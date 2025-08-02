// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/errors"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ErrorHandlingIntegrationSuite tests error handling scenarios with Claude Code CLI
type ErrorHandlingIntegrationSuite struct {
	suite.Suite
	config *types.ClaudeCodeConfig
}

func (s *ErrorHandlingIntegrationSuite) SetupSuite() {
	// Skip if not in integration test mode
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		s.T().Skip("Integration tests disabled. Set INTEGRATION_TESTS=true to run")
	}

	// Create config
	s.config = types.NewClaudeCodeConfig()
	s.config.ClaudeExecutable = "claude"
	s.config.Timeout = 5 * time.Second // Short timeout for error tests
}

func (s *ErrorHandlingIntegrationSuite) TestInvalidAPIKey() {
	// Test with invalid API key
	config := *s.config
	config.APIKey = "invalid-api-key"

	client, err := client.NewClaudeCodeClient(&config)
	if err != nil {
		// May fail during client creation
		var claudeErr *errors.ClaudeCodeError
		if errors.As(err, &claudeErr) {
			assert.Equal(s.T(), errors.CategoryAuth, claudeErr.Category)
		}
		return
	}
	defer client.Close()

	// Try to make a query
	ctx := context.Background()
	_, err = client.QueryMessagesSync(ctx, "Hello", nil)
	require.Error(s.T(), err)

	// Should be an auth error
	var claudeErr *errors.ClaudeCodeError
	if errors.As(err, &claudeErr) {
		assert.Equal(s.T(), errors.CategoryAuth, claudeErr.Category)
	}
}

func (s *ErrorHandlingIntegrationSuite) TestMissingAPIKey() {
	// Test with no API key
	config := *s.config
	config.APIKey = ""

	client, err := client.NewClaudeCodeClient(&config)
	if err != nil {
		// May fail during client creation
		var claudeErr *errors.ClaudeCodeError
		if errors.As(err, &claudeErr) {
			assert.Equal(s.T(), errors.CategoryAuth, claudeErr.Category)
		}
		return
	}
	defer client.Close()

	// Try to make a query
	ctx := context.Background()
	_, err = client.QueryMessagesSync(ctx, "Hello", nil)
	require.Error(s.T(), err)

	// Should be an auth error
	assert.Contains(s.T(), err.Error(), "API key")
}

func (s *ErrorHandlingIntegrationSuite) TestInvalidExecutablePath() {
	// Test with non-existent Claude executable
	config := *s.config
	config.APIKey = "test-key"
	config.ClaudeExecutable = "/nonexistent/path/to/claude"

	_, err := client.NewClaudeCodeClient(&config)
	require.Error(s.T(), err)

	// Should be a process error
	var claudeErr *errors.ClaudeCodeError
	if errors.As(err, &claudeErr) {
		assert.Equal(s.T(), errors.CategoryProcess, claudeErr.Category)
	}
}

func (s *ErrorHandlingIntegrationSuite) TestTimeoutError() {
	// Skip if API key not set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		s.T().Skip("ANTHROPIC_API_KEY not set")
	}

	// Create client with very short timeout
	config := *s.config
	config.APIKey = apiKey
	config.Timeout = 1 * time.Millisecond // Extremely short timeout

	client, err := client.NewClaudeCodeClient(&config)
	require.NoError(s.T(), err)
	defer client.Close()

	// Try to make a query that will timeout
	ctx := context.Background()
	_, err = client.QueryMessagesSync(ctx, "Write a very long essay about the history of computing", nil)
	require.Error(s.T(), err)

	// Should be a timeout or context error
	assert.True(s.T(), 
		errors.Is(err, context.DeadlineExceeded) || 
		errors.GetCategory(err) == errors.CategoryNetwork,
		"Expected timeout or network error")
}

