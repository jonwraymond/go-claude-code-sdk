# Basic Client Example

This example demonstrates the fundamental patterns for creating and configuring a Claude Code client. It covers various authentication methods, configuration options, and basic client operations.

## What You'll Learn

- How to create a basic Claude Code client
- Different authentication methods (API key vs subscription)
- Custom configuration options
- Working directory management
- Environment variable configuration
- Proper client lifecycle management

## Code Overview

The example includes six different client creation patterns:

### 1. Basic Client Creation
```go
config := types.NewClaudeCodeConfig()
claudeClient, err := client.NewClaudeCodeClient(ctx, config)
```

The simplest way to create a client using default configuration.

### 2. API Key Authentication
```go
config := types.NewClaudeCodeConfig()
config.APIKey = os.Getenv("ANTHROPIC_API_KEY")
config.AuthMethod = types.AuthTypeAPIKey
```

Uses your Anthropic API key for authentication. Requires `ANTHROPIC_API_KEY` environment variable.

### 3. Subscription Authentication
```go
config := types.NewClaudeCodeConfig()
config.AuthMethod = types.AuthTypeSubscription
```

Uses Claude Code CLI's built-in subscription authentication. No API key required.

### 4. Custom Configuration
```go
config := &types.ClaudeCodeConfig{
    Model:           "claude-3-opus-20240229",
    SessionID:       "my-custom-session",
    AuthMethod:      types.AuthTypeAPIKey,
    APIKey:          os.Getenv("ANTHROPIC_API_KEY"),
    WorkingDirectory: "/tmp",
    Environment: map[string]string{
        "CUSTOM_VAR": "custom_value",
        "DEBUG":      "true",
    },
    Debug:    true,
    TestMode: false,
}
```

Demonstrates full control over client configuration including custom models, session IDs, and environment variables.

### 5. Working Directory Configuration
```go
config := types.NewClaudeCodeConfig()
config.WorkingDirectory = currentDir

// Change working directory after creation
claudeClient.SetWorkingDirectory(ctx, "/tmp")
```

Shows how to set and change the working directory for Claude Code operations.

### 6. Environment Variables
```go
config.Environment = map[string]string{
    "CLAUDE_DEBUG":      "true",
    "CLAUDE_LOG_LEVEL":  "info",
    "PROJECT_NAME":      "go-claude-sdk-examples",
    "CUSTOM_TOOL_PATH":  "/usr/local/bin/custom-tools",
}
```

Demonstrates setting custom environment variables that will be available to Claude Code.

## Prerequisites

1. **Go 1.21+** installed
2. **Claude Code CLI** installed: `npm install -g @anthropic/claude-code`
3. **Authentication** configured (choose one):
   - Set `ANTHROPIC_API_KEY` environment variable
   - Authenticate with Claude Code CLI subscription

## Running the Example

### Setup Authentication

**Option 1: API Key**
```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

**Option 2: Claude Code CLI Subscription**
```bash
# Install Claude Code CLI
npm install -g @anthropic/claude-code

# Authenticate (if not already done)
claude auth
```

### Run the Example
```bash
cd examples/basic_client
go run main.go
```

## Expected Output

```
=== Claude Code Client Initialization Examples ===

--- Example 1: Basic Client ---
✓ Basic client created successfully
  Working Directory: /current/directory
  Model: claude-3-5-sonnet-20241022
  Session ID: session-1704123456789-abc123

--- Example 2: API Key Authentication ---
✓ Using API key from environment
✓ API key client created successfully
  Auth Method: api_key

--- Example 3: Subscription Authentication ---
✓ Subscription client created successfully
  Auth Method: subscription
  Note: Uses Claude Code CLI's built-in subscription authentication

--- Example 4: Custom Configuration ---
✓ Custom client created successfully
  Model: claude-3-opus-20240229
  Working Directory: /tmp
  Debug: true
  Environment Variables: 2 custom vars

--- Example 5: Working Directory Configuration ---
✓ Client with working directory created successfully
  Configured Directory: /current/directory
  Project Context Directory: /current/directory
✓ Working directory changed to: /tmp

--- Example 6: Environment Variables ---
✓ Client with environment variables created successfully
  Custom Environment Variables:
    CLAUDE_DEBUG=true
    CLAUDE_LOG_LEVEL=info
    PROJECT_NAME=go-claude-sdk-examples
    CUSTOM_TOOL_PATH=/usr/local/bin/custom-tools
```

## Key Concepts

### Configuration Defaults
- **Model**: `claude-3-5-sonnet-20241022`
- **Session ID**: Auto-generated UUID
- **Working Directory**: Current directory
- **Auth Method**: `subscription` (if no API key provided)

### Best Practices

1. **Always call `defer claudeClient.Close()`** to ensure proper cleanup
2. **Use environment variables** for sensitive data like API keys
3. **Check errors** from client creation and operations
4. **Choose appropriate auth method** based on your deployment scenario

### Authentication Methods

| Method | Pros | Cons | Use Case |
|--------|------|------|----------|
| API Key | Direct API access, programmatic control | Requires managing API keys | Production applications, CI/CD |
| Subscription | No key management, easy setup | Requires Claude Code CLI | Development, personal projects |

### Configuration Tips

- **Test Mode**: Set `config.TestMode = true` for testing without API calls
- **Debug Mode**: Enable `config.Debug = true` for detailed logging
- **Custom Models**: Specify different Claude models as needed
- **Session IDs**: Use meaningful session IDs for session tracking

## Troubleshooting

### Common Issues

1. **"Failed to create client" Error**
   - Check if Claude Code CLI is installed: `claude --version`
   - Verify authentication is set up properly

2. **API Key Not Found**
   - Ensure `ANTHROPIC_API_KEY` environment variable is set
   - Verify the API key is valid and has proper permissions

3. **Working Directory Errors**
   - Ensure the specified directory exists and is accessible
   - Check file permissions for the working directory

## Next Steps

After understanding basic client creation, explore:
- [Advanced Client](../advanced_client/) - MCP servers, timeouts, monitoring
- [Session Lifecycle](../session_lifecycle/) - Session management
- [Sync Queries](../sync_queries/) - Making API calls
- [Streaming Queries](../streaming_queries/) - Real-time responses

## Related Documentation

- [Configuration Types](../../pkg/types/config.go)
- [Client Package](../../pkg/client/)
- [Authentication Guide](../../docs/auth.md)