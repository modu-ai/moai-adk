package loop

import (
	"testing"
)

func TestIsImproved(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		prev *Feedback
		curr *Feedback
		want bool
	}{
		{
			name: "fewer test failures",
			prev: &Feedback{TestsFailed: 5, LintErrors: 3, Coverage: 72.0},
			curr: &Feedback{TestsFailed: 2, LintErrors: 3, Coverage: 72.0},
			want: true,
		},
		{
			name: "fewer lint errors",
			prev: &Feedback{TestsFailed: 2, LintErrors: 3, Coverage: 72.0},
			curr: &Feedback{TestsFailed: 2, LintErrors: 1, Coverage: 72.0},
			want: true,
		},
		{
			name: "higher coverage",
			prev: &Feedback{TestsFailed: 2, LintErrors: 1, Coverage: 72.0},
			curr: &Feedback{TestsFailed: 2, LintErrors: 1, Coverage: 80.0},
			want: true,
		},
		{
			name: "all metrics improved",
			prev: &Feedback{TestsFailed: 5, LintErrors: 3, Coverage: 72.0},
			curr: &Feedback{TestsFailed: 2, LintErrors: 1, Coverage: 80.0},
			want: true,
		},
		{
			name: "no change (stagnant)",
			prev: &Feedback{TestsFailed: 2, LintErrors: 1, Coverage: 80.0},
			curr: &Feedback{TestsFailed: 2, LintErrors: 1, Coverage: 80.0},
			want: false,
		},
		{
			name: "regression in all metrics",
			prev: &Feedback{TestsFailed: 2, LintErrors: 1, Coverage: 80.0},
			curr: &Feedback{TestsFailed: 5, LintErrors: 3, Coverage: 72.0},
			want: false,
		},
		{
			name: "nil prev",
			prev: nil,
			curr: &Feedback{TestsFailed: 0, LintErrors: 0, Coverage: 90.0},
			want: false,
		},
		{
			name: "nil curr",
			prev: &Feedback{TestsFailed: 5, LintErrors: 3, Coverage: 72.0},
			curr: nil,
			want: false,
		},
		{
			name: "both nil",
			prev: nil,
			curr: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsImproved(tt.prev, tt.curr)
			if got != tt.want {
				t.Errorf("IsImproved() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsStagnant(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		prev *Feedback
		curr *Feedback
		want bool
	}{
		{
			name: "identical metrics",
			prev: &Feedback{TestsFailed: 2, LintErrors: 1, Coverage: 80.0},
			curr: &Feedback{TestsFailed: 2, LintErrors: 1, Coverage: 80.0},
			want: true,
		},
		{
			name: "zero metrics identical",
			prev: &Feedback{TestsFailed: 0, LintErrors: 0, Coverage: 0.0},
			curr: &Feedback{TestsFailed: 0, LintErrors: 0, Coverage: 0.0},
			want: true,
		},
		{
			name: "test failures changed",
			prev: &Feedback{TestsFailed: 2, LintErrors: 1, Coverage: 80.0},
			curr: &Feedback{TestsFailed: 1, LintErrors: 1, Coverage: 80.0},
			want: false,
		},
		{
			name: "lint errors changed",
			prev: &Feedback{TestsFailed: 2, LintErrors: 1, Coverage: 80.0},
			curr: &Feedback{TestsFailed: 2, LintErrors: 0, Coverage: 80.0},
			want: false,
		},
		{
			name: "coverage changed",
			prev: &Feedback{TestsFailed: 2, LintErrors: 1, Coverage: 80.0},
			curr: &Feedback{TestsFailed: 2, LintErrors: 1, Coverage: 82.0},
			want: false,
		},
		{
			name: "nil prev",
			prev: nil,
			curr: &Feedback{TestsFailed: 0, LintErrors: 0, Coverage: 90.0},
			want: false,
		},
		{
			name: "nil curr",
			prev: &Feedback{TestsFailed: 0, LintErrors: 0, Coverage: 90.0},
			curr: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsStagnant(tt.prev, tt.curr)
			if got != tt.want {
				t.Errorf("IsStagnant() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMeetsQualityGate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fb   *Feedback
		want bool
	}{
		{
			name: "all criteria met (87.5%)",
			fb:   &Feedback{TestsFailed: 0, LintErrors: 0, BuildSuccess: true, Coverage: 87.5},
			want: true,
		},
		{
			name: "all criteria met (exactly 85%)",
			fb:   &Feedback{TestsFailed: 0, LintErrors: 0, BuildSuccess: true, Coverage: 85.0},
			want: true,
		},
		{
			name: "all criteria met (92.3%)",
			fb:   &Feedback{TestsFailed: 0, LintErrors: 0, BuildSuccess: true, Coverage: 92.3},
			want: true,
		},
		{
			name: "coverage below 85%",
			fb:   &Feedback{TestsFailed: 0, LintErrors: 0, BuildSuccess: true, Coverage: 82.0},
			want: false,
		},
		{
			name: "test failures present",
			fb:   &Feedback{TestsFailed: 1, LintErrors: 0, BuildSuccess: true, Coverage: 90.0},
			want: false,
		},
		{
			name: "lint errors present",
			fb:   &Feedback{TestsFailed: 0, LintErrors: 1, BuildSuccess: true, Coverage: 90.0},
			want: false,
		},
		{
			name: "build failed",
			fb:   &Feedback{TestsFailed: 0, LintErrors: 0, BuildSuccess: false, Coverage: 90.0},
			want: false,
		},
		{
			name: "all bad",
			fb:   &Feedback{TestsFailed: 5, LintErrors: 3, BuildSuccess: false, Coverage: 50.0},
			want: false,
		},
		{
			name: "nil feedback",
			fb:   nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MeetsQualityGate(tt.fb)
			if got != tt.want {
				t.Errorf("MeetsQualityGate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindPreviousReviewFeedback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		feedbacks        []Feedback
		currentIteration int
		wantNil          bool
		wantIteration    int
	}{
		{
			name:             "empty feedback list",
			feedbacks:        nil,
			currentIteration: 2,
			wantNil:          true,
		},
		{
			name: "no review feedback",
			feedbacks: []Feedback{
				{Phase: PhaseAnalyze, Iteration: 1},
				{Phase: PhaseImplement, Iteration: 1},
			},
			currentIteration: 2,
			wantNil:          true,
		},
		{
			name: "find previous review",
			feedbacks: []Feedback{
				{Phase: PhaseAnalyze, Iteration: 1},
				{Phase: PhaseReview, Iteration: 1},
				{Phase: PhaseAnalyze, Iteration: 2},
				{Phase: PhaseReview, Iteration: 2},
			},
			currentIteration: 2,
			wantNil:          false,
			wantIteration:    1,
		},
		{
			name: "skip current iteration review",
			feedbacks: []Feedback{
				{Phase: PhaseReview, Iteration: 1},
				{Phase: PhaseReview, Iteration: 2},
				{Phase: PhaseReview, Iteration: 3},
			},
			currentIteration: 3,
			wantNil:          false,
			wantIteration:    2,
		},
		{
			name: "first iteration has no previous",
			feedbacks: []Feedback{
				{Phase: PhaseReview, Iteration: 1},
			},
			currentIteration: 1,
			wantNil:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FindPreviousReviewFeedback(tt.feedbacks, tt.currentIteration)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got feedback with iteration %d", got.Iteration)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil feedback, got nil")
			}
			if got.Iteration != tt.wantIteration {
				t.Errorf("Iteration = %d, want %d", got.Iteration, tt.wantIteration)
			}
		})
	}
}
