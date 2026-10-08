package outbox

// The purge-vs-held-lock test (review gate finding, P2): the corrupted-
// queue repair taught the purge to ignore the queue mutation's parse
// failure, but the blanket `_ =` also swallowed a LOCK-ACQUIRE failure —
// a purge that runs while another holder owns the queue lock removed the
// file anyway, and the holder's save resurrected the purged report: the
// exact resurrection the lock-serialized queue step exists to prevent. A
// lock-acquire failure must propagate; only a corrupted-queue parse
// failure may be ignored.

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/feedback"
)

func TestPurgePropagatesWhenTheQueueLockIsHeld(t *testing.T) {
	spoolFixture(t)
	consentOn(t)

	store := BugreportQueueStore()
	if err := store.Mutate(func(rec *feedback.QueueRecord) error {
		rec.LastSeq++
		rec.Items = append(rec.Items, feedback.QueueItem{ID: "f1", Title: "held-lock fixture", Fingerprint: "aaaaaaaaaaaaaaaa"})
		return nil
	}); err != nil {
		t.Fatalf("seed queue: %v", err)
	}
	queuePath, err := StorePath(QueueFileName)
	if err != nil {
		t.Fatalf("store path: %v", err)
	}

	// A foreign holder claims the queue lock and keeps it for the whole
	// purge attempt. ClaimSection's owner records come from this process,
	// so a second claim in the same process is refused by the CAS, not by
	// ownership.
	release, err := atomicfile.ClaimSection(context.Background(), store.LockPath(), 0o600, 1, time.Millisecond)
	if err != nil {
		t.Fatalf("claim the queue lock as the foreign holder: %v", err)
	}
	defer func() { _ = release() }()

	perr := PurgeStores()
	if perr == nil {
		t.Fatal("purge returned nil while another holder owned the queue lock — the lock-acquire failure was swallowed and the queue file was removed under the holder")
	}
	if _, serr := os.Lstat(queuePath); serr != nil {
		t.Fatalf("the queued report's file was removed while the lock was held (lstat: %v)", serr)
	}
}
