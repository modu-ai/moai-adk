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

	modelPolicy := existingPrefs.ModelPolicy
	if modelPolicy == "" {
		modelPolicy = "high"
	}
	model := existingPrefs.Model
	bypass := existingPrefs.Bypass

	statuslinePreset := existingPrefs.StatuslinePreset
	if statuslinePreset == "" {
		statuslinePreset = "full"
	}
	statuslineTheme := existingPrefs.StatuslineTheme
	if statuslineTheme == "" {
		statuslineTheme = "default"
	}
	// Statusline segment toggles
	segModel := getSegmentDefault(existingPrefs.StatuslineSegments, "model", true)
	segContext := getSegmentDefault(existingPrefs.StatuslineSegments, "context", true)
	segOutputStyle := getSegmentDefault(existingPrefs.StatuslineSegments, "output_style", true)
	segDirectory := getSegmentDefault(existingPrefs.StatuslineSegments, "directory", true)
	segGitStatus := getSegmentDefault(existingPrefs.StatuslineSegments, "git_status", true)
	segClaudeVersion := getSegmentDefault(existingPrefs.StatuslineSegments, "claude_version", true)
	segMoaiVersion := getSegmentDefault(existingPrefs.StatuslineSegments, "moai_version", true)
	segGitBranch := getSegmentDefault(existingPrefs.StatuslineSegments, "git_branch", true)

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

		// Section 3: Model Settings (policy + model override)
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.ModelPolicyTitle).
				Description(t.ModelPolicyDesc).
				Options(
					huh.NewOption(t.ModelPolicyHigh, "high"),
					huh.NewOption(t.ModelPolicyMedium, "medium"),
					huh.NewOption(t.ModelPolicyLow, "low"),
				).
				Value(&modelPolicy),
			huh.NewSelect[string]().
				Title(t.ModelOverrideTitle).
				Description(t.ModelOverrideDesc).
				Options(
					huh.NewOption(t.ModelDefault, ""),
					huh.NewOption(t.ModelOpus, "claude-opus-4-6"),
					huh.NewOption(t.ModelSonnet, "claude-sonnet-4-6"),
					huh.NewOption(t.ModelHaiku, "claude-haiku-4-5-20251001"),
					huh.NewOption(t.ModelOpusPlan, "opusplan"),
				).
				Value(&model),
			huh.NewConfirm().
				Title(t.BypassTitle).
				Description(t.BypassDesc).
				Value(&bypass),
		).Title(t.ModelSettingsTitle),

		// Section 5: Display
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.StatuslineTitle).
				Description(t.StatuslineDesc).
				Options(
					huh.NewOption(t.StatuslineFull, "full"),
					huh.NewOption(t.StatuslineCompact, "compact"),
					huh.NewOption(t.StatuslineMinimal, "minimal"),
					huh.NewOption(t.StatuslineCustom, "custom"),
				).
				Value(&statuslinePreset),
			huh.NewSelect[string]().
				Title(t.StatuslineThemeTitle).
				Description(t.StatuslineThemeDesc).
				Options(
					huh.NewOption(t.ThemeDefault, "default"),
					huh.NewOption(t.ThemeCatppuccinMocha, "catppuccin-mocha"),
					huh.NewOption(t.ThemeCatppuccinLatte, "catppuccin-latte"),
				).
				Value(&statuslineTheme),
		).Title(t.DisplayTitle),

		// Section 5b: Custom segments (shown only when preset is "custom")
		huh.NewGroup(
			huh.NewConfirm().Title(t.SegModel).Value(&segModel),
			huh.NewConfirm().Title(t.SegContext).Value(&segContext),
			huh.NewConfirm().Title(t.SegOutputStyle).Value(&segOutputStyle),
			huh.NewConfirm().Title(t.SegDirectory).Value(&segDirectory),
			huh.NewConfirm().Title(t.SegGitStatus).Value(&segGitStatus),
			huh.NewConfirm().Title(t.SegClaudeVersion).Value(&segClaudeVersion),
			huh.NewConfirm().Title(t.SegMoaiVersion).Value(&segMoaiVersion),
			huh.NewConfirm().Title(t.SegGitBranch).Value(&segGitBranch),
		).Title(t.SegmentsTitle).
			WithHideFunc(func() bool { return statuslinePreset != "custom" }),

	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), t.SetupCancelled)
			return nil
		}
		return fmt.Errorf("wizard error: %w", err)
	}

	// Build and save preferences
	prefs := profile.ProfilePreferences{
		UserName:         userName,
		ConversationLang: convLang,
		GitCommitLang:    gitCommitLang,
		CodeCommentLang:  codeCommentLang,
		DocLang:          docLang,
		ModelPolicy:      modelPolicy,
		Model:            model,
		Bypass:           bypass,
		StatuslinePreset: statuslinePreset,
		StatuslineTheme:  statuslineTheme,
		TeammateDisplay:  "auto",
	}

	// Only persist segment toggles for custom preset
	if statuslinePreset == "custom" {
		prefs.StatuslineSegments = map[string]bool{
			"model":          segModel,
			"context":        segContext,
			"output_style":   segOutputStyle,
			"directory":      segDirectory,
			"git_status":     segGitStatus,
			"claude_version": segClaudeVersion,
			"moai_version":   segMoaiVersion,
			"git_branch":     segGitBranch,
		}
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

// getSegmentDefault returns the segment toggle value from the map, or the default.
func getSegmentDefault(segments map[string]bool, name string, defaultVal bool) bool {
	if segments == nil {
		return defaultVal
	}
	if val, ok := segments[name]; ok {
		return val
	}
	return defaultVal
}
