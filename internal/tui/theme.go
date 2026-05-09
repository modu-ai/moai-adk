// Package tui provides the MoAI-ADK terminal UI design system v2.
//
// Source: .moai/design/SPEC-V3R3-CLI-TUI-001/source/project/tui.jsx:9-69
// All token values are derived verbatim from TOK.light and TOK.dark in that file.
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
	// Accent is the primary brand accent colour (deep teal).
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
// All values are copied verbatim from tui.jsx:9-38 (TOK.light).
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
		Accent:       "#144a46",
		AccentDeep:   "#0a2825",
		AccentSoft:   "rgba(20,74,70,0.10)",
		AccentSofter: "rgba(20,74,70,0.05)",
		Success:      "#0e7a6c",
		SuccessSoft:  "rgba(14,122,108,0.12)",
		Warning:      "#a86412",
		WarningSoft:  "rgba(168,100,18,0.13)",
		Danger:       "#b1432f",
		DangerSoft:   "rgba(177,67,47,0.12)",
		Info:         "#1f6f72",
		InfoSoft:     "rgba(31,111,114,0.12)",
		Cursor:       "#144a46",
		Selection:    "rgba(20,74,70,0.18)",
		PromptArrow:  "#144a46",
		PromptPath:   "#1f6f72",
		Shadow:       "0 24px 48px -22px rgba(9,17,15,0.22), 0 1px 0 rgba(255,255,255,0.6) inset",
	}
}

// DarkTheme returns the dark-mode design tokens.
// All values are copied verbatim from tui.jsx:39-68 (TOK.dark).
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
		Accent:       "#3eb3a4",
		AccentDeep:   "#22938a",
		AccentSoft:   "rgba(62,179,164,0.16)",
		AccentSofter: "rgba(62,179,164,0.07)",
		Success:      "#3fcfa6",
		SuccessSoft:  "rgba(63,207,166,0.14)",
		Warning:      "#e3a14a",
		WarningSoft:  "rgba(227,161,74,0.14)",
		Danger:       "#ed7d6b",
		DangerSoft:   "rgba(237,125,107,0.15)",
		Info:         "#5cc7c9",
		InfoSoft:     "rgba(92,199,201,0.14)",
		Cursor:       "#3eb3a4",
		Selection:    "rgba(62,179,164,0.25)",
		PromptArrow:  "#3eb3a4",
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
