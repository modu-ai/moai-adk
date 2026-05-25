package defs

// Top-level directory names used by MoAI-ADK projects.
const (
	// MoAIDir is the hidden directory that stores MoAI project state.
	MoAIDir = ".moai"

	// ClaudeDir is the hidden directory that stores Claude Code configuration.
	ClaudeDir = ".claude"

	// BackupsDir is the directory where project backups are stored.
	BackupsDir = ".moai-backups"

	// NamespaceBackupsSubdir is the subdirectory under .moai/ that stores
	// user-owned namespace backups created before destructive moai update
	// operations. Path: .moai/backups/update-<ISO-8601-UTC>/.
	//
	// Distinct from BackupsDir (.moai-backups/) which holds only config
	// backups, and from .moai/archive/skills/v2.16-drift-*/ which holds
	// archive-drift backups. See SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001
	// REQ-UNP-010.
	NamespaceBackupsSubdir = "backups"
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
// Extended by SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 M2 (2026-05-25, v3.0.0-rc2):
// Category B = 31 v.2.x-era NEW entries + Category C = 3 rc1-stage staging
// artifacts. After extension: 43 entries total per spec.md §A.4 Canonical
// DeprecatedPaths Derivation Table.
//
// @MX:ANCHOR: SSOT for v.2.x → v3 cleanup targets.
// @MX:REASON: External-user cleanup correctness depends on the 43-entry total
// + 9/31/3 category split; modifications MUST update both this slice and
// spec.md §A.4 atomically. See SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001.
var DeprecatedPaths = []DeprecatedPathEntry{
	// ====================================================================
	// Category A — Pre-existing entries (9; SPEC-AGENCY-ABSORB-001)
	// ====================================================================
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
	// ====================================================================
	// Category B — v.2.x-era NEW entries (31; SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001)
	// All paths use v.2.x FLAT layout verified by `git ls-tree -r 1bd083725^`.
	// ====================================================================
	// v2 directories
	{
		Path:            ".agency",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".agency.archived",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	// agency agents (v.2.x FLAT layout)
	{
		Path:            ".claude/agents/moai/planner.md",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/agents/moai/designer.md",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/agents/moai/builder.md",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/agents/moai/evaluator.md",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	// retired manager agents (v.2.x FLAT layout)
	{
		Path:            ".claude/agents/moai/manager-strategy.md",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/agents/moai/manager-quality.md",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/agents/moai/manager-brain.md",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/agents/moai/manager-project.md",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	// retired meta agents (v.2.x FLAT layout)
	{
		Path:            ".claude/agents/moai/claude-code-guide.md",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/agents/moai/researcher.md",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	// retired expert agents (v.2.x FLAT layout)
	{
		Path:            ".claude/agents/moai/expert-backend.md",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/agents/moai/expert-frontend.md",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/agents/moai/expert-security.md",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/agents/moai/expert-devops.md",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/agents/moai/expert-performance.md",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/agents/moai/expert-refactoring.md",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	// deprecated config yaml files
	{
		Path:            ".moai/config/sections/design.yaml",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".moai/config/sections/db.yaml",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".moai/config/sections/gate.yaml",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".moai/config/sections/github-actions.yaml",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".moai/config/sections/memo.yaml",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	// design skill directories
	{
		Path:            ".claude/skills/moai-domain-brand-design",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/skills/moai-domain-copywriting",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/skills/moai-domain-design-handoff",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	// design workflow skill directories
	{
		Path:            ".claude/skills/moai-workflow-design",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/skills/moai-workflow-gan-loop",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	// design rule directory
	{
		Path:            ".claude/rules/moai/design",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	// brand + db directories
	{
		Path:            ".moai/project/brand",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".moai/db",
		DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	// ====================================================================
	// Category C — rc1-stage staging artifacts (3; SPEC-V3R6-AGENT-FOLDER-SPLIT-001)
	// These directories never existed in v.2.x; they were introduced by
	// commit 1bd083725 (2026-05-22) and are reverted by SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001
	// M2a. Inclusion ensures rc1-stage early adopters receive the
	// layout-correction cleanup through `moai update`.
	// ====================================================================
	{
		Path:            ".claude/agents/core",
		DeprecatedSince: "SPEC-V3R6-AGENT-FOLDER-SPLIT-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/agents/expert",
		DeprecatedSince: "SPEC-V3R6-AGENT-FOLDER-SPLIT-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
		RemovalSchedule: "v3.0.0",
	},
	{
		Path:            ".claude/agents/meta",
		DeprecatedSince: "SPEC-V3R6-AGENT-FOLDER-SPLIT-001",
		DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
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
