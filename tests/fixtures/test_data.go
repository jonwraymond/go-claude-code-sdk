// Package fixtures provides test data and fixtures for the Go Claude Code SDK tests.
package fixtures

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// TestMessages provides sample messages for testing
var TestMessages = struct {
	Simple       []types.Message
	MultiTurn    []types.Message
	WithTools    []types.Message
	WithImages   []types.Message
	EdgeCases    []types.Message
	LargeContext []types.Message
}{
	Simple: []types.Message{
		{
			Role:    types.RoleUser,
			Content: "Hello, Claude!",
		},
	},
	MultiTurn: []types.Message{
		{
			Role:    types.RoleUser,
			Content: "What is the capital of France?",
		},
		{
			Role:    types.RoleAssistant,
			Content: "The capital of France is Paris.",
		},
		{
			Role:    types.RoleUser,
			Content: "What is its population?",
		},
	},
	WithTools: []types.Message{
		{
			Role:    types.RoleUser,
			Content: "Read the file test.txt",
		},
		{
			Role:    types.RoleAssistant,
			Content: "I'll read the file test.txt for you.",
			ToolCalls: []types.ToolCall{
				{
					ID:   "tool_123",
					Type: "function",
					Function: types.FunctionCall{
						Name:      "read_file",
						Arguments: `{"path": "test.txt"}`,
					},
				},
			},
		},
		{
			Role:       types.RoleTool,
			Content:    "File content: Hello World",
			ToolCallID: "tool_123",
		},
	},
	WithImages: []types.Message{
		{
			Role:    types.RoleUser,
			Content: "What's in this image? [image data would be attached separately]",
		},
	},
	EdgeCases: []types.Message{
		{
			Role:    types.RoleUser,
			Content: "", // Empty content
		},
		{
			Role:    types.RoleUser,
			Content: string(make([]byte, 10000)), // Large content
		},
		{
			Role:    types.RoleUser,
			Content: "Special chars: 你好 🌍 émojis ñ", // Unicode
		},
	},
	LargeContext: generateLargeContext(100),
}

// TestResponses provides sample API responses
var TestResponses = struct {
	Success        types.QueryResponse
	WithToolUse    types.QueryResponse
	Streaming      []types.StreamEvent
	Error          types.QueryResponse
	RateLimited    types.QueryResponse
	PartialContent types.QueryResponse
}{
	Success: types.QueryResponse{
		ID:   "msg_01234567890",
		Type: "message",
		Role: types.RoleAssistant,
		Content: []types.ContentBlock{
			{
				Type: "text",
				Text: "This is a successful response.",
			},
		},
		Model:      "claude-3-5-sonnet-20241022",
		StopReason: "end_turn",
		Usage: &types.TokenUsage{
			InputTokens:  50,
			OutputTokens: 100,
			TotalTokens:  150,
		},
		CreatedAt: time.Now(),
	},
	WithToolUse: types.QueryResponse{
		ID:   "msg_with_tools",
		Type: "message",
		Role: types.RoleAssistant,
		Content: []types.ContentBlock{
			{
				Type: "text",
				Text: "I'll help you with that file.",
			},
			{
				Type: "tool_use",
				ID:   "tool_use_123",
				Name: "read_file",
				Input: map[string]interface{}{
					"path": "example.txt",
				},
			},
		},
		Model:      "claude-3-5-sonnet-20241022",
		StopReason: "tool_use",
		Usage: &types.TokenUsage{
			InputTokens:  75,
			OutputTokens: 125,
			TotalTokens:  200,
		},
		CreatedAt: time.Now(),
	},
	Streaming: []types.StreamEvent{
		{
			Type: types.StreamEventMessageStart,
			Message: &types.StreamMessage{
				ID:    "msg_stream",
				Type:  "message",
				Role:  types.RoleAssistant,
				Model: "claude-3-5-sonnet-20241022",
			},
		},
		{
			Type:  types.StreamEventContentBlockStart,
			Index: 0,
			ContentBlock: &types.ContentBlock{
				Type: "text",
				Text: "",
			},
		},
		{
			Type:  types.StreamEventContentBlockDelta,
			Index: 0,
			ContentDelta: &types.ContentDelta{
				Type: "text_delta",
				Text: "Hello, ",
			},
		},
		{
			Type:  types.StreamEventContentBlockDelta,
			Index: 0,
			ContentDelta: &types.ContentDelta{
				Type: "text_delta",
				Text: "world!",
			},
		},
		{
			Type:  types.StreamEventContentBlockStop,
			Index: 0,
		},
		{
			Type: types.StreamEventMessageStop,
			MessageDelta: &types.MessageDelta{
				StopReason: "end_turn",
			},
			Usage: &types.TokenUsage{
				OutputTokens: 10,
			},
		},
	},
	Error: types.QueryResponse{
		ID:   "msg_error",
		Type: "message",
		Role: types.RoleAssistant,
		Content: []types.ContentBlock{
			{
				Type: "text",
				Text: "Error: Invalid model specified",
			},
		},
		Model: "claude-3-5-sonnet-20241022",
	},
	RateLimited: types.QueryResponse{
		ID:   "msg_rate_limited",
		Type: "message",
		Role: types.RoleAssistant,
		Content: []types.ContentBlock{
			{
				Type: "text",
				Text: "Error: Rate limit exceeded",
			},
		},
		Model: "claude-3-5-sonnet-20241022",
	},
	PartialContent: types.QueryResponse{
		ID:   "msg_partial",
		Type: "message",
		Role: types.RoleAssistant,
		Content: []types.ContentBlock{
			{
				Type: "text",
				Text: "This response was cut off due to...",
			},
		},
		Model:      "claude-3-5-sonnet-20241022",
		StopReason: "max_tokens",
		Usage: &types.TokenUsage{
			InputTokens:  100,
			OutputTokens: 4096,
			TotalTokens:  4196,
		},
		CreatedAt: time.Now(),
	},
}

