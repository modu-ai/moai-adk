package tui

import (
	"bytes"
	"os"
	"sync"

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

// outputProfile lazily detects the colour profile of the process stdout.
//
// Lip Gloss v2 removed the v1 global default renderer: v2 Style.Render always
// emits full-fidelity (24-bit) ANSI and delegates downsampling/stripping to a
// colorprofile-aware writer. To preserve the v1 observable behaviour — styles
// were automatically degraded (or fully stripped on non-TTY output such as
// pipes, CI logs, and `go test`) by the default renderer bound to os.Stdout —
// this package detects the profile once from os.Stdout + environment and
// re-encodes every rendered string through it (see downsample).
//
// colorprofile.Detect respects NO_COLOR, CLICOLOR, and CLICOLOR_FORCE, and
// performs no terminal query (isatty + env inspection only), matching the v1
// renderer's detection cost.
var (
	outputProfileOnce sync.Once
	outputProfileVal  colorprofile.Profile
)

func outputProfile() colorprofile.Profile {
	outputProfileOnce.Do(func() {
		outputProfileVal = colorprofile.Detect(os.Stdout, os.Environ())
	})
	return outputProfileVal
}

// downsample re-encodes a styled string for the detected terminal colour
// profile, reproducing the lipgloss v1 default-renderer degradation semantics
// under lipgloss v2:
//
//   - TrueColor: passthrough (no re-encoding)
//   - ANSI256 / ANSI / ASCII: colours downsampled to the supported palette
//   - NoTTY (pipes, CI, go test): all ANSI sequences stripped, plain text
//
// The helper is unexported and string-in/string-out, so no lipgloss (or
// colorprofile) type leaks across the package's public string-token boundary
// (design decision D-1: the public contract is plain string tokens).
func downsample(s string) string {
	p := outputProfile()
	if p == colorprofile.TrueColor {
		return s
	}
	var buf bytes.Buffer
	w := colorprofile.Writer{Forward: &buf, Profile: p}
	_, _ = w.WriteString(s)
	return buf.String()
}
