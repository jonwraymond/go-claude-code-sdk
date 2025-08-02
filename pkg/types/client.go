package types

import (
	"context"
)

// Client defines the core interface for interacting with Claude Code.
// It provides both synchronous and streaming query capabilities, as well as
// Claude Code-specific command execution support.
//
// Example usage:
//
//	client := claude.NewClient(ctx, config)
//	response, err := client.Query(ctx, request)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(response.Content)
type Client interface {
	// Query sends a synchronous request to Claude Code and returns the complete response.
	// This method blocks until the full response is received.
	Query(ctx context.Context, request *QueryRequest) (*QueryResponse, error)

	// QueryStream sends a request to Claude Code and returns a streaming response.
	// This method returns immediately with a stream that yields response chunks as they arrive.
	QueryStream(ctx context.Context, request *QueryRequest) (QueryStream, error)

	// Close gracefully shuts down the client and releases any resources.
	// It should be called when the client is no longer needed.
	Close() error
}

// ClaudeCodeClient extends the basic Client interface with Claude Code-specific features.
// This includes command execution, project context management, and session handling.
type ClaudeCodeClient interface {
	Client

	// ExecuteCommand executes a Claude Code command and returns the result.
	ExecuteCommand(ctx context.Context, cmd *Command) (*CommandResult, error)

	// ExecuteSlashCommand executes a Claude Code slash command (e.g., "/read file.go").
	ExecuteSlashCommand(ctx context.Context, slashCommand string) (*CommandResult, error)

	// GetProjectContext returns information about the current project context.
	GetProjectContext(ctx context.Context) (*ProjectContext, error)

	// SetWorkingDirectory changes the working directory for Claude Code operations.
	SetWorkingDirectory(ctx context.Context, path string) error
}

// QueryStream represents a streaming response from the Claude Code API.
// It provides methods to read response chunks and handle streaming events.
//
// Example usage:
//
//	stream, err := client.QueryStream(ctx, request)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer stream.Close()
//
//	for {
//		chunk, err := stream.Recv()
//		if chunk != nil && chunk.Done {
//			break
//		}
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Print(chunk.Content)
//	}
type QueryStream interface {
	// Recv receives the next chunk from the streaming response.
	// Returns an error when the stream encounters an issue, or nil when successful.
	// Check the chunk's Done field to determine if the stream is complete.
	Recv() (*StreamChunk, error)

	// Close closes the stream and releases associated resources.
	Close() error
}

// StreamChunk represents a single chunk of data in a streaming response.
// Each chunk contains partial content and metadata about the streaming state.
type StreamChunk struct {
	// Type indicates the type of this chunk (content, metadata, tool_call, etc.)
	Type ChunkType `json:"type"`

	// Content contains the text content for this chunk (if applicable)
	Content string `json:"content,omitempty"`

	// Delta represents incremental changes to the response state
	Delta *StreamDelta `json:"delta,omitempty"`

	// Metadata contains additional information about this chunk
	Metadata map[string]any `json:"metadata,omitempty"`

	// Done indicates whether this is the final chunk in the stream
	Done bool `json:"done"`
}

// ChunkType represents the different types of chunks that can be received in a stream.
type ChunkType string

const (
	// ChunkTypeContent indicates a chunk containing message content
	ChunkTypeContent ChunkType = "content"

	// ChunkTypeMetadata indicates a chunk containing response metadata
	ChunkTypeMetadata ChunkType = "metadata"

	// ChunkTypeToolCall indicates a chunk containing tool call information
	ChunkTypeToolCall ChunkType = "tool_call"

	// ChunkTypeError indicates a chunk containing error information
	ChunkTypeError ChunkType = "error"

	// ChunkTypeDone indicates the final chunk marking stream completion
	ChunkTypeDone ChunkType = "done"
)

// StreamDelta represents incremental changes in a streaming response.
// This follows the delta format commonly used in streaming APIs.
type StreamDelta struct {
	// Role indicates the role being updated (if applicable)
	Role string `json:"role,omitempty"`

	// Content represents content changes
	Content string `json:"content,omitempty"`

	// ToolCalls represents tool call updates
	ToolCalls []ToolCallDelta `json:"tool_calls,omitempty"`
}

// ToolCallDelta represents incremental updates to tool calls in a streaming response.
type ToolCallDelta struct {
	// Index is the index of the tool call being updated
	Index int `json:"index"`

	// ID is the unique identifier for this tool call
	ID string `json:"id,omitempty"`

	// Type is the type of tool being called
	Type string `json:"type,omitempty"`

	// Function contains updates to function call parameters
	Function *FunctionCallDelta `json:"function,omitempty"`
}

// FunctionCallDelta represents incremental updates to function call parameters.
type FunctionCallDelta struct {
	// Name is the name of the function being called
	Name string `json:"name,omitempty"`

	// Arguments contains incremental updates to function arguments (JSON string)
	Arguments string `json:"arguments,omitempty"`
}

// SessionManager defines the interface for managing conversation sessions.
// Sessions allow for maintaining context across multiple queries.
type SessionManager interface {
	// CreateSession creates a new conversation session with the given configuration.
	CreateSession(ctx context.Context, config *SessionConfig) (*Session, error)

	// GetSession retrieves an existing session by ID.
	GetSession(ctx context.Context, sessionID string) (*Session, error)

	// ListSessions returns all active sessions for the current user.
	ListSessions(ctx context.Context) ([]*Session, error)

	// DeleteSession removes a session and all associated data.
	DeleteSession(ctx context.Context, sessionID string) error
}

// Session represents a conversation session with Claude Code.
// Sessions maintain context and conversation history across multiple queries.
type Session struct {
	// ID is the unique identifier for this session
	ID string `json:"id"`

	// Title is a human-readable title for the session
	Title string `json:"title,omitempty"`

	// CreatedAt is the timestamp when the session was created
	CreatedAt int64 `json:"created_at"`

	// UpdatedAt is the timestamp when the session was last updated
	UpdatedAt int64 `json:"updated_at"`

	// MessageCount is the number of messages in this session
	MessageCount int `json:"message_count"`

	// Metadata contains additional session information
	Metadata map[string]any `json:"metadata,omitempty"`
}

// SessionConfig contains configuration options for creating new sessions.
type SessionConfig struct {
	// Title is an optional title for the session
	Title string `json:"title,omitempty"`

	// Model specifies which Claude model to use for this session
	Model string `json:"model,omitempty"`

	// MaxTokens sets the maximum number of tokens for responses in this session
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature controls randomness in responses (0.0 to 1.0)
	Temperature float64 `json:"temperature,omitempty"`

	// Metadata contains additional configuration options
	Metadata map[string]any `json:"metadata,omitempty"`
}
