package worktree

// landing_predicate.go — SPEC-GITHUB-FLOW-DEFAULT-001 M2-A (card t1453),
// REQ-GFD-002 / design D-5: the squash-safe landing predicate.
//
// A card branch is "landed" when its work is on the integration ref. Under
// git-flow the card is merged with a merge commit, so the tip is an ancestor
// of the integration ref and one `git merge-base --is-ancestor` answers. Under
// github-flow the card is squash-merged: the tip is never an ancestor, and
// `git cherry` only recognises a one-commit card (measured: a two-commit card
// reads two "+"). The predicate therefore has three layers, tried in order:
//
//  1. ancestry — `git merge-base --is-ancestor <tip> <ref>`. The callers keep
//     their own layer 1 (done: the rev-list count that feeds its refusal
//     text; sweep: the three-way sweepAncestor; session exit: its own arm)
//     so the existing git-flow output is untouched; this file starts at the
//     question layer 1 answered "no" to.
//  2. cumulative patch-id — the patch-id of `git diff <merge-base> <tip>`
//     compared with the patch-id of every commit on <ref> since the
//     merge-base. A squash commit carries the card's cumulative change, so
//     the ids are equal. Capped at landingPatchIDCommitCap commits; over the
//     cap the layer cannot answer.
//  3. PR merged state — `gh pr list --head <branch> --state merged`, bounded by
//     landingGHTimeout, fail-closed on any gh failure. Landed only when a PR
//     is MERGED, its headRefOid IS the local tip (a later local commit would
//     otherwise be deleted with the branch), and its merge commit is an
//     ancestor of <ref> (a fetch lag keeps the answer "preserve").
//
// `moai worktree done` and `moai worktree sweep` run all three
// (landedBeyondAncestry); session-exit cleanup runs layers 1-2 only through
// LandedByPatchID and never reaches layer 3 — REQ-WSS-304 keeps that shared
// exit path free of network calls, and REQ-GFD-002's second clause restates it.
// A layer that cannot answer is never a "landed": an unconfirmable landing
// preserves the tree.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// landingPatchIDCommitCap bounds how many integration-ref commits since the
// merge-base the cumulative patch-id layer compares (design D-5: 500 — about
// two weeks of card merges, measured in the SPEC ledger as 248 per week).
var landingPatchIDCommitCap = 500

// landingGHTimeout bounds one `gh` call of the PR-state layer (design D-5:
// 10 s, no retry).
var landingGHTimeout = 10 * time.Second

// landingGH is the `gh` execution seam of the PR-state layer: it runs `gh` in
// dir under ctx and returns stdout. Tests replace it with a double; nothing in
// the test binary may reach the real gh CLI or the network.
var landingGH = func(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GH_PROMPT_DISABLED=1")
	// A gh that ignores the kill must not hold the call past its bound.
	cmd.WaitDelay = 2 * time.Second
	return cmd.Output()
}

// runLandingGit runs `git -C dir <args>` with optional stdin and returns raw
// stdout (never trimmed: a trailing newline is part of a diff). The exit code
// is returned beside the error so `merge-base --is-ancestor` can tell "no"
// (exit 1) from "cannot tell" (anything else); -1 means git did not run.
func runLandingGit(dir, stdin string, args ...string) (stdout string, exit int, err error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	runErr := cmd.Run()
	if runErr == nil {
		return out.String(), 0, nil
	}
	if ee, ok := runErr.(*exec.ExitError); ok {
		return out.String(), ee.ExitCode(), runErr
	}
	return out.String(), -1, runErr
}

// landingVerbatimSupport is the support seam of the whitespace-faithful
// patch-id mode: it answers whether this git accepts `git patch-id
// --verbatim`. nil means "not yet measured" — the next caller runs the real
// probe (probeLandingVerbatimSupport). Tests assign it to pin both answers
// without depending on the installed git, the same way landingGH is doubled.
var landingVerbatimSupport func(dir string) (bool, error)

// verbatimSupported resolves the seam above, measuring when it is unset.
func verbatimSupported(dir string) (bool, error) {
	if landingVerbatimSupport != nil {
		return landingVerbatimSupport(dir)
	}
	return probeLandingVerbatimSupport(dir)
}

