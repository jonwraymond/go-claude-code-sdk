# Command Execution Example

This example demonstrates comprehensive command execution capabilities with Claude Code, including individual commands, slash commands, command options, file operations, Git integration, code analysis, and development workflow automation.

## What You'll Learn

- How to execute individual Claude Code commands programmatically
- Using slash command syntax for quick operations
- Command configuration with options and context
- File operation commands (read, write, edit)
- Git integration commands (status, diff, log)
- Code analysis and explanation commands
- Development workflow automation (build, test, install, refactor)

## Code Overview

The example includes seven command execution patterns:

### 1. Basic Command Execution
```go
// Create command using builder pattern
readCommand := client.ReadFile("go.mod", client.WithSummary(true))

// Execute the command
result, err := claudeClient.ExecuteCommand(ctx, readCommand)

// Process results
if result.Success {
    fmt.Printf("Command output: %s\n", result.Output)
    fmt.Printf("Truncated: %t\n", result.IsTruncated)
}
```

Demonstrates the fundamental command execution pattern with result processing.

### 2. Slash Command Execution
```go
slashCommands := []string{
    "/analyze .",
    "/search func main",
    "/explain goroutines", 
    "/read go.mod",
}

for _, slashCmd := range slashCommands {
    result, err := claudeClient.ExecuteSlashCommand(ctx, slashCmd)
    if err != nil {
        log.Printf("Slash command failed: %v", err)
        continue
    }
    
    fmt.Printf("Command: %s\n", slashCmd)
    fmt.Printf("Success: %t\n", result.Success)
    fmt.Printf("Output: %s\n", result.Output)
}
```

Shows slash command execution for quick operations and prototyping.

### 3. Commands with Options
```go
// Search command with advanced options
searchCommand := client.SearchCode(
    "interface",
    client.WithPattern("*.go"),
    client.WithLimit(5),
    client.WithContext("language", "go"),
)

// Analysis command with depth configuration
analysisCommand := client.AnalyzeCode(
    ".",
    client.WithDepth("detailed"),
    client.WithContext("focus", "architecture"),
    client.WithVerboseOutput(),
)

// Git command with formatting options
gitCommand := client.GitStatus(client.WithLimit(10))
```

Demonstrates command customization with options, context, and parameters.

### 4. File Operation Commands
```go
// Write file command
testContent := `package main

import "fmt"

func HelloWorld(name string) {
    fmt.Printf("Hello, %s!\n", name)
}

func main() {
    HelloWorld("Claude")
}`

writeCommand := client.WriteFile(testFile, testContent)
result, err := claudeClient.ExecuteCommand(ctx, writeCommand)

// Read file with summary
readCommand := client.ReadFile(testFile, client.WithSummary(true))
result, err = claudeClient.ExecuteCommand(ctx, readCommand)

// Custom edit command
editCommand := &types.Command{
    Type: client.CommandEdit,
    Args: []string{testFile, "Add error handling to the HelloWorld function"},
    Options: map[string]any{
        "backup": true,
        "format": "gofmt",
    },
    Context: map[string]any{
        "language":    "go",
        "style_guide": "effective_go",
    },
}
```

Shows comprehensive file operations including reading, writing, and editing.

### 5. Git Integration Commands
```go
// Git status with options
statusCommand := &types.Command{
    Type:    client.CommandGitStatus,
    Options: map[string]any{"short": true},
}

// Git diff with stat information
diffCommand := &types.Command{
    Type: client.CommandGitDiff,
    Args: []string{"HEAD"},
    Options: map[string]any{
        "stat":      true,
        "name-only": false,
    },
}

// Git log with formatting
logCommand := &types.Command{
    Type: client.CommandGitLog,
    Options: map[string]any{
        "limit":   5,
        "oneline": true,
    },
}
```

Demonstrates Git integration for version control operations.

### 6. Code Analysis Commands
```go
// Project structure analysis
structureCommand := &types.Command{
    Type: client.CommandAnalyze,
    Args: []string{"."},
    Options: map[string]any{
        "depth":  "structure",
        "format": "detailed",
    },
    Context: map[string]any{
        "focus":    "architecture",
        "language": "go",
    },
}

// Code quality analysis
qualityCommand := &types.Command{
    Type: client.CommandAnalyze,
    Args: []string{"pkg/"},
    Options: map[string]any{
        "depth":   "quality",
        "metrics": true,
    },
    Context: map[string]any{
        "focus":     "maintainability",
        "standards": "go_best_practices",
    },
}

// Code explanation
explainCommand := &types.Command{
    Type: client.CommandExplain,
    Args: []string{"goroutines and channels"},
    Options: map[string]any{
        "detail":   "high",
        "examples": true,
    },
    Context: map[string]any{
        "audience": "intermediate_developer",
        "format":   "tutorial",
    },
}
```

