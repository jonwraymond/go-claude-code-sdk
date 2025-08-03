# Subscription Authentication Example

This example demonstrates focused subscription authentication patterns with the Claude Code SDK, including setup, validation, troubleshooting, and advanced subscription management techniques.

## What You'll Learn

- How to configure and use Claude Code CLI subscription authentication
- Subscription authentication setup and validation 
- Advanced subscription management and monitoring
- Troubleshooting subscription authentication issues
- Best practices for subscription-based deployments
- Integration patterns for development workflows

## Code Overview

The example focuses on subscription authentication patterns:

### 1. Basic Subscription Setup
```go
config := &types.ClaudeCodeConfig{
    WorkingDirectory: ".",
    AuthMethod:       types.AuthTypeSubscription,
}

// Validate subscription before client creation
subscriptionAuth := &types.SubscriptionAuth{}
if !subscriptionAuth.IsValid(ctx) {
    return fmt.Errorf("subscription authentication not available")
}

client, err := client.NewClaudeCodeClient(ctx, config)
```

Demonstrates basic subscription authentication setup and validation.

### 2. Subscription Status Monitoring
```go
func monitorSubscriptionStatus(ctx context.Context) {
    subscriptionAuth := &types.SubscriptionAuth{}
    
    // Check current status
    isValid := subscriptionAuth.IsValid(ctx)
    fmt.Printf("Subscription status: %v\n", isValid)
    
    // Get detailed subscription info
    info, err := subscriptionAuth.GetSubscriptionInfo(ctx)
    if err != nil {
        log.Printf("Failed to get subscription info: %v", err)
        return
    }
    
    fmt.Printf("Subscription details:\n")
    fmt.Printf("  User: %s\n", info.User)
    fmt.Printf("  Plan: %s\n", info.Plan)
    fmt.Printf("  Valid until: %s\n", info.ExpiresAt.Format("2006-01-02"))
}
```

Shows subscription status monitoring and detailed information retrieval.

### 3. Subscription Renewal and Management
```go
func handleSubscriptionRenewal(ctx context.Context) error {
    subscriptionAuth := &types.SubscriptionAuth{}
    
    // Check if renewal is needed
    info, err := subscriptionAuth.GetSubscriptionInfo(ctx)
    if err != nil {
        return fmt.Errorf("failed to get subscription info: %w", err)
    }
    
    // Check expiration
    if time.Until(info.ExpiresAt) < 7*24*time.Hour {
        fmt.Println("⚠️  Subscription expires within 7 days")
        fmt.Println("💡 Consider renewing your subscription")
        
        // Automated renewal (if supported)
        if info.AutoRenewal {
            fmt.Println("✅ Auto-renewal is enabled")
        } else {
            fmt.Println("❌ Auto-renewal is disabled")
            fmt.Println("   Please renew manually via: claude subscription renew")
        }
    }
    
    return nil
}
```

Demonstrates subscription lifecycle management and renewal handling.

### 4. Development Workflow Integration
```go
func setupDevelopmentWithSubscription(ctx context.Context) error {
    // Ensure subscription is available
    subscriptionAuth := &types.SubscriptionAuth{}
    if !subscriptionAuth.IsValid(ctx) {
        return fmt.Errorf("development requires valid subscription")
    }
    
    config := &types.ClaudeCodeConfig{
        WorkingDirectory: ".",
        AuthMethod:       types.AuthTypeSubscription,
        Debug:            true, // Enable debug for development
    }
    
    client, err := client.NewClaudeCodeClient(ctx, config)
    if err != nil {
        return fmt.Errorf("failed to create client: %w", err)
    }
    defer client.Close()
    
    // Test subscription functionality
    request := &types.QueryRequest{
        Messages: []types.Message{
            {Role: types.RoleUser, Content: "Test subscription authentication"},
        },
    }
    
    response, err := client.Query(ctx, request)
    if err != nil {
        return fmt.Errorf("subscription test failed: %w", err)
    }
    
    fmt.Printf("✅ Subscription authentication working\n")
    fmt.Printf("Response preview: %s\n", 
        extractFirstTextContent(response.Content)[:100])
    
    return nil
}
```

Shows integration of subscription auth in development workflows.

## Prerequisites

