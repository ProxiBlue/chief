package uplink

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// initGitRepo creates a git repo in dir with an initial commit.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
	// Create a file and commit so HEAD exists.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "initial commit")
}

func TestStateCollectorDiscoverProjects(t *testing.T) {
	workspace := t.TempDir()

	// Create two git repos and one non-git dir.
	for _, name := range []string{"alpha", "beta"} {
		dir := filepath.Join(workspace, name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		initGitRepo(t, dir)
	}
	// Non-git directory — should be ignored.
	if err := os.MkdirAll(filepath.Join(workspace, "not-a-repo"), 0755); err != nil {
		t.Fatal(err)
	}

	sc := NewStateCollector(workspace, "1.0.0")
	state, err := sc.Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}

	if len(state.Projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(state.Projects))
	}

	names := map[string]bool{}
	for _, p := range state.Projects {
		names[p.Name] = true
	}
	if !names["alpha"] || !names["beta"] {
		t.Errorf("expected projects alpha and beta, got %v", names)
	}
}

func TestStateCollectorProjectGitInfo(t *testing.T) {
	workspace := t.TempDir()
	dir := filepath.Join(workspace, "myproject")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, dir)

	// Add a remote.
	cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/test/myproject.git")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git remote add failed: %v\n%s", err, out)
	}

	sc := NewStateCollector(workspace, "1.0.0")
	state, err := sc.Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}

	if len(state.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(state.Projects))
	}

	proj := state.Projects[0]

	if proj.Name != "myproject" {
		t.Errorf("name = %q, want %q", proj.Name, "myproject")
	}
	if proj.Path == "" {
		t.Error("path should not be empty")
	}
	if proj.GitBranch == "" {
		t.Error("git branch should not be empty")
	}
	if proj.GitRemote != "https://github.com/test/myproject.git" {
		t.Errorf("git remote = %q, want origin URL", proj.GitRemote)
	}
	if proj.GitSHA == "" {
		t.Error("git SHA should not be empty")
	}
	if proj.GitStatus != "clean" {
		t.Errorf("git status = %q, want %q", proj.GitStatus, "clean")
	}
	if proj.LastCommit == nil {
		t.Fatal("last commit should not be nil")
	}
	if proj.LastCommit.SHA != proj.GitSHA {
		t.Errorf("last commit SHA = %q, want %q", proj.LastCommit.SHA, proj.GitSHA)
	}
	if proj.LastCommit.Message != "initial commit" {
		t.Errorf("last commit message = %q, want %q", proj.LastCommit.Message, "initial commit")
	}
	if proj.LastCommit.Author != "Test" {
		t.Errorf("last commit author = %q, want %q", proj.LastCommit.Author, "Test")
	}
}

func TestStateCollectorDirtyStatus(t *testing.T) {
	workspace := t.TempDir()
	dir := filepath.Join(workspace, "dirty-repo")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, dir)

	// Create an untracked file to make it dirty.
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	sc := NewStateCollector(workspace, "1.0.0")
	state, err := sc.Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}

	if state.Projects[0].GitStatus != "dirty" {
		t.Errorf("git status = %q, want %q", state.Projects[0].GitStatus, "dirty")
	}
}

func TestStateCollectorDeviceInfo(t *testing.T) {
	workspace := t.TempDir()

	sc := NewStateCollector(workspace, "2.5.0")
	state, err := sc.Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}

	if state.Device == nil {
		t.Fatal("device info should not be nil")
	}
	if state.Device.DeviceID == "" {
		t.Error("device ID should not be empty")
	}
	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if state.Device.Platform != expectedPlatform {
		t.Errorf("platform = %q, want %q", state.Device.Platform, expectedPlatform)
	}
	if state.Device.Version != "2.5.0" {
		t.Errorf("version = %q, want %q", state.Device.Version, "2.5.0")
	}
}

