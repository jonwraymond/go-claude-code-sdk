/*
Package auth provides authentication mechanisms for the Claude Code Go SDK.

This package handles all authentication concerns for interacting with Claude Code,
including API key management, environment variable support, and secure credential
storage. It integrates with the Claude Code CLI's authentication system.

# Authentication Methods

The package supports multiple authentication approaches:

	// API Key authentication (most common)
	auth := auth.NewAPIKeyAuth("sk-ant-...")

	// Bearer token authentication
	auth := auth.NewBearerTokenAuth("bearer-token")

	// Environment variable authentication (automatic)
	auth := auth.NewFromEnvironment()

# Configuration Integration

Authentication integrates seamlessly with SDK configuration:

	config := types.NewClaudeCodeConfig()
	config.APIKey = "your-api-key"  // Automatically creates APIKeyAuth

	// Or use environment variables
	// export ANTHROPIC_API_KEY="sk-ant-..."
	config := types.NewConfigFromEnvironment()

# Credential Manager

Secure credential storage and retrieval:

	manager := auth.NewCredentialManager()

	// Store credentials securely
	err := manager.StoreCredential("my-project", &auth.Credential{
		Type:  auth.CredentialTypeAPIKey,
		Value: "sk-ant-...",
	})

	// Retrieve credentials
	cred, err := manager.GetCredential("my-project")

	// List all stored credentials
	creds := manager.ListCredentials()

# Environment Variables

The package recognizes standard environment variables:

	ANTHROPIC_API_KEY    // Primary API key
	CLAUDE_API_KEY       // Alternative API key name
	CLAUDE_API_BASE_URL  // Custom API base URL

# Security Best Practices

The auth package implements security best practices:

	// API keys are never logged or included in errors
	err := auth.ValidateAPIKey(apiKey)
	// Error message will not contain the actual key

	// Credentials are stored with appropriate permissions
	// Configuration files use 0600 permissions

	// Sensitive data is sanitized in string representations
	fmt.Println(auth.String()) // Shows "sk-ant-...XXXX"

# Validation

Built-in validation for authentication credentials:

	// Validate API key format
	if err := auth.ValidateAPIKey(apiKey); err != nil {
		// Handle invalid key format
	}

	// Validate bearer token
	if err := auth.ValidateBearerToken(token); err != nil {
		// Handle invalid token
	}

# Authentication Headers

The package handles HTTP header generation:

	auth := auth.NewAPIKeyAuth("sk-ant-...")
	headers := auth.Headers()
	// Returns: {"X-API-Key": "sk-...", "api-version": "..."}

# Multi-Project Support

Support for different credentials per project:

	manager := auth.NewCredentialManager()

	// Store project-specific credentials
	manager.StoreCredential("project-1", &auth.Credential{
		Type:  auth.CredentialTypeAPIKey,
		Value: "sk-ant-project1-...",
	})

	manager.StoreCredential("project-2", &auth.Credential{
		Type:  auth.CredentialTypeAPIKey,
		Value: "sk-ant-project2-...",
	})

	// Use appropriate credentials based on context
	cred := manager.GetCredential(currentProject)

# Error Handling

Authentication errors are clearly categorized:

	auth, err := auth.NewFromEnvironment()
	if err != nil {
		switch err.(type) {
		case *auth.MissingCredentialsError:
			// No credentials found
		case *auth.InvalidCredentialsError:
			// Credentials are malformed
		default:
			// Other error
		}
	}

# Integration with Claude Code CLI

The auth package ensures compatibility with Claude Code CLI's authentication:

	// Credentials work seamlessly with CLI subprocess
	config := types.NewClaudeCodeConfig()
	config.APIKey = getAPIKey()

	// The SDK passes credentials to Claude Code CLI
	// via environment variables or command flags

# Example Usage

	// Simple API key authentication
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY not set")
	}

	config := types.NewClaudeCodeConfig()
	config.APIKey = apiKey

	client, err := client.NewClaudeCodeClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// Using credential manager
	manager := auth.NewCredentialManager()
	cred, err := manager.GetCredential("my-project")
	if err != nil {
		// Fall back to environment
		cred = &auth.Credential{
			Type:  auth.CredentialTypeAPIKey,
			Value: os.Getenv("ANTHROPIC_API_KEY"),
		}
	}

	config.APIKey = cred.Value

The auth package provides secure, flexible authentication for the Claude Code Go SDK
while maintaining compatibility with the Claude Code CLI authentication system.
*/
package auth
