// MCP (Model Context Protocol) server integration example
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
	// Setup client
	config := types.NewClaudeCodeConfig()
	config.APIKey = os.Getenv("ANTHROPIC_API_KEY")

	claudeClient, err := client.NewClaudeCodeClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()

	// Example 1: Basic MCP server setup
	fmt.Println("=== Example 1: Basic MCP Server Setup ===")
	basicMCPSetup(claudeClient)

	// Example 2: Multiple MCP servers
	fmt.Println("\n=== Example 2: Multiple MCP Servers ===")
	multipleMCPServers(claudeClient)

	// Example 3: Common MCP server patterns
	fmt.Println("\n=== Example 3: Common MCP Patterns ===")
	commonMCPPatterns(claudeClient)
}

func basicMCPSetup(client *client.ClaudeCodeClient) {
	ctx := context.Background()
	mcpManager := client.MCP()

	// Configure a SQLite MCP server
	sqliteConfig := &types.MCPServerConfig{
		Name:        "sqlite-db",
		Command:     "sqlite-mcp-server",
		Args:        []string{"./example.db"},
		Description: "SQLite database for example data",
		Environment: map[string]string{
			"SQLITE_READONLY": "false",
		},
	}

	// Register the server
	err := mcpManager.RegisterServer("sqlite-db", sqliteConfig)
	if err != nil {
		log.Printf("Failed to register SQLite server: %v", err)
		return
	}
	fmt.Println("✓ Registered SQLite MCP server")

	// Start the server
	err = mcpManager.StartServer(ctx, "sqlite-db")
	if err != nil {
		log.Printf("Failed to start SQLite server: %v", err)
		return
	}
	fmt.Println("✓ Started SQLite MCP server")
	defer mcpManager.StopServer(ctx, "sqlite-db")

	// List available MCP tools
	fmt.Println("\nAvailable MCP tools from sqlite-db:")
	tools := client.Tools().GetMCPTools("sqlite-db")
	for name, tool := range tools {
		fmt.Printf("- %s: %s\n", name, tool.Description)
	}

	// Use the MCP server in a query
	options := &client.QueryOptions{
		SystemPrompt: "You have access to a SQLite database. Use it to answer questions.",
	}

	fmt.Println("\nQuerying with MCP server available:")
	messages, err := client.QueryMessages(ctx,
		"Create a users table with id, name, and email fields, then insert some sample data",
		options)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	for msg := range messages {
		if msg.Role == types.MessageRoleAssistant {
			fmt.Printf("Claude: %s\n", msg.GetText())
		}
		if msg.HasToolUse() {
			for _, tool := range msg.GetToolUses() {
				fmt.Printf("🔧 Used MCP tool: %s\n", tool.Name)
			}
		}
	}
}

func multipleMCPServers(client *client.ClaudeCodeClient) {
	ctx := context.Background()
	mcpManager := client.MCP()

	// Configure multiple MCP servers
	servers := map[string]*types.MCPServerConfig{
		"filesystem": {
			Name:        "fs-server",
			Command:     "fs-mcp-server",
			Args:        []string{"--root", "./data"},
			Description: "File system access for data directory",
		},
		"web-search": {
			Name:        "search-server",
			Command:     "web-search-mcp-server",
			Args:        []string{},
			Description: "Web search capabilities",
			Environment: map[string]string{
				"SEARCH_API_KEY": os.Getenv("SEARCH_API_KEY"),
			},
		},
		"git": {
			Name:        "git-server",
			Command:     "git-mcp-server",
			Args:        []string{"--repo", "."},
			Description: "Git repository operations",
		},
	}

	// Register all servers
	for name, config := range servers {
		err := mcpManager.RegisterServer(name, config)
		if err != nil {
			log.Printf("Failed to register %s: %v", name, err)
			continue
		}
		fmt.Printf("✓ Registered %s MCP server\n", name)
	}

	// Start all servers
	for name := range servers {
		err := mcpManager.StartServer(ctx, name)
		if err != nil {
			log.Printf("Failed to start %s: %v", name, err)
			continue
		}
		fmt.Printf("✓ Started %s MCP server\n", name)
		defer mcpManager.StopServer(ctx, name)
	}

	// List all active servers
	fmt.Println("\nActive MCP servers:")
	activeServers := mcpManager.ListServers()
	for _, server := range activeServers {
		status := "stopped"
		if mcpManager.IsServerRunning(server) {
			status = "running"
		}
		fmt.Printf("- %s: %s\n", server, status)
	}

	// Use multiple servers in a complex query
	options := &client.QueryOptions{
		SystemPrompt: "You have access to filesystem, web search, and git operations. Use them as needed.",
		MaxTurns:     10,
	}

	fmt.Println("\nComplex query using multiple MCP servers:")
	result, err := client.QueryMessagesSync(ctx,
		"Search for best practices for Go error handling, save them to a file, and create a git commit",
		options)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	// Track which servers were used
	serversUsed := make(map[string]bool)
	for _, msg := range result.Messages {
		if msg.HasToolUse() {
			for _, tool := range msg.GetToolUses() {
				// Determine which server the tool belongs to
				for server, tools := range client.Tools().GetAllMCPTools() {
					if _, exists := tools[tool.Name]; exists {
						serversUsed[server] = true
					}
				}
			}
		}
	}

	fmt.Println("\nMCP servers used:")
	for server := range serversUsed {
		fmt.Printf("- %s\n", server)
	}
}

