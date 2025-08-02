package client

import (
	"context"
	"testing"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func TestParseSlashCommand(t *testing.T) {
	// Create a mock command executor
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
	
	executor := NewCommandExecutor(client)

	tests := []struct {
		name     string
		input    string
		expected *Command
		hasError bool
	}{
		{
			name:  "simple read command",
			input: "/read file.go",
			expected: &Command{
				Type: CommandRead,
				Args: []string{"file.go"},
				Options: map[string]interface{}{},
			},
		},
		{
			name:  "search with pattern",
			input: "/search function pattern=*.go",
			expected: &Command{
				Type: CommandSearch,
				Args: []string{"function"},
				Options: map[string]interface{}{
					"pattern": "*.go",
				},
			},
		},
		{
			name:  "analyze with multiple args",
			input: "/analyze main.go performance",
			expected: &Command{
				Type: CommandAnalyze,
				Args: []string{"main.go", "performance"},
				Options: map[string]interface{}{},
			},
		},
		{
			name:     "invalid command - no slash",
			input:    "read file.go",
			hasError: true,
		},
		{
			name:     "empty command",
			input:    "/",
			hasError: true,
		},
		{
			name:  "git status",
			input: "/git-status",
			expected: &Command{
				Type: CommandGitStatus,
				Args: []string{},
				Options: map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.ParseSlashCommand(tt.input)
			
			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if result.Type != tt.expected.Type {
				t.Errorf("Expected command type %s, got %s", tt.expected.Type, result.Type)
			}
			
			if len(result.Args) != len(tt.expected.Args) {
				t.Errorf("Expected %d args, got %d", len(tt.expected.Args), len(result.Args))
			}
			
			for i, arg := range tt.expected.Args {
				if i >= len(result.Args) || result.Args[i] != arg {
					t.Errorf("Expected arg[%d] = %s, got %s", i, arg, result.Args[i])
				}
			}
		})
	}
}

func TestBuildCommandPrompt(t *testing.T) {
	// Create a mock command executor
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
	
	executor := NewCommandExecutor(client)

	tests := []struct {
		name     string
		command  *Command
		expected string
		hasError bool
	}{
		{
			name: "read file command",
			command: &Command{
				Type: CommandRead,
				Args: []string{"main.go"},
			},
			expected: "Please read the file 'main.go'",
		},
		{
			name: "read file with summary",
			command: &Command{
				Type: CommandRead,
				Args: []string{"main.go"},
				Options: map[string]interface{}{
					"summarize": true,
				},
			},
			expected: "Please read the file 'main.go' and provide a summary",
		},
		{
			name: "search with pattern",
			command: &Command{
				Type: CommandSearch,
				Args: []string{"function"},
				Options: map[string]interface{}{
					"pattern": "*.go",
				},
			},
			expected: "Please search for 'function' in the codebase in files matching pattern '*.go'",
		},
		{
			name: "analyze with context",
			command: &Command{
				Type: CommandAnalyze,
				Args: []string{"performance"},
				Context: map[string]interface{}{
					"focus": "memory usage",
				},
			},
			expected: "Please analyze the codebase focusing on 'performance'\n\nAdditional context:\n- focus: memory usage",
		},
		{
			name: "git status",
			command: &Command{
				Type: CommandGitStatus,
			},
			expected: "Please show git status",
		},
		{
			name: "write command - insufficient args",
			command: &Command{
				Type: CommandWrite,
				Args: []string{"file.go"}, // Missing content
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.buildCommandPrompt(tt.command)
			
			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if result != tt.expected {
				t.Errorf("Expected prompt:\n%s\n\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestCommandBuilders(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() *Command
		expected *Command
	}{
		{
			name: "ReadFile basic",
			builder: func() *Command {
				return ReadFile("main.go")
			},
			expected: &Command{
				Type: CommandRead,
				Args: []string{"main.go"},
				Options: map[string]interface{}{},
			},
		},
		{
			name: "ReadFile with summary",
			builder: func() *Command {
				return ReadFile("main.go", WithSummary(true))
			},
			expected: &Command{
				Type: CommandRead,
				Args: []string{"main.go"},
				Options: map[string]interface{}{
					"summarize": true,
				},
			},
		},
		{
			name: "WriteFile basic",
			builder: func() *Command {
				return WriteFile("test.go", "package main")
			},
			expected: &Command{
				Type: CommandWrite,
				Args: []string{"test.go", "package main"},
				Options: map[string]interface{}{},
			},
		},
		{
			name: "AnalyzeCode with depth",
			builder: func() *Command {
				return AnalyzeCode("performance", WithDepth("deep"))
			},
			expected: &Command{
				Type: CommandAnalyze,
				Args: []string{"performance"},
				Options: map[string]interface{}{
					"depth": "deep",
				},
			},
		},
		{
			name: "SearchCode with pattern",
			builder: func() *Command {
				return SearchCode("func main", WithPattern("*.go"))
			},
			expected: &Command{
				Type: CommandSearch,
				Args: []string{"func main"},
				Options: map[string]interface{}{
					"pattern": "*.go",
				},
			},
		},
		{
			name: "GitStatus basic",
			builder: func() *Command {
				return GitStatus()
			},
			expected: &Command{
				Type: CommandGitStatus,
				Options: map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder()
			
			if result.Type != tt.expected.Type {
				t.Errorf("Expected command type %s, got %s", tt.expected.Type, result.Type)
			}
			
			if len(result.Args) != len(tt.expected.Args) {
				t.Errorf("Expected %d args, got %d", len(tt.expected.Args), len(result.Args))
			}
			
			for i, arg := range tt.expected.Args {
				if i >= len(result.Args) || result.Args[i] != arg {
					t.Errorf("Expected arg[%d] = %s, got %s", i, arg, result.Args[i])
				}
			}
			
			for key, expectedValue := range tt.expected.Options {
				actualValue, exists := result.Options[key]
				if !exists {
					t.Errorf("Expected option %s not found", key)
					continue
				}
				if actualValue != expectedValue {
					t.Errorf("Expected option %s = %v, got %v", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestExtractTextContent(t *testing.T) {
	// Create a mock command executor
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
	
	executor := NewCommandExecutor(client)

	tests := []struct {
		name     string
		content  []types.ContentBlock
		expected string
	}{
		{
			name: "single text block",
			content: []types.ContentBlock{
				{Type: "text", Text: "Hello world"},
			},
			expected: "Hello world",
		},
		{
			name: "multiple text blocks",
			content: []types.ContentBlock{
				{Type: "text", Text: "Hello "},
				{Type: "text", Text: "world"},
			},
			expected: "Hello world",
		},
		{
			name: "mixed content types",
			content: []types.ContentBlock{
				{Type: "text", Text: "Before "},
				{Type: "image", Text: "image-data"},
				{Type: "text", Text: "after"},
			},
			expected: "Before after",
		},
		{
			name:     "empty content",
			content:  []types.ContentBlock{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.ExtractTextContent(tt.content)
			if result != tt.expected {
				t.Errorf("Expected text content '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestCommandTypes(t *testing.T) {
	// Test that all command types are properly defined
	commandTypes := []CommandType{
		CommandRead,
		CommandWrite,
		CommandEdit,
		CommandSearch,
		CommandAnalyze,
		CommandExplain,
		CommandRefactor,
		CommandTest,
		CommandDebug,
		CommandGitStatus,
		CommandGitCommit,
		CommandGitDiff,
		CommandGitLog,
		CommandBuild,
		CommandRun,
		CommandInstall,
		CommandClean,
		CommandHistory,
		CommandClear,
		CommandSave,
		CommandLoad,
	}

	for _, cmdType := range commandTypes {
		if string(cmdType) == "" {
			t.Errorf("Command type should not be empty: %v", cmdType)
		}
	}
}

func TestCommandOptionsChaining(t *testing.T) {
	// Test that multiple options can be chained
	cmd := ReadFile("test.go", 
		WithSummary(true),
		WithContext("project", "test-project"),
		WithLimit(10),
	)

	if cmd.Type != CommandRead {
		t.Errorf("Expected command type %s, got %s", CommandRead, cmd.Type)
	}

	if len(cmd.Args) != 1 || cmd.Args[0] != "test.go" {
		t.Errorf("Expected args [test.go], got %v", cmd.Args)
	}

	expectedOptions := map[string]interface{}{
		"summarize": true,
		"limit": 10,
	}

	for key, expectedValue := range expectedOptions {
		actualValue, exists := cmd.Options[key]
		if !exists {
			t.Errorf("Expected option %s not found", key)
			continue
		}
		if actualValue != expectedValue {
			t.Errorf("Expected option %s = %v, got %v", key, expectedValue, actualValue)
		}
	}

	expectedContext := map[string]interface{}{
		"project": "test-project",
	}

	for key, expectedValue := range expectedContext {
		actualValue, exists := cmd.Context[key]
		if !exists {
			t.Errorf("Expected context %s not found", key)
			continue
		}
		if actualValue != expectedValue {
			t.Errorf("Expected context %s = %v, got %v", key, expectedValue, actualValue)
		}
	}
}

// Integration test for command execution (requires mocking or real Claude Code)
func TestExecuteCommand_MockResponse(t *testing.T) {
	// This would need a mock implementation or real Claude Code installation
	// For now, we'll test the command building and parsing logic
	
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
	
	executor := NewCommandExecutor(client)

	// Test that we can build a command and generate a prompt
	cmd := ReadFile("main.go", WithSummary(true))
	
	prompt, err := executor.buildCommandPrompt(cmd)
	if err != nil {
		t.Fatalf("Failed to build command prompt: %v", err)
	}

	expectedPrompt := "Please read the file 'main.go' and provide a summary"
	if prompt != expectedPrompt {
		t.Errorf("Expected prompt '%s', got '%s'", expectedPrompt, prompt)
	}

	// Test slash command parsing and execution preparation
	slashCmd := "/read main.go summarize=true"
	parsedCmd, err := executor.ParseSlashCommand(slashCmd)
	if err != nil {
		t.Fatalf("Failed to parse slash command: %v", err)
	}

	if parsedCmd.Type != CommandRead {
		t.Errorf("Expected command type %s, got %s", CommandRead, parsedCmd.Type)
	}

	if len(parsedCmd.Args) != 1 || parsedCmd.Args[0] != "main.go" {
		t.Errorf("Expected args [main.go], got %v", parsedCmd.Args)
	}

	if summarize, ok := parsedCmd.Options["summarize"].(string); !ok || summarize != "true" {
		t.Errorf("Expected summarize option to be 'true', got %v", parsedCmd.Options["summarize"])
	}
}