Shows comprehensive code analysis capabilities for understanding and improving code.

### 7. Development Workflow Commands
```go
// Build command with options
buildCommand := &types.Command{
    Type: client.CommandBuild,
    Args: []string{"./..."},
    Options: map[string]any{
        "verbose": true,
        "race":    true,
    },
}

// Test command with coverage
testCommand := &types.Command{
    Type: client.CommandTest,
    Args: []string{"./..."},
    Options: map[string]any{
        "type":     "unit",
        "verbose":  true,
        "coverage": true,
    },
}

// Refactor command
refactorCommand := &types.Command{
    Type: client.CommandRefactor,
    Args: []string{"pkg/client/", "improve error handling"},
    Options: map[string]any{
        "approach": "safe",
        "preview":  true,
    },
    Context: map[string]any{
        "style":  "go_best_practices",
        "safety": "high",
    },
}
```

Demonstrates automation of common development workflow tasks.

## Prerequisites

1. **Go 1.21+** installed
2. **Claude Code CLI** installed: `npm install -g @anthropic/claude-code`
3. **Git** installed (for Git integration examples)
4. **Authentication** configured

## Running the Example

### Setup
```bash
# Install Claude Code CLI
npm install -g @anthropic/claude-code

# Setup authentication
export ANTHROPIC_API_KEY="your-api-key"
# OR use subscription auth
claude setup-token

# Ensure you're in a Git repository for Git examples
git init  # if not already a Git repo
```

### Run the Example
```bash
cd examples/command_execution
go run main.go
```

## Expected Output

```
=== Command Execution Examples ===

--- Example 1: Basic Command Execution ---
✓ Command executed successfully
  Command type: read
  Success: true
  Output length: 245 characters
  Output preview:
    module github.com/jonwraymond/go-claude-code-sdk
    
    go 1.21
    
    require (
        github.com/google/uuid v1.3.0
    ... (output was truncated)

--- Example 2: Slash Command Execution ---
Executing 4 slash commands...

Slash command 1: /analyze .
  ✓ Success: true
  Output: Project Structure Analysis...

Slash command 2: /search func main
  ✓ Success: true
  Output: Found 3 matches for "func main"...

Slash command 3: /explain goroutines
  ✓ Success: true
  Output: Goroutines are Go's lightweight threads...

Slash command 4: /read go.mod
  ✓ Success: true
  Output: module github.com/jonwraymond/go-claude-code-sdk...

--- Example 3: Commands with Options ---
✓ Search command executed
  Found results: true
  Results preview:
    pkg/client/claude_code_client.go:45: type ClientInterface interface {
    pkg/types/api.go:12: type QueryRequest interface {
    pkg/auth/manager.go:8: type AuthManager interface {

✓ Analysis command executed
  Success: true
  Output length: 1247
  Truncated: false

✓ Git status command executed
  Success: true
  Status: On branch main...

--- Example 4: File Operation Commands ---
✓ File write command executed
  Success: true
  Created file: /tmp/claude_test.go

✓ File read command executed
  Success: true
  File content/summary:
    package main
    
    import "fmt"
    
    // HelloWorld prints a greeting message
    func HelloWorld(name string) {

✓ File edit command executed
  Success: true
  Edit result: Updated HelloWorld function with error handling...

✓ Cleaned up test file

--- Example 5: Git Integration Commands ---
✓ Git status command executed
  Success: true
  Status:
    M examples/README.md
    M pkg/client/commands.go
    ?? examples/command_execution/README.md

✓ Git diff command executed
  Success: true
  Diff summary:  examples/README.md | 12 ++++++++++--
   pkg/client/commands.go | 8 ++++++--
   2 files changed, 16 insertions(+), 4 deletions(-)

✓ Git log command executed
  Success: true
  Recent commits:
    a1b2c3d feat: implement command execution examples
    e4f5g6h docs: update README with command examples
    h7i8j9k fix: improve error handling in commands

--- Example 6: Code Analysis Commands ---
✓ Project structure analysis executed
  Success: true
  Analysis:
    Project follows Go standard layout
    - cmd/ (command-line applications)
    - pkg/ (library code)
    - examples/ (usage examples)
    - tests/ (test utilities)
    Strong separation of concerns
    Well-organized package structure

✓ Code quality analysis executed
  Success: true
  Quality report: Overall code quality: GOOD
  Maintainability Index: 85/100
  - Clear naming conventions
  - Appropriate error handling
  - Good test coverage (78%)

✓ Code explanation executed
  Success: true
  Explanation:
    Goroutines and Channels in Go
    
    Goroutines are lightweight threads managed by the Go runtime...
    Channels provide communication between goroutines...
    Example usage patterns:
      - Producer-consumer pattern
      - Fan-out/fan-in pattern

--- Example 7: Development Workflow Commands ---
✓ Build command executed
  Success: true
  Build completed successfully

✓ Test command executed
  Success: true
  Test results:
    PASS: TestClaudeCodeClient (0.01s)
    PASS: TestQueryRequest (0.00s)
    PASS: TestAuthMethods (0.02s)
    coverage: 78.5% of statements

✓ Install command executed
  Success: true
  Dependencies installed

✓ Refactor command executed
  Success: true
  Refactor suggestions: Safe refactoring opportunities identified:
  1. Extract error handling patterns into utility functions
  2. Consolidate duplicate validation logic
```