// getTestAPIKey returns a test API key that doesn't trigger security scanners
func getTestAPIKey() string {
	// Use environment variable if set, otherwise use a non-secret test value
	if key := os.Getenv("TEST_API_KEY"); key != "" {
		return key
	}
	return "test-key-for-unit-tests"
}

// TestConfigs provides sample configurations
var TestConfigs = struct {
	Default     types.ClaudeCodeConfig
	WithAPIKey  types.ClaudeCodeConfig
	WithSession types.ClaudeCodeConfig
	Custom      types.ClaudeCodeConfig
	Invalid     types.ClaudeCodeConfig
}{
	Default: types.ClaudeCodeConfig{
		Model: "claude-3-5-sonnet-20241022",
	},
	WithAPIKey: types.ClaudeCodeConfig{
		APIKey: getTestAPIKey(), // Test fixture, not a real API key
		Model:  "claude-3-5-sonnet-20241022",
	},
	WithSession: types.ClaudeCodeConfig{
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		Model:     "claude-3-5-sonnet-20241022",
	},
	Custom: types.ClaudeCodeConfig{
		Model:            "claude-3-opus-20240229",
		MaxTokens:        4096,
		Temperature:      0.7,
		WorkingDirectory: "/tmp/test",
		Environment: map[string]string{
			"TEST_ENV": "true",
		},
		Timeout: 60 * time.Second,
	},
	Invalid: types.ClaudeCodeConfig{
		Model:       "",  // Invalid: empty model
		Temperature: 2.5, // Invalid: out of range
		MaxTokens:   -1,  // Invalid: negative
	},
}

// TestTools provides sample tool definitions
var TestTools = struct {
	FileOperations []types.Tool
	WebTools       []types.Tool
	SystemTools    []types.Tool
}{
	FileOperations: []types.Tool{
		{
			Name:        "read_file",
			Description: "Read the contents of a file",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.ToolProperty{
					"path": {
						Type:        "string",
						Description: "Path to the file to read",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "write_file",
			Description: "Write content to a file",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.ToolProperty{
					"path": {
						Type:        "string",
						Description: "Path to the file to write",
					},
					"content": {
						Type:        "string",
						Description: "Content to write to the file",
					},
				},
				Required: []string{"path", "content"},
			},
		},
	},
	WebTools: []types.Tool{
		{
			Name:        "fetch_url",
			Description: "Fetch content from a URL",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.ToolProperty{
					"url": {
						Type:        "string",
						Description: "URL to fetch",
					},
				},
				Required: []string{"url"},
			},
		},
	},
	SystemTools: []types.Tool{
		{
			Name:        "run_command",
			Description: "Execute a system command",
			InputSchema: types.ToolInputSchema{
				Type: "object",
				Properties: map[string]types.ToolProperty{
					"command": {
						Type:        "string",
						Description: "Command to execute",
					},
					"args": {
						Type:        "array",
						Description: "Command arguments",
						Items: &types.ToolProperty{
							Type: "string",
						},
					},
				},
				Required: []string{"command"},
			},
		},
	},
}

