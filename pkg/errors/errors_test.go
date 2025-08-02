package errors

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"
)

// Test BaseError implementation
func TestBaseError(t *testing.T) {
	t.Run("basic creation", func(t *testing.T) {
		err := NewBaseError(CategoryAPI, SeverityHigh, "TEST_CODE", "test message")

		if err.Category() != CategoryAPI {
			t.Errorf("Expected category %s, got %s", CategoryAPI, err.Category())
		}
		if err.Severity() != SeverityHigh {
			t.Errorf("Expected severity %s, got %s", SeverityHigh, err.Severity())
		}
		if err.Code() != "TEST_CODE" {
			t.Errorf("Expected code TEST_CODE, got %s", err.Code())
		}
		if err.Message() != "test message" {
			t.Errorf("Expected message 'test message', got %s", err.Message())
		}
		if err.Error() != "test message" {
			t.Errorf("Expected error 'test message', got %s", err.Error())
		}
		if err.IsRetryable() {
			t.Error("Expected error to not be retryable by default")
		}
	})

	t.Run("with cause", func(t *testing.T) {
		cause := fmt.Errorf("underlying error")
		err := NewBaseError(CategoryNetwork, SeverityMedium, "NET_ERROR", "network failed").
			WithCause(cause)

		if err.Unwrap() != cause {
			t.Error("Expected Unwrap() to return the cause")
		}
		if !strings.Contains(err.Error(), "underlying error") {
			t.Errorf("Expected error message to contain cause, got: %s", err.Error())
		}
	})

	t.Run("with details", func(t *testing.T) {
		err := NewBaseError(CategoryValidation, SeverityLow, "VALID_ERROR", "validation failed").
			WithDetail("field", "username").
			WithDetail("value", "invalid")

		details := err.Details()
		if details["field"] != "username" {
			t.Errorf("Expected field detail to be 'username', got %v", details["field"])
		}
		if details["value"] != "invalid" {
			t.Errorf("Expected value detail to be 'invalid', got %v", details["value"])
		}
	})

	t.Run("with HTTP status and request ID", func(t *testing.T) {
		err := NewBaseError(CategoryAPI, SeverityHigh, "API_ERROR", "API failed").
			WithHTTPStatus(500).
			WithRequestID("req-123").
			WithRetryable(true)

		if err.HTTPStatusCode() != 500 {
			t.Errorf("Expected HTTP status 500, got %d", err.HTTPStatusCode())
		}
		if err.RequestID() != "req-123" {
			t.Errorf("Expected request ID 'req-123', got %s", err.RequestID())
		}
		if !err.IsRetryable() {
			t.Error("Expected error to be retryable")
		}
	})
}

// Test error wrapping and unwrapping
func TestErrorWrapping(t *testing.T) {
	t.Run("wrap with Go 1.13+ errors", func(t *testing.T) {
		originalErr := fmt.Errorf("original error")
		sdkErr := WrapError(originalErr, CategoryNetwork, "WRAP_ERROR", "wrapped error")

		// Test that errors.Is works
		if !errors.Is(sdkErr, originalErr) {
			t.Error("Expected errors.Is to work with wrapped error")
		}

		// Test that errors.As works
		var baseErr *BaseError
		if !errors.As(sdkErr, &baseErr) {
			t.Error("Expected errors.As to work with wrapped error")
		}

		// Test unwrapping
		if sdkErr.Unwrap() != originalErr {
			t.Error("Expected Unwrap() to return original error")
		}
	})

	t.Run("wrap preserves properties", func(t *testing.T) {
		originalErr := NewAPIError(429, "RATE_LIMIT", "rate_limit_error", "Rate limited").
			WithRequestID("req-456").
			WithRetryable(true)

		wrappedErr := WrapError(originalErr, CategoryNetwork, "NETWORK_WRAP", "Network wrapper")

		// Should preserve request ID and retryability
		if wrappedErr.RequestID() != "req-456" {
			t.Errorf("Expected wrapped error to preserve request ID, got %s", wrappedErr.RequestID())
		}
		if !wrappedErr.IsRetryable() {
			t.Error("Expected wrapped error to preserve retryability")
		}
	})
}

