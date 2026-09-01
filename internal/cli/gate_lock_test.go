package cli

// gate_lock_test.go — SPEC-GATE-THREE-AXES-001 M3, unit level (card t235).
//
// The gate-run lock's substrate behaviour: contention sentinel, holder
// identity in the artifact, the bounded wait's notice and degradation, the
// dead-holder fast path, and the read-only-directory degradation. The
// end-to-end shape — two full `moai gate` runs serialized against one project
// — is verified in gate_lock_cli_test.go; what is verified HERE is the lock
// machinery itself, exercised through its own API.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// --- substrate ---------------------------------------------------------------

// TestGateLock_AcquireRecordsHolderIdentityAndReleaseAllowsReacquire — the
// artifact names its current owner (the identity a waiting run's notice
// prints and a stale clear decides about), and a released lock is
// re-acquirable.
func TestGateLock_AcquireRecordsHolderIdentityAndReleaseAllowsReacquire(t *testing.T) {
	dir := t.TempDir()

	lock, err := AcquireGateLock(dir)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}

	holder, err := ReadGateLockHolder(dir)
	if err != nil {
		t.Fatalf("reading the holder the artifact records: %v", err)
	}
	if holder.PID != os.Getpid() {
		t.Errorf("artifact records pid %d, want this process's pid %d", holder.PID, os.Getpid())
	}
	if holder.CreatedAt == "" {
		t.Errorf("artifact records no creation time; the identity block is incomplete: %+v", holder)
	}

	if err := lock.Release(); err != nil {
		t.Fatalf("release failed: %v", err)
	}
	// Release is idempotent; a second call must not error or re-lock.
	if err := lock.Release(); err != nil {
		t.Fatalf("second release errored (release must be idempotent): %v", err)
	}

	second, err := AcquireGateLock(dir)
	if err != nil {
		t.Fatalf("re-acquire after release failed: %v", err)
	}
	if err := second.Release(); err != nil {
		t.Fatalf("releasing the re-acquired lock failed: %v", err)
	}
}

// TestGateLock_ContentionReturnsSentinel — a second acquisition while one is
// held fails fast with the contention sentinel, never blocks. The two
// acquisitions are two open file descriptions, which is the same
// cross-process surface two OS processes present to the kernel.
func TestGateLock_ContentionReturnsSentinel(t *testing.T) {
	dir := t.TempDir()

	first, err := AcquireGateLock(dir)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	defer func() {
		if err := first.Release(); err != nil {
			t.Errorf("release: %v", err)
		}
	}()

	second, err := AcquireGateLock(dir)
	if !IsGateLockHeld(err) {
		t.Fatalf("second acquire while held returned %v, want the contention sentinel %v", err, ErrGateLockHeld)
	}
	if second != nil {
		if err := second.Release(); err != nil {
			t.Errorf("release of an unexpected second lock: %v", err)
		}
	}
}

// --- the bounded wait ---------------------------------------------------------

// TestGateLock_WaitNoticeNamesHolder — AC-GTA-012 (unit half).
//
// While the lock is held by a process whose PID is recorded in the artifact,
// a waiting run's notice names THAT pid. A generic "another run is in
// progress" line is the criterion's named mutant; the pid is knowable only
// by reading the artifact.
func TestGateLock_WaitNoticeNamesHolder(t *testing.T) {
	dir := t.TempDir()

	holder, err := AcquireGateLock(dir)
	if err != nil {
		t.Fatalf("holder acquire failed: %v", err)
	}
	defer func() {
		if err := holder.Release(); err != nil {
			t.Errorf("release: %v", err)
		}
	}()

	var notice bytes.Buffer
	res := waitForGateLock(dir, 300*time.Millisecond, &notice)
	if res.lock != nil {
		t.Fatalf("wait returned a lock while one was held")
	}
	out := notice.String()
	want := fmt.Sprintf("held by pid %d", os.Getpid())
	if !bytes.Contains(notice.Bytes(), []byte(want)) {
		t.Fatalf("waiting notice does not name the recorded holder pid; want %q in:\n%s", want, out)
	}
	if res.holderPID != os.Getpid() {
		t.Errorf("result names holder pid %d, want %d", res.holderPID, os.Getpid())
	}
}

// TestGateLock_WaitIsBoundedAndDegrades — AC-GTA-013 (unit half).
//
// Against a live holder, the wait lasts at least the budget (it really
// waited) and at most the budget plus a stated slack (it is bounded — the
// criterion's Mutant A is waiting indefinitely). The result carries no lock
// and reports the degradation.
func TestGateLock_WaitIsBoundedAndDegrades(t *testing.T) {
	dir := t.TempDir()

	holder, err := AcquireGateLock(dir)
	if err != nil {
		t.Fatalf("holder acquire failed: %v", err)
	}
	defer func() {
		if err := holder.Release(); err != nil {
			t.Errorf("release: %v", err)
		}
	}()

	const budget = 400 * time.Millisecond
	// The retry delay is 100ms, so anything past the budget by more than a
	// couple of retry delays — plus headroom for a loaded machine — is a
	// stall, not a bound.
	const slack = 5 * time.Second

	var notice bytes.Buffer
	res := waitForGateLock(dir, budget, &notice)
	if res.lock != nil {
		t.Fatalf("wait returned a lock while one was held")
	}
	if res.waited < budget {
		t.Errorf("wait lasted %s, want at least the budget %s — the run did not actually wait", res.waited, budget)
	}
	if res.waited > budget+slack {
		t.Errorf("wait lasted %s, want at most budget+slack (%s) — the wait is not bounded", res.waited, budget+slack)
	}
	if res.line == "" || !bytes.Contains([]byte(res.line), []byte("unserialized")) {
		t.Errorf("degradation verdict line %q does not state that the run is unserialized", res.line)
	}
}

