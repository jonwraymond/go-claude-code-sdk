//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Testing SDK CommandList Functionality ===")

	ctx := context.Background()
	config := &types.ClaudeCodeConfig{
		Model: "claude-3-5-sonnet-20241022",
	}

	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()

	// Create test files
	fmt.Println("\nPreparing test files...")
	writeCmd1 := client.WriteFile("test_list_1.txt", "Content of file 1")
	writeCmd2 := client.WriteFile("test_list_2.txt", "Content of file 2")
	writeCmd3 := client.WriteFile("test_list_3.go", "package main\n\nfunc main() {\n\tprintln(\"Hello from test\")\n}")

	setupList := client.NewCommandList(writeCmd1, writeCmd2, writeCmd3)
	setupResult, err := claudeClient.ExecuteCommands(ctx, setupList)
	if err != nil {
		log.Printf("Failed to create test files: %v", err)
		return
	}

	fmt.Printf("✅ Setup complete: %d/%d commands succeeded\n",
		setupResult.SuccessfulCommands, setupResult.TotalCommands)

	// Test 1: Sequential Command Execution
	fmt.Println("\nTest 1: Sequential Command Execution...")

	sequentialCmds := client.NewCommandList(
		client.ReadFile("test_list_1.txt"),
		client.ReadFile("test_list_2.txt"),
		client.SearchCode("Hello"),
	)

	startTime := time.Now()
	result1, err := claudeClient.ExecuteCommands(ctx, sequentialCmds)
	if err != nil {
		log.Printf("❌ FAILED: Sequential execution error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Sequential execution completed")
		fmt.Printf("   Total commands: %d\n", result1.TotalCommands)
		fmt.Printf("   Successful: %d\n", result1.SuccessfulCommands)
		fmt.Printf("   Failed: %d\n", result1.FailedCommands)
		fmt.Printf("   Execution time: %dms\n", result1.ExecutionTime)
		fmt.Printf("   Actual time: %v\n", time.Since(startTime))

		if result1.Success {
			fmt.Println("   ✅ All commands succeeded")
		}

		// Show individual results
		for i, cmdResult := range result1.Results {
			fmt.Printf("   Command %d: Success=%v\n", i+1, cmdResult.Success)
		}
	}

	// Test 2: Parallel Command Execution
	fmt.Println("\nTest 2: Parallel Command Execution...")

	parallelCmds := client.NewParallelCommandList(2,
		client.ReadFile("test_list_1.txt"),
		client.ReadFile("test_list_2.txt"),
		client.ReadFile("test_list_3.go"),
		client.GitStatus(),
	)

	startTime = time.Now()
	result2, err := claudeClient.ExecuteCommands(ctx, parallelCmds)
	if err != nil {
		log.Printf("❌ FAILED: Parallel execution error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Parallel execution completed")
		fmt.Printf("   Total commands: %d\n", result2.TotalCommands)
		fmt.Printf("   Successful: %d\n", result2.SuccessfulCommands)
		fmt.Printf("   Failed: %d\n", result2.FailedCommands)
		fmt.Printf("   Execution time: %dms\n", result2.ExecutionTime)
		fmt.Printf("   Actual time: %v\n", time.Since(startTime))
		fmt.Printf("   Max parallel: %d\n", parallelCmds.MaxParallel)

		if result2.ExecutionTime < result1.ExecutionTime {
			fmt.Println("   ✅ Parallel execution was faster than sequential")
		}
	}

	// Test 3: Error Handling with StopOnError
	fmt.Println("\nTest 3: Error Handling (StopOnError=true)...")

	errorCmds := client.CreateCommandList([]*types.Command{
		client.ReadFile("test_list_1.txt"),
		client.ReadFile("nonexistent_file.txt"), // This should fail
		client.ReadFile("test_list_2.txt"),      // This might not execute
	}, client.WithStopOnError(true))

	result3, err := claudeClient.ExecuteCommands(ctx, errorCmds)
	if err != nil {
		log.Printf("Execution error: %v", err)
	} else {
		fmt.Println("✅ Execution completed with errors handled")
		fmt.Printf("   Total commands: %d\n", result3.TotalCommands)
		fmt.Printf("   Executed: %d\n", len(result3.Results))
		fmt.Printf("   Successful: %d\n", result3.SuccessfulCommands)
		fmt.Printf("   Failed: %d\n", result3.FailedCommands)

		if !result3.Success {
			fmt.Println("   ✅ Overall success is false (as expected)")
		}

		if len(result3.Errors) > 0 {
			fmt.Println("   Errors:")
			for _, err := range result3.Errors {
				fmt.Printf("   - %s\n", err)
			}
		}

		// Check if execution stopped after error
		if len(result3.Results) == 2 {
			fmt.Println("   ✅ Execution stopped after first error (as expected)")
		}
	}

	// Test 4: Continue on Error
	fmt.Println("\nTest 4: Continue on Error (StopOnError=false)...")

	continueOnErrorCmds := client.CreateCommandList([]*types.Command{
		client.ReadFile("test_list_1.txt"),
		client.ReadFile("nonexistent_file.txt"), // This should fail
		client.ReadFile("test_list_2.txt"),      // This should still execute
	}, client.WithStopOnError(false))

	result4, err := claudeClient.ExecuteCommands(ctx, continueOnErrorCmds)
	if err != nil {
		log.Printf("Execution error: %v", err)
	} else {
		fmt.Println("✅ Execution completed (continued after error)")
		fmt.Printf("   Total commands: %d\n", result4.TotalCommands)
		fmt.Printf("   Executed: %d\n", len(result4.Results))
		fmt.Printf("   Successful: %d\n", result4.SuccessfulCommands)
		fmt.Printf("   Failed: %d\n", result4.FailedCommands)

		if len(result4.Results) == 3 {
			fmt.Println("   ✅ All commands were attempted (as expected)")
		}
	}

	// Test 5: Mixed Command Types
	fmt.Println("\nTest 5: Mixed Command Types...")

	mixedCmds := client.NewCommandList(
		client.AnalyzeCode("test_list_3.go"),
		client.SearchCode("func main"),
		client.GitStatus(),
	)

	result5, err := claudeClient.ExecuteCommands(ctx, mixedCmds)
	if err != nil {
		log.Printf("❌ FAILED: Mixed commands error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Mixed command types executed")
		fmt.Printf("   Commands executed: %d\n", result5.TotalCommands)
		fmt.Printf("   All succeeded: %v\n", result5.Success)
	}

	// Clean up
	fmt.Println("\nCleaning up test files...")
	cleanupList := client.NewCommandList(
		&types.Command{Type: types.CommandWrite, Args: []string{"test_list_1.txt", ""}},
		&types.Command{Type: types.CommandWrite, Args: []string{"test_list_2.txt", ""}},
		&types.Command{Type: types.CommandWrite, Args: []string{"test_list_3.go", ""}},
	)
	claudeClient.ExecuteCommands(ctx, cleanupList)

	fmt.Println("\n=== CommandList Tests Complete ===")

	// Summary
	fmt.Println("\nFeatures tested:")
	fmt.Println("- ✅ Sequential command execution")
	fmt.Println("- ✅ Parallel command execution with concurrency limit")
	fmt.Println("- ✅ Error handling with StopOnError option")
	fmt.Println("- ✅ Continue on error functionality")
	fmt.Println("- ✅ Mixed command types in single list")
	fmt.Println("- ✅ Execution time tracking")
	fmt.Println("- ✅ Comprehensive result aggregation")
}
