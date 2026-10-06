package cli

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// The auto-done receipt gate (SPEC-FACTORY-COMPLETION-RECOVERY-001 M1,
// REQ-FCR-003/004): a factory-linked candidate whose leader approval receipt
// does not verify never closes; the scan continues around it. Non-factory
// candidates keep the receipt-less close.

// fcApprove issues a receipt through the leader path for a card whose uuid
// was minted by the runtime record write.
func fcApprove(t *testing.T, root, cardID, runID string, version int64, evidenceHash string) {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	store := factory.NewBacklogStore(todoBacklogPath(root))
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	var cardUUID string
	for i := range rec.Items {
		if rec.Items[i].ID == cardID {
			cardUUID = todoCardUUID(&rec.Items[i])
			break
		}
	}
	if cardUUID == "" {
		t.Fatalf("card %s has no projected identity", cardID)
	}
	_, err = db.IssueLeaderApproval(context.Background(), homestate.LeaderApproval{
		CardUUID: cardUUID, RunID: runID, CardID: cardID, FactoryVersion: version,
		EvidenceHash: evidenceHash, Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
	})
	if err != nil {
		t.Fatalf("issue approval: %v", err)
	}
}

// fcLinkRuntime mints the card's runtime identity (the factory-linked
// marker's uuid source) the way a real dispatch does.
func fcLinkRuntime(t *testing.T, root, cardID string) {
	t.Helper()
	if err := factory.RecordFactoryCardState(root, fcRun, cardID, "worker-1", "", "picked", "card.assigned"); err != nil {
		t.Fatalf("record runtime assignment: %v", err)
	}
}

// AC-FCR-004 — a factory-linked candidate with no verified receipt is a
// skip, the scan continues around it, a factory-linked candidate WITH a
// matching receipt still closes, and the non-factory control closes
// receipt-less (REQ-FCR-002's scope sentence).
func TestAutoDoneReceiptSkip(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t900", "linked, receipt missing", factory.BacklogStateQueued)
	seedCard(t, store, "t901", "linked, receipt valid", factory.BacklogStatePicked)
	seedCard(t, store, "t902", "plain control", factory.BacklogStatePicked)
	seedCard(t, store, "t903", "linked, receipt stale", factory.BacklogStateQueued)

	// t900/t901/t903 are factory-linked; t902 has no factory record at all.
	fcLinkRuntime(t, root, "t900")
	fcLinkRuntime(t, root, "t901")
	fcLinkRuntime(t, root, "t903")

	// The archive-moment rows. t903 carries a version-bumped row (2) — the
	// state after a factory transition invalidated the v1 receipt.
	fcPlaceFactoryCard(t, root, "t900", 1, "sha-evidence", "2026-09-26T00:00:00Z")
	fcPlaceFactoryCard(t, root, "t901", 1, "sha-evidence", "2026-09-26T00:00:00Z")
	fcPlaceFactoryCard(t, root, "t903", 2, "sha-evidence", "2026-09-26T00:00:00Z")

	commitOnRef(t, root, "Merge branch 'WT-a' into develop (card t900)")
	commitOnRef(t, root, "Merge branch 'WT-b' into develop (card t901)")
	commitOnRef(t, root, "Merge branch 'WT-c' into develop (card t902)")
	commitOnRef(t, root, "Merge branch 'WT-d' into develop (card t903)")
	materializeOriginDevelop(t, root)

	fcApprove(t, root, "t901", fcRun, 1, "sha-evidence")
	fcApprove(t, root, "t903", fcRun, 1, "sha-evidence") // binds v1; the row is at v2

	stdout, _, err := runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("auto-done: %v", err)
	}
	if !strings.Contains(stdout, "skip t900 reason=leader-unapproved") {
		t.Errorf("stdout %q lacks skip t900 reason=leader-unapproved", stdout)
	}
	if !strings.Contains(stdout, "skip t903 reason=leader-unapproved") {
		t.Errorf("stdout %q lacks skip t903 reason=leader-unapproved (the stale v1 receipt against the v2 row)", stdout)
	}
	if !strings.Contains(stdout, "done t901 landing=landed") {
		t.Errorf("stdout %q lacks the receipt-backed close", stdout)
	}
	if !strings.Contains(stdout, "done t902 landing=landed") {
		t.Errorf("stdout %q lacks the non-factory control close", stdout)
	}
	if _, err := store.LoadPure(); err != nil {
		t.Fatalf("load: %v", err)
	}
	if _, ok := liveItemOK(t, store, "t900"); !ok {
		t.Error("t900 was archived without a receipt")
	}
	if _, ok := liveItemOK(t, store, "t903"); !ok {
		t.Error("t903 was archived with a stale receipt")
	}
	if _, ok := liveItemOK(t, store, "t902"); ok {
		t.Error("t902 (non-factory control) stayed live — the control failed")
	}
}

