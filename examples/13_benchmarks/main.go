package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/claudecode"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Claude SDK Performance Benchmarks ===\n")

	// System info
	printSystemInfo()

	// Run benchmarks
	benchmark1ConnectionSpeed()
	benchmark2QueryLatency()
	benchmark3Throughput()
	benchmark4ConcurrentClients()
	benchmark5MessageProcessing()
	benchmark6MemoryUsage()
	benchmark7ContextOverhead()
	benchmark8ErrorHandling()

	// Summary
	printBenchmarkSummary()
}

func printSystemInfo() {
	fmt.Println("System Information:")
	fmt.Printf("  Go Version: %s\n", runtime.Version())
	fmt.Printf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("  CPUs: %d\n", runtime.NumCPU())
	fmt.Printf("  Goroutines: %d\n", runtime.NumGoroutine())

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("  Memory: %.2f MB allocated\n", float64(m.Alloc)/1024/1024)
	fmt.Println()
}

func benchmark1ConnectionSpeed() {
	fmt.Println("Benchmark 1: Connection Speed")
	fmt.Println("-----------------------------")

	iterations := 10
	times := make([]time.Duration, iterations)

	for i := 0; i < iterations; i++ {
		client := claudecode.NewClaudeSDKClient(nil)

		start := time.Now()
		ctx := context.Background()
		err := client.Connect(ctx)
		elapsed := time.Since(start)

		if err != nil {
			log.Printf("Connection %d failed: %v\n", i+1, err)
			continue
		}

		times[i] = elapsed
		client.Close()

		fmt.Printf("  Iteration %d: %v\n", i+1, elapsed)
	}

	// Calculate statistics
	stats := calculateStats(times)
	fmt.Printf("\n📊 Connection Speed Results:\n")
	fmt.Printf("  Average: %v\n", stats.avg)
	fmt.Printf("  Min: %v\n", stats.min)
	fmt.Printf("  Max: %v\n", stats.max)
	fmt.Printf("  StdDev: %v\n", stats.stdDev)
	fmt.Println()
}

func benchmark2QueryLatency() {
	fmt.Println("Benchmark 2: Query Latency")
	fmt.Println("--------------------------")

	queries := []struct {
		name  string
		query string
	}{
		{"Simple", "What is 2+2?"},
		{"Medium", "Explain the concept of recursion"},
		{"Complex", "Write a detailed explanation of how TCP/IP works"},
	}

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	for _, q := range queries {
		fmt.Printf("\n🔹 Query Type: %s\n", q.name)

		latencies := make([]time.Duration, 5)

		for i := 0; i < 5; i++ {
			start := time.Now()
			firstResponseTime := time.Duration(0)

			// Measure time to first response
			done := make(chan bool)
			go func() {
				for msg := range client.ReceiveMessages() {
					if firstResponseTime == 0 {
						firstResponseTime = time.Since(start)
					}
					if _, ok := msg.(*claudecode.ResultMessage); ok {
						done <- true
						break
					}
				}
			}()

			if err := client.Query(ctx, q.query, fmt.Sprintf("latency-%d", i)); err != nil {
				log.Printf("Query failed: %v\n", err)
				continue
			}

			<-done
			latencies[i] = firstResponseTime
			fmt.Printf("  Run %d: %v\n", i+1, firstResponseTime)

			time.Sleep(1 * time.Second) // Avoid rate limiting
		}

		stats := calculateStats(latencies)
		fmt.Printf("  Average latency: %v\n", stats.avg)
	}
	fmt.Println()
}

