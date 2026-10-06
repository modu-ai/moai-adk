package cli

// factory_card_pr.go — the github-flow card delivery edge of
// `moai factory complete` (SPEC-GITHUB-FLOW-DEFAULT-001 M2-B, REQ-GFD-004/005/006,
// design D-1, D-4, D-20).
//
// Under github-flow `complete` does not merge locally and does not take the
// integration window. It checks merge-readiness against the integration target
// (the condition triple, before anything is pushed or opened), pushes the card
// branch to origin, opens a pull request whose base is the target and whose
// head is the card branch, asks for auto-merge with the configured merge
// method, and records `pr-open`. A later `complete` reads the PR through
// `gh pr view`; only when it reads MERGED — from the card's own tip, with the
// merge commit on origin's integration branch — does the record move to
// `merged-pr`. Every gh failure, timeout or unreadable answer is "cannot
// confirm": the verb reports it and the record keeps what it had.
//
// git-flow never reaches this file: factoryCompleteCard hands over only when
// the active git strategy's workflow is github-flow. The merge queue stays
// deferred (design D-3): one PR per card, no merge_group.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorylane"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// factoryGHTimeout bounds one `gh` call of the delivery edge (design D-5: 10 s,
// no retry; the delivery edge uses the bound the landing predicate uses).
var factoryGHTimeout = 10 * time.Second

// factoryGH is the `gh` execution seam of the delivery edge: it runs `gh` in
// dir under ctx and returns stdout. Tests replace it with a double; nothing in
// the test binary may reach the real gh CLI or the network.
var factoryGH = func(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GH_PROMPT_DISABLED=1")
	// A gh that ignores the kill must not hold the call past its bound.
	cmd.WaitDelay = 2 * time.Second
	return cmd.Output()
}

// factoryPRFields is the `gh pr view --json` field list the edge reads.
const factoryPRFields = "number,url,state,baseRefName,headRefName,headRefOid,mergeCommit"

// factoryPR is the subset of a pull request the delivery edge reads.
type factoryPR struct {
	Number      int    `json:"number"`
	URL         string `json:"url"`
	State       string `json:"state"`
	BaseRefName string `json:"baseRefName"`
	HeadRefName string `json:"headRefName"`
	HeadRefOid  string `json:"headRefOid"`
	MergeCommit *struct {
		Oid string `json:"oid"`
	} `json:"mergeCommit"`
}

// factoryGitHubFlow reports whether the project's active git strategy is
// github-flow — the only configuration the delivery edge is live under. An
// absent or unreadable git-strategy.yaml, and every other workflow, is not.
func factoryGitHubFlow(root string) bool {
	return config.LoadGitFlowIntegrationConfig(root).Workflow == config.WorkflowGitHubFlow
}

// factoryPRMergeFlag maps git_strategy.<mode>.merge_method to the `gh pr merge`
// flag. An absent or unrecognized method is the default, squash (spec-workflow
// § SPEC Phase Discipline: the configured merge_method, default squash).
func factoryPRMergeFlag(method string) string {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case "merge":
		return "--merge"
	case "rebase":
		return "--rebase"
	default:
		return "--squash"
	}
}

// factoryGHCall runs one gh call under the bound and returns stdout. A timeout
// and a non-zero exit are both errors; gh's own stderr rides in the message so
// a caller can tell "no pull requests found" from a failure.
func factoryGHCall(dir string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), factoryGHTimeout)
	defer cancel()
	out, err := factoryGH(ctx, dir, args...)
	name := "gh " + strings.Join(args[:min(2, len(args))], " ")
	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, fmt.Errorf("%s did not answer within %s: %w", name, factoryGHTimeout, ctxErr)
	}
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return nil, fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return out, nil
}

// factoryReadPR reads the pull request whose head is branch. found is false
// only when gh says there is none; every other failure is an error ("cannot
// confirm").
func factoryReadPR(dir, branch string) (pr *factoryPR, found bool, err error) {
	out, err := factoryGHCall(dir, "pr", "view", branch, "--json", factoryPRFields)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no pull requests found") {
			return nil, false, nil
		}
		return nil, false, err
	}
	var got factoryPR
	if err := json.Unmarshal(out, &got); err != nil {
		return nil, false, fmt.Errorf("gh pr view for %s: unreadable answer: %w", branch, err)
	}
	if got.Number <= 0 || strings.TrimSpace(got.URL) == "" {
		return nil, false, fmt.Errorf("gh pr view for %s: the answer carries no pull request number and URL", branch)
	}
	return &got, true, nil
}

