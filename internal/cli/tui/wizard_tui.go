// Package tui provides Bubble Tea TUI components for MoAI-ADK.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents different screens in the wizard.
type Screen int

const (
	// ScreenWelcome is the welcome screen.
	ScreenWelcome Screen = iota
	// ScreenProjectName is the project name input screen.
	ScreenProjectName
	// ScreenSuccess is the success screen.
	ScreenSuccess
)

// String returns the screen name.
func (s Screen) String() string {
	switch s {
	case ScreenWelcome:
		return "Welcome"
	case ScreenProjectName:
		return "ProjectName"
	case ScreenSuccess:
		return "Success"
	default:
		return "Unknown"
	}
}

// WizardModel is the main model for the wizard TUI.
type WizardModel struct {
	// Current screen
	currentScreen Screen

	// User input
	projectName string

	// Text input for project name
	textInput textinput.Model

	// Error message
	err error

	// Window dimensions
	width  int
	height int

	// Whether user is quitting
	quitting bool

	// Final result
	result *WizardResult
}

// NewWizardModel creates a new wizard model.
func NewWizardModel() WizardModel {
	ti := textinput.New()
	ti.Placeholder = "my-awesome-project"
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 40

	return WizardModel{
		currentScreen: ScreenWelcome,
		textInput:     ti,
		width:         80,
		height:        24,
	}
}

// Init implements tea.Model.
func (m WizardModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			return m.handleEnter()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Update text input based on current screen
	switch m.currentScreen {
	case ScreenProjectName:
		m.textInput, cmd = m.textInput.Update(msg)
		// Capture project name as user types
		m.projectName = m.textInput.Value()
	}

	return m, cmd
}

// handleEnter handles the Enter key press based on current screen.
func (m WizardModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.currentScreen {
	case ScreenWelcome:
		// Move to project name screen
		m.currentScreen = ScreenProjectName
		return m, nil

	case ScreenProjectName:
		// Validate project name
		if err := validateProjectName(m.projectName); err != nil {
			m.err = err
			return m, nil
		}

		// Create result and move to success screen
		m.result = &WizardResult{
			ProjectName: m.projectName,
		}
		m.currentScreen = ScreenSuccess
		m.quitting = true
		return m, tea.Quit

	case ScreenSuccess:
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// validateProjectName validates the project name.
func validateProjectName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("project name is required")
	}
	if len(name) < 2 {
		return fmt.Errorf("project name must be at least 2 characters")
	}
	if len(name) > 50 {
		return fmt.Errorf("project name must be at most 50 characters")
	}
	// Check for invalid characters
	for _, r := range name {
		if (r < 'a' || r > 'z') &&
			(r < 'A' || r > 'Z') &&
			(r < '0' || r > '9') &&
			(r != '-' && r != '_') {
			return fmt.Errorf("project name can only contain letters, numbers, hyphens, and underscores")
		}
	}
	return nil
}

// View implements tea.Model.
func (m WizardModel) View() string {
	if m.quitting {
		return ""
	}

	switch m.currentScreen {
	case ScreenWelcome:
		return m.renderWelcome()
	case ScreenProjectName:
		return m.renderProjectName()
	case ScreenSuccess:
		return m.renderSuccess()
	default:
		return "Unknown screen"
	}
}

// renderWelcome renders the welcome screen with a box border.
func (m WizardModel) renderWelcome() string {
	innerTitle := "MoAI-ADK"
	innerSubtitle := "Project Initialization Wizard"

	instructions := []string{
		"",
		"This wizard will guide you through creating a new",
		"MoAI-ADK project with AI-powered development tools.",
		"",
		"Press Enter to continue",
		"Press Q or Esc to quit",
	}

	// Calculate dimensions
	boxWidth := min(60, m.width-4)

	// Build inner content
	var innerContent strings.Builder
	innerContent.WriteString(styles.title.Render(innerTitle))
	innerContent.WriteString("\n")
	innerContent.WriteString(styles.subtitle.Render(innerSubtitle))
	innerContent.WriteString("\n")
	for _, line := range instructions {
		if line == "" {
			innerContent.WriteString("\n")
		} else {
			innerContent.WriteString(styles.instruction.Render(line))
			if line != instructions[len(instructions)-1] {
				innerContent.WriteString("\n")
			}
		}
	}

	// Wrap content in a bordered box
	boxedContent := styles.border.Width(boxWidth).Render(innerContent.String())

	// Add vertical centering
	topMargin := (m.height - 12) / 2
	if topMargin < 1 {
		topMargin = 1
	}
	topPadding := strings.Repeat("\n", topMargin)

	return topPadding + boxedContent
}

