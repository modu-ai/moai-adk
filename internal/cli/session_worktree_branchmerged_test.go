package cli

// session_worktree_branchmerged_test.go — SPEC-WORKTREE-REAPER-001 post-audit
// repair F1: a parser-level battery over `git branch --merged <base>` stdout.
//
// Why this file exists at all: AC-WR-002's fixture swaps the
// sessionWorktreeGitBranchMerged seam and feeds fabricated, marker-free branch
// names, so nothing in the suite ever ran the REAL parser. Measured on this
// repository at the time of the repair, `git branch --merged origin/main`
// emitted 149 entries of which 119 carried git's `+` linked-worktree marker —
// and `+` is precisely the marker every candidate this sweep evaluates carries,
// because a candidate is by definition checked out in a linked worktree.

import (
	"reflect"
	"testing"
)

// TestParseGitBranchMergedOutput binds the real parser against realistic
// stdout: the current-branch marker, the linked-worktree marker, plain
// entries, and the two non-branch lines git emits for detached checkouts.
func TestParseGitBranchMergedOutput(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want []string
	}{
		{
			name: "current-branch marker is stripped",
			out:  "* main\n  WT-plain\n",
			want: []string{"main", "WT-plain"},
		},
		{
			name: "linked-worktree marker is stripped",
			out:  "+ WT-agent-toml-dual\n",
			want: []string{"WT-agent-toml-dual"},
		},
		{
			name: "mixed markers in one listing",
			out:  "* main\n+ WT-worktree-reaper\n  WT-docs-redesign\n",
			want: []string{"main", "WT-worktree-reaper", "WT-docs-redesign"},
		},
		{
			name: "detached entries are not branch names",
			out:  "* (HEAD detached at 301841e0f)\n+ (no branch)\n  WT-real\n",
			want: []string{"WT-real"},
		},
		{
			name: "empty stdout yields no branches",
			out:  "\n",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseGitBranchMergedOutput(tt.out)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseGitBranchMergedOutput(%q) = %#v, want %#v", tt.out, got, tt.want)
			}
		})
	}
}

// TestBranchMergedForCleanup_LinkedWorktreeMarker binds the DEFECT at the
// decision level rather than the parser level: gh gives no answer, the git
// fallback is consulted (REQ-WR-002), and the branch it must recognise is one
// checked out in a linked worktree — the `+`-prefixed form.
func TestBranchMergedForCleanup_LinkedWorktreeMarker(t *testing.T) {
	orig := sessionWorktreeGitBranchMerged
	t.Cleanup(func() { sessionWorktreeGitBranchMerged = orig })
	// Route the seam through the REAL parser so the fixture cannot bypass it.
	sessionWorktreeGitBranchMerged = func() ([]string, error) {
		return parseGitBranchMergedOutput("* main\n+ WT-worktree-reaper\n"), nil
	}

	if got := branchMergedForCleanup("WT-worktree-reaper", false); got != mergeStateMerged {
		t.Errorf("a branch checked out in a linked worktree must be recognised as merged; got %v", got)
	}
}
