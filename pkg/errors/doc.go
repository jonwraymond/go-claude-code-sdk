// Package errors defines error types for the Claude Code SDK.
//
// This package provides structured error types that help identify and handle
// various failure scenarios when interacting with Claude Code:
//
//	- CLINotFoundError: The Claude CLI is not installed or not in PATH
//	- CLIConnectionError: Failed to establish connection with Claude CLI
//	- ProcessError: The Claude CLI process exited with an error
//	- CLITimeoutError: Operation timed out
//	- ClaudeSDKError: General SDK errors
//
// All error types implement the standard error interface and provide additional
// context about the failure. Example usage:
//
//	messages, err := claudecode.Query(ctx, prompt, nil)
//	if err != nil {
//	    var cliErr *errors.CLINotFoundError
//	    if errors.As(err, &cliErr) {
//	        fmt.Println("Please install Claude CLI")
//	    }
//	}
package errors