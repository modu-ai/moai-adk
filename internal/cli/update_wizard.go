package cli

// Init/reconfigure wizard half of the update command. Extracted verbatim from
// update.go by SPEC-CLIFIX-HYGIENE-001 M5 (mechanical move, no logic change)
// to bring update.go under the 1,200-line ceiling.

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config/atomicfile"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/tui"
	"github.com/modu-ai/moai-adk/pkg/version"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// runInitWizard runs the configuration wizard for reconfiguring an existing project.
// Used by 'moai update -c/--config' to edit project settings.
// @MX:NOTE: [AUTO] runInitWizard — M4-S4d-3 DDD migration. tui.Pill (uninitialized warning,
// cancelled, success) + tui.Section (Reconfiguration header; AC-017 emoji removed).
func runInitWizard(cmd *cobra.Command, reconfigure bool) error {
	out := cmd.OutOrStdout()
	th := resolveTheme()

	// Verify the project is initialized
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	if _, err := os.Stat(filepath.Join(cwd, defs.MoAIDir)); os.IsNotExist(err) {
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillWarn, Solid: false, Label: "Project not initialized · Run 'moai init' first", Theme: &th}))
		return fmt.Errorf("project not initialized")
	}

	// Print banner and welcome message
	uikit.PrintBanner(version.GetVersion())
	if reconfigure {
		_, _ = fmt.Fprintln(out, tui.Section("Project Reconfiguration Wizard", tui.SectionOpts{Theme: &th}))
		_, _ = fmt.Fprintln(out, "This wizard will help you update your project configuration.")
	} else {
		uikit.PrintWelcomeMessage()
	}

	// REQ-1: Read locale from language.yaml
	locale := wizard.ReadLocaleFromProject(cwd)

	// REQ-2: Read existing username from config (used as default value)
	existingGitHubUsername := wizard.ReadGitHubUsernameFromConfig(cwd)
	existingGitLabUsername := wizard.ReadGitLabUsernameFromConfig(cwd)

	// REQ-3: Check whether gh CLI is authenticated
	ghAuthenticated := wizard.IsGhAuthenticated()

	// Generate default questions and set defaults from existing values
	questions := wizard.ReconfigureQuestions(cwd)
	if existingGitHubUsername != "" {
		if q := wizard.QuestionByID(questions, "github_username"); q != nil {
			q.Default = existingGitHubUsername
		}
	}
	if existingGitLabUsername != "" {
		if q := wizard.QuestionByID(questions, "gitlab_username"); q != nil {
			q.Default = existingGitLabUsername
		}
	}
	// REQ-3: Skip github_token question when gh auth is authenticated
	if ghAuthenticated {
		if q := wizard.QuestionByID(questions, "github_token"); q != nil {
			q.Condition = func(_ *wizard.WizardResult) bool { return false }
		}
	}

	// Run wizard with locale and custom questions
	result, err := wizard.RunWithLocale(questions, nil, locale)
	if err != nil {
		if errors.Is(err, wizard.ErrCancelled) {
			_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillNeutral, Solid: false, Label: "Configuration cancelled", Theme: &th}))
			return nil
		}
		return fmt.Errorf("wizard failed: %w", err)
	}

	// Apply configuration updates to .moai/config/sections/
	// This updates the YAML configuration files based on wizard results,
	// then runs the interactive workflow-toggle step (chain ② link d).
	if err := applyWizardReconfigureSteps(out, cwd, result); err != nil {
		return fmt.Errorf("apply configuration: %w", err)
	}

	// F1 security fix: a token entered in the wizard is NOT persisted to any
	// git-tracked config. Tell the user so the credential is not silently
	// dropped — direct them to the provider CLI credential store.
	if result.GitHubToken != "" {
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillWarn, Solid: false, Label: "GitHub token was not saved (never stored in plaintext). Run 'gh auth login' to authenticate.", Theme: &th}))
	}
	if result.GitLabToken != "" {
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillWarn, Solid: false, Label: "GitLab token was not saved (never stored in plaintext). Run 'glab auth login' to authenticate.", Theme: &th}))
	}

	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: "Configuration updated successfully", Theme: &th}))

	return nil
}