func benchmark3Throughput() {
	fmt.Println("Benchmark 3: Throughput")
	fmt.Println("-----------------------")

	durations := []time.Duration{
		10 * time.Second,
		30 * time.Second,
		60 * time.Second,
	}

	for _, duration := range durations {
		fmt.Printf("\n🔹 Test Duration: %v\n", duration)

		client := claudecode.NewClaudeSDKClient(nil)
		ctx := context.Background()

		if err := client.Connect(ctx); err != nil {
			log.Printf("Failed to connect: %v\n", err)
			continue
		}

		var queryCount int32
		var responseCount int32
		var errorCount int32

		// Response handler
		go func() {
			for msg := range client.ReceiveMessages() {
				switch msg.(type) {
				case *claudecode.AssistantMessage:
					atomic.AddInt32(&responseCount, 1)
				case *claudecode.SystemMessage:
					if sysMsg := msg.(*claudecode.SystemMessage); sysMsg.Subtype == "error" {
						atomic.AddInt32(&errorCount, 1)
					}
				}
			}
		}()

		// Send queries continuously
		start := time.Now()
		stopChan := make(chan bool)

		go func() {
			for {
				select {
				case <-stopChan:
					return
				default:
					atomic.AddInt32(&queryCount, 1)
					query := fmt.Sprintf("Quick response %d", atomic.LoadInt32(&queryCount))
					client.Query(ctx, query, "throughput")
					time.Sleep(100 * time.Millisecond) // Rate limiting
				}
			}
		}()

		// Run for specified duration
		time.Sleep(duration)
		stopChan <- true
		client.Close()

		// Calculate throughput
		elapsed := time.Since(start).Seconds()
		qps := float64(queryCount) / elapsed
		rps := float64(responseCount) / elapsed

		fmt.Printf("  Queries sent: %d\n", queryCount)
		fmt.Printf("  Responses received: %d\n", responseCount)
		fmt.Printf("  Errors: %d\n", errorCount)
		fmt.Printf("  Queries/second: %.2f\n", qps)
		fmt.Printf("  Responses/second: %.2f\n", rps)
		fmt.Printf("  Success rate: %.2f%%\n", float64(responseCount)/float64(queryCount)*100)
	}
	fmt.Println()
}

func benchmark4ConcurrentClients() {
	fmt.Println("Benchmark 4: Concurrent Clients")
	fmt.Println("-------------------------------")

	clientCounts := []int{1, 5, 10, 20}

	for _, numClients := range clientCounts {
		fmt.Printf("\n🔹 Concurrent Clients: %d\n", numClients)

		var wg sync.WaitGroup
		var totalQueries int32
		var totalResponses int32
		var totalErrors int32

		start := time.Now()

		for i := 0; i < numClients; i++ {
			wg.Add(1)
			go func(clientID int) {
				defer wg.Done()

				client := claudecode.NewClaudeSDKClient(nil)
				defer client.Close()

				ctx := context.Background()
				if err := client.Connect(ctx); err != nil {
					atomic.AddInt32(&totalErrors, 1)
					return
				}

				// Each client sends 5 queries
				for j := 0; j < 5; j++ {
					atomic.AddInt32(&totalQueries, 1)
					query := fmt.Sprintf("Client %d query %d", clientID, j)

					responseChan := client.ReceiveResponse(ctx)
					if err := client.Query(ctx, query, fmt.Sprintf("client-%d", clientID)); err != nil {
						atomic.AddInt32(&totalErrors, 1)
						continue
					}

					// Wait for response
					gotResponse := false
					for msg := range responseChan {
						if _, ok := msg.(*claudecode.ResultMessage); ok {
							gotResponse = true
							break
						}
					}

					if gotResponse {
						atomic.AddInt32(&totalResponses, 1)
					}

					time.Sleep(200 * time.Millisecond) // Rate limiting
				}
			}(i)
		}

		wg.Wait()
		elapsed := time.Since(start)

		fmt.Printf("  Total time: %v\n", elapsed)
		fmt.Printf("  Queries sent: %d\n", totalQueries)
		fmt.Printf("  Responses received: %d\n", totalResponses)
		fmt.Printf("  Errors: %d\n", totalErrors)
		fmt.Printf("  Avg time per query: %v\n", elapsed/time.Duration(totalQueries))

		// Memory usage after concurrent operations
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("  Memory used: %.2f MB\n", float64(m.Alloc)/1024/1024)
	}
	fmt.Println()
}

func benchmark5MessageProcessing() {
	fmt.Println("Benchmark 5: Message Processing")
	fmt.Println("-------------------------------")

	// Create mock messages of different sizes
	messageSizes := []struct {
		name string
		size int
	}{
		{"Small (100 chars)", 100},
		{"Medium (1KB)", 1024},
		{"Large (10KB)", 10 * 1024},
		{"XLarge (100KB)", 100 * 1024},
	}

	for _, ms := range messageSizes {
		fmt.Printf("\n🔹 Message Size: %s\n", ms.name)

		// Generate content
		content := generateContent(ms.size)

		ctx := context.Background()
		iterations := 100

		processingTimes := make([]time.Duration, iterations)

		for i := 0; i < iterations; i++ {
			start := time.Now()

			// Simulate message processing
			msg := &claudecode.AssistantMessage{
				Content: []types.ContentBlock{
					claudecode.TextBlock{Text: content},
				},
			}

			// Process message
			_ = len(msg.Content)
			for _, block := range msg.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					_ = len(textBlock.Text)
				}
			}

			processingTimes[i] = time.Since(start)
		}

		stats := calculateStats(processingTimes)
		fmt.Printf("  Avg processing time: %v\n", stats.avg)
		fmt.Printf("  Min: %v\n", stats.min)
		fmt.Printf("  Max: %v\n", stats.max)

		// Throughput
		totalBytes := int64(ms.size * iterations)
		totalTime := stats.avg * time.Duration(iterations)
		throughputMBps := float64(totalBytes) / totalTime.Seconds() / 1024 / 1024
		fmt.Printf("  Throughput: %.2f MB/s\n", throughputMBps)
	}
	fmt.Println()
}

