package hook

import (
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// SPEC-FACTORY-STALE-RUN-HEAL-001 REQ-SRH-002 / quality gate "i18n parity":
// the four locales of the stale-run notice (N5 and N6 among them) carry the
// same protocol tokens — the same format verbs per field, the quoted commands,
// and the "stale run" prefix — so a locale cannot drop or reorder a command
// line argument the operator has to run.
var (
	staleRunVerbRE    = regexp.MustCompile(`%\[\d+\][a-z]`)
	staleRunCommandRE = regexp.MustCompile(`'moai [^']+'`)
)

func TestStaleRunLocalesProtocolTokenParity(t *testing.T) {
	locales := []string{langEnglish, "ko", "ja", "zh"}
	for _, lang := range locales {
		if _, ok := staleRunLocales[lang]; !ok {
			t.Fatalf("staleRunLocales has no %q entry", lang)
		}
	}
	if len(staleRunLocales) != len(locales) {
		t.Fatalf("staleRunLocales carries %d locales, want exactly %v", len(staleRunLocales), locales)
	}
	base := reflect.ValueOf(staleRunLocales[langEnglish])
	typ := base.Type()
	fields := 0
	for i := 0; i < typ.NumField(); i++ {
		name := typ.Field(i).Name
		fields++
		want := staleRunTokens(base.Field(i).String())
		for _, lang := range locales {
			got := reflect.ValueOf(staleRunLocales[lang]).Field(i).String()
			if strings.TrimSpace(got) == "" {
				t.Errorf("%s[%s] is empty", name, lang)
				continue
			}
			if got != strings.TrimSpace(got) {
				t.Errorf("%s[%s] carries leading or trailing whitespace", name, lang)
			}
			if tokens := staleRunTokens(got); !reflect.DeepEqual(tokens, want) {
				t.Errorf("%s[%s] protocol tokens = %v, want %v (the English table's)", name, lang, tokens, want)
			}
			if gotNL, wantNL := strings.Count(got, "\n"), strings.Count(base.Field(i).String(), "\n"); gotNL != wantNL {
				t.Errorf("%s[%s] has %d interior newlines, want %d", name, lang, gotNL, wantNL)
			}
		}
	}
	if fields == 0 {
		t.Fatal("no fields compared")
	}
}

// staleRunTokens is the sorted multiset of protocol tokens one message
// carries: its format verbs (with their explicit argument indices) and its
// quoted commands.
func staleRunTokens(msg string) []string {
	tokens := append(staleRunVerbRE.FindAllString(msg, -1), staleRunCommandRE.FindAllString(msg, -1)...)
	sort.Strings(tokens)
	return tokens
}