## Key Concepts

### Command Types

The SDK supports various command types:

| Command | Purpose | Example |
|---------|---------|---------|
| `read` | Read files | `client.ReadFile("main.go")` |
| `write` | Write files | `client.WriteFile("test.go", content)` |
| `edit` | Edit files | `client.EditFile("main.go", "add logging")` |
| `search` | Search code | `client.SearchCode("interface")` |
| `analyze` | Analyze code | `client.AnalyzeCode(".")` |
| `explain` | Explain concepts | `client.ExplainCode("goroutines")` |
| `git_status` | Git status | `client.GitStatus()` |
| `git_diff` | Git diff | `client.GitDiff("HEAD")` |
| `build` | Build project | `client.BuildProject("./...")` |
| `test` | Run tests | `client.RunTests("./...")` |

### Command Structure

```go
type Command struct {
    Type          string         `json:"type"`
    Args          []string       `json:"args,omitempty"`
    Options       map[string]any `json:"options,omitempty"`
    Context       map[string]any `json:"context,omitempty"`
    VerboseOutput bool           `json:"verbose_output,omitempty"`
}

type CommandResult struct {
    Command      *Command `json:"command"`
    Success      bool     `json:"success"`
    Output       string   `json:"output"`
    Error        string   `json:"error,omitempty"`
    OutputLength int      `json:"output_length"`
    IsTruncated  bool     `json:"is_truncated"`
    FullOutput   string   `json:"full_output,omitempty"`
    Metadata     any      `json:"metadata,omitempty"`
}
```

### Builder Pattern Usage

```go
// Use builder functions for common commands
readCmd := client.ReadFile("main.go", 
    client.WithSummary(true),
    client.WithLimit(100),
)

searchCmd := client.SearchCode("interface",
    client.WithPattern("*.go"),
    client.WithContext("language", "go"),
    client.WithLimit(10),
)

analyzeCmd := client.AnalyzeCode(".",
    client.WithDepth("detailed"),
    client.WithVerboseOutput(),
)
```

## Advanced Usage

### Custom Command Creation
```go
customCommand := &types.Command{
    Type: "custom_analysis",
    Args: []string{"pkg/", "analyze performance"},
    Options: map[string]any{
        "depth":      "deep",
        "focus":      []string{"loops", "allocations"},
        "benchmark":  true,
        "suggest":    true,
    },
    Context: map[string]any{
        "language":     "go",
        "version":      "1.21",
        "target_arch":  "amd64",
        "optimization": "speed",
    },
    VerboseOutput: true,
}

result, err := claudeClient.ExecuteCommand(ctx, customCommand)
```

### Command Pipelines
```go
func executeCommandPipeline(commands []*types.Command) error {
    for i, cmd := range commands {
        fmt.Printf("Executing command %d/%d: %s\n", i+1, len(commands), cmd.Type)
        
        result, err := claudeClient.ExecuteCommand(ctx, cmd)
        if err != nil {
            return fmt.Errorf("command %d failed: %w", i+1, err)
        }
        
        if !result.Success {
            return fmt.Errorf("command %d unsuccessful: %s", i+1, result.Error)
        }
        
        // Use result for next command if needed
        if i < len(commands)-1 && len(result.Output) > 0 {
            commands[i+1].Context["previous_output"] = result.Output
        }
    }
    
    return nil
}
```

### Parallel Command Execution
```go
func executeCommandsConcurrently(commands []*types.Command) []CommandResult {
    results := make([]CommandResult, len(commands))
    var wg sync.WaitGroup
    
    for i, cmd := range commands {
        wg.Add(1)
        go func(index int, command *types.Command) {
            defer wg.Done()
            
            result, err := claudeClient.ExecuteCommand(ctx, command)
            if err != nil {
                results[index] = CommandResult{
                    Command: command,
                    Success: false,
                    Error:   err.Error(),
                }
                return
            }
            
            results[index] = CommandResult{
                Command: result.Command,
                Success: result.Success,
                Output:  result.Output,
                Error:   result.Error,
            }
        }(i, cmd)
    }
    
    wg.Wait()
    return results
}
```

