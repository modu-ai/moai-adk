package cli

// t1619_doctor_lock_markers_coverage_test.go — card t1619, bundle B1: the unexecuted
// lines of doctor_lock_markers.go. The lock-file states are decided from the file's
// contents (a record with no pid is ownerless; a record naming a live pid is alive), so
// the tests need no flock and run on every platform.

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

// TestDoctorUserLockMarkersPidlessLockFileIsOwnerless — a lock file with no pid record
// is ownerless: the row warns, states the never-auto-reclaimed contract, and gives the
// single-quoted removal procedure. The row is report-only and removes nothing.
func TestDoctorUserLockMarkersPidlessLockFileIsOwnerless(t *testing.T) {
	home := t.TempDir()
	lockPath := userassets.LockPath(home)
	t1619MustWrite(t, lockPath, "left behind without a pid record\n")

	check := checkUserLockMarkers(home, false)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("an ownerless lock file produced %s (want warn): %q", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "never auto-reclaimed") {
		t.Errorf("the row does not state the never-auto-reclaim contract: %q", check.Message)
	}
	if want := "rm '" + lockPath + "'"; !strings.Contains(check.Message, want) {
		t.Errorf("the row does not name the single-quoted removal procedure %q: %q", want, check.Message)
	}
	if check.Detail != "" {
		t.Errorf("a non-verbose check must carry no detail, got %q", check.Detail)
	}
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("the doctor row removed the lock file (report-only contract broken): %v", err)
	}
}

// TestDoctorUserLockMarkersVerboseOwnerlessAddsDetail — the verbose report explains why
// the ownerless marker is refused rather than reclaimed.
func TestDoctorUserLockMarkersVerboseOwnerlessAddsDetail(t *testing.T) {
	home := t.TempDir()
	t1619MustWrite(t, userassets.LockPath(home), "left behind without a pid record\n")

	check := checkUserLockMarkers(home, true)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("an ownerless lock file produced %s (want warn): %q", check.Status, check.Message)
	}
	if !strings.Contains(check.Detail, "predates PID ownership records") {
		t.Errorf("the verbose detail does not explain the refusal: %q", check.Detail)
	}
}

// TestDoctorUserLockMarkersLiveOwnerReportsHeldLock — a lock file whose pid is a running
// process is a held lock: the row warns and names the holder and the way out.
func TestDoctorUserLockMarkersLiveOwnerReportsHeldLock(t *testing.T) {
	home := t.TempDir()
	t1619MustWrite(t, userassets.LockPath(home),
		fmt.Sprintf("pid=%d acquired=2026-10-13T00:00:00Z\n", os.Getpid()))

	check := checkUserLockMarkers(home, false)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("a lock held by a running process produced %s (want warn): %q", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "held by a running process") {
		t.Errorf("the row does not report the held lock: %q", check.Message)
	}
	if !strings.Contains(check.Message, fmt.Sprintf("(pid %d)", os.Getpid())) {
		t.Errorf("the row does not name the holder's pid: %q", check.Message)
	}
}

// TestDoctorUserLockMarkersDeadGuardMarkerIsReclaimedByNextRun — a guard marker whose
// owner is dead is informational: the next run reclaims it, so the row reports ok.
func TestDoctorUserLockMarkersDeadGuardMarkerIsReclaimedByNextRun(t *testing.T) {
	home := t.TempDir()
	t1619MustWrite(t, userassets.GuardMarkerPath(home), "pid=2147483647 acquired=2026-10-08T00:00:00Z\n")

	check := checkUserLockMarkers(home, false)
	if check.Status != uikit.CheckOK {
		t.Fatalf("a guard marker with a dead owner produced %s (want ok — the next run reclaims it): %q",
			check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "a stale guard marker") {
		t.Errorf("the row does not name the stale guard marker: %q", check.Message)
	}
	if !strings.Contains(check.Message, "the next moai run reclaims it automatically") {
		t.Errorf("the row does not state the reclaim condition for the guard marker: %q", check.Message)
	}
}
