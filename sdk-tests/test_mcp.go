//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Testing SDK MCP Integration ===")

	ctx := context.Background()

	// Create a test directory for MCP configuration
	testDir, err := os.MkdirTemp("", "sdk-mcp-test-*")
	if err != nil {
		log.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	config := &types.ClaudeCodeConfig{
		WorkingDirectory: testDir,
		Model:            "claude-3-5-sonnet-20241022",
	}

	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()

	// Get MCP manager
	mcpManager := claudeClient.MCP()

	// Test 1: Add MCP server
	fmt.Println("\nTest 1: Adding MCP Server...")
	testServerConfig := &types.MCPServerConfig{
		Command:     "echo",
		Args:        []string{"test-mcp-server"},
		Environment: map[string]string{"TEST": "true"},
		Enabled:     true,
	}

	err = mcpManager.AddServer("test-server", testServerConfig)
	if err != nil {
		log.Printf("❌ FAILED: Add MCP server error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: MCP server added")
		fmt.Printf("   Server name: test-server\n")
		fmt.Printf("   Command: %s\n", testServerConfig.Command)
	}

	// Test 2: List MCP servers
	fmt.Println("\nTest 2: Listing MCP Servers...")
	servers := mcpManager.ListServers()
	fmt.Printf("✅ MCP servers count: %d\n", len(servers))
	for serverName := range servers {
		fmt.Printf("   - %s\n", serverName)
	}

	// Test 3: Get server configuration
	fmt.Println("\nTest 3: Getting Server Configuration...")
	if len(servers) > 0 {
		var firstServerName string
		for name := range servers {
			firstServerName = name
			break
		}

		serverConfig, err := mcpManager.GetServer(firstServerName)
		if err != nil {
			log.Printf("❌ FAILED: Get server error: %v", err)
		} else {
			fmt.Println("✅ SUCCESS: Server configuration retrieved")
			fmt.Printf("   Command: %s\n", serverConfig.Command)
			fmt.Printf("   Enabled: %v\n", serverConfig.Enabled)
		}
	}

	// Test 4: Enable/Disable server
	fmt.Println("\nTest 4: Enable/Disable MCP Server...")
	if len(servers) > 0 {
		var firstServerName string
		for name := range servers {
			firstServerName = name
			break
		}

		// Disable
		err = mcpManager.DisableServer(firstServerName)
		if err != nil {
			log.Printf("❌ FAILED: Disable server error: %v", err)
		} else {
			fmt.Println("✅ SUCCESS: Server disabled")

			// Check enabled servers
			enabledServers := mcpManager.GetEnabledServers()
			fmt.Printf("   Enabled servers: %d\n", len(enabledServers))
		}

		// Re-enable
		err = mcpManager.EnableServer(firstServerName)
		if err != nil {
			log.Printf("❌ FAILED: Enable server error: %v", err)
		} else {
			fmt.Println("✅ SUCCESS: Server re-enabled")
		}
	}

	// Test 5: Save and load configuration
	fmt.Println("\nTest 5: Save/Load MCP Configuration...")
	configPath := filepath.Join(testDir, "mcp-test.json")

	err = mcpManager.SaveToFile(configPath)
	if err != nil {
		log.Printf("❌ FAILED: Save configuration error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Configuration saved")
		fmt.Printf("   Path: %s\n", configPath)

		// Create new manager and load
		newMcpManager := client.NewMCPManager(claudeClient)
		err = newMcpManager.LoadFromFile(configPath)
		if err != nil {
			log.Printf("❌ FAILED: Load configuration error: %v", err)
		} else {
			fmt.Println("✅ SUCCESS: Configuration loaded")
			loadedServers := newMcpManager.ListServers()
			fmt.Printf("   Loaded servers: %d\n", len(loadedServers))
		}
	}

	// Test 6: Apply configuration to client
	fmt.Println("\nTest 6: Apply MCP Configuration...")
	err = mcpManager.ApplyConfiguration(ctx)
	if err != nil {
		log.Printf("❌ FAILED: Apply configuration error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: MCP configuration applied")

		// Check if .claude directory was created
		claudeDir := filepath.Join(testDir, ".claude")
		if _, err := os.Stat(claudeDir); err == nil {
			fmt.Printf("   .claude directory created: %s\n", claudeDir)

			// Check for mcp.json
			mcpJsonPath := filepath.Join(claudeDir, "mcp.json")
			if _, err := os.Stat(mcpJsonPath); err == nil {
				fmt.Println("   ✅ mcp.json file created")
			}
		}
	}

	// Test 7: Remove server
	fmt.Println("\nTest 7: Remove MCP Server...")
	if len(servers) > 0 {
		var firstServerName string
		for name := range servers {
			firstServerName = name
			break
		}

		err = mcpManager.RemoveServer(firstServerName)
		if err != nil {
			log.Printf("❌ FAILED: Remove server error: %v", err)
		} else {
			fmt.Println("✅ SUCCESS: Server removed")
			remainingServers := mcpManager.ListServers()
			fmt.Printf("   Remaining servers: %d\n", len(remainingServers))
		}
	}

	fmt.Println("\n=== MCP Integration Tests Complete ===")
}
