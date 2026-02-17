package github

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/quality"
)

// mockPRReviewer implements PRReviewer for testing.
type mockPRReviewer struct {
	report *ReviewReport
	err    error
}

func (m *mockPRReviewer) Review(_ context.Context, _ int, _ string, _ *ReviewInput) (*ReviewReport, error) {
	return m.report, m.err
}

// mustNewPRMerger is a test helper that calls NewPRMerger and fails the test on error.
func mustNewPRMerger(t *testing.T, gh GHClient, rev PRReviewer, logger *slog.Logger) *prMerger {
	t.Helper()
	m, err := NewPRMerger(gh, rev, logger)
	if err != nil {
		t.Fatalf("NewPRMerger() error = %v", err)
	}
	return m
}

func TestNewPRMerger(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{}
	rev := &mockPRReviewer{}
	m, err := NewPRMerger(gh, rev, nil)
	if err != nil {
		t.Fatalf("NewPRMerger() error = %v", err)
	}
	if m == nil {
		t.Fatal("NewPRMerger returned nil")
	}
}

func TestNewPRMerger_NilGHClient(t *testing.T) {
	t.Parallel()

	_, err := NewPRMerger(nil, &mockPRReviewer{}, nil)
	if !errors.Is(err, ErrNilGHClient) {
		t.Errorf("NewPRMerger(nil gh) error = %v, want ErrNilGHClient", err)
	}
}

func TestNewPRMerger_NilReviewer(t *testing.T) {
	t.Parallel()

	_, err := NewPRMerger(&mockGHClient{}, nil, nil)
	if !errors.Is(err, ErrNilReviewer) {
		t.Errorf("NewPRMerger(nil reviewer) error = %v, want ErrNilReviewer", err)
	}
}

func TestMerge_AllConditionsMet(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number:    100,
			State:     "OPEN",
			Mergeable: "MERGEABLE",
		},
		prChecksResult: &CheckStatus{Overall: CheckPass},
	}
	rev := &mockPRReviewer{
		report: &ReviewReport{
			PRNumber:      100,
			Decision:      ReviewApprove,
			QualityReport: &quality.Report{Passed: true, Score: 1.0},
		},
	}
	merger := mustNewPRMerger(t, gh, rev, nil)

	result, err := merger.Merge(context.Background(), 100, MergeOptions{
		AutoMerge:     true,
		Method:        MergeMethodSquash,
		RequireReview: true,
		RequireChecks: true,
		SpecID:        "SPEC-ISSUE-100",
	})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}
	if !result.Merged {
		t.Error("Merged = false, want true")
	}
	if result.PRNumber != 100 {
		t.Errorf("PRNumber = %d, want 100", result.PRNumber)
	}
	if result.Method != MergeMethodSquash {
		t.Errorf("Method = %q, want %q", result.Method, MergeMethodSquash)
	}
	if !gh.prMergeCalled {
		t.Error("gh.PRMerge not called")
	}
	if gh.prMergeMethod != MergeMethodSquash {
		t.Errorf("gh merge method = %q, want %q", gh.prMergeMethod, MergeMethodSquash)
	}
}

func TestMerge_AutoMergeNotRequested(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{}
	rev := &mockPRReviewer{}
	merger := mustNewPRMerger(t, gh, rev, nil)

	_, err := merger.Merge(context.Background(), 101, MergeOptions{
		AutoMerge: false,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrAutoMergeNotRequested) {
		t.Errorf("error = %v, want ErrAutoMergeNotRequested", err)
	}
}

func TestMerge_ReviewNotApproved(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number:    102,
			State:     "OPEN",
			Mergeable: "MERGEABLE",
		},
		prChecksResult: &CheckStatus{Overall: CheckPass},
	}
	rev := &mockPRReviewer{
		report: &ReviewReport{
			PRNumber:      102,
			Decision:      ReviewRequestChanges,
			QualityReport: &quality.Report{Passed: false, Score: 0.5},
		},
	}
	merger := mustNewPRMerger(t, gh, rev, nil)

	_, err := merger.Merge(context.Background(), 102, MergeOptions{
		AutoMerge:     true,
		RequireReview: true,
		RequireChecks: true,
		SpecID:        "SPEC-ISSUE-102",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrMergeBlocked) {
		t.Errorf("error = %v, want ErrMergeBlocked", err)
	}
}

func TestMerge_CIFailed(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number:    103,
			State:     "OPEN",
			Mergeable: "MERGEABLE",
		},
		prChecksResult: &CheckStatus{
			Overall: CheckFail,
			Checks: []Check{
				{Name: "test", Status: "completed", Conclusion: "failure"},
			},
		},
	}
	rev := &mockPRReviewer{
		report: &ReviewReport{
			PRNumber:      103,
			Decision:      ReviewApprove,
			QualityReport: &quality.Report{Passed: true, Score: 1.0},
		},
	}
	merger := mustNewPRMerger(t, gh, rev, nil)

	_, err := merger.Merge(context.Background(), 103, MergeOptions{
		AutoMerge:     true,
		RequireReview: true,
		RequireChecks: true,
		SpecID:        "SPEC-ISSUE-103",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrMergeBlocked) {
		t.Errorf("error = %v, want ErrMergeBlocked", err)
	}
}

