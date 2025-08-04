package claudecode

import (
	"context"
	"os"

	"github.com/jonwraymond/go-claude-code-sdk/internal/adapter"
	"github.com/jonwraymond/go-claude-code-sdk/internal/transport"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/errors"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// Query performs a one-shot or unidirectional streaming interaction with Claude Code.
//
// This function is ideal for simple, stateless queries where you don't need
// bidirectional communication or conversation management. For interactive,
// stateful conversations, use ClaudeSDKClient instead.
//
// Key differences from ClaudeSDKClient:
// - **Unidirectional**: Send all messages upfront, receive all responses
// - **Stateless**: Each query is independent, no conversation state
// - **Simple**: Fire-and-forget style, no connection management
// - **No interrupts**: Cannot interrupt or send follow-up messages
//
// When to use Query():
// - Simple one-off questions ("What is 2+2?")
// - Batch processing of independent prompts
// - Code generation or analysis tasks
// - Automated scripts and CI/CD pipelines
// - When you know all inputs upfront
//
// When to use ClaudeSDKClient:
// - Interactive conversations with follow-ups
// - Chat applications or REPL-like interfaces
// - When you need to send messages based on responses
// - When you need interrupt capabilities
// - Long-running sessions with state
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - prompt: The prompt to send to Claude (string for simple queries)
//   - options: Optional configuration (uses default if nil)
//
// Returns a channel of Messages from the conversation.
//
// Example - Simple query:
//
//	ctx := context.Background()
//	msgChan := Query(ctx, "What is the capital of France?", nil)
//	for msg := range msgChan {
//	    if assistantMsg, ok := msg.(*AssistantMessage); ok {
//	        for _, block := range assistantMsg.Content {
//	            if textBlock, ok := block.(TextBlock); ok {
//	                fmt.Println("Claude:", textBlock.Text)
//	            }
//	        }
//	    }
//	}
//
// Example - With options:
//
//	options := NewClaudeCodeOptions()
//	options.SystemPrompt = StringPtr("You are an expert Python developer")
//	options.SetCWD("/home/user/project")
//
//	msgChan := Query(ctx, "Create a Python web server", options)
//	for msg := range msgChan {
//	    // Handle messages...
//	}
func Query(ctx context.Context, prompt string, options *ClaudeCodeOptions) <-chan types.Message {
	if options == nil {
		options = NewClaudeCodeOptions()
	}

	// Set environment variable for SDK entrypoint
	os.Setenv("CLAUDE_CODE_ENTRYPOINT", "sdk-go")

	// Create message channel
	msgChan := make(chan types.Message, 100)

	// Process query in goroutine
	go func() {
		defer close(msgChan)

		// Convert options to internal format
		internalOptions := adapter.ConvertToInternalOptions(options)

		// Create transport
		t := transport.NewSubprocessCLITransport(prompt, internalOptions)

		// Connect
		if err := t.Connect(); err != nil {
			// Send error as system message
			msgChan <- &types.SystemMessage{
				Subtype: "error",
				Data: map[string]interface{}{
					"error": adapter.ConvertFromInternalError(err).Error(),
				},
			}
			return
		}

		defer t.Disconnect()

		// Send initial message
		message := map[string]interface{}{
			"type":               "user",
			"message":            map[string]interface{}{"role": "user", "content": prompt},
			"parent_tool_use_id": nil,
			"session_id":         "default",
		}

		if err := t.SendRequest([]map[string]interface{}{message}, map[string]interface{}{"session_id": "default"}); err != nil {
			msgChan <- &types.SystemMessage{
				Subtype: "error",
				Data: map[string]interface{}{
					"error": adapter.ConvertFromInternalError(err).Error(),
				},
			}
			return
		}

		// Receive messages
		for {
			select {
			case <-ctx.Done():
				return
			case data, ok := <-t.ReceiveMessages():
				if !ok {
					return
				}

				rawMsg, err := transport.ParseMessage(data)
				if err != nil {
					// Send error but continue processing
					msgChan <- &types.SystemMessage{
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
					// Send error but continue processing
					msgChan <- &types.SystemMessage{
						Subtype: "error",
						Data: map[string]interface{}{
							"error":    err.Error(),
							"raw_data": string(data),
						},
					}
					continue
				}

				msgChan <- msg

				// Stop after result message
				if _, isResult := msg.(*types.ResultMessage); isResult {
					return
				}
			}
		}
	}()

	return msgChan
}

// QuerySync is a synchronous version of Query that collects all messages.
//
// This is a convenience function that collects all messages from Query
// into a slice and returns them along with any error that occurred.
//
// Example:
//
//	messages, err := QuerySync(ctx, "What is 2+2?", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, msg := range messages {
//	    // Handle messages...
//	}
func QuerySync(ctx context.Context, prompt string, options *ClaudeCodeOptions) ([]types.Message, error) {
	var messages []types.Message
	var lastError error

	msgChan := Query(ctx, prompt, options)
	for msg := range msgChan {
		messages = append(messages, msg)

		// Check for error messages
		if sysMsg, ok := msg.(*types.SystemMessage); ok && sysMsg.Subtype == "error" {
			if errStr, ok := sysMsg.Data["error"].(string); ok {
				lastError = errors.NewClaudeSDKError(errStr, nil)
			}
		}
	}

	return messages, lastError
}

// StringPtr is a helper function to create a pointer to a string.
// This is useful for setting optional string fields in ClaudeCodeOptions.
//
// Example:
//
//	options := NewClaudeCodeOptions()
//	options.SystemPrompt = StringPtr("You are a helpful assistant")
func StringPtr(s string) *string {
	return types.StringPtr(s)
}

// IntPtr is a helper function to create a pointer to an int.
// This is useful for setting optional int fields in ClaudeCodeOptions.
//
// Example:
//
//	options := NewClaudeCodeOptions()
//	options.MaxTurns = IntPtr(5)
func IntPtr(i int) *int {
	return types.IntPtr(i)
}

// PermissionModePtr is a helper function to create a pointer to a PermissionMode.
// This is useful for setting the PermissionMode field in ClaudeCodeOptions.
//
// Example:
//
//	options := NewClaudeCodeOptions()
//	options.PermissionMode = PermissionModePtr(PermissionModeAcceptEdits)
func PermissionModePtr(pm PermissionMode) *PermissionMode {
	result := types.PermissionModePtr(types.PermissionMode(pm))
	// Convert back to local type
	pm2 := PermissionMode(*result)
	return &pm2
}