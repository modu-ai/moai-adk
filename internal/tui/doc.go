// Package tui provides the MoAI-ADK terminal UI design system v2.
//
// This package is the single source of truth for all visual output in the moai
// CLI. Every colour, border, and layout primitive used by the 14 command surfaces
// is rendered through the functions defined here. No other package in internal/
// or pkg/ may hard-code a hex colour constant (REQ-CLI-TUI-013).
//
// # Design Source
//
// Token values are extracted verbatim from:
//
//	.moai/design/SPEC-V3R3-CLI-TUI-001/source/project/tui.jsx:9-69
//
// The 28 tokens cover both light and dark modes. The mapping is 1:1: every
// field in Theme corresponds to a key in TOK.light or TOK.dark in the design
// source. Token values must not be edited here without a corresponding update
// to the design source (SPEC-V3R3-CLI-TUI-001 governance).
//
// # Token Rationale
//
// The 28 tokens are grouped by semantic role:
//
//	Chrome / ChromeBorder / TitleBar  — window frame colours
//	Bg / Panel                        — background surfaces (ivory / ink)
//	Fg / Body / Dim / Faint           — foreground text hierarchy
//	Rule / RuleSoft                   — divider and separator colours
//	Accent / AccentDeep / AccentSoft / AccentSofter — brand deep teal
//	Success / SuccessSoft             — pass / ok states
//	Warning / WarningSoft             — caution states
//	Danger / DangerSoft               — error / fail states
//	Info / InfoSoft                   — informational states
//	Cursor / Selection                — interactive highlight colours
//	PromptArrow / PromptPath          — shell prompt colours
//	Shadow                            — CSS box-shadow (screenshot mode only)
//
// # lipgloss Usage Convention
//
// All ANSI styling is produced through the charmbracelet/lipgloss library.
// The canonical pattern for applying a token colour is:
//
//	th := tui.LightTheme()
//	style := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent))
//
// RGBA token values (e.g. AccentSoft = "rgba(20,74,70,0.10)") are stored as
// strings for documentation and screenshot-mode rendering. They are not passed
// directly to lipgloss.Color in production terminal output because most
// terminals do not support RGBA; only the hex sub-tokens are passed.
//
// # 16-Language Neutrality
//
// All primitives in this package accept label text as plain string arguments.
// There are no language-specific branches (if lang == "ko" etc.) anywhere in
// internal/tui/. Korean character width handling is encapsulated exclusively in
// internal/tui/internal/runeguard.go via the mattn/go-runewidth library.
//
// # NO_COLOR Support
//
// When NO_COLOR is set in the environment, callers should obtain a
// MonochromeTheme() and pass it to all primitives. MonochromeTheme returns a
// Theme where all colour fields are empty strings. lipgloss treats
// lipgloss.Color("") as no colour, producing plain text output with zero
// ANSI escape sequences (REQ-CLI-TUI-009, AC-CLI-TUI-005).
//
// # SPEC
//
// SPEC-V3R3-CLI-TUI-001 M1 — TUI package skeleton and core tokens.
// Acceptance: AC-CLI-TUI-001 (token 1:1), AC-CLI-TUI-014 (#FFF/#000 ban),
// AC-CLI-TUI-016 (global hex sweep), AC-CLI-TUI-017 (emoji sweep).
package tui
