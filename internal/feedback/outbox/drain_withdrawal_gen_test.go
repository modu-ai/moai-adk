package outbox

// The withdrawal invalidation test (review gate finding, P1): discardAll
// (consent off) emptied the queue and the spool but never advanced the
// store generation — so another flush's in-flight batch re-enqueued a
// discarded report, and a sender mid-search published it. The withdrawal
// bumps the generation like the purge does: every in-flight reader stops.

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

func TestWithdrawalBumpsTheGeneration(t *testing.T) {
	spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	before, gerr := bugreport.SpoolGeneration()
	if gerr != nil {
		t.Fatalf("read generation: %v", gerr)
	}

	// The withdrawal: consent off routes the drain through discardAll.
	consentOff(t)
	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}

	after, gerr := bugreport.SpoolGeneration()
	if gerr != nil {
		t.Fatalf("read generation: %v", gerr)
	}
	if after == before {
		t.Fatal("the withdrawal left the generation unchanged — an in-flight batch or sender keeps treating the withdrawn store as live")
	}
}
