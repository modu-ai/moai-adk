package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/settings"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/models"
	"github.com/spf13/cobra"
)

// Wizard default constants.
const (
	defaultPermissionMode = "acceptEdits"
)

// acceptEditsConfirmationLine is the deterministic confirmation emitted by the
// wizard when the user selects "acceptEdits" as permissionMode. REQ-CCI-006
// requires the wizard to surface the empty-string normalization so the user
// does not perceive the selection as a silent no-op. The anchor tokens
// ("acceptEdits", "project default", "settings.local.json") are grep-stable
// and asserted by TestEmitAcceptEditsConfirmationAnchor (AC-CCI-006).
const acceptEditsConfirmationLine = "Note: \"acceptEdits\" is the project default, so no settings.local.json defaultMode override will be written."

// emitAcceptEditsConfirmation writes the acceptEdits confirmation line to out.
// Called from runProfileSetup immediately after the acceptEdits→"" normalization
// so the user sees why nothing was persisted to settings.local.json.
func emitAcceptEditsConfirmation(out io.Writer) {
	_, _ = fmt.Fprintln(out, acceptEditsConfirmationLine)
}

// @MX:NOTE: [AUTO] Wizard v3 migration — normalizes deprecated Claude model IDs to canonical aliases.
// @MX:REASON: Prevents silent loss of existing prefs values in huh.Select bindings after the "claude-opus-4-7" option was removed from the previous wizard.
//
// The alias↔canonical-id mapping is owned by template.ModelAliasTable (single
// SSOT). This function performs the reverse direction (full-id → short alias)
// via template.ModelAliasFromCanonicalID, so adding a new model only requires
// one new row in ModelAliasTable rather than touching this switch too.
func normalizeModel(m string) string {
	// Empty and canonical aliases pass through unchanged.
	if m == "" {
		return m
	}
	for _, alias := range template.ModelAliasPickerValues() {
		if m == alias {
			return m
		}
	}
	// Split the [1m] suffix so the reverse lookup can match the base id.
	base, suffix := splitModelSuffix(m)
	alias := template.ModelAliasFromCanonicalID(base)
	if alias == base {
		// base is not a known canonical id. It may still be a bare short alias
		// that the picker no longer offers (the 1M unification dropped the bare
		// opus/sonnet/fable options); those are handled by promoteTo1M below.
		// Otherwise the legacy " 1M" suffix is the last deprecated form left.
		if _, known := template.ModelAliasTable[base]; !known {
			return normalizeModelLegacy1M(m)
		}
	}
	if suffix == "" {
		return promoteTo1M(alias)
	}
	return alias + suffix
}

// promoteTo1M advances a bare short alias to its "[1m]" form when the picker
// offers only that form. The 1M unification exposes opus/sonnet/fable solely as
// [1m] variants, so a prefs value carrying the bare alias — either stored before
// the unification or resolved from a deprecated full id — must migrate rather
// than reset to the runtime default. Aliases with no [1m] variant on the picker
// (haiku, and the opusplan routing alias) are returned unchanged.
func promoteTo1M(alias string) string {
	oneM := alias + "[1m]"
	for _, v := range template.ModelAliasPickerValues() {
		if v == oneM {
			return oneM
		}
	}
	return alias
}

// normalizeModelLegacy1M handles the deprecated " <version> 1M" suffix form
// (e.g. "claude-opus-4-6 1M") that predates the "[1m]" convention. It maps
// those legacy strings to the current alias + "[1m]" form via the central
// table's reverse lookup. Unknown legacy forms reset to the runtime default.
func normalizeModelLegacy1M(m string) string {
	const legacy1MSuffix = " 1M"
	if !strings.HasSuffix(m, legacy1MSuffix) {
		return ""
	}
	base := strings.TrimSuffix(m, legacy1MSuffix)
	alias := template.ModelAliasFromCanonicalID(base)
	if alias == base {
		return ""
	}
	return alias + "[1m]"
}

