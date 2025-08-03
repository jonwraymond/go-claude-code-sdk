# Streaming Queries Example

This example demonstrates real-time streaming capabilities with Claude Code, including basic streaming, context cancellation, advanced chunk processing, and the QueryMessages streaming API.

## What You'll Learn

- How to execute streaming queries for real-time responses
- Context cancellation and timeout handling with streams
- Advanced chunk processing and analysis
- Using the QueryMessages streaming API
- Performance monitoring and statistics
- Error handling in streaming contexts

## Code Overview

The example includes four streaming patterns:

### 1. Basic Streaming Query
```go
request := &types.QueryRequest{
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "Write a short poem about Go programming."},
    },
}

stream, err := claudeClient.QueryStream(ctx, request)
defer stream.Close()

for {
    chunk, err := stream.Recv()
    if err != nil || chunk.Done {
        break
    }
    fmt.Print(chunk.Content) // Real-time output
}
```

Demonstrates the fundamental streaming query pattern with real-time content display.

### 2. Streaming with Cancellation
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

stream, err := claudeClient.QueryStream(ctx, request)
for {
    chunk, err := stream.Recv()
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            fmt.Printf("Stream cancelled due to timeout")
        }
        break
    }
    // Process chunk...
}
```

Shows how to handle timeouts and cancellation during streaming operations.

### 3. Advanced Chunk Processing
```go
var (
    chunkCount    int
    totalBytes    int
    wordCount     int
    startTime     = time.Now()
)

for {
    chunk, err := stream.Recv()
    if err != nil || chunk.Done {
        break
    }
    
    // Update statistics
    chunkCount++
    totalBytes += len(chunk.Content)
    wordCount += len(strings.Fields(chunk.Content))
    
    fmt.Print(chunk.Content)
}

elapsed := time.Since(startTime)
fmt.Printf("Throughput: %.2f words/second\n", float64(wordCount)/elapsed.Seconds())
```

Demonstrates streaming analytics and performance monitoring.

### 4. QueryMessages Streaming API
```go
options := &client.QueryOptions{
    Model:          "claude-3-5-sonnet-20241022",
    Stream:         true,
    PermissionMode: client.PermissionModeAsk,
}

messageChan, err := claudeClient.QueryMessages(ctx, prompt, options)

for message := range messageChan {
    switch message.Role {
    case types.RoleUser:
        fmt.Printf("👤 User: %s\n", message.Content)
    case types.RoleAssistant:
        fmt.Printf("🤖 Claude: %s\n", message.Content)
    }
}
```

Shows the higher-level QueryMessages streaming API for conversation management.

## Prerequisites

1. **Go 1.21+** installed
2. **Claude Code CLI** installed: `npm install -g @anthropic/claude-code`
3. **Authentication** configured

## Running the Example

### Setup Authentication
```bash
# Option 1: API Key
export ANTHROPIC_API_KEY="your-api-key-here"

# Option 2: Claude Code CLI subscription
claude auth
```

### Run the Example
```bash
cd examples/streaming_queries
go run main.go
```

## Expected Output

```
=== Claude Code Streaming Queries Examples ===

--- Example 1: Basic Streaming Query ---
Starting streaming query...
Response:
In circuits of logic, where bits dance free,
Go stands as guardian of simplicity.
Goroutines like rivers flow side by side,
While channels connect what might otherwise hide.

Concurrent and clean, with syntax so bright,
Go makes complex systems feel surprisingly light.
From Google's design to the cloud up above,
Programming in Go is a developer's love.

✓ Stream completed
Full response length: 324 characters

--- Example 2: Streaming with Cancellation ---
Starting streaming query with 5-second timeout...
Response (will timeout after 5 seconds):
Once upon a time in a world of code,
Where developers traveled the debugging road,
There lived a language called Go so neat,
Making concurrent programming quite the feat...

⚠ Stream cancelled due to timeout

--- Example 3: Advanced Chunk Processing ---
Starting streaming query with chunk analysis...
Response:
---
Go's concurrency model is built around three main concepts: goroutines, channels, and the select statement.

Goroutines are lightweight threads managed by the Go runtime. You can create thousands of them without significant overhead. Here's a simple example:

