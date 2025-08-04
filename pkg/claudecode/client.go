package claudecode

import (
	"context"
	"os"
	"sync"

	"github.com/jonwraymond/go-claude-code-sdk/internal/adapter"
	"github.com/jonwraymond/go-claude-code-sdk/internal/transport"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// ClaudeSDKClient provides bidirectional, interactive conversations with Claude Code.
//
// This client provides full control over the conversation flow with support
// for streaming, interrupts, and dynamic message sending. For simple one-shot
// queries, consider using the Query() function instead.
//
// Key features:
// - **Bidirectional**: Send and receive messages at any time
// - **Stateful**: Maintains conversation context across messages
// - **Interactive**: Send follow-ups based on responses
// - **Control flow**: Support for interrupts and session management
//
// When to use ClaudeSDKClient:
// - Building chat interfaces or conversational UIs
// - Interactive debugging or exploration sessions
// - Multi-turn conversations with context
// - When you need to react to Claude's responses
// - Real-time applications with user input
// - When you need interrupt capabilities
//
// When to use Query() instead:
// - Simple one-off questions
// - Batch processing of prompts
// - Fire-and-forget automation scripts
// - When all inputs are known upfront
// - Stateless operations
//
// Example - Interactive conversation:
//
//	client := NewClaudeSDKClient(nil)
//	defer client.Close()
//
//	ctx := context.Background()
//	if err := client.Connect(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Send initial message
//	if err := client.Query(ctx, "Let's solve a math problem step by step", "default"); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Receive and process response
//	for msg := range client.ReceiveMessages() {
//	    if assistantMsg, ok := msg.(*AssistantMessage); ok {
//	        for _, block := range assistantMsg.Content {
//	            if textBlock, ok := block.(TextBlock); ok {
//	                if strings.Contains(strings.ToLower(textBlock.Text), "ready") {
//	                    goto next
//	                }
//	            }
//	        }
//	    }
//	}
//
//	next:
//	// Send follow-up based on response
//	client.Query(ctx, "What's 15% of 80?", "default")
//
// Example - With interrupt:
//
//	client := NewClaudeSDKClient(nil)
//	defer client.Close()
//
//	if err := client.Connect(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Start a long task
//	client.Query(ctx, "Count to 1000", "default")
//
//	// Interrupt after 2 seconds
//	go func() {
//	    time.Sleep(2 * time.Second)
//	    client.Interrupt()
//	}()
//
//	// Send new instruction after interrupt
//	client.Query(ctx, "Never mind, what's 2+2?", "default")
type ClaudeSDKClient struct {
	options     *ClaudeCodeOptions
	transport   *transport.SubprocessCLITransport
	mu          sync.RWMutex
	connected   bool
	messageChan chan types.Message
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewClaudeSDKClient creates a new Claude SDK client.
//
// Parameters:
//   - options: Optional configuration (uses default if nil)
//
// Example:
//
//	// Default options
//	client := NewClaudeSDKClient(nil)
//
//	// Custom options
//	options := NewClaudeCodeOptions()
//	options.SystemPrompt = StringPtr("You are a helpful coding assistant")
//	options.AllowedTools = []string{"Read", "Write"}
//	client := NewClaudeSDKClient(options)
func NewClaudeSDKClient(options *ClaudeCodeOptions) *ClaudeSDKClient {
	if options == nil {
		options = NewClaudeCodeOptions()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ClaudeSDKClient{
		options:     options,
		messageChan: make(chan types.Message, 100),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Connect establishes connection to Claude with optional initial prompt.
//
// This creates the subprocess connection and prepares for message exchange.
// If no initial prompt is provided, the connection remains open for interactive use.
//
// Parameters:
//   - ctx: Context for connection timeout and cancellation
//   - initialPrompt: Optional initial prompt to send (can be nil for interactive mode)
//
// Example:
//
//	// Interactive mode (no initial prompt)
//	err := client.Connect(ctx)
//
//	// With initial prompt
//	err := client.Connect(ctx)
func (c *ClaudeSDKClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// Set environment variable for SDK entrypoint
	os.Setenv("CLAUDE_CODE_ENTRYPOINT", "sdk-go-client")

	// Convert options to internal format
	internalOptions := adapter.ConvertToInternalOptions(c.options)

	// Create transport with empty prompt for interactive mode
	c.transport = transport.NewSubprocessCLITransport(nil, internalOptions)

	// Connect
	if err := c.transport.Connect(); err != nil {
		return adapter.ConvertFromInternalError(err)
	}

	c.connected = true

	// Start message processing in background
	go c.processMessages()

	return nil
}

// processMessages processes incoming messages from transport.
func (c *ClaudeSDKClient) processMessages() {
	defer close(c.messageChan)

	for {
		select {
		case <-c.ctx.Done():
			return
		case data, ok := <-c.transport.ReceiveMessages():
			if !ok {
				return
			}

			rawMsg, err := transport.ParseMessage(data)
			if err != nil {
				// Send error as system message
				c.messageChan <- &types.SystemMessage{
					Subtype: "error",
					Data: map[string]interface{}{
						"error":    adapter.ConvertFromInternalError(err).Error(),
						"raw_data": string(data),
					},
				}
				continue
			}

			msg, err := adapter.ParseMessageFromRaw(rawMsg)
			if err != nil {
				// Send error as system message
				c.messageChan <- &types.SystemMessage{
					Subtype: "error",
					Data: map[string]interface{}{
						"error":    err.Error(),
						"raw_data": string(data),
					},
				}
				continue
			}

			c.messageChan <- msg
		}
	}
}

// ReceiveMessages returns a channel of messages from Claude.
//
// This channel will receive all messages from the conversation, including
// UserMessage, AssistantMessage, SystemMessage, and ResultMessage types.
//
// The channel is closed when the client is disconnected.
//
// Example:
//
//	for msg := range client.ReceiveMessages() {
//	    switch m := msg.(type) {
//	    case *AssistantMessage:
//	        for _, block := range m.Content {
//	            if textBlock, ok := block.(TextBlock); ok {
//	                fmt.Println("Claude:", textBlock.Text)
//	            }
//	        }
//	    case *ResultMessage:
//	        fmt.Printf("Cost: $%.4f\n", *m.TotalCostUSD)
//	        return // Conversation ended
//	    }
//	}
func (c *ClaudeSDKClient) ReceiveMessages() <-chan types.Message {
	return c.messageChan
}

// Query sends a new request in streaming mode.
//
// This sends a message to Claude and allows for immediate response processing.
// Unlike the standalone Query function, this maintains conversation state.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - prompt: The message to send to Claude
//   - sessionID: Session identifier for the conversation (defaults to "default")
//
// Example:
//
//	err := client.Query(ctx, "What's the weather like?", "session1")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (c *ClaudeSDKClient) Query(ctx context.Context, prompt string, sessionID string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return adapter.ConvertFromInternalError(transport.NewCLIConnectionError("Not connected. Call Connect() first.", nil))
	}

	if sessionID == "" {
		sessionID = "default"
	}

	message := map[string]interface{}{
		"type":               "user",
		"message":            map[string]interface{}{"role": "user", "content": prompt},
		"parent_tool_use_id": nil,
		"session_id":         sessionID,
	}

	err := c.transport.SendRequest([]map[string]interface{}{message}, map[string]interface{}{"session_id": sessionID})
	return adapter.ConvertFromInternalError(err)
}

// ReceiveResponse receives messages until and including a ResultMessage.
//
// This is a convenience method that returns messages one at a time and
// automatically stops after receiving a ResultMessage. It's useful for
// single-response workflows.
//
// The ResultMessage IS included in the returned messages.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//
// Returns a channel that yields messages and closes after a ResultMessage.
//
// Example:
//
//	client.Query(ctx, "What's the capital of France?", "default")
//
//	for msg := range client.ReceiveResponse(ctx) {
//	    if assistantMsg, ok := msg.(*AssistantMessage); ok {
//	        for _, block := range assistantMsg.Content {
//	            if textBlock, ok := block.(TextBlock); ok {
//	                fmt.Printf("Claude: %s\n", textBlock.Text)
//	            }
//	        }
//	    } else if resultMsg, ok := msg.(*ResultMessage); ok {
//	        if resultMsg.TotalCostUSD != nil {
//	            fmt.Printf("Cost: $%.4f\n", *resultMsg.TotalCostUSD)
//	        }
//	        // Channel will close after this message
//	    }
//	}
func (c *ClaudeSDKClient) ReceiveResponse(ctx context.Context) <-chan types.Message {
	responseChan := make(chan types.Message, 10)

	go func() {
		defer close(responseChan)

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-c.ReceiveMessages():
				if !ok {
					return
				}

				select {
				case responseChan <- msg:
				case <-ctx.Done():
					return
				}

				// Stop after result message
				if _, isResult := msg.(*types.ResultMessage); isResult {
					return
				}
			}
		}
	}()

	return responseChan
}

// Interrupt sends interrupt signal to Claude CLI.
//
// This can be used to stop long-running operations and send new instructions.
// Interrupts only work when messages are being actively consumed.
//
// Example:
//
//	// Start a long task
//	client.Query(ctx, "Count from 1 to 1000", "default")
//
//	// Interrupt after some time
//	time.Sleep(2 * time.Second)
//	if err := client.Interrupt(); err != nil {
//	    log.Printf("Failed to interrupt: %v", err)
//	}
//
//	// Send new instruction
//	client.Query(ctx, "Never mind, tell me a joke", "default")
func (c *ClaudeSDKClient) Interrupt() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return adapter.ConvertFromInternalError(transport.NewCLIConnectionError("Not connected. Call Connect() first.", nil))
	}

	err := c.transport.Interrupt()
	return adapter.ConvertFromInternalError(err)
}

// Close disconnects from Claude and cleans up resources.
//
// This should be called when done with the client to ensure proper cleanup.
// It's safe to call multiple times.
//
// Example:
//
//	client := NewClaudeSDKClient(nil)
//	defer client.Close()
//
//	// Use client...
//
//	// Explicit close (optional due to defer)
//	if err := client.Close(); err != nil {
//	    log.Printf("Failed to close client: %v", err)
//	}
func (c *ClaudeSDKClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.connected = false
	c.cancel() // Cancel context to stop background processing

	if c.transport != nil {
		err := c.transport.Disconnect()
		return adapter.ConvertFromInternalError(err)
	}

	return nil
}

// IsConnected returns whether the client is currently connected.
//
// Example:
//
//	if client.IsConnected() {
//	    client.Query(ctx, "Hello", "default")
//	} else {
//	    client.Connect(ctx)
//	}
func (c *ClaudeSDKClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}