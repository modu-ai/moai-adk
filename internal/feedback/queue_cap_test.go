package feedback

// The manual-queue cap test (review gate finding, P2): the load-side cap
// was added with the automatic participation queue in mind, but the
// manual entry point (EnqueueMasked) enforces nothing at save time — a
// manual queue grown past the cap became UNREADABLE, and with the queue
// unreadable every resend, remove, and add failed with it. The save now
// rejects an enqueue that would push the store past the same cap, so a
// manual queue can always be read back.

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnqueueMaskedRejectsAQueuePastTheCap(t *testing.T) {
	s := NewQueueStore(filepath.Join(t.TempDir(), "queue.json"))

	// Large manual reports until the next one would cross the cap. The
	// last REJECTED save must leave the store readable.
	for i := 0; i < 16; i++ {
		if _, err := s.EnqueueMasked(Result{Verdict: VerdictOK, Title: "manual fixture", Body: strings.Repeat("b", 62*1024)}); err != nil {
			t.Fatalf("enqueue %d: %v", i, err)
		}
	}
	if _, lerr := s.Load(); lerr != nil {
		t.Fatalf("the store is already unreadable before the rejecting enqueue: %v", lerr)
	}
	_, ferr := s.EnqueueMasked(Result{Verdict: VerdictOK, Title: "overflow", Body: strings.Repeat("c", 62*1024)})
	if ferr == nil {
		t.Fatal("an enqueue past the size cap succeeded — the store it creates cannot be loaded back")
	}
	if !errors.Is(ferr, ErrQueueFull) {
		t.Fatalf("the over-cap enqueue failed with %v, want ErrQueueFull", ferr)
	}
	if _, lerr := s.Load(); lerr != nil {
		t.Fatalf("the store became unreadable anyway: %v", lerr)
	}
}
