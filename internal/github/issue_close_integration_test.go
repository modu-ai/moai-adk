package github

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/i18n"
)

// TestIntegration_CommentGenerator_To_IssueCloser verifies the end-to-end flow:
// generate a multilingual comment via i18n.CommentGenerator, then pass it to
// IssueCloser.Close() and verify all 3 gh CLI commands are executed correctly.
func TestIntegration_CommentGenerator_To_IssueCloser(t *testing.T) {
	t.Parallel()

	// Step 1: Generate a Korean comment.
	gen := i18n.NewCommentGenerator()
	data := &i18n.CommentData{
		Summary:         "Added user authentication feature",
		PRNumber:        456,
		IssueNumber:     123,
		MergedAt:        time.Date(2026, 2, 16, 16, 30, 0, 0, time.UTC),
		TimeZone:        "KST",
		CoveragePercent: 92,
	}

	comment, err := gen.Generate("ko", data)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	// Verify the comment is Korean.
	if !strings.Contains(comment, "성공적으로 해결") {
		t.Fatalf("expected Korean comment, got:\n%s", comment)
	}

	// Step 2: Pass the generated comment to IssueCloser.
	var calls []ghCall
	mockExec := func(_ context.Context, _ string, args ...string) (string, error) {
		calls = append(calls, ghCall{args: args})
		return "", nil
	}

	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	closer := NewIssueCloser("/tmp/repo",
		WithExecFunc(mockExec),
		WithRetryDelay(0),
		WithCloserLogger(logger),
	)

	result, err := closer.Close(context.Background(), 123, comment)
	if err != nil {
		t.Fatalf("Close() error: %v", err)
	}

	// Step 3: Verify the 3-step execution.
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

	// Verify exactly 3 gh commands.
	if len(calls) != 3 {
		t.Fatalf("expected 3 gh calls, got %d", len(calls))
	}

	// Call 1: gh issue comment 123 --body <Korean comment>
	assertGHCallContains(t, calls[0], "comment")
	assertGHCallContains(t, calls[0], "123")

	// Verify the Korean comment was passed to gh.
	commentBody := findArg(calls[0].args, "--body")
	if commentBody == "" {
		t.Error("missing --body in comment call")
	} else if !strings.Contains(commentBody, "성공적으로 해결") {
		t.Error("comment body should contain Korean text")
	}

	// Call 2: gh issue edit 123 --add-label resolved
	assertGHCallContains(t, calls[1], "--add-label")
	assertGHCallContains(t, calls[1], "resolved")

	// Call 3: gh issue close 123
	assertGHCallContains(t, calls[2], "close")
	assertGHCallContains(t, calls[2], "123")
}

// TestIntegration_MultilingualComments tests comment generation for all 4
// supported languages plus fallback behavior, then verifies each comment
// can be passed to IssueCloser without issues.
func TestIntegration_MultilingualComments(t *testing.T) {
	t.Parallel()

	gen := i18n.NewCommentGenerator()
	data := &i18n.CommentData{
		Summary:         "Fixed login redirect bug",
		PRNumber:        789,
		IssueNumber:     42,
		MergedAt:        time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC),
		TimeZone:        "UTC",
		CoveragePercent: 88,
	}

	tests := []struct {
		lang     string
		contains string
	}{
		{"en", "resolved successfully"},
		{"ko", "성공적으로 해결"},
		{"ja", "正常に解決されました"},
		{"zh", "已成功解决"},
		{"de", "resolved successfully"}, // Fallback to English.
		{"", "resolved successfully"},   // Empty lang fallback.
	}

	for _, tt := range tests {
		t.Run("lang_"+tt.lang, func(t *testing.T) {
			t.Parallel()

			comment, err := gen.Generate(tt.lang, data)
			if err != nil {
				t.Fatalf("Generate(%q) error: %v", tt.lang, err)
			}

			if !strings.Contains(comment, tt.contains) {
				t.Errorf("comment for lang %q missing %q\ngot:\n%s", tt.lang, tt.contains, comment)
			}

			// Verify the comment contains the PR link and timestamp.
			if !strings.Contains(comment, "#789") {
				t.Errorf("comment missing PR link #789")
			}
			if !strings.Contains(comment, "2026-02-16") {
				t.Errorf("comment missing date")
			}

			// Verify the comment can be passed to IssueCloser without error.
			mockExec := func(_ context.Context, _ string, _ ...string) (string, error) {
				return "", nil
			}
			closer := NewIssueCloser("/tmp/repo",
				WithExecFunc(mockExec),
				WithRetryDelay(0),
			)

			result, err := closer.Close(context.Background(), 42, comment)
			if err != nil {
				t.Fatalf("Close() error: %v", err)
			}
			if !result.IssueClosed {
				t.Error("IssueClosed should be true")
			}
		})
	}
}