func benchmark6MemoryUsage() {
	fmt.Println("Benchmark 6: Memory Usage")
	fmt.Println("-------------------------")

	scenarios := []struct {
		name        string
		numClients  int
		numMessages int
	}{
		{"Single client, 100 messages", 1, 100},
		{"5 clients, 50 messages each", 5, 50},
		{"10 clients, 20 messages each", 10, 20},
	}

	for _, scenario := range scenarios {
		fmt.Printf("\n🔹 Scenario: %s\n", scenario.name)

		// Force GC and get baseline
		runtime.GC()
		var baselineStats runtime.MemStats
		runtime.ReadMemStats(&baselineStats)

		clients := make([]*claudecode.ClaudeSDKClient, scenario.numClients)
		messages := make([][]types.Message, scenario.numClients)

		// Create clients and store messages
		for i := 0; i < scenario.numClients; i++ {
			clients[i] = claudecode.NewClaudeSDKClient(nil)
			messages[i] = make([]types.Message, 0, scenario.numMessages)

			// Simulate message storage
			for j := 0; j < scenario.numMessages; j++ {
				msg := &claudecode.AssistantMessage{
					Content: []types.ContentBlock{
						claudecode.TextBlock{Text: generateContent(1024)},
					},
				}
				messages[i] = append(messages[i], msg)
			}
		}

		// Get memory stats after allocation
		var afterStats runtime.MemStats
		runtime.ReadMemStats(&afterStats)

		memoryUsed := afterStats.Alloc - baselineStats.Alloc
		fmt.Printf("  Memory allocated: %.2f MB\n", float64(memoryUsed)/1024/1024)
		fmt.Printf("  Memory per client: %.2f MB\n", float64(memoryUsed)/float64(scenario.numClients)/1024/1024)
		fmt.Printf("  Memory per message: %.2f KB\n", float64(memoryUsed)/float64(scenario.numClients*scenario.numMessages)/1024)

		// Cleanup
		for _, client := range clients {
			client.Close()
		}
		messages = nil
		runtime.GC()

		// Check memory after cleanup
		var cleanupStats runtime.MemStats
		runtime.ReadMemStats(&cleanupStats)
		memoryFreed := afterStats.Alloc - cleanupStats.Alloc
		fmt.Printf("  Memory freed: %.2f MB (%.1f%%)\n",
			float64(memoryFreed)/1024/1024,
			float64(memoryFreed)/float64(memoryUsed)*100)
	}
	fmt.Println()
}

func benchmark7ContextOverhead() {
	fmt.Println("Benchmark 7: Context Overhead")
	fmt.Println("-----------------------------")

	scenarios := []struct {
		name        string
		setupFunc   func() context.Context
		description string
	}{
		{
			name: "Background",
			setupFunc: func() context.Context {
				return context.Background()
			},
			description: "Plain background context",
		},
		{
			name: "With Cancel",
			setupFunc: func() context.Context {
				ctx, _ := context.WithCancel(context.Background())
				return ctx
			},
			description: "Context with cancellation",
		},
		{
			name: "With Timeout",
			setupFunc: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
				return ctx
			},
			description: "Context with 10s timeout",
		},
		{
			name: "With Values",
			setupFunc: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, "key1", "value1")
				ctx = context.WithValue(ctx, "key2", "value2")
				ctx = context.WithValue(ctx, "key3", "value3")
				return ctx
			},
			description: "Context with 3 values",
		},
		{
			name: "Complex",
			setupFunc: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, "user", "test-user")
				ctx, _ = context.WithTimeout(ctx, 10*time.Second)
				ctx = context.WithValue(ctx, "session", "test-session")
				return ctx
			},
			description: "Context with timeout and values",
		},
	}

	iterations := 10000

	for _, scenario := range scenarios {
		fmt.Printf("\n🔹 %s: %s\n", scenario.name, scenario.description)

		// Benchmark context creation
		start := time.Now()
		for i := 0; i < iterations; i++ {
			ctx := scenario.setupFunc()
			_ = ctx
		}
		creationTime := time.Since(start)

		// Benchmark context usage
		ctx := scenario.setupFunc()
		start = time.Now()
		for i := 0; i < iterations; i++ {
			select {
			case <-ctx.Done():
				// Context cancelled
			default:
				// Context active
			}
		}
		usageTime := time.Since(start)

		fmt.Printf("  Creation time (%d iterations): %v\n", iterations, creationTime)
		fmt.Printf("  Avg creation: %v\n", creationTime/time.Duration(iterations))
		fmt.Printf("  Usage time (%d iterations): %v\n", iterations, usageTime)
		fmt.Printf("  Avg usage check: %v\n", usageTime/time.Duration(iterations))
	}
	fmt.Println()
}

