package ledger

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// lane-19 24th review P2: a card that leaves MERGING while its queue row is still open and
// is claimed again at a higher epoch has a row nobody can write (the epoch fence) and
// Enqueue refuses to replace ("already in the queue"). Enqueue restarts a row of an older
// card epoch the way it restarts a failed one.
func TestEnqueue_RestartsARowOfAnOlderCardEpoch(t *testing.T) {
	for _, state := range []QueueState{QueueWaiting, QueueVerifying} {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
			t.Fatal(err)
		}
		if state == QueueVerifying {
			if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueVerifying, "", ""); err != nil {
				t.Fatal(err)
			}
		}
		if err := s.Move(bg, "a", c.Epoch, core.Position{Stage: core.StageReviewing}, "lane", "back out of the queue"); err != nil {
			t.Fatalf("%s: leaving MERGING: %v", state, err)
		}
		if err := s.Release(bg, "a", c.Epoch); err != nil {
			t.Fatal(err)
		}
		c2, err := s.Claim(bg, req("next", 77))
		if err != nil || c2.CardID != "a" || c2.Epoch != c.Epoch+1 {
			t.Fatalf("%s: claimed again: %+v %v", state, c2, err)
		}
		// the old row cannot be written by anyone
		if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueFailed, "", "x"); !errors.Is(err, ErrStaleEpoch) {
			t.Fatalf("%s: the old row is fenced: %v", state, err)
		}
		for _, e := range []Evidence{{Kind: KindGate, TreeSHA: "t2", SubjectSHA: "c2", Result: ResultPass}} {
			if err := s.AddEvidence(bg, "a", c2.Epoch, e); err != nil {
				t.Fatal(err)
			}
		}
		if err := s.Enqueue(bg, "a", c2.Epoch, "c2"); err != nil {
			t.Fatalf("%s: queued again after the row went stale: %v", state, err)
		}
		items, _ := s.Queue(bg)
		if len(items) != 1 || items[0].CardEpoch != c2.Epoch || items[0].CandidateSHA != "c2" || items[0].State != QueueWaiting {
			t.Errorf("%s: the row restarts at the new epoch: %+v", state, items)
		}
		notes, _ := s.NotesSince(bg, 0)
		found := false
		for _, n := range notes {
			found = found || (n.Kind == NoteDecision && n.CardID == "a" && strings.Contains(n.Text, "c1") && strings.Contains(n.Text, "c2") && strings.Contains(n.Text, string(state)))
		}
		if !found {
			t.Errorf("%s: no note keeps the replaced row: %+v", state, notes)
		}
	}
}

// A current-epoch row that is still open is not replaced.
func TestEnqueue_KeepsAnOpenRowOfTheCurrentEpoch(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err == nil {
		t.Error("a second Enqueue of an open current row must be refused")
	}
}

func TestQueueUpdate_UnknownEpochKindIsInvalid(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	for _, k := range []EpochKind{"", "epoch", "CARD"} {
		if err := s.QueueUpdate(bg, "a", QueueEpoch{Kind: k, N: c.Epoch}, QueueVerifying, "", ""); !errors.Is(err, ErrInvalid) {
			t.Errorf("kind %q: %v", k, err)
		}
	}
}

// A landed row is history: not even a higher card epoch replaces it.
func TestEnqueue_NeverReplacesALandedRow(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	landOne(t, s, "a", c.Epoch)
	for _, q := range []string{`UPDATE cards SET epoch=2 WHERE id='a'`, `UPDATE leases SET epoch=2 WHERE card_id='a'`} {
		if _, err := s.db.ExecContext(bg, q); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.AddEvidence(bg, "a", 2, Evidence{Kind: KindGate, TreeSHA: "t9", SubjectSHA: "c9", Result: ResultPass}); err != nil {
		t.Fatal(err)
	}
	if err := s.Enqueue(bg, "a", 2, "c9"); !errors.Is(err, ErrInvalid) {
		t.Errorf("a landed row is not replaced: %v", err)
	}
	if items, _ := s.Queue(bg); len(items) != 1 || items[0].CandidateSHA != "c1" || items[0].State != QueueLanded {
		t.Errorf("%+v", items)
	}
}
