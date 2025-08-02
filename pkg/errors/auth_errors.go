package errors

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// AuthenticationError represents authentication credential errors.
type AuthenticationError struct {
	*BaseError
	AuthType   string // Type of authentication (api_key, bearer_token, etc.)
	Reason     string // Specific reason for authentication failure
	Suggestion string // Suggestion for resolving the issue
}

// NewAuthenticationError creates a new authentication error.
func NewAuthenticationError(authType, reason string) *AuthenticationError {
	message := "Authentication failed"
	if reason != "" {
		message = fmt.Sprintf("Authentication failed: %s", reason)
	}

	// Provide helpful suggestions based on the reason
	suggestion := getAuthSuggestion(reason)

	err := &AuthenticationError{
		BaseError: NewBaseError(CategoryAuth, SeverityCritical, "AUTHENTICATION_ERROR", message).
			WithHTTPStatus(http.StatusUnauthorized).
			WithRetryable(false), // Auth errors are not retryable without fixing credentials
		AuthType:   authType,
		Reason:     reason,
		Suggestion: suggestion,
	}

	err.WithDetail("auth_type", authType).
		WithDetail("reason", reason).
		WithDetail("suggestion", suggestion)

	return err
}

// AuthorizationError represents permission and access control errors.
type AuthorizationError struct {
	*BaseError
	Resource     string   // The resource being accessed
	Permission   string   // The required permission
	UserRole     string   // The user's current role (if available)
	RequiredRole string   // The required role for access
	Scopes       []string // Required scopes (if applicable)
}

// NewAuthorizationError creates a new authorization error.
func NewAuthorizationError(resource, permission string) *AuthorizationError {
	message := "Access denied"
	if resource != "" {
		message = fmt.Sprintf("Access denied to %s", resource)
		if permission != "" {
			message = fmt.Sprintf("Access denied to %s: %s permission required", resource, permission)
		}
	}

	err := &AuthorizationError{
		BaseError: NewBaseError(CategoryAuth, SeverityCritical, "AUTHORIZATION_ERROR", message).
			WithHTTPStatus(http.StatusForbidden).
			WithRetryable(false), // Authorization errors are not retryable without permission changes
		Resource:   resource,
		Permission: permission,
	}

	err.WithDetail("resource", resource).
		WithDetail("permission", permission)

	return err
}

// WithUserRole sets the user's current role.
func (e *AuthorizationError) WithUserRole(role string) *AuthorizationError {
	e.UserRole = role
	e.WithDetail("user_role", role)
	return e
}

// WithRequiredRole sets the required role for access.
func (e *AuthorizationError) WithRequiredRole(role string) *AuthorizationError {
	e.RequiredRole = role
	e.WithDetail("required_role", role)

	// Update the message to include role information
	if e.Resource != "" && role != "" {
		e.message = fmt.Sprintf("Access denied to %s: %s role required", e.Resource, role)
	}

	return e
}

// WithScopes sets the required scopes.
func (e *AuthorizationError) WithScopes(scopes []string) *AuthorizationError {
	e.Scopes = scopes
	e.WithDetail("required_scopes", scopes)

	// Update the message to include scope information
	if e.Resource != "" && len(scopes) > 0 {
		e.message = fmt.Sprintf("Access denied to %s: requires scopes [%s]",
			e.Resource, strings.Join(scopes, ", "))
	}

	return e
}

// TokenExpiredError represents expired authentication token errors.
type TokenExpiredError struct {
	*BaseError
	TokenType string    // Type of token (access_token, refresh_token, etc.)
	ExpiresAt time.Time // When the token expired
	IssuedAt  time.Time // When the token was issued
}

// NewTokenExpiredError creates a new token expired error.
func NewTokenExpiredError(tokenType string, expiresAt, issuedAt time.Time) *TokenExpiredError {
	message := fmt.Sprintf("%s has expired", tokenType)
	if !expiresAt.IsZero() {
		message = fmt.Sprintf("%s expired at %s", tokenType, expiresAt.Format(time.RFC3339))
	}

	// Token expiration may be retryable if we can refresh the token
	retryable := tokenType == "access_token" // Access tokens can often be refreshed

	err := &TokenExpiredError{
		BaseError: NewBaseError(CategoryAuth, SeverityHigh, "TOKEN_EXPIRED", message).
			WithHTTPStatus(http.StatusUnauthorized).
			WithRetryable(retryable),
		TokenType: tokenType,
		ExpiresAt: expiresAt,
		IssuedAt:  issuedAt,
	}

	err.WithDetail("token_type", tokenType)
	if !expiresAt.IsZero() {
		err.WithDetail("expires_at", expiresAt.Format(time.RFC3339))
	}
	if !issuedAt.IsZero() {
		err.WithDetail("issued_at", issuedAt.Format(time.RFC3339))
	}

	return err
}