// factoryGitRead runs `git -C dir <args>` and returns trimmed stdout.
func factoryGitRead(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(ee.Stderr)))
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

// factoryCompleteGitHubFlow is the github-flow body of `moai factory complete`
// (called from factoryCompleteCard once the Codex refusal has passed). It takes
// no integration window and never merges locally.
func factoryCompleteGitHubFlow(ctx context.Context, out io.Writer, root, cardID, remeasure, run, lane string) error {
	if strings.TrimSpace(remeasure) != "" {
		return fmt.Errorf("factory complete: the re-measure positional is the git-flow merge record; under github-flow the delivery evidence is the pull request, so none is taken")
	}
	runID, err := resolveFactoryCardRun(ctx, root, run)
	if err != nil {
		return fmt.Errorf("factory complete: %w", err)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return fmt.Errorf("factory complete: %w", err)
	}
	defer func() { _ = db.Close() }()
	card, err := db.LoadCard(ctx, runID, cardID)
	if err != nil {
		return fmt.Errorf("factory complete: %w", err)
	}
	cfg := config.LoadGitFlowIntegrationConfig(root)
	target := cfg.IntegrationTarget
	if target == "" {
		return fmt.Errorf("factory complete: no integration target is configured, so the pull request's base would be a guess: %s", cfg.EmptyTargetGuidance(root, "set the workflow"))
	}
	cardBranch := factoryBranchOfWorktree(card.WorktreePath)
	if cardBranch == "" || cardBranch == "HEAD" || cardBranch == target {
		return fmt.Errorf("factory complete: refused — card %s's worktree %q is not on a card branch (reads %q, integration target %q)", card.CardID, card.WorktreePath, cardBranch, target)
	}

	switch card.State {
	case homestate.CardPROpen:
		return factoryObservePRMerge(ctx, out, db, root, runID, card, cardBranch, target, lane)
	case homestate.CardMergedPR:
		_, _ = fmt.Fprintf(out, "%s %s merge=%s — nothing to do\n", card.CardID, card.State, card.MergeSHA)
		return nil
	case homestate.CardMergeReady, homestate.CardMerging:
		return factoryDeliverByPR(ctx, out, db, root, runID, card, cardBranch, target, cfg.MergeMethod, lane)
	}
	return fmt.Errorf("factory complete: card %s is %s; complete takes a merge-ready, merging or pr-open card", card.CardID, card.State)
}

// factoryPRReadiness runs the condition triple (sync audit PASS record,
// conflict-free merge-tree, tree identity) against the integration target's
// remote ref — the ref the PR will merge into — and records the run. It never
// merges and takes no window.
func factoryPRReadiness(out io.Writer, root string, card homestate.Card, lane, cardBranch, target string) (factorylane.MergeCheckRun, error) {
	wt := card.WorktreePath
	ref := target
	if _, err := factoryGitRead(wt, "fetch", "origin", target); err != nil {
		_, _ = fmt.Fprintf(out, "  note: could not fetch origin/%s (%v); probing the last fetched ref\n", target, err)
	}
	if _, err := factoryGitRead(wt, "rev-parse", "--verify", "--quiet", "refs/remotes/origin/"+target); err == nil {
		ref = "origin/" + target
	}
	specDir := ""
	if card.SpecID != "" { // a SPEC-less card reads its verdict file (empty SpecDir)
		specDir = filepath.Join(wt, ".moai", "specs", card.SpecID)
	}
	run, err := factorylane.EvaluateMergeTriple(factorylane.MergeTripleInput{
		Lane:    lane,
		Card:    card.CardID,
		SpecDir: specDir,
		Branch:  cardBranch,
		Develop: ref,
		RepoDir: wt,
	}, factorylane.ExecGitRunner{Dir: wt})
	if err != nil {
		return run, err
	}
	return factorylane.NewStore(root, nil).RecordMergeCheckRun(run)
}

