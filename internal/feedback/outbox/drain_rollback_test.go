package outbox

// The rollback ownership tests (review-gate finding 4, P2): a queue
// mutation that fails at LOCK ACQUISITION ran no callback and created NO
// reservation, so its rollback must not touch the ledger at all — the
// unconditional delete used to erase a PRIOR SUCCESS's fingerprint record,
// re-opening the dedupe window for a report already queued (double
// publication on the next drain). The rollback acts only on a reservation
// its own attempt verifiably owns: the recorded stamp equals the attempt's.

import (
	"context"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

// rollbackPriorStamp / rollbackAttemptStamp: two distinct RFC3339 stamps an
// hour apart — the prior success's and the failed attempt's.
var (
	rollbackPriorStamp   = time.Date(2026, 10, 8, 9, 0, 0, 0, time.UTC).Format(time.RFC3339)
	rollbackAttemptStamp = time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC).Format(time.RFC3339)
)

// rollbackLedgerFixture establishes the temporary MOAI_HOME (spoolFixture
// seeds an empty spool — the rollback only needs the store path) and loads
// the current ledger for seeding.
func rollbackLedgerFixture(t *testing.T) {
	t.Helper()
	spoolFixture(t, bugreport.KindPanic)
}

func rollbackSeedLedger(t *testing.T, seed func(l *Ledger)) {
	t.Helper()
	ledger, err := loadLedger()
	if err != nil {
		t.Fatalf("load ledger: %v", err)
	}
	seed(ledger)
	if err := saveLedger(ledger); err != nil {
		t.Fatalf("save ledger: %v", err)
	}
}

func rollbackLoadLedger(t *testing.T) *Ledger {
	t.Helper()
	ledger, err := loadLedger()
	if err != nil {
		t.Fatalf("load ledger: %v", err)
	}
	return ledger
}

// TestRollbackNeverDeletesAPriorSuccessRecord: the fp's ledger record
// belongs to a PRIOR success (a different stamp); a failed attempt's
// rollback must leave it — and its QueuedAt entry — exactly in place.
func TestRollbackNeverDeletesAPriorSuccessRecord(t *testing.T) {
	rollbackLedgerFixture(t)
	const fp = "aaaaaaaaaaaaaaaa"
	rollbackSeedLedger(t, func(l *Ledger) {
		l.FingerprintSeen[fp] = rollbackPriorStamp
		l.QueuedAt = append(l.QueuedAt, rollbackPriorStamp)
	})

	rollbackLedgerRecord(context.Background(), fp, rollbackAttemptStamp)

	ledger := rollbackLoadLedger(t)
	if got, ok := ledger.FingerprintSeen[fp]; !ok || got != rollbackPriorStamp {
		t.Fatalf("the prior success's fingerprint record was deleted (got %q, ok=%v) — the rollback erased a record this attempt never wrote", got, ok)
	}
	if len(ledger.QueuedAt) != 1 || ledger.QueuedAt[0] != rollbackPriorStamp {
		t.Fatalf("the prior success's QueuedAt entry was removed: %v — the rollback erased a reservation this attempt never wrote", ledger.QueuedAt)
	}
}

// TestRollbackRemovesExactlyItsOwnReservation: the fp's record IS this
// attempt's stamp — the rollback removes it and exactly its own QueuedAt
// entry, leaving the other reservation alone.
func TestRollbackRemovesExactlyItsOwnReservation(t *testing.T) {
	rollbackLedgerFixture(t)
	const fp = "bbbbbbbbbbbbbbbb"
	rollbackSeedLedger(t, func(l *Ledger) {
		l.QueuedAt = append(l.QueuedAt, rollbackPriorStamp) // another fingerprint's entry
		l.FingerprintSeen[fp] = rollbackAttemptStamp
		l.QueuedAt = append(l.QueuedAt, rollbackAttemptStamp)
	})

	rollbackLedgerRecord(context.Background(), fp, rollbackAttemptStamp)

	ledger := rollbackLoadLedger(t)
	if _, ok := ledger.FingerprintSeen[fp]; ok {
		t.Fatalf("the attempt's own fingerprint record survived its rollback")
	}
	if len(ledger.QueuedAt) != 1 || ledger.QueuedAt[0] != rollbackPriorStamp {
		t.Fatalf("the rollback disturbed another fingerprint's reservation: %v", ledger.QueuedAt)
	}
}
