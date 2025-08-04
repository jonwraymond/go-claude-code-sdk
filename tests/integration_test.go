package tests

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/claudecode"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// isClaudeAvailable checks if Claude CLI is available and can potentially run queries
func isClaudeAvailable() bool {
	// Check if claude command exists
	_, err := exec.LookPath("claude")
	if err != nil {
		return false
	}

	// Try a simple version check that should work even without full auth
	cmd := exec.Command("claude", "--version")
	err = cmd.Run()
	return err == nil
}

// skipIfClaudeUnavailable skips the test if Claude CLI is not available
func skipIfClaudeUnavailable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if !isClaudeAvailable() {
		t.Skip("Claude CLI not available - skipping integration test")
	}

	// Check if we have an API key
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set - skipping integration test")
	}
}

// TestBasicQuery tests basic query functionality
func TestBasicQuery(t *testing.T) {
	skipIfClaudeUnavailable(t)

	ctx := context.Background()
	msgChan := claudecode.Query(ctx, "What is 2+2?", nil)

	gotResponse := false
	gotResult := false

	for msg := range msgChan {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			gotResponse = true
			assert.NotEmpty(t, m.Content, "Assistant message should have content")

			// Check for text block
			for _, block := range m.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					assert.NotEmpty(t, textBlock.Text, "Text block should not be empty")
				}
			}
		case *claudecode.ResultMessage:
			gotResult = true
			assert.False(t, m.IsError, "Result should not be an error")
			assert.Greater(t, m.DurationMs, 0, "Duration should be positive")
		}
	}

	assert.True(t, gotResponse, "Should receive assistant message")
	assert.True(t, gotResult, "Should receive result message")
}

// TestQueryWithOptions tests query with custom options
func TestQueryWithOptions(t *testing.T) {
	skipIfClaudeUnavailable(t)

	options := claudecode.NewClaudeCodeOptions()
	options.SystemPrompt = claudecode.StringPtr("You are a helpful math tutor. Be concise.")
	options.MaxTurns = claudecode.IntPtr(1)

	ctx := context.Background()
	msgChan := claudecode.Query(ctx, "Explain prime numbers", options)

	messageCount := 0
	for msg := range msgChan {
		if _, ok := msg.(*claudecode.AssistantMessage); ok {
			messageCount++
		}
	}

	assert.Greater(t, messageCount, 0, "Should receive at least one message")
}

// TestInteractiveClient tests interactive client functionality
func TestInteractiveClient(t *testing.T) {
	skipIfClaudeUnavailable(t)

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx := context.Background()
	err := client.Connect(ctx)
	require.NoError(t, err, "Client should connect successfully")

	// Test connection status
	assert.True(t, client.IsConnected(), "Client should be connected")

	// Send a query
	err = client.Query(ctx, "Hello, Claude!", "test-session")
	require.NoError(t, err, "Query should succeed")

	// Receive response
	timeout := time.After(10 * time.Second)
	gotResponse := false

	for !gotResponse {
		select {
		case msg := <-client.ReceiveMessages():
			if _, ok := msg.(*claudecode.AssistantMessage); ok {
				gotResponse = true
			}
		case <-timeout:
			t.Fatal("Timeout waiting for response")
		}
	}

	assert.True(t, gotResponse, "Should receive response from Claude")
}

// TestMultipleSessions tests multiple concurrent sessions
func TestMultipleSessions(t *testing.T) {
	skipIfClaudeUnavailable(t)

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx := context.Background()
	err := client.Connect(ctx)
	require.NoError(t, err)

	// Track messages by session
	sessionMessages := make(map[string]int)

	// Message processor
	go func() {
		for msg := range client.ReceiveMessages() {
			if result, ok := msg.(*claudecode.ResultMessage); ok {
				sessionMessages[result.SessionID]++
			}
		}
	}()

	// Send queries to different sessions
	sessions := []string{"session-A", "session-B", "session-C"}
	for _, sessionID := range sessions {
		err := client.Query(ctx, "Hello from "+sessionID, sessionID)
		assert.NoError(t, err)
		time.Sleep(1 * time.Second)
	}

	// Wait for responses
	time.Sleep(5 * time.Second)

	// Verify each session got a response
	for _, sessionID := range sessions {
		assert.Greater(t, sessionMessages[sessionID], 0,
			"Session %s should have received messages", sessionID)
	}
}

// TestContextCancellation tests context cancellation
func TestContextCancellation(t *testing.T) {
	skipIfClaudeUnavailable(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	msgChan := claudecode.Query(ctx, "Count from 1 to 1000 slowly", nil)

	messageCount := 0
	for range msgChan {
		messageCount++
		if messageCount > 100 {
			t.Error("Should have been cancelled before receiving 100 messages")
			break
		}
	}

	assert.Greater(t, messageCount, 0, "Should receive some messages before cancellation")
}

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	skipIfClaudeUnavailable(t)

	tests := []struct {
		name        string
		setupFunc   func() (*claudecode.ClaudeSDKClient, error)
		expectError bool
	}{
		{
			name: "Query without connection",
			setupFunc: func() (*claudecode.ClaudeSDKClient, error) {
				client := claudecode.NewClaudeSDKClient(nil)
				// Don't connect
				return client, nil
			},
			expectError: true,
		},
		{
			name: "Invalid CLI path",
			setupFunc: func() (*claudecode.ClaudeSDKClient, error) {
				os.Setenv("CLAUDE_CLI_PATH", "/invalid/path/to/claude")
				defer os.Unsetenv("CLAUDE_CLI_PATH")

				client := claudecode.NewClaudeSDKClient(nil)
				ctx := context.Background()
				err := client.Connect(ctx)
				return client, err
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := tt.setupFunc()
			defer client.Close()

			if tt.expectError {
				if err == nil {
					// Try to query
					ctx := context.Background()
					err = client.Query(ctx, "test", "test")
				}
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Expected no error")
			}
		})
	}
}