// Test utility functions
func TestUtilityFunctions(t *testing.T) {
	t.Run("IsRetryable", func(t *testing.T) {
		retryableErr := NewBaseError(CategoryNetwork, SeverityMedium, "NET_ERROR", "network error").
			WithRetryable(true)
		nonRetryableErr := NewBaseError(CategoryValidation, SeverityHigh, "VALID_ERROR", "validation error").
			WithRetryable(false)

		if !IsRetryable(retryableErr) {
			t.Error("Expected retryable error to be detected as retryable")
		}
		if IsRetryable(nonRetryableErr) {
			t.Error("Expected non-retryable error to be detected as non-retryable")
		}
		if IsRetryable(nil) {
			t.Error("Expected nil error to be non-retryable")
		}
	})

	t.Run("GetCategory", func(t *testing.T) {
		err := NewBaseError(CategoryAPI, SeverityMedium, "API_ERROR", "API error")

		if GetCategory(err) != CategoryAPI {
			t.Errorf("Expected category %s, got %s", CategoryAPI, GetCategory(err))
		}
		if GetCategory(nil) != "" {
			t.Errorf("Expected empty category for nil error, got %s", GetCategory(nil))
		}

		// Test with non-SDK error
		genericErr := fmt.Errorf("generic error")
		if GetCategory(genericErr) != CategoryInternal {
			t.Errorf("Expected internal category for generic error, got %s", GetCategory(genericErr))
		}
	})

	t.Run("GetSeverity", func(t *testing.T) {
		err := NewBaseError(CategorySecurity, SeverityCritical, "SEC_ERROR", "Security error")

		if GetSeverity(err) != SeverityCritical {
			t.Errorf("Expected severity %s, got %s", SeverityCritical, GetSeverity(err))
		}
		if GetSeverity(nil) != SeverityLow {
			t.Errorf("Expected low severity for nil error, got %s", GetSeverity(nil))
		}
	})
}

// Test HTTP error creation
func TestHTTPErrorFromStatus(t *testing.T) {
	testCases := []struct {
		statusCode int
		message    string
		retryable  bool
		severity   ErrorSeverity
	}{
		{500, "", true, SeverityHigh},
		{502, "", true, SeverityHigh},
		{503, "", true, SeverityHigh},
		{429, "", true, SeverityMedium},
		{400, "", false, SeverityCritical},
		{401, "", false, SeverityCritical},
		{404, "", false, SeverityCritical},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("status_%d", tc.statusCode), func(t *testing.T) {
			err := HTTPErrorFromStatus(tc.statusCode, tc.message)

			if err.HTTPStatusCode() != tc.statusCode {
				t.Errorf("Expected HTTP status %d, got %d", tc.statusCode, err.HTTPStatusCode())
			}
			if err.IsRetryable() != tc.retryable {
				t.Errorf("Expected retryable=%v for status %d, got %v", tc.retryable, tc.statusCode, err.IsRetryable())
			}
			if err.Severity() != tc.severity {
				t.Errorf("Expected severity %s for status %d, got %s", tc.severity, tc.statusCode, err.Severity())
			}
		})
	}
}

