package client

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// Re-export types from types package for convenience
type CommandType = types.CommandType
type Command = types.Command
type CommandResult = types.CommandResult
type CommandList = types.CommandList
type CommandListResult = types.CommandListResult
type CommandExecutionMode = types.CommandExecutionMode

// Re-export execution mode constants from types package
const (
	ExecutionModeSequential = types.ExecutionModeSequential
	ExecutionModeParallel   = types.ExecutionModeParallel
	ExecutionModeDependency = types.ExecutionModeDependency
)

// Common command types as string constants for convenience
// These are not part of the official SDK API - users can send any prompt
const (
	CommandRead     CommandType = "read"
	CommandWrite    CommandType = "write"
	CommandEdit     CommandType = "edit"
	CommandSearch   CommandType = "search"
	CommandAnalyze  CommandType = "analyze"
	CommandExplain  CommandType = "explain"
	CommandRefactor CommandType = "refactor"
	CommandTest     CommandType = "test"
	CommandDebug    CommandType = "debug"

	CommandGitStatus CommandType = "git-status"
	CommandGitCommit CommandType = "git-commit"
	CommandGitDiff   CommandType = "git-diff"
	CommandGitLog    CommandType = "git-log"

	CommandBuild   CommandType = "build"
	CommandRun     CommandType = "run"
	CommandInstall CommandType = "install"
	CommandClean   CommandType = "clean"

	CommandHistory CommandType = "history"
	CommandClear   CommandType = "clear"
	CommandSave    CommandType = "save"
	CommandLoad    CommandType = "load"
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

	// Check for truncation indicators
	isTruncated := ce.isOutputTruncated(output)
	outputLength := len(output)

	result := &CommandResult{
		Command:      cmd,
		Success:      true,
		Output:       output,
		IsTruncated:  isTruncated,
		OutputLength: outputLength,
		Metadata: map[string]any{
			"stop_reason": response.StopReason,
		},
	}

	// If truncated and verbose output requested, try to get full output
	if isTruncated && cmd.VerboseOutput {
		fullOutput, err := ce.getFullOutput(ctx, cmd, output)
		if err == nil && fullOutput != "" {
			result.FullOutput = fullOutput
			result.OutputLength = len(fullOutput)
		}
	}

	return result, nil
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

// ExecuteCommands executes a list of commands according to the specified execution mode
func (ce *CommandExecutor) ExecuteCommands(ctx context.Context, cmdList *CommandList) (*CommandListResult, error) {
	if cmdList == nil || len(cmdList.Commands) == 0 {
		return &CommandListResult{
			Success:       true,
			TotalCommands: 0,
			Results:       []*CommandResult{},
		}, nil
	}

	startTime := time.Now()

	// Default execution mode is sequential
	mode := cmdList.ExecutionMode
	if mode == "" {
		mode = ExecutionModeSequential
	}

	var result *CommandListResult
	var err error

	switch mode {
	case ExecutionModeSequential:
		result, err = ce.executeSequential(ctx, cmdList)
	case ExecutionModeParallel:
		result, err = ce.executeParallel(ctx, cmdList)
	case ExecutionModeDependency:
		// For now, treat dependency mode as sequential
		// Future enhancement: implement dependency graph execution
		result, err = ce.executeSequential(ctx, cmdList)
	default:
		return nil, fmt.Errorf("unknown execution mode: %s", mode)
	}

	if result != nil {
		result.ExecutionTime = time.Since(startTime).Milliseconds()
	}

	return result, err
}

// executeSequential executes commands one after another
func (ce *CommandExecutor) executeSequential(ctx context.Context, cmdList *CommandList) (*CommandListResult, error) {
	result := &CommandListResult{
		Results:       make([]*CommandResult, 0, len(cmdList.Commands)),
		TotalCommands: len(cmdList.Commands),
		Success:       true,
		Errors:        make([]string, 0),
	}

	for i, cmd := range cmdList.Commands {
		select {
		case <-ctx.Done():
			result.Success = false
			result.Errors = append(result.Errors, fmt.Sprintf("execution cancelled at command %d", i+1))
			return result, ctx.Err()
		default:
		}

		cmdResult, err := ce.ExecuteCommand(ctx, cmd)
		if err != nil {
			result.FailedCommands++
			result.Success = false
			result.Errors = append(result.Errors, fmt.Sprintf("command %d error: %v", i+1, err))

			// Create a failed result
			cmdResult = &CommandResult{
				Command: cmd,
				Success: false,
				Error:   err.Error(),
			}
		} else if cmdResult.Success {
			result.SuccessfulCommands++
		} else {
			result.FailedCommands++
			result.Success = false
			if cmdResult.Error != "" {
				result.Errors = append(result.Errors, fmt.Sprintf("command %d: %s", i+1, cmdResult.Error))
			}
		}

		result.Results = append(result.Results, cmdResult)

		// Stop on error if requested
		if cmdList.StopOnError && !cmdResult.Success {
			break
		}
	}

	return result, nil
}

// executeParallel executes commands in parallel
func (ce *CommandExecutor) executeParallel(ctx context.Context, cmdList *CommandList) (*CommandListResult, error) {
	result := &CommandListResult{
		Results:       make([]*CommandResult, len(cmdList.Commands)),
		TotalCommands: len(cmdList.Commands),
		Success:       true,
		Errors:        make([]string, 0),
	}

	// Determine max parallel execution
	maxParallel := cmdList.MaxParallel
	if maxParallel <= 0 {
		maxParallel = 5 // Default max parallel
	}
	if maxParallel > len(cmdList.Commands) {
		maxParallel = len(cmdList.Commands)
	}

	// Create semaphore for parallel limit
	semaphore := make(chan struct{}, maxParallel)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, cmd := range cmdList.Commands {
		select {
		case <-ctx.Done():
			result.Success = false
			result.Errors = append(result.Errors, "execution cancelled")
			return result, ctx.Err()
		default:
		}

		wg.Add(1)
		go func(index int, command *Command) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Execute command
			cmdResult, err := ce.ExecuteCommand(ctx, command)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result.FailedCommands++
				result.Success = false
				result.Errors = append(result.Errors, fmt.Sprintf("command %d error: %v", index+1, err))

				// Create a failed result
				cmdResult = &CommandResult{
					Command: command,
					Success: false,
					Error:   err.Error(),
				}
			} else if cmdResult.Success {
				result.SuccessfulCommands++
			} else {
				result.FailedCommands++
				result.Success = false
				if cmdResult.Error != "" {
					result.Errors = append(result.Errors, fmt.Sprintf("command %d: %s", index+1, cmdResult.Error))
				}
			}

			result.Results[index] = cmdResult
		}(i, cmd)
	}

	// Wait for all commands to complete
	wg.Wait()

	return result, nil
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

	// Add verbose output request if enabled
	if cmd.VerboseOutput {
		prompt.WriteString("\n\nPlease provide the complete output without any truncation. Include all details and content.")
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
	options := make(map[string]any)

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

// isOutputTruncated checks if the output appears to be truncated
func (ce *CommandExecutor) isOutputTruncated(output string) bool {
	// Common truncation indicators
	truncationIndicators := []string{
		"...",
		"[truncated]",
		"[output truncated]",
		"... (truncated)",
		"[more content]",
		"[rest of output omitted]",
	}

	output = strings.TrimSpace(output)
	lowerOutput := strings.ToLower(output)

	// Check if output ends with truncation indicators
	for _, indicator := range truncationIndicators {
		if strings.HasSuffix(output, indicator) || strings.HasSuffix(lowerOutput, indicator) {
			return true
		}
	}

	// Check if output is suspiciously short for certain commands
	if len(output) < 10 && output == "..." {
		return true
	}

	// Check for incomplete patterns (e.g., cut off mid-sentence or mid-JSON)
	if len(output) > 20 { // Reduced threshold for better detection
		lastChar := output[len(output)-1]

		// Check for incomplete JSON/code structures
		openBraces := strings.Count(output, "{") - strings.Count(output, "}")
		openBrackets := strings.Count(output, "[") - strings.Count(output, "]")
		openQuotes := strings.Count(output, `"`) % 2

		if openBraces > 0 || openBrackets > 0 || openQuotes != 0 {
			return true
		}

		// If doesn't end with typical punctuation, might be truncated
		if lastChar != '.' && lastChar != '!' && lastChar != '?' && lastChar != '\n' && lastChar != '}' && lastChar != ']' && lastChar != '"' && lastChar != '\'' {
			// Additional check: if it looks like it ends mid-word or mid-sentence
			lastWords := strings.Fields(output)
			if len(lastWords) > 0 {
				lastWord := lastWords[len(lastWords)-1]
				// If the last word doesn't end with punctuation and seems incomplete
				if len(lastWord) > 0 && !strings.ContainsAny(lastWord, ".!?,;:)]}\"'") {
					// Check if it might be a complete word by looking at context
					// Short outputs might just be brief responses
					if len(output) < 50 && !strings.Contains(output, "...") {
						return false
					}
					return true
				}
			}
		}
	}

	return false
}

// getFullOutput attempts to retrieve the full output for a truncated command
func (ce *CommandExecutor) getFullOutput(ctx context.Context, cmd *Command, truncatedOutput string) (string, error) {
	// Build a follow-up prompt to get the complete output
	prompt := fmt.Sprintf("The previous command output was truncated. Please provide the complete output for the %s command", cmd.Type)
	if len(cmd.Args) > 0 {
		prompt += fmt.Sprintf(" with arguments: %v", cmd.Args)
	}
	prompt += ". Please include ALL content without any truncation."

	request := &types.QueryRequest{
		Messages: []types.Message{
			{Role: types.RoleUser, Content: prompt},
		},
		MaxTokens: 8000, // Request more tokens for full output
	}

	response, err := ce.client.Query(ctx, request)
	if err != nil {
		return "", err
	}

	return ce.ExtractTextContent(response.Content), nil
}

// Helper functions for truncation detection have been removed as they were unused

// Common command builder methods

// ReadFile creates a command to read a file
func ReadFile(filename string, options ...func(*Command)) *Command {
	cmd := &Command{
		Type:    CommandRead,
		Args:    []string{filename},
		Options: make(map[string]any),
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
		Options: make(map[string]any),
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
		Options: make(map[string]any),
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
		Options: make(map[string]any),
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
		Options: make(map[string]any),
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
			cmd.Options = make(map[string]any)
		}
		cmd.Options["summarize"] = summarize
	}
}

