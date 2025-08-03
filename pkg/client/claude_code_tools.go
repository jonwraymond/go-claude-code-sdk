package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	sdkerrors "github.com/jonwraymond/go-claude-code-sdk/pkg/errors"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// Dangerous command patterns that should be blocked for security
var dangerousCommandPatterns = []*regexp.Regexp{
	regexp.MustCompile(`rm\s+-rf\s+/`),           // Dangerous rm operations
	regexp.MustCompile(`;\s*rm\s+-rf`),           // Command chaining with rm
	regexp.MustCompile(`\|\s*rm\s+-rf`),          // Piped rm operations
	regexp.MustCompile(`&&\s*rm\s+-rf`),          // Chained rm operations  
	regexp.MustCompile(`curl.*\|\s*sh`),          // Download and execute
	regexp.MustCompile(`wget.*\|\s*sh`),          // Download and execute
	regexp.MustCompile(`eval\s*\$\(`),            // Code injection via eval
	regexp.MustCompile(`\$\(.*curl`),             // Command substitution with curl
	regexp.MustCompile(`nc\s+-l`),                // Netcat listeners
	regexp.MustCompile(`/dev/tcp/`),              // TCP connections
}

// validateCommand checks if a command contains dangerous patterns
func validateCommand(command string) error {
	command = strings.TrimSpace(command)
	if command == "" {
		return sdkerrors.NewValidationError("command", "", "non-empty string", "command cannot be empty")
	}

	// Check for dangerous patterns
	for _, pattern := range dangerousCommandPatterns {
		if pattern.MatchString(command) {
			return sdkerrors.NewValidationError("command", command, "safe command", "command contains potentially dangerous pattern")
		}
	}

	return nil
}

// validateFilePath checks if a file path is safe to access
func validateFilePath(path, workingDir string) error {
	// Clean the path to normalize it
	cleanPath := filepath.Clean(path)
	
	// Convert to absolute path
	var absPath string
	if filepath.IsAbs(cleanPath) {
		absPath = cleanPath
	} else {
		absPath = filepath.Join(workingDir, cleanPath)
	}
	
	// Check if the resolved path is within the working directory
	absWorkingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return fmt.Errorf("invalid working directory: %v", err)
	}
	
	relPath, err := filepath.Rel(absWorkingDir, absPath)
	if err != nil {
		return fmt.Errorf("invalid file path: %v", err)
	}
	
	// Check for directory traversal attempts
	if strings.HasPrefix(relPath, "..") || strings.Contains(relPath, "/../") {
		return fmt.Errorf("path traversal detected: %s", path)
	}
	
	return nil
}

// ClaudeCodeToolManager manages tool execution for Claude Code CLI integration.
// Unlike the traditional ToolManager that executes functions directly, this manager
// works with Claude Code's subprocess-based tool system and MCP servers.
//
// The tool manager supports:
// - Claude Code built-in tools (read, write, edit, search, etc.)
// - MCP server tool integration
// - Tool discovery and introspection
// - Tool execution via Claude Code CLI
// - Result parsing and error handling
//
// Example usage:
//
//	toolManager := NewClaudeCodeToolManager(client)
//
//	// Discover available tools
//	tools, err := toolManager.DiscoverTools(ctx)
//
//	// Execute a built-in tool
//	result, err := toolManager.ExecuteTool(ctx, &ClaudeCodeTool{
//		Name: "read_file",
//		Parameters: map[string]any{
//			"path": "main.go",
//		},
//	})
//
//	// Execute an MCP tool
//	result, err := toolManager.ExecuteMCPTool(ctx, "filesystem", "read_file", params)
type ClaudeCodeToolManager struct {
	client       *ClaudeCodeClient
	builtInTools map[string]*ClaudeCodeToolDefinition
	mcpTools     map[string]map[string]*ClaudeCodeToolDefinition // serverName -> toolName -> definition
	mu           sync.RWMutex

	// Tool execution configuration
	config *ClaudeCodeToolConfig
}

// ClaudeCodeToolConfig provides configuration for tool execution.
type ClaudeCodeToolConfig struct {
	// MaxExecutionTime limits tool execution time (default: 30s)
	MaxExecutionTime time.Duration

	// EnableCaching enables caching of tool results (default: true)
	EnableCaching bool

	// CacheDuration sets how long to cache tool results (default: 5m)
	CacheDuration time.Duration

	// AllowFileSystemAccess controls file system tool access (default: true)
	AllowFileSystemAccess bool

	// AllowNetworkAccess controls network tool access (default: false)
	AllowNetworkAccess bool
}

