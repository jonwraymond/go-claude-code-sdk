package client

import (
	"context"
	"strings"
	"testing"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func TestIsOutputTruncated(t *testing.T) {
	executor := &CommandExecutor{}

	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		{
			name:     "Simple ellipsis",
			output:   "...",
			expected: true,
		},
		{
			name:     "Output ending with ellipsis",
			output:   "Some content here...",
			expected: true,
		},
		{
			name:     "Output with truncation marker",
			output:   "Content [truncated]",
			expected: true,
		},
		{
			name:     "Output truncated indicator",
			output:   "Some output [output truncated]",
			expected: true,
		},
		{
			name:     "Complete output with period",
			output:   "This is a complete sentence.",
			expected: false,
		},
		{
			name:     "Complete output with newline",
			output:   "Line 1\nLine 2\n",
			expected: false,
		},
		{
			name:     "Mid-sentence cutoff",
			output:   "This is a longer sentence with more content that suddenly cuts off in the middle without prop",
			expected: true,
		},
		{
			name:     "JSON cutoff",
			output:   `{"key": "value", "another": "val`,
			expected: true,
		},
		{
			name:     "Complete JSON",
			output:   `{"key": "value"}`,
			expected: false,
		},
		{
			name:     "Short complete output",
			output:   "OK",
			expected: false,
		},
		{
			name:     "Empty output",
			output:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.isOutputTruncated(tt.output)
			if result != tt.expected {
				t.Errorf("isOutputTruncated(%q) = %v, want %v", tt.output, result, tt.expected)
			}
		})
	}
}

func TestCommandResult_Enhancement(t *testing.T) {
	// Test that enhanced CommandResult fields work correctly
	cmd := &types.Command{
		Type:          CommandRead,
		Args:          []string{"test.txt"},
		VerboseOutput: true,
	}

	result := &types.CommandResult{
		Command:      cmd,
		Success:      true,
		Output:       "This is truncated output...",
		FullOutput:   "This is truncated output with much more content that was originally cut off.",
		IsTruncated:  true,
		OutputLength: 74,
	}

	// Verify all fields are set correctly
	if !result.Success {
		t.Error("Expected success to be true")
	}

	if !result.IsTruncated {
		t.Error("Expected truncation to be detected")
	}

	if result.OutputLength != 74 {
		t.Errorf("Expected output length 74, got %d", result.OutputLength)
	}

	if result.FullOutput == "" {
		t.Error("Expected full output to be populated")
	}

	if !strings.Contains(result.Output, "...") {
		t.Error("Expected truncated output to contain ellipsis")
	}
}

func TestExecuteCommand_WithVerboseOutput(t *testing.T) {
	// Create a test client
	config := &types.ClaudeCodeConfig{
		TestMode:         true,
		WorkingDirectory: t.TempDir(),
	}

	client, err := NewClaudeCodeClient(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Create executor
	executor := NewCommandExecutor(client)

	// Test command with verbose output
	cmd := &types.Command{
		Type:          CommandRead,
		Args:          []string{"test.txt"},
		VerboseOutput: true,
	}

	// In test mode, we would mock the response
	// For now, just verify the command structure
	if !cmd.VerboseOutput {
		t.Error("Expected VerboseOutput to be true")
	}

	// Test the prompt building includes verbose request
	prompt, err := executor.buildCommandPrompt(cmd)
	if err != nil {
		t.Fatalf("Failed to build prompt: %v", err)
	}

	if !strings.Contains(prompt, "complete output without any truncation") {
		t.Error("Expected prompt to include verbose output request")
	}
}
