package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
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

// normalizeModel maps a stored model id onto the alias form the picker
// offers.
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

// schemaSelectOptions builds the option list for a schema select field from
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
// The list is version neutral ({Label, Value}, design.md §3): the absorbed v2
// profile wizard takes it as wizard.Option arguments.
//
// @MX:NOTE: [AUTO] Single derivation site for every wizard select backed by the shared schema.
func schemaSelectOptions(t profileSetupText, field string, withEmpty bool) []wizard.Option {
	defs := settings.FieldOptionDefs(field)
	opts := make([]wizard.Option, 0, len(defs)+1)
	if withEmpty {
		if empty := settings.EmptyLabelFor(field); empty != "" {
			opts = append(opts, wizard.Option{Label: empty, Value: ""})
		}
	}
	for _, d := range defs {
		opts = append(opts, wizard.Option{Label: optionLabelFor(t, d), Value: d.Value})
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
	// Routed through the runProfileSetupFn seam (profile.go) so the explicit
	// entry is countable alongside `moai profile --setup` (REQ-ITI-001).
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProfileSetupFn(cmd, args)
	},
}

func init() {
	profileCmd.AddCommand(profileSetupCmd)
}

// initialProfileResult maps the stored preferences onto the absorbed wizard's
// initial values — the binding point of the v2 profile wizard. Every select
// pre-selects the stored value (REQ-ITI-005 (5)); the empty language slots
// pre-select English; the empty permission mode pre-selects acceptEdits
// (REQ-ITI-005 (9)); development_mode initializes from the CURRENT project
// config, not from the profile store (SPEC-WEB-CONSOLE-003). Outside a MoAI
// project (no .moai dir) it stays empty "(project default)" and the save is a
// no-op.
//
// @MX:NOTE: [AUTO] Wizard v3 migration — normalizes deprecated Claude model IDs to canonical aliases before binding.
// @MX:REASON: Prevents silent loss of existing prefs values in the wizard bindings after the "claude-opus-4-7" option was removed from the picker; the mapping stays owned by template.ModelAliasTable (single SSOT) via normalizeModel.
func initialProfileResult(existingPrefs profile.ProfilePreferences) wizard.ProfileResult {
	result := wizard.ProfileResult{
		UserName:        existingPrefs.UserName,
		Model:           normalizeModel(existingPrefs.Model),
		ModelPolicy:     existingPrefs.ModelPolicy,
		EffortLevel:     existingPrefs.EffortLevel,
		PermissionMode:  existingPrefs.PermissionMode,
		GitCommitLang:   existingPrefs.GitCommitLang,
		CodeCommentLang: existingPrefs.CodeCommentLang,
		DocLang:         existingPrefs.DocLang,
	}
	for _, field := range []*string{
		&result.ConversationLang, &result.GitCommitLang,
		&result.CodeCommentLang, &result.DocLang,
	} {
		if *field == "" {
			*field = "en"
		}
	}
	if result.PermissionMode == "" {
		result.PermissionMode = defaultPermissionMode
	}
	if cwd, err := os.Getwd(); err == nil {
		if info, statErr := os.Stat(filepath.Join(cwd, ".moai")); statErr == nil && info.IsDir() {
			if dm, _, readErr := readCurrentProjectConfig(cwd); readErr == nil {
				result.DevelopmentMode = dm
			}
		}
	}
	return result
}

