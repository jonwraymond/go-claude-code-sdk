package client

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	sdkerrors "github.com/jonwraymond/go-claude-code-sdk/pkg/errors"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// StreamQuery creates a new streaming query with advanced options
func (c *ClaudeCodeClient) StreamQuery(ctx context.Context, request *types.QueryRequest, opts *types.StreamOptions) (*types.StreamingResponse, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return nil, sdkerrors.NewInternalError("CLIENT_CLOSED", "client has been closed")
	}
	c.mu.RUnlock()

	if request == nil {
		return nil, sdkerrors.NewValidationError("request", "", "required", "request cannot be nil")
	}

	if opts == nil {
		opts = types.DefaultStreamOptions()
	}

	// Ensure streaming is enabled
	request.Stream = true

	// Build claude command arguments for streaming
	args, err := c.buildClaudeArgs(request, true)
	if err != nil {
		return nil, sdkerrors.WrapError(err, sdkerrors.CategoryValidation, "ARGS_BUILD", "failed to build claude streaming arguments")
	}

	// Create and start claude process
	cmd := exec.CommandContext(ctx, c.claudeCodeCmd, args...)
	cmd.Dir = c.workingDir
	cmd.Env = append(os.Environ(), c.buildEnvironment()...)

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "PIPE_CREATION", "failed to create stdout pipe")
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdout.Close()
		return nil, sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "PIPE_CREATION", "failed to create stderr pipe")
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		return nil, sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "PROCESS_START", "failed to start claude process")
	}

	// Track the process
	processID := fmt.Sprintf("stream-%d", cmd.Process.Pid)
	c.processMu.Lock()
	c.activeProcesses[processID] = cmd
	c.processMu.Unlock()

	// Create channels for streaming
	eventChan := make(chan *types.StreamEvent, opts.BufferSize)
	errorChan := make(chan error, 1)
	doneChan := make(chan struct{})

	// Create cancel context
	streamCtx, cancel := context.WithCancel(ctx)

	// Create reader for parsing stream
	reader := &advancedStreamReader{
		cmd:       cmd,
		stdout:    stdout,
		stderr:    stderr,
		processID: processID,
		client:    c,
		opts:      opts,
	}

	// Start stream processing goroutine
	go reader.processStream(streamCtx, eventChan, errorChan, doneChan)

	return &types.StreamingResponse{
		Events: eventChan,
		Errors: errorChan,
		Done:   doneChan,
		Cancel: cancel,
	}, nil
}

// advancedStreamReader handles the actual stream processing
type advancedStreamReader struct {
	cmd       *exec.Cmd
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	processID string
	client    *ClaudeCodeClient
	opts      *types.StreamOptions
}

// processStream reads from the claude process and sends events to channels
func (r *advancedStreamReader) processStream(ctx context.Context, eventChan chan<- *types.StreamEvent, errorChan chan<- error, doneChan chan<- struct{}) {
	defer close(eventChan)
	defer close(errorChan)
	defer close(doneChan)
	defer r.cleanup()

	// Create scanner for stdout
	scanner := bufio.NewScanner(r.stdout)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	// Read stderr in separate goroutine
	errChan := make(chan error, 1)
	go func() {
		errData, err := io.ReadAll(r.stderr)
		if err != nil {
			errChan <- err
			return
		}
		if len(errData) > 0 {
			errChan <- errors.New(string(errData))
		}
		close(errChan)
	}()

	var currentMessage *types.StreamMessage
	contentBlocks := make([]types.ContentBlock, 0)
	
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		
		// Parse streaming event
		event, err := r.parseStreamEvent(line)
		if err != nil {
			// Skip non-JSON lines
			if !strings.HasPrefix(line, "{") && !strings.HasPrefix(line, "[") {
				continue
			}
			if r.opts.OnError != nil {
				r.opts.OnError(err)
			}
			continue
		}

		// Process event based on type
		switch event.Type {
		case types.StreamEventMessageStart:
			currentMessage = event.Message
			if r.opts.OnMessage != nil {
				r.opts.OnMessage(event.Message)
			}
			
		case types.StreamEventContentBlockStart:
			if event.ContentBlock != nil {
				contentBlocks = append(contentBlocks, *event.ContentBlock)
			}
			
		case types.StreamEventContentBlockDelta:
			if event.ContentDelta != nil && r.opts.OnContentDelta != nil {
				r.opts.OnContentDelta(event.ContentDelta)
			}
			
		case types.StreamEventMessageDelta:
			if event.MessageDelta != nil && r.opts.OnMessageDelta != nil {
				r.opts.OnMessageDelta(event.MessageDelta)
			}
			
		case types.StreamEventContentBlockStop:
			if event.Index < len(contentBlocks) && r.opts.OnContentBlock != nil {
				r.opts.OnContentBlock(event.Index, &contentBlocks[event.Index])
			}
			
		case types.StreamEventMessageStop:
			if currentMessage != nil && r.opts.OnComplete != nil {
				currentMessage.Content = contentBlocks
				if event.Usage != nil {
					currentMessage.Usage = event.Usage
				}
				r.opts.OnComplete(currentMessage)
			}
		}

		// Send event to channel
		select {
		case eventChan <- event:
		case <-ctx.Done():
			return
		}
	}

	// Check for scanner error
	if err := scanner.Err(); err != nil {
		select {
		case errorChan <- err:
		case <-ctx.Done():
		}
	}

	// Check for stderr output
	select {
	case err := <-errChan:
		if err != nil {
			select {
			case errorChan <- err:
			case <-ctx.Done():
			}
		}
	default:
	}
}

