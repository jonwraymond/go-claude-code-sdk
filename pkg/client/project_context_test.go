package client

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func TestNewProjectContextManager(t *testing.T) {
	tempDir := t.TempDir()
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	manager := client.ProjectContext()
	if manager == nil {
		t.Fatal("Project context manager should not be nil")
	}

	// Check initial cache info
	cacheInfo := manager.GetCacheInfo()
	if cacheInfo["is_cached"].(bool) {
		t.Error("Expected cache to be empty initially")
	}
}

func TestProjectContextManager_GetEnhancedProjectContext(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a sample Go project structure
	setupGoProject(t, tempDir)
	
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	manager := client.ProjectContext()

	// Get enhanced context
	context, err := manager.GetEnhancedProjectContext(ctx)
	if err != nil {
		t.Fatalf("Failed to get enhanced project context: %v", err)
	}

	// Basic checks
	if context.WorkingDirectory != tempDir {
		t.Errorf("Expected working directory %s, got %s", tempDir, context.WorkingDirectory)
	}

	if context.Language != "Go" {
		t.Errorf("Expected language Go, got %s", context.Language)
	}

	if context.Framework != "Go Modules" {
		t.Errorf("Expected framework 'Go Modules', got %s", context.Framework)
	}

	// Check enhanced metadata
	if context.Metadata == nil {
		t.Fatal("Expected metadata to be populated")
	}

	// Should have architecture info
	if _, exists := context.Metadata["architecture"]; !exists {
		t.Error("Expected architecture metadata")
	}

	// Should have dependencies info
	if _, exists := context.Metadata["dependencies"]; !exists {
		t.Error("Expected dependencies metadata")
	}

	// Check cache is populated
	cacheInfo := manager.GetCacheInfo()
	if !cacheInfo["is_cached"].(bool) {
		t.Error("Expected cache to be populated after first call")
	}
}

func TestProjectContextManager_CacheManagement(t *testing.T) {
	tempDir := t.TempDir()
	setupGoProject(t, tempDir)
	
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	manager := client.ProjectContext()

	// Set short cache duration for testing
	manager.SetCacheDuration(100 * time.Millisecond)

	// First call - should populate cache
	context1, err := manager.GetEnhancedProjectContext(ctx)
	if err != nil {
		t.Fatalf("Failed to get enhanced project context: %v", err)
	}

	// Second call immediately - should use cache
	start := time.Now()
	context2, err := manager.GetEnhancedProjectContext(ctx)
	if err != nil {
		t.Fatalf("Failed to get enhanced project context: %v", err)
	}
	duration := time.Since(start)

	// Should be very fast (cached)
	if duration > 10*time.Millisecond {
		t.Errorf("Second call took too long (%v), should be cached", duration)
	}

	// Should be the same instance (pointer equality)
	if context1 != context2 {
		t.Error("Expected cached context to be the same instance")
	}

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Third call - should repopulate cache
	context3, err := manager.GetEnhancedProjectContext(ctx)
	if err != nil {
		t.Fatalf("Failed to get enhanced project context: %v", err)
	}

	// Should be a different instance (cache expired)
	if context1 == context3 {
		t.Error("Expected expired cache to create new context instance")
	}

	// Test manual cache invalidation
	manager.InvalidateCache()
	cacheInfo := manager.GetCacheInfo()
	if cacheInfo["is_cached"].(bool) {
		t.Error("Expected cache to be invalidated")
	}
}

func TestProjectContextManager_AnalyzeArchitecture(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create microservices architecture
	os.MkdirAll(filepath.Join(tempDir, "services", "user"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "services", "order"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "api"), 0755)
	os.WriteFile(filepath.Join(tempDir, "docker-compose.yml"), []byte("version: '3'"), 0644)
	
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	manager := client.ProjectContext()
	
	// Analyze architecture
	arch, err := manager.analyzeArchitecture(tempDir)
	if err != nil {
		t.Fatalf("Failed to analyze architecture: %v", err)
	}

	if arch.Pattern != "microservices" {
		t.Errorf("Expected microservices pattern, got %s", arch.Pattern)
	}

	if len(arch.ConfigFiles) == 0 {
		t.Error("Expected to find configuration files")
	}

	// Check for docker-compose.yml
	found := false
	for _, file := range arch.ConfigFiles {
		if file == "docker-compose.yml" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find docker-compose.yml in config files")
	}
}

