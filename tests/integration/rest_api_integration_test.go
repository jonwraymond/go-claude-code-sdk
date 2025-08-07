//go:build integration
// +build integration

package integration

import (
    "context"
    "os"
    "testing"
    "time"

    api "github.com/jonwraymond/go-claude-code-sdk/pkg/api"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/suite"
)

type RESTIntegrationSuite struct {
    suite.Suite
    client *api.Client
}

func (s *RESTIntegrationSuite) SetupSuite() {
    if os.Getenv("INTEGRATION_TESTS") != "true" {
        s.T().Skip("Integration tests disabled. Set INTEGRATION_TESTS=true to run")
    }
    apiKey := os.Getenv("ANTHROPIC_API_KEY")
    if apiKey == "" {
        s.T().Skip("ANTHROPIC_API_KEY not set")
    }
    cfg := api.DefaultConfig()
    cfg.APIKey = apiKey
    cfg.Timeout = 60 * time.Second
    c, err := api.NewClient(cfg)
    require.NoError(s.T(), err)
    s.client = c
}

func (s *RESTIntegrationSuite) TestMessagesCreate() {
    ctx := context.Background()
    body := map[string]any{
        "model":      "claude-3-opus-20240229",
        "max_tokens": 64,
        "messages": []map[string]any{
            {"role": "user", "content": "Say 'ok'."},
        },
    }
    resp, err := s.client.MessagesCreate(ctx, body)
    require.NoError(s.T(), err)
    assert.NotEmpty(s.T(), resp["id"])
}

func (s *RESTIntegrationSuite) TestMessagesCountTokens() {
    ctx := context.Background()
    body := map[string]any{
        "model": "claude-3-opus-20240229",
        "messages": []map[string]any{
            {"role": "user", "content": "Count these tokens."},
        },
    }
    resp, err := s.client.MessagesCountTokens(ctx, body)
    require.NoError(s.T(), err)
    // The response should contain input_tokens
    _, ok := resp["input_tokens"]
    assert.True(s.T(), ok, "expected input_tokens in response")
}

func (s *RESTIntegrationSuite) TestBatchesList() {
    ctx := context.Background()
    resp, err := s.client.BatchesList(ctx, nil)
    require.NoError(s.T(), err)
    assert.NotNil(s.T(), resp)
}

func TestRESTIntegrationSuite(t *testing.T) {
    suite.Run(t, new(RESTIntegrationSuite))
}


