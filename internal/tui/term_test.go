// Package tui provides the MoAI-ADK terminal UI design system.
package tui_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// TestTermScreenshotMode verifies Term renders when MOAI_SCREENSHOT=1.
func TestTermScreenshotMode(t *testing.T) {
	t.Setenv("MOAI_SCREENSHOT", "1")
	th := tui.LightTheme()
	out := tui.Term("MoAI Terminal", 60, "body content here", tui.TermOpts{Theme: &th})
	if out == "" {
		t.Fatal("Term MOAI_SCREENSHOT=1 returned empty string")
	}
	checkGolden(t, "term-light-screenshot", out)
}

// TestTermNoScreenshot verifies Term returns empty string when MOAI_SCREENSHOT is not set.
func TestTermNoScreenshot(t *testing.T) {
	// Unset MOAI_SCREENSHOT (no-op if already unset)
	t.Setenv("MOAI_SCREENSHOT", "")
	th := tui.LightTheme()
	out := tui.Term("MoAI Terminal", 60, "body content here", tui.TermOpts{Theme: &th})
	if out != "" {
		t.Errorf("Term without MOAI_SCREENSHOT: expected empty string, got %q", out)
	}
}

// TestTermScreenshotDark verifies Term in dark theme with MOAI_SCREENSHOT=1.
func TestTermScreenshotDark(t *testing.T) {
	t.Setenv("MOAI_SCREENSHOT", "1")
	th := tui.DarkTheme()
	out := tui.Term("moai doctor", 70, "doctor output here", tui.TermOpts{Theme: &th})
	if out == "" {
		t.Fatal("Term dark screenshot returned empty string")
	}
	checkGolden(t, "term-dark-screenshot", out)
}

// TestTermScreenshotValue1 verifies that MOAI_SCREENSHOT=1 is the exact trigger.
func TestTermScreenshotValue1(t *testing.T) {
	t.Setenv("MOAI_SCREENSHOT", "1")
	th := tui.LightTheme()
	out := tui.Term("Test", 40, "content", tui.TermOpts{Theme: &th})
	if out == "" {
		t.Fatal("Term MOAI_SCREENSHOT=1 returned empty")
	}
}

// TestTermScreenshotValueOther verifies MOAI_SCREENSHOT=0 does NOT trigger rendering.
func TestTermScreenshotValueOther(t *testing.T) {
	t.Setenv("MOAI_SCREENSHOT", "0")
	th := tui.LightTheme()
	out := tui.Term("Test", 40, "content", tui.TermOpts{Theme: &th})
	if out != "" {
		t.Errorf("Term MOAI_SCREENSHOT=0: expected empty string, got %q", out)
	}
}
