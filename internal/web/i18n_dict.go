package web

// i18n_dict.go — a read accessor over the embedded i18n dictionary for
// in-process consumers (SPEC-FEEDBACK-PARTICIPATION-001: the internal/cli
// test that pins the console participation description to the wizard
// question's locale text). The dictionary itself stays an embedded asset;
// this only exposes values so a test in another package can assert the
// shared-locale-strings property without reaching for the file on disk.

import (
	"regexp"
	"strings"
)

// i18nDictLocaleRe matches one locale block opener: two spaces, the locale,
// ": {". Named to avoid the test-side helper of the same purpose.
var i18nDictLocaleRe = regexp.MustCompile(`(?m)^\s{2}(en|ko|ja|zh): \{$`)

// I18nDictionaryValue returns the value bound to key in the given locale's
// block of the embedded i18n.js. ok is false when the locale block or the key
// is absent.
func I18nDictionaryValue(locale, key string) (string, bool) {
	data, err := assetsFS.ReadFile("assets/i18n.js")
	if err != nil {
		return "", false
	}
	js := string(data)

	// Locate the locale block: from its opener to the next block opener.
	locs := i18nDictLocaleRe.FindAllStringSubmatchIndex(js, -1)
	start, end := -1, len(js)
	for i, loc := range locs {
		if loc[2] < 0 || loc[3] < 0 {
			continue
		}
		if js[loc[2]:loc[3]] != locale {
			continue
		}
		start = loc[1]
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		break
	}
	if start < 0 {
		return "", false
	}
	block := js[start:end]

	keyRe := regexp.MustCompile(`"` + regexp.QuoteMeta(key) + `":\s*"((?:[^"\\]|\\.)*)"`)
	m := keyRe.FindStringSubmatch(block)
	if m == nil {
		return "", false
	}
	return strings.ReplaceAll(m[1], `\"`, `"`), true
}
