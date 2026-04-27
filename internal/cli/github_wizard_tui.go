// Package cli provides bubbletea TUI components for GitHub Init Wizard.
package cli

import (
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LanguageModel is the bubbletea Model for language selection.
type LanguageModel struct {
	Choices []string
	Cursor  int
	Selected bool
}

// NewLanguageModel creates a new language selection model.
func NewLanguageModel() LanguageModel {
	return LanguageModel{
		Choices: []string{"한국어 (ko)", "English (en)", "日本語 (ja)", "中文 (zh)"},
		Cursor:  0,
	}
}

// Init implements bubbletea.Model.
func (m LanguageModel) Init() tea.Cmd {
	return nil
}

// Update implements bubbletea.Model.
func (m LanguageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Choices)-1 {
				m.Cursor++
			}
		case "enter", "r":
			m.Selected = true
			return m, tea.Quit
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View implements bubbletea.Model.
func (m LanguageModel) View() string {
	var b strings.Builder

	b.WriteString("\n? Select your language / 사용하실 언어를 선택하세요:\n\n")

	for i, choice := range m.Choices {
		cursor := " "
		if m.Cursor == i {
			cursor = ">"
		}
		b.WriteString(cursor + " " + choice + "\n")
	}

	b.WriteString("\n")
	b.WriteString(dimmedStyle.Render("Use ↑↓ to move, Enter to select / ↑↓로 이동, Enter로 선택"))

	return b.String()
}

// GetSelectedLanguage returns the selected language code.
func (m LanguageModel) GetSelectedLanguage() string {
	languages := []string{"ko", "en", "ja", "zh"}
	if m.Cursor >= 0 && m.Cursor < len(languages) {
		return languages[m.Cursor]
	}
	return "en" // Default
}

// LLMModel is the bubbletea Model for LLM selection (multiple).
type LLMModel struct {
	Choices     []LLMChoice
	Selected    map[string]bool
	Cursor      int
	Finished    bool
	Messages    *Messages
}

// LLMChoice represents an LLM option.
type LLMChoice struct {
	Name   string
	Value  string
	Notice string // Additional notice (e.g., "Private repos only")
}

// NewLLMModel creates a new LLM selection model.
func NewLLMModel(messages *Messages) LLMModel {
	return LLMModel{
		Choices: []LLMChoice{
			{Name: "Claude (Anthropic)", Value: "claude"},
			{Name: "Codex (OpenAI)", Value: "codex", Notice: "Private repos only"},
			{Name: "Gemini (Google)", Value: "gemini"},
			{Name: "GLM (Z.AI)", Value: "glm"},
		},
		Selected: make(map[string]bool),
		Cursor:   0,
		Messages: messages,
	}
}

// Init implements bubbletea.Model.
func (m LLMModel) Init() tea.Cmd {
	return nil
}

// Update implements bubbletea.Model.
func (m LLMModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Choices)-1 {
				m.Cursor++
			}
		case " ", "x": // Space or x to toggle selection
			choice := m.Choices[m.Cursor]
			m.Selected[choice.Value] = !m.Selected[choice.Value]
		case "enter", "r":
			// Check if at least one LLM is selected
			hasSelection := false
			for _, v := range m.Selected {
				if v {
					hasSelection = true
					break
				}
			}
			if !hasSelection {
				// Don't finish if nothing selected
				return m, nil
			}
			m.Finished = true
			return m, tea.Quit
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View implements bubbletea.Model.
func (m LLMModel) View() string {
	var b strings.Builder

	// Use localized messages
	if m.Messages != nil {
		b.WriteString("\n" + m.Messages.SelectLLM + "\n\n")
	} else {
		b.WriteString("\n? Select LLMs for code review (Space to toggle, Enter to confirm)\n")
		b.WriteString("? 코드 리뷰에 사용할 LLM을 선택하세요 (Space로 선택, Enter로 확인)\n\n")
	}

	for i, choice := range m.Choices {
		cursor := " "
		if m.Cursor == i {
			cursor = ">"
		}

		checked := " "
		if m.Selected[choice.Value] {
			checked = "✓"
		}

		b.WriteString(cursor + " [" + checked + "] " + choice.Name)

		if choice.Notice != "" {
			b.WriteString(" - " + choice.Notice)
		}

		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(dimmedStyle.Render("↑↓ Move, Space Toggle, Enter Confirm / ↑↓ 이동, Space 선택, Enter 확인"))

	return b.String()
}

// GetSelectedLLMs returns selected LLM values.
func (m LLMModel) GetSelectedLLMs() []string {
	var result []string
	for _, choice := range m.Choices {
		if m.Selected[choice.Value] {
			result = append(result, choice.Value)
		}
	}
	return result
}

// ModelChoiceModel is the bubbletea Model for model selection for a specific LLM.
type ModelChoiceModel struct {
	LLMName     string
	LLMNameKey   string // For message lookup
	Choices     []ModelChoice
	Cursor      int
	Selected    string
	Cancelled   bool
	Messages    *Messages
}

// ModelChoice represents a model option.
type ModelChoice struct {
	Name  string
	Value string
}

// NewModelChoiceModel creates a new model selection model for the given LLM.
func NewModelChoiceModel(llm string, messages *Messages) ModelChoiceModel {
	var choices []ModelChoice
	var modelName string
	var nameKey string

	switch llm {
	case "claude":
		modelName = "Claude"
		nameKey = "SelectClaudeModel"
		choices = []ModelChoice{
			{Name: "Opus 4.7 (최고 성능 / Highest performance)", Value: "claude-opus-4-7"},
			{Name: "Sonnet 4.6 (균형 / Balanced)", Value: "claude-sonnet-4-6"},
			{Name: "Haiku 4.5 (빠른 속도 / Fastest)", Value: "claude-haiku-4-5"},
		}
	case "codex":
		modelName = "Codex"
		nameKey = "SelectCodexModel"
		choices = []ModelChoice{
			{Name: "GPT-4", Value: "gpt-4"},
			{Name: "GPT-3.5 Turbo", Value: "gpt-3.5-turbo"},
		}
	case "gemini":
		modelName = "Gemini"
		nameKey = "SelectGeminiModel"
		choices = []ModelChoice{
			{Name: "Gemini Pro", Value: "gemini-pro"},
			{Name: "Gemini Flash", Value: "gemini-flash"},
		}
	case "glm":
		modelName = "GLM (Z.AI)"
		nameKey = "SelectZAIModel"
		choices = []ModelChoice{
			{Name: "GLM-4", Value: "glm-4"},
			{Name: "GLM-3-Turbo", Value: "glm-3-turbo"},
		}
	}

	return ModelChoiceModel{
		LLMName:   modelName,
		LLMNameKey: nameKey,
		Choices:   choices,
		Cursor:    0,
		Messages:  messages,
	}
}

// Init implements bubbletea.Model.
func (m ModelChoiceModel) Init() tea.Cmd {
	return nil
}

// Update implements bubbletea.Model.
func (m ModelChoiceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Choices)-1 {
				m.Cursor++
			}
		case "enter", "r":
			m.Selected = m.Choices[m.Cursor].Value
			return m, tea.Quit
		case "ctrl+c", "q":
			m.Cancelled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View implements bubbletea.Model.
func (m ModelChoiceModel) View() string {
	var b strings.Builder

	// Use localized messages
	prompt := "? Select model for " + m.LLMName + " / " + m.LLMName + " 모델 선택:"
	if m.Messages != nil {
		// Try to get localized prompt
		if localizedPrompt, ok := m.GetModelSelectionPrompt(m.LLMNameKey); ok {
			prompt = localizedPrompt
		}
	}

	b.WriteString("\n" + prompt + "\n\n")

	for i, choice := range m.Choices {
		cursor := " "
		if m.Cursor == i {
			cursor = ">"
		}
		b.WriteString(cursor + " " + choice.Name + "\n")
	}

	b.WriteString("\n")
	b.WriteString(dimmedStyle.Render("Use ↑↓ to move, Enter to select / ↑↓로 이동, Enter로 선택"))

	return b.String()
}

// GetModelSelectionPrompt returns localized prompt for model selection.
func (m ModelChoiceModel) GetModelSelectionPrompt(key string) (string, bool) {
	if m.Messages == nil {
		return "", false
	}

	prompts := map[string]string{
		"SelectClaudeModel":  m.Messages.SelectClaudeModel,
		"SelectCodexModel":   m.Messages.SelectCodexModel,
		"SelectGeminiModel": m.Messages.SelectGeminiModel,
		"SelectZAIModel":    m.Messages.SelectZAIModel,
	}

	if prompt, ok := prompts[key]; ok {
		return prompt, true
	}
	return "", false
}

// MessagesLanguageCode returns the language code from Messages.
func (m ModelChoiceModel) MessagesLanguageCode() string {
	// Extract from Messages or default to English
	// This is a simplified version - in production, Messages would have a LanguageCode() method
	return "en" // Default fallback
}

// YesNoModel is the bubbletea Model for Yes/No confirmation.
type YesNoModel struct {
	Message    string
	Confirmed bool
	Cancelled bool
	Messages   *Messages
}

// NewYesNoModel creates a new Yes/No confirmation model.
func NewYesNoModel(message string, messages *Messages) YesNoModel {
	return YesNoModel{
		Message:  message,
		Messages: messages,
	}
}

// Init implements bubbletea.Model.
func (m YesNoModel) Init() tea.Cmd {
	return nil
}

// Update implements bubbletea.Model.
func (m YesNoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			m.Confirmed = true
			return m, tea.Quit
		case "right", "l":
			m.Cancelled = true
			return m, tea.Quit
		case "ctrl+c", "q":
			m.Cancelled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View implements bubbletea.Model.
func (m YesNoModel) View() string {
	var yesLabel, noLabel string

	if m.Messages != nil {
		yesLabel, noLabel = m.GetYesNoLabels()
	} else {
		yesLabel, noLabel = "Yes", "No"
	}

	left := selectedStyle.Render("← "+yesLabel)
	right := noLabel + " →"

	return m.Message + "\n\n" + left + "    " + right
}

// GetYesNoLabels returns localized Yes/No labels.
func (m YesNoModel) GetYesNoLabels() (string, string) {
	if m.Messages == nil {
		return "Yes", "No"
	}

	return m.Messages.YesLabel, m.Messages.NoLabel
}

// Styling
var (
	dimmedStyle    = lipgloss.NewStyle().Faint(true)
	selectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("228"))
)