// Test API errors
func TestAPIErrors(t *testing.T) {
	t.Run("basic API error", func(t *testing.T) {
		err := NewAPIError(400, "invalid_request", "invalid_request_error", "Invalid request")

		if err.HTTPStatusCode() != 400 {
			t.Errorf("Expected HTTP status 400, got %d", err.HTTPStatusCode())
		}
		if err.APICode != "invalid_request" {
			t.Errorf("Expected API code 'invalid_request', got %s", err.APICode)
		}
		if err.APIType != "invalid_request_error" {
			t.Errorf("Expected API type 'invalid_request_error', got %s", err.APIType)
		}
	})

	t.Run("rate limit error", func(t *testing.T) {
		retryAfter := 60 * time.Second
		err := NewRateLimitError(retryAfter, 100, 0, time.Now().Add(time.Hour))

		if err.RetryAfter != retryAfter {
			t.Errorf("Expected retry after %v, got %v", retryAfter, err.RetryAfter)
		}
		if err.Limit != 100 {
			t.Errorf("Expected limit 100, got %d", err.Limit)
		}
		if !err.IsRetryable() {
			t.Error("Expected rate limit error to be retryable")
		}
	})

	t.Run("quota exceeded error", func(t *testing.T) {
		err := NewQuotaExceededError("tokens", 1000, 1000, time.Now().Add(24*time.Hour))

		if err.QuotaType != "tokens" {
			t.Errorf("Expected quota type 'tokens', got %s", err.QuotaType)
		}
		if err.Used != 1000 {
			t.Errorf("Expected used 1000, got %d", err.Used)
		}
		if err.Limit != 1000 {
			t.Errorf("Expected limit 1000, got %d", err.Limit)
		}
		if err.IsRetryable() {
			t.Error("Expected quota exceeded error to not be retryable")
		}
	})

	t.Run("model unavailable error", func(t *testing.T) {
		availableModels := []string{"claude-3-sonnet", "claude-3-haiku"}
		err := NewModelUnavailableError("claude-4", "model not released", "claude-3-sonnet", availableModels)

		if err.ModelID != "claude-4" {
			t.Errorf("Expected model ID 'claude-4', got %s", err.ModelID)
		}
		if err.UnavailableReason != "model not released" {
			t.Errorf("Expected reason 'model not released', got %s", err.UnavailableReason)
		}
		if err.SuggestedModel != "claude-3-sonnet" {
			t.Errorf("Expected suggested model 'claude-3-sonnet', got %s", err.SuggestedModel)
		}
	})

	t.Run("parse API error from response", func(t *testing.T) {
		responseBody := `{"error": {"type": "rate_limit_error", "message": "Rate limit exceeded"}}`
		err := ParseAPIErrorFromResponse(429, []byte(responseBody))

		if err.HTTPStatusCode() != 429 {
			t.Errorf("Expected HTTP status 429, got %d", err.HTTPStatusCode())
		}
		if err.APIType != "rate_limit_error" {
			t.Errorf("Expected API type 'rate_limit_error', got %s", err.APIType)
		}
		if err.APIMessage != "Rate limit exceeded" {
			t.Errorf("Expected API message 'Rate limit exceeded', got %s", err.APIMessage)
		}
	})
}

