# Authentication Methods Example

This example demonstrates comprehensive authentication patterns with the Claude Code SDK, including subscription auth, API key auth, automatic detection, and authentication status checking.

## What You'll Learn

- How to use subscription authentication (Claude Code CLI)
- How to configure API key authentication 
- Automatic authentication method detection
- Authentication status validation and troubleshooting
- Best practices for different deployment scenarios
- Fallback authentication strategies

## Code Overview

The example includes four authentication patterns:

### 1. Subscription Authentication
```go
config := &types.ClaudeCodeConfig{
    WorkingDirectory: ".",
    AuthMethod:       types.AuthTypeSubscription,
}

// Validate subscription auth availability
subscriptionAuth := &types.SubscriptionAuth{}
if !subscriptionAuth.IsValid(ctx) {
    fmt.Println("❌ Subscription authentication not available")
    return
}

client, err := client.NewClaudeCodeClient(ctx, config)
```

Uses Claude Code CLI's built-in subscription authentication. No API key management required.

### 2. API Key Authentication
```go
config := &types.ClaudeCodeConfig{
    WorkingDirectory: ".",
    SessionID:        "api-key-test",
    APIKey:           os.Getenv("ANTHROPIC_API_KEY"),
    AuthMethod:       types.AuthTypeAPIKey,
}

client, err := client.NewClaudeCodeClient(ctx, config)
```

Direct API key authentication for programmatic control and production deployments.

### 3. Automatic Detection
```go
config := &types.ClaudeCodeConfig{
    WorkingDirectory: ".",
    SessionID:        "auto-detect-test",
    // No explicit AuthMethod - let it auto-detect
}

config.ApplyDefaults() // Triggers auto-detection logic
```

Automatic detection chooses the best available authentication method.

### 4. Authentication Status Check
```go
// Check subscription availability
subscriptionAuth := &types.SubscriptionAuth{}
subscriptionAvailable := subscriptionAuth.IsValid(context.Background())

// Check API key availability  
apiKeyAvailable := os.Getenv("ANTHROPIC_API_KEY") != ""

// Provide recommendations
if subscriptionAvailable {
    fmt.Println("✅ Use subscription authentication for best experience")
} else {
    fmt.Println("💡 Run 'claude setup-token' to enable subscription")
}
```

Comprehensive status checking and user guidance.

## Prerequisites

1. **Go 1.21+** installed
2. **Claude Code CLI** installed: `npm install -g @anthropic/claude-code`
3. **Authentication** configured (choose one or both):
   - Subscription: `claude setup-token`
   - API Key: Set `ANTHROPIC_API_KEY` environment variable

## Running the Example

### Setup Authentication

**Option 1: Subscription Authentication (Recommended)**
```bash
# Install Claude Code CLI
npm install -g @anthropic/claude-code

# Setup subscription authentication
claude setup-token
```

**Option 2: API Key Authentication**
```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

**Option 3: Both (for comparison)**
```bash
# Setup both methods to see all examples work
claude setup-token
export ANTHROPIC_API_KEY="your-api-key-here"
```

### Run the Example
```bash
cd examples/auth_methods
go run main.go
```

## Expected Output

```
Claude Code SDK - Authentication Methods Demo
==================================================

1. Testing Subscription Authentication
----------------------------------------
✅ Subscription authentication configured successfully
✅ Subscription query successful: Hello! I'm Claude, an AI assistant...

2. Testing API Key Authentication
----------------------------------------
✅ Using API key from environment
✅ API key client created successfully
  Auth Method: api_key
✅ API key query successful: Hello! I'd be happy to help...

3. Testing Automatic Detection
----------------------------------------
🔍 Auto-detected authentication method: subscription
✅ Auto-detection successful
   Using subscription authentication

4. Authentication Status Check
----------------------------------------
Authentication Status:
  Subscription auth available: true
  API key configured: true

Recommendations:
  ✅ Use subscription authentication for the best experience
  ✅ API key authentication is available as fallback
```

## Key Concepts

### Authentication Methods Comparison

| Method | Pros | Cons | Use Case |
|--------|------|------|----------|
| **Subscription** | No key management, easy setup, integrated experience | Requires Claude Code CLI | Development, personal projects, interactive use |
| **API Key** | Direct API access, programmatic control, no CLI dependency | Key management required | Production apps, CI/CD, server deployments |

### Auto-Detection Logic

The SDK automatically chooses authentication in this order:
1. **Explicit Config**: If `AuthMethod` is set, use that method
2. **API Key Present**: If `ANTHROPIC_API_KEY` exists, use API key auth
3. **Subscription Available**: If Claude Code CLI is authenticated, use subscription
4. **Default**: Falls back to subscription auth

### Authentication Validation

```go
// Check subscription auth
subscriptionAuth := &types.SubscriptionAuth{}
isValid := subscriptionAuth.IsValid(ctx)

// Check API key auth
apiKey := os.Getenv("ANTHROPIC_API_KEY")
hasAPIKey := apiKey != ""

