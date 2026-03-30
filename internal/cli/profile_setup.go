package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/spf13/cobra"
)

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
// The first question is language selection; all subsequent UI text
// is displayed in the selected language.
func runProfileSetup(cmd *cobra.Command, args []string) error {
	profileName := "default"
	if len(args) > 0 {
		profileName = args[0]
	}

	// Load existing config as defaults
	existingPrefs, err := profile.ReadPreferences(profileName)
	if err != nil {
		return fmt.Errorf("read existing preferences: %w", err)
	}

	// Form values pre-filled from existing config
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

	model := existingPrefs.Model
	permissionMode := existingPrefs.PermissionMode
	if permissionMode == "" {
		permissionMode = "acceptEdits" // project default
	}

	statuslineMode := existingPrefs.StatuslineMode
	if statuslineMode == "" {
		statuslineMode = "default"
	}
	statuslineTheme := existingPrefs.StatuslineTheme
	if statuslineTheme == "" {
		statuslineTheme = "default"
	}
	// ====== Step 1: Language Selection ======
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

	// ====== Step 2: Remaining form in selected language ======
	t := getProfileText(convLang)

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), t.ConfiguringProfile+"\n\n", profileName)

	form := huh.NewForm(
		// Section 1: Identity
		huh.NewGroup(
			huh.NewInput().
				Title(t.UserNameTitle).
				Description(t.UserNameDesc).
				Value(&userName),
		).Title(t.IdentityTitle),

		// Section 2: Languages (remaining)
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

		// Section 3: Model Settings (model override + permission mode)
		// Note: model_policy is now configured per-project via moai init / moai update -c.
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.ModelOverrideTitle).
				Description(t.ModelOverrideDesc).
				Options(
					huh.NewOption(t.ModelDefault, ""),
					huh.NewOption(t.ModelOpus, "claude-opus-4-6"),
					huh.NewOption(t.ModelOpus1M, "claude-opus-4-6[1m]"),
					huh.NewOption(t.ModelSonnet, "claude-sonnet-4-6"),
					huh.NewOption(t.ModelSonnet1M, "claude-sonnet-4-6[1m]"),
					huh.NewOption(t.ModelHaiku, "claude-haiku-4-5-20251001"),
					huh.NewOption(t.ModelOpusPlan, "opusplan"),
				).
				Value(&model),
			huh.NewSelect[string]().
				Title(t.PermissionModeTitle).
				Description(t.PermissionModeDesc).
				Options(
					huh.NewOption(t.PermAuto, "auto"),
					huh.NewOption(t.PermAcceptEdits, "acceptEdits"),
					huh.NewOption(t.PermDefault, "default"),
					huh.NewOption(t.PermPlan, "plan"),
					huh.NewOption(t.PermBypass, "bypassPermissions"),
					huh.NewOption(t.PermDontAsk, "dontAsk"),
				).
				Value(&permissionMode),
		).Title(t.ModelSettingsTitle),

		// Section 5: Display — Mode, Theme, and Preset in one screen
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.StatuslineModeTitle).
				Description(t.StatuslineModeDesc).
				Options(
					huh.NewOption(t.ModeDefault, "default"),
					huh.NewOption(t.ModeFull, "full"),
				).
				Value(&statuslineMode),
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

	// Normalize permission mode: "acceptEdits" is the project default,
	// so store as empty to avoid unnecessary override.
	if permissionMode == "acceptEdits" {
		permissionMode = ""
	}

	// Build and save preferences
	prefs := profile.ProfilePreferences{
		UserName:         userName,
		ConversationLang: convLang,
		GitCommitLang:    gitCommitLang,
		CodeCommentLang:  codeCommentLang,
		DocLang:          docLang,
		Model:            model,
		PermissionMode:   permissionMode,
		StatuslineMode:   statuslineMode,
		StatuslineTheme:  statuslineTheme,
		TeammateDisplay:  "auto",
	}

	if err := profile.WritePreferences(profileName, prefs); err != nil {
		return fmt.Errorf("save preferences: %w", err)
	}

	// Sync preferences to project config if inside a MoAI project
	if cwd, err := os.Getwd(); err == nil {
		moaiDir := filepath.Join(cwd, ".moai")
		if info, err := os.Stat(moaiDir); err == nil && info.IsDir() {
			if err := profile.SyncToProjectConfig(cwd, prefs); err != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: failed to sync profile to project config: %v\n", err)
			}
		}
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), t.SavedProfile,
		profileName,
		profile.GetPreferencesPath(profileName))
	return nil
}
