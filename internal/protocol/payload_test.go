package protocol

import (
	"encoding/json"
	"reflect"
	"testing"
)

// roundTrip marshals v to JSON, then unmarshals into a new T and compares.
func roundTrip[T any](t *testing.T, name string, v T) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("%s: marshal: %v", name, err)
	}
	var got T
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("%s: unmarshal: %v", name, err)
	}
	if !reflect.DeepEqual(v, got) {
		t.Errorf("%s: round-trip mismatch\n  want: %+v\n  got:  %+v", name, v, got)
	}
}

// --- Control payloads ---

func TestWelcomeRoundTrip(t *testing.T) {
	roundTrip(t, "Welcome", Welcome{
		ServerVersion: "1.0.0",
		ConnectionID:  "conn-abc",
	})
}

func TestAckRoundTrip(t *testing.T) {
	roundTrip(t, "Ack", Ack{RefID: "msg_123"})
}

func TestErrorRoundTrip(t *testing.T) {
	roundTrip(t, "Error", Error{
		Code:    400,
		Message: "bad request",
		RefID:   "msg_456",
	})
}

func TestErrorNoRefID(t *testing.T) {
	roundTrip(t, "ErrorNoRef", Error{
		Code:    500,
		Message: "internal",
	})
}

// --- Shared types ---

func TestProjectRoundTrip(t *testing.T) {
	roundTrip(t, "Project", Project{
		ID:        "proj-1",
		Path:      "/home/user/project",
		Name:      "my-project",
		GitRemote: "git@github.com:user/repo.git",
		GitBranch: "main",
		GitSHA:    "abc123",
	})
}

func TestProjectMinimal(t *testing.T) {
	roundTrip(t, "ProjectMinimal", Project{
		ID:   "proj-2",
		Path: "/tmp/p",
		Name: "p",
	})
}

func TestPRDRoundTrip(t *testing.T) {
	roundTrip(t, "PRD", PRD{
		ID:        "prd-1",
		ProjectID: "proj-1",
		Title:     "Feature X",
		Status:    "active",
		Content:   "some content",
		Progress:  "50%",
		SessionID: "sess-1",
	})
}

func TestRunRoundTrip(t *testing.T) {
	roundTrip(t, "Run", Run{
		ID:         "run-1",
		PRDID:      "prd-1",
		Status:     "completed",
		Result:     "success",
		StartedAt:  "2026-01-01T00:00:00Z",
		FinishedAt: "2026-01-01T00:01:00Z",
	})
}

func TestDiffEntryRoundTrip(t *testing.T) {
	roundTrip(t, "DiffEntry", DiffEntry{
		Path:   "src/main.go",
		Status: "modified",
		Patch:  "@@ -1,3 +1,4 @@",
	})
}

func TestFileEntryRoundTrip(t *testing.T) {
	size := 1024
	roundTrip(t, "FileEntry", FileEntry{
		Name:  "readme.md",
		IsDir: false,
		Size:  &size,
	})
}

func TestFileEntryNoSize(t *testing.T) {
	roundTrip(t, "FileEntryNoSize", FileEntry{
		Name:  "src",
		IsDir: true,
	})
}

func TestLogEntryRoundTrip(t *testing.T) {
	roundTrip(t, "LogEntry", LogEntry{
		Timestamp: "2026-01-01T00:00:00Z",
		Level:     "info",
		Message:   "started",
	})
}

func TestDeviceInfoRoundTrip(t *testing.T) {
	roundTrip(t, "DeviceInfo", DeviceInfo{
		DeviceID: "dev-1",
		Name:     "MacBook",
		Platform: "darwin",
		Version:  "1.0.0",
	})
}

func TestGitCommitRoundTrip(t *testing.T) {
	roundTrip(t, "GitCommit", GitCommit{
		SHA:     "abc123def",
		Message: "initial commit",
		Author:  "dev",
		Date:    "2026-01-01T00:00:00Z",
	})
}

// --- State payloads ---

