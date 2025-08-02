package client

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	sdkerrors "github.com/jonwraymond/go-claude-code-sdk/pkg/errors"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// ClaudeCodeClient implements the Client interface by executing the Claude Code CLI subprocess.
// This provides programmatic access to Claude Code's capabilities through its command-line interface.
//
// The client manages subprocess lifecycle, handles streaming responses, and provides
// session management for multi-turn conversations with Claude Code.
//
// Architecture:
// - Subprocess Management: Spawns and manages claude CLI processes
// - Session Persistence: Maintains conversation context across interactions
// - Project Awareness: Automatically detects and uses project context
// - MCP Integration: Supports Model Context Protocol for tool extensions
// - Streaming Support: Real-time response processing with chunk handling
//
// Example usage:
//
//	config := &types.ClaudeCodeConfig{
//		WorkingDirectory: "/path/to/project",
//		SessionID:        "my-session",
//		Model:           "claude-3-5-sonnet-20241022",
//	}
//	client, err := NewClaudeCodeClient(ctx, config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	response, err := client.Query(ctx, &types.QueryRequest{
//		Messages: []types.Message{
//			{Role: types.RoleUser, Content: "Analyze this codebase"},
//		},
//	})
type ClaudeCodeClient struct {
	config        *types.ClaudeCodeConfig
	workingDir    string
	sessionID     string
	claudeCodeCmd string
	mu            sync.RWMutex
	closed        bool

	// Process management
	activeProcesses map[string]*exec.Cmd
	processMu       sync.Mutex

	// MCP management
	mcpManager *MCPManager

	// Project context management
	projectContextManager *ProjectContextManager

	// Tool management
	toolManager *ClaudeCodeToolManager

	// Session management
	sessionManager *ClaudeCodeSessionManager
}

// NewClaudeCodeClient creates a new Claude Code client with subprocess management.
// The client will spawn claude CLI processes as needed and manage their lifecycle.
//
// Configuration options:
// - WorkingDirectory: Project directory for context (defaults to current directory)
// - SessionID: Session identifier for conversation persistence
// - Model: Claude model to use (defaults to claude-3-5-sonnet-20241022)
// - ClaudeCodePath: Path to claude executable (auto-detected if not provided)
// - MCPServers: MCP server configurations for tool extensions
// - Environment: Environment variables for subprocess execution
func NewClaudeCodeClient(ctx context.Context, config *types.ClaudeCodeConfig) (*ClaudeCodeClient, error) {
	if config == nil {
		return nil, sdkerrors.NewConfigurationError("config", "configuration is required")
	}

	// Set defaults
	if config.WorkingDirectory == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, sdkerrors.WrapError(err, sdkerrors.CategoryConfiguration, "WORKING_DIR", "failed to get current directory")
		}
		config.WorkingDirectory = wd
	}

	if config.SessionID == "" {
		// Generate a simple UUID-like ID for session
		config.SessionID = generateSessionID()
	}

	if config.Model == "" {
		config.Model = "claude-3-5-sonnet-20241022"
	}

	// Apply defaults including authentication method detection
	config.ApplyDefaults()

	// Find claude command (skip in test mode)
	var claudeCmd string
	if config.TestMode {
		claudeCmd = "/bin/echo" // Use echo as a mock command for testing
	} else {
		var err error
		claudeCmd, err = findClaudeCodeCommand(config.ClaudeCodePath)
		if err != nil {
			return nil, sdkerrors.WrapError(err, sdkerrors.CategoryConfiguration, "CLAUDE_CODE_PATH", "failed to locate claude code executable")
		}
	}

	// Validate working directory
	if _, err := os.Stat(config.WorkingDirectory); os.IsNotExist(err) {
		return nil, sdkerrors.NewConfigurationError("working_directory", "working directory does not exist: "+config.WorkingDirectory)
	}

	client := &ClaudeCodeClient{
		config:          config,
		workingDir:      config.WorkingDirectory,
		sessionID:       config.SessionID,
		claudeCodeCmd:   claudeCmd,
		activeProcesses: make(map[string]*exec.Cmd),
	}

	// Initialize MCP manager
	client.mcpManager = NewMCPManager(client)

	// Initialize project context manager
	client.projectContextManager = NewProjectContextManager(client)

	// Initialize tool manager
	client.toolManager = NewClaudeCodeToolManager(client)

	// Initialize session manager
	client.sessionManager = NewClaudeCodeSessionManager(client)

	return client, nil
}

