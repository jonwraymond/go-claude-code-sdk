# CLI Flag Compatibility Fixes

This document summarizes the fixes made to ensure the Go Claude Code SDK uses the correct CLI flags.

## Changes Made

### 1. Fixed System Prompt Flag
- **Old**: `--system`
- **New**: `--append-system-prompt`
- **Files**: 
  - `pkg/client/claude_code_client.go` (line 801)
  - `pkg/client/query.go` (line 381)

### 2. Removed Unsupported Flags
- **Removed**: `--max-tokens` (line 805-807 in claude_code_client.go)
- **Removed**: `--temperature` (line 810-812 in claude_code_client.go)
- **Removed**: `--stream` (line 788 in claude_code_client.go)
- **Removed**: `--timeout` (query.go)
- **Note**: Added comments explaining these are not supported by Claude CLI

### 3. Fixed Tool Permissions Flag
- **Old**: `--tools`
- **New**: `--allowedTools`
- **File**: `pkg/client/query.go` (line 400)

### 4. Fixed Permission Mode
- **Old**: `--accept-edits` and `--reject-edits` as separate flags
- **New**: `--permission-mode` with values "acceptEdits", "default", etc.
- **File**: `pkg/client/query.go` (lines 390-394)

### 5. Correctly Maintained Flags
These flags were already correct:
- `--model` ✓
- `--session-id` ✓
- `--print` ✓
- `--mcp-config` ✓

## Verification

Created `test_fixed_flags.go` to verify all flag fixes work correctly. The test confirms:
- System prompts work with `--append-system-prompt`
- Tool restrictions work with `--allowedTools`
- Permission modes work with `--permission-mode`
- Model and session ID flags continue to work correctly

## Impact

These changes ensure the SDK is fully compatible with the actual Claude Code CLI implementation, preventing errors from invalid flags and ensuring all features work as expected.