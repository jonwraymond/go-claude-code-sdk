package api

import "context"

// Convenience wrappers with typed signatures (lightweight to reduce scope now).

type MessageParam struct {
    Role    string      `json:"role"`
    Content interface{} `json:"content"`
}

type CreateRequest struct {
    Model     string         `json:"model"`
    MaxTokens int            `json:"max_tokens"`
    Messages  []MessageParam `json:"messages"`
    Stream    bool           `json:"stream,omitempty"`
    Tools     interface{}    `json:"tools,omitempty"`
    ToolChoice interface{}   `json:"tool_choice,omitempty"`
}

func (c *Client) CreateMessage(ctx context.Context, r *CreateRequest) (map[string]any, error) {
    return c.MessagesCreate(ctx, map[string]any{
        "model":      r.Model,
        "max_tokens": r.MaxTokens,
        "messages":   r.Messages,
        "stream":     r.Stream,
        "tools":      r.Tools,
        "tool_choice": r.ToolChoice,
    })
}

type CountTokensRequest struct {
    Model    string         `json:"model"`
    Messages []MessageParam `json:"messages"`
}

func (c *Client) CountTokens(ctx context.Context, r *CountTokensRequest) (map[string]any, error) {
    return c.MessagesCountTokens(ctx, map[string]any{
        "model":    r.Model,
        "messages": r.Messages,
    })
}


