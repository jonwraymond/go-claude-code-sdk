//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ProjectContextIntegrationSuite tests project context detection with Claude Code CLI
type ProjectContextIntegrationSuite struct {
	suite.Suite
	client                *client.ClaudeCodeClient
	projectContextManager *client.ProjectContextManager
	config                *types.ClaudeCodeConfig
	testDir               string
}

func (s *ProjectContextIntegrationSuite) SetupSuite() {
	// Skip if not in integration test mode
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		s.T().Skip("Integration tests disabled. Set INTEGRATION_TESTS=true to run")
	}

	// Ensure API key is set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		s.T().Skip("ANTHROPIC_API_KEY not set")
	}

	// Create test directory
	var err error
	s.testDir, err = os.MkdirTemp("", "claude-code-context-test-*")
	require.NoError(s.T(), err)

	// Create config
	s.config = types.NewClaudeCodeConfig()
	s.config.APIKey = apiKey
	s.config.ClaudeExecutable = "claude"
	s.config.WorkingDirectory = s.testDir
	s.config.Timeout = 30 * time.Second
	
	// Enable TestMode in CI environment to skip Claude Code CLI requirement
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		s.config.TestMode = true
	}

	// Create client
	ctx := context.Background()
	s.client, err = client.NewClaudeCodeClient(ctx, s.config)
	require.NoError(s.T(), err)

	// Get project context manager
	s.projectContextManager = s.client.ProjectContext()
}

func (s *ProjectContextIntegrationSuite) TearDownSuite() {
	if s.client != nil {
		s.client.Close()
	}
	if s.testDir != "" {
		os.RemoveAll(s.testDir)
	}
}

func (s *ProjectContextIntegrationSuite) TestDetectGoProject() {
	ctx := context.Background()

	// Create a Go project structure
	goMod := `module github.com/test/project

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/stretchr/testify v1.8.4
)`

	mainGo := `package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello World"})
    })
    r.Run()
}`

	// Create files
	err := os.WriteFile(filepath.Join(s.testDir, "go.mod"), []byte(goMod), 0644)
	require.NoError(s.T(), err)
	
	err = os.WriteFile(filepath.Join(s.testDir, "main.go"), []byte(mainGo), 0644)
	require.NoError(s.T(), err)

	// Detect project context
	context, err := s.projectContextManager.GetEnhancedProjectContext(ctx)
	require.NoError(s.T(), err)

	// Verify we got a context back
	assert.NotNil(s.T(), context)
	assert.NotEmpty(s.T(), context.WorkingDirectory)
}

func (s *ProjectContextIntegrationSuite) TestDetectPythonProject() {
	ctx := context.Background()

	// Create a Python project structure
	requirementsTxt := `Flask==2.3.0
pytest==7.4.0
requests==2.31.0`

	appPy := `from flask import Flask, jsonify

app = Flask(__name__)

@app.route('/')
def hello_world():
    return jsonify({"message": "Hello World"})

if __name__ == '__main__':
    app.run(debug=True)`

	// Create files
	err := os.WriteFile(filepath.Join(s.testDir, "requirements.txt"), []byte(requirementsTxt), 0644)
	require.NoError(s.T(), err)
	
	err = os.WriteFile(filepath.Join(s.testDir, "app.py"), []byte(appPy), 0644)
	require.NoError(s.T(), err)

	// Detect project context
	context, err := s.projectContextManager.GetEnhancedProjectContext(ctx)
	require.NoError(s.T(), err)

	// Verify we got a context back
	assert.NotNil(s.T(), context)
	assert.NotEmpty(s.T(), context.WorkingDirectory)
}

func (s *ProjectContextIntegrationSuite) TestDetectNodeProject() {
	ctx := context.Background()

	// Create a Node.js project structure
	packageJSON := `{
  "name": "test-project",
  "version": "1.0.0",
  "description": "Test Node.js project",
  "main": "index.js",
  "scripts": {
    "start": "node index.js",
    "test": "jest"
  },
  "dependencies": {
    "express": "^4.18.2",
    "axios": "^1.4.0"
  },
  "devDependencies": {
    "jest": "^29.5.0",
    "eslint": "^8.42.0"
  }
}`

	indexJS := `const express = require('express');
const app = express();
const port = 3000;

app.get('/', (req, res) => {
  res.json({ message: 'Hello World' });
});

app.listen(port, () => {
  console.log('Server running on port ' + port);
});`

	// Create files
	err := os.WriteFile(filepath.Join(s.testDir, "package.json"), []byte(packageJSON), 0644)
	require.NoError(s.T(), err)
	
	err = os.WriteFile(filepath.Join(s.testDir, "index.js"), []byte(indexJS), 0644)
	require.NoError(s.T(), err)

	// Detect project context
	context, err := s.projectContextManager.GetEnhancedProjectContext(ctx)
	require.NoError(s.T(), err)

	// Verify we got a context back
	assert.NotNil(s.T(), context)
	assert.NotEmpty(s.T(), context.WorkingDirectory)
}

