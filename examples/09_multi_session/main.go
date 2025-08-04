package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/claudecode"
)

func main() {
	fmt.Println("=== Multi-Session Examples ===")

	// Example 1: Basic multi-session
	example1BasicMultiSession()

	// Example 2: Parallel sessions
	example2ParallelSessions()

	// Example 3: Session isolation
	example3SessionIsolation()

	// Example 4: Session coordination
	example4SessionCoordination()

	// Example 5: Advanced session management
	example5AdvancedSessionManagement()
}

func example1BasicMultiSession() {
	fmt.Println("Example 1: Basic Multi-Session")
	fmt.Println("------------------------------")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Track messages by session
	sessionMessages := make(map[string][]string)
	var mu sync.Mutex

	// Message processor
	go func() {
		for msg := range client.ReceiveMessages() {
			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				// Extract session from context if available
				for _, block := range m.Content {
					if textBlock, ok := block.(claudecode.TextBlock); ok {
						mu.Lock()
						// We'll track by detecting session mentions
						sessionMessages["current"] = append(sessionMessages["current"], textBlock.Text)
						mu.Unlock()
					}
				}
			case *claudecode.ResultMessage:
				mu.Lock()
				sessionMessages[m.SessionID] = append(sessionMessages[m.SessionID],
					fmt.Sprintf("Session %s complete (%d turns)", m.SessionID, m.NumTurns))
				mu.Unlock()
			}
		}
	}()

	// Create multiple sessions
	sessions := []struct {
		id       string
		topic    string
		messages []string
	}{
		{
			id:    "math-session",
			topic: "Mathematics",
			messages: []string{
				"Let's work on some math problems",
				"What is 15% of 240?",
				"Now calculate the square root of 144",
			},
		},
		{
			id:    "coding-session",
			topic: "Programming",
			messages: []string{
				"Help me with Go programming",
				"How do I create a goroutine?",
				"What's the difference between a channel and a mutex?",
			},
		},
		{
			id:    "general-session",
			topic: "General Knowledge",
			messages: []string{
				"Tell me about the solar system",
				"Which planet is largest?",
				"How far is Mars from Earth?",
			},
		},
	}

	// Run sessions sequentially
	for _, session := range sessions {
		fmt.Printf("\n📚 Starting %s session (ID: %s)\n", session.topic, session.id)

		for i, msg := range session.messages {
			fmt.Printf("   [%d] You: %s\n", i+1, msg)
			if err := client.Query(ctx, msg, session.id); err != nil {
				log.Printf("Failed to send query: %v\n", err)
			}
			time.Sleep(2 * time.Second)
		}

		fmt.Printf("   ✅ %s session complete\n", session.topic)
	}

	// Summary
	time.Sleep(2 * time.Second)
	mu.Lock()
	fmt.Println("\n📊 Session Summary:")
	for sessionID, messages := range sessionMessages {
		fmt.Printf("   %s: %d messages\n", sessionID, len(messages))
	}
	mu.Unlock()
	fmt.Println()
}

