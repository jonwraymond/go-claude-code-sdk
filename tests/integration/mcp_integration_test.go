// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MCPIntegrationSuite tests MCP server integration with Claude Code CLI
type MCPIntegrationSuite struct {
	suite.Suite
	client     *client.ClaudeCodeClient
	mcpManager *client.MCPManager
	config     *types.ClaudeCodeConfig
}

func (s *MCPIntegrationSuite) SetupSuite() {
	// Skip if not in integration test mode
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		s.T().Skip("Integration tests disabled. Set INTEGRATION_TESTS=true to run")
	}

	// Skip if MCP tests are disabled
	if os.Getenv("SKIP_MCP_TESTS") == "true" {
		s.T().Skip("MCP tests disabled. Remove SKIP_MCP_TESTS to run")
	}

	// Ensure API key is set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		s.T().Skip("ANTHROPIC_API_KEY not set")
	}

	// Create config
	s.config = types.NewClaudeCodeConfig()
	s.config.APIKey = apiKey
	s.config.ClaudeExecutable = "claude"
	s.config.Timeout = 30 * time.Second
	
	// Enable TestMode in CI environment to skip Claude Code CLI requirement
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		s.config.TestMode = true
	}

	// Create client
	var err error
	ctx := context.Background()
	s.client, err = client.NewClaudeCodeClient(ctx, s.config)
	require.NoError(s.T(), err)

	// Get MCP manager
	s.mcpManager = s.client.MCP()
}

func (s *MCPIntegrationSuite) TearDownSuite() {
	if s.client != nil {
		// Remove all MCP servers
		if s.mcpManager != nil {
			servers := s.mcpManager.ListServers()
			for name := range servers {
				s.mcpManager.RemoveServer(name)
			}
		}
		s.client.Close()
	}
}

func (s *MCPIntegrationSuite) TestRegisterMCPServer() {
	// Test basic MCP server registration
	serverConfig := &types.MCPServerConfig{
		Command: "echo",
		Args:    []string{"MCP Echo Server"},
		Environment: map[string]string{
			"TEST_ENV": "test_value",
		},
		Enabled: true,
	}

	err := s.mcpManager.AddServer("test-echo", serverConfig)
	require.NoError(s.T(), err)

	// Check if server is registered
	servers := s.mcpManager.ListServers()
	assert.Contains(s.T(), servers, "test-echo")
}

func (s *MCPIntegrationSuite) TestEnableDisableMCPServer() {
	// Register a simple server
	serverConfig := &types.MCPServerConfig{
		Command: "sleep",
		Args:    []string{"30"}, // Sleep for 30 seconds
		Enabled: false, // Start disabled
	}

	err := s.mcpManager.AddServer("test-sleep", serverConfig)
	require.NoError(s.T(), err)

	// Enable the server
	err = s.mcpManager.EnableServer("test-sleep")
	require.NoError(s.T(), err)

	// Check if server is in enabled servers list
	enabledServers := s.mcpManager.GetEnabledServers()
	assert.Contains(s.T(), enabledServers, "test-sleep")

	// Disable the server
	err = s.mcpManager.DisableServer("test-sleep")
	require.NoError(s.T(), err)

	// Check server is not in enabled list
	enabledServers = s.mcpManager.GetEnabledServers()
	assert.NotContains(s.T(), enabledServers, "test-sleep")

	// Clean up
	s.mcpManager.RemoveServer("test-sleep")
}

func (s *MCPIntegrationSuite) TestMCPServerConfiguration() {
	// Skip this test if no MCP servers are available
	mcpServerPath := os.Getenv("TEST_MCP_SERVER_PATH")
	if mcpServerPath == "" {
		s.T().Skip("TEST_MCP_SERVER_PATH not set. Skipping MCP server tools test")
	}
	_ = mcpServerPath // Mark as used for test

	// Register an actual MCP server (if available)
	serverConfig := &types.MCPServerConfig{
		Command: mcpServerPath,
		Args:    []string{},
		Enabled: true,
	}

	err := s.mcpManager.AddServer("test-mcp", serverConfig)
	require.NoError(s.T(), err)

	// Get the server configuration back
	retrievedConfig, err := s.mcpManager.GetServer("test-mcp")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), mcpServerPath, retrievedConfig.Command)
	assert.True(s.T(), retrievedConfig.Enabled)

	// Clean up
	s.mcpManager.RemoveServer("test-mcp")

	// Test that we can list tools after MCP server registration
	allTools := s.client.Tools().ListTools()
	assert.NotNil(s.T(), allTools, "Should be able to list tools")
}

func (s *MCPIntegrationSuite) TestMultipleMCPServers() {
	ctx := context.Background()

	// Register multiple servers
	servers := []struct {
		name   string
		config *types.MCPServerConfig
	}{
		{
			name: "server1",
			config: &types.MCPServerConfig{
				Command: "echo",
				Args:    []string{"Server 1"},
				Enabled: true,
			},
		},
		{
			name: "server2",
			config: &types.MCPServerConfig{
				Command: "echo",
				Args:    []string{"Server 2"},
				Enabled: true,
			},
		},
	}

	// Register all servers
	for _, server := range servers {
		err := s.mcpManager.AddServer(server.name, server.config)
		require.NoError(s.T(), err)
	}

	// List all servers
	registeredServers := s.mcpManager.ListServers()
	assert.GreaterOrEqual(s.T(), len(registeredServers), 2)

	// Check that all servers are registered
	for _, server := range servers {
		assert.Contains(s.T(), registeredServers, server.name)
	}

	// Clean up
	for _, server := range servers {
		s.mcpManager.RemoveServer(server.name)
	}
}

func (s *MCPIntegrationSuite) TestMCPServerGetAndRemove() {
	// Register a server
	serverConfig := &types.MCPServerConfig{
		Command: "sleep",
		Args:    []string{"10"},
		Enabled: true,
	}

	err := s.mcpManager.AddServer("get-test", serverConfig)
	require.NoError(s.T(), err)

	// Get the server config
	retrievedConfig, err := s.mcpManager.GetServer("get-test")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "sleep", retrievedConfig.Command)
	assert.Equal(s.T(), []string{"10"}, retrievedConfig.Args)
	assert.True(s.T(), retrievedConfig.Enabled)

	// Remove the server
	err = s.mcpManager.RemoveServer("get-test")
	require.NoError(s.T(), err)

	// Verify it's removed
	servers := s.mcpManager.ListServers()
	assert.NotContains(s.T(), servers, "get-test")
}

func (s *MCPIntegrationSuite) TestMCPServerEnabledDisabledLists() {
	// Register a server that's enabled
	enabledConfig := &types.MCPServerConfig{
		Command: "echo",
		Args:    []string{"enabled"},
		Enabled: true,
	}

	err := s.mcpManager.AddServer("enabled-test", enabledConfig)
	require.NoError(s.T(), err)

	// Register a server that's disabled
	disabledConfig := &types.MCPServerConfig{
		Command: "echo",
		Args:    []string{"disabled"},
		Enabled: false,
	}

	err = s.mcpManager.AddServer("disabled-test", disabledConfig)
	require.NoError(s.T(), err)

	// Check enabled servers list
	enabledServers := s.mcpManager.GetEnabledServers()
	assert.Contains(s.T(), enabledServers, "enabled-test")
	assert.NotContains(s.T(), enabledServers, "disabled-test")

	// Clean up
	s.mcpManager.RemoveServer("enabled-test")
	s.mcpManager.RemoveServer("disabled-test")
}

func TestMCPIntegrationSuite(t *testing.T) {
	suite.Run(t, new(MCPIntegrationSuite))
}