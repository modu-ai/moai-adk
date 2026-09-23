package factorymsg

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestRegisterPeerOnBrokerWithoutHandoffTables: the hook hot path can open a
// broker created before the handoff tables existed, without schema setup. The
// REQ-FLH-017/018 reads inside RegisterPeer then find no handoff and no
// tombstone, and both registration kinds follow t1074 unchanged.
func TestRegisterPeerOnBrokerWithoutHandoffTables(t *testing.T) {
	f := newBindSeed(t)
	for _, table := range []string{"lane_handoffs", "lane_endpoint_tombstones"} {
		if _, err := f.s.db.Exec("DROP TABLE " + table); err != nil {
			t.Fatal(err)
		}
	}
	legacy, err := OpenExistingWithDeadline(f.root, f.run, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = legacy.Close() })
	ctx := context.Background()

	turn := f.source
	turn.SessionUUID = "post-cd-uuid"
	turn.ProcessStart = f.ownerStart
	got, err := legacy.RegisterPeer(ctx, turn)
	if err != nil || got.Generation != f.source.Generation+1 {
		t.Fatalf("turn registration on a pre-handoff broker = %+v err=%v", got, err)
	}

	relaunch := f.source
	relaunch.SessionUUID = ""
	relaunch.ProcessStart = "fake-relaunch-start"
	if _, err := legacy.RegisterLaunchPending(ctx, relaunch); err == nil {
		t.Fatal("launcher registration displaced a live owner on a pre-handoff broker")
	}
	relaunch.Slot = "lane-2"
	if _, err := legacy.RegisterLaunchPending(ctx, relaunch); err != nil {
		t.Fatalf("launcher registration on a pre-handoff broker: %v", err)
	}
}

// TestTurnRegistrationDuringHandoffIsPendingForAnyIdentity: REQ-FLH-018
// rejects every turn-hook registration on a lane with a non-final handoff with
// ENDPOINT_HANDOFF_PENDING — also one whose owner identity differs from a live
// source owner, which t1074 alone would refuse only as "live owner".
func TestTurnRegistrationDuringHandoffIsPendingForAnyIdentity(t *testing.T) {
	f := newRaceFixture(t, false)
	f.toSwitchPending(t)
	before := readEndpointRow(t, f.seed.db, raceSlot)
	other := f.postCD("post-cd-uuid")
	other.ProcessStart = "another-process-start"
	_, err := f.seed.RegisterPeer(context.Background(), other)
	requireNackReason(t, err, NackEndpointHandoffPending)
	if after := readEndpointRow(t, f.seed.db, raceSlot); after != before {
		t.Fatalf("refused registration wrote the row: before=%+v after=%+v", before, after)
	}
}

// TestHandoffStepHookAbortsItsTransaction: an error returned by the step
// observer rolls back the whole write transaction it fires in — registration,
// launcher finalize, reservation, and rebind alike — so a race test's barrier
// can never leave a partial write behind.
func TestHandoffStepHookAbortsItsTransaction(t *testing.T) {
	injected := errors.New("injected at step")
	failAt := func(name string) context.Context {
		return WithStepHook(context.Background(), func(step string) error {
			if step == name {
				return injected
			}
			return nil
		})
	}

	t.Run(StepRegisterHandoffRead, func(t *testing.T) {
		f := newBindSeed(t)
		before := readEndpointRow(t, f.s.db, bindSlot)
		p := f.source
		p.SessionUUID = "post-cd-uuid"
		p.ProcessStart = f.ownerStart
		if _, err := f.s.RegisterPeer(failAt(StepRegisterHandoffRead), p); !errors.Is(err, injected) {
			t.Fatalf("err=%v, want the injected abort", err)
		}
		if after := readEndpointRow(t, f.s.db, bindSlot); after != before {
			t.Fatalf("aborted registration wrote the row: before=%+v after=%+v", before, after)
		}
	})

	t.Run(StepRegisterFinalize, func(t *testing.T) {
		f := newBindFixture(t, HandoffModeInteractive)
		before := readEndpointRow(t, f.s.db, bindSlot)
		p := f.source
		p.SessionUUID = ""
		p.ProcessStart = f.ownerStart
		if _, err := f.s.RegisterLaunchPending(failAt(StepRegisterFinalize), p); !errors.Is(err, injected) {
			t.Fatalf("err=%v, want the injected abort", err)
		}
		if after := readEndpointRow(t, f.s.db, bindSlot); after != before {
			t.Fatalf("aborted launcher registration wrote the row: before=%+v after=%+v", before, after)
		}
		if st, _ := f.handoffState(t); st != HandoffSwitchPendingInteractive {
			t.Fatalf("aborted launcher registration moved the handoff to %s", st)
		}
	})

	t.Run(StepReserveInserted, func(t *testing.T) {
		f := newBindSeed(t)
		if _, err := f.s.ReserveHandoff(failAt(StepReserveInserted), validReservation(f.root)); !errors.Is(err, injected) {
			t.Fatalf("err=%v, want the injected abort", err)
		}
		if n := countLaneHandoffs(t, f.s.db, bindSlot); n != 0 {
			t.Fatalf("aborted reservation left %d handoff rows", n)
		}
	})

	t.Run(StepBindBegun, func(t *testing.T) {
		f := newBindFixture(t, HandoffModeInteractive)
		if _, err := f.s.BindHandoff(failAt(StepBindBegun), f.h, f.evidence()); !errors.Is(err, injected) {
			t.Fatalf("err=%v, want the injected abort", err)
		}
		if got := f.counts(t); got != (bindCounts{}) {
			t.Fatalf("aborted rebind wrote %+v", got)
		}
	})
}

// TestRacerHandleProbes pins the two probes the hook-package race test uses to
// prove its racers hold distinct broker handles.
func TestRacerHandleProbes(t *testing.T) {
	f := newBindSeed(t)
	other, err := Open(f.root, f.run)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = other.Close() })
	if !SharesHandle(f.s, f.s) {
		t.Fatal("a handle does not share itself")
	}
	if SharesHandle(f.s, other) {
		t.Fatal("two production Open handles reported as shared")
	}
	if !SharesHandle(f.s, &Store{db: f.s.db}) {
		t.Fatal("two Stores over one *sql.DB reported as distinct")
	}
	if st := f.s.HandleStats(); st.MaxOpenConnections != 1 {
		t.Fatalf("handle stats = %+v, want the single-connection pool", st)
	}
}
