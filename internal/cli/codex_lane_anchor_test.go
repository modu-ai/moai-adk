package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// t1628 — the codex -l launch judges the parent-checkout rule from the
// directory the operator ran it from, not from CLAUDE_PROJECT_DIR (the project
// anchor frozen by whichever session exported it). The anchor rule is changed
// on the codex path only; the cc and glm paths stay with t1604.

// t1628LaneFixture builds a primary checkout with one card worktree and returns
// (primary, card). HOME is a temporary directory for the test.
func t1628LaneFixture(t *testing.T) (string, string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", os.Getenv("HOME"))
	primary := t.TempDir()
	t1628GitFixture(t, primary, "init", "-q")
	t1628GitFixture(t, primary, "-c", "user.email=fixture@example.com", "-c", "user.name=fixture",
		"commit", "-q", "--allow-empty", "-m", "init")
	card := filepath.Join(t.TempDir(), "card")
	t1628GitFixture(t, primary, "worktree", "add", "-q", "-b", "WT-lane-anchor-fixture", card)
	return primary, card
}

// t1628Enter changes the working directory for the test and restores it.
func t1628Enter(t *testing.T, dir string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
}

// t1628StubCodex makes the codex binary look present, without running it. The
// -l launch refuses before any launch when the factory join fails.
func t1628StubCodex(t *testing.T) {
	t.Helper()
	orig := codexLookPath
	codexLookPath = func(string) (string, error) { return "/nonexistent/codex-stub", nil }
	t.Cleanup(func() { codexLookPath = orig })
}

// t1628RunLane runs `moai codex -l` through the root command, the path the
// binary takes, and returns the error it produces.
func t1628RunLane(t *testing.T) error {
	t.Helper()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"codex", "-l"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()
	return runFang(context.Background(), rootCmd)
}

// A launch from the parent checkout must not be refused because a stale
// CLAUDE_PROJECT_DIR names a card worktree.
func TestCodexLaneAnchorIgnoresStaleProjectDirEnv(t *testing.T) {
	primary, card := t1628LaneFixture(t)
	t.Setenv("CLAUDE_PROJECT_DIR", card)
	t1628Enter(t, primary)
	t1628StubCodex(t)

	err := t1628RunLane(t)
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "parent checkout") {
		t.Errorf("a launch from the parent checkout must not be refused against CLAUDE_PROJECT_DIR (%s): %v", card, err)
	}
}

// A launch from a card worktree must be refused by the parent-checkout rule even
// when CLAUDE_PROJECT_DIR names the primary checkout.
func TestCodexLaneRefusesCardTreeEvenWhenEnvNamesPrimary(t *testing.T) {
	primary, card := t1628LaneFixture(t)
	t.Setenv("CLAUDE_PROJECT_DIR", primary)
	t1628Enter(t, card)
	t1628StubCodex(t)

	err := t1628RunLane(t)
	if err == nil {
		t.Fatal("a launch from a card worktree must be refused by the parent-checkout rule")
	}
	if !strings.Contains(err.Error(), "parent checkout") {
		t.Errorf("the refusal must come from the parent-checkout rule, got: %v", err)
	}
}
