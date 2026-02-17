package workflow

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/core/quality"
)

// specIDPattern validates the expected SPEC-ISSUE-{number} format.
var specIDPattern = regexp.MustCompile(`^SPEC-ISSUE-\d+$`)

// WorkflowPhaseStatus tracks the completion state of a workflow phase.
type WorkflowPhaseStatus string

const (
	// PhaseStatusPending indicates the phase has not started.
	PhaseStatusPending WorkflowPhaseStatus = "pending"

	// PhaseStatusRunning indicates the phase is currently executing.
	PhaseStatusRunning WorkflowPhaseStatus = "running"

	// PhaseStatusCompleted indicates the phase completed successfully.
	PhaseStatusCompleted WorkflowPhaseStatus = "completed"

	// PhaseStatusFailed indicates the phase failed.
	PhaseStatusFailed WorkflowPhaseStatus = "failed"

	// PhaseStatusSkipped indicates the phase was skipped.
	PhaseStatusSkipped WorkflowPhaseStatus = "skipped"
)

// WorktreeContext holds information about the current worktree environment.
type WorktreeContext struct {
	// SpecID is the SPEC identifier (e.g., "SPEC-ISSUE-123").
	SpecID string

	// WorktreeDir is the absolute filesystem path to the worktree.
	WorktreeDir string

	// Branch is the Git branch name (e.g., "fix/issue-123").
	Branch string

	// BaseBranch is the base branch to merge into (e.g., "main").
	BaseBranch string

	// IssueNumber is the original GitHub issue number.
	IssueNumber int
}

// WorkflowResult summarizes the Plan-Run-Sync execution.
type WorkflowResult struct {
	SpecID        string
	PlanStatus    WorkflowPhaseStatus
	RunStatus     WorkflowPhaseStatus
	SyncStatus    WorkflowPhaseStatus
	QualityReport *quality.Report
	StartedAt     time.Time
	CompletedAt   time.Time
}

// ReviewReadiness indicates whether a worktree is ready for PR creation.
type ReviewReadiness struct {
	Ready          bool
	QualityPassed  bool
	QualityReport  *quality.Report
	FailureReasons []string
}

// PhaseExecutor runs a single workflow phase (plan, run, or sync).
// This abstraction allows the orchestrator to be tested without actually
// invoking Claude Code commands.
type PhaseExecutor interface {
	// ExecutePlan runs the plan phase for a SPEC.
	ExecutePlan(ctx context.Context, specID, workDir string) error

	// ExecuteRun runs the implementation phase for a SPEC.
	ExecuteRun(ctx context.Context, specID, workDir string) error

	// ExecuteSync runs the documentation/sync phase for a SPEC.
	ExecuteSync(ctx context.Context, specID, workDir string) error
}

// WorktreeOrchestrator coordinates the Plan-Run-Sync workflow within a worktree.
type WorktreeOrchestrator interface {
	// DetectWorktreeContext identifies the current worktree and loads SPEC metadata.
	DetectWorktreeContext(ctx context.Context, dir string) (*WorktreeContext, error)

	// ExecuteWorkflow runs the Plan-Run-Sync sequence for a given SPEC.
	ExecuteWorkflow(ctx context.Context, specID string) (*WorkflowResult, error)

	// ValidateQuality runs TRUST 5 quality gates on the worktree.
	ValidateQuality(ctx context.Context, specID string) (*quality.Report, error)

	// PrepareForReview verifies quality gates and marks worktree ready for PR.
	PrepareForReview(ctx context.Context, specID string) (*ReviewReadiness, error)
}

// defaultBranchDetectorFunc returns the repository's default branch name.
type defaultBranchDetectorFunc func(ctx context.Context, root string) string

// worktreeOrchestrator implements WorktreeOrchestrator.
type worktreeOrchestrator struct {
	worktreeMgr    git.WorktreeManager
	validator      quality.WorktreeValidator
	executor       PhaseExecutor
	detectBranch   defaultBranchDetectorFunc
	logger         *slog.Logger
}

// Compile-time interface compliance check.
var _ WorktreeOrchestrator = (*worktreeOrchestrator)(nil)

