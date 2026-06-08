package web

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// writeProjectSections seeds a temp project root with the in-scope config
// sections plus the out-of-scope sections for round-trip / isolation tests.
// SPEC-WEB-CONSOLE-003 M2.
func writeProjectSections(t *testing.T, devMode, convention string) string {
	t.Helper()
	root := t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"quality.yaml":        "constitution:\n  development_mode: " + devMode + "\n  test_coverage_target: 85\n",
		"git-convention.yaml": "git_convention:\n  convention: " + convention + "\n  validation:\n    max_length: 72\n",
		// Out-of-scope sections — must round-trip unchanged.
		"workflow.yaml":     "workflow:\n  sentinel: DO_NOT_TOUCH\n",
		"harness.yaml":      "harness:\n  sentinel: DO_NOT_TOUCH\n",
		"git-strategy.yaml": "git_strategy:\n  sentinel: DO_NOT_TOUCH\n",
		"llm.yaml":          "llm:\n  mode: \"\"\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(sectionsDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return root
}

// readProjectSectionValues reads the persisted development_mode + convention via
// the config manager (LoadRaw round-trip), the same path the app read seam uses.
func readProjectSectionValues(t *testing.T, root string) (devMode, convention string) {
	t.Helper()
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	return string(cfg.Quality.DevelopmentMode), cfg.GitConvention.Convention
}

// TestValidateProjectConfig covers REQ-WC3-001/002: empty allowed, canonical
// allowed, out-of-list rejected with the correct field key. Case sensitivity
// (EC-4) and whitespace (EC-6) are non-canonical → error.
func TestValidateProjectConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		devMode    string
		convention string
		wantKeys   []string // field keys expected in the error map (empty = no error)
	}{
		{"both empty", "", "", nil},
		{"both canonical", "ddd", "angular", nil},
		{"dev tdd, conv auto", "tdd", "auto", nil},
		{"conv conventional-commits", "", "conventional-commits", nil},
		{"conv karma", "", "karma", nil},
		{"conv custom rejected (engine removed)", "", "custom", []string{"git_convention"}},
		{"bogus dev only", "xyz", "angular", []string{"development_mode"}},
		{"bogus conv only", "ddd", "gitflow", []string{"git_convention"}},
		{"both bogus", "xyz", "gitflow", []string{"development_mode", "git_convention"}},
		{"EC-4 uppercase dev", "TDD", "", []string{"development_mode"}},
		{"EC-4 uppercase conv", "", "Angular", []string{"git_convention"}},
		{"EC-6 whitespace dev", " ", "", []string{"development_mode"}},
		{"EC-6 whitespace conv", "", " ", []string{"git_convention"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs := validateProjectConfig(tt.devMode, tt.convention)
			if len(errs) != len(tt.wantKeys) {
				t.Fatalf("validateProjectConfig(%q,%q) = %v, want keys %v", tt.devMode, tt.convention, errs, tt.wantKeys)
			}
			for _, k := range tt.wantKeys {
				if _, ok := errs[k]; !ok {
					t.Errorf("expected error key %q, got map %v", k, errs)
				}
			}
		})
	}
}

// TestReadProjectConfig covers REQ-WC3-004: the real read seam returns the
// persisted development_mode + convention from quality.yaml/git-convention.yaml.
func TestReadProjectConfig(t *testing.T) {
	t.Parallel()
	root := writeProjectSections(t, "ddd", "karma")

	devMode, convention, err := readProjectConfig(root)
	if err != nil {
		t.Fatalf("readProjectConfig: %v", err)
	}
	if devMode != "ddd" {
		t.Errorf("devMode = %q, want ddd", devMode)
	}
	if convention != "karma" {
		t.Errorf("convention = %q, want karma", convention)
	}
}

