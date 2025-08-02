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
	// Example 1: Automatic Authentication Detection (Recommended)
	fmt.Println("=== Example 1: Automatic Authentication Detection ===")
	autoConfig := types.NewClaudeCodeConfig()
	// The SDK will automatically detect and use the best available authentication method
	// Priority: 1. Claude subscription (if logged in), 2. API key (if set)
	
	ctx := context.Background()
	autoClient, err := client.NewClaudeCodeClient(ctx, autoConfig)
	if err != nil {
		log.Printf("Auto-detection failed: %v", err)
	} else {
		fmt.Printf("Successfully connected using: %s\n", autoConfig.AuthMethod)
		autoClient.Close()
	}

	fmt.Println("\n=== Example 2: Explicit Subscription Authentication ===")
	// Force subscription authentication (requires claude login)
	subConfig := &types.ClaudeCodeConfig{
		AuthMethod: types.AuthTypeSubscription,
		Debug:      true, // Show CLI commands for debugging
	}
	
	subClient, err := client.NewClaudeCodeClient(ctx, subConfig)
	if err != nil {
		log.Printf("Subscription auth failed: %v", err)
		fmt.Println("Make sure you're logged in with: claude login")
	} else {
		fmt.Println("Successfully connected using Claude subscription!")
		
		// Use the client
		messages, err := subClient.QueryMessages(ctx, "Hello! What authentication method am I using?", nil)
		if err != nil {
			log.Printf("Query failed: %v", err)
		} else {
			for msg := range messages {
				fmt.Printf("Claude: %s\n", msg.GetText())
			}
		}
		subClient.Close()
	}

	fmt.Println("\n=== Example 3: Fallback to API Key ===")
	// Explicitly use API key authentication
	apiConfig := &types.ClaudeCodeConfig{
		APIKey:     os.Getenv("ANTHROPIC_API_KEY"),
		AuthMethod: types.AuthTypeAPIKey,
	}
	
	if apiConfig.APIKey != "" {
		apiClient, err := client.NewClaudeCodeClient(ctx, apiConfig)
		if err != nil {
			log.Printf("API key auth failed: %v", err)
		} else {
			fmt.Println("Successfully connected using API key!")
			apiClient.Close()
		}
	} else {
		fmt.Println("No API key found in ANTHROPIC_API_KEY environment variable")
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println("The Go Claude Code SDK now supports both authentication methods:")
	fmt.Println("1. Claude subscription login (claude login)")
	fmt.Println("2. Anthropic API key (ANTHROPIC_API_KEY)")
	fmt.Println("\nThe SDK can automatically detect which method to use!")
}