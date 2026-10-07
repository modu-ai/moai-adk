package cli

// factory_complete_mwq19_test.go — M6 tests for the substantive completion
// gate (card t1479, AC-MWQ-019 and O1): the one merge path (complete calls
// the REQ-MWQ-017 step), the re-measure-and-re-acquire refusal on a moved
// base, the post-merge transition conflict (commit left, hold naming the
// SHA, complete's own exit code), and O1's foreign-holder invariant. Every
// fixture is fresh per case (codex-P2), and the lane environment is set
// deliberately — complete is a lane command that ADMITS on the role claim.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// mwq19Fixture builds the complete environment: parent checkout with the
// git-flow develop branch provisioned in an integration worktree, one card
// worktree on its WT- branch, the card recorded merge-ready with a LIVE
// lease held by lane-1, and a valid re-measure record keyed to the
// candidate tree. The card id is t1 on run-cli, the same shapes the
// REQ-SD-013 fixtures use.
func mwq19Fixture(t *testing.T) (root, integ string, cardWT sdCardTree) {
	t.Helper()
	// The lane env leaks into tests from the lane session running them
	// (the recorded lesson: env-reading guards go locally red here) — the
	// fixture clears what would refuse its setup, then pins what the lane
	// verbs admit on.
	t.Setenv(config.EnvFactoryRole, "")
	t.Setenv(config.EnvMoaiFactoryWorker, "")
	t.Setenv(config.EnvFactoryBackend, "")
	root, integ, cards := sdMergeFixture(t, true, true, false, 1)
	cardWT = cards[0]
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
	sdLaneEnv(t, "lane-1", "")

	// The card: merge-ready with a live lease (the fixture helper stamps a
	// lapsed lease, so it is placed directly).
	fcPlace(t, root, homestate.Card{
		CardID: "t1", RunID: fcRun,
		State: homestate.CardMergeReady, Stage: homestate.CardMergeReady,
		OwnerLabel: "lane-1", LeaseHolder: "lane-1",
		LeaseExpiresAt: time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
		HeartbeatAt:    time.Now().UTC().Format(time.RFC3339),
		WorktreePath:   cardWT.wt, Version: 1,
	})

	// The re-measure: keyed to the candidate tree, valid (a tool with no
	// recognized report is valid on exit zero).
	if _, err := factory.RunRemeasure(root, cardWT.wt, "develop", "true"); err != nil {
		t.Fatalf("remeasure: %v", err)
	}
	return root, integ, cardWT
}

func TestMWQ19_Scenario2_OneMergePath(t *testing.T) {
	// AC-MWQ-019 scenario 2: a valid record and an unmoved base — the merge
	// runs through the merge step (one --no-ff merge commit whose second
	// parent is the card tip), the card lands merged-local, and the window
	// releases (complete's deferred release).
	root, integ, cardWT := mwq19Fixture(t)
	_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	if err != nil {
		t.Fatalf("complete must succeed: %v", err)
	}
	card := fcCard(t, root, "t1")
	if card.State != homestate.CardMergedLocal {
		t.Fatalf("the card must land merged-local, got %q", card.State)
	}
	if card.MergeSHA == "" {
		t.Fatalf("the merged-local record must carry the merge SHA")
	}
	parents := strings.Fields(fcGit(t, integ, "rev-list", "--parents", "-n", "1", card.MergeSHA))
	cardTip := fcGit(t, cardWT.wt, "rev-parse", "HEAD")
	if len(parents) != 3 || parents[2] != cardTip {
		t.Fatalf("the merge commit's second parent must be the card tip: %v (card tip %s)", parents, cardTip)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("the window must be released after the transitions: %+v", lock)
	}
	// The integration worktree is clean (no residue).
	if status := fcGit(t, integ, "status", "--porcelain", "--untracked-files=all"); status != "" {
		t.Fatalf("the integration worktree must be clean: %q", status)
	}
}

