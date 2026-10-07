//go:build unix

// The release-path FIFO fixture is unix-only (FIFO semantics; the B1
// split rule). The release read its lock file with a plain os.ReadFile:
// a path swapped for a FIFO after the claim parked the release past every
// deadline — and the release runs in the caller's defer, so no caller
// deadline reaches it. The release must refuse a non-regular path
// without opening it and remove nothing it cannot verify as its own.

package atomicfile

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestReleaseRefusesFifoWithoutBlocking(t *testing.T) {
	path := filepath.Join(t.TempDir(), "section.lock")
	release, err := ClaimSection(context.Background(), path, 0o600, 1, time.Millisecond)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}

	// Swap the claimed regular file for a FIFO AFTER the claim — the
	// interleave the finding reproduced.
	if rerr := os.Remove(path); rerr != nil {
		t.Fatalf("clear the claimed lock: %v", rerr)
	}
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Skipf("mkfifo unavailable on this platform: %v", err)
	}

	start := time.Now()
	rerr := release()
	elapsed := time.Since(start)
	if elapsed > 300*time.Millisecond {
		t.Fatalf("the lock release blocked %s on a FIFO lock path — a non-regular file must be refused without opening it", elapsed)
	}
	if rerr != nil {
		t.Fatalf("release: %v", rerr)
	}
}
