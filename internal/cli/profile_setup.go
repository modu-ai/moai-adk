package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/pkg/models"
	"github.com/spf13/cobra"
)

// statuslineThemeCanonical holds the canonical theme values offered by the
// wizard — must stay in sync with the huh.NewOption block in the Display group.
// The statuslineModeCanonical + statuslinePresetCanonical lists and their
// helpers were removed by SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001 (runtime mode
// was inert; named presets were redundant with the segment map).
// KEEP IN SYNC with huh.NewOption block in the Display group.
var statuslineThemeCanonical = []string{defaultStatuslineTheme, "catppuccin-latte"}

// Wizard default constants.
const (
	defaultStatuslineTheme = "catppuccin-mocha"
	defaultPermissionMode  = "acceptEdits"
)

// isCanonicalStatuslineTheme reports whether s is a canonical value in the wizard option slice.
func isCanonicalStatuslineTheme(s string) bool {
	for _, v := range statuslineThemeCanonical {
		if v == s {
			return true
		}
	}
	return false
}

// normalizeStatuslineTheme returns a theme name compatible with wizard options.
// Legacy values such as "default" are converted to "catppuccin-mocha" so that the Select
// widget highlights the correct item.
func normalizeStatuslineTheme(theme string) string {
	if isCanonicalStatuslineTheme(theme) {
		return theme
	}
	return defaultStatuslineTheme
}

// statuslineAllSegments lists the 15 canonical segment keys that the MultiSelect
// widget offers. Order MUST match statusline.yaml segment definitions and
// Segment* fields in profileSetupText. The segment step is now unconditional
// (SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001 retired the preset==custom gate).
var statuslineAllSegments = []string{
	"claude_version", "context", "directory", "effort_thinking",
	"git_branch", "git_status", "moai_version", "model",
	"output_style", "pr", "session_time", "task",
	"usage_5h", "usage_7d", "worktree",
}

// @MX:NOTE: [AUTO] Wizard v3 migration — normalizes deprecated Claude model IDs to canonical aliases.
// @MX:REASON: Prevents silent loss of existing prefs values in huh.Select bindings after the "claude-opus-4-7" option was removed from the previous wizard.
func normalizeModel(m string) string {
	switch m {
	// canonical aliases pass through unchanged
	case "", "opus", "opus[1m]", "sonnet", "sonnet[1m]", "haiku", "opusplan":
		return m
	// deprecated full-ID → canonical alias
	case "claude-opus-4-7", "claude-opus-4-6":
		return "opus"
	case "claude-opus-4-7[1m]", "claude-opus-4-6[1m]", "claude-opus-4-6 1M":
		return "opus[1m]"
	case "claude-sonnet-4-6":
		return "sonnet"
	case "claude-sonnet-4-6[1m]", "claude-sonnet-4-6 1M":
		return "sonnet[1m]"
	case "claude-haiku-4-5":
		return "haiku"
	default:
		// Unknown values are reset to the runtime default.
		return ""
	}
}

// readCurrentProjectConfig reads the current development_mode + git_convention
// values from the project config (quality.yaml / git-convention.yaml) via the
// config manager. SPEC-WEB-CONSOLE-003 — these are project-config values, NOT
// ProfilePreferences fields, so the wizard initializes their selects from here
// rather than from existingPrefs. An absent config dir yields LoadRaw defaults.
func readCurrentProjectConfig(projectRoot string) (devMode, convention string, err error) {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return "", "", fmt.Errorf("read project config: %w", err)
	}
	return string(cfg.Quality.DevelopmentMode), cfg.GitConvention.Convention, nil
}

// persistProjectConfig writes the selected development_mode + git_convention
// values into the project config via the config-manager API (LoadRaw → mutate
// only non-empty → SetSection → Save). It writes ONLY the quality
// (development_mode) and git_convention (convention) sections; every other
// section round-trips unchanged. Empty values keep the existing persisted value
// (EC-1). This is the TUI counterpart to the web layer's writeProjectConfig —
// same canonical persistence path, no direct yaml.Marshal/os.WriteFile.
// SPEC-WEB-CONSOLE-003 REQ-WC3-006/007.
func persistProjectConfig(projectRoot, devMode, convention string) error {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return fmt.Errorf("load project config: %w", err)
	}

	changed := false

	if devMode != "" && string(cfg.Quality.DevelopmentMode) != devMode {
		quality := cfg.Quality
		quality.DevelopmentMode = models.DevelopmentMode(devMode)
		if err := mgr.SetSection("quality", quality); err != nil {
			return fmt.Errorf("set quality section: %w", err)
		}
		changed = true
	}

	if convention != "" && cfg.GitConvention.Convention != convention {
		gc := cfg.GitConvention
		gc.Convention = convention
		if err := mgr.SetSection("git_convention", gc); err != nil {
			return fmt.Errorf("set git_convention section: %w", err)
		}
		changed = true
	}

	if changed {
		if err := mgr.Save(); err != nil {
			return fmt.Errorf("save project config: %w", err)
		}
	}
	return nil
}

