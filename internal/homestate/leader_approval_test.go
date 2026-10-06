package homestate

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The leader approval receipt (SPEC-FACTORY-COMPLETION-RECOVERY-001 M1,
// REQ-FCR-001/002/005). The receipt is the leader's evidence-review record
// bound to four values — card UUID, run id, factory version, evidence hash —
// and the done edges (T18, T20) verify it inside FactoryDB.Transition.

func frApprove(t *testing.T, db *FactoryDB, a LeaderApproval) {
	t.Helper()
	if _, err := db.IssueLeaderApproval(context.Background(), a); err != nil {
		t.Fatalf("issue approval for %s: %v", a.CardID, err)
	}
}

func TestLeaderApprovalBindingValidator(t *testing.T) {
	base := LeaderApproval{
		CardUUID: "uuid-1", RunID: "run-a", CardID: "t1", FactoryVersion: 3,
		EvidenceHash: "sha-a", Issuer: "lead", IssuerRole: ApprovalIssuerLeader,
		IssuedAt: "2026-10-06T00:00:00Z",
	}
	if err := base.VerifyBinding("uuid-1", "run-a", 3, "sha-a", "worker-1"); err != nil {
		t.Fatalf("matching binding refused: %v", err)
	}
	for name, mutate := range map[string]func(a *LeaderApproval){
		"uuid":      func(a *LeaderApproval) { a.CardUUID = "uuid-other" },
		"run":       func(a *LeaderApproval) { a.RunID = "run-other" },
		"stale":     func(a *LeaderApproval) { a.FactoryVersion = 2 },
		"hash":      func(a *LeaderApproval) { a.EvidenceHash = "sha-other" },
		"issuer":    func(a *LeaderApproval) { a.IssuerRole = ApprovalIssuerLane },
		"performer": func(a *LeaderApproval) { a.Issuer = "worker-1" },
	} {
		a := base
		mutate(&a)
		err := a.VerifyBinding("uuid-1", "run-a", 3, "sha-a", "worker-1")
		if err == nil {
			t.Fatalf("%s variant accepted", name)
		}
		var want error
		switch name {
		case "uuid":
			want = ErrApprovalCardMismatch
		case "run":
			want = ErrApprovalRunMismatch
		case "stale":
			want = ErrApprovalStale
		case "hash":
			want = ErrApprovalHashMismatch
		case "issuer", "performer":
			want = ErrApprovalIssuer
		}
		if !errors.Is(err, want) {
			t.Fatalf("%s variant err = %v, want %v", name, err, want)
		}
	}
	// The uuid axis is the caller's identity knowledge: a verifier that does
	// not know it (the factory row carries no uuid) checks the other three
	// and skips it — it does not fail it.
	row := base
	row.CardUUID = "uuid-1"
	if err := row.VerifyBinding("", "run-a", 3, "sha-a", "worker-1"); err != nil {
		t.Fatalf("uuid-agnostic verification refused: %v", err)
	}
	// A row that names no owner skips the performer axis too.
	ownerless := base
	ownerless.Issuer = "worker-1"
	if err := ownerless.VerifyBinding("uuid-1", "run-a", 3, "sha-a", ""); err != nil {
		t.Fatalf("ownerless row verification refused: %v", err)
	}
}

// Issuance is the leader path's act (REQ-FCR-014): a receipt minted with the
// performing lane's role marker is refused at issuance.
func TestLeaderApprovalIssuanceLeaderOnly(t *testing.T) {
	db := frOpen(t)
	_, err := db.IssueLeaderApproval(context.Background(), LeaderApproval{
		CardUUID: "uuid-1", RunID: frRun, CardID: "t1", FactoryVersion: 1,
		EvidenceHash: "sha-a", Issuer: "worker-1", IssuerRole: ApprovalIssuerLane,
	})
	if !errors.Is(err, ErrApprovalIssuer) {
		t.Fatalf("lane-role issuance err = %v, want ErrApprovalIssuer", err)
	}
	frApprove(t, db, LeaderApproval{
		CardUUID: "uuid-1", RunID: frRun, CardID: "t1", FactoryVersion: 1,
		EvidenceHash: "sha-a", Issuer: "lead", IssuerRole: ApprovalIssuerLeader,
	})
	got, err := db.FindLeaderApproval(context.Background(), "uuid-1")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.RunID != frRun || got.FactoryVersion != 1 || got.EvidenceHash != "sha-a" {
		t.Fatalf("stored receipt = %+v", got)
	}
	// Re-issuance at a new version replaces (the leader re-approves after a
	// transition bumped the card).
	frApprove(t, db, LeaderApproval{
		CardUUID: "uuid-1", RunID: frRun, CardID: "t1", FactoryVersion: 2,
		EvidenceHash: "sha-b", Issuer: "lead", IssuerRole: ApprovalIssuerLeader,
	})
	got, err = db.FindLeaderApproval(context.Background(), "uuid-1")
	if err != nil {
		t.Fatalf("find after re-issue: %v", err)
	}
	if got.FactoryVersion != 2 || got.EvidenceHash != "sha-b" {
		t.Fatalf("re-issued receipt = %+v, want the replacement", got)
	}
}

// HoldApprovalGate is the serialization point: a factory write transaction
// held across a backlog close's persistence, so the archive-moment
// verification cannot interleave with a factory transition. A card with no
// factory row is not factory-linked — the gate does not apply.
func TestApprovalGateLinkedAndUnlinked(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, false)
	ctx := context.Background()
	c := Card{RunID: frRun, CardID: "gated", State: CardMergedLocal, Version: 1, OwnerLabel: "worker-1", WorktreePath: repo.Dir, MergeSHA: repo.Merge, EvidenceSHA: repo.Commit}
	frPlace(t, db, c)
	frBindDispatch(t, db, "gated", frRun)

	gate, err := db.HoldApprovalGate(ctx)
	if err != nil {
		t.Fatalf("hold: %v", err)
	}
	// A card with no factory row reads as not linked — the caller skips
	// verification entirely and the close proceeds (the gate does not
	// apply; the CLI wrapper encodes exactly this branch).
	card, linked, err := gate.Row(ctx, "absent")
	if err != nil || linked {
		t.Fatalf("absent card: linked=%v err=%v, want not linked", linked, err)
	}
	if card.CardID != "" {
		t.Fatalf("absent card returned a row: %+v", card)
	}
	if err := gate.Rollback(); err != nil {
		t.Fatalf("rollback unlinked gate: %v", err)
	}

	// Issued receipt bound to another uuid: the card-uuid sentinel. Issued
	// BEFORE the gate is held — the gate holds the pool's single connection
	// for its whole life, so nothing else in this process can reach the
	// factory DB while it is open (the receipt is always issued before the
	// close it guards).
	frApprove(t, db, LeaderApproval{
		CardUUID: "uuid-other", RunID: frRun, CardID: "gated", FactoryVersion: 1,
		EvidenceHash: repo.Commit, Issuer: "lead", IssuerRole: ApprovalIssuerLeader,
	})
	gate, err = db.HoldApprovalGate(ctx)
	if err != nil {
		t.Fatalf("hold: %v", err)
	}
	defer func() {
		if err := gate.Rollback(); err != nil {
			t.Fatalf("rollback: %v", err)
		}
	}()
	card, linked, rowErr := gate.Row(ctx, "gated")
	if rowErr != nil || !linked {
		t.Fatalf("gated card: linked=%v err=%v, want linked", linked, rowErr)
	}
	if card.CardID != "gated" || card.Version != 1 || card.EvidenceSHA != repo.Commit {
		t.Fatalf("gate card = %+v", card)
	}
	// No receipt for THIS card's uuid: the missing sentinel.
	if err := gate.Verify(ctx, card, "uuid-1"); !errors.Is(err, ErrApprovalMissing) {
		t.Fatalf("verify without receipt err = %v, want ErrApprovalMissing", err)
	}
	// The receipt bound to the card's own uuid verifies — the gate's
	// positive control.
	if err := gate.Verify(ctx, card, "uuid-other"); err != nil {
		t.Fatalf("verify with the matching receipt: %v", err)
	}
	// The archive-moment row is what verification binds: a row read at
	// version 2 against the version-1 receipt is stale.
	bumped := card
	bumped.Version = 2
	if err := gate.Verify(ctx, bumped, "uuid-other"); !errors.Is(err, ErrApprovalStale) {
		t.Fatalf("verify against bumped row err = %v, want ErrApprovalStale", err)
	}
}

