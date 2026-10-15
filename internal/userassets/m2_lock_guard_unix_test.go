//go:build unix

package userassets

// m2_lock_guard_unix_test.go — the unix-execution arms of the M2 policy
// table (gate round 18/19): the flock is the guard truth on unix, so these
// fixtures need syscall.Flock and live under the unix build tag — the
// windows compile of the whole test package must stay clean (the round-13
// P1 class).

import (
	"os"
	"syscall"
	"testing"
)

// TestUnixFlockLeftoverClassification — a pid-less, CLEANLY-RELEASED unix
// guard file is the normal post-release shape: flock is the truth, so a
// free file classifies absent (no recovery target) and a flock-held file
// classifies owner-alive.
func TestUnixFlockLeftoverClassification(t *testing.T) {
	marker := filepathJoin(t, "guard.guard")
	if err := os.WriteFile(marker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if state, _ := ClassifyGuardMarker(marker); state != GuardMarkerAbsent {
		t.Fatalf("a clean flock leftover classified %s, want absent", state)
	}
	fd, err := os.Open(marker)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fd.Close() }()
	if err := syscall.Flock(int(fd.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatalf("flock the fixture: %v", err)
	}
	defer func() { _ = syscall.Flock(int(fd.Fd()), syscall.LOCK_UN) }()
	if state, _ := ClassifyGuardMarker(marker); state != GuardMarkerOwnerAlive {
		t.Fatalf("a flock-held guard classified %s, want owner-alive", state)
	}
}