func TestMWQ19_Scenario3_MovedBaseRefusesWithReacquireCode(t *testing.T) {
	// AC-MWQ-019 scenario 3: develop advanced past the record's base — no
	// merge commit, everything unchanged, exit code = the
	// re-measure-and-re-acquire code (factory.MergeExitBaseMoved = 2).
	root, integ, cardWT := mwq19Fixture(t)
	before := fcCard(t, root, "t1")
	beforeLock := sdWindow(t, root)
	// Another lane lands on develop.
	fcGit(t, integ, "commit", "-q", "--allow-empty", "-m", "another lane landed")
	_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	if err == nil {
		t.Fatalf("a moved base must refuse")
	}
	if !strings.Contains(err.Error(), "re-acquire") {
		t.Fatalf("the refusal must name the re-acquire path: %v", err)
	}
	sdCardUnchanged(t, "scenario 3", root, "t1", before)
	after := sdWindow(t, root)
	if after.SessionID != beforeLock.SessionID || after.Branch != beforeLock.Branch {
		t.Fatalf("the integration branch must be unchanged: %+v vs %+v", beforeLock, after)
	}
	_ = cardWT
}

func TestMWQ19_Scenario8_PostMergeTransitionConflict(t *testing.T) {
	// AC-MWQ-019 scenario 8: the card's version bumps after the merge step
	// succeeds — the merge commit STAYS on the integration branch, the card
	// state is unchanged, the policy holds with cause
	// post-merge-transition-conflict naming the merge SHA, and the exit
	// code differs from the thirteen (complete's own).
	root, integ, _ := mwq19Fixture(t)
	db := fcOpen(t, root)
	factoryCompleteTransitionHook = func() {
		res, err := db.DB.Exec("UPDATE cards SET version = version + 1 WHERE card_id = 't1' AND run_id = ?", fcRun)
		if err != nil {
			t.Errorf("hook bump: %v", err)
			return
		}
		if n, _ := res.RowsAffected(); n != 1 {
			t.Errorf("hook bump touched %d rows, want 1 (the fixture's card row must exist under run-cli/t1)", n)
		}
	}
	t.Cleanup(func() { factoryCompleteTransitionHook = nil })

	_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	if err == nil {
		t.Fatalf("a post-merge transition conflict must fail the run")
	}
	// The merge commit stays on the integration branch.
	head := fcGit(t, integ, "rev-parse", "HEAD")
	parents := strings.Fields(fcGit(t, integ, "rev-list", "--parents", "-n", "1", head))
	if len(parents) != 3 {
		t.Fatalf("the merge commit must stay on the integration branch: %v", parents)
	}
	if !strings.Contains(err.Error(), head[:12]) {
		t.Fatalf("the failure must name the merge SHA %s: %v", head[:12], err)
	}
	// The card state is unchanged (still merge-ready, version bumped by the
	// hook).
	card := fcCard(t, root, "t1")
	if card.State != homestate.CardMergeReady {
		t.Fatalf("the card state must not change: %q", card.State)
	}
	// The policy holds naming the cause and the SHA.
	policy, err := factory.ReadIntegrationWindowPolicy(root)
	if err != nil || policy.Policy != factory.PolicyHold {
		t.Fatalf("the policy must hold after the conflict: %+v err=%v", policy, err)
	}
	if !strings.Contains(policy.Reason, "post-merge-transition-conflict") || !strings.Contains(policy.Reason, head[:12]) {
		t.Fatalf("the hold must name the cause and the merge SHA: %+v", policy)
	}
	// No queued ticket is promoted onto the conflict (there is none here;
	// the holder is gone after the release).
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("the window must be released after the hold: %+v", lock)
	}
}

