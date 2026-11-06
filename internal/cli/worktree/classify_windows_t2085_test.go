//go:build windows

package worktree

// classify_windows_t2085_test.go — card t2085.
//
// git prints worktree paths with forward slashes on Windows ("C:/Users/x"),
// while the engine builds every other path with the OS separator. A record's
// Path kept git's spelling, so a plan row's Source ("C:/...") never matched the
// tree path the caller held ("C:\..."), and the first state-store lookup after a
// move used the other spelling (a canonical fallback absorbed that one).
// ParseWorktreeList converts once.

import "testing"

func TestParseWorktreeList_WindowsPathUsesTheOSSeparator(t *testing.T) {
	recs, err := ParseWorktreeList("worktree C:/Users/x/repo\nHEAD abc\nbranch refs/heads/main\n\nworktree C:/Users/x/repo/.moai/worktrees/legacy6\nHEAD def\ndetached\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want := []string{`C:\Users\x\repo`, `C:\Users\x\repo\.moai\worktrees\legacy6`}
	if len(recs) != len(want) {
		t.Fatalf("records = %d, want %d", len(recs), len(want))
	}
	for i, w := range want {
		if recs[i].Path != w {
			t.Errorf("record %d Path = %q, want %q", i, recs[i].Path, w)
		}
	}
}
