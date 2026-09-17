package mission

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/goal"
)

func TestAutoMissionStateRoundTripAndRefusals(t *testing.T) {
	root := t.TempDir()
	state := AutoMission{SessionID: "018f4f4a-7b7c-7a11-8f4d-111111111111", Text: "승인된 임무", MissionMode: ModeAuto, ProgressionMode: goal.ProgressionAutonomous, State: StateDraft}
	if err := SaveAutoMission(root, state); err != nil {
		t.Fatal(err)
	}
	got, err := LoadAutoMission(root, state.SessionID)
	if err != nil || got == nil || got.MissionMode != ModeAuto || got.ProgressionMode != goal.ProgressionAutonomous {
		t.Fatalf("round trip=%+v err=%v", got, err)
	}
	for _, path := range []string{filepath.Dir(autoMissionPath(root, state.SessionID)), autoMissionPath(root, state.SessionID)} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		want := os.FileMode(0o600)
		if info.IsDir() {
			want = 0o700
		}
		if info.Mode().Perm() != want {
			t.Fatalf("mode %s=%o want %o", path, info.Mode().Perm(), want)
		}
	}
	if missing, err := LoadAutoMission(root, "018f4f4a-7b7c-7a11-8f4d-444444444444"); err != nil || missing != nil {
		t.Fatalf("missing=%+v err=%v", missing, err)
	}
	for _, invalid := range []AutoMission{{}, {SessionID: "s", Text: "x"}, {SessionID: "s", MissionMode: ModeAuto}} {
		if err := SaveAutoMission(root, invalid); err == nil {
			t.Fatalf("invalid state saved: %+v", invalid)
		}
	}
	badID := "018f4f4a-7b7c-7a11-8f4d-222222222222"
	if err := os.WriteFile(autoMissionPath(root, badID), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAutoMission(root, badID); err == nil {
		t.Fatal("corrupt state accepted")
	}
}

func TestValidateAutoMissionIntegrityRejectsEveryLineageMutant(t *testing.T) {
	contract := completeContract()
	contract.MissionID = "018f4f4a-7b7c-7a11-8f4d-cccccccccccc"
	sealed, err := SealMissionContract(contract)
	if err != nil {
		t.Fatal(err)
	}
	valid := AutoMission{SessionID: contract.MissionID, Text: "x", MissionMode: ModeAuto, State: StateRunning, Contract: &contract, ContractHash: sealed.Hash, Snapshot: &MissionSnapshot{MissionID: contract.MissionID, ContractHash: sealed.Hash, PolicyVersion: sealed.Version, SnapshotHash: "snapshot"}, OperationIDs: []string{"op-0123456789abcdef0123456789abcdef"}}
	if err := ValidateAutoMissionIntegrity(&valid); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAutoMissionIntegrity(nil); err == nil {
		t.Fatal("nil mission accepted")
	}
	mutants := []func(*AutoMission){
		func(s *AutoMission) { s.SessionID = "bad" },
		func(s *AutoMission) { s.MissionMode = "other" },
		func(s *AutoMission) { s.ContractHash = "wrong" },
		func(s *AutoMission) { s.Contract.PolicyVersion = "wrong" },
		func(s *AutoMission) { s.Snapshot.MissionID = "other" },
		func(s *AutoMission) { s.Snapshot.ContractHash = "other" },
		func(s *AutoMission) { s.Snapshot.PolicyVersion = "other" },
		func(s *AutoMission) { s.Snapshot.SnapshotHash = "" },
		func(s *AutoMission) { s.OperationIDs = []string{"short"} },
		func(s *AutoMission) { s.OperationIDs = []string{"op-zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"} },
	}
	for i, mutate := range mutants {
		copyState := valid
		copyContract := contract
		copySnapshot := *valid.Snapshot
		copyState.Contract, copyState.Snapshot = &copyContract, &copySnapshot
		mutate(&copyState)
		if err := ValidateAutoMissionIntegrity(&copyState); err == nil {
			t.Fatalf("mutant %d accepted", i)
		}
	}
	for _, mutate := range []func(*AutoMission){
		func(s *AutoMission) { s.ContractHash = "x" },
		func(s *AutoMission) { s.Snapshot = &MissionSnapshot{} },
		func(s *AutoMission) { s.OperationIDs = []string{"op-0123456789abcdef0123456789abcdef"} },
	} {
		draft := AutoMission{SessionID: contract.MissionID, Text: "x", MissionMode: ModeAuto}
		mutate(&draft)
		if err := ValidateAutoMissionIntegrity(&draft); err == nil {
			t.Fatal("unsealed lineage accepted")
		}
	}
}

