package git

import "testing"

func TestDetectBranchPrefix(t *testing.T) {
	tests := []struct {
		name   string
		labels []string
		want   string
	}{
		{"bug label", []string{"bug"}, "fix/"},
		{"feature label", []string{"feature"}, "feat/"},
		{"enhancement label", []string{"enhancement"}, "feat/"},
		{"documentation label", []string{"documentation"}, "docs/"},
		{"docs label", []string{"docs"}, "docs/"},
		{"no labels", []string{}, "feat/"},
		{"nil labels", nil, "feat/"},
		{"unknown label", []string{"question"}, "feat/"},
		{"multiple labels bug first", []string{"bug", "feature"}, "fix/"},
		{"multiple labels feature first", []string{"feature", "bug"}, "feat/"},
		{"multiple unknown labels", []string{"help wanted", "good first issue"}, "feat/"},
		{"mixed known and unknown", []string{"help wanted", "bug"}, "fix/"},
		{"docs among others", []string{"question", "documentation"}, "docs/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectBranchPrefix(tt.labels)
			if got != tt.want {
				t.Errorf("DetectBranchPrefix(%v) = %q, want %q", tt.labels, got, tt.want)
			}
		})
	}
}

func TestFormatIssueBranch(t *testing.T) {
	tests := []struct {
		name        string
		labels      []string
		issueNumber int
		want        string
	}{
		{"bug issue", []string{"bug"}, 123, "fix/issue-123"},
		{"feature issue", []string{"feature"}, 456, "feat/issue-456"},
		{"enhancement issue", []string{"enhancement"}, 789, "feat/issue-789"},
		{"docs issue", []string{"documentation"}, 42, "docs/issue-42"},
		{"no labels", []string{}, 100, "feat/issue-100"},
		{"unknown labels", []string{"question"}, 1, "feat/issue-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatIssueBranch(tt.labels, tt.issueNumber)
			if err != nil {
				t.Fatalf("FormatIssueBranch(%v, %d) error = %v", tt.labels, tt.issueNumber, err)
			}
			if got != tt.want {
				t.Errorf("FormatIssueBranch(%v, %d) = %q, want %q", tt.labels, tt.issueNumber, got, tt.want)
			}
		})
	}
}

func TestFormatIssueBranch_InvalidNumber(t *testing.T) {
	tests := []struct {
		name        string
		issueNumber int
	}{
		{"zero", 0},
		{"negative", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FormatIssueBranch(nil, tt.issueNumber)
			if err == nil {
				t.Errorf("FormatIssueBranch(nil, %d) expected error, got nil", tt.issueNumber)
			}
		})
	}
}
