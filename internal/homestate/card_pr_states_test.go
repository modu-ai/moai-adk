package homestate

// card_pr_states_test.go — SPEC-GITHUB-FLOW-DEFAULT-001 M2-B (card t1453):
// the github-flow delivery states `pr-open` and `merged-pr` (design D-4, S-a)
// and the three state consumers the SPEC names — the predecessor pass
// (card_picked.go), its refusal wording, and the merged-local → done refusal
// (card_transition.go) a PR-merged card must not meet.

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const prCardBranch = "WT-pr-card"

// prRepo is frNewRepo with the worktree moved onto a card branch that carries
// one commit past the integration branch and is pushed to origin. HEAD is the
// card tip and origin/WT-pr-card equals it. The fixture's `integration` branch
// (pushed, carrying merge M) stands for the PR's base.
func prRepo(t *testing.T) frRepo {
	t.Helper()
	repo := frNewRepo(t, true)
	frGit(t, repo.Dir, "checkout", "-q", "-b", prCardBranch)
	frWrite(t, filepath.Join(repo.Dir, "c.txt"), "card\n")
	frGit(t, repo.Dir, "add", "-A")
	frGit(t, repo.Dir, "commit", "-q", "-m", "card change")
	frGit(t, repo.Dir, "push", "-q", "origin", prCardBranch)
	return repo
}

func prTip(t *testing.T, repo frRepo) string {
	t.Helper()
	return frGit(t, repo.Dir, "rev-parse", "HEAD")
}

// prCard places a card in state with the worktree and lease a lane holds.
func prCard(t *testing.T, db *FactoryDB, repo frRepo, cardID, state string) {
	t.Helper()
	c := Card{RunID: frRun, CardID: cardID, State: state, OwnerLabel: "worker-1", WorktreePath: repo.Dir, EvidenceSHA: repo.Commit}
	if IsLeaseHoldingState(state) {
		c.LeaseHolder = "worker-1"
		c.LeaseExpiresAt = frLeaseUntil(time.Hour)
		c.HeartbeatAt = frNow.Format(time.RFC3339Nano)
	}
	frPlace(t, db, c)
}

func prOpenRequest(cardID string, repo frRepo) TransitionRequest {
	return TransitionRequest{
		RunID: frRun, CardID: cardID, To: CardPROpen, ExpectedVersion: 1, Actor: "worker-1",
		IntegrationBranch: repo.Integration, PRNumber: "7", PRURL: "https://github.example/org/repo/pull/7", Now: frNow,
	}
}

// AC-GFD-004 — the homestate half: the two state consumers the SPEC names
// (design D-4, consumers (1)+(2) and (3)).
func TestFactoryCompleteGitHubFlowPR(t *testing.T) {
	t.Run("predecessor_merged_by_pr_does_not_block_the_next_card", func(t *testing.T) {
		db := frOpen(t)
		frRegisterWorker(t, db, "worker-1")
		ctx := context.Background()
		frPlace(t, db, Card{RunID: frRun, CardID: "pred", State: CardMergedPR, Version: 1})
		frPlace(t, db, Card{RunID: frRun, CardID: "succ", State: CardPicked, Version: 1, HintAfter: "pred"})
		got, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "succ", To: CardAssigned, ExpectedVersion: 1, Actor: "assign", Owner: "worker-1", Now: frNow})
		if err != nil {
			t.Fatalf("a card behind a merged-pr predecessor was refused: %v", err)
		}
		if got.State != CardAssigned {
			t.Fatalf("successor = %s, want assigned", got.State)
		}
	})

	t.Run("merged_pr_reaches_done_with_a_remote_configured", func(t *testing.T) {
		db := frOpen(t)
		repo := frNewRepo(t, true) // the repository has a remote: merged-local → done is refused here
		ctx := context.Background()
		frPlace(t, db, Card{RunID: frRun, CardID: "viaPR", State: CardMergedPR, Version: 1, WorktreePath: repo.Dir, MergeSHA: repo.Merge})
		done, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "viaPR", To: CardDone, ExpectedVersion: 1, Actor: "operator", Decider: DeciderHuman, Now: frNow})
		if err != nil {
			t.Fatalf("merged-pr → done with a remote configured: %v", err)
		}
		if done.State != CardDone {
			t.Fatalf("state = %s, want done", done.State)
		}
		// The control: the git-flow state keeps its refusal.
		frPlace(t, db, Card{RunID: frRun, CardID: "viaLocal", State: CardMergedLocal, Version: 1, WorktreePath: repo.Dir, MergeSHA: repo.Merge})
		if _, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "viaLocal", To: CardDone, ExpectedVersion: 1, Actor: "operator", Decider: DeciderHuman, Now: frNow}); !errors.Is(err, ErrIllegalTransition) {
			t.Fatalf("merged-local → done with a remote: err = %v, want ErrIllegalTransition (git-flow unchanged)", err)
		}
	})
}