// TestIntegration_IssueClose_RetryRecovery verifies the retry mechanism
// with exponential backoff when comment posting fails transiently.
func TestIntegration_IssueClose_RetryRecovery(t *testing.T) {
	t.Parallel()

	gen := i18n.NewCommentGenerator()
	comment, err := gen.Generate("en", &i18n.CommentData{
		Summary:     "Retry recovery test",
		PRNumber:    100,
		IssueNumber: 50,
		MergedAt:    time.Now(),
		TimeZone:    "UTC",
	})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	// Mock: comment fails 2 times, then succeeds.
	commentAttempt := 0
	mockExec := func(_ context.Context, _ string, args ...string) (string, error) {
		for _, a := range args {
			if a == "comment" {
				commentAttempt++
				if commentAttempt <= 2 {
					return "", fmt.Errorf("transient network error")
				}
				return "", nil
			}
		}
		return "", nil
	}

	closer := NewIssueCloser("/tmp/repo",
		WithExecFunc(mockExec),
		WithMaxRetries(5),
		WithRetryDelay(1*time.Millisecond), // Fast retries for test.
	)

	start := time.Now()
	result, err := closer.Close(context.Background(), 50, comment)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Close() should succeed after retry, got: %v", err)
	}

	if !result.CommentPosted {
		t.Error("CommentPosted should be true after retry recovery")
	}
	if !result.IssueClosed {
		t.Error("IssueClosed should be true")
	}

	// Verify that retries actually happened (comment was attempted 3 times).
	if commentAttempt != 3 {
		t.Errorf("commentAttempt = %d, want 3 (2 failures + 1 success)", commentAttempt)
	}

	// Verify exponential backoff introduced some delay.
	// With 1ms base and 2 retries: ~1ms + ~2ms = ~3ms minimum.
	if elapsed < 1*time.Millisecond {
		t.Errorf("elapsed = %v, expected at least 1ms for backoff delays", elapsed)
	}
}

// TestIntegration_IssueClose_PartialFailure verifies that label failure
// (non-critical) does not prevent issue closure.
func TestIntegration_IssueClose_PartialFailure(t *testing.T) {
	t.Parallel()

	gen := i18n.NewCommentGenerator()
	comment, err := gen.Generate("ja", &i18n.CommentData{
		Summary:         "Partial failure test",
		PRNumber:        200,
		IssueNumber:     75,
		MergedAt:        time.Now(),
		TimeZone:        "JST",
		CoveragePercent: 95,
	})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	// Verify Japanese comment was generated.
	if !strings.Contains(comment, "解決されました") {
		t.Fatalf("expected Japanese comment, got:\n%s", comment)
	}

	// Mock: label operation always fails, others succeed.
	var calls []ghCall
	mockExec := func(_ context.Context, _ string, args ...string) (string, error) {
		calls = append(calls, ghCall{args: args})
		if slices.Contains(args, "--add-label") {
			return "", fmt.Errorf("label 'resolved' does not exist")
		}
		return "", nil
	}

	closer := NewIssueCloser("/tmp/repo",
		WithExecFunc(mockExec),
		WithMaxRetries(1), // No retries for label to speed up test.
		WithRetryDelay(0),
	)

	result, err := closer.Close(context.Background(), 75, comment)
	if err != nil {
		t.Fatalf("Close() should succeed despite label failure, got: %v", err)
	}

	// Verify partial result.
	if !result.CommentPosted {
		t.Error("CommentPosted should be true")
	}
	if result.LabelAdded {
		t.Error("LabelAdded should be false (label operation failed)")
	}
	if !result.IssueClosed {
		t.Error("IssueClosed should be true (closure proceeds despite label failure)")
	}
	if result.IssueNumber != 75 {
		t.Errorf("IssueNumber = %d, want 75", result.IssueNumber)
	}

	// Verify all 3 gh commands were attempted (label failure doesn't skip close).
	if len(calls) != 3 {
		t.Fatalf("expected 3 gh calls (comment, label attempt, close), got %d", len(calls))
	}
	assertGHCallContains(t, calls[0], "comment")
	assertGHCallContains(t, calls[1], "--add-label")
	assertGHCallContains(t, calls[2], "close")
}

// --- Integration Test Helpers ---

// ghCall records a single gh CLI invocation.
type ghCall struct {
	args []string
}

// assertGHCallContains verifies that at least one argument matches target.
func assertGHCallContains(t *testing.T, call ghCall, target string) {
	t.Helper()
	if slices.Contains(call.args, target) {
		return
	}
	t.Errorf("gh call args %v does not contain %q", call.args, target)
}

// findArg returns the argument following a flag (e.g., "--body" returns the next arg).
func findArg(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}
