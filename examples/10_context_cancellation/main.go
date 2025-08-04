package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/claudecode"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Context Cancellation Examples ===\n")

	// Example 1: Basic timeout
	example1BasicTimeout()

	// Example 2: Manual cancellation
	example2ManualCancellation()

	// Example 3: Deadline context
	example3DeadlineContext()

	// Example 4: Cascading cancellation
	example4CascadingCancellation()

	// Example 5: Context with values
	example5ContextWithValues()

	// Example 6: Advanced cancellation patterns
	example6AdvancedPatterns()
}

func example1BasicTimeout() {
	fmt.Println("Example 1: Basic Timeout")
	fmt.Println("------------------------")

	// Different timeout scenarios
	timeoutTests := []struct {
		name     string
		timeout  time.Duration
		query    string
		expected string
	}{
		{
			name:     "Very short timeout",
			timeout:  100 * time.Millisecond,
			query:    "Count to 1000",
			expected: "Should timeout quickly",
		},
		{
			name:     "Medium timeout",
			timeout:  3 * time.Second,
			query:    "Explain quantum physics in detail",
			expected: "Might partially complete",
		},
		{
			name:     "Long timeout",
			timeout:  10 * time.Second,
			query:    "What is 2+2?",
			expected: "Should complete easily",
		},
	}

	for _, test := range timeoutTests {
		fmt.Printf("\n🔹 %s (timeout: %v)\n", test.name, test.timeout)
		fmt.Printf("   Query: %s\n", test.query)
		fmt.Printf("   Expected: %s\n", test.expected)

		ctx, cancel := context.WithTimeout(context.Background(), test.timeout)
		defer cancel()

		start := time.Now()
		msgChan := claudecode.Query(ctx, test.query, nil)

		completed := false
		messageCount := 0

		for msg := range msgChan {
			messageCount++
			
			select {
			case <-ctx.Done():
				elapsed := time.Since(start)
				fmt.Printf("   ⏱️ Timed out after %v (%d messages received)\n", 
					elapsed.Round(time.Millisecond), messageCount)
				goto next
			default:
				if _, ok := msg.(*claudecode.ResultMessage); ok {
					completed = true
					elapsed := time.Since(start)
					fmt.Printf("   ✅ Completed in %v\n", elapsed.Round(time.Millisecond))
				}
			}
		}

		if !completed && ctx.Err() == nil {
			fmt.Println("   ❓ Channel closed without timeout or completion")
		}

	next:
		// Small delay between tests
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println()
}

func example2ManualCancellation() {
	fmt.Println("Example 2: Manual Cancellation")
	fmt.Println("------------------------------")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Track progress
	var progressMu sync.Mutex
	progressSteps := []string{}
	cancelled := false

	// Message processor
	go func() {
		for msg := range client.ReceiveMessages() {
			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range m.Content {
					if textBlock, ok := block.(claudecode.TextBlock); ok {
						progressMu.Lock()
						progressSteps = append(progressSteps, 
							fmt.Sprintf("Step %d: %.50s...", len(progressSteps)+1, textBlock.Text))
						progressMu.Unlock()
						
						// Cancel after 3 steps
						if len(progressSteps) == 3 && !cancelled {
							cancelled = true
							fmt.Println("\n🛑 Cancelling after 3 steps...")
							cancel()
						}
					}
				}
			case *claudecode.ResultMessage:
				if !cancelled {
					fmt.Printf("✅ Completed naturally (%d turns)\n", m.NumTurns)
				}
			}
		}
	}()

	// Start a multi-step task
	fmt.Println("Starting multi-step task...")
	query := `Create a step-by-step plan to build a web application:
1. Choose technology stack
2. Design database schema  
3. Create API endpoints
4. Build frontend
5. Add authentication
6. Deploy to production`

	if err := client.Query(ctx, query, "manual-cancel"); err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Println("✅ Query cancelled successfully")
		} else {
			log.Printf("Query error: %v\n", err)
		}
	}

	time.Sleep(3 * time.Second)

	// Show progress
	progressMu.Lock()
	fmt.Printf("\n📊 Progress before cancellation:\n")
	for _, step := range progressSteps {
		fmt.Printf("   %s\n", step)
	}
	fmt.Printf("   Total steps completed: %d\n", len(progressSteps))
	progressMu.Unlock()
	fmt.Println()
}

