package transport

import (
	"fmt"
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

// McpServerConfig represents MCP server configuration.
type McpServerConfig map[string]interface{}

// ClaudeCodeOptions represents query options for the Claude SDK.
type ClaudeCodeOptions struct {
	AllowedTools             []string                    `json:"allowed_tools,omitempty"`
	MaxThinkingTokens        int                         `json:"max_thinking_tokens,omitempty"`
	SystemPrompt             *string                     `json:"system_prompt,omitempty"`
	AppendSystemPrompt       *string                     `json:"append_system_prompt,omitempty"`
	MCPTools                 []string                    `json:"mcp_tools,omitempty"`
	MCPServers               map[string]McpServerConfig  `json:"mcp_servers,omitempty"`
	PermissionMode           *PermissionMode             `json:"permission_mode,omitempty"`
	ContinueConversation     bool                        `json:"continue_conversation,omitempty"`
	Resume                   *string                     `json:"resume,omitempty"`
	MaxTurns                 *int                        `json:"max_turns,omitempty"`
	DisallowedTools          []string                    `json:"disallowed_tools,omitempty"`
	Model                    *string                     `json:"model,omitempty"`
	PermissionPromptToolName *string                     `json:"permission_prompt_tool_name,omitempty"`
	CWD                      *string                     `json:"cwd,omitempty"`
	Settings                 *string                     `json:"settings,omitempty"`
	AddDirs                  []string                    `json:"add_dirs,omitempty"`
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

// ClaudeSDKError is the base error type for Claude SDK errors.
type ClaudeSDKError struct {
	Message string
	Cause   error
}

func (e *ClaudeSDKError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("ClaudeSDKError: %s (caused by: %v)", e.Message, e.Cause)
	}
	return fmt.Sprintf("ClaudeSDKError: %s", e.Message)
}

func (e *ClaudeSDKError) Unwrap() error {
	return e.Cause
}

// NewClaudeSDKError creates a new ClaudeSDKError.
func NewClaudeSDKError(message string, cause error) *ClaudeSDKError {
	return &ClaudeSDKError{
		Message: message,
		Cause:   cause,
	}
}

// CLIConnectionError represents errors connecting to Claude CLI.
type CLIConnectionError struct {
	*ClaudeSDKError
}

// NewCLIConnectionError creates a new CLIConnectionError.
func NewCLIConnectionError(message string, cause error) *CLIConnectionError {
	return &CLIConnectionError{
		ClaudeSDKError: NewClaudeSDKError(message, cause),
	}
}

// CLINotFoundError represents errors when Claude CLI is not found.
type CLINotFoundError struct {
	*ClaudeSDKError
}

// NewCLINotFoundError creates a new CLINotFoundError.
func NewCLINotFoundError(message string, cause error) *CLINotFoundError {
	return &CLINotFoundError{
		ClaudeSDKError: NewClaudeSDKError(message, cause),
	}
}

// ProcessError represents errors during subprocess execution.
type ProcessError struct {
	*ClaudeSDKError
	ExitCode int
}

// NewProcessError creates a new ProcessError.
func NewProcessError(message string, exitCode int, cause error) *ProcessError {
	return &ProcessError{
		ClaudeSDKError: NewClaudeSDKError(message, cause),
		ExitCode:       exitCode,
	}
}

// CLIJSONDecodeError represents JSON decoding errors.
type CLIJSONDecodeError struct {
	*ClaudeSDKError
	RawData string
}

// NewCLIJSONDecodeError creates a new CLIJSONDecodeError.
func NewCLIJSONDecodeError(message string, rawData string, cause error) *CLIJSONDecodeError {
	return &CLIJSONDecodeError{
		ClaudeSDKError: NewClaudeSDKError(message, cause),
		RawData:        rawData,
	}
}

