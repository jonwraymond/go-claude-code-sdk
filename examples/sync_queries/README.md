# Synchronous Queries Example

This example demonstrates synchronous (non-streaming) query patterns with Claude Code, including basic queries, multi-turn conversations, system prompts, the QueryMessagesSync API, and robust error handling with retry mechanisms.

## What You'll Learn

- How to execute synchronous queries for simple request-response patterns
- Building multi-turn conversations with context preservation
- Using system prompts to customize Claude's behavior
- The QueryMessagesSync API for conversation management
- Error handling patterns and retry strategies
- Performance considerations for synchronous operations

## Code Overview

The example includes five synchronous query patterns:

### 1. Basic Synchronous Query
```go
request := &types.QueryRequest{
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "What are the key features of Go?"},
    },
    Model: "claude-3-5-sonnet-20241022",
}

response, err := claudeClient.Query(ctx, request)
if err != nil {
    log.Printf("Query failed: %v", err)
    return
}

// Extract and display response
content := extractTextContent(response.Content)
fmt.Printf("Response: %s\n", content)
```

Demonstrates the fundamental synchronous query pattern.

### 2. Multi-turn Conversation
```go
var messages []types.Message

// Turn 1
messages = append(messages, types.Message{
    Role:    types.RoleUser,
    Content: "Explain goroutines in Go.",
})

response1, err := claudeClient.Query(ctx, &types.QueryRequest{Messages: messages})
messages = append(messages, types.Message{
    Role:    types.RoleAssistant,
    Content: extractTextContent(response1.Content),
})

// Turn 2
messages = append(messages, types.Message{
    Role:    types.RoleUser,
    Content: "Can you provide a code example?",
})

response2, err := claudeClient.Query(ctx, &types.QueryRequest{Messages: messages})
```

Shows how to build and maintain conversation context across multiple turns.

### 3. System Prompt Usage
```go
request := &types.QueryRequest{
    System: "You are a Go programming expert. Provide concise, practical answers with code examples.",
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "How do channels work in Go?"},
    },
}

response, err := claudeClient.Query(ctx, request)
```

Demonstrates customizing Claude's behavior with system prompts.

### 4. QueryMessagesSync API
```go
options := &client.QueryOptions{
    Model:          "claude-3-5-sonnet-20241022",
    MaxTurns:       3,
    Stream:         false,
    PermissionMode: client.PermissionModeAcceptEdits,
    SystemPrompt:   "You are a helpful Go programming assistant.",
}

result, err := claudeClient.QueryMessagesSync(ctx, "Write a Go function to reverse a string", options)

// Access full conversation
for i, message := range result.Messages {
    fmt.Printf("%d. %s: %s\n", i+1, message.Role, formatContent(message.Content))
}
```

Shows the higher-level API for conversation management.

### 5. Error Handling and Retries
```go
func executeWithRetries(request *types.QueryRequest, maxRetries int) (*types.QueryResponse, error) {
    for attempt := 1; attempt <= maxRetries; attempt++ {
        queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
        
        response, err := claudeClient.Query(queryCtx, request)
        cancel()
        
        if err == nil {
            return response, nil
        }
        
        if attempt < maxRetries {
            waitTime := time.Duration(attempt*attempt) * time.Second
            time.Sleep(waitTime) // Exponential backoff
        }
    }
    return nil, fmt.Errorf("all %d attempts failed", maxRetries)
}
```

Demonstrates robust error handling with exponential backoff retry logic.

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
cd examples/sync_queries
go run main.go
```

## Expected Output

```
=== Claude Code Synchronous Queries Examples ===

--- Example 1: Basic Synchronous Query ---
Sending synchronous query...
✓ Query completed in 1.23s
Stop reason: end_turn
Response:
Go programming language has several key features that make it popular for modern software development:

1. **Simplicity and Readability**: Clean, minimalist syntax that's easy to learn and maintain
2. **Concurrency**: Built-in goroutines and channels for efficient concurrent programming
3. **Performance**: Compiled to native machine code for fast execution
4. **Garbage Collection**: Automatic memory management with low-latency GC
5. **Static Typing**: Compile-time type checking for better reliability
6. **Fast Compilation**: Quick build times even for large codebases
7. **Standard Library**: Comprehensive built-in packages for common tasks
8. **Cross-Platform**: Supports multiple operating systems and architectures

--- Example 2: Multi-turn Conversation ---
Turn 1 - User: Explain goroutines in Go.
Turn 1 - Claude: Goroutines are Go's lightweight threads for concurrent programming. They're managed by the Go runtime and allow you to run functions concurrently...

Turn 2 - User: Can you provide a simple code example?
Turn 2 - Claude: Here's a simple example demonstrating goroutines:

```go
package main

import (
    "fmt"
    "time"
)

func sayHello() {
    for i := 0; i < 3; i++ {
        fmt.Println("Hello", i)
        time.Sleep(100 * time.Millisecond)
    }
}

func main() {
    go sayHello() // Start as goroutine
    
    for i := 0; i < 3; i++ {
        fmt.Println("World", i)
        time.Sleep(100 * time.Millisecond)
    }
    
    time.Sleep(500 * time.Millisecond) // Wait for goroutine
}
```

