package feedback

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

// previousBootFixture writes a lock file whose recorded owner ran on a
// PREVIOUS boot (a boot id different from this machine's wire form) — a
// verified-dead owner by the boot-comparison path, whatever lives at the
// pid today. The current pid keeps the record self-consistent otherwise.
func previousBootFixture(t *testing.T, path string) {
	t.Helper()
	identity := bootIDIdentity()
	if identity == "" {
		t.Skip("no boot identity on this platform; the boot-comparison path is not exercised")
	}
	owner := lockOwner{PID: os.Getpid(), BootID: "previous-boot-" + identity}
	raw, err := json.Marshal(owner)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, raw, queueFilePerm); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

// TestStaleLockReclaimDoesNotDeleteTheNewLock is review-gate finding #6:
// two processes can both read the same stale lock as verified-dead; the
// first reclaims and ACQUIRES the fresh lock, and the second reclaimer —
// still holding its stale verdict — then DELETES that fresh lock, leaving
// two writers inside one critical section. The reproduction: N concurrent
// Mutates against a previous-boot fixture; before the fix, at least one
// reclaimer removes a lock another reclaimer already re-acquired, and the
// concurrent-holder detector fires.
//
// The fix makes verify+break+reacquire ONE critical section: the reclaimer
// does not remove the stale lock — it RENAMES it away (atomic; only the
// first reclaimer's rename succeeds) and loops to Claim the now-free lock
// under its own label. A second reclaimer's rename fails (the stale lock
// is gone) and it re-loops, finding a live labelled lock it must block on.
func TestStaleLockReclaimDoesNotDeleteTheNewLock(t *testing.T) {
	dir := t.TempDir()
	store := NewQueueStore(filepath.Join(dir, "queue.json"))
	if err := os.MkdirAll(dir, queueDirPerm); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	previousBootFixture(t, store.LockPath())

	var inside atomic.Int64
	var doubleEntry atomic.Bool
	const reclaimers = 16
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < reclaimers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			// Each reclaimer runs against a FRESH previous-boot fixture, so
			// every round re-opens the reclaim window the finding names.
			for round := 0; round < 50; round++ {
				previousBootFixture(t, store.LockPath())
				_ = store.Mutate(func(rec *QueueRecord) error {
					n := inside.Add(1)
					if n > 1 {
						doubleEntry.Store(true)
					}
					defer inside.Add(-1)
					rec.LastSeq++
					rec.Items = append(rec.Items, QueueItem{ID: "f1"})
					return nil
				})
			}
		}()
	}
	close(start)
	wg.Wait()

	if doubleEntry.Load() {
		t.Fatal("two writers were inside the critical section at once — a reclaimer deleted a live lock")
	}
	if _, err := os.Stat(store.LockPath()); !os.IsNotExist(err) {
		t.Fatalf("the lock survived every release: %v", err)
	}
}

// TestBreakAbortsWhenTheFileChangedUnderneath is the DETERMINISTIC form of
// review-gate finding #6: a reclaimer that verified a stale snapshot must
// not remove whatever sits at the lock path NOW. The repro drives
// breakStaleLock against a snapshot of the stale bytes while the file at
// the path has already been replaced by a fresh labelled lock — exactly the
// A-claimed-B-removed interleaving the concurrent repro widens for.
// Before the fix, breakStaleLock removed by path and deleted the FRESH
// lock; with the fix it re-reads and aborts unless the bytes are still the
// ones it verified.
func TestBreakAbortsWhenTheFileChangedUnderneath(t *testing.T) {
	dir := t.TempDir()
	store := NewQueueStore(filepath.Join(dir, "queue.json"))
	if err := os.MkdirAll(dir, queueDirPerm); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	lockPath := store.LockPath()

	// (1) The stale record the reclaimer verifies.
	previousBootFixture(t, lockPath)
	staleBytes, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("read stale fixture: %v", err)
	}
	var staleOwner lockOwner
	if err := json.Unmarshal(staleBytes, &staleOwner); err != nil {
		t.Fatalf("decode stale: %v", err)
	}
	if !lockOwnerIsDead(staleOwner) {
		t.Fatal("fixture: the previous-boot owner must verify dead")
	}

	// (2) Meanwhile, another reclaimer already reclaimed and acquired: the
	// path now holds a FRESH, labelled, live lock.
	fresh := lockOwner{PID: os.Getpid(), BootID: bootIDIdentity()}
	freshBytes, err := json.Marshal(fresh)
	if err != nil {
		t.Fatalf("marshal fresh: %v", err)
	}
	if err := os.WriteFile(lockPath, freshBytes, queueFilePerm); err != nil {
		t.Fatalf("write fresh lock: %v", err)
	}

	// (3) The stale-armed reclaimer acts. The fix aborts: the bytes at the
	// path are not the ones it verified.
	if breakStaleLock(lockPath) {
		t.Fatal("breakStaleLock removed a lock whose bytes had changed under it")
	}
	after, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("the fresh lock was destroyed: %v", err)
	}
	if string(after) != string(freshBytes) {
		t.Fatalf("the fresh lock's bytes were altered: %s", after)
	}
}

// TestBreakStillFiresOnUnchangedStaleLock pins the other half: the re-read
// abort must not disarm the break against the lock it verified.
func TestBreakStillFiresOnUnchangedStaleLock(t *testing.T) {
	dir := t.TempDir()
	store := NewQueueStore(filepath.Join(dir, "queue.json"))
	if err := os.MkdirAll(dir, queueDirPerm); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	lockPath := store.LockPath()
	previousBootFixture(t, lockPath)
	if !breakStaleLock(lockPath) {
		t.Fatal("breakStaleLock did not fire on an unchanged verified-dead lock")
	}
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("the stale lock survived the break: %v", err)
	}
}
