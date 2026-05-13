package defs

// Top-level directory names used by MoAI-ADK projects.
const (
	// MoAIDir is the hidden directory that stores MoAI project state.
	MoAIDir = ".moai"

	// ClaudeDir is the hidden directory that stores Claude Code configuration.
	ClaudeDir = ".claude"

	// BackupsDir is the directory where project backups are stored.
	BackupsDir = ".moai-backups"
)

// MoAI subdirectory segments (relative to MoAIDir).
const (
	ConfigSubdir   = "config"
	SectionsSubdir = "config/sections"
	StateSubdir    = "state"
	LogsSubdir     = "logs"

	// EvolutionSubdir is the subdirectory for the Reflective Learning system.
	// It holds telemetry records, learning entries, and rate-limit state.
	EvolutionSubdir = "evolution"

	// SearchSubdir is the subdirectory for the search index database.
	// It holds the SQLite FTS5 database for session search.
	SearchSubdir = ".moai/search"
)

// Claude subdirectory segments (relative to ClaudeDir).
const (
	AgentsMoaiSubdir   = "agents/moai"
	SkillsSubdir       = "skills"
	CommandsMoaiSubdir = "commands/moai"
	RulesMoaiSubdir    = "rules/moai"
	OutputStylesSubdir = "output-styles"
	HooksMoaiSubdir    = "hooks/moai"
)
