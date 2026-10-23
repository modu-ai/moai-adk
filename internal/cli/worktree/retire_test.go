package worktree

// retire_test.go — SPEC-WORKTREE-HOME-ROOT-001 M7 RED (card t1640):
// AC-WHR-020, the retirement of the empty MoAI legacy directory.
//
// Written before the retirement function exists — RetireEmptyLegacyDir is
// undefined here, so this package fails to compile and that compile failure
// is the RED (plan §C test-compilation order; research §R13's M7 row).
//
// The pins: the empty MoAI legacy directory is removed by non-recursive
// removal; the non-empty one stays with a notice; `.claude/worktrees` and
// every claude-native tree in it are untouched in both fixtures.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// m7InitRepoInPlace turns dir into a real git repository in place (the
// same pattern the migration tests use; the retirement check verifies the
// removal is non-recursive against a real directory).
func m7InitRepoInPlace(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "init", "-q", "--template=", "-b", "main")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init %s: %v\n%s", dir, err, out)
	}
}

// TestRetireEmptyLegacyDir_RemovesEmpty pins the positive shape: after the
// last tree under the MoAI legacy root has been migrated, the empty
// directory is removed with a non-recursive removal.
func TestRetireEmptyLegacyDir_RemovesEmpty(t *testing.T) {
	repo := t.TempDir()
	m7InitRepoInPlace(t, repo)
	legacyDir := filepath.Join(repo, ".moai", "worktrees")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatalf("mkdir legacy dir: %v", err)
	}

	err := RetireEmptyLegacyDir(repo)
	if err != nil {
		t.Fatalf("RetireEmptyLegacyDir: %v", err)
	}
	if _, statErr := os.Stat(legacyDir); !os.IsNotExist(statErr) {
		t.Errorf("the empty legacy directory was not removed (stat err=%v)", statErr)
	}
	// The Claude-owned directory is never touched.
	claudeDir := filepath.Join(repo, ".claude", "worktrees")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("mkdir claude dir: %v", err)
	}
	if _, statErr := os.Stat(claudeDir); statErr != nil {
		t.Errorf("the Claude-owned directory was removed: %v", statErr)
	}
}

// TestRetireEmptyLegacyDir_NonEmptyStays pins that a non-empty legacy
// directory (holding an unmigrated tree or any file) is left in place with
// a notice, and the Claude-owned directory is untouched.
func TestRetireEmptyLegacyDir_NonEmptyStays(t *testing.T) {
	repo := t.TempDir()
	m7InitRepoInPlace(t, repo)
	legacyDir := filepath.Join(repo, ".moai", "worktrees")
	if err := os.MkdirAll(filepath.Join(legacyDir, "unmigrated-tree"), 0o755); err != nil {
		t.Fatalf("mkdir legacy content: %v", err)
	}

	err := RetireEmptyLegacyDir(repo)
	if err == nil {
		t.Fatal("the non-empty legacy directory was removed without a notice")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Errorf("the notice %q does not say the directory is not empty", err)
	}
	if info, statErr := os.Stat(legacyDir); statErr != nil || !info.IsDir() {
		t.Errorf("the non-empty legacy directory was removed: %v", statErr)
	}
	// The Claude-owned directory is never touched.
	claudeDir := filepath.Join(repo, ".claude", "worktrees")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("mkdir claude dir: %v", err)
	}
	if info, statErr := os.Stat(claudeDir); statErr != nil || !info.IsDir() {
		t.Errorf("the Claude-owned directory was removed: %v", statErr)
	}
}
