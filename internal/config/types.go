package config

import (
	"github.com/modu-ai/moai-adk/pkg/models"
)

// Config is the root configuration aggregate containing all sections.
// It imports types from pkg/models for shared types (UserConfig, LanguageConfig,
// QualityConfig, ProjectConfig) and defines internal types for the rest.
type Config struct {
	User          models.UserConfig          `yaml:"user"`
	Language      models.LanguageConfig      `yaml:"language"`
	Quality       models.QualityConfig       `yaml:"quality"`
	Project       models.ProjectConfig       `yaml:"project"`
	GitStrategy   GitStrategyConfig          `yaml:"git_strategy"`
	GitConvention models.GitConventionConfig `yaml:"git_convention"`
	System        SystemConfig               `yaml:"system"`
	LLM           LLMConfig                  `yaml:"llm"`
	Pricing       PricingConfig              `yaml:"pricing"`
	Ralph         RalphConfig                `yaml:"ralph"`
	Workflow      WorkflowConfig             `yaml:"workflow"`
}

// GitStrategyConfig represents the git strategy configuration section.
type GitStrategyConfig struct {
	AutoBranch        bool   `yaml:"auto_branch"`
	BranchPrefix      string `yaml:"branch_prefix"`
	CommitStyle       string `yaml:"commit_style"`
	WorktreeRoot      string `yaml:"worktree_root"`
	Provider          string `yaml:"provider"`            // "github", "gitlab"
	GitLabInstanceURL string `yaml:"gitlab_instance_url"` // GitLab instance URL
}

// SystemConfig represents the system configuration section.
type SystemConfig struct {
	Version        string `yaml:"version"`
	LogLevel       string `yaml:"log_level"`
	LogFormat      string `yaml:"log_format"`
	NoColor        bool   `yaml:"no_color"`
	NonInteractive bool   `yaml:"non_interactive"`
}

// LLMConfig represents the LLM configuration section.
type LLMConfig struct {
	// Mode selection: claude-only, glm-only, mashup
	Mode string `yaml:"mode"`
	// Environment variable name for GLM API key
	GLMEnvVar string `yaml:"glm_env_var"`
	// GLM API configuration
	GLM GLMSettings `yaml:"glm"`
	// Legacy fields (kept for backward compatibility)
	DefaultModel string `yaml:"default_model"`
	QualityModel string `yaml:"quality_model"`
	SpeedModel   string `yaml:"speed_model"`
}

// GLMSettings represents GLM API configuration.
type GLMSettings struct {
	BaseURL string    `yaml:"base_url"`
	Models  GLMModels `yaml:"models"`
}

// GLMModels represents GLM model mappings.
type GLMModels struct {
	Haiku  string `yaml:"haiku"`
	Sonnet string `yaml:"sonnet"`
	Opus   string `yaml:"opus"`
}

// PricingConfig represents the pricing configuration section.
type PricingConfig struct {
	TokenBudget  int  `yaml:"token_budget"`
	CostTracking bool `yaml:"cost_tracking"`
}

// RalphConfig represents the Ralph engine configuration section.
type RalphConfig struct {
	MaxIterations int  `yaml:"max_iterations"`
	AutoConverge  bool `yaml:"auto_converge"`
	HumanReview   bool `yaml:"human_review"`
}

// WorkflowConfig represents the workflow configuration section.
type WorkflowConfig struct {
	AutoClear  bool `yaml:"auto_clear"`
	PlanTokens int  `yaml:"plan_tokens"`
	RunTokens  int  `yaml:"run_tokens"`
	SyncTokens int  `yaml:"sync_tokens"`
}

// LSPQualityGates represents LSP quality gate configuration.
type LSPQualityGates struct {
	Enabled         bool     `yaml:"enabled"`
	Plan            PlanGate `yaml:"plan"`
	Run             RunGate  `yaml:"run"`
	Sync            SyncGate `yaml:"sync"`
	CacheTTLSeconds int      `yaml:"cache_ttl_seconds"`
	TimeoutSeconds  int      `yaml:"timeout_seconds"`
}

// PlanGate represents the plan phase quality gate.
type PlanGate struct {
	RequireBaseline bool `yaml:"require_baseline"`
}

// RunGate represents the run phase quality gate.
type RunGate struct {
	MaxErrors       int  `yaml:"max_errors"`
	MaxTypeErrors   int  `yaml:"max_type_errors"`
	MaxLintErrors   int  `yaml:"max_lint_errors"`
	AllowRegression bool `yaml:"allow_regression"`
}

// SyncGate represents the sync phase quality gate.
type SyncGate struct {
	MaxErrors       int  `yaml:"max_errors"`
	MaxWarnings     int  `yaml:"max_warnings"`
	RequireCleanLSP bool `yaml:"require_clean_lsp"`
}

// sectionNames lists all valid configuration section names.
var sectionNames = []string{
	"user", "language", "quality", "project",
	"git_strategy", "git_convention", "system", "llm",
	"pricing", "ralph", "workflow",
}

// IsValidSectionName checks if the given name is a valid section name.
func IsValidSectionName(name string) bool {
	for _, s := range sectionNames {
		if s == name {
			return true
		}
	}
	return false
}

// ValidSectionNames returns all valid section names.
func ValidSectionNames() []string {
	result := make([]string, len(sectionNames))
	copy(result, sectionNames)
	return result
}

// YAML file wrapper types for proper unmarshaling with top-level keys.
// Each section file wraps its content under a top-level key.

type userFileWrapper struct {
	User models.UserConfig `yaml:"user"`
}

type languageFileWrapper struct {
	Language models.LanguageConfig `yaml:"language"`
}

// qualityFileWrapper handles the quality.yaml file which uses "constitution:"
// as the top-level key (Python MoAI-ADK backward compatibility).
type qualityFileWrapper struct {
	Constitution models.QualityConfig `yaml:"constitution"`
}

// gitConventionFileWrapper handles the git-convention.yaml section file.
type gitConventionFileWrapper struct {
	GitConvention models.GitConventionConfig `yaml:"git_convention"`
}
