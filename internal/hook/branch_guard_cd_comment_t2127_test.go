package hook

// branch_guard_cd_comment_t2127_test.go — card t2127 (PR #1969 Codex P2).
// cdPrimaryTarget scanned the quote-collapsed command without eliding shell
// comments, so a `cd <primary>` that only appears in a trailing comment was
// returned as the command's target and a legitimate worktree commit was judged
// against the primary checkout's protected branch. The cd scan must read the
// command the shell runs, not its comments; a `#` inside quotes stays data.

import "testing"

// TestCdPrimaryTargetIgnoresCdInShellComment pins t2127: a cd word that exists
// only inside a comment is no target, while a real cd word (before or after the
// git command, or next to a quoted `#`) still names the primary checkout.
func TestCdPrimaryTargetIgnoresCdInShellComment(t *testing.T) {
	skipWindowsPOSIXCommandText(t)
	primary, worktree := worktreeFixture(t)

	cases := []struct {
		name    string
		command string
		want    bool
	}{
		// Comment-only cd: no target (the defect).
		{"trailing comment", "git commit -m x # cd " + primary, false},
		{"comment after quoted message", `git commit -m "it's" # cd ` + primary, false},
		{"whole-line comment on a later line", "git commit -m x\n# cd " + primary, false},
		// Positive controls: a real cd word is still found.
		{"cd before commit", "cd " + primary + " && git commit -m x", true},
		{"cd after commit", "git commit -m x; cd " + primary, true},
		{"real cd then comment", "cd " + primary + " && git commit -m x # note", true},
		// A `#` inside quotes is data and opens no comment: the real cd after it
		// is still scanned.
		{"hash inside quotes then cd", `git commit -m "a # b"; cd ` + primary, true},
		{"cd inside quotes is data", `git commit -m "x; cd ` + primary + `"`, false},
		// A mid-word `#` is not a comment either.
		{"mid-word hash then cd", "echo a#b; cd " + primary, true},
		// A worktree target never classifies as primary.
		{"cd to worktree", "cd " + worktree + " && git commit -m x", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, ok := cdPrimaryTarget(tc.command)
			if ok != tc.want {
				t.Fatalf("cdPrimaryTarget(%q) = (%q, %v), want ok=%v", tc.command, got, ok, tc.want)
			}
			if ok && got != primary {
				t.Fatalf("cdPrimaryTarget(%q) target = %q, want %q", tc.command, got, primary)
			}
		})
	}
}
