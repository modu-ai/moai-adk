// Package tui provides the MoAI-ADK terminal UI design system.
package tui_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// TestHelpBarEmpty verifies HelpBar with no items returns an empty string.
func TestHelpBarEmpty(t *testing.T) {
	th := tui.LightTheme()
	out := tui.HelpBar(nil, &th)
	if out != "" {
		t.Errorf("HelpBar empty: expected empty string, got %q", out)
	}
}

// TestHelpBarSingle verifies HelpBar with a single item.
func TestHelpBarSingle(t *testing.T) {
	th := tui.LightTheme()
	out := tui.HelpBar([]tui.KeyHint{
		{Key: "enter", Label: "선택"},
	}, &th)
	if out == "" {
		t.Fatal("HelpBar single returned empty string")
	}
	if !strings.Contains(out, "enter") {
		t.Error("HelpBar: key not in output")
	}
	if !strings.Contains(out, "선택") {
		t.Error("HelpBar: label not in output")
	}
	checkGolden(t, "helpbar-light-single", out)
}

// TestHelpBarMulti verifies HelpBar with multiple items.
func TestHelpBarMulti(t *testing.T) {
	th := tui.LightTheme()
	out := tui.HelpBar([]tui.KeyHint{
		{Key: "↑↓", Label: "이동"},
		{Key: "enter", Label: "선택"},
		{Key: "q", Label: "종료"},
	}, &th)
	if out == "" {
		t.Fatal("HelpBar multi returned empty string")
	}
	checkGolden(t, "helpbar-light-multi", out)
}

// TestHelpBarDark verifies HelpBar in dark theme.
func TestHelpBarDark(t *testing.T) {
	th := tui.DarkTheme()
	out := tui.HelpBar([]tui.KeyHint{
		{Key: "tab", Label: "next field"},
		{Key: "esc", Label: "go back"},
	}, &th)
	if out == "" {
		t.Fatal("HelpBar dark returned empty string")
	}
	checkGolden(t, "helpbar-dark-multi", out)
}

// TestHelpBarKoEnMixed verifies HelpBar with Korean+English mixed labels.
func TestHelpBarKoEnMixed(t *testing.T) {
	th := tui.LightTheme()
	out := tui.HelpBar([]tui.KeyHint{
		{Key: "↑↓", Label: "이동 · move"},
		{Key: "enter", Label: "선택 · select"},
		{Key: "ctrl+c", Label: "취소"},
	}, &th)
	if out == "" {
		t.Fatal("HelpBar ko-en mixed returned empty string")
	}
	checkGolden(t, "helpbar-light-koen", out)
}
