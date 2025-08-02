# Claude Code SDK Compatibility Guide

This guide details the compatibility between the Go Claude Code SDK and the official Claude CLI, including feature parity, known limitations, and migration guidance.

## Version Compatibility

| SDK Version | Claude CLI Version | Compatibility |
|-------------|-------------------|---------------|
| v0.1.x | v0.1.13+ | Full compatibility |
| v0.2.x | v0.1.13+ | Full compatibility with streaming |

## Feature Compatibility Matrix

### Core Features

| Feature | SDK Support | CLI Support | Notes |
|---------|-------------|-------------|-------|
| **Basic Queries** | ✅ Full | ✅ Full | Complete parity |
| **Streaming Queries** | ✅ Full | ✅ Full | stream-json format |
| **Session Management** | ✅ Full | ✅ Full | UUID validation added |
| **Command Execution** | ✅ Full | ✅ Full | All commands supported |
| **MCP Integration** | ✅ Full | ✅ Full | Server management |
| **Project Context** | ✅ Full | ✅ Full | Auto-detection |
| **Tool Usage** | ✅ Full | ✅ Full | Via allowedTools flag |

### API Options

| Option | SDK | CLI | Mapping |
|--------|-----|-----|---------|
| `--model` | ✅ | ✅ | Direct |
| `--temperature` | ✅ | ✅ | Direct |
| `--append-system-prompt` | ✅ | ✅ | From `System` field |
| `--format` | ✅ | ✅ | Auto-set for streaming |
| `--allowedTools` | ✅ | ✅ | From `Tools` config |
| `--session` | ✅ | ✅ | Session ID |
| `--no-cache` | ✅ | ✅ | Config option |
| `--max-tokens` | ❌ | ❌ | Managed internally |
| `--top-p` | ⚠️ | ❌ | Warning logged |
| `--top-k` | ⚠️ | ❌ | Warning logged |

### Command Support

| Command Type | SDK | CLI | Example |
|-------------|-----|-----|---------|
| Read | ✅ | ✅ | `/read file.txt` |
| Write | ✅ | ✅ | `/write file.txt` |
| Search | ✅ | ✅ | `/search pattern` |
| List | ✅ | ✅ | `/ls directory` |
| Git | ✅ | ✅ | `/git status` |
| Custom Slash | ✅ | ✅ | `/help` |

## Known Differences

### 1. Flag Naming Changes
The SDK has been updated to match CLI flags exactly:
- ❌ `--system` → ✅ `--append-system-prompt`
- ❌ `--tools` → ✅ `--allowedTools`
- ❌ `--max-tokens` → Handled internally (no CLI flag)

### 2. Session ID Requirements
- **CLI**: Accepts any string as session ID
- **SDK**: Validates and normalizes to UUID format
- **Migration**: SDK auto-converts invalid IDs to valid UUIDs

### 3. Streaming Format
- **CLI**: Supports multiple formats (json, stream-json, text)
- **SDK**: Uses stream-json for QueryStream, json for Query
- **Behavior**: Automatic format selection based on method

### 4. Output Handling
- **CLI**: May truncate output with "..."
- **SDK**: Detects truncation and provides metadata
- **Enhancement**: VerboseOutput option for full content

## Migration Guide

### From CLI to SDK

#### Basic Query
**CLI**:
```bash
claude "What is 2+2?" --model claude-3-5-sonnet-20241022
```

**SDK**:
```go
request := &types.QueryRequest{
    Model: "claude-3-5-sonnet-20241022",
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "What is 2+2?"},
    },
    MaxTokens: 100,
}
response, err := client.Query(ctx, request)
```

#### Streaming Query
**CLI**:
```bash
claude "Write a story" --format stream-json
```

**SDK**:
```go
request := &types.QueryRequest{
    Model: "claude-3-5-sonnet-20241022",
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "Write a story"},
    },
    MaxTokens: 500,
    Stream: true,
}
stream, err := client.QueryStream(ctx, request)
```

#### Session Usage
**CLI**:
```bash
claude --session my-session "Hello"
```

**SDK**:
```go
session, err := client.CreateSession(ctx, "my-session")
response, err := session.Query(ctx, request)
```

### From Raw API to SDK

The SDK provides higher-level abstractions while maintaining compatibility:

1. **Automatic subprocess management**
2. **Session persistence**
3. **Streaming event parsing**
4. **Command execution helpers**
5. **MCP server integration**

## Compatibility Warnings

### Runtime Warnings
The SDK logs warnings for incompatible options:
```
WARNING: top_p parameter is not supported by Claude CLI and will be ignored
WARNING: top_k parameter is not supported by Claude CLI and will be ignored
```

### Build-Time Validation
Use the Validate() method to catch issues early:
```go
if err := request.Validate(); err != nil {
    // Handle validation error
}
```

## Platform Compatibility

| Platform | SDK Support | CLI Required | Notes |
|----------|-------------|--------------|-------|
| Linux | ✅ Full | Yes | All features |
| macOS | ✅ Full | Yes | All features |
| Windows | ✅ Full | Yes | Path handling automatic |
| WSL | ✅ Full | Yes | Use Linux CLI |

### Installation Requirements
1. Claude CLI must be installed and in PATH
2. Go 1.21+ required for SDK
3. Git recommended for repository context

## Best Practices

### 1. Version Checking
```go
// Check CLI availability on startup
if err := client.CheckCLIVersion(); err != nil {
    log.Fatal("Claude CLI not found or incompatible version")
}
```

### 2. Feature Detection
```go
// Check for specific features
if client.SupportsStreaming() {
    // Use streaming
} else {
    // Fall back to regular queries
}
```

### 3. Error Handling
```go
// Handle CLI-specific errors
if errors.Is(err, sdkerrors.ErrCLINotFound) {
    // Guide user to install CLI
}
```

### 4. Configuration
```go
config := &types.ClaudeCodeConfig{
    // Use explicit paths if needed
    ClaudeCodePath: "/usr/local/bin/claude",
    // Enable compatibility mode
    CompatibilityMode: true,
}
```

## Troubleshooting

### Common Issues

1. **"command not found: claude"**
   - Install Claude CLI: `npm install -g @anthropic-ai/claude-cli`
   - Or set explicit path in config

2. **"invalid session ID"**
   - SDK now auto-converts to UUID
   - Use GenerateSessionID() for new sessions

3. **"unknown option '--max-tokens'"**
   - This is handled internally now
   - Remove from direct CLI calls

4. **Streaming not working**
   - Ensure using QueryStream method
   - Check context cancellation
   - Verify CLI supports streaming

### Debug Mode
Enable debug logging for detailed CLI interaction:
```go
config := &types.ClaudeCodeConfig{
    Debug: true,
    LogLevel: "debug",
}
```

## Future Compatibility

### Planned Enhancements
1. **Native streaming protocol** - Direct streaming without CLI
2. **Extended tool support** - More tool types
3. **Batch operations** - Multiple queries in one call
4. **Offline mode** - Cache and replay capabilities

### Maintaining Compatibility
- SDK will track CLI changes
- Compatibility layer for deprecated features
- Migration tools for breaking changes
- Version-specific adapters

## Support

### Resources
- [SDK Issues](https://github.com/jonwraymond/go-claude-code-sdk/issues)
- [CLI Documentation](https://github.com/anthropics/claude-cli)
- [API Reference](https://docs.anthropic.com/claude)

### Compatibility Reports
Report compatibility issues with:
1. SDK version
2. CLI version (`claude --version`)
3. Platform/OS
4. Error messages
5. Minimal reproduction code