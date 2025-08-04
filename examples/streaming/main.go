// Package main demonstrates comprehensive streaming mode examples.
//
// This file demonstrates various patterns for building applications with
// the ClaudeSDKClient streaming interface.
//
// The queries are intentionally simplistic. In reality, a query can be a more
// complex task that Claude SDK uses its agentic capabilities and tools (e.g. run
// bash commands, edit files, search the web, fetch web content) to accomplish.
//
// Usage:
//
//	go run streaming_mode.go                    - List the examples
//	go run streaming_mode.go all                - Run all examples
//	go run streaming_mode.go basic_streaming    - Run a specific example
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/claudecode"
)

// displayMessage provides standardized message display function.
//
// - UserMessage: "User: <content>"
// - AssistantMessage: "Claude: <content>"
// - SystemMessage: ignored
// - ResultMessage: "Result ended" + cost if available
func displayMessage(msg claudecode.Message) {
	switch m := msg.(type) {
	case *claudecode.UserMessage:
		if content, ok := m.Content.(string); ok {
			fmt.Printf("User: %s\n", content)
		}
	case *claudecode.AssistantMessage:
		for _, block := range m.Content {
			if textBlock, ok := block.(claudecode.TextBlock); ok {
				fmt.Printf("Claude: %s\n", textBlock.Text)
			}
		}
	case *claudecode.SystemMessage:
		// Ignore system messages
	case *claudecode.ResultMessage:
		fmt.Println("Result ended")
		if m.TotalCostUSD != nil && *m.TotalCostUSD > 0 {
			fmt.Printf("Cost: $%.4f\n", *m.TotalCostUSD)
		}
	}
}

func exampleBasicStreaming() {
	fmt.Println("=== Basic Streaming Example ===")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}

	fmt.Println("User: What is 2+2?")
	if err := client.Query(ctx, "What is 2+2?", "default"); err != nil {
		log.Printf("Failed to send query: %v", err)
		return
	}

	// Receive complete response using the helper method
	for msg := range client.ReceiveResponse(ctx) {
		displayMessage(msg)
	}

	fmt.Println()
}

func exampleMultiTurnConversation() {
	fmt.Println("=== Multi-Turn Conversation Example ===")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}

	// First turn
	fmt.Println("User: What's the capital of France?")
	if err := client.Query(ctx, "What's the capital of France?", "default"); err != nil {
		log.Printf("Failed to send query: %v", err)
		return
	}

	// Extract and print response
	for msg := range client.ReceiveResponse(ctx) {
		displayMessage(msg)
	}

	// Second turn - follow-up
	fmt.Println("\nUser: What's the population of that city?")
	if err := client.Query(ctx, "What's the population of that city?", "default"); err != nil {
		log.Printf("Failed to send query: %v", err)
		return
	}

	for msg := range client.ReceiveResponse(ctx) {
		displayMessage(msg)
	}

	fmt.Println()
}

func exampleConcurrentResponses() {
	fmt.Println("=== Concurrent Send/Receive Example ===")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}

	// Background goroutine to continuously receive messages
	go func() {
		for msg := range client.ReceiveMessages() {
			displayMessage(msg)
		}
	}()

	// Send multiple messages with delays
	questions := []string{
		"What is 2 + 2?",
		"What is the square root of 144?",
		"What is 10% of 80?",
	}

	for _, question := range questions {
		fmt.Printf("\nUser: %s\n", question)
		if err := client.Query(ctx, question, "default"); err != nil {
			log.Printf("Failed to send query: %v", err)
			continue
		}
		time.Sleep(3 * time.Second) // Wait between messages
	}

	// Give time for final responses
	time.Sleep(5 * time.Second)

	fmt.Println()
}

