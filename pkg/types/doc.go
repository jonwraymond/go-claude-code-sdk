// Package types defines the message types and content blocks used by the Claude Code SDK.
//
// This package contains all the type definitions needed for interacting with Claude Code,
// including message types (UserMessage, AssistantMessage, SystemMessage) and content
// block types (TextBlock, ToolUseBlock, ToolResultBlock).
//
// Message Types:
//
//   - UserMessage: Represents a message from the user
//   - AssistantMessage: Represents a message from Claude
//   - SystemMessage: Represents system-level messages
//   - ResultMessage: Represents the final result of a conversation
//
// Content Blocks:
//
//   - TextBlock: Plain text content
//   - ToolUseBlock: Tool invocation by Claude
//   - ToolResultBlock: Results from tool execution
//
// The types in this package mirror the structure of the official Python SDK,
// providing type-safe representations of Claude Code's communication protocol.
package types