// TestGateLock_DeadHolderArtifactAcquiresFarBelowBudget — AC-GTA-014.
//
// An artifact whose recorded holder is not alive must not cost the wait
// budget. On the Unix substrate the kernel has already released the flock,
// so acquisition succeeds outright; on Windows the clear path removes the
// artifact. Either way the acquisition lands far below the budget and no
// degradation is reported.
func TestGateLock_DeadHolderArtifactAcquiresFarBelowBudget(t *testing.T) {
	if os.PathSeparator != '/' {
		t.Skip("the /bin/sleep probe below is not a Windows command; the Windows clear path is exercised by the CI Windows matrix")
	}

	dir := t.TempDir()
	pid := deadProcessPID(t)

	// A killed holder's artifact: identity published, no lock held.
	stateDir := filepath.Join(dir, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("create state dir: %v", err)
	}
	artifact, err := json.MarshalIndent(GateLockOwner{PID: pid, CreatedAt: time.Now().UTC().Format(time.RFC3339)}, "", "  ")
	if err != nil {
		t.Fatalf("marshal owner record: %v", err)
	}
	if err := os.WriteFile(gateLockPath(dir), append(artifact, '\n'), 0o644); err != nil {
		t.Fatalf("write stale artifact: %v", err)
	}

	const budget = 30 * time.Second
	var notice bytes.Buffer
	started := time.Now()
	res := waitForGateLock(dir, budget, &notice)
	if res.lock == nil {
		t.Fatalf("a stale artifact blocked acquisition for the full run: %s", res.line)
	}
	defer func() {
		if err := res.lock.Release(); err != nil {
			t.Errorf("release: %v", err)
		}
	}()
	if elapsed := time.Since(started); elapsed > 5*time.Second {
		t.Errorf("acquiring over a dead holder's artifact took %s, want far below the %s budget", elapsed, budget)
	}
	if res.line != "" {
		t.Errorf("a dead-holder acquisition reported a degradation verdict: %q", res.line)
	}
	if notice.Len() != 0 {
		t.Errorf("a dead-holder acquisition emitted waiting notices:\n%s", notice.String())
	}
}

// TestGateLock_ReadOnlyStateDirIsUnavailableNotFatal — AC-GTA-015 (unit
// half). When the lock machinery cannot even create its directory, the wait
// reports the lock as unavailable and the caller still receives a usable
// result — the gate's own verdict decides the run, never the lock.
func TestGateLock_ReadOnlyStateDirIsUnavailableNotFatal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("posix permission bits are advisory on windows — os.Chmod cannot make the state directory read-only, so the fixture premise is unbuildable there")
	}
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatalf("create .moai: %v", err)
	}
	// Reading .moai stays possible (the config loader still works); writing
	// the state directory is what must fail.
	if err := os.Chmod(filepath.Join(dir, ".moai"), 0o500); err != nil {
		t.Skipf("cannot make .moai read-only on this filesystem: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(filepath.Join(dir, ".moai"), 0o755); err != nil {
			t.Logf("restore .moai permissions: %v", err)
		}
	})

	var notice bytes.Buffer
	res := waitForGateLock(dir, time.Second, &notice)
	if res.lock != nil {
		t.Fatalf("a lock was acquired against a read-only state directory")
	}
	if res.line == "" || !bytes.Contains([]byte(res.line), []byte("unavailable")) {
		t.Errorf("verdict line %q does not report the lock as unavailable", res.line)
	}
	if !bytes.Contains([]byte(res.line), []byte("unserialized")) {
		t.Errorf("verdict line %q does not state that the run proceeds unserialized", res.line)
	}
}

// deadProcessPID returns the pid of a process that has already exited, and
// verifies through the same probe the lock uses that it reads dead.
func deadProcessPID(t *testing.T) int {
	t.Helper()
	cmd := exec.Command("/bin/sleep", "0")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start the short-lived holder process: %v", err)
	}
	pid := cmd.Process.Pid
	if err := cmd.Wait(); err != nil {
		t.Fatalf("wait for the short-lived holder process: %v", err)
	}
	if kanban.FactoryProcessAlive(pid) {
		t.Fatalf("pid %d still reads live after exiting; the dead-holder fixture is not a dead holder", pid)
	}
	return pid
}
