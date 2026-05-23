package merge

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/modu-ai/moai-adk/internal/tui"
)

const (
	// Input validation limits
	maxMergeFiles = 1000 // Maximum files in a single merge operation
	maxPathLength = 1024 // Maximum file path length (security)
)

// MergeAnalysis holds analysis results for multiple files to be merged.
type MergeAnalysis struct {
	Files        []FileAnalysis
	HasConflicts bool
	SafeToMerge  bool
	Summary      string
	RiskLevel    string
}

// FileAnalysis contains merge analysis for a single file.
type FileAnalysis struct {
	Path      string
	Changes   string
	Strategy  MergeStrategy
	RiskLevel string // "low", "medium", "high"
	Note      string
}

// confirmModel is the Bubble Tea model for merge confirmation UI.
type confirmModel struct {
	analysis        MergeAnalysis
	decision        bool   // true = proceed, false = cancel
	done            bool   // true = user made a decision
	cursor          int    // current cursor position
	selectedFiles   []bool // selection state for each file
	showSelection   bool   // true = legacy single-table selection UI
	showLowExpanded bool   // true = expand the Low-risk group in Cargo view
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			// If in selection mode, only proceed with selected files
			if m.showSelection {
				// Check if any files are selected
				hasSelection := false
				for _, s := range m.selectedFiles {
					if s {
						hasSelection = true
						break
					}
				}
				if !hasSelection {
					// No selection, proceed with all files
					m.decision = true
				} else {
					// Filter to only selected files
					m.analysis = m.filterSelectedFiles()
					m.decision = true
				}
			} else {
				m.decision = true
			}
			m.done = true
			return m, tea.Quit
		case "n", "N":
			m.decision = false
			m.done = true
			return m, tea.Quit
		case "ctrl+c":
			m.decision = false
			m.done = true
			return m, tea.Quit
		case "s", "S":
			// Toggle selection mode
			if len(m.analysis.Files) > 0 {
				m.showSelection = !m.showSelection
				// Initialize selectedFiles slice if needed
				if m.showSelection && len(m.selectedFiles) != len(m.analysis.Files) {
					m.selectedFiles = make([]bool, len(m.analysis.Files))
					// Select all by default
					for i := range m.selectedFiles {
						m.selectedFiles[i] = true
					}
				}
			}
		case "a", "A":
			// Select all files
			if m.showSelection {
				for i := range m.selectedFiles {
					m.selectedFiles[i] = true
				}
			}
		case "d", "D":
			// Deselect all files
			if m.showSelection {
				for i := range m.selectedFiles {
					m.selectedFiles[i] = false
				}
			}
		case " ":
			// Toggle current file selection
			if m.showSelection && len(m.selectedFiles) > 0 {
				m.selectedFiles[m.cursor] = !m.selectedFiles[m.cursor]
			}
		case "up", "k":
			if m.showSelection && m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.showSelection && m.cursor < len(m.analysis.Files)-1 {
				m.cursor++
			}
		case "v", "V":
			// Expand / collapse the Low-risk group in Cargo view.
			if !m.showSelection {
				m.showLowExpanded = !m.showLowExpanded
			}
		}
	}
	return m, nil
}

// filterSelectedFiles returns a new MergeAnalysis with only selected files.
func (m confirmModel) filterSelectedFiles() MergeAnalysis {
	var selectedFiles []FileAnalysis
	var highRisk, medRisk, lowRisk int

	for i, file := range m.analysis.Files {
		if i < len(m.selectedFiles) && m.selectedFiles[i] {
			selectedFiles = append(selectedFiles, file)
			switch file.RiskLevel {
			case "high":
				highRisk++
			case "medium":
				medRisk++
			case "low":
				lowRisk++
			}
		}
	}

	overallRisk := "low"
	hasConflicts := false
	if highRisk > 0 {
		overallRisk = "high"
		hasConflicts = true
	} else if medRisk > 0 {
		overallRisk = "medium"
	}

	summary := fmt.Sprintf("Found %d files to sync", len(selectedFiles))
	if highRisk > 0 {
		summary += fmt.Sprintf(" (%d high-risk files)", highRisk)
	}

	return MergeAnalysis{
		Files:        selectedFiles,
		HasConflicts: hasConflicts,
		SafeToMerge:  highRisk == 0,
		Summary:      summary,
		RiskLevel:    overallRisk,
	}
}

