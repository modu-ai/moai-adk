package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/web"
)

// TestParticipationConsoleDescMatchesWizardText pins the shared-locale-strings
// property (SPEC-FEEDBACK-PARTICIPATION-001 REQ-ANON-005/020): the console's
// participation description is byte-equal to the wizard question description
// per locale, so the two enablement surfaces cannot drift apart. The test
// lives here because this package is the only place that can name both the
// wizard texts (through GetLocalizedQuestion) and the web dictionary.
func TestParticipationConsoleDescMatchesWizardText(t *testing.T) {
	enQ := wizard.QuestionByID(wizard.InitQuestions(t.TempDir()), wizard.ParticipationQuestionID)
	if enQ == nil {
		t.Fatal("no participation question in the init set")
	}

	for _, locale := range []string{"en", "ko", "ja", "zh"} {
		lq := wizard.GetLocalizedQuestion(enQ, locale)
		if lq.Description == "" {
			t.Fatalf("locale %s: wizard participation description is empty", locale)
		}
		got, ok := web.I18nDictionaryValue(locale, "f.feedback.participation.desc")
		if !ok {
			t.Fatalf("web i18n dictionary has no f.feedback.participation.desc for locale %s", locale)
		}
		if got != lq.Description {
			t.Errorf("locale %s: console desc drifts from the wizard question text.\nconsole: %q\nwizard:  %q", locale, got, lq.Description)
		}
	}
}
