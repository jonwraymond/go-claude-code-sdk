package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/claudecode"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Custom Options Examples ===\n")

	// Example 1: Tool restrictions
	example1ToolRestrictions()

	// Example 2: Permission modes
	example2PermissionModes()

	// Example 3: Working directory
	example3WorkingDirectory()

	// Example 4: System prompts
	example4SystemPrompts()

	// Example 5: MCP server configuration
	example5MCPServers()

	// Example 6: Advanced options
	example6AdvancedOptions()
}

func example1ToolRestrictions() {
	fmt.Println("Example 1: Tool Restrictions")
	fmt.Println("----------------------------")

	// Test with different tool sets
	toolSets := []struct {
		name  string
		tools []string
		query string
	}{
		{
			name:  "Read-only",
			tools: []string{"Read", "Glob", "Grep"},
			query: "What files are in the current directory?",
		},
		{
			name:  "Write-only",
			tools: []string{"Write"},
			query: "Create a hello.txt file",
		},
		{
			name:  "No tools",
			tools: []string{},
			query: "Explain what goroutines are",
		},
	}

	for _, ts := range toolSets {
		fmt.Printf("\n🔧 Testing with %s tools: %v\n", ts.name, ts.tools)

		options := claudecode.NewClaudeCodeOptions()
		options.AllowedTools = ts.tools

		ctx := context.Background()
		msgChan := claudecode.Query(ctx, ts.query, options)

		toolsUsed := make(map[string]int)

		for msg := range msgChan {
			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range m.Content {
					if toolUse, ok := block.(claudecode.ToolUseBlock); ok {
						toolsUsed[toolUse.Name]++
					}
				}
			case *claudecode.ResultMessage:
				fmt.Printf("✅ Completed. Tools used: %v\n", toolsUsed)
			}
		}
	}
	fmt.Println()
}

func example2PermissionModes() {
	fmt.Println("Example 2: Permission Modes")
	fmt.Println("---------------------------")

	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "claude-sdk-test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test different permission modes
	modes := []struct {
		name string
		mode claudecode.PermissionMode
		desc string
	}{
		{
			name: "Default",
			mode: claudecode.PermissionModeDefault,
			desc: "Asks for permission for each edit",
		},
		{
			name: "Accept Edits",
			mode: claudecode.PermissionModeAcceptEdits,
			desc: "Automatically accepts all edits",
		},
		{
			name: "Bypass Permissions",
			mode: claudecode.PermissionModeBypassPermission,
			desc: "Bypasses all permission checks",
		},
	}

	for _, pm := range modes {
		fmt.Printf("\n🔐 Testing %s mode: %s\n", pm.name, pm.desc)

		options := claudecode.NewClaudeCodeOptions()
		options.PermissionMode = &pm.mode
		options.CWD = &tempDir
		options.AllowedTools = []string{"Write", "Edit"}

		// Use WithPermissionMode helper
		alternativeOptions := claudecode.NewClaudeCodeOptions()
		claudecode.WithPermissionMode(pm.mode)(alternativeOptions)
		claudecode.WithCWD(tempDir)(alternativeOptions)
		claudecode.WithAllowedTools([]string{"Write", "Edit"})(alternativeOptions)

		ctx := context.Background()
		msgChan := claudecode.Query(ctx, fmt.Sprintf("Create a file called %s.txt", pm.name), options)

		for msg := range msgChan {
			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				// Check for permission-related messages
				for _, block := range m.Content {
					if textBlock, ok := block.(claudecode.TextBlock); ok {
						if len(textBlock.Text) > 0 && len(textBlock.Text) < 100 {
							fmt.Printf("   %s\n", textBlock.Text)
						}
					}
				}
			case *claudecode.ResultMessage:
				// Check if file was created
				filePath := filepath.Join(tempDir, fmt.Sprintf("%s.txt", pm.name))
				if _, err := os.Stat(filePath); err == nil {
					fmt.Printf("✅ File created successfully\n")
				} else {
					fmt.Printf("❌ File not created\n")
				}
			}
		}
	}
	fmt.Println()
}

func example3WorkingDirectory() {
	fmt.Println("Example 3: Working Directory")
	fmt.Println("----------------------------")

	// Create test directories
	testDir, err := os.MkdirTemp("", "claude-wd-test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	// Create subdirectories
	subDirs := []string{"src", "tests", "docs"}
	for _, dir := range subDirs {
		os.MkdirAll(filepath.Join(testDir, dir), 0755)
	}

	// Test CWD setting
	options := claudecode.NewClaudeCodeOptions()
	options.SetCWD(testDir)

	// Alternative using helper function
	options2 := claudecode.NewClaudeCodeOptions()
	claudecode.WithCWD(testDir)(options2)

	ctx := context.Background()
	fmt.Printf("Working directory set to: %s\n", testDir)

	msgChan := claudecode.Query(ctx, "List all directories in the current working directory", options)

	for msg := range msgChan {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range m.Content {
				switch b := block.(type) {
				case claudecode.TextBlock:
					// Look for directory listings
					if contains(b.Text, subDirs...) {
						fmt.Printf("✅ Found expected directories\n")
					}
				case claudecode.ToolUseBlock:
					if b.Name == "Bash" || b.Name == "LS" {
						fmt.Printf("🔧 Using %s tool in working directory\n", b.Name)
					}
				}
			}
		}
	}
	fmt.Println()
}

