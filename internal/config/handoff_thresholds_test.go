package config

import "testing"

// TestHandoffThresholdConstants locks the SPEC-HANDOFF-THRESHOLD-001 (M4) band
// boundary constants. These are compiled-in (NOT config-overridable) per the
// "M3 lands / M4 consumes" contract; renderer.go references them by name so no
// inline band literals remain (§14 hardcoding-prevention). AC-THRESHOLD-004.
func TestHandoffThresholdConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  int
		want int
	}{
		{"HandoffSoftLargePct", HandoffSoftLargePct, 50},
		{"HandoffSoftStandardPct", HandoffSoftStandardPct, 90},
		{"HandoffLargeWindowCutoff", HandoffLargeWindowCutoff, 500000},
		{"HandoffHardCeilingCapPct", HandoffHardCeilingCapPct, 95},
		{"HandoffHardCeilingMarginPct", HandoffHardCeilingMarginPct, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Fatalf("%s = %d, want %d", tt.name, tt.got, tt.want)
			}
		})
	}
}
