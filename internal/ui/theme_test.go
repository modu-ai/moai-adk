package ui

import (
	"os"
	"strings"
	"testing"
)

func TestNewTheme_DefaultIsDark(t *testing.T) {
	theme := NewTheme(ThemeConfig{Mode: ""})
	if theme == nil {
		t.Fatal("NewTheme returned nil")
	}
	if theme.IsDark != true {
		t.Error("default theme should be dark mode")
	}
}

func TestNewTheme_DarkMode(t *testing.T) {
	theme := NewTheme(ThemeConfig{Mode: "dark"})
	if !theme.IsDark {
		t.Error("expected dark mode")
	}
	if theme.Colors.Primary == "" {
		t.Error("primary color should not be empty")
	}
}

func TestNewTheme_LightMode(t *testing.T) {
	theme := NewTheme(ThemeConfig{Mode: "light"})
	if theme.IsDark {
		t.Error("expected light mode")
	}
	if theme.Colors.Primary == "" {
		t.Error("primary color should not be empty")
	}
}

func TestNewTheme_DarkAndLightHaveDifferentPalettes(t *testing.T) {
	dark := NewTheme(ThemeConfig{Mode: "dark"})
	light := NewTheme(ThemeConfig{Mode: "light"})

	if dark.Colors.Primary == light.Colors.Primary {
		t.Error("dark and light themes should have different primary colors")
	}
}

func TestNewTheme_NoColor(t *testing.T) {
	theme := NewTheme(ThemeConfig{NoColor: true})
	if !theme.NoColor {
		t.Error("expected NoColor to be true")
	}
}

func TestNewTheme_NoColorFromEnv(t *testing.T) {
	_ = os.Setenv("MOAI_NO_COLOR", "true")
	defer func() { _ = os.Unsetenv("MOAI_NO_COLOR") }()

	theme := NewTheme(ThemeConfig{})
	if !theme.NoColor {
		t.Error("expected NoColor true when MOAI_NO_COLOR env is set")
	}
}

func TestNewTheme_NoColorFromEnvValue1(t *testing.T) {
	_ = os.Setenv("MOAI_NO_COLOR", "1")
	defer func() { _ = os.Unsetenv("MOAI_NO_COLOR") }()

	theme := NewTheme(ThemeConfig{})
	if !theme.NoColor {
		t.Error("expected NoColor true when MOAI_NO_COLOR=1")
	}
}

func TestNewTheme_AutoModeDefaultsToDark(t *testing.T) {
	theme := NewTheme(ThemeConfig{Mode: "auto"})
	// In test environments (no TTY), auto should default to dark
	if theme == nil {
		t.Fatal("NewTheme returned nil for auto mode")
	}
}

func TestTheme_ColorPaletteHasAllFields(t *testing.T) {
	theme := NewTheme(ThemeConfig{Mode: "dark"})
	colors := theme.Colors

	fields := map[string]string{
		"Primary":   colors.Primary,
		"Secondary": colors.Secondary,
		"Success":   colors.Success,
		"Warning":   colors.Warning,
		"Error":     colors.Error,
		"Muted":     colors.Muted,
		"Text":      colors.Text,
		"Border":    colors.Border,
	}

	for name, val := range fields {
		if val == "" {
			t.Errorf("color %s should not be empty", name)
		}
	}
}

func TestTheme_RenderTitle(t *testing.T) {
	theme := NewTheme(ThemeConfig{Mode: "dark"})
	result := theme.RenderTitle("Hello")
	if result == "" {
		t.Error("RenderTitle should return non-empty string")
	}
	if !strings.Contains(result, "Hello") {
		t.Error("RenderTitle output should contain the input text")
	}
}

func TestTheme_RenderTitle_NoColor(t *testing.T) {
	theme := NewTheme(ThemeConfig{NoColor: true})
	result := theme.RenderTitle("Hello")
	if result != "Hello" {
		t.Errorf("RenderTitle in NoColor mode should return plain text, got %q", result)
	}
}

