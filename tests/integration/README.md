# Claude Code Go SDK Integration Tests

This directory contains integration tests for the Claude Code Go SDK. These tests interact with the actual Claude Code CLI to ensure the SDK works correctly in real-world scenarios.

## Prerequisites

1. **Claude Code CLI**: Must be installed and available in your PATH
   ```bash
   npm install -g @anthropic-ai/claude-code
   ```

2. **API Key**: Set your Anthropic API key
   ```bash
   export ANTHROPIC_API_KEY="your-api-key-here"
   ```

3. **Enable Integration Tests**: Set the environment variable
   ```bash
   export INTEGRATION_TESTS=true
   ```

## Running Tests

### Run All Integration Tests
```bash
cd tests/integration
go test -v -tags=integration ./...
```

### Run Specific Test Suite
```bash
# Client tests only
go test -v -tags=integration -run TestClaudeCodeIntegrationSuite ./...

# Session tests only
go test -v -tags=integration -run TestSessionIntegrationSuite ./...

# Tool tests only
go test -v -tags=integration -run TestToolsIntegrationSuite ./...
```

### Run with Timeout
```bash
go test -v -tags=integration -timeout 10m ./...
```

## Test Suites

### 1. Client Integration Tests (`client_integration_test.go`)
Tests basic client functionality:
- Simple synchronous queries
- Streaming responses
- Query options (temperature, max tokens, model)
- Context cancellation and timeouts

### 2. Session Integration Tests (`session_integration_test.go`)
Tests conversation persistence:
- Creating and using sessions
- Conversation context retention
- Listing and managing sessions
- Session deletion
- Project-aware sessions
- Concurrent session handling

### 3. Tools Integration Tests (`tools_integration_test.go`)
Tests Claude Code's tool system:
- Listing available tools
- File read/write operations
- Code search functionality
- Command execution
- Tool permissions (AcceptEdits, RejectEdits, Ask)
- Tool usage in conversations

### 4. MCP Integration Tests (`mcp_integration_test.go`)
Tests Model Context Protocol servers:
- Server registration and configuration
- Starting and stopping servers
- Server health checks
- Multiple server management
- MCP-provided tools

**Note**: MCP tests can be skipped with:
```bash
export SKIP_MCP_TESTS=true
```

### 5. Project Context Integration Tests (`project_context_integration_test.go`)
Tests project detection and analysis:
- Language detection (Go, Python, JavaScript)
- Framework detection
- Dependency analysis
- Code structure analysis
- Enhanced project context

### 6. Error Handling Integration Tests (`error_handling_integration_test.go`)
Tests error scenarios:
- Invalid API keys
- Missing credentials
- Timeout handling
- Validation errors
- Retryable errors
- Error wrapping and categorization

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `INTEGRATION_TESTS` | Set to "true" to enable integration tests | Yes |
| `ANTHROPIC_API_KEY` | Your Anthropic API key | Yes |
| `SKIP_MCP_TESTS` | Set to "true" to skip MCP tests | No |
| `TEST_MCP_SERVER_PATH` | Path to a test MCP server binary | No |

## CI/CD Integration

For CI/CD pipelines, you can run integration tests conditionally:

```yaml
# GitHub Actions example
- name: Run Integration Tests
  if: env.ANTHROPIC_API_KEY != ''
  env:
    INTEGRATION_TESTS: true
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  run: |
    go test -v -tags=integration -timeout 15m ./tests/integration/...
```

## Writing New Integration Tests

1. **Use Build Tags**: Add `// +build integration` at the top of test files
2. **Check Environment**: Skip tests if `INTEGRATION_TESTS != "true"`
3. **Handle API Keys**: Skip tests if API key is not available
4. **Clean Up**: Always clean up resources (sessions, files, etc.)
5. **Use Testify**: Use the testify suite for better test organization

Example structure:
```go
// +build integration

package integration

import (
    "testing"
    "github.com/stretchr/testify/suite"
)

type MyIntegrationSuite struct {
    suite.Suite
    // suite fields
}

func (s *MyIntegrationSuite) SetupSuite() {
    if os.Getenv("INTEGRATION_TESTS") != "true" {
        s.T().Skip("Integration tests disabled")
    }
}

func TestMyIntegrationSuite(t *testing.T) {
    suite.Run(t, new(MyIntegrationSuite))
}
```

## Troubleshooting

### Tests Skipped
- Ensure `INTEGRATION_TESTS=true` is set
- Check that `ANTHROPIC_API_KEY` is set and valid

### Claude Command Not Found
- Verify Claude Code CLI is installed: `which claude`
- Ensure it's in your PATH

### Timeout Errors
- Increase test timeout: `-timeout 20m`
- Check network connectivity
- Verify API key is valid

### Permission Errors
- Ensure test directory is writable
- Check file permissions for tool tests

## Notes

- Integration tests make real API calls and may incur costs
- Tests create temporary files and directories that are cleaned up
- Some tests may be flaky due to network conditions
- MCP tests require MCP server binaries to be available