// Test network errors
func TestNetworkErrors(t *testing.T) {
	t.Run("basic network error", func(t *testing.T) {
		cause := fmt.Errorf("connection failed")
		err := NewNetworkError("dial", "example.com:443", cause)

		if err.Operation != "dial" {
			t.Errorf("Expected operation 'dial', got %s", err.Operation)
		}
		if err.Address != "example.com:443" {
			t.Errorf("Expected address 'example.com:443', got %s", err.Address)
		}
		if err.Unwrap() != cause {
			t.Error("Expected error to wrap the cause")
		}
	})

	t.Run("timeout error", func(t *testing.T) {
		timeout := 30 * time.Second
		elapsed := 35 * time.Second
		err := NewTimeoutError("connect", timeout, elapsed)

		if err.Operation != "connect" {
			t.Errorf("Expected operation 'connect', got %s", err.Operation)
		}
		if err.Timeout != timeout {
			t.Errorf("Expected timeout %v, got %v", timeout, err.Timeout)
		}
		if err.Elapsed != elapsed {
			t.Errorf("Expected elapsed %v, got %v", elapsed, err.Elapsed)
		}
		if !err.IsRetryable() {
			t.Error("Expected timeout error to be retryable")
		}
	})

	t.Run("connection error", func(t *testing.T) {
		cause := fmt.Errorf("connection refused")
		err := NewConnectionError("example.com:443", "connection refused", cause)

		if err.Address != "example.com:443" {
			t.Errorf("Expected address 'example.com:443', got %s", err.Address)
		}
		if err.Reason != "connection refused" {
			t.Errorf("Expected reason 'connection refused', got %s", err.Reason)
		}
	})

	t.Run("TLS error", func(t *testing.T) {
		cause := &tls.CertificateVerificationError{}
		err := NewTLSError("example.com:443", "certificate verification failed", cause)

		if err.Address != "example.com:443" {
			t.Errorf("Expected address 'example.com:443', got %s", err.Address)
		}
		if err.Reason != "certificate verification failed" {
			t.Errorf("Expected reason 'certificate verification failed', got %s", err.Reason)
		}
		if err.IsRetryable() {
			t.Error("Expected TLS error to not be retryable")
		}
	})

	t.Run("DNS error", func(t *testing.T) {
		cause := fmt.Errorf("no such host")
		err := NewDNSError("nonexistent.example.com", "A", cause)

		if err.Hostname != "nonexistent.example.com" {
			t.Errorf("Expected hostname 'nonexistent.example.com', got %s", err.Hostname)
		}
		if err.DNSType != "A" {
			t.Errorf("Expected DNS type 'A', got %s", err.DNSType)
		}
		if !err.IsRetryable() {
			t.Error("Expected DNS error to be retryable")
		}
	})

	t.Run("classify network errors", func(t *testing.T) {
		// Context cancellation
		err := ClassifyNetworkError(context.Canceled)
		if err.Category() != CategoryNetwork {
			t.Errorf("Expected network category, got %s", err.Category())
		}
		if err.IsRetryable() {
			t.Error("Expected context cancellation to not be retryable")
		}

		// Context timeout
		err = ClassifyNetworkError(context.DeadlineExceeded)
		if _, ok := err.(*BaseError); !ok {
			t.Error("Expected timeout error to be wrapped in BaseError")
		}

		// URL error
		urlErr := &url.Error{Op: "Get", URL: "http://example.com", Err: fmt.Errorf("connection refused")}
		err = ClassifyNetworkError(urlErr)
		if err.Category() != CategoryNetwork {
			t.Errorf("Expected network category, got %s", err.Category())
		}
	})
}

