# MCP Integration Example

This example demonstrates Model Context Protocol (MCP) server integration with Claude Code, including listing servers, adding new servers, setting up common servers, and using MCP context in queries.

## What You'll Learn

- How to list and manage MCP servers
- Adding custom MCP server configurations
- Setting up common MCP servers automatically
- Using MCP context to enhance Claude's capabilities
- MCP server validation and troubleshooting
- Best practices for MCP integration

## Code Overview

The example includes four MCP integration patterns:

### 1. List MCP Servers
```go
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
    fmt.Printf("    Enabled: %t\n", config.Enabled)
}
```

Shows how to discover and inspect currently configured MCP servers.

### 2. Add MCP Server
```go
mcpConfig := &types.MCPServerConfig{
    Command: "npx",
    Args:    []string{"@modelcontextprotocol/server-filesystem", "/tmp"},
    Environment: map[string]string{
        "NODE_ENV": "production",
    },
    Enabled: true,
}

err := claudeClient.AddMCPServer(ctx, "filesystem", mcpConfig)
if err != nil {
    fmt.Printf("Failed to add filesystem MCP server: %v\n", err)
} else {
    fmt.Println("Successfully added filesystem MCP server")
}
```

Demonstrates adding a custom MCP server with specific configuration.

### 3. Setup Common MCP Servers
```go
err := claudeClient.SetupCommonMCPServers(ctx)
if err != nil {
    fmt.Printf("Failed to setup common MCP servers: %v\n", err)
} else {
    fmt.Println("Successfully setup common MCP servers")
    
    // List all servers after setup
    servers := claudeClient.ListMCPServers()
    fmt.Printf("Total MCP servers configured: %d\n", len(servers))
}
```

Shows automatic setup of commonly used MCP servers.

### 4. Query with MCP Context
```go
// Check if MCP servers are available
servers := claudeClient.ListMCPServers()
if len(servers) == 0 {
    fmt.Println("No MCP servers available for queries")
    return
}

// Create query that benefits from MCP context
request := &types.QueryRequest{
    Messages: []types.Message{
        {
            Role:    types.RoleUser,
            Content: "What tools and capabilities are available through the configured MCP servers?",
        },
    },
    MaxTokens: 500,
}

response, err := claudeClient.Query(ctx, request)
if err != nil {
    log.Printf("Query failed: %v", err)
    return
}

// Display response
if len(response.Content) > 0 {
    fmt.Printf("Response:\n%s\n", response.Content[0].Text)
}
```

Demonstrates using MCP context to enhance Claude's responses with additional tools and capabilities.

## Prerequisites

1. **Go 1.21+** installed
2. **Claude Code CLI** installed: `npm install -g @anthropic/claude-code`
3. **Node.js and npm** (for npm-based MCP servers)
4. **Authentication** configured
5. **MCP servers** installed (optional, for full functionality)

## MCP Server Installation

### Common MCP Servers
```bash
# Filesystem server (file system access)
npm install -g @modelcontextprotocol/server-filesystem

# Git server (Git repository operations)
npm install -g @modelcontextprotocol/server-git

# Postgres server (database operations)
npm install -g @modelcontextprotocol/server-postgres

# Web search server (web search capabilities)
npm install -g @modelcontextprotocol/server-web-search
```

### Verify Installation
```bash
# Test filesystem server
npx @modelcontextprotocol/server-filesystem --version

# Test Git server
npx @modelcontextprotocol/server-git --version
```

## Running the Example

### Setup
```bash
# Install Claude Code CLI
npm install -g @anthropic/claude-code

# Setup authentication
export ANTHROPIC_API_KEY="your-api-key"
# OR use subscription auth
claude setup-token

# Install MCP servers (optional but recommended)
npm install -g @modelcontextprotocol/server-filesystem
npm install -g @modelcontextprotocol/server-git
```

### Run the Example
```bash
cd examples/mcp_integration
go run main.go
```

## Expected Output