// InvalidCredentialsError represents invalid API key or token errors.
type InvalidCredentialsError struct {
	*BaseError
	CredentialType string // Type of credential (api_key, token, etc.)
	Format         string // Expected format if applicable
	Hint           string // Hint about where to find valid credentials
}

// NewInvalidCredentialsError creates a new invalid credentials error.
func NewInvalidCredentialsError(credentialType, format, hint string) *InvalidCredentialsError {
	message := fmt.Sprintf("Invalid %s", credentialType)
	if format != "" {
		message = fmt.Sprintf("Invalid %s format", credentialType)
	}

	err := &InvalidCredentialsError{
		BaseError: NewBaseError(CategoryAuth, SeverityCritical, "INVALID_CREDENTIALS", message).
			WithHTTPStatus(http.StatusUnauthorized).
			WithRetryable(false),
		CredentialType: credentialType,
		Format:         format,
		Hint:           hint,
	}

	err.WithDetail("credential_type", credentialType).
		WithDetail("format", format).
		WithDetail("hint", hint)

	return err
}

// MissingCredentialsError represents missing authentication credentials.
type MissingCredentialsError struct {
	*BaseError
	RequiredCredentials []string // List of required credential types
	ProvidedCredentials []string // List of provided credential types
}

// NewMissingCredentialsError creates a new missing credentials error.
func NewMissingCredentialsError(required, provided []string) *MissingCredentialsError {
	message := "Authentication credentials missing"
	if len(required) > 0 {
		message = fmt.Sprintf("Missing required credentials: %s", strings.Join(required, ", "))
	}

	err := &MissingCredentialsError{
		BaseError: NewBaseError(CategoryAuth, SeverityCritical, "MISSING_CREDENTIALS", message).
			WithHTTPStatus(http.StatusUnauthorized).
			WithRetryable(false),
		RequiredCredentials: required,
		ProvidedCredentials: provided,
	}

	err.WithDetail("required_credentials", required).
		WithDetail("provided_credentials", provided)

	return err
}

// APIKeyError represents API key specific errors.
type APIKeyError struct {
	*AuthenticationError
	KeyFormat string // Expected key format
	KeyPrefix string // Expected key prefix (e.g., "sk-")
	KeySource string // Where the key was sourced from (env, config, etc.)
}

// NewAPIKeyError creates a new API key error.
func NewAPIKeyError(reason, keyFormat, keyPrefix, keySource string) *APIKeyError {
	authErr := NewAuthenticationError("api_key", reason)

	err := &APIKeyError{
		AuthenticationError: authErr,
		KeyFormat:           keyFormat,
		KeyPrefix:           keyPrefix,
		KeySource:           keySource,
	}

	err.WithDetail("key_format", keyFormat).
		WithDetail("key_prefix", keyPrefix).
		WithDetail("key_source", keySource)

	return err
}

// BearerTokenError represents bearer token specific errors.
type BearerTokenError struct {
	*AuthenticationError
	TokenFormat string    // Expected token format (JWT, opaque, etc.)
	TokenSource string    // Where the token was sourced from
	ExpiresAt   time.Time // Token expiration time if available
}

// NewBearerTokenError creates a new bearer token error.
func NewBearerTokenError(reason, tokenFormat, tokenSource string) *BearerTokenError {
	authErr := NewAuthenticationError("bearer_token", reason)

	err := &BearerTokenError{
		AuthenticationError: authErr,
		TokenFormat:         tokenFormat,
		TokenSource:         tokenSource,
	}

	err.WithDetail("token_format", tokenFormat).
		WithDetail("token_source", tokenSource)

	return err
}

// WithExpiresAt sets the token expiration time.
func (e *BearerTokenError) WithExpiresAt(expiresAt time.Time) *BearerTokenError {
	e.ExpiresAt = expiresAt
	if !expiresAt.IsZero() {
		e.WithDetail("expires_at", expiresAt.Format(time.RFC3339))
	}
	return e
}

// Helper functions for authentication errors

