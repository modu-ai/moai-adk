package worktree

import "fmt"

// WorktreeError is the base error type for worktree operations.
// It provides structured error information with SPEC-ID reference and recovery commands.
//
// @MX:NOTE: SPEC-WORKTREE-002 S4 implementation - centralized error handling
// @MX:SPEC: SPEC-WORKTREE-002
type WorktreeError struct {
	prefix   string // error message prefix (e.g., "Worktree creation failed")
	SpecID   string // SPEC reference
	Err      error  // underlying error
	Recovery string // recovery command suggestion
}

// Error returns the formatted error message including SPEC-ID, error details, and recovery command.
func (e *WorktreeError) Error() string {
	msg := fmt.Sprintf("%s: %s", e.prefix, e.SpecID)
	if e.Err != nil {
		msg = fmt.Sprintf("%s. %v", msg, e.Err)
	}
	if e.Recovery != "" {
		msg = fmt.Sprintf("%s. Recovery: `%s`", msg, e.Recovery)
	}
	return msg
}

// Unwrap returns the underlying error for errors.Is/As unwrapping.
func (e *WorktreeError) Unwrap() error {
	return e.Err
}

// NewWorktreeCreateError creates an error for worktree creation failures.
// Error template: "Worktree creation failed: {SPEC-ID}. {error}. Recovery: `moai worktree new {SPEC-ID}`"
func NewWorktreeCreateError(specID string, err error) *WorktreeError {
	return &WorktreeError{
		prefix:   "Worktree creation failed",
		SpecID:   specID,
		Err:      err,
		Recovery: fmt.Sprintf("moai worktree new %s", specID),
	}
}

// NewTmuxNotAvailableError creates an error when tmux is not available.
// Error template: "tmux not available: {SPEC-ID}. Recovery: `cd {path} && /moai run {SPEC-ID}`"
func NewTmuxNotAvailableError(specID, worktreePath string) *WorktreeError {
	return &WorktreeError{
		prefix:   "tmux not available",
		SpecID:   specID,
		Err:      nil,
		Recovery: fmt.Sprintf("cd %s && /moai run %s", worktreePath, specID),
	}
}

// NewAutoMergeBlockedError creates an error when auto-merge is blocked.
// Error templates:
// - CI failed: "Auto-merge blocked: {SPEC-ID}. CI checks failed. Recovery: `Fix issues and re-run: /moai sync {SPEC-ID}`"
// - Conflicts: "Auto-merge blocked: {SPEC-ID}. Merge conflicts detected. Recovery: `Resolve manually in PR`"
func NewAutoMergeBlockedError(specID, reason string) *WorktreeError {
	var recovery string
	if reason == "CI checks failed" {
		recovery = fmt.Sprintf("Fix issues and re-run: /moai sync %s", specID)
	} else if reason == "Merge conflicts detected" {
		recovery = "Resolve manually in PR"
	} else {
		recovery = fmt.Sprintf("/moai sync %s", specID)
	}

	return &WorktreeError{
		prefix:   "Auto-merge blocked",
		SpecID:   specID,
		Err:      fmt.Errorf("%s", reason),
		Recovery: recovery,
	}
}

// NewCleanupFailedError creates an error for worktree cleanup failures.
// Error template: "Worktree cleanup failed: {SPEC-ID}. {error}. Recovery: `moai worktree done {SPEC-ID}`"
func NewCleanupFailedError(specID string, err error) *WorktreeError {
	return &WorktreeError{
		prefix:   "Worktree cleanup failed",
		SpecID:   specID,
		Err:      err,
		Recovery: fmt.Sprintf("moai worktree done %s", specID),
	}
}
