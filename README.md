# Claude Code SDK for Go

A Go wrapper around the Claude CLI tool, providing type-safe interfaces for automating Claude Code interactions from Go applications.

> **⚠️ Important**: This SDK requires the Claude CLI to be installed and acts as a wrapper around it. It does not provide direct API access to Claude's services - it manages Claude CLI processes and parses their structured output.

## Features

- **CLI Process Management**: Automated spawning and lifecycle management of Claude CLI processes
- **Simple Query Interface**: One-shot CLI queries with the `Query()` function
- **Interactive CLI Client**: Long-running CLI sessions with `ClaudeSDKClient`
- **Type Safety**: Strongly typed parsing of CLI JSON output into Go structures
- **Tool Support**: Handle CLI tool use and result blocks automatically
- **Error Handling**: Comprehensive error types for CLI failures and process management
- **Context Support**: Full Go context integration for CLI process timeouts and cancellation

## Installation

```bash
go get github.com/jonwraymond/go-claude-code-sdk
```

## Prerequisites

⚠️ **Required**: This SDK is a wrapper around the Claude CLI tool. You must have:

1. **Claude CLI installed**: Follow the [Claude CLI installation guide](https://docs.anthropic.com/claude/docs/claude-code)
2. **Claude CLI authenticated**: Run `claude auth` and complete authentication
3. **CLI in PATH**: Ensure the `claude` command is accessible from your shell

```bash
# Verify Claude CLI is installed and working
claude --version
claude auth status
```

This SDK does not communicate directly with Claude's API - it spawns and manages `claude` CLI processes and parses their structured JSON output.

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
    // Create context for the CLI process (30 second timeout)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Query spawns a 'claude' CLI process and streams the parsed JSON output
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
    // Create a client that manages a long-running CLI process
    client := claude_code_sdk.NewClaudeSDKClient(nil)
    defer client.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    // Connect starts a persistent 'claude' CLI session
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

Spawns a Claude CLI process for one-shot queries and streams the parsed JSON responses.

- **Process**: Executes `claude query` command with structured arguments
- **When to use**: Simple questions, batch processing, automation scripts  
- **Returns**: Channel of parsed CLI JSON messages
- **Lifecycle**: CLI process terminates after receiving a `ResultMessage`

#### `QuerySync(ctx context.Context, prompt string, options *ClaudeCodeOptions) ([]Message, error)`

Synchronous version of `Query` that spawns a CLI process and waits for completion.

### Types

#### `ClaudeSDKClient`

Interactive client that manages persistent Claude CLI sessions for bidirectional conversations.

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

- `ClaudeSDKError`: Base error type for all SDK failures
- `CLIConnectionError`: Failed to spawn or communicate with CLI process
- `CLINotFoundError`: Claude CLI executable not found in PATH - ensure CLI is installed
- `ProcessError`: CLI subprocess returned non-zero exit code or crashed
- `CLIJSONDecodeError`: Failed to parse CLI's JSON output format

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
- Simple one-off questions (spawns single CLI process)
- Batch processing of independent prompts  
- Code generation or analysis tasks
- Automated scripts and CI/CD pipelines
- When you know all inputs upfront
- When CLI startup overhead is acceptable

### Use `ClaudeSDKClient` for:
- Interactive conversations with follow-ups (persistent CLI session)
- Chat applications or REPL-like interfaces
- When you need to send messages based on responses
- When you need interrupt capabilities
- Long-running sessions with state
- When minimizing CLI startup overhead is important

### When NOT to use this SDK:
- Direct API access (use official Claude API SDKs instead)
- High-frequency requests (CLI process overhead per request)
- Applications where CLI dependency is problematic
- Environments where Claude CLI cannot be installed

## Requirements

- Go 1.20 or later
- **Claude CLI installed and authenticated** (this is the core dependency)
- Claude CLI executable must be accessible in your system's PATH
- Valid Claude authentication (run `claude auth` to set up)

## Troubleshooting

### Common Issues

**"CLI not found" errors:**
```bash
# Verify Claude CLI is installed and in PATH
which claude
claude --version

# If not found, install Claude CLI first
npm install -g @anthropic-ai/claude-code
```

**Authentication errors:**
```bash
# Check authentication status
claude auth status

# Re-authenticate if needed
claude auth
```

**Process timeout errors:**
- Increase context timeout for complex queries
- Check if CLI is hanging on permission prompts
- Use appropriate `PermissionMode` settings in options
- Verify CLI works independently: `echo "test" | claude query`

**JSON parsing errors:**
- Ensure you're using a compatible Claude CLI version
- Check if CLI output format has changed
- Enable debug logging to inspect raw CLI output

## License

This project follows the same license as the Python Claude Code SDK.

## Contributing

Contributions are welcome! This SDK aims to maintain exact parity with the Python SDK functionality.