func commonMCPPatterns(client *client.ClaudeCodeClient) {
	ctx := context.Background()
	mcpManager := client.MCP()

	// Pattern 1: Use common preset configurations
	fmt.Println("\nUsing preset MCP configurations:")
	
	presets := mcpManager.GetCommonPresets()
	for name, preset := range presets {
		fmt.Printf("- %s: %s\n", name, preset.Description)
	}

	// Use a preset
	err := mcpManager.RegisterFromPreset("sqlite")
	if err != nil {
		log.Printf("Failed to register SQLite preset: %v", err)
	} else {
		fmt.Println("✓ Registered SQLite from preset")
	}

	// Pattern 2: Configure MCP server with retries and health checks
	robustConfig := &types.MCPServerConfig{
		Name:    "robust-server",
		Command: "example-mcp-server",
		Args:    []string{},
		Environment: map[string]string{
			"LOG_LEVEL": "debug",
		},
		StartTimeout: 30 * time.Second,
		MaxRetries:   3,
	}

	err = mcpManager.RegisterServer("robust", robustConfig)
	if err != nil {
		log.Printf("Failed to register robust server: %v", err)
		return
	}

	// Pattern 3: Apply configuration to Claude Code
	fmt.Println("\nApplying MCP configuration to Claude Code...")
	err = mcpManager.ApplyConfiguration(ctx)
	if err != nil {
		log.Printf("Failed to apply configuration: %v", err)
		return
	}
	fmt.Println("✓ Configuration applied to Claude Code")

	// Pattern 4: Monitor server health
	fmt.Println("\nMonitoring MCP server health:")
	servers := mcpManager.ListServers()
	for _, server := range servers {
		if mcpManager.IsServerRunning(server) {
			// In a real application, you might check server health endpoints
			fmt.Printf("- %s: healthy ✓\n", server)
		} else {
			fmt.Printf("- %s: not running ✗\n", server)
		}
	}

	// Pattern 5: Graceful shutdown
	fmt.Println("\nPerforming graceful shutdown...")
	for _, server := range servers {
		if mcpManager.IsServerRunning(server) {
			err := mcpManager.StopServer(ctx, server)
			if err != nil {
				log.Printf("Failed to stop %s: %v", server, err)
			} else {
				fmt.Printf("✓ Stopped %s\n", server)
			}
		}
	}
}

// Example of custom MCP server implementation wrapper
type CustomMCPServer struct {
	manager *client.MCPManager
	name    string
	config  *types.MCPServerConfig
}

func NewCustomMCPServer(manager *client.MCPManager, name string) *CustomMCPServer {
	return &CustomMCPServer{
		manager: manager,
		name:    name,
		config: &types.MCPServerConfig{
			Name:        name,
			Command:     "custom-mcp-server",
			Args:        []string{"--port", "8080"},
			Description: "Custom MCP server with extended functionality",
		},
	}
}

func (s *CustomMCPServer) Start(ctx context.Context) error {
	// Register server
	if err := s.manager.RegisterServer(s.name, s.config); err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	// Start with retries
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		err := s.manager.StartServer(ctx, s.name)
		if err == nil {
			return nil
		}
		if i < maxRetries-1 {
			log.Printf("Start attempt %d failed, retrying: %v", i+1, err)
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}

	return fmt.Errorf("failed to start after %d attempts", maxRetries)
}

func (s *CustomMCPServer) Stop(ctx context.Context) error {
	return s.manager.StopServer(ctx, s.name)
}

func (s *CustomMCPServer) IsHealthy() bool {
	// Custom health check logic
	return s.manager.IsServerRunning(s.name)
}