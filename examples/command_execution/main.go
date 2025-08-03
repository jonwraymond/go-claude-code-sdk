// Package main demonstrates command execution capabilities of Claude Code SDK.
// This example covers individual commands, slash commands, and command lists.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Command Execution Examples ===")

	// Setup client for all examples
	claudeClient := setupClient()
	if claudeClient == nil {
		return
	}
	defer claudeClient.Close()

	// Example 1: Basic command execution
	basicCommandExample(claudeClient)

	// Example 2: Slash command execution
	slashCommandExample(claudeClient)

	// Example 3: Command with options
	commandOptionsExample(claudeClient)

	// Example 4: File operation commands
	fileOperationCommands(claudeClient)

	// Example 5: Git integration commands
	gitCommandsExample(claudeClient)

	// Example 6: Code analysis commands
	codeAnalysisCommands(claudeClient)

	// Example 7: Development workflow commands
	developmentWorkflowCommands(claudeClient)
}

// setupClient creates a configured Claude Code client
func setupClient() *client.ClaudeCodeClient {
	ctx := context.Background()
	config := types.NewClaudeCodeConfig()

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
		config.AuthMethod = types.AuthTypeAPIKey
	} else {
		config.AuthMethod = types.AuthTypeSubscription
		fmt.Println("Note: Using subscription auth (set ANTHROPIC_API_KEY for API key auth)")
	}

	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create Claude client: %v", err)
		return nil
	}

	return claudeClient
}

// basicCommandExample demonstrates basic command execution
func basicCommandExample(claudeClient *client.ClaudeCodeClient) {
	fmt.Println("--- Example 1: Basic Command Execution ---")

	ctx := context.Background()

	// Create a basic command using the builder pattern
	readCommand := client.ReadFile("go.mod", client.WithSummary(true))

	// Execute the command
	result, err := claudeClient.ExecuteCommand(ctx, readCommand)
	if err != nil {
		log.Printf("Command execution failed: %v", err)
		return
	}

	fmt.Printf("✓ Command executed successfully\n")
	fmt.Printf("  Command type: %s\n", result.Command.Type)
	fmt.Printf("  Success: %t\n", result.Success)
	fmt.Printf("  Output length: %d characters\n", result.OutputLength)

	if result.Success && result.Output != "" {
		// Show first few lines of output
		lines := splitLines(result.Output, 5)
		fmt.Printf("  Output preview:\n")
		for _, line := range lines {
			fmt.Printf("    %s\n", line)
		}
		if result.IsTruncated {
			fmt.Printf("    ... (output was truncated)\n")
		}
	}

	if !result.Success && result.Error != "" {
		fmt.Printf("  Error: %s\n", result.Error)
	}

	fmt.Println()
}

// slashCommandExample demonstrates slash command execution
func slashCommandExample(claudeClient *client.ClaudeCodeClient) {
	fmt.Println("--- Example 2: Slash Command Execution ---")

	ctx := context.Background()

	// Define various slash commands to try
	slashCommands := []string{
		"/analyze .",
		"/search func main",
		"/explain goroutines",
		"/read go.mod",
	}

	fmt.Printf("Executing %d slash commands...\n", len(slashCommands))

	for i, slashCmd := range slashCommands {
		fmt.Printf("\nSlash command %d: %s\n", i+1, slashCmd)

		result, err := claudeClient.ExecuteSlashCommand(ctx, slashCmd)
		if err != nil {
			log.Printf("Slash command failed: %v", err)
			continue
		}

		fmt.Printf("  ✓ Success: %t\n", result.Success)

		if result.Success && result.Output != "" {
			preview := result.Output
			if len(preview) > 150 {
				preview = preview[:150] + "..."
			}
			fmt.Printf("  Output: %s\n", preview)
		}

		if !result.Success && result.Error != "" {
			fmt.Printf("  Error: %s\n", result.Error)
		}

		// Show execution metadata if available
		if result.Metadata != nil {
			fmt.Printf("  Metadata: %v\n", result.Metadata)
		}
	}

	fmt.Println()
}