func TestMWQ19_O1_ForeignHolderRecordUntouched(t *testing.T) {
	// O1: a window held by ANOTHER session is never released or altered by
	// a complete that refuses — the foreign holder's record bytes are
	// unchanged (and the refusal names the holder, not a release).
	root, _, _ := mwq19Fixture(t)
	// A foreign holder takes the window AFTER the fixture (lane-1's own
	// window would be reused).
	if _, err := factory.AcquireIntegrationWindow(root, factory.IntegrationLock{
		SessionID: "sess-foreign", SessionName: "lane-9",
		Branch: "develop", BranchSource: "config", Worktree: "/repo/integ", Card: "other-card",
	}, false, nil); err != nil {
		t.Fatal(err)
	}
	before := sdWindow(t, root)
	beforeBytes := fmt.Sprint(before)
	_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	if err == nil {
		t.Fatalf("a foreign-held window must refuse the complete")
	}
	if !strings.Contains(err.Error(), "lane-9") {
		t.Fatalf("the refusal must name the holder: %v", err)
	}
	after := sdWindow(t, root)
	if fmt.Sprint(after) != beforeBytes {
		t.Fatalf("the foreign holder's record must be byte-shape unchanged: %+v vs %+v", before, after)
	}
}

func TestMWQ19_Scenario5_AdoptionRefusedAfterNewCommit(t *testing.T) {
	// AC-MWQ-019 scenario 5 (fixture — verb, new commit, complete): the lane
	// merged through the verb, then committed once more on its WT- branch —
	// the earlier merge is NOT adopted (its second parent is no longer the
	// branch tip); with no valid record for the NEW tip complete refuses,
	// and no merge commit is created.
	root, integ, cardWT := mwq19Fixture(t)
	// The lane holds the window first — the verb merges as the holder.
	sdHoldWindow(t, root, "sess-lane-1", "lane-1", "develop", factory.BranchSourceConfig, integ, "t1")
	// The lane merges through the merge verb (the real one — holder, valid
	// record).
	if _, _, err := runIntegrationMerge(t, "merge", "--card", "t1", "--run", fcRun); err != nil {
		t.Fatalf("the verb merge must succeed: %v", err)
	}
	// One more commit on the card branch.
	if err := os.WriteFile(filepath.Join(cardWT.wt, "late.txt"), []byte("late\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, cardWT.wt, "add", "late.txt")
	fcGit(t, cardWT.wt, "commit", "-q", "-m", "late card commit")
	// The card is still merge-ready (the fixture's lease still holds).
	before := fcCard(t, root, "t1")
	beforeTip := fcGit(t, integ, "rev-parse", "develop")
	_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	if err == nil {
		t.Fatalf("an adoption after a new card commit must refuse")
	}
	if tip := fcGit(t, integ, "rev-parse", "develop"); tip != beforeTip {
		t.Fatalf("the integration branch must be unchanged: %s vs %s", beforeTip, tip)
	}
	after := fcCard(t, root, "t1")
	if after.State != before.State {
		t.Fatalf("the card state must be unchanged: %q vs %q", after.State, before.State)
	}
}

func TestMWQ19_Step3RefusalReleasesCompleteOwnAcquisition(t *testing.T) {
	// t1576 review round 2: on the path where the lane does not hold the
	// window yet, complete ACQUIRES it itself before the flow reaches step
	// 3 — and the step-3 refusal (no valid re-measure record for the
	// candidate tree) returned with the window still held, parking the next
	// lane until the lease lapsed. The acquisition complete made in this
	// invocation is released by the refusal.
	root, _, cardWT := mwq19Fixture(t)
	// Invalidate the record by gaining a commit: the candidate tree changes
	// and no record keys the new tree (the scenario-5 shape without a prior
	// verb merge, so no adoption applies and the lane holds no window).
	if err := os.WriteFile(filepath.Join(cardWT.wt, "late.txt"), []byte("late\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, cardWT.wt, "add", "late.txt")
	fcGit(t, cardWT.wt, "commit", "-q", "-m", "late card commit")
	_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	if err == nil {
		t.Fatalf("a complete with no record for the candidate tree must refuse")
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("complete's own acquisition must be released by the step-3 refusal: %+v", lock)
	}
}

// runIntegrationMerge runs the integration merge verb in-process.
func runIntegrationMerge(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := newIntegrationMergeCmd()
	var out, errBuf strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errBuf.String(), err
}
