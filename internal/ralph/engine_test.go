package ralph

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/loop"
	"github.com/modu-ai/moai-adk/internal/lsp/gopls"
)

func TestRalphEngine_Decide(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		cfg           config.RalphConfig
		state         *loop.LoopState
		feedback      *loop.Feedback
		wantAction    string
		wantConverged bool
		wantReasonSub string
		wantErr       bool
	}{
		{
			name: "max iterations reached (iteration=5, max=5)",
			cfg:  config.RalphConfig{MaxIterations: 5, AutoConverge: true, HumanReview: false},
			state: &loop.LoopState{
				SpecID:    "SPEC-TEST",
				Phase:     loop.PhaseReview,
				Iteration: 5,
				MaxIter:   5,
			},
			feedback:      &loop.Feedback{TestsFailed: 1, LintErrors: 0, BuildSuccess: true, Coverage: 80.0},
			wantAction:    loop.ActionAbort,
			wantConverged: false,
			wantReasonSub: "max iterations",
		},
		{
			name: "perfect success converges immediately",
			cfg:  config.RalphConfig{MaxIterations: 5, AutoConverge: true, HumanReview: false},
			state: &loop.LoopState{
				SpecID:    "SPEC-TEST",
				Phase:     loop.PhaseReview,
				Iteration: 1,
				MaxIter:   5,
			},
			feedback:      &loop.Feedback{TestsFailed: 0, LintErrors: 0, BuildSuccess: true, Coverage: 92.3},
			wantAction:    loop.ActionConverge,
			wantConverged: true,
			wantReasonSub: "quality gate",
		},
		{
			name: "perfect success at exactly 85% coverage",
			cfg:  config.RalphConfig{MaxIterations: 5, AutoConverge: true, HumanReview: false},
			state: &loop.LoopState{
				SpecID:    "SPEC-TEST",
				Phase:     loop.PhaseReview,
				Iteration: 2,
				MaxIter:   5,
			},
			feedback:      &loop.Feedback{TestsFailed: 0, LintErrors: 0, BuildSuccess: true, Coverage: 85.0},
			wantAction:    loop.ActionConverge,
			wantConverged: true,
			wantReasonSub: "quality gate",
		},
		{
			name: "stagnation detected with auto_converge on",
			cfg:  config.RalphConfig{MaxIterations: 5, AutoConverge: true, HumanReview: false},
			state: &loop.LoopState{
				SpecID:    "SPEC-TEST",
				Phase:     loop.PhaseReview,
				Iteration: 2,
				MaxIter:   5,
				Feedback: []loop.Feedback{
					{Phase: loop.PhaseReview, Iteration: 1, TestsFailed: 2, LintErrors: 1, Coverage: 78.5},
				},
			},
			feedback:      &loop.Feedback{Phase: loop.PhaseReview, Iteration: 2, TestsFailed: 2, LintErrors: 1, Coverage: 78.5},
			wantAction:    loop.ActionConverge,
			wantConverged: true,
			wantReasonSub: "stagnant",
		},
		{
			name: "stagnation ignored with auto_converge off",
			cfg:  config.RalphConfig{MaxIterations: 5, AutoConverge: false, HumanReview: false},
			state: &loop.LoopState{
				SpecID:    "SPEC-TEST",
				Phase:     loop.PhaseReview,
				Iteration: 2,
				MaxIter:   5,
				Feedback: []loop.Feedback{
					{Phase: loop.PhaseReview, Iteration: 1, TestsFailed: 2, LintErrors: 1, Coverage: 78.5},
				},
			},
			feedback:      &loop.Feedback{Phase: loop.PhaseReview, Iteration: 2, TestsFailed: 2, LintErrors: 1, Coverage: 78.5},
			wantAction:    loop.ActionContinue,
			wantConverged: false,
		},
		{
			name: "human review requested at review phase",
			cfg:  config.RalphConfig{MaxIterations: 5, AutoConverge: true, HumanReview: true},
			state: &loop.LoopState{
				SpecID:    "SPEC-TEST",
				Phase:     loop.PhaseReview,
				Iteration: 1,
				MaxIter:   5,
			},
			feedback:      &loop.Feedback{TestsFailed: 1, LintErrors: 0, BuildSuccess: true, Coverage: 80.0},
			wantAction:    loop.ActionRequestReview,
			wantConverged: false,
			wantReasonSub: "human review",
		},
		{
			name: "human review disabled continues",
			cfg:  config.RalphConfig{MaxIterations: 5, AutoConverge: false, HumanReview: false},
			state: &loop.LoopState{
				SpecID:    "SPEC-TEST",
				Phase:     loop.PhaseReview,
				Iteration: 1,
				MaxIter:   5,
			},
			feedback:      &loop.Feedback{TestsFailed: 1, LintErrors: 0, BuildSuccess: true, Coverage: 80.0},
			wantAction:    loop.ActionContinue,
			wantConverged: false,
		},
		{
			name: "continue when improvement detected",
			cfg:  config.RalphConfig{MaxIterations: 5, AutoConverge: true, HumanReview: false},
			state: &loop.LoopState{
				SpecID:    "SPEC-TEST",
				Phase:     loop.PhaseReview,
				Iteration: 2,
				MaxIter:   5,
				Feedback: []loop.Feedback{
					{Phase: loop.PhaseReview, Iteration: 1, TestsFailed: 5, LintErrors: 3, Coverage: 72.0},
				},
			},
			feedback:      &loop.Feedback{Phase: loop.PhaseReview, Iteration: 2, TestsFailed: 2, LintErrors: 1, Coverage: 80.0},
			wantAction:    loop.ActionContinue,
			wantConverged: false,
		},
		{
			name: "iteration 4 with max 5 continues",
			cfg:  config.RalphConfig{MaxIterations: 5, AutoConverge: false, HumanReview: false},
			state: &loop.LoopState{
				SpecID:    "SPEC-TEST",
				Phase:     loop.PhaseReview,
				Iteration: 4,
				MaxIter:   5,
			},
			feedback:      &loop.Feedback{TestsFailed: 1, LintErrors: 0, BuildSuccess: true, Coverage: 80.0},
			wantAction:    loop.ActionContinue,
			wantConverged: false,
		},
		{
			name: "coverage below 85% continues",
			cfg:  config.RalphConfig{MaxIterations: 5, AutoConverge: false, HumanReview: false},
			state: &loop.LoopState{
				SpecID:    "SPEC-TEST",
				Phase:     loop.PhaseReview,
				Iteration: 1,
				MaxIter:   5,
			},
			feedback:      &loop.Feedback{TestsFailed: 0, LintErrors: 0, BuildSuccess: true, Coverage: 82.0},
			wantAction:    loop.ActionContinue,
			wantConverged: false,
		},
		{
			name: "first iteration no stagnation possible",
			cfg:  config.RalphConfig{MaxIterations: 5, AutoConverge: true, HumanReview: false},
			state: &loop.LoopState{
				SpecID:    "SPEC-TEST",
				Phase:     loop.PhaseReview,
				Iteration: 1,
				MaxIter:   5,
			},
			feedback:      &loop.Feedback{TestsFailed: 2, LintErrors: 1, Coverage: 78.5},
			wantAction:    loop.ActionContinue,
			wantConverged: false,
		},
		{
			name:     "nil state returns error",
			cfg:      config.RalphConfig{MaxIterations: 5},
			state:    nil,
			feedback: &loop.Feedback{},
			wantErr:  true,
		},
		{
			name: "nil feedback returns error",
			cfg:  config.RalphConfig{MaxIterations: 5},
			state: &loop.LoopState{
				SpecID:    "SPEC-TEST",
				Phase:     loop.PhaseReview,
				Iteration: 1,
				MaxIter:   5,
			},
			feedback: nil,
			wantErr:  true,
		},
		{
			name: "perfect success overrides human review",
			cfg:  config.RalphConfig{MaxIterations: 5, AutoConverge: true, HumanReview: true},
			state: &loop.LoopState{
				SpecID:    "SPEC-TEST",
				Phase:     loop.PhaseReview,
				Iteration: 1,
				MaxIter:   5,
			},
			feedback:      &loop.Feedback{TestsFailed: 0, LintErrors: 0, BuildSuccess: true, Coverage: 90.0},
			wantAction:    loop.ActionConverge,
			wantConverged: true,
			wantReasonSub: "quality gate",
		},
		{
			name: "max iterations overrides everything",
			cfg:  config.RalphConfig{MaxIterations: 3, AutoConverge: true, HumanReview: true},
			state: &loop.LoopState{
				SpecID:    "SPEC-TEST",
				Phase:     loop.PhaseReview,
				Iteration: 3,
				MaxIter:   3,
			},
			feedback:      &loop.Feedback{TestsFailed: 0, LintErrors: 0, BuildSuccess: true, Coverage: 90.0},
			wantAction:    loop.ActionAbort,
			wantConverged: false,
			wantReasonSub: "max iterations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			engine := NewRalphEngine(tt.cfg)
			decision, err := engine.Decide(context.Background(), tt.state, tt.feedback)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if decision.Action != tt.wantAction {
				t.Errorf("Action = %q, want %q", decision.Action, tt.wantAction)
			}
			if decision.Converged != tt.wantConverged {
				t.Errorf("Converged = %v, want %v", decision.Converged, tt.wantConverged)
			}
			if tt.wantReasonSub != "" && !strings.Contains(decision.Reason, tt.wantReasonSub) {
				t.Errorf("Reason = %q, want to contain %q", decision.Reason, tt.wantReasonSub)
			}
		})
	}
}