// DefaultClaudeCodeToolConfig returns default configuration for Claude Code tools.
func DefaultClaudeCodeToolConfig() *ClaudeCodeToolConfig {
	return &ClaudeCodeToolConfig{
		MaxExecutionTime:      30 * time.Second,
		EnableCaching:         true,
		CacheDuration:         5 * time.Minute,
		AllowFileSystemAccess: true,
		AllowNetworkAccess:    false,
	}
}

// ClaudeCodeToolDefinition defines a Claude Code tool.
type ClaudeCodeToolDefinition struct {
	// Name is the tool name
	Name string `json:"name"`

	// Description describes what the tool does
	Description string `json:"description"`

	// Category groups related tools
	Category string `json:"category"`

	// Parameters defines the tool's input parameters
	Parameters map[string]ToolParameter `json:"parameters"`

	// RequiredParameters lists required parameter names
	RequiredParameters []string `json:"required_parameters"`

	// Source indicates where the tool comes from (builtin, mcp:servername)
	Source string `json:"source"`

	// Permissions lists required permissions
	Permissions []string `json:"permissions"`
}

// ToolParameter defines a tool parameter.
type ToolParameter struct {
	// Type is the parameter type (string, number, boolean, array, object)
	Type string `json:"type"`

	// Description describes the parameter
	Description string `json:"description"`

	// Default is the default value (if any)
	Default any `json:"default,omitempty"`

	// Enum lists allowed values (if restricted)
	Enum []any `json:"enum,omitempty"`

	// Pattern is a regex pattern for validation (strings only)
	Pattern string `json:"pattern,omitempty"`
}

// ClaudeCodeTool represents a tool to execute.
type ClaudeCodeTool struct {
	// Name is the tool name
	Name string

	// Parameters are the tool parameters
	Parameters map[string]any

	// MCPServer is the MCP server name (if this is an MCP tool)
	MCPServer string
}

// ClaudeCodeToolResult represents the result of tool execution.
type ClaudeCodeToolResult struct {
	// Success indicates if the tool executed successfully
	Success bool `json:"success"`

	// Output contains the tool output
	Output any `json:"output,omitempty"`

	// Error contains error information
	Error string `json:"error,omitempty"`

	// ExecutionTime is how long the tool took
	ExecutionTime time.Duration `json:"execution_time"`

	// Metadata contains additional information
	Metadata map[string]any `json:"metadata,omitempty"`
}

// NewClaudeCodeToolManager creates a new Claude Code tool manager.
func NewClaudeCodeToolManager(client *ClaudeCodeClient) *ClaudeCodeToolManager {
	config := DefaultClaudeCodeToolConfig()
	return NewClaudeCodeToolManagerWithConfig(client, config)
}

// NewClaudeCodeToolManagerWithConfig creates a new tool manager with custom configuration.
func NewClaudeCodeToolManagerWithConfig(client *ClaudeCodeClient, config *ClaudeCodeToolConfig) *ClaudeCodeToolManager {
	manager := &ClaudeCodeToolManager{
		client:       client,
		builtInTools: make(map[string]*ClaudeCodeToolDefinition),
		mcpTools:     make(map[string]map[string]*ClaudeCodeToolDefinition),
		config:       config,
	}

	// Initialize built-in tools
	manager.initializeBuiltInTools()

	return manager
}

