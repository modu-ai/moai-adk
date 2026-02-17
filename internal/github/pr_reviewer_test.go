package github

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/quality"
)

// mockQualityGate implements quality.Gate for testing.
type mockQualityGate struct {
	report *quality.Report
	err    error
}

func (m *mockQualityGate) Validate(_ context.Context) (*quality.Report, error) {
	return m.report, m.err
}

func (m *mockQualityGate) ValidatePrinciple(_ context.Context, _ string) (*quality.PrincipleResult, error) {
	return nil, nil
}

// mustNewPRReviewer is a test helper that calls NewPRReviewer and fails the test on error.
func mustNewPRReviewer(t *testing.T, gh GHClient, gate quality.Gate, logger *slog.Logger) *prReviewer {
	t.Helper()
	r, err := NewPRReviewer(gh, gate, logger)
	if err != nil {
		t.Fatalf("NewPRReviewer() error = %v", err)
	}
	return r
}

func TestNewPRReviewer(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{}
	gate := &mockQualityGate{}
	r, err := NewPRReviewer(gh, gate, nil)
	if err != nil {
		t.Fatalf("NewPRReviewer() error = %v", err)
	}
	if r == nil {
		t.Fatal("NewPRReviewer returned nil")
	}
}

func TestNewPRReviewer_NilGHClient(t *testing.T) {
	t.Parallel()

	_, err := NewPRReviewer(nil, &mockQualityGate{}, nil)
	if !errors.Is(err, ErrNilGHClient) {
		t.Errorf("NewPRReviewer(nil gh) error = %v, want ErrNilGHClient", err)
	}
}

func TestNewPRReviewer_NilQualityGate(t *testing.T) {
	t.Parallel()

	_, err := NewPRReviewer(&mockGHClient{}, nil, nil)
	if !errors.Is(err, ErrNilQualityGate) {
		t.Errorf("NewPRReviewer(nil gate) error = %v, want ErrNilQualityGate", err)
	}
}

func TestReview_AllChecksPassed_Approve(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number: 42,
			Title:  "feat: add auth",
			State:  "OPEN",
		},
		prChecksResult: &CheckStatus{
			Overall: CheckPass,
			Checks: []Check{
				{Name: "build", Status: "completed", Conclusion: "success"},
				{Name: "test", Status: "completed", Conclusion: "success"},
			},
		},
	}
	gate := &mockQualityGate{
		report: &quality.Report{Passed: true, Score: 0.95},
	}
	r := mustNewPRReviewer(t, gh, gate, nil)

	report, err := r.Review(context.Background(), 42, "SPEC-ISSUE-42", nil)
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}
	if report.Decision != ReviewApprove {
		t.Errorf("Decision = %q, want %q", report.Decision, ReviewApprove)
	}
	if len(report.Issues) != 0 {
		t.Errorf("Issues = %v, want empty", report.Issues)
	}
}

func TestReview_QualityFailed_RequestChanges(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number: 43,
			State:  "OPEN",
		},
		prChecksResult: &CheckStatus{Overall: CheckPass},
	}
	gate := &mockQualityGate{
		report: &quality.Report{
			Passed: false,
			Score:  0.5,
			Principles: map[string]quality.PrincipleResult{
				"tested": {
					Name:   "tested",
					Passed: false,
					Score:  0.5,
					Issues: []quality.Issue{
						{
							File:     "main.go",
							Line:     10,
							Severity: quality.SeverityError,
							Message:  "undefined: foo",
							Rule:     "type-error",
						},
					},
				},
			},
		},
	}
	r := mustNewPRReviewer(t, gh, gate, nil)

	report, err := r.Review(context.Background(), 43, "SPEC-ISSUE-43", nil)
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}
	if report.Decision != ReviewRequestChanges {
		t.Errorf("Decision = %q, want %q", report.Decision, ReviewRequestChanges)
	}
	if len(report.Issues) == 0 {
		t.Error("Issues is empty, want at least one issue")
	}
}