func TestPRStatesAreInTheClosedStateSet(t *testing.T) {
	for _, s := range []string{CardPROpen, CardMergedPR} {
		if !IsCardState(s) {
			t.Errorf("%q is not an F1 card state", s)
		}
	}
	if CardPROpen != "pr-open" || CardMergedPR != "merged-pr" {
		t.Fatalf("state names = %q, %q", CardPROpen, CardMergedPR)
	}
	if got := len(CardStates()); got != 21 {
		t.Fatalf("CardStates = %d, want 21 (19 + pr-open + merged-pr)", got)
	}
	if IsLeaseHoldingState(CardPROpen) || IsLeaseHoldingState(CardMergedPR) {
		t.Error("pr-open and merged-pr hold no lease: the lane is free once the PR is open")
	}
	if IsTerminalCardState(CardPROpen) || IsTerminalCardState(CardMergedPR) {
		t.Error("neither PR state is terminal")
	}
	if (Card{State: CardMergedLocal}).Legacy() || (Card{State: CardPROpen}).Legacy() || (Card{State: CardMergedPR}).Legacy() {
		t.Error("merged-local stays a current state and neither PR state reads as a legacy row")
	}
}

// The merging → pr-open edge: the lease holder's, and it reads the git
// evidence itself — the card tip must be the pushed tip.
func TestPROpenEdgeReadsThePushedTip(t *testing.T) {
	ctx := context.Background()
	t.Run("pushed_tip_with_a_pr_identity_is_accepted_and_releases_the_lease", func(t *testing.T) {
		db := frOpen(t)
		repo := prRepo(t)
		prCard(t, db, repo, "c1", CardMerging)
		got, err := db.Transition(ctx, prOpenRequest("c1", repo))
		if err != nil {
			t.Fatalf("merging → pr-open: %v", err)
		}
		if got.State != CardPROpen || got.LeaseHolder != "" || got.LeaseExpiresAt != "" {
			t.Fatalf("card = %s holder=%q expiry=%q, want pr-open with no lease", got.State, got.LeaseHolder, got.LeaseExpiresAt)
		}
		ev := frEvents(t, db, "card.transition")
		if len(ev) != 1 || !strings.Contains(ev[0], `"pr_number":"7"`) || !strings.Contains(ev[0], "pull/7") || !strings.Contains(ev[0], prTip(t, repo)) {
			t.Fatalf("transition event = %v, want the PR identity and the pushed tip", ev)
		}
	})

	refusals := []struct {
		name   string
		mutate func(t *testing.T, repo frRepo, req *TransitionRequest)
		want   error
		text   string // a phrase the refusal must carry (the sentinel alone is shared with "unknown target state")
	}{
		{"tip_ahead_of_the_pushed_branch", func(t *testing.T, repo frRepo, _ *TransitionRequest) {
			frWrite(t, filepath.Join(repo.Dir, "d.txt"), "unpushed\n")
			frGit(t, repo.Dir, "add", "-A")
			frGit(t, repo.Dir, "commit", "-q", "-m", "unpushed")
		}, ErrEvidence, "pushed"},
		{"branch_never_pushed", func(t *testing.T, repo frRepo, _ *TransitionRequest) {
			frGit(t, repo.Dir, "checkout", "-q", "-b", "WT-never-pushed")
		}, ErrEvidence, "remote-tracking"},
		{"no_pr_number", func(_ *testing.T, _ frRepo, req *TransitionRequest) { req.PRNumber = "" }, ErrInvalidCardInput, "pull request"},
		{"non_numeric_pr_number", func(_ *testing.T, _ frRepo, req *TransitionRequest) { req.PRNumber = "seven" }, ErrInvalidCardInput, "pull request"},
		{"no_pr_url", func(_ *testing.T, _ frRepo, req *TransitionRequest) { req.PRURL = "" }, ErrInvalidCardInput, "pull request"},
		{"card_branch_is_the_integration_branch", func(_ *testing.T, _ frRepo, req *TransitionRequest) { req.IntegrationBranch = prCardBranch }, ErrEvidence, "integration branch"},
		{"not_the_lease_holder", func(_ *testing.T, _ frRepo, req *TransitionRequest) { req.Actor = "worker-2" }, ErrLeaseHolder, "lease"},
	}
	for _, tc := range refusals {
		t.Run("refused_"+tc.name, func(t *testing.T) {
			db := frOpen(t)
			repo := prRepo(t)
			prCard(t, db, repo, "c1", CardMerging)
			req := prOpenRequest("c1", repo)
			tc.mutate(t, repo, &req)
			before := frRowDump(t, db, frRun, "c1")
			if _, err := db.Transition(ctx, req); !errors.Is(err, tc.want) || !strings.Contains(err.Error(), tc.text) {
				t.Fatalf("err = %v, want %v carrying %q", err, tc.want, tc.text)
			}
			if after := frRowDump(t, db, frRun, "c1"); after != before {
				t.Fatalf("a refusal changed the record:\nbefore %s\nafter  %s", before, after)
			}
		})
	}

	t.Run("only_merging_reaches_pr_open", func(t *testing.T) {
		for _, from := range []string{CardMergeReady, CardRun, CardMergedLocal, CardPicked, CardDone} {
			db := frOpen(t)
			repo := prRepo(t)
			prCard(t, db, repo, "c1", from)
			if _, err := db.Transition(ctx, prOpenRequest("c1", repo)); !errors.Is(err, ErrIllegalTransition) {
				t.Errorf("%s → pr-open: err = %v, want ErrIllegalTransition", from, err)
			}
		}
	})
}

