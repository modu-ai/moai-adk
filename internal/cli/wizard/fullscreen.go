package wizard

// @MX:NOTE: [AUTO] Full-screen Bubble Tea wizard for moai init (layout v2.2).
// Replaces the sequential huh.Form pattern that produced verbose breadcrumb
// accumulation. Uses tea.WithAltScreen so every step redraws a single page.

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// wizardState tracks the high-level position of the bubbletea model.
type wizardState int

const (
	stateEditing wizardState = iota
	stateReview
	stateDone
	stateCancelled
)

// reviewChoice indices used for the review pane.
const (
	reviewIdxProceed = 0
	reviewIdxEdit    = 1
	reviewIdxCancel  = 2
)

// wizardModel is the bubbletea.Model that drives the init wizard.
type wizardModel struct {
	questions []Question
	result    *WizardResult
	locale    string
	theme     tui.Theme

	visible   []int // indices into questions visible under the current result
	stepIdx   int   // 0-based position within visible
	selectIdx int   // cursor inside a select question's options
	textInput textinput.Model

	width  int
	height int

	state     wizardState
	reviewIdx int
	errMsg    string
}

// newWizardModel constructs an initial model for the given question list.
func newWizardModel(questions []Question, locale string) wizardModel {
	ti := textinput.New()
	ti.Prompt = "▸ "
	ti.CharLimit = 200
	ti.Width = 48

	m := wizardModel{
		questions: questions,
		result:    &WizardResult{},
		locale:    locale,
		theme:     tui.ResolveOS(),
		textInput: ti,
		state:     stateEditing,
	}
	m.recomputeVisible()
	m.activateCurrent()
	return m
}

// recomputeVisible refreshes the visible-question index list according to the
// condition predicates evaluated against the current result.
func (m *wizardModel) recomputeVisible() {
	m.visible = m.visible[:0]
	for i := range m.questions {
		q := &m.questions[i]
		if q.Condition != nil && !q.Condition(m.result) {
			continue
		}
		m.visible = append(m.visible, i)
	}
}

// activateCurrent primes the input or select cursor for the question at
// stepIdx using either the previously saved answer or the question default.
func (m *wizardModel) activateCurrent() {
	if m.stepIdx >= len(m.visible) {
		return
	}
	q := &m.questions[m.visible[m.stepIdx]]

	if q.Type == QuestionTypeInput {
		prev := storedAnswer(q.ID, m.result)
		if prev == "" && q.Default != "" {
			m.textInput.Placeholder = q.Default
			m.textInput.SetValue("")
		} else {
			m.textInput.Placeholder = ""
			m.textInput.SetValue(prev)
		}
		m.textInput.Focus()
		return
	}

	m.selectIdx = 0
	prev := storedAnswer(q.ID, m.result)
	if prev == "" {
		prev = q.Default
	}
	if prev != "" {
		for i, opt := range q.Options {
			if opt.Value == prev {
				m.selectIdx = i
				break
			}
		}
	}
}

// Init satisfies tea.Model.
func (m wizardModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles tea.Msg dispatch by state.
func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch m.state {
		case stateEditing:
			return m.updateEditing(msg)
		case stateReview:
			return m.updateReview(msg)
		}
	}
	return m, nil
}

func (m wizardModel) updateEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.stepIdx >= len(m.visible) {
		m.state = stateReview
		m.reviewIdx = reviewIdxProceed
		return m, nil
	}
	q := &m.questions[m.visible[m.stepIdx]]

	switch msg.String() {
	case "ctrl+c":
		m.state = stateCancelled
		return m, tea.Quit
	}

	if q.Type == QuestionTypeSelect {
		return m.updateSelect(msg, q)
	}
	return m.updateInput(msg, q)
}

func (m wizardModel) updateSelect(msg tea.KeyMsg, q *Question) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateCancelled
		return m, tea.Quit
	case "up", "k":
		if m.selectIdx > 0 {
			m.selectIdx--
		} else {
			m.selectIdx = len(q.Options) - 1
		}
	case "down", "j":
		if m.selectIdx < len(q.Options)-1 {
			m.selectIdx++
		} else {
			m.selectIdx = 0
		}
	case "enter":
		chosen := q.Options[m.selectIdx].Value
		saveAnswer(q.ID, chosen, m.result, nil)
		return m.advance()
	}
	return m, nil
}

