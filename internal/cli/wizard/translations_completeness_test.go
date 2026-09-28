package wizard

import "testing"

// localizableLocales is the set of non-English locales the wizard ships
// translations for. English is the source language (no translation table).
var localizableLocales = []string{"ko", "ja", "zh"}

// optionTranslationExemptIDs are Select questions whose options are
// intentionally NOT carried in the fixed-length translation table:
//   - conversation_language: option labels are native language names, never
//     translated (GetLocalizedQuestion leaves them untouched).
var optionTranslationExemptIDs = map[string]bool{
	"conversation_language": true,
}

// TestWizardQuestionTranslationCompleteness is the regression guard the user's
// bug proved we need: EVERY interactive question in the init wizard set
// (InitQuestions: pages 1-3) must have a title + description
// translation for ko/ja/zh, and every Select question with a static option set
// (all except the exempt IDs) must have matching-length, non-empty option
// translations. Adding a new question without translations will FAIL this test.
func TestWizardQuestionTranslationCompleteness(t *testing.T) {
	root := t.TempDir()
	questions := InitQuestions(root)

	for _, locale := range localizableLocales {
		langTrans, ok := translations[locale]
		if !ok {
			t.Fatalf("translations for locale %q not found", locale)
		}
		for i := range questions {
			q := &questions[i]
			trans, ok := langTrans[q.ID]
			if !ok {
				t.Errorf("locale %q: question %q has NO translation entry", locale, q.ID)
				continue
			}
			if trans.Title == "" {
				t.Errorf("locale %q: question %q has empty title translation", locale, q.ID)
			}
			if trans.Description == "" {
				t.Errorf("locale %q: question %q has empty description translation", locale, q.ID)
			}
			if q.Type != QuestionTypeSelect || optionTranslationExemptIDs[q.ID] {
				continue
			}
			if len(trans.Options) != len(q.Options) {
				t.Errorf("locale %q: question %q has %d option translations, want %d",
					locale, q.ID, len(trans.Options), len(q.Options))
				continue
			}
			for j, opt := range trans.Options {
				if opt.Label == "" {
					t.Errorf("locale %q: question %q option %d has empty label", locale, q.ID, j)
				}
				if opt.Desc == "" {
					t.Errorf("locale %q: question %q option %d has empty desc", locale, q.ID, j)
				}
			}
		}
	}
}

// TestConfirmButtonLocalization verifies GetUIStrings returns localized
// affirmative/negative button labels for every supported locale (bound onto the
// huh Confirm via Affirmative/Negative in buildConfirmField).
func TestConfirmButtonLocalization(t *testing.T) {
	tests := []struct {
		locale  string
		wantYes string
		wantNo  string
	}{
		{"en", "Yes", "No"},
		{"ko", "예", "아니오"},
		{"ja", "はい", "いいえ"},
		{"zh", "是", "否"},
	}
	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			ui := GetUIStrings(tt.locale)
			if ui.ConfirmYes != tt.wantYes {
				t.Errorf("locale %q: ConfirmYes = %q, want %q", tt.locale, ui.ConfirmYes, tt.wantYes)
			}
			if ui.ConfirmNo != tt.wantNo {
				t.Errorf("locale %q: ConfirmNo = %q, want %q", tt.locale, ui.ConfirmNo, tt.wantNo)
			}
		})
	}

	// Unknown locale must fall back to English (existing GetUIStrings contract).
	if got := GetUIStrings("xx"); got.ConfirmYes != "Yes" || got.ConfirmNo != "No" {
		t.Errorf("unknown locale fallback: got Yes=%q No=%q, want Yes/No", got.ConfirmYes, got.ConfirmNo)
	}
}
