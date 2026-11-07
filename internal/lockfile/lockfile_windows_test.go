//go:build windows

package lockfile

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestWindowsLockAllowsDataIOWhileExcludingContenders pins the property the
// Unix flock already has and the first LockFileEx arm lacked: the lock is
// ADVISORY between lockers. It must exclude a second locker, yet leave the file's
// bytes readable and writable through other handles, because call sites lock a
// data file and read it back through a separate handle
// (internal/cli/taskledger, internal/harness retention).
//
// A byte-range lock on bytes 0.. is mandatory on Win32: another handle's
// ReadFile of that range fails with ERROR_LOCK_VIOLATION. The lock therefore
// has to sit on a region that overlaps no data.
func TestWindowsLockAllowsDataIOWhileExcludingContenders(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.txt")
	const content = "tasklist row 1\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	holder, err := os.OpenFile(path, os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = holder.Close() }()
	if err := Lock(holder); err != nil {
		t.Fatalf("Lock: %v", err)
	}

	// Data I/O through OTHER handles succeeds while the lock is held.
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile through another handle while locked: %v", err)
	}
	if string(got) != content {
		t.Fatalf("ReadFile = %q, want %q", got, content)
	}
	appender, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("open for append while locked: %v", err)
	}
	if _, err := appender.WriteString("row 2\n"); err != nil {
		_ = appender.Close()
		t.Fatalf("append through another handle while locked: %v", err)
	}
	_ = appender.Close()

	// A second locker is still refused, with the sentinel the callers test for.
	contender, err := os.OpenFile(path, os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = contender.Close() }()
	if err := TryLock(contender); !errors.Is(err, errLockHeld) {
		t.Fatalf("TryLock while held = %v, want errLockHeld", err)
	}

	// Release must undo exactly what Lock took, so the contender now succeeds.
	if err := Unlock(holder); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if err := TryLock(contender); err != nil {
		t.Fatalf("TryLock after Unlock: %v", err)
	}
	if err := Unlock(contender); err != nil {
		t.Fatalf("Unlock contender: %v", err)
	}
}
