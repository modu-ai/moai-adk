package ciwatch_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/ciwatch"
)

// TestFormatStatusUpdate validates the 30-second status report formatter
// per AC-CIAUT-004 and strategy W2-T05.
func TestFormatStatusUpdate(t *testing.T) {
	tests := []struct {
		name       string
		state      ciwatch.CIState
		wantSubstr []string
		noANSI     bool
		maxLen     int
	}{
		{
			name: "all_pending",
			state: ciwatch.CIState{
				PRNumber:        785,
				Branch:          "main",
				RequiredPassed:  0,
				RequiredFailed:  nil,
				RequiredPending: []ciwatch.CheckResult{
					{Name: "Lint", Status: "in_progress"},
					{Name: "Test (ubuntu-latest)", Status: "queued"},
					{Name: "CodeQL", Status: "queued"},
				},
				AuxiliaryFailed: nil,
			},
			wantSubstr: []string{"PR #785", "0/3", "3 pending", "advisory 0 fail"},
			noANSI:     true,
			maxLen:     200,
		},
		{
			name: "partial_pass",
			state: ciwatch.CIState{
				PRNumber:       786,
				Branch:         "main",
				RequiredPassed: 4,
				RequiredFailed: nil,
				RequiredPending: []ciwatch.CheckResult{
					{Name: "CodeQL", Status: "in_progress"},
					{Name: "Build (linux/amd64)", Status: "in_progress"},
				},
				AuxiliaryFailed: []ciwatch.CheckResult{
					{Name: "claude-code-review", Status: "completed", Conclusion: "failure"},
				},
			},
			wantSubstr: []string{"PR #786", "4/6", "2 pending", "advisory 1 fail"},
			noANSI:     true,
			maxLen:     200,
		},
		{
			name: "fail_imminent",
			state: ciwatch.CIState{
				PRNumber:       787,
				Branch:         "feat/test",
				RequiredPassed: 5,
				RequiredFailed: []ciwatch.CheckResult{
					{Name: "Lint", Status: "completed", Conclusion: "failure"},
				},
				RequiredPending: nil,
				AuxiliaryFailed: nil,
			},
			wantSubstr: []string{"PR #787", "5/6", "1 failed", "advisory 0 fail"},
			noANSI:     true,
			maxLen:     200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ciwatch.FormatStatusUpdate(tt.state)

			// No ANSI escape codes.
			if tt.noANSI && containsANSI(got) {
				t.Errorf("output contains ANSI codes: %q", got)
			}

			// Max line length.
			if tt.maxLen > 0 && len(got) > tt.maxLen {
				t.Errorf("output length %d exceeds max %d: %q", len(got), tt.maxLen, got)
			}

			// Required substrings.
			for _, sub := range tt.wantSubstr {
				if !strings.Contains(got, sub) {
					t.Errorf("output missing %q: %q", sub, got)
				}
			}
		})
	}
}

// containsANSI returns true if s contains any ANSI escape sequence.
func containsANSI(s string) bool {
	return strings.Contains(s, "\x1b[")
}
