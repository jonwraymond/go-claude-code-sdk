# Claude Code Go SDK Examples

This directory contains comprehensive examples demonstrating various features and usage patterns of the Claude Code Go SDK.

## Prerequisites

Before running these examples:

1. Install Claude Code CLI:
   ```bash
   npm install -g @anthropic-ai/claude-code
   ```

2. Set your Anthropic API key:
   ```bash
   export ANTHROPIC_API_KEY="your-api-key-here"
   ```

3. Install the SDK:
   ```bash
   go get github.com/jonwraymond/go-claude-code-sdk
   ```

## Examples Overview

### 📚 [Basic Example](./basic/main.go)
Demonstrates fundamental SDK usage:
- Creating a Claude Code client
- Simple synchronous queries
- Streaming message responses
- Query options configuration

```bash
cd basic
go run main.go
```

### 🔄 [Session Management](./session/main.go)
Shows conversation persistence and context:
- Creating and managing sessions
- Maintaining conversation context
- Project-aware sessions
- Multiple concurrent sessions

```bash
cd session
go run main.go
```

### 🛠️ [Tool System](./tools/main.go)
Explores Claude Code's tool capabilities:
- Listing available tools
- Direct tool execution
- Tool usage in conversations
- Custom tool registration

```bash
cd tools
go run main.go
```

### 🔌 [MCP Server Integration](./mcp/main.go)
Demonstrates Model Context Protocol servers:
- Basic MCP server setup
- Multiple server management
- Common MCP patterns
- Server health monitoring

```bash
cd mcp
go run main.go
```

### 🚀 [Advanced Patterns](./advanced/main.go)
Complex workflows and error handling:
- Comprehensive error handling
- Context cancellation
- Concurrent operations
- Complex code review workflow

```bash
cd advanced
go run main.go
```

## Common Patterns

### Error Handling
```go
result, err := client.QueryMessagesSync(ctx, "query", nil)
if err != nil {
    switch e := err.(type) {
    case *errors.ClaudeCodeError:
        if e.IsRetryable() {
            // Retry logic
        }
    case *errors.ValidationError:
        // Handle validation error
    default:
        // Handle other errors
    }
}
```

### Streaming with Cancellation
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

messages, err := client.QueryMessages(ctx, "query", nil)
if err != nil {
    log.Fatal(err)
}

for msg := range messages {
    select {
    case <-ctx.Done():
        return // Cancelled or timed out
    default:
        // Process message
    }
}
```

### Project-Aware Queries
```go
projectCtx, _ := client.ProjectContext().GetEnhancedProjectContext(ctx)
options := &client.QueryOptions{
    SystemPrompt: fmt.Sprintf("You are helping with a %s project", 
        projectCtx.Language),
}
```

## Tips

1. **API Key Management**: Never hardcode API keys. Use environment variables or secure configuration.

2. **Context Usage**: Always use context for cancellation and timeouts:
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
   defer cancel()
   ```

3. **Resource Cleanup**: Always close clients and sessions:
   ```go
   defer client.Close()
   defer session.Close()
   ```

4. **Error Handling**: Check for specific error types to handle different scenarios appropriately.

5. **Tool Permissions**: Be explicit about allowed tools and permission modes:
   ```go
   options := &client.QueryOptions{
       AllowedTools:   []string{"read_file", "write_file"},
       PermissionMode: client.PermissionModeAsk,
   }
   ```

## Troubleshooting

- **"Claude Code not found"**: Ensure Claude Code CLI is installed and in your PATH
- **"Invalid API key"**: Check that ANTHROPIC_API_KEY is set correctly
- **Timeout errors**: Increase timeout in context or QueryOptions
- **Permission denied**: Check file permissions and PermissionMode settings

## Contributing

See more examples or contribute your own in the [main repository](https://github.com/jonwraymond/go-claude-code-sdk).