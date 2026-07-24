package config

import (
	"os"
	"path/filepath"
	"testing"
)

// loader_gate_test.go — gate.yaml section loader tests.
//
// The gate section previously had a registry entry and a GateConfig struct but
// no loader path, so AstGrepGate.Enabled could never become true from config.
// These tests prove the loader completes the pair while keeping the ast-grep
// sub-gate OFF by default (opt-in enable).

// writeGateYAML writes a gate.yaml into a temp .moai/config/sections dir and
// returns the .moai dir suitable for Loader.Load.
func writeGateYAML(t *testing.T, content string) string {
	t.Helper()
	moaiDir := t.TempDir()
	sectionsDir := filepath.Join(moaiDir, "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if content != "" {
		if err := os.WriteFile(filepath.Join(sectionsDir, "gate.yaml"), []byte(content), 0o644); err != nil {
			t.Fatalf("write gate.yaml: %v", err)
		}
	}
	return moaiDir
}

// TestLoadGateSection_AstGrepEnable verifies a gate.yaml with
// ast_grep_gate.enabled: true loads into AstGrepGate.Enabled=true.
func TestLoadGateSection_AstGrepEnable(t *testing.T) {
	moaiDir := writeGateYAML(t, "gate:\n  ast_grep_gate:\n    enabled: true\n    block_on_error: true\n")

	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Gate.AstGrepGate.Enabled {
		t.Error("AstGrepGate.Enabled = false, want true when gate.yaml enables it")
	}
	if !cfg.Gate.AstGrepGate.BlockOnError {
		t.Error("AstGrepGate.BlockOnError = false, want true from gate.yaml")
	}
	if !loader.LoadedSections()["gate"] {
		t.Error("loadedSections[gate] = false, want true when gate.yaml present")
	}
}

// TestLoadGateSection_DefaultOff is the default-OFF characterization:
// with no gate.yaml the ast-grep sub-gate stays disabled and the outer
// gate defaults are unchanged (byte-identical pre-existing behavior).
func TestLoadGateSection_DefaultOff(t *testing.T) {
	moaiDir := writeGateYAML(t, "")

	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Gate.AstGrepGate.Enabled {
		t.Error("AstGrepGate.Enabled = true, want false by default (opt-in invariant)")
	}
	if !cfg.Gate.Enabled {
		t.Error("Gate.Enabled = false, want true default (unchanged)")
	}
	if cfg.Gate.Timeouts.Vet != 30 || cfg.Gate.Timeouts.Lint != 60 || cfg.Gate.Timeouts.Test != 120 {
		t.Errorf("Gate.Timeouts = %+v, want defaults 30/60/120", cfg.Gate.Timeouts)
	}
	if loader.LoadedSections()["gate"] {
		t.Error("loadedSections[gate] = true, want false when gate.yaml absent")
	}
}

// TestLoadGateSection_PartialOverride verifies a gate.yaml that sets only one
// key keeps the seeded defaults for omitted keys (partial-override contract,
// parallel to loadArchiveSection).
func TestLoadGateSection_PartialOverride(t *testing.T) {
	moaiDir := writeGateYAML(t, "gate:\n  skip_tests: true\n")

	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Gate.SkipTests {
		t.Error("Gate.SkipTests = false, want true from gate.yaml")
	}
	if cfg.Gate.Timeouts.Vet != 30 {
		t.Errorf("Gate.Timeouts.Vet = %d, want seeded default 30", cfg.Gate.Timeouts.Vet)
	}
	if cfg.Gate.AstGrepGate.Enabled {
		t.Error("AstGrepGate.Enabled = true, want false (not set in yaml)")
	}
}
