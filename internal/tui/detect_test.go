// Package tui provides the MoAI-ADK terminal UI design system.
package tui_test

import (
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// staticEnv is a test double for the Env interface that returns fixed values.
type staticEnv struct {
	noColor   bool
	moaiTheme string
	darkBg    bool
}

func (e staticEnv) NoColor() bool     { return e.noColor }
func (e staticEnv) MoaiTheme() string { return e.moaiTheme }
func (e staticEnv) DetectDark() bool  { return e.darkBg }

// TestResolve verifies Resolve priority chain: NO_COLOR > MOAI_THEME > DetectDark > default-dark.
func TestResolve(t *testing.T) {
	tests := []struct {
		name     string
		env      staticEnv
		wantType string // "monochrome", "light", "dark"
	}{
		{
			name:     "NO_COLOR=1 forces MonochromeTheme regardless of MOAI_THEME",
			env:      staticEnv{noColor: true, moaiTheme: "dark", darkBg: true},
			wantType: "monochrome",
		},
		{
			name:     "MOAI_THEME=invalid falls back to default-dark",
			env:      staticEnv{noColor: false, moaiTheme: "invalid", darkBg: false},
			wantType: "dark",
		},
		{
			name:     "MOAI_THEME=light returns LightTheme",
			env:      staticEnv{noColor: false, moaiTheme: "light", darkBg: true},
			wantType: "light",
		},
		{
			name:     "MOAI_THEME=dark returns DarkTheme",
			env:      staticEnv{noColor: false, moaiTheme: "dark", darkBg: false},
			wantType: "dark",
		},
		{
			name:     "HasDarkBackground=false returns LightTheme when MOAI_THEME unset",
			env:      staticEnv{noColor: false, moaiTheme: "", darkBg: false},
			wantType: "light",
		},
		{
			name:     "HasDarkBackground=true returns DarkTheme when MOAI_THEME unset",
			env:      staticEnv{noColor: false, moaiTheme: "", darkBg: true},
			wantType: "dark",
		},
		{
			name:     "MOAI_THEME=auto defers to DetectDark (dark bg -> dark)",
			env:      staticEnv{noColor: false, moaiTheme: "auto", darkBg: true},
			wantType: "dark",
		},
		{
			name:     "MOAI_THEME=auto defers to DetectDark (light bg -> light)",
			env:      staticEnv{noColor: false, moaiTheme: "auto", darkBg: false},
			wantType: "light",
		},
	}

	light := tui.LightTheme()
	dark := tui.DarkTheme()
	mono := tui.MonochromeTheme()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tui.Resolve(tt.env)

			switch tt.wantType {
			case "monochrome":
				if got != mono {
					t.Errorf("Resolve() = %v, want MonochromeTheme", got)
				}
			case "light":
				if got != light {
					t.Errorf("Resolve() got dark/mono theme, want LightTheme")
				}
			case "dark":
				if got != dark {
					t.Errorf("Resolve() got light/mono theme, want DarkTheme")
				}
			default:
				t.Fatalf("unknown wantType: %s", tt.wantType)
			}
		})
	}
}

// TestThemeResolve is the matrix from AC-CLI-TUI-012 (8 cases).
func TestThemeResolve(t *testing.T) {
	tests := []struct {
		name     string
		env      staticEnv
		wantType string
	}{
		{"1 unset unset light-bg", staticEnv{false, "", false}, "light"},
		{"2 unset unset dark-bg", staticEnv{false, "", true}, "dark"},
		{"3 unset light dark-bg", staticEnv{false, "light", true}, "light"},
		{"4 unset dark light-bg", staticEnv{false, "dark", false}, "dark"},
		{"5 unset auto light-bg", staticEnv{false, "auto", false}, "light"},
		{"6 unset auto dark-bg", staticEnv{false, "auto", true}, "dark"},
		{"7 nocolor dark dark-bg", staticEnv{true, "dark", true}, "monochrome"},
		{"8 nocolor unset light-bg", staticEnv{true, "", false}, "monochrome"},
	}

	light := tui.LightTheme()
	dark := tui.DarkTheme()
	mono := tui.MonochromeTheme()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tui.Resolve(tt.env)
			switch tt.wantType {
			case "monochrome":
				if got != mono {
					t.Errorf("Resolve() != MonochromeTheme")
				}
			case "light":
				if got != light {
					t.Errorf("Resolve() != LightTheme")
				}
			case "dark":
				if got != dark {
					t.Errorf("Resolve() != DarkTheme")
				}
			}
		})
	}
}

// TestOSEnvDetectDark_NonTTYDoesNotHang runs under `go test`, where stdin and
// stdout are pipes, not terminals. Before the non-TTY guard, OSEnv.DetectDark
// issued an OSC 11 background-color query and blocked reading a reply that no
// terminal sends — an indefinite [syscall] hang on Windows CI. The guard must
// short-circuit to the safe-dark default (true) without querying. A hang here
// fails the whole package via the test binary's timeout, which is the exact
// failure this reproduces and guards against.
func TestOSEnvDetectDark_NonTTYDoesNotHang(t *testing.T) {
	done := make(chan bool, 1)
	go func() { done <- tui.OSEnv{}.DetectDark() }()

	select {
	case got := <-done:
		if !got {
			t.Errorf("OSEnv.DetectDark() under non-TTY = %v, want true (safe-dark default)", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("OSEnv.DetectDark() did not return under non-TTY — terminal query blocked (the Windows CI hang)")
	}
}
