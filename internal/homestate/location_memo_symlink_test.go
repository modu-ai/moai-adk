package homestate

// location_memo_symlink_test.go — a symlink re-pointed to another repository under a remembered answer (card t2027).
//
// The fingerprint of a directory reads the `.git` entry and the `config` stamp of whatever the path resolves to,
// but not which directory the path resolves to. Two repositories whose config files carry the same size and
// modification time (an archive restore, a filesystem with a coarse timestamp) therefore looked identical to a
// symlink that had been moved from one to the other.

import (
	"os"
	"path/filepath"
	"testing"
)

func lmSymlinkFixture(t *testing.T) (link, repoA, repoB string) {
	t.Helper()
	repoA = lmRepo(t)
	repoB = lmRepo(t)
	link = filepath.Join(lmDir(t), "link")
	if err := os.Symlink(repoA, link); err != nil {
		t.Fatal(err)
	}
	return link, repoA, repoB
}

func lmRepoint(t *testing.T, link, to string) {
	t.Helper()
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(to, link); err != nil {
		t.Fatal(err)
	}
}

// Two repositories created back to back: their config stamps differ, so the move was always noticed.
func TestLocationMemo_SymlinkRepointedToAnotherRepository(t *testing.T) {
	link, _, repoB := lmSymlinkFixture(t)
	resetLocationMemo()
	if got := CanonicalProjectRoot(link); got == repoB {
		t.Fatalf("control: the link answers %q before it is moved", got)
	}
	lmRepoint(t, link, repoB)
	if got := CanonicalProjectRoot(link); got != repoB {
		t.Errorf("after the symlink was re-pointed, CanonicalProjectRoot = %q, want %q (a stale remembered answer)", got, repoB)
	}
}

// The same move between two repositories whose config files have the same size and modification time.
func TestLocationMemo_SymlinkRepointedBetweenRepositoriesWithEqualConfigStamps(t *testing.T) {
	link, repoA, repoB := lmSymlinkFixture(t)
	ia, err := os.Stat(filepath.Join(repoA, ".git", "config"))
	if err != nil {
		t.Fatal(err)
	}
	cfgB := filepath.Join(repoB, ".git", "config")
	if err := os.Chtimes(cfgB, ia.ModTime(), ia.ModTime()); err != nil {
		t.Fatal(err)
	}
	ib, err := os.Stat(cfgB)
	if err != nil {
		t.Fatal(err)
	}
	if ia.Size() != ib.Size() || !ia.ModTime().Equal(ib.ModTime()) {
		t.Fatalf("control: the config stamps differ (%d %v vs %d %v)", ia.Size(), ia.ModTime(), ib.Size(), ib.ModTime())
	}
	resetLocationMemo()
	if got := CanonicalProjectRoot(link); got != repoA {
		t.Fatalf("control: the link answers %q before it is moved, want %q", got, repoA)
	}
	lmRepoint(t, link, repoB)
	if got := CanonicalProjectRoot(link); got != repoB {
		t.Errorf("after the symlink was re-pointed, CanonicalProjectRoot = %q, want %q (a stale remembered answer)", got, repoB)
	}
}
