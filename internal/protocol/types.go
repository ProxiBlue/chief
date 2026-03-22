package protocol

import "encoding/json"

// Shared domain types used across state and command payloads.

// Project represents a tracked project.
type Project struct {
	ID         string     `json:"id"`
	Path       string     `json:"path"`
	Name       string     `json:"name"`
	GitRemote  string     `json:"git_remote,omitempty"`
	GitBranch  string     `json:"git_branch,omitempty"`
	GitSHA     string     `json:"git_sha,omitempty"`
	GitStatus  string     `json:"git_status,omitempty"`
	LastCommit *GitCommit `json:"last_commit,omitempty"`
}

// PRD represents a Product Requirements Document.
type PRD struct {
	ID          string            `json:"id"`
	ProjectID   string            `json:"project_id"`
	Title       string            `json:"title"`
	Status      string            `json:"status"`
	Content     string            `json:"content,omitempty"`
	Progress    string            `json:"progress,omitempty"`
	ChatHistory []json.RawMessage `json:"chat_history,omitempty"`
	SessionID   string            `json:"session_id,omitempty"`
}

// Run represents a single execution run of a PRD story or task.
type Run struct {
	ID         string `json:"id"`
	PRDID      string `json:"prd_id"`
	Status     string `json:"status"`
	Result     string `json:"result,omitempty"`
	Error      string `json:"error,omitempty"`
	StartedAt  string `json:"started_at,omitempty"`
	FinishedAt string `json:"finished_at,omitempty"`
}

// DiffEntry represents a single file diff.
type DiffEntry struct {
	Path   string `json:"path"`
	Status string `json:"status"`
	Patch  string `json:"patch,omitempty"`
}

// FileEntry represents a file or directory in a listing.
type FileEntry struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  *int   `json:"size,omitempty"`
}

// LogEntry represents a single log entry in a log response.
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

// DeviceInfo holds metadata about a connected device.
type DeviceInfo struct {
	DeviceID  string `json:"device_id"`
	Name      string `json:"name,omitempty"`
	Platform  string `json:"platform,omitempty"`
	Version   string `json:"version,omitempty"`
}

// GitCommit represents a git commit reference.
type GitCommit struct {
	SHA     string `json:"sha"`
	Message string `json:"message,omitempty"`
	Author  string `json:"author,omitempty"`
	Date    string `json:"date,omitempty"`
}

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
