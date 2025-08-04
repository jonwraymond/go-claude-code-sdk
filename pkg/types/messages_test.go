package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClaudeCodeOptions(t *testing.T) {
	options := NewClaudeCodeOptions()

	assert.NotNil(t, options)
	assert.Equal(t, []string{}, options.AllowedTools)
	assert.Equal(t, 8000, options.MaxThinkingTokens)
	assert.Equal(t, []string{}, options.MCPTools)
	assert.Equal(t, map[string]McpServerConfig{}, options.MCPServers)
	assert.False(t, options.ContinueConversation)
	assert.Equal(t, []string{}, options.DisallowedTools)
	assert.Equal(t, []string{}, options.AddDirs)
}

func TestClaudeCodeOptionsSetCWD(t *testing.T) {
	options := NewClaudeCodeOptions()

	// Test with string
	options.SetCWD("/home/user")
	assert.Equal(t, "/home/user", *options.CWD)

	// Test with string pointer
	cwd := "/home/test"
	options.SetCWD(&cwd)
	assert.Equal(t, "/home/test", *options.CWD)
}

func TestPermissionMode(t *testing.T) {
	assert.Equal(t, PermissionMode("default"), PermissionModeDefault)
	assert.Equal(t, PermissionMode("acceptEdits"), PermissionModeAcceptEdits)
	assert.Equal(t, PermissionMode("bypassPermissions"), PermissionModeBypassPermissions)
}

func TestHelperFunctions(t *testing.T) {
	// Test StringPtr
	s := "test"
	ptr := StringPtr(s)
	assert.Equal(t, &s, ptr)

	// Test IntPtr
	i := 42
	iPtr := IntPtr(i)
	assert.Equal(t, &i, iPtr)

	// Test PermissionModePtr
	pm := PermissionModeAcceptEdits
	pmPtr := PermissionModePtr(pm)
	assert.Equal(t, &pm, pmPtr)
}

func TestContentBlockTypes(t *testing.T) {
	// Test TextBlock
	textBlock := TextBlock{Text: "Hello"}
	assert.Equal(t, "Hello", textBlock.Text)

	// Test ToolUseBlock
	input := map[string]interface{}{"command": "ls -la"}
	toolUseBlock := ToolUseBlock{
		ID:    "tool123",
		Name:  "Bash",
		Input: input,
	}
	assert.Equal(t, "tool123", toolUseBlock.ID)
	assert.Equal(t, "Bash", toolUseBlock.Name)
	assert.Equal(t, input, toolUseBlock.Input)

	// Test ToolResultBlock
	isError := false
	toolResultBlock := ToolResultBlock{
		ToolUseID: "tool123",
		Content:   "output",
		IsError:   &isError,
	}
	assert.Equal(t, "tool123", toolResultBlock.ToolUseID)
	assert.Equal(t, "output", toolResultBlock.Content)
	assert.Equal(t, &isError, toolResultBlock.IsError)
}

func TestMessageTypes(t *testing.T) {
	// Test UserMessage
	userMsg := &UserMessage{Content: "Hello"}
	assert.Equal(t, "Hello", userMsg.Content)

	// Test AssistantMessage
	content := []ContentBlock{TextBlock{Text: "Hi there"}}
	assistantMsg := &AssistantMessage{Content: content}
	assert.Equal(t, content, assistantMsg.Content)

	// Test SystemMessage
	data := map[string]interface{}{"key": "value"}
	systemMsg := &SystemMessage{
		Subtype: "info",
		Data:    data,
	}
	assert.Equal(t, "info", systemMsg.Subtype)
	assert.Equal(t, data, systemMsg.Data)

	// Test ResultMessage
	cost := 0.001
	usage := map[string]interface{}{"tokens": 100}
	resultMsg := &ResultMessage{
		Subtype:       "completion",
		DurationMs:    1000,
		DurationAPIMs: 800,
		IsError:       false,
		NumTurns:      1,
		SessionID:     "session123",
		TotalCostUSD:  &cost,
		Usage:         usage,
	}
	assert.Equal(t, "completion", resultMsg.Subtype)
	assert.Equal(t, 1000, resultMsg.DurationMs)
	assert.False(t, resultMsg.IsError)
	assert.Equal(t, &cost, resultMsg.TotalCostUSD)
	assert.Equal(t, usage, resultMsg.Usage)
}

func TestContentBlockMarshaling(t *testing.T) {
	// Test TextBlock marshaling - simple JSON without custom marshaling
	textBlock := TextBlock{Text: "Hello world"}
	data, err := json.Marshal(textBlock)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "\"Hello world\"")

	// Test ToolUseBlock marshaling
	input := map[string]interface{}{"command": "echo test"}
	toolUseBlock := ToolUseBlock{
		ID:    "tool1",
		Name:  "Bash",
		Input: input,
	}
	data, err = json.Marshal(toolUseBlock)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "\"tool1\"")
	assert.Contains(t, string(data), "\"Bash\"")

	// Test ToolResultBlock marshaling
	isError := false
	toolResultBlock := ToolResultBlock{
		ToolUseID: "tool1",
		Content:   "test output",
		IsError:   &isError,
	}
	data, err = json.Marshal(toolResultBlock)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "\"tool1\"")
}

func TestUnmarshalContentBlock(t *testing.T) {
	// Test unmarshaling TextBlock
	textData := []byte(`{"text":"Hello world"}`)
	block, err := UnmarshalContentBlock(textData)
	assert.NoError(t, err)
	textBlock, ok := block.(TextBlock)
	assert.True(t, ok)
	assert.Equal(t, "Hello world", textBlock.Text)

	// Test unmarshaling ToolUseBlock
	toolData := []byte(`{"id":"tool1","name":"Bash","input":{"command":"echo test"}}`)
	block, err = UnmarshalContentBlock(toolData)
	assert.NoError(t, err)
	toolUseBlock, ok := block.(ToolUseBlock)
	assert.True(t, ok)
	assert.Equal(t, "tool1", toolUseBlock.ID)
	assert.Equal(t, "Bash", toolUseBlock.Name)

	// Test unmarshaling ToolResultBlock
	resultData := []byte(`{"tool_use_id":"tool1","content":"output","is_error":false}`)
	block, err = UnmarshalContentBlock(resultData)
	assert.NoError(t, err)
	toolResultBlock, ok := block.(ToolResultBlock)
	assert.True(t, ok)
	assert.Equal(t, "tool1", toolResultBlock.ToolUseID)
}

func TestMessageInterface(t *testing.T) {
	// Test that all message types implement the Message interface
	var msg Message

	msg = &UserMessage{Content: "test"}
	assert.NotNil(t, msg)

	msg = &AssistantMessage{Content: []ContentBlock{}}
	assert.NotNil(t, msg)

	msg = &SystemMessage{Subtype: "test", Data: map[string]interface{}{}}
	assert.NotNil(t, msg)

	msg = &ResultMessage{Subtype: "test", SessionID: "test"}
	assert.NotNil(t, msg)
}

func TestContentBlockInterface(t *testing.T) {
	// Test that all content block types implement the ContentBlock interface
	var block ContentBlock

	block = TextBlock{Text: "test"}
	assert.NotNil(t, block)

	block = ToolUseBlock{ID: "test", Name: "test", Input: map[string]interface{}{}}
	assert.NotNil(t, block)

	block = ToolResultBlock{ToolUseID: "test"}
	assert.NotNil(t, block)
}
