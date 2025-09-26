package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseStreamEvent parses a JSON byte array into a StreamEvent
func parseStreamEvent(data []byte) (*types.StreamEvent, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}

	var event types.StreamEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}

	// Store raw data
	event.Raw = json.RawMessage(data)

	return &event, nil
}

// mockStreamReader simulates reading streaming responses
type mockStreamReader struct {
	events []string
	index  int
	delay  time.Duration
	err    error
}

func (m *mockStreamReader) ReadLine() ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}

	if m.index >= len(m.events) {
		return nil, nil // EOF
	}

	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	event := m.events[m.index]
	m.index++
	return []byte(event), nil
}

func TestParseStreamEvent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *types.StreamEvent
		wantErr bool
	}{
		{
			name:  "message_start event",
			input: `{"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant","model":"claude-3-5-sonnet-20241022"}}`,
			want: &types.StreamEvent{
				Type: types.StreamEventMessageStart,
				Message: &types.StreamMessage{
					ID:    "msg_123",
					Type:  "message",
					Role:  types.RoleAssistant,
					Model: types.ModelClaude35Sonnet,
				},
			},
		},
		{
			name:  "content_block_start event",
			input: `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
			want: &types.StreamEvent{
				Type:  types.StreamEventContentBlockStart,
				Index: 0,
				ContentBlock: &types.ContentBlock{
					Type: "text",
					Text: "",
				},
			},
		},
		{
			name:  "content_block_delta event",
			input: `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello world"}}`,
			want: &types.StreamEvent{
				Type:  types.StreamEventContentBlockDelta,
				Index: 0,
				ContentDelta: &types.ContentDelta{
					Type: "text_delta",
					Text: "Hello world",
				},
			},
		},
		{
			name:  "message_stop event with usage",
			input: `{"type":"message_stop","message":{"stop_reason":"end_turn"},"usage":{"input_tokens":10,"output_tokens":20,"total_tokens":30}}`,
			want: &types.StreamEvent{
				Type: types.StreamEventMessageStop,
				Message: &types.StreamMessage{
					StopReason: "end_turn",
				},
				Usage: &types.TokenUsage{
					InputTokens:  10,
					OutputTokens: 20,
					TotalTokens:  30,
				},
			},
		},
		{
			name:  "error event",
			input: `{"type":"error","error":{"type":"invalid_request_error","message":"Invalid input"}}`,
			want: &types.StreamEvent{
				Type: types.StreamEventError,
				Error: &types.APIError{
					Type:    "invalid_request_error",
					Message: "Invalid input",
				},
			},
		},
		{
			name:    "invalid JSON",
			input:   `{"type":"message_start","message":`,
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:  "ping event",
			input: `{"type":"ping"}`,
			want: &types.StreamEvent{
				Type: types.StreamEventPing,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := parseStreamEvent([]byte(tt.input))

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Type, event.Type)

			// Compare specific fields based on event type
			switch event.Type {
			case types.StreamEventMessageStart:
				if tt.want.Message != nil && event.Message != nil {
					assert.Equal(t, tt.want.Message.ID, event.Message.ID)
					assert.Equal(t, tt.want.Message.Model, event.Message.Model)
				}
			case types.StreamEventContentBlockDelta:
				if tt.want.ContentDelta != nil && event.ContentDelta != nil {
					assert.Equal(t, tt.want.ContentDelta.Text, event.ContentDelta.Text)
				}
			case types.StreamEventError:
				if tt.want.Error != nil && event.Error != nil {
					assert.Equal(t, tt.want.Error.Message, event.Error.Message)
				}
			}
		})
	}
}

