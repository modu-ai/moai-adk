package merge

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	analysis MergeAnalysis
	decision bool // true = proceed, false = cancel
	done     bool // true = user made a decision
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.decision = true
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
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DA7756")). // MoAI brand color
		Bold(true).
		MarginBottom(1)
	b.WriteString(titleStyle.Render("📊 Merge Analysis Results"))
	b.WriteString("\n\n")

	// Summary
	if m.analysis.Summary != "" {
		b.WriteString(fmt.Sprintf("📝 %s\n", m.analysis.Summary))
	}

	// Risk level
	if m.analysis.RiskLevel != "" {
		riskStyle := lipgloss.NewStyle()
		switch strings.ToLower(m.analysis.RiskLevel) {
		case "low":
			riskStyle = riskStyle.Foreground(lipgloss.Color("#10B981")) // Success green
		case "medium":
			riskStyle = riskStyle.Foreground(lipgloss.Color("#F59E0B")) // Warning yellow
		case "high":
			riskStyle = riskStyle.Foreground(lipgloss.Color("#EF4444")) // Error red
		}
		b.WriteString(riskStyle.Render(fmt.Sprintf("⚠️  Risk Level: %s", m.analysis.RiskLevel)))
		b.WriteString("\n\n")
	}

	// File table
	if len(m.analysis.Files) > 0 {
		tableStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4B5563")).
			Padding(1, 2)

		// Table header
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#DA7756"))
		var tableContent strings.Builder
		tableContent.WriteString(headerStyle.Render("File") + strings.Repeat(" ", 20))
		tableContent.WriteString(headerStyle.Render("Changes") + strings.Repeat(" ", 10))
		tableContent.WriteString(headerStyle.Render("Strategy") + strings.Repeat(" ", 8))
		tableContent.WriteString(headerStyle.Render("Risk"))
		tableContent.WriteString("\n")
		tableContent.WriteString(strings.Repeat("─", 80))
		tableContent.WriteString("\n")

		// Table rows
		for _, file := range m.analysis.Files {
			// Truncate file path if too long
			filePath := file.Path
			if len(filePath) > 24 {
				filePath = "..." + filePath[len(filePath)-21:]
			}

			// Truncate changes if too long
			changes := file.Changes
			if len(changes) > 18 {
				changes = changes[:15] + "..."
			}

			// Risk color
			riskStyle := lipgloss.NewStyle()
			switch strings.ToLower(file.RiskLevel) {
			case "low":
				riskStyle = riskStyle.Foreground(lipgloss.Color("#10B981"))
			case "medium":
				riskStyle = riskStyle.Foreground(lipgloss.Color("#F59E0B"))
			case "high":
				riskStyle = riskStyle.Foreground(lipgloss.Color("#EF4444"))
			}

			tableContent.WriteString(fmt.Sprintf("%-24s %-18s %-16s %s\n",
				filePath,
				changes,
				file.Strategy,
				riskStyle.Render(file.RiskLevel),
			))
		}

		b.WriteString(tableStyle.Render(tableContent.String()))
		b.WriteString("\n")
	}

	// Conflict warning
	if m.analysis.HasConflicts {
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)
		conflictCount := 0
		for _, file := range m.analysis.Files {
			if strings.ToLower(file.RiskLevel) == "high" {
				conflictCount++
			}
		}
		b.WriteString("\n")
		b.WriteString(warningStyle.Render(fmt.Sprintf("⚠️  Warning: %d file(s) with high risk conflicts detected", conflictCount)))
		b.WriteString("\n")
	}

	// Prompt
	promptStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")). // Secondary color
		MarginTop(1)
	b.WriteString("\n")
	b.WriteString(promptStyle.Render("Proceed with merge? (y/n) "))

	return b.String()
}

// ConfirmMerge displays an interactive confirmation UI and returns user's decision.
// Returns true if user confirms, false if user cancels.
func ConfirmMerge(analysis MergeAnalysis) (bool, error) {
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
