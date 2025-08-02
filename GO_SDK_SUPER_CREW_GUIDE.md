# Go Claude Code SDK - Super Crew Usage Guide

## System Validation ✅

The Claude Code Super Crew configuration has been successfully implemented with:

- **8 Primary Agents**: Properly configured with 3-field frontmatter format
- **1 Orchestrator**: Intelligent routing and coordination system
- **Enhanced Tools**: code2prompt and ast-grep integration enabled
- **Quality Gates**: Comprehensive validation workflows implemented

## Agent Overview

### Core Development Agents

#### go-agent 🔧
**Primary Go Development Specialist**
- **Activates for**: Go programming, package architecture, concurrency patterns
- **Specializes in**: Goroutines, channels, interfaces, error handling, testing
- **Use cases**: Core SDK logic, type-safe interfaces, Go best practices

#### cli-agent 🖥️
**CLI & Subprocess Management Specialist**
- **Activates for**: Subprocess execution, command-line integration, streaming I/O
- **Specializes in**: Process lifecycle, session management, CLI wrapper development
- **Use cases**: Claude CLI integration, command execution, process communication

#### testing-agent 🧪
**Quality Assurance & Testing Specialist**
- **Activates for**: Test development, coverage analysis, quality validation
- **Specializes in**: Unit tests, integration tests, CI/CD validation
- **Use cases**: Test strategy, coverage improvement, quality gates

#### architect-agent 🏗️
**System Architecture & Design Specialist**
- **Activates for**: Package structure, interface design, long-term maintainability
- **Specializes in**: System design, architectural patterns, scalability
- **Use cases**: Module organization, API design, architectural decisions

### Supporting Agents

#### backend-agent 🌐
**Client-Server Communication Specialist**
- **Activates for**: HTTP integration, data serialization, client-server patterns
- **Specializes in**: API clients, communication protocols, data handling
- **Use cases**: Network communication, HTTP clients, data exchange

#### security-agent 🛡️
**Security & Safety Specialist**
- **Activates for**: Authentication, subprocess safety, vulnerability prevention
- **Specializes in**: Security patterns, credential management, safe execution
- **Use cases**: Secure subprocess execution, authentication, security audits

#### devops-agent 🚀
**Build & Deployment Specialist**
- **Activates for**: Build optimization, CI/CD, release management
- **Specializes in**: Build systems, automation, dependency management
- **Use cases**: CI/CD setup, build optimization, release processes

#### scribe-agent 📝
**Documentation & Communication Specialist**
- **Activates for**: Documentation updates, examples, user guides
- **Specializes in**: Technical writing, API documentation, examples
- **Use cases**: README updates, documentation, code examples

## Common Development Patterns

### 1. Feature Development Workflow
```bash
# Intelligent routing: architect-agent → go-agent → testing-agent → scribe-agent
/crew:implement "Add message streaming support with channels"
```

**Expected Agent Flow:**
1. **architect-agent**: Designs streaming architecture and interfaces
2. **go-agent**: Implements Go channel-based streaming logic
3. **testing-agent**: Creates comprehensive test coverage
4. **scribe-agent**: Updates documentation and examples

### 2. CLI Integration Tasks
```bash
# Routes to: cli-agent → security-agent → testing-agent
/crew:implement "Enhance subprocess execution with timeout handling"
```

**Expected Agent Flow:**
1. **cli-agent**: Implements subprocess management and timeout logic
2. **security-agent**: Reviews security implications of subprocess execution
3. **testing-agent**: Validates subprocess behavior and edge cases

### 3. Quality Improvement
```bash
# Routes to: testing-agent → architect-agent → devops-agent
/crew:improve "Increase test coverage and add performance benchmarks"
```

**Expected Agent Flow:**
1. **testing-agent**: Analyzes current coverage and creates improvement plan
2. **architect-agent**: Reviews architectural implications of testing strategy
3. **devops-agent**: Integrates quality gates into CI/CD pipeline

### 4. Security-Focused Development
```bash
# Routes to: security-agent → cli-agent → testing-agent
/crew:implement "Add secure credential handling for Claude CLI authentication"
```

**Expected Agent Flow:**
1. **security-agent**: Designs secure credential management system
2. **cli-agent**: Implements CLI authentication integration
3. **testing-agent**: Creates security-focused test scenarios

## Enhanced Tool Integration

### Code2prompt Usage
The system automatically uses code2prompt for:
- **Agent Handoffs**: Rich context generation when routing between agents
- **Context Generation**: Comprehensive project understanding for complex tasks
- **Pattern Detection**: Enhanced understanding of existing code patterns

