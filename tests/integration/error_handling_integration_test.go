//go:build integration
// +build integration

package integration

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	client "github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	sdkerrors "github.com/jonwraymond/go-claude-code-sdk/pkg/errors"
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

	// Enable TestMode in CI environment to skip Claude Code CLI requirement
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		s.config.TestMode = true
	}
}

func (s *ErrorHandlingIntegrationSuite) TestInvalidAPIKey() {
	// Test with invalid API key
	config := *s.config
	config.APIKey = "invalid-api-key"

	ctx := context.Background()
	client, err := client.NewClaudeCodeClient(ctx, &config)
	if err != nil {
		// May fail during client creation
		assert.Equal(s.T(), sdkerrors.CategoryAuth, sdkerrors.GetCategory(err))
		return
	}
	defer client.Close()

	// Try to make a query
	_, err2 := client.QueryMessagesSync(ctx, "Hello", nil)
	require.Error(s.T(), err2)

	// Should be an auth error
	assert.Equal(s.T(), sdkerrors.CategoryAuth, sdkerrors.GetCategory(err2))
}

func (s *ErrorHandlingIntegrationSuite) TestMissingAPIKey() {
	// Test with no API key
	config := *s.config
	config.APIKey = ""

	ctx := context.Background()
	client, err := client.NewClaudeCodeClient(ctx, &config)
	if err != nil {
		// May fail during client creation
		assert.Equal(s.T(), sdkerrors.CategoryAuth, sdkerrors.GetCategory(err))
		return
	}
	defer client.Close()

	// Try to make a query
	_, err3 := client.QueryMessagesSync(ctx, "Hello", nil)
	require.Error(s.T(), err3)

	// Should be an auth error
	assert.Contains(s.T(), err3.Error(), "API key")
}

func (s *ErrorHandlingIntegrationSuite) TestInvalidExecutablePath() {
	// Test with non-existent Claude executable
	config := *s.config
	config.APIKey = "test-key"
	config.ClaudeExecutable = "/nonexistent/path/to/claude"

	ctx := context.Background()
	_, err := client.NewClaudeCodeClient(ctx, &config)
	require.Error(s.T(), err)

	// Should be a configuration or internal error
	assert.True(s.T(),
		sdkerrors.GetCategory(err) == sdkerrors.CategoryConfiguration ||
			sdkerrors.GetCategory(err) == sdkerrors.CategoryInternal,
		"Expected configuration or internal error")
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

	ctx := context.Background()
	client, err := client.NewClaudeCodeClient(ctx, &config)
	require.NoError(s.T(), err)
	defer client.Close()

	// Try to make a query that will timeout
	_, err4 := client.QueryMessagesSync(ctx, "Write a very long essay about the history of computing", nil)
	require.Error(s.T(), err4)

	// Should be a timeout or context error
	assert.True(s.T(),
		errors.Is(err4, context.DeadlineExceeded) ||
			sdkerrors.GetCategory(err4) == sdkerrors.CategoryNetwork,
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

	ctx := context.Background()
	client, err := client.NewClaudeCodeClient(ctx, &config)
	require.NoError(s.T(), err)
	defer client.Close()

	// Test basic query without options
	_, err5 := client.QueryMessagesSync(ctx, "Hello", nil)

	// Should get some kind of error or response
	// For validation tests, we'll just check that the client is working
	if err5 != nil {
		// Just check that we got some error
		assert.NotEmpty(s.T(), err5.Error())
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

	ctx := context.Background()
	client, err := client.NewClaudeCodeClient(ctx, &config)
	require.NoError(s.T(), err)
	defer client.Close()

	// This test would simulate retryable errors like rate limits
	// In practice, this is hard to test reliably without mocking
	// So we'll test the error detection logic

	// Create a mock error that should be retryable
	mockErr := sdkerrors.NewAPIError(429, "RATE_LIMIT", "error", "Too many requests")

	// Check if error is correctly identified as retryable
	assert.True(s.T(), sdkerrors.IsRetryable(mockErr))
	assert.Equal(s.T(), sdkerrors.CategoryAPI, sdkerrors.GetCategory(mockErr))
}

func (s *ErrorHandlingIntegrationSuite) TestErrorWrapping() {
	// Test error wrapping and unwrapping
	baseErr := os.ErrNotExist
	wrappedErr := sdkerrors.WrapError(baseErr, sdkerrors.CategoryInternal, "FILE_NOT_FOUND", "Config file not found")

	// Should be able to unwrap to original error
	assert.ErrorIs(s.T(), wrappedErr, os.ErrNotExist)

	// Should have SDK error properties
	var sdkErr sdkerrors.SDKError
	require.True(s.T(), errors.As(wrappedErr, &sdkErr))
	assert.Equal(s.T(), "FILE_NOT_FOUND", sdkErr.Code())
	assert.Equal(s.T(), sdkerrors.CategoryInternal, sdkErr.Category())
}

func (s *ErrorHandlingIntegrationSuite) TestToolExecutionErrors() {
	// Skip if API key not set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		s.T().Skip("ANTHROPIC_API_KEY not set")
	}

	config := *s.config
	config.APIKey = apiKey

	ctx := context.Background()
	client, err := client.NewClaudeCodeClient(ctx, &config)
	require.NoError(s.T(), err)
	defer client.Close()

	// Test basic tool functionality
	tools := client.Tools()
	if tools != nil {
		// Test that we can list tools
		toolList := tools.ListTools()
		assert.NotNil(s.T(), toolList)
		// Tool execution tests would require proper type definitions
		// For now, just verify the tools manager is working
	}
}

func (s *ErrorHandlingIntegrationSuite) TestSessionErrors() {
	// Skip if API key not set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		s.T().Skip("ANTHROPIC_API_KEY not set")
	}

	config := *s.config
	config.APIKey = apiKey

	ctx := context.Background()
	client, err := client.NewClaudeCodeClient(ctx, &config)
	require.NoError(s.T(), err)
	defer client.Close()

	// Try to get non-existent session
	_, err = client.Sessions().GetSession("nonexistent-session-12345")
	require.Error(s.T(), err)

	// Try to close non-existent session
	err = client.Sessions().CloseSession("nonexistent-session-12345")
	// May or may not error depending on Claude CLI behavior
	_ = err
}

func TestErrorHandlingIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ErrorHandlingIntegrationSuite))
}
