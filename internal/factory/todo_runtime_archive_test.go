package factory

import (
	"strings"
	"testing"
	"time"
)

// todo_runtime_archive_test.go — card t1542: the runtime completion's
// archive authority. The P1 audit finding was that a completion the runtime
// recorded had no path that archives the card: the report row landed and the
// card sat live forever, so a leader had to run `todo done` for every lane
// completion. ArchiveOnRuntimeCompletion is that path.

// TestArchiveOnRuntimeCompletionClosesTheLaneCard — one locked write
// archives the live card with the verdict the mechanical edge answered; a
// re-run reconciles (already closed reads as closed); a card in neither the
// queue nor the archive refuses, and the refusal writes nothing.
func TestArchiveOnRuntimeCompletionClosesTheLaneCard(t *testing.T) {
	_, s, card := runtimeFixture(t)
	verdict := LandingVerdict{Verdict: LandingLanded, Ref: "origin/main", At: time.Now().UTC().Format(time.RFC3339)}
	if err := s.ArchiveOnRuntimeCompletion(card, verdict); err != nil {
		t.Fatalf("close: %v", err)
	}
	rec, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Items) != 0 {
		t.Fatalf("live items = %d, want 0", len(rec.Items))
	}
	if len(rec.Archived) != 1 || rec.Archived[0].Item.ID != card {
		t.Fatalf("archive holds %d entries, want exactly %s", len(rec.Archived), card)
	}
	v := rec.Archived[0].LandingVerdict
	if v == nil || v.Verdict != LandingLanded || v.Ref != "origin/main" {
		t.Fatalf("landing verdict = %+v, want landed against origin/main", v)
	}
	// Idempotent reconciliation: closing again is a no-op, never an error —
	// the delivery edge runs it on every observation.
	if err := s.ArchiveOnRuntimeCompletion(card, verdict); err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	// An unknown card refuses, and the refusal writes nothing.
	if err := s.ArchiveOnRuntimeCompletion("t404", verdict); err == nil || !strings.Contains(err.Error(), "t404") {
		t.Fatalf("unknown card = %v, want a named refusal", err)
	}
}
