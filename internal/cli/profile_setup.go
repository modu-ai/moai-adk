package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/statusline"
	"github.com/spf13/cobra"
)

// statuslineModeCanonical holds the canonical values offered by the wizard — must stay in sync with the huh.NewOption block (~line 230).
// KEEP IN SYNC with huh.NewOption block at ~line 230
var statuslineModeCanonical = []string{defaultStatuslineMode, "full"}

// KEEP IN SYNC with huh.NewOption block at ~line 240
var statuslineThemeCanonical = []string{defaultStatuslineTheme, "catppuccin-latte"}

// Wizard default constants.
const (
	defaultStatuslineMode  = "default"
	defaultStatuslineTheme = "catppuccin-mocha"
	defaultPermissionMode  = "acceptEdits"
)

// isCanonicalStatuslineMode reports whether s is a canonical value in the wizard option slice.
func isCanonicalStatuslineMode(s string) bool {
	for _, v := range statuslineModeCanonical {
		if v == s {
			return true
		}
	}
	return false
}

// isCanonicalStatuslineTheme reports whether s is a canonical value in the wizard option slice.
func isCanonicalStatuslineTheme(s string) bool {
	for _, v := range statuslineThemeCanonical {
		if v == s {
			return true
		}
	}
	return false
}

// normalizeStatuslineModeRaw calls statusline.NormalizeMode to encapsulate string conversion noise.
func normalizeStatuslineModeRaw(s string) string {
	return string(statusline.NormalizeMode(statusline.StatuslineMode(s)))
}

// normalizeStatuslineMode returns a mode value compatible with wizard options.
// Converts deprecated names from older statusline versions to v3 names,
// and falls back to "default" for values not present in the option set.
func normalizeStatuslineMode(mode string) string {
	if mode == "" {
		return defaultStatuslineMode
	}
	normalized := normalizeStatuslineModeRaw(mode)
	if isCanonicalStatuslineMode(normalized) {
		return normalized
	}
	return defaultStatuslineMode
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
	permissionMode := existingPrefs.PermissionMode
	if permissionMode == "" {
		permissionMode = defaultPermissionMode
	}

	// W-4: preserve raw values for migration banner output
	rawStatuslineMode := existingPrefs.StatuslineMode
	rawStatuslineTheme := existingPrefs.StatuslineTheme

	statuslineMode := normalizeStatuslineMode(existingPrefs.StatuslineMode)
	statuslineTheme := normalizeStatuslineTheme(existingPrefs.StatuslineTheme)

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

	// W-4: statusline mode/theme migration banner — print before displaying the form
	if rawStatuslineMode != "" && rawStatuslineMode != statuslineMode {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), t.MigrationNoticeStatuslineMode+"\n", rawStatuslineMode, statuslineMode)
	}
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

		// Section 3: Model settings (model override + permission mode)
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

		// Section 4: Display — mode, theme
		// KEEP IN SYNC with statuslineModeCanonical at top of file
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.StatuslineModeTitle).
				Description(t.StatuslineModeDesc).
				Options(
					huh.NewOption(t.ModeDefault, "default"),
					huh.NewOption(t.ModeFull, "full"),
				).
				Value(&statuslineMode),
			// KEEP IN SYNC with statuslineThemeCanonical at top of file
			huh.NewSelect[string]().
				Title(t.StatuslineThemeTitle).
				Description(t.StatuslineThemeDesc).
				Options(
					huh.NewOption(t.ThemeMoaiDark, "catppuccin-mocha"),
					huh.NewOption(t.ThemeMoaiLight, "catppuccin-latte"),
				).
				Value(&statuslineTheme),
		).Title(t.DisplayTitle),
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

	// Save preferences.
	prefs := profile.ProfilePreferences{
		UserName:         userName,
		ConversationLang: convLang,
		GitCommitLang:    gitCommitLang,
		CodeCommentLang:  codeCommentLang,
		DocLang:          docLang,
		Model:            model,
		EffortLevel:      effortLevel,
		PermissionMode:   permissionMode,
		StatuslineMode:   statuslineMode,
		StatuslineTheme:  statuslineTheme,
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
	// S-7: combine 7 fields into a single Fprintf call
	_, _ = fmt.Fprintf(out,
		"%s\n"+
			"  %s: %s\n"+
			"  %s: %s / %s / %s / %s\n"+
			"  %s: %s\n"+
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
		// S-2: StatuslineMode/Theme are always non-empty after normalization, so valueOrDefault is unnecessary
		t.SummaryPermission, valueOrDefault(prefs.PermissionMode, defaultPermissionMode),
		t.SummaryStatuslineMode, prefs.StatuslineMode,
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
