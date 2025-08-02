package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ErrorCategory represents the category of an error for classification and handling.
type ErrorCategory string

const (
	// CategoryAPI represents errors from the Claude API service
	CategoryAPI ErrorCategory = "api"
	// CategoryNetwork represents network connectivity errors
	CategoryNetwork ErrorCategory = "network"
	// CategoryAuth represents authentication and authorization errors
	CategoryAuth ErrorCategory = "auth"
	// CategoryValidation represents request/response validation errors
	CategoryValidation ErrorCategory = "validation"
	// CategoryConfiguration represents SDK configuration errors
	CategoryConfiguration ErrorCategory = "configuration"
	// CategorySecurity represents security-related errors
	CategorySecurity ErrorCategory = "security"
	// CategoryInternal represents internal SDK errors
	CategoryInternal ErrorCategory = "internal"
)

// ErrorSeverity represents the severity level of an error.
type ErrorSeverity string

const (
	// SeverityLow represents non-critical errors that don't affect core functionality
	SeverityLow ErrorSeverity = "low"
	// SeverityMedium represents errors that affect some functionality but aren't critical
	SeverityMedium ErrorSeverity = "medium"
	// SeverityHigh represents critical errors that significantly impact functionality
	SeverityHigh ErrorSeverity = "high"
	// SeverityCritical represents critical errors that prevent core functionality
	SeverityCritical ErrorSeverity = "critical"
)

// SDKError is the interface that all SDK errors must implement.
// It extends the standard error interface with additional metadata and functionality.
type SDKError interface {
	error

	// Category returns the error category for classification
	Category() ErrorCategory

	// Severity returns the error severity level
	Severity() ErrorSeverity

	// Code returns a machine-readable error code
	Code() string

	// Message returns a user-friendly error message
	Message() string

	// Details returns additional error details for debugging
	Details() map[string]interface{}

	// IsRetryable returns true if the operation can be retried
	IsRetryable() bool

	// Unwrap returns the underlying error if any (for Go 1.13+ error wrapping)
	Unwrap() error

	// HTTPStatusCode returns the associated HTTP status code if applicable
	HTTPStatusCode() int

	// RequestID returns the request ID associated with this error if available
	RequestID() string

	// Timestamp returns when the error occurred
	Timestamp() time.Time
}

// BaseError provides a base implementation of SDKError that other error types can embed.
type BaseError struct {
	category      ErrorCategory
	severity      ErrorSeverity
	code          string
	message       string
	details       map[string]interface{}
	retryable     bool
	cause         error
	httpStatus    int
	requestID     string
	timestamp     time.Time
	stackTrace    []string
}

// NewBaseError creates a new BaseError with the provided parameters.
func NewBaseError(
	category ErrorCategory,
	severity ErrorSeverity,
	code string,
	message string,
) *BaseError {
	return &BaseError{
		category:   category,
		severity:   severity,
		code:       code,
		message:    message,
		details:    make(map[string]interface{}),
		retryable:  false,
		timestamp:  time.Now().UTC(),
		stackTrace: captureStackTrace(2), // Skip NewBaseError and calling function
	}
}