// runProfileSetup runs the interactive profile configuration wizard through
// the absorbed v2 wizard (design.md §2.2): the ten profile questions as one
// multi-group form; the first question is language selection and all
// subsequent groups re-render in the chosen language.
//
// SPEC-SESSION-WORKTREE-001 M6/M4: session-worktree auto-entry + exit disposal.
// Auto-entry is wired HERE (not in runProfileCmd / list / current / delete) so
// it gates to the only profile subverb that mutates the PROJECT tree
// (persistProjectConfig writes .moai/config/sections/*.yaml). Read-only
// subverbs (list / current) and global-dir-only mutators (delete) do NOT
// trigger auto-entry (EC-7). The auto-entry runs BEFORE any shared-state
// mutation (REQ-SW-002). When the feature is OFF (default) the wrapper returns
// "" and runProfileSetup is byte-identical to the baseline (REQ-SW-001).
//
// REQ-SW-017: the worktree scopes to the PROJECT, NEVER to the global profile
// dir at ~/.moai/claude-profiles/. emitProfileScopeNotice states this
// explicitly so the user is not misled into thinking the launch-ledger race is
// solved. The profile dir remains GLOBAL; isolating it is out of scope (§3).
//
// The named return (err error) lets the deferred M4 cleanup derive cleanExit
// (err == nil means exit 0). wtPath == "" makes cleanup a no-op.
func runProfileSetup(cmd *cobra.Command, args []string) (err error) {
	// SPEC-SESSION-WORKTREE-001 M6: session-worktree auto-entry. Runs before
	// any shared-state mutation (ReadPreferences / WritePreferences /
	// persistProjectConfig). Mirrors the M2 init / M3 web wiring.
	swCfg := loadSessionWorktreeConfig(cmd)
	wtPath := enterSessionWorktreeFn(swCfg, "profile", cmd.ErrOrStderr())
	if wtPath != "" {
		// REQ-SW-017: honest scope notice — names BOTH paths and states the
		// profile dir is NOT isolated.
		emitProfileScopeNotice(cmd.ErrOrStderr(), wtPath)
		if cerr := os.Chdir(wtPath); cerr != nil {
			// Chdir failure is a fail-back (REQ-SW-004): continue in the
			// shared checkout. The worktree was materialized but unusable
			// from this process.
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
				"moai: session-worktree chdir into %s failed (%v); continuing in shared checkout for %q\n",
				wtPath, cerr, "profile")
		}
	}
	defer func() {
		// SPEC-WORKTREE-KEY-WIRING-001 M2: auto-merge runs BEFORE disposal —
		// merge-then-dispose is the only safe order, and independent of
		// auto_cleanup (REQ-WKW-013).
		sessionExitAutoMerge(swCfg, wtPath, err == nil, cmd.ErrOrStderr())
		// cleanup goes through the test seam (t586 M5, AC-ITI-006/007
		// preservation) which profile.go binds to cleanupSessionWorktree.
		cleanupSessionWorktreeFn(swCfg, wtPath, err == nil, cmd.ErrOrStderr())
	}()

	profileName := "default"
	if len(args) > 0 {
		profileName = args[0]
	}

	// Load existing preferences as the wizard's initial values.
	existingPrefs, err := profile.ReadPreferences(profileName)
	if err != nil {
		return fmt.Errorf("read existing preferences: %w", err)
	}

	initial := initialProfileResult(existingPrefs)

	// The announcement and the option labels resolve in the run's initial
	// locale — the stored conversation language, falling back to English for
	// a fresh profile (design.md §3).
	t := getProfileText(initial.ConversationLang)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), t.ConfiguringProfile+"\n\n", profileName)

	// Run the absorbed v2 wizard (design.md §2.2) behind the
	// profileWizardRunner seam: ten questions, five groups, one form.
	result, err := profileWizardRunner(initial, buildProfileOptions(t), initial.ConversationLang)
	if err != nil {
		if errors.Is(err, wizard.ErrCancelled) {
			// The cancellation message renders in the language the user had
			// reached: the answered conversation language, else the stored
			// value, else English (design.md §5).
			cancelLocale := initial.ConversationLang
			if result != nil && result.ConversationLang != "" {
				cancelLocale = result.ConversationLang
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), getProfileText(cancelLocale).SetupCancelled)
			return nil
		}
		return err
	}

	// The saved/summary text renders in the language the wizard ended with.
	t = getProfileText(result.ConversationLang)
	permissionMode := result.PermissionMode
	developmentMode := result.DevelopmentMode

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
		UserName:           result.UserName,
		ConversationLang:   result.ConversationLang,
		GitCommitLang:      result.GitCommitLang,
		CodeCommentLang:    result.CodeCommentLang,
		DocLang:            result.DocLang,
		Model:              result.Model,
		ModelPolicy:        result.ModelPolicy,
		EffortLevel:        result.EffortLevel,
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
