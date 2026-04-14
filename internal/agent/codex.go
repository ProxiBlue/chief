package agent

import (
	"context"
	"os/exec"
	"strings"

	"github.com/minicodemonkey/chief/internal/loop"
)

// CodexProvider implements loop.Provider for the Codex CLI.
type CodexProvider struct {
	cliPath string
	model   string
}

// NewCodexProvider returns a Provider for the Codex CLI.
// If cliPath is empty, "codex" is used.
func NewCodexProvider(cliPath string) *CodexProvider {
	if cliPath == "" {
		cliPath = "codex"
	}
	return &CodexProvider{cliPath: cliPath}
}

// Name implements loop.Provider.
func (p *CodexProvider) Name() string { return "Codex" }

// CLIPath implements loop.Provider.
func (p *CodexProvider) CLIPath() string { return p.cliPath }

// SetModel implements loop.Provider.
func (p *CodexProvider) SetModel(model string) { p.model = model }

// LoopCommand implements loop.Provider.
func (p *CodexProvider) LoopCommand(ctx context.Context, prompt, workDir string) *exec.Cmd {
	args := []string{"exec", "--json", "--yolo", "--skip-git-repo-check"}
	if p.model != "" {
		args = append(args, "--model", p.model)
	}
	args = append(args, "-C", workDir, "-")
	cmd := exec.CommandContext(ctx, p.cliPath, args...)
	cmd.Dir = workDir
	cmd.Stdin = strings.NewReader(prompt)
	return cmd
}

// InteractiveCommand implements loop.Provider.
func (p *CodexProvider) InteractiveCommand(workDir, prompt string) *exec.Cmd {
	cmd := exec.Command(p.cliPath, prompt)
	cmd.Dir = workDir
	return cmd
}

// ParseLine implements loop.Provider.
func (p *CodexProvider) ParseLine(line string) *loop.Event {
	return loop.ParseLineCodex(line)
}

// LogFileName implements loop.Provider.
func (p *CodexProvider) LogFileName() string { return "codex.log" }

// CleanOutput implements loop.Provider - Codex doesn't use a special format.
func (p *CodexProvider) CleanOutput(output string) string { return output }
