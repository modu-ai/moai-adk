package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// The leader-approval receipt gate on the backlog completion surfaces
// (SPEC-FACTORY-COMPLETION-RECOVERY-001 M1, REQ-FCR-001..005). A
// factory-linked card — one with a factory card row — completes only with a
// receipt bound to the four values (card uuid, run id, factory version,
// evidence hash) and issued off the performing owner; a card with no factory
// row keeps the receipt-less completion (REQ-FCR-002's scope sentence).

// fcLinkedCard seeds a live backlog card and mints its runtime identity the
// way a real dispatch does (the runtime record write ensures identities), so
// the card reads back with the projected CardUUID the gate binds against.
// Returns the projected uuid.
func fcLinkedCard(t *testing.T, root string, store *factory.BacklogStore, id, text string) string {
	t.Helper()
	if _, _, err := runTodo(t, "add", text); err != nil {
		t.Fatalf("todo add: %v", err)
	}
	if err := factory.RecordFactoryCardState(root, fcRun, id, "worker-1", "", "picked", "card.assigned"); err != nil {
		t.Fatalf("record runtime assignment: %v", err)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for i := range rec.Items {
		if rec.Items[i].ID == id {
			if rec.Items[i].CardUUID == nil {
				t.Fatalf("card %s carries no projected identity", id)
			}
			return *rec.Items[i].CardUUID
		}
	}
	t.Fatalf("card %s not found after seeding", id)
	return ""
}

// Regression pin for round-4 continuation P1 (card t1538): EVERY assignment
// path updates the dispatch binding — including a reassignment into a run
// that already holds a row for the card. The old run's receipt then cannot
// close the card's new work.
func TestFactoryAssignUpdatesDispatchBinding(t *testing.T) {
	root, store := fcFixture(t)
	if _, _, err := runTodo(t, "add", "reassigned factory card"); err != nil {
		t.Fatalf("todo add: %v", err)
	}
	if err := store.Mutate(func(r *factory.BacklogRecord) error {
		for i := range r.Items {
			if r.Items[i].ID == "t1" {
				r.Items[i].State = factory.BacklogStatePicked
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	// First dispatch: run-old is created and becomes the binding.
	if _, _, err := runFactory(t, "assign", "t1", "--to", "worker-1", "--run", "run-old"); err != nil {
		t.Fatalf("assign run-old: %v", err)
	}
	// Reassignment into run-cli, which ALREADY holds a picked row — the
	// pre-fix mirror skipped RecordPicked here and left the binding stale.
	fcPlace(t, root, homestate.Card{CardID: "t1", RunID: fcRun, State: homestate.CardPicked, OwnerLabel: "worker-2", Version: 1})
	if _, _, err := runFactory(t, "assign", "t1", "--to", "worker-2", "--run", fcRun); err != nil {
		t.Fatalf("assign into the existing run-cli row: %v", err)
	}
	db := fcOpen(t, root)
	row, linked, err := db.RecordedCardRowReadonly(context.Background(), "t1")
	if err != nil || !linked {
		t.Fatalf("binding after reassignment: linked=%v err=%v, want run-cli", linked, err)
	}
	_ = db.Close()
	if row.RunID != fcRun {
		t.Fatalf("binding run = %s, want %s", row.RunID, fcRun)
	}

	// Completion scoping follows the binding: the old run's receipt is
	// refused; the recorded run's receipt closes.
	uuid := recheckUUID(t, root, store, "t1")
	fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
		CardUUID: uuid, RunID: "run-old", CardID: "t1", FactoryVersion: 2,
		EvidenceHash: "", Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
	})
	if _, _, err := runTodo(t, "done", "t1"); err == nil {
		t.Fatal("done closed on the old run's receipt after reassignment")
	}
	if !fcLiveItem(t, store, "t1") {
		t.Fatal("the refused done archived the card")
	}
	// The assigned row's evidence is empty until a commit lands; approve
	// binds the current row as it stands.
	if _, _, err := runFactory(t, "approve", "t1", "--run", fcRun, "--issuer", "lead"); err != nil {
		t.Fatalf("approve the recorded run: %v", err)
	}
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done with the recorded run's receipt: %v", err)
	}
}

// Regression pin for round-4 review P1-1/P1-2 (card t1538): all four
// assignment shapes keep the dispatch binding aligned — new-row assignment,
// reassignment into an existing target row, idempotent same-lane re-assign
// over a stale binding, and legacy-row recovery at assigned+. The shapes
// drive the dispatch mirror directly, the same writer every dispatch path
// shares.
func TestDispatchBindingCoversAssignmentShapes(t *testing.T) {
	root, store := fcFixture(t)
	assign := func(card, run, lane string) error {
		return writeFactoryAssignment(context.Background(), root, store, run, card, lane)
	}
	pick := func(card string) {
		if err := store.Mutate(func(r *factory.BacklogRecord) error {
			for i := range r.Items {
				if r.Items[i].ID == card {
					r.Items[i].State = factory.BacklogStatePicked
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	bindingRun := func(card string) string {
		db := fcOpen(t, root)
		defer func() { _ = db.Close() }()
		row, linked, err := db.RecordedCardRowReadonly(context.Background(), card)
		if err != nil || !linked {
			t.Fatalf("%s binding: linked=%v err=%v, want linked", card, linked, err)
		}
		return row.RunID
	}

	// Shape 1 — new-row assignment records the binding.
	shapeNew := addShapeCard(t, "shape new row")
	pick(shapeNew)
	if err := assign(shapeNew, "run-old", "worker-1"); err != nil {
		t.Fatalf("new-row assign: %v", err)
	}
	if got := bindingRun(shapeNew); got != "run-old" {
		t.Fatalf("new-row binding = %s, want run-old", got)
	}

	// Shape 3 — idempotent same-lane re-assign repairs a stale binding.
	fcBindDispatch(t, root, shapeNew, "run-stale")
	if err := assign(shapeNew, "run-old", "worker-1"); err != nil {
		t.Fatalf("idempotent assign: %v", err)
	}
	if got := bindingRun(shapeNew); got != "run-old" {
		t.Fatalf("idempotent binding = %s, want run-old", got)
	}

	// Shape 4 — legacy row at assigned+ with no binding recovers through
	// the idempotent path, then completes with a matching receipt.
	shapeLegacy := addShapeCard(t, "shape legacy recovery")
	fcPlace(t, root, homestate.Card{CardID: shapeLegacy, RunID: "run-old", State: homestate.CardAssigned, OwnerLabel: "worker-1", Version: 1, UpdatedAt: "2026-09-26T01:00:00Z"})
	pick(shapeLegacy)
	if err := assign(shapeLegacy, "run-old", "worker-1"); err != nil {
		t.Fatalf("legacy recovery assign: %v", err)
	}
	if got := bindingRun(shapeLegacy); got != "run-old" {
		t.Fatalf("legacy binding = %s, want run-old", got)
	}
	if _, _, err := runFactory(t, "approve", shapeLegacy, "--run", "run-old", "--issuer", "lead"); err != nil {
		t.Fatalf("approve legacy row: %v", err)
	}
	if _, _, err := runTodo(t, "done", shapeLegacy); err != nil {
		t.Fatalf("legacy recovery done: %v", err)
	}

	// Shape 2 — reassignment into an existing target row moves the binding.
	shapeExisting := addShapeCard(t, "shape existing target")
	pick(shapeExisting)
	fcPlace(t, root, homestate.Card{CardID: shapeExisting, RunID: fcRun, State: homestate.CardPicked, OwnerLabel: "worker-2", Version: 1})
	if err := assign(shapeExisting, fcRun, "worker-2"); err != nil {
		t.Fatalf("existing-target assign: %v", err)
	}
	if got := bindingRun(shapeExisting); got != fcRun {
		t.Fatalf("existing-target binding = %s, want %s", got, fcRun)
	}
}

// addShapeCard adds one queue card and returns its generated id.
func addShapeCard(t *testing.T, text string) string {
	t.Helper()
	out, _, err := runTodo(t, "add", text)
	if err != nil {
		t.Fatalf("todo add: %v", err)
	}
	id, _, ok := strings.Cut(strings.TrimSpace(out), " ")
	if !ok {
		t.Fatalf("add output %q lacks an id", out)
	}
	return id
}

// fcBindDispatch records the card's dispatch binding — the current-run
// authority the completion gate and scan resolve through. Idempotent.
func fcBindDispatch(t *testing.T, root, cardID, runID string) {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`INSERT INTO card_dispatch(card_id,run_id,recorded_at) VALUES(?,?,?) ON CONFLICT(card_id) DO UPDATE SET run_id=excluded.run_id,recorded_at=excluded.recorded_at`, cardID, runID, "2026-09-26T00:00:00Z"); err != nil {
		t.Fatalf("bind dispatch %s -> %s: %v", cardID, runID, err)
	}
}

// fcPlaceFactoryCard places the card's factory row (the factory-linked
// marker) with the version and evidence hash named, inside the recorded
// dispatch run.
func fcPlaceFactoryCard(t *testing.T, root, cardID string, version int64, evidenceSHA, updated string) {
	t.Helper()
	fcBindDispatch(t, root, cardID, fcRun)
	fcPlace(t, root, homestate.Card{CardID: cardID, RunID: fcRun, State: homestate.CardMergedLocal, OwnerLabel: "worker-1", Version: version, EvidenceSHA: evidenceSHA, MergeSHA: evidenceSHA, UpdatedAt: updated})
}

// fcPlaceApprovalRaw inserts a receipt row directly — the shape an
// issuer-role fixture needs, because the issuance surface refuses the
// performing lane's role marker before anything is stored.
func fcPlaceApprovalRaw(t *testing.T, root string, a homestate.LeaderApproval) {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	if a.IssuedAt == "" {
		a.IssuedAt = "2026-09-26T00:00:00Z"
	}
	if _, err := db.DB.Exec(`INSERT INTO leader_approvals(card_uuid,run_id,card_id,factory_version,evidence_hash,issuer,issuer_role,issued_at) VALUES(?,?,?,?,?,?,?,?)
		ON CONFLICT(card_uuid,run_id) DO UPDATE SET card_id=excluded.card_id,factory_version=excluded.factory_version,evidence_hash=excluded.evidence_hash,issuer=excluded.issuer,issuer_role=excluded.issuer_role,issued_at=excluded.issued_at`,
		a.CardUUID, a.RunID, a.CardID, a.FactoryVersion, a.EvidenceHash, a.Issuer, a.IssuerRole, a.IssuedAt); err != nil {
		t.Fatalf("place approval: %v", err)
	}
}

func fcLiveItem(t *testing.T, store *factory.BacklogStore, id string) bool {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for i := range rec.Items {
		if rec.Items[i].ID == id {
			return true
		}
	}
	return false
}

// AC-FCR-001 — a factory-linked card's done is refused without a leader
// approval receipt; the refusal names the reason on stderr and the row stays
// live.
func TestLeaderReceiptGateRejectsMissingReceipt(t *testing.T) {
	root, store := fcFixture(t)
	_ = fcLinkedCard(t, root, store, "t1", "factory-linked card")
	fcPlaceFactoryCard(t, root, "t1", 1, "sha-t1", "2026-09-26T00:00:00Z")

	_, stderr, err := runTodo(t, "done", "t1")
	if err == nil {
		t.Fatal("done accepted without a leader approval receipt")
	}
	if !strings.Contains(stderr, "leader approval") {
		t.Errorf("stderr %q does not name the leader approval reason", stderr)
	}
	if !fcLiveItem(t, store, "t1") {
		t.Error("the refused done archived the card")
	}

	// The scope sentence (REQ-FCR-002): a card with no factory row completes
	// as before, receipt-less.
	if _, _, err := runTodo(t, "add", "plain card"); err != nil {
		t.Fatalf("todo add plain: %v", err)
	}
	out, _, err := runTodo(t, "done", "t2")
	if err != nil {
		t.Fatalf("plain card done: %v", err)
	}
	if !strings.Contains(out, "done t2") {
		t.Errorf("plain card done output = %q", out)
	}
}

// AC-FCR-002 — the receipt's four bindings are each checked against the
// card's archive-moment factory row: uuid, run id (the cross-run receipt —
// the same card re-run restarts the version, so the run binding is the axis
// that stops the previous run's approval closing the new run's work), factory
// version, and evidence hash. All four matching closes.
func TestLeaderReceiptGateBinding(t *testing.T) {
	root, store := fcFixture(t)
	uuid := fcLinkedCard(t, root, store, "t1", "factory-linked card")

	// The archive-moment row: run R2, version 1, evidence sha-r2 — the
	// recorded active dispatch (newer-created run).
	fcBindDispatch(t, root, "t1", "run-r2")
	fcPlace(t, root, homestate.Card{CardID: "t1", RunID: "run-r2", State: homestate.CardMergedLocal, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-r2", UpdatedAt: "2026-09-26T02:00:00Z"})
	// The previous run's row: identical uuid, version, and evidence — the
	// shape whose receipt must NOT close run R2 (REQ-FCR-001).
	fcPlace(t, root, homestate.Card{CardID: "t1", RunID: "run-r1", State: homestate.CardDone, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-r2", UpdatedAt: "2026-09-26T01:00:00Z"})

	base := homestate.LeaderApproval{
		CardUUID: uuid, RunID: "run-r2", CardID: "t1", FactoryVersion: 1,
		EvidenceHash: "sha-r2", Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
	}
	for name, mutate := range map[string]func(a *homestate.LeaderApproval){
		"uuid":          func(a *homestate.LeaderApproval) { a.CardUUID = "uuid-other" },
		"run":           func(a *homestate.LeaderApproval) { a.RunID = "run-r1" },
		"stale version": func(a *homestate.LeaderApproval) { a.FactoryVersion = 0 },
		"evidence hash": func(a *homestate.LeaderApproval) { a.EvidenceHash = "sha-other" },
	} {
		a := base
		mutate(&a)
		fcPlaceApprovalRaw(t, root, a)
		_, stderr, err := runTodo(t, "done", "t1")
		if err == nil {
			t.Fatalf("%s variant: done accepted a receipt that does not bind the card", name)
		}
		if !strings.Contains(stderr, "leader approval") {
			t.Errorf("%s variant: stderr %q does not name the reason", name, stderr)
		}
		if !fcLiveItem(t, store, "t1") {
			t.Fatalf("%s variant: the refused done archived the card", name)
		}
	}

	// All four matching: the close goes through.
	fcPlaceApprovalRaw(t, root, base)
	out, _, err := runTodo(t, "done", "t1")
	if err != nil {
		t.Fatalf("done with the matching receipt: %v", err)
	}
	if !strings.Contains(out, "done t1") {
		t.Errorf("done output = %q", out)
	}
	if fcLiveItem(t, store, "t1") {
		t.Error("the card stayed live after the gated close")
	}
}

// AC-FCR-003 (authored as TestLeaderReceiptGateIssuerIndependence — renamed
// from acceptance.md's TestLeaderReceiptGateSelfIssued cite per review debt
// D16: the axis is performer ≠ approver, and the old name names only one
// arm). The card's performing owner cannot approve its own work; the leader
// issuing the receipt and executing the done is the normal single-leader
// flow and passes.
func TestLeaderReceiptGateIssuerIndependence(t *testing.T) {
	root, store := fcFixture(t)
	uuid := fcLinkedCard(t, root, store, "t1", "factory-linked card")
	fcPlaceFactoryCard(t, root, "t1", 1, "sha-t1", "2026-09-26T00:00:00Z")

	// The performing owner (worker-1, the factory row's owner) issued the
	// receipt for its own work — refused on the issuer axis alone.
	fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
		CardUUID: uuid, RunID: fcRun, CardID: "t1", FactoryVersion: 1,
		EvidenceHash: "sha-t1", Issuer: "worker-1", IssuerRole: homestate.ApprovalIssuerLane,
	})
	_, stderr, err := runTodo(t, "done", "t1")
	if !errors.Is(err, homestate.ErrApprovalIssuer) && (err == nil || !strings.Contains(stderr+err.Error(), "performing owner")) {
		t.Fatalf("performer-issued receipt: err=%v stderr=%q, want the issuer refusal", err, stderr)
	}
	if !fcLiveItem(t, store, "t1") {
		t.Fatal("the performer-issued receipt closed the card")
	}

	// The leader issued the receipt and this same leader process executes
	// the done — the single-leader normal flow. The receipt row carries
	// Issuer=lead; the executing CLI carries no issuer identity at all, so
	// the pass shows issuer ≠ refusal reason.
	fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
		CardUUID: uuid, RunID: fcRun, CardID: "t1", FactoryVersion: 1,
		EvidenceHash: "sha-t1", Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
	})
	out, _, err := runTodo(t, "done", "t1")
	if err != nil {
		t.Fatalf("leader-issued, leader-executed done: %v", err)
	}
	if !strings.Contains(out, "done t1") {
		t.Errorf("done output = %q", out)
	}
}

// REQ-FCR-005's performer axis on the backlog gate: a receipt whose Issuer
// equals the factory row's performing owner is refused even when its role
// marker reads leader — the marker is data, and the performer check is the
// gate's own comparison.
func TestLeaderReceiptGatePerformerOwnerRefused(t *testing.T) {
	root, store := fcFixture(t)
	_ = fcLinkedCard(t, root, store, "t1", "factory-linked card")
	fcPlaceFactoryCard(t, root, "t1", 1, "sha-t1", "2026-09-26T00:00:00Z")
	fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
		CardUUID: "uuid-performer", RunID: fcRun, CardID: "t1", FactoryVersion: 1,
		EvidenceHash: "sha-t1", Issuer: "worker-1", IssuerRole: homestate.ApprovalIssuerLeader,
	})
	_, _, err := runTodo(t, "done", "t1")
	if err == nil {
		t.Fatal("a receipt issued by the performing owner was accepted despite the leader role marker")
	}
	if !fcLiveItem(t, store, "t1") {
		t.Fatal("the performing owner's receipt closed the card")
	}
}

// TestTodoAutoGateRefusesWithoutReceipt — the --auto cycle's archive point
// runs the same gate (REQ-FCR-002a, surface todo_auto.go): a factory-linked
// picked card whose worker evidence arrives is NOT archived without a
// receipt — the cycle reports the non-finding and the card stays picked.
func TestTodoAutoGateRefusesWithoutReceipt(t *testing.T) {
	root, store := autoFixture(t, map[string]string{"a": "card a"})
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	id := rec.Items[0].ID
	autoSetState(t, store, id, factory.BacklogStatePicked)
	if err := factory.RecordFactoryCardState(root, fcRun, id, "worker-1", "", "picked", "card.assigned"); err != nil {
		t.Fatalf("record runtime assignment: %v", err)
	}
	fcPlaceFactoryCard(t, root, id, 1, "sha-auto", "2026-09-26T00:00:00Z")

	lv := autoTestLiveness(root, id, true, true, nil)
	tick := 0
	opts := autoOptions{
		wait:      5 * time.Minute,
		liveness:  lv,
		sessionID: "operator-session-fixture",
		sleep: func(time.Duration) {
			tick++
			path := filepath.Join(root, ".moai", "reports", id, "evidence.md")
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("# evidence: verbatim output\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		},
		now: func() time.Time { return time.Unix(0, 0).Add(time.Duration(tick) * time.Minute) },
	}

	var out strings.Builder
	if err := runAutoCycle(&out, store, root, opts); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if strings.Contains(got, "done "+id) {
		t.Errorf("the --auto cycle archived a factory-linked card without a receipt:\n%s", got)
	}
	if !strings.Contains(got, "non-finding") {
		t.Errorf("the refusal carried no labelled non-finding:\n%s", got)
	}
	if !fcLiveItem(t, store, id) {
		t.Error("the refused --auto archive removed the card from the live queue")
	}
}

// The issuance surface (REQ-FCR-014): `factory approve` is the leader path's
// receipt mint — refused in a lane session, refused for a card with no
// factory row, and binding the CURRENT row's version and evidence hash (the
// caller names nothing the row does not carry).
func TestFactoryApproveIssuance(t *testing.T) {
	root, store := fcFixture(t)
	_ = fcLinkedCard(t, root, store, "t1", "factory-linked card")

	// A lane session cannot mint a receipt.
	t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
	if _, _, err := runFactory(t, "approve", "t1", "--run", fcRun); err == nil {
		t.Fatal("a lane session issued a leader approval receipt")
	}
	t.Setenv(config.EnvFactoryRole, "")

	// No factory row: the card is not factory-linked; nothing to approve.
	if _, _, err := runTodo(t, "add", "plain card"); err != nil {
		t.Fatalf("todo add plain: %v", err)
	}
	if _, stderr, err := runFactory(t, "approve", "t2", "--run", fcRun); err == nil || !strings.Contains(stderr, "factory") {
		t.Fatalf("approve without a factory row: err=%v stderr=%q, want refusal", err, stderr)
	}

	// The happy path: the receipt binds the current row, and the gated done
	// accepts it — issuance to close, end to end.
	fcPlaceFactoryCard(t, root, "t1", 1, "sha-t1", "2026-09-26T00:00:00Z")
	out, _, err := runFactory(t, "approve", "t1", "--run", fcRun, "--issuer", "lead")
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if !strings.Contains(out, "t1") || !strings.Contains(out, fcRun) {
		t.Errorf("approve output = %q, want the card id and run", out)
	}
	doneOut, _, err := runTodo(t, "done", "t1")
	if err != nil {
		t.Fatalf("done after approve: %v", err)
	}
	if !strings.Contains(doneOut, "done t1") {
		t.Errorf("done output = %q", doneOut)
	}
}

// Regression pin for round-2 review P1 and round-3 review P1 (card t1538):
// the completion gate keys on the DISPATCH BINDING the runs table carries —
// an active run wins over a retired one, and a reassignment into a newer
// run wins — never on a row's modification time and never on the backlog's
// runtime assignment record (which `factory assign` does not update). The
// old run's row carries the newest updated_at AND a matching receipt; the
// gate must still resolve the active run and refuse.
func TestLeaderReceiptGateKeysOnActiveDispatch(t *testing.T) {
	root, store := fcFixture(t)
	uuid := fcLinkedCard(t, root, store, "t1", "factory-linked card")
	// run-old: retired, but its row carries the newest updated_at and a
	// matching receipt — the forgery bait both prior reviews used.
	fcBindDispatch(t, root, "t1", fcRun)
	fcPlace(t, root, homestate.Card{CardID: "t1", RunID: "run-old", State: homestate.CardMergedLocal, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-old", UpdatedAt: "2026-09-26T03:00:00Z"})
	fcPlace(t, root, homestate.Card{CardID: "t1", RunID: fcRun, State: homestate.CardMergedLocal, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-cur", UpdatedAt: "2026-09-26T01:00:00Z"})
	fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
		CardUUID: uuid, RunID: "run-old", CardID: "t1", FactoryVersion: 1,
		EvidenceHash: "sha-old", Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
	})

	_, stderr, err := runTodo(t, "done", "t1")
	if err == nil {
		t.Fatal("done closed on the retired run's receipt — the gate keyed on modification time")
	}
	if !strings.Contains(stderr, "leader approval") {
		t.Errorf("stderr %q does not name the leader approval reason", stderr)
	}
	if !fcLiveItem(t, store, "t1") {
		t.Fatal("the refused done archived the card")
	}

	// The active run's own receipt closes (repro shape 2).
	fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
		CardUUID: uuid, RunID: fcRun, CardID: "t1", FactoryVersion: 1,
		EvidenceHash: "sha-cur", Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
	})
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done with the active run's receipt: %v", err)
	}
}

// Regression pin for review P1-1 (card t1538): the close's verification must
// read the archive-moment row even when a factory transition commits while
// the done command is in flight. A concurrent writer holds the factory write
// lock, bumps the card to version 2, and commits mid-command; the done's
// gate blocks on the lock, then reads the bumped row and refuses the
// version-1 receipt as stale.
func TestLeaderReceiptGateReadsArchiveMomentRow(t *testing.T) {
	root, store := fcFixture(t)
	uuid := fcLinkedCard(t, root, store, "t1", "factory-linked card")
	fcPlaceFactoryCard(t, root, "t1", 1, "sha-t1", "2026-09-26T00:00:00Z")
	fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
		CardUUID: uuid, RunID: fcRun, CardID: "t1", FactoryVersion: 1,
		EvidenceHash: "sha-t1", Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
	})

	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	blocker, err := homestate.OpenFactoryPath(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = blocker.Close() }()
	ctx := context.Background()
	tx, err := blocker.DB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	committed := make(chan error, 1)
	go func() {
		time.Sleep(300 * time.Millisecond)
		if _, err := tx.ExecContext(ctx, `UPDATE cards SET version=2, state='needs-decision', updated_at='2026-09-26T09:00:00Z' WHERE card_id='t1'`); err != nil {
			committed <- err
			_ = tx.Rollback()
			return
		}
		committed <- tx.Commit()
	}()
	// Let the blocker take the write lock before the done starts, so the
	// done's gate is the one that waits.
	time.Sleep(50 * time.Millisecond)

	_, stderr, err := runTodo(t, "done", "t1")
	if err == nil {
		t.Fatal("done closed on a version-1 receipt while the row was bumped mid-flight")
	}
	if !strings.Contains(stderr, "leader approval") {
		t.Errorf("stderr %q does not name the leader approval reason", stderr)
	}
	if !fcLiveItem(t, store, "t1") {
		t.Fatal("the refused done archived the card")
	}
	if err := <-committed; err != nil {
		t.Fatalf("concurrent bump: %v", err)
	}
	if c := fcCard(t, root, "t1"); c.Version != 2 {
		t.Fatalf("bump did not land: version=%d", c.Version)
	}
}

// Regression pin for review P1-2 (card t1538): only a MISSING factory
// database exempts the gate. A database that cannot even be stat'ed (here:
// permission denied on its directory) is a database the close cannot verify
// against — the close is refused, never silently completed unchecked.
func TestLeaderReceiptGateRefusesUnstatableFactoryDB(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("permission bits do not bind root")
	}
	root, store := fcFixture(t)
	_ = fcLinkedCard(t, root, store, "t1", "factory-linked card")
	fcPlaceFactoryCard(t, root, "t1", 1, "sha-t1", "2026-09-26T00:00:00Z")

	factoryDir, err := homestate.FactoryDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(factoryDir, 0o000); err != nil {
		t.Fatalf("chmod factory dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(factoryDir, 0o700) })

	_, stderr, err := runTodo(t, "done", "t1")
	if err == nil {
		t.Fatal("done completed while the factory database could not be read")
	}
	if !strings.Contains(stderr, "leader approval") {
		t.Errorf("stderr %q does not name the leader approval reason", stderr)
	}
	if !fcLiveItem(t, store, "t1") {
		t.Fatal("the refused done archived the card")
	}
}

// Regression pin for round-4 review P2-1 (card t1538): the gate's factory
// transaction is deliberately opened on a detached context — a request
// cancellation after verification must not roll the lock back while the
// guarded backlog save continues. The helper accepts a canceled context
// and still returns a live, usable gate.
func TestApprovalGateOutlivesContextCancel(t *testing.T) {
	root, store := fcFixture(t)
	uuid := fcLinkedCard(t, root, store, "t1", "factory-linked card")
	fcPlaceFactoryCard(t, root, "t1", 1, "sha-t1", "2026-09-26T00:00:00Z")
	fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
		CardUUID: uuid, RunID: fcRun, CardID: "t1", FactoryVersion: 1,
		EvidenceHash: "sha-t1", Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	gate, err := holdDoneApprovalGate(ctx, root)
	if err != nil {
		t.Fatalf("gate died with the canceled request context: %v", err)
	}
	if gate == nil {
		t.Fatal("gate missing with an existing factory database")
	}
	if err := gate.verifyForClose(context.Background(), "t1", uuid); err != nil {
		t.Fatalf("verify after request cancellation: %v", err)
	}
	gate.release()
}

// Regression pin for round-5 review P2 (card t1538): refreshDoneApprovalGate
// is the archive-moment re-check. With no factory database it stays nil (the
// card is not factory-linked); once the racing first dispatch has created
// the database and binding, the refresh opens a real gate whose verification
// refuses an approval-less close.
func TestRefreshDoneApprovalGateRacingFirstDispatch(t *testing.T) {
	root, store := fcFixture(t)
	// Deliberately no factory database yet: the pre-read sees none.
	gate, err := holdDoneApprovalGate(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	refreshed, err := refreshDoneApprovalGate(context.Background(), root, gate)
	if err != nil {
		t.Fatal(err)
	}
	if refreshed != nil {
		t.Fatal("refresh opened a gate although no factory database existed")
	}
	gate.release()

	// The racing dispatch: card row + dispatch binding now exist.
	uuid := fcLinkedCard(t, root, store, "t1", "racing factory-linked card")
	fcPlaceFactoryCard(t, root, "t1", 1, "sha-t1", "2026-09-26T00:00:00Z")

	refreshed, err = refreshDoneApprovalGate(context.Background(), root, refreshed)
	if err != nil {
		t.Fatal(err)
	}
	if refreshed == nil {
		t.Fatal("refresh did not open the gate after the factory database appeared")
	}
	if err := refreshed.verifyForClose(context.Background(), "t1", uuid); !errors.Is(err, homestate.ErrApprovalMissing) {
		t.Fatalf("racing dispatch verify err = %v, want ErrApprovalMissing", err)
	}
	refreshed.release()
}

// Regression pin for round-2 review P2-1 (card t1538): `factory approve
// --run <run>` reads THE NAMED RUN's row — never the most recently modified
// row of any run — and refuses a run that has no row for the card.
func TestFactoryApproveBindsNamedRun(t *testing.T) {
	root, store := fcFixture(t)
	_ = fcLinkedCard(t, root, store, "t1", "factory-linked card")
	fcPlace(t, root, homestate.Card{CardID: "t1", RunID: "run-old", State: homestate.CardMergedLocal, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-old", UpdatedAt: "2026-09-26T01:00:00Z"})
	fcPlace(t, root, homestate.Card{CardID: "t1", RunID: fcRun, State: homestate.CardMergedLocal, OwnerLabel: "worker-1", Version: 4, EvidenceSHA: "sha-new", UpdatedAt: "2026-09-26T02:00:00Z"})

	// A run with no row for the card is refused outright.
	if _, _, err := runFactory(t, "approve", "t1", "--run", "run-ghost"); err == nil {
		t.Fatal("approve accepted a run with no factory row for the card")
	}
	// The named run's OWN row is what the receipt binds (version 1 and
	// sha-old — not the latest row's version 4 / sha-new).
	out, _, err := runFactory(t, "approve", "t1", "--run", "run-old")
	if err != nil {
		t.Fatalf("approve run-old: %v", err)
	}
	if !strings.Contains(out, "version=1") || !strings.Contains(out, "sha-old") {
		t.Errorf("approve output = %q, want run-old's own version and evidence", out)
	}
}
