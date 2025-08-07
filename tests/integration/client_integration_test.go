//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ClaudeCodeIntegrationSuite tests real interactions with Claude Code CLI
type ClaudeCodeIntegrationSuite struct {
	suite.Suite
	client *client.ClaudeCodeClient
	config *types.ClaudeCodeConfig
}

func (s *ClaudeCodeIntegrationSuite) SetupSuite() {
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
	s.config.ClaudeExecutable = "claude" // Assume it's in PATH
	s.config.Timeout = 30 * time.Second
	
	// Enable TestMode in CI environment to skip Claude Code CLI requirement
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		s.config.TestMode = true
	}

	// Create client
	var err error
	ctx := context.Background()
	s.client, err = client.NewClaudeCodeClient(ctx, s.config)
	require.NoError(s.T(), err)
}

func (s *ClaudeCodeIntegrationSuite) TearDownSuite() {
	if s.client != nil {
		s.client.Close()
	}
}

func (s *ClaudeCodeIntegrationSuite) TestBasicQuery() {
	ctx := context.Background()

	// Simple synchronous query
	result, err := s.client.QueryMessagesSync(ctx, "What is 2 + 2?", nil)
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), result.Messages)
	// Check that the assistant's response contains "4"
	containsAnswer := false
	for _, msg := range result.Messages {
		if msg.Role == "assistant" && strings.Contains(strings.ToLower(msg.Content), "4") {
			containsAnswer = true
			break
		}
	}
	assert.True(s.T(), containsAnswer, "Response should contain the answer '4'")
}

func (s *ClaudeCodeIntegrationSuite) TestStreamingQuery() {
	ctx := context.Background()

	// Streaming query
	messages, err := s.client.QueryMessages(ctx, "Write a haiku about Go programming", nil)
	require.NoError(s.T(), err)

	messageCount := 0
	for msg := range messages {
		assert.NotNil(s.T(), msg)
		messageCount++
		
		// Should have role and content
		assert.NotEmpty(s.T(), msg.Role)
		assert.NotEmpty(s.T(), msg.GetText())
	}

	assert.Greater(s.T(), messageCount, 0, "Should receive at least one message")
}

func (s *ClaudeCodeIntegrationSuite) TestQueryWithOptions() {
	ctx := context.Background()

	options := &client.QueryOptions{
		Model: "claude-3-opus-20240229",
	}

	result, err := s.client.QueryMessagesSync(ctx, "Explain recursion in one sentence", options)
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), result.Messages)
	
	// Check that we got a response
	hasAssistantResponse := false
	for _, msg := range result.Messages {
		if msg.Role == "assistant" && len(msg.Content) > 0 {
			hasAssistantResponse = true
			break
		}
	}
	assert.True(s.T(), hasAssistantResponse, "Should have assistant response")
}

func (s *ClaudeCodeIntegrationSuite) TestContextCancellation() {
	// Create a context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start a query that would normally take longer
	messages, err := s.client.QueryMessages(ctx, "Write a detailed essay about software architecture", nil)
	
	if err != nil {
		// Should get context error
		assert.ErrorIs(s.T(), err, context.DeadlineExceeded)
		return
	}

	// If no immediate error, should get cancelled while reading messages
	messageCount := 0
	for range messages {
		messageCount++
	}
	
	// Should have received partial response before cancellation
	assert.Greater(s.T(), messageCount, 0)
}

func TestClaudeCodeIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ClaudeCodeIntegrationSuite))
}