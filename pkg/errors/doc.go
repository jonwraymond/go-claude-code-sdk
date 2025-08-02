/*
Package errors provides comprehensive error handling for the Claude Code Go SDK.

This package defines a rich error hierarchy that captures detailed information about
failures while maintaining Go idioms. All errors in the SDK flow through this package,
providing consistent error handling and recovery strategies.

# Error Hierarchy

The package defines a base ClaudeCodeError type with specialized error types:

	// Base error type
	type ClaudeCodeError struct {
		Code       string                 // Unique error code
		Message    string                 // Human-readable message
		Category   string                 // Error category for classification
		HTTPStatus int                    // HTTP status code if applicable
		Retryable  bool                   // Whether the operation can be retried
		Details    map[string]any // Additional error context
		Cause      error                  // Underlying error if wrapped
	}

# Error Categories

Errors are classified into categories for better handling:

	errors.CategoryNetwork    // Network connectivity issues
	errors.CategoryValidation // Input validation failures
	errors.CategoryAuth       // Authentication/authorization errors
	errors.CategoryInternal   // Internal SDK errors
	errors.CategoryAPI        // Claude Code API errors
	errors.CategoryProcess    // Subprocess execution errors

# Creating Errors

The package provides constructor functions for common error types:

	// Validation error
	err := errors.NewValidationError("field", value, "required", "field is required")

	// API error with retry information
	err := errors.NewAPIError("RATE_LIMIT", "Too many requests", 429, true)

	// Process error for CLI failures
	err := errors.NewProcessError("CLAUDE_NOT_FOUND", "claude command not found")

# Error Wrapping

Support for Go 1.13+ error wrapping:

	// Wrap an existing error
	err := errors.WrapError(originalErr, errors.CategoryNetwork, "NETWORK_TIMEOUT",
		"failed to connect to Claude Code")

	// Check wrapped errors
	var claudeErr *errors.ClaudeCodeError
	if errors.As(err, &claudeErr) {
		if claudeErr.IsRetryable() {
			// Retry logic
		}
	}

# Error Checking

The package provides utilities for error classification:

	// Check if error is retryable
	if errors.IsRetryable(err) {
		// Implement retry logic
	}

	// Check error category
	if errors.GetCategory(err) == errors.CategoryNetwork {
		// Handle network errors
	}

	// Check specific error codes
	if errors.HasCode(err, "RATE_LIMIT") {
		// Handle rate limiting
	}

# HTTP Error Mapping

Convert HTTP status codes to appropriate errors:

	err := errors.HTTPErrorFromStatus(429, "Rate limit exceeded")
	// Returns appropriate error with retry information

# Validation Errors

Rich validation error support:

	type ValidationError struct {
		Field      string      // Field that failed validation
		Value      any // Actual value provided
		Constraint string      // Validation constraint violated
		Message    string      // Human-readable message
	}

	// Create validation error
	err := errors.NewValidationError("max_tokens", -1, "positive",
		"max_tokens must be positive")

# Process Errors

Subprocess execution errors with exit codes:

	type ProcessError struct {
		*ClaudeCodeError
		ExitCode int    // Process exit code
		Stderr   string // Captured stderr output
	}

# Error Context

Add context to errors for better debugging:

	err := errors.NewInternalError("PARSE_FAILURE", "Failed to parse response").
		WithDetail("response", responseBody).
		WithDetail("contentType", contentType)

# Best Practices

1. Always use specific error types when possible
2. Include relevant context in error details
3. Set Retryable flag appropriately
4. Use error wrapping to preserve error chains
5. Check error types with errors.As() for type-specific handling

# Example Usage

	result, err := client.Query(ctx, request)
	if err != nil {
		// Type-specific handling
		var validationErr *errors.ValidationError
		if errors.As(err, &validationErr) {
			log.Printf("Validation failed for field %s: %s",
				validationErr.Field, validationErr.Message)
			return
		}

		// Category-based handling
		switch errors.GetCategory(err) {
		case errors.CategoryNetwork:
			if errors.IsRetryable(err) {
				// Implement retry with backoff
			}
		case errors.CategoryAuth:
			// Re-authenticate
		default:
			// Generic error handling
		}
	}

The errors package ensures consistent, informative error handling throughout the
Claude Code Go SDK, making it easier to build robust applications.
*/
package errors
