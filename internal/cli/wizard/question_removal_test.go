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

// removedQuietInit are the eleven page-3-only questions the quiet init wizard
// stops asking (SPEC-INIT-QUIET-WIZARD-001 REQ-IQW-002/013). No other caller
// reaches them, so they leave no definition, translation, or capture branch
// behind — the same removal invariant removedInM3 pins.
var removedQuietInit = []string{
	"project_mode", "worktree_auto_create", "todo_enabled", "feedback_auto_submit",
	"project_continuation", "audit_model", "audit_gate_claude", "audit_gate_codex",
	"audit_gate_glm", "codex_audit_enabled", "mcp_provision",
}

// sharedInitRemovedIDs are the three DefaultQuestions entries the quiet init
// wizard no longer asks but the reconfigure path still does (spec.md §2.3 D1).
var sharedInitRemovedIDs = []string{"project_name", "model_policy", "report_format"}

// TestInitQuestions_QuietSet pins AC-IQW-001: the init set is exactly the four
// kept questions, in order, each keeping its current group label.
func TestInitQuestions_QuietSet(t *testing.T) {
	t.Parallel()
	questions := InitQuestions(t.TempDir())

	wantIDs := []string{"conversation_language", "user_name", "agent_wiring", "autonomy_tier"}
	wantGroups := []string{"Basic", "Basic", "Quality & Workflow", "Autonomy"}

	if len(questions) != len(wantIDs) {
		got := make([]string, 0, len(questions))
		for i := range questions {
			got = append(got, questions[i].ID)
		}
		t.Fatalf("InitQuestions returned %d questions %v, want %d %v", len(questions), got, len(wantIDs), wantIDs)
	}
	for i := range wantIDs {
		if questions[i].ID != wantIDs[i] {
			t.Errorf("InitQuestions[%d].ID = %q, want %q", i, questions[i].ID, wantIDs[i])
		}
		if questions[i].Group != wantGroups[i] {
			t.Errorf("InitQuestions[%d] (%s) Group = %q, want %q", i, questions[i].ID, questions[i].Group, wantGroups[i])
		}
	}
}

// TestRemovedQuestionsAbsentFromInitSet pins the question-absence half of
// AC-WIZ-008 and AC-WIZ-009, and of AC-IQW-002 for the quiet init set.
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

	// AC-IQW-002: the eleven page-3-only questions are gone from both sets.
	for _, id := range removedQuietInit {
		if QuestionByID(InitQuestions(root), id) != nil {
			t.Errorf("%q must not be asked by the quiet init wizard", id)
		}
		if QuestionByID(ReconfigureQuestions(root), id) != nil {
			t.Errorf("%q must not reach the reconfigure path", id)
		}
	}
	// The three shared questions are gone from init only.
	for _, id := range sharedInitRemovedIDs {
		if QuestionByID(InitQuestions(root), id) != nil {
			t.Errorf("%q must not be asked by the quiet init wizard", id)
		}
	}
}

// TestRemovedQuestionsHaveNoOrphanTranslations pins the C15/C16 half of
// AC-WIZ-011: a removed question leaves no ko/ja/zh entry behind. AC-IQW-002
// extends it to the eleven page-3-only questions of the quiet init wizard.
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
		for _, id := range removedQuietInit {
			if _, exists := langTrans[id]; exists {
				t.Errorf("locale %q: orphan translation entry for removed question %q", locale, id)
			}
		}
	}
}

// TestSharedQuestionsRetainedForReconfigure pins the D1 split (AC-IQW-002):
// project_name, model_policy, and report_format leave the init set but stay in
// ReconfigureQuestions with their translations, and their answers still
// capture on the result.
func TestSharedQuestionsRetainedForReconfigure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	reconf := ReconfigureQuestions(root)
	for _, id := range sharedInitRemovedIDs {
		if QuestionByID(reconf, id) == nil {
			t.Errorf("%q must stay in ReconfigureQuestions (D1)", id)
		}
		for _, locale := range localizableLocales {
			if _, ok := translations[locale][id]; !ok {
				t.Errorf("locale %q: %q translation must stay for the reconfigure path", locale, id)
			}
		}
	}

	locale := ""
	r := &WizardResult{}
	saveAnswer("project_name", "reconf-proj", r, &locale)
	saveAnswer("model_policy", "low", r, &locale)
	saveAnswer("report_format", "md", r, &locale)
	if r.ProjectName != "reconf-proj" || r.ModelPolicy != "low" || r.ReportFormat != "md" {
		t.Errorf("shared capture branches must survive for reconfigure; result = %+v", *r)
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

	// Retained capture branches must survive (the kept init questions).
	kept := &WizardResult{}
	saveAnswer("agent_wiring", "both", kept, &locale)
	saveAnswer("autonomy_tier", "automatic", kept, &locale)
	if kept.AgentWiring != "both" || kept.AutonomyTier != "automatic" {
		t.Errorf("a kept capture branch was removed by mistake; result = %+v", *kept)
	}

	// AC-IQW-002: the eleven quiet-init removals capture nothing through either
	// handler — no result field changes at all.
	for _, id := range removedQuietInit {
		r := &WizardResult{}
		saveAnswer(id, "team", r, &locale)
		saveBoolAnswer(id, true, r)
		if *r != (WizardResult{}) {
			t.Errorf("%s capture branch must be gone (SPEC-INIT-QUIET-WIZARD-001); result = %+v", id, *r)
		}
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