// Query sends a synchronous request to Claude Code and returns the complete response.
// This executes the claude CLI in non-interactive mode and captures the full output.
//
// The request is converted to appropriate CLI arguments and the response is parsed
// from the command output. Project context is automatically included based on the
// working directory configuration.
func (c *ClaudeCodeClient) Query(ctx context.Context, request *types.QueryRequest) (*types.QueryResponse, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return nil, sdkerrors.NewInternalError("CLIENT_CLOSED", "client has been closed")
	}
	c.mu.RUnlock()

	if request == nil {
		return nil, sdkerrors.NewValidationError("request", "", "required", "request cannot be nil")
	}

	// Build claude command arguments
	args, err := c.buildClaudeArgs(request, false)
	if err != nil {
		return nil, sdkerrors.WrapError(err, sdkerrors.CategoryValidation, "ARGS_BUILD", "failed to build claude arguments")
	}

	// Execute claude command
	cmd := exec.CommandContext(ctx, c.claudeCodeCmd, args...)
	cmd.Dir = c.workingDir

	// Set environment variables
	cmd.Env = append(os.Environ(), c.buildEnvironment()...)

	// Debug: print the command being executed
	if c.config.Debug {
		fmt.Printf("[DEBUG] Executing: %s %s\n", c.claudeCodeCmd, strings.Join(args, " "))
		fmt.Printf("[DEBUG] Working directory: %s\n", c.workingDir)
		fmt.Printf("[DEBUG] Environment: %v\n", c.buildEnvironment())
	}

	// Capture output
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, sdkerrors.NewInternalError("CLAUDE_EXECUTION", fmt.Sprintf("claude command failed: %s", string(exitErr.Stderr)))
		}
		return nil, sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "CLAUDE_EXECUTION", "failed to execute claude command")
	}

	// Parse response
	response, err := c.parseClaudeOutput(string(output))
	if err != nil {
		return nil, sdkerrors.WrapError(err, sdkerrors.CategoryAPI, "RESPONSE_PARSE", "failed to parse claude output")
	}

	return response, nil
}

// QueryStream sends a streaming request to Claude Code and returns a streaming response.
// This executes claude in streaming mode and returns a stream interface for real-time
// processing of response chunks.
//
// The stream must be closed when done to prevent resource leaks and properly
// terminate the underlying claude process.
func (c *ClaudeCodeClient) QueryStream(ctx context.Context, request *types.QueryRequest) (types.QueryStream, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return nil, sdkerrors.NewInternalError("CLIENT_CLOSED", "client has been closed")
	}
	c.mu.RUnlock()

	if request == nil {
		return nil, sdkerrors.NewValidationError("request", "", "required", "request cannot be nil")
	}

	// Build claude command arguments for streaming
	args, err := c.buildClaudeArgs(request, true)
	if err != nil {
		return nil, sdkerrors.WrapError(err, sdkerrors.CategoryValidation, "ARGS_BUILD", "failed to build claude streaming arguments")
	}

	// Create and start claude process
	cmd := exec.CommandContext(ctx, c.claudeCodeCmd, args...)
	cmd.Dir = c.workingDir
	cmd.Env = append(os.Environ(), c.buildEnvironment()...)

	// Create pipes for stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "PIPE_CREATION", "failed to create stdout pipe")
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "PROCESS_START", "failed to start claude process")
	}

	// Track the process
	processID := fmt.Sprintf("stream-%d", time.Now().UnixNano())
	c.processMu.Lock()
	c.activeProcesses[processID] = cmd
	c.processMu.Unlock()

	// Create streaming query stream
	stream := &claudeCodeQueryStream{
		cmd:       cmd,
		stdout:    stdout,
		ctx:       ctx,
		processID: processID,
		client:    c,
	}

	return stream, nil
}

// Close gracefully shuts down the client and terminates any active processes.
// This method should be called when the client is no longer needed to prevent
// resource leaks and orphaned processes.
func (c *ClaudeCodeClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	// Close session manager
	if c.sessionManager != nil {
		c.sessionManager.Close()
	}

	// Terminate all active processes
	c.processMu.Lock()
	for processID, cmd := range c.activeProcesses {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		delete(c.activeProcesses, processID)
	}
	c.processMu.Unlock()

	return nil
}

