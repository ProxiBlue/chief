package protocol

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixturesDir() string {
	// Walk up from internal/protocol to repo root
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "contract", "fixtures")); err == nil {
			return filepath.Join(dir, "contract", "fixtures")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func TestNewEnvelope(t *testing.T) {
	env := NewEnvelope(TypeSync, "device-123")

	if env.Type != TypeSync {
		t.Errorf("Type = %q, want %q", env.Type, TypeSync)
	}
	if env.DeviceID != "device-123" {
		t.Errorf("DeviceID = %q, want %q", env.DeviceID, "device-123")
	}
	if !strings.HasPrefix(env.ID, "msg_") {
		t.Errorf("ID = %q, want msg_ prefix", env.ID)
	}
	if env.Timestamp == "" {
		t.Error("Timestamp is empty")
	}
}

func TestNewEnvelopeUniqueIDs(t *testing.T) {
	e1 := NewEnvelope(TypeAck, "d1")
	e2 := NewEnvelope(TypeAck, "d1")
	if e1.ID == e2.ID {
		t.Errorf("expected unique IDs, got %q twice", e1.ID)
	}
}

func TestMarshalRoundtrip(t *testing.T) {
	env := NewEnvelope(TypeWelcome, "dev-abc")
	env.Payload = json.RawMessage(`{"server_version":"1.0.0","connection_id":"conn-1"}`)

	data, err := env.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	parsed, err := ParseEnvelope(data)
	if err != nil {
		t.Fatalf("ParseEnvelope: %v", err)
	}

	if parsed.Type != env.Type {
		t.Errorf("Type = %q, want %q", parsed.Type, env.Type)
	}
	if parsed.ID != env.ID {
		t.Errorf("ID = %q, want %q", parsed.ID, env.ID)
	}
	if parsed.DeviceID != env.DeviceID {
		t.Errorf("DeviceID = %q, want %q", parsed.DeviceID, env.DeviceID)
	}
}

func TestParseEnvelopeInvalid(t *testing.T) {
	_, err := ParseEnvelope([]byte(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDecodePayload(t *testing.T) {
	type WelcomePayload struct {
		ServerVersion string `json:"server_version"`
		ConnectionID  string `json:"connection_id"`
	}

	env := Envelope{
		Type:     TypeWelcome,
		ID:       "msg_test",
		DeviceID: "dev-1",
		Payload:  json.RawMessage(`{"server_version":"1.0.0","connection_id":"conn-xyz"}`),
	}

	p, err := DecodePayload[WelcomePayload](env)
	if err != nil {
		t.Fatalf("DecodePayload: %v", err)
	}
	if p.ServerVersion != "1.0.0" {
		t.Errorf("ServerVersion = %q, want %q", p.ServerVersion, "1.0.0")
	}
	if p.ConnectionID != "conn-xyz" {
		t.Errorf("ConnectionID = %q, want %q", p.ConnectionID, "conn-xyz")
	}
}

func TestDecodePayloadNoPayload(t *testing.T) {
	env := Envelope{Type: TypeAck, ID: "msg_test", DeviceID: "d1"}
	_, err := DecodePayload[map[string]string](env)
	if err == nil {
		t.Error("expected error for nil payload")
	}
}

func TestParseFixtures(t *testing.T) {
	root := fixturesDir()
	if root == "" {
		t.Skip("fixtures directory not found")
	}

	validFiles := []string{
		"envelope/valid_minimal.json",
		"envelope/valid_with_payload.json",
		"control/valid_welcome.json",
		"control/valid_ack.json",
		"control/valid_error.json",
		"state/valid_sync.json",
		"state/valid_prd_updated.json",
		"state/valid_run_completed.json",
		"cmd/valid_run_start.json",
		"cmd/valid_prd_create.json",
	}

	for _, f := range validFiles {
		t.Run("valid/"+f, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, f))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			env, err := ParseEnvelope(data)
			if err != nil {
				t.Fatalf("ParseEnvelope: %v", err)
			}
			if env.Type == "" {
				t.Error("Type is empty")
			}
			if env.ID == "" {
				t.Error("ID is empty")
			}
			if env.DeviceID == "" {
				t.Error("DeviceID is empty")
			}
			if env.Timestamp == "" {
				t.Error("Timestamp is empty")
			}
		})
	}
}

func TestTypeConstants(t *testing.T) {
	// Verify all 34 type constants are defined and non-empty
	types := []string{
		// Control
		TypeWelcome, TypeAck, TypeError,
		// State
		TypeSync, TypeProjectsUpdated, TypePRDCreated, TypePRDUpdated,
		TypePRDDeleted, TypePRDChatOutput, TypeRunStarted, TypeRunProgress,
		TypeRunOutput, TypeRunStopped, TypeRunCompleted, TypeDiffsResponse,
		TypeLogOutput, TypeLogResponse, TypeSettingsUpdated, TypeDeviceHeartbeat,
		TypeFilesList, TypeFileResponse, TypeProjectCloneProgress,
		// Command
		TypePRDCreate, TypePRDMessage, TypePRDUpdate, TypePRDDelete,
		TypeRunStart, TypeRunStop, TypeProjectClone, TypeDiffsGet,
		TypeLogGet, TypeFileGet, TypeSettingsGet, TypeSettingsUpdate,
	}

	if len(types) != 34 {
		t.Errorf("expected 34 type constants, got %d", len(types))
	}

	seen := make(map[string]bool)
	for _, typ := range types {
		if typ == "" {
			t.Error("found empty type constant")
		}
		if seen[typ] {
			t.Errorf("duplicate type constant: %q", typ)
		}
		seen[typ] = true
	}
}
