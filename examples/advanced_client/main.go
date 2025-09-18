// Package main demonstrates advanced Claude Code client initialization
// including MCP servers, custom timeouts, and resource management.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Advanced Claude Code Client Examples ===")

	// Example 1: Client with timeout configuration
	timeoutConfigurationExample()

	// Example 2: Client with MCP server setup
	mcpServerExample()

	// Example 3: Client lifecycle management
	lifecycleManagementExample()

	// Example 4: Client with resource monitoring
	resourceMonitoringExample()

	// Example 5: Client with custom Claude Code path
	customClaudePathExample()
}

// timeoutConfigurationExample demonstrates timeout and context management
func timeoutConfigurationExample() {
	fmt.Println("--- Example 1: Timeout Configuration ---")

	// Create context with timeout for client operations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	config := types.NewClaudeCodeConfig()
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}

	// Create client with timeout context
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create client with timeout: %v", err)
		return
	}
	defer claudeClient.Close()

	fmt.Printf("✓ Client created with 30-second timeout\n")

	// Demonstrate query with timeout
	queryCtx, queryCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer queryCancel()

	request := &types.QueryRequest{
		Messages: []types.Message{
			{Role: types.RoleUser, Content: "What is Go programming language?"},
		},
	}

	fmt.Printf("Executing query with 10-second timeout...\n")
	response, err := claudeClient.Query(queryCtx, request)
	if err != nil {
		if queryCtx.Err() == context.DeadlineExceeded {
			fmt.Printf("⚠ Query timed out after 10 seconds\n")
		} else {
			fmt.Printf("Query failed: %v\n", err)
		}
	} else {
		fmt.Printf("✓ Query completed successfully\n")
		if len(response.Content) > 0 {
			content := response.Content[0].Text
			if len(content) > 100 {
				content = content[:100] + "..."
			}
			fmt.Printf("  Response preview: %s\n", content)
		}
	}
	fmt.Println()
}

// mcpServerExample demonstrates MCP server configuration
func mcpServerExample() {
	fmt.Println("--- Example 2: MCP Server Configuration ---")

	ctx := context.Background()
	config := types.NewClaudeCodeConfig()
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}

	// Create client
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create client for MCP example: %v", err)
		return
	}
	defer claudeClient.Close()

	fmt.Printf("✓ Client created for MCP configuration\n")

	// List current MCP servers
	servers := claudeClient.ListMCPServers()
	fmt.Printf("Current MCP servers: %d configured\n", len(servers))
	for name, config := range servers {
		fmt.Printf("  - %s: %s\n", name, config.Command)
	}

	// Add a common MCP server (filesystem)
	fmt.Printf("Adding filesystem MCP server...\n")
	err = claudeClient.AddMCPServer(ctx, "filesystem", &types.MCPServerConfig{
		Command: "npx",
		Args:    []string{"@modelcontextprotocol/server-filesystem", "/tmp"},
	})
	if err != nil {
		fmt.Printf("⚠ Failed to add filesystem MCP server: %v\n", err)
	} else {
		fmt.Printf("✓ Filesystem MCP server added\n")
	}

	// Setup common MCP servers
	fmt.Printf("Setting up common MCP servers...\n")
	err = claudeClient.SetupCommonMCPServers(ctx)
	if err != nil {
		fmt.Printf("⚠ Failed to setup common MCP servers: %v\n", err)
	} else {
		fmt.Printf("✓ Common MCP servers configured\n")
	}

	// List servers again
	servers = claudeClient.ListMCPServers()
	fmt.Printf("MCP servers after setup: %d configured\n", len(servers))

	fmt.Println()
}

