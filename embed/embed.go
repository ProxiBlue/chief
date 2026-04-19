// Package embed provides embedded prompt templates used by Chief.
// All prompts are embedded at compile time using Go's embed directive.
// Prompts can be overridden at runtime by placing custom files in .chief/prompts/.
package embed

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed prompt.txt
var promptTemplate string

//go:embed init_prompt.txt
var initPromptTemplate string

//go:embed edit_prompt.txt
var editPromptTemplate string

//go:embed detect_setup_prompt.txt
var detectSetupPromptTemplate string

//go:embed evaluator_prompt.txt
var evaluatorPromptTemplate string

//go:embed deliberation_prompt.txt
var deliberationPromptTemplate string

//go:embed security_evaluator_prompt.txt
var securityEvaluatorPromptTemplate string

// baseDir is the project root directory, used to locate .chief/prompts/ overrides.
var baseDir string
var baseDirOnce sync.Once

// SetBaseDir sets the project root directory for prompt override resolution.
// Call this once during startup before any Get*Prompt functions are called.
// Override files are read from <baseDir>/.chief/prompts/<filename>.
func SetBaseDir(dir string) {
	baseDirOnce.Do(func() {
		baseDir = dir
	})
}

// resolveTemplate returns the override content if a file exists at
// .chief/prompts/<filename>, otherwise returns the embedded default.
func resolveTemplate(embedded string, filename string) string {
	if baseDir == "" {
		return embedded
	}
	overridePath := filepath.Join(baseDir, ".chief", "prompts", filename)
	data, err := os.ReadFile(overridePath)
	if err != nil {
		return embedded
	}
	content := string(data)
	if strings.TrimSpace(content) == "" {
		return embedded
	}
	return content
}

// GetPrompt returns the agent prompt with the progress path and
// current story context substituted. The storyContext is the JSON of the
// current story to work on, inlined directly into the prompt so that the
// agent does not need to read the entire prd.md file.
func GetPrompt(progressPath, storyContext, storyID, storyTitle string) string {
	tmpl := resolveTemplate(promptTemplate, "prompt.txt")
	result := strings.ReplaceAll(tmpl, "{{PROGRESS_PATH}}", progressPath)
	result = strings.ReplaceAll(result, "{{STORY_CONTEXT}}", storyContext)
	result = strings.ReplaceAll(result, "{{STORY_ID}}", storyID)
	return strings.ReplaceAll(result, "{{STORY_TITLE}}", storyTitle)
}

// GetInitPrompt returns the PRD generator prompt with the PRD directory and optional context substituted.
func GetInitPrompt(prdDir, context string) string {
	if context == "" {
		context = "No additional context provided. Ask the user what they want to build."
	}
	tmpl := resolveTemplate(initPromptTemplate, "init_prompt.txt")
	result := strings.ReplaceAll(tmpl, "{{PRD_DIR}}", prdDir)
	return strings.ReplaceAll(result, "{{CONTEXT}}", context)
}

// GetEditPrompt returns the PRD editor prompt with the PRD directory substituted.
func GetEditPrompt(prdDir string) string {
	tmpl := resolveTemplate(editPromptTemplate, "edit_prompt.txt")
	return strings.ReplaceAll(tmpl, "{{PRD_DIR}}", prdDir)
}

// GetDetectSetupPrompt returns the prompt for detecting project setup commands.
func GetDetectSetupPrompt() string {
	return resolveTemplate(detectSetupPromptTemplate, "detect_setup_prompt.txt")
}

// GetEvaluatorPrompt returns the adversarial evaluator prompt with placeholders substituted.
// Uses single-pass replacement to prevent re-expansion if values contain placeholder strings.
func GetEvaluatorPrompt(evaluatorID, storyContext, diff string) string {
	tmpl := resolveTemplate(evaluatorPromptTemplate, "evaluator_prompt.txt")
	r := strings.NewReplacer(
		"{{EVALUATOR_ID}}", evaluatorID,
		"{{STORY_CONTEXT}}", storyContext,
		"{{DIFF}}", diff,
	)
	return r.Replace(tmpl)
}

// GetSecurityEvaluatorPrompt returns the security-focused evaluator prompt with placeholders substituted.
// Uses single-pass replacement to prevent re-expansion if values contain placeholder strings.
func GetSecurityEvaluatorPrompt(evaluatorID, storyContext, diff string) string {
	tmpl := resolveTemplate(securityEvaluatorPromptTemplate, "security_evaluator_prompt.txt")
	r := strings.NewReplacer(
		"{{EVALUATOR_ID}}", evaluatorID,
		"{{STORY_CONTEXT}}", storyContext,
		"{{DIFF}}", diff,
	)
	return r.Replace(tmpl)
}

// GetDeliberationPrompt returns the deliberation round prompt with placeholders substituted.
// Uses single-pass replacement to prevent re-expansion if values contain placeholder strings.
func GetDeliberationPrompt(evaluatorID, storyContext, ownFindings, otherFindings string) string {
	tmpl := resolveTemplate(deliberationPromptTemplate, "deliberation_prompt.txt")
	r := strings.NewReplacer(
		"{{EVALUATOR_ID}}", evaluatorID,
		"{{STORY_CONTEXT}}", storyContext,
		"{{OWN_FINDINGS}}", ownFindings,
		"{{OTHER_FINDINGS}}", otherFindings,
	)
	return r.Replace(tmpl)
}
