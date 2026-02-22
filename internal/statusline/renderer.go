package statusline

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Renderer formats StatusData into a single-line statusline string.
// Matches Python output format with emojis and context bar graph.
type Renderer struct {
	separator     string
	noColor       bool
	mutedStyle    lipgloss.Style
	segmentConfig map[string]bool
}

// NewRenderer creates a Renderer with the specified theme, color mode, and
// segment configuration. When segmentConfig is nil or empty, all segments
// are displayed (backward compatible).
func NewRenderer(themeName string, noColor bool, segmentConfig map[string]bool) *Renderer {
	r := &Renderer{
		separator:     " | ",
		noColor:       noColor,
		segmentConfig: segmentConfig,
	}

	if noColor {
		r.mutedStyle = lipgloss.NewStyle()
		return r
	}

	// All themes use the same muted style for consistency
	r.mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))

	return r
}

// Render formats the StatusData into a single-line string based on the mode.
// Format: 🤖 Model | 🔋/🪫 Context Graph | 💬 Style | 📁 Directory | 📊 Changes | 🔅 Claude Code Ver | 🗿 MoAI Ver | 🔀 Branch
func (r *Renderer) Render(data *StatusData, mode StatuslineMode) string {
	if data == nil {
		return "MoAI"
	}

	var sections []string

	switch mode {
	case ModeMinimal:
		sections = r.renderMinimal(data)
	case ModeVerbose:
		sections = r.renderVerbose(data)
	default: // ModeDefault (compact)
		sections = r.renderCompact(data)
	}

	// Filter empty sections
	filtered := make([]string, 0, len(sections))
	for _, s := range sections {
		if s != "" {
			filtered = append(filtered, s)
		}
	}

	if len(filtered) == 0 {
		return "MoAI"
	}

	return strings.Join(filtered, r.separator)
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

// renderCompact returns sections for compact mode with full emoji format.
// Format: 🤖 Model | 🔋/🪫 Context Graph | 💬 Style | 📁 Directory | 📊 Changes | 🔅 Claude Code Ver | 🗿 MoAI Ver | 🔀 Branch
// Each segment is filtered by isSegmentEnabled() based on the segment config.
func (r *Renderer) renderCompact(data *StatusData) []string {
	var sections []string

	// 1. Model with emoji
	if r.isSegmentEnabled(SegmentModel) && data.Metrics.Available && data.Metrics.Model != "" {
		sections = append(sections, fmt.Sprintf("🤖 %s", data.Metrics.Model))
	}

	// 2. Context window with battery emoji and bar graph
	if r.isSegmentEnabled(SegmentContext) && data.Memory.Available {
		if graph := r.renderContextGraph(data); graph != "" {
			sections = append(sections, graph)
		}
	}

	// 3. Output style with emoji
	if r.isSegmentEnabled(SegmentOutputStyle) && data.OutputStyle != "" {
		sections = append(sections, fmt.Sprintf("💬 %s", data.OutputStyle))
	}

	// 4. Directory with emoji
	if r.isSegmentEnabled(SegmentDirectory) && data.Directory != "" {
		sections = append(sections, fmt.Sprintf("📁 %s", data.Directory))
	}

	// 5. Git status with emoji
	if r.isSegmentEnabled(SegmentGitStatus) {
		if git := r.renderGitStatus(data); git != "" {
			sections = append(sections, fmt.Sprintf("📊 %s", git))
		}
	}

	// 6. Claude Code version with emoji (from JSON input)
	if r.isSegmentEnabled(SegmentClaudeVersion) && data.ClaudeCodeVersion != "" {
		sections = append(sections, fmt.Sprintf("🔅 v%s", data.ClaudeCodeVersion))
	}

	// 7. MoAI-ADK version with emoji (from config) + update notification
	if r.isSegmentEnabled(SegmentMoaiVersion) && data.Version.Available && data.Version.Current != "" {
		versionStr := fmt.Sprintf("🗿 v%s", data.Version.Current)
		if data.Version.UpdateAvailable && data.Version.Latest != "" {
			versionStr += fmt.Sprintf(" ⬆️ v%s", data.Version.Latest)
		}
		sections = append(sections, versionStr)
	}

	// 8. Branch with emoji
	if r.isSegmentEnabled(SegmentGitBranch) && data.Git.Available && data.Git.Branch != "" {
		sections = append(sections, fmt.Sprintf("🔀 %s", data.Git.Branch))
	}

	return sections
}

// renderMinimal returns sections for minimal mode: model + context graph only.
// Format: 🤖 Model | 🔋/🪫 Context Graph
func (r *Renderer) renderMinimal(data *StatusData) []string {
	var sections []string

	if data.Metrics.Available && data.Metrics.Model != "" {
		sections = append(sections, fmt.Sprintf("🤖 %s", data.Metrics.Model))
	}

	if data.Memory.Available {
		if graph := r.renderContextGraph(data); graph != "" {
			sections = append(sections, graph)
		}
	}

	// Add git status if it fits
	if git := r.renderGitStatus(data); git != "" {
		statusLabel := fmt.Sprintf("📊 %s", git)
		// Only add if total length would be under 40 chars
		currentLen := len(strings.Join(sections, r.separator))
		if currentLen+len(statusLabel)+len(r.separator) <= 40 {
			sections = append(sections, statusLabel)
		}
	}

	return sections
}

// renderVerbose returns sections for verbose mode: same as compact.
// Python uses the same format for both compact and extended.
func (r *Renderer) renderVerbose(data *StatusData) []string {
	return r.renderCompact(data)
}

// renderContextGraph renders the context window usage as a bar graph.
// Format: 🔋 ████░░░░░░░░ 41% or 🪫 ██████████░ 85%
// Battery icon: 🔋 (<=70% used), 🪫 (>70% used)
func (r *Renderer) renderContextGraph(data *StatusData) string {
	if !data.Memory.Available {
		return ""
	}

	used := data.Memory.TokensUsed
	total := data.Memory.TokenBudget

	if total <= 0 {
		return ""
	}

	pct := usagePercent(used, total)

	// Determine battery icon based on usage
	// 🔋 (70% or less used, 30%+ remaining) | 🪫 (over 70% used, less than 30% remaining)
	icon := "🔋"
	if pct > 70 {
		icon = "🪫"
	}

	// Build bar graph with 12 character width
	bar := r.buildBar(pct, 12)

	return fmt.Sprintf("%s  %s %d%%", icon, bar, pct)
}

// buildBar constructs a horizontal bar graph using Unicode block characters.
// Width is total bar width in characters.
// Uses full block (█) for used portion and light block (░) for remaining.
func (r *Renderer) buildBar(pct int, width int) string {
	if width <= 0 {
		return ""
	}

	// Calculate filled blocks based on percentage
	filled := min((pct*width)/100, width)

	empty := width - filled

	// Build bar using Unicode block characters
	filledChar := "█" // Full block for used
	emptyChar := "░"  // Light block for remaining

	return strings.Repeat(filledChar, filled) + strings.Repeat(emptyChar, empty)
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