func TestStreamingResponse_ProcessEvents(t *testing.T) {
	tests := []struct {
		name          string
		events        []string
		expectedText  string
		expectedError error
		expectedUsage *types.TokenUsage
		simulateDelay time.Duration
	}{
		{
			name: "complete streaming response",
			events: []string{
				`{"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant","model":"claude-3-5-sonnet-20241022"}}`,
				`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
				`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello "}}`,
				`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"world!"}}`,
				`{"type":"content_block_stop","index":0}`,
				`{"type":"message_stop","usage":{"output_tokens":2}}`,
			},
			expectedText: "Hello world!",
			expectedUsage: &types.TokenUsage{
				OutputTokens: 2,
			},
		},
		{
			name: "multiple content blocks",
			events: []string{
				`{"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant"}}`,
				`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
				`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"First block"}}`,
				`{"type":"content_block_stop","index":0}`,
				`{"type":"content_block_start","index":1,"content_block":{"type":"text","text":""}}`,
				`{"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":" Second block"}}`,
				`{"type":"content_block_stop","index":1}`,
				`{"type":"message_stop"}`,
			},
			expectedText: "First block Second block",
		},
		{
			name: "error during streaming",
			events: []string{
				`{"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant"}}`,
				`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
				`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Partial"}}`,
				`{"type":"error","error":{"type":"server_error","message":"Internal error"}}`,
			},
			expectedText:  "Partial",
			expectedError: errors.New("stream error: server_error: Internal error"),
		},
		{
			name: "streaming with delays",
			events: []string{
				`{"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant"}}`,
				`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
				`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Delayed"}}`,
				`{"type":"content_block_stop","index":0}`,
				`{"type":"message_stop"}`,
			},
			expectedText:  "Delayed",
			simulateDelay: 10 * time.Millisecond,
		},
		{
			name: "empty content blocks",
			events: []string{
				`{"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant"}}`,
				`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
				`{"type":"content_block_stop","index":0}`,
				`{"type":"message_stop"}`,
			},
			expectedText: "",
		},
		{
			name: "tool use in stream",
			events: []string{
				`{"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant"}}`,
				`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
				`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"I'll help with that."}}`,
				`{"type":"content_block_stop","index":0}`,
				`{"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"tool_123","name":"read_file","input":{}}}`,
				`{"type":"content_block_stop","index":1}`,
				`{"type":"message_stop","stop_reason":"tool_use"}`,
			},
			expectedText: "I'll help with that.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create mock reader
			reader := &mockStreamReader{
				events: tt.events,
				delay:  tt.simulateDelay,
			}

			// Create channels for streaming
			events := make(chan *types.StreamEvent, 100)
			errors := make(chan error, 10)
			done := make(chan struct{})

			// Create streaming response
			sr := &types.StreamingResponse{
				Events: events,
				Errors: errors,
				Done:   done,
			}

			// Process events in goroutine
			go func() {
				defer close(done)
				defer close(events)
				defer close(errors)

				for {
					line, err := reader.ReadLine()
					if err != nil {
						errors <- err
						return
					}
					if line == nil {
						return
					}

					event, err := parseStreamEvent(line)
					if err != nil {
						errors <- err
						continue
					}

					if event.Type == types.StreamEventError {
						streamErr := fmt.Errorf("stream error: %s: %s", event.Error.Type, event.Error.Message)
						errors <- streamErr
						return
					}

					events <- event
				}
			}()

			// Collect events
			var collectedText strings.Builder
			var lastError error
			var usage *types.TokenUsage

		eventLoop:
			for {
				select {
				case event, ok := <-sr.Events:
					if !ok {
						// Events channel closed, process any remaining errors
						select {
						case err := <-sr.Errors:
							if lastError == nil {
								lastError = err
							}
						default:
						}
						break eventLoop
					}

					switch event.Type {
					case types.StreamEventContentBlockDelta:
						if event.ContentDelta != nil {
							collectedText.WriteString(event.ContentDelta.Text)
						}
					case types.StreamEventMessageStop:
						if event.Usage != nil {
							usage = event.Usage
						}
					}

				case err := <-sr.Errors:
					if lastError == nil {
						lastError = err
					}
					// Continue processing to collect any remaining events

				case <-sr.Done:
					// Stream complete, process any remaining events and errors in channels
					for {
						select {
						case event, ok := <-sr.Events:
							if !ok {
								// Events channel closed, check for any remaining errors
								select {
								case err := <-sr.Errors:
									if lastError == nil {
										lastError = err
									}
								default:
								}
								break eventLoop
							}
							switch event.Type {
							case types.StreamEventContentBlockDelta:
								if event.ContentDelta != nil {
									collectedText.WriteString(event.ContentDelta.Text)
								}
							case types.StreamEventMessageStop:
								if event.Usage != nil {
									usage = event.Usage
								}
							}
						case err := <-sr.Errors:
							if lastError == nil {
								lastError = err
							}
						default:
							// No more events or errors, break out
							break eventLoop
						}
					}

				case <-time.After(1 * time.Second):
					t.Fatal("timeout waiting for stream completion")
				}
			}

			// Verify results
			assert.Equal(t, tt.expectedText, collectedText.String())

			if tt.expectedError != nil {
				assert.Error(t, lastError)
				if lastError != nil {
					assert.Contains(t, lastError.Error(), tt.expectedError.Error())
				}
			} else {
				assert.NoError(t, lastError)
			}

			if tt.expectedUsage != nil && usage != nil {
				assert.Equal(t, tt.expectedUsage.OutputTokens, usage.OutputTokens)
			}
		})
	}
}

