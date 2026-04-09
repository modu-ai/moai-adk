package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Style definitions (lipgloss-based)
var (
	headerStyle = lipgloss.NewStyle().Bold(true).Underline(true)
	greenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))  // green (improvement)
	redStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))  // red (regression)
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // gray (no change)
	mustStyle   = lipgloss.NewStyle().Bold(true)                       // MUST label
)

// RenderDashboard renders the full terminal dashboard including
// progress bar, experiment statistics, and per-criterion breakdown.
func RenderDashboard(data *DashboardData) string {
	if data == nil {
		return ""
	}

	var b strings.Builder

	// Header
	b.WriteString(headerStyle.Render("Research Dashboard"))
	b.WriteString("\n\n")

	// Target and score summary
	fmt.Fprintf(&b,"  Target: %s\n", data.Target)

	// Current score + percentage
	scorePct := int(data.CurrentScore * 100)
	targetPct := int(data.TargetScore * 100)
	fmt.Fprintf(&b,"  Score:  %d%% / %d%% (target)\n", scorePct, targetPct)

	// Delta display (current - baseline)
	delta := data.CurrentScore - data.Baseline
	deltaPct := int(delta * 100)
	var deltaStr string
	if delta > 0 {
		deltaStr = greenStyle.Render(fmt.Sprintf("+%d%%", deltaPct))
	} else if delta < 0 {
		deltaStr = redStyle.Render(fmt.Sprintf("%d%%", deltaPct))
	} else {
		deltaStr = dimStyle.Render("0%")
	}
	fmt.Fprintf(&b,"  Delta:  %s from baseline\n", deltaStr)

	// Overall progress bar
	scoreRatio := data.CurrentScore
	fmt.Fprintf(&b,"  Progress: %s %d%%\n", renderProgressBar(scoreRatio, 25), scorePct)
	b.WriteString("\n")

	// Experiment statistics
	fmt.Fprintf(&b,"  Experiments: %d/%d", data.Experiments, data.MaxExperiments)
	fmt.Fprintf(&b,"  (Keep: %d, Discard: %d)\n", data.KeepCount, data.DiscardCount)

	// Per-criterion breakdown
	if len(data.PerCriterion) > 0 {
		b.WriteString("\n")
		b.WriteString(headerStyle.Render("Per-Criterion Breakdown"))
		b.WriteString("\n\n")

		// Calculate maximum name length for alignment
		maxNameLen := 0
		for _, cs := range data.PerCriterion {
			if len(cs.Name) > maxNameLen {
				maxNameLen = len(cs.Name)
			}
		}

		for _, cs := range data.PerCriterion {
			b.WriteString("  ")
			b.WriteString(renderCriterionLine(cs, maxNameLen))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// RenderCompact renders a single-line summary for use in status bars.
// Format: "Research: {target} {score}% ({experiments}/{max}) {keep}K/{discard}D"
func RenderCompact(data *DashboardData) string {
	if data == nil {
		return ""
	}

	scorePct := int(data.CurrentScore * 100)
	return fmt.Sprintf("Research: %s %d%% (%d/%d) %dK/%dD",
		data.Target,
		scorePct,
		data.Experiments,
		data.MaxExperiments,
		data.KeepCount,
		data.DiscardCount,
	)
}

// renderProgressBar generates a unicode block progress bar of the specified width.
// ratio is clamped to the range 0.0~1.0.
// Uses █ (filled) and ░ (empty) characters.
func renderProgressBar(ratio float64, width int) string {
	// Clamp range
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	return strings.Repeat("█", filled) + strings.Repeat("░", empty)
}

// renderCriterionLine renders a single criterion's name, bar, percentage, and MUST label.
// Names are left-padded based on maxNameLen for alignment.
func renderCriterionLine(cs CriterionStatus, maxNameLen int) string {
	// Name padding
	paddedName := fmt.Sprintf("%-*s", maxNameLen, cs.Name)

	// Percentage
	pct := int(cs.PassRate * 100)

	// Mini progress bar (width 15)
	bar := renderProgressBar(cs.PassRate, 15)

	// MUST label
	var label string
	if cs.Weight == "MUST" {
		label = " " + mustStyle.Render("MUST")
	}

	return fmt.Sprintf("%s %s %3d%%%s", paddedName, bar, pct, label)
}
