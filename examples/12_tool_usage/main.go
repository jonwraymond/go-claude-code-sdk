package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/claudecode"
)

func main() {
	fmt.Println("=== Tool Usage Examples ===")

	// Example 1: Basic tool usage
	example1BasicToolUsage()

	// Example 2: Tool restrictions
	example2ToolRestrictions()

	// Example 3: Tool result handling
	example3ToolResultHandling()

	// Example 4: Complex tool workflows
	example4ComplexToolWorkflows()

	// Example 5: Tool error handling
	example5ToolErrorHandling()

	// Example 6: Custom tool patterns
	example6CustomToolPatterns()
}

func example1BasicToolUsage() {
	fmt.Println("Example 1: Basic Tool Usage")
	fmt.Println("--------------------------")

	// Enable common tools
	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Read", "Write", "Edit", "Bash"}

	ctx := context.Background()

	// Test different tools
	toolTests := []struct {
		name          string
		query         string
		expectedTools []string
	}{
		{
			name:          "File creation",
			query:         "Create a file called hello.txt with 'Hello, World!' content",
			expectedTools: []string{"Write"},
		},
		{
			name:          "File reading",
			query:         "Read the contents of hello.txt",
			expectedTools: []string{"Read"},
		},
		{
			name:          "File editing",
			query:         "Edit hello.txt and change 'World' to 'Claude'",
			expectedTools: []string{"Edit"},
		},
		{
			name:          "Command execution",
			query:         "Run 'echo Test Complete' using bash",
			expectedTools: []string{"Bash"},
		},
	}

	for _, test := range toolTests {
		fmt.Printf("\n🔧 %s\n", test.name)
		fmt.Printf("   Query: %s\n", test.query)

		msgChan := claudecode.Query(ctx, test.query, options)

		toolsUsed := make(map[string]int)

		for msg := range msgChan {
			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range m.Content {
					if toolUse, ok := block.(claudecode.ToolUseBlock); ok {
						toolsUsed[toolUse.Name]++
						fmt.Printf("   📌 Tool used: %s (ID: %s)\n", toolUse.Name, toolUse.ID)

						// Show tool input
						if inputJSON, err := json.MarshalIndent(toolUse.Input, "      ", "  "); err == nil {
							fmt.Printf("      Input: %s\n", string(inputJSON))
						}
					}
				}
			case *claudecode.SystemMessage:
				if m.Data != nil {
					if toolResult, ok := m.Data["tool_result"]; ok {
						fmt.Printf("   ✅ Tool result: %v\n", toolResult)
					}
				}
			}
		}

		fmt.Printf("   Summary: Used %d tools\n", len(toolsUsed))
		for tool, count := range toolsUsed {
			fmt.Printf("      %s: %d times\n", tool, count)
		}
	}
	fmt.Println()
}

func example2ToolRestrictions() {
	fmt.Println("Example 2: Tool Restrictions")
	fmt.Println("----------------------------")

	ctx := context.Background()

	// Test different tool restriction scenarios
	scenarios := []struct {
		name          string
		allowedTools  []string
		query         string
		shouldSucceed bool
	}{
		{
			name:          "Only Read allowed",
			allowedTools:  []string{"Read"},
			query:         "Create a new file called test.txt",
			shouldSucceed: false,
		},
		{
			name:          "Only Write allowed",
			allowedTools:  []string{"Write"},
			query:         "Create a new file called allowed.txt with content 'This works'",
			shouldSucceed: true,
		},
		{
			name:          "No tools allowed",
			allowedTools:  []string{},
			query:         "Tell me about Go programming",
			shouldSucceed: true,
		},
		{
			name:          "Multiple tools allowed",
			allowedTools:  []string{"Read", "Write", "Edit"},
			query:         "Create config.json, read it, then edit it to add a new field",
			shouldSucceed: true,
		},
	}

	for _, scenario := range scenarios {
		fmt.Printf("\n🚦 Scenario: %s\n", scenario.name)
		fmt.Printf("   Allowed tools: %v\n", scenario.allowedTools)
		fmt.Printf("   Query: %s\n", scenario.query)

		options := claudecode.NewClaudeCodeOptions()
		options.AllowedTools = scenario.allowedTools

		msgChan := claudecode.Query(ctx, scenario.query, options)

		toolAttempts := 0
		toolSuccesses := 0

		for msg := range msgChan {
			if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
				for _, block := range assistantMsg.Content {
					if toolUse, ok := block.(claudecode.ToolUseBlock); ok {
						toolAttempts++
						allowed := false
						for _, allowedTool := range scenario.allowedTools {
							if toolUse.Name == allowedTool {
								allowed = true
								toolSuccesses++
								break
							}
						}

						if allowed {
							fmt.Printf("   ✅ Tool %s: Allowed\n", toolUse.Name)
						} else {
							fmt.Printf("   ❌ Tool %s: Not allowed\n", toolUse.Name)
						}
					}
				}
			}
		}

		fmt.Printf("   Result: %d/%d tools allowed\n", toolSuccesses, toolAttempts)
		if scenario.shouldSucceed && toolSuccesses > 0 || !scenario.shouldSucceed && toolSuccesses == 0 {
			fmt.Println("   ✅ Behaved as expected")
		} else {
			fmt.Println("   ❓ Unexpected behavior")
		}
	}
	fmt.Println()
}