func TestStateSyncRoundTrip(t *testing.T) {
	roundTrip(t, "StateSync", StateSync{
		Projects: []Project{{ID: "p1", Path: "/p", Name: "proj"}},
		PRDs:     []PRD{{ID: "prd1", ProjectID: "p1", Title: "T", Status: "draft"}},
		Runs:     []Run{{ID: "r1", PRDID: "prd1", Status: "pending"}},
		Settings: json.RawMessage(`{"theme":"dark"}`),
	})
}

func TestStateSyncNoSettings(t *testing.T) {
	roundTrip(t, "StateSyncNoSettings", StateSync{
		Projects: []Project{},
		PRDs:     []PRD{},
		Runs:     []Run{},
	})
}

func TestStateProjectsUpdatedRoundTrip(t *testing.T) {
	roundTrip(t, "StateProjectsUpdated", StateProjectsUpdated{
		Projects: []Project{{ID: "p1", Path: "/p", Name: "proj"}},
	})
}

func TestStatePRDCreatedRoundTrip(t *testing.T) {
	roundTrip(t, "StatePRDCreated", StatePRDCreated{
		PRD: PRD{ID: "prd1", ProjectID: "p1", Title: "New", Status: "draft"},
	})
}

func TestStatePRDUpdatedRoundTrip(t *testing.T) {
	roundTrip(t, "StatePRDUpdated", StatePRDUpdated{
		PRD: PRD{ID: "prd1", ProjectID: "p1", Title: "Updated", Status: "active"},
	})
}

func TestStatePRDDeletedRoundTrip(t *testing.T) {
	roundTrip(t, "StatePRDDeleted", StatePRDDeleted{
		PRDID:     "prd1",
		ProjectID: "p1",
	})
}

func TestStatePRDChatOutputRoundTrip(t *testing.T) {
	roundTrip(t, "StatePRDChatOutput", StatePRDChatOutput{
		PRDID:   "prd1",
		Role:    "assistant",
		Content: "Hello!",
	})
}

func TestStateRunStartedRoundTrip(t *testing.T) {
	roundTrip(t, "StateRunStarted", StateRunStarted{
		Run: Run{ID: "r1", PRDID: "prd1", Status: "running"},
	})
}

func TestStateRunProgressRoundTrip(t *testing.T) {
	pct := 42
	roundTrip(t, "StateRunProgress", StateRunProgress{
		RunID:      "r1",
		PRDID:      "prd1",
		Message:    "Building...",
		Percentage: &pct,
	})
}

func TestStateRunProgressNoPercentage(t *testing.T) {
	roundTrip(t, "StateRunProgressNoPct", StateRunProgress{
		RunID:   "r1",
		PRDID:   "prd1",
		Message: "Starting...",
	})
}

func TestStateRunOutputRoundTrip(t *testing.T) {
	roundTrip(t, "StateRunOutput", StateRunOutput{
		RunID:  "r1",
		Stream: "stdout",
		Data:   "hello world\n",
	})
}

func TestStateRunStoppedRoundTrip(t *testing.T) {
	roundTrip(t, "StateRunStopped", StateRunStopped{
		RunID:  "r1",
		PRDID:  "prd1",
		Reason: "user cancelled",
	})
}

func TestStateRunCompletedRoundTrip(t *testing.T) {
	roundTrip(t, "StateRunCompleted", StateRunCompleted{
		RunID:      "r1",
		PRDID:      "prd1",
		Status:     "completed",
		Result:     "all tests passed",
		FinishedAt: "2026-01-01T00:05:00Z",
	})
}

func TestStateRunCompletedFailed(t *testing.T) {
	roundTrip(t, "StateRunCompletedFailed", StateRunCompleted{
		RunID:      "r1",
		PRDID:      "prd1",
		Status:     "failed",
		Result:     "tests failed",
		Error:      "exit code 1",
		FinishedAt: "2026-01-01T00:05:00Z",
	})
}

