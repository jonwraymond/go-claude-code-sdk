package client

import (
	"context"
	"os"
	"testing"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func TestNewClaudeCodeClient(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
		SessionID:        "test-session",
		Model:            "claude-3-5-sonnet-20241022",
		APIKey:           "test-key",
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	if client.workingDir != tempDir {
		t.Errorf("Expected working directory %s, got %s", tempDir, client.workingDir)
	}

	if client.sessionID != "test-session" {
		t.Errorf("Expected session ID 'test-session', got %s", client.sessionID)
	}
}

func TestClaudeCodeClient_ConfigDefaults(t *testing.T) {
	config := &types.ClaudeCodeConfig{}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	// Check that defaults were applied
	if client.config.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected default model 'claude-3-5-sonnet-20241022', got %s", client.config.Model)
	}

	// Working directory should be set to current directory
	currentDir, _ := os.Getwd()
	if client.workingDir != currentDir {
		t.Errorf("Expected working directory to be current directory %s, got %s", currentDir, client.workingDir)
	}

	// Session ID should be auto-generated
	if client.sessionID == "" {
		t.Error("Expected session ID to be auto-generated")
	}
}

func TestClaudeCodeClient_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClaudeCodeConfig
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "non-existent working directory",
			config: &types.ClaudeCodeConfig{
				WorkingDirectory: "/non/existent/directory",
			},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClaudeCodeClient(ctx, tt.config)
			if err == nil {
				t.Error("Expected error for invalid config, got nil")
			}
		})
	}
}

func TestBuildClaudeArgs(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
		SessionID:        "test-session",
		Model:            "claude-3-5-sonnet-20241022",
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	request := &types.QueryRequest{
		Model: "claude-3-opus-20240229",
		Messages: []types.Message{
			{Role: types.RoleUser, Content: "Hello Claude!"},
		},
		MaxTokens:   1000,
		Temperature: 0.5,
		System:      "You are a helpful assistant.",
	}

	args, err := client.buildClaudeArgs(request, false)
	if err != nil {
		t.Fatalf("Failed to build claude arguments: %v", err)
	}

	// Check that arguments contain expected values
	expectedArgs := map[string]bool{
		"--model":       false,
		"--session":     false,
		"--max-tokens":  false,
		"--temperature": false,
		"--system":      false,
	}

	for i, arg := range args {
		if _, exists := expectedArgs[arg]; exists {
			expectedArgs[arg] = true
			// Check that flag has a value (next argument)
			if i+1 >= len(args) {
				t.Errorf("Flag %s has no value", arg)
			}
		}
	}

	for flag, found := range expectedArgs {
		if !found {
			t.Errorf("Expected flag %s not found in arguments", flag)
		}
	}

	// Test streaming args
	streamArgs, err := client.buildClaudeArgs(request, true)
	if err != nil {
		t.Fatalf("Failed to build streaming claude arguments: %v", err)
	}

	// Check that streaming flag is present
	streamFlagFound := false
	for _, arg := range streamArgs {
		if arg == "--stream" {
			streamFlagFound = true
			break
		}
	}
	if !streamFlagFound {
		t.Error("Expected --stream flag in streaming arguments")
	}
}

