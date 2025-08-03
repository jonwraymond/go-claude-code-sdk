# Go Claude Code SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/jonwraymond/go-claude-code-sdk.svg)](https://pkg.go.dev/github.com/jonwraymond/go-claude-code-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/jonwraymond/go-claude-code-sdk)](https://goreportcard.com/report/github.com/jonwraymond/go-claude-code-sdk)

A comprehensive Go SDK for interacting with Claude Code CLI, providing programmatic access to AI-powered development tools through a clean, idiomatic Go interface.

## 🚀 Features

- **Subprocess Management**: Spawns and manages Claude Code CLI processes
- **Session Persistence**: Maintains conversation context across interactions
- **Project Awareness**: Automatically detects and uses project context
- **MCP Integration**: Supports Model Context Protocol for tool extensions
- **Streaming Support**: Real-time response processing with advanced event handling
- **Command Execution**: High-level command abstraction for common operations
- **Authentication**: Multiple authentication methods (API keys, session tokens)
- **Error Handling**: Comprehensive error types with retry logic
- **Tool Management**: Discover and execute Claude Code tools
- **UUID Validation**: Proper session ID validation and generation

## 📦 Architecture Overview

The SDK is organized into several packages, each serving a specific purpose:

```
pkg/
├── client/          # Core client implementation and managers
├── types/           # Type definitions and data structures
├── auth/            # Authentication and credential management
├── errors/          # Error types and handling utilities
└── mocks/           # Test mocks and utilities
```

### Core Components

- **ClaudeCodeClient**: Main client for interacting with Claude Code CLI
- **SessionManager**: Manages conversation sessions with proper UUID handling
- **MCPManager**: Handles Model Context Protocol server integration
- **ProjectContextManager**: Manages project analysis and context
- **ToolManager**: Discovers and executes available tools
- **CommandExecutor**: High-level command execution with smart truncation detection

## 🛠 Installation

```bash
go get github.com/jonwraymond/go-claude-code-sdk
```

## 📚 Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "github.com/jonwraymond/go-claude-code-sdk/pkg/client"
    "github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
    ctx := context.Background()

    // Create client configuration
    config := &types.ClaudeCodeConfig{
        WorkingDirectory: "/path/to/project",
        SessionID:        "my-session-id",
        Model:           "claude-3-5-sonnet-20241022",
        AuthMethod:      types.AuthTypeAPIKey,
        APIKey:          "your-api-key",
    }

    // Create client
    client, err := client.NewClaudeCodeClient(ctx, config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Send a query
    response, err := client.Query(ctx, &types.QueryRequest{
        Messages: []types.Message{
            {
                Role:    types.RoleUser,
                Content: "Analyze this codebase and suggest improvements",
            },
        },
        MaxTokens: 1000,
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Response: %s", response.GetTextContent())
}
```

### Streaming Responses

```go
// Create streaming request
stream, err := client.QueryStream(ctx, &types.QueryRequest{
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "Write a Go HTTP server"},
    },
    MaxTokens: 2000,
})
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

// Process streaming chunks
for {
    chunk, err := stream.Recv()
    if err != nil {
        break
    }
    if chunk.Done {
        break
    }
    fmt.Print(chunk.Content)
}
```

### Session Management

```go
// Create a new session
sessionID := client.GenerateSessionID() // Returns UUID v4
session, err := client.CreateSession(ctx, sessionID)
if err != nil {
    log.Fatal(err)
}

// Use session for conversation
response, err := session.Query(ctx, &types.QueryRequest{
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "Hello Claude"},
    },
})

// Session automatically maintains conversation context
```

### Command Execution

```go
// Execute high-level commands
executor := client.Commands()

// Read a file
result, err := executor.ExecuteCommand(ctx, &types.Command{
    Type: client.CommandRead,
    Args: []string{"main.go"},
    Options: map[string]any{
        "summarize": true,
    },
})

// Execute multiple commands in parallel
commands := client.NewParallelCommandList(3,
    client.ReadFile("main.go"),
    client.AnalyzeCode(".", client.WithDepth("detailed")),
    client.GitStatus(),
)

