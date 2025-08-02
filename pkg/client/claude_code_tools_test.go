package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func TestNewClaudeCodeToolManager(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	toolManager := client.Tools()
	if toolManager == nil {
		t.Fatal("Tool manager should not be nil")
	}

	// Check that built-in tools are initialized
	tools := toolManager.ListTools()
	if len(tools) == 0 {
		t.Error("Expected built-in tools to be initialized")
	}

	// Check for specific built-in tools
	expectedTools := []string{
		"read_file",
		"write_file",
		"edit_file",
		"list_files",
		"search_code",
		"analyze_code",
		"run_command",
		"git_status",
		"git_diff",
	}

	for _, expectedName := range expectedTools {
		found := false
		for _, tool := range tools {
			if tool.Name == expectedName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find built-in tool: %s", expectedName)
		}
	}
}

func TestClaudeCodeToolManager_GetTool(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	toolManager := client.Tools()

	tests := []struct {
		name        string
		toolName    string
		expectError bool
	}{
		{
			name:        "Get existing tool",
			toolName:    "read_file",
			expectError: false,
		},
		{
			name:        "Get non-existent tool",
			toolName:    "non_existent_tool",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, err := toolManager.GetTool(tt.toolName)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tool == nil {
					t.Error("Expected tool to be returned")
				}
				if tool != nil && tool.Name != tt.toolName {
					t.Errorf("Expected tool name %s, got %s", tt.toolName, tool.Name)
				}
			}
		})
	}
}

func TestClaudeCodeToolManager_ExecuteReadFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file
	testContent := "Hello, World!\nThis is a test file."
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	toolManager := client.Tools()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		checkResult func(*ClaudeCodeToolResult) error
	}{
		{
			name: "Read existing file with relative path",
			params: map[string]interface{}{
				"path": "test.txt",
			},
			expectError: false,
			checkResult: func(result *ClaudeCodeToolResult) error {
				if !result.Success {
					return fmt.Errorf("Expected success, got failure: %s", result.Error)
				}
				if result.Output != testContent {
					return fmt.Errorf("Expected content %q, got %q", testContent, result.Output)
				}
				return nil
			},
		},
		{
			name: "Read existing file with absolute path",
			params: map[string]interface{}{
				"path": testFile,
			},
			expectError: false,
			checkResult: func(result *ClaudeCodeToolResult) error {
				if !result.Success {
					return fmt.Errorf("Expected success, got failure: %s", result.Error)
				}
				if result.Output != testContent {
					return fmt.Errorf("Expected content %q, got %q", testContent, result.Output)
				}
				return nil
			},
		},
		{
			name: "Read non-existent file",
			params: map[string]interface{}{
				"path": "non_existent.txt",
			},
			expectError: false, // Tool execution doesn't error, but result shows failure
			checkResult: func(result *ClaudeCodeToolResult) error {
				if result.Success {
					return fmt.Errorf("Expected failure for non-existent file")
				}
				if result.Error == "" {
					return fmt.Errorf("Expected error message for non-existent file")
				}
				return nil
			},
		},
		{
			name:        "Missing path parameter",
			params:      map[string]interface{}{},
			expectError: true,
		},
		{
			name: "Invalid path type",
			params: map[string]interface{}{
				"path": 123,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := &ClaudeCodeTool{
				Name:       "read_file",
				Parameters: tt.params,
			}

			result, err := toolManager.ExecuteTool(ctx, tool)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("Expected result to be returned")
				}
				if tt.checkResult != nil {
					if err := tt.checkResult(result); err != nil {
						t.Error(err)
					}
				}
			}
		})
	}
}

