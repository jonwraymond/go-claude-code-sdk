# Go Claude Code SDK Examples

This directory contains comprehensive examples demonstrating all functionality exposed by the Go Claude Code SDK. Each example is designed to test specific features and provide practical usage patterns.

## 📚 Examples Overview

### 01. Basic Query (`01_basic_query/`)
Demonstrates fundamental query operations:
- Simple one-shot queries
- System prompts configuration
- Tool restrictions
- Synchronous queries with `QuerySync`
- Timeout handling

### 02. Streaming Query (`02_streaming_query/`)
Real-time streaming capabilities:
- Streaming responses
- Progress tracking
- Tool usage monitoring
- Token counting
- Response formatting

### 03. Interactive Client (`03_interactive_client/`)
Bidirectional client interactions:
- Client lifecycle management
- Multi-turn conversations
- Stateful coding sessions
- Context management
- Session persistence

### 04. Interrupt Handling (`04_interrupt_handling/`)
Graceful interruption patterns:
- Signal-based interrupts (Ctrl+C)
- Time-based interrupts
- Tool count interrupts
- Conditional interrupts
- Clean shutdown

### 05. Custom Options (`05_custom_options/`)
Comprehensive configuration:
- Tool restrictions
- Permission modes
- MCP server configuration
- System prompts
- Working directory setup
- Token limits

### 06. Error Handling (`06_error_handling/`)
Robust error management:
- Connection errors
- Query errors
- Tool errors
- Permission errors
- Custom error handlers
- Retry logic

### 07. MCP Servers (`07_mcp_servers/`)
Model Context Protocol integration:
- Filesystem server
- GitHub server
- Multiple servers
- Environment variables
- Custom server implementations

### 08. Permission Modes (`08_permission_modes/`)
Security and permission control:
- Default mode
- Accept edits mode
- Bypass permissions mode
- Custom permission policies
- Audit logging

### 09. Multi-Session (`09_multi_session/`)
Session management patterns:
- Parallel sessions
- Session isolation
- Session coordination
- Priority-based execution
- Session persistence

### 10. Context Cancellation (`10_context_cancellation/`)
Advanced context patterns:
- Timeout contexts
- Manual cancellation
- Deadline contexts
- Cascading cancellation
- Context with values
- Graceful shutdown

### 11. Message Types (`11_message_types/`)
Message handling patterns:
- User messages
- Assistant messages
- System messages
- Result messages
- Content blocks (Text, ToolUse, ToolResult)
- Message flow analysis

### 12. Tool Usage (`12_tool_usage/`)
Tool interaction patterns:
- Basic tool usage (Read, Write, Edit, Bash)
- Tool restrictions
- Tool result handling
- Complex workflows
- Error handling
- Custom patterns (chaining, conditional, batch)

### 13. Benchmarks (`13_benchmarks/`)
Performance testing:
- Connection speed
- Query latency
- Throughput testing
- Concurrent clients
- Message processing
- Memory usage
- Context overhead
- Error handling performance

### 14. Real World Applications (`14_real_world/`)
Practical applications:
- Automated code reviewer
- Documentation generator
- Test generator
- Bug analyzer
- API client generator
- Performance optimizer

## 🚀 Running Examples

### Prerequisites
1. Go 1.20 or later
2. Claude CLI installed and authenticated
3. SDK installed: `go get github.com/jonwraymond/go-claude-code-sdk`

### Run Individual Examples
```bash
# Run basic query examples
go run examples/01_basic_query/main.go

# Run streaming examples
go run examples/02_streaming_query/main.go

# Run interactive client
go run examples/03_interactive_client/main.go

# Run benchmarks
go run examples/13_benchmarks/main.go

# Run integration tests
go test ./tests/... -v
```

### Run All Examples
```bash
# Run all examples sequentially
for dir in examples/*/; do
    if [ -f "$dir/main.go" ]; then
        echo "Running $dir"
        go run "$dir/main.go"
    fi
done
```

## 📝 Code Patterns

