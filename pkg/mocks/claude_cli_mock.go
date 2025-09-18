// Package mocks provides mock implementations for testing the Go Claude Code SDK.
package mocks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// ClaudeCLIMock represents a mock implementation of the Claude CLI for testing.
type ClaudeCLIMock struct {
	mu sync.RWMutex

	// Configuration
	Responses      map[string]interface{} // Map of command patterns to responses
	DefaultDelay   time.Duration          // Default delay for responses
	FailureRate    float32                // Probability of command failure (0-1)
	StreamingDelay time.Duration          // Delay between streaming chunks

	// State tracking
	CommandHistory []CommandRecord
	Sessions       map[string]*SessionState
	OpenStreams    map[string]*StreamState

	// Behavior flags
	SimulateErrors     bool
	SimulateTimeouts   bool
	RecordInteractions bool
}

// CommandRecord represents a recorded command interaction.
type CommandRecord struct {
	Timestamp time.Time
	Command   string
	Args      []string
	Env       []string
	Response  interface{}
	Error     error
	Duration  time.Duration
}

// SessionState tracks the state of a mock session.
type SessionState struct {
	ID        string
	CreatedAt time.Time
	Messages  []types.Message
	Model     string
	Active    bool
}

// StreamState tracks the state of a streaming response.
type StreamState struct {
	ID       string
	Messages []string
	Position int
	Done     bool
	Error    error
}

// NewClaudeCLIMock creates a new Claude CLI mock with default configuration.
func NewClaudeCLIMock() *ClaudeCLIMock {
	return &ClaudeCLIMock{
		Responses:          make(map[string]interface{}),
		Sessions:           make(map[string]*SessionState),
		OpenStreams:        make(map[string]*StreamState),
		DefaultDelay:       10 * time.Millisecond,
		StreamingDelay:     5 * time.Millisecond,
		RecordInteractions: true,
	}
}

// SetResponse sets a mock response for a specific command pattern.
func (m *ClaudeCLIMock) SetResponse(pattern string, response interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Responses[pattern] = response
}

// SetStreamingResponse sets a streaming response for testing streaming APIs.
func (m *ClaudeCLIMock) SetStreamingResponse(pattern string, chunks []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	streamID := fmt.Sprintf("stream_%d", time.Now().UnixNano())
	m.OpenStreams[streamID] = &StreamState{
		ID:       streamID,
		Messages: chunks,
		Position: 0,
		Done:     false,
	}
	m.Responses[pattern] = streamID
}

