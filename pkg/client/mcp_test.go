package client

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jraymond/claude-code-go-sdk/pkg/types"
)

func TestNewMCPManager(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
		MCPServers: map[string]*types.MCPServerConfig{
			"test-server": {
				Command: "echo",
				Args:    []string{"hello"},
				Enabled: true,
			},
		},
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	manager := client.MCP()
	if manager == nil {
		t.Fatal("MCP manager should not be nil")
	}

	// Check that existing servers were loaded
	servers := manager.ListServers()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(servers))
	}

	if _, exists := servers["test-server"]; !exists {
		t.Error("Expected test-server to be loaded")
	}
}

func TestMCPManager_AddServer(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	manager := client.MCP()

	// Test adding a server
	serverConfig := &types.MCPServerConfig{
		Command: "npx",
		Args:    []string{"@modelcontextprotocol/server-filesystem", tempDir},
		Enabled: true,
	}

	err = manager.AddServer("filesystem", serverConfig)
	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}

	// Verify server was added
	servers := manager.ListServers()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(servers))
	}

	server, exists := servers["filesystem"]
	if !exists {
		t.Fatal("Expected filesystem server to exist")
	}

	if server.Command != "npx" {
		t.Errorf("Expected command 'npx', got '%s'", server.Command)
	}

	if !server.Enabled {
		t.Error("Expected server to be enabled")
	}
}