func example4SystemPrompts() {
	fmt.Println("Example 4: System Prompts")
	fmt.Println("-------------------------")

	// Different system prompts for different behaviors
	prompts := []struct {
		name   string
		prompt string
		query  string
	}{
		{
			name:   "Pirate Mode",
			prompt: "You are a pirate. Speak like a pirate in all your responses, matey!",
			query:  "How do I write a hello world program?",
		},
		{
			name:   "Concise Mode",
			prompt: "Be extremely concise. Use bullet points. No explanations unless asked.",
			query:  "What are the main features of Go?",
		},
		{
			name:   "Teacher Mode",
			prompt: "You are a patient teacher. Explain concepts step-by-step with examples.",
			query:  "What is a pointer?",
		},
		{
			name:   "Code Review Mode",
			prompt: "You are a strict code reviewer. Point out any issues and suggest improvements.",
			query:  "Review this: func add(a, b int) { return a + b }",
		},
	}

	for _, p := range prompts {
		fmt.Printf("\n🎭 %s:\n", p.name)

		options := claudecode.NewClaudeCodeOptions()
		options.SystemPrompt = claudecode.StringPtr(p.prompt)

		// Also test WithSystemPrompt helper
		options2 := claudecode.NewClaudeCodeOptions()
		claudecode.WithSystemPrompt(p.prompt)(options2)

		ctx := context.Background()
		msgChan := claudecode.Query(ctx, p.query, options)

		for msg := range msgChan {
			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range m.Content {
					if textBlock, ok := block.(claudecode.TextBlock); ok {
						// Show first 200 chars to see the style
						preview := textBlock.Text
						if len(preview) > 200 {
							preview = preview[:200] + "..."
						}
						fmt.Printf("Response: %s\n", preview)
					}
				}
			}
		}
	}
	fmt.Println()
}

func example5MCPServers() {
	fmt.Println("Example 5: MCP Server Configuration")
	fmt.Println("-----------------------------------")

	options := claudecode.NewClaudeCodeOptions()

	// Configure MCP servers
	options.MCPServers = map[string]types.McpServerConfig{
		"filesystem": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
			Env: map[string]string{
				"MCP_READ_ONLY": "true",
			},
		},
		"github": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-github"},
			Env: map[string]string{
				"GITHUB_TOKEN": os.Getenv("GITHUB_TOKEN"),
			},
		},
	}

	// Specify which MCP tools to use
	options.MCPTools = []string{"mcp_filesystem_read", "mcp_github_search"}

	fmt.Println("Configured MCP servers:")
	for name, config := range options.MCPServers {
		fmt.Printf("  - %s: %s %v\n", name, config.Command, config.Args)
	}
	fmt.Printf("Enabled MCP tools: %v\n\n", options.MCPTools)

	ctx := context.Background()
	msgChan := claudecode.Query(ctx, "Use MCP tools to explore the filesystem", options)

	mcpToolsUsed := 0
	for msg := range msgChan {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range m.Content {
				if toolUse, ok := block.(claudecode.ToolUseBlock); ok {
					if len(toolUse.Name) > 4 && toolUse.Name[:4] == "mcp_" {
						mcpToolsUsed++
						fmt.Printf("🔧 MCP tool used: %s\n", toolUse.Name)
					}
				}
			}
		case *claudecode.ResultMessage:
			fmt.Printf("✅ Completed. MCP tools used: %d\n", mcpToolsUsed)
		}
	}
	fmt.Println()
}

func example6AdvancedOptions() {
	fmt.Println("Example 6: Advanced Options")
	fmt.Println("---------------------------")

	options := claudecode.NewClaudeCodeOptions()

	// Set multiple advanced options
	options.MaxTurns = claudecode.IntPtr(5)
	options.MaxThinkingTokens = 16000
	options.DisallowedTools = []string{"Bash", "Edit"}
	options.Model = claudecode.StringPtr("claude-3-opus-20240229")
	options.AddDirs = []string{"/tmp/test1", "/tmp/test2"}
	options.ContinueConversation = false

	// Using helper functions
	options2 := claudecode.NewClaudeCodeOptions()
	claudecode.WithMaxTurns(5)(options2)

	fmt.Println("Advanced configuration:")
	fmt.Printf("  Max turns: %d\n", *options.MaxTurns)
	fmt.Printf("  Max thinking tokens: %d\n", options.MaxThinkingTokens)
	fmt.Printf("  Disallowed tools: %v\n", options.DisallowedTools)
	fmt.Printf("  Model: %s\n", *options.Model)
	fmt.Printf("  Additional directories: %v\n", options.AddDirs)
	fmt.Printf("  Continue conversation: %v\n\n", options.ContinueConversation)

	// Create a client to test turn limits
	client := claudecode.NewClaudeSDKClient(options)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v\n", err)
		return
	}

	turnCount := 0
	go func() {
		for msg := range client.ReceiveMessages() {
			switch m := msg.(type) {
			case *claudecode.UserMessage:
				turnCount++
				fmt.Printf("📤 Turn %d (User)\n", turnCount)
			case *claudecode.AssistantMessage:
				fmt.Printf("📥 Turn %d (Assistant)\n", turnCount)
			case *claudecode.ResultMessage:
				fmt.Printf("\n✅ Conversation ended after %d turns\n", m.NumTurns)
				fmt.Printf("   Total tokens used: %v\n", m.Usage)
			}
		}
	}()

	// Send multiple queries to test turn limit
	queries := []string{
		"What is 1+1?",
		"What is 2+2?",
		"What is 3+3?",
		"What is 4+4?",
		"What is 5+5?",
		"What is 6+6?", // This should exceed turn limit
	}

	for i, query := range queries {
		if err := client.Query(ctx, query, "turn-test"); err != nil {
			fmt.Printf("❌ Query %d failed: %v\n", i+1, err)
			break
		}
		time.Sleep(2 * time.Second)
	}

	time.Sleep(3 * time.Second)
}

// Helper function
func contains(text string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(substr) > 0 && len(text) > 0 {
			return true // Simplified for example
		}
	}
	return false
}