func exampleWithInterrupt() {
	fmt.Println("=== Interrupt Example ===")
	fmt.Println("IMPORTANT: Interrupts require active message consumption.")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}

	// Start a long-running task
	fmt.Println("\nUser: Count from 1 to 100 slowly")
	if err := client.Query(ctx, "Count from 1 to 100 slowly, with a brief pause between each number", "default"); err != nil {
		log.Printf("Failed to send query: %v", err)
		return
	}

	// Create background task to consume messages
	messagesReceived := []claudecode.Message{}
	interruptSent := false

	go func() {
		// Consume messages in the background to enable interrupt processing
		for msg := range client.ReceiveMessages() {
			messagesReceived = append(messagesReceived, msg)
			if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
				for _, block := range assistantMsg.Content {
					if textBlock, ok := block.(claudecode.TextBlock); ok {
						// Print first few numbers
						text := textBlock.Text
						if len(text) > 50 {
							text = text[:50] + "..."
						}
						fmt.Printf("Claude: %s\n", text)
					}
				}
			} else if _, ok := msg.(*claudecode.ResultMessage); ok {
				displayMessage(msg)
				if interruptSent {
					return
				}
			}
		}
	}()

	// Wait 2 seconds then send interrupt
	time.Sleep(2 * time.Second)
	fmt.Println("\n[After 2 seconds, sending interrupt...]")
	interruptSent = true
	if err := client.Interrupt(); err != nil {
		log.Printf("Failed to interrupt: %v", err)
	}

	// Wait a bit for interrupt to be processed
	time.Sleep(1 * time.Second)

	// Send new instruction after interrupt
	fmt.Println("\nUser: Never mind, just tell me a quick joke")
	if err := client.Query(ctx, "Never mind, just tell me a quick joke", "default"); err != nil {
		log.Printf("Failed to send query: %v", err)
		return
	}

	// Get the joke
	for msg := range client.ReceiveResponse(ctx) {
		displayMessage(msg)
	}

	fmt.Println()
}

func exampleManualMessageHandling() {
	fmt.Println("=== Manual Message Handling Example ===")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}

	if err := client.Query(ctx, "List 5 programming languages and their main use cases", "default"); err != nil {
		log.Printf("Failed to send query: %v", err)
		return
	}

	// Manually process messages with custom logic
	languagesFound := []string{}

	for msg := range client.ReceiveMessages() {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					text := textBlock.Text
					fmt.Printf("Claude: %s\n", text)
					// Custom logic: extract language names
					languages := []string{"Python", "JavaScript", "Java", "C++", "Go", "Rust", "Ruby"}
					for _, lang := range languages {
						if strings.Contains(text, lang) {
							found := false
							for _, found_lang := range languagesFound {
								if found_lang == lang {
									found = true
									break
								}
							}
							if !found {
								languagesFound = append(languagesFound, lang)
								fmt.Printf("Found language: %s\n", lang)
							}
						}
					}
				}
			}
		} else if resultMsg, ok := msg.(*claudecode.ResultMessage); ok {
			displayMessage(resultMsg)
			fmt.Printf("Total languages mentioned: %d\n", len(languagesFound))
			break
		}
	}

	fmt.Println()
}

func exampleWithOptions() {
	fmt.Println("=== Custom Options Example ===")

	// Configure options
	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Read", "Write"} // Allow file operations
	options.MaxThinkingTokens = 10000
	options.SystemPrompt = claudecode.StringPtr("You are a helpful coding assistant.")

	client := claudecode.NewClaudeSDKClient(options)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}

	fmt.Println("User: Create a simple hello.txt file with a greeting message")
	if err := client.Query(ctx, "Create a simple hello.txt file with a greeting message", "default"); err != nil {
		log.Printf("Failed to send query: %v", err)
		return
	}

	toolUses := []string{}
	for msg := range client.ReceiveResponse(ctx) {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			displayMessage(msg)
			for _, block := range assistantMsg.Content {
				if toolUseBlock, ok := block.(claudecode.ToolUseBlock); ok {
					toolUses = append(toolUses, toolUseBlock.Name)
				}
			}
		} else {
			displayMessage(msg)
		}
	}

	if len(toolUses) > 0 {
		fmt.Printf("Tools used: %s\n", strings.Join(toolUses, ", "))
	}

	fmt.Println()
}

