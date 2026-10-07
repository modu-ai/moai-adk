package feedback

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// TestBreakGateAbortsWhenContentSwapsBeforeRemove drives the re-read gate
// with the removal seam: between the verdict's re-read and the remove, the
// file at the path is swapped for a fresh labelled lock — the exact
// interleaving of review-gate finding #6 (a second reclaimer already
// acquired; the stale-armed reclaimer must not delete their lock). The gate
// must abort and leave the fresh lock byte-identical.
func TestBreakGateAbortsWhenContentSwapsBeforeRemove(t *testing.T) {
	dir := t.TempDir()
	store := NewQueueStore(filepath.Join(dir, "queue.json"))
	if err := os.MkdirAll(dir, queueDirPerm); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	lockPath := store.LockPath()
	previousBootFixture(t, lockPath)

	freshBytes := []byte(`{"pid":` + strconv.Itoa(os.Getpid()) + `,"boot_id":"` + bootIDIdentity() + `"}`)
	// The injection point is the VERDICT read: it returns the stale bytes
	// the break will verify dead, and — modelling the second reclaimer's
	// re-acquire landing between that verdict and the gate's re-read —
	// replaces the file with a fresh labelled lock on the way out. The gate
	// re-read must see the change and abort.
	prevReread := staleLockRereadFn
	t.Cleanup(func() { staleLockRereadFn = prevReread })
	firstRead := true
	staleLockRereadFn = func(path string) ([]byte, error) {
		raw, err := prevReread(path)
		if firstRead {
			firstRead = false
			if werr := os.WriteFile(path, freshBytes, queueFilePerm); werr != nil {
				return nil, werr
			}
		}
		return raw, err
	}

	if breakStaleLock(lockPath) {
		t.Fatal("the break fired although the file changed between the verdict read and the gate re-read")
	}
	after, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("read after aborted break: %v", err)
	}
	if string(after) != string(freshBytes) {
		t.Fatalf("the fresh lock's bytes were altered: %s", after)
	}
}
