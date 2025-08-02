// Advanced example demonstrating error handling, cancellation, and complex workflows
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	claudeerrors "github.com/jraymond/claude-code-go-sdk/pkg/errors"
	"github.com/jraymond/claude-code-go-sdk/pkg/client"
	"github.com/jraymond/claude-code-go-sdk/pkg/types"
)

func main() {
	// Setup client
	config := types.NewClaudeCodeConfig()
	config.APIKey = os.Getenv("ANTHROPIC_API_KEY")
	
	// Configure timeouts and retries
	config.Timeout = 30 * time.Second
	config.MaxRetries = 3

	claudeClient, err := client.NewClaudeCodeClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()

	// Example 1: Error handling patterns
	fmt.Println("=== Example 1: Error Handling ===")
	errorHandlingExample(claudeClient)

	// Example 2: Context cancellation
	fmt.Println("\n=== Example 2: Context Cancellation ===")
	cancellationExample(claudeClient)

	// Example 3: Concurrent operations
	fmt.Println("\n=== Example 3: Concurrent Operations ===")
	concurrentExample(claudeClient)

	// Example 4: Complex workflow
	fmt.Println("\n=== Example 4: Complex Workflow ===")
	complexWorkflowExample(claudeClient)
}

func errorHandlingExample(client *client.ClaudeCodeClient) {
	ctx := context.Background()

	// Try various operations that might fail
	operations := []struct {
		name string
		fn   func() error
	}{
		{
			name: "Invalid query",
			fn: func() error {
				_, err := client.QueryMessagesSync(ctx, "", nil)
				return err
			},
		},
		{
			name: "Tool with missing arguments",
			fn: func() error {
				_, err := client.Tools().ExecuteTool(ctx, &client.ClaudeCodeTool{
					Name:      "read_file",
					Arguments: map[string]interface{}{}, // Missing required 'path'
				})
				return err
			},
		},
		{
			name: "Session with invalid ID",
			fn: func() error {
				_, err := client.Sessions().GetSession("")
				return err
			},
		},
	}

	for _, op := range operations {
		fmt.Printf("\nTrying %s:\n", op.name)
		err := op.fn()
		if err != nil {
			handleError(err)
		}
	}
}

func handleError(err error) {
	// Type switch for specific error handling
	switch e := err.(type) {
	case *claudeerrors.ClaudeCodeError:
		fmt.Printf("Claude Code Error:\n")
		fmt.Printf("  Code: %s\n", e.Code)
		fmt.Printf("  Message: %s\n", e.Message)
		fmt.Printf("  Category: %s\n", e.Category)
		fmt.Printf("  Retryable: %v\n", e.IsRetryable())
		
	case *claudeerrors.ValidationError:
		fmt.Printf("Validation Error:\n")
		fmt.Printf("  Field: %s\n", e.Field)
		fmt.Printf("  Value: %v\n", e.Value)
		fmt.Printf("  Constraint: %s\n", e.Constraint)
		
	default:
		// Check for wrapped errors
		var claudeErr *claudeerrors.ClaudeCodeError
		if errors.As(err, &claudeErr) {
			fmt.Printf("Wrapped Claude Error: %s\n", claudeErr.Message)
		} else {
			fmt.Printf("General Error: %v\n", err)
		}
	}
}

func cancellationExample(client *client.ClaudeCodeClient) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start a long-running query
	fmt.Println("Starting long-running query (5s timeout)...")
	messages, err := client.QueryMessages(ctx, 
		"Write a detailed analysis of all Go files in this project", nil)
	if err != nil {
		fmt.Printf("Failed to start query: %v\n", err)
		return
	}

	// Process messages until timeout or completion
	messageCount := 0
	for msg := range messages {
		messageCount++
		fmt.Printf("Received message %d: [%s]\n", messageCount, msg.Role)
		
		// Check if context was cancelled
		select {
		case <-ctx.Done():
			fmt.Println("Context cancelled/timed out")
			return
		default:
			// Continue processing
		}
	}

	fmt.Printf("Completed successfully with %d messages\n", messageCount)
}

