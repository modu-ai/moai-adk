package outbox

// The orphaned-ledger tests (review-gate residual, P2): the ledger record
// lands inside the queue-lock callback, but the queue-file replacement
// happens after it — a crash or write failure in between leaves a record
// whose report never entered the queue. The next drain used to judge the
// report `deduped` (from the ledger) and consume it: a lost report. The
// dedupe check now consults the two places a LEGITIMATE record's report
// lives — the live queue and the sent history — and re-queues when both
// lack the fingerprint.

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

// TestOrphanedLedgerRecordRequeuesInsteadOfDeduping drives the write-
// failure shape: drain one records the ledger and then fails the queue
// replacement; drain two (failure cleared) must re-queue the report, not
// dedupe it away.
func TestOrphanedLedgerRecordRequeuesInsteadOfDeduping(t *testing.T) {
	spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	prev := queueCommitFailForTest
	t.Cleanup(func() { queueCommitFailForTest = prev })
	queueCommitFailForTest = true
	if err := Drain(); err != nil {
		t.Fatalf("drain one: %v", err)
	}
	if got := queuedBugreportCount(t); got != 0 {
		t.Fatalf("the failed write queued %d items, want none", got)
	}

	queueCommitFailForTest = false
	if err := Drain(); err != nil {
		t.Fatalf("drain two: %v", err)
	}
	if got := queuedBugreportCount(t); got != 1 {
		t.Fatalf("queue holds %d items, want the orphaned report re-queued — it was deduped away and lost", got)
	}
}

// TestCrashOrphanedLedgerRecordRequeues seeds the crash shape directly —
// the ledger already carries the fingerprint, the queue does not, and the
// spool still holds the report — and pins the same recovery.
func TestCrashOrphanedLedgerRecordRequeues(t *testing.T) {
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
		t.Fatalf("queue holds %d items, want the crash-orphaned report re-queued — it was deduped away and lost", got)
	}
	if remaining := spoolKinds(t, spoolPath); len(remaining) != 0 {
		t.Fatalf("the spool kept %v after the report was queued", remaining)
	}
}

// TestSentFingerprintStillDedupes pins the discriminator's other side: a
// record whose report was SENT (the outbox log carries the sent row) stays
// deduped inside the window — the repair must not re-queue a handled
// fingerprint.
func TestSentFingerprintStillDedupes(t *testing.T) {
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
	// The sent history: the report was queued, sent, and removed.
	title := "[auto-report] panic " + fp
	if err := AppendOutbox(OutboxRow{Outcome: "sent", Title: title, Fingerpr: fp}); err != nil {
		t.Fatalf("seed outbox: %v", err)
	}

	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}
	if got := queuedBugreportCount(t); got != 0 {
		t.Fatalf("queue holds %d items, want the sent fingerprint deduped", got)
	}
	if remaining := spoolKinds(t, spoolPath); len(remaining) != 0 {
		t.Fatalf("the deduped report stayed in the spool (remaining: %v)", remaining)
	}
}
