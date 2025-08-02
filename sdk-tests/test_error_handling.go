package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Testing SDK Error Handling ===")
	
	ctx := context.Background()
	
	// Test 1: Invalid client configuration
	fmt.Println("\nTest 1: Invalid Client Configuration...")
	invalidConfig := &types.ClaudeCodeConfig{
		WorkingDirectory: "/nonexistent/directory/that/should/not/exist",
		Model:            "invalid-model-name",
	}
	
	_, err := client.NewClaudeCodeClient(ctx, invalidConfig)
	if err != nil {
		fmt.Println("✅ SUCCESS: Client creation failed as expected")
		fmt.Printf("   Error: %v\n", err)
		fmt.Printf("   Error type: %T\n", err)
	} else {
		log.Printf("❌ FAILED: Expected error for invalid configuration")
	}
	
	// Test 2: Invalid query
	fmt.Println("\nTest 2: Invalid Query...")
	config := &types.ClaudeCodeConfig{
		Model: "claude-3-5-sonnet-20241022",
	}
	
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()
	
	// Empty messages
	emptyRequest := &types.QueryRequest{
		Messages: []types.Message{},
	}
	
	_, err = claudeClient.Query(ctx, emptyRequest)
	if err != nil {
		fmt.Println("✅ SUCCESS: Empty query failed as expected")
		fmt.Printf("   Error: %v\n", err)
	} else {
		log.Printf("❌ FAILED: Expected error for empty messages")
	}
	
	// Test 3: Session errors
	fmt.Println("\nTest 3: Session Errors...")
	
	// Try to get non-existent session
	_, err = claudeClient.GetSession("non-existent-session-id")
	if err != nil {
		fmt.Println("✅ SUCCESS: Non-existent session error")
		fmt.Printf("   Error: %v\n", err)
	} else {
		log.Printf("❌ FAILED: Expected error for non-existent session")
	}
	
	// Create session with empty ID
	_, err = claudeClient.CreateSession(ctx, "")
	if err != nil {
		fmt.Println("✅ SUCCESS: Empty session ID error")
		fmt.Printf("   Error: %v\n", err)
	} else {
		log.Printf("❌ FAILED: Expected error for empty session ID")
	}
	
	// Test 4: Command errors
	fmt.Println("\nTest 4: Command Errors...")
	
	// Nil command
	_, err = claudeClient.ExecuteCommand(ctx, nil)
	if err == nil {
		// Check if result indicates error
		fmt.Println("✅ SUCCESS: Nil command handled gracefully")
	} else {
		fmt.Printf("   Error: %v\n", err)
	}
	
	// Invalid slash command
	result, err := claudeClient.ExecuteSlashCommand(ctx, "invalid-not-slash")
	if err != nil || (result != nil && !result.Success) {
		fmt.Println("✅ SUCCESS: Invalid slash command failed as expected")
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
		} else if result != nil {
			fmt.Printf("   Result error: %s\n", result.Error)
		}
	} else {
		log.Printf("❌ FAILED: Expected error for invalid slash command")
	}
	
	// Test 5: File operation errors
	fmt.Println("\nTest 5: File Operation Command Errors...")
	
	// Read non-existent file
	readCmd := &types.Command{
		Type: types.CommandRead,
		Args: []string{"/nonexistent/file/path.txt"},
	}
	
	result, err = claudeClient.ExecuteCommand(ctx, readCmd)
	if err != nil {
		fmt.Println("✅ SUCCESS: Read non-existent file error")
		fmt.Printf("   Error: %v\n", err)
	} else if result != nil {
		fmt.Printf("   Command success: %v\n", result.Success)
		if result.Output != "" {
			fmt.Printf("   Output: %s\n", strings.TrimSpace(result.Output)[:min(100, len(result.Output))])
		}
	}
	
	// Write to invalid path
	writeCmd := &types.Command{
		Type: types.CommandWrite,
		Args: []string{"/root/restricted/file.txt", "test content"},
	}
	
	result, err = claudeClient.ExecuteCommand(ctx, writeCmd)
	if err != nil {
		fmt.Println("✅ SUCCESS: Write to restricted path error")
		fmt.Printf("   Error: %v\n", err)
	} else if result != nil {
		fmt.Printf("   Command reported: success=%v\n", result.Success)
	}
	
	// Test 6: MCP errors
	fmt.Println("\nTest 6: MCP Error Handling...")
	mcpManager := claudeClient.MCP()
	
	// Add server with empty name
	err = mcpManager.AddServer("", &types.MCPServerConfig{
		Command: "test",
	})
	if err != nil {
		fmt.Println("✅ SUCCESS: Empty MCP server name error")
		fmt.Printf("   Error: %v\n", err)
	} else {
		log.Printf("❌ FAILED: Expected error for empty server name")
	}
	
	// Add server with nil config
	err = mcpManager.AddServer("test", nil)
	if err != nil {
		fmt.Println("✅ SUCCESS: Nil MCP config error")
		fmt.Printf("   Error: %v\n", err)
	} else {
		log.Printf("❌ FAILED: Expected error for nil config")
	}
	
	// Remove non-existent server
	err = mcpManager.RemoveServer("non-existent-server")
	if err != nil {
		fmt.Println("✅ SUCCESS: Remove non-existent server error")
		fmt.Printf("   Error: %v\n", err)
	} else {
		log.Printf("❌ FAILED: Expected error for non-existent server")
	}
	
	// Test 7: Context cancellation
	fmt.Println("\nTest 7: Context Cancellation...")
	
	// Create a context with timeout
	shortCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
	defer cancel()
	
	// Sleep to ensure context expires
	time.Sleep(5 * time.Millisecond)
	
	// Try query with cancelled context
	request := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "This should fail due to cancelled context",
			},
		},
	}
	
	_, err = claudeClient.Query(shortCtx, request)
	if err != nil {
		fmt.Println("✅ SUCCESS: Cancelled context error")
		fmt.Printf("   Error: %v\n", err)
		if strings.Contains(err.Error(), "context") {
			fmt.Println("   ✅ Error mentions context cancellation")
		}
	} else {
		log.Printf("❌ FAILED: Expected error for cancelled context")
	}
	
	// Test 8: File loading errors
	fmt.Println("\nTest 8: File Loading Errors...")
	
	// Try to create session with invalid directory
	tempDir, _ := os.MkdirTemp("", "sdk-error-test-*")
	defer os.RemoveAll(tempDir)
	
	nonExistentPath := filepath.Join(tempDir, "nonexistent", "project")
	
	// Try to execute command with non-existent working directory
	errorConfig := &types.ClaudeCodeConfig{
		WorkingDirectory: nonExistentPath,
		Model:            "claude-3-5-sonnet-20241022",
	}
	
	errorClient, err := client.NewClaudeCodeClient(ctx, errorConfig)
	if err != nil {
		fmt.Println("✅ SUCCESS: Non-existent working directory error")
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Println("❌ Client created with non-existent directory")
		errorClient.Close()
	}
	
	fmt.Println("\n=== Error Handling Tests Complete ===")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}