```go
func main() {
    go sayHello()
    go sayWorld()
    time.Sleep(time.Second)
}

func sayHello() {
    fmt.Println("Hello")
}

func sayWorld() {
    fmt.Println("World")
}
```

Channels provide communication between goroutines...
---
✓ Streaming completed
Statistics:
  Chunks received: 47
  Total bytes: 1,247
  Word count: 203
  Duration: 2.34s
  Throughput: 86.75 words/second

--- Example 4: QueryMessages Streaming API ---
Starting QueryMessages stream for: Help me understand goroutines with a simple example
Conversation:
---
👤 User: Help me understand goroutines with a simple example

🤖 Claude: I'd be happy to explain goroutines! Goroutines are Go's way of handling concurrent execution - they're like lightweight threads that make it easy to run multiple tasks simultaneously.

Here's a simple example:

```go
package main

import (
    "fmt"
    "time"
)

func printNumbers() {
    for i := 1; i <= 5; i++ {
        fmt.Printf("Number: %d\n", i)
        time.Sleep(1 * time.Second)
    }
}

func printLetters() {
    for _, letter := range []string{"A", "B", "C", "D", "E"} {
        fmt.Printf("Letter: %s\n", letter)
        time.Sleep(1 * time.Second)
    }
}

func main() {
    // Start both functions as goroutines
    go printNumbers()
    go printLetters()
    
    // Wait for both to complete
    time.Sleep(6 * time.Second)
    fmt.Println("All done!")
}
```

Key points about this example:
- The `go` keyword creates a new goroutine
- Both functions run concurrently (at the same time)
- The output will be interleaved: Number: 1, Letter: A, Number: 2, Letter: B, etc.
- We use `time.Sleep(6 * time.Second)` to wait for both goroutines to finish

---
✓ QueryMessages stream completed
  Total messages: 2
```

## Key Concepts

### Streaming vs Synchronous

| Aspect | Streaming | Synchronous |
|--------|-----------|-------------|
| **Response Time** | Immediate/incremental | Complete response |
| **User Experience** | Real-time feedback | Wait for completion |
| **Memory Usage** | Low (chunk-based) | High (full response) |
| **Use Cases** | Interactive, long responses | Simple Q&A |

### Stream Lifecycle

1. **Initialization**: Create stream with `QueryStream()`
2. **Processing**: Receive chunks with `stream.Recv()`
3. **Completion**: Detect `chunk.Done` or receive error
4. **Cleanup**: Always call `stream.Close()`

### Chunk Structure
```go
type StreamChunk struct {
    Content string `json:"content"`
    Done    bool   `json:"done"`
    Error   string `json:"error,omitempty"`
}
```

## Advanced Usage

### Custom Stream Processing
```go
type StreamProcessor struct {
    buffer     strings.Builder
    chunkCount int
    startTime  time.Time
}

func (p *StreamProcessor) ProcessChunk(chunk *types.StreamChunk) {
    if chunk.Done {
        p.finalize()
        return
    }
    
    p.chunkCount++
    p.buffer.WriteString(chunk.Content)
    
    // Custom processing logic
    if p.shouldFlush() {
        p.flush()
    }
}
```

### Stream Buffering
```go
buffer := make([]string, 0, 100)
for {
    chunk, err := stream.Recv()
    if err != nil || chunk.Done {
        break
    }
    
    buffer = append(buffer, chunk.Content)
    
    // Process buffer when full
    if len(buffer) >= 10 {
        processBuffer(buffer)
        buffer = buffer[:0] // Reset
    }
}
```

### Performance Monitoring
```go
type StreamMetrics struct {
    StartTime    time.Time
    ChunkCount   int
    TotalBytes   int
    WordCount    int
    AvgChunkSize float64
}

