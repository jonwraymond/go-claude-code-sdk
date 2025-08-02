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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test subscription authentication with a simple configuration
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: ".",
		AuthMethod:      types.AuthTypeSubscription,
		Model:           "claude-3-5-sonnet-20241022",
		Timeout:         15 * time.Second,
		Debug:           true,
	}

	fmt.Printf("Creating client with auth method: %s\n", config.AuthMethod)
	
	client, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	fmt.Printf("Client created successfully, session ID: %s\n", config.SessionID)
	fmt.Println("Testing a simple query...")

	response, err := client.Query(ctx, &types.QueryRequest{
		Messages: []types.Message{
			{Role: types.RoleUser, Content: "Just say 'Hello from subscription auth!' and nothing else."},
		},
	})

	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Printf("Success! Response: %s\n", extractTextContent(response.Content))
}

func extractTextContent(content []types.ContentBlock) string {
	for _, block := range content {
		if block.Type == "text" {
			return block.Text
		}
	}
	return "(no text content)"
}