func TestClaudeCodeToolManager_ExecuteWriteFile(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	toolManager := client.Tools()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		checkResult func(*ClaudeCodeToolResult) error
		checkFile   func() error
	}{
		{
			name: "Write file with relative path",
			params: map[string]interface{}{
				"path":    "output.txt",
				"content": "Test content",
			},
			expectError: false,
			checkResult: func(result *ClaudeCodeToolResult) error {
				if !result.Success {
					return fmt.Errorf("Expected success, got failure: %s", result.Error)
				}
				return nil
			},
			checkFile: func() error {
				content, err := os.ReadFile(filepath.Join(tempDir, "output.txt"))
				if err != nil {
					return fmt.Errorf("Failed to read written file: %v", err)
				}
				if string(content) != "Test content" {
					return fmt.Errorf("Expected content 'Test content', got %q", string(content))
				}
				return nil
			},
		},
		{
			name: "Write file with directory creation",
			params: map[string]interface{}{
				"path":        "subdir/nested/file.txt",
				"content":     "Nested content",
				"create_dirs": true,
			},
			expectError: false,
			checkResult: func(result *ClaudeCodeToolResult) error {
				if !result.Success {
					return fmt.Errorf("Expected success, got failure: %s", result.Error)
				}
				return nil
			},
			checkFile: func() error {
				content, err := os.ReadFile(filepath.Join(tempDir, "subdir", "nested", "file.txt"))
				if err != nil {
					return fmt.Errorf("Failed to read written file: %v", err)
				}
				if string(content) != "Nested content" {
					return fmt.Errorf("Expected content 'Nested content', got %q", string(content))
				}
				return nil
			},
		},
		{
			name:        "Missing path parameter",
			params:      map[string]interface{}{"content": "test"},
			expectError: true,
		},
		{
			name:        "Missing content parameter",
			params:      map[string]interface{}{"path": "test.txt"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := &ClaudeCodeTool{
				Name:       "write_file",
				Parameters: tt.params,
			}

			result, err := toolManager.ExecuteTool(ctx, tool)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("Expected result to be returned")
				}
				if tt.checkResult != nil {
					if err := tt.checkResult(result); err != nil {
						t.Error(err)
					}
				}
				if tt.checkFile != nil {
					if err := tt.checkFile(); err != nil {
						t.Error(err)
					}
				}
			}
		})
	}
}

func TestClaudeCodeToolManager_ExecuteListFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create test directory structure
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(tempDir, "file2.go"), []byte("content2"), 0644)
	os.MkdirAll(filepath.Join(tempDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tempDir, "subdir", "file3.txt"), []byte("content3"), 0644)
	os.WriteFile(filepath.Join(tempDir, "subdir", "file4.go"), []byte("content4"), 0644)

	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	toolManager := client.Tools()

	tests := []struct {
		name          string
		params        map[string]interface{}
		expectError   bool
		expectedCount int
		checkFiles    func([]string) error
	}{
		{
			name:          "List files in current directory",
			params:        map[string]interface{}{},
			expectError:   false,
			expectedCount: 3, // file1.txt, file2.go, subdir
			checkFiles: func(files []string) error {
				// Should not include files from subdirectories
				for _, file := range files {
					if file == "subdir/file3.txt" || file == "subdir/file4.go" {
						return fmt.Errorf("Non-recursive list should not include %s", file)
					}
				}
				return nil
			},
		},
		{
			name: "List files recursively",
			params: map[string]interface{}{
				"recursive": true,
			},
			expectError:   false,
			expectedCount: 6, // All files including current dir (.) and subdir
		},
		{
			name: "List files with pattern",
			params: map[string]interface{}{
				"pattern":   "*.go",
				"recursive": true,
			},
			expectError:   false,
			expectedCount: 2, // file2.go and subdir/file4.go
			checkFiles: func(files []string) error {
				for _, file := range files {
					matched, err := filepath.Match("*.go", filepath.Base(file))
					if err != nil || !matched {
						return fmt.Errorf("File %s doesn't match pattern *.go", file)
					}
				}
				return nil
			},
		},
		{
			name: "List files in subdirectory",
			params: map[string]interface{}{
				"path": "subdir",
			},
			expectError:   false,
			expectedCount: 2, // file3.txt and file4.go
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := &ClaudeCodeTool{
				Name:       "list_files",
				Parameters: tt.params,
			}

			result, err := toolManager.ExecuteTool(ctx, tool)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("Expected result to be returned")
				}
				if !result.Success {
					t.Errorf("Expected success, got failure: %s", result.Error)
				}

				files, ok := result.Output.([]string)
				if !ok {
					t.Fatalf("Expected output to be []string, got %T", result.Output)
				}

				if len(files) != tt.expectedCount {
					t.Errorf("Expected %d files, got %d: %v", tt.expectedCount, len(files), files)
				}

				if tt.checkFiles != nil {
					if err := tt.checkFiles(files); err != nil {
						t.Error(err)
					}
				}
			}
		})
	}
}

