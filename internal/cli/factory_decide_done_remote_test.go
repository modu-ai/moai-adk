// factory_decide_done_remote_test.go — card t2059: the verb path that sends a
// merged-local card to done in a repository WITH a remote (t2057 admitted the
// T18 edge at the store; PushGateTarget still routes such a card to pushed).
package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// t2059Fixture places a merged-local card in a repository that has a remote,
// bound to its queue identity, and returns the pieces a decide test needs.
func t2059Fixture(t *testing.T, title string) (root, cardID, uuid, merge string) {
	t.Helper()
	root, store := fcFixture(t)
	cardID = addShapeCard(t, title)
	fcLinkRuntime(t, root, cardID)
	fcBindDispatch(t, root, cardID, fcRun)
	repo, merge := fcRepo(t, true)
	fcPlace(t, root, homestate.Card{CardID: cardID, RunID: fcRun, State: homestate.CardMergedLocal, OwnerLabel: "worker-1", Version: 1, MergeSHA: merge, WorktreePath: repo, EvidenceSHA: merge})
	uuid = recheckUUID(t, root, store, cardID)
	return root, cardID, uuid, merge
}

// TestT2059_DecidePushDoneWithRemoteNeedsTheReceipt — `--gate push --choice
// done` asks for the T18 edge on a merged-local card whatever the remote
// state: without the leader's receipt it is refused for the receipt (and the
// record does not move); with the receipt the card reaches done.
func TestT2059_DecidePushDoneWithRemoteNeedsTheReceipt(t *testing.T) {
	root, cardID, uuid, merge := t2059Fixture(t, "decide done with a remote")
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	if _, err := decideOne(ctx, db, root, fcRun, cardID, "push", "done", ""); !errors.Is(err, homestate.ErrApprovalMissing) {
		t.Fatalf("decide push done with a remote and no receipt: err = %v, want ErrApprovalMissing", err)
	}
	if c := fcCard(t, root, cardID); c.State != homestate.CardMergedLocal {
		t.Fatalf("the refused decide moved the card to %s", c.State)
	}

	fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
		CardUUID: uuid, RunID: fcRun, CardID: cardID, FactoryVersion: 1,
		EvidenceHash: merge, Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
	})
	got, err := decideOne(ctx, db, root, fcRun, cardID, "push", "done", "")
	if err != nil {
		t.Fatalf("decide push done with a remote and the receipt: %v", err)
	}
	if got.State != homestate.CardDone {
		t.Fatalf("decide = %s, want done", got.State)
	}
}

// TestT2059_DecideVerbAcceptsPushDone — the verb surface admits the choice
// (the validation switch listed only an empty choice for --gate push) and
// still refuses any other choice for that gate.
func TestT2059_DecideVerbAcceptsPushDone(t *testing.T) {
	t.Setenv(config.EnvFactoryRole, "operator")
	root, cardID, uuid, merge := t2059Fixture(t, "decide verb done with a remote")

	if _, _, err := runFactory(t, "decide", cardID, "--gate", "push", "--choice", "bogus", "--run", fcRun); err == nil {
		t.Fatal("--gate push with an unknown choice was accepted")
	}
	if _, _, err := runFactory(t, "decide", cardID, "--gate", "push", "--choice", "done", "--run", fcRun); err == nil {
		t.Fatal("--gate push --choice done without a receipt was accepted")
	}
	if c := fcCard(t, root, cardID); c.State != homestate.CardMergedLocal {
		t.Fatalf("the refused verb moved the card to %s", c.State)
	}

	fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
		CardUUID: uuid, RunID: fcRun, CardID: cardID, FactoryVersion: 1,
		EvidenceHash: merge, Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
	})
	if _, _, err := runFactory(t, "decide", cardID, "--gate", "push", "--choice", "done", "--run", fcRun); err != nil {
		t.Fatalf("--gate push --choice done with the receipt: %v", err)
	}
	if c := fcCard(t, root, cardID); c.State != homestate.CardDone {
		t.Fatalf("card = %s, want done", c.State)
	}
}