### Basic Query Pattern
```go
ctx := context.Background()
msgChan := claudecode.Query(ctx, "What is the capital of France?", nil)

for msg := range msgChan {
    switch m := msg.(type) {
    case *claudecode.AssistantMessage:
        for _, block := range m.Content {
            if textBlock, ok := block.(claudecode.TextBlock); ok {
                fmt.Println(textBlock.Text)
            }
        }
    case *claudecode.ResultMessage:
        fmt.Printf("Completed in %dms\n", m.DurationMs)
    }
}
```

### Interactive Client Pattern
```go
client := claudecode.NewClaudeSDKClient(nil)
defer client.Close()

ctx := context.Background()
if err := client.Connect(ctx); err != nil {
    log.Fatal("Failed to connect:", err)
}

// Send query
if err := client.Query(ctx, "Help me write a function", "session-1"); err != nil {
    log.Fatal("Query failed:", err)
}

// Receive responses
for msg := range client.ReceiveMessages() {
    // Process messages
}
```

### Tool Usage Pattern
```go
options := claudecode.NewClaudeCodeOptions()
options.AllowedTools = []string{"Read", "Write", "Edit"}

msgChan := claudecode.Query(ctx, "Create a config file", options)

for msg := range msgChan {
    if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
        for _, block := range assistantMsg.Content {
            if toolUse, ok := block.(claudecode.ToolUseBlock); ok {
                fmt.Printf("Tool: %s\n", toolUse.Name)
            }
        }
    }
}
```

### Error Handling Pattern
```go
messages, err := claudecode.QuerySync(ctx, "Complex task", options)
if err != nil {
    var cliErr *pkgerrors.CLIConnectionError
    if errors.As(err, &cliErr) {
        // Handle CLI connection error
        log.Printf("CLI error: %v", cliErr)
    }
    return err
}
```

## 🔧 Configuration Options

### Environment Variables
- `CLAUDE_CLI_PATH`: Custom path to Claude CLI
- `CLAUDE_API_KEY`: API key for authentication
- `CLAUDE_LOG_LEVEL`: Logging verbosity (debug, info, warn, error)

### Common Options
```go
options := claudecode.NewClaudeCodeOptions()
options.SystemPrompt = claudecode.StringPtr("You are a helpful coding assistant")
options.MaxTurns = claudecode.IntPtr(5)
options.AllowedTools = []string{"Read", "Write"}
options.PermissionMode = &claudecode.PermissionModeAcceptEdits
options.CWD = claudecode.StringPtr("/tmp/workspace")
```

## 🧪 Testing

### Run Unit Tests
```bash
go test ./pkg/... -v
```

### Run Integration Tests
```bash
go test ./tests/... -v --tags=integration
```

### Run Benchmarks
```bash
go test -bench=. ./examples/13_benchmarks/...
```

## 📊 Performance Tips

1. **Reuse Clients**: Create one client and reuse for multiple queries
2. **Use Contexts**: Leverage context for timeouts and cancellation
3. **Stream Responses**: Use streaming for real-time feedback
4. **Batch Operations**: Group related queries when possible
5. **Monitor Resources**: Use the benchmark examples to profile your usage

## 🐛 Debugging

### Enable Debug Logging
```go
os.Setenv("CLAUDE_LOG_LEVEL", "debug")
```

### Common Issues

1. **Connection Failed**: Ensure Claude CLI is installed and authenticated
2. **Permission Denied**: Check file permissions and allowed tools
3. **Timeout Errors**: Increase timeout values for complex queries
4. **Memory Issues**: Monitor goroutine leaks with `runtime.NumGoroutine()`

## 📚 Additional Resources

- [Main SDK Documentation](../README.md)
- [API Reference](../docs/API.md)
- [Claude CLI Documentation](https://claude.ai/docs/cli)
- [Go Best Practices](https://go.dev/doc/effective_go)

## 🤝 Contributing

To add new examples:
1. Create a new directory with descriptive name
2. Include a `main.go` with clear example code
3. Add comments explaining key concepts
4. Update this README with the example description
5. Test thoroughly before submitting

## 📄 License

These examples are part of the Go Claude Code SDK and follow the same license terms.