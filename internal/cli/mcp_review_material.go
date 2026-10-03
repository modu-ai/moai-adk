// Package cli — review material for backends that cannot read the tree.
//
// The codex backend is a subprocess launched inside the tree under review: it
// reads the working copy itself, and `target` tells it what to look at. The GLM
// backend is an HTTPS call to z.ai. It has no working directory, no filesystem,
// and no way to fetch anything — so whatever it is to review has to travel in
// the request body.
//
// Nothing sent it any. The audit prompt said "Review the proposed change" and
// stopped there, and the model, having no change in front of it, answered from
// imagination: a live run returned a confident `fail` citing a repository
// whitelist that does not exist in this codebase (card t178). A backend given
// nothing does not report that it has nothing; it invents something to report.
//
// This file collects the change as a diff so the request carries the code. When
// no diff can be produced, the caller must fail open to inconclusive rather than
// ask for a verdict anyway — an honest "I could not tell" is worth more than a
// confident answer about nothing.
package cli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// reviewDiffMaxBytes bounds the diff carried into a backend request. A review is
// a focused pass over a change, not a whole-repository upload: past this size
// the request stops being a review request and starts being a way to exhaust the
// response budget before the model reaches the interesting hunk. Oversized diffs
// are truncated with a visible marker so the model — and anyone reading the
// request — knows the material is partial rather than complete.
const reviewDiffMaxBytes = 200_000

// reviewDiffTruncationNote is appended to a truncated diff. It is prose rather
// than a comment marker because its reader is a language model, and it must not
// be mistakable for part of the patch.
const reviewDiffTruncationNote = "\n\n[diff truncated: the change exceeds the review size limit; the tail is not shown]\n"

// collectReviewDiff returns the change named by target, as a unified diff read
// from the git tree at root.
//
// The two targets mirror the codex ones so both backends review the same thing
// when asked for the same target:
//
//   - uncommittedChanges — everything not yet committed, staged and unstaged
//     alike (`git diff HEAD`), which is what a pre-commit review looks at.
//   - baseBranch — the branch's own commits, measured from the merge base so an
//     unrelated advance of the base does not appear in the review.
//
// An error means no material could be produced: root is not a git tree, git is
// absent, or the command failed. An empty diff with no error means the tree is
// genuinely clean. Callers MUST treat both as "nothing to review" — never as a
// reason to ask for a verdict anyway.
func collectReviewDiff(root, target string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("no project root to read the change from")
	}
	args, err := reviewDiffArgs(root, target)
	if err != nil {
		return "", err
	}
	out, err := runReviewGit(root, args...)
	if err != nil {
		return "", fmt.Errorf("cannot read the change from %s: %w", root, err)
	}
	return truncateDiff(out), nil
}

// reviewDiffArgs maps a review target onto the git arguments that produce it.
// baseBranch needs the merge base resolved first, which is a second git call and
// is why this returns args rather than a single fixed command.
func reviewDiffArgs(root, target string) ([]string, error) {
	switch target {
	case codexTargetBaseBranch:
		base, err := resolveReviewBase(root)
		if err != nil {
			return nil, err
		}
		return []string{"diff", base.MergeBase + "...HEAD"}, nil
	case codexTargetUncommitted, "":
		return []string{"diff", "HEAD"}, nil
	default:
		return nil, fmt.Errorf("unknown review target %q", target)
	}
}

// collectReviewDiffAt is collectReviewDiff for a baseBranch review whose base
// the caller has ALREADY resolved: the diff is measured from exactly that merge
// base, so the base a result reports and the material the backend saw cannot
// drift apart (card t1426 — a base ref may move while the model is thinking).
func collectReviewDiffAt(root string, base reviewBase) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("no project root to read the change from")
	}
	out, err := runReviewGit(root, "diff", base.MergeBase+"...HEAD")
	if err != nil {
		return "", fmt.Errorf("cannot read the change from %s: %w", root, err)
	}
	return truncateDiff(out), nil
}

// reviewBase is ONE resolution of the base a baseBranch review is measured
// against: the ref that was actually selected and the merge-base COMMIT both
// backends compare from — GLM measures its diff from it, codex is sent it as the
// baseBranch target. Resolved together so the backends can never land on
// different steps of the chain (card t1426).
//
// Ref is the selected ref in its short, resolvable form — `develop` for a local
// branch, `origin/develop` for a remote-tracking one — never a bare name that
// was only true of a different ref.
type reviewBase struct {
	Ref       string
	MergeBase string
}

