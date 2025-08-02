package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// APIError represents errors from the Claude API service.
type APIError struct {
	*BaseError
	APICode    string `json:"code,omitempty"`    // Claude API error code
	APIMessage string `json:"message,omitempty"` // Claude API error message
	APIType    string `json:"type,omitempty"`    // Claude API error type
}

// NewAPIError creates a new API error with Claude API error details.
func NewAPIError(statusCode int, apiCode, apiType, apiMessage string) *APIError {
	severity := SeverityMedium
	retryable := false
	
	// Determine severity and retryability based on status code
	switch statusCode {
	case http.StatusTooManyRequests:
		severity = SeverityMedium
		retryable = true
	case http.StatusInternalServerError, http.StatusBadGateway, 
		 http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		severity = SeverityHigh
		retryable = true
	case http.StatusUnauthorized, http.StatusForbidden:
		severity = SeverityCritical
		retryable = false
	default:
		if statusCode >= 400 && statusCode < 500 {
			severity = SeverityHigh
			retryable = false
		}
	}
	
	// Create user-friendly message
	userMessage := formatAPIErrorMessage(statusCode, apiCode, apiType, apiMessage)
	
	err := &APIError{
		BaseError: NewBaseError(CategoryAPI, severity, apiCode, userMessage).
			WithHTTPStatus(statusCode).
			WithRetryable(retryable).
			WithDetail("api_code", apiCode).
			WithDetail("api_type", apiType).
			WithDetail("api_message", apiMessage),
		APICode:    apiCode,
		APIMessage: apiMessage,
		APIType:    apiType,
	}
	
	return err
}

// ParseAPIErrorFromResponse parses an API error from an HTTP response body.
func ParseAPIErrorFromResponse(statusCode int, responseBody []byte) *APIError {
	// Try to parse the response as Claude API error format
	var apiErrorResponse struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	
	apiCode := fmt.Sprintf("HTTP_%d", statusCode)
	apiType := "unknown_error"
	apiMessage := ""
	
	if err := json.Unmarshal(responseBody, &apiErrorResponse); err == nil {
		if apiErrorResponse.Error.Type != "" {
			apiType = apiErrorResponse.Error.Type
			apiCode = apiType
		}
		if apiErrorResponse.Error.Message != "" {
			apiMessage = apiErrorResponse.Error.Message
		}
	}
	
	return NewAPIError(statusCode, apiCode, apiType, apiMessage)
}

// RateLimitError represents rate limiting errors with retry information.
type RateLimitError struct {
	*APIError
	RetryAfter time.Duration // How long to wait before retrying
	Limit      int64         // The rate limit
	Remaining  int64         // Remaining requests in the current window
	Reset      time.Time     // When the rate limit resets
}

// NewRateLimitError creates a new rate limit error.
func NewRateLimitError(retryAfter time.Duration, limit, remaining int64, reset time.Time) *RateLimitError {
	message := fmt.Sprintf("Rate limit exceeded. Retry after %v", retryAfter)
	if limit > 0 {
		message = fmt.Sprintf("Rate limit exceeded (%d/%d requests). Retry after %v", 
			limit-remaining, limit, retryAfter)
	}
	
	apiErr := NewAPIError(http.StatusTooManyRequests, "rate_limit_error", "rate_limit_error", message)
	
	err := &RateLimitError{
		APIError:   apiErr,
		RetryAfter: retryAfter,
		Limit:      limit,
		Remaining:  remaining,
		Reset:      reset,
	}
	
	// Add rate limit details
	err.WithDetail("retry_after_seconds", retryAfter.Seconds()).
		WithDetail("limit", limit).
		WithDetail("remaining", remaining).
		WithDetail("reset", reset.Format(time.RFC3339))
	
	return err
}

// ParseRateLimitErrorFromHeaders creates a rate limit error from HTTP headers.
func ParseRateLimitErrorFromHeaders(headers http.Header) *RateLimitError {
	var retryAfter time.Duration
	var limit, remaining int64
	var reset time.Time
	
	// Parse Retry-After header
	if retryAfterStr := headers.Get("Retry-After"); retryAfterStr != "" {
		if seconds, err := strconv.Atoi(retryAfterStr); err == nil {
			retryAfter = time.Duration(seconds) * time.Second
		}
	}
	
	// Parse X-RateLimit headers (common format)
	if limitStr := headers.Get("X-RateLimit-Limit"); limitStr != "" {
		limit, _ = strconv.ParseInt(limitStr, 10, 64)
	}
	if remainingStr := headers.Get("X-RateLimit-Remaining"); remainingStr != "" {
		remaining, _ = strconv.ParseInt(remainingStr, 10, 64)
	}
	if resetStr := headers.Get("X-RateLimit-Reset"); resetStr != "" {
		if resetTimestamp, err := strconv.ParseInt(resetStr, 10, 64); err == nil {
			reset = time.Unix(resetTimestamp, 0)
		}
	}
	
	// Default retry after if not specified
	if retryAfter == 0 {
		retryAfter = 60 * time.Second // Default to 60 seconds
	}
	
	return NewRateLimitError(retryAfter, limit, remaining, reset)
}