// ExecuteCommand executes a Claude Code command and returns the result.
func (c *ClaudeCodeClient) ExecuteCommand(ctx context.Context, cmd *types.Command) (*types.CommandResult, error) {
	executor := NewCommandExecutor(c)
	return executor.ExecuteCommand(ctx, cmd)
}

// ExecuteSlashCommand executes a Claude Code slash command (e.g., "/read file.go").
func (c *ClaudeCodeClient) ExecuteSlashCommand(ctx context.Context, slashCommand string) (*types.CommandResult, error) {
	executor := NewCommandExecutor(c)
	return executor.ExecuteSlashCommand(ctx, slashCommand)
}

// GetProjectContext returns information about the current project context.
func (c *ClaudeCodeClient) GetProjectContext(ctx context.Context) (*types.ProjectContext, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return nil, sdkerrors.NewInternalError("CLIENT_CLOSED", "client has been closed")
	}
	c.mu.RUnlock()

	context := &types.ProjectContext{
		WorkingDirectory: c.workingDir,
		ProjectName:      filepath.Base(c.workingDir),
	}

	// Detect primary language
	if lang := c.detectPrimaryLanguage(); lang != "" {
		context.Language = lang
	}

	// Detect framework
	if framework := c.detectFramework(); framework != "" {
		context.Framework = framework
	}

	// Get git information if available
	if gitInfo := c.getGitInfo(); gitInfo != nil {
		context.GitRepository = gitInfo
	}

	// Get file information
	if files := c.getProjectFiles(); files != nil {
		context.Files = files
	}

	return context, nil
}

// SetWorkingDirectory changes the working directory for Claude Code operations.
func (c *ClaudeCodeClient) SetWorkingDirectory(ctx context.Context, path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return sdkerrors.NewInternalError("CLIENT_CLOSED", "client has been closed")
	}

	// Validate the directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return sdkerrors.NewValidationError("path", path, "exists", "directory does not exist")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return sdkerrors.WrapError(err, sdkerrors.CategoryValidation, "PATH_ABS", "failed to convert to absolute path")
	}

	c.workingDir = absPath
	c.config.WorkingDirectory = absPath

	return nil
}

// detectPrimaryLanguage detects the primary programming language in the project.
func (c *ClaudeCodeClient) detectPrimaryLanguage() string {
	languageCounts := make(map[string]int)

	// Language extension mappings
	languageMap := map[string]string{
		".go":    "Go",
		".js":    "JavaScript",
		".ts":    "TypeScript",
		".py":    "Python",
		".java":  "Java",
		".cpp":   "C++",
		".c":     "C",
		".cs":    "C#",
		".rb":    "Ruby",
		".php":   "PHP",
		".rs":    "Rust",
		".swift": "Swift",
		".kt":    "Kotlin",
		".scala": "Scala",
	}

	// Walk the directory and count files by extension
	_ = filepath.Walk(c.workingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if lang, exists := languageMap[ext]; exists {
			languageCounts[lang]++
		}
		return nil
	})

	// Find the language with the most files
	maxCount := 0
	primaryLang := ""
	for lang, count := range languageCounts {
		if count > maxCount {
			maxCount = count
			primaryLang = lang
		}
	}

	return primaryLang
}

// detectFramework detects the primary framework used in the project.
func (c *ClaudeCodeClient) detectFramework() string {
	// Check for common framework indicators
	frameworkFiles := map[string]func() string{
		"package.json":     c.detectJSFramework,
		"go.mod":           func() string { return "Go Modules" },
		"Cargo.toml":       func() string { return "Cargo" },
		"pom.xml":          func() string { return "Maven" },
		"build.gradle":     func() string { return "Gradle" },
		"requirements.txt": func() string { return "Python" },
		"Pipfile":          func() string { return "Pipenv" },
		"composer.json":    func() string { return "Composer" },
	}

	for filename, detector := range frameworkFiles {
		if _, err := os.Stat(filepath.Join(c.workingDir, filename)); err == nil {
			return detector()
		}
	}

	return ""
}

