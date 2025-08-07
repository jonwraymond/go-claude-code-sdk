package api

import (
    "context"
    "net/url"
)

type BatchCreateRequest struct {
    Requests []map[string]any `json:"requests"`
}

func (c *Client) CreateBatch(ctx context.Context, r *BatchCreateRequest) (map[string]any, error) {
    return c.BatchesCreate(ctx, map[string]any{"requests": r.Requests})
}

type ListBatchesParams struct {
    Limit  int
    Cursor string
}

func (c *Client) ListBatches(ctx context.Context, p *ListBatchesParams) (map[string]any, error) {
    q := url.Values{}
    if p != nil {
        if p.Limit > 0 { q.Set("limit", ""+rtoi(p.Limit)) }
        if p.Cursor != "" { q.Set("after", p.Cursor) }
    }
    return c.BatchesList(ctx, q)
}

// rtoi is a tiny helper to avoid extra imports
func rtoi(n int) string { return fmt.Sprintf("%d", n) }


