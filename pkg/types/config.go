package types

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Config contains configuration options for the Claude Code SDK client.
// It provides settings for authentication, networking, and behavior customization.
//
// Example usage:
//
//	config := &types.Config{
//		Auth: &types.APIKeyAuth{
//			APIKey: "your-api-key-here",
//		},
//		BaseURL: "https://example.com",
//		Model:   types.ModelClaude35Sonnet,
//	}
//	client := claude.NewClient(ctx, config)
type Config struct {
	// Auth provides authentication for API requests
	Auth Authenticator `json:"-"`

	// BaseURL is the base URL for the Claude Code API
	BaseURL string `json:"base_url,omitempty"`

	// Model is the default model to use for requests
	Model string `json:"model,omitempty"`

	// MaxTokens is the default maximum tokens for responses
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature is the default temperature for responses (0.0 to 1.0)
	Temperature float64 `json:"temperature,omitempty"`

	// Timeout is the default timeout for API requests
	Timeout time.Duration `json:"timeout,omitempty"`

	// RetryConfig defines retry behavior for failed requests
	RetryConfig *RetryConfig `json:"retry_config,omitempty"`

	// HTTPClient allows providing a custom HTTP client
	HTTPClient *http.Client `json:"-"`

	// UserAgent overrides the default user agent string
	UserAgent string `json:"user_agent,omitempty"`

	// Debug enables debug logging when true
	Debug bool `json:"debug,omitempty"`

	// RateLimiting contains rate limiting configuration
	RateLimiting *RateLimitConfig `json:"rate_limiting,omitempty"`

	// Proxy configuration for HTTP requests
	Proxy *ProxyConfig `json:"proxy,omitempty"`

	// Headers contains default headers to include with all requests
	Headers map[string]string `json:"headers,omitempty"`

	// TLSConfig contains TLS configuration options
	TLSConfig *TLSConfig `json:"tls_config,omitempty"`

	// SessionConfig contains default session configuration
	SessionConfig *SessionConfig `json:"session_config,omitempty"`
}

// ClaudeCodeConfig contains configuration options for the Claude Code CLI client.
// This configuration is used for the subprocess-based client that executes the
// claude CLI rather than making direct HTTP API calls.
//
// Example usage:
//
//	config := &types.ClaudeCodeConfig{
//		WorkingDirectory: "/path/to/project",
//		SessionID:        "my-session",
//		Model:           "claude-3-5-sonnet-20241022",
//		APIKey:          "your-api-key",
//	}
//	client, err := NewClaudeCodeClient(ctx, config)
type ClaudeCodeConfig struct {
	// WorkingDirectory is the project directory for context (defaults to current directory)
	WorkingDirectory string `json:"working_directory,omitempty"`

	// SessionID is the session identifier for conversation persistence
	SessionID string `json:"session_id,omitempty"`

	// Model is the Claude model to use (defaults to claude-3-5-sonnet-20241022)
	Model string `json:"model,omitempty"`

	// APIKey is the Anthropic API key for authentication
	APIKey string `json:"api_key,omitempty"`

	// AuthMethod specifies the authentication method to use
	// Options: "api_key" (default), "subscription"
	AuthMethod AuthType `json:"auth_method,omitempty"`

	// ClaudeCodePath is the path to the claude executable (auto-detected if not provided)
	ClaudeCodePath string `json:"claude_code_path,omitempty"`

	// MCPServers contains MCP server configurations for tool extensions
	MCPServers map[string]*MCPServerConfig `json:"mcp_servers,omitempty"`

	// Environment contains environment variables for subprocess execution
	Environment map[string]string `json:"environment,omitempty"`

	// MaxTokens is the default maximum tokens for responses
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature is the default temperature for responses (0.0 to 1.0)
	Temperature float64 `json:"temperature,omitempty"`

	// System is the default system prompt
	System string `json:"system,omitempty"`

	// Timeout is the default timeout for CLI execution
	Timeout time.Duration `json:"timeout,omitempty"`

	// Debug enables debug logging
	Debug bool `json:"debug,omitempty"`

	// TestMode bypasses Claude Code CLI requirement for testing
	// When true, the client will skip CLI validation and use mock behavior
	TestMode bool `json:"test_mode,omitempty"`

	// ClaudeExecutable is an alias for ClaudeCodePath for backward compatibility
	ClaudeExecutable string `json:"claude_executable,omitempty"`
}

// NewClaudeCodeConfig creates a new ClaudeCodeConfig with sensible defaults.
func NewClaudeCodeConfig() *ClaudeCodeConfig {
	return &ClaudeCodeConfig{
		Model:       ModelClaude35Sonnet,
		MaxTokens:   4096,
		Temperature: 0.0,
		Timeout:     30 * time.Second,
		MCPServers:  make(map[string]*MCPServerConfig),
		Environment: make(map[string]string),
	}
}

// NewClaudeCodeConfigFromEnvironment creates a new ClaudeCodeConfig from environment variables.
func NewClaudeCodeConfigFromEnvironment() *ClaudeCodeConfig {
	config := NewClaudeCodeConfig()

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}

	if workDir := os.Getenv("CLAUDE_WORKING_DIR"); workDir != "" {
		config.WorkingDirectory = workDir
	}

	if model := os.Getenv("CLAUDE_MODEL"); model != "" {
		config.Model = model
	}

	if debug := os.Getenv("CLAUDE_DEBUG"); debug == "true" || debug == "1" {
		config.Debug = true
	}

	return config
}

// MCPServerConfig defines configuration for a Model Context Protocol server.
type MCPServerConfig struct {
	// Command is the command to execute for the MCP server
	Command string `json:"command"`

	// Args are arguments to pass to the MCP server command
	Args []string `json:"args,omitempty"`

	// Environment contains environment variables for the MCP server
	Environment map[string]string `json:"environment,omitempty"`

	// WorkingDirectory is the working directory for the MCP server
	WorkingDirectory string `json:"working_directory,omitempty"`

	// Enabled indicates whether this MCP server should be used
	Enabled bool `json:"enabled"`
}

