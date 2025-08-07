# Getting Started

## Installation

To install the Go Claude Code SDK, use `go get`:

```bash
go get github.com/jonwraymond/go-claude-code-sdk
```

## Usage

Here's a simple example of how to use the SDK:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jonwraymond/go-claude-code-sdk/pkg/client"
    "github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
    ctx := context.Background()

    // Minimal config: auto-detect auth (subscription or API key)
    cfg := types.NewClaudeCodeConfig()

    // Create the Claude Code client (CLI wrapper)
    cc, err := client.NewClaudeCodeClient(ctx, cfg)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer cc.Close()

    // Simple query
    resp, err := cc.Query(ctx, &types.QueryRequest{
        Messages: []types.Message{
            {Role: types.RoleUser, Content: "Hello Claude Code"},
        },
        MaxTokens: 1024,
    })
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }

    fmt.Println(resp.GetTextContent())
}
```
