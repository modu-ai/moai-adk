package workflow

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/core/quality"
)

// --- Mocks ---

// mockWorktreeManager implements git.WorktreeManager for testing.
type mockWorktreeManager struct {
	worktrees []git.Worktree
	listErr   error
	addErr    error
	root      string
}

func (m *mockWorktreeManager) Add(_, _ string) error          { return m.addErr }
func (m *mockWorktreeManager) List() ([]git.Worktree, error)  { return m.worktrees, m.listErr }
func (m *mockWorktreeManager) Remove(_ string, _ bool) error  { return nil }
func (m *mockWorktreeManager) Prune() error                   { return nil }
func (m *mockWorktreeManager) Repair() error                  { return nil }
func (m *mockWorktreeManager) Root() string                   { return m.root }
func (m *mockWorktreeManager) Sync(_, _, _ string) error      { return nil }
func (m *mockWorktreeManager) DeleteBranch(_ string) error    { return nil }
func (m *mockWorktreeManager) IsBranchMerged(_, _ string) (bool, error) {
	return false, nil
}

// mockWorktreeValidator implements quality.WorktreeValidator for testing.
type mockWorktreeValidator struct {
	report *quality.Report
	err    error
}

func (m *mockWorktreeValidator) Validate(_ context.Context, _ string) (*quality.Report, error) {
	return m.report, m.err
}

func (m *mockWorktreeValidator) ValidateWithConfig(_ context.Context, _ string, _ quality.QualityConfig) (*quality.Report, error) {
	return m.report, m.err
}

// mockPhaseExecutor implements PhaseExecutor for testing.
type mockPhaseExecutor struct {
	planErr error
	runErr  error
	syncErr error

	planCalled bool
	runCalled  bool
	syncCalled bool
}

func (m *mockPhaseExecutor) ExecutePlan(_ context.Context, _, _ string) error {
	m.planCalled = true
	return m.planErr
}

func (m *mockPhaseExecutor) ExecuteRun(_ context.Context, _, _ string) error {
	m.runCalled = true
	return m.runErr
}

func (m *mockPhaseExecutor) ExecuteSync(_ context.Context, _, _ string) error {
	m.syncCalled = true
	return m.syncErr
}

// --- Helper ---

// setupWorktree creates a temp directory structure simulating a worktree with SPEC.
func setupWorktree(t *testing.T, specID string) string {
	t.Helper()
	dir := t.TempDir()
	wtPath := filepath.Join(dir, specID)
	specDir := filepath.Join(wtPath, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("create spec dir: %v", err)
	}
	specFile := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specFile, []byte("# "+specID), 0o644); err != nil {
		t.Fatalf("create spec.md: %v", err)
	}
	return wtPath
}

// mustNewWorktreeOrchestrator is a test helper that calls NewWorktreeOrchestrator and fails on error.
func mustNewWorktreeOrchestrator(t *testing.T, wm git.WorktreeManager, wv quality.WorktreeValidator, pe PhaseExecutor, logger *slog.Logger) *worktreeOrchestrator {
	t.Helper()
	o, err := NewWorktreeOrchestrator(wm, wv, pe, logger)
	if err != nil {
		t.Fatalf("NewWorktreeOrchestrator() error = %v", err)
	}
	return o
}

// --- Tests ---

func TestNewWorktreeOrchestrator(t *testing.T) {
	t.Parallel()

	wm := &mockWorktreeManager{}
	wv := &mockWorktreeValidator{}
	pe := &mockPhaseExecutor{}

	o, err := NewWorktreeOrchestrator(wm, wv, pe, nil)
	if err != nil {
		t.Fatalf("NewWorktreeOrchestrator() error = %v", err)
	}
	if o == nil {
		t.Fatal("NewWorktreeOrchestrator returned nil")
	}
}

func TestNewWorktreeOrchestrator_NilWorktreeManager(t *testing.T) {
	t.Parallel()

	_, err := NewWorktreeOrchestrator(nil, &mockWorktreeValidator{}, &mockPhaseExecutor{}, nil)
	if !errors.Is(err, ErrNilWorktreeManager) {
		t.Errorf("NewWorktreeOrchestrator(nil wm) error = %v, want ErrNilWorktreeManager", err)
	}
}

func TestNewWorktreeOrchestrator_NilValidator(t *testing.T) {
	t.Parallel()

	_, err := NewWorktreeOrchestrator(&mockWorktreeManager{}, nil, &mockPhaseExecutor{}, nil)
	if !errors.Is(err, ErrNilValidator) {
		t.Errorf("NewWorktreeOrchestrator(nil validator) error = %v, want ErrNilValidator", err)
	}
}

