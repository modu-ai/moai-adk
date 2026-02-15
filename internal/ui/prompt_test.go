package ui

import (
	"testing"
)

// --- Constructor tests ---

func TestNewPrompt_ReturnsNonNil(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	p := NewPrompt(theme, hm)
	if p == nil {
		t.Error("NewPrompt should not return nil")
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

func TestPromptHeadless_HeadlessDefaultOverridesOption(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{"name": "headless-value"})

	p := NewPrompt(theme, hm)
	result, err := p.Input("name", WithDefault("option-value"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "headless-value" {
		t.Errorf("expected 'headless-value', got %q", result)
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

func TestInputOptions_Combined(t *testing.T) {
	cfg := inputConfig{}
	opts := []InputOption{
		WithPlaceholder("hint"),
		WithDefault("def"),
		WithValidation(func(s string) error { return nil }),
	}
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.placeholder != "hint" {
		t.Errorf("expected placeholder 'hint', got %q", cfg.placeholder)
	}
	if cfg.defaultVal != "def" {
		t.Errorf("expected defaultVal 'def', got %q", cfg.defaultVal)
	}
	if cfg.validate == nil {
		t.Error("expected validate function to be set")
	}
}

// --- buildInputField tests ---

func TestBuildInputField_Basic(t *testing.T) {
	var value string
	cfg := inputConfig{}
	field := buildInputField("Enter name", cfg, &value)
	if field == nil {
		t.Fatal("buildInputField should not return nil")
	}
}

func TestBuildInputField_WithPlaceholder(t *testing.T) {
	var value string
	cfg := inputConfig{placeholder: "your name here"}
	field := buildInputField("Name", cfg, &value)
	if field == nil {
		t.Fatal("expected non-nil field with placeholder")
	}
}

func TestBuildInputField_WithValidation(t *testing.T) {
	var value string
	cfg := inputConfig{
		validate: func(s string) error {
			if s == "" {
				return ErrNoItems // reuse existing error for test
			}
			return nil
		},
	}
	field := buildInputField("Name", cfg, &value)
	if field == nil {
		t.Fatal("expected non-nil field with validation")
	}
}

func TestBuildInputField_WithAllOptions(t *testing.T) {
	var value string
	cfg := inputConfig{
		placeholder: "hint",
		defaultVal:  "default",
		validate:    func(s string) error { return nil },
	}
	field := buildInputField("Project", cfg, &value)
	if field == nil {
		t.Fatal("expected non-nil field")
	}
}

func TestBuildInputField_BindsValuePointer(t *testing.T) {
	var value string
	cfg := inputConfig{}
	field := buildInputField("Test", cfg, &value)
	if field == nil {
		t.Fatal("expected non-nil field")
	}
	if value != "" {
		t.Errorf("expected empty initial value, got %q", value)
	}
}

func TestBuildInputField_NoPlaceholder_NoValidation(t *testing.T) {
	var value string
	cfg := inputConfig{}
	field := buildInputField("Plain input", cfg, &value)
	if field == nil {
		t.Fatal("expected non-nil field")
	}
}

// --- buildConfirmField tests ---

func TestBuildConfirmField_Basic(t *testing.T) {
	var value bool
	field := buildConfirmField("Continue?", &value)
	if field == nil {
		t.Fatal("buildConfirmField should not return nil")
	}
}

func TestBuildConfirmField_DefaultTrue(t *testing.T) {
	value := true
	field := buildConfirmField("Proceed?", &value)
	if field == nil {
		t.Fatal("expected non-nil field")
	}
	if !value {
		t.Error("value should still be true")
	}
}

func TestBuildConfirmField_DefaultFalse(t *testing.T) {
	value := false
	field := buildConfirmField("Delete?", &value)
	if field == nil {
		t.Fatal("expected non-nil field")
	}
	if value {
		t.Error("value should still be false")
	}
}

func TestBuildConfirmField_BindsValuePointer(t *testing.T) {
	value := false
	field := buildConfirmField("Confirm?", &value)
	if field == nil {
		t.Fatal("expected non-nil field")
	}
	// Value pointer should be bound but unchanged
	if value != false {
		t.Error("value should not have changed")
	}
}
