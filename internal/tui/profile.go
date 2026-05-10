package tui

import (
	"os"

	"github.com/charmbracelet/colorprofile"
)

// ProfileEnv extends Env with colour-depth detection.
// It is used by Profile to combine light/dark selection with colour-capability
// degradation.
//
// @MX:ANCHOR: [AUTO] ProfileEnv extends Env for colour-depth detection; high fan_in expected
// @MX:REASON: Profile() and GetColorProfile() both take this interface;
// any future surface requiring colour-depth control must implement it.
type ProfileEnv interface {
	Env
	// ColorProfile returns the detected terminal colour profile.
	// In production, use DetectedProfileEnv which calls colorprofile.Env.
	ColorProfile() colorprofile.Profile
}

// Profile returns the active Theme, taking both the light/dark selection
// (via Resolve) and the terminal colour profile into account.
//
// Degradation rules:
//   - colorprofile.NoTTY  → MonochromeTheme (no terminal, no ANSI)
//   - colorprofile.ASCII  → MonochromeTheme (ASCII-only, no colour)
//   - colorprofile.ANSI   → Resolve(env) with ANSI palette
//   - colorprofile.ANSI256 → Resolve(env) (256-colour)
//   - colorprofile.TrueColor → Resolve(env) (24-bit colour, preferred)
//
// @MX:NOTE: [AUTO] Profile() composes Resolve() with colour-depth degradation;
// callers should prefer Profile() over Resolve() when colour support is uncertain.
func Profile(env ProfileEnv) Theme {
	p := env.ColorProfile()
	if p <= colorprofile.ASCII {
		// No colour support: override theme to monochrome.
		return MonochromeTheme()
	}
	return Resolve(env)
}

// GetColorProfile returns the colour profile reported by env.
// It is a thin accessor exposed for testing and introspection.
func GetColorProfile(env ProfileEnv) colorprofile.Profile {
	return env.ColorProfile()
}

// DetectedProfileEnv is the production implementation of ProfileEnv.
// It reads NO_COLOR and MOAI_THEME from the process environment,
// uses [lipgloss.HasDarkBackground] for dark-background detection,
// and calls [colorprofile.Env] to detect the terminal colour capability.
//
// @MX:NOTE: [AUTO] DetectedProfileEnv is the production ProfileEnv used by CLI
// entry points; it shares the same env-probe helpers as OSEnv in detect.go.
type DetectedProfileEnv struct{}

// NoColor reports whether NO_COLOR is set.
func (DetectedProfileEnv) NoColor() bool { return isEnvSet("NO_COLOR") }

// MoaiTheme returns the MOAI_THEME environment variable value.
func (DetectedProfileEnv) MoaiTheme() string { return envLookup("MOAI_THEME") }

// DetectDark delegates to [lipgloss.HasDarkBackground].
func (DetectedProfileEnv) DetectDark() bool { return OSEnv{}.DetectDark() }

// ColorProfile detects the terminal colour capability using [colorprofile.Env].
// It reads the current process environment variables (TERM, COLORTERM, NO_COLOR,
// CLICOLOR, CLICOLOR_FORCE) to determine the best supported profile.
func (DetectedProfileEnv) ColorProfile() colorprofile.Profile {
	return colorprofile.Env(os.Environ())
}

// ProfileOS returns the active Theme using the production DetectedProfileEnv.
// Most CLI commands should call this instead of ResolveOS for correct degradation.
func ProfileOS() Theme {
	return Profile(DetectedProfileEnv{})
}