func (m wizardModel) updateInput(msg tea.KeyMsg, q *Question) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		val := strings.TrimSpace(m.textInput.Value())
		if val == "" {
			val = q.Default
		}
		if q.Required && val == "" {
			m.errMsg = "This field is required."
			return m, nil
		}
		m.errMsg = ""
		saveAnswer(q.ID, val, m.result, nil)
		return m.advance()
	case "esc":
		m.state = stateCancelled
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// advance moves to the next visible question or transitions to Review.
func (m wizardModel) advance() (tea.Model, tea.Cmd) {
	m.stepIdx++
	m.recomputeVisible()
	if m.stepIdx >= len(m.visible) {
		m.state = stateReview
		m.reviewIdx = reviewIdxProceed
		return m, nil
	}
	m.activateCurrent()
	return m, nil
}

func (m wizardModel) updateReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		m.state = stateCancelled
		return m, tea.Quit
	case "up", "k":
		if m.reviewIdx > 0 {
			m.reviewIdx--
		} else {
			m.reviewIdx = reviewIdxCancel
		}
	case "down", "j":
		if m.reviewIdx < reviewIdxCancel {
			m.reviewIdx++
		} else {
			m.reviewIdx = reviewIdxProceed
		}
	case "enter":
		switch m.reviewIdx {
		case reviewIdxProceed:
			m.state = stateDone
			return m, tea.Quit
		case reviewIdxEdit:
			m.result = &WizardResult{}
			m.stepIdx = 0
			m.state = stateEditing
			m.errMsg = ""
			m.recomputeVisible()
			m.activateCurrent()
			return m, nil
		case reviewIdxCancel:
			m.state = stateCancelled
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the current page in alt-screen mode.
func (m wizardModel) View() string {
	if m.width == 0 {
		return ""
	}
	switch m.state {
	case stateReview:
		return m.renderReview()
	case stateDone, stateCancelled:
		return ""
	}
	return m.renderEditing()
}

// renderEditing returns the full wizard page when the user is answering a
// question (header + progress bar + 2-pane body + footer).
func (m wizardModel) renderEditing() string {
	if m.stepIdx >= len(m.visible) {
		return ""
	}
	q := &m.questions[m.visible[m.stepIdx]]
	th := m.theme

	accent := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent)).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Dim))

	projectName := m.result.ProjectName
	if projectName == "" {
		projectName = "(new project)"
	}
	header := accent.Render(fmt.Sprintf("moai init · %s", projectName))

	total := len(m.visible)
	current := m.stepIdx + 1
	stepLabel := fmt.Sprintf("Step %d of %d", current, total)
	barWidth := 24
	filled := current * barWidth / total
	if filled > barWidth {
		filled = barWidth
	}
	progressBar := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent)).
		Render(strings.Repeat("━", filled)) +
		lipgloss.NewStyle().Foreground(lipgloss.Color(th.Faint)).
		Render(strings.Repeat("░", barWidth-filled))
	pct := current * 100 / total
	progressLine := stepLabel + "   " + progressBar + "  " + dim.Render(fmt.Sprintf("%d%%", pct))

	progressPanel := m.renderProgressPanel(th)
	questionPanel := m.renderQuestionPanel(q, th)
	body := lipgloss.JoinHorizontal(lipgloss.Top, progressPanel, "  ", questionPanel)

	var footer string
	if q.Type == QuestionTypeInput {
		footer = dim.Render("type to edit  ·  ⏎ next  ·  esc cancel")
	} else {
		footer = dim.Render("↑↓ navigate  ·  ⏎ select  ·  esc cancel")
	}

	var errLine string
	if m.errMsg != "" {
		errLine = "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color(th.Danger)).
			Render("✗ "+m.errMsg)
	}

	content := header + "\n\n" + progressLine + "\n\n" + body + errLine + "\n\n" + footer

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(th.Accent)).
		Padding(1, 2).
		Render(content)
}

// renderProgressPanel returns the left-hand progress side panel showing
// answered (✓), current (▸), and pending (·) steps.
func (m wizardModel) renderProgressPanel(th tui.Theme) string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Dim)).Bold(true)
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Success))
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent)).Bold(true)
	faintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Faint))

	lines := []string{dim.Render("Progress")}
	for i, qIdx := range m.visible {
		q := &m.questions[qIdx]
		label := shortLabel(q.ID)
		switch {
		case i < m.stepIdx:
			lines = append(lines, successStyle.Render("✓ "+label))
		case i == m.stepIdx:
			lines = append(lines, accentStyle.Render("▸ "+label))
		default:
			lines = append(lines, faintStyle.Render("· "+label))
		}
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(th.Rule)).
		Padding(0, 1).
		Width(20).
		Render(strings.Join(lines, "\n"))
}