// schemaSelectOptions builds a huh option list for a schema select field from
// settings.FieldOptionDefs — the SHARED option-list SSOT that internal/web already
// reads — so the TUI wizard and the web console cannot drift apart. Each option's
// wire value is the schema's canonical value (for "model" that is the short alias
// form from template.ModelAliasPickerValues, matching what normalizeModel returns
// and what the web validator accepts); the human label is resolved through
// schemaOptionBridge, falling back to the wire value when a field has no localized
// option labels.
//
// withEmpty controls whether the field's canonical empty option
// (settings.EmptyLabelFor) is prepended. It is per-call-site rather than derived
// from the schema because permission_mode declares an empty label for the web
// console but the wizard has never offered one (it defaults to acceptEdits and
// normalizes that back to "" on save).
//
// @MX:NOTE: [AUTO] Single derivation site for every wizard select backed by the shared schema.
func schemaSelectOptions(t profileSetupText, field string, withEmpty bool) []huh.Option[string] {
	defs := settings.FieldOptionDefs(field)
	opts := make([]huh.Option[string], 0, len(defs)+1)
	if withEmpty {
		if empty := settings.EmptyLabelFor(field); empty != "" {
			opts = append(opts, huh.NewOption(empty, ""))
		}
	}
	for _, d := range defs {
		opts = append(opts, huh.NewOption(optionLabelFor(t, d), d.Value))
	}
	return opts
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

// The 7 nested project-config fields (quality coverage targets + git-convention
// auto-detection) are no longer collected by this wizard, so the TUI-side input
// struct and its read/write wrappers were removed. The underlying shared seam
// (settings.ReadProjectNestedConfig / settings.WriteProjectNestedConfig) is
// untouched — the web console still drives it, and internal/settings owns its
// round-trip and empty=preserve tests.

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

	// SPEC-WEB-CONSOLE-003: initialize the development_mode select from the CURRENT
	// project config (quality.yaml) — NOT from existingPrefs, since development_mode
	// is a project-config value, not a ProfilePreferences field. Outside a MoAI
	// project (no .moai dir) the select defaults to empty "(project default)" and
	// the save is a no-op. The sibling git_convention select was removed from the
	// wizard, so its persisted value is read by nobody here and left untouched on save.
	var developmentMode string
	if cwd, err := os.Getwd(); err == nil {
		if info, statErr := os.Stat(filepath.Join(cwd, ".moai")); statErr == nil && info.IsDir() {
			if dm, _, readErr := readCurrentProjectConfig(cwd); readErr == nil {
				developmentMode = dm
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
	).WithTheme(moaiHuhTheme())

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

	// The statusline-theme migration banner was removed alongside the theme Select:
	// with the theme fixed to fixedStatuslineTheme there is no stored value being
	// normalized onto a widget, so there is nothing to notify the user about.

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

		// Section 3: Model settings (model override + policy + permission mode).
		// The option LISTS (values + order) and the empty-option labels are both
		// single-sourced from the settings schema via schemaSelectOptions, so the
		// wizard and the web console render the same canonical value set. Before
		// this, the wizard re-declared the model list inline and emitted canonical
		// model ids (claude-opus-5), which the web validator rejects and which
		// normalizeModel never produces — the two surfaces write the same
		// preferences.yaml, so the value shape has to match. The verbose option
		// labels stay localized (resolved through schemaOptionBridge).
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.ModelOverrideTitle).
				Description(t.ModelOverrideDesc).
				Options(schemaSelectOptions(t, "model", true)...).
				Value(&model),
			// model_policy stays a CLI-only field: the web console dropped it (it
			// duplicates the agentfm performance tier), so it has no schema entry
			// and its options are declared here against template.ValidModelPolicies().
			// The empty-option label is still single-sourced from the schema.
			huh.NewSelect[string]().
				Title(t.ModelPolicyTitle).
				Description(t.ModelPolicyDesc).
				Options(
					huh.NewOption(settings.EmptyLabelFor("model_policy"), ""),
					huh.NewOption(t.ModelPolicyHigh, "high"),
					huh.NewOption(t.ModelPolicyMedium, "medium"),
					huh.NewOption(t.ModelPolicyLow, "low"),
				).
				Value(&modelPolicy),
			huh.NewSelect[string]().
				Title(t.EffortLevelTitle).
				Description(t.EffortLevelDesc).
				Options(schemaSelectOptions(t, "effort_level", true)...).
				Value(&effortLevel),
			// S-4: option order — acceptEdits, auto, default, plan, bypass, dontAsk.
			// The schema's permissionModeOptions() mirrors that exact order, and the
			// wizard offers no empty option here (acceptEdits is the default and is
			// normalized back to "" on save).
			huh.NewSelect[string]().
				Title(t.PermissionModeTitle).
				Description(t.PermissionModeDesc).
				Options(schemaSelectOptions(t, "permission_mode", false)...).
				Value(&permissionMode),
		).Title(t.ModelSettingsTitle),

		// Section 4: Project config — development_mode only. Persisted to the project
		// config (quality.yaml) via the config manager, NOT the profile store. The
		// empty-option label is single-sourced from the schema (REQ-WC10-013).
		//
		// Removed from the wizard (settings now use their stored/default value and are
		// never written by a wizard run): the statusline theme Select (fixed to
		// fixedStatuslineTheme), the 16-segment MultiSelect, the git_convention Select,
		// the 3 nested quality fields, and the 4 nested git auto-detection fields.
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.DevelopmentModeTitle).
				Description(t.DevelopmentModeDesc).
				Options(schemaSelectOptions(t, "development_mode", true)...).
				Value(&developmentMode),
		).Title(t.DevelopmentModeTitle),
	).WithTheme(moaiHuhTheme())

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), t.SetupCancelled)
			return nil
		}
		return fmt.Errorf("wizard error: %w", err)
	}

	// Normalize permission mode: "acceptEdits" is the project default, so store
	// empty string to avoid an unnecessary override. The normalization is NOT
	// silent — emitAcceptEditsConfirmation surfaces it to the user so the
	// selection is not perceived as a no-op (REQ-CCI-006).
	if permissionMode == defaultPermissionMode {
		permissionMode = ""
		emitAcceptEditsConfirmation(cmd.OutOrStdout())
	}

	// Save preferences.
	//
	// StatuslineSegments carries the profile's STORED map through untouched. The
	// segment MultiSelect is gone, so the wizard has no opinion about segments — but
	// WritePreferences marshals the whole struct over the file, so passing nil here
	// would DELETE statusline_segments from preferences.yaml. Carrying the existing
	// value through is what makes "removed from the wizard" mean "left alone" rather
	// than "blanked". A profile that never stored segments keeps its nil.
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
		StatuslineSegments: existingPrefs.StatuslineSegments,
		// StatuslineTheme is deliberately left at its zero value. The wizard no
		// longer manages the statusline theme at all: the terminal's apparent
		// statusline colour turned out to come from the Claude Code theme, not
		// from this setting, so the knob was removed rather than fixed to a value.
		// Empty means `omitempty` drops statusline_theme from preferences.yaml and
		// syncStatusline leaves .moai/config/sections/statusline.yaml untouched
		// (internal/profile/sync.go) — removed, not overwritten.
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
			// The project sync gets a nil segment map even when the profile store
			// carries one: syncStatusline treats nil as "preserve what is on disk"
			// (internal/profile/sync.go), and with the MultiSelect removed, editing
			// .moai/config/sections/statusline.yaml by hand is the ONLY way left to
			// change segments. Pushing the profile's map would silently clobber that
			// edit on every wizard run. The theme needs no special handling here: it
			// is already empty in prefs, which syncStatusline also treats as preserve.
			syncPrefs := prefs
			syncPrefs.StatuslineSegments = nil
			if err := profile.SyncToProjectConfig(cwd, syncPrefs); err != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: failed to sync profile to project config: %v\n", err)
			} else {
				syncedProjectRoot = cwd
			}
			// SPEC-WEB-CONSOLE-003 REQ-WC3-006: persist development_mode to
			// quality.yaml via the config manager — the SAME write path as the web
			// console, NOT into ProfilePreferences. The empty convention argument is
			// deliberate: the git_convention Select was removed from the wizard, and
			// persistProjectConfig writes only non-empty values (EC-1), so the stored
			// git-convention.yaml value is preserved untouched.
			if err := persistProjectConfig(cwd, developmentMode, ""); err != nil {
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
	// row was removed by SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001 (mode retired);
	// the statusline THEME row is gone too, because the wizard no longer collects
	// or writes a theme. Reporting a value the wizard did not apply would be
	// worse than reporting nothing.
	_, _ = fmt.Fprintf(out,
		"%s\n"+
			"  %s: %s\n"+
			"  %s: %s / %s / %s / %s\n"+
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

// The validateOptional* helpers were removed with the nested quality /
// git-auto-detection Inputs they validated. The equivalent validation for the web
// console lives in internal/settings + internal/web and is untouched.
