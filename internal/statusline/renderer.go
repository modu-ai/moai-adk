package statusline

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Renderer formats StatusData into a multiline statusline string.
// Supports v3 layouts: default(3L), full(5L).
type Renderer struct {
	separator     string
	noColor       bool
	mutedStyle    lipgloss.Style
	segmentConfig map[string]bool
	theme         Theme
}

// NewRenderer creates a Renderer with the specified theme, color mode, and
// segment configuration. When segmentConfig is nil or empty, all segments
// are displayed (backward compatible).
func NewRenderer(themeName string, noColor bool, segmentConfig map[string]bool) *Renderer {
	theme := NewTheme(themeName)

	r := &Renderer{
		// v3 separator: U+2502 box drawing vertical line
		separator:     " │ ",
		noColor:       noColor,
		segmentConfig: segmentConfig,
		theme:         theme,
	}

	if noColor {
		r.mutedStyle = lipgloss.NewStyle()
		return r
	}

	// Set muted style from theme color (REQ-SLE-017)
	r.mutedStyle = lipgloss.NewStyle().Foreground(theme.Muted())

	return r
}

// Render formats the StatusData into a statusline string based on the mode.
//
// v3 mode mapping:
//   - ModeDefault, ModeCompact, ModeMinimal → 3-line default layout
//   - ModeFull, ModeVerbose                 → 5-line full layout
//   - unknown                               → 3-line default layout (fallback)
//
// @MX:ANCHOR: [AUTO] Single entry point for all mode rendering - called from Build() in builder.go
// @MX:REASON: Public API boundary; contains mode routing logic
func (r *Renderer) Render(data *StatusData, mode StatuslineMode) string {
	if data == nil {
		return "MoAI"
	}

	// Normalize deprecated mode names to v3 names
	normalizedMode := NormalizeMode(mode)

	var result string
	switch normalizedMode {
	case ModeFull:
		result = r.renderFullV3(data)
	default: // ModeDefault or unknown mode
		result = r.renderDefaultV3(data)
	}

	if result == "" {
		return "MoAI"
	}
	return result
}