var profileSetupCmd = &cobra.Command{
	Use:   "setup [name]",
	Short: "Interactive setup wizard for profile preferences",
	Long: `Configure per-profile preferences through an interactive wizard.

Settings are stored in:
  ~/.moai/claude-profiles/<name>/preferences.yaml  (identity, language, model, display)

Examples:
  moai profile setup          # Configure default profile
  moai profile setup work     # Configure 'work' profile`,
	Args: cobra.MaximumNArgs(1),
	RunE: runProfileSetup,
}

func init() {
	profileCmd.AddCommand(profileSetupCmd)
}

// runProfileSetup runs the interactive profile configuration wizard.
// The first question is language selection; all subsequent UI text is displayed in the chosen language.
func runProfileSetup(cmd *cobra.Command, args []string) error {
	profileName := "default"
	if len(args) > 0 {
		profileName = args[0]
	}

	// Load existing preferences as defaults.
	existingPrefs, err := profile.ReadPreferences(profileName)
	if err != nil {
		return fmt.Errorf("read existing preferences: %w", err)
	}

	// Initialize form values from existing preferences.
	userName := existingPrefs.UserName

	convLang := existingPrefs.ConversationLang
	if convLang == "" {
		convLang = "en"
	}
	gitCommitLang := existingPrefs.GitCommitLang
	if gitCommitLang == "" {
		gitCommitLang = "en"
	}
	codeCommentLang := existingPrefs.CodeCommentLang
	if codeCommentLang == "" {
		codeCommentLang = "en"
	}
	docLang := existingPrefs.DocLang
	if docLang == "" {
		docLang = "en"
	}

	// C-1: normalize deprecated model IDs to canonical aliases
	model := normalizeModel(existingPrefs.Model)
	effortLevel := existingPrefs.EffortLevel
	// SPEC-WEB-CONSOLE-002 REQ-WC2-006: model_policy parity with the web console.
	modelPolicy := existingPrefs.ModelPolicy
	permissionMode := existingPrefs.PermissionMode
	if permissionMode == "" {
		permissionMode = defaultPermissionMode
	}

	// W-4: preserve raw theme value for migration banner output. The mode /
	// preset migration banners were removed by SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001
	// (runtime mode was inert; presets were retired alongside it).
	rawStatuslineTheme := existingPrefs.StatuslineTheme

	statuslineTheme := normalizeStatuslineTheme(existingPrefs.StatuslineTheme)

	// Extract enabled segment keys for MultiSelect default selection. The segment
	// step is now unconditional (preset==custom gate removed). When
	// prefs.StatuslineSegments is nil (new user), default to all 15 segments enabled
	// (matching .moai/config/sections/statusline.yaml 15-segment baseline, NOT the
	// 11-segment defaultStatuslineSegments() in internal/profile/sync.go which serves a
	// different purpose: yaml fallback when statusline.yaml is absent).
	statuslineSegmentsSelection := make([]string, 0, len(statuslineAllSegments))
	if existingPrefs.StatuslineSegments != nil {
		for _, key := range statuslineAllSegments {
			if existingPrefs.StatuslineSegments[key] {
				statuslineSegmentsSelection = append(statuslineSegmentsSelection, key)
			}
		}
	} else {
		statuslineSegmentsSelection = append(statuslineSegmentsSelection, statuslineAllSegments...)
	}

	// SPEC-WEB-CONSOLE-003: initialize the two project-config selects from the
	// CURRENT project config (quality.yaml / git-convention.yaml) — NOT from
	// existingPrefs, since development_mode/convention are project-config values,
	// not ProfilePreferences fields. Outside a MoAI project (no .moai dir) the
	// selects default to empty "(project default)" and the save is a no-op.
	var developmentMode, gitConvention string
	if cwd, err := os.Getwd(); err == nil {
		if info, statErr := os.Stat(filepath.Join(cwd, ".moai")); statErr == nil && info.IsDir() {
			if dm, gc, readErr := readCurrentProjectConfig(cwd); readErr == nil {
				developmentMode, gitConvention = dm, gc
			}
		}
	}

	// ====== Step 1: Language selection ======
	langOptions := []huh.Option[string]{
		huh.NewOption("English", "en"),
		huh.NewOption("Korean (한국어)", "ko"),
		huh.NewOption("Japanese (日本語)", "ja"),
		huh.NewOption("Chinese (中文)", "zh"),
	}

	langForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select your language").
				Description("Language for this wizard and Claude's responses.").
				Options(langOptions...).
				Value(&convLang),
		).Title("Language"),
	)

	if err := langForm.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Setup cancelled.")
			return nil
		}
		return fmt.Errorf("wizard error: %w", err)
	}

	// ====== Step 2: Display remaining forms in the selected language ======
	t := getProfileText(convLang)

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), t.ConfiguringProfile+"\n\n", profileName)

	// W-4: statusline theme migration banner — print before displaying the form.
	// The mode migration banner was removed by SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001.
	if rawStatuslineTheme != "" && rawStatuslineTheme != statuslineTheme {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), t.MigrationNoticeStatuslineTheme+"\n", rawStatuslineTheme, statuslineTheme)
	}

	form := huh.NewForm(
		// Section 1: User information
		huh.NewGroup(
			huh.NewInput().
				Title(t.UserNameTitle).
				Description(t.UserNameDesc).
				Value(&userName),
		).Title(t.IdentityTitle),

		// Section 2: Languages (after conversation language)
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.GitCommitLangTitle).
				Description(t.GitCommitLangDesc).
				Options(langOptions...).
				Value(&gitCommitLang),
			huh.NewSelect[string]().
				Title(t.CodeCommentLangTitle).
				Description(t.CodeCommentLangDesc).
				Options(langOptions...).
				Value(&codeCommentLang),
			huh.NewSelect[string]().
				Title(t.DocLangTitle).
				Description(t.DocLangDesc).
				Options(langOptions...).
				Value(&docLang),
		).Title(t.LanguagesTitle),

		// Section 3: Model settings (model override + policy + permission mode)
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.ModelOverrideTitle).
				Description(t.ModelOverrideDesc).
				Options(
					huh.NewOption(t.ModelDefault, ""),
					huh.NewOption(t.ModelOpus, "opus"),
					huh.NewOption(t.ModelOpus1M, "opus[1m]"),
					huh.NewOption(t.ModelSonnet, "sonnet"),
					huh.NewOption(t.ModelSonnet1M, "sonnet[1m]"),
					huh.NewOption(t.ModelHaiku, "haiku"),
					huh.NewOption(t.ModelOpusPlan, "opusplan"),
				).
				Value(&model),
			// SPEC-WEB-CONSOLE-002 REQ-WC2-006: model_policy select — parity with
			// the web console. Options mirror template.ValidModelPolicies() plus an
			// empty "(project default)" option (reusing the ModelDefault label).
			huh.NewSelect[string]().
				Title(t.ModelPolicyTitle).
				Description(t.ModelPolicyDesc).
				Options(
					huh.NewOption(t.ModelDefault, ""),
					huh.NewOption(t.ModelPolicyHigh, "high"),
					huh.NewOption(t.ModelPolicyMedium, "medium"),
					huh.NewOption(t.ModelPolicyLow, "low"),
				).
				Value(&modelPolicy),
			huh.NewSelect[string]().
				Title(t.EffortLevelTitle).
				Description(t.EffortLevelDesc).
				Options(
					huh.NewOption(t.EffortLevelDefault, ""),
					huh.NewOption(t.EffortLevelLow, "low"),
					huh.NewOption(t.EffortLevelMedium, "medium"),
					huh.NewOption(t.EffortLevelHigh, "high"),
					huh.NewOption(t.EffortLevelXHigh, "xhigh"),
					huh.NewOption(t.EffortLevelMax, "max"),
				).
				Value(&effortLevel),
			// S-4: option order — acceptEdits, auto, default, plan, bypass, dontAsk
			huh.NewSelect[string]().
				Title(t.PermissionModeTitle).
				Description(t.PermissionModeDesc).
				Options(
					huh.NewOption(t.PermAcceptEdits, "acceptEdits"),
					huh.NewOption(t.PermAuto, "auto"),
					huh.NewOption(t.PermDefault, "default"),
					huh.NewOption(t.PermPlan, "plan"),
					huh.NewOption(t.PermBypass, "bypassPermissions"),
					huh.NewOption(t.PermDontAsk, "dontAsk"),
				).
				Value(&permissionMode),
		).Title(t.ModelSettingsTitle),

		// Section 4: Display — theme only. The mode + preset Selects were removed
		// by SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001 (runtime mode was inert; named
		// presets were redundant with the segment map). Segments are configured in
		// the next unconditional section.
		// KEEP IN SYNC with statuslineThemeCanonical at top of file.
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.StatuslineThemeTitle).
				Description(t.StatuslineThemeDesc).
				Options(
					huh.NewOption(t.ThemeMoaiDark, "catppuccin-mocha"),
					huh.NewOption(t.ThemeMoaiLight, "catppuccin-latte"),
				).
				Value(&statuslineTheme),
		).Title(t.DisplayTitle),

		// Section 5: Segments — now unconditional (preset==custom gate removed by
		// SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001). 15 segments KEEP IN SYNC with
		// statuslineAllSegments slice at top of file.
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(t.StatuslineSegmentsTitle).
				Description(t.StatuslineSegmentsDesc).
				Options(
					huh.NewOption(t.SegmentClaudeVersion, "claude_version"),
					huh.NewOption(t.SegmentContext, "context"),
					huh.NewOption(t.SegmentDirectory, "directory"),
					huh.NewOption(t.SegmentEffortThinking, "effort_thinking"),
					huh.NewOption(t.SegmentGitBranch, "git_branch"),
					huh.NewOption(t.SegmentGitStatus, "git_status"),
					huh.NewOption(t.SegmentMoaiVersion, "moai_version"),
					huh.NewOption(t.SegmentModel, "model"),
					huh.NewOption(t.SegmentOutputStyle, "output_style"),
					huh.NewOption(t.SegmentPR, "pr"),
					huh.NewOption(t.SegmentSessionTime, "session_time"),
					huh.NewOption(t.SegmentTask, "task"),
					huh.NewOption(t.SegmentUsage5h, "usage_5h"),
					huh.NewOption(t.SegmentUsage7d, "usage_7d"),
					huh.NewOption(t.SegmentWorktree, "worktree"),
				).
				Value(&statuslineSegmentsSelection),
		).Title(t.StatuslineSegmentsTitle),

		// Section 6: Project config (development_mode + git_convention) — SPEC-WEB-CONSOLE-003.
		// Parity with the web console "Project" fieldset. Persisted to project config
		// (quality.yaml / git-convention.yaml), NOT the profile store.
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.DevelopmentModeTitle).
				Description(t.DevelopmentModeDesc).
				Options(
					huh.NewOption(t.ProjectDefaultOption, ""),
					huh.NewOption(t.DevelopmentModeDDD, "ddd"),
					huh.NewOption(t.DevelopmentModeTDD, "tdd"),
				).
				Value(&developmentMode),
			huh.NewSelect[string]().
				Title(t.GitConventionTitle).
				Description(t.GitConventionDesc).
				Options(
					huh.NewOption(t.ProjectDefaultOption, ""),
					huh.NewOption("auto", "auto"),
					huh.NewOption("conventional-commits", "conventional-commits"),
					huh.NewOption("angular", "angular"),
					huh.NewOption("karma", "karma"),
				).
				Value(&gitConvention),
		).Title(t.DevelopmentModeTitle),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), t.SetupCancelled)
			return nil
		}
		return fmt.Errorf("wizard error: %w", err)
	}

	// Normalize permission mode: "acceptEdits" is the project default, so store empty string to avoid an unnecessary override.
	if permissionMode == defaultPermissionMode {
		permissionMode = ""
	}

	// Build the segments map unconditionally (SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001
	// retired the preset==custom gate). The wizard always emits a full 15-key map.
	selected := make(map[string]bool, len(statuslineSegmentsSelection))
	for _, key := range statuslineSegmentsSelection {
		selected[key] = true
	}
	statuslineSegmentsMap := make(map[string]bool, len(statuslineAllSegments))
	for _, key := range statuslineAllSegments {
		statuslineSegmentsMap[key] = selected[key]
	}

	// Save preferences.
	prefs := profile.ProfilePreferences{
		UserName:           userName,
		ConversationLang:   convLang,
		GitCommitLang:      gitCommitLang,
		CodeCommentLang:    codeCommentLang,
		DocLang:            docLang,
		Model:              model,
		ModelPolicy:        modelPolicy,
		EffortLevel:        effortLevel,
		PermissionMode:     permissionMode,
		StatuslineSegments: statuslineSegmentsMap,
		StatuslineTheme:    statuslineTheme,
	}

	if err := profile.WritePreferences(profileName, prefs); err != nil {
		return fmt.Errorf("save preferences: %w", err)
	}

	// When inside a MoAI project, sync preferences to the project configuration.
	// When syncedProjectRoot is set, the final report shows the statusline.yaml path
	// so the user can verify where the changes were applied.
	var syncedProjectRoot string
	if cwd, err := os.Getwd(); err == nil {
		moaiDir := filepath.Join(cwd, ".moai")
		if info, err := os.Stat(moaiDir); err == nil && info.IsDir() {
			if err := profile.SyncToProjectConfig(cwd, prefs); err != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: failed to sync profile to project config: %v\n", err)
			} else {
				syncedProjectRoot = cwd
			}
			// SPEC-WEB-CONSOLE-003 REQ-WC3-006: persist the two project-config
			// selects (development_mode + git_convention) to quality.yaml /
			// git-convention.yaml via the config manager — the SAME write path as
			// the web console, NOT into ProfilePreferences. Empty values keep the
			// existing project-config value (EC-1).
			if err := persistProjectConfig(cwd, developmentMode, gitConvention); err != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: failed to persist project config: %v\n", err)
			}
		}
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), t.SavedProfile,
		profileName,
		profile.GetPreferencesPath(profileName))

	// Print a structured summary so the user can visually confirm all captured values.
	printProfileSummary(cmd.OutOrStdout(), &t, &prefs, syncedProjectRoot)
	return nil
}

