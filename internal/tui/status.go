// Package tui provides the MoAI-ADK terminal UI design system v2.
package tui

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusIcon returns a single-glyph status indicator for the given kind.
// Allowed kinds: "ok", "warn", "err", "info", "run", "skip", "dot".
// Unknown kinds fall back to the dot glyph.
// All glyphs are in the AC-CLI-TUI-017 whitelist (✓ ✗ ! · ● ○ ◆ ◇).
//
// @MX:ANCHOR: [AUTO] StatusIcon is a core primitive; expected fan_in >= 4 across M4-M5
// @MX:REASON: doctor CheckLine, banner, init wizard, and status command all use StatusIcon
func StatusIcon(kind string) string {
	switch kind {
	case "ok":
		return "✓"
	case "warn":
		return "!"
	case "err":
		return "✗"
	case "info":
		return "·"
	case "run":
		return "●"
	case "skip":
		return "○"
	default: // "dot" and unknown
		return "·"
	}
}

// reducedMotion returns true when MOAI_REDUCED_MOTION is set to a non-empty,
// non-"0" value (AC-CLI-TUI-015).
func reducedMotion() bool {
	v := os.Getenv("MOAI_REDUCED_MOTION")
	return v != "" && v != "0"
}

// Spinner returns a stateless spinner label for the given text.
// When MOAI_REDUCED_MOTION=1, a static dot (●) is produced instead of the
// animated frame character (AC-CLI-TUI-015).
// th may be nil; LightTheme is used in that case.
//
// Spinner is stateless: no goroutines, no tickers. The caller is responsible
// for re-rendering on each display tick.
func Spinner(label string, th *Theme) string {
	t := LightTheme()
	if th != nil {
		t = *th
	}

	var frame string
	if reducedMotion() {
		frame = "●"
	} else {
		// Default animated frame — caller re-renders; we always return frame 0.
		// The choice of "⠋" is a safe Unicode spinner character outside the
		// emoji codepoint ranges (AC-CLI-TUI-017); it is in Braille Patterns
		// U+2800–U+28FF, which is not in the EMOJI_RANGES list.
		frame = "⠋"
	}

	spinStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Accent))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Body))

	return spinStyle.Render(frame) + " " + labelStyle.Render(label)
}

// ProgressOpts configures a Progress render call.
type ProgressOpts struct {
	// Theme provides design tokens. If nil, LightTheme is used.
	Theme *Theme
	// Width is the total bar width in columns. Defaults to 20.
	Width int
}

// Progress renders a stateless progress bar.
// value is the current progress value; max is the maximum value.
// When MOAI_REDUCED_MOTION=1, a fully filled bar is rendered instead of a
// partial one (AC-CLI-TUI-015 — no animated fill transition).
//
// Progress is stateless: no goroutines, no tickers.
//
// @MX:ANCHOR: [AUTO] Progress is a core primitive; expected fan_in >= 3 across M4-M6
// @MX:REASON: status command, loop stepper, and update command use Progress
func Progress(value, max int, opts ProgressOpts) string {
	t := LightTheme()
	if opts.Theme != nil {
		t = *opts.Theme
	}

	width := opts.Width
	if width <= 0 {
		width = 20
	}

	var filled int
	if reducedMotion() {
		// Static fallback: fully filled bar regardless of progress value.
		filled = width
	} else {
		if max > 0 {
			filled = (value * width) / max
			if filled > width {
				filled = width
			}
		}
	}
	empty := width - filled

	filledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Accent))
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Rule))

	bar := filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", empty))

	return bar
}

// Stepper renders a compact step indicator like "● 1 / 6 ○ ○ ○ ○ ○".
// current is 1-based; total is the total number of steps.
// th may be nil; LightTheme is used in that case.
//
// @MX:ANCHOR: [AUTO] Stepper is a core primitive; expected fan_in >= 3 in M5-M6
// @MX:REASON: init wizard (M5) and loop start (M6) both display step progress
func Stepper(current, total int, th *Theme) string {
	t := LightTheme()
	if th != nil {
		t = *th
	}

	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Accent)).Bold(true)
	doneStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Accent))
	futureStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Faint))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Dim))

	var b strings.Builder
	for i := 1; i <= total; i++ {
		if i > 1 {
			b.WriteString(" ")
		}
		switch {
		case i < current:
			b.WriteString(doneStyle.Render("●"))
		case i == current:
			b.WriteString(activeStyle.Render("●"))
		default:
			b.WriteString(futureStyle.Render("○"))
		}
	}

	label := labelStyle.Render(" " + itoa(current) + " / " + itoa(total))
	return b.String() + label
}

// itoa converts an int to a string without importing strconv at top-level.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits [20]byte
	pos := 20
	for n > 0 {
		pos--
		digits[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		digits[pos] = '-'
	}
	return string(digits[pos:])
}
