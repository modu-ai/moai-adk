package ui

import (
	"context"
	"testing"
)

// --- Wizard headless tests (100% coverage target) ---

func TestWizardHeadless_FullDefaults(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{
		"project_name": "my-project",
		"language":     "Go",
		"framework":    "Cobra CLI",
		"features":     "LSP,Quality Gates",
		"user_name":    "GOOS",
		"conv_lang":    "ko",
	})

	w := NewWizard(theme, hm)
	result, err := w.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil WizardResult")
	}
	if result.ProjectName != "my-project" {
		t.Errorf("expected ProjectName 'my-project', got %q", result.ProjectName)
	}
	if result.Language != "Go" {
		t.Errorf("expected Language 'Go', got %q", result.Language)
	}
	if result.Framework != "Cobra CLI" {
		t.Errorf("expected Framework 'Cobra CLI', got %q", result.Framework)
	}
	if len(result.Features) != 2 {
		t.Fatalf("expected 2 features, got %d", len(result.Features))
	}
	if result.Features[0] != "LSP" {
		t.Errorf("expected first feature 'LSP', got %q", result.Features[0])
	}
	if result.Features[1] != "Quality Gates" {
		t.Errorf("expected second feature 'Quality Gates', got %q", result.Features[1])
	}
	if result.UserName != "GOOS" {
		t.Errorf("expected UserName 'GOOS', got %q", result.UserName)
	}
	if result.ConvLang != "ko" {
		t.Errorf("expected ConvLang 'ko', got %q", result.ConvLang)
	}
}

func TestWizardHeadless_NoDefaults_ReturnsError(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	// No defaults set

	w := NewWizard(theme, hm)
	result, err := w.Run(context.Background())
	if err != ErrHeadlessNoDefaults {
		t.Errorf("expected ErrHeadlessNoDefaults, got %v", err)
	}
	if result != nil {
		t.Error("expected nil WizardResult on error")
	}
}

func TestWizardHeadless_PartialDefaults_UsesEmptyForMissing(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{
		"project_name": "partial-project",
		"language":     "Python",
	})

	w := NewWizard(theme, hm)
	result, err := w.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ProjectName != "partial-project" {
		t.Errorf("expected 'partial-project', got %q", result.ProjectName)
	}
	if result.Language != "Python" {
		t.Errorf("expected 'Python', got %q", result.Language)
	}
	if result.Framework != "" {
		t.Errorf("expected empty framework, got %q", result.Framework)
	}
}

func TestWizardHeadless_ContextCancelled(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{"project_name": "test"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	w := NewWizard(theme, hm)
	result, err := w.Run(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if result != nil {
		t.Error("expected nil result on cancellation")
	}
}

func TestWizardHeadless_EmptyFeatures(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{
		"project_name": "proj",
		"language":     "Go",
		"features":     "",
	})

	w := NewWizard(theme, hm)
	result, err := w.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Features) != 0 {
		t.Errorf("expected 0 features for empty string, got %d", len(result.Features))
	}
}

// --- Wizard data helpers ---

func TestLanguageOptions_NotEmpty(t *testing.T) {
	opts := languageOptions()
	if len(opts) == 0 {
		t.Error("expected non-empty language options")
	}
	// Verify Go is in the list
	found := false
	for _, o := range opts {
		if o.Value == "Go" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'Go' in language options")
	}
}

func TestFrameworkOptions_Go(t *testing.T) {
	opts := frameworkOptions("Go")
	if len(opts) == 0 {
		t.Error("expected non-empty framework options for Go")
	}
}

func TestFrameworkOptions_Python(t *testing.T) {
	opts := frameworkOptions("Python")
	if len(opts) == 0 {
		t.Error("expected non-empty framework options for Python")
	}
}

func TestFrameworkOptions_Unknown(t *testing.T) {
	opts := frameworkOptions("UnknownLang")
	if len(opts) == 0 {
		t.Error("expected at least one fallback framework option")
	}
}

func TestFeatureOptions_NotEmpty(t *testing.T) {
	opts := featureOptions()
	if len(opts) == 0 {
		t.Error("expected non-empty feature options")
	}
}

func TestConvLangOptions_NotEmpty(t *testing.T) {
	opts := convLangOptions()
	if len(opts) == 0 {
		t.Error("expected non-empty conversation language options")
	}
	// Verify Korean is in the list
	found := false
	for _, o := range opts {
		if o.Value == "ko" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'ko' in conversation language options")
	}
}

// --- Comprehensive data option coverage ---

