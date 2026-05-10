package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Env is the interface used by Resolve to inspect the execution environment.
// A small interface allows unit tests to inject fixed values without mutating
// process-level environment variables (t.Setenv has race issues in parallel tests).
//
// # API Verification (M7)
//
// lipgloss v1.1.0 still exports [lipgloss.HasDarkBackground] (confirmed via
// go/pkg/mod/github.com/charmbracelet/lipgloss@v1.1.0/renderer.go). The function
// queries the default renderer's background-color detection.  We expose it
// through this interface so callers can override it in tests via DetectDark().
//
// colorprofile.Detect() (github.com/charmbracelet/colorprofile v0.4.1) is used
// in profile.go for colour-depth detection; it is NOT used here because
// HasDarkBackground is the correct primitive for light/dark selection.
//
// @MX:ANCHOR: [AUTO] Env interface is the invariant contract for theme resolution;
// @MX:REASON: All callers of Resolve depend on this interface; adding a method
// breaks the staticEnv test double and every future caller simultaneously.
type Env interface {
	// NoColor reports whether the NO_COLOR environment variable is set
	// to any non-empty value (https://no-color.org/).
	NoColor() bool

	// MoaiTheme returns the MOAI_THEME environment variable value.
	// Valid explicit values are "light", "dark", "auto", and "" (unset).
	// Any other value is treated as unset (auto-detect fallback).
	MoaiTheme() string

	// DetectDark reports whether the terminal background is dark.
	// The production implementation delegates to [lipgloss.HasDarkBackground].
	DetectDark() bool
}

// Resolve returns the appropriate Theme for the given environment using the
// following priority chain (highest to lowest):
//
//  1. NO_COLOR set to any non-empty string → MonochromeTheme
//  2. MOAI_THEME="light" → LightTheme
//  3. MOAI_THEME="dark"  → DarkTheme
//  4. MOAI_THEME="auto" or unset/invalid → env.DetectDark()
//  5. env.DetectDark()==false → LightTheme
//  6. default → DarkTheme (safe default, REQ-CLI-TUI-010)
//
// See AC-CLI-TUI-012 for the full 8-case acceptance matrix.
//
// @MX:NOTE: [AUTO] Priority chain: NO_COLOR > MOAI_THEME(light/dark) > DetectDark > dark-default.
// MOAI_THEME="auto" and invalid/empty values delegate directly to DetectDark.
func Resolve(env Env) Theme {
	if env.NoColor() {
		return MonochromeTheme()
	}

	switch env.MoaiTheme() {
	case "light":
		return LightTheme()
	case "dark":
		return DarkTheme()
	case "auto", "":
		// Defer to terminal background detection
	default:
		// Invalid value: use safe dark default without querying the terminal
		// (REQ-CLI-TUI-010, plan.md §9.3 "safe default").
		return DarkTheme()
	}

	if env.DetectDark() {
		return DarkTheme()
	}
	return LightTheme()
}

// OSEnv is the production implementation of Env that reads from the process
// environment and uses [lipgloss.HasDarkBackground] for terminal background detection.
//
// @MX:NOTE: [AUTO] OSEnv is the production Env; used by CLI entry points to
// obtain the active theme without injecting test-specific overrides.
type OSEnv struct{}

// NoColor reports whether NO_COLOR is set in the process environment.
func (OSEnv) NoColor() bool {
	return isEnvSet("NO_COLOR")
}

// MoaiTheme returns the MOAI_THEME environment variable value.
func (OSEnv) MoaiTheme() string {
	return envLookup("MOAI_THEME")
}

// DetectDark delegates to [lipgloss.HasDarkBackground].
func (OSEnv) DetectDark() bool {
	return lipgloss.HasDarkBackground()
}

// ResolveOS is a convenience wrapper that calls Resolve with the production
// OSEnv. Most CLI commands should call this to obtain the active theme.
func ResolveOS() Theme {
	return Resolve(OSEnv{})
}

// isEnvSet reports whether the named environment variable is set to any
// non-empty value (following the NO_COLOR standard: any non-empty string is set).
func isEnvSet(name string) bool {
	v, ok := os.LookupEnv(name)
	return ok && v != ""
}

// envLookup returns the value of the named environment variable, or "" if unset.
func envLookup(name string) string {
	return os.Getenv(name)
}
