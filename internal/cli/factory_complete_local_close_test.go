// factory_complete_local_close_test.go — card t2057: the local integration
// route closes the lane's own queue card at merged-local, the same way the
// github-flow edge does at merged-pr.
package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestFactoryCompleteLocalMergeClosesTheQueueCard — the operator's Done
// standard is local verification + the card's merge into the local
// integration branch + recorded evidence. The github-flow edge already closes
// the lane's own queue card at merged-pr (factoryCloseLaneCard); the local
// integration-window route must close it the same way once the card reaches
// merged-local, with the landing verdict naming the integration branch. The
// approval-gated done edges (T18/T20) are a separate, untouched surface.
func TestFactoryCompleteLocalMergeClosesTheQueueCard(t *testing.T) {
	sdClearLaneEnv(t)
	root, integWT, cards := sdMergeFixture(t, true, true, false, 1)
	sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
	sdHoldWindow(t, root, "sess-lane-1", "lane-1", "develop", factory.BranchSourceConfig, integWT, "t1")
	if _, remErr := factory.RunRemeasure(root, cards[0].wt, "develop", "true"); remErr != nil {
		t.Fatalf("remeasure on the candidate tree: %v", remErr)
	}
	sdLaneEnv(t, "lane-1", "")
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
	if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if c := fcCard(t, root, "t1"); c.State != homestate.CardMergedLocal {
		t.Fatalf("t1 = %s, want merged-local", c.State)
	}
	rec, err := factory.NewBacklogStore(todoBacklogPath(root)).LoadPure()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	if len(rec.Items) != 0 {
		t.Fatalf("the queue still holds %d live card(s) after merged-local", len(rec.Items))
	}
	if len(rec.Archived) != 1 || rec.Archived[0].Item.ID != "t1" {
		t.Fatalf("archive holds %d entries, want exactly t1", len(rec.Archived))
	}
	if v := rec.Archived[0].LandingVerdict; v == nil || v.Verdict != factory.LandingLanded || v.Ref != "develop" {
		t.Fatalf("landing verdict = %+v, want landed against develop", v)
	}
}

// TestFactoryCompleteRerunReconcilesAMergedLocalCard — closing the queue card
// is fail-open (a failed archive prints a note), so a re-run of complete on a
// card already at merged-local must retry the close instead of refusing the
// card as "not merge-ready"; the pull-request edge reconciles the same way at
// merged-pr. The card record and the integration branch are not touched.
func TestFactoryCompleteRerunReconcilesAMergedLocalCard(t *testing.T) {
	sdClearLaneEnv(t)
	root, _, cards := sdMergeFixture(t, true, true, false, 1)
	fcPlace(t, root, homestate.Card{
		CardID: "t1", State: homestate.CardMergedLocal, Stage: homestate.CardMergedLocal,
		OwnerLabel: "lane-1", LeaseHolder: "lane-1",
		WorktreePath: cards[0].wt, MergeSHA: "0123456789abcdef0123456789abcdef01234567",
	})
	before := fcCard(t, root, "t1")
	sdLaneEnv(t, "lane-1", "")
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
	if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
		t.Fatalf("re-run of complete on a merged-local card: %v", err)
	}
	sdCardUnchanged(t, "merged-local re-run", root, "t1", before)
	rec, err := factory.NewBacklogStore(todoBacklogPath(root)).LoadPure()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	if len(rec.Items) != 0 || len(rec.Archived) != 1 || rec.Archived[0].Item.ID != "t1" {
		t.Fatalf("queue after the re-run: live=%d archived=%d, want t1 archived", len(rec.Items), len(rec.Archived))
	}
}
