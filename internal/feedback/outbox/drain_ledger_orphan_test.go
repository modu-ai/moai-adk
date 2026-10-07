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
	"time"

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

// The adopt-or-cancel tests (review-gate findings 5 and 9, P2): the
// orphaned reservation already occupies a rolling-cap slot. The recovery
// used to judge the caps INCLUDING that slot and then RecordQueued a SECOND
// entry — at the daily cap the recovery target was cap-judged and DELETED
// from the spool (count-and-discard), and below it the report counted twice
// into the caps. The recovery now ADOPTS the orphan's slot: the caps
// judgment excludes the orphan's own reservation, and the re-queue replaces
// the orphan's QueuedAt entry instead of appending.

// TestOrphanRecoveryAtTheCapAdoptsItsSlot (finding 9's fixture): the
// orphaned reservation plus two other reports fill the DAILY CAP of 3. The
// recovery target must not be cap-judged by its own orphan slot and
// consumed — it re-queues into that slot.
func TestOrphanRecoveryAtTheCapAdoptsItsSlot(t *testing.T) {
	spoolPath := spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	now := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	SetClockForTest(func() time.Time { return now })
	t.Cleanup(func() { SetClockForTest(nil) })

	entries, err := bugreport.ReadSpool()
	if err != nil || len(entries) != 1 {
		t.Fatalf("read spool: %v (%d entries)", err, len(entries))
	}
	fp := fingerprintOf(entries[0])
	ledger, err := loadLedger()
	if err != nil {
		t.Fatalf("load ledger: %v", err)
	}
	ledger.RecordQueued(fp, now) // the orphan's slot
	ledger.QueuedAt = append(ledger.QueuedAt, // two other reports' slots
		now.Add(-1*time.Hour).UTC().Format(time.RFC3339),
		now.Add(-2*time.Hour).UTC().Format(time.RFC3339))
	if err := saveLedger(ledger); err != nil {
		t.Fatalf("seed ledger: %v", err)
	}

	// The recovery runs an hour later: the three seeded slots are still
	// inside the rolling 24-hour window.
	now = now.Add(time.Hour)
	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}

	if got := queuedBugreportCount(t); got != 1 {
		t.Fatalf("queue holds %d items — the recovery target was cap-judged against its own orphan slot and deleted from the spool", got)
	}
	if remaining := spoolKinds(t, spoolPath); len(remaining) != 0 {
		t.Fatalf("the spool kept %v — the report should have been re-queued", remaining)
	}
	ledger, err = loadLedger()
	if err != nil {
		t.Fatalf("load ledger: %v", err)
	}
	if len(ledger.QueuedAt) != 3 {
		t.Fatalf("the ledger carries %d QueuedAt entries, want 3 — the adopted slot must not add a second reservation for one report", len(ledger.QueuedAt))
	}
	if got := ledger.FingerprintSeen[fp]; got != now.UTC().Format(time.RFC3339) {
		t.Fatalf("the adopted record kept the orphan's stamp %q, want the recovery's own stamp %q", got, now.UTC().Format(time.RFC3339))
	}
}

// TestOrphanRecoveryBelowTheCapDoesNotDoubleCount (finding 5's shape): with
// only the orphan's slot in the ledger, the re-queue must REUSE that slot —
// one report, one reservation.
func TestOrphanRecoveryBelowTheCapDoesNotDoubleCount(t *testing.T) {
	spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	now := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	SetClockForTest(func() time.Time { return now })
	t.Cleanup(func() { SetClockForTest(nil) })

	entries, err := bugreport.ReadSpool()
	if err != nil || len(entries) != 1 {
		t.Fatalf("read spool: %v (%d entries)", err, len(entries))
	}
	fp := fingerprintOf(entries[0])
	ledger, err := loadLedger()
	if err != nil {
		t.Fatalf("load ledger: %v", err)
	}
	ledger.RecordQueued(fp, now)
	if err := saveLedger(ledger); err != nil {
		t.Fatalf("seed ledger: %v", err)
	}

	now = now.Add(time.Hour)
	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}

	if got := queuedBugreportCount(t); got != 1 {
		t.Fatalf("queue holds %d items, want the orphan re-queued", got)
	}
	ledger, err = loadLedger()
	if err != nil {
		t.Fatalf("load ledger: %v", err)
	}
	if len(ledger.QueuedAt) != 1 {
		t.Fatalf("the ledger carries %d QueuedAt entries for ONE report — the recovery double-counted into the caps", len(ledger.QueuedAt))
	}
}
