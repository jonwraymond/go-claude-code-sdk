// Package main demonstrates session lifecycle management with Claude Code.
// This example shows how to create, manage, and persist conversations across sessions.
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
	fmt.Println("=== Claude Code Session Lifecycle Examples ===")

	// Example 1: Basic session creation and management
	basicSessionExample()

	// Example 2: Session persistence across client instances
	sessionPersistenceExample()

	// Example 3: Multiple concurrent sessions
	multipleConcurrentSessionsExample()

	// Example 4: Session with custom configuration
	customSessionConfigurationExample()

	// Example 5: Session cleanup and resource management
	sessionCleanupExample()
}

// basicSessionExample demonstrates basic session creation and usage
func basicSessionExample() {
	fmt.Println("--- Example 1: Basic Session Management ---")

	ctx := context.Background()
	config := types.NewClaudeCodeConfig()

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

	// Generate a new session ID
	sessionID := claudeClient.GenerateSessionID()
	fmt.Printf("Generated session ID: %s\n", sessionID)

	// Create a new session
	session, err := claudeClient.CreateSession(ctx, sessionID)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		return
	}

	fmt.Printf("✓ Session created successfully\n")
	fmt.Printf("  Session ID: %s\n", session.ID)

	// Use the session for a conversation
	options := &client.QueryOptions{
		SessionID:      session.ID,
		MaxTurns:       5,
		PermissionMode: client.PermissionModeAsk,
	}

	// Send first message
	fmt.Printf("\nStarting conversation in session...\n")
	result1, err := claudeClient.QueryMessagesSync(ctx, "Hello! Can you remember this conversation?", options)
	if err != nil {
		log.Printf("Failed to send first message: %v", err)
		return
	}

	if len(result1.Messages) > 1 {
		lastMessage := result1.Messages[len(result1.Messages)-1]
		fmt.Printf("Claude: %s\n", truncateText(formatContent(lastMessage.Content), 150))
	}

	// Send follow-up message to test session persistence
	result2, err := claudeClient.QueryMessagesSync(ctx, "What did I just ask you about?", options)
	if err != nil {
		log.Printf("Failed to send follow-up message: %v", err)
		return
	}

	if len(result2.Messages) > 0 {
		lastMessage := result2.Messages[len(result2.Messages)-1]
		fmt.Printf("Claude: %s\n", truncateText(formatContent(lastMessage.Content), 150))
	}

	// List all sessions
	sessions := claudeClient.ListSessions()
	fmt.Printf("\nActive sessions: %d\n", len(sessions))
	for i, sessionID := range sessions {
		fmt.Printf("  %d. %s\n", i+1, sessionID)
	}

	fmt.Println()
}

