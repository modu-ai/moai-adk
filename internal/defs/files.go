package defs

// Common file names used across the project.
const (
	// SettingsJSON is the Claude Code project settings file.
	SettingsJSON = "settings.json"

	// SettingsLocalJSON is the Claude Code local settings override file.
	SettingsLocalJSON = "settings.local.json"

	// MCPJSON is the MCP server configuration file.
	MCPJSON = ".mcp.json"

	// ManifestJSON is the MoAI manifest file that tracks deployed templates.
	ManifestJSON = "manifest.json"

	// ClaudeMD is the main Claude Code execution directive file.
	ClaudeMD = "CLAUDE.md"

	// CredentialsJSON is the rank service credentials file.
	CredentialsJSON = "credentials.json"

	// GithubSpecRegistryJSON is the file that maps GitHub issues to SPEC IDs.
	GithubSpecRegistryJSON = "github-spec-registry.json"
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
)
