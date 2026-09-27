package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/mission"
)

// Characterization of the two existing lead dispatch paths (t1239 M6), taken
// before the factory-record mirror lands: the queue runtime row and the
// command-visible result each path produces today. The mirror must leave both
// exactly as recorded here.

// fcRuntimeRow is the stable part of a queue runtime assignment row.
type fcRuntimeRow struct{ RunID, CardID, Owner, State, Event string }

func fcRuntimeRows(t *testing.T, store *kanban.BacklogStore) []fcRuntimeRow {
	t.Helper()
	record, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	var out []fcRuntimeRow
	for _, a := range record.Runtime.Assignments {
		out = append(out, fcRuntimeRow{a.RunID, a.CardID, a.OwnerLabel, a.ReportedState, a.EventKind})
	}
	return out
}

// runGTDCapture runs one `moai gtd` invocation and returns stdout and stderr.
func runGTDCapture(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := NewGTDCommand()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

// fcGTDDispatch walks one item through capture → clarify → organize → engage
// --pick --dispatch and returns the engage stdout and stderr and the published
// card id.
func fcGTDDispatch(t *testing.T, store *kanban.BacklogStore, content, event, lane, runID string) (string, string, string) {
	t.Helper()
	out, _, err := runGTDCapture(t, "capture", content, "--event", event, "--source", "user", "--sensitivity", "private", "--json")
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	var captured struct {
		ItemID string `json:"item_id"`
	}
	if err := json.Unmarshal([]byte(out), &captured); err != nil || captured.ItemID == "" {
		t.Fatalf("capture output=%q err=%v", out, err)
	}
	if _, _, err := runGTDCapture(t, "clarify", captured.ItemID, "--disposition", "action", "--outcome", "landed", "--evidence", "CI", "--authority", "queue,dispatch", "--trusted"); err != nil {
		t.Fatalf("clarify: %v", err)
	}
	if _, _, err := runGTDCapture(t, "organize", captured.ItemID, "--class", "action", "--context", "computer"); err != nil {
		t.Fatalf("organize: %v", err)
	}
	stdout, stderr, err := runGTDCapture(t, "engage", captured.ItemID, "--approve", "--fresh", "--dependencies-ready", "--lane", lane, "--resources", "--pick", "--dispatch", "--run-id", runID, "--json")
	if err != nil {
		t.Fatalf("engage: %v (stderr %s)", err, stderr)
	}
	item, err := kanban.LoadGTDItem(context.Background(), store, captured.ItemID)
	if err != nil {
		t.Fatalf("load gtd item: %v", err)
	}
	return stdout, stderr, item.CardID
}

// fcGoalDispatch drives one auto mission through publish, pick, and dispatch
// (the goal.go owner-adapter path) and returns the dispatch operation, the
// dispatched card id, and the supervise stderr.
func fcGoalDispatch(t *testing.T, root string, store *kanban.BacklogStore, content, event, session, lane, runID string) (kanban.GTDOperation, string, string) {
	t.Helper()
	ctx := context.Background()
	item, err := kanban.CaptureGTDItem(ctx, store, kanban.CaptureInput{Content: content, Source: "user", SourceAllowed: true, Sensitivity: kanban.SensitivityPrivate, EventID: event})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := kanban.ClarifyGTDItem(ctx, store, kanban.ClarifyInput{ItemID: item.ItemID, Disposition: kanban.DispositionAction, DesiredOutcome: "dispatched", CompletionEvidence: "assignment", Authority: "queue,dispatch", SourceTrusted: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := kanban.OrganizeGTDItem(ctx, store, kanban.OrganizeInput{ItemID: item.ItemID, Class: kanban.ClassAction}); err != nil {
		t.Fatal(err)
	}
	target := "gtd:" + item.ItemID
	var stderr bytes.Buffer
	run := func(args ...string) error {
		cmd := newGoalCmd()
		cmd.SetOut(io.Discard)
		cmd.SetErr(&stderr)
		cmd.SetArgs(args)
		return cmd.Execute()
	}
	if err := run("--auto", "--session", session, "publish pick dispatch"); err != nil {
		t.Fatal(err)
	}
	if err := run("approve", "--session", session, "--scope", target, "--action", "publish", "--action", "pick", "--action", "dispatch", "--completion-evidence", "assignment"); err != nil {
		t.Fatal(err)
	}
	if _, err := kanban.AcquireSlotLease(root, kanban.SlotLeaseRequest{Resource: lane, SessionID: session, MaxDuration: time.Hour}); err != nil {
		t.Fatal(err)
	}
	head := gitFixtureCLI(t, root, "rev-parse", "HEAD")
	dir := filepath.Join(root, ".moai", "state", "mission", "governance")
	writeReceipts := func(action mission.Action, revision int64) {
		t.Helper()
		state, err := mission.LoadAutoMission(root, session)
		if err != nil {
			t.Fatal(err)
		}
		snapshot := autoSnapshotHash(state.ContractHash, target, revision)
		base := mission.GovernanceReceipt{Version: 1, MissionID: session, ContractHash: state.ContractHash, SnapshotHash: snapshot, Action: action, Targets: []string{target}, ExpiresAt: time.Now().Add(time.Hour), HeadSHA: head}
		decision := base
		decision.Kind, decision.Issuer, decision.Status = mission.GovernanceDecision, "mission-governor", mission.GovernanceRecommended
		audit := base
		audit.Kind, audit.Issuer, audit.Status = mission.GovernanceAudit, "sync-auditor", mission.GovernancePassed
		if err := mission.WriteGovernanceReceipt(root, filepath.Join(dir, string(action)+"-decision.json"), decision); err != nil {
			t.Fatal(err)
		}
		if err := mission.WriteGovernanceReceipt(root, filepath.Join(dir, string(action)+"-audit.json"), audit); err != nil {
			t.Fatal(err)
		}
	}
	current, err := kanban.LoadGTDItem(ctx, store, item.ItemID)
	if err != nil {
		t.Fatal(err)
	}
	writeReceipts(mission.ActionPublish, current.SourceRevision)
	writeReceipts(mission.ActionPick, current.SourceRevision+1)
	writeReceipts(mission.ActionDispatch, current.SourceRevision+1)
	governorPattern := filepath.Join(dir, "{action}-decision.json")
	auditPattern := filepath.Join(dir, "{action}-audit.json")
	completionPath := filepath.Join(dir, "completion.json")
	// The first supervised run executes the three operations and then stops
	// at the missing completion receipt; the dispatch has happened by then.
	_ = run("run", "--supervise", "--session", session, "--target", target, "--governor-receipt", governorPattern, "--audit-receipt", auditPattern, "--completion-receipt", completionPath, "--lane", lane, "--run-id", runID)
	state, err := mission.LoadAutoMission(root, session)
	if err != nil {
		t.Fatal(err)
	}
	var dispatchOp kanban.GTDOperation
	for _, id := range state.OperationIDs {
		op, err := kanban.LoadGTDOperation(ctx, store, id)
		if err != nil {
			t.Fatal(err)
		}
		if op.Action == string(mission.ActionDispatch) {
			dispatchOp = op
		}
	}
	if dispatchOp.OperationID == "" {
		t.Fatalf("no dispatch operation recorded (ops %v, blocker %q)", state.OperationIDs, state.LastBlocker)
	}
	published, err := kanban.LoadGTDItem(ctx, store, item.ItemID)
	if err != nil {
		t.Fatal(err)
	}
	return dispatchOp, published.CardID, stderr.String()
}

// AC-024 (characterization) — the queue dispatch path today: one runtime row
// `picked` / `card.assigned` for the lane, and the engage result naming the card.
func TestFR_AC024_CharacterizeGTDDispatch(t *testing.T) {
	_, store := todoFixture(t)
	stdout, _, cardID := fcGTDDispatch(t, store, "characterize gtd dispatch", "fr-char-gtd", "worker-2", "run-gtd")
	if cardID != "t1" {
		t.Fatalf("published card = %q, want t1", cardID)
	}
	want := []fcRuntimeRow{{"run-gtd", "t1", "worker-2", "picked", "card.assigned"}}
	if got := fcRuntimeRows(t, store); len(got) != 1 || got[0] != want[0] {
		t.Fatalf("runtime rows = %+v, want %+v", got, want)
	}
	if strings.TrimSpace(stdout) != fcGTDEngageOutput {
		t.Fatalf("engage output = %q, want %q", strings.TrimSpace(stdout), fcGTDEngageOutput)
	}
}

// AC-025 (characterization) — the auto-mission dispatch path today: one
// runtime row for the lane and a dispatch operation in the applied state.
func TestFR_AC025_CharacterizeGoalDispatch(t *testing.T) {
	root, store := todoFixture(t)
	op, cardID, _ := fcGoalDispatch(t, root, store, "characterize goal dispatch", "fr-char-goal", "018f4f4a-7b7c-7a11-8f4d-f22222222222", "worker-3", "auto-run-2")
	want := []fcRuntimeRow{{"auto-run-2", cardID, "worker-3", "picked", "card.assigned"}}
	if got := fcRuntimeRows(t, store); len(got) != 1 || got[0] != want[0] {
		t.Fatalf("runtime rows = %+v, want %+v", got, want)
	}
	fcAssertGoalDispatchOp(t, op, "018f4f4a-7b7c-7a11-8f4d-f22222222222")
}

// fcAssertGoalDispatchOp checks the stable fields of the recorded dispatch
// operation receipt: reconciled, action dispatch, the mission, a gtd target.
func fcAssertGoalDispatchOp(t *testing.T, op kanban.GTDOperation, session string) {
	t.Helper()
	if op.State != "reconciled" || op.Action != "dispatch" || op.MissionID != session || !strings.HasPrefix(op.Target, "gtd:") {
		t.Fatalf("dispatch operation = state %s action %s mission %s target %s", op.State, op.Action, op.MissionID, op.Target)
	}
}

const fcGTDEngageOutput = `{"actionable":true,"card_id":"t1","reasons":null}`
