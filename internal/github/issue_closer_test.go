package github

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"testing"
	"time"
)

// commandCall records a single execGH invocation for assertion.
type commandCall struct {
	args []string
}

func TestIssueCloser_Close_Success(t *testing.T) {
	t.Parallel()

	var calls []commandCall
	exec := func(_ context.Context, _ string, args ...string) (string, error) {
		calls = append(calls, commandCall{args: args})
		return "", nil
	}

	closer := NewIssueCloser("/tmp/repo", WithExecFunc(exec), WithRetryDelay(0))
	result, err := closer.Close(context.Background(), 123, "Test comment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.CommentPosted {
		t.Error("CommentPosted should be true")
	}
	if !result.LabelAdded {
		t.Error("LabelAdded should be true")
	}
	if !result.IssueClosed {
		t.Error("IssueClosed should be true")
	}
	if result.IssueNumber != 123 {
		t.Errorf("IssueNumber = %d, want 123", result.IssueNumber)
	}

	// Verify 3 gh commands were issued: comment, edit --add-label, close
	if len(calls) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(calls))
	}
	assertArgsContain(t, calls[0].args, "comment")
	assertArgsContain(t, calls[1].args, "--add-label")
	assertArgsContain(t, calls[2].args, "close")
}

func TestIssueCloser_Close_CommentFails_AllRetries(t *testing.T) {
	t.Parallel()

	commentErr := fmt.Errorf("network timeout")
	exec := func(_ context.Context, _ string, args ...string) (string, error) {
		if slices.Contains(args, "comment") {
			return "", commentErr
		}
		return "", nil
	}

	closer := NewIssueCloser("/tmp/repo",
		WithExecFunc(exec),
		WithMaxRetries(3),
		WithRetryDelay(0),
	)

	result, err := closer.Close(context.Background(), 42, "Test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrMaxRetriesExceeded) {
		t.Errorf("error = %v, want ErrMaxRetriesExceeded", err)
	}
	if result.CommentPosted {
		t.Error("CommentPosted should be false")
	}
	// Label and close should not have been attempted since comment failed.
	if result.LabelAdded {
		t.Error("LabelAdded should be false when comment failed")
	}
	if result.IssueClosed {
		t.Error("IssueClosed should be false when comment failed")
	}
}

func TestIssueCloser_Close_RetryThenSuccess(t *testing.T) {
	t.Parallel()

	attempt := 0
	exec := func(_ context.Context, _ string, args ...string) (string, error) {
		for _, a := range args {
			if a == "comment" {
				attempt++
				if attempt < 3 {
					return "", fmt.Errorf("transient error")
				}
				return "", nil
			}
		}
		return "", nil
	}

	closer := NewIssueCloser("/tmp/repo",
		WithExecFunc(exec),
		WithMaxRetries(3),
		WithRetryDelay(0),
	)

	result, err := closer.Close(context.Background(), 99, "Retry test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.CommentPosted {
		t.Error("CommentPosted should be true after retry")
	}
	if !result.IssueClosed {
		t.Error("IssueClosed should be true")
	}
}

func TestIssueCloser_Close_LabelFails_CloseSucceeds(t *testing.T) {
	t.Parallel()

	exec := func(_ context.Context, _ string, args ...string) (string, error) {
		if slices.Contains(args, "--add-label") {
			return "", fmt.Errorf("label error")
		}
		return "", nil
	}

	closer := NewIssueCloser("/tmp/repo",
		WithExecFunc(exec),
		WithMaxRetries(1),
		WithRetryDelay(0),
	)

	result, err := closer.Close(context.Background(), 55, "Label test")
	// Should succeed overall; label failure is non-critical.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.CommentPosted {
		t.Error("CommentPosted should be true")
	}
	if result.LabelAdded {
		t.Error("LabelAdded should be false")
	}
	if !result.IssueClosed {
		t.Error("IssueClosed should be true despite label failure")
	}
}

func TestIssueCloser_Close_CloseFails_AllRetries(t *testing.T) {
	t.Parallel()

	exec := func(_ context.Context, _ string, args ...string) (string, error) {
		if slices.Contains(args, "close") {
			return "", fmt.Errorf("close error")
		}
		return "", nil
	}

	closer := NewIssueCloser("/tmp/repo",
		WithExecFunc(exec),
		WithMaxRetries(2),
		WithRetryDelay(0),
	)

	result, err := closer.Close(context.Background(), 77, "Close fail test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrMaxRetriesExceeded) {
		t.Errorf("error = %v, want ErrMaxRetriesExceeded", err)
	}
	if !result.CommentPosted {
		t.Error("CommentPosted should be true")
	}
	if result.IssueClosed {
		t.Error("IssueClosed should be false since close failed")
	}
}

