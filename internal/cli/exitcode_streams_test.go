package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestGithubDryRun verifies REQ-CONT-001-002: moai github link-spec --dry-run
// prints the planned mutation and performs NO registry write.
func TestGithubDryRun(t *testing.T) {
	// The github --dry-run flag binds to the package-level dryRun var (shared
	// with spec_status); reset it after the test to avoid cross-test bleed.
	origDryRun := dryRun
	t.Cleanup(func() { dryRun = origDryRun })
	dryRun = true

	var buf bytes.Buffer
	cmd := &cobra.Command{Use: "link-spec"}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := runLinkSpec(cmd, []string{"123", "SPEC-ISSUE-123"})
	if err != nil {
		t.Fatalf("runLinkSpec --dry-run returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Would link Issue #123") {
		t.Errorf("dry-run output missing planned mutation; got:\n%s", out)
	}
	if !strings.Contains(out, "no mutation performed") {
		t.Errorf("dry-run output missing no-mutation notice; got:\n%s", out)
	}
}

// TestSpecStatusConfirm verifies REQ-CONT-001-003: --confirm dead flag removed,
// --yes is the gate, and non-TTY without --yes aborts (no hang, no mutation).
func TestSpecStatusConfirm(t *testing.T) {
	t.Run("FlagRemoved", func(t *testing.T) {
		cmd := newSpecStatusCmd()
		if f := cmd.Flags().Lookup("confirm"); f != nil {
			t.Errorf("--confirm flag should be removed (dead flag), still registered: %v", f)
		}
		if f := cmd.Flags().Lookup("yes"); f == nil {
			t.Errorf("--yes flag should be registered as the non-interactive gate")
		}
	})

	t.Run("NonTTYAbort", func(t *testing.T) {
		// Set up a temp git repo with one non-completed SPEC referenced in git
		// log, so syncGitSpecStatuses reaches the !autoConfirm branch.
		dir := t.TempDir()
		gitInit(t, dir)
		specDir := filepath.Join(dir, ".moai", "specs", "SPEC-TEST-001")
		if err := os.MkdirAll(specDir, 0o755); err != nil {
			t.Fatalf("mkdir spec dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(specDir, "spec.md"),
			[]byte("---\nstatus: draft\n---\n"), 0o644); err != nil {
			t.Fatalf("write spec.md: %v", err)
		}
		gitCommit(t, dir, "feat(SPEC-TEST-001): test commit")

		// Point findProjectRootFn at the temp repo (overridable package var).
		orig := findProjectRootFn
		t.Cleanup(func() { findProjectRootFn = orig })
		findProjectRootFn = func() (string, error) { return dir, nil }

		// Force the non-TTY path (CI/piped-stdin behavior) deterministically.
		origTTY := stdinIsTerminalFn
		t.Cleanup(func() { stdinIsTerminalFn = origTTY })
		stdinIsTerminalFn = func() bool { return false }

		var out, errBuf bytes.Buffer
		cmd := &cobra.Command{}
		cmd.SetOut(&out)
		cmd.SetErr(&errBuf)

		// autoConfirm=false in a non-TTY test context → must abort, not hang.
		err := syncGitSpecStatuses(cmd, false)
		if err == nil {
			t.Fatalf("syncGitSpecStatuses(autoConfirm=false) in non-TTY should abort with an error, got nil")
		}
		if !strings.Contains(err.Error(), "non-TTY") && !strings.Contains(err.Error(), "--yes") {
			t.Errorf("abort error should explain the non-TTY/--yes cause; got: %v", err)
		}
		// No git mutation: the SPEC status file is still draft.
		content, _ := os.ReadFile(filepath.Join(specDir, "spec.md"))
		if !strings.Contains(string(content), "status: draft") {
			t.Errorf("SPEC status was mutated despite non-TTY abort; file now:\n%s", content)
		}
	})
}

// TestPrePushStderr verifies REQ-CONT-001-006: on exit 2 the violation details
// appear on stderr (so git + Claude Code surface the reason), not solely stdout.
func TestPrePushStderr(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)

	t.Setenv("MOAI_ENFORCE_ON_PUSH", "true")
	t.Setenv("MOAI_GIT_CONVENTION", "conventional-commits")
	t.Setenv(config.EnvClaudeProjectDir, dir)

	// Feed a violating commit message via stdin (conventional-commits requires a
	// type prefix; "bad message no type" violates it).
	origStdin := os.Stdin
	t.Cleanup(func() { os.Stdin = origStdin })
	r, w, errPipe := os.Pipe()
	if errPipe != nil {
		t.Fatalf("os.Pipe: %v", errPipe)
	}
	if _, err := w.Write([]byte("bad message no type\n")); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	w.Close()
	os.Stdin = r

	var outBuf, errBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	err := runPrePush(cmd, nil)
	// Exit 2 (enforce + violation) is carried as an ExitCoder return.
	assertExitCode(t, err, 2)

	// Violation details must be on stderr, not solely stdout.
	if !strings.Contains(errBuf.String(), "violate") {
		t.Errorf("violation details missing from stderr; stderr=\n%s", errBuf.String())
	}
}

// gitInit initializes a bare-enough git repo in dir for convention validation.
func gitInit(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test"},
	} {
		if err := exec.Command("git", append([]string{"-C", dir}, args...)...).Run(); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
}

// gitCommit stages everything in dir and commits with the given message.
func gitCommit(t *testing.T, dir, msg string) {
	t.Helper()
	if err := exec.Command("git", "-C", dir, "add", "-A").Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "commit", "-m", msg).Run(); err != nil {
		t.Fatalf("git commit: %v", err)
	}
}
