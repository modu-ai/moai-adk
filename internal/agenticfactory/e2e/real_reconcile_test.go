package e2e

import (
	"context"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/reconcile"
)

// Scenarios 2-4 (real, lane-20 reconcile): recovery decisions come from the
// probed actual state, never from the record alone.

type fakeProber struct {
	act reconcile.Actual
	err error
}

func (f fakeProber) Probe(context.Context, reconcile.Recorded) (reconcile.Actual, error) {
	return f.act, f.err
}

func kinds(r reconcile.Report) map[reconcile.Kind]reconcile.Action {
	m := map[reconcile.Kind]reconcile.Action{}
	for _, d := range r.Divergences {
		m[d.Kind] = d.Action
	}
	return m
}

// Scenario 4: only a dead holder with an expired lease makes a new attempt eligible.
func TestReal_InterruptionRecovery_OnlyDeadHolderEligible(t *testing.T) {
	rec := reconcile.Recorded{Issue: 1948, LeaseExpired: true, Delivery: effect.Executed}
	cases := []struct {
		name string
		live reconcile.Liveness
		want reconcile.Kind
		act  reconcile.Action
	}{
		{"dead", reconcile.Dead, reconcile.KindLeaseDead, reconcile.ActionNewAttempt},
		{"live", reconcile.Live, reconcile.KindLeaseLapsed, reconcile.ActionMarkLapsed},
		{"indeterminate", reconcile.Indeterminate, reconcile.KindLeaseLapsed, reconcile.ActionMarkLapsed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := fakeProber{act: reconcile.Actual{Liveness: tc.live, Delivery: effect.Executed}}
			got := kinds(reconcile.Reconcile(context.Background(), p, rec))
			if got[tc.want] != tc.act {
				t.Errorf("%s: got %v, want %s=%s", tc.name, got, tc.want, tc.act)
			}
			for _, a := range got {
				if a == reconcile.ActionNewAttempt && tc.live != reconcile.Dead {
					t.Errorf("%s holder made a new attempt eligible", tc.name)
				}
			}
		})
	}
}

// Scenario 4/3: a failed or undecidable probe yields only "unconfirmed".
func TestReal_UnreadableActual_OnlyUnconfirmed(t *testing.T) {
	rec := reconcile.Recorded{Issue: 1948, LeaseExpired: true, Delivery: effect.Executed}
	for name, p := range map[string]fakeProber{
		"error":       {err: errors.New("probe down")},
		"unavailable": {act: reconcile.Actual{Liveness: reconcile.Unavailable}},
	} {
		t.Run(name, func(t *testing.T) {
			rep := reconcile.Reconcile(context.Background(), p, rec)
			if len(rep.Divergences) != 1 || rep.Divergences[0].Kind != reconcile.KindUnconfirmed ||
				rep.Divergences[0].Action != reconcile.ActionNone {
				t.Errorf("got %+v, want a single unconfirmed/none", rep.Divergences)
			}
		})
	}
}

// Scenario 3: the record is ahead of what the receiver shows -> advice is to
// query first; no action may resend or kill.
func TestReal_RecordAhead_NoResendAction(t *testing.T) {
	rec := reconcile.Recorded{Issue: 1948, Delivery: effect.Received}
	p := fakeProber{act: reconcile.Actual{Liveness: reconcile.Live, Delivery: effect.Sent}}
	rep := reconcile.Reconcile(context.Background(), p, rec)
	if kinds(rep)[reconcile.KindRecordAhead] != reconcile.ActionNone {
		t.Errorf("record-ahead action = %v, want none", kinds(rep))
	}
}

// Scenario 2: a pending dispatch colliding with an active one is flagged.
func TestReal_DuplicateDispatch_Flagged(t *testing.T) {
	rec := reconcile.Recorded{Issue: 1948, PendingDispatch: "d-2", Delivery: effect.Sent}
	p := fakeProber{act: reconcile.Actual{Liveness: reconcile.Live, Delivery: effect.Sent, ActiveDispatchIDs: []string{"d-1"}}}
	if _, ok := kinds(reconcile.Reconcile(context.Background(), p, rec))[reconcile.KindDuplicateDispatch]; !ok {
		t.Error("colliding dispatch not flagged")
	}
}
