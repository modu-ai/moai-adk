package wizard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	all := make([]Question, 0, len(base)+len(page3))
	all = append(all, base...)
	all = append(all, page3...)
	return all
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

// canonicalHarnessProfiles is the fallback list used when the evaluator-profiles
// directory is absent or empty (AC-IWE-002 EC-2).
var canonicalHarnessProfiles = []string{"default", "strict", "lenient", "frontend"}

// loadHarnessProfiles reads profile filenames from projectRoot/.moai/config/evaluator-profiles/.
// Falls back to canonicalHarnessProfiles when the directory is absent or empty.
// A warning is printed to stderr when the fallback is triggered (EC-2).
func loadHarnessProfiles(projectRoot string) []Option {
	dir := filepath.Join(projectRoot, ".moai", "config", "evaluator-profiles")
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr,
				"warning: evaluator-profiles directory not found at %s; using canonical fallback list [default, strict, lenient, frontend]\n",
				dir)
		}
		return harnessProfileOptions(canonicalHarnessProfiles)
	}

	var profiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if base, ok := strings.CutSuffix(e.Name(), ".md"); ok {
			profiles = append(profiles, base)
		}
	}

	if len(profiles) == 0 {
		_, _ = fmt.Fprintf(os.Stderr,
			"warning: no .md profile files found in %s; using canonical fallback list\n", dir)
		return harnessProfileOptions(canonicalHarnessProfiles)
	}

	return harnessProfileOptions(profiles)
}

func harnessProfileOptions(profiles []string) []Option {
	opts := make([]Option, 0, len(profiles))
	for i, p := range profiles {
		desc := ""
		switch p {
		case "default":
			desc = "Standard quality scoring"
		case "strict":
			desc = "Stricter thresholds — fewer false PASS"
		case "lenient":
			desc = "Relaxed thresholds — faster iteration"
		case "frontend":
			desc = "Frontend-optimised scoring dimensions"
		}
		label := p
		if i == 0 {
			label = p + " (Recommended)"
		}
		opts = append(opts, Option{Label: label, Value: p, Desc: desc})
	}
	return opts
}

// Page3Questions returns page 3 of the `moai init` set, "Quality & Workflow".
//
// The questions are UNCONDITIONAL: the former mode gate was removed
// (REQ-WIZ-001/002), so every user sees them and they merge into one page. The
// single exception is claude_design_enabled, which stays nested on
// design_enabled (REQ-WIZ-006) and therefore renders as its own sub-group.
//
// The constructor is named for the page it builds rather than for the retired
// mode taxonomy (REQ-WIZ-018): no flag selects it any more.
//
// Order is load-bearing: design_enabled MUST precede claude_design_enabled so
// huh has the design answer before evaluating the nested hide func.
func Page3Questions(projectRoot string) []Question {
	return []Question{
		// B3 — lsp.enabled. Enabled by default since
		// SPEC-CLI-WIZARD-RESTRUCTURE-001 (REQ-WIZ-010): the diagnostics are
		// worth more than the startup cost, and opting out is one keystroke.
		{
			ID:          "lsp_enabled",
			Group:       "Quality & Workflow",
			Type:        QuestionTypeConfirm,
			Title:       "Enable LSP integration? (default: Yes)",
			Description: "LSP provides language-server diagnostics during the run phase. Enabled by default; answer No to opt out.",
			Default:     "true",
			Required:    false,
		},
		// B5 — quality.enforce_quality
		{
			ID:          "enforce_quality",
			Group:       "Quality & Workflow",
			Type:        QuestionTypeConfirm,
			Title:       "Enforce quality gates? (default: Yes)",
			Description: "When enabled, TRUST 5 quality gates block implementation progress on failure.",
			Default:     "true",
			Required:    false,
		},
		// B1 — project.mode
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
		// B8 — design.enabled
		{
			ID:          "design_enabled",
			Group:       "Quality & Workflow",
			Type:        QuestionTypeConfirm,
			Title:       "Enable design workflow? (default: Yes)",
			Description: "Enables the MoAI design pipeline (GAN loop, brand context, Claude Design integration).",
			Default:     "true",
			Required:    false,
		},
		// B8 — design.claude_design.enabled. The ONLY conditional question on
		// page 3: nested on design_enabled (REQ-WIZ-006). The mode half of the
		// former two-term predicate is gone — only DesignEnabled remains.
		{
			ID:          "claude_design_enabled",
			Group:       "Quality & Workflow",
			Type:        QuestionTypeConfirm,
			Title:       "Enable Claude Design integration? (default: Yes)",
			Description: "Enables the Claude Design handoff workflow within the design pipeline.",
			Default:     "true",
			Required:    false,
			Condition:   func(r *WizardResult) bool { return r.DesignEnabled },
		},
	}
}
