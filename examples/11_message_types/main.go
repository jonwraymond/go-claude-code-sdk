package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/claudecode"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Message Types Examples ===\n")

	// Example 1: User messages
	example1UserMessages()

	// Example 2: Assistant messages
	example2AssistantMessages()

	// Example 3: System messages
	example3SystemMessages()

	// Example 4: Result messages
	example4ResultMessages()

	// Example 5: Content blocks
	example5ContentBlocks()

	// Example 6: Message flow analysis
	example6MessageFlowAnalysis()
}

func example1UserMessages() {
	fmt.Println("Example 1: User Messages")
	fmt.Println("------------------------")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Track user messages
	userMessages := []claudecode.UserMessage{}

	go func() {
		for msg := range client.ReceiveMessages() {
			switch m := msg.(type) {
			case *claudecode.UserMessage:
				userMessages = append(userMessages, *m)
				fmt.Printf("👤 User Message captured:\n")
				fmt.Printf("   Content: %v\n", m.Content)
				fmt.Printf("   Type: %T\n", m)
			}
		}
	}()

	// Send various user messages
	testQueries := []string{
		"Hello Claude!",
		"What's the weather like?",
		"Can you help me with coding?",
	}

	for i, query := range testQueries {
		fmt.Printf("\n📤 Sending query %d: %s\n", i+1, query)
		if err := client.Query(ctx, query, "user-msg-demo"); err != nil {
			log.Printf("Query failed: %v\n", err)
		}
		time.Sleep(2 * time.Second)
	}

	fmt.Printf("\n📊 User Message Summary:\n")
	fmt.Printf("   Total user messages: %d\n", len(userMessages))
	for i, msg := range userMessages {
		fmt.Printf("   Message %d: %v\n", i+1, msg.Content)
	}
	fmt.Println()
}

func example2AssistantMessages() {
	fmt.Println("Example 2: Assistant Messages")
	fmt.Println("-----------------------------")

	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Read", "Write", "Bash"}

	ctx := context.Background()
	
	// Different types of assistant responses
	queries := []struct {
		query       string
		expectTools bool
		desc        string
	}{
		{
			query:       "What is the capital of France?",
			expectTools: false,
			desc:        "Simple text response",
		},
		{
			query:       "Create a file called test.txt with 'Hello World' content",
			expectTools: true,
			desc:        "Response with tool use",
		},
		{
			query:       "Explain recursion and then create a recursive function example",
			expectTools: true,
			desc:        "Mixed content response",
		},
	}

	for _, test := range queries {
		fmt.Printf("\n🔹 Test: %s\n", test.desc)
		fmt.Printf("   Query: %s\n", test.query)
		
		msgChan := claudecode.Query(ctx, test.query, options)
		
		assistantStats := struct {
			textBlocks      int
			toolUseBlocks   int
			toolResultBlocks int
			totalContent    int
		}{}

		for msg := range msgChan {
			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				fmt.Printf("\n   🤖 Assistant Message:\n")
				fmt.Printf("      Content blocks: %d\n", len(m.Content))
				
				for i, block := range m.Content {
					assistantStats.totalContent++
					
					switch b := block.(type) {
					case claudecode.TextBlock:
						assistantStats.textBlocks++
						preview := b.Text
						if len(preview) > 100 {
							preview = preview[:100] + "..."
						}
						fmt.Printf("      [%d] Text: %s\n", i+1, preview)
						
					case claudecode.ToolUseBlock:
						assistantStats.toolUseBlocks++
						fmt.Printf("      [%d] Tool Use: %s\n", i+1, b.Name)
						fmt.Printf("          ID: %s\n", b.ID)
						if b.Input != nil {
							inputJSON, _ := json.MarshalIndent(b.Input, "          ", "  ")
							fmt.Printf("          Input: %s\n", string(inputJSON))
						}
						
					case claudecode.ToolResultBlock:
						assistantStats.toolResultBlocks++
						fmt.Printf("      [%d] Tool Result:\n", i+1)
						fmt.Printf("          Tool Use ID: %s\n", b.ToolUseID)
						fmt.Printf("          Content: %v\n", b.Content)
						if b.IsError != nil && *b.IsError {
							fmt.Printf("          ⚠️ Error: true\n")
						}
					}
				}
			}
		}
		
		fmt.Printf("\n   📊 Content Summary:\n")
		fmt.Printf("      Text blocks: %d\n", assistantStats.textBlocks)
		fmt.Printf("      Tool use blocks: %d\n", assistantStats.toolUseBlocks)
		fmt.Printf("      Tool result blocks: %d\n", assistantStats.toolResultBlocks)
		fmt.Printf("      Expected tools: %v\n", test.expectTools)
	}
	fmt.Println()
}

