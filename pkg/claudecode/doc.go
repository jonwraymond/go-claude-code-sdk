// Package claudecode provides a Go client for interacting with Claude Code.
//
// This package contains the primary client implementation and the main query
// functionality. It provides both one-shot queries and bidirectional streaming
// conversations with Claude Code via the CLI.
//
// Basic usage:
//
//	// One-shot query
//	messages, err := claudecode.Query(ctx, "Hello, Claude!", nil)
//	for msg := range messages {
//	    fmt.Println(msg.GetRole(), msg.GetContent())
//	}
//
//	// Interactive client
//	client := claudecode.NewClaudeSDKClient(nil)
//	err := client.Connect(ctx, "Let's have a conversation")
//	err = client.SendMessage(ctx, "Tell me about Go")
//	for msg := range client.ReceiveMessages() {
//	    fmt.Println(msg.GetContent())
//	}
//
// The package follows Go conventions for error handling and context support,
// providing idiomatic interfaces while maintaining compatibility with the
// official Python SDK functionality.
package claudecode