// commandOptionsExample demonstrates commands with various options
func commandOptionsExample(claudeClient *client.ClaudeCodeClient) {
	fmt.Println("--- Example 3: Commands with Options ---")

	ctx := context.Background()

	// Example 1: Search command with pattern
	searchCommand := client.SearchCode(
		"interface",
		client.WithPattern("*.go"),
		client.WithLimit(5),
		client.WithContext("language", "go"),
	)

	result, err := claudeClient.ExecuteCommand(ctx, searchCommand)
	if err != nil {
		log.Printf("Search command failed: %v", err)
	} else {
		fmt.Printf("✓ Search command executed\n")
		fmt.Printf("  Found results: %t\n", result.Success)
		if result.Success && len(result.Output) > 0 {
			lines := splitLines(result.Output, 3)
			fmt.Printf("  Results preview:\n")
			for _, line := range lines {
				fmt.Printf("    %s\n", line)
			}
		}
	}

	// Example 2: Analysis command with depth
	analysisCommand := client.AnalyzeCode(
		".",
		client.WithDepth("detailed"),
		client.WithContext("focus", "architecture"),
		client.WithVerboseOutput(),
	)

	result, err = claudeClient.ExecuteCommand(ctx, analysisCommand)
	if err != nil {
		log.Printf("Analysis command failed: %v", err)
	} else {
		fmt.Printf("\n✓ Analysis command executed\n")
		fmt.Printf("  Success: %t\n", result.Success)
		fmt.Printf("  Output length: %d\n", result.OutputLength)
		fmt.Printf("  Truncated: %t\n", result.IsTruncated)

		if result.IsTruncated && result.FullOutput != "" {
			fmt.Printf("  Full output available: %d characters\n", len(result.FullOutput))
		}
	}

	// Example 3: Git status command with limit
	gitCommand := client.GitStatus(client.WithLimit(10))

	result, err = claudeClient.ExecuteCommand(ctx, gitCommand)
	if err != nil {
		log.Printf("Git command failed: %v", err)
	} else {
		fmt.Printf("\n✓ Git status command executed\n")
		fmt.Printf("  Success: %t\n", result.Success)
		if result.Success && len(result.Output) > 0 {
			preview := result.Output
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			fmt.Printf("  Status: %s\n", preview)
		}
	}

	fmt.Println()
}

// fileOperationCommands demonstrates file-related commands
func fileOperationCommands(claudeClient *client.ClaudeCodeClient) {
	fmt.Println("--- Example 4: File Operation Commands ---")

	ctx := context.Background()

	// Create a temporary file for demonstration
	testFile := "/tmp/claude_test.go"
	testContent := `package main

import "fmt"

// HelloWorld prints a greeting message
func HelloWorld(name string) {
    fmt.Printf("Hello, %s!\n", name)
}

func main() {
    HelloWorld("Claude")
}`

	// Example 1: Write file command
	writeCommand := client.WriteFile(testFile, testContent)

	result, err := claudeClient.ExecuteCommand(ctx, writeCommand)
	if err != nil {
		log.Printf("Write command failed: %v", err)
	} else {
		fmt.Printf("✓ File write command executed\n")
		fmt.Printf("  Success: %t\n", result.Success)
		if result.Success {
			fmt.Printf("  Created file: %s\n", testFile)
		}
	}

	// Example 2: Read file command with summary
	readCommand := client.ReadFile(testFile, client.WithSummary(true))

	result, err = claudeClient.ExecuteCommand(ctx, readCommand)
	if err != nil {
		log.Printf("Read command failed: %v", err)
	} else {
		fmt.Printf("\n✓ File read command executed\n")
		fmt.Printf("  Success: %t\n", result.Success)
		if result.Success && len(result.Output) > 0 {
			fmt.Printf("  File content/summary:\n")
			lines := splitLines(result.Output, 5)
			for _, line := range lines {
				fmt.Printf("    %s\n", line)
			}
		}
	}

	// Example 3: Custom command for file editing
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
		VerboseOutput: false,
	}

	result, err = claudeClient.ExecuteCommand(ctx, editCommand)
	if err != nil {
		log.Printf("Edit command failed: %v", err)
	} else {
		fmt.Printf("\n✓ File edit command executed\n")
		fmt.Printf("  Success: %t\n", result.Success)
		if result.Success && len(result.Output) > 0 {
			preview := result.Output
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			fmt.Printf("  Edit result: %s\n", preview)
		}
	}

	// Clean up
	_ = os.Remove(testFile) // Ignore error, best effort cleanup
	fmt.Printf("✓ Cleaned up test file\n")

	fmt.Println()
}

