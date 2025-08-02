package types

import (
	"encoding/json"
	"time"
)

// StreamEventType represents the type of streaming event
type StreamEventType string

const (
	// StreamEventMessage indicates a message event
	StreamEventMessage StreamEventType = "message"
	// StreamEventMessageStart indicates the start of a message
	StreamEventMessageStart StreamEventType = "message_start"
	// StreamEventContentBlockStart indicates the start of a content block
	StreamEventContentBlockStart StreamEventType = "content_block_start"
	// StreamEventContentBlockDelta indicates a content block update
	StreamEventContentBlockDelta StreamEventType = "content_block_delta"
	// StreamEventContentBlockStop indicates the end of a content block
	StreamEventContentBlockStop StreamEventType = "content_block_stop"
	// StreamEventMessageDelta indicates a message update
	StreamEventMessageDelta StreamEventType = "message_delta"
	// StreamEventMessageStop indicates the end of a message
	StreamEventMessageStop StreamEventType = "message_stop"
	// StreamEventPing indicates a keep-alive ping
	StreamEventPing StreamEventType = "ping"
	// StreamEventError indicates an error occurred
	StreamEventError StreamEventType = "error"
)

// StreamEvent represents a single event in a streaming response
type StreamEvent struct {
	// Type indicates the event type
	Type StreamEventType `json:"type"`

	// Index is the index of the content block (for content events)
	Index int `json:"index,omitempty"`

	// Message contains message data (for message events)
	Message *StreamMessage `json:"message,omitempty"`

	// ContentBlock contains content block data (for content block events)
	ContentBlock *ContentBlock `json:"content_block,omitempty"`

	// MessageDelta contains message-level incremental updates (for message_delta events)
	MessageDelta *MessageDelta `json:"message_delta,omitempty"`
	
	// ContentDelta contains content-level incremental updates (for content_block_delta events)
	ContentDelta *ContentDelta `json:"content_delta,omitempty"`

	// Error contains error information (for error events)
	Error *APIError `json:"error,omitempty"`

	// Usage contains token usage information (for message_stop events)
	Usage *TokenUsage `json:"usage,omitempty"`

	// Timestamp is when the event was generated
	Timestamp time.Time `json:"timestamp,omitempty"`

	// Raw contains the raw JSON data for the event
	Raw json.RawMessage `json:"-"`
}

// StreamMessage represents a message in a streaming response
type StreamMessage struct {
	// ID is the message identifier
	ID string `json:"id"`

	// Type is the message type (usually "message")
	Type string `json:"type"`

	// Role is the message role (usually "assistant")
	Role Role `json:"role"`

	// Content contains the message content blocks
	Content []ContentBlock `json:"content,omitempty"`

	// Model indicates which model is generating the response
	Model string `json:"model"`

	// StopReason indicates why the message ended (in message_stop)
	StopReason string `json:"stop_reason,omitempty"`

	// StopSequence contains the stop sequence that ended generation
	StopSequence string `json:"stop_sequence,omitempty"`

	// Usage contains token usage information
	Usage *TokenUsage `json:"usage,omitempty"`
}

// ContentDelta represents incremental content updates in streaming
type ContentDelta struct {
	// Type indicates the delta type (e.g., "text_delta")
	Type string `json:"type,omitempty"`
	
	// Text contains incremental text updates
	Text string `json:"text,omitempty"`
}

// StreamOptions contains options for streaming responses
type StreamOptions struct {
	// OnMessage is called for each complete message
	OnMessage func(*StreamMessage) error

	// OnContentBlock is called for each complete content block
	OnContentBlock func(index int, block *ContentBlock) error

	// OnContentDelta is called for each incremental content update
	OnContentDelta func(delta *ContentDelta) error
	
	// OnMessageDelta is called for message-level updates (stop reason, usage)
	OnMessageDelta func(delta *MessageDelta) error

	// OnError is called when an error occurs
	OnError func(error) error

	// OnComplete is called when streaming completes
	OnComplete func(*StreamMessage) error

	// BufferSize sets the size of the event channel buffer
	BufferSize int

	// IncludeRawEvents indicates whether to include raw JSON in events
	IncludeRawEvents bool
}