func example2ParallelSessions() {
	fmt.Println("Example 2: Parallel Sessions")
	fmt.Println("----------------------------")

	// Create multiple clients for true parallel sessions
	numSessions := 3
	clients := make([]*claudecode.ClaudeSDKClient, numSessions)

	for i := range clients {
		clients[i] = claudecode.NewClaudeSDKClient(nil)
		defer clients[i].Close()
	}

	ctx := context.Background()
	var wg sync.WaitGroup

	// Session definitions
	sessions := []struct {
		name    string
		queries []string
		color   string
	}{
		{
			name: "Research",
			queries: []string{
				"Research the history of computing",
				"Who invented the first computer?",
				"What was ENIAC?",
			},
			color: "🔵",
		},
		{
			name: "Development",
			queries: []string{
				"Write a bubble sort algorithm in Go",
				"Now optimize it",
				"Add comments explaining the optimization",
			},
			color: "🟢",
		},
		{
			name: "Analysis",
			queries: []string{
				"Analyze the time complexity of quicksort",
				"Compare it with mergesort",
				"Which is better for large datasets?",
			},
			color: "🟡",
		},
	}

	// Start time for synchronization
	startTime := time.Now()

	// Run sessions in parallel
	for i, session := range sessions {
		wg.Add(1)
		go func(idx int, sess struct {
			name    string
			queries []string
			color   string
		}) {
			defer wg.Done()

			client := clients[idx]
			if err := client.Connect(ctx); err != nil {
				log.Printf("Session %s failed to connect: %v\n", sess.name, err)
				return
			}

			// Message processor for this session
			go func() {
				for msg := range client.ReceiveMessages() {
					elapsed := time.Since(startTime).Seconds()
					switch m := msg.(type) {
					case *claudecode.AssistantMessage:
						for _, block := range m.Content {
							if textBlock, ok := block.(claudecode.TextBlock); ok {
								// Show first 80 chars
								preview := textBlock.Text
								if len(preview) > 80 {
									preview = preview[:80] + "..."
								}
								fmt.Printf("[%.1fs] %s %s: %s\n",
									elapsed, sess.color, sess.name, preview)
							}
						}
					case *claudecode.ResultMessage:
						fmt.Printf("[%.1fs] %s %s: Turn %d complete\n",
							elapsed, sess.color, sess.name, m.NumTurns)
					}
				}
			}()

			// Send queries with slight delays
			for i, query := range sess.queries {
				time.Sleep(time.Duration(i*500) * time.Millisecond) // Stagger queries

				elapsed := time.Since(startTime).Seconds()
				fmt.Printf("[%.1fs] %s %s: Sending query %d\n",
					elapsed, sess.color, sess.name, i+1)

				if err := client.Query(ctx, query, fmt.Sprintf("%s-%d", sess.name, idx)); err != nil {
					log.Printf("Query failed: %v\n", err)
				}
			}

			// Keep session alive for responses
			time.Sleep(5 * time.Second)

		}(i, session)
	}

	wg.Wait()
	fmt.Println("\n✅ All parallel sessions complete")
	fmt.Println()
}

func example3SessionIsolation() {
	fmt.Println("Example 3: Session Isolation")
	fmt.Println("----------------------------")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	go func() {
		for msg := range client.ReceiveMessages() {
			switch m := msg.(type) {
			case *claudecode.ResultMessage:
				fmt.Printf("Session %s: %d turns\n", m.SessionID, m.NumTurns)
			}
		}
	}()

	// Test 1: Variable isolation
	fmt.Println("\n📝 Test 1: Variable Isolation")

	// Session A sets a variable
	client.Query(ctx, "Remember that x = 42", "session-A")
	time.Sleep(2 * time.Second)

	// Session B sets a different value
	client.Query(ctx, "Remember that x = 100", "session-B")
	time.Sleep(2 * time.Second)

	// Check if sessions maintain separate values
	client.Query(ctx, "What is the value of x?", "session-A")
	time.Sleep(2 * time.Second)

	client.Query(ctx, "What is the value of x?", "session-B")
	time.Sleep(2 * time.Second)

	// Test 2: Context isolation
	fmt.Println("\n📚 Test 2: Context Isolation")

	// Session C discusses Python
	client.Query(ctx, "I want to learn Python. What should I start with?", "session-C")
	time.Sleep(2 * time.Second)

	// Session D discusses JavaScript
	client.Query(ctx, "I want to learn JavaScript. What should I start with?", "session-D")
	time.Sleep(2 * time.Second)

	// Continue conversations
	client.Query(ctx, "What about advanced features?", "session-C") // Should be Python context
	time.Sleep(2 * time.Second)

	client.Query(ctx, "What about advanced features?", "session-D") // Should be JavaScript context
	time.Sleep(2 * time.Second)

	// Test 3: Tool usage isolation
	fmt.Println("\n🔧 Test 3: Tool Usage Isolation")

	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Write"}

	client2 := claudecode.NewClaudeSDKClient(options)
	defer client2.Close()

	if err := client2.Connect(ctx); err != nil {
		log.Printf("Failed to connect client2: %v\n", err)
		return
	}

	// Different sessions create different files
	sessionFiles := map[string]string{
		"project-A": "project_a_config.json",
		"project-B": "project_b_config.json",
		"project-C": "project_c_config.json",
	}

	for session, filename := range sessionFiles {
		query := fmt.Sprintf("Create a configuration file named %s with project settings", filename)
		fmt.Printf("\n🗂️ %s: Creating %s\n", session, filename)

		client2.Query(ctx, query, session)
		time.Sleep(3 * time.Second)
	}

	fmt.Println("\n✅ Session isolation tests complete")
	fmt.Println()
}

