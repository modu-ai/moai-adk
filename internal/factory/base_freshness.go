package factory

// base_freshness.go — the pre-integration base-freshness gate
// (SPEC-INTEGRATION-STALE-BASE-001, card t1834).
//
// Before a candidate may take the release-integration window, or queue for it,
// the gate asks two questions of the candidate's worktree:
//
//  1. Is there a VALID re-measure record keyed to the candidate's CURRENT tree?
//     An absorb that changes content changes the tree, so a record keyed to the
//     old tree cannot match; a record keyed to the current tree was produced on
//     content that includes everything the lane absorbed.
//  2. Is the base that record names at most K landings behind the local
//     integration branch? A landing is one FIRST-PARENT commit of that branch
//     (an integration merge or a direct commit), counted from the record's
//     stored Base — the very field the merge step's cause 2 compares with the
//     tip (record.Base != tip). At K = 0 the gate therefore tests the predicate
//     cause 2 tests, earlier and without consuming a window turn.
//
// The gate is an early, cheap refusal in front of a strict authority; it never
// replaces that authority. It judges the record through ReadRemeasureRecord and
// ValidateRemeasureRecord, runs no absorb, and writes nothing: no record, no
// window state, no queue ticket, no factory-record field (REQ-ISB-006, -014).
//
// Failure posture: fail-open on uncertainty, fail-closed on positive evidence.
// A configuration, branch, or git read that fails admits the entry with one
// line naming the cause; only a measured L > K, a stale (non-ancestor) base, or
// a missing/invalid record refuses (REQ-ISB-009). The merge step re-enforces
// the strict form, so a wrongly admitted lane costs one window turn.

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/factorylane"
)

const (
	// BaseFreshnessSentinel prefixes every refusal so a caller can match the
	// source without parsing the text (REQ-ISB-011).
	BaseFreshnessSentinel = "BASE_FRESHNESS_VIOLATION:"

	// BaseFreshnessExitRefused is the process exit code of a refusal. It is
	// distinct from the merge step's 1-15 and from complete's own 20
	// (REQ-ISB-010, REQ-ISB-011).
	BaseFreshnessExitRefused = 16

	// BaseFreshnessKey names the tolerance in refusal text and notices.
	BaseFreshnessKey = "workflow.base_freshness_gate.max_behind_landings"

	baseFreshnessNoticePrefix   = "base-freshness: not measured — "
	baseFreshnessAdvisoryPrefix = "base-freshness (advisory, gate off): would refuse — "

	// gitExitNotAncestor is the exit code `git merge-base --is-ancestor` uses to
	// answer "no"; any other failure is an error, not an answer.
	gitExitNotAncestor = 1
)

// BaseFreshnessInput is what the gate judges.
type BaseFreshnessInput struct {
	// Root is the project root holding the re-measure store — the same root the
	// remeasure verb writes to and the merge step reads from.
	Root string
	// Dir is the candidate tree: the worktree whose HEAD tree keys the record.
	Dir string
	// CardID is the card the entry is for. Empty means the gate does not apply
	// (a leader acquiring without a card).
	CardID string
	// IntegrationBranch is the local branch the candidate lands on: the configured
	// integration target, never a fallback to the caller's current branch.
	IntegrationBranch string
	// RequestedBranch is the branch an explicit `--branch` on the refused verb
	// named; empty when the verb was given none. The gate measures the configured
	// IntegrationBranch only, so a request for another branch is a refusal (or,
	// with the gate off, an advisory) and is never measured.
	RequestedBranch string
	// K is the tolerance, read from configuration by the caller. The gate holds
	// no threshold of its own (REQ-ISB-007).
	K int
	// Enabled gates ONLY the refusal; measurement and the advisory line run
	// either way (REQ-ISB-008).
	Enabled bool
	// ConfigErr is non-nil when the caller could not read the configuration the
	// other fields come from: the verdict is then undeterminable.
	ConfigErr error
}

// BaseFreshnessSeams carries the gate's test seam.
type BaseFreshnessSeams struct {
	// Git overrides the git runner. It runs in the candidate tree; nil is a real
	// git bound to BaseFreshnessInput.Dir.
	Git func(args ...string) (string, error)
}

func (s BaseFreshnessSeams) runner(dir string) func(args ...string) (string, error) {
	if s.Git != nil {
		return s.Git
	}
	return factorylane.ExecGitRunner{Dir: dir}.Git
}

