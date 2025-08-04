package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/claudecode"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Permission Mode Examples ===\n")

	// Example 1: Default permission mode
	example1DefaultMode()

	// Example 2: Accept edits mode
	example2AcceptEditsMode()

	// Example 3: Bypass permissions mode
	example3BypassMode()

	// Example 4: Interactive permission handling
	example4InteractivePermissions()

	// Example 5: Custom permission workflow
	example5CustomPermissionWorkflow()
}

func example1DefaultMode() {
	fmt.Println("Example 1: Default Permission Mode")
	fmt.Println("----------------------------------")
	fmt.Println("In default mode, Claude will ask for permission before making changes.\n")

	// Create test directory
	testDir, err := os.MkdirTemp("", "claude-perm-default")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	options := claudecode.NewClaudeCodeOptions()
	options.CWD = &testDir
	options.AllowedTools = []string{"Write", "Edit", "Read"}
	// Default mode - no need to set explicitly

	ctx := context.Background()
	
	// Create initial file
	testFile := filepath.Join(testDir, "config.json")
	os.WriteFile(testFile, []byte(`{"version": "1.0", "debug": false}`), 0644)
	
	fmt.Printf("📁 Working directory: %s\n", testDir)
	fmt.Println("📝 Task: Update config.json to enable debug mode")
	fmt.Println("\nIn default mode, you would normally see permission prompts.")
	fmt.Println("For this example, we'll simulate the behavior.\n")

	msgChan := claudecode.Query(ctx, 
		fmt.Sprintf("Update the config.json file to set debug to true. The file is at %s", testFile), 
		options)

	for msg := range msgChan {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range m.Content {
				switch b := block.(type) {
				case claudecode.TextBlock:
					if strings.Contains(b.Text, "permission") || 
					   strings.Contains(b.Text, "allow") ||
					   strings.Contains(b.Text, "edit") {
						fmt.Printf("🔐 Permission-related message: %s\n", b.Text)
					}
				case claudecode.ToolUseBlock:
					fmt.Printf("🔧 Tool request: %s\n", b.Name)
					if b.Name == "Edit" || b.Name == "Write" {
						fmt.Println("   [In default mode, this would prompt for permission]")
					}
				}
			}
		case *claudecode.SystemMessage:
			if m.Subtype == "permission_request" {
				fmt.Printf("❓ Permission request: %v\n", m.Data)
			}
		}
	}
	
	// Check if file was modified
	content, _ := os.ReadFile(testFile)
	fmt.Printf("\n📄 Final file content: %s\n", string(content))
	fmt.Println()
}

func example2AcceptEditsMode() {
	fmt.Println("Example 2: Accept Edits Mode")
	fmt.Println("----------------------------")
	fmt.Println("In accept-edits mode, file edits are automatically approved.\n")

	testDir, err := os.MkdirTemp("", "claude-perm-accept")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	options := claudecode.NewClaudeCodeOptions()
	options.CWD = &testDir
	options.AllowedTools = []string{"Write", "Edit", "Read", "Bash"}
	
	// Set to accept edits mode
	acceptMode := claudecode.PermissionModeAcceptEdits
	options.PermissionMode = &acceptMode

	// Create test files
	files := []string{
		"main.go",
		"utils.go", 
		"config.yaml",
		"README.md",
	}
	
	for _, file := range files {
		path := filepath.Join(testDir, file)
		content := fmt.Sprintf("// Original content of %s\n", file)
		os.WriteFile(path, []byte(content), 0644)
	}

	fmt.Printf("📁 Working directory: %s\n", testDir)
	fmt.Printf("📝 Created files: %v\n", files)
	fmt.Println("\n🔓 Permission mode: ACCEPT EDITS")
	fmt.Println("   All file modifications will be automatically approved.\n")

	ctx := context.Background()
	msgChan := claudecode.Query(ctx,
		"Add a comment header to all .go files with copyright notice and current year",
		options)

	editsPerformed := 0
	filesModified := []string{}

	for msg := range msgChan {
		switch m := msg.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range m.Content {
				switch b := block.(type) {
				case claudecode.ToolUseBlock:
					if b.Name == "Edit" || b.Name == "Write" {
						editsPerformed++
						if input, ok := b.Input.(map[string]interface{}); ok {
							if filePath, ok := input["file_path"].(string); ok {
								filesModified = append(filesModified, filepath.Base(filePath))
							}
						}
						fmt.Printf("✅ Auto-approved edit #%d\n", editsPerformed)
					}
				}
			}
		}
	}

	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("   Edits performed: %d\n", editsPerformed)
	fmt.Printf("   Files modified: %v\n", filesModified)
	
	// Show modified content
	fmt.Println("\n📄 Modified files:")
	for _, file := range files {
		if strings.HasSuffix(file, ".go") {
			path := filepath.Join(testDir, file)
			content, _ := os.ReadFile(path)
			fmt.Printf("\n--- %s ---\n%s\n", file, string(content))
		}
	}
	fmt.Println()
}

