package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// This file is the regression guard for the wizard questions that were REMOVED
// from `moai profile setup`:
//
//	statusline theme Select        → fixed to fixedStatuslineTheme, never asked
//	statusline 16-segment MultiSelect → never asked, on-disk map left alone
//	git_convention Select          → never asked, git-convention.yaml left alone
//	nested quality group (3 fields)  → never asked, quality.yaml left alone
//	nested git auto-detection group (4 fields) → never asked, left alone
//
// The guard has two halves that must be read together:
//
//   - The SOURCE half pins what runProfileSetup builds and what arguments it hands
//     to its persistence seams. runProfileSetup drives huh forms and needs a TTY, so
//     it cannot be invoked from a unit test; a source assertion is how the wizard's
//     own composition is observed.
//   - The BEHAVIOUR half runs those exact seam calls, with those exact arguments,
//     against a temp project and asserts the on-disk values survive.
//
// Neither half alone is sufficient: source-only would not prove preservation
// actually happens, behaviour-only would assert against a fixture the test wrote
// itself rather than against what the wizard passes.

// removedWidgetMarkers are i18n/binding tokens that exist ONLY inside the removed
// question groups. Any of them reappearing in non-comment wizard code means a
// removed question was re-introduced.
var removedWidgetMarkers = []string{
	// statusline theme Select
	"StatuslineThemeTitle", "ThemeMoaiDark", "ThemeMoaiLight", "&statuslineTheme",
	// statusline segment MultiSelect
	"NewMultiSelect", "StatuslineSegmentsTitle", "&statuslineSegmentsSelection",
	// git convention Select
	"GitConventionTitle", "&gitConvention",
	// nested quality group
	"QualityCoverageTargetTitle", "QualityMinCoverageTitle", "QualityEnforceQualityTitle",
	"&nestedCoverageTarget", "&nestedMinCoverage", "&nestedEnforceQuality",
	// nested git auto-detection group
	"GitAutoEnabledTitle", "GitConfidenceTitle", "GitSampleSizeTitle", "GitEnforceOnPushTitle",
	"&nestedAutoDetection", "&nestedConfidence", "&nestedSampleSize", "&nestedEnforceOnPush",
	// the nested-config write path the wizard no longer drives
	"persistProjectNestedConfig", "readCurrentNestedConfig",
}

// TestWizardOmitsRemovedQuestions asserts the wizard source contains no widget for
// any removed setting. Comment lines are stripped so the explanatory prose in
// profile_setup.go (which names these groups) does not false-positive.
func TestWizardOmitsRemovedQuestions(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("profile_setup.go")
	if err != nil {
		t.Fatalf("read profile_setup.go: %v", err)
	}
	code := nonCommentLines(string(data))
	for _, marker := range removedWidgetMarkers {
		if strings.Contains(code, marker) {
			t.Errorf("profile_setup.go references %q — a removed wizard question appears to be back", marker)
		}
	}
}

// TestWizardWritesNoStatuslineTheme pins the product decision: the wizard does
// not manage the statusline theme AT ALL.
//
// History worth keeping, because it explains why this is a removal and not a
// default: a statusline whose text looked washed out was first attributed to
// this setting, and the knob was briefly pinned to MoAI Light. The colour turned
// out to come from the Claude Code theme instead — this setting was never the
// cause — so the knob was removed rather than given a value.
//
// The assertions therefore check for ABSENCE, which is what makes the removal
// stick: naming ANY theme value in the wizard source fails this test, so a
// future "just default it to X" edit cannot land silently.
func TestWizardWritesNoStatuslineTheme(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("profile_setup.go")
	if err != nil {
		t.Fatalf("read profile_setup.go: %v", err)
	}
	code := nonCommentLines(string(data))

	// No theme value may be assigned to the saved prefs. Leaving the field at its
	// zero value is what makes `omitempty` drop statusline_theme from
	// preferences.yaml and makes syncStatusline preserve statusline.yaml.
	if strings.Contains(code, "StatuslineTheme:") {
		t.Error("profile_setup.go assigns StatuslineTheme — the wizard must leave it " +
			"at the zero value so the theme is neither stored nor synced")
	}
	for _, theme := range []string{"catppuccin-mocha", "catppuccin-latte"} {
		if strings.Contains(code, theme) {
			t.Errorf("profile_setup.go still names the theme %q — the wizard no longer "+
				"manages the statusline theme in any form", theme)
		}
	}
}

