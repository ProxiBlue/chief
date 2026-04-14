package evaluation

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"strings"
)

// AgentProvider is the minimal interface evaluation needs from an agent CLI.
type AgentProvider interface {
	Name() string
	LoopCommand(ctx context.Context, prompt, workDir string) *exec.Cmd
	ParseLine(line string) *AgentEvent
}

// AgentEvent is a simplified event type for evaluation output parsing.
type AgentEvent struct {
	IsText bool
	Text   string
}

// evaluatorOutput is the expected JSON structure from an evaluator agent.
// Uses interface{} for Score to handle float/string coercion from LLM output.
type evaluatorOutput struct {
	Scores []evaluatorScoreRaw `json:"scores"`
}

type evaluatorScoreRaw struct {
	Criterion string      `json:"criterion"`
	Score     interface{} `json:"score"`
	Failure   string      `json:"failure,omitempty"`
}

// RunEvaluator invokes a single evaluator agent and returns its scored results.
func RunEvaluator(ctx context.Context, id int, provider AgentProvider, workDir string, story StoryContext, cfg *Config) (*EvaluatorResult, error) {
	prompt := BuildEvaluatorPrompt(id, story)

	cmd := provider.LoopCommand(ctx, prompt, workDir)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start evaluator %d: %w", id, err)
	}

	// Collect all output lines to extract JSON
	var lines []string
	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("evaluator %d scanner: %w", id, err)
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("evaluator %d exited: %w", id, err)
	}

	// Parse the output to find JSON scores
	scores, err := parseEvaluatorOutput(lines, provider)
	if err != nil {
		return nil, fmt.Errorf("parse evaluator %d output: %w", id, err)
	}

	// Determine pass/fail using configured threshold
	threshold := cfg.PassThreshold
	if threshold <= 0 {
		threshold = 7
	}
	pass := true
	for _, s := range scores {
		if s.Score < threshold {
			pass = false
			break
		}
	}

	return &EvaluatorResult{
		EvaluatorID: id,
		Scores:      scores,
		Pass:        pass,
	}, nil
}

// parseEvaluatorOutput extracts criterion scores from agent output.
func parseEvaluatorOutput(lines []string, provider AgentProvider) ([]CriterionScore, error) {
	// First try: parse through the provider's parser to extract assistant text
	var allText strings.Builder
	for _, line := range lines {
		event := provider.ParseLine(line)
		if event != nil && event.IsText {
			allText.WriteString(event.Text)
			allText.WriteString("\n")
		}
	}

	// Try to find JSON in the extracted text
	if scores, err := extractScoresJSON(allText.String()); err == nil {
		return scores, nil
	}

	// Fallback: try to find JSON directly in raw lines
	for _, line := range lines {
		if scores, err := extractScoresJSON(line); err == nil {
			return scores, nil
		}
	}

	// Provide helpful error with truncated output
	preview := allText.String()
	if len(preview) > 500 {
		preview = preview[:500] + "...(truncated)"
	}
	return nil, fmt.Errorf("no valid scores JSON found in evaluator output. Preview:\n%s", preview)
}

// extractScoresJSON tries to extract a scores JSON object from text.
// Strips markdown code fences and uses string-aware brace matching.
func extractScoresJSON(text string) ([]CriterionScore, error) {
	text = stripMarkdownFences(text)

	start := strings.Index(text, "{")
	if start < 0 {
		return nil, fmt.Errorf("no JSON object found")
	}

	end := findMatchingBrace(text, start)
	if end < 0 {
		return nil, fmt.Errorf("no complete JSON object found")
	}

	var output evaluatorOutput
	if err := json.Unmarshal([]byte(text[start:end]), &output); err != nil {
		return nil, err
	}

	if len(output.Scores) == 0 {
		return nil, fmt.Errorf("empty scores array")
	}

	// Coerce raw scores to typed CriterionScore
	scores := make([]CriterionScore, len(output.Scores))
	for i, raw := range output.Scores {
		score, err := coerceScore(raw.Score)
		if err != nil {
			return nil, fmt.Errorf("score %d (%q): %w", i, raw.Criterion, err)
		}
		scores[i] = CriterionScore{
			Criterion: raw.Criterion,
			Score:     score,
			Failure:   raw.Failure,
		}
	}

	return scores, nil
}

// coerceScore converts an interface{} score value to int, clamped to 1-10.
// Handles float64 (JSON default), string numbers, and int.
func coerceScore(v interface{}) (int, error) {
	var raw int
	switch val := v.(type) {
	case float64:
		raw = int(math.Round(val))
	case int:
		raw = val
	case string:
		var f float64
		if _, err := fmt.Sscanf(val, "%f", &f); err != nil {
			return 0, fmt.Errorf("cannot parse score %q as number", val)
		}
		raw = int(math.Round(f))
	case nil:
		return 0, fmt.Errorf("score is null")
	default:
		return 0, fmt.Errorf("unexpected score type %T", v)
	}
	// Clamp to valid range
	if raw < 1 {
		raw = 1
	}
	if raw > 10 {
		raw = 10
	}
	return raw, nil
}

// stripMarkdownFences removes ```json ... ``` wrappers from text.
func stripMarkdownFences(text string) string {
	text = strings.TrimSpace(text)
	// Remove opening fence
	if idx := strings.Index(text, "```json"); idx >= 0 {
		text = text[idx+7:]
	} else if idx := strings.Index(text, "```"); idx >= 0 {
		text = text[idx+3:]
	}
	// Remove closing fence
	if idx := strings.LastIndex(text, "```"); idx >= 0 {
		text = text[:idx]
	}
	return strings.TrimSpace(text)
}

// findMatchingBrace finds the closing } for the { at position start.
// Properly handles JSON strings (ignores braces inside quotes).
func findMatchingBrace(text string, start int) int {
	depth := 0
	inString := false
	escaped := false

	for i := start; i < len(text); i++ {
		ch := text[i]

		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && inString {
			escaped = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}

		switch ch {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}

	return -1
}