// Test validation errors
func TestValidationErrors(t *testing.T) {
	t.Run("basic validation error", func(t *testing.T) {
		err := NewValidationError("username", "invalid_user", "must be alphanumeric", "Invalid username")

		if err.Field != "username" {
			t.Errorf("Expected field 'username', got %s", err.Field)
		}
		if err.Value != "invalid_user" {
			t.Errorf("Expected value 'invalid_user', got %s", err.Value)
		}
		if err.Constraint != "must be alphanumeric" {
			t.Errorf("Expected constraint 'must be alphanumeric', got %s", err.Constraint)
		}
		if err.IsRetryable() {
			t.Error("Expected validation error to not be retryable")
		}
	})

	t.Run("validation error with violations", func(t *testing.T) {
		violations := []ValidationViolation{
			{Field: "username", Code: "required", Message: "Username is required"},
			{Field: "email", Code: "invalid_format", Message: "Invalid email format"},
		}
		err := NewValidationErrorWithViolations(violations)

		if len(err.Violations) != 2 {
			t.Errorf("Expected 2 violations, got %d", len(err.Violations))
		}
		if err.Violations[0].Field != "username" {
			t.Errorf("Expected first violation field 'username', got %s", err.Violations[0].Field)
		}
	})

	t.Run("request validation error", func(t *testing.T) {
		violations := []ValidationViolation{
			{Field: "model", Code: "required", Message: "Model is required"},
		}
		err := NewRequestValidationError("chat_completion", "POST", "/v1/chat/completions", violations)

		if err.RequestType != "chat_completion" {
			t.Errorf("Expected request type 'chat_completion', got %s", err.RequestType)
		}
		if err.Method != "POST" {
			t.Errorf("Expected method 'POST', got %s", err.Method)
		}
		if err.Endpoint != "/v1/chat/completions" {
			t.Errorf("Expected endpoint '/v1/chat/completions', got %s", err.Endpoint)
		}
	})

	t.Run("parameter validation error", func(t *testing.T) {
		err := NewParameterValidationError("max_tokens", "integer", 0, "must be greater than 0").
			WithMinValue(1).
			WithMaxValue(4096)

		if err.ParameterName != "max_tokens" {
			t.Errorf("Expected parameter name 'max_tokens', got %s", err.ParameterName)
		}
		if err.ParameterType != "integer" {
			t.Errorf("Expected parameter type 'integer', got %s", err.ParameterType)
		}
		if err.MinValue != 1 {
			t.Errorf("Expected min value 1, got %v", err.MinValue)
		}
		if err.MaxValue != 4096 {
			t.Errorf("Expected max value 4096, got %v", err.MaxValue)
		}
	})

	t.Run("validation helper functions", func(t *testing.T) {
		// Test required validation
		violation := ValidateRequired("username", "")
		if violation == nil {
			t.Error("Expected required validation to fail for empty string")
		}
		if violation.Code != "required" {
			t.Errorf("Expected violation code 'required', got %s", violation.Code)
		}

		// Test string length validation
		violation = ValidateStringLength("description", "short", 10, 100)
		if violation == nil {
			t.Error("Expected string length validation to fail for short string")
		}
		if violation.Code != "min_length" {
			t.Errorf("Expected violation code 'min_length', got %s", violation.Code)
		}

		// Test numeric range validation
		violation = ValidateNumericRange("temperature", 2.0, 0.0, 1.0)
		if violation == nil {
			t.Error("Expected numeric range validation to fail for out-of-range value")
		}
		if violation.Code != "max_value" {
			t.Errorf("Expected violation code 'max_value', got %s", violation.Code)
		}

		// Test enum validation
		violation = ValidateEnum("model", "invalid-model", []string{"claude-3-sonnet", "claude-3-haiku"})
		if violation == nil {
			t.Error("Expected enum validation to fail for invalid value")
		}
		if violation.Code != "invalid_enum" {
			t.Errorf("Expected violation code 'invalid_enum', got %s", violation.Code)
		}
	})
}

