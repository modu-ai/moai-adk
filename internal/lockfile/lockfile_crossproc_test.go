package lockfile

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// The tests in this file carry NO build constraint, like lockfile_contract_test.go:
// they run on every GOOS, so the Windows arm is exercised by the Windows CI leg.
//
// What they pin that the same-process contract tests cannot: exclusion between
// two PROCESSES. The Windows arm was a process-local mutex until card t2001,
// which passed every same-process test while a second process walked straight
// through the lock; a registry test that started a real second process was the
// only thing that noticed.

// xprocEnv names the file the re-executed helper process tries to lock.
const xprocEnv = "LOCKFILE_XPROC_TARGET"

// Exit codes of the helper process, distinct so a wrong outcome is readable.
const (
	xprocAcquired = 0
	xprocOpenFail = 3
	xprocRefused  = 7
)

// TestLockfileXprocHelper is the re-execution target of the child process: it
// opens the target file, makes ONE non-blocking TryLock attempt, and exits with
// a code that says what happened. Ordinary runs skip it (no env guard).
func TestLockfileXprocHelper(t *testing.T) {
	target := os.Getenv(xprocEnv)
	if target == "" {
		t.Skip("helper process only")
	}
	f, err := os.OpenFile(target, os.O_RDWR, 0o600)
	if err != nil {
		os.Exit(xprocOpenFail)
	}
	if err := TryLock(f); err != nil {
		os.Exit(xprocRefused)
	}
	_ = Unlock(f)
	_ = f.Close()
	os.Exit(xprocAcquired)
}

// runXprocChild starts the helper as a separate process against target and
// returns its exit code.
func runXprocChild(t *testing.T, target string) int {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestLockfileXprocHelper$")
	cmd.Env = append(os.Environ(), xprocEnv+"="+target)
	err := cmd.Run()
	if err == nil {
		return xprocAcquired
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	t.Fatalf("run the helper process: %v", err)
	return -1
}

// TestTryLockRefusesASecondProcessWhileHeld is the negative regression: while
// this process holds the lock, a second process's TryLock must be refused, and
// once this process releases it the second process must acquire.
//
// Non-vacuous in both directions. A TryLock that never refuses (a no-op, or the
// process-local mutex of the old Windows arm) makes the first child exit 0; an
// Unlock that does not release makes the second child exit 7. A helper that
// never ran (selector matched no test) exits 0 too, so it fails the first check.
func TestTryLockRefusesASecondProcessWhileHeld(t *testing.T) {
	target := filepath.Join(t.TempDir(), "xproc.lock")
	f, err := os.Create(target)
	if err != nil {
		t.Fatalf("create lock file: %v", err)
	}
	defer func() { _ = f.Close() }()

	if err := Lock(f); err != nil {
		t.Fatalf("Lock: %v", err)
	}
	if code := runXprocChild(t, target); code != xprocRefused {
		_ = Unlock(f)
		t.Fatalf("a second process's TryLock while the lock is held exited %d, want %d (refused)", code, xprocRefused)
	}

	if err := Unlock(f); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if code := runXprocChild(t, target); code != xprocAcquired {
		t.Fatalf("a second process's TryLock after release exited %d, want %d (acquired)", code, xprocAcquired)
	}
}
