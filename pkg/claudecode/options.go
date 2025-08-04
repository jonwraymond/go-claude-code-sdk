package claudecode

import (
	"path/filepath"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// ClaudeCodeOptions represents configuration options for Claude Code.
// This is the public API version of the options.
type ClaudeCodeOptions = types.ClaudeCodeOptions

// PermissionMode represents how Claude Code handles file edit permissions.
type PermissionMode = types.PermissionMode

// Permission mode constants
const (
	PermissionModeDefault         = types.PermissionModeDefault
	PermissionModeAcceptEdits     = types.PermissionModeAcceptEdits
	PermissionModeBypassPermission = types.PermissionModeBypassPermissions
)

// NewClaudeCodeOptions creates a new options instance with defaults.
func NewClaudeCodeOptions() *ClaudeCodeOptions {
	return types.NewClaudeCodeOptions()
}

// WithSystemPrompt sets the system prompt for the conversation.
func WithSystemPrompt(prompt string) func(*ClaudeCodeOptions) {
	return func(opts *ClaudeCodeOptions) {
		opts.SystemPrompt = &prompt
	}
}

// WithMaxTurns sets the maximum number of conversation turns.
func WithMaxTurns(turns int) func(*ClaudeCodeOptions) {
	return func(opts *ClaudeCodeOptions) {
		opts.MaxTurns = &turns
	}
}

// WithAllowedTools sets the list of allowed tools.
func WithAllowedTools(tools []string) func(*ClaudeCodeOptions) {
	return func(opts *ClaudeCodeOptions) {
		opts.AllowedTools = tools
	}
}

// WithPermissionMode sets how file edits are handled.
func WithPermissionMode(mode PermissionMode) func(*ClaudeCodeOptions) {
	return func(opts *ClaudeCodeOptions) {
		opts.PermissionMode = &mode
	}
}

// WithCWD sets the current working directory.
func WithCWD(dir string) func(*ClaudeCodeOptions) {
	return func(opts *ClaudeCodeOptions) {
		absPath, _ := filepath.Abs(dir)
		opts.CWD = &absPath
	}
}