package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeDecisionGateFile writes a minimal interview.yaml carrying the given
// decision_gate value into t.TempDir() and returns its path.
func writeDecisionGateFile(t *testing.T, gateLine string) string {
	t.Helper()
	content := "interview:\n  enabled: true\n  clarity_threshold: 4\n" + gateLine
	path := filepath.Join(t.TempDir(), "interview.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return path
}

// TestDecisionGateDefaultIsOff pins the distributed default: absent key
// resolves to off, and defaultInterviewConfig carries the explicit off value.
// Maps to: REQ-DA-001, REQ-DA-002, REQ-DA-004; AC-DA-001/003.
func TestDecisionGateDefaultIsOff(t *testing.T) {
	def := defaultInterviewConfig()
	if def.DecisionGate != "off" {
		t.Errorf("defaultInterviewConfig().DecisionGate: want \"off\", got %q", def.DecisionGate)
	}
	if got := def.ResolvedDecisionGate(); got != "off" {
		t.Errorf("default resolved gate: want \"off\", got %q", got)
	}
}

// TestDecisionGateAbsentKeyIsOff verifies that a section file without the
// decision_gate key resolves to off (the exact status quo at base).
// Maps to: REQ-DA-002; AC-DA-002.
func TestDecisionGateAbsentKeyIsOff(t *testing.T) {
	path := writeDecisionGateFile(t, "")
	cfg, err := LoadInterviewConfig(path)
	if err != nil {
		t.Fatalf("LoadInterviewConfig(absent key): unexpected error: %v", err)
	}
	if cfg.DecisionGate != "" {
		t.Errorf("DecisionGate on file without the key: want \"\", got %q", cfg.DecisionGate)
	}
	if got := cfg.ResolvedDecisionGate(); got != "off" {
		t.Errorf("resolved gate on file without the key: want \"off\", got %q", got)
	}
}

// TestDecisionGateOnRoundTrip verifies that a section file setting
// decision_gate: on resolves to on through the loader.
// Maps to: REQ-DA-001; AC-DA-002.
func TestDecisionGateOnRoundTrip(t *testing.T) {
	path := writeDecisionGateFile(t, "  decision_gate: on\n")
	cfg, err := LoadInterviewConfig(path)
	if err != nil {
		t.Fatalf("LoadInterviewConfig(on): unexpected error: %v", err)
	}
	if cfg.DecisionGate != "on" {
		t.Errorf("DecisionGate: want \"on\", got %q", cfg.DecisionGate)
	}
	if got := cfg.ResolvedDecisionGate(); got != "on" {
		t.Errorf("resolved gate: want \"on\", got %q", got)
	}
}

// TestDecisionGateUnknownValueResolvesOff verifies the fail-safe triple on a
// bogus value: the load succeeds, the resolved gate is off, and the
// unrecognized value is recorded verbatim rather than discarded.
// Maps to: REQ-DA-003; AC-DA-002.
func TestDecisionGateUnknownValueResolvesOff(t *testing.T) {
	path := writeDecisionGateFile(t, "  decision_gate: bogus\n")
	cfg, err := LoadInterviewConfig(path)
	if err != nil {
		t.Fatalf("LoadInterviewConfig(bogus): unexpected error (REQ-DA-003 requires the load to succeed): %v", err)
	}
	if cfg.DecisionGate != "bogus" {
		t.Errorf("DecisionGate: want verbatim \"bogus\", got %q", cfg.DecisionGate)
	}
	if got := cfg.ResolvedDecisionGate(); got != "off" {
		t.Errorf("resolved gate: want \"off\" fallback, got %q", got)
	}
}

// TestDecisionGateEmptyAndOffLiteralResolveOff covers the two remaining
// off-direction inputs: the empty string (behaves like absence) and the
// explicit "off" literal. Case variants are unrecognized (no normalization).
// Maps to: REQ-DA-002, REQ-DA-003.
func TestDecisionGateEmptyAndOffLiteralResolveOff(t *testing.T) {
	for _, line := range []string{"  decision_gate: \"\"\n", "  decision_gate: off\n", "  decision_gate: On\n"} {
		path := writeDecisionGateFile(t, line)
		cfg, err := LoadInterviewConfig(path)
		if err != nil {
			t.Fatalf("LoadInterviewConfig(%q): unexpected error: %v", line, err)
		}
		if got := cfg.ResolvedDecisionGate(); got != "off" {
			t.Errorf("resolved gate for %q: want \"off\", got %q", line, got)
		}
	}
}

// TestDecisionGateOrthogonalToRecommendationMode asserts the axis
// independence explicitly (REQ-DA-018): setting one key and leaving the other
// absent must resolve each axis independently, and no resolver may read the
// other axis's field. Maps to: AC-DA-017.
func TestDecisionGateOrthogonalToRecommendationMode(t *testing.T) {
	// decision_gate on + recommendation_mode absent: gate resolves on, mode
	// resolves to its own default push.
	onGate := defaultInterviewConfig()
	onGate.DecisionGate = "on"
	if got := onGate.ResolvedDecisionGate(); got != "on" {
		t.Errorf("gate with mode absent: want \"on\", got %q", got)
	}
	if got := onGate.ResolvedRecommendationMode(); got != "push" {
		t.Errorf("mode with gate on and mode absent: want \"push\", got %q", got)
	}

	// recommendation_mode pull + decision_gate absent: mode resolves pull,
	// gate resolves off.
	pullMode := defaultInterviewConfig()
	pullMode.RecommendationMode = "pull"
	if got := pullMode.ResolvedRecommendationMode(); got != "pull" {
		t.Errorf("mode with gate absent: want \"pull\", got %q", got)
	}
	if got := pullMode.ResolvedDecisionGate(); got != "off" {
		t.Errorf("gate with mode pull and gate absent: want \"off\", got %q", got)
	}
}
