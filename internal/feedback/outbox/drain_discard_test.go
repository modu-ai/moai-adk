package outbox

// The terminal-discard tests (review-gate finding 8, P2): a send-attempt-
// exhausted discard removed the queue item and logged a `dropped` row, but
// left the LEDGER record — and the record, with no queue item and no sent
// row, read exactly like an ORPHAN. The next drain inside the 7-day window
// re-enrolled the discarded report: attempts reset, the model budget spent
// again. A terminal discard marker in the ledger closes the shape: the
// recovery path re-queues only genuinely unfinished reservations.

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

// TestDiscardedReportIsNotReEnrolled: the ledger carries the fingerprint
// (queued once), the item was dropped at the attempt limit, and the spool
// still holds a recurring capture of the same crash. The drain must judge
// the report TERMINALLY DISCARDED — consumed with a decision row, never
// re-queued.
func TestDiscardedReportIsNotReEnrolled(t *testing.T) {
	spoolPath := spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	entries, err := bugreport.ReadSpool()
	if err != nil || len(entries) != 1 {
		t.Fatalf("read spool: %v (%d entries)", err, len(entries))
	}
	fp := fingerprintOf(entries[0])
	ledger, err := loadLedger()
	if err != nil {
		t.Fatalf("load ledger: %v", err)
	}
	ledger.RecordQueued(fp, clock())
	ledger.MarkDiscarded(fp, clock())
	if err := saveLedger(ledger); err != nil {
		t.Fatalf("seed ledger: %v", err)
	}

	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}
	if got := queuedBugreportCount(t); got != 0 {
		t.Fatalf("queue holds %d items, want the discarded report NOT re-enrolled — the recovery reset its attempts and re-spent the model budget", got)
	}
	if remaining := spoolKinds(t, spoolPath); len(remaining) != 0 {
		t.Fatalf("the discarded report stayed in the spool (remaining: %v)", remaining)
	}
}

// TestUnfinishedReservationStillRequeues: the same ledger shape WITHOUT the
// discard marker — the genuine orphan — still re-queues. The marker gates
// the recovery, it does not remove it.
func TestUnfinishedReservationStillRequeues(t *testing.T) {
	spoolPath := spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	entries, err := bugreport.ReadSpool()
	if err != nil || len(entries) != 1 {
		t.Fatalf("read spool: %v (%d entries)", err, len(entries))
	}
	fp := fingerprintOf(entries[0])
	ledger, err := loadLedger()
	if err != nil {
		t.Fatalf("load ledger: %v", err)
	}
	ledger.RecordQueued(fp, clock())
	if err := saveLedger(ledger); err != nil {
		t.Fatalf("seed ledger: %v", err)
	}

	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}
	if got := queuedBugreportCount(t); got != 1 {
		t.Fatalf("queue holds %d items, want the genuine orphan re-queued", got)
	}
	if remaining := spoolKinds(t, spoolPath); len(remaining) != 0 {
		t.Fatalf("the re-queued orphan stayed in the spool (remaining: %v)", remaining)
	}
}
