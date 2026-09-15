package mission

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

type gtdE2EPublishOwner struct {
	store  *kanban.BacklogStore
	itemID string
}

func (o gtdE2EPublishOwner) Readback(ctx context.Context, _ kanban.GTDOperation) (bool, error) {
	item, err := kanban.LoadGTDItem(ctx, o.store, o.itemID)
	return err == nil && item.CardID != "", err
}

func (o gtdE2EPublishOwner) Apply(ctx context.Context, _ kanban.GTDOperation) error {
	_, err := kanban.EngageGTDItem(ctx, o.store, kanban.EngageInput{ItemID: o.itemID, Authorized: true, EvidenceFresh: true, DependenciesReady: true, LaneAvailable: true, ResourcesAvailable: true})
	return err
}

type gtdE2EPickOwner struct {
	store  *kanban.BacklogStore
	cardID string
}

func (o gtdE2EPickOwner) Readback(_ context.Context, _ kanban.GTDOperation) (bool, error) {
	record, err := o.store.LoadPure()
	if err != nil {
		return false, err
	}
	for _, item := range record.Items {
		if item.ID == o.cardID {
			return item.State == kanban.BacklogStatePicked, nil
		}
	}
	return false, errors.New("card missing")
}

func (o gtdE2EPickOwner) Apply(_ context.Context, _ kanban.GTDOperation) error {
	return o.store.Mutate(func(record *kanban.BacklogRecord) error {
		for i := range record.Items {
			if record.Items[i].ID == o.cardID {
				if record.Items[i].State != kanban.BacklogStateQueued {
					return errors.New("card not queued")
				}
				record.Items[i].State = kanban.BacklogStatePicked
				return nil
			}
		}
		return errors.New("card missing")
	})
}

type gtdE2EDispatchOwner struct {
	root, runID, cardID, lane string
	crashAfterEffect          bool
}

func (o *gtdE2EDispatchOwner) Readback(_ context.Context, _ kanban.GTDOperation) (bool, error) {
	record, err := kanban.NewBacklogStore(kanban.BacklogPathForRoot(o.root)).LoadPure()
	if err != nil {
		return false, err
	}
	for _, assignment := range record.Runtime.Assignments {
		if assignment.RunID == o.runID && assignment.CardID == o.cardID && assignment.OwnerLabel == o.lane {
			return true, nil
		}
	}
	return false, nil
}

func (o *gtdE2EDispatchOwner) Apply(_ context.Context, _ kanban.GTDOperation) error {
	if err := kanban.RecordFactoryCardAssignment(o.root, o.runID, o.cardID, o.lane, ""); err != nil {
		return err
	}
	if o.crashAfterEffect {
		o.crashAfterEffect = false
		return errors.New("crash after dispatch effect")
	}
	return nil
}

