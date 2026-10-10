package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// t1628 — `moai codex -l` run from a card worktree is refused by the
// parent-checkout rule (REQ-SD-010). The refusal must name the verb the
// operator ran. It read "factory next: refused", a verb this launch never runs.
func TestCodexLaneRefusalNamesItsVerb(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", os.Getenv("HOME"))
	t.Setenv("CLAUDE_PROJECT_DIR", "")

	parent := t.TempDir()
	t1628GitFixture(t, parent, "init", "-q")
	t1628GitFixture(t, parent, "-c", "user.email=fixture@example.com", "-c", "user.name=fixture",
		"commit", "-q", "--allow-empty", "-m", "init")
	worktree := filepath.Join(t.TempDir(), "card")
	t1628GitFixture(t, parent, "worktree", "add", "-q", "-b", "WT-lane-refusal-fixture", worktree)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(worktree); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"codex", "-l"})
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()

	runErr := runFang(context.Background(), rootCmd)
	if runErr == nil {
		t.Fatal("codex -l from a card worktree must be refused by the parent-checkout rule")
	}
	msg := runErr.Error()
	if strings.Contains(strings.ToLower(msg), "factory next") {
		t.Errorf("refusal names the wrong verb (factory next) for codex -l: %v", runErr)
	}
	if !strings.Contains(msg, "codex -l") {
		t.Errorf("refusal should name the verb it was given (codex -l): %v", runErr)
	}
}

// t1628GitFixture runs git in dir for the refusal fixture.
func t1628GitFixture(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