// initializeBuiltInTools initializes Claude Code's built-in tools.
func (tm *ClaudeCodeToolManager) initializeBuiltInTools() {
	// File system tools
	tm.builtInTools["read_file"] = &ClaudeCodeToolDefinition{
		Name:        "read_file",
		Description: "Read the contents of a file",
		Category:    "file_system",
		Parameters: map[string]ToolParameter{
			"path": {
				Type:        "string",
				Description: "Path to the file to read",
			},
			"encoding": {
				Type:        "string",
				Description: "File encoding (default: utf-8)",
				Default:     "utf-8",
			},
		},
		RequiredParameters: []string{"path"},
		Source:             "builtin",
		Permissions:        []string{"file_read"},
	}

	tm.builtInTools["write_file"] = &ClaudeCodeToolDefinition{
		Name:        "write_file",
		Description: "Write content to a file",
		Category:    "file_system",
		Parameters: map[string]ToolParameter{
			"path": {
				Type:        "string",
				Description: "Path to the file to write",
			},
			"content": {
				Type:        "string",
				Description: "Content to write to the file",
			},
			"encoding": {
				Type:        "string",
				Description: "File encoding (default: utf-8)",
				Default:     "utf-8",
			},
			"create_dirs": {
				Type:        "boolean",
				Description: "Create parent directories if they don't exist",
				Default:     true,
			},
		},
		RequiredParameters: []string{"path", "content"},
		Source:             "builtin",
		Permissions:        []string{"file_write"},
	}

	tm.builtInTools["edit_file"] = &ClaudeCodeToolDefinition{
		Name:        "edit_file",
		Description: "Edit specific parts of a file",
		Category:    "file_system",
		Parameters: map[string]ToolParameter{
			"path": {
				Type:        "string",
				Description: "Path to the file to edit",
			},
			"edits": {
				Type:        "array",
				Description: "List of edits to apply",
			},
		},
		RequiredParameters: []string{"path", "edits"},
		Source:             "builtin",
		Permissions:        []string{"file_write"},
	}

	tm.builtInTools["list_files"] = &ClaudeCodeToolDefinition{
		Name:        "list_files",
		Description: "List files in a directory",
		Category:    "file_system",
		Parameters: map[string]ToolParameter{
			"path": {
				Type:        "string",
				Description: "Directory path to list",
				Default:     ".",
			},
			"recursive": {
				Type:        "boolean",
				Description: "List files recursively",
				Default:     false,
			},
			"pattern": {
				Type:        "string",
				Description: "File pattern to match (glob)",
			},
		},
		RequiredParameters: []string{},
		Source:             "builtin",
		Permissions:        []string{"file_read"},
	}

	// Code analysis tools
	tm.builtInTools["search_code"] = &ClaudeCodeToolDefinition{
		Name:        "search_code",
		Description: "Search for patterns in code",
		Category:    "code_analysis",
		Parameters: map[string]ToolParameter{
			"pattern": {
				Type:        "string",
				Description: "Search pattern (regex supported)",
			},
			"path": {
				Type:        "string",
				Description: "Path to search in",
				Default:     ".",
			},
			"file_pattern": {
				Type:        "string",
				Description: "File pattern to search in",
			},
			"case_sensitive": {
				Type:        "boolean",
				Description: "Case sensitive search",
				Default:     true,
			},
		},
		RequiredParameters: []string{"pattern"},
		Source:             "builtin",
		Permissions:        []string{"file_read"},
	}

	tm.builtInTools["analyze_code"] = &ClaudeCodeToolDefinition{
		Name:        "analyze_code",
		Description: "Analyze code structure and patterns",
		Category:    "code_analysis",
		Parameters: map[string]ToolParameter{
			"path": {
				Type:        "string",
				Description: "File or directory to analyze",
			},
			"analysis_type": {
				Type:        "string",
				Description: "Type of analysis to perform",
				Enum:        []any{"complexity", "dependencies", "structure", "quality"},
			},
		},
		RequiredParameters: []string{"path", "analysis_type"},
		Source:             "builtin",
		Permissions:        []string{"file_read"},
	}

	// Terminal/command tools
	tm.builtInTools["run_command"] = &ClaudeCodeToolDefinition{
		Name:        "run_command",
		Description: "Execute a shell command",
		Category:    "terminal",
		Parameters: map[string]ToolParameter{
			"command": {
				Type:        "string",
				Description: "Command to execute",
			},
			"working_dir": {
				Type:        "string",
				Description: "Working directory for the command",
				Default:     ".",
			},
			"timeout": {
				Type:        "number",
				Description: "Command timeout in seconds",
				Default:     30,
			},
		},
		RequiredParameters: []string{"command"},
		Source:             "builtin",
		Permissions:        []string{"command_execute"},
	}

	// Git tools
	tm.builtInTools["git_status"] = &ClaudeCodeToolDefinition{
		Name:        "git_status",
		Description: "Get git repository status",
		Category:    "git",
		Parameters: map[string]ToolParameter{
			"path": {
				Type:        "string",
				Description: "Repository path",
				Default:     ".",
			},
			"verbose": {
				Type:        "boolean",
				Description: "Show verbose output",
				Default:     false,
			},
		},
		RequiredParameters: []string{},
		Source:             "builtin",
		Permissions:        []string{"git_read"},
	}

	tm.builtInTools["git_diff"] = &ClaudeCodeToolDefinition{
		Name:        "git_diff",
		Description: "Show git differences",
		Category:    "git",
		Parameters: map[string]ToolParameter{
			"path": {
				Type:        "string",
				Description: "Repository path",
				Default:     ".",
			},
			"staged": {
				Type:        "boolean",
				Description: "Show staged changes",
				Default:     false,
			},
			"commit": {
				Type:        "string",
				Description: "Compare with specific commit",
			},
		},
		RequiredParameters: []string{},
		Source:             "builtin",
		Permissions:        []string{"git_read"},
	}
}

