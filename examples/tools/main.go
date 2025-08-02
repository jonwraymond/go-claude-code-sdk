// Tool system example demonstrating file operations and code search
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	// Setup client
	config := types.NewClaudeCodeConfig()
	config.APIKey = os.Getenv("ANTHROPIC_API_KEY")

	claudeClient, err := client.NewClaudeCodeClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()

	// Example 1: List available tools
	fmt.Println("=== Example 1: Available Tools ===")
	listTools(claudeClient)

	// Example 2: Direct tool execution
	fmt.Println("\n=== Example 2: Direct Tool Execution ===")
	directToolExecution(claudeClient)

	// Example 3: Tools in conversation
	fmt.Println("\n=== Example 3: Tools in Conversation ===")
	toolsInConversation(claudeClient)

	// Example 4: Custom tool usage
	fmt.Println("\n=== Example 4: Custom Tool Usage ===")
	customToolUsage(claudeClient)
}

func listTools(client *client.ClaudeCodeClient) {
	// Get all available tools
	tools := client.Tools().GetAvailableTools()

	fmt.Println("Built-in tools:")
	for name, tool := range tools {
		fmt.Printf("- %s: %s\n", name, tool.Description)
		
		// Show tool parameters
		if schema, ok := tool.InputSchema["properties"].(map[string]interface{}); ok {
			fmt.Println("  Parameters:")
			for param, details := range schema {
				if paramInfo, ok := details.(map[string]interface{}); ok {
					fmt.Printf("    - %s: %s\n", param, paramInfo["description"])
				}
			}
		}
	}
}

func directToolExecution(client *client.ClaudeCodeClient) {
	ctx := context.Background()

	// Example 1: Read a file
	fmt.Println("\nReading go.mod file:")
	readResult, err := client.Tools().ExecuteTool(ctx, &client.ClaudeCodeTool{
		Name: "read_file",
		Arguments: map[string]interface{}{
			"path": "go.mod",
		},
	})
	if err != nil {
		log.Printf("Failed to read file: %v", err)
	} else {
		fmt.Printf("Content:\n%s\n", readResult.Output)
	}

	// Example 2: Search for code patterns
	fmt.Println("\nSearching for function definitions:")
	searchResult, err := client.Tools().ExecuteTool(ctx, &client.ClaudeCodeTool{
		Name: "search_code",
		Arguments: map[string]interface{}{
			"pattern": "func ",
			"path":    "./pkg",
		},
	})
	if err != nil {
		log.Printf("Failed to search: %v", err)
	} else {
		fmt.Printf("Search results:\n%s\n", searchResult.Output)
	}

	// Example 3: List directory contents
	fmt.Println("\nListing directory contents:")
	listResult, err := client.Tools().ExecuteTool(ctx, &client.ClaudeCodeTool{
		Name: "list_directory",
		Arguments: map[string]interface{}{
			"path": ".",
		},
	})
	if err != nil {
		log.Printf("Failed to list directory: %v", err)
	} else {
		fmt.Printf("Directory contents:\n%s\n", listResult.Output)
	}
}

func toolsInConversation(client *client.ClaudeCodeClient) {
	ctx := context.Background()

	// Create a temporary test file
	testFile := "test_example.go"
	testContent := `package main

import "fmt"

func greet(name string) {
    fmt.Printf("Hello, %s!\n", name)
}

func main() {
    greet("World")
}
`

	// Write the test file first
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		log.Printf("Failed to create test file: %v", err)
		return
	}
	defer os.Remove(testFile) // Clean up

	// Query Claude to analyze and improve the code
	options := &client.QueryOptions{
		AllowedTools:   []string{"read_file", "write_file", "analyze_code"},
		PermissionMode: client.PermissionModeAcceptEdits,
	}

	fmt.Printf("Asking Claude to improve %s...\n", testFile)
	messages, err := client.QueryMessages(ctx,
		fmt.Sprintf("Please read %s, analyze it, and add error handling to the greet function", testFile),
		options)
	if err != nil {
		log.Printf("Failed to query: %v", err)
		return
	}

	// Track tool usage
	toolsUsed := make(map[string]int)

	for msg := range messages {
		// Track tool calls
		if msg.HasToolUse() {
			for _, tool := range msg.GetToolUses() {
				toolsUsed[tool.Name]++
				fmt.Printf("🔧 Using tool: %s\n", tool.Name)
			}
		}

		// Show assistant responses
		if msg.Role == types.MessageRoleAssistant && msg.GetText() != "" {
			fmt.Printf("Claude: %s\n", msg.GetText())
		}
	}

	// Summary of tools used
	fmt.Println("\nTools used in this conversation:")
	for tool, count := range toolsUsed {
		fmt.Printf("- %s: %d times\n", tool, count)
	}

	// Read the modified file to see changes
	modified, err := os.ReadFile(testFile)
	if err == nil {
		fmt.Printf("\nModified file content:\n%s\n", string(modified))
	}
}

func customToolUsage(client *client.ClaudeCodeClient) {
	ctx := context.Background()

	// Register a custom tool (this would normally connect to your tool implementation)
	customTool := &client.ClaudeCodeToolDefinition{
		Name:        "project_stats",
		Description: "Get statistics about the current project",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"include_tests": map[string]interface{}{
					"type":        "boolean",
					"description": "Include test files in statistics",
				},
			},
		},
	}

	err := client.Tools().RegisterTool("project_stats", customTool)
	if err != nil {
		log.Printf("Failed to register custom tool: %v", err)
		return
	}

	// Use the custom tool in a conversation
	options := &client.QueryOptions{
		AllowedTools: []string{"project_stats", "read_file"},
	}

	fmt.Println("Using custom tool in conversation...")
	result, err := client.QueryMessagesSync(ctx,
		"Can you analyze the project statistics including test coverage?",
		options)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	// Show results
	for _, msg := range result.Messages {
		if msg.HasToolUse() {
			fmt.Printf("Tool calls: %v\n", msg.GetToolUses())
		}
		if text := msg.GetText(); text != "" {
			fmt.Printf("[%s]: %s\n", msg.Role, text)
		}
	}
}

// Helper function to create a sample project structure
func createSampleProject() error {
	dirs := []string{
		"sample_project/src",
		"sample_project/tests",
		"sample_project/docs",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	files := map[string]string{
		"sample_project/main.go": `package main

func main() {
    println("Sample project")
}`,
		"sample_project/src/utils.go": `package src

func Add(a, b int) int {
    return a + b
}`,
		"sample_project/tests/utils_test.go": `package tests

import "testing"

func TestAdd(t *testing.T) {
    // Test implementation
}`,
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

// Clean up sample project
func cleanupSampleProject() {
	os.RemoveAll("sample_project")
}