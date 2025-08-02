# Claude Code Super Crew - Go SDK Command Patterns

## Command Reference for Go SDK Development

This guide provides specific command patterns optimized for Go Claude Code SDK development workflows.

## Core Development Commands

### `/crew:implement` - Feature Implementation

#### Go Language Features
```bash
# Interface design and implementation
/crew:implement "Add configurable retry policies with exponential backoff"
# Routes to: architect-agent → go-agent → testing-agent

# Concurrency patterns
/crew:implement "Implement worker pool for parallel tool execution"
# Routes to: go-agent → architect-agent → testing-agent → security-agent

# Error handling improvements
/crew:implement "Add structured error types with wrapping and context"
# Routes to: go-agent → architect-agent → testing-agent
```

#### CLI Integration Features
```bash
# Subprocess management
/crew:implement "Add timeout and cancellation support for CLI commands"
# Routes to: cli-agent → security-agent → testing-agent

# Session management
/crew:implement "Implement session persistence and recovery"
# Routes to: cli-agent → backend-agent → testing-agent

# Streaming enhancements
/crew:implement "Add real-time message parsing with custom handlers"
# Routes to: cli-agent → go-agent → testing-agent
```

#### Testing & Quality
```bash
# Test infrastructure
/crew:implement "Add integration test suite for CLI subprocess calls"
# Routes to: testing-agent → cli-agent → devops-agent

# Mocking and testing utilities
/crew:implement "Create mock CLI client for unit testing"
# Routes to: testing-agent → go-agent → architect-agent

# Performance testing
/crew:implement "Add benchmarks for streaming performance"
# Routes to: testing-agent → go-agent → cli-agent
```

### `/crew:improve` - Quality Enhancement

#### Performance Optimization
```bash
# Memory and CPU optimization
/crew:improve "Optimize memory usage in streaming message handling"
# Routes to: go-agent → architect-agent → testing-agent

# Concurrency improvements
/crew:improve "Enhance goroutine management and resource cleanup"
# Routes to: go-agent → architect-agent → testing-agent → security-agent

# CLI performance
/crew:improve "Reduce subprocess creation overhead"
# Routes to: cli-agent → go-agent → testing-agent
```

#### Code Quality
```bash
# Architecture improvements
/crew:improve "Refactor package structure for better modularity"
# Routes to: architect-agent → go-agent → testing-agent → scribe-agent

# Error handling enhancements
/crew:improve "Improve error context and debugging information"
# Routes to: go-agent → architect-agent → testing-agent

# Interface design
/crew:improve "Simplify and streamline public API interfaces"
# Routes to: architect-agent → go-agent → scribe-agent
```

#### Testing Coverage
```bash
# Coverage analysis
/crew:improve "Increase test coverage to 95%+ with focus on edge cases"
# Routes to: testing-agent → go-agent → architect-agent

# Integration testing
/crew:improve "Enhance integration test reliability and speed"
# Routes to: testing-agent → cli-agent → devops-agent

# Test quality
/crew:improve "Add property-based testing for complex state machines"
# Routes to: testing-agent → go-agent → architect-agent
```

### `/crew:analyze` - Code Analysis

#### Architecture Analysis
```bash
# System design review
/crew:analyze "Review package dependencies and coupling"
# Routes to: architect-agent → go-agent → testing-agent

# Interface design analysis
/crew:analyze "Evaluate API usability and consistency"
# Routes to: architect-agent → go-agent → scribe-agent

# Performance analysis
/crew:analyze "Profile memory usage and identify bottlenecks"
# Routes to: go-agent → architect-agent → testing-agent
```

#### Security Analysis
```bash
# Security review
/crew:analyze "Audit subprocess execution for security vulnerabilities"
# Routes to: security-agent → cli-agent → testing-agent

# Credential handling
/crew:analyze "Review authentication and credential management"
# Routes to: security-agent → backend-agent → testing-agent

# Input validation
/crew:analyze "Validate input sanitization throughout the SDK"
# Routes to: security-agent → go-agent → testing-agent
```

### `/crew:test` - Testing Operations

#### Test Development
```bash
# Unit test creation
/crew:test "Generate comprehensive unit tests for core interfaces"
# Routes to: testing-agent → go-agent

# Integration test development
/crew:test "Create end-to-end tests for CLI integration scenarios"
# Routes to: testing-agent → cli-agent → security-agent

# Performance testing
/crew:test "Develop benchmarks for concurrent operations"
# Routes to: testing-agent → go-agent → cli-agent
```

#### Test Execution and Analysis
```bash
# Coverage analysis
/crew:test "Run coverage analysis and identify gaps"
# Routes to: testing-agent → go-agent

# Performance benchmarking
/crew:test "Execute performance benchmarks and analyze results"
# Routes to: testing-agent → go-agent → cli-agent

# Integration validation
/crew:test "Validate CLI integration across different environments"
# Routes to: testing-agent → cli-agent → devops-agent
```

