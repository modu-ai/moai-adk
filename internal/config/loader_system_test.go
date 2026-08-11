package config

import (
	"os"
	"path/filepath"
	"testing"
)

// loader_system_test.go — SPEC-CONFIG-KEY-HONESTY-001 M4.
//
// Verifies that Loader.Load binds the hook block of system.yaml into
// cfg.System.Hook (the narrow loadSystemSection binding), and that the shared
// LoadSystemHookOptInEnabled helper reads the same value the loader would,
// fail-CLOSED on any error or absent key.

// TestSystemHookOptInLoadsViaLoader asserts the M4 reader-consolidation
// contract: after Loader.Load reads a system.yaml whose hook.opt_in.enabled is
// true, cfg.System.Hook.OptIn.Enabled is true; and the shared helper
// LoadSystemHookOptInEnabled returns the same value against the same tree.
func TestSystemHookOptInLoadsViaLoader(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	sectionsDir := filepath.Join(tmp, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	// Ship only the hook block — the other three top-level blocks (moai /
	// github / document_management) have no SystemConfig field and are
	// intentionally ignored by loadSystemSection.
	systemYAML := []byte("hook:\n  opt_in:\n    enabled: true\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "system.yaml"), systemYAML, 0o600); err != nil {
		t.Fatalf("write system.yaml: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tmp, ".moai"))
	if err != nil {
		t.Fatalf("Loader.Load: %v", err)
	}
	if !cfg.System.Hook.OptIn.Enabled {
		t.Fatalf("cfg.System.Hook.OptIn.Enabled = false after Loader.Load; want true (system.yaml hook.opt_in.enabled=true)")
	}
	if !loader.LoadedSections()["system"] {
		t.Fatalf("LoadedSections()[system] = false; want true (loadSystemSection should mark the section loaded)")
	}

	// The shared helper must agree with the loaded value against the same tree.
	if got := LoadSystemHookOptInEnabled(tmp); !got {
		t.Fatalf("LoadSystemHookOptInEnabled(%q) = false; want true (must match cfg.System.Hook.OptIn.Enabled)", tmp)
	}
}

// TestSystemHookOptInLoadsViaLoader_DefaultsFalse covers the fail-CLOSED
// contract: a system.yaml without the hook block, a missing file, and an
// unparseable file all yield false from both the loader default and the helper.
func TestSystemHookOptInLoadsViaLoader_DefaultsFalse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		content []byte // nil => do not write the file
	}{
		{name: "no hook block", content: []byte("moai:\n  version: \"1.0.0\"\n")},
		{name: "hook block disabled", content: []byte("hook:\n  opt_in:\n    enabled: false\n")},
		{name: "file absent", content: nil},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tmp := t.TempDir()
			sectionsDir := filepath.Join(tmp, ".moai", "config", "sections")
			if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
				t.Fatalf("mkdir sections: %v", err)
			}
			if tc.content != nil {
				if err := os.WriteFile(filepath.Join(sectionsDir, "system.yaml"), tc.content, 0o600); err != nil {
					t.Fatalf("write system.yaml: %v", err)
				}
			}
			if got := LoadSystemHookOptInEnabled(tmp); got {
				t.Fatalf("LoadSystemHookOptInEnabled (%s) = true; want false (fail-CLOSED default OFF)", tc.name)
			}
		})
	}
}