// detectJSFramework detects JavaScript/TypeScript framework from package.json.
func (c *ClaudeCodeClient) detectJSFramework() string {
	packagePath := filepath.Join(c.workingDir, "package.json")
	data, err := os.ReadFile(packagePath)
	if err != nil {
		return "Node.js"
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return "Node.js"
	}

	// Check for popular frameworks
	frameworks := []struct {
		name    string
		modules []string
	}{
		{"Next.js", []string{"next"}},
		{"React", []string{"react"}},
		{"Vue.js", []string{"vue"}},
		{"Angular", []string{"@angular/core"}},
		{"Express", []string{"express"}},
		{"Fastify", []string{"fastify"}},
		{"NestJS", []string{"@nestjs/core"}},
	}

	allDeps := make(map[string]string)
	for k, v := range pkg.Dependencies {
		allDeps[k] = v
	}
	for k, v := range pkg.DevDependencies {
		allDeps[k] = v
	}

	for _, framework := range frameworks {
		for _, module := range framework.modules {
			if _, exists := allDeps[module]; exists {
				return framework.name
			}
		}
	}

	return "Node.js"
}

// getGitInfo retrieves git repository information.
func (c *ClaudeCodeClient) getGitInfo() *types.GitInfo {
	gitDir := filepath.Join(c.workingDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil
	}

	gitInfo := &types.GitInfo{}

	// Get current branch
	if branch := c.getGitBranch(); branch != "" {
		gitInfo.Branch = branch
	}

	// Get remote URL
	if remoteURL := c.getGitRemoteURL(); remoteURL != "" {
		gitInfo.RemoteURL = remoteURL
	}

	// Get current commit hash
	if commit := c.getGitCommitHash(); commit != "" {
		gitInfo.CommitHash = commit
	}

	// Check if repo is dirty
	gitInfo.IsDirty = c.isGitDirty()

	return gitInfo
}

// getGitBranch gets the current git branch.
func (c *ClaudeCodeClient) getGitBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = c.workingDir
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getGitRemoteURL gets the git remote URL.
func (c *ClaudeCodeClient) getGitRemoteURL() string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = c.workingDir
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getGitCommitHash gets the current commit hash.
func (c *ClaudeCodeClient) getGitCommitHash() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = c.workingDir
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// isGitDirty checks if the git repository has uncommitted changes.
func (c *ClaudeCodeClient) isGitDirty() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = c.workingDir
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// getProjectFiles gets information about project files.
func (c *ClaudeCodeClient) getProjectFiles() *types.ProjectFiles {
	files := &types.ProjectFiles{
		Languages:      make(map[string]int),
		ImportantFiles: make([]string, 0),
		Structure:      make(map[string]any),
	}

	// Count files by extension
	extensionMap := map[string]string{
		".go":   "Go",
		".js":   "JavaScript",
		".ts":   "TypeScript",
		".py":   "Python",
		".java": "Java",
		".cpp":  "C++",
		".c":    "C",
		".rs":   "Rust",
		".rb":   "Ruby",
		".php":  "PHP",
	}

	totalFiles := 0
	importantFileNames := []string{
		"README.md", "README.txt", "LICENSE", "Makefile",
		"package.json", "go.mod", "Cargo.toml", "requirements.txt",
		"docker-compose.yml", "Dockerfile", ".gitignore",
	}

	_ = filepath.Walk(c.workingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(info.Name(), ".") && info.Name() != ".gitignore" {
			return nil
		}

		totalFiles++

		// Count by language
		ext := filepath.Ext(path)
		if lang, exists := extensionMap[ext]; exists {
			files.Languages[lang]++
		}

		// Check for important files
		for _, important := range importantFileNames {
			if info.Name() == important {
				relPath, _ := filepath.Rel(c.workingDir, path)
				files.ImportantFiles = append(files.ImportantFiles, relPath)
				break
			}
		}

		return nil
	})

	files.TotalFiles = totalFiles
	return files
}

// MCP returns the MCP manager for managing Model Context Protocol servers.
func (c *ClaudeCodeClient) MCP() *MCPManager {
	return c.mcpManager
}

// EnableMCPServer enables an MCP server by name.
func (c *ClaudeCodeClient) EnableMCPServer(ctx context.Context, name string) error {
	if err := c.mcpManager.EnableServer(name); err != nil {
		return err
	}
	return c.mcpManager.ApplyConfiguration(ctx)
}

// DisableMCPServer disables an MCP server by name.
func (c *ClaudeCodeClient) DisableMCPServer(ctx context.Context, name string) error {
	if err := c.mcpManager.DisableServer(name); err != nil {
		return err
	}
	return c.mcpManager.ApplyConfiguration(ctx)
}