func example3BypassMode() {
	fmt.Println("Example 3: Bypass Permissions Mode")
	fmt.Println("----------------------------------")
	fmt.Println("In bypass mode, all operations proceed without any permission checks.\n")

	testDir, err := os.MkdirTemp("", "claude-perm-bypass")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	options := claudecode.NewClaudeCodeOptions()
	options.CWD = &testDir
	options.AllowedTools = []string{"Write", "Edit", "Read", "Bash", "LS"}
	
	// Set to bypass permissions mode
	bypassMode := claudecode.PermissionModeBypassPermission
	options.PermissionMode = &bypassMode

	fmt.Printf("📁 Working directory: %s\n", testDir)
	fmt.Println("\n⚡ Permission mode: BYPASS PERMISSIONS")
	fmt.Println("   All operations will proceed without permission checks.")
	fmt.Println("   Use with caution!\n")

	// Track operations
	operations := make(map[string]int)
	
	client := claudecode.NewClaudeSDKClient(options)
	defer client.Close()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Process messages
	go func() {
		for msg := range client.ReceiveMessages() {
			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range m.Content {
					if toolUse, ok := block.(claudecode.ToolUseBlock); ok {
						operations[toolUse.Name]++
						fmt.Printf("⚡ Bypassed permission for: %s\n", toolUse.Name)
					}
				}
			}
		}
	}()

	// Complex task that would normally require many permissions
	tasks := []string{
		"Create a project structure with src/, tests/, and docs/ directories",
		"Create a main.go file in src/ with a hello world program",
		"Create a README.md with project documentation",
		"Run 'ls -la' to show all files",
		"Create a Makefile with build instructions",
	}

	for i, task := range tasks {
		fmt.Printf("\n📌 Task %d: %s\n", i+1, task)
		if err := client.Query(ctx, task, "bypass-demo"); err != nil {
			log.Printf("Query failed: %v\n", err)
		}
		time.Sleep(2 * time.Second)
	}

	// Summary
	fmt.Printf("\n📊 Operations Summary (all bypassed):\n")
	total := 0
	for op, count := range operations {
		fmt.Printf("   %s: %d times\n", op, count)
		total += count
	}
	fmt.Printf("   Total: %d operations\n", total)

	// Show created structure
	fmt.Println("\n📁 Created structure:")
	filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(testDir, path)
		if relPath != "." {
			indent := strings.Count(relPath, string(os.PathSeparator))
			fmt.Printf("%s%s\n", strings.Repeat("  ", indent), filepath.Base(path))
		}
		return nil
	})
	fmt.Println()
}

