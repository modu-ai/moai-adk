// state_lock_test.go — the cross-process exclusion of the shared file-lock
// substrate (acquireStateLockImpl), which the todo queue, the integration lock,
// and the slot lease all acquire through (SPEC-KANBAN-BOARD-001 REQ-KB-019's
// substrate property, kept after the board went: SPEC-LAUNCHER-ENTRY-FLAGS-001
// M6).
//
// The exclusion is exercised by SEPARATE OS PROCESSES — sessions are distinct
// processes, and a goroutine test would measure the harness, not the
// requirement (AP-19; internal/lockfile's in-process mutex is the repository's
// own worked example of that gap).
package factory

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// startLockHoldHelper spawns a subprocess that acquires the state lock at
// <root>/state.lock, prints ACQUIRED (or HELD), waits for the release file,
// then releases. Returns the command and its stdout scanner.
func startLockHoldHelper(t *testing.T, root, releaseFile string) (*exec.Cmd, *bufio.Scanner, io.ReadCloser) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=TestFactoryHelperProcess", "--")
	cmd.Env = append(os.Environ(),
		"MOAI_KANBAN_HELPER=lock-hold",
		"HELPER_ROOT="+root,
		"HELPER_RELEASE_FILE="+releaseFile,
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start lock-hold helper: %v", err)
	}
	scanner := bufio.NewScanner(stdout)
	return cmd, scanner, stdout
}

// readHelperLine reads one stdout line from a started helper with a deadline.
func readHelperLine(t *testing.T, cmd *exec.Cmd, scanner *bufio.Scanner) string {
	t.Helper()
	lineCh := make(chan string, 1)
	go func() {
		if scanner.Scan() {
			lineCh <- scanner.Text()
		} else {
			lineCh <- ""
		}
	}()
	select {
	case line := <-lineCh:
		return line
	case <-time.After(30 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("helper produced no output within 30s")
		return ""
	}
}

// TestStateLock_ExcludesAcrossProcesses — the substrate property, in separate
// processes: while one OS process holds the state lock, another OS process's
// acquisition attempt is refused with ErrStateLockHeld; after the holder
// releases, re-acquisition succeeds.
func TestStateLock_ExcludesAcrossProcesses(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("helper re-exec plumbing exercised on unix; windows substrate covered by GOOS=windows build")
	}
	root := t.TempDir()
	releaseFile := filepath.Join(root, "release-flag")

	holder, scanner, stdout := startLockHoldHelper(t, root, releaseFile)
	defer func() { _ = stdout.Close() }()

	first := readHelperLine(t, holder, scanner)
	if first != "ACQUIRED" {
		t.Fatalf("first holder output = %q, want ACQUIRED", first)
	}

	// A second OS process must observe contention while the first holds.
	contender, scanner2, stdout2 := startLockHoldHelper(t, root, releaseFile)
	defer func() { _ = stdout2.Close() }()
	second := readHelperLine(t, contender, scanner2)
	if second != "HELD" {
		t.Fatalf("second process output = %q, want HELD — the lock excluded nothing across processes", second)
	}
	if err := contender.Wait(); err == nil {
		t.Log("contender exited 0 on HELD; helper contract is exit-non-zero, tolerated here")
	}

	// Release the holder, then a fresh acquisition must succeed.
	if err := os.WriteFile(releaseFile, []byte("go"), 0o644); err != nil {
		t.Fatalf("write release flag: %v", err)
	}
	if err := holder.Wait(); err != nil {
		t.Fatalf("holder wait: %v", err)
	}

	lock, err := acquireStateLockImpl(filepath.Join(root, substrateLockFileName))
	if err != nil {
		t.Fatalf("re-acquisition after release failed: %v", err)
	}
	if err := lock.release(); err != nil {
		t.Fatalf("release: %v", err)
	}
}
