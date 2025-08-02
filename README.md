# Go Claude Code SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/jonwraymond/go-claude-code-sdk.svg)](https://pkg.go.dev/github.com/jonwraymond/go-claude-code-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/jonwraymond/go-claude-code-sdk)](https://goreportcard.com/report/github.com/jonwraymond/go-claude-code-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

A comprehensive Go SDK for [Claude Code](https://github.com/anthropics/claude-code), the AI-powered CLI tool that lives in your terminal. This SDK provides idiomatic Go interfaces for interacting with Claude Code's subprocess-based architecture, enabling seamless integration of AI-powered coding assistance into your Go applications.

## 🎯 Purpose

The Go Claude Code SDK wraps the Claude Code CLI tool to provide:

- **Type-safe interfaces** for all Claude Code operations
- **Streaming support** for real-time responses
- **Session management** for conversation persistence
- **Tool execution** including file operations and code analysis
- **MCP server integration** for extended capabilities
- **Project-aware context** for intelligent code assistance

This SDK focuses exclusively on Claude Code CLI functionality and does not include Anthropic API features.

## 📊 Feature Parity with Official Claude Code SDKs

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
| **Integration** |
| CLI Subprocess | ✅ Native | ✅ Native | ✅ Native | Direct CLI integration |
| Error Handling | ✅ Standard | ✅ Standard | ✅ Enhanced | Go error patterns |
| Cancellation | ✅ Basic | ✅ Basic | ✅ Context-based | Full context.Context |

### Go SDK Advantages

- **Strong Type Safety**: Compile-time type checking for all operations
- **Concurrency Control**: Native goroutine support with proper synchronization
- **Context Cancellation**: First-class `context.Context` support throughout
- **Error Handling**: Idiomatic Go error handling with detailed error types
- **Performance**: Efficient subprocess management and streaming

## 🚀 Getting Started

### Prerequisites

- Go 1.20 or higher
- [Claude Code CLI](https://github.com/anthropics/claude-code) installed:
  ```bash
  npm install -g @anthropic-ai/claude-code
  ```
- Authentication via one of:
  - **Claude Subscription** (recommended): Run `claude setup-token` 
  - **API Key**: Valid Anthropic API key

### Authentication Options

The SDK supports two authentication methods:

#### 1. Subscription Authentication (Recommended)

If you have a Claude subscription, set up long-lived authentication:

```bash
# Set up subscription authentication
claude setup-token
```

Then use in your Go code:

```go
config := &types.ClaudeCodeConfig{
    WorkingDirectory: ".",
    AuthMethod:      types.AuthTypeSubscription,
}
```

#### 2. API Key Authentication

For API key authentication, set your key as an environment variable:

```bash
export ANTHROPIC_API_KEY="sk-ant-api03-your-key-here"
```

Then use in your Go code:

```go
config := &types.ClaudeCodeConfig{
    WorkingDirectory: ".",
    APIKey:          "sk-ant-api03-your-key-here", // or omit to use env var
    AuthMethod:      types.AuthTypeAPIKey,
}
```

#### 3. Automatic Detection

The SDK can automatically detect your preferred authentication method:

```go
// Automatically detects subscription or API key auth
config := types.NewClaudeCodeConfig()
// AuthMethod will be set automatically based on what's available
```

### Installation

```bash
go get github.com/jonwraymond/go-claude-code-sdk
```

### Quick Start

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
    
    // Create a new Claude Code client with automatic auth detection
    config := types.NewClaudeCodeConfig()
    // AuthMethod will be automatically detected (subscription or API key)
    
    client, err := client.NewClaudeCodeClient(ctx, config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Simple query
    response, err := client.Query(ctx, &types.QueryRequest{
        Messages: []types.Message{
            {Role: types.RoleUser, Content: "Explain this Go code"},
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Extract and print the response
    for _, block := range response.Content {
        if block.Type == "text" {
            fmt.Println(block.Text)
        }
    }
}
```

## 💻 Usage Examples

### Basic Session Creation and Query

```go
// Create a session for persistent conversations
session, err := client.Sessions().CreateSession(ctx, "my-session")
if err != nil {
    log.Fatal(err)
}
defer session.Close()

// Execute a command in the session
result, err := session.ExecuteCommand(ctx, &types.Command{
    Type: types.CommandRead,
    Args: []string{"main.go"},
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Output)
```

### Message Streaming with Options

```go
// Configure query options
options := &client.QueryOptions{
    SystemPrompt:   "You are a helpful Go programming assistant",
    MaxTurns:       10,
    AllowedTools:   []string{"read_file", "write_file", "search_code"},
    PermissionMode: client.PermissionModeAcceptEdits,
    Model:          "claude-3-opus",
}

// Stream messages with options
messages, err := client.QueryMessages(ctx, "Help me refactor this function", options)
if err != nil {
    log.Fatal(err)
}

for msg := range messages {
    switch msg.Role {
    case types.MessageRoleAssistant:
        fmt.Printf("Claude: %s\n", msg.GetText())
    case types.MessageRoleTool:
        fmt.Printf("Tool Result: %s\n", msg.GetText())
    }
}
```

### Command Execution

```go
// Execute various commands
commands := []types.Command{
    {Type: types.CommandAnalyze, Args: []string{"src/"}},
    {Type: types.CommandTest, Args: []string{"./..."}},
    {Type: types.CommandGitStatus},
}

for _, cmd := range commands {
    result, err := session.ExecuteCommand(ctx, &cmd)
    if err != nil {
        log.Printf("Command %s failed: %v", cmd.Type, err)
        continue
    }
    fmt.Printf("%s: %s\n", cmd.Type, result.Output)
}
```

### Tool System Usage

```go
// Get available tools
tools := client.Tools().GetAvailableTools()
for name, tool := range tools {
    fmt.Printf("Tool: %s - %s\n", name, tool.Description)
}

// Execute a specific tool
result, err := client.Tools().ExecuteTool(ctx, &client.ClaudeCodeTool{
    Name: "search_code",
    Arguments: map[string]interface{}{
        "pattern": "func main",
        "path":    "./",
    },
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Search results: %s\n", result.Output)
```

### MCP Server Integration

```go
// Configure MCP servers
mcpConfig := &types.MCPServerConfig{
    Name:    "sqlite",
    Command: "sqlite-mcp-server",
    Args:    []string{"./database.db"},
}

err = client.MCP().RegisterServer("sqlite", mcpConfig)
if err != nil {
    log.Fatal(err)
}

// Start MCP server
err = client.MCP().StartServer(ctx, "sqlite")
if err != nil {
    log.Fatal(err)
}

// Use MCP-provided tools
result, err := client.QueryMessagesSync(ctx, "Query the users table", nil)
if err != nil {
    log.Fatal(err)
}
```

### Project Context Detection

```go
// Get enhanced project context
context, err := client.ProjectContext().GetEnhancedProjectContext(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Project: %s\n", context.ProjectName)
fmt.Printf("Language: %s\n", context.Language)
fmt.Printf("Framework: %s\n", context.Framework)
fmt.Printf("Dependencies: %d found\n", len(context.Dependencies))

// Use project context in queries
options := &client.QueryOptions{
    SystemPrompt: fmt.Sprintf("You are helping with a %s project using %s", 
        context.Language, context.Framework),
}
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

- **Operating Systems**: Linux, macOS, Windows
- **Go Versions**: 1.20, 1.21, 1.22
- **Claude Code CLI**: Latest version recommended
- **Shell Requirements**: Bash, Zsh, PowerShell, or CMD

## 📦 Project Structure

```
claude-code-go-sdk/
├── pkg/
│   ├── client/          # Main client implementation
│   ├── types/           # Type definitions
│   ├── errors/          # Error types and handling
│   └── auth/            # Authentication helpers
├── examples/            # Example applications
├── .github/
│   └── workflows/       # CI/CD configuration
├── go.mod              # Go module definition
├── LICENSE             # MIT License
├── README.md           # This file
├── CHANGELOG.md        # Version history
└── CONTRIBUTING.md     # Contribution guidelines
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details on:

- Code style and standards
- Commit message format
- Testing requirements
- Pull request process

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Anthropic](https://www.anthropic.com) for creating Claude and Claude Code
- The Go community for excellent tooling and standards
- Contributors and users of this SDK

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/jonwraymond/go-claude-code-sdk/issues)
- **Discussions**: [GitHub Discussions](https://github.com/jonwraymond/go-claude-code-sdk/discussions)
- **Documentation**: [pkg.go.dev](https://pkg.go.dev/github.com/jonwraymond/go-claude-code-sdk)

---

Built with ❤️ for the Go and Claude Code communities