// parseStreamEvent parses a line of streaming output into a StreamEvent
func (r *advancedStreamReader) parseStreamEvent(line string) (*types.StreamEvent, error) {
	// Claude CLI streaming format
	// Try to parse as JSON event
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return nil, err
	}

	// Extract event type
	var eventType string
	if typeData, ok := raw["type"]; ok {
		json.Unmarshal(typeData, &eventType)
	}

	event := &types.StreamEvent{
		Type:      types.StreamEventType(eventType),
		Timestamp: time.Now(),
	}

	// Extract index if present
	if indexData, ok := raw["index"]; ok {
		json.Unmarshal(indexData, &event.Index)
	}

	// Parse based on event type
	switch eventType {
	case "message_start":
		if msgData, ok := raw["message"]; ok {
			var msg streamMessage
			if err := json.Unmarshal(msgData, &msg); err == nil {
				event.Message = &types.StreamMessage{
					ID:    msg.ID,
					Type:  msg.Type,
					Role:  types.Role(msg.Role),
					Model: msg.Model,
				}
			}
		}
		
	case "content_block_start":
		if blockData, ok := raw["content_block"]; ok {
			var block streamContentBlock
			if err := json.Unmarshal(blockData, &block); err == nil {
				event.ContentBlock = &types.ContentBlock{
					Type: block.Type,
					Text: block.Text,
				}
			}
		}
		
	case "content_block_delta":
		if deltaData, ok := raw["delta"]; ok {
			var delta streamDelta
			if err := json.Unmarshal(deltaData, &delta); err == nil {
				event.ContentDelta = &types.ContentDelta{
					Type: delta.Type,
					Text: delta.Text,
				}
			}
		}
		
	case "message_delta":
		if deltaData, ok := raw["delta"]; ok {
			var delta streamDelta
			if err := json.Unmarshal(deltaData, &delta); err == nil {
				event.MessageDelta = &types.MessageDelta{
					StopReason:   delta.StopReason,
					StopSequence: delta.StopSequence,
					Usage:        delta.Usage,
				}
			}
		}
		
	case "message_stop":
		// Extract usage if present
		if usageData, ok := raw["usage"]; ok {
			var usage types.TokenUsage
			if err := json.Unmarshal(usageData, &usage); err == nil {
				event.Usage = &usage
			}
		}
		
	case "error":
		if errorData, ok := raw["error"]; ok {
			var errorInfo map[string]any
			if err := json.Unmarshal(errorData, &errorInfo); err == nil {
				event.Error = &types.APIError{
					Type:    getString(errorInfo, "type"),
					Message: getString(errorInfo, "message"),
					Details: errorInfo,
				}
			}
		}
	}

	// Include raw JSON if requested
	if r.opts.IncludeRawEvents {
		event.Raw = json.RawMessage(line)
	}

	return event, nil
}

// cleanup releases resources
func (r *advancedStreamReader) cleanup() {
	if r.stdout != nil {
		r.stdout.Close()
	}
	if r.stderr != nil {
		r.stderr.Close()
	}
	if r.cmd != nil && r.cmd.Process != nil {
		r.cmd.Process.Kill()
		r.cmd.Wait()
	}
	
	// Remove from active processes
	r.client.processMu.Lock()
	delete(r.client.activeProcesses, r.processID)
	r.client.processMu.Unlock()
}

// streamMessage represents message data in stream events
type streamMessage struct {
	ID      string            `json:"id"`
	Type    string            `json:"type"`
	Role    string            `json:"role"`
	Model   string            `json:"model"`
	Content []json.RawMessage `json:"content,omitempty"`
}

// streamContentBlock represents content block data in stream events
type streamContentBlock struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
	Text string `json:"text,omitempty"`
}

// streamDelta represents delta updates in stream events
type streamDelta struct {
	Type         string            `json:"type,omitempty"`
	Text         string            `json:"text,omitempty"`
	StopReason   string            `json:"stop_reason,omitempty"`
	StopSequence string            `json:"stop_sequence,omitempty"`
	Usage        *types.TokenUsage `json:"usage,omitempty"`
}

// getString safely extracts a string from a map
func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}