// The pr-open → merged-pr edge reads the merge commit off the remote-tracking
// ref of the integration branch — a PR "merged" claim alone is never enough.
func TestMergedPREdgeReadsTheMergeOnTheRemoteRef(t *testing.T) {
	ctx := context.Background()
	mergedReq := func(repo frRepo) TransitionRequest {
		return TransitionRequest{
			RunID: frRun, CardID: "c1", To: CardMergedPR, ExpectedVersion: 1, Actor: "worker-1",
			MergeSHA: repo.Merge, IntegrationBranch: repo.Integration, PRNumber: "7", Now: frNow,
		}
	}
	t.Run("merge_commit_on_the_remote_ref_is_accepted_and_recorded", func(t *testing.T) {
		db := frOpen(t)
		repo := prRepo(t)
		prCard(t, db, repo, "c1", CardPROpen)
		got, err := db.Transition(ctx, mergedReq(repo))
		if err != nil {
			t.Fatalf("pr-open → merged-pr: %v", err)
		}
		if got.State != CardMergedPR || got.MergeSHA != repo.Merge || got.MergeTree != repo.MergeTree {
			t.Fatalf("card = %s merge=%s tree=%s, want merged-pr with merge %s tree %s", got.State, got.MergeSHA, got.MergeTree, repo.Merge, repo.MergeTree)
		}
	})
	refusals := []struct {
		name   string
		mutate func(t *testing.T, repo frRepo, req *TransitionRequest)
	}{
		{"merge_commit_only_local", func(t *testing.T, repo frRepo, req *TransitionRequest) {
			frGit(t, repo.Dir, "checkout", "-q", "integration")
			frWrite(t, filepath.Join(repo.Dir, "local.txt"), "local\n")
			frGit(t, repo.Dir, "add", "-A")
			frGit(t, repo.Dir, "commit", "-q", "-m", "local only")
			req.MergeSHA = frGit(t, repo.Dir, "rev-parse", "HEAD")
		}},
		{"no_merge_sha", func(_ *testing.T, _ frRepo, req *TransitionRequest) { req.MergeSHA = "" }},
		{"not_a_sha", func(_ *testing.T, _ frRepo, req *TransitionRequest) { req.MergeSHA = "main" }},
		{"unusable_integration_branch", func(_ *testing.T, _ frRepo, req *TransitionRequest) { req.IntegrationBranch = "" }},
	}
	for _, tc := range refusals {
		t.Run("refused_"+tc.name, func(t *testing.T) {
			db := frOpen(t)
			repo := prRepo(t)
			prCard(t, db, repo, "c1", CardPROpen)
			req := mergedReq(repo)
			tc.mutate(t, repo, &req)
			before := frRowDump(t, db, frRun, "c1")
			if _, err := db.Transition(ctx, req); !errors.Is(err, ErrEvidence) {
				t.Fatalf("err = %v, want ErrEvidence", err)
			}
			if after := frRowDump(t, db, frRun, "c1"); after != before {
				t.Fatalf("a refusal changed the record:\nbefore %s\nafter  %s", before, after)
			}
		})
	}
	t.Run("only_pr_open_reaches_merged_pr", func(t *testing.T) {
		for _, from := range []string{CardMerging, CardMergeReady, CardMergedLocal, CardPushed, CardDone} {
			db := frOpen(t)
			repo := prRepo(t)
			prCard(t, db, repo, "c1", from)
			if _, err := db.Transition(ctx, mergedReq(repo)); !errors.Is(err, ErrIllegalTransition) {
				t.Errorf("%s → merged-pr: err = %v, want ErrIllegalTransition", from, err)
			}
		}
	})
}

