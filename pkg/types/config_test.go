package types

import (
	"os"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name    string
		options []OptionFunc
		wantErr bool
	}{
		{
			name: "valid config with API key",
			options: []OptionFunc{
				WithAPIKey("test-api-key"),
				WithModel(ModelClaude35Sonnet),
			},
			wantErr: false,
		},
		{
			name: "invalid config - empty API key",
			options: []OptionFunc{
				WithAPIKey(""),
			},
			wantErr: true,
		},
		{
			name: "invalid config - invalid temperature",
			options: []OptionFunc{
				WithAPIKey("test-api-key"),
				WithTemperature(1.5),
			},
			wantErr: true,
		},
		{
			name: "invalid config - invalid base URL",
			options: []OptionFunc{
				WithAPIKey("test-api-key"),
				WithBaseURL("not-a-url"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewConfig(tt.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && config == nil {
				t.Error("NewConfig() returned nil config when expecting valid config")
			}
		})
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	// Save original environment
	originalEnv := map[string]string{
		EnvAPIKey:      os.Getenv(EnvAPIKey),
		EnvBaseURL:     os.Getenv(EnvBaseURL),
		EnvModel:       os.Getenv(EnvModel),
		EnvMaxTokens:   os.Getenv(EnvMaxTokens),
		EnvTemperature: os.Getenv(EnvTemperature),
		EnvTimeout:     os.Getenv(EnvTimeout),
		EnvDebug:       os.Getenv(EnvDebug),
	}

	// Clean environment for testing
	defer func() {
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	// Test loading from environment
	os.Setenv(EnvAPIKey, "env-api-key")
	os.Setenv(EnvBaseURL, "https://api.example.com")
	os.Setenv(EnvModel, ModelClaude3Haiku)
	os.Setenv(EnvMaxTokens, "2000")
	os.Setenv(EnvTemperature, "0.5")
	os.Setenv(EnvTimeout, "30")
	os.Setenv(EnvDebug, "true")

	config := &Config{}
	err := config.LoadFromEnvironment()
	if err != nil {
		t.Fatalf("LoadFromEnvironment() error = %v", err)
	}

	// Verify values were loaded
	if config.Auth == nil {
		t.Error("Expected Auth to be set from environment")
	}
	if apiKeyAuth, ok := config.Auth.(*APIKeyAuth); !ok || apiKeyAuth.APIKey != "env-api-key" {
		t.Error("Expected APIKeyAuth with correct API key")
	}
	if config.BaseURL != "https://api.example.com" {
		t.Errorf("Expected BaseURL to be 'https://api.example.com', got %s", config.BaseURL)
	}
	if config.Model != ModelClaude3Haiku {
		t.Errorf("Expected Model to be %s, got %s", ModelClaude3Haiku, config.Model)
	}
	if config.MaxTokens != 2000 {
		t.Errorf("Expected MaxTokens to be 2000, got %d", config.MaxTokens)
	}
	if config.Temperature != 0.5 {
		t.Errorf("Expected Temperature to be 0.5, got %f", config.Temperature)
	}
	if config.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout to be 30s, got %v", config.Timeout)
	}
	if !config.Debug {
		t.Error("Expected Debug to be true")
	}
}

func TestNewConfigFromEnvironment(t *testing.T) {
	// Save original environment
	originalAPIKey := os.Getenv(EnvAPIKey)
	defer func() {
		if originalAPIKey == "" {
			os.Unsetenv(EnvAPIKey)
		} else {
			os.Setenv(EnvAPIKey, originalAPIKey)
		}
	}()

	// Test with valid environment
	os.Setenv(EnvAPIKey, "test-env-key")

	config, err := NewConfigFromEnvironment()
	if err != nil {
		t.Fatalf("NewConfigFromEnvironment() error = %v", err)
	}

	if config.Auth == nil {
		t.Error("Expected Auth to be set")
	}

	// Verify defaults were applied
	if config.BaseURL == "" {
		t.Error("Expected BaseURL to have default value")
	}
	if config.Model == "" {
		t.Error("Expected Model to have default value")
	}
}

func TestConfigBuilder(t *testing.T) {
	config, err := NewConfigBuilder().
		WithAPIKey("builder-api-key").
		WithModel(ModelClaude35Sonnet).
		WithTimeout(45 * time.Second).
		WithTemperature(0.8).
		WithDebug(true).
		Build()

	if err != nil {
		t.Fatalf("ConfigBuilder.Build() error = %v", err)
	}

	if config.Auth == nil {
		t.Error("Expected Auth to be set")
	}
	if apiKeyAuth, ok := config.Auth.(*APIKeyAuth); !ok || apiKeyAuth.APIKey != "builder-api-key" {
		t.Error("Expected APIKeyAuth with correct API key")
	}
	if config.Model != ModelClaude35Sonnet {
		t.Errorf("Expected Model to be %s, got %s", ModelClaude35Sonnet, config.Model)
	}
	if config.Timeout != 45*time.Second {
		t.Errorf("Expected Timeout to be 45s, got %v", config.Timeout)
	}
	if config.Temperature != 0.8 {
		t.Errorf("Expected Temperature to be 0.8, got %f", config.Temperature)
	}
	if !config.Debug {
		t.Error("Expected Debug to be true")
	}
}

func TestConfigBuilderErrors(t *testing.T) {
	// Test builder with validation errors
	_, err := NewConfigBuilder().
		WithAPIKey(""). // Empty API key should cause error
		Build()

	if err == nil {
		t.Error("Expected ConfigBuilder.Build() to return error for empty API key")
	}

	// Test builder with invalid temperature
	_, err = NewConfigBuilder().
		WithAPIKey("test-key").
		WithTemperature(2.0). // Invalid temperature
		Build()

	if err == nil {
		t.Error("Expected ConfigBuilder.Build() to return error for invalid temperature")
	}
}

func TestConfigRedact(t *testing.T) {
	config := &Config{
		Auth: &APIKeyAuth{APIKey: "secret-api-key"},
		Proxy: &ProxyConfig{
			URL:      "http://proxy.example.com",
			Username: "user",
			Password: "secret-password",
		},
		Headers: map[string]string{
			"Authorization": "Bearer secret-token",
			"X-API-Key":     "another-secret",
			"Content-Type":  "application/json",
		},
	}

	redacted := config.Redact()

	// Check that sensitive data is redacted
	if redactedAuth, ok := redacted.Auth.(*RedactedAuth); !ok {
		t.Error("Expected Auth to be RedactedAuth")
	} else if redactedAuth.AuthType != AuthTypeAPIKey {
		t.Error("Expected RedactedAuth to preserve auth type")
	}

	if redacted.Proxy.Password != "[REDACTED]" {
		t.Error("Expected proxy password to be redacted")
	}
	if redacted.Proxy.Username != "user" {
		t.Error("Expected proxy username to be preserved")
	}

	if redacted.Headers["Authorization"] != "[REDACTED]" {
		t.Error("Expected Authorization header to be redacted")
	}
	if redacted.Headers["X-API-Key"] != "[REDACTED]" {
		t.Error("Expected X-API-Key header to be redacted")
	}
	if redacted.Headers["Content-Type"] != "application/json" {
		t.Error("Expected Content-Type header to be preserved")
	}

	// Verify original is unchanged
	if config.Proxy.Password != "secret-password" {
		t.Error("Expected original config to be unchanged")
	}
}

func TestConfigSanitize(t *testing.T) {
	config := &Config{
		Auth: &APIKeyAuth{APIKey: "secret-api-key"},
		Proxy: &ProxyConfig{
			URL:      "http://proxy.example.com",
			Username: "user",
			Password: "secret-password",
		},
		Headers: map[string]string{
			"Authorization": "Bearer secret-token",
			"Content-Type":  "application/json",
		},
	}

	// Create a copy and sanitize it
	sanitized := config.Clone()
	sanitized.Sanitize()

	// Check that sensitive data is removed
	if sanitized.Auth != nil {
		t.Error("Expected Auth to be removed")
	}
	if sanitized.Proxy.Username != "" || sanitized.Proxy.Password != "" {
		t.Error("Expected proxy credentials to be cleared")
	}
	if _, exists := sanitized.Headers["Authorization"]; exists {
		t.Error("Expected Authorization header to be removed")
	}
	if sanitized.Headers["Content-Type"] != "application/json" {
		t.Error("Expected Content-Type header to be preserved")
	}
}

func TestWithEnvironment(t *testing.T) {
	tests := []struct {
		name string
		env  Environment
		want func(*Config) bool
	}{
		{
			name: "production environment",
			env:  EnvironmentProduction,
			want: func(c *Config) bool {
				return !c.Debug && c.Timeout == 60*time.Second && c.RetryConfig.MaxRetries == 3
			},
		},
		{
			name: "development environment",
			env:  EnvironmentDevelopment,
			want: func(c *Config) bool {
				return c.Debug && c.Timeout == 30*time.Second && c.RetryConfig.MaxRetries == 1
			},
		},
		{
			name: "testing environment",
			env:  EnvironmentTesting,
			want: func(c *Config) bool {
				return !c.Debug && c.Timeout == 10*time.Second && c.RetryConfig.MaxRetries == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewConfig(
				WithAPIKey("test-key"),
				WithEnvironment(tt.env),
			)
			if err != nil {
				t.Fatalf("NewConfig() error = %v", err)
			}

			if !tt.want(config) {
				t.Errorf("Environment configuration not applied correctly for %s", tt.env)
			}
		})
	}
}

func TestValidateEnvironment(t *testing.T) {
	// Save original environment
	originalAPIKey := os.Getenv(EnvAPIKey)
	defer func() {
		if originalAPIKey == "" {
			os.Unsetenv(EnvAPIKey)
		} else {
			os.Setenv(EnvAPIKey, originalAPIKey)
		}
	}()

	// Test missing API key
	os.Unsetenv(EnvAPIKey)
	err := ValidateEnvironment()
	if err == nil {
		t.Error("Expected ValidateEnvironment() to return error for missing API key")
	}

	// Test valid environment
	os.Setenv(EnvAPIKey, "test-key")
	err = ValidateEnvironment()
	if err != nil {
		t.Errorf("ValidateEnvironment() error = %v", err)
	}
}

func TestIsZeroConfig(t *testing.T) {
	// Test nil config
	var config *Config
	if !config.IsZeroConfig() {
		t.Error("Expected nil config to be zero config")
	}

	// Test empty config
	config = &Config{}
	if !config.IsZeroConfig() {
		t.Error("Expected empty config to be zero config")
	}

	// Test config with values
	config = &Config{
		Auth: &APIKeyAuth{APIKey: "test"},
	}
	if config.IsZeroConfig() {
		t.Error("Expected config with auth to not be zero config")
	}
}