func example4InteractivePermissions() {
	fmt.Println("Example 4: Interactive Permission Handling")
	fmt.Println("------------------------------------------")
	fmt.Println("Simulating interactive permission decisions.\n")

	testDir, err := os.MkdirTemp("", "claude-perm-interactive")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	// Simulate different permission decisions
	permissionDecisions := map[string]bool{
		"create_sensitive.txt": false,  // Deny
		"create_safe.txt":      true,   // Allow
		"edit_config.json":     true,   // Allow
		"delete_important.log": false,  // Deny
		"run_safe_command":     true,   // Allow
		"run_risky_command":    false,  // Deny
	}

	fmt.Println("📋 Simulated permission decisions:")
	for action, allowed := range permissionDecisions {
		emoji := "❌"
		if allowed {
			emoji = "✅"
		}
		fmt.Printf("   %s %s: %v\n", emoji, action, allowed)
	}
	fmt.Println()

	// Create a custom permission handler simulation
	options := claudecode.NewClaudeCodeOptions()
	options.CWD = &testDir
	options.AllowedTools = []string{"Write", "Edit", "Bash", "Read"}

	ctx := context.Background()
	
	// Test various operations
	testOperations := []struct {
		query      string
		shouldWork bool
		reason     string
	}{
		{
			query:      "Create a file called safe.txt with 'This is safe content'",
			shouldWork: true,
			reason:     "Non-sensitive file creation",
		},
		{
			query:      "Create a file called sensitive.txt with password information",
			shouldWork: false,
			reason:     "Contains sensitive data",
		},
		{
			query:      "Edit config.json to update version number",
			shouldWork: true,
			reason:     "Safe configuration update",
		},
		{
			query:      "Delete important.log file",
			shouldWork: false,
			reason:     "Deleting important files",
		},
		{
			query:      "Run 'echo Hello World'",
			shouldWork: true,
			reason:     "Safe command execution",
		},
		{
			query:      "Run 'rm -rf /'",
			shouldWork: false,
			reason:     "Dangerous command",
		},
	}

	for i, test := range testOperations {
		fmt.Printf("\n🧪 Test %d: %s\n", i+1, test.reason)
		fmt.Printf("   Query: %s\n", test.query)
		fmt.Printf("   Expected: %v\n", test.shouldWork)
		
		// In real usage, this would involve actual permission prompts
		// Here we simulate the outcome
		if test.shouldWork {
			fmt.Println("   Result: ✅ Operation allowed")
		} else {
			fmt.Println("   Result: ❌ Operation denied")
		}
	}
	fmt.Println()
}

func example5CustomPermissionWorkflow() {
	fmt.Println("Example 5: Custom Permission Workflow")
	fmt.Println("-------------------------------------")

	// Create a permission policy simulator
	policy := &PermissionPolicy{
		AllowedPaths: []string{"/tmp", "/home/user/safe"},
		DeniedPaths:  []string{"/etc", "/sys", "/root"},
		AllowedCommands: []string{"ls", "echo", "cat", "grep"},
		DeniedCommands:  []string{"rm", "dd", "mkfs", "sudo"},
		MaxFileSize: 1024 * 1024, // 1MB
	}

	fmt.Println("📜 Custom Permission Policy:")
	fmt.Printf("   Allowed paths: %v\n", policy.AllowedPaths)
	fmt.Printf("   Denied paths: %v\n", policy.DeniedPaths)
	fmt.Printf("   Allowed commands: %v\n", policy.AllowedCommands)
	fmt.Printf("   Denied commands: %v\n", policy.DeniedCommands)
	fmt.Printf("   Max file size: %d bytes\n\n", policy.MaxFileSize)

	// Test the policy with various scenarios
	scenarios := []struct {
		operation string
		details   map[string]interface{}
	}{
		{
			operation: "write_file",
			details: map[string]interface{}{
				"path": "/tmp/test.txt",
				"size": 100,
			},
		},
		{
			operation: "write_file",
			details: map[string]interface{}{
				"path": "/etc/passwd",
				"size": 100,
			},
		},
		{
			operation: "write_file",
			details: map[string]interface{}{
				"path": "/tmp/huge.bin",
				"size": 2 * 1024 * 1024, // 2MB
			},
		},
		{
			operation: "run_command",
			details: map[string]interface{}{
				"command": "ls -la",
			},
		},
		{
			operation: "run_command",
			details: map[string]interface{}{
				"command": "rm -rf /",
			},
		},
		{
			operation: "edit_file",
			details: map[string]interface{}{
				"path": "/home/user/safe/config.ini",
			},
		},
	}

	fmt.Println("🔍 Policy Evaluation Results:")
	for i, scenario := range scenarios {
		allowed, reason := policy.Evaluate(scenario.operation, scenario.details)
		
		emoji := "✅"
		if !allowed {
			emoji = "❌"
		}
		
		fmt.Printf("\n%d. %s %s\n", i+1, emoji, scenario.operation)
		fmt.Printf("   Details: %v\n", scenario.details)
		fmt.Printf("   Decision: %v\n", allowed)
		fmt.Printf("   Reason: %s\n", reason)
	}

	// Demonstrate permission mode switching based on context
	fmt.Println("\n🔄 Dynamic Permission Mode Switching:")
	
	contexts := []struct {
		context string
		mode    claudecode.PermissionMode
		reason  string
	}{
		{
			context: "Development environment",
			mode:    claudecode.PermissionModeAcceptEdits,
			reason:  "Trust level: High, Environment: Isolated",
		},
		{
			context: "Production environment",
			mode:    claudecode.PermissionModeDefault,
			reason:  "Trust level: Low, Environment: Critical",
		},
		{
			context: "Automated testing",
			mode:    claudecode.PermissionModeBypassPermission,
			reason:  "Trust level: Full, Environment: CI/CD",
		},
		{
			context: "User demonstration",
			mode:    claudecode.PermissionModeDefault,
			reason:  "Trust level: Medium, Environment: Interactive",
		},
	}

	for _, ctx := range contexts {
		fmt.Printf("\n📍 Context: %s\n", ctx.context)
		fmt.Printf("   Selected mode: %s\n", ctx.mode)
		fmt.Printf("   Reason: %s\n", ctx.reason)
		
		// Show how to apply this mode
		options := claudecode.NewClaudeCodeOptions()
		options.PermissionMode = &ctx.mode
		
		fmt.Printf("   Configuration applied ✓\n")
	}

	// Advanced permission handling with audit logging
	fmt.Println("\n📝 Permission Audit Log Example:")
	
	auditLog := []PermissionAuditEntry{
		{
			Timestamp: time.Now().Add(-5 * time.Minute),
			User:      "developer1",
			Operation: "Edit",
			Target:    "/project/src/main.go",
			Allowed:   true,
			Mode:      "accept-edits",
		},
		{
			Timestamp: time.Now().Add(-3 * time.Minute),
			User:      "contractor",
			Operation: "Write",
			Target:    "/etc/shadow",
			Allowed:   false,
			Mode:      "default",
		},
		{
			Timestamp: time.Now().Add(-1 * time.Minute),
			User:      "ci-bot",
			Operation: "Bash",
			Target:    "npm test",
			Allowed:   true,
			Mode:      "bypass",
		},
	}

	for _, entry := range auditLog {
		emoji := "✅"
		if !entry.Allowed {
			emoji = "❌"
		}
		
		fmt.Printf("\n%s [%s] %s\n", emoji, entry.Timestamp.Format("15:04:05"), entry.User)
		fmt.Printf("   Operation: %s on %s\n", entry.Operation, entry.Target)
		fmt.Printf("   Mode: %s, Allowed: %v\n", entry.Mode, entry.Allowed)
	}
}

