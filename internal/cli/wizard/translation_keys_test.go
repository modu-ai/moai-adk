package wizard

// AC-ITI-014 (REQ-ITI-013) — wizard translation table half: the ko/ja/zh key
// sets equal en's, and every en key is referenced in non-test code outside
// translations.go (the table is the definition, its entries are not reads).

import (
	"os"
	"slices"
	"strings"
	"testing"
)

// TestWizardTranslationKeys_ReferencedAndLocaleEquivalent is AC-ITI-014's
// wizard-table half.
func TestWizardTranslationKeys_ReferencedAndLocaleEquivalent(t *testing.T) {
	// The en key set IS the question set: en lives in the question literals
	// (GetLocalizedQuestion returns them untouched), so the table only needs
	// ko/ja/zh. Collect every question id the three sets can render.
	questionIDs := map[string]bool{}
	for _, q := range DefaultQuestions("/tmp/key-sweep") {
		questionIDs[q.ID] = true
	}
	for _, q := range Page3Questions("/tmp/key-sweep") {
		questionIDs[q.ID] = true
	}
	for _, q := range ReconfigureQuestions("/tmp/key-sweep") {
		questionIDs[q.ID] = true
	}
	wantKeys := make([]string, 0, len(questionIDs))
	for id := range questionIDs {
		wantKeys = append(wantKeys, id)
	}
	slices.Sort(wantKeys)

	// Locale equivalence: ko/ja/zh key sets equal each other and the question
	// set (en's key set is the question set by construction).
	prevKeys := []string(nil)
	for _, loc := range []string{"ko", "ja", "zh"} {
		got := make([]string, 0, len(translations[loc]))
		for key := range translations[loc] {
			got = append(got, key)
		}
		slices.Sort(got)
		if prevKeys != nil && !slices.Equal(prevKeys, got) {
			t.Errorf("locale %s key set differs from the previous locale:\n got  %v\n want %v", loc, got, prevKeys)
		}
		if !slices.Equal(wantKeys, got) {
			t.Errorf("locale %s key set differs from the question set:\n table %v\n want  %v", loc, got, wantKeys)
		}
		prevKeys = got
	}

	// Reference sweep: each key's ID literal must be read in non-test code
	// outside translations.go (questions.go, saveAnswer, the profile question
	// table, ...).
	corpus := wizardNonTestCorpus(t)
	for _, key := range wantKeys {
		if !strings.Contains(corpus, `"`+key+`"`) {
			t.Errorf("wizard translation key %q is never referenced in non-test code — delete the key and its seeds (REQ-ITI-013)", key)
		}
	}
}

// wizardNonTestCorpus concatenates the wizard package's non-test .go files,
// excluding translations.go (the key definitions).
func wizardNonTestCorpus(t *testing.T) string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read wizard dir: %v", err)
	}
	var b strings.Builder
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || name == "translations.go" || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		b.Write(raw)
		b.WriteString("\n")
	}
	return b.String()
}