// T18 (merged-local → done, the no-remote edge) is admitted only by a leader
// approval receipt (REQ-FCR-002b): refused without one, admitted with the
// correct one. The row's guards above this gate are untouched.
func TestFR_FCR_T18DoneRequiresReceipt(t *testing.T) {
	db := frOpen(t)
	bare := frNewRepo(t, false)
	ctx := context.Background()
	c := Card{RunID: frRun, CardID: "noreceipt", State: CardMergedLocal, Version: 1, OwnerLabel: "worker-1", WorktreePath: bare.Dir, MergeSHA: bare.Merge, EvidenceSHA: bare.Commit}
	frPlace(t, db, c)
	req := TransitionRequest{RunID: frRun, CardID: c.CardID, To: CardDone, ExpectedVersion: c.Version, Actor: "lead", Decider: DeciderHuman, Now: frNow}
	before := frRowDump(t, db, frRun, c.CardID)
	if _, err := db.Transition(ctx, req); !errors.Is(err, ErrApprovalMissing) {
		t.Fatalf("done without receipt err = %v, want ErrApprovalMissing", err)
	}
	if after := frRowDump(t, db, frRun, c.CardID); after != before {
		t.Fatal("refused done changed the record")
	}
	frApprove(t, db, LeaderApproval{
		CardUUID: "uuid-nr", RunID: frRun, CardID: c.CardID, FactoryVersion: 1,
		EvidenceHash: bare.Commit, Issuer: "lead", IssuerRole: ApprovalIssuerLeader,
		IssuedAt: frNow.Format(time.RFC3339Nano),
	})
	got, err := db.Transition(ctx, req)
	if err != nil {
		t.Fatalf("done with receipt: %v", err)
	}
	if got.State != CardDone {
		t.Fatalf("done with receipt = %s, want done", got.State)
	}
	// A stale receipt (the card moved on) is refused: bump the version and
	// try again with the version-1 receipt still stored.
	frPlace(t, db, Card{RunID: frRun, CardID: "stale", State: CardMergedLocal, Version: 2, OwnerLabel: "worker-1", WorktreePath: bare.Dir, MergeSHA: bare.Merge, EvidenceSHA: bare.Commit})
	frApprove(t, db, LeaderApproval{
		CardUUID: "uuid-st", RunID: frRun, CardID: "stale", FactoryVersion: 1,
		EvidenceHash: bare.Commit, Issuer: "lead", IssuerRole: ApprovalIssuerLeader,
	})
	req2 := TransitionRequest{RunID: frRun, CardID: "stale", To: CardDone, ExpectedVersion: 2, Actor: "lead", Decider: DeciderHuman, Now: frNow}
	if _, err := db.Transition(ctx, req2); !errors.Is(err, ErrApprovalStale) {
		t.Fatalf("done with stale receipt err = %v, want ErrApprovalStale", err)
	}
}