results, err := executor.ExecuteCommands(ctx, commands)
```

### MCP Server Integration

```go
// Add an MCP server
err := client.AddMCPServer(ctx, "filesystem", &types.MCPServerConfig{
    Command: "npx",
    Args:    []string{"@modelcontextprotocol/server-filesystem", "/path/to/project"},
    Env:     map[string]string{"NODE_ENV": "production"},
})

// Setup common MCP servers
err = client.SetupCommonMCPServers(ctx)

// List available tools after MCP setup
tools, err := client.DiscoverTools(ctx)
```

### Error Handling

```go
response, err := client.Query(ctx, request)
if err != nil {
    // Check for specific error types
    if apiErr, ok := err.(*types.APIError); ok {
        if apiErr.IsRateLimited() {
            log.Printf("Rate limited: %s", apiErr.Message)
            // Implement backoff logic
        } else if apiErr.IsAuthenticationError() {
            log.Printf("Auth error: %s", apiErr.Message)
            // Handle authentication
        }
    } else if valErr, ok := err.(*errors.ValidationError); ok {
        log.Printf("Validation error on field %s: %s", valErr.Field, valErr.Message)
    }
    return
}
```

## 📋 API Reference

### Core Client (`pkg/client`)

#### ClaudeCodeClient

The main client for interacting with Claude Code CLI.

```go
type ClaudeCodeClient struct {
    // Internal fields...
}

// NewClaudeCodeClient creates a new Claude Code client
func NewClaudeCodeClient(ctx context.Context, config *types.ClaudeCodeConfig) (*ClaudeCodeClient, error)

// Query sends a synchronous request to Claude Code
func (c *ClaudeCodeClient) Query(ctx context.Context, request *types.QueryRequest) (*types.QueryResponse, error)

// QueryStream sends a streaming request to Claude Code
func (c *ClaudeCodeClient) QueryStream(ctx context.Context, request *types.QueryRequest) (types.QueryStream, error)

// Close gracefully shuts down the client
func (c *ClaudeCodeClient) Close() error
```

#### Session Management

```go
// CreateSession creates a new conversation session
func (c *ClaudeCodeClient) CreateSession(ctx context.Context, sessionID string) (*ClaudeCodeSession, error)

// GenerateSessionID generates a UUID v4 session ID
func (c *ClaudeCodeClient) GenerateSessionID() string

// GetSession retrieves an existing session
func (c *ClaudeCodeClient) GetSession(sessionID string) (*ClaudeCodeSession, error)

// ListSessions returns all active session IDs
func (c *ClaudeCodeClient) ListSessions() []string
```

#### MCP Management

```go
// MCP returns the MCP manager
func (c *ClaudeCodeClient) MCP() *MCPManager

// EnableMCPServer enables an MCP server by name
func (c *ClaudeCodeClient) EnableMCPServer(ctx context.Context, name string) error

// AddMCPServer adds an MCP server configuration
func (c *ClaudeCodeClient) AddMCPServer(ctx context.Context, name string, config *types.MCPServerConfig) error

// ListMCPServers returns all configured MCP servers
func (c *ClaudeCodeClient) ListMCPServers() map[string]*types.MCPServerConfig
```

#### Tool Management

```go
// DiscoverTools discovers all available tools
func (c *ClaudeCodeClient) DiscoverTools(ctx context.Context) ([]*ClaudeCodeToolDefinition, error)

// ExecuteTool executes a Claude Code tool
func (c *ClaudeCodeClient) ExecuteTool(ctx context.Context, tool *ClaudeCodeTool) (*ClaudeCodeToolResult, error)

// ListTools returns all available tools
func (c *ClaudeCodeClient) ListTools() []*ClaudeCodeToolDefinition
```

#### Command Execution

```go
// ExecuteCommand executes a Claude Code command
func (c *ClaudeCodeClient) ExecuteCommand(ctx context.Context, cmd *types.Command) (*types.CommandResult, error)

// ExecuteSlashCommand executes a slash command (e.g., "/read file.go")
func (c *ClaudeCodeClient) ExecuteSlashCommand(ctx context.Context, slashCommand string) (*types.CommandResult, error)

// ExecuteCommands executes multiple commands
func (c *ClaudeCodeClient) ExecuteCommands(ctx context.Context, cmdList *types.CommandList) (*types.CommandListResult, error)
```

#### Project Context

```go
// GetProjectContext returns basic project information
func (c *ClaudeCodeClient) GetProjectContext(ctx context.Context) (*types.ProjectContext, error)

