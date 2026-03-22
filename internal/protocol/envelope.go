package protocol

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Envelope is the outer message format for CLI-server communication.
type Envelope struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	DeviceID  string          `json:"device_id"`
	Timestamp string          `json:"timestamp"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// NewEnvelope creates a new Envelope with a generated message ID and current timestamp.
func NewEnvelope(msgType, deviceID string) Envelope {
	return Envelope{
		Type:      msgType,
		ID:        generateID(),
		DeviceID:  deviceID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// Marshal serializes the envelope to JSON bytes.
func (e Envelope) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// ParseEnvelope deserializes JSON bytes into an Envelope.
func ParseEnvelope(data []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return Envelope{}, fmt.Errorf("parse envelope: %w", err)
	}
	return env, nil
}

// DecodePayload extracts a typed payload from an envelope.
func DecodePayload[T any](env Envelope) (T, error) {
	var payload T
	if env.Payload == nil {
		return payload, fmt.Errorf("envelope has no payload")
	}
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		return payload, fmt.Errorf("decode payload: %w", err)
	}
	return payload, nil
}

func generateID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return "msg_" + hex.EncodeToString(b)
}
