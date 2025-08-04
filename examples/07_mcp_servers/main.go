package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/claudecode"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== MCP Server Examples ===")

	// Example 1: Basic MCP server setup
	example1BasicMCPSetup()

	// Example 2: Multiple MCP servers
	example2MultipleMCPServers()

	// Example 3: MCP with environment variables
	example3MCPWithEnv()

	// Example 4: MCP tool filtering
	example4MCPToolFiltering()

	// Example 5: Custom MCP servers
	example5CustomMCPServers()
}

func example1BasicMCPSetup() {
	fmt.Println("Example 1: Basic MCP Server Setup")
	fmt.Println("---------------------------------")

	options := claudecode.NewClaudeCodeOptions()

	// Configure a filesystem MCP server
	options.MCPServers = map[string]types.McpServerConfig{
		"filesystem": {
			"command": "npx",
			"args":    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
			"env": map[string]string{
				"MCP_READ_ONLY": "true",
			},
		},
	}

	// Allow all MCP tools from this server
	options.MCPTools = []string{"mcp_filesystem_*"}

	fmt.Println("📡 Configured MCP filesystem server")
	if fs, ok := options.MCPServers["filesystem"]; ok {
		fmt.Printf("   Command: %v %v\n", fs["command"], fs["args"])
		fmt.Printf("   Environment: %v\n", fs["env"])
	}

	ctx := context.Background()
	msgChan := claudecode.Query(ctx, "Use the MCP filesystem server to list files in /tmp", options)

	mcpToolsUsed := []string{}

	for msg := range msgChan {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range m.Content {
				switch b := block.(type) {
				case claudecode.TextBlock:
					fmt.Printf("Claude: %s\n", b.Text)
				case claudecode.ToolUseBlock:
					if strings.HasPrefix(b.Name, "mcp_") {
						mcpToolsUsed = append(mcpToolsUsed, b.Name)
						fmt.Printf("🔧 MCP Tool: %s\n", b.Name)
						fmt.Printf("   Input: %v\n", b.Input)
					}
				}
			}
		case *claudecode.SystemMessage:
			if m.Subtype == "tool_result" && strings.Contains(fmt.Sprintf("%v", m.Data), "mcp_") {
				fmt.Printf("✅ MCP tool result received\n")
			}
		case *claudecode.ResultMessage:
			fmt.Printf("\n📊 MCP tools used: %v\n", mcpToolsUsed)
		}
	}
	fmt.Println()
}

func example2MultipleMCPServers() {
	fmt.Println("Example 2: Multiple MCP Servers")
	fmt.Println("-------------------------------")

	options := claudecode.NewClaudeCodeOptions()

	// Configure multiple MCP servers
	options.MCPServers = map[string]types.McpServerConfig{
		"filesystem": {
			"command": "npx",
			"args":    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
		},
		"github": {
			"command": "npx",
			"args":    []string{"-y", "@modelcontextprotocol/server-github"},
			"env": map[string]string{
				"GITHUB_TOKEN": os.Getenv("GITHUB_TOKEN"),
			},
		},
		"postgres": {
			"command": "npx",
			"args":    []string{"-y", "@modelcontextprotocol/server-postgres"},
			"env": map[string]string{
				"POSTGRES_URL": os.Getenv("DATABASE_URL"),
			},
		},
	}

	// Enable tools from all servers
	options.MCPTools = []string{
		"mcp_filesystem_read",
		"mcp_filesystem_write",
		"mcp_github_search_repos",
		"mcp_github_get_repo",
		"mcp_postgres_query",
	}

	fmt.Println("📡 Configured multiple MCP servers:")
	for name, config := range options.MCPServers {
		if cmd, ok := config["command"]; ok {
			fmt.Printf("   - %s: %v %v\n", name, cmd, config["args"])
		}
	}
	fmt.Printf("\n🔧 Enabled MCP tools: %v\n\n", options.MCPTools)

	// Create interactive client for better demonstration
	client := claudecode.NewClaudeSDKClient(options)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v\n", err)
		return
	}

	// Track MCP usage
	mcpUsage := make(map[string]int)

	go func() {
		for msg := range client.ReceiveMessages() {
			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range m.Content {
					if toolUse, ok := block.(claudecode.ToolUseBlock); ok {
						if strings.HasPrefix(toolUse.Name, "mcp_") {
							// Extract server name from tool name
							parts := strings.Split(toolUse.Name, "_")
							if len(parts) >= 2 {
								server := parts[1]
								mcpUsage[server]++
								fmt.Printf("🔧 %s server: %s\n", server, toolUse.Name)
							}
						}
					}
				}
			}
		}
	}()

	// Test different MCP servers
	queries := []string{
		"Check if there's a file called test.txt in /tmp using the filesystem MCP",
		"Search for Go repositories on GitHub using the GitHub MCP",
		"If you have access to postgres MCP, show me how to list tables",
	}

	for i, query := range queries {
		fmt.Printf("\n📝 Query %d: %s\n", i+1, query)
		if err := client.Query(ctx, query, "mcp-multi"); err != nil {
			log.Printf("Query failed: %v\n", err)
		}
		time.Sleep(3 * time.Second)
	}

	fmt.Printf("\n📊 MCP Usage Summary:\n")
	for server, count := range mcpUsage {
		fmt.Printf("   %s: %d calls\n", server, count)
	}
	fmt.Println()
}