// SetWorkingDirectory changes the working directory
func (c *ClaudeCodeClient) SetWorkingDirectory(ctx context.Context, path string) error

// GetEnhancedProjectContext returns enhanced project analysis
func (c *ClaudeCodeClient) GetEnhancedProjectContext(ctx context.Context) (*types.ProjectContext, error)
```

### Data Types (`pkg/types`)

#### Core Request/Response Types

```go
// QueryRequest represents a request to Claude Code
type QueryRequest struct {
    Model         string          `json:"model"`
    Messages      []Message       `json:"messages"`
    MaxTokens     int             `json:"max_tokens"`
    Temperature   float64         `json:"temperature,omitempty"`
    TopP          float64         `json:"top_p,omitempty"`
    TopK          int             `json:"top_k,omitempty"`
    StopSequences []string        `json:"stop_sequences,omitempty"`
    Stream        bool            `json:"stream,omitempty"`
    Tools         []Tool          `json:"tools,omitempty"`
    ToolChoice    any             `json:"tool_choice,omitempty"`
    System        string          `json:"system,omitempty"`
    Metadata      map[string]any  `json:"metadata,omitempty"`
}

// QueryResponse represents a response from Claude Code
type QueryResponse struct {
    ID           string          `json:"id"`
    Type         string          `json:"type"`
    Role         Role            `json:"role"`
    Content      []ContentBlock  `json:"content"`
    Model        string          `json:"model"`
    StopReason   string          `json:"stop_reason,omitempty"`
    StopSequence string          `json:"stop_sequence,omitempty"`
    Usage        *TokenUsage     `json:"usage,omitempty"`
    CreatedAt    time.Time       `json:"created_at,omitempty"`
    Metadata     map[string]any  `json:"metadata,omitempty"`
}
```

#### Message Types

```go
// Message represents a single message in a conversation
type Message struct {
    ID          string        `json:"id,omitempty"`
    Role        Role          `json:"role"`
    Content     string        `json:"content"`
    ToolCalls   []ToolCall    `json:"tool_calls,omitempty"`
    ToolCallID  string        `json:"tool_call_id,omitempty"`
    Attachments []Attachment  `json:"attachments,omitempty"`
    Metadata    map[string]any `json:"metadata,omitempty"`
    Timestamp   time.Time     `json:"timestamp,omitempty"`
    TokenCount  int           `json:"token_count,omitempty"`
}

// Role represents message roles
type Role string

const (
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
    RoleSystem    Role = "system"
    RoleTool      Role = "tool"
)

// ContentBlock represents content within a message
type ContentBlock struct {
    Type    string         `json:"type"`
    Text    string         `json:"text,omitempty"`
    Data    any            `json:"data,omitempty"`
    ID      string         `json:"id,omitempty"`
    Name    string         `json:"name,omitempty"`
    Input   map[string]any `json:"input,omitempty"`
}
```

#### Streaming Types

```go
// StreamEvent represents a streaming event
type StreamEvent struct {
    Type         StreamEventType `json:"type"`
    Index        int             `json:"index,omitempty"`
    Message      *StreamMessage  `json:"message,omitempty"`
    ContentBlock *ContentBlock   `json:"content_block,omitempty"`
    MessageDelta *MessageDelta   `json:"message_delta,omitempty"`
    ContentDelta *ContentDelta   `json:"delta,omitempty"`
    Error        *APIError       `json:"error,omitempty"`
    Usage        *TokenUsage     `json:"usage,omitempty"`
    Timestamp    time.Time       `json:"timestamp,omitempty"`
}

// StreamEventType represents streaming event types
type StreamEventType string