func TestIssueCloser_Close_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	exec := func(c context.Context, _ string, _ ...string) (string, error) {
		return "", c.Err()
	}

	closer := NewIssueCloser("/tmp/repo",
		WithExecFunc(exec),
		WithRetryDelay(0),
	)

	_, err := closer.Close(ctx, 1, "Cancelled")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRetryError_Error(t *testing.T) {
	t.Parallel()

	inner := fmt.Errorf("connection reset")
	retryErr := &RetryError{
		Operation: "post comment",
		Attempts:  3,
		LastError: inner,
	}

	got := retryErr.Error()
	want := "github: post comment failed after 3 attempts: connection reset"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestRetryError_Unwrap(t *testing.T) {
	t.Parallel()

	inner := fmt.Errorf("inner error")
	retryErr := &RetryError{
		Operation: "close issue",
		Attempts:  2,
		LastError: inner,
	}

	if !errors.Is(retryErr, inner) {
		t.Error("Unwrap should expose inner error for errors.Is")
	}
}

func TestIssueCloser_Close_RetryDelay(t *testing.T) {
	t.Parallel()

	// Verify that delay is actually applied (non-zero delay should make it take longer).
	attempt := 0
	exec := func(_ context.Context, _ string, args ...string) (string, error) {
		for _, a := range args {
			if a == "comment" {
				attempt++
				if attempt < 2 {
					return "", fmt.Errorf("transient")
				}
				return "", nil
			}
		}
		return "", nil
	}

	closer := NewIssueCloser("/tmp/repo",
		WithExecFunc(exec),
		WithMaxRetries(3),
		WithRetryDelay(10*time.Millisecond),
	)

	start := time.Now()
	result, err := closer.Close(context.Background(), 1, "Delay test")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.CommentPosted {
		t.Error("CommentPosted should be true")
	}
	// With 1 retry and 10ms delay, it should take at least 10ms.
	if elapsed < 10*time.Millisecond {
		t.Errorf("elapsed = %v, expected at least 10ms for retry delay", elapsed)
	}
}

func TestIssueCloser_WithCloserLogger(t *testing.T) {
	t.Parallel()

	exec := func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", nil
	}

	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	closer := NewIssueCloser("/tmp/repo",
		WithExecFunc(exec),
		WithCloserLogger(logger),
		WithRetryDelay(0),
	)

	result, err := closer.Close(context.Background(), 10, "Logger test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IssueClosed {
		t.Error("IssueClosed should be true")
	}
}

func TestIssueCloser_Close_ContextCancelledDuringRetryWait(t *testing.T) {
	t.Parallel()

	// Cancel context while the closer is waiting in the retry delay.
	attempt := 0
	exec := func(_ context.Context, _ string, args ...string) (string, error) {
		if slices.Contains(args, "comment") {
			attempt++
			return "", fmt.Errorf("always fail")
		}
		return "", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	closer := NewIssueCloser("/tmp/repo",
		WithExecFunc(exec),
		WithMaxRetries(10),                   // Many retries to ensure we hit the delay.
		WithRetryDelay(500*time.Millisecond), // Long delay so context cancels during wait.
	)

	_, err := closer.Close(ctx, 1, "Cancel during wait")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestIssueCloser_WithMaxRetries_IgnoresZero(t *testing.T) {
	t.Parallel()

	exec := func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", nil
	}

	// MaxRetries of 0 should be ignored, keeping default of 3.
	closer := NewIssueCloser("/tmp/repo", WithExecFunc(exec), WithMaxRetries(0))
	if closer.maxRetries != 3 {
		t.Errorf("maxRetries = %d, want 3 (default preserved)", closer.maxRetries)
	}
}

func TestIssueCloser_Close_InvalidIssueNumber(t *testing.T) {
	t.Parallel()

	exec := func(_ context.Context, _ string, _ ...string) (string, error) {
		t.Fatal("exec should not be called for invalid issue numbers")
		return "", nil
	}

	closer := NewIssueCloser("/tmp/repo", WithExecFunc(exec), WithRetryDelay(0))

	tests := []struct {
		name   string
		number int
	}{
		{"zero", 0},
		{"negative", -1},
		{"large negative", -999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := closer.Close(context.Background(), tt.number, "comment")
			if err == nil {
				t.Fatal("expected error for invalid issue number")
			}
			if result != nil {
				t.Error("result should be nil for invalid issue number")
			}
			want := fmt.Sprintf("close issue: invalid issue number %d", tt.number)
			if err.Error() != want {
				t.Errorf("error = %q, want %q", err.Error(), want)
			}
		})
	}
}

// assertArgsContain checks that at least one arg matches the target.
func assertArgsContain(t *testing.T, args []string, target string) {
	t.Helper()
	if slices.Contains(args, target) {
		return
	}
	t.Errorf("args %v does not contain %q", args, target)
}
