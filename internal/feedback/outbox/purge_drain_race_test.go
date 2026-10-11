package outbox

// The purge-vs-drain tests (review gate findings, P2):
//
//   - A batch the drain read BEFORE the purge must not resurrect purged
//     reports: once the purge completes, the store is GONE — the drain's
//     in-memory batch is stale, and enqueueing it recreates reports the
//     user withdrew. The drain re-checks the spool generation before each
//     item and stops when the store was purged mid-batch.
//   - A queue grown past the size cap must still be purgable: the cap
//     failure is an UNREADABLE store, the same purge-must-win class as a
//     corrupted one.

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

func TestDrainStopsWhenTheStoreWasPurgedMidBatch(t *testing.T) {
	spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	// The seam runs between the drain's read and its item loop: the purge
	// completes there, exactly the interleaving the finding reproduced.
	spoolAfterReadForTest = func() {
		if err := PurgeStores(); err != nil {
			t.Errorf("purge: %v", err)
		}
	}
	defer func() { spoolAfterReadForTest = nil }()

	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}
	if got := queuedBugreportCount(t); got != 0 {
		t.Fatalf("the drain enqueued %d item(s) after the purge completed — the in-memory batch resurrected withdrawn reports", got)
	}
}

func TestPurgeRemovesAnOversizedQueue(t *testing.T) {
	spoolFixture(t)
	consentOn(t)

	queuePath, err := StorePath(QueueFileName)
	if err != nil {
		t.Fatalf("store path: %v", err)
	}
	if err := os.WriteFile(queuePath, []byte(strings.Repeat("x", 1024*1024+64)), 0o600); err != nil {
		t.Fatalf("seed oversized queue: %v", err)
	}

	perr := PurgeStores()
	if perr != nil {
		t.Fatalf("purge: %v", perr)
	}
	if _, serr := os.Lstat(queuePath); !os.IsNotExist(serr) {
		t.Fatalf("the oversized queue file survived the purge (lstat: %v)", serr)
	}
}
