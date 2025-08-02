# Contributing to Claude Code Go SDK

Thank you for your interest in contributing to the Claude Code Go SDK! We welcome contributions from the community and are grateful for any help you can provide.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct:
- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive criticism
- Accept responsibility and apologize for mistakes
- Prioritize the project's best interests

## Getting Started

### Prerequisites

1. Go 1.20 or higher
2. Claude Code CLI installed (`npm install -g @anthropic-ai/claude-code`)
3. Git for version control
4. Your favorite Go IDE or editor

### Development Setup

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/claude-code-go-sdk.git
   cd claude-code-go-sdk
   ```

3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/jonwraymond/go-claude-code-sdk.git
   ```

4. Set up development environment:
   ```bash
   make dev-setup
   ```

5. Create a new branch for your feature:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Workflow

### 1. Before You Start

- Check existing issues and pull requests to avoid duplicate work
- For major changes, open an issue first to discuss your proposal
- Ensure your fork is up to date with the upstream repository

### 2. Making Changes

#### Code Style

We use standard Go formatting and conventions:
- Run `make fmt` to format your code
- Run `make lint` to check for linting issues
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Write clear, self-documenting code with meaningful variable names

#### Testing

All contributions must include appropriate tests:

**Unit Tests:**
```bash
make test
```

**Integration Tests:**
```bash
export ANTHROPIC_API_KEY="your-key"
make test-integration
```

**Writing Tests:**
- Place unit tests in `*_test.go` files in the same package
- Place integration tests in `tests/integration/`
- Use table-driven tests where appropriate
- Mock external dependencies in unit tests
- Use testify/assert for assertions

#### Documentation

- Update relevant documentation for any API changes
- Add godoc comments for all exported types and functions
- Update README.md if adding new features
- Include examples in the examples/ directory for new functionality

### 3. Commit Guidelines

We follow conventional commit messages:

```
type(scope): subject

body

footer
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Test additions or modifications
- `chore`: Maintenance tasks

**Examples:**
```
feat(client): add support for custom timeouts
fix(session): handle concurrent session access correctly
docs(examples): add MCP server integration example
test(tools): improve coverage for tool execution
```

### 4. Pull Request Process

1. **Ensure all checks pass:**
   ```bash
   make check  # Runs fmt, vet, lint, and tests
   ```

2. **Update documentation:**
   - Add/update godoc comments
   - Update README.md if needed
   - Add entries to CHANGELOG.md under "Unreleased"

3. **Create pull request:**
   - Use a clear, descriptive title
   - Reference any related issues
   - Provide a detailed description of changes
   - Include examples of usage if applicable

4. **PR Requirements:**
   - All CI checks must pass
   - Code coverage should not decrease
   - At least one maintainer approval required
   - No merge conflicts

## Testing Guidelines

### Unit Tests

Focus on testing individual components in isolation:
```go
func TestClaudeCodeClient_QueryMessagesSync(t *testing.T) {
    // Test setup
    client := setupTestClient(t)
    
    // Test execution
    result, err := client.QueryMessagesSync(ctx, "test query", nil)
    
    // Assertions
    assert.NoError(t, err)
    assert.NotEmpty(t, result.Content)
}
```

### Integration Tests

Test real interactions with Claude Code CLI:
```go
// +build integration

func TestRealCLIInteraction(t *testing.T) {
    if os.Getenv("INTEGRATION_TESTS") != "true" {
        t.Skip("Integration tests disabled")
    }
    // Test implementation
}
```

### Benchmarks

Add benchmarks for performance-critical code:
```go
func BenchmarkMessageParsing(b *testing.B) {
    for i := 0; i < b.N; i++ {
        parseMessage(testMessage)
    }
}
```

## Project Structure

```
claude-code-go-sdk/
├── pkg/
│   ├── client/          # Main client implementation
│   ├── types/           # Type definitions
│   ├── errors/          # Error types and handling
│   └── auth/            # Authentication
├── examples/            # Example programs
├── tests/
│   └── integration/     # Integration tests
├── .github/
│   └── workflows/       # CI/CD configuration
└── docs/                # Additional documentation
```

## Common Tasks

### Running Tests
```bash
make test              # Unit tests only
make test-integration  # Integration tests
make test-all         # All tests
```

### Code Quality
```bash
make fmt              # Format code
make lint             # Run linter
make vet              # Run go vet
make check            # Run all checks
```

### Building
```bash
make build            # Build all packages
make install          # Install the SDK
```

## Debugging Tips

1. **Enable verbose logging:**
   ```go
   config.Debug = true
   ```

2. **Use integration tests for debugging:**
   ```bash
   go test -v -tags=integration -run TestSpecificTest ./tests/integration/
   ```

3. **Check Claude Code CLI directly:**
   ```bash
   claude --version
   claude --help
   ```

## Release Process

Releases are managed by maintainers:

1. Update version in relevant files
2. Update CHANGELOG.md with release date
3. Create and push version tag: `git tag v0.1.0`
4. GitHub Actions will create the release

## Getting Help

- Check the [documentation](README.md)
- Look at [examples](examples/)
- Open an issue for bugs or feature requests
- Join discussions in GitHub Discussions

## Recognition

Contributors will be recognized in:
- The project's README.md
- Release notes
- GitHub's contributor graph

Thank you for contributing to Claude Code Go SDK! 🎉