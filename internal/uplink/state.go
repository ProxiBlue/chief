package uplink

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/minicodemonkey/chief/internal/protocol"
)

// StateCollector discovers projects in a workspace directory and gathers
// their state into a protocol.StateSync payload.
type StateCollector struct {
	workspaceDir string
	version      string
	projects     []string // manually added project paths
}

// NewStateCollector creates a collector that scans the given workspace directory.
// version is the chief CLI version string.
func NewStateCollector(workspaceDir, version string) *StateCollector {
	return &StateCollector{
		workspaceDir: workspaceDir,
		version:      version,
	}
}

// AddProject adds a project path to be included in state collection,
// in addition to any projects discovered by scanning the workspace directory.
func (sc *StateCollector) AddProject(path string) {
	sc.projects = append(sc.projects, path)
}

// Collect gathers device info, discovers projects, collects git state and PRDs,
// and returns a StateSync payload ready for envelope wrapping.
func (sc *StateCollector) Collect() (protocol.StateSync, error) {
	device, err := sc.collectDeviceInfo()
	if err != nil {
		return protocol.StateSync{}, fmt.Errorf("collect device info: %w", err)
	}

	projectPaths, err := sc.discoverProjects()
	if err != nil {
		return protocol.StateSync{}, fmt.Errorf("discover projects: %w", err)
	}

	var projects []protocol.Project
	var prds []protocol.PRD
	for _, dir := range projectPaths {
		proj := sc.collectProject(dir)
		projects = append(projects, proj)

		foundPRDs := sc.collectPRDs(dir, proj.ID)
		prds = append(prds, foundPRDs...)
	}

	return protocol.StateSync{
		Device:   &device,
		Projects: projects,
		PRDs:     prds,
		Runs:     []protocol.Run{},
	}, nil
}

// collectDeviceInfo gathers hostname, OS, arch, and chief version.
func (sc *StateCollector) collectDeviceInfo() (protocol.DeviceInfo, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return protocol.DeviceInfo{}, fmt.Errorf("get hostname: %w", err)
	}

	platform := runtime.GOOS + "/" + runtime.GOARCH

	return protocol.DeviceInfo{
		DeviceID: hostname,
		Name:     hostname,
		Platform: platform,
		Version:  sc.version,
	}, nil
}

// discoverProjects scans the workspace directory for subdirectories containing
// .git/ and merges them with manually added projects. Returns deduplicated paths.
func (sc *StateCollector) discoverProjects() ([]string, error) {
	seen := make(map[string]bool)
	var paths []string

	// Add manually registered projects first.
	for _, p := range sc.projects {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if !seen[abs] {
			seen[abs] = true
			paths = append(paths, abs)
		}
	}

	// Scan workspace directory for git repos (one level deep).
	entries, err := os.ReadDir(sc.workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("read workspace dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(sc.workspaceDir, entry.Name())
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			abs, err := filepath.Abs(dir)
			if err != nil {
				continue
			}
			if !seen[abs] {
				seen[abs] = true
				paths = append(paths, abs)
			}
		}
	}

	return paths, nil
}

// collectProject gathers project metadata and git state for a directory.
func (sc *StateCollector) collectProject(dir string) protocol.Project {
	name := filepath.Base(dir)
	id := generateProjectID(dir)

	proj := protocol.Project{
		ID:   id,
		Path: dir,
		Name: name,
	}

	// Git branch
	if branch, err := gitCommand(dir, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		proj.GitBranch = branch
	}

	// Git remote
	if remote, err := gitCommand(dir, "config", "--get", "remote.origin.url"); err == nil {
		proj.GitRemote = remote
	}

	// Git SHA (HEAD)
	if sha, err := gitCommand(dir, "rev-parse", "HEAD"); err == nil {
		proj.GitSHA = sha
	}

	// Git status (clean/dirty)
	if status, err := gitCommand(dir, "status", "--porcelain"); err == nil {
		if strings.TrimSpace(status) == "" {
			proj.GitStatus = "clean"
		} else {
			proj.GitStatus = "dirty"
		}
	}

	// Last commit info
	if proj.GitSHA != "" {
		commit := &protocol.GitCommit{SHA: proj.GitSHA}
		if msg, err := gitCommand(dir, "log", "-1", "--format=%s"); err == nil {
			commit.Message = msg
		}
		if author, err := gitCommand(dir, "log", "-1", "--format=%an"); err == nil {
			commit.Author = author
		}
		if date, err := gitCommand(dir, "log", "-1", "--format=%aI"); err == nil {
			commit.Date = date
		}
		proj.LastCommit = commit
	}

	return proj
}

// collectPRDs reads .chief/prds/*/prd.md and optional progress.md for a project.
func (sc *StateCollector) collectPRDs(projectDir, projectID string) []protocol.PRD {
	prdsDir := filepath.Join(projectDir, ".chief", "prds")
	entries, err := os.ReadDir(prdsDir)
	if err != nil {
		return nil // no PRDs directory is fine
	}

	var prds []protocol.PRD
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		prdDir := filepath.Join(prdsDir, entry.Name())
		prdFile := filepath.Join(prdDir, "prd.md")

		content, err := os.ReadFile(prdFile)
		if err != nil {
			continue // skip dirs without prd.md
		}

		prd := protocol.PRD{
			ID:        generatePRDID(projectID, entry.Name()),
			ProjectID: projectID,
			Title:     entry.Name(),
			Status:    "active",
			Content:   string(content),
		}

		// Read optional progress.md
		if progress, err := os.ReadFile(filepath.Join(prdDir, "progress.md")); err == nil {
			prd.Progress = string(progress)
		}

		prds = append(prds, prd)
	}

	return prds
}

// gitCommand runs a git command in the given directory and returns trimmed stdout.
func gitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// generateProjectID creates a deterministic ID from the project path.
func generateProjectID(path string) string {
	h := sha256.Sum256([]byte(path))
	return fmt.Sprintf("proj_%x", h[:8])
}

// generatePRDID creates a deterministic ID from project ID and PRD name.
func generatePRDID(projectID, name string) string {
	h := sha256.Sum256([]byte(projectID + ":" + name))
	return fmt.Sprintf("prd_%x", h[:8])
}
