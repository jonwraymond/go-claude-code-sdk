package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// Re-export types from types package for convenience
type CommandType = types.CommandType
type Command = types.Command
type CommandResult = types.CommandResult

// Re-export command type constants from types package
const (
	// File operations
	CommandRead   = types.CommandRead
	CommandWrite  = types.CommandWrite
	CommandEdit   = types.CommandEdit
	CommandSearch = types.CommandSearch

	// Code operations
	CommandAnalyze  = types.CommandAnalyze
	CommandExplain  = types.CommandExplain
	CommandRefactor = types.CommandRefactor
	CommandTest     = types.CommandTest
	CommandDebug    = types.CommandDebug

	// Git operations
	CommandGitStatus = types.CommandGitStatus
	CommandGitCommit = types.CommandGitCommit
	CommandGitDiff   = types.CommandGitDiff
	CommandGitLog    = types.CommandGitLog

	// Project operations
	CommandBuild   = types.CommandBuild
	CommandRun     = types.CommandRun
	CommandInstall = types.CommandInstall
	CommandClean   = types.CommandClean

	// Session operations
	CommandHistory = types.CommandHistory
	CommandClear   = types.CommandClear
	CommandSave    = types.CommandSave
	CommandLoad    = types.CommandLoad
)

// CommandExecutor provides command execution capabilities for Claude Code
type CommandExecutor struct {
	client *ClaudeCodeClient
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor(client *ClaudeCodeClient) *CommandExecutor {
	return &CommandExecutor{
		client: client,
	}
}

// ExecuteCommand executes a Claude Code command
func (ce *CommandExecutor) ExecuteCommand(ctx context.Context, cmd *Command) (*CommandResult, error) {
	if cmd == nil {
		return &CommandResult{
			Success: false,
			Error:   "command cannot be nil",
		}, nil
	}

	// Build the command prompt
	prompt, err := ce.buildCommandPrompt(cmd)
	if err != nil {
		return &CommandResult{
			Command: cmd,
			Success: false,
			Error:   fmt.Sprintf("failed to build command prompt: %v", err),
		}, nil
	}

	// Execute via Claude Code client
	request := &types.QueryRequest{
		Messages: []types.Message{
			{Role: types.RoleUser, Content: prompt},
		},
		MaxTokens: 4000,
	}

	response, err := ce.client.Query(ctx, request)
	if err != nil {
		return &CommandResult{
			Command: cmd,
			Success: false,
			Error:   fmt.Sprintf("command execution failed: %v", err),
		}, nil
	}

	// Parse response
	output := ce.ExtractTextContent(response.Content)

	return &CommandResult{
		Command: cmd,
		Success: true,
		Output:  output,
		Metadata: map[string]interface{}{
			"stop_reason": response.StopReason,
		},
	}, nil
}

// ExecuteSlashCommand executes a Claude Code slash command (e.g., "/read file.go")
func (ce *CommandExecutor) ExecuteSlashCommand(ctx context.Context, slashCommand string) (*CommandResult, error) {
	cmd, err := ce.ParseSlashCommand(slashCommand)
	if err != nil {
		return &CommandResult{
			Success: false,
			Error:   fmt.Sprintf("failed to parse slash command: %v", err),
		}, nil
	}

	return ce.ExecuteCommand(ctx, cmd)
}

// buildCommandPrompt constructs a prompt for the given command
func (ce *CommandExecutor) buildCommandPrompt(cmd *Command) (string, error) {
	var prompt strings.Builder

	switch cmd.Type {
	case CommandRead:
		prompt.WriteString("Please read the file")
		if len(cmd.Args) > 0 {
			prompt.WriteString(fmt.Sprintf(" '%s'", cmd.Args[0]))
		}
		if summarize, ok := cmd.Options["summarize"].(bool); ok && summarize {
			prompt.WriteString(" and provide a summary")
		}

	case CommandWrite:
		if len(cmd.Args) < 2 {
			return "", fmt.Errorf("write command requires filename and content")
		}
		prompt.WriteString(fmt.Sprintf("Please write to file '%s' with the following content:\n%s", cmd.Args[0], cmd.Args[1]))

	case CommandEdit:
		if len(cmd.Args) < 1 {
			return "", fmt.Errorf("edit command requires filename")
		}
		prompt.WriteString(fmt.Sprintf("Please edit the file '%s'", cmd.Args[0]))
		if len(cmd.Args) > 1 {
			prompt.WriteString(fmt.Sprintf(" with these changes: %s", cmd.Args[1]))
		}

	case CommandSearch:
		if len(cmd.Args) < 1 {
			return "", fmt.Errorf("search command requires search term")
		}
		prompt.WriteString(fmt.Sprintf("Please search for '%s' in the codebase", cmd.Args[0]))
		if filePattern, ok := cmd.Options["pattern"].(string); ok {
			prompt.WriteString(fmt.Sprintf(" in files matching pattern '%s'", filePattern))
		}

	case CommandAnalyze:
		prompt.WriteString("Please analyze the codebase")
		if len(cmd.Args) > 0 {
			prompt.WriteString(fmt.Sprintf(" focusing on '%s'", cmd.Args[0]))
		}
		if depth, ok := cmd.Options["depth"].(string); ok {
			prompt.WriteString(fmt.Sprintf(" with %s analysis", depth))
		}

	case CommandExplain:
		if len(cmd.Args) < 1 {
			return "", fmt.Errorf("explain command requires code or concept to explain")
		}
		prompt.WriteString(fmt.Sprintf("Please explain '%s'", cmd.Args[0]))
		if detail, ok := cmd.Options["detail"].(string); ok {
			prompt.WriteString(fmt.Sprintf(" with %s level of detail", detail))
		}

	case CommandRefactor:
		if len(cmd.Args) < 1 {
			return "", fmt.Errorf("refactor command requires filename or code")
		}
		prompt.WriteString(fmt.Sprintf("Please refactor '%s'", cmd.Args[0]))
		if approach, ok := cmd.Options["approach"].(string); ok {
			prompt.WriteString(fmt.Sprintf(" using %s approach", approach))
		}

	case CommandTest:
		prompt.WriteString("Please run tests")
		if len(cmd.Args) > 0 {
			prompt.WriteString(fmt.Sprintf(" for '%s'", cmd.Args[0]))
		}
		if testType, ok := cmd.Options["type"].(string); ok {
			prompt.WriteString(fmt.Sprintf(" (%s tests)", testType))
		}

	case CommandDebug:
		if len(cmd.Args) < 1 {
			return "", fmt.Errorf("debug command requires issue description")
		}
		prompt.WriteString(fmt.Sprintf("Please debug the issue: %s", cmd.Args[0]))

	case CommandGitStatus:
		prompt.WriteString("Please show git status")

	case CommandGitCommit:
		prompt.WriteString("Please create a git commit")
		if len(cmd.Args) > 0 {
			prompt.WriteString(fmt.Sprintf(" with message: %s", cmd.Args[0]))
		}

	case CommandGitDiff:
		prompt.WriteString("Please show git diff")
		if len(cmd.Args) > 0 {
			prompt.WriteString(fmt.Sprintf(" for '%s'", cmd.Args[0]))
		}

	case CommandGitLog:
		prompt.WriteString("Please show git log")
		if limit, ok := cmd.Options["limit"].(int); ok {
			prompt.WriteString(fmt.Sprintf(" (last %d commits)", limit))
		}

	case CommandBuild:
		prompt.WriteString("Please build the project")
		if len(cmd.Args) > 0 {
			prompt.WriteString(fmt.Sprintf(" target: %s", cmd.Args[0]))
		}

	case CommandRun:
		prompt.WriteString("Please run the project")
		if len(cmd.Args) > 0 {
			prompt.WriteString(fmt.Sprintf(" with args: %s", strings.Join(cmd.Args, " ")))
		}

	case CommandInstall:
		prompt.WriteString("Please install dependencies")
		if len(cmd.Args) > 0 {
			prompt.WriteString(fmt.Sprintf(" for: %s", strings.Join(cmd.Args, ", ")))
		}

	case CommandClean:
		prompt.WriteString("Please clean the project")

	case CommandHistory:
		prompt.WriteString("Please show conversation history")
		if limit, ok := cmd.Options["limit"].(int); ok {
			prompt.WriteString(fmt.Sprintf(" (last %d entries)", limit))
		}

	case CommandClear:
		prompt.WriteString("Please clear the conversation history")

	case CommandSave:
		prompt.WriteString("Please save the current session")
		if len(cmd.Args) > 0 {
			prompt.WriteString(fmt.Sprintf(" as '%s'", cmd.Args[0]))
		}

	case CommandLoad:
		if len(cmd.Args) < 1 {
			return "", fmt.Errorf("load command requires session name")
		}
		prompt.WriteString(fmt.Sprintf("Please load session '%s'", cmd.Args[0]))

	default:
		return "", fmt.Errorf("unknown command type: %s", cmd.Type)
	}

	// Add context if provided
	if len(cmd.Context) > 0 {
		prompt.WriteString("\n\nAdditional context:")
		for key, value := range cmd.Context {
			prompt.WriteString(fmt.Sprintf("\n- %s: %v", key, value))
		}
	}

	return prompt.String(), nil
}

// ParseSlashCommand parses a slash command string into a Command struct
func (ce *CommandExecutor) ParseSlashCommand(slashCommand string) (*Command, error) {
	if !strings.HasPrefix(slashCommand, "/") {
		return nil, fmt.Errorf("slash command must start with '/'")
	}

	// Remove leading slash and split into parts
	parts := strings.Fields(slashCommand[1:])
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	cmdType := CommandType(parts[0])
	args := parts[1:]

	// Parse options (simple key=value parsing)
	var cleanArgs []string
	options := make(map[string]interface{})

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			kv := strings.SplitN(arg, "=", 2)
			if len(kv) == 2 {
				options[kv[0]] = kv[1]
			}
		} else {
			cleanArgs = append(cleanArgs, arg)
		}
	}

	return &Command{
		Type:    cmdType,
		Args:    cleanArgs,
		Options: options,
	}, nil
}