// Validate performs validation on the Claude Code configuration.
func (c *ClaudeCodeConfig) Validate() error {
	if c.WorkingDirectory != "" {
		if _, err := os.Stat(c.WorkingDirectory); os.IsNotExist(err) {
			return &ValidationError{
				Field:   "working_directory",
				Message: "working directory does not exist: " + c.WorkingDirectory,
			}
		}
	}

	if c.MaxTokens < 0 {
		return &ValidationError{
			Field:   "max_tokens",
			Message: "max_tokens cannot be negative",
		}
	}

	if c.Temperature < 0 || c.Temperature > 1 {
		return &ValidationError{
			Field:   "temperature",
			Message: "temperature must be between 0.0 and 1.0",
		}
	}

	if c.Timeout < 0 {
		return &ValidationError{
			Field:   "timeout",
			Message: "timeout cannot be negative",
		}
	}

	return nil
}

// ApplyDefaults applies default values to unset configuration options.
func (c *ClaudeCodeConfig) ApplyDefaults() {
	if c.Model == "" {
		c.Model = DefaultModel
	}

	if c.MaxTokens == 0 {
		c.MaxTokens = DefaultMaxTokens
	}

	if c.Temperature == 0 {
		c.Temperature = DefaultTemperature
	}

	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}

	if c.Environment == nil {
		c.Environment = make(map[string]string)
	}

	// Add API key from environment if not set
	if c.APIKey == "" {
		if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
			c.APIKey = apiKey
		}
	}

	// Set default auth method if not specified
	if c.AuthMethod == "" {
		if c.APIKey != "" {
			c.AuthMethod = AuthTypeAPIKey
		} else {
			// Try to detect if subscription auth is available
			if c.isSubscriptionAuthAvailable() {
				c.AuthMethod = AuthTypeSubscription
			} else {
				c.AuthMethod = AuthTypeAPIKey // Default fallback
			}
		}
	}
}

// isSubscriptionAuthAvailable checks if subscription authentication is available.
func (c *ClaudeCodeConfig) isSubscriptionAuthAvailable() bool {
	// Check if Claude CLI is available
	candidates := []string{"claude", "npx claude"}

	for _, candidate := range candidates {
		if strings.Contains(candidate, " ") {
			// For commands like "npx claude", test by running with --version
			parts := strings.Fields(candidate)
			cmd := exec.Command(parts[0], append(parts[1:], "--version")...) // #nosec G204 - using predefined safe candidates
			if err := cmd.Run(); err == nil {
				return true
			}
		} else {
			// For single commands, check if available in PATH
			if _, err := exec.LookPath(candidate); err == nil {
				return true
			}
		}
	}

	return false
}

// GetAuthenticator returns the appropriate authenticator based on the configured auth method.
func (c *ClaudeCodeConfig) GetAuthenticator() Authenticator {
	switch c.AuthMethod {
	case AuthTypeAPIKey:
		if c.APIKey != "" {
			return &APIKeyAuth{APIKey: c.APIKey}
		}
	case AuthTypeSubscription:
		return &SubscriptionAuth{}
	}

	// Fallback: try to auto-detect
	if c.APIKey != "" {
		return &APIKeyAuth{APIKey: c.APIKey}
	}

	// Default to subscription auth if available
	if c.isSubscriptionAuthAvailable() {
		return &SubscriptionAuth{}
	}

	// Final fallback to API key auth (will fail validation later if no key)
	return &APIKeyAuth{}
}

// IsUsingSubscriptionAuth returns true if the configuration is set up for subscription authentication.
func (c *ClaudeCodeConfig) IsUsingSubscriptionAuth() bool {
	return c.AuthMethod == AuthTypeSubscription
}

// IsUsingAPIKeyAuth returns true if the configuration is set up for API key authentication.
func (c *ClaudeCodeConfig) IsUsingAPIKeyAuth() bool {
	return c.AuthMethod == AuthTypeAPIKey
}

// Validate performs validation on the configuration.
func (c *Config) Validate() error {
	if c.Auth == nil {
		return &ValidationError{
			Field:   "auth",
			Message: "authentication is required",
		}
	}

	// Validate authenticator itself
	if !c.Auth.IsValid(context.Background()) {
		// Try to get a more specific error by attempting authentication
		if err := c.Auth.Authenticate(context.Background(), nil); err != nil {
			return &ValidationError{
				Field:   "auth",
				Message: "authentication validation failed: " + err.Error(),
			}
		}
	}

	if c.BaseURL != "" {
		if u, err := url.Parse(c.BaseURL); err != nil {
			return &ValidationError{
				Field:   "base_url",
				Message: "invalid base URL: " + err.Error(),
			}
		} else if !u.IsAbs() {
			return &ValidationError{
				Field:   "base_url",
				Message: "base URL must be absolute (include scheme like https://)",
			}
		}
	}

	if c.MaxTokens < 0 {
		return &ValidationError{
			Field:   "max_tokens",
			Message: "max_tokens cannot be negative",
		}
	}

	if c.Temperature < 0 || c.Temperature > 1 {
		return &ValidationError{
			Field:   "temperature",
			Message: "temperature must be between 0.0 and 1.0",
		}
	}

	if c.Timeout < 0 {
		return &ValidationError{
			Field:   "timeout",
			Message: "timeout cannot be negative",
		}
	}

	return nil
}

// ApplyDefaults applies default values to unset configuration options.
func (c *Config) ApplyDefaults() {
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}

	if c.Model == "" {
		c.Model = DefaultModel
	}

	if c.MaxTokens == 0 {
		c.MaxTokens = DefaultMaxTokens
	}

	if c.Temperature == 0 {
		c.Temperature = DefaultTemperature
	}

	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}

	if c.RetryConfig == nil {
		c.RetryConfig = DefaultRetryConfig()
	}

	if c.UserAgent == "" {
		c.UserAgent = DefaultUserAgent
	}

	if c.Headers == nil {
		c.Headers = make(map[string]string)
	}

	// Set default headers
	if c.Headers["Content-Type"] == "" {
		c.Headers["Content-Type"] = "application/json"
	}
}

