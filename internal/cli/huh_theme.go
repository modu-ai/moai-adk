package cli

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// huhThemeIsDark resolves the light/dark axis for the MoAI-branded theme applied
// to the non-wizard huh surfaces (the init/update profile-setup confirms and the
// profile_setup language/main forms). It is a package-level var so tests can
// force either axis without mutating the process environment, mirroring
// wizard.wizardIsDark.
var huhThemeIsDark = tui.IsDarkOS

// moaiHuhTheme returns the MoAI-branded github.com/charmbracelet/huh (v1) theme
// with the light/dark axis resolved once via tui.IsDarkOS (NO_COLOR >
// MOAI_THEME > terminal detection).
//
// It mirrors the wizard's newMoAIWizardTheme fix (internal/cli/wizard), but is a
// separate factory because the four non-wizard surfaces still run on huh v1,
// whose Theme is a struct type distinct from the wizard's charm.land/huh/v2
// Theme interface — the v2 factory cannot cross the version boundary. The
// readability defect is the same on both APIs: huh's default theme leans on the
// terminal's async background reply for its adaptive colours, so before that
// reply lands (or on terminals that never answer) the near-black body token
// renders over a dark background, leaving the prompt descriptions unreadable.
// Resolving the axis here paints the tui token that matches the actual
// background.
func moaiHuhTheme() *huh.Theme {
	return moaiHuhStyles(huhThemeIsDark())
}

// moaiHuhStyles builds the MoAI-branded v1 huh theme for the resolved
// background. Colour values are derived from internal/tui LightTheme / DarkTheme
// tokens (no hex literals outside internal/tui/), mirroring the wizard's
// moaiWizardStyles role mapping: Title=Accent (Claude coral), Description=Body,
// Error=Danger, Muted=Dim, Secondary=Info.
//
// # Divergence audit against the huh v2 sibling factory
//
// The two factories stay separate (the v1 huh.Theme struct and the v2
// charm.land/huh Theme interface do not cross the library-major boundary), but
// they now apply the SAME token-to-role assignment. The fields the v2 factory
// set and this one did not — Base/Card borders, SelectedPrefix,
// UnselectedPrefix, the SelectSelector string, the NoteTitle bottom margin, and
// Next — are all present on the v1 FieldStyles type, so every gap the audit
// surfaced is CLOSED below rather than reasoned away.
//
// Border token: both factories paint the field card with Theme.Rule. The
// preview TUI reserves Theme.ChromeBorder for the bubbles table's own chrome,
// which has no huh analogue; the preview's card primitive (tui.Box) likewise
// draws its non-accent border from Rule. Rule is therefore the preview-aligned
// choice for a form card, not a leftover.
//
// The prefix and selector strings ("◆ ", "◇ ", "▸ ") match the v2 factory
// verbatim so a user moving between the wizard and the non-wizard forms sees
// one vocabulary.
func moaiHuhStyles(isDark bool) *huh.Theme {
	t := huh.ThemeBase()

	th := tui.LightTheme()
	buttonFg := th.Bg // white-ish text on filled button (light background)
	if isDark {
		th = tui.DarkTheme()
		buttonFg = th.Fg
	}

	fg := func(token string) lipgloss.Color { return lipgloss.Color(token) }

	brand := func(fs huh.FieldStyles) huh.FieldStyles {
		// Base carries the field card's border. ThemeBase already gives the
		// blurred variant a hidden border, so recolouring here preserves the
		// focused/blurred distinction the v2 factory also relies on.
		fs.Base = fs.Base.BorderForeground(fg(th.Rule))
		fs.Card = fs.Base
		fs.Title = fs.Title.Foreground(fg(th.Accent)).Bold(true)
		fs.NoteTitle = fs.NoteTitle.Foreground(fg(th.Accent)).Bold(true).MarginBottom(1)
		fs.Description = fs.Description.Foreground(fg(th.Body))
		fs.ErrorIndicator = fs.ErrorIndicator.Foreground(fg(th.Danger))
		fs.ErrorMessage = fs.ErrorMessage.Foreground(fg(th.Danger))
		fs.SelectSelector = fs.SelectSelector.Foreground(fg(th.Accent)).SetString("▸ ")
		fs.NextIndicator = fs.NextIndicator.Foreground(fg(th.Accent))
		fs.PrevIndicator = fs.PrevIndicator.Foreground(fg(th.Accent))
		fs.Option = fs.Option.Foreground(fg(th.Fg))
		fs.MultiSelectSelector = fs.MultiSelectSelector.Foreground(fg(th.Accent))
		fs.SelectedOption = fs.SelectedOption.Foreground(fg(th.Success))
		fs.SelectedPrefix = lipgloss.NewStyle().Foreground(fg(th.Success)).SetString("◆ ")
		fs.UnselectedOption = fs.UnselectedOption.Foreground(fg(th.Fg))
		fs.UnselectedPrefix = lipgloss.NewStyle().Foreground(fg(th.Dim)).SetString("◇ ")
		fs.TextInput.Cursor = fs.TextInput.Cursor.Foreground(fg(th.Accent))
		fs.TextInput.Placeholder = fs.TextInput.Placeholder.Foreground(fg(th.Dim))
		fs.TextInput.Prompt = fs.TextInput.Prompt.Foreground(fg(th.Info))
		fs.FocusedButton = fs.FocusedButton.
			Foreground(fg(buttonFg)).
			Background(fg(th.Accent))
		fs.BlurredButton = fs.BlurredButton.
			Foreground(fg(th.Fg)).
			Background(fg(th.Panel))
		fs.Next = fs.FocusedButton
		return fs
	}

	t.Focused = brand(t.Focused)
	t.Blurred = brand(t.Blurred)
	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description

	return t
}