// gitCommandsExample demonstrates Git integration commands
func gitCommandsExample(claudeClient *client.ClaudeCodeClient) {
	fmt.Println("--- Example 5: Git Integration Commands ---")

	ctx := context.Background()

	// Example 1: Git status
	statusCommand := &types.Command{
		Type:    client.CommandGitStatus,
		Options: map[string]any{"short": true},
	}

	result, err := claudeClient.ExecuteCommand(ctx, statusCommand)
	if err != nil {
		log.Printf("Git status failed: %v", err)
	} else {
		fmt.Printf("✓ Git status command executed\n")
		fmt.Printf("  Success: %t\n", result.Success)
		if result.Success && len(result.Output) > 0 {
			lines := splitLines(result.Output, 5)
			fmt.Printf("  Status:\n")
			for _, line := range lines {
				fmt.Printf("    %s\n", line)
			}
		}
	}

	// Example 2: Git diff
	diffCommand := &types.Command{
		Type: client.CommandGitDiff,
		Args: []string{"HEAD"},
		Options: map[string]any{
			"stat":      true,
			"name-only": false,
		},
	}

	result, err = claudeClient.ExecuteCommand(ctx, diffCommand)
	if err != nil {
		log.Printf("Git diff failed: %v", err)
	} else {
		fmt.Printf("\n✓ Git diff command executed\n")
		fmt.Printf("  Success: %t\n", result.Success)
		if result.Success && len(result.Output) > 0 {
			preview := result.Output
			if len(preview) > 300 {
				preview = preview[:300] + "..."
			}
			fmt.Printf("  Diff summary: %s\n", preview)
		}
	}

	// Example 3: Git log with limit
	logCommand := &types.Command{
		Type: client.CommandGitLog,
		Options: map[string]any{
			"limit":   5,
			"oneline": true,
		},
	}

	result, err = claudeClient.ExecuteCommand(ctx, logCommand)
	if err != nil {
		log.Printf("Git log failed: %v", err)
	} else {
		fmt.Printf("\n✓ Git log command executed\n")
		fmt.Printf("  Success: %t\n", result.Success)
		if result.Success && len(result.Output) > 0 {
			lines := splitLines(result.Output, 5)
			fmt.Printf("  Recent commits:\n")
			for _, line := range lines {
				fmt.Printf("    %s\n", line)
			}
		}
	}

	fmt.Println()
}