// Clone creates a deep copy of the configuration.
func (c *Config) Clone() *Config {
	clone := &Config{
		Auth:         c.Auth,
		BaseURL:      c.BaseURL,
		Model:        c.Model,
		MaxTokens:    c.MaxTokens,
		Temperature:  c.Temperature,
		Timeout:      c.Timeout,
		HTTPClient:   c.HTTPClient,
		UserAgent:    c.UserAgent,
		Debug:        c.Debug,
		RateLimiting: c.RateLimiting,
		Proxy:        c.Proxy,
		TLSConfig:    c.TLSConfig,
	}

	// Deep copy retry config
	if c.RetryConfig != nil {
		retryConfig := *c.RetryConfig
		clone.RetryConfig = &retryConfig
	}

	// Deep copy headers
	if c.Headers != nil {
		clone.Headers = make(map[string]string)
		for k, v := range c.Headers {
			clone.Headers[k] = v
		}
	}

	// Deep copy session config
	if c.SessionConfig != nil {
		sessionConfig := *c.SessionConfig
		clone.SessionConfig = &sessionConfig
	}

	return clone
}

// RateLimitConfig defines rate limiting behavior.
type RateLimitConfig struct {
	// Enabled indicates whether rate limiting is active
	Enabled bool `json:"enabled"`

	// RequestsPerMinute is the maximum requests per minute
	RequestsPerMinute int `json:"requests_per_minute"`

	// TokensPerMinute is the maximum tokens per minute
	TokensPerMinute int `json:"tokens_per_minute"`

	// BurstSize allows bursts up to this size
	BurstSize int `json:"burst_size"`

	// WaitOnLimit indicates whether to wait when limits are hit
	WaitOnLimit bool `json:"wait_on_limit"`
}

// ProxyConfig defines HTTP proxy settings.
type ProxyConfig struct {
	// URL is the proxy server URL
	URL string `json:"url"`

	// Username for proxy authentication
	Username string `json:"username,omitempty"`

	// Password for proxy authentication
	Password string `json:"password,omitempty"`

	// NoProxy contains hosts that should bypass the proxy
	NoProxy []string `json:"no_proxy,omitempty"`
}

// TLSConfig defines TLS/SSL configuration options.
type TLSConfig struct {
	// InsecureSkipVerify disables certificate verification (not recommended for production)
	InsecureSkipVerify bool `json:"insecure_skip_verify,omitempty"`

	// CertFile is the path to the client certificate file
	CertFile string `json:"cert_file,omitempty"`

	// KeyFile is the path to the client private key file
	KeyFile string `json:"key_file,omitempty"`

	// CAFile is the path to the CA certificate file
	CAFile string `json:"ca_file,omitempty"`

	// ServerName is used to verify the hostname on the returned certificates
	ServerName string `json:"server_name,omitempty"`

	// MinVersion is the minimum TLS version to support
	MinVersion uint16 `json:"min_version,omitempty"`

	// MaxVersion is the maximum TLS version to support
	MaxVersion uint16 `json:"max_version,omitempty"`
}

// Constants for default configuration values
const (
	// DefaultBaseURL is the default Claude Code API base URL
	DefaultBaseURL = "https://example.com"

	// DefaultModel is the default Claude model to use
	DefaultModel = ModelClaude35Sonnet

	// DefaultMaxTokens is the default maximum tokens for responses
	DefaultMaxTokens = 4000

	// DefaultTemperature is the default temperature for responses
	DefaultTemperature = 0.7

	// DefaultTimeout is the default timeout for API requests
	DefaultTimeout = 60 * time.Second

	// DefaultUserAgent is the default user agent string
	DefaultUserAgent = "claude-code-go-sdk/1.0.0"

	// APIVersion is the current API version
	APIVersion = "2023-06-01"
)

// Model constants for available Claude models
const (
	// ModelClaude35Sonnet is the Claude 3.5 Sonnet model
	ModelClaude35Sonnet = "claude-3-5-sonnet-20241022"

	// ModelClaude3Opus is the Claude 3 Opus model
	ModelClaude3Opus = "claude-3-opus-20240229"

	// ModelClaude3Sonnet is the Claude 3 Sonnet model
	ModelClaude3Sonnet = "claude-3-sonnet-20240229"

	// ModelClaude3Haiku is the Claude 3 Haiku model
	ModelClaude3Haiku = "claude-3-haiku-20240307"
)

// OptionFunc is a function type for configuring clients using the functional options pattern.
type OptionFunc func(*Config) error

// WithAuth sets the authentication method for the client.
func WithAuth(auth Authenticator) OptionFunc {
	return func(c *Config) error {
		c.Auth = auth
		return nil
	}
}

// WithAPIKey sets API key authentication for the client.
func WithAPIKey(apiKey string) OptionFunc {
	return func(c *Config) error {
		c.Auth = &APIKeyAuth{APIKey: apiKey}
		return nil
	}
}

// WithModel sets the default model for the client.
func WithModel(model string) OptionFunc {
	return func(c *Config) error {
		c.Model = model
		return nil
	}
}

// WithBaseURL sets the base URL for the API.
func WithBaseURL(baseURL string) OptionFunc {
	return func(c *Config) error {
		if baseURL == "" {
			c.BaseURL = baseURL
			return nil
		}
		if u, err := url.Parse(baseURL); err != nil {
			return &ValidationError{
				Field:   "base_url",
				Message: "invalid base URL: " + err.Error(),
			}
		} else if !u.IsAbs() {
			return &ValidationError{
				Field:   "base_url",
				Message: "base URL must be absolute (include scheme like https://)",
			}
		}
		c.BaseURL = baseURL
		return nil
	}
}

// WithMaxTokens sets the default maximum tokens for responses.
func WithMaxTokens(maxTokens int) OptionFunc {
	return func(c *Config) error {
		if maxTokens <= 0 {
			return &ValidationError{
				Field:   "max_tokens",
				Message: "max_tokens must be greater than 0",
			}
		}
		c.MaxTokens = maxTokens
		return nil
	}
}

// WithTemperature sets the default temperature for responses.
func WithTemperature(temperature float64) OptionFunc {
	return func(c *Config) error {
		if temperature < 0 || temperature > 1 {
			return &ValidationError{
				Field:   "temperature",
				Message: "temperature must be between 0.0 and 1.0",
			}
		}
		c.Temperature = temperature
		return nil
	}
}

// WithTimeout sets the request timeout.
func WithTimeout(timeout time.Duration) OptionFunc {
	return func(c *Config) error {
		if timeout < 0 {
			return &ValidationError{
				Field:   "timeout",
				Message: "timeout cannot be negative",
			}
		}
		c.Timeout = timeout
		return nil
	}
}

