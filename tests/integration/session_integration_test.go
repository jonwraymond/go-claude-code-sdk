// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jraymond/claude-code-go-sdk/pkg/client"
	"github.com/jraymond/claude-code-go-sdk/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// SessionIntegrationSuite tests session management with Claude Code CLI
type SessionIntegrationSuite struct {
	suite.Suite
	client         *client.ClaudeCodeClient
	sessionManager *client.SessionManager
	config         *types.ClaudeCodeConfig
}

func (s *SessionIntegrationSuite) SetupSuite() {
	// Skip if not in integration test mode
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		s.T().Skip("Integration tests disabled. Set INTEGRATION_TESTS=true to run")
	}

	// Ensure API key is set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		s.T().Skip("ANTHROPIC_API_KEY not set")
	}

	// Create config
	s.config = types.NewClaudeCodeConfig()
	s.config.APIKey = apiKey
	s.config.ClaudeExecutable = "claude"
	s.config.Timeout = 30 * time.Second

	// Create client
	var err error
	s.client, err = client.NewClaudeCodeClient(s.config)
	require.NoError(s.T(), err)

	// Get session manager
	s.sessionManager = s.client.Sessions()
}

func (s *SessionIntegrationSuite) TearDownSuite() {
	if s.client != nil {
		s.client.Close()
	}
}

func (s *SessionIntegrationSuite) TestCreateAndUseSession() {
	ctx := context.Background()
	sessionID := fmt.Sprintf("test-session-%d", time.Now().Unix())

	// Create session
	session, err := s.sessionManager.CreateSession(ctx, sessionID)
	require.NoError(s.T(), err)
	defer session.Close()

	assert.Equal(s.T(), sessionID, session.ID())
	assert.True(s.T(), session.IsActive())

	// First query in session
	req1 := &types.QueryRequest{
		Messages: []types.Message{
			{Role: types.MessageRoleUser, Content: "My name is TestBot. Remember this."},
		},
	}
	
	resp1, err := session.Query(ctx, req1)
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), resp1.Content)

	// Second query should remember context
	req2 := &types.QueryRequest{
		Messages: []types.Message{
			{Role: types.MessageRoleUser, Content: "What is my name?"},
		},
	}
	
	resp2, err := session.Query(ctx, req2)
	require.NoError(s.T(), err)
	assert.Contains(s.T(), resp2.Content, "TestBot")
}

func (s *SessionIntegrationSuite) TestListSessions() {
	ctx := context.Background()

	// Create a test session
	sessionID := fmt.Sprintf("test-list-%d", time.Now().Unix())
	session, err := s.sessionManager.CreateSession(ctx, sessionID)
	require.NoError(s.T(), err)
	defer session.Close()

	// List sessions
	sessions, err := s.sessionManager.ListSessions(ctx)
	require.NoError(s.T(), err)

	// Should contain our session
	found := false
	for _, s := range sessions {
		if s.ID == sessionID {
			found = true
			break
		}
	}
	assert.True(s.T(), found, "Created session should be in list")
}

func (s *SessionIntegrationSuite) TestGetSession() {
	ctx := context.Background()
	sessionID := fmt.Sprintf("test-get-%d", time.Now().Unix())

	// Create session
	session1, err := s.sessionManager.CreateSession(ctx, sessionID)
	require.NoError(s.T(), err)
	defer session1.Close()

	// Get the same session
	session2, err := s.sessionManager.GetSession(ctx, sessionID)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), session1.ID(), session2.ID())
	assert.True(s.T(), session2.IsActive())
}

func (s *SessionIntegrationSuite) TestDeleteSession() {
	ctx := context.Background()
	sessionID := fmt.Sprintf("test-delete-%d", time.Now().Unix())

	// Create session
	session, err := s.sessionManager.CreateSession(ctx, sessionID)
	require.NoError(s.T(), err)

	// Delete session
	err = s.sessionManager.DeleteSession(ctx, sessionID)
	require.NoError(s.T(), err)

	// Session should no longer be active
	assert.False(s.T(), session.IsActive())

	// Getting deleted session should fail
	_, err = s.sessionManager.GetSession(ctx, sessionID)
	assert.Error(s.T(), err)
}

func (s *SessionIntegrationSuite) TestSessionWithProjectContext() {
	ctx := context.Background()
	sessionID := fmt.Sprintf("test-project-%d", time.Now().Unix())

	// Create session with project context
	session, err := s.sessionManager.CreateSessionWithProject(ctx, sessionID, ".")
	require.NoError(s.T(), err)
	defer session.Close()

	// Query about the project
	req := &types.QueryRequest{
		Messages: []types.Message{
			{Role: types.MessageRoleUser, Content: "What kind of project is this?"},
		},
	}
	
	resp, err := session.Query(ctx, req)
	require.NoError(s.T(), err)
	
	// Should understand it's a Go SDK project
	assert.Contains(s.T(), resp.Content, "Go")
}

func (s *SessionIntegrationSuite) TestMultipleConcurrentSessions() {
	ctx := context.Background()

	// Create multiple sessions
	session1ID := fmt.Sprintf("test-concurrent-1-%d", time.Now().Unix())
	session2ID := fmt.Sprintf("test-concurrent-2-%d", time.Now().Unix())

	session1, err := s.sessionManager.CreateSession(ctx, session1ID)
	require.NoError(s.T(), err)
	defer session1.Close()

	session2, err := s.sessionManager.CreateSession(ctx, session2ID)
	require.NoError(s.T(), err)
	defer session2.Close()

	// Use both sessions concurrently
	done1 := make(chan bool)
	done2 := make(chan bool)

	go func() {
		req := &types.QueryRequest{
			Messages: []types.Message{
				{Role: types.MessageRoleUser, Content: "Count from 1 to 3"},
			},
		}
		_, err := session1.Query(ctx, req)
		assert.NoError(s.T(), err)
		done1 <- true
	}()

	go func() {
		req := &types.QueryRequest{
			Messages: []types.Message{
				{Role: types.MessageRoleUser, Content: "Count from 4 to 6"},
			},
		}
		_, err := session2.Query(ctx, req)
		assert.NoError(s.T(), err)
		done2 <- true
	}()

	// Wait for both to complete
	select {
	case <-done1:
		<-done2
	case <-time.After(30 * time.Second):
		s.T().Fatal("Concurrent sessions timed out")
	}
}

func TestSessionIntegrationSuite(t *testing.T) {
	suite.Run(t, new(SessionIntegrationSuite))
}