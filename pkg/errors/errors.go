package errors

import (
	"fmt"
)

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