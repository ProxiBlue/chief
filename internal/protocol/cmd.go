package protocol

import "encoding/json"

// CmdPRDCreate is the payload for a "prd-create" command message.
type CmdPRDCreate struct {
	ProjectID string `json:"project_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
}

// CmdPRDMessage is the payload for a "prd-message" command message.
type CmdPRDMessage struct {
	PRDID   string `json:"prd_id"`
	Message string `json:"message"`
}

// CmdPRDUpdate is the payload for a "prd-update" command message.
type CmdPRDUpdate struct {
	PRDID   string `json:"prd_id"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

// CmdPRDDelete is the payload for a "prd-delete" command message.
type CmdPRDDelete struct {
	PRDID string `json:"prd_id"`
}

// CmdRunStart is the payload for a "run-start" command message.
type CmdRunStart struct {
	PRDID string `json:"prd_id"`
	RunID string `json:"run_id"`
}

// CmdRunStop is the payload for a "run-stop" command message.
type CmdRunStop struct {
	RunID string `json:"run_id"`
}

// CmdProjectClone is the payload for a "project-clone" command message.
type CmdProjectClone struct {
	GitURL string `json:"git_url"`
	Branch string `json:"branch,omitempty"`
}

// CmdDiffsGet is the payload for a "diffs-get" command message.
type CmdDiffsGet struct {
	ProjectID string `json:"project_id"`
}

// CmdLogGet is the payload for a "log-get" command message.
type CmdLogGet struct {
	ProjectID string `json:"project_id"`
	Limit     *int   `json:"limit,omitempty"`
}

// CmdFilesList is the payload for a "files-list" command message.
type CmdFilesList struct {
	ProjectID string `json:"project_id"`
	Path      string `json:"path,omitempty"`
}

// CmdFileGet is the payload for a "file-get" command message.
type CmdFileGet struct {
	ProjectID string `json:"project_id"`
	Path      string `json:"path"`
}

// CmdSettingsGet is the payload for a "settings-get" command message.
type CmdSettingsGet struct{}

// CmdSettingsUpdate is the payload for a "settings-update" command message.
type CmdSettingsUpdate struct {
	Settings json.RawMessage `json:"settings"`
}
