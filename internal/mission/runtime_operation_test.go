package mission

import (
	"context"
	"errors"
	"testing"
)

type fakeRuntimeProbe struct {
	start, reconnect, replace, credential, identity bool
}

func (f fakeRuntimeProbe) StartSupported(context.Context) bool           { return f.start }
func (f fakeRuntimeProbe) ReconnectSupported(context.Context) bool       { return f.reconnect }
func (f fakeRuntimeProbe) ReplaceSupported(context.Context) bool         { return f.replace }
func (f fakeRuntimeProbe) CredentialValid(context.Context) bool          { return f.credential }
func (f fakeRuntimeProbe) ProcessIdentitySupported(context.Context) bool { return f.identity }

func TestMissionRuntimeCapabilityMatrix(t *testing.T) {
	ctx := context.Background()
	full := ProbeMissionRuntimeCapability(ctx, fakeRuntimeProbe{true, true, true, true, true})
	if full.Mode != RuntimeDurable || !full.CanReconnect || !full.CanReplace {
		t.Fatalf("full probe = %+v", full)
	}
	partial := ProbeMissionRuntimeCapability(ctx, fakeRuntimeProbe{start: true, credential: true, identity: true})
	if partial.Mode != RuntimeActiveSessionOnly || partial.CanReconnect || partial.CanReplace {
		t.Fatalf("partial probe = %+v", partial)
	}
	unsupported := ProbeMissionRuntimeCapability(ctx, fakeRuntimeProbe{})
	if unsupported.Mode != RuntimeActiveSessionOnly {
		t.Fatalf("unsupported probe = %+v", unsupported)
	}
	expired := ProbeMissionRuntimeCapability(ctx, fakeRuntimeProbe{true, true, true, false, true})
	if expired.Mode != RuntimeActiveSessionOnly || expired.CredentialValid {
		t.Fatalf("expired probe = %+v", expired)
	}
	if CanReplaceRuntimeOwner(full, false) || !CanReplaceRuntimeOwner(full, true) {
		t.Fatal("replacement did not require prior-owner effect stop")
	}
	lease := RuntimeLease{MissionID: "m", ContractHash: "c", SnapshotHash: "s", OwnerID: "owner-a", Version: 1}
	taken, err := TakeoverRuntimeLease(lease, "owner-b", true, "c", "s")
	if err != nil || taken.OwnerID != "owner-b" || taken.Version != 2 {
		t.Fatalf("takeover = %+v err=%v", taken, err)
	}
	if _, err := TakeoverRuntimeLease(lease, "owner-b", true, "different", "s"); err == nil {
		t.Fatal("snapshot-lineage mismatch takeover allowed")
	}
}

type fakeReadback struct {
	applied bool
	err     error
	calls   int
}

func (f *fakeReadback) OperationApplied(context.Context, OperationReceipt) (bool, error) {
	f.calls++
	return f.applied, f.err
}

type fakeInvoker struct {
	calls int
	ids   []string
}

func (f *fakeInvoker) Invoke(_ context.Context, r OperationReceipt) error {
	f.calls++
	f.ids = append(f.ids, r.OperationID)
	return nil
}

func TestReconcileMissionOperationCrashCuts(t *testing.T) {
	actions := []Action{ActionPublish, ActionDispatch, ActionCommit, ActionLocalMerge, ActionBatchPush, ActionReleaseBranch, ActionReleasePR, ActionMainMerge}
	for _, action := range actions {
		t.Run(string(action), func(t *testing.T) {
			receipt := OperationReceipt{OperationID: "stable-" + string(action), MissionID: "m", Action: action, State: ReceiptInvoked}
			readback := &fakeReadback{applied: true}
			invoker := &fakeInvoker{}
			got, err := ReconcileMissionOperation(context.Background(), receipt, readback, invoker)
			if err != nil || got.State != ReceiptReconciled || invoker.calls != 0 || readback.calls != 1 {
				t.Fatalf("already applied = %+v err=%v invokes=%d reads=%d", got, err, invoker.calls, readback.calls)
			}
			readback = &fakeReadback{applied: false}
			got, err = ReconcileMissionOperation(context.Background(), receipt, readback, invoker)
			if err != nil || got.State != ReceiptInvoked || invoker.calls != 1 || invoker.ids[0] != receipt.OperationID {
				t.Fatalf("confirmed absent retry = %+v err=%v invokes=%v", got, err, invoker.ids)
			}
			readback = &fakeReadback{err: errors.New("ambiguous")}
			before := invoker.calls
			if _, err := ReconcileMissionOperation(context.Background(), receipt, readback, invoker); err == nil || invoker.calls != before {
				t.Fatal("ambiguous readback retried side effect")
			}
		})
	}
}