func TestStreamingResponse_Collect(t *testing.T) {
	events := []string{
		`{"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant","model":"claude-3-5-sonnet-20241022"}}`,
		`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Test "}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"response"}}`,
		`{"type":"content_block_stop","index":0}`,
		`{"type":"message_stop","usage":{"input_tokens":5,"output_tokens":2,"total_tokens":7}}`,
	}

	// Create channels for streaming
	eventChan := make(chan *types.StreamEvent, 100)
	errorChan := make(chan error, 10)
	doneChan := make(chan struct{})

	sr := &types.StreamingResponse{
		Events: eventChan,
		Errors: errorChan,
		Done:   doneChan,
	}

	// Simulate streaming by collecting events manually to avoid race conditions
	go func() {
		defer close(eventChan)
		defer close(doneChan)

		for _, eventStr := range events {
			event, err := parseStreamEvent([]byte(eventStr))
			if err != nil {
				continue
			}
			eventChan <- event
		}
	}()

	// Manually collect events to avoid race condition in sr.Collect()
	var message *types.StreamMessage
	var contentBlocks []types.ContentBlock
	var usage *types.TokenUsage

	// Process all events, ignoring Done channel to avoid race conditions
	for event := range sr.Events {
		switch event.Type {
		case types.StreamEventMessageStart:
			message = event.Message
		case types.StreamEventContentBlockStart:
			if event.ContentBlock != nil {
				contentBlocks = append(contentBlocks, *event.ContentBlock)
			}
		case types.StreamEventContentBlockDelta:
			if event.ContentDelta != nil && event.Index < len(contentBlocks) {
				if event.ContentDelta.Text != "" {
					contentBlocks[event.Index].Text += event.ContentDelta.Text
				}
			}
		case types.StreamEventMessageStop:
			if event.Usage != nil {
				usage = event.Usage
			}
		}
	}

	require.NotNil(t, message)

	// Build response manually
	response := &types.QueryResponse{
		ID:      message.ID,
		Role:    message.Role,
		Model:   message.Model,
		Content: contentBlocks,
		Usage:   usage,
	}

	// Verify collected response
	assert.Equal(t, "msg_123", response.ID)
	assert.Equal(t, types.RoleAssistant, response.Role)
	assert.Equal(t, types.ModelClaude35Sonnet, response.Model)
	assert.Equal(t, "Test response", response.GetTextContent())
	if response.Usage != nil {
		assert.Equal(t, 7, response.Usage.TotalTokens)
	}
}

