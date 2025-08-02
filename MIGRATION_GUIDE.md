# Migration Guide

This guide helps you migrate to the latest version of the Go Claude Code SDK, which includes several improvements for better Claude CLI compatibility.

## Quick Start

### What's Changed

1. **CLI Flag Compatibility** - All flags now match Claude CLI exactly
2. **UUID Session Validation** - Session IDs are validated and normalized
3. **Streaming API** - New advanced streaming with callbacks
4. **Command Lists** - Batch command execution support
5. **Output Enhancement** - Truncation detection and verbose mode

### Breaking Changes

None! All changes are backward compatible with enhanced functionality.

## Key Improvements

### 1. Session ID Handling

**Before:**
```go
// Could fail with invalid session IDs
session, err := client.CreateSession(ctx, "my session name")
```

**After:**
```go
// Automatically converts to valid UUID
session, err := client.CreateSession(ctx, "my session name")
// Or use the new helper
sessionID := client.GenerateSessionID()
session, err := client.CreateSession(ctx, sessionID)
```

### 2. Streaming Responses

**New Streaming API:**
```go
// Simple streaming
stream, err := client.QueryStream(ctx, request)
for {
    chunk, err := stream.Recv()
    if chunk.Done {
        break
    }
    fmt.Print(chunk.Content)
}

// Advanced streaming with callbacks
opts := &types.StreamOptions{
    OnContentDelta: func(delta *types.ContentDelta) error {
        fmt.Print(delta.Text)
        return nil
    },
    OnComplete: func(msg *types.StreamMessage) error {
        fmt.Printf("Tokens used: %d\n", msg.Usage.TotalTokens)
        return nil
    },
}
response, err := client.StreamQuery(ctx, request, opts)
```

### 3. Command Execution

**Batch Commands:**
```go
// Execute multiple commands efficiently
commands := client.NewCommandList(
    client.ReadFile("config.json"),
    client.SearchCode("TODO"),
    client.GitStatus(),
)
results, err := client.ExecuteCommands(ctx, commands)

// Parallel execution
commands := client.NewParallelCommandList(3,
    client.ReadFile("file1.txt"),
    client.ReadFile("file2.txt"),
    client.ReadFile("file3.txt"),
)
```

### 4. Output Handling

**Detect Truncation:**
```go
result, err := client.ExecuteCommand(ctx, cmd)
if result.IsTruncated {
    fmt.Printf("Output was truncated at %d characters\n", result.OutputLength)
}

// Get full output
cmd := client.ReadFile("large_file.txt", client.WithVerboseOutput())
result, err := client.ExecuteCommand(ctx, cmd)
// result.FullOutput contains complete content
```

## Step-by-Step Migration

### 1. Update Your Imports
No changes needed - same import paths.

### 2. Update Session Creation
```go
// Old way (still works)
session, err := client.CreateSession(ctx, "session-name")

// New recommended way
sessionID := client.GenerateSessionID()
session, err := client.CreateSession(ctx, sessionID)
```

### 3. Add Streaming Support
```go
// Add streaming to existing queries
request.Stream = true
stream, err := client.QueryStream(ctx, request)

// Or use advanced streaming
streamResp, err := client.StreamQuery(ctx, request, nil)
response, err := streamResp.Collect() // Gather complete response
```

### 4. Leverage Command Lists
```go
// Replace multiple ExecuteCommand calls
// Old:
result1, _ := client.ExecuteCommand(ctx, cmd1)
result2, _ := client.ExecuteCommand(ctx, cmd2)
result3, _ := client.ExecuteCommand(ctx, cmd3)

// New:
results, err := client.ExecuteCommands(ctx, 
    client.NewCommandList(cmd1, cmd2, cmd3))
```

## Configuration Updates

### Required CLI Version
Ensure Claude CLI v0.1.13 or later:
```bash
claude --version
```

### Recommended Config
```go
config := &types.ClaudeCodeConfig{
    Model: "claude-3-5-sonnet-20241022",
    WorkingDirectory: "/path/to/project",
    // New: Session management
    SessionConfig: &types.SessionConfig{
        AutoGenerateID: true,
        PersistSessions: true,
    },
    // New: Performance options
    CommandExecution: &types.CommandConfig{
        EnableParallel: true,
        MaxConcurrent: 5,
    },
}
```

## Testing Your Migration

Run the included test suite to verify compatibility:
```bash
cd sdk-tests
go run test_basic_init.go
go run test_streaming.go
go run test_command_list.go
```

## Common Issues

### Issue: "invalid session ID"
**Solution**: Use `GenerateSessionID()` or let SDK auto-convert

### Issue: "unknown option '--max-tokens'"
**Solution**: Already fixed - update to latest SDK

### Issue: Streaming not working
**Solution**: Use `QueryStream()` method, not `Query()` with stream flag

## Need Help?

- Check [COMPATIBILITY_GUIDE.md](docs/COMPATIBILITY_GUIDE.md)
- Review [QUERY_OPTIONS.md](docs/QUERY_OPTIONS.md)
- See [examples/](examples/) directory
- Open an issue on GitHub

## What's Next?

Future enhancements coming soon:
- Native streaming protocol
- Advanced tool management
- Session branching
- Offline mode support

Thank you for using the Go Claude Code SDK!