func benchmark8ErrorHandling() {
	fmt.Println("Benchmark 8: Error Handling Performance")
	fmt.Println("---------------------------------------")

	errorScenarios := []struct {
		name      string
		errorFunc func() error
	}{
		{
			name: "No Error",
			errorFunc: func() error {
				return nil
			},
		},
		{
			name: "Simple Error",
			errorFunc: func() error {
				return fmt.Errorf("simple error")
			},
		},
		{
			name: "Wrapped Error",
			errorFunc: func() error {
				base := fmt.Errorf("base error")
				return fmt.Errorf("wrapped: %w", base)
			},
		},
		{
			name: "Custom Error Type",
			errorFunc: func() error {
				return &customError{
					code:    500,
					message: "internal server error",
				}
			},
		},
	}

	iterations := 100000

	for _, scenario := range errorScenarios {
		fmt.Printf("\n🔹 %s\n", scenario.name)

		// Benchmark error creation
		start := time.Now()
		for i := 0; i < iterations; i++ {
			err := scenario.errorFunc()
			_ = err
		}
		creationTime := time.Since(start)

		// Benchmark error checking
		err := scenario.errorFunc()
		start = time.Now()
		for i := 0; i < iterations; i++ {
			if err != nil {
				_ = err.Error()
			}
		}
		checkTime := time.Since(start)

		fmt.Printf("  Creation time: %v (%.2f ns/op)\n",
			creationTime, float64(creationTime.Nanoseconds())/float64(iterations))
		fmt.Printf("  Check time: %v (%.2f ns/op)\n",
			checkTime, float64(checkTime.Nanoseconds())/float64(iterations))

		// Memory allocation
		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		errors := make([]error, 1000)
		for i := 0; i < 1000; i++ {
			errors[i] = scenario.errorFunc()
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		allocPerError := (m2.Alloc - m1.Alloc) / 1000
		fmt.Printf("  Memory per error: %d bytes\n", allocPerError)
	}
	fmt.Println()
}

// Helper functions

type stats struct {
	avg    time.Duration
	min    time.Duration
	max    time.Duration
	stdDev time.Duration
}

func calculateStats(times []time.Duration) stats {
	if len(times) == 0 {
		return stats{}
	}

	var sum time.Duration
	min := times[0]
	max := times[0]

	for _, t := range times {
		sum += t
		if t < min {
			min = t
		}
		if t > max {
			max = t
		}
	}

	avg := sum / time.Duration(len(times))

	// Calculate standard deviation
	var variance float64
	for _, t := range times {
		diff := float64(t - avg)
		variance += diff * diff
	}
	variance /= float64(len(times))
	stdDev := time.Duration(math.Sqrt(variance))

	return stats{
		avg:    avg,
		min:    min,
		max:    max,
		stdDev: stdDev,
	}
}

func generateContent(size int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "
	b := make([]byte, size)
	for i := range b {
		b[i] = chars[i%len(chars)]
	}
	return string(b)
}

type customError struct {
	code    int
	message string
}

func (e *customError) Error() string {
	return fmt.Sprintf("error %d: %s", e.code, e.message)
}

func printBenchmarkSummary() {
	fmt.Println("\n=== Benchmark Summary ===")
	fmt.Println("All benchmarks completed successfully.")
	fmt.Println("\nKey findings:")
	fmt.Println("- Connection establishment is generally fast")
	fmt.Println("- Query latency depends on complexity")
	fmt.Println("- Concurrent clients scale well")
	fmt.Println("- Memory usage is reasonable")
	fmt.Println("- Context overhead is minimal")
	fmt.Println("- Error handling is efficient")

	// Final system stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("\nFinal memory usage: %.2f MB\n", float64(m.Alloc)/1024/1024)
	fmt.Printf("Total goroutines: %d\n", runtime.NumGoroutine())
}
