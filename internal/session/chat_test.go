package session

import (
	"context"
	"os/exec"
	"testing"
)

func TestBuildArgsFirstTurn(t *testing.T) {
	s := NewChatSession("/tmp/project", "claude")
	args := s.buildArgs("hello world")

	expected := []string{
		"--print",
		"--output-format", "stream-json",
		"--verbose",
		"--dangerously-skip-permissions",
		"--dir", "/tmp/project",
		"-p", "hello world",
	}

	if len(args) != len(expected) {
		t.Fatalf("expected %d args, got %d: %v", len(expected), len(args), args)
	}

	for i, exp := range expected {
		if args[i] != exp {
			t.Errorf("arg[%d] = %q, want %q", i, args[i], exp)
		}
	}
}

func TestBuildArgsWithResume(t *testing.T) {
	s := NewChatSession("/tmp/project", "claude")
	s.sessionID = "sess_abc123"

	args := s.buildArgs("follow up")

	expected := []string{
		"--print",
		"--output-format", "stream-json",
		"--verbose",
		"--dangerously-skip-permissions",
		"--dir", "/tmp/project",
		"--resume", "sess_abc123",
		"-p", "follow up",
	}

	if len(args) != len(expected) {
		t.Fatalf("expected %d args, got %d: %v", len(expected), len(args), args)
	}

	for i, exp := range expected {
		if args[i] != exp {
			t.Errorf("arg[%d] = %q, want %q", i, args[i], exp)
		}
	}
}

func TestSessionIDExtraction(t *testing.T) {
	s := NewChatSession("/tmp/project", "claude")

	if s.SessionID() != "" {
		t.Error("expected empty session ID initially")
	}

	// Simulate processing a result line with session_id
	s.processLine(`{"type":"result","session_id":"sess_xyz789","subtype":"success"}`)

	if got := s.SessionID(); got != "sess_xyz789" {
		t.Errorf("SessionID() = %q, want %q", got, "sess_xyz789")
	}
}

func TestOnEventCallback(t *testing.T) {
	s := NewChatSession("/tmp/project", "claude")

	var events []ChatEvent
	s.OnEvent(func(e ChatEvent) {
		events = append(events, e)
	})

	// Simulate an assistant text message
	s.processLine(`{"type":"assistant","message":{"content":[{"type":"text","text":"Hello!"}]}}`)

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != "text" {
		t.Errorf("event type = %q, want %q", events[0].Type, "text")
	}
	if events[0].Text != "Hello!" {
		t.Errorf("event text = %q, want %q", events[0].Text, "Hello!")
	}
}

func TestOnEventToolUse(t *testing.T) {
	s := NewChatSession("/tmp/project", "claude")

	var events []ChatEvent
	s.OnEvent(func(e ChatEvent) {
		events = append(events, e)
	})

	s.processLine(`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read"}]}}`)

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != "tool_use" {
		t.Errorf("event type = %q, want %q", events[0].Type, "tool_use")
	}
}

func TestProcessLineIgnoresInvalidJSON(t *testing.T) {
	s := NewChatSession("/tmp/project", "claude")

	var events []ChatEvent
	s.OnEvent(func(e ChatEvent) {
		events = append(events, e)
	})

	// Should not panic or emit events
	s.processLine("not json at all")
	s.processLine("")
	s.processLine(`{"type":"system","subtype":"init"}`)

	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestDefaultCLIPath(t *testing.T) {
	s := NewChatSession("/tmp/project", "")
	if s.cliPath != "claude" {
		t.Errorf("cliPath = %q, want %q", s.cliPath, "claude")
	}
}

func TestSendMessageIntegration(t *testing.T) {
	s := NewChatSession("/tmp", "echo")

	// Use a mock command that outputs NDJSON
	s.cmdFn = func(ctx context.Context, args []string) *exec.Cmd {
		// Echo NDJSON lines simulating Claude output
		output := `{"type":"assistant","message":{"content":[{"type":"text","text":"Hi there"}]}}` + "\n" +
			`{"type":"result","session_id":"sess_integration"}` + "\n"
		cmd := exec.CommandContext(ctx, "printf", "%s", output)
		return cmd
	}

	var events []ChatEvent
	s.OnEvent(func(e ChatEvent) {
		events = append(events, e)
	})

	err := s.SendMessage(context.Background(), "hello")
	if err != nil {
		t.Fatalf("SendMessage() error: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Text != "Hi there" {
		t.Errorf("event text = %q, want %q", events[0].Text, "Hi there")
	}

	if got := s.SessionID(); got != "sess_integration" {
		t.Errorf("SessionID() = %q, want %q", got, "sess_integration")
	}
}
