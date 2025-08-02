package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	ctx := context.Background()

	fmt.Println("Claude Code SDK - Authentication Methods Demo")
	fmt.Println(strings.Repeat("=", 50))

	// Demonstrate different ways to configure authentication
	demonstrateAuthMethods(ctx)
}

func demonstrateAuthMethods(ctx context.Context) {
	fmt.Println("\n1. Testing Subscription Authentication")
	fmt.Println(strings.Repeat("-", 40))
	testSubscriptionAuth(ctx)

	fmt.Println("\n2. Testing API Key Authentication")
	fmt.Println(strings.Repeat("-", 40))
	testAPIKeyAuth(ctx)

	fmt.Println("\n3. Testing Automatic Detection")
	fmt.Println(strings.Repeat("-", 40))
	testAutoDetection(ctx)

	fmt.Println("\n4. Authentication Status Check")
	fmt.Println(strings.Repeat("-", 40))
	checkAuthStatus()
}

func testSubscriptionAuth(ctx context.Context) {
	// Test subscription authentication
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: ".",
		// Let the client generate a proper UUID
		AuthMethod:      types.AuthTypeSubscription,
	}

	// Check if subscription auth is available before creating client
	subscriptionAuth := &types.SubscriptionAuth{}
	if !subscriptionAuth.IsValid(ctx) {
		fmt.Println("❌ Subscription authentication not available")
		fmt.Println("   Run 'claude setup-token' to configure subscription authentication")
		return
	}

	client, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		fmt.Printf("❌ Failed to create subscription client: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Println("✅ Subscription authentication configured successfully")
	
	// Test a simple query
	response, err := client.Query(ctx, &types.QueryRequest{
		Messages: []types.Message{
			{Role: types.RoleUser, Content: "Hello from subscription auth!"},
		},
	})

	if err != nil {
		fmt.Printf("❌ Subscription query failed: %v\n", err)
	} else {
		fmt.Printf("✅ Subscription query successful: %s\n", 
			extractFirstTextContent(response.Content))
	}
}

func testAPIKeyAuth(ctx context.Context) {
	// Check for API key in environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ API key not found in ANTHROPIC_API_KEY environment variable")
		fmt.Println("   Set ANTHROPIC_API_KEY to test API key authentication")
		return
	}

	config := &types.ClaudeCodeConfig{
		WorkingDirectory: ".",
		SessionID:        "api-key-test",
		APIKey:          apiKey,
		AuthMethod:      types.AuthTypeAPIKey,
	}

	client, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		fmt.Printf("❌ Failed to create API key client: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Println("✅ API key authentication configured successfully")
	
	// Test a simple query
	response, err := client.Query(ctx, &types.QueryRequest{
		Messages: []types.Message{
			{Role: types.RoleUser, Content: "Hello from API key auth!"},
		},
	})

	if err != nil {
		fmt.Printf("❌ API key query failed: %v\n", err)
	} else {
		fmt.Printf("✅ API key query successful: %s\n", 
			extractFirstTextContent(response.Content))
	}
}

func testAutoDetection(ctx context.Context) {
	// Test automatic authentication detection
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: ".",
		SessionID:        "auto-detect-test",
		// No explicit AuthMethod - let it auto-detect
	}

	// Apply defaults to trigger auto-detection
	config.ApplyDefaults()

	fmt.Printf("🔍 Auto-detected authentication method: %s\n", config.AuthMethod)

	client, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		fmt.Printf("❌ Failed to create auto-detected client: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Println("✅ Auto-detection successful")

	// Show which method was selected
	if config.IsUsingSubscriptionAuth() {
		fmt.Println("   Using subscription authentication")
	} else if config.IsUsingAPIKeyAuth() {
		fmt.Println("   Using API key authentication")
	}
}

func checkAuthStatus() {
	fmt.Println("Authentication Status:")
	
	// Check subscription authentication availability
	subscriptionAuth := &types.SubscriptionAuth{}
	subscriptionAvailable := subscriptionAuth.IsValid(context.Background())
	fmt.Printf("  Subscription auth available: %v\n", subscriptionAvailable)
	
	// Check API key availability
	apiKeyAvailable := os.Getenv("ANTHROPIC_API_KEY") != ""
	fmt.Printf("  API key configured: %v\n", apiKeyAvailable)
	
	// Show recommendations
	fmt.Println("\nRecommendations:")
	if subscriptionAvailable {
		fmt.Println("  ✅ Use subscription authentication for the best experience")
	} else {
		fmt.Println("  💡 Run 'claude setup-token' to enable subscription authentication")
	}
	
	if apiKeyAvailable {
		fmt.Println("  ✅ API key authentication is available as fallback")
	} else {
		fmt.Println("  💡 Set ANTHROPIC_API_KEY environment variable for API key authentication")
	}

	if !subscriptionAvailable && !apiKeyAvailable {
		fmt.Println("  ⚠️  No authentication method available")
		fmt.Println("     Please configure either subscription or API key authentication")
	}
}

func extractFirstTextContent(content []types.ContentBlock) string {
	for _, block := range content {
		if block.Type == "text" && len(block.Text) > 0 {
			// Return first 100 characters for demo
			if len(block.Text) > 100 {
				return block.Text[:100] + "..."
			}
			return block.Text
		}
	}
	return "(no text content)"
}