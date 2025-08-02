//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Testing SDK Command Execution ===")

	ctx := context.Background()

	// Create a test directory with some files
	testDir, err := os.MkdirTemp("", "sdk-command-test-*")
	if err != nil {
		log.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create test files
	testFile1 := filepath.Join(testDir, "test1.txt")
	testFile2 := filepath.Join(testDir, "test2.go")

	err = os.WriteFile(testFile1, []byte("Hello from test file 1!\nThis is a test."), 0644)
	if err != nil {
		log.Fatalf("Failed to create test file 1: %v", err)
	}

	err = os.WriteFile(testFile2, []byte("package main\n\nfunc main() {\n\tprintln(\"Hello, SDK!\")\n}"), 0644)
	if err != nil {
		log.Fatalf("Failed to create test file 2: %v", err)
	}

	// Create client with test directory as working directory
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: testDir,
		Model:            "claude-3-5-sonnet-20241022",
	}

	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()

	// Test 1: Read file command
	fmt.Println("\nTest 1: Read File Command...")
	readCmd := &types.Command{
		Type: types.CommandRead,
		Args: []string{"test1.txt"},
	}

	result1, err := claudeClient.ExecuteCommand(ctx, readCmd)
	if err != nil {
		log.Printf("❌ FAILED: Read command error: %v", err)
	} else if result1 != nil {
		fmt.Println("✅ SUCCESS: Read command executed")
		fmt.Printf("   Output preview: %s...\n", strings.TrimSpace(result1.Output)[:min(50, len(result1.Output))])
		if result1.Success {
			fmt.Println("   ✅ Command reported success")
		}
	}

	// Test 2: Use slash command to list files
	fmt.Println("\nTest 2: List Files (via slash command)...")
	result2, err := claudeClient.ExecuteSlashCommand(ctx, "/ls")
	if err != nil {
		// Try alternative command if /ls doesn't exist
		result2, err = claudeClient.ExecuteSlashCommand(ctx, "/list")
		if err != nil {
			log.Printf("❌ FAILED: List slash command error: %v", err)
		}
	}

	if err == nil && result2 != nil {
		fmt.Println("✅ SUCCESS: List command executed")
		output := strings.TrimSpace(result2.Output)
		fmt.Printf("   Output preview: %s...\n", output[:min(200, len(output))])
		if strings.Contains(output, "test1.txt") || strings.Contains(output, "test2.go") {
			fmt.Println("   ✅ Test files mentioned in output")
		}
	}

	// Test 3: Search command
	fmt.Println("\nTest 3: Search Command...")
	searchCmd := &types.Command{
		Type: types.CommandSearch,
		Args: []string{"Hello"},
	}

	result3, err := claudeClient.ExecuteCommand(ctx, searchCmd)
	if err != nil {
		log.Printf("❌ FAILED: Search command error: %v", err)
	} else if result3 != nil {
		fmt.Println("✅ SUCCESS: Search command executed")
		if strings.Contains(result3.Output, "test1.txt") || strings.Contains(result3.Output, "test2.go") {
			fmt.Println("   ✅ Search found matches in test files")
		}
		fmt.Printf("   Results preview: %s...\n", strings.TrimSpace(result3.Output)[:min(100, len(result3.Output))])
	}

	// Test 4: Write file command
	fmt.Println("\nTest 4: Write File Command...")
	testFile3 := filepath.Join(testDir, "test3.txt")
	writeCmd := &types.Command{
		Type: types.CommandWrite,
		Args: []string{testFile3, "This file was created by the SDK!"},
	}

	result4, err := claudeClient.ExecuteCommand(ctx, writeCmd)
	if err != nil {
		log.Printf("❌ FAILED: Write command error: %v", err)
	} else if result4 != nil {
		fmt.Println("✅ SUCCESS: Write command executed")
		// Verify file was created
		if _, err := os.Stat(testFile3); err == nil {
			content, _ := os.ReadFile(testFile3)
			fmt.Printf("   File created with content: %s\n", string(content))
			fmt.Println("   ✅ File creation verified")
		}
	}

	// Test 5: Slash command
	fmt.Println("\nTest 5: Slash Command Execution...")
	slashResult, err := claudeClient.ExecuteSlashCommand(ctx, "/help")
	if err != nil {
		log.Printf("❌ FAILED: Slash command error: %v", err)
	} else if slashResult != nil {
		fmt.Println("✅ SUCCESS: Slash command executed")
		fmt.Printf("   Output preview: %s...\n", strings.TrimSpace(slashResult.Output)[:min(100, len(slashResult.Output))])
	}

	fmt.Println("\n=== Command Execution Tests Complete ===")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
