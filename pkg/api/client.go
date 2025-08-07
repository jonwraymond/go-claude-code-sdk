package api

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
    "time"
)

// Client is a minimal REST client for Anthropic API parity features.
type Client struct {
    cfg *Config
}

func NewClient(cfg *Config) (*Client, error) {
    if cfg == nil {
        cfg = DefaultConfig()
    }
    if cfg.APIKey == "" {
        return nil, errors.New("api: API key is required")
    }
    if cfg.HTTPClient == nil {
        cfg.HTTPClient = &http.Client{Timeout: cfg.Timeout}
    }
    return &Client{cfg: cfg}, nil
}

// MessagesCreate implements a minimal messages.create equivalent.
func (c *Client) MessagesCreate(ctx context.Context, body map[string]any) (map[string]any, error) {
    return c.doJSON(ctx, http.MethodPost, "/messages", body)
}

// MessagesCountTokens implements messages.count_tokens.
func (c *Client) MessagesCountTokens(ctx context.Context, body map[string]any) (map[string]any, error) {
    return c.doJSON(ctx, http.MethodPost, "/messages/count_tokens", body)
}

// BatchesCreate implements messages.batches.create.
func (c *Client) BatchesCreate(ctx context.Context, body map[string]any) (map[string]any, error) {
    return c.doJSON(ctx, http.MethodPost, "/messages/batches", body)
}

// BatchesList implements messages.batches.list (cursor-based pagination supported via query).
func (c *Client) BatchesList(ctx context.Context, query url.Values) (map[string]any, error) {
    path := "/messages/batches"
    if len(query) > 0 {
        path += "?" + query.Encode()
    }
    return c.doJSON(ctx, http.MethodGet, path, nil)
}

// BatchesResults streams results. Here we return an iterator-like pull interface.
func (c *Client) BatchesResults(ctx context.Context, batchID string) (io.ReadCloser, *http.Response, error) {
    req, err := c.newRequest(ctx, http.MethodGet, fmt.Sprintf("/messages/batches/%s/results", batchID), nil)
    if err != nil { return nil, nil, err }
    resp, err := c.cfg.HTTPClient.Do(req)
    if err != nil { return nil, nil, err }
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        defer resp.Body.Close()
        return nil, resp, fmt.Errorf("api: status %d", resp.StatusCode)
    }
    return resp.Body, resp, nil
}

// BetaFilesUpload implements beta files.upload; content can be bytes, path, or multipart placeholder.
func (c *Client) BetaFilesUpload(ctx context.Context, filename string, content []byte, mediaType string) (map[string]any, error) {
    // Minimal placeholder: send raw bytes with content-type and filename header
    headers := map[string]string{
        "Content-Type": mediaType,
        "X-Filename":   filename,
    }
    return c.doRaw(ctx, http.MethodPost, "/beta/files", content, headers)
}

// Helper methods
func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
    base := strings.TrimRight(c.cfg.BaseURL, "/")
    req, err := http.NewRequestWithContext(ctx, method, base+path, body)
    if err != nil { return nil, err }
    req.Header.Set("anthropic-version", c.cfg.APIVersion)
    req.Header.Set("x-api-key", c.cfg.APIKey)
    for k, v := range c.cfg.DefaultHeaders { req.Header.Set(k, v) }
    return req, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, payload map[string]any) (map[string]any, error) {
    var reader io.Reader
    if payload != nil {
        b, err := json.Marshal(payload)
        if err != nil { return nil, err }
        reader = strings.NewReader(string(b))
    }
    req, err := c.newRequest(ctx, method, path, reader)
    if err != nil { return nil, err }
    if payload != nil { req.Header.Set("Content-Type", "application/json") }
    return c.do(req)
}

func (c *Client) doRaw(ctx context.Context, method, path string, payload []byte, headers map[string]string) (map[string]any, error) {
    var reader io.Reader
    if payload != nil { reader = strings.NewReader(string(payload)) }
    req, err := c.newRequest(ctx, method, path, reader)
    if err != nil { return nil, err }
    for k, v := range headers { req.Header.Set(k, v) }
    return c.do(req)
}

func (c *Client) do(req *http.Request) (map[string]any, error) {
    // Simple retries
    var lastErr error
    backoff := c.cfg.InitialBackoff
    for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
        resp, err := c.cfg.HTTPClient.Do(req)
        if err != nil { lastErr = err; time.Sleep(backoff); backoff = minDur(backoff*2, c.cfg.MaxBackoff); continue }
        defer resp.Body.Close()
        if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
            b, _ := io.ReadAll(resp.Body)
            lastErr = fmt.Errorf("api: retryable status %d: %s", resp.StatusCode, string(b))
            time.Sleep(backoff)
            backoff = minDur(backoff*2, c.cfg.MaxBackoff)
            continue
        }
        if resp.StatusCode < 200 || resp.StatusCode >= 300 {
            b, _ := io.ReadAll(resp.Body)
            return nil, fmt.Errorf("api: status %d: %s", resp.StatusCode, string(b))
        }
        var out map[string]any
        if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return nil, err }
        return out, nil
    }
    if lastErr == nil { lastErr = errors.New("api: request failed") }
    return nil, lastErr
}

func minDur(a, b time.Duration) time.Duration { if a < b { return a }; return b }


