# Working Examples Plan for Go Claude Code SDK

## Analysis Summary

Based on research of the current SDK implementation and existing examples, this plan outlines working examples that demonstrate actual SDK capabilities rather than non-existent APIs.

## Current SDK Capabilities (What Actually Works)

### Core Client Features
1. **Client Initialization** - `NewClaudeCodeClient(ctx, config)`
2. **Basic Queries** - `Query(ctx, QueryRequest)` with QueryResponse
3. **Streaming Queries** - `QueryStream(ctx, QueryRequest)` returns QueryStream
4. **Session Management** - String-based session IDs, basic lifecycle
5. **Project Context** - Working directory only (`ProjectContext.WorkingDirectory`)
6. **MCP Server Management** - Basic configuration and listing
7. **Authentication** - API key and subscription auth methods
8. **Command Execution** - Basic command execution via ExecuteCommand

### Type System (What Exists)
- `types.ClaudeCodeConfig` - Main configuration struct
- `types.QueryRequest/QueryResponse` - Query message types
- `types.Message` with `types.RoleUser/RoleAssistant`
- `types.ContentBlock` for message content
- `types.MCPServerConfig` - MCP server configuration
- `types.AuthType` - Authentication types
- `types.StreamChunk` - Streaming response chunks

## Problems with Current Examples

### Non-Existent API Methods
- `client.QueryWithProject()` - doesn't exist
- `client.StartNewSession()` - doesn't exist
- `client.Tools().GetAvailableTools()` - doesn't exist
- `mcpManager.RegisterServer()` - doesn't exist
- `ListSessions()` returns 1 value, not 2

### Non-Existent Types/Fields
- `client.QueryOptions` type - doesn't exist
- `ProjectContext.Language/ProjectName/Framework` - only WorkingDirectory exists
- `MCPServerConfig.Env/Name/Description` - only Command, Args, Environment exist
- `client.ClaudeCodeTool*` types - don't exist

### Missing Context Parameters
Many examples missing required `context.Context` parameter in method calls.

## New Working Examples Structure

### 1. Basic Examples (`examples/basic/`)

#### `basic_client_init.go`
```go
// Demonstrates basic client creation with various auth methods
- Basic config with defaults
- API key authentication  
- Subscription authentication
- Working directory configuration
- Environment variable usage
```

#### `basic_queries.go`
```go
// Demonstrates basic query operations
- Simple text queries with QueryRequest/QueryResponse
- Multi-turn conversations
- Message content handling
- Error handling patterns
```

#### `basic_streaming.go`
```go
// Demonstrates streaming responses
- QueryStream usage
- StreamChunk processing
- Proper stream cleanup
- Stream error handling
```

### 2. Session Management (`examples/sessions/`)

#### `session_lifecycle.go`
```go
// Demonstrates session management
- Creating sessions with custom IDs
- Session persistence across client instances
- UUID session ID generation
- Session listing and management
```

#### `session_conversations.go`
```go
// Demonstrates conversation continuity
- Multi-turn conversations in sessions
- Message history handling
- Session context preservation
```

### 3. MCP Integration (`examples/mcp/`)

#### `mcp_server_config.go`
```go
// Demonstrates MCP server configuration
- Adding/removing MCP servers
- Enabling/disabling servers
- Common MCP server setup
- MCP configuration management
```

#### `mcp_tools_basic.go`
```go
// Demonstrates basic MCP tool usage
- Listing configured MCP servers
- Server status checking
- Basic tool discovery (what actually works)
```

### 4. Project Context (`examples/project/`)

#### `project_context_basics.go`
```go
// Demonstrates project context usage
- Getting working directory context
- Changing working directories
- Context validation
- Directory-based operations
```

### 5. Advanced Features (`examples/advanced/`)

#### `error_handling.go`
```go
// Comprehensive error handling patterns
- SDK error types and handling
- Timeout management
- Retry patterns
- Graceful degradation
```

#### `resource_management.go`
```go
// Resource management and lifecycle
- Proper client cleanup
- Resource monitoring
- Memory management
- Process lifecycle
```