func TestMerge_MergeConflicts(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number:    104,
			State:     "OPEN",
			Mergeable: "CONFLICTING",
		},
		prChecksResult: &CheckStatus{Overall: CheckPass},
	}
	rev := &mockPRReviewer{
		report: &ReviewReport{
			PRNumber:      104,
			Decision:      ReviewApprove,
			QualityReport: &quality.Report{Passed: true, Score: 1.0},
		},
	}
	merger := mustNewPRMerger(t, gh, rev, nil)

	_, err := merger.Merge(context.Background(), 104, MergeOptions{
		AutoMerge:     true,
		RequireReview: true,
		RequireChecks: true,
		SpecID:        "SPEC-ISSUE-104",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrMergeBlocked) {
		t.Errorf("error = %v, want ErrMergeBlocked", err)
	}
}

func TestMerge_DefaultMergeMethod(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number:    105,
			State:     "OPEN",
			Mergeable: "MERGEABLE",
		},
	}
	rev := &mockPRReviewer{
		report: &ReviewReport{
			PRNumber:      105,
			Decision:      ReviewApprove,
			QualityReport: &quality.Report{Passed: true, Score: 1.0},
		},
	}
	merger := mustNewPRMerger(t, gh, rev, nil)

	result, err := merger.Merge(context.Background(), 105, MergeOptions{
		AutoMerge: true,
		// Method intentionally left empty to test default.
	})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}
	if result.Method != MergeMethodMerge {
		t.Errorf("Method = %q, want %q (default)", result.Method, MergeMethodMerge)
	}
}

func TestMerge_RebaseMethod(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number:    106,
			State:     "OPEN",
			Mergeable: "MERGEABLE",
		},
	}
	rev := &mockPRReviewer{}
	merger := mustNewPRMerger(t, gh, rev, nil)

	result, err := merger.Merge(context.Background(), 106, MergeOptions{
		AutoMerge: true,
		Method:    MergeMethodRebase,
	})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}
	if gh.prMergeMethod != MergeMethodRebase {
		t.Errorf("gh merge method = %q, want %q", gh.prMergeMethod, MergeMethodRebase)
	}
	if result.Method != MergeMethodRebase {
		t.Errorf("result.Method = %q, want %q", result.Method, MergeMethodRebase)
	}
}

func TestCheckPrerequisites_AllMet(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number:    110,
			State:     "OPEN",
			Mergeable: "MERGEABLE",
		},
		prChecksResult: &CheckStatus{Overall: CheckPass},
	}
	rev := &mockPRReviewer{
		report: &ReviewReport{
			PRNumber:      110,
			Decision:      ReviewApprove,
			QualityReport: &quality.Report{Passed: true, Score: 1.0},
		},
	}
	merger := mustNewPRMerger(t, gh, rev, nil)

	prereqs, err := merger.CheckPrerequisites(context.Background(), 110, MergeOptions{
		AutoMerge:     true,
		RequireReview: true,
		RequireChecks: true,
		SpecID:        "SPEC-ISSUE-110",
	})
	if err != nil {
		t.Fatalf("CheckPrerequisites() error = %v", err)
	}
	if !prereqs.AllMet {
		t.Errorf("AllMet = false, want true; reasons: %v", prereqs.FailureReasons)
	}
	if !prereqs.AutoMergeFlag {
		t.Error("AutoMergeFlag = false, want true")
	}
	if !prereqs.ReviewApproved {
		t.Error("ReviewApproved = false, want true")
	}
	if !prereqs.ChecksPassed {
		t.Error("ChecksPassed = false, want true")
	}
	if !prereqs.QualityPassed {
		t.Error("QualityPassed = false, want true")
	}
	if !prereqs.Mergeable {
		t.Error("Mergeable = false, want true")
	}
}

func TestCheckPrerequisites_PartiallyMet(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number:    111,
			State:     "OPEN",
			Mergeable: "MERGEABLE",
		},
		prChecksResult: &CheckStatus{Overall: CheckFail},
	}
	rev := &mockPRReviewer{
		report: &ReviewReport{
			PRNumber:      111,
			Decision:      ReviewApprove,
			QualityReport: &quality.Report{Passed: true, Score: 1.0},
		},
	}
	merger := mustNewPRMerger(t, gh, rev, nil)

	prereqs, err := merger.CheckPrerequisites(context.Background(), 111, MergeOptions{
		AutoMerge:     true,
		RequireReview: true,
		RequireChecks: true,
		SpecID:        "SPEC-ISSUE-111",
	})
	if err != nil {
		t.Fatalf("CheckPrerequisites() error = %v", err)
	}
	if prereqs.AllMet {
		t.Error("AllMet = true, want false")
	}
	if prereqs.ChecksPassed {
		t.Error("ChecksPassed = true, want false")
	}
	if len(prereqs.FailureReasons) == 0 {
		t.Error("FailureReasons is empty, want at least one")
	}
}

