package uplink

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/minicodemonkey/chief/internal/protocol"
)

func makeCommandEnvelope(t *testing.T, msgType string, payload any) protocol.Envelope {
	t.Helper()
	env := protocol.NewEnvelope(msgType, "dev_test")
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		env.Payload = data
	}
	return env
}

func TestHandlerDispatchRunStart(t *testing.T) {
	h := NewHandler("dev_test")

	var got protocol.CmdRunStart
	h.OnRunStart(func(cmd protocol.CmdRunStart) error {
		got = cmd
		return nil
	})

	env := makeCommandEnvelope(t, protocol.TypeRunStart, protocol.CmdRunStart{
		PRDID: "prd_123",
		RunID: "run_456",
	})

	resp := h.Handle(env)

	if resp.Type != protocol.TypeAck {
		t.Fatalf("expected ack, got %s", resp.Type)
	}

	var ack protocol.Ack
	if err := json.Unmarshal(resp.Payload, &ack); err != nil {
		t.Fatalf("unmarshal ack: %v", err)
	}
	if ack.RefID != env.ID {
		t.Errorf("ack ref_id = %q, want %q", ack.RefID, env.ID)
	}
	if got.PRDID != "prd_123" {
		t.Errorf("PRDID = %q, want %q", got.PRDID, "prd_123")
	}
	if got.RunID != "run_456" {
		t.Errorf("RunID = %q, want %q", got.RunID, "run_456")
	}
}

func TestHandlerDispatchPRDCreate(t *testing.T) {
	h := NewHandler("dev_test")

	var got protocol.CmdPRDCreate
	h.OnPRDCreate(func(cmd protocol.CmdPRDCreate) error {
		got = cmd
		return nil
	})

	env := makeCommandEnvelope(t, protocol.TypePRDCreate, protocol.CmdPRDCreate{
		ProjectID: "proj_abc",
		Title:     "My PRD",
		Content:   "Some content",
	})

	resp := h.Handle(env)

	if resp.Type != protocol.TypeAck {
		t.Fatalf("expected ack, got %s", resp.Type)
	}
	if got.ProjectID != "proj_abc" {
		t.Errorf("ProjectID = %q, want %q", got.ProjectID, "proj_abc")
	}
	if got.Title != "My PRD" {
		t.Errorf("Title = %q, want %q", got.Title, "My PRD")
	}
}

func TestHandlerUnknownCommand(t *testing.T) {
	h := NewHandler("dev_test")

	env := makeCommandEnvelope(t, "totally-unknown", nil)
	resp := h.Handle(env)

	if resp.Type != protocol.TypeError {
		t.Fatalf("expected error, got %s", resp.Type)
	}

	var errPayload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		RefID   string `json:"ref_id"`
	}
	if err := json.Unmarshal(resp.Payload, &errPayload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if errPayload.Code != "unknown_command" {
		t.Errorf("code = %q, want %q", errPayload.Code, "unknown_command")
	}
	if errPayload.RefID != env.ID {
		t.Errorf("ref_id = %q, want %q", errPayload.RefID, env.ID)
	}
}