func exampleBashCommand() {
	fmt.Println("=== Bash Command Example ===")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}

	fmt.Println("User: Run a bash echo command")
	if err := client.Query(ctx, "Run a bash echo command that says 'Hello from bash!'", "default"); err != nil {
		log.Printf("Failed to send query: %v", err)
		return
	}

	// Track all message types received
	messageTypes := []string{}

	for msg := range client.ReceiveMessages() {
		msgType := fmt.Sprintf("%T", msg)
		messageTypes = append(messageTypes, msgType)

		if userMsg, ok := msg.(*claudecode.UserMessage); ok {
			// User messages can contain tool results
			if content, ok := userMsg.Content.(string); ok {
				fmt.Printf("User: %s\n", content)
			}
		} else if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			// Assistant messages can contain tool use blocks
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					fmt.Printf("Claude: %s\n", textBlock.Text)
				} else if toolUseBlock, ok := block.(claudecode.ToolUseBlock); ok {
					fmt.Printf("Tool Use: %s (id: %s)\n", toolUseBlock.Name, toolUseBlock.ID)
					if toolUseBlock.Name == "Bash" {
						if command, ok := toolUseBlock.Input["command"].(string); ok {
							fmt.Printf("  Command: %s\n", command)
						}
					}
				}
			}
		} else if resultMsg, ok := msg.(*claudecode.ResultMessage); ok {
			fmt.Println("Result ended")
			if resultMsg.TotalCostUSD != nil {
				fmt.Printf("Cost: $%.4f\n", *resultMsg.TotalCostUSD)
			}
			break
		}
	}

	// Remove duplicates and display
	uniqueTypes := make(map[string]bool)
	for _, msgType := range messageTypes {
		uniqueTypes[msgType] = true
	}
	var types []string
	for msgType := range uniqueTypes {
		types = append(types, msgType)
	}

	fmt.Printf("\nMessage types received: %s\n", strings.Join(types, ", "))
	fmt.Println()
}

func exampleErrorHandling() {
	fmt.Println("=== Error Handling Example ===")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // Short timeout for demo
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Printf("Connection error: %v", err)
		return
	}

	// Send a message that will take time to process
	fmt.Println("User: Run a bash sleep command for 60 seconds")
	if err := client.Query(ctx, "Run a bash sleep command for 60 seconds", "default"); err != nil {
		log.Printf("Query error: %v", err)
		return
	}

	// Try to receive response with the context timeout
	messages := []claudecode.Message{}
	for msg := range client.ReceiveResponse(ctx) {
		messages = append(messages, msg)
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					text := textBlock.Text
					if len(text) > 50 {
						text = text[:50] + "..."
					}
					fmt.Printf("Claude: %s\n", text)
				}
			}
		} else if _, ok := msg.(*claudecode.ResultMessage); ok {
			displayMessage(msg)
			break
		}
	}

	// Check if context was cancelled due to timeout
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Printf("\nResponse timeout after 15 seconds - demonstrating graceful handling\n")
		fmt.Printf("Received %d messages before timeout\n", len(messages))
	}

	fmt.Println()
}

func main() {
	examples := map[string]func(){
		"basic_streaming":         exampleBasicStreaming,
		"multi_turn_conversation": exampleMultiTurnConversation,
		"concurrent_responses":    exampleConcurrentResponses,
		"with_interrupt":          exampleWithInterrupt,
		"manual_message_handling": exampleManualMessageHandling,
		"with_options":            exampleWithOptions,
		"bash_command":            exampleBashCommand,
		"error_handling":          exampleErrorHandling,
	}

	if len(os.Args) < 2 {
		// List available examples
		fmt.Println("Usage: go run streaming_mode.go <example_name>")
		fmt.Println("\nAvailable examples:")
		fmt.Println("  all - Run all examples")
		for name := range examples {
			fmt.Printf("  %s\n", name)
		}
		os.Exit(0)
	}

	exampleName := os.Args[1]

	if exampleName == "all" {
		// Run all examples
		fmt.Println("Claude Code SDK - Go Streaming Examples")
		fmt.Println("========================================")

		for name, fn := range examples {
			fmt.Printf("Running example: %s\n", name)
			fn()
			fmt.Println(strings.Repeat("-", 50) + "\n")
		}
	} else if fn, exists := examples[exampleName]; exists {
		// Run specific example
		fn()
	} else {
		fmt.Printf("Error: Unknown example '%s'\n", exampleName)
		fmt.Println("\nAvailable examples:")
		fmt.Println("  all - Run all examples")
		for name := range examples {
			fmt.Printf("  %s\n", name)
		}
		os.Exit(1)
	}
}
