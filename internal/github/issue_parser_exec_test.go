package github

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
)

// TestParseIssue_InvalidNumber tests ParseIssue with invalid (non-positive) issue numbers.
func TestParseIssue_InvalidNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		number int
	}{
		{name: "zero", number: 0},
		{name: "negative", number: -1},
		{name: "large negative", number: -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// execFn should never be called for invalid numbers.
			parser := newIssueParserWithExec("/tmp/repo", func(_ context.Context, _ string, _ ...string) (string, error) {
				t.Error("execFn should not be called for invalid issue numbers")
				return "", nil
			})

			_, err := parser.ParseIssue(context.Background(), tt.number)
			if err == nil {
				t.Fatalf("ParseIssue(%d) expected error, got nil", tt.number)
			}
		})
	}
}

// TestParseIssue_Success tests successful issue parsing.
func TestParseIssue_Success(t *testing.T) {
	t.Parallel()

	issueJSON := `{
		"number": 42,
		"title": "Fix the bug",
		"body": "Detailed description",
		"labels": [{"name": "bug"}, {"name": "priority:high"}],
		"author": {"login": "bugfinder"},
		"comments": [{"body": "Confirmed", "author": {"login": "ops"}}]
	}`

	var capturedArgs []string
	parser := newIssueParserWithExec("/tmp/repo", func(_ context.Context, _ string, args ...string) (string, error) {
		capturedArgs = args
		return issueJSON, nil
	})

	issue, err := parser.ParseIssue(context.Background(), 42)
	if err != nil {
		t.Fatalf("ParseIssue() error = %v", err)
	}

	if issue.Number != 42 {
		t.Errorf("Number = %d, want 42", issue.Number)
	}
	if issue.Title != "Fix the bug" {
		t.Errorf("Title = %q, want %q", issue.Title, "Fix the bug")
	}
	if issue.Body != "Detailed description" {
		t.Errorf("Body = %q, want %q", issue.Body, "Detailed description")
	}
	if len(issue.Labels) != 2 {
		t.Errorf("Labels count = %d, want 2", len(issue.Labels))
	}
	if issue.Author.Login != "bugfinder" {
		t.Errorf("Author.Login = %q, want %q", issue.Author.Login, "bugfinder")
	}
	if len(issue.Comments) != 1 {
		t.Errorf("Comments count = %d, want 1", len(issue.Comments))
	}

	// Verify exec was called with correct args.
	if len(capturedArgs) < 4 {
		t.Fatalf("exec called with too few args: %v", capturedArgs)
	}
	if capturedArgs[0] != "issue" || capturedArgs[1] != "view" {
		t.Errorf("exec args[0:2] = %v, want [issue view]", capturedArgs[0:2])
	}
	if capturedArgs[2] != "42" {
		t.Errorf("exec args[2] = %q, want %q", capturedArgs[2], "42")
	}
}

// TestParseIssue_ExecError tests ParseIssue when exec fails.
func TestParseIssue_ExecError(t *testing.T) {
	t.Parallel()

	parser := newIssueParserWithExec("/tmp/repo", func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("gh: issue view failed")
	})

	_, err := parser.ParseIssue(context.Background(), 1)
	if err == nil {
		t.Fatal("ParseIssue() expected error, got nil")
	}
}

// TestParseIssue_InvalidJSON tests ParseIssue when JSON response is invalid.
func TestParseIssue_InvalidJSON(t *testing.T) {
	t.Parallel()

	parser := newIssueParserWithExec("/tmp/repo", func(_ context.Context, _ string, _ ...string) (string, error) {
		return "{invalid json", nil
	})

	_, err := parser.ParseIssue(context.Background(), 5)
	if err == nil {
		t.Fatal("ParseIssue() expected error for invalid JSON, got nil")
	}
}

// TestParseIssue_EmptyTitle tests ParseIssue when title is empty.
func TestParseIssue_EmptyTitle(t *testing.T) {
	t.Parallel()

	parser := newIssueParserWithExec("/tmp/repo", func(_ context.Context, _ string, _ ...string) (string, error) {
		return `{"number": 10, "title": "", "body": "some body"}`, nil
	})

	_, err := parser.ParseIssue(context.Background(), 10)
	if err == nil {
		t.Fatal("ParseIssue() expected error for empty title, got nil")
	}
}

// TestParseIssue_VerifiesJSONFields tests that all expected JSON fields are requested.
func TestParseIssue_VerifiesJSONFields(t *testing.T) {
	t.Parallel()

	var capturedArgs []string
	parser := newIssueParserWithExec("/tmp/repo", func(_ context.Context, _ string, args ...string) (string, error) {
		capturedArgs = args
		return `{"number": 7, "title": "Test issue"}`, nil
	})

	_, err := parser.ParseIssue(context.Background(), 7)
	if err != nil {
		t.Fatalf("ParseIssue() error = %v", err)
	}

	// Find --json flag and verify the fields.
	for i, arg := range capturedArgs {
		if arg == "--json" && i+1 < len(capturedArgs) {
			fields := capturedArgs[i+1]
			if fields != issueFields {
				t.Errorf("--json fields = %q, want %q", fields, issueFields)
			}
			return
		}
	}
	t.Errorf("--json flag not found in args: %v", capturedArgs)
}

// TestParseIssue_LargeIssueNumber tests ParseIssue with a large issue number.
func TestParseIssue_LargeIssueNumber(t *testing.T) {
	t.Parallel()

	parser := newIssueParserWithExec("/tmp/repo", func(_ context.Context, _ string, args ...string) (string, error) {
		// Verify the number is correctly stringified.
		if slices.Contains(args, "99999") {
			return `{"number": 99999, "title": "Large number issue"}`, nil
		}
		return "", fmt.Errorf("did not find expected number in args: %v", args)
	})

	issue, err := parser.ParseIssue(context.Background(), 99999)
	if err != nil {
		t.Fatalf("ParseIssue() error = %v", err)
	}
	if issue.Number != 99999 {
		t.Errorf("Number = %d, want 99999", issue.Number)
	}
}
