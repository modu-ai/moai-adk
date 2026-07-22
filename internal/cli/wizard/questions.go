package wizard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// The wizard question set is split across three constructors:
//
//   - DefaultQuestions      — the `moai init` set (NO Git questions)
//   - GitQuestions          — the 7 Git questions, on their own
//   - ReconfigureQuestions  — DefaultQuestions with GitQuestions spliced back in,
//     used by the `moai update --reconfigure` path
//
// DefaultQuestions returns the question set asked during project
// initialization, in this order:
//  1. Conversation language (drives the rendering language of every later question)
//  2. User name (optional)
//  3. Project name (required)
//  4. Model policy
//  5. Report format
//  6. Advanced-settings bridge (conditional — hidden when StandardMode is preset by flag)
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
			Group:       "Language",
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
			Group:       "Identity",
			Type:        QuestionTypeInput,
			Title:       "Enter your name",
			Description: "How MoAI addresses you. Persisted to user.yaml (user.name). Leave empty to skip.",
			Default:     "",
			Required:    false,
		},
		// 3. Project Name
		{
			ID:          "project_name",
			Group:       "Project",
			Type:        QuestionTypeInput,
			Title:       "Enter project name",
			Description: "The name of your project.",
			Default:     defaultProjectName,
			Required:    true,
		},
		// 2. Model Policy
		{
			ID:          "model_policy",
			Group:       "Project",
			Type:        QuestionTypeSelect,
			Title:       "Select model policy",
			Description: "Controls which Claude model tier is assigned to each agent. Match to your Claude plan.",
			Options: []Option{
				{Label: "High (Recommended)", Value: "high", Desc: "Opus for critical agents — Max $200 plan"},
				{Label: "Medium", Value: "medium", Desc: "Opus for key agents, sonnet for rest — Max $100 plan"},
				{Label: "Low", Value: "low", Desc: "Sonnet and haiku only — Plus $20 plan"},
			},
			Default:  "high",
			Required: true,
		},
		// 3. Report Format — html+md vs md.
		// The value set mirrors internal/settings reportFormatValues (the closed
		// set {"html+md", "md"} consumed by the moai-domain-html-report skill via
		// report.format). Keep these two Values in sync with that SSOT.
		{
			ID:          "report_format",
			Group:       "Project",
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
		// 6. Advanced-settings bridge (Quick mode only). Answering Yes flips
		// StandardMode on, which reveals the Phase 1 questions (gated on
		// r.StandardMode) in the same wizard run. Hidden when --standard/--advanced
		// already preset StandardMode (Condition returns false), so the flag path
		// never double-asks.
		{
			ID:          "advanced_bridge",
			Group:       "Options",
			Type:        QuestionTypeConfirm,
			Title:       "Configure advanced settings? (default: No)",
			Description: "Reveals project mode, harness profile, LSP, quality gates, and design options in this run.",
			Default:     "false",
			Required:    false,
			Condition:   func(r *WizardResult) bool { return !r.StandardMode },
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
// RunWithDefaultsModes builds DefaultQuestions (+ Phase1/Phase2), which contain
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
// back in at their original position — after report_format, before
// advanced_bridge — so the reconfigure wizard's question order is unchanged.
func ReconfigureQuestions(projectRoot string) []Question {
	base := DefaultQuestions(projectRoot)
	git := GitQuestions()

	// Splice point: immediately after report_format. Falling back to
	// "before the trailing advanced_bridge" keeps the order correct if the
	// base set is ever reordered.
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

// Phase1Questions returns the additional Phase 1 questions exposed by --standard/--advanced.
// Each question is gated on r.StandardMode == true so Quick mode is unaffected.
func Phase1Questions(projectRoot string) []Question {
	return []Question{
		// B1 — project.mode
		{
			ID:          "project_mode",
			Group:       "Options",
			Type:        QuestionTypeSelect,
			Title:       "Select project mode",
			Description: "Controls collaboration settings. 'personal' is the recommended default for solo developers.",
			Options: []Option{
				{Label: "Personal (Recommended)", Value: "personal", Desc: "Solo developer — no team coordination overhead"},
				{Label: "Team", Value: "team", Desc: "Multi-developer setup — enables team collaboration features"},
			},
			Default:   "personal",
			Required:  true,
			Condition: func(r *WizardResult) bool { return r.StandardMode },
		},
		// B2 — harness.default_profile (dynamic enumeration)
		{
			ID:          "harness_profile",
			Group:       "Options",
			Type:        QuestionTypeSelect,
			Title:       "Select default harness evaluator profile",
			Description: "Controls quality scoring depth. Profiles are loaded from .moai/config/evaluator-profiles/.",
			Options:     loadHarnessProfiles(projectRoot),
			Default:     "default",
			Required:    true,
			Condition:   func(r *WizardResult) bool { return r.StandardMode },
		},
		// B3 — lsp.enabled
		{
			ID:          "lsp_enabled",
			Group:       "Options",
			Type:        QuestionTypeConfirm,
			Title:       "Enable LSP integration? (default: No)",
			Description: "LSP provides language-server diagnostics during the run phase. Default is off (opt-in).",
			Default:     "false",
			Required:    false,
			Condition:   func(r *WizardResult) bool { return r.StandardMode },
		},
		// B5 — quality.enforce_quality
		{
			ID:          "enforce_quality",
			Group:       "Options",
			Type:        QuestionTypeConfirm,
			Title:       "Enforce quality gates? (default: Yes)",
			Description: "When enabled, TRUST 5 quality gates block implementation progress on failure.",
			Default:     "true",
			Required:    false,
			Condition:   func(r *WizardResult) bool { return r.StandardMode },
		},
		// B5 — quality.coverage_exemptions.enabled
		{
			ID:          "coverage_exemptions_enabled",
			Group:       "Options",
			Type:        QuestionTypeConfirm,
			Title:       "Allow coverage exemptions? (default: No)",
			Description: "Permits specific files or packages to be excluded from the coverage target.",
			Default:     "false",
			Required:    false,
			Condition:   func(r *WizardResult) bool { return r.StandardMode },
		},
		// B8 — design.enabled
		{
			ID:          "design_enabled",
			Group:       "Options",
			Type:        QuestionTypeConfirm,
			Title:       "Enable design workflow? (default: Yes)",
			Description: "Enables the MoAI design pipeline (GAN loop, brand context, Claude Design integration).",
			Default:     "true",
			Required:    false,
			Condition:   func(r *WizardResult) bool { return r.StandardMode },
		},
		// B8 — design.claude_design.enabled (conditional on design_enabled=true)
		{
			ID:          "claude_design_enabled",
			Group:       "Options",
			Type:        QuestionTypeConfirm,
			Title:       "Enable Claude Design integration? (default: Yes)",
			Description: "Enables the Claude Design handoff workflow within the design pipeline.",
			Default:     "true",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return r.StandardMode && r.DesignEnabled
			},
		},
	}
}
