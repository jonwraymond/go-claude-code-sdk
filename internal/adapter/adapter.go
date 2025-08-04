// Package adapter provides type conversion utilities between public and internal types.
package adapter

import (
	"github.com/jonwraymond/go-claude-code-sdk/internal/transport"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/errors"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// ConvertFromInternalOptions converts internal options to public options.
func ConvertFromInternalOptions(internalOpts *transport.ClaudeCodeOptions) *types.ClaudeCodeOptions {
	if internalOpts == nil {
		return nil
	}

	// Convert MCP servers
	mcpServers := make(map[string]types.McpServerConfig)
	for k, v := range internalOpts.MCPServers {
		mcpServers[k] = types.McpServerConfig(v)
	}

	opts := &types.ClaudeCodeOptions{
		AllowedTools:             internalOpts.AllowedTools,
		MaxThinkingTokens:        internalOpts.MaxThinkingTokens,
		SystemPrompt:             internalOpts.SystemPrompt,
		AppendSystemPrompt:       internalOpts.AppendSystemPrompt,
		MCPTools:                 internalOpts.MCPTools,
		MCPServers:               mcpServers,
		ContinueConversation:     internalOpts.ContinueConversation,
		Resume:                   internalOpts.Resume,
		MaxTurns:                 internalOpts.MaxTurns,
		DisallowedTools:          internalOpts.DisallowedTools,
		Model:                    internalOpts.Model,
		PermissionPromptToolName: internalOpts.PermissionPromptToolName,
		CWD:                      internalOpts.CWD,
		Settings:                 internalOpts.Settings,
		AddDirs:                  internalOpts.AddDirs,
	}

	if internalOpts.PermissionMode != nil {
		pm := types.PermissionMode(*internalOpts.PermissionMode)
		opts.PermissionMode = &pm
	}

	return opts
}

// ConvertToInternalOptions converts public options to internal options.
func ConvertToInternalOptions(opts *types.ClaudeCodeOptions) *transport.ClaudeCodeOptions {
	if opts == nil {
		return &transport.ClaudeCodeOptions{
			AllowedTools:         []string{},
			MaxThinkingTokens:    8000,
			MCPTools:             []string{},
			MCPServers:           make(map[string]transport.McpServerConfig),
			ContinueConversation: false,
			DisallowedTools:      []string{},
			AddDirs:              []string{},
		}
	}

	// Convert MCP servers
	internalMcpServers := make(map[string]transport.McpServerConfig)
	for k, v := range opts.MCPServers {
		internalMcpServers[k] = transport.McpServerConfig(v)
	}

	internalOpts := &transport.ClaudeCodeOptions{
		AllowedTools:             opts.AllowedTools,
		MaxThinkingTokens:        opts.MaxThinkingTokens,
		SystemPrompt:             opts.SystemPrompt,
		AppendSystemPrompt:       opts.AppendSystemPrompt,
		MCPTools:                 opts.MCPTools,
		MCPServers:               internalMcpServers,
		ContinueConversation:     opts.ContinueConversation,
		Resume:                   opts.Resume,
		MaxTurns:                 opts.MaxTurns,
		DisallowedTools:          opts.DisallowedTools,
		Model:                    opts.Model,
		PermissionPromptToolName: opts.PermissionPromptToolName,
		CWD:                      opts.CWD,
		Settings:                 opts.Settings,
		AddDirs:                  opts.AddDirs,
	}

	if opts.PermissionMode != nil {
		pm := transport.PermissionMode(*opts.PermissionMode)
		internalOpts.PermissionMode = &pm
	}

	return internalOpts
}

// ConvertFromInternalError converts internal error to public error.
func ConvertFromInternalError(err error) error {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case *transport.CLIConnectionError:
		return &errors.CLIConnectionError{
			ClaudeSDKError: &errors.ClaudeSDKError{
				Message: e.Message,
				Cause:   e.Cause,
			},
		}
	case *transport.CLINotFoundError:
		return &errors.CLINotFoundError{
			ClaudeSDKError: &errors.ClaudeSDKError{
				Message: e.Message,
				Cause:   e.Cause,
			},
		}
	case *transport.ProcessError:
		return &errors.ProcessError{
			ClaudeSDKError: &errors.ClaudeSDKError{
				Message: e.Message,
				Cause:   e.Cause,
			},
			ExitCode: e.ExitCode,
		}
	case *transport.CLIJSONDecodeError:
		return &errors.CLIJSONDecodeError{
			ClaudeSDKError: &errors.ClaudeSDKError{
				Message: e.Message,
				Cause:   e.Cause,
			},
			RawData: e.RawData,
		}
	case *transport.ClaudeSDKError:
		return &errors.ClaudeSDKError{
			Message: e.Message,
			Cause:   e.Cause,
		}
	default:
		return err
	}
}

