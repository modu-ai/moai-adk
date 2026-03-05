package statusline

import "testing"

// TestNormalizeMode verifies backward-compatible mode name normalization.
// REQ-V3-MODE-001: "minimal" → "compact" conversion
// REQ-V3-MODE-002: "verbose" → "full" conversion
func TestNormalizeMode(t *testing.T) {
	tests := []struct {
		name  string
		input StatuslineMode
		want  StatuslineMode
	}{
		// Backward compatibility: old name → new name conversion
		{"minimal converts to compact", "minimal", ModeCompact},
		{"verbose converts to full", "verbose", ModeFull},
		// Current names remain unchanged
		{"default unchanged", "default", ModeDefault},
		{"compact unchanged", "compact", ModeCompact},
		{"full unchanged", "full", ModeFull},
		// Edge cases
		{"empty unchanged", "", ""},
		{"unknown unchanged", "custom", "custom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeMode(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeMode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
