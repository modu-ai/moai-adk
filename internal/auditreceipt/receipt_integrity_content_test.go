// receipt_integrity_content_test.go — card t1632 card-review repair P1-1: the
// tree digest carries the CONTENT of the judged state — a file rewritten in
// place (same HEAD, same porcelain rows) must move the digest, or a previous
// PASS receipt stays reusable across a changed diff.
package auditreceipt

import (
	"os"
	"path/filepath"
	"testing"
)

func initContentRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if out, err := runGit(t, root, "init", "-q", "-b", "main"); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("package a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := runGit(t, root, "add", "-A"); err != nil {
		t.Fatal(err)
	}
	if _, err := runGit(t, root, "-c", "user.email=fx@example.com", "-c", "user.name=fx", "commit", "-q", "-m", "init"); err != nil {
		t.Fatal(err)
	}
	return root
}

func runGit(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	return gitIn(dir, args...)
}

// TestTreeState_ReeditMovesDigest pins the content binding: rewriting a
// tracked-modified file in place — HEAD unchanged, the same single porcelain
// row — CHANGES the digest. The old path-only digest read both states as
// identical, keeping a previous PASS receipt reusable across a changed diff.
func TestTreeState_ReeditMovesDigest(t *testing.T) {
	root := initContentRepo(t)
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("package a\n\n// modified v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, digest1, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState v1: %v", err)
	}

	// The re-edit: same file, same modified status row, different content.
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("package a\n\n// modified v2 — a different change\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, digest2, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState v2: %v", err)
	}

	if digest1 == digest2 {
		t.Fatal("digest unchanged after an in-place re-edit — the old PASS receipt would be reusable across a changed diff (card-review P1-1)")
	}
}

// TestTreeState_WhitespaceFilenameContentMovesDigest pins the l57-r2 audit
// finding: the -z split path is taken VERBATIM — a leading space (` a.go`) or
// trailing space (`b.go `) is part of the filename. Trimming it made the
// content read collide with the trimmed name (the tracked a.go), so a content
// change to the whitespace-named untracked file moved nothing and an earlier
// PASS receipt stayed fresh across a changed tree.
func TestTreeState_WhitespaceFilenameContentMovesDigest(t *testing.T) {
	root := initContentRepo(t)
	// `a.go` (tracked) and `b.go` (plain, below) exist so the trimmed names
	// resolve — the exact collision shape the finding describes.
	leading := filepath.Join(root, " a.go")
	if err := os.WriteFile(leading, []byte("leading-space v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, digest1, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState v1: %v", err)
	}
	if err := os.WriteFile(leading, []byte("leading-space v2 — changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, digest2, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState v2: %v", err)
	}
	if digest1 == digest2 {
		t.Fatal("digest unchanged after a leading-space untracked file's content changed — the trimmed path read the tracked a.go instead (l57-r2 finding 1)")
	}

	if err := os.WriteFile(filepath.Join(root, "b.go"), []byte("package b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	trailing := filepath.Join(root, "b.go ")
	if err := os.WriteFile(trailing, []byte("trailing-space v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, digest3, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState v3: %v", err)
	}
	if digest2 == digest3 {
		t.Fatal("digest unchanged after a trailing-space untracked file was added — the trailing space was trimmed away (l57-r2 finding 1)")
	}
	if err := os.WriteFile(trailing, []byte("trailing-space v2 — changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, digest4, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState v4: %v", err)
	}
	if digest3 == digest4 {
		t.Fatal("digest unchanged after a trailing-space untracked file's content changed (l57-r2 finding 1)")
	}
}

// TestTreeState_ChmodMovesUntrackedDigest pins the l57-r5 finding 2: the
// untracked digest binds the file's MODE as well as its bytes. A chmod +x on
// an untracked script changes nothing in the porcelain rows, the diff, or the
// content — the path-only digest read both states as identical, and a PASS
// receipt taken before the mode change stayed fresh across an executable
// (a materially different judged state).
func TestTreeState_ChmodMovesUntrackedDigest(t *testing.T) {
	root := initContentRepo(t)
	script := filepath.Join(root, "run.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nexit 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, digest1, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState v1: %v", err)
	}

	// The change: same path, same content, executable bit added.
	if err := os.Chmod(script, 0o755); err != nil {
		t.Fatal(err)
	}
	_, digest2, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState v2: %v", err)
	}

	if digest1 == digest2 {
		t.Fatal("digest unchanged after chmod +x on an untracked script — a pre-change PASS receipt reads fresh across an executable file (l57-r5 finding 2)")
	}
}