// TestWizardCarriesStoredSegmentsIntoPrefs pins the preservation decision for the
// PROFILE store: preferences.yaml is rewritten wholesale by
// profile.WritePreferences (yaml.Marshal over the file), so the saved struct must
// carry the profile's existing segment map. Passing nil there would delete
// statusline_segments from preferences.yaml — removing a question must not blank
// the stored value.
func TestWizardCarriesStoredSegmentsIntoPrefs(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("profile_setup.go")
	if err != nil {
		t.Fatalf("read profile_setup.go: %v", err)
	}
	code := nonCommentLines(string(data))
	if !strings.Contains(code, "StatuslineSegments: existingPrefs.StatuslineSegments") {
		t.Error("the saved prefs must carry existingPrefs.StatuslineSegments through; a nil would blank preferences.yaml")
	}
	// ...and the PROJECT sync must get nil so statusline.yaml is left alone.
	if !strings.Contains(code, "syncPrefs.StatuslineSegments = nil") {
		t.Error("the project sync must receive a nil segment map so statusline.yaml segments are preserved")
	}
	// The git_convention argument to persistProjectConfig must stay empty
	// (persistProjectConfig writes only non-empty values).
	if !strings.Contains(code, `persistProjectConfig(cwd, developmentMode, "")`) {
		t.Error(`persistProjectConfig must be called with an empty convention so git-convention.yaml is preserved`)
	}
}

// TestWizardPreferencesRoundTripKeepsSegments is the behavioural half of the
// profile-store claim: writing the struct the wizard now builds (stored segments
// carried through) leaves statusline_segments intact in preferences.yaml.
func TestWizardPreferencesRoundTripKeepsSegments(t *testing.T) {
	base := t.TempDir()
	orig := profile.BaseDirOverride
	profile.BaseDirOverride = base
	t.Cleanup(func() { profile.BaseDirOverride = orig })

	const name = "default"
	stored := map[string]bool{"model": true, "context": false, "git_branch": true}
	if err := profile.WritePreferences(name, profile.ProfilePreferences{
		UserName:           "Goos",
		StatuslineSegments: stored,
		StatuslineTheme:    "catppuccin-mocha",
	}); err != nil {
		t.Fatalf("seed WritePreferences: %v", err)
	}

	existingPrefs, err := profile.ReadPreferences(name)
	if err != nil {
		t.Fatalf("ReadPreferences: %v", err)
	}

	// The struct runProfileSetup now composes for the statusline fields: the
	// segment map is carried through, and StatuslineTheme is left unset.
	saved := profile.ProfilePreferences{
		UserName:           "Goos",
		StatuslineSegments: existingPrefs.StatuslineSegments,
	}
	if err := profile.WritePreferences(name, saved); err != nil {
		t.Fatalf("WritePreferences: %v", err)
	}

	got, err := profile.ReadPreferences(name)
	if err != nil {
		t.Fatalf("re-read preferences: %v", err)
	}
	if got.StatuslineSegments == nil {
		t.Fatal("statusline_segments was blanked by the save — the wizard must carry the stored map through")
	}
	for key, want := range stored {
		if got.StatuslineSegments[key] != want {
			t.Errorf("segment %q = %v, want %v (stored value must survive a wizard save)", key, got.StatuslineSegments[key], want)
		}
	}
	// The wizard writes no theme, and preferences.yaml tags the field omitempty,
	// so the previously-stored catppuccin-mocha must be GONE rather than kept or
	// replaced. This is the "removed, not defaulted" half of the decision.
	if got.StatuslineTheme != "" {
		t.Errorf("StatuslineTheme = %q, want empty — the wizard must not store a "+
			"statusline theme at all", got.StatuslineTheme)
	}
}

// seedRemovedQuestionProject writes a temp project whose statusline.yaml,
// git-convention.yaml and quality.yaml all carry values the wizard used to
// overwrite, so a save can be checked for clobbering.
func seedRemovedQuestionProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		// A deliberately partial, hand-edited segment map: with the MultiSelect gone,
		// editing this file is the only way left to change segments.
		"statusline.yaml": "statusline:\n" +
			"  theme: catppuccin-mocha\n" +
			"  segments:\n" +
			"    model: true\n" +
			"    context: false\n" +
			"    pr: false\n",
		"quality.yaml": "constitution:\n" +
			"  development_mode: tdd\n" +
			"  enforce_quality: true\n" +
			"  test_coverage_target: 77\n" +
			"  tdd_settings:\n" +
			"    min_coverage_per_commit: 66\n",
		"git-convention.yaml": "git_convention:\n" +
			"  convention: angular\n" +
			"  auto_detection:\n" +
			"    enabled: true\n" +
			"    confidence_threshold: 0.6\n" +
			"    sample_size: 100\n" +
			"  validation:\n" +
			"    enforce_on_push: true\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(sectionsDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return root
}

