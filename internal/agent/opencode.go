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
}

func NewOpenCodeProvider(cliPath string) *OpenCodeProvider {
	if cliPath == "" {
		cliPath = "opencode"
	}
	return &OpenCodeProvider{cliPath: cliPath}
}

func (p *OpenCodeProvider) Name() string { return "OpenCode" }

func (p *OpenCodeProvider) CLIPath() string { return p.cliPath }

func (p *OpenCodeProvider) LoopCommand(ctx context.Context, prompt, workDir string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, p.cliPath, "run", "--format", "json", prompt)
	cmd.Dir = workDir
	return cmd
}

func (p *OpenCodeProvider) InteractiveCommand(workDir, prompt string) *exec.Cmd {
	cmd := exec.Command(p.cliPath, "--prompt", prompt)
	cmd.Dir = workDir
	return cmd
}

func (p *OpenCodeProvider) ConvertCommand(workDir, prompt string) (*exec.Cmd, loop.OutputMode, string, error) {
	cmd := exec.Command(p.cliPath, "run", "--format", "json", "--", prompt)
	cmd.Dir = workDir
	cmd.Stdin = strings.NewReader(prompt)
	return cmd, loop.OutputStdout, "", nil
}

func (p *OpenCodeProvider) FixJSONCommand(prompt string) (*exec.Cmd, loop.OutputMode, string, error) {
	cmd := exec.Command(p.cliPath, "run", "--format", "json", "--", prompt)
	cmd.Stdin = strings.NewReader(prompt)
	return cmd, loop.OutputStdout, "", nil
}

func (p *OpenCodeProvider) ParseLine(line string) *loop.Event {
	return loop.ParseLineOpenCode(line)
}

func (p *OpenCodeProvider) LogFileName() string { return "opencode.log" }

// CleanOutput extracts JSON from opencode's NDJSON output format.
func (p *OpenCodeProvider) CleanOutput(output string) string {
	output = strings.TrimSpace(output)

	if !strings.Contains(output, "\n") || !strings.Contains(output, `"type":"step_start"`) || !strings.Contains(output, `"type":"text"`) {
		return output
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, `"type":"text"`) {
			var ev struct {
				Type string `json:"type"`
				Part struct {
					Text string `json:"text"`
				} `json:"part"`
			}
			if json.Unmarshal([]byte(line), &ev) == nil && ev.Part.Text != "" {
				return ev.Part.Text
			}
		}
	}
	return output
}
