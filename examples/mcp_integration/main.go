// Package main demonstrates MCP (Model Context Protocol) server integration
// with the Claude Code SDK.
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
	fmt.Println("=== MCP Server Integration Example ===")

	ctx := context.Background()

	// Create configuration
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
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()

	// Example 1: List existing MCP servers
	fmt.Println("--- Example 1: List MCP Servers ---")
	listMCPServers(claudeClient)

	// Example 2: Add a filesystem MCP server
	fmt.Println("\n--- Example 2: Add Filesystem MCP Server ---")
	addFilesystemMCPServer(ctx, claudeClient)

	// Example 3: Setup common MCP servers
	fmt.Println("\n--- Example 3: Setup Common MCP Servers ---")
	setupCommonMCPServers(ctx, claudeClient)

	// Example 4: Use MCP in queries
	fmt.Println("\n--- Example 4: Query with MCP Context ---")
	queryWithMCP(ctx, claudeClient)
}

// listMCPServers demonstrates how to list configured MCP servers
func listMCPServers(claudeClient *client.ClaudeCodeClient) {
	servers := claudeClient.ListMCPServers()

	if len(servers) == 0 {
		fmt.Println("No MCP servers configured")
		return
	}

	fmt.Printf("Found %d MCP servers:\n", len(servers))
	for name, config := range servers {
		fmt.Printf("  - %s\n", name)
		fmt.Printf("    Command: %s\n", config.Command)
		if len(config.Args) > 0 {
			fmt.Printf("    Args: %v\n", config.Args)
		}
		if config.WorkingDirectory != "" {
			fmt.Printf("    Working Dir: %s\n", config.WorkingDirectory)
		}
		fmt.Printf("    Enabled: %t\n", config.Enabled)
	}
}

// addFilesystemMCPServer demonstrates adding a filesystem MCP server
func addFilesystemMCPServer(ctx context.Context, claudeClient *client.ClaudeCodeClient) {
	// Create MCP server configuration
	mcpConfig := &types.MCPServerConfig{
		Command: "npx",
		Args:    []string{"@modelcontextprotocol/server-filesystem", "/tmp"},
		Environment: map[string]string{
			"NODE_ENV": "production",
		},
		Enabled: true,
	}

	// Add the MCP server
	err := claudeClient.AddMCPServer(ctx, "filesystem", mcpConfig)
	if err != nil {
		fmt.Printf("Failed to add filesystem MCP server: %v\n", err)
		fmt.Println("Note: This requires the MCP server to be installed")
		fmt.Println("Install with: npm install -g @modelcontextprotocol/server-filesystem")
	} else {
		fmt.Println("Successfully added filesystem MCP server")

		// List servers again to confirm
		servers := claudeClient.ListMCPServers()
		if config, exists := servers["filesystem"]; exists {
			fmt.Printf("Verified: filesystem server added with command: %s\n", config.Command)
		}
	}
}

// setupCommonMCPServers demonstrates setting up common MCP servers
func setupCommonMCPServers(ctx context.Context, claudeClient *client.ClaudeCodeClient) {
	fmt.Println("Setting up common MCP servers...")

	err := claudeClient.SetupCommonMCPServers(ctx)
	if err != nil {
		fmt.Printf("Failed to setup common MCP servers: %v\n", err)
		fmt.Println("Note: This may require MCP servers to be installed first")
	} else {
		fmt.Println("Successfully setup common MCP servers")

		// List all servers
		servers := claudeClient.ListMCPServers()
		fmt.Printf("Total MCP servers configured: %d\n", len(servers))
	}
}

// queryWithMCP demonstrates using MCP context in queries
func queryWithMCP(ctx context.Context, claudeClient *client.ClaudeCodeClient) {
	// Check if we have any MCP servers configured
	servers := claudeClient.ListMCPServers()
	if len(servers) == 0 {
		fmt.Println("No MCP servers available for queries")
		fmt.Println("MCP servers would provide additional context for Claude's responses")
		return
	}

	// Create a query that could benefit from MCP context
	request := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "What tools and capabilities are available through the configured MCP servers?",
			},
		},
		MaxTokens: 500,
	}

	fmt.Println("Querying with MCP context enabled...")
	response, err := claudeClient.Query(ctx, request)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	// Display response
	if len(response.Content) > 0 {
		fmt.Printf("\nResponse:\n%s\n", response.Content[0].Text)
	}

	// Show token usage
	if response.Usage != nil {
		fmt.Printf("\nToken usage - Input: %d, Output: %d\n",
			response.Usage.InputTokens, response.Usage.OutputTokens)
	}
}
