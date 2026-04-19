package evaluation

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// RunDeliberation runs a deliberation round where each evaluator reviews the others' findings.
// All evaluators run in parallel.
func RunDeliberation(ctx context.Context, provider AgentProvider, workDir string, story StoryContext, results []*EvaluatorResult, cfg *Config) ([]DeliberationResponse, error) {
	responses := make([]DeliberationResponse, len(results))
	errs := make([]error, len(results))

	var wg sync.WaitGroup
	for i, result := range results {
		wg.Add(1)
		go func(idx int, own *EvaluatorResult) {
			defer wg.Done()

			var others []*EvaluatorResult
			for j, r := range results {
				if j != idx {
					others = append(others, r)
				}
			}

			resp, err := runDeliberator(ctx, provider, workDir, story, own, others, cfg)
			if err != nil {
				errs[idx] = err
				return
			}
			responses[idx] = *resp
		}(i, result)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("deliberation evaluator %d: %w", i+1, err)
		}
	}

	return responses, nil
}

// runDeliberator runs a single evaluator's deliberation round.
func runDeliberator(ctx context.Context, provider AgentProvider, workDir string, story StoryContext, own *EvaluatorResult, others []*EvaluatorResult, cfg *Config) (*DeliberationResponse, error) {
	prompt := BuildDeliberationPrompt(own.EvaluatorID, story, own, others)

	cmd := provider.LoopCommand(ctx, prompt, workDir)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start deliberation: %w", err)
	}

	var lines []string
	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("deliberation scanner: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("deliberation exited: %w", err)
	}

	return parseDeliberationOutput(lines, own.EvaluatorID, provider)
}

// parseDeliberationOutput extracts a DeliberationResponse from agent output.
// It captures all text before the JSON as the model's discussion/reasoning.
func parseDeliberationOutput(lines []string, evaluatorID int, provider AgentProvider) (*DeliberationResponse, error) {
	var allText strings.Builder
	for _, line := range lines {
		event := provider.ParseLine(line)
		if event != nil && event.IsText {
			allText.WriteString(event.Text)
			allText.WriteString("\n")
		}
	}

	fullText := allText.String()
	resp, err := extractDeliberationJSON(fullText, evaluatorID)
	if err != nil {
		for _, line := range lines {
			if resp, err := extractDeliberationJSON(line, evaluatorID); err == nil {
				return resp, nil
			}
		}
		preview := fullText
		if len(preview) > 500 {
			preview = preview[:500] + "...(truncated)"
		}
		return nil, fmt.Errorf("no valid deliberation JSON found. Preview:\n%s", preview)
	}

	// Extract discussion text: everything before the JSON object
	resp.Discussion = extractDiscussion(fullText)
	return resp, nil
}

// extractDiscussion extracts the reasoning text before the JSON in deliberation output.
func extractDiscussion(text string) string {
	stripped := stripMarkdownFences(text)
	jsonStart := strings.Index(stripped, "{")
	if jsonStart <= 0 {
		return ""
	}
	discussion := strings.TrimSpace(stripped[:jsonStart])
	// Remove any trailing markdown fence artifacts
	discussion = strings.TrimSuffix(discussion, "```")
	discussion = strings.TrimSpace(discussion)
	return discussion
}

// extractDeliberationJSON tries to extract a deliberation JSON from text.
// Uses string-aware brace matching and strips markdown fences.
func extractDeliberationJSON(text string, evaluatorID int) (*DeliberationResponse, error) {
	text = stripMarkdownFences(text)

	start := strings.Index(text, "{")
	if start < 0 {
		return nil, fmt.Errorf("no JSON found")
	}

	end := findMatchingBrace(text, start)
	if end < 0 {
		return nil, fmt.Errorf("incomplete JSON")
	}

	var resp DeliberationResponse
	if err := json.Unmarshal([]byte(text[start:end]), &resp); err != nil {
		return nil, err
	}
	resp.EvaluatorID = evaluatorID
	return &resp, nil
}
