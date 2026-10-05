package wizard

import (
	"reflect"
	"sort"
	"testing"
)

var profileLocales = []string{"en", "ko", "ja", "zh"}

func sortedKeys(m map[string]QuestionTranslation) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// TestProfileTranslations_FourLocaleKeyParity: every locale carries exactly the
// profile question id set, with a non-empty title and description per id.
func TestProfileTranslations_FourLocaleKeyParity(t *testing.T) {
	want := append([]string(nil), profileFieldIDs...)
	sort.Strings(want)
	if len(profileQuestionTexts) != len(profileLocales) {
		t.Fatalf("profile translation table carries %d locales, want %d (%v)", len(profileQuestionTexts), len(profileLocales), profileLocales)
	}
	for _, loc := range profileLocales {
		table, ok := profileQuestionTexts[loc]
		if !ok {
			t.Errorf("locale %q missing from the profile translation table", loc)
			continue
		}
		if got := sortedKeys(table); !reflect.DeepEqual(got, want) {
			t.Errorf("locale %q key set:\n got  %v\n want %v", loc, got, want)
		}
		for id, tr := range table {
			if tr.Title == "" || tr.Description == "" {
				t.Errorf("locale %q id %q: empty title or description", loc, id)
			}
			if len(tr.Options) != 0 {
				t.Errorf("locale %q id %q: option labels are resolved by the caller, the table must not carry any", loc, id)
			}
		}
	}
}

// TestProfileTranslations_NonEnglishLocalesAreTranslated guards against a
// locale block that was copied from English and never translated.
func TestProfileTranslations_NonEnglishLocalesAreTranslated(t *testing.T) {
	en := profileQuestionTexts["en"]
	if len(en) == 0 {
		t.Fatal("english profile translation block is empty")
	}
	for _, loc := range []string{"ko", "ja", "zh"} {
		for id, tr := range profileQuestionTexts[loc] {
			if tr.Title == en[id].Title {
				t.Errorf("locale %q id %q: title equals the English title %q", loc, id, tr.Title)
			}
			if tr.Description == en[id].Description {
				t.Errorf("locale %q id %q: description equals the English description", loc, id)
			}
		}
	}
}

// TestProfileQuestions_EnglishTextSourcedFromTable: the question set's own
// Title/Description are the English table entries (single source).
func TestProfileQuestions_EnglishTextSourcedFromTable(t *testing.T) {
	qs := ProfileQuestions(sampleProfileOptions(), ProfileResult{})
	if len(qs) != len(profileFieldIDs) {
		t.Fatalf("want %d questions, got %d", len(profileFieldIDs), len(qs))
	}
	for i := range qs {
		tr := profileQuestionTexts["en"][qs[i].ID]
		if qs[i].Title != tr.Title || qs[i].Description != tr.Description {
			t.Errorf("%s: question text (%q, %q) differs from the en table (%q, %q)",
				qs[i].ID, qs[i].Title, qs[i].Description, tr.Title, tr.Description)
		}
	}
}

func TestLocalizeProfileQuestion(t *testing.T) {
	qs := ProfileQuestions(sampleProfileOptions(), ProfileResult{ModelPolicy: "high"})
	var policy *Question
	for i := range qs {
		if qs[i].ID == "model_policy" {
			policy = &qs[i]
		}
	}
	if policy == nil {
		t.Fatal("model_policy question missing")
	}

	for _, loc := range []string{"ko", "ja", "zh"} {
		got := LocalizeProfileQuestion(policy, loc)
		want := profileQuestionTexts[loc]["model_policy"]
		if got.Title != want.Title || got.Description != want.Description {
			t.Errorf("locale %q: got (%q, %q), want (%q, %q)", loc, got.Title, got.Description, want.Title, want.Description)
		}
		if !reflect.DeepEqual(got.Options, policy.Options) || got.Default != policy.Default || got.ID != policy.ID || got.Group != policy.Group {
			t.Errorf("locale %q: localization changed a non-text field: %+v", loc, got)
		}
	}
	for _, loc := range []string{"en", "", "fr"} {
		got := LocalizeProfileQuestion(policy, loc)
		if got.Title != policy.Title || got.Description != policy.Description {
			t.Errorf("locale %q: want the English text, got (%q, %q)", loc, got.Title, got.Description)
		}
	}
	unknown := Question{ID: "not_a_profile_question", Title: "T", Description: "D"}
	if got := LocalizeProfileQuestion(&unknown, "ko"); got.Title != "T" || got.Description != "D" {
		t.Errorf("unknown id: want the question unchanged, got (%q, %q)", got.Title, got.Description)
	}
}
