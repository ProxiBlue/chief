package evaluation

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/minicodemonkey/chief/internal/prd"
)

// maxDiffSize limits the diff embedded in evaluator prompts to avoid token overflow.
const maxDiffSize = 200_000 // ~200KB

// ProgressFunc is a callback for reporting evaluation progress.
type ProgressFunc func(storyID, message string)

// RunEvaluation orchestrates the full evaluation cycle for a completed story:
// 1. Gather story context + diff
// 2. Run N evaluators in parallel
// 3. Run deliberation round
// 4. Compute final verdict
// 5. Persist results
func RunEvaluation(ctx context.Context, provider AgentProvider, workDir, prdPath, storyID string, cfg *Config, progress ProgressFunc) (*EvaluationResult, error) {
	progress(storyID, "Starting adversarial evaluation")

	// Load story
	p, err := prd.LoadPRD(prdPath)
	if err != nil {
		return nil, fmt.Errorf("load PRD: %w", err)
	}

	var story *prd.UserStory
	for i := range p.UserStories {
		if p.UserStories[i].ID == storyID {
			story = &p.UserStories[i]
			break
		}
	}
	if story == nil {
		return nil, fmt.Errorf("story %s not found", storyID)
	}

	// Get git diff for the story's commit
	diff, err := getStoryDiff(ctx, workDir, storyID, story.Title)
	if err != nil {
		diff = "(diff unavailable: " + err.Error() + ")"
	}

	// Truncate huge diffs to avoid token overflow
	if len(diff) > maxDiffSize {
		diff = diff[:maxDiffSize] + "\n\n[diff truncated at 200KB]"
	}

	storyCtx := StoryContext{
		Story: *story,
		Diff:  diff,
	}

	numAgents := cfg.Agents
	if numAgents <= 0 {
		numAgents = 3
	}

	// Phase 1: Run evaluators in parallel
	progress(storyID, fmt.Sprintf("Running %d evaluators in parallel", numAgents))

	results := make([]*EvaluatorResult, numAgents)
	evalErrs := make([]error, numAgents)
	var wg sync.WaitGroup

	for i := 0; i < numAgents; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			result, err := RunEvaluator(ctx, id+1, provider, workDir, storyCtx, cfg)
			if err != nil {
				evalErrs[id] = err
				return
			}
			results[id] = result
			progress(storyID, fmt.Sprintf("Evaluator %d done (pass=%v)", id+1, result.Pass))
		}(i)
	}
	wg.Wait()

	// Collect valid results
	var validResults []*EvaluatorResult
	for i, r := range results {
		if r != nil {
			validResults = append(validResults, r)
		} else if evalErrs[i] != nil {
			progress(storyID, fmt.Sprintf("Evaluator %d failed: %s", i+1, evalErrs[i]))
		}
	}

	if len(validResults) == 0 {
		return nil, fmt.Errorf("all evaluators failed")
	}

	// Phase 2: Deliberation round
	progress(storyID, "Starting deliberation round")

	deliberation, err := RunDeliberation(ctx, provider, workDir, storyCtx, validResults, cfg)
	if err != nil {
		progress(storyID, fmt.Sprintf("Deliberation failed: %s, using initial scores", err))
		deliberation = nil
	}

	// Phase 3: Compute final scores and verdict
	finalScores := computeFinalScores(validResults, deliberation, cfg)
	pass := verdictPass(finalScores, cfg)

	result := &EvaluationResult{
		StoryID:      storyID,
		Pass:         pass,
		Evaluators:   derefResults(validResults),
		Deliberation: deliberation,
		FinalScores:  finalScores,
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	// Persist result
	prdName := prd.ExtractPRDName(prdPath)
	if err := SaveResult(prdPath, prdName, storyID, result); err != nil {
		progress(storyID, fmt.Sprintf("Warning: failed to save result: %s", err))
	}

	return result, nil
}

