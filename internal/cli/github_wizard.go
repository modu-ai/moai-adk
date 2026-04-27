// Package cli provides GitHub Init Wizard for interactive Multi-LLM CI setup.
package cli

import (
	"fmt"
	"io"

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

// Run executes the Wizard flow.
func (w *Wizard) Run() (*WizardState, error) {
	// Step 0: Check internet connection
	if err := w.checkInternetConnection(); err != nil {
		return nil, err
	}

	// Step 1: Language selection
	if err := w.selectLanguage(); err != nil {
		return nil, err
	}

	// Load messages for selected language
	w.loadMessages()

	// Step 2: LLM selection (multiple)
	if err := w.selectLLMs(); err != nil {
		return nil, err
	}

	// Step 3: Model configuration
	if err := w.selectModels(); err != nil {
		return nil, err
	}

	// Step 4: Trigger configuration
	if err := w.configureTriggers(); err != nil {
		return nil, err
	}

	// Step 5: Summary and confirmation
	if err := w.confirmAndFinish(); err != nil {
		return nil, err
	}

	return w.state, nil
}

// checkInternetConnection verifies internet connectivity before proceeding.
func (w *Wizard) checkInternetConnection() error {
	if err := github.CheckInternetConnection(); err != nil {
		fmt.Fprintf(w.out, "\n❌ %s\n\n", err.Error())
		fmt.Fprintf(w.out, "moai github init requires an internet connection:\n")
		fmt.Fprintf(w.out, "  • Download GitHub Actions Runner\n")
		fmt.Fprintf(w.out, "  • Claude Code OAuth authentication\n")
		fmt.Fprintf(w.out, "  • LLM API integration\n\n")
		fmt.Fprintf(w.out, "Troubleshooting:\n")
		fmt.Fprintf(w.out, "  1. Check your internet connection\n")
		fmt.Fprintf(w.out, "  2. Verify firewall settings\n")
		fmt.Fprintf(w.out, "  3. Configure proxy if needed\n\n")
		return fmt.Errorf("internet connection required")
	}
	return nil
}

// selectLanguage prompts for language selection.
func (w *Wizard) selectLanguage() error {
	fmt.Fprintln(w.out)
	fmt.Fprintln(w.out, "? Select your language / 사용하실 언어를 선택하세요:")
	fmt.Fprintln(w.out)
	fmt.Fprintln(w.out, "  1. 한국어 (ko)")
	fmt.Fprintln(w.out, "  2. English (en)")
	fmt.Fprintln(w.out, "  3. 日本語 (ja)")
	fmt.Fprintln(w.out, "  4. 中文 (zh)")
	fmt.Fprintln(w.out)

	var choice int
	for {
		fmt.Fprint(w.out, "Select (1-4) / 선택 (1-4): ")
		if _, err := fmt.Scanln(&choice); err != nil {
			fmt.Fprintln(w.out, "Invalid input. Please try again.")
			continue
		}
		if choice >= 1 && choice <= 4 {
			break
		}
		fmt.Fprintln(w.out, "Please enter a number between 1 and 4.")
	}

	languages := []string{"ko", "en", "ja", "zh"}
	w.state.Language = languages[choice-1]
	return nil
}

// loadMessages loads messages for the selected language.
func (w *Wizard) loadMessages() {
	w.messages = GetMessages(w.state.Language)
}

// selectLLMs prompts for LLM selection (multiple selection allowed).
func (w *Wizard) selectLLMs() error {
	fmt.Fprintln(w.out)
	fmt.Fprintln(w.out, w.messages.SelectLLM)
	fmt.Fprintln(w.out)
	fmt.Fprintln(w.out, "  1. Claude (Anthropic)")
	fmt.Fprintln(w.out, "  2. Codex (OpenAI) - Private repos only / 비공개 레포 전용")
	fmt.Fprintln(w.out, "  3. Gemini (Google)")
	fmt.Fprintln(w.out, "  4. GLM (Zhipu AI)")
	fmt.Fprintln(w.out)
	fmt.Fprintln(w.out, "Separate multiple choices with comma / 쉼표(,)로 구분하여 여러 개 선택 가능:")
	fmt.Fprintln(w.out)

	var input string
	fmt.Fprint(w.out, "Select / 선택 (e.g., 1,3): ")
	if _, err := fmt.Scanln(&input); err != nil {
		return err
	}

	llms := parseLLMSelection(input)
	if len(llms) == 0 {
		fmt.Fprintln(w.out, "At least one LLM must be selected / 최소 하나의 LLM을 선택해야 합니다.")
		return w.selectLLMs()
	}

	w.state.SelectedLLMs = llms
	return nil
}

// parseLLMSelection parses user input and returns selected LLMs.
// TODO: Implement robust parsing - "1,3" -> ["claude", "gemini"]
func parseLLMSelection(input string) []string {
	// Temporary: return default values
	return []string{"claude", "gemini"}
}

// selectModels configures models for each selected LLM.
func (w *Wizard) selectModels() error {
	for _, llm := range w.state.SelectedLLMs {
		if err := w.selectModelForLLM(llm); err != nil {
			return err
		}
	}
	return nil
}

// selectModelForLLM prompts for model selection for a specific LLM.
// TODO: Implement interactive model selection per LLM
func (w *Wizard) selectModelForLLM(llm string) error {
	switch llm {
	case "claude":
		w.state.ModelChoices[llm] = "claude-opus-4-7" // Default
	case "gemini":
		w.state.ModelChoices[llm] = "gemini-pro" // Default
	case "codex":
		w.state.ModelChoices[llm] = "gpt-4" // Default
	case "glm":
		w.state.ModelChoices[llm] = "glm-4" // Default
	}
	return nil
}

// configureTriggers configures when code reviews run.
// TODO: Implement interactive trigger configuration
func (w *Wizard) configureTriggers() error {
	// Default configuration
	w.state.Triggers.AutoOnPR = true
	w.state.Triggers.OnComment = true
	return nil
}

// confirmAndFinish displays configuration summary and confirms.
func (w *Wizard) confirmAndFinish() error {
	fmt.Fprintln(w.out)
	fmt.Fprintln(w.out, "✅ Configuration Summary / 설정 요약:")
	fmt.Fprintln(w.out)
	fmt.Fprintf(w.out, "Language / 언어: %s\n", w.state.Language)
	fmt.Fprintf(w.out, "LLMs: %v\n", w.state.SelectedLLMs)
	fmt.Fprintf(w.out, "Models / 모델: %v\n", w.state.ModelChoices)
	fmt.Fprintln(w.out)
	fmt.Fprintln(w.out, "? Proceed with this configuration? / 이 설정으로 진행할까요? (Y/n)")

	var confirm string
	fmt.Scanln(&confirm)
	if confirm == "n" || confirm == "N" {
		return fmt.Errorf("cancelled by user / 사용자가 취소했습니다")
	}

	return nil
}