// WithPattern adds file pattern option to a command
func WithPattern(pattern string) func(*Command) {
	return func(cmd *Command) {
		if cmd.Options == nil {
			cmd.Options = make(map[string]any)
		}
		cmd.Options["pattern"] = pattern
	}
}

// WithDepth adds analysis depth option to a command
func WithDepth(depth string) func(*Command) {
	return func(cmd *Command) {
		if cmd.Options == nil {
			cmd.Options = make(map[string]any)
		}
		cmd.Options["depth"] = depth
	}
}

// WithLimit adds limit option to a command
func WithLimit(limit int) func(*Command) {
	return func(cmd *Command) {
		if cmd.Options == nil {
			cmd.Options = make(map[string]any)
		}
		cmd.Options["limit"] = limit
	}
}

// WithContext adds context information to a command
func WithContext(key string, value any) func(*Command) {
	return func(cmd *Command) {
		if cmd.Context == nil {
			cmd.Context = make(map[string]any)
		}
		cmd.Context[key] = value
	}
}

// WithVerboseOutput enables verbose output for a command
func WithVerboseOutput() func(*Command) {
	return func(cmd *Command) {
		cmd.VerboseOutput = true
	}
}

// Command list builder functions

// NewCommandList creates a new command list with the specified commands
func NewCommandList(commands ...*Command) *CommandList {
	return &CommandList{
		Commands:      commands,
		ExecutionMode: ExecutionModeSequential, // Default to sequential
		StopOnError:   true,                    // Default to stopping on error
	}
}

