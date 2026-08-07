package wizard

import (
	"testing"
)

// TestPage3QuestionsStructure verifies each page-3 Question entry has the
// required fields, in the REQ-WIZ-005 order. The page-3 questions are
// unconditional; claude_design_enabled (nested on design_enabled) is the only
// one that still carries a Condition.
func TestPage3QuestionsStructure(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	questions := Page3Questions(tmpDir)

	// Expected IDs, their types, and whether they are gated.
	type entry struct {
		id      string
		qtype   QuestionType
		hasOpts bool // whether Options slice is non-empty
		gated   bool // whether Condition is non-nil
	}

	want := []entry{
		{"project_mode", QuestionTypeSelect, true, false},
		{"worktree_auto_create", QuestionTypeConfirm, false, false},
		// SPEC-MOAI-MCP-SERVER-001 M4 (REQ-MCP-015 / AC-MCP-020): audit + MCP
		// opt-in selection, grouped under "Audit & MCP". Enum values reuse the
		// M3 typed-config constants (see mcp_audit_test.go).
		{"audit_model", QuestionTypeSelect, true, false},
		{"audit_gate_claude", QuestionTypeSelect, true, false},
		{"audit_gate_codex", QuestionTypeSelect, true, false},
		{"audit_gate_glm", QuestionTypeSelect, true, false},
		{"codex_audit_enabled", QuestionTypeConfirm, false, false},
		{"mcp_tools_opt_in", QuestionTypeConfirm, false, false},
	}

	if len(questions) != len(want) {
		t.Fatalf("Page3Questions() returned %d questions, want %d", len(questions), len(want))
	}

	for i, w := range want {
		q := questions[i]
		if q.ID != w.id {
			t.Errorf("questions[%d].ID = %q, want %q", i, q.ID, w.id)
		}
		if q.Type != w.qtype {
			t.Errorf("questions[%d].Type = %v, want %v", i, q.Type, w.qtype)
		}
		if w.hasOpts && len(q.Options) == 0 {
			t.Errorf("questions[%d] (%s) has no Options", i, q.ID)
		}
		if got := q.Condition != nil; got != w.gated {
			t.Errorf("questions[%d] (%s) gated = %v, want %v", i, q.ID, got, w.gated)
		}
	}
}

// TestPage3VisibleWithoutBridge supersedes TestAdvancedBridgeRevealsPhase1.
// The advanced_bridge no longer reveals page 3: page 3 is unconditional, so it
// is visible from the start (REQ-WIZ-001/002) with no bridge answered. The
// coverage intent is preserved but inverted — this now asserts the ABSENCE of
// the gate, through the visibility lens (FilteredQuestions).
func TestPage3VisibleWithoutBridge(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	all := InitQuestions(tmpDir)

	// A fresh result with no bridge answered and no mode selected.
	quick := &WizardResult{EnforceQuality: true, DesignEnabled: true, ClaudeDesignEnabled: true}
	visible := FilteredQuestions(all, quick)

	if QuestionByID(visible, "project_mode") == nil {
		t.Error(`page-3 question "project_mode" must be visible with no bridge answered`)
	}
}

// TestPage3Questions_Ungated supersedes the former mode-gating test: C3 removed
// the gate and REQ-WIZ-018 retired the flag it read, so the page-3 questions are
// visible for any result whose DesignEnabled reveals the one nested question.
// The gate cannot be re-asserted through a field that no longer exists, so the
// invariant is pinned structurally instead: only claude_design_enabled carries a
// Condition, and the whole page is visible in one pass.
func TestPage3Questions_Ungated(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	questions := Page3Questions(tmpDir)

	visible := FilteredQuestions(questions, &WizardResult{DesignEnabled: true})
	if QuestionByID(visible, "project_mode") == nil {
		t.Error(`page-3 question "project_mode" must be visible`)
	}

	// Structural half: nothing on page 3 is gated.
	for _, q := range questions {
		if q.Condition != nil {
			t.Errorf("page-3 question %q carries a Condition — page 3 must be ungated", q.ID)
		}
	}
}

// TestPage3Questions_NoConditional verifies page 3 now has NO Condition-gated
// question: claude_design_enabled (the last conditional) was removed together
// with the four default-true confirms (2026-08-03).
func TestPage3Questions_NoConditional(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	for _, q := range Page3Questions(tmpDir) {
		if q.Condition != nil {
			t.Errorf("page-3 question %q carries a Condition — page 3 must have no conditional questions", q.ID)
		}
	}
}

// TestProjectModeQuestion verifies project_mode has exactly 2 options (personal/team).
func TestProjectModeQuestion(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	questions := Page3Questions(tmpDir)

	q := QuestionByID(questions, "project_mode")
	if q == nil {
		t.Fatal("project_mode question not found")
	}
	if len(q.Options) != 2 {
		t.Errorf("project_mode has %d options, want 2", len(q.Options))
	}
	// Verify values
	values := make([]string, len(q.Options))
	for i, o := range q.Options {
		values[i] = o.Value
	}
	if values[0] != "personal" || values[1] != "team" {
		t.Errorf("project_mode option values = %v, want [personal, team]", values)
	}
}

// TestSaveAnswerPhase1 verifies saveAnswer stores the page-3 string field.
// harness_profile is deliberately absent — its capture branch was removed with
// the question (REQ-WIZ-012); the absence is asserted in
// TestRemovedQuestionsHaveNoCaptureBranch.
func TestSaveAnswerPhase1(t *testing.T) {
	t.Parallel()
	locale := ""
	result := &WizardResult{}

	saveAnswer("project_mode", "team", result, &locale)
	if result.ProjectMode != "team" {
		t.Errorf("ProjectMode = %q, want 'team'", result.ProjectMode)
	}
}

