package api

import (
    "context"
)

type FileUploadResponse map[string]any

// UploadFile wraps BetaFilesUpload.
func (c *Client) UploadFile(ctx context.Context, filename string, content []byte, mediaType string) (FileUploadResponse, error) {
    return c.BetaFilesUpload(ctx, filename, content, mediaType)
}


