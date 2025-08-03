# Go Claude Code SDK Examples

This directory contains comprehensive working examples demonstrating how to use the Go Claude Code SDK across different use cases and complexity levels.

## 🚀 Quick Start Examples

### Basic Usage
- **[basic_client](./basic_client/)** - Client initialization, configuration, and authentication methods
- **[auth_methods](./auth_methods/)** - Complete authentication comparison (API key vs subscription)

### Core Operations
- **[sync_queries](./sync_queries/)** - Synchronous query patterns with error handling and retries
- **[streaming_queries](./streaming_queries/)** - Real-time streaming responses with advanced chunk processing
- **[command_execution](./command_execution/)** - Command execution, slash commands, and development workflows

### Advanced Features
- **[session_lifecycle](./session_lifecycle/)** - Session management, persistence, and multi-session handling  
- **[advanced_client](./advanced_client/)** - Advanced features (timeouts, MCP servers, performance monitoring)
- **[mcp_integration](./mcp_integration/)** - Model Context Protocol server integration and configuration

### Specialized Examples  
- **[subscription_auth](./subscription_auth/)** - Focused subscription authentication patterns

## 📋 Example Categories

### 🚀 Beginner Examples
Start here if you're new to the Claude Code SDK:
- **[basic_client](./basic_client/)** - Your first Claude Code client
- **[auth_methods](./auth_methods/)** - Understanding authentication options

### 💬 Query Examples  
Learn different ways to interact with Claude:
- **[sync_queries](./sync_queries/)** - Traditional request-response patterns
- **[streaming_queries](./streaming_queries/)** - Real-time streaming responses

### ⚙️ Advanced Features
Explore powerful SDK capabilities:
- **[command_execution](./command_execution/)** - Automate development workflows
- **[session_lifecycle](./session_lifecycle/)** - Manage conversation state
- **[mcp_integration](./mcp_integration/)** - Extend Claude with external tools
- **[advanced_client](./advanced_client/)** - Performance optimization and monitoring

### 🔐 Authentication Focus
Deep dive into authentication patterns:
- **[subscription_auth](./subscription_auth/)** - Claude Code CLI integration

## Prerequisites

### Required Software
1. **Go 1.21+** - [Install Go](https://golang.org/doc/install)
2. **Claude Code CLI** - Install via npm:
   ```bash
   npm install -g @anthropic/claude-code
   ```
3. **Node.js 18+** - [Install Node.js](https://nodejs.org/) (for MCP servers)

### Authentication Setup
Choose your preferred authentication method:

**Option A: Subscription Authentication (Recommended for Development)**
```bash
# Setup subscription authentication (one-time setup)
claude setup-token
```

**Option B: API Key Authentication (Production/CI/CD)**
```bash
# Set your Anthropic API key
export ANTHROPIC_API_KEY="your-api-key-here"
```

### Verify Installation
```bash
# Check Claude Code CLI
claude --version

# Check authentication status  
claude auth status

# Test connection
claude --help
```

## 🚀 Quick Start

### 1. Clone and Setup
```bash
# Clone the repository
git clone https://github.com/jonwraymond/go-claude-code-sdk.git
cd go-claude-code-sdk

# Install Go dependencies
go mod download
```

### 2. Run Your First Example
```bash
# Start with the basic client example
cd examples/basic_client
go run main.go
```

### 3. Explore More Examples
```bash
# Try authentication methods
cd ../auth_methods
go run main.go

# Experience real-time streaming
cd ../streaming_queries
go run main.go

# Automate development tasks
cd ../command_execution
go run main.go
```

## 📖 Learning Path

Follow this recommended learning path:

### 🌱 **Beginner Path**
1. [basic_client](./basic_client/) - Learn client setup
2. [auth_methods](./auth_methods/) - Understand authentication
3. [sync_queries](./sync_queries/) - Make your first queries

### 🌿 **Intermediate Path**  
4. [streaming_queries](./streaming_queries/) - Real-time responses
5. [session_lifecycle](./session_lifecycle/) - Conversation management
6. [command_execution](./command_execution/) - Development automation

### 🌳 **Advanced Path**
7. [mcp_integration](./mcp_integration/) - External tool integration
8. [advanced_client](./advanced_client/) - Performance optimization
9. [subscription_auth](./subscription_auth/) - Advanced authentication

## 🛠️ Common Patterns

### Basic Client Setup
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/jonwraymond/go-claude-code-sdk/pkg/client"
    "github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
    ctx := context.Background()
    
    // Create configuration
    config := types.NewClaudeCodeConfig()
    
    // Configure authentication (choose one)
    if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
        config.APIKey = apiKey
        config.AuthMethod = types.AuthTypeAPIKey
    } else {
        config.AuthMethod = types.AuthTypeSubscription
    }
    
    // Create client
    claudeClient, err := client.NewClaudeCodeClient(ctx, config)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer claudeClient.Close()
    
    // Your code here...
}
```

### Making Queries
```go
// Synchronous query
request := &types.QueryRequest{
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "Explain Go interfaces"},
    },
    Model: "claude-3-5-sonnet-20241022",
}