// factoryDeliverByPR is the merge-ready/merging half: check, push, open, ask
// for auto-merge, record pr-open. Everything that can refuse without changing a
// record runs before the first record change.
func factoryDeliverByPR(ctx context.Context, out io.Writer, db *homestate.FactoryDB, root, runID string, card homestate.Card, cardBranch, target, method, lane string) error {
	// A card already at merging is a delivery RETRY: the merge-ready entry
	// took the T14 lease-holder edge, but this path ran the push, the pull
	// request, and the auto-merge request with no owner check at all — a
	// foreign lane's work landed before the record refused it (card t1533,
	// review-gate r2 finding c). Only the recorded lease holder re-enters a
	// mutating path, and only on a lease that has not expired — an expired
	// holder's retry ran the same remote mutations before the F1 expiry
	// refusal (review-gate r5) — and the refusal precedes the push.
	if card.State == homestate.CardMerging {
		if holder := strings.TrimSpace(card.LeaseHolder); holder == "" || holder != lane {
			return fmt.Errorf("factory complete: refused — card %s is merging under lease holder %s; %s cannot retry the delivery", card.CardID, dash(holder), dash(lane))
		}
		if card.LeaseExpired(factoryCardNow()) {
			return fmt.Errorf("factory complete: refused — card %s's merging lease held by %s expired at %s; the expiry must be collected before the delivery is retried", card.CardID, dash(card.LeaseHolder), card.LeaseExpiresAt)
		}
	}
	wt := card.WorktreePath
	if _, err := factoryGitRead(wt, "config", "--get", "remote.origin.url"); err != nil {
		return fmt.Errorf("factory complete: refused — the repository has no remote named origin to push %s to", cardBranch)
	}
	// REQ-GFD-005: the readiness check precedes the PR; a failing check opens nothing.
	run, err := factoryPRReadiness(out, root, card, lane, cardBranch, target)
	if err != nil {
		return fmt.Errorf("factory complete: the merge-readiness check could not run: %w", err)
	}
	if !run.AllPassed {
		_, _ = fmt.Fprintf(out, "merge-readiness: REFUSED — failing condition: %s\nno pull request was opened\n", run.FailedCondition)
		printMergeChecks(out, run.Checks)
		return fmt.Errorf("factory complete: refused — merge-readiness failed on %s against %s; no branch was pushed and no pull request was opened", run.FailedCondition, target)
	}

	cur := card
	if card.State == homestate.CardMergeReady {
		// T14 — the lease holder's edge, refused by F1 verbatim for any other lane.
		cur, err = db.Transition(ctx, homestate.TransitionRequest{
			RunID: runID, CardID: card.CardID, To: homestate.CardMerging,
			ExpectedVersion: card.Version, Actor: lane, Now: factoryCardNow(),
		})
		if err != nil {
			return fmt.Errorf("factory complete: %w", err)
		}
	}
	stays := fmt.Sprintf("card %s stays in merging", card.CardID)

	if _, err := factoryGitRead(wt, "push", "origin", cardBranch); err != nil {
		return fmt.Errorf("factory complete: %s; the push of %s failed: %w", stays, cardBranch, err)
	}
	pr, found, err := factoryReadPR(wt, cardBranch)
	if err != nil {
		return fmt.Errorf("factory complete: %s; the pull request could not be read, so none is opened: %w", stays, err)
	}
	if !found {
		title, body := factoryPRText(wt, card, target)
		if _, err := factoryGHCall(wt, "pr", "create", "--base", target, "--head", cardBranch, "--title", title, "--body", body); err != nil {
			return fmt.Errorf("factory complete: %s; opening the pull request failed: %w", stays, err)
		}
		if pr, found, err = factoryReadPR(wt, cardBranch); err != nil || !found {
			return fmt.Errorf("factory complete: %s; the pull request was opened but could not be read back (re-run complete): %v", stays, errOr(err, "no pull request found"))
		}
	}
	if strings.EqualFold(pr.State, "CLOSED") {
		return fmt.Errorf("factory complete: %s; pull request #%d is closed without merging — an operator decides", stays, pr.Number)
	}
	if pr.BaseRefName != target {
		return fmt.Errorf("factory complete: %s; pull request #%d targets %q, not the integration target %q", stays, pr.Number, pr.BaseRefName, target)
	}
	mergeFlag := factoryPRMergeFlag(method)
	if !strings.EqualFold(pr.State, "MERGED") {
		if _, err := factoryGHCall(wt, "pr", "merge", strconv.Itoa(pr.Number), "--auto", mergeFlag); err != nil {
			return fmt.Errorf("factory complete: %s; pull request #%d is open but the auto-merge request failed: %w", stays, pr.Number, err)
		}
	}
	open, err := db.Transition(ctx, homestate.TransitionRequest{
		RunID: runID, CardID: card.CardID, To: homestate.CardPROpen,
		ExpectedVersion: cur.Version, Actor: lane, IntegrationBranch: target,
		PRNumber: strconv.Itoa(pr.Number), PRURL: pr.URL, Now: factoryCardNow(),
	})
	if err != nil {
		return fmt.Errorf("factory complete: %s; pull request #%d is open but the F1 record refused pr-open: %w", stays, pr.Number, err)
	}
	_, _ = fmt.Fprintf(out, "%s %s pr=%s branch=%s base=%s\n", open.CardID, open.State, pr.URL, cardBranch, target)
	_, _ = fmt.Fprintf(out, "  auto-merge requested (%s); run moai factory complete %s again once the pull request merges to record merged-pr\n", mergeFlag, open.CardID)
	if strings.EqualFold(pr.State, "MERGED") {
		return factoryObservePRMerge(ctx, out, db, root, runID, open, cardBranch, target, lane)
	}
	return nil
}