func TestAutoMissionRejectsTraversalAbsoluteAndSymlink(t *testing.T) {
	root := t.TempDir()
	for _, id := range []string{"../../escaped", "/tmp/escaped", "a/b", `a\\b`, ".", "not-a-uuid"} {
		state := AutoMission{SessionID: id, Text: "safe", MissionMode: ModeAuto, ProgressionMode: goal.ProgressionAutonomous, State: StateDraft}
		if err := SaveAutoMission(root, state); err == nil {
			t.Fatalf("unsafe session saved: %q", id)
		}
		if _, err := LoadAutoMission(root, id); err == nil {
			t.Fatalf("unsafe session loaded: %q", id)
		}
	}
	out := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai", "state"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(out, filepath.Join(root, ".moai", "state", "mission")); err != nil {
		t.Fatal(err)
	}
	id := "018f4f4a-7b7c-7a11-8f4d-333333333333"
	state := AutoMission{SessionID: id, Text: "safe", MissionMode: ModeAuto, ProgressionMode: goal.ProgressionAutonomous, State: StateDraft}
	if err := SaveAutoMission(root, state); err == nil {
		t.Fatal("symlink mission directory accepted")
	}
	if _, err := os.Stat(filepath.Join(out, id+".json")); !os.IsNotExist(err) {
		t.Fatalf("symlink escaped state exists: %v", err)
	}
}

func TestAutoMissionClearAndStateFileSymlinkRefusals(t *testing.T) {
	root := t.TempDir()
	id := "018f4f4a-7b7c-7a11-8f4d-888888888888"
	state := AutoMission{SessionID: id, Text: "x", MissionMode: ModeAuto, ProgressionMode: goal.ProgressionAutonomous, State: StateDraft}
	if err := SaveAutoMission(root, state); err != nil {
		t.Fatal(err)
	}
	if err := ClearAutoMission(root, id); err != nil {
		t.Fatal(err)
	}
	if got, err := LoadAutoMission(root, id); err != nil || got != nil {
		t.Fatalf("cleared=%+v err=%v", got, err)
	}
	if err := ClearAutoMission(root, id); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(out, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	path := autoMissionPath(root, id)
	if err := os.Symlink(out, path); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAutoMission(root, id); err == nil {
		t.Fatal("state symlink loaded")
	}
	if err := ClearAutoMission(root, id); err == nil {
		t.Fatal("state symlink cleared")
	}
	if err := SaveAutoMission(filepath.Join(root, "missing"), state); err == nil {
		t.Fatal("missing project root accepted")
	}
}

func TestMissionPolicyFailClosedBranches(t *testing.T) {
	sealed, err := SealMissionContract(completeContract())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	evidence := map[string]string{"approval": "explicit", "clarified": "true", "fresh_snapshot": "1", "organized": "true"}
	snapshot := MissionSnapshot{MissionID: "mission-1", ContractHash: sealed.Hash, PolicyVersion: sealed.Version, SnapshotHash: "snap", EvidenceRevision: 1, Evidence: evidence, State: StateRunning}
	base := Decision{DecisionID: "d", MissionID: "mission-1", ContractHash: sealed.Hash, PolicyVersion: sealed.Version, SnapshotHash: "snap", EvidenceRevision: 1, Action: ActionPublish, Targets: []string{"card:t1"}, RequiredEvidence: requiredEvidenceForAction(ActionPublish), Evidence: evidence, ExpiresAt: now.Add(time.Minute)}
	tests := []struct {
		name string
		snap MissionSnapshot
		dec  Decision
	}{
		{"not-running", func() MissionSnapshot { x := snapshot; x.State = StateDraft; return x }(), base},
		{"mission-mismatch", func() MissionSnapshot { x := snapshot; x.MissionID = "other"; return x }(), base},
		{"policy-mismatch", func() MissionSnapshot { x := snapshot; x.PolicyVersion = "old"; return x }(), base},
		{"expired", snapshot, func() Decision { x := base; x.ExpiresAt = now; return x }()},
		{"prohibited", snapshot, func() Decision { x := base; x.Action = ActionForcePush; return x }()},
		{"target-missing", snapshot, func() Decision { x := base; x.Targets = nil; return x }()},
		{"resource-limit", func() MissionSnapshot {
			x := snapshot
			x.OperationsUsed = sealed.Contract.ResourceLimits.MaxOperations
			return x
		}(), base},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ValidateMissionDecision(sealed, tc.snap, tc.dec, now); err == nil {
				t.Fatal("fail-closed branch allowed")
			}
		})
	}
}

