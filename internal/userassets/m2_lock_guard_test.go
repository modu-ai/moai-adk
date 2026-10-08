// m2_lock_guard_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 M2, the lock
// guard recovery tests (REQ-LOCK-001/002):
//   - the PLATFORM-NEUTRAL policy table (marker state × owner liveness)
//     judging guardMarkerDead/reclaimGuardMarker — the one predicate both
//     platform guards decide their reclaim through;
//   - the AC-001 case-2 contract: an ownerless marker is never
//     auto-reclaimed and resolves through the doctor's visible-recovery
//     row + explicit confirmed removal.
//
// The windows runtime arms live in m0_lock_red_test.go (skipped on darwin —
// lock_guard_windows.go is an IgnoredGoFiles here; the darwin gap is
// declared in progress.md §E.2).
package userassets

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestMarkerHelperProcessExit is the helper-process fixture: under the
// GO_USERASSETS_MARKER_HELPER env it returns immediately, giving the caller
// a reaped (PROVEN dead) pid to stamp into markers — the same evidence a
// crashed owner's pid record carries.
func TestMarkerHelperProcessExit(t *testing.T) {
	if os.Getenv("GO_USERASSETS_MARKER_HELPER") != "1" {
		t.Skip("helper process for TestGuardMarkerReclaimPolicy")
	}
}

// deadProcessPID runs the helper process to completion and returns its pid.
func deadProcessPID(t *testing.T) int {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestMarkerHelperProcessExit$", "-test.timeout=60s")
	cmd.Env = append(os.Environ(), "GO_USERASSETS_MARKER_HELPER=1")
	if err := cmd.Run(); err != nil {
		t.Fatalf("helper process: %v", err)
	}
	return cmd.Process.Pid
}

