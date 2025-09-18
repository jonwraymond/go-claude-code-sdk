// Package main demonstrates synchronous queries with Claude Code.
// This example shows how to use the non-streaming API for simple request-response patterns.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Claude Code Synchronous Queries Examples ===")

	// Example 1: Basic synchronous query
	basicSyncQueryExample()

	// Example 2: Multi-turn conversation
	multiTurnConversationExample()

	// Example 3: Using system prompts
	systemPromptExample()

	// Example 4: QueryMessagesSync API
	queryMessagesSyncExample()

	// Example 5: Error handling and retries
	errorHandlingExample()
}

// basicSyncQueryExample demonstrates a simple synchronous query
func basicSyncQueryExample() {
	fmt.Println("--- Example 1: Basic Synchronous Query ---")

	ctx := context.Background()
	config := types.NewClaudeCodeConfig()

	// Use API key if available
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
		config.AuthMethod = types.AuthTypeAPIKey
	} else {
		config.AuthMethod = types.AuthTypeSubscription
	}

	// Create client
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer claudeClient.Close()

	// Create synchronous request
	request := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "What are the key features of the Go programming language?",
			},
		},
		Model: "claude-3-5-sonnet-20241022",
	}

	fmt.Printf("Sending synchronous query...\n")
	start := time.Now()

	// Execute synchronous query
	response, err := claudeClient.Query(ctx, request)
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		return
	}

	elapsed := time.Since(start)

	fmt.Printf("✓ Query completed in %v\n", elapsed)
	fmt.Printf("Stop reason: %s\n", response.StopReason)

	// Print response content
	if len(response.Content) > 0 {
		content := extractTextContent(response.Content)
		fmt.Printf("Response:\n%s\n", content)
	}

	fmt.Println()
}

// multiTurnConversationExample demonstrates a multi-turn conversation
func multiTurnConversationExample() {
	fmt.Println("--- Example 2: Multi-turn Conversation ---")

	ctx := context.Background()
	config := types.NewClaudeCodeConfig()

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
		config.AuthMethod = types.AuthTypeAPIKey
	} else {
		config.AuthMethod = types.AuthTypeSubscription
	}

	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer claudeClient.Close()

	// Build conversation history
	var messages []types.Message

	// Turn 1: Initial question
	messages = append(messages, types.Message{
		Role:    types.RoleUser,
		Content: "Explain goroutines in Go.",
	})

	fmt.Printf("Turn 1 - User: %s\n", messages[0].Content)

	response1, err := claudeClient.Query(ctx, &types.QueryRequest{
		Messages: messages,
	})
	if err != nil {
		log.Printf("Failed to execute first query: %v", err)
		return
	}

	// Add assistant response to conversation
	assistantContent := extractTextContent(response1.Content)
	messages = append(messages, types.Message{
		Role:    types.RoleAssistant,
		Content: assistantContent,
	})

	fmt.Printf("Turn 1 - Claude: %s\n\n", truncateText(assistantContent, 200))

	// Turn 2: Follow-up question
	messages = append(messages, types.Message{
		Role:    types.RoleUser,
		Content: "Can you provide a simple code example?",
	})

	fmt.Printf("Turn 2 - User: %s\n", messages[2].Content)

	response2, err := claudeClient.Query(ctx, &types.QueryRequest{
		Messages: messages,
	})
	if err != nil {
		log.Printf("Failed to execute second query: %v", err)
		return
	}

	assistantContent2 := extractTextContent(response2.Content)
	fmt.Printf("Turn 2 - Claude: %s\n\n", truncateText(assistantContent2, 300))

	fmt.Printf("✓ Multi-turn conversation completed\n")
	fmt.Printf("  Total messages in conversation: %d\n", len(messages)+1)
	fmt.Println()
}

// systemPromptExample demonstrates using system prompts
func systemPromptExample() {
	fmt.Println("--- Example 3: System Prompt Usage ---")

	ctx := context.Background()
	config := types.NewClaudeCodeConfig()

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
		config.AuthMethod = types.AuthTypeAPIKey
	} else {
		config.AuthMethod = types.AuthTypeSubscription
	}

	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer claudeClient.Close()

	// Request with system prompt
	request := &types.QueryRequest{
		System: "You are a Go programming expert. Provide concise, practical answers with code examples when appropriate. Always explain concepts clearly for developers who are learning Go.",
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "How do channels work in Go?",
			},
		},
	}

	fmt.Printf("Sending query with system prompt...\n")
	fmt.Printf("System: %s\n", truncateText(request.System, 100))
	fmt.Printf("User: %s\n", request.Messages[0].Content)

	response, err := claudeClient.Query(ctx, request)
	if err != nil {
		log.Printf("Failed to execute query with system prompt: %v", err)
		return
	}

	content := extractTextContent(response.Content)
	fmt.Printf("Claude: %s\n\n", truncateText(content, 400))

	fmt.Printf("✓ System prompt query completed\n")
	fmt.Println()
}

