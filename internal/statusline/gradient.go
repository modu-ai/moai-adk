package statusline

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// interpolateGradientColor returns the gradient RGB value for a block position (0.0~1.0).
// Path: Green(0,255,0) → Yellow(255,255,0) → Red(255,0,0)
//
//	0.0~0.5: Green → Yellow
//	0.5~1.0: Yellow → Red
func interpolateGradientColor(blockPct float64) (r, g, b int) {
	// Clamp out-of-range input
	if blockPct <= 0 {
		return 0, 255, 0
	}
	if blockPct >= 1 {
		return 255, 0, 0
	}

	if blockPct <= 0.5 {
		// Green → Yellow: R increases, G stays 255, B=0
		t := blockPct * 2 // 0.0 ~ 1.0
		r = int(float64(255) * t)
		g = 255
		b = 0
	} else {
		// Yellow → Red: R stays 255, G decreases, B=0
		t := (blockPct - 0.5) * 2 // 0.0 ~ 1.0
		r = 255
		g = int(float64(255) * (1.0 - t))
		b = 0
	}
	return
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
	if width <= 0 {
		return ""
	}

	// Calculate filled block count (max width)
	filled := min((pct*width)/100, width)
	empty := width - filled

	filledChar := "█" // Used block
	emptyChar := "░"  // Remaining block

	// Return plain string in noColor mode or when no filled blocks
	if noColor || filled == 0 {
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
		r, g, b := interpolateGradientColor(blockPct)
		hex := fmt.Sprintf("#%02X%02X%02X", r, g, b)
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Render(filledChar))
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