func TestHandlerUnregisteredCommand(t *testing.T) {
	h := NewHandler("dev_test")
	// Don't register any handler for run-start (a known command type)

	env := makeCommandEnvelope(t, protocol.TypeRunStart, protocol.CmdRunStart{
		PRDID: "prd_123",
		RunID: "run_456",
	})
	resp := h.Handle(env)

	if resp.Type != protocol.TypeError {
		t.Fatalf("expected error, got %s", resp.Type)
	}

	var errPayload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(resp.Payload, &errPayload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if errPayload.Code != "command_failed" {
		t.Errorf("code = %q, want %q", errPayload.Code, "command_failed")
	}
}

func TestHandlerCallbackError(t *testing.T) {
	h := NewHandler("dev_test")

	h.OnRunStop(func(cmd protocol.CmdRunStop) error {
		return errors.New("something went wrong")
	})

	env := makeCommandEnvelope(t, protocol.TypeRunStop, protocol.CmdRunStop{
		RunID: "run_789",
	})
	resp := h.Handle(env)

	if resp.Type != protocol.TypeError {
		t.Fatalf("expected error, got %s", resp.Type)
	}

	var errPayload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(resp.Payload, &errPayload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if errPayload.Code != "command_failed" {
		t.Errorf("code = %q, want %q", errPayload.Code, "command_failed")
	}
	if errPayload.Message != "something went wrong" {
		t.Errorf("message = %q, want %q", errPayload.Message, "something went wrong")
	}
}

func TestHandlerAckEnvelopeFields(t *testing.T) {
	h := NewHandler("dev_abc")

	h.OnSettingsGet(func(cmd protocol.CmdSettingsGet) error {
		return nil
	})

	env := makeCommandEnvelope(t, protocol.TypeSettingsGet, protocol.CmdSettingsGet{})
	resp := h.Handle(env)

	if resp.Type != protocol.TypeAck {
		t.Fatalf("expected ack, got %s", resp.Type)
	}
	if resp.DeviceID != "dev_abc" {
		t.Errorf("DeviceID = %q, want %q", resp.DeviceID, "dev_abc")
	}
	if resp.ID == "" {
		t.Error("response ID should not be empty")
	}
	if resp.Timestamp == "" {
		t.Error("response Timestamp should not be empty")
	}
}

func TestHandlerAllCommandTypes(t *testing.T) {
	h := NewHandler("dev_test")

	// Register all 13 command handlers and verify each one dispatches correctly.
	called := make(map[string]bool)

	h.OnPRDCreate(func(cmd protocol.CmdPRDCreate) error { called["prd-create"] = true; return nil })
	h.OnPRDMessage(func(cmd protocol.CmdPRDMessage) error { called["prd-message"] = true; return nil })
	h.OnPRDUpdate(func(cmd protocol.CmdPRDUpdate) error { called["prd-update"] = true; return nil })
	h.OnPRDDelete(func(cmd protocol.CmdPRDDelete) error { called["prd-delete"] = true; return nil })
	h.OnRunStart(func(cmd protocol.CmdRunStart) error { called["run-start"] = true; return nil })
	h.OnRunStop(func(cmd protocol.CmdRunStop) error { called["run-stop"] = true; return nil })
	h.OnProjectClone(func(cmd protocol.CmdProjectClone) error { called["project-clone"] = true; return nil })
	h.OnDiffsGet(func(cmd protocol.CmdDiffsGet) error { called["diffs-get"] = true; return nil })
	h.OnLogGet(func(cmd protocol.CmdLogGet) error { called["log-get"] = true; return nil })
	h.OnFilesList(func(cmd protocol.CmdFilesList) error { called["files-list"] = true; return nil })
	h.OnFileGet(func(cmd protocol.CmdFileGet) error { called["file-get"] = true; return nil })
	h.OnSettingsGet(func(cmd protocol.CmdSettingsGet) error { called["settings-get"] = true; return nil })
	h.OnSettingsUpdate(func(cmd protocol.CmdSettingsUpdate) error { called["settings-update"] = true; return nil })

	commands := []struct {
		typ     string
		payload any
	}{
		{protocol.TypePRDCreate, protocol.CmdPRDCreate{ProjectID: "p", Title: "t", Content: "c"}},
		{protocol.TypePRDMessage, protocol.CmdPRDMessage{PRDID: "p", Message: "m"}},
		{protocol.TypePRDUpdate, protocol.CmdPRDUpdate{PRDID: "p"}},
		{protocol.TypePRDDelete, protocol.CmdPRDDelete{PRDID: "p"}},
		{protocol.TypeRunStart, protocol.CmdRunStart{PRDID: "p", RunID: "r"}},
		{protocol.TypeRunStop, protocol.CmdRunStop{RunID: "r"}},
		{protocol.TypeProjectClone, protocol.CmdProjectClone{GitURL: "url"}},
		{protocol.TypeDiffsGet, protocol.CmdDiffsGet{ProjectID: "p"}},
		{protocol.TypeLogGet, protocol.CmdLogGet{ProjectID: "p"}},
		{protocol.TypeFilesList, protocol.CmdFilesList{ProjectID: "p"}},
		{protocol.TypeFileGet, protocol.CmdFileGet{ProjectID: "p", Path: "/"}},
		{protocol.TypeSettingsGet, protocol.CmdSettingsGet{}},
		{protocol.TypeSettingsUpdate, protocol.CmdSettingsUpdate{Settings: json.RawMessage(`{}`)}},
	}

	for _, cmd := range commands {
		env := makeCommandEnvelope(t, cmd.typ, cmd.payload)
		resp := h.Handle(env)
		if resp.Type != protocol.TypeAck {
			t.Errorf("%s: expected ack, got %s", cmd.typ, resp.Type)
		}
	}

	if len(called) != 13 {
		t.Errorf("expected 13 handlers called, got %d", len(called))
		for _, cmd := range commands {
			if !called[cmd.typ] {
				t.Errorf("handler not called: %s", cmd.typ)
			}
		}
	}
}