func example4SessionCoordination() {
	fmt.Println("Example 4: Session Coordination")
	fmt.Println("-------------------------------")

	// Shared data structure for coordination
	sharedData := &SharedSessionData{
		results: make(map[string]interface{}),
	}

	// Create coordinator client
	coordinator := claudecode.NewClaudeSDKClient(nil)
	defer coordinator.Close()

	// Create worker clients
	numWorkers := 3
	workers := make([]*claudecode.ClaudeSDKClient, numWorkers)
	for i := range workers {
		workers[i] = claudecode.NewClaudeSDKClient(nil)
		defer workers[i].Close()
	}

	ctx := context.Background()

	// Connect all clients
	if err := coordinator.Connect(ctx); err != nil {
		log.Fatal("Coordinator failed to connect:", err)
	}

	for i, worker := range workers {
		if err := worker.Connect(ctx); err != nil {
			log.Printf("Worker %d failed to connect: %v\n", i, err)
		}
	}

	// Coordinator message processor
	go processCoordinatorMessages(coordinator, sharedData)

	// Worker message processors
	for i, worker := range workers {
		go processWorkerMessages(worker, i, sharedData)
	}

	// Coordination example: Distributed calculation
	fmt.Println("\n🧮 Distributed Calculation Example")

	// Coordinator assigns tasks
	coordinator.Query(ctx, "We need to calculate statistics for these number sets. Coordinate the work.", "coordinator")
	time.Sleep(1 * time.Second)

	// Workers process different parts
	tasks := []string{
		"Calculate mean and median of: 12, 45, 67, 23, 89, 34, 56",
		"Calculate mean and median of: 98, 76, 54, 32, 21, 43, 65",
		"Calculate mean and median of: 11, 22, 33, 44, 55, 66, 77",
	}

	for i, task := range tasks {
		workers[i].Query(ctx, task, fmt.Sprintf("worker-%d", i))
		time.Sleep(2 * time.Second)
	}

	// Coordinator aggregates results
	time.Sleep(2 * time.Second)
	coordinator.Query(ctx, "Summarize all the statistics calculated by the workers", "coordinator")
	time.Sleep(3 * time.Second)

	// Show shared results
	sharedData.mu.Lock()
	fmt.Println("\n📊 Shared Results:")
	for key, value := range sharedData.results {
		fmt.Printf("   %s: %v\n", key, value)
	}
	sharedData.mu.Unlock()

	fmt.Println("\n✅ Coordinated session complete")
	fmt.Println()
}

