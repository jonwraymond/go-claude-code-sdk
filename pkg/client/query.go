package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// PermissionMode defines how Claude Code handles file edit permissions
type PermissionMode string

const (
	// PermissionModeAsk prompts the user for permission before edits
	PermissionModeAsk PermissionMode = "ask"

	// PermissionModeAcceptEdits automatically accepts file edits
	PermissionModeAcceptEdits PermissionMode = "acceptEdits"

	// PermissionModeRejectEdits automatically rejects file edits
	PermissionModeRejectEdits PermissionMode = "rejectEdits"
)

// QueryOptions configures the behavior of a query execution
type QueryOptions struct {
	// SystemPrompt sets the system prompt for Claude Code
	SystemPrompt string

	// MaxTurns limits the number of conversation turns
	MaxTurns int

	// AllowedTools specifies which tools Claude can use
	AllowedTools []string

	// PermissionMode controls how file edits are handled
	PermissionMode PermissionMode

	// CWD sets the current working directory
	CWD string

	// Model specifies which Claude model to use
	Model string

	// SessionID allows resuming a previous session
	SessionID string

	// Stream enables streaming responses
	Stream bool

	// Timeout sets the query timeout
	Timeout int

	// Environment variables to pass to Claude Code
	Env map[string]string
}

// QueryResult represents the result of a query execution
type QueryResult struct {
	Messages []types.Message
	Error    error
	Metadata map[string]any
}

// QueryMessages executes a query against Claude Code and returns a channel of messages
// This matches the pattern used by the official Python and TypeScript SDKs
func (c *ClaudeCodeClient) QueryMessages(ctx context.Context, prompt string, options *QueryOptions) (<-chan *types.Message, error) {
	messageChan := make(chan *types.Message, 100)

	// Set defaults
	if options == nil {
		options = &QueryOptions{
			MaxTurns:       10,
			Stream:         true,
			PermissionMode: PermissionModeAsk,
		}
	}

	// Create session using session manager
	session, err := c.sessionManager.CreateSession(ctx, options.SessionID)
	if err != nil {
		close(messageChan)
		return messageChan, fmt.Errorf("failed to create session: %w", err)
	}

	// Configure session
	if options.Model != "" {
		session.model = options.Model
	}
	if options.CWD != "" {
		session.projectDir = options.CWD
	}

	// Start processing in goroutine
	go func() {
		defer close(messageChan)
		defer func() {
			if !session.closed {
				session.Close()
			}
		}()

		// Send initial message
		userMsg := &types.Message{
			Role:    types.RoleUser,
			Content: prompt,
		}
		messageChan <- userMsg

		// Build command for chat
		cmd := &types.Command{
			Type:    types.CommandType("chat"), // Using chat as command type
			Args:    []string{prompt},
			Options: c.convertQueryOptionsToCommandOptions(options),
		}

		// Execute with streaming
		c.executeQueryWithStreaming(ctx, session, cmd, messageChan, options)
	}()

	return messageChan, nil
}

// QueryMessagesSync executes a query synchronously and returns all messages
func (c *ClaudeCodeClient) QueryMessagesSync(ctx context.Context, prompt string, options *QueryOptions) (*QueryResult, error) {
	messages := make([]types.Message, 0)
	messageChan, err := c.QueryMessages(ctx, prompt, options)
	if err != nil {
		return nil, err
	}

	for msg := range messageChan {
		if msg != nil {
			messages = append(messages, *msg)
		}
	}

	// Check if last message indicates an error
	var queryErr error
	if len(messages) > 0 {
		lastMsg := messages[len(messages)-1]
		if lastMsg.Role == types.RoleSystem && strings.Contains(lastMsg.Content, "Error:") {
			queryErr = fmt.Errorf("query failed: %s", lastMsg.Content)
		}
	}

	return &QueryResult{
		Messages: messages,
		Error:    queryErr,
		Metadata: map[string]any{
			"turn_count": len(messages) / 2, // Approximate turn count
		},
	}, nil
}