func TestCheckPrerequisites_NoReviewRequired(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number:    112,
			State:     "OPEN",
			Mergeable: "MERGEABLE",
		},
	}
	rev := &mockPRReviewer{}
	merger := mustNewPRMerger(t, gh, rev, nil)

	prereqs, err := merger.CheckPrerequisites(context.Background(), 112, MergeOptions{
		AutoMerge:     true,
		RequireReview: false,
		RequireChecks: false,
	})
	if err != nil {
		t.Fatalf("CheckPrerequisites() error = %v", err)
	}
	if !prereqs.AllMet {
		t.Errorf("AllMet = false, want true; reasons: %v", prereqs.FailureReasons)
	}
}

func TestCheckPrerequisites_PRViewError(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewErr: ErrPRNotFound,
	}
	rev := &mockPRReviewer{}
	merger := mustNewPRMerger(t, gh, rev, nil)

	_, err := merger.CheckPrerequisites(context.Background(), 999, MergeOptions{
		AutoMerge: true,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrPRNotFound) {
		t.Errorf("error = %v, want ErrPRNotFound", err)
	}
}

func TestMerge_PRMergeError(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number:    200,
			State:     "OPEN",
			Mergeable: "MERGEABLE",
		},
		prMergeErr: errors.New("merge API error"),
	}
	rev := &mockPRReviewer{
		report: &ReviewReport{
			PRNumber:      200,
			Decision:      ReviewApprove,
			QualityReport: &quality.Report{Passed: true, Score: 1.0},
		},
	}
	merger := mustNewPRMerger(t, gh, rev, nil)

	_, err := merger.Merge(context.Background(), 200, MergeOptions{
		AutoMerge:     true,
		RequireReview: true,
		SpecID:        "SPEC-ISSUE-200",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheckPrerequisites_ReviewError(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number:    201,
			State:     "OPEN",
			Mergeable: "MERGEABLE",
		},
		prChecksResult: &CheckStatus{Overall: CheckPass},
	}
	rev := &mockPRReviewer{
		err: errors.New("review service unavailable"),
	}
	merger := mustNewPRMerger(t, gh, rev, nil)

	prereqs, err := merger.CheckPrerequisites(context.Background(), 201, MergeOptions{
		AutoMerge:     true,
		RequireReview: true,
		RequireChecks: true,
		SpecID:        "SPEC-ISSUE-201",
	})
	if err != nil {
		t.Fatalf("CheckPrerequisites() error = %v", err)
	}
	if prereqs.AllMet {
		t.Error("AllMet = true, want false (review error)")
	}
	if prereqs.ReviewApproved {
		t.Error("ReviewApproved = true, want false")
	}
	if len(prereqs.FailureReasons) == 0 {
		t.Error("FailureReasons is empty")
	}
}

func TestCheckPrerequisites_ChecksError(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number:    202,
			State:     "OPEN",
			Mergeable: "MERGEABLE",
		},
		prChecksErr: errors.New("checks API error"),
	}
	rev := &mockPRReviewer{
		report: &ReviewReport{
			PRNumber:      202,
			Decision:      ReviewApprove,
			QualityReport: &quality.Report{Passed: true, Score: 1.0},
		},
	}
	merger := mustNewPRMerger(t, gh, rev, nil)

	prereqs, err := merger.CheckPrerequisites(context.Background(), 202, MergeOptions{
		AutoMerge:     true,
		RequireReview: true,
		RequireChecks: true,
		SpecID:        "SPEC-ISSUE-202",
	})
	if err != nil {
		t.Fatalf("CheckPrerequisites() error = %v", err)
	}
	if prereqs.AllMet {
		t.Error("AllMet = true, want false (checks error)")
	}
	if prereqs.ChecksPassed {
		t.Error("ChecksPassed = true, want false")
	}
	if len(prereqs.FailureReasons) == 0 {
		t.Error("FailureReasons is empty")
	}
}

func TestCheckPrerequisites_UnknownMergeability(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number:    203,
			State:     "OPEN",
			Mergeable: "UNKNOWN",
		},
	}
	rev := &mockPRReviewer{}
	merger := mustNewPRMerger(t, gh, rev, nil)

	prereqs, err := merger.CheckPrerequisites(context.Background(), 203, MergeOptions{
		AutoMerge: true,
	})
	if err != nil {
		t.Fatalf("CheckPrerequisites() error = %v", err)
	}
	if !prereqs.Mergeable {
		t.Error("Mergeable = false, want true (UNKNOWN treated as potentially mergeable)")
	}
}
