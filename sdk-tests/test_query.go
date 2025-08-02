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
	fmt.Println("=== Testing SDK Query Functionality ===")
	
	ctx := context.Background()
	config := &types.ClaudeCodeConfig{
		Model: "claude-3-5-sonnet-20241022",
	}
	
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()
	
	// Test 1: Simple query
	fmt.Println("\nTest 1: Simple Query...")
	request1 := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "What is 5 + 7?",
			},
		},
		MaxTokens: 100,
	}
	
	response1, err := claudeClient.Query(ctx, request1)
	if err != nil {
		log.Printf("❌ FAILED: Simple query error: %v", err)
	} else if response1 != nil && len(response1.Content) > 0 {
		fmt.Println("✅ SUCCESS: Simple query completed")
		fmt.Printf("   Response: %s\n", response1.Content[0].Text)
		// Check if response contains "12"
		if strings.Contains(response1.Content[0].Text, "12") {
			fmt.Println("   ✅ Correct answer detected")
		}
	}
	
	// Test 2: Multi-turn conversation
	fmt.Println("\nTest 2: Multi-turn Conversation...")
	request2 := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "Remember the number 42.",
			},
			{
				Role:    types.RoleAssistant,
				Content: "I'll remember the number 42 for our conversation.",
			},
			{
				Role:    types.RoleUser,
				Content: "What number did I ask you to remember?",
			},
		},
		MaxTokens: 200,
	}
	
	response2, err := claudeClient.Query(ctx, request2)
	if err != nil {
		log.Printf("❌ FAILED: Multi-turn query error: %v", err)
	} else if response2 != nil && len(response2.Content) > 0 {
		fmt.Println("✅ SUCCESS: Multi-turn query completed")
		fmt.Printf("   Response: %s\n", response2.Content[0].Text)
		if strings.Contains(response2.Content[0].Text, "42") {
			fmt.Println("   ✅ Context retention verified")
		}
	}
	
	// Test 3: Query with system prompt
	fmt.Println("\nTest 3: Query with System Prompt...")
	request3 := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "Hello",
			},
		},
		System:    "You are a helpful assistant who always responds in haiku format.",
		MaxTokens: 100,
	}
	
	response3, err := claudeClient.Query(ctx, request3)
	if err != nil {
		log.Printf("❌ FAILED: System prompt query error: %v", err)
	} else if response3 != nil && len(response3.Content) > 0 {
		fmt.Println("✅ SUCCESS: System prompt query completed")
		fmt.Printf("   Response: %s\n", response3.Content[0].Text)
		// Count lines to see if it might be a haiku (3 lines)
		lines := strings.Split(strings.TrimSpace(response3.Content[0].Text), "\n")
		if len(lines) >= 3 {
			fmt.Println("   ✅ Response appears to be in haiku format")
		}
	}
	
	// Test 4: Query with temperature
	fmt.Println("\nTest 4: Query with Custom Temperature...")
	request4 := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "Generate a creative name for a pet fish",
			},
		},
		MaxTokens:   100,
		Temperature: 0.9, // High temperature for creativity
	}
	
	response4, err := claudeClient.Query(ctx, request4)
	if err != nil {
		log.Printf("❌ FAILED: Temperature query error: %v", err)
	} else if response4 != nil && len(response4.Content) > 0 {
		fmt.Println("✅ SUCCESS: Temperature query completed")
		fmt.Printf("   Response: %s\n", response4.Content[0].Text)
	}
	
	fmt.Println("\n=== Query Tests Complete ===")
}