func TestMessagesToPrompt(t *testing.T) {
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

	tests := []struct {
		name     string
		messages []types.Message
		expected string
	}{
		{
			name:     "empty messages",
			messages: []types.Message{},
			expected: "",
		},
		{
			name: "single user message",
			messages: []types.Message{
				{Role: types.RoleUser, Content: "Hello"},
			},
			expected: "Hello",
		},
		{
			name: "multi-turn conversation",
			messages: []types.Message{
				{Role: types.RoleUser, Content: "Hello"},
				{Role: types.RoleAssistant, Content: "Hi there!"},
				{Role: types.RoleUser, Content: "How are you?"},
			},
			expected: "Human: Hello\n\nAssistant: Hi there!\n\nHuman: How are you?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.messagesToPrompt(tt.messages)
			if err != nil {
				t.Fatalf("messagesToPrompt failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected prompt '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestParseClaudeOutput(t *testing.T) {
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

	tests := []struct {
		name   string
		output string
	}{
		{
			name:   "plain text output",
			output: "This is a simple response from Claude.",
		},
		{
			name:   "multi-line output",
			output: "Line 1\nLine 2\nLine 3",
		},
		{
			name:   "empty output",
			output: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := client.parseClaudeOutput(tt.output)
			if err != nil {
				t.Fatalf("parseClaudeOutput failed: %v", err)
			}

			if len(response.Content) == 0 {
				t.Error("Expected content in response")
			}

			if response.Content[0].Type != "text" {
				t.Errorf("Expected content type 'text', got %s", response.Content[0].Type)
			}

			expectedText := tt.output
			if tt.output != "" {
				expectedText = tt.output // parseClaudeOutput trims whitespace
			}
			if response.Content[0].Text != expectedText {
				t.Errorf("Expected content text '%s', got '%s'", expectedText, response.Content[0].Text)
			}
		})
	}
}

func TestBuildEnvironment(t *testing.T) {
	config := &types.ClaudeCodeConfig{
		APIKey: "test-api-key",
		Environment: map[string]string{
			"CUSTOM_VAR": "custom_value",
			"DEBUG":      "true",
		},
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	env := client.buildEnvironment()

	// Check that API key is included
	apiKeyFound := false
	customVarFound := false
	debugFound := false

	for _, envVar := range env {
		if envVar == "ANTHROPIC_API_KEY=test-api-key" {
			apiKeyFound = true
		}
		if envVar == "CUSTOM_VAR=custom_value" {
			customVarFound = true
		}
		if envVar == "DEBUG=true" {
			debugFound = true
		}
	}

	if !apiKeyFound {
		t.Error("Expected ANTHROPIC_API_KEY in environment")
	}
	if !customVarFound {
		t.Error("Expected CUSTOM_VAR in environment")
	}
	if !debugFound {
		t.Error("Expected DEBUG in environment")
	}
}

func TestFindClaudeCodeCommand(t *testing.T) {
	// Test with custom path (use a known executable for testing)
	customPath := "/bin/echo" // Use echo as a test executable
	cmd, err := findClaudeCodeCommand(customPath)
	if err != nil {
		t.Errorf("findClaudeCodeCommand with valid custom path failed: %v", err)
	}
	if cmd != customPath {
		t.Errorf("Expected command '%s', got '%s'", customPath, cmd)
	}

	// Test with non-existent custom path
	_, err = findClaudeCodeCommand("/non/existent/path")
	if err == nil {
		t.Error("Expected error for non-existent custom path")
	}

	// Test auto-detection (this may fail if claude is not installed)
	_, err = findClaudeCodeCommand("")
	// Don't fail the test if claude is not installed, just log
	if err != nil {
		t.Logf("Auto-detection failed (expected if claude not installed): %v", err)
	}
}

func TestClaudeCodeClient_Close(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}

	// Close the client
	err = client.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify client is closed
	if !client.closed {
		t.Error("Expected client to be closed")
	}

	// Test that operations fail after close
	request := &types.QueryRequest{
		Messages: []types.Message{
			{Role: types.RoleUser, Content: "Hello"},
		},
	}

	_, err = client.Query(ctx, request)
	if err == nil {
		t.Error("Expected error when using closed client")
	}

	_, err = client.QueryStream(ctx, request)
	if err == nil {
		t.Error("Expected error when using closed client for streaming")
	}
}

// Mock test for client operations (since we don't have claude installed in CI)
func TestClaudeCodeClientIntegration(t *testing.T) {
	// Skip this test if CLAUDE_CODE_INTEGRATION_TEST is not set
	if os.Getenv("CLAUDE_CODE_INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test (set CLAUDE_CODE_INTEGRATION_TEST to run)")
	}

	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
		SessionID:        "integration-test",
		APIKey:           os.Getenv("ANTHROPIC_API_KEY"), // Use real API key from environment
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	request := &types.QueryRequest{
		Messages: []types.Message{
			{Role: types.RoleUser, Content: "Say 'Hello, World!'"},
		},
		MaxTokens: 100,
	}

	// Test synchronous query
	response, err := client.Query(ctx, request)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(response.Content) == 0 {
		t.Error("Expected content in response")
	}

	// Test streaming query
	stream, err := client.QueryStream(ctx, request)
	if err != nil {
		t.Fatalf("QueryStream failed: %v", err)
	}
	defer stream.Close()

	// Read first chunk
	chunk, err := stream.Recv()
	if err != nil {
		t.Fatalf("Stream.Recv failed: %v", err)
	}

	if chunk.Done {
		t.Error("Expected content chunk, got done signal immediately")
	}
}
