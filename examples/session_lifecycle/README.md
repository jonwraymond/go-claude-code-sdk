# Session Lifecycle Example

This example demonstrates comprehensive session management with Claude Code, including session creation, persistence, multi-session handling, custom configuration, and proper cleanup.

## What You'll Learn

- How to create and manage Claude Code sessions
- Session persistence across client instances
- Managing multiple concurrent sessions
- Custom session configuration
- Session cleanup and resource management
- Session isolation and context separation

## Code Overview

The example includes five session management patterns:

### 1. Basic Session Management
```go
// Generate a new session ID
sessionID := claudeClient.GenerateSessionID()

// Create a new session
session, err := claudeClient.CreateSession(ctx, sessionID)

// Use the session for conversations
options := &client.QueryOptions{
    SessionID:      session.ID,
    MaxTurns:       5,
    PermissionMode: client.PermissionModeAsk,
}

result, err := claudeClient.QueryMessagesSync(ctx, "Hello!", options)
```

Demonstrates basic session creation and usage patterns.

### 2. Session Persistence
```go
// Create session with first client
client1, _ := client.NewClaudeCodeClient(ctx, config1)
sessionID := client1.GenerateSessionID()
session, _ := client1.CreateSession(ctx, sessionID)

// Use session with first client
result1, _ := client1.QueryMessagesSync(ctx, "Remember this: Go is great!", options)
client1.Close()

// Create second client with same session ID
config2.SessionID = sessionID
client2, _ := client.NewClaudeCodeClient(ctx, config2)

// Continue conversation with second client
result2, _ := client2.QueryMessagesSync(ctx, "What did I tell you to remember?", options)
```

Shows how sessions persist across different client instances.

### 3. Multiple Concurrent Sessions
```go
sessions := make(map[string]*client.ClaudeCodeSession)
topics := []string{"golang", "databases", "algorithms"}

for _, topic := range topics {
    sessionID := claudeClient.GenerateSessionID()
    session, _ := claudeClient.CreateSession(ctx, sessionID)
    sessions[topic] = session
}

// Have separate conversations in each session
for topic, session := range sessions {
    options := &client.QueryOptions{SessionID: session.ID}
    result, _ := claudeClient.QueryMessagesSync(ctx, prompts[topic], options)
}
```

Demonstrates managing multiple sessions for different contexts.

### 4. Custom Session Configuration
```go
config := &types.ClaudeCodeConfig{
    Model:     "claude-3-5-sonnet-20241022",
    SessionID: fmt.Sprintf("custom-session-%d", time.Now().Unix()),
    Debug:     true,
}

options := &client.QueryOptions{
    SessionID:      session.ID,
    Model:          "claude-3-5-sonnet-20241022",
    SystemPrompt:   "You are a Go programming assistant.",
    MaxTurns:       3,
    PermissionMode: client.PermissionModeAcceptEdits,
}
```

Shows custom session configuration with specific models and prompts.

### 5. Session Cleanup
```go
// Create temporary sessions
var tempSessions []*client.ClaudeCodeSession
for i := 0; i < 3; i++ {
    session, _ := claudeClient.CreateSession(ctx, sessionID)
    tempSessions = append(tempSessions, session)
}

// Clean up sessions explicitly
for _, session := range tempSessions {
    err := session.Close()
    if err != nil {
        log.Printf("Error closing session: %v", err)
    }
}
```

Demonstrates proper session cleanup and resource management.

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
cd examples/session_lifecycle
go run main.go
```

## Expected Output

```
=== Claude Code Session Lifecycle Examples ===

--- Example 1: Basic Session Management ---
Generated session ID: session-1704123456789-abc123
✓ Session created successfully
  Session ID: session-1704123456789-abc123

Starting conversation in session...
Claude: Hello! Yes, I can remember this conversation. Each session maintains...

Claude: You asked me about remembering conversations. Sessions allow me to...

Active sessions: 1
  1. session-1704123456789-abc123

--- Example 2: Session Persistence ---
Created session with first client: session-1704123456790-def456
Client 1 - Claude: I'll remember that your favorite programming language is Go...
✓ First client closed
Created second client with same session ID
✓ Retrieved session: session-1704123456790-def456
Client 2 - Claude: You mentioned that your favorite programming language is Go...
✓ Session persistence verified

--- Example 3: Multiple Concurrent Sessions ---
✓ Created session for golang: 12345678...
✓ Created session for databases: 23456789...
✓ Created session for algorithms: 34567890...

Conversation in golang session:
  Q: Explain Go's interface system briefly.
  A: Go's interface system is based on implicit satisfaction - types automatically...

Conversation in databases session:
  Q: What's the difference between SQL and NoSQL databases?
  A: SQL databases are relational and use structured schemas with ACID properties...

Conversation in algorithms session:
  Q: Explain the Big O notation concept.
  A: Big O notation describes the upper bound of algorithm complexity, showing how...

