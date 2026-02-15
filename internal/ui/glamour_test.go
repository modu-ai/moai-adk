package ui

import (
	"os"
	"testing"
)

func TestRenderMarkdown_SimpleText(t *testing.T) {
	_ = os.Unsetenv("MOAI_NO_COLOR")
	_ = os.Unsetenv("NO_COLOR")

	md := "# Hello\n\nThis is **bold** text."
	out, err := RenderMarkdown(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty rendered output")
	}
}

func TestRenderMarkdown_NoColor_ReturnsUnchanged(t *testing.T) {
	_ = os.Setenv("MOAI_NO_COLOR", "true")
	defer func() { _ = os.Unsetenv("MOAI_NO_COLOR") }()

	md := "# Hello\n\nSome text."
	out, err := RenderMarkdown(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != md {
		t.Errorf("expected unchanged output in no-color mode, got %q", out)
	}
}

func TestRenderMarkdown_NoColor_EnvValue1(t *testing.T) {
	_ = os.Setenv("MOAI_NO_COLOR", "1")
	defer func() { _ = os.Unsetenv("MOAI_NO_COLOR") }()

	md := "# Hello"
	out, err := RenderMarkdown(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != md {
		t.Errorf("expected unchanged output, got %q", out)
	}
}

func TestRenderMarkdown_NoColor_EnvNO_COLOR(t *testing.T) {
	_ = os.Setenv("NO_COLOR", "1")
	defer func() { _ = os.Unsetenv("NO_COLOR") }()

	md := "# Hello"
	out, err := RenderMarkdown(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != md {
		t.Errorf("expected unchanged output with NO_COLOR, got %q", out)
	}
}

func TestRenderMarkdownWithTheme_DarkMode(t *testing.T) {
	_ = os.Unsetenv("MOAI_NO_COLOR")
	_ = os.Unsetenv("NO_COLOR")

	theme := NewTheme(ThemeConfig{Mode: "dark"})
	md := "# Hello\n\nDark mode text."
	out, err := RenderMarkdownWithTheme(md, theme)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty rendered output")
	}
}

func TestRenderMarkdownWithTheme_LightMode(t *testing.T) {
	_ = os.Unsetenv("MOAI_NO_COLOR")
	_ = os.Unsetenv("NO_COLOR")

	theme := NewTheme(ThemeConfig{Mode: "light"})
	md := "# Hello\n\nLight mode text."
	out, err := RenderMarkdownWithTheme(md, theme)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty rendered output")
	}
}

func TestRenderMarkdownWithTheme_NoColor(t *testing.T) {
	theme := NewTheme(ThemeConfig{NoColor: true})
	md := "# Hello\n\nNo color text."
	out, err := RenderMarkdownWithTheme(md, theme)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != md {
		t.Errorf("expected unchanged output in no-color mode, got %q", out)
	}
}

// --- noColorFromEnv tests ---

func TestNoColorFromEnv_Default(t *testing.T) {
	_ = os.Unsetenv("MOAI_NO_COLOR")
	_ = os.Unsetenv("NO_COLOR")

	if noColorFromEnv() {
		t.Error("expected false when no env vars set")
	}
}

func TestNoColorFromEnv_MOAI_NO_COLOR_True(t *testing.T) {
	_ = os.Setenv("MOAI_NO_COLOR", "true")
	defer func() { _ = os.Unsetenv("MOAI_NO_COLOR") }()

	if !noColorFromEnv() {
		t.Error("expected true when MOAI_NO_COLOR=true")
	}
}

func TestNoColorFromEnv_MOAI_NO_COLOR_1(t *testing.T) {
	_ = os.Setenv("MOAI_NO_COLOR", "1")
	defer func() { _ = os.Unsetenv("MOAI_NO_COLOR") }()

	if !noColorFromEnv() {
		t.Error("expected true when MOAI_NO_COLOR=1")
	}
}

func TestNoColorFromEnv_MOAI_NO_COLOR_Other(t *testing.T) {
	_ = os.Setenv("MOAI_NO_COLOR", "false")
	defer func() { _ = os.Unsetenv("MOAI_NO_COLOR") }()

	if noColorFromEnv() {
		t.Error("expected false when MOAI_NO_COLOR=false")
	}
}

func TestNoColorFromEnv_NO_COLOR(t *testing.T) {
	_ = os.Setenv("NO_COLOR", "anything")
	defer func() { _ = os.Unsetenv("NO_COLOR") }()

	if !noColorFromEnv() {
		t.Error("expected true when NO_COLOR is set")
	}
}

// --- terminalWidth tests ---

func TestTerminalWidth_Default(t *testing.T) {
	_ = os.Unsetenv("COLUMNS")

	w := terminalWidth()
	if w != defaultTerminalWidth {
		t.Errorf("expected default width %d, got %d", defaultTerminalWidth, w)
	}
}

func TestTerminalWidth_FromEnv(t *testing.T) {
	_ = os.Setenv("COLUMNS", "120")
	defer func() { _ = os.Unsetenv("COLUMNS") }()

	w := terminalWidth()
	if w != 120 {
		t.Errorf("expected width 120, got %d", w)
	}
}

func TestTerminalWidth_InvalidEnv(t *testing.T) {
	_ = os.Setenv("COLUMNS", "not-a-number")
	defer func() { _ = os.Unsetenv("COLUMNS") }()

	w := terminalWidth()
	if w != defaultTerminalWidth {
		t.Errorf("expected default width %d for invalid env, got %d", defaultTerminalWidth, w)
	}
}

func TestTerminalWidth_ZeroEnv(t *testing.T) {
	_ = os.Setenv("COLUMNS", "0")
	defer func() { _ = os.Unsetenv("COLUMNS") }()

	w := terminalWidth()
	if w != defaultTerminalWidth {
		t.Errorf("expected default width %d for zero env, got %d", defaultTerminalWidth, w)
	}
}

func TestTerminalWidth_NegativeEnv(t *testing.T) {
	_ = os.Setenv("COLUMNS", "-1")
	defer func() { _ = os.Unsetenv("COLUMNS") }()

	w := terminalWidth()
	if w != defaultTerminalWidth {
		t.Errorf("expected default width %d for negative env, got %d", defaultTerminalWidth, w)
	}
}
