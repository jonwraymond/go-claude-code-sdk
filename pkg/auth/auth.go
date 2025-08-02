package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jraymond/claude-code-go-sdk/pkg/types"
)

var (
	// ErrInvalidCredentials indicates that the provided credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrAuthenticationFailed indicates that authentication failed
	ErrAuthenticationFailed = errors.New("authentication failed")

	// ErrSessionExpired indicates that the session has expired
	ErrSessionExpired = errors.New("session expired")

	// ErrMissingCredentials indicates that required credentials are missing
	ErrMissingCredentials = errors.New("missing required credentials")
)

// Authenticator defines the interface for authentication mechanisms
type Authenticator interface {
	// Authenticate performs the authentication and returns auth info
	Authenticate(ctx context.Context) (*types.AuthInfo, error)

	// Refresh refreshes the authentication if supported
	Refresh(ctx context.Context) error

	// IsValid checks if the current authentication is valid
	IsValid() bool

	// GetHeaders returns the authentication headers
	GetHeaders() map[string]string
}

// BaseAuthenticator provides common functionality for authenticators
type BaseAuthenticator struct {
	mu          sync.RWMutex
	authInfo    *types.AuthInfo
	lastRefresh time.Time
}

// IsValid checks if the authentication is currently valid
func (b *BaseAuthenticator) IsValid() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.authInfo == nil {
		return false
	}

	// Check if session-based auth has expired
	if b.authInfo.ExpiresAt != nil && time.Now().After(*b.authInfo.ExpiresAt) {
		return false
	}

	return true
}

// GetHeaders returns authentication headers
func (b *BaseAuthenticator) GetHeaders() map[string]string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.authInfo == nil {
		return nil
	}

	return b.authInfo.Headers
}

// setAuthInfo safely sets the authentication info
func (b *BaseAuthenticator) setAuthInfo(info *types.AuthInfo) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.authInfo = info
	b.lastRefresh = time.Now()
}

// getAuthInfo safely gets the authentication info
func (b *BaseAuthenticator) getAuthInfo() *types.AuthInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.authInfo == nil {
		return nil
	}

	// Return a copy to prevent external modifications
	infoCopy := *b.authInfo
	if b.authInfo.Headers != nil {
		infoCopy.Headers = make(map[string]string, len(b.authInfo.Headers))
		for k, v := range b.authInfo.Headers {
			infoCopy.Headers[k] = v
		}
	}

	return &infoCopy
}

// APIKeyAuthenticator implements API key-based authentication
type APIKeyAuthenticator struct {
	BaseAuthenticator
	apiKey string
}

// NewAPIKeyAuthenticator creates a new API key authenticator
func NewAPIKeyAuthenticator(apiKey string) (*APIKeyAuthenticator, error) {
	if apiKey == "" {
		return nil, ErrMissingCredentials
	}

	auth := &APIKeyAuthenticator{
		apiKey: apiKey,
	}

	// Validate the API key format
	validator := NewValidator()
	if err := validator.ValidateAPIKey(apiKey); err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	// Pre-populate auth info since API keys don't require authentication calls
	auth.setAuthInfo(&types.AuthInfo{
		Type: types.AuthTypeAPIKey,
		Headers: map[string]string{
			"X-API-Key": apiKey,
		},
		CreatedAt: time.Now(),
	})

	return auth, nil
}

// Authenticate performs API key authentication
func (a *APIKeyAuthenticator) Authenticate(ctx context.Context) (*types.AuthInfo, error) {
	// API key authentication doesn't require a separate auth call
	// The key is validated during creation
	return a.getAuthInfo(), nil
}

// Refresh is a no-op for API key authentication
func (a *APIKeyAuthenticator) Refresh(ctx context.Context) error {
	// API keys don't need refreshing
	return nil
}

// SessionAuthenticator implements session-based authentication
type SessionAuthenticator struct {
	BaseAuthenticator
	sessionKey    string
	refreshBefore time.Duration // How long before expiry to refresh
}

// NewSessionAuthenticator creates a new session authenticator
func NewSessionAuthenticator(sessionKey string) (*SessionAuthenticator, error) {
	if sessionKey == "" {
		return nil, ErrMissingCredentials
	}

	auth := &SessionAuthenticator{
		sessionKey:    sessionKey,
		refreshBefore: 5 * time.Minute, // Default: refresh 5 minutes before expiry
	}

	// Validate session key format
	validator := NewValidator()
	if err := validator.ValidateSessionKey(sessionKey); err != nil {
		return nil, fmt.Errorf("invalid session key: %w", err)
	}

	return auth, nil
}

// Authenticate performs session authentication
func (s *SessionAuthenticator) Authenticate(ctx context.Context) (*types.AuthInfo, error) {
	// In a real implementation, this would make an API call to validate the session
	// For now, we'll simulate it
	expiresAt := time.Now().Add(time.Hour)

	authInfo := &types.AuthInfo{
		Type: types.AuthTypeSession,
		Headers: map[string]string{
			"Cookie": fmt.Sprintf("sessionKey=%s", s.sessionKey),
		},
		CreatedAt: time.Now(),
		ExpiresAt: &expiresAt,
		Metadata: map[string]interface{}{
			"sessionId": generateSessionID(),
		},
	}

	s.setAuthInfo(authInfo)
	return s.getAuthInfo(), nil
}

// Refresh refreshes the session authentication
func (s *SessionAuthenticator) Refresh(ctx context.Context) error {
	// Check if refresh is needed
	authInfo := s.getAuthInfo()
	if authInfo == nil {
		_, err := s.Authenticate(ctx)
		return err
	}

	if authInfo.ExpiresAt != nil {
		timeUntilExpiry := time.Until(*authInfo.ExpiresAt)
		if timeUntilExpiry > s.refreshBefore {
			// No need to refresh yet
			return nil
		}
	}

	// Perform refresh (in real implementation, this would call an API)
	_, err := s.Authenticate(ctx)
	return err
}

// IsValid checks if the session is still valid
func (s *SessionAuthenticator) IsValid() bool {
	if !s.BaseAuthenticator.IsValid() {
		return false
	}

	// Check if we should proactively refresh
	authInfo := s.getAuthInfo()
	if authInfo != nil && authInfo.ExpiresAt != nil {
		timeUntilExpiry := time.Until(*authInfo.ExpiresAt)
		if timeUntilExpiry <= s.refreshBefore {
			// Session is about to expire, consider it invalid to trigger refresh
			return false
		}
	}

	return true
}

// generateSessionID generates a mock session ID for demonstration
func generateSessionID() string {
	// In production, use a proper UUID or random string generator
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// AuthOption configures an authenticator
type AuthOption func(Authenticator) error

// WithRefreshBefore sets how long before expiry to refresh session tokens
func WithRefreshBefore(duration time.Duration) AuthOption {
	return func(auth Authenticator) error {
		if sessionAuth, ok := auth.(*SessionAuthenticator); ok {
			sessionAuth.refreshBefore = duration
			return nil
		}
		return errors.New("option only applicable to session authenticator")
	}
}
