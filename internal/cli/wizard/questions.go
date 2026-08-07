package wizard

import (
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/config"
)

// The wizard question set is split across these constructors:
//
//   - DefaultQuestions      — pages 1-2 of the `moai init` set (NO Git questions)
//   - Page3Questions        — page 3 of the `moai init` set
//   - InitQuestions         — DefaultQuestions + Page3Questions: the full
//     3-page `moai init` set, assembled once for the wizard entry point
//   - GitQuestions          — the 7 Git questions, on their own
//   - ReconfigureQuestions  — DefaultQuestions with GitQuestions spliced back
//     in, used by the `moai update --reconfigure` path. It deliberately does
//     NOT include the page-3 questions, so the reconfigure set keeps its
//     pre-restructure membership.
//
// The `moai init` wizard renders three topic pages
// (SPEC-CLI-WIZARD-RESTRUCTURE-001 REQ-WIZ-003..005):
//
//	Page 1 "Basic"              — conversation_language, user_name, project_name
//	Page 2 "Model & Report"     — model_policy, report_format
//	Page 3 "Quality & Workflow" — lsp_enabled, enforce_quality, project_mode,
//	                              design_enabled, claude_design_enabled (nested)
//
// A page is a run of consecutive UNCONDITIONAL questions sharing one Group
// label: buildFormGroups (wizard.go) merges each such run into a single huh
// group. Pages 1-2 come from DefaultQuestions, page 3 from Page3Questions.
//
// DefaultQuestions returns pages 1-2, in this order:
//  1. Conversation language (drives the rendering language of every later question)
//  2. User name (optional)
//  3. Project name (required)
//  4. Model policy
//  5. Report format
//
// Git mode and provider are NOT asked here: `moai init` derives them from the
// repository's configured remotes (see detectGitConfig in internal/cli), so a
// fresh init no longer interrogates the user about Git. The reconfigure path
// still asks them — see ReconfigureQuestions.
//
// plan_type and development_mode are NO LONGER interactive questions — they
// default silently (plan_type → subscription at persistence; development_mode →
// tdd in init.go) and remain overridable via the --plan-type / --mode flags.
func DefaultQuestions(projectRoot string) []Question {
	// Use current directory name as default project name
	defaultProjectName := filepath.Base(projectRoot)
	if defaultProjectName == "." || defaultProjectName == "/" || defaultProjectName == `\` {
		defaultProjectName = "my-project"
	}

	return []Question{
		// 1. Conversation language — asked first so every subsequent question
		// renders in the chosen language (saveAnswer updates the live locale).
		// The default is pre-filled from the profile's ConversationLang by the
		// wizard entry point when available (else "en"). Option labels carry the
		// native language name and are never translated (GetLocalizedQuestion
		// leaves them untouched because the translation entry supplies no options).
		{
			ID:          "conversation_language",
			Group:       "Basic",
			Type:        QuestionTypeSelect,
			Title:       "Select conversation language",
			Description: "The language MoAI uses when talking with you. The wizard switches to it immediately.",
			Options: []Option{
				{Label: "English", Value: "en", Desc: "English"},
				{Label: "Korean (한국어)", Value: "ko", Desc: "한국어"},
				{Label: "Japanese (日本語)", Value: "ja", Desc: "日本語"},
				{Label: "Chinese (中文)", Value: "zh", Desc: "中文"},
			},
			Default:  "en",
			Required: true,
		},
		// 2. User Name — optional; pre-filled from the profile's UserName by the
		// wizard entry point when available. Empty is allowed.
		{
			ID:          "user_name",
			Group:       "Basic",
			Type:        QuestionTypeInput,
			Title:       "Enter your name",
			Description: "How MoAI addresses you. Persisted to user.yaml (user.name). Leave empty to skip.",
			Default:     "",
			Required:    false,
		},
		// 3. Project Name
		{
			ID:          "project_name",
			Group:       "Basic",
			Type:        QuestionTypeInput,
			Title:       "Enter project name",
			Description: "The name of your project.",
			Default:     defaultProjectName,
			Required:    true,
		},
		// 2. Model Policy
		{
			ID:          "model_policy",
			Group:       "Model & Report",
			Type:        QuestionTypeSelect,
			Title:       "Select model policy",
			Description: "Controls which Claude model tier is assigned to each agent. Match to your Claude plan.",
			// Labels use the v3.0.1 tier naming (Max / Medium / Low). Values stay
			// "high"/"medium"/"low": the internal ModelPolicy vocabulary
			// (internal/template/model_policy.go IsValidModelPolicy) expects those
			// values and NormalizeToTier maps high→max downstream.
			// Descriptions mirror the actual per-tier assignments in the profile
			// matrix SSOT (internal/template/profile_matrix.go defaultProfileMatrix,
			// Matrix A): max leans on Fable + Opus for core agents; medium/low mix
			// Opus and Sonnet across effort levels. Keep these in sync with that
			// matrix, not with a marketing summary.
			// The (Recommended) marker tracks the Default below: Medium is the default
			// for new projects (SPEC-CLI-WIZARD-RESTRUCTURE-001 REQ-WIZ-008). Max/High
			// remains a fully selectable tier — only the DEFAULT moved.
			Options: []Option{
				{Label: "Max", Value: "high", Desc: "Opus 5 (max~medium) + Sonnet (low, single-shot rows only) — Max $200 plan"},
				{Label: "Medium (Recommended)", Value: "medium", Desc: "Opus 5 (high~low) + Sonnet (low, single-shot rows only) — Max $100 plan"},
				{Label: "Low", Value: "low", Desc: "Opus 5 (medium~low) + Sonnet (low, docs/e2e/single-shot rows) — Plus $20 plan"},
			},
			Default:  "medium",
			Required: true,
		},
		// 3. Report Format — html+md vs md.
		// The value set mirrors internal/settings reportFormatValues (the closed
		// set {"html+md", "md"} consumed by the moai-domain-html-report skill via
		// report.format). Keep these two Values in sync with that SSOT.
		{
			ID:          "report_format",
			Group:       "Model & Report",
			Type:        QuestionTypeSelect,
			Title:       "Select report format",
			Description: "Controls whether reports are generated as HTML+Markdown or Markdown only.",
			Options: []Option{
				{Label: "HTML + Markdown (Recommended)", Value: "html+md", Desc: "Generate both an HTML report (browser-viewable) and Markdown"},
				{Label: "Markdown only", Value: "md", Desc: "Generate Markdown reports only (lighter, diff-friendly)"},
			},
			Default:  "html+md",
			Required: true,
		},
	}
}

// GitQuestions returns the 7 Git questions, in their canonical order:
//  1. Git mode
//  2. Git provider (conditional)
//  3. GitLab instance URL (conditional)
//  4. GitHub username (conditional)
//  5. GitHub token (conditional)
//  6. GitLab username (conditional)
//  7. GitLab token (conditional)
//
// These are asked ONLY by the `moai update --reconfigure` path — runInitWizard
// (internal/cli/update.go) is the sole caller and it builds ReconfigureQuestions.
// The interactive `moai init` path is DISTINCT: init.go -> runWizardFn ->
// RunWithDefaults builds InitQuestions (DefaultQuestions + Page3Questions),
// which contain
// NO Git questions; `moai init` auto-detects mode and provider from the
// repository's remotes via detectGitConfig instead. This split is enforced by
// TestInitWizardQuestionSetHasNoGitCredentialQuestions.
func GitQuestions() []Question {
	return []Question{
		// 1. Git Mode
		{
			ID:          "git_mode",
			Group:       "Git",
			Type:        QuestionTypeSelect,
			Title:       "Select Git automation mode",
			Description: "Controls how much Git automation Claude can perform.",
			Options: []Option{
				{Label: "Manual", Value: "manual", Desc: "AI never commits or pushes"},
				{Label: "Personal", Value: "personal", Desc: "AI can create branches and commit"},
				{Label: "Team", Value: "team", Desc: "AI can create branches, commit, and open PRs"},
			},
			Default:  "manual",
			Required: true,
		},
		// 2. Git Provider (conditional - only for personal/team modes)
		{
			ID:          "git_provider",
			Group:       "Git",
			Type:        QuestionTypeSelect,
			Title:       "Select your Git provider",
			Description: "Choose the Git hosting platform for your project.",
			Options: []Option{
				{Label: "GitHub", Value: "github", Desc: "GitHub.com"},
				{Label: "GitLab", Value: "gitlab", Desc: "GitLab.com or self-hosted GitLab"},
			},
			Default:  "github",
			Required: true,
			Condition: func(r *WizardResult) bool {
				return r.GitMode == "personal" || r.GitMode == "team"
			},
		},
		// 3. GitLab Instance URL (conditional - only for gitlab provider)
		{
			ID:          "gitlab_instance_url",
			Group:       "Git",
			Type:        QuestionTypeInput,
			Title:       "Enter GitLab instance URL",
			Description: "For GitLab.com use https://gitlab.com. For self-hosted, enter your instance URL.",
			Default:     "https://gitlab.com",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return (r.GitMode == "personal" || r.GitMode == "team") && r.GitProvider == "gitlab"
			},
		},
		// 4. GitHub Username (conditional - only for github provider)
		{
			ID:          "github_username",
			Group:       "Git",
			Type:        QuestionTypeInput,
			Title:       "Enter your GitHub username",
			Description: "Required for Git automation features.",
			Default:     "",
			Required:    false, // Conditional requirement handled by wizard
			Condition: func(r *WizardResult) bool {
				return (r.GitMode == "personal" || r.GitMode == "team") && r.GitProvider == "github"
			},
		},
		// 5. GitHub Token (conditional - only for github provider)
		{
			ID:          "github_token",
			Group:       "Git",
			Type:        QuestionTypeInput,
			Title:       "Enter GitHub personal access token (optional)",
			Description: "Prefer 'gh auth login' — it stores credentials securely outside the project and MoAI never writes them to disk. A token pasted here is NOT saved to any file; if you use one, scope it minimally (fine-grained: Contents + Pull requests read/write; classic: repo). Leave empty to use the gh CLI.",
			Default:     "",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return (r.GitMode == "personal" || r.GitMode == "team") && r.GitProvider == "github"
			},
		},
		// 6. GitLab Username (conditional - only for gitlab provider)
		{
			ID:          "gitlab_username",
			Group:       "Git",
			Type:        QuestionTypeInput,
			Title:       "Enter your GitLab username",
			Description: "Required for Git automation features with GitLab.",
			Default:     "",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return (r.GitMode == "personal" || r.GitMode == "team") && r.GitProvider == "gitlab"
			},
		},
		// 7. GitLab Token (conditional - only for gitlab provider)
		{
			ID:          "gitlab_token",
			Group:       "Git",
			Type:        QuestionTypeInput,
			Title:       "Enter GitLab personal access token (optional)",
			Description: "Prefer 'glab auth login' — it stores credentials securely outside the project and MoAI never writes them to disk. A token pasted here is NOT saved to any file; if you use one, scope it minimally (write_repository, api). Leave empty to use the glab CLI.",
			Default:     "",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return (r.GitMode == "personal" || r.GitMode == "team") && r.GitProvider == "gitlab"
			},
		},
	}
}

// ReconfigureQuestions returns the full question set used by the
// `moai update --reconfigure` path: DefaultQuestions with GitQuestions spliced
// back in at their original position — immediately after report_format — so the
// reconfigure wizard's question order is unchanged.
func ReconfigureQuestions(projectRoot string) []Question {
	base := DefaultQuestions(projectRoot)
	git := GitQuestions()

	// Splice point: immediately after report_format. Falling back to the end of
	// the base set keeps the order correct if the base set is ever reordered.
	splice := len(base)
	for i, q := range base {
		if q.ID == "report_format" {
			splice = i + 1
			break
		}
	}

	merged := make([]Question, 0, len(base)+len(git))
	merged = append(merged, base[:splice]...)
	merged = append(merged, git...)
	merged = append(merged, base[splice:]...)
	return merged
}

// InitQuestions returns the FULL `moai init` question set: pages 1-2
// (DefaultQuestions) followed by page 3 (Page3Questions). It is the single
// assembly point consumed by the wizard entry point, so the init set cannot
// drift from what the tests exercise.
//
// ReconfigureQuestions deliberately does NOT build on this: the page-3
// questions must not leak into `moai update --reconfigure` (AC-WIZ-012a).
func InitQuestions(projectRoot string) []Question {
	base := DefaultQuestions(projectRoot)
	page3 := Page3Questions(projectRoot)
	all := make([]Question, 0, len(base)+len(page3)+1)
	all = append(all, base...)
	all = append(all, page3...)
	// SPEC-AUTONOMY-TIERS-001 (REQ-001 / AC-001): the interactive autonomy-tier
	// selector page, appended after the Quality & Workflow page. The flag
	// (M4) validates the closed set fail-loud; this page PROMPTS the user.
	all = append(all, AutonomyTierQuestion())
	return all
}

// AutonomyTierQuestion returns the interactive autonomy-tier selector question
// (SPEC-AUTONOMY-TIERS-001 REQ-001 / AC-001). It offers the 3-tier closed set
// {semi-auto, automatic, fully-autonomous} with semi-auto pre-selected (REQ-006:
// no fully-autonomous default ships). The fully-autonomous tier is gated at
// apply time (init.go) via config.EffectiveTierWithGates — the sandbox-proof +
// kill-switch gating is NOT duplicated in the static option set; instead a
// fully-autonomous selection made without proof / under the kill-switch is
// downgraded to automatic with an advisory (AC-005).
func AutonomyTierQuestion() Question {
	return Question{
		ID:          "autonomy_tier",
		Group:       "Autonomy",
		Type:        QuestionTypeSelect,
		Title:       "Select autonomy tier",
		Description: "Controls how many permission prompts Claude Code shows. 'Semi-auto' is the safe default.",
		Options: []Option{
			{Label: "Semi-auto (Recommended)", Value: config.AutonomyTierSemiAuto, Desc: "Per-tool prompt — today's behavior"},
			{Label: "Automatic", Value: config.AutonomyTierAutomatic, Desc: "Per-tool auto-approval"},
			{Label: "Fully-autonomous", Value: config.AutonomyTierFullyAutonomous, Desc: "All prompts skipped (bypassPermissions); requires sandbox proof, gated by the kill-switch"},
		},
		Default:  config.AutonomyTierSemiAuto,
		Required: true,
	}
}

// FilteredQuestions returns questions filtered by their conditions.
// Questions whose conditions return false are excluded.
func FilteredQuestions(questions []Question, result *WizardResult) []Question {
	filtered := make([]Question, 0, len(questions))
	for _, q := range questions {
		if q.Condition == nil || q.Condition(result) {
			filtered = append(filtered, q)
		}
	}
	return filtered
}

// TotalVisibleQuestions counts questions that would be visible given current state.
func TotalVisibleQuestions(questions []Question, result *WizardResult) int {
	count := 0
	for _, q := range questions {
		if q.Condition == nil || q.Condition(result) {
			count++
		}
	}
	return count
}

// QuestionByID finds a question by its ID.
func QuestionByID(questions []Question, id string) *Question {
	for i := range questions {
		if questions[i].ID == id {
			return &questions[i]
		}
	}
	return nil
}

// Page3Questions returns page 3 of the `moai init` set, "Quality & Workflow".
//
// The four former confirm questions — lsp_enabled, enforce_quality,
// design_enabled, claude_design_enabled — are FIXED at their shipped true
// defaults and no longer asked (removed 2026-08-03). Their values are seeded
// by RunWithDefaults (see wizard.go), so interactive `moai init` writes the
// true default for each without prompting. project_mode and the worktree
// auto-create toggle remain interactive on this page.
//
// The constructor is named for the page it builds rather than for the retired
// mode taxonomy (REQ-WIZ-018): no flag selects it any more.
func Page3Questions(projectRoot string) []Question {
	return []Question{
		// B1 — project.mode.
		{
			ID:          "project_mode",
			Group:       "Quality & Workflow",
			Type:        QuestionTypeSelect,
			Title:       "Select project mode",
			Description: "Controls collaboration settings. 'personal' is the recommended default for solo developers.",
			Options: []Option{
				{Label: "Personal (Recommended)", Value: "personal", Desc: "Solo developer — no team coordination overhead"},
				{Label: "Team", Value: "team", Desc: "Multi-developer setup — enables team collaboration features"},
			},
			Default:  "personal",
			Required: true,
		},
		// Worktree auto-create (Issue 3). Persisted to
		// workflow.worktree.auto_create via the workflow seam. Default false
		// matches the config default (internal/config/defaults.go AutoCreate:
		// false). When enabled, `moai init` patches workflow.yaml so the
		// orchestrator auto-creates an L1 worktree for run-phase isolation.
		{
			ID:          "worktree_auto_create",
			Group:       "Quality & Workflow",
			Type:        QuestionTypeConfirm,
			Title:       "Enable worktree auto-creation?",
			Description: "When enabled, MoAI automatically creates an isolated git worktree for run-phase work. Default is off (Claude Code runtime handles L1 worktrees autonomously).",
			Default:     "false",
			Required:    false,
		},
		// SPEC-MOAI-MCP-SERVER-001 M4 (REQ-MCP-015 / AC-MCP-020) — the audit +
		// MCP opt-in selection. Grouped under "Audit & MCP" so they render as a
		// distinct form page. The enum Values reuse the M3 typed-config constants
		// (internal/config AuditModel* / AuditGate*) — the wizard and the audit
		// backend share ONE interpreter (no fork; AC-MCP-021 relies on the same
		// interpreter via the web console schema fields).
		//
		// audit_model — the active audit backend. `multi` is a declared token
		// only (convergence is a follow-up SPEC); it is offered for forward
		// compatibility and stored verbatim.
		{
			ID:          "audit_model",
			Group:       "Audit & MCP",
			Type:        QuestionTypeSelect,
			Title:       "Select audit model",
			Description: "The active review backend that gates merges. 'claude' uses the session model; 'codex'/'glm' add an external reviewer (fail-open if absent).",
			Options: []Option{
				{Label: "Claude (Recommended)", Value: config.AuditModelClaude, Desc: "Session model review — no external dependency"},
				{Label: "Codex", Value: config.AuditModelCodex, Desc: "codex CLI review (requires codex installed; fails open)"},
				{Label: "GLM", Value: config.AuditModelGLM, Desc: "z.ai GLM review (requires GLM key; fails open)"},
				{Label: "Multi", Value: config.AuditModelMulti, Desc: "Declare multi-auditor (convergence is a follow-up; stored only)"},
			},
			Default:  config.AuditModelClaude,
			Required: true,
		},
		// Per-auditor audit_gate. The distributed default gate is `required`
		// (§G.3); glm ships advisory in the locked default profile, applied at
		// the config-default layer (defaults.go), not at the prompt. Each gate
		// reuses the M3 AuditGate* constants.
		{
			ID:          "audit_gate_claude",
			Group:       "Audit & MCP",
			Type:        QuestionTypeSelect,
			Title:       "Claude audit gate",
			Description: "off = skip; advisory = warn only; required = block merge until PASS.",
			Options: []Option{
				{Label: "Required (Recommended)", Value: config.AuditGateRequired, Desc: "Block merge until Claude review PASS"},
				{Label: "Advisory", Value: config.AuditGateAdvisory, Desc: "Warn only — no block"},
				{Label: "Off", Value: config.AuditGateOff, Desc: "Skip the Claude auditor"},
			},
			Default:  config.AuditGateRequired,
			Required: true,
		},
		{
			ID:          "audit_gate_codex",
			Group:       "Audit & MCP",
			Type:        QuestionTypeSelect,
			Title:       "Codex audit gate",
			Description: "off = skip; advisory = warn only; required = block merge until PASS (fails open if codex absent).",
			Options: []Option{
				{Label: "Required (Recommended)", Value: config.AuditGateRequired, Desc: "Block merge until codex PASS (fail-open on missing codex)"},
				{Label: "Advisory", Value: config.AuditGateAdvisory, Desc: "Warn only — no block"},
				{Label: "Off", Value: config.AuditGateOff, Desc: "Skip the codex auditor"},
			},
			Default:  config.AuditGateRequired,
			Required: true,
		},
		{
			ID:          "audit_gate_glm",
			Group:       "Audit & MCP",
			Type:        QuestionTypeSelect,
			Title:       "GLM audit gate",
			Description: "off = skip; advisory = warn only; required = block merge until PASS (fails open if GLM key absent).",
			Options: []Option{
				{Label: "Required", Value: config.AuditGateRequired, Desc: "Block merge until GLM PASS (fail-open on missing key)"},
				{Label: "Advisory (Recommended)", Value: config.AuditGateAdvisory, Desc: "Warn only — no block; the locked default profile for glm (a distributed user with no GLM key is never hard-blocked)"},
				{Label: "Off", Value: config.AuditGateOff, Desc: "Skip the GLM auditor"},
			},
			Default:  config.AuditGateAdvisory,
			Required: true,
		},
		// codex_audit_enabled — master toggle for the codex backend + the
		// Stop-hook review gate. Default false (opt-in). When true, `moai init`
		// persists workflow.codex.review_gate.enabled=true so the M2 review-gate
		// hook activates.
		{
			ID:          "codex_audit_enabled",
			Group:       "Audit & MCP",
			Type:        QuestionTypeConfirm,
			Title:       "Enable codex review gate?",
			Description: "When enabled, MoAI activates the codex Stop-hook review gate (workflow.codex.review_gate.enabled). Default is off — the gate ships dormant.",
			Default:     "false",
			Required:    false,
		},
		// mcp_tools_opt_in — provisions the single neutral `.mcp.json` moai
		// entry (the §G.5 locked shape). Default false (C6 — opt-in default-off).
		// When true, `moai init` writes the entry via the M1 atomic-config seam.
		{
			ID:          "mcp_tools_opt_in",
			Group:       "Audit & MCP",
			Type:        QuestionTypeConfirm,
			Title:       "Provision the moai MCP server (.mcp.json)?",
			Description: "When enabled, MoAI writes the `moai mcp-server` stdio entry to .mcp.json (opt-in; a fresh project ships it inert). Default is off.",
			Default:     "false",
			Required:    false,
		},
	}
}
