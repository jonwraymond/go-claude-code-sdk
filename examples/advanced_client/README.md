# Advanced Client Example

This example demonstrates advanced Claude Code client features including timeout configuration, MCP server setup, lifecycle management, resource monitoring, and custom executable paths.

## What You'll Learn

- Context timeout and cancellation patterns
- MCP (Model Context Protocol) server configuration
- Client lifecycle management best practices
- Resource monitoring and performance optimization
- Cache management and configuration
- Custom Claude Code executable paths

## Code Overview

The example includes five advanced configuration patterns:

### 1. Timeout Configuration
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

claudeClient, err := client.NewClaudeCodeClient(ctx, config)
```

Demonstrates proper context management for client creation and operations with timeouts.

### 2. MCP Server Setup
```go
// List current MCP servers
servers := claudeClient.ListMCPServers()

// Add a specific MCP server
err = claudeClient.AddMCPServer(ctx, "filesystem", &types.MCPServerConfig{
    Command: "npx",
    Args:    []string{"@modelcontextprotocol/server-filesystem", "/tmp"},
})

// Setup common MCP servers automatically
err = claudeClient.SetupCommonMCPServers(ctx)
```

Shows how to configure MCP servers for enhanced Claude Code capabilities.

### 3. Lifecycle Management
```go
useClientSession := func(sessionName string) error {
    claudeClient, err := client.NewClaudeCodeClient(ctx, config)
    if err != nil {
        return err
    }
    
    defer func() {
        if err := claudeClient.Close(); err != nil {
            log.Printf("Error closing client: %v", err)
        }
    }()
    
    // Use client...
    return nil
}
```

Demonstrates proper client lifecycle management with cleanup.

### 4. Resource Monitoring
```go
// Enable debug mode
config.Debug = true

// Test project context caching
start := time.Now()
ctx1, _ := claudeClient.GetProjectContext(ctx)
first := time.Since(start)

start = time.Now()
ctx2, _ := claudeClient.GetProjectContext(ctx)
second := time.Since(start) // Should be much faster (cached)

// Configure cache
claudeClient.SetProjectContextCacheDuration(5 * time.Minute)
claudeClient.InvalidateProjectContextCache()
```

Shows resource monitoring, caching behavior, and performance optimization.

### 5. Custom Claude Code Path
```go
config.ClaudeCodePath = "/usr/local/bin/claude"
// or
config.ClaudeCodePath = "npx" // Use npx to run Claude Code
```

Demonstrates using custom Claude Code executable paths.

## Prerequisites

1. **Go 1.21+** installed
2. **Claude Code CLI** installed: `npm install -g @anthropic/claude-code`
3. **Node.js and npm** (for MCP server examples)
4. **Authentication** configured

## Running the Example

### Setup
```bash
# Install Claude Code CLI
npm install -g @anthropic/claude-code

# Set up authentication (choose one)
export ANTHROPIC_API_KEY="your-api-key"
# OR authenticate with CLI
claude auth

# Install MCP server dependencies (optional, for MCP examples)
npm install -g @modelcontextprotocol/server-filesystem
```

### Run the Example
```bash
cd examples/advanced_client
go run main.go
```

## Expected Output

```
=== Advanced Claude Code Client Examples ===

--- Example 1: Timeout Configuration ---
✓ Client created with 30-second timeout
Executing query with 10-second timeout...
✓ Query completed successfully
  Response preview: Go is a programming language developed by Google...

--- Example 2: MCP Server Configuration ---
✓ Client created for MCP configuration
Current MCP servers: 0 configured
Adding filesystem MCP server...
✓ Filesystem MCP server added
Setting up common MCP servers...
✓ Common MCP servers configured
MCP servers after setup: 3 configured

--- Example 3: Client Lifecycle Management ---
✓ Client created for session: session1
  Working in directory: /current/path
✓ Client for session session1 closed successfully
✓ Client created for session: session2
  Working in directory: /current/path
✓ Client for session session2 closed successfully
✓ Client created for session: session3
  Working in directory: /current/path
✓ Client for session session3 closed successfully

--- Example 4: Resource Monitoring ---
✓ Client created with debug mode enabled
Testing project context cache...
  First call: 45.123ms
  Cached call: 1.234ms
  Same result: true
✓ Project context cache invalidated
✓ Cache duration set to 5 minutes
Cache info:
  cached: true
  cache_duration: 5m0s
  last_updated: 2024-01-15T10:30:45Z

