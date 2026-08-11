package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This file is the SPEC-GATE-ASTGREP-REPAIR-001 regression guard for the D3
// defect: `moai gate` (internal/cli/gate.go) MUST load the ast-grep gate
// configuration from .moai/config/sections/gate.yaml via the shared
// config.loadGateSection loader, instead of hardcoding quality.DefaultGateConfig().
//
// AC-GAR-008 (REQ-GAR-006): the CLI gate path consults gate.yaml.
// AC-GAR-009 (REQ-GAR-007): warn_only_mode:true is honored (advisory, not hard-block).
// AC-GAR-010 (REQ-GAR-008): ast_grep_gate.enabled:false skips the ast-grep sub-scan.

// TestGAR_AC008_GateGoLoadsGateYAML (AC-GAR-008):
// gate.go source MUST route runGate through the loadGateCfgForCLI helper so
// gate.yaml is consulted. Before M3 runGate hardcoded
// quality.DefaultGateConfig() as its sole config source. The helper itself
// retains DefaultGateConfig as a fallback for missing/unparseable gate.yaml
// (AP-GAR-004), so this test asserts runGate calls the helper rather than
// asserting the literal absence of DefaultGateConfig anywhere in the file.
func TestGAR_AC008_GateGoLoadsGateYAML(t *testing.T) {
	src, err := os.ReadFile("gate.go")
	if err != nil {
		t.Fatalf("read gate.go: %v", err)
	}
	body := string(src)

	if !strings.Contains(body, "loadGateCfgForCLI(") {
		t.Errorf("AC-GAR-008: gate.go does not call loadGateCfgForCLI; " +
			"the standalone moai gate CLI path does not consult gate.yaml (D3 not wired)")
	}

	// runGate must obtain its cfg from the helper, not from a direct
	// DefaultGateConfig() call at the runGate call site. Match the line shape
	// that existed pre-M3: `cfg := quality.DefaultGateConfig()` appearing
	// directly inside runGate.
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "cfg := quality.DefaultGateConfig()") {
			t.Errorf("AC-GAR-008: runGate still assigns "+
				"`cfg := quality.DefaultGateConfig()` directly; "+
				"route through loadGateCfgForCLI instead. Line: %q", line)
		}
	}
}

// writeGateYAML writes a gate.yaml under <tmp>/.moai/config/sections/ with the
// given ast_grep_gate body.
func writeGateYAML(t *testing.T, tmp, astGrepBody string) {
	t.Helper()
	sections := filepath.Join(tmp, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	content := "gate:\n  enabled: true\n  ast_grep_gate:\n" + astGrepBody
	if err := os.WriteFile(filepath.Join(sections, "gate.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write gate.yaml: %v", err)
	}
}

// TestGAR_AC009_WarnOnlyModeReflectedInCLIConfig (AC-GAR-009 / REQ-GAR-007):
// when gate.yaml sets ast_grep_gate.warn_only_mode: true + block_on_error:
// false, the CLI gate config MUST reflect the advisory intent. This exercises
// the loadGateCfgForCLI helper extracted in M3 end-to-end against a real
// gate.yaml in a temp dir.
func TestGAR_AC009_WarnOnlyModeReflectedInCLIConfig(t *testing.T) {
	tmp := t.TempDir()
	writeGateYAML(t, tmp, "    enabled: true\n    block_on_error: false\n    warn_only_mode: true\n")

	cfg := loadGateCfgForCLI(tmp)
	if cfg.AstGrepGate == nil {
		t.Fatal("AC-GAR-009: cfg.AstGrepGate is nil after loading gate.yaml")
	}
	if !cfg.AstGrepGate.WarnOnlyMode {
		t.Errorf("AC-GAR-009: WarnOnlyMode = false; want true (gate.yaml advisory intent)")
	}
	if cfg.AstGrepGate.BlockOnError {
		t.Errorf("AC-GAR-009: BlockOnError = true; want false (gate.yaml advisory intent)")
	}
}

// TestGAR_AC010_EnabledFalseSkipsAstGrep (AC-GAR-010 / REQ-GAR-008):
// when gate.yaml sets ast_grep_gate.enabled: false, the CLI gate config MUST
// carry Enabled=false so the ast-grep sub-scan is skipped.
func TestGAR_AC010_EnabledFalseSkipsAstGrep(t *testing.T) {
	tmp := t.TempDir()
	writeGateYAML(t, tmp, "    enabled: false\n")

	cfg := loadGateCfgForCLI(tmp)
	if cfg.AstGrepGate == nil {
		t.Fatal("AC-GAR-010: cfg.AstGrepGate is nil after loading gate.yaml")
	}
	if cfg.AstGrepGate.Enabled {
		t.Errorf("AC-GAR-010: AstGrepGate.Enabled = true; want false (explicit opt-out)")
	}
}

// TestGAR_D3_FallbackOnMissingGateYAML (AP-GAR-004):
// when gate.yaml is absent, loadGateCfgForCLI MUST fall back to the hardcoded
// default (NOT silently hard-block). The fallback path returns a non-nil cfg
// whose ProjectDir matches the input — the gate still runs, it just runs with
// the production-default config rather than failing closed.
func TestGAR_D3_FallbackOnMissingGateYAML(t *testing.T) {
	tmp := t.TempDir()
	// No gate.yaml written — entirely empty project dir.

	cfg := loadGateCfgForCLI(tmp)
	if cfg == nil {
		t.Fatal("AP-GAR-004: loadGateCfgForCLI returned nil for missing gate.yaml")
	}
	if cfg.ProjectDir != tmp {
		t.Errorf("AP-GAR-004: ProjectDir = %q; want %q", cfg.ProjectDir, tmp)
	}
	if cfg.AstGrepGate == nil {
		t.Error("AP-GAR-004: fallback cfg.AstGrepGate is nil; expected the default config")
	}
}
