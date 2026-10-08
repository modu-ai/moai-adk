package feedback

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

// previousBootFixture writes a lock file whose recorded owner ran on a
// PREVIOUS boot (a boot id different from this machine's wire form) — a
// verified-dead owner by the boot-comparison path, whatever lives at the
// pid today. The current pid keeps the record self-consistent otherwise.
func previousBootFixture(t *testing.T, path string) {
	t.Helper()
	identity := atomicfile.BootIDIdentity()
	if identity == "" {
		t.Skip("no boot identity on this platform; the boot-comparison path is not exercised")
	}
	owner := atomicfile.LockOwner{PID: os.Getpid(), BootID: "previous-boot-" + identity}
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
// The break machinery's deterministic pins live beside the machinery in
// internal/config/atomicfile/section_test.go; this test drives the QUEUE
// level, where the finding manifests as a lost update.
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
