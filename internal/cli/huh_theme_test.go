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

// TestMoAIHuhTheme_ClosesWizardFactoryDivergence pins the fields the huh v2
// wizard factory set and this v1 factory previously left at the huh defaults.
// The two factories stay separate — the Theme types differ across the library
// major boundary — but their token-to-role assignment is now the same, so a
// regression that drops one of these closures is caught here rather than
// discovered visually.
//
// Expected values are read from the tui tokens, never written as hex literals.
func TestMoAIHuhTheme_ClosesWizardFactoryDivergence(t *testing.T) {
	for _, dark := range []bool{false, true} {
		th := tui.LightTheme()
		if dark {
			th = tui.DarkTheme()
		}

		styles := moaiHuhStyles(dark)
		f := styles.Focused

		t.Run(map[bool]string{false: "light", true: "dark"}[dark], func(t *testing.T) {
			assertFg := func(name string, got lipgloss.TerminalColor, want string) {
				t.Helper()
				col, ok := got.(lipgloss.Color)
				if !ok {
					t.Fatalf("%s foreground: expected lipgloss.Color, got %T", name, got)
				}
				if string(col) != want {
					t.Errorf("%s foreground = %q, want tui token %q", name, string(col), want)
				}
			}

			// Base / Card border — the principal gap the audit surfaced.
			assertFg("Focused.Base border", f.Base.GetBorderTopForeground(), th.Rule)
			assertFg("Focused.Card border", f.Card.GetBorderTopForeground(), th.Rule)

			// Prefix and selector strings, matching the wizard factory verbatim.
			if got := f.SelectSelector.Value(); got != "▸ " {
				t.Errorf("Focused.SelectSelector = %q, want %q", got, "▸ ")
			}
			if got := f.SelectedPrefix.Value(); got != "◆ " {
				t.Errorf("Focused.SelectedPrefix = %q, want %q", got, "◆ ")
			}
			if got := f.UnselectedPrefix.Value(); got != "◇ " {
				t.Errorf("Focused.UnselectedPrefix = %q, want %q", got, "◇ ")
			}
			assertFg("Focused.SelectedPrefix", f.SelectedPrefix.GetForeground(), th.Success)
			assertFg("Focused.UnselectedPrefix", f.UnselectedPrefix.GetForeground(), th.Dim)

			// Next mirrors FocusedButton, as it does in the wizard factory.
			if f.Next.GetBackground() != f.FocusedButton.GetBackground() {
				t.Errorf("Focused.Next background does not mirror FocusedButton")
			}

			// NoteTitle gained the bottom margin the wizard factory applies.
			if got := f.NoteTitle.GetMarginBottom(); got != 1 {
				t.Errorf("Focused.NoteTitle margin-bottom = %d, want 1", got)
			}

			// The blurred variant keeps its hidden border, so focus stays legible.
			if styles.Blurred.Base.GetBorderStyle() != lipgloss.HiddenBorder() {
				t.Errorf("Blurred.Base lost its hidden border")
			}
		})
	}
}
