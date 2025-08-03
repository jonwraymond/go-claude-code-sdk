// Package main demonstrates streaming queries with Claude Code.
// This example shows how to use the streaming API for real-time responses.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Claude Code Streaming Queries Examples ===")

	// Example 1: Basic streaming query
	basicStreamingExample()

	// Example 2: Streaming with context cancellation
	streamingWithCancellationExample()

	// Example 3: Processing streaming chunks
	streamingChunkProcessingExample()

	// Example 4: Using the QueryMessages API for streaming
	queryMessagesStreamingExample()
}

// basicStreamingExample demonstrates a simple streaming query
func basicStreamingExample() {
	fmt.Println("--- Example 1: Basic Streaming Query ---")

	ctx := context.Background()
	config := types.NewClaudeCodeConfig()
	
	// Use API key if available
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
		config.AuthMethod = types.AuthTypeAPIKey
	} else {
		config.AuthMethod = types.AuthTypeSubscription
	}

	// Create client
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer claudeClient.Close()

	// Create streaming request
	request := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "Write a short poem about Go programming language.",
			},
		},
		Model: "claude-3-5-sonnet-20241022",
	}

	fmt.Printf("Starting streaming query...\n")
	
	// Execute streaming query
	stream, err := claudeClient.QueryStream(ctx, request)
	if err != nil {
		log.Printf("Failed to start streaming query: %v", err)
		return
	}
	defer stream.Close()

	fmt.Printf("Response:\n")
	var fullResponse strings.Builder

	// Process streaming chunks
	for {
		chunk, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving chunk: %v", err)
			break
		}

		if chunk.Done {
			fmt.Printf("\n✓ Stream completed\n")
			break
		}

		// Print chunk content in real-time
		fmt.Print(chunk.Content)
		fullResponse.WriteString(chunk.Content)
	}

	fmt.Printf("Full response length: %d characters\n", fullResponse.Len())
	fmt.Println()
}

// streamingWithCancellationExample demonstrates cancelling a streaming query
func streamingWithCancellationExample() {
	fmt.Println("--- Example 2: Streaming with Cancellation ---")

	config := types.NewClaudeCodeConfig()
	
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
		config.AuthMethod = types.AuthTypeAPIKey
	} else {
		config.AuthMethod = types.AuthTypeSubscription
	}

	// Create client
	claudeClient, err := client.NewClaudeCodeClient(context.Background(), config)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer claudeClient.Close()

	// Create context with timeout for cancellation demo
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	request := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "Write a very long story about software development. Make it at least 1000 words.",
			},
		},
	}

	fmt.Printf("Starting streaming query with 5-second timeout...\n")
	
	stream, err := claudeClient.QueryStream(ctx, request)
	if err != nil {
		log.Printf("Failed to start streaming query: %v", err)
		return
	}
	defer stream.Close()

	fmt.Printf("Response (will timeout after 5 seconds):\n")
	
	// Process chunks until timeout or completion
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				fmt.Printf("\n⚠ Stream cancelled due to timeout\n")
			} else {
				fmt.Printf("\nError: %v\n", err)
			}
			break
		}

		if chunk.Done {
			fmt.Printf("\n✓ Stream completed before timeout\n")
			break
		}

		fmt.Print(chunk.Content)
	}

	fmt.Println()
}