// TestReadProjectConfig_AbsentFiles covers EC-5: missing config files → the
// read seam returns the LoadRaw default values (compiled-in defaults: tdd / auto),
// not a panic and not an error. The exact defaults are owned by internal/config;
// the contract this test pins is "no panic, no error, defaults applied".
func TestReadProjectConfig_AbsentFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir() // no .moai/config/sections at all

	devMode, convention, err := readProjectConfig(root)
	if err != nil {
		t.Fatalf("readProjectConfig on absent config should not error: %v", err)
	}
	// LoadRaw applies compiled-in defaults when the sections dir is absent.
	// The read seam must surface whatever the config manager returns (no panic);
	// it must NOT invent its own empty-string behavior. Cross-check against the
	// same LoadRaw round-trip the seam itself uses.
	wantDev, wantConv := readProjectSectionValues(t, root)
	if devMode != wantDev {
		t.Errorf("devMode = %q, want %q (LoadRaw default round-trip)", devMode, wantDev)
	}
	if convention != wantConv {
		t.Errorf("convention = %q, want %q (LoadRaw default round-trip)", convention, wantConv)
	}
}

// TestWriteProjectConfig covers REQ-WC3-005/007: the real write seam persists
// non-empty values via the config manager round-trip, leaves a non-targeted
// quality field unchanged, and creates the section on a fresh project (EC-5).
func TestWriteProjectConfig(t *testing.T) {
	t.Parallel()

	t.Run("persists both values", func(t *testing.T) {
		t.Parallel()
		root := writeProjectSections(t, "tdd", "auto")
		if err := writeProjectConfig(root, "ddd", "angular"); err != nil {
			t.Fatalf("writeProjectConfig: %v", err)
		}
		devMode, convention := readProjectSectionValues(t, root)
		if devMode != "ddd" {
			t.Errorf("devMode = %q, want ddd", devMode)
		}
		if convention != "angular" {
			t.Errorf("convention = %q, want angular", convention)
		}
	})

	t.Run("empty keeps existing (EC-1)", func(t *testing.T) {
		t.Parallel()
		root := writeProjectSections(t, "tdd", "auto")
		// dev=ddd, convention empty → quality becomes ddd, git-convention stays auto.
		if err := writeProjectConfig(root, "ddd", ""); err != nil {
			t.Fatalf("writeProjectConfig: %v", err)
		}
		devMode, convention := readProjectSectionValues(t, root)
		if devMode != "ddd" {
			t.Errorf("devMode = %q, want ddd", devMode)
		}
		if convention != "auto" {
			t.Errorf("convention = %q, want auto (empty submission must not clobber)", convention)
		}
	})

	t.Run("non-targeted quality field preserved", func(t *testing.T) {
		t.Parallel()
		root := writeProjectSections(t, "tdd", "auto")
		if err := writeProjectConfig(root, "ddd", ""); err != nil {
			t.Fatalf("writeProjectConfig: %v", err)
		}
		mgr := config.NewConfigManager()
		cfg, err := mgr.LoadRaw(root)
		if err != nil {
			t.Fatalf("LoadRaw: %v", err)
		}
		if cfg.Quality.TestCoverageTarget != 85 {
			t.Errorf("test_coverage_target = %d, want 85 (must round-trip unchanged)", cfg.Quality.TestCoverageTarget)
		}
	})

	t.Run("absent config creates section (EC-5)", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		if err := writeProjectConfig(root, "ddd", "angular"); err != nil {
			t.Fatalf("writeProjectConfig on fresh project: %v", err)
		}
		devMode, convention := readProjectSectionValues(t, root)
		if devMode != "ddd" || convention != "angular" {
			t.Errorf("fresh write yielded devMode=%q convention=%q, want ddd/angular", devMode, convention)
		}
	})
}

// TestDefaultAppHasProjectConfigSeams covers AC-WC3-005: newApp wires real,
// non-nil readProjectConfig / writeProjectConfig seams.
func TestDefaultAppHasProjectConfigSeams(t *testing.T) {
	t.Parallel()
	a := newApp(Config{ProjectRoot: t.TempDir()})
	if a.readProjectConfig == nil {
		t.Error("newApp must wire a non-nil readProjectConfig seam")
	}
	if a.writeProjectConfig == nil {
		t.Error("newApp must wire a non-nil writeProjectConfig seam")
	}
}
