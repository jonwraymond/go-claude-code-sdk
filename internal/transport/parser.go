package transport

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
	// Check for explicit type field
	if msgType, ok := raw["type"].(string); ok {
		return msgType
	}

	// Check for role field (common in message structures)
	if role, ok := raw["role"].(string); ok {
		switch role {
		case "user":
			return "user"
		case "assistant":
			return "assistant"
		case "system":
			return "system"
		}
	}

	// Check for result message indicators
	if _, hasSubtype := raw["subtype"]; hasSubtype {
		if _, hasDuration := raw["duration_ms"]; hasDuration {
			return "result"
		}
	}

	// Default to assistant message
	return "assistant"
}

// DetermineContentBlockType determines the type of content block from raw data.
func DetermineContentBlockType(raw map[string]interface{}) string {
	// Check for explicit type field
	if blockType, ok := raw["type"].(string); ok {
		return blockType
	}

	// Check for text field
	if _, hasText := raw["text"]; hasText {
		return "text"
	}

	// Check for tool use fields
	if _, hasName := raw["name"]; hasName {
		if _, hasInput := raw["input"]; hasInput {
			return "tool_use"
		}
	}

	// Check for tool result fields
	if _, hasToolUseID := raw["tool_use_id"]; hasToolUseID {
		return "tool_result"
	}

	// Default to text
	return "text"
}