func TestStateCollectorPRDs(t *testing.T) {
	workspace := t.TempDir()
	dir := filepath.Join(workspace, "project-with-prds")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, dir)

	// Create PRD structure: .chief/prds/feature-a/prd.md and progress.md
	prdDir := filepath.Join(dir, ".chief", "prds", "feature-a")
	if err := os.MkdirAll(prdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(prdDir, "prd.md"), []byte("# Feature A\nBuild feature A"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(prdDir, "progress.md"), []byte("## Progress\n- Step 1 done"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a second PRD without progress.md
	prdDir2 := filepath.Join(dir, ".chief", "prds", "feature-b")
	if err := os.MkdirAll(prdDir2, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(prdDir2, "prd.md"), []byte("# Feature B"), 0644); err != nil {
		t.Fatal(err)
	}

	sc := NewStateCollector(workspace, "1.0.0")
	state, err := sc.Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}

	if len(state.PRDs) != 2 {
		t.Fatalf("expected 2 PRDs, got %d", len(state.PRDs))
	}

	prdByTitle := map[string]struct {
		content  string
		progress string
	}{}
	for _, prd := range state.PRDs {
		prdByTitle[prd.Title] = struct {
			content  string
			progress string
		}{prd.Content, prd.Progress}
	}

	a, ok := prdByTitle["feature-a"]
	if !ok {
		t.Fatal("feature-a PRD not found")
	}
	if !strings.Contains(a.content, "Feature A") {
		t.Errorf("feature-a content = %q, want to contain 'Feature A'", a.content)
	}
	if !strings.Contains(a.progress, "Step 1 done") {
		t.Errorf("feature-a progress = %q, want to contain 'Step 1 done'", a.progress)
	}

	b, ok := prdByTitle["feature-b"]
	if !ok {
		t.Fatal("feature-b PRD not found")
	}
	if !strings.Contains(b.content, "Feature B") {
		t.Errorf("feature-b content = %q, want to contain 'Feature B'", b.content)
	}
	if b.progress != "" {
		t.Errorf("feature-b progress should be empty, got %q", b.progress)
	}

	// Verify PRDs have correct project ID.
	projID := state.Projects[0].ID
	for _, prd := range state.PRDs {
		if prd.ProjectID != projID {
			t.Errorf("PRD %q project_id = %q, want %q", prd.Title, prd.ProjectID, projID)
		}
	}
}

func TestStateCollectorAddProject(t *testing.T) {
	workspace := t.TempDir()

	// Create a project outside the workspace.
	externalDir := t.TempDir()
	projDir := filepath.Join(externalDir, "external")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, projDir)

	sc := NewStateCollector(workspace, "1.0.0")
	sc.AddProject(projDir)

	state, err := sc.Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}

	if len(state.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(state.Projects))
	}
	if state.Projects[0].Name != "external" {
		t.Errorf("name = %q, want %q", state.Projects[0].Name, "external")
	}
}

func TestStateCollectorAddProjectDeduplication(t *testing.T) {
	workspace := t.TempDir()
	dir := filepath.Join(workspace, "duped")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, dir)

	// Add the same project that will also be discovered.
	sc := NewStateCollector(workspace, "1.0.0")
	sc.AddProject(dir)

	state, err := sc.Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}

	if len(state.Projects) != 1 {
		t.Fatalf("expected 1 project (deduplicated), got %d", len(state.Projects))
	}
}

func TestStateCollectorEmptyWorkspace(t *testing.T) {
	workspace := t.TempDir()

	sc := NewStateCollector(workspace, "1.0.0")
	state, err := sc.Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}

	if len(state.Projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(state.Projects))
	}
	if len(state.PRDs) != 0 {
		t.Errorf("expected 0 PRDs, got %d", len(state.PRDs))
	}
	if len(state.Runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(state.Runs))
	}
}

func TestStateCollectorProjectID(t *testing.T) {
	workspace := t.TempDir()
	dir := filepath.Join(workspace, "stable-id")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, dir)

	// Collect twice — IDs should be deterministic.
	sc1 := NewStateCollector(workspace, "1.0.0")
	state1, _ := sc1.Collect()

	sc2 := NewStateCollector(workspace, "1.0.0")
	state2, _ := sc2.Collect()

	if state1.Projects[0].ID != state2.Projects[0].ID {
		t.Errorf("project IDs not deterministic: %q vs %q", state1.Projects[0].ID, state2.Projects[0].ID)
	}
	if !strings.HasPrefix(state1.Projects[0].ID, "proj_") {
		t.Errorf("project ID should start with proj_, got %q", state1.Projects[0].ID)
	}
}
