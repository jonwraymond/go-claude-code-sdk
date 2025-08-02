package types

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// Authenticator defines the interface for handling authentication with the Claude Code API.
// Different authentication methods can implement this interface.
type Authenticator interface {
	// Authenticate adds authentication information to the HTTP request.
	// This method is called before each API request to ensure proper authentication.
	Authenticate(ctx context.Context, req *http.Request) error

	// IsValid returns true if the current authentication credentials are valid.
	// This can be used to check if reauthentication is needed.
	IsValid(ctx context.Context) bool

	// Refresh attempts to refresh the authentication credentials if supported.
	// Returns an error if refresh is not supported or fails.
	Refresh(ctx context.Context) error

	// Type returns the type of authentication being used.
	Type() AuthType
}

// AuthType represents the different types of authentication supported by the SDK.
type AuthType string

// AuthInfo contains information about the current authentication state
type AuthInfo struct {
	// Type indicates the type of authentication being used
	Type AuthType

	// Headers contains the authentication headers to be added to requests
	Headers map[string]string

	// CreatedAt is when the authentication was established
	CreatedAt time.Time

	// ExpiresAt is when the authentication expires (optional)
	ExpiresAt *time.Time

	// Metadata contains additional authentication-specific information
	Metadata map[string]interface{}
}

const (
	// AuthTypeAPIKey indicates API key-based authentication
	AuthTypeAPIKey AuthType = "api_key"

	// AuthTypeSession indicates session-based authentication
	AuthTypeSession AuthType = "session"

	// AuthTypeSubscription indicates subscription-based authentication
	AuthTypeSubscription AuthType = "subscription"

	// AuthTypeOAuth indicates OAuth-based authentication (future)
	AuthTypeOAuth AuthType = "oauth"

	// AuthTypeBearer indicates bearer token authentication
	AuthTypeBearer AuthType = "bearer"
)

// APIKeyAuth implements API key-based authentication for the Claude Code API.
// This is the primary authentication method currently supported.
//
// Example usage:
//
//	auth := &types.APIKeyAuth{
//		APIKey: "your-api-key-here",
//	}
//	client := claude.NewClient(ctx, &types.Config{
//		Auth: auth,
//	})
type APIKeyAuth struct {
	// APIKey is the API key provided by Anthropic for Claude Code access
	APIKey string

	// HeaderName is the HTTP header name to use for the API key.
	// Defaults to "Authorization" if empty.
	HeaderName string

	// Prefix is the prefix to add before the API key in the header.
	// Defaults to "Bearer " if empty.
	Prefix string
}

// Authenticate adds the API key to the request headers.
func (a *APIKeyAuth) Authenticate(ctx context.Context, req *http.Request) error {
	if a.APIKey == "" {
		return &AuthError{
			Type:    "missing_api_key",
			Message: "API key is required but not provided",
		}
	}

	headerName := a.HeaderName
	if headerName == "" {
		headerName = "Authorization"
	}

	prefix := a.Prefix
	if prefix == "" {
		prefix = "Bearer "
	}

	req.Header.Set(headerName, prefix+a.APIKey)
	return nil
}

// IsValid checks if the API key is present (basic validation).
func (a *APIKeyAuth) IsValid(ctx context.Context) bool {
	return a.APIKey != ""
}

// Refresh is not supported for API key authentication.
func (a *APIKeyAuth) Refresh(ctx context.Context) error {
	return &AuthError{
		Type:    "refresh_not_supported",
		Message: "API key authentication does not support refresh",
	}
}

// Type returns the authentication type.
func (a *APIKeyAuth) Type() AuthType {
	return AuthTypeAPIKey
}

// BearerTokenAuth implements bearer token-based authentication.
// This can be used for various token-based authentication schemes.
type BearerTokenAuth struct {
	// Token is the bearer token to use for authentication
	Token string

	// ExpiresAt is the timestamp when the token expires (optional)
	ExpiresAt *time.Time

	// RefreshToken is used to refresh the access token (optional)
	RefreshToken string

	// RefreshFunc is called to refresh the token when needed (optional)
	RefreshFunc func(ctx context.Context, refreshToken string) (*TokenResponse, error)
}

// Authenticate adds the bearer token to the request headers.
func (b *BearerTokenAuth) Authenticate(ctx context.Context, req *http.Request) error {
	if b.Token == "" {
		return &AuthError{
			Type:    "missing_token",
			Message: "Bearer token is required but not provided",
		}
	}

	// Check if token is expired and refresh if possible
	if b.ExpiresAt != nil && time.Now().After(*b.ExpiresAt) {
		if err := b.Refresh(ctx); err != nil {
			return err
		}
	}

	req.Header.Set("Authorization", "Bearer "+b.Token)
	return nil
}

