# API Reference

This section provides a detailed reference for the SDK's API.

## Client (Claude Code CLI Wrapper)

### NewClaudeCodeClient

`NewClaudeCodeClient(ctx, config)` creates a new client for interacting with the Claude Code CLI.

**Parameters:**

*   `ctx` (context.Context): Context for cancellation/timeouts
*   `config` (*types.ClaudeCodeConfig): Client configuration

**Returns:**

*   `*client.ClaudeCodeClient`: A new client.
*   `error`: An error if the client could not be created.

### Common Methods

```go
// Synchronous query
Query(ctx context.Context, req *types.QueryRequest) (*types.QueryResponse, error)

// Streaming query
QueryStream(ctx context.Context, req *types.QueryRequest) (types.QueryStream, error)

// Commands
ExecuteCommand(ctx context.Context, cmd *types.Command) (*types.CommandResult, error)
ExecuteSlashCommand(ctx context.Context, slash string) (*types.CommandResult, error)

// Sessions
CreateSession(ctx context.Context, sessionID string) (*client.ClaudeCodeSession, error)
GetSession(sessionID string) (*client.ClaudeCodeSession, error)
ListSessions() []string

// Tools
DiscoverTools(ctx context.Context) ([]*client.ClaudeCodeToolDefinition, error)
ExecuteTool(ctx context.Context, tool *client.ClaudeCodeTool) (*client.ClaudeCodeToolResult, error)
ListTools() []*client.ClaudeCodeToolDefinition
```
