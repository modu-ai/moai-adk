package wizard

import (
	"fmt"

	"charm.land/huh/v2"
)

// downgradeConfirmText is the localized title format (current, target) and
// description of the `moai update --version` downgrade confirmation.
type downgradeConfirmText struct {
	TitleFormat string
	Description string
}

// downgradeConfirmTexts holds the downgrade confirmation strings for the four
// locales. They move into the translations.go table in plan.md M7.
var downgradeConfirmTexts = map[string]downgradeConfirmText{
	"en": {TitleFormat: "Downgrade %s → %s?", Description: "The requested tag is older than the running version."},
	"ko": {TitleFormat: "다운그레이드할까요? %s → %s", Description: "요청한 태그가 지금 실행 중인 버전보다 오래되었습니다."},
	"ja": {TitleFormat: "ダウングレードしますか？ %s → %s", Description: "指定したタグは実行中のバージョンより古いバージョンです。"},
	"zh": {TitleFormat: "要降级吗？%s → %s", Description: "请求的标签比当前运行的版本旧。"},
}

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
		Value(value)
	return huh.NewForm(huh.NewGroup(conf)).
		WithTheme(newMoAIWizardTheme()).
		WithKeyMap(localizedKeyMap(locale)).
		WithAccessible(false)
}
