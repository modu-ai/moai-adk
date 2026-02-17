package workflow

import "errors"

// Sentinel errors for the workflow package.
var (
	// ErrNotInWorktree indicates the current directory is not inside a worktree.
	ErrNotInWorktree = errors.New("workflow: not inside a worktree")

	// ErrSPECNotFound indicates the SPEC document was not found in the worktree.
	ErrSPECNotFound = errors.New("workflow: SPEC document not found")

	// ErrInvalidSPECID indicates the SPEC ID does not match the expected format.
	ErrInvalidSPECID = errors.New("workflow: invalid SPEC ID format")

	// ErrWorkflowInProgress indicates a workflow is already running.
	ErrWorkflowInProgress = errors.New("workflow: workflow already in progress")

	// ErrQualityGateFailed indicates TRUST 5 quality gates did not pass.
	ErrQualityGateFailed = errors.New("workflow: quality gate validation failed")

	// ErrPlanPhaseFailed indicates the Plan phase failed.
	ErrPlanPhaseFailed = errors.New("workflow: plan phase failed")

	// ErrRunPhaseFailed indicates the Run phase failed.
	ErrRunPhaseFailed = errors.New("workflow: run phase failed")

	// ErrSyncPhaseFailed indicates the Sync phase failed.
	ErrSyncPhaseFailed = errors.New("workflow: sync phase failed")

	// ErrNilWorktreeManager indicates a nil WorktreeManager was provided.
	ErrNilWorktreeManager = errors.New("workflow: WorktreeManager must not be nil")

	// ErrNilValidator indicates a nil WorktreeValidator was provided.
	ErrNilValidator = errors.New("workflow: WorktreeValidator must not be nil")

	// ErrNilExecutor indicates a nil PhaseExecutor was provided.
	ErrNilExecutor = errors.New("workflow: PhaseExecutor must not be nil")
)
