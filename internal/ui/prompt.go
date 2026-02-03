package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// promptImpl implements the Prompt interface.
type promptImpl struct {
	theme    *Theme
	headless *HeadlessManager
}

// NewPrompt creates a Prompt backed by the given theme and headless manager.
func NewPrompt(theme *Theme, hm *HeadlessManager) Prompt {
	return &promptImpl{theme: theme, headless: hm}
}

// Input displays a text input prompt and returns the entered value.
// In headless mode it returns the default from HeadlessManager or WithDefault option.
func (p *promptImpl) Input(label string, opts ...InputOption) (string, error) {
	cfg := inputConfig{}
	for _, o := range opts {
		o(&cfg)
	}

	if p.headless.IsHeadless() {
		return p.inputHeadless(label, cfg)
	}

	return p.inputInteractive(label, cfg)
}

// Confirm displays a Yes/No prompt and returns the boolean result.
// In headless mode it returns the provided default value immediately.
func (p *promptImpl) Confirm(label string, defaultVal bool) (bool, error) {
	if p.headless.IsHeadless() {
		return defaultVal, nil
	}

	return p.confirmInteractive(label, defaultVal)
}

// inputHeadless returns the headless default or the WithDefault option value.
func (p *promptImpl) inputHeadless(label string, cfg inputConfig) (string, error) {
	if val, ok := p.headless.GetDefault(label); ok {
		return val, nil
	}
	return cfg.defaultVal, nil
}

// inputInteractive runs a bubbletea program for text input.
func (p *promptImpl) inputInteractive(label string, cfg inputConfig) (string, error) {
	m := newPromptModel(p.theme, label, cfg)

	finalModel, err := programRunner(m)
	if err != nil {
		return "", fmt.Errorf("prompt: %w", err)
	}

	result := finalModel.(promptModel)
	if result.cancelled {
		return "", ErrCancelled
	}

	return result.value, nil
}

// confirmInteractive runs a bubbletea program for Yes/No confirmation.
func (p *promptImpl) confirmInteractive(label string, defaultVal bool) (bool, error) {
	m := newConfirmModel(p.theme, label, defaultVal)

	finalModel, err := programRunner(m)
	if err != nil {
		return false, fmt.Errorf("confirm: %w", err)
	}

	result := finalModel.(confirmModel)
	if result.cancelled {
		return false, ErrCancelled
	}

	return result.value, nil
}

// --- promptModel (bubbletea Model for text input) ---

// promptModel is the bubbletea Model for a text input prompt.
type promptModel struct {
	theme     *Theme
	label     string
	value     string
	cfg       inputConfig
	errMsg    string
	cancelled bool
	done      bool
}

// newPromptModel creates a promptModel with the initial state.
func newPromptModel(theme *Theme, label string, cfg inputConfig) promptModel {
	return promptModel{
		theme: theme,
		label: label,
		value: cfg.defaultVal,
		cfg:   cfg,
	}
}

// Init is the bubbletea initialization command.
func (m promptModel) Init() tea.Cmd {
	return nil
}

// Update processes messages and returns the updated model.
func (m promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		case tea.KeyEnter:
			if err := m.validateInput(); err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			m.done = true
			return m, tea.Quit
		case tea.KeyBackspace:
			m = m.backspace()
		case tea.KeyRunes:
			m = m.addRunes(msg.Runes)
		}
	}
	return m, nil
}

// View renders the prompt to a string.
func (m promptModel) View() string {
	if m.done || m.cancelled {
		return ""
	}

	var b strings.Builder

	b.WriteString(m.theme.RenderTitle(m.label))
	b.WriteString("\n")

	if m.value == "" && m.cfg.placeholder != "" {
		b.WriteString(m.theme.RenderMuted(m.cfg.placeholder))
	} else {
		b.WriteString(m.value)
	}
	b.WriteString("\n")

	if m.errMsg != "" {
		b.WriteString(m.theme.RenderError(m.errMsg))
		b.WriteString("\n")
	}

	return b.String()
}

// addRunes appends runes to the current value.
func (m promptModel) addRunes(runes []rune) promptModel {
	m.value += string(runes)
	m.errMsg = "" // clear error on new input
	return m
}

// backspace removes the last rune from the current value.
func (m promptModel) backspace() promptModel {
	if len(m.value) == 0 {
		return m
	}
	runes := []rune(m.value)
	m.value = string(runes[:len(runes)-1])
	m.errMsg = "" // clear error on edit
	return m
}

// validateInput runs the configured validation function.
func (m *promptModel) validateInput() error {
	if m.cfg.validate == nil {
		return nil
	}
	return m.cfg.validate(m.value)
}

// --- confirmModel (bubbletea Model for Yes/No) ---

// confirmModel is the bubbletea Model for a Yes/No confirmation prompt.
type confirmModel struct {
	theme     *Theme
	label     string
	value     bool
	cancelled bool
	done      bool
}

// newConfirmModel creates a confirmModel with the initial state.
func newConfirmModel(theme *Theme, label string, defaultVal bool) confirmModel {
	return confirmModel{
		theme: theme,
		label: label,
		value: defaultVal,
	}
}

// Init is the bubbletea initialization command.
func (m confirmModel) Init() tea.Cmd {
	return nil
}

// Update processes messages and returns the updated model.
func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		case tea.KeyEnter:
			m.done = true
			return m, tea.Quit
		case tea.KeyLeft, tea.KeyRight:
			m = m.toggle()
		case tea.KeyRunes:
			if len(msg.Runes) == 1 {
				switch msg.Runes[0] {
				case 'y', 'Y':
					m.value = true
				case 'n', 'N':
					m.value = false
				}
			}
		}
	}
	return m, nil
}

// View renders the confirmation prompt to a string.
func (m confirmModel) View() string {
	if m.done || m.cancelled {
		return ""
	}

	var b strings.Builder

	b.WriteString(m.theme.RenderTitle(m.label))
	b.WriteString(" ")

	if m.value {
		b.WriteString(m.theme.RenderHighlight("Yes"))
		b.WriteString(" / ")
		b.WriteString(m.theme.RenderMuted("No"))
	} else {
		b.WriteString(m.theme.RenderMuted("Yes"))
		b.WriteString(" / ")
		b.WriteString(m.theme.RenderHighlight("No"))
	}
	b.WriteString("\n")

	return b.String()
}

// toggle flips the current value.
func (m confirmModel) toggle() confirmModel {
	m.value = !m.value
	return m
}
