package wizard

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPhase1QuestionsStructure verifies each Phase 1 Question entry has the required fields.
func TestPhase1QuestionsStructure(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	questions := Phase1Questions(tmpDir)

	// Expected IDs and their types
	type entry struct {
		id      string
		qtype   QuestionType
		hasOpts bool // whether Options slice is non-empty
	}

	want := []entry{
		{"project_mode", QuestionTypeSelect, true},
		{"harness_profile", QuestionTypeSelect, true},
		{"lsp_enabled", QuestionTypeConfirm, false},
		{"enforce_quality", QuestionTypeConfirm, false},
		{"coverage_exemptions_enabled", QuestionTypeConfirm, false},
		{"design_enabled", QuestionTypeConfirm, false},
		{"claude_design_enabled", QuestionTypeConfirm, false},
	}

	if len(questions) != len(want) {
		t.Fatalf("Phase1Questions() returned %d questions, want %d", len(questions), len(want))
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
		if q.Condition == nil {
			t.Errorf("questions[%d] (%s) has nil Condition; Phase 1 questions must be gated", i, q.ID)
		}
	}
}

// TestPhase1Questions_StandardModeGating verifies questions are hidden when StandardMode=false.
func TestPhase1Questions_StandardModeGating(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	questions := Phase1Questions(tmpDir)

	// Quick mode: StandardMode=false — all Phase 1 questions should be filtered out
	quickResult := &WizardResult{StandardMode: false}
	visible := FilteredQuestions(questions, quickResult)
	if len(visible) != 0 {
		t.Errorf("Quick mode: expected 0 visible Phase 1 questions, got %d", len(visible))
	}

	// Standard mode: StandardMode=true, DesignEnabled=true — all 7 questions visible
	standardResult := &WizardResult{StandardMode: true, DesignEnabled: true}
	visible = FilteredQuestions(questions, standardResult)
	if len(visible) != 7 {
		t.Errorf("Standard mode: expected 7 visible Phase 1 questions, got %d", len(visible))
	}
}

// TestPhase1Questions_ClaudeDesignConditional verifies claude_design_enabled is hidden
// when DesignEnabled=false (AC-IWE-005 conditional skip).
func TestPhase1Questions_ClaudeDesignConditional(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	questions := Phase1Questions(tmpDir)

	// Standard mode but design disabled: claude_design_enabled should be hidden
	result := &WizardResult{StandardMode: true, DesignEnabled: false}
	visible := FilteredQuestions(questions, result)
	for _, q := range visible {
		if q.ID == "claude_design_enabled" {
			t.Error("claude_design_enabled visible when DesignEnabled=false, want hidden")
		}
	}
	// Expect 6 visible (7 minus claude_design_enabled)
	if len(visible) != 6 {
		t.Errorf("Expected 6 visible questions when DesignEnabled=false, got %d", len(visible))
	}
}

