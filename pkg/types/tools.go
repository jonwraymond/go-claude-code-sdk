package types

import (
	"context"
	"encoding/json"
)

// Tool represents a function or capability that Claude can use during conversations.
// Tools allow Claude to perform actions like web searches, calculations, or API calls.
//
// Example usage:
//
//	tool := &types.Tool{
//		Name:        "search_web",
//		Description: "Search the web for information",
//		InputSchema: types.ToolInputSchema{
//			Type: "object",
//			Properties: map[string]types.ToolProperty{
//				"query": {
//					Type:        "string",
//					Description: "The search query",
//				},
//			},
//			Required: []string{"query"},
//		},
//	}
type Tool struct {
	// Name is the unique identifier for this tool
	Name string `json:"name"`

	// Description explains what this tool does and when to use it
	Description string `json:"description"`

	// InputSchema defines the expected input parameters for this tool
	InputSchema ToolInputSchema `json:"input_schema"`

	// Metadata contains additional tool information
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ToolInputSchema defines the JSON schema for tool input parameters.
type ToolInputSchema struct {
	// Type is the root type of the schema (usually "object")
	Type string `json:"type"`

	// Properties defines the properties of an object schema
	Properties map[string]ToolProperty `json:"properties,omitempty"`

	// Required lists the required property names
	Required []string `json:"required,omitempty"`

	// AdditionalProperties controls whether additional properties are allowed
	AdditionalProperties any `json:"additionalProperties,omitempty"`

	// Description provides additional context for the schema
	Description string `json:"description,omitempty"`
}

// ToolProperty defines a single property in a tool's input schema.
type ToolProperty struct {
	// Type is the data type of this property
	Type string `json:"type"`

	// Description explains what this property is for
	Description string `json:"description,omitempty"`

	// Enum lists allowed values for this property
	Enum []any `json:"enum,omitempty"`

	// Default provides a default value for this property
	Default any `json:"default,omitempty"`

	// Items defines the schema for array items (when type is "array")
	Items *ToolProperty `json:"items,omitempty"`

	// Properties defines nested object properties (when type is "object")
	Properties map[string]ToolProperty `json:"properties,omitempty"`

	// Required lists required properties for nested objects
	Required []string `json:"required,omitempty"`

	// Format provides additional format constraints (e.g., "email", "date")
	Format string `json:"format,omitempty"`

	// Minimum/Maximum define numeric constraints
	Minimum *float64 `json:"minimum,omitempty"`
	Maximum *float64 `json:"maximum,omitempty"`

	// MinLength/MaxLength define string length constraints
	MinLength *int `json:"minLength,omitempty"`
	MaxLength *int `json:"maxLength,omitempty"`

	// Pattern defines a regex pattern for string validation
	Pattern string `json:"pattern,omitempty"`
}

// ToolExecutor defines the interface for executing tools.
// Implementations handle the actual execution of tool functions.
type ToolExecutor interface {
	// Execute runs the tool with the provided input and returns the result.
	Execute(ctx context.Context, toolName string, input map[string]any) (*ToolResult, error)

	// GetTool returns information about a specific tool.
	GetTool(toolName string) (*Tool, error)

	// ListTools returns all available tools.
	ListTools() ([]Tool, error)

	// RegisterTool adds a new tool to the executor.
	RegisterTool(tool *Tool, handler ToolHandler) error

	// UnregisterTool removes a tool from the executor.
	UnregisterTool(toolName string) error
}

// ToolHandler is a function that implements the actual tool logic.
type ToolHandler func(ctx context.Context, input map[string]any) (*ToolResult, error)

// ToolUse represents a tool use request from Claude.
type ToolUse struct {
	// ID is the unique identifier for this tool use
	ID string `json:"id"`

	// Name is the name of the tool to use
	Name string `json:"name"`

	// Input contains the parameters to pass to the tool
	Input map[string]any `json:"input"`
}

// ToolResult represents the result of executing a tool.
type ToolResult struct {
	// ToolUseID is the ID of the tool use that generated this result
	ToolUseID string `json:"tool_use_id,omitempty"`

	// IsError indicates whether this result represents an error
	IsError bool `json:"is_error,omitempty"`

	// Content contains the result content blocks
	Content []ContentBlock `json:"content,omitempty"`

	// Success indicates whether the tool execution was successful
	Success bool `json:"success"`

	// Error contains error information if the tool failed
	Error string `json:"error,omitempty"`

	// Metadata contains additional result information
	Metadata map[string]any `json:"metadata,omitempty"`

	// ExecutionTime is how long the tool took to execute
	ExecutionTime int64 `json:"execution_time,omitempty"`

	// Usage contains resource usage information
	Usage *ToolUsage `json:"usage,omitempty"`
}

// ToolUsage contains information about resource usage during tool execution.
type ToolUsage struct {
	// TokensUsed is the number of tokens consumed (if applicable)
	TokensUsed int `json:"tokens_used,omitempty"`

	// APICallsUsed is the number of API calls made
	APICallsUsed int `json:"api_calls_used,omitempty"`

	// DataProcessed is the amount of data processed in bytes
	DataProcessed int64 `json:"data_processed,omitempty"`

	// Cost is the estimated cost of the operation
	Cost float64 `json:"cost,omitempty"`

	// Currency is the currency for the cost
	Currency string `json:"currency,omitempty"`
}

// ToolChoice represents the different ways Claude can choose to use tools.
type ToolChoice interface {
	toolChoice() // private method to make this a sealed interface
}

// AutoToolChoice indicates Claude should automatically decide whether to use tools.
type AutoToolChoice struct{}

func (AutoToolChoice) toolChoice() {}

// NoneToolChoice indicates Claude should not use any tools.
type NoneToolChoice struct{}

func (NoneToolChoice) toolChoice() {}

// AnyToolChoice indicates Claude must use at least one tool.
type AnyToolChoice struct{}

func (AnyToolChoice) toolChoice() {}

// SpecificToolChoice indicates Claude must use a specific tool.
type SpecificToolChoice struct {
	// Name is the name of the required tool
	Name string `json:"name"`
}

func (SpecificToolChoice) toolChoice() {}

// MarshalJSON implements custom JSON marshaling for ToolChoice.
func (tc AutoToolChoice) MarshalJSON() ([]byte, error) {
	return json.Marshal("auto")
}

func (tc NoneToolChoice) MarshalJSON() ([]byte, error) {
	return json.Marshal("none")
}

func (tc AnyToolChoice) MarshalJSON() ([]byte, error) {
	return json.Marshal("any")
}

func (tc SpecificToolChoice) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type": "tool",
		"name": tc.Name,
	})
}

