package factorymsg

import (
	"context"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func fixedProbe(start string, state homestate.ProcessIdentityState) func(int) (string, homestate.ProcessIdentityState) {
	return func(int) (string, homestate.ProcessIdentityState) { return start, state }
}

// TestAbandonLaneOwnerStates pins the three-state owner decision of the
// operator abandon store transaction: only a provably not-current owner is
// abandoned; current and unestablished owners refuse and write nothing.
func TestAbandonLaneOwnerStates(t *testing.T) {
	ctx := context.Background()
	recorded := "fake-source-start"
	cases := []struct {
		name  string
		probe func(int) (string, homestate.ProcessIdentityState)
		want  string // "" = abandoned
	}{
		{"live_same_start", fixedProbe(recorded, homestate.ProcessIdentityLive), NackSourceOwnerLive},
		{"live_unknown_start", fixedProbe("", homestate.ProcessIdentityLive), NackSourceOwnerLive},
		{"indeterminate", fixedProbe("", homestate.ProcessIdentityIndeterminate), NackSourceOwnerLive},
		{"live_reused_pid", fixedProbe("other-start", homestate.ProcessIdentityLive), ""},
		{"dead", fixedProbe("", homestate.ProcessIdentityDead), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newBindFixture(t, HandoffModeInteractive)
			before := f.counts(t)
			h, err := f.s.AbandonLane(ctx, bindSlot, tc.probe)
			if tc.want != "" {
				if got, _ := HandoffNackReason(err); got != tc.want {
					t.Fatalf("err=%v, want %s", err, tc.want)
				}
				if st, _ := f.handoffState(t); st != HandoffSwitchPendingInteractive {
					t.Fatalf("state = %s after a refusal", st)
				}
				return
			}
			if err != nil || h.State != HandoffAbandoned || h.Reason != AbandonOperator {
				t.Fatalf("abandon = %+v err=%v", h, err)
			}
			if st, reason := f.handoffState(t); st != HandoffAbandoned || reason != AbandonOperator {
				t.Fatalf("stored = %s/%s", st, reason)
			}
			if after := f.counts(t); after != before {
				t.Fatalf("abandon wrote BOUND facts: before=%+v after=%+v", before, after)
			}
		})
	}
}

func TestAbandonLaneRefusals(t *testing.T) {
	ctx := context.Background()
	f := newBindSeed(t)
	if _, err := f.s.AbandonLane(ctx, "bad slot!", fixedProbe("", homestate.ProcessIdentityDead)); err == nil {
		t.Fatal("invalid slot accepted")
	}
	if _, err := f.s.AbandonLane(ctx, bindSlot, nil); err == nil {
		t.Fatal("nil probe accepted")
	}
	if _, err := f.s.AbandonLane(ctx, bindSlot, fixedProbe("", homestate.ProcessIdentityDead)); err == nil {
		t.Fatal("abandon with no handoff accepted")
	} else if got, _ := HandoffNackReason(err); got != NackHandoffNotPending {
		t.Fatalf("err=%v, want %s", err, NackHandoffNotPending)
	}
}

func TestAbandonHandoffAndReceiptReadback(t *testing.T) {
	ctx := context.Background()
	f := newBindFixture(t, HandoffModeInteractive)
	if _, err := f.s.AbandonHandoff(ctx, f.h, ""); err == nil {
		t.Fatal("ABANDONED without a reason accepted")
	}
	if _, ok, err := f.s.HandoffReceipt(ctx, f.h.ID); err != nil || ok {
		t.Fatalf("receipt before BOUND ok=%v err=%v", ok, err)
	}
	b, err := f.s.BindHandoff(ctx, f.h, f.evidence())
	if err != nil {
		t.Fatal(err)
	}
	got, ok, err := f.s.HandoffReceipt(ctx, f.h.ID)
	if err != nil || !ok || got.ReceiptID != b.ReceiptID || got.New != b.New || got.Old != b.Old {
		t.Fatalf("receipt readback = %+v ok=%v err=%v, want %+v", got, ok, err, b)
	}
	if _, err := f.s.AbandonHandoff(ctx, f.h, NackTargetDirty); err == nil {
		t.Fatal("BOUND handoff abandoned")
	}

	g := newBindFixture(t, HandoffModeHeadless)
	h, err := g.s.AbandonHandoff(ctx, g.h, NackTargetDirty)
	if err != nil || h.State != HandoffAbandoned || h.Reason != NackTargetDirty {
		t.Fatalf("abandon = %+v err=%v", h, err)
	}
}
