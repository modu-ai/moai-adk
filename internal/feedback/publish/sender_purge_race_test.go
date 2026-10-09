package publish

// The purge-vs-send race (review gate finding, P1): the duplicate search
// is a gh round-trip, and a purge completing DURING the search left the
// sender holding a withdrawn report — which it then created an issue
// from. The send re-checks the store generation right before the model
// call and the create: a purge that advanced the generation mid-search
// stops the send instead of filing over the user's withdrawal.

import (
	"context"
	"testing"

	"github.com/modu-ai/moai-adk/internal/feedback/outbox"
)

func TestSendStopsWhenTheStoreWasPurgedMidSearch(t *testing.T) {
	consentOn(t)
	_, item := payloadFixture(t)
	seedQueue(t, item)

	stub := newStubRunner(true)
	stub.onSearch = func(s *stubRunner) {
		// The purge completes while the gh search runs — the store the
		// item belonged to is gone by the time the search returns.
		_ = outbox.PurgeStores()
	}

	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	_, creates, comments := stub.recorded()
	if creates != 0 || comments != 0 {
		t.Fatalf("creates=%d comments=%d after the purge completed — the send published over a withdrawn report", creates, comments)
	}
}