// probeLandingVerbatimSupport measures whether this git accepts
// `git patch-id --verbatim` by running the mode once on a fixed one-line
// patch: a git without the flag exits with a usage error, a git with it
// prints a patch-id. `git --version` is never parsed — distros, builds and
// locales vary, so the feature itself is observed.
func probeLandingVerbatimSupport(dir string) (bool, error) {
	const probePatch = "diff --git a/f.txt b/f.txt\n" +
		"index 0000000..7898192 100644\n" +
		"--- a/f.txt\n" +
		"+++ b/f.txt\n" +
		"@@ -0,0 +1 @@\n" +
		"+a\n"
	out, exit, err := runLandingGit(dir, probePatch, "patch-id", "--verbatim")
	if err != nil {
		if exit > 1 { // a usage error means this git has no --verbatim flag
			return false, nil
		}
		return false, fmt.Errorf("git patch-id --verbatim probe: %w", err)
	}
	if strings.TrimSpace(out) == "" {
		return false, fmt.Errorf("git patch-id --verbatim probe: empty answer")
	}
	return true, nil
}

// landingPatchIDs runs `git patch-id --verbatim` over stream (a diff or a
// `git log -p` stream) and returns the patch-ids in order. The mode is the
// whitespace-faithful one (SPEC-GFD-PATCHID-VERBATIM-001 REQ-GPV-001): the
// default mode normalizes away whitespace and hunk line numbers, which read
// a remote squash amended with a whitespace-only reformat as the card's
// landing. A git without the verbatim mode cannot answer (REQ-GPV-003) —
// the layer errors instead of falling back to the normalizing mode, so the
// caller preserves and landedBeyondAncestry still reaches the PR layer.
func landingPatchIDs(dir, stream string) ([]string, error) {
	supported, err := verbatimSupported(dir)
	if err != nil {
		return nil, fmt.Errorf("patch-id mode probe: %w", err)
	}
	if !supported {
		return nil, fmt.Errorf("git patch-id --verbatim is not supported by this git: the whitespace-faithful comparison cannot answer")
	}
	out, _, err := runLandingGit(dir, stream, "patch-id", "--verbatim")
	if err != nil {
		return nil, fmt.Errorf("git patch-id: %w", err)
	}
	var ids []string
	for ln := range strings.SplitSeq(out, "\n") {
		if fields := strings.Fields(ln); len(fields) > 0 {
			ids = append(ids, fields[0])
		}
	}
	return ids, nil
}

// landingDiffFlags keep a diff and a `git log -p` stream comparable: no
// external diff drivers, no colour, no rename detection (it is configuration
// dependent and changes the hunk shape), binary content included.
var landingDiffFlags = []string{"--no-ext-diff", "--no-color", "--no-renames", "--binary"}

// LandedByPatchID is layer 2 — and the only layer past ancestry that session-exit
// cleanup may use. It reports whether the cumulative patch of tip (relative to
// its merge-base with ref) already exists as a commit on ref. A nil error means
// the layer answered, yes or no; an error means it could not (no merge-base, a
// range over the commit cap, a branch with no net change, a git failure) and the
// caller must treat the landing as unconfirmed — never as landed.
func LandedByPatchID(dir, tip, ref string) (bool, error) {
	mbOut, _, err := runLandingGit(dir, "", "merge-base", tip, ref)
	mb := strings.TrimSpace(mbOut)
	if err != nil || mb == "" {
		return false, fmt.Errorf("no merge-base of %s and %s: %w", tip, ref, errOrEmpty(err))
	}
	countOut, _, err := runLandingGit(dir, "", "rev-list", "--count", mb+".."+ref)
	if err != nil {
		return false, fmt.Errorf("count commits %s..%s: %w", mb, ref, err)
	}
	n, convErr := strconv.Atoi(strings.TrimSpace(countOut))
	if convErr != nil {
		return false, fmt.Errorf("count commits %s..%s: unexpected output %q", mb, ref, strings.TrimSpace(countOut))
	}
	if n > landingPatchIDCommitCap {
		return false, fmt.Errorf("%d commits on %s since the merge-base exceed the patch-id comparison cap %d", n, ref, landingPatchIDCommitCap)
	}
	if n == 0 {
		return false, nil // nothing landed on ref since the card forked: nothing to match
	}
	diff, _, err := runLandingGit(dir, "", append(append([]string{"diff"}, landingDiffFlags...), mb, tip)...)
	if err != nil {
		return false, fmt.Errorf("git diff %s %s: %w", mb, tip, err)
	}
	if strings.TrimSpace(diff) == "" {
		return false, fmt.Errorf("%s carries no net change against %s: a patch-id comparison has nothing to match", tip, mb)
	}
	cardIDs, err := landingPatchIDs(dir, diff)
	if err != nil {
		return false, err
	}
	if len(cardIDs) != 1 {
		return false, fmt.Errorf("cumulative diff of %s yielded %d patch-ids, want 1", tip, len(cardIDs))
	}
	logArgs := append(append([]string{"log", "-p", "--no-merges", "--format=commit %H"}, landingDiffFlags...), mb+".."+ref)
	stream, _, err := runLandingGit(dir, "", logArgs...)
	if err != nil {
		return false, fmt.Errorf("git log -p %s..%s: %w", mb, ref, err)
	}
	refIDs, err := landingPatchIDs(dir, stream)
	if err != nil {
		return false, err
	}
	for _, id := range refIDs {
		if id == cardIDs[0] {
			return true, nil
		}
	}
	return false, nil
}