func TestNewWorktreeOrchestrator_NilExecutor(t *testing.T) {
	t.Parallel()

	_, err := NewWorktreeOrchestrator(&mockWorktreeManager{}, &mockWorktreeValidator{}, nil, nil)
	if !errors.Is(err, ErrNilExecutor) {
		t.Errorf("NewWorktreeOrchestrator(nil executor) error = %v, want ErrNilExecutor", err)
	}
}

func TestDetectWorktreeContext_ValidWorktree(t *testing.T) {
	t.Parallel()

	wtPath := setupWorktree(t, "SPEC-ISSUE-123")
	wm := &mockWorktreeManager{
		worktrees: []git.Worktree{
			{Path: wtPath, Branch: "fix/issue-123", HEAD: "abc123"},
		},
	}
	wv := &mockWorktreeValidator{}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	ctx := context.Background()
	wtCtx, err := o.DetectWorktreeContext(ctx, wtPath)
	if err != nil {
		t.Fatalf("DetectWorktreeContext() error = %v", err)
	}
	if wtCtx.SpecID != "SPEC-ISSUE-123" {
		t.Errorf("SpecID = %q, want %q", wtCtx.SpecID, "SPEC-ISSUE-123")
	}
	if wtCtx.Branch != "fix/issue-123" {
		t.Errorf("Branch = %q, want %q", wtCtx.Branch, "fix/issue-123")
	}
	if wtCtx.IssueNumber != 123 {
		t.Errorf("IssueNumber = %d, want %d", wtCtx.IssueNumber, 123)
	}
}

func TestDetectWorktreeContext_NotInWorktree(t *testing.T) {
	t.Parallel()

	wm := &mockWorktreeManager{
		worktrees: []git.Worktree{},
	}
	wv := &mockWorktreeValidator{}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	_, err := o.DetectWorktreeContext(context.Background(), "/some/random/dir")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotInWorktree) {
		t.Errorf("error = %v, want ErrNotInWorktree", err)
	}
}

func TestDetectWorktreeContext_SPECNotFound(t *testing.T) {
	t.Parallel()

	// Create worktree dir without SPEC.
	dir := t.TempDir()
	wtPath := filepath.Join(dir, "SPEC-ISSUE-999")
	if err := os.MkdirAll(wtPath, 0o755); err != nil {
		t.Fatalf("create worktree dir: %v", err)
	}

	wm := &mockWorktreeManager{
		worktrees: []git.Worktree{
			{Path: wtPath, Branch: "feat/issue-999", HEAD: "def456"},
		},
	}
	wv := &mockWorktreeValidator{}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	_, err := o.DetectWorktreeContext(context.Background(), wtPath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrSPECNotFound) {
		t.Errorf("error = %v, want ErrSPECNotFound", err)
	}
}

func TestExecuteWorkflow_AllPhasesSucceed(t *testing.T) {
	t.Parallel()

	wtPath := setupWorktree(t, "SPEC-ISSUE-42")
	wm := &mockWorktreeManager{
		worktrees: []git.Worktree{
			{Path: wtPath, Branch: "feat/issue-42", HEAD: "aaa"},
		},
	}
	wv := &mockWorktreeValidator{
		report: &quality.Report{Passed: true, Score: 1.0},
	}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	result, err := o.ExecuteWorkflow(context.Background(), "SPEC-ISSUE-42")
	if err != nil {
		t.Fatalf("ExecuteWorkflow() error = %v", err)
	}
	if result.PlanStatus != PhaseStatusCompleted {
		t.Errorf("PlanStatus = %q, want %q", result.PlanStatus, PhaseStatusCompleted)
	}
	if result.RunStatus != PhaseStatusCompleted {
		t.Errorf("RunStatus = %q, want %q", result.RunStatus, PhaseStatusCompleted)
	}
	if result.SyncStatus != PhaseStatusCompleted {
		t.Errorf("SyncStatus = %q, want %q", result.SyncStatus, PhaseStatusCompleted)
	}
	if !pe.planCalled {
		t.Error("Plan phase not executed")
	}
	if !pe.runCalled {
		t.Error("Run phase not executed")
	}
	if !pe.syncCalled {
		t.Error("Sync phase not executed")
	}
}

