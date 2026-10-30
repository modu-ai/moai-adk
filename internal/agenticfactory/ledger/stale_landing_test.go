package ledger

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// staleLanding builds the case the leader decided on: a lander holds the merge lock and has
// taken the row to landing at card epoch 1; the card's holder backs out of MERGING, lets go,
// and the card is claimed again at epoch 2. The row is still landing, at the older epoch.
func staleLanding(t *testing.T) (s *Store, epoch2, lock int64) {
	t.Helper()
	s = openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	var err error
	if lock, err = s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq"}, LandingLockTTL); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueLanding, "t0", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(bg, "a", c.Epoch, core.Position{Stage: core.StageReviewing}, "lane", "back out of the queue"); err != nil {
		t.Fatal(err)
	}
	if err := s.Release(bg, "a", c.Epoch); err != nil {
		t.Fatal(err)
	}
	c2, err := s.Claim(bg, req("next", 77))
	if err != nil || c2.Epoch != c.Epoch+1 {
		t.Fatalf("claimed again: %+v %v", c2, err)
	}
	if err := s.AddEvidence(bg, "a", c2.Epoch, Evidence{Kind: KindGate, TreeSHA: "t2", SubjectSHA: "c2", Result: ResultPass}); err != nil {
		t.Fatal(err)
	}
	return s, c2.Epoch, lock
}

func rowOf(t *testing.T, s *Store) QueueItem {
	t.Helper()
	items, _ := s.Queue(bg)
	if len(items) != 1 {
		t.Fatalf("%+v", items)
	}
	return items[0]
}

// A landing row may have a lander working under the merge lock: it is not restarted.
func TestEnqueue_RefusesARowStillLandingAtAnOlderEpoch(t *testing.T) {
	s, e2, _ := staleLanding(t)
	if err := s.Enqueue(bg, "a", e2, "c2"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("the landing row waits for its settlement: %v", err)
	}
	if it := rowOf(t, s); it.State != QueueLanding || it.CandidateSHA != "c1" || it.CardEpoch != 1 {
		t.Errorf("the row was touched: %+v", it)
	}
}

// The live merge-lock holder settles the older landing row from the trunk ancestry result;
// the row keeps its epoch; a failed settlement lets the card be queued again.
func TestQueueUpdate_LockHolderSettlesAnOlderLandingRow(t *testing.T) {
	for _, to := range []QueueState{QueueFailed, QueueLanded} {
		s, e2, lock := staleLanding(t)
		reason := ""
		if to == QueueFailed {
			reason = "candidate c1 is not an ancestor of trunk"
		}
		if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), to, "", reason); err != nil {
			t.Fatalf("%s: %v", to, err)
		}
		if it := rowOf(t, s); it.State != to || it.CardEpoch != 1 || it.CandidateSHA != "c1" {
			t.Errorf("%s: %+v", to, it)
		}
		notes, _ := s.NotesSince(bg, 0)
		found := false
		for _, n := range notes {
			found = found || (n.Kind == NoteDecision && n.CardID == "a" && strings.Contains(n.Text, "settled") && strings.Contains(n.Text, string(to)))
		}
		if !found {
			t.Errorf("%s: no note records the settlement: %+v", to, notes)
		}
		err := s.Enqueue(bg, "a", e2, "c2")
		if to == QueueLanded {
			if !errors.Is(err, ErrInvalid) {
				t.Errorf("a landed row is final: %v", err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("queued again after the failed settlement: %v", err)
		}
		// and it lands once
		for _, step := range []struct {
			ep    QueueEpoch
			state QueueState
		}{{CardEpoch(e2), QueueVerifying}, {LockEpoch(lock, "c1"), QueueLanding}, {LockEpoch(lock, "c1"), QueueLanded}} {
			if err := s.QueueUpdate(bg, "a", step.ep, step.state, "t1", ""); err != nil {
				t.Fatalf("%s: %v", step.state, err)
			}
		}
		if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueLanded, "", ""); !errors.Is(err, ErrInvalid) {
			t.Errorf("landing twice: %v", err)
		}
	}
}

// Only the live lock holder, and only to landed or failed.
func TestQueueUpdate_OlderLandingRowSettlementNeedsTheLiveLock(t *testing.T) {
	s, e2, lock := staleLanding(t)
	for name, c := range map[string]struct {
		ep    QueueEpoch
		state QueueState
	}{
		"old card epoch":      {CardEpoch(1), QueueFailed},
		"current card epoch":  {CardEpoch(e2), QueueFailed},
		"a stale lock epoch":  {LockEpoch(lock+1, "c1"), QueueFailed},
		"card kind landed":    {CardEpoch(1), QueueLanded},
		"lock holder waiting": {LockEpoch(lock, "c1"), QueueWaiting},
		"lock holder verify":  {LockEpoch(lock, "c1"), QueueVerifying},
	} {
		if err := s.QueueUpdate(bg, "a", c.ep, c.state, "", "x"); !errors.Is(err, ErrStaleEpoch) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if err := s.ReleaseLock(bg, LockMerge, lock); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueFailed, "", "x"); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("a released lock settles nothing: %v", err)
	}
	if it := rowOf(t, s); it.State != QueueLanding {
		t.Errorf("a refused settlement changed the row: %+v", it)
	}
}
