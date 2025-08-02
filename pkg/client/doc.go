/*
Package client provides the main client implementation for the Claude Code Go SDK.

The client package wraps the Claude Code CLI tool, providing idiomatic Go interfaces
for AI-powered coding assistance. It uses a subprocess-based architecture to interact
with the Claude Code command-line interface, enabling streaming responses, session
management, and tool execution.

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

	config := types.NewClaudeCodeConfig()
	config.APIKey = os.Getenv("ANTHROPIC_API_KEY")

	client, err := client.NewClaudeCodeClient(config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Simple query
	result, err := client.QueryMessagesSync(ctx, "Explain this code", nil)

# Streaming Responses

The SDK supports streaming responses through channels:

	messages, err := client.QueryMessages(ctx, "Write a function", nil)
	if err != nil {
		log.Fatal(err)
	}

	for msg := range messages {
		fmt.Printf("[%s]: %s\n", msg.Role, msg.GetText())
	}

# Session Management

Sessions provide conversation persistence:

	session, err := client.Sessions().CreateSession(ctx, "my-session")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	// Multiple queries in the same conversation context
	session.Query(ctx, &types.QueryRequest{...})

# Tool System

Claude Code provides various tools for file operations and code analysis:

	// Execute a tool directly
	result, err := client.Tools().ExecuteTool(ctx, &ClaudeCodeTool{
		Name: "read_file",
		Arguments: map[string]any{
			"path": "main.go",
		},
	})

	// Tools are automatically available in conversations
	options := &QueryOptions{
		AllowedTools: []string{"read_file", "write_file", "search_code"},
	}

# MCP Server Integration

The Model Context Protocol extends Claude Code with additional capabilities:

	// Register an MCP server
	err := client.MCP().RegisterServer("sqlite", &types.MCPServerConfig{
		Command: "sqlite-mcp-server",
		Args:    []string{"./database.db"},
	})

	// Start the server
	err = client.MCP().StartServer(ctx, "sqlite")

# Project Context

The SDK automatically detects project information:

	context, err := client.ProjectContext().GetEnhancedProjectContext(ctx)
	// Returns language, framework, dependencies, architecture patterns, etc.

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
