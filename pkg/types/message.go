package types

import (
	"encoding/json"
	"time"
)

// Message represents a single message in a conversation with Claude Code.
// Messages can be from users, Claude, or the system, and may contain various types of content.
//
// Example usage:
//
//	msg := &types.Message{
//		Role:    types.RoleUser,
//		Content: "Hello Claude, can you help me with Go programming?",
//	}
type Message struct {
	// ID is a unique identifier for this message
	ID string `json:"id,omitempty"`

	// Role indicates who sent the message (user, assistant, system)
	Role Role `json:"role"`

	// Content is the main text content of the message
	Content string `json:"content"`

	// ToolCalls contains any tool calls made in this message
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// ToolCallID is used to reference a specific tool call (for tool responses)
	ToolCallID string `json:"tool_call_id,omitempty"`

	// Attachments contains file attachments or other media
	Attachments []Attachment `json:"attachments,omitempty"`

	// Metadata contains additional message information
	Metadata map[string]any `json:"metadata,omitempty"`

	// Timestamp is when the message was created
	Timestamp time.Time `json:"timestamp,omitempty"`

	// TokenCount is the number of tokens in this message (if available)
	TokenCount int `json:"token_count,omitempty"`
}

// GetText returns the text content of the message.
// For messages with content blocks, it concatenates all text blocks.
func (m *Message) GetText() string {
	// For now, just return the Content field
	// In the future, this could handle content blocks
	return m.Content
}

// HasToolUse returns true if the message contains tool calls.
func (m *Message) HasToolUse() bool {
	return len(m.ToolCalls) > 0
}

// GetToolUses returns all tool calls in the message.
func (m *Message) GetToolUses() []ToolCall {
	return m.ToolCalls
}

// Role represents the different roles that can send messages in a conversation.
type Role string

const (
	// RoleUser indicates a message from the user
	RoleUser Role = "user"

	// RoleAssistant indicates a message from Claude
	RoleAssistant Role = "assistant"

	// RoleSystem indicates a system message (instructions, context, etc.)
	RoleSystem Role = "system"

	// RoleTool indicates a response from a tool call
	RoleTool Role = "tool"
)

// Backward compatibility aliases
const (
	MessageRoleUser      = RoleUser
	MessageRoleAssistant = RoleAssistant
	MessageRoleSystem    = RoleSystem
	MessageRoleTool      = RoleTool
)

// IsValid checks if the role is a valid message role.
func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleAssistant, RoleSystem, RoleTool:
		return true
	default:
		return false
	}
}

// ToolCall represents a function call made by Claude during a conversation.
// This allows Claude to use external tools and functions to assist with requests.
type ToolCall struct {
	// ID is a unique identifier for this tool call
	ID string `json:"id"`

	// Type is the type of tool being called (usually "function")
	Type string `json:"type"`

	// Function contains the function call details
	Function FunctionCall `json:"function"`
}

// FunctionCall represents the details of a specific function call.
type FunctionCall struct {
	// Name is the name of the function being called
	Name string `json:"name"`

	// Arguments contains the function arguments as a JSON string
	Arguments string `json:"arguments"`

	// ParsedArguments contains the parsed function arguments (if available)
	ParsedArguments map[string]any `json:"parsed_arguments,omitempty"`
}

// ParseArguments parses the JSON arguments string into a map.
func (f *FunctionCall) ParseArguments() (map[string]any, error) {
	if f.ParsedArguments != nil {
		return f.ParsedArguments, nil
	}

	var args map[string]any
	if err := json.Unmarshal([]byte(f.Arguments), &args); err != nil {
		return nil, err
	}

	f.ParsedArguments = args
	return args, nil
}

// Attachment represents a file or media attachment to a message.
type Attachment struct {
	// ID is a unique identifier for this attachment
	ID string `json:"id,omitempty"`

	// Type indicates the type of attachment (image, file, etc.)
	Type AttachmentType `json:"type"`

	// Name is the original filename or name
	Name string `json:"name,omitempty"`

	// URL is the location where the attachment can be accessed
	URL string `json:"url,omitempty"`

	// Data contains the raw attachment data (for small attachments)
	Data []byte `json:"data,omitempty"`

	// MimeType is the MIME type of the attachment
	MimeType string `json:"mime_type,omitempty"`

	// Size is the size of the attachment in bytes
	Size int64 `json:"size,omitempty"`

	// Metadata contains additional attachment information
	Metadata map[string]any `json:"metadata,omitempty"`
}