// codeAnalysisCommands demonstrates code analysis commands
func codeAnalysisCommands(claudeClient *client.ClaudeCodeClient) {
	fmt.Println("--- Example 6: Code Analysis Commands ---")

	ctx := context.Background()

	// Example 1: Analyze project structure
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

	result, err := claudeClient.ExecuteCommand(ctx, structureCommand)
	if err != nil {
		log.Printf("Structure analysis failed: %v", err)
	} else {
		fmt.Printf("✓ Project structure analysis executed\n")
		fmt.Printf("  Success: %t\n", result.Success)
		if result.Success && len(result.Output) > 0 {
			lines := splitLines(result.Output, 8)
			fmt.Printf("  Analysis:\n")
			for _, line := range lines {
				fmt.Printf("    %s\n", line)
			}
		}
	}

	// Example 2: Code quality analysis
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

	result, err = claudeClient.ExecuteCommand(ctx, qualityCommand)
	if err != nil {
		log.Printf("Quality analysis failed: %v", err)
	} else {
		fmt.Printf("\n✓ Code quality analysis executed\n")
		fmt.Printf("  Success: %t\n", result.Success)
		if result.Success && len(result.Output) > 0 {
			preview := result.Output
			if len(preview) > 250 {
				preview = preview[:250] + "..."
			}
			fmt.Printf("  Quality report: %s\n", preview)
		}
	}

	// Example 3: Explain complex code
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

	result, err = claudeClient.ExecuteCommand(ctx, explainCommand)
	if err != nil {
		log.Printf("Explain command failed: %v", err)
	} else {
		fmt.Printf("\n✓ Code explanation executed\n")
		fmt.Printf("  Success: %t\n", result.Success)
		if result.Success && len(result.Output) > 0 {
			lines := splitLines(result.Output, 6)
			fmt.Printf("  Explanation:\n")
			for _, line := range lines {
				fmt.Printf("    %s\n", line)
			}
		}
	}

	fmt.Println()
}

// developmentWorkflowCommands demonstrates development workflow commands
func developmentWorkflowCommands(claudeClient *client.ClaudeCodeClient) {
	fmt.Println("--- Example 7: Development Workflow Commands ---")

	ctx := context.Background()

	// Example 1: Build command
	buildCommand := &types.Command{
		Type: client.CommandBuild,
		Args: []string{"./..."},
		Options: map[string]any{
			"verbose": true,
			"race":    true,
		},
	}

	result, err := claudeClient.ExecuteCommand(ctx, buildCommand)
	if err != nil {
		log.Printf("Build command failed: %v", err)
	} else {
		fmt.Printf("✓ Build command executed\n")
		fmt.Printf("  Success: %t\n", result.Success)
		if result.Success {
			fmt.Printf("  Build completed successfully\n")
		} else if result.Error != "" {
			fmt.Printf("  Build errors: %s\n", result.Error)
		}
	}

	// Example 2: Test command
	testCommand := &types.Command{
		Type: client.CommandTest,
		Args: []string{"./..."},
		Options: map[string]any{
			"type":     "unit",
			"verbose":  true,
			"coverage": true,
		},
	}

	result, err = claudeClient.ExecuteCommand(ctx, testCommand)
	if err != nil {
		log.Printf("Test command failed: %v", err)
	} else {
		fmt.Printf("\n✓ Test command executed\n")
		fmt.Printf("  Success: %t\n", result.Success)
		if result.Success && len(result.Output) > 0 {
			lines := splitLines(result.Output, 5)
			fmt.Printf("  Test results:\n")
			for _, line := range lines {
				fmt.Printf("    %s\n", line)
			}
		}
	}

	// Example 3: Install dependencies
	installCommand := &types.Command{
		Type: client.CommandInstall,
		Args: []string{"dependencies"},
		Options: map[string]any{
			"update": false,
			"clean":  true,
		},
	}

	result, err = claudeClient.ExecuteCommand(ctx, installCommand)
	if err != nil {
		log.Printf("Install command failed: %v", err)
	} else {
		fmt.Printf("\n✓ Install command executed\n")
		fmt.Printf("  Success: %t\n", result.Success)
		if result.Success {
			fmt.Printf("  Dependencies installed\n")
		}
	}

	// Example 4: Refactor command
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

	result, err = claudeClient.ExecuteCommand(ctx, refactorCommand)
	if err != nil {
		log.Printf("Refactor command failed: %v", err)
	} else {
		fmt.Printf("\n✓ Refactor command executed\n")
		fmt.Printf("  Success: %t\n", result.Success)
		if result.Success && len(result.Output) > 0 {
			preview := result.Output
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			fmt.Printf("  Refactor suggestions: %s\n", preview)
		}
	}

	fmt.Println()
}

// Helper function to split text into limited lines
func splitLines(text string, maxLines int) []string {
	lines := strings.Split(text, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, fmt.Sprintf("... (%d more lines)", len(strings.Split(text, "\n"))-maxLines))
	}
	return lines
}