// Test authentication errors
func TestAuthenticationErrors(t *testing.T) {
	t.Run("authentication error", func(t *testing.T) {
		err := NewAuthenticationError("api_key", "invalid API key")

		if err.AuthType != "api_key" {
			t.Errorf("Expected auth type 'api_key', got %s", err.AuthType)
		}
		if err.Reason != "invalid API key" {
			t.Errorf("Expected reason 'invalid API key', got %s", err.Reason)
		}
		if err.HTTPStatusCode() != 401 {
			t.Errorf("Expected HTTP status 401, got %d", err.HTTPStatusCode())
		}
		if err.IsRetryable() {
			t.Error("Expected authentication error to not be retryable")
		}
	})

	t.Run("authorization error", func(t *testing.T) {
		err := NewAuthorizationError("messages", "read").
			WithUserRole("user").
			WithRequiredRole("admin").
			WithScopes([]string{"messages:read", "messages:write"})

		if err.Resource != "messages" {
			t.Errorf("Expected resource 'messages', got %s", err.Resource)
		}
		if err.Permission != "read" {
			t.Errorf("Expected permission 'read', got %s", err.Permission)
		}
		if err.UserRole != "user" {
			t.Errorf("Expected user role 'user', got %s", err.UserRole)
		}
		if err.RequiredRole != "admin" {
			t.Errorf("Expected required role 'admin', got %s", err.RequiredRole)
		}
		if len(err.Scopes) != 2 {
			t.Errorf("Expected 2 scopes, got %d", len(err.Scopes))
		}
	})

	t.Run("token expired error", func(t *testing.T) {
		expiresAt := time.Now().Add(-time.Hour)
		issuedAt := time.Now().Add(-2 * time.Hour)
		err := NewTokenExpiredError("access_token", expiresAt, issuedAt)

		if err.TokenType != "access_token" {
			t.Errorf("Expected token type 'access_token', got %s", err.TokenType)
		}
		if !err.ExpiresAt.Equal(expiresAt) {
			t.Errorf("Expected expires at %v, got %v", expiresAt, err.ExpiresAt)
		}
		if err.IsRetryable() != true {
			t.Error("Expected access token expiration to be retryable")
		}
	})

	t.Run("API key error", func(t *testing.T) {
		err := NewAPIKeyError("invalid format", "sk-...", "sk-", "environment")

		if err.KeyFormat != "sk-..." {
			t.Errorf("Expected key format 'sk-...', got %s", err.KeyFormat)
		}
		if err.KeyPrefix != "sk-" {
			t.Errorf("Expected key prefix 'sk-', got %s", err.KeyPrefix)
		}
		if err.KeySource != "environment" {
			t.Errorf("Expected key source 'environment', got %s", err.KeySource)
		}
	})

	t.Run("bearer token error", func(t *testing.T) {
		expiresAt := time.Now().Add(time.Hour)
		err := NewBearerTokenError("malformed token", "JWT", "header").
			WithExpiresAt(expiresAt)

		if err.TokenFormat != "JWT" {
			t.Errorf("Expected token format 'JWT', got %s", err.TokenFormat)
		}
		if err.TokenSource != "header" {
			t.Errorf("Expected token source 'header', got %s", err.TokenSource)
		}
		if !err.ExpiresAt.Equal(expiresAt) {
			t.Errorf("Expected expires at %v, got %v", expiresAt, err.ExpiresAt)
		}
	})
}

// Test configuration errors
func TestConfigurationErrors(t *testing.T) {
	t.Run("configuration error", func(t *testing.T) {
		err := NewConfigurationError("base_url", "Invalid base URL format")

		if err.Field != "base_url" {
			t.Errorf("Expected field 'base_url', got %s", err.Field)
		}
		if err.Category() != CategoryConfiguration {
			t.Errorf("Expected category %s, got %s", CategoryConfiguration, err.Category())
		}
		if err.Severity() != SeverityHigh {
			t.Errorf("Expected severity %s, got %s", SeverityHigh, err.Severity())
		}
	})

	t.Run("internal error", func(t *testing.T) {
		err := NewInternalError("INTERNAL_BUG", "Unexpected nil pointer")

		if err.Code() != "INTERNAL_BUG" {
			t.Errorf("Expected code 'INTERNAL_BUG', got %s", err.Code())
		}
		if err.Category() != CategoryInternal {
			t.Errorf("Expected category %s, got %s", CategoryInternal, err.Category())
		}
	})
}

// Test error serialization and formatting
func TestErrorSerialization(t *testing.T) {
	t.Run("error string representation", func(t *testing.T) {
		err := NewAPIError(429, "rate_limit_error", "rate_limit_error", "Rate limit exceeded").
			WithRequestID("req-123").
			WithRetryable(true).
			WithDetail("retry_after", 60)

		str := err.String()
		expectedParts := []string{
			"Category: api",
			"Severity: medium",
			"Code: rate_limit_error",
			"Message: Rate limit exceeded",
			"HTTP Status: 429",
			"Request ID: req-123",
			"Retryable: true",
		}

		for _, part := range expectedParts {
			if !strings.Contains(str, part) {
				t.Errorf("Expected string representation to contain '%s', got: %s", part, str)
			}
		}
	})

	t.Run("details serialization", func(t *testing.T) {
		details := map[string]interface{}{
			"field":  "username",
			"value":  "test",
			"number": 42,
		}
		err := NewValidationError("test", "test", "test", "test").WithDetails(details)

		retrievedDetails := err.Details()
		if retrievedDetails["field"] != "username" {
			t.Errorf("Expected field 'username', got %v", retrievedDetails["field"])
		}
		if retrievedDetails["number"] != 42 {
			t.Errorf("Expected number 42, got %v", retrievedDetails["number"])
		}
	})
}