const (
    StreamEventMessage            StreamEventType = "message"
    StreamEventMessageStart      StreamEventType = "message_start"
    StreamEventContentBlockStart StreamEventType = "content_block_start"
    StreamEventContentBlockDelta StreamEventType = "content_block_delta"
    StreamEventContentBlockStop  StreamEventType = "content_block_stop"
    StreamEventMessageDelta      StreamEventType = "message_delta"
    StreamEventMessageStop       StreamEventType = "message_stop"
    StreamEventPing              StreamEventType = "ping"
    StreamEventError             StreamEventType = "error"
)
```

#### Configuration Types

```go
// ClaudeCodeConfig contains client configuration
type ClaudeCodeConfig struct {
    WorkingDirectory string                    `json:"working_directory,omitempty"`
    SessionID        string                    `json:"session_id,omitempty"`
    Model            string                    `json:"model,omitempty"`
    ClaudeCodePath   string                    `json:"claude_code_path,omitempty"`
    AuthMethod       AuthType                  `json:"auth_method,omitempty"`
    APIKey           string                    `json:"api_key,omitempty"`
    MCPServers       map[string]*MCPServerConfig `json:"mcp_servers,omitempty"`
    Environment      map[string]string         `json:"environment,omitempty"`
    Debug            bool                      `json:"debug,omitempty"`
    TestMode         bool                      `json:"test_mode,omitempty"`
}

// MCPServerConfig configures an MCP server
type MCPServerConfig struct {
    Command     string            `json:"command"`
    Args        []string          `json:"args,omitempty"`
    Env         map[string]string `json:"env,omitempty"`
    Enabled     bool              `json:"enabled"`
    Description string            `json:"description,omitempty"`
}
```

#### Command Types

```go
// Command represents a Claude Code command
type Command struct {
    Type          CommandType    `json:"type"`
    Args          []string       `json:"args,omitempty"`
    Options       map[string]any `json:"options,omitempty"`
    Context       map[string]any `json:"context,omitempty"`
    VerboseOutput bool           `json:"verbose_output,omitempty"`
}

// CommandResult represents command execution results
type CommandResult struct {
    Command      *Command       `json:"command"`
    Success      bool           `json:"success"`
    Output       string         `json:"output,omitempty"`
    FullOutput   string         `json:"full_output,omitempty"`
    IsTruncated  bool           `json:"is_truncated,omitempty"`
    OutputLength int            `json:"output_length,omitempty"`
    Error        string         `json:"error,omitempty"`
    Metadata     map[string]any `json:"metadata,omitempty"`
}

// CommandExecutionMode specifies execution strategy
type CommandExecutionMode string

const (
    ExecutionModeSequential CommandExecutionMode = "sequential"
    ExecutionModeParallel   CommandExecutionMode = "parallel"
    ExecutionModeDependency CommandExecutionMode = "dependency"
)
```

### Authentication (`pkg/auth`)

#### Authentication Manager

```go
// Manager handles credential management
type Manager struct {
    // Internal fields...
}

// NewManager creates a new credential manager
func NewManager(store CredentialStore) *Manager

// StoreAPIKey stores an API key credential
func (m *Manager) StoreAPIKey(ctx context.Context, id, apiKey string) error

// StoreSessionKey stores a session key credential
func (m *Manager) StoreSessionKey(ctx context.Context, id, sessionKey string, expiresAt *time.Time) error

// GetAuthenticator retrieves an authenticator for a credential
func (m *Manager) GetAuthenticator(ctx context.Context, id string) (Authenticator, error)

// DeleteCredential removes a credential
func (m *Manager) DeleteCredential(ctx context.Context, id string) error

// ListCredentials returns all stored credential IDs
func (m *Manager) ListCredentials(ctx context.Context) ([]string, error)
```

#### Credential Storage

```go
// CredentialStore defines credential storage interface
type CredentialStore interface {
    Store(ctx context.Context, id string, cred *StoredCredential) error
    Retrieve(ctx context.Context, id string) (*StoredCredential, error)
    Delete(ctx context.Context, id string) error
    List(ctx context.Context) ([]string, error)
}

// StoredCredential represents a stored credential
type StoredCredential struct {
    ID         string         `json:"id"`
    Type       types.AuthType `json:"type"`
    Credential string         `json:"credential"`
    Metadata   map[string]any `json:"metadata,omitempty"`
    CreatedAt  time.Time      `json:"created_at"`
    LastUsed   *time.Time     `json:"last_used,omitempty"`
    ExpiresAt  *time.Time     `json:"expires_at,omitempty"`
}

// MemoryStore implements in-memory credential storage
type MemoryStore struct { /* ... */ }

