package publish

// The unreadable-generation test (review finding, r8 — the ubuntu-only
// creates=0/green-locally class): in front of a PUBLIC act the generation
// check must be FAIL-CLOSED. A store whose generation marker cannot be
// read is indistinguishable from a withdrawn one — the check cannot open
// the gate on an unreadable answer, so the send stops with a dropped row
// instead of filing.

import (
	"context"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

func TestSendStopsWhenTheGenerationIsUnreadable(t *testing.T) {
	consentOn(t)
	_, item := payloadFixture(t)
	seedQueue(t, item)

	// Make the generation marker a DIRECTORY: its Stat answers, and the
	// answer is "not a regular file" — unreadable, deterministically, on
	// every platform and every runner speed.
	genPath, gerr := bugreport.SpoolGenerationPath()
	if gerr != nil {
		t.Fatalf("generation path: %v", gerr)
	}
	if err := os.MkdirAll(genPath, 0o700); err != nil {
		t.Fatalf("plant an unreadable generation marker: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(genPath) })

	stub := newStubRunner(true)
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	_, creates, comments := stub.recorded()
	if creates != 0 || comments != 0 {
		t.Fatalf("creates=%d comments=%d with an UNREADABLE generation marker — the check passed an answer it could not read and published over a state it could not verify", creates, comments)
	}
}