// streamingChunkProcessingExample demonstrates advanced chunk processing
func streamingChunkProcessingExample() {
	fmt.Println("--- Example 3: Advanced Chunk Processing ---")

	ctx := context.Background()
	config := types.NewClaudeCodeConfig()
	
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
		config.AuthMethod = types.AuthTypeAPIKey
	} else {
		config.AuthMethod = types.AuthTypeSubscription
	}

	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer claudeClient.Close()

	request := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "Explain the Go programming language's concurrency model with examples.",
			},
		},
	}

	fmt.Printf("Starting streaming query with chunk analysis...\n")
	
	stream, err := claudeClient.QueryStream(ctx, request)
	if err != nil {
		log.Printf("Failed to start streaming query: %v", err)
		return
	}
	defer stream.Close()

	// Track streaming statistics
	var (
		chunkCount    int
		totalBytes    int
		wordCount     int
		startTime     = time.Now()
	)

	fmt.Printf("Response:\n")
	fmt.Println("---")

	for {
		chunk, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving chunk: %v", err)
			break
		}

		if chunk.Done {
			break
		}

		// Update statistics
		chunkCount++
		totalBytes += len(chunk.Content)
		wordCount += len(strings.Fields(chunk.Content))

		// Print content
		fmt.Print(chunk.Content)
	}

	elapsed := time.Since(startTime)
	
	fmt.Println("\n---")
	fmt.Printf("✓ Streaming completed\n")
	fmt.Printf("Statistics:\n")
	fmt.Printf("  Chunks received: %d\n", chunkCount)
	fmt.Printf("  Total bytes: %d\n", totalBytes)
	fmt.Printf("  Word count: %d\n", wordCount)
	fmt.Printf("  Duration: %v\n", elapsed)
	if elapsed.Seconds() > 0 {
		fmt.Printf("  Throughput: %.2f words/second\n", float64(wordCount)/elapsed.Seconds())
	}
	fmt.Println()
}

// queryMessagesStreamingExample demonstrates the QueryMessages streaming API
func queryMessagesStreamingExample() {
	fmt.Println("--- Example 4: QueryMessages Streaming API ---")

	ctx := context.Background()
	config := types.NewClaudeCodeConfig()
	
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
		config.AuthMethod = types.AuthTypeAPIKey
	} else {
		config.AuthMethod = types.AuthTypeSubscription
	}

	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer claudeClient.Close()

	// Configure query options
	options := &client.QueryOptions{
		Model:          "claude-3-5-sonnet-20241022",
		MaxTurns:       5,
		Stream:         true,
		PermissionMode: client.PermissionModeAsk,
	}

	prompt := "Help me understand goroutines with a simple example"
	fmt.Printf("Starting QueryMessages stream for: %s\n", prompt)

	// Start streaming conversation
	messageChan, err := claudeClient.QueryMessages(ctx, prompt, options)
	if err != nil {
		log.Printf("Failed to start QueryMessages stream: %v", err)
		return
	}

	fmt.Printf("Conversation:\n")
	fmt.Println("---")

	var messageCount int
	
	// Process messages from the stream
	for message := range messageChan {
		if message == nil {
			continue
		}

		messageCount++
		
		switch message.Role {
		case types.RoleUser:
			fmt.Printf("👤 User: %s\n\n", formatContent(message.Content))
		case types.RoleAssistant:
			fmt.Printf("🤖 Claude: %s\n\n", formatContent(message.Content))
		case types.RoleSystem:
			fmt.Printf("🔧 System: %s\n\n", formatContent(message.Content))
		case types.RoleTool:
			fmt.Printf("🛠️ Tool: %s\n\n", formatContent(message.Content))
		}

		// Handle tool calls
		if len(message.ToolCalls) > 0 {
			for _, toolCall := range message.ToolCalls {
				fmt.Printf("🔧 Tool Call: %s(%s)\n", toolCall.Function.Name, toolCall.Function.Arguments)
			}
			fmt.Println()
		}
	}

	fmt.Println("---")
	fmt.Printf("✓ QueryMessages stream completed\n")
	fmt.Printf("  Total messages: %d\n", messageCount)
	fmt.Println()
}

// formatContent formats message content for display
func formatContent(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []types.ContentBlock:
		var result strings.Builder
		for _, block := range v {
			if block.Type == "text" {
				result.WriteString(block.Text)
			}
		}
		return result.String()
	default:
		return fmt.Sprintf("%v", content)
	}
}