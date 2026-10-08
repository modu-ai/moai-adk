package publish

// The generation-read-order test (review gate finding, P1): the send read
// the queue and THEN read the generation — a purge completing between the
// two left the withdrawn report wearing the NEW generation number, and
// every later check passed it. The generation is read BEFORE the queue
// data: the baseline is taken first, so a purge that lands anywhere after
// the read shows up as an advance.

import (
	"context"
	"testing"

	"github.com/modu-ai/moai-adk/internal/feedback/outbox"
)

func TestSendStopsWhenThePurgeLandsAfterTheQueueRead(t *testing.T) {
	consentOn(t)
	_, item := payloadFixture(t)
	seedQueue(t, item)

	afterLiveQueueReadForTest = func() {
		// The purge completes between the queue read and (the old code's)
		// generation read: the report was withdrawn, but the new code
		// must still see it — the baseline was taken BEFORE the read.
		_ = outbox.PurgeStores()
	}
	defer func() { afterLiveQueueReadForTest = nil }()

	stub := newStubRunner(true)
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	_, creates, comments := stub.recorded()
	if creates != 0 || comments != 0 {
		t.Fatalf("creates=%d comments=%d — the report withdrawn between the queue read and the generation read published anyway (it wore the new generation)", creates, comments)
	}
}