// Error implements the error interface.
func (e *BaseError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

// Category returns the error category.
func (e *BaseError) Category() ErrorCategory {
	return e.category
}

// Severity returns the error severity.
func (e *BaseError) Severity() ErrorSeverity {
	return e.severity
}

// Code returns the error code.
func (e *BaseError) Code() string {
	return e.code
}

// Message returns the user-friendly message.
func (e *BaseError) Message() string {
	return e.message
}

// Details returns the error details.
func (e *BaseError) Details() map[string]interface{} {
	// Return a copy to prevent external modification
	details := make(map[string]interface{})
	for k, v := range e.details {
		details[k] = v
	}
	return details
}

// IsRetryable returns whether the error is retryable.
func (e *BaseError) IsRetryable() bool {
	return e.retryable
}

// Unwrap returns the underlying error.
func (e *BaseError) Unwrap() error {
	return e.cause
}

// HTTPStatusCode returns the HTTP status code.
func (e *BaseError) HTTPStatusCode() int {
	return e.httpStatus
}

// RequestID returns the request ID.
func (e *BaseError) RequestID() string {
	return e.requestID
}

// Timestamp returns when the error occurred.
func (e *BaseError) Timestamp() time.Time {
	return e.timestamp
}

// WithCause sets the underlying cause of the error.
func (e *BaseError) WithCause(cause error) *BaseError {
	e.cause = cause
	return e
}

// WithDetail adds a detail key-value pair to the error.
func (e *BaseError) WithDetail(key string, value interface{}) *BaseError {
	if e.details == nil {
		e.details = make(map[string]interface{})
	}
	e.details[key] = value
	return e
}

// WithDetails sets multiple details at once.
func (e *BaseError) WithDetails(details map[string]interface{}) *BaseError {
	if e.details == nil {
		e.details = make(map[string]interface{})
	}
	for k, v := range details {
		e.details[k] = v
	}
	return e
}

// WithHTTPStatus sets the HTTP status code.
func (e *BaseError) WithHTTPStatus(status int) *BaseError {
	e.httpStatus = status
	return e
}

// WithRequestID sets the request ID.
func (e *BaseError) WithRequestID(requestID string) *BaseError {
	e.requestID = requestID
	return e
}

// WithRetryable sets whether the error is retryable.
func (e *BaseError) WithRetryable(retryable bool) *BaseError {
	e.retryable = retryable
	return e
}

// String returns a detailed string representation of the error.
func (e *BaseError) String() string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("Category: %s", e.category))
	parts = append(parts, fmt.Sprintf("Severity: %s", e.severity))
	parts = append(parts, fmt.Sprintf("Code: %s", e.code))
	parts = append(parts, fmt.Sprintf("Message: %s", e.message))
	
	if e.httpStatus > 0 {
		parts = append(parts, fmt.Sprintf("HTTP Status: %d", e.httpStatus))
	}
	
	if e.requestID != "" {
		parts = append(parts, fmt.Sprintf("Request ID: %s", e.requestID))
	}
	
	if e.retryable {
		parts = append(parts, "Retryable: true")
	}
	
	if len(e.details) > 0 {
		if detailsJSON, err := json.Marshal(e.details); err == nil {
			parts = append(parts, fmt.Sprintf("Details: %s", string(detailsJSON)))
		}
	}
	
	if e.cause != nil {
		parts = append(parts, fmt.Sprintf("Cause: %v", e.cause))
	}
	
	return strings.Join(parts, ", ")
}

// InternalError represents internal SDK errors that are not user-facing.
type InternalError struct {
	*BaseError
}

// NewInternalError creates a new internal error.
func NewInternalError(code, message string) *InternalError {
	return &InternalError{
		BaseError: NewBaseError(CategoryInternal, SeverityHigh, code, message),
	}
}

// ConfigurationError represents configuration-related errors.
type ConfigurationError struct {
	*BaseError
	Field string // The configuration field that has an issue
}

// NewConfigurationError creates a new configuration error.
func NewConfigurationError(field, message string) *ConfigurationError {
	return &ConfigurationError{
		BaseError: NewBaseError(CategoryConfiguration, SeverityHigh, "CONFIGURATION_ERROR", message).
			WithDetail("field", field),
		Field: field,
	}
}

// Utility functions for error handling

// IsRetryable checks if an error is retryable by examining the error chain.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	
	// Check if the error implements SDKError
	var sdkErr SDKError
	if errors.As(err, &sdkErr) {
		return sdkErr.IsRetryable()
	}
	
	// For non-SDK errors, use heuristics
	return isRetryableHeuristic(err)
}

// GetCategory extracts the error category from an error.
func GetCategory(err error) ErrorCategory {
	if err == nil {
		return ""
	}
	
	var sdkErr SDKError
	if errors.As(err, &sdkErr) {
		return sdkErr.Category()
	}
	
	return CategoryInternal
}

