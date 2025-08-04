package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/claudecode"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Interactive Client Examples ===\n")

	// Example 1: Basic interactive session
	example1BasicInteractive()

	// Example 2: Multi-turn conversation
	example2MultiTurnConversation()

	// Example 3: Stateful coding session
	example3StatefulCoding()

	// Example 4: Context-aware responses
	example4ContextAware()
}

func example1BasicInteractive() {
	fmt.Println("Example 1: Basic Interactive Session")
	fmt.Println("------------------------------------")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Start message processor
	go processMessages(client)

	// Send initial message
	fmt.Println("You: Hello! Can you help me with Go programming?")
	if err := client.Query(ctx, "Hello! Can you help me with Go programming?", "session1"); err != nil {
		log.Printf("Failed to send query: %v\n", err)
	}

	// Wait for response
	time.Sleep(3 * time.Second)

	// Send follow-up
	fmt.Println("\nYou: What are goroutines?")
	if err := client.Query(ctx, "What are goroutines?", "session1"); err != nil {
		log.Printf("Failed to send query: %v\n", err)
	}

	// Wait for response
	time.Sleep(3 * time.Second)
	fmt.Println("\n---Session 1 Complete---\n")
}

func example2MultiTurnConversation() {
	fmt.Println("Example 2: Multi-turn Conversation")
	fmt.Println("----------------------------------")

	options := claudecode.NewClaudeCodeOptions()
	options.SystemPrompt = claudecode.StringPtr("You are a helpful coding tutor. Keep track of what we've discussed.")

	client := claudecode.NewClaudeSDKClient(options)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Start message processor
	done := make(chan bool)
	go func() {
		processMessagesWithDone(client, done)
	}()

	// Conversation flow
	queries := []string{
		"Let's build a simple web server in Go",
		"Great! Now how do I add middleware?",
		"Can you show me how to add logging middleware?",
		"Perfect! Now let's add authentication",
	}

	sessionID := "tutorial"
	for i, query := range queries {
		fmt.Printf("\n[Turn %d] You: %s\n", i+1, query)
		if err := client.Query(ctx, query, sessionID); err != nil {
			log.Printf("Failed to send query: %v\n", err)
			continue
		}
		time.Sleep(4 * time.Second) // Wait for response
	}

	close(done)
	fmt.Println("\n---Multi-turn Tutorial Complete---\n")
}

func example3StatefulCoding() {
	fmt.Println("Example 3: Stateful Coding Session")
	fmt.Println("----------------------------------")

	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Read", "Write", "Edit", "Bash"}
	options.SystemPrompt = claudecode.StringPtr("You are a pair programming partner. Help build a project step by step.")

	client := claudecode.NewClaudeSDKClient(options)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Create result channel for tracking
	results := make(chan string, 100)
	go processMessagesWithResults(client, results)

	// Step-by-step project building
	projectSteps := []struct {
		query string
		wait  time.Duration
	}{
		{"Create a new Go module called 'todoapp'", 3 * time.Second},
		{"Create a main.go file with a basic structure", 3 * time.Second},
		{"Add a Todo struct with ID, Title, and Completed fields", 3 * time.Second},
		{"Create a slice to store todos and add functions to add and list todos", 4 * time.Second},
		{"Add a function to mark a todo as completed", 3 * time.Second},
		{"Create a simple CLI interface to interact with the todos", 5 * time.Second},
	}

	sessionID := "project-build"
	for i, step := range projectSteps {
		fmt.Printf("\n🔨 Step %d: %s\n", i+1, step.query)
		if err := client.Query(ctx, step.query, sessionID); err != nil {
			log.Printf("Failed to send query: %v\n", err)
			continue
		}
		time.Sleep(step.wait)

		// Check results
		select {
		case result := <-results:
			fmt.Printf("✅ Result: %s\n", result)
		default:
		}
	}

	fmt.Println("\n---Stateful Project Building Complete---\n")
}

func example4ContextAware() {
	fmt.Println("Example 4: Context-Aware Interactive Session")
	fmt.Println("--------------------------------------------")

	// This example shows how the client maintains context
	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Interactive REPL-like interface
	scanner := bufio.NewScanner(os.Stdin)
	sessionID := "interactive"

	// Start async message processor
	go func() {
		for msg := range client.ReceiveMessages() {
			handleInteractiveMessage(msg)
		}
	}()

	fmt.Println("Interactive session started. Type 'exit' to quit.")
	fmt.Println("Try asking about something, then refer back to it!\n")

	// Demo conversation
	demoQueries := []string{
		"My favorite programming language is Go",
		"What did I just tell you my favorite language is?",
		"Can you give me a tip about that language?",
		"exit",
	}

	for _, query := range demoQueries {
		fmt.Printf("\nYou: %s\n", query)

		if query == "exit" {
			break
		}

		if err := client.Query(ctx, query, sessionID); err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		// Wait for response
		time.Sleep(3 * time.Second)
	}

	fmt.Println("\n---Context-Aware Session Complete---")
}

// Helper functions

func processMessages(client *claudecode.ClaudeSDKClient) {
	for msg := range client.ReceiveMessages() {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			fmt.Print("\nClaude: ")
			for _, block := range m.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					fmt.Print(textBlock.Text)
				}
			}
			fmt.Println()
		case *claudecode.ResultMessage:
			fmt.Printf("\n[Session stats: %dms, turns: %d]\n", m.DurationMs, m.NumTurns)
		}
	}
}

func processMessagesWithDone(client *claudecode.ClaudeSDKClient, done chan bool) {
	for {
		select {
		case msg := <-client.ReceiveMessages():
			handleMessage(msg)
		case <-done:
			return
		}
	}
}

func processMessagesWithResults(client *claudecode.ClaudeSDKClient, results chan string) {
	for msg := range client.ReceiveMessages() {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range m.Content {
				switch b := block.(type) {
				case claudecode.TextBlock:
					fmt.Printf("💬 %s\n", b.Text)
				case claudecode.ToolUseBlock:
					fmt.Printf("🔧 Using tool: %s\n", b.Name)
				}
			}
		case *claudecode.SystemMessage:
			if m.Subtype == "tool_result" {
				if content, ok := m.Data["content"].(string); ok {
					results <- content
				}
			}
		}
	}
}

func handleMessage(msg types.Message) {
	switch m := msg.(type) {
	case *claudecode.AssistantMessage:
		fmt.Print("\nClaude: ")
		for _, block := range m.Content {
			if textBlock, ok := block.(claudecode.TextBlock); ok {
				fmt.Print(textBlock.Text)
			}
		}
		fmt.Println()
	}
}

func handleInteractiveMessage(msg types.Message) {
	switch m := msg.(type) {
	case *claudecode.AssistantMessage:
		fmt.Print("Claude: ")
		for _, block := range m.Content {
			if textBlock, ok := block.(claudecode.TextBlock); ok {
				// Format for better readability
				lines := strings.Split(textBlock.Text, "\n")
				for i, line := range lines {
					if i > 0 {
						fmt.Print("        ") // Indent continuation
					}
					fmt.Println(line)
				}
			}
		}
	case *claudecode.SystemMessage:
		if m.Subtype == "error" {
			fmt.Printf("❌ Error: %v\n", m.Data["error"])
		}
	}
}
