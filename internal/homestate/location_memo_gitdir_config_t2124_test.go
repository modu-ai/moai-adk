package homestate

// Card t2124 (PR #1969 Codex review P2, location_memo.go): when `.git` is a FILE naming an external git
// directory, the fingerprint recorded only whether that directory has HEAD and commondir. A change to the
// directory's `config` (for example `core.worktree`) moved the answer git gives while the fingerprint stayed
// the same, so a long-running process kept returning the old location.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocationMemo_CoreWorktreeChangedInAnExternalGitDir(t *testing.T) {
	base := lmDir(t)
	wt := filepath.Join(base, "wt")
	other := filepath.Join(base, "other")
	for _, d := range []string{wt, other} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	gitdir := filepath.Join(base, "meta.git")
	lmGit(t, wt, "init", "-q", "-b", "main", "--separate-git-dir", gitdir)
	lmGit(t, wt, "commit", "-q", "--allow-empty", "-m", "seed")
	if fi, err := os.Lstat(filepath.Join(wt, ".git")); err != nil || fi.IsDir() {
		t.Fatalf("control: %s/.git must be a file naming the external git directory (err %v)", wt, err)
	}

	resetLocationMemo()
	before := CanonicalProjectRoot(wt)
	fpBefore := locationFingerprint(wt)

	// Git now places the working tree somewhere else; nothing at wt or in HEAD/commondir changed.
	lmGit(t, wt, "config", "core.worktree", other)

	fpAfter := locationFingerprint(wt)
	got := CanonicalProjectRoot(wt)
	resetLocationMemo()
	want := CanonicalProjectRoot(wt)
	if got != want {
		t.Errorf("CanonicalProjectRoot = %q after core.worktree changed, a fresh lookup answers %q (a stale remembered answer; before = %q)", got, want, before)
	}
	if fpAfter == fpBefore {
		t.Errorf("fingerprint unchanged after the external git directory's config changed")
	}
}