func TestExecuteWorkflow_PlanFails(t *testing.T) {
	t.Parallel()

	wtPath := setupWorktree(t, "SPEC-ISSUE-50")
	wm := &mockWorktreeManager{
		worktrees: []git.Worktree{
			{Path: wtPath, Branch: "feat/issue-50", HEAD: "bbb"},
		},
	}
	wv := &mockWorktreeValidator{}
	pe := &mockPhaseExecutor{planErr: errors.New("plan failed")}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	result, err := o.ExecuteWorkflow(context.Background(), "SPEC-ISSUE-50")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrPlanPhaseFailed) {
		t.Errorf("error = %v, want ErrPlanPhaseFailed", err)
	}
	if result.PlanStatus != PhaseStatusFailed {
		t.Errorf("PlanStatus = %q, want %q", result.PlanStatus, PhaseStatusFailed)
	}
	if result.RunStatus != PhaseStatusSkipped {
		t.Errorf("RunStatus = %q, want %q", result.RunStatus, PhaseStatusSkipped)
	}
	if result.SyncStatus != PhaseStatusSkipped {
		t.Errorf("SyncStatus = %q, want %q", result.SyncStatus, PhaseStatusSkipped)
	}
}

func TestExecuteWorkflow_RunFails(t *testing.T) {
	t.Parallel()

	wtPath := setupWorktree(t, "SPEC-ISSUE-60")
	wm := &mockWorktreeManager{
		worktrees: []git.Worktree{
			{Path: wtPath, Branch: "feat/issue-60", HEAD: "ccc"},
		},
	}
	wv := &mockWorktreeValidator{}
	pe := &mockPhaseExecutor{runErr: errors.New("run failed")}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	result, err := o.ExecuteWorkflow(context.Background(), "SPEC-ISSUE-60")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrRunPhaseFailed) {
		t.Errorf("error = %v, want ErrRunPhaseFailed", err)
	}
	if result.PlanStatus != PhaseStatusCompleted {
		t.Errorf("PlanStatus = %q, want %q", result.PlanStatus, PhaseStatusCompleted)
	}
	if result.RunStatus != PhaseStatusFailed {
		t.Errorf("RunStatus = %q, want %q", result.RunStatus, PhaseStatusFailed)
	}
	if result.SyncStatus != PhaseStatusSkipped {
		t.Errorf("SyncStatus = %q, want %q", result.SyncStatus, PhaseStatusSkipped)
	}
}

func TestExecuteWorkflow_SyncFails(t *testing.T) {
	t.Parallel()

	wtPath := setupWorktree(t, "SPEC-ISSUE-70")
	wm := &mockWorktreeManager{
		worktrees: []git.Worktree{
			{Path: wtPath, Branch: "feat/issue-70", HEAD: "ddd"},
		},
	}
	wv := &mockWorktreeValidator{
		report: &quality.Report{Passed: true, Score: 1.0},
	}
	pe := &mockPhaseExecutor{syncErr: errors.New("sync failed")}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	result, err := o.ExecuteWorkflow(context.Background(), "SPEC-ISSUE-70")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrSyncPhaseFailed) {
		t.Errorf("error = %v, want ErrSyncPhaseFailed", err)
	}
	if result.PlanStatus != PhaseStatusCompleted {
		t.Errorf("PlanStatus = %q, want %q", result.PlanStatus, PhaseStatusCompleted)
	}
	if result.RunStatus != PhaseStatusCompleted {
		t.Errorf("RunStatus = %q, want %q", result.RunStatus, PhaseStatusCompleted)
	}
	if result.SyncStatus != PhaseStatusFailed {
		t.Errorf("SyncStatus = %q, want %q", result.SyncStatus, PhaseStatusFailed)
	}
}

func TestExecuteWorkflow_InvalidSPECID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		specID string
	}{
		{"empty", ""},
		{"no prefix", "ISSUE-123"},
		{"wrong prefix", "SPEC-123"},
		{"no number", "SPEC-ISSUE-abc"},
		{"extra parts", "SPEC-ISSUE-123-extra"},
	}

	wm := &mockWorktreeManager{}
	wv := &mockWorktreeValidator{}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := o.ExecuteWorkflow(context.Background(), tt.specID)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, ErrInvalidSPECID) {
				t.Errorf("error = %v, want ErrInvalidSPECID", err)
			}
		})
	}
}

func TestValidateQuality_Passing(t *testing.T) {
	t.Parallel()

	wtPath := setupWorktree(t, "SPEC-ISSUE-80")
	wm := &mockWorktreeManager{
		worktrees: []git.Worktree{
			{Path: wtPath, Branch: "feat/issue-80", HEAD: "eee"},
		},
	}
	wv := &mockWorktreeValidator{
		report: &quality.Report{
			Passed: true,
			Score:  0.95,
			Principles: map[string]quality.PrincipleResult{
				"tested":   {Name: "tested", Passed: true, Score: 0.9},
				"readable": {Name: "readable", Passed: true, Score: 1.0},
			},
		},
	}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	report, err := o.ValidateQuality(context.Background(), "SPEC-ISSUE-80")
	if err != nil {
		t.Fatalf("ValidateQuality() error = %v", err)
	}
	if !report.Passed {
		t.Error("report.Passed = false, want true")
	}
}