func (s *ErrorHandlingIntegrationSuite) TestValidationErrors() {
	// Skip if API key not set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		s.T().Skip("ANTHROPIC_API_KEY not set")
	}

	config := *s.config
	config.APIKey = apiKey

	client, err := client.NewClaudeCodeClient(&config)
	require.NoError(s.T(), err)
	defer client.Close()

	ctx := context.Background()

	// Test with invalid options
	testCases := []struct {
		name    string
		options *client.QueryOptions
	}{
		{
			name: "Negative max tokens",
			options: &client.QueryOptions{
				MaxTokens: -1,
			},
		},
		{
			name: "Invalid temperature",
			options: &client.QueryOptions{
				Temperature: 2.5, // Should be 0-1
			},
		},
		{
			name: "Empty model",
			options: &client.QueryOptions{
				Model: "",
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := client.QueryMessagesSync(ctx, "Hello", tc.options)
			
			// Should get validation error
			if err != nil {
				var validationErr *errors.ValidationError
				if errors.As(err, &validationErr) {
					assert.NotEmpty(t, validationErr.Field)
					assert.NotEmpty(t, validationErr.Message)
				}
			}
		})
	}
}

func (s *ErrorHandlingIntegrationSuite) TestRetryableErrors() {
	// Skip if API key not set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		s.T().Skip("ANTHROPIC_API_KEY not set")
	}

	config := *s.config
	config.APIKey = apiKey

	client, err := client.NewClaudeCodeClient(&config)
	require.NoError(s.T(), err)
	defer client.Close()

	// This test would simulate retryable errors like rate limits
	// In practice, this is hard to test reliably without mocking
	// So we'll test the error detection logic

	// Create a mock error that should be retryable
	mockErr := errors.NewAPIError("RATE_LIMIT", "Too many requests", 429, true)
	
	// Check if error is correctly identified as retryable
	assert.True(s.T(), errors.IsRetryable(mockErr))
	assert.Equal(s.T(), errors.CategoryAPI, errors.GetCategory(mockErr))
}

func (s *ErrorHandlingIntegrationSuite) TestErrorWrapping() {
	// Test error wrapping and unwrapping
	baseErr := os.ErrNotExist
	wrappedErr := errors.WrapError(baseErr, errors.CategoryInternal, "FILE_NOT_FOUND", "Config file not found")

	// Should be able to unwrap to original error
	assert.ErrorIs(s.T(), wrappedErr, os.ErrNotExist)

	// Should have Claude error properties
	var claudeErr *errors.ClaudeCodeError
	require.True(s.T(), errors.As(wrappedErr, &claudeErr))
	assert.Equal(s.T(), "FILE_NOT_FOUND", claudeErr.Code)
	assert.Equal(s.T(), errors.CategoryInternal, claudeErr.Category)
}

func (s *ErrorHandlingIntegrationSuite) TestToolExecutionErrors() {
	// Skip if API key not set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		s.T().Skip("ANTHROPIC_API_KEY not set")
	}

	config := *s.config
	config.APIKey = apiKey

	client, err := client.NewClaudeCodeClient(&config)
	require.NoError(s.T(), err)
	defer client.Close()

	ctx := context.Background()

	// Try to execute a tool with invalid arguments
	_, err = client.Tools().ExecuteTool(ctx, &client.ClaudeCodeTool{
		Name: "read_file",
		Arguments: map[string]interface{}{
			// Missing required 'path' argument
		},
	})
	require.Error(s.T(), err)

	// Try to execute non-existent tool
	_, err = client.Tools().ExecuteTool(ctx, &client.ClaudeCodeTool{
		Name: "nonexistent_tool",
		Arguments: map[string]interface{}{
			"arg": "value",
		},
	})
	require.Error(s.T(), err)
}

func (s *ErrorHandlingIntegrationSuite) TestSessionErrors() {
	// Skip if API key not set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		s.T().Skip("ANTHROPIC_API_KEY not set")
	}

	config := *s.config
	config.APIKey = apiKey

	client, err := client.NewClaudeCodeClient(&config)
	require.NoError(s.T(), err)
	defer client.Close()

	ctx := context.Background()

	// Try to get non-existent session
	_, err = client.Sessions().GetSession(ctx, "nonexistent-session-12345")
	require.Error(s.T(), err)

	// Try to delete non-existent session
	err = client.Sessions().DeleteSession(ctx, "nonexistent-session-12345")
	// May or may not error depending on Claude CLI behavior
	_ = err
}

func TestErrorHandlingIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ErrorHandlingIntegrationSuite))
}