response, err := claudeClient.Query(ctx, request)
if err != nil {
    log.Printf("Query failed: %v", err)
    return
}

// Extract text content
for _, block := range response.Content {
    if block.Type == "text" {
        fmt.Println(block.Text)
    }
}
```

### Streaming Responses
```go
// Create streaming query
stream, err := claudeClient.QueryStream(ctx, request)
if err != nil {
    log.Fatalf("Failed to create stream: %v", err)
}
defer stream.Close()

// Process stream chunks
for {
    chunk, err := stream.Recv()
    if err != nil || chunk.Done {
        break
    }
    fmt.Print(chunk.Content) // Real-time output
}
```

### Command Execution
```go
// Execute Claude Code commands
readCommand := client.ReadFile("main.go", client.WithSummary(true))
result, err := claudeClient.ExecuteCommand(ctx, readCommand)
if err != nil {
    log.Printf("Command failed: %v", err)
    return
}

if result.Success {
    fmt.Printf("File content: %s\n", result.Output)
}
```

## 🔧 Configuration Options

### Client Configuration
```go
config := &types.ClaudeCodeConfig{
    Model:            "claude-3-5-sonnet-20241022",  // AI model to use
    SessionID:        "my-session",                  // Session identifier
    WorkingDirectory: "/path/to/project",            // Working directory
    AuthMethod:       types.AuthTypeSubscription,   // Authentication method
    Debug:            true,                          // Enable debug logging
    TestMode:         false,                         // Test mode (no API calls)
    Environment: map[string]string{                  // Environment variables
        "CUSTOM_VAR": "value",
    },
}
```

### Query Options
```go
options := &client.QueryOptions{
    Model:          "claude-3-5-sonnet-20241022",    // Override model
    SessionID:      "conversation-1",                // Session for context
    SystemPrompt:   "You are a Go expert.",         // System instruction
    MaxTurns:       5,                               // Conversation limit
    Stream:         true,                            // Enable streaming
    PermissionMode: client.PermissionModeAsk,       // Permission handling
}
```

## ⚠️ Error Handling

The SDK provides comprehensive error types for robust error handling:

### Error Types
```go
import "github.com/jonwraymond/go-claude-code-sdk/pkg/errors"

// Handle different error types
response, err := claudeClient.Query(ctx, request)
if err != nil {
    switch e := err.(type) {
    case *errors.APIError:
        fmt.Printf("API error: %s (code: %s)\n", e.Message, e.Code)
    case *errors.ValidationError:
        fmt.Printf("Validation error: %s\n", e.Message)
    case *errors.NetworkError:
        fmt.Printf("Network error: %s\n", e.Message)
    default:
        fmt.Printf("Unknown error: %v\n", err)
    }
    return
}
```

### Best Practices
```go
// Always use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Always close clients
defer claudeClient.Close()

// Always check errors
if err != nil {
    log.Printf("Operation failed: %v", err)
    return
}

// Handle streaming errors
for {
    chunk, err := stream.Recv()
    if err != nil {
        if err == io.EOF {
            break // Normal end
        }
        log.Printf("Stream error: %v", err)
        break
    }
    // Process chunk...
}
```

## 🚨 Troubleshooting

### Common Issues

**1. "claude command not found"**
```bash
# Install Claude Code CLI
npm install -g @anthropic/claude-code

# Verify installation
claude --version
```

**2. "Authentication failed"**
```bash
# For subscription auth
claude setup-token

# For API key auth
export ANTHROPIC_API_KEY="your-key-here"
```

**3. "Connection timeout"**
```go
// Increase timeout
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()
```

**4. "Permission denied"**
```bash
# Check file permissions
ls -la go.mod

# Fix permissions if needed
chmod 644 go.mod
```

### Debug Mode
Enable debug mode for detailed logging:
```go
config := types.NewClaudeCodeConfig()
config.Debug = true
config.Environment = map[string]string{
    "CLAUDE_LOG_LEVEL": "debug",
}
```

### Getting Help
- **Documentation**: [SDK Documentation](../../docs/)
- **Issues**: [GitHub Issues](https://github.com/jonwraymond/go-claude-code-sdk/issues)
- **Examples**: All examples in this directory
- **CLI Help**: `claude --help`

## 📝 Example Output Format

Each example provides clear output showing:
- ✅ **Success indicators** - What worked
- ❌ **Error messages** - What failed and why  
- 📊 **Metrics** - Performance and usage data
- 💡 **Tips** - Helpful suggestions and next steps

## 🎯 Next Steps

After exploring the examples:

1. **Build Your Application** - Integrate the SDK into your project
2. **Explore Advanced Features** - MCP servers, streaming, sessions
3. **Optimize Performance** - Caching, batching, monitoring
4. **Join the Community** - Contribute examples and improvements

## 📚 Additional Resources

- **[SDK Documentation](../../README.md)** - Complete SDK documentation
- **[API Reference](../../pkg/)** - Detailed API documentation  
- **[Development Guide](../../docs/development.md)** - Development best practices
- **[Performance Guide](../../docs/performance.md)** - Optimization techniques

---

💡 **Tip**: Start with [basic_client](./basic_client/) and work your way through the examples at your own pace. Each example builds on concepts from previous ones.