func (m confirmModel) View() string {
	if m.done {
		return ""
	}

	formatter := NewAnalysisFormatterWithSelection(m.analysis, m.cursor, m.selectedFiles, m.showSelection)
	formatter.showLowExpanded = m.showLowExpanded
	return formatter.Render()
}

// AnalysisFormatter handles formatting of merge analysis for display.
type AnalysisFormatter struct {
	analysis        MergeAnalysis
	styles          formatterStyles
	cursor          int
	selectedFiles   []bool
	showSelection   bool
	showLowExpanded bool
}

type formatterStyles struct {
	title          lipgloss.Style
	lowRisk        lipgloss.Style
	mediumRisk     lipgloss.Style
	highRisk       lipgloss.Style
	prompt         lipgloss.Style
	warning        lipgloss.Style
	headerStyle    lipgloss.Style
	tableHeaderRow lipgloss.Style
	tableBorder    lipgloss.Style
	tableRowEven   lipgloss.Style
	tableRowOdd    lipgloss.Style
	columnFile     lipgloss.Style
	columnChanges  lipgloss.Style
	columnStrategy lipgloss.Style
	columnRisk     lipgloss.Style
}

// NewAnalysisFormatter creates a new formatter for the given analysis.
func NewAnalysisFormatter(analysis MergeAnalysis) *AnalysisFormatter {
	return &AnalysisFormatter{
		analysis:      analysis,
		styles:        initFormatterStyles(),
		showSelection: false,
		selectedFiles: nil,
		cursor:        0,
	}
}

// NewAnalysisFormatterWithSelection creates a new formatter with file selection UI.
func NewAnalysisFormatterWithSelection(analysis MergeAnalysis, cursor int, selectedFiles []bool, showSelection bool) *AnalysisFormatter {
	return &AnalysisFormatter{
		analysis:      analysis,
		styles:        initFormatterStyles(),
		cursor:        cursor,
		selectedFiles: selectedFiles,
		showSelection: showSelection,
	}
}

func initFormatterStyles() formatterStyles {
	lt := tui.LightTheme()
	dt := tui.DarkTheme()
	accent := lipgloss.AdaptiveColor{Light: lt.Accent, Dark: dt.Accent}
	success := lipgloss.AdaptiveColor{Light: lt.Success, Dark: dt.Success}
	warning := lipgloss.AdaptiveColor{Light: lt.Warning, Dark: dt.Warning}
	danger := lipgloss.AdaptiveColor{Light: lt.Danger, Dark: dt.Danger}
	body := lipgloss.AdaptiveColor{Light: lt.Body, Dark: dt.Body}
	dim := lipgloss.AdaptiveColor{Light: lt.Dim, Dark: dt.Dim}
	rule := lipgloss.AdaptiveColor{Light: lt.Rule, Dark: dt.Rule}

	return formatterStyles{
		title:       lipgloss.NewStyle().Bold(true).Foreground(accent),
		lowRisk:     lipgloss.NewStyle().Foreground(success),
		mediumRisk:  lipgloss.NewStyle().Foreground(warning),
		highRisk:    lipgloss.NewStyle().Foreground(danger),
		prompt:      lipgloss.NewStyle().Foreground(dim),
		warning:     lipgloss.NewStyle().Foreground(warning).Bold(true),
		headerStyle: lipgloss.NewStyle().Bold(true).Foreground(accent),
		tableHeaderRow: lipgloss.NewStyle().
			Bold(true).
			Foreground(dim).
			Padding(0, 1),
		tableBorder:    lipgloss.NewStyle().Foreground(rule),
		tableRowEven:   lipgloss.NewStyle().Foreground(body),
		tableRowOdd:    lipgloss.NewStyle().Foreground(dim),
		columnFile:     lipgloss.NewStyle().Width(39).MaxWidth(39),
		columnChanges:  lipgloss.NewStyle().Width(16).MaxWidth(16),
		columnStrategy: lipgloss.NewStyle().Width(14).MaxWidth(14),
		columnRisk:     lipgloss.NewStyle().Width(8).MaxWidth(8),
	}
}