```
=== MCP Server Integration Example ===

--- Example 1: List MCP Servers ---
No MCP servers configured

--- Example 2: Add Filesystem MCP Server ---
Successfully added filesystem MCP server
Verified: filesystem server added with command: npx

--- Example 3: Setup Common MCP Servers ---
Setting up common MCP servers...
Successfully setup common MCP servers
Total MCP servers configured: 4

--- Example 4: Query with MCP Context ---
Querying with MCP context enabled...

Response:
Based on the configured MCP servers, I have access to several powerful tools and capabilities:

**Filesystem Operations:**
- Read and write files in the specified directory (/tmp)
- List directory contents
- Create and manage file structures
- Monitor file changes

**Git Repository Management:**
- Access Git repository information
- Read commit history and branches
- Analyze repository structure
- Track changes and differences

**Web Search Capabilities:**
- Search the web for current information
- Retrieve and analyze web content
- Access real-time data and news

**Database Operations (if configured):**
- Query PostgreSQL databases
- Execute SQL commands
- Analyze database schemas
- Generate database reports

These tools extend my capabilities beyond text generation to include file manipulation, version control, web research, and database operations, making me more useful for development tasks and system administration.

Token usage - Input: 45, Output: 187
```

## Key Concepts

### MCP (Model Context Protocol)

MCP is a protocol that allows Claude Code to integrate with external tools and services, extending its capabilities beyond text generation.

**Benefits:**
- **Extended Capabilities**: Access to file systems, databases, web APIs
- **Real-time Data**: Current information from external sources
- **Tool Integration**: Use existing tools and services
- **Contextual Awareness**: Better understanding of your specific environment

### MCP Server Types

| Server Type | Purpose | Example Use Cases |
|-------------|---------|-------------------|
| **Filesystem** | File operations | Reading configs, analyzing codebases, managing files |
| **Git** | Version control | Analyzing repo history, tracking changes, branch management |
| **Database** | Data operations | Querying databases, schema analysis, data reports |
| **Web Search** | Internet access | Research, current events, API documentation |
| **Custom** | Specialized tools | Company-specific APIs, development tools |

### Server Configuration

```go
type MCPServerConfig struct {
    Command         string            `json:"command"`
    Args            []string          `json:"args,omitempty"`
    Environment     map[string]string `json:"env,omitempty"`
    WorkingDirectory string           `json:"cwd,omitempty"`
    Enabled         bool              `json:"enabled"`
}
```

## Advanced Usage

### Custom MCP Server Setup
```go
func setupCustomMCPServer(client *client.ClaudeCodeClient) error {
    // Database server configuration
    dbConfig := &types.MCPServerConfig{
        Command: "npx",
        Args:    []string{"@modelcontextprotocol/server-postgres"},
        Environment: map[string]string{
            "POSTGRES_URL": "postgresql://user:pass@localhost:5432/db",
            "NODE_ENV":     "production",
        },
        Enabled: true,
    }
    
    err := client.AddMCPServer(ctx, "database", dbConfig)
    if err != nil {
        return fmt.Errorf("failed to add database server: %w", err)
    }
    
    // Custom tool server
    customConfig := &types.MCPServerConfig{
        Command: "python",
        Args:    []string{"/path/to/custom-mcp-server.py"},
        Environment: map[string]string{
            "API_KEY":     os.Getenv("CUSTOM_API_KEY"),
            "LOG_LEVEL":   "info",
        },
        WorkingDirectory: "/opt/tools",
        Enabled:          true,
    }
    
    return client.AddMCPServer(ctx, "custom-tools", customConfig)
}
```

### MCP Server Validation
```go
func validateMCPServers(client *client.ClaudeCodeClient) error {
    servers := client.ListMCPServers()
    
    for name, config := range servers {
        if !config.Enabled {
            log.Printf("Server %s is disabled", name)
            continue
        }
        
        // Test server connectivity
        testRequest := &types.QueryRequest{
            Messages: []types.Message{
                {
                    Role:    types.RoleUser,
                    Content: fmt.Sprintf("Test connectivity to %s server", name),
                },
            },
        }
        
        _, err := client.Query(ctx, testRequest)
        if err != nil {
            log.Printf("Server %s failed connectivity test: %v", name, err)
        } else {
            log.Printf("Server %s is operational", name)
        }
    }
    
    return nil
}
```

### Dynamic MCP Management
```go
type MCPManager struct {
    client *client.ClaudeCodeClient
    active map[string]bool
}

func (m *MCPManager) EnableServer(name string) error {
    servers := m.client.ListMCPServers()
    config, exists := servers[name]
    if !exists {
        return fmt.Errorf("server %s not found", name)
    }
    
    config.Enabled = true
    m.active[name] = true
    
    return m.client.UpdateMCPServer(ctx, name, config)
}

func (m *MCPManager) DisableServer(name string) error {
    servers := m.client.ListMCPServers()
    config, exists := servers[name]
    if !exists {
        return fmt.Errorf("server %s not found", name)
    }
    
    config.Enabled = false
    m.active[name] = false
    
    return m.client.UpdateMCPServer(ctx, name, config)
}

func (m *MCPManager) GetActiveServers() []string {
    var active []string
    for name, isActive := range m.active {
        if isActive {
            active = append(active, name)
        }
    }
    return active
}
```

