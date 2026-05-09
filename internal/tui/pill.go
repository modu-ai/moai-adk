// Package tui provides the MoAI-ADK terminal UI design system v2.
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// PillKind identifies the semantic variant of a Pill.
// Each kind maps to a foreground+background colour pair from Theme.
type PillKind string

const (
	// PillInfo is the informational pill variant.
	PillInfo PillKind = "info"
	// PillOk is the success / passing pill variant.
	PillOk PillKind = "ok"
	// PillWarn is the warning / caution pill variant.
	PillWarn PillKind = "warn"
	// PillErr is the error / danger pill variant.
	PillErr PillKind = "err"
	// PillPrimary is the brand-accent pill variant.
	PillPrimary PillKind = "primary"
	// PillNeutral is the muted / neutral pill variant.
	PillNeutral PillKind = "neutral"
)

// PillOpts configures a Pill render call.
type PillOpts struct {
	// Kind selects the semantic colour pair. Defaults to PillInfo.
	Kind PillKind
	// Solid uses a filled background when true; outline (tinted bg) when false.
	Solid bool
	// Label is the text content of the pill.
	Label string
	// Theme provides the design tokens. If nil, LightTheme() is used.
	Theme *Theme
}

// pillColors returns the foreground and background colour strings for a given
// PillKind and Theme, following the tui.jsx Pill component colour map:
//
//	info:    [t.info, t.infoSoft]
//	ok:      [t.success, t.successSoft]
//	warn:    [t.warning, t.warningSoft]
//	err:     [t.danger, t.dangerSoft]
//	primary: [t.accent, t.accentSoft]
//	neutral: [t.dim, t.ruleSoft]
func pillColors(k PillKind, th Theme) (fg, bg string) {
	switch k {
	case PillOk:
		return th.Success, th.SuccessSoft
	case PillWarn:
		return th.Warning, th.WarningSoft
	case PillErr:
		return th.Danger, th.DangerSoft
	case PillPrimary:
		return th.Accent, th.AccentSoft
	case PillNeutral:
		return th.Dim, th.RuleSoft
	default: // PillInfo
		return th.Info, th.InfoSoft
	}
}

// Pill renders a compact rounded label — semantically equivalent to the CSS
// pill / badge pattern in tui.jsx (Pill component, lines 157-177).
//
// In solid mode the foreground colour becomes the background and the label
// renders in the terminal's default foreground. In outline mode the background
// is the soft tint and the foreground is the full-saturation accent colour.
//
// All colours are sourced from opts.Theme; no hex literals appear in this
// file (REQ-CLI-TUI-013).
//
// @MX:ANCHOR: [AUTO] Pill is a core primitive; expected fan_in >= 5 across M3-M6
// @MX:REASON: doctor summary, banner, version card, status all use Pill per design source
func Pill(opts PillOpts) string {
	th := LightTheme()
	if opts.Theme != nil {
		th = *opts.Theme
	}

	kind := opts.Kind
	if kind == "" {
		kind = PillInfo
	}

	fgHex, bgHex := pillColors(kind, th)

	// In monochrome mode (NO_COLOR), all hex values are empty.
	// lipgloss.Color("") produces no ANSI output, so the pill degrades to
	// plain bracketed text.
	if fgHex == "" && bgHex == "" {
		return "[" + opts.Label + "]"
	}

	var style lipgloss.Style
	if opts.Solid {
		// Solid: filled background = fgHex, label in bg colour for contrast.
		// For neutral, use Bg colour (dark/light bg); for others use th.Bg.
		labelColor := th.Bg
		if labelColor == "" {
			labelColor = th.Fg
		}
		style = lipgloss.NewStyle().
			Background(lipgloss.Color(fgHex)).
			Foreground(lipgloss.Color(labelColor)).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)
	} else {
		// Outline: soft tinted background + saturated foreground.
		// RGBA backgrounds are not supported by all terminals; we use
		// lipgloss.Color which handles hex but not rgba strings.
		// For soft/rgba tokens we fall back to a styled foreground-only pill.
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(fgHex)).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)
	}

	return style.Render(opts.Label)
}
