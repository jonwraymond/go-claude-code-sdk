# Go Claude Code SDK - Python SDK Parity Report

## 📊 Overall Parity Status: ✅ COMPLETE

The Go Claude Code SDK has achieved full feature parity with the Python SDK, with enhancements specific to Go's idioms and patterns.

## 🔍 Detailed Comparison

### Core Functionality

| Feature | Python SDK | Go SDK | Status |
|---------|------------|---------|---------|
| One-shot Query | `query()` async function | `Query()` function | ✅ Equivalent |
| Sync Query | N/A (async only) | `QuerySync()` function | ✅ Go Enhancement |
| Interactive Client | `ClaudeSDKClient` class | `ClaudeSDKClient` struct | ✅ Equivalent |
| Streaming Support | AsyncIterator | Channel-based | ✅ Idiomatic |
| Context Management | `async with` | `context.Context` | ✅ Idiomatic |
| Interrupt Support | `await client.interrupt()` | `client.Interrupt()` | ✅ Equivalent |

### Type System

| Type | Python SDK | Go SDK | Status |
|------|------------|---------|---------|
| ClaudeCodeOptions | `@dataclass` | `struct` with builder pattern | ✅ Enhanced |
| Message Types | `UserMessage`, `AssistantMessage`, `SystemMessage`, `ResultMessage` | Same types | ✅ Identical |
| Content Blocks | `TextBlock`, `ToolUseBlock`, `ToolResultBlock` | Same types | ✅ Identical |
| Permission Modes | `Literal["default", "acceptEdits", "bypassPermissions"]` | Typed constants | ✅ Type-safe |
| MCP Server Config | `TypedDict` unions | Interface-based | ✅ More flexible |

### Error Handling

| Error Type | Python SDK | Go SDK | Status |
|------------|------------|---------|---------|
| Base Error | `ClaudeSDKError` | `ClaudeSDKError` | ✅ Equivalent |
| CLI Connection | `CLIConnectionError` | `CLIConnectionError` | ✅ Equivalent |
| CLI Not Found | `CLINotFoundError` | `CLINotFoundError` | ✅ Equivalent |
| Process Error | `ProcessError` | `ProcessError` | ✅ Equivalent |
| JSON Decode | `CLIJSONDecodeError` | `CLIJSONDecodeError` | ✅ Equivalent |

### API Surface

#### Python SDK Exports
```python
# Main functions
query()
ClaudeSDKClient

# Types
PermissionMode
McpServerConfig
UserMessage, AssistantMessage, SystemMessage, ResultMessage
Message
ClaudeCodeOptions
TextBlock, ToolUseBlock, ToolResultBlock
ContentBlock

# Errors
ClaudeSDKError
CLIConnectionError
CLINotFoundError
ProcessError
CLIJSONDecodeError
```

#### Go SDK Exports
```go
// Main functions
Query()           // Async equivalent
QuerySync()       // Sync enhancement
ClaudeSDKClient

// Types
PermissionMode (with constants)
McpServerConfig
UserMessage, AssistantMessage, SystemMessage, ResultMessage
Message
ClaudeCodeOptions
TextBlock, ToolUseBlock, ToolResultBlock
ContentBlock

// Helper functions (Go enhancement)
NewClaudeCodeOptions()
WithSystemPrompt(), WithMaxTurns(), etc.
StringPtr(), IntPtr(), PermissionModePtr()

// Errors (in pkg/errors)
ClaudeSDKError
CLIConnectionError
CLINotFoundError
ProcessError
CLIJSONDecodeError
```

### Package Structure Comparison

#### Python SDK Structure
```
claude-code-sdk-python/
├── src/claude_code_sdk/
│   ├── __init__.py          # Main exports
│   ├── query.py             # One-shot query
│   ├── client.py            # Interactive client
│   ├── types.py             # Type definitions
│   ├── _errors.py           # Error types
│   └── _internal/           # Implementation
│       ├── client.py
│       ├── message_parser.py
│       └── transport/
│           └── subprocess_cli.py
```

#### Go SDK Structure
```
go-claude-code-sdk/
├── pkg/
│   ├── claudecode/          # Main package
│   │   ├── query.go         # One-shot query
│   │   ├── client.go        # Interactive client
│   │   ├── types.go         # Type exports
│   │   └── options.go       # Options & helpers
│   ├── types/               # Type definitions
│   │   ├── messages.go      # Message types
│   │   ├── options.go       # Options types
│   │   └── mcp.go          # MCP types
│   └── errors/              # Error types
│       └── errors.go
└── internal/                # Implementation
    ├── adapter/
    ├── transport/
    └── parser/
```

## ✨ Go SDK Enhancements

### 1. Synchronous Support
- `QuerySync()` - Blocking version for simpler use cases
- Useful for CLI tools and scripts

### 2. Builder Pattern
- Functional options for `ClaudeCodeOptions`
- Type-safe configuration with `With*` functions

### 3. Helper Functions
- `StringPtr()`, `IntPtr()` - Convenience for optional fields
- `NewClaudeCodeOptions()` - Default configuration

### 4. Context Integration
- Native `context.Context` support
- Proper cancellation and timeout handling

### 5. Channel-Based Streaming
- Idiomatic Go channels instead of AsyncIterator
- Natural concurrency patterns

### 6. Comprehensive Examples
- 14 example categories (vs 3 in Python)
- Benchmarks and performance testing
- Real-world application examples

## 📋 Feature Checklist

- [x] One-shot query functionality
- [x] Interactive client with streaming
- [x] All message types (User, Assistant, System, Result)
- [x] All content blocks (Text, ToolUse, ToolResult)
- [x] Tool restrictions and allowed tools
- [x] Permission modes (default, acceptEdits, bypassPermissions)
- [x] MCP server configuration
- [x] Error types with proper error chains
- [x] Interrupt support
- [x] Session management
- [x] Working directory configuration
- [x] System prompts
- [x] Max turns limitation
- [x] Context cancellation
- [x] Comprehensive documentation
- [x] Unit tests
- [x] Integration tests
- [x] Examples for all features

## 🎯 Conclusion

The Go Claude Code SDK has achieved **complete feature parity** with the Python SDK while adding idiomatic Go enhancements. The SDK provides:

1. **All core functionality** from the Python SDK
2. **Go-specific improvements** (sync support, builders, context)
3. **Comprehensive examples** demonstrating all features
4. **Thorough testing** including integration tests
5. **Performance benchmarks** for optimization

The Go SDK is ready for production use and offers developers a familiar, idiomatic experience when working with Claude Code from Go applications.