// executeQueryWithStreaming handles the streaming execution of a query
func (c *ClaudeCodeClient) executeQueryWithStreaming(
	ctx context.Context,
	session *ClaudeCodeSession,
	cmd *types.Command,
	messageChan chan<- *types.Message,
	options *QueryOptions,
) {
	// Build command
	cmdArgs := c.buildQueryCommand(session, cmd, options)

	// Create and start claude process
	process := exec.CommandContext(ctx, c.claudeCodeCmd, cmdArgs...)
	process.Dir = c.workingDir

	// Create pipes for stdout
	stdout, err := process.StdoutPipe()
	if err != nil {
		messageChan <- &types.Message{
			Role:    types.RoleSystem,
			Content: fmt.Sprintf("Error creating stdout pipe: %v", err),
		}
		return
	}

	// Start the process
	if err := process.Start(); err != nil {
		messageChan <- &types.Message{
			Role:    types.RoleSystem,
			Content: fmt.Sprintf("Error starting Claude Code: %v", err),
		}
		return
	}

	// Track the process
	processID := fmt.Sprintf("query_%s", session.ID)
	c.processMu.Lock()
	c.activeProcesses[processID] = process
	c.processMu.Unlock()

	defer func() {
		c.processMu.Lock()
		delete(c.activeProcesses, processID)
		c.processMu.Unlock()
		process.Process.Kill()
	}()

	// Parse streaming output
	c.parseStreamingOutput(stdout, messageChan, options)
}

// parseStreamingOutput parses the streaming output from Claude Code
func (c *ClaudeCodeClient) parseStreamingOutput(
	stdout any,
	messageChan chan<- *types.Message,
	options *QueryOptions,
) {
	scanner := bufio.NewScanner(stdout.(interface{ Read([]byte) (int, error) }))

	var currentMessage *types.Message
	var contentBuffer strings.Builder
	inAssistantMessage := false
	turnCount := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Parse different output patterns
		if strings.HasPrefix(line, "Claude:") || strings.HasPrefix(line, "Assistant:") {
			// Start of assistant message
			if currentMessage != nil && contentBuffer.Len() > 0 {
				currentMessage.Content = strings.TrimSpace(contentBuffer.String())
				messageChan <- currentMessage
				contentBuffer.Reset()
			}

			currentMessage = &types.Message{
				Role: types.RoleAssistant,
			}
			inAssistantMessage = true

			// Extract content after prefix
			content := strings.TrimPrefix(line, "Claude:")
			content = strings.TrimPrefix(content, "Assistant:")
			contentBuffer.WriteString(strings.TrimSpace(content))

		} else if strings.HasPrefix(line, "Tool:") {
			// Tool usage detected
			if currentMessage != nil && contentBuffer.Len() > 0 {
				currentMessage.Content = strings.TrimSpace(contentBuffer.String())
				messageChan <- currentMessage
				contentBuffer.Reset()
			}

			// Parse tool information
			toolInfo := c.parseToolUsage(line)
			if toolInfo != nil {
				currentMessage = &types.Message{
					Role: types.RoleAssistant,
					ToolCalls: []types.ToolCall{
						{
							ID:   toolInfo.ID,
							Type: "function",
							Function: types.FunctionCall{
								Name:      toolInfo.Name,
								Arguments: toolInfo.Arguments,
							},
						},
					},
				}
				messageChan <- currentMessage
			}

		} else if strings.HasPrefix(line, "Result:") {
			// Tool result
			result := strings.TrimPrefix(line, "Result:")
			resultMsg := &types.Message{
				Role:    types.RoleTool,
				Content: strings.TrimSpace(result),
			}
			messageChan <- resultMsg

		} else if inAssistantMessage && line != "" {
			// Continue building assistant message
			if contentBuffer.Len() > 0 {
				contentBuffer.WriteString("\n")
			}
			contentBuffer.WriteString(line)

		} else if strings.Contains(line, "Turn limit reached") ||
			strings.Contains(line, "Max turns reached") {
			// Turn limit reached
			if currentMessage != nil && contentBuffer.Len() > 0 {
				currentMessage.Content = strings.TrimSpace(contentBuffer.String())
				messageChan <- currentMessage
			}

			// Send system message about turn limit
			messageChan <- &types.Message{
				Role:    types.RoleSystem,
				Content: fmt.Sprintf("Turn limit reached (%d turns)", options.MaxTurns),
			}
			break
		}

		// Check turn count
		if strings.HasPrefix(line, "User:") || strings.HasPrefix(line, "Human:") {
			turnCount++
			if options.MaxTurns > 0 && turnCount >= options.MaxTurns {
				break
			}
		}
	}

	// Send final message if exists
	if currentMessage != nil && contentBuffer.Len() > 0 {
		currentMessage.Content = strings.TrimSpace(contentBuffer.String())
		messageChan <- currentMessage
	}
}