// BaseFreshnessResult is the gate's verdict.
type BaseFreshnessResult struct {
	// Applicable is false when the call is outside the gate's scope (no card, the
	// tree is on the integration branch, or HEAD is detached); such a call
	// admits silently.
	Applicable bool
	// Admit is false only for an enabled refusal.
	Admit bool
	// Refusal is the sentinel-prefixed text of an enabled refusal; empty on admit.
	Refusal string
	// Notice is the single line to print on an admit that carries something to
	// say: an uncertainty cause, or the gate-off advisory. Empty when silent.
	Notice string
	// Behind is the landings-behind count, or -1 when it was not measured.
	Behind int
	// Stale reports a stored base that is not an ancestor of the tip.
	Stale bool
	// Tip and Tree are the integration tip and the candidate tree, when read.
	Tip, Tree string
}

// EvaluateBaseFreshness judges one entry.
//
// @MX:ANCHOR: [AUTO] the single base-freshness verdict both window-entry verbs consult
// @MX:REASON: `moai integration acquire`, `moai factory complete`, and the gate tests all call it — a drift here changes what every lane is refused at the window door. Expected fan-in: 3 production-adjacent callers (two adapters, the tests).
// @MX:SPEC: SPEC-INTEGRATION-STALE-BASE-001
func EvaluateBaseFreshness(in BaseFreshnessInput, seams BaseFreshnessSeams) BaseFreshnessResult {
	res := BaseFreshnessResult{Admit: true, Behind: -1}
	if strings.TrimSpace(in.CardID) == "" {
		return res
	}
	// A card id is present, so the call is applicable until the tree proves it
	// is not: an unreadable HEAD is uncertainty, not inapplicability.
	res.Applicable = true
	git := seams.runner(in.Dir)

	head, err := git("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return uncertain(res, gitCause("rev-parse", err))
	}
	// branch is the candidate worktree's real current branch; it is compared with
	// the configured IntegrationBranch, which the caller never derives from it.
	branch := strings.TrimSpace(head)
	if branch == "HEAD" || (in.IntegrationBranch != "" && branch == in.IntegrationBranch) {
		res.Applicable = false
		return res
	}

	switch {
	case in.ConfigErr != nil:
		return uncertain(res, "configuration unreadable: "+in.ConfigErr.Error())
	case in.K < 0:
		return uncertain(res, fmt.Sprintf("%s is negative (%d)", BaseFreshnessKey, in.K))
	case strings.TrimSpace(in.IntegrationBranch) == "":
		return uncertain(res, "the integration branch is not configured")
	}

	// The gate measures the configured integration target only. A request for any
	// other branch is refused (the gate off: advised) before anything is read, so
	// the candidate is never measured against the other branch.
	if requested := strings.TrimSpace(in.RequestedBranch); requested != "" && requested != in.IntegrationBranch {
		reason := fmt.Sprintf("--branch %s differs from the configured integration target %s, and the gate measures the configured integration target only", requested, in.IntegrationBranch)
		return violation(res, in, reason, fmt.Sprintf("Acquire without --branch (the configured target %s), or change the configured integration target, then retry.", in.IntegrationBranch))
	}

	tipOut, err := git("rev-parse", "--verify", integrationBranchRef(in.IntegrationBranch))
	if err != nil {
		return uncertain(res, fmt.Sprintf("integration branch %s is unresolvable (%s)", in.IntegrationBranch, gitCause("rev-parse", err)))
	}
	res.Tip = strings.TrimSpace(tipOut)
	treeOut, err := git("rev-parse", "HEAD^{tree}")
	if err != nil {
		return uncertain(res, gitCause("rev-parse", err))
	}
	res.Tree = strings.TrimSpace(treeOut)

	// The record: judged through the one verifier every merge path consults.
	rec, err := ReadRemeasureRecord(in.Root, res.Tree)
	if err == nil {
		err = ValidateRemeasureRecord(rec)
	}
	if err != nil {
		reason := fmt.Sprintf("no valid re-measure record for tree %s (%s)", ShortSHA(res.Tree), recordCause(rec, err))
		return violation(res, in, reason, "Run moai integration remeasure after the last absorb, then retry.")
	}
	// The record file is keyed by the current tree, but the reader returns the
	// tree field as written (the merge step's cause 4 relies on that), so a
	// forged or corrupted file can name another tree. Only a record that names
	// the tree being judged is a record for it.
	if rec.Tree != res.Tree {
		reason := fmt.Sprintf("the re-measure record stored for tree %s names tree %s, not the current tree", ShortSHA(res.Tree), ShortSHA(rec.Tree))
		return violation(res, in, reason, "Run moai integration remeasure after the last absorb, then retry.")
	}

	// A base the tip does not descend from is stale whatever K is.
	if _, err := git("merge-base", "--is-ancestor", rec.Base, res.Tip); err != nil {
		var exit *factorylane.GitExitError
		if !errors.As(err, &exit) || exit.ExitCode != gitExitNotAncestor {
			return uncertain(res, gitCause("merge-base", err))
		}
		res.Stale = true
		reason := fmt.Sprintf("card %s is stale against local %s (tip %s): the re-measure record's base %s is not an ancestor of the tip, whatever %s=%d",
			in.CardID, in.IntegrationBranch, ShortSHA(res.Tip), ShortSHA(rec.Base), BaseFreshnessKey, in.K)
		return violation(res, in, reason, absorbRemedy(in.IntegrationBranch))
	}

	countOut, err := git("rev-list", "--count", "--first-parent", rec.Base+".."+res.Tip)
	if err != nil {
		return uncertain(res, gitCause("rev-list", err))
	}
	behind, err := strconv.Atoi(strings.TrimSpace(countOut))
	if err != nil {
		return uncertain(res, "git rev-list printed a non-numeric count")
	}
	res.Behind = behind
	if behind > in.K {
		reason := fmt.Sprintf("card %s is %d landings behind local %s (tip %s), more than %s=%d",
			in.CardID, behind, in.IntegrationBranch, ShortSHA(res.Tip), BaseFreshnessKey, in.K)
		return violation(res, in, reason, absorbRemedy(in.IntegrationBranch))
	}
	return res
}

