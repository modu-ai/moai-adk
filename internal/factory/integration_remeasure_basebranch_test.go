package factory

// integration_remeasure_basebranch_test.go — card t1790, GH #1817: an empty
// integration branch never reaches git. Before the guard the empty base made
// the fork-point lookup for `refs/heads/` die with exit status 128 AFTER the
// measured command had already run.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// remeasureBaseRepo is a real git repository on main with one commit: with the
// guard disabled the measured command RUNS and the fork-point lookup then dies,
// so the marker and the error-text assertions below can tell the guard from the
// absence of a repository (sync-audit iter1 F3).
func remeasureBaseRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitFixture(t, dir, "init", "-q", "-b", "main")
	gitFixture(t, dir, "config", "user.email", "t@example.invalid")
	gitFixture(t, dir, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "seed.txt"), []byte("seed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitFixture(t, dir, "add", "seed.txt")
	gitFixture(t, dir, "commit", "-q", "-m", "seed")
	return dir
}

func TestRunRemeasureRefusesEmptyBaseBeforeRunningTheCommand(t *testing.T) {
	tree := remeasureBaseRepo(t)
	marker := filepath.Join(t.TempDir(), "ran")
	for _, base := range []string{"", "   ", "\t"} {
		_, err := RunRemeasure(t.TempDir(), tree, base, "touch "+marker)
		if err == nil {
			t.Fatalf("base %q: an empty integration branch was accepted", base)
		}
		if !strings.Contains(err.Error(), "no integration branch") {
			t.Fatalf("base %q: the refusal must name the missing branch, got %v", base, err)
		}
		if strings.Contains(err.Error(), "merge-base") {
			t.Fatalf("base %q: git was called with the empty ref: %v", base, err)
		}
		if _, statErr := os.Stat(marker); statErr == nil {
			t.Fatalf("base %q: the measured command ran although the base was empty", base)
		}
	}
}

// A base carrying surrounding whitespace is measured as its trimmed name: the
// guard and the ref read the same value (sync-audit iter1 F6), so a padded
// branch no longer passes the guard and then dies in the fork-point lookup.
func TestRunRemeasureTrimsThePaddedBase(t *testing.T) {
	tree := remeasureBaseRepo(t)
	rec, err := RunRemeasure(t.TempDir(), tree, " main\t", "true")
	if err != nil {
		t.Fatalf("a padded base name must be measured as its trimmed name: %v", err)
	}
	if rec.Base == "" {
		t.Fatalf("the record names no base")
	}
}
