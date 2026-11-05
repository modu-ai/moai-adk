// t2057_done_with_remote_test.go — card t2057: the merged-local → done edge
// (T18) is admitted while a remote is configured, and it still requires the
// leader's approval receipt (REQ-FCR-002b); no path to done exists without it.
package homestate

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// TestT2057_T18DoneWithRemoteRequiresReceipt — the operator's Done standard:
// local verification + the card's merge into the local integration branch +
// evidence closes the card, and a configured remote no longer refuses it. The
// leader's evidence receipt stays mandatory: without it the edge refuses with
// ErrApprovalMissing and the record does not change; with the correct receipt
// it lands at done.
func TestT2057_T18DoneWithRemoteRequiresReceipt(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	c := Card{RunID: frRun, CardID: "remote-done", State: CardMergedLocal, Version: 1, OwnerLabel: "worker-1", WorktreePath: repo.Dir, MergeSHA: repo.Merge, EvidenceSHA: repo.Commit}
	frPlace(t, db, c)
	req := TransitionRequest{RunID: frRun, CardID: c.CardID, To: CardDone, ExpectedVersion: c.Version, Actor: "lead", Decider: DeciderHuman, IntegrationBranch: repo.Integration, Now: frNow, ApprovalUUID: "uuid-remote-done"}

	before := frRowDump(t, db, frRun, c.CardID)
	if _, err := db.Transition(ctx, req); !errors.Is(err, ErrApprovalMissing) {
		t.Fatalf("done with a remote and no receipt: err = %v, want ErrApprovalMissing (the remote must no longer be the reason)", err)
	}
	if after := frRowDump(t, db, frRun, c.CardID); after != before {
		t.Fatal("the refused done changed the record")
	}

	frApprove(t, db, LeaderApproval{
		CardUUID: "uuid-remote-done", RunID: frRun, CardID: c.CardID, FactoryVersion: 1,
		EvidenceHash: repo.Commit, Issuer: "lead", IssuerRole: ApprovalIssuerLeader,
	})
	got, err := db.Transition(ctx, req)
	if err != nil {
		t.Fatalf("done with a remote and the leader's receipt: %v", err)
	}
	if got.State != CardDone {
		t.Fatalf("state = %s, want done", got.State)
	}
	events := frEvents(t, db, "card.transition")
	var payload map[string]any
	if err := json.Unmarshal([]byte(events[len(events)-1]), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["note"] != "leader approval verified" {
		t.Fatalf("event note = %v, want %q (the no-remote note must not be claimed while a remote exists)", payload["note"], "leader approval verified")
	}
}

// TestT2057_T18RemoteDoneRefusesAReceiptForAnotherCard — the receipt binding
// holds on the with-remote path too: a receipt issued under another identity
// for the same (run, card) pair does not open the edge.
func TestT2057_T18RemoteDoneRefusesAReceiptForAnotherCard(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	c := Card{RunID: frRun, CardID: "remote-other", State: CardMergedLocal, Version: 1, OwnerLabel: "worker-1", WorktreePath: repo.Dir, MergeSHA: repo.Merge, EvidenceSHA: repo.Commit}
	frPlace(t, db, c)
	frApprove(t, db, LeaderApproval{
		CardUUID: "uuid-someone-else", RunID: frRun, CardID: c.CardID, FactoryVersion: 1,
		EvidenceHash: repo.Commit, Issuer: "lead", IssuerRole: ApprovalIssuerLeader,
	})
	req := TransitionRequest{RunID: frRun, CardID: c.CardID, To: CardDone, ExpectedVersion: c.Version, Actor: "lead", Decider: DeciderHuman, IntegrationBranch: repo.Integration, Now: frNow, ApprovalUUID: "uuid-remote-other"}
	before := frRowDump(t, db, frRun, c.CardID)
	if _, err := db.Transition(ctx, req); !errors.Is(err, ErrApprovalCardMismatch) {
		t.Fatalf("done with a receipt bound to another identity: err = %v, want ErrApprovalCardMismatch", err)
	}
	if after := frRowDump(t, db, frRun, c.CardID); after != before {
		t.Fatal("the refused done changed the record")
	}
}