// FormatTitle returns the formatted title section.
func (f *AnalysisFormatter) FormatTitle() string {
	return f.styles.title.Render("Merge Analysis")
}

// FormatSummary returns the formatted summary section.
func (f *AnalysisFormatter) FormatSummary() string {
	if f.analysis.Summary == "" {
		return ""
	}
	return f.analysis.Summary
}

// FormatRiskLevel returns the formatted risk level with color.
func (f *AnalysisFormatter) FormatRiskLevel(level string) string {
	if level == "" {
		return ""
	}
	style := f.getRiskStyle(level)
	return style.Render(level)
}

// FormatOverallRisk returns the formatted overall risk level section.
func (f *AnalysisFormatter) FormatOverallRisk() string {
	if f.analysis.RiskLevel == "" {
		return ""
	}
	style := f.getRiskStyle(f.analysis.RiskLevel)
	return style.Render(fmt.Sprintf("Risk: %s", f.analysis.RiskLevel))
}

func (f *AnalysisFormatter) getRiskStyle(level string) lipgloss.Style {
	switch strings.ToLower(level) {
	case "high":
		return f.styles.highRisk
	case "medium":
		return f.styles.mediumRisk
	default:
		return f.styles.lowRisk
	}
}

// FormatFileTable returns the formatted file list table with modern TUI design.
func (f *AnalysisFormatter) FormatFileTable() string {
	if len(f.analysis.Files) == 0 {
		return ""
	}

	var tableContent strings.Builder

	// Column widths - adjust for selection mode
	const (
		colSelect   = 3
		colFile     = 39
		colChanges  = 16
		colStrategy = 14
		colRisk     = 8
		dividerLen  = colSelect + colFile + colChanges + colStrategy + colRisk + 6
	)

	// Table header with background
	var header string
	if f.showSelection {
		header = fmt.Sprintf(" %-2s %-38s %-15s %-13s %-7s ",
			"✓",
			"File Path",
			"Changes",
			"Strategy",
			"Risk",
		)
	} else {
		header = fmt.Sprintf(" %-41s %-15s %-13s %-7s ",
			"File Path",
			"Changes",
			"Strategy",
			"Risk",
		)
	}
	tableContent.WriteString(f.styles.tableHeaderRow.Render(header))
	tableContent.WriteString("\n")

	// Header separator
	tableContent.WriteString(f.styles.tableBorder.Render("╭" + strings.Repeat("─", dividerLen) + "╮"))
	tableContent.WriteString("\n")

	// Table rows with alternating styles
	for i, file := range f.analysis.Files {
		filePath := f.truncatePath(file.Path)
		changes := f.truncateChanges(file.Changes)
		riskColored := f.FormatRiskLevel(file.RiskLevel)

		// Determine row style
		rowStyle := f.styles.tableRowEven
		cursorStyle := rowStyle
		if i%2 == 1 {
			rowStyle = f.styles.tableRowOdd
			cursorStyle = rowStyle
		}

		// Highlight cursor row in selection mode
		if f.showSelection && i == f.cursor {
			cursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#111827", Dark: "#E5E7EB"}).
				Background(lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#4B5563"}).
				Bold(true)
		}

		// Format row based on mode
		var row string
		if f.showSelection {
			// Selection checkbox
			checkMark := " "
			if i < len(f.selectedFiles) && f.selectedFiles[i] {
				checkMark = "✓"
			}
			cursor := " "
			if i == f.cursor {
				cursor = "→"
			}

			row = fmt.Sprintf("│%s%s %-38s %-15s %-13s %-7s │",
				cursor, checkMark,
				filePath,
				changes,
				file.Strategy,
				riskColored,
			)
		} else {
			row = fmt.Sprintf("│ %-45s %-15s %-13s %-7s │",
				filePath,
				changes,
				file.Strategy,
				riskColored,
			)
		}
		tableContent.WriteString(cursorStyle.Render(row))
		tableContent.WriteString("\n")
	}

	// Table bottom border
	tableContent.WriteString(f.styles.tableBorder.Render("╰" + strings.Repeat("─", dividerLen) + "╯"))

	return tableContent.String()
}

// truncatePath shortens the path for display while keeping important parts.
// For skills, shows "skills/{skill-name}" instead of full ".claude/skills/..." path.
// For other files, truncates from the beginning if too long.
func (f *AnalysisFormatter) truncatePath(path string) string {
	// Special handling for skills: show "skills/{skill-name}/..." format
	if strings.Contains(path, ".claude/skills/") {
		// Extract skill name from .claude/skills/{skill-name}/...
		parts := strings.Split(path, string(filepath.Separator))
		for i, part := range parts {
			if part == "skills" && i+1 < len(parts) {
				// Show "skills/{skill-name}/remaining/path"
				skillName := parts[i+1]
				remaining := strings.Join(parts[i+2:], string(filepath.Separator))
				if remaining != "" {
					return "skills/" + skillName + "/" + remaining
				}
				return "skills/" + skillName
			}
		}
	}

	// Special handling for agents: show "agents/{agent-name}/..." format
	if strings.Contains(path, ".claude/agents/") {
		parts := strings.Split(path, string(filepath.Separator))
		for i, part := range parts {
			if part == "agents" && i+1 < len(parts) {
				agentName := parts[i+1]
				remaining := strings.Join(parts[i+2:], string(filepath.Separator))
				if remaining != "" {
					return "agents/" + agentName + "/" + remaining
				}
				return "agents/" + agentName
			}
		}
	}

	// Special handling for rules: show "rules/..." format
	if strings.Contains(path, ".claude/rules/") {
		parts := strings.Split(path, string(filepath.Separator))
		for i, part := range parts {
			if part == "rules" && i+1 < len(parts) {
				remaining := strings.Join(parts[i+1:], string(filepath.Separator))
				return "rules/" + remaining
			}
		}
	}

	// Special handling for commands: show "commands/..." format
	if strings.Contains(path, ".claude/commands/") {
		parts := strings.Split(path, string(filepath.Separator))
		for i, part := range parts {
			if part == "commands" && i+1 < len(parts) {
				remaining := strings.Join(parts[i+1:], string(filepath.Separator))
				return "commands/" + remaining
			}
		}
	}

	// For other files, truncate if too long
	const maxWidth = 38 // Maximum width for file path column (39 - 1 for padding)
	if len(path) > maxWidth {
		return "..." + path[len(path)-maxWidth+3:]
	}
	return path
}

func (f *AnalysisFormatter) truncateChanges(changes string) string {
	const maxWidth = 13 // Maximum width for changes column
	if len(changes) > maxWidth {
		return changes[:maxWidth-3] + "..."
	}
	return changes
}

// FormatConflictWarning returns the formatted conflict warning if applicable.
func (f *AnalysisFormatter) FormatConflictWarning() string {
	if !f.analysis.HasConflicts {
		return ""
	}

	conflictCount := 0
	for _, file := range f.analysis.Files {
		if strings.ToLower(file.RiskLevel) == "high" {
			conflictCount++
		}
	}

	return f.styles.warning.Render(
		fmt.Sprintf("! Warning: %d file(s) with high risk conflicts detected", conflictCount),
	)
}

// FormatPrompt returns the formatted user prompt.
func (f *AnalysisFormatter) FormatPrompt() string {
	if f.showSelection {
		return f.styles.prompt.Render(
			"↑↓ navigate  ·  space toggle  ·  a select-all  ·  d deselect-all  ·  s exit-selection  ·  y proceed  ·  n cancel",
		)
	}
	expandHint := "v expand"
	if f.showLowExpanded {
		expandHint = "v collapse"
	}
	return f.styles.prompt.Render(fmt.Sprintf("[Y/n]  ·  s select  ·  %s", expandHint))
}

// Render returns the complete formatted view. Selection mode (cursor-driven
// per-file picker) keeps the legacy single-table layout; normal mode uses the
// Cargo / uv style aligned text layout (no boxes, no emoji, sparse color).
func (f *AnalysisFormatter) Render() string {
	if f.showSelection {
		return f.renderSelectionView()
	}
	return f.renderCargoView()
}

// renderSelectionView preserves the legacy single-table layout used when the
// user enters per-file selection mode (cursor + space toggle).
func (f *AnalysisFormatter) renderSelectionView() string {
	var b strings.Builder

	b.WriteString(f.FormatTitle())
	b.WriteString("\n\n")
	if summary := f.FormatSummary(); summary != "" {
		b.WriteString(summary)
		b.WriteString("\n")
	}
	if riskLevel := f.FormatOverallRisk(); riskLevel != "" {
		b.WriteString(riskLevel)
		b.WriteString("\n\n")
	}

	selectedCount := 0
	for _, s := range f.selectedFiles {
		if s {
			selectedCount++
		}
	}
	if len(f.selectedFiles) > 0 {
		fmt.Fprintf(&b, "Selected: %d / %d files\n\n", selectedCount, len(f.selectedFiles))
	}

	if table := f.FormatFileTable(); table != "" {
		b.WriteString(table)
		b.WriteString("\n")
	}
	if warning := f.FormatConflictWarning(); warning != "" {
		b.WriteString("\n")
		b.WriteString(warning)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(f.FormatPrompt())
	return b.String()
}

// renderCargoView is the default normal-mode view. It follows the Cargo / uv
// CLI aesthetic: verb-prefixed key/value header, one aligned line per file
// (color from risk level), low-risk collapsed to path-prefix groups, and a
// single `Continue? [Y/n]` prompt at the bottom.
func (f *AnalysisFormatter) renderCargoView() string {
	var b strings.Builder

	// Header: verb-prefixed key/value lines (Cargo-style).
	f.writeVerbLine(&b, "Analyzing", fmt.Sprintf("%d files", len(f.analysis.Files)))
	if f.analysis.Summary != "" && !strings.HasPrefix(f.analysis.Summary, "Found ") {
		// Don't repeat the generic "Found N files to sync" — only show the
		// custom Summary (typically version delta) when present.
		f.writeVerbLine(&b, "Summary", f.analysis.Summary)
	}
	if f.analysis.RiskLevel != "" {
		f.writeVerbLine(&b, "Risk", f.styledRisk(f.analysis.RiskLevel, true))
	}
	b.WriteString("\n")

	// File list — high, medium, then low (collapsed or expanded).
	high, med, low := groupFilesByRisk(f.analysis.Files)
	for _, file := range high {
		b.WriteString(f.formatCargoRow(file, "high"))
		b.WriteString("\n")
	}
	for _, file := range med {
		b.WriteString(f.formatCargoRow(file, "medium"))
		b.WriteString("\n")
	}
	if len(low) > 0 {
		if f.showLowExpanded {
			for _, file := range low {
				b.WriteString(f.formatCargoRow(file, "low"))
				b.WriteString("\n")
			}
		} else {
			for _, group := range groupByPathPrefix(low, 4) {
				b.WriteString(f.formatCargoGroupRow(group, "low"))
				b.WriteString("\n")
			}
			b.WriteString(f.styles.tableRowOdd.Render(
				fmt.Sprintf("    … %d low-risk files (press v to expand)", len(low))))
			b.WriteString("\n")
		}
	}

	// Trailing warning + prompt.
	if warning := f.FormatConflictWarning(); warning != "" {
		b.WriteString("\n  ")
		b.WriteString(warning)
		b.WriteString("\n")
	}
	b.WriteString("\n  Continue? ")
	b.WriteString(f.FormatPrompt())
	return b.String()
}

// writeVerbLine appends a Cargo-style header line: `   <verb>  <value>`,
// where the verb column is right-aligned within 10 cells (Cargo convention).
func (f *AnalysisFormatter) writeVerbLine(b *strings.Builder, verb, value string) {
	verbStyle := f.styles.headerStyle
	fmt.Fprintf(b, "%s  %s\n", verbStyle.Render(fmt.Sprintf("%10s", verb)), value)
}

// styledRisk renders a risk-level string with the matching color. When bold
// is true the text is also rendered bold.
func (f *AnalysisFormatter) styledRisk(level string, bold bool) string {
	style := f.getRiskStyle(level)
	if bold {
		style = style.Bold(true)
	}
	return style.Render(level)
}

// formatCargoRow renders a single file row in Cargo style:
//
//	"  <icon> <path:34>  <strategy:14>  <risk:6>"
//
// The icon and risk word inherit the risk color; the path/strategy stay plain
// to keep the column dense and easy to scan.
func (f *AnalysisFormatter) formatCargoRow(file FileAnalysis, riskLevel string) string {
	icon := riskIcon(riskLevel)
	iconStyle := f.getRiskStyle(riskLevel)
	path := truncateRowField(file.Path, 34)
	strategy := truncateRowField(string(file.Strategy), 14)
	risk := truncateRowField(strings.ToLower(riskLevel), 6)
	return fmt.Sprintf("  %s  %-34s  %-14s  %s",
		iconStyle.Render(icon),
		path,
		strategy,
		f.styledRisk(risk, false),
	)
}

// formatCargoGroupRow renders a collapsed Low-risk group line:
//
//	"  ✓ scripts/ci-mirror/* (18)                       low"
func (f *AnalysisFormatter) formatCargoGroupRow(group pathGroup, riskLevel string) string {
	icon := riskIcon(riskLevel)
	iconStyle := f.getRiskStyle(riskLevel)
	label := fmt.Sprintf("%s* (%d)", group.prefix, group.count)
	label = truncateRowField(label, 34)
	return fmt.Sprintf("  %s  %-34s  %-14s  %s",
		iconStyle.Render(icon),
		label,
		"",
		f.styledRisk(strings.ToLower(riskLevel), false),
	)
}

// truncateRowField truncates s to at most max cells, replacing the last cell
// with a 1-cell ellipsis when needed.
func truncateRowField(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 1 {
		return "…"
	}
	return s[:max-1] + "…"
}

// groupFilesByRisk splits files into high / medium / low buckets while
// preserving the original order within each bucket.
func groupFilesByRisk(files []FileAnalysis) (high, med, low []FileAnalysis) {
	for _, f := range files {
		switch strings.ToLower(f.RiskLevel) {
		case "high":
			high = append(high, f)
		case "medium":
			med = append(med, f)
		default:
			low = append(low, f)
		}
	}
	return high, med, low
}

// pathGroup is one collapsed Low-risk group entry: a directory prefix and the
// number of files that share it.
type pathGroup struct {
	prefix string
	count  int
}

// groupByPathPrefix collapses a flat file list into the top N directory
// prefixes (first two path segments) sorted by file count descending.
func groupByPathPrefix(files []FileAnalysis, topN int) []pathGroup {
	counts := make(map[string]int, len(files))
	order := make([]string, 0, len(files))
	for _, file := range files {
		prefix := topTwoSegments(file.Path)
		if _, seen := counts[prefix]; !seen {
			order = append(order, prefix)
		}
		counts[prefix]++
	}
	groups := make([]pathGroup, 0, len(order))
	for _, p := range order {
		groups = append(groups, pathGroup{prefix: p, count: counts[p]})
	}
	sort.SliceStable(groups, func(i, j int) bool { return groups[i].count > groups[j].count })
	if topN > 0 && len(groups) > topN {
		groups = groups[:topN]
	}
	return groups
}

// topTwoSegments returns the first two path segments joined by "/", e.g.
// "scripts/ci-mirror/" for a 3-segment path. Single-segment paths return
// the path verbatim.
func topTwoSegments(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) <= 1 {
		return path
	}
	if len(parts) == 2 {
		return parts[0] + "/" + parts[1]
	}
	return parts[0] + "/" + parts[1] + "/"
}

// riskIcon returns the ASCII icon used in the file row prefix.
func riskIcon(riskLevel string) string {
	switch strings.ToLower(riskLevel) {
	case "high":
		return "✗"
	case "medium":
		return "!"
	default:
		return "✓"
	}
}

// validateAnalysis checks if a MergeAnalysis struct contains valid data.
// It prevents DoS attacks (file count limit), injection attacks (risk level whitelist),
// and path traversal attacks (path length limit).
func validateAnalysis(analysis MergeAnalysis) error {
	// DoS prevention: Limit file count to prevent memory exhaustion
	if len(analysis.Files) > maxMergeFiles {
		return fmt.Errorf("too many files: %d > %d", len(analysis.Files), maxMergeFiles)
	}

	// Validate risk level to prevent injection attacks
	validRiskLevels := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
	}
	if analysis.RiskLevel != "" && !validRiskLevels[strings.ToLower(analysis.RiskLevel)] {
		return fmt.Errorf("invalid risk level: %s", analysis.RiskLevel)
	}

	// Validate each file
	for i, file := range analysis.Files {
		// Path length validation to prevent buffer overflow
		if len(file.Path) > maxPathLength {
			return fmt.Errorf("file path too long at index %d: %d > %d", i, len(file.Path), maxPathLength)
		}

		// Validate file risk level
		if file.RiskLevel != "" && !validRiskLevels[strings.ToLower(file.RiskLevel)] {
			return fmt.Errorf("invalid file risk level at index %d: %s", i, file.RiskLevel)
		}
	}

	return nil
}

