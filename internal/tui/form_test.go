// Package tui provides the MoAI-ADK terminal UI design system.
package tui_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// TestRadioRowSelected verifies that the selected option uses ◆ prefix.
func TestRadioRowSelected(t *testing.T) {
	th := tui.LightTheme()
	opts := []tui.RadioOption{
		{Label: "Go", Selected: true},
		{Label: "Python", Selected: false},
		{Label: "TypeScript", Selected: false},
	}
	out := tui.RadioRow(opts, &th)
	if out == "" {
		t.Fatal("RadioRow returned empty string")
	}
	if !strings.Contains(out, "◆") {
		t.Error("RadioRow: selected item missing ◆ prefix")
	}
	if !strings.Contains(out, "◇") {
		t.Error("RadioRow: unselected item missing ◇ prefix")
	}
	checkGolden(t, "radio-light-selected", out)
}

// TestRadioRowDark verifies RadioRow in dark theme.
func TestRadioRowDark(t *testing.T) {
	th := tui.DarkTheme()
	opts := []tui.RadioOption{
		{Label: "Go", Selected: false},
		{Label: "Python", Selected: true},
	}
	out := tui.RadioRow(opts, &th)
	if out == "" {
		t.Fatal("RadioRow dark returned empty string")
	}
	if !strings.Contains(out, "◆") {
		t.Error("RadioRow dark: selected item missing ◆ prefix")
	}
	checkGolden(t, "radio-dark-selected", out)
}

// TestRadioRowNoneSelected verifies RadioRow with no selected item uses ◇ for all.
func TestRadioRowNoneSelected(t *testing.T) {
	th := tui.LightTheme()
	opts := []tui.RadioOption{
		{Label: "A", Selected: false},
		{Label: "B", Selected: false},
	}
	out := tui.RadioRow(opts, &th)
	if strings.Contains(out, "◆") {
		t.Error("RadioRow none-selected: unexpected ◆ prefix")
	}
	checkGolden(t, "radio-light-none-selected", out)
}

// TestRadioRowEmpty verifies RadioRow with empty options returns empty string.
func TestRadioRowEmpty(t *testing.T) {
	th := tui.LightTheme()
	out := tui.RadioRow(nil, &th)
	if out != "" {
		t.Errorf("RadioRow empty opts: expected empty string, got %q", out)
	}
}

// TestCheckRowChecked verifies CheckRow with checked and unchecked items.
func TestCheckRowChecked(t *testing.T) {
	th := tui.LightTheme()
	opts := []tui.CheckOption{
		{Label: "Enable feature A", Checked: true},
		{Label: "Enable feature B", Checked: false},
		{Label: "Enable feature C", Checked: true},
	}
	out := tui.CheckRow(opts, &th)
	if out == "" {
		t.Fatal("CheckRow returned empty string")
	}
	checkGolden(t, "check-light-mixed", out)
}

// TestCheckRowDark verifies CheckRow in dark theme.
func TestCheckRowDark(t *testing.T) {
	th := tui.DarkTheme()
	opts := []tui.CheckOption{
		{Label: "Option X", Checked: true},
	}
	out := tui.CheckRow(opts, &th)
	if out == "" {
		t.Fatal("CheckRow dark returned empty string")
	}
	checkGolden(t, "check-dark-single", out)
}

// TestCheckRowEmpty verifies CheckRow with empty options returns empty string.
func TestCheckRowEmpty(t *testing.T) {
	th := tui.LightTheme()
	out := tui.CheckRow(nil, &th)
	if out != "" {
		t.Errorf("CheckRow empty: expected empty string, got %q", out)
	}
}