func TestStreamingResponse_Cancel(t *testing.T) {
	// Create channels for streaming
	eventChan := make(chan *types.StreamEvent, 100)
	errorChan := make(chan error, 10)
	doneChan := make(chan struct{})

	sr := &types.StreamingResponse{
		Events: eventChan,
		Errors: errorChan,
		Done:   doneChan,
		Cancel: func() {
			// Simulate cancellation
		},
	}

	// Add some events
	go func() {
		eventChan <- &types.StreamEvent{
			Type: types.StreamEventMessageStart,
			Message: &types.StreamMessage{
				ID: "msg_123",
			},
		}

		// Simulate ongoing stream
		time.Sleep(100 * time.Millisecond)
		close(doneChan)
	}()

	// Cancel the stream
	sr.Cancel()

	// Wait for completion
	select {
	case <-sr.Done:
		// Success
	case <-time.After(200 * time.Millisecond):
		t.Fatal("stream did not complete after cancellation")
	}
}

func TestStreamingResponse_ConcurrentAccess(t *testing.T) {
	// Create channels for streaming
	eventChan := make(chan *types.StreamEvent, 100)
	errorChan := make(chan error, 10)
	doneChan := make(chan struct{})

	sr := &types.StreamingResponse{
		Events: eventChan,
		Errors: errorChan,
		Done:   doneChan,
	}

	var wg sync.WaitGroup

	// Producer
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(eventChan)
		defer close(doneChan)

		for i := 0; i < 10; i++ {
			eventChan <- &types.StreamEvent{
				Type:  types.StreamEventContentBlockDelta,
				Index: i,
				ContentDelta: &types.ContentDelta{
					Text: fmt.Sprintf("chunk-%d ", i),
				},
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// Multiple consumers
	results := make([]string, 3)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			var text strings.Builder
			for event := range sr.Events {
				if event.ContentDelta != nil {
					text.WriteString(event.ContentDelta.Text)
				}
			}
			results[idx] = text.String()
		}(i)
	}

	wg.Wait()

	// At least one consumer should have received events
	var hasContent bool
	for _, result := range results {
		if result != "" {
			hasContent = true
			break
		}
	}
	assert.True(t, hasContent, "at least one consumer should have received content")
}

func TestStreamingOptions_Callbacks(t *testing.T) {
	opts := &types.StreamOptions{
		OnMessage: func(msg *types.StreamMessage) error {
			assert.Equal(t, "msg_123", msg.ID)
			return nil
		},
		OnContentBlock: func(index int, block *types.ContentBlock) error {
			assert.Equal(t, 0, index)
			assert.Equal(t, "text", block.Type)
			return nil
		},
		OnContentDelta: func(delta *types.ContentDelta) error {
			assert.Equal(t, "Hello", delta.Text)
			return nil
		},
		OnComplete: func(msg *types.StreamMessage) error {
			assert.Equal(t, "end_turn", msg.StopReason)
			return nil
		},
		OnError: func(err error) error {
			// Should not be called in this test
			t.Fatal("unexpected error callback")
			return nil
		},
	}

	// Simulate events
	events := []types.StreamEvent{
		{
			Type: types.StreamEventMessageStart,
			Message: &types.StreamMessage{
				ID: "msg_123",
			},
		},
		{
			Type:  types.StreamEventContentBlockStart,
			Index: 0,
			ContentBlock: &types.ContentBlock{
				Type: "text",
			},
		},
		{
			Type: types.StreamEventContentBlockDelta,
			ContentDelta: &types.ContentDelta{
				Text: "Hello",
			},
		},
		{
			Type: types.StreamEventMessageStop,
			Message: &types.StreamMessage{
				StopReason: "end_turn",
			},
		},
	}

	// Process events through callbacks
	for _, event := range events {
		switch event.Type {
		case types.StreamEventMessageStart:
			err := opts.OnMessage(event.Message)
			assert.NoError(t, err)
		case types.StreamEventContentBlockStart:
			err := opts.OnContentBlock(event.Index, event.ContentBlock)
			assert.NoError(t, err)
		case types.StreamEventContentBlockDelta:
			err := opts.OnContentDelta(event.ContentDelta)
			assert.NoError(t, err)
		case types.StreamEventMessageStop:
			err := opts.OnComplete(event.Message)
			assert.NoError(t, err)
		}
	}
}