// lifecycleManagementExample demonstrates proper client lifecycle management
func lifecycleManagementExample() {
	fmt.Println("--- Example 3: Client Lifecycle Management ---")

	ctx := context.Background()

	// Function to create and use a client
	useClientSession := func(sessionName string) error {
		config := types.NewClaudeCodeConfig()
		config.SessionID = fmt.Sprintf("lifecycle-session-%s-%d", sessionName, time.Now().Unix())

		if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
			config.APIKey = apiKey
		}

		// Create client
		claudeClient, err := client.NewClaudeCodeClient(ctx, config)
		if err != nil {
			return fmt.Errorf("failed to create client for session %s: %w", sessionName, err)
		}

		// Ensure cleanup happens even if function panics
		defer func() {
			if err := claudeClient.Close(); err != nil {
				log.Printf("Error closing client for session %s: %v", sessionName, err)
			} else {
				fmt.Printf("✓ Client for session %s closed successfully\n", sessionName)
			}
		}()

		fmt.Printf("✓ Client created for session: %s\n", sessionName)

		// Use the client (simulate some work)
		projectCtx, err := claudeClient.GetProjectContext(ctx)
		if err != nil {
			return fmt.Errorf("failed to get project context: %w", err)
		}

		fmt.Printf("  Working in directory: %s\n", projectCtx.WorkingDirectory)
		return nil
	}

	// Create multiple client sessions to demonstrate lifecycle
	sessions := []string{"session1", "session2", "session3"}
	for _, sessionName := range sessions {
		if err := useClientSession(sessionName); err != nil {
			log.Printf("Error in session %s: %v", sessionName, err)
		}
	}

	fmt.Println()
}

// resourceMonitoringExample demonstrates resource monitoring and management
func resourceMonitoringExample() {
	fmt.Println("--- Example 4: Resource Monitoring ---")

	ctx := context.Background()
	config := types.NewClaudeCodeConfig()
	config.Debug = true // Enable debug logging

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}

	// Create client
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create client for monitoring: %v", err)
		return
	}
	defer claudeClient.Close()

	fmt.Printf("✓ Client created with debug mode enabled\n")

	// Demonstrate project context caching
	fmt.Printf("Testing project context cache...\n")

	start := time.Now()
	ctx1, err := claudeClient.GetProjectContext(ctx)
	if err != nil {
		log.Printf("Failed to get project context: %v", err)
		return
	}
	first := time.Since(start)

	start = time.Now()
	ctx2, err := claudeClient.GetProjectContext(ctx)
	if err != nil {
		log.Printf("Failed to get project context (cached): %v", err)
		return
	}
	second := time.Since(start)

	fmt.Printf("  First call: %v\n", first)
	fmt.Printf("  Cached call: %v\n", second)
	fmt.Printf("  Same result: %t\n", ctx1.WorkingDirectory == ctx2.WorkingDirectory)

	// Demonstrate cache invalidation
	claudeClient.InvalidateProjectContextCache()
	fmt.Printf("✓ Project context cache invalidated\n")

	// Set custom cache duration
	claudeClient.SetProjectContextCacheDuration(5 * time.Minute)
	fmt.Printf("✓ Cache duration set to 5 minutes\n")

	// Get cache info
	cacheInfo := claudeClient.GetProjectContextCacheInfo()
	fmt.Printf("Cache info:\n")
	for key, value := range cacheInfo {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println()
}

// customClaudePathExample demonstrates using a custom Claude Code executable path
func customClaudePathExample() {
	fmt.Println("--- Example 5: Custom Claude Code Path ---")

	ctx := context.Background()

	// Create configuration with custom Claude Code path
	config := types.NewClaudeCodeConfig()

	// Try to find claude in common locations
	possiblePaths := []string{
		"/usr/local/bin/claude",
		"/opt/homebrew/bin/claude",
		"claude", // Default PATH lookup
	}

	var claudePath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			claudePath = path
			break
		}
	}

	if claudePath == "" {
		// Use npx as fallback (most common installation method)
		claudePath = "npx"
		config.ClaudeCodePath = claudePath
		fmt.Printf("Using npx to run Claude Code\n")
	} else {
		config.ClaudeCodePath = claudePath
		fmt.Printf("Found Claude Code at: %s\n", claudePath)
	}

	// Set test mode to avoid actually calling Claude CLI for this example
	config.TestMode = true

	// Create client with custom path
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create client with custom path: %v", err)
		return
	}
	defer claudeClient.Close()

	fmt.Printf("✓ Client created with custom Claude Code path\n")
	fmt.Printf("  Path: %s\n", config.ClaudeCodePath)
	fmt.Printf("  Test Mode: %t\n", config.TestMode)

	// In test mode, we can still test basic functionality
	projectCtx, err := claudeClient.GetProjectContext(ctx)
	if err != nil {
		log.Printf("Failed to get project context: %v", err)
		return
	}

	fmt.Printf("✓ Project context retrieved successfully\n")
	fmt.Printf("  Working Directory: %s\n", projectCtx.WorkingDirectory)

	fmt.Println()
}
