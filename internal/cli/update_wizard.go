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
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
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
	// This updates the YAML configuration files based on wizard results
	if err := applyWizardConfig(cwd, result); err != nil {
		return fmt.Errorf("apply configuration: %w", err)
	}

	// SPEC-WT-DOC-001 (branch-guard config surface): the four workflow toggles
	// are not part of the wizard question set — they ride a dedicated huh form
	// so the existing ReconfigureQuestions order/length (locked by tests in
	// internal/cli/wizard) stays stable. The step reads workflow.yaml for
	// current defaults, prompts once, then writes back via the comment-
	// preserving yamlpatch seam. A TTY is required (huh falls back to no-op on
	// non-terminal stdin, matching the wizard's own behaviour).
	if err := runWorkflowConfigStep(out, cwd); err != nil {
		return fmt.Errorf("workflow settings step: %w", err)
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
		if err := os.WriteFile(userPath, updatedData, defs.FilePerm); err != nil {
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
		if err := os.WriteFile(langPath, updatedData, defs.FilePerm); err != nil {
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
		if err := os.WriteFile(gitStratPath, updatedData, defs.FilePerm); err != nil {
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
				_ = os.WriteFile(systemPath, updatedData, defs.FilePerm)
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
		if err := os.WriteFile(qualityPath, updatedData, defs.FilePerm); err != nil {
			return fmt.Errorf("write quality.yaml: %w", err)
		}
	}

	return nil
}

// runWorkflowConfigStep is the SPEC-WT-DOC-001 reconfigure-wizard surface for
// the four workflow toggles that do not fit the existing wizard question set
// (workflow.branch_guard.enabled + the three workflow.worktree.auto_* keys).
// It reads the current values from workflow.yaml, prompts once via a themed
// huh form, then writes back through the comment-preserving yamlpatch seam.
//
// The step is a no-op when stdin is not a TTY (CI / non-interactive reconfigure)
// or when workflow.yaml is absent — the latter would only occur on a corrupted
// project, and silently skipping keeps the reconfigure path idempotent. The
// interactive surface is split from applyWorkflowConfigEdits so callers and
// tests can exercise the write path directly without driving the TUI.
func runWorkflowConfigStep(out io.Writer, projectRoot string) error {
	workflowPath := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir, defs.WorkflowYAML)
	if _, err := os.Stat(workflowPath); err != nil {
		// Absent workflow.yaml — nothing to reconfigure. Not an error.
		return nil
	}
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		// Non-interactive context (CI, piped stdin) — skip the prompt, leave
		// the deployed defaults untouched. Mirrors the wizard's own gating.
		return nil
	}

	curBranch, curCreate, curMerge, curCleanup, err := readWorkflowToggleDefaults(workflowPath)
	if err != nil {
		return fmt.Errorf("read workflow.yaml: %w", err)
	}

	got, err := promptWorkflowToggles(curBranch, curCreate, curMerge, curCleanup)
	if err != nil {
		return err
	}

	edits := buildWorkflowToggleEdits(curBranch, curCreate, curMerge, curCleanup,
		got.branchGuard, got.autoCreate, got.autoMerge, got.autoCleanup)
	if len(edits) == 0 {
		// User accepted every default — nothing to persist.
		return nil
	}
	if err := yamlpatch.PatchFile(workflowPath, edits); err != nil {
		return fmt.Errorf("patch workflow.yaml: %w", err)
	}
	// Pill write failure is non-fatal — discard the error so it never aborts the wizard.
	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{
		Kind:  tui.PillOk,
		Solid: false,
		Label: "Workflow settings updated",
		Theme: ptrThemeOrDefault(),
	}))
	return nil
}

// workflowToggleValues carries the four bool toggles read from or written to
// workflow.yaml. Keeping the four values together avoids a multi-return sprawl
// across the helper chain.
type workflowToggleValues struct {
	branchGuard bool
	autoCreate  bool
	autoMerge   bool
	autoCleanup bool
}