// TestWizardSavePreservesStatuslineSegments is the behavioural half of the project
// claim for statusline.yaml: the sync the wizard now performs (nil segments, no
// theme) leaves BOTH a hand-edited segment map and the on-disk theme untouched.
//
// With every statusline widget removed, editing this file by hand is the only
// way left to change either value, so a wizard run must be a content no-op here.
func TestWizardSavePreservesStatuslineSegments(t *testing.T) {
	root := seedRemovedQuestionProject(t)

	// Exactly what runProfileSetup hands to profile.SyncToProjectConfig: no
	// segment map and no theme.
	syncPrefs := profile.ProfilePreferences{
		UserName: "Goos",
	}
	syncPrefs.StatuslineSegments = nil
	if err := profile.SyncToProjectConfig(root, syncPrefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "statusline.yaml"))
	if err != nil {
		t.Fatalf("read statusline.yaml: %v", err)
	}
	out := string(data)
	// The hand-edited segment map must survive verbatim.
	for _, want := range []string{"model: true", "context: false", "pr: false"} {
		if !strings.Contains(out, want) {
			t.Errorf("statusline.yaml lost the hand-edited segment %q; got:\n%s", want, out)
		}
	}
	// The wizard must not have expanded the partial map into a full 16-key map.
	if strings.Contains(out, "worktree:") || strings.Contains(out, "usage_5h:") {
		t.Errorf("statusline.yaml gained segments the wizard never asked about; got:\n%s", out)
	}
	// The seeded theme must survive verbatim. The wizard no longer manages the
	// statusline theme, so overwriting it here — with EITHER theme — is a
	// regression, which is why this asserts the seeded value rather than absence.
	if !strings.Contains(out, "theme: catppuccin-mocha") {
		t.Errorf("statusline.yaml lost its seeded theme catppuccin-mocha; the wizard "+
			"must not write a theme at all; got:\n%s", out)
	}
}

// TestWizardSavePreservesGitConventionAndNestedQuality is the behavioural half of
// the project claim for quality.yaml / git-convention.yaml: the only project write
// the wizard still performs is development_mode, and it leaves every value from the
// removed groups untouched.
func TestWizardSavePreservesGitConventionAndNestedQuality(t *testing.T) {
	root := seedRemovedQuestionProject(t)

	// Exactly what runProfileSetup hands to persistProjectConfig: a chosen
	// development_mode and an EMPTY convention.
	if err := persistProjectConfig(root, "ddd", ""); err != nil {
		t.Fatalf("persistProjectConfig: %v", err)
	}

	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	// The one value the wizard still collects.
	if string(cfg.Quality.DevelopmentMode) != "ddd" {
		t.Errorf("development_mode = %q, want ddd", cfg.Quality.DevelopmentMode)
	}
	// git_convention: removed from the wizard, must be untouched.
	if cfg.GitConvention.Convention != "angular" {
		t.Errorf("convention = %q, want angular (the wizard no longer asks — it must not clobber)", cfg.GitConvention.Convention)
	}

	// The 7 nested values: removed from the wizard, must all be untouched. Read back
	// through the shared seam the web console uses.
	cur, err := settings.ReadProjectNestedConfig(root)
	if err != nil {
		t.Fatalf("ReadProjectNestedConfig: %v", err)
	}
	for _, c := range []struct {
		field     string
		got, want any
	}{
		{"CoverageTarget", cur.CoverageTarget, "77"},
		{"MinCoverage", cur.MinCoverage, "66"},
		{"EnforceQuality", cur.EnforceQuality, true},
		{"AutoDetectionEnabled", cur.AutoDetectionEnabled, true},
		{"ConfidenceThreshold", cur.ConfidenceThreshold, "0.6"},
		{"SampleSize", cur.SampleSize, "100"},
		{"EnforceOnPush", cur.EnforceOnPush, true},
	} {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v (removed from the wizard — must be preserved)", c.field, c.got, c.want)
		}
	}
}