// TestClassifyFeedback_DiagnosticError verifies that when Diagnostics contains
// an Error-severity entry, ClassifyFeedback includes ErrorLevelManual.
func TestClassifyFeedback_DiagnosticError(t *testing.T) {
	t.Parallel()

	fb := &loop.Feedback{
		BuildSuccess: true,
		Diagnostics: []gopls.Diagnostic{
			{Severity: gopls.SeverityError, Message: "undeclared name: foo", Source: "compiler"},
		},
	}

	classified := ClassifyFeedback(fb)

	if len(classified) == 0 {
		t.Fatal("ClassifyFeedback returned empty slice despite a diagnostic error")
	}

	maxLevel := MaxErrorLevel(classified)
	if maxLevel < ErrorLevelManual {
		t.Errorf("MaxErrorLevel = %d, expected >= %d (ErrorLevelManual)", maxLevel, ErrorLevelManual)
	}

	// A diagnostic-error classification entry must be present.
	found := false
	for _, ce := range classified {
		if ce.Level == ErrorLevelManual && strings.Contains(ce.Description, "diagnostic") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("could not find an ErrorLevelManual classification entry for the diagnostic error. classified = %v", classified)
	}
}

// TestClassifyFeedback_DiagnosticWarning verifies that when Diagnostics
// contains only Warning-severity entries, ClassifyFeedback includes
// ErrorLevelLogOnly.
func TestClassifyFeedback_DiagnosticWarning(t *testing.T) {
	t.Parallel()

	fb := &loop.Feedback{
		BuildSuccess: true,
		Diagnostics: []gopls.Diagnostic{
			{Severity: gopls.SeverityWarning, Message: "unused variable", Source: "staticcheck"},
			{Severity: gopls.SeverityWarning, Message: "deprecated use", Source: "gopls"},
		},
	}

	classified := ClassifyFeedback(fb)

	if len(classified) == 0 {
		t.Fatal("ClassifyFeedback returned empty slice despite a diagnostic warning")
	}

	found := false
	for _, ce := range classified {
		if ce.Level == ErrorLevelLogOnly && strings.Contains(ce.Description, "diagnostic") {
			found = true
			if ce.Count != 2 {
				t.Errorf("warning Count = %d, expected 2", ce.Count)
			}
			break
		}
	}
	if !found {
		t.Errorf("could not find an ErrorLevelLogOnly classification entry for the diagnostic warning. classified = %v", classified)
	}
}