func example3DeadlineContext() {
	fmt.Println("Example 3: Deadline Context")
	fmt.Println("---------------------------")

	// Test different deadline scenarios
	now := time.Now()
	deadlineTests := []struct {
		name     string
		deadline time.Time
		query    string
	}{
		{
			name:     "Past deadline",
			deadline: now.Add(-1 * time.Second),
			query:    "This should fail immediately",
		},
		{
			name:     "Very near deadline",
			deadline: now.Add(500 * time.Millisecond),
			query:    "Quick response needed",
		},
		{
			name:     "Comfortable deadline",
			deadline: now.Add(5 * time.Second),
			query:    "Explain how contexts work in Go",
		},
	}

	for _, test := range deadlineTests {
		fmt.Printf("\n🕐 %s\n", test.name)
		timeUntilDeadline := time.Until(test.deadline)
		fmt.Printf("   Deadline: %v from now\n", timeUntilDeadline.Round(time.Millisecond))
		fmt.Printf("   Query: %s\n", test.query)

		ctx, cancel := context.WithDeadline(context.Background(), test.deadline)
		defer cancel()

		// Check if already exceeded
		if ctx.Err() != nil {
			fmt.Printf("   ❌ Deadline already exceeded: %v\n", ctx.Err())
			continue
		}

		// Create client with deadline
		client := claudecode.NewClaudeSDKClient(nil)
		defer client.Close()

		// Track timing
		connectStart := time.Now()
		
		// Try to connect
		if err := client.Connect(ctx); err != nil {
			connectDuration := time.Since(connectStart)
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("   ⏰ Deadline exceeded during connection (after %v)\n", 
					connectDuration.Round(time.Millisecond))
			} else {
				fmt.Printf("   ❌ Connection error: %v\n", err)
			}
			continue
		}

		// Process messages
		responseReceived := false
		go func() {
			for msg := range client.ReceiveMessages() {
				if _, ok := msg.(*claudecode.AssistantMessage); ok && !responseReceived {
					responseReceived = true
					remaining := time.Until(test.deadline)
					fmt.Printf("   ✅ Response received with %v remaining\n", 
						remaining.Round(time.Millisecond))
				}
			}
		}()

		// Send query
		queryStart := time.Now()
		if err := client.Query(ctx, test.query, "deadline-test"); err != nil {
			queryDuration := time.Since(queryStart)
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("   ⏰ Deadline exceeded during query (after %v)\n", 
					queryDuration.Round(time.Millisecond))
			} else {
				fmt.Printf("   ❌ Query error: %v\n", err)
			}
		}

		// Wait a bit for response
		time.Sleep(1 * time.Second)
		
		if !responseReceived && ctx.Err() == context.DeadlineExceeded {
			fmt.Println("   ⏰ Deadline exceeded while waiting for response")
		}
	}
	fmt.Println()
}

func example4CascadingCancellation() {
	fmt.Println("Example 4: Cascading Cancellation")
	fmt.Println("---------------------------------")

	// Create parent context
	parentCtx, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	// Track cancellations
	var cancelledOps int32
	var wg sync.WaitGroup

	// Create child operations
	operations := []struct {
		name  string
		delay time.Duration
	}{
		{"Database Query", 2 * time.Second},
		{"API Call", 3 * time.Second},
		{"File Processing", 4 * time.Second},
		{"Cache Update", 1 * time.Second},
		{"Notification Send", 2 * time.Second},
	}

	fmt.Println("Starting 5 parallel operations with cascading cancellation...")
	startTime := time.Now()

	for _, op := range operations {
		wg.Add(1)
		go func(operation struct {
			name  string
			delay time.Duration
		}) {
			defer wg.Done()

			// Create child context
			childCtx, childCancel := context.WithCancel(parentCtx)
			defer childCancel()

			// Simulate operation with Claude query
			fmt.Printf("🔄 %s: Starting...\n", operation.name)
			
			msgChan := claudecode.Query(childCtx, 
				fmt.Sprintf("Simulate %s operation", operation.name), nil)

			// Wait for completion or cancellation
			timer := time.NewTimer(operation.delay)
			defer timer.Stop()

			select {
			case <-timer.C:
				fmt.Printf("✅ %s: Completed after %v\n", 
					operation.name, time.Since(startTime).Round(time.Millisecond))
			case <-childCtx.Done():
				atomic.AddInt32(&cancelledOps, 1)
				fmt.Printf("❌ %s: Cancelled after %v\n", 
					operation.name, time.Since(startTime).Round(time.Millisecond))
			case msg := <-msgChan:
				if msg != nil {
					fmt.Printf("📨 %s: Got response after %v\n", 
						operation.name, time.Since(startTime).Round(time.Millisecond))
				}
			}
		}(op)
	}

	// Cancel parent after 2.5 seconds
	go func() {
		time.Sleep(2500 * time.Millisecond)
		fmt.Println("\n🛑 Cancelling parent context...")
		parentCancel()
	}()

	// Wait for all operations
	wg.Wait()

	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("   Total operations: %d\n", len(operations))
	fmt.Printf("   Cancelled operations: %d\n", atomic.LoadInt32(&cancelledOps))
	fmt.Printf("   Completed operations: %d\n", len(operations)-int(atomic.LoadInt32(&cancelledOps)))
	fmt.Println()
}