Testing session isolation:
  Database session response to Go question: I don't recall discussing Go interfaces...
✓ Multiple concurrent sessions demonstrated

--- Example 4: Custom Session Configuration ---
✓ Client created with custom session configuration
  Session ID: custom-session-1704123456791
  Model: claude-3-5-sonnet-20241022
  Debug: true

Using custom session configuration...
✓ Query executed with custom configuration
Response: Here's a simple Go function to calculate factorial:

func factorial(n int) int {
    if n <= 1 {
        return 1
    }
    return n * factorial(n-1)
}

--- Example 5: Session Cleanup and Resource Management ---
✓ Created temporary session 1: 45678901...
✓ Created temporary session 2: 56789012...
✓ Created temporary session 3: 67890123...

Sessions before cleanup: 4
Cleaning up temporary sessions...
✓ Closed temporary session 1
✓ Closed temporary session 2
✓ Closed temporary session 3
Sessions after cleanup: 1

Session resource information:
  Active sessions: 1
✓ Session cleanup example completed
  (Client will be cleaned up automatically)
```

## Key Concepts

### Session Lifecycle

1. **Creation**: Generate ID and create session
2. **Usage**: Conduct conversations within session context
3. **Persistence**: Sessions survive client restarts
4. **Cleanup**: Explicit session closure when done

### Session IDs

- **Auto-generated**: UUID format with timestamp
- **Custom**: Use meaningful names for tracking
- **Persistence**: Same ID = same conversation context

### Session Isolation

Each session maintains:
- **Separate conversation history**
- **Independent context**
- **Isolated memory**
- **Individual configuration**

### Session Options

```go
type QueryOptions struct {
    SessionID      string
    Model          string
    SystemPrompt   string
    MaxTurns       int
    PermissionMode PermissionMode
    Stream         bool
}
```

## Advanced Usage

### Session with Custom System Prompt
```go
options := &client.QueryOptions{
    SessionID:      session.ID,
    SystemPrompt:   "You are an expert Go developer. Provide concise, practical advice with code examples.",
    PermissionMode: client.PermissionModeAcceptEdits,
}
```

### Session with Turn Limits
```go
options := &client.QueryOptions{
    SessionID: session.ID,
    MaxTurns:  5, // Limit conversation length
}
```

### Session Monitoring
```go
// List all active sessions
sessions := claudeClient.ListSessions()
fmt.Printf("Active sessions: %d\n", len(sessions))

// Get specific session
session, err := claudeClient.GetSession(sessionID)
if err != nil {
    log.Printf("Session not found: %v", err)
}
```

## Best Practices

### 1. Session Naming
```go
// Use descriptive session IDs
sessionID := fmt.Sprintf("user-%s-task-%s-%d", userID, taskType, timestamp)
```

### 2. Resource Management
```go
// Always clean up sessions
defer session.Close()

// Or use automatic cleanup
defer claudeClient.Close() // Closes all sessions
```

### 3. Error Handling
```go
session, err := claudeClient.CreateSession(ctx, sessionID)
if err != nil {
    log.Printf("Failed to create session: %v", err)
    return
}
```

### 4. Session Isolation
```go
// Use different sessions for different contexts
userSession := claudeClient.CreateSession(ctx, "user-session")
adminSession := claudeClient.CreateSession(ctx, "admin-session")
```

## Performance Considerations

### Session Overhead
- Each session maintains conversation history
- Memory usage grows with conversation length
- Consider session cleanup for long-running applications

### Concurrent Sessions
- Multiple sessions can run simultaneously
- Each session is independent
- No shared state between sessions

### Session Persistence
- Sessions are stored by Claude Code CLI
- Persistence survives application restarts
- Consider cleanup for temporary sessions

## Troubleshooting

### Common Issues

1. **Session Not Found**
   - Verify session ID is correct
   - Check if session was properly created
   - Ensure client has access to session

2. **Memory Issues**
   - Clean up unused sessions
   - Limit conversation history length
   - Monitor session count

3. **Persistence Issues**
   - Ensure Claude Code CLI is properly installed
   - Check file permissions
   - Verify session storage location

### Debug Session Issues
```go
// Enable debug mode
config.Debug = true

// List all sessions
sessions := claudeClient.ListSessions()
for _, sessionID := range sessions {
    fmt.Printf("Active session: %s\n", sessionID)
}
```

## Next Steps

After understanding session management, explore:
- [Sync Queries](../sync_queries/) - Making API calls within sessions
- [Streaming Queries](../streaming_queries/) - Real-time responses in sessions
- [Advanced Client](../advanced_client/) - Advanced client features

## Related Documentation

- [Session Types](../../pkg/types/session.go)
- [Client Package](../../pkg/client/)
- [Query Options Guide](../../docs/query-options.md)