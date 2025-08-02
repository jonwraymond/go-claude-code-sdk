package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Testing SDK Query Functionality (CLI-Compatible) ===")
	
	ctx := context.Background()
	config := &types.ClaudeCodeConfig{
		Model: "claude-3-5-sonnet-20241022",
	}
	
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()
	
	// Test 1: Simple query (without unsupported fields)
	fmt.Println("\nTest 1: Simple Query...")
	request1 := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "What is 5 + 7? Please answer with just the number.",
			},
		},
	}
	
	response1, err := claudeClient.Query(ctx, request1)
	if err != nil {
		log.Printf("❌ FAILED: Simple query error: %v", err)
	} else if response1 != nil && len(response1.Content) > 0 {
		fmt.Println("✅ SUCCESS: Simple query completed")
		fmt.Printf("   Response: %s\n", strings.TrimSpace(response1.Content[0].Text))
		// Check if response contains "12"
		if strings.Contains(response1.Content[0].Text, "12") {
			fmt.Println("   ✅ Correct answer detected")
		}
	}
	
	// Test 2: Query with model specification
	fmt.Println("\nTest 2: Query with Model Specification...")
	request2 := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "Say 'Hello SDK!' and nothing else.",
			},
		},
		Model: "claude-3-5-sonnet-20241022", // Explicitly set model
	}
	
	response2, err := claudeClient.Query(ctx, request2)
	if err != nil {
		log.Printf("❌ FAILED: Model query error: %v", err)
	} else if response2 != nil && len(response2.Content) > 0 {
		fmt.Println("✅ SUCCESS: Model query completed")
		fmt.Printf("   Response: %s\n", strings.TrimSpace(response2.Content[0].Text))
		if strings.Contains(response2.Content[0].Text, "Hello SDK!") {
			fmt.Println("   ✅ Expected response received")
		}
	}
	
	// Test 3: Multi-message conversation (checking if this works)
	fmt.Println("\nTest 3: Testing Message Handling...")
	request3 := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "What is the capital of France? Answer in one word.",
			},
		},
	}
	
	response3, err := claudeClient.Query(ctx, request3)
	if err != nil {
		log.Printf("❌ FAILED: Message handling error: %v", err)
	} else if response3 != nil && len(response3.Content) > 0 {
		fmt.Println("✅ SUCCESS: Message handling completed")
		fmt.Printf("   Response: %s\n", strings.TrimSpace(response3.Content[0].Text))
		if strings.Contains(strings.ToLower(response3.Content[0].Text), "paris") {
			fmt.Println("   ✅ Correct answer detected")
		}
	}
	
	fmt.Println("\n=== Query Tests Complete ===")
}