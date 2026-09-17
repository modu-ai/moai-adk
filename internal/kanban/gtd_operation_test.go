package kanban

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
)

type failingGTDOperationOwner struct {
	readErr, applyErr error
	applied           bool
}

func (f failingGTDOperationOwner) Readback(context.Context, GTDOperation) (bool, error) {
	return f.applied, f.readErr
}
func (f failingGTDOperationOwner) Apply(context.Context, GTDOperation) error { return f.applyErr }

type fakeGTDOperationOwner struct {
	mu      sync.Mutex
	applied map[string]bool
	calls   int
}

func TestPersistentGTDOperationFailClosedBranches(t *testing.T) {
	ctx := context.Background()
	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	if _, err := ExecuteGTDOperation(ctx, store, GTDOperation{}, nil); err == nil {
		t.Fatal("invalid operation allowed")
	}
	op := GTDOperation{OperationID: "branches", MissionID: "m", Action: "publish", Target: "x", SnapshotHash: "s", ReceiptJSON: []byte(`{}`)}
	if _, err := ExecuteGTDOperation(ctx, store, op, failingGTDOperationOwner{readErr: errors.New("read")}); err == nil {
		t.Fatal("readback error hidden")
	}
	op.OperationID = "apply"
	if _, err := ExecuteGTDOperation(ctx, store, op, failingGTDOperationOwner{applyErr: errors.New("apply")}); err == nil {
		t.Fatal("apply error hidden")
	}
	if err := markGTDOperationState(ctx, store, "missing", GTDOperationPrepared, GTDOperationInvoking); err == nil {
		t.Fatal("missing transition allowed")
	}
	if _, err := LoadGTDOperation(ctx, store, "missing"); err == nil {
		t.Fatal("missing operation loaded")
	}
}

func (f *fakeGTDOperationOwner) Readback(_ context.Context, op GTDOperation) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.applied[op.OperationID], nil
}

func (f *fakeGTDOperationOwner) Apply(_ context.Context, op GTDOperation) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.applied[op.OperationID] = true
	return nil
}

func TestPersistentGTDOperationExactlyOnceAcrossStoresAndCrashCuts(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "backlog.json")
	stores := []*BacklogStore{NewBacklogStore(path), NewBacklogStore(path)}
	owner := &fakeGTDOperationOwner{applied: map[string]bool{}}
	op := GTDOperation{OperationID: "op-publish-1", MissionID: "mission-1", Action: "publish", Target: "gtd-1234567890abcdef", SnapshotHash: "snapshot-1", ReceiptJSON: []byte(`{"decision":"d1"}`)}
	var wg sync.WaitGroup
	results := make(chan GTDOperation, 2)
	for _, store := range stores {
		wg.Add(1)
		go func(s *BacklogStore) {
			defer wg.Done()
			got, _ := ExecuteGTDOperation(ctx, s, op, owner)
			results <- got
		}(store)
	}
	wg.Wait()
	close(results)
	if owner.calls != 1 {
		t.Fatalf("owner calls=%d want 1", owner.calls)
	}
	got, err := LoadGTDOperation(ctx, stores[0], op.OperationID)
	if err != nil || got.State != GTDOperationReconciled {
		t.Fatalf("operation=%+v err=%v", got, err)
	}

	before := GTDOperation{OperationID: "op-before", MissionID: "mission-1", Action: "pick", Target: "t1", SnapshotHash: "snapshot-1", ReceiptJSON: []byte(`{}`)}
	if _, _, err := PrepareGTDOperation(ctx, stores[0], before); err != nil {
		t.Fatal(err)
	}
	if _, err := ExecuteGTDOperation(ctx, stores[1], before, owner); err != nil {
		t.Fatal(err)
	}

	after := GTDOperation{OperationID: "op-after", MissionID: "mission-1", Action: "dispatch", Target: "t1", SnapshotHash: "snapshot-1", ReceiptJSON: []byte(`{}`)}
	if _, _, err := PrepareGTDOperation(ctx, stores[0], after); err != nil {
		t.Fatal(err)
	}
	if err := markGTDOperationState(ctx, stores[0], after.OperationID, GTDOperationPrepared, GTDOperationInvoking); err != nil {
		t.Fatal(err)
	}
	owner.applied[after.OperationID] = true
	calls := owner.calls
	if got, err := ExecuteGTDOperation(ctx, stores[1], after, owner); err != nil || got.State != GTDOperationReconciled {
		t.Fatalf("crash readback=%+v err=%v", got, err)
	}
	if owner.calls != calls {
		t.Fatalf("post-effect crash reinvoked owner: %d -> %d", calls, owner.calls)
	}
}

func TestPersistentGTDOperationRejectsIdentityCollision(t *testing.T) {
	ctx := context.Background()
	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	base := GTDOperation{OperationID: "same", MissionID: "m", Action: "publish", Target: "gtd-1234567890abcdef", SnapshotHash: "s", ReceiptJSON: []byte(`{}`)}
	if _, _, err := PrepareGTDOperation(ctx, store, base); err != nil {
		t.Fatal(err)
	}
	base.Target = "gtd-fedcba0987654321"
	if _, _, err := PrepareGTDOperation(ctx, store, base); err == nil {
		t.Fatal("operation identity collision accepted")
	}
}

func TestPersistentGTDOperationRequiresAuthoritativeReadback(t *testing.T) {
	ctx := context.Background()
	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	base := GTDOperation{OperationID: "readback-missing", MissionID: "mission", Action: "commit", Target: "repo", SnapshotHash: "snapshot", ReceiptJSON: []byte(`{"approved":true}`)}
	if _, err := ExecuteGTDOperation(ctx, store, base, failingGTDOperationOwner{}); err == nil || !errors.Is(err, context.Canceled) && err.Error() != "gtd operation: authoritative_readback_missing" {
		t.Fatalf("missing authoritative readback err=%v", err)
	}
	stored, err := LoadGTDOperation(ctx, store, base.OperationID)
	if err != nil || stored.State != GTDOperationInvoking {
		t.Fatalf("uncertain operation=%+v err=%v", stored, err)
	}
	owner := &fakeGTDOperationOwner{applied: map[string]bool{base.OperationID: true}}
	if got, err := ExecuteGTDOperation(ctx, store, base, owner); err != nil || got.State != GTDOperationReconciled || owner.calls != 0 {
		t.Fatalf("reconcile=%+v calls=%d err=%v", got, owner.calls, err)
	}
	if got, err := ExecuteGTDOperation(ctx, store, base, owner); err != nil || got.State != GTDOperationReconciled || owner.calls != 0 {
		t.Fatalf("already reconciled=%+v calls=%d err=%v", got, owner.calls, err)
	}
	if _, err := ExecuteGTDOperation(ctx, store, base, nil); err == nil {
		t.Fatal("nil owner accepted")
	}

	uncertain := base
	uncertain.OperationID = "uncertain"
	if _, _, err := PrepareGTDOperation(ctx, store, uncertain); err != nil {
		t.Fatal(err)
	}
	if err := markGTDOperationState(ctx, store, uncertain.OperationID, GTDOperationPrepared, GTDOperationInvoking); err != nil {
		t.Fatal(err)
	}
	notApplied := &fakeGTDOperationOwner{applied: map[string]bool{}}
	if got, err := ExecuteGTDOperation(ctx, store, uncertain, notApplied); err != nil || got.State != GTDOperationInvoking || notApplied.calls != 0 {
		t.Fatalf("uncertain retry=%+v calls=%d err=%v", got, notApplied.calls, err)
	}
}