## Best Practices

### 1. Error Handling
```go
result, err := claudeClient.ExecuteCommand(ctx, command)
if err != nil {
    log.Printf("Command execution failed: %v", err)
    return
}

if !result.Success {
    log.Printf("Command unsuccessful: %s", result.Error)
    return
}

// Check for truncated output
if result.IsTruncated && result.FullOutput != "" {
    content := result.FullOutput
} else {
    content := result.Output
}
```

### 2. Context and Timeouts
```go
// Set appropriate timeouts for different command types
var timeout time.Duration
switch command.Type {
case "analyze", "build":
    timeout = 2 * time.Minute
case "test":
    timeout = 5 * time.Minute
default:
    timeout = 30 * time.Second
}

ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()

result, err := claudeClient.ExecuteCommand(ctx, command)
```

### 3. Command Validation
```go
func validateCommand(cmd *types.Command) error {
    if cmd.Type == "" {
        return fmt.Errorf("command type is required")
    }
    
    // Validate required args for specific commands
    switch cmd.Type {
    case "read", "write":
        if len(cmd.Args) == 0 {
            return fmt.Errorf("%s command requires file path", cmd.Type)
        }
    case "search":
        if len(cmd.Args) == 0 {
            return fmt.Errorf("search command requires search pattern")
        }
    }
    
    return nil
}
```

### 4. Output Processing
```go
func processCommandOutput(result *types.CommandResult) {
    if !result.Success {
        fmt.Printf("❌ Command failed: %s\n", result.Error)
        return
    }
    
    fmt.Printf("✅ Command succeeded\n")
    
    // Handle different output types
    switch result.Command.Type {
    case "analyze":
        processAnalysisOutput(result.Output)
    case "test":
        processTestOutput(result.Output)
    case "git_status":
        processGitOutput(result.Output)
    default:
        fmt.Printf("Output: %s\n", result.Output)
    }
    
    // Show metadata if available
    if result.Metadata != nil {
        fmt.Printf("Metadata: %v\n", result.Metadata)
    }
}
```

## Integration Patterns

### CI/CD Integration
```go
func runCICommands() error {
    commands := []*types.Command{
        client.BuildProject("./..."),
        client.RunTests("./...", client.WithCoverage(true)),
        client.RunLinter(),
        client.SecurityScan(),
    }
    
    for _, cmd := range commands {
        result, err := claudeClient.ExecuteCommand(ctx, cmd)
        if err != nil || !result.Success {
            return fmt.Errorf("CI command failed: %s", cmd.Type)
        }
    }
    
    return nil
}
```

### Development Automation
```go
func autoFormat() error {
    files, err := findGoFiles(".")
    if err != nil {
        return err
    }
    
    for _, file := range files {
        cmd := client.FormatFile(file)
        result, err := claudeClient.ExecuteCommand(ctx, cmd)
        if err != nil || !result.Success {
            log.Printf("Failed to format %s: %v", file, err)
        }
    }
    
    return nil
}
```

## Troubleshooting

### Common Issues

1. **Command Not Found**
   - Verify Claude Code CLI is installed and accessible
   - Check PATH environment variable
   - Test with `claude --version`

2. **Permission Errors**
   - Ensure proper file permissions for file operations
   - Check directory access rights
   - Verify Git repository permissions

3. **Timeout Issues**
   - Increase context timeout for long-running commands
   - Consider breaking complex commands into smaller parts
   - Use async patterns for time-consuming operations

### Debug Commands
```go
// Enable debug mode
config.Debug = true

// Add debug context to commands
command.Context["debug"] = true
command.VerboseOutput = true

// Log command details
log.Printf("Executing command: type=%s, args=%v", command.Type, command.Args)
```

## Performance Tips

1. **Use Appropriate Command Types**: Choose the most specific command for your task
2. **Limit Output**: Use options to limit output size for large results
3. **Parallel Execution**: Run independent commands concurrently
4. **Cache Results**: Cache command results for repeated operations
5. **Timeout Management**: Set appropriate timeouts based on command complexity

## Next Steps

After mastering command execution, explore:
- [Basic Client](../basic_client/) - Client configuration
- [Sync Queries](../sync_queries/) - Direct API queries
- [Streaming Queries](../streaming_queries/) - Real-time responses  
- [Session Lifecycle](../session_lifecycle/) - Session management

## Related Documentation

- [Command Types](../../pkg/types/commands.go)
- [Client Package](../../pkg/client/)
- [Development Guide](../../docs/development.md)