// sessionPersistenceExample demonstrates session persistence across client instances
func sessionPersistenceExample() {
	fmt.Println("--- Example 2: Session Persistence ---")

	ctx := context.Background()

	// Create first client instance
	config1 := types.NewClaudeCodeConfig()
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config1.APIKey = apiKey
		config1.AuthMethod = types.AuthTypeAPIKey
	} else {
		config1.AuthMethod = types.AuthTypeSubscription
	}

	client1, err := client.NewClaudeCodeClient(ctx, config1)
	if err != nil {
		log.Printf("Failed to create first client: %v", err)
		return
	}

	// Create session with first client
	sessionID := client1.GenerateSessionID()
	session, err := client1.CreateSession(ctx, sessionID)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		_ = client1.Close() // Ignore error during cleanup
		return
	}

	fmt.Printf("Created session with first client: %s\n", session.ID)

	// Start conversation with first client
	options := &client.QueryOptions{
		SessionID:      session.ID,
		PermissionMode: client.PermissionModeAcceptEdits,
	}

	result1, err := client1.QueryMessagesSync(ctx, "My favorite programming language is Go. Please remember this.", options)
	if err != nil {
		log.Printf("Failed to send message with first client: %v", err)
		_ = client1.Close() // Ignore error during cleanup
		return
	}

	if len(result1.Messages) > 1 {
		lastMessage := result1.Messages[len(result1.Messages)-1]
		fmt.Printf("Client 1 - Claude: %s\n", truncateText(formatContent(lastMessage.Content), 100))
	}

	// Close first client
	_ = client1.Close() // Ignore error during cleanup
	fmt.Printf("✓ First client closed\n")

	// Create second client instance with same session ID
	config2 := types.NewClaudeCodeConfig()
	config2.SessionID = sessionID // Use the same session ID
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config2.APIKey = apiKey
		config2.AuthMethod = types.AuthTypeAPIKey
	} else {
		config2.AuthMethod = types.AuthTypeSubscription
	}

	client2, err := client.NewClaudeCodeClient(ctx, config2)
	if err != nil {
		log.Printf("Failed to create second client: %v", err)
		return
	}
	defer client2.Close()

	fmt.Printf("Created second client with same session ID\n")

	// Retrieve the existing session
	retrievedSession, err := client2.GetSession(sessionID)
	if err != nil {
		log.Printf("Failed to retrieve session: %v", err)
		return
	}

	fmt.Printf("✓ Retrieved session: %s\n", retrievedSession.ID)

	// Continue conversation with second client
	result2, err := client2.QueryMessagesSync(ctx, "What programming language did I say was my favorite?", options)
	if err != nil {
		log.Printf("Failed to send message with second client: %v", err)
		return
	}

	if len(result2.Messages) > 0 {
		lastMessage := result2.Messages[len(result2.Messages)-1]
		fmt.Printf("Client 2 - Claude: %s\n", truncateText(formatContent(lastMessage.Content), 150))
	}

	fmt.Printf("✓ Session persistence verified\n")
	fmt.Println()
}

// multipleConcurrentSessionsExample demonstrates managing multiple sessions
func multipleConcurrentSessionsExample() {
	fmt.Println("--- Example 3: Multiple Concurrent Sessions ---")

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

	// Create multiple sessions for different topics
	sessions := make(map[string]*client.ClaudeCodeSession)
	topics := []string{"golang", "databases", "algorithms"}

	for _, topic := range topics {
		sessionID := claudeClient.GenerateSessionID()
		session, err := claudeClient.CreateSession(ctx, sessionID)
		if err != nil {
			log.Printf("Failed to create session for %s: %v", topic, err)
			continue
		}
		sessions[topic] = session
		fmt.Printf("✓ Created session for %s: %s\n", topic, session.ID[:8]+"...")
	}

	// Have conversations in each session
	for topic, session := range sessions {
		fmt.Printf("\nConversation in %s session:\n", topic)

		options := &client.QueryOptions{
			SessionID:      session.ID,
			PermissionMode: client.PermissionModeAsk,
		}

		var prompt string
		switch topic {
		case "golang":
			prompt = "Explain Go's interface system briefly."
		case "databases":
			prompt = "What's the difference between SQL and NoSQL databases?"
		case "algorithms":
			prompt = "Explain the Big O notation concept."
		}

		result, err := claudeClient.QueryMessagesSync(ctx, prompt, options)
		if err != nil {
			log.Printf("Failed to query in %s session: %v", topic, err)
			continue
		}

		if len(result.Messages) > 1 {
			lastMessage := result.Messages[len(result.Messages)-1]
			response := formatContent(lastMessage.Content)
			fmt.Printf("  Q: %s\n", prompt)
			fmt.Printf("  A: %s\n", truncateText(response, 200))
		}
	}

	// Show session isolation - ask about topics in different sessions
	fmt.Printf("\nTesting session isolation:\n")

	// Ask about Go in the databases session
	dbSession := sessions["databases"]
	options := &client.QueryOptions{
		SessionID:      dbSession.ID,
		PermissionMode: client.PermissionModeAsk,
	}

	result, err := claudeClient.QueryMessagesSync(ctx, "What did we discuss about Go interfaces?", options)
	if err != nil {
		log.Printf("Failed to test session isolation: %v", err)
	} else if len(result.Messages) > 0 {
		lastMessage := result.Messages[len(result.Messages)-1]
		response := formatContent(lastMessage.Content)
		fmt.Printf("  Database session response to Go question: %s\n", truncateText(response, 150))
	}

	fmt.Printf("✓ Multiple concurrent sessions demonstrated\n")
	fmt.Println()
}