// Determine best method
if isValid && hasAPIKey {
    // Both available - user preference
} else if isValid {
    // Use subscription
} else if hasAPIKey {
    // Use API key
} else {
    // No auth available
}
```

## Advanced Configuration

### Custom Authentication Flow
```go
func setupAuthWithFallback(ctx context.Context) (*client.ClaudeCodeClient, error) {
    // Try subscription first
    config := &types.ClaudeCodeConfig{
        AuthMethod: types.AuthTypeSubscription,
    }
    
    subscriptionAuth := &types.SubscriptionAuth{}
    if subscriptionAuth.IsValid(ctx) {
        return client.NewClaudeCodeClient(ctx, config)
    }
    
    // Fall back to API key
    apiKey := os.Getenv("ANTHROPIC_API_KEY")
    if apiKey == "" {
        return nil, fmt.Errorf("no authentication method available")
    }
    
    config.AuthMethod = types.AuthTypeAPIKey
    config.APIKey = apiKey
    return client.NewClaudeCodeClient(ctx, config)
}
```

### Environment-Specific Auth
```go
func setupAuthForEnvironment(env string) *types.ClaudeCodeConfig {
    config := types.NewClaudeCodeConfig()
    
    switch env {
    case "development":
        // Use subscription for development
        config.AuthMethod = types.AuthTypeSubscription
    case "staging", "production":
        // Use API key for deployed environments
        config.AuthMethod = types.AuthTypeAPIKey
        config.APIKey = os.Getenv("ANTHROPIC_API_KEY")
    default:
        // Auto-detect for unknown environments
        config.ApplyDefaults()
    }
    
    return config
}
```

## Best Practices

### 1. Environment Variables
```bash
# For development
export CLAUDE_AUTH_METHOD="subscription"

# For production
export CLAUDE_AUTH_METHOD="api_key"
export ANTHROPIC_API_KEY="your-production-key"
```

### 2. Configuration Validation
```go
func validateAuthConfig(config *types.ClaudeCodeConfig) error {
    if config.AuthMethod == types.AuthTypeAPIKey && config.APIKey == "" {
        return fmt.Errorf("API key required for API key authentication")
    }
    
    if config.AuthMethod == types.AuthTypeSubscription {
        subscriptionAuth := &types.SubscriptionAuth{}
        if !subscriptionAuth.IsValid(context.Background()) {
            return fmt.Errorf("subscription authentication not available")
        }
    }
    
    return nil
}
```

### 3. Graceful Degradation
```go
func createClientWithGracefulAuth(ctx context.Context) (*client.ClaudeCodeClient, error) {
    // Try preferred method first
    if client := trySubscriptionAuth(ctx); client != nil {
        return client, nil
    }
    
    // Fall back to API key
    if client := tryAPIKeyAuth(ctx); client != nil {
        return client, nil
    }
    
    return nil, fmt.Errorf("no authentication method available")
}
```

## Troubleshooting

### Common Issues

1. **Subscription Auth Not Available**
   ```bash
   # Setup Claude Code CLI authentication
   claude setup-token
   
   # Verify authentication status
   claude auth status
   ```

2. **API Key Not Found**
   ```bash
   # Set API key environment variable
   export ANTHROPIC_API_KEY="your-key-here"
   
   # Verify it's set
   echo $ANTHROPIC_API_KEY
   ```

3. **Auto-Detection Issues**
   - Check that at least one auth method is properly configured
   - Enable debug mode: `config.Debug = true`
   - Check logs for authentication attempts

### Debug Authentication
```go
config := types.NewClaudeCodeConfig()
config.Debug = true

// Enable detailed auth logging
config.Environment = map[string]string{
    "CLAUDE_LOG_LEVEL": "debug",
    "CLAUDE_AUTH_DEBUG": "true",
}
```

### Verify Installation
```bash
# Check Claude Code CLI installation
claude --version

# Check Go installation
go version

# Test Claude Code CLI auth
claude auth status
```

## Security Considerations

### API Key Security
- **Never hardcode** API keys in source code
- **Use environment variables** for API key storage
- **Rotate keys regularly** in production
- **Limit API key scope** when possible

### Subscription Security  
- **Protect session tokens** from unauthorized access
- **Use subscription auth** only on trusted machines
- **Monitor usage** for unexpected activity

### Production Deployment
```bash
# Production environment setup
export ANTHROPIC_API_KEY="prod-key-here"
export CLAUDE_AUTH_METHOD="api_key"
export CLAUDE_LOG_LEVEL="warn"
```

## Next Steps

After understanding authentication methods, explore:
- [Basic Client](../basic_client/) - Basic client configuration
- [Sync Queries](../sync_queries/) - Making authenticated API calls
- [Session Lifecycle](../session_lifecycle/) - Session management
- [Command Execution](../command_execution/) - Development workflows

## Related Documentation

- [Authentication Types](../../pkg/types/auth.go)
- [Client Configuration](../../pkg/client/)
- [Security Guide](../../docs/security.md)