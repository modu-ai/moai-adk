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

// DeprecatedPathEntry describes a single deprecated template path.
type DeprecatedPathEntry struct {
	// Path is the slash-separated path relative to the project root.
	Path string
	// DeprecatedSince is the SPEC ID that removed this path from templates.
	DeprecatedSince string
	// DeprecatedBy is the SPEC ID authorising cleanup.
	DeprecatedBy string
	// RemovalSchedule is the version when deletion is expected.
	RemovalSchedule string
}

// DeprecatedPaths is the single source of truth for paths that have been removed
// from the MoAI-ADK templates but may still exist in user projects.
// Populated by SPEC-AGENCY-ABSORB-001 (2026-04-23, commit 3e8b61e80).
// Managed by: SPEC-V3R3-UPDATE-CLEANUP-001 (REQ-UPC-006).
var DeprecatedPaths = []DeprecatedPathEntry{
	{
		Path:            ".claude/commands/agency/agency.md",
		DeprecatedSince: "SPEC-AGENCY-ABSORB-001",
		DeprecatedBy:    "SPEC-V3R3-UPDATE-CLEANUP-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/commands/agency/brief.md",
		DeprecatedSince: "SPEC-AGENCY-ABSORB-001",
		DeprecatedBy:    "SPEC-V3R3-UPDATE-CLEANUP-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/commands/agency/build.md",
		DeprecatedSince: "SPEC-AGENCY-ABSORB-001",
		DeprecatedBy:    "SPEC-V3R3-UPDATE-CLEANUP-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/commands/agency/evolve.md",
		DeprecatedSince: "SPEC-AGENCY-ABSORB-001",
		DeprecatedBy:    "SPEC-V3R3-UPDATE-CLEANUP-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/commands/agency/learn.md",
		DeprecatedSince: "SPEC-AGENCY-ABSORB-001",
		DeprecatedBy:    "SPEC-V3R3-UPDATE-CLEANUP-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/commands/agency/profile.md",
		DeprecatedSince: "SPEC-AGENCY-ABSORB-001",
		DeprecatedBy:    "SPEC-V3R3-UPDATE-CLEANUP-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/commands/agency/resume.md",
		DeprecatedSince: "SPEC-AGENCY-ABSORB-001",
		DeprecatedBy:    "SPEC-V3R3-UPDATE-CLEANUP-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/commands/agency/review.md",
		DeprecatedSince: "SPEC-AGENCY-ABSORB-001",
		DeprecatedBy:    "SPEC-V3R3-UPDATE-CLEANUP-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/rules/agency/constitution.md",
		DeprecatedSince: "SPEC-AGENCY-ABSORB-001",
		DeprecatedBy:    "SPEC-V3R3-UPDATE-CLEANUP-001",
		RemovalSchedule: "v3.0.0",
	},
}

// MoAI subdirectory segments (relative to MoAIDir).
const (
	ConfigSubdir   = "config"
	SectionsSubdir = "config/sections"
	StateSubdir    = "state"
	LogsSubdir     = "logs"

	// EvolutionSubdir is the subdirectory for the Reflective Learning system.
	// It holds telemetry records, learning entries, and rate-limit state.
	EvolutionSubdir = "evolution"
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