// Regression pin for review P2-4 (card t1538): the scan's receipt selection
// matches the archive gate's — the receipt bound to the card's CURRENT run
// wins over a NEWER receipt minted against an older run. A re-dispatched
// card whose leader re-approved the current run early must close even if a
// stale approval for the previous run was recorded later; and with only the
// previous run's receipt the card skips leader-unapproved.
func TestAutoDoneReceiptCurrentRunPreferred(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t950", "current run approved first", factory.BacklogStateQueued)
	seedCard(t, store, "t951", "only old run approved", factory.BacklogStateQueued)
	fcLinkRuntime(t, root, "t950")
	fcLinkRuntime(t, root, "t951")

	// Both cards' current factory engagement is the recorded run (the
	// dispatch binding); run-a is the previous engagement with identical
	// bindings — but only the binding decides which receipt judges.
	fcBindDispatch(t, root, "t950", fcRun)
	fcBindDispatch(t, root, "t951", fcRun)
	fcPlace(t, root, homestate.Card{CardID: "t950", RunID: "run-a", State: homestate.CardDone, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-ev", UpdatedAt: "2026-09-26T01:00:00Z"})
	fcPlace(t, root, homestate.Card{CardID: "t950", RunID: fcRun, State: homestate.CardMergedLocal, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-ev", UpdatedAt: "2026-09-26T02:00:00Z"})
	fcPlace(t, root, homestate.Card{CardID: "t951", RunID: "run-a", State: homestate.CardDone, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-ev", UpdatedAt: "2026-09-26T01:00:00Z"})
	fcPlace(t, root, homestate.Card{CardID: "t951", RunID: fcRun, State: homestate.CardMergedLocal, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-ev", UpdatedAt: "2026-09-26T02:00:00Z"})

	commitOnRef(t, root, "Merge branch 'WT-p' into develop (card t950)")
	commitOnRef(t, root, "Merge branch 'WT-q' into develop (card t951)")
	materializeOriginDevelop(t, root)

	// t950: the current-run receipt is the OLDER one; the old run's receipt
	// was recorded LATER. The current-run receipt must win.
	uuid950 := recheckUUID(t, root, store, "t950")
	uuid951 := recheckUUID(t, root, store, "t951")
	db := fcOpen(t, root)
	if _, err := db.IssueLeaderApproval(context.Background(), homestate.LeaderApproval{
		CardUUID: uuid950, RunID: fcRun, CardID: "t950", FactoryVersion: 1,
		EvidenceHash: "sha-ev", Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
		IssuedAt: "2026-09-26T00:30:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.IssueLeaderApproval(context.Background(), homestate.LeaderApproval{
		CardUUID: uuid950, RunID: "run-a", CardID: "t950", FactoryVersion: 1,
		EvidenceHash: "sha-ev", Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
		IssuedAt: "2026-09-26T03:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	// t951: only the old run's (latest) receipt.
	fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
		CardUUID: uuid951, RunID: "run-a", CardID: "t951", FactoryVersion: 1,
		EvidenceHash: "sha-ev", Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
		IssuedAt: "2026-09-26T03:00:00Z",
	})

	stdout, _, err := runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("auto-done: %v", err)
	}
	if !strings.Contains(stdout, "done t950 landing=landed") {
		t.Errorf("stdout %q lacks the current-run close — the old run's newer receipt shadowed it", stdout)
	}
	if !strings.Contains(stdout, "skip t951 reason=leader-unapproved") {
		t.Errorf("stdout %q lacks skip t951 reason=leader-unapproved", stdout)
	}
}

// recheckUUID reads a card's projected identity.
func recheckUUID(t *testing.T, root string, store *factory.BacklogStore, id string) string {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for i := range rec.Items {
		if rec.Items[i].ID == id {
			if rec.Items[i].CardUUID == nil {
				t.Fatalf("card %s has no projected identity", id)
			}
			return *rec.Items[i].CardUUID
		}
	}
	t.Fatalf("card %s not found", id)
	return ""
}

// Regression pin for round-2 review P2-2 (card t1538): the scan — dry-run
// included — opens the factory database strictly read-only. On a store
// whose leader_approvals table is absent (an older schema), the scan reads
// factory-linked cards as unverified and migrates NOTHING: reopening the
// database afterwards must still find no leader_approvals table.
func TestAutoDoneScanDoesNotMigrateOldSchema(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t960", "old-schema store", factory.BacklogStateQueued)
	fcLinkRuntime(t, root, "t960")
	fcPlaceFactoryCard(t, root, "t960", 1, "sha-ev", "2026-09-26T00:00:00Z")
	commitOnRef(t, root, "Merge branch 'WT-r' into develop (card t960)")
	materializeOriginDevelop(t, root)

	// The non-factory control carries landing evidence but no factory row:
	// an old-schema store must not downgrade it to query-inconclusive.
	seedCard(t, store, "t961", "plain old-schema control", factory.BacklogStateQueued)
	commitOnRef(t, root, "Merge branch 'WT-s' into develop (card t961)")
	materializeOriginDevelop(t, root)

	db := fcOpen(t, root)
	if _, err := db.DB.Exec(`DROP TABLE leader_approvals`); err != nil {
		t.Fatalf("drop approvals table: %v", err)
	}
	if _, err := db.DB.Exec(`DROP TABLE card_dispatch`); err != nil {
		t.Fatalf("drop dispatch table: %v", err)
	}
	_ = db.Close()

	stdout, _, err := runTodo(t, "auto-done", "--dry-run")
	if err != nil {
		t.Fatalf("auto-done dry-run: %v", err)
	}
	if !strings.Contains(stdout, "skip t960 reason=leader-unapproved") {
		t.Errorf("stdout %q lacks skip t960 reason=leader-unapproved (missing-table read)", stdout)
	}
	if strings.Contains(stdout, "skip t961 reason=query-inconclusive") {
		t.Errorf("the ordinary card was misclassified query-inconclusive:\n%s", stdout)
	}
	if !strings.Contains(stdout, "done t961 landing=landed") {
		t.Errorf("stdout %q lacks the ordinary card's dry-run close", stdout)
	}

	// Reopen STRICTLY READ-ONLY for the check — a write-mode open runs the
	// schema DDL and would recreate the table itself, masking the defect.
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	ro, err := homestate.OpenFactoryReadonly(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ro.Close() }()
	present, err := ro.FactoryTablePresent(context.Background(), "leader_approvals")
	if err != nil {
		t.Fatal(err)
	}
	if present {
		t.Fatal("the scan migrated the store: leader_approvals was recreated")
	}
}

// fcApprovalsRow bumps a card's factory row version directly — the
// archive-moment state after a concurrent factory transition.
func fcBumpRowVersion(t *testing.T, root, cardID string, version int64) {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`UPDATE cards SET version=?, updated_at='2026-09-26T09:00:00Z' WHERE card_id=?`, version, cardID); err != nil {
		t.Fatalf("bump row: %v", err)
	}
}

// fcReidentityItemColumn rewires the card's identity in the identity table —
// the table every record read projects CardUUID from (todoIdentitySnapshot
// apply overwrites unconditionally). Returns the fresh uuid so the caller
// can restore the original afterwards.
func fcReidentityItemColumn(t *testing.T, root, cardID string) string {
	t.Helper()
	fresh, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	fcSetItemUUID(t, root, cardID, fresh.String())
	return fresh.String()
}

// fcSetItemUUID sets the card's projected identity in the identity table.
func fcSetItemUUID(t *testing.T, root, cardID, value string) {
	t.Helper()
	dbPath := filepath.Join(filepath.Dir(todoBacklogPath(root)), "backlog.db")
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(dbPath))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(`UPDATE todo_identities SET uuid=? WHERE entity_kind='card' AND local_id=?`, value, cardID); err != nil {
		t.Fatalf("re-identity: %v", err)
	}
}

// AC-FCR-005 — the lock re-verification compares the full snapshot (UUID,
// body, state, SPEC, landing) and, when all five match, re-verifies the
// receipt's bindings at the archive moment inside the same lock. Every
// mismatch downgrades the planned close instead of archiving.
func TestAutoDoneRecheckStaleRow(t *testing.T) {
	root, store := autoDoneFixture(t)
	const text = "recheck target card"
	seedCard(t, store, "t100", text, factory.BacklogStatePicked)
	fcLinkRuntime(t, root, "t100")
	fcPlaceFactoryCard(t, root, "t100", 1, "sha-evidence", "2026-09-26T00:00:00Z")
	fcApprove(t, root, "t100", fcRun, 1, "sha-evidence")

	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	var item factory.BacklogItem
	for i := range rec.Items {
		if rec.Items[i].ID == "t100" {
			item = rec.Items[i]
			break
		}
	}
	snapshot := autoDoneOutcome{
		id: "t100", closed: true, form: factory.AutoDoneFormSubject,
		snapUUID:    todoCardUUID(&item),
		snapText:    item.Text,
		snapState:   item.State,
		snapSpec:    "",
		snapLanding: item.Landing,
	}

	t.Run("text edited under the lock", func(t *testing.T) {
		seed := freshRecheckRecord(t, store, "t100", text)
		if err := store.Mutate(func(r *factory.BacklogRecord) error {
			for i := range r.Items {
				if r.Items[i].ID == "t100" {
					r.Items[i].Text = text + " (edited)"
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		_ = seed
		out := runRecheck(t, store, root, snapshot)
		assertRecheckDowngraded(t, store, out, "t100", factory.AutoDoneSkipQueryInconclusive)
	})

	t.Run("state changed under the lock", func(t *testing.T) {
		if err := store.Mutate(func(r *factory.BacklogRecord) error {
			for i := range r.Items {
				if r.Items[i].ID == "t100" {
					r.Items[i].Text = text
					r.Items[i].State = factory.BacklogStateQueued
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		out := runRecheck(t, store, root, snapshot)
		assertRecheckDowngraded(t, store, out, "t100", factory.AutoDoneSkipQueryInconclusive)
	})

	t.Run("spec changed under the lock", func(t *testing.T) {
		if err := store.Mutate(func(r *factory.BacklogRecord) error {
			for i := range r.Items {
				if r.Items[i].ID == "t100" {
					r.Items[i].State = factory.BacklogStatePicked
					v := "SPEC-RECHECK-FIXTURE"
					r.Items[i].SpecID = &v
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		out := runRecheck(t, store, root, snapshot)
		assertRecheckDowngraded(t, store, out, "t100", factory.AutoDoneSkipQueryInconclusive)
	})

	t.Run("landing recorded under the lock", func(t *testing.T) {
		if err := store.Mutate(func(r *factory.BacklogRecord) error {
			for i := range r.Items {
				if r.Items[i].ID == "t100" {
					r.Items[i].SpecID = nil
					r.Items[i].Landing = &factory.LandingEvidence{
						Ref: "origin/develop", RefHead: "aa11", ObservedAt: "2026-09-26T08:00:00Z",
					}
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		out := runRecheck(t, store, root, snapshot)
		assertRecheckDowngraded(t, store, out, "t100", factory.AutoDoneSkipQueryInconclusive)
	})

	// Reset to the pristine row for the identity and factory-state variants.
	resetRecheckRow(t, store, "t100", text)

	t.Run("uuid re-identified under the lock", func(t *testing.T) {
		// The row now carries a fresh identity the snapshot never saw —
		// the UUID comparison fires before any receipt work.
		fresh := fcReidentityItemColumn(t, root, "t100")
		out := runRecheck(t, store, root, snapshot)
		assertRecheckDowngraded(t, store, out, "t100", factory.AutoDoneSkipQueryInconclusive)
		// Restore the snapshot's identity for the later subtests.
		fcSetItemUUID(t, root, "t100", snapshot.snapUUID)
		_ = fresh
	})

	t.Run("factory version bumped after approval", func(t *testing.T) {
		fcBumpRowVersion(t, root, "t100", 2)
		out := runRecheck(t, store, root, snapshot)
		assertRecheckDowngraded(t, store, out, "t100", factory.AutoDoneSkipLeaderUnapproved)
	})

	t.Run("run replaced after approval", func(t *testing.T) {
		// A re-dispatch updates the DISPATCH BINDING to a newer run whose
		// row carries the same uuid, version, and evidence — exactly the
		// cross-run shape REQ-FCR-001 forbids reusing. The receipt bound to
		// the original run no longer matches.
		fcBindDispatch(t, root, "t100", "run-late")
		db := fcOpen(t, root)
		defer func() { _ = db.Close() }()
		if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,owner_label,state,version,evidence_sha,updated_at,merge_sha) VALUES('run-late','t100','worker-1','picked',1,'sha-evidence','2026-09-26T10:00:00Z','sha-evidence')`); err != nil {
			t.Fatalf("place replacement run row: %v", err)
		}
		out := runRecheck(t, store, root, snapshot)
		assertRecheckDowngraded(t, store, out, "t100", factory.AutoDoneSkipLeaderUnapproved)
	})

	t.Run("matching snapshot and receipt closes", func(t *testing.T) {
		// Restore the pristine live row and factory state: restore the
		// dispatch binding, drop the replacement-run row, and restore
		// version 1.
		resetRecheckRow(t, store, "t100", text)
		fcBindDispatch(t, root, "t100", fcRun)
		db := fcOpen(t, root)
		if _, err := db.DB.Exec(`DELETE FROM cards WHERE run_id='run-late'`); err != nil {
			t.Fatal(err)
		}
		_ = db.Close()
		fcBumpRowVersion(t, root, "t100", 1)
		out := runRecheck(t, store, root, snapshot)
		if len(out) != 1 || !out[0].closed {
			t.Fatalf("matching snapshot did not close: %+v", out)
		}
	})
}

// freshRecheckRecord re-reads the live item (a no-op read kept for symmetry
// with the mutation subtests).
func freshRecheckRecord(t *testing.T, store *factory.BacklogStore, id, text string) *factory.BacklogItem {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for i := range rec.Items {
		if rec.Items[i].ID == id {
			return &rec.Items[i]
		}
	}
	t.Fatalf("card %s not found", id)
	return nil
}

// resetRecheckRow restores the pristine live row after a mutation subtest.
func resetRecheckRow(t *testing.T, store *factory.BacklogStore, id, text string) {
	t.Helper()
	if err := store.Mutate(func(r *factory.BacklogRecord) error {
		for i := range r.Items {
			if r.Items[i].ID == id {
				r.Items[i].Text = text
				r.Items[i].State = factory.BacklogStatePicked
				r.Items[i].SpecID = nil
				r.Items[i].Landing = nil
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// runRecheck drives applyAutoDoneCloses over one planned close.
func runRecheck(t *testing.T, store *factory.BacklogStore, root string, snapshot autoDoneOutcome) []autoDoneOutcome {
	t.Helper()
	outcomes := []autoDoneOutcome{snapshot}
	applied, err := applyAutoDoneCloses(context.Background(), root, store, outcomes)
	if err != nil {
		t.Fatalf("applyAutoDoneCloses: %v", err)
	}
	return applied
}

func assertRecheckDowngraded(t *testing.T, store *factory.BacklogStore, out []autoDoneOutcome, id, wantReason string) {
	t.Helper()
	if len(out) != 1 {
		t.Fatalf("outcomes = %d, want 1", len(out))
	}
	if out[0].closed {
		t.Fatalf("%s was archived despite the stale row", id)
	}
	if out[0].reason != wantReason {
		t.Fatalf("%s downgrade reason = %q, want %q", id, out[0].reason, wantReason)
	}
	if _, ok := liveItemOK(t, store, id); !ok {
		t.Fatalf("%s vanished from the live queue", id)
	}
}