// FileStore implements file-based credential storage
type FileStore struct { /* ... */ }
```

### Error Handling (`pkg/errors`)

#### Error Types

```go
// ValidationError represents input validation errors
type ValidationError struct {
    *BaseError
    Field      string                `json:"field"`
    Value      string                `json:"value"`
    Constraint string                `json:"constraint"`
    Violations []ValidationViolation `json:"violations"`
}

// APIError represents API-level errors
type APIError struct {
    Type      string         `json:"type"`
    Message   string         `json:"message"`
    Code      int            `json:"code,omitempty"`
    Details   map[string]any `json:"details,omitempty"`
    RequestID string         `json:"request_id,omitempty"`
}

// BaseError provides common error functionality
type BaseError struct {
    category   ErrorCategory `json:"category"`
    severity   Severity      `json:"severity"`
    code       string        `json:"code"`
    message    string        `json:"message"`
    details    map[string]any `json:"details,omitempty"`
    timestamp  time.Time     `json:"timestamp"`
    retryable  bool          `json:"retryable"`
}
```

#### Error Categories

```go
type ErrorCategory string

const (
    CategoryAPI           ErrorCategory = "api"
    CategoryNetwork       ErrorCategory = "network"
    CategoryValidation    ErrorCategory = "validation"
    CategoryAuthentication ErrorCategory = "authentication"
    CategoryConfiguration ErrorCategory = "configuration"
    CategoryInternal      ErrorCategory = "internal"
)

type Severity string

const (
    SeverityLow      Severity = "low"
    SeverityMedium   Severity = "medium"
    SeverityHigh     Severity = "high"
    SeverityCritical Severity = "critical"
)
```

## 🔧 Configuration

### Environment Variables

- `ANTHROPIC_API_KEY`: Your Anthropic API key for authentication
- `CLAUDE_CODE_PATH`: Custom path to Claude Code executable
- `CLAUDE_CODE_DEBUG`: Enable debug logging

### Configuration Options

```go
config := &types.ClaudeCodeConfig{
    // Working directory for Claude Code operations
    WorkingDirectory: "/path/to/project",
    
    // Session ID for conversation persistence (UUID format recommended)
    SessionID: "550e8400-e29b-41d4-a716-446655440000",
    
    // Claude model to use
    Model: "claude-3-5-sonnet-20241022",
    
    // Authentication method
    AuthMethod: types.AuthTypeAPIKey,
    APIKey:     "your-api-key",
    
    // Custom Claude Code executable path
    ClaudeCodePath: "/usr/local/bin/claude",
    
    // MCP server configurations
    MCPServers: map[string]*types.MCPServerConfig{
        "filesystem": {
            Command: "npx",
            Args:    []string{"@modelcontextprotocol/server-filesystem", "."},
            Enabled: true,
        },
    },
    
    // Environment variables for subprocess
    Environment: map[string]string{
        "NODE_ENV": "production",
    },
    
    // Enable debug output
    Debug: true,
}
```

## 🧪 Testing

The SDK includes comprehensive test coverage with mocks for external dependencies:

```go
// Use test mode for unit tests
config := &types.ClaudeCodeConfig{
    TestMode: true,  // Uses mock implementation
    Debug:    true,  // Enable test debugging
}

client, err := client.NewClaudeCodeClient(ctx, config)
// Client will use test mocks instead of real Claude Code CLI
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -cover ./...

# Run integration tests
go test -tags=integration ./tests/integration/...
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Write comprehensive tests for new functionality
- Update documentation for API changes
- Use semantic versioning for releases
- Ensure backward compatibility when possible

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

## 🔗 Related Projects

- [Claude Code CLI](https://github.com/anthropic-ai/claude-code) - The official Claude Code command-line interface
- [Model Context Protocol](https://github.com/modelcontextprotocol) - Protocol for integrating external tools
- [Anthropic API](https://docs.anthropic.com/) - Direct API access to Claude models

## 📞 Support

- GitHub Issues: [Report bugs or request features](https://github.com/jonwraymond/go-claude-code-sdk/issues)
- Documentation: [API Reference](https://pkg.go.dev/github.com/jonwraymond/go-claude-code-sdk)
- Examples: [Example implementations](../examples/)

---

**Made with ❤️ for the Go and AI communities**