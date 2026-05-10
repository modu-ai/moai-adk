// Package tui provides the MoAI-ADK terminal UI design system.
package tui_test

import (
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// profileEnv is a test double that combines staticEnv with a fixed ColorProfile.
type profileEnv struct {
	staticEnv
	profile colorprofile.Profile
}

func (e profileEnv) ColorProfile() colorprofile.Profile { return e.profile }

// TestProfileDegradation verifies that Profile() returns the colour profile and
// that the theme is degraded to MonochromeTheme when the profile is NoTTY or ASCII.
// Matrix: 4 ColorProfile values × 2 theme selections (light / dark).
func TestProfileDegradation(t *testing.T) {
	tests := []struct {
		name     string
		env      profileEnv
		wantMono bool
	}{
		{
			name:     "TrueColor + dark theme",
			env:      profileEnv{staticEnv{false, "dark", true}, colorprofile.TrueColor},
			wantMono: false,
		},
		{
			name:     "TrueColor + light theme",
			env:      profileEnv{staticEnv{false, "light", false}, colorprofile.TrueColor},
			wantMono: false,
		},
		{
			name:     "ANSI256 + dark theme",
			env:      profileEnv{staticEnv{false, "dark", true}, colorprofile.ANSI256},
			wantMono: false,
		},
		{
			name:     "ANSI256 + light theme",
			env:      profileEnv{staticEnv{false, "light", false}, colorprofile.ANSI256},
			wantMono: false,
		},
		{
			name:     "ANSI + dark theme",
			env:      profileEnv{staticEnv{false, "dark", true}, colorprofile.ANSI},
			wantMono: false,
		},
		{
			name:     "ANSI + light theme",
			env:      profileEnv{staticEnv{false, "light", false}, colorprofile.ANSI},
			wantMono: false,
		},
		{
			name:     "NoTTY + dark theme → MonochromeTheme",
			env:      profileEnv{staticEnv{false, "dark", true}, colorprofile.NoTTY},
			wantMono: true,
		},
		{
			name:     "ASCII + light theme → MonochromeTheme",
			env:      profileEnv{staticEnv{false, "light", false}, colorprofile.ASCII},
			wantMono: true,
		},
	}

	mono := tui.MonochromeTheme()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tui.Profile(tt.env)
			if tt.wantMono {
				if got != mono {
					t.Errorf("Profile() should return MonochromeTheme for %s, got non-mono", tt.name)
				}
			} else {
				if got == mono {
					t.Errorf("Profile() returned MonochromeTheme for %s, want coloured theme", tt.name)
				}
			}
		})
	}
}

// TestProfileGetProfile verifies the colour profile accessor.
func TestProfileGetProfile(t *testing.T) {
	env := profileEnv{staticEnv{false, "dark", true}, colorprofile.TrueColor}
	p := tui.GetColorProfile(env)
	if p != colorprofile.TrueColor {
		t.Errorf("GetColorProfile() = %v, want TrueColor", p)
	}
}
