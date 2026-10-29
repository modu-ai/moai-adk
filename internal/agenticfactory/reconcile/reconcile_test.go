package reconcile

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

const (
	headA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	headB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

type fakeProber struct {
	act   Actual
	err   error
	calls int
}

func (f *fakeProber) Probe(context.Context, Recorded) (Actual, error) {
	f.calls++
	return f.act, f.err
}

func kinds(r Report) []Kind {
	var ks []Kind
	for _, d := range r.Divergences {
		ks = append(ks, d.Kind)
	}
	return ks
}

func has(r Report, k Kind) bool {
	for _, d := range r.Divergences {
		if d.Kind == k {
			return true
		}
	}
	return false
}

func TestProbeRunsOnceAndUnavailableIsUnconfirmed(t *testing.T) {
	// Recorded state has mismatches a record-only decision would act on.
	rec := Recorded{Issue: 7, HeadSHA: headA, LeaseExpired: true, Delivery: effect.Executed}
	for _, p := range []*fakeProber{
		{err: errors.New("probe failed")},
		{act: Actual{Liveness: Unavailable}},
	} {
		rep := Reconcile(context.Background(), p, rec)
		if p.calls != 1 {
			t.Fatalf("probe calls = %d", p.calls)
		}
		if got := kinds(rep); len(got) != 1 || got[0] != KindUnconfirmed {
			t.Fatalf("kinds = %v, want only unconfirmed", got)
		}
		if rep.Divergences[0].Action != ActionNone {
			t.Errorf("action = %s, want none", rep.Divergences[0].Action)
		}
	}
}

func TestExpiredLeaseLiveOrIndeterminateNeverTakenOver(t *testing.T) {
	for _, l := range []Liveness{Live, Indeterminate} {
		rep := Reconcile(context.Background(), &fakeProber{act: Actual{Liveness: l, WorktreeHead: headA, Delivery: effect.Executed}},
			Recorded{Issue: 7, HeadSHA: headA, LeaseExpired: true, Delivery: effect.Executed})
		if !has(rep, KindLeaseLapsed) {
			t.Fatalf("%s: kinds = %v", l, kinds(rep))
		}
		for _, d := range rep.Divergences {
			if d.Action == ActionNewAttempt {
				t.Errorf("%s: new attempt offered for a holder that is not dead", l)
			}
		}
	}
}

func TestNewAttemptOnlyWhenDeadAndExpired(t *testing.T) {
	for _, l := range []Liveness{Live, Dead, Indeterminate} {
		for _, expired := range []bool{true, false} {
			rep := Reconcile(context.Background(), &fakeProber{act: Actual{Liveness: l, WorktreeHead: headA, Delivery: effect.Executed}},
				Recorded{Issue: 7, HeadSHA: headA, LeaseExpired: expired, Delivery: effect.Executed})
			offered := false
			for _, d := range rep.Divergences {
				if d.Action == ActionNewAttempt {
					offered = true
				}
			}
			if want := l == Dead && expired; offered != want {
				t.Errorf("liveness=%s expired=%v: new attempt offered=%v want %v", l, expired, offered, want)
			}
		}
	}
}

func TestStaleHeadRecordedNotActedOn(t *testing.T) {
	rep := Reconcile(context.Background(), &fakeProber{act: Actual{Liveness: Live, WorktreeHead: headB, Delivery: effect.Executed}},
		Recorded{Issue: 7, HeadSHA: headA, Delivery: effect.Executed})
	if !has(rep, KindStaleHead) {
		t.Fatalf("kinds = %v", kinds(rep))
	}
	for _, d := range rep.Divergences {
		if d.Kind == KindStaleHead && d.Action != ActionNone {
			t.Errorf("stale head action = %s", d.Action)
		}
	}
}

func TestDeliveryMismatchBothDirections(t *testing.T) {
	behind := Reconcile(context.Background(), &fakeProber{act: Actual{Liveness: Live, WorktreeHead: headA, Delivery: effect.Executed}},
		Recorded{Issue: 7, HeadSHA: headA, Delivery: effect.Sent})
	if !has(behind, KindRecordBehind) {
		t.Errorf("record behind: kinds = %v", kinds(behind))
	}
	ahead := Reconcile(context.Background(), &fakeProber{act: Actual{Liveness: Live, WorktreeHead: headA, Delivery: effect.Received}},
		Recorded{Issue: 7, HeadSHA: headA, Delivery: effect.Completed})
	if !has(ahead, KindRecordAhead) {
		t.Errorf("record ahead: kinds = %v", kinds(ahead))
	}
}

func TestDuplicateDispatchRefusedWithCollision(t *testing.T) {
	rep := Reconcile(context.Background(), &fakeProber{act: Actual{Liveness: Live, WorktreeHead: headA, Delivery: effect.Executed, ActiveDispatchIDs: []string{"cmd-old"}}},
		Recorded{Issue: 7, HeadSHA: headA, Delivery: effect.Executed, PendingDispatch: "cmd-new"})
	var d *Divergence
	for i := range rep.Divergences {
		if rep.Divergences[i].Kind == KindDuplicateDispatch {
			d = &rep.Divergences[i]
		}
	}
	if d == nil {
		t.Fatalf("kinds = %v", kinds(rep))
	}
	if d.Action != ActionNone || !strings.Contains(d.Detail, "cmd-old") {
		t.Errorf("duplicate = %+v, want action none naming cmd-old", *d)
	}
}

func TestConsistentStateYieldsNoDivergence(t *testing.T) {
	rep := Reconcile(context.Background(), &fakeProber{act: Actual{Liveness: Live, WorktreeHead: headA, Delivery: effect.Executed}},
		Recorded{Issue: 7, HeadSHA: headA, Delivery: effect.Executed})
	if len(rep.Divergences) != 0 {
		t.Errorf("divergences = %v", kinds(rep))
	}
}

func TestDetailIsMasked(t *testing.T) {
	rep := Reconcile(context.Background(), &fakeProber{act: Actual{Liveness: Live, WorktreeHead: headA, Delivery: effect.Executed, ActiveDispatchIDs: []string{"ghp_abcdefghijklmnopqrstuvwxyz0123456789"}}},
		Recorded{Issue: 7, HeadSHA: headA, Delivery: effect.Executed, PendingDispatch: "cmd-new"})
	for _, d := range rep.Divergences {
		if strings.Contains(d.Detail, "ghp_abcdefghijklmnopqrstuvwxyz0123456789") {
			t.Errorf("secret in detail: %q", d.Detail)
		}
	}
}

// The package must offer no action that kills a process or re-sends a command.
func TestNoDestructiveActionExists(t *testing.T) {
	for _, a := range []Action{ActionNone, ActionMarkLapsed, ActionNewAttempt} {
		s := strings.ToLower(string(a))
		if strings.Contains(s, "kill") || strings.Contains(s, "resend") || strings.Contains(s, "redispatch") {
			t.Errorf("action %q is destructive", a)
		}
	}
}