// AddMCPServer adds an MCP server configuration.
func (c *ClaudeCodeClient) AddMCPServer(ctx context.Context, name string, config *types.MCPServerConfig) error {
	if err := c.mcpManager.AddServer(name, config); err != nil {
		return err
	}
	return c.mcpManager.ApplyConfiguration(ctx)
}

// RemoveMCPServer removes an MCP server configuration.
func (c *ClaudeCodeClient) RemoveMCPServer(ctx context.Context, name string) error {
	if err := c.mcpManager.RemoveServer(name); err != nil {
		return err
	}
	return c.mcpManager.ApplyConfiguration(ctx)
}

// ListMCPServers returns a list of all configured MCP servers.
func (c *ClaudeCodeClient) ListMCPServers() map[string]*types.MCPServerConfig {
	return c.mcpManager.ListServers()
}

// GetMCPServer returns the configuration for a specific MCP server.
func (c *ClaudeCodeClient) GetMCPServer(name string) (*types.MCPServerConfig, error) {
	return c.mcpManager.GetServer(name)
}

// SetupCommonMCPServers adds commonly used MCP servers with default configurations.
func (c *ClaudeCodeClient) SetupCommonMCPServers(ctx context.Context) error {
	if err := c.mcpManager.AddCommonServers(); err != nil {
		return err
	}
	return c.mcpManager.ApplyConfiguration(ctx)
}

// ProjectContext returns the project context manager for advanced project analysis.
func (c *ClaudeCodeClient) ProjectContext() *ProjectContextManager {
	return c.projectContextManager
}

// GetEnhancedProjectContext returns an enhanced project context with deep analysis.
func (c *ClaudeCodeClient) GetEnhancedProjectContext(ctx context.Context) (*types.ProjectContext, error) {
	return c.projectContextManager.GetEnhancedProjectContext(ctx)
}

// InvalidateProjectContextCache invalidates the cached project context.
func (c *ClaudeCodeClient) InvalidateProjectContextCache() {
	c.projectContextManager.InvalidateCache()
}

// SetProjectContextCacheDuration sets the cache duration for project context.
func (c *ClaudeCodeClient) SetProjectContextCacheDuration(duration time.Duration) {
	c.projectContextManager.SetCacheDuration(duration)
}

// GetProjectContextCacheInfo returns information about the project context cache status.
func (c *ClaudeCodeClient) GetProjectContextCacheInfo() map[string]any {
	return c.projectContextManager.GetCacheInfo()
}

// Tools returns the tool manager for managing Claude Code tools.
func (c *ClaudeCodeClient) Tools() *ClaudeCodeToolManager {
	return c.toolManager
}

// DiscoverTools discovers all available tools including MCP server tools.
func (c *ClaudeCodeClient) DiscoverTools(ctx context.Context) ([]*ClaudeCodeToolDefinition, error) {
	return c.toolManager.DiscoverTools(ctx)
}

// ExecuteTool executes a Claude Code tool.
func (c *ClaudeCodeClient) ExecuteTool(ctx context.Context, tool *ClaudeCodeTool) (*ClaudeCodeToolResult, error) {
	return c.toolManager.ExecuteTool(ctx, tool)
}

// ListTools returns all available tools.
func (c *ClaudeCodeClient) ListTools() []*ClaudeCodeToolDefinition {
	return c.toolManager.ListTools()
}

// GetTool retrieves a tool definition by name.
func (c *ClaudeCodeClient) GetTool(name string) (*ClaudeCodeToolDefinition, error) {
	return c.toolManager.GetTool(name)
}

// Sessions returns the session manager for managing Claude Code conversations.
func (c *ClaudeCodeClient) Sessions() *ClaudeCodeSessionManager {
	return c.sessionManager
}

// CreateSession creates a new conversation session with Claude Code.
func (c *ClaudeCodeClient) CreateSession(ctx context.Context, sessionID string) (*ClaudeCodeSession, error) {
	return c.sessionManager.CreateSession(ctx, sessionID)
}

// GetSession retrieves an existing session by ID.
func (c *ClaudeCodeClient) GetSession(sessionID string) (*ClaudeCodeSession, error) {
	return c.sessionManager.GetSession(sessionID)
}

// ListSessions returns all active session IDs.
func (c *ClaudeCodeClient) ListSessions() []string {
	return c.sessionManager.ListSessions()
}

