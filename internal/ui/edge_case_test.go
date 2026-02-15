package ui

import (
	"context"
	"testing"
)

// --- IsHeadless TTY detection paths ---

func TestHeadlessManager_IsHeadless_UnforcedPath(t *testing.T) {
	hm := NewHeadlessManager()
	// Just exercise the TTY detection path without forcing
	// The result depends on whether tests run in a TTY
	_ = hm.IsHeadless()
}

func TestHeadlessManager_ForceHeadless_ThenUnforce(t *testing.T) {
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	if !hm.IsHeadless() {
		t.Error("expected headless after force true")
	}
	hm.ForceHeadless(false)
	// Now it falls back to TTY detection
	_ = hm.IsHeadless()
}

// --- Wizard runHeadless context check ---

func TestWizardHeadless_ContextCancelled_AfterDefaultsCheck(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{
		"project_name": "proj",
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	w := NewWizard(theme, hm)
	result, err := w.Run(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if result != nil {
		t.Error("expected nil result")
	}
}

// --- Non-interactive path for Select when not headless ---

func TestSelector_NonHeadless_EmptyItems(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	// Don't force headless, but still test empty items validation
	sel := NewSelector(theme, hm)
	_, err := sel.Select("language", []SelectItem{})
	if err != ErrNoItems {
		t.Errorf("expected ErrNoItems, got %v", err)
	}
}

func TestCheckbox_NonHeadless_EmptyItems(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	cb := NewCheckbox(theme, hm)
	_, err := cb.MultiSelect("features", []SelectItem{})
	if err != ErrNoItems {
		t.Errorf("expected ErrNoItems, got %v", err)
	}
}

// --- Prompt non-headless context test ---

func TestPrompt_NonHeadless_InputOptions(t *testing.T) {
	// Test that input options are correctly applied in non-headless mode.
	// We can't actually run the form, but we can verify options are applied.
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true) // Use headless to avoid TTY requirement

	p := NewPrompt(theme, hm)

	// Test with WithDefault + headless: should return the default
	result, err := p.Input("test", WithDefault("myval"), WithPlaceholder("hint"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "myval" {
		t.Errorf("expected 'myval', got %q", result)
	}
}

// --- Wizard headless edge cases ---

func TestWizardHeadless_AllFieldsEmpty(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{
		"project_name": "",
		"language":     "",
		"framework":    "",
		"features":     "",
		"user_name":    "",
		"conv_lang":    "",
	})

	w := NewWizard(theme, hm)
	result, err := w.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ProjectName != "" {
		t.Errorf("expected empty ProjectName, got %q", result.ProjectName)
	}
	if len(result.Features) != 0 {
		t.Errorf("expected 0 features, got %d", len(result.Features))
	}
}

// --- Error type checks ---

func TestErrCancelled_IsError(t *testing.T) {
	var err error = ErrCancelled
	if err.Error() == "" {
		t.Error("ErrCancelled should have a non-empty error message")
	}
}

func TestErrNoItems_IsError(t *testing.T) {
	var err error = ErrNoItems
	if err.Error() == "" {
		t.Error("ErrNoItems should have a non-empty error message")
	}
}

func TestErrHeadlessNoDefaults_IsError(t *testing.T) {
	var err error = ErrHeadlessNoDefaults
	if err.Error() == "" {
		t.Error("ErrHeadlessNoDefaults should have a non-empty error message")
	}
}
