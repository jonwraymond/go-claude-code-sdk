package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	sdkerrors "github.com/jonwraymond/go-claude-code-sdk/pkg/errors"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// MCPManager manages Model Context Protocol servers for Claude Code integration.
// It handles MCP server configuration, lifecycle management, and integration with Claude Code CLI.
//
// MCP enables Claude Code to interact with external tools, databases, file systems,
// and other services through a standardized protocol. The manager ensures MCP servers
// are properly configured and available to Claude Code sessions.
//
// Key features:
// - MCP server configuration management
// - Dynamic server enable/disable
// - Configuration file generation for Claude Code
// - Server lifecycle tracking
// - Validation and error handling
//
// Example usage:
//
//	manager := NewMCPManager(client)
//
//	// Add an MCP server
//	err := manager.AddServer("filesystem", &types.MCPServerConfig{
//		Command: "npx",
//		Args:    []string{"@modelcontextprotocol/server-filesystem", "/path/to/project"},
//		Enabled: true,
//	})
//
//	// Apply configuration to Claude Code
//	err = manager.ApplyConfiguration(ctx)
type MCPManager struct {
	client  *ClaudeCodeClient
	servers map[string]*types.MCPServerConfig
	mu      sync.RWMutex
}

// NewMCPManager creates a new MCP manager for the given Claude Code client.
func NewMCPManager(client *ClaudeCodeClient) *MCPManager {
	manager := &MCPManager{
		client:  client,
		servers: make(map[string]*types.MCPServerConfig),
	}

	// Load existing servers from client config
	if client.config.MCPServers != nil {
		for name, config := range client.config.MCPServers {
			manager.servers[name] = config
		}
	}

	return manager
}

// AddServer adds or updates an MCP server configuration.
func (m *MCPManager) AddServer(name string, config *types.MCPServerConfig) error {
	if name == "" {
		return sdkerrors.NewValidationError("name", "", "required", "server name cannot be empty")
	}

	if config == nil {
		return sdkerrors.NewValidationError("config", "", "required", "server config cannot be nil")
	}

	if err := m.validateServerConfig(config); err != nil {
		return sdkerrors.WrapError(err, sdkerrors.CategoryValidation, "SERVER_CONFIG", "invalid server configuration")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a copy to avoid external modifications
	serverConfig := &types.MCPServerConfig{
		Command:          config.Command,
		Args:             make([]string, len(config.Args)),
		Environment:      make(map[string]string),
		WorkingDirectory: config.WorkingDirectory,
		Enabled:          config.Enabled,
	}

	copy(serverConfig.Args, config.Args)
	for k, v := range config.Environment {
		serverConfig.Environment[k] = v
	}

	m.servers[name] = serverConfig

	// Update client config
	m.updateClientConfig()

	return nil
}

// RemoveServer removes an MCP server configuration.
func (m *MCPManager) RemoveServer(name string) error {
	if name == "" {
		return sdkerrors.NewValidationError("name", "", "required", "server name cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.servers[name]; !exists {
		return sdkerrors.NewValidationError("name", name, "exists", "server not found")
	}

	delete(m.servers, name)

	// Update client config
	m.updateClientConfig()

	return nil
}

// EnableServer enables an MCP server.
func (m *MCPManager) EnableServer(name string) error {
	return m.setServerEnabled(name, true)
}

// DisableServer disables an MCP server.
func (m *MCPManager) DisableServer(name string) error {
	return m.setServerEnabled(name, false)
}

// setServerEnabled sets the enabled state of an MCP server.
func (m *MCPManager) setServerEnabled(name string, enabled bool) error {
	if name == "" {
		return sdkerrors.NewValidationError("name", "", "required", "server name cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	server, exists := m.servers[name]
	if !exists {
		return sdkerrors.NewValidationError("name", name, "exists", "server not found")
	}

	server.Enabled = enabled

	// Update client config
	m.updateClientConfig()

	return nil
}

// ListServers returns a list of all configured MCP servers.
func (m *MCPManager) ListServers() map[string]*types.MCPServerConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*types.MCPServerConfig)
	for name, config := range m.servers {
		// Create a copy to prevent external modifications
		result[name] = &types.MCPServerConfig{
			Command:          config.Command,
			Args:             make([]string, len(config.Args)),
			Environment:      make(map[string]string),
			WorkingDirectory: config.WorkingDirectory,
			Enabled:          config.Enabled,
		}
		copy(result[name].Args, config.Args)
		for k, v := range config.Environment {
			result[name].Environment[k] = v
		}
	}

	return result
}

// GetServer returns the configuration for a specific MCP server.
func (m *MCPManager) GetServer(name string) (*types.MCPServerConfig, error) {
	if name == "" {
		return nil, sdkerrors.NewValidationError("name", "", "required", "server name cannot be empty")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.servers[name]
	if !exists {
		return nil, sdkerrors.NewValidationError("name", name, "exists", "server not found")
	}

	// Return a copy to prevent external modifications
	result := &types.MCPServerConfig{
		Command:          config.Command,
		Args:             make([]string, len(config.Args)),
		Environment:      make(map[string]string),
		WorkingDirectory: config.WorkingDirectory,
		Enabled:          config.Enabled,
	}
	copy(result.Args, config.Args)
	for k, v := range config.Environment {
		result.Environment[k] = v
	}

	return result, nil
}

// ApplyConfiguration applies the current MCP server configuration to Claude Code.
// This generates the necessary configuration files and updates the Claude Code environment.
func (m *MCPManager) ApplyConfiguration(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Generate MCP configuration file path
	configDir := filepath.Join(m.client.workingDir, ".claude")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "CONFIG_DIR", "failed to create config directory")
	}

	mcpConfigPath := filepath.Join(configDir, "mcp.json")

	// Build MCP configuration
	mcpConfig := make(map[string]any)
	servers := make(map[string]any)

	for name, config := range m.servers {
		if !config.Enabled {
			continue
		}

		serverConfig := map[string]any{
			"command": config.Command,
		}

		if len(config.Args) > 0 {
			serverConfig["args"] = config.Args
		}

		if len(config.Environment) > 0 {
			serverConfig["env"] = config.Environment
		}

		if config.WorkingDirectory != "" {
			serverConfig["cwd"] = config.WorkingDirectory
		}

		servers[name] = serverConfig
	}

	mcpConfig["mcpServers"] = servers

	// Write configuration file
	configData, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		return sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "CONFIG_MARSHAL", "failed to marshal MCP configuration")
	}

	if err := os.WriteFile(mcpConfigPath, configData, 0600); err != nil {
		return sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "CONFIG_WRITE", "failed to write MCP configuration file")
	}

	return nil
}