// ExtractTextContent extracts text content from content blocks
func (ce *CommandExecutor) ExtractTextContent(content []types.ContentBlock) string {
	var text strings.Builder
	for _, block := range content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	return text.String()
}

// Common command builder methods

// ReadFile creates a command to read a file
func ReadFile(filename string, options ...func(*Command)) *Command {
	cmd := &Command{
		Type:    CommandRead,
		Args:    []string{filename},
		Options: make(map[string]interface{}),
	}

	for _, opt := range options {
		opt(cmd)
	}

	return cmd
}

// WriteFile creates a command to write to a file
func WriteFile(filename, content string, options ...func(*Command)) *Command {
	cmd := &Command{
		Type:    CommandWrite,
		Args:    []string{filename, content},
		Options: make(map[string]interface{}),
	}

	for _, opt := range options {
		opt(cmd)
	}

	return cmd
}

// AnalyzeCode creates a command to analyze code
func AnalyzeCode(target string, options ...func(*Command)) *Command {
	cmd := &Command{
		Type:    CommandAnalyze,
		Args:    []string{target},
		Options: make(map[string]interface{}),
	}

	for _, opt := range options {
		opt(cmd)
	}

	return cmd
}

// SearchCode creates a command to search in codebase
func SearchCode(query string, options ...func(*Command)) *Command {
	cmd := &Command{
		Type:    CommandSearch,
		Args:    []string{query},
		Options: make(map[string]interface{}),
	}

	for _, opt := range options {
		opt(cmd)
	}

	return cmd
}

