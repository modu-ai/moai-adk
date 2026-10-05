package wizard

// The wizard's help-line mechanics: huh's default key map relabelled per
// locale. The LABEL TABLE (helpActionLabels) lives in translations.go with
// the other wizard string tables.

import (
	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
)

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

// HelpActionLabels returns the en help-line action labels for locale — the
// identity for en itself, the localized labels for ko/ja/zh, and nothing for
// an unknown locale. Exported for the profile golden test's closed-exception
// assertion (AC-ITI-008): the en values are the strings that must not leak
// onto a localized screen.
func HelpActionLabels(locale string) []string {
	labels, ok := helpActionLabels[locale]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(labels))
	for _, action := range []string{"next", "submit", "back", "select", "up", "down",
		"filter", "set filter", "clear filter", "toggle", "complete"} {
		out = append(out, labels[action])
	}
	return out
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