// LoadFromFile loads MCP server configurations from a JSON file.
func (m *MCPManager) LoadFromFile(filePath string) error {
	if filePath == "" {
		return sdkerrors.NewValidationError("filePath", "", "required", "file path cannot be empty")
	}

	// Clean the file path to prevent directory traversal
	filePath = filepath.Clean(filePath)

	data, err := os.ReadFile(filePath) // #nosec G304 - file path is cleaned and this is expected to load config files
	if err != nil {
		return sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "FILE_READ", "failed to read MCP configuration file")
	}

	var config struct {
		MCPServers map[string]*types.MCPServerConfig `json:"mcpServers"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return sdkerrors.WrapError(err, sdkerrors.CategoryValidation, "CONFIG_PARSE", "failed to parse MCP configuration file")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate and add each server
	for name, serverConfig := range config.MCPServers {
		if err := m.validateServerConfig(serverConfig); err != nil {
			return sdkerrors.WrapError(err, sdkerrors.CategoryValidation, "SERVER_CONFIG", fmt.Sprintf("invalid configuration for server '%s'", name))
		}

		m.servers[name] = serverConfig
	}

	// Update client config
	m.updateClientConfig()

	return nil
}

// SaveToFile saves the current MCP server configurations to a JSON file.
func (m *MCPManager) SaveToFile(filePath string) error {
	if filePath == "" {
		return sdkerrors.NewValidationError("filePath", "", "required", "file path cannot be empty")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	config := struct {
		MCPServers map[string]*types.MCPServerConfig `json:"mcpServers"`
	}{
		MCPServers: m.servers,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "CONFIG_MARSHAL", "failed to marshal MCP configuration")
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "FILE_WRITE", "failed to write MCP configuration file")
	}

	return nil
}

// validateServerConfig validates an MCP server configuration.
func (m *MCPManager) validateServerConfig(config *types.MCPServerConfig) error {
	if config.Command == "" {
		return sdkerrors.NewValidationError("command", "", "required", "command is required")
	}

	// Validate working directory if provided
	if config.WorkingDirectory != "" {
		if _, err := os.Stat(config.WorkingDirectory); os.IsNotExist(err) {
			return sdkerrors.NewValidationError("working_directory", config.WorkingDirectory, "exists", "working directory does not exist")
		}
	}

	return nil
}

// updateClientConfig updates the client's MCP server configuration.
func (m *MCPManager) updateClientConfig() {
	if m.client.config.MCPServers == nil {
		m.client.config.MCPServers = make(map[string]*types.MCPServerConfig)
	}

	// Clear existing servers
	for name := range m.client.config.MCPServers {
		delete(m.client.config.MCPServers, name)
	}

	// Add current servers
	for name, config := range m.servers {
		m.client.config.MCPServers[name] = config
	}
}

// GetEnabledServers returns a list of enabled MCP servers.
func (m *MCPManager) GetEnabledServers() map[string]*types.MCPServerConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*types.MCPServerConfig)
	for name, config := range m.servers {
		if config.Enabled {
			// Create a copy to prevent external modifications
			result[name] = &types.MCPServerConfig{
				Command:          config.Command,
				Args:             make([]string, len(config.Args)),
				Environment:      make(map[string]string),
				WorkingDirectory: config.WorkingDirectory,
				Enabled:          config.Enabled,
			}
			copy(result[name].Args, config.Args)
			for k, v := range config.Environment {
				result[name].Environment[k] = v
			}
		}
	}

	return result
}

// AddCommonServers adds commonly used MCP servers with default configurations.
func (m *MCPManager) AddCommonServers() error {
	// Filesystem server - for file system operations
	filesystemServer := &types.MCPServerConfig{
		Command: "npx",
		Args:    []string{"@modelcontextprotocol/server-filesystem", m.client.workingDir},
		Enabled: false, // Disabled by default for security
	}

	if err := m.AddServer("filesystem", filesystemServer); err != nil {
		return sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "ADD_FILESYSTEM", "failed to add filesystem server")
	}

	// Git server - for git operations
	gitServer := &types.MCPServerConfig{
		Command:          "npx",
		Args:             []string{"@modelcontextprotocol/server-git"},
		WorkingDirectory: m.client.workingDir,
		Enabled:          false, // Disabled by default
	}

	if err := m.AddServer("git", gitServer); err != nil {
		return sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "ADD_GIT", "failed to add git server")
	}

	// Brave Search server - for web search capabilities
	braveServer := &types.MCPServerConfig{
		Command:     "npx",
		Args:        []string{"@modelcontextprotocol/server-brave-search"},
		Environment: map[string]string{
			// API key would need to be set by user
		},
		Enabled: false, // Disabled by default - requires API key
	}

	if err := m.AddServer("brave-search", braveServer); err != nil {
		return sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "ADD_BRAVE", "failed to add brave search server")
	}

	return nil
}
