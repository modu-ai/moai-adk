package wizard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultQuestions returns the standard set of questions for project initialization.
// The questions follow this order:
// 1. Project name (required)
// 2. Development mode
// 3. Git mode
// 4. Git provider (conditional)
// 5. GitLab instance URL (conditional)
// 6. GitHub username (conditional)
// 7. GitHub token (conditional)
// 8. GitLab username (conditional)
// 9. GitLab token (conditional)
func DefaultQuestions(projectRoot string) []Question {
	// Use current directory name as default project name
	defaultProjectName := filepath.Base(projectRoot)
	if defaultProjectName == "." || defaultProjectName == "/" || defaultProjectName == `\` {
		defaultProjectName = "my-project"
	}

	return []Question{
		// 1. Project Name
		{
			ID:          "project_name",
			Type:        QuestionTypeInput,
			Title:       "Enter project name",
			Description: "The name of your project.",
			Default:     defaultProjectName,
			Required:    true,
		},
		// 2. Model Policy
		{
			ID:          "model_policy",
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
		// 3. Development Mode
		{
			ID:          "development_mode",
			Type:        QuestionTypeSelect,
			Title:       "Select development methodology",
			Description: "Controls the development workflow cycle used during implementation.",
			Options: []Option{
				{Label: "TDD (Recommended)", Value: "tdd", Desc: "Test-Driven Development: RED-GREEN-REFACTOR"},
				{Label: "DDD", Value: "ddd", Desc: "Domain-Driven Development: ANALYZE-PRESERVE-IMPROVE"},
			},
			Default:  "tdd",
			Required: true,
		},
		// 3. Git Mode
		{
			ID:          "git_mode",
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
		// 4. Git Provider (conditional - only for personal/team modes)
		{
			ID:          "git_provider",
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
		// 5. GitLab Instance URL (conditional - only for gitlab provider)
		{
			ID:          "gitlab_instance_url",
			Type:        QuestionTypeInput,
			Title:       "Enter GitLab instance URL",
			Description: "For GitLab.com use https://gitlab.com. For self-hosted, enter your instance URL.",
			Default:     "https://gitlab.com",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return (r.GitMode == "personal" || r.GitMode == "team") && r.GitProvider == "gitlab"
			},
		},
		// 6. GitHub Username (conditional - only for github provider)
		{
			ID:          "github_username",
			Type:        QuestionTypeInput,
			Title:       "Enter your GitHub username",
			Description: "Required for Git automation features.",
			Default:     "",
			Required:    false, // Conditional requirement handled by wizard
			Condition: func(r *WizardResult) bool {
				return (r.GitMode == "personal" || r.GitMode == "team") && r.GitProvider == "github"
			},
		},
		// 7. GitHub Token (conditional - only for github provider)
		{
			ID:          "github_token",
			Type:        QuestionTypeInput,
			Title:       "Enter GitHub personal access token (optional)",
			Description: "Required for PR creation and pushing. Leave empty to skip or use gh CLI.",
			Default:     "",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return (r.GitMode == "personal" || r.GitMode == "team") && r.GitProvider == "github"
			},
		},
		// 8. GitLab Username (conditional - only for gitlab provider)
		{
			ID:          "gitlab_username",
			Type:        QuestionTypeInput,
			Title:       "Enter your GitLab username",
			Description: "Required for Git automation features with GitLab.",
			Default:     "",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return (r.GitMode == "personal" || r.GitMode == "team") && r.GitProvider == "gitlab"
			},
		},
		// 9. GitLab Token (conditional - only for gitlab provider)
		{
			ID:          "gitlab_token",
			Type:        QuestionTypeInput,
			Title:       "Enter GitLab personal access token (optional)",
			Description: "Required for MR creation and pushing. Leave empty to skip or use glab CLI.",
			Default:     "",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return (r.GitMode == "personal" || r.GitMode == "team") && r.GitProvider == "gitlab"
			},
		},
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
		name := e.Name()
		if strings.HasSuffix(name, ".md") {
			profiles = append(profiles, strings.TrimSuffix(name, ".md"))
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