// GitStatus creates a command to show git status
func GitStatus(options ...func(*Command)) *Command {
	cmd := &Command{
		Type:    CommandGitStatus,
		Options: make(map[string]interface{}),
	}

	for _, opt := range options {
		opt(cmd)
	}

	return cmd
}

// Command option builders

// WithSummary adds summarize option to a command
func WithSummary(summarize bool) func(*Command) {
	return func(cmd *Command) {
		if cmd.Options == nil {
			cmd.Options = make(map[string]interface{})
		}
		cmd.Options["summarize"] = summarize
	}
}

// WithPattern adds file pattern option to a command
func WithPattern(pattern string) func(*Command) {
	return func(cmd *Command) {
		if cmd.Options == nil {
			cmd.Options = make(map[string]interface{})
		}
		cmd.Options["pattern"] = pattern
	}
}

// WithDepth adds analysis depth option to a command
func WithDepth(depth string) func(*Command) {
	return func(cmd *Command) {
		if cmd.Options == nil {
			cmd.Options = make(map[string]interface{})
		}
		cmd.Options["depth"] = depth
	}
}

// WithLimit adds limit option to a command
func WithLimit(limit int) func(*Command) {
	return func(cmd *Command) {
		if cmd.Options == nil {
			cmd.Options = make(map[string]interface{})
		}
		cmd.Options["limit"] = limit
	}
}

// WithContext adds context information to a command
func WithContext(key string, value interface{}) func(*Command) {
	return func(cmd *Command) {
		if cmd.Context == nil {
			cmd.Context = make(map[string]interface{})
		}
		cmd.Context[key] = value
	}
}
