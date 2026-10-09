//go:build unix

package userassets

// m7_coverage_topup_unix_test.go — the unix-only arms of the M7.1 coverage
// top-up: the injections need unix permission semantics (a chmod-0555
// parent makes the pinned O_EXCL create fail with a NON-EEXIST error) or
// cross-fd flock contention. Windows fails these injections differently
// (dir chmod is a no-op; the guard serializes via LockFileEx), so the file
// is build-tagged from first write per the standing M0 rule — the windows
// compile of the test package must stay clean.

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestUnixConfinedWriteCreateErrorSurfaces — a read-only parent makes the
// pinned handle's O_CREATE|O_EXCL fail with EACCES on every retry attempt;
// the non-EEXIST error must SURFACE (not be swallowed as a collision) and
// nothing may land. Runs only as a non-root principal — root ignores the
// mode bits the injection relies on.
func TestUnixConfinedWriteCreateErrorSurfaces(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root ignores mode bits — the chmod injection cannot fail")
	}
	home := t.TempDir()
	root := resolvedSkillsRoot(t, home)
	parent := filepath.Join(root.dir, "sub")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	inst := f_installerFor(t, home)
	if err := os.Chmod(parent, 0o555); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(parent, 0o755) }()

	err := inst.confinedWrite(root, "sub/leaf.txt", []byte("payload\n"), false)
	if err == nil {
		t.Fatal("a write into a read-only parent succeeded, want the create error surfaced")
	}
	if strings.Contains(err.Error(), "no exclusive temp name") {
		t.Fatalf("the create error was mis-routed to the exhausted-retry arm: %v", err)
	}
	entries, readErr := os.ReadDir(parent)
	if readErr != nil || len(entries) != 0 {
		t.Fatalf("the refused write left entries behind: %v (%v)", entries, readErr)
	}
}

// TestUnixGuardContentionTimesOut — a guard file whose flock is HELD (by
// this test process through a second open fd) blocks acquisition: the
// caller gets the acquire-guard error after its timeout, never the lock.
func TestUnixGuardContentionTimesOut(t *testing.T) {
	home := t.TempDir()
	guard := GuardMarkerPath(home)
	if err := os.MkdirAll(filepath.Dir(guard), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(guard, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	fd, err := os.Open(guard)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fd.Close() }()
	if err := syscall.Flock(int(fd.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatalf("flock the fixture: %v", err)
	}
	defer func() { _ = syscall.Flock(int(fd.Fd()), syscall.LOCK_UN) }()

	if _, err := AcquireUserLock(home, 80*time.Millisecond); err == nil || !strings.Contains(err.Error(), "acquire guard") {
		t.Fatalf("acquire under a held guard = %v, want acquire-guard timeout", err)
	}
}

// f_installerFor builds a minimal Installer over a temp home — the confined
// write paths under test never consult the catalog or source.
func f_installerFor(t *testing.T, home string) *Installer {
	t.Helper()
	return &Installer{Home: home, MoaiVersion: "test"}
}

// TestUnixStageJournalFailureRefusesTheRun — a read-only .moai home makes
// the STAGE journal persist fail before any per-file write; the run must
// refuse (fail-closed) instead of installing with no recovery journal.
// Runs only as non-root — root ignores the mode bits the injection relies
// on.
func TestUnixStageJournalFailureRefusesTheRun(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root ignores mode bits — the chmod injection cannot fail")
	}
	f := newFixture(t)
	inst := f.installer(t)
	moai := MoaiHome(f.home)
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(moai, 0o555); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(moai, 0o755) }()

	if _, err := inst.Install(nil); err == nil || !strings.Contains(err.Error(), "stage journal") {
		t.Fatalf("Install with an unwritable journal home = %v, want a stage-journal refusal", err)
	}
}
