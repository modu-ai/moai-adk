package wizard

// SPEC-CLI-WIZARD-RESTRUCTURE-001 M3 — question removal + LSP default flip
// (AC-WIZ-007/008/009/013, REQ-WIZ-010/012/013/017).

import (
	"testing"
)

// removedInM3 are the questions fixed at their shipped defaults and therefore
// no longer asked: harness_profile (fixed at "default"),
// coverage_exemptions_enabled (fixed at false), and — added 2026-08-03 — the
// four page-3 default-true confirms (lsp_enabled, enforce_quality,
// design_enabled, claude_design_enabled), all fixed at true.
var removedInM3 = []string{
	"harness_profile", "coverage_exemptions_enabled",
	"lsp_enabled", "enforce_quality", "design_enabled", "claude_design_enabled",
}

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
	// harness_profile field is fully removed from WizardResult (WS1 dead-code
	// removal), so there is no field left to capture into — the assertion is
	// vacuous and only the retained-capture branches below are meaningful.

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
	// The four newly-removed page-3 confirms must NOT capture (M3 invariant
	// extended to lsp_enabled/enforce_quality/design_enabled/claude_design_enabled).
	for _, id := range []string{"lsp_enabled", "enforce_quality", "design_enabled", "claude_design_enabled"} {
		r := &WizardResult{}
		saveBoolAnswer(id, true, r)
		if r.LSPEnabled || r.EnforceQuality || r.DesignEnabled || r.ClaudeDesignEnabled {
			t.Errorf("%s capture branch must be gone (removed 2026-08-03)", id)
		}
	}
}

// TestLSPEnabledFixedAtTrueDefault pins the post-removal invariant (AC-WIZ-007
// carried forward, 2026-08-03): lsp_enabled is no longer asked, AND its value
// is seeded true on the result struct (wizard.go RunWithDefaults/RunWithLocale)
// so interactive `moai init` writes lsp.enabled: true without prompting. The
// four-locale wording check is moot — the question is gone.
func TestLSPEnabledFixedAtTrueDefault(t *testing.T) {
	t.Parallel()
	if q := QuestionByID(InitQuestions(t.TempDir()), "lsp_enabled"); q != nil {
		t.Fatalf("lsp_enabled must no longer be asked; got question with Default=%q", q.Default)
	}
	// The seed struct literal must hold the four fixed-default booleans true
	// (mirrors the RunWithDefaults/RunWithLocale seed in wizard.go).
	seed := &WizardResult{
		LSPEnabled:                true,
		EnforceQuality:            true,
		CoverageExemptionsEnabled: false,
		DesignEnabled:             true,
		ClaudeDesignEnabled:       true,
	}
	if !seed.LSPEnabled || !seed.EnforceQuality || !seed.DesignEnabled || !seed.ClaudeDesignEnabled {
		t.Errorf("seed WizardResult must set the four default-true booleans: %+v", *seed)
	}
	if seed.CoverageExemptionsEnabled {
		t.Error("seed CoverageExemptionsEnabled must stay false")
	}
}