func TestClaudeCodeToolManager_ExecuteSearchCode(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files with searchable content
	goFile := `package main

func main() {
	fmt.Println("Hello, World!")
}

func helper() {
	fmt.Println("Helper function")
}`
	os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(goFile), 0644)

	txtFile := `This is a test file.
It contains some text.
Hello, World!`
	os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte(txtFile), 0644)

	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	toolManager := client.Tools()

	tests := []struct {
		name            string
		params          map[string]interface{}
		expectError     bool
		expectedMatches int
	}{
		{
			name: "Search for Hello",
			params: map[string]interface{}{
				"pattern": "Hello",
			},
			expectError:     false,
			expectedMatches: 2, // One in main.go, one in test.txt
		},
		{
			name: "Search case insensitive",
			params: map[string]interface{}{
				"pattern":        "hello",
				"case_sensitive": false,
			},
			expectError:     false,
			expectedMatches: 2,
		},
		{
			name: "Search with file pattern",
			params: map[string]interface{}{
				"pattern":      "Hello",
				"file_pattern": "*.go",
			},
			expectError:     false,
			expectedMatches: 1, // Only in main.go
		},
		{
			name: "Search for non-existent pattern",
			params: map[string]interface{}{
				"pattern": "NonExistentPattern",
			},
			expectError:     false,
			expectedMatches: 0,
		},
		{
			name:        "Missing pattern parameter",
			params:      map[string]interface{}{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := &ClaudeCodeTool{
				Name:       "search_code",
				Parameters: tt.params,
			}

			result, err := toolManager.ExecuteTool(ctx, tool)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("Expected result to be returned")
				}
				if !result.Success {
					t.Errorf("Expected success, got failure: %s", result.Error)
				}

				matches, ok := result.Output.([]map[string]interface{})
				if !ok {
					// Could be empty array
					if emptyMatches, ok := result.Output.([]string); ok && len(emptyMatches) == 0 {
						matches = []map[string]interface{}{}
					} else {
						t.Fatalf("Expected output to be []map[string]interface{}, got %T", result.Output)
					}
				}

				if len(matches) != tt.expectedMatches {
					t.Errorf("Expected %d matches, got %d", tt.expectedMatches, len(matches))
				}
			}
		})
	}
}