func concurrentExample(client *client.ClaudeCodeClient) {
	ctx := context.Background()
	
	// Define multiple queries to run concurrently
	queries := []string{
		"What is the main purpose of this Go SDK?",
		"List all the main features of this SDK",
		"What are the key design patterns used?",
		"How does error handling work in this SDK?",
	}

	// Use WaitGroup to wait for all queries
	var wg sync.WaitGroup
	results := make([]string, len(queries))
	errors := make([]error, len(queries))

	// Run queries concurrently
	for i, query := range queries {
		wg.Add(1)
		go func(index int, q string) {
			defer wg.Done()

			// Create a timeout context for each query
			queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			// Execute query
			result, err := client.QueryMessagesSync(queryCtx, q, &client.QueryOptions{
				MaxTurns: 1,
			})
			
			if err != nil {
				errors[index] = err
				return
			}

			// Extract response
			for _, msg := range result.Messages {
				if msg.Role == types.MessageRoleAssistant {
					results[index] = msg.GetText()
					break
				}
			}
		}(i, query)
	}

	// Wait for all queries to complete
	wg.Wait()

	// Display results
	fmt.Println("Concurrent query results:")
	for i, query := range queries {
		fmt.Printf("\nQ%d: %s\n", i+1, query)
		if errors[i] != nil {
			fmt.Printf("Error: %v\n", errors[i])
		} else {
			fmt.Printf("A: %.100s...\n", results[i]) // First 100 chars
		}
	}
}

func complexWorkflowExample(client *client.ClaudeCodeClient) {
	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, shutting down...")
		cancel()
	}()

	// Complex workflow: Code review pipeline
	fmt.Println("Starting code review workflow...")

	// Step 1: Create specialized session
	session, err := client.Sessions().CreateSession(ctx, "code-review-session")
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		return
	}
	defer session.Close()

	// Step 2: Configure for code review
	reviewOptions := &client.QueryOptions{
		SystemPrompt: `You are a senior code reviewer. Focus on:
- Code quality and best practices
- Potential bugs and edge cases
- Performance considerations
- Security vulnerabilities
- Suggestions for improvement`,
		SessionID:    session.ID,
		AllowedTools: []string{"read_file", "search_code", "list_directory"},
		MaxTurns:     20,
	}

	// Step 3: Get project context
	projectCtx, err := client.ProjectContext().GetEnhancedProjectContext(ctx)
	if err != nil {
		log.Printf("Failed to get project context: %v", err)
		return
	}

	// Step 4: Build review request based on project
	reviewRequest := fmt.Sprintf(`Please perform a comprehensive code review of this %s project.
Focus on:
1. The main package structure
2. Error handling patterns
3. Concurrency safety
4. API design
5. Test coverage

Project: %s
Language: %s
Framework: %s`,
		projectCtx.Language,
		projectCtx.ProjectName,
		projectCtx.Language,
		projectCtx.Framework)

	// Step 5: Execute review with progress tracking
	fmt.Println("\nStarting code review...")
	startTime := time.Now()
	
	messages, err := client.QueryMessages(ctx, reviewRequest, reviewOptions)
	if err != nil {
		log.Printf("Failed to start review: %v", err)
		return
	}

	// Track review progress
	var (
		messageCount int
		toolUseCount int
		filesReviewed = make(map[string]bool)
	)

	// Process review messages
	for msg := range messages {
		select {
		case <-ctx.Done():
			fmt.Println("Review cancelled")
			return
		default:
			messageCount++

			// Track tool usage
			if msg.HasToolUse() {
				for _, tool := range msg.GetToolUses() {
					toolUseCount++
					if tool.Name == "read_file" {
						if path, ok := tool.Input["path"].(string); ok {
							filesReviewed[path] = true
						}
					}
					fmt.Printf("🔧 Reviewing: %s\n", tool.Name)
				}
			}

			// Show assistant responses
			if msg.Role == types.MessageRoleAssistant && msg.GetText() != "" {
				fmt.Printf("\n📝 Review finding:\n%s\n", msg.GetText())
			}
		}
	}

	// Step 6: Summary
	duration := time.Since(startTime)
	fmt.Printf("\n=== Review Complete ===\n")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Messages processed: %d\n", messageCount)
	fmt.Printf("Tool uses: %d\n", toolUseCount)
	fmt.Printf("Files reviewed: %d\n", len(filesReviewed))
	fmt.Println("\nFiles reviewed:")
	for file := range filesReviewed {
		fmt.Printf("  - %s\n", file)
	}
}

// Example of a custom retry mechanism
func retryWithBackoff(ctx context.Context, maxRetries int, operation func() error) error {
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		// Check context before retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// Try operation
		err := operation()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// Check if error is retryable
		var claudeErr *claudeerrors.ClaudeCodeError
		if errors.As(err, &claudeErr) && !claudeErr.IsRetryable() {
			return err
		}
		
		// Calculate backoff
		if i < maxRetries-1 {
			backoff := time.Duration(i+1) * time.Second
			fmt.Printf("Retry %d/%d after %v: %v\n", i+1, maxRetries, backoff, err)
			
			select {
			case <-time.After(backoff):
				// Continue to next retry
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	
	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}