package statusline

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Renderer formats StatusData into a single-line statusline string.
// It supports multiple display modes and color themes, with NO_COLOR fallback.
type Renderer struct {
	separator   string
	noColor     bool
	useNerd     bool
	okStyle     lipgloss.Style
	warnStyle   lipgloss.Style
	errStyle    lipgloss.Style
	mutedStyle  lipgloss.Style
	branchStyle lipgloss.Style
}

// NewRenderer creates a Renderer with the specified theme and color mode.
// Supported theme names: "default", "minimal", "nerd".
// When noColor is true, all styling is disabled for plain text output.
func NewRenderer(themeName string, noColor bool) *Renderer {
	r := &Renderer{
		separator: " | ",
		noColor:   noColor,
	}

	if noColor {
		r.okStyle = lipgloss.NewStyle()
		r.warnStyle = lipgloss.NewStyle()
		r.errStyle = lipgloss.NewStyle()
		r.mutedStyle = lipgloss.NewStyle()
		r.branchStyle = lipgloss.NewStyle()
		return r
	}

	switch themeName {
	case "nerd":
		r.useNerd = true
		r.okStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
		r.warnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
		r.errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
		r.mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
		r.branchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
	case "minimal":
		r.okStyle = lipgloss.NewStyle()
		r.warnStyle = lipgloss.NewStyle()
		r.errStyle = lipgloss.NewStyle()
		r.mutedStyle = lipgloss.NewStyle()
		r.branchStyle = lipgloss.NewStyle()
	default: // "default"
		r.okStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
		r.warnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
		r.errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
		r.mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
		r.branchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
	}

	return r
}

// Render formats the StatusData into a single-line string based on the mode.
// Empty sections are automatically filtered out.
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
	default: // ModeDefault
		sections = r.renderDefault(data)
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

// renderMinimal returns sections for minimal mode: model + context percentage.
func (r *Renderer) renderMinimal(data *StatusData) []string {
	var sections []string

	if data.Metrics.Available && data.Metrics.Model != "" {
		sections = append(sections, data.Metrics.Model)
	}

	if data.Memory.Available {
		sections = append(sections, r.renderContextBrief(data))
	}

	return sections
}

// renderDefault returns sections for default mode: git + context + cost.
func (r *Renderer) renderDefault(data *StatusData) []string {
	var sections []string

	if git := r.renderGitBrief(data); git != "" {
		sections = append(sections, git)
	}

	if data.Memory.Available {
		sections = append(sections, r.renderContextBrief(data))
	}

	if data.Metrics.Available && data.Metrics.CostUSD > 0 {
		sections = append(sections, formatCost(data.Metrics.CostUSD))
	}

	return sections
}

// renderVerbose returns sections for verbose mode: all data with full detail.
func (r *Renderer) renderVerbose(data *StatusData) []string {
	var sections []string

	if git := r.renderGitVerbose(data); git != "" {
		sections = append(sections, git)
	}

	if data.Memory.Available {
		sections = append(sections, r.renderContextVerbose(data))
	}

	if data.Metrics.Available {
		sections = append(sections, formatCost(data.Metrics.CostUSD))
	}

	if ver := r.renderVersion(data); ver != "" {
		sections = append(sections, ver)
	}

	return sections
}

// renderGitBrief renders git info for default mode: "main +3 ~2"
func (r *Renderer) renderGitBrief(data *StatusData) string {
	if !data.Git.Available || data.Git.Branch == "" {
		return ""
	}

	parts := []string{data.Git.Branch}

	if data.Git.Staged > 0 {
		parts = append(parts, fmt.Sprintf("+%d", data.Git.Staged))
	}
	if data.Git.Modified > 0 {
		parts = append(parts, fmt.Sprintf("~%d", data.Git.Modified))
	}

	return strings.Join(parts, " ")
}

// renderGitVerbose renders git info for verbose mode: "main +3 ~2 ^1 v0"
func (r *Renderer) renderGitVerbose(data *StatusData) string {
	if !data.Git.Available || data.Git.Branch == "" {
		return ""
	}

	parts := []string{data.Git.Branch}

	if data.Git.Staged > 0 {
		parts = append(parts, fmt.Sprintf("+%d", data.Git.Staged))
	}
	if data.Git.Modified > 0 {
		parts = append(parts, fmt.Sprintf("~%d", data.Git.Modified))
	}

	parts = append(parts, fmt.Sprintf("^%d", data.Git.Ahead))
	parts = append(parts, fmt.Sprintf("v%d", data.Git.Behind))

	return strings.Join(parts, " ")
}

// renderContextBrief renders context usage for default/minimal: "Ctx: 25%"
func (r *Renderer) renderContextBrief(data *StatusData) string {
	if !data.Memory.Available {
		return ""
	}

	pct := usagePercent(data.Memory.TokensUsed, data.Memory.TokenBudget)
	text := fmt.Sprintf("Ctx: %d%%", pct)

	return r.applyContextStyle(text, data.Memory.TokensUsed, data.Memory.TokenBudget)
}

// renderContextVerbose renders context for verbose: "Ctx: 50K/200K (25%)"
func (r *Renderer) renderContextVerbose(data *StatusData) string {
	if !data.Memory.Available {
		return ""
	}

	pct := usagePercent(data.Memory.TokensUsed, data.Memory.TokenBudget)
	text := fmt.Sprintf("Ctx: %s/%s (%d%%)",
		formatTokens(data.Memory.TokensUsed),
		formatTokens(data.Memory.TokenBudget),
		pct,
	)

	return r.applyContextStyle(text, data.Memory.TokensUsed, data.Memory.TokenBudget)
}

// renderVersion renders version info for verbose mode.
func (r *Renderer) renderVersion(data *StatusData) string {
	if !data.Version.Available || data.Version.Current == "" {
		return ""
	}

	text := "v" + data.Version.Current
	if data.Version.UpdateAvailable {
		text += " (update!)"
	}

	return text
}

// applyContextStyle applies color based on context usage level.
// In NoColor mode, returns the text unmodified.
func (r *Renderer) applyContextStyle(text string, used, total int) string {
	if r.noColor {
		return text
	}

	level := contextUsageLevel(used, total)

	switch level {
	case levelError:
		return r.errStyle.Render(text)
	case levelWarn:
		return r.warnStyle.Render(text)
	default:
		return r.okStyle.Render(text)
	}
}