// TestTreeState_SymlinkReplacementMovesUntrackedDigest pins the symlink half
// of the l57-r5 finding 2: replacing an untracked file with a SAME-CONTENT
// symlink moves the digest. os.ReadFile follows the link, so the content-only
// digest read the two states as identical; the file TYPE (lstat) and the link
// target are part of the state the audit judged.
func TestTreeState_SymlinkReplacementMovesUntrackedDigest(t *testing.T) {
	root := initContentRepo(t)
	payload := "payload\n"
	script := filepath.Join(root, "run.sh")
	if err := os.WriteFile(script, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	_, digest1, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState v1: %v", err)
	}

	// The replacement: a symlink whose target — OUTSIDE the tree, so the
	// untracked path set itself is unchanged — carries the SAME bytes. The
	// only state delta is the file's own type and its link target.
	target := filepath.Join(t.TempDir(), "payload-source")
	if err := os.WriteFile(target, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(script); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, script); err != nil {
		t.Fatal(err)
	}
	_, digest2, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState v2: %v", err)
	}

	if digest1 == digest2 {
		t.Fatal("digest unchanged after a same-content symlink replacement — a pre-change PASS receipt reads fresh (l57-r5 finding 2)")
	}
}

// TestTreeState_DanglingSymlinkComputesAndBindsLinkTarget pins the l57-r6
// finding 1: an untracked symlink is bound as a LINK — its Readlink target
// beside the lstat metadata — and is NEVER opened. os.ReadFile follows the
// link, so binding its result bound the TARGET's content: a dangling symlink
// (a target that does not exist) failed the read, the read failure failed the
// whole TreeState, and the failed capture marked every receipt from that audit
// BindCaptureFailed — all of them unusable; a FIFO- or device-backed target
// could block the read instead. The r5 pins stay: a regular file keeps its
// content read, and a symlink still moves the digest when re-pointed.
func TestTreeState_DanglingSymlinkComputesAndBindsLinkTarget(t *testing.T) {
	root := initContentRepo(t)
	link := filepath.Join(root, "dangling")
	if err := os.Symlink(filepath.Join(t.TempDir(), "no-such-target"), link); err != nil {
		t.Fatal(err)
	}

	// THE PIN: TreeState computes over a dangling untracked symlink — no
	// error, hence no BindCaptureFailed path on the receipt write.
	_, digest1, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState failed on a dangling untracked symlink — ReadFile followed the link and bound the target's content (l57-r6 finding 1): %v", err)
	}
	if digest1 == "" {
		t.Fatal("TreeState returned an empty digest for a state it could compute")
	}

	// The link TARGET is part of the bound state: re-pointing the link moves
	// the digest even though the link's path, and its danglingness, are
	// unchanged in the porcelain rows and the diff.
	other := filepath.Join(t.TempDir(), "another-missing-target")
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(other, link); err != nil {
		t.Fatal(err)
	}
	_, digest2, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState v2: %v", err)
	}
	if digest1 == digest2 {
		t.Fatal("digest unchanged after the symlink was re-pointed at another target — the link target is not bound (l57-r6 finding 1)")
	}
}

// TestTreeState_UntrackedContentMovesDigest pins the untracked half: changing
// the CONTENT of an untracked file (its porcelain row unchanged) moves the
// digest.
func TestTreeState_UntrackedContentMovesDigest(t *testing.T) {
	root := initContentRepo(t)
	if err := os.WriteFile(filepath.Join(root, "untracked_secret.go"), []byte("key: v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, digest1, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState v1: %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, "untracked_secret.go"), []byte("key: v2 — rotated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, digest2, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState v2: %v", err)
	}

	if digest1 == digest2 {
		t.Fatal("digest unchanged after an untracked file's content changed (card-review P1-1)")
	}
}