// String renders the base for a result's review_base field.
func (b reviewBase) String() string {
	return b.Ref + " (merge base " + b.MergeBase + ")"
}

// shortReviewRef turns a full ref the resolver selected into the short form git
// resolves to that same ref: refs/heads/X → X, refs/remotes/O/X → O/X.
func shortReviewRef(ref string) string {
	if s, ok := strings.CutPrefix(ref, "refs/heads/"); ok {
		return s
	}
	if s, ok := strings.CutPrefix(ref, "refs/remotes/"); ok {
		return s
	}
	return ref
}

// resolveReviewBase walks the base chain once and returns the first step whose
// branch resolves as a ref in this tree AND shares a merge base with HEAD. A
// step that names an existing branch with no common history falls through —
// for both backends at once, because both read this one result.
//
//  0. git_strategy.worktree_base_branch (card t1426), when set: in a git-flow
//     repository card branches are cut from `develop` while the remote default
//     head stays `main`, and measuring a card against main reviews everything
//     develop carries that main lacks. Local branch first, then origin's.
//  1. the remote default head (the tree a pull request would be measured
//     against), named by stripping `origin/` off refs/remotes/origin/HEAD.
//  2. main — origin's remote-tracking ref first, then the local branch, because
//     a worktree cut for a card may have no remote-tracking ref for its base yet.
//
// A project that never configured step 0 resolves as before.
//
// [HARD] A ref is returned only after it is confirmed to resolve in this tree
// (SPEC-CODEX-REVIEW-TARGET-001), and it is returned as the ref that was
// selected: a base that exists only as origin/develop is reported as
// origin/develop, because the bare name `develop` does not resolve there.
func resolveReviewBase(root string) (reviewBase, error) {
	// Each step lists the refs it may be satisfied by, in preference order.
	var steps [][]string
	if name := config.LoadWorktreeBaseBranch(root); name != "" {
		steps = append(steps, []string{"refs/heads/" + name, "refs/remotes/origin/" + name})
	}
	if out, err := runReviewGit(root, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil {
		if name := strings.TrimPrefix(strings.TrimSpace(out), "origin/"); name != "" {
			steps = append(steps, []string{"refs/remotes/origin/" + name, "refs/heads/" + name})
		}
	}
	steps = append(steps, []string{"refs/remotes/origin/main", "refs/heads/main"})

	var lastErr error
	for _, refs := range steps {
		for _, ref := range refs {
			if _, err := runReviewGit(root, "rev-parse", "--verify", "--quiet", ref); err != nil {
				continue // the name does not resolve through this ref
			}
			out, err := runReviewGit(root, "merge-base", ref, "HEAD")
			if err == nil && strings.TrimSpace(out) != "" {
				return reviewBase{Ref: shortReviewRef(ref), MergeBase: strings.TrimSpace(out)}, nil
			}
			lastErr = err
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no base branch resolves")
	}
	return reviewBase{}, fmt.Errorf("cannot resolve the base commit in %s: %w", root, lastErr)
}

// resolveReviewMergeBase is the commit half of resolveReviewBase.
func resolveReviewMergeBase(root string) (string, error) {
	b, err := resolveReviewBase(root)
	return b.MergeBase, err
}

// resolveReviewBaseBranchName is the ref half of resolveReviewBase (the
// audit_multi receipt records it); it reads the SAME single resolution as the
// merge base, so the two can never disagree.
func resolveReviewBaseBranchName(root string) (string, error) {
	b, err := resolveReviewBase(root)
	return b.Ref, err
}

// runReviewGit runs one git command in the named tree and returns its stdout. It
// uses -C rather than changing directory: a cd would apply to the process, and
// two backends reviewing two trees run concurrently in this package.
func runReviewGit(root string, args ...string) (string, error) {
	full := append([]string{"-C", root}, args...)
	cmd := exec.Command("git", full...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// truncateDiff bounds a diff to reviewDiffMaxBytes, cutting at a line boundary
// so the tail of the material is not a half-written hunk header.
func truncateDiff(diff string) string {
	if len(diff) <= reviewDiffMaxBytes {
		return diff
	}
	cut := diff[:reviewDiffMaxBytes]
	if nl := strings.LastIndexByte(cut, '\n'); nl > 0 {
		cut = cut[:nl]
	}
	return cut + reviewDiffTruncationNote
}
