package cli

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// headlessOwner is the current process that owns the relocated headless
// thread's endpoint; it must be live for the rebind to accept it.
func headlessOwner(t *testing.T) laneHandoffBind {
	t.Helper()
	start := homestate.CurrentProcessFingerprint()
	if start == "" {
		t.Fatal("process-start identity of the test process is unavailable")
	}
	return laneHandoffBind{OwnerPID: os.Getpid(), OwnerProcessStart: start}
}

func (f *laneHandoffFixture) requireBound(t *testing.T, h factorymsg.Handoff, b factorymsg.HandoffBinding, thread string) {
	t.Helper()
	if st := f.storedHandoff(t, h.ID); st.State != factorymsg.HandoffBound {
		t.Fatalf("stored state = %s, want BOUND", st.State)
	}
	row := f.endpointRow(t)
	if row.Session != thread || row.Generation != f.source.Generation+1 {
		t.Fatalf("endpoint = %+v, want %s at generation %d", row, thread, f.source.Generation+1)
	}
	if b.New.SessionUUID != thread || b.Old.SessionUUID != f.source.SessionUUID || b.ReceiptID == "" {
		t.Fatalf("binding = %+v", b)
	}
	for q, want := range map[string]int{
		`SELECT count(*) FROM lane_endpoint_tombstones WHERE session_uuid='src-uuid'`: 1,
		`SELECT count(*) FROM lane_handoff_receipts`:                                  1,
		`SELECT count(*) FROM lane_dispatch_releases`:                                 1,
	} {
		if n := f.count(t, q); n != want {
			t.Fatalf("%s = %d, want %d", q, n, want)
		}
	}
}

