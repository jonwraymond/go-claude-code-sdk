// Package main demonstrates quick start examples for Claude Code SDK.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/claudecode"
)

func basicExample() {
	fmt.Println("=== Basic Example ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	msgChan := claudecode.Query(ctx, "What is 2 + 2?", nil)
	for msg := range msgChan {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					fmt.Printf("Claude: %s\n", textBlock.Text)
				}
			}
		}
	}
	fmt.Println()
}

func withOptionsExample() {
	fmt.Println("=== With Options Example ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	options := claudecode.NewClaudeCodeOptions()
	options.SystemPrompt = claudecode.StringPtr("You are a helpful assistant that explains things simply.")
	options.MaxTurns = claudecode.IntPtr(1)

	msgChan := claudecode.Query(ctx, "Explain what Python is in one sentence.", options)
	for msg := range msgChan {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					fmt.Printf("Claude: %s\n", textBlock.Text)
				}
			}
		}
	}
	fmt.Println()
}

func withToolsExample() {
	fmt.Println("=== With Tools Example ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Read", "Write"}
	options.SystemPrompt = claudecode.StringPtr("You are a helpful file assistant.")

	msgChan := claudecode.Query(ctx, "Create a file called hello.txt with 'Hello, World!' in it", options)
	for msg := range msgChan {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					fmt.Printf("Claude: %s\n", textBlock.Text)
				}
			}
		} else if resultMsg, ok := msg.(*claudecode.ResultMessage); ok {
			if resultMsg.TotalCostUSD != nil && *resultMsg.TotalCostUSD > 0 {
				fmt.Printf("\nCost: $%.4f\n", *resultMsg.TotalCostUSD)
			}
		}
	}
	fmt.Println()
}

func syncExample() {
	fmt.Println("=== Synchronous Example ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages, err := claudecode.QuerySync(ctx, "What is the capital of Japan?", nil)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Received %d messages:\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("Message %d: %T\n", i+1, msg)
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					fmt.Printf("  Content: %s\n", textBlock.Text)
				}
			}
		}
	}
	fmt.Println()
}

func main() {
	fmt.Println("Claude Code SDK - Go Quick Start Examples")
	fmt.Println("==========================================")

	basicExample()
	withOptionsExample()
	withToolsExample()
	syncExample()

	fmt.Println("All examples completed!")
}