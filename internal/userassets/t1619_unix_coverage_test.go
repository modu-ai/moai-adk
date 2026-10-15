//go:build unix

// t1619_unix_coverage_test.go — card t1619 bundle A: the unix-only arms of the
// changed-line closure in internal/userassets. The lock-guard helpers are
// defined in lock_guard_unix.go, and the mid-write case needs the POSIX
// file-size limit. Each test drives a real input and asserts the observed
// outcome. No production code is changed.
package userassets

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// TestClassifyPidlessMarkerFailsClosedWhenProbeCannotOpen covers the probe
// open failure. A marker that cannot be opened reads as held, never as free.
func TestClassifyPidlessMarkerFailsClosedWhenProbeCannotOpen(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "absent.acquire-guard.guard")
	state, pid := classifyPidlessMarker(marker)
	if state != GuardMarkerOwnerAlive || pid != 0 {
		t.Fatalf("classifyPidlessMarker of an unopenable marker = (%v, %d), want (owner-alive, 0)", state, pid)
	}
}

// TestReadLockRecordIrregularWhenOpenFailsNotAbsent covers the open failure that
// is not a missing file. A parent path that is a regular file yields ENOTDIR,
// which must read as irregular and not as absent.
func TestReadLockRecordIrregularWhenOpenFailsNotAbsent(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "plain")
	if err := os.WriteFile(blocker, []byte("not a directory\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rec := readLockRecord(filepath.Join(blocker, "lock-record"))
	if rec.status != recordIrregular {
		t.Fatalf("readLockRecord under a regular file = status %d, want recordIrregular (%d)", rec.status, recordIrregular)
	}
}

// TestConfinedWriteRefusesParentPathThatIsAFile covers the destination-parent
// pin. The mkdirs flag is a parameter no production call sets to false, so this
// test exercises the function's own contract. A regular file at the parent path
// makes the pin fail, and nothing is written.
func TestConfinedWriteRefusesParentPathThatIsAFile(t *testing.T) {
	home := t.TempDir()
	root := resolvedSkillsRoot(t, home)
	blocker := filepath.Join(root.dir, "afile")
	if err := os.WriteFile(blocker, []byte("occupied\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	inst := &Installer{Home: home, MoaiVersion: "test"}

	err := inst.confinedWrite(root, "afile/leaf.txt", []byte("payload\n"), false)
	if err == nil || !strings.Contains(err.Error(), "pin destination parent") {
		t.Fatalf("confinedWrite under a file parent = %v, want the destination-parent pin refusal", err)
	}
	if got := readBytes(t, blocker); string(got) != "occupied\n" {
		t.Fatalf("the parent file was changed by a refused write: %q", got)
	}
}

// TestConfinedWriteRemovesTempWhenWriteFails covers the mid-write failure arm.
// The file-size limit is lowered to zero for the duration of one write, so the
// temp file's write fails with EFBIG. The arm must remove the temp file and
// return the write error, and the limit is restored in cleanup.
func TestConfinedWriteRemovesTempWhenWriteFails(t *testing.T) {
	home := t.TempDir()
	root := resolvedSkillsRoot(t, home)
	inst := &Installer{Home: home, MoaiVersion: "test"}

	var saved syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_FSIZE, &saved); err != nil {
		t.Fatalf("getrlimit RLIMIT_FSIZE: %v", err)
	}
	t.Cleanup(func() { _ = syscall.Setrlimit(syscall.RLIMIT_FSIZE, &saved) })
	zero := syscall.Rlimit{Cur: 0, Max: saved.Max}
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &zero); err != nil {
		t.Fatalf("lower RLIMIT_FSIZE for one write: %v", err)
	}

	err := inst.confinedWrite(root, "sub/leaf.txt", []byte("payload\n"), true)
	if !errors.Is(err, syscall.EFBIG) {
		t.Fatalf("confinedWrite with a zero file-size limit = %v, want EFBIG", err)
	}
	entries, readErr := os.ReadDir(filepath.Join(root.dir, "sub"))
	if readErr != nil || len(entries) != 0 {
		t.Fatalf("the failed write left entries behind: %v (%v)", entries, readErr)
	}
	if _, statErr := os.Lstat(filepath.Join(root.dir, "sub", "leaf.txt")); !os.IsNotExist(statErr) {
		t.Fatalf("the destination was created by a failed write: %v", statErr)
	}
}