// renderQuestionPanel renders the right-hand panel that shows the current
// question and its input (select option list or text input).
func (m wizardModel) renderQuestionPanel(q *Question, th tui.Theme) string {
	lq := GetLocalizedQuestion(q, m.locale)

	title := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent)).Bold(true).Render(lq.Title)
	desc := ""
	if lq.Description != "" {
		desc = lipgloss.NewStyle().Foreground(lipgloss.Color(th.Dim)).Render(lq.Description)
	}

	lines := []string{title}
	if desc != "" {
		lines = append(lines, "", desc)
	}
	lines = append(lines, "")

	if q.Type == QuestionTypeSelect {
		for i, opt := range lq.Options {
			cursor := "  "
			prefix := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Faint)).Render("◇")
			labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Fg))
			if i == m.selectIdx {
				cursor = lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent)).Render("▸ ")
				prefix = lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent)).Render("◆")
				labelStyle = labelStyle.Bold(true)
			}
			line := cursor + prefix + " " + labelStyle.Render(opt.Label)
			if opt.Desc != "" {
				line += "\n      " + lipgloss.NewStyle().Foreground(lipgloss.Color(th.Dim)).Render(opt.Desc)
			}
			lines = append(lines, line)
		}
	} else {
		lines = append(lines, m.textInput.View())
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(th.Rule)).
		Padding(0, 1).
		Width(54).
		Render(strings.Join(lines, "\n"))
}

// renderReview returns the Review page shown after the last question.
func (m wizardModel) renderReview() string {
	th := m.theme
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent)).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Dim))
	body := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Fg))

	maxLabel := 0
	for _, qIdx := range m.visible {
		l := shortLabel(m.questions[qIdx].ID)
		if len(l) > maxLabel {
			maxLabel = len(l)
		}
	}
	var rows []string
	for _, qIdx := range m.visible {
		q := &m.questions[qIdx]
		rows = append(rows, fmt.Sprintf("  %-*s   %s",
			maxLabel,
			body.Render(shortLabel(q.ID)),
			body.Render(reviewValue(q, m.result))))
	}

	choices := []string{
		"Proceed — deploy templates",
		"Edit answers — restart wizard",
		"Cancel",
	}
	var choiceLines []string
	for i, c := range choices {
		cursor := "  "
		prefix := dim.Render("◇")
		style := body
		if i == m.reviewIdx {
			cursor = lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent)).Render("▸ ")
			prefix = lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent)).Render("◆")
			style = body.Bold(true)
		}
		choiceLines = append(choiceLines, cursor+prefix+" "+style.Render(c))
	}

	content := accent.Render("Review") + "\n\n" +
		strings.Join(rows, "\n") + "\n\n" +
		dim.Render("Proceed with these settings?") + "\n\n" +
		strings.Join(choiceLines, "\n") + "\n\n" +
		dim.Render("↑↓ navigate  ·  ⏎ confirm  ·  esc cancel")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(th.Accent)).
		Padding(1, 2).
		Render(content)
}

// RunFullScreen runs the bubbletea wizard for the supplied question list.
// On user cancellation the function returns (nil, ErrCancelled).
func RunFullScreen(questions []Question, locale string) (*WizardResult, error) {
	if len(questions) == 0 {
		return nil, ErrNoQuestions
	}
	m := newWizardModel(questions, locale)
	final, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return nil, fmt.Errorf("wizard: %w", err)
	}
	fm, ok := final.(wizardModel)
	if !ok {
		return nil, fmt.Errorf("wizard: unexpected model type %T", final)
	}
	if fm.state == stateCancelled {
		return nil, ErrCancelled
	}
	return fm.result, nil
}

// RunFullScreenDefaults is the bubbletea analogue of RunWithDefaults: it
// builds the standard question set for the given project root and runs the
// full-screen wizard.
func RunFullScreenDefaults(projectRoot, locale string) (*WizardResult, error) {
	return RunFullScreen(DefaultQuestions(projectRoot), locale)
}