func example3ToolResultHandling() {
	fmt.Println("Example 3: Tool Result Handling")
	fmt.Println("-------------------------------")

	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Read", "Write", "Edit", "Bash"}

	ctx := context.Background()

	// Track tool results
	toolResults := []ToolResult{}
	var resultsMu sync.Mutex

	// Complex query that uses multiple tools
	query := `Please do the following:
1. Create a file called data.json with sample user data
2. Read the file to verify it was created
3. Edit the file to add a timestamp field
4. Use bash to check the file size`

	fmt.Printf("📝 Query: %s\n\n", query)

	msgChan := claudecode.Query(ctx, query, options)

	for msg := range msgChan {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range m.Content {
				switch b := block.(type) {
				case claudecode.ToolUseBlock:
					fmt.Printf("🔧 Tool Use: %s (ID: %s)\n", b.Name, b.ID)

					result := ToolResult{
						ToolName:  b.Name,
						ToolID:    b.ID,
						Input:     b.Input,
						StartTime: time.Now(),
					}

					resultsMu.Lock()
					toolResults = append(toolResults, result)
					resultsMu.Unlock()

				case claudecode.ToolResultBlock:
					fmt.Printf("📊 Tool Result (ID: %s)\n", b.ToolUseID)
					fmt.Printf("   Success: %v\n", b.IsError == nil || !*b.IsError)

					// Update result
					resultsMu.Lock()
					for i := range toolResults {
						if toolResults[i].ToolID == b.ToolUseID {
							toolResults[i].EndTime = time.Now()
							toolResults[i].Result = b.Content
							toolResults[i].IsError = b.IsError != nil && *b.IsError
							toolResults[i].Duration = toolResults[i].EndTime.Sub(toolResults[i].StartTime)
							break
						}
					}
					resultsMu.Unlock()

					// Show content preview
					contentStr := fmt.Sprintf("%v", b.Content)
					if len(contentStr) > 100 {
						contentStr = contentStr[:100] + "..."
					}
					fmt.Printf("   Content: %s\n", contentStr)
					fmt.Println()
				}
			}
		}
	}

	// Summary of tool results
	fmt.Println("\n📊 Tool Results Summary:")
	fmt.Printf("   Total tools used: %d\n", len(toolResults))

	for i, result := range toolResults {
		fmt.Printf("\n   [%d] %s:\n", i+1, result.ToolName)
		fmt.Printf("       Duration: %v\n", result.Duration.Round(time.Millisecond))
		fmt.Printf("       Success: %v\n", !result.IsError)
		if result.IsError {
			fmt.Printf("       ❌ Error occurred\n")
		}
	}
	fmt.Println()
}

