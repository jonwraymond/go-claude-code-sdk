package client

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func TestNewClaudeCodeSessionManager(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		TestMode:         true, // Skip Claude Code CLI requirement for testing
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	sessionManager := client.Sessions()
	if sessionManager == nil {
		t.Fatal("Session manager should not be nil")
	}

	// Check initial state
	if sessionManager.GetSessionCount() != 0 {
		t.Error("Expected no sessions initially")
	}

	sessions := sessionManager.ListSessions()
	if len(sessions) != 0 {
		t.Error("Expected empty session list initially")
	}
}

func TestClaudeCodeSessionManager_CreateSession(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		TestMode:         true, // Skip Claude Code CLI requirement for testing
		WorkingDirectory: tempDir,
		Model:            types.ModelClaude35Sonnet,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	sessionManager := client.Sessions()

	// Generate proper UUID session IDs
	sessionID1 := client.GenerateSessionID()
	sessionID2 := client.GenerateSessionID()

	tests := []struct {
		name        string
		sessionID   string
		expectError bool
	}{
		{
			name:        "Create new session",
			sessionID:   sessionID1,
			expectError: false,
		},
		{
			name:        "Create another session",
			sessionID:   sessionID2,
			expectError: false,
		},
		{
			name:        "Get existing session",
			sessionID:   sessionID1,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := sessionManager.CreateSession(ctx, tt.sessionID)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if session == nil {
					t.Fatal("Expected session to be created")
				}
				if session.ID != tt.sessionID {
					t.Errorf("Expected session ID %s, got %s", tt.sessionID, session.ID)
				}

				// Check metadata
				metadata := session.GetMetadata()
				if metadata["project_dir"] != tempDir {
					t.Errorf("Expected project_dir %s in metadata, got %v", tempDir, metadata["project_dir"])
				}
				if metadata["model"] != config.Model {
					t.Errorf("Expected model %s in metadata, got %v", config.Model, metadata["model"])
				}
			}
		})
	}

	// Check session count
	if sessionManager.GetSessionCount() != 2 {
		t.Errorf("Expected 2 sessions, got %d", sessionManager.GetSessionCount())
	}

	// Check session list
	sessions := sessionManager.ListSessions()
	if len(sessions) != 2 {
		t.Errorf("Expected 2 session IDs, got %d", len(sessions))
	}
}

func TestClaudeCodeSessionManager_MaxSessions(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		TestMode:         true, // Skip Claude Code CLI requirement for testing
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	// Create session manager with low max sessions
	sessionConfig := &ClaudeCodeSessionConfig{
		MaxSessions:      3,
		SessionTimeout:   1 * time.Hour,
		CleanupInterval:  10 * time.Minute,
		PersistSessions:  true,
		SessionDirectory: ".claude/sessions",
	}
	sessionManager := NewClaudeCodeSessionManagerWithConfig(client, sessionConfig)
	defer sessionManager.Close()

	// Create sessions up to the limit
	for i := 0; i < 3; i++ {
		sessionID := fmt.Sprintf("session-%d", i)
		_, err := sessionManager.CreateSession(ctx, sessionID)
		if err != nil {
			t.Fatalf("Failed to create session %s: %v", sessionID, err)
		}
	}

	// Try to create one more session (should fail)
	_, err = sessionManager.CreateSession(ctx, "session-overflow")
	if err == nil {
		t.Error("Expected error when exceeding max sessions")
	}
}

func TestClaudeCodeSessionManager_GetSession(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		TestMode:         true, // Skip Claude Code CLI requirement for testing
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	sessionManager := client.Sessions()

	// Create a session with proper UUID
	sessionID := client.GenerateSessionID()
	session1, err := sessionManager.CreateSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Get the session using the normalized ID
	session2, err := sessionManager.GetSession(session1.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if session2.ID != session1.ID {
		t.Errorf("Expected session ID %s, got %s", session1.ID, session2.ID)
	}

	// Try to get non-existent session
	_, err = sessionManager.GetSession("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent session")
	}
}

