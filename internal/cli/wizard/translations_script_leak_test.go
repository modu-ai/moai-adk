package wizard

import (
	"fmt"
	"reflect"
	"testing"
	"unicode"
)

// forbiddenScripts lists, per locale, the writing systems that must never
// appear in that locale's strings. Latin is allowed everywhere: brand names
// (Claude, Codex, GitHub), flags (--llm), paths (.claude/) and tier names
// (Max, Opus 5.5) are deliberately left untranslated. Han is allowed in ja
// and zh, where it is native.
var forbiddenScripts = map[string][]*unicode.RangeTable{
	"en": {unicode.Hangul, unicode.Hiragana, unicode.Katakana, unicode.Han},
	"ko": {unicode.Hiragana, unicode.Katakana, unicode.Han},
	"ja": {unicode.Hangul},
	"zh": {unicode.Hangul, unicode.Hiragana, unicode.Katakana},
}

var scriptNames = map[*unicode.RangeTable]string{
	unicode.Hangul: "Hangul", unicode.Hiragana: "Hiragana", unicode.Katakana: "Katakana", unicode.Han: "Han",
}

// localeStrings flattens every per-locale string table in translations.go into
// "path -> value" pairs for one locale, so a new table field is covered without
// touching this test.
func localeStrings(locale string) map[string]string {
	out := map[string]string{}
	add := func(prefix string, v reflect.Value) {
		var walk func(path string, v reflect.Value)
		walk = func(path string, v reflect.Value) {
			switch v.Kind() {
			case reflect.String:
				out[path] = v.String()
			case reflect.Struct:
				for i := 0; i < v.NumField(); i++ {
					walk(path+"."+v.Type().Field(i).Name, v.Field(i))
				}
			case reflect.Slice:
				for i := 0; i < v.Len(); i++ {
					walk(fmt.Sprintf("%s[%d]", path, i), v.Index(i))
				}
			case reflect.Map:
				for _, k := range v.MapKeys() {
					walk(path+"."+k.String(), v.MapIndex(k))
				}
			}
		}
		walk(prefix, v)
	}
	add("translations", reflect.ValueOf(translations[locale]))
	add("uiStrings", reflect.ValueOf(uiStrings[locale]))
	add("helpActionLabels", reflect.ValueOf(helpActionLabels[locale]))
	add("downgradeConfirmTexts", reflect.ValueOf(downgradeConfirmTexts[locale]))
	add("profileQuestionTexts", reflect.ValueOf(profileQuestionTexts[locale]))
	return out
}

// TestTranslationsNoForeignScriptLeak catches a locale's string carrying
// another locale's script — the shape of the defect where the ja agent_wiring
// title shipped as a Korean sentence and no existing translation test noticed,
// because they only check that a value is present, not what it is written in.
func TestTranslationsNoForeignScriptLeak(t *testing.T) {
	for locale, forbidden := range forbiddenScripts {
		strs := localeStrings(locale)
		if len(strs) == 0 {
			t.Errorf("locale %q: no strings collected — the walk is not reaching the tables", locale)
			continue
		}
		for path, s := range strs {
			for _, r := range s {
				for _, tbl := range forbidden {
					if unicode.Is(tbl, r) {
						t.Errorf("locale %q: %s contains %s %q: %q", locale, path, scriptNames[tbl], r, s)
						goto next
					}
				}
			}
		next:
		}
	}
}
