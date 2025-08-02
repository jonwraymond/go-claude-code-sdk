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
}

// CommandResult represents the result of executing a command
type CommandResult struct {
	// Command is the original command
	Command *Command `json:"command"`

	// Success indicates if the command succeeded
	Success bool `json:"success"`

	// Output contains the command output
	Output string `json:"output,omitempty"`

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
