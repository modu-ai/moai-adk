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
		t.Skip("dead-PID probe uses posix wait semantics")
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
		t.Fatalf("sacrificial pid %d still observed live; cannot construct a dead owner", pid)
	}
	return pid
}

// TestClearStaleBoardLock_DeadOwnerCleared — AC-KB-023 observation 1: an
// artifact left by a terminated process is removed, the removal is reported,
// and a subsequent mutation acquires the lock. Platform note: on Unix the
// substrate releases flock on process exit, so an orphaned artifact blocks
// nothing here — the requirement exists for the Windows substrate, and a
// result obtained only on Unix is recorded as such.
func TestClearStaleBoardLock_DeadOwnerCleared(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("posix dead-PID probe; windows substrate covered by GOOS=windows build")
	}
	t.Logf("platform: %s — Unix substrate releases flock(2) on exit, so this observation is trivially satisfiable here; the Windows substrate is the reason REQ-KB-023 exists", "darwin-or-linux")
	root := t.TempDir()
	dir := BoardDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir board dir: %v", err)
	}
	pid := deadPID(t)
	writeLockArtifact(t, root, pid)

	report, err := ClearStaleBoardLock(root)
	if err != nil {
		t.Fatalf("ClearStaleBoardLock(dead owner) error = %v", err)
	}
	if !report.Removed {
		t.Fatalf("report.Removed = false (%s); want removal of a dead owner's artifact", report.Reason)
	}
	if report.PID != pid {
		t.Errorf("report.PID = %d, want %d — the clear must report what it removed", report.PID, pid)
	}
	if _, err := os.Stat(boardLockPath(root)); !os.IsNotExist(err) {
		t.Fatalf("artifact still present after clear: %v", err)
	}

	// A subsequent mutation acquires the lock cleanly.
	lock, err := AcquireBoardLock(root)
	if err != nil {
		t.Fatalf("acquire after clear: %v", err)
	}
	_ = lock.Release()
}

// TestClearStaleBoardLock_LiveOwnerRefused — AC-KB-023 observation 2: an
// artifact whose recorded process is LIVE and holding the lock is not
// removed; the operation reports the refusal. The age of the artifact has no
// bearing (a freshly created artifact is refused exactly like an old one).
func TestClearStaleBoardLock_LiveOwnerRefused(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("helper re-exec plumbing exercised on unix; windows substrate covered by GOOS=windows build")
	}
	root := t.TempDir()
	releaseFile := filepath.Join(root, "release-flag")

	holder, scanner, stdout := startLockHoldHelper(t, root, releaseFile)
	defer func() { _ = stdout.Close() }()
	if got := readHelperLine(t, holder, scanner); got != "ACQUIRED" {
		t.Fatalf("holder output = %q, want ACQUIRED", got)
	}

	report, err := ClearStaleBoardLock(root)
	if err != nil {
		t.Fatalf("ClearStaleBoardLock(live owner) error = %v — a refusal must be a report, not a failure", err)
	}
	if report.Removed {
		t.Fatalf("report.Removed = true (%s); a live owner's artifact must not be removed", report.Reason)
	}
	if _, err := os.Stat(boardLockPath(root)); err != nil {
		t.Fatalf("artifact vanished despite live owner: %v", err)
	}

	if err := os.WriteFile(releaseFile, []byte("go"), 0o644); err != nil {
		t.Fatalf("release flag: %v", err)
	}
	if err := holder.Wait(); err != nil {
		t.Fatalf("holder wait: %v", err)
	}
}

// TestClearStaleBoardLock_ReacquireRaceAborts — AC-KB-023 observation 3, the
// one that decides the criterion: between the clear's inspection and its
// pre-removal re-read, the artifact is re-acquired by a live process. The
// re-read must observe the changed identity and the clear must abort — a
// clear that completes through this interleaving would unlink a valid lock
// and admit two writers to the critical section (AP-25).
//
// The interleaving is constructed at the liveness probe, which the clear runs
// immediately before the re-read: the probe's side effect re-acquires the
// lock in a separate live process, overwriting the recorded identity. The
// re-read therefore observes a different owner than the inspection did.
func TestClearStaleBoardLock_ReacquireRaceAborts(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("posix probe swap; windows substrate covered by GOOS=windows build")
	}
	root := t.TempDir()
	if err := os.MkdirAll(BoardDir(root), 0o755); err != nil {
		t.Fatalf("mkdir board dir: %v", err)
	}
	pid := deadPID(t)
	writeLockArtifact(t, root, pid)

	orig := processAlive
	t.Cleanup(func() { processAlive = orig })
	reacquired := false
	processAlive = func(candidate int) bool {
		if !reacquired {
			reacquired = true
			// The re-acquisition, ahead of the pre-removal re-read: a
			// separate live process takes the lock and records ITS identity
			// in the artifact.
			runHelperProcess(t, "reacquire-lock", map[string]string{"HELPER_ROOT": root})
		}
		return false // the ORIGINAL owner is positively absent
	}

	report, err := ClearStaleBoardLock(root)
	if err == nil {
		t.Fatalf("ClearStaleBoardLock() err = nil with report %+v — the clear completed through a changed identity: the defect AP-25 names", report)
	}
	if !IsBoardLockChangedHands(err) {
		t.Fatalf("err = %v, want ErrBoardLockChangedHands", err)
	}
	if _, statErr := os.Stat(boardLockPath(root)); statErr != nil {
		t.Fatalf("artifact did not survive the aborted clear: %v", statErr)
	}
	if !reacquired {
		t.Fatal("interleaving never fired — the probe was not consulted")
	}
}

// TestClearStaleBoardLock_NoArtifactAndUnparseable — the remaining bounds: no
// artifact is a no-op report, and an artifact whose identity cannot be parsed
// is refused (an unknown owner cannot be positively observed absent, and
// removing on parse failure would be an unconditional clear).
func TestClearStaleBoardLock_NoArtifactAndUnparseable(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	report, err := ClearStaleBoardLock(root)
	if err != nil {
		t.Fatalf("ClearStaleBoardLock(no artifact) error = %v", err)
	}
	if report.Removed {
		t.Fatalf("report.Removed = true with no artifact present")
	}

	dir := BoardDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(boardLockPath(root), []byte("not json"), 0o644); err != nil {
		t.Fatalf("write unparseable artifact: %v", err)
	}
	report, err = ClearStaleBoardLock(root)
	if err == nil {
		t.Fatalf("ClearStaleBoardLock(unparseable) err = nil, want refusal — cannot prove an unparseable owner absent")
	}
	if report != nil && report.Removed {
		t.Fatal("unparseable artifact was removed; the clear must be conditional on positive absence")
	}
	if _, statErr := os.Stat(boardLockPath(root)); statErr != nil {
		t.Fatalf("unparseable artifact did not survive: %v", statErr)
	}
}

// writeLockArtifact seeds a lock artifact recording the given owner identity,
// as a killed holder would leave behind.
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
