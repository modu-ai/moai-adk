package migration

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"testing"
)

// TestVersionFile_RoundTrip verifies version-file read/write round-trip.
// REQ-V3R2-RT-007-013: writeVersion updates the version-file atomically.
func TestVersionFile_RoundTrip(t *testing.T) {
	root := t.TempDir()

	// Absent file → 0.
	v, err := readVersion(root)
	if err != nil || v != 0 {
		t.Fatalf("initial readVersion: got %d, %v; want 0, nil", v, err)
	}

	if err := writeVersion(root, 42); err != nil {
		t.Fatalf("writeVersion: %v", err)
	}
	got, err := readVersion(root)
	if err != nil {
		t.Fatalf("readVersion after write: %v", err)
	}
	if got != 42 {
		t.Errorf("round-trip: got %d, want 42", got)
	}

	// Overwrite.
	if err := writeVersion(root, 99); err != nil {
		t.Fatalf("writeVersion overwrite: %v", err)
	}
	got, _ = readVersion(root)
	if got != 99 {
		t.Errorf("overwrite: got %d, want 99", got)
	}
}

// TestVersionFile_AtomicRename verifies crash safety via atomic rename.
// REQ-V3R2-RT-007-013: version-file updates use the *.tmp + os.Rename pattern.
func TestVersionFile_AtomicRename(t *testing.T) {
	root := t.TempDir()
	if err := writeVersion(root, 7); err != nil {
		t.Fatalf("writeVersion: %v", err)
	}
	stateDir := filepath.Join(root, ".moai", "state")
	// After atomic write, the .tmp file must NOT remain (rename succeeded).
	if _, err := os.Stat(filepath.Join(stateDir, versionTmpFileName)); !os.IsNotExist(err) {
		t.Errorf("tmp file should not remain after atomic rename; stat err=%v", err)
	}
	// The version file must exist and contain the version.
	data, err := os.ReadFile(filepath.Join(stateDir, versionFileName))
	if err != nil {
		t.Fatalf("read version file: %v", err)
	}
	if string(data) != strconv.Itoa(7) {
		t.Errorf("version file content: got %q, want %q", string(data), "7")
	}
	// The lock file must have been created during the write and left behind.
	// This is Unix-only: the Unix lock is a flock(2) on a persistent file,
	// whereas the Windows lock is an O_CREATE|O_EXCL file mutex whose
	// releaseLock deletes the file by design (version_windows.go). Asserting
	// residue on Windows would assert against the documented contract.
	if runtime.GOOS != "windows" {
		if _, err := os.Stat(filepath.Join(stateDir, versionFileName+".lock")); err != nil {
			t.Errorf("lock file should exist after write; stat err=%v", err)
		}
	}
}

// TestVersionFile_AbsentMeansZero verifies the default of 0 when the version-file is absent.
// REQ-V3R2-RT-007-030: an absent version-file means the current version is 0.
func TestVersionFile_AbsentMeansZero(t *testing.T) {
	root := t.TempDir()
	v, err := readVersion(root)
	if err != nil {
		t.Fatalf("readVersion on absent file: unexpected err %v", err)
	}
	if v != 0 {
		t.Errorf("absent version file: got %d, want 0", v)
	}
}

// TestVersionFile_AdvisoryLock_HighContention verifies advisory locking under contention.
// REQ-V3R2-RT-007-031: version-file updates are protected by an advisory lock.
func TestVersionFile_AdvisoryLock_HighContention(t *testing.T) {
	root := t.TempDir()
	const goroutines = 20
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if err := writeVersion(root, n); err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("concurrent writeVersion error: %v", err)
	}
	// Final version must be readable and in range.
	v, err := readVersion(root)
	if err != nil {
		t.Fatalf("final readVersion: %v", err)
	}
	if v < 0 || v >= goroutines {
		t.Errorf("final version %d out of range [0,%d)", v, goroutines)
	}
}

// TestVersionFile_detectInFlightState exercises the crash-detection helper directly.
func TestVersionFile_detectInFlightState(t *testing.T) {
	root := t.TempDir()
	if detectInFlightState(root) {
		t.Error("detectInFlightState should be false with no .tmp file")
	}
	markerTmpFile(t, root)
	if !detectInFlightState(root) {
		t.Error("detectInFlightState should be true with .tmp file present")
	}
	if err := cleanupInFlightState(root); err != nil {
		t.Fatalf("cleanupInFlightState: %v", err)
	}
	if detectInFlightState(root) {
		t.Error("detectInFlightState should be false after cleanup")
	}
	// cleanup on already-absent .tmp is idempotent (no error).
	if err := cleanupInFlightState(root); err != nil {
		t.Errorf("cleanupInFlightState idempotent: %v", err)
	}
}

// TestVersionFile_ReadInvalidContent verifies readVersion errors on non-integer content.
func TestVersionFile_ReadInvalidContent(t *testing.T) {
	root := t.TempDir()
	stateDir := filepath.Join(root, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, versionFileName), []byte("not-a-number"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readVersion(root); err == nil {
		t.Error("readVersion with non-integer content should return an error")
	}
}