func errOr(err error, fallback string) string {
	if err != nil {
		return err.Error()
	}
	return fallback
}

// factoryPRText composes the pull request's title and body: the card id leads
// the title (the traceability carrier), the body names the evidence path and
// closes with the attribution line.
func factoryPRText(wt string, card homestate.Card, target string) (title, body string) {
	subject, err := factoryGitRead(wt, "log", "-1", "--format=%s")
	if err != nil || subject == "" {
		subject = "card " + card.CardID
	}
	title = "[" + card.CardID + "] " + subject
	var b strings.Builder
	fmt.Fprintf(&b, "Card: %s\n", card.CardID)
	if card.SpecID != "" {
		fmt.Fprintf(&b, "SPEC: %s\n", card.SpecID)
	}
	fmt.Fprintf(&b, "Evidence: .moai/reports/%s/verdict.md\n\n", card.CardID)
	fmt.Fprintf(&b, "Opened by `moai factory complete` after the merge-readiness check against %s (sync audit PASS record, conflict-free merge, tree identity).\n\n", target)
	b.WriteString("🗿 MoAI")
	return title, b.String()
}

// factoryObservePRMerge is the pr-open half: read the PR; record merged-pr only
// when it reads MERGED from exactly this card's tip and the merge commit is on
// origin's integration branch. An open PR is a clean "not yet"; every other
// answer is "cannot confirm" and records nothing.
func factoryObservePRMerge(ctx context.Context, out io.Writer, db *homestate.FactoryDB, root, runID string, card homestate.Card, cardBranch, target, lane string) error {
	wt := card.WorktreePath
	stays := fmt.Sprintf("card %s stays pr-open", card.CardID)
	pr, found, err := factoryReadPR(wt, cardBranch)
	if err != nil {
		return fmt.Errorf("factory complete: %s; cannot confirm the pull request: %w", stays, err)
	}
	if !found {
		return fmt.Errorf("factory complete: %s; no pull request is found for %s — cannot confirm", stays, cardBranch)
	}
	switch strings.ToUpper(pr.State) {
	case "OPEN":
		_, _ = fmt.Fprintf(out, "%s %s pr=%s — the pull request is still open; nothing recorded, run complete again after it merges\n", card.CardID, card.State, pr.URL)
		return nil
	case "MERGED":
	default:
		return fmt.Errorf("factory complete: %s; pull request #%d is %s without merging — an operator decides", stays, pr.Number, strings.ToLower(pr.State))
	}
	tip, err := factoryGitRead(wt, "rev-parse", "--verify", "--quiet", "HEAD^{commit}")
	if err != nil {
		return fmt.Errorf("factory complete: %s; cannot read the card tip: %w", stays, err)
	}
	if pr.HeadRefOid != tip {
		return fmt.Errorf("factory complete: %s; pull request #%d merged head %s but the card worktree holds %s — what merged is not what the card holds, cannot confirm", stays, pr.Number, pr.HeadRefOid, tip)
	}
	if pr.MergeCommit == nil || strings.TrimSpace(pr.MergeCommit.Oid) == "" {
		return fmt.Errorf("factory complete: %s; pull request #%d reads MERGED but names no merge commit — cannot confirm", stays, pr.Number)
	}
	if _, err := factoryGitRead(wt, "fetch", "origin", target); err != nil {
		return fmt.Errorf("factory complete: %s; cannot fetch origin/%s to verify the merge commit: %w", stays, target, err)
	}
	merged, err := db.Transition(ctx, homestate.TransitionRequest{
		RunID: runID, CardID: card.CardID, To: homestate.CardMergedPR,
		ExpectedVersion: card.Version, Actor: lane, MergeSHA: pr.MergeCommit.Oid, IntegrationBranch: target,
		PRNumber: strconv.Itoa(pr.Number), Now: factoryCardNow(),
	})
	if err != nil {
		return fmt.Errorf("factory complete: %s; the F1 record refused merged-pr: %w", stays, err)
	}
	_, _ = fmt.Fprintf(out, "%s %s merge=%s pr=%s branch=%s base=%s\n", merged.CardID, merged.State, merged.MergeSHA, pr.URL, cardBranch, target)
	factoryPrintClearPolicyLine(out, root)
	return nil
}