func TestClaudeCodeSessionManager_CloseSession(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		TestMode:         true, // Skip Claude Code CLI requirement for testing
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	sessionManager := client.Sessions()

	// Create a session with proper UUID
	sessionID := client.GenerateSessionID()
	session, err := sessionManager.CreateSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Verify session exists
	if sessionManager.GetSessionCount() != 1 {
		t.Error("Expected 1 session")
	}

	// Close the session using the normalized ID
	err = sessionManager.CloseSession(session.ID)
	if err != nil {
		t.Errorf("Failed to close session: %v", err)
	}

	// Verify session is removed
	if sessionManager.GetSessionCount() != 0 {
		t.Error("Expected 0 sessions after closing")
	}

	// Try to get the closed session
	_, err = sessionManager.GetSession(session.ID)
	if err == nil {
		t.Error("Expected error for closed session")
	}
}

func TestClaudeCodeSessionManager_Cleanup(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		TestMode:         true, // Skip Claude Code CLI requirement for testing
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	// Create session manager with short timeout
	sessionConfig := &ClaudeCodeSessionConfig{
		MaxSessions:      10,
		SessionTimeout:   100 * time.Millisecond,
		CleanupInterval:  50 * time.Millisecond,
		PersistSessions:  true,
		SessionDirectory: ".claude/sessions",
	}
	sessionManager := NewClaudeCodeSessionManagerWithConfig(client, sessionConfig)
	defer sessionManager.Close()

	// Create a session
	_, err = sessionManager.CreateSession(ctx, "test-session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Verify session exists
	if sessionManager.GetSessionCount() != 1 {
		t.Error("Expected 1 session")
	}

	// Wait for session to expire and be cleaned up
	time.Sleep(200 * time.Millisecond)

	// Verify session is cleaned up
	if sessionManager.GetSessionCount() != 0 {
		t.Error("Expected 0 sessions after cleanup")
	}
}

func TestClaudeCodeSession_Metadata(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		TestMode:         true, // Skip Claude Code CLI requirement for testing
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	session, err := client.CreateSession(ctx, "test-session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Set metadata
	session.SetMetadata("user_id", "user-123")
	session.SetMetadata("project_type", "web-app")

	// Get metadata
	metadata := session.GetMetadata()
	if metadata["user_id"] != "user-123" {
		t.Errorf("Expected user_id 'user-123', got %v", metadata["user_id"])
	}
	if metadata["project_type"] != "web-app" {
		t.Errorf("Expected project_type 'web-app', got %v", metadata["project_type"])
	}

	// Verify original metadata is still there
	if metadata["project_dir"] != tempDir {
		t.Errorf("Expected project_dir %s, got %v", tempDir, metadata["project_dir"])
	}
}

func TestClaudeCodeSession_ProjectDirectory(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		TestMode:         true, // Skip Claude Code CLI requirement for testing
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	session, err := client.CreateSession(ctx, "test-session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Get initial project directory
	projectDir := session.GetProjectDirectory()
	if projectDir != tempDir {
		t.Errorf("Expected project directory %s, got %s", tempDir, projectDir)
	}

	// Change project directory
	newDir := t.TempDir()
	err = session.SetProjectDirectory(newDir)
	if err != nil {
		t.Errorf("Failed to set project directory: %v", err)
	}

	// Verify change
	projectDir = session.GetProjectDirectory()
	if projectDir != newDir {
		t.Errorf("Expected project directory %s, got %s", newDir, projectDir)
	}

	// Check metadata update
	metadata := session.GetMetadata()
	if metadata["project_dir"] != newDir {
		t.Errorf("Expected project_dir %s in metadata, got %v", newDir, metadata["project_dir"])
	}
}

func TestClaudeCodeSession_Lifecycle(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		TestMode:         true, // Skip Claude Code CLI requirement for testing
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	// Create session with short timeout
	sessionConfig := &ClaudeCodeSessionConfig{
		MaxSessions:      10,
		SessionTimeout:   100 * time.Millisecond,
		CleanupInterval:  1 * time.Minute,
		PersistSessions:  true,
		SessionDirectory: ".claude/sessions",
	}
	sessionManager := NewClaudeCodeSessionManagerWithConfig(client, sessionConfig)
	defer sessionManager.Close()

	session, err := sessionManager.CreateSession(ctx, "test-session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Check initial state
	if session.IsExpired() {
		t.Error("New session should not be expired")
	}

	// Check age
	age := session.GetAge()
	if age > 100*time.Millisecond {
		t.Errorf("New session age should be minimal, got %v", age)
	}

	// Check idle time
	idleTime := session.GetIdleTime()
	if idleTime > 100*time.Millisecond {
		t.Errorf("New session idle time should be minimal, got %v", idleTime)
	}

	// Wait for session to expire
	time.Sleep(150 * time.Millisecond)

	// Check expired state
	if !session.IsExpired() {
		t.Error("Session should be expired after timeout")
	}

	// Refresh the session
	session.Refresh()

	// Check not expired after refresh
	if session.IsExpired() {
		t.Error("Session should not be expired after refresh")
	}

	// Close the session
	err = session.Close()
	if err != nil {
		t.Errorf("Failed to close session: %v", err)
	}

	// Try to use closed session (should fail)
	request := &types.QueryRequest{
		Messages: []types.Message{
			{Role: types.RoleUser, Content: "Test"},
		},
	}
	_, err = session.Query(ctx, request)
	if err == nil {
		t.Error("Expected error when using closed session")
	}
}

func TestClaudeCodeSession_ExecuteCommand(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		TestMode:         true, // Skip Claude Code CLI requirement for testing
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	session, err := client.CreateSession(ctx, "test-session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test with valid command (would require Claude Code to be installed)
	cmd := &types.Command{
		Type: CommandRead,
		Args: []string{"test.txt"},
	}

	// This will fail without Claude Code installed, but we're testing the session functionality
	_, err = session.ExecuteCommand(ctx, cmd)
	// We expect an error since Claude Code isn't installed in test environment
	if err == nil {
		t.Skip("Claude Code appears to be installed - skipping error test")
	}
}

func TestClaudeCodeSession_ExecuteSlashCommand(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		TestMode:         true, // Skip Claude Code CLI requirement for testing
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	session, err := client.CreateSession(ctx, "test-session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test with slash command (would require Claude Code to be installed)
	// This will fail without Claude Code installed, but we're testing the session functionality
	_, err = session.ExecuteSlashCommand(ctx, "/help")
	// We expect an error since Claude Code isn't installed in test environment
	if err == nil {
		t.Skip("Claude Code appears to be installed - skipping error test")
	}
}

func TestClaudeCodeClient_SessionIntegration(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		TestMode:         true, // Skip Claude Code CLI requirement for testing
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	// Test client-level session methods with proper UUIDs
	sessionID1 := client.GenerateSessionID()
	sessionID2 := client.GenerateSessionID()

	session1, err := client.CreateSession(ctx, sessionID1)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	session2, err := client.CreateSession(ctx, sessionID2)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// List sessions
	sessions := client.ListSessions()
	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}

	// Get session using the normalized ID
	retrievedSession, err := client.GetSession(session1.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrievedSession.ID != session1.ID {
		t.Errorf("Expected session ID %s, got %s", session1.ID, retrievedSession.ID)
	}

	// Close session through manager using normalized ID
	err = client.Sessions().CloseSession(session2.ID)
	if err != nil {
		t.Errorf("Failed to close session: %v", err)
	}

	// Verify only one session remains
	sessions = client.ListSessions()
	if len(sessions) != 1 {
		t.Errorf("Expected 1 session after closing, got %d", len(sessions))
	}

	if sessions[0] != session1.ID {
		t.Errorf("Expected remaining session to be %s, got %s", session1.ID, sessions[0])
	}
}