1. **Go 1.21+** installed
2. **Claude Code CLI** installed: `npm install -g @anthropic/claude-code`
3. **Active Claude subscription** (required for subscription auth)
4. **Claude Code CLI authentication** configured

## Setup Instructions

### 1. Install Claude Code CLI
```bash
# Install via npm
npm install -g @anthropic/claude-code

# Verify installation
claude --version
```

### 2. Setup Subscription Authentication
```bash
# Initialize subscription authentication
claude setup-token

# Follow the prompts to authenticate with your Claude account
# This will open a browser window for authentication
```

### 3. Verify Subscription Status
```bash
# Check authentication status
claude auth status

# View subscription details
claude subscription status
```

## Running the Example

### Run the Example
```bash
cd examples/subscription_auth
go run main.go
```

## Expected Output

```
=== Subscription Authentication Example ===

--- Subscription Setup and Validation ---
✅ Subscription authentication is available
✅ Claude Code client created successfully
  Auth method: subscription
  Working directory: /current/path

--- Subscription Status Monitoring ---
Subscription status: true
Subscription details:
  User: user@example.com
  Plan: Pro
  Valid until: 2024-12-31
  Auto-renewal: true
  Usage this month: 1,247 queries
  Monthly limit: 10,000 queries

--- Subscription Test Query ---
Testing subscription authentication...
✅ Subscription query successful
Response preview: Hello! I'm Claude, an AI assistant created by Anthropic. I'm here to help you with...

--- Development Workflow Integration ---
✅ Development environment setup complete
✅ Subscription authentication working
✅ Debug mode enabled for development
✅ All development checks passed

--- Subscription Health Check ---
Performing comprehensive subscription health check...
✅ Authentication token valid
✅ Subscription active
✅ API access confirmed
✅ No rate limiting detected
✅ All systems operational
```

## Key Concepts

### Subscription Authentication Benefits

| Aspect | Subscription Auth | API Key Auth |
|--------|------------------|--------------|
| **Setup** | One-time browser auth | Environment variable management |
| **Security** | Managed by Claude Code CLI | Manual key management |
| **Usage Tracking** | Automatic billing integration | Manual monitoring |
| **Rate Limits** | Subscription-based limits | API key limits |
| **Development** | Seamless local development | Requires key distribution |

### Authentication Flow

1. **Initial Setup**: User authenticates via browser
2. **Token Storage**: CLI securely stores authentication tokens
3. **Automatic Refresh**: Tokens refreshed automatically
4. **Session Management**: Each client session uses stored tokens
5. **Validation**: Real-time validation of subscription status

### Subscription Management

```go
type SubscriptionInfo struct {
    User        string    `json:"user"`
    Plan        string    `json:"plan"`
    ExpiresAt   time.Time `json:"expires_at"`
    AutoRenewal bool      `json:"auto_renewal"`
    Usage       UsageInfo `json:"usage"`
}

type UsageInfo struct {
    QueriesThisMonth int `json:"queries_this_month"`
    MonthlyLimit     int `json:"monthly_limit"`
    ResetDate        time.Time `json:"reset_date"`
}
```

## Advanced Usage

### Subscription Health Monitoring
```go
type SubscriptionMonitor struct {
    client          *client.ClaudeCodeClient
    checkInterval   time.Duration
    alertThreshold  float64
    notificationFn  func(alert Alert)
}

func (sm *SubscriptionMonitor) StartMonitoring(ctx context.Context) error {
    ticker := time.NewTicker(sm.checkInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := sm.performHealthCheck(); err != nil {
                log.Printf("Health check failed: %v", err)
            }
        }
    }
}

func (sm *SubscriptionMonitor) performHealthCheck() error {
    subscriptionAuth := &types.SubscriptionAuth{}
    
    // Check basic validity
    if !subscriptionAuth.IsValid(context.Background()) {
        sm.notificationFn(Alert{
            Type:    "subscription_invalid",
            Message: "Subscription authentication is no longer valid",
            Severity: "critical",
        })
        return fmt.Errorf("subscription invalid")
    }
    
    // Check usage limits
    info, err := subscriptionAuth.GetSubscriptionInfo(context.Background())
    if err != nil {
        return fmt.Errorf("failed to get subscription info: %w", err)
    }
    
    usagePercent := float64(info.Usage.QueriesThisMonth) / float64(info.Usage.MonthlyLimit)
    if usagePercent > sm.alertThreshold {
        sm.notificationFn(Alert{
            Type:    "usage_high",
            Message: fmt.Sprintf("Usage at %.1f%% of monthly limit", usagePercent*100),
            Severity: "warning",
        })
    }
    
    // Check expiration
    if time.Until(info.ExpiresAt) < 7*24*time.Hour {
        sm.notificationFn(Alert{
            Type:    "expiring_soon",
            Message: "Subscription expires within 7 days",
            Severity: "warning", 
        })
    }
    
    return nil
}
```

