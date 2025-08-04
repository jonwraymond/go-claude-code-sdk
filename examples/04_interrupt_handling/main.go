package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/claudecode"
)

func main() {
	fmt.Println("=== Interrupt Handling Examples ===")

	// Example 1: Basic interrupt
	example1BasicInterrupt()

	// Example 2: Interrupt with follow-up
	example2InterruptWithFollowup()

	// Example 3: Signal-based interrupt
	example3SignalInterrupt()

	// Example 4: Conditional interrupt
	example4ConditionalInterrupt()
}

func example1BasicInterrupt() {
	fmt.Println("Example 1: Basic Interrupt")
	fmt.Println("--------------------------")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Message counter
	messageCount := 0
	var mu sync.Mutex

	// Start message processor
	go func() {
		for msg := range client.ReceiveMessages() {
			mu.Lock()
			messageCount++
			mu.Unlock()

			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range m.Content {
					if textBlock, ok := block.(claudecode.TextBlock); ok {
						// Show truncated output
						content := textBlock.Text
						if len(content) > 50 {
							content = content[:50] + "..."
						}
						fmt.Printf("Claude: %s\n", content)
					}
				}
			}
		}
	}()

	// Start a long task
	fmt.Println("Starting long counting task...")
	if err := client.Query(ctx, "Count from 1 to 1000, showing each number", "interrupt-demo"); err != nil {
		log.Printf("Failed to send query: %v\n", err)
		return
	}

	// Wait 2 seconds then interrupt
	time.Sleep(2 * time.Second)

	mu.Lock()
	count := messageCount
	mu.Unlock()

	fmt.Printf("\n🛑 Interrupting after %d messages...\n", count)
	if err := client.Interrupt(); err != nil {
		log.Printf("Failed to interrupt: %v\n", err)
	}

	// Give some time to see the effect
	time.Sleep(1 * time.Second)
	fmt.Println("✅ Interrupt handled successfully")
}

func example2InterruptWithFollowup() {
	fmt.Println("Example 2: Interrupt with Follow-up")
	fmt.Println("-----------------------------------")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	done := make(chan bool)
	interrupted := false

	// Message processor
	go func() {
		for msg := range client.ReceiveMessages() {
			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range m.Content {
					if textBlock, ok := block.(claudecode.TextBlock); ok {
						if interrupted {
							fmt.Printf("🔄 New task: %s\n", textBlock.Text)
						} else {
							// Show progress
							if len(textBlock.Text) > 30 {
								fmt.Print(".")
							}
						}
					}
				}
			case *claudecode.ResultMessage:
				if interrupted {
					fmt.Printf("\n✅ Follow-up completed in %dms\n", m.DurationMs)
					done <- true
				}
			}
		}
	}()

	// Start complex task
	fmt.Println("Starting complex analysis...")
	if err := client.Query(ctx, "Analyze all prime numbers between 1 and 10000 and explain their distribution", "task1"); err != nil {
		log.Printf("Failed to send query: %v\n", err)
		return
	}

	// Interrupt after 3 seconds
	time.Sleep(3 * time.Second)
	fmt.Println("\n\n🛑 Interrupting complex task...")
	if err := client.Interrupt(); err != nil {
		log.Printf("Failed to interrupt: %v\n", err)
		return
	}

	interrupted = true
	time.Sleep(500 * time.Millisecond)

	// Send new, simpler task
	fmt.Println("📤 Sending new task...")
	if err := client.Query(ctx, "Just tell me what 2+2 equals", "task2"); err != nil {
		log.Printf("Failed to send follow-up: %v\n", err)
		return
	}

	// Wait for completion
	select {
	case <-done:
		fmt.Println("✅ Interrupt and follow-up handled successfully")
	case <-time.After(5 * time.Second):
		fmt.Println("⏱️ Timeout waiting for follow-up")
	}
	fmt.Println()
}

