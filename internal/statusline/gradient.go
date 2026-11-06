package statusline

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	tuipkg "github.com/modu-ai/moai-adk/internal/tui"
)

// barGradientStops returns the theme's usage-bar palette stops, green →
// yellow → peach → red (card t1822: pastel Catppuccin stops replace the
// saturated pure-RGB path; Latte keeps its darker stops for light backgrounds).
func barGradientStops(theme Theme) []string {
	if _, ok := theme.(*catppuccinLatte); ok {
		return []string{tuipkg.CatppuccinLatteSuccess, tuipkg.CatppuccinLatteWarning,
			tuipkg.CatppuccinLattePeach, tuipkg.CatppuccinLatteDanger}
	}
	return []string{tuipkg.CatppuccinMochaSuccess, tuipkg.CatppuccinMochaWarning,
		tuipkg.CatppuccinMochaPeach, tuipkg.CatppuccinMochaDanger}
}

// interpolateGradientColor returns the RGB value for a block position
// (0.0~1.0) interpolated across evenly spaced hex stops. Out-of-range input
// clamps to the first or last stop.
func interpolateGradientColor(stops []string, blockPct float64) (r, g, b int) {
	segments := len(stops) - 1
	pos := min(max(blockPct, 0), 1) * float64(segments)
	i := min(int(pos), segments-1)
	f := pos - float64(i)
	r0, g0, b0 := hexRGB(stops[i])
	r1, g1, b1 := hexRGB(stops[i+1])
	lerp := func(a, b int) int { return int(math.Round(float64(a) + (float64(b)-float64(a))*f)) }
	return lerp(r0, r1), lerp(g0, g1), lerp(b0, b1)
}

// hexRGB parses a "#RRGGBB" color; a malformed value reads as black.
func hexRGB(hex string) (r, g, b int) {
	v, err := strconv.ParseUint(strings.TrimPrefix(hex, "#"), 16, 32)
	if err != nil {
		return 0, 0, 0
	}
	return int(v >> 16 & 0xff), int(v >> 8 & 0xff), int(v & 0xff)
}

// BuildGradientBar generates a progress bar with continuous RGB gradient.
//
// pct: usage percentage (0~100)
// width: total number of blocks in the bar
// noColor: when true, outputs only unicode block characters without ANSI escapes
//
// Each filled block gets an individual RGB color (REQ-V3-BAR-002).
// Returns plain block string when noColor or width <= 0 (REQ-V3-BAR-003).
//
// @MX:ANCHOR: [AUTO] Core function for all usage bar (CW/5H/7D) rendering - called from renderer.go
// @MX:REASON: [AUTO] fan_in >= 3 (renderUsageBar → called from 3 bar rendering paths)
func BuildGradientBar(pct int, width int, noColor bool) string {
	if noColor {
		return buildGradientBar(pct, width, nil)
	}
	return buildGradientBar(pct, width, newBarStyle(lipgloss.DefaultRenderer(), NewTheme("")))
}

// buildGradientBar renders the bar with bs's renderer and theme stops; a nil
// bs renders plain blocks. The statusline passes its own forced-profile
// renderer so the bar stays colored on a piped stdout (card t1822).
func buildGradientBar(pct int, width int, bs *barStyle) string {
	if width <= 0 {
		return ""
	}

	// Calculate filled block count (max width)
	filled := min((pct*width)/100, width)
	empty := width - filled

	filledChar := "█" // Used block
	emptyChar := "░"  // Remaining block

	// Return plain string in noColor mode or when no filled blocks
	if bs == nil || filled == 0 {
		return strings.Repeat(filledChar, filled) + strings.Repeat(emptyChar, empty)
	}

	// Apply individual gradient color per block
	var sb strings.Builder
	for i := 0; i < filled; i++ {
		// blockPct: 0.0 (first block) ~ 1.0 (last block)
		var blockPct float64
		if filled > 1 {
			blockPct = float64(i) / float64(filled-1)
		}
		r, g, b := interpolateGradientColor(bs.stops, blockPct)
		hex := fmt.Sprintf("#%02X%02X%02X", r, g, b)
		sb.WriteString(bs.lg.NewStyle().Foreground(lipgloss.Color(hex)).Render(filledChar))
	}
	sb.WriteString(strings.Repeat(emptyChar, empty))

	return sb.String()
}

// BatteryIcon returns a battery icon based on usage percentage.
// <= 70%: 🔋, > 70%: 🪫
//
// @MX:NOTE: [AUTO] 70% threshold per AC-V3-13 - update BatteryIcon tests in usage_test.go if changed
func BatteryIcon(pct int) string {
	if pct > 70 {
		return "🪫"
	}
	return "🔋"
}