// IsValid checks if the bearer token is present and not expired.
func (b *BearerTokenAuth) IsValid(ctx context.Context) bool {
	if b.Token == "" {
		return false
	}
	if b.ExpiresAt != nil && time.Now().After(*b.ExpiresAt) {
		return false
	}
	return true
}

// Refresh attempts to refresh the bearer token using the refresh token.
func (b *BearerTokenAuth) Refresh(ctx context.Context) error {
	if b.RefreshToken == "" || b.RefreshFunc == nil {
		return &AuthError{
			Type:    "refresh_not_available",
			Message: "Token refresh is not available (missing refresh token or refresh function)",
		}
	}

	response, err := b.RefreshFunc(ctx, b.RefreshToken)
	if err != nil {
		return &AuthError{
			Type:    "refresh_failed",
			Message: "Failed to refresh token: " + err.Error(),
		}
	}

	b.Token = response.AccessToken
	if response.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)
		b.ExpiresAt = &expiresAt
	}
	if response.RefreshToken != "" {
		b.RefreshToken = response.RefreshToken
	}

	return nil
}

// Type returns the authentication type.
func (b *BearerTokenAuth) Type() AuthType {
	return AuthTypeBearer
}

// TokenResponse represents the response from a token refresh operation.
type TokenResponse struct {
	// AccessToken is the new access token
	AccessToken string `json:"access_token"`

	// TokenType is the type of token (usually "Bearer")
	TokenType string `json:"token_type"`

	// ExpiresIn is the number of seconds until the token expires
	ExpiresIn int64 `json:"expires_in"`

	// RefreshToken is the new refresh token (if provided)
	RefreshToken string `json:"refresh_token,omitempty"`

	// Scope contains the token scope information
	Scope string `json:"scope,omitempty"`
}

// AuthError represents an authentication-related error.
type AuthError struct {
	// Type is the error type identifier
	Type string `json:"type"`

	// Message is a human-readable error message
	Message string `json:"message"`

	// Code is an optional error code
	Code int `json:"code,omitempty"`

	// Details contains additional error information
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *AuthError) Error() string {
	return e.Message
}

// Is implements error comparison for Go 1.13+ error handling.
func (e *AuthError) Is(target error) bool {
	if ae, ok := target.(*AuthError); ok {
		return e.Type == ae.Type
	}
	return false
}

// SubscriptionAuth implements subscription-based authentication for Claude Code.
// This authentication method works with users who have a Claude subscription
// and have set up authentication using `claude setup-token`.
//
// The authenticator will automatically detect and use the CLI's authentication
// state without requiring manual token management.
//
// Example usage:
//
//	auth := &types.SubscriptionAuth{}
//	client := claude.NewClient(ctx, &types.Config{
//		Auth: auth,
//	})
type SubscriptionAuth struct {
	// TokenPath is the path to the CLI token file (auto-detected if empty)
	TokenPath string

	// ConfigPath is the path to the CLI config directory (auto-detected if empty)
	ConfigPath string

	// cached token information
	cachedToken      string
	cachedExpiration *time.Time
	lastChecked      time.Time
}

// Authenticate handles subscription-based authentication by delegating to the Claude CLI.
// This method doesn't set headers directly since the Claude Code SDK uses subprocess execution.
func (s *SubscriptionAuth) Authenticate(ctx context.Context, req *http.Request) error {
	// For ClaudeCodeClient, authentication is handled by the CLI subprocess,
	// so we just need to verify that subscription authentication is available
	if !s.IsValid(ctx) {
		return &AuthError{
			Type:    "subscription_unavailable",
			Message: "Claude subscription authentication is not available. Please run 'claude setup-token' to configure subscription authentication.",
		}
	}
	
	// No headers need to be set since the CLI handles authentication
	return nil
}

// IsValid checks if subscription authentication is properly configured.
// This verifies that the Claude CLI has been set up with a valid subscription token.
func (s *SubscriptionAuth) IsValid(ctx context.Context) bool {
	// Check if Claude CLI is available
	if !s.isCLIAvailable() {
		return false
	}

	// Check if authentication token exists and is valid
	return s.hasValidToken(ctx)
}

// Refresh attempts to refresh the subscription authentication.
// For subscription auth, this would typically involve re-running the CLI setup.
func (s *SubscriptionAuth) Refresh(ctx context.Context) error {
	// Clear cached token
	s.cachedToken = ""
	s.cachedExpiration = nil
	s.lastChecked = time.Time{}

	// Check if authentication is still valid after clearing cache
	if s.IsValid(ctx) {
		return nil
	}

	return &AuthError{
		Type:    "refresh_required",
		Message: "Subscription authentication needs to be refreshed. Please run 'claude setup-token' to re-authenticate.",
	}
}

