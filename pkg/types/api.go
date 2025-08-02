package types

import (
	"encoding/json"
	"time"
)

// QueryRequest represents a request to the Claude Code API.
// It contains the message content and configuration for how Claude should respond.
//
// Example usage:
//
//	request := &types.QueryRequest{
//		Model: "claude-3-5-sonnet-20241022",
//		Messages: []types.Message{
//			{
//				Role:    types.RoleUser,
//				Content: "Hello Claude, help me write a Go function",
//			},
//		},
//		MaxTokens: 1000,
//	}
type QueryRequest struct {
	// Model specifies which Claude model to use for the request
	Model string `json:"model"`

	// Messages contains the conversation history and current message
	Messages []Message `json:"messages"`

	// MaxTokens sets the maximum number of tokens in the response
	MaxTokens int `json:"max_tokens"`

	// Temperature controls randomness in the response (0.0 to 1.0)
	Temperature float64 `json:"temperature,omitempty"`

	// TopP controls nucleus sampling (0.0 to 1.0)
	TopP float64 `json:"top_p,omitempty"`

	// TopK controls top-k sampling (integer)
	TopK int `json:"top_k,omitempty"`

	// StopSequences defines sequences that will stop generation
	StopSequences []string `json:"stop_sequences,omitempty"`

	// Stream indicates whether to use streaming response
	Stream bool `json:"stream,omitempty"`

	// Tools defines available tools that Claude can use
	Tools []Tool `json:"tools,omitempty"`

	// ToolChoice controls how Claude uses tools
	ToolChoice interface{} `json:"tool_choice,omitempty"`

	// System provides system-level instructions
	System string `json:"system,omitempty"`

	// Metadata contains additional request information
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Validate performs basic validation on the query request.
func (q *QueryRequest) Validate() error {
	if q.Model == "" {
		return &ValidationError{
			Field:   "model",
			Message: "model is required",
		}
	}

	if len(q.Messages) == 0 {
		return &ValidationError{
			Field:   "messages",
			Message: "at least one message is required",
		}
	}

	if q.MaxTokens <= 0 {
		return &ValidationError{
			Field:   "max_tokens",
			Message: "max_tokens must be greater than 0",
		}
	}

	if q.Temperature < 0 || q.Temperature > 1 {
		return &ValidationError{
			Field:   "temperature",
			Message: "temperature must be between 0.0 and 1.0",
		}
	}

	// Validate messages
	for i, msg := range q.Messages {
		if !msg.Role.IsValid() {
			return &ValidationError{
				Field:   "messages[" + string(rune(i)) + "].role",
				Message: "invalid message role: " + string(msg.Role),
			}
		}
	}

	return nil
}

// QueryResponse represents a response from the Claude Code API.
// It contains Claude's response message and metadata about the request.
type QueryResponse struct {
	// ID is a unique identifier for this response
	ID string `json:"id"`

	// Type indicates the type of response (usually "message")
	Type string `json:"type"`

	// Role indicates the responder (usually "assistant")
	Role Role `json:"role"`

	// Content contains the response content blocks
	Content []ContentBlock `json:"content"`

	// Model indicates which model generated the response
	Model string `json:"model"`

	// StopReason indicates why the response ended
	StopReason string `json:"stop_reason,omitempty"`

	// StopSequence contains the stop sequence that ended generation (if applicable)
	StopSequence string `json:"stop_sequence,omitempty"`

	// Usage contains token usage information
	Usage *TokenUsage `json:"usage,omitempty"`

	// CreatedAt is when the response was generated
	CreatedAt time.Time `json:"created_at,omitempty"`

	// Metadata contains additional response information
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GetTextContent extracts all text content from the response.
func (r *QueryResponse) GetTextContent() string {
	var content string
	for _, block := range r.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}
	return content
}

// GetToolCalls extracts all tool calls from the response.
func (r *QueryResponse) GetToolCalls() []ToolCall {
	var toolCalls []ToolCall
	for _, block := range r.Content {
		if block.Type == "tool_use" {
			// Parse tool call from block data
			if data, ok := block.Data.(map[string]interface{}); ok {
				toolCall := ToolCall{
					Type: "function",
				}
				if id, ok := data["id"].(string); ok {
					toolCall.ID = id
				}
				if name, ok := data["name"].(string); ok {
					toolCall.Function.Name = name
				}
				if input, ok := data["input"]; ok {
					if inputBytes, err := json.Marshal(input); err == nil {
						toolCall.Function.Arguments = string(inputBytes)
					}
				}
				toolCalls = append(toolCalls, toolCall)
			}
		}
	}
	return toolCalls
}

// APIError represents an error response from the Claude Code API.
type APIError struct {
	// Type is the error type identifier
	Type string `json:"type"`

	// Message is a human-readable error message
	Message string `json:"message"`

	// Code is the HTTP status code
	Code int `json:"code,omitempty"`

	// Details contains additional error information
	Details map[string]interface{} `json:"details,omitempty"`

	// RequestID is the ID of the failed request
	RequestID string `json:"request_id,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return e.Message
}

// IsRateLimited returns true if the error is due to rate limiting.
func (e *APIError) IsRateLimited() bool {
	return e.Type == "rate_limit_error" || e.Code == 429
}

// IsInvalidRequest returns true if the error is due to an invalid request.
func (e *APIError) IsInvalidRequest() bool {
	return e.Type == "invalid_request_error" || e.Code == 400
}

// IsAuthenticationError returns true if the error is due to authentication issues.
func (e *APIError) IsAuthenticationError() bool {
	return e.Type == "authentication_error" || e.Code == 401
}

// IsPermissionError returns true if the error is due to permission issues.
func (e *APIError) IsPermissionError() bool {
	return e.Type == "permission_error" || e.Code == 403
}

// IsNotFound returns true if the requested resource was not found.
func (e *APIError) IsNotFound() bool {
	return e.Type == "not_found_error" || e.Code == 404
}

// IsServerError returns true if the error is a server-side error.
func (e *APIError) IsServerError() bool {
	return e.Code >= 500
}

// ValidationError represents a client-side validation error.
type ValidationError struct {
	// Field is the name of the field that failed validation
	Field string `json:"field"`

	// Message is a description of the validation failure
	Message string `json:"message"`

	// Value is the invalid value (optional)
	Value interface{} `json:"value,omitempty"`
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return "validation error on field '" + e.Field + "': " + e.Message
}

// RequestOptions contains options for customizing API requests.
type RequestOptions struct {
	// Timeout specifies the request timeout
	Timeout time.Duration `json:"timeout,omitempty"`

	// RetryConfig specifies retry behavior
	RetryConfig *RetryConfig `json:"retry_config,omitempty"`

	// Headers contains additional HTTP headers to send
	Headers map[string]string `json:"headers,omitempty"`

	// UserAgent overrides the default user agent
	UserAgent string `json:"user_agent,omitempty"`

	// TraceID can be used for request tracing
	TraceID string `json:"trace_id,omitempty"`
}

// RetryConfig defines retry behavior for failed requests.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int `json:"max_retries"`

	// InitialDelay is the initial delay before the first retry
	InitialDelay time.Duration `json:"initial_delay"`

	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration `json:"max_delay"`

	// Multiplier is the backoff multiplier for exponential backoff
	Multiplier float64 `json:"multiplier"`

	// RetryableErrors defines which error types should trigger retries
	RetryableErrors []string `json:"retryable_errors,omitempty"`
}

// DefaultRetryConfig returns a sensible default retry configuration.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		RetryableErrors: []string{
			"rate_limit_error",
			"server_error",
			"timeout_error",
		},
	}
}

// HealthCheckRequest represents a request to check API health.
type HealthCheckRequest struct {
	// IncludeModels indicates whether to include model availability
	IncludeModels bool `json:"include_models,omitempty"`

	// IncludeUsage indicates whether to include usage statistics
	IncludeUsage bool `json:"include_usage,omitempty"`
}

// HealthCheckResponse represents the response from a health check.
type HealthCheckResponse struct {
	// Status indicates the overall API status
	Status string `json:"status"`

	// Timestamp is when the health check was performed
	Timestamp time.Time `json:"timestamp"`

	// Version is the API version
	Version string `json:"version,omitempty"`

	// Models contains available models (if requested)
	Models []ModelInfo `json:"models,omitempty"`

	// Usage contains usage statistics (if requested)
	Usage *UsageInfo `json:"usage,omitempty"`

	// Details contains additional health information
	Details map[string]interface{} `json:"details,omitempty"`
}

// ModelInfo contains information about an available model.
type ModelInfo struct {
	// ID is the model identifier
	ID string `json:"id"`

	// Name is the human-readable model name
	Name string `json:"name"`

	// Description describes the model's capabilities
	Description string `json:"description,omitempty"`

	// MaxTokens is the maximum context length
	MaxTokens int `json:"max_tokens,omitempty"`

	// Available indicates if the model is currently available
	Available bool `json:"available"`

	// Pricing contains pricing information (if available)
	Pricing *ModelPricing `json:"pricing,omitempty"`
}

// ModelPricing contains pricing information for a model.
type ModelPricing struct {
	// InputCostPer1K is the cost per 1000 input tokens
	InputCostPer1K float64 `json:"input_cost_per_1k"`

	// OutputCostPer1K is the cost per 1000 output tokens
	OutputCostPer1K float64 `json:"output_cost_per_1k"`

	// Currency is the pricing currency (e.g., "USD")
	Currency string `json:"currency"`
}

// UsageInfo contains usage statistics.
type UsageInfo struct {
	// RequestCount is the number of requests made
	RequestCount int64 `json:"request_count"`

	// TokenCount is the total number of tokens used
	TokenCount int64 `json:"token_count"`

	// InputTokenCount is the number of input tokens
	InputTokenCount int64 `json:"input_token_count"`

	// OutputTokenCount is the number of output tokens
	OutputTokenCount int64 `json:"output_token_count"`

	// Period describes the time period for these statistics
	Period string `json:"period,omitempty"`

	// ResetTime is when the usage counters reset
	ResetTime *time.Time `json:"reset_time,omitempty"`
}