// LandedByCommitPatchIDs is the commit-level whitespace-faithful equivalence
// predicate — the arm of the session-exit cleanup where `git cherry` stood
// (SPEC-GFD-PATCHID-VERBATIM-001, decision Q3). It reports whether every
// commit of tip missing from ref has its patch on ref under the verbatim
// patch-id: the per-commit confirmation cherry provided, recomputed in the
// same mode the cumulative layer (LandedByPatchID) compares in, so a
// whitespace-divergent squash can no longer read a card commit as landed
// (REQ-GPV-002 keeps the per-commit ability; REQ-GPV-001 keeps the bytes).
//
// cherry's internal equivalence is not parameterizable and normalizes like
// the default patch-id mode, which is exactly the t1561 defect at the
// session exit. The range of ref-only commits is capped at
// landingPatchIDCommitCap like the cumulative layer; over the cap, without
// verbatim support, or on any git failure the layer cannot answer — the
// caller preserves (REQ-WSS-302) and the next sweep decides. Merge and
// empty commits carry no patch of their own: they are skipped on the head
// side, and every patch-carrying commit of the card must still match.
//
// @MX:NOTE: [AUTO] the only commit-level landing check past ancestry on the
// session-exit path; its answer lets cleanupSessionWorktree remove a card
// worktree whose branch is confirmed landed.
// @MX:SPEC: SPEC-GFD-PATCHID-VERBATIM-001
func LandedByCommitPatchIDs(dir, tip, ref string) (bool, error) {
	countOut, _, err := runLandingGit(dir, "", "rev-list", "--count", tip+".."+ref)
	if err != nil {
		return false, fmt.Errorf("count commits %s..%s: %w", tip, ref, err)
	}
	n, convErr := strconv.Atoi(strings.TrimSpace(countOut))
	if convErr != nil {
		return false, fmt.Errorf("count commits %s..%s: unexpected output %q", tip, ref, strings.TrimSpace(countOut))
	}
	if n > landingPatchIDCommitCap {
		return false, fmt.Errorf("%d commits on %s beyond %s exceed the patch-id comparison cap %d", n, ref, tip, landingPatchIDCommitCap)
	}
	headStream, _, err := runLandingGit(dir, "", append(append([]string{"log", "-p", "--no-merges", "--format=commit %H"}, landingDiffFlags...), ref+".."+tip)...)
	if err != nil {
		return false, fmt.Errorf("git log -p %s..%s: %w", ref, tip, err)
	}
	headIDs, err := landingPatchIDs(dir, headStream)
	if err != nil {
		return false, err
	}
	if len(headIDs) == 0 {
		return true, nil // nothing of the card is off the ref: cherry's empty answer
	}
	refStream, _, err := runLandingGit(dir, "", append(append([]string{"log", "-p", "--no-merges", "--format=commit %H"}, landingDiffFlags...), tip+".."+ref)...)
	if err != nil {
		return false, fmt.Errorf("git log -p %s..%s: %w", tip, ref, err)
	}
	refIDs, err := landingPatchIDs(dir, refStream)
	if err != nil {
		return false, err
	}
	matched := make(map[string]struct{}, len(refIDs))
	for _, id := range refIDs {
		matched[id] = struct{}{}
	}
	for _, id := range headIDs {
		if _, ok := matched[id]; !ok {
			return false, nil
		}
	}
	return true, nil
}