// QuotaExceededError represents quota or usage limit exceeded errors.
type QuotaExceededError struct {
	*APIError
	QuotaType string    // Type of quota exceeded (tokens, requests, etc.)
	Used      int64     // Amount used
	Limit     int64     // The quota limit
	ResetsAt  time.Time // When the quota resets
}

// NewQuotaExceededError creates a new quota exceeded error.
func NewQuotaExceededError(quotaType string, used, limit int64, resetsAt time.Time) *QuotaExceededError {
	message := fmt.Sprintf("%s quota exceeded (%d/%d)", quotaType, used, limit)
	if !resetsAt.IsZero() {
		message += fmt.Sprintf(". Resets at %s", resetsAt.Format(time.RFC3339))
	}
	
	apiErr := NewAPIError(http.StatusPaymentRequired, "quota_exceeded", "quota_exceeded", message)
	// Quota errors are typically not retryable until the quota resets
	apiErr.WithRetryable(false)
	
	err := &QuotaExceededError{
		APIError:  apiErr,
		QuotaType: quotaType,
		Used:      used,
		Limit:     limit,
		ResetsAt:  resetsAt,
	}
	
	err.WithDetail("quota_type", quotaType).
		WithDetail("used", used).
		WithDetail("limit", limit).
		WithDetail("resets_at", resetsAt.Format(time.RFC3339))
	
	return err
}

// ModelUnavailableError represents errors when a requested model is unavailable.
type ModelUnavailableError struct {
	*APIError
	ModelID           string   // The requested model ID
	AvailableModels   []string // List of available models
	SuggestedModel    string   // Suggested alternative model
	UnavailableReason string   // Reason why the model is unavailable
}

// NewModelUnavailableError creates a new model unavailable error.
func NewModelUnavailableError(modelID, reason, suggestedModel string, availableModels []string) *ModelUnavailableError {
	message := fmt.Sprintf("Model '%s' is not available", modelID)
	if reason != "" {
		message += fmt.Sprintf(": %s", reason)
	}
	if suggestedModel != "" {
		message += fmt.Sprintf(". Try using '%s' instead", suggestedModel)
	}
	
	apiErr := NewAPIError(http.StatusBadRequest, "model_unavailable", "invalid_request_error", message)
	apiErr.WithRetryable(false) // Model unavailability is typically not retryable
	
	err := &ModelUnavailableError{
		APIError:          apiErr,
		ModelID:           modelID,
		AvailableModels:   availableModels,
		SuggestedModel:    suggestedModel,
		UnavailableReason: reason,
	}
	
	err.WithDetail("model_id", modelID).
		WithDetail("reason", reason).
		WithDetail("suggested_model", suggestedModel).
		WithDetail("available_models", availableModels)
	
	return err
}

// ContentPolicyError represents errors related to content policy violations.
type ContentPolicyError struct {
	*APIError
	PolicyType string   // Type of policy violated
	ViolatedRules []string // Specific rules that were violated
}

// NewContentPolicyError creates a new content policy error.
func NewContentPolicyError(policyType string, violatedRules []string) *ContentPolicyError {
	message := "Content violates usage policies"
	if policyType != "" {
		message = fmt.Sprintf("Content violates %s policy", policyType)
	}
	
	apiErr := NewAPIError(http.StatusBadRequest, "content_policy_violation", "invalid_request_error", message)
	apiErr.WithRetryable(false) // Policy violations are not retryable
	
	err := &ContentPolicyError{
		APIError:      apiErr,
		PolicyType:    policyType,
		ViolatedRules: violatedRules,
	}
	
	err.WithDetail("policy_type", policyType).
		WithDetail("violated_rules", violatedRules)
	
	return err
}

// InvalidRequestError represents errors in request format or parameters.
type InvalidRequestError struct {
	*APIError
	Parameter string // The invalid parameter
	Value     string // The invalid value (sanitized)
	Expected  string // What was expected
}

