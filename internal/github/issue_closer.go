package github

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

// ExecFunc is the function signature for executing gh CLI commands.
// It matches the signature of execGH for seamless integration.
type ExecFunc func(ctx context.Context, dir string, args ...string) (string, error)

// CloseResult holds the outcome of an issue closure operation.
type CloseResult struct {
	// IssueNumber is the GitHub issue that was processed.
	IssueNumber int

	// CommentPosted indicates whether the success comment was posted.
	CommentPosted bool

	// LabelAdded indicates whether the resolved label was added.
	LabelAdded bool

	// IssueClosed indicates whether the issue was closed.
	IssueClosed bool
}

// IssueCloser posts a success comment, adds a resolved label,
// and closes a GitHub issue.
type IssueCloser interface {
	// Close posts a comment, adds the resolved label, and closes the issue.
	// Returns a partial CloseResult even on failure to indicate which steps succeeded.
	Close(ctx context.Context, issueNumber int, comment string) (*CloseResult, error)
}

// DefaultIssueCloser implements IssueCloser using the gh CLI.
type DefaultIssueCloser struct {
	root       string
	exec       ExecFunc
	maxRetries int
	retryDelay time.Duration
	logger     *slog.Logger
}

// Compile-time interface compliance check.
var _ IssueCloser = (*DefaultIssueCloser)(nil)

// IssueCloserOption configures a DefaultIssueCloser.
type IssueCloserOption func(*DefaultIssueCloser)

// WithExecFunc sets a custom execution function (used for testing).
func WithExecFunc(fn ExecFunc) IssueCloserOption {
	return func(c *DefaultIssueCloser) {
		c.exec = fn
	}
}

// WithMaxRetries sets the maximum number of retry attempts per operation.
func WithMaxRetries(n int) IssueCloserOption {
	return func(c *DefaultIssueCloser) {
		if n > 0 {
			c.maxRetries = n
		}
	}
}

// WithRetryDelay sets the base delay between retry attempts.
// Actual delay doubles with each attempt (exponential backoff).
// Zero disables backoff; negative values are ignored.
func WithRetryDelay(d time.Duration) IssueCloserOption {
	return func(c *DefaultIssueCloser) {
		if d >= 0 {
			c.retryDelay = d
		}
	}
}

// WithCloserLogger sets the logger for the issue closer.
func WithCloserLogger(l *slog.Logger) IssueCloserOption {
	return func(c *DefaultIssueCloser) {
		c.logger = l
	}
}

// NewIssueCloser creates a new DefaultIssueCloser rooted at the given directory.
func NewIssueCloser(root string, opts ...IssueCloserOption) *DefaultIssueCloser {
	c := &DefaultIssueCloser{
		root:       root,
		exec:       execGH,
		maxRetries: 3,
		retryDelay: 2 * time.Second,
		logger:     slog.Default().With("module", "github.issue_closer"),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Close posts a comment, adds the "resolved" label, and closes the issue.
// It retries each step up to maxRetries times with exponential backoff.
//
// The operation is sequential: comment -> label -> close.
// If comment posting fails, label and close are skipped.
// If label fails, close is still attempted (label is non-critical).
// The result tracks which steps succeeded for partial failure recovery.
func (c *DefaultIssueCloser) Close(ctx context.Context, issueNumber int, comment string) (*CloseResult, error) {
	if issueNumber <= 0 {
		return nil, fmt.Errorf("close issue: invalid issue number %d", issueNumber)
	}

	result := &CloseResult{
		IssueNumber: issueNumber,
	}

	num := strconv.Itoa(issueNumber)

	// Step 1: Post success comment (critical).
	if err := c.retryOp(ctx, "post comment", func() error {
		_, execErr := c.exec(ctx, c.root, "issue", "comment", num, "--body", comment)
		return execErr
	}); err != nil {
		return result, fmt.Errorf("close issue #%d: %w", issueNumber, err)
	}
	result.CommentPosted = true

	// Step 2: Add "resolved" label (non-critical).
	if err := c.retryOp(ctx, "add label", func() error {
		_, execErr := c.exec(ctx, c.root, "issue", "edit", num, "--add-label", "resolved")
		return execErr
	}); err != nil {
		c.logger.Warn("failed to add resolved label",
			"issue", issueNumber,
			"error", err,
		)
		// Continue to close; label failure is non-critical.
	} else {
		result.LabelAdded = true
	}

	// Step 3: Close issue (critical).
	if err := c.retryOp(ctx, "close issue", func() error {
		_, execErr := c.exec(ctx, c.root, "issue", "close", num)
		return execErr
	}); err != nil {
		return result, fmt.Errorf("close issue #%d: %w", issueNumber, err)
	}
	result.IssueClosed = true

	c.logger.Info("issue closed successfully",
		"issue", issueNumber,
		"comment_posted", result.CommentPosted,
		"label_added", result.LabelAdded,
	)

	return result, nil
}

// retryOp executes an operation with exponential backoff retry.
// Returns ErrMaxRetriesExceeded (wrapped in RetryError) if all attempts fail.
func (c *DefaultIssueCloser) retryOp(ctx context.Context, operation string, fn func() error) error {
	var lastErr error
	delay := c.retryDelay

	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		c.logger.Debug("operation failed, retrying",
			"operation", operation,
			"attempt", attempt,
			"max_retries", c.maxRetries,
			"error", lastErr,
		)

		if attempt < c.maxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			delay *= 2 // Exponential backoff.
		}
	}

	return &RetryError{
		Operation: operation,
		Attempts:  c.maxRetries,
		LastError: fmt.Errorf("%w: %v", ErrMaxRetriesExceeded, lastErr),
	}
}
