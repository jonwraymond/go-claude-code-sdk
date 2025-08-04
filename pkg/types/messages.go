// Package claude_code_sdk provides a Go SDK for interacting with Claude Code.
//
// This SDK mirrors the Python SDK functionality, providing simple interfaces
// for querying Claude Code and handling streaming responses.
package types

import (
	"encoding/json"
	"path/filepath"
)

// PermissionMode defines the permission modes for tool execution.
type PermissionMode string

const (
	// PermissionModeDefault prompts for dangerous tools
	PermissionModeDefault PermissionMode = "default"
	// PermissionModeAcceptEdits auto-accepts file edits
	PermissionModeAcceptEdits PermissionMode = "acceptEdits"
	// PermissionModeBypassPermissions allows all tools (use with caution)
	PermissionModeBypassPermissions PermissionMode = "bypassPermissions"
)

// McpServerConfig represents an MCP server configuration.
type McpServerConfig map[string]interface{}

// TextBlock represents a text content block.
type TextBlock struct {
	Text string `json:"text"`
}

// ToolUseBlock represents a tool use content block.
type ToolUseBlock struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// ToolResultBlock represents a tool result content block.
type ToolResultBlock struct {
	ToolUseID string      `json:"tool_use_id"`
	Content   interface{} `json:"content,omitempty"`
	IsError   *bool       `json:"is_error,omitempty"`
}

// ContentBlock is an interface that represents any content block type.
type ContentBlock interface {
	contentBlock()
}

func (TextBlock) contentBlock()       {}
func (ToolUseBlock) contentBlock()    {}
func (ToolResultBlock) contentBlock() {}

// UserMessage represents a user message.
type UserMessage struct {
	Content interface{} `json:"content"` // Can be string or []ContentBlock
}

// AssistantMessage represents an assistant message with content blocks.
type AssistantMessage struct {
	Content []ContentBlock `json:"content"`
}

// SystemMessage represents a system message with metadata.
type SystemMessage struct {
	Subtype string                 `json:"subtype"`
	Data    map[string]interface{} `json:"data"`
}

// ResultMessage represents a result message with cost and usage information.
type ResultMessage struct {
	Subtype       string                 `json:"subtype"`
	DurationMs    int                    `json:"duration_ms"`
	DurationAPIMs int                    `json:"duration_api_ms"`
	IsError       bool                   `json:"is_error"`
	NumTurns      int                    `json:"num_turns"`
	SessionID     string                 `json:"session_id"`
	TotalCostUSD  *float64               `json:"total_cost_usd,omitempty"`
	Usage         map[string]interface{} `json:"usage,omitempty"`
	Result        *string                `json:"result,omitempty"`
}

// Message is an interface that represents any message type.
type Message interface {
	message()
}

func (*UserMessage) message()      {}
func (*AssistantMessage) message() {}
func (*SystemMessage) message()    {}
func (*ResultMessage) message()    {}

// ClaudeCodeOptions represents query options for the Claude SDK.
type ClaudeCodeOptions struct {
	AllowedTools             []string                   `json:"allowed_tools,omitempty"`
	MaxThinkingTokens        int                        `json:"max_thinking_tokens,omitempty"`
	SystemPrompt             *string                    `json:"system_prompt,omitempty"`
	AppendSystemPrompt       *string                    `json:"append_system_prompt,omitempty"`
	MCPTools                 []string                   `json:"mcp_tools,omitempty"`
	MCPServers               map[string]McpServerConfig `json:"mcp_servers,omitempty"`
	PermissionMode           *PermissionMode            `json:"permission_mode,omitempty"`
	ContinueConversation     bool                       `json:"continue_conversation,omitempty"`
	Resume                   *string                    `json:"resume,omitempty"`
	MaxTurns                 *int                       `json:"max_turns,omitempty"`
	DisallowedTools          []string                   `json:"disallowed_tools,omitempty"`
	Model                    *string                    `json:"model,omitempty"`
	PermissionPromptToolName *string                    `json:"permission_prompt_tool_name,omitempty"`
	CWD                      *string                    `json:"cwd,omitempty"`
	Settings                 *string                    `json:"settings,omitempty"`
	AddDirs                  []string                   `json:"add_dirs,omitempty"`
}

// NewClaudeCodeOptions creates a new ClaudeCodeOptions with default values.
func NewClaudeCodeOptions() *ClaudeCodeOptions {
	return &ClaudeCodeOptions{
		AllowedTools:         []string{},
		MaxThinkingTokens:    8000,
		MCPTools:             []string{},
		MCPServers:           make(map[string]McpServerConfig),
		ContinueConversation: false,
		DisallowedTools:      []string{},
		AddDirs:              []string{},
	}
}

// SetCWD sets the current working directory, accepting string or path.
func (o *ClaudeCodeOptions) SetCWD(cwd interface{}) {
	switch v := cwd.(type) {
	case string:
		o.CWD = &v
	case *string:
		o.CWD = v
	default:
		// Try to convert to string for filepath compatibility
		cwdStr := filepath.Clean(v.(string))
		o.CWD = &cwdStr
	}
}

// UnmarshalContentBlock unmarshals a ContentBlock from JSON.
func UnmarshalContentBlock(data []byte) (ContentBlock, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	// Determine type based on fields present
	if _, hasText := raw["text"]; hasText {
		var block TextBlock
		err := json.Unmarshal(data, &block)
		return block, err
	} else if name, hasName := raw["name"]; hasName && name != "" {
		if _, hasInput := raw["input"]; hasInput {
			var block ToolUseBlock
			err := json.Unmarshal(data, &block)
			return block, err
		}
	} else if _, hasToolUseID := raw["tool_use_id"]; hasToolUseID {
		var block ToolResultBlock
		err := json.Unmarshal(data, &block)
		return block, err
	}

	// Default to text block
	var block TextBlock
	err := json.Unmarshal(data, &block)
	return block, err
}
