// Session management example demonstrating persistent conversations
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	// Setup client
	config := types.NewClaudeCodeConfig()
	config.APIKey = os.Getenv("ANTHROPIC_API_KEY")

	claudeClient, err := client.NewClaudeCodeClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()

	// Example 1: Basic session management
	fmt.Println("=== Example 1: Basic Session ===")
	basicSession(claudeClient)

	// Example 2: Session with project context
	fmt.Println("\n=== Example 2: Session with Project Context ===")
	sessionWithContext(claudeClient)

	// Example 3: Multiple sessions
	fmt.Println("\n=== Example 3: Multiple Sessions ===")
	multipleSessions(claudeClient)
}

func basicSession(client *client.ClaudeCodeClient) {
	ctx := context.Background()

	// Create a new session
	sessionManager := client.Sessions()
	session, err := sessionManager.CreateSession(ctx, "demo-session-1")
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		return
	}
	defer session.Close()

	fmt.Printf("Created session: %s\n", session.ID)

	// Execute multiple commands in the same session
	// The conversation context is maintained
	queries := []string{
		"My name is Alice and I'm learning Go",
		"What's my name?", // Claude will remember from previous message
		"What am I learning?", // Claude will remember this too
	}

	for _, query := range queries {
		fmt.Printf("\nYou: %s\n", query)

		result, err := session.Query(ctx, &types.QueryRequest{
			Messages: []types.Message{
				{Role: types.RoleUser, Content: query},
			},
			MaxTokens: 500,
		})
		if err != nil {
			log.Printf("Query failed: %v", err)
			continue
		}

		fmt.Printf("Claude: %s\n", result.Content)
	}
}

func sessionWithContext(client *client.ClaudeCodeClient) {
	ctx := context.Background()

	// Get project context first
	projectCtx, err := client.ProjectContext().GetEnhancedProjectContext(ctx)
	if err != nil {
		log.Printf("Failed to get project context: %v", err)
		return
	}

	// Create session with project awareness
	session, err := client.Sessions().CreateSession(ctx, "project-aware-session")
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		return
	}
	defer session.Close()

	// Build a context-aware system prompt
	systemPrompt := fmt.Sprintf(
		"You are helping with a %s project named '%s'. The main framework is %s. "+
			"Be specific to this technology stack.",
		projectCtx.Language,
		projectCtx.ProjectName,
		projectCtx.Framework,
	)

	// Use project context in conversation
	options := &client.QueryOptions{
		SystemPrompt: systemPrompt,
		SessionID:    session.ID,
	}

	messages, err := client.QueryMessages(ctx, 
		"What are the best practices for error handling in this project?", 
		options)
	if err != nil {
		log.Printf("Failed to query: %v", err)
		return
	}

	for msg := range messages {
		if msg.Role == types.MessageRoleAssistant {
			fmt.Printf("Claude (context-aware): %s\n", msg.GetText())
		}
	}

	// Show project information
	fmt.Printf("\nProject Context:\n")
	fmt.Printf("- Name: %s\n", projectCtx.ProjectName)
	fmt.Printf("- Language: %s\n", projectCtx.Language)
	fmt.Printf("- Framework: %s\n", projectCtx.Framework)
	fmt.Printf("- Working Dir: %s\n", projectCtx.WorkingDirectory)
}

func multipleSessions(client *client.ClaudeCodeClient) {
	ctx := context.Background()
	sessionManager := client.Sessions()

	// Create multiple sessions for different purposes
	sessions := map[string]string{
		"code-review":    "You are a code reviewer. Focus on finding issues and suggesting improvements.",
		"documentation":  "You are a technical writer. Help create clear, comprehensive documentation.",
		"debugging":      "You are a debugging assistant. Help identify and fix bugs.",
	}

	for purpose, systemPrompt := range sessions {
		session, err := sessionManager.CreateSession(ctx, purpose)
		if err != nil {
			log.Printf("Failed to create %s session: %v", purpose, err)
			continue
		}
		defer session.Close()

		fmt.Printf("\n=== %s Session ===\n", purpose)

		// Query each session with its specific purpose
		options := &client.QueryOptions{
			SystemPrompt: systemPrompt,
			SessionID:    session.ID,
			MaxTurns:     1,
		}

		result, err := client.QueryMessagesSync(ctx,
			"Review this Go function: func add(a, b int) int { return a + b }",
			options)
		if err != nil {
			log.Printf("Query failed: %v", err)
			continue
		}

		// Print the specialized response
		for _, msg := range result.Messages {
			if msg.Role == types.MessageRoleAssistant {
				fmt.Printf("%s response: %s\n", purpose, msg.GetText())
			}
		}
	}

	// List all active sessions
	fmt.Println("\n=== Active Sessions ===")
	activeSessions := sessionManager.ListSessions()
	for _, session := range activeSessions {
		fmt.Printf("- %s (created: %v ago)\n", 
			session.ID, 
			session.GetAge())
	}
}