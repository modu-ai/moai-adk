package outbox

// The partial-consumption tests (review-gate finding, P2): a retryable
// queue-write failure — context expiry during a lock wait, a lost budget,
// an unreadable ledger — used to record the failure row and then consume
// the WHOLE spool batch. The failed item and everything after it were lost
// (nil return, spool deleted, report gone). A retryable failure excludes
// the item from consumption: only the successfully processed prefix leaves
// the spool, and the next drain retries the rest.

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

// spoolKinds reads back which kinds remain in the spool after a drain.
func spoolKinds(t *testing.T, path string) []bugreport.Kind {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read spool: %v", err)
	}
	var kinds []bugreport.Kind
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry bugreport.SpoolEntry
		if json.Unmarshal([]byte(line), &entry) != nil {
			continue
		}
		kinds = append(kinds, entry.Kind)
	}
	return kinds
}

func containsKind(kinds []bugreport.Kind, k bugreport.Kind) bool {
	for _, got := range kinds {
		if got == k {
			return true
		}
	}
	return false
}

func queuedBugreportCount(t *testing.T) int {
	t.Helper()
	rec, err := BugreportQueueStore().Load()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	return len(rec.Items)
}

func blockQueueWriteForTest(t *testing.T, blocked bugreport.Kind) {
	t.Helper()
	prev := queueWriteBlockForTest
	t.Cleanup(func() { queueWriteBlockForTest = prev })
	queueWriteBlockForTest = func(entry bugreport.SpoolEntry) bool {
		return entry.Kind == blocked
	}
}

// TestRetryableFailureKeepsTheItemAndTheTailInTheSpool: the middle entry's
// queue write fails — the first queues, the failed second AND the untouched
// third stay in the spool for the next drain.
func TestRetryableFailureKeepsTheItemAndTheTailInTheSpool(t *testing.T) {
	spoolPath := spoolFixture(t,
		bugreport.KindPanic,
		bugreport.KindHookHandlerFailure,
		bugreport.KindInternalError,
	)
	consentOn(t)
	validBuildForTest(t)
	blockQueueWriteForTest(t, bugreport.KindHookHandlerFailure)

	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}

	if got := queuedBugreportCount(t); got != 1 {
		t.Fatalf("queue holds %d items, want exactly the first entry's", got)
	}
	remaining := spoolKinds(t, spoolPath)
	if !containsKind(remaining, bugreport.KindHookHandlerFailure) {
		t.Fatalf("the failed item was consumed from the spool — a lost report (remaining: %v)", remaining)
	}
	if !containsKind(remaining, bugreport.KindInternalError) {
		t.Fatalf("the entry after the failure was consumed from the spool (remaining: %v)", remaining)
	}
	if containsKind(remaining, bugreport.KindPanic) {
		t.Fatalf("the successfully queued entry stayed in the spool (remaining: %v)", remaining)
	}
	rows := outboxRows(t)
	if countOutcome(rows, "queued") != 1 {
		t.Fatalf("queued rows = %d, want exactly the first entry's", countOutcome(rows, "queued"))
	}
	if !hasOutcome(rows, "dropped") {
		t.Fatalf("the failed item recorded no dropped row: %v", rows)
	}
}

// TestRetryableFailureOnLastItemKeepsItInTheSpool is the finding's literal
// repro: the deadline (or any retryable failure) lands during the LAST
// item's lock wait — the earlier items queue, the failed item stays in the
// spool instead of being consumed as lost.
func TestRetryableFailureOnLastItemKeepsItInTheSpool(t *testing.T) {
	spoolPath := spoolFixture(t,
		bugreport.KindPanic,
		bugreport.KindInternalError,
	)
	consentOn(t)
	validBuildForTest(t)
	blockQueueWriteForTest(t, bugreport.KindInternalError)

	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}

	if got := queuedBugreportCount(t); got != 1 {
		t.Fatalf("queue holds %d items, want exactly the first entry's", got)
	}
	remaining := spoolKinds(t, spoolPath)
	if !containsKind(remaining, bugreport.KindInternalError) {
		t.Fatalf("the last item was consumed from the spool despite its failed write — a lost report (remaining: %v)", remaining)
	}
	if containsKind(remaining, bugreport.KindPanic) {
		t.Fatalf("the successfully queued entry stayed in the spool (remaining: %v)", remaining)
	}
}

// TestCleanDrainConsumesTheWholeBatch: with no failure every entry queues
// and the whole read batch leaves the spool — the partial rule never
// retires the full-batch case.
func TestCleanDrainConsumesTheWholeBatch(t *testing.T) {
	spoolPath := spoolFixture(t,
		bugreport.KindPanic,
		bugreport.KindHookHandlerFailure,
		bugreport.KindInternalError,
	)
	consentOn(t)
	validBuildForTest(t)

	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}

	if got := queuedBugreportCount(t); got != 3 {
		t.Fatalf("queue holds %d items, want all three", got)
	}
	if remaining := spoolKinds(t, spoolPath); len(remaining) != 0 {
		t.Fatalf("the spool kept %v after a clean drain", remaining)
	}
}
