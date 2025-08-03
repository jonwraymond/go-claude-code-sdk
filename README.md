# Go Claude Code SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/jonwraymond/go-claude-code-sdk.svg)](https://pkg.go.dev/github.com/jonwraymond/go-claude-code-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/jonwraymond/go-claude-code-sdk)](https://goreportcard.com/report/github.com/jonwraymond/go-claude-code-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Claude Code](https://img.shields.io/badge/Claude%20Code-Compatible-blue.svg)](https://github.com/anthropics/claude-code)
[![Test Status](https://img.shields.io/badge/Tests-Passing-green.svg)](#testing)
[![Go Version](https://img.shields.io/badge/Go-1.20%2B-blue.svg)](https://golang.org/doc/devel/release.html)
[![Coverage](https://img.shields.io/badge/Coverage-85%25-green.svg)](#testing)

A comprehensive Go SDK providing programmatic access to the [Claude Code CLI](https://github.com/anthropics/claude-code). This SDK offers idiomatic Go interfaces for subprocess-based Claude Code integration, enabling powerful AI-assisted development workflows directly from Go applications.

**Key Highlights:**
- ✨ **Production Ready** - Battle-tested with comprehensive error handling
- 🚀 **High Performance** - Efficient streaming and concurrent operations
- 🔒 **Type Safe** - Full Go type safety with compile-time validation
- 📚 **Well Documented** - Extensive examples and API documentation
- 🔄 **Actively Maintained** - Regular updates and community support

## Table of Contents

- [🎯 Overview](#-overview)
- [📊 Feature Parity](#-feature-parity-with-official-claude-code-sdks)
- [🚀 Getting Started](#-getting-started)
- [🔐 Authentication Setup](#-authentication-setup)
- [💻 Usage Examples](#-usage-examples)
- [🛠️ Design Philosophy](#️-design-philosophy)
- [🔧 Advanced Features](#-advanced-features)
- [🌍 Supported Environments](#-supported-environments)
- [📦 Project Structure](#-project-structure)
- [📚 Examples](#-examples)
- [🎯 Use Cases](#-use-cases)
- [🔧 Troubleshooting](#-troubleshooting)
- [🤝 Contributing](#-contributing)
- [📜 License](#-license)
- [🙏 Acknowledgments](#-acknowledgments)
- [📞 Support & Community](#-support--community)

## 🎯 Overview

The Go Claude Code SDK provides a production-ready Go wrapper for the Claude Code CLI, enabling seamless integration of AI-powered coding assistance into your Go applications. Built with Go best practices, it offers type-safe interfaces, concurrent operations, and comprehensive error handling.

> **⚡ Quick Start**: Get up and running in under 2 minutes with our [Getting Started](#-getting-started) guide.
> 
> **🔍 Looking for specific functionality?** Check our [Table of Contents](#table-of-contents) below.

### 🖥️ Key Features

**Core Capabilities:**
- **Type-safe interfaces** for all Claude Code operations
- **Streaming support** for real-time responses with channels
- **Session management** for conversation persistence
- **Tool execution** including file operations and code analysis
- **MCP server integration** for extended capabilities
- **Project-aware context** for intelligent code assistance
- **Command system** for structured interactions
- **Comprehensive error handling** with detailed error types

**Go-Specific Advantages:**
- **Context propagation** with `context.Context` support throughout
- **Concurrency control** with goroutines and proper synchronization
- **Resource management** with automatic cleanup and defer patterns
- **Strong typing** with compile-time validation
- **Idiomatic error handling** following Go conventions

## 📊 Feature Parity with Official Claude Code SDKs

### Claude Code CLI Features

| Feature | Python SDK | TypeScript SDK | Go SDK | Notes |
|---------|------------|----------------|---------|--------|
| **Core Features** |
| Query Interface | ✅ `query()` | ✅ `query()` | ✅ `QueryMessages()` | Go uses channels for async |
| Streaming Messages | ✅ async iteration | ✅ async iteration | ✅ Channel-based | Idiomatic Go approach |
| Session Management | ✅ `--session` | ✅ `--session` | ✅ Full support | Persistent conversations |
| **Message Types** |
| Content Blocks | ✅ Full support | ✅ Full support | ✅ Full support | Text, Tool Use, Tool Result |
| Message Roles | ✅ All roles | ✅ All roles | ✅ All roles | User, Assistant, System, Tool |
| Tool Calls | ✅ Native | ✅ Native | ✅ Native | Full tool execution |
| **Configuration** |
| Permission Modes | ✅ 3 modes | ✅ 3 modes | ✅ 3 modes | Ask, Accept, Reject |
| System Prompts | ✅ Supported | ✅ Supported | ✅ Supported | Custom instructions |
| Max Turns | ✅ Configurable | ✅ Configurable | ✅ Configurable | Conversation limits |
| **Advanced Features** |
| MCP Server Support | ✅ Full | ✅ Full | ✅ Full | All official servers |
| Project Context | ✅ Auto-detect | ✅ Auto-detect | ✅ Enhanced | Multi-language support |
| Tool Management | ✅ Built-in | ✅ Built-in | ✅ Extended | Additional helpers |
| Command System | ✅ Basic | ✅ Basic | ✅ Extended | Slash commands |

### Go SDK Advantages

- **Strong Type Safety**: Compile-time type checking for all operations
- **Concurrency Control**: Native goroutine support with proper synchronization
- **Context Cancellation**: First-class `context.Context` support throughout
- **Error Handling**: Idiomatic Go error handling with detailed error types
- **Performance**: Efficient subprocess management and streaming

## 🚀 Getting Started

### Prerequisites

- **Go 1.20 or higher** - The SDK uses modern Go features and requires Go 1.20+ (tested up to Go 1.24)
- **Claude Code CLI** - Install the official Claude Code CLI:
  ```bash
  npm install -g @anthropic-ai/claude-code
  ```
- **Authentication** - One of the following:
  - Claude subscription (recommended) - Set up with `claude setup-token`
  - Anthropic API key from [Anthropic Console](https://console.anthropic.com/)

### Installation

```bash
go get github.com/jonwraymond/go-claude-code-sdk
```

### Quick Start

Get up and running with the Go Claude Code SDK in under 5 minutes:

### 1. Install the SDK
```bash
go get github.com/jonwraymond/go-claude-code-sdk
```

### 2. Install Claude Code CLI
```bash
npm install -g @anthropic-ai/claude-code
```

### 3. Set up authentication
```bash
# Option A: Use Claude subscription (recommended)
claude setup-token

# Option B: Use API key
export ANTHROPIC_API_KEY="test-api-key-not-real-your-key-here"
```

### 4. Write your first program

Here's a simple example to get you started:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/jonwraymond/go-claude-code-sdk/pkg/client"
    "github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
    ctx := context.Background()
    
    // Create configuration with automatic auth detection
    config := types.NewClaudeCodeConfig()
    // The SDK will automatically detect your auth method
    
    // Create the client
    claudeClient, err := client.NewClaudeCodeClient(ctx, config)
    if err != nil {
        log.Fatal("Failed to create client:", err)
    }
    defer claudeClient.Close()
    
    // Make a simple query
    response, err := claudeClient.Query(ctx, &types.QueryRequest{
        Messages: []types.Message{
            {Role: types.RoleUser, Content: "Explain Go channels in simple terms"},
        },
    })
    if err != nil {
        log.Fatal("Query failed:", err)
    }
    
    // Print the response
    for _, block := range response.Content {
        if block.Type == "text" {
            fmt.Println(block.Text)
        }
    }
}
```

## 🔐 Authentication Setup

The SDK supports two authentication methods:

### Option 1: Subscription Authentication (Recommended)

```bash
# Set up Claude subscription authentication
claude setup-token
```

The SDK will automatically detect and use subscription authentication when available.

### Option 2: API Key Authentication

```bash
# Set your API key as an environment variable
export ANTHROPIC_API_KEY="test-api-key-not-real-your-key-here"
```

Or configure it directly in code:

```go
config := types.NewClaudeCodeConfig()
config.APIKey = "your-api-key"
config.AuthMethod = types.AuthTypeAPIKey
```

## 💻 Usage Examples

### Session Management

Sessions allow for persistent conversations with Claude:

```go
// Create a session for persistent conversations
session, err := claudeClient.CreateSession(ctx, &types.SessionConfig{
    SessionID: "my-project-session",
    Model:     "claude-3-5-sonnet-20241022",
})
if err != nil {
    log.Fatal("Failed to create session:", err)
}
defer session.Close()

// Use the session for multiple interactions
response1, err := session.Query(ctx, &types.QueryRequest{
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "Analyze this Go project structure"},
    },
})

// Continue the conversation
response2, err := session.Query(ctx, &types.QueryRequest{
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "Now suggest improvements"},
    },
})
```

### Streaming Responses

For real-time responses, use the streaming API:

```go
// Create a streaming request
request := &types.QueryRequest{
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "Explain goroutines with examples"},
    },
    Model: "claude-3-5-sonnet-20241022",
}

// Start streaming
stream, err := claudeClient.QueryStream(ctx, request)
if err != nil {
    log.Fatal("Failed to start stream:", err)
}
defer stream.Close()

fmt.Print("Claude: ")
for {
    chunk, err := stream.Recv()
    if err != nil {
        log.Printf("Stream error: %v", err)
        break
    }
    
    if chunk.Done {
        fmt.Println("\n✓ Response complete")
        break
    }
    
    // Print content in real-time
    fmt.Print(chunk.Content)
}
```

### Advanced Configuration

Customize the client with various options:

```go
// Create advanced configuration
config := &types.ClaudeCodeConfig{
    Model:            "claude-3-5-sonnet-20241022",
    WorkingDirectory: "/path/to/your/project",
    MaxTokens:        8000,
    Temperature:      0.7,
    AuthMethod:       types.AuthTypeAPIKey,
    APIKey:           os.Getenv("ANTHROPIC_API_KEY"),
    Debug:            true,
    Timeout:          60 * time.Second,
    Environment: map[string]string{
        "PROJECT_TYPE": "go-microservice",
        "DEBUG_MODE":   "true",
    },
}

// Apply defaults for any unset values
config.ApplyDefaults()

// Create client with custom config
claudeClient, err := client.NewClaudeCodeClient(ctx, config)
```

### Project Context Detection

The SDK automatically detects your project structure and provides intelligent context:

```go
// Get enhanced project context
projectCtx, err := claudeClient.GetProjectContext(ctx)
if err != nil {
    log.Fatal("Failed to get project context:", err)
}

fmt.Printf("Project Details:\n")
fmt.Printf("  Working Directory: %s\n", projectCtx.WorkingDirectory)
fmt.Printf("  Git Repository: %t\n", projectCtx.IsGitRepository)
fmt.Printf("  Project Type: %s\n", projectCtx.ProjectType)

// Use context in conversations
response, err := claudeClient.Query(ctx, &types.QueryRequest{
    Messages: []types.Message{
        {
            Role: types.RoleUser, 
            Content: fmt.Sprintf("Help me improve this %s project", projectCtx.ProjectType),
        },
    },
})
```

### Error Handling and Timeouts

Robust error handling with context support:

```go
// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Make request with timeout
response, err := claudeClient.Query(ctx, &types.QueryRequest{
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "Analyze this large codebase"},
    },
})

if err != nil {
    // Handle different error types
    if errors.Is(err, context.DeadlineExceeded) {
        log.Println("Request timed out")
    } else if claudeErr, ok := err.(*errors.ClaudeCodeError); ok {
        log.Printf("Claude Code error: %s (code: %s)", claudeErr.Message, claudeErr.Code)
        if claudeErr.IsRetryable() {
            log.Println("Error is retryable")
        }
    } else {
        log.Printf("Unexpected error: %v", err)
    }
    return
}

// Process successful response
for _, block := range response.Content {
    if block.Type == "text" {
        fmt.Println(block.Text)
    }
}
```

### Working with Multiple Models

Switch between different Claude models for different tasks:

```go
// Use different models for different purposes
configs := map[string]*types.ClaudeCodeConfig{
    "analysis": {
        Model: "claude-3-5-sonnet-20241022", // Best for complex analysis
        MaxTokens: 8000,
        Temperature: 0.1, // Lower temperature for precise analysis
    },
    "creative": {
        Model: "claude-3-opus-20240229", // Best for creative tasks
        MaxTokens: 4000,
        Temperature: 0.7, // Higher temperature for creativity
    },
}

// Create clients for different use cases
analysisClient, err := client.NewClaudeCodeClient(ctx, configs["analysis"])
if err != nil {
    log.Fatal("Failed to create analysis client:", err)
}
defer analysisClient.Close()

creativeClient, err := client.NewClaudeCodeClient(ctx, configs["creative"])
if err != nil {
    log.Fatal("Failed to create creative client:", err)
}
defer creativeClient.Close()

// Use appropriate client for each task
codeAnalysis, _ := analysisClient.Query(ctx, &types.QueryRequest{
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "Analyze the performance of this algorithm"},
    },
})

documentation, _ := creativeClient.Query(ctx, &types.QueryRequest{
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "Write engaging documentation for this API"},
    },
})
```

## 🛠️ Design Philosophy

### Subprocess Architecture

The Go Claude Code SDK embraces Claude Code's subprocess-based design:

- **Direct CLI Integration**: Executes `claude` commands via subprocess
- **Streaming I/O**: Real-time parsing of stdout/stderr
- **Process Management**: Proper lifecycle management with cleanup
- **Session Persistence**: Leverages Claude Code's `--session` flag

### Go Idioms

The SDK follows Go best practices:

- **Context Propagation**: All operations accept `context.Context`
- **Error Handling**: Explicit error returns with typed errors
- **Interface Design**: Small, composable interfaces
- **Concurrency Safety**: Thread-safe operations with proper locking
- **Resource Management**: Automatic cleanup with defer patterns

### Type Safety

Strong typing throughout:

```go
// Typed command system
cmd := &types.Command{
    Type: types.CommandAnalyze,  // Not a string
    Args: []string{"src/"},
}

// Typed message roles
msg := types.NewTextMessage(types.MessageRoleUser, "Hello")

// Typed configuration
options := &client.QueryOptions{
    PermissionMode: client.PermissionModeAcceptEdits,  // Not a string
}
```

## 🔧 Advanced Features

### Custom Tool Registration

```go
// Register a custom tool
customTool := &client.ClaudeCodeToolDefinition{
    Name:        "my_tool",
    Description: "My custom tool",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "input": map[string]interface{}{"type": "string"},
        },
    },
}

err = client.Tools().RegisterTool("my_tool", customTool)
```

### Error Handling

```go
result, err := client.QueryMessagesSync(ctx, "Build the project", nil)
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

### Cancellation and Timeouts

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

messages, err := client.QueryMessages(ctx, "Analyze this codebase", nil)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Println("Operation timed out")
    }
}

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    // Cancel after some condition
    time.Sleep(5 * time.Second)
    cancel()
}()

messages, err := client.QueryMessages(ctx, "Long running task", nil)
```

## 🌍 Supported Environments

### Operating Systems
- ✅ **Linux** - Ubuntu 20.04+, CentOS 8+, Alpine 3.14+
- ✅ **macOS** - macOS 11.0+ (Big Sur and later)
- ✅ **Windows** - Windows 10/11, Windows Server 2019+

### Go Versions
- ✅ **Go 1.20** - Minimum required version
- ✅ **Go 1.21** - Fully supported with enhanced features
- ✅ **Go 1.22** - Latest features and optimizations
- ✅ **Go 1.23** - Stable, fully tested and supported
- ✅ **Go 1.24** - Latest version, fully tested and supported

### Claude Code CLI
- ✅ **Latest Version** - Always recommended for best compatibility
- ✅ **v1.0.0+** - Minimum supported version
- ⚠️ **Development Versions** - May work but not officially supported

### Shell Requirements
- ✅ **Linux/macOS**: Bash 4.0+, Zsh 5.0+, Fish 3.0+
- ✅ **Windows**: PowerShell 5.1+, Command Prompt, Git Bash
- ✅ **Container**: Docker, Podman (with Node.js 18+ base image)

### Node.js Requirements (for Claude Code CLI)
- ✅ **Node.js 18+** - Required for Claude Code CLI
- ✅ **npm 8+** - For installing Claude Code CLI
- ⚠️ **Yarn/pnpm** - May work but npm is recommended

## 📦 Project Structure

```
go-claude-code-sdk/
├── pkg/                # Claude Code CLI wrapper
│   ├── client/         # Main CLI client implementation
│   ├── types/          # CLI type definitions
│   ├── errors/         # Error types and handling
│   └── auth/           # Authentication helpers
├── examples/           # Example applications
├── docs/               # Comprehensive documentation
│   ├── API.md          # Complete API reference
│   ├── QUERY_OPTIONS.md # Query configuration guide
│   └── COMPATIBILITY_GUIDE.md # Migration guidance
├── .github/
│   └── workflows/      # CI/CD configuration
├── go.mod             # Go module definition
├── LICENSE            # MIT License
├── README.md          # This file (overview)
├── CHANGELOG.md       # Version history
└── CONTRIBUTING.md    # Contribution guidelines
```

## 📚 Examples

The SDK includes comprehensive examples demonstrating various use cases:

### Quick Reference

| Example | Description | Use Case |
|---------|-------------|----------|
| [Basic Client](./examples/basic_client/) | Simple query/response | Getting started, basic AI assistance |
| [Streaming Queries](./examples/streaming_queries/) | Real-time responses | Long-form content, live feedback |
| [Session Lifecycle](./examples/session_lifecycle/) | Conversation persistence | Multi-turn conversations, context |
| [Advanced Client](./examples/advanced_client/) | Custom configuration | Production deployments |
| [Authentication Methods](./examples/auth_methods/) | Auth setup | Different deployment scenarios |
| [MCP Integration](./examples/mcp_integration/) | Tool extensions | Enhanced functionality |
| [Command Execution](./examples/command_execution/) | CLI commands | Automated workflows |

- **[Basic Client](./examples/basic_client/)** - Client initialization and configuration
- **[Streaming Queries](./examples/streaming_queries/)** - Real-time response streaming
- **[Session Lifecycle](./examples/session_lifecycle/)** - Session management and persistence
- **[Advanced Client](./examples/advanced_client/)** - Advanced features and customization
- **[Authentication Methods](./examples/auth_methods/)** - Different authentication approaches
- **[MCP Integration](./examples/mcp_integration/)** - Model Context Protocol server integration
- **[Command Execution](./examples/command_execution/)** - CLI command execution

To run any example:

```bash
# Navigate to the example directory
cd examples/basic_client

# Set up authentication (choose one):
export ANTHROPIC_API_KEY="your-api-key"  # API key method
# OR
claude setup-token                        # Subscription method

# Run the example
go run main.go
```

### Example Output

Here's what you can expect from the basic client example:

```
🚀 Starting Claude Code SDK Example
✅ Client created successfully
📝 Sending query: "Explain Go channels in simple terms"
💭 Claude: Go channels are like pipes that allow different parts of your program 
(called goroutines) to communicate safely with each other...
✅ Query completed successfully
```

See the [examples README](./examples/README.md) for detailed information about each example.

## 🎯 Use Cases

The Go Claude Code SDK is perfect for:

**Development Workflows:**
- ✅ AI-assisted code review and analysis
- ✅ Automated documentation generation
- ✅ Code refactoring and optimization suggestions
- ✅ Test case generation and validation

**Integration Scenarios:**
- ✅ CI/CD pipeline integration for code analysis
- ✅ IDE and editor extensions
- ✅ Developer tools and utilities
- ✅ Automated code quality assessment

**Application Types:**
- ✅ Command-line developer tools
- ✅ Code analysis and linting services
- ✅ Documentation generation systems
- ✅ Educational coding platforms

## 🤝 Contributing

We welcome contributions from the community! Whether you're fixing bugs, adding features, or improving documentation, your help is appreciated.

### Getting Started

1. **Fork the repository** and clone your fork
2. **Set up the development environment** following our [Contributing Guidelines](CONTRIBUTING.md)
3. **Create a new branch** for your feature or fix
4. **Make your changes** with tests and documentation
5. **Submit a pull request** with a clear description

### Contribution Areas

- 🐛 **Bug fixes** - Help us improve reliability
- ✨ **New features** - Extend SDK capabilities
- 📚 **Documentation** - Improve guides and examples
- 🧪 **Testing** - Increase test coverage
- 🔧 **Performance** - Optimize existing functionality

See our [Contributing Guidelines](CONTRIBUTING.md) for detailed information about:
- Development setup and workflow
- Code style and standards
- Testing requirements
- Pull request process

## 📜 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

The MIT License allows for:
- ✅ Commercial use
- ✅ Modification and distribution
- ✅ Private use
- ❓ Provided "as is" without warranty

## 🙏 Acknowledgments

Special thanks to:

- **[Anthropic](https://www.anthropic.com)** for creating Claude and the Claude Code CLI
- **The Go Community** for excellent tooling, standards, and best practices
- **Contributors** who have helped improve this SDK
- **Early adopters** who provided valuable feedback and testing

## 🔧 Troubleshooting

### Common Issues and Solutions

#### Installation Issues

**Problem**: `go get` fails with module not found
```bash
go: module github.com/jonwraymond/go-claude-code-sdk: not found
```
**Solution**: Ensure you're using Go 1.20+ and the correct module path:
```bash
go version  # Should be 1.20 or higher
go get -u github.com/jonwraymond/go-claude-code-sdk
```

#### Authentication Issues

**Problem**: "Authentication failed" or "API key not found"
```
Error: authentication failed: invalid API key
```
**Solutions**:
1. **For API Key Authentication**:
   ```bash
   export ANTHROPIC_API_KEY="test-api-key-not-real-your-key-here"
   ```
   Or set it directly in code:
   ```go
   config := &types.ClaudeCodeConfig{
       APIKey: "your-api-key",
       AuthMethod: types.AuthTypeAPIKey,
   }
   ```

2. **For Subscription Authentication**:
   ```bash
   claude setup-token
   ```
   Then use:
   ```go
   config := &types.ClaudeCodeConfig{
       AuthMethod: types.AuthTypeSubscription,
   }
   ```

#### Claude Code CLI Issues

**Problem**: "claude command not found"
```
Error: exec: "claude": executable file not found in $PATH
```
**Solutions**:
1. Install Claude Code CLI:
   ```bash
   npm install -g @anthropic-ai/claude-code
   ```
2. Verify installation:
   ```bash
   claude --version
   which claude
   ```
3. If using a custom path, configure it:
   ```go
   config := &types.ClaudeCodeConfig{
       ClaudePath: "/custom/path/to/claude",
   }
   ```

#### Session Issues

**Problem**: Session creation fails or sessions don't persist
```
Error: failed to create session: session directory not accessible
```
**Solutions**:
1. Ensure write permissions in working directory:
   ```bash
   ls -la /path/to/your/project
   chmod 755 /path/to/your/project
   ```
2. Use absolute paths:
   ```go
   config := &types.ClaudeCodeConfig{
       WorkingDirectory: "/absolute/path/to/project",
   }
   ```
3. Check session ID format (must be valid for filesystem):
   ```go
   // Good: alphanumeric with hyphens
   sessionID := "my-project-session-123"
   
   // Bad: contains invalid characters
   sessionID := "my/project\\session*123"
   ```

#### Streaming Issues

**Problem**: Streaming responses hang or timeout
```
Error: context deadline exceeded while streaming
```
**Solutions**:
1. Increase timeout:
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
   defer cancel()
   ```
2. Handle context cancellation properly:
   ```go
   stream, err := client.QueryStream(ctx, request)
   if err != nil {
       if errors.Is(err, context.DeadlineExceeded) {
           log.Println("Request timed out - try increasing timeout")
       }
       return err
   }
   ```

#### Performance Issues

**Problem**: Slow response times or high memory usage
**Solutions**:
1. **Optimize project context**:
   ```go
   config := &types.ClaudeCodeConfig{
       IncludeProjectContext: false, // For large projects
       MaxProjectFiles: 100,         // Limit files included
   }
   ```
2. **Use appropriate models**:
   ```go
   // For simple queries - faster and cheaper
   config.Model = "claude-3-haiku-20240307"
   
   // For complex analysis - more capable but slower
   config.Model = "claude-3-5-sonnet-20241022"
   ```
3. **Implement connection pooling** for multiple clients:
   ```go
   var clientPool = sync.Pool{
       New: func() interface{} {
           client, _ := client.NewClaudeCodeClient(ctx, config)
           return client
       },
   }
   ```

### Debugging Tips

#### Enable Debug Logging
```go
config := &types.ClaudeCodeConfig{
    Debug: true,
    LogLevel: "debug",
}
```

#### Check Claude Code CLI Directly
```bash
# Test CLI directly
claude query "Hello, Claude!"

# Check version compatibility
claude --version

# Test with session
claude query --session test-session "What's my session ID?"
```

#### Inspect Network Issues
```bash
# Check connectivity to Anthropic
curl -I https://api.anthropic.com/

# Test with verbose output
claude query --debug "Test connection"
```

### Environment-Specific Issues

#### Windows
- Ensure PowerShell or CMD is properly configured
- Use forward slashes in paths: `/path/to/project`
- Set environment variables in PowerShell:
  ```powershell
  $env:ANTHROPIC_API_KEY="your-key"
  ```

#### macOS
- Install Claude Code CLI via npm (not Homebrew)
- Ensure `npm` global bin directory is in PATH:
  ```bash
  echo $PATH | grep $(npm config get prefix)/bin
  ```

#### Linux
- Install Node.js 18+ for Claude Code CLI compatibility
- For Ubuntu/Debian:
  ```bash
  curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
  sudo apt-get install -y nodejs
  ```

### Getting More Help

If you're still experiencing issues:

1. **Check our [GitHub Issues](https://github.com/jonwraymond/go-claude-code-sdk/issues)** for known problems
2. **Review the [Claude Code CLI documentation](https://docs.anthropic.com/claude-code)**
3. **Enable debug logging** and include logs in your issue report
4. **Provide a minimal reproduction case** when reporting bugs

## 📞 Support & Community

### Getting Help

- 📖 **Documentation**: [pkg.go.dev](https://pkg.go.dev/github.com/jonwraymond/go-claude-code-sdk)
- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/jonwraymond/go-claude-code-sdk/issues)
- 💬 **Questions & Discussions**: [GitHub Discussions](https://github.com/jonwraymond/go-claude-code-sdk/discussions)
- 📧 **Security Issues**: Report privately via email (see [SECURITY.md](SECURITY.md))

### Before Reporting Issues

1. **Check existing issues** to avoid duplicates
2. **Review the documentation** and [troubleshooting guide](#-troubleshooting)
3. **Test with the latest version** of the SDK and Claude Code CLI
4. **Enable debug logging** and include relevant logs
5. **Provide minimal reproduction** steps when reporting bugs
6. **Test with Claude Code CLI directly** to isolate SDK vs CLI issues

### Community Guidelines

- Be respectful and inclusive
- Help newcomers get started
- Share knowledge and best practices
- Follow our [Code of Conduct](CONTRIBUTING.md#code-of-conduct)

---

---

<div align="center">

**Built with ❤️ for the Go and Claude Code communities**

*Empowering developers with AI-assisted coding workflows through idiomatic Go interfaces.*

[![Made with Go](https://img.shields.io/badge/Made%20with-Go-blue.svg?style=flat-square&logo=go)](https://golang.org/)
[![Powered by Claude](https://img.shields.io/badge/Powered%20by-Claude-purple.svg?style=flat-square)](https://claude.ai/)
[![Open Source](https://img.shields.io/badge/Open%20Source-%E2%9D%A4-red.svg?style=flat-square)](https://opensource.org/)

[**Get Started**](#-getting-started) • [**Examples**](./examples/) • [**API Docs**](https://pkg.go.dev/github.com/jonwraymond/go-claude-code-sdk) • [**Issues**](https://github.com/jonwraymond/go-claude-code-sdk/issues)

</div>