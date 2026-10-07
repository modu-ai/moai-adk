package defs

// Common file names used across the project.
const (
	// SettingsJSON is the Claude Code project settings file.
	SettingsJSON = "settings.json"

	// SettingsLocalJSON is the Claude Code local settings override file.
	SettingsLocalJSON = "settings.local.json"

	// ManifestJSON is the MoAI manifest file that tracks deployed templates.
	ManifestJSON = "manifest.json"

	// ClaudeMD is the legacy Claude Code execution directive file. Retired as
	// a deploy payload — the name survives for legacy-project detection and
	// the frozen-instruction-files contract only.
	ClaudeMD = "CLAUDE.md"

	// AgentsMD is the cross-harness primary instruction file (AGENTS.md-primary
	// product): every `moai init --llm` value scaffolds it and no CLAUDE.md.
	AgentsMD = "AGENTS.md"
)

// Section YAML file names under .moai/config/sections/.
const (
	UserYAML        = "user.yaml"
	LanguageYAML    = "language.yaml"
	QualityYAML     = "quality.yaml"
	WorkflowYAML    = "workflow.yaml"
	ProjectYAML     = "project.yaml"
	GitStrategyYAML = "git-strategy.yaml"
	SystemYAML      = "system.yaml"
	StatuslineYAML  = "statusline.yaml"
	HarnessYAML     = "harness.yaml"
	LSPYAML         = "lsp.yaml"
	DesignYAML      = "design.yaml"
	ReportYAML      = "report.yaml"
	FeedbackYAML    = "feedback.yaml"
)