func example3SystemMessages() {
	fmt.Println("Example 3: System Messages")
	fmt.Println("--------------------------")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Track system messages by subtype
	systemMessages := make(map[string][]claudecode.SystemMessage)

	go func() {
		for msg := range client.ReceiveMessages() {
			switch m := msg.(type) {
			case *claudecode.SystemMessage:
				subtype := m.Subtype
				if subtype == "" {
					subtype = "unknown"
				}
				systemMessages[subtype] = append(systemMessages[subtype], *m)
				
				fmt.Printf("\n📋 System Message (subtype: %s)\n", subtype)
				
				// Pretty print data
				if m.Data != nil {
					for key, value := range m.Data {
						fmt.Printf("   %s: %v\n", key, value)
					}
				}
			}
		}
	}()

	// Trigger different system messages
	fmt.Println("Triggering various system messages...")

	// 1. Tool usage (generates tool_result system messages)
	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Bash"}
	
	client2 := claudecode.NewClaudeSDKClient(options)
	defer client2.Close()
	
	if err := client2.Connect(ctx); err == nil {
		client2.Query(ctx, "Run 'echo Hello System Messages'", "system-msg-1")
		time.Sleep(3 * time.Second)
	}

	// 2. Error scenarios
	client.Query(ctx, "Try to use a tool that doesn't exist: UseMagicTool", "system-msg-2")
	time.Sleep(2 * time.Second)

	// 3. Permission-related
	client.Query(ctx, "Try to access /etc/passwd", "system-msg-3")
	time.Sleep(2 * time.Second)

	// Summary
	fmt.Printf("\n📊 System Message Summary:\n")
	for subtype, messages := range systemMessages {
		fmt.Printf("   %s: %d messages\n", subtype, len(messages))
		
		// Show sample data keys
		if len(messages) > 0 && messages[0].Data != nil {
			keys := []string{}
			for k := range messages[0].Data {
				keys = append(keys, k)
			}
			fmt.Printf("      Data keys: %v\n", keys)
		}
	}
	fmt.Println()
}

func example4ResultMessages() {
	fmt.Println("Example 4: Result Messages")
	fmt.Println("--------------------------")

	// Test different scenarios that produce different result messages
	scenarios := []struct {
		name    string
		query   string
		options *claudecode.ClaudeCodeOptions
	}{
		{
			name:  "Simple query",
			query: "What is 2+2?",
			options: nil,
		},
		{
			name:  "With tools",
			query: "Create three files: a.txt, b.txt, c.txt",
			options: func() *claudecode.ClaudeCodeOptions {
				opts := claudecode.NewClaudeCodeOptions()
				opts.AllowedTools = []string{"Write"}
				return opts
			}(),
		},
		{
			name:  "Multi-turn limited",
			query: "Let's have a conversation",
			options: func() *claudecode.ClaudeCodeOptions {
				opts := claudecode.NewClaudeCodeOptions()
				opts.MaxTurns = claudecode.IntPtr(2)
				return opts
			}(),
		},
		{
			name:  "Error case",
			query: "",
			options: nil,
		},
	}

	for _, scenario := range scenarios {
		fmt.Printf("\n🔹 Scenario: %s\n", scenario.name)
		
		ctx := context.Background()
		msgChan := claudecode.Query(ctx, scenario.query, scenario.options)
		
		var resultMsg *claudecode.ResultMessage
		
		for msg := range msgChan {
			if result, ok := msg.(*claudecode.ResultMessage); ok {
				resultMsg = result
			}
		}
		
		if resultMsg != nil {
			fmt.Printf("   📊 Result Message:\n")
			fmt.Printf("      Subtype: %s\n", resultMsg.Subtype)
			fmt.Printf("      Duration: %dms (API: %dms)\n", 
				resultMsg.DurationMs, resultMsg.DurationAPIMs)
			fmt.Printf("      Is Error: %v\n", resultMsg.IsError)
			fmt.Printf("      Num Turns: %d\n", resultMsg.NumTurns)
			fmt.Printf("      Session ID: %s\n", resultMsg.SessionID)
			
			if resultMsg.TotalCostUSD != nil {
				fmt.Printf("      Cost: $%.6f\n", *resultMsg.TotalCostUSD)
			}
			
			if resultMsg.Usage != nil {
				fmt.Printf("      Usage:\n")
				usageJSON, _ := json.MarshalIndent(resultMsg.Usage, "        ", "  ")
				fmt.Printf("        %s\n", string(usageJSON))
			}
			
			if resultMsg.Result != nil {
				fmt.Printf("      Result: %s\n", *resultMsg.Result)
			}
		} else {
			fmt.Println("   ❌ No result message received")
		}
	}
	fmt.Println()
}