// SimpleToolExecutor provides a basic implementation of ToolExecutor.
// It maintains a registry of tools and their handlers.
type SimpleToolExecutor struct {
	tools    map[string]*Tool
	handlers map[string]ToolHandler
}

// NewSimpleToolExecutor creates a new simple tool executor.
func NewSimpleToolExecutor() *SimpleToolExecutor {
	return &SimpleToolExecutor{
		tools:    make(map[string]*Tool),
		handlers: make(map[string]ToolHandler),
	}
}

// Execute runs the specified tool with the provided input.
func (e *SimpleToolExecutor) Execute(ctx context.Context, toolName string, input map[string]any) (*ToolResult, error) {
	handler, exists := e.handlers[toolName]
	if !exists {
		return &ToolResult{
			Success: false,
			Error:   "tool not found: " + toolName,
		}, nil
	}

	// Validate input against tool schema if desired
	// (Implementation would go here)

	result, err := handler(ctx, input)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return result, nil
}

// GetTool returns information about a specific tool.
func (e *SimpleToolExecutor) GetTool(toolName string) (*Tool, error) {
	tool, exists := e.tools[toolName]
	if !exists {
		return nil, &ToolError{
			Type:    "tool_not_found",
			Message: "tool not found: " + toolName,
		}
	}
	return tool, nil
}