// buildClaudeArgs constructs command-line arguments for the claude CLI based on the request.
func (c *ClaudeCodeClient) buildClaudeArgs(request *types.QueryRequest, streaming bool) ([]string, error) {
	args := make([]string, 0)

	// Add print flag for non-interactive use
	if !streaming {
		args = append(args, "--print")
	}

	// Add model selection
	if request.Model != "" {
		args = append(args, "--model", request.Model)
	} else if c.config.Model != "" {
		args = append(args, "--model", c.config.Model)
	}

	// Add session ID for conversation persistence
	if c.sessionID != "" {
		args = append(args, "--session-id", c.sessionID)
	}

	// Add streaming flag if requested
	if streaming {
		args = append(args, "--stream")
	}

	// Add MCP configuration if there are enabled servers
	if enabledServers := c.mcpManager.GetEnabledServers(); len(enabledServers) > 0 {
		configPath := filepath.Join(c.workingDir, ".claude", "mcp.json")
		if _, err := os.Stat(configPath); err == nil {
			args = append(args, "--mcp-config", configPath)
		}
	}

	// Add system prompt if provided
	if request.System != "" {
		args = append(args, "--system", request.System)
	}

	// Add max tokens if specified
	if request.MaxTokens > 0 {
		args = append(args, "--max-tokens", fmt.Sprintf("%d", request.MaxTokens))
	}

	// Add temperature if specified
	if request.Temperature > 0 {
		args = append(args, "--temperature", fmt.Sprintf("%.2f", request.Temperature))
	}

	// Convert messages to prompt
	if len(request.Messages) > 0 {
		prompt, err := c.messagesToPrompt(request.Messages)
		if err != nil {
			return nil, sdkerrors.WrapError(err, sdkerrors.CategoryValidation, "MESSAGE_CONVERSION", "failed to convert messages to prompt")
		}
		args = append(args, prompt)
	}

	return args, nil
}

// messagesToPrompt converts a slice of messages into a single prompt string for claude CLI.
func (c *ClaudeCodeClient) messagesToPrompt(messages []types.Message) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	// For single user message, return directly
	if len(messages) == 1 && messages[0].Role == types.RoleUser {
		return c.extractTextContent(messages[0].Content), nil
	}

	// For multi-turn conversation, format as conversation
	var prompt strings.Builder
	for i, msg := range messages {
		if i > 0 {
			prompt.WriteString("\n\n")
		}

		switch msg.Role {
		case types.RoleUser:
			prompt.WriteString("Human: ")
		case types.RoleAssistant:
			prompt.WriteString("Assistant: ")
		case types.RoleSystem:
			prompt.WriteString("System: ")
		}

		prompt.WriteString(c.extractTextContent(msg.Content))
	}

	return prompt.String(), nil
}

// extractTextContent extracts text content from message content blocks.
func (c *ClaudeCodeClient) extractTextContent(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []types.ContentBlock:
		var text strings.Builder
		for _, block := range v {
			if block.Type == "text" {
				text.WriteString(block.Text)
			}
		}
		return text.String()
	default:
		return fmt.Sprintf("%v", content)
	}
}

