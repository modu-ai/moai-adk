package outbox

// The ledger-reservation tests (review-gate hardening round, P2): a queue
// mutation that FAILED after recording the ledger used to leave its
// reservation behind — the fingerprint stamp AND a QueuedAt entry — and
// every retry added another. Three failed saves reached the DAILY CAP with
// an empty queue, and the next run classified the report `capped` (a
// decided outcome) and consumed it from the spool: data loss. A failed
// save's reservation is rolled back; only a save that LANDED counts.

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

func TestFailedSavesDoNotDoubleCountIntoTheCaps(t *testing.T) {
	spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	prev := queueCommitFailForTest
	t.Cleanup(func() { queueCommitFailForTest = prev })
	queueCommitFailForTest = true
	for i := 0; i < 3; i++ {
		if err := Drain(); err != nil {
			t.Fatalf("drain %d: %v", i+1, err)
		}
	}
	if got := queuedBugreportCount(t); got != 0 {
		t.Fatalf("the failed writes queued %d items, want none", got)
	}

	// No phantom reservations: the three failed saves left nothing in the
	// ledger's rolling window.
	ledger, err := loadLedger()
	if err != nil {
		t.Fatalf("load ledger: %v", err)
	}
	if len(ledger.QueuedAt) != 0 {
		t.Fatalf("the ledger carries %d QueuedAt entries after three FAILED saves — phantom reservations reached for the cap", len(ledger.QueuedAt))
	}

	// The report survives in the spool and queues once the failure clears.
	queueCommitFailForTest = false
	if err := Drain(); err != nil {
		t.Fatalf("clean drain: %v", err)
	}
	if got := queuedBugreportCount(t); got != 1 {
		t.Fatalf("queue holds %d items, want the report queued — it must not be capped or lost by phantom reservations", got)
	}
	ledger, err = loadLedger()
	if err != nil {
		t.Fatalf("load ledger: %v", err)
	}
	if len(ledger.QueuedAt) != 1 {
		t.Fatalf("the ledger carries %d QueuedAt entries, want exactly the successful save's", len(ledger.QueuedAt))
	}
}
