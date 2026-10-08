//go:build unix

// The spool consume reread's FIFO fixture is unix-only (FIFO semantics;
// the B1 split rule). The first spool read is bounded, but the consume
// step's reread was a plain os.ReadFile: if the spool file is swapped for
// a FIFO between the two reads, the flush blocks forever WITH the spool
// section lock held. The reread must apply the same non-regular-file
// refusal, size cap, and time box the first read has.

package bugreport

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestSpoolConsumeRereadRefusesFifoWithoutBlocking(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv("CI", "")
	path, err := SpoolPath()
	if err != nil {
		t.Fatalf("spool path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create spool directory: %v", err)
	}
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Skipf("mkfifo unavailable on this platform: %v", err)
	}

	// FIFO open ordering: open the read end non-blocking first so the
	// writer's blocking open succeeds, then release our descriptor — the
	// consume reread's open would otherwise block forever.
	r, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		t.Fatalf("open fifo read end: %v", err)
	}
	opened := make(chan struct{})
	go func() {
		f, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err != nil {
			close(opened)
			return
		}
		defer func() { _ = f.Close() }()
		close(opened)
		time.Sleep(10 * time.Second)
	}()
	t.Cleanup(func() {
		select {
		case <-opened:
		case <-time.After(100 * time.Millisecond):
		}
	})
	select {
	case <-opened:
	case <-time.After(2 * time.Second):
		t.Skip("writer could not open the fifo")
	}
	_ = r.Close()

	start := time.Now()
	_ = ConsumeSpoolPrefix([]byte("anything"))
	elapsed := time.Since(start)
	if elapsed > 300*time.Millisecond {
		t.Fatalf("the spool consume reread blocked %s on a FIFO spool file — the reread must refuse a non-regular file like the first read does", elapsed)
	}
	// The consume's unreadable-file contract is nil (leave it for the next
	// drain); what the bound owes is a RETURN, not an error.
	_ = filepath.Join
}
