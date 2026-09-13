package wizard

import (
	"fmt"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// The downgrade confirmation's title/description strings
// (downgradeConfirmTexts) live in translations.go with the other wizard
// string tables.

// NewDowngradeConfirmForm builds the one-confirm form that asks whether to
// install target over the running current version. locale is already
// resolved by the caller (design.md §6); an unknown locale renders English.
// Title and description come from downgradeConfirmTexts, the buttons from
// ConfirmYes/ConfirmNo, and the help line from the localized key map.
func NewDowngradeConfirmForm(locale, current, target string, value *bool) *huh.Form {
	txt, ok := downgradeConfirmTexts[locale]
	if !ok {
		locale = "en"
		txt = downgradeConfirmTexts[locale]
	}
	ui := GetUIStrings(locale)
	conf := huh.NewConfirm().
		Title(fmt.Sprintf(txt.TitleFormat, current, target)).
		Description(txt.Description).
		Affirmative(ui.ConfirmYes).
		Negative(ui.ConfirmNo).
		WithButtonAlignment(lipgloss.Left).
		Value(value)
	return huh.NewForm(huh.NewGroup(conf)).
		WithTheme(newMoAIWizardTheme()).
		WithKeyMap(localizedKeyMap(locale)).
		WithAccessible(false)
}