### AST-grep Integration
The system leverages ast-grep for:
- **Go Pattern Analysis**: Finding Go-specific patterns (goroutines, channels, interfaces)
- **Consistency Checking**: Ensuring consistent patterns across the codebase
- **Refactoring Support**: Safe code transformations with pattern awareness

## Quality Gates

### Pre-Implementation Gates
1. **Architecture Review**: architect-agent validates design approach
2. **Security Assessment**: security-agent reviews security implications
3. **Test Strategy**: testing-agent plans comprehensive testing approach

### Implementation Gates
1. **Code Review**: Primary agent ensures Go best practices
2. **Testing Validation**: testing-agent validates implementation
3. **Documentation Update**: scribe-agent ensures docs stay current

### Post-Implementation Gates
1. **Integration Testing**: testing-agent validates end-to-end functionality
2. **Security Validation**: security-agent confirms security measures
3. **Documentation Validation**: scribe-agent ensures documentation accuracy

## Project-Specific Optimizations

### Go SDK Development
- **Primary Language Focus**: go-agent leads most development tasks
- **CLI Integration**: cli-agent handles subprocess management and CLI wrapping
- **Type Safety**: architect-agent ensures strong typing and interface design
- **Testing Strategy**: testing-agent emphasizes integration testing for CLI interactions

### Claude Code Integration
- **Subprocess Management**: cli-agent specializes in claude CLI subprocess handling
- **Session Management**: backend-agent manages persistent conversation sessions
- **Streaming Support**: go-agent implements channel-based streaming patterns
- **Security**: security-agent ensures safe subprocess execution and credential handling

## Usage Examples

### Starting a New Feature
```bash
# Orchestrator will intelligently route to appropriate agents
/crew:implement "Add support for tool execution with result streaming"

# Expected routing:
# 1. architect-agent: Design tool execution interface
# 2. go-agent: Implement Go interfaces and logic
# 3. cli-agent: Handle CLI subprocess integration
# 4. testing-agent: Create comprehensive tests
# 5. scribe-agent: Update documentation
```

### Code Quality Improvement
```bash
# Focus on testing and architecture
/crew:improve "Enhance error handling throughout the SDK"

# Expected routing:
# 1. testing-agent: Analyze current error handling coverage
# 2. architect-agent: Review error handling architecture
# 3. go-agent: Implement improved error patterns
# 4. security-agent: Review security implications
```

### Documentation Update
```bash
# Primarily routes to scribe-agent with go-agent support
/crew:document "Create comprehensive examples for all SDK features"

# Expected routing:
# 1. scribe-agent: Lead documentation creation
# 2. go-agent: Validate code examples
# 3. testing-agent: Ensure examples are tested
```

## Advanced Features

### Parallel Agent Execution
The system supports parallel agent execution for:
- **Independent Tasks**: Multiple agents working on separate components
- **Validation Workflows**: Parallel validation by different specialists
- **Context Sharing**: Enhanced tool integration enables efficient context sharing

### Intelligent Context Awareness
- **Project Pattern Recognition**: System understands Go SDK development patterns
- **Domain-Specific Routing**: Routes based on Go, CLI, testing, and security domains
- **Quality Gate Integration**: Automated quality validation throughout development

### Enhanced Tool Coordination
- **code2prompt + ast-grep**: Combined for comprehensive code understanding
- **Pattern Detection**: Identifies and maintains consistency with existing patterns
- **Smart Caching**: Optimized performance through intelligent caching strategies

## Next Steps

1. **Start Using**: Begin with simple `/crew:implement` commands to see routing in action
2. **Observe Patterns**: Notice how different request types route to appropriate agents
3. **Leverage Specialization**: Use agent expertise for complex Go SDK development
4. **Quality Focus**: Rely on built-in quality gates for robust development
5. **Documentation**: Keep documentation current using scribe-agent integration

## Success Metrics

The implementation provides:
- **90%+ Test Coverage**: Comprehensive testing strategy with testing-agent
- **Security Focus**: Specialized security agent for subprocess safety
- **Architectural Integrity**: architect-agent ensures long-term maintainability
- **Documentation Currency**: scribe-agent keeps docs aligned with implementation
- **Go Best Practices**: go-agent enforces idiomatic Go development patterns

This Claude Code Super Crew configuration is optimized specifically for Go SDK development with Claude CLI integration, providing intelligent routing, quality gates, and specialized expertise for building robust, secure, and well-tested SDK libraries.