func TestStateDiffsResponseRoundTrip(t *testing.T) {
	roundTrip(t, "StateDiffsResponse", StateDiffsResponse{
		ProjectID: "p1",
		Ref:       "HEAD~1",
		Diffs: []DiffEntry{
			{Path: "main.go", Status: "modified", Patch: "@@"},
			{Path: "new.go", Status: "added"},
		},
	})
}

func TestStateLogOutputRoundTrip(t *testing.T) {
	roundTrip(t, "StateLogOutput", StateLogOutput{
		Level:   "error",
		Message: "something broke",
		Context: json.RawMessage(`{"file":"main.go"}`),
	})
}

func TestStateLogResponseRoundTrip(t *testing.T) {
	roundTrip(t, "StateLogResponse", StateLogResponse{
		RunID: "r1",
		Entries: []LogEntry{
			{Timestamp: "2026-01-01T00:00:00Z", Level: "info", Message: "start"},
			{Timestamp: "2026-01-01T00:01:00Z", Level: "error", Message: "fail"},
		},
	})
}

func TestStateSettingsUpdatedRoundTrip(t *testing.T) {
	roundTrip(t, "StateSettingsUpdated", StateSettingsUpdated{
		Settings: json.RawMessage(`{"auto_run":true}`),
	})
}

func TestStateDeviceHeartbeatRoundTrip(t *testing.T) {
	roundTrip(t, "StateDeviceHeartbeat", StateDeviceHeartbeat{
		UptimeSeconds: 3600,
		ActiveRuns:    2,
	})
}

func TestStateFilesListRoundTrip(t *testing.T) {
	size := 512
	roundTrip(t, "StateFilesList", StateFilesList{
		ProjectID: "p1",
		Path:      "src",
		Files: []FileEntry{
			{Name: "main.go", IsDir: false, Size: &size},
			{Name: "pkg", IsDir: true},
		},
	})
}

func TestStateFileResponseRoundTrip(t *testing.T) {
	roundTrip(t, "StateFileResponse", StateFileResponse{
		ProjectID: "p1",
		Path:      "main.go",
		Content:   "package main",
		Encoding:  "utf-8",
	})
}

func TestStateProjectCloneProgressRoundTrip(t *testing.T) {
	pct := 75
	roundTrip(t, "StateProjectCloneProgress", StateProjectCloneProgress{
		ProjectID:  "p1",
		Status:     "cloning",
		Message:    "Receiving objects...",
		Percentage: &pct,
	})
}

// --- Command payloads ---

func TestCmdPRDCreateRoundTrip(t *testing.T) {
	roundTrip(t, "CmdPRDCreate", CmdPRDCreate{
		ProjectID: "p1",
		Title:     "New Feature",
		Content:   "Build something",
	})
}

func TestCmdPRDMessageRoundTrip(t *testing.T) {
	roundTrip(t, "CmdPRDMessage", CmdPRDMessage{
		PRDID:   "prd1",
		Message: "What about edge cases?",
	})
}

func TestCmdPRDUpdateRoundTrip(t *testing.T) {
	roundTrip(t, "CmdPRDUpdate", CmdPRDUpdate{
		PRDID:   "prd1",
		Title:   "Updated Title",
		Content: "Updated content",
	})
}

func TestCmdPRDUpdatePartial(t *testing.T) {
	roundTrip(t, "CmdPRDUpdatePartial", CmdPRDUpdate{
		PRDID: "prd1",
		Title: "Only Title",
	})
}

func TestCmdPRDDeleteRoundTrip(t *testing.T) {
	roundTrip(t, "CmdPRDDelete", CmdPRDDelete{PRDID: "prd1"})
}

func TestCmdRunStartRoundTrip(t *testing.T) {
	roundTrip(t, "CmdRunStart", CmdRunStart{
		PRDID: "prd1",
		RunID: "run1",
	})
}

func TestCmdRunStopRoundTrip(t *testing.T) {
	roundTrip(t, "CmdRunStop", CmdRunStop{RunID: "run1"})
}

func TestCmdProjectCloneRoundTrip(t *testing.T) {
	roundTrip(t, "CmdProjectClone", CmdProjectClone{
		GitURL: "git@github.com:user/repo.git",
		Branch: "develop",
	})
}

