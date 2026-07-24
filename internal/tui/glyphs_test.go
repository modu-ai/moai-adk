package tui

import "testing"

// TestGlyphConstants pins the canonical status-glyph runes (REQ-TUXIU-001).
// These constants are the single source of truth for the status-glyph
// vocabulary; every render site in tui/printer/uikit resolves its glyph
// from one of them rather than redeclaring the raw rune literal.
func TestGlyphConstants(t *testing.T) {
	cases := []struct {
		name string
		got  rune
		want rune
	}{
		{"GlyphDone", GlyphDone, '✓'},
		{"GlyphErr", GlyphErr, '✗'},
		{"GlyphRun", GlyphRun, '●'},
		{"GlyphSkip", GlyphSkip, '○'},
		{"GlyphWarn", GlyphWarn, '!'},
		{"GlyphInfo", GlyphInfo, '·'},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}

// TestStatusIconResolvesFromGlyphConstants asserts StatusIcon returns the
// canonical constants (not its own private rune literals), so there is
// exactly ONE glyph source (REQ-TUXIU-001, AC-TUXIU-004).
func TestStatusIconResolvesFromGlyphConstants(t *testing.T) {
	cases := []struct {
		kind string
		want rune
	}{
		{"ok", GlyphDone},
		{"warn", GlyphWarn},
		{"err", GlyphErr},
		{"info", GlyphInfo},
		{"run", GlyphRun},
		{"skip", GlyphSkip},
		{"dot", GlyphInfo},
		{"unknown-kind-falls-back-to-dot", GlyphInfo},
	}
	for _, c := range cases {
		if got := StatusIcon(c.kind); got != string(c.want) {
			t.Errorf("StatusIcon(%q) = %q, want %q", c.kind, got, string(c.want))
		}
	}
}
