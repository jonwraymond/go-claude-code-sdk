# Testing Guide for Go Claude Code SDK

## Overview

This guide provides best practices and guidelines for writing tests for the Go Claude Code SDK. Our goal is to maintain >90% code coverage while ensuring tests are reliable, maintainable, and fast.

## Quick Start

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detection
make test-race

# Run specific package tests
go test ./pkg/client -v

# Run integration tests
go test -tags=integration ./tests/integration/...

# Run SDK tests
cd tests/sdk-tests && ./run_all_tests.sh
```

### Writing a New Test

```go
func TestFeatureName(t *testing.T) {
    // Arrange
    client := setupTestClient(t)
    expected := "expected result"
    
    // Act
    result, err := client.DoSomething()
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

## Test Categories

### 1. Unit Tests

Unit tests focus on testing individual functions and methods in isolation.

**Location**: Next to source files (`*_test.go`)

**Example**:
```go
func TestQueryResponse_GetTextContent(t *testing.T) {
    tests := []struct {
        name     string
        response *types.QueryResponse
        want     string
    }{
        {
            name: "nil response",
            response: nil,
            want: "",
        },
        {
            name: "response with text content",
            response: &types.QueryResponse{
                Content: []types.ContentBlock{
                    {Type: "text", Text: "Hello"},
                    {Type: "text", Text: " World"},
                },
            },
            want: "Hello World",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.response.GetTextContent()
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### 2. Integration Tests

Integration tests verify that different components work together correctly.

**Location**: `tests/integration/`

**Build Tag**: `//go:build integration`

**Example**:
```go
//go:build integration

func TestClient_QueryWithMockCLI(t *testing.T) {
    // Use mock CLI for integration testing
    mock := mocks.NewClaudeCLIMock()
    mock.SetResponse("query", fixtures.TestResponses.Success)
    
    client := setupClientWithMock(t, mock)
    
    response, err := client.Query(ctx, &types.QueryRequest{
        Messages: fixtures.TestMessages.Simple,
    })
    
    require.NoError(t, err)
    assert.NotNil(t, response)
    assert.Equal(t, "This is a successful response.", response.GetTextContent())
}
```

### 3. End-to-End Tests

E2E tests verify complete user workflows.

**Location**: `tests/e2e/`

**Requirements**: Real Claude CLI or comprehensive mocks

**Example**:
```go
func TestCompleteUserFlow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }
    
    // Test complete user flow
    client := createE2EClient(t)
    
    // 1. Create session
    session, err := client.CreateSession(ctx)
    require.NoError(t, err)
    
    // 2. Send query
    response, err := session.Query(ctx, testQuery)
    require.NoError(t, err)
    
    // 3. Use tools
    toolResult, err := session.ExecuteTool(ctx, response.GetToolCalls()[0])
    require.NoError(t, err)
    
    // 4. Close session
    err = session.Close()
    require.NoError(t, err)
}
```

## Best Practices

### 1. Use Table-Driven Tests

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
        errType error
    }{
        {"valid input", "test-123", false, nil},
        {"empty input", "", true, ErrEmptyInput},
        {"invalid format", "test 123", true, ErrInvalidFormat},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                if tt.errType != nil {
                    assert.ErrorIs(t, err, tt.errType)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 2. Use Test Fixtures

```go
// Use predefined test data
response := fixtures.TestResponses.Success
messages := fixtures.TestMessages.MultiTurn

// Create test files
testFile := fixtures.GetTestFile("hello.txt")
```

### 3. Mock External Dependencies

```go
// Create and configure mock
mock := mocks.NewClaudeCLIMock()
mock.SetResponse("--query", `{"content": [{"type": "text", "text": "Mock response"}]}`)
mock.SimulateErrors = false

// Use mock in tests
client := NewClaudeCodeClient(ctx, &types.Config{
    CommandExecutor: mock.CreateMockExecutor(),
})
```

### 4. Test Error Scenarios

```go
func TestErrorHandling(t *testing.T) {
    scenarios := []string{
        "timeout",
        "network",
        "auth",
        "rate_limit",
    }
    
    for _, scenario := range scenarios {
        t.Run(scenario, func(t *testing.T) {
            mock := mocks.NewClaudeCLIMock()
            mock.SimulateErrors = true
            
            client := setupClientWithMock(t, mock)
            _, err := client.Query(ctx, testRequest)
            
            assert.Error(t, err)
            // Verify error is handled appropriately
        })
    }
}
```

### 5. Use Proper Assertions

```go
// Use require for critical assertions that should stop the test
require.NoError(t, err)
require.NotNil(t, response)

// Use assert for non-critical checks
assert.Equal(t, expected, actual)
assert.Contains(t, output, "expected text")

// Check specific error types
assert.ErrorIs(t, err, ErrInvalidInput)
assert.ErrorAs(t, err, &validationErr)
```

### 6. Test Concurrency

```go
func TestConcurrentAccess(t *testing.T) {
    client := setupTestClient(t)
    
    var wg sync.WaitGroup
    errors := make(chan error, 10)
    
    // Run concurrent operations
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            _, err := client.Query(ctx, makeRequest(id))
            if err != nil {
                errors <- err
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    // Check for errors
    for err := range errors {
        t.Errorf("Concurrent operation failed: %v", err)
    }
}
```

### 7. Use Test Helpers

```go
// Test setup helpers
func setupTestClient(t *testing.T) *client.ClaudeCodeClient {
    t.Helper()
    
    config := &types.ClaudeCodeConfig{
        Model: "claude-3-5-sonnet-20241022",
    }
    
    client, err := client.NewClaudeCodeClient(context.Background(), config)
    require.NoError(t, err)
    
    t.Cleanup(func() {
        client.Close()
    })
    
    return client
}

// Assertion helpers
func assertJSONEqual(t *testing.T, expected, actual string) {
    t.Helper()
    
    var expectedJSON, actualJSON interface{}
    require.NoError(t, json.Unmarshal([]byte(expected), &expectedJSON))
    require.NoError(t, json.Unmarshal([]byte(actual), &actualJSON))
    
    assert.Equal(t, expectedJSON, actualJSON)
}
```

## Testing Patterns

### Testing Streaming APIs

```go
func TestStreamingResponse(t *testing.T) {
    // Create mock streaming events
    events := []string{
        `{"type":"message_start","message":{"id":"msg_123"}}`,
        `{"type":"content_block_delta","delta":{"text":"Hello"}}`,
        `{"type":"message_stop"}`,
    }
    
    mock := mocks.NewClaudeCLIMock()
    mock.SetStreamingResponse("--stream", events)
    
    client := setupClientWithMock(t, mock)
    stream, err := client.QueryStream(ctx, request)
    require.NoError(t, err)
    
    // Collect streaming response
    var content strings.Builder
    for {
        chunk, err := stream.Recv()
        if err == io.EOF {
            break
        }
        require.NoError(t, err)
        content.WriteString(chunk.Content)
    }
    
    assert.Equal(t, "Hello", content.String())
}
```

### Testing with Timeouts

```go
func TestTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    
    mock := mocks.NewClaudeCLIMock()
    mock.DefaultDelay = 200 * time.Millisecond
    
    client := setupClientWithMock(t, mock)
    _, err := client.Query(ctx, request)
    
    assert.Error(t, err)
    assert.ErrorIs(t, err, context.DeadlineExceeded)
}
```

### Testing File Operations

```go
func TestFileOperations(t *testing.T) {
    // Create temporary test directory
    tmpDir := t.TempDir()
    
    // Test file operations
    manager := NewFileManager(tmpDir)
    
    // Write test
    err := manager.WriteFile("test.txt", []byte("content"))
    require.NoError(t, err)
    
    // Read test
    content, err := manager.ReadFile("test.txt")
    require.NoError(t, err)
    assert.Equal(t, "content", string(content))
    
    // List test
    files, err := manager.ListFiles()
    require.NoError(t, err)
    assert.Contains(t, files, "test.txt")
}
```

## Coverage Guidelines

### Minimum Coverage Requirements

- **Overall**: 85%
- **Core packages** (`pkg/client`, `pkg/auth`): 90%
- **Utility packages** (`pkg/types`, `pkg/errors`): 85%
- **Integration tests**: 80%

### Checking Coverage

```bash
# Generate coverage report
make test-coverage

# View coverage in browser
go tool cover -html=coverage.out

# Check package-specific coverage
go test -cover ./pkg/client
```

### Improving Coverage

1. **Identify gaps**: Use coverage reports to find untested code
2. **Add edge cases**: Test error paths and boundary conditions
3. **Test all branches**: Ensure all if/else paths are covered
4. **Mock external calls**: Don't skip tests due to external dependencies

## CI/CD Integration

### GitHub Actions

Our CI pipeline runs:

1. **Quick checks**: Format, lint, mod tidy
2. **Unit tests**: All packages with coverage
3. **Integration tests**: With mocks
4. **SDK tests**: Specific SDK functionality
5. **Benchmarks**: Performance regression detection

### PR Requirements

- All tests must pass
- Coverage must not decrease
- No new linting issues
- Benchmarks show no significant regression

## Debugging Tests

### Verbose Output

```bash
# Run with verbose output
go test -v ./pkg/client

# Run specific test with logging
go test -v -run TestStreamingAPI ./pkg/client
```

### Debug Failed CI Tests

1. Check the GitHub Actions logs
2. Run the exact same command locally
3. Use the same Go version as CI
4. Check for environment-specific issues

### Common Issues

1. **Flaky tests**: Add proper synchronization
2. **Timeout failures**: Increase timeouts in CI
3. **Path issues**: Use relative paths or `t.TempDir()`
4. **Race conditions**: Run with `-race` flag

## Performance Testing

### Writing Benchmarks

```go
func BenchmarkQuery(b *testing.B) {
    client := setupBenchmarkClient(b)
    request := createBenchmarkRequest()
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        _, err := client.Query(context.Background(), request)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Running Benchmarks

```bash
# Run all benchmarks
make bench

# Run specific benchmark
go test -bench=BenchmarkQuery -benchmem ./pkg/client

# Compare benchmarks
go test -bench=. -count=10 > new.txt
benchstat old.txt new.txt
```

## Troubleshooting

### Test Hangs

- Add timeouts to all tests
- Use `go test -timeout 30s`
- Check for deadlocks with `-race`

### Inconsistent Results

- Ensure proper test isolation
- Don't rely on test execution order
- Use `t.Parallel()` carefully

### Mock Issues

- Verify mock configuration
- Check mock call expectations
- Enable mock debug logging

## Contributing

When adding new features:

1. Write tests first (TDD)
2. Ensure tests are deterministic
3. Add both positive and negative test cases
4. Update test documentation
5. Verify CI passes locally first

Remember: Good tests are the foundation of a reliable SDK!