// Type returns the authentication type.
func (s *SubscriptionAuth) Type() AuthType {
	return AuthTypeSubscription
}

// isCLIAvailable checks if the Claude CLI is available in the system.
func (s *SubscriptionAuth) isCLIAvailable() bool {
	// Try to find claude command
	candidates := []string{"claude", "npx claude"}
	
	for _, candidate := range candidates {
		if strings.Contains(candidate, " ") {
			// For commands like "npx claude", test by running with --version
			parts := strings.Fields(candidate)
			cmd := exec.Command(parts[0], append(parts[1:], "--version")...)
			if err := cmd.Run(); err == nil {
				return true
			}
		} else {
			// For single commands, check if available in PATH
			if _, err := exec.LookPath(candidate); err == nil {
				return true
			}
		}
	}
	
	return false
}

// hasValidToken checks if there's a valid subscription token available.
func (s *SubscriptionAuth) hasValidToken(ctx context.Context) bool {
	// Check cache first (with 5-minute TTL)
	if !s.lastChecked.IsZero() && time.Since(s.lastChecked) < 5*time.Minute {
		if s.cachedExpiration != nil && time.Now().Before(*s.cachedExpiration) {
			return s.cachedToken != ""
		}
	}

	// Try to run a simple Claude command to verify authentication
	cmd := exec.CommandContext(ctx, "claude", "--version")
	if err := cmd.Run(); err != nil {
		return false
	}

	// Update cache
	s.lastChecked = time.Now()
	// Set a conservative expiration (subscription tokens are typically long-lived)
	expiration := time.Now().Add(24 * time.Hour)
	s.cachedExpiration = &expiration
	s.cachedToken = "valid" // We don't need the actual token value
	
	return true
}

// CredentialManager defines the interface for managing authentication credentials.
// This interface allows for secure storage and retrieval of authentication information.
type CredentialManager interface {
	// Store securely stores authentication credentials.
	Store(ctx context.Context, key string, credentials interface{}) error

	// Retrieve retrieves stored authentication credentials.
	Retrieve(ctx context.Context, key string) (interface{}, error)

	// Delete removes stored credentials.
	Delete(ctx context.Context, key string) error

	// List returns all available credential keys.
	List(ctx context.Context) ([]string, error)

	// Exists checks if credentials exist for the given key.
	Exists(ctx context.Context, key string) bool
}

// MemoryCredentialManager is a simple in-memory implementation of CredentialManager.
// This should only be used for testing or development environments.
//
// Example usage:
//
//	manager := &types.MemoryCredentialManager{}
//	err := manager.Store(ctx, "default", &types.APIKeyAuth{
//		APIKey: "your-api-key",
//	})
type MemoryCredentialManager struct {
	credentials map[string]interface{}
}

// NewMemoryCredentialManager creates a new in-memory credential manager.
func NewMemoryCredentialManager() *MemoryCredentialManager {
	return &MemoryCredentialManager{
		credentials: make(map[string]interface{}),
	}
}

// Store stores credentials in memory.
func (m *MemoryCredentialManager) Store(ctx context.Context, key string, credentials interface{}) error {
	if m.credentials == nil {
		m.credentials = make(map[string]interface{})
	}
	m.credentials[key] = credentials
	return nil
}

// Retrieve retrieves credentials from memory.
func (m *MemoryCredentialManager) Retrieve(ctx context.Context, key string) (interface{}, error) {
	if m.credentials == nil {
		return nil, &AuthError{
			Type:    "credentials_not_found",
			Message: "No credentials found for key: " + key,
		}
	}

	credentials, exists := m.credentials[key]
	if !exists {
		return nil, &AuthError{
			Type:    "credentials_not_found",
			Message: "No credentials found for key: " + key,
		}
	}

	return credentials, nil
}

// Delete removes credentials from memory.
func (m *MemoryCredentialManager) Delete(ctx context.Context, key string) error {
	if m.credentials != nil {
		delete(m.credentials, key)
	}
	return nil
}

// List returns all available credential keys.
func (m *MemoryCredentialManager) List(ctx context.Context) ([]string, error) {
	keys := make([]string, 0, len(m.credentials))
	for key := range m.credentials {
		keys = append(keys, key)
	}
	return keys, nil
}

// Exists checks if credentials exist for the given key.
func (m *MemoryCredentialManager) Exists(ctx context.Context, key string) bool {
	if m.credentials == nil {
		return false
	}
	_, exists := m.credentials[key]
	return exists
}