// queryMessagesSyncExample demonstrates the QueryMessagesSync API
func queryMessagesSyncExample() {
	fmt.Println("--- Example 4: QueryMessagesSync API ---")

	ctx := context.Background()
	config := types.NewClaudeCodeConfig()

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
		config.AuthMethod = types.AuthTypeAPIKey
	} else {
		config.AuthMethod = types.AuthTypeSubscription
	}

	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer claudeClient.Close()

	// Configure query options
	options := &client.QueryOptions{
		Model:          "claude-3-5-sonnet-20241022",
		MaxTurns:       3,
		Stream:         false, // Synchronous mode
		PermissionMode: client.PermissionModeAcceptEdits,
		SystemPrompt:   "You are a helpful Go programming assistant.",
	}

	prompt := "Write a simple Go function that reverses a string"
	fmt.Printf("Executing QueryMessagesSync with prompt: %s\n", prompt)

	// Execute synchronous query
	result, err := claudeClient.QueryMessagesSync(ctx, prompt, options)
	if err != nil {
		log.Printf("Failed to execute QueryMessagesSync: %v", err)
		return
	}

	if result.Error != nil {
		log.Printf("Query completed with error: %v", result.Error)
		return
	}

	fmt.Printf("✓ QueryMessagesSync completed successfully\n")
	fmt.Printf("Messages in conversation:\n")

	for i, message := range result.Messages {
		role := string(message.Role)
		content := formatContent(message.Content)

		fmt.Printf("  %d. %s: %s\n", i+1, role, truncateText(content, 150))

		// Show tool calls if any
		if len(message.ToolCalls) > 0 {
			for _, toolCall := range message.ToolCalls {
				fmt.Printf("     Tool: %s(%s)\n", toolCall.Function.Name, toolCall.Function.Arguments)
			}
		}
	}

	// Show metadata
	fmt.Printf("Metadata:\n")
	for key, value := range result.Metadata {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println()
}

// errorHandlingExample demonstrates error handling and retry patterns
func errorHandlingExample() {
	fmt.Println("--- Example 5: Error Handling and Retries ---")

	ctx := context.Background()
	config := types.NewClaudeCodeConfig()
	config.TestMode = true // Use test mode to simulate errors safely

	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer claudeClient.Close()

	// Function to attempt a query with retries
	executeWithRetries := func(request *types.QueryRequest, maxRetries int) (*types.QueryResponse, error) {
		var lastErr error

		for attempt := 1; attempt <= maxRetries; attempt++ {
			fmt.Printf("  Attempt %d/%d...\n", attempt, maxRetries)

			// Create context with timeout for each attempt
			queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

			response, err := claudeClient.Query(queryCtx, request)
			cancel()

			if err == nil {
				fmt.Printf("  ✓ Success on attempt %d\n", attempt)
				return response, nil
			}

			lastErr = err
			fmt.Printf("  ✗ Failed on attempt %d: %v\n", attempt, err)

			// Wait before retrying (exponential backoff)
			if attempt < maxRetries {
				waitTime := time.Duration(attempt*attempt) * time.Second
				fmt.Printf("  Waiting %v before retry...\n", waitTime)
				time.Sleep(waitTime)
			}
		}

		return nil, fmt.Errorf("all %d attempts failed, last error: %w", maxRetries, lastErr)
	}

	// Test with a simple request
	request := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "Hello, Claude!",
			},
		},
	}

	fmt.Printf("Testing query with retry logic...\n")

	response, err := executeWithRetries(request, 3)
	if err != nil {
		fmt.Printf("✗ Query failed after retries: %v\n", err)
	} else {
		fmt.Printf("✓ Query succeeded\n")
		if len(response.Content) > 0 {
			content := extractTextContent(response.Content)
			fmt.Printf("Response: %s\n", truncateText(content, 100))
		}
	}

	// Demonstrate timeout handling
	fmt.Printf("\nTesting timeout handling...\n")

	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond) // Very short timeout
	defer cancel()

	_, err = claudeClient.Query(timeoutCtx, request)
	if err != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			fmt.Printf("✓ Timeout handled correctly: %v\n", err)
		} else {
			fmt.Printf("✗ Unexpected error: %v\n", err)
		}
	} else {
		fmt.Printf("? Query completed faster than expected\n")
	}

	fmt.Println()
}

// Helper functions

// extractTextContent extracts text from content blocks
func extractTextContent(content []types.ContentBlock) string {
	var result strings.Builder
	for _, block := range content {
		if block.Type == "text" {
			result.WriteString(block.Text)
		}
	}
	return result.String()
}

// formatContent formats message content for display
func formatContent(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []types.ContentBlock:
		return extractTextContent(v)
	default:
		return fmt.Sprintf("%v", content)
	}
}

// truncateText truncates text to a maximum length
func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}
