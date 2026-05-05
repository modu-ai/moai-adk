// Package config provides configuration management for MoAI-ADK Go Edition.
// It loads YAML section files, applies defaults, validates, and provides
// thread-safe access to configuration values.
package config

// @MX:NOTE: [AUTO] Environment variable key constants centralize all env var names to prevent typos and enable IDE navigation
//
// Environment variable key constants.
//
// Centralizes all environment variable names used across the codebase
// to prevent typos and enable IDE navigation.

// MoAI configuration environment variables.
const (
	// EnvConfigDir overrides the MoAI configuration directory path.
	EnvConfigDir = "MOAI_CONFIG_DIR"

	// EnvDevelopmentMode overrides the development methodology (ddd or tdd).
	EnvDevelopmentMode = "MOAI_DEVELOPMENT_MODE"

	// EnvLogLevel overrides the log level (debug, info, warn, error).
	EnvLogLevel = "MOAI_LOG_LEVEL"

	// EnvLogFormat overrides the log format (text or json).
	EnvLogFormat = "MOAI_LOG_FORMAT"

	// EnvNoColor disables color output when set to "true" or "1".
	EnvNoColor = "MOAI_NO_COLOR"

	// EnvStatuslineMode selects the statusline display mode.
	EnvStatuslineMode = "MOAI_STATUSLINE_MODE"

	// EnvStatuslineContextSize overrides the context window size used by the
	// statusline gauge. Useful when the upstream provider reports a context
	// size that does not match the actual API limit (e.g., GLM models served
	// behind the Anthropic-compatible endpoint, see issue #653).
	// Value is parsed as int (tokens). Zero or invalid → fall back to stdin.
	EnvStatuslineContextSize = "MOAI_STATUSLINE_CONTEXT_SIZE"

	// EnvSkipBinaryUpdate skips binary self-update when set to "1".
	EnvSkipBinaryUpdate = "MOAI_SKIP_BINARY_UPDATE"

	// EnvGitConvention overrides the git commit convention.
	EnvGitConvention = "MOAI_GIT_CONVENTION"

	// EnvEnforceOnPush overrides the pre-push convention enforcement flag.
	EnvEnforceOnPush = "MOAI_ENFORCE_ON_PUSH"

	// EnvUpdateSource overrides the update source ("github" or "local").
	EnvUpdateSource = "MOAI_UPDATE_SOURCE"

	// EnvReleasesDir specifies a local directory for release archives.
	EnvReleasesDir = "MOAI_RELEASES_DIR"

	// EnvUpdateURL overrides the GitHub releases API URL.
	EnvUpdateURL = "MOAI_UPDATE_URL"
)

// MoAI test-only environment variables.
const (
	// EnvTestMode enables test mode behavior when set to "1".
	EnvTestMode = "MOAI_TEST_MODE"

	// EnvTestGLMKey provides a test GLM API key for integration tests.
	EnvTestGLMKey = "MOAI_TEST_GLM_KEY"
)

// Claude Code environment variables (set by Claude Code runtime).
const (
	// EnvClaudeProjectDir is the project root directory set by Claude Code.
	EnvClaudeProjectDir = "CLAUDE_PROJECT_DIR"

	// EnvClaudeConfigDir is the Claude Code configuration directory.
	EnvClaudeConfigDir = "CLAUDE_CONFIG_DIR"

	// EnvClaudeEnvFile is the path to Claude Code's environment file.
	EnvClaudeEnvFile = "CLAUDE_ENV_FILE"

	// EnvClaudeAutoCompactPct overrides the auto-compact percentage threshold.
	EnvClaudeAutoCompactPct = "CLAUDE_AUTOCOMPACT_PCT_OVERRIDE"

	// EnvClaudeCodeEffortLevel sets the session effort level for Claude Code.
	// Valid values: "low", "medium", "high", "xhigh", "max".
	// "xhigh" and "max" are supported on Opus 4.7+.
	// When empty, the runtime default applies.
	EnvClaudeCodeEffortLevel = "CLAUDE_CODE_EFFORT_LEVEL"
)

// Anthropic API environment variables.
const (
	// EnvAnthropicBaseURL overrides the Anthropic API base URL.
	EnvAnthropicBaseURL = "ANTHROPIC_BASE_URL"

	// EnvAnthropicAuthToken provides the Anthropic API authentication token.
	EnvAnthropicAuthToken = "ANTHROPIC_AUTH_TOKEN"

	// EnvAnthropicDefaultHaikuModel overrides the default Haiku model ID.
	EnvAnthropicDefaultHaikuModel = "ANTHROPIC_DEFAULT_HAIKU_MODEL"

	// EnvAnthropicDefaultSonnetModel overrides the default Sonnet model ID.
	EnvAnthropicDefaultSonnetModel = "ANTHROPIC_DEFAULT_SONNET_MODEL"

	// EnvAnthropicDefaultOpusModel overrides the default Opus model ID.
	EnvAnthropicDefaultOpusModel = "ANTHROPIC_DEFAULT_OPUS_MODEL"
)