// Consumer (2): the refusal wording names both merged states, and a card whose
// PR is merely open is not a merged predecessor.
func TestPredecessorUnmergedWordingNamesBothMergedStates(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	frPlace(t, db, Card{RunID: frRun, CardID: "pred", State: CardPROpen, Version: 1})
	frPlace(t, db, Card{RunID: frRun, CardID: "succ", State: CardPicked, Version: 1, HintAfter: "pred"})
	_, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "succ", To: CardAssigned, ExpectedVersion: 1, Actor: "assign", Owner: "worker-1", Now: frNow})
	if !errors.Is(err, ErrPredecessorUnmerged) {
		t.Fatalf("a pr-open predecessor: err = %v, want ErrPredecessorUnmerged", err)
	}
	for _, want := range []string{CardMergedLocal, CardMergedPR} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal %q does not name %s", err.Error(), want)
		}
	}
}

// An operator can still abandon a card whose PR never merged (T25 over the new
// states); the human decider remains required.
func TestPRStatesCanBeAbandoned(t *testing.T) {
	ctx := context.Background()
	for _, state := range []string{CardPROpen, CardMergedPR} {
		db := frOpen(t)
		frPlace(t, db, Card{RunID: frRun, CardID: "c1", State: state, Version: 1})
		if _, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "c1", To: CardAbandoned, ExpectedVersion: 1, Actor: "operator", Now: frNow}); !errors.Is(err, ErrDecider) {
			t.Errorf("%s → abandoned without a human decider: err = %v, want ErrDecider", state, err)
		}
		got, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "c1", To: CardAbandoned, ExpectedVersion: 1, Actor: "operator", Decider: DeciderHuman, Now: frNow})
		if err != nil || got.State != CardAbandoned {
			t.Errorf("%s → abandoned: state=%s err=%v", state, got.State, err)
		}
	}
}