// NewWorktreeOrchestrator creates an orchestrator for worktree-based workflows.
// Returns an error if worktreeMgr, validator, or executor is nil.
func NewWorktreeOrchestrator(
	worktreeMgr git.WorktreeManager,
	validator quality.WorktreeValidator,
	executor PhaseExecutor,
	logger *slog.Logger,
) (*worktreeOrchestrator, error) {
	if worktreeMgr == nil {
		return nil, ErrNilWorktreeManager
	}
	if validator == nil {
		return nil, ErrNilValidator
	}
	if executor == nil {
		return nil, ErrNilExecutor
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &worktreeOrchestrator{
		worktreeMgr:  worktreeMgr,
		validator:    validator,
		executor:     executor,
		detectBranch: detectDefaultBranch,
		logger:       logger.With("module", "worktree-orchestrator"),
	}, nil
}

// DetectWorktreeContext identifies which worktree the given directory belongs to
// and extracts SPEC metadata from it.
func (o *worktreeOrchestrator) DetectWorktreeContext(ctx context.Context, dir string) (*WorktreeContext, error) {
	worktrees, err := o.worktreeMgr.List()
	if err != nil {
		return nil, fmt.Errorf("list worktrees: %w", err)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve absolute path %q: %w", dir, err)
	}

	// Find the worktree that contains the given directory.
	var matched *git.Worktree
	for i := range worktrees {
		wt := &worktrees[i]
		wtAbs, err := filepath.Abs(wt.Path)
		if err != nil {
			continue
		}
		if absDir == wtAbs || strings.HasPrefix(absDir, wtAbs+string(filepath.Separator)) {
			matched = wt
			break
		}
	}

	if matched == nil {
		return nil, fmt.Errorf("directory %q: %w", absDir, ErrNotInWorktree)
	}

	// Extract SPEC ID from worktree path (last directory component).
	specID := filepath.Base(matched.Path)
	if !specIDPattern.MatchString(specID) {
		return nil, fmt.Errorf("worktree %q has invalid SPEC ID %q: %w", matched.Path, specID, ErrInvalidSPECID)
	}

	// Verify SPEC document exists in the worktree.
	specDir := filepath.Join(matched.Path, ".moai", "specs", specID)
	if _, err := os.Stat(specDir); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("SPEC directory %q: %w", specDir, ErrSPECNotFound)
		}
		return nil, fmt.Errorf("stat SPEC directory %q: %w", specDir, err)
	}

	issueNumber := extractIssueNumber(specID)

	return &WorktreeContext{
		SpecID:      specID,
		WorktreeDir: matched.Path,
		Branch:      matched.Branch,
		BaseBranch:  o.detectBranch(ctx, o.worktreeMgr.Root()),
		IssueNumber: issueNumber,
	}, nil
}

// ExecuteWorkflow runs the Plan-Run-Sync sequence for the given SPEC.
func (o *worktreeOrchestrator) ExecuteWorkflow(ctx context.Context, specID string) (*WorkflowResult, error) {
	if !specIDPattern.MatchString(specID) {
		return nil, fmt.Errorf("SPEC ID %q: %w", specID, ErrInvalidSPECID)
	}

	wtCtx, err := o.findWorktreeForSpec(ctx, specID)
	if err != nil {
		return nil, fmt.Errorf("find worktree for %s: %w", specID, err)
	}

	result := &WorkflowResult{
		SpecID:     specID,
		PlanStatus: PhaseStatusPending,
		RunStatus:  PhaseStatusPending,
		SyncStatus: PhaseStatusPending,
		StartedAt:  time.Now(),
	}

	o.logger.Info("starting workflow execution", "spec_id", specID, "worktree", wtCtx.WorktreeDir)

	// Phase 1: Plan
	result.PlanStatus = PhaseStatusRunning
	if err := o.executor.ExecutePlan(ctx, specID, wtCtx.WorktreeDir); err != nil {
		result.PlanStatus = PhaseStatusFailed
		result.RunStatus = PhaseStatusSkipped
		result.SyncStatus = PhaseStatusSkipped
		result.CompletedAt = time.Now()
		return result, fmt.Errorf("plan phase for %s: %w", specID, ErrPlanPhaseFailed)
	}
	result.PlanStatus = PhaseStatusCompleted

	// Phase 2: Run
	result.RunStatus = PhaseStatusRunning
	if err := o.executor.ExecuteRun(ctx, specID, wtCtx.WorktreeDir); err != nil {
		result.RunStatus = PhaseStatusFailed
		result.SyncStatus = PhaseStatusSkipped
		result.CompletedAt = time.Now()
		return result, fmt.Errorf("run phase for %s: %w", specID, ErrRunPhaseFailed)
	}
	result.RunStatus = PhaseStatusCompleted

	// Validate quality after run phase.
	qualityReport, err := o.validator.Validate(ctx, wtCtx.WorktreeDir)
	if err != nil {
		o.logger.Warn("quality validation failed", "spec_id", specID, "error", err)
	}
	result.QualityReport = qualityReport

	// Phase 3: Sync
	result.SyncStatus = PhaseStatusRunning
	if err := o.executor.ExecuteSync(ctx, specID, wtCtx.WorktreeDir); err != nil {
		result.SyncStatus = PhaseStatusFailed
		result.CompletedAt = time.Now()
		return result, fmt.Errorf("sync phase for %s: %w", specID, ErrSyncPhaseFailed)
	}
	result.SyncStatus = PhaseStatusCompleted

	result.CompletedAt = time.Now()

	o.logger.Info("workflow execution complete",
		"spec_id", specID,
		"plan", string(result.PlanStatus),
		"run", string(result.RunStatus),
		"sync", string(result.SyncStatus),
	)

	return result, nil
}