// NewInvalidRequestError creates a new invalid request error.
func NewInvalidRequestError(parameter, value, expected string) *InvalidRequestError {
	message := fmt.Sprintf("Invalid request parameter '%s'", parameter)
	if expected != "" {
		message += fmt.Sprintf(": expected %s", expected)
	}
	
	// Sanitize the value to avoid exposing sensitive data
	sanitizedValue := sanitizeValue(value)
	
	apiErr := NewAPIError(http.StatusBadRequest, "invalid_request_error", "invalid_request_error", message)
	apiErr.WithRetryable(false) // Invalid requests are not retryable without changes
	
	err := &InvalidRequestError{
		APIError:  apiErr,
		Parameter: parameter,
		Value:     sanitizedValue,
		Expected:  expected,
	}
	
	err.WithDetail("parameter", parameter).
		WithDetail("value", sanitizedValue).
		WithDetail("expected", expected)
	
	return err
}

// Helper functions

// formatAPIErrorMessage creates a user-friendly error message from API error details.
func formatAPIErrorMessage(statusCode int, apiCode, apiType, apiMessage string) string {
	// If we have a clear API message, use it
	if apiMessage != "" {
		return apiMessage
	}
	
	// Otherwise, create a message based on the error type or status code
	switch apiType {
	case "rate_limit_error":
		return "Request rate limit exceeded. Please wait before making more requests."
	case "quota_exceeded":
		return "Usage quota exceeded. Please check your account limits."
	case "invalid_request_error":
		return "The request was invalid. Please check your parameters and try again."
	case "authentication_error":
		return "Authentication failed. Please check your API key."
	case "permission_error":
		return "Permission denied. Your API key may not have access to this resource."
	case "not_found_error":
		return "The requested resource was not found."
	case "overloaded_error":
		return "The API is temporarily overloaded. Please try again later."
	case "api_error":
		return "An API error occurred. Please try again later."
	}
	
	// Fallback to status code-based messages
	switch statusCode {
	case http.StatusBadRequest:
		return "Bad request. Please check your request parameters."
	case http.StatusUnauthorized:
		return "Authentication required. Please provide a valid API key."
	case http.StatusForbidden:
		return "Access forbidden. Your API key may not have the required permissions."
	case http.StatusNotFound:
		return "Resource not found."
	case http.StatusTooManyRequests:
		return "Too many requests. Please slow down your request rate."
	case http.StatusInternalServerError:
		return "Internal server error. Please try again later."
	case http.StatusBadGateway:
		return "Bad gateway. The server is temporarily unavailable."
	case http.StatusServiceUnavailable:
		return "Service unavailable. Please try again later."
	case http.StatusGatewayTimeout:
		return "Gateway timeout. The request took too long to process."
	default:
		return fmt.Sprintf("API error occurred (HTTP %d)", statusCode)
	}
}

// sanitizeValue sanitizes a parameter value to prevent exposing sensitive information.
func sanitizeValue(value string) string {
	if len(value) == 0 {
		return "(empty)"
	}
	
	// If the value looks like a key or token, redact most of it
	if isLikelySecret(value) {
		if len(value) <= 8 {
			return "[REDACTED]"
		}
		return fmt.Sprintf("%s...[REDACTED]", value[:4])
	}
	
	// If the value is very long, truncate it
	if len(value) > 100 {
		return fmt.Sprintf("%s...(truncated)", value[:97])
	}
	
	return value
}

// isLikelySecret checks if a value looks like it might be a secret.
func isLikelySecret(value string) bool {
	lowerValue := fmt.Sprintf("%s", value) // Don't convert to lower to avoid allocations
	
	// Check for patterns that might indicate secrets
	secretPatterns := []string{
		"key", "token", "secret", "password", "auth", "bearer",
		"sk-", "pk-", "api_", "ey", // Common prefixes
	}
	
	for _, pattern := range secretPatterns {
		if len(value) > 10 && containsIgnoreCase(lowerValue, pattern) {
			return true
		}
	}
	
	// If it's a long string of alphanumeric characters, it might be a key
	if len(value) > 20 && isAlphaNumeric(value) {
		return true
	}
	
	return false
}

// containsIgnoreCase checks if a string contains a substring (case-insensitive).
func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive contains check without allocating strings.ToLower
	if len(substr) > len(s) {
		return false
	}
	
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			
			// Convert to lowercase for comparison
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 32
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 32
			}
			
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	
	return false
}

// isAlphaNumeric checks if a string contains only alphanumeric characters.
func isAlphaNumeric(s string) bool {
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}
	return true
}