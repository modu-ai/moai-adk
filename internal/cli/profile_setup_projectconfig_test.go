package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// seedProjectConfig writes a temp project with the in-scope + out-of-scope
// sections so persistProjectConfig round-trips can be asserted.
// SPEC-WEB-CONSOLE-003 M4.
func seedProjectConfig(t *testing.T, devMode, convention string) string {
	t.Helper()
	root := t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"quality.yaml":        "constitution:\n  development_mode: " + devMode + "\n  test_coverage_target: 85\n",
		"git-convention.yaml": "git_convention:\n  convention: " + convention + "\n",
		"workflow.yaml":       "workflow:\n  sentinel: DO_NOT_TOUCH\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(sectionsDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return root
}

// TestPersistProjectConfig_LandsInProjectConfig covers AC-WC3-006b: the TUI save
// path persists development_mode + git_convention into quality.yaml /
// git-convention.yaml via the config manager.
func TestPersistProjectConfig_LandsInProjectConfig(t *testing.T) {
	root := seedProjectConfig(t, "tdd", "auto")

	if err := persistProjectConfig(root, "ddd", "angular"); err != nil {
		t.Fatalf("persistProjectConfig: %v", err)
	}

	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	if string(cfg.Quality.DevelopmentMode) != "ddd" {
		t.Errorf("development_mode = %q, want ddd", cfg.Quality.DevelopmentMode)
	}
	if cfg.GitConvention.Convention != "angular" {
		t.Errorf("convention = %q, want angular", cfg.GitConvention.Convention)
	}
}

// TestPersistProjectConfig_NotInPreferences covers AC-WC3-006b: the two
// project-config values are NOT ProfilePreferences fields. The ProfilePreferences
// struct has no slot for development_mode/convention — this test documents that
// structural guarantee by reading back the profile and asserting no such state.
func TestPersistProjectConfig_NotInPreferences(t *testing.T) {
	base := t.TempDir()
	orig := profile.BaseDirOverride
	profile.BaseDirOverride = base
	t.Cleanup(func() { profile.BaseDirOverride = orig })

	// Write a profile with the standard fields.
	const name = "default"
	if err := profile.WritePreferences(name, profile.ProfilePreferences{
		UserName:       "Goos",
		PermissionMode: "acceptEdits",
	}); err != nil {
		t.Fatalf("WritePreferences: %v", err)
	}

	root := seedProjectConfig(t, "tdd", "auto")
	if err := persistProjectConfig(root, "ddd", "angular"); err != nil {
		t.Fatalf("persistProjectConfig: %v", err)
	}

	// Read back the profile preferences file content and confirm it carries no
	// development_mode / convention keys (structurally impossible — the struct has
	// no such fields — but the test pins the contract).
	prefsPath := profile.GetPreferencesPath(name)
	data, err := os.ReadFile(prefsPath)
	if err != nil {
		t.Fatalf("read preferences.yaml: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "development_mode") {
		t.Errorf("preferences.yaml must NOT contain development_mode; got:\n%s", content)
	}
	if strings.Contains(content, "convention") {
		t.Errorf("preferences.yaml must NOT contain convention; got:\n%s", content)
	}
}

// TestPersistProjectConfig_EmptyKeepsExisting covers EC-1 parity for the TUI: an
// empty submitted value leaves the existing persisted value unchanged.
func TestPersistProjectConfig_EmptyKeepsExisting(t *testing.T) {
	root := seedProjectConfig(t, "tdd", "auto")
	if err := persistProjectConfig(root, "ddd", ""); err != nil {
		t.Fatalf("persistProjectConfig: %v", err)
	}
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	if string(cfg.Quality.DevelopmentMode) != "ddd" {
		t.Errorf("development_mode = %q, want ddd", cfg.Quality.DevelopmentMode)
	}
	if cfg.GitConvention.Convention != "auto" {
		t.Errorf("convention = %q, want auto (empty submission must not clobber)", cfg.GitConvention.Convention)
	}
}

// TestPersistProjectConfig_ReadCurrent covers the wizard init read: the helper
// that reads current project-config values returns the persisted values.
func TestPersistProjectConfig_ReadCurrent(t *testing.T) {
	root := seedProjectConfig(t, "ddd", "karma")
	devMode, convention, err := readCurrentProjectConfig(root)
	if err != nil {
		t.Fatalf("readCurrentProjectConfig: %v", err)
	}
	if devMode != "ddd" {
		t.Errorf("devMode = %q, want ddd", devMode)
	}
	if convention != "karma" {
		t.Errorf("convention = %q, want karma", convention)
	}
}

// TestProfileSetupConstructsProjectSelects is the AC-ITI-010 S5 guard,
// re-aimed per design.md §10 from a source grep to a behavior test over the
// ABSORBED question set: the profile wizard asks development_mode offering the
// canonical option values. The git_convention half of this guard moved to a
// NEGATIVE assertion in profile_setup_removed_questions_test.go: that Select
// was removed from the wizard, so binding it again is now the regression, not
// the requirement.
func TestProfileSetupConstructsProjectSelects(t *testing.T) {
	t.Parallel()
	opts := buildProfileOptions(getProfileText("en"))
	qs := wizard.ProfileQuestions(opts, wizard.ProfileResult{})
	q := wizard.QuestionByID(qs, "development_mode")
	if q == nil {
		t.Fatal("the absorbed profile question set has no development_mode question")
	}
	offered := map[string]bool{}
	for _, o := range q.Options {
		offered[o.Value] = true
	}
	for _, v := range []string{"", "ddd", "tdd"} {
		if !offered[v] {
			t.Errorf("development_mode select does not offer canonical option value %q", v)
		}
	}

	// The canonical option values must also be OFFERED by the shared schema
	// the option list derives from — a stronger check than the wizard-side
	// list alone: it fails if the schema stops offering a value.
	txt := getProfileText("en")
	schemaOffered := map[string]bool{}
	for _, o := range schemaSelectOptions(txt, "development_mode", false) {
		schemaOffered[o.Value] = true
	}
	for _, v := range []string{"ddd", "tdd"} {
		if !schemaOffered[v] {
			t.Errorf("schema development_mode options do not offer canonical value %q", v)
		}
	}
}
