package outbox

// The in-lock generation re-check test (review gate finding, P2): the
// pre-loop check alone left a window — a purge completing AFTER the
// checks and BEFORE the queue mutation let the stale batch enqueue a
// withdrawn report over the removed store. The enqueue's own critical
// section (the queue lock) re-checks the generation, so a purge that
// lands at the mutation's doorstep stops the insert.

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

func TestQueueInsertRechecksTheGenerationUnderTheLock(t *testing.T) {
	spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	// The purge lands after BOTH pre-loop generation checks — at the
	// queue mutation's doorstep.
	beforeQueueMutationForTest = func() {
		if err := PurgeStores(); err != nil {
			t.Errorf("purge: %v", err)
		}
	}
	defer func() { beforeQueueMutationForTest = nil }()

	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}
	if got := queuedBugreportCount(t); got != 0 {
		t.Fatalf("the drain enqueued %d item(s) after a purge that completed at the mutation's doorstep — the in-lock re-check is missing", got)
	}
}