## Best Practices

### 1. Server Lifecycle Management
```go
// Setup servers during initialization
func initializeMCPServers(client *client.ClaudeCodeClient) error {
    // Add required servers
    requiredServers := map[string]*types.MCPServerConfig{
        "filesystem": {
            Command: "npx",
            Args:    []string{"@modelcontextprotocol/server-filesystem", "."},
            Enabled: true,
        },
        "git": {
            Command: "npx", 
            Args:    []string{"@modelcontextprotocol/server-git", "."},
            Enabled: true,
        },
    }
    
    for name, config := range requiredServers {
        if err := client.AddMCPServer(ctx, name, config); err != nil {
            log.Printf("Failed to add %s server: %v", name, err)
        }
    }
    
    return nil
}

// Cleanup during shutdown
func cleanupMCPServers(client *client.ClaudeCodeClient) {
    servers := client.ListMCPServers()
    for name := range servers {
        if err := client.RemoveMCPServer(ctx, name); err != nil {
            log.Printf("Failed to remove %s server: %v", name, err)
        }
    }
}
```

### 2. Error Handling
```go
func safeAddMCPServer(client *client.ClaudeCodeClient, name string, config *types.MCPServerConfig) error {
    // Check if server already exists
    existing := client.ListMCPServers()
    if _, exists := existing[name]; exists {
        log.Printf("Server %s already exists, updating configuration", name)
        return client.UpdateMCPServer(ctx, name, config)
    }
    
    // Validate configuration
    if err := validateMCPConfig(config); err != nil {
        return fmt.Errorf("invalid MCP config: %w", err)
    }
    
    // Add server
    return client.AddMCPServer(ctx, name, config)
}

func validateMCPConfig(config *types.MCPServerConfig) error {
    if config.Command == "" {
        return fmt.Errorf("command is required")
    }
    
    // Check if command exists
    _, err := exec.LookPath(config.Command)
    if err != nil {
        return fmt.Errorf("command not found: %s", config.Command)
    }
    
    return nil
}
```

### 3. Security Considerations
```go
func secureAddMCPServer(name string, config *types.MCPServerConfig) error {
    // Validate server source
    if !isAllowedServer(name, config.Command) {
        return fmt.Errorf("server not in allowlist: %s", name)
    }
    
    // Sanitize environment variables
    for key, value := range config.Environment {
        if containsSensitiveData(key, value) {
            return fmt.Errorf("sensitive data detected in environment: %s", key)
        }
    }
    
    // Restrict working directory
    if config.WorkingDirectory != "" {
        if !isAllowedDirectory(config.WorkingDirectory) {
            return fmt.Errorf("working directory not allowed: %s", config.WorkingDirectory)
        }
    }
    
    return claudeClient.AddMCPServer(ctx, name, config)
}

func isAllowedServer(name, command string) bool {
    allowedServers := map[string][]string{
        "filesystem": {"npx"},
        "git":        {"npx"},
        "database":   {"npx", "python"},
    }
    
    allowed, exists := allowedServers[name]
    if !exists {
        return false
    }
    
    for _, allowedCommand := range allowed {
        if command == allowedCommand {
            return true
        }
    }
    
    return false
}
```

## Integration Patterns

### Development Environment Setup
```go
func setupDevelopmentMCP(client *client.ClaudeCodeClient, projectPath string) error {
    configs := map[string]*types.MCPServerConfig{
        "project-fs": {
            Command: "npx",
            Args:    []string{"@modelcontextprotocol/server-filesystem", projectPath},
            Enabled: true,
        },
        "project-git": {
            Command: "npx",
            Args:    []string{"@modelcontextprotocol/server-git", projectPath},
            Enabled: true,
        },
    }
    
    for name, config := range configs {
        if err := client.AddMCPServer(ctx, name, config); err != nil {
            return fmt.Errorf("failed to setup %s: %w", name, err)
        }
    }
    
    return nil
}
```