// TestGuardMarkerReclaimPolicy — the marker-state × owner-liveness matrix
// (design §3). The reclaim license exists ONLY for a pid-recorded, PROVEN
// dead owner; every other shape fails closed and the marker stays.
func TestGuardMarkerReclaimPolicy(t *testing.T) {
	dead := deadProcessPID(t)

	t.Run("pid_recorded_dead_owner_is_reclaimable", func(t *testing.T) {
		marker := filepathJoin(t, "guard.guard")
		if err := os.WriteFile(marker, []byte("pid="+itoaTest(dead)+" acquired=2026-10-08T00:00:00Z\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if !guardMarkerDead(marker) {
			t.Fatal("a pid-recorded marker with a proven-dead owner was not judged reclaimable — REQ-LOCK-001's recovery arm is broken")
		}
		if !reclaimGuardMarker(marker) {
			t.Fatal("the proven-dead marker was not reclaimed")
		}
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatalf("the marker survived its own reclaim: %v", err)
		}
	})

	t.Run("ownerless_marker_is_never_reclaimable", func(t *testing.T) {
		marker := filepathJoin(t, "guard.guard")
		if err := os.WriteFile(marker, []byte("orphaned marker — no pid record\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if guardMarkerDead(marker) {
			t.Fatal("an ownerless marker was judged reclaimable — the suspended-process shape must never grade as dead (design §3)")
		}
		if reclaimGuardMarker(marker) {
			t.Fatal("an ownerless marker was reclaimed — REQ-LOCK-001 forbids auto-reclaim without ownership evidence")
		}
		if _, err := os.Stat(marker); err != nil {
			t.Fatalf("the ownerless marker must survive for the doctor's visible-recovery row: %v", err)
		}
	})

	t.Run("live_owner_marker_is_never_reclaimable", func(t *testing.T) {
		marker := filepathJoin(t, "guard.guard")
		if err := os.WriteFile(marker, []byte("pid="+itoaTest(os.Getpid())+" acquired=2026-10-08T00:00:00Z\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if guardMarkerDead(marker) {
			t.Fatal("a live owner's marker was judged reclaimable — PID reuse can delay recovery but cannot authorize a second live writer")
		}
	})

	t.Run("malformed_record_fails_closed", func(t *testing.T) {
		for _, record := range []string{"", "pid=invalid", "pid=0", "pid=-1", "pid=4294967296"} {
			marker := filepathJoin(t, "guard.guard")
			if err := os.WriteFile(marker, []byte(record), 0o644); err != nil {
				t.Fatal(err)
			}
			if guardMarkerDead(marker) {
				t.Fatalf("malformed record %q graded as reclaimable — lockOwnerGone's fail-closed posture broke", record)
			}
		}
	})

	t.Run("unreadable_marker_fails_closed", func(t *testing.T) {
		marker := filepathJoin(t, "guard-dir") // a directory: present, unreadable as a file
		if err := os.MkdirAll(marker, 0o755); err != nil {
			t.Fatal(err)
		}
		if guardMarkerDead(marker) {
			t.Fatal("an unreadable marker graded as reclaimable — fail-closed broke")
		}
	})

	t.Run("classifies_for_the_doctor_row", func(t *testing.T) {
		home := t.TempDir()
		marker := GuardMarkerPath(home)
		if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
			t.Fatal(err)
		}
		if state, _ := ClassifyGuardMarker(marker); state != GuardMarkerAbsent {
			t.Fatalf("absent marker classified %s", state)
		}
		if err := os.WriteFile(marker, []byte("orphaned\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if state, pid := ClassifyGuardMarker(marker); state != GuardMarkerOwnerless || pid != 0 {
			t.Fatalf("ownerless marker classified %s pid=%d", state, pid)
		}
		if err := os.WriteFile(marker, []byte("pid="+itoaTest(dead)+" acquired=2026-10-08T00:00:00Z\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if state, got := ClassifyGuardMarker(marker); state != GuardMarkerOwnerDead || got != dead {
			t.Fatalf("dead-owner marker classified %s pid=%d, want owner-dead %d", state, got, dead)
		}
		if err := os.WriteFile(marker, []byte("pid="+itoaTest(os.Getpid())+" acquired=2026-10-08T00:00:00Z\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if state, _ := ClassifyGuardMarker(marker); state != GuardMarkerOwnerAlive {
			t.Fatalf("live-owner marker classified %s", state)
		}
		// Malformed pid records classify ownerless (fail closed) — the same
		// records lockOwnerGone refuses.
		for _, record := range []string{"pid=invalid", "pid=0", "pid=-1", "pid=4294967296", "acquired=2026-10-08T00:00:00Z"} {
			if err := os.WriteFile(marker, []byte(record), 0o644); err != nil {
				t.Fatal(err)
			}
			if state, pid := ClassifyGuardMarker(marker); state != GuardMarkerOwnerless || pid != 0 {
				t.Fatalf("malformed record %q classified %s pid=%d, want ownerless/0", record, state, pid)
			}
		}
	})
}

// filepathJoin builds a unique marker path under the test's temp dir.
func filepathJoin(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(t.TempDir(), name)
}

// TestGuardMarkerStateString pins the doctor-facing state renderings.
func TestGuardMarkerStateString(t *testing.T) {
	for state, want := range map[GuardMarkerState]string{
		GuardMarkerAbsent:     "absent",
		GuardMarkerOwnerDead:  "owner-dead",
		GuardMarkerOwnerAlive: "owner-alive",
		GuardMarkerOwnerless:  "ownerless",
		GuardMarkerState(99):  "unknown",
	} {
		if got := state.String(); got != want {
			t.Fatalf("GuardMarkerState(%d).String() = %q, want %q", int(state), got, want)
		}
	}
}

// TestGuardMarkerPathDerivation pins the doctor-visible marker path to the
// acquisition-produced shape (the M0.1 lock-shape test observed
// user-assets.acquire-guard.guard on disk): the path helper must name the
// same file.
func TestGuardMarkerPathDerivation(t *testing.T) {
	home := t.TempDir()
	got := GuardMarkerPath(home)
	if !strings.HasSuffix(got, "user-assets.acquire-guard.guard") {
		t.Fatalf("GuardMarkerPath = %s, want the acquisition-produced user-assets.acquire-guard.guard shape", got)
	}
}

// TestLockStaleDeadOwnerTakenOver — the .lock file's own recovery arm (the
// posture the guard markers converge toward): a stale lock recording a
// PROVEN-dead owner is taken over by the next acquirer, atomically (the
// rename-reclaim branch).
func TestLockStaleDeadOwnerTakenOver(t *testing.T) {
	home := t.TempDir()
	path := LockPath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("pid="+itoaTest(deadProcessPID(t))+" acquired=2026-10-08T00:00:00Z\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	lock, err := acquireUserLockStale(path, time.Second, time.Millisecond)
	if err != nil {
		t.Fatalf("a stale lock with a proven-dead owner was not taken over: %v", err)
	}
	defer func() { _ = lock.Release() }()
	// The takeover wrote THIS caller's ownership record.
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "pid="+itoaTest(os.Getpid())) {
		t.Fatalf("the takeover did not re-stamp the new owner: %q", raw)
	}
}
