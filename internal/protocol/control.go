package protocol

// Welcome is the payload for a "welcome" control message.
type Welcome struct {
	ServerVersion string `json:"server_version"`
	ConnectionID  string `json:"connection_id"`
}

// Ack is the payload for an "ack" control message.
type Ack struct {
	RefID string `json:"ref_id"`
}

// Error is the payload for an "error" control message.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	RefID   string `json:"ref_id,omitempty"`
}
