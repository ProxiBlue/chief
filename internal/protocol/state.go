package protocol

import "encoding/json"

// StateSync is the payload for a "sync" state message.
type StateSync struct {
	Projects []Project        `json:"projects"`
	PRDs     []PRD            `json:"prds"`
	Runs     []Run            `json:"runs"`
	Settings json.RawMessage  `json:"settings,omitempty"`
}

// StateProjectsUpdated is the payload for a "projects-updated" state message.
type StateProjectsUpdated struct {
	Projects []Project `json:"projects"`
}

// StatePRDCreated is the payload for a "prd-created" state message.
type StatePRDCreated struct {
	PRD PRD `json:"prd"`
}

// StatePRDUpdated is the payload for a "prd-updated" state message.
type StatePRDUpdated struct {
	PRD PRD `json:"prd"`
}

// StatePRDDeleted is the payload for a "prd-deleted" state message.
type StatePRDDeleted struct {
	PRDID     string `json:"prd_id"`
	ProjectID string `json:"project_id"`
}

// StatePRDChatOutput is the payload for a "prd-chat-output" state message.
type StatePRDChatOutput struct {
	PRDID   string `json:"prd_id"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

// StateRunStarted is the payload for a "run-started" state message.
type StateRunStarted struct {
	Run Run `json:"run"`
}

// StateRunProgress is the payload for a "run-progress" state message.
type StateRunProgress struct {
	RunID      string `json:"run_id"`
	PRDID      string `json:"prd_id"`
	Message    string `json:"message"`
	Percentage *int   `json:"percentage,omitempty"`
}

// StateRunOutput is the payload for a "run-output" state message.
type StateRunOutput struct {
	RunID  string `json:"run_id"`
	Stream string `json:"stream"`
	Data   string `json:"data"`
}

// StateRunStopped is the payload for a "run-stopped" state message.
type StateRunStopped struct {
	RunID  string `json:"run_id"`
	PRDID  string `json:"prd_id"`
	Reason string `json:"reason,omitempty"`
}

// StateRunCompleted is the payload for a "run-completed" state message.
type StateRunCompleted struct {
	RunID      string `json:"run_id"`
	PRDID      string `json:"prd_id"`
	Status     string `json:"status"`
	Result     string `json:"result"`
	Error      string `json:"error,omitempty"`
	FinishedAt string `json:"finished_at,omitempty"`
}

// StateDiffsResponse is the payload for a "diffs-response" state message.
type StateDiffsResponse struct {
	ProjectID string      `json:"project_id"`
	Ref       string      `json:"ref,omitempty"`
	Diffs     []DiffEntry `json:"diffs"`
}

// StateLogOutput is the payload for a "log-output" state message.
type StateLogOutput struct {
	Level   string          `json:"level"`
	Message string          `json:"message"`
	Context json.RawMessage `json:"context,omitempty"`
}

// StateLogResponse is the payload for a "log-response" state message.
type StateLogResponse struct {
	RunID   string     `json:"run_id"`
	Entries []LogEntry `json:"entries"`
}

// StateSettingsUpdated is the payload for a "settings-updated" state message.
type StateSettingsUpdated struct {
	Settings json.RawMessage `json:"settings"`
}

// StateDeviceHeartbeat is the payload for a "device-heartbeat" state message.
type StateDeviceHeartbeat struct {
	UptimeSeconds int `json:"uptime_seconds"`
	ActiveRuns    int `json:"active_runs"`
}

// StateFilesList is the payload for a "files-list" state message (response).
type StateFilesList struct {
	ProjectID string      `json:"project_id"`
	Path      string      `json:"path"`
	Files     []FileEntry `json:"files"`
}

// StateFileResponse is the payload for a "file-response" state message.
type StateFileResponse struct {
	ProjectID string `json:"project_id"`
	Path      string `json:"path"`
	Content   string `json:"content"`
	Encoding  string `json:"encoding"`
}

// StateProjectCloneProgress is the payload for a "project-clone-progress" state message.
type StateProjectCloneProgress struct {
	ProjectID  string `json:"project_id"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	Percentage *int   `json:"percentage,omitempty"`
}