func TestProjectContextManager_AnalyzeDependencies(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(string)
		expectedMgr    string
		expectedModule string
	}{
		{
			name:        "Go project",
			setupFunc:   func(dir string) { setupGoProject(nil, dir) },
			expectedMgr: "go",
		},
		{
			name:        "Node.js project",
			setupFunc:   setupNodeProject,
			expectedMgr: "npm",
		},
		{
			name:        "Python project",
			setupFunc:   setupPythonProject,
			expectedMgr: "pip",
		},
		{
			name:        "Rust project",
			setupFunc:   setupRustProject,
			expectedMgr: "cargo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tt.setupFunc(tempDir)

			config := &types.ClaudeCodeConfig{
				WorkingDirectory: tempDir,
			}

			ctx := context.Background()
			client, err := NewClaudeCodeClient(ctx, config)
			if err != nil {
				t.Fatalf("Failed to create Claude Code client: %v", err)
			}
			defer client.Close()

			manager := client.ProjectContext()

			deps, err := manager.analyzeDependencies(tempDir)
			if err != nil {
				t.Fatalf("Failed to analyze dependencies: %v", err)
			}

			if deps.Manager != tt.expectedMgr {
				t.Errorf("Expected manager %s, got %s", tt.expectedMgr, deps.Manager)
			}

			if len(deps.Dependencies) == 0 {
				t.Error("Expected to find dependencies")
			}
		})
	}
}

func TestProjectContextManager_AnalyzeDevTools(t *testing.T) {
	tempDir := t.TempDir()
	
	// Setup development tools
	os.WriteFile(filepath.Join(tempDir, "Dockerfile"), []byte("FROM node:16"), 0644)
	os.WriteFile(filepath.Join(tempDir, "docker-compose.yml"), []byte("version: '3'"), 0644)
	os.MkdirAll(filepath.Join(tempDir, ".github", "workflows"), 0755)
	os.WriteFile(filepath.Join(tempDir, ".github", "workflows", "ci.yml"), []byte("name: CI"), 0644)
	os.WriteFile(filepath.Join(tempDir, ".eslintrc.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(tempDir, ".prettierrc"), []byte("{}"), 0644)
	
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	manager := client.ProjectContext()

	tools, err := manager.analyzeDevTools(tempDir)
	if err != nil {
		t.Fatalf("Failed to analyze dev tools: %v", err)
	}

	// Check for Docker
	if dockerInfo, exists := tools["docker"]; exists {
		dockerMap := dockerInfo.(map[string]interface{})
		if !dockerMap["dockerfile"].(bool) {
			t.Error("Expected to detect Dockerfile")
		}
		if !dockerMap["compose"].(bool) {
			t.Error("Expected to detect docker-compose.yml")
		}
	} else {
		t.Error("Expected to detect Docker tools")
	}

	// Check for CI/CD
	if cicdInfo, exists := tools["ci_cd"]; exists {
		cicdMap := cicdInfo.(map[string]interface{})
		if !cicdMap["github_actions"].(bool) {
			t.Error("Expected to detect GitHub Actions")
		}
	} else {
		t.Error("Expected to detect CI/CD tools")
	}

	// Check for linting
	if lintInfo, exists := tools["linting"]; exists {
		lintMap := lintInfo.(map[string]interface{})
		if !lintMap["eslint"].(bool) {
			t.Error("Expected to detect ESLint")
		}
	} else {
		t.Error("Expected to detect linting tools")
	}

	// Check for formatting
	if formatInfo, exists := tools["formatting"]; exists {
		formatMap := formatInfo.(map[string]interface{})
		if !formatMap["prettier"].(bool) {
			t.Error("Expected to detect Prettier")
		}
	} else {
		t.Error("Expected to detect formatting tools")
	}
}

func TestProjectContextManager_AnalyzeCodePatterns(t *testing.T) {
	tempDir := t.TempDir()
	
	// Setup test files
	os.MkdirAll(filepath.Join(tempDir, "tests"), 0755)
	os.WriteFile(filepath.Join(tempDir, "main_test.go"), []byte("package main\n\nfunc TestMain(t *testing.T) {}"), 0644)
	os.WriteFile(filepath.Join(tempDir, "tests", "integration_test.go"), []byte("package tests"), 0644)
	
	// Setup API files
	os.MkdirAll(filepath.Join(tempDir, "api"), 0755)
	os.WriteFile(filepath.Join(tempDir, "api", "handler.go"), []byte("package api"), 0644)
	os.WriteFile(filepath.Join(tempDir, "router.go"), []byte("package main"), 0644)
	
	// Setup Go mod for dependencies
	goMod := `module test-project

go 1.21

require (
	github.com/gin-gonic/gin v1.9.0
	github.com/stretchr/testify v1.8.0
)`
	os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644)
	
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	manager := client.ProjectContext()

	patterns, err := manager.analyzeCodePatterns(tempDir)
	if err != nil {
		t.Fatalf("Failed to analyze code patterns: %v", err)
	}

	// Check testing pattern
	if testingPattern, exists := patterns["testing"]; exists {
		if testingPattern.Confidence <= 0 {
			t.Error("Expected testing pattern to have confidence > 0")
		}
		if len(testingPattern.Evidence) == 0 {
			t.Error("Expected testing pattern to have evidence")
		}
	} else {
		t.Error("Expected to detect testing pattern")
	}

	// Check API pattern
	if apiPattern, exists := patterns["api"]; exists {
		if apiPattern.Confidence <= 0 {
			t.Error("Expected API pattern to have confidence > 0")
		}
		if len(apiPattern.Evidence) == 0 {
			t.Error("Expected API pattern to have evidence")
		}
	} else {
		t.Error("Expected to detect API pattern")
	}
}

func TestClaudeCodeClient_ProjectContextIntegration(t *testing.T) {
	tempDir := t.TempDir()
	setupGoProject(t, tempDir)
	
	config := &types.ClaudeCodeConfig{
		WorkingDirectory: tempDir,
	}

	ctx := context.Background()
	client, err := NewClaudeCodeClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Claude Code client: %v", err)
	}
	defer client.Close()

	// Test client project context methods
	manager := client.ProjectContext()
	if manager == nil {
		t.Fatal("Project context manager should not be nil")
	}

	// Get enhanced context through client
	enhancedContext, err := client.GetEnhancedProjectContext(ctx)
	if err != nil {
		t.Fatalf("Failed to get enhanced project context: %v", err)
	}

	if enhancedContext.Language != "Go" {
		t.Errorf("Expected language Go, got %s", enhancedContext.Language)
	}

	// Test cache management through client
	cacheInfo := client.GetProjectContextCacheInfo()
	if !cacheInfo["is_cached"].(bool) {
		t.Error("Expected cache to be populated")
	}

	// Test cache duration setting
	client.SetProjectContextCacheDuration(1 * time.Second)
	newCacheInfo := client.GetProjectContextCacheInfo()
	if newCacheInfo["cache_duration"].(string) != "1s" {
		t.Error("Expected cache duration to be updated")
	}

	// Test cache invalidation
	client.InvalidateProjectContextCache()
	invalidatedCacheInfo := client.GetProjectContextCacheInfo()
	if invalidatedCacheInfo["is_cached"].(bool) {
		t.Error("Expected cache to be invalidated")
	}
}

