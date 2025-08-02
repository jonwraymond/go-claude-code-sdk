/*
Package types provides core type definitions for the Claude Code Go SDK.

This package contains all the fundamental types used throughout the SDK, including
configuration, messages, commands, and responses. These types are designed to map
closely to Claude Code CLI's data structures while providing idiomatic Go interfaces.

# Configuration Types

The package provides configuration types for the Claude Code client:

	config := types.NewClaudeCodeConfig()
	config.APIKey = "your-api-key"
	config.WorkingDirectory = "/path/to/project"
	config.Model = "claude-3-opus"

Configuration supports environment variables and builder patterns:

	config := types.NewConfigBuilder().
		WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")).
		WithTimeout(30 * time.Second).
		Build()

# Message Types

Messages are the core communication unit with Claude Code:

	// Create different message types
	userMsg := types.NewTextMessage(types.MessageRoleUser, "Hello Claude")
	assistantMsg := types.NewTextMessage(types.MessageRoleAssistant, "Hello! How can I help?")
	
	// Messages with tool usage
	toolMsg := types.NewToolUseMessage("tool-123", "read_file", map[string]interface{}{
		"path": "main.go",
	})

Messages support content blocks for structured data:

	msg := &types.Message{
		Role: types.MessageRoleAssistant,
		Content: []types.ContentBlock{
			types.NewTextBlock("Here's the file content:"),
			types.NewToolUseBlock("123", "read_file", args),
		},
	}

# Command Types

Commands represent Claude Code CLI operations:

	cmd := &types.Command{
		Type: types.CommandAnalyze,
		Args: []string{"src/"},
		Options: map[string]interface{}{
			"include_tests": true,
		},
	}

Supported command types include:
  - File operations: read, write, edit, search
  - Code operations: analyze, explain, refactor, test
  - Git operations: status, commit, diff, log
  - Project operations: build, run, install, clean

# Query Types

Query requests and responses for Claude Code interactions:

	request := &types.QueryRequest{
		Messages: []types.Message{
			{Role: types.MessageRoleUser, Content: "Explain this code"},
		},
		MaxTokens: 1000,
		Temperature: 0.7,
	}

	response := &types.QueryResponse{
		Content: "Here's the explanation...",
		TokensUsed: 450,
		Model: "claude-3-opus",
	}

# Tool Types

Tool definitions for Claude Code's built-in and MCP tools:

	tool := &types.Tool{
		Name: "search_code",
		Description: "Search for code patterns",
		InputSchema: types.ToolInputSchema{
			Type: "object",
			Properties: map[string]types.ToolProperty{
				"pattern": {Type: "string", Description: "Search pattern"},
				"path": {Type: "string", Description: "Directory to search"},
			},
			Required: []string{"pattern"},
		},
	}

# MCP Server Types

Configuration for Model Context Protocol servers:

	mcpConfig := &types.MCPServerConfig{
		Name: "sqlite",
		Command: "sqlite-mcp-server",
		Args: []string{"database.db"},
		Environment: map[string]string{
			"SQLITE_READONLY": "false",
		},
	}

# Project Context Types

Project analysis and detection results:

	context := &types.ProjectContext{
		WorkingDirectory: "/path/to/project",
		ProjectName: "my-app",
		Language: "Go",
		Framework: "gin",
		Dependencies: map[string]string{
			"github.com/gin-gonic/gin": "v1.9.1",
		},
	}

# Error Constants

The package defines error categories and codes used throughout the SDK:

	// Error categories
	types.CategoryNetwork    // Network-related errors
	types.CategoryValidation // Input validation errors
	types.CategoryAuth       // Authentication errors
	types.CategoryInternal   // Internal SDK errors

# Type Safety

All types are designed with Go idioms in mind:
  - Strong typing for all fields
  - Validation methods where appropriate
  - JSON marshaling/unmarshaling support
  - Builder patterns for complex types
  - Immutability where it makes sense

The types package ensures type safety across the entire SDK while maintaining
flexibility for Claude Code's dynamic nature.
*/
package types