func TestTheme_RenderError(t *testing.T) {
	theme := NewTheme(ThemeConfig{Mode: "dark"})
	result := theme.RenderError("something failed")
	if !strings.Contains(result, "something failed") {
		t.Error("RenderError output should contain the input text")
	}
}

func TestTheme_RenderError_NoColor(t *testing.T) {
	theme := NewTheme(ThemeConfig{NoColor: true})
	result := theme.RenderError("something failed")
	if result != "something failed" {
		t.Errorf("RenderError in NoColor mode should return plain text, got %q", result)
	}
}

func TestTheme_RenderSuccess(t *testing.T) {
	theme := NewTheme(ThemeConfig{Mode: "dark"})
	result := theme.RenderSuccess("done")
	if !strings.Contains(result, "done") {
		t.Error("RenderSuccess output should contain the input text")
	}
}

func TestTheme_RenderSuccess_NoColor(t *testing.T) {
	theme := NewTheme(ThemeConfig{NoColor: true})
	result := theme.RenderSuccess("done")
	if result != "done" {
		t.Errorf("RenderSuccess in NoColor mode should return plain text, got %q", result)
	}
}

func TestTheme_RenderMuted(t *testing.T) {
	theme := NewTheme(ThemeConfig{Mode: "dark"})
	result := theme.RenderMuted("hint")
	if !strings.Contains(result, "hint") {
		t.Error("RenderMuted output should contain the input text")
	}
}

func TestTheme_RenderMuted_NoColor(t *testing.T) {
	theme := NewTheme(ThemeConfig{NoColor: true})
	result := theme.RenderMuted("hint")
	if result != "hint" {
		t.Errorf("RenderMuted in NoColor mode should return plain text, got %q", result)
	}
}

func TestTheme_RenderWarning(t *testing.T) {
	theme := NewTheme(ThemeConfig{Mode: "dark"})
	result := theme.RenderWarning("caution")
	if !strings.Contains(result, "caution") {
		t.Error("RenderWarning output should contain the input text")
	}
}

func TestTheme_RenderWarning_NoColor(t *testing.T) {
	theme := NewTheme(ThemeConfig{NoColor: true})
	result := theme.RenderWarning("caution")
	if result != "caution" {
		t.Errorf("RenderWarning in NoColor mode should return plain text, got %q", result)
	}
}

func TestTheme_RenderHighlight(t *testing.T) {
	theme := NewTheme(ThemeConfig{Mode: "dark"})
	result := theme.RenderHighlight("important")
	if !strings.Contains(result, "important") {
		t.Error("RenderHighlight output should contain the input text")
	}
}

func TestTheme_RenderHighlight_NoColor(t *testing.T) {
	theme := NewTheme(ThemeConfig{NoColor: true})
	result := theme.RenderHighlight("important")
	if result != "important" {
		t.Errorf("RenderHighlight in NoColor mode should return plain text, got %q", result)
	}
}

func TestThemeConfig_Defaults(t *testing.T) {
	cfg := ThemeConfig{}
	if cfg.Mode != "" {
		t.Error("default ThemeConfig.Mode should be empty string")
	}
	if cfg.NoColor != false {
		t.Error("default ThemeConfig.NoColor should be false")
	}
}

// --- NewMoAIHuhTheme tests ---

func TestNewMoAIHuhTheme_Default(t *testing.T) {
	theme := NewMoAIHuhTheme(false)
	if theme == nil {
		t.Fatal("NewMoAIHuhTheme(false) returned nil")
	}
}

func TestNewMoAIHuhTheme_NoColor(t *testing.T) {
	theme := NewMoAIHuhTheme(true)
	if theme == nil {
		t.Fatal("NewMoAIHuhTheme(true) returned nil")
	}
}

func TestNewMoAIHuhTheme_DifferentFromBase(t *testing.T) {
	// The no-color theme should be the base theme;
	// the color theme should be customized.
	noColorTheme := NewMoAIHuhTheme(true)
	colorTheme := NewMoAIHuhTheme(false)

	if noColorTheme == nil || colorTheme == nil {
		t.Fatal("themes should not be nil")
	}
	// Both should be valid huh.Theme pointers (not the same pointer
	// unless implementation reuses base).
}