// applyWizardReconfigureSteps applies the reconfigure wizard's config updates
// and then runs the workflow-toggle step (SPEC-INIT-WIZARD-REPAIR-001 REQ-007
// / SPEC-WT-DOC-001 Surface 3). The workflow step is a no-op when stdin is not
// a TTY or workflow.yaml is absent, so CI / non-interactive reconfigures are
// unchanged. Extracted from runInitWizard so the post-wizard composition is
// testable without driving the TUI wizard.
//
// @MX:SPEC: SPEC-INIT-WIZARD-REPAIR-001
func applyWizardReconfigureSteps(out io.Writer, cwd string, result *wizard.WizardResult) error {
	if err := applyWizardConfig(cwd, result); err != nil {
		return fmt.Errorf("apply configuration: %w", err)
	}
	if err := runWorkflowConfigStep(out, cwd); err != nil {
		return fmt.Errorf("workflow config step: %w", err)
	}
	return nil
}

// applyWizardConfig applies wizard results to the project configuration files.
func applyWizardConfig(projectRoot string, result *wizard.WizardResult) error {
	// F3: validate externally-supplied identity fields before any write so a
	// malformed URL or username never lands in a persisted config file.
	if err := validateWizardInput(result); err != nil {
		return err
	}

	sectionsDir := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir)

	// user.yaml: Save GitHub/GitLab username plus the wizard-collected user
	// display name. Access tokens are DELIBERATELY NOT persisted here (F1
	// security fix): user.yaml is git-tracked and world-readable, so writing a
	// secret into it leaks the credential. Credentials are delegated to the
	// `gh` / `glab` CLI (see the advisory surfaced by runInitWizard). The block
	// triggers on any non-secret user field (username or display name), so a
	// reconfigure that answers only the user_name question still persists
	// user.name — but a token-only answer creates nothing.
	hasUserFields := result.GitHubUsername != "" ||
		result.GitLabUsername != "" ||
		result.UserName != ""
	if hasUserFields {
		userPath := filepath.Join(sectionsDir, defs.UserYAML)
		// Read existing file
		userData, err := os.ReadFile(userPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read user.yaml: %w", err)
		}

		// Parse YAML
		var user map[string]any
		if len(userData) > 0 {
			if err := yaml.Unmarshal(userData, &user); err != nil {
				return fmt.Errorf("parse user.yaml: %w", err)
			}
		} else {
			user = make(map[string]any)
		}

		// Ensure user.user section exists
		var userConfig map[string]any
		if existingUser, ok := user["user"].(map[string]any); ok {
			userConfig = existingUser
		} else {
			userConfig = make(map[string]any)
		}

		// Save GitHub/GitLab usernames (non-secret). Access tokens
		// (result.GitHubToken / result.GitLabToken) are intentionally NOT
		// written — persisting a secret into the git-tracked user.yaml is the
		// F1 vulnerability this fix closes. Tokens are delegated to the
		// `gh` / `glab` CLI credential store instead.
		if result.GitHubUsername != "" {
			userConfig["github_username"] = result.GitHubUsername
		}
		if result.GitLabUsername != "" {
			userConfig["gitlab_username"] = result.GitLabUsername
		}

		// Save user display name (reconfigure wizard user_name question)
		if result.UserName != "" {
			userConfig["name"] = result.UserName
		}

		user["user"] = userConfig

		// Save to file
		updatedData, err := yaml.Marshal(user)
		if err != nil {
			return fmt.Errorf("marshal user.yaml: %w", err)
		}
		if err := atomicfile.Write(userPath, updatedData, defs.FilePerm); err != nil {
			return fmt.Errorf("write user.yaml: %w", err)
		}
	}

	// language.yaml: Save conversation language (reconfigure wizard language
	// question). Sets both conversation_language and conversation_language_name
	// while preserving all sibling keys (agent_prompt_language, code_comments, ...).
	if result.ConversationLang != "" {
		langPath := filepath.Join(sectionsDir, defs.LanguageYAML)
		langData, err := os.ReadFile(langPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read language.yaml: %w", err)
		}

		var lang map[string]any
		if len(langData) > 0 {
			if err := yaml.Unmarshal(langData, &lang); err != nil {
				return fmt.Errorf("parse language.yaml: %w", err)
			}
		} else {
			lang = make(map[string]any)
		}

		// Ensure language section exists
		var language map[string]any
		if existing, ok := lang["language"].(map[string]any); ok {
			language = existing
		} else {
			language = make(map[string]any)
		}

		language["conversation_language"] = result.ConversationLang
		language["conversation_language_name"] = result.ConversationLang
		lang["language"] = language

		updatedData, err := yaml.Marshal(lang)
		if err != nil {
			return fmt.Errorf("marshal language.yaml: %w", err)
		}
		if err := atomicfile.Write(langPath, updatedData, defs.FilePerm); err != nil {
			return fmt.Errorf("write language.yaml: %w", err)
		}
	}

	// git-strategy.yaml: Save git mode and provider (REQ-4)
	if result.GitMode != "" || result.GitProvider != "" {
		gitStratPath := filepath.Join(sectionsDir, defs.GitStrategyYAML)
		gsData, err := os.ReadFile(gitStratPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read git-strategy.yaml: %w", err)
		}

		var gs map[string]any
		if len(gsData) > 0 {
			if err := yaml.Unmarshal(gsData, &gs); err != nil {
				return fmt.Errorf("parse git-strategy.yaml: %w", err)
			}
		} else {
			gs = make(map[string]any)
		}

		// Ensure git_strategy section exists
		var gitStrategy map[string]any
		if existing, ok := gs["git_strategy"].(map[string]any); ok {
			gitStrategy = existing
		} else {
			gitStrategy = make(map[string]any)
		}

		if result.GitMode != "" {
			gitStrategy["mode"] = result.GitMode
		}
		if result.GitProvider != "" {
			gitStrategy["provider"] = result.GitProvider
		}

		// Save GitLab instance URL (REQ-5)
		if result.GitLabInstanceURL != "" {
			var gitlabSection map[string]any
			if existing, ok := gitStrategy["gitlab"].(map[string]any); ok {
				gitlabSection = existing
			} else {
				gitlabSection = make(map[string]any)
			}
			gitlabSection["instance_url"] = result.GitLabInstanceURL
			gitStrategy["gitlab"] = gitlabSection
		}

		gs["git_strategy"] = gitStrategy

		updatedData, err := yaml.Marshal(gs)
		if err != nil {
			return fmt.Errorf("marshal git-strategy.yaml: %w", err)
		}
		if err := atomicfile.Write(gitStratPath, updatedData, defs.FilePerm); err != nil {
			return fmt.Errorf("write git-strategy.yaml: %w", err)
		}
	}

	// SPEC-MODEL-PROFILE-MATRIX-001 (REQ-MPM-016/024): the former plan_type × tier
	// agent-frontmatter mutation is RETIRED — persist the resolved model policy to
	// llm.profile (normalized {high,medium,low} → {max,medium,low}) instead of
	// mutating agent frontmatter, which stays at model: inherit.
	if result.ModelPolicy != "" {
		policy := template.ModelPolicy(result.ModelPolicy)
		if template.IsValidModelPolicy(string(policy)) {
			if err := template.ApplyProfile(projectRoot, template.NormalizeToTier(result.ModelPolicy)); err != nil {
				return fmt.Errorf("apply profile: %w", err)
			}
			// Persist model_policy to system.yaml so it survives future updates
			systemPath := filepath.Join(sectionsDir, defs.SystemYAML)
			systemData, _ := os.ReadFile(systemPath)
			var sys map[string]any
			if len(systemData) > 0 {
				_ = yaml.Unmarshal(systemData, &sys)
			}
			if sys == nil {
				sys = make(map[string]any)
			}
			moaiSection, _ := sys["moai"].(map[string]any)
			if moaiSection == nil {
				moaiSection = make(map[string]any)
			}
			moaiSection["model_policy"] = string(policy)
			sys["moai"] = moaiSection
			if updatedData, err := yaml.Marshal(sys); err == nil {
				_ = atomicfile.Write(systemPath, updatedData, defs.FilePerm)
			}
		}
	}

	// quality.yaml: Save development mode (REQ-4)
	if result.DevelopmentMode != "" {
		qualityPath := filepath.Join(sectionsDir, defs.QualityYAML)
		qualityData, err := os.ReadFile(qualityPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read quality.yaml: %w", err)
		}

		var quality map[string]any
		if len(qualityData) > 0 {
			if err := yaml.Unmarshal(qualityData, &quality); err != nil {
				return fmt.Errorf("parse quality.yaml: %w", err)
			}
		} else {
			quality = make(map[string]any)
		}

		// Ensure constitution section exists
		var constitution map[string]any
		if existing, ok := quality["constitution"].(map[string]any); ok {
			constitution = existing
		} else {
			constitution = make(map[string]any)
		}

		constitution["development_mode"] = result.DevelopmentMode
		quality["constitution"] = constitution

		updatedData, err := yaml.Marshal(quality)
		if err != nil {
			return fmt.Errorf("marshal quality.yaml: %w", err)
		}
		if err := atomicfile.Write(qualityPath, updatedData, defs.FilePerm); err != nil {
			return fmt.Errorf("write quality.yaml: %w", err)
		}
	}

	return nil
}
