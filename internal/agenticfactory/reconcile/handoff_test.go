package reconcile

import (
	"context"
	"errors"
	"testing"
)

type fakeHandoff struct {
	rec      Receipt
	recFound bool
	recErr   error
	pr       RemotePR
	prErr    error
}

func (f *fakeHandoff) LookupHandoff(context.Context, string) (Receipt, bool, error) {
	return f.rec, f.recFound, f.recErr
}
func (f *fakeHandoff) RemotePR(context.Context, int) (RemotePR, error) { return f.pr, f.prErr }

func id() string { return HandoffID(7, 55, headA) }

func decide(f *fakeHandoff) HandoffDecision {
	return ReconcileHandoff(context.Background(), f, 7, id())
}

func TestHandoffIDIsDeterministic(t *testing.T) {
	first, second := HandoffID(7, 55, headA), HandoffID(7, 55, headA) // two calls, one input
	if first != second || first == HandoffID(7, 55, headB) {
		t.Error("handoff id must depend on issue, PR and head only")
	}
}

// HANDOFF01: PR published, acceptance lost -> not done, claim kept, no next dispatch.
func TestPublishedPRWithLostAcceptanceHoldsEverything(t *testing.T) {
	d := decide(&fakeHandoff{pr: RemotePR{Found: true, Number: 55, Head: headA}})
	if d.Kind != HandoffAcceptanceUnconfirmed || d.ReleaseClaim || d.NextDispatch || d.Done {
		t.Errorf("decision = %+v", d)
	}
	if d.ClosureOwner != OwnerAgent {
		t.Errorf("closure owner = %q before acceptance, want the original agent", d.ClosureOwner)
	}
}

// HANDOFF02: duplicate and out-of-order receipts release the claim exactly once.
func TestReleaseHappensOnceAcrossDuplicatesAndOrder(t *testing.T) {
	// Acceptance arrives before the PR is visible on the remote: hold.
	f := &fakeHandoff{rec: Receipt{HandoffID: id(), Issue: 7, PR: 55, Head: headA, LeaderAccepted: true}, recFound: true}
	if d := decide(f); d.ReleaseClaim || d.Kind != HandoffPRMissing {
		t.Fatalf("acceptance before PR visible: %+v", d)
	}
	f.pr = RemotePR{Found: true, Number: 55, Head: headA}
	first := decide(f)
	if !first.ReleaseClaim || !first.NextDispatch || first.ClosureOwner != OwnerLeader {
		t.Fatalf("first = %+v", first)
	}
	f.rec.ClaimReleased = true // the control service recorded the release
	for i := 0; i < 3; i++ {
		if again := decide(f); again.ReleaseClaim {
			t.Fatalf("pass %d released the claim a second time", i)
		}
	}
}

// HANDOFF03: crash before / after acceptance recovers from remote PR + receipt without a second writer.
func TestCrashAroundAcceptanceRecoversWithoutDuplicateWriter(t *testing.T) {
	before := decide(&fakeHandoff{pr: RemotePR{Found: true, Number: 55, Head: headA}, rec: Receipt{HandoffID: id(), Head: headA}, recFound: true})
	if before.ReleaseClaim || before.NewWriterAllowed || before.ClosureOwner != OwnerAgent {
		t.Errorf("before accept: %+v", before)
	}
	after := decide(&fakeHandoff{pr: RemotePR{Found: true, Number: 55, Head: headA}, rec: Receipt{HandoffID: id(), Head: headA, LeaderAccepted: true}, recFound: true})
	if !after.ReleaseClaim || after.NewWriterAllowed || after.ClosureOwner != OwnerLeader {
		t.Errorf("after accept, before release: %+v", after)
	}
}

// HANDOFF08: a HEAD that moved after the handoff voids the handoff.
func TestHeadChangedAfterHandoffInvalidatesIt(t *testing.T) {
	d := decide(&fakeHandoff{pr: RemotePR{Found: true, Number: 55, Head: headB}, rec: Receipt{HandoffID: id(), Head: headA, LeaderAccepted: true}, recFound: true})
	if d.Kind != HandoffStale || d.ReleaseClaim || d.NextDispatch || !d.NeedsNewHandoff {
		t.Errorf("decision = %+v", d)
	}
}

func TestUnreadableFactsAreUnconfirmedAndChangeNothing(t *testing.T) {
	for name, f := range map[string]*fakeHandoff{
		"receipt lookup": {recErr: errors.New("store down")},
		"remote lookup":  {prErr: errors.New("gh 502")},
	} {
		d := decide(f)
		if d.Kind != HandoffUnconfirmed || d.ReleaseClaim || d.NextDispatch || d.Done {
			t.Errorf("%s: %+v", name, d)
		}
	}
}

func TestMergedPRIsStillNotDoneByHandoffAlone(t *testing.T) {
	d := decide(&fakeHandoff{pr: RemotePR{Found: true, Number: 55, Head: headA, Merged: true}, rec: Receipt{HandoffID: id(), Head: headA, LeaderAccepted: true, ClaimReleased: true}, recFound: true})
	if d.Done {
		t.Error("handoff reconcile must never declare done; completion needs the done evidence")
	}
}