// ListTools returns all available tools.
func (e *SimpleToolExecutor) ListTools() ([]Tool, error) {
	tools := make([]Tool, 0, len(e.tools))
	for _, tool := range e.tools {
		tools = append(tools, *tool)
	}
	return tools, nil
}

// RegisterTool adds a new tool to the executor.
func (e *SimpleToolExecutor) RegisterTool(tool *Tool, handler ToolHandler) error {
	if tool.Name == "" {
		return &ToolError{
			Type:    "invalid_tool",
			Message: "tool name is required",
		}
	}

	if handler == nil {
		return &ToolError{
			Type:    "invalid_handler",
			Message: "tool handler is required",
		}
	}

	e.tools[tool.Name] = tool
	e.handlers[tool.Name] = handler
	return nil
}

// UnregisterTool removes a tool from the executor.
func (e *SimpleToolExecutor) UnregisterTool(toolName string) error {
	delete(e.tools, toolName)
	delete(e.handlers, toolName)
	return nil
}

// ToolError represents an error that occurs during tool operations.
type ToolError struct {
	// Type is the error type identifier
	Type string `json:"type"`

	// Message is a human-readable error message
	Message string `json:"message"`

	// ToolName is the name of the tool that caused the error
	ToolName string `json:"tool_name,omitempty"`

	// Details contains additional error information
	Details map[string]any `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *ToolError) Error() string {
	return e.Message
}

// BuiltinTools contains definitions for commonly used tools.
var BuiltinTools = struct {
	WebSearch    *Tool
	Calculator   *Tool
	FileRead     *Tool
	FileWrite    *Tool
	CodeExecutor *Tool
}{
	WebSearch: &Tool{
		Name:        "web_search",
		Description: "Search the web for information on a given topic",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]ToolProperty{
				"query": {
					Type:        "string",
					Description: "The search query",
				},
				"limit": {
					Type:        "integer",
					Description: "Maximum number of results to return",
					Default:     10,
					Minimum:     func() *float64 { v := 1.0; return &v }(),
					Maximum:     func() *float64 { v := 50.0; return &v }(),
				},
			},
			Required: []string{"query"},
		},
	},

	Calculator: &Tool{
		Name:        "calculator",
		Description: "Perform mathematical calculations",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]ToolProperty{
				"expression": {
					Type:        "string",
					Description: "The mathematical expression to evaluate",
				},
			},
			Required: []string{"expression"},
		},
	},

	FileRead: &Tool{
		Name:        "file_read",
		Description: "Read the contents of a file",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]ToolProperty{
				"path": {
					Type:        "string",
					Description: "The path to the file to read",
				},
				"encoding": {
					Type:        "string",
					Description: "The text encoding to use",
					Default:     "utf-8",
					Enum:        []any{"utf-8", "ascii", "latin-1"},
				},
			},
			Required: []string{"path"},
		},
	},

	FileWrite: &Tool{
		Name:        "file_write",
		Description: "Write content to a file",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]ToolProperty{
				"path": {
					Type:        "string",
					Description: "The path to the file to write",
				},
				"content": {
					Type:        "string",
					Description: "The content to write to the file",
				},
				"mode": {
					Type:        "string",
					Description: "The write mode",
					Default:     "write",
					Enum:        []any{"write", "append"},
				},
			},
			Required: []string{"path", "content"},
		},
	},

	CodeExecutor: &Tool{
		Name:        "code_executor",
		Description: "Execute code in a specified programming language",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]ToolProperty{
				"language": {
					Type:        "string",
					Description: "The programming language",
					Enum:        []any{"python", "javascript", "bash", "go"},
				},
				"code": {
					Type:        "string",
					Description: "The code to execute",
				},
				"timeout": {
					Type:        "integer",
					Description: "Execution timeout in seconds",
					Default:     30,
					Minimum:     func() *float64 { v := 1.0; return &v }(),
					Maximum:     func() *float64 { v := 300.0; return &v }(),
				},
			},
			Required: []string{"language", "code"},
		},
	},
}
