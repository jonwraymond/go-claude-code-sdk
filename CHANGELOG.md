# Changelog

All notable changes to the Claude Code Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **NEW**: Support for latest Claude models (Claude 4 Opus, Claude 3.7 Sonnet, Claude 3.5 Sonnet v2, Claude 3.5 Haiku)
- **NEW**: MaxThinkingTokens parameter for enhanced reasoning tasks (Claude 3.5+)
- **NEW**: Enhanced TokenUsage tracking with cache metrics, web search requests, and cost estimation
- **NEW**: Comprehensive model comparison and selection guide in documentation
- Initial release of the Claude Code Go SDK
- Core client implementation with subprocess-based architecture
- Session management for conversation persistence
- Tool system integration with Claude Code CLI
- MCP (Model Context Protocol) server support
- Project context detection and analysis
- Streaming message responses
- Comprehensive error handling and categorization
- Query options with model selection, temperature, and token limits
- Permission mode configuration (Ask, AcceptEdits, RejectEdits)
- Content block types (TextBlock, ToolUseBlock, ToolResultBlock)
- Authentication via API key and environment variables
- Concurrent session support
- Context cancellation and timeout handling
- Comprehensive examples directory with all examples (consolidated from .examples)
- Integration test suite covering all major features
- GitHub Actions CI/CD pipeline with multi-version testing
- Package documentation for all public APIs
- Feature parity with official Python and TypeScript SDKs

### Changed
- **BREAKING**: Default model updated from Claude 3.5 Sonnet to Claude 3.5 Sonnet v2 (latest stable)
- **Updated**: All examples and documentation to use new model constants instead of hardcoded strings
- **Enhanced**: TokenUsage struct with additional fields for latest Claude features
- Consolidated `.examples` directory into `examples` for better visibility and consistency

### Features
- **Client Package**: Main client for interacting with Claude Code CLI
  - Subprocess management with proper cleanup
  - Streaming and synchronous query methods
  - Tool execution support
  - MCP server integration
  
- **Types Package**: Core type definitions
  - Configuration types with builder pattern
  - Message and content block types
  - Query request/response types
  - Tool and MCP server configurations
  
- **Errors Package**: Rich error handling
  - Categorized errors (Network, Validation, Auth, Process)
  - Retryable error detection
  - Error wrapping with context preservation
  - Validation error details
  
- **Auth Package**: Authentication management
  - API key authentication
  - Environment variable support
  - Credential manager for multi-project setups
  - Security best practices

### Documentation
- Comprehensive README with feature comparison table
- Examples directory with well-commented code samples
- Integration test documentation
- Package-level documentation (doc.go files)
- API reference documentation

### Testing
- Unit tests for all packages
- Integration tests with real Claude Code CLI
- CI/CD pipeline with Go 1.20-1.22 support
- Cross-platform testing (Linux, macOS, Windows)
- Security scanning with gosec
- Dependency vulnerability checking

### Developer Experience
- Makefile with common development tasks
- Linting configuration with golangci-lint
- Example programs demonstrating all features
- Clear error messages and handling
- Context-aware operations

## [0.1.0] - TBD

### Notes
This is the first public release of the Claude Code Go SDK. The SDK provides idiomatic Go interfaces for the Claude Code CLI tool, enabling AI-powered coding assistance in Go applications.

### Migration Guide
For users coming from HTTP-based Anthropic API clients:
1. This SDK uses subprocess architecture instead of HTTP
2. Authentication is still via API key
3. Streaming is supported through Go channels
4. Session management provides conversation persistence
5. Tool usage is integrated into the conversation flow

### Known Issues
- MCP server integration requires external MCP server binaries
- Windows support for some features may require additional configuration
- Integration tests require Claude Code CLI to be installed

### Future Plans
- Additional MCP server examples
- Performance optimizations for large responses
- Enhanced project context detection
- More sophisticated error recovery mechanisms

[Unreleased]: https://github.com/jonwraymond/go-claude-code-sdk/compare/v0.1.0...HEAD