func example5AdvancedSessionManagement() {
	fmt.Println("Example 5: Advanced Session Management")
	fmt.Println("-------------------------------------")

	// Create session manager
	sessionMgr := NewSessionManager()

	ctx := context.Background()

	// Define session configurations
	configs := []SessionConfig{
		{
			ID:          "priority-high",
			Priority:    3,
			MaxTokens:   4000,
			Timeout:     30 * time.Second,
			RetryCount:  3,
			Description: "High priority financial analysis",
		},
		{
			ID:          "priority-medium",
			Priority:    2,
			MaxTokens:   2000,
			Timeout:     20 * time.Second,
			RetryCount:  2,
			Description: "Medium priority research task",
		},
		{
			ID:          "priority-low",
			Priority:    1,
			MaxTokens:   1000,
			Timeout:     10 * time.Second,
			RetryCount:  1,
			Description: "Low priority general query",
		},
	}

	// Create sessions with configurations
	for _, config := range configs {
		session, err := sessionMgr.CreateSession(config)
		if err != nil {
			log.Printf("Failed to create session %s: %v\n", config.ID, err)
			continue
		}
		fmt.Printf("✅ Created session: %s (priority: %d)\n", session.ID, session.Priority)
	}

	// Test 1: Priority-based execution
	fmt.Println("\n🎯 Test 1: Priority-based Execution")

	// Queue tasks for different priority sessions
	tasks := []struct {
		sessionID string
		query     string
	}{
		{"priority-low", "What is 2+2?"},
		{"priority-high", "Analyze market trends for Q4 2024"},
		{"priority-medium", "Explain quantum computing basics"},
		{"priority-high", "Calculate ROI for investment portfolio"},
		{"priority-low", "Tell me a fun fact"},
	}

	// Execute tasks based on priority
	sessionMgr.ExecuteByPriority(ctx, tasks)

	// Test 2: Session lifecycle management
	fmt.Println("\n🔄 Test 2: Session Lifecycle")

	// Create ephemeral session
	ephemeral := SessionConfig{
		ID:          "ephemeral-task",
		Priority:    2,
		MaxTokens:   1000,
		Timeout:     5 * time.Second,
		AutoClose:   true,
		Description: "Auto-closing session",
	}

	session, _ := sessionMgr.CreateSession(ephemeral)
	fmt.Printf("📝 Created ephemeral session: %s\n", session.ID)

	// Use ephemeral session
	sessionMgr.Query(ctx, session.ID, "Quick calculation: 123 * 456")
	time.Sleep(6 * time.Second)

	// Check if auto-closed
	if sessionMgr.IsSessionActive(session.ID) {
		fmt.Println("❌ Ephemeral session still active (unexpected)")
	} else {
		fmt.Println("✅ Ephemeral session auto-closed")
	}

	// Test 3: Session metrics
	fmt.Println("\n📊 Test 3: Session Metrics")

	metrics := sessionMgr.GetMetrics()
	fmt.Printf("   Total sessions: %d\n", metrics.TotalSessions)
	fmt.Printf("   Active sessions: %d\n", metrics.ActiveSessions)
	fmt.Printf("   Total queries: %d\n", metrics.TotalQueries)
	fmt.Printf("   Average response time: %.2fs\n", metrics.AvgResponseTime)

	// Test 4: Session persistence
	fmt.Println("\n💾 Test 4: Session Persistence")

	// Save session state
	state := sessionMgr.SaveState()
	fmt.Printf("   Saved state for %d sessions\n", len(state.Sessions))

	// Simulate restart by creating new manager
	newSessionMgr := NewSessionManager()

	// Restore state
	err := newSessionMgr.RestoreState(state)
	if err != nil {
		fmt.Printf("❌ Failed to restore state: %v\n", err)
	} else {
		fmt.Printf("✅ Restored %d sessions\n", len(state.Sessions))
	}

	// Clean up
	sessionMgr.CloseAll()
	fmt.Println("\n✅ Advanced session management complete")
}

// Helper types and functions

type SharedSessionData struct {
	mu      sync.Mutex
	results map[string]interface{}
}

func processCoordinatorMessages(client *claudecode.ClaudeSDKClient, shared *SharedSessionData) {
	for msg := range client.ReceiveMessages() {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range m.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					fmt.Printf("🎯 Coordinator: %s\n", textBlock.Text)
				}
			}
		}
	}
}