func TestReview_CIFailed_RequestChanges(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number: 44,
			State:  "OPEN",
		},
		prChecksResult: &CheckStatus{
			Overall: CheckFail,
			Checks: []Check{
				{Name: "build", Status: "completed", Conclusion: "failure"},
			},
		},
	}
	gate := &mockQualityGate{
		report: &quality.Report{Passed: true, Score: 1.0},
	}
	r := mustNewPRReviewer(t, gh, gate, nil)

	report, err := r.Review(context.Background(), 44, "SPEC-ISSUE-44", nil)
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}
	if report.Decision != ReviewRequestChanges {
		t.Errorf("Decision = %q, want %q", report.Decision, ReviewRequestChanges)
	}
}

func TestReview_CIPending_Comment(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number: 45,
			State:  "OPEN",
		},
		prChecksResult: &CheckStatus{
			Overall: CheckPending,
			Checks: []Check{
				{Name: "build", Status: "in_progress", Conclusion: ""},
			},
		},
	}
	gate := &mockQualityGate{
		report: &quality.Report{Passed: true, Score: 1.0},
	}
	r := mustNewPRReviewer(t, gh, gate, nil)

	report, err := r.Review(context.Background(), 45, "SPEC-ISSUE-45", nil)
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}
	if report.Decision != ReviewComment {
		t.Errorf("Decision = %q, want %q", report.Decision, ReviewComment)
	}
}

func TestReview_PRNotFound(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewErr: ErrPRNotFound,
	}
	gate := &mockQualityGate{}
	r := mustNewPRReviewer(t, gh, gate, nil)

	_, err := r.Review(context.Background(), 999, "SPEC-ISSUE-999", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrPRNotFound) {
		t.Errorf("error = %v, want ErrPRNotFound", err)
	}
}

func TestReview_PRNotOpen(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number: 46,
			State:  "MERGED",
		},
	}
	gate := &mockQualityGate{}
	r := mustNewPRReviewer(t, gh, gate, nil)

	_, err := r.Review(context.Background(), 46, "SPEC-ISSUE-46", nil)
	if err == nil {
		t.Fatal("expected error for non-OPEN PR, got nil")
	}
}

func TestReview_QualityValidationError(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number: 47,
			State:  "OPEN",
		},
		prChecksResult: &CheckStatus{Overall: CheckPass},
	}
	gate := &mockQualityGate{
		err: errors.New("lsp timeout"),
	}
	r := mustNewPRReviewer(t, gh, gate, nil)

	report, err := r.Review(context.Background(), 47, "SPEC-ISSUE-47", nil)
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}
	if report.Decision != ReviewRequestChanges {
		t.Errorf("Decision = %q, want %q", report.Decision, ReviewRequestChanges)
	}
}

func TestReview_SummaryGeneration(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number: 48,
			State:  "OPEN",
		},
		prChecksResult: &CheckStatus{Overall: CheckPass},
	}
	gate := &mockQualityGate{
		report: &quality.Report{Passed: true, Score: 0.92},
	}
	r := mustNewPRReviewer(t, gh, gate, nil)

	report, err := r.Review(context.Background(), 48, "SPEC-ISSUE-48", nil)
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}
	if report.Summary == "" {
		t.Error("Summary is empty")
	}
	if report.Decision != ReviewApprove {
		t.Errorf("Decision = %q, want %q", report.Decision, ReviewApprove)
	}
}

func TestReview_CIChecksError(t *testing.T) {
	t.Parallel()

	gh := &mockGHClient{
		prViewResult: &PRDetails{
			Number: 49,
			State:  "OPEN",
		},
		prChecksErr: errors.New("checks API timeout"),
	}
	gate := &mockQualityGate{
		report: &quality.Report{Passed: true, Score: 1.0},
	}
	r := mustNewPRReviewer(t, gh, gate, nil)

	report, err := r.Review(context.Background(), 49, "SPEC-ISSUE-49", nil)
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}
	// CI check error with quality passing results in COMMENT (not REQUEST_CHANGES).
	if report.Decision != ReviewComment {
		t.Errorf("Decision = %q, want %q", report.Decision, ReviewComment)
	}
	if len(report.Issues) == 0 {
		t.Error("Issues is empty, want at least one issue for CI error")
	}
}
