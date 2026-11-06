// factory_push_gate_default_test.go — card t2059 (root cause): the push gate's
// DEFAULT destination for a merged-local card no longer depends on whether the
// repository has a remote. The operator's Done standard is local verification
// + the merge into local main + the leader's evidence receipt; the pushed →
// ci-green → done route (T17/T19/T20) is an explicit choice.
package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// t2059PushGateCard places a merged-local card in a repository WITH a remote
// whose integration branch already carries the merge (so the explicit push
// route has its evidence), and returns the pieces the tests need.
func t2059PushGateCard(t *testing.T, title string) (root, cardID, uuid, merge string) {
	t.Helper()
	root, store := fcFixture(t)
	fcGitFlowConfig(t, root)
	dir, merge := fcRepo(t, true)
	cardID = addShapeCard(t, title)
	fcGit(t, dir, "push", "-q", "origin", "integration")
	fcLinkRuntime(t, root, cardID)
	fcBindDispatch(t, root, cardID, fcRun)
	fcPlace(t, root, homestate.Card{CardID: cardID, RunID: fcRun, State: homestate.CardMergedLocal, OwnerLabel: "worker-1", Version: 1, MergeSHA: merge, WorktreePath: dir, EvidenceSHA: merge})
	uuid = recheckUUID(t, root, store, cardID)
	return root, cardID, uuid, merge
}

// TestT2059_PushGateDefaultsToDoneWithARemote — the default decision for a
// merged-local card with a remote is done, not pushed: with the leader's
// receipt the card reaches done, and the pushed route is not taken.
func TestT2059_PushGateDefaultsToDoneWithARemote(t *testing.T) {
	t.Setenv(config.EnvFactoryRole, "operator")
	root, cardID, uuid, merge := t2059PushGateCard(t, "push gate default with a remote")
	fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
		CardUUID: uuid, RunID: fcRun, CardID: cardID, FactoryVersion: 1,
		EvidenceHash: merge, Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
	})
	if _, _, err := runFactory(t, "decide", cardID, "--gate", "push", "--run", fcRun); err != nil {
		t.Fatalf("default push gate with a remote and the receipt: %v", err)
	}
	if c := fcCard(t, root, cardID); c.State != homestate.CardDone {
		t.Fatalf("card = %s, want done (the remote must not decide the default destination)", c.State)
	}
}

// TestT2059_PushGateDefaultNeverFallsBackToPushedWithoutAReceipt — without the
// receipt the default refuses for the receipt and the card does not move; it
// must not slip onto the pushed route (which needs no receipt) on its own.
func TestT2059_PushGateDefaultNeverFallsBackToPushedWithoutAReceipt(t *testing.T) {
	t.Setenv(config.EnvFactoryRole, "operator")
	root, cardID, _, _ := t2059PushGateCard(t, "push gate default without a receipt")
	if _, _, err := runFactory(t, "decide", cardID, "--gate", "push", "--run", fcRun); err == nil {
		t.Fatal("the default push gate accepted a merged-local card with no receipt")
	}
	if c := fcCard(t, root, cardID); c.State != homestate.CardMergedLocal {
		t.Fatalf("card = %s, want merged-local (the default must not take the pushed route)", c.State)
	}
}

// TestT2059_PushGateExplicitPushChoiceTakesThePushedRoute — the pushed →
// ci-green → done route stays available as an explicit choice, and the receipt
// still chains across T17 (the version bump re-stamps the receipt).
func TestT2059_PushGateExplicitPushChoiceTakesThePushedRoute(t *testing.T) {
	t.Setenv(config.EnvFactoryRole, "operator")
	root, cardID, uuid, merge := t2059PushGateCard(t, "push gate explicit push")
	fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
		CardUUID: uuid, RunID: fcRun, CardID: cardID, FactoryVersion: 1,
		EvidenceHash: merge, Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
	})
	if _, _, err := runFactory(t, "decide", cardID, "--gate", "push", "--choice", "push", "--run", fcRun); err != nil {
		t.Fatalf("decide --gate push --choice push: %v", err)
	}
	if c := fcCard(t, root, cardID); c.State != homestate.CardPushed {
		t.Fatalf("card = %s, want pushed", c.State)
	}
}
