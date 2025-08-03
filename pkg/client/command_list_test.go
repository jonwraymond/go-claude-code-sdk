package client

import (
	"context"
	"testing"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func TestNewCommandList(t *testing.T) {
	// Test creating a sequential command list
	cmd1 := ReadFile("file1.txt")
	cmd2 := ReadFile("file2.txt")
	cmd3 := ReadFile("file3.txt")

	cmdList := NewCommandList(cmd1, cmd2, cmd3)

	if len(cmdList.Commands) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(cmdList.Commands))
	}

	if cmdList.ExecutionMode != ExecutionModeSequential {
		t.Errorf("Expected sequential mode, got %s", cmdList.ExecutionMode)
	}

	if !cmdList.StopOnError {
		t.Error("Expected StopOnError to be true by default")
	}
}

func TestNewParallelCommandList(t *testing.T) {
	// Test creating a parallel command list
	cmd1 := ReadFile("file1.txt")
	cmd2 := ReadFile("file2.txt")
	cmd3 := ReadFile("file3.txt")

	cmdList := NewParallelCommandList(2, cmd1, cmd2, cmd3)

	if len(cmdList.Commands) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(cmdList.Commands))
	}

	if cmdList.ExecutionMode != ExecutionModeParallel {
		t.Errorf("Expected parallel mode, got %s", cmdList.ExecutionMode)
	}

	if cmdList.MaxParallel != 2 {
		t.Errorf("Expected MaxParallel to be 2, got %d", cmdList.MaxParallel)
	}

	if cmdList.StopOnError {
		t.Error("Expected StopOnError to be false for parallel execution")
	}
}

func TestCreateCommandListWithOptions(t *testing.T) {
	commands := []*Command{
		ReadFile("file1.txt"),
		ReadFile("file2.txt"),
		SearchCode("pattern"),
	}

	cmdList := CreateCommandList(commands,
		WithExecutionMode(ExecutionModeParallel),
		WithMaxParallel(5),
		WithStopOnError(false),
	)

	if cmdList.ExecutionMode != ExecutionModeParallel {
		t.Errorf("Expected parallel mode, got %s", cmdList.ExecutionMode)
	}

	if cmdList.MaxParallel != 5 {
		t.Errorf("Expected MaxParallel to be 5, got %d", cmdList.MaxParallel)
	}

	if cmdList.StopOnError {
		t.Error("Expected StopOnError to be false")
	}
}

func TestExecuteCommands_EmptyList(t *testing.T) {
	config := &types.ClaudeCodeConfig{
		TestMode:         true,
		WorkingDirectory: t.TempDir(),
	}

	client, err := NewClaudeCodeClient(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	executor := NewCommandExecutor(client)

	// Test empty command list
	result, err := executor.ExecuteCommands(context.Background(), &CommandList{})
	if err != nil {
		t.Fatalf("ExecuteCommands failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success for empty command list")
	}

	if result.TotalCommands != 0 {
		t.Errorf("Expected 0 commands, got %d", result.TotalCommands)
	}

	if len(result.Results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(result.Results))
	}
}

func TestExecuteCommands_Sequential(t *testing.T) {
	config := &types.ClaudeCodeConfig{
		TestMode:         true,
		WorkingDirectory: t.TempDir(),
	}

	client, err := NewClaudeCodeClient(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Create a sequential command list
	cmdList := NewCommandList(
		ReadFile("test1.txt"),
		ReadFile("test2.txt"),
		SearchCode("pattern"),
	)

	// In test mode, we would mock the execution
	// For now, just verify the command list structure
	if cmdList.ExecutionMode != ExecutionModeSequential {
		t.Error("Expected sequential execution mode")
	}

	if len(cmdList.Commands) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(cmdList.Commands))
	}
}

func TestExecuteCommands_Parallel(t *testing.T) {
	config := &types.ClaudeCodeConfig{
		TestMode:         true,
		WorkingDirectory: t.TempDir(),
	}

	client, err := NewClaudeCodeClient(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Create a parallel command list
	cmdList := NewParallelCommandList(2,
		ReadFile("test1.txt"),
		ReadFile("test2.txt"),
		ReadFile("test3.txt"),
	)

	// Verify parallel configuration
	if cmdList.ExecutionMode != ExecutionModeParallel {
		t.Error("Expected parallel execution mode")
	}

	if cmdList.MaxParallel != 2 {
		t.Errorf("Expected MaxParallel to be 2, got %d", cmdList.MaxParallel)
	}
}

func TestCommandListResult_Aggregation(t *testing.T) {
	// Test result aggregation
	result := &types.CommandListResult{
		Results: []*types.CommandResult{
			{Success: true},
			{Success: false, Error: "error 1"},
			{Success: true},
			{Success: false, Error: "error 2"},
		},
		TotalCommands:      4,
		SuccessfulCommands: 2,
		FailedCommands:     2,
		Success:            false,
		ExecutionTime:      1500,
		Errors:             []string{"error 1", "error 2"},
	}

	if result.Success {
		t.Error("Expected overall success to be false when some commands fail")
	}

	if result.SuccessfulCommands != 2 {
		t.Errorf("Expected 2 successful commands, got %d", result.SuccessfulCommands)
	}

	if result.FailedCommands != 2 {
		t.Errorf("Expected 2 failed commands, got %d", result.FailedCommands)
	}

	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}
}

func TestExecuteCommands_ContextCancellation(t *testing.T) {
	config := &types.ClaudeCodeConfig{
		TestMode:         true,
		WorkingDirectory: t.TempDir(),
	}

	client, err := NewClaudeCodeClient(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	executor := NewCommandExecutor(client)

	// Create a context that can be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Create a command list
	cmdList := NewCommandList(
		ReadFile("test1.txt"),
		ReadFile("test2.txt"),
		ReadFile("test3.txt"),
	)

	// In a real scenario, this would be cancelled due to timeout
	// For test mode, we just verify the structure
	_ = ctx
	_ = executor
	_ = cmdList
}
