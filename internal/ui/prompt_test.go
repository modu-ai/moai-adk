package ui

import (
	"errors"
	"strings"
	"testing"
)

// --- promptModel unit tests ---

func TestNewPromptModel(t *testing.T) {
	m := newPromptModel(testTheme(), "Project name", inputConfig{})
	if m.label != "Project name" {
		t.Errorf("expected label 'Project name', got %q", m.label)
	}
	if m.value != "" {
		t.Error("initial value should be empty")
	}
}

func TestPromptModel_View_ShowsLabel(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{})
	view := m.View()
	if !strings.Contains(view, "Name") {
		t.Error("View should contain the label")
	}
}

func TestPromptModel_View_ShowsPlaceholder(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{placeholder: "Enter name"})
	view := m.View()
	if !strings.Contains(view, "Enter name") {
		t.Error("View should show placeholder when value is empty")
	}
}

func TestPromptModel_View_HidesPlaceholderWhenTyping(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{placeholder: "Enter name"})
	m.value = "hello"
	view := m.View()
	if strings.Contains(view, "Enter name") {
		t.Error("View should hide placeholder when value is not empty")
	}
}

func TestPromptModel_TypeRunes(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{})
	m = m.addRunes([]rune("hello"))
	if m.value != "hello" {
		t.Errorf("expected 'hello', got %q", m.value)
	}
}

func TestPromptModel_Backspace(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{})
	m.value = "hello"
	m = m.backspace()
	if m.value != "hell" {
		t.Errorf("expected 'hell', got %q", m.value)
	}
}

func TestPromptModel_Backspace_Empty(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{})
	m = m.backspace()
	if m.value != "" {
		t.Error("backspace on empty should remain empty")
	}
}

func TestPromptModel_Validate_Success(t *testing.T) {
	validate := func(s string) error {
		if s == "" {
			return errors.New("cannot be empty")
		}
		return nil
	}
	m := newPromptModel(testTheme(), "Name", inputConfig{validate: validate})
	m.value = "valid"
	err := m.validateInput()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if m.errMsg != "" {
		t.Errorf("expected no error message, got %q", m.errMsg)
	}
}

func TestPromptModel_Validate_Failure(t *testing.T) {
	validate := func(s string) error {
		if s == "" {
			return errors.New("cannot be empty")
		}
		return nil
	}
	m := newPromptModel(testTheme(), "Name", inputConfig{validate: validate})
	m.value = ""
	err := m.validateInput()
	if err == nil {
		t.Error("expected validation error")
	}
}

func TestPromptModel_Validate_ShowsErrorInView(t *testing.T) {
	validate := func(s string) error {
		if s == "" {
			return errors.New("cannot be empty")
		}
		return nil
	}
	m := newPromptModel(testTheme(), "Name", inputConfig{validate: validate})
	m.value = ""
	m.validateInput()
	m.errMsg = "cannot be empty"
	view := m.View()
	if !strings.Contains(view, "cannot be empty") {
		t.Error("View should show validation error message")
	}
}

func TestPromptModel_DefaultValue(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{defaultVal: "default-name"})
	if m.value != "default-name" {
		t.Errorf("expected default value 'default-name', got %q", m.value)
	}
}

// --- confirmModel unit tests ---

func TestNewConfirmModel(t *testing.T) {
	m := newConfirmModel(testTheme(), "Continue?", true)
	if m.label != "Continue?" {
		t.Errorf("expected label 'Continue?', got %q", m.label)
	}
	if !m.value {
		t.Error("default value should be true")
	}
}

func TestConfirmModel_View_ShowsLabel(t *testing.T) {
	m := newConfirmModel(testTheme(), "Continue?", true)
	view := m.View()
	if !strings.Contains(view, "Continue?") {
		t.Error("View should contain the label")
	}
}

func TestConfirmModel_View_ShowsYesNo(t *testing.T) {
	m := newConfirmModel(testTheme(), "Continue?", true)
	view := m.View()
	if !strings.Contains(view, "Yes") && !strings.Contains(view, "yes") &&
		!strings.Contains(view, "Y") && !strings.Contains(view, "y") {
		t.Error("View should show Yes option")
	}
}

func TestConfirmModel_ToggleValue(t *testing.T) {
	m := newConfirmModel(testTheme(), "Continue?", true)
	m = m.toggle()
	if m.value {
		t.Error("value should be false after toggle")
	}
	m = m.toggle()
	if !m.value {
		t.Error("value should be true after double toggle")
	}
}

// --- Headless prompt tests ---

func TestPromptHeadless_ReturnsDefault(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{"project_name": "ci-project"})

	p := NewPrompt(theme, hm)
	result, err := p.Input("project_name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ci-project" {
		t.Errorf("expected 'ci-project', got %q", result)
	}
}

func TestPromptHeadless_NoDefault_ReturnsEmpty(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)

	p := NewPrompt(theme, hm)
	result, err := p.Input("project_name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestPromptHeadless_WithDefaultOption(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)

	p := NewPrompt(theme, hm)
	result, err := p.Input("project_name", WithDefault("fallback"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "fallback" {
		t.Errorf("expected 'fallback', got %q", result)
	}
}

func TestConfirmHeadless_ReturnsDefault(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)

	p := NewPrompt(theme, hm)
	result, err := p.Confirm("Proceed?", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true for default confirm")
	}
}

func TestConfirmHeadless_DefaultFalse(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)

	p := NewPrompt(theme, hm)
	result, err := p.Confirm("Proceed?", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result {
		t.Error("expected false for default confirm")
	}
}

// --- InputOption tests ---

func TestWithPlaceholder(t *testing.T) {
	cfg := inputConfig{}
	WithPlaceholder("Enter value")(&cfg)
	if cfg.placeholder != "Enter value" {
		t.Errorf("expected placeholder 'Enter value', got %q", cfg.placeholder)
	}
}

func TestWithValidation(t *testing.T) {
	fn := func(s string) error { return nil }
	cfg := inputConfig{}
	WithValidation(fn)(&cfg)
	if cfg.validate == nil {
		t.Error("expected validate function to be set")
	}
}

func TestWithDefault(t *testing.T) {
	cfg := inputConfig{}
	WithDefault("default-val")(&cfg)
	if cfg.defaultVal != "default-val" {
		t.Errorf("expected 'default-val', got %q", cfg.defaultVal)
	}
}