func TestPrepareForReview_QualityPassed(t *testing.T) {
	t.Parallel()

	wtPath := setupWorktree(t, "SPEC-ISSUE-90")
	wm := &mockWorktreeManager{
		worktrees: []git.Worktree{
			{Path: wtPath, Branch: "feat/issue-90", HEAD: "fff"},
		},
	}
	wv := &mockWorktreeValidator{
		report: &quality.Report{Passed: true, Score: 1.0},
	}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	readiness, err := o.PrepareForReview(context.Background(), "SPEC-ISSUE-90")
	if err != nil {
		t.Fatalf("PrepareForReview() error = %v", err)
	}
	if !readiness.Ready {
		t.Errorf("Ready = false, want true")
	}
	if !readiness.QualityPassed {
		t.Errorf("QualityPassed = false, want true")
	}
}

func TestPrepareForReview_QualityFailed(t *testing.T) {
	t.Parallel()

	wtPath := setupWorktree(t, "SPEC-ISSUE-91")
	wm := &mockWorktreeManager{
		worktrees: []git.Worktree{
			{Path: wtPath, Branch: "feat/issue-91", HEAD: "ggg"},
		},
	}
	wv := &mockWorktreeValidator{
		report: &quality.Report{
			Passed: false,
			Score:  0.5,
			Principles: map[string]quality.PrincipleResult{
				"tested": {
					Name:   "tested",
					Passed: false,
					Score:  0.5,
					Issues: []quality.Issue{
						{File: "main.go", Line: 10, Severity: quality.SeverityError, Message: "type error"},
					},
				},
			},
		},
	}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	readiness, err := o.PrepareForReview(context.Background(), "SPEC-ISSUE-91")
	if err != nil {
		t.Fatalf("PrepareForReview() error = %v", err)
	}
	if readiness.Ready {
		t.Error("Ready = true, want false")
	}
	if readiness.QualityPassed {
		t.Error("QualityPassed = true, want false")
	}
	if len(readiness.FailureReasons) == 0 {
		t.Error("FailureReasons is empty, want at least one reason")
	}
}

func TestDetectWorktreeContext_ListError(t *testing.T) {
	t.Parallel()

	wm := &mockWorktreeManager{
		listErr: errors.New("git error"),
	}
	wv := &mockWorktreeValidator{}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	_, err := o.DetectWorktreeContext(context.Background(), "/some/dir")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDetectWorktreeContext_InvalidSPECIDInWorktreeName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Create a worktree with non-SPEC name.
	wtPath := filepath.Join(dir, "my-feature-branch")
	if err := os.MkdirAll(wtPath, 0o755); err != nil {
		t.Fatalf("create dir: %v", err)
	}

	wm := &mockWorktreeManager{
		worktrees: []git.Worktree{
			{Path: wtPath, Branch: "feat/my-feature", HEAD: "abc123"},
		},
	}
	wv := &mockWorktreeValidator{}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	_, err := o.DetectWorktreeContext(context.Background(), wtPath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidSPECID) {
		t.Errorf("error = %v, want ErrInvalidSPECID", err)
	}
}

func TestValidateQuality_InvalidSPECID(t *testing.T) {
	t.Parallel()

	wm := &mockWorktreeManager{}
	wv := &mockWorktreeValidator{}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	_, err := o.ValidateQuality(context.Background(), "INVALID-SPEC")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidSPECID) {
		t.Errorf("error = %v, want ErrInvalidSPECID", err)
	}
}

func TestValidateQuality_WorktreeNotFound(t *testing.T) {
	t.Parallel()

	wm := &mockWorktreeManager{
		worktrees: []git.Worktree{},
	}
	wv := &mockWorktreeValidator{}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	_, err := o.ValidateQuality(context.Background(), "SPEC-ISSUE-999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotInWorktree) {
		t.Errorf("error = %v, want ErrNotInWorktree", err)
	}
}

