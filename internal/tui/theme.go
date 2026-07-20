// Package tui provides the MoAI-ADK terminal UI design system v2.
//
// Source: original design tokens from SPEC-V3R3-CLI-TUI-001; palette migrated
// to Claude coral (#cc785c) per brand alignment 2026-07. The docs-site "Claude
// Warm Editorial" system already uses this coral as its primary. Backgrounds
// (ivory/ink) and chrome tokens are brand-neutral and were kept verbatim; only
// the accent / success / warning / danger / info family was re-grounded on the
// coral palette. Canonical coral mid-point: #cc785c.
//
// # Design Tokens
//
// The 28 tokens in each Theme are the single source of truth for all colour
// decisions in the MoAI-ADK terminal output. No file outside internal/tui/ may
// hard-code a hex colour constant (REQ-CLI-TUI-013).
//
// Token names follow the camelCase convention from the design source, with the
// first letter capitalised for Go export. RGBA and CSS gradient values are stored
// as plain strings; lipgloss.Color interprets hex sub-strings automatically.
//
// # Usage
//
//	th := tui.LightTheme()
//	style := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent))
//
// # MonochromeTheme
//
// MonochromeTheme returns a Theme with all colour-bearing fields set to empty
// string. It is used when NO_COLOR is set (REQ-CLI-TUI-009). lipgloss treats an
// empty string in lipgloss.Color("") as no-colour, producing plain text output.
package tui

// Theme holds the 28 design tokens for one display mode (light or dark).
// Every field maps 1:1 to the corresponding key in the design source
// tui.jsx:9-69 (TOK.light / TOK.dark). Fields containing CSS gradients or
// RGBA values are stored as plain strings for documentation purposes; only
// the hex fields are passed directly to lipgloss.Color.
type Theme struct {
	// Chrome is the window chrome / titlebar background colour.
	Chrome string
	// ChromeBorder is the colour for the window chrome border.
	ChromeBorder string
	// TitleBar is the CSS gradient for the macOS-style title bar (screenshot
	// mode only; not used in live terminal rendering).
	TitleBar string
	// Bg is the main terminal background colour (ivory in light, ink in dark).
	Bg string
	// Panel is the secondary panel / sidebar background colour.
	Panel string
	// Fg is the primary foreground / text colour.
	Fg string
	// Body is the secondary body text colour.
	Body string
	// Dim is the muted / helper text colour.
	Dim string
	// Faint is the placeholder / caption colour.
	Faint string
	// Rule is the standard divider / border colour.
	Rule string
	// RuleSoft is the subtle divider colour.
	RuleSoft string
	// Accent is the primary brand accent colour (Claude coral).
	Accent string
	// AccentDeep is a deeper variant of the accent colour.
	AccentDeep string
	// AccentSoft is a semi-transparent accent overlay (RGBA string).
	AccentSoft string
	// AccentSofter is an even more transparent accent overlay (RGBA string).
	AccentSofter string
	// Success is the success / pass colour.
	Success string
	// SuccessSoft is a semi-transparent success overlay (RGBA string).
	SuccessSoft string
	// Warning is the warning / caution colour.
	Warning string
	// WarningSoft is a semi-transparent warning overlay (RGBA string).
	WarningSoft string
	// Danger is the error / danger colour.
	Danger string
	// DangerSoft is a semi-transparent danger overlay (RGBA string).
	DangerSoft string
	// Info is the informational colour.
	Info string
	// InfoSoft is a semi-transparent info overlay (RGBA string).
	InfoSoft string
	// Cursor is the cursor blink colour (same as Accent in most modes).
	Cursor string
	// Selection is the text selection highlight colour (RGBA string).
	Selection string
	// PromptArrow is the colour for the prompt chevron symbol.
	PromptArrow string
	// PromptPath is the colour for the prompt working-directory path.
	PromptPath string
	// Shadow is the CSS box-shadow definition (screenshot mode only).
	Shadow string
}

