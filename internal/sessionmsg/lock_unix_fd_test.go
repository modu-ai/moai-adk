//go:build !windows

package sessionmsg

import (
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

// TestAgentLockReleasesDescriptorZero is the regression guard for the fd-0
// sentinel bug: release() treated fd 0 as "never acquired" and returned nil
// WITHOUT closing, so a lock acquired while stdin was closed leaked its flock
// for the process lifetime and every later acquire on that path failed with
// EWOULDBLOCK.
//
// Pre-fix behaviour (observed on f33cd0564): acquire → fd=0; release → nil,
// no close; second acquire → "resource temporarily unavailable".
//
// MUST NOT be parallel: it mutates process-global descriptor 0.
//
// Mutant this test catches that a shallower one would not: a version that
// asserts only `release() == nil` passes on the broken code, because the
// broken release() returns nil precisely by doing nothing. The load-bearing
// assertion is that a SECOND acquire on the same path succeeds.
func TestAgentLockReleasesDescriptorZero(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "fdzero.lock")

	// Preserve stdin so the rest of the suite is unaffected, then free fd 0
	// so the kernel hands it to the next open().
	savedStdin, err := unix.Dup(0)
	if err != nil {
		t.Skipf("cannot dup stdin on this platform: %v", err)
	}
	t.Cleanup(func() {
		_ = unix.Dup2(savedStdin, 0)
		_ = unix.Close(savedStdin)
	})
	if err := unix.Close(0); err != nil {
		t.Skipf("cannot close stdin: %v", err)
	}

	first := newAgentLock()
	if err := first.acquire(lockPath); err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	if first.fd != 0 {
		// The descriptor allocator did not hand us 0 (another goroutine may
		// have taken it). Nothing to prove here; stay honest rather than
		// assert on an unreproduced condition.
		_ = first.release()
		t.Skipf("acquired fd=%d, not 0 — cannot exercise the fd-0 sentinel path", first.fd)
	}
	if err := first.release(); err != nil {
		t.Fatalf("release of fd 0: %v", err)
	}

	// The real assertion: the flock is gone, so a fresh lock can take it.
	second := newAgentLock()
	if err := second.acquire(lockPath); err != nil {
		t.Fatalf("second acquire after releasing fd 0 failed — the flock leaked: %v", err)
	}
	if err := second.release(); err != nil {
		t.Errorf("second release: %v", err)
	}
}
