package claudecode

import "github.com/jonwraymond/go-claude-code-sdk/pkg/types"

// Type aliases for public API message types
type Message = types.Message
type UserMessage = types.UserMessage
type AssistantMessage = types.AssistantMessage
type SystemMessage = types.SystemMessage
type ResultMessage = types.ResultMessage

// Type aliases for content blocks
type ContentBlock = types.ContentBlock
type TextBlock = types.TextBlock
type ToolUseBlock = types.ToolUseBlock
type ToolResultBlock = types.ToolResultBlock

// Type aliases for MCP server configuration
type McpServerConfig = types.McpServerConfig