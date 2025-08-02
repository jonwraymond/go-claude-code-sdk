//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Testing SDK Common MCP Servers ===")

	ctx := context.Background()
	config := &types.ClaudeCodeConfig{
		Model: "claude-3-5-sonnet-20241022",
	}

	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()

	// Get MCP manager
	mcpManager := claudeClient.MCP()

	// Test 1: Add common MCP servers
	fmt.Println("\nTest 1: Adding Common MCP Servers...")
	err = mcpManager.AddCommonServers()
	if err != nil {
		log.Printf("❌ FAILED: Add common servers error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Common servers added")

		// List all servers
		servers := mcpManager.ListServers()
		fmt.Printf("   Total servers: %d\n", len(servers))
		for name, config := range servers {
			fmt.Printf("   - %s: %s (enabled: %v)\n", name, config.Command, config.Enabled)
		}
	}

	// Test 2: Using client helper method
	fmt.Println("\nTest 2: Using Client Helper Method...")

	// Create a new client to test the helper
	newClient, err := client.NewClaudeCodeClient(ctx, &types.ClaudeCodeConfig{
		Model: "claude-3-5-sonnet-20241022",
	})
	if err != nil {
		log.Fatalf("Failed to create new client: %v", err)
	}
	defer newClient.Close()

	err = newClient.SetupCommonMCPServers(ctx)
	if err != nil {
		log.Printf("❌ FAILED: Setup common servers error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Common servers setup via client helper")

		// List servers from the new client
		servers := newClient.ListMCPServers()
		fmt.Printf("   Total servers: %d\n", len(servers))
		for name, config := range servers {
			fmt.Printf("   - %s: %s\n", name, config.Command)
		}
	}

	// Test 3: Enable specific server
	fmt.Println("\nTest 3: Enable Specific Server...")
	err = newClient.EnableMCPServer(ctx, "filesystem")
	if err != nil {
		log.Printf("❌ FAILED: Enable filesystem server error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Filesystem server enabled")

		// Check enabled servers
		enabledServers := newClient.MCP().GetEnabledServers()
		fmt.Printf("   Enabled servers: %d\n", len(enabledServers))
		for name := range enabledServers {
			fmt.Printf("   - %s\n", name)
		}
	}

	fmt.Println("\n=== Common MCP Servers Tests Complete ===")
}