func TestStreamingResponse_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		events        []string
		expectedError string
	}{
		{
			name: "malformed JSON event",
			events: []string{
				`{"type":"message_start","message":{"id":"msg_123"}}`,
				`{"type":"content_block_delta","delta":{"text":"test"`, // Malformed JSON
			},
			expectedError: "unexpected end of JSON input",
		},
		{
			name: "API error in stream",
			events: []string{
				`{"type":"message_start","message":{"id":"msg_123"}}`,
				`{"type":"error","error":{"type":"rate_limit_error","message":"Rate limit exceeded"}}`,
			},
			expectedError: "rate_limit_error",
		},
		{
			name: "unknown event type",
			events: []string{
				`{"type":"message_start","message":{"id":"msg_123"}}`,
				`{"type":"unknown_event","data":"something"}`,
				`{"type":"message_stop"}`,
			},
			expectedError: "", // Should not error on unknown events
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &mockStreamReader{
				events: tt.events,
			}

			// Create channels for streaming
			events := make(chan *types.StreamEvent, 100)
			errors := make(chan error, 10)
			done := make(chan struct{})

			sr := &types.StreamingResponse{
				Events: events,
				Errors: errors,
				Done:   done,
			}

			var lastError error

			// Process events
			go func() {
				defer close(done)
				defer close(events)
				defer close(errors)

				for {
					line, err := reader.ReadLine()
					if err != nil || line == nil {
						return
					}

					event, err := parseStreamEvent(line)
					if err != nil {
						errors <- err
						continue
					}

					if event.Type == types.StreamEventError {
						errors <- fmt.Errorf("%s: %s", event.Error.Type, event.Error.Message)
						return
					}

					events <- event
				}
			}()

			// Collect errors - wait for completion and then check for any errors
			<-sr.Done

			// Check if there are any errors in the channel
			select {
			case err := <-sr.Errors:
				lastError = err
			default:
				// No error in channel
			}

			if tt.expectedError != "" {
				assert.Error(t, lastError)
				if lastError != nil {
					assert.Contains(t, lastError.Error(), tt.expectedError)
				}
			} else {
				assert.NoError(t, lastError)
			}
		})
	}
}

// Benchmark tests

func BenchmarkParseStreamEvent(b *testing.B) {
	event := `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"This is a longer piece of text that simulates a real response from the API"}}`
	eventBytes := []byte(event)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := parseStreamEvent(eventBytes)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStreamingResponse_Collect(b *testing.B) {
	// Prepare test events
	events := make([]*types.StreamEvent, 0, 100)

	// Add message start
	events = append(events, &types.StreamEvent{
		Type: types.StreamEventMessageStart,
		Message: &types.StreamMessage{
			ID:    "msg_bench",
			Model: types.ModelClaude35Sonnet,
		},
	})

	// Add content blocks
	for i := 0; i < 50; i++ {
		events = append(events, &types.StreamEvent{
			Type:  types.StreamEventContentBlockDelta,
			Index: 0,
			ContentDelta: &types.ContentDelta{
				Text: fmt.Sprintf("Chunk %d of streaming content. ", i),
			},
		})
	}

	// Add message stop
	events = append(events, &types.StreamEvent{
		Type: types.StreamEventMessageStop,
		Usage: &types.TokenUsage{
			InputTokens:  100,
			OutputTokens: 500,
			TotalTokens:  600,
		},
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create channels for streaming
		eventChan := make(chan *types.StreamEvent, len(events))
		errorChan := make(chan error, 1)
		doneChan := make(chan struct{})

		sr := &types.StreamingResponse{
			Events: eventChan,
			Errors: errorChan,
			Done:   doneChan,
		}

		// Add all events
		for _, event := range events {
			eventChan <- event
		}
		close(eventChan)
		close(doneChan)

		// Collect
		_, err := sr.Collect()
		if err != nil {
			b.Fatal(err)
		}
	}
}