// Helper functions to setup test projects

func setupGoProject(t *testing.T, dir string) {
	if t != nil {
		t.Helper()
	}
	
	goMod := `module test-project

go 1.21

require (
	github.com/gin-gonic/gin v1.9.0
	github.com/stretchr/testify v1.8.0
)`
	
	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`
	
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0644)
}

func setupNodeProject(dir string) {
	packageJson := `{
  "name": "test-project",
  "version": "1.0.0",
  "dependencies": {
    "express": "^4.18.0",
    "lodash": "^4.17.21"
  },
  "devDependencies": {
    "jest": "^29.0.0",
    "eslint": "^8.0.0"
  },
  "scripts": {
    "start": "node index.js",
    "test": "jest"
  }
}`
	
	indexJs := `const express = require('express');
const app = express();

app.get('/', (req, res) => {
    res.send('Hello World!');
});

app.listen(3000);`
	
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJson), 0644)
	os.WriteFile(filepath.Join(dir, "index.js"), []byte(indexJs), 0644)
}

func setupPythonProject(dir string) {
	requirements := `flask==2.3.0
requests==2.28.0
pytest==7.4.0`
	
	mainPy := `from flask import Flask

app = Flask(__name__)

@app.route('/')
def hello():
    return 'Hello, World!'

if __name__ == '__main__':
    app.run()`
	
	os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte(requirements), 0644)
	os.WriteFile(filepath.Join(dir, "main.py"), []byte(mainPy), 0644)
}

func setupRustProject(dir string) {
	cargoToml := `[package]
name = "test-project"
version = "0.1.0"
edition = "2021"

[dependencies]
serde = { version = "1.0", features = ["derive"] }
tokio = { version = "1.0", features = ["full"] }

[dev-dependencies]
tokio-test = "0.4"`
	
	mainRs := `fn main() {
    println!("Hello, world!");
}`
	
	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(cargoToml), 0644)
	os.WriteFile(filepath.Join(dir, "src", "main.rs"), []byte(mainRs), 0644)
}