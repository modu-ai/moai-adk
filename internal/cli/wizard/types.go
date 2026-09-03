// Package wizard provides an interactive Bubble Tea-based wizard
// for MoAI project initialization.
package wizard

import (
	"errors"

	"github.com/modu-ai/moai-adk/pkg/models"
)

// WizardResult holds the user's selections from the init wizard.
type WizardResult struct {
	// Identity / locale (asked first so the rest of the wizard renders in the
	// chosen language).
	ConversationLang string // conversation_language code: en, ko, ja, zh
	UserName         string // user.name display value (empty allowed)

	// Core settings
	ProjectName string // Project name (required)

	// Development methodology
	DevelopmentMode string // Development mode: ddd, tdd

	// Model policy (project-level) — the model+effort profile selection
	// {high, medium, low} normalized to {max, medium, low} at persistence.
	ModelPolicy string // Token tier: high, medium, low

	// Report format — html+md or md. Persisted to report.yaml at init.
	// Empty resolves to the html+md default at persistence time.
	ReportFormat string // Report output format: html+md, md

	// Git settings
	GitMode           string // Git automation mode: manual, personal, team
	GitProvider       string // Git provider: "github", "gitlab"
	GitHubUsername    string // GitHub username (required for personal/team modes)
	GitHubToken       string // GitHub personal access token (optional)
	GitLabInstanceURL string // GitLab instance URL (for self-hosted, e.g. "https://gitlab.company.com")
	GitLabUsername    string // GitLab username (for personal/team modes with gitlab provider)
	GitLabToken       string // GitLab personal access token (optional)

	// Page-3 fields ("Quality & Workflow"). Only project_mode is still asked;
	// the four booleans are fixed at their shipped defaults and no longer
	// prompted (removed 2026-08-03) — seeded by RunWithDefaults /
	// RunWithLocale. The two mode-flag fields that formerly gated them are
	// retired (REQ-WIZ-018).
	ProjectMode               string // project.mode: personal, team (B1) — asked
	LSPEnabled                bool   // lsp.enabled: true (fixed default, no longer asked)
	EnforceQuality            bool   // quality.enforce_quality: true (fixed default, no longer asked)
	CoverageExemptionsEnabled bool   // quality.coverage_exemptions.enabled: false (fixed default)
	DesignEnabled             bool   // design.enabled: true (fixed default, no longer asked)
	ClaudeDesignEnabled       bool   // design.claude_design.enabled: true (fixed default, no longer asked)

	// Worktree advisory. Wizard-collected;
	// seeded false when the question is skipped (--non-interactive).
	WorktreeAutoCreate bool // workflow.worktree.auto_create

	// TodoEnabled is the workflow.todo.enabled answer. It is a pointer, unlike
	// its Page-3 neighbours, because the underlying config gate is default-ON
	// and absence must stay distinguishable from a "no": nil means the question
	// was never asked (--non-interactive), and an unasked question writes
	// nothing rather than falling through a zero value into `enabled: false`.
	TodoEnabled *bool // workflow.todo.enabled

	// FeedbackAutoSubmit is the feedback.auto_submit answer. Like TodoEnabled
	// it is a pointer so that "never asked" (--non-interactive) stays
	// distinguishable from an explicit "no": an unasked question writes
	// nothing rather than restating the shipped default in every project.
	FeedbackAutoSubmit *bool // feedback.auto_submit

	// ProjectContinuation (SPEC-PROJECT-CONTINUATION-KEY-001 REQ-PCK-010): the
	// /moai project Phase 14 completion selection, reusing the
	// config.ProjectContinuation* enum. A plain string, not a pointer: the
	// default is a NAMED token of the domain (card), so "" and "card" carry the
	// same meaning to the resolver, and "" additionally means the wizard did not
	// reach the question (--non-interactive) — which the writer reads as "write
	// nothing".
	ProjectContinuation string // workflow.project.continuation: none|card|pipeline

	// AutonomyTier (SPEC-AUTONOMY-TIERS-001 M7): the interactive autonomy-tier
	// selection. Reuses the config.AutonomyTier* enum. Empty when the wizard did
	// not collect a selection (--non-interactive → downstream resolves semi-auto).
	AutonomyTier string // workflow.autonomy_tier

	// M4 audit + MCP opt-in selection (SPEC-MOAI-MCP-SERVER-001 REQ-MCP-015 /
	// AC-MCP-020). Reuses the M3 typed-config vocabulary (config.AuditModel* /
	// AuditGate*). Empty strings mean "wizard did not run"; the init writer
	// treats the whole block as opt-in via AuditConfigSet.
	AuditModel        string // audit.model: claude|codex|glm|multi
	AuditGateClaude   string // audit.gates.claude: off|advisory|required
	AuditGateCodex    string // audit.gates.codex: off|advisory|required
	AuditGateGLM      string // audit.gates.glm: off|advisory|required
	CodexAuditEnabled bool   // codex.review_gate.enabled (M2 Stop-hook opt-in)
	MCPProvision      bool   // moai MCP server provisioning (default-on per SPEC-MCP-DEFAULT-ON-001)

	// AgentWiring (SPEC-INIT-HARNESS-PROMPT-001 REQ-IHP-001/002): the
	// interactive agent-harness selection, reusing the --agent closed set
	// {claude, codex, both}. Empty means the wizard did not run (or did not
	// reach the question), which resolveAgentWiringWithWizard resolves to the
	// claude fallback — the same fallback the flag path has always used.
	AgentWiring string // agent harness: claude, codex, both
}

// QuestionType represents the type of wizard question.
type QuestionType int

const (
	// QuestionTypeSelect is a single-choice selection question.
	QuestionTypeSelect QuestionType = iota
	// QuestionTypeInput is a text input question.
	QuestionTypeInput
	// QuestionTypeConfirm is a yes/no boolean question.
	QuestionTypeConfirm
)

// Question defines a single wizard question.
type Question struct {
	ID          string                   // Unique identifier
	Type        QuestionType             // Select or Input
	Title       string                   // Question title
	Description string                   // Additional description
	Options     []Option                 // Options for select questions
	Default     string                   // Default value
	Required    bool                     // Whether the field is required
	Condition   func(*WizardResult) bool // Condition for showing this question
	// Group is an optional partition label for the unified multi-group form
	// (SPEC-CLI-TUX-V3-002 REQ-TUX2-006): consecutive unconditional questions
	// sharing the same Group label render on one form page. Empty is valid
	// (label-less partition).
	Group string
}

// Option represents a selectable option.
type Option struct {
	Label string // Display label
	Value string // Actual value stored
	Desc  string // Optional description
}

// State represents the current state of the wizard.
type State int

const (
	// StateRunning means the wizard is actively running.
	StateRunning State = iota
	// StateCompleted means the wizard finished successfully.
	StateCompleted
	// StateCancelled means the user cancelled the wizard.
	StateCancelled
)

// Error definitions for the wizard package.
var (
	// ErrCancelled is returned when the user cancels the wizard.
	ErrCancelled = errors.New("wizard cancelled by user")
	// ErrNoQuestions is returned when no questions are provided.
	ErrNoQuestions = errors.New("no questions provided")
	// ErrInvalidQuestion is returned when a question index is out of bounds.
	ErrInvalidQuestion = errors.New("invalid question index")
)

// LangNameMap is an alias to the canonical language map in pkg/models.
// Deprecated: Use models.LangNameMap directly.
var LangNameMap = models.LangNameMap

// GetLanguageName returns the full language name for a code.
// Returns "English" if the code is not found.
func GetLanguageName(code string) string {
	return models.GetLanguageName(code)
}