func example4ComplexToolWorkflows() {
	fmt.Println("Example 4: Complex Tool Workflows")
	fmt.Println("---------------------------------")

	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Read", "Write", "Edit", "Bash", "MultiEdit"}

	ctx := context.Background()

	// Workflow 1: Code generation and testing
	fmt.Println("\n🔄 Workflow 1: Code Generation and Testing")

	workflow1 := `Create a complete Go program that:
1. Defines a Calculator struct with Add, Subtract, Multiply, Divide methods
2. Includes proper error handling for division by zero
3. Create a test file with unit tests
4. Run the tests using go test`

	executeWorkflow(ctx, "Code Generation", workflow1, options)

	// Workflow 2: Data processing pipeline
	fmt.Println("\n🔄 Workflow 2: Data Processing Pipeline")

	workflow2 := `Create a data processing pipeline:
1. Create a CSV file with sample sales data (product, quantity, price)
2. Write a Python script to read the CSV and calculate total revenue
3. Execute the script and show the results
4. Create a summary report file with the findings`

	executeWorkflow(ctx, "Data Pipeline", workflow2, options)

	// Workflow 3: Configuration management
	fmt.Println("\n🔄 Workflow 3: Configuration Management")

	workflow3 := `Set up a configuration system:
1. Create a base config.yaml with application settings
2. Create environment-specific configs (dev.yaml, prod.yaml)
3. Write a script to merge configurations
4. Test the configuration loading`

	executeWorkflow(ctx, "Config Management", workflow3, options)
	fmt.Println()
}

func example5ToolErrorHandling() {
	fmt.Println("Example 5: Tool Error Handling")
	fmt.Println("------------------------------")

	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Read", "Write", "Edit", "Bash"}

	ctx := context.Background()

	// Test error scenarios
	errorScenarios := []struct {
		name          string
		query         string
		expectedError string
	}{
		{
			name:          "Read non-existent file",
			query:         "Read the contents of /definitely/does/not/exist/file.txt",
			expectedError: "file not found",
		},
		{
			name:          "Write to protected location",
			query:         "Create a file at /etc/protected.conf",
			expectedError: "permission denied",
		},
		{
			name:          "Edit with invalid old_string",
			query:         "Edit hello.txt and replace 'NonExistentString' with 'NewString'",
			expectedError: "string not found",
		},
		{
			name:          "Bash command failure",
			query:         "Run the command 'exit 1' and check the exit code",
			expectedError: "non-zero exit",
		},
	}

	for _, scenario := range errorScenarios {
		fmt.Printf("\n❌ Testing: %s\n", scenario.name)
		fmt.Printf("   Query: %s\n", scenario.query)

		msgChan := claudecode.Query(ctx, scenario.query, options)

		errorCount := 0
		var errors []string

		for msg := range msgChan {
			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range m.Content {
					if toolResult, ok := block.(claudecode.ToolResultBlock); ok {
						if toolResult.IsError != nil && *toolResult.IsError {
							errorCount++
							errorStr := fmt.Sprintf("%v", toolResult.Content)
							errors = append(errors, errorStr)
							fmt.Printf("   🚨 Tool error detected: %s\n", errorStr)
						}
					}
				}
			case *claudecode.SystemMessage:
				if m.Subtype == "error" {
					errorCount++
					if errMsg, ok := m.Data["error"].(string); ok {
						errors = append(errors, errMsg)
						fmt.Printf("   🚨 System error: %s\n", errMsg)
					}
				}
			}
		}

		fmt.Printf("   Total errors: %d\n", errorCount)
		if errorCount > 0 {
			fmt.Println("   ✅ Error handling working correctly")
		} else {
			fmt.Println("   ⚠️ Expected error but none occurred")
		}
	}
	fmt.Println()
}