// filterEmpty removes empty strings from a slice.
func filterEmpty(sections []string) []string {
	filtered := make([]string, 0, len(sections))
	for _, s := range sections {
		if s != "" {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// isSegmentEnabled checks whether a segment should be rendered based on config.
// Returns true (enabled) when segmentConfig is nil/empty (backward compatible),
// or when the key is not present in the config (unknown segments default to enabled).
func (r *Renderer) isSegmentEnabled(key string) bool {
	if len(r.segmentConfig) == 0 {
		return true
	}
	enabled, exists := r.segmentConfig[key]
	if !exists {
		return true
	}
	return enabled
}

// joinSegments filters a segment slice and joins them with the separator.
// Returns empty string if all segments are empty.
func (r *Renderer) joinSegments(segments []string) string {
	filtered := filterEmpty(segments)
	if len(filtered) == 0 {
		return ""
	}
	return strings.Join(filtered, r.separator)
}

// ─────────────────────────────────────────────────────────────────────────────
// v3 layout renderers
// ─────────────────────────────────────────────────────────────────────────────

// renderDefaultV3 renders the default mode 3-line layout.
//
// L1: 🤖 Model │ 🔅 v2.1.50 │ 🗿 v2.8.0 │ ⏳ 2h 34m │ 💬 MoAI
// L2: CW: 🪫 ██████████ 88% │ 5H: 🔋 ██████████ 45% │ 7D: 🪫 ██████████ 82%
// L3: 📁 moai-adk-go │ 🔀 feat/auth ↑2↓1 │ 📊 +3 M2 ?1
func (r *Renderer) renderDefaultV3(data *StatusData) string {
	var lines []string

	// L1: model, Claude version, MoAI version, session time, output style
	l1 := r.renderInfoLine(data, false)
	if l1 != "" {
		lines = append(lines, l1)
	}

	// L2: CW/5H/7D bars inline (10 blocks) - always show all 3 bars
	l2 := r.renderBarsInline(data, 10)
	if l2 != "" {
		lines = append(lines, l2)
	}

	// L3: directory, branch, git status
	l3 := r.renderDirGitLine(data)
	if l3 != "" {
		lines = append(lines, l3)
	}

	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

// renderFullV3 renders the full mode 5-line layout.
//
// L1: 🤖 Model │ 🔅 v2.1.50 │ 🗿 v2.8.0 │ ⏳ 2h 34m │ 💬 MoAI
// L2: CW: 🪫 ████████████████████████████████████░░░░ 88%
// L3: 5H: 🔋 ██████████████████░░░░░░░░░░░░░░░░░░░░░░ 45%
// L4: 7D: 🪫 ████████████████████████████████░░░░░░░░ 82%
// L5: 📁 moai-adk-go │ 🔀 feat/auth ↑2↓1 │ 📊 +3 M2 ?1
func (r *Renderer) renderFullV3(data *StatusData) string {
	var lines []string

	// L1: model, Claude version, MoAI version, session time, output style (no prefix)
	l1 := r.renderInfoLine(data, false)
	if l1 != "" {
		lines = append(lines, l1)
	}

	// L2: CW bar (40 blocks, standalone line)
	cwPct := r.contextPercent(data)
	if cwPct >= 0 {
		lines = append(lines, renderUsageBar("CW:", cwPct, 40, r.noColor))
	}

	// L3: 5H bar (40 blocks, standalone line) with reset time - defaults to 0% when no data
	// Prefer RateLimits (from Claude Code v2.1.80+ statusline JSON) over Usage (MoAI API call)
	pct5H := 0
	var reset5H string
	if data.RateLimits != nil && data.RateLimits.FiveHour != nil {
		pct5H = int(data.RateLimits.FiveHour.UsedPercentage)
		reset5H = formatResetTimeRelative(data.RateLimits.FiveHour.ResetsAt)
	} else if data.Usage != nil && data.Usage.Usage5H != nil {
		pct5H = int(data.Usage.Usage5H.Percentage)
		reset5H = formatResetTimeRelative(data.Usage.Usage5H.ResetsAt)
	}
	lines = append(lines, renderUsageBarWithReset("5H:", pct5H, 40, r.noColor, reset5H))

	// L4: 7D bar (40 blocks, standalone line) with reset date - defaults to 0% when no data
	pct7D := 0
	var reset7D string
	if data.RateLimits != nil && data.RateLimits.SevenDay != nil {
		pct7D = int(data.RateLimits.SevenDay.UsedPercentage)
		reset7D = formatResetTimeAbsolute(data.RateLimits.SevenDay.ResetsAt)
	} else if data.Usage != nil && data.Usage.Usage7D != nil {
		pct7D = int(data.Usage.Usage7D.Percentage)
		reset7D = formatResetTimeAbsolute(data.Usage.Usage7D.ResetsAt)
	}
	lines = append(lines, renderUsageBarWithReset("7D:", pct7D, 40, r.noColor, reset7D))

	// L5: directory, branch, git status
	l5 := r.renderDirGitLine(data)
	if l5 != "" {
		lines = append(lines, l5)
	}

	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

// ─────────────────────────────────────────────────────────────────────────────
// Common line renderers
// ─────────────────────────────────────────────────────────────────────────────

// renderInfoLine renders the L1 info line (shared by default/full).
// withPrefix=true: full mode format ("Claude v...", "MoAI v...")
// withPrefix=false: default mode format ("v...")
func (r *Renderer) renderInfoLine(data *StatusData, withPrefix bool) string {
	var segs []string

	// Model
	if r.isSegmentEnabled(SegmentModel) && data.Metrics.Available && data.Metrics.Model != "" {
		segs = append(segs, fmt.Sprintf("🤖 %s", data.Metrics.Model))
	}

	// Claude version
	if r.isSegmentEnabled(SegmentClaudeVersion) && data.ClaudeCodeVersion != "" {
		if withPrefix {
			segs = append(segs, fmt.Sprintf("🔅 Claude v%s", data.ClaudeCodeVersion))
		} else {
			segs = append(segs, fmt.Sprintf("🔅 v%s", data.ClaudeCodeVersion))
		}
	}

	// MoAI version
	if r.isSegmentEnabled(SegmentMoaiVersion) && data.Version.Available && data.Version.Current != "" {
		var versionStr string
		if withPrefix {
			versionStr = fmt.Sprintf("🗿 MoAI v%s", data.Version.Current)
		} else {
			versionStr = fmt.Sprintf("🗿 v%s", data.Version.Current)
		}
		if data.Version.UpdateAvailable && data.Version.Latest != "" {
			versionStr += fmt.Sprintf(" ⬆️ v%s", data.Version.Latest)
		}
		segs = append(segs, versionStr)
	}

	// Session time
	if r.isSegmentEnabled(SegmentSessionTime) && data.Metrics.Available {
		if st := renderSessionTime(data.Metrics.SessionDurationMS); st != "" {
			segs = append(segs, st)
		}
	}

	// Output style (integrated into L1)
	if r.isSegmentEnabled(SegmentOutputStyle) && data.OutputStyle != "" {
		segs = append(segs, fmt.Sprintf("💬 %s", data.OutputStyle))
	}

	return r.joinSegments(segs)
}

// renderBarsInline renders CW/5H/7D bars inline on a single line (default mode L2).
// width: number of blocks per bar
func (r *Renderer) renderBarsInline(data *StatusData, width int) string {
	var segs []string

	// CW bar
	if r.isSegmentEnabled(SegmentContext) && data.Memory.Available && data.Memory.TokenBudget > 0 {
		pct := usagePercent(data.Memory.TokensUsed, data.Memory.TokenBudget)
		segs = append(segs, renderUsageBar("CW:", pct, width, r.noColor))
	}

	// 5H bar - always shown, defaults to 0% when no data
	if r.isSegmentEnabled(SegmentUsage5H) {
		pct5H := 0
		if data.Usage != nil && data.Usage.Usage5H != nil {
			pct5H = int(data.Usage.Usage5H.Percentage)
		}
		segs = append(segs, renderUsageBar("5H:", pct5H, width, r.noColor))
	}

	// 7D bar - always shown, defaults to 0% when no data
	if r.isSegmentEnabled(SegmentUsage7D) {
		pct7D := 0
		if data.Usage != nil && data.Usage.Usage7D != nil {
			pct7D = int(data.Usage.Usage7D.Percentage)
		}
		segs = append(segs, renderUsageBar("7D:", pct7D, width, r.noColor))
	}

	return r.joinSegments(segs)
}

// renderDirGitLine renders the directory + branch + git status line (default L3, full L5).
// Format: 📁 moai-adk-go │ 🔀 feat/auth ↑2↓1 │ 📊 +3 M2 ?1
func (r *Renderer) renderDirGitLine(data *StatusData) string {
	var segs []string

	// Directory
	if r.isSegmentEnabled(SegmentDirectory) && data.Directory != "" {
		segs = append(segs, fmt.Sprintf("📁 %s", data.Directory))
	}

	// Branch + ahead/behind
	if r.isSegmentEnabled(SegmentGitBranch) {
		if branch := renderGitBranch(data); branch != "" {
			segs = append(segs, branch)
		}
	}

	// Git status
	if r.isSegmentEnabled(SegmentGitStatus) {
		if git := r.renderGitStatus(data); git != "" {
			segs = append(segs, fmt.Sprintf("📊 %s", git))
		}
	}

	return r.joinSegments(segs)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helper functions
// ─────────────────────────────────────────────────────────────────────────────

// renderUsageBar renders label + battery icon + gradient bar + percentage.
// Format: {label} {BatteryIcon(pct)} {BuildGradientBar(pct, width, noColor)} {pct}%
// Example: CW: 🪫 ████████████████████████████████████░░░░ 88%
func renderUsageBar(label string, pct int, width int, noColor bool) string {
	icon := BatteryIcon(pct)
	bar := BuildGradientBar(pct, width, noColor)
	return fmt.Sprintf("%s %s %s %d%%", label, icon, bar, pct)
}

// renderUsageBarWithReset renders a usage bar with optional reset time suffix.
// Format: {label} {icon} {bar} {pct}% (Resets {resetStr})
func renderUsageBarWithReset(label string, pct int, width int, noColor bool, resetStr string) string {
	base := renderUsageBar(label, pct, width, noColor)
	if resetStr == "" {
		return base
	}
	return fmt.Sprintf("%s (Resets %s)", base, resetStr)
}

// formatResetTimeRelative formats an ISO 8601 timestamp as relative time "in Xh Ym".
// Returns empty string on parse failure or if reset is in the past.
func formatResetTimeRelative(isoTimestamp string) string {
	if isoTimestamp == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, isoTimestamp)
	if err != nil {
		// Try without timezone
		t, err = time.Parse("2006-01-02T15:04:05", isoTimestamp)
		if err != nil {
			return ""
		}
	}
	remaining := time.Until(t)
	if remaining <= 0 {
		return ""
	}
	hours := int(remaining.Hours())
	minutes := int(remaining.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("in %dh%dm", hours, minutes)
	}
	return fmt.Sprintf("in %dm", minutes)
}

// formatResetTimeAbsolute formats an ISO 8601 timestamp as "Jan 21 at 2pm".
// Returns empty string on parse failure.
func formatResetTimeAbsolute(isoTimestamp string) string {
	if isoTimestamp == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, isoTimestamp)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05", isoTimestamp)
		if err != nil {
			return ""
		}
	}
	// Convert to local time for display
	t = t.Local()
	hour := t.Hour()
	ampm := "am"
	if hour >= 12 {
		ampm = "pm"
	}
	hour12 := hour % 12
	if hour12 == 0 {
		hour12 = 12
	}
	return fmt.Sprintf("%s at %d%s", t.Format("Jan 2"), hour12, ampm)
}

// contextPercent returns the context window usage percentage (0~100).
// Returns -1 if unavailable or total budget is 0.
func (r *Renderer) contextPercent(data *StatusData) int {
	if !data.Memory.Available || data.Memory.TokenBudget <= 0 {
		return -1
	}
	return usagePercent(data.Memory.TokensUsed, data.Memory.TokenBudget)
}

// renderGitBranch renders the git branch string with ahead/behind info.
// REQ-V3-GIT-001: Ahead > 0 → "🔀 branch ↑N"
// REQ-V3-GIT-002: Behind > 0 → "🔀 branch ↓N"
// REQ-V3-GIT-003: Both → "🔀 branch ↑N↓M"
// REQ-V3-GIT-004: Neither → "🔀 branch"
func renderGitBranch(data *StatusData) string {
	if !data.Git.Available || data.Git.Branch == "" {
		return ""
	}

	branch := data.Git.Branch
	var suffix string

	if data.Git.Ahead > 0 && data.Git.Behind > 0 {
		suffix = fmt.Sprintf(" ↑%d↓%d", data.Git.Ahead, data.Git.Behind)
	} else if data.Git.Ahead > 0 {
		suffix = fmt.Sprintf(" ↑%d", data.Git.Ahead)
	} else if data.Git.Behind > 0 {
		suffix = fmt.Sprintf(" ↓%d", data.Git.Behind)
	}

	return fmt.Sprintf("🔀 %s%s", branch, suffix)
}

// renderSessionTime converts milliseconds to a session time string in "⏳ Xh Ym" format.
// REQ-V3-TIME-002: >= 60min: "⏳ Xh Ym", < 60min: "⏳ Xm", >= 24h: "⏳ Xd Yh"
// REQ-V3-TIME-004: returns empty string when ms is 0
func renderSessionTime(ms int) string {
	if ms <= 0 {
		return ""
	}

	totalMinutes := ms / 60000
	totalHours := totalMinutes / 60

	// >= 24 hours: "⏳ Xd Yh"
	if totalHours >= 24 {
		days := totalHours / 24
		hours := totalHours % 24
		return fmt.Sprintf("⏳ %dd %dh", days, hours)
	}

	// >= 1 hour: "⏳ Xh Ym"
	if totalHours >= 1 {
		minutes := totalMinutes % 60
		return fmt.Sprintf("⏳ %dh %dm", totalHours, minutes)
	}

	// < 1 hour: "⏳ Xm"
	return fmt.Sprintf("⏳ %dm", totalMinutes)
}

// renderGitStatus renders git status in Python format.
// Format: +{staged} M{modified} ?{untracked}
// Example: "+0 M1066 ?2" (0 staged, 1066 modified, 2 untracked)
func (r *Renderer) renderGitStatus(data *StatusData) string {
	if !data.Git.Available {
		return ""
	}

	// Only show git status if there are changes
	if data.Git.Staged == 0 && data.Git.Modified == 0 && data.Git.Untracked == 0 {
		return ""
	}

	var parts []string

	// Staged files
	parts = append(parts, fmt.Sprintf("+%d", data.Git.Staged))

	// Modified files (uses M instead of ~)
	parts = append(parts, fmt.Sprintf("M%d", data.Git.Modified))

	// Untracked files
	parts = append(parts, fmt.Sprintf("?%d", data.Git.Untracked))

	return strings.Join(parts, " ")
}