--- Example 5: Custom Claude Code Path ---
Found Claude Code at: /usr/local/bin/claude
✓ Client created with custom Claude Code path
  Path: /usr/local/bin/claude
  Test Mode: true
✓ Project context retrieved successfully
  Working Directory: /current/path
```

## Key Concepts

### Context Management

**Timeouts**: Set appropriate timeouts for different operations
```go
// Client creation timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

// Query timeout
queryCtx, queryCancel := context.WithTimeout(context.Background(), 10*time.Second)
```

**Cancellation**: Proper cleanup when operations are cancelled
```go
if ctx.Err() == context.DeadlineExceeded {
    fmt.Printf("Operation timed out")
}
```

### MCP Server Integration

MCP (Model Context Protocol) servers extend Claude Code's capabilities:

- **Filesystem Server**: File system access
- **Database Server**: Database operations
- **Web Server**: Web scraping and API access
- **Custom Servers**: Your own extensions

### Resource Monitoring

The SDK includes built-in performance monitoring:

- **Project Context Caching**: Automatic caching of project information
- **Cache Management**: Configure cache duration and invalidation
- **Debug Logging**: Detailed operation logging
- **Performance Metrics**: Track operation timing

### Custom Executable Paths

Support for different Claude Code installations:

- **Global Installation**: `/usr/local/bin/claude`
- **NPX**: `npx @anthropic/claude-code`
- **Custom Builds**: `/path/to/custom/claude`

## Advanced Configuration

### MCP Server Configuration
```go
// Add custom MCP server
config := &types.MCPServerConfig{
    Command: "node",
    Args:    []string{"/path/to/server.js"},
    Env: map[string]string{
        "SERVER_PORT": "8080",
    },
}
claudeClient.AddMCPServer(ctx, "my-server", config)
```

### Performance Tuning
```go
// Configure cache duration
claudeClient.SetProjectContextCacheDuration(10 * time.Minute)

// Monitor cache performance
cacheInfo := claudeClient.GetProjectContextCacheInfo()
fmt.Printf("Cache hit rate: %v\n", cacheInfo["hit_rate"])
```

### Debug Configuration
```go
config.Debug = true
config.Environment = map[string]string{
    "CLAUDE_LOG_LEVEL": "debug",
    "CLAUDE_TRACE":     "true",
}
```

## Best Practices

### 1. Always Use Contexts
```go
// Create contexts with appropriate timeouts
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

### 2. Proper Resource Cleanup
```go
defer func() {
    if err := claudeClient.Close(); err != nil {
        log.Printf("Error closing client: %v", err)
    }
}()
```

### 3. Monitor Performance
```go
start := time.Now()
result, err := claudeClient.SomeOperation(ctx)
elapsed := time.Since(start)
log.Printf("Operation took %v", elapsed)
```

### 4. Handle Errors Gracefully
```go
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        // Handle timeout
    } else {
        // Handle other errors
    }
}
```

## Troubleshooting

### Common Issues

1. **MCP Server Errors**
   - Ensure Node.js is installed for npm-based servers
   - Check server installation: `npm list -g @modelcontextprotocol/server-filesystem`
   - Verify server permissions and paths

2. **Timeout Issues**
   - Increase context timeout for complex operations
   - Check network connectivity
   - Monitor resource usage

3. **Custom Path Issues**
   - Verify Claude Code executable exists at specified path
   - Check file permissions
   - Test with `claude --version`

4. **Cache Issues**
   - Invalidate cache if seeing stale data
   - Adjust cache duration based on project changes
   - Monitor cache performance in debug mode

## Performance Tips

1. **Use Caching**: Project context caching can significantly improve performance
2. **Set Appropriate Timeouts**: Balance responsiveness with operation complexity
3. **Monitor Resource Usage**: Use debug mode to identify bottlenecks
4. **Reuse Clients**: Create clients once and reuse them

## Next Steps

After understanding advanced client features, explore:
- [Session Lifecycle](../session_lifecycle/) - Session management
- [Sync Queries](../sync_queries/) - Making API calls
- [Streaming Queries](../streaming_queries/) - Real-time responses

## Related Documentation

- [MCP Protocol Documentation](https://modelcontextprotocol.io/)
- [Client Package API](../../pkg/client/)
- [Performance Guide](../../docs/performance.md)