// printProfileSummary writes a multi-line summary of the applied settings to out.
// When sync has been performed, the project-level YAML paths holding the values are also printed.
func printProfileSummary(out io.Writer, t *profileSetupText, prefs *profile.ProfilePreferences, syncedProjectRoot string) {
	// S-7: combine fields into a single Fprintf call. The SummaryStatuslineMode
	// row was removed by SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001 (mode retired).
	_, _ = fmt.Fprintf(out,
		"%s\n"+
			"  %s: %s\n"+
			"  %s: %s / %s / %s / %s\n"+
			"  %s: %s\n"+
			"  %s: %s\n"+
			"  %s: %s\n"+
			"  %s: %s\n",
		t.SummaryHeader,
		t.SummaryUserName, valueOrDash(prefs.UserName),
		t.SummaryLanguages,
		valueOrDash(prefs.ConversationLang),
		valueOrDash(prefs.GitCommitLang),
		valueOrDash(prefs.CodeCommentLang),
		valueOrDash(prefs.DocLang),
		t.SummaryModel, valueOrDefault(prefs.Model, t.SummaryDefault),
		t.SummaryEffort, valueOrDefault(prefs.EffortLevel, t.SummaryDefault),
		t.SummaryPermission, valueOrDefault(prefs.PermissionMode, defaultPermissionMode),
		t.SummaryStatuslineTheme, prefs.StatuslineTheme,
	)

	if syncedProjectRoot != "" {
		// S-1: print relative paths (syncedProjectRoot == cwd, so relative paths are hardcoded)
		_, _ = fmt.Fprintf(out, "\n%s\n", t.SummarySyncedHeader)
		_, _ = fmt.Fprintf(out, "  statusline.yaml -> .moai/config/sections/statusline.yaml\n")
		_, _ = fmt.Fprintf(out, "  language.yaml   -> .moai/config/sections/language.yaml\n")
	} else {
		_, _ = fmt.Fprintf(out, "\n%s\n", t.SummarySyncSkipped)
	}
}

// valueOrDash returns "-" when v is empty.
// Used for fields such as user name or language where an empty value means "not set".
func valueOrDash(v string) string {
	if v == "" {
		return "-"
	}
	return v
}

// valueOrDefault returns fallback when v is empty.
// Used for slots where an empty string means "use runtime default".
func valueOrDefault(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
