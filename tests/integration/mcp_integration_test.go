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

	// Create client
	var err error
	s.client, err = client.NewClaudeCodeClient(s.config)
	require.NoError(s.T(), err)

	// Get MCP manager
	s.mcpManager = s.client.MCP()
}

func (s *MCPIntegrationSuite) TearDownSuite() {
	if s.client != nil {
		// Stop all MCP servers
		if s.mcpManager != nil {
			ctx := context.Background()
			servers := s.mcpManager.ListServers()
			for name := range servers {
				s.mcpManager.StopServer(ctx, name)
			}
		}
		s.client.Close()
	}
}

func (s *MCPIntegrationSuite) TestRegisterMCPServer() {
	// Test basic MCP server registration
	serverConfig := &types.MCPServerConfig{
		Name:    "test-echo-server",
		Command: "echo",
		Args:    []string{"MCP Echo Server"},
		Environment: map[string]string{
			"TEST_ENV": "test_value",
		},
	}

	err := s.mcpManager.RegisterServer("test-echo", serverConfig)
	require.NoError(s.T(), err)

	// Check if server is registered
	servers := s.mcpManager.ListServers()
	assert.Contains(s.T(), servers, "test-echo")
}

func (s *MCPIntegrationSuite) TestStartStopMCPServer() {
	ctx := context.Background()

	// Register a simple server
	serverConfig := &types.MCPServerConfig{
		Name:    "test-server",
		Command: "sleep",
		Args:    []string{"30"}, // Sleep for 30 seconds
	}

	err := s.mcpManager.RegisterServer("test-sleep", serverConfig)
	require.NoError(s.T(), err)

	// Start the server
	err = s.mcpManager.StartServer(ctx, "test-sleep")
	require.NoError(s.T(), err)

	// Check server status
	status, err := s.mcpManager.GetServerStatus(ctx, "test-sleep")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "running", status.State)

	// Stop the server
	err = s.mcpManager.StopServer(ctx, "test-sleep")
	require.NoError(s.T(), err)

	// Check server is stopped
	status, err = s.mcpManager.GetServerStatus(ctx, "test-sleep")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "stopped", status.State)
}

func (s *MCPIntegrationSuite) TestMCPServerTools() {
	// Skip this test if no MCP servers are available
	mcpServerPath := os.Getenv("TEST_MCP_SERVER_PATH")
	if mcpServerPath == "" {
		s.T().Skip("TEST_MCP_SERVER_PATH not set. Skipping MCP server tools test")
	}

	ctx := context.Background()

	// Register an actual MCP server (if available)
	serverConfig := &types.MCPServerConfig{
		Name:    "test-mcp-server",
		Command: mcpServerPath,
		Args:    []string{},
	}

	err := s.mcpManager.RegisterServer("test-mcp", serverConfig)
	require.NoError(s.T(), err)

	// Start the server
	err = s.mcpManager.StartServer(ctx, "test-mcp")
	require.NoError(s.T(), err)
	defer s.mcpManager.StopServer(ctx, "test-mcp")

	// Wait for server to be ready
	time.Sleep(2 * time.Second)

	// Get tools provided by the MCP server
	tools, err := s.mcpManager.GetServerTools(ctx, "test-mcp")
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), tools, "MCP server should provide tools")

	// List all tools (including MCP tools)
	allTools, err := s.client.Tools().ListTools(ctx)
	require.NoError(s.T(), err)

	// Check that MCP tools are included
	mcpToolFound := false
	for _, tool := range allTools {
		if tool.Source == "test-mcp" {
			mcpToolFound = true
			break
		}
	}
	assert.True(s.T(), mcpToolFound, "MCP server tools should be available")
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
				Name:    "echo-server-1",
				Command: "echo",
				Args:    []string{"Server 1"},
			},
		},
		{
			name: "server2",
			config: &types.MCPServerConfig{
				Name:    "echo-server-2",
				Command: "echo",
				Args:    []string{"Server 2"},
			},
		},
	}

	// Register all servers
	for _, server := range servers {
		err := s.mcpManager.RegisterServer(server.name, server.config)
		require.NoError(s.T(), err)
	}

	// List all servers
	registeredServers := s.mcpManager.ListServers()
	assert.GreaterOrEqual(s.T(), len(registeredServers), 2)

	// Start all servers
	for _, server := range servers {
		err := s.mcpManager.StartServer(ctx, server.name)
		// Echo command will exit immediately, so we might get an error
		// but the server should still be registered
		_ = err
	}

	// Check that all servers are registered
	for _, server := range servers {
		assert.Contains(s.T(), registeredServers, server.name)
	}
}

func (s *MCPIntegrationSuite) TestMCPServerRestart() {
	ctx := context.Background()

	// Register a server
	serverConfig := &types.MCPServerConfig{
		Name:    "restart-test",
		Command: "sleep",
		Args:    []string{"10"},
	}

	err := s.mcpManager.RegisterServer("restart-test", serverConfig)
	require.NoError(s.T(), err)

	// Start the server
	err = s.mcpManager.StartServer(ctx, "restart-test")
	require.NoError(s.T(), err)

	// Get initial status
	status1, err := s.mcpManager.GetServerStatus(ctx, "restart-test")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "running", status1.State)

	// Restart the server
	err = s.mcpManager.RestartServer(ctx, "restart-test")
	require.NoError(s.T(), err)

	// Check status after restart
	status2, err := s.mcpManager.GetServerStatus(ctx, "restart-test")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "running", status2.State)

	// Stop the server
	err = s.mcpManager.StopServer(ctx, "restart-test")
	require.NoError(s.T(), err)
}

func (s *MCPIntegrationSuite) TestMCPServerHealthCheck() {
	ctx := context.Background()

	// Register a server that will exit quickly
	serverConfig := &types.MCPServerConfig{
		Name:    "health-test",
		Command: "echo",
		Args:    []string{"Quick exit"},
	}

	err := s.mcpManager.RegisterServer("health-test", serverConfig)
	require.NoError(s.T(), err)

	// Start the server (it will exit immediately)
	_ = s.mcpManager.StartServer(ctx, "health-test")

	// Wait a moment
	time.Sleep(100 * time.Millisecond)

	// Health check should show server is not running
	healthy, err := s.mcpManager.HealthCheck(ctx, "health-test")
	require.NoError(s.T(), err)
	assert.False(s.T(), healthy, "Echo server should have exited")
}

func TestMCPIntegrationSuite(t *testing.T) {
	suite.Run(t, new(MCPIntegrationSuite))
}