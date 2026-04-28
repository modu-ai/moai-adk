// Package cli provides GitHub Init Wizard for interactive Multi-LLM CI setup.
package cli

import (
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/modu-ai/moai-adk/internal/github"
)

// WizardState stores the current state of the GitHub Init Wizard.
type WizardState struct {
	Language     string            // ko, en, ja, zh
	SelectedLLMs  []string          // claude, codex, gemini, glm
	ModelChoices map[string]string // llm -> model mapping
	Triggers     TriggerConfig     // auto, comment, both
	ProjectDir   string
}

// TriggerConfig represents trigger configuration.
type TriggerConfig struct {
	AutoOnPR     bool   // Auto-run on PR creation
	OnComment    bool   // Trigger via comment only
	CommentWords string // Trigger comment words (default: /claude, /codex, etc.)
}

// Wizard provides interactive configuration for GitHub Actions integration.
type Wizard struct {
	state    *WizardState
	out      io.Writer
	messages *Messages
}

// NewWizard creates a new Wizard instance.
func NewWizard(out io.Writer) *Wizard {
	return &Wizard{
		state: &WizardState{
			ModelChoices: make(map[string]string),
		},
		out: out,
	}
}

// Run executes the Wizard flow using bubbletea TUI.
func (w *Wizard) Run() (*WizardState, error) {
	// Step 0: Check internet connection
	if err := w.checkInternetConnection(); err != nil {
		return nil, err
	}

	// Step 1: Language selection (TUI)
	if err := w.selectLanguageTUI(); err != nil {
		return nil, err
	}

	// Load messages for selected language
	w.loadMessages()

	// Step 2: LLM selection (TUI with multiple selection)
	if err := w.selectLLMsTUI(); err != nil {
		return nil, err
	}

	// Step 3: Model selection for each selected LLM (TUI)
	if err := w.selectModelsTUI(); err != nil {
		return nil, err
	}

	// Step 4: Trigger configuration
	if err := w.configureTriggers(); err != nil {
		return nil, err
	}

	// Step 5: Summary and confirmation (Yes/No TUI)
	if err := w.confirmAndFinishTUI(); err != nil {
		return nil, err
	}

	return w.state, nil
}

// checkInternetConnection verifies internet connectivity before proceeding.
func (w *Wizard) checkInternetConnection() error {
	if err := github.CheckInternetConnection(); err != nil {
		_, _ = fmt.Fprintf(w.out, "\n❌ %s\n\n", err.Error())
		_, _ = fmt.Fprintf(w.out, "moai github init requires an internet connection:\n")
		_, _ = fmt.Fprintf(w.out, "  • Download GitHub Actions Runner\n")
		_, _ = fmt.Fprintf(w.out, "  • Claude Code OAuth authentication\n")
		_, _ = fmt.Fprintf(w.out, "  • LLM API integration\n\n")
		_, _ = fmt.Fprintf(w.out, "Troubleshooting:\n")
		_, _ = fmt.Fprintf(w.out, "  1. Check your internet connection\n")
		_, _ = fmt.Fprintf(w.out, "  2. Verify firewall settings\n")
		_, _ = fmt.Fprintf(w.out, "  3. Configure proxy if needed\n\n")
		return fmt.Errorf("internet connection required")
	}
	return nil
}

// selectLanguageTUI runs the language selection TUI.
func (w *Wizard) selectLanguageTUI() error {
	model := NewLanguageModel()
	p := tea.NewProgram(model, tea.WithOutput(w.out))

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI execution failed: %w", err)
	}

	languageModel, ok := finalModel.(LanguageModel)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	w.state.Language = languageModel.GetSelectedLanguage()
	return nil
}

// loadMessages loads messages for the selected language.
func (w *Wizard) loadMessages() {
	w.messages = GetMessages(w.state.Language)
}

// selectLLMsTUI runs the LLM selection TUI with multiple selection.
func (w *Wizard) selectLLMsTUI() error {
	model := NewLLMModel(w.messages)
	p := tea.NewProgram(model, tea.WithOutput(w.out))

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI execution failed: %w", err)
	}

	llmModel, ok := finalModel.(LLMModel)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	w.state.SelectedLLMs = llmModel.GetSelectedLLMs()
	return nil
}

// selectModelsTUI runs model selection for each selected LLM using TUI.
func (w *Wizard) selectModelsTUI() error {
	for _, llm := range w.state.SelectedLLMs {
		model := NewModelChoiceModel(llm, w.messages)
		p := tea.NewProgram(model, tea.WithOutput(w.out))

		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("TUI execution failed for %s: %w", llm, err)
		}

		modelChoice, ok := finalModel.(ModelChoiceModel)
		if !ok {
			return fmt.Errorf("unexpected model type")
		}

		if modelChoice.Cancelled {
			return fmt.Errorf("model selection cancelled for %s", llm)
		}

		w.state.ModelChoices[llm] = modelChoice.Selected
	}
	return nil
}

// configureTriggers configures when code reviews run.
func (w *Wizard) configureTriggers() error {
	// Default configuration
	w.state.Triggers.AutoOnPR = true
	w.state.Triggers.OnComment = true
	return nil
}

// confirmAndFinishTUI displays configuration summary and confirms using Yes/No TUI.
func (w *Wizard) confirmAndFinishTUI() error {
	_, _ = fmt.Fprintln(w.out)
	_, _ = fmt.Fprintln(w.out, "✅ Configuration Summary / 설정 요약:")
	_, _ = fmt.Fprintln(w.out)
	_, _ = fmt.Fprintf(w.out, "Language / 언어: %s\n", w.state.Language)
	_, _ = fmt.Fprintf(w.out, "LLMs: %v\n", w.state.SelectedLLMs)
	_, _ = fmt.Fprintf(w.out, "Models / 모델:\n")
	for llm, model := range w.state.ModelChoices {
		_, _ = fmt.Fprintf(w.out, "  - %s: %s\n", llm, model)
	}
	_, _ = fmt.Fprintln(w.out)

	// Use Yes/No TUI for confirmation
	summary := "? Proceed with this configuration? / 이 설정으로 진행할까요?"
	model := NewYesNoModel(summary, w.messages)
	p := tea.NewProgram(model, tea.WithOutput(w.out))

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI execution failed: %w", err)
	}

	yesNoModel, ok := finalModel.(YesNoModel)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	if yesNoModel.Cancelled {
		return fmt.Errorf("cancelled by user / 사용자가 취소했습니다")
	}

	return nil
}