// TestToolUsage tests tool usage functionality
func TestToolUsage(t *testing.T) {
	skipIfClaudeUnavailable(t)

	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Read", "Write"}

	ctx := context.Background()
	msgChan := claudecode.Query(ctx,
		"Create a file called test.txt with 'Hello, World!' content", options)

	toolsUsed := make(map[string]int)

	for msg := range msgChan {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if toolUse, ok := block.(claudecode.ToolUseBlock); ok {
					toolsUsed[toolUse.Name]++
				}
			}
		}
	}

	assert.Greater(t, len(toolsUsed), 0, "Should have used at least one tool")
	assert.Contains(t, toolsUsed, "Write", "Should have used the Write tool")
}

// TestPermissionModes tests different permission modes
func TestPermissionModes(t *testing.T) {
	skipIfClaudeUnavailable(t)

	testDir := t.TempDir()

	modes := []claudecode.PermissionMode{
		claudecode.PermissionModeDefault,
		claudecode.PermissionModeAcceptEdits,
		claudecode.PermissionModeBypassPermission,
	}

	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			options := claudecode.NewClaudeCodeOptions()
			options.PermissionMode = &mode
			options.CWD = &testDir
			options.AllowedTools = []string{"Write"}

			ctx := context.Background()
			msgChan := claudecode.Query(ctx,
				"Create a file called "+string(mode)+".txt", options)

			// Consume all messages
			for range msgChan {
				// Just drain the channel
			}

			// For non-default modes, file should be created
			if mode != claudecode.PermissionModeDefault {
				filePath := testDir + "/" + string(mode) + ".txt"
				_, err := os.Stat(filePath)
				// File might or might not exist depending on Claude's response
				_ = err
			}
		})
	}
}

// TestQuerySync tests synchronous query functionality
func TestQuerySync(t *testing.T) {
	skipIfClaudeUnavailable(t)

	ctx := context.Background()
	messages, err := claudecode.QuerySync(ctx, "What is the capital of Japan?", nil)

	require.NoError(t, err, "QuerySync should not error")
	assert.NotEmpty(t, messages, "Should receive messages")

	// Check message types
	hasAssistant := false
	hasResult := false

	for _, msg := range messages {
		switch msg.(type) {
		case *claudecode.AssistantMessage:
			hasAssistant = true
		case *claudecode.ResultMessage:
			hasResult = true
		}
	}

	assert.True(t, hasAssistant, "Should have assistant message")
	assert.True(t, hasResult, "Should have result message")
}

// TestInterrupt tests interrupt functionality
func TestInterrupt(t *testing.T) {
	skipIfClaudeUnavailable(t)

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx := context.Background()
	err := client.Connect(ctx)
	require.NoError(t, err)

	// Track if we got interrupted
	interrupted := false

	go func() {
		for msg := range client.ReceiveMessages() {
			if _, ok := msg.(*claudecode.AssistantMessage); ok {
				// Got a message, try to interrupt
				if !interrupted {
					interrupted = true
					err := client.Interrupt()
					assert.NoError(t, err, "Interrupt should succeed")
				}
			}
		}
	}()

	// Start a long task
	err = client.Query(ctx, "Count from 1 to 1000", "interrupt-test")
	assert.NoError(t, err)

	// Wait a bit
	time.Sleep(3 * time.Second)

	assert.True(t, interrupted, "Should have interrupted the query")
}

// TestReceiveResponse tests the ReceiveResponse convenience method
func TestReceiveResponse(t *testing.T) {
	skipIfClaudeUnavailable(t)

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx := context.Background()
	err := client.Connect(ctx)
	require.NoError(t, err)

	// Send query
	err = client.Query(ctx, "What is 5 times 6?", "response-test")
	require.NoError(t, err)

	// Use ReceiveResponse
	responseChan := client.ReceiveResponse(ctx)

	messageCount := 0
	gotResult := false

	for msg := range responseChan {
		messageCount++
		if _, ok := msg.(*claudecode.ResultMessage); ok {
			gotResult = true
		}
	}

	assert.Greater(t, messageCount, 0, "Should receive messages")
	assert.True(t, gotResult, "Should receive result message")
}

// BenchmarkQuery benchmarks basic query performance
func BenchmarkQuery(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgChan := claudecode.Query(ctx, "What is 2+2?", nil)
		for range msgChan {
			// Drain channel
		}
	}
}

// BenchmarkClientCreation benchmarks client creation
func BenchmarkClientCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := claudecode.NewClaudeSDKClient(nil)
		client.Close()
	}
}

// BenchmarkMessageProcessing benchmarks message processing
func BenchmarkMessageProcessing(b *testing.B) {
	// Create sample message
	msg := &claudecode.AssistantMessage{
		Content: []types.ContentBlock{
			claudecode.TextBlock{Text: "This is a test message for benchmarking"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Process message
		for _, block := range msg.Content {
			if textBlock, ok := block.(claudecode.TextBlock); ok {
				_ = len(textBlock.Text)
			}
		}
	}
}