// WithRetryConfig sets the retry configuration.
func WithRetryConfig(retryConfig *RetryConfig) OptionFunc {
	return func(c *Config) error {
		c.RetryConfig = retryConfig
		return nil
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) OptionFunc {
	return func(c *Config) error {
		c.HTTPClient = client
		return nil
	}
}

// WithDebug enables or disables debug logging.
func WithDebug(debug bool) OptionFunc {
	return func(c *Config) error {
		c.Debug = debug
		return nil
	}
}

// WithHeader adds a default header to all requests.
func WithHeader(key, value string) OptionFunc {
	return func(c *Config) error {
		if c.Headers == nil {
			c.Headers = make(map[string]string)
		}
		c.Headers[key] = value
		return nil
	}
}

// WithRateLimiting configures rate limiting.
func WithRateLimiting(config *RateLimitConfig) OptionFunc {
	return func(c *Config) error {
		c.RateLimiting = config
		return nil
	}
}

// WithProxy configures HTTP proxy settings.
func WithProxy(config *ProxyConfig) OptionFunc {
	return func(c *Config) error {
		c.Proxy = config
		return nil
	}
}

// WithProxyURL configures HTTP proxy with just a URL.
func WithProxyURL(proxyURL string) OptionFunc {
	return func(c *Config) error {
		if proxyURL == "" {
			return nil // Allow clearing proxy
		}
		if _, err := url.Parse(proxyURL); err != nil {
			return &ValidationError{
				Field:   "proxy_url",
				Message: "invalid proxy URL: " + err.Error(),
			}
		}
		if c.Proxy == nil {
			c.Proxy = &ProxyConfig{}
		}
		c.Proxy.URL = proxyURL
		return nil
	}
}

// WithProxyAuth configures HTTP proxy with authentication.
func WithProxyAuth(proxyURL, username, password string) OptionFunc {
	return func(c *Config) error {
		if proxyURL == "" {
			return &ValidationError{
				Field:   "proxy_url",
				Message: "proxy URL cannot be empty when setting proxy auth",
			}
		}
		if _, err := url.Parse(proxyURL); err != nil {
			return &ValidationError{
				Field:   "proxy_url",
				Message: "invalid proxy URL: " + err.Error(),
			}
		}
		c.Proxy = &ProxyConfig{
			URL:      proxyURL,
			Username: username,
			Password: password,
		}
		return nil
	}
}

// WithTLS configures TLS settings.
func WithTLS(config *TLSConfig) OptionFunc {
	return func(c *Config) error {
		c.TLSConfig = config
		return nil
	}
}

// WithTLSInsecure configures TLS to skip certificate verification.
// WARNING: This should only be used in development/testing environments.
func WithTLSInsecure(insecure bool) OptionFunc {
	return func(c *Config) error {
		if c.TLSConfig == nil {
			c.TLSConfig = &TLSConfig{}
		}
		c.TLSConfig.InsecureSkipVerify = insecure
		return nil
	}
}

// WithTLSCerts configures TLS client certificates.
func WithTLSCerts(certFile, keyFile, caFile string) OptionFunc {
	return func(c *Config) error {
		if certFile == "" && keyFile == "" && caFile == "" {
			return nil // Allow clearing TLS config
		}
		if c.TLSConfig == nil {
			c.TLSConfig = &TLSConfig{}
		}
		c.TLSConfig.CertFile = certFile
		c.TLSConfig.KeyFile = keyFile
		c.TLSConfig.CAFile = caFile
		return nil
	}
}

// WithSessionConfig configures default session settings.
func WithSessionConfig(config *SessionConfig) OptionFunc {
	return func(c *Config) error {
		c.SessionConfig = config
		return nil
	}
}

// WithUserAgent sets a custom user agent string.
func WithUserAgent(userAgent string) OptionFunc {
	return func(c *Config) error {
		c.UserAgent = userAgent
		return nil
	}
}

// WithEnvironment configures the client for a specific environment.
// This is a convenience function that sets appropriate defaults for different environments.
func WithEnvironment(env Environment) OptionFunc {
	return func(c *Config) error {
		if !env.IsValid() {
			return &ValidationError{
				Field:   "environment",
				Message: "invalid environment: " + string(env),
			}
		}

		switch env {
		case EnvironmentProduction:
			// Production: secure defaults, longer timeouts
			c.Debug = false
			if c.Timeout == 0 {
				c.Timeout = 60 * time.Second
			}
			if c.RetryConfig == nil {
				c.RetryConfig = &RetryConfig{
					MaxRetries:   3,
					InitialDelay: 2 * time.Second,
					MaxDelay:     30 * time.Second,
					Multiplier:   2.0,
				}
			}

		case EnvironmentStaging:
			// Staging: balanced settings for testing
			c.Debug = false
			if c.Timeout == 0 {
				c.Timeout = 45 * time.Second
			}
			if c.RetryConfig == nil {
				c.RetryConfig = &RetryConfig{
					MaxRetries:   2,
					InitialDelay: 1 * time.Second,
					MaxDelay:     15 * time.Second,
					Multiplier:   1.5,
				}
			}

		case EnvironmentDevelopment:
			// Development: faster feedback, debug enabled
			c.Debug = true
			if c.Timeout == 0 {
				c.Timeout = 30 * time.Second
			}
			if c.RetryConfig == nil {
				c.RetryConfig = &RetryConfig{
					MaxRetries:   1,
					InitialDelay: 500 * time.Millisecond,
					MaxDelay:     5 * time.Second,
					Multiplier:   1.0,
				}
			}

		case EnvironmentTesting:
			// Testing: minimal delays, no retries
			c.Debug = false
			if c.Timeout == 0 {
				c.Timeout = 10 * time.Second
			}
			if c.RetryConfig == nil {
				c.RetryConfig = &RetryConfig{
					MaxRetries:   0, // No retries in tests
					InitialDelay: 0,
					MaxDelay:     0,
					Multiplier:   1.0,
				}
			}
		}

		return nil
	}
}

// WithDefaults applies the default configuration for production use.
// This is equivalent to calling ApplyDefaults() but in the functional options style.
func WithDefaults() OptionFunc {
	return func(c *Config) error {
		c.ApplyDefaults()
		return nil
	}
}

// WithHeaders sets multiple headers at once.
func WithHeaders(headers map[string]string) OptionFunc {
	return func(c *Config) error {
		if c.Headers == nil {
			c.Headers = make(map[string]string)
		}
		for k, v := range headers {
			c.Headers[k] = v
		}
		return nil
	}
}

// NewConfig creates a new configuration with the provided options.
func NewConfig(options ...OptionFunc) (*Config, error) {
	config := &Config{}

	for _, option := range options {
		if err := option(config); err != nil {
			return nil, err
		}
	}

	config.ApplyDefaults()

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Environment represents different deployment environments.
type Environment string

const (
	// EnvironmentProduction is the production environment
	EnvironmentProduction Environment = "production"

	// EnvironmentStaging is the staging environment
	EnvironmentStaging Environment = "staging"

	// EnvironmentDevelopment is the development environment
	EnvironmentDevelopment Environment = "development"

	// EnvironmentTesting is the testing environment
	EnvironmentTesting Environment = "testing"
)

// IsValid checks if the environment is valid.
func (e Environment) IsValid() bool {
	switch e {
	case EnvironmentProduction, EnvironmentStaging, EnvironmentDevelopment, EnvironmentTesting:
		return true
	default:
		return false
	}
}

// LogLevel represents different logging levels.
type LogLevel string

const (
	// LogLevelDebug enables debug-level logging
	LogLevelDebug LogLevel = "debug"

	// LogLevelInfo enables info-level logging
	LogLevelInfo LogLevel = "info"

	// LogLevelWarn enables warning-level logging
	LogLevelWarn LogLevel = "warn"

	// LogLevelError enables error-level logging
	LogLevelError LogLevel = "error"

	// LogLevelOff disables logging
	LogLevelOff LogLevel = "off"
)

// IsValid checks if the log level is valid.
func (l LogLevel) IsValid() bool {
	switch l {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError, LogLevelOff:
		return true
	default:
		return false
	}
}

// Environment variable names for configuration
const (
	// EnvAPIKey is the environment variable for the API key
	EnvAPIKey = "CLAUDE_API_KEY" // #nosec G101 - This is just the environment variable name, not a credential

	// EnvBaseURL is the environment variable for the base URL
	EnvBaseURL = "CLAUDE_BASE_URL"

	// EnvModel is the environment variable for the default model
	EnvModel = "CLAUDE_MODEL"

	// EnvMaxTokens is the environment variable for max tokens
	EnvMaxTokens = "CLAUDE_MAX_TOKENS"

	// EnvTemperature is the environment variable for temperature
	EnvTemperature = "CLAUDE_TEMPERATURE"

	// EnvTimeout is the environment variable for request timeout
	EnvTimeout = "CLAUDE_TIMEOUT"

	// EnvDebug is the environment variable for debug mode
	EnvDebug = "CLAUDE_DEBUG"

	// EnvLogLevel is the environment variable for log level
	EnvLogLevel = "CLAUDE_LOG_LEVEL"

	// EnvEnvironment is the environment variable for deployment environment
	EnvEnvironment = "CLAUDE_ENVIRONMENT"

	// EnvUserAgent is the environment variable for custom user agent
	EnvUserAgent = "CLAUDE_USER_AGENT"

	// EnvRetryAttempts is the environment variable for retry attempts
	EnvRetryAttempts = "CLAUDE_RETRY_ATTEMPTS"

	// EnvRetryDelay is the environment variable for retry delay
	EnvRetryDelay = "CLAUDE_RETRY_DELAY"

	// EnvProxyURL is the environment variable for proxy URL
	EnvProxyURL = "CLAUDE_PROXY_URL"

	// EnvTLSInsecure is the environment variable for TLS insecure skip verify
	EnvTLSInsecure = "CLAUDE_TLS_INSECURE"
)

// LoadFromEnvironment loads configuration values from environment variables.
// This function populates a Config struct with values from environment variables
// using the standard CLAUDE_* prefix. Environment variables take precedence over
// default values but are overridden by explicit configuration.
//
// Supported environment variables:
//   - CLAUDE_API_KEY: API key for authentication
//   - CLAUDE_BASE_URL: Base URL for the API
//   - CLAUDE_MODEL: Default model to use
//   - CLAUDE_MAX_TOKENS: Maximum tokens for responses
//   - CLAUDE_TEMPERATURE: Temperature for responses (0.0-1.0)
//   - CLAUDE_TIMEOUT: Request timeout in seconds
//   - CLAUDE_DEBUG: Enable debug mode (true/false)
//   - CLAUDE_LOG_LEVEL: Logging level (debug/info/warn/error/off)
//   - CLAUDE_ENVIRONMENT: Deployment environment
//   - CLAUDE_USER_AGENT: Custom user agent string
//   - CLAUDE_RETRY_ATTEMPTS: Number of retry attempts
//   - CLAUDE_RETRY_DELAY: Delay between retries in seconds
//   - CLAUDE_PROXY_URL: HTTP proxy URL
//   - CLAUDE_TLS_INSECURE: Skip TLS verification (true/false)
//
// Example usage:
//
//	// Load from environment with defaults
//	config := &Config{}
//	config.LoadFromEnvironment()
//	config.ApplyDefaults()
//
//	// Or use the convenience function
//	config, err := NewConfigFromEnvironment()
func (c *Config) LoadFromEnvironment() error {
	var errs []error

	// Load API key
	if apiKey := os.Getenv(EnvAPIKey); apiKey != "" {
		c.Auth = &APIKeyAuth{APIKey: apiKey}
	}

	// Load base URL
	if baseURL := os.Getenv(EnvBaseURL); baseURL != "" {
		if _, err := url.Parse(baseURL); err != nil {
			errs = append(errs, &ValidationError{
				Field:   "base_url",
				Message: fmt.Sprintf("invalid base URL from %s: %v", EnvBaseURL, err),
			})
		} else {
			c.BaseURL = baseURL
		}
	}

	// Load model
	if model := os.Getenv(EnvModel); model != "" {
		c.Model = model
	}

	// Load max tokens
	if maxTokensStr := os.Getenv(EnvMaxTokens); maxTokensStr != "" {
		if maxTokens, err := strconv.Atoi(maxTokensStr); err != nil {
			errs = append(errs, &ValidationError{
				Field:   "max_tokens",
				Message: fmt.Sprintf("invalid max_tokens from %s: %v", EnvMaxTokens, err),
			})
		} else if maxTokens <= 0 {
			errs = append(errs, &ValidationError{
				Field:   "max_tokens",
				Message: fmt.Sprintf("max_tokens from %s must be greater than 0", EnvMaxTokens),
			})
		} else {
			c.MaxTokens = maxTokens
		}
	}

	// Load temperature
	if tempStr := os.Getenv(EnvTemperature); tempStr != "" {
		if temp, err := strconv.ParseFloat(tempStr, 64); err != nil {
			errs = append(errs, &ValidationError{
				Field:   "temperature",
				Message: fmt.Sprintf("invalid temperature from %s: %v", EnvTemperature, err),
			})
		} else if temp < 0 || temp > 1 {
			errs = append(errs, &ValidationError{
				Field:   "temperature",
				Message: fmt.Sprintf("temperature from %s must be between 0.0 and 1.0", EnvTemperature),
			})
		} else {
			c.Temperature = temp
		}
	}

	// Load timeout
	if timeoutStr := os.Getenv(EnvTimeout); timeoutStr != "" {
		if timeoutSecs, err := strconv.Atoi(timeoutStr); err != nil {
			errs = append(errs, &ValidationError{
				Field:   "timeout",
				Message: fmt.Sprintf("invalid timeout from %s: %v", EnvTimeout, err),
			})
		} else if timeoutSecs < 0 {
			errs = append(errs, &ValidationError{
				Field:   "timeout",
				Message: fmt.Sprintf("timeout from %s cannot be negative", EnvTimeout),
			})
		} else {
			c.Timeout = time.Duration(timeoutSecs) * time.Second
		}
	}

	// Load debug flag
	if debugStr := os.Getenv(EnvDebug); debugStr != "" {
		if debug, err := strconv.ParseBool(debugStr); err != nil {
			errs = append(errs, &ValidationError{
				Field:   "debug",
				Message: fmt.Sprintf("invalid debug flag from %s: %v", EnvDebug, err),
			})
		} else {
			c.Debug = debug
		}
	}

	// Load user agent
	if userAgent := os.Getenv(EnvUserAgent); userAgent != "" {
		c.UserAgent = userAgent
	}

	// Load retry configuration
	retryConfig := &RetryConfig{}
	hasRetryConfig := false

	if attemptsStr := os.Getenv(EnvRetryAttempts); attemptsStr != "" {
		if attempts, err := strconv.Atoi(attemptsStr); err != nil {
			errs = append(errs, &ValidationError{
				Field:   "retry_attempts",
				Message: fmt.Sprintf("invalid retry attempts from %s: %v", EnvRetryAttempts, err),
			})
		} else if attempts < 0 {
			errs = append(errs, &ValidationError{
				Field:   "retry_attempts",
				Message: fmt.Sprintf("retry attempts from %s cannot be negative", EnvRetryAttempts),
			})
		} else {
			retryConfig.MaxRetries = attempts
			hasRetryConfig = true
		}
	}

	if delayStr := os.Getenv(EnvRetryDelay); delayStr != "" {
		if delaySecs, err := strconv.Atoi(delayStr); err != nil {
			errs = append(errs, &ValidationError{
				Field:   "retry_delay",
				Message: fmt.Sprintf("invalid retry delay from %s: %v", EnvRetryDelay, err),
			})
		} else if delaySecs < 0 {
			errs = append(errs, &ValidationError{
				Field:   "retry_delay",
				Message: fmt.Sprintf("retry delay from %s cannot be negative", EnvRetryDelay),
			})
		} else {
			retryConfig.InitialDelay = time.Duration(delaySecs) * time.Second
			hasRetryConfig = true
		}
	}

	if hasRetryConfig {
		if c.RetryConfig == nil {
			c.RetryConfig = retryConfig
		} else {
			// Merge with existing retry config
			if retryConfig.MaxRetries > 0 {
				c.RetryConfig.MaxRetries = retryConfig.MaxRetries
			}
			if retryConfig.InitialDelay > 0 {
				c.RetryConfig.InitialDelay = retryConfig.InitialDelay
			}
		}
	}

	// Load proxy configuration
	if proxyURL := os.Getenv(EnvProxyURL); proxyURL != "" {
		if _, err := url.Parse(proxyURL); err != nil {
			errs = append(errs, &ValidationError{
				Field:   "proxy_url",
				Message: fmt.Sprintf("invalid proxy URL from %s: %v", EnvProxyURL, err),
			})
		} else {
			if c.Proxy == nil {
				c.Proxy = &ProxyConfig{}
			}
			c.Proxy.URL = proxyURL
		}
	}

	// Load TLS configuration
	if tlsInsecureStr := os.Getenv(EnvTLSInsecure); tlsInsecureStr != "" {
		if tlsInsecure, err := strconv.ParseBool(tlsInsecureStr); err != nil {
			errs = append(errs, &ValidationError{
				Field:   "tls_insecure",
				Message: fmt.Sprintf("invalid TLS insecure flag from %s: %v", EnvTLSInsecure, err),
			})
		} else {
			if c.TLSConfig == nil {
				c.TLSConfig = &TLSConfig{}
			}
			c.TLSConfig.InsecureSkipVerify = tlsInsecure
		}
	}

	// Return combined errors if any
	if len(errs) > 0 {
		return &ValidationError{
			Field:   "environment",
			Message: fmt.Sprintf("environment variable validation failed: %v", errs),
		}
	}

	return nil
}

// NewConfigFromEnvironment creates a new configuration loading values from environment variables
// and applying defaults. This is a convenience function that combines environment loading,
// default application, and validation.
//
// Example usage:
//
//	config, err := NewConfigFromEnvironment()
//	if err != nil {
//		log.Fatal("Configuration error:", err)
//	}
//	client := claude.NewClient(ctx, config)
func NewConfigFromEnvironment(options ...OptionFunc) (*Config, error) {
	config := &Config{}

	// Load from environment first
	if err := config.LoadFromEnvironment(); err != nil {
		return nil, err
	}

	// Apply any additional options
	for _, option := range options {
		if err := option(config); err != nil {
			return nil, err
		}
	}

	// Apply defaults for any unset values
	config.ApplyDefaults()

	// Validate the final configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// ConfigBuilder provides a fluent interface for building configurations with validation.
// This builder pattern allows for step-by-step configuration with immediate validation
// and provides better error reporting than functional options alone.
//
// Example usage:
//
//	config, err := NewConfigBuilder().
//		WithAPIKey("sk-...").
//		WithModel(ModelClaude35Sonnet).
//		WithTimeout(30 * time.Second).
//		LoadFromEnvironment().
//		Build()
type ConfigBuilder struct {
	config *Config
	errors []error
}

// NewConfigBuilder creates a new configuration builder.
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: &Config{},
		errors: make([]error, 0),
	}
}

// WithAPIKey sets the API key for authentication.
func (b *ConfigBuilder) WithAPIKey(apiKey string) *ConfigBuilder {
	if apiKey == "" {
		b.errors = append(b.errors, &ValidationError{
			Field:   "api_key",
			Message: "API key cannot be empty",
		})
		return b
	}
	b.config.Auth = &APIKeyAuth{APIKey: apiKey}
	return b
}

// WithAuth sets the authentication method.
func (b *ConfigBuilder) WithAuth(auth Authenticator) *ConfigBuilder {
	if auth == nil {
		b.errors = append(b.errors, &ValidationError{
			Field:   "auth",
			Message: "authenticator cannot be nil",
		})
		return b
	}
	b.config.Auth = auth
	return b
}

// WithModel sets the default model.
func (b *ConfigBuilder) WithModel(model string) *ConfigBuilder {
	if model == "" {
		b.errors = append(b.errors, &ValidationError{
			Field:   "model",
			Message: "model cannot be empty",
		})
		return b
	}
	b.config.Model = model
	return b
}

// WithBaseURL sets the base URL with validation.
func (b *ConfigBuilder) WithBaseURL(baseURL string) *ConfigBuilder {
	if baseURL == "" {
		b.errors = append(b.errors, &ValidationError{
			Field:   "base_url",
			Message: "base URL cannot be empty",
		})
		return b
	}
	if _, err := url.Parse(baseURL); err != nil {
		b.errors = append(b.errors, &ValidationError{
			Field:   "base_url",
			Message: "invalid base URL: " + err.Error(),
		})
		return b
	}
	b.config.BaseURL = baseURL
	return b
}

// WithTimeout sets the request timeout with validation.
func (b *ConfigBuilder) WithTimeout(timeout time.Duration) *ConfigBuilder {
	if timeout < 0 {
		b.errors = append(b.errors, &ValidationError{
			Field:   "timeout",
			Message: "timeout cannot be negative",
		})
		return b
	}
	b.config.Timeout = timeout
	return b
}

// WithTemperature sets the temperature with validation.
func (b *ConfigBuilder) WithTemperature(temperature float64) *ConfigBuilder {
	if temperature < 0 || temperature > 1 {
		b.errors = append(b.errors, &ValidationError{
			Field:   "temperature",
			Message: "temperature must be between 0.0 and 1.0",
		})
		return b
	}
	b.config.Temperature = temperature
	return b
}

// WithMaxTokens sets the maximum tokens with validation.
func (b *ConfigBuilder) WithMaxTokens(maxTokens int) *ConfigBuilder {
	if maxTokens <= 0 {
		b.errors = append(b.errors, &ValidationError{
			Field:   "max_tokens",
			Message: "max_tokens must be greater than 0",
		})
		return b
	}
	b.config.MaxTokens = maxTokens
	return b
}

// WithDebug enables or disables debug mode.
func (b *ConfigBuilder) WithDebug(debug bool) *ConfigBuilder {
	b.config.Debug = debug
	return b
}

// WithRetryConfig sets the retry configuration.
func (b *ConfigBuilder) WithRetryConfig(retryConfig *RetryConfig) *ConfigBuilder {
	if retryConfig == nil {
		b.errors = append(b.errors, &ValidationError{
			Field:   "retry_config",
			Message: "retry config cannot be nil",
		})
		return b
	}
	b.config.RetryConfig = retryConfig
	return b
}

// LoadFromEnvironment loads values from environment variables.
// This can be called at any point in the builder chain.
func (b *ConfigBuilder) LoadFromEnvironment() *ConfigBuilder {
	if err := b.config.LoadFromEnvironment(); err != nil {
		b.errors = append(b.errors, err)
	}
	return b
}

// ApplyDefaults applies default values to unset configuration options.
func (b *ConfigBuilder) ApplyDefaults() *ConfigBuilder {
	b.config.ApplyDefaults()
	return b
}

// Build creates the final configuration with validation.
// Returns an error if any validation errors occurred during building.
func (b *ConfigBuilder) Build() (*Config, error) {
	// Check for any errors accumulated during building
	if len(b.errors) > 0 {
		return nil, &ValidationError{
			Field:   "config_builder",
			Message: fmt.Sprintf("configuration build failed: %v", b.errors),
		}
	}

	// Apply defaults if not already applied
	b.config.ApplyDefaults()

	// Final validation
	if err := b.config.Validate(); err != nil {
		return nil, err
	}

	return b.config, nil
}

// Redact returns a copy of the configuration with sensitive values redacted for logging.
// This method is essential for security as it prevents accidental logging of sensitive
// information like API keys, passwords, and tokens.
//
// Redacted fields:
//   - Auth (replaced with type information only)
//   - Proxy passwords
//   - TLS certificate/key file contents (paths are preserved)
//   - Any custom headers that might contain sensitive data
//
// Example usage:
//
//	// Safe to log
//	log.Printf("Config: %+v", config.Redact())
//
//	// NEVER do this (logs sensitive data)
//	log.Printf("Config: %+v", config)
func (c *Config) Redact() *Config {
	if c == nil {
		return nil
	}

	// Create a shallow copy
	redacted := *c

	// Redact authentication
	if c.Auth != nil {
		redacted.Auth = &RedactedAuth{AuthType: c.Auth.Type()}
	}

	// Redact proxy password
	if c.Proxy != nil {
		proxy := *c.Proxy
		if proxy.Password != "" {
			proxy.Password = "[REDACTED]"
		}
		redacted.Proxy = &proxy
	}

	// Redact potentially sensitive headers
	if c.Headers != nil {
		headers := make(map[string]string)
		for k, v := range c.Headers {
			// Redact headers that commonly contain sensitive data
			lowerKey := strings.ToLower(k)
			if strings.Contains(lowerKey, "auth") ||
				strings.Contains(lowerKey, "token") ||
				strings.Contains(lowerKey, "key") ||
				strings.Contains(lowerKey, "secret") ||
				strings.Contains(lowerKey, "password") {
				headers[k] = "[REDACTED]"
			} else {
				headers[k] = v
			}
		}
		redacted.Headers = headers
	}

	return &redacted
}

// RedactedAuth is a placeholder authenticator used for logging redacted configurations.
// It implements the Authenticator interface but only provides type information.
type RedactedAuth struct {
	AuthType AuthType
}

// Authenticate always returns an error since this is only for logging.
func (r *RedactedAuth) Authenticate(ctx context.Context, req *http.Request) error {
	return fmt.Errorf("redacted authenticator cannot be used for authentication")
}

// IsValid always returns false since this is only for logging.
func (r *RedactedAuth) IsValid(ctx context.Context) bool {
	return false
}

// Refresh always returns an error since this is only for logging.
func (r *RedactedAuth) Refresh(ctx context.Context) error {
	return fmt.Errorf("redacted authenticator cannot be refreshed")
}

// Type returns the original authenticator type.
func (r *RedactedAuth) Type() AuthType {
	return r.AuthType
}

// String provides a safe string representation for logging.
func (r *RedactedAuth) String() string {
	return fmt.Sprintf("RedactedAuth{Type: %s}", r.AuthType)
}

// Sanitize removes or replaces potentially sensitive values in the configuration
// for safe serialization. Unlike Redact, this method modifies sensitive fields
// in place and is intended for scenarios where the configuration needs to be
// serialized (e.g., saved to disk, sent over network) without sensitive data.
//
// Sanitized fields:
//   - Auth is removed entirely
//   - Proxy passwords are cleared
//   - Sensitive headers are removed
//
// Example usage:
//
//	// Create a copy for serialization
//	safeConfig := config.Clone()
//	safeConfig.Sanitize()
//	data, _ := json.Marshal(safeConfig)
func (c *Config) Sanitize() {
	if c == nil {
		return
	}

	// Remove authentication entirely
	c.Auth = nil

	// Remove proxy credentials
	if c.Proxy != nil {
		c.Proxy.Username = ""
		c.Proxy.Password = ""
	}

	// Remove sensitive headers
	if c.Headers != nil {
		for k := range c.Headers {
			lowerKey := strings.ToLower(k)
			if strings.Contains(lowerKey, "auth") ||
				strings.Contains(lowerKey, "token") ||
				strings.Contains(lowerKey, "key") ||
				strings.Contains(lowerKey, "secret") ||
				strings.Contains(lowerKey, "password") {
				delete(c.Headers, k)
			}
		}
	}
}

// IsZeroConfig returns true if the configuration is effectively empty
// (all values are at their zero state). This is useful for determining
// whether configuration loading was successful.
func (c *Config) IsZeroConfig() bool {
	if c == nil {
		return true
	}

	return c.Auth == nil &&
		c.BaseURL == "" &&
		c.Model == "" &&
		c.MaxTokens == 0 &&
		c.Temperature == 0 &&
		c.Timeout == 0 &&
		c.RetryConfig == nil &&
		c.HTTPClient == nil &&
		c.UserAgent == "" &&
		!c.Debug &&
		c.RateLimiting == nil &&
		c.Proxy == nil &&
		len(c.Headers) == 0 &&
		c.TLSConfig == nil &&
		c.SessionConfig == nil
}

// ValidateEnvironment validates that all required environment variables are set
// and accessible. This is useful for deployment health checks and configuration
// validation in containerized environments.
//
// Returns validation errors for:
//   - Missing required environment variables
//   - Invalid values that would cause runtime errors
//   - Inaccessible file paths or resources
//
// Example usage:
//
//	if err := types.ValidateEnvironment(); err != nil {
//		log.Fatal("Environment validation failed:", err)
//	}
func ValidateEnvironment() error {
	var errs []error

	// Check required API key
	if apiKey := os.Getenv(EnvAPIKey); apiKey == "" {
		errs = append(errs, &ValidationError{
			Field:   "api_key",
			Message: fmt.Sprintf("required environment variable %s is not set", EnvAPIKey),
		})
	}

	// Validate optional but critical values
	if baseURL := os.Getenv(EnvBaseURL); baseURL != "" {
		if _, err := url.Parse(baseURL); err != nil {
			errs = append(errs, &ValidationError{
				Field:   "base_url",
				Message: fmt.Sprintf("invalid base URL in %s: %v", EnvBaseURL, err),
			})
		}
	}

	if maxTokensStr := os.Getenv(EnvMaxTokens); maxTokensStr != "" {
		if maxTokens, err := strconv.Atoi(maxTokensStr); err != nil {
			errs = append(errs, &ValidationError{
				Field:   "max_tokens",
				Message: fmt.Sprintf("invalid max_tokens in %s: %v", EnvMaxTokens, err),
			})
		} else if maxTokens <= 0 {
			errs = append(errs, &ValidationError{
				Field:   "max_tokens",
				Message: fmt.Sprintf("max_tokens in %s must be greater than 0", EnvMaxTokens),
			})
		}
	}

	if tempStr := os.Getenv(EnvTemperature); tempStr != "" {
		if temp, err := strconv.ParseFloat(tempStr, 64); err != nil {
			errs = append(errs, &ValidationError{
				Field:   "temperature",
				Message: fmt.Sprintf("invalid temperature in %s: %v", EnvTemperature, err),
			})
		} else if temp < 0 || temp > 1 {
			errs = append(errs, &ValidationError{
				Field:   "temperature",
				Message: fmt.Sprintf("temperature in %s must be between 0.0 and 1.0", EnvTemperature),
			})
		}
	}

	if timeoutStr := os.Getenv(EnvTimeout); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err != nil {
			errs = append(errs, &ValidationError{
				Field:   "timeout",
				Message: fmt.Sprintf("invalid timeout in %s: %v", EnvTimeout, err),
			})
		} else if timeout < 0 {
			errs = append(errs, &ValidationError{
				Field:   "timeout",
				Message: fmt.Sprintf("timeout in %s cannot be negative", EnvTimeout),
			})
		}
	}

	// Return combined errors if any
	if len(errs) > 0 {
		return &ValidationError{
			Field:   "environment_validation",
			Message: fmt.Sprintf("environment validation failed: %v", errs),
		}
	}

	return nil
}
