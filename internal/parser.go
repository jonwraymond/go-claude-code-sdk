package internal

import (
	"encoding/json"
)

// ParseMessage parses raw message data from Claude CLI into a generic message map.
// The calling code should then convert this to the appropriate message type.
func ParseMessage(data []byte) (map[string]interface{}, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, NewCLIJSONDecodeError("Failed to parse message JSON", string(data), err)
	}
	return raw, nil
}

// ParseContentBlock parses a content block from raw JSON.
func ParseContentBlock(data json.RawMessage) (map[string]interface{}, error) {
	var block map[string]interface{}
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, NewCLIJSONDecodeError("Failed to parse content block", string(data), err)
	}
	return block, nil
}

// DetermineMessageType determines the type of message from raw data.
func DetermineMessageType(raw map[string]interface{}) string {
	// Check for user/assistant role
	if role, ok := raw["role"].(string); ok {
		return role
	}

	// Check for system message
	if _, ok := raw["subtype"].(string); ok {
		if _, hasData := raw["data"]; hasData {
			return "system"
		}
		// Could also be a result message
		if _, hasDuration := raw["duration_ms"]; hasDuration {
			return "result"
		}
	}

	// Check for result message by other fields
	if _, hasDuration := raw["duration_ms"]; hasDuration {
		return "result"
	}

	// Default to assistant
	return "assistant"
}

// DetermineContentBlockType determines the type of content block.
func DetermineContentBlockType(block map[string]interface{}) string {
	// Check for text field
	if _, hasText := block["text"]; hasText {
		return "text"
	}

	// Check for tool use (has name and input)
	if _, hasName := block["name"]; hasName {
		if _, hasInput := block["input"]; hasInput {
			return "tool_use"
		}
	}

	// Check for tool result (has tool_use_id)
	if _, hasToolUseID := block["tool_use_id"]; hasToolUseID {
		return "tool_result"
	}

	// Default to text
	return "text"
}