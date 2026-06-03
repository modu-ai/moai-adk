package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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

// TestProfileSetupConstructsProjectSelects is a construction grep guard
// (AC-WC3-006a): the wizard source must bind a development_mode and a
// git_convention huh.Select with the canonical option values.
func TestProfileSetupConstructsProjectSelects(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("profile_setup.go")
	if err != nil {
		t.Fatalf("read profile_setup.go: %v", err)
	}
	src := string(data)
	// The two selects must be bound to the value variables.
	for _, marker := range []string{"&developmentMode", "&gitConvention"} {
		if !strings.Contains(src, marker) {
			t.Errorf("profile_setup.go must bind a select to %s", marker)
		}
	}
	// The canonical convention values must appear as huh.NewOption values.
	for _, v := range []string{`"ddd"`, `"tdd"`, `"auto"`, `"conventional-commits"`, `"angular"`, `"karma"`, `"custom"`} {
		if !strings.Contains(src, v) {
			t.Errorf("profile_setup.go missing canonical option value %s", v)
		}
	}
}
