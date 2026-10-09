// m0_lock_red_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 M0, lock family:
// AC-001 (guard marker reclaim after owner death) as the platform-split
// table the acceptance names — a darwin run exercises the unix flock
// semantics (born-green) plus the platform-neutral lock-record contract; the
// windows-marker arm is observable only on a windows execution environment
// (B1: `GOOS=windows go build` is its M0 gate, the runtime RED is a
// windows-run observation recorded as a Gap on darwin).
//
// M0 discipline: observation only — no production change.
package userassets

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestGuardMarkerReclaimAfterOwnerDeath — AC-001 (ledger 1/10-P1,
// REQ-LOCK-001/002). The scenario family: a previous owner died holding the
// acquire-guard; the next acquirer must confirm the death and reclaim, so
// acquisition succeeds within the timeout.
func TestGuardMarkerReclaimAfterOwnerDeath(t *testing.T) {
	t.Run("unix_orphan_guard_file_is_reclaimed", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("unix flock semantics — exercised by lock_guard_unix.go on unix only")
		}
		home := t.TempDir()
		path := LockPath(home)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		// A dead owner leaves the guard FILE behind (flock died with the
		// process; the file is residue). The unix guard checks the LOCK, not
		// the file's existence — re-acquisition must succeed (born-green).
		if err := os.WriteFile(path+".guard", []byte("orphan\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		release, err := acquireGuard(path, 2*time.Second)
		if err != nil {
			t.Fatalf("born-green regression: acquiring past an orphaned unix guard file failed: %v", err)
		}
		release()
	})

	t.Run("windows_dead_owner_marker_is_confirmed_dead_and_reclaimed", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("windows marker semantics (lock_guard_windows.go) — darwin cannot execute this arm; the shared policy is judged by TestGuardMarkerReclaimPolicy here and the runtime by a windows execution environment (darwin gap declared in §E.2)")
		}
		home := t.TempDir()
		path := LockPath(home)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		// A windows owner died holding the guard marker: the file remains,
		// carrying the PID record M2 stamps at create time (the fixture
		// writes the same shape a crashed owner leaves).
		if err := os.WriteFile(path+".guard", []byte("pid="+itoaTest(deadProcessPID(t))+" acquired=2026-10-08T00:00:00Z\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		release, err := acquireGuard(path, 2*time.Second)
		if err != nil {
			t.Fatalf("the death-proven marker was not reclaimed — REQ-LOCK-001's recovery arm broke on windows: %v", err)
		}
		release()
	})

	t.Run("windows_ownerless_marker_is_never_reclaimed", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("windows marker semantics — the ownerless contract is judged platform-neutrally by TestGuardMarkerReclaimPolicy/ownerless on every platform")
		}
		home := t.TempDir()
		path := LockPath(home)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		// AC-001 case 2 (design §3 round-8 redefinition): a marker with NO
		// pid record — the suspended-process shape — is NEVER auto-
		// reclaimed. Acquisition is refused within the timeout and the
		// marker survives for the doctor's visible-recovery row (doctor
		// report + explicit confirmed removal).
		if err := os.WriteFile(path+".guard", []byte("orphaned marker — owner process is gone\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		release, err := acquireGuard(path, 400*time.Millisecond)
		if err == nil {
			release()
			t.Fatal("an ownerless marker was auto-reclaimed — REQ-LOCK-001 forbids reclaim without ownership evidence (the suspended-process shape)")
		}
		if _, statErr := os.Stat(path + ".guard"); statErr != nil {
			t.Fatalf("the ownerless marker must survive for the visible-recovery path: %v", statErr)
		}
	})

	t.Run("legacy_marker_without_pid_is_never_age_reclaimed", func(t *testing.T) {
		// AC-001 case 2 (design §3 round-8 redefinition): a lock file with NO
		// pid record (the current windows-marker shape; also what a suspended
		// survivor's marker looks like) is NEVER reclaimed by age. The
		// acquisition is refused and the marker survives for user-visible
		// reporting (doctor) + explicit-confirmation removal — that resolution
		// surface is M2; M0 pins the refusal half, which is the
		// platform-neutral lock-record contract (lock.go lockOwnerGone fails
		// closed).
		home := t.TempDir()
		path := LockPath(home)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("no pid record here\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := acquireUserLockStale(path, 300*time.Millisecond, time.Nanosecond)
		if !errors.Is(err, ErrLocked) {
			t.Fatalf("RED-guard (intended GREEN at HEAD): a pid-less stale lock was not refused with ErrLocked (got %v) — age-based auto-reclaim of an ownerless marker is forbidden", err)
		}
		if _, statErr := os.Stat(path); statErr != nil {
			t.Fatalf("the refused marker must survive for user-visible reporting: %v", statErr)
		}
	})

	t.Run("lock_record_with_live_pid_is_never_reclaimed", func(t *testing.T) {
		// The companion invariant: age alone never takes a live holder's lock
		// — lockOwnerGone must gate the takeover even when the file is older
		// than staleAfter. Uses THIS test process's own pid (alive by
		// construction).
		home := t.TempDir()
		path := LockPath(home)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		content := "pid=" + itoaTest(os.Getpid()) + " token=self acquired=2026-10-08T00:00:00Z\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := acquireUserLockStale(path, 300*time.Millisecond, time.Nanosecond)
		if !errors.Is(err, ErrLocked) {
			t.Fatalf("born-green regression: a live-pid lock was taken over after only age (err %v) — acquisition must stay refused", err)
		}
	})
}

func itoaTest(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}

// TestLockGuardPathShape documents the guard marker the REAL acquisition
// produces (M0.1, gate round 13: the former form re-derived the path with
// the same formula the code uses — a tautological pass). The observation is
// the disk diff across one acquisition: exactly the lock file and ONE guard
// marker appear, the marker sits in the lock's own name family, and a
// second acquisition while held does not take over.
func TestLockGuardPathShape(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(MoaiHome(home), 0o755); err != nil {
		t.Fatal(err)
	}
	before := map[string]bool{}
	if entries, err := os.ReadDir(MoaiHome(home)); err == nil {
		for _, e := range entries {
			before[e.Name()] = true
		}
	}

	lock, err := AcquireUserLock(home, 2*time.Second)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	defer func() { _ = lock.Release() }()

	entries, err := os.ReadDir(MoaiHome(home))
	if err != nil {
		t.Fatal(err)
	}
	var created []string
	for _, e := range entries {
		if !before[e.Name()] {
			created = append(created, e.Name())
		}
	}
	// Acquisition creates the lock file itself plus the guard marker — and
	// nothing else.
	if len(created) != 2 {
		t.Fatalf("acquisition created %d files (%v), want exactly the lock and its guard marker", len(created), created)
	}
	var lockFile, guardMarker string
	for _, name := range created {
		switch {
		case name == filepath.Base(LockPath(home)):
			lockFile = name
		case strings.HasSuffix(name, ".guard"):
			guardMarker = name
		// Gate round 44-3: the windows guard removes the .guard marker
		// BEFORE acquire returns — the surviving serialized shape is
		// .guard.serialize beside the .lock, so the marker-shape judgment
		// is split per platform.
		case runtime.GOOS == "windows" && strings.HasSuffix(name, ".guard.serialize"):
			guardMarker = name
		}
	}
	if lockFile == "" {
		t.Fatalf("acquisition created no lock file: %v", created)
	}
	if guardMarker == "" {
		t.Fatalf("acquisition created no guard marker (.guard on unix, .guard.serialize on windows): %v", created)
	}
	// The marker belongs to the lock's own name family and is a SEPARATE
	// file from the lock — the observed shape both platform guards must
	// produce (REQ-LOCK-002).
	if !strings.HasPrefix(guardMarker, "user-assets") {
		t.Fatalf("guard marker %q is outside the lock's name family", guardMarker)
	}
	if guardMarker == lockFile {
		t.Fatalf("the guard marker and the lock file are the same file: %q", guardMarker)
	}
	t.Logf("observed acquisition marker: %s", guardMarker)

	// While held, a second acquisition does not take over (the guard is a
	// real serialization surface, not a decorative file).
	if _, err := AcquireUserLock(home, 150*time.Millisecond); err == nil {
		t.Fatal("a second acquisition took the lock while it was held — the guard does not serialize")
	}
}
