package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Testing Basic SDK Initialization ===")
	
	// Test 1: Create client with default config
	fmt.Println("\nTest 1: Creating client with default config...")
	ctx := context.Background()
	config := &types.ClaudeCodeConfig{
		// Use default values
	}
	
	client1, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("❌ FAILED: Error creating client with default config: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Client created with default config")
		defer client1.Close()
	}
	
	// Test 2: Create client with custom config
	fmt.Println("\nTest 2: Creating client with custom config...")
	customConfig := &types.ClaudeCodeConfig{
		WorkingDirectory: "/tmp",
		Model:            "claude-3-5-sonnet-20241022",
		SessionID:        "test-session-123",
		Debug:            true,
		MaxTokens:        2048,
		Temperature:      0.5,
	}
	
	client2, err := client.NewClaudeCodeClient(ctx, customConfig)
	if err != nil {
		log.Printf("❌ FAILED: Error creating client with custom config: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Client created with custom config")
		fmt.Printf("   Working Directory: %s\n", customConfig.WorkingDirectory)
		fmt.Printf("   Model: %s\n", customConfig.Model)
		fmt.Printf("   SessionID: %s\n", customConfig.SessionID)
		fmt.Printf("   MaxTokens: %d\n", customConfig.MaxTokens)
		fmt.Printf("   Temperature: %.1f\n", customConfig.Temperature)
		defer client2.Close()
	}
	
	// Test 3: Verify Claude Code command is found
	fmt.Println("\nTest 3: Verifying Claude Code CLI availability...")
	// This is already tested in NewClaudeCodeClient, so if client creation succeeded, CLI is available
	if client1 != nil || client2 != nil {
		fmt.Println("✅ SUCCESS: Claude Code CLI is available")
	}
	
	fmt.Println("\n=== Basic Initialization Tests Complete ===")
}