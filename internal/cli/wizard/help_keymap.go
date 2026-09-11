package wizard

import (
	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
)

// helpActionLabels maps the English help action strings of huh's default key
// map to each locale (design.md §7). The en column is the identity so every
// locale carries the same action set. Key names (enter, ↑, …) are never
// translated; only the action after them is.
var helpActionLabels = map[string]map[string]string{
	"en": {"next": "next", "submit": "submit", "back": "back", "select": "select", "up": "up", "down": "down",
		"filter": "filter", "set filter": "set filter", "clear filter": "clear filter", "toggle": "toggle", "complete": "complete"},
	"ko": {"next": "다음", "submit": "제출", "back": "이전", "select": "선택", "up": "위", "down": "아래",
		"filter": "검색", "set filter": "검색 적용", "clear filter": "검색 해제", "toggle": "전환", "complete": "자동 완성"},
	"ja": {"next": "次へ", "submit": "送信", "back": "戻る", "select": "選択", "up": "上", "down": "下",
		"filter": "絞り込み", "set filter": "絞り込み確定", "clear filter": "絞り込み解除", "toggle": "切替", "complete": "補完"},
	"zh": {"next": "下一步", "submit": "提交", "back": "返回", "select": "选择", "up": "上", "down": "下",
		"filter": "筛选", "set filter": "应用筛选", "clear filter": "清除筛选", "toggle": "切换", "complete": "补全"},
}

// localizedKeyMap returns huh's default key map with the help action of every
// binding listed in helpActionLabels relabelled for locale; the bound keys and
// key names are unchanged. An unknown locale gets the default (English) map.
// A confirm field's y/n help needs no entry: huh v2 Confirm.View rebuilds those
// two from the Affirmative/Negative button labels (plan.md V-b).
func localizedKeyMap(locale string) *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	labels, ok := helpActionLabels[locale]
	if !ok {
		return km
	}
	for _, b := range keyMapBindings(km) {
		h := b.Help()
		if l, ok := labels[h.Desc]; ok {
			b.SetHelp(h.Key, l)
		}
	}
	return km
}

// keyMapBindings lists every binding of km that carries help text.
func keyMapBindings(km *huh.KeyMap) []*key.Binding {
	return []*key.Binding{
		&km.Input.AcceptSuggestion, &km.Input.Next, &km.Input.Prev, &km.Input.Submit,
		&km.Text.Next, &km.Text.Prev, &km.Text.NewLine, &km.Text.Editor, &km.Text.Submit,
		&km.Select.Next, &km.Select.Prev, &km.Select.Up, &km.Select.Down, &km.Select.HalfPageUp,
		&km.Select.HalfPageDown, &km.Select.GotoTop, &km.Select.GotoBottom, &km.Select.Left,
		&km.Select.Right, &km.Select.Filter, &km.Select.SetFilter, &km.Select.ClearFilter, &km.Select.Submit,
		&km.MultiSelect.Next, &km.MultiSelect.Prev, &km.MultiSelect.Up, &km.MultiSelect.Down,
		&km.MultiSelect.HalfPageUp, &km.MultiSelect.HalfPageDown, &km.MultiSelect.GotoTop,
		&km.MultiSelect.GotoBottom, &km.MultiSelect.Toggle, &km.MultiSelect.Filter,
		&km.MultiSelect.SetFilter, &km.MultiSelect.ClearFilter, &km.MultiSelect.Submit,
		&km.MultiSelect.SelectAll, &km.MultiSelect.SelectNone,
		&km.FilePicker.Open, &km.FilePicker.Close, &km.FilePicker.GotoTop, &km.FilePicker.GotoBottom,
		&km.FilePicker.PageUp, &km.FilePicker.PageDown, &km.FilePicker.Back, &km.FilePicker.Select,
		&km.FilePicker.Up, &km.FilePicker.Down, &km.FilePicker.Prev, &km.FilePicker.Next, &km.FilePicker.Submit,
		&km.Note.Next, &km.Note.Prev, &km.Note.Submit,
		&km.Confirm.Next, &km.Confirm.Prev, &km.Confirm.Toggle, &km.Confirm.Submit,
		&km.Confirm.Accept, &km.Confirm.Reject,
	}
}