// DiscoverTools discovers all available tools including MCP server tools.
func (tm *ClaudeCodeToolManager) DiscoverTools(ctx context.Context) ([]*ClaudeCodeToolDefinition, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tools := make([]*ClaudeCodeToolDefinition, 0)

	// Add built-in tools
	for _, tool := range tm.builtInTools {
		tools = append(tools, tool)
	}

	// Discover MCP server tools
	enabledServers := tm.client.mcpManager.GetEnabledServers()
	for serverName, serverConfig := range enabledServers {
		// Query MCP server for available tools
		serverTools, err := tm.queryMCPServerTools(ctx, serverName, serverConfig)
		if err != nil {
			// Log error but continue with other servers
			continue
		}

		// Add server tools to registry
		if _, exists := tm.mcpTools[serverName]; !exists {
			tm.mcpTools[serverName] = make(map[string]*ClaudeCodeToolDefinition)
		}

		for _, tool := range serverTools {
			tool.Source = fmt.Sprintf("mcp:%s", serverName)
			tm.mcpTools[serverName][tool.Name] = tool
			tools = append(tools, tool)
		}
	}

	return tools, nil
}

// queryMCPServerTools queries an MCP server for available tools.
func (tm *ClaudeCodeToolManager) queryMCPServerTools(ctx context.Context, serverName string, config *types.MCPServerConfig) ([]*ClaudeCodeToolDefinition, error) {
	// This would typically involve executing the MCP server with a discovery command
	// For now, return empty as MCP servers would need to be running
	return []*ClaudeCodeToolDefinition{}, nil
}

// GetTool retrieves a tool definition by name.
func (tm *ClaudeCodeToolManager) GetTool(name string) (*ClaudeCodeToolDefinition, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// Check built-in tools
	if tool, exists := tm.builtInTools[name]; exists {
		return tool, nil
	}

	// Check MCP tools
	for _, serverTools := range tm.mcpTools {
		if tool, exists := serverTools[name]; exists {
			return tool, nil
		}
	}

	return nil, sdkerrors.NewValidationError("tool", name, "exists", "tool not found")
}

// ListTools returns all available tools.
func (tm *ClaudeCodeToolManager) ListTools() []*ClaudeCodeToolDefinition {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tools := make([]*ClaudeCodeToolDefinition, 0)

	// Add built-in tools
	for _, tool := range tm.builtInTools {
		tools = append(tools, tool)
	}

	// Add MCP tools
	for _, serverTools := range tm.mcpTools {
		for _, tool := range serverTools {
			tools = append(tools, tool)
		}
	}

	return tools
}

// ExecuteTool executes a Claude Code tool.
func (tm *ClaudeCodeToolManager) ExecuteTool(ctx context.Context, tool *ClaudeCodeTool) (*ClaudeCodeToolResult, error) {
	if tool == nil {
		return nil, sdkerrors.NewValidationError("tool", "", "required", "tool cannot be nil")
	}

	start := time.Now()

	// Determine tool type and execute appropriately
	if tool.MCPServer != "" {
		result, err := tm.executeMCPTool(ctx, tool)
		if err != nil {
			return &ClaudeCodeToolResult{
				Success:       false,
				Error:         err.Error(),
				ExecutionTime: time.Since(start),
			}, err
		}
		result.ExecutionTime = time.Since(start)
		return result, nil
	}

	// Execute built-in tool
	result, err := tm.executeBuiltInTool(ctx, tool)
	if err != nil {
		return &ClaudeCodeToolResult{
			Success:       false,
			Error:         err.Error(),
			ExecutionTime: time.Since(start),
		}, err
	}

	result.ExecutionTime = time.Since(start)
	return result, nil
}

// executeBuiltInTool executes a built-in Claude Code tool.
func (tm *ClaudeCodeToolManager) executeBuiltInTool(ctx context.Context, tool *ClaudeCodeTool) (*ClaudeCodeToolResult, error) {
	// Get tool definition
	definition, exists := tm.builtInTools[tool.Name]
	if !exists {
		return nil, sdkerrors.NewValidationError("tool", tool.Name, "exists", "built-in tool not found")
	}

	// Check permissions
	if err := tm.checkPermissions(definition); err != nil {
		return nil, err
	}

	// Validate parameters
	if err := tm.validateParameters(definition, tool.Parameters); err != nil {
		return nil, err
	}

	// Execute based on tool name
	switch tool.Name {
	case "read_file":
		return tm.executeReadFile(ctx, tool.Parameters)
	case "write_file":
		return tm.executeWriteFile(ctx, tool.Parameters)
	case "edit_file":
		return tm.executeEditFile(ctx, tool.Parameters)
	case "list_files":
		return tm.executeListFiles(ctx, tool.Parameters)
	case "search_code":
		return tm.executeSearchCode(ctx, tool.Parameters)
	case "analyze_code":
		return tm.executeAnalyzeCode(ctx, tool.Parameters)
	case "run_command":
		return tm.executeRunCommand(ctx, tool.Parameters)
	case "git_status":
		return tm.executeGitStatus(ctx, tool.Parameters)
	case "git_diff":
		return tm.executeGitDiff(ctx, tool.Parameters)
	default:
		return nil, sdkerrors.NewValidationError("tool", tool.Name, "supported", "unsupported built-in tool")
	}
}

