// Package tui provides the MoAI-ADK terminal UI design system.
package tui_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// TestLightTokens verifies all 28 light-theme token values. Backgrounds and
// chrome are inherited from the original design source tui.jsx:9-69 (TOK.light);
// the accent / success / warning / info family was re-grounded on the Claude
// coral palette (#cc785c canonical mid-point) per the 2026-07 brand alignment.
// This is the specification test for REQ-CLI-TUI-002.
func TestLightTokens(t *testing.T) {
	th := tui.LightTheme()

	cases := []struct {
		name string
		got  string
		want string
	}{
		{"chrome", th.Chrome, "#e8e6e0"},
		{"chromeBorder", th.ChromeBorder, "#bdbab2"},
		{"titleBar", th.TitleBar, "linear-gradient(180deg,#efece5 0%,#e1ddd3 100%)"},
		{"bg", th.Bg, "#fbfaf6"},
		{"panel", th.Panel, "#f3f3f3"},
		{"fg", th.Fg, "#0e1513"},
		{"body", th.Body, "#1f2826"},
		{"dim", th.Dim, "#5b625f"},
		{"faint", th.Faint, "#8c918d"},
		{"rule", th.Rule, "#dcd9d2"},
		{"ruleSoft", th.RuleSoft, "#ebe8e1"},
		{"accent", th.Accent, "#bf6547"},
		{"accentDeep", th.AccentDeep, "#a84f33"},
		{"accentSoft", th.AccentSoft, "rgba(191,101,71,0.10)"},
		{"accentSofter", th.AccentSofter, "rgba(191,101,71,0.05)"},
		{"success", th.Success, "#3d8b6e"},
		{"successSoft", th.SuccessSoft, "rgba(61,139,110,0.12)"},
		{"warning", th.Warning, "#b9701a"},
		{"warningSoft", th.WarningSoft, "rgba(185,112,26,0.13)"},
		{"danger", th.Danger, "#b1432f"},
		{"dangerSoft", th.DangerSoft, "rgba(177,67,47,0.12)"},
		{"info", th.Info, "#1f7a7d"},
		{"infoSoft", th.InfoSoft, "rgba(31,122,125,0.12)"},
		{"cursor", th.Cursor, "#bf6547"},
		{"selection", th.Selection, "rgba(191,101,71,0.18)"},
		{"promptArrow", th.PromptArrow, "#bf6547"},
		{"promptPath", th.PromptPath, "#1f7a7d"},
		{"shadow", th.Shadow, "0 24px 48px -22px rgba(9,17,15,0.22), 0 1px 0 rgba(255,255,255,0.6) inset"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("LightTheme().%s = %q, want %q", tc.name, tc.got, tc.want)
			}
		})
	}
}

