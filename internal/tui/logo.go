package tui

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// moaiLogoArt is the restored 6-line "MoAI-ADK" ASCII-art logo, byte-identical
// to the const retired by SPEC-CLI-TUX-V3-004 REQ-TUX4-006 (commit 77893579e).
// The leading and trailing newline reproduce the original const shape. The
// block/box-drawing runes (█ ╗ ╔ ╚ ╝ ═ ║) are a decorative-art category exempt
// from the AC-CLI-TUI-017 status-glyph whitelist (REQ-TUXIU-056); they predate
// the whitelist and introduce no emoji-range codepoint.
const moaiLogoArt = `
███╗   ███╗          █████╗ ██╗       █████╗ ██████╗ ██╗  ██╗
████╗ ████║ ██████╗ ██╔══██╗██║      ██╔══██╗██╔══██╗██║ ██╔╝
██╔████╔██║██║   ██║███████║██║█████╗███████║██║  ██║█████╔╝
██║╚██╔╝██║██║   ██║██╔══██║██║╚════╝██╔══██║██║  ██║██╔═██╗
██║ ╚═╝ ██║╚██████╔╝██║  ██║██║      ██║  ██║██████╔╝██║  ██╗
╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝      ╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝
`

// CoralRamp returns an n-stop vertical coral gradient derived from the theme
// accent tokens (Accent → AccentDeep) via lipgloss.Blend1D (CIELAB linear
// interpolation). The ramp is the single colour source for the logo's per-line
// gradient; no hex literal is used — the endpoints are Theme tokens
// (REQ-TUXIU-051).
//
// CoralRamp returns the gradient COLOURS; whether they are displayed is a
// render-time decision (Logo applies NO_COLOR / non-TTY degradation). When the
// active theme resolves to monochrome (NO_COLOR set → empty accent tokens), the
// ramp falls back to LightTheme accents so the returned stops remain the brand
// coral gradient rather than collapsing to a single non-colour.
//
// @MX:ANCHOR: [AUTO] CoralRamp is the logo gradient SSOT; fan_in via Logo + M3 PrintBanner
// @MX:REASON: [AUTO] the coral ramp endpoints come only from Theme.Accent/AccentDeep,
// keeping all colour decisions inside the internal/tui theme-SSOT boundary (REQ-TUXIU-051)
func CoralRamp(n int) []color.Color {
	th := ResolveOS()
	if th.Accent == "" { // resolved to monochrome (NO_COLOR); use coral endpoints
		th = LightTheme()
	}
	return coralRamp(th, n)
}

// coralRamp builds the n-stop coral ramp from an explicit theme. Both the
// exported CoralRamp (active theme) and Logo (caller-supplied theme) route
// through here, so the gradient math lives in exactly one place.
func coralRamp(th Theme, n int) []color.Color {
	if n < 1 {
		return nil
	}
	return lipgloss.Blend1D(n, lipgloss.Color(th.Accent), lipgloss.Color(th.AccentDeep))
}

// Logo renders the restored 6-line MoAI-ADK logo with a per-row vertical coral
// gradient — top light → bottom deep — each row painted with one successive
// CoralRamp stop (row i ← ramp[i], REQ-TUXIU-052). Colour is sourced ONLY from
// the supplied theme's accent tokens; this file contains zero hex literals
// (REQ-TUXIU-051).
//
// Degradation (REQ-TUXIU-053):
//   - A monochrome theme (Accent == "", the theme resolved under NO_COLOR)
//     renders plain art with zero ANSI colour.
//   - Non-TTY / reduced-palette output is downsampled by the shared downsample
//     helper (the same path Spinner/Progress/Stepper use): a truecolor terminal
//     shows the full per-row gradient; a pipe collapses to <=1 distinct colour.
//
// The logo is static art — no animation is introduced, so MOAI_REDUCED_MOTION
// has no logo-specific effect.
//
// Logo is a foundation primitive: this M1 helper CREATES the logo but does NOT
// wire it into uikit.PrintBanner (the 3-surface stacking is M3).
//
// @MX:ANCHOR: [AUTO] Logo is the restored large-banner primitive; fan_in via M3 PrintBanner (3 surfaces)
// @MX:REASON: [AUTO] single logo render surface consumed by root/init/update PrintBanner stacking
func Logo(theme Theme) string {
	rows := strings.Split(strings.Trim(moaiLogoArt, "\n"), "\n")

	// Monochrome theme → plain art, no colour (NO_COLOR path).
	if theme.Accent == "" {
		return strings.Join(rows, "\n") + "\n"
	}

	ramp := coralRamp(theme, len(rows))
	var b strings.Builder
	for i, row := range rows {
		style := lipgloss.NewStyle().Foreground(ramp[i])
		b.WriteString(downsample(style.Render(row)))
		b.WriteByte('\n')
	}
	return b.String()
}
