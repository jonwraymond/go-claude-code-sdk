# Contributing to Go Claude Code SDK

Thank you for your interest in contributing to the Go Claude Code SDK! We welcome contributions from the community and are grateful for any help you can provide.

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
2. Claude CLI installed (`npm install -g @anthropic-ai/claude-code`)
3. Git for version control
4. Your favorite Go IDE or editor

### Development Setup

1. Fork the repository on GitHub

2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/go-claude-code-sdk.git
   cd go-claude-code-sdk
   ```

3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/jonwraymond/go-claude-code-sdk.git
   ```

4. Install dependencies and development tools:
   ```bash
   make deps
   make install-tools
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
- Keep lines under 120 characters

#### Testing

All contributions must include appropriate tests:

**Unit Tests:**
```bash
make test
```

**Integration Tests:**
```bash
# Ensure Claude CLI is authenticated first
claude auth login
make integration-test
```

**Coverage Report:**
```bash
make coverage
```

**Writing Tests:**
- Place unit tests in `*_test.go` files in the same package
- Place integration tests in `tests/`
- Use table-driven tests where appropriate
- Mock external dependencies in unit tests
- Test error cases thoroughly
- Aim for >80% code coverage

#### Documentation

- Update relevant documentation for any API changes
- Add godoc comments for all exported types and functions
- Update README.md if adding new features
- Include examples in the examples/ directory for new functionality
- Update CHANGELOG.md under "Unreleased" section

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
- `ci`: CI/CD changes
- `perf`: Performance improvements

**Examples:**
```
feat(client): add support for custom timeouts
fix(session): handle concurrent session access correctly
docs(examples): add MCP server integration example
test(query): improve coverage for error cases
```

### 4. Pull Request Process

1. **Ensure all checks pass:**
   ```bash
   make check  # Runs fmt, vet, lint, and tests
   make security  # Run security checks
   ```

2. **Update documentation:**
   - Add/update godoc comments
   - Update README.md if needed
   - Add entries to CHANGELOG.md under "Unreleased"
   - Update examples if applicable

3. **Create pull request:**
   - Use a clear, descriptive title
   - Reference any related issues with "Fixes #123"
   - Provide a detailed description of changes
   - Include examples of usage if applicable
   - Check all boxes in the PR template

4. **PR Requirements:**
   - All CI checks must pass
   - Code coverage should not decrease
   - At least one maintainer approval required
   - No merge conflicts
   - Signed commits preferred

## Testing Guidelines

### Unit Tests

Focus on testing individual components in isolation:
```go
func TestQuery(t *testing.T) {
    tests := []struct {
        name    string
        prompt  string
        want    string
        wantErr bool
    }{
        {
            name:   "simple query",
            prompt: "Hello",
            want:   "response",
        },
        {
            name:    "empty prompt",
            prompt:  "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Integration Tests

Test real interactions with Claude CLI:
```go
//go:build integration
// +build integration

func TestRealCLIInteraction(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
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
go-claude-code-sdk/
├── pkg/                 # Public API packages
│   ├── claudecode/      # Main SDK functionality
│   ├── types/           # Type definitions
│   └── errors/          # Error types
├── internal/            # Internal packages
│   ├── adapter/         # Type conversions
│   ├── transport/       # CLI communication
│   └── parser/          # Message parsing
├── examples/            # Example applications
├── tests/               # Integration tests
├── .github/            
│   └── workflows/       # CI/CD configuration
└── docs/                # Additional documentation
```

## Common Tasks

### Running Tests
```bash
make test                # Unit tests only
make integration-test    # Integration tests
make coverage           # Coverage report
make bench              # Benchmarks
```

### Code Quality
```bash
make fmt                # Format code
make lint               # Run linter
make vet                # Run go vet
make security           # Security scan
make check              # Run all checks
```

### Building
```bash
make build              # Build all packages
make clean              # Clean build artifacts
```

### Documentation
```bash
make docs               # Generate documentation
godoc -http=:6060      # View docs locally
```

## Debugging Tips

1. **Enable verbose logging:**
   ```go
   os.Setenv("CLAUDE_LOG_LEVEL", "debug")
   ```

2. **Use integration tests for debugging:**
   ```bash
   go test -v -tags=integration -run TestSpecificTest ./tests/...
   ```

3. **Check Claude CLI directly:**
   ```bash
   claude --version
   which claude
   ```

4. **Inspect CLI communication:**
   ```go
   // Add to test
   client.SetDebug(true)
   ```

## API Design Guidelines

- Keep the API surface small and focused
- Use interfaces for extensibility
- Return errors explicitly (no panic)
- Use context for cancellation
- Follow Go idioms and conventions
- Maintain backward compatibility

## Performance Guidelines

- Minimize allocations in hot paths
- Use sync.Pool for frequently allocated objects
- Benchmark critical code paths
- Profile before optimizing
- Document performance characteristics

## Security Guidelines

- Never log sensitive information
- Validate all inputs
- Use secure defaults
- Follow OWASP guidelines
- Report security issues privately

## Release Process

Releases are managed by maintainers:

1. Update version in relevant files
2. Update CHANGELOG.md with release date
3. Create and push version tag: `git tag v1.0.0`
4. GitHub Actions will create the release

## Getting Help

- Check the [documentation](README.md)
- Look at [examples](examples/)
- Open an issue for bugs or feature requests
- Join discussions in GitHub Discussions
- Check existing issues and PRs

## Recognition

Contributors will be recognized in:
- The project's README.md
- Release notes
- GitHub's contributor graph
- CHANGELOG.md

Thank you for contributing to Go Claude Code SDK! 🎉