func example5ContentBlocks() {
	fmt.Println("Example 5: Content Blocks")
	fmt.Println("-------------------------")

	// Create different content block scenarios
	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Read", "Write", "Edit", "Bash"}

	ctx := context.Background()
	
	// Complex query that should generate various content blocks
	query := `Please do the following:
1. Explain what a linked list is
2. Create a file called linkedlist.go with a basic implementation
3. Show me the contents of the file
4. Run a simple test of the linked list`

	fmt.Printf("📝 Query: %s\n\n", query)

	msgChan := claudecode.Query(ctx, query, options)
	
	// Track all content blocks
	blockStats := map[string]int{
		"TextBlock":       0,
		"ToolUseBlock":    0,
		"ToolResultBlock": 0,
		"Unknown":         0,
	}
	
	blockDetails := []struct {
		Type    string
		Content interface{}
		Index   int
	}{}

	messageIndex := 0
	for msg := range msgChan {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			messageIndex++
			fmt.Printf("📨 Assistant Message #%d\n", messageIndex)
			
			for i, block := range m.Content {
				blockType := reflect.TypeOf(block).Name()
				blockStats[blockType]++
				
				blockDetail := struct {
					Type    string
					Content interface{}
					Index   int
				}{
					Type:    blockType,
					Content: block,
					Index:   i,
				}
				blockDetails = append(blockDetails, blockDetail)
				
				switch b := block.(type) {
				case claudecode.TextBlock:
					fmt.Printf("   [%d] 📝 Text Block:\n", i)
					lines := strings.Split(b.Text, "\n")
					for j, line := range lines {
						if j < 3 || j >= len(lines)-3 { // Show first and last 3 lines
							fmt.Printf("       %s\n", line)
						} else if j == 3 {
							fmt.Printf("       ... (%d lines omitted) ...\n", len(lines)-6)
						}
					}
					
				case claudecode.ToolUseBlock:
					fmt.Printf("   [%d] 🔧 Tool Use Block:\n", i)
					fmt.Printf("       Tool: %s\n", b.Name)
					fmt.Printf("       ID: %s\n", b.ID)
					
					// Show input parameters
					if inputMap, ok := b.Input.(map[string]interface{}); ok {
						fmt.Printf("       Parameters:\n")
						for key, value := range inputMap {
							valueStr := fmt.Sprintf("%v", value)
							if len(valueStr) > 50 {
								valueStr = valueStr[:50] + "..."
							}
							fmt.Printf("         %s: %s\n", key, valueStr)
						}
					}
					
				case claudecode.ToolResultBlock:
					fmt.Printf("   [%d] ✅ Tool Result Block:\n", i)
					fmt.Printf("       Tool Use ID: %s\n", b.ToolUseID)
					if b.IsError != nil && *b.IsError {
						fmt.Printf("       ❌ Error: true\n")
					}
					
					// Show content preview
					contentStr := fmt.Sprintf("%v", b.Content)
					if len(contentStr) > 100 {
						contentStr = contentStr[:100] + "..."
					}
					fmt.Printf("       Content: %s\n", contentStr)
					
				default:
					fmt.Printf("   [%d] ❓ Unknown Block Type: %T\n", i, block)
				}
			}
			fmt.Println()
		}
	}

	// Summary
	fmt.Printf("📊 Content Block Summary:\n")
	totalBlocks := 0
	for blockType, count := range blockStats {
		if count > 0 {
			fmt.Printf("   %s: %d\n", blockType, count)
			totalBlocks += count
		}
	}
	fmt.Printf("   Total blocks: %d\n", totalBlocks)
	
	// Analyze block patterns
	fmt.Printf("\n🔍 Block Pattern Analysis:\n")
	if len(blockDetails) > 0 {
		// Find tool use -> result patterns
		toolPatterns := 0
		for i := 0; i < len(blockDetails)-1; i++ {
			if blockDetails[i].Type == "ToolUseBlock" {
				// Check if followed by result
				for j := i + 1; j < len(blockDetails); j++ {
					if blockDetails[j].Type == "ToolResultBlock" {
						toolPatterns++
						break
					}
				}
			}
		}
		fmt.Printf("   Tool Use → Result patterns: %d\n", toolPatterns)
		
		// Text block positioning
		firstBlock := blockDetails[0].Type
		lastBlock := blockDetails[len(blockDetails)-1].Type
		fmt.Printf("   First block type: %s\n", firstBlock)
		fmt.Printf("   Last block type: %s\n", lastBlock)
	}
	fmt.Println()
}