// Test security features
func TestSecurityFeatures(t *testing.T) {
	t.Run("sanitize validation values", func(t *testing.T) {
		// Test API key sanitization
		sanitized := sanitizeValidationValue("sk-1234567890abcdef1234567890abcdef")
		if !strings.Contains(sanitized, "[REDACTED]") {
			t.Errorf("Expected API key to be redacted, got: %s", sanitized)
		}

		// Test long value truncation
		longValue := strings.Repeat("a", 250)
		sanitized = sanitizeValidationValue(longValue)
		if len(sanitized) >= len(longValue) {
			t.Errorf("Expected long value to be truncated, got length: %d", len(sanitized))
		}

		// Test normal value passthrough
		normalValue := "normal_value"
		sanitized = sanitizeValidationValue(normalValue)
		if sanitized != normalValue {
			t.Errorf("Expected normal value to pass through, got: %s", sanitized)
		}
	})

	t.Run("sanitize addresses", func(t *testing.T) {
		// Test URL with credentials
		addr := "https://user:pass@example.com:443/path"
		sanitized := sanitizeAddress(addr)
		if strings.Contains(sanitized, "pass") {
			t.Errorf("Expected password to be redacted, got: %s", sanitized)
		}
		if !strings.Contains(sanitized, "[REDACTED]") {
			t.Errorf("Expected credentials to be redacted, got: %s", sanitized)
		}

		// Test normal host:port
		addr = "example.com:443"
		sanitized = sanitizeAddress(addr)
		if sanitized != addr {
			t.Errorf("Expected normal address to pass through, got: %s", sanitized)
		}
	})

	t.Run("no sensitive data in errors", func(t *testing.T) {
		// Create errors with potentially sensitive data
		apiKey := "sk-1234567890abcdef1234567890abcdef"

		// API key error should not expose the key
		err := NewAPIKeyError("invalid format", "", "", "").
			WithDetail("raw_key", apiKey)

		errorStr := err.Error()
		if strings.Contains(errorStr, apiKey) {
			t.Errorf("Error message should not contain raw API key: %s", errorStr)
		}

		// Validation error should sanitize values
		validationErr := NewValidationError("api_key", apiKey, "invalid format", "")
		if strings.Contains(validationErr.Value, "1234567890abcdef") {
			t.Errorf("Validation error should sanitize API key values: %s", validationErr.Value)
		}
	})
}

// Benchmark tests
func BenchmarkErrorCreation(b *testing.B) {
	b.Run("BaseError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewBaseError(CategoryAPI, SeverityMedium, "BENCH_ERROR", "benchmark error")
		}
	})

	b.Run("APIError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewAPIError(400, "invalid_request", "invalid_request_error", "Invalid request")
		}
	})

	b.Run("NetworkError", func(b *testing.B) {
		cause := fmt.Errorf("connection failed")
		for i := 0; i < b.N; i++ {
			_ = NewNetworkError("dial", "example.com:443", cause)
		}
	})

	b.Run("ValidationError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewValidationError("field", "value", "constraint", "message")
		}
	})
}

func BenchmarkErrorWrapping(b *testing.B) {
	originalErr := fmt.Errorf("original error")

	b.Run("WrapError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = WrapError(originalErr, CategoryNetwork, "WRAP_ERROR", "wrapped error")
		}
	})

	b.Run("ErrorsIs", func(b *testing.B) {
		wrappedErr := WrapError(originalErr, CategoryNetwork, "WRAP_ERROR", "wrapped error")
		for i := 0; i < b.N; i++ {
			_ = errors.Is(wrappedErr, originalErr)
		}
	})
}