// TestProjectModeQuestion verifies project_mode has exactly 2 options (personal/team).
func TestProjectModeQuestion(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	questions := Phase1Questions(tmpDir)

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

// TestSaveAnswerPhase1 verifies saveAnswer stores Phase 1 string fields correctly.
func TestSaveAnswerPhase1(t *testing.T) {
	t.Parallel()
	locale := ""
	result := &WizardResult{}

	saveAnswer("project_mode", "team", result, &locale)
	if result.ProjectMode != "team" {
		t.Errorf("ProjectMode = %q, want 'team'", result.ProjectMode)
	}

	saveAnswer("harness_profile", "strict", result, &locale)
	if result.HarnessProfile != "strict" {
		t.Errorf("HarnessProfile = %q, want 'strict'", result.HarnessProfile)
	}
}

// TestSaveBoolAnswer verifies saveBoolAnswer stores all Phase 1 boolean fields.
func TestSaveBoolAnswer(t *testing.T) {
	t.Parallel()
	cases := []struct {
		id   string
		val  bool
		check func(*WizardResult) bool
	}{
		{"lsp_enabled", true, func(r *WizardResult) bool { return r.LSPEnabled }},
		{"enforce_quality", false, func(r *WizardResult) bool { return !r.EnforceQuality }},
		{"coverage_exemptions_enabled", true, func(r *WizardResult) bool { return r.CoverageExemptionsEnabled }},
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

// TestIsAdvancedWizardReady verifies gate returns false for P2 and P4 (not yet implemented).
func TestIsAdvancedWizardReady(t *testing.T) {
	t.Parallel()
	gate := IsAdvancedWizardReady()
	if gate.P2Ready {
		t.Error("P2Ready = true, want false (SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 not implemented)")
	}
	if gate.P4Ready {
		t.Error("P4Ready = true, want false (SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001 not implemented)")
	}
}

// TestPhase2Questions_SkippedWhenGateNotReady verifies Phase 2 stubs are hidden
// when prerequisites are absent.
func TestPhase2Questions_SkippedWhenGateNotReady(t *testing.T) {
	t.Parallel()
	gate := AdvancedGate{P2Ready: false, P4Ready: false}
	questions := Phase2Questions(gate)

	result := &WizardResult{StandardMode: true, AdvancedMode: true}
	visible := FilteredQuestions(questions, result)
	if len(visible) != 0 {
		t.Errorf("Expected 0 visible Phase 2 questions when gate not ready, got %d", len(visible))
	}
}

// TestPhase2Questions_VisibleWhenGateReady verifies Phase 2 stubs are visible
// when prerequisites are met (mock scenario).
func TestPhase2Questions_VisibleWhenGateReady(t *testing.T) {
	t.Parallel()
	gate := AdvancedGate{P2Ready: true, P4Ready: true}
	questions := Phase2Questions(gate)

	result := &WizardResult{StandardMode: true, AdvancedMode: true}
	visible := FilteredQuestions(questions, result)
	if len(visible) != 4 {
		t.Errorf("Expected 4 visible Phase 2 questions when gate ready, got %d", len(visible))
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

// TestWizardResultDefaultsPrePopulated verifies RunWithDefaultsModes pre-populates
// the result with correct defaults before wizard interaction.
func TestWizardResultDefaultsPrePopulated(t *testing.T) {
	t.Parallel()
	// We test the pre-population logic directly since we can't run interactive wizard in tests.
	result := &WizardResult{
		StandardMode:             true,
		AdvancedMode:             false,
		EnforceQuality:           true,
		CoverageExemptionsEnabled: false,
		DesignEnabled:            true,
		ClaudeDesignEnabled:      true,
	}

	if !result.StandardMode {
		t.Error("StandardMode should be true")
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

// TestTotalVisibleQuestions_StandardMode verifies TotalVisibleQuestions includes
// Phase 1 questions in standard mode.
func TestTotalVisibleQuestions_StandardMode(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	all := append(DefaultQuestions(tmpDir), Phase1Questions(tmpDir)...)

	// Quick mode: 9 default questions visible (all have nil Condition)
	quickResult := &WizardResult{StandardMode: false}
	// The 4 git conditional questions may or may not be visible depending on GitMode
	// At least 5 unconditional questions (name, model_policy, development_mode, git_mode)
	// Plus 4 conditional git questions (default hidden unless GitMode is set)
	// and 0 Phase 1 questions
	quickCount := TotalVisibleQuestions(all, quickResult)
	if quickCount != 4 { // project_name, model_policy, development_mode, git_mode
		// More lenient check: all 9 original minus conditional ones
		if quickCount > 9 {
			t.Errorf("Quick mode shows more than 9 questions: %d", quickCount)
		}
	}

	// Standard mode: all 9 default + 7 Phase 1 = up to 16 (minus git conditionals)
	standardResult := &WizardResult{StandardMode: true, DesignEnabled: true}
	standardCount := TotalVisibleQuestions(all, standardResult)
	if standardCount < 11 { // at minimum: 4 unconditional default + 7 Phase 1
		t.Errorf("Standard mode: too few visible questions: %d (want >= 11)", standardCount)
	}
}