// DefaultStreamOptions returns default streaming options
func DefaultStreamOptions() *StreamOptions {
	return &StreamOptions{
		BufferSize: 100,
	}
}

// StreamReader provides an interface for reading streaming events
type StreamReader interface {
	// Next returns the next event in the stream
	Next() (*StreamEvent, error)

	// Close closes the stream reader
	Close() error
}

// StreamingResponse represents a streaming response from the API
type StreamingResponse struct {
	// Events is a channel of streaming events
	Events <-chan *StreamEvent

	// Errors is a channel of errors
	Errors <-chan error

	// Done indicates when streaming is complete
	Done <-chan struct{}

	// Cancel cancels the streaming operation
	Cancel func()

	// reader is the underlying stream reader
	reader StreamReader
}

// Collect waits for the stream to complete and returns the full message
func (sr *StreamingResponse) Collect() (*QueryResponse, error) {
	var message *StreamMessage
	var contentBlocks []ContentBlock
	var lastError error

	for {
		select {
		case event, ok := <-sr.Events:
			if !ok {
				// Channel closed
				if message == nil {
					return nil, lastError
				}
				return sr.buildQueryResponse(message, contentBlocks), nil
			}

			switch event.Type {
			case StreamEventMessageStart:
				message = event.Message
			case StreamEventContentBlockStart:
				if event.ContentBlock != nil {
					contentBlocks = append(contentBlocks, *event.ContentBlock)
				}
			case StreamEventContentBlockDelta:
				if event.ContentDelta != nil && event.Index < len(contentBlocks) {
					// Update the content block with delta
					if event.ContentDelta.Text != "" {
						contentBlocks[event.Index].Text += event.ContentDelta.Text
					}
				}
			case StreamEventMessageStop:
				if event.Message != nil {
					message = event.Message
				}
				if event.Usage != nil && message != nil {
					message.Usage = event.Usage
				}
			}

		case err := <-sr.Errors:
			lastError = err

		case <-sr.Done:
			if message == nil {
				return nil, lastError
			}
			return sr.buildQueryResponse(message, contentBlocks), nil
		}
	}
}

// buildQueryResponse converts streaming data to a QueryResponse
func (sr *StreamingResponse) buildQueryResponse(message *StreamMessage, contentBlocks []ContentBlock) *QueryResponse {
	return &QueryResponse{
		ID:           message.ID,
		Type:         message.Type,
		Role:         message.Role,
		Content:      contentBlocks,
		Model:        message.Model,
		StopReason:   message.StopReason,
		StopSequence: message.StopSequence,
		Usage:        message.Usage,
		CreatedAt:    time.Now(),
	}
}

// StreamCallback is a simple callback function for streaming events
type StreamCallback func(event *StreamEvent) error

// SimpleStreamOptions creates StreamOptions with a single callback
func SimpleStreamOptions(callback StreamCallback) *StreamOptions {
	opts := DefaultStreamOptions()
	
	// Route all events through the single callback
	opts.OnMessage = func(msg *StreamMessage) error {
		return callback(&StreamEvent{
			Type:    StreamEventMessage,
			Message: msg,
		})
	}
	
	opts.OnContentBlock = func(index int, block *ContentBlock) error {
		return callback(&StreamEvent{
			Type:         StreamEventContentBlockStart,
			Index:        index,
			ContentBlock: block,
		})
	}
	
	opts.OnContentDelta = func(delta *ContentDelta) error {
		return callback(&StreamEvent{
			Type:         StreamEventContentBlockDelta,
			ContentDelta: delta,
		})
	}
	
	opts.OnMessageDelta = func(delta *MessageDelta) error {
		return callback(&StreamEvent{
			Type:         StreamEventMessageDelta,
			MessageDelta: delta,
		})
	}
	
	opts.OnError = func(err error) error {
		if apiErr, ok := err.(*APIError); ok {
			return callback(&StreamEvent{
				Type:  StreamEventError,
				Error: apiErr,
			})
		}
		return err
	}
	
	return opts
}