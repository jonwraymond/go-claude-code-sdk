package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

var (
	// ErrCredentialNotFound indicates the requested credential was not found
	ErrCredentialNotFound = errors.New("credential not found")

	// ErrCredentialExists indicates a credential with the same ID already exists
	ErrCredentialExists = errors.New("credential already exists")
)

// CredentialStore defines the interface for credential storage
type CredentialStore interface {
	// Store saves a credential
	Store(ctx context.Context, id string, cred *StoredCredential) error

	// Retrieve gets a credential by ID
	Retrieve(ctx context.Context, id string) (*StoredCredential, error)

	// Delete removes a credential
	Delete(ctx context.Context, id string) error

	// List returns all credential IDs
	List(ctx context.Context) ([]string, error)
}

// StoredCredential represents a credential in storage
type StoredCredential struct {
	ID         string         `json:"id"`
	Type       types.AuthType `json:"type"`
	Credential string         `json:"credential"` // Encrypted in production
	Metadata   map[string]any `json:"metadata,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	LastUsed   *time.Time     `json:"last_used,omitempty"`
	ExpiresAt  *time.Time     `json:"expires_at,omitempty"`
}

// Manager handles credential management and lifecycle
type Manager struct {
	store          CredentialStore
	authenticators map[string]Authenticator
	mu             sync.RWMutex
	validator      *Validator
}

// NewManager creates a new credential manager
func NewManager(store CredentialStore) *Manager {
	return &Manager{
		store:          store,
		authenticators: make(map[string]Authenticator),
		validator:      NewValidator(),
	}
}

// StoreAPIKey stores an API key credential
func (m *Manager) StoreAPIKey(ctx context.Context, id, apiKey string) error {
	// Validate the API key
	if err := m.validator.ValidateAPIKey(apiKey); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	cred := &StoredCredential{
		ID:         id,
		Type:       types.AuthTypeAPIKey,
		Credential: apiKey, // In production, encrypt this
		CreatedAt:  time.Now(),
		Metadata: map[string]any{
			"prefix": apiKey[:10] + "...", // Store prefix for identification
		},
	}

	return m.store.Store(ctx, id, cred)
}

// StoreSessionKey stores a session key credential
func (m *Manager) StoreSessionKey(ctx context.Context, id, sessionKey string, expiresAt *time.Time) error {
	// Validate the session key
	if err := m.validator.ValidateSessionKey(sessionKey); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	cred := &StoredCredential{
		ID:         id,
		Type:       types.AuthTypeSession,
		Credential: sessionKey, // In production, encrypt this
		CreatedAt:  time.Now(),
		ExpiresAt:  expiresAt,
	}

	return m.store.Store(ctx, id, cred)
}

// GetAuthenticator retrieves or creates an authenticator for a credential
func (m *Manager) GetAuthenticator(ctx context.Context, id string) (Authenticator, error) {
	m.mu.RLock()
	auth, exists := m.authenticators[id]
	m.mu.RUnlock()

	if exists && auth.IsValid() {
		return auth, nil
	}

	// Retrieve credential from store
	cred, err := m.store.Retrieve(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}

	// Update last used time
	now := time.Now()
	cred.LastUsed = &now
	_ = m.store.Store(ctx, id, cred) // Ignore error for last used update

	// Create appropriate authenticator
	var newAuth Authenticator
	switch cred.Type {
	case types.AuthTypeAPIKey:
		newAuth, err = NewAPIKeyAuthenticator(cred.Credential)
	case types.AuthTypeSession:
		newAuth, err = NewSessionAuthenticator(cred.Credential)
	default:
		return nil, fmt.Errorf("unsupported credential type: %s", cred.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator: %w", err)
	}

	// Cache the authenticator
	m.mu.Lock()
	m.authenticators[id] = newAuth
	m.mu.Unlock()

	return newAuth, nil
}

// DeleteCredential removes a credential and its associated authenticator
func (m *Manager) DeleteCredential(ctx context.Context, id string) error {
	// Remove from cache
	m.mu.Lock()
	delete(m.authenticators, id)
	m.mu.Unlock()

	// Remove from store
	return m.store.Delete(ctx, id)
}

// ListCredentials returns all stored credential IDs
func (m *Manager) ListCredentials(ctx context.Context) ([]string, error) {
	return m.store.List(ctx)
}

// RefreshCredential refreshes a credential if supported
func (m *Manager) RefreshCredential(ctx context.Context, id string) error {
	auth, err := m.GetAuthenticator(ctx, id)
	if err != nil {
		return err
	}

	return auth.Refresh(ctx)
}

// MemoryStore implements an in-memory credential store
type MemoryStore struct {
	mu          sync.RWMutex
	credentials map[string]*StoredCredential
}

// NewMemoryStore creates a new in-memory credential store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		credentials: make(map[string]*StoredCredential),
	}
}

// Store saves a credential in memory
func (s *MemoryStore) Store(ctx context.Context, id string, cred *StoredCredential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.credentials[id] = cred
	return nil
}

// Retrieve gets a credential from memory
func (s *MemoryStore) Retrieve(ctx context.Context, id string) (*StoredCredential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cred, exists := s.credentials[id]
	if !exists {
		return nil, ErrCredentialNotFound
	}

	// Return a copy to prevent external modifications
	credCopy := *cred
	return &credCopy, nil
}

// Delete removes a credential from memory
func (s *MemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.credentials, id)
	return nil
}

// List returns all credential IDs
func (s *MemoryStore) List(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.credentials))
	for id := range s.credentials {
		ids = append(ids, id)
	}

	return ids, nil
}

// FileStore implements file-based credential storage
type FileStore struct {
	basePath string
	mu       sync.RWMutex
}

// NewFileStore creates a new file-based credential store
func NewFileStore(basePath string) (*FileStore, error) {
	// Ensure the directory exists
	if err := os.MkdirAll(basePath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create credential directory: %w", err)
	}

	return &FileStore{
		basePath: basePath,
	}, nil
}

// Store saves a credential to a file
func (s *FileStore) Store(ctx context.Context, id string, cred *StoredCredential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credential: %w", err)
	}

	filePath := filepath.Join(s.basePath, id+".json")
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write credential file: %w", err)
	}

	return nil
}

// Retrieve gets a credential from a file
func (s *FileStore) Retrieve(ctx context.Context, id string) (*StoredCredential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filePath := filepath.Join(s.basePath, id+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrCredentialNotFound
		}
		return nil, fmt.Errorf("failed to read credential file: %w", err)
	}

	var cred StoredCredential
	if err := json.Unmarshal(data, &cred); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credential: %w", err)
	}

	return &cred, nil
}

// Delete removes a credential file
func (s *FileStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := filepath.Join(s.basePath, id+".json")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete credential file: %w", err)
	}

	return nil
}

// List returns all credential IDs from files
func (s *FileStore) List(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credential directory: %w", err)
	}

	var ids []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			id := strings.TrimSuffix(entry.Name(), ".json")
			ids = append(ids, id)
		}
	}

	return ids, nil
}