### Development Team Integration
```go
func setupTeamDevelopment(ctx context.Context, teamConfig TeamConfig) error {
    // Validate team subscription requirements
    subscriptionAuth := &types.SubscriptionAuth{}
    if !subscriptionAuth.IsValid(ctx) {
        return fmt.Errorf("team development requires valid subscription")
    }
    
    info, err := subscriptionAuth.GetSubscriptionInfo(ctx)
    if err != nil {
        return fmt.Errorf("failed to get subscription info: %w", err)
    }
    
    // Check if subscription supports team usage
    if !supportsTeamUsage(info.Plan) {
        return fmt.Errorf("subscription plan %s does not support team usage", info.Plan)
    }
    
    // Setup team-specific configuration
    config := &types.ClaudeCodeConfig{
        WorkingDirectory: teamConfig.ProjectPath,
        AuthMethod:       types.AuthTypeSubscription,
        Environment: map[string]string{
            "TEAM_ID":      teamConfig.TeamID,
            "PROJECT_NAME": teamConfig.ProjectName,
        },
    }
    
    client, err := client.NewClaudeCodeClient(ctx, config)
    if err != nil {
        return fmt.Errorf("failed to create team client: %w", err)
    }
    defer client.Close()
    
    fmt.Printf("✅ Team development setup complete for %s\n", teamConfig.TeamID)
    return nil
}

type TeamConfig struct {
    TeamID      string
    ProjectName string
    ProjectPath string
    Members     []string
}

func supportsTeamUsage(plan string) bool {
    teamPlans := []string{"Pro", "Team", "Enterprise"}
    for _, teamPlan := range teamPlans {
        if plan == teamPlan {
            return true
        }
    }
    return false
}
```

### Subscription Analytics
```go
type SubscriptionAnalytics struct {
    client     *client.ClaudeCodeClient
    startTime  time.Time
    queryCount int
    errors     []error
}

func (sa *SubscriptionAnalytics) TrackQuery(request *types.QueryRequest) {
    sa.queryCount++
    
    // Track usage patterns
    log.Printf("Query %d: %d messages, %d tokens estimated", 
        sa.queryCount, len(request.Messages), estimateTokens(request))
}

func (sa *SubscriptionAnalytics) TrackError(err error) {
    sa.errors = append(sa.errors, err)
    
    // Analyze error patterns
    if isSubscriptionError(err) {
        log.Printf("Subscription-related error: %v", err)
    }
}

func (sa *SubscriptionAnalytics) GenerateReport() SubscriptionReport {
    return SubscriptionReport{
        Duration:    time.Since(sa.startTime),
        QueryCount:  sa.queryCount,
        ErrorCount:  len(sa.errors),
        ErrorRate:   float64(len(sa.errors)) / float64(sa.queryCount),
        QueriesPerHour: float64(sa.queryCount) / time.Since(sa.startTime).Hours(),
    }
}

type SubscriptionReport struct {
    Duration       time.Duration
    QueryCount     int
    ErrorCount     int
    ErrorRate      float64
    QueriesPerHour float64
}
```

## Best Practices

### 1. Subscription Validation
```go
func createClientWithValidation(ctx context.Context) (*client.ClaudeCodeClient, error) {
    // Always validate subscription before client creation
    subscriptionAuth := &types.SubscriptionAuth{}
    if !subscriptionAuth.IsValid(ctx) {
        return nil, fmt.Errorf("subscription authentication required")
    }
    
    // Check subscription health
    info, err := subscriptionAuth.GetSubscriptionInfo(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to validate subscription: %w", err)
    }
    
    // Warn if near limits
    usagePercent := float64(info.Usage.QueriesThisMonth) / float64(info.Usage.MonthlyLimit)
    if usagePercent > 0.9 {
        log.Printf("⚠️  Warning: Using %.1f%% of monthly query limit", usagePercent*100)
    }
    
    config := &types.ClaudeCodeConfig{
        AuthMethod: types.AuthTypeSubscription,
    }
    
    return client.NewClaudeCodeClient(ctx, config)
}
```

