// +build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ToolsIntegrationSuite tests tool system with Claude Code CLI
type ToolsIntegrationSuite struct {
	suite.Suite
	client      *client.ClaudeCodeClient
	toolManager *client.ClaudeCodeToolManager
	config      *types.ClaudeCodeConfig
	testDir     string
}

func (s *ToolsIntegrationSuite) SetupSuite() {
	// Skip if not in integration test mode
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		s.T().Skip("Integration tests disabled. Set INTEGRATION_TESTS=true to run")
	}

	// Ensure API key is set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		s.T().Skip("ANTHROPIC_API_KEY not set")
	}

	// Create test directory
	var err error
	s.testDir, err = os.MkdirTemp("", "claude-code-test-*")
	require.NoError(s.T(), err)

	// Create config
	s.config = types.NewClaudeCodeConfig()
	s.config.APIKey = apiKey
	s.config.ClaudeExecutable = "claude"
	s.config.WorkingDirectory = s.testDir
	s.config.Timeout = 30 * time.Second
	
	// Enable TestMode in CI environment to skip Claude Code CLI requirement
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		s.config.TestMode = true
	}

	// Create client
	ctx := context.Background()
	s.client, err = client.NewClaudeCodeClient(ctx, s.config)
	require.NoError(s.T(), err)

	// Get tool manager
	s.toolManager = s.client.Tools()
}

func (s *ToolsIntegrationSuite) TearDownSuite() {
	if s.client != nil {
		s.client.Close()
	}
	if s.testDir != "" {
		os.RemoveAll(s.testDir)
	}
}

func (s *ToolsIntegrationSuite) TestListTools() {
	ctx := context.Background()

	tools := s.toolManager.ListTools()

	// Should have standard tools
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	// Check for common tools
	assert.True(s.T(), toolNames["read_file"], "Should have read_file tool")
	assert.True(s.T(), toolNames["write_file"], "Should have write_file tool")
	assert.True(s.T(), toolNames["search_code"], "Should have search_code tool")
	assert.True(s.T(), toolNames["run_command"], "Should have run_command tool")
}

func (s *ToolsIntegrationSuite) TestReadWriteFile() {
	ctx := context.Background()

	// Create a test file
	testFile := filepath.Join(s.testDir, "test.txt")
	testContent := "Hello from Claude Code SDK!"

	// Write file using tool
	writeResult, err := s.toolManager.ExecuteTool(ctx, &client.ClaudeCodeTool{
		Name: "write_file",
		Arguments: map[string]interface{}{
			"path":    testFile,
			"content": testContent,
		},
	})
	require.NoError(s.T(), err)
	assert.Contains(s.T(), writeResult, "success")

	// Read file using tool
	readResult, err := s.toolManager.ExecuteTool(ctx, &client.ClaudeCodeTool{
		Name: "read_file",
		Arguments: map[string]interface{}{
			"path": testFile,
		},
	})
	require.NoError(s.T(), err)
	assert.Contains(s.T(), readResult, testContent)
}

func (s *ToolsIntegrationSuite) TestSearchCode() {
	ctx := context.Background()

	// Create test files with code
	goFile := filepath.Join(s.testDir, "main.go")
	goContent := `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}

func helper() {
    fmt.Println("Helper function")
}`

	err := os.WriteFile(goFile, []byte(goContent), 0644)
	require.NoError(s.T(), err)

	// Search for code pattern
	searchResult, err := s.toolManager.ExecuteTool(ctx, &client.ClaudeCodeTool{
		Name: "search_code",
		Arguments: map[string]interface{}{
			"pattern": "func.*\\(",
			"path":    s.testDir,
		},
	})
	require.NoError(s.T(), err)

	// Should find both functions
	assert.Contains(s.T(), searchResult, "main")
	assert.Contains(s.T(), searchResult, "helper")
}

func (s *ToolsIntegrationSuite) TestRunCommand() {
	ctx := context.Background()

	// Run a simple command
	result, err := s.toolManager.ExecuteTool(ctx, &client.ClaudeCodeTool{
		Name: "run_command",
		Arguments: map[string]interface{}{
			"command": "echo 'Hello from command'",
		},
	})
	require.NoError(s.T(), err)
	assert.Contains(s.T(), result, "Hello from command")
}

func (s *ToolsIntegrationSuite) TestToolInConversation() {
	ctx := context.Background()

	// Create a test file
	testFile := filepath.Join(s.testDir, "data.json")
	jsonContent := `{
    "name": "Claude Code SDK",
    "version": "1.0.0",
    "language": "Go"
}`
	err := os.WriteFile(testFile, []byte(jsonContent), 0644)
	require.NoError(s.T(), err)

	// Query that should use tools
	options := &client.QueryOptions{
		AllowedTools: []string{"read_file"},
	}

	result, err := s.client.QueryMessagesSync(ctx, 
		"Read the data.json file in the current directory and tell me what language it uses", 
		options)
	require.NoError(s.T(), err)

	// Should have read the file and understood the content
	assert.Contains(s.T(), result.Content, "Go")
}

func (s *ToolsIntegrationSuite) TestToolPermissions() {
	ctx := context.Background()

	// Test with different permission modes
	testCases := []struct {
		name           string
		permissionMode client.PermissionMode
		expectSuccess  bool
	}{
		{
			name:           "AcceptEdits mode",
			permissionMode: client.PermissionModeAcceptEdits,
			expectSuccess:  true,
		},
		{
			name:           "RejectEdits mode",
			permissionMode: client.PermissionModeRejectEdits,
			expectSuccess:  false,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join(s.testDir, "perm-test.txt")
			
			options := &client.QueryOptions{
				AllowedTools:   []string{"write_file"},
				PermissionMode: tc.permissionMode,
			}

			// Try to write a file
			query := "Create a file called perm-test.txt with the content 'Permission test'"
			result, err := s.client.QueryMessagesSync(ctx, query, options)

			if tc.expectSuccess {
				require.NoError(t, err)
				// File should exist
				_, err = os.Stat(testFile)
				assert.NoError(t, err)
				os.Remove(testFile) // Cleanup
			} else {
				// With RejectEdits, the tool use should be mentioned but not executed
				assert.NotNil(t, result)
				// File should not exist
				_, err = os.Stat(testFile)
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}

func (s *ToolsIntegrationSuite) TestCustomTool() {
	ctx := context.Background()

	// Register a custom tool (this would typically be done via MCP)
	// For this test, we'll simulate by using the existing tool system
	
	// Create a script that acts as a custom tool
	scriptPath := filepath.Join(s.testDir, "custom-tool.sh")
	scriptContent := `#!/bin/bash
echo "Custom tool output: $1"`
	
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(s.T(), err)

	// Execute the custom tool via run_command
	result, err := s.toolManager.ExecuteTool(ctx, &client.ClaudeCodeTool{
		Name: "run_command",
		Arguments: map[string]interface{}{
			"command": scriptPath + " 'test argument'",
		},
	})
	require.NoError(s.T(), err)
	assert.Contains(s.T(), result, "Custom tool output: test argument")
}

func TestToolsIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ToolsIntegrationSuite))
}