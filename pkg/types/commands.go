package types

// CommandType represents the type of command being executed
// Simplified to not prescribe specific command types - users can send any prompt
type CommandType string

// Command represents a Claude Code command with parameters
type Command struct {
	// Type is the command type
	Type CommandType `json:"type"`

	// Args are command arguments
	Args []string `json:"args,omitempty"`

	// Options are command options/flags
	Options map[string]any `json:"options,omitempty"`

	// Context provides additional context for the command
	Context map[string]any `json:"context,omitempty"`

	// VerboseOutput requests complete output without truncation
	VerboseOutput bool `json:"verbose_output,omitempty"`
}

// CommandResult represents the result of executing a command
type CommandResult struct {
	// Command is the original command
	Command *Command `json:"command"`

	// Success indicates if the command succeeded
	Success bool `json:"success"`

	// Output contains the command output
	Output string `json:"output,omitempty"`

	// FullOutput contains the complete output if it was truncated
	FullOutput string `json:"full_output,omitempty"`

	// IsTruncated indicates if the output was truncated
	IsTruncated bool `json:"is_truncated,omitempty"`

	// OutputLength is the actual length of the output before truncation
	OutputLength int `json:"output_length,omitempty"`

	// Error contains error information if the command failed
	Error string `json:"error,omitempty"`

	// Metadata contains additional result information
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ProjectContext represents basic project information
// Simplified to match official SDK scope (just working directory)
type ProjectContext struct {
	// WorkingDirectory is the current working directory
	WorkingDirectory string `json:"working_directory"`
}

// CommandList represents a list of commands to execute
type CommandList struct {
	// Commands is the list of commands to execute
	Commands []*Command `json:"commands"`

	// ExecutionMode specifies how to execute the commands
	ExecutionMode CommandExecutionMode `json:"execution_mode,omitempty"`

	// StopOnError determines if execution should stop when a command fails
	StopOnError bool `json:"stop_on_error,omitempty"`

	// MaxParallel limits the number of parallel commands (for parallel mode)
	MaxParallel int `json:"max_parallel,omitempty"`
}

// CommandExecutionMode specifies how commands in a list should be executed
type CommandExecutionMode string

const (
	// ExecutionModeSequential executes commands one after another
	ExecutionModeSequential CommandExecutionMode = "sequential"

	// ExecutionModeParallel executes commands in parallel
	ExecutionModeParallel CommandExecutionMode = "parallel"

	// ExecutionModeDependency executes commands based on dependencies
	ExecutionModeDependency CommandExecutionMode = "dependency"
)

// CommandListResult represents the results of executing a command list
type CommandListResult struct {
	// Results contains the individual command results
	Results []*CommandResult `json:"results"`

	// Success indicates if all commands succeeded
	Success bool `json:"success"`

	// TotalCommands is the total number of commands executed
	TotalCommands int `json:"total_commands"`

	// SuccessfulCommands is the number of commands that succeeded
	SuccessfulCommands int `json:"successful_commands"`

	// FailedCommands is the number of commands that failed
	FailedCommands int `json:"failed_commands"`

	// ExecutionTime is the total execution time in milliseconds
	ExecutionTime int64 `json:"execution_time,omitempty"`

	// Errors contains any errors that occurred
	Errors []string `json:"errors,omitempty"`
}