func example5ContextWithValues() {
	fmt.Println("Example 5: Context with Values")
	fmt.Println("------------------------------")

	// Define context keys
	type contextKey string
	const (
		userIDKey     contextKey = "userID"
		sessionIDKey  contextKey = "sessionID"
		priorityKey   contextKey = "priority"
		debugModeKey  contextKey = "debugMode"
	)

	// Create base context with values
	ctx := context.Background()
	ctx = context.WithValue(ctx, userIDKey, "user-12345")
	ctx = context.WithValue(ctx, sessionIDKey, "session-abcdef")
	ctx = context.WithValue(ctx, priorityKey, "high")
	ctx = context.WithValue(ctx, debugModeKey, true)

	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	fmt.Println("Context values:")
	fmt.Printf("   User ID: %v\n", ctx.Value(userIDKey))
	fmt.Printf("   Session ID: %v\n", ctx.Value(sessionIDKey))
	fmt.Printf("   Priority: %v\n", ctx.Value(priorityKey))
	fmt.Printf("   Debug Mode: %v\n", ctx.Value(debugModeKey))

	// Create custom query based on context
	userID := ctx.Value(userIDKey).(string)
	priority := ctx.Value(priorityKey).(string)
	debugMode := ctx.Value(debugModeKey).(bool)

	query := fmt.Sprintf("Process request for user %s with %s priority", userID, priority)
	if debugMode {
		query += " (include debug information)"
	}

	fmt.Printf("\nGenerated query: %s\n", query)

	// Use context values to configure options
	options := claudecode.NewClaudeCodeOptions()
	if priority == "high" {
		options.MaxTurns = claudecode.IntPtr(10)
		fmt.Println("📈 High priority: Increased max turns to 10")
	}

	// Execute query with context
	msgChan := claudecode.Query(ctx, query, options)

	for msg := range msgChan {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			// Check if debug mode for verbose output
			if debugMode {
				fmt.Println("\n[DEBUG] Full response:")
				for _, block := range m.Content {
					if textBlock, ok := block.(claudecode.TextBlock); ok {
						fmt.Printf("[DEBUG] %s\n", textBlock.Text)
					}
				}
			} else {
				fmt.Println("\n✅ Response received (debug mode off)")
			}
		case *claudecode.ResultMessage:
			// Log with session ID
			sessionID := ctx.Value(sessionIDKey).(string)
			fmt.Printf("\n📊 Session %s completed:\n", sessionID)
			fmt.Printf("   Duration: %dms\n", m.DurationMs)
			if m.TotalCostUSD != nil {
				fmt.Printf("   Cost: $%.6f\n", *m.TotalCostUSD)
			}
		}
	}
	fmt.Println()
}

func example6AdvancedPatterns() {
	fmt.Println("Example 6: Advanced Cancellation Patterns")
	fmt.Println("-----------------------------------------")

	// Pattern 1: Graceful shutdown with signal handling
	fmt.Println("\n🎯 Pattern 1: Graceful Shutdown")
	demonstrateGracefulShutdown()

	// Pattern 2: Racing contexts
	fmt.Println("\n🎯 Pattern 2: Racing Contexts")
	demonstrateRacingContexts()

	// Pattern 3: Context merger
	fmt.Println("\n🎯 Pattern 3: Context Merger")
	demonstrateContextMerger()

	// Pattern 4: Retry with backoff
	fmt.Println("\n🎯 Pattern 4: Retry with Backoff")
	demonstrateRetryWithBackoff()
}

