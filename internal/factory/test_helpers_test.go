package factory

import (
	"os/exec"
	"runtime"
	"testing"
)

// Test helpers moved verbatim out of the board test files deleted at M6
// (SPEC-LAUNCHER-ENTRY-FLAGS-001): retained tests (the status reader, the
// integration lock, the slot lease, the worktree probes) still call them.

// runtimeIsWindows reports the test's runtime OS.
func runtimeIsWindows() bool {
	return runtime.GOOS == "windows"
}

// runGitAt runs git in dir, failing the test on error.
func runGitAt(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %s: %v", args, dir, string(out), err)
	}
}

// deadPID returns the PID of a process that has positively terminated, by
// spawning and reaping a child.
func deadPID(t *testing.T) int {
	t.Helper()
	if runtimeIsWindows() {
		t.Skip("posix dead-PID probe")
	}
	cmd := exec.Command("true")
	if err := cmd.Start(); err != nil {
		t.Fatalf("spawn sacrificial process: %v", err)
	}
	pid := cmd.Process.Pid
	if err := cmd.Wait(); err != nil {
		t.Fatalf("wait sacrificial process: %v", err)
	}
	// Wait reaped the child, so the pid is positively terminated (the
	// liveness probe itself now lives behind the windows tag with the clear).
	return pid
}
