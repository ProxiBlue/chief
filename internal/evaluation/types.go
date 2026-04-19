// Package evaluation provides adversarial evaluation of completed stories.
// After the generator agent commits code, evaluator agents independently
// score acceptance criteria, deliberate, and produce a pass/fail verdict.
package evaluation

import (
	"github.com/minicodemonkey/chief/internal/config"
	"github.com/minicodemonkey/chief/internal/prd"
)

// Config is a convenience alias used within the evaluation package.
type Config = config.EvaluationConfig

// CriterionScore holds the evaluation score for a single acceptance criterion.
type CriterionScore struct {
	Criterion string `json:"criterion"`
	Score     int    `json:"score"`
	Failure   string `json:"failure,omitempty"`
}

// EvaluatorResult holds one evaluator's complete scoring output.
type EvaluatorResult struct {
	EvaluatorID int              `json:"evaluatorId"`
	Scores      []CriterionScore `json:"scores"`
	Pass        bool             `json:"pass"`
}

// DeliberationChallenge represents a challenged finding during deliberation.
type DeliberationChallenge struct {
	Criterion string `json:"criterion"`
	Reason    string `json:"reason"`
}

// DeliberationNewIssue represents a new issue raised during deliberation.
type DeliberationNewIssue struct {
	Criterion string `json:"criterion"`
	Score     int    `json:"score"`
	Failure   string `json:"failure"`
}

// DeliberationResponse is one evaluator's response during the deliberation round.
type DeliberationResponse struct {
	EvaluatorID int                     `json:"evaluatorId"`
	Discussion  string                  `json:"discussion,omitempty"`
	Agree       []int                   `json:"agree"`
	Challenges  []DeliberationChallenge `json:"challenges"`
	NewIssues   []DeliberationNewIssue  `json:"newIssues"`
}

// EvaluationResult holds the full result of an evaluation cycle.
type EvaluationResult struct {
	StoryID      string                 `json:"storyId"`
	Pass         bool                   `json:"pass"`
	Attempt      int                    `json:"attempt"`
	MaxAttempts  int                    `json:"maxAttempts"`
	Evaluators   []EvaluatorResult      `json:"evaluators"`
	Deliberation []DeliberationResponse `json:"deliberation"`
	FinalScores  []CriterionScore       `json:"finalScores"`
	Timestamp    string                 `json:"timestamp"`
}

// StoryContext bundles what evaluators need to assess a story.
type StoryContext struct {
	Story prd.UserStory
	Diff  string
}