// GetSeverity extracts the error severity from an error.
func GetSeverity(err error) ErrorSeverity {
	if err == nil {
		return SeverityLow
	}
	
	var sdkErr SDKError
	if errors.As(err, &sdkErr) {
		return sdkErr.Severity()
	}
	
	return SeverityMedium
}

// GetRequestID extracts the request ID from an error if available.
func GetRequestID(err error) string {
	if err == nil {
		return ""
	}
	
	var sdkErr SDKError
	if errors.As(err, &sdkErr) {
		return sdkErr.RequestID()
	}
	
	return ""
}

// GetHTTPStatusCode extracts the HTTP status code from an error if available.
func GetHTTPStatusCode(err error) int {
	if err == nil {
		return 0
	}
	
	var sdkErr SDKError
	if errors.As(err, &sdkErr) {
		return sdkErr.HTTPStatusCode()
	}
	
	return 0
}

// WrapError wraps an existing error with additional context while preserving the error chain.
func WrapError(err error, category ErrorCategory, code, message string) SDKError {
	if err == nil {
		return nil
	}
	
	// Determine severity based on the underlying error
	severity := SeverityMedium
	if sdkErr, ok := err.(SDKError); ok {
		severity = sdkErr.Severity()
	}
	
	baseErr := NewBaseError(category, severity, code, message).WithCause(err)
	
	// Preserve request ID and HTTP status if available from underlying error
	if sdkErr, ok := err.(SDKError); ok {
		if requestID := sdkErr.RequestID(); requestID != "" {
			baseErr.WithRequestID(requestID)
		}
		if httpStatus := sdkErr.HTTPStatusCode(); httpStatus > 0 {
			baseErr.WithHTTPStatus(httpStatus)
		}
		baseErr.WithRetryable(sdkErr.IsRetryable())
	}
	
	return baseErr
}

// isRetryableHeuristic determines if a non-SDK error might be retryable.
func isRetryableHeuristic(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := strings.ToLower(err.Error())
	
	// Network-related errors that are typically retryable
	retryablePatterns := []string{
		"timeout",
		"connection reset",
		"connection refused", 
		"no such host",
		"network is unreachable",
		"temporary failure",
		"service unavailable",
		"too many requests",
		"rate limit",
		"server error",
		"internal server error",
		"bad gateway",
		"service unavailable",
		"gateway timeout",
	}
	
	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	
	return false
}

// captureStackTrace captures a simple stack trace for debugging.
// This is a simplified implementation - in production you might want to use
// a more sophisticated stack trace library.
func captureStackTrace(skip int) []string {
	// For now, return an empty slice. In a full implementation,
	// you would capture the actual stack trace here.
	// This could use runtime.Caller() or a library like github.com/pkg/errors
	return []string{}
}

// HTTPErrorFromStatus creates an appropriate error based on HTTP status code.
func HTTPErrorFromStatus(statusCode int, message string) SDKError {
	code := fmt.Sprintf("HTTP_%d", statusCode)
	category := CategoryAPI
	severity := SeverityMedium
	retryable := false
	
	// Categorize based on status code
	switch {
	case statusCode >= 500:
		// Server errors are typically retryable
		severity = SeverityHigh
		retryable = true
		if message == "" {
			message = fmt.Sprintf("Server error (HTTP %d)", statusCode)
		}
	case statusCode == http.StatusTooManyRequests:
		// Rate limiting is retryable
		retryable = true
		if message == "" {
			message = "Rate limit exceeded"
		}
	case statusCode >= 400:
		// Client errors are typically not retryable
		severity = SeverityCritical
		if message == "" {
			message = fmt.Sprintf("Client error (HTTP %d)", statusCode)
		}
	default:
		if message == "" {
			message = fmt.Sprintf("HTTP error (status %d)", statusCode)
		}
	}
	
	return NewBaseError(category, severity, code, message).
		WithHTTPStatus(statusCode).
		WithRetryable(retryable)
}