// sanitizePath removes potentially dangerous path components.
// It cleans the path and strips leading slashes and dots to prevent path traversal.
func sanitizePath(path string) string {
	// Normalize path separators and resolve . and ..
	cleaned := filepath.Clean(path)

	// Remove leading path separators to prevent absolute path attacks
	// Handle both forward and backward slashes for cross-platform safety
	cleaned = strings.TrimLeft(cleaned, `/\`)

	// Remove leading ./ and ../ sequences using OS-native separator
	sep := string(filepath.Separator)
	dotSep := "." + sep
	dotDotSep := ".." + sep
	for strings.HasPrefix(cleaned, dotSep) || strings.HasPrefix(cleaned, dotDotSep) {
		cleaned = strings.TrimPrefix(cleaned, dotSep)
		cleaned = strings.TrimPrefix(cleaned, dotDotSep)
	}

	return cleaned
}

// ConfirmMerge displays an interactive confirmation UI for template merge operations.
// It shows a table of files to be modified, their risk levels, and merge strategies.
// Users can approve (y) or reject (n) the merge. The UI is built with Bubble Tea.
//
// Input validation:
//   - File count limited to 1000 (DoS prevention)
//   - Risk levels must be "low", "medium", or "high"
//   - File paths limited to 1024 characters
//   - Paths are automatically sanitized (path traversal prevention)
//
// Returns true if user approves, false if rejected or error occurs.
func ConfirmMerge(analysis MergeAnalysis) (bool, error) {
	// Input validation
	if err := validateAnalysis(analysis); err != nil {
		return false, fmt.Errorf("invalid analysis: %w", err)
	}

	// Sanitize file paths
	for i := range analysis.Files {
		analysis.Files[i].Path = sanitizePath(analysis.Files[i].Path)
	}
	m := confirmModel{
		analysis: analysis,
		decision: false,
		done:     false,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("run confirmation UI: %w", err)
	}

	result := finalModel.(confirmModel)
	return result.decision, nil
}
