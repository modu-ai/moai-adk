package ui

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
)

// promptImpl implements the Prompt interface using huh.Input and huh.Confirm.
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

// buildInputField creates a configured huh.Input field for text input.
// It applies placeholder and validation from the inputConfig.
// This function is separated from inputInteractive to enable unit testing
// of field configuration without requiring a TTY.
func buildInputField(label string, cfg inputConfig, value *string) *huh.Input {
	inp := huh.NewInput().
		Title(label).
		Value(value)

	if cfg.placeholder != "" {
		inp = inp.Placeholder(cfg.placeholder)
	}
	if cfg.validate != nil {
		inp = inp.Validate(cfg.validate)
	}

	return inp
}

// buildConfirmField creates a configured huh.Confirm field for Yes/No confirmation.
// This function is separated from confirmInteractive to enable unit testing
// of field configuration without requiring a TTY.
func buildConfirmField(label string, value *bool) *huh.Confirm {
	return huh.NewConfirm().
		Title(label).
		Value(value).
		Affirmative("Yes").
		Negative("No")
}

// inputInteractive runs a huh.Input form for text input.
func (p *promptImpl) inputInteractive(label string, cfg inputConfig) (string, error) {
	value := cfg.defaultVal
	inp := buildInputField(label, cfg, &value)

	form := huh.NewForm(huh.NewGroup(inp)).
		WithTheme(NewMoAIHuhTheme(p.theme.NoColor)).
		WithAccessible(p.headless.IsHeadless())

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrCancelled
		}
		return "", fmt.Errorf("prompt: %w", err)
	}

	return value, nil
}

// confirmInteractive runs a huh.Confirm form for Yes/No confirmation.
func (p *promptImpl) confirmInteractive(label string, defaultVal bool) (bool, error) {
	value := defaultVal
	conf := buildConfirmField(label, &value)

	form := huh.NewForm(huh.NewGroup(conf)).
		WithTheme(NewMoAIHuhTheme(p.theme.NoColor)).
		WithAccessible(p.headless.IsHeadless())

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, ErrCancelled
		}
		return false, fmt.Errorf("confirm: %w", err)
	}

	return value, nil
}
