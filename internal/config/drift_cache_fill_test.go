package config

// drift_cache_fill_test.go — AC-DCF-011: the opt-out resolves through the
// CONFIG LOADER, never through the presence of text in a YAML file.
//
// The mutant this excludes is the one that would actually ship: the YAML key
// added to both workflow.yaml files with no typed field behind it. It decodes
// into nothing, the opt-out does not exist, and the feature is permanently on —
// while every text-presence check passes.

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeDriftFillWorkflowYAML(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if body == "" {
		return
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
}

// TestDriftCacheFillEnabledResolvesThroughLoader covers both halves of
// AC-DCF-011: (a) a project declaring no key at all resolves to enabled, and
// (b) an explicit `enabled: false` resolves to disabled. Both read the resolved
// config value.
func TestDriftCacheFillEnabledResolvesThroughLoader(t *testing.T) {
	cases := []struct {
		name string
		body string
		want bool
	}{
		{name: "key absent entirely", body: "workflow:\n    default_mode: \"\"\n", want: true},
		{name: "no workflow.yaml at all", body: "", want: true},
		{name: "explicit false", body: "workflow:\n    drift_cache_fill:\n        enabled: false\n", want: false},
		{name: "explicit true", body: "workflow:\n    drift_cache_fill:\n        enabled: true\n", want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeDriftFillWorkflowYAML(t, root, tc.body)
			if got := DriftCacheFillEnabledForRoot(root); got != tc.want {
				t.Fatalf("DriftCacheFillEnabledForRoot = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestDriftCacheFillEnabledNilConfig asserts the fail-open posture: a caller
// that could not load config gets the feature rather than silent inactivity.
func TestDriftCacheFillEnabledNilConfig(t *testing.T) {
	var cfg *Config
	if !cfg.DriftCacheFillEnabled() {
		t.Fatal("nil *Config DriftCacheFillEnabled() = false, want true (fail-open)")
	}
	if !DriftCacheFillEnabledForRoot("") {
		t.Fatal("empty root = disabled, want enabled (fail-open)")
	}
}

// TestDriftCacheFillDefaultsCarryTheField is the continued-firing carrier for
// AC-DCF-011(a): the construction-time default is asserted directly, so
// removing the defaults.go entry (which would leave the zero value false and
// ship the feature permanently OFF) turns this red.
func TestDriftCacheFillDefaultsCarryTheField(t *testing.T) {
	cfg := NewDefaultConfig()
	if !cfg.Workflow.DriftCacheFill.Enabled {
		t.Fatal("NewDefaultConfig().Workflow.DriftCacheFill.Enabled = false, want true")
	}
	if !cfg.DriftCacheFillEnabled() {
		t.Fatal("NewDefaultConfig().DriftCacheFillEnabled() = false, want true")
	}
}

// TestDriftCacheFillTTLCoversTheChildDeadline is AC-DCF-008(c), asserted over
// the RESOLVED constants rather than over literals.
//
// A suppression TTL shorter than the child's own deadline lets a second session
// reclaim a record whose child is still legitimately computing, which
// reintroduces the very burst the record exists to suppress — one TTL later.
func TestDriftCacheFillTTLCoversTheChildDeadline(t *testing.T) {
	if DefaultDriftCacheFillTimeout <= 0 {
		t.Fatalf("DefaultDriftCacheFillTimeout = %v, want a positive duration", DefaultDriftCacheFillTimeout)
	}
	if DriftCacheFillTTLSlack <= 0 {
		t.Fatalf("DriftCacheFillTTLSlack = %v, want a positive duration", DriftCacheFillTTLSlack)
	}
	min := DefaultDriftCacheFillTimeout + DriftCacheFillTTLSlack
	if DefaultDriftCacheFillTTL < min {
		t.Fatalf("DefaultDriftCacheFillTTL = %v, want >= deadline+slack = %v", DefaultDriftCacheFillTTL, min)
	}
	// Guard against a degenerate "slack" that makes the inequality vacuous.
	if DriftCacheFillTTLSlack < time.Second {
		t.Fatalf("DriftCacheFillTTLSlack = %v, want at least 1s of real headroom", DriftCacheFillTTLSlack)
	}
}
