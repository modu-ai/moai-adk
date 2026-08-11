package cli

import (
	"os"
	"strings"
	"testing"
)

// This file is the SPEC-GATE-ASTGREP-REPAIR-001 M0 RED guard for the D3
// defect: `moai gate` (internal/cli/gate.go) MUST load the ast-grep gate
// configuration from .moai/config/sections/gate.yaml via the shared
// config.loadGateSection loader, instead of hardcoding quality.DefaultGateConfig().
//
// AC-GAR-008 (REQ-GAR-006): the CLI gate path consults gate.yaml.
//
// M0 scope: this grep test is the binary RED signal — it compiles against the
// pre-M3 source and fails because gate.go contains neither loadGateSection
// nor any other gate.yaml loader call. Behavioral coverage (AC-GAR-009/010)
// is added in M3 alongside the extracted loader helper.

// TestGAR_AC008_GateGoLoadsGateYAML (AC-GAR-008):
// gate.go source MUST reference the gate.yaml loader (loadGateSection or an
// equivalent config.Loader path). Before M3 the source hardcoded
// quality.DefaultGateConfig() with no config read; this grep test fails RED
// until the wiring lands.
func TestGAR_AC008_GateGoLoadsGateYAML(t *testing.T) {
	src, err := os.ReadFile("gate.go")
	if err != nil {
		t.Fatalf("read gate.go: %v", err)
	}
	body := string(src)

	// A loadGate* reference (the shared loader or a thin wrapper) MUST be
	// present. Pre-M3 gate.go has none.
	if !strings.Contains(body, "loadGate") {
		t.Errorf("AC-GAR-008: gate.go does not reference any loadGate* loader; " +
			"gate.yaml is ignored by the standalone moai gate CLI path (D3 not wired)")
	}

	// The unconditional hardcoded default as the SOLE config source MUST be
	// gone. Pre-M3 gate.go has exactly `cfg := quality.DefaultGateConfig()`.
	if strings.Contains(body, "cfg := quality.DefaultGateConfig()") {
		t.Errorf("AC-GAR-008: gate.go still hardcodes `cfg := quality.DefaultGateConfig()` " +
			"as the sole config source; D3 not wired")
	}
}
