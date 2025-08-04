# Claude Code SDK for Go

A Go SDK for interacting with Claude Code, providing simple interfaces for querying Claude and handling streaming responses.

## Features

- **Simple Query Interface**: One-shot queries with the `Query()` function
- **Interactive Client**: Bidirectional streaming with `ClaudeSDKClient` 
- **Type Safety**: Strongly typed message and content block structures
- **Tool Support**: Handle tool use and result blocks automatically
- **Error Handling**: Comprehensive error types for different failure modes
- **Context Support**: Full Go context integration for timeouts and cancellation

## Installation

```bash
go get github.com/jonwraymond/go-claude-code-sdk
```

## Quick Start

### Simple Query

```go
package main

import (
    "context"
    "fmt"
    "time"

    claude_code_sdk "github.com/jonwraymond/go-claude-code-sdk"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    msgChan := claude_code_sdk.Query(ctx, "What is 2 + 2?", nil)
    for msg := range msgChan {
        if assistantMsg, ok := msg.(*claude_code_sdk.AssistantMessage); ok {
            for _, block := range assistantMsg.Content {
                if textBlock, ok := block.(claude_code_sdk.TextBlock); ok {
                    fmt.Printf("Claude: %s\n", textBlock.Text)
                }
            }
        }
    }
}
```

### Interactive Client

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    claude_code_sdk "github.com/jonwraymond/go-claude-code-sdk"
)

func main() {
    client := claude_code_sdk.NewClaudeSDKClient(nil)
    defer client.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }

    // Send message
    if err := client.Query(ctx, "What's the capital of France?", "default"); err != nil {
        log.Fatal(err)
    }

    // Receive response
    for msg := range client.ReceiveResponse(ctx) {
        if assistantMsg, ok := msg.(*claude_code_sdk.AssistantMessage); ok {
            for _, block := range assistantMsg.Content {
                if textBlock, ok := block.(claude_code_sdk.TextBlock); ok {
                    fmt.Printf("Claude: %s\n", textBlock.Text)
                }
            }
        } else if resultMsg, ok := msg.(*claude_code_sdk.ResultMessage); ok {
            if resultMsg.TotalCostUSD != nil {
                fmt.Printf("Cost: $%.4f\n", *resultMsg.TotalCostUSD)
            }
        }
    }
}
```

## Configuration

### Options

```go
options := claude_code_sdk.NewClaudeCodeOptions()
options.SystemPrompt = claude_code_sdk.StringPtr("You are a helpful assistant")
options.AllowedTools = []string{"Read", "Write"}
options.MaxTurns = claude_code_sdk.IntPtr(5)
options.PermissionMode = claude_code_sdk.PermissionModePtr(claude_code_sdk.PermissionModeAcceptEdits)
options.SetCWD("/path/to/working/directory")

msgChan := claude_code_sdk.Query(ctx, "Help me with my project", options)
```

### Permission Modes

- `PermissionModeDefault`: CLI prompts for dangerous tools
- `PermissionModeAcceptEdits`: Auto-accept file edits
- `PermissionModeBypassPermissions`: Allow all tools (use with caution)

## API Reference

### Functions

#### `Query(ctx context.Context, prompt string, options *ClaudeCodeOptions) <-chan Message`

One-shot query function for simple, stateless interactions.

- **When to use**: Simple questions, batch processing, automation scripts
- **Returns**: Channel of messages from the conversation
- **Stops**: Automatically after receiving a `ResultMessage`

#### `QuerySync(ctx context.Context, prompt string, options *ClaudeCodeOptions) ([]Message, error)`

Synchronous version of `Query` that collects all messages into a slice.

### Types

#### `ClaudeSDKClient`

Interactive client for bidirectional conversations.

**Methods:**
- `Connect(ctx context.Context) error`
- `Query(ctx context.Context, prompt string, sessionID string) error`
- `ReceiveMessages() <-chan Message`
- `ReceiveResponse(ctx context.Context) <-chan Message`
- `Interrupt() error`
- `Close() error`
- `IsConnected() bool`

#### `ClaudeCodeOptions`

Configuration options for queries and clients.

**Fields:**
- `AllowedTools []string`
- `MaxThinkingTokens int`
- `SystemPrompt *string`
- `PermissionMode *PermissionMode`
- `MaxTurns *int`
- `CWD *string`
- And more...

#### Message Types

- `UserMessage`: User messages
- `AssistantMessage`: Claude's responses with content blocks
- `SystemMessage`: System notifications and metadata
- `ResultMessage`: Conversation results with cost and usage info

#### Content Block Types

- `TextBlock`: Plain text content
- `ToolUseBlock`: Tool invocation requests
- `ToolResultBlock`: Tool execution results

### Error Types

- `ClaudeSDKError`: Base error type
- `CLIConnectionError`: Connection failures
- `CLINotFoundError`: Claude CLI not found
- `ProcessError`: Subprocess execution errors
- `CLIJSONDecodeError`: JSON parsing errors

## Examples

See the `examples/` directory for comprehensive examples:

- `examples/quick_start/`: Basic usage patterns
- `examples/streaming_mode/`: Advanced streaming scenarios

To run examples:

```bash
go run examples/quick_start/main.go
go run examples/streaming_mode/main.go all
```

## When to Use What

### Use `Query()` for:
- Simple one-off questions
- Batch processing of independent prompts  
- Code generation or analysis tasks
- Automated scripts and CI/CD pipelines
- When you know all inputs upfront

### Use `ClaudeSDKClient` for:
- Interactive conversations with follow-ups
- Chat applications or REPL-like interfaces
- When you need to send messages based on responses
- When you need interrupt capabilities
- Long-running sessions with state

## Requirements

- Go 1.20 or later
- Claude CLI installed and accessible in PATH

## License

This project follows the same license as the Python Claude Code SDK.

## Contributing

Contributions are welcome! This SDK aims to maintain exact parity with the Python SDK functionality.