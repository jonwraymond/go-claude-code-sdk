package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Testing SDK Session Management ===")
	
	ctx := context.Background()
	config := &types.ClaudeCodeConfig{
		Model: "claude-3-5-sonnet-20241022",
	}
	
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()
	
	// Test 1: Create a new session using a generated UUID
	fmt.Println("\nTest 1: Creating New Session...")
	sessionID1 := claudeClient.GenerateSessionID()
	session1, err := claudeClient.CreateSession(ctx, sessionID1)
	if err != nil {
		log.Printf("❌ FAILED: Create session error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Session created")
		fmt.Printf("   Session ID: %s\n", session1.ID)
		defer session1.Close()
		
		// Test query within session
		request := &types.QueryRequest{
			Messages: []types.Message{
				{
					Role:    types.RoleUser,
					Content: "Remember the word 'ELEPHANT'. What word did I just ask you to remember?",
				},
			},
		}
		
		response, err := session1.Query(ctx, request)
		if err != nil {
			log.Printf("❌ FAILED: Session query error: %v", err)
		} else if response != nil && len(response.Content) > 0 {
			fmt.Println("✅ SUCCESS: Query within session completed")
			fmt.Printf("   Response: %s\n", strings.TrimSpace(response.Content[0].Text))
			if strings.Contains(strings.ToUpper(response.Content[0].Text), "ELEPHANT") {
				fmt.Println("   ✅ Session context working correctly")
			}
		}
	}
	
	// Test 2: Create multiple sessions
	fmt.Println("\nTest 2: Multiple Sessions...")
	sessionID2 := claudeClient.GenerateSessionID()
	session2, err := claudeClient.CreateSession(ctx, sessionID2)
	if err != nil {
		log.Printf("❌ FAILED: Create second session error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Second session created")
		fmt.Printf("   Session ID: %s\n", session2.ID)
		defer session2.Close()
	}
	
	// Test 3: Test UUID normalization with non-UUID input
	fmt.Println("\nTest 3: Testing UUID Normalization...")
	// This should automatically convert to a valid UUID
	session3, err := claudeClient.CreateSession(ctx, "my-custom-session-name")
	if err != nil {
		log.Printf("❌ FAILED: Create session with custom name error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Session created with custom name")
		fmt.Printf("   Original input: my-custom-session-name\n")
		fmt.Printf("   Normalized ID: %s\n", session3.ID)
		defer session3.Close()
	}
	
	// Test 4: List sessions
	fmt.Println("\nTest 4: Listing Active Sessions...")
	sessions := claudeClient.ListSessions()
	fmt.Printf("✅ Active sessions count: %d\n", len(sessions))
	for i, sessionID := range sessions {
		fmt.Printf("   %d. %s\n", i+1, sessionID)
	}
	
	// Test 5: Get existing session
	fmt.Println("\nTest 5: Retrieving Existing Session...")
	if len(sessions) > 0 {
		retrievedSession, err := claudeClient.GetSession(sessions[0])
		if err != nil {
			log.Printf("❌ FAILED: Get session error: %v", err)
		} else {
			fmt.Println("✅ SUCCESS: Session retrieved")
			fmt.Printf("   Retrieved Session ID: %s\n", retrievedSession.ID)
		}
	}
	
	// Test 6: Session metadata
	fmt.Println("\nTest 6: Session Metadata...")
	if session1 != nil {
		session1.SetMetadata("test_key", "test_value")
		session1.SetMetadata("timestamp", time.Now().Format(time.RFC3339))
		
		metadata := session1.GetMetadata()
		fmt.Println("✅ SUCCESS: Metadata operations completed")
		fmt.Printf("   Metadata entries: %d\n", len(metadata))
		for key, value := range metadata {
			fmt.Printf("   - %s: %v\n", key, value)
		}
	}
	
	// Test 7: Close specific session
	fmt.Println("\nTest 7: Closing Sessions...")
	if len(sessions) > 1 {
		err = claudeClient.Sessions().CloseSession(sessions[1])
		if err != nil {
			log.Printf("❌ FAILED: Close session error: %v", err)
		} else {
			fmt.Println("✅ SUCCESS: Session closed")
			
			// Verify session was closed
			remainingSessions := claudeClient.ListSessions()
			fmt.Printf("   Remaining sessions: %d\n", len(remainingSessions))
		}
	}
	
	fmt.Println("\n=== Session Management Tests Complete ===")
}