func TestMCPManager_AddServerValidation(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	manager := client.MCP()

	tests := []struct {
		name        string
		serverName  string
		config      *types.MCPServerConfig
		expectError bool
	}{
		{
			name:        "empty server name",
			serverName:  "",
			config:      &types.MCPServerConfig{Command: "echo"},
			expectError: true,
		},
		{
			name:        "nil config",
			serverName:  "test",
			config:      nil,
			expectError: true,
		},
		{
			name:        "empty command",
			serverName:  "test",
			config:      &types.MCPServerConfig{Command: ""},
			expectError: true,
		},
		{
			name:        "invalid working directory",
			serverName:  "test",
			config:      &types.MCPServerConfig{Command: "echo", WorkingDirectory: "/non/existent"},
			expectError: true,
		},
		{
			name:       "valid config",
			serverName: "test",
			config:     &types.MCPServerConfig{Command: "echo", Enabled: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.AddServer(tt.serverName, tt.config)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestMCPManager_RemoveServer(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	manager := client.MCP()

	// Add a server first
	serverConfig := &types.MCPServerConfig{
		Command: "echo",
		Enabled: true,
	}

	err = manager.AddServer("test-server", serverConfig)
	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}

	// Remove the server
	err = manager.RemoveServer("test-server")
	if err != nil {
		t.Fatalf("Failed to remove server: %v", err)
	}

	// Verify server was removed
	servers := manager.ListServers()
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers, got %d", len(servers))
	}

	// Test removing non-existent server
	err = manager.RemoveServer("non-existent")
	if err == nil {
		t.Error("Expected error when removing non-existent server")
	}
}

func TestMCPManager_EnableDisableServer(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	manager := client.MCP()

	// Add a disabled server
	serverConfig := &types.MCPServerConfig{
		Command: "echo",
		Enabled: false,
	}

	err = manager.AddServer("test-server", serverConfig)
	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}

	// Enable the server
	err = manager.EnableServer("test-server")
	if err != nil {
		t.Fatalf("Failed to enable server: %v", err)
	}

	// Verify server is enabled
	server, err := manager.GetServer("test-server")
	if err != nil {
		t.Fatalf("Failed to get server: %v", err)
	}

	if !server.Enabled {
		t.Error("Expected server to be enabled")
	}

	// Disable the server
	err = manager.DisableServer("test-server")
	if err != nil {
		t.Fatalf("Failed to disable server: %v", err)
	}

	// Verify server is disabled
	server, err = manager.GetServer("test-server")
	if err != nil {
		t.Fatalf("Failed to get server: %v", err)
	}

	if server.Enabled {
		t.Error("Expected server to be disabled")
	}
}

func TestMCPManager_GetEnabledServers(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	manager := client.MCP()

	// Add multiple servers with different enabled states
	servers := map[string]bool{
		"enabled-1":  true,
		"enabled-2":  true,
		"disabled-1": false,
		"disabled-2": false,
	}

	for name, enabled := range servers {
		serverConfig := &types.MCPServerConfig{
			Command: "echo",
			Enabled: enabled,
		}
		err = manager.AddServer(name, serverConfig)
		if err != nil {
			t.Fatalf("Failed to add server %s: %v", name, err)
		}
	}

	// Get enabled servers
	enabledServers := manager.GetEnabledServers()

	// Should have 2 enabled servers
	if len(enabledServers) != 2 {
		t.Errorf("Expected 2 enabled servers, got %d", len(enabledServers))
	}

	// Check specific servers
	expectedEnabled := []string{"enabled-1", "enabled-2"}
	for _, name := range expectedEnabled {
		if _, exists := enabledServers[name]; !exists {
			t.Errorf("Expected %s to be in enabled servers", name)
		}
	}

	// Check that disabled servers are not included
	unexpectedEnabled := []string{"disabled-1", "disabled-2"}
	for _, name := range unexpectedEnabled {
		if _, exists := enabledServers[name]; exists {
			t.Errorf("Did not expect %s to be in enabled servers", name)
		}
	}
}

func TestMCPManager_ApplyConfiguration(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	manager := client.MCP()

	// Add some servers
	servers := []struct {
		name    string
		config  *types.MCPServerConfig
		enabled bool
	}{
		{
			name: "filesystem",
			config: &types.MCPServerConfig{
				Command: "npx",
				Args:    []string{"@modelcontextprotocol/server-filesystem", tempDir},
				Enabled: true,
			},
		},
		{
			name: "git",
			config: &types.MCPServerConfig{
				Command:          "npx",
				Args:             []string{"@modelcontextprotocol/server-git"},
				WorkingDirectory: tempDir,
				Enabled:          false, // Disabled server should not appear in config
			},
		},
	}

	for _, server := range servers {
		err = manager.AddServer(server.name, server.config)
		if err != nil {
			t.Fatalf("Failed to add server %s: %v", server.name, err)
		}
	}

	// Apply configuration
	err = manager.ApplyConfiguration(ctx)
	if err != nil {
		t.Fatalf("Failed to apply configuration: %v", err)
	}

	// Check that configuration file was created
	configPath := filepath.Join(tempDir, ".claude", "mcp.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Expected MCP configuration file to be created")
	}

	// Read and verify configuration file content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var mcpConfig map[string]interface{}
	err = json.Unmarshal(data, &mcpConfig)
	if err != nil {
		t.Fatalf("Failed to parse config file: %v", err)
	}

	// Check that mcpServers section exists
	servers_config, exists := mcpConfig["mcpServers"].(map[string]interface{})
	if !exists {
		t.Fatal("Expected mcpServers section in config")
	}

	// Should only contain enabled server (filesystem)
	if len(servers_config) != 1 {
		t.Errorf("Expected 1 server in config, got %d", len(servers_config))
	}

	if _, exists := servers_config["filesystem"]; !exists {
		t.Error("Expected filesystem server in config")
	}

	if _, exists := servers_config["git"]; exists {
		t.Error("Did not expect disabled git server in config")
	}
}