func TestCmdProjectCloneNoBranch(t *testing.T) {
	roundTrip(t, "CmdProjectCloneNoBranch", CmdProjectClone{
		GitURL: "https://github.com/user/repo.git",
	})
}

func TestCmdDiffsGetRoundTrip(t *testing.T) {
	roundTrip(t, "CmdDiffsGet", CmdDiffsGet{ProjectID: "p1"})
}

func TestCmdLogGetRoundTrip(t *testing.T) {
	limit := 50
	roundTrip(t, "CmdLogGet", CmdLogGet{
		ProjectID: "p1",
		Limit:     &limit,
	})
}

func TestCmdLogGetNoLimit(t *testing.T) {
	roundTrip(t, "CmdLogGetNoLimit", CmdLogGet{ProjectID: "p1"})
}

func TestCmdFilesListRoundTrip(t *testing.T) {
	roundTrip(t, "CmdFilesList", CmdFilesList{
		ProjectID: "p1",
		Path:      "src/internal",
	})
}

func TestCmdFilesListNoPath(t *testing.T) {
	roundTrip(t, "CmdFilesListNoPath", CmdFilesList{ProjectID: "p1"})
}

func TestCmdFileGetRoundTrip(t *testing.T) {
	roundTrip(t, "CmdFileGet", CmdFileGet{
		ProjectID: "p1",
		Path:      "main.go",
	})
}

func TestCmdSettingsGetRoundTrip(t *testing.T) {
	roundTrip(t, "CmdSettingsGet", CmdSettingsGet{})
}

func TestCmdSettingsUpdateRoundTrip(t *testing.T) {
	roundTrip(t, "CmdSettingsUpdate", CmdSettingsUpdate{
		Settings: json.RawMessage(`{"theme":"light"}`),
	})
}

// --- Envelope + DecodePayload integration ---

func TestDecodePayloadStateSync(t *testing.T) {
	payload := StateSync{
		Projects: []Project{{ID: "p1", Path: "/p", Name: "proj"}},
		PRDs:     []PRD{{ID: "prd1", ProjectID: "p1", Title: "T", Status: "draft"}},
		Runs:     []Run{{ID: "r1", PRDID: "prd1", Status: "pending"}},
	}
	data, _ := json.Marshal(payload)
	env := Envelope{
		Type:     TypeSync,
		ID:       "msg_test",
		DeviceID: "dev-1",
		Payload:  data,
	}
	got, err := DecodePayload[StateSync](env)
	if err != nil {
		t.Fatalf("DecodePayload: %v", err)
	}
	if len(got.Projects) != 1 || got.Projects[0].ID != "p1" {
		t.Errorf("unexpected projects: %+v", got.Projects)
	}
}

func TestDecodePayloadCmdRunStart(t *testing.T) {
	payload := CmdRunStart{PRDID: "prd1", RunID: "run1"}
	data, _ := json.Marshal(payload)
	env := Envelope{
		Type:     TypeRunStart,
		ID:       "msg_test",
		DeviceID: "dev-1",
		Payload:  data,
	}
	got, err := DecodePayload[CmdRunStart](env)
	if err != nil {
		t.Fatalf("DecodePayload: %v", err)
	}
	if got.PRDID != "prd1" || got.RunID != "run1" {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestDecodePayloadWelcome(t *testing.T) {
	payload := Welcome{ServerVersion: "2.0.0", ConnectionID: "conn-1"}
	data, _ := json.Marshal(payload)
	env := Envelope{
		Type:     TypeWelcome,
		ID:       "msg_test",
		DeviceID: "dev-1",
		Payload:  data,
	}
	got, err := DecodePayload[Welcome](env)
	if err != nil {
		t.Fatalf("DecodePayload: %v", err)
	}
	if got.ServerVersion != "2.0.0" {
		t.Errorf("ServerVersion = %q, want %q", got.ServerVersion, "2.0.0")
	}
}