// getAuthSuggestion provides helpful suggestions based on authentication failure reasons.
func getAuthSuggestion(reason string) string {
	lowerReason := strings.ToLower(reason)

	switch {
	case strings.Contains(lowerReason, "api key"):
		return "Verify your API key is correct and has the required permissions"
	case strings.Contains(lowerReason, "token"):
		if strings.Contains(lowerReason, "expired") {
			return "Refresh your authentication token and try again"
		}
		return "Verify your authentication token is valid and properly formatted"
	case strings.Contains(lowerReason, "missing"):
		return "Provide valid authentication credentials (API key or bearer token)"
	case strings.Contains(lowerReason, "invalid"):
		return "Check that your credentials are correctly formatted and have not been revoked"
	case strings.Contains(lowerReason, "permission"):
		return "Contact your administrator to request the necessary permissions"
	case strings.Contains(lowerReason, "rate limit"):
		return "Wait before retrying or upgrade your plan for higher rate limits"
	case strings.Contains(lowerReason, "quota"):
		return "Check your usage quota and upgrade your plan if necessary"
	default:
		return "Verify your authentication credentials and try again"
	}
}

// ClassifyAuthError analyzes an HTTP response and creates an appropriate auth error.
func ClassifyAuthError(statusCode int, responseBody []byte, headers http.Header) SDKError {
	switch statusCode {
	case http.StatusUnauthorized:
		return classifyUnauthorizedError(responseBody, headers)
	case http.StatusForbidden:
		return classifyForbiddenError(responseBody, headers)
	default:
		return NewAuthenticationError("unknown", fmt.Sprintf("HTTP %d", statusCode)).
			BaseError.WithHTTPStatus(statusCode)
	}
}

// classifyUnauthorizedError analyzes 401 responses to create specific auth errors.
func classifyUnauthorizedError(responseBody []byte, headers http.Header) SDKError {
	// Try to parse error details from response body
	bodyStr := strings.ToLower(string(responseBody))

	// Check for specific error patterns
	switch {
	case strings.Contains(bodyStr, "expired"):
		return NewTokenExpiredError("access_token", time.Time{}, time.Time{})
	case strings.Contains(bodyStr, "invalid api key"):
		return NewAPIKeyError("invalid API key", "", "sk-", "")
	case strings.Contains(bodyStr, "missing api key"):
		return NewMissingCredentialsError([]string{"api_key"}, []string{})
	case strings.Contains(bodyStr, "invalid token"):
		return NewBearerTokenError("invalid token", "", "")
	case strings.Contains(bodyStr, "malformed"):
		return NewInvalidCredentialsError("credentials", "malformed", "")
	default:
		return NewAuthenticationError("unknown", "authentication failed")
	}
}

// classifyForbiddenError analyzes 403 responses to create specific auth errors.
func classifyForbiddenError(responseBody []byte, headers http.Header) SDKError {
	bodyStr := strings.ToLower(string(responseBody))

	// Check for specific permission patterns
	switch {
	case strings.Contains(bodyStr, "insufficient"):
		return NewAuthorizationError("resource", "insufficient permissions")
	case strings.Contains(bodyStr, "access denied"):
		return NewAuthorizationError("resource", "access denied")
	case strings.Contains(bodyStr, "forbidden"):
		return NewAuthorizationError("resource", "forbidden")
	case strings.Contains(bodyStr, "scope"):
		return NewAuthorizationError("resource", "insufficient scope")
	default:
		return NewAuthorizationError("resource", "permission denied")
	}
}

// IsAuthenticationError checks if an error is an authentication error.
func IsAuthenticationError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific auth error types
	var authErr *AuthenticationError
	var tokenErr *TokenExpiredError
	var credErr *InvalidCredentialsError
	var missingErr *MissingCredentialsError

	return AsError(err, &authErr) || AsError(err, &tokenErr) ||
		AsError(err, &credErr) || AsError(err, &missingErr)
}

// IsAuthorizationError checks if an error is an authorization error.
func IsAuthorizationError(err error) bool {
	if err == nil {
		return false
	}

	var authzErr *AuthorizationError
	return AsError(err, &authzErr)
}

// AsError is a wrapper around errors.As for convenience.
func AsError(err error, target any) bool {
	// This would use errors.As in a real implementation
	// For now, we'll use a simple type assertion approach
	switch target := target.(type) {
	case **AuthenticationError:
		if authErr, ok := err.(*AuthenticationError); ok {
			*target = authErr
			return true
		}
	case **AuthorizationError:
		if authzErr, ok := err.(*AuthorizationError); ok {
			*target = authzErr
			return true
		}
	case **TokenExpiredError:
		if tokenErr, ok := err.(*TokenExpiredError); ok {
			*target = tokenErr
			return true
		}
	case **InvalidCredentialsError:
		if credErr, ok := err.(*InvalidCredentialsError); ok {
			*target = credErr
			return true
		}
	case **MissingCredentialsError:
		if missingErr, ok := err.(*MissingCredentialsError); ok {
			*target = missingErr
			return true
		}
	}
	return false
}