func example3MCPWithEnv() {
	fmt.Println("Example 3: MCP with Environment Variables")
	fmt.Println("-----------------------------------------")

	// Set up test environment variables
	testEnvVars := map[string]string{
		"MCP_TEST_VAR":    "test_value",
		"MCP_API_KEY":     "sk-test-12345",
		"MCP_DEBUG":       "true",
		"MCP_CUSTOM_PATH": "/custom/path",
	}

	// Set environment variables
	for key, value := range testEnvVars {
		_ = os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	options := claudecode.NewClaudeCodeOptions()

	// Configure MCP server with environment variables
	options.MCPServers = map[string]types.McpServerConfig{
		"custom": {
			"command": "node",
			"args":    []string{"./custom-mcp-server.js"},
			"env": map[string]string{
				"API_KEY":     "${MCP_API_KEY}",     // Will be expanded
				"DEBUG":       "${MCP_DEBUG}",       // Will be expanded
				"CUSTOM_PATH": "${MCP_CUSTOM_PATH}", // Will be expanded
				"STATIC_VAR":  "static_value",       // Static value
			},
		},
	}

	fmt.Println("📡 MCP Server Environment Configuration:")
	fmt.Println("   Environment variables set:")
	for key, value := range testEnvVars {
		fmt.Printf("     %s=%s\n", key, value)
	}
	fmt.Println("\n   MCP server will receive:")
	if customConfig, ok := options.MCPServers["custom"]; ok {
		if env, ok := customConfig["env"].(map[string]string); ok {
			for key, value := range env {
				expanded := os.ExpandEnv(value)
				fmt.Printf("     %s=%s", key, value)
				if expanded != value {
					fmt.Printf(" → %s", expanded)
				}
				fmt.Println()
			}
		}
	}

	// Test with environment-dependent behavior
	ctx := context.Background()
	msgChan := claudecode.Query(ctx, "Show me the MCP server configuration", options)

	for msg := range msgChan {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					// Claude should acknowledge the MCP configuration
					fmt.Printf("\nClaude: %s\n", textBlock.Text)
				}
			}
		}
	}
	fmt.Println()
}

func example4MCPToolFiltering() {
	fmt.Println("Example 4: MCP Tool Filtering")
	fmt.Println("-----------------------------")

	options := claudecode.NewClaudeCodeOptions()

	// Set up MCP server with many tools
	options.MCPServers = map[string]types.McpServerConfig{
		"toolkit": {
			"command": "npx",
			"args":    []string{"-y", "@modelcontextprotocol/server-toolkit"},
		},
	}

	// Test different filtering strategies
	filterTests := []struct {
		name    string
		filters []string
		desc    string
	}{
		{
			name:    "All tools",
			filters: []string{"mcp_toolkit_*"},
			desc:    "Allow all tools from toolkit server",
		},
		{
			name:    "Specific tools",
			filters: []string{"mcp_toolkit_read", "mcp_toolkit_write"},
			desc:    "Only allow read and write tools",
		},
		{
			name:    "Pattern matching",
			filters: []string{"mcp_toolkit_*_file", "mcp_toolkit_list_*"},
			desc:    "Allow tools matching patterns",
		},
		{
			name:    "No MCP tools",
			filters: []string{},
			desc:    "MCP server configured but no tools allowed",
		},
	}

	for _, test := range filterTests {
		fmt.Printf("\n🔍 Test: %s\n", test.name)
		fmt.Printf("   Description: %s\n", test.desc)
		fmt.Printf("   Filters: %v\n", test.filters)

		testOptions := claudecode.NewClaudeCodeOptions()
		testOptions.MCPServers = options.MCPServers
		testOptions.MCPTools = test.filters

		ctx := context.Background()
		msgChan := claudecode.Query(ctx, "Try to use MCP toolkit tools", testOptions)

		toolsUsed := []string{}
		toolsBlocked := 0

		for msg := range msgChan {
			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range m.Content {
					switch b := block.(type) {
					case claudecode.ToolUseBlock:
						if strings.HasPrefix(b.Name, "mcp_") {
							toolsUsed = append(toolsUsed, b.Name)
						}
					case claudecode.TextBlock:
						if strings.Contains(strings.ToLower(b.Text), "not available") ||
							strings.Contains(strings.ToLower(b.Text), "not allowed") {
							toolsBlocked++
						}
					}
				}
			}
		}

		fmt.Printf("   Results: %d tools used, %d tools blocked\n", len(toolsUsed), toolsBlocked)
		if len(toolsUsed) > 0 {
			fmt.Printf("   Used: %v\n", toolsUsed)
		}
	}
	fmt.Println()
}

