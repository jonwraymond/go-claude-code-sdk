# Go Claude Code SDK Tests

This directory contains comprehensive tests for the Go Claude Code SDK.

## Test Files

- `test_basic_init.go` - Tests SDK initialization with various configurations
- `test_query_simple.go` - Tests basic query functionality (CLI-compatible version)
- `test_sessions.go` - Tests session management and context persistence
- `test_commands.go` - Tests command execution (read, write, search, etc.)
- `test_mcp.go` - Tests MCP (Model Context Protocol) integration
- `test_mcp_common.go` - Tests common MCP server setup
- `test_error_handling.go` - Tests error handling and edge cases

## Running Tests

### Run Individual Tests

```bash
go run test_basic_init.go
go run test_query_simple.go
go run test_sessions.go
go run test_commands.go
go run test_mcp.go
go run test_mcp_common.go
go run test_error_handling.go
```

### Run All Tests

```bash
./run_all_tests.sh
```

## Test Results

See `SDK_TEST_RESULTS.md` for comprehensive test results and findings.

## Key Findings

1. **SDK Initialization**: Works correctly with both default and custom configurations
2. **Query Execution**: Successfully executes queries, though some CLI flags are not supported
3. **Session Management**: Maintains context correctly across multiple sessions
4. **Command Execution**: All command types work, though output may be minimal
5. **MCP Integration**: Full MCP functionality works as expected
6. **Error Handling**: Appropriate error messages for all tested error conditions

## Known Limitations

- Claude Code CLI doesn't support `--max-tokens` and `--system` flags
- Some session operations require UUID-formatted session IDs
- Command outputs may be minimal ("...") which is normal CLI behavior
- The SDK expects some types that don't exist (e.g., `CommandList`)

## Requirements

- Go 1.20 or later
- Claude Code CLI installed and accessible in PATH
- Write permissions for temporary directories (for MCP tests)
