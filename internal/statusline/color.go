package statusline

import (
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// SGR sequences framing every color-enabled statusline line (card t1822).
const (
	// sgrNormalIntensity cancels any faint/bold attribute the host carries
	// into the line, so the text is never rendered dim.
	sgrNormalIntensity = "\x1b[22m"
	// sgrReset is the reset lipgloss/termenv closes every styled run with.
	sgrReset = "\x1b[0m"
)

// statuslineColorProfile picks the color profile the statusline forces.
// Claude Code captures the command's stdout through a pipe, so TTY detection
// always resolves to the Ascii profile; the profile is therefore chosen from
// COLORTERM alone: TrueColor for "truecolor"/"24bit", ANSI256 otherwise.
func statuslineColorProfile(colorterm string) termenv.Profile {
	switch strings.ToLower(strings.TrimSpace(colorterm)) {
	case "truecolor", "24bit":
		return termenv.TrueColor
	default:
		return termenv.ANSI256
	}
}

// newForcedColorRenderer returns a lipgloss renderer owned by the statusline
// with its profile forced, leaving the process-wide default renderer (and so
// every other command) untouched.
func newForcedColorRenderer(colorterm string) *lipgloss.Renderer {
	lr := lipgloss.NewRenderer(io.Discard)
	lr.SetColorProfile(statuslineColorProfile(colorterm))
	return lr
}

// foregroundSequence returns the SGR sequence setting hex as the foreground
// under the renderer's profile.
func foregroundSequence(lr *lipgloss.Renderer, hex lipgloss.Color) string {
	return termenv.CSI + lr.ColorProfile().Color(string(hex)).Sequence(false) + "m"
}

// applyBaseColor gives every rune of s an explicit foreground: each line opens
// with SGR 22 plus the base text color, the base color is re-applied after
// every reset an inner styled run emits, and the line closes with a reset so
// no attribute leaks past it. Visible text is unchanged.
func applyBaseColor(s, baseSeq string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		line = strings.ReplaceAll(line, sgrReset, sgrReset+baseSeq)
		lines[i] = sgrNormalIntensity + baseSeq + line + sgrReset
	}
	return strings.Join(lines, "\n")
}

// barStyle carries the color inputs of the usage bars; a nil *barStyle
// renders them plain (NO_COLOR).
type barStyle struct {
	lg    *lipgloss.Renderer
	stops []string
	muted lipgloss.Color
}

// newBarStyle binds a renderer to the theme's gradient stops and muted color.
func newBarStyle(lg *lipgloss.Renderer, theme Theme) *barStyle {
	return &barStyle{lg: lg, stops: barGradientStops(theme), muted: theme.Muted()}
}

// placeholderBar renders an all-empty bar of width blocks in the muted color,
// standing in for a usage window whose reading is n/a (card t1822). It never
// carries a fill or a percentage (REQ-SOU-004).
func placeholderBar(width int, bs *barStyle) string {
	if width <= 0 {
		return ""
	}
	blocks := strings.Repeat("░", width)
	if bs == nil {
		return blocks
	}
	return bs.lg.NewStyle().Foreground(bs.muted).Render(blocks)
}
