# Claude Code SDK Package Architecture

## Overview

The Go Claude Code SDK provides a programmatic interface to the Claude Code CLI tool. The SDK follows a modular package architecture designed for maintainability, extensibility, and clear separation of concerns.

## Architecture Overview

Unlike traditional API clients that make HTTP requests, this SDK wraps the Claude Code CLI tool, executing it as a subprocess and managing its lifecycle. This approach provides:

- Direct integration with Claude Code's features
- Session persistence through CLI's built-in support
- Access to Claude Code's project-aware capabilities
- MCP (Model Context Protocol) server integration

## Package Structure

```
pkg/
├── client/          # Claude Code client and subprocess management
├── auth/            # Authentication handling (API keys, environment)
├── types/           # Core data structures and shared types
└── errors/          # Custom error types and error handling
```

## Core Components

### client/
The main package containing Claude Code-specific functionality:

- **claude_code_client.go**: Core client that manages Claude Code subprocess execution
- **claude_code_session.go**: Session management for conversation persistence
- **claude_code_tools.go**: Tool system for executing Claude Code's built-in tools
- **commands.go**: Command system for slash commands and operations
- **mcp.go**: Model Context Protocol (MCP) server management
- **project_context.go**: Project analysis and context management

### auth/
Authentication and credential management:
- API key validation and storage
- Environment variable handling
- Secure credential management

### types/
Shared data structures used across the SDK:
- Command types and structures
- Configuration options
- Message formats
- Tool definitions
- MCP server configurations

### errors/
Comprehensive error handling system:
- Structured error types for different failure scenarios
- Error categorization (API, Network, Auth, Validation)
- Go 1.13+ error wrapping support
- Security-conscious error messages

## Key Design Decisions

### 1. Subprocess Architecture
Instead of HTTP requests, the SDK executes `claude` CLI commands as subprocesses:
- Manages process lifecycle
- Handles streaming output
- Maintains session state through CLI flags

### 2. Tool System
Adapts Claude Code's tool architecture:
- Built-in tools (file operations, code search, git)
- MCP tool integration
- Subprocess-based tool execution

### 3. Session Management
Leverages Claude Code's --session flag:
- Persistent conversations across queries
- Project-aware session contexts
- Automatic session cleanup

### 4. Project Context
Deep integration with Claude Code's project awareness:
- Automatic language and framework detection
- Dependency analysis
- Development tool discovery

## Package Dependencies

```
client → auth, types, errors
auth → types, errors
types → (foundation package, no dependencies)
errors → (foundation package, no dependencies)
```

## Usage Example

```go
import (
    "context"
    "github.com/jonwraymond/go-claude-code-sdk/pkg/client"
    "github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// Create client
ctx := context.Background()
config := &types.ClaudeCodeConfig{
    WorkingDirectory: "./my-project",
    APIKey: "your-api-key",
}
client, err := client.NewClaudeCodeClient(ctx, config)

// Create a session
session, err := client.CreateSession(ctx, "my-session")

// Send a query
response, err := session.Query(ctx, &types.QueryRequest{
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "Analyze this codebase"},
    },
})

// Execute a tool
tool := &client.ClaudeCodeTool{
    Name: "read_file",
    Parameters: map[string]any{
        "path": "main.go",
    },
}
result, err := client.ExecuteTool(ctx, tool)
```

## Design Principles

1. **Subprocess Safety**: Careful process lifecycle management with proper cleanup
2. **Session Persistence**: Leverage Claude Code's built-in session support
3. **Tool Integration**: Seamless access to Claude Code's tool ecosystem
4. **Error Handling**: Comprehensive error types with security in mind
5. **Project Awareness**: Deep integration with project context and analysis

## Future Extensibility

The architecture supports future enhancements:
- Additional MCP server integrations
- Enhanced tool capabilities
- Extended project analysis features
- Performance optimizations for subprocess management