### `/crew:document` - Documentation

#### API Documentation
```bash
# Interface documentation
/crew:document "Generate comprehensive API documentation with examples"
# Routes to: scribe-agent → go-agent → testing-agent

# Usage guides
/crew:document "Create getting started guide with common patterns"
# Routes to: scribe-agent → go-agent → cli-agent

# Architecture documentation
/crew:document "Document system architecture and design decisions"
# Routes to: scribe-agent → architect-agent → go-agent
```

#### Code Examples
```bash
# Example development
/crew:document "Create working examples for all major features"
# Routes to: scribe-agent → go-agent → testing-agent

# Tutorial creation
/crew:document "Build step-by-step tutorial for SDK integration"
# Routes to: scribe-agent → go-agent → cli-agent

# Best practices guide
/crew:document "Document Go SDK development best practices"
# Routes to: scribe-agent → go-agent → architect-agent
```

### `/crew:build` - Build and Deployment

#### Build Optimization
```bash
# Build system improvement
/crew:build "Optimize build process and reduce compilation time"
# Routes to: devops-agent → go-agent → testing-agent

# Dependency management
/crew:build "Review and optimize module dependencies"
# Routes to: devops-agent → go-agent → architect-agent

# Release preparation
/crew:build "Prepare release pipeline with automated testing"
# Routes to: devops-agent → testing-agent → scribe-agent
```

## Advanced Patterns

### Multi-Domain Commands

#### Security + Performance
```bash
# Secure optimization
/crew:improve "Optimize subprocess execution while maintaining security"
# Routes to: security-agent + go-agent → cli-agent → testing-agent
```

#### Architecture + Testing
```bash
# Test-driven architecture
/crew:implement "Design new feature with comprehensive test strategy"
# Routes to: architect-agent + testing-agent → go-agent → scribe-agent
```

#### CLI + Backend Integration
```bash
# Full-stack integration
/crew:implement "Add HTTP client with CLI fallback support"
# Routes to: backend-agent + cli-agent → go-agent → security-agent → testing-agent
```

### Workflow-Specific Patterns

#### Bug Investigation and Fix
```bash
# 1. Analysis phase
/crew:analyze "Investigate memory leak in long-running sessions"
# Routes to: go-agent → architect-agent → testing-agent

# 2. Fix implementation
/crew:implement "Fix identified memory leak with proper cleanup"
# Routes to: go-agent → testing-agent → security-agent

# 3. Validation
/crew:test "Validate memory leak fix with stress testing"
# Routes to: testing-agent → go-agent
```

#### Feature Development Cycle
```bash
# 1. Architecture design
/crew:analyze "Design architecture for new feature X"
# Routes to: architect-agent → go-agent → security-agent

# 2. Implementation
/crew:implement "Implement feature X with designed architecture"
# Routes to: go-agent → testing-agent → scribe-agent

# 3. Quality assurance
/crew:improve "Optimize and harden feature X implementation"
# Routes to: go-agent → security-agent → testing-agent

# 4. Documentation
/crew:document "Create comprehensive documentation for feature X"
# Routes to: scribe-agent → go-agent → testing-agent
```

## Command Flags and Modifiers

### Enhanced Tool Integration
```bash
# Force enhanced tools
/crew:implement "Add new feature" --code2prompt --ast-grep

# Disable enhanced tools for basic operations
/crew:improve "Fix simple bug" --no-enhanced-tools
```

### Persona Targeting
```bash
# Direct agent assignment (when needed)
/crew:implement "Go-specific optimization" --persona-go-agent

# Multi-agent coordination
/crew:implement "Security-focused feature" --persona-security-agent --persona-cli-agent
```

### Quality Gates
```bash
# Comprehensive validation
/crew:implement "Critical feature" --validate --safe-mode

# Performance focus
/crew:improve "Optimization task" --focus performance
```

## Best Practices

### Command Selection
1. **Use `/crew:implement`** for new features and functionality
2. **Use `/crew:improve`** for enhancing existing code
3. **Use `/crew:analyze`** for investigation and understanding
4. **Use `/crew:test`** for testing-specific tasks
5. **Use `/crew:document`** for documentation work

### Routing Optimization
1. **Be specific** in command descriptions for better routing
2. **Include domain keywords** (Go, CLI, security, testing)
3. **Specify scope** (package, interface, system-wide)
4. **Mention context** (performance, security, usability)

### Quality Considerations
1. **Always include testing** in feature development
2. **Consider security implications** for CLI integration
3. **Document architectural decisions** for complex features
4. **Validate performance impact** for optimization work

This command reference provides specific patterns optimized for Go Claude Code SDK development, ensuring efficient routing to appropriate agents and comprehensive coverage of development workflows.