// customSessionConfigurationExample demonstrates session with custom configuration
func customSessionConfigurationExample() {
	fmt.Println("--- Example 4: Custom Session Configuration ---")

	ctx := context.Background()
	config := types.NewClaudeCodeConfig()

	// Customize session configuration
	config.Model = types.ModelClaude35Sonnet
	config.SessionID = fmt.Sprintf("custom-session-%d", time.Now().Unix())
	config.Debug = true

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

	fmt.Printf("✓ Client created with custom session configuration\n")
	fmt.Printf("  Session ID: %s\n", config.SessionID)
	fmt.Printf("  Model: %s\n", config.Model)
	fmt.Printf("  Debug: %t\n", config.Debug)

	// The session is automatically created with the client configuration
	session, err := claudeClient.GetSession(config.SessionID)
	if err != nil {
		// Session doesn't exist yet, create it
		session, err = claudeClient.CreateSession(ctx, config.SessionID)
		if err != nil {
			log.Printf("Failed to create session: %v", err)
			return
		}
	}

	// Use the session with specific options
	options := &client.QueryOptions{
		SessionID:      session.ID,
		Model:          types.ModelClaude35Sonnet,
		SystemPrompt:   "You are a helpful programming assistant specializing in Go.",
		MaxTurns:       3,
		PermissionMode: client.PermissionModeAcceptEdits,
	}

	fmt.Printf("\nUsing custom session configuration...\n")

	result, err := claudeClient.QueryMessagesSync(ctx, "Write a simple Go function to calculate factorial", options)
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		return
	}

	fmt.Printf("✓ Query executed with custom configuration\n")
	if len(result.Messages) > 1 {
		lastMessage := result.Messages[len(result.Messages)-1]
		response := formatContent(lastMessage.Content)
		fmt.Printf("Response: %s\n", truncateText(response, 300))
	}

	fmt.Println()
}

// sessionCleanupExample demonstrates proper session cleanup
func sessionCleanupExample() {
	fmt.Println("--- Example 5: Session Cleanup and Resource Management ---")

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

	// Create several temporary sessions
	var tempSessions []*client.ClaudeCodeSession

	for i := 0; i < 3; i++ {
		sessionID := claudeClient.GenerateSessionID()
		session, err := claudeClient.CreateSession(ctx, sessionID)
		if err != nil {
			log.Printf("Failed to create temporary session %d: %v", i+1, err)
			continue
		}
		tempSessions = append(tempSessions, session)
		fmt.Printf("✓ Created temporary session %d: %s\n", i+1, session.ID[:8]+"...")

		// Use each session briefly
		options := &client.QueryOptions{
			SessionID:      session.ID,
			PermissionMode: client.PermissionModeAsk,
		}

		_, err = claudeClient.QueryMessagesSync(ctx, fmt.Sprintf("Hello from session %d", i+1), options)
		if err != nil {
			log.Printf("Failed to use temporary session %d: %v", i+1, err)
		}
	}

	// List all sessions before cleanup
	allSessions := claudeClient.ListSessions()
	fmt.Printf("\nSessions before cleanup: %d\n", len(allSessions))

	// Clean up sessions explicitly
	fmt.Printf("Cleaning up temporary sessions...\n")
	for i, session := range tempSessions {
		err := session.Close()
		if err != nil {
			log.Printf("Error closing session %d: %v", i+1, err)
		} else {
			fmt.Printf("✓ Closed temporary session %d\n", i+1)
		}
	}

	// List sessions after cleanup
	remainingSessions := claudeClient.ListSessions()
	fmt.Printf("Sessions after cleanup: %d\n", len(remainingSessions))

	// Demonstrate session resource monitoring
	fmt.Printf("\nSession resource information:\n")
	fmt.Printf("  Active sessions: %d\n", len(remainingSessions))

	// Client cleanup happens automatically when defer claudeClient.Close() is called
	fmt.Printf("✓ Session cleanup example completed\n")
	fmt.Printf("  (Client will be cleaned up automatically)\n")

	fmt.Println()
}

// Helper functions

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

// truncateText truncates text to a maximum length
func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}