// TestSaveBoolAnswer verifies saveBoolAnswer is now a no-op: the four page-3
// confirm questions (lsp_enabled, enforce_quality, design_enabled,
// claude_design_enabled) are no longer asked (removed 2026-08-03), so no
// boolean answer is captured. A zero result must stay zero.
func TestSaveBoolAnswer(t *testing.T) {
	t.Parallel()
	result := &WizardResult{}
	for _, id := range []string{"lsp_enabled", "enforce_quality", "design_enabled", "claude_design_enabled"} {
		saveBoolAnswer(id, true, result)
	}
	if result.LSPEnabled || result.EnforceQuality || result.DesignEnabled || result.ClaudeDesignEnabled {
		t.Errorf("saveBoolAnswer must be a no-op for the four removed IDs; result = %+v", *result)
	}
}

// TestBuildQuestionGroup_ConfirmType verifies buildQuestionGroup handles QuestionTypeConfirm.
func TestBuildQuestionGroup_ConfirmType(t *testing.T) {
	t.Parallel()
	locale := ""
	result := &WizardResult{}
	q := &Question{
		ID:      "lsp_enabled",
		Type:    QuestionTypeConfirm,
		Title:   "Enable LSP?",
		Default: "false",
	}
	// buildQuestionGroup should not panic for Confirm type
	group := buildQuestionGroup(q, result, &locale)
	if group == nil {
		t.Error("buildQuestionGroup returned nil for QuestionTypeConfirm")
	}
}

// TestWizardResultDefaultsPrePopulated verifies RunWithDefaults pre-populates
// the result with correct defaults before wizard interaction. The two mode
// fields it used to seed are retired (REQ-WIZ-018), so only the four boolean
// defaults remain.
func TestWizardResultDefaultsPrePopulated(t *testing.T) {
	t.Parallel()
	// We test the pre-population logic directly since we can't run interactive wizard in tests.
	result := &WizardResult{
		LSPEnabled:                true,
		EnforceQuality:            true,
		CoverageExemptionsEnabled: false,
		DesignEnabled:             true,
		ClaudeDesignEnabled:       true,
	}

	if !result.LSPEnabled {
		t.Error("LSPEnabled default should be true")
	}
	if !result.EnforceQuality {
		t.Error("EnforceQuality default should be true")
	}
	if result.CoverageExemptionsEnabled {
		t.Error("CoverageExemptionsEnabled default should be false")
	}
	if !result.DesignEnabled {
		t.Error("DesignEnabled default should be true")
	}
	if !result.ClaudeDesignEnabled {
		t.Error("ClaudeDesignEnabled default should be true")
	}
}

// TestBuildConfirmField_DefaultFalse verifies buildConfirmField initializes
// with the correct default value when Default="false".
func TestBuildConfirmField_DefaultFalse(t *testing.T) {
	t.Parallel()
	locale := ""
	result := &WizardResult{}
	q := &Question{
		ID:      "lsp_enabled",
		Type:    QuestionTypeConfirm,
		Title:   "Enable LSP?",
		Default: "false",
	}
	field := buildConfirmField(q, result, &locale)
	if field == nil {
		t.Error("buildConfirmField returned nil")
	}
}

// TestBuildConfirmField_DefaultTrue verifies buildConfirmField parses "true" correctly.
func TestBuildConfirmField_DefaultTrue(t *testing.T) {
	t.Parallel()
	locale := ""
	result := &WizardResult{}
	q := &Question{
		ID:      "enforce_quality",
		Type:    QuestionTypeConfirm,
		Title:   "Enforce quality?",
		Default: "true",
	}
	field := buildConfirmField(q, result, &locale)
	if field == nil {
		t.Error("buildConfirmField returned nil")
	}
}

// TestTotalVisibleQuestions_Page3AlwaysCounted supersedes the former
// mode-parameterised stepper test: the denominator no longer depends on any
// mode flag for the page-3 questions, and REQ-WIZ-018 retired the flag itself.
func TestTotalVisibleQuestions_Page3AlwaysCounted(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	all := InitQuestions(tmpDir)

	// No mode is selected — page 3 no longer needs one.
	// DesignEnabled reveals the nested claude_design_enabled.
	res := &WizardResult{DesignEnabled: true}
	got := TotalVisibleQuestions(all, res)
	// Page 1 (3) + Page 2 (2) + Page 3 "Quality & Workflow" (2: project_mode +
	// worktree_auto_create — Issue 3) + Page 3 "Audit & MCP" (6: audit_model +
	// 3 audit_gate + codex_audit_enabled + mcp_tools_opt_in — SPEC-MOAI-MCP-
	// SERVER-001 M4) + Autonomy page (1, autonomy_tier — SPEC-AUTONOMY-TIERS-001
	// AC-001) = 14.
	if got != 14 {
		t.Errorf("TotalVisibleQuestions = %d, want 14 (3 Basic + 2 Model & Report + 2 Quality & Workflow + 6 Audit & MCP + 1 Autonomy)", got)
	}
	// Page 3 "Quality & Workflow" still has project_mode + worktree_auto_create
	// (Issue 3). The M4 audit selection is a separate group ("Audit & MCP").
	n := 0
	for _, q := range FilteredQuestions(all, res) {
		if q.Group == "Quality & Workflow" {
			n++
		}
	}
	if n != 2 {
		t.Errorf("visible Quality & Workflow questions = %d, want 2", n)
	}
	// The M4 audit selection is its own group of 6.
	a := 0
	for _, q := range FilteredQuestions(all, res) {
		if q.Group == "Audit & MCP" {
			a++
		}
	}
	if a != 6 {
		t.Errorf("visible Audit & MCP questions = %d, want 6 (SPEC-MOAI-MCP-SERVER-001 M4)", a)
	}
}
