package types

// StringPtr is a helper function to create a pointer to a string.
// This is useful for setting optional string fields in ClaudeCodeOptions.
//
// Example:
//
//	options := NewClaudeCodeOptions()
//	options.SystemPrompt = StringPtr("You are a helpful assistant")
func StringPtr(s string) *string {
	return &s
}

// IntPtr is a helper function to create a pointer to an int.
// This is useful for setting optional int fields in ClaudeCodeOptions.
//
// Example:
//
//	options := NewClaudeCodeOptions()
//	options.MaxTurns = IntPtr(5)
func IntPtr(i int) *int {
	return &i
}

// PermissionModePtr is a helper function to create a pointer to a PermissionMode.
// This is useful for setting the PermissionMode field in ClaudeCodeOptions.
//
// Example:
//
//	options := NewClaudeCodeOptions()
//	options.PermissionMode = PermissionModePtr(PermissionModeAcceptEdits)
func PermissionModePtr(pm PermissionMode) *PermissionMode {
	return &pm
}