func TestMCPManager_SaveLoadFromFile(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	manager := client.MCP()

	// Add some servers
	serverConfig := &types.MCPServerConfig{
		Command: "npx",
		Args:    []string{"@modelcontextprotocol/server-filesystem", tempDir},
		Environment: map[string]string{
			"NODE_ENV": "development",
		},
		WorkingDirectory: tempDir,
		Enabled:          true,
	}

	err = manager.AddServer("filesystem", serverConfig)
	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}

	// Save to file
	configFile := filepath.Join(tempDir, "mcp-config.json")
	err = manager.SaveToFile(configFile)
	if err != nil {
		t.Fatalf("Failed to save to file: %v", err)
	}

	// Create new manager and load from file
	client2, err := NewClaudeCodeClient(ctx, &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	})
	if err != nil {
		t.Fatalf("Failed to create second client: %v", err)
	}
	defer client2.Close()

	manager2 := client2.MCP()
	err = manager2.LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("Failed to load from file: %v", err)
	}

	// Verify servers were loaded
	servers := manager2.ListServers()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(servers))
	}

	server, exists := servers["filesystem"]
	if !exists {
		t.Fatal("Expected filesystem server to be loaded")
	}

	if server.Command != "npx" {
		t.Errorf("Expected command 'npx', got '%s'", server.Command)
	}

	if len(server.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(server.Args))
	}

	if server.Environment["NODE_ENV"] != "development" {
		t.Errorf("Expected NODE_ENV=development, got '%s'", server.Environment["NODE_ENV"])
	}
}

func TestMCPManager_AddCommonServers(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	manager := client.MCP()

	// Add common servers
	err = manager.AddCommonServers()
	if err != nil {
		t.Fatalf("Failed to add common servers: %v", err)
	}

	// Verify common servers were added
	servers := manager.ListServers()
	expectedServers := []string{"filesystem", "git", "brave-search"}

	for _, expected := range expectedServers {
		if _, exists := servers[expected]; !exists {
			t.Errorf("Expected %s server to be added", expected)
		}
	}

	// All servers should be disabled by default for security
	for _, server := range servers {
		if server.Enabled {
			t.Error("Expected common servers to be disabled by default")
		}
	}
}

func TestClaudeCodeClient_MCPIntegration(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	// Test client MCP methods
	serverConfig := &types.MCPServerConfig{
		Command: "echo",
		Args:    []string{"hello"},
		Enabled: false,
	}

	// Add server through client
	err = client.AddMCPServer(ctx, "test-server", serverConfig)
	if err != nil {
		t.Fatalf("Failed to add MCP server: %v", err)
	}

	// Enable server through client
	err = client.EnableMCPServer(ctx, "test-server")
	if err != nil {
		t.Fatalf("Failed to enable MCP server: %v", err)
	}

	// Verify server is enabled
	server, err := client.GetMCPServer("test-server")
	if err != nil {
		t.Fatalf("Failed to get MCP server: %v", err)
	}

	if !server.Enabled {
		t.Error("Expected server to be enabled")
	}

	// List servers through client
	servers := client.ListMCPServers()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(servers))
	}

	// Setup common servers
	err = client.SetupCommonMCPServers(ctx)
	if err != nil {
		t.Fatalf("Failed to setup common MCP servers: %v", err)
	}

	// Should now have more servers
	servers = client.ListMCPServers()
	if len(servers) < 4 { // test-server + 3 common servers
		t.Errorf("Expected at least 4 servers, got %d", len(servers))
	}

	// Disable server through client
	err = client.DisableMCPServer(ctx, "test-server")
	if err != nil {
		t.Fatalf("Failed to disable MCP server: %v", err)
	}

	// Verify server is disabled
	server, err = client.GetMCPServer("test-server")
	if err != nil {
		t.Fatalf("Failed to get MCP server: %v", err)
	}

	if server.Enabled {
		t.Error("Expected server to be disabled")
	}

	// Remove server through client
	err = client.RemoveMCPServer(ctx, "test-server")
	if err != nil {
		t.Fatalf("Failed to remove MCP server: %v", err)
	}

	// Verify server was removed
	_, err = client.GetMCPServer("test-server")
	if err == nil {
		t.Error("Expected error when getting removed server")
	}
}