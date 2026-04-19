package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configFile = ".chief/config.yaml"

// Config holds project-level settings for Chief.
type Config struct {
	Worktree           WorktreeConfig         `yaml:"worktree"`
	OnComplete         OnCompleteConfig       `yaml:"onComplete"`
	Agent              AgentConfig            `yaml:"agent"`
	Evaluation         EvaluationConfig       `yaml:"evaluation"`
	SecurityEvaluation SecurityEvaluationConfig `yaml:"securityEvaluation"`
}

// EvaluationConfig holds adversarial evaluation settings.
type EvaluationConfig struct {
	Enabled       bool   `yaml:"enabled"`       // opt-in, default false
	Agents        int    `yaml:"agents"`         // number of evaluator agents, default 3
	PassThreshold int    `yaml:"passThreshold"`  // minimum score per criterion (1-10), default 7
	MaxRetries    int    `yaml:"maxRetries"`     // retry attempts per story on failure, default 3
	Mode          string `yaml:"mode"`           // evaluator output style, default "caveman"
	Provider      string `yaml:"provider"`       // defaults to same as main agent provider
	Model         string `yaml:"model"`          // model override for evaluators (e.g. "claude-sonnet-4-5-20250514")
}

// SecurityEvaluationConfig holds security-focused evaluation settings.
// Operates independently of EvaluationConfig — can be enabled separately or together.
type SecurityEvaluationConfig struct {
	Enabled       bool   `yaml:"enabled"`       // opt-in, default false
	Agents        int    `yaml:"agents"`         // number of security evaluator agents, default 1
	PassThreshold int    `yaml:"passThreshold"`  // minimum score per criterion (1-10), default 7
	MaxRetries    int    `yaml:"maxRetries"`     // retry attempts per story on failure, default 3
	Provider      string `yaml:"provider"`       // defaults to same as main agent provider
	Model         string `yaml:"model"`          // model override for security evaluators
}

// DefaultEvaluation returns sensible defaults for evaluation config.
func DefaultEvaluation() EvaluationConfig {
	return EvaluationConfig{
		Enabled:       false,
		Agents:        3,
		PassThreshold: 7,
		MaxRetries:    3,
		Mode:          "caveman",
	}
}

// DefaultSecurityEvaluation returns sensible defaults for security evaluation config.
func DefaultSecurityEvaluation() SecurityEvaluationConfig {
	return SecurityEvaluationConfig{
		Enabled:       false,
		Agents:        1,
		PassThreshold: 7,
		MaxRetries:    3,
	}
}

// AgentConfig holds agent CLI settings (Claude, Codex, OpenCode, or Cursor).
type AgentConfig struct {
	Provider string `yaml:"provider"` // "claude" (default) | "codex" | "opencode" | "cursor"
	CLIPath  string `yaml:"cliPath"`  // optional custom path to CLI binary
}

// WorktreeConfig holds worktree-related settings.
type WorktreeConfig struct {
	Setup string `yaml:"setup"`
}

// OnCompleteConfig holds post-completion automation settings.
type OnCompleteConfig struct {
	Push     bool `yaml:"push"`
	CreatePR bool `yaml:"createPR"`
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		Evaluation:         DefaultEvaluation(),
		SecurityEvaluation: DefaultSecurityEvaluation(),
	}
}

// ApplyEvaluationDefaults fills in zero-value evaluation fields with defaults.
// This preserves user-set values while filling gaps.
func (c *Config) ApplyEvaluationDefaults() {
	d := DefaultEvaluation()
	if c.Evaluation.Agents == 0 {
		c.Evaluation.Agents = d.Agents
	}
	if c.Evaluation.PassThreshold == 0 {
		c.Evaluation.PassThreshold = d.PassThreshold
	}
	if c.Evaluation.MaxRetries == 0 {
		c.Evaluation.MaxRetries = d.MaxRetries
	}
	if c.Evaluation.Mode == "" {
		c.Evaluation.Mode = d.Mode
	}
}

// ApplySecurityEvaluationDefaults fills in zero-value security evaluation fields with defaults.
func (c *Config) ApplySecurityEvaluationDefaults() {
	d := DefaultSecurityEvaluation()
	if c.SecurityEvaluation.Agents == 0 {
		c.SecurityEvaluation.Agents = d.Agents
	}
	if c.SecurityEvaluation.PassThreshold == 0 {
		c.SecurityEvaluation.PassThreshold = d.PassThreshold
	}
	if c.SecurityEvaluation.MaxRetries == 0 {
		c.SecurityEvaluation.MaxRetries = d.MaxRetries
	}
}

// configPath returns the full path to the config file.
func configPath(baseDir string) string {
	return filepath.Join(baseDir, configFile)
}

// Exists checks if the config file exists.
func Exists(baseDir string) bool {
	_, err := os.Stat(configPath(baseDir))
	return err == nil
}

// Load reads the config from .chief/config.yaml.
// Returns Default() when the file doesn't exist (no error).
func Load(baseDir string) (*Config, error) {
	path := configPath(baseDir)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Fill in any zero-value evaluation fields with defaults
	cfg.ApplyEvaluationDefaults()
	cfg.ApplySecurityEvaluationDefaults()

	return cfg, nil
}

// Save writes the config to .chief/config.yaml.
func Save(baseDir string, cfg *Config) error {
	path := configPath(baseDir)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
