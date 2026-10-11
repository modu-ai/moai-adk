package outbox

// The purge-vs-writer race test (review gate finding, P2): a queue mutation
// loaded a report and paused while HOLDING the queue lock; PurgeStores'
// queue removal was a bare os.Remove that IGNORED the lock — it deleted the
// file under the paused writer, and the writer's save then RESURRECTED the
// purged report. The purge's queue step now goes through the queue's own
// mutation path: it waits for the paused writer's save to commit, empties
// the queue under the same lock, and only then removes the file — the
// writer has already finished, so nothing resurrects.

import (
	"context"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/feedback"
)

func TestPurgeDoesNotResurrectAQueueReport(t *testing.T) {
	spoolFixture(t) // establishes the test's isolated MOAI_HOME
	consentOn(t)

	store := BugreportQueueStore()
	// One committed report in the queue.
	if err := store.Mutate(func(rec *feedback.QueueRecord) error {
		rec.LastSeq++
		rec.Items = append(rec.Items, feedback.QueueItem{ID: "f1", Title: "purge-race fixture", Fingerprint: "aaaaaaaaaaaaaaaa"})
		return nil
	}); err != nil {
		t.Fatalf("seed queue: %v", err)
	}

	// The paused writer: it acquires the queue lock, signals, and parks —
	// its record (carrying the report) saves only when the gate opens.
	locked := make(chan struct{})
	gate := make(chan struct{})
	writerDone := make(chan error, 1)
	go func() {
		writerDone <- store.MutateContext(context.Background(), func(rec *feedback.QueueRecord) error {
			close(locked)
			<-gate
			return nil // save the loaded record: the resurrection vehicle
		})
	}()
	select {
	case <-locked:
	case <-time.After(2 * time.Second):
		t.Fatal("the writer never acquired the queue lock")
	}

	// Purge runs while the writer holds the lock. GREEN: its queue step
	// blocks on the lock. RED: the bare removal deletes the file under the
	// paused writer at once.
	purgeDone := make(chan error, 1)
	go func() { purgeDone <- PurgeStores() }()
	time.Sleep(150 * time.Millisecond)

	close(gate)
	if err := <-writerDone; err != nil {
		t.Fatalf("paused writer: %v", err)
	}
	select {
	case err := <-purgeDone:
		if err != nil {
			t.Fatalf("purge: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("purge never completed — the queue step deadlocked behind the paused writer")
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rec.Items) != 0 {
		t.Fatalf("queue holds %d item(s) after purge — the paused writer's save resurrected the purged report", len(rec.Items))
	}
}
