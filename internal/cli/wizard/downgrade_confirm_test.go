package wizard

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/ptycaptest"
)

// designHelpLabels is design.md §7's per-key help action table, the oracle the
// wizard's helpActionLabels table must equal.
var designHelpLabels = map[string]map[string]string{
	"ko": {"next": "다음", "submit": "제출", "back": "이전", "select": "선택", "up": "위", "down": "아래",
		"filter": "검색", "set filter": "검색 적용", "clear filter": "검색 해제", "toggle": "전환", "complete": "자동 완성"},
	"ja": {"next": "次へ", "submit": "送信", "back": "戻る", "select": "選択", "up": "上", "down": "下",
		"filter": "絞り込み", "set filter": "絞り込み確定", "clear filter": "絞り込み解除", "toggle": "切替", "complete": "補完"},
	"zh": {"next": "下一步", "submit": "提交", "back": "返回", "select": "选择", "up": "上", "down": "下",
		"filter": "筛选", "set filter": "应用筛选", "clear filter": "清除筛选", "toggle": "切换", "complete": "补全"},
}

// TestHelpActionLabels_MatchDesignTable pins the help label table to
// design.md §7: en is the identity, ko/ja/zh equal the design values, and all
// four locales carry the same action set.
func TestHelpActionLabels_MatchDesignTable(t *testing.T) {
	en, ok := helpActionLabels["en"]
	if !ok || len(en) == 0 {
		t.Fatalf("help label table has no en entries (locales %v)", slices.Sorted(maps.Keys(helpActionLabels)))
	}
	wantActions := slices.Sorted(maps.Keys(designHelpLabels["ko"]))
	for _, loc := range []string{"en", "ko", "ja", "zh"} {
		got := helpActionLabels[loc]
		if keys := slices.Sorted(maps.Keys(got)); !slices.Equal(keys, wantActions) {
			t.Errorf("locale %s action set %q, want %q", loc, keys, wantActions)
		}
	}
	for action, label := range en {
		if label != action {
			t.Errorf("en %q maps to %q, want itself", action, label)
		}
	}
	for loc, want := range designHelpLabels {
		for action, label := range want {
			if got := helpActionLabels[loc][action]; got != label {
				t.Errorf("locale %s %q = %q, want design.md §7 %q", loc, action, got, label)
			}
		}
	}
}

// TestDowngradeConfirmTexts_FourLocales — every locale has a title format
// with both version slots and a description.
func TestDowngradeConfirmTexts_FourLocales(t *testing.T) {
	for _, loc := range []string{"en", "ko", "ja", "zh"} {
		txt, ok := downgradeConfirmTexts[loc]
		if !ok {
			t.Errorf("locale %s: no downgrade confirm text", loc)
			continue
		}
		if strings.Count(txt.TitleFormat, "%s") != 2 {
			t.Errorf("locale %s: title format %q, want two %%s slots", loc, txt.TitleFormat)
		}
		if strings.TrimSpace(txt.Description) == "" {
			t.Errorf("locale %s: empty description", loc)
		}
	}
}

// TestNewDowngradeConfirmForm_KoRender draws the ko confirm and reads the
// rendered help and buttons: action labels from the design table, button
// labels from ConfirmYes/ConfirmNo, and no English action label left.
func TestNewDowngradeConfirmForm_KoRender(t *testing.T) {
	var v bool
	view := ptycaptest.StripANSI(newFormDriver(t, NewDowngradeConfirmForm("ko", "v9.9.9", "v1.0.0", &v)).view())
	ui := GetUIStrings("ko")
	ptycaptest.RequireLines(t, view, "v9.9.9 → v1.0.0")
	for _, want := range []string{ui.ConfirmYes, ui.ConfirmNo, "전환", "제출"} {
		if !strings.Contains(view, want) {
			t.Errorf("ko confirm lacks %q; view:\n%s", want, view)
		}
	}
	for _, en := range []string{"toggle", "submit", "Downgrade", "Yes", "No"} {
		if strings.Contains(view, en) {
			t.Errorf("ko confirm still renders %q; view:\n%s", en, view)
		}
	}
}
