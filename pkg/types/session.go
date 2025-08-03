package types

import (
	"time"
)

// SimpleSessionInfo represents basic session information
// This aligns with official Claude Code SDK capabilities
type SimpleSessionInfo struct {
	// ID is the session identifier (UUID format)
	ID string `json:"id"`

	// CreatedAt is when the session was created
	CreatedAt time.Time `json:"created_at"`

	// LastUsedAt is when the session was last accessed
	LastUsedAt time.Time `json:"last_used_at"`

	// Model is the Claude model being used
	Model string `json:"model,omitempty"`

	// Metadata for basic session data (kept minimal)
	Metadata map[string]any `json:"metadata,omitempty"`
}

// SessionMessage represents a basic message in the session
// Simplified to match official SDK message structure
type SessionMessage struct {
	Role        Role           `json:"role"`
	Content     []ContentBlock `json:"content"`
	Timestamp   time.Time      `json:"timestamp"`
	ToolCalls   []ToolCall     `json:"tool_calls,omitempty"`
	ToolResults []ToolResult   `json:"tool_results,omitempty"`
}

// SimpleSessionHistory represents minimal session history
// Only includes what's necessary for basic session tracking
type SimpleSessionHistory struct {
	SessionInfo SimpleSessionInfo `json:"session_info"`
	Messages    []SessionMessage  `json:"messages"`
}
