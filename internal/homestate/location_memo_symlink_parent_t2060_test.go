package homestate

// Card t2060 (PR #1969 Codex review P2, location_memo.go): when the looked-up path is a symlink to a subdirectory of
// another directory, git walks up from the RESOLVED directory, but the fingerprint walked up only from the path
// as given. A repository created (or removed) above the link's target left the fingerprint unchanged.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocationMemo_RepositoryCreatedAboveASymlinkTarget(t *testing.T) {
	base := lmDir(t)
	target := filepath.Join(base, "sub")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(lmDir(t), "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	resetLocationMemo()
	// Target is outside any repository, so the answer is the (resolved) directory itself.
	before := CanonicalProjectRoot(link)
	if before == base {
		t.Fatalf("control: before git init the answer is already %q", before)
	}
	fpBefore := locationFingerprint(link)
	lmGit(t, base, "init", "-q", "-b", "main")
	if fpAfter := locationFingerprint(link); fpAfter == fpBefore {
		t.Errorf("fingerprint unchanged after `git init` above the symlink target")
	}
	if got := CanonicalProjectRoot(link); got != base {
		t.Errorf("after `git init` above the link target, CanonicalProjectRoot = %q, want %q (a stale remembered answer)", got, base)
	}
}
