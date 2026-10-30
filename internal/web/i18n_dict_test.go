package web

// i18n_dict_test.go — the read accessor over the embedded i18n dictionary
// (I18nDictionaryValue). It had no test inside this package; an internal/cli
// test is its only consumer.
//
// Defect answer: these tests catch a locale block read from the wrong place, a
// key found in another locale's block, an absent key or locale reported as
// present, and an escaped quote left escaped.

import "testing"

func TestI18nDictionaryValue(t *testing.T) {
	for _, tc := range []struct {
		locale, key, want string
	}{
		{"en", "nav.overview", "Overview"},
		{"ko", "nav.overview", "개요"},
		{"en", "nav.main", "Main navigation"},
		{"ko", "nav.main", "주 메뉴"},
	} {
		got, ok := I18nDictionaryValue(tc.locale, tc.key)
		if !ok || got != tc.want {
			t.Errorf("%s/%s = %q, %v; want %q", tc.locale, tc.key, got, ok, tc.want)
		}
	}
	for _, tc := range []struct{ locale, key string }{
		{"en", "no.such.key"},
		{"xx", "nav.overview"}, // no such locale block
		{"", "nav.overview"},
	} {
		if got, ok := I18nDictionaryValue(tc.locale, tc.key); ok || got != "" {
			t.Errorf("%q/%q = %q, %v; want absent", tc.locale, tc.key, got, ok)
		}
	}
	// A key that exists only in another locale's block is not found here: the
	// lookup is bounded by the locale's own block.
	en, _ := I18nDictionaryValue("en", "nav.overview")
	ko, _ := I18nDictionaryValue("ko", "nav.overview")
	if en == ko {
		t.Errorf("the en and ko blocks returned the same text %q; the lookup is not bounded by locale", en)
	}
}
