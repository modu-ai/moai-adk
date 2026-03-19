package dashboard

import (
	"fmt"
	"strings"
	"time"
)

// Renderer produces CLI-friendly output for dashboard data.
// Uses plain text formatting (lipgloss integration deferred).
type Renderer struct{}

// NewRenderer creates a new dashboard renderer.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// RenderSummary renders the main dashboard summary.
func (r *Renderer) RenderSummary(summary DashboardSummary) string {
	var sb strings.Builder

	sb.WriteString("╔══════════════════════════════════════╗\n")
	sb.WriteString("║       MoAI Dashboard Summary         ║\n")
	sb.WriteString("╚══════════════════════════════════════╝\n\n")

	sb.WriteString(fmt.Sprintf("  Active SPECs:    %d\n", summary.ActiveSPECs))
	sb.WriteString(fmt.Sprintf("  Completed SPECs: %d\n", summary.CompletedSPECs))
	sb.WriteString(fmt.Sprintf("  Total Tokens:    %s\n", formatTokens(summary.TotalTokensUsed)))
	sb.WriteString("\n")

	// Budget
	r.renderBudgetInline(&sb, summary.Budget)

	// Top agents
	if len(summary.TopAgents) > 0 {
		sb.WriteString("\n  Top Agents:\n")
		for i, a := range summary.TopAgents {
			if i >= 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("    %-20s %3d calls  %.0f%% success\n",
				a.AgentName, a.TotalCalls, a.SuccessRate))
		}
	}

	return sb.String()
}

// RenderSPECCosts renders cost analysis per SPEC.
func (r *Renderer) RenderSPECCosts(costs []SPECCost) string {
	var sb strings.Builder

	sb.WriteString("SPEC Cost Analysis\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	sb.WriteString(fmt.Sprintf("%-20s %12s %8s %6s\n", "SPEC ID", "Tokens", "Duration", "Calls"))
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	for _, c := range costs {
		sb.WriteString(fmt.Sprintf("%-20s %12s %8s %6d\n",
			c.SpecID,
			formatTokens(c.TotalTokens),
			formatDuration(c.Duration),
			c.AgentCalls))
	}

	return sb.String()
}

// RenderAgentStats renders agent performance statistics.
func (r *Renderer) RenderAgentStats(stats []AgentStats) string {
	var sb strings.Builder

	sb.WriteString("Agent Performance Statistics\n")
	sb.WriteString(strings.Repeat("─", 70) + "\n")
	sb.WriteString(fmt.Sprintf("%-22s %6s %6s %6s %8s %10s\n",
		"Agent", "Total", "Pass", "Fail", "Rate", "Avg Tokens"))
	sb.WriteString(strings.Repeat("─", 70) + "\n")

	for _, s := range stats {
		sb.WriteString(fmt.Sprintf("%-22s %6d %6d %6d %7.1f%% %10s\n",
			s.AgentName, s.TotalCalls, s.SuccessCount, s.FailureCount,
			s.SuccessRate, formatTokens(s.AvgTokens)))
	}

	return sb.String()
}

// RenderTrends renders quality trend report.
func (r *Renderer) RenderTrends(report *TrendReport) string {
	var sb strings.Builder

	sb.WriteString("Quality Trends\n")
	sb.WriteString(strings.Repeat("─", 40) + "\n")
	sb.WriteString(fmt.Sprintf("  Period:          %s\n", report.Period))
	sb.WriteString(fmt.Sprintf("  SPECs Created:   %d\n", report.SPECsCreated))
	sb.WriteString(fmt.Sprintf("  Avg Quality:     %.1f%%\n", report.AvgQuality))
	sb.WriteString(fmt.Sprintf("  Quality Trend:   %s\n", trendIndicator(report.QualityTrend)))

	return sb.String()
}

// RenderBudget renders budget status.
func (r *Renderer) RenderBudget(status BudgetStatus) string {
	var sb strings.Builder

	sb.WriteString("Budget Status\n")
	sb.WriteString(strings.Repeat("─", 40) + "\n")

	r.renderBudgetInline(&sb, status)

	return sb.String()
}

func (r *Renderer) renderBudgetInline(sb *strings.Builder, status BudgetStatus) {
	alertMarker := ""
	if status.IsAlert {
		alertMarker = " ⚠ ALERT"
	}

	sb.WriteString(fmt.Sprintf("  Budget:  %s / %s (%.1f%%)%s\n",
		formatTokens(status.UsedTokens),
		formatTokens(status.TotalBudget),
		status.UsagePercent,
		alertMarker))

	// Progress bar
	barWidth := 30
	filled := int(status.UsagePercent / 100.0 * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	sb.WriteString(fmt.Sprintf("  [%s]\n", bar))
}

func formatTokens(tokens int64) string {
	if tokens >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(tokens)/1_000_000)
	}
	if tokens >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(tokens)/1_000)
	}
	return fmt.Sprintf("%d", tokens)
}

func formatDuration(d time.Duration) string {
	if d >= time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	if d >= time.Minute {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func trendIndicator(trend string) string {
	switch trend {
	case "improving":
		return "↑ Improving"
	case "declining":
		return "↓ Declining"
	case "stable":
		return "→ Stable"
	default:
		return "? Unknown"
	}
}