#### `configuration_patterns.go`
```go
// Advanced configuration patterns
- Environment-based configuration
- Configuration validation
- Custom Claude Code paths
- Debug and test modes
```

### 6. Command Execution (`examples/commands/`)

#### `command_basics.go`
```go
// Basic command execution
- ExecuteCommand usage
- Command result handling
- Error management
- Simple command patterns
```

#### `command_lists.go`
```go
// Command list execution (if working)
- ExecuteCommands usage
- Sequential vs parallel execution
- Command dependencies
- Batch operations
```

### 7. Integration Patterns (`examples/integration/`)

#### `cli_integration.go`
```go
// Integration with CLI workflows
- Subprocess management
- CLI argument handling
- Environment setup
- Output processing
```

#### `testing_patterns.go`
```go
// Testing patterns for SDK usage
- Test mode configuration
- Mock setups
- Integration test patterns
- Unit testing approaches
```

### 8. Real-World Use Cases (`examples/usecases/`)

#### `code_analysis.go`
```go
// Code analysis use case
- Project context for code analysis
- Multi-file query patterns
- Code review workflows
- Documentation generation
```

#### `development_assistant.go`
```go
// Development assistant use case
- Session-based development conversations
- MCP tool integration
- Project-aware responses
- Iterative development patterns
```

## Implementation Priorities

### Phase 1: Core Working Examples
1. `basic_client_init.go` - Essential for all users
2. `basic_queries.go` - Core functionality demonstration
3. `basic_streaming.go` - Streaming API usage
4. `session_lifecycle.go` - Session management basics

### Phase 2: Integration Examples
1. `mcp_server_config.go` - MCP functionality
2. `project_context_basics.go` - Project awareness
3. `error_handling.go` - Production readiness
4. `command_basics.go` - Command execution

### Phase 3: Advanced Examples
1. `resource_management.go` - Best practices
2. `configuration_patterns.go` - Advanced config
3. `testing_patterns.go` - Development support
4. Real-world use cases

## Testing Strategy

### Build Validation
```go
// Add to CI pipeline
func TestAllExamplesCompile(t *testing.T) {
    // Compile all examples to ensure they stay in sync
    // Test basic functionality without external dependencies
}
```

### Example Tests
Each example should include:
1. Compile-time validation
2. Basic functionality tests (with mocks)
3. Integration tests (optional, with real API)
4. Documentation tests (README validation)

## Documentation Structure

### Per-Example Documentation
Each example includes:
1. Clear purpose statement
2. Prerequisites and setup
3. Step-by-step walkthrough
4. Common pitfalls and solutions
5. Related examples and next steps

### Main README Structure
```markdown
# Go Claude Code SDK Examples

## Getting Started
- Installation and setup
- Authentication configuration
- Basic usage patterns

## Example Categories
- Basic Usage (start here)
- Session Management
- MCP Integration  
- Project Context
- Advanced Features
- Command Execution
- Integration Patterns
- Real-World Use Cases

## Running Examples
- Prerequisites
- Environment setup
- Individual example instructions
- Troubleshooting guide
```

## Migration from Current Examples

### Immediate Actions
1. Audit existing examples for non-working code
2. Create new structure with working examples
3. Preserve working code patterns from current examples
4. Add comprehensive build testing

### Backward Compatibility
1. Keep existing examples with clear "legacy" marking
2. Provide migration guide for users
3. Cross-reference old → new example mappings
4. Deprecation timeline for non-working examples

## Success Metrics

### Developer Experience
1. New users can run examples immediately
2. Examples compile and run successfully
3. Clear progression from basic to advanced
4. Real-world use case coverage

### Maintenance
1. Examples stay in sync with SDK changes
2. CI validates all examples continuously
3. Clear contribution guidelines for new examples
4. Regular review and update process

## Next Steps

1. **Immediate**: Create Phase 1 examples with working APIs only
2. **Short-term**: Add build validation and testing
3. **Medium-term**: Complete all example categories
4. **Long-term**: Expand based on user feedback and SDK evolution

This plan focuses on demonstrating actual SDK capabilities rather than aspirational APIs, ensuring developers can immediately use the examples productively.