// NewParallelCommandList creates a new command list for parallel execution
func NewParallelCommandList(maxParallel int, commands ...*Command) *CommandList {
	return &CommandList{
		Commands:      commands,
		ExecutionMode: ExecutionModeParallel,
		MaxParallel:   maxParallel,
		StopOnError:   false, // Default to not stopping on error for parallel
	}
}

// CommandListOption is a function that modifies a CommandList
type CommandListOption func(*CommandList)

// WithExecutionMode sets the execution mode for a command list
func WithExecutionMode(mode CommandExecutionMode) CommandListOption {
	return func(cl *CommandList) {
		cl.ExecutionMode = mode
	}
}

// WithStopOnError sets whether to stop execution on first error
func WithStopOnError(stop bool) CommandListOption {
	return func(cl *CommandList) {
		cl.StopOnError = stop
	}
}

// WithMaxParallel sets the maximum number of parallel commands
func WithMaxParallel(max int) CommandListOption {
	return func(cl *CommandList) {
		cl.MaxParallel = max
	}
}

// CreateCommandList creates a command list with options
func CreateCommandList(commands []*Command, options ...CommandListOption) *CommandList {
	cl := &CommandList{
		Commands:      commands,
		ExecutionMode: ExecutionModeSequential,
		StopOnError:   true,
	}

	for _, opt := range options {
		opt(cl)
	}

	return cl
}