// executeMCPTool executes an MCP server tool.
func (tm *ClaudeCodeToolManager) executeMCPTool(ctx context.Context, tool *ClaudeCodeTool) (*ClaudeCodeToolResult, error) {
	// Get MCP server configuration
	serverConfig, err := tm.client.mcpManager.GetServer(tool.MCPServer)
	if err != nil {
		return nil, sdkerrors.WrapError(err, sdkerrors.CategoryValidation, "MCP_SERVER", "MCP server not found")
	}

	if !serverConfig.Enabled {
		return nil, sdkerrors.NewValidationError("mcp_server", tool.MCPServer, "enabled", "MCP server is disabled")
	}

	// Execute MCP tool via Claude Code
	// This would involve calling Claude with the MCP tool request
	// For now, return a placeholder
	return &ClaudeCodeToolResult{
		Success: false,
		Error:   "MCP tool execution not yet implemented",
	}, nil
}

// Tool execution implementations

func (tm *ClaudeCodeToolManager) executeReadFile(ctx context.Context, params map[string]any) (*ClaudeCodeToolResult, error) {
	path, ok := params["path"].(string)
	if !ok {
		return nil, sdkerrors.NewValidationError("path", "", "string", "path must be a string")
	}

	// Validate file path for security
	if err := validateFilePath(path, tm.client.workingDir); err != nil {
		return &ClaudeCodeToolResult{
			Success: false,
			Error:   fmt.Sprintf("file path validation failed: %v", err),
		}, nil
	}

	// Resolve path relative to working directory
	if !filepath.IsAbs(path) {
		path = filepath.Join(tm.client.workingDir, path)
	}

	// Read file - path validated above
	content, err := os.ReadFile(path) // #nosec G304 - path validated above
	if err != nil {
		return &ClaudeCodeToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read file: %v", err),
		}, nil
	}

	return &ClaudeCodeToolResult{
		Success: true,
		Output:  string(content),
		Metadata: map[string]any{
			"path": path,
			"size": len(content),
		},
	}, nil
}

func (tm *ClaudeCodeToolManager) executeWriteFile(ctx context.Context, params map[string]any) (*ClaudeCodeToolResult, error) {
	path, ok := params["path"].(string)
	if !ok {
		return nil, sdkerrors.NewValidationError("path", "", "string", "path must be a string")
	}

	content, ok := params["content"].(string)
	if !ok {
		return nil, sdkerrors.NewValidationError("content", "", "string", "content must be a string")
	}

	// Validate file path for security
	if err := validateFilePath(path, tm.client.workingDir); err != nil {
		return &ClaudeCodeToolResult{
			Success: false,
			Error:   fmt.Sprintf("file path validation failed: %v", err),
		}, nil
	}

	// Resolve path relative to working directory
	if !filepath.IsAbs(path) {
		path = filepath.Join(tm.client.workingDir, path)
	}

	// Create directories if needed
	createDirs := true
	if val, exists := params["create_dirs"]; exists {
		createDirs, _ = val.(bool)
	}

	if createDirs {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return &ClaudeCodeToolResult{
				Success: false,
				Error:   fmt.Sprintf("failed to create directories: %v", err),
			}, nil
		}
	}

	// Write file
	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		return &ClaudeCodeToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write file: %v", err),
		}, nil
	}

	return &ClaudeCodeToolResult{
		Success: true,
		Output:  fmt.Sprintf("File written successfully: %s", path),
		Metadata: map[string]any{
			"path": path,
			"size": len(content),
		},
	}, nil
}

func (tm *ClaudeCodeToolManager) executeEditFile(ctx context.Context, params map[string]any) (*ClaudeCodeToolResult, error) {
	// This would be implemented using Claude Code's edit functionality
	// For now, return a placeholder
	return &ClaudeCodeToolResult{
		Success: false,
		Error:   "edit_file tool not yet implemented",
	}, nil
}

