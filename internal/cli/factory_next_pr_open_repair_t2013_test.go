// factory_next_pr_open_repair_t2013_test.go — card t2013 (R200): after the
// completion path leaves a card at `pr-open`, the registered owner lane resumes
// repair with `moai factory next --card <id>`: the card is leased to the owner
// again at the run stage, keeping its worktree and evidence, in one record
// transaction that never touches the worktree. Another lane stays refused
// exactly as today.
package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// ASSUMPTION: the queue item is taken to stay `picked` after the PR-opening
// step. That was read from factory_card_pr.go (the PR edge writes no queue
// state) and NOT traced across every closure path; a path that moves the item
// elsewhere would make the nominated lease refuse the owner on the queue-state
// check before the record is consulted.
func prOpenRepairSetup(t *testing.T) reacqFixture {
	t.Helper()
	root, store := nmBase(t, factory.BacklogStatePicked)
	wt := reacqWorktree(t)
	fcPlace(t, root, homestate.Card{
		CardID: "t1", State: homestate.CardPROpen, OwnerLabel: "lane-1",
		Stage: homestate.CardMergeReady, Version: 5,
		WorktreePath: wt, EvidenceSHA: reacqEvidenceSHA, EvidencePath: reacqEvidencePath,
	})
	return reacqFixture{root: root, store: store, wt: wt}
}

func TestT2013PROpenOwnerNextResumesRepair(t *testing.T) {
	fx := prOpenRepairSetup(t)
	nmLaneEnv(t, "lane-1", "")
	before := fcCard(t, fx.root, "t1")
	snapshot := reacqWorktreeSnapshot(t, fx.wt)

	out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	if err != nil {
		t.Fatalf("owner next --card t1 on its pr-open card: %v (stderr %q), want exit 0", err, stderr)
	}
	if nmLeasedHead(out) == "" {
		t.Errorf("stdout = %q, want the leased card's head line", out)
	}
	after := fcCard(t, fx.root, "t1")
	if after.State != homestate.CardLeased || after.LeaseHolder != "lane-1" || after.OwnerLabel != "lane-1" {
		t.Fatalf("t1 = %s owner=%q holder=%q, want leased to lane-1", after.State, after.OwnerLabel, after.LeaseHolder)
	}
	if after.Stage != homestate.CardRun {
		t.Errorf("stage = %q, want %q", after.Stage, homestate.CardRun)
	}
	if after.WorktreePath != before.WorktreePath || after.EvidenceSHA != before.EvidenceSHA || after.EvidencePath != before.EvidencePath {
		t.Errorf("worktree/evidence changed: %q %q %q → %q %q %q",
			before.WorktreePath, before.EvidenceSHA, before.EvidencePath, after.WorktreePath, after.EvidenceSHA, after.EvidencePath)
	}
	if got := reacqWorktreeSnapshot(t, fx.wt); got != snapshot {
		t.Errorf("the verb touched the worktree:\nbefore:\n%s\nafter:\n%s", snapshot, got)
	}
	if _, n := reacqIndexOf(reacqEventKinds(t, fx.root), "pr.repair_resumed"); n != 1 {
		t.Errorf("pr.repair_resumed events = %d, want exactly 1", n)
	}
}

func TestT2013PROpenOtherLaneNextStaysRefused(t *testing.T) {
	fx := prOpenRepairSetup(t)
	nmLaneEnv(t, "lane-2", "")
	rowsBefore := fcRows(t, fx.root)

	out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	nmAssertRefused(t, out, stderr, err, "recorded")
	if after := fcRows(t, fx.root); after != rowsBefore {
		t.Errorf("a refused nomination changed the rows or events:\nbefore:\n%s\nafter:\n%s", rowsBefore, after)
	}
}
