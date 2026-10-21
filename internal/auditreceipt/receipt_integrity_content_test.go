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