// AttachmentType represents the different types of attachments supported.
type AttachmentType string

const (
	// AttachmentTypeImage indicates an image attachment
	AttachmentTypeImage AttachmentType = "image"

	// AttachmentTypeFile indicates a general file attachment
	AttachmentTypeFile AttachmentType = "file"

	// AttachmentTypeCode indicates a code file attachment
	AttachmentTypeCode AttachmentType = "code"

	// AttachmentTypeDocument indicates a document attachment
	AttachmentTypeDocument AttachmentType = "document"
)

// Conversation represents a complete conversation thread with Claude Code.
// It contains all messages and maintains conversation state.
type Conversation struct {
	// ID is the unique identifier for this conversation
	ID string `json:"id"`

	// Title is a human-readable title for the conversation
	Title string `json:"title,omitempty"`

	// Messages contains all messages in the conversation in chronological order
	Messages []Message `json:"messages"`

	// CreatedAt is when the conversation was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the conversation was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// TokenCount is the total token count for the conversation
	TokenCount int `json:"token_count,omitempty"`

	// Metadata contains additional conversation information
	Metadata map[string]any `json:"metadata,omitempty"`
}

// AddMessage adds a new message to the conversation.
func (c *Conversation) AddMessage(msg Message) {
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	c.Messages = append(c.Messages, msg)
	c.UpdatedAt = time.Now()
}

// GetLastMessage returns the most recent message in the conversation.
func (c *Conversation) GetLastMessage() *Message {
	if len(c.Messages) == 0 {
		return nil
	}
	return &c.Messages[len(c.Messages)-1]
}

// GetMessagesByRole returns all messages from a specific role.
func (c *Conversation) GetMessagesByRole(role Role) []Message {
	var messages []Message
	for _, msg := range c.Messages {
		if msg.Role == role {
			messages = append(messages, msg)
		}
	}
	return messages
}

