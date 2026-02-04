package ralph

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/loop"
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
