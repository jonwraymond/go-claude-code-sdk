package client

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	sdkerrors "github.com/jonwraymond/go-claude-code-sdk/pkg/errors"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// ClaudeCodeSessionManager manages conversation sessions for Claude Code CLI.
// Unlike the API-based SessionManager, this works with Claude Code's
// session persistence through the --session flag and manages conversation
// history via the CLI's built-in session support.
//
// The session manager provides:
// - Session creation and management
// - Project-aware session contexts
// - Session persistence across Claude Code invocations
// - Session metadata and configuration
// - Automatic session cleanup
//
// Example usage:
//
//	sessionManager := client.Sessions()
//	session, err := sessionManager.CreateSession(ctx, "user-123")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer session.Close()
//
//	response, err := session.Query(ctx, &types.QueryRequest{
//		Messages: []types.Message{
//			{Role: types.RoleUser, Content: "Analyze this codebase"},
//		},
//		MaxTokens: 1000,
//	})
type ClaudeCodeSessionManager struct {
	client   *ClaudeCodeClient
	sessions map[string]*ClaudeCodeSession
	mu       sync.RWMutex

	// Configuration
	config *ClaudeCodeSessionConfig

	// Background cleanup
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
	wg            sync.WaitGroup
}

// ClaudeCodeSessionConfig provides configuration for session management.
type ClaudeCodeSessionConfig struct {
	// MaxSessions limits the number of concurrent sessions (default: 50)
	MaxSessions int

	// SessionTimeout after which inactive sessions are cleaned up (default: 1h)
	SessionTimeout time.Duration

	// CleanupInterval for running session cleanup (default: 10m)
	CleanupInterval time.Duration

	// PersistSessions enables session persistence to disk (default: true)
	PersistSessions bool

	// SessionDirectory where sessions are persisted (default: .claude/sessions)
	SessionDirectory string
}

// DefaultClaudeCodeSessionConfig returns default session configuration.
func DefaultClaudeCodeSessionConfig() *ClaudeCodeSessionConfig {
	return &ClaudeCodeSessionConfig{
		MaxSessions:      50,
		SessionTimeout:   1 * time.Hour,
		CleanupInterval:  10 * time.Minute,
		PersistSessions:  true,
		SessionDirectory: ".claude/sessions",
	}
}

// ClaudeCodeSession represents a conversation session with Claude Code.
type ClaudeCodeSession struct {
	ID      string
	client  *ClaudeCodeClient
	manager *ClaudeCodeSessionManager

	// Session configuration
	projectDir string
	model      string

	// Session metadata
	metadata map[string]interface{}

	// Session lifecycle
	createdAt  time.Time
	lastUsedAt time.Time
	timeout    time.Duration
	closed     bool

	// Thread safety
	mu sync.RWMutex
}

// NewClaudeCodeSessionManager creates a new session manager.
func NewClaudeCodeSessionManager(client *ClaudeCodeClient) *ClaudeCodeSessionManager {
	config := DefaultClaudeCodeSessionConfig()
	return NewClaudeCodeSessionManagerWithConfig(client, config)
}

// NewClaudeCodeSessionManagerWithConfig creates a new session manager with custom configuration.
func NewClaudeCodeSessionManagerWithConfig(client *ClaudeCodeClient, config *ClaudeCodeSessionConfig) *ClaudeCodeSessionManager {
	sm := &ClaudeCodeSessionManager{
		client:      client,
		sessions:    make(map[string]*ClaudeCodeSession),
		config:      config,
		stopCleanup: make(chan struct{}),
	}

	// Start background cleanup
	sm.cleanupTicker = time.NewTicker(config.CleanupInterval)
	sm.wg.Add(1)
	go sm.runCleanup()

	return sm
}

// CreateSession creates a new Claude Code conversation session.
func (sm *ClaudeCodeSessionManager) CreateSession(ctx context.Context, sessionID string) (*ClaudeCodeSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if session already exists
	if existingSession, exists := sm.sessions[sessionID]; exists {
		if !existingSession.IsExpired() {
			return existingSession, nil
		}
		// Clean up expired session
		delete(sm.sessions, sessionID)
	}

	// Check session limit
	if len(sm.sessions) >= sm.config.MaxSessions {
		return nil, sdkerrors.NewValidationError("sessions", fmt.Sprintf("%d", len(sm.sessions)),
			fmt.Sprintf("max %d", sm.config.MaxSessions), "maximum number of sessions reached")
	}

	// Create new session
	session := &ClaudeCodeSession{
		ID:         sessionID,
		client:     sm.client,
		manager:    sm,
		projectDir: sm.client.workingDir,
		model:      sm.client.config.Model,
		metadata:   make(map[string]interface{}),
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
		timeout:    sm.config.SessionTimeout,
	}

	// Initialize session metadata
	session.metadata["project_dir"] = session.projectDir
	session.metadata["model"] = session.model

	// Get project context for the session
	if projectCtx, err := sm.client.GetProjectContext(ctx); err == nil {
		session.metadata["language"] = projectCtx.Language
		session.metadata["framework"] = projectCtx.Framework
		session.metadata["project_name"] = projectCtx.ProjectName
	}

	sm.sessions[sessionID] = session
	return session, nil
}