// ValidateQuality runs TRUST 5 quality gates on the worktree for the given SPEC.
func (o *worktreeOrchestrator) ValidateQuality(ctx context.Context, specID string) (*quality.Report, error) {
	if !specIDPattern.MatchString(specID) {
		return nil, fmt.Errorf("SPEC ID %q: %w", specID, ErrInvalidSPECID)
	}

	wtCtx, err := o.findWorktreeForSpec(ctx, specID)
	if err != nil {
		return nil, fmt.Errorf("find worktree for %s: %w", specID, err)
	}

	return o.validator.Validate(ctx, wtCtx.WorktreeDir)
}

// PrepareForReview checks quality gates and returns readiness status.
func (o *worktreeOrchestrator) PrepareForReview(ctx context.Context, specID string) (*ReviewReadiness, error) {
	if !specIDPattern.MatchString(specID) {
		return nil, fmt.Errorf("SPEC ID %q: %w", specID, ErrInvalidSPECID)
	}

	wtCtx, err := o.findWorktreeForSpec(ctx, specID)
	if err != nil {
		return nil, fmt.Errorf("find worktree for %s: %w", specID, err)
	}

	readiness := &ReviewReadiness{
		Ready:          false,
		QualityPassed:  false,
		FailureReasons: []string{},
	}

	report, err := o.validator.Validate(ctx, wtCtx.WorktreeDir)
	if err != nil {
		readiness.FailureReasons = append(readiness.FailureReasons,
			fmt.Sprintf("quality validation error: %v", err))
		return readiness, nil
	}

	readiness.QualityReport = report
	readiness.QualityPassed = report.Passed

	if !report.Passed {
		readiness.FailureReasons = append(readiness.FailureReasons, "TRUST 5 quality gates not passed")
		for _, issue := range report.AllIssues() {
			if issue.Severity == quality.SeverityError {
				readiness.FailureReasons = append(readiness.FailureReasons,
					fmt.Sprintf("%s:%d: %s", issue.File, issue.Line, issue.Message))
			}
		}
		return readiness, nil
	}

	readiness.Ready = true
	return readiness, nil
}

// findWorktreeForSpec looks up the worktree directory for a given SPEC ID.
func (o *worktreeOrchestrator) findWorktreeForSpec(ctx context.Context, specID string) (*WorktreeContext, error) {
	worktrees, err := o.worktreeMgr.List()
	if err != nil {
		return nil, fmt.Errorf("list worktrees: %w", err)
	}

	for _, wt := range worktrees {
		if filepath.Base(wt.Path) == specID {
			return &WorktreeContext{
				SpecID:      specID,
				WorktreeDir: wt.Path,
				Branch:      wt.Branch,
				BaseBranch:  o.detectBranch(ctx, o.worktreeMgr.Root()),
				IssueNumber: extractIssueNumber(specID),
			}, nil
		}
	}

	return nil, fmt.Errorf("no worktree found for %s: %w", specID, ErrNotInWorktree)
}

// detectDefaultBranch determines the repository's default branch by reading
// the symbolic ref for origin/HEAD. Falls back to "main" if the git command
// fails or returns an empty result.
func detectDefaultBranch(ctx context.Context, root string) string {
	out, err := exec.CommandContext(ctx, "git", "-C", root, "symbolic-ref", "refs/remotes/origin/HEAD", "--short").Output()
	if err != nil {
		return "main"
	}
	ref := strings.TrimSpace(string(out))
	if ref == "" {
		return "main"
	}
	// The output is e.g. "origin/main" — strip the "origin/" prefix.
	if after, ok := strings.CutPrefix(ref, "origin/"); ok && after != "" {
		return after
	}
	return ref
}

// extractIssueNumber parses the issue number from a SPEC ID like "SPEC-ISSUE-123".
func extractIssueNumber(specID string) int {
	parts := strings.Split(specID, "-")
	if len(parts) < 3 {
		return 0
	}
	n, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return 0
	}
	return n
}