func TestClaudeCodeToolManager_ExecuteRunCommand(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	toolManager := client.Tools()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		checkResult func(*ClaudeCodeToolResult) error
	}{
		{
			name: "Run echo command",
			params: map[string]interface{}{
				"command": "echo 'Hello, World!'",
			},
			expectError: false,
			checkResult: func(result *ClaudeCodeToolResult) error {
				if !result.Success {
					return fmt.Errorf("Expected success, got failure: %s", result.Error)
				}
				output := result.Output.(string)
				if output != "Hello, World!\n" {
					return fmt.Errorf("Expected output 'Hello, World!\\n', got %q", output)
				}
				return nil
			},
		},
		{
			name: "Run pwd command",
			params: map[string]interface{}{
				"command": "pwd",
			},
			expectError: false,
			checkResult: func(result *ClaudeCodeToolResult) error {
				if !result.Success {
					return fmt.Errorf("Expected success, got failure: %s", result.Error)
				}
				output := result.Output.(string)
				if output != tempDir+"\n" {
					return fmt.Errorf("Expected output %q, got %q", tempDir+"\n", output)
				}
				return nil
			},
		},
		{
			name: "Run command with custom working directory",
			params: map[string]interface{}{
				"command":     "pwd",
				"working_dir": "/tmp",
			},
			expectError: false,
			checkResult: func(result *ClaudeCodeToolResult) error {
				if !result.Success {
					return fmt.Errorf("Expected success, got failure: %s", result.Error)
				}
				output := result.Output.(string)
				if output != "/tmp\n" {
					return fmt.Errorf("Expected output '/tmp\\n', got %q", output)
				}
				return nil
			},
		},
		{
			name: "Run failing command",
			params: map[string]interface{}{
				"command": "exit 1",
			},
			expectError: false, // Tool execution doesn't error, but result shows failure
			checkResult: func(result *ClaudeCodeToolResult) error {
				if result.Success {
					return fmt.Errorf("Expected failure for failing command")
				}
				metadata := result.Metadata
				if exitCode, ok := metadata["exit_code"].(int); ok {
					if exitCode != 1 {
						return fmt.Errorf("Expected exit code 1, got %d", exitCode)
					}
				} else {
					return fmt.Errorf("Expected exit_code in metadata")
				}
				return nil
			},
		},
		{
			name: "Run command with timeout",
			params: map[string]interface{}{
				"command": "sleep 2",
				"timeout": 0.1, // 100ms timeout
			},
			expectError: false,
			checkResult: func(result *ClaudeCodeToolResult) error {
				if result.Success {
					return fmt.Errorf("Expected failure due to timeout")
				}
				return nil
			},
		},
		{
			name:        "Missing command parameter",
			params:      map[string]interface{}{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := &ClaudeCodeTool{
				Name:       "run_command",
				Parameters: tt.params,
			}

			result, err := toolManager.ExecuteTool(ctx, tool)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("Expected result to be returned")
				}
				if tt.checkResult != nil {
					if err := tt.checkResult(result); err != nil {
						t.Error(err)
					}
				}
			}
		})
	}
}

func TestClaudeCodeToolManager_ConvertToClaudeAPITools(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	toolManager := client.Tools()

	// Convert tools
	apiTools := toolManager.ConvertToClaudeAPITools()

	if len(apiTools) == 0 {
		t.Error("Expected API tools to be generated")
	}

	// Check specific tool conversion
	var readFileTool *types.Tool
	for i := range apiTools {
		if apiTools[i].Name == "read_file" {
			readFileTool = &apiTools[i]
			break
		}
	}

	if readFileTool == nil {
		t.Fatal("Expected read_file tool to be converted")
	}

	// Check tool properties
	if readFileTool.Description == "" {
		t.Error("Expected tool description to be set")
	}

	if readFileTool.InputSchema.Type != "object" {
		t.Errorf("Expected input schema type to be 'object', got %s", readFileTool.InputSchema.Type)
	}

	if len(readFileTool.InputSchema.Properties) == 0 {
		t.Error("Expected input schema to have properties")
	}

	// Check path property
	pathProp, exists := readFileTool.InputSchema.Properties["path"]
	if !exists {
		t.Error("Expected 'path' property in read_file tool")
	}

	if pathProp.Type != "string" {
		t.Errorf("Expected path property type to be 'string', got %s", pathProp.Type)
	}

	// Check required parameters
	if len(readFileTool.InputSchema.Required) != 1 || readFileTool.InputSchema.Required[0] != "path" {
		t.Errorf("Expected required parameters to be ['path'], got %v", readFileTool.InputSchema.Required)
	}
}

