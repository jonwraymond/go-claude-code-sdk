package auth

import (
	"context"
	"testing"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func TestNewAPIKeyAuthenticator(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid api key",
			apiKey:  "sk-ant-api03-valid-test-key-123456789abcdefghijklmnopqrstuvwxyz",
			wantErr: false,
		},
		{
			name:    "empty api key",
			apiKey:  "",
			wantErr: true,
			errMsg:  "missing required credentials",
		},
		{
			name:    "invalid format",
			apiKey:  "invalid-key",
			wantErr: true,
			errMsg:  "invalid API key",
		},
		{
			name:    "test key rejected",
			apiKey:  "sk-ant-test-key-123456789",
			wantErr: true,
			errMsg:  "test or example API keys are not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewAPIKeyAuthenticator(tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAPIKeyAuthenticator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg && !contains(err.Error(), tt.errMsg) {
					t.Errorf("NewAPIKeyAuthenticator() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
			if !tt.wantErr && auth == nil {
				t.Error("NewAPIKeyAuthenticator() returned nil authenticator")
			}
		})
	}
}

func TestAPIKeyAuthenticator_Authenticate(t *testing.T) {
	auth, err := NewAPIKeyAuthenticator("sk-ant-api03-valid-test-key-123456789abcdefghijklmnopqrstuvwxyz")
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	ctx := context.Background()
	authInfo, err := auth.Authenticate(ctx)
	if err != nil {
		t.Errorf("Authenticate() error = %v", err)
	}

	if authInfo == nil {
		t.Fatal("Authenticate() returned nil AuthInfo")
	}

	if authInfo.Type != types.AuthTypeAPIKey {
		t.Errorf("AuthInfo.Type = %v, want %v", authInfo.Type, types.AuthTypeAPIKey)
	}

	if authInfo.Headers == nil {
		t.Fatal("AuthInfo.Headers is nil")
	}

	if apiKey, ok := authInfo.Headers["X-API-Key"]; !ok || apiKey != "sk-ant-api03-valid-test-key-123456789abcdefghijklmnopqrstuvwxyz" {
		t.Errorf("AuthInfo.Headers[X-API-Key] = %v, want %v", apiKey, "sk-ant-api03-valid-test-key-123456789abcdefghijklmnopqrstuvwxyz")
	}
}

func TestAPIKeyAuthenticator_IsValid(t *testing.T) {
	auth, err := NewAPIKeyAuthenticator("sk-ant-api03-valid-test-key-123456789abcdefghijklmnopqrstuvwxyz")
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	if !auth.IsValid() {
		t.Error("IsValid() = false, want true")
	}
}

func TestNewSessionAuthenticator(t *testing.T) {
	tests := []struct {
		name       string
		sessionKey string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid session key",
			sessionKey: "sess_abcdefghijklmnopqrstuvwxyz123456",
			wantErr:    false,
		},
		{
			name:       "empty session key",
			sessionKey: "",
			wantErr:    true,
			errMsg:     "missing required credentials",
		},
		{
			name:       "too short",
			sessionKey: "short",
			wantErr:    true,
			errMsg:     "session key too short",
		},
		{
			name:       "invalid characters",
			sessionKey: "sess_abc!@#$%^&*()_+{}[]|\\:\";<>?,./~`123456789012",
			wantErr:    true,
			errMsg:     "session key contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewSessionAuthenticator(tt.sessionKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSessionAuthenticator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("NewSessionAuthenticator() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
			if !tt.wantErr && auth == nil {
				t.Error("NewSessionAuthenticator() returned nil authenticator")
			}
		})
	}
}

func TestSessionAuthenticator_Authenticate(t *testing.T) {
	auth, err := NewSessionAuthenticator("sess_abcdefghijklmnopqrstuvwxyz123456")
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	ctx := context.Background()
	authInfo, err := auth.Authenticate(ctx)
	if err != nil {
		t.Errorf("Authenticate() error = %v", err)
	}

	if authInfo == nil {
		t.Fatal("Authenticate() returned nil AuthInfo")
	}

	if authInfo.Type != types.AuthTypeSession {
		t.Errorf("AuthInfo.Type = %v, want %v", authInfo.Type, types.AuthTypeSession)
	}

	if authInfo.Headers == nil {
		t.Fatal("AuthInfo.Headers is nil")
	}

	if cookie, ok := authInfo.Headers["Cookie"]; !ok || !contains(cookie, "sessionKey=") {
		t.Errorf("AuthInfo.Headers[Cookie] = %v, want to contain sessionKey=", cookie)
	}

	if authInfo.ExpiresAt == nil {
		t.Error("AuthInfo.ExpiresAt is nil, expected expiration time")
	}
}

func TestSessionAuthenticator_IsValid(t *testing.T) {
	auth, err := NewSessionAuthenticator("sess_abcdefghijklmnopqrstuvwxyz123456")
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	// Before authentication, should be invalid
	if auth.IsValid() {
		t.Error("IsValid() = true before authentication, want false")
	}

	// After authentication, should be valid
	ctx := context.Background()
	_, err = auth.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	if !auth.IsValid() {
		t.Error("IsValid() = false after authentication, want true")
	}

	// Test with expired session
	expiredTime := time.Now().Add(-1 * time.Hour)
	auth.setAuthInfo(&types.AuthInfo{
		Type:      types.AuthTypeSession,
		Headers:   map[string]string{"Cookie": "sessionKey=test"},
		CreatedAt: time.Now(),
		ExpiresAt: &expiredTime,
	})

	if auth.IsValid() {
		t.Error("IsValid() = true for expired session, want false")
	}
}

func TestSessionAuthenticator_Refresh(t *testing.T) {
	auth, err := NewSessionAuthenticator("sess_abcdefghijklmnopqrstuvwxyz123456")
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	ctx := context.Background()

	// First authenticate
	_, err = auth.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	// Refresh should succeed
	err = auth.Refresh(ctx)
	if err != nil {
		t.Errorf("Refresh() error = %v", err)
	}

	// Get auth info after refresh
	authInfo2 := auth.getAuthInfo()
	if authInfo2 == nil {
		t.Fatal("getAuthInfo() returned nil after refresh")
	}

	// Verify refresh was successful by checking the auth info is not nil
	// In a real implementation, we would verify the session was actually refreshed
}

func TestAuthOption_WithRefreshBefore(t *testing.T) {
	auth, err := NewSessionAuthenticator("sess_abcdefghijklmnopqrstuvwxyz123456")
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	// Apply option
	opt := WithRefreshBefore(10 * time.Minute)
	err = opt(auth)
	if err != nil {
		t.Errorf("WithRefreshBefore() error = %v", err)
	}

	if auth.refreshBefore != 10*time.Minute {
		t.Errorf("refreshBefore = %v, want %v", auth.refreshBefore, 10*time.Minute)
	}

	// Test with API key authenticator (should fail)
	apiAuth, _ := NewAPIKeyAuthenticator("sk-ant-api03-valid-test-key-123456789abcdefghijklmnopqrstuvwxyz")
	err = opt(apiAuth)
	if err == nil {
		t.Error("WithRefreshBefore() on API key authenticator should return error")
	}
}

func TestBaseAuthenticator_ThreadSafety(t *testing.T) {
	auth, err := NewAPIKeyAuthenticator("sk-ant-api03-valid-test-key-123456789abcdefghijklmnopqrstuvwxyz")
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	// Test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = auth.IsValid()
			_ = auth.GetHeaders()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && len(substr) > 0 && s[:len(substr)] == substr || len(s) > len(substr) && s[len(s)-len(substr):] == substr || len(substr) > 0 && len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
