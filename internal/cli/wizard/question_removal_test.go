package wizard

// SPEC-CLI-WIZARD-RESTRUCTURE-001 M3 — question removal + LSP default flip
// (AC-WIZ-007/008/009/013, REQ-WIZ-010/012/013/017).

import (
	"strings"
	"testing"
)

// removedInM3 are the questions fixed at their shipped defaults and therefore
// no longer asked: harness_profile (fixed at "default") and
// coverage_exemptions_enabled (fixed at false).
var removedInM3 = []string{"harness_profile", "coverage_exemptions_enabled"}

// TestRemovedQuestionsAbsentFromInitSet pins the question-absence half of
// AC-WIZ-008 and AC-WIZ-009.
func TestRemovedQuestionsAbsentFromInitSet(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	for _, id := range removedInM3 {
		if QuestionByID(InitQuestions(root), id) != nil {
			t.Errorf("%q must not be asked — it is fixed at its shipped default", id)
		}
		if QuestionByID(ReconfigureQuestions(root), id) != nil {
			t.Errorf("%q must not be asked on the reconfigure path either", id)
		}
	}
}

// TestRemovedQuestionsHaveNoOrphanTranslations pins the C15/C16 half of
// AC-WIZ-011: a removed question leaves no ko/ja/zh entry behind.
func TestRemovedQuestionsHaveNoOrphanTranslations(t *testing.T) {
	t.Parallel()
	for _, locale := range localizableLocales {
		langTrans, ok := translations[locale]
		if !ok {
			t.Fatalf("translations for locale %q not found", locale)
		}
		for _, id := range removedInM3 {
			if _, exists := langTrans[id]; exists {
				t.Errorf("locale %q: orphan translation entry for removed question %q", locale, id)
			}
		}
	}
}

// TestRemovedQuestionsHaveNoCaptureBranch pins the C18/C19 half of AC-WIZ-013:
// the answer-capture cases are gone, so feeding the removed IDs stores nothing.
// The retained IDs must still capture (guards against over-deletion).
func TestRemovedQuestionsHaveNoCaptureBranch(t *testing.T) {
	t.Parallel()

	locale := ""
	result := &WizardResult{}
	saveAnswer("harness_profile", "strict", result, &locale)
	if result.HarnessProfile != "" {
		t.Errorf("saveAnswer still captures harness_profile (got %q) — the case must be gone", result.HarnessProfile)
	}

	boolResult := &WizardResult{}
	saveBoolAnswer("coverage_exemptions_enabled", true, boolResult)
	if boolResult.CoverageExemptionsEnabled {
		t.Error("saveBoolAnswer still captures coverage_exemptions_enabled — the case must be gone")
	}

	// Retained capture branches must survive.
	kept := &WizardResult{}
	saveAnswer("project_mode", "team", kept, &locale)
	if kept.ProjectMode != "team" {
		t.Error("project_mode capture branch was removed by mistake")
	}
	for _, tc := range []struct {
		id    string
		check func(*WizardResult) bool
	}{
		{"lsp_enabled", func(r *WizardResult) bool { return r.LSPEnabled }},
		{"enforce_quality", func(r *WizardResult) bool { return r.EnforceQuality }},
		{"design_enabled", func(r *WizardResult) bool { return r.DesignEnabled }},
		{"claude_design_enabled", func(r *WizardResult) bool { return r.ClaudeDesignEnabled }},
	} {
		r := &WizardResult{}
		saveBoolAnswer(tc.id, true, r)
		if !tc.check(r) {
			t.Errorf("%s capture branch was removed by mistake", tc.id)
		}
	}
}

// lspDefaultWording records, PER LOCALE, the affirmative default token the
// lsp_enabled title must now carry and the negative token it must no longer
// carry. Each locale is listed explicitly with its OWN punctuation: ko/ja use a
// halfwidth colon, zh uses a FULLWIDTH one (：). A single shared halfwidth
// pattern would be structurally blind to zh (plan.md §G).
var lspDefaultWording = map[string]struct{ want, gone string }{
	"ko": {"기본값: 예", "기본값: 아니오"},
	"ja": {"デフォルト: はい", "デフォルト: いいえ"},
	"zh": {"默认：是", "默认：否"},
}

// TestLSPEnabledDefaultsToTrue pins AC-WIZ-007: the default flips to true and
// the enabled-by-default wording lands in all four locales.
func TestLSPEnabledDefaultsToTrue(t *testing.T) {
	t.Parallel()
	q := QuestionByID(InitQuestions(t.TempDir()), "lsp_enabled")
	if q == nil {
		t.Fatal("lsp_enabled question not found")
	}

	if q.Default != "true" {
		t.Errorf("lsp_enabled default = %q, want %q", q.Default, "true")
	}

	// English wording lives inline on the question.
	if strings.Contains(q.Title, "default: No") {
		t.Errorf("lsp_enabled English title still says disabled-by-default: %q", q.Title)
	}
	if !strings.Contains(q.Title, "default: Yes") {
		t.Errorf("lsp_enabled English title must state the Yes default, got %q", q.Title)
	}

	for locale, w := range lspDefaultWording {
		localized := GetLocalizedQuestion(q, locale)
		if !strings.Contains(localized.Title, w.want) {
			t.Errorf("locale %q: lsp_enabled title %q must contain %q", locale, localized.Title, w.want)
		}
		if strings.Contains(localized.Title, w.gone) {
			t.Errorf("locale %q: lsp_enabled title %q still carries the disabled-by-default wording %q",
				locale, localized.Title, w.gone)
		}
		if localized.Description == "" {
			t.Errorf("locale %q: lsp_enabled description is empty", locale)
		}
	}
}
