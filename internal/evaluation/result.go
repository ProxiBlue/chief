package evaluation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EvalDir returns the evaluation directory for a story.
// Path: <baseDir>/.chief/evaluations/{prdName}/{storyID}/
// It locates the project root by searching upward for a .chief directory,
// rather than assuming a fixed number of parent levels.
func EvalDir(prdPath, prdName, storyID string) string {
	// Sanitize inputs to prevent path traversal
	prdName = filepath.Base(prdName)
	storyID = filepath.Base(storyID)

	baseDir := findBaseDir(prdPath)
	return filepath.Join(baseDir, ".chief", "evaluations", prdName, storyID)
}

// findBaseDir walks up from prdPath to find the directory containing .chief/.
func findBaseDir(prdPath string) string {
	dir := filepath.Dir(prdPath)
	for i := 0; i < 10; i++ { // cap iterations to avoid infinite walk
		if _, err := os.Stat(filepath.Join(dir, ".chief")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root
		}
		dir = parent
	}
	// Fallback: 3 levels up (original behavior)
	return filepath.Dir(filepath.Dir(filepath.Dir(prdPath)))
}

// SaveResult writes the evaluation result as JSON.
func SaveResult(prdPath, prdName, storyID string, result *EvaluationResult) error {
	dir := EvalDir(prdPath, prdName, storyID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "result.json"), data, 0o644)
}

// LoadResult reads a saved evaluation result.
func LoadResult(prdPath, prdName, storyID string) (*EvaluationResult, error) {
	dir := EvalDir(prdPath, prdName, storyID)
	data, err := os.ReadFile(filepath.Join(dir, "result.json"))
	if err != nil {
		return nil, err
	}

	var result EvaluationResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// WriteFeedback writes a human/agent-readable feedback file for a failed evaluation.
// passThreshold is used to determine which scores count as failures.
func WriteFeedback(prdPath, prdName, storyID string, result *EvaluationResult, passThreshold int) error {
	dir := EvalDir(prdPath, prdName, storyID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	if passThreshold <= 0 {
		passThreshold = 7
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("## Evaluation Failed (Attempt %d/%d)\n\n", result.Attempt, result.MaxAttempts))
	b.WriteString("### Failures:\n\n")

	for _, s := range result.FinalScores {
		if s.Score < passThreshold {
			b.WriteString(fmt.Sprintf("- criterion: %q\n  score: %d/10\n", s.Criterion, s.Score))
			if s.Failure != "" {
				b.WriteString(fmt.Sprintf("  issue: %s\n", sanitizeFeedback(s.Failure)))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("### Action Required:\n\nFix the above failures and ensure all acceptance criteria pass.\n")

	return os.WriteFile(filepath.Join(dir, "feedback.md"), []byte(b.String()), 0o644)
}

// ReadFeedback reads the feedback file for a story, if it exists.
func ReadFeedback(prdPath, prdName, storyID string) (string, error) {
	dir := EvalDir(prdPath, prdName, storyID)
	data, err := os.ReadFile(filepath.Join(dir, "feedback.md"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// CleanFeedback removes the feedback file for a story (e.g., after max retries).
func CleanFeedback(prdPath, prdName, storyID string) {
	dir := EvalDir(prdPath, prdName, storyID)
	os.Remove(filepath.Join(dir, "feedback.md"))
}

// sanitizeFeedback strips content that could be used for prompt injection
// when feedback is embedded into the generator's next prompt.
func sanitizeFeedback(s string) string {
	// Remove chief-done markers that could trick the generator into stopping early
	s = strings.ReplaceAll(s, "<chief-done/>", "[chief-done]")
	s = strings.ReplaceAll(s, "<chief-done>", "[chief-done]")
	// Remove XML/HTML tags that could break prompt structure (e.g. </story>, </diff>)
	s = stripXMLTags(s)
	// Remove markdown heading markers that could break prompt structure
	s = strings.ReplaceAll(s, "\n#", "\n ")
	// Collapse multiple newlines to prevent section injection
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	return s
}

// stripXMLTags removes XML/HTML-style tags from text to prevent prompt structure breakout.
func stripXMLTags(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] == '<' {
			// Look for closing >
			j := i + 1
			for j < len(s) && s[j] != '>' && s[j] != '\n' {
				j++
			}
			if j < len(s) && s[j] == '>' {
				// Replace tag with bracket notation: <tag> -> [tag]
				b.WriteByte('[')
				b.WriteString(s[i+1 : j])
				b.WriteByte(']')
				i = j + 1
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}