// Event represents a real-time event in the Claude Code system.
// Events are used for streaming responses and system notifications.
type Event struct {
	// ID is a unique identifier for this event
	ID string `json:"id,omitempty"`

	// Type indicates the type of event
	Type EventType `json:"type"`

	// Data contains the event payload
	Data any `json:"data,omitempty"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// Source indicates where the event originated
	Source string `json:"source,omitempty"`

	// ConversationID links the event to a specific conversation
	ConversationID string `json:"conversation_id,omitempty"`
}

// EventType represents the different types of events that can occur.
type EventType string

const (
	// EventTypeMessageStart indicates a new message is starting
	EventTypeMessageStart EventType = "message_start"

	// EventTypeContentBlockStart indicates a content block is starting
	EventTypeContentBlockStart EventType = "content_block_start"

	// EventTypeContentBlockDelta indicates incremental content updates
	EventTypeContentBlockDelta EventType = "content_block_delta"

	// EventTypeContentBlockStop indicates a content block is complete
	EventTypeContentBlockStop EventType = "content_block_stop"

	// EventTypeMessageDelta indicates incremental message updates
	EventTypeMessageDelta EventType = "message_delta"

	// EventTypeMessageStop indicates a message is complete
	EventTypeMessageStop EventType = "message_stop"

	// EventTypeToolCallStart indicates a tool call is starting
	EventTypeToolCallStart EventType = "tool_call_start"

	// EventTypeToolCallDelta indicates incremental tool call updates
	EventTypeToolCallDelta EventType = "tool_call_delta"

	// EventTypeToolCallStop indicates a tool call is complete
	EventTypeToolCallStop EventType = "tool_call_stop"

	// EventTypeError indicates an error occurred
	EventTypeError EventType = "error"

	// EventTypePing is used for connection keep-alive
	EventTypePing EventType = "ping"
)

// MessageStartEvent contains data for message start events.
type MessageStartEvent struct {
	Message Message `json:"message"`
}

// ContentBlockStartEvent contains data for content block start events.
type ContentBlockStartEvent struct {
	Index        int          `json:"index"`
	ContentBlock ContentBlock `json:"content_block"`
}

// ContentBlockDeltaEvent contains data for content block delta events.
type ContentBlockDeltaEvent struct {
	Index int               `json:"index"`
	Delta ContentBlockDelta `json:"delta"`
}

// ContentBlockStopEvent contains data for content block stop events.
type ContentBlockStopEvent struct {
	Index int `json:"index"`
}

// MessageDeltaEvent contains data for message delta events.
type MessageDeltaEvent struct {
	Delta MessageDelta `json:"delta"`
}

// MessageStopEvent contains data for message stop events.
type MessageStopEvent struct {
	// Empty for now, may contain stop reason or other metadata in the future
}

// ContentBlock represents a block of content within a message.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Data any    `json:"data,omitempty"`

	// Tool-specific fields
	ID    string         `json:"id,omitempty"`    // For tool_use blocks
	Name  string         `json:"name,omitempty"`  // For tool_use blocks
	Input map[string]any `json:"input,omitempty"` // For tool_use blocks

	// Tool result fields
	ToolUseID string         `json:"tool_use_id,omitempty"` // For tool_result blocks
	Content   []ContentBlock `json:"content,omitempty"`     // For tool_result blocks
	IsError   bool           `json:"is_error,omitempty"`    // For tool_result blocks
}

// NewTextBlock creates a new text content block
func NewTextBlock(text string) ContentBlock {
	return ContentBlock{
		Type: "text",
		Text: text,
	}
}

// NewToolUseBlock creates a new tool use content block
func NewToolUseBlock(id, name string, input map[string]any) ContentBlock {
	return ContentBlock{
		Type:  "tool_use",
		ID:    id,
		Name:  name,
		Input: input,
	}
}

// NewToolResultBlock creates a new tool result content block
func NewToolResultBlock(toolUseID string, content []ContentBlock, isError bool) ContentBlock {
	return ContentBlock{
		Type:      "tool_result",
		ToolUseID: toolUseID,
		Content:   content,
		IsError:   isError,
	}
}

// ContentBlockDelta represents incremental changes to a content block.
type ContentBlockDelta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// MessageDelta represents incremental changes to a message.
type MessageDelta struct {
	StopReason   string      `json:"stop_reason,omitempty"`
	StopSequence string      `json:"stop_sequence,omitempty"`
	Usage        *TokenUsage `json:"usage,omitempty"`
}

// TokenUsage contains information about token consumption.
type TokenUsage struct {
	// InputTokens is the number of tokens in the input
	InputTokens int `json:"input_tokens"`

	// OutputTokens is the number of tokens in the output
	OutputTokens int `json:"output_tokens"`

	// TotalTokens is the total token count
	TotalTokens int `json:"total_tokens"`
}

// EventHandler defines the interface for handling streaming events.
// Implementations can process different types of events as they occur.
type EventHandler interface {
	// OnEvent is called when any event is received
	OnEvent(event *Event) error

	// OnMessage is called when a complete message is received
	OnMessage(message *Message) error

	// OnError is called when an error event is received
	OnError(err error) error

	// OnComplete is called when the stream is complete
	OnComplete() error
}

// SimpleEventHandler provides a basic implementation of EventHandler.
// It can be embedded in custom handlers to provide default behavior.
type SimpleEventHandler struct {
	// MessageHandler is called for each complete message
	MessageHandler func(*Message) error

	// ErrorHandler is called for each error
	ErrorHandler func(error) error

	// CompleteHandler is called when streaming is complete
	CompleteHandler func() error
}

// OnEvent provides default event handling.
func (h *SimpleEventHandler) OnEvent(event *Event) error {
	// Default implementation does nothing
	return nil
}

// OnMessage calls the MessageHandler if provided.
func (h *SimpleEventHandler) OnMessage(message *Message) error {
	if h.MessageHandler != nil {
		return h.MessageHandler(message)
	}
	return nil
}

// OnError calls the ErrorHandler if provided.
func (h *SimpleEventHandler) OnError(err error) error {
	if h.ErrorHandler != nil {
		return h.ErrorHandler(err)
	}
	return nil
}

// OnComplete calls the CompleteHandler if provided.
func (h *SimpleEventHandler) OnComplete() error {
	if h.CompleteHandler != nil {
		return h.CompleteHandler()
	}
	return nil
}