✓ Multi-turn conversation completed
  Total messages in conversation: 3

--- Example 3: System Prompt Usage ---
Sending query with system prompt...
System: You are a Go programming expert. Provide concise, practical answers...
User: How do channels work in Go?
Claude: Channels in Go provide communication between goroutines. Here's how they work:

**Basic Channel Operations:**
```go
ch := make(chan int)        // Create channel
ch <- 42                    // Send value
value := <-ch               // Receive value
```

**Key Points:**
- Channels are typed: `chan int`, `chan string`, etc.
- Sends and receives block until both goroutines are ready
- Use `close(ch)` to signal no more values
- Range over channels: `for value := range ch`

**Example:**
```go
func main() {
    ch := make(chan string)
    
    go func() {
        ch <- "Hello from goroutine"
    }()
    
    message := <-ch
    fmt.Println(message)
}
```

✓ System prompt query completed

--- Example 4: QueryMessagesSync API ---
Executing QueryMessagesSync with prompt: Write a simple Go function that reverses a string
✓ QueryMessagesSync completed successfully
Messages in conversation:
  1. user: Write a simple Go function that reverses a string
  2. assistant: Here's a simple Go function to reverse a string:

```go
func reverseString(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}
```

This function:
- Converts string to rune slice (handles Unicode correctly)
- Uses two pointers to swap characters from both ends
- Returns the reversed string

Usage:
```go
fmt.Println(reverseString("Hello")) // Output: "olleH"
```

Metadata:
  model: claude-3-5-sonnet-20241022
  turn_count: 1

--- Example 5: Error Handling and Retries ---
Testing query with retry logic...
  Attempt 1/3...
  ✓ Success on attempt 1
✓ Query succeeded
Response: Hello! I'm Claude, an AI assistant. I'm here to help you with questions, tasks, or conversations...

Testing timeout handling...
✓ Timeout handled correctly: context deadline exceeded
```

## Key Concepts

### Request-Response Pattern

Synchronous queries follow a simple pattern:
1. **Send**: Submit complete request
2. **Wait**: Block until response received
3. **Process**: Handle complete response

### Message Types

```go
type Message struct {
    Role    MessageRole `json:"role"`
    Content interface{} `json:"content"`
}

// MessageRoles
const (
    RoleUser      MessageRole = "user"
    RoleAssistant MessageRole = "assistant"
    RoleSystem    MessageRole = "system"
    RoleTool      MessageRole = "tool"
)
```

### Response Structure

```go
type QueryResponse struct {
    Content    []ContentBlock `json:"content"`
    StopReason string         `json:"stop_reason"`
    Usage      UsageInfo      `json:"usage"`
}
```

## Advanced Usage

### Complex System Prompts
```go
systemPrompt := `You are a senior Go developer with 10+ years of experience.

Instructions:
- Provide production-ready code examples
- Include error handling
- Explain performance implications
- Suggest best practices
- Use modern Go idioms (Go 1.21+)

Format:
- Start with a brief explanation
- Provide complete, runnable code
- End with usage notes`

request := &types.QueryRequest{
    System:   systemPrompt,
    Messages: messages,
}
```

### Conversation State Management
```go
type ConversationManager struct {
    messages []types.Message
    client   *client.ClaudeCodeClient
}

func (cm *ConversationManager) AddUserMessage(content string) {
    cm.messages = append(cm.messages, types.Message{
        Role:    types.RoleUser,
        Content: content,
    })
}

func (cm *ConversationManager) GetResponse(ctx context.Context) (string, error) {
    response, err := cm.client.Query(ctx, &types.QueryRequest{
        Messages: cm.messages,
    })
    if err != nil {
        return "", err
    }
    
    content := extractTextContent(response.Content)
    cm.messages = append(cm.messages, types.Message{
        Role:    types.RoleAssistant,
        Content: content,
    })
    
    return content, nil
}
```

### Response Processing
```go
func processResponse(response *types.QueryResponse) {
    // Extract text content
    text := extractTextContent(response.Content)
    
    // Analyze response metadata
    fmt.Printf("Stop reason: %s\n", response.StopReason)
    fmt.Printf("Input tokens: %d\n", response.Usage.InputTokens)
    fmt.Printf("Output tokens: %d\n", response.Usage.OutputTokens)
    
    // Check for tool calls
    for _, block := range response.Content {
        if block.Type == "tool_use" {
            fmt.Printf("Tool called: %s\n", block.Name)
        }
    }
}
```

## Error Handling Strategies

### 1. Basic Error Handling
```go
response, err := claudeClient.Query(ctx, request)
if err != nil {
    log.Printf("Query failed: %v", err)
    return
}
```

### 2. Timeout Handling
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := claudeClient.Query(ctx, request)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        return fmt.Errorf("query timed out after 30 seconds")
    }
    return fmt.Errorf("query failed: %w", err)
}
```