// buildEnvironment constructs environment variables for the claude subprocess.
func (c *ClaudeCodeClient) buildEnvironment() []string {
	env := make([]string, 0)

	// Handle authentication based on configured method
	switch c.config.AuthMethod {
	case types.AuthTypeAPIKey:
		// Add API key from config for API key authentication
		if c.config.APIKey != "" {
			env = append(env, "ANTHROPIC_API_KEY="+c.config.APIKey)
		}
	case types.AuthTypeSubscription:
		// For subscription auth, the CLI handles authentication automatically
		// No additional environment variables needed
	default:
		// Fallback: if API key is available, use it
		if c.config.APIKey != "" {
			env = append(env, "ANTHROPIC_API_KEY="+c.config.APIKey)
		}
	}

	// Add custom environment variables
	for key, value := range c.config.Environment {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	return env
}

// parseClaudeOutput parses the output from claude CLI into a QueryResponse.
func (c *ClaudeCodeClient) parseClaudeOutput(output string) (*types.QueryResponse, error) {
	// Try to parse as JSON first (in case of structured output)
	var jsonResponse types.QueryResponse
	if err := json.Unmarshal([]byte(output), &jsonResponse); err == nil {
		return &jsonResponse, nil
	}

	// Fall back to treating as plain text response
	response := &types.QueryResponse{
		Content: []types.ContentBlock{
			{
				Type: "text",
				Text: strings.TrimSpace(output),
			},
		},
		StopReason: "end_turn",
	}

	return response, nil
}

// findClaudeCodeCommand locates the claude executable in the system.
func findClaudeCodeCommand(customPath string) (string, error) {
	// If custom path is provided, use it
	if customPath != "" {
		if _, err := os.Stat(customPath); err != nil {
			return "", fmt.Errorf("claude executable not found at %s: %w", customPath, err)
		}
		return customPath, nil
	}

	// Try common locations
	candidates := []string{
		"claude",     // In PATH
		"npx claude", // Via npx
		filepath.Join(os.Getenv("HOME"), ".local", "bin", "claude"), // Local install
		"/usr/local/bin/claude", // System install
	}

	for _, candidate := range candidates {
		if strings.Contains(candidate, " ") {
			// For commands like "npx claude", test by running with --version
			parts := strings.Fields(candidate)
			cmd := exec.Command(parts[0], append(parts[1:], "--version")...)
			if err := cmd.Run(); err == nil {
				return candidate, nil
			}
		} else {
			// For single commands, check if file exists
			if _, err := exec.LookPath(candidate); err == nil {
				return candidate, nil
			}
		}
	}

	return "", fmt.Errorf("claude code executable not found. Please install claude code via 'npm install -g @anthropic-ai/claude-code' or provide custom path")
}

// generateSessionID generates a UUID v4 session ID for Claude CLI
func generateSessionID() string {
	// Generate a proper UUID v4
	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		// Fallback to time-based if crypto/rand fails
		now := time.Now().UnixNano()
		for i := range uuid {
			uuid[i] = byte(now >> (8 * (i % 8)))
		}
	}
	
	// Set version (4) and variant bits
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant bits
	
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uint32(uuid[0])<<24|uint32(uuid[1])<<16|uint32(uuid[2])<<8|uint32(uuid[3]),
		uint16(uuid[4])<<8|uint16(uuid[5]),
		uint16(uuid[6])<<8|uint16(uuid[7]),
		uint16(uuid[8])<<8|uint16(uuid[9]),
		uint64(uuid[10])<<40|uint64(uuid[11])<<32|uint64(uuid[12])<<24|uint64(uuid[13])<<16|uint64(uuid[14])<<8|uint64(uuid[15]),
	)
}

// claudeCodeQueryStream implements QueryStream for Claude Code subprocess streaming.
type claudeCodeQueryStream struct {
	cmd       *exec.Cmd
	stdout    io.ReadCloser
	ctx       context.Context
	processID string
	client    *ClaudeCodeClient
	scanner   *bufio.Scanner
	closed    bool
	mu        sync.Mutex
}

// Recv receives the next chunk from the streaming Claude Code process.
func (s *claudeCodeQueryStream) Recv() (*types.StreamChunk, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, sdkerrors.NewInternalError("STREAM_CLOSED", "stream has been closed")
	}

	// Check context cancellation
	select {
	case <-s.ctx.Done():
		return nil, sdkerrors.WrapError(s.ctx.Err(), sdkerrors.CategoryNetwork, "CONTEXT_CANCELED", "request context canceled")
	default:
	}

	// Initialize scanner if not already done
	if s.scanner == nil {
		s.scanner = bufio.NewScanner(s.stdout)
		s.scanner.Buffer(make([]byte, 64*1024), 64*1024)
	}

	// Read the next line
	if !s.scanner.Scan() {
		// Check for scanning errors
		if err := s.scanner.Err(); err != nil {
			return nil, sdkerrors.WrapError(err, sdkerrors.CategoryNetwork, "STREAM_READ", "failed to read from claude process")
		}
		// Stream ended, check if process completed successfully
		if err := s.cmd.Wait(); err != nil {
			return nil, sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "PROCESS_ERROR", "claude process failed")
		}
		return &types.StreamChunk{Done: true}, nil
	}

	line := s.scanner.Text()

	// Parse the line into a stream chunk
	chunk := &types.StreamChunk{
		Content: line + "\n",
		Done:    false,
	}

	return chunk, nil
}

// Close terminates the streaming Claude Code process and releases resources.
func (s *claudeCodeQueryStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true

	// Close stdout pipe
	if s.stdout != nil {
		s.stdout.Close()
	}

	// Terminate the process
	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
	}

	// Remove from client's active processes
	s.client.processMu.Lock()
	delete(s.client.activeProcesses, s.processID)
	s.client.processMu.Unlock()

	return nil
}