// ExecuteCommand simulates executing a Claude CLI command.
func (m *ClaudeCLIMock) ExecuteCommand(ctx context.Context, args ...string) ([]byte, error) {
	start := time.Now()

	// Record the command if enabled
	if m.RecordInteractions {
		defer func() {
			m.mu.Lock()
			m.CommandHistory = append(m.CommandHistory, CommandRecord{
				Timestamp: start,
				Command:   "claude",
				Args:      args,
				Duration:  time.Since(start),
			})
			m.mu.Unlock()
		}()
	}

	// Simulate delay
	if m.DefaultDelay > 0 {
		select {
		case <-time.After(m.DefaultDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Check for timeout simulation
	if m.SimulateTimeouts && len(args) > 0 && strings.Contains(args[0], "timeout") {
		<-ctx.Done()
		return nil, context.DeadlineExceeded
	}

	// Find matching response
	m.mu.RLock()
	defer m.mu.RUnlock()

	commandStr := strings.Join(args, " ")

	// Look for exact match first
	if response, ok := m.Responses[commandStr]; ok {
		return m.formatResponse(response)
	}

	// Look for pattern match
	for pattern, response := range m.Responses {
		if strings.Contains(commandStr, pattern) {
			return m.formatResponse(response)
		}
	}

	// Handle specific command types
	switch {
	case contains(args, "--query") || contains(args, "-q"):
		return m.handleQueryCommand(ctx, args)
	case contains(args, "--session"):
		return m.handleSessionCommand(ctx, args)
	case contains(args, "--mode") && contains(args, "mcp"):
		return m.handleMCPCommand(ctx, args)
	default:
		return m.handleDefaultCommand(ctx, args)
	}
}

// handleQueryCommand handles mock query commands.
func (m *ClaudeCLIMock) handleQueryCommand(ctx context.Context, args []string) ([]byte, error) {
	// Extract session ID if provided
	sessionID := extractFlag(args, "--session", "default-session")

	// Create or get session
	session, ok := m.Sessions[sessionID]
	if !ok {
		session = &SessionState{
			ID:        sessionID,
			CreatedAt: time.Now(),
			Model:     "claude-3-5-sonnet-20241022",
			Active:    true,
		}
		m.Sessions[sessionID] = session
	}

	// Check if streaming
	if contains(args, "--stream") {
		return m.handleStreamingQuery(ctx, args, session)
	}

	// Return standard query response
	response := types.QueryResponse{
		ID:   fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Type: "message",
		Role: types.RoleAssistant,
		Content: []types.ContentBlock{
			{
				Type: "text",
				Text: "This is a mock response from the Claude CLI mock.",
			},
		},
		Model:      session.Model,
		StopReason: "end_turn",
		Usage: &types.TokenUsage{
			InputTokens:  10,
			OutputTokens: 15,
			TotalTokens:  25,
		},
		CreatedAt: time.Now(),
	}

	return json.Marshal(response)
}

// handleStreamingQuery handles streaming query responses.
func (m *ClaudeCLIMock) handleStreamingQuery(ctx context.Context, args []string, session *SessionState) ([]byte, error) {
	// Create streaming events
	events := []string{
		`{"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant","model":"claude-3-5-sonnet-20241022"}}`,
		`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello "}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"from "}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"mock!"}}`,
		`{"type":"content_block_stop","index":0}`,
		`{"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":5}}`,
		`{"type":"message_stop"}`,
	}

	// Return events as newline-delimited JSON
	return []byte(strings.Join(events, "\n")), nil
}

// handleSessionCommand handles session management commands.
func (m *ClaudeCLIMock) handleSessionCommand(ctx context.Context, args []string) ([]byte, error) {
	// Extract action (list, create, delete, etc.)
	if contains(args, "list") {
		var sessions []map[string]interface{}
		for id, session := range m.Sessions {
			sessions = append(sessions, map[string]interface{}{
				"id":        id,
				"createdAt": session.CreatedAt,
				"active":    session.Active,
			})
		}
		if sessions == nil {
			sessions = []map[string]interface{}{} // Return empty array instead of null
		}
		return json.Marshal(sessions)
	}

	return []byte(`{"success": true}`), nil
}

// handleMCPCommand handles MCP-related commands.
func (m *ClaudeCLIMock) handleMCPCommand(ctx context.Context, args []string) ([]byte, error) {
	// Return mock MCP server configuration
	mcpConfig := map[string]interface{}{
		"servers": map[string]interface{}{
			"mock-server": map[string]interface{}{
				"command": "mock-mcp-server",
				"args":    []string{"--mock"},
				"enabled": true,
			},
		},
	}
	return json.Marshal(mcpConfig)
}

// handleDefaultCommand handles unmatched commands with a default response.
func (m *ClaudeCLIMock) handleDefaultCommand(ctx context.Context, args []string) ([]byte, error) {
	if m.SimulateErrors {
		return nil, fmt.Errorf("mock error: command not recognized")
	}

	return []byte(`{"status": "ok", "message": "Mock command executed successfully"}`), nil
}

// formatResponse converts various response types to byte array.
func (m *ClaudeCLIMock) formatResponse(response interface{}) ([]byte, error) {
	switch v := response.(type) {
	case string:
		// Check if it's a stream ID
		if stream, ok := m.OpenStreams[v]; ok {
			return m.getNextStreamChunk(stream)
		}
		return []byte(v), nil
	case []byte:
		return v, nil
	case error:
		return nil, v
	default:
		return json.Marshal(v)
	}
}

// getNextStreamChunk returns the next chunk from a streaming response.
func (m *ClaudeCLIMock) getNextStreamChunk(stream *StreamState) ([]byte, error) {
	if stream.Done || stream.Position >= len(stream.Messages) {
		stream.Done = true
		return nil, io.EOF
	}

	chunk := stream.Messages[stream.Position]
	stream.Position++

	// Simulate streaming delay
	if m.StreamingDelay > 0 {
		time.Sleep(m.StreamingDelay)
	}

	return []byte(chunk), nil
}

// Utility functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func extractFlag(args []string, flag, defaultValue string) string {
	for i, arg := range args {
		if arg == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return defaultValue
}

// CreateMockExecutor creates a command executor that uses the mock.
func (m *ClaudeCLIMock) CreateMockExecutor() func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		if name != "claude" {
			// Return real command for non-Claude commands
			return exec.CommandContext(ctx, name, args...)
		}

		// Create a mock command that calls our mock handler
		mockScript := fmt.Sprintf(`#!/bin/bash
echo '%s'
`, m.getMockResponse(args))

		// Write mock script to temp file
		tmpFile, err := os.CreateTemp("", "claude-mock-*.sh")
		if err != nil {
			// Fallback to a basic mock if temp file creation fails
			return exec.CommandContext(ctx, "echo", `{"error": "mock creation failed"}`)
		}
		if _, err := tmpFile.WriteString(mockScript); err != nil {
			_ = tmpFile.Close()           // Ignore error, cleanup attempt
			_ = os.Remove(tmpFile.Name()) // Ignore error, cleanup attempt
			return exec.CommandContext(ctx, "echo", `{"error": "mock script write failed"}`)
		}
		if err := tmpFile.Close(); err != nil {
			_ = os.Remove(tmpFile.Name()) // Ignore error, cleanup attempt
			return exec.CommandContext(ctx, "echo", `{"error": "mock script close failed"}`)
		}
		// #nosec G302 - temporary test script needs execute permission, 0700 is secure (owner only)
		if err := os.Chmod(tmpFile.Name(), 0700); err != nil {
			_ = os.Remove(tmpFile.Name()) // Ignore error, cleanup attempt
			return exec.CommandContext(ctx, "echo", `{"error": "mock script chmod failed"}`)
		}

		// Return command that executes our mock script
		cmd := exec.CommandContext(ctx, tmpFile.Name()) // #nosec G204 - tmpFile.Name() is a controlled temporary file path

		// Clean up temp file after command completes
		go func() {
			<-ctx.Done()
			_ = os.Remove(tmpFile.Name()) // Ignore error, cleanup attempt
		}()

		return cmd
	}
}

// getMockResponse generates a mock response for the given arguments.
func (m *ClaudeCLIMock) getMockResponse(args []string) string {
	response, err := m.ExecuteCommand(context.Background(), args...)
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error())
	}
	return string(response)
}

// Reset clears all mock state.
func (m *ClaudeCLIMock) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CommandHistory = nil
	m.Sessions = make(map[string]*SessionState)
	m.OpenStreams = make(map[string]*StreamState)
	m.Responses = make(map[string]interface{})
}

// GetCommandHistory returns the recorded command history.
func (m *ClaudeCLIMock) GetCommandHistory() []CommandRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history := make([]CommandRecord, len(m.CommandHistory))
	copy(history, m.CommandHistory)
	return history
}
