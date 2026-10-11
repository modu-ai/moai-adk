package publish

// The purge-vs-model-call race (review gate finding, P1): the summarizer's
// model call can take seconds, and a purge completing DURING it left the
// send holding a withdrawn report — the generation was only re-checked
// BEFORE the call, so CreateIssue still filed the issue. The generation is
// re-checked AFTER the summary too: a purge landing mid-summary stops the
// create.

import (
	"context"
	"testing"

	"github.com/modu-ai/moai-adk/internal/feedback/outbox"
)

func TestSendStopsWhenTheStoreWasPurgedMidSummary(t *testing.T) {
	consentOn(t)
	_, item := payloadFixture(t)
	seedQueue(t, item)

	afterSummaryForTest = func() {
		_ = outbox.PurgeStores()
	}
	defer func() { afterSummaryForTest = nil }()

	stub := newStubRunner(true)
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	_, creates, _ := stub.recorded()
	if creates != 0 {
		t.Fatalf("creates=%d after a purge that completed during the model call — the send filed an issue over a withdrawn report", creates)
	}
}
