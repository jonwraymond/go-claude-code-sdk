package internal

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
)

// SubprocessCLITransport handles communication with Claude CLI via subprocess.
type SubprocessCLITransport struct {
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	stdout      io.ReadCloser
	stderr      io.ReadCloser
	options     *ClaudeCodeOptions
	prompt      interface{}
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	connected   bool
	messageChan chan []byte
}

// NewSubprocessCLITransport creates a new subprocess transport.
func NewSubprocessCLITransport(prompt interface{}, options *ClaudeCodeOptions) *SubprocessCLITransport {
	ctx, cancel := context.WithCancel(context.Background())
	return &SubprocessCLITransport{
		options:     options,
		prompt:      prompt,
		ctx:         ctx,
		cancel:      cancel,
		messageChan: make(chan []byte, 100),
	}
}

// Connect establishes connection to Claude CLI.
func (t *SubprocessCLITransport) Connect() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.connected {
		return nil
	}

	// Find Claude CLI executable
	claudeCmd, err := exec.LookPath("claude")
	if err != nil {
		return NewCLINotFoundError("Claude CLI not found in PATH", err)
	}

	// Set environment variable for SDK entrypoint
	env := os.Environ()
	env = append(env, "CLAUDE_CODE_ENTRYPOINT=sdk-go")

	// Create the command with appropriate args
	args := []string{"code", "start", "--streaming"}
	
	// Add options as CLI flags
	if t.options != nil {
		if t.options.SystemPrompt != nil && *t.options.SystemPrompt != "" {
			args = append(args, "--system-prompt", *t.options.SystemPrompt)
		}
		if t.options.MaxTurns != nil {
			args = append(args, "--max-turns", fmt.Sprintf("%d", *t.options.MaxTurns))
		}
		if t.options.PermissionMode != nil {
			args = append(args, "--permission-mode", string(*t.options.PermissionMode))
		}
		if t.options.CWD != nil && *t.options.CWD != "" {
			args = append(args, "--cwd", *t.options.CWD)
		}
		if len(t.options.AllowedTools) > 0 {
			for _, tool := range t.options.AllowedTools {
				args = append(args, "--allowed-tool", tool)
			}
		}
		if len(t.options.DisallowedTools) > 0 {
			for _, tool := range t.options.DisallowedTools {
				args = append(args, "--disallowed-tool", tool)
			}
		}
		if t.options.Model != nil && *t.options.Model != "" {
			args = append(args, "--model", *t.options.Model)
		}
	}

	// Create command with context
	t.cmd = exec.CommandContext(t.ctx, claudeCmd, args...)
	t.cmd.Env = env

	// Set working directory if specified
	if t.options != nil && t.options.CWD != nil {
		if filepath.IsAbs(*t.options.CWD) {
			t.cmd.Dir = *t.options.CWD
		} else {
			// Make relative path absolute
			if abs, err := filepath.Abs(*t.options.CWD); err == nil {
				t.cmd.Dir = abs
			}
		}
	}

	// Create pipes
	stdin, err := t.cmd.StdinPipe()
	if err != nil {
		return NewCLIConnectionError("Failed to create stdin pipe", err)
	}
	t.stdin = stdin

	stdout, err := t.cmd.StdoutPipe()
	if err != nil {
		return NewCLIConnectionError("Failed to create stdout pipe", err)
	}
	t.stdout = stdout

	stderr, err := t.cmd.StderrPipe()
	if err != nil {
		return NewCLIConnectionError("Failed to create stderr pipe", err)
	}
	t.stderr = stderr

	// Start the process
	if err := t.cmd.Start(); err != nil {
		return NewCLIConnectionError("Failed to start Claude CLI", err)
	}

	t.connected = true

	// Start reading output in background
	go t.readOutput()

	return nil
}

// readOutput reads output from Claude CLI and forwards to message channel.
func (t *SubprocessCLITransport) readOutput() {
	defer close(t.messageChan)

	scanner := bufio.NewScanner(t.stdout)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) > 0 {
			// Copy the bytes since scanner reuses the buffer
			lineCopy := make([]byte, len(line))
			copy(lineCopy, line)
			
			select {
			case t.messageChan <- lineCopy:
			case <-t.ctx.Done():
				return
			}
		}
	}

	if err := scanner.Err(); err != nil && t.ctx.Err() == nil {
		// Log error but don't fail - connection might be closing
		fmt.Fprintf(os.Stderr, "Error reading from Claude CLI: %v\n", err)
	}
}

// SendRequest sends a request to Claude CLI.
func (t *SubprocessCLITransport) SendRequest(messages []map[string]interface{}, metadata map[string]interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected {
		return NewCLIConnectionError("Not connected to Claude CLI", nil)
	}

	// Create the request payload
	request := map[string]interface{}{
		"messages": messages,
	}
	if metadata != nil {
		for k, v := range metadata {
			request[k] = v
		}
	}

	// Send as JSON
	data, err := json.Marshal(request)
	if err != nil {
		return NewClaudeSDKError("Failed to marshal request", err)
	}

	_, err = t.stdin.Write(append(data, '\n'))
	if err != nil {
		return NewCLIConnectionError("Failed to write to Claude CLI", err)
	}

	return nil
}

// ReceiveMessages returns a channel of messages from Claude CLI.
func (t *SubprocessCLITransport) ReceiveMessages() <-chan []byte {
	return t.messageChan
}

// Interrupt sends interrupt signal to Claude CLI.
func (t *SubprocessCLITransport) Interrupt() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected || t.cmd == nil || t.cmd.Process == nil {
		return NewCLIConnectionError("Not connected to Claude CLI", nil)
	}

	// Send SIGINT to the process
	if err := t.cmd.Process.Signal(syscall.SIGINT); err != nil {
		return NewProcessError("Failed to send interrupt signal", 0, err)
	}

	return nil
}

// Disconnect closes the connection to Claude CLI.
func (t *SubprocessCLITransport) Disconnect() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected {
		return nil
	}

	t.connected = false
	t.cancel() // Cancel context

	// Close stdin to signal end of input
	if t.stdin != nil {
		t.stdin.Close()
	}

	// Wait for process to exit
	if t.cmd != nil && t.cmd.Process != nil {
		if err := t.cmd.Wait(); err != nil {
			// Process might exit with non-zero code, which is often normal
			if exitErr, ok := err.(*exec.ExitError); ok {
				// Log but don't return error for non-zero exit codes
				fmt.Fprintf(os.Stderr, "Claude CLI exited with code %d\n", exitErr.ExitCode())
			} else {
				return NewProcessError("Error waiting for Claude CLI to exit", 0, err)
			}
		}
	}

	// Close remaining pipes
	if t.stdout != nil {
		t.stdout.Close()
	}
	if t.stderr != nil {
		t.stderr.Close()
	}

	return nil
}