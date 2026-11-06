// factory_complete_local_close_test.go — card t2057: the local integration
// route closes the lane's own queue card at merged-local, the same way the
// github-flow edge does at merged-pr.
package cli

import (
	"os"
	"path/filepath"
	"strings"
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
	root, integWT, cards := sdMergeFixture(t, true, true, false, 1)
	// A real --no-ff merge of the card branch into develop: the recorded
	// MergeSHA is reachable from the configured integration target.
	fcGit(t, integWT, "merge", "-q", "--no-ff", "-m", "M t1", cards[0].branch)
	merge := fcGit(t, integWT, "rev-parse", "HEAD")
	fcPlace(t, root, homestate.Card{
		CardID: "t1", State: homestate.CardMergedLocal, Stage: homestate.CardMergedLocal,
		OwnerLabel: "lane-1", LeaseHolder: "lane-1",
		WorktreePath: cards[0].wt, MergeSHA: merge,
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

// TestFactoryCompleteRerunDoesNotGuessTheLandingTarget — card-review P2: the
// card record keeps no integration branch, so the re-run reads the target from
// configuration. Where configuration names none, the target would fall back to
// the card's own work branch, and archiving the card as landed there would be
// a false landing verdict: the re-run reports and leaves the queue card live.
func TestFactoryCompleteRerunDoesNotGuessTheLandingTarget(t *testing.T) {
	sdClearLaneEnv(t)
	root, _, cards := sdMergeFixture(t, false, true, false, 1)
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	body := "git_strategy:\n  mode: manual\n  manual:\n    workflow: git-flow\n"
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	fcPlace(t, root, homestate.Card{
		CardID: "t1", State: homestate.CardMergedLocal, Stage: homestate.CardMergedLocal,
		OwnerLabel: "lane-1", LeaseHolder: "lane-1",
		WorktreePath: cards[0].wt, MergeSHA: "0123456789abcdef0123456789abcdef01234567",
	})
	sdLaneEnv(t, "lane-1", "")
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
	if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
		t.Fatalf("re-run of complete: %v", err)
	}
	rec, err := factory.NewBacklogStore(todoBacklogPath(root)).LoadPure()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	if len(rec.Archived) != 0 || len(rec.Items) != 1 {
		t.Fatalf("queue after the re-run: live=%d archived=%d, want the card left live (no configured landing target)", len(rec.Items), len(rec.Archived))
	}
}

// TestFactoryCompleteRerunRefusesAMergeNotOnTheTarget — card-review finding 3,
// redesigned (operator decision 11-06): the re-run closes the queue card only
// when the recorded MergeSHA is an ancestor of the CURRENT integration target;
// when it is not (the target configuration changed after the merge, or the
// merge never reached it), the re-run reports the reason, closes nothing and
// records no landed verdict.
func TestFactoryCompleteRerunRefusesAMergeNotOnTheTarget(t *testing.T) {
	sdClearLaneEnv(t)
	root, _, cards := sdMergeFixture(t, true, true, false, 1)
	// The card branch tip is NOT merged into develop: it is not an ancestor
	// of the configured target.
	notOnTarget := fcGit(t, cards[0].wt, "rev-parse", "HEAD")
	fcPlace(t, root, homestate.Card{
		CardID: "t1", State: homestate.CardMergedLocal, Stage: homestate.CardMergedLocal,
		OwnerLabel: "lane-1", LeaseHolder: "lane-1",
		WorktreePath: cards[0].wt, MergeSHA: notOnTarget,
	})
	before := fcCard(t, root, "t1")
	sdLaneEnv(t, "lane-1", "")
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
	stdout, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	if err != nil {
		t.Fatalf("re-run of complete: %v", err)
	}
	sdCardUnchanged(t, "unreachable merge re-run", root, "t1", before)
	rec, err := factory.NewBacklogStore(todoBacklogPath(root)).LoadPure()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	if len(rec.Archived) != 0 || len(rec.Items) != 1 {
		t.Fatalf("queue after the re-run: live=%d archived=%d, want the card left live (MergeSHA not on develop)", len(rec.Items), len(rec.Archived))
	}
	if !strings.Contains(stdout, "not reachable") {
		t.Errorf("the re-run output does not say why the card was left open: %q", stdout)
	}
}

// TestFactoryCompleteRerunReachabilityReadsTheBranchNotATag — card-review
// (c6452cde99): a tag named like the integration branch outranks the branch in
// git's name resolution, so a merge recorded only on that tag would read as
// reachable from "develop". The comparison names refs/heads/<branch>, and the
// re-run leaves the card open.
func TestFactoryCompleteRerunReachabilityReadsTheBranchNotATag(t *testing.T) {
	sdClearLaneEnv(t)
	root, _, cards := sdMergeFixture(t, true, true, false, 1)
	notOnTarget := fcGit(t, cards[0].wt, "rev-parse", "HEAD")
	// A tag "develop" on a commit the branch does not contain.
	fcGit(t, root, "tag", "develop", notOnTarget)
	fcPlace(t, root, homestate.Card{
		CardID: "t1", State: homestate.CardMergedLocal, Stage: homestate.CardMergedLocal,
		OwnerLabel: "lane-1", LeaseHolder: "lane-1",
		WorktreePath: cards[0].wt, MergeSHA: notOnTarget,
	})
	sdLaneEnv(t, "lane-1", "")
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
	if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
		t.Fatalf("re-run of complete: %v", err)
	}
	rec, err := factory.NewBacklogStore(todoBacklogPath(root)).LoadPure()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	if len(rec.Archived) != 0 || len(rec.Items) != 1 {
		t.Fatalf("queue after the re-run: live=%d archived=%d, want the card left live (the tag, not the branch, contains the merge)", len(rec.Items), len(rec.Archived))
	}
}