// parseToolUsage extracts tool information from a tool usage line
func (c *ClaudeCodeClient) parseToolUsage(line string) *struct {
	ID        string
	Name      string
	Arguments string
} {
	// Try to parse JSON-formatted tool usage
	if idx := strings.Index(line, "{"); idx >= 0 {
		jsonStr := line[idx:]
		var toolData map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &toolData); err == nil {
			return &struct {
				ID        string
				Name      string
				Arguments string
			}{
				ID:        fmt.Sprintf("%v", toolData["id"]),
				Name:      fmt.Sprintf("%v", toolData["name"]),
				Arguments: fmt.Sprintf("%v", toolData["input"]),
			}
		}
	}

	// Fallback to simple parsing
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		return &struct {
			ID        string
			Name      string
			Arguments string
		}{
			ID:        fmt.Sprintf("tool_%d", len(parts)),
			Name:      parts[1],
			Arguments: "{}",
		}
	}

	return nil
}

// buildQueryCommand builds the command arguments for a query
func (c *ClaudeCodeClient) buildQueryCommand(
	session *ClaudeCodeSession,
	cmd *types.Command,
	options *QueryOptions,
) []string {
	args := []string{c.claudeCodeCmd}

	// Add session ID
	if session.ID != "" {
		args = append(args, "--session", session.ID)
	}

	// Add model
	if options.Model != "" {
		args = append(args, "--model", options.Model)
	}

	// Add system prompt
	if options.SystemPrompt != "" {
		args = append(args, "--system-prompt", options.SystemPrompt)
	}

	// Add max turns
	if options.MaxTurns > 0 {
		args = append(args, "--max-turns", fmt.Sprintf("%d", options.MaxTurns))
	}

	// Add permission mode
	switch options.PermissionMode {
	case PermissionModeAcceptEdits:
		args = append(args, "--accept-edits")
	case PermissionModeRejectEdits:
		args = append(args, "--reject-edits")
	}

	// Add allowed tools
	if len(options.AllowedTools) > 0 {
		args = append(args, "--tools", strings.Join(options.AllowedTools, ","))
	}

	// Add timeout
	if options.Timeout > 0 {
		args = append(args, "--timeout", fmt.Sprintf("%d", options.Timeout))
	}

	// Add the prompt
	args = append(args, cmd.Args[0])

	return args
}

// convertQueryOptionsToCommandOptions converts QueryOptions to command options map
func (c *ClaudeCodeClient) convertQueryOptionsToCommandOptions(options *QueryOptions) map[string]any {
	opts := make(map[string]any)

	if options.SystemPrompt != "" {
		opts["system_prompt"] = options.SystemPrompt
	}
	if options.MaxTurns > 0 {
		opts["max_turns"] = options.MaxTurns
	}
	if len(options.AllowedTools) > 0 {
		opts["allowed_tools"] = options.AllowedTools
	}
	if options.PermissionMode != "" {
		opts["permission_mode"] = string(options.PermissionMode)
	}
	if options.Timeout > 0 {
		opts["timeout"] = options.Timeout
	}

	return opts
}
