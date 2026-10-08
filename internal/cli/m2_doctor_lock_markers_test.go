// m2_doctor_lock_markers_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 M2, the
// doctor's visible-recovery row (REQ-LOCK-001): an ownerless lock marker is
// NEVER auto-reclaimed and resolves through this report + explicit
// confirmed removal. The row is report-only — it must never remove
// anything.
package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

func TestDoctorUserLockMarkersReportsOwnerless(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("the ownerless state is windows-decided — on unix the flock is the truth and a pid-less free file classifies absent (gate round 19)")
	}
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(userassets.GuardMarkerPath(home)), 0o755); err != nil {
		t.Fatal(err)
	}
	marker := userassets.GuardMarkerPath(home)
	if err := os.WriteFile(marker, []byte("orphaned marker — no pid record\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	check := checkUserLockMarkers(home, false)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("an ownerless marker produced %s (want warn) — the visible-recovery row must surface it", check.Status)
	}
	if !strings.Contains(check.Message, "never auto-reclaimed") {
		t.Errorf("the row does not state the never-auto-reclaim contract: %q", check.Message)
	}
	if !strings.Contains(check.Message, "rm \"") {
		t.Errorf("the row does not name the explicit removal procedure: %q", check.Message)
	}
	if !strings.Contains(check.Message, marker) {
		t.Errorf("the row does not name the marker path: %q", check.Message)
	}
	// Report-only: the marker survives the check.
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("the doctor row removed the marker (report-only contract broken): %v", err)
	}
}

func TestDoctorUserLockMarkersAbsentIsOK(t *testing.T) {
	check := checkUserLockMarkers(t.TempDir(), false)
	if check.Status != uikit.CheckOK {
		t.Fatalf("a clean home produced %s (want ok): %q", check.Status, check.Message)
	}
}

// TestDoctorUserLockMarkersUnixFlockLeftoverIsOK — gate round 19: on unix a
// pid-less, CLEANLY-RELEASED guard file is the normal post-release shape;
// the row must treat it as no-recovery-needed, not warn with rm guidance.
func TestDoctorUserLockMarkersUnixFlockLeftoverIsOK(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix flock semantics")
	}
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(userassets.GuardMarkerPath(home)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(userassets.GuardMarkerPath(home), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	check := checkUserLockMarkers(home, false)
	if check.Status != uikit.CheckOK {
		t.Fatalf("a clean unix flock leftover produced %s (want ok): %q", check.Status, check.Message)
	}
}

// TestDoctorUserLockMarkersIrregularIsSurfaced — gate round 19: a FIFO at
// a marker path is warned, never read (this test returning is the proof
// the row did not hang).
func TestDoctorUserLockMarkersIrregularIsSurfaced(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("FIFO semantics are unix")
	}
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(userassets.GuardMarkerPath(home)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(userassets.GuardMarkerPath(home), 0o644); err != nil {
		t.Fatal(err)
	}
	check := checkUserLockMarkers(home, false)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("an irregular marker produced %s (want warn): %q", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "irregular object") {
		t.Errorf("the row does not name the irregular object: %q", check.Message)
	}
}

func TestDoctorUserLockMarkersDeadOwnerIsInformational(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(userassets.LockPath(home)), 0o755); err != nil {
		t.Fatal(err)
	}
	// The LOCK file carries the same pid-record posture; its dead-owner
	// takeover additionally waits out the stale window — the guidance names
	// the condition the code actually applies (gate round 19).
	if err := os.WriteFile(userassets.LockPath(home), []byte("pid=2147483647 acquired=2026-10-08T00:00:00Z\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	check := checkUserLockMarkers(home, false)
	if check.Status != uikit.CheckOK {
		t.Fatalf("a dead-owner lock produced %s (want ok — a later run takes it over): %q", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "stale window") {
		t.Errorf("the .lock guidance does not state the stale-window condition: %q", check.Message)
	}
}