// getStoryDiff gets the git diff for the most recent commit matching the story.
// All git commands respect the provided context for cancellation.
func getStoryDiff(ctx context.Context, workDir, storyID, storyTitle string) (string, error) {
	// Use --fixed-strings to avoid regex interpretation of story IDs
	commitPattern := fmt.Sprintf("feat: %s - %s", storyID, storyTitle)
	cmd := exec.CommandContext(ctx, "git", "log", "--fixed-strings", fmt.Sprintf("--grep=%s", commitPattern), "-1", "--format=%H")
	cmd.Dir = workDir
	out, err := cmd.Output()

	// Fallback: try just the story ID if full pattern didn't match
	if err != nil || strings.TrimSpace(string(out)) == "" {
		commitPattern = fmt.Sprintf("feat: %s", storyID)
		cmd = exec.CommandContext(ctx, "git", "log", "--fixed-strings", fmt.Sprintf("--grep=%s", commitPattern), "-1", "--format=%H")
		cmd.Dir = workDir
		out, err = cmd.Output()
	}

	if err != nil || strings.TrimSpace(string(out)) == "" {
		// Fallback: get diff of last commit, handling initial commit
		return getLastCommitDiff(ctx, workDir)
	}

	hash := strings.TrimSpace(string(out))
	cmd = exec.CommandContext(ctx, "git", "diff", hash+"~1.."+hash)
	cmd.Dir = workDir
	out, err = cmd.Output()
	if err != nil {
		// hash~1 might not exist (initial commit)
		return getDiffAgainstEmpty(ctx, workDir, hash)
	}
	return string(out), nil
}

// getLastCommitDiff returns the diff for HEAD, handling the initial commit case.
func getLastCommitDiff(ctx context.Context, workDir string) (string, error) {
	// Check if HEAD~1 exists
	check := exec.CommandContext(ctx, "git", "rev-parse", "--verify", "HEAD~1")
	check.Dir = workDir
	if err := check.Run(); err != nil {
		// Initial commit — diff against empty tree
		return getDiffAgainstEmpty(ctx, workDir, "HEAD")
	}

	cmd := exec.CommandContext(ctx, "git", "diff", "HEAD~1..HEAD")
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// getDiffAgainstEmpty shows all changes in a commit that has no parent.
func getDiffAgainstEmpty(ctx context.Context, workDir, ref string) (string, error) {
	// 4b825dc is git's empty tree hash
	cmd := exec.CommandContext(ctx, "git", "diff", "4b825dc642cb6eb9a060e54bf8d69288fbee4904", ref)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// computeFinalScores merges evaluator scores with deliberation adjustments.
func computeFinalScores(results []*EvaluatorResult, deliberation []DeliberationResponse, cfg *Config) []CriterionScore {
	criterionScores := make(map[string][]CriterionScore)
	for _, r := range results {
		for _, s := range r.Scores {
			criterionScores[s.Criterion] = append(criterionScores[s.Criterion], s)
		}
	}

	// Add new issues from deliberation
	if deliberation != nil {
		for _, d := range deliberation {
			for _, issue := range d.NewIssues {
				criterionScores[issue.Criterion] = append(criterionScores[issue.Criterion], CriterionScore{
					Criterion: issue.Criterion,
					Score:     issue.Score,
					Failure:   issue.Failure,
				})
			}
		}
	}

	var final []CriterionScore
	for criterion, scores := range criterionScores {
		if len(scores) == 0 {
			continue
		}

		total := 0
		var worstFailure string
		worstScore := 10
		for _, s := range scores {
			total += s.Score
			if s.Score < worstScore {
				worstScore = s.Score
				worstFailure = s.Failure
			}
		}
		avg := total / len(scores)

		final = append(final, CriterionScore{
			Criterion: criterion,
			Score:     avg,
			Failure:   worstFailure,
		})
	}

	return final
}

// verdictPass returns true if all final scores meet the threshold.
func verdictPass(finalScores []CriterionScore, cfg *Config) bool {
	threshold := cfg.PassThreshold
	if threshold <= 0 {
		threshold = 7
	}
	for _, s := range finalScores {
		if s.Score < threshold {
			return false
		}
	}
	return true
}

// derefResults converts []*EvaluatorResult to []EvaluatorResult for serialization.
func derefResults(ptrs []*EvaluatorResult) []EvaluatorResult {
	out := make([]EvaluatorResult, len(ptrs))
	for i, p := range ptrs {
		out[i] = *p
	}
	return out
}
