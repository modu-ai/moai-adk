package defs

// Common file names used across the project.
const (
	// SettingsJSON is the Claude Code project settings file.
	SettingsJSON = "settings.json"

	// SettingsLocalJSON is the Claude Code local settings override file.
	SettingsLocalJSON = "settings.local.json"

	// ManifestJSON is the MoAI manifest file that tracks deployed templates.
	ManifestJSON = "manifest.json"

	// ClaudeMD is the canonical shared instruction file (CLAUDE.md-canonical
	// product, SPEC-CLAUDE-INSTRUCTION-FILES-001): every `moai init --llm`
	// value scaffolds it and no AGENTS.md.
	ClaudeMD = "CLAUDE.md"

	// AgentsMD is the legacy cross-harness instruction file. Retired as a
	// deploy payload — the name survives for legacy-project detection
	// (reader fallback, migration sources) and the frozen-instruction-files
	// contract only.
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