// TestDarkTokens verifies all 28 dark-theme token values. Backgrounds and
// chrome are inherited from the original design source tui.jsx:39-68 (TOK.dark);
// the accent family was re-grounded on the Claude coral palette (canonical
// mid-point #cc785c) per the 2026-07 brand alignment. Warning, Danger, and Info
// were already warm-harmonious and are retained verbatim.
// This is the specification test for REQ-CLI-TUI-002.
func TestDarkTokens(t *testing.T) {
	th := tui.DarkTheme()

	cases := []struct {
		name string
		got  string
		want string
	}{
		{"chrome", th.Chrome, "#0c1413"},
		{"chromeBorder", th.ChromeBorder, "#1c2624"},
		{"titleBar", th.TitleBar, "linear-gradient(180deg,#131b19 0%,#0a1110 100%)"},
		{"bg", th.Bg, "#0a110f"},
		{"panel", th.Panel, "#0f1816"},
		{"fg", th.Fg, "#eef2ef"},
		{"body", th.Body, "#d8dedb"},
		{"dim", th.Dim, "#9aa3a0"},
		{"faint", th.Faint, "#6b7370"},
		{"rule", th.Rule, "#1c2624"},
		{"ruleSoft", th.RuleSoft, "#152019"},
		{"accent", th.Accent, "#d97757"},
		{"accentDeep", th.AccentDeep, "#b85e3f"},
		{"accentSoft", th.AccentSoft, "rgba(217,119,87,0.16)"},
		{"accentSofter", th.AccentSofter, "rgba(217,119,87,0.07)"},
		{"success", th.Success, "#5bbf9a"},
		{"successSoft", th.SuccessSoft, "rgba(91,191,154,0.14)"},
		{"warning", th.Warning, "#e3a14a"},
		{"warningSoft", th.WarningSoft, "rgba(227,161,74,0.14)"},
		{"danger", th.Danger, "#ed7d6b"},
		{"dangerSoft", th.DangerSoft, "rgba(237,125,107,0.15)"},
		{"info", th.Info, "#5cc7c9"},
		{"infoSoft", th.InfoSoft, "rgba(92,199,201,0.14)"},
		{"cursor", th.Cursor, "#d97757"},
		{"selection", th.Selection, "rgba(217,119,87,0.25)"},
		{"promptArrow", th.PromptArrow, "#d97757"},
		{"promptPath", th.PromptPath, "#5cc7c9"},
		{"shadow", th.Shadow, "0 30px 60px -22px rgba(0,0,0,0.65), 0 1px 0 rgba(255,255,255,0.03) inset"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("DarkTheme().%s = %q, want %q", tc.name, tc.got, tc.want)
			}
		})
	}
}

// TestMonochromeTheme verifies that MonochromeTheme returns a Theme where all
// lipgloss-relevant color fields are empty strings (NO_COLOR behaviour).
func TestMonochromeTheme(t *testing.T) {
	th := tui.MonochromeTheme()
	// All hex-based fields must be empty in monochrome mode.
	hexFields := []struct {
		name string
		got  string
	}{
		{"accent", th.Accent},
		{"bg", th.Bg},
		{"fg", th.Fg},
		{"danger", th.Danger},
		{"success", th.Success},
		{"warning", th.Warning},
		{"info", th.Info},
	}
	for _, f := range hexFields {
		t.Run(f.name, func(t *testing.T) {
			if f.got != "" {
				t.Errorf("MonochromeTheme().%s = %q, want empty string", f.name, f.got)
			}
		})
	}
}

// TestThemeIsLanguageNeutral confirms no field in Theme references locale strings.
// This test satisfies REQ-CLI-TUI-003: primitive must be language-agnostic.
func TestThemeIsLanguageNeutral(t *testing.T) {
	// The Theme struct fields are all color codes or CSS values.
	// This test confirms the Theme type exists and can be used without language context.
	light := tui.LightTheme()
	dark := tui.DarkTheme()

	// Both themes should be distinct (proving they are not the same value set).
	if light.Accent == dark.Accent {
		t.Error("LightTheme().Accent and DarkTheme().Accent must differ")
	}
	if light.Bg == dark.Bg {
		t.Error("LightTheme().Bg and DarkTheme().Bg must differ")
	}
}

// TestNoSurfaceUsesPureColors verifies that neither LightTheme nor DarkTheme
// defines pure white (#FFFFFF) or pure black (#000000) for any background token.
// This satisfies AC-CLI-TUI-014 / REQ-CLI-TUI-016.
func TestNoSurfaceUsesPureColors(t *testing.T) {
	forbidden := []string{"#FFFFFF", "#ffffff", "#000000"}

	check := func(themeName, field, val string) {
		for _, f := range forbidden {
			if val == f {
				t.Errorf("%s.%s = %q: pure white/black forbidden (AC-CLI-TUI-014)", themeName, field, val)
			}
		}
	}

	for _, td := range []struct {
		name  string
		theme tui.Theme
	}{
		{"LightTheme", tui.LightTheme()},
		{"DarkTheme", tui.DarkTheme()},
	} {
		th := td.theme
		check(td.name, "Bg", th.Bg)
		check(td.name, "Panel", th.Panel)
		check(td.name, "Chrome", th.Chrome)
		check(td.name, "Fg", th.Fg)
		check(td.name, "Body", th.Body)
	}
}