// LightTheme returns the light-mode design tokens.
// Backgrounds and chrome are inherited verbatim from the original tui.jsx:9-38
// (TOK.light). The accent / success / warning / danger / info family is
// re-grounded on the Claude coral palette (canonical mid-point #cc785c); the
// light-mode accent is a slightly deeper coral (#bf6547) for contrast on ivory.
//
// @MX:ANCHOR: [AUTO] LightTheme is the canonical light token source; fan_in expected >=3
// @MX:REASON: Box, Pill, and all future M2 primitives call this; single source per REQ-CLI-TUI-001
func LightTheme() Theme {
	return Theme{
		Chrome:       "#e8e6e0",
		ChromeBorder: "#bdbab2",
		TitleBar:     "linear-gradient(180deg,#efece5 0%,#e1ddd3 100%)",
		Bg:           "#fbfaf6",
		Panel:        "#f3f3f3",
		Fg:           "#0e1513",
		Body:         "#1f2826",
		Dim:          "#5b625f",
		Faint:        "#8c918d",
		Rule:         "#dcd9d2",
		RuleSoft:     "#ebe8e1",
		Accent:       "#bf6547",
		AccentDeep:   "#a84f33",
		AccentSoft:   "rgba(191,101,71,0.10)",
		AccentSofter: "rgba(191,101,71,0.05)",
		Success:      "#3d8b6e",
		SuccessSoft:  "rgba(61,139,110,0.12)",
		Warning:      "#b9701a",
		WarningSoft:  "rgba(185,112,26,0.13)",
		Danger:       "#b1432f",
		DangerSoft:   "rgba(177,67,47,0.12)",
		Info:         "#1f7a7d",
		InfoSoft:     "rgba(31,122,125,0.12)",
		Cursor:       "#bf6547",
		Selection:    "rgba(191,101,71,0.18)",
		PromptArrow:  "#bf6547",
		PromptPath:   "#1f7a7d",
		Shadow:       "0 24px 48px -22px rgba(9,17,15,0.22), 0 1px 0 rgba(255,255,255,0.6) inset",
	}
}

// DarkTheme returns the dark-mode design tokens.
// Backgrounds and chrome are inherited verbatim from the original tui.jsx:39-68
// (TOK.dark). The accent family is re-grounded on the Claude coral palette; the
// dark-mode accent is a brighter coral (#d97757) for legibility on ink. Warning,
// Danger, and Info were already warm-harmonious and are retained verbatim.
//
// @MX:ANCHOR: [AUTO] DarkTheme is the canonical dark token source; fan_in expected >=3
// @MX:REASON: Box, Pill, and all future M2 primitives call this; single source per REQ-CLI-TUI-001
func DarkTheme() Theme {
	return Theme{
		Chrome:       "#0c1413",
		ChromeBorder: "#1c2624",
		TitleBar:     "linear-gradient(180deg,#131b19 0%,#0a1110 100%)",
		Bg:           "#0a110f",
		Panel:        "#0f1816",
		Fg:           "#eef2ef",
		Body:         "#d8dedb",
		Dim:          "#9aa3a0",
		Faint:        "#6b7370",
		Rule:         "#1c2624",
		RuleSoft:     "#152019",
		Accent:       "#d97757",
		AccentDeep:   "#b85e3f",
		AccentSoft:   "rgba(217,119,87,0.16)",
		AccentSofter: "rgba(217,119,87,0.07)",
		Success:      "#5bbf9a",
		SuccessSoft:  "rgba(91,191,154,0.14)",
		Warning:      "#e3a14a",
		WarningSoft:  "rgba(227,161,74,0.14)",
		Danger:       "#ed7d6b",
		DangerSoft:   "rgba(237,125,107,0.15)",
		Info:         "#5cc7c9",
		InfoSoft:     "rgba(92,199,201,0.14)",
		Cursor:       "#d97757",
		Selection:    "rgba(217,119,87,0.25)",
		PromptArrow:  "#d97757",
		PromptPath:   "#5cc7c9",
		Shadow:       "0 30px 60px -22px rgba(0,0,0,0.65), 0 1px 0 rgba(255,255,255,0.03) inset",
	}
}

// MonochromeTheme returns a Theme with all colour fields empty.
// It is used when NO_COLOR is set (REQ-CLI-TUI-009, AC-CLI-TUI-005).
// lipgloss.Color("") produces no ANSI colour escape, rendering plain text.
// Non-colour metadata fields (TitleBar, Shadow) retain empty strings too.
func MonochromeTheme() Theme {
	return Theme{}
}
