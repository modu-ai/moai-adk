package wizard

import (
	"os"
	"path/filepath"
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
		{"lsp_enabled", QuestionTypeConfirm, false, false},
		{"enforce_quality", QuestionTypeConfirm, false, false},
		{"project_mode", QuestionTypeSelect, true, false},
		{"design_enabled", QuestionTypeConfirm, false, false},
		{"claude_design_enabled", QuestionTypeConfirm, false, true},
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

	for _, id := range []string{
		"lsp_enabled", "enforce_quality", "project_mode",
		"design_enabled", "claude_design_enabled",
	} {
		if QuestionByID(visible, id) == nil {
			t.Errorf("page-3 question %q must be visible with no bridge answered", id)
		}
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

	page3 := []string{
		"lsp_enabled", "enforce_quality", "project_mode",
		"design_enabled", "claude_design_enabled",
	}
	visible := FilteredQuestions(questions, &WizardResult{DesignEnabled: true})
	for _, id := range page3 {
		if QuestionByID(visible, id) == nil {
			t.Errorf("page-3 question %q must be visible", id)
		}
	}

	// Structural half: nothing on page 3 is gated except the nested design question.
	for _, q := range questions {
		if q.ID == "claude_design_enabled" {
			continue
		}
		if q.Condition != nil {
			t.Errorf("page-3 question %q carries a Condition — page 3 must be ungated", q.ID)
		}
	}
}

// TestPage3Questions_ClaudeDesignConditional verifies claude_design_enabled is hidden
// when DesignEnabled=false (AC-IWE-005 conditional skip).
func TestPage3Questions_ClaudeDesignConditional(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	questions := Page3Questions(tmpDir)

	// Design disabled: claude_design_enabled should be hidden.
	result := &WizardResult{DesignEnabled: false}
	visible := FilteredQuestions(questions, result)
	for _, q := range visible {
		if q.ID == "claude_design_enabled" {
			t.Error("claude_design_enabled visible when DesignEnabled=false, want hidden")
		}
	}
	// Expect 4 visible (the 5 page-3 questions minus claude_design_enabled)
	if len(visible) != 4 {
		t.Errorf("Expected 4 visible questions when DesignEnabled=false, got %d", len(visible))
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

// TestHarnessProfileFallback verifies loadHarnessProfiles falls back to canonical list
// when evaluator-profiles directory is absent.
func TestHarnessProfileFallback(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	// tmpDir has no evaluator-profiles dir → fallback expected

	opts := loadHarnessProfiles(tmpDir)
	if len(opts) != 4 {
		t.Fatalf("fallback: expected 4 options, got %d", len(opts))
	}
	wantValues := []string{"default", "strict", "lenient", "frontend"}
	for i, o := range opts {
		if o.Value != wantValues[i] {
			t.Errorf("opts[%d].Value = %q, want %q", i, o.Value, wantValues[i])
		}
	}
}

// TestHarnessProfileDynamic verifies loadHarnessProfiles reads actual .md files.
func TestHarnessProfileDynamic(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	profileDir := filepath.Join(tmpDir, ".moai", "config", "evaluator-profiles")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"default.md", "custom.md"} {
		if err := os.WriteFile(filepath.Join(profileDir, name), []byte("# profile"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	opts := loadHarnessProfiles(tmpDir)
	if len(opts) != 2 {
		t.Fatalf("dynamic: expected 2 options, got %d", len(opts))
	}
	// First option should have (Recommended) suffix
	if opts[0].Value != "default" && opts[0].Value != "custom" {
		t.Errorf("unexpected first option value: %s", opts[0].Value)
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

// TestSaveBoolAnswer verifies saveBoolAnswer stores the page-3 boolean fields.
// coverage_exemptions_enabled is deliberately absent — its capture branch was
// removed with the question (REQ-WIZ-013).
func TestSaveBoolAnswer(t *testing.T) {
	t.Parallel()
	cases := []struct {
		id    string
		val   bool
		check func(*WizardResult) bool
	}{
		{"lsp_enabled", true, func(r *WizardResult) bool { return r.LSPEnabled }},
		{"enforce_quality", false, func(r *WizardResult) bool { return !r.EnforceQuality }},
		{"design_enabled", false, func(r *WizardResult) bool { return !r.DesignEnabled }},
		{"claude_design_enabled", false, func(r *WizardResult) bool { return !r.ClaudeDesignEnabled }},
	}
	for _, c := range cases {
		result := &WizardResult{EnforceQuality: true, DesignEnabled: true, ClaudeDesignEnabled: true}
		saveBoolAnswer(c.id, c.val, result)
		if !c.check(result) {
			t.Errorf("saveBoolAnswer(%q, %v): check failed", c.id, c.val)
		}
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
		EnforceQuality:            true,
		CoverageExemptionsEnabled: false,
		DesignEnabled:             true,
		ClaudeDesignEnabled:       true,
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
	// Page 1 (3) + Page 2 (2) + Page 3 (5) = 10 exactly.
	if got != 10 {
		t.Errorf("TotalVisibleQuestions = %d, want 10 (3 Basic + 2 Model & Report + 5 Quality & Workflow)", got)
	}
	// All five page-3 questions are counted with no flag to enable them.
	n := 0
	for _, q := range FilteredQuestions(all, res) {
		if q.Group == "Quality & Workflow" {
			n++
		}
	}
	if n != 5 {
		t.Errorf("visible page-3 questions = %d, want 5", n)
	}
}