func (m *StreamMetrics) UpdateFromChunk(chunk *types.StreamChunk) {
    m.ChunkCount++
    m.TotalBytes += len(chunk.Content)
    m.WordCount += len(strings.Fields(chunk.Content))
    m.AvgChunkSize = float64(m.TotalBytes) / float64(m.ChunkCount)
}
```

## Error Handling

### Stream Error Patterns
```go
for {
    chunk, err := stream.Recv()
    if err != nil {
        if err == io.EOF {
            // Normal stream end
            break
        }
        if ctx.Err() == context.DeadlineExceeded {
            // Timeout
            return fmt.Errorf("stream timeout: %w", err)
        }
        // Other errors
        return fmt.Errorf("stream error: %w", err)
    }
    
    if chunk.Done {
        break
    }
    
    // Process chunk
}
```

### Graceful Degradation
```go
func executeWithFallback(ctx context.Context, request *types.QueryRequest) {
    // Try streaming first
    stream, err := claudeClient.QueryStream(ctx, request)
    if err != nil {
        // Fall back to synchronous
        response, err := claudeClient.Query(ctx, request)
        if err != nil {
            log.Printf("Both streaming and sync failed: %v", err)
            return
        }
        processResponse(response)
        return
    }
    
    processStream(stream)
}
```

## Best Practices

### 1. Always Close Streams
```go
stream, err := claudeClient.QueryStream(ctx, request)
if err != nil {
    return err
}
defer stream.Close() // Essential for resource cleanup
```

### 2. Handle Context Cancellation
```go
select {
case <-ctx.Done():
    return ctx.Err()
default:
    chunk, err := stream.Recv()
    // Process chunk...
}
```

### 3. Implement Timeout Protection
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

### 4. Buffer for UI Updates
```go
// Avoid excessive UI updates
if time.Since(lastUpdate) > 100*time.Millisecond {
    updateUI(buffer.String())
    lastUpdate = time.Now()
}
```

## Performance Tips

1. **Chunk Processing**: Process chunks as received, don't accumulate
2. **UI Updates**: Batch UI updates to avoid excessive redraws
3. **Memory Management**: Use bounded buffers for long streams
4. **Error Recovery**: Implement retry logic for transient failures

## Use Cases

### Interactive Chat Applications
```go
for {
    chunk, err := stream.Recv()
    if err != nil || chunk.Done {
        break
    }
    
    // Update chat UI in real-time
    chatWidget.AppendText(chunk.Content)
    chatWidget.ScrollToBottom()
}
```

### Code Generation with Preview
```go
codeBuffer := strings.Builder{}
for {
    chunk, err := stream.Recv()
    if err != nil || chunk.Done {
        break
    }
    
    codeBuffer.WriteString(chunk.Content)
    
    // Update code preview
    if isValidCode(codeBuffer.String()) {
        codeEditor.SetContent(codeBuffer.String())
    }
}
```

### Progress Indicators
```go
totalExpected := estimateResponseLength(request)
received := 0

for {
    chunk, err := stream.Recv()
    if err != nil || chunk.Done {
        break
    }
    
    received += len(chunk.Content)
    progress := float64(received) / float64(totalExpected)
    progressBar.SetProgress(progress)
}
```

## Troubleshooting

### Common Issues

1. **Stream Hangs**
   - Implement context timeouts
   - Check network connectivity
   - Monitor for infinite loops in processing

2. **Memory Leaks**
   - Always call `stream.Close()`
   - Avoid accumulating all chunks in memory
   - Use bounded buffers

3. **UI Freezing**
   - Process chunks in background goroutines
   - Batch UI updates
   - Use channels for UI communication

### Debug Streaming Issues
```go
stream, err := claudeClient.QueryStream(ctx, request)
if err != nil {
    log.Printf("Failed to create stream: %v", err)
    return
}

log.Printf("Stream created successfully")
defer func() {
    if err := stream.Close(); err != nil {
        log.Printf("Error closing stream: %v", err)
    } else {
        log.Printf("Stream closed successfully")
    }
}()
```

## Next Steps

After understanding streaming queries, explore:
- [Sync Queries](../sync_queries/) - Compare with synchronous patterns
- [Session Lifecycle](../session_lifecycle/) - Session management
- [Advanced Client](../advanced_client/) - Advanced client features

## Related Documentation

- [Stream Types](../../pkg/types/stream.go)
- [Query Options](../../docs/query-options.md)
- [Performance Guide](../../docs/performance.md)