// Helper types

type PermissionPolicy struct {
	AllowedPaths    []string
	DeniedPaths     []string
	AllowedCommands []string
	DeniedCommands  []string
	MaxFileSize     int
}

func (p *PermissionPolicy) Evaluate(operation string, details map[string]interface{}) (bool, string) {
	switch operation {
	case "write_file", "edit_file":
		path, _ := details["path"].(string)
		size, _ := details["size"].(int)
		
		// Check denied paths
		for _, denied := range p.DeniedPaths {
			if strings.HasPrefix(path, denied) {
				return false, fmt.Sprintf("Path %s is in denied list", path)
			}
		}
		
		// Check allowed paths
		allowed := false
		for _, allowedPath := range p.AllowedPaths {
			if strings.HasPrefix(path, allowedPath) {
				allowed = true
				break
			}
		}
		if !allowed {
			return false, fmt.Sprintf("Path %s is not in allowed list", path)
		}
		
		// Check file size
		if size > p.MaxFileSize {
			return false, fmt.Sprintf("File size %d exceeds maximum %d", size, p.MaxFileSize)
		}
		
		return true, "All checks passed"
		
	case "run_command":
		command, _ := details["command"].(string)
		parts := strings.Fields(command)
		if len(parts) == 0 {
			return false, "Empty command"
		}
		
		cmd := parts[0]
		
		// Check denied commands
		for _, denied := range p.DeniedCommands {
			if cmd == denied {
				return false, fmt.Sprintf("Command %s is denied", cmd)
			}
		}
		
		// Check allowed commands
		for _, allowed := range p.AllowedCommands {
			if cmd == allowed {
				return true, "Command is allowed"
			}
		}
		
		return false, fmt.Sprintf("Command %s is not in allowed list", cmd)
		
	default:
		return false, "Unknown operation"
	}
}

type PermissionAuditEntry struct {
	Timestamp time.Time
	User      string
	Operation string
	Target    string
	Allowed   bool
	Mode      string
}