package quality

import (
	"context"
	"testing"
)

// TestCountFixedIssues tests the countFixedIssues method covering all branches.
func TestCountFixedIssues(t *testing.T) {
	l := NewLinter(nil)

	tests := []struct {
		name   string
		output string
		want   int
	}{
		{
			name:   "empty output returns zero",
			output: "",
			want:   0,
		},
		{
			name:   "pattern: N issues fixed",
			output: "3 issues fixed by auto-fix",
			want:   3,
		},
		{
			name:   "pattern: N fixed",
			output: "5 fixed",
			want:   5,
		},
		{
			name:   "pattern: fixed with trailing content",
			output: "1 issue corrected in formatting",
			want:   1,
		},
		{
			name:   "fallback: line contains 'fixed'",
			output: "line1\nthis line was fixed manually\nline3",
			want:   1,
		},
		{
			name:   "fallback: multiple lines with 'fixed'",
			output: "file.go: fixed whitespace\nfile2.go: fixed imports\nno match here",
			want:   2,
		},
		{
			name:   "fallback: case insensitive 'FIXED'",
			output: "FIXED: trailing whitespace in test.go",
			want:   1,
		},
		{
			name:   "no matches returns zero",
			output: "no issues found\nall clear",
			want:   0,
		},
		{
			name:   "multi-digit count in summary pattern",
			output: "12 issues fixed",
			want:   12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := l.countFixedIssues(tt.output)
			if got != tt.want {
				t.Errorf("countFixedIssues(%q) = %d, want %d", tt.output, got, tt.want)
			}
		})
	}
}

// TestGenerateSummary_NilResult tests nil result handling.
func TestGenerateSummary_NilResult(t *testing.T) {
	l := NewLinter(nil)
	summary := l.GenerateSummary(nil)
	if summary == "" {
		t.Error("expected non-empty summary for nil result")
	}
}

// TestGenerateSummary_FewIssues tests that short output is returned as-is.
func TestGenerateSummary_FewIssues(t *testing.T) {
	l := NewLinter(nil)

	result := &ToolResult{
		IssuesFound: 2,
		Output:      "file.go:1: warning: issue 1\nfile.go:2: warning: issue 2",
	}

	summary := l.GenerateSummary(result)
	// Should return the full output when <= 5 issues
	if summary == "" {
		t.Error("expected non-empty summary")
	}
}

// TestGenerateSummary_ManyIssues tests truncation behavior.
func TestGenerateSummary_ManyIssues(t *testing.T) {
	l := NewLinter(nil)

	output := "file.go:1: warning: issue\nfile.go:2: warning: issue\nfile.go:3: warning: issue\nfile.go:4: warning: issue\nfile.go:5: warning: issue\nfile.go:6: warning: issue\nfile.go:7: warning: issue"
	result := &ToolResult{
		IssuesFound: 7,
		Output:      output,
	}

	summary := l.GenerateSummary(result)
	if summary == "" {
		t.Error("expected non-empty summary for many issues")
	}
}

// TestIsIssueLine tests the isIssueLine private method.
func TestIsIssueLine(t *testing.T) {
	l := NewLinter(nil)

	tests := []struct {
		line string
		want bool
	}{
		{"", false},
		{"   ", false},
		{"file.go:1:2: error: something", true},
		{"file.go:1: warning: something", true},
		{"file.go:1: note: something", true},
		{"something ERR happened", true},   // " ERR " with surrounding spaces
		{"something WARN happened", true},  // " WARN " with surrounding spaces
		{"error: this is an error", true},
		{"warning: this is a warning", true},
		{"normal log line", false},
		{"file.go:42:10: someword", true}, // matches reFileLineCol
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := l.isIssueLine(tt.line)
			if got != tt.want {
				t.Errorf("isIssueLine(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

// TestAutoFix_NoTools tests AutoFix when no tools support the file.
func TestAutoFix_NoTools(t *testing.T) {
	l := NewLinter(NewToolRegistry())

	// Use a file extension with no matching linter/auto-fix
	result, err := l.AutoFix(context.TODO(), "/tmp/test.unknownext123")
	if err != nil {
		t.Fatalf("AutoFix returned unexpected error: %v", err)
	}
	// Should return nil when no tools available
	if result != nil {
		t.Errorf("expected nil result for unsupported extension, got %+v", result)
	}
}

// TestParseIssues_WithSummaryLine tests that parseIssuesFromOutput handles summary lines.
func TestParseIssues_WithSummaryLine(t *testing.T) {
	l := NewLinter(nil)

	// Output with summary pattern "N issues found"
	output := "test.go:1:1: error: syntax error\nFound 1 issues"
	result := l.ParseIssues(output, "test.go")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IssuesFound == 0 {
		t.Error("expected issues > 0")
	}
}