func TestValidateQuality_ListError(t *testing.T) {
	t.Parallel()

	wm := &mockWorktreeManager{
		listErr: errors.New("git list failed"),
	}
	wv := &mockWorktreeValidator{}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	_, err := o.ValidateQuality(context.Background(), "SPEC-ISSUE-100")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPrepareForReview_InvalidSPECID(t *testing.T) {
	t.Parallel()

	wm := &mockWorktreeManager{}
	wv := &mockWorktreeValidator{}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	_, err := o.PrepareForReview(context.Background(), "BAD-SPEC")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidSPECID) {
		t.Errorf("error = %v, want ErrInvalidSPECID", err)
	}
}

func TestPrepareForReview_WorktreeNotFound(t *testing.T) {
	t.Parallel()

	wm := &mockWorktreeManager{
		worktrees: []git.Worktree{},
	}
	wv := &mockWorktreeValidator{}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	_, err := o.PrepareForReview(context.Background(), "SPEC-ISSUE-999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotInWorktree) {
		t.Errorf("error = %v, want ErrNotInWorktree", err)
	}
}

func TestPrepareForReview_ValidatorError(t *testing.T) {
	t.Parallel()

	wtPath := setupWorktree(t, "SPEC-ISSUE-95")
	wm := &mockWorktreeManager{
		worktrees: []git.Worktree{
			{Path: wtPath, Branch: "feat/issue-95", HEAD: "hhh"},
		},
	}
	wv := &mockWorktreeValidator{
		err: errors.New("lsp timeout"),
	}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	readiness, err := o.PrepareForReview(context.Background(), "SPEC-ISSUE-95")
	if err != nil {
		t.Fatalf("PrepareForReview() error = %v", err)
	}
	if readiness.Ready {
		t.Error("Ready = true, want false (validator error)")
	}
	if len(readiness.FailureReasons) == 0 {
		t.Error("FailureReasons is empty, want at least one")
	}
}

func TestExecuteWorkflow_WorktreeNotFound(t *testing.T) {
	t.Parallel()

	wm := &mockWorktreeManager{
		worktrees: []git.Worktree{},
	}
	wv := &mockWorktreeValidator{}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	_, err := o.ExecuteWorkflow(context.Background(), "SPEC-ISSUE-999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotInWorktree) {
		t.Errorf("error = %v, want ErrNotInWorktree", err)
	}
}

func TestExecuteWorkflow_QualityValidationError(t *testing.T) {
	t.Parallel()

	wtPath := setupWorktree(t, "SPEC-ISSUE-85")
	wm := &mockWorktreeManager{
		worktrees: []git.Worktree{
			{Path: wtPath, Branch: "feat/issue-85", HEAD: "iii"},
		},
	}
	wv := &mockWorktreeValidator{
		err: errors.New("quality check failed"),
	}
	pe := &mockPhaseExecutor{}
	o := mustNewWorktreeOrchestrator(t, wm, wv, pe, nil)

	// Quality validation error is non-fatal; workflow should still complete.
	result, err := o.ExecuteWorkflow(context.Background(), "SPEC-ISSUE-85")
	if err != nil {
		t.Fatalf("ExecuteWorkflow() error = %v", err)
	}
	if result.QualityReport != nil {
		t.Error("QualityReport should be nil when validator errors")
	}
	if result.SyncStatus != PhaseStatusCompleted {
		t.Errorf("SyncStatus = %q, want %q", result.SyncStatus, PhaseStatusCompleted)
	}
}

func TestDetectDefaultBranch(t *testing.T) {
	t.Parallel()

	t.Run("non-git directory falls back to main", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		got := detectDefaultBranch(context.Background(), dir)
		if got != "main" {
			t.Errorf("detectDefaultBranch(non-git) = %q, want %q", got, "main")
		}
	})

	t.Run("empty root falls back to main", func(t *testing.T) {
		t.Parallel()
		got := detectDefaultBranch(context.Background(), "")
		if got != "main" {
			t.Errorf("detectDefaultBranch(\"\") = %q, want %q", got, "main")
		}
	})

	t.Run("nonexistent path falls back to main", func(t *testing.T) {
		t.Parallel()
		got := detectDefaultBranch(context.Background(), "/nonexistent/path/that/does/not/exist")
		if got != "main" {
			t.Errorf("detectDefaultBranch(nonexistent) = %q, want %q", got, "main")
		}
	})
}

func TestExtractIssueNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  int
	}{
		{"SPEC-ISSUE-123", 123},
		{"SPEC-ISSUE-1", 1},
		{"SPEC-ISSUE-0", 0},
		{"INVALID", 0},
		{"SPEC-ISSUE-abc", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got := extractIssueNumber(tt.input)
			if got != tt.want {
				t.Errorf("extractIssueNumber(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
