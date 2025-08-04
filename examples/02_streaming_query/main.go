package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/claudecode"
)

func main() {
	fmt.Println("=== Streaming Query Examples ===")

	// Example 1: Basic streaming with real-time output
	example1BasicStreaming()

	// Example 2: Streaming with progress tracking
	example2StreamingProgress()

	// Example 3: Streaming code generation
	example3StreamingCode()

	// Example 4: Streaming with tool use
	example4StreamingTools()
}

func example1BasicStreaming() {
	fmt.Println("Example 1: Basic Streaming")
	fmt.Println("--------------------------")

	ctx := context.Background()
	msgChan := claudecode.Query(ctx, "Write a short story about a robot learning to paint. Stream it word by word.", nil)

	var fullStory strings.Builder
	tokenCount := 0

	for msg := range msgChan {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range m.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					// Simulate streaming by showing partial content
					words := strings.Fields(textBlock.Text)
					for _, word := range words {
						fmt.Print(word + " ")
						fullStory.WriteString(word + " ")
						tokenCount++
						if tokenCount%10 == 0 {
							fmt.Println() // New line every 10 words
						}
					}
				}
			}
		case *claudecode.ResultMessage:
			fmt.Printf("\n\nStreaming completed!")
			fmt.Printf("\nTotal words: %d", tokenCount)
			fmt.Printf("\nDuration: %dms\n\n", m.DurationMs)
		}
	}
}

func example2StreamingProgress() {
	fmt.Println("Example 2: Streaming with Progress")
	fmt.Println("----------------------------------")

	ctx := context.Background()
	options := claudecode.NewClaudeCodeOptions()
	options.SystemPrompt = claudecode.StringPtr("When asked to count, show progress indicators.")

	msgChan := claudecode.Query(ctx, "Count from 1 to 20, showing progress", options)

	for msg := range msgChan {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range m.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					// Show streaming progress
					lines := strings.Split(textBlock.Text, "\n")
					for _, line := range lines {
						if strings.TrimSpace(line) != "" {
							fmt.Printf("\r%s", line)
							// In real streaming, each line would come separately
						}
					}
				}
			}
		case *claudecode.ResultMessage:
			fmt.Printf("\n\nProgress tracking completed in %dms\n\n", m.DurationMs)
		}
	}
}

func example3StreamingCode() {
	fmt.Println("Example 3: Streaming Code Generation")
	fmt.Println("------------------------------------")

	ctx := context.Background()
	msgChan := claudecode.Query(ctx, "Write a Go function that implements binary search. Include comments.", nil)

	fmt.Println("Generated code:")
	fmt.Println("```go")

	for msg := range msgChan {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range m.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					// Extract code blocks
					if strings.Contains(textBlock.Text, "```") {
						// Parse and display code
						code := extractCode(textBlock.Text)
						fmt.Print(code)
					} else {
						fmt.Print(textBlock.Text)
					}
				}
			}
		case *claudecode.ResultMessage:
			fmt.Println("```")
			fmt.Printf("\nCode generation completed in %dms\n\n", m.DurationMs)
		}
	}
}

func example4StreamingTools() {
	fmt.Println("Example 4: Streaming with Tool Usage")
	fmt.Println("------------------------------------")

	ctx := context.Background()
	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Read", "Write", "Bash"}

	msgChan := claudecode.Query(ctx, "Create a project structure for a web server with main.go, handlers/, and middleware/ directories", options)

	toolUseCount := 0
	var lastToolUse string

	for msg := range msgChan {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range m.Content {
				switch b := block.(type) {
				case claudecode.TextBlock:
					if b.Text != "" {
						fmt.Println("Assistant:", b.Text)
					}
				case claudecode.ToolUseBlock:
					toolUseCount++
					lastToolUse = b.Name
					fmt.Printf("\n🔧 Tool Use #%d: %s\n", toolUseCount, b.Name)

					// Show tool parameters
					for key, value := range b.Input {
						fmt.Printf("   %s: %v\n", key, value)
					}
				}
			}
		case *claudecode.SystemMessage:
			if m.Subtype == "tool_result" {
				fmt.Printf("✅ %s completed\n", lastToolUse)
			}
		case *claudecode.ResultMessage:
			fmt.Printf("\n📊 Summary:")
			fmt.Printf("\n   Tools used: %d times", toolUseCount)
			fmt.Printf("\n   Duration: %dms", m.DurationMs)
			if m.TotalCostUSD != nil {
				fmt.Printf("\n   Cost: $%.6f", *m.TotalCostUSD)
			}
			fmt.Println()
		}
	}
}

// Helper function to extract code from markdown
func extractCode(text string) string {
	lines := strings.Split(text, "\n")
	var code strings.Builder
	inCode := false

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			inCode = !inCode
			continue
		}
		if inCode {
			code.WriteString(line + "\n")
		}
	}

	return code.String()
}