// TestClassifyFeedback_DiagnosticMixed verifies that when Error + Warning
// entries coexist, the final MaxErrorLevel is ErrorLevelManual.
func TestClassifyFeedback_DiagnosticMixed(t *testing.T) {
	t.Parallel()

	fb := &loop.Feedback{
		BuildSuccess: true,
		Diagnostics: []gopls.Diagnostic{
			{Severity: gopls.SeverityError, Message: "type error"},
			{Severity: gopls.SeverityWarning, Message: "lint warning"},
		},
	}

	classified := ClassifyFeedback(fb)
	maxLevel := MaxErrorLevel(classified)

	if maxLevel < ErrorLevelManual {
		t.Errorf("MaxErrorLevel = %d, expected >= %d when Error+Warning coexist", maxLevel, ErrorLevelManual)
	}
}

// TestClassifyFeedback_NilDiagnostics verifies that when Diagnostics is nil
// the legacy integer-based classification continues to work (fallback guarantee).
func TestClassifyFeedback_NilDiagnostics(t *testing.T) {
	t.Parallel()

	fb := &loop.Feedback{
		TestsFailed:  3,
		LintErrors:   2,
		BuildSuccess: true,
	}

	classified := ClassifyFeedback(fb)

	if len(classified) == 0 {
		t.Fatal("ClassifyFeedback returned empty slice despite legacy errors")
	}

	// Legacy classification: lint=AutoFix, tests=LogOnly (<= 5).
	foundLint := false
	foundTest := false
	for _, ce := range classified {
		if ce.Level == ErrorLevelAutoFix && strings.Contains(ce.Description, "lint") {
			foundLint = true
		}
		if ce.Level == ErrorLevelLogOnly && strings.Contains(ce.Description, "test") {
			foundTest = true
		}
	}
	if !foundLint {
		t.Errorf("could not find a lint-error classification. classified = %v", classified)
	}
	if !foundTest {
		t.Errorf("could not find a test-failure classification. classified = %v", classified)
	}
}

// TestClassifyFeedback_EmptyDiagnostics verifies that no extra classification
// entries are produced when Diagnostics is an empty slice.
func TestClassifyFeedback_EmptyDiagnostics(t *testing.T) {
	t.Parallel()

	fb := &loop.Feedback{
		BuildSuccess: true,
		Diagnostics:  []gopls.Diagnostic{}, // empty slice
	}

	classified := ClassifyFeedback(fb)
	// With empty diagnostics, no diagnostic classification should be added.
	for _, ce := range classified {
		if strings.Contains(ce.Description, "diagnostic") {
			t.Errorf("a diagnostic classification was produced despite empty diagnostics: %v", ce)
		}
	}
}
