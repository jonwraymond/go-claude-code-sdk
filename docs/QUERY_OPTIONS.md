# Claude Code SDK Query Options Documentation

This document provides a comprehensive guide to the query options supported by the Claude Code SDK and their compatibility with the Claude CLI.

## Table of Contents
- [Overview](#overview)
- [Supported Options](#supported-options)
- [Compatibility Matrix](#compatibility-matrix)
- [Usage Examples](#usage-examples)
- [CLI Flag Mappings](#cli-flag-mappings)
- [Best Practices](#best-practices)

## Overview

The Claude Code SDK supports various query options that control how Claude processes and responds to requests. These options are passed through the `QueryRequest` structure and are translated to appropriate CLI flags when executing the claude subprocess.

## Supported Options

### Core Options

#### Model (`model`)
- **Type**: `string`
- **Required**: Yes
- **Default**: `"claude-3-5-sonnet-20241022"`
- **Description**: Specifies which Claude model to use
- **CLI Flag**: `--model`
- **Example**: 
```go
request := &types.QueryRequest{
    Model: "claude-3-5-sonnet-20241022",
    // ... other fields
}
```

#### Messages (`messages`)
- **Type**: `[]Message`
- **Required**: Yes
- **Description**: The conversation history and current message
- **CLI Handling**: Passed as positional argument or through stdin
- **Example**:
```go
request := &types.QueryRequest{
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "Hello, Claude"},
        {Role: types.RoleAssistant, Content: "Hello! How can I help you?"},
        {Role: types.RoleUser, Content: "Can you explain recursion?"},
    },
}
```

#### Max Tokens (`max_tokens`)
- **Type**: `int`
- **Required**: Yes (SDK enforces this)
- **Default**: Model-dependent (typically 4096)
- **Description**: Maximum number of tokens in the response
- **CLI Flag**: Not directly supported by Claude CLI
- **Handling**: SDK manages this internally
- **Example**:
```go
request := &types.QueryRequest{
    MaxTokens: 1000,
    // ... other fields
}
```

### Generation Options

#### Temperature (`temperature`)
- **Type**: `float64`
- **Required**: No
- **Default**: `0.0`
- **Range**: `0.0` to `1.0`
- **Description**: Controls randomness in responses (0 = deterministic, 1 = maximum randomness)
- **CLI Flag**: `--temperature`
- **Example**:
```go
request := &types.QueryRequest{
    Temperature: 0.7,
    // ... other fields
}
```

#### Top P (`top_p`)
- **Type**: `float64`
- **Required**: No
- **Default**: `1.0`
- **Range**: `0.0` to `1.0`
- **Description**: Nucleus sampling parameter
- **CLI Flag**: Not supported by Claude CLI
- **Handling**: SDK includes warning when used

#### Top K (`top_k`)
- **Type**: `int`
- **Required**: No
- **Default**: Not set
- **Description**: Top-k sampling parameter
- **CLI Flag**: Not supported by Claude CLI
- **Handling**: SDK includes warning when used

#### Stop Sequences (`stop_sequences`)
- **Type**: `[]string`
- **Required**: No
- **Default**: Empty
- **Description**: Sequences that will stop generation
- **CLI Flag**: Not directly supported
- **Handling**: SDK manages internally

### Streaming Options

#### Stream (`stream`)
- **Type**: `bool`
- **Required**: No
- **Default**: `false`
- **Description**: Enable streaming response
- **CLI Flag**: `--format stream-json` (when true)
- **Example**:
```go
request := &types.QueryRequest{
    Stream: true,
    // ... other fields
}
```

### Tool Usage Options

#### Tools (`tools`)
- **Type**: `[]Tool`
- **Required**: No
- **Default**: Empty
- **Description**: Available tools that Claude can use
- **CLI Flag**: Managed through `--allowedTools`
- **Example**:
```go
request := &types.QueryRequest{
    Tools: []types.Tool{
        {
            Type: "function",
            Function: types.FunctionTool{
                Name: "get_weather",
                Description: "Get current weather",
                Parameters: weatherSchema,
            },
        },
    },
}
```

#### Tool Choice (`tool_choice`)
- **Type**: `any`
- **Required**: No
- **Default**: `"auto"`
- **Description**: Controls how Claude uses tools
- **Values**: `"auto"`, `"none"`, or specific tool name
- **CLI Handling**: SDK manages tool selection

### System Instructions

#### System (`system`)
- **Type**: `string`
- **Required**: No
- **Default**: Empty
- **Description**: System-level instructions for Claude
- **CLI Flag**: `--append-system-prompt`
- **Example**:
```go
request := &types.QueryRequest{
    System: "You are a helpful coding assistant.",
    // ... other fields
}
```

### Metadata

#### Metadata (`metadata`)
- **Type**: `map[string]any`
- **Required**: No
- **Default**: Empty
- **Description**: Additional request metadata
- **CLI Handling**: Not directly passed to CLI
- **Usage**: For SDK-level tracking and custom handling

## Compatibility Matrix

| Option | SDK Support | CLI Support | Notes |
|--------|-------------|-------------|-------|
| `model` | ✅ Full | ✅ Full | Direct mapping |
| `messages` | ✅ Full | ✅ Full | Passed as input |
| `max_tokens` | ✅ Full | ❌ No flag | SDK manages internally |
| `temperature` | ✅ Full | ✅ Full | Direct mapping |
| `top_p` | ⚠️ Warning | ❌ Not supported | SDK logs warning |
| `top_k` | ⚠️ Warning | ❌ Not supported | SDK logs warning |
| `stop_sequences` | ✅ Full | ❌ Not supported | SDK manages internally |
| `stream` | ✅ Full | ✅ Full | Maps to --format flag |
| `tools` | ✅ Full | ✅ Partial | Maps to --allowedTools |
| `tool_choice` | ✅ Full | ❌ Not supported | SDK manages selection |
| `system` | ✅ Full | ✅ Full | Maps to --append-system-prompt |
| `metadata` | ✅ Full | ❌ Not applicable | SDK-only feature |

### Legend:
- ✅ Full: Complete support
- ⚠️ Warning: Supported with warnings
- ❌ Not supported: Feature not available

## Usage Examples

### Basic Query
```go
request := &types.QueryRequest{
    Model: "claude-3-5-sonnet-20241022",
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "What is the capital of France?"},
    },
    MaxTokens: 100,
}

response, err := client.Query(ctx, request)
```

### Streaming Query
```go
request := &types.QueryRequest{
    Model: "claude-3-5-sonnet-20241022",
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "Write a short story"},
    },
    MaxTokens: 500,
    Stream: true,
}

stream, err := client.QueryStream(ctx, request)
```

### Query with Tools
```go
request := &types.QueryRequest{
    Model: "claude-3-5-sonnet-20241022",
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "What's the weather in Paris?"},
    },
    MaxTokens: 200,
    Tools: []types.Tool{weatherTool},
    ToolChoice: "auto",
}
```

### Query with System Prompt
```go
request := &types.QueryRequest{
    Model: "claude-3-5-sonnet-20241022",
    System: "You are a Python programming expert. Always provide code examples.",
    Messages: []types.Message{
        {Role: types.RoleUser, Content: "How do I read a CSV file?"},
    },
    MaxTokens: 500,
    Temperature: 0.3,
}
```

## CLI Flag Mappings

The SDK automatically translates QueryRequest options to CLI flags:

| SDK Option | CLI Flag | Example |
|------------|----------|---------|
| `Model` | `--model` | `--model claude-3-5-sonnet-20241022` |
| `Temperature` | `--temperature` | `--temperature 0.7` |
| `System` | `--append-system-prompt` | `--append-system-prompt "You are helpful"` |
| `Stream: true` | `--format stream-json` | `--format stream-json` |
| `Tools` | `--allowedTools` | `--allowedTools all` |

### Tool Permission Mappings

The SDK intelligently maps tool configurations:

```go
// No tools specified
request.Tools = nil
// CLI: --allowedTools none

// Specific tools specified
request.Tools = []Tool{tool1, tool2}
// CLI: --allowedTools all (SDK filters in response)

// All tools allowed (via config)
config.AllowAllTools = true
// CLI: --allowedTools all
```

## Best Practices

### 1. Always Set MaxTokens
The SDK requires MaxTokens to be set. Choose appropriate values based on expected response length:
- Short responses: 100-500 tokens
- Medium responses: 500-2000 tokens
- Long responses: 2000-4000 tokens

### 2. Use Appropriate Temperature
- `0.0`: Deterministic, best for factual/technical responses
- `0.3-0.7`: Balanced creativity and consistency
- `0.7-1.0`: Maximum creativity, may be less coherent

### 3. Handle Unsupported Options
When using options not supported by the CLI:
```go
// The SDK will log warnings for unsupported options
request := &types.QueryRequest{
    TopP: 0.9,  // Will trigger warning
    TopK: 50,   // Will trigger warning
    // ... other fields
}
```

### 4. Stream for Long Responses
Enable streaming for better user experience with long responses:
```go
if expectedLength > 1000 {
    request.Stream = true
}
```

### 5. Validate Before Sending
Always validate requests before sending:
```go
if err := request.Validate(); err != nil {
    log.Printf("Invalid request: %v", err)
    return err
}
```

## Troubleshooting

### Common Issues

1. **"unknown option" errors**
   - The SDK has been updated to use correct CLI flags
   - Ensure you're using the latest SDK version

2. **MaxTokens validation errors**
   - Always set MaxTokens > 0
   - Check model-specific limits

3. **Tool execution not working**
   - Verify tools are properly defined
   - Check Claude CLI has necessary permissions

4. **Streaming not working**
   - Ensure context is not cancelled prematurely
   - Check for process execution errors

### Debug Mode

Enable debug logging to see exact CLI commands:
```go
config := &types.ClaudeCodeConfig{
    Debug: true,  // Logs all CLI commands
    // ... other config
}
```

## Future Enhancements

Planned improvements for query options:

1. **Native top_p and top_k support** - When CLI adds these flags
2. **Advanced tool choice** - More granular tool selection
3. **Response format options** - JSON mode, structured outputs
4. **Retry configuration** - Built-in retry with backoff
5. **Request priority** - Queue management for multiple requests

## References

- [Claude API Documentation](https://docs.anthropic.com/claude/reference/messages)
- [Claude CLI Documentation](https://github.com/anthropics/claude-cli)
- [SDK Implementation](https://github.com/jonwraymond/go-claude-code-sdk)