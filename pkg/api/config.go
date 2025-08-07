package api

import (
    "net/http"
    "time"
)

// Config contains configuration for the Anthropic REST API client.
type Config struct {
    APIKey         string
    BaseURL        string
    APIVersion     string
    Timeout        time.Duration
    MaxRetries     int
    InitialBackoff time.Duration
    MaxBackoff     time.Duration
    HTTPClient     *http.Client
    DefaultHeaders map[string]string
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() *Config {
    return &Config{
        BaseURL:        "https://api.anthropic.com/v1",
        APIVersion:     "2023-06-01",
        Timeout:        10 * time.Minute,
        MaxRetries:     2,
        InitialBackoff: 500 * time.Millisecond,
        MaxBackoff:     10 * time.Second,
        DefaultHeaders: map[string]string{},
    }
}


