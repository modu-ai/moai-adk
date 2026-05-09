// Package tui provides the MoAI-ADK terminal UI design system v2.
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// BorderKind selects the lipgloss border style used by Box and ThickBox.
type BorderKind string

const (
	// BorderRounded uses lipgloss.RoundedBorder (default).
	BorderRounded BorderKind = "rounded"
	// BorderThick uses lipgloss.ThickBorder for heavy-weight containers.
	BorderThick BorderKind = "thick"
	// BorderDashed uses lipgloss.NormalBorder to approximate a dashed look.
	// lipgloss does not natively support CSS dashed borders; NormalBorder is
	// used as the closest semantic alternative.
	BorderDashed BorderKind = "dashed"
)

// BoxOpts configures a Box or ThickBox render call.
type BoxOpts struct {
	// Title is the optional label shown inset into the top border.
	Title string
	// Body is the text content rendered inside the box.
	Body string
	// Width is the total outer width in terminal columns. 0 means auto.
	Width int
	// Theme provides the design tokens for colour selection.
	// If nil, LightTheme() is used.
	Theme *Theme
	// Accent applies the accent colour to the border and background when true.
	Accent bool
	// Border selects the border kind. Defaults to BorderRounded.
	Border BorderKind
}

// theme returns the effective theme, defaulting to LightTheme if opts.Theme is nil.
func (o BoxOpts) theme() Theme {
	if o.Theme == nil {
		t := LightTheme()
		return t
	}
	return *o.Theme
}

// borderStyle returns the lipgloss.Border for the given BorderKind.
func borderStyle(k BorderKind) lipgloss.Border {
	switch k {
	case BorderThick:
		return lipgloss.ThickBorder()
	case BorderDashed:
		return lipgloss.NormalBorder()
	default:
		return lipgloss.RoundedBorder()
	}
}

// borderColor returns the border foreground colour for the given theme and accent flag.
func borderColor(th Theme, accent bool) lipgloss.Color {
	if accent {
		return lipgloss.Color(th.Accent)
	}
	return lipgloss.Color(th.Rule)
}

// Box renders a rounded-border panel using lipgloss.Border().
// The border colour is set to Theme.Rule by default, or Theme.Accent when
// Accent is true. Padding is 1 line / 2 columns (matches tui.jsx padding:
// "12px 16px" mapped to cell units).
//
// All colours are sourced from the Theme argument; no hex literal appears
// in this file (REQ-CLI-TUI-013).
//
// @MX:ANCHOR: [AUTO] Box is a core primitive; expected fan_in >= 5 across M3-M6
// @MX:REASON: All 14 CLI command surfaces call Box for panel/card rendering
func Box(opts BoxOpts) string {
	return renderBox(opts, false)
}

// ThickBox renders a heavy-bordered panel using lipgloss.ThickBorder().
// It is used for emphasis containers such as the version card or error output.
//
// @MX:ANCHOR: [AUTO] ThickBox is a core primitive; expected fan_in >= 3 across M4-M6
// @MX:REASON: version/error/doctor summary surfaces use ThickBox per design source
func ThickBox(opts BoxOpts) string {
	// Override border kind to thick regardless of opts.Border.
	opts.Border = BorderThick
	return renderBox(opts, true)
}

// renderBox is the shared implementation for Box and ThickBox.
// thick selects ThickBorder and adjusts font weight styling.
func renderBox(opts BoxOpts, thick bool) string {
	th := opts.theme()

	bk := opts.Border
	if thick {
		bk = BorderThick
	} else if bk == "" {
		bk = BorderRounded
	}

	border := borderStyle(bk)
	bc := borderColor(th, opts.Accent)

	style := lipgloss.NewStyle().
		Border(border).
		BorderForeground(bc).
		Padding(1, 2).
		Foreground(lipgloss.Color(th.Fg))

	if opts.Width > 0 {
		// Width includes border (1 cell each side) and padding (2 cells each side).
		// lipgloss.Style.Width sets the inner content width.
		innerW := max(opts.Width-2-4, 1) // 2 border cols + 4 padding cols, min 1
		style = style.Width(innerW)
	}

	content := opts.Body
	if opts.Title != "" {
		content = opts.Title + "\n" + content
	}

	return style.Render(content)
}
