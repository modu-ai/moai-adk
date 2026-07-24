package uikit

import "github.com/modu-ai/moai-adk/internal/tui"

// StatusIcon returns a colored Unicode icon for the check status.
//
// Moved from internal/cli/doctor.go as part of the uikit kernel extraction
// (SPEC-CLI-UIKIT-KERNEL-001). Generic status→icon renderer consumed by
// RenderStatusLine in this package. Glyphs resolve from the canonical tui
// glyph SSOT (REQ-TUXIU-003).
func StatusIcon(s CheckStatus) string {
	switch s {
	case CheckOK:
		return SuccessStyle.Render(string(tui.GlyphDone))
	case CheckWarn:
		return WarnStyle.Render(string(tui.GlyphWarn))
	case CheckFail:
		return ErrorStyle.Render(string(tui.GlyphErr))
	default:
		return "?"
	}
}
