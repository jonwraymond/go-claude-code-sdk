// Basic example demonstrating simple Claude Code SDK usage
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	// Create configuration
	// API key can be set via ANTHROPIC_API_KEY environment variable
	config := types.NewClaudeCodeConfig()
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}

	// Create context
	ctx := context.Background()

	// Create Claude Code client
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()

	// Example 1: Simple synchronous query
	fmt.Println("=== Example 1: Simple Query ===")
	simpleQuery(ctx, claudeClient)

	// Example 2: Streaming messages
	fmt.Println("\n=== Example 2: Streaming Messages ===")
	streamingExample(ctx, claudeClient)

	// Example 3: Query with options
	fmt.Println("\n=== Example 3: Query with Options ===")
	queryWithOptions(ctx, claudeClient)
}

func simpleQuery(ctx context.Context, claudeClient *client.ClaudeCodeClient) {

	// Execute a simple query and get all messages
	result, err := claudeClient.QueryMessagesSync(ctx, "What is Go programming language?", nil)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	// Print all messages
	for _, msg := range result.Messages {
		fmt.Printf("[%s]: %s\n", msg.Role, msg.GetText())
	}
}

func streamingExample(ctx context.Context, claudeClient *client.ClaudeCodeClient) {

	// Stream messages in real-time
	messages, err := claudeClient.QueryMessages(ctx, "Write a Hello World program in Go", nil)
	if err != nil {
		log.Printf("Failed to start streaming: %v", err)
		return
	}

	// Process messages as they arrive
	for msg := range messages {
		switch msg.Role {
		case types.MessageRoleUser:
			fmt.Printf("You: %s\n", msg.GetText())
		case types.MessageRoleAssistant:
			fmt.Printf("Claude: %s\n", msg.GetText())
		case types.MessageRoleTool:
			fmt.Printf("Tool: %s\n", msg.GetText())
		}
	}
}

func queryWithOptions(ctx context.Context, claudeClient *client.ClaudeCodeClient) {

	// Configure query options
	options := &client.QueryOptions{
		SystemPrompt: "You are a helpful Go programming expert. Always provide clear, concise answers with code examples.",
		MaxTurns:     5,
		Model:        "claude-3-opus",
		// Allow file operations
		AllowedTools: []string{"read_file", "write_file"},
		// Automatically accept file edits
		PermissionMode: client.PermissionModeAcceptEdits,
	}

	// Execute query with options
	messages, err := claudeClient.QueryMessages(ctx, "Create a simple HTTP server in Go", options)
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		return
	}

	// Process messages
	for msg := range messages {
		// Check if message contains tool usage
		if msg.HasToolUse() {
			tools := msg.GetToolUses()
			for _, tool := range tools {
				fmt.Printf("Using tool: %s\n", tool.Function.Name)
			}
		}

		// Print message content
		if text := msg.GetText(); text != "" {
			fmt.Printf("[%s]: %s\n", msg.Role, text)
		}
	}
}