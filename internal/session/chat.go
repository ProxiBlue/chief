// Package session provides chat session management for PRD conversations
// using Claude Code with --resume support for multi-turn interactions.
package session

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
)

// ChatEvent represents a streaming event from the chat session.
type ChatEvent struct {
	Type string // "text", "tool_use", "tool_result", "result", "error"
	Text string
}

// ChatSession manages a multi-turn Claude Code conversation for a PRD.
type ChatSession struct {
	dir       string
	cliPath   string
	sessionID string
	onEvent   func(ChatEvent)
	mu        sync.Mutex
	cmdFn     func(ctx context.Context, args []string) *exec.Cmd // for testing
}

// NewChatSession creates a new chat session for the given working directory.
// cliPath is the path to the Claude CLI binary (defaults to "claude" if empty).
func NewChatSession(dir, cliPath string) *ChatSession {
	if cliPath == "" {
		cliPath = "claude"
	}
	return &ChatSession{
		dir:     dir,
		cliPath: cliPath,
	}
}

// SessionID returns the current session ID, or empty string if no session yet.
func (s *ChatSession) SessionID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessionID
}

// OnEvent registers a callback for streaming events.
func (s *ChatSession) OnEvent(fn func(ChatEvent)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onEvent = fn
}

// SendMessage sends a message to the chat session and processes the streaming response.
// On the first call, it spawns a new Claude process. Subsequent calls resume the session.
func (s *ChatSession) SendMessage(ctx context.Context, message string) error {
	args := s.buildArgs(message)

	var cmd *exec.Cmd
	if s.cmdFn != nil {
		cmd = s.cmdFn(ctx, args)
	} else {
		cmd = exec.CommandContext(ctx, s.cliPath, args...)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start claude: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // 1MB max line size

	for scanner.Scan() {
		line := scanner.Text()
		s.processLine(line)
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("claude exited with error: %w", err)
	}

	return nil
}

// buildArgs constructs the CLI arguments for a Claude invocation.
func (s *ChatSession) buildArgs(message string) []string {
	args := []string{
		"--print",
		"--output-format", "stream-json",
		"--verbose",
		"--dangerously-skip-permissions",
		"--dir", s.dir,
	}

	s.mu.Lock()
	sid := s.sessionID
	s.mu.Unlock()

	if sid != "" {
		args = append(args, "--resume", sid)
	}

	args = append(args, "-p", message)
	return args
}

// streamMessage is the top-level NDJSON structure from Claude's stream-json output.
type streamMessage struct {
	Type      string          `json:"type"`
	Subtype   string          `json:"subtype,omitempty"`
	SessionID string          `json:"session_id,omitempty"`
	Message   json.RawMessage `json:"message,omitempty"`
}

// assistantMessage holds the content blocks from an assistant message.
type assistantMessage struct {
	Content []contentBlock `json:"content"`
}

// contentBlock represents a single content block (text or tool_use).
type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Name string `json:"name,omitempty"`
}

// processLine parses a single NDJSON line and emits events.
func (s *ChatSession) processLine(line string) {
	var msg streamMessage
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return
	}

	switch msg.Type {
	case "result":
		if msg.SessionID != "" {
			s.mu.Lock()
			s.sessionID = msg.SessionID
			s.mu.Unlock()
		}

	case "assistant":
		if msg.Message == nil {
			return
		}
		var am assistantMessage
		if err := json.Unmarshal(msg.Message, &am); err != nil {
			return
		}
		for _, block := range am.Content {
			s.emit(ChatEvent{Type: block.Type, Text: block.Text})
		}
	}
}

// emit fires the onEvent callback if registered.
func (s *ChatSession) emit(event ChatEvent) {
	s.mu.Lock()
	fn := s.onEvent
	s.mu.Unlock()
	if fn != nil {
		fn(event)
	}
}
