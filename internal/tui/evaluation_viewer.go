package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/minicodemonkey/chief/internal/evaluation"
)

// EvaluationViewer displays an evaluation transcript for a completed story.
type EvaluationViewer struct {
	result    *evaluation.EvaluationResult
	scrollPos int
	width     int
	height    int
	lines     []string // Pre-rendered lines
}

// NewEvaluationViewer creates a new evaluation viewer.
func NewEvaluationViewer() *EvaluationViewer {
	return &EvaluationViewer{}
}

// SetResult loads an evaluation result for display.
func (v *EvaluationViewer) SetResult(result *evaluation.EvaluationResult) {
	v.result = result
	v.scrollPos = 0
	v.lines = v.render()
}

// SetSize updates the viewport dimensions.
func (v *EvaluationViewer) SetSize(width, height int) {
	v.width = width
	v.height = height
	if v.result != nil {
		v.lines = v.render()
	}
}

// ScrollUp scrolls up one line.
func (v *EvaluationViewer) ScrollUp() {
	if v.scrollPos > 0 {
		v.scrollPos--
	}
}

// ScrollDown scrolls down one line.
func (v *EvaluationViewer) ScrollDown() {
	maxScroll := len(v.lines) - v.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if v.scrollPos < maxScroll {
		v.scrollPos++
	}
}

// PageUp scrolls up by half a page.
func (v *EvaluationViewer) PageUp() {
	v.scrollPos -= v.height / 2
	if v.scrollPos < 0 {
		v.scrollPos = 0
	}
}

// PageDown scrolls down by half a page.
func (v *EvaluationViewer) PageDown() {
	maxScroll := len(v.lines) - v.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	v.scrollPos += v.height / 2
	if v.scrollPos > maxScroll {
		v.scrollPos = maxScroll
	}
}

// View returns the rendered evaluation transcript.
func (v *EvaluationViewer) View() string {
	if v.result == nil {
		return "No evaluation data"
	}

	if len(v.lines) == 0 {
		return "Empty evaluation"
	}

	// Window into the lines
	end := v.scrollPos + v.height
	if end > len(v.lines) {
		end = len(v.lines)
	}
	start := v.scrollPos
	if start > len(v.lines) {
		start = len(v.lines)
	}

	visible := v.lines[start:end]
	return strings.Join(visible, "\n")
}

// render pre-renders all lines for the evaluation transcript.
func (v *EvaluationViewer) render() []string {
	if v.result == nil {
		return nil
	}

	r := v.result
	var lines []string

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(PrimaryColor)
	verdictStyle := lipgloss.NewStyle().Bold(true)
	if r.Pass {
		verdictStyle = verdictStyle.Foreground(SuccessColor)
	} else {
		verdictStyle = verdictStyle.Foreground(ErrorColor)
	}

	verdict := "PASS"
	if !r.Pass {
		verdict = "FAIL"
	}

	lines = append(lines, headerStyle.Render(fmt.Sprintf("Evaluation: %s", r.StoryID)))
	lines = append(lines, verdictStyle.Render(fmt.Sprintf("Verdict: %s", verdict)))
	if r.Attempt > 0 {
		lines = append(lines, fmt.Sprintf("Attempt: %d/%d", r.Attempt, r.MaxAttempts))
	}
	lines = append(lines, fmt.Sprintf("Timestamp: %s", r.Timestamp))
	lines = append(lines, "")

	// Final Scores
	lines = append(lines, headerStyle.Render("Final Scores"))
	lines = append(lines, strings.Repeat("─", min(v.width-4, 60)))
	for _, s := range r.FinalScores {
		scoreStyle := lipgloss.NewStyle().Bold(true)
		if s.Score >= 7 {
			scoreStyle = scoreStyle.Foreground(SuccessColor)
		} else {
			scoreStyle = scoreStyle.Foreground(ErrorColor)
		}
		lines = append(lines, fmt.Sprintf("  %s  %s", scoreStyle.Render(fmt.Sprintf("%2d/10", s.Score)), s.Criterion))
		if s.Failure != "" {
			failStyle := lipgloss.NewStyle().Foreground(ErrorColor)
			lines = append(lines, failStyle.Render(fmt.Sprintf("         %s", s.Failure)))
		}
	}
	lines = append(lines, "")

	// Per-evaluator detail
	for _, ev := range r.Evaluators {
		lines = append(lines, headerStyle.Render(fmt.Sprintf("Evaluator %d", ev.EvaluatorID)))
		passLabel := "PASS"
		if !ev.Pass {
			passLabel = "FAIL"
		}
		evStyle := lipgloss.NewStyle().Bold(true)
		if ev.Pass {
			evStyle = evStyle.Foreground(SuccessColor)
		} else {
			evStyle = evStyle.Foreground(ErrorColor)
		}
		lines = append(lines, evStyle.Render(fmt.Sprintf("  Result: %s", passLabel)))
		for _, s := range ev.Scores {
			scoreStyle := lipgloss.NewStyle()
			if s.Score >= 7 {
				scoreStyle = scoreStyle.Foreground(SuccessColor)
			} else {
				scoreStyle = scoreStyle.Foreground(ErrorColor)
			}
			lines = append(lines, fmt.Sprintf("  %s  %s", scoreStyle.Render(fmt.Sprintf("%2d", s.Score)), s.Criterion))
			if s.Failure != "" {
				lines = append(lines, lipgloss.NewStyle().Foreground(MutedColor).Render(fmt.Sprintf("         %s", s.Failure)))
			}
		}
		lines = append(lines, "")
	}

	// Deliberation
	if len(r.Deliberation) > 0 {
		lines = append(lines, headerStyle.Render("Deliberation"))
		lines = append(lines, strings.Repeat("─", min(v.width-4, 60)))
		discussionStyle := lipgloss.NewStyle().Foreground(TextColor)
		for _, d := range r.Deliberation {
			lines = append(lines, headerStyle.Render(fmt.Sprintf("  Evaluator %d:", d.EvaluatorID)))
			// Show model discussion/reasoning if present
			if d.Discussion != "" {
				lines = append(lines, "")
				for _, dLine := range strings.Split(d.Discussion, "\n") {
					lines = append(lines, discussionStyle.Render("    "+dLine))
				}
				lines = append(lines, "")
			}
			if len(d.Agree) > 0 {
				lines = append(lines, fmt.Sprintf("    Agrees with: %v", d.Agree))
			}
			for _, c := range d.Challenges {
				lines = append(lines, lipgloss.NewStyle().Foreground(WarningColor).Render(
					fmt.Sprintf("    Challenge: %s - %s", c.Criterion, c.Reason)))
			}
			for _, n := range d.NewIssues {
				lines = append(lines, lipgloss.NewStyle().Foreground(ErrorColor).Render(
					fmt.Sprintf("    New issue: %s [%d/10] %s", n.Criterion, n.Score, n.Failure)))
			}
			lines = append(lines, "")
		}
	}

	return lines
}