// TestFactoryLaneHandoffHeadlessAppServerStateMachine is AC-FLH-004
// (REQ-FLH-002/007): a headless lane relocates through the official
// app-server and binds directly from the returned thread id plus the
// controller's own target readback — no SessionStart wait, no empty
// turn/start, no turn/steer.
func TestFactoryLaneHandoffHeadlessAppServerStateMachine(t *testing.T) {
	cases := []struct {
		name, sourceThread, newThread string
		wantMethods                   []string
	}{
		{"stored_history_forks", "thr-source", "thr-forked", []string{codexMethodInitialize, codexMethodThreadFork}},
		{"no_history_starts", "", "thr-fresh", []string{codexMethodInitialize, codexMethodThreadStart}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newLaneHandoffFixture(t, "develop", true)
			srv := withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: tc.newThread})
			h := f.wtReady(t, factorymsg.HandoffModeHeadless)
			pending, err := switchLaneHandoffHeadless(context.Background(), h, f.switchRequest(laneActivityIdle, tc.sourceThread), laneHandoffDeps{})
			if err != nil {
				t.Fatalf("switch: %v", err)
			}
			bind := headlessOwner(t)
			bind.ProjectRoot = f.primary
			b, err := bindLaneHandoffHeadless(context.Background(), pending, bind)
			if err != nil {
				t.Fatalf("bind: %v", err)
			}
			f.requireBound(t, h, b, tc.newThread)
			if !reflect.DeepEqual(srv.methods(), tc.wantMethods) {
				t.Fatalf("app-server requests = %v, want %v (no turn/start, turn/steer, or SessionStart wait)", srv.methods(), tc.wantMethods)
			}
			if srv.starts != 1 {
				t.Fatalf("app-server sessions = %d, want 1", srv.starts)
			}
		})
	}

	for _, activity := range []laneHandoffActivity{laneActivityActiveTurn, laneActivityPermissionWait, laneActivityInterrupting} {
		t.Run("rejects_"+string(activity), func(t *testing.T) {
			f := newLaneHandoffFixture(t, "develop", true)
			srv := withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: "thr-x"})
			h := f.wtReady(t, factorymsg.HandoffModeHeadless)
			before := f.endpointRow(t)
			_, err := switchLaneHandoffHeadless(context.Background(), h, f.switchRequest(activity, "thr-source"), laneHandoffDeps{})
			if _, ok := factorymsg.HandoffNackReason(err); !ok {
				t.Fatalf("non-idle switch err=%v, want NACK", err)
			}
			bind := headlessOwner(t)
			bind.ProjectRoot = f.primary
			if _, err := bindLaneHandoffHeadless(context.Background(), f.storedHandoff(t, h.ID), bind); err == nil {
				t.Fatal("NACKed handoff was bound")
			}
			if srv.starts != 0 {
				t.Fatalf("app-server sessions = %d for a non-idle lane", srv.starts)
			}
			f.requireNoEndpointEffects(t, before)
		})
	}

	// The controller reads the target back at bind time: a target that moved
	// after the relocation never reaches BOUND.
	t.Run("readback_mismatch_after_relocation", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: "thr-forked"})
		h := f.wtReady(t, factorymsg.HandoffModeHeadless)
		pending, err := switchLaneHandoffHeadless(context.Background(), h, f.switchRequest(laneActivityIdle, "thr-source"), laneHandoffDeps{})
		if err != nil {
			t.Fatal(err)
		}
		before := f.endpointRow(t)
		handoffGit(t, h.TargetPath, "commit", "-q", "--allow-empty", "-m", "moved after relocation")
		bind := headlessOwner(t)
		bind.ProjectRoot = f.primary
		_, err = bindLaneHandoffHeadless(context.Background(), pending, bind)
		requireHandoffNack(t, err, factorymsg.NackTargetReadbackMismatch)
		if st := f.storedHandoff(t, h.ID); st.State != factorymsg.HandoffNack {
			t.Fatalf("stored = %s, want NACK", st.State)
		}
		f.requireNoEndpointEffects(t, before)
	})

	// A target that cannot be read back at all never reaches BOUND.
	t.Run("target_unreadable_at_bind", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: "thr-forked"})
		h := f.wtReady(t, factorymsg.HandoffModeHeadless)
		pending, err := switchLaneHandoffHeadless(context.Background(), h, f.switchRequest(laneActivityIdle, "thr-source"), laneHandoffDeps{})
		if err != nil {
			t.Fatal(err)
		}
		before := f.endpointRow(t)
		if err := os.RemoveAll(h.TargetPath); err != nil {
			t.Fatal(err)
		}
		bind := headlessOwner(t)
		bind.ProjectRoot = f.primary
		_, err = bindLaneHandoffHeadless(context.Background(), pending, bind)
		requireHandoffNack(t, err, factorymsg.NackTargetReadbackMismatch)
		f.requireNoEndpointEffects(t, before)
	})

	// The headless binder refuses an interactive handoff outright.
	t.Run("interactive_handoff_refused", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		h := f.wtReady(t, factorymsg.HandoffModeInteractive)
		bind := headlessOwner(t)
		bind.ProjectRoot = f.primary
		if _, err := bindLaneHandoffHeadless(context.Background(), h, bind); err == nil {
			t.Fatal("interactive handoff took the headless binder")
		}
		if st := f.storedHandoff(t, h.ID); st.State != factorymsg.HandoffWTReady {
			t.Fatalf("stored = %s, want WT_READY untouched", st.State)
		}
	})

	// Headless never binds from a SessionStart-shaped claim: without recorded
	// official relocation evidence there is nothing to bind.
	t.Run("no_relocation_evidence", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		h := f.wtReady(t, factorymsg.HandoffModeHeadless)
		pending, err := f.store.MarkHandoffSwitchPendingHeadless(context.Background(), h)
		if err != nil {
			t.Fatal(err)
		}
		before := f.endpointRow(t)
		bind := headlessOwner(t)
		bind.ProjectRoot = f.primary
		_, err = bindLaneHandoffHeadless(context.Background(), pending, bind)
		requireHandoffNack(t, err, factorymsg.NackRelocationEvidenceInvalid)
		f.requireNoEndpointEffects(t, before)
	})
}