// GetSession retrieves an existing session by ID.
func (sm *ClaudeCodeSessionManager) GetSession(sessionID string) (*ClaudeCodeSession, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, sdkerrors.NewValidationError("sessionID", sessionID, "existing session", "session not found")
	}

	if session.IsExpired() {
		return nil, sdkerrors.NewValidationError("sessionID", sessionID, "active session", "session has expired")
	}

	return session, nil
}

// ListSessions returns all active session IDs.
func (sm *ClaudeCodeSessionManager) ListSessions() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessionIDs := make([]string, 0, len(sm.sessions))
	for id, session := range sm.sessions {
		if !session.IsExpired() {
			sessionIDs = append(sessionIDs, id)
		}
	}

	return sessionIDs
}

// CloseSession closes and removes a session.
func (sm *ClaudeCodeSessionManager) CloseSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if exists {
		session.Close()
		delete(sm.sessions, sessionID)
	}

	return nil
}

// GetSessionCount returns the number of active sessions.
func (sm *ClaudeCodeSessionManager) GetSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	count := 0
	for _, session := range sm.sessions {
		if !session.IsExpired() {
			count++
		}
	}

	return count
}

// Close shuts down the session manager and closes all sessions.
func (sm *ClaudeCodeSessionManager) Close() error {
	// Stop cleanup goroutine
	close(sm.stopCleanup)
	sm.cleanupTicker.Stop()
	sm.wg.Wait()

	// Close all sessions
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, session := range sm.sessions {
		session.Close()
	}
	sm.sessions = make(map[string]*ClaudeCodeSession)

	return nil
}

// runCleanup runs periodic cleanup of expired sessions.
func (sm *ClaudeCodeSessionManager) runCleanup() {
	defer sm.wg.Done()

	for {
		select {
		case <-sm.cleanupTicker.C:
			sm.cleanupExpiredSessions()
		case <-sm.stopCleanup:
			return
		}
	}
}

// cleanupExpiredSessions removes expired sessions.
func (sm *ClaudeCodeSessionManager) cleanupExpiredSessions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for id, session := range sm.sessions {
		if session.IsExpired() {
			session.Close()
			delete(sm.sessions, id)
		}
	}
}

// Query sends a query within this session using Claude Code.
// The session ID is passed to claude via the --session flag to maintain
// conversation context across queries.
func (s *ClaudeCodeSession) Query(ctx context.Context, request *types.QueryRequest) (*types.QueryResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, sdkerrors.NewInternalError("SESSION_CLOSED", "session has been closed")
	}

	if s.IsExpired() {
		return nil, sdkerrors.NewInternalError("SESSION_EXPIRED", "session has expired")
	}

	// Update last used time
	s.lastUsedAt = time.Now()

	// Create a session-aware request
	sessionRequest := s.buildSessionRequest(request)

	// Use the client's session ID for this query
	originalSessionID := s.client.sessionID
	s.client.sessionID = s.ID
	defer func() {
		s.client.sessionID = originalSessionID
	}()

	// Send the query
	response, err := s.client.Query(ctx, sessionRequest)
	if err != nil {
		return nil, sdkerrors.WrapError(err, sdkerrors.CategoryAPI, "SESSION_QUERY", "session query failed")
	}

	return response, nil
}

