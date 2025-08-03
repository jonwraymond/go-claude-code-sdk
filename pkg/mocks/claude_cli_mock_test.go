package mocks

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeCLIMock_BasicFunctionality(t *testing.T) {
	mock := NewClaudeCLIMock()

	// Test setting and getting responses
	mock.SetResponse("test-command", "test response")

	response, err := mock.ExecuteCommand(context.Background(), "test-command")
	require.NoError(t, err)
	assert.Equal(t, "test response", string(response))
}

func TestClaudeCLIMock_QueryCommand(t *testing.T) {
	mock := NewClaudeCLIMock()

	// Execute a query command
	response, err := mock.ExecuteCommand(context.Background(), "--query", "-q", "Hello")
	require.NoError(t, err)

	// Response should be valid JSON
	assert.Contains(t, string(response), `"type":"message"`)
	assert.Contains(t, string(response), `"role":"assistant"`)
	assert.Contains(t, string(response), "mock response")
}

func TestClaudeCLIMock_StreamingResponse(t *testing.T) {
	mock := NewClaudeCLIMock()

	// Execute a streaming query command
	response, err := mock.ExecuteCommand(context.Background(), "--query", "-q", "test", "--stream")
	require.NoError(t, err)

	// Response should contain streaming events as newline-delimited JSON
	responseStr := string(response)

	// Verify it contains the expected streaming event types
	assert.Contains(t, responseStr, `"type":"message_start"`)
	assert.Contains(t, responseStr, `"type":"content_block_start"`)
	assert.Contains(t, responseStr, `"type":"content_block_delta"`)
	assert.Contains(t, responseStr, `"type":"content_block_stop"`)
	assert.Contains(t, responseStr, `"type":"message_stop"`)

	// Verify it contains actual content (split across deltas)
	assert.Contains(t, responseStr, "Hello ")
	assert.Contains(t, responseStr, "from ")
	assert.Contains(t, responseStr, "mock!")

	// Verify events are newline-delimited
	lines := strings.Split(strings.TrimSpace(responseStr), "\n")
	assert.GreaterOrEqual(t, len(lines), 5, "Should have at least 5 streaming events")

	// Verify each line is valid JSON
	for _, line := range lines {
		var event map[string]interface{}
		err := json.Unmarshal([]byte(line), &event)
		assert.NoError(t, err, "Each line should be valid JSON")
	}
}

func TestClaudeCLIMock_SessionCommand(t *testing.T) {
	mock := NewClaudeCLIMock()

	// Test session list command
	response, err := mock.ExecuteCommand(context.Background(), "--session", "list")
	require.NoError(t, err)

	// Should return JSON array
	assert.True(t, strings.HasPrefix(string(response), "[") || string(response) == "[]")
}

func TestClaudeCLIMock_ErrorSimulation(t *testing.T) {
	mock := NewClaudeCLIMock()
	mock.SimulateErrors = true

	// Should return error for unrecognized command
	_, err := mock.ExecuteCommand(context.Background(), "unknown-command")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock error")
}

func TestClaudeCLIMock_CommandHistory(t *testing.T) {
	mock := NewClaudeCLIMock()
	mock.RecordInteractions = true

	// Execute some commands
	mock.ExecuteCommand(context.Background(), "command1")
	mock.ExecuteCommand(context.Background(), "command2", "--flag")

	// Check history
	history := mock.GetCommandHistory()
	assert.Len(t, history, 2)
	assert.Equal(t, "command1", history[0].Args[0])
	assert.Equal(t, "command2", history[1].Args[0])
	assert.Equal(t, "--flag", history[1].Args[1])
}

func TestClaudeCLIMock_Timeout(t *testing.T) {
	mock := NewClaudeCLIMock()
	mock.SimulateTimeouts = true
	mock.DefaultDelay = 100 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := mock.ExecuteCommand(ctx, "timeout-command")
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestClaudeCLIMock_Reset(t *testing.T) {
	mock := NewClaudeCLIMock()

	// Add some data
	mock.SetResponse("test", "response")
	mock.ExecuteCommand(context.Background(), "command")
	mock.Sessions["session1"] = &SessionState{ID: "session1"}

	// Reset
	mock.Reset()

	// Verify everything is cleared
	assert.Empty(t, mock.Responses)
	assert.Empty(t, mock.CommandHistory)
	assert.Empty(t, mock.Sessions)
	assert.Empty(t, mock.OpenStreams)
}
