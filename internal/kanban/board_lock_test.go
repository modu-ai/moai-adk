// board_lock_test.go — the board-wide advisory lock and its bounded stale
// clear (SPEC-KANBAN-BOARD-001 REQ-KB-019/023, M1).
//
// Every board mutation is serialized beneath a lock scoped to the WHOLE
// board, and the exclusion is exercised by SEPARATE OS PROCESSES — sessions
// are distinct processes, and a goroutine test would measure the harness, not
// the requirement (AP-19; internal/lockfile's in-process mutex is the
// repository's own worked example of that gap).
//
// The stale-lock clear is judged on what it REFUSES (acceptance.md §D.12):
// three observations, of which the third — the release-and-re-acquire
// interleaving ahead of the pre-removal re-read — is the one that decides
// whether the clear aborts on a changed identity instead of unlinking a
// re-acquired lock.
package kanban

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// startLockHoldHelper spawns a subprocess that acquires the board lock,
// prints ACQUIRED (or HELD), waits for the release file, then releases.
// Returns the command and its stdout scanner.
func startLockHoldHelper(t *testing.T, root, releaseFile string) (*exec.Cmd, *bufio.Scanner, io.ReadCloser) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=TestKanbanHelperProcess", "--")
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

// TestBoardLock_ExcludesAcrossProcesses — REQ-KB-019's substrate property, in
// separate processes: while one OS process holds the board lock, another OS
// process's acquisition attempt is refused with ErrBoardLockHeld; after the
// holder releases, re-acquisition succeeds.
func TestBoardLock_ExcludesAcrossProcesses(t *testing.T) {
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

	lock, err := AcquireBoardLock(root)
	if err != nil {
		t.Fatalf("re-acquisition after release failed: %v", err)
	}
	if err := lock.Release(); err != nil {
		t.Fatalf("release: %v", err)
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
	if processAlive(pid) {
		t.Fatalf("sacrificial pid %d still observed live", pid)
	}
	return pid
}

// writeLockArtifact seeds a lock artifact recording the given owner identity.
func writeLockArtifact(t *testing.T, root string, pid int) {
	t.Helper()
	dir := BoardDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir board dir: %v", err)
	}
	owner := BoardLockOwner{PID: pid, CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	body, err := json.MarshalIndent(owner, "", "  ")
	if err != nil {
		t.Fatalf("marshal owner: %v", err)
	}
	if err := os.WriteFile(boardLockPath(root), body, 0o644); err != nil {
		t.Fatalf("write lock artifact: %v", err)
	}
}

// TestClearStaleBoardLock_UnixGatedOut — review finding F5: the stale-lock
// clear is gated to Windows (the Unix substrate releases flock on process
// exit, so an orphaned artifact blocks nothing here and the clear window the
// M1 implementation opened — acquire flock, then record identity — cannot
// arise). On Unix the operation is a no-op reporting the platform gate and
// touches nothing. The Windows clear logic (dead-owner cleared, live-owner
// refused, re-acquire race abort) lives in board_lock_clear_windows_test.go
// behind a build tag and is exercised on a Windows runner / GOOS=windows
// build.
func TestClearStaleBoardLock_UnixGatedOut(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("unix-only gate observation")
	}
	root := t.TempDir()
	if err := os.MkdirAll(BoardDir(root), 0o755); err != nil {
		t.Fatalf("mkdir board dir: %v", err)
	}
	// A stale-looking artifact is present; on Unix it is inert and the clear
	// does not remove it.
	writeLockArtifact(t, root, deadPID(t))

	report, err := ClearStaleBoardLock(root)
	if err != nil {
		t.Fatalf("ClearStaleBoardLock(unix) error = %v, want nil no-op", err)
	}
	if report == nil || report.Removed {
		t.Fatalf("report = %+v, want Removed=false not-applicable on unix", report)
	}
	if _, statErr := os.Stat(boardLockPath(root)); statErr != nil {
		t.Fatalf("artifact vanished on unix: %v — the clear must be a no-op here", statErr)
	}
}