func example5CustomMCPServers() {
	fmt.Println("Example 5: Custom MCP Servers")
	fmt.Println("-----------------------------")

	// Example custom MCP server configurations
	customServers := []struct {
		name   string
		config types.McpServerConfig
		desc   string
	}{
		{
			name: "docker",
			config: types.McpServerConfig{
				"command": "docker",
				"args":    []string{"run", "--rm", "-i", "custom/mcp-docker:latest"},
				"env": map[string]string{
					"DOCKER_HOST": "unix:///var/run/docker.sock",
				},
			},
			desc: "MCP server running in Docker",
		},
		{
			name: "python-script",
			config: types.McpServerConfig{
				"command": "python3",
				"args":    []string{"-m", "mcp_server", "--port", "8080"},
				"env": map[string]string{
					"PYTHONPATH": "./mcp_modules",
				},
			},
			desc: "Python-based MCP server",
		},
		{
			name: "remote-server",
			config: types.McpServerConfig{
				"command": "mcp-client",
				"args":    []string{"--connect", "wss://mcp.example.com/ws"},
				"env": map[string]string{
					"MCP_AUTH_TOKEN": "${MCP_REMOTE_TOKEN}",
					"MCP_CLIENT_ID":  "claude-sdk-example",
				},
			},
			desc: "Remote MCP server via WebSocket",
		},
		{
			name: "local-binary",
			config: types.McpServerConfig{
				"command": "./bin/mcp-server",
				"args":    []string{"--config", "./config/mcp.yaml"},
				"env":     map[string]string{},
			},
			desc: "Local binary MCP server",
		},
	}

	fmt.Println("📡 Custom MCP Server Configurations:")

	for _, server := range customServers {
		fmt.Printf("🔧 %s - %s\n", server.name, server.desc)
		if cmd, ok := server.config["command"]; ok {
			fmt.Printf("   Command: %v %v\n", cmd, server.config["args"])
		}
		if envMap, ok := server.config["env"].(map[string]string); ok && len(envMap) > 0 {
			fmt.Printf("   Environment:\n")
			for key, value := range envMap {
				fmt.Printf("     %s=%s\n", key, value)
			}
		}
		fmt.Println()

		// Demonstrate usage
		options := claudecode.NewClaudeCodeOptions()
		options.MCPServers = map[string]types.McpServerConfig{
			server.name: server.config,
		}
		options.MCPTools = []string{fmt.Sprintf("mcp_%s_*", server.name)}

		// Show how Claude would interact with this server
		ctx := context.Background()
		query := fmt.Sprintf("If the %s MCP server is available, describe what tools it might provide", server.name)

		msgChan := claudecode.Query(ctx, query, options)

		fmt.Printf("   Claude's assessment:\n")
		for msg := range msgChan {
			if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
				for _, block := range assistantMsg.Content {
					if textBlock, ok := block.(claudecode.TextBlock); ok {
						// Show first 200 chars
						preview := textBlock.Text
						if len(preview) > 200 {
							preview = preview[:200] + "..."
						}
						fmt.Printf("   %s\n", preview)
						break
					}
				}
			}
		}
		fmt.Println()
	}

	// Advanced MCP configuration example
	fmt.Println("🚀 Advanced MCP Configuration Example:")
	fmt.Println("-------------------------------------")

	advancedOptions := claudecode.NewClaudeCodeOptions()

	// Complex multi-server setup
	advancedOptions.MCPServers = map[string]types.McpServerConfig{
		"orchestrator": {
			"command": "mcp-orchestrator",
			"args":    []string{"--mode", "distributed"},
			"env": map[string]string{
				"MCP_CLUSTER_NODES": "node1:8080,node2:8080,node3:8080",
				"MCP_LOAD_BALANCE":  "round-robin",
			},
		},
		"cache": {
			"command": "mcp-cache-server",
			"args":    []string{"--ttl", "3600", "--max-size", "1GB"},
		},
		"security": {
			"command": "mcp-security-server",
			"args":    []string{"--audit-log", "/var/log/mcp-audit.log"},
			"env": map[string]string{
				"MCP_SECURITY_LEVEL": "high",
				"MCP_ENCRYPT_DATA":   "true",
			},
		},
	}

	// Selective tool enabling
	advancedOptions.MCPTools = []string{
		"mcp_orchestrator_dispatch",
		"mcp_orchestrator_status",
		"mcp_cache_get",
		"mcp_cache_set",
		"mcp_security_audit",
	}

	// Also combine with regular tools
	advancedOptions.AllowedTools = []string{"Read", "Write"}

	fmt.Println("Configuration:")
	fmt.Printf("  MCP Servers: %d configured\n", len(advancedOptions.MCPServers))
	fmt.Printf("  MCP Tools: %d enabled\n", len(advancedOptions.MCPTools))
	fmt.Printf("  Regular Tools: %v\n", advancedOptions.AllowedTools)

	// Demonstrate combined usage
	ctx := context.Background()
	msgChan := claudecode.Query(ctx,
		"Demonstrate how MCP orchestrator, cache, and security servers could work together",
		advancedOptions)

	for msg := range msgChan {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					fmt.Printf("\nClaude: %s\n", textBlock.Text)
				}
			}
		}
	}
}