func (tm *ClaudeCodeToolManager) executeListFiles(ctx context.Context, params map[string]any) (*ClaudeCodeToolResult, error) {
	path := "."
	if p, ok := params["path"].(string); ok {
		path = p
	}

	// Resolve path relative to working directory
	if !filepath.IsAbs(path) {
		path = filepath.Join(tm.client.workingDir, path)
	}

	recursive := false
	if r, ok := params["recursive"].(bool); ok {
		recursive = r
	}

	pattern := ""
	if p, ok := params["pattern"].(string); ok {
		pattern = p
	}

	var files []string

	if recursive {
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}

			// Apply pattern matching if specified
			if pattern != "" {
				matched, err := filepath.Match(pattern, filepath.Base(filePath))
				if err != nil || !matched {
					return nil
				}
			}

			relPath, _ := filepath.Rel(tm.client.workingDir, filePath)
			files = append(files, relPath)
			return nil
		})

		if err != nil {
			return &ClaudeCodeToolResult{
				Success: false,
				Error:   fmt.Sprintf("failed to walk directory: %v", err),
			}, nil
		}
	} else {
		entries, err := os.ReadDir(path)
		if err != nil {
			return &ClaudeCodeToolResult{
				Success: false,
				Error:   fmt.Sprintf("failed to read directory: %v", err),
			}, nil
		}

		for _, entry := range entries {
			// Apply pattern matching if specified
			if pattern != "" {
				matched, err := filepath.Match(pattern, entry.Name())
				if err != nil || !matched {
					continue
				}
			}

			relPath, _ := filepath.Rel(tm.client.workingDir, filepath.Join(path, entry.Name()))
			files = append(files, relPath)
		}
	}

	return &ClaudeCodeToolResult{
		Success: true,
		Output:  files,
		Metadata: map[string]any{
			"path":      path,
			"count":     len(files),
			"recursive": recursive,
			"pattern":   pattern,
		},
	}, nil
}

func (tm *ClaudeCodeToolManager) executeSearchCode(ctx context.Context, params map[string]any) (*ClaudeCodeToolResult, error) {
	pattern, ok := params["pattern"].(string)
	if !ok {
		return nil, sdkerrors.NewValidationError("pattern", "", "string", "pattern must be a string")
	}

	// Use grep or similar for code search
	searchPath := "."
	if p, ok := params["path"].(string); ok {
		searchPath = p
	}

	// Resolve path relative to working directory
	if !filepath.IsAbs(searchPath) {
		searchPath = filepath.Join(tm.client.workingDir, searchPath)
	}

	// Build grep command
	args := []string{"-r", "-n"}

	if cs, ok := params["case_sensitive"].(bool); ok && !cs {
		args = append(args, "-i")
	}

	if fp, ok := params["file_pattern"].(string); ok && fp != "" {
		args = append(args, "--include="+fp)
	}

	args = append(args, pattern, searchPath)

	// Execute grep
	cmd := exec.CommandContext(ctx, "grep", args...)
	output, err := cmd.Output()

	// grep returns exit code 1 when no matches found
	if err != nil && cmd.ProcessState.ExitCode() == 1 {
		return &ClaudeCodeToolResult{
			Success: true,
			Output:  []string{}, // No matches
			Metadata: map[string]any{
				"pattern": pattern,
				"path":    searchPath,
				"matches": 0,
			},
		}, nil
	}

	if err != nil {
		return &ClaudeCodeToolResult{
			Success: false,
			Error:   fmt.Sprintf("search failed: %v", err),
		}, nil
	}

	// Parse grep output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	matches := make([]map[string]any, 0)

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse grep output format: filename:line_number:content
		parts := strings.SplitN(line, ":", 3)
		if len(parts) >= 3 {
			matches = append(matches, map[string]any{
				"file":    parts[0],
				"line":    parts[1],
				"content": parts[2],
			})
		}
	}

	return &ClaudeCodeToolResult{
		Success: true,
		Output:  matches,
		Metadata: map[string]any{
			"pattern": pattern,
			"path":    searchPath,
			"matches": len(matches),
		},
	}, nil
}

func (tm *ClaudeCodeToolManager) executeAnalyzeCode(ctx context.Context, params map[string]any) (*ClaudeCodeToolResult, error) {
	// This would integrate with code analysis tools
	// For now, return a placeholder
	return &ClaudeCodeToolResult{
		Success: false,
		Error:   "analyze_code tool not yet implemented",
	}, nil
}

func (tm *ClaudeCodeToolManager) executeRunCommand(ctx context.Context, params map[string]any) (*ClaudeCodeToolResult, error) {
	command, ok := params["command"].(string)
	if !ok {
		return nil, sdkerrors.NewValidationError("command", "", "string", "command must be a string")
	}

	// Validate command for security
	if err := validateCommand(command); err != nil {
		return &ClaudeCodeToolResult{
			Success: false,
			Error:   fmt.Sprintf("command validation failed: %v", err),
		}, nil
	}

	workingDir := tm.client.workingDir
	if wd, ok := params["working_dir"].(string); ok {
		if !filepath.IsAbs(wd) {
			workingDir = filepath.Join(tm.client.workingDir, wd)
		} else {
			workingDir = wd
		}
	}

	timeout := 30.0
	if t, ok := params["timeout"].(float64); ok {
		timeout = t
	}

	// Create command with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Execute command through shell - validated for security above
	cmd := exec.CommandContext(ctx, "sh", "-c", command) // #nosec G204 - command validated above
	cmd.Dir = workingDir

	output, err := cmd.CombinedOutput()

	if err != nil {
		return &ClaudeCodeToolResult{
			Success: false,
			Error:   fmt.Sprintf("command failed: %v", err),
			Output:  string(output),
			Metadata: map[string]any{
				"command":     command,
				"working_dir": workingDir,
				"exit_code":   cmd.ProcessState.ExitCode(),
			},
		}, nil
	}

	return &ClaudeCodeToolResult{
		Success: true,
		Output:  string(output),
		Metadata: map[string]any{
			"command":     command,
			"working_dir": workingDir,
			"exit_code":   0,
		},
	}, nil
}

