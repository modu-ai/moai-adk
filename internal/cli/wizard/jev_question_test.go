package wizard

import (
	"strings"
	"testing"
)

// jev_question_test.go — the init-only Jev opt-in question
// (SPEC-JEV-OPTIN-MEASURE-001 REQ-JEVO-001/003/004/005/006/007;
// AC-JEVO-004/005/007/008/009).
//
// The placement is the load-bearing part. `ReconfigureQuestions` deliberately
// excludes the page-3 set to preserve its pre-restructure membership, so a
// question placed there is genuinely UNREACHABLE from `moai update
// --reconfigure` — not merely inconvenient. That is the operator's decision,
// and its cost is paid by a sentence: the question text names `moai web` as
// the place the setting can later be changed.

// jevQuestionID is the id under test. It is written out rather than imported
// from the production constant so a rename has to be a deliberate two-file act.
const jevQuestionIDUnderTest = "jev_enabled"

// TestInitQuestions_CarriesExactlyOneJevQuestion is AC-JEVO-007.
func TestInitQuestions_CarriesExactlyOneJevQuestion(t *testing.T) {
	qs := InitQuestions(t.TempDir())
	if len(qs) != 5 {
		ids := make([]string, 0, len(qs))
		for _, q := range qs {
			ids = append(ids, q.ID)
		}
		t.Fatalf("InitQuestions length = %d, want 5; got %v", len(qs), ids)
	}
	count := 0
	for _, q := range qs {
		if q.ID == jevQuestionIDUnderTest {
			count++
			if q.Type != QuestionTypeConfirm {
				t.Errorf("Jev question Type = %v, want QuestionTypeConfirm (a yes/no opt-in)", q.Type)
			}
		}
	}
	if count != 1 {
		t.Errorf("Jev questions in InitQuestions = %d, want exactly 1", count)
	}
}

// TestJevQuestion_AbsentFromDefaultAndReconfigure is AC-JEVO-008: neither the
// shared set nor the reconfigure set carries it, and the reconfigure set keeps
// its pre-SPEC membership exactly.
func TestJevQuestion_AbsentFromDefaultAndReconfigure(t *testing.T) {
	root := t.TempDir()

	for name, qs := range map[string][]Question{
		"DefaultQuestions":     DefaultQuestions(root),
		"ReconfigureQuestions": ReconfigureQuestions(root),
	} {
		for _, q := range qs {
			if q.ID == jevQuestionIDUnderTest {
				t.Errorf("%s contains %q — the operator decision is init-only", name, jevQuestionIDUnderTest)
			}
		}
	}

	// Membership pin. The pre-SPEC reconfigure set is DefaultQuestions with
	// GitQuestions spliced after report_format; spelling it out here is what
	// makes "unchanged" a measurement rather than an assertion.
	wantReconfigure := []string{
		"conversation_language", "user_name", "project_name", "model_policy", "report_format",
		"git_mode", "git_provider", "gitlab_instance_url",
		"github_username", "github_token", "gitlab_username", "gitlab_token",
	}
	got := make([]string, 0, len(wantReconfigure))
	for _, q := range ReconfigureQuestions(root) {
		got = append(got, q.ID)
	}
	if strings.Join(got, ",") != strings.Join(wantReconfigure, ",") {
		t.Errorf("ReconfigureQuestions membership changed\n got: %v\nwant: %v", got, wantReconfigure)
	}
}

// jevLocaleTokens are the per-locale substrings that carry the two obligations
// the question text must satisfy: the privacy statement (REQ-JEVO-003) and the
// `moai web` pointer (REQ-JEVO-007). `moai web` is a command name and stays
// verbatim in every locale; the privacy token is the locale's own word for a
// third party / external server.
var jevLocaleTokens = map[string][]string{
	"en": {"third-party", "moai web"},
	"ko": {"외부", "moai web"},
	"ja": {"外部", "moai web"},
	"zh": {"外部", "moai web"},
}

