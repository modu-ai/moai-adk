package config

import "testing"

// === SPEC-CC2178-MODEL-POLICY-REPAIR-001 M3 (RED tests for ResolveCycleType) ===
//
// These tests assert the ResolveCycleType contract defined in plan.md §F.3.1.
// They are written BEFORE the symbol is authored (TDD RED phase) and reference
// a function that does not yet exist in the package. They FAIL at compile time
// until cycle_type.go provides the implementation.
//
// AC bindings: AC-MPR-004 (minimal->ddd), AC-MPR-005 (thorough->tdd),
// AC-MPR-006 (explicit pin preserved), AC-MPR-014 (standard->tdd).

func TestResolveCycleType(t *testing.T) {
	tests := []struct {
		name         string
		harnessLevel string
		explicitPin  string
		want         string
	}{
		// AC-MPR-004: minimal harness resolves to lightweight ddd (NOT tdd)
		{"minimal_no_pin", "minimal", "", "ddd"},
		// AC-MPR-005: thorough harness resolves to full tdd
		{"thorough_no_pin", "thorough", "", "tdd"},
		// AC-MPR-014: standard harness resolves to tdd (current default unchanged)
		{"standard_no_pin", "standard", "", "tdd"},
		// AC-MPR-006: explicit development_mode pin wins over harness level (AG-01 backward-compat)
		{"minimal_with_tdd_pin", "minimal", "tdd", "tdd"},
		{"thorough_with_ddd_pin", "thorough", "ddd", "ddd"},
		{"standard_with_tdd_pin", "standard", "tdd", "tdd"},
		// Unknown/empty harness level: safe fallback to tdd (never returns empty)
		{"empty_level_no_pin", "", "", "tdd"},
		{"unknown_level_no_pin", "nonexistent-level", "", "tdd"},
		// Explicit pin wins even when harness level is unknown
		{"unknown_level_with_pin", "nonexistent-level", "ddd", "ddd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveCycleType(tt.harnessLevel, tt.explicitPin)
			if got != tt.want {
				t.Errorf("ResolveCycleType(%q, %q) = %q, want %q",
					tt.harnessLevel, tt.explicitPin, got, tt.want)
			}
		})
	}
}

// TestResolveCycleType_AlwaysReturnsNonEmpty verifies the function never
// returns an empty string (the safe fallback is always "tdd").
func TestResolveCycleType_AlwaysReturnsNonEmpty(t *testing.T) {
	cases := []struct {
		level string
		pin   string
	}{
		{"", ""},
		{"minimal", ""},
		{"standard", ""},
		{"thorough", ""},
		{"garbage", ""},
		{"", "tdd"},
		{"", "ddd"},
	}
	for _, c := range cases {
		if got := ResolveCycleType(c.level, c.pin); got == "" {
			t.Errorf("ResolveCycleType(%q, %q) returned empty string; must always return a non-empty cycle_type", c.level, c.pin)
		}
	}
}
