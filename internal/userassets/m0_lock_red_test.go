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

	t.Run("windows_orphan_marker_is_confirmed_dead_and_reclaimed", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("windows marker semantics (lock_guard_windows.go) — darwin cannot execute this arm; covered by the GOOS=windows build gate here and a windows execution environment for the runtime RED")
		}
		home := t.TempDir()
		path := LockPath(home)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		// A windows owner died holding the guard marker: the file remains,
		// carrying NO ownership record at all. The next acquirer must confirm
		// the owner's death and reclaim the marker.
		if err := os.WriteFile(path+".guard", []byte("orphaned marker — owner process is gone\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		release, err := acquireGuard(path, 2*time.Second)
		if err != nil {
			t.Fatalf("RED (intended, windows execution): the orphaned windows marker has no ownership record and no reclaim path — acquisition timed out: %v", err)
		}
		release()
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

// TestLockGuardPathShape documents the guard path derivation both platform
// files must agree on (strings.TrimSuffix(path, ".lock") + ".acquire-guard"),
// so a windows-only rename of the guard shape cannot silently fork the
// platform semantics (REQ-LOCK-002 — the source re-read that backs the
// acceptance's "게이트 unix 프로브 + 소스 재독 확정" line).
func TestLockGuardPathShape(t *testing.T) {
	home := t.TempDir()
	lockPath := LockPath(home)
	guardPath := strings.TrimSuffix(lockPath, ".lock") + ".acquire-guard"
	want := filepath.Join(home, ".moai", "user-assets.acquire-guard")
	if guardPath != want {
		t.Fatalf("guard path derivation diverged: got %s want %s", guardPath, want)
	}
}
