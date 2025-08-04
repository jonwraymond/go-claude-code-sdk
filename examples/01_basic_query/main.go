package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/claudecode"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Basic Query Examples ===\n")

	// Example 1: Simple query with default options
	example1SimpleQuery()

	// Example 2: Query with custom system prompt
	example2SystemPrompt()

	// Example 3: Query with specific tools allowed
	example3AllowedTools()

	// Example 4: QuerySync for synchronous operation
	example4QuerySync()

	// Example 5: Query with timeout
	example5QueryWithTimeout()
}

func example1SimpleQuery() {
	fmt.Println("Example 1: Simple Query")
	fmt.Println("----------------------")

	ctx := context.Background()
	msgChan := claudecode.Query(ctx, "What is the capital of France?", nil)

	for msg := range msgChan {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			fmt.Println("Claude's response:")
			for _, block := range m.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					fmt.Println(textBlock.Text)
				}
			}
		case *claudecode.ResultMessage:
			if m.TotalCostUSD != nil {
				fmt.Printf("Cost: $%.6f\n", *m.TotalCostUSD)
			}
			fmt.Printf("Duration: %dms\n\n", m.DurationMs)
		}
	}
}

func example2SystemPrompt() {
	fmt.Println("Example 2: Query with System Prompt")
	fmt.Println("-----------------------------------")

	options := claudecode.NewClaudeCodeOptions()
	options.SystemPrompt = claudecode.StringPtr("You are a helpful geography teacher. Always explain your answers in a educational way.")

	ctx := context.Background()
	msgChan := claudecode.Query(ctx, "Tell me about the geography of Japan", options)

	for msg := range msgChan {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			fmt.Println("Claude's educational response:")
			for _, block := range m.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					fmt.Println(textBlock.Text)
				}
			}
		case *claudecode.ResultMessage:
			fmt.Printf("Response completed in %dms\n\n", m.DurationMs)
		}
	}
}

func example3AllowedTools() {
	fmt.Println("Example 3: Query with Specific Tools")
	fmt.Println("------------------------------------")

	options := claudecode.NewClaudeCodeOptions()
	// Only allow Read and Write tools
	options.AllowedTools = []string{"Read", "Write"}
	options.SystemPrompt = claudecode.StringPtr("You are a code assistant. Use only Read and Write tools.")

	ctx := context.Background()
	msgChan := claudecode.Query(ctx, "Create a simple hello.txt file with 'Hello, World!' content", options)

	for msg := range msgChan {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range m.Content {
				switch b := block.(type) {
				case claudecode.TextBlock:
					fmt.Println("Text:", b.Text)
				case claudecode.ToolUseBlock:
					fmt.Printf("Tool used: %s\n", b.Name)
					fmt.Printf("Tool input: %v\n", b.Input)
				}
			}
		case *claudecode.SystemMessage:
			if m.Subtype == "tool_result" {
				fmt.Printf("Tool result: %v\n", m.Data)
			}
		case *claudecode.ResultMessage:
			fmt.Printf("Task completed in %dms\n\n", m.DurationMs)
		}
	}
}

func example4QuerySync() {
	fmt.Println("Example 4: Synchronous Query")
	fmt.Println("----------------------------")

	ctx := context.Background()
	messages, err := claudecode.QuerySync(ctx, "What is 2 + 2?", nil)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Received %d messages\n", len(messages))
	for _, msg := range messages {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range m.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					fmt.Println("Answer:", textBlock.Text)
				}
			}
		}
	}
	fmt.Println()
}

func example5QueryWithTimeout() {
	fmt.Println("Example 5: Query with Timeout")
	fmt.Println("-----------------------------")

	// Create a context with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msgChan := claudecode.Query(ctx, "Count from 1 to 100 slowly", nil)

	for msg := range msgChan {
		select {
		case <-ctx.Done():
			fmt.Println("Query timed out!")
			return
		default:
			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range m.Content {
					if textBlock, ok := block.(claudecode.TextBlock); ok {
						// Only print first 100 chars to avoid spam
						content := textBlock.Text
						if len(content) > 100 {
							content = content[:100] + "..."
						}
						fmt.Println("Claude:", content)
					}
				}
			case *claudecode.ResultMessage:
				fmt.Printf("Completed in %dms\n", m.DurationMs)
			}
		}
	}
}