// T20 (ci-green → done) is the reserved edge M1 admits — not by a CI reader
// (that opens T19 only, and is M2's), but by the receipt gate inside
// FactoryDB.Transition (REQ-FCR-002b, REQ-FCR-010). T19 stays reserved.
func TestFR_FCR_T20DoneRequiresReceipt(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()

	c := frFixtureCard(repo, "t20", CardCIGreen, CardDone)
	c.Version = 1
	frPlace(t, db, c)
	req := TransitionRequest{RunID: frRun, CardID: c.CardID, To: CardDone, ExpectedVersion: c.Version, Actor: "lead", Decider: DeciderHuman, Now: frNow}
	before := frRowDump(t, db, frRun, c.CardID)
	if _, err := db.Transition(ctx, req); !errors.Is(err, ErrApprovalMissing) {
		t.Fatalf("T20 without receipt err = %v, want ErrApprovalMissing", err)
	}
	if after := frRowDump(t, db, frRun, c.CardID); after != before {
		t.Fatal("refused T20 changed the record")
	}
	frApprove(t, db, LeaderApproval{
		CardUUID: "uuid-t20", RunID: frRun, CardID: c.CardID, FactoryVersion: 1,
		EvidenceHash: repo.Commit, Issuer: "lead", IssuerRole: ApprovalIssuerLeader,
	})
	got, err := db.Transition(ctx, req)
	if err != nil {
		t.Fatalf("T20 with receipt: %v", err)
	}
	if got.State != CardDone {
		t.Fatalf("T20 with receipt = %s, want done", got.State)
	}
}