type errorInvoker struct{}

func (errorInvoker) Invoke(context.Context, OperationReceipt) error {
	return errors.New("invoke failed")
}

func TestMissionRuntimeOperationAndGovernanceRefusals(t *testing.T) {
	if got := ProbeMissionRuntimeCapability(context.Background(), nil); got.Mode != RuntimeActiveSessionOnly {
		t.Fatalf("nil probe=%+v", got)
	}
	lease := RuntimeLease{ContractHash: "c", SnapshotHash: "s"}
	if _, err := TakeoverRuntimeLease(lease, "", true, "c", "s"); err == nil {
		t.Fatal("empty owner takeover allowed")
	}
	if got := SuppressAutoMissionQuestions(false, ""); got.Report != "outside_sealed_scope" {
		t.Fatalf("default blocker=%+v", got)
	}
	ctx := context.Background()
	if _, err := ReconcileMissionOperation(ctx, OperationReceipt{}, &fakeReadback{}, &fakeInvoker{}); err == nil {
		t.Fatal("empty receipt reconciled")
	}
	reconciled := OperationReceipt{OperationID: "op", State: ReceiptReconciled}
	if got, err := ReconcileMissionOperation(ctx, reconciled, &fakeReadback{}, &fakeInvoker{}); err != nil || got.State != ReceiptReconciled {
		t.Fatalf("reconciled=%+v err=%v", got, err)
	}
	if _, err := ReconcileMissionOperation(ctx, OperationReceipt{OperationID: "op"}, &fakeReadback{}, errorInvoker{}); err == nil {
		t.Fatal("invoke failure hidden")
	}
}

func TestMissionDispatchAndDeliveryRefusals(t *testing.T) {
	registry := NewDispatchRegistry()
	if _, err := registry.Dispatch(DispatchInput{}); err == nil {
		t.Fatal("invalid card dispatched")
	}
	base := DispatchInput{MissionID: "m", CardID: "t1", CardState: "picked", CardRevision: 1, ExpectedRevision: 1, Lane: "lane-1", LaneAvailable: true, LeaseHeld: true}
	blocked := base
	blocked.LaneAvailable = false
	if _, err := registry.Dispatch(blocked); err == nil {
		t.Fatal("unavailable lane dispatched")
	}
	if got, err := registry.Dispatch(base); err != nil || !got.Applied {
		t.Fatalf("valid dispatch=%+v err=%v", got, err)
	}
	if got, err := registry.Dispatch(base); err != nil || got.Applied {
		t.Fatalf("duplicate dispatch=%+v err=%v", got, err)
	}
	badPath := CardIntegrationInput{OwnerRole: "manager-git", LauncherEntered: true, WorktreeBranch: "WT-x", IntegrationWorktree: "/repo/.claude/worktrees/develop", IntegrationLeaseHeld: true, LocalDevelopBaseSHA: "a", CardHeadSHA: "b", LocalDevelopMergeSHA: "c", ExplicitPaths: []string{"../escape"}, MergeStrategy: "--no-ff"}
	if _, err := IntegrateCardIntoLocalDevelop(badPath); err == nil {
		t.Fatal("unsafe path accepted")
	}
	if got := FinalizeOrRevokeMission(FinalizationInput{}); got.State != StateRunning || !got.AllowNewWork {
		t.Fatalf("default finalization=%+v", got)
	}
}