func (tm *ClaudeCodeToolManager) executeGitStatus(ctx context.Context, params map[string]any) (*ClaudeCodeToolResult, error) {
	path := tm.client.workingDir
	if p, ok := params["path"].(string); ok {
		if !filepath.IsAbs(p) {
			path = filepath.Join(tm.client.workingDir, p)
		} else {
			path = p
		}
	}

	args := []string{"status"}
	if verbose, ok := params["verbose"].(bool); ok && verbose {
		args = append(args, "-v")
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = path

	output, err := cmd.Output()
	if err != nil {
		return &ClaudeCodeToolResult{
			Success: false,
			Error:   fmt.Sprintf("git status failed: %v", err),
		}, nil
	}

	return &ClaudeCodeToolResult{
		Success: true,
		Output:  string(output),
		Metadata: map[string]any{
			"path": path,
		},
	}, nil
}

func (tm *ClaudeCodeToolManager) executeGitDiff(ctx context.Context, params map[string]any) (*ClaudeCodeToolResult, error) {
	path := tm.client.workingDir
	if p, ok := params["path"].(string); ok {
		if !filepath.IsAbs(p) {
			path = filepath.Join(tm.client.workingDir, p)
		} else {
			path = p
		}
	}

	args := []string{"diff"}

	if staged, ok := params["staged"].(bool); ok && staged {
		args = append(args, "--staged")
	}

	if commit, ok := params["commit"].(string); ok && commit != "" {
		args = append(args, commit)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = path

	output, err := cmd.Output()
	if err != nil {
		return &ClaudeCodeToolResult{
			Success: false,
			Error:   fmt.Sprintf("git diff failed: %v", err),
		}, nil
	}

	return &ClaudeCodeToolResult{
		Success: true,
		Output:  string(output),
		Metadata: map[string]any{
			"path": path,
		},
	}, nil
}

// checkPermissions checks if the tool has required permissions.
func (tm *ClaudeCodeToolManager) checkPermissions(definition *ClaudeCodeToolDefinition) error {
	for _, perm := range definition.Permissions {
		switch perm {
		case "file_read":
			if !tm.config.AllowFileSystemAccess {
				return sdkerrors.NewValidationError("permission", perm, "allowed", "file system access is disabled")
			}
		case "file_write":
			if !tm.config.AllowFileSystemAccess {
				return sdkerrors.NewValidationError("permission", perm, "allowed", "file system access is disabled")
			}
		case "network":
			if !tm.config.AllowNetworkAccess {
				return sdkerrors.NewValidationError("permission", perm, "allowed", "network access is disabled")
			}
		}
	}
	return nil
}

// validateParameters validates tool parameters against the definition.
func (tm *ClaudeCodeToolManager) validateParameters(definition *ClaudeCodeToolDefinition, params map[string]any) error {
	// Check required parameters
	for _, required := range definition.RequiredParameters {
		if _, exists := params[required]; !exists {
			return sdkerrors.NewValidationError("parameter", required, "provided", fmt.Sprintf("required parameter '%s' is missing", required))
		}
	}

	// Validate parameter types and constraints
	for name, value := range params {
		paramDef, exists := definition.Parameters[name]
		if !exists {
			// Skip unknown parameters
			continue
		}

		// Type validation
		if err := tm.validateParameterType(name, value, paramDef); err != nil {
			return err
		}

		// Enum validation
		if len(paramDef.Enum) > 0 {
			valid := false
			for _, allowed := range paramDef.Enum {
				if value == allowed {
					valid = true
					break
				}
			}
			if !valid {
				return sdkerrors.NewValidationError("parameter", name, fmt.Sprintf("one of %v", paramDef.Enum), fmt.Sprintf("invalid value: %v", value))
			}
		}
	}

	return nil
}

// validateParameterType validates a parameter's type.
func (tm *ClaudeCodeToolManager) validateParameterType(name string, value any, paramDef ToolParameter) error {
	switch paramDef.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return sdkerrors.NewValidationError("parameter", name, "string", fmt.Sprintf("expected string, got %T", value))
		}
	case "number":
		switch value.(type) {
		case float64, float32, int, int32, int64:
			// Valid number types
		default:
			return sdkerrors.NewValidationError("parameter", name, "number", fmt.Sprintf("expected number, got %T", value))
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return sdkerrors.NewValidationError("parameter", name, "boolean", fmt.Sprintf("expected boolean, got %T", value))
		}
	case "array":
		if _, ok := value.([]any); !ok {
			return sdkerrors.NewValidationError("parameter", name, "array", fmt.Sprintf("expected array, got %T", value))
		}
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return sdkerrors.NewValidationError("parameter", name, "object", fmt.Sprintf("expected object, got %T", value))
		}
	}
	return nil
}

