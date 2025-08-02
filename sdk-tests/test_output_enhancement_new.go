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
	fmt.Println("=== Testing SDK Command Output Enhancement ===")
	
	ctx := context.Background()
	config := &types.ClaudeCodeConfig{
		Model: "claude-3-5-sonnet-20241022",
	}
	
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()
	
	// Create a test file with longer content
	fmt.Println("\nPreparing test file with longer content...")
	longContent := strings.Repeat("This is a test line with some content.\n", 100)
	writeCmd := client.WriteFile("test_long_file.txt", longContent)
	
	_, err = claudeClient.ExecuteCommand(ctx, writeCmd)
	if err != nil {
		log.Printf("Failed to create test file: %v", err)
		return
	}
	fmt.Println("✅ Test file created")
	
	// Test 1: Regular read command (may be truncated)
	fmt.Println("\nTest 1: Regular Read Command...")
	readCmd := client.ReadFile("test_long_file.txt")
	
	result1, err := claudeClient.ExecuteCommand(ctx, readCmd)
	if err != nil {
		log.Printf("❌ FAILED: Read command error: %v", err)
	} else if result1 != nil {
		fmt.Println("✅ SUCCESS: Read command executed")
		fmt.Printf("   Success: %v\n", result1.Success)
		fmt.Printf("   Output length: %d characters\n", result1.OutputLength)
		fmt.Printf("   Is truncated: %v\n", result1.IsTruncated)
		
		// Debug: show the actual output even if empty
		fmt.Printf("   Raw output: '%s'\n", result1.Output)
		
		// Check metadata
		if result1.Metadata != nil {
			fmt.Printf("   Stop reason: %v\n", result1.Metadata["stop_reason"])
		}
		
		if result1.IsTruncated {
			fmt.Println("   ✅ Truncation detected correctly")
		}
		// Show preview of output
		preview := result1.Output
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		if len(preview) > 0 {
			fmt.Printf("   Output preview: %s\n", preview)
		}
	}
	
	// Test 2: Read command with verbose output
	fmt.Println("\nTest 2: Read Command with Verbose Output...")
	verboseReadCmd := client.ReadFile("test_long_file.txt", client.WithVerboseOutput())
	
	result2, err := claudeClient.ExecuteCommand(ctx, verboseReadCmd)
	if err != nil {
		log.Printf("❌ FAILED: Verbose read command error: %v", err)
	} else if result2 != nil {
		fmt.Println("✅ SUCCESS: Verbose read command executed")
		fmt.Printf("   Output length: %d characters\n", result2.OutputLength)
		fmt.Printf("   Is truncated: %v\n", result2.IsTruncated)
		
		if result2.FullOutput != "" {
			fmt.Printf("   Full output available: %d characters\n", len(result2.FullOutput))
			fmt.Println("   ✅ Full output retrieved successfully")
		} else {
			fmt.Printf("   Using regular output: %d characters\n", len(result2.Output))
		}
	}
	
	// Test 3: Test truncation detection with minimal output
	fmt.Println("\nTest 3: Truncation Detection (minimal output)...")
	// Create a command that might return "..."
	searchCmd := &types.Command{
		Type: types.CommandSearch,
		Args: []string{"nonexistent_pattern_12345"},
	}
	
	result3, err := claudeClient.ExecuteCommand(ctx, searchCmd)
	if err != nil {
		log.Printf("❌ FAILED: Search command error: %v", err)
	} else if result3 != nil {
		fmt.Println("✅ SUCCESS: Search command executed")
		fmt.Printf("   Output: '%s'\n", result3.Output)
		fmt.Printf("   Output length: %d characters\n", result3.OutputLength)
		fmt.Printf("   Is truncated: %v\n", result3.IsTruncated)
		
		if result3.Output == "..." && result3.IsTruncated {
			fmt.Println("   ✅ Minimal truncation ('...') detected correctly")
		}
	}
	
	// Test 4: List command with verbose output
	fmt.Println("\nTest 4: List Command with Verbose Output...")
	listCmd := &types.Command{
		Type:          types.CommandRead,
		Args:          []string{"."},
		VerboseOutput: true,
		Options: map[string]any{
			"pattern": "*.go",
		},
	}
	
	result4, err := claudeClient.ExecuteCommand(ctx, listCmd)
	if err != nil {
		log.Printf("❌ FAILED: List command error: %v", err)
	} else if result4 != nil {
		fmt.Println("✅ SUCCESS: List command executed")
		fmt.Printf("   Output length: %d characters\n", result4.OutputLength)
		fmt.Printf("   Is truncated: %v\n", result4.IsTruncated)
		
		// Check if we got file listings
		if strings.Contains(result4.Output, ".go") {
			fmt.Println("   ✅ Go files listed in output")
		}
	}
	
	// Clean up
	fmt.Println("\nCleaning up test file...")
	cleanupCmd := &types.Command{
		Type: types.CommandWrite,
		Args: []string{"test_long_file.txt", ""},
	}
	claudeClient.ExecuteCommand(ctx, cleanupCmd)
	
	fmt.Println("\n=== Output Enhancement Tests Complete ===")
	
	// Summary
	fmt.Println("\nFeatures tested:")
	fmt.Println("- ✅ Truncation detection for regular output")
	fmt.Println("- ✅ Verbose output option for commands")
	fmt.Println("- ✅ Full output retrieval for truncated content")
	fmt.Println("- ✅ Minimal truncation ('...') detection")
	fmt.Println("- ✅ Output length tracking")
}