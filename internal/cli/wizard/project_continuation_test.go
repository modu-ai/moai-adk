package wizard

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// findQuestion returns the question carrying id, or nil.
func findQuestion(questions []Question, id string) *Question {
	for i := range questions {
		if questions[i].ID == id {
			return &questions[i]
		}
	}
	return nil
}

// TestProjectContinuationQuestion covers AC-PCK-010 conjunct 1: the question
// exists as a closed select whose default is card and whose option values are
// exactly config.ValidProjectContinuations().
func TestProjectContinuationQuestion(t *testing.T) {
	q := findQuestion(InitQuestions(t.TempDir()), "project_continuation")
	if q == nil {
		t.Fatal("question project_continuation not found in InitQuestions")
	}

	if q.Type != QuestionTypeSelect {
		t.Errorf("Type = %v, want %v", q.Type, QuestionTypeSelect)
	}
	if q.Default != config.ProjectContinuationCard {
		t.Errorf("Default = %q, want %q", q.Default, config.ProjectContinuationCard)
	}

	want := config.ValidProjectContinuations()
	if len(q.Options) != len(want) {
		t.Fatalf("len(Options) = %d, want %d", len(q.Options), len(want))
	}
	for i, opt := range q.Options {
		if opt.Value != want[i] {
			t.Errorf("Options[%d].Value = %q, want %q", i, opt.Value, want[i])
		}
		if opt.Label == "" {
			t.Errorf("Options[%d] (%q) has an empty English Label", i, opt.Value)
		}
		if opt.Desc == "" {
			t.Errorf("Options[%d] (%q) has an empty English Desc", i, opt.Value)
		}
	}
}

// TestProjectContinuationTranslationsExist covers AC-PCK-010 conjunct 2: each
// of ko / ja / zh carries a Title, a Description, and an Options slice of
// length 3 with a non-empty Label AND Desc per element — the audit_model shape.
//
// It duplicates part of what TestWizardQuestionTranslationCompleteness sweeps
// deliberately: the sweep is open-scope and would also go red for an unrelated
// question, so this per-question assertion is what names THIS question when it
// fails.
func TestProjectContinuationTranslationsExist(t *testing.T) {
	q := findQuestion(InitQuestions(t.TempDir()), "project_continuation")
	if q == nil {
		t.Fatal("question project_continuation not found in InitQuestions")
	}

	for _, locale := range localizableLocales {
		langTrans, ok := translations[locale]
		if !ok {
			t.Fatalf("translations for locale %q not found", locale)
		}
		trans, ok := langTrans["project_continuation"]
		if !ok {
			t.Errorf("locale %q: project_continuation translation missing", locale)
			continue
		}
		if trans.Title == "" {
			t.Errorf("locale %q: empty Title", locale)
		}
		if trans.Description == "" {
			t.Errorf("locale %q: empty Description", locale)
		}
		if len(trans.Options) != len(q.Options) {
			t.Errorf("locale %q: %d option translations, want %d", locale, len(trans.Options), len(q.Options))
			continue
		}
		for i, opt := range trans.Options {
			if opt.Label == "" {
				t.Errorf("locale %q: option %d has empty Label", locale, i)
			}
			if opt.Desc == "" {
				t.Errorf("locale %q: option %d has empty Desc", locale, i)
			}
		}
	}
}

// TestProjectContinuationNotOptionTranslationExempt covers AC-PCK-010 conjunct
// 3's second half: the question must NOT be added to optionTranslationExemptIDs
// to satisfy TestWizardQuestionTranslationCompleteness. Exempting it would make
// that sweep pass while the option translations stayed English — the shortcut
// the criterion exists to forbid.
func TestProjectContinuationNotOptionTranslationExempt(t *testing.T) {
	if optionTranslationExemptIDs["project_continuation"] {
		t.Error("project_continuation is in optionTranslationExemptIDs; AC-PCK-010 conjunct 3 forbids exempting it")
	}
}