func (s *ProjectContextIntegrationSuite) TestGetEnhancedProjectContext() {
	ctx := context.Background()

	// Create a multi-file Go project
	files := map[string]string{
		"go.mod": `module github.com/test/enhanced

go 1.21

require github.com/gorilla/mux v1.8.0`,
		"main.go": `package main

import "github.com/gorilla/mux"

func main() {
    r := mux.NewRouter()
    // Router setup
}`,
		"handlers/user.go": `package handlers

import "net/http"

func GetUser(w http.ResponseWriter, r *http.Request) {
    // Handler implementation
}`,
		"models/user.go": `package models

type User struct {
    ID    string
    Name  string
    Email string
}`,
		"README.md": `# Enhanced Test Project

A Go web service using Gorilla Mux.`,
	}

	// Create directory structure
	err := os.MkdirAll(filepath.Join(s.testDir, "handlers"), 0755)
	require.NoError(s.T(), err)
	err = os.MkdirAll(filepath.Join(s.testDir, "models"), 0755)
	require.NoError(s.T(), err)

	// Create files
	for path, content := range files {
		err := os.WriteFile(filepath.Join(s.testDir, path), []byte(content), 0644)
		require.NoError(s.T(), err)
	}

	// Get enhanced context
	context, err := s.projectContextManager.GetEnhancedProjectContext(ctx)
	require.NoError(s.T(), err)

	// Verify we got a context back
	assert.NotNil(s.T(), context)
	assert.NotEmpty(s.T(), context.WorkingDirectory)
}

func (s *ProjectContextIntegrationSuite) TestAnalyzeCodeStructure() {
	ctx := context.Background()

	// Create a structured project
	structureFiles := map[string]string{
		"cmd/api/main.go": `package main

func main() {
    // API server
}`,
		"internal/service/user.go": `package service

type UserService struct{}

func (s *UserService) GetUser(id string) error {
    return nil
}`,
		"internal/repository/user.go": `package repository

type UserRepository interface {
    FindByID(id string) (*User, error)
}`,
		"pkg/utils/validation.go": `package utils

func ValidateEmail(email string) bool {
    return true
}`,
	}

	// Create directories
	dirs := []string{"cmd/api", "internal/service", "internal/repository", "pkg/utils"}
	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(s.testDir, dir), 0755)
		require.NoError(s.T(), err)
	}

	// Create files
	for path, content := range structureFiles {
		err := os.WriteFile(filepath.Join(s.testDir, path), []byte(content), 0644)
		require.NoError(s.T(), err)
	}

	// Get project context
	context, err := s.projectContextManager.GetEnhancedProjectContext(ctx)
	require.NoError(s.T(), err)

	// Verify we got a context back
	assert.NotNil(s.T(), context)
	assert.NotEmpty(s.T(), context.WorkingDirectory)
}

func (s *ProjectContextIntegrationSuite) TestProjectContextWithQuery() {
	ctx := context.Background()

	// Create a simple Go project
	goMod := `module github.com/test/context-query

go 1.21`
	
	mainGo := `package main

import "fmt"

func calculateSum(a, b int) int {
    return a + b
}

func main() {
    result := calculateSum(5, 3)
    fmt.Printf("Sum: %d\n", result)
}`

	err := os.WriteFile(filepath.Join(s.testDir, "go.mod"), []byte(goMod), 0644)
	require.NoError(s.T(), err)
	err = os.WriteFile(filepath.Join(s.testDir, "main.go"), []byte(mainGo), 0644)
	require.NoError(s.T(), err)

	// Get project context
	projectCtx, err := s.projectContextManager.GetEnhancedProjectContext(ctx)
	require.NoError(s.T(), err)

	// Use project context in a query
	options := &client.QueryOptions{
		SystemPrompt: "You are helping with a Go project",
		MaxTokens:    500,
	}

	// Query about the project
	result, err := s.client.QueryMessagesSync(ctx, 
		"What does the calculateSum function do in this project?", 
		options)
	require.NoError(s.T(), err)

	// Should understand the project context
	assert.Contains(s.T(), result.Content, "calculateSum")
	assert.Contains(s.T(), result.Content, "add")
}

func TestProjectContextIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ProjectContextIntegrationSuite))
}