### 2. Error Handling
```go
func handleSubscriptionErrors(err error) error {
    if err == nil {
        return nil
    }
    
    // Check for subscription-specific errors
    if isSubscriptionExpiredError(err) {
        return fmt.Errorf("subscription has expired, please renew: %w", err)
    }
    
    if isSubscriptionLimitExceededError(err) {
        return fmt.Errorf("monthly query limit exceeded, upgrade plan or wait for reset: %w", err)
    }
    
    if isSubscriptionInvalidError(err) {
        return fmt.Errorf("subscription authentication invalid, please re-authenticate: %w", err)
    }
    
    return err
}

func isSubscriptionExpiredError(err error) bool {
    return strings.Contains(err.Error(), "subscription expired")
}

func isSubscriptionLimitExceededError(err error) bool {
    return strings.Contains(err.Error(), "quota exceeded") ||
           strings.Contains(err.Error(), "limit exceeded")
}

func isSubscriptionInvalidError(err error) bool {
    return strings.Contains(err.Error(), "invalid subscription") ||
           strings.Contains(err.Error(), "authentication failed")
}
```

### 3. Development Workflow Integration
```go
func setupDevelopmentWorkflow(ctx context.Context, projectPath string) error {
    // Validate development environment
    if err := validateDevelopmentEnvironment(); err != nil {
        return fmt.Errorf("development environment validation failed: %w", err)
    }
    
    // Setup subscription-based client
    client, err := createClientWithValidation(ctx)
    if err != nil {
        return fmt.Errorf("failed to create subscription client: %w", err)
    }
    defer client.Close()
    
    // Configure for development
    if err := setupDevelopmentMCPServers(client, projectPath); err != nil {
        log.Printf("Warning: MCP setup failed: %v", err)
    }
    
    // Test development setup
    if err := testDevelopmentSetup(ctx, client); err != nil {
        return fmt.Errorf("development setup test failed: %w", err)
    }
    
    fmt.Println("✅ Development workflow setup complete")
    return nil
}

func validateDevelopmentEnvironment() error {
    // Check Claude Code CLI installation
    if _, err := exec.LookPath("claude"); err != nil {
        return fmt.Errorf("claude CLI not found: %w", err)
    }
    
    // Check subscription status
    subscriptionAuth := &types.SubscriptionAuth{}
    if !subscriptionAuth.IsValid(context.Background()) {
        return fmt.Errorf("valid subscription required for development")
    }
    
    return nil
}
```

## Troubleshooting

### Common Issues

1. **Subscription Not Found**
   ```bash
   # Re-authenticate
   claude setup-token
   
   # Check status
   claude auth status
   ```

2. **Authentication Expired**
   ```bash
   # Refresh authentication
   claude auth refresh
   
   # Or re-authenticate completely
   claude setup-token
   ```

3. **Usage Limits Exceeded**
   ```bash
   # Check usage status
   claude subscription status
   
   # Upgrade plan if needed
   claude subscription upgrade
   ```

### Debug Subscription Issues
```go
func debugSubscriptionAuth(ctx context.Context) {
    subscriptionAuth := &types.SubscriptionAuth{}
    
    // Basic validation
    isValid := subscriptionAuth.IsValid(ctx)
    fmt.Printf("Subscription valid: %v\n", isValid)
    
    if !isValid {
        fmt.Println("❌ Subscription authentication issues detected")
        fmt.Println("💡 Try: claude setup-token")
        return
    }
    
    // Detailed information
    info, err := subscriptionAuth.GetSubscriptionInfo(ctx)
    if err != nil {
        fmt.Printf("❌ Failed to get subscription info: %v\n", err)
        fmt.Println("💡 Try: claude auth refresh")
        return
    }
    
    // Display debug information
    fmt.Printf("Debug Information:\n")
    fmt.Printf("  User: %s\n", info.User)
    fmt.Printf("  Plan: %s\n", info.Plan)
    fmt.Printf("  Expires: %s\n", info.ExpiresAt.Format("2006-01-02 15:04:05"))
    fmt.Printf("  Usage: %d/%d queries this month\n", 
        info.Usage.QueriesThisMonth, info.Usage.MonthlyLimit)
    
    // Health checks
    if time.Until(info.ExpiresAt) < 0 {
        fmt.Println("⚠️  Subscription has expired")
    }
    
    usagePercent := float64(info.Usage.QueriesThisMonth) / float64(info.Usage.MonthlyLimit)
    if usagePercent > 0.8 {
        fmt.Printf("⚠️  High usage: %.1f%% of monthly limit\n", usagePercent*100)
    }
}
```