// ParseMessageFromRaw converts a raw message map to a proper Message interface.
func ParseMessageFromRaw(raw map[string]interface{}) (types.Message, error) {
	msgType := transport.DetermineMessageType(raw)

	switch msgType {
	case "user":
		return parseUserMessageFromRaw(raw)
	case "assistant":
		return parseAssistantMessageFromRaw(raw)
	case "system":
		return parseSystemMessageFromRaw(raw)
	case "result":
		return parseResultMessageFromRaw(raw)
	default:
		// Default to assistant message
		return parseAssistantMessageFromRaw(raw)
	}
}

func parseUserMessageFromRaw(raw map[string]interface{}) (*types.UserMessage, error) {
	return &types.UserMessage{
		Content: raw["content"],
	}, nil
}

func parseAssistantMessageFromRaw(raw map[string]interface{}) (*types.AssistantMessage, error) {
	var content []types.ContentBlock

	if contentRaw, ok := raw["content"].([]interface{}); ok {
		for _, blockRaw := range contentRaw {
			if blockMap, ok := blockRaw.(map[string]interface{}); ok {
				block, err := parseContentBlockFromRaw(blockMap)
				if err != nil {
					return nil, err
				}
				content = append(content, block)
			}
		}
	}

	return &types.AssistantMessage{
		Content: content,
	}, nil
}

func parseSystemMessageFromRaw(raw map[string]interface{}) (*types.SystemMessage, error) {
	subtype, _ := raw["subtype"].(string)
	data, _ := raw["data"].(map[string]interface{})

	return &types.SystemMessage{
		Subtype: subtype,
		Data:    data,
	}, nil
}

func parseResultMessageFromRaw(raw map[string]interface{}) (*types.ResultMessage, error) {
	msg := &types.ResultMessage{}

	if subtype, ok := raw["subtype"].(string); ok {
		msg.Subtype = subtype
	}

	if durationMs, ok := raw["duration_ms"].(float64); ok {
		msg.DurationMs = int(durationMs)
	}

	if durationAPIMs, ok := raw["duration_api_ms"].(float64); ok {
		msg.DurationAPIMs = int(durationAPIMs)
	}

	if isError, ok := raw["is_error"].(bool); ok {
		msg.IsError = isError
	}

	if numTurns, ok := raw["num_turns"].(float64); ok {
		msg.NumTurns = int(numTurns)
	}

	if sessionID, ok := raw["session_id"].(string); ok {
		msg.SessionID = sessionID
	}

	if totalCostUSD, ok := raw["total_cost_usd"].(float64); ok {
		msg.TotalCostUSD = &totalCostUSD
	}

	if usage, ok := raw["usage"].(map[string]interface{}); ok {
		msg.Usage = usage
	}

	if result, ok := raw["result"].(string); ok {
		msg.Result = &result
	}

	return msg, nil
}

func parseContentBlockFromRaw(raw map[string]interface{}) (types.ContentBlock, error) {
	blockType := transport.DetermineContentBlockType(raw)

	switch blockType {
	case "text":
		if text, ok := raw["text"].(string); ok {
			return types.TextBlock{Text: text}, nil
		}
		return types.TextBlock{}, nil

	case "tool_use":
		block := types.ToolUseBlock{}
		if id, ok := raw["id"].(string); ok {
			block.ID = id
		}
		if name, ok := raw["name"].(string); ok {
			block.Name = name
		}
		if input, ok := raw["input"].(map[string]interface{}); ok {
			block.Input = input
		}
		return block, nil

	case "tool_result":
		block := types.ToolResultBlock{}
		if toolUseID, ok := raw["tool_use_id"].(string); ok {
			block.ToolUseID = toolUseID
		}
		if content := raw["content"]; content != nil {
			block.Content = content
		}
		if isError, ok := raw["is_error"].(bool); ok {
			block.IsError = &isError
		}
		return block, nil

	default:
		// Default to text block
		return types.TextBlock{}, nil
	}
}