// TestSessions provides sample session data
var TestSessions = struct {
	Active  map[string]types.Session
	Expired map[string]types.Session
	Invalid map[string]types.Session
}{
	Active: map[string]types.Session{
		"session-1": {
			ID:           "550e8400-e29b-41d4-a716-446655440001",
			Title:        "Test Session 1",
			CreatedAt:    time.Now().Add(-1 * time.Hour).Unix(),
			UpdatedAt:    time.Now().Unix(),
			MessageCount: 10,
		},
		"session-2": {
			ID:           "550e8400-e29b-41d4-a716-446655440002",
			Title:        "Test Session 2",
			CreatedAt:    time.Now().Add(-30 * time.Minute).Unix(),
			UpdatedAt:    time.Now().Unix(),
			MessageCount: 5,
		},
	},
	Expired: map[string]types.Session{
		"expired-1": {
			ID:           "550e8400-e29b-41d4-a716-446655440003",
			Title:        "Expired Session",
			CreatedAt:    time.Now().Add(-25 * time.Hour).Unix(),
			UpdatedAt:    time.Now().Add(-25 * time.Hour).Unix(),
			MessageCount: 2,
		},
	},
	Invalid: map[string]types.Session{
		"invalid-1": {
			ID:           "not-a-valid-uuid",
			Title:        "Invalid Session",
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
			MessageCount: 0,
		},
	},
}

// Helper functions

// generateLargeContext creates a large conversation history for testing
func generateLargeContext(turns int) []types.Message {
	messages := make([]types.Message, 0, turns*2)

	for i := 0; i < turns; i++ {
		messages = append(messages,
			types.Message{
				Role:    types.RoleUser,
				Content: generateLoremIpsum(100 + i*10),
			},
			types.Message{
				Role:    types.RoleAssistant,
				Content: generateLoremIpsum(150 + i*15),
			},
		)
	}

	return messages
}

// generateLoremIpsum generates Lorem Ipsum text of specified word count
func generateLoremIpsum(words int) string {
	loremWords := []string{
		"lorem", "ipsum", "dolor", "sit", "amet", "consectetur",
		"adipiscing", "elit", "sed", "do", "eiusmod", "tempor",
		"incididunt", "ut", "labore", "et", "dolore", "magna",
		"aliqua", "enim", "ad", "minim", "veniam", "quis",
	}

	result := make([]string, words)
	for i := 0; i < words; i++ {
		result[i] = loremWords[i%len(loremWords)]
	}

	return strings.Join(result, " ")
}

// GetTestFile returns test file content
func GetTestFile(name string) []byte {
	files := map[string][]byte{
		"hello.txt":     []byte("Hello, World!"),
		"code.go":       []byte("package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}"),
		"data.json":     []byte(`{"name": "test", "value": 42}`),
		"empty.txt":     []byte(""),
		"binary.bin":    {0x00, 0x01, 0x02, 0x03, 0xFF},
		"large.txt":     []byte(generateLoremIpsum(1000)),
		"unicode.txt":   []byte("Hello 世界 🌍 émojis"),
		"multiline.txt": []byte("Line 1\nLine 2\nLine 3\n"),
	}

	if content, ok := files[name]; ok {
		return content
	}
	return nil
}

// GetErrorScenario returns test data for error scenarios
func GetErrorScenario(scenario string) error {
	scenarios := map[string]error{
		"timeout":      fmt.Errorf("context deadline exceeded"),
		"network":      fmt.Errorf("dial tcp: connection refused"),
		"auth":         fmt.Errorf("authentication failed: invalid API key"),
		"rate_limit":   fmt.Errorf("rate limit exceeded: retry after 60s"),
		"invalid_json": fmt.Errorf("invalid character 'x' looking for beginning of value"),
		"permission":   fmt.Errorf("permission denied"),
		"not_found":    fmt.Errorf("file not found"),
	}

	if err, ok := scenarios[scenario]; ok {
		return err
	}
	return fmt.Errorf("unknown error scenario: %s", scenario)
}
