// Package tui provides the MoAI-ADK terminal UI design system.
package tui_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// TestPromptBasic verifies that Prompt returns a non-empty string with host
// and path components.
func TestPromptBasic(t *testing.T) {
	th := tui.LightTheme()
	out := tui.Prompt("yuna@air", "~/work/my-app", "main", false, "", &th)
	if out == "" {
		t.Fatal("Prompt returned empty string")
	}
	if !strings.Contains(out, "yuna@air") {
		t.Error("Prompt: host not in output")
	}
	if !strings.Contains(out, "~/work/my-app") {
		t.Error("Prompt: path not in output")
	}
	checkGolden(t, "prompt-light-basic", out)
}

// TestPromptDark verifies Prompt in dark theme.
func TestPromptDark(t *testing.T) {
	th := tui.DarkTheme()
	out := tui.Prompt("goos@mac", "~/MoAI/moai-adk", "feat/SPEC-001", true, "", &th)
	if out == "" {
		t.Fatal("Prompt dark returned empty string")
	}
	checkGolden(t, "prompt-dark-basic", out)
}

// TestPromptDirty verifies that the dirty flag appears in the prompt.
func TestPromptDirty(t *testing.T) {
	th := tui.LightTheme()
	clean := tui.Prompt("user@host", "~/project", "main", false, "", &th)
	dirty := tui.Prompt("user@host", "~/project", "main", true, "", &th)
	if clean == dirty {
		t.Error("Prompt: dirty/clean outputs should differ")
	}
	checkGolden(t, "prompt-light-dirty", dirty)
}

// TestPromptWithCmd verifies Prompt with a command argument.
func TestPromptWithCmd(t *testing.T) {
	th := tui.LightTheme()
	out := tui.Prompt("yuna@air", "~/work", "main", false, "moai doctor", &th)
	if !strings.Contains(out, "moai doctor") {
		t.Error("Prompt: cmd not in output")
	}
	checkGolden(t, "prompt-light-cmd", out)
}

// TestPromptNoTheme verifies Prompt with nil theme falls back to light.
func TestPromptNoTheme(t *testing.T) {
	out := tui.Prompt("user@host", "~/", "main", false, "", nil)
	if out == "" {
		t.Fatal("Prompt nil theme returned empty string")
	}
}

// TestCursor verifies Cursor returns a non-empty string.
func TestCursor(t *testing.T) {
	th := tui.LightTheme()
	out := tui.Cursor(&th)
	if out == "" {
		t.Fatal("Cursor returned empty string")
	}
	checkGolden(t, "cursor-light", out)
}

// TestCursorDark verifies Cursor in dark theme.
func TestCursorDark(t *testing.T) {
	th := tui.DarkTheme()
	out := tui.Cursor(&th)
	if out == "" {
		t.Fatal("Cursor dark returned empty string")
	}
	checkGolden(t, "cursor-dark", out)
}
