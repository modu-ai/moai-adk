package cli

import (
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// TestMoAIHuhTheme_ReturnsNonNil verifies the factory yields a usable theme.
func TestMoAIHuhTheme_ReturnsNonNil(t *testing.T) {
	if moaiHuhTheme() == nil {
		t.Fatal("expected non-nil huh theme")
	}
}

// TestMoAIHuhTheme_ResolvesDarkAxis reproduces the dark-terminal contrast defect
// for the non-wizard huh (v1) surfaces: huh's default theme leans on the
// terminal's async background reply, so before that reply lands (or on terminals
// that never answer) the near-black body token renders over a dark background,
// leaving the prompt descriptions unreadable. The factory must resolve the axis
// itself via tui.IsDarkOS (overridable through huhThemeIsDark for tests) and
// paint the tui Body token that matches the actual background.
//
// The foreground is inspected by type-asserting to lipgloss.Color and comparing
// the token string (deterministic, independent of the renderer's colour
// profile). Expected values come from the tui tokens, never a hex literal.
func TestMoAIHuhTheme_ResolvesDarkAxis(t *testing.T) {
	tests := []struct {
		name     string
		dark     bool
		wantBody string
	}{
		{"resolver says dark", true, tui.DarkTheme().Body},
		{"resolver says light", false, tui.LightTheme().Body},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := huhThemeIsDark
			t.Cleanup(func() { huhThemeIsDark = orig })
			huhThemeIsDark = func() bool { return tt.dark }

			got := moaiHuhTheme().Focused.Description.GetForeground()
			col, ok := got.(lipgloss.Color)
			if !ok {
				t.Fatalf("Focused.Description foreground: expected lipgloss.Color, got %T", got)
			}
			if string(col) != tt.wantBody {
				t.Errorf("Focused.Description foreground = %q, want tui Body token %q",
					string(col), tt.wantBody)
			}
		})
	}
}
