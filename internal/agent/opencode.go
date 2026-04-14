package agent

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/minicodemonkey/chief/internal/loop"
)

type OpenCodeProvider struct {
	cliPath string
	model   string
}

func NewOpenCodeProvider(cliPath string) *OpenCodeProvider {
	if cliPath == "" {
		cliPath = "opencode"
	}
	return &OpenCodeProvider{cliPath: cliPath}
}

func (p *OpenCodeProvider) Name() string { return "OpenCode" }

func (p *OpenCodeProvider) CLIPath() string { return p.cliPath }

// SetModel implements loop.Provider.
func (p *OpenCodeProvider) SetModel(model string) { p.model = model }

func (p *OpenCodeProvider) LoopCommand(ctx context.Context, prompt, workDir string) *exec.Cmd {
	args := []string{"run", "--format", "json"}
	if p.model != "" {
		args = append(args, "--model", p.model)
	}
	args = append(args, prompt)
	cmd := exec.CommandContext(ctx, p.cliPath, args...)
	cmd.Dir = workDir
	return cmd
}

func (p *OpenCodeProvider) InteractiveCommand(workDir, prompt string) *exec.Cmd {
	cmd := exec.Command(p.cliPath, "--prompt", prompt)
	cmd.Dir = workDir
	return cmd
}

func (p *OpenCodeProvider) ParseLine(line string) *loop.Event {
	return loop.ParseLineOpenCode(line)
}

func (p *OpenCodeProvider) LogFileName() string { return "opencode.log" }

// CleanOutput extracts JSON from opencode's NDJSON output format.
// It looks for the last "text" event line and returns its part.text content.
func (p *OpenCodeProvider) CleanOutput(output string) string {
	output = strings.TrimSpace(output)
	if !strings.Contains(output, "\n") {
		return output
	}

	var lastText string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var ev struct {
			Type string `json:"type"`
			Part struct {
				Text string `json:"text"`
			} `json:"part"`
		}
		if json.Unmarshal([]byte(line), &ev) == nil && ev.Type == "text" && ev.Part.Text != "" {
			lastText = ev.Part.Text
		}
	}
	if lastText != "" {
		return lastText
	}
	return output
}