func example6CustomToolPatterns() {
	fmt.Println("Example 6: Custom Tool Patterns")
	fmt.Println("-------------------------------")

	ctx := context.Background()

	// Pattern 1: Tool chaining
	fmt.Println("\n🔗 Pattern 1: Tool Chaining")

	options1 := claudecode.NewClaudeCodeOptions()
	options1.AllowedTools = []string{"Write", "Read", "Edit", "Bash"}

	chainQuery := `Create a chain of operations:
1. Write initial.txt with 'Step 1'
2. Read initial.txt
3. Edit initial.txt to append ' -> Step 2'
4. Read the updated file
5. Use bash to copy it to final.txt`

	msgChan := claudecode.Query(ctx, chainQuery, options1)

	toolChain := []string{}

	for msg := range msgChan {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if toolUse, ok := block.(claudecode.ToolUseBlock); ok {
					toolChain = append(toolChain, toolUse.Name)
				}
			}
		}
	}

	fmt.Printf("   Tool chain: %v\n", toolChain)
	fmt.Printf("   Chain length: %d tools\n", len(toolChain))

	// Pattern 2: Conditional tool usage
	fmt.Println("\n🎯 Pattern 2: Conditional Tool Usage")

	conditionalQuery := `Check if config.json exists:
- If it exists, read it and add a new field
- If it doesn't exist, create it with default values`

	msgChan2 := claudecode.Query(ctx, conditionalQuery, options1)

	conditionMet := false
	toolsAfterCondition := []string{}

	for msg := range msgChan2 {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					if contains(textBlock.Text, "exists") || contains(textBlock.Text, "doesn't exist") {
						conditionMet = true
					}
				}
				if toolUse, ok := block.(claudecode.ToolUseBlock); ok && conditionMet {
					toolsAfterCondition = append(toolsAfterCondition, toolUse.Name)
				}
			}
		}
	}

	fmt.Printf("   Conditional logic detected: %v\n", conditionMet)
	fmt.Printf("   Tools used after condition: %v\n", toolsAfterCondition)

	// Pattern 3: Batch operations
	fmt.Println("\n📦 Pattern 3: Batch Operations")

	batchQuery := `Create multiple files in one operation:
- file1.txt with 'Content 1'
- file2.txt with 'Content 2'  
- file3.txt with 'Content 3'
Then list all files using bash ls command`

	msgChan3 := claudecode.Query(ctx, batchQuery, options1)

	writeCount := 0
	bashCount := 0

	for msg := range msgChan3 {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if toolUse, ok := block.(claudecode.ToolUseBlock); ok {
					switch toolUse.Name {
					case "Write":
						writeCount++
					case "Bash":
						bashCount++
					}
				}
			}
		}
	}

	fmt.Printf("   Batch writes: %d\n", writeCount)
	fmt.Printf("   Verification commands: %d\n", bashCount)

	// Pattern 4: Tool retry pattern
	fmt.Println("\n🔄 Pattern 4: Tool Retry Pattern")

	retryQuery := `Try to create a file in /tmp/test-retry.txt:
- If it fails, try again with a different approach
- Ensure the file is created successfully`

	msgChan4 := claudecode.Query(ctx, retryQuery, options1)

	attempts := 0
	successes := 0

	for msg := range msgChan4 {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for i, block := range assistantMsg.Content {
				if toolUse, ok := block.(claudecode.ToolUseBlock); ok {
					if toolUse.Name == "Write" {
						attempts++
					}
				}
				// Check next block for result
				if i+1 < len(assistantMsg.Content) {
					if toolResult, ok := assistantMsg.Content[i+1].(claudecode.ToolResultBlock); ok {
						if toolResult.IsError == nil || !*toolResult.IsError {
							successes++
						}
					}
				}
			}
		}
	}

	fmt.Printf("   Write attempts: %d\n", attempts)
	fmt.Printf("   Successful writes: %d\n", successes)
	if successes > 0 {
		fmt.Println("   ✅ Retry pattern successful")
	}

	fmt.Println()
}

// Helper types and functions

type ToolResult struct {
	ToolName  string
	ToolID    string
	Input     interface{}
	Result    interface{}
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	IsError   bool
}

func executeWorkflow(ctx context.Context, name string, query string, options *claudecode.ClaudeCodeOptions) {
	fmt.Printf("\n📋 Executing: %s\n", name)

	start := time.Now()
	msgChan := claudecode.Query(ctx, query, options)

	toolSequence := []string{}
	stepCount := 0

	for msg := range msgChan {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				switch b := block.(type) {
				case claudecode.TextBlock:
					// Count steps mentioned
					text := b.Text
					for i := 1; i <= 10; i++ {
						stepMarker := fmt.Sprintf("%d.", i)
						if contains(text, stepMarker) {
							stepCount = i
						}
					}
				case claudecode.ToolUseBlock:
					toolSequence = append(toolSequence, b.Name)
					fmt.Printf("   → %s\n", b.Name)
				}
			}
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("   ✅ Workflow complete\n")
	fmt.Printf("      Steps detected: %d\n", stepCount)
	fmt.Printf("      Tools used: %d\n", len(toolSequence))
	fmt.Printf("      Duration: %v\n", elapsed.Round(time.Millisecond))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