// ConvertToClaudeAPITools converts Claude Code tools to API tool format.
func (tm *ClaudeCodeToolManager) ConvertToClaudeAPITools() []types.Tool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tools := make([]types.Tool, 0)

	// Convert built-in tools
	for _, toolDef := range tm.builtInTools {
		tool := types.Tool{
			Name:        toolDef.Name,
			Description: toolDef.Description,
			InputSchema: tm.convertParametersToSchema(toolDef),
		}
		tools = append(tools, tool)
	}

	// Convert MCP tools
	for _, serverTools := range tm.mcpTools {
		for _, toolDef := range serverTools {
			tool := types.Tool{
				Name:        toolDef.Name,
				Description: toolDef.Description,
				InputSchema: tm.convertParametersToSchema(toolDef),
			}
			tools = append(tools, tool)
		}
	}

	return tools
}

// convertParametersToSchema converts tool parameters to API schema format.
func (tm *ClaudeCodeToolManager) convertParametersToSchema(toolDef *ClaudeCodeToolDefinition) types.ToolInputSchema {
	schema := types.ToolInputSchema{
		Type:       "object",
		Properties: make(map[string]types.ToolProperty),
		Required:   toolDef.RequiredParameters,
	}

	for name, param := range toolDef.Parameters {
		prop := types.ToolProperty{
			Type:        param.Type,
			Description: param.Description,
		}

		// Add enum if present
		if len(param.Enum) > 0 {
			prop.Enum = make([]any, len(param.Enum))
			for i, v := range param.Enum {
				prop.Enum[i] = fmt.Sprintf("%v", v)
			}
		}

		// Add pattern if present
		if param.Pattern != "" {
			prop.Pattern = param.Pattern
		}

		schema.Properties[name] = prop
	}

	return schema
}

// HandleToolUse processes a tool use request from Claude.
func (tm *ClaudeCodeToolManager) HandleToolUse(ctx context.Context, toolUse *types.ToolUse) (*types.ToolResult, error) {
	if toolUse == nil {
		return nil, sdkerrors.NewValidationError("tool_use", "", "required", "tool use cannot be nil")
	}

	// Create Claude Code tool from tool use
	tool := &ClaudeCodeTool{
		Name:       toolUse.Name,
		Parameters: toolUse.Input,
	}

	// Check if it's an MCP tool (name contains server prefix)
	if strings.Contains(toolUse.Name, ":") {
		parts := strings.SplitN(toolUse.Name, ":", 2)
		if len(parts) == 2 {
			tool.MCPServer = parts[0]
			tool.Name = parts[1]
		}
	}

	// Execute the tool
	result, err := tm.ExecuteTool(ctx, tool)
	if err != nil {
		return &types.ToolResult{
			ToolUseID: toolUse.ID,
			IsError:   true,
			Content: []types.ContentBlock{
				{
					Type: "text",
					Text: fmt.Sprintf("Tool execution failed: %v", err),
				},
			},
		}, nil
	}

	// Convert result to content blocks
	content, err := tm.resultToContent(result)
	if err != nil {
		return &types.ToolResult{
			ToolUseID: toolUse.ID,
			IsError:   true,
			Content: []types.ContentBlock{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to serialize tool result: %v", err),
				},
			},
		}, nil
	}

	return &types.ToolResult{
		ToolUseID: toolUse.ID,
		IsError:   !result.Success,
		Content:   content,
	}, nil
}

// resultToContent converts a tool result to content blocks.
func (tm *ClaudeCodeToolManager) resultToContent(result *ClaudeCodeToolResult) ([]types.ContentBlock, error) {
	if !result.Success {
		return []types.ContentBlock{
			{
				Type: "text",
				Text: result.Error,
			},
		}, nil
	}

	// Convert output to JSON for consistent formatting
	output, err := json.MarshalIndent(map[string]any{
		"output":   result.Output,
		"metadata": result.Metadata,
	}, "", "  ")
	if err != nil {
		// Fallback to string representation
		return []types.ContentBlock{
			{
				Type: "text",
				Text: fmt.Sprintf("%v", result.Output),
			},
		}, nil
	}

	return []types.ContentBlock{
		{
			Type: "text",
			Text: string(output),
		},
	}, nil
}
