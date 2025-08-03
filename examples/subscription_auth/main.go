package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	ctx := context.Background()

	// Example 1: Explicit subscription authentication
	fmt.Println("=== Example 1: Explicit Subscription Authentication ===")
	config1 := &types.ClaudeCodeConfig{
		WorkingDirectory: ".",
		SessionID:        "subscription-example",
		Model:            "claude-3-5-sonnet-20241022",
		AuthMethod:       types.AuthTypeSubscription,
	}

	client1, err := client.NewClaudeCodeClient(ctx, config1)
	if err != nil {
		log.Printf("Failed to create client with explicit subscription auth: %v", err)
		log.Println("Make sure you have run 'claude setup-token' to configure subscription authentication")
	} else {
		defer client1.Close()

		response1, err := client1.Query(ctx, &types.QueryRequest{
			Messages: []types.Message{
				{Role: types.RoleUser, Content: "Hello! Are you using subscription authentication?"},
			},
		})

		if err != nil {
			log.Printf("Query failed: %v", err)
		} else {
			fmt.Printf("Response: %s\n", extractTextContent(response1.Content))
		}
	}

	// Example 2: Automatic authentication method detection
	fmt.Println("\n=== Example 2: Automatic Authentication Detection ===")
	config2 := &types.ClaudeCodeConfig{
		WorkingDirectory: ".",
		SessionID:        "auto-detect-example",
		Model:            "claude-3-5-sonnet-20241022",
		// AuthMethod not specified - will be auto-detected
	}

	client2, err := client.NewClaudeCodeClient(ctx, config2)
	if err != nil {
		log.Printf("Failed to create client with auto-detection: %v", err)
	} else {
		defer client2.Close()

		// Show which authentication method was detected
		if config2.IsUsingSubscriptionAuth() {
			fmt.Println("Detected: Subscription authentication")
		} else if config2.IsUsingAPIKeyAuth() {
			fmt.Println("Detected: API key authentication")
		}

		response2, err := client2.Query(ctx, &types.QueryRequest{
			Messages: []types.Message{
				{Role: types.RoleUser, Content: "What authentication method are you using?"},
			},
		})

		if err != nil {
			log.Printf("Query failed: %v", err)
		} else {
			fmt.Printf("Response: %s\n", extractTextContent(response2.Content))
		}
	}

	// Example 3: Fallback to API key if subscription is not available
	fmt.Println("\n=== Example 3: API Key Fallback ===")
	config3 := &types.ClaudeCodeConfig{
		WorkingDirectory: ".",
		SessionID:        "fallback-example",
		Model:            "claude-3-5-sonnet-20241022",
		APIKey:           "test-api-key-not-real-your-api-key-here", // Replace with your API key
		AuthMethod:       types.AuthTypeAPIKey,
	}

	client3, err := client.NewClaudeCodeClient(ctx, config3)
	if err != nil {
		log.Printf("Failed to create client with API key: %v", err)
	} else {
		defer client3.Close()

		fmt.Println("Using API key authentication as fallback")
		// Note: This would work if you provide a valid API key
	}

	// Example 4: Check authentication availability
	fmt.Println("\n=== Example 4: Authentication Method Checks ===")

	// Create a minimal config to test availability
	testConfig := types.NewClaudeCodeConfig()
	testConfig.ApplyDefaults()

	fmt.Printf("Subscription auth available: %v\n", testConfig.IsUsingSubscriptionAuth())
	fmt.Printf("API key auth configured: %v\n", testConfig.IsUsingAPIKeyAuth())
	fmt.Printf("Detected auth method: %s\n", testConfig.AuthMethod)
}

// extractTextContent extracts text from content blocks
func extractTextContent(content []types.ContentBlock) string {
	for _, block := range content {
		if block.Type == "text" {
			return block.Text
		}
	}
	return ""
}
