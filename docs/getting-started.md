# Getting Started

## Installation

To install the Go Claude Code SDK, use `go get`:

```bash
go get github.com/your-username/go-claude-code-sdk
```

## Usage

Here's a simple example of how to use the SDK:

```go
package main

import (
	"fmt"
	"log"

	"github.com/your-username/go-claude-code-sdk/pkg/client"
)

func main() {
	// Create a new client
	c, err := client.NewClient("your-api-key")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Use the client to interact with the Claude API
	// ...
}
```
