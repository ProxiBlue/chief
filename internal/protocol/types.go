package protocol

// Message type constants for all protocol message types.
const (
	// Control messages
	TypeWelcome = "welcome"
	TypeAck     = "ack"
	TypeError   = "error"

	// State messages
	TypeSync                 = "sync"
	TypeProjectsUpdated      = "projects-updated"
	TypePRDCreated           = "prd-created"
	TypePRDUpdated           = "prd-updated"
	TypePRDDeleted           = "prd-deleted"
	TypePRDChatOutput        = "prd-chat-output"
	TypeRunStarted           = "run-started"
	TypeRunProgress          = "run-progress"
	TypeRunOutput            = "run-output"
	TypeRunStopped           = "run-stopped"
	TypeRunCompleted         = "run-completed"
	TypeDiffsResponse        = "diffs-response"
	TypeLogOutput            = "log-output"
	TypeLogResponse          = "log-response"
	TypeSettingsUpdated      = "settings-updated"
	TypeDeviceHeartbeat      = "device-heartbeat"
	TypeFilesList            = "files-list"
	TypeFileResponse         = "file-response"
	TypeProjectCloneProgress = "project-clone-progress"

	// Command messages
	TypePRDCreate      = "prd-create"
	TypePRDMessage     = "prd-message"
	TypePRDUpdate      = "prd-update"
	TypePRDDelete      = "prd-delete"
	TypeRunStart       = "run-start"
	TypeRunStop        = "run-stop"
	TypeProjectClone   = "project-clone"
	TypeDiffsGet       = "diffs-get"
	TypeLogGet         = "log-get"
	TypeFileGet        = "file-get"
	TypeSettingsGet    = "settings-get"
	TypeSettingsUpdate = "settings-update"
)
