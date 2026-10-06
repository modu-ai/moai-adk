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
	// The assigned row's evidence lands with the commit; approve binds it.
	db = fcOpen(t, root)
	if _, err := db.DB.Exec(`UPDATE cards SET evidence_sha='sha-final' WHERE card_id='t1' AND run_id=?`, fcRun); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
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
	fcPlace(t, root, homestate.Card{CardID: shapeLegacy, RunID: "run-old", State: homestate.CardAssigned, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-legacy", UpdatedAt: "2026-09-26T01:00:00Z"})
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

// Regression matrix for round-7 review P1-1/P1-2 (card t1538): every
// successful assignment/recovery shape records the binding for the run the
// command targeted — regardless of row version, prior state, or a stale
// prior binding — and failed/refused shapes leave it untouched.
func TestFactoryAssignBindingMatrix(t *testing.T) {
	root, store := fcFixture(t)
	addCard := func(text string) string {
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
	assign := func(card, run, lane string) error {
		_, _, err := runFactory(t, "assign", card, "--to", lane, "--run", run)
		return err
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

	// --to-less success at v2+ on an existing picked row: the binding moves
	// to the targeted run (review round-7 P1-1).
	cardID := addCard("matrix --to-less v2")
	pick(cardID)
	fcPlace(t, root, homestate.Card{CardID: cardID, RunID: fcRun, State: homestate.CardPicked, OwnerLabel: "worker-1", Version: 2})
	fcBindDispatch(t, root, cardID, "run-old")
	if _, _, err := runFactory(t, "assign", cardID, "--run", fcRun); err != nil {
		t.Fatalf("--to-less assign: %v", err)
	}
	if got := bindingRun(cardID); got != fcRun {
		t.Fatalf("--to-less binding = %s, want %s", got, fcRun)
	}

	// State-preserving recovery at every post-assigned state, including
	// terminal done (review round-7 P1-2). Each row starts bound to the
	// WRONG run; the same-owner recovery must re-bind to the row's run.
	for _, state := range []string{homestate.CardAssigned, homestate.CardMergeReady, homestate.CardMergedLocal, homestate.CardPushed, homestate.CardCIGreen, homestate.CardDone} {
		recoverID := addCard("matrix recovery " + state)
		pick(recoverID)
		fcPlace(t, root, homestate.Card{CardID: recoverID, RunID: "run-old", State: state, OwnerLabel: "worker-1", Version: 3})
		if err := assign(recoverID, "run-old", "worker-1"); err != nil {
			t.Fatalf("recovery assign %s: %v", state, err)
		}
		if got := bindingRun(recoverID); got != "run-old" {
			t.Fatalf("recovery %s binding = %s, want run-old", state, got)
		}
		// State and version are preserved: binding-only, no transition.
		db := fcOpen(t, root)
		row, rowErr := db.LoadCard(context.Background(), "run-old", recoverID)
		_ = db.Close()
		if rowErr != nil {
			t.Fatalf("recovery %s load: %v", state, rowErr)
		}
		if row.State != state || row.Version != 3 {
			t.Fatalf("recovery %s mutated the row: state=%s version=%d", state, row.State, row.Version)
		}
	}

	// A different-lane re-bind on the SAME non-picked row is refused and
	// leaves the binding untouched. (Assigning to a different run creates
	// that run's own row and is the legitimate reassignment shape above.)
	refuseID := addCard("matrix lane refusal")
	pick(refuseID)
	fcPlace(t, root, homestate.Card{CardID: refuseID, RunID: "run-old", State: homestate.CardAssigned, OwnerLabel: "worker-1", Version: 1})
	fcBindDispatch(t, root, refuseID, "run-old")
	if _, _, err := runFactory(t, "assign", refuseID, "--to", "worker-9", "--run", "run-old"); err == nil {
		t.Fatal("different-lane re-bind was accepted")
	}
	if got := bindingRun(refuseID); got != "run-old" {
		t.Fatalf("refused re-bind moved the binding: %s, want run-old", got)
	}
}

// Regression pin for round-7 review P1-REPEAT (card t1538): a FIRST dispatch
// serializes with a completion on the queue lock. The dispatch acquires the
// queue lock first and creates the factory database inside it; the done
// command then waits for the lock, opens the gate, sees the binding, and
// refuses the approval-less close — the archive can never slip through a
// stat-then-assign window.
func TestFirstDispatchSerializesWithCompletion(t *testing.T) {
	root, store := fcFixture(t)
	if _, _, err := runTodo(t, "add", "serial first dispatch card"); err != nil {
		t.Fatal(err)
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

	locked := make(chan struct{})
	dispatchErr := make(chan error, 1)
	go func() {
		dispatchErr <- store.WithLock(func(l *factory.LockedBacklog) error {
			close(locked)
			// Hold the lock long enough that the concurrent done is
			// demonstrably waiting behind the dispatch.
			time.Sleep(300 * time.Millisecond)
			db, err := homestate.OpenFactory(root)
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()
			now := time.Now().UTC()
			if _, err := db.RecordPicked(context.Background(), fcRun, "t1", homestate.CardFields{}, "dispatch", now); err != nil {
				return err
			}
			if _, err := db.Transition(context.Background(), homestate.TransitionRequest{RunID: fcRun, CardID: "t1", To: homestate.CardAssigned, ExpectedVersion: 1, Actor: "dispatch", Owner: "worker-1", Now: now}); err != nil {
				return err
			}
			return db.RecordDispatchBinding(context.Background(), "t1", fcRun, now)
		})
	}()
	<-locked

	_, stderr, err := runTodo(t, "done", "t1")
	if err == nil {
		t.Fatal("done completed while the racing first dispatch had created the factory record")
	}
	if !strings.Contains(stderr, "leader approval") {
		t.Errorf("stderr %q does not name the leader approval reason", stderr)
	}
	if !fcLiveItem(t, store, "t1") {
		t.Fatal("the refused done archived the card")
	}
	if err := <-dispatchErr; err != nil {
		t.Fatalf("racing dispatch failed: %v", err)
	}
	if c := fcCard(t, root, "t1"); c.State != homestate.CardAssigned {
		t.Fatalf("racing dispatch row = %s, want assigned", c.State)
	}
}

// Regression pin for round-7 review P2 (card t1538): decide reads the
// approval uuid from the queue at the PROJECT ROOT the factory DB was
// opened at — never from the server cwd's queue. The decoy environment
// points CLAUDE_PROJECT_DIR elsewhere; pre-fix the valid approval was
// refused with an identity error.
func TestFactoryDecideUsesTargetProjectRoot(t *testing.T) {
	root, store := fcFixture(t)
	cardID := addShapeCard(t, "target root decide card")
	fcLinkRuntime(t, root, cardID)
	fcBindDispatch(t, root, cardID, fcRun)
	bare, bareMerge := fcRepo(t, false)
	fcPlace(t, root, homestate.Card{CardID: cardID, RunID: fcRun, State: homestate.CardMergedLocal, OwnerLabel: "worker-1", Version: 1, MergeSHA: bareMerge, WorktreePath: bare, EvidenceSHA: bareMerge})
	uuid := recheckUUID(t, root, store, cardID)
	fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
		CardUUID: uuid, RunID: fcRun, CardID: cardID, FactoryVersion: 1,
		EvidenceHash: bareMerge, Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
	})

	// Point the environment at a decoy project with an empty queue: the
	// target root argument must win for the identity lookup.
	decoy := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", decoy)

	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	card, err := decideOne(context.Background(), db, root, fcRun, cardID, "push", "", "")
	if err != nil {
		t.Fatalf("decide with a decoy cwd queue: %v", err)
	}
	if card.State != homestate.CardDone {
		t.Fatalf("decide = %s, want done", card.State)
	}
}

// Regression pin for round-10 P2-2 (card t1538): factory-only recovery
// decisions never touch the queue record — a corrupt backlog database
// cannot block an abandon.
func TestDecideAbandonWorksWithCorruptQueue(t *testing.T) {
	root, _ := fcFixture(t)
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "worker-1", Version: 1})
	// Materialize the queue, then corrupt it.
	if _, _, err := runTodo(t, "add", "queue materializer"); err != nil {
		t.Fatal(err)
	}
	// Corrupt the queue database: the decide path must not read it.
	dbPath := filepath.Join(filepath.Dir(todoBacklogPath(root)), "backlog.db")
	if err := os.WriteFile(dbPath, []byte("this is not a database"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runFactory(t, "decide", "t1", "--choice", "abandon", "--run", fcRun); err != nil {
		t.Fatalf("abandon with a corrupt queue: %v", err)
	}
	if c := fcCard(t, root, "t1"); c.State != homestate.CardAbandoned {
		t.Fatalf("card = %s, want abandoned", c.State)
	}
}

// Regression pin for round-13 P1 (card t1538): a LANE cannot run the
// state-preserving re-bind — rotating the binding onto a past row would
// revive that row's old approval and let the in-flight work close without
// the leader's approval. The refusal leaves the binding and both receipts
// inert.
func TestLaneRebindRefusedAndOldApprovalInert(t *testing.T) {
	root, store := fcFixture(t)
	cardID := addShapeCard(t, "lane re-bind target")
	pick := func() {
		if err := store.Mutate(func(r *factory.BacklogRecord) error {
			for i := range r.Items {
				if r.Items[i].ID == cardID {
					r.Items[i].State = factory.BacklogStatePicked
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	// In-flight work: run-cur holds the card at merged-local v2 with the
	// dispatch binding and NO approval yet.
	fcPlace(t, root, homestate.Card{CardID: cardID, RunID: "run-old", State: homestate.CardDone, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-old", UpdatedAt: "2026-09-26T01:00:00Z"})
	fcPlace(t, root, homestate.Card{CardID: cardID, RunID: fcRun, State: homestate.CardMergedLocal, OwnerLabel: "worker-1", Version: 2, EvidenceSHA: "sha-new", UpdatedAt: "2026-09-26T02:00:00Z"})
	fcBindDispatch(t, root, cardID, fcRun)
	// The past row's approval, recorded under run-old at its own version.
	fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
		CardUUID: "uuid-lane", RunID: "run-old", CardID: cardID, FactoryVersion: 1,
		EvidenceHash: "sha-old", Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
	})

	pick()
	t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
	if _, _, err := runFactory(t, "assign", cardID, "--run", "run-old", "--to", "worker-1"); err == nil {
		t.Fatal("a lane session rotated the dispatch binding")
	}
	t.Setenv(config.EnvFactoryRole, "")

	// The binding still names the in-flight run, whose completion gate
	// refuses (no receipt for run-cli) — the old receipt stayed inert.
	bindDB := fcOpen(t, root)
	row, linked, rerr := bindDB.RecordedCardRowReadonly(context.Background(), cardID)
	_ = bindDB.Close()
	if rerr != nil || !linked {
		t.Fatalf("binding read: linked=%v err=%v", linked, rerr)
	}
	if row.RunID != fcRun {
		t.Fatalf("binding run = %s, want %s", row.RunID, fcRun)
	}
	if _, _, err := runTodo(t, "done", cardID); err == nil {
		t.Fatal("the revived old approval closed the in-flight work")
	}
	if !fcLiveItem(t, store, cardID) {
		t.Fatal("the card was archived on the revived approval")
	}
}

// Regression pin for round-13 P1 continuation (card t1538): a re-selection
// (`todo pick`) into a new run re-points the dispatch binding with the
// assignment record — the leftover old run's approval goes inert, and the
// done gate refuses the approval-less new work.
func TestTodoPickUpdatesDispatchBinding(t *testing.T) {
	root, store := fcFixture(t)
	cardID := addShapeCard(t, "re-selected card")
	// The old run: a done row with a valid-looking approval.
	fcPlace(t, root, homestate.Card{CardID: cardID, RunID: "run-old", State: homestate.CardDone, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-old", UpdatedAt: "2026-09-26T01:00:00Z"})
	fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
		CardUUID: "uuid-old", RunID: "run-old", CardID: cardID, FactoryVersion: 1,
		EvidenceHash: "sha-old", Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
	})
	// The new run: a picked row at v1 with no approval.
	fcPlaceFactoryCard(t, root, cardID, 1, "sha-new", "2026-09-26T02:00:00Z")

	// Re-selection records the assignment under run-new (env-driven) and —
	// the fix — re-points the dispatch binding in the same flow. `todo
	// next --card` is the re-selection surface the overlay drove.
	t.Setenv(config.EnvFactoryRunID, fcRun)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	if _, _, err := runTodo(t, "next", cardID); err != nil {
		t.Fatalf("next: %v", err)
	}

	bindDB := fcOpen(t, root)
	row, linked, rerr := bindDB.RecordedCardRowReadonly(context.Background(), cardID)
	_ = bindDB.Close()
	if rerr != nil || !linked {
		t.Fatalf("binding read: linked=%v err=%v", linked, rerr)
	}
	if row.RunID != fcRun {
		t.Fatalf("binding run = %s, want %s", row.RunID, fcRun)
	}
	// The old run's approval is inert: the done gate resolves run-cli's
	// row, holds no receipt for it, and refuses.
	_, stderr, err := runTodo(t, "done", cardID)
	if err == nil {
		t.Fatal("done closed on the old run's approval after re-selection")
	}
	if !strings.Contains(stderr, "leader approval") {
		t.Errorf("stderr %q does not name the leader approval reason", stderr)
	}
	if !fcLiveItem(t, store, cardID) {
		t.Fatal("the refused done archived the card")
	}
}

// Regression pin for round-14 P1-1 (card t1538): a FAILED binding update
// refuses the pick — the old run's approval is never left armed on a card
// that just moved runs. The factory database path is turned into a
// directory so the record open fails while the existence probe succeeds.
func TestPickRefusedWhenBindingUpdateFails(t *testing.T) {
	root, store := fcFixture(t)
	cardID := addShapeCard(t, "binding failure card")
	fcPlaceRun(t, root, fcRun, "active", "2026-09-25T00:00:00Z")
	// A row already exists in the target run: the pick's binding write has
	// real work to do.
	fcPlace(t, root, homestate.Card{CardID: cardID, RunID: fcRun, State: homestate.CardPicked, OwnerLabel: "worker-1", Version: 1})

	factoryDB, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(factoryDB); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(factoryDB, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(factoryDB) })

	t.Setenv(config.EnvFactoryRunID, fcRun)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	if _, _, err := runTodo(t, "next", cardID); err == nil {
		t.Fatal("pick succeeded although the binding update could not run")
	}
	// The pick is refused ATOMICALLY: the queue write aborted, so the card
	// is back to queued (never stuck picked), nothing archived, and the
	// binding never moved.
	if _, ok := liveItemOK(t, store, cardID); !ok {
		t.Fatal("the refused pick removed the card")
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for i := range rec.Items {
		if rec.Items[i].ID == cardID {
			if rec.Items[i].State != factory.BacklogStateQueued {
				t.Fatalf("card stuck in %s after the refused pick, want queued", rec.Items[i].State)
			}
		}
	}
	// The factory database path is a directory at this point (the failure
	// injection) — no binding assertion is possible or needed: the queue
	// state above is the contract.
}

// Regression pin for round-14 P1-2 (card t1538): the binding write lands at
// the CALLER-SPECIFIED project root — never a server-cwd/env fallback. The
// environment points at a decoy project; the explicit root still receives
// the binding.
func TestRecordDispatchBindingAtExplicitRoot(t *testing.T) {
	target := t.TempDir()
	fcPlaceRun(t, target, fcRun, "active", "2026-09-25T00:00:00Z")
	fcPlace(t, target, homestate.Card{CardID: "t9", RunID: fcRun, State: homestate.CardPicked, OwnerLabel: "worker-1", Version: 1})

	decoy := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", decoy)

	if err := recordDispatchBindingAtRoot("t9", fcRun, target); err != nil {
		t.Fatalf("binding at explicit root: %v", err)
	}
	db := fcOpen(t, target)
	defer func() { _ = db.Close() }()
	row, linked, err := db.RecordedCardRowReadonly(context.Background(), "t9")
	if err != nil || !linked {
		t.Fatalf("binding read: linked=%v err=%v", linked, err)
	}
	if row.RunID != fcRun {
		t.Fatalf("binding run = %s, want %s", row.RunID, fcRun)
	}
}

// fcPlaceRun records a run row — the runs-table metadata some fixtures
// need beside their card rows. Idempotent.
func fcPlaceRun(t *testing.T, root, runID, status, created string) {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`INSERT INTO runs(run_id,status,manifest_json,created_at,updated_at) VALUES(?,?,'{}',?,?) ON CONFLICT(run_id) DO NOTHING`, runID, status, created, created); err != nil {
		t.Fatalf("place run %s: %v", runID, err)
	}
}

// Regression pin for round-15 P1-3 (card t1538): a claim whose dispatch
// binding cannot be recorded is REVERTED before anything prints — the card
// returns to queued with no lease, and no selection stands unbound.
func TestClaimRevertedWhenBindingUpdateFails(t *testing.T) {
	root, store := fcFixture(t)
	if _, _, err := runTodo(t, "add", "claim revert card"); err != nil {
		t.Fatal(err)
	}
	fcPlaceRun(t, root, fcRun, "active", "2026-09-25T00:00:00Z")
	fcPlace(t, root, homestate.Card{CardID: "t1", RunID: fcRun, State: string(factory.BacklogStateQueued), OwnerLabel: "worker-1", Version: 1})

	factoryDB, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(factoryDB); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(factoryDB, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(factoryDB) })

	t.Setenv(config.EnvFactoryRunID, fcRun)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	if _, _, err := runTodo(t, "claim"); err == nil {
		t.Fatal("claim succeeded although the binding update could not run")
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for i := range rec.Items {
		if rec.Items[i].ID == "t1" {
			if rec.Items[i].State != factory.BacklogStateQueued {
				t.Fatalf("card stuck in %s after the reverted claim, want queued", rec.Items[i].State)
			}
			if rec.Items[i].PickedBy != nil {
				t.Fatalf("reverted claim kept picked_by %q", *rec.Items[i].PickedBy)
			}
			if rec.Items[i].LeaseExpiresAt != nil {
				t.Fatalf("reverted claim kept a lease %q", *rec.Items[i].LeaseExpiresAt)
			}
			return
		}
	}
	t.Fatal("claim card vanished")
}

// Regression pin for round-16 P1-1 + P2 (card t1538): the claim and its
// dispatch binding land under ONE held queue lock, and the success line
// prints exactly ONCE. A binding failure rolls the claim back under the
// same lock and prints only the refusal.
func TestClaimBindsUnderLockAndPrintsOnce(t *testing.T) {
	root, _ := fcFixture(t)
	if _, _, err := runTodo(t, "add", "claim single print card"); err != nil {
		t.Fatal(err)
	}
	fcPlaceRun(t, root, fcRun, "active", "2026-09-25T00:00:00Z")
	fcPlace(t, root, homestate.Card{CardID: "t1", RunID: fcRun, State: string(factory.BacklogStateQueued), OwnerLabel: "worker-1", Version: 1})

	t.Setenv(config.EnvFactoryRunID, fcRun)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	out, _, err := runTodo(t, "claim")
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if got := strings.Count(out, "claimed t1 "); got != 1 {
		t.Fatalf("claimed line count = %d, want 1 (output %q)", got, out)
	}
	// The binding followed the claim (same lock).
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	row, linked, rerr := db.RecordedCardRowReadonly(context.Background(), "t1")
	if rerr != nil || !linked {
		t.Fatalf("binding read: linked=%v rerr=%v", linked, rerr)
	}
	if row.RunID != fcRun {
		t.Fatalf("binding run = %s, want %s", row.RunID, fcRun)
	}

	// A claim whose binding write fails rolls back and prints only the
	// refusal — never a claimed line for an unbound selection.
	factoryDB, ferr := homestate.FactoryDBPath(root)
	if ferr != nil {
		t.Fatal(ferr)
	}
	if err := os.Remove(factoryDB); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(factoryDB, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(factoryDB) })
	if _, _, err := runTodo(t, "add", "second claim card"); err != nil {
		t.Fatal(err)
	}
	if outp, _, err := runTodo(t, "claim"); err == nil {
		t.Fatal("claim with a failing binding write succeeded")
	} else if strings.Contains(outp, "claimed t2") {
		t.Fatal("the rolled-back claim still printed a claimed line")
	}
}

// Regression pin for round-18 P2 (card t1538): selecting an ORDINARY card
// (no factory row) while another card's factory database exists does NOT
// create a dangling dispatch binding — the ordinary card keeps its
// existing completion behavior.
func TestOrdinaryCardSelectionSkipsBinding(t *testing.T) {
	root, _ := fcFixture(t)
	// t1: the ordinary card — added to the queue, NO factory row anywhere.
	if _, _, err := runTodo(t, "add", "ordinary card"); err != nil {
		t.Fatal(err)
	}
	// t2: a factory-linked card (its row creates the factory database), so
	// the selection of t1 has a live database to wrongly bind into.
	fcPlaceFactoryCard(t, root, "t2", 1, "sha-t2", "2026-09-26T00:00:00Z")

	t.Setenv(config.EnvFactoryRunID, fcRun)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	if _, _, err := runTodo(t, "next", "t1"); err != nil {
		t.Fatalf("select the ordinary card: %v", err)
	}

	// No dangling binding: the factory database has no dispatch record for
	// the ordinary card, and no factory row either.
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	var n int
	if err := db.DB.QueryRow(`SELECT count(*) FROM card_dispatch WHERE card_id='t1'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("a dangling dispatch binding was created for the ordinary card (%d rows)", n)
	}
	var rows int
	if err := db.DB.QueryRow(`SELECT count(*) FROM cards WHERE card_id='t1'`).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 0 {
		t.Fatalf("a factory row was created for the ordinary card (%d rows)", rows)
	}
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

// Regression pin for round-6 review P2 (card t1538), end-to-end through the
// done verb: the pre-read finds no factory database, a racing FIRST
// dispatch creates it during the refresh window, and the archive-moment
// refresh must then refuse the approval-less close instead of archiving.
// The racing dispatch writes the factory store directly (a completion holds
// the queue lock, so a real assign would serialize — pinned separately by
// TestFirstDispatchSerializesWithCompletion).
func TestDoneGateRacingFirstDispatchUnderQueueLock(t *testing.T) {
	root, store := fcFixture(t)
	if _, _, err := runTodo(t, "add", "racing first dispatch card"); err != nil {
		t.Fatal(err)
	}

	raced := false
	prevStat := approvalGateStat
	approvalGateStat = func(path string) (os.FileInfo, error) {
		if !raced {
			raced = true
			return nil, os.ErrNotExist
		}
		// The racing dispatch lands between the pre-read and the refresh.
		db, err := homestate.OpenFactory(root)
		if err != nil {
			t.Errorf("racing dispatch open: %v", err)
			return prevStat(path)
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		if _, err := db.DB.Exec(`INSERT INTO runs(run_id,status,manifest_json,created_at,updated_at) VALUES(?,'active','{}',?,?) ON CONFLICT(run_id) DO NOTHING`, fcRun, now, now); err != nil {
			t.Errorf("racing runs row: %v", err)
		}
		if _, err := db.DB.Exec(`INSERT INTO card_dispatch(card_id,run_id,recorded_at) VALUES(?,?,?) ON CONFLICT(card_id) DO UPDATE SET run_id=excluded.run_id,recorded_at=excluded.recorded_at`, "t1", fcRun, now); err != nil {
			t.Errorf("racing binding: %v", err)
		}
		if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,owner_label,state,version,evidence_sha,updated_at) VALUES(?,'t1','worker-1','picked',1,'sha-t1',?)`, fcRun, now); err != nil {
			t.Errorf("racing card row: %v", err)
		}
		_ = db.Close()
		return prevStat(path)
	}
	t.Cleanup(func() { approvalGateStat = prevStat })

	_, stderr, err := runTodo(t, "done", "t1")
	if err == nil {
		t.Fatal("done archived a card that became factory-linked mid-flight without an approval")
	}
	if !strings.Contains(stderr, "leader approval") {
		t.Errorf("stderr %q does not name the leader approval reason", stderr)
	}
	if !fcLiveItem(t, store, "t1") {
		t.Fatal("the racing dispatch archived the card without an approval")
	}
	if c := fcCard(t, root, "t1"); c.CardID != "t1" {
		t.Fatalf("racing factory card missing: %+v", c)
	}
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