func processWorkerMessages(client *claudecode.ClaudeSDKClient, workerID int, shared *SharedSessionData) {
	for msg := range client.ReceiveMessages() {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range m.Content {
				if textBlock, ok := block.(claudecode.TextBlock); ok {
					// Extract results (simplified)
					shared.mu.Lock()
					shared.results[fmt.Sprintf("worker-%d", workerID)] = textBlock.Text
					shared.mu.Unlock()

					// Show truncated output
					preview := textBlock.Text
					if len(preview) > 60 {
						preview = preview[:60] + "..."
					}
					fmt.Printf("👷 Worker %d: %s\n", workerID, preview)
				}
			}
		}
	}
}

type SessionConfig struct {
	ID          string
	Priority    int
	MaxTokens   int
	Timeout     time.Duration
	RetryCount  int
	AutoClose   bool
	Description string
}

type Session struct {
	SessionConfig
	Client      *claudecode.ClaudeSDKClient
	CreatedAt   time.Time
	LastUsed    time.Time
	QueryCount  int
	TotalTokens int
	Active      bool
}

type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (sm *SessionManager) CreateSession(config SessionConfig) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	client := claudecode.NewClaudeSDKClient(nil)

	session := &Session{
		SessionConfig: config,
		Client:        client,
		CreatedAt:     time.Now(),
		LastUsed:      time.Now(),
		Active:        true,
	}

	sm.sessions[config.ID] = session

	// Connect in background
	go func() {
		ctx := context.Background()
		if err := client.Connect(ctx); err != nil {
			log.Printf("Session %s failed to connect: %v\n", config.ID, err)
			session.Active = false
		}
	}()

	return session, nil
}

func (sm *SessionManager) ExecuteByPriority(ctx context.Context, tasks []struct {
	sessionID string
	query     string
}) {
	// Sort by priority
	// In real implementation, would use a priority queue
	for i := 3; i >= 1; i-- { // High to low priority
		for _, task := range tasks {
			sm.mu.RLock()
			session, exists := sm.sessions[task.sessionID]
			sm.mu.RUnlock()

			if exists && session.Priority == i {
				fmt.Printf("⚡ Executing priority %d: %s\n", i, task.sessionID)
				sm.Query(ctx, task.sessionID, task.query)
				time.Sleep(2 * time.Second)
			}
		}
	}
}

func (sm *SessionManager) Query(ctx context.Context, sessionID, query string) error {
	sm.mu.Lock()
	session, exists := sm.sessions[sessionID]
	if exists {
		session.LastUsed = time.Now()
		session.QueryCount++
	}
	sm.mu.Unlock()

	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	return session.Client.Query(ctx, query, sessionID)
}

func (sm *SessionManager) IsSessionActive(sessionID string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	return exists && session.Active
}

type SessionMetrics struct {
	TotalSessions   int
	ActiveSessions  int
	TotalQueries    int
	AvgResponseTime float64
}

func (sm *SessionManager) GetMetrics() SessionMetrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	metrics := SessionMetrics{
		TotalSessions: len(sm.sessions),
	}

	for _, session := range sm.sessions {
		if session.Active {
			metrics.ActiveSessions++
		}
		metrics.TotalQueries += session.QueryCount
	}

	// Simulated average response time
	if metrics.TotalQueries > 0 {
		metrics.AvgResponseTime = 2.5
	}

	return metrics
}

type SessionState struct {
	Sessions map[string]SessionConfig
	SavedAt  time.Time
}

func (sm *SessionManager) SaveState() SessionState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	state := SessionState{
		Sessions: make(map[string]SessionConfig),
		SavedAt:  time.Now(),
	}

	for id, session := range sm.sessions {
		state.Sessions[id] = session.SessionConfig
	}

	return state
}

func (sm *SessionManager) RestoreState(state SessionState) error {
	for id, config := range state.Sessions {
		_, err := sm.CreateSession(config)
		if err != nil {
			return fmt.Errorf("failed to restore session %s: %w", id, err)
		}
	}
	return nil
}

func (sm *SessionManager) CloseAll() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for id, session := range sm.sessions {
		session.Client.Close()
		session.Active = false
		fmt.Printf("🔒 Closed session: %s\n", id)
	}
}
