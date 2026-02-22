package ui

import (
	"context"
	"strings"
)

// wizardImpl implements the Wizard interface.
type wizardImpl struct {
	theme    *Theme
	headless *HeadlessManager
}

// NewWizard creates a Wizard backed by the given theme and headless manager.
func NewWizard(theme *Theme, hm *HeadlessManager) Wizard {
	return &wizardImpl{theme: theme, headless: hm}
}

// Run executes the multi-step project initialization wizard.
// In headless mode it returns a WizardResult from stored defaults.
// It respects context cancellation at every step.
func (w *wizardImpl) Run(ctx context.Context) (*WizardResult, error) {
	// Check context before starting
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if w.headless.IsHeadless() {
		return w.runHeadless(ctx)
	}

	return w.runInteractive(ctx)
}

// runHeadless builds a WizardResult from stored defaults.
func (w *wizardImpl) runHeadless(ctx context.Context) (*WizardResult, error) {
	if !w.headless.HasDefaults() {
		return nil, ErrHeadlessNoDefaults
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &WizardResult{}

	if v, ok := w.headless.GetDefault("project_name"); ok {
		result.ProjectName = v
	}
	if v, ok := w.headless.GetDefault("language"); ok {
		result.Language = v
	}
	if v, ok := w.headless.GetDefault("framework"); ok {
		result.Framework = v
	}
	if v, ok := w.headless.GetDefault("features"); ok && v != "" {
		parts := strings.SplitSeq(v, ",")
		for p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result.Features = append(result.Features, trimmed)
			}
		}
	}
	if result.Features == nil {
		result.Features = []string{}
	}
	if v, ok := w.headless.GetDefault("user_name"); ok {
		result.UserName = v
	}
	if v, ok := w.headless.GetDefault("conv_lang"); ok {
		result.ConvLang = v
	}

	return result, nil
}

// runInteractive executes the wizard with bubbletea-based UI components.
// Each step checks context cancellation before proceeding.
func (w *wizardImpl) runInteractive(ctx context.Context) (*WizardResult, error) {
	sel := NewSelector(w.theme, w.headless)
	prompt := NewPrompt(w.theme, w.headless)
	cb := NewCheckbox(w.theme, w.headless)

	result := &WizardResult{}

	// Step 1: Project name
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	name, err := prompt.Input("Project name", WithPlaceholder("my-project"))
	if err != nil {
		return nil, err
	}
	result.ProjectName = name

	// Step 2: Language
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	lang, err := sel.Select("Programming language", languageOptions())
	if err != nil {
		return nil, err
	}
	result.Language = lang

	// Step 3: Framework (dynamic based on language)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	fw, err := sel.Select("Framework", frameworkOptions(lang))
	if err != nil {
		return nil, err
	}
	result.Framework = fw

	// Step 4: Features
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	features, err := cb.MultiSelect("Features", featureOptions())
	if err != nil {
		return nil, err
	}
	result.Features = features

	// Step 5: User name
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	userName, err := prompt.Input("User name", WithPlaceholder("Your name"))
	if err != nil {
		return nil, err
	}
	result.UserName = userName

	// Step 6: Conversation language
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	convLang, err := sel.Select("Conversation language", convLangOptions())
	if err != nil {
		return nil, err
	}
	result.ConvLang = convLang

	return result, nil
}

// --- Wizard data options ---

// languageOptions returns the available programming language choices.
func languageOptions() []SelectItem {
	return []SelectItem{
		{Label: "Go", Value: "Go", Desc: "Compiled, concurrent language"},
		{Label: "Python", Value: "Python", Desc: "Versatile scripting language"},
		{Label: "TypeScript", Value: "TypeScript", Desc: "Typed JavaScript"},
		{Label: "Java", Value: "Java", Desc: "Enterprise platform"},
		{Label: "Rust", Value: "Rust", Desc: "Systems programming"},
		{Label: "PHP", Value: "PHP", Desc: "Web development"},
	}
}

// frameworkOptions returns framework choices filtered by language.
func frameworkOptions(language string) []SelectItem {
	switch language {
	case "Go":
		return []SelectItem{
			{Label: "Cobra CLI", Value: "Cobra CLI", Desc: "CLI application framework"},
			{Label: "Gin", Value: "Gin", Desc: "HTTP web framework"},
			{Label: "Echo", Value: "Echo", Desc: "High performance web framework"},
			{Label: "Fiber", Value: "Fiber", Desc: "Express-inspired web framework"},
		}
	case "Python":
		return []SelectItem{
			{Label: "FastAPI", Value: "FastAPI", Desc: "Modern async API framework"},
			{Label: "Django", Value: "Django", Desc: "Full-stack web framework"},
			{Label: "Flask", Value: "Flask", Desc: "Lightweight web framework"},
		}
	case "TypeScript":
		return []SelectItem{
			{Label: "Next.js", Value: "Next.js", Desc: "React full-stack framework"},
			{Label: "NestJS", Value: "NestJS", Desc: "Progressive Node.js framework"},
			{Label: "Express", Value: "Express", Desc: "Minimal web framework"},
		}
	case "Java":
		return []SelectItem{
			{Label: "Spring Boot", Value: "Spring Boot", Desc: "Enterprise Java framework"},
			{Label: "Quarkus", Value: "Quarkus", Desc: "Cloud-native Java framework"},
		}
	case "Rust":
		return []SelectItem{
			{Label: "Axum", Value: "Axum", Desc: "Ergonomic web framework"},
			{Label: "Rocket", Value: "Rocket", Desc: "Web framework for Rust"},
		}
	case "PHP":
		return []SelectItem{
			{Label: "Laravel", Value: "Laravel", Desc: "PHP web framework"},
			{Label: "Symfony", Value: "Symfony", Desc: "Enterprise PHP framework"},
		}
	default:
		return []SelectItem{
			{Label: "None", Value: "None", Desc: "No framework"},
		}
	}
}

// featureOptions returns the available feature choices.
func featureOptions() []SelectItem {
	return []SelectItem{
		{Label: "LSP", Value: "LSP", Desc: "Language Server Protocol integration"},
		{Label: "Quality Gates", Value: "Quality Gates", Desc: "TRUST 5 quality validation"},
		{Label: "Git Hooks", Value: "Git Hooks", Desc: "Pre-commit and pre-push hooks"},
		{Label: "Statusline", Value: "Statusline", Desc: "Custom status display"},
	}
}

// convLangOptions returns the available conversation language choices.
func convLangOptions() []SelectItem {
	return []SelectItem{
		{Label: "English", Value: "en", Desc: "English"},
		{Label: "Korean", Value: "ko", Desc: "Korean"},
		{Label: "Japanese", Value: "ja", Desc: "Japanese"},
		{Label: "Chinese", Value: "zh", Desc: "Chinese"},
		{Label: "Spanish", Value: "es", Desc: "Spanish"},
		{Label: "French", Value: "fr", Desc: "French"},
		{Label: "German", Value: "de", Desc: "German"},
	}
}