func example3SignalInterrupt() {
	fmt.Println("Example 3: Signal-based Interrupt (Ctrl+C)")
	fmt.Println("------------------------------------------")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	taskComplete := make(chan bool)
	var processingTask bool
	var mu sync.Mutex

	// Message processor
	go func() {
		for msg := range client.ReceiveMessages() {
			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				mu.Lock()
				processingTask = true
				mu.Unlock()

				for _, block := range m.Content {
					if _, ok := block.(claudecode.TextBlock); ok {
						// Show streaming output
						fmt.Print(".")
					}
				}
			case *claudecode.ResultMessage:
				fmt.Printf("\n✅ Task completed in %dms\n", m.DurationMs)
				taskComplete <- true
			}
		}
	}()

	// Start long task
	fmt.Println("Starting long task... Press Ctrl+C to interrupt")
	if err := client.Query(ctx, "Write a detailed 1000-word essay about artificial intelligence", "signal-task"); err != nil {
		log.Printf("Failed to send query: %v\n", err)
		return
	}

	// Wait for signal or completion
	select {
	case <-sigChan:
		mu.Lock()
		processing := processingTask
		mu.Unlock()

		if processing {
			fmt.Println("\n\n🛑 Received interrupt signal!")
			if err := client.Interrupt(); err != nil {
				log.Printf("Failed to interrupt: %v\n", err)
			} else {
				fmt.Println("✅ Task interrupted successfully")
			}
		}
	case <-taskComplete:
		fmt.Println("Task completed naturally")
	case <-time.After(30 * time.Second):
		fmt.Println("Demo timeout")
	}

	// Clean up signal handler
	signal.Stop(sigChan)
	fmt.Println()
}

func example4ConditionalInterrupt() {
	fmt.Println("Example 4: Conditional Interrupt")
	fmt.Println("--------------------------------")

	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Bash", "Write"}

	client := claudecode.NewClaudeSDKClient(options)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	toolUseCount := 0
	shouldInterrupt := false
	interrupted := make(chan bool)

	// Message processor with conditional logic
	go func() {
		for msg := range client.ReceiveMessages() {
			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range m.Content {
					switch b := block.(type) {
					case claudecode.TextBlock:
						fmt.Printf("Claude: %s\n", b.Text)
					case claudecode.ToolUseBlock:
						toolUseCount++
						fmt.Printf("🔧 Tool use #%d: %s\n", toolUseCount, b.Name)

						// Interrupt if too many tool uses
						if toolUseCount >= 3 && !shouldInterrupt {
							shouldInterrupt = true
							go func() {
								time.Sleep(500 * time.Millisecond)
								fmt.Println("\n⚠️ Too many tool uses, interrupting...")
								if err := client.Interrupt(); err != nil {
									log.Printf("Failed to interrupt: %v\n", err)
								}
								interrupted <- true
							}()
						}
					}
				}
			case *claudecode.SystemMessage:
				if m.Subtype == "tool_result" {
					fmt.Println("✅ Tool completed")
				}
			case *claudecode.ResultMessage:
				if !shouldInterrupt {
					fmt.Printf("\n✅ Task completed normally with %d tool uses\n", toolUseCount)
				}
			}
		}
	}()

	// Start task that might use many tools
	fmt.Println("Starting task with tool use limit...")
	if err := client.Query(ctx, "Create 5 different test files with different content in each", "conditional"); err != nil {
		log.Printf("Failed to send query: %v\n", err)
		return
	}

	// Wait for interrupt or timeout
	select {
	case <-interrupted:
		fmt.Println("✅ Conditional interrupt executed")

		// Send simpler task
		time.Sleep(1 * time.Second)
		fmt.Println("\n📤 Sending simpler task...")
		if err := client.Query(ctx, "Just create one file called test.txt", "simple"); err != nil {
			log.Printf("Failed to send follow-up: %v\n", err)
		}
		time.Sleep(3 * time.Second)
	case <-time.After(15 * time.Second):
		fmt.Println("✅ Task completed without hitting interrupt condition")
	}
}
