// Package tui provides the MoAI-ADK terminal UI design system.
package tui_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// TestPillVariants runs 6 kinds × 2 themes × 2 solid = 24 cases.
// Each case verifies Pill() returns non-empty output and matches a golden snapshot.
func TestPillVariants(t *testing.T) {
	kinds := []tui.PillKind{
		tui.PillInfo,
		tui.PillOk,
		tui.PillWarn,
		tui.PillErr,
		tui.PillPrimary,
		tui.PillNeutral,
	}

	themes := []struct {
		name  string
		theme tui.Theme
	}{
		{"light", tui.LightTheme()},
		{"dark", tui.DarkTheme()},
	}

	solids := []struct {
		name  string
		solid bool
	}{
		{"outline", false},
		{"solid", true},
	}

	for _, k := range kinds {
		for _, th := range themes {
			for _, s := range solids {
				k, th, s := k, th, s // capture loop vars
				name := "pill-" + string(k) + "-" + th.name + "-" + s.name
				t.Run(name, func(t *testing.T) {
					thCopy := th.theme
					out := tui.Pill(tui.PillOpts{
						Kind:  k,
						Solid: s.solid,
						Label: string(k),
						Theme: &thCopy,
					})
					if out == "" {
						t.Fatalf("Pill(%s, %s, %v) returned empty string", k, th.name, s.solid)
					}
					checkGolden(t, name, out)
				})
			}
		}
	}
}