func demonstrateGracefulShutdown() {
	// Create shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Simulate signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	defer signal.Stop(sigChan)

	// Shutdown handler
	go func() {
		select {
		case <-sigChan:
			fmt.Println("\n📢 Shutdown signal received (simulated)")
			cancel()
		case <-time.After(3 * time.Second):
			fmt.Println("\n⏰ Simulating shutdown after 3 seconds")
			cancel()
		}
	}()

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		log.Printf("Connection failed: %v\n", err)
		return
	}

	// Graceful shutdown handler
	var shutdownOnce sync.Once
	handleShutdown := func() {
		shutdownOnce.Do(func() {
			fmt.Println("🔄 Performing graceful shutdown...")
			
			// Save state
			fmt.Println("   💾 Saving state...")
			time.Sleep(500 * time.Millisecond)
			
			// Close connections
			fmt.Println("   🔌 Closing connections...")
			client.Close()
			
			// Final cleanup
			fmt.Println("   🧹 Cleanup complete")
		})
	}

	// Process with shutdown handling
	done := make(chan bool)
	go func() {
		defer close(done)
		defer handleShutdown()
		
		for {
			select {
			case <-ctx.Done():
				fmt.Println("   ⚠️ Context cancelled, initiating shutdown")
				return
			default:
				// Simulate work
				fmt.Print(".")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	<-done
	fmt.Println("\n✅ Graceful shutdown complete")
}

func demonstrateRacingContexts() {
	// Create multiple contexts that race
	timeout1 := 2 * time.Second
	timeout2 := 3 * time.Second
	timeout3 := 4 * time.Second

	ctx1, cancel1 := context.WithTimeout(context.Background(), timeout1)
	defer cancel1()

	ctx2, cancel2 := context.WithTimeout(context.Background(), timeout2)
	defer cancel2()

	ctx3, cancel3 := context.WithTimeout(context.Background(), timeout3)
	defer cancel3()

	fmt.Printf("Racing 3 contexts: %v, %v, %v\n", timeout1, timeout2, timeout3)

	// Race contexts
	start := time.Now()
	
	select {
	case <-ctx1.Done():
		elapsed := time.Since(start)
		fmt.Printf("🏁 Context 1 won after %v\n", elapsed.Round(time.Millisecond))
	case <-ctx2.Done():
		elapsed := time.Since(start)
		fmt.Printf("🏁 Context 2 won after %v\n", elapsed.Round(time.Millisecond))
	case <-ctx3.Done():
		elapsed := time.Since(start)
		fmt.Printf("🏁 Context 3 won after %v\n", elapsed.Round(time.Millisecond))
	}
}

func demonstrateContextMerger() {
	// Create multiple parent contexts
	parent1, cancel1 := context.WithCancel(context.Background())
	parent2, cancel2 := context.WithCancel(context.Background())
	defer cancel1()
	defer cancel2()

	// Merge contexts
	mergedCtx, mergedCancel := mergeContexts(parent1, parent2)
	defer mergedCancel()

	fmt.Println("Created merged context from 2 parents")

	// Test cancellation propagation
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("   Cancelling parent 1...")
		cancel1()
	}()

	select {
	case <-mergedCtx.Done():
		fmt.Println("✅ Merged context cancelled when parent 1 was cancelled")
	case <-time.After(2 * time.Second):
		fmt.Println("❌ Merged context not cancelled (unexpected)")
	}
}

func demonstrateRetryWithBackoff() {
	ctx := context.Background()
	
	// Retry configuration
	maxRetries := 3
	baseDelay := 500 * time.Millisecond
	
	// Simulate operation that might fail
	attemptQuery := func(attempt int) error {
		// Create timeout for this attempt
		attemptCtx, cancel := context.WithTimeout(ctx, time.Duration(attempt)*time.Second)
		defer cancel()
		
		fmt.Printf("   Attempt %d (timeout: %v)\n", attempt, time.Duration(attempt)*time.Second)
		
		// Simulate failure on first attempts
		if attempt < 3 {
			time.Sleep(time.Duration(attempt+1) * time.Second) // Will timeout
			return context.DeadlineExceeded
		}
		
		// Success on last attempt
		msgChan := claudecode.Query(attemptCtx, "What is 2+2?", nil)
		for msg := range msgChan {
			if _, ok := msg.(*claudecode.ResultMessage); ok {
				return nil
			}
		}
		
		return fmt.Errorf("no result received")
	}

	// Execute with exponential backoff
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		lastErr = attemptQuery(attempt)
		
		if lastErr == nil {
			fmt.Printf("✅ Success on attempt %d\n", attempt)
			break
		}
		
		fmt.Printf("   ❌ Attempt %d failed: %v\n", attempt, lastErr)
		
		if attempt < maxRetries {
			delay := baseDelay * time.Duration(1<<(attempt-1)) // Exponential backoff
			fmt.Printf("   ⏳ Waiting %v before retry...\n", delay)
			time.Sleep(delay)
		}
	}
	
	if lastErr != nil {
		fmt.Printf("❌ All %d attempts failed\n", maxRetries)
	}
}

// Helper function to merge contexts
func mergeContexts(ctx1, ctx2 context.Context) (context.Context, context.CancelFunc) {
	merged, cancel := context.WithCancel(context.Background())
	
	go func() {
		select {
		case <-ctx1.Done():
			cancel()
		case <-ctx2.Done():
			cancel()
		case <-merged.Done():
		}
	}()
	
	return merged, cancel
}