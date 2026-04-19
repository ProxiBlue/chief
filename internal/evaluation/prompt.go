package evaluation

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/minicodemonkey/chief/embed"
)

// BuildEvaluatorPrompt constructs the prompt for a single evaluator.
func BuildEvaluatorPrompt(evaluatorID int, story StoryContext) string {
	storyJSON, err := json.MarshalIndent(story.Story, "", "  ")
	if err != nil {
		storyJSON = []byte(fmt.Sprintf(`{"id": %q, "title": %q, "error": "marshal failed"}`, story.Story.ID, story.Story.Title))
	}
	return embed.GetEvaluatorPrompt(
		fmt.Sprintf("%d", evaluatorID),
		string(storyJSON),
		story.Diff,
	)
}

// BuildSecurityEvaluatorPrompt constructs the prompt for the security-focused evaluator.
func BuildSecurityEvaluatorPrompt(evaluatorID int, story StoryContext) string {
	storyJSON, err := json.MarshalIndent(story.Story, "", "  ")
	if err != nil {
		storyJSON = []byte(fmt.Sprintf(`{"id": %q, "title": %q, "error": "marshal failed"}`, story.Story.ID, story.Story.Title))
	}
	return embed.GetSecurityEvaluatorPrompt(
		fmt.Sprintf("SEC-%d", evaluatorID),
		string(storyJSON),
		story.Diff,
	)
}

// BuildDeliberationPrompt constructs the prompt for a deliberation round.
func BuildDeliberationPrompt(evaluatorID int, story StoryContext, own *EvaluatorResult, others []*EvaluatorResult) string {
	storyJSON, err := json.MarshalIndent(story.Story, "", "  ")
	if err != nil {
		storyJSON = []byte(fmt.Sprintf(`{"id": %q, "title": %q, "error": "marshal failed"}`, story.Story.ID, story.Story.Title))
	}

	ownJSON, err := json.MarshalIndent(own.Scores, "", "  ")
	if err != nil {
		ownJSON = []byte("[]")
	}

	var otherParts []string
	for _, o := range others {
		j, err := json.MarshalIndent(o.Scores, "", "  ")
		if err != nil {
			j = []byte("[]")
		}
		otherParts = append(otherParts, fmt.Sprintf("Evaluator %d:\n%s", o.EvaluatorID, string(j)))
	}

	return embed.GetDeliberationPrompt(
		fmt.Sprintf("%d", evaluatorID),
		string(storyJSON),
		string(ownJSON),
		strings.Join(otherParts, "\n\n"),
	)
}