func errOrEmpty(err error) error {
	if err != nil {
		return err
	}
	return fmt.Errorf("empty answer")
}

// landingPR is the subset of `gh pr list --json` the PR layer reads.
type landingPR struct {
	State       string `json:"state"`
	HeadRefOid  string `json:"headRefOid"`
	MergeCommit *struct {
		Oid string `json:"oid"`
	} `json:"mergeCommit"`
}

// landedByMergedPR is layer 3. It reads the merged PRs whose head is branch and
// reports landed only when one of them is MERGED, was merged from exactly the
// local tip, and has its merge commit reachable from ref. Any gh failure —
// non-zero exit, timeout, unreadable JSON — is "cannot answer" (fail-closed).
func landedByMergedPR(dir, branch, ref string) (bool, error) {
	tipOut, _, err := runLandingGit(dir, "", "rev-parse", "--verify", "--quiet", branch+"^{commit}")
	tip := strings.TrimSpace(tipOut)
	if err != nil || tip == "" {
		return false, fmt.Errorf("resolve local tip of %s: %w", branch, errOrEmpty(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), landingGHTimeout)
	defer cancel()
	out, err := landingGH(ctx, dir, "pr", "list", "--head", branch, "--state", "merged",
		"--limit", "20", "--json", "number,state,headRefOid,mergeCommit")
	if ctxErr := ctx.Err(); ctxErr != nil {
		return false, fmt.Errorf("gh pr list for %s did not answer within %s: %w", branch, landingGHTimeout, ctxErr)
	}
	if err != nil {
		return false, fmt.Errorf("gh pr list for %s: %w", branch, err)
	}
	var prs []landingPR
	if err := json.Unmarshal(out, &prs); err != nil {
		return false, fmt.Errorf("gh pr list for %s: unreadable answer: %w", branch, err)
	}
	var lastErr error
	for _, pr := range prs {
		if !strings.EqualFold(pr.State, "MERGED") || pr.HeadRefOid != tip {
			continue // a PR merged from another head: the local tip is not what merged
		}
		if pr.MergeCommit == nil || pr.MergeCommit.Oid == "" {
			continue
		}
		_, exit, ancErr := runLandingGit(dir, "", "merge-base", "--is-ancestor", pr.MergeCommit.Oid, ref)
		switch {
		case ancErr == nil:
			return true, nil
		case exit == 1:
			// MERGED, but the merge commit is not on the local ref yet (a
			// fetch lag): the third condition is false, so preserve.
		default:
			lastErr = fmt.Errorf("merge commit %s of %s is not resolvable against %s: %w", pr.MergeCommit.Oid, branch, ref, ancErr)
		}
	}
	return false, lastErr
}

// landedBeyondAncestry is the shared predicate `moai worktree done` and
// `moai worktree sweep` run once their own layer 1 has said "not an ancestor of
// ref": layer 2, then layer 3. It returns the layer that confirmed the landing
// (2 or 3), or 0 when nothing did — in which case the caller preserves.
//
// @MX:ANCHOR: [AUTO] the single landing predicate past ancestry — done, sweep and (layers 1-2 only, via LandedByPatchID) session-exit cleanup all answer "is this card on the integration ref" here
// @MX:REASON: three call sites decide whether a card worktree is deleted on this answer; a divergence between them is how a squash-merged card would be kept forever by one path and deleted by another (SPEC-GITHUB-FLOW-DEFAULT-001 REQ-GFD-002)
// @MX:SPEC: SPEC-GITHUB-FLOW-DEFAULT-001
func landedBeyondAncestry(dir, branch, ref string) (landed bool, layer int) {
	if ok, err := LandedByPatchID(dir, branch, ref); err == nil && ok {
		return true, 2
	}
	if ok, err := landedByMergedPR(dir, branch, ref); err == nil && ok {
		return true, 3
	}
	return false, 0
}
