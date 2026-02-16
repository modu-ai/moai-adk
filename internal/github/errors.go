package github

import (
	"errors"
	"fmt"
)

// Sentinel errors for the github package.
var (
	// ErrGHNotFound indicates the gh CLI binary is not in PATH.
	ErrGHNotFound = errors.New("github: gh CLI not found")

	// ErrGHNotAuthenticated indicates gh is not authenticated.
	ErrGHNotAuthenticated = errors.New("github: gh not authenticated")

	// ErrPRNotFound indicates the specified pull request does not exist.
	ErrPRNotFound = errors.New("github: pull request not found")

	// ErrPRAlreadyExists indicates a PR already exists for this branch.
	ErrPRAlreadyExists = errors.New("github: pull request already exists for branch")

	// ErrMergeBlocked indicates one or more merge prerequisites are not met.
	ErrMergeBlocked = errors.New("github: merge blocked by unmet prerequisites")

	// ErrMergeConflict indicates the PR has merge conflicts.
	ErrMergeConflict = errors.New("github: merge conflicts detected")

	// ErrCIFailed indicates CI/CD checks failed.
	ErrCIFailed = errors.New("github: CI/CD checks failed")

	// ErrReviewRequired indicates PR review has not been approved.
	ErrReviewRequired = errors.New("github: review approval required")

	// ErrAutoMergeNotRequested indicates the --auto-merge flag was not specified.
	ErrAutoMergeNotRequested = errors.New("github: auto-merge not requested")

	// ErrIssueNotFound indicates the specified issue does not exist.
	ErrIssueNotFound = errors.New("github: issue not found")

	// ErrCommentFailed indicates a failure posting a comment to an issue.
	ErrCommentFailed = errors.New("github: failed to post comment")

	// ErrCloseFailed indicates a failure closing an issue.
	ErrCloseFailed = errors.New("github: failed to close issue")

	// ErrLabelFailed indicates a failure adding a label to an issue.
	ErrLabelFailed = errors.New("github: failed to add label")

	// ErrMaxRetriesExceeded indicates all retry attempts have been exhausted.
	ErrMaxRetriesExceeded = errors.New("github: maximum retries exceeded")

	// ErrMappingExists indicates the issue is already linked to a SPEC.
	ErrMappingExists = errors.New("github: issue already linked to a SPEC")

	// ErrMappingNotFound indicates no SPEC is linked to the issue.
	ErrMappingNotFound = errors.New("github: no SPEC linked to issue")

	// ErrNilGHClient indicates a nil GHClient was provided to a constructor.
	ErrNilGHClient = errors.New("github: GHClient must not be nil")

	// ErrNilQualityGate indicates a nil quality.Gate was provided to a constructor.
	ErrNilQualityGate = errors.New("github: quality gate must not be nil")

	// ErrNilReviewer indicates a nil PRReviewer was provided to a constructor.
	ErrNilReviewer = errors.New("github: PRReviewer must not be nil")
)

// RetryError wraps the final error after all retry attempts are exhausted.
type RetryError struct {
	// Operation describes the operation that was retried.
	Operation string

	// Attempts is the total number of attempts made.
	Attempts int

	// LastError is the error from the final attempt.
	LastError error
}

// Error returns a human-readable description of the retry failure.
func (e *RetryError) Error() string {
	return fmt.Sprintf("github: %s failed after %d attempts: %v", e.Operation, e.Attempts, e.LastError)
}

// Unwrap returns the underlying error for errors.Is/errors.As chain.
func (e *RetryError) Unwrap() error {
	return e.LastError
}