### Production Environment
```go
func setupProductionMCP(client *client.ClaudeCodeClient) error {
    // Only enable essential, secure servers in production
    configs := map[string]*types.MCPServerConfig{
        "secure-db": {
            Command: "npx",
            Args:    []string{"@modelcontextprotocol/server-postgres"},
            Environment: map[string]string{
                "POSTGRES_URL": os.Getenv("SECURE_DB_URL"),
                "NODE_ENV":     "production",
            },
            Enabled: true,
        },
    }
    
    for name, config := range configs {
        if err := client.AddMCPServer(ctx, name, config); err != nil {
            return fmt.Errorf("failed to setup %s: %w", name, err)
        }
    }
    
    return nil
}
```

## Troubleshooting

### Common Issues

1. **MCP Server Not Found**
   ```bash
   # Install missing server
   npm install -g @modelcontextprotocol/server-filesystem
   
   # Verify installation
   npx @modelcontextprotocol/server-filesystem --version
   ```

2. **Permission Errors**
   ```bash
   # Check file permissions
   ls -la /path/to/mcp/server
   
   # Fix permissions if needed
   chmod +x /path/to/mcp/server
   ```

3. **Connection Issues**
   - Check network connectivity
   - Verify server configuration
   - Test server independently

### Debug MCP Issues
```go
// Enable MCP debugging
config := types.NewClaudeCodeConfig()
config.Debug = true
config.Environment = map[string]string{
    "MCP_DEBUG":     "true",
    "MCP_LOG_LEVEL": "debug",
}

// Test MCP server connectivity
func testMCPServer(name string) error {
    testQuery := &types.QueryRequest{
        Messages: []types.Message{
            {Role: types.RoleUser, Content: fmt.Sprintf("Test %s server capabilities", name)},
        },
    }
    
    response, err := claudeClient.Query(ctx, testQuery)
    if err != nil {
        return fmt.Errorf("MCP test failed: %w", err)
    }
    
    fmt.Printf("MCP test response: %s\n", response.Content[0].Text)
    return nil
}
```

### Health Checks
```go
func performMCPHealthCheck(client *client.ClaudeCodeClient) map[string]bool {
    servers := client.ListMCPServers()
    health := make(map[string]bool)
    
    for name, config := range servers {
        if !config.Enabled {
            health[name] = false
            continue
        }
        
        // Perform basic connectivity test
        err := testMCPServer(name)
        health[name] = (err == nil)
        
        if err != nil {
            log.Printf("Health check failed for %s: %v", name, err)
        }
    }
    
    return health
}
```

## Performance Optimization

### 1. Selective Server Enabling
```go
// Enable only servers needed for current task
func enableServersForTask(task string, client *client.ClaudeCodeClient) error {
    taskServers := map[string][]string{
        "file_analysis":  {"filesystem", "git"},
        "web_research":   {"web-search"},
        "data_analysis":  {"database", "filesystem"},
        "development":    {"filesystem", "git"},
    }
    
    servers := client.ListMCPServers()
    requiredServers := taskServers[task]
    
    for name, config := range servers {
        shouldEnable := contains(requiredServers, name)
        if config.Enabled != shouldEnable {
            config.Enabled = shouldEnable
            client.UpdateMCPServer(ctx, name, config)
        }
    }
    
    return nil
}
```

### 2. Server Connection Pooling
```go
type MCPConnectionPool struct {
    connections map[string]*MCPConnection
    mutex       sync.RWMutex
}

func (p *MCPConnectionPool) GetConnection(serverName string) (*MCPConnection, error) {
    p.mutex.RLock()
    conn, exists := p.connections[serverName]
    p.mutex.RUnlock()
    
    if exists && conn.IsHealthy() {
        return conn, nil
    }
    
    // Create new connection
    p.mutex.Lock()
    defer p.mutex.Unlock()
    
    newConn, err := createMCPConnection(serverName)
    if err != nil {
        return nil, err
    }
    
    p.connections[serverName] = newConn
    return newConn, nil
}
```

## Next Steps

After understanding MCP integration, explore:
- [Command Execution](../command_execution/) - Using MCP-enhanced commands
- [Advanced Client](../advanced_client/) - Client optimization with MCP
- [Sync Queries](../sync_queries/) - MCP-enhanced queries
- [Streaming Queries](../streaming_queries/) - Real-time MCP integration

## Related Documentation

- [MCP Protocol Documentation](https://modelcontextprotocol.io/)
- [MCP Types](../../pkg/types/mcp.go)
- [Client Package](../../pkg/client/)
- [Integration Guide](../../docs/mcp-integration.md)