// TestJevQuestion_TranslatedInFourLocales is AC-JEVO-005, read against the
// tree rather than against the AC's literal wording.
//
// The AC says "an entry exists under each of the four locale keys". This
// package's `translations` map has THREE keys — ko, ja, zh — because
// GetLocalizedQuestion returns the base Question unchanged for "en" and ""
// before it ever consults the map. An "en" entry would therefore be data no
// code path can read, added solely to satisfy a literal reading. All four
// locales ARE covered; en is covered by the base value, which this test reads
// through GetLocalizedQuestion so the coverage is asserted at the surface the
// wizard actually renders from.
func TestJevQuestion_TranslatedInFourLocales(t *testing.T) {
	base := findQuestion(t, InitQuestions(t.TempDir()), jevQuestionIDUnderTest)

	for _, loc := range []string{"en", "ko", "ja", "zh"} {
		q := GetLocalizedQuestion(&base, loc)
		if strings.TrimSpace(q.Title) == "" {
			t.Errorf("locale %q: Title is empty", loc)
		}
		if strings.TrimSpace(q.Description) == "" {
			t.Errorf("locale %q: Description is empty", loc)
		}
		if loc != "en" {
			if _, ok := translations[loc][jevQuestionIDUnderTest]; !ok {
				t.Errorf("translations[%q][%q] missing", loc, jevQuestionIDUnderTest)
			}
			// A missing entry silently falls back to English, so an assertion
			// that only checked non-emptiness would pass over it.
			if q.Title == base.Title {
				t.Errorf("locale %q Title is byte-identical to the English base — untranslated", loc)
			}
		}
	}
}

// TestJevQuestion_PrivacyAndPointerInEveryLocale is AC-JEVO-004 + AC-JEVO-009,
// one sub-case per locale. The English sub-case reads the base Question value
// as well, because that is what renders when no locale entry is selected.
func TestJevQuestion_PrivacyAndPointerInEveryLocale(t *testing.T) {
	base := findQuestion(t, InitQuestions(t.TempDir()), jevQuestionIDUnderTest)

	for _, loc := range []string{"en", "ko", "ja", "zh"} {
		t.Run(loc, func(t *testing.T) {
			q := GetLocalizedQuestion(&base, loc)
			text := q.Title + " " + q.Description
			for _, token := range jevLocaleTokens[loc] {
				if !strings.Contains(text, token) {
					t.Errorf("locale %s question text lacks %q\ntext: %s", loc, token, text)
				}
			}
		})
	}

	baseText := base.Title + " " + base.Description
	for _, token := range jevLocaleTokens["en"] {
		if !strings.Contains(baseText, token) {
			t.Errorf("the base (untranslated) question text lacks %q\ntext: %s", token, baseText)
		}
	}

	// Positive control: the token search fires. A token the text cannot
	// contain must be reported absent, otherwise the assertions above would
	// pass over any text at all.
	if strings.Contains(baseText, "\x00absent-control\x00") {
		t.Error("positive control failed: the substring search matched an impossible token")
	}
	if !strings.Contains(baseText, "Jev") {
		t.Error("positive control failed: the question text does not even name the capability")
	}
}

// TestJevQuestion_DefaultsOff: the opt-in defaults to declining. A capability
// that reaches a third party must not be pre-selected.
func TestJevQuestion_DefaultsOff(t *testing.T) {
	q := findQuestion(t, InitQuestions(t.TempDir()), jevQuestionIDUnderTest)
	if q.Default != "" && q.Default != "false" {
		t.Errorf("Default = %q, want empty or \"false\" — the opt-in must not be pre-selected", q.Default)
	}
	if q.Required {
		t.Error("Required = true; a declined opt-in is a valid answer")
	}
}

// TestSaveBoolAnswer_CapturesJev: the confirm answer reaches the result.
func TestSaveBoolAnswer_CapturesJev(t *testing.T) {
	res := &WizardResult{}
	saveBoolAnswer(jevQuestionIDUnderTest, true, res)
	if !res.JevEnabled {
		t.Error("saveBoolAnswer did not capture the Jev opt-in")
	}
	// Negative control: an unrelated id must not flip it.
	other := &WizardResult{}
	saveBoolAnswer("some_other_confirm", true, other)
	if other.JevEnabled {
		t.Error("an unrelated confirm id flipped JevEnabled")
	}
}

func findQuestion(t *testing.T, qs []Question, id string) Question {
	t.Helper()
	for _, q := range qs {
		if q.ID == id {
			return q
		}
	}
	t.Fatalf("question %q not found", id)
	return Question{}
}