### Subscription Recovery
```go
func recoverSubscriptionAuth(ctx context.Context) error {
    subscriptionAuth := &types.SubscriptionAuth{}
    
    // Try standard validation first
    if subscriptionAuth.IsValid(ctx) {
        return nil // Already working
    }
    
    fmt.Println("🔄 Attempting subscription recovery...")
    
    // Step 1: Try refreshing authentication
    if err := refreshAuthentication(); err != nil {
        log.Printf("Refresh failed: %v", err)
    } else {
        if subscriptionAuth.IsValid(ctx) {
            fmt.Println("✅ Subscription recovered via refresh")
            return nil
        }
    }
    
    // Step 2: Clear and re-authenticate
    fmt.Println("💡 Please re-authenticate manually:")
    fmt.Println("   claude setup-token")
    
    return fmt.Errorf("manual re-authentication required")
}

func refreshAuthentication() error {
    cmd := exec.Command("claude", "auth", "refresh")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("auth refresh failed: %w (output: %s)", err, output)
    }
    return nil
}
```

## Performance Optimization

### 1. Subscription Usage Optimization
```go
type UsageOptimizer struct {
    client        *client.ClaudeCodeClient
    monthlyLimit  int
    usageTarget   float64
    queryHistory  []QueryMetrics
}

func (uo *UsageOptimizer) ShouldExecuteQuery(request *types.QueryRequest) bool {
    // Check current usage
    info, err := uo.getSubscriptionInfo()
    if err != nil {
        log.Printf("Failed to get usage info: %v", err)
        return true // Default to allowing
    }
    
    currentUsage := float64(info.Usage.QueriesThisMonth) / float64(info.Usage.MonthlyLimit)
    
    // If we're near the limit, be more selective
    if currentUsage > uo.usageTarget {
        return uo.isHighPriorityQuery(request)
    }
    
    return true
}

func (uo *UsageOptimizer) isHighPriorityQuery(request *types.QueryRequest) bool {
    // Implement logic to determine query priority
    // E.g., based on message content, user context, etc.
    return true // Simplified
}
```

### 2. Query Batching
```go
func batchQueries(queries []string, client *client.ClaudeCodeClient) error {
    // Combine multiple queries into one to save on usage
    batchedContent := "Please answer the following questions:\n\n"
    for i, query := range queries {
        batchedContent += fmt.Sprintf("%d. %s\n", i+1, query)
    }
    
    request := &types.QueryRequest{
        Messages: []types.Message{
            {Role: types.RoleUser, Content: batchedContent},
        },
    }
    
    response, err := client.Query(context.Background(), request)
    if err != nil {
        return fmt.Errorf("batched query failed: %w", err)
    }
    
    // Process batched response
    content := extractFirstTextContent(response.Content)
    processBatchedResponse(content, len(queries))
    
    return nil
}
```

## Security Considerations

### 1. Subscription Token Security
- **Never log** authentication tokens or subscription details
- **Secure storage** of subscription credentials by Claude Code CLI
- **Regular rotation** of authentication tokens
- **Access control** for development environments

### 2. Usage Monitoring
- **Track usage patterns** to detect anomalies
- **Monitor for unauthorized access** to subscription
- **Set up alerts** for unusual usage spikes
- **Regular audits** of subscription access

## Next Steps

After mastering subscription authentication, explore:
- [Auth Methods](../auth_methods/) - Compare auth methods
- [Basic Client](../basic_client/) - Client configuration
- [Sync Queries](../sync_queries/) - Making authenticated queries
- [Advanced Client](../advanced_client/) - Advanced features

## Related Documentation

- [Subscription Types](../../pkg/types/auth.go)
- [Client Configuration](../../pkg/client/)
- [Authentication Guide](../../docs/authentication.md)