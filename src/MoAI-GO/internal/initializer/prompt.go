package initializer

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/moai-adk-go/internal/output"
)

// LanguageCode represents a supported language code
type LanguageCode string

const (
	LanguageEnglish  LanguageCode = "en"
	LanguageKorean   LanguageCode = "ko"
	LanguageJapanese LanguageCode = "ja"
	LanguageChinese  LanguageCode = "zh"
)

// LanguageInfo contains information about a supported language
type LanguageInfo struct {
	Code        LanguageCode
	Name        string
	NativeName  string
	DisplayName string // e.g., "English (en)"
}

// SupportedLanguages lists all supported languages
var SupportedLanguages = []LanguageInfo{
	{Code: LanguageEnglish, Name: "English", NativeName: "English", DisplayName: "English (en)"},
	{Code: LanguageKorean, Name: "Korean", NativeName: "한국어", DisplayName: "Korean (ko)"},
	{Code: LanguageJapanese, Name: "Japanese", NativeName: "日本語", DisplayName: "Japanese (ja)"},
	{Code: LanguageChinese, Name: "Chinese", NativeName: "中文", DisplayName: "Chinese (zh)"},
}

// Prompter handles interactive user prompts
type Prompter struct {
	reader *bufio.Reader
}

// NewPrompter creates a new prompter
func NewPrompter() *Prompter {
	return &Prompter{
		reader: bufio.NewReader(os.Stdin),
	}
}

// PromptLanguage prompts the user to select a language
func (p *Prompter) PromptLanguage() (LanguageCode, error) {
	fmt.Println(output.InfoStyle.Render("\nSelect your preferred language:"))
	fmt.Println()

	// Display language options
	for i, lang := range SupportedLanguages {
		fmt.Printf("  %d. %s\n", i+1, lang.DisplayName)
	}
	fmt.Println()

	// Prompt for selection
	fmt.Print(output.InfoStyle.Render("Enter language number [1-4, default=1]: "))

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading language selection: %w", err)
	}

	input = strings.TrimSpace(input)

	// Default to English if no input
	if input == "" {
		fmt.Println(output.MutedStyle.Render("Selected: English (en)"))
		return LanguageEnglish, nil
	}

	// Parse selection
	var selection int
	if _, err := fmt.Sscanf(input, "%d", &selection); err != nil {
		return "", fmt.Errorf("invalid language selection: %w", err)
	}

	// Validate selection
	if selection < 1 || selection > len(SupportedLanguages) {
		return "", fmt.Errorf("invalid language selection: %d (must be 1-%d)", selection, len(SupportedLanguages))
	}

	lang := SupportedLanguages[selection-1]
	fmt.Println(output.MutedStyle.Render(fmt.Sprintf("Selected: %s", lang.DisplayName)))
	return lang.Code, nil
}

// PromptUserName prompts the user for their name
func (p *Prompter) PromptUserName() (string, error) {
	fmt.Println(output.InfoStyle.Render("\nEnter your name (for documentation):"))
	fmt.Print(output.InfoStyle.Render("Name: "))

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading user name: %w", err)
	}

	name := strings.TrimSpace(input)

	// Provide default if empty
	if name == "" {
		name = "Developer"
		fmt.Println(output.MutedStyle.Render("Using default name: Developer"))
	} else {
		fmt.Println(output.MutedStyle.Render(fmt.Sprintf("Hello, %s!", name)))
	}

	return name, nil
}

// PromptConfirm prompts the user for yes/no confirmation
func (p *Prompter) PromptConfirm(message string, defaultValue bool) (bool, error) {
	prompt := message
	if defaultValue {
		prompt += " [Y/n]: "
	} else {
		prompt += " [y/N]: "
	}

	fmt.Print(output.WarningStyle.Render(prompt))

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("error reading confirmation: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))

	// Default handling
	if input == "" {
		return defaultValue, nil
	}

	// Parse input
	return input == "y" || input == "yes", nil
}