func TestLanguageOptions_AllExpected(t *testing.T) {
	opts := languageOptions()
	expected := map[string]bool{
		"Go":         false,
		"Python":     false,
		"TypeScript": false,
		"Java":       false,
		"Rust":       false,
		"PHP":        false,
	}
	for _, o := range opts {
		expected[o.Value] = true
	}
	for lang, found := range expected {
		if !found {
			t.Errorf("expected %q in language options", lang)
		}
	}
}

func TestLanguageOptions_HaveDescriptions(t *testing.T) {
	opts := languageOptions()
	for _, o := range opts {
		if o.Label == "" {
			t.Errorf("language option with value %q has empty label", o.Value)
		}
		if o.Desc == "" {
			t.Errorf("language option %q has empty description", o.Value)
		}
	}
}

func TestFeatureOptions_AllExpected(t *testing.T) {
	opts := featureOptions()
	expected := map[string]bool{
		"LSP":           false,
		"Quality Gates": false,
		"Git Hooks":     false,
		"Statusline":    false,
	}
	for _, o := range opts {
		expected[o.Value] = true
	}
	for feat, found := range expected {
		if !found {
			t.Errorf("expected %q in feature options", feat)
		}
	}
}

func TestFeatureOptions_HaveDescriptions(t *testing.T) {
	opts := featureOptions()
	for _, o := range opts {
		if o.Label == "" {
			t.Errorf("feature option with value %q has empty label", o.Value)
		}
		if o.Desc == "" {
			t.Errorf("feature option %q has empty description", o.Value)
		}
	}
}

func TestConvLangOptions_AllExpected(t *testing.T) {
	opts := convLangOptions()
	expected := map[string]bool{
		"en": false,
		"ko": false,
		"ja": false,
		"zh": false,
		"es": false,
		"fr": false,
		"de": false,
	}
	for _, o := range opts {
		expected[o.Value] = true
	}
	for lang, found := range expected {
		if !found {
			t.Errorf("expected %q in conversation language options", lang)
		}
	}
}

func TestConvLangOptions_HaveDescriptions(t *testing.T) {
	opts := convLangOptions()
	for _, o := range opts {
		if o.Label == "" {
			t.Errorf("conv lang option with value %q has empty label", o.Value)
		}
		if o.Desc == "" {
			t.Errorf("conv lang option %q has empty description", o.Value)
		}
	}
}

func TestFrameworkOptions_AllLanguages_NonEmpty(t *testing.T) {
	languages := []string{"Go", "Python", "TypeScript", "Java", "Rust", "PHP", "Unknown"}
	for _, lang := range languages {
		opts := frameworkOptions(lang)
		if len(opts) == 0 {
			t.Errorf("expected non-empty framework options for %s", lang)
		}
		for _, o := range opts {
			if o.Label == "" {
				t.Errorf("framework option for %s has empty label", lang)
			}
			if o.Value == "" {
				t.Errorf("framework option for %s has empty value", lang)
			}
		}
	}
}

// --- Wizard constructor and context tests ---

func TestNewWizard_ReturnsNonNil(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	w := NewWizard(theme, hm)
	if w == nil {
		t.Error("NewWizard should not return nil")
	}
}

func TestWizard_NonHeadless_CancelledContext(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	// NOT headless

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before Run

	w := NewWizard(theme, hm)
	result, err := w.Run(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if result != nil {
		t.Error("expected nil result on cancellation")
	}
}

func TestWizardHeadless_FeaturesWithWhitespace(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{
		"project_name": "proj",
		"features":     " LSP , , Quality Gates , ",
	})

	w := NewWizard(theme, hm)
	result, err := w.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Features) != 2 {
		t.Fatalf("expected 2 features (trimmed), got %d: %v", len(result.Features), result.Features)
	}
	if result.Features[0] != "LSP" {
		t.Errorf("expected first feature 'LSP', got %q", result.Features[0])
	}
	if result.Features[1] != "Quality Gates" {
		t.Errorf("expected second feature 'Quality Gates', got %q", result.Features[1])
	}
}

func TestWizardHeadless_ConvLangDefault(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{
		"project_name": "proj",
		"conv_lang":    "ja",
	})

	w := NewWizard(theme, hm)
	result, err := w.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ConvLang != "ja" {
		t.Errorf("expected ConvLang 'ja', got %q", result.ConvLang)
	}
}

func TestWizardHeadless_UserNameDefault(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{
		"project_name": "proj",
		"user_name":    "testuser",
	})

	w := NewWizard(theme, hm)
	result, err := w.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserName != "testuser" {
		t.Errorf("expected UserName 'testuser', got %q", result.UserName)
	}
}