// QueryStream sends a streaming query within this session.
func (s *ClaudeCodeSession) QueryStream(ctx context.Context, request *types.QueryRequest) (types.QueryStream, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, sdkerrors.NewInternalError("SESSION_CLOSED", "session has been closed")
	}

	if s.IsExpired() {
		return nil, sdkerrors.NewInternalError("SESSION_EXPIRED", "session has expired")
	}

	// Update last used time
	s.lastUsedAt = time.Now()

	// Create a session-aware request
	sessionRequest := s.buildSessionRequest(request)

	// Use the client's session ID for this query
	originalSessionID := s.client.sessionID
	s.client.sessionID = s.ID

	// Send the streaming query
	stream, err := s.client.QueryStream(ctx, sessionRequest)
	if err != nil {
		s.client.sessionID = originalSessionID
		return nil, sdkerrors.WrapError(err, sdkerrors.CategoryAPI, "SESSION_STREAM", "session streaming query failed")
	}

	// Wrap the stream to restore session ID on close
	return &claudeCodeSessionStream{
		QueryStream:       stream,
		session:           s,
		originalSessionID: originalSessionID,
	}, nil
}

// ExecuteCommand executes a Claude Code command within this session.
func (s *ClaudeCodeSession) ExecuteCommand(ctx context.Context, cmd *types.Command) (*types.CommandResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, sdkerrors.NewInternalError("SESSION_CLOSED", "session has been closed")
	}

	if s.IsExpired() {
		return nil, sdkerrors.NewInternalError("SESSION_EXPIRED", "session has expired")
	}

	// Update last used time
	s.lastUsedAt = time.Now()

	// Use the client's session ID for this command
	originalSessionID := s.client.sessionID
	s.client.sessionID = s.ID
	defer func() {
		s.client.sessionID = originalSessionID
	}()

	// Execute the command
	return s.client.ExecuteCommand(ctx, cmd)
}

// ExecuteSlashCommand executes a slash command within this session.
func (s *ClaudeCodeSession) ExecuteSlashCommand(ctx context.Context, slashCommand string) (*types.CommandResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, sdkerrors.NewInternalError("SESSION_CLOSED", "session has been closed")
	}

	if s.IsExpired() {
		return nil, sdkerrors.NewInternalError("SESSION_EXPIRED", "session has expired")
	}

	// Update last used time
	s.lastUsedAt = time.Now()

	// Use the client's session ID for this command
	originalSessionID := s.client.sessionID
	s.client.sessionID = s.ID
	defer func() {
		s.client.sessionID = originalSessionID
	}()

	// Execute the slash command
	return s.client.ExecuteSlashCommand(ctx, slashCommand)
}

// buildSessionRequest creates a request configured for this session.
func (s *ClaudeCodeSession) buildSessionRequest(request *types.QueryRequest) *types.QueryRequest {
	// Create a copy of the request
	sessionRequest := *request

	// Use session's model if not specified
	if sessionRequest.Model == "" {
		sessionRequest.Model = s.model
	}

	return &sessionRequest
}

// GetMetadata returns session metadata.
func (s *ClaudeCodeSession) GetMetadata() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	metadata := make(map[string]interface{}, len(s.metadata))
	for k, v := range s.metadata {
		metadata[k] = v
	}
	return metadata
}

// SetMetadata sets session metadata.
func (s *ClaudeCodeSession) SetMetadata(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metadata[key] = value
}

// GetProjectDirectory returns the project directory for this session.
func (s *ClaudeCodeSession) GetProjectDirectory() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.projectDir
}

// SetProjectDirectory changes the project directory for this session.
func (s *ClaudeCodeSession) SetProjectDirectory(dir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return sdkerrors.WrapError(err, sdkerrors.CategoryValidation, "PROJECT_DIR", "failed to resolve project directory")
	}

	s.projectDir = absDir
	s.metadata["project_dir"] = absDir

	return nil
}

// IsExpired checks if the session has expired.
func (s *ClaudeCodeSession) IsExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return time.Since(s.lastUsedAt) > s.timeout
}

// GetAge returns how long the session has existed.
func (s *ClaudeCodeSession) GetAge() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return time.Since(s.createdAt)
}

// GetIdleTime returns how long the session has been idle.
func (s *ClaudeCodeSession) GetIdleTime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return time.Since(s.lastUsedAt)
}

// Refresh updates the last used time to prevent expiration.
func (s *ClaudeCodeSession) Refresh() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastUsedAt = time.Now()
}

// Close closes the session and cleans up resources.
func (s *ClaudeCodeSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.closed {
		s.closed = true
		s.metadata = nil
	}

	return nil
}

// claudeCodeSessionStream wraps a QueryStream for session management.
type claudeCodeSessionStream struct {
	types.QueryStream
	session           *ClaudeCodeSession
	originalSessionID string
}

// Close closes the stream and restores the original session ID.
func (css *claudeCodeSessionStream) Close() error {
	// Restore original session ID
	css.session.client.sessionID = css.originalSessionID

	// Close the underlying stream
	return css.QueryStream.Close()
}
