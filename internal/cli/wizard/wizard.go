package wizard

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// statuslineSegmentPrefix is the prefix used for statusline segment question IDs.
const statuslineSegmentPrefix = "statusline_seg_"

// Model is the Bubble Tea model for the wizard.
type Model struct {
	questions    []Question
	currentIndex int
	visibleIndex int // 1-based index for display (e.g., "[1/7]")
	result       *WizardResult
	styles       *Styles
	state        State
	cursor       int        // For select questions
	inputValue   string     // For input questions
	errorMsg     string     // Current error message
	allQuestions []Question // All questions (including conditional ones)
	locale       string     // Current UI locale for translations
}

// New creates a new wizard Model with the given questions.
func New(questions []Question, styles *Styles) Model {
	if styles == nil {
		styles = NewStyles()
	}

	// Store all questions and filter for initial display
	result := &WizardResult{}

	return Model{
		questions:    questions,
		allQuestions: questions,
		currentIndex: 0,
		visibleIndex: 1,
		result:       result,
		styles:       styles,
		state:        StateRunning,
		inputValue:   questions[0].Default,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}
	return m, nil
}

// handleKeyMsg processes keyboard input.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle cancellation
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		m.state = StateCancelled
		return m, tea.Quit
	}

	// Route to appropriate handler based on question type
	q := m.currentQuestion()
	if q == nil {
		m.state = StateCompleted
		return m, tea.Quit
	}

	switch q.Type {
	case QuestionTypeSelect:
		return m.handleSelectInput(msg)
	case QuestionTypeInput:
		return m.handleTextInput(msg)
	}

	return m, nil
}

// handleSelectInput handles input for select-type questions.
func (m Model) handleSelectInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	q := m.currentQuestion()
	numOptions := len(q.Options)

	switch msg.Type {
	case tea.KeyUp, tea.KeyShiftTab:
		m.cursor--
		if m.cursor < 0 {
			m.cursor = numOptions - 1
		}
	case tea.KeyDown, tea.KeyTab:
		m.cursor = (m.cursor + 1) % numOptions
	case tea.KeyEnter:
		if numOptions > 0 {
			m.saveSelectAnswer(q.Options[m.cursor].Value)
			return m.advance()
		}
	}

	return m, nil
}

// handleTextInput handles input for text-type questions.
func (m Model) handleTextInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		return m.submitTextInput()
	case tea.KeyBackspace:
		if len(m.inputValue) > 0 {
			m.inputValue = m.inputValue[:len(m.inputValue)-1]
		}
	case tea.KeyRunes:
		m.inputValue += string(msg.Runes)
		m.errorMsg = ""
	}

	return m, nil
}

// submitTextInput validates and saves text input.
func (m Model) submitTextInput() (tea.Model, tea.Cmd) {
	q := m.currentQuestion()
	value := strings.TrimSpace(m.inputValue)

	// Use default if empty and not required
	if value == "" && q.Default != "" {
		value = q.Default
	}

	// Validate required fields
	if q.Required && value == "" {
		uiStr := GetUIStrings(m.locale)
		m.errorMsg = uiStr.ErrorRequired
		return m, nil
	}

	m.saveInputAnswer(value)
	return m.advance()
}

// advance moves to the next question or completes the wizard.
func (m Model) advance() (tea.Model, tea.Cmd) {
	m.currentIndex++
	m.errorMsg = ""

	// Skip questions whose conditions are not met
	for m.currentIndex < len(m.allQuestions) {
		q := &m.allQuestions[m.currentIndex]
		if q.Condition == nil || q.Condition(m.result) {
			break
		}
		m.currentIndex++
	}

	// Check if we've finished all questions
	if m.currentIndex >= len(m.allQuestions) {
		m.state = StateCompleted
		return m, tea.Quit
	}

	// Update visible index
	m.visibleIndex++

	// Reset state for new question
	m.cursor = 0
	q := m.currentQuestion()
	if q != nil {
		m.inputValue = q.Default

		// Set cursor to default option for select questions
		if q.Type == QuestionTypeSelect {
			for i, opt := range q.Options {
				if opt.Value == q.Default {
					m.cursor = i
					break
				}
			}
		}
	}

	return m, nil
}

// currentQuestion returns the current question or nil if done.
func (m *Model) currentQuestion() *Question {
	if m.currentIndex >= len(m.allQuestions) {
		return nil
	}
	return &m.allQuestions[m.currentIndex]
}

// saveSelectAnswer saves the answer for a select question.
func (m *Model) saveSelectAnswer(value string) {
	m.saveAnswer(m.currentQuestion().ID, value)
}

// saveInputAnswer saves the answer for an input question.
func (m *Model) saveInputAnswer(value string) {
	m.saveAnswer(m.currentQuestion().ID, value)
}