func TestClaudeCodeToolManager_HandleToolUse(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file
	testContent := "Test content for HandleToolUse"
	testFile := filepath.Join(tempDir, "test_handle.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	toolManager := client.Tools()

	tests := []struct {
		name        string
		toolUse     *types.ToolUse
		expectError bool
		checkResult func(*types.ToolResult) error
	}{
		{
			name: "Handle read_file tool use",
			toolUse: &types.ToolUse{
				ID:   "tool-use-123",
				Name: "read_file",
				Input: map[string]interface{}{
					"path": "test_handle.txt",
				},
			},
			expectError: false,
			checkResult: func(result *types.ToolResult) error {
				if result.IsError {
					return fmt.Errorf("Expected success, got error")
				}
				if result.ToolUseID != "tool-use-123" {
					return fmt.Errorf("Expected tool use ID 'tool-use-123', got %s", result.ToolUseID)
				}
				if len(result.Content) == 0 {
					return fmt.Errorf("Expected content in result")
				}
				return nil
			},
		},
		{
			name: "Handle tool with error",
			toolUse: &types.ToolUse{
				ID:   "tool-use-456",
				Name: "read_file",
				Input: map[string]interface{}{
					"path": "non_existent_file.txt",
				},
			},
			expectError: false, // HandleToolUse doesn't return error, it sets IsError in result
			checkResult: func(result *types.ToolResult) error {
				if !result.IsError {
					return fmt.Errorf("Expected error result for non-existent file")
				}
				if result.ToolUseID != "tool-use-456" {
					return fmt.Errorf("Expected tool use ID 'tool-use-456', got %s", result.ToolUseID)
				}
				return nil
			},
		},
		{
			name: "Handle MCP tool use",
			toolUse: &types.ToolUse{
				ID:   "tool-use-789",
				Name: "filesystem:read_file", // MCP server:tool format
				Input: map[string]interface{}{
					"path": "test.txt",
				},
			},
			expectError: false,
			checkResult: func(result *types.ToolResult) error {
				// MCP tools not implemented yet, should return error
				if !result.IsError {
					return fmt.Errorf("Expected error for unimplemented MCP tool")
				}
				return nil
			},
		},
		{
			name:        "Handle nil tool use",
			toolUse:     nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toolManager.HandleToolUse(ctx, tt.toolUse)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("Expected result to be returned")
				}
				if tt.checkResult != nil {
					if err := tt.checkResult(result); err != nil {
						t.Error(err)
					}
				}
			}
		})
	}
}

func TestClaudeCodeToolManager_Permissions(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	// Create tool manager with restricted permissions
	restrictedConfig := &ClaudeCodeToolConfig{
		MaxExecutionTime:      30 * time.Second,
		EnableCaching:         true,
		CacheDuration:         5 * time.Minute,
		AllowFileSystemAccess: false,
		AllowNetworkAccess:    false,
	}

	toolManager := NewClaudeCodeToolManagerWithConfig(client, restrictedConfig)

	// Try to execute file system tool with restricted permissions
	tool := &ClaudeCodeTool{
		Name: "read_file",
		Parameters: map[string]interface{}{
			"path": "test.txt",
		},
	}

	result, err := toolManager.ExecuteTool(ctx, tool)
	if err == nil {
		t.Error("Expected error for restricted file system access")
	}
	if result != nil && result.Success {
		t.Error("Expected failure for restricted file system access")
	}
}

func TestClaudeCodeToolManager_ParameterValidation(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	toolManager := client.Tools()

	tests := []struct {
		name        string
		tool        *ClaudeCodeTool
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid parameters",
			tool: &ClaudeCodeTool{
				Name: "analyze_code",
				Parameters: map[string]interface{}{
					"path":          "main.go",
					"analysis_type": "complexity",
				},
			},
			expectError: false,
		},
		{
			name: "Invalid enum value",
			tool: &ClaudeCodeTool{
				Name: "analyze_code",
				Parameters: map[string]interface{}{
					"path":          "main.go",
					"analysis_type": "invalid_type",
				},
			},
			expectError: true,
			errorMsg:    "invalid value",
		},
		{
			name: "Wrong parameter type",
			tool: &ClaudeCodeTool{
				Name: "list_files",
				Parameters: map[string]interface{}{
					"recursive": "yes", // Should be boolean
				},
			},
			expectError: true,
			errorMsg:    "expected boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := toolManager.ExecuteTool(ctx, tt.tool)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !errContains(err, tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to check if error contains message
func errContains(err error, msg string) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), msg)
}
