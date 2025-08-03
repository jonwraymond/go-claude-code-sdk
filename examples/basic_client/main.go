// Package main demonstrates basic Claude Code client initialization
// and configuration options. This example shows the fundamental
// patterns for creating and configuring a Claude Code client.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Claude Code Client Initialization Examples ===")

	// Example 1: Basic client with minimal configuration
	basicClientExample()

	// Example 2: Client with API key authentication
	apiKeyClientExample()

	// Example 3: Client with subscription authentication
	subscriptionClientExample()

	// Example 4: Client with custom configuration
	customConfigurationExample()

	// Example 5: Client with working directory
	workingDirectoryExample()

	// Example 6: Client with environment variables
	environmentVariablesExample()
}

// basicClientExample demonstrates the simplest way to create a client
func basicClientExample() {
	fmt.Println("--- Example 1: Basic Client ---")

	ctx := context.Background()

	// Create a basic configuration
	config := types.NewClaudeCodeConfig()

	// Create the client
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create basic client: %v", err)
		return
	}
	defer claudeClient.Close()

	fmt.Printf("✓ Basic client created successfully\n")
	fmt.Printf("  Working Directory: %s\n", config.WorkingDirectory)
	fmt.Printf("  Model: %s\n", config.Model)
	fmt.Printf("  Session ID: %s\n", config.SessionID)
	fmt.Println()
}

// apiKeyClientExample demonstrates API key authentication
func apiKeyClientExample() {
	fmt.Println("--- Example 2: API Key Authentication ---")

	ctx := context.Background()

	// Create configuration with API key
	config := types.NewClaudeCodeConfig()
	
	// Try to get API key from environment
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
		config.AuthMethod = types.AuthTypeAPIKey
		fmt.Printf("✓ Using API key from environment\n")
	} else {
		// Use a placeholder for demonstration
		config.APIKey = "sk-ant-api03-..." // Your actual API key
		config.AuthMethod = types.AuthTypeAPIKey
		fmt.Printf("⚠ Using placeholder API key (set ANTHROPIC_API_KEY environment variable)\n")
	}

	// Create the client
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create API key client: %v", err)
		return
	}
	defer claudeClient.Close()

	fmt.Printf("✓ API key client created successfully\n")
	fmt.Printf("  Auth Method: %s\n", config.AuthMethod)
	fmt.Println()
}

// subscriptionClientExample demonstrates subscription authentication
func subscriptionClientExample() {
	fmt.Println("--- Example 3: Subscription Authentication ---")

	ctx := context.Background()

	// Create configuration with subscription auth
	config := types.NewClaudeCodeConfig()
	config.AuthMethod = types.AuthTypeSubscription

	// Create the client
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create subscription client: %v", err)
		return
	}
	defer claudeClient.Close()

	fmt.Printf("✓ Subscription client created successfully\n")
	fmt.Printf("  Auth Method: %s\n", config.AuthMethod)
	fmt.Printf("  Note: Uses Claude Code CLI's built-in subscription authentication\n")
	fmt.Println()
}

// customConfigurationExample demonstrates custom configuration options
func customConfigurationExample() {
	fmt.Println("--- Example 4: Custom Configuration ---")

	ctx := context.Background()

	// Create a custom configuration
	config := &types.ClaudeCodeConfig{
		Model:           "claude-3-opus-20240229",
		SessionID:       "my-custom-session-" + fmt.Sprintf("%d", time.Now().Unix()),
		AuthMethod:      types.AuthTypeAPIKey,
		APIKey:          os.Getenv("ANTHROPIC_API_KEY"),
		WorkingDirectory: "/tmp", // Custom working directory
		Environment: map[string]string{
			"CUSTOM_VAR": "custom_value",
			"DEBUG":      "true",
		},
		Debug:    true,
		TestMode: false,
	}

	// Apply defaults to fill in missing values
	config.ApplyDefaults()

	// Create the client
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create custom client: %v", err)
		return
	}
	defer claudeClient.Close()

	fmt.Printf("✓ Custom client created successfully\n")
	fmt.Printf("  Model: %s\n", config.Model)
	fmt.Printf("  Working Directory: %s\n", config.WorkingDirectory)
	fmt.Printf("  Debug: %t\n", config.Debug)
	fmt.Printf("  Environment Variables: %d custom vars\n", len(config.Environment))
	fmt.Println()
}

// workingDirectoryExample demonstrates setting a specific working directory
func workingDirectoryExample() {
	fmt.Println("--- Example 5: Working Directory Configuration ---")

	ctx := context.Background()

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get current directory: %v", err)
		return
	}

	// Create configuration with specific working directory
	config := types.NewClaudeCodeConfig()
	config.WorkingDirectory = currentDir

	// Create the client
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create client with working directory: %v", err)
		return
	}
	defer claudeClient.Close()

	// Get project context to verify working directory
	projectCtx, err := claudeClient.GetProjectContext(ctx)
	if err != nil {
		log.Printf("Failed to get project context: %v", err)
		return
	}

	fmt.Printf("✓ Client with working directory created successfully\n")
	fmt.Printf("  Configured Directory: %s\n", config.WorkingDirectory)
	fmt.Printf("  Project Context Directory: %s\n", projectCtx.WorkingDirectory)

	// Demonstrate changing working directory
	newDir := "/tmp"
	err = claudeClient.SetWorkingDirectory(ctx, newDir)
	if err != nil {
		log.Printf("Failed to change working directory: %v", err)
	} else {
		fmt.Printf("✓ Working directory changed to: %s\n", newDir)
	}
	fmt.Println()
}

// environmentVariablesExample demonstrates environment variable configuration
func environmentVariablesExample() {
	fmt.Println("--- Example 6: Environment Variables ---")

	ctx := context.Background()

	// Create configuration with custom environment
	config := types.NewClaudeCodeConfig()
	
	// Add custom environment variables
	config.Environment = map[string]string{
		"CLAUDE_DEBUG":      "true",
		"CLAUDE_LOG_LEVEL":  "info",
		"PROJECT_NAME":      "go-claude-sdk-examples",
		"CUSTOM_TOOL_PATH":  "/usr/local/bin/custom-tools",
	}

	// Also demonstrate reading from actual environment
	if debugVal := os.Getenv("DEBUG"); debugVal != "" {
		config.Environment["DEBUG"] = debugVal
		config.Debug = debugVal == "true"
	}

	// Create the client
	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create client with environment variables: %v", err)
		return
	}
	defer claudeClient.Close()

	fmt.Printf("✓ Client with environment variables created successfully\n")
	fmt.Printf("  Custom Environment Variables:\n")
	for key, value := range config.Environment {
		fmt.Printf("    %s=%s\n", key, value)
	}
	fmt.Println()
}