func TestGTDAutonomyEndToEnd(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store := kanban.NewBacklogStore(kanban.BacklogPathForRoot(root))
	restartedStore := kanban.NewBacklogStore(kanban.BacklogPathForRoot(root))
	item, err := kanban.CaptureGTDItem(ctx, store, kanban.CaptureInput{Content: "implement approved card", Source: "user", SourceAllowed: true, Sensitivity: kanban.SensitivityPrivate, EventID: "e2e-1"})
	if err != nil {
		t.Fatal(err)
	}
	clarified, err := kanban.ClarifyGTDItem(ctx, store, kanban.ClarifyInput{ItemID: item.ItemID, Disposition: kanban.DispositionAction, DesiredOutcome: "merged", CompletionEvidence: "CI+landed", Authority: "queue,dispatch", SourceTrusted: true})
	if err != nil || !clarified.Publishable {
		t.Fatalf("clarify=%+v err=%v", clarified, err)
	}
	if _, err := kanban.OrganizeGTDItem(ctx, store, kanban.OrganizeInput{ItemID: item.ItemID, Class: kanban.ClassAction, Context: "computer"}); err != nil {
		t.Fatal(err)
	}
	contract := completeContract()
	contract.Scope = []string{"gtd:" + item.ItemID}
	contract.AllowedActions = []Action{ActionPublish}
	sealed, err := SealMissionContract(contract)
	if err != nil {
		t.Fatal(err)
	}
	evidence := map[string]string{"approval": "explicit", "clarified": "true", "fresh_snapshot": "1", "organized": "true"}
	snapshot := MissionSnapshot{MissionID: contract.MissionID, ContractHash: sealed.Hash, PolicyVersion: sealed.Version, SnapshotHash: "e2e-snapshot", EvidenceRevision: 1, Evidence: evidence, State: StateRunning}
	decision := Decision{DecisionID: "publish", MissionID: contract.MissionID, ContractHash: sealed.Hash, PolicyVersion: sealed.Version, SnapshotHash: snapshot.SnapshotHash, EvidenceRevision: 1, Action: ActionPublish, Targets: contract.Scope, RequiredEvidence: requiredEvidenceForAction(ActionPublish), Evidence: evidence, ExpiresAt: time.Now().Add(time.Minute)}
	if _, err := ValidateMissionDecision(sealed, snapshot, decision, time.Now()); err != nil {
		t.Fatalf("sealed publication denied: %v", err)
	}
	publish := kanban.GTDOperation{OperationID: "op-publish", MissionID: contract.MissionID, Action: "publish", Target: contract.Scope[0], ReceiptJSON: []byte(`{"status":"passed"}`), SnapshotHash: snapshot.SnapshotHash}
	publishResults := make(chan error, 2)
	var publishWG sync.WaitGroup
	for _, operationStore := range []*kanban.BacklogStore{store, restartedStore} {
		publishWG.Add(1)
		go func(operationStore *kanban.BacklogStore) {
			defer publishWG.Done()
			_, err := kanban.ExecuteGTDOperation(ctx, operationStore, publish, gtdE2EPublishOwner{store: operationStore, itemID: item.ItemID})
			publishResults <- err
		}(operationStore)
	}
	publishWG.Wait()
	close(publishResults)
	for err := range publishResults {
		if err != nil {
			t.Fatal(err)
		}
	}
	itemAfterPublish, err := kanban.LoadGTDItem(ctx, restartedStore, item.ItemID)
	if err != nil || itemAfterPublish.CardID == "" {
		t.Fatalf("published item=%+v err=%v", itemAfterPublish, err)
	}
	record, err := store.Load()
	if err != nil || len(record.Items) != 1 {
		t.Fatalf("concurrent publish created %d cards: %v", len(record.Items), err)
	}
	cardID := itemAfterPublish.CardID
	pick := kanban.GTDOperation{OperationID: "op-pick", MissionID: contract.MissionID, Action: "pick", Target: contract.Scope[0], ReceiptJSON: []byte(`{"status":"passed"}`), SnapshotHash: "pick-snapshot"}
	if _, err := kanban.ExecuteGTDOperation(ctx, restartedStore, pick, gtdE2EPickOwner{store: restartedStore, cardID: cardID}); err != nil {
		t.Fatal(err)
	}
	lease, err := kanban.AcquireSlotLease(root, kanban.SlotLeaseRequest{Resource: "lane-10", SessionID: contract.MissionID, MaxDuration: time.Minute})
	if err != nil || lease.SessionID != contract.MissionID {
		t.Fatalf("lease=%+v err=%v", lease, err)
	}
	dispatch := kanban.GTDOperation{OperationID: "op-dispatch", MissionID: contract.MissionID, Action: "dispatch", Target: contract.Scope[0], ReceiptJSON: []byte(`{"status":"passed"}`), SnapshotHash: "dispatch-snapshot"}
	dispatchOwner := &gtdE2EDispatchOwner{root: root, runID: "run-1", cardID: cardID, lane: "lane-10", crashAfterEffect: true}
	if _, err := kanban.ExecuteGTDOperation(ctx, store, dispatch, dispatchOwner); err == nil {
		t.Fatal("dispatch crash cut was not observed")
	}
	reconciled, err := kanban.ExecuteGTDOperation(ctx, restartedStore, dispatch, dispatchOwner)
	if err != nil || reconciled.State != kanban.GTDOperationReconciled {
		t.Fatalf("dispatch reconciliation=%+v err=%v", reconciled, err)
	}
	final, err := restartedStore.LoadPure()
	if err != nil || len(final.Items) != 1 || final.Items[0].State != kanban.BacklogStatePicked || len(final.Runtime.Assignments) != 1 {
		t.Fatalf("final=%+v err=%v", final, err)
	}
	if final.Runtime.Assignments[0].CardID != cardID || final.Runtime.Assignments[0].OwnerLabel != "lane-10" {
		t.Fatalf("assignment=%+v", final.Runtime.Assignments[0])
	}
}

func TestFinalizeOrRevokeMission(t *testing.T) {
	revoked := FinalizeOrRevokeMission(FinalizationInput{Current: StateRunning, Revoked: true, InFlightEffects: 1})
	if revoked.State != StateRevoking || revoked.AllowNewWork || !revoked.ReconcileInFlight {
		t.Fatalf("revoked=%+v", revoked)
	}
	complete := FinalizeOrRevokeMission(FinalizationInput{Current: StateRunning, RequiredEvidence: map[string]bool{"ci": true, "landed": true}, MainLandedAncestry: true})
	if complete.State != StateCompleted {
		t.Fatalf("complete=%+v", complete)
	}
	for _, weak := range []FinalizationInput{{Current: StateRunning, AgentIdle: true}, {Current: StateRunning, LeaseExpired: true}} {
		if got := FinalizeOrRevokeMission(weak); got.State == StateCompleted {
			t.Fatalf("weak completion=%+v", got)
		}
	}
}