func example6MessageFlowAnalysis() {
	fmt.Println("Example 6: Message Flow Analysis")
	fmt.Println("--------------------------------")

	client := claudecode.NewClaudeSDKClient(nil)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Message flow tracker
	messageFlow := []MessageFlowEntry{}
	startTime := time.Now()

	// Message processor with detailed tracking
	go func() {
		for msg := range client.ReceiveMessages() {
			entry := MessageFlowEntry{
				Timestamp: time.Now(),
				Elapsed:   time.Since(startTime),
				Type:      reflect.TypeOf(msg).String(),
				Message:   msg,
			}
			
			// Extract key information
			switch m := msg.(type) {
			case *claudecode.UserMessage:
				entry.Summary = fmt.Sprintf("User: %v", m.Content)
				
			case *claudecode.AssistantMessage:
				contentTypes := []string{}
				for _, block := range m.Content {
					contentTypes = append(contentTypes, reflect.TypeOf(block).Name())
				}
				entry.Summary = fmt.Sprintf("Assistant: %d blocks (%v)", 
					len(m.Content), strings.Join(contentTypes, ", "))
				
			case *claudecode.SystemMessage:
				entry.Summary = fmt.Sprintf("System (%s): %v", m.Subtype, m.Data)
				
			case *claudecode.ResultMessage:
				entry.Summary = fmt.Sprintf("Result: %dms, %d turns, error=%v", 
					m.DurationMs, m.NumTurns, m.IsError)
			}
			
			messageFlow = append(messageFlow, entry)
			
			// Real-time display
			fmt.Printf("[%s] %s\n", 
				entry.Elapsed.Round(time.Millisecond), 
				entry.Summary)
		}
	}()

	// Execute a multi-step conversation
	conversation := []struct {
		query  string
		delay  time.Duration
	}{
		{"Hello! I need help with a coding project.", 1 * time.Second},
		{"I want to build a REST API in Go.", 2 * time.Second},
		{"Can you show me a basic example?", 2 * time.Second},
		{"How do I add authentication?", 2 * time.Second},
		{"Thanks for your help!", 1 * time.Second},
	}

	fmt.Println("\n🗣️ Starting conversation...")
	for i, turn := range conversation {
		fmt.Printf("\n[Turn %d] %s\n", i+1, turn.query)
		if err := client.Query(ctx, turn.query, "flow-analysis"); err != nil {
			log.Printf("Query failed: %v\n", err)
		}
		time.Sleep(turn.delay)
	}

	// Wait for final messages
	time.Sleep(2 * time.Second)

	// Analyze message flow
	fmt.Println("\n📊 Message Flow Analysis:")
	fmt.Printf("   Total messages: %d\n", len(messageFlow))
	fmt.Printf("   Total duration: %s\n", messageFlow[len(messageFlow)-1].Elapsed.Round(time.Millisecond))
	
	// Message type distribution
	typeCount := make(map[string]int)
	for _, entry := range messageFlow {
		typeCount[entry.Type]++
	}
	
	fmt.Println("\n   Message Type Distribution:")
	for msgType, count := range typeCount {
		percentage := float64(count) * 100 / float64(len(messageFlow))
		fmt.Printf("     %s: %d (%.1f%%)\n", msgType, count, percentage)
	}
	
	// Timing analysis
	fmt.Println("\n   Timing Analysis:")
	var totalResponseTime time.Duration
	responseCount := 0
	
	for i := 0; i < len(messageFlow)-1; i++ {
		curr := messageFlow[i]
		next := messageFlow[i+1]
		
		// If user message followed by assistant message
		if strings.Contains(curr.Type, "UserMessage") && 
		   strings.Contains(next.Type, "AssistantMessage") {
			responseTime := next.Timestamp.Sub(curr.Timestamp)
			totalResponseTime += responseTime
			responseCount++
			fmt.Printf("     Response %d: %s\n", responseCount, responseTime.Round(time.Millisecond))
		}
	}
	
	if responseCount > 0 {
		avgResponseTime := totalResponseTime / time.Duration(responseCount)
		fmt.Printf("     Average response time: %s\n", avgResponseTime.Round(time.Millisecond))
	}
	
	// Message sequence pattern
	fmt.Println("\n   Message Sequence Pattern:")
	for i, entry := range messageFlow {
		indent := ""
		if strings.Contains(entry.Type, "Assistant") {
			indent = "  "
		} else if strings.Contains(entry.Type, "System") {
			indent = "    "
		} else if strings.Contains(entry.Type, "Result") {
			indent = "      "
		}
		
		fmt.Printf("   %s[%d] %s\n", indent, i+1, entry.Type)
	}
}

// Helper types

type MessageFlowEntry struct {
	Timestamp time.Time
	Elapsed   time.Duration
	Type      string
	Summary   string
	Message   types.Message
}