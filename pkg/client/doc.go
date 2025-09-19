/*
Package client provides the main client implementation for the Claude Code Go SDK.

The client package wraps the Claude Code CLI tool, providing idiomatic Go interfaces
for AI-powered coding assistance. It uses a subprocess-based architecture to interact
with the Claude Code command-line interface, enabling streaming responses, session
management, tool execution, MCP integration, and comprehensive error handling.

This package is the primary entry point for interacting with Claude Code programmatically,
offering both high-level convenience methods and low-level control for advanced use cases.

# Architecture

The client follows a subprocess model where all interactions with Claude Code happen
through the CLI binary. This design ensures compatibility with Claude Code's native
features while providing type-safe Go interfaces.

Key components:
  - ClaudeCodeClient: Main client for executing queries and managing the CLI subprocess
  - SessionManager: Handles conversation persistence using Claude Code's --session flag
  - ToolManager: Manages built-in and MCP-provided tools for code operations
  - MCPManager: Integrates Model Context Protocol servers for extended capabilities
  - ProjectContextManager: Detects and analyzes project structure, languages, and frameworks

# Basic Usage

Create a client and execute a simple query:

	config := &types.ClaudeCodeConfig{
		WorkingDirectory: "/path/to/project",
		SessionID:        "my-session",
		Model:           types.ModelClaude35Sonnet,
		AuthMethod:      types.AuthTypeAPIKey,
		APIKey:          os.Getenv("ANTHROPIC_API_KEY"),
	}

	client, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Simple query
	response, err := client.Query(ctx, &types.QueryRequest{
		Messages: []types.Message{
			{Role: types.RoleUser, Content: "Analyze this codebase"},
		},
		MaxTokens: 1000,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response: %s\n", response.GetTextContent())

# Streaming Responses

The SDK supports streaming responses for real-time output:

	stream, err := client.QueryStream(ctx, &types.QueryRequest{
		Messages: []types.Message{
			{Role: types.RoleUser, Content: "Write a Go HTTP server"},
		},
		MaxTokens: 2000,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	// Process streaming chunks
	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}
		if chunk.Done {
			break
		}
		fmt.Print(chunk.Content)
	}

	// Or collect the complete response
	response, err := stream.Collect()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Complete response: %s\n", response.GetTextContent())

# Session Management

Sessions provide conversation persistence with UUID validation:

	// Generate a proper UUID for session ID
	sessionID := client.GenerateSessionID()

	// Create a new session
	session, err := client.CreateSession(ctx, sessionID)
	if err != nil {
		log.Fatal(err)
	}

	// Use session for conversation
	response, err := session.Query(ctx, &types.QueryRequest{
		Messages: []types.Message{
			{Role: types.RoleUser, Content: "Hello Claude"},
		},
	})

	// Session automatically maintains conversation context
	// Subsequent queries will include previous context

	// Get existing session
	existingSession, err := client.GetSession(sessionID)

	// List all active sessions
	sessionIDs := client.ListSessions()

# Tool System

Claude Code provides various tools for file operations and code analysis:

	// Discover all available tools
	tools, err := client.DiscoverTools(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, tool := range tools {
		fmt.Printf("Tool: %s - %s\n", tool.Name, tool.Description)
	}

	// Execute a tool directly
	result, err := client.ExecuteTool(ctx, &ClaudeCodeTool{
		Name: "read_file",
		Arguments: map[string]any{
			"path": "main.go",
		},
	})

	// Get a specific tool definition
	readTool, err := client.GetTool("read_file")
	if err != nil {
		log.Fatal(err)
	}

	// List all available tools
	allTools := client.ListTools()

# MCP Server Integration

The Model Context Protocol extends Claude Code with additional capabilities:

	// Add an MCP server
	err := client.AddMCPServer(ctx, "filesystem", &types.MCPServerConfig{
		Command: "npx",
		Args:    []string{"@modelcontextprotocol/server-filesystem", "/path/to/project"},
		Env:     map[string]string{"NODE_ENV": "production"},
		Enabled: true,
	})

	// Setup common MCP servers automatically
	err = client.SetupCommonMCPServers(ctx)

	// Enable/disable servers
	err = client.EnableMCPServer(ctx, "filesystem")
	err = client.DisableMCPServer(ctx, "filesystem")

	// List all configured servers
	servers := client.ListMCPServers()
	for name, config := range servers {
		fmt.Printf("Server: %s, Enabled: %v\n", name, config.Enabled)
	}

	// Get specific server config
	config, err := client.GetMCPServer("filesystem")

	// Remove a server
	err = client.RemoveMCPServer(ctx, "filesystem")

# Project Context

The SDK automatically detects and analyzes project information:

	// Get basic project context
	context, err := client.GetProjectContext(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Working directory: %s\n", context.WorkingDirectory)

	// Get enhanced project analysis
	enhanced, err := client.GetEnhancedProjectContext(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// Returns language, framework, dependencies, architecture patterns, etc.

	// Change working directory
	err = client.SetWorkingDirectory(ctx, "/new/project/path")

	// Invalidate cached context when project changes
	client.InvalidateProjectContextCache()

	// Configure cache duration
	client.SetProjectContextCacheDuration(5 * time.Minute)

	// Get cache information
	cacheInfo := client.GetProjectContextCacheInfo()
	fmt.Printf("Cache info: %+v\n", cacheInfo)

# Error Handling

The client provides detailed error types for different scenarios:

	result, err := client.QueryMessagesSync(ctx, "query", nil)
	if err != nil {
		var claudeErr *errors.ClaudeCodeError
		if errors.As(err, &claudeErr) {
			if claudeErr.IsRetryable() {
				// Retry logic
			}
		}
	}

# Subprocess Management

The client manages the Claude Code CLI subprocess lifecycle:
  - Automatic process creation and cleanup
  - Proper signal handling and graceful shutdown
  - Resource leak prevention
  - Concurrent operation safety

All subprocess operations are handled internally, providing a clean API while
ensuring proper resource management.
*/
package client