// integrationBranchRef is the one fully qualified spelling of the integration
// branch: the gate reads its tip through it and the re-measure names its absorbed
// base through it, so a tag sharing the branch's name cannot make the two
// disagree (card t1834, sync-audit iter10 H2).
func integrationBranchRef(branch string) string {
	return "refs/heads/" + branch
}

// gitCause renders a git failure as the operation and, for a process exit, its
// code. Git's own text (stderr, the error string) is never forwarded: it quotes
// configuration values and paths the operator did not mean to print.
func gitCause(op string, err error) string {
	var exit *factorylane.GitExitError
	if errors.As(err, &exit) {
		return fmt.Sprintf("git %s failed (exit %d)", op, exit.ExitCode)
	}
	return fmt.Sprintf("git %s could not run", op)
}

// recordCause names why a record is not usable from a fixed phrase set. Record
// fields (command, build identity, failing names, the stored SHAs) are untrusted
// text and are never printed; only the exit code, a number, is.
func recordCause(rec *RemeasureRecord, err error) string {
	switch {
	case rec == nil && strings.HasPrefix(err.Error(), "no re-measure record"):
		return "no re-measure record is stored for it"
	case rec == nil:
		return "the stored re-measure record is unreadable"
	case rec.Tree == "" || rec.Base == "" || rec.Command == "" || rec.BuildIdentity == "" ||
		!isFullSHA(rec.Tree) || !isFullSHA(rec.Base):
		return "the record is incomplete or malformed"
	case rec.ExitCode != 0:
		return fmt.Sprintf("its command exited %d", rec.ExitCode)
	default:
		return "its structured test report is missing or empty"
	}
}

func absorbRemedy(branch string) string {
	return fmt.Sprintf("Absorb local %s into the card branch, re-run moai integration remeasure, then retry.", branch)
}

// violation turns a positive finding into a refusal when the gate is enabled
// and into the one advisory line when it is not.
func violation(res BaseFreshnessResult, in BaseFreshnessInput, reason, remedy string) BaseFreshnessResult {
	if !in.Enabled {
		res.Notice = oneLine(baseFreshnessAdvisoryPrefix + reason)
		return res
	}
	res.Admit = false
	res.Refusal = oneLine(BaseFreshnessSentinel + " integration window refused: " + reason + ". " + remedy)
	return res
}

// uncertain admits with the one line naming why no verdict was reached.
func uncertain(res BaseFreshnessResult, cause string) BaseFreshnessResult {
	res.Notice = oneLine(baseFreshnessNoticePrefix + cause)
	return res
}

// oneLine collapses any whitespace run, newlines included, so a notice or a
// refusal is exactly one line however many lines a git error spans.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