### 3. Retry with Exponential Backoff
```go
func queryWithRetry(client *client.ClaudeCodeClient, request *types.QueryRequest) (*types.QueryResponse, error) {
    maxRetries := 3
    baseDelay := time.Second
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        if attempt > 0 {
            delay := time.Duration(1<<uint(attempt)) * baseDelay
            time.Sleep(delay)
        }
        
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        response, err := client.Query(ctx, request)
        cancel()
        
        if err == nil {
            return response, nil
        }
        
        // Don't retry certain errors
        if isNonRetriableError(err) {
            return nil, err
        }
        
        log.Printf("Attempt %d failed: %v", attempt+1, err)
    }
    
    return nil, fmt.Errorf("all %d attempts failed", maxRetries)
}
```

### 4. Circuit Breaker Pattern
```go
type CircuitBreaker struct {
    failureCount int
    lastFailTime time.Time
    threshold    int
    timeout      time.Duration
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.isOpen() {
        return fmt.Errorf("circuit breaker is open")
    }
    
    err := fn()
    if err != nil {
        cb.recordFailure()
        return err
    }
    
    cb.recordSuccess()
    return nil
}
```

## Performance Optimization

### Request Optimization
```go
// Use appropriate model for task complexity
request := &types.QueryRequest{
    Model: "claude-3-haiku-20240307", // Faster for simple tasks
    Messages: messages,
}

// Limit response length for better performance
request.MaxTokens = 1000
```

### Concurrent Requests
```go
func processQueriesConcurrently(queries []string) []string {
    results := make([]string, len(queries))
    var wg sync.WaitGroup
    
    for i, query := range queries {
        wg.Add(1)
        go func(index int, q string) {
            defer wg.Done()
            
            response, err := claudeClient.Query(ctx, &types.QueryRequest{
                Messages: []types.Message{{Role: types.RoleUser, Content: q}},
            })
            if err != nil {
                results[index] = fmt.Sprintf("Error: %v", err)
                return
            }
            
            results[index] = extractTextContent(response.Content)
        }(i, query)
    }
    
    wg.Wait()
    return results
}
```

### Response Caching
```go
type QueryCache struct {
    cache map[string]*types.QueryResponse
    mutex sync.RWMutex
    ttl   time.Duration
}

func (qc *QueryCache) Get(key string) (*types.QueryResponse, bool) {
    qc.mutex.RLock()
    defer qc.mutex.RUnlock()
    
    response, exists := qc.cache[key]
    return response, exists
}

func (qc *QueryCache) Set(key string, response *types.QueryResponse) {
    qc.mutex.Lock()
    defer qc.mutex.Unlock()
    
    qc.cache[key] = response
}
```

## Best Practices

### 1. Request Composition
```go
// Use clear, specific prompts
prompt := "Write a Go function that validates email addresses using regex. Include error handling and unit tests."

// Structure complex requests
request := &types.QueryRequest{
    System: "You are a Go expert focused on production-ready code.",
    Messages: []types.Message{
        {Role: types.RoleUser, Content: prompt},
    },
    Model: "claude-3-5-sonnet-20241022",
}
```

### 2. Response Validation
```go
func validateResponse(response *types.QueryResponse) error {
    if len(response.Content) == 0 {
        return fmt.Errorf("empty response content")
    }
    
    if response.StopReason != "end_turn" && response.StopReason != "stop_sequence" {
        return fmt.Errorf("unexpected stop reason: %s", response.StopReason)
    }
    
    return nil
}
```

### 3. Resource Management
```go
// Always use contexts with timeouts
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Monitor token usage
response, err := claudeClient.Query(ctx, request)
if err == nil {
    log.Printf("Tokens used: input=%d, output=%d", 
        response.Usage.InputTokens, response.Usage.OutputTokens)
}
```

## Troubleshooting

### Common Issues

1. **Slow Responses**
   - Use faster models for simple tasks
   - Reduce max_tokens for shorter responses
   - Implement request timeout

2. **Rate Limiting**
   - Implement exponential backoff
   - Use request queuing
   - Monitor API usage

3. **Large Context Issues**
   - Limit conversation history length
   - Summarize old messages
   - Use session management

### Debug Techniques
```go
// Enable debug logging
config.Debug = true

// Log request details
log.Printf("Sending request: model=%s, messages=%d", request.Model, len(request.Messages))

// Log response details
log.Printf("Response: stop_reason=%s, tokens=%d", response.StopReason, response.Usage.OutputTokens)
```

## Next Steps

After mastering synchronous queries, explore:
- [Streaming Queries](../streaming_queries/) - Real-time responses
- [Session Lifecycle](../session_lifecycle/) - Session management
- [Advanced Client](../advanced_client/) - Advanced features

## Related Documentation

- [Query Types](../../pkg/types/query.go)
- [Client Package](../../pkg/client/)
- [Error Handling Guide](../../docs/error-handling.md)