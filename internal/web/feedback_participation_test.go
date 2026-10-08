package web

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// i18nLocaleBlock returns the dictionary segment for one locale: from its
// `  <locale>: {` line to the next locale block or the end of the map.
func i18nLocaleBlock(t *testing.T, js, locale string) string {
	t.Helper()
	start := strings.Index(js, "\n  "+locale+": {")
	if start < 0 {
		t.Fatalf("i18n.js has no %q locale block", locale)
	}
	rest := js[start+len("\n  "+locale+": {"):]
	end := len(rest)
	for _, other := range []string{"en", "ko", "ja", "zh"} {
		if i := strings.Index(rest, "\n  "+other+": {"); i >= 0 && i < end {
			end = i
		}
	}
	return rest[:end]
}

// i18nLocaleValue extracts the string value bound to key inside one locale
// block. A missing key is a test failure.
func i18nLocaleValue(t *testing.T, js, locale, key string) string {
	t.Helper()
	block := i18nLocaleBlock(t, js, locale)
	re := regexp.MustCompile(`"` + regexp.QuoteMeta(key) + `":\s*"((?:[^"\\]|\\.)*)"`)
	m := re.FindStringSubmatch(block)
	if m == nil {
		t.Fatalf("i18n.js locale %s has no entry for %q", locale, key)
	}
	return strings.ReplaceAll(m[1], `\"`, `"`)
}

// TestFeedbackParticipationI18nKeysInAllLocales asserts the four-locale title
// and description keys exist, and that each description carries the full
// REQ-ANON-005 required-fact set — the same disclosure the wizard question
// carries, so console-first enablement is informed before the value is
// written.
func TestFeedbackParticipationI18nKeysInAllLocales(t *testing.T) {
	raw, err := os.ReadFile("assets/i18n.js")
	if err != nil {
		t.Fatalf("read i18n.js: %v", err)
	}
	js := string(raw)

	for _, locale := range []string{"en", "ko", "ja", "zh"} {
		for _, key := range []string{"f.feedback.participation.title", "f.feedback.participation.desc"} {
			i18nLocaleValue(t, js, locale, key) // failure names the missing entry
		}
	}

	facts := map[string][]string{
		"en": {
			"public GitHub issue", "your own GitHub account", "creation time",
			"no per-report confirmation", "later comments", "publicly associated",
			"fixed fields", "moai feedback participation preview", "moai web",
			"cannot be recalled", "model-subscription tokens", "template text",
		},
		"ko": {
			"공개 이슈", "GitHub 계정", "생성 시각", "확인을 받지 않고", "구독",
			"공개적으로 연결", "고정 필드", "moai feedback participation preview",
			"moai web", "취소할 수 없습니다", "모델 구독 토큰", "템플릿 본문",
		},
		"ja": {
			"公開イシュー", "GitHub アカウント", "作成日時", "確認はありません", "購読",
			"公開的に結び付", "固定フィールド", "moai feedback participation preview",
			"moai web", "取り消すことはできません", "モデル購読トークン", "テンプレート本文",
		},
		"zh": {
			"公开 issue", "GitHub 账号", "创建时间", "确认", "订阅", "公开关联",
			"固定字段", "moai feedback participation preview", "moai web",
			"无法", "模型订阅", "模板文本",
		},
	}
	for locale, parts := range facts {
		desc := i18nLocaleValue(t, js, locale, "f.feedback.participation.desc")
		for _, fact := range parts {
			if !strings.Contains(desc, fact) {
				t.Errorf("locale %s: console participation description misses required fact %q", locale, fact)
			}
		}
	}
}

// TestFeedbackParticipationRendersAsRadioPair asserts the widget shape and
// the non-rendering of the asked marker on the rendered console page.
func TestFeedbackParticipationRendersAsRadioPair(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	html := renderConsolePage(t)

	if got := strings.Count(html, `name="feedback.participation"`); got != 2 {
		t.Errorf("feedback.participation has %d inputs, want exactly 2 (the two-option radio pair)", got)
	}
	if !strings.Contains(html, `name="feedback.participation__present"`) {
		t.Error("feedback.participation lost its hidden __present companion")
	}
	if n := countCheckboxInputs(html); n != 0 {
		t.Errorf("rendered console has %d checkbox inputs, want 0", n)
	}
	if strings.Contains(html, `name="feedback.participation_asked"`) {
		t.Error("the asked marker key rendered; it must never appear as a form field")
	}
}