// renderProjectName renders the project name input screen.

// renderProjectName renders the project name input screen with a box border.
func (m WizardModel) renderProjectName() string {
	title := "Project Name"

	question := "Enter a name for your project:"

	instructions := []string{
		"",
		"Requirements:",
		"  • 2-50 characters",
		"  • Letters, numbers, hyphens, and underscores only",
		"",
		"Press Enter to continue",
		"Press Q or Esc to quit",
	}

	// Calculate dimensions
	boxWidth := min(60, m.width-4)
	contentWidth := boxWidth - 4

	// Build inner content
	var innerContent strings.Builder
	innerContent.WriteString(styles.title.Render(title))
	innerContent.WriteString("\n")
	innerContent.WriteString("\n")
	innerContent.WriteString(styles.question.Render(question))
	innerContent.WriteString("\n")
	// Text input
	m.textInput.Width = min(40, contentWidth-4)
	innerContent.WriteString(m.textInput.View())

	// Error message if any
	if m.err != nil {
		innerContent.WriteString("\n\n")
		innerContent.WriteString(styles.error.Render("⚠ " + m.err.Error()))
	}

	innerContent.WriteString("\n\n")
	for _, line := range instructions {
		if line == "" {
			innerContent.WriteString("\n")
		} else {
			innerContent.WriteString(styles.instruction.Render(line))
		}
	}

	// Wrap content in a bordered box
	boxedContent := styles.border.Width(boxWidth).Render(innerContent.String())

	// Add vertical centering
	topMargin := (m.height - 14) / 2
	if topMargin < 1 {
		topMargin = 1
	}
	topPadding := strings.Repeat("\n", topMargin)

	return topPadding + boxedContent
}

// renderSuccess renders the success screen.
func (m WizardModel) renderSuccess() string {
	title := "Project Initialized!"

	message := fmt.Sprintf("Your project '%s' is ready.", m.projectName)

	nextSteps := []string{
		"Next steps:",
		"  1. Review your project configuration",
		"  2. Start development with Claude Code",
		"  3. Run 'moai status' to verify setup",
		"",
		"Press Enter to exit",
	}

	// Calculate dimensions
	boxWidth := min(60, m.width-4)

	// Build inner content
	var innerContent strings.Builder
	innerContent.WriteString(styles.successTitle.Render("✓ " + title))
	innerContent.WriteString("\n")
	innerContent.WriteString(styles.subtitle.Render(message))
	innerContent.WriteString("\n\n")
	for _, line := range nextSteps {
		if line == "" {
			innerContent.WriteString("\n")
		} else {
			innerContent.WriteString(styles.instruction.Render(line))
			if line != nextSteps[len(nextSteps)-1] {
				innerContent.WriteString("\n")
			}
		}
	}

	// Wrap content in a bordered box
	boxedContent := styles.border.Width(boxWidth).Render(innerContent.String())

	// Add vertical centering
	topMargin := (m.height - 13) / 2
	if topMargin < 1 {
		topMargin = 1
	}
	topPadding := strings.Repeat("\n", topMargin)

	return topPadding + boxedContent
}

// GetResult returns the wizard result.
func (m *WizardModel) GetResult() *WizardResult {
	return m.result
}

// RunWizardTUI starts the wizard TUI and returns the result.
func RunWizardTUI() (*WizardResult, error) {
	model := NewWizardModel()
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	m, ok := finalModel.(WizardModel)
	if !ok {
		return nil, fmt.Errorf("unexpected model type")
	}

	return m.GetResult(), nil
}
