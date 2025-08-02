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
	fmt.Println("=== Testing Fixed CLI Flags ===")
	
	ctx := context.Background()
	
	// Test 1: Query with system prompt (should use --append-system-prompt)
	fmt.Println("\nTest 1: Query with System Prompt...")
	config := &types.ClaudeCodeConfig{
		Model: "claude-3-5-sonnet-20241022",
	}
	
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()
	
	// Test with system prompt
	request := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "What is 2 + 2?",
			},
		},
		System: "You are a helpful math tutor. Always explain your reasoning.",
	}
	
	response, err := claudeClient.Query(ctx, request)
	if err != nil {
		log.Printf("❌ FAILED: Query with system prompt error: %v", err)
	} else if response != nil && len(response.Content) > 0 {
		fmt.Println("✅ SUCCESS: Query with system prompt completed")
		output := strings.TrimSpace(response.Content[0].Text)
		fmt.Printf("   Response preview: %s...\n", output[:min(100, len(output))])
		if strings.Contains(output, "4") {
			fmt.Println("   ✅ Correct answer included")
		}
	}
	
	// Test 2: Query with model specification
	fmt.Println("\nTest 2: Query with Model Specification...")
	request2 := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "Say 'Hello from Claude' and nothing else.",
			},
		},
		Model: "claude-3-5-sonnet-20241022",
	}
	
	response2, err := claudeClient.Query(ctx, request2)
	if err != nil {
		log.Printf("❌ FAILED: Query with model error: %v", err)
	} else if response2 != nil && len(response2.Content) > 0 {
		fmt.Println("✅ SUCCESS: Query with model specification completed")
		output := strings.TrimSpace(response2.Content[0].Text)
		fmt.Printf("   Response: %s\n", output)
	}
	
	// Test 3: Query with session ID
	fmt.Println("\nTest 3: Session-based Query...")
	
	// Create a session
	session, err := claudeClient.CreateSession(ctx, "test-flags-session")
	if err != nil {
		log.Printf("❌ FAILED: Create session error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Session created")
		defer session.Close()
		
		// Query within session
		sessionRequest := &types.QueryRequest{
			Messages: []types.Message{
				{
					Role:    types.RoleUser,
					Content: "Remember the number 42. What number did I just ask you to remember?",
				},
			},
		}
		
		sessionResponse, err := session.Query(ctx, sessionRequest)
		if err != nil {
			log.Printf("❌ FAILED: Session query error: %v", err)
		} else if sessionResponse != nil && len(sessionResponse.Content) > 0 {
			fmt.Println("✅ SUCCESS: Session query completed")
			output := strings.TrimSpace(sessionResponse.Content[0].Text)
			fmt.Printf("   Response: %s\n", output)
			if strings.Contains(output, "42") {
				fmt.Println("   ✅ Session context working")
			}
		}
	}
	
	// Test 4: Query with allowed tools (should use --allowedTools)
	fmt.Println("\nTest 4: Query with Tool Restrictions...")
	
	// Create query options with allowed tools
	queryOptions := &client.QueryOptions{
		AllowedTools:   []string{"Read", "Search"},
		PermissionMode: client.PermissionModeAcceptEdits,
	}
	
	// Use QueryMessagesSync for this test
	result, err := claudeClient.QueryMessagesSync(ctx, 
		"List the files in the current directory", 
		queryOptions)
	
	if err != nil {
		log.Printf("❌ FAILED: Query with tools error: %v", err)
	} else if result != nil && len(result.Messages) > 0 {
		fmt.Println("✅ SUCCESS: Query with tool restrictions completed")
		// Check if any messages were returned
		fmt.Printf("   Messages returned: %d\n", len(result.Messages))
		for i, msg := range result.Messages {
			if i < 3 { // Show first 3 messages
				preview := fmt.Sprintf("%v", msg.Content)
				if len(preview) > 50 {
					preview = preview[:50] + "..."
				}
				fmt.Printf("   Message %d (%s): %s\n", i+1, msg.Role, preview)
			}
		}
	}
	
	fmt.Println("\n=== Fixed CLI Flags Tests Complete ===")
	
	// Summary
	fmt.Println("\nFlag Mapping Summary:")
	fmt.Println("  ✓ System prompt: --append-system-prompt (not --system)")
	fmt.Println("  ✓ Allowed tools: --allowedTools (not --tools)")
	fmt.Println("  ✓ Permission mode: --permission-mode with values")
	fmt.Println("  ✓ Session ID: --session-id (correct)")
	fmt.Println("  ✓ Model: --model (correct)")
	fmt.Println("  ✗ Max tokens: Not supported by CLI")
	fmt.Println("  ✗ Temperature: Not supported by CLI")
	fmt.Println("  ✗ Stream: Not a direct flag")
	fmt.Println("  ✗ Timeout: Not supported by CLI")
	fmt.Println("  ✗ Max turns: Not supported by CLI")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}