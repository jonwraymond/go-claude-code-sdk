package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func TestManager_StoreAPIKey(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store)
	ctx := context.Background()

	tests := []struct {
		name    string
		id      string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid API key",
			id:      "test-api-key",
			apiKey:  "sk-ant-api03j9h8f7d6s5a4l3k2m1n0",
			wantErr: false,
		},
		{
			name:    "invalid API key",
			id:      "invalid-key",
			apiKey:  "invalid",
			wantErr: true,
		},
		{
			name:    "empty API key",
			id:      "empty-key",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.StoreAPIKey(ctx, tt.id, tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("StoreAPIKey() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify it was stored
				cred, err := store.Retrieve(ctx, tt.id)
				if err != nil {
					t.Errorf("Failed to retrieve stored credential: %v", err)
				}
				if cred == nil {
					t.Error("Retrieved credential is nil")
				} else if cred.Type != types.AuthTypeAPIKey {
					t.Errorf("Stored credential type = %v, want %v", cred.Type, types.AuthTypeAPIKey)
				}
			}
		})
	}
}

func TestManager_StoreSessionKey(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store)
	ctx := context.Background()

	expiresAt := time.Now().Add(time.Hour)

	tests := []struct {
		name       string
		id         string
		sessionKey string
		expiresAt  *time.Time
		wantErr    bool
	}{
		{
			name:       "valid session key",
			id:         "test-session",
			sessionKey: "sess_abcdefghijklmnopqrstuvwxyz123456",
			expiresAt:  &expiresAt,
			wantErr:    false,
		},
		{
			name:       "invalid session key",
			id:         "invalid-session",
			sessionKey: "short",
			wantErr:    true,
		},
		{
			name:       "no expiration",
			id:         "no-expire",
			sessionKey: "sess_abcdefghijklmnopqrstuvwxyz789012",
			expiresAt:  nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.StoreSessionKey(ctx, tt.id, tt.sessionKey, tt.expiresAt)
			if (err != nil) != tt.wantErr {
				t.Errorf("StoreSessionKey() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify it was stored
				cred, err := store.Retrieve(ctx, tt.id)
				if err != nil {
					t.Errorf("Failed to retrieve stored credential: %v", err)
				}
				if cred == nil {
					t.Error("Retrieved credential is nil")
				} else if cred.Type != types.AuthTypeSession {
					t.Errorf("Stored credential type = %v, want %v", cred.Type, types.AuthTypeSession)
				}
			}
		})
	}
}

func TestManager_GetAuthenticator(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store)
	ctx := context.Background()

	// Store test credentials
	apiKeyID := "test-api-key"
	sessionKeyID := "test-session"

	err := manager.StoreAPIKey(ctx, apiKeyID, "sk-ant-api03j9h8f7d6s5a4l3k2m1n0")
	if err != nil {
		t.Fatalf("Failed to store API key: %v", err)
	}

	err = manager.StoreSessionKey(ctx, sessionKeyID, "sess_abcdefghijklmnopqrstuvwxyz123456", nil)
	if err != nil {
		t.Fatalf("Failed to store session key: %v", err)
	}

	// Test retrieving API key authenticator
	auth, err := manager.GetAuthenticator(ctx, apiKeyID)
	if err != nil {
		t.Errorf("GetAuthenticator() for API key error = %v", err)
	}
	if auth == nil {
		t.Error("GetAuthenticator() returned nil for API key")
	} else if _, ok := auth.(*APIKeyAuthenticator); !ok {
		t.Error("GetAuthenticator() returned wrong type for API key")
	}

	// Test retrieving session authenticator
	auth, err = manager.GetAuthenticator(ctx, sessionKeyID)
	if err != nil {
		t.Errorf("GetAuthenticator() for session error = %v", err)
	}
	if auth == nil {
		t.Error("GetAuthenticator() returned nil for session")
	} else if _, ok := auth.(*SessionAuthenticator); !ok {
		t.Error("GetAuthenticator() returned wrong type for session")
	}

	// Test non-existent credential
	_, err = manager.GetAuthenticator(ctx, "non-existent")
	if err == nil {
		t.Error("GetAuthenticator() should fail for non-existent credential")
	}
}

func TestManager_DeleteCredential(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store)
	ctx := context.Background()

	// Store and then delete
	id := "test-delete"
	err := manager.StoreAPIKey(ctx, id, "sk-ant-api03j9h8f7d6s5a4l3k2m1n0")
	if err != nil {
		t.Fatalf("Failed to store credential: %v", err)
	}

	// Get authenticator to cache it
	_, err = manager.GetAuthenticator(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get authenticator: %v", err)
	}

	// Delete
	err = manager.DeleteCredential(ctx, id)
	if err != nil {
		t.Errorf("DeleteCredential() error = %v", err)
	}

	// Verify it's gone from store
	_, err = store.Retrieve(ctx, id)
	if err != ErrCredentialNotFound {
		t.Error("Credential still exists after deletion")
	}

	// Verify it's gone from cache
	_, err = manager.GetAuthenticator(ctx, id)
	if err == nil {
		t.Error("Authenticator still cached after deletion")
	}
}

func TestManager_ListCredentials(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store)
	ctx := context.Background()

	// Store multiple credentials
	creds := map[string]string{
		"api-1": "sk-ant-api03j9h8f7d6s5a4l3k2m1n1",
		"api-2": "sk-ant-api03j9h8f7d6s5a4l3k2m1n2",
		"api-3": "sk-ant-api03j9h8f7d6s5a4l3k2m1n3",
	}

	for id, key := range creds {
		err := manager.StoreAPIKey(ctx, id, key)
		if err != nil {
			t.Fatalf("Failed to store credential %s: %v", id, err)
		}
	}

	// List credentials
	ids, err := manager.ListCredentials(ctx)
	if err != nil {
		t.Errorf("ListCredentials() error = %v", err)
	}

	if len(ids) != len(creds) {
		t.Errorf("ListCredentials() returned %d items, want %d", len(ids), len(creds))
	}

	// Verify all IDs are present
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	for id := range creds {
		if !idMap[id] {
			t.Errorf("Credential ID %s not found in list", id)
		}
	}
}

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	cred := &StoredCredential{
		ID:         "test-cred",
		Type:       types.AuthTypeAPIKey,
		Credential: "test-key",
		CreatedAt:  time.Now(),
	}

	// Test Store
	err := store.Store(ctx, cred.ID, cred)
	if err != nil {
		t.Errorf("Store() error = %v", err)
	}

	// Test Retrieve
	retrieved, err := store.Retrieve(ctx, cred.ID)
	if err != nil {
		t.Errorf("Retrieve() error = %v", err)
	}
	if retrieved.ID != cred.ID {
		t.Errorf("Retrieved credential ID = %v, want %v", retrieved.ID, cred.ID)
	}

	// Test List
	ids, err := store.List(ctx)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(ids) != 1 || ids[0] != cred.ID {
		t.Errorf("List() = %v, want [%s]", ids, cred.ID)
	}

	// Test Delete
	err = store.Delete(ctx, cred.ID)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err = store.Retrieve(ctx, cred.ID)
	if err != ErrCredentialNotFound {
		t.Error("Credential still exists after deletion")
	}
}

func TestFileStore(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "auth_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := NewFileStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create FileStore: %v", err)
	}

	ctx := context.Background()

	cred := &StoredCredential{
		ID:         "test-file-cred",
		Type:       types.AuthTypeSession,
		Credential: "test-session-key",
		CreatedAt:  time.Now(),
		Metadata: map[string]any{
			"test": "value",
		},
	}

	// Test Store
	err = store.Store(ctx, cred.ID, cred)
	if err != nil {
		t.Errorf("Store() error = %v", err)
	}

	// Verify file was created
	filePath := filepath.Join(tempDir, cred.ID+".json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Credential file was not created")
	}

	// Test Retrieve
	retrieved, err := store.Retrieve(ctx, cred.ID)
	if err != nil {
		t.Errorf("Retrieve() error = %v", err)
	}
	if retrieved.ID != cred.ID {
		t.Errorf("Retrieved credential ID = %v, want %v", retrieved.ID, cred.ID)
	}

	// Test List
	ids, err := store.List(ctx)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(ids) != 1 || ids[0] != cred.ID {
		t.Errorf("List() = %v, want [%s]", ids, cred.ID)
	}

	// Test Delete
	err = store.Delete(ctx, cred.ID)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("Credential file still exists after deletion")
	}
}

func TestManager_RefreshCredential(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store)
	ctx := context.Background()

	// Store session credential
	sessionID := "refresh-test"
	err := manager.StoreSessionKey(ctx, sessionID, "sess_abcdefghijklmnopqrstuvwxyz123456", nil)
	if err != nil {
		t.Fatalf("Failed to store session key: %v", err)
	}

	// Get authenticator
	auth, err := manager.GetAuthenticator(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get authenticator: %v", err)
	}

	// Authenticate first
	_, err = auth.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	// Test refresh
	err = manager.RefreshCredential(ctx, sessionID)
	if err != nil {
		t.Errorf("RefreshCredential() error = %v", err)
	}

	// Test refresh for API key (should not error)
	apiKeyID := "api-refresh-test"
	err = manager.StoreAPIKey(ctx, apiKeyID, "sk-ant-api03j9h8f7d6s5a4l3k2m1n0")
	if err != nil {
		t.Fatalf("Failed to store API key: %v", err)
	}

	err = manager.RefreshCredential(ctx, apiKeyID)
	if err != nil {
		t.Errorf("RefreshCredential() for API key error = %v", err)
	}
}
