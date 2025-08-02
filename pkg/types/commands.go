package types

// CommandType represents different types of Claude Code commands
type CommandType string

const (
	// File operations
	CommandRead   CommandType = "read"
	CommandWrite  CommandType = "write"
	CommandEdit   CommandType = "edit"
	CommandSearch CommandType = "search"

	// Code operations
	CommandAnalyze  CommandType = "analyze"
	CommandExplain  CommandType = "explain"
	CommandRefactor CommandType = "refactor"
	CommandTest     CommandType = "test"
	CommandDebug    CommandType = "debug"

	// Git operations
	CommandGitStatus CommandType = "git-status"
	CommandGitCommit CommandType = "git-commit"
	CommandGitDiff   CommandType = "git-diff"
	CommandGitLog    CommandType = "git-log"

	// Project operations
	CommandBuild   CommandType = "build"
	CommandRun     CommandType = "run"
	CommandInstall CommandType = "install"
	CommandClean   CommandType = "clean"

	// Session operations
	CommandHistory CommandType = "history"
	CommandClear   CommandType = "clear"
	CommandSave    CommandType = "save"
	CommandLoad    CommandType = "load"
)

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

// ProjectContext represents information about the current project
type ProjectContext struct {
	// WorkingDirectory is the current working directory
	WorkingDirectory string `json:"working_directory"`

	// ProjectName is the name of the project (derived from directory or config)
	ProjectName string `json:"project_name,omitempty"`

	// Language is the primary programming language detected
	Language string `json:"language,omitempty"`

	// Framework is the primary framework detected (if any)
	Framework string `json:"framework,omitempty"`

	// GitRepository contains git repository information
	GitRepository *GitInfo `json:"git_repository,omitempty"`

	// Files contains information about project files
	Files *ProjectFiles `json:"files,omitempty"`

	// Metadata contains additional project information
	Metadata map[string]any `json:"metadata,omitempty"`
}

// GitInfo contains information about a git repository
type GitInfo struct {
	// RemoteURL is the remote origin URL
	RemoteURL string `json:"remote_url,omitempty"`

	// Branch is the current branch name
	Branch string `json:"branch,omitempty"`

	// CommitHash is the current commit hash
	CommitHash string `json:"commit_hash,omitempty"`

	// IsDirty indicates if there are uncommitted changes
	IsDirty bool `json:"is_dirty"`

	// Status contains git status information
	Status map[string]any `json:"status,omitempty"`
}

// ProjectFiles contains information about project file structure
type ProjectFiles struct {
	// TotalFiles is the total number of files in the project
	TotalFiles int `json:"total_files"`

	// Languages maps file extensions to counts
	Languages map[string]int `json:"languages,omitempty"`

	// ImportantFiles lists key project files
	ImportantFiles []string `json:"important_files,omitempty"`

	// Structure provides a high-level view of the project structure
	Structure map[string]any `json:"structure,omitempty"`
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
