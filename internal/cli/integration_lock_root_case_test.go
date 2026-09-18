package cli

// integration_lock_root_case_test.go — card t766.
//
// integrationLockRoot() returned CLAUDE_PROJECT_DIR verbatim, and everything
// downstream joined onto that string: the preserved copy's directory and the
// ledger row's preserved_path. source_path and worktree, by contrast, come
// from `git rev-parse --show-toplevel`, which answers the on-disk spelling. So
// one row could name two different spellings of one tree, and the real ledger
// does: three rows from 2026-09-08 carry a preserved_path under
// `/Users/goos/moai/…` while their source_path and worktree carry
// `/Users/goos/MoAI/…`.
//
// [HARD] Every comparison here is EXACT string equality. The local filesystem
// is case-insensitive, so "the path resolves" and "the file is there" are true
// of the wrong spelling too — they would pass against the defect. What is
// being repaired is the spelling itself, so the spelling is what is asserted.
//
// The probe behind this: with CLAUDE_PROJECT_DIR=<…>/T765LOWER and a tree at
// <…>/CaseRepo, the preserved copy landed in a T765LOWER directory that did
// not previously exist. On a case-sensitive filesystem that is a second
// directory tree, not a cosmetic record error.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// caseFixtureRepo builds a git repository whose directory name has upper-case
// letters, so a wrong-case spelling of it is constructible.
func caseFixtureRepo(t *testing.T, name string) string {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("fixture: resolve temp dir: %v", err)
	}
	dir := filepath.Join(base, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("fixture: mkdir %s: %v", dir, err)
	}
	for _, args := range [][]string{
		{"init", "-b", "main"},
		{"config", "user.name", "t766 fixture"},
		{"config", "user.email", "t766@example.invalid"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("fixture: git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	return dir
}

// lowerLeaf returns the same path with its final element lower-cased — a
// spelling of the same tree that git and the filesystem both accept on a
// case-insensitive volume, and that names nothing on a case-sensitive one.
func lowerLeaf(path string) string {
	return filepath.Join(filepath.Dir(path), strings.ToLower(filepath.Base(path)))
}

// TestIntegrationLockRootKeepsACanonicalEnvSpellingExactly — the no-op control.
// It runs first for a reason: without it, a repair that returned some OTHER
// correct-looking path would still satisfy the case test below.
func TestIntegrationLockRootKeepsACanonicalEnvSpellingExactly(t *testing.T) {
	repo := caseFixtureRepo(t, "CaseRepo")
	t.Setenv("CLAUDE_PROJECT_DIR", repo)

	if got := integrationLockRoot(); got != repo {
		t.Errorf("integrationLockRoot() = %q, want exactly %q", got, repo)
	}
}

// TestIntegrationLockRootCorrectsAWrongCaseEnvSpellingExactly is the card.
//
// The assertion is `got != canonical`, not os.Stat or filepath.Clean: on this
// filesystem the wrong-case spelling opens the same directory, so any check
// that asks "does it work" passes against the unrepaired code.
func TestIntegrationLockRootCorrectsAWrongCaseEnvSpellingExactly(t *testing.T) {
	canonical := caseFixtureRepo(t, "CaseRepo")
	wrong := lowerLeaf(canonical)
	if wrong == canonical {
		t.Fatalf("fixture: %q and %q are the same string; the test would be vacuous", wrong, canonical)
	}
	// Positive control on the fixture itself: the wrong spelling must reach the
	// same tree, or this measures an unreadable path rather than a spelling.
	if _, err := os.Stat(wrong); err != nil {
		t.Skipf("filesystem is case-sensitive; %q names nothing (%v)", wrong, err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", wrong)

	got := integrationLockRoot()
	if got != canonical {
		t.Errorf("integrationLockRoot() = %q, want exactly %q (git's own spelling of the same tree)", got, canonical)
	}
	if got == wrong {
		t.Errorf("integrationLockRoot() returned the caller's spelling %q verbatim", got)
	}
}

// TestIntegrationLockRootFallsBackToTheCallerValueForANonRepoPath — a
// CLAUDE_PROJECT_DIR that is not a git repository must keep working exactly as
// it does today: returned verbatim, no error, no substitution.
func TestIntegrationLockRootFallsBackToTheCallerValueForANonRepoPath(t *testing.T) {
	plain, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("fixture: resolve temp dir: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(plain, ".git")); statErr == nil {
		t.Fatalf("fixture: %q is a git repository; the fallback would not be exercised", plain)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", plain)

	if got := integrationLockRoot(); got != plain {
		t.Errorf("integrationLockRoot() = %q, want exactly %q returned verbatim", got, plain)
	}
}

// TestIntegrationLockRootFallsBackForAPathThatDoesNotExist — the shape the
// t766 probe actually used (CLAUDE_PROJECT_DIR naming a directory that did not
// exist). git cannot answer for it, so the caller's value stands, and the
// downstream preserve then creates it as it does today.
func TestIntegrationLockRootFallsBackForAPathThatDoesNotExist(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "T766LOWER")
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatalf("fixture: %q already exists (stat err=%v)", missing, err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", missing)

	if got := integrationLockRoot(); got != missing {
		t.Errorf("integrationLockRoot() = %q, want exactly %q returned verbatim", got, missing)
	}
}