// saveAnswer stores an answer in the result.
func (m *Model) saveAnswer(id, value string) {
	switch id {
	case "locale":
		m.result.Locale = value
		m.locale = value // Update UI locale for translations
	case "user_name":
		m.result.UserName = value
	case "project_name":
		m.result.ProjectName = value
	case "git_mode":
		m.result.GitMode = value
	case "git_provider":
		m.result.GitProvider = value
	case "github_username":
		m.result.GitHubUsername = value
	case "github_token":
		m.result.GitHubToken = value
	case "gitlab_instance_url":
		m.result.GitLabInstanceURL = value
	case "gitlab_username":
		m.result.GitLabUsername = value
	case "gitlab_token":
		m.result.GitLabToken = value
	case "git_commit_lang":
		m.result.GitCommitLang = value
	case "code_comment_lang":
		m.result.CodeCommentLang = value
	case "doc_lang":
		m.result.DocLang = value
	case "model_policy":
		m.result.ModelPolicy = value
	case "agent_teams_mode":
		m.result.AgentTeamsMode = value
	case "max_teammates":
		m.result.MaxTeammates = value
	case "default_model":
		m.result.DefaultModel = value
	case "statusline_preset":
		m.result.StatuslinePreset = value
	default:
		// Handle statusline segment toggles (statusline_seg_*)
		if strings.HasPrefix(id, statuslineSegmentPrefix) {
			segName := strings.TrimPrefix(id, statuslineSegmentPrefix)
			if m.result.StatuslineSegments == nil {
				m.result.StatuslineSegments = make(map[string]bool)
			}
			m.result.StatuslineSegments[segName] = (value == "true")
		}
	}
}

// View implements tea.Model.
func (m Model) View() string {
	if m.state == StateCompleted || m.state == StateCancelled {
		return ""
	}

	q := m.currentQuestion()
	if q == nil {
		return ""
	}

	// Get localized question based on current locale
	localizedQ := GetLocalizedQuestion(q, m.locale)

	var b strings.Builder

	// Progress indicator
	total := TotalVisibleQuestions(m.allQuestions, m.result)
	progress := m.styles.Progress.Render(fmt.Sprintf("[%d/%d]", m.visibleIndex, total))
	b.WriteString(progress)
	b.WriteString(" ")

	// Title
	b.WriteString(m.styles.Title.Render(localizedQ.Title))
	b.WriteString("\n")

	// Description
	if localizedQ.Description != "" {
		b.WriteString(m.styles.Description.Render(localizedQ.Description))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Render based on question type
	switch localizedQ.Type {
	case QuestionTypeSelect:
		b.WriteString(m.renderSelect(&localizedQ))
	case QuestionTypeInput:
		b.WriteString(m.renderInput(&localizedQ))
	}

	// Error message (localized)
	if m.errorMsg != "" {
		b.WriteString("\n")
		b.WriteString(m.styles.Error.Render(m.errorMsg))
	}

	// Help text (localized)
	b.WriteString("\n")
	b.WriteString(m.renderHelp(localizedQ.Type))

	return b.String()
}

// renderSelect renders a select question.
func (m Model) renderSelect(q *Question) string {
	var b strings.Builder

	for i, opt := range q.Options {
		cursor := "  "
		if i == m.cursor {
			cursor = m.styles.Cursor.Render("> ")
		}

		label := opt.Label
		if i == m.cursor {
			label = m.styles.SelectedOption.Render(label)
		} else {
			label = m.styles.Option.Render(label)
		}

		b.WriteString(cursor)
		b.WriteString(label)

		if opt.Desc != "" {
			b.WriteString(m.styles.Muted.Render(" - " + opt.Desc))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderInput renders an input question.
func (m Model) renderInput(q *Question) string {
	var b strings.Builder

	// Input prompt
	b.WriteString("> ")

	if m.inputValue == "" && q.Default != "" {
		// Show default as placeholder
		b.WriteString(m.styles.Placeholder.Render(q.Default))
	} else if m.inputValue == "" {
		// Show cursor position indicator
		b.WriteString(m.styles.Muted.Render("_"))
	} else {
		b.WriteString(m.inputValue)
		b.WriteString(m.styles.Muted.Render("_"))
	}

	return b.String()
}

// renderHelp renders the help text based on question type.
func (m Model) renderHelp(qType QuestionType) string {
	uiStr := GetUIStrings(m.locale)
	var help string
	switch qType {
	case QuestionTypeSelect:
		help = uiStr.HelpSelect
	case QuestionTypeInput:
		help = uiStr.HelpInput
	}
	return m.styles.Help.Render(help)
}

// Result returns the wizard result. Only valid after completion.
func (m Model) Result() *WizardResult {
	return m.result
}

// State returns the current wizard state.
func (m Model) State() State {
	return m.state
}

// Run executes the wizard and returns the result.
// This is a convenience function that handles the Bubble Tea program lifecycle.
func Run(questions []Question, styles *Styles) (*WizardResult, error) {
	if len(questions) == 0 {
		return nil, ErrNoQuestions
	}

	model := New(questions, styles)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("wizard error: %w", err)
	}

	m := finalModel.(Model)
	if m.State() == StateCancelled {
		return nil, ErrCancelled
	}

	return m.Result(), nil
}

// RunWithDefaults runs the wizard with default questions for the given project root.
func RunWithDefaults(projectRoot string) (*WizardResult, error) {
	questions := DefaultQuestions(projectRoot)
	return Run(questions, nil)
}
