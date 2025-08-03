package client

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func TestProjectContextManager_Basic(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "go-claude-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a basic test client
	config := types.NewClaudeCodeConfig()
	config.WorkingDirectory = tempDir
	config.TestMode = true

	client, err := NewClaudeCodeClient(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Create project context manager
	manager := NewProjectContextManager(client)

	// Test basic context retrieval
	ctx := context.Background()
	projectContext, err := manager.GetEnhancedProjectContext(ctx)
	if err != nil {
		t.Fatalf("Failed to get enhanced project context: %v", err)
	}

	// Basic checks - simplified to match official SDK scope
	if projectContext.WorkingDirectory != tempDir {
		t.Errorf("Expected working directory %s, got %s", tempDir, projectContext.WorkingDirectory)
	}

	// Note: Official SDKs only provide working directory
	// Complex analysis features (Language, Framework, Metadata) were removed
	// to match official API surface

	// Check cache is populated
	cacheInfo := manager.GetCacheInfo()
	if !cacheInfo["is_cached"].(bool) {
		t.Error("Expected context to be cached")
	}
}

func TestProjectContextManager_Cache(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "go-claude-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a basic test client
	config := types.NewClaudeCodeConfig()
	config.WorkingDirectory = tempDir
	config.TestMode = true

	client, err := NewClaudeCodeClient(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Create project context manager with short cache duration
	manager := NewProjectContextManager(client)
	manager.SetCacheDuration(100 * time.Millisecond)

	ctx := context.Background()

	// First call should populate cache
	_, err = manager.GetEnhancedProjectContext(ctx)
	if err != nil {
		t.Fatalf("Failed to get enhanced project context: %v", err)
	}

	cacheInfo := manager.GetCacheInfo()
	if !cacheInfo["is_cached"].(bool) {
		t.Error("Expected context to be cached after first call")
	}

	// Second call should use cache
	_, err = manager.GetEnhancedProjectContext(ctx)
	if err != nil {
		t.Fatalf("Failed to get enhanced project context from cache: %v", err)
	}

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Third call should refresh cache
	_, err = manager.GetEnhancedProjectContext(ctx)
	if err != nil {
		t.Fatalf("Failed to get enhanced project context after cache expiry: %v", err)
	}
}

func TestProjectContextManager_InvalidateCache(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "go-claude-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a basic test client
	config := types.NewClaudeCodeConfig()
	config.WorkingDirectory = tempDir
	config.TestMode = true

	client, err := NewClaudeCodeClient(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Create project context manager
	manager := NewProjectContextManager(client)

	ctx := context.Background()

	// First call should populate cache
	_, err = manager.GetEnhancedProjectContext(ctx)
	if err != nil {
		t.Fatalf("Failed to get enhanced project context: %v", err)
	}

	cacheInfo := manager.GetCacheInfo()
	if !cacheInfo["is_cached"].(bool) {
		t.Error("Expected context to be cached")
	}

	// Invalidate cache
	manager.InvalidateCache()

	cacheInfo = manager.GetCacheInfo()
	if cacheInfo["is_cached"].(bool) {
		t.Error("Expected cache to be invalidated")
	}

	// Next call should repopulate cache
	_, err = manager.GetEnhancedProjectContext(ctx)
	if err != nil {
		t.Fatalf("Failed to get enhanced project context after invalidation: %v", err)
	}

	cacheInfo = manager.GetCacheInfo()
	if !cacheInfo["is_cached"].(bool) {
		t.Error("Expected context to be cached after invalidation and refetch")
	}
}

func TestProjectContextManager_AbsolutePath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "go-claude-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with relative path - create test directory first
	relativeDir := "./test"
	if err := os.MkdirAll(relativeDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(relativeDir)

	absDir, _ := filepath.Abs(relativeDir)

	// Create a basic test client
	config := types.NewClaudeCodeConfig()
	config.WorkingDirectory = relativeDir
	config.TestMode = true

	client, err := NewClaudeCodeClient(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Create project context manager
	manager := NewProjectContextManager(client)

	ctx := context.Background()

	// Get context - should convert to absolute path
	projectContext, err := manager.GetEnhancedProjectContext(ctx)
	if err != nil {
		t.Fatalf("Failed to get enhanced project context: %v", err)
	}

	// Check that working directory is now absolute
	if projectContext.WorkingDirectory != absDir {
		t.Errorf("Expected absolute directory %s, got %s", absDir, projectContext.WorkingDirectory)
	}
}
