package uplink

import (
	"encoding/json"
	"fmt"

	"github.com/minicodemonkey/chief/internal/protocol"
)

// Handler dispatches incoming server commands to registered callbacks.
type Handler struct {
	deviceID string
	handlers map[string]func(protocol.Envelope) error
}

// NewHandler creates a new Handler for the given device.
func NewHandler(deviceID string) *Handler {
	return &Handler{
		deviceID: deviceID,
		handlers: make(map[string]func(protocol.Envelope) error),
	}
}

// Handle decodes an incoming envelope, dispatches to the registered handler,
// and returns an ack envelope on success or an error envelope on failure.
func (h *Handler) Handle(env protocol.Envelope) protocol.Envelope {
	fn, ok := h.handlers[env.Type]
	if !ok {
		// Check if this is a known command type at all.
		if !isCommandType(env.Type) {
			return h.errorEnvelope(env.ID, "unknown_command", fmt.Sprintf("unknown command type: %s", env.Type))
		}
		return h.errorEnvelope(env.ID, "command_failed", fmt.Sprintf("no handler registered for command: %s", env.Type))
	}

	if err := fn(env); err != nil {
		return h.errorEnvelope(env.ID, "command_failed", err.Error())
	}

	return h.ackEnvelope(env.ID)
}

func (h *Handler) ackEnvelope(refID string) protocol.Envelope {
	env := protocol.NewEnvelope(protocol.TypeAck, h.deviceID)
	payload, _ := json.Marshal(protocol.Ack{RefID: refID})
	env.Payload = payload
	return env
}

func (h *Handler) errorEnvelope(refID, code, message string) protocol.Envelope {
	env := protocol.NewEnvelope(protocol.TypeError, h.deviceID)
	payload, _ := json.Marshal(struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		RefID   string `json:"ref_id,omitempty"`
	}{
		Code:    code,
		Message: message,
		RefID:   refID,
	})
	env.Payload = payload
	return env
}

// dispatch is a generic helper that decodes the payload and calls the typed callback.
func dispatch[T any](env protocol.Envelope, fn func(T) error) error {
	payload, err := protocol.DecodePayload[T](env)
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}
	return fn(payload)
}

func (h *Handler) register(msgType string, fn func(protocol.Envelope) error) {
	h.handlers[msgType] = fn
}

// OnPRDCreate registers a handler for "prd-create" commands.
func (h *Handler) OnPRDCreate(fn func(protocol.CmdPRDCreate) error) {
	h.register(protocol.TypePRDCreate, func(env protocol.Envelope) error {
		return dispatch(env, fn)
	})
}

// OnPRDMessage registers a handler for "prd-message" commands.
func (h *Handler) OnPRDMessage(fn func(protocol.CmdPRDMessage) error) {
	h.register(protocol.TypePRDMessage, func(env protocol.Envelope) error {
		return dispatch(env, fn)
	})
}

// OnPRDUpdate registers a handler for "prd-update" commands.
func (h *Handler) OnPRDUpdate(fn func(protocol.CmdPRDUpdate) error) {
	h.register(protocol.TypePRDUpdate, func(env protocol.Envelope) error {
		return dispatch(env, fn)
	})
}

// OnPRDDelete registers a handler for "prd-delete" commands.
func (h *Handler) OnPRDDelete(fn func(protocol.CmdPRDDelete) error) {
	h.register(protocol.TypePRDDelete, func(env protocol.Envelope) error {
		return dispatch(env, fn)
	})
}

// OnRunStart registers a handler for "run-start" commands.
func (h *Handler) OnRunStart(fn func(protocol.CmdRunStart) error) {
	h.register(protocol.TypeRunStart, func(env protocol.Envelope) error {
		return dispatch(env, fn)
	})
}

// OnRunStop registers a handler for "run-stop" commands.
func (h *Handler) OnRunStop(fn func(protocol.CmdRunStop) error) {
	h.register(protocol.TypeRunStop, func(env protocol.Envelope) error {
		return dispatch(env, fn)
	})
}

// OnProjectClone registers a handler for "project-clone" commands.
func (h *Handler) OnProjectClone(fn func(protocol.CmdProjectClone) error) {
	h.register(protocol.TypeProjectClone, func(env protocol.Envelope) error {
		return dispatch(env, fn)
	})
}

// OnDiffsGet registers a handler for "diffs-get" commands.
func (h *Handler) OnDiffsGet(fn func(protocol.CmdDiffsGet) error) {
	h.register(protocol.TypeDiffsGet, func(env protocol.Envelope) error {
		return dispatch(env, fn)
	})
}

// OnLogGet registers a handler for "log-get" commands.
func (h *Handler) OnLogGet(fn func(protocol.CmdLogGet) error) {
	h.register(protocol.TypeLogGet, func(env protocol.Envelope) error {
		return dispatch(env, fn)
	})
}

// OnFilesList registers a handler for "files-list" commands.
func (h *Handler) OnFilesList(fn func(protocol.CmdFilesList) error) {
	h.register(protocol.TypeFilesList, func(env protocol.Envelope) error {
		return dispatch(env, fn)
	})
}

// OnFileGet registers a handler for "file-get" commands.
func (h *Handler) OnFileGet(fn func(protocol.CmdFileGet) error) {
	h.register(protocol.TypeFileGet, func(env protocol.Envelope) error {
		return dispatch(env, fn)
	})
}

// OnSettingsGet registers a handler for "settings-get" commands.
func (h *Handler) OnSettingsGet(fn func(protocol.CmdSettingsGet) error) {
	h.register(protocol.TypeSettingsGet, func(env protocol.Envelope) error {
		return dispatch(env, fn)
	})
}

// OnSettingsUpdate registers a handler for "settings-update" commands.
func (h *Handler) OnSettingsUpdate(fn func(protocol.CmdSettingsUpdate) error) {
	h.register(protocol.TypeSettingsUpdate, func(env protocol.Envelope) error {
		return dispatch(env, fn)
	})
}

// isCommandType returns true if the given type string is a known command type.
func isCommandType(t string) bool {
	switch t {
	case protocol.TypePRDCreate,
		protocol.TypePRDMessage,
		protocol.TypePRDUpdate,
		protocol.TypePRDDelete,
		protocol.TypeRunStart,
		protocol.TypeRunStop,
		protocol.TypeProjectClone,
		protocol.TypeDiffsGet,
		protocol.TypeLogGet,
		protocol.TypeFilesList,
		protocol.TypeFileGet,
		protocol.TypeSettingsGet,
		protocol.TypeSettingsUpdate:
		return true
	}
	return false
}
