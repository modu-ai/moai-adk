package wizard

import (
	"fmt"

	"charm.land/huh/v2"
)

// downgradeConfirmText is the localized title format and description of the
// `moai update --version` downgrade confirmation.
type downgradeConfirmText struct {
	TitleFormat string
	Description string
}

// downgradeConfirmTexts — stub for the RED step (empty).
var downgradeConfirmTexts = map[string]downgradeConfirmText{}

// NewDowngradeConfirmForm — stub for the RED step: the v1 English text on a
// v2 confirm, no localization.
func NewDowngradeConfirmForm(locale, current, target string, value *bool) *huh.Form {
	_ = locale
	conf := huh.NewConfirm().
		Title(fmt.Sprintf("Downgrade %s → %s?", current, target)).
		Description("The requested tag is older than the running version.").
		Value(value)
	return huh.NewForm(huh.NewGroup(conf)).WithTheme(newMoAIWizardTheme()).WithAccessible(false)
}