// REQ-FCR-005's performer axis on the done transitions: a receipt whose
// Issuer equals the row's performing owner is refused even with the leader
// role marker — the overlay repro (issuer=worker-1, owner=worker-1 → done)
// must refuse on both T18 and the gate path.
func TestFR_FCR_PerformerOwnerRefusedOnTransition(t *testing.T) {
	db := frOpen(t)
	bare := frNewRepo(t, false)
	ctx := context.Background()
	c := Card{RunID: frRun, CardID: "self-approved", State: CardMergedLocal, Version: 1, OwnerLabel: "worker-1", WorktreePath: bare.Dir, MergeSHA: bare.Merge, EvidenceSHA: bare.Commit}
	frPlace(t, db, c)
	frApprove(t, db, LeaderApproval{
		CardUUID: "uuid-self", RunID: frRun, CardID: c.CardID, FactoryVersion: 1,
		EvidenceHash: bare.Commit, Issuer: "worker-1", IssuerRole: ApprovalIssuerLeader,
	})
	req := TransitionRequest{RunID: frRun, CardID: c.CardID, To: CardDone, ExpectedVersion: c.Version, Actor: "lead", Decider: DeciderHuman, Now: frNow}
	before := frRowDump(t, db, frRun, c.CardID)
	if _, err := db.Transition(ctx, req); !errors.Is(err, ErrApprovalIssuer) {
		t.Fatalf("self-issued T18 err = %v, want ErrApprovalIssuer", err)
	}
	if after := frRowDump(t, db, frRun, c.CardID); after != before {
		t.Fatal("the self-issued receipt moved the card")
	}
}

// TestFR_AC019_ReservedCIEdgesRefused narrows to the still-reserved pair:
// pushed → ci-green stays refused (the CI verdict reader that admits it is
// M2's, REQ-FCR-010). ci-green → done left the reserved set in M1 — the
// receipt gate admits it (TestFR_FCR_T20DoneRequiresReceipt).
func TestFR_AC019_ReservedCIEdgesRefused(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	claim := filepath.Join(t.TempDir(), "ci-green.json")
	frWrite(t, claim, `{"merge_sha":"`+repo.Merge+`","conclusion":"success"}`)
	for _, pair := range [][2]string{{CardPushed, CardCIGreen}} {
		cardID := "reserved-" + pair[0]
		c := frFixtureCard(repo, cardID, pair[0], pair[1])
		c.Version = 1
		frPlace(t, db, c)
		before := frRowDump(t, db, frRun, cardID)
		for _, withClaim := range []bool{false, true} {
			req := frFullRequest(repo, c, pair[1])
			if withClaim {
				req.ArtifactPath = claim
				req.RemeasurePath = claim
			}
			_, err := db.Transition(ctx, req)
			if !errors.Is(err, ErrReservedEdge) || !strings.Contains(err.Error(), "F3") {
				t.Fatalf("%s → %s (claim=%v) err = %v, want ErrReservedEdge naming F3", pair[0], pair[1], withClaim, err)
			}
			if after := frRowDump(t, db, frRun, cardID); after != before {
				t.Fatalf("%s → %s changed the record", pair[0], pair[1])
			}
		}
	}
}