// readWorkflowToggleDefaults parses workflow.yaml and extracts the current
// values of the four reconfigurable toggles. Missing keys default to false
// (matching the distributed template: branch_guard absent, worktree.auto_create
// false; auto_merge/auto_cleanup ship as true in the template but the defaults
// here are only the parse-failure fallback — the deployed values are read
// verbatim from disk when present).
func readWorkflowToggleDefaults(path string) (branchGuard, autoCreate, autoMerge, autoCleanup bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, false, false, false, err
	}
	var doc struct {
		Workflow struct {
			BranchGuard struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"branch_guard"`
			Worktree struct {
				AutoCreate  bool `yaml:"auto_create"`
				AutoMerge   bool `yaml:"auto_merge"`
				AutoCleanup bool `yaml:"auto_cleanup"`
			} `yaml:"worktree"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return false, false, false, false, err
	}
	return doc.Workflow.BranchGuard.Enabled,
		doc.Workflow.Worktree.AutoCreate,
		doc.Workflow.Worktree.AutoMerge,
		doc.Workflow.Worktree.AutoCleanup,
		nil
}

// promptWorkflowToggles presents a single-page huh form carrying the four
// workflow toggles. The current values seed each confirm as the default.
func promptWorkflowToggles(curBranch, curCreate, curMerge, curCleanup bool) (workflowToggleValues, error) {
	v := workflowToggleValues{
		branchGuard: curBranch,
		autoCreate:  curCreate,
		autoMerge:   curMerge,
		autoCleanup: curCleanup,
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable branch guard?").
				Description("Blocks branch-state changes (switch/reset/rebase) in the shared primary checkout.").
				Value(&v.branchGuard),
			huh.NewConfirm().
				Title("Auto-create worktree per task?").
				Description("L1 worktree creation when starting a SPEC run phase (default: false).").
				Value(&v.autoCreate),
			huh.NewConfirm().
				Title("Auto-merge worktree branch on completion?").
				Description("Merge the worktree branch back to its base when the SPEC completes.").
				Value(&v.autoMerge),
			huh.NewConfirm().
				Title("Auto-cleanup worktree after merge?").
				Description("Remove the worktree directory after a successful merge.").
				Value(&v.autoCleanup),
		),
	).WithTheme(moaiHuhTheme())
	if err := form.Run(); err != nil {
		return workflowToggleValues{}, fmt.Errorf("workflow form: %w", err)
	}
	return v, nil
}

// buildWorkflowToggleEdits diffs the (current, chosen) bool pairs and produces
// a KeyEdit slice carrying only the keys whose value the user changed. An
// unchanged key is NOT included so yamlpatch leaves the deployed byte sequence
// (and any comments) untouched. branch_guard.enabled is always included when
// the user toggled it, even if the key was absent on disk — yamlpatch upserts
// the nested mapping on first write.
func buildWorkflowToggleEdits(curBranch, curCreate, curMerge, curCleanup,
	newBranch, newCreate, newMerge, newCleanup bool) []yamlpatch.KeyEdit {
	var edits []yamlpatch.KeyEdit
	if newBranch != curBranch {
		edits = append(edits, yamlpatch.KeyEdit{
			Path:  []string{"workflow", "branch_guard", "enabled"},
			Value: strconv.FormatBool(newBranch),
		})
	}
	if newCreate != curCreate {
		edits = append(edits, yamlpatch.KeyEdit{
			Path:  []string{"workflow", "worktree", "auto_create"},
			Value: strconv.FormatBool(newCreate),
		})
	}
	if newMerge != curMerge {
		edits = append(edits, yamlpatch.KeyEdit{
			Path:  []string{"workflow", "worktree", "auto_merge"},
			Value: strconv.FormatBool(newMerge),
		})
	}
	if newCleanup != curCleanup {
		edits = append(edits, yamlpatch.KeyEdit{
			Path:  []string{"workflow", "worktree", "auto_cleanup"},
			Value: strconv.FormatBool(newCleanup),
		})
	}
	return edits
}

// ptrThemeOrDefault returns the resolved theme for tui.Pill banners without
// forcing every call site to import tui itself. Falls back to the zero theme
// when the resolver is unavailable; the Pill helper handles a zero-value
// Theme by rendering plain text.
func ptrThemeOrDefault() *tui.Theme {
	th := resolveTheme()
	return &th
}
