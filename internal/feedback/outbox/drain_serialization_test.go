package outbox

import (
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/feedback"
)

// TestConcurrentDrainsEnqueueFingerprintOnce pins review-gate finding
// (P2, drain.go): the dedupe check and the ledger update were separate
// steps around the queue mutation, so two concurrent drains both read an
// empty ledger, both passed the dedupe window, and both enqueued the SAME
// fingerprint — the queue lock alone does not cover the ledger. The
// dedupe-check → enqueue → ledger-record sequence must be ONE cross-process
// critical section.
func TestConcurrentDrainsEnqueueFingerprintOnce(t *testing.T) {
	spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	const drains = 2
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < drains; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_ = Drain()
		}()
	}
	close(start)
	wg.Wait()

	rec, err := BugreportQueueStore().Load()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	if len(rec.Items) != 1 {
		t.Fatalf("queue holds %d items for one fingerprint — %d drains passed the dedupe check together", len(rec.Items), drains)
	}
}

// TestConcurrentDrainsRespectDailyCap extends the same critical section to
// the rolling caps: four spool entries carry three DISTINCT fingerprints
// (kind is a fingerprint input), and the daily cap is 3 — concurrent
// drains must never enqueue past it.
func TestConcurrentDrainsRespectDailyCap(t *testing.T) {
	spoolFixture(t,
		bugreport.KindPanic,
		bugreport.KindHookHandlerFailure,
		bugreport.KindHookHandlerFailure,
		bugreport.KindInternalError,
	)
	consentOn(t)
	validBuildForTest(t)

	const drains = 4
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < drains; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_ = Drain()
		}()
	}
	close(start)
	wg.Wait()

	rec, err := BugreportQueueStore().Load()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	if len(rec.Items) > 3 {
		t.Fatalf("queue holds %d items — the daily cap of 3 was breached by concurrent drains", len(rec.Items))
	}
	seen := map[string]bool{}
	for _, it := range rec.Items {
		if seen[it.Fingerprint] {
			t.Fatalf("fingerprint %s enqueued twice", it.Fingerprint)
		}
		seen[it.Fingerprint] = true
	}
}

// TestQueueBoundHoldsAtEnqueue pins review-gate finding (P2, drain.go):
// EnforceQueueBound existed but the drain never called it — repeated runs
// grew the queue to 21 over the configured cap of 20. The append and the
// oldest-item removal must share the enqueue mutation.
func TestQueueBoundHoldsAtEnqueue(t *testing.T) {
	spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	store := BugreportQueueStore()
	// Pre-fill the queue to its bound.
	err := store.Mutate(func(rec *feedback.QueueRecord) error {
		for i := 0; i < 20; i++ {
			rec.LastSeq++
			rec.Items = append(rec.Items, feedback.QueueItem{ID: "old" + itoa(i)})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed queue: %v", err)
	}
	rowsBefore := len(outboxRows(t))

	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	if len(rec.Items) != 20 {
		t.Fatalf("queue holds %d items after the drain, want the bound 20 — the bound was not enforced at enqueue", len(rec.Items))
	}
	// The oldest pre-filled item was the one dropped.
	if rec.Items[0].ID != "old1" {
		t.Fatalf("oldest surviving item = %s, want old1 (old0 was dropped)", rec.Items[0].ID)
	}
	dropped := countOutcome(outboxRows(t)[rowsBefore:], "dropped")
	if dropped == 0 {
		t.Fatal("no dropped row was recorded for the bound eviction")
	}
}
