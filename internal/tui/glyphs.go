package tui

// Canonical status-glyph runes — the single source of truth for the status
// glyph vocabulary rendered across the tui, printer, and uikit packages
// (REQ-TUXIU-001). Every render site that emits one of these glyphs resolves
// it from one of these constants; no package redeclares the raw rune literal.
// Callers that need a string use string(GlyphX).
//
// All runes are within the AC-CLI-TUI-017 status-glyph whitelist
// (✓ ✗ ! · ● ○ ◆ ◇); no emoji-range codepoint is introduced (REQ-TUXIU-004).
// The large-logo art (internal/tui/logo.go) is a SEPARATE decorative-rune
// category, exempt from this status-glyph vocabulary (REQ-TUXIU-056).
//
// @MX:ANCHOR: [AUTO] Glyph* is the canonical status-glyph SSOT; fan_in >= 4
// across tui.StatusIcon, printer step/spinner markers, and uikit sym helpers
// @MX:REASON: [AUTO] consolidates the previously-duplicated ✓/✗/●/○/!/·
// literals so a glyph change happens in exactly one place (REQ-TUXIU-001..003)
const (
	GlyphDone rune = '✓' // ok / done / success
	GlyphErr  rune = '✗' // error / fail
	GlyphRun  rune = '●' // running / in-progress
	GlyphSkip rune = '○' // pending / skip
	GlyphWarn rune = '!' // warning
	GlyphInfo rune = '·' // info / neutral dot
)
