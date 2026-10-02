package homestate

// card_pr_states.go — the github-flow card delivery states
// (SPEC-GITHUB-FLOW-DEFAULT-001 M2-B, design D-4 option S-a).
//
// git-flow keeps `merging → merged-local`; github-flow delivers a card by pull
// request and adds two states beside it rather than overloading one:
//
//	merging ─▶ pr-open ─▶ merged-pr ─▶ done
//
// `merged-pr` is named for where the merge happened, the way `merged-local` is:
// "local" would be a false name for a merge GitHub performed, and renaming
// `merged-local` would migrate every recorded card and the MCP surface (design
// D-4, S-b and S-c rejected). Existing records are read unchanged.
//
// Neither state holds a lease: the lane is free once the PR is open (the
// serial slot releases at `merging` already), so `merging → pr-open` is the
// lease holder's last edge and the two later edges are evidence-gated, not
// holder-gated. Each edge reads its git evidence itself (REQ-FR-010); the
// caller supplies identities only — the PR number and URL, the merge commit's
// SHA. That the PR was observed MERGED is the caller's `gh` read; what the
// record verifies is that the merge commit is on the remote-tracking ref of
// the integration branch.

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

const (
	// CardPROpen is a card whose branch is pushed and whose pull request is
	// open against the integration target with auto-merge requested.
	CardPROpen = "pr-open"
	// CardMergedPR is a card whose pull request was observed merged.
	CardMergedPR = "merged-pr"
)

// The github-flow edge guards. They sit above the iota block of card_transition.go.
const (
	guardPROpen edgeGuard = iota + 100
	guardPRMerged
	guardPRDone
)

func init() {
	// merging → pr-open is the lease holder's edge; the other two are not.
	holderGuards[guardPROpen] = true
}

// prTransitionEdges are the requested edges the github-flow states add to the
// transition table: `merging → pr-open → merged-pr → done`. Abandon (T25)
// covers the new states through the cardStates loop.
func prTransitionEdges() []transitionEdge {
	return []transitionEdge{
		{"TP1", CardMerging, CardPROpen, guardPROpen},
		{"TP2", CardPROpen, CardMergedPR, guardPRMerged},
		{"TP3", CardMergedPR, CardDone, guardPRDone},
	}
}

// planPREdge checks one github-flow edge's guard and computes its effect.
func planPREdge(ctx context.Context, cur Card, edge transitionEdge, req TransitionRequest, plan transitionPlan, nowText string) (transitionPlan, error) {
	switch edge.guard {
	case guardPROpen:
		ev, err := verifyPROpen(ctx, cur.WorktreePath, req)
		if err != nil {
			return plan, err
		}
		plan.evidence = ev
	case guardPRMerged:
		ref, err := verifyPushed(ctx, cur.WorktreePath, req.MergeSHA, req.IntegrationBranch)
		if err != nil {
			return plan, err
		}
		full, err := resolveCommit(ctx, cur.WorktreePath, req.MergeSHA)
		if err != nil {
			return plan, err
		}
		tree, err := gitRead(ctx, cur.WorktreePath, "rev-parse", full+"^{tree}")
		if err != nil {
			return plan, evidenceErr("read tree of %s: %v", full, err)
		}
		plan.next.MergeSHA, plan.next.MergeTree = full, tree
		plan.evidence["merge_sha"], plan.evidence["remote_ref"] = full, ref
		if n := strings.TrimSpace(req.PRNumber); n != "" {
			plan.evidence["pr_number"] = n
		}
	case guardPRDone:
		// The human closure of a PR-delivered card, like the no-remote T18: the
		// merge is already on the remote, so there is nothing for the
		// merged-local → done remote refusal to protect.
		if req.Decider != DeciderHuman {
			return plan, fmt.Errorf("%w: F1 accepts only decider %q, got %q", ErrDecider, DeciderHuman, req.Decider)
		}
		plan.next.Decider, plan.next.DecidedAt = DeciderHuman, nowText
		plan.note = "merged by pull request — the integration branch's CI verdict is not read by this edge"
	}
	return plan, nil
}

// verifyPROpen is the merging → pr-open reader: the PR's identity is present,
// the card worktree is on a branch that is not the integration branch, and that
// branch's tip is the tip on the remote-tracking ref — what the PR carries is
// what the worktree holds.
func verifyPROpen(ctx context.Context, dir string, req TransitionRequest) (map[string]string, error) {
	number, url := strings.TrimSpace(req.PRNumber), strings.TrimSpace(req.PRURL)
	if n, err := strconv.Atoi(number); err != nil || n <= 0 || url == "" || strings.ContainsAny(url, " \t\r\n") {
		return nil, fmt.Errorf("%w: moving a card to pr-open requires the pull request's number and URL", ErrInvalidCardInput)
	}
	if !validBranchName(req.IntegrationBranch) {
		return nil, evidenceErr("integration branch %q is not usable", req.IntegrationBranch)
	}
	branch, err := gitRead(ctx, dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil || !validBranchName(branch) || branch == "HEAD" {
		return nil, evidenceErr("the card worktree is not on a branch")
	}
	if branch == req.IntegrationBranch {
		return nil, evidenceErr("card branch %q is the integration branch: a pull request never carries its base as its head", branch)
	}
	tip, err := gitRead(ctx, dir, "rev-parse", "--verify", "--quiet", "HEAD^{commit}")
	if err != nil || tip == "" {
		return nil, evidenceErr("the card worktree HEAD does not resolve")
	}
	remote, err := cardRemote(ctx, dir)
	if err != nil {
		return nil, err
	}
	ref := "refs/remotes/" + remote + "/" + branch
	pushed, err := gitRead(ctx, dir, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
	if err != nil || pushed == "" {
		return nil, evidenceErr("remote-tracking ref %s does not exist: the card branch was never pushed", ref)
	}
	if pushed != tip {
		return nil, evidenceErr("card tip %s is not the pushed tip of %s (%s): push the branch before the pull request", tip, branch, pushed)
	}
	return map[string]string{"pr_number": number, "pr_url": url, "branch": branch, "tip": tip, "remote_ref": ref}, nil
}

// cardRemote names the remote the card's branch was pushed to — origin when
// there is one, else the first configured (the choice verifyPushed makes).
func cardRemote(ctx context.Context, dir string) (string, error) {
	remotes, err := gitRemotes(ctx, dir)
	if err != nil || len(remotes) == 0 {
		return "", evidenceErr("no remote is configured")
	}
	if slices.Contains(remotes, "origin") {
		return "origin", nil
	}
	return remotes[0], nil
}
