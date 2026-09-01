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
		base, err := resolveReviewMergeBase(root)
		if err != nil {
			return nil, err
		}
		return []string{"diff", base + "...HEAD"}, nil
	case codexTargetUncommitted, "":
		return []string{"diff", "HEAD"}, nil
	default:
		return nil, fmt.Errorf("unknown review target %q", target)
	}
}

// resolveReviewMergeBase finds the commit this branch diverged from. It tries
// the remote default branch first (the tree a pull request would be measured
// against) and falls back to the local one, because a worktree cut for a card
// may have no remote-tracking ref for its base yet.
func resolveReviewMergeBase(root string) (string, error) {
	var lastErr error
	for _, ref := range []string{"origin/HEAD", "origin/main", "main"} {
		out, err := runReviewGit(root, "merge-base", ref, "HEAD")
		if err == nil && strings.TrimSpace(out) != "" {
			return strings.TrimSpace(out), nil
		}
		lastErr = err
	}
	return "", fmt.Errorf("cannot resolve the base commit in %s: %w", root, lastErr)
}

// resolveReviewBaseBranchName returns the base branch NAME for the tree at root.
//
// It is the name-layer sibling of resolveReviewMergeBase, and reads the SAME
// fallback chain in the same order (SPEC-CODEX-REVIEW-TARGET-001 §A.7): the
// remote default head first, then `main`. The two functions exist separately
// because they answer different questions — a merge base is a commit, and
// codex's baseBranch review target is a branch name — not because the backends
// disagree about which base to use. They must not diverge: a codex review and a
// GLM review asked for the same target have to be looking at the same change,
// and the asymmetry that produced this SPEC is exactly what a second chain
// would recreate in the opposite direction.
//
// git_strategy.worktree_base_branch is deliberately NOT read here. If that key
// should be the base, it should be the base for BOTH backends, which is a change
// to resolveReviewMergeBase and a different card.
//
// resolveReviewMergeBase's chain lists origin/main and main as separate steps
// because each names a different ref to compute a merge base FROM. At the name
// layer both produce the same string, and two steps that no observation can tell
// apart are one step, so they are merged here.
//
// [HARD] A name is returned only after it is confirmed to resolve as a ref in
// this tree. Stripping the `origin/` prefix off a symbolic-ref yields a string,
// not a guarantee: the remote-tracking ref can exist while nothing by that name
// does. Returning an unconfirmed name would send codex to review against a
// branch it cannot find, and that failure reappears as an `inconclusive` in the
// very place this SPEC closed one.
func resolveReviewBaseBranchName(root string) (string, error) {
	// 1. the remote default head
	if out, err := runReviewGit(root, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil {
		name := strings.TrimPrefix(strings.TrimSpace(out), "origin/")
		if name != "" && reviewRefResolves(root, name) {
			return name, nil
		}
	}
	// 2. main — as a remote-tracking ref or as a local branch
	if reviewRefResolves(root, "main") {
		return "main", nil
	}
	return "", fmt.Errorf("cannot resolve a base branch in %s", root)
}

// reviewRefResolves reports whether name resolves as a branch in the tree,
// looking at local heads and origin's remote-tracking refs. The refs are named
// in full rather than handed to git as a bare revision so a file or directory
// sharing the name cannot be mistaken for a branch.
func reviewRefResolves(root, name string) bool {
	for _, ref := range []string{"refs/heads/" + name, "refs/remotes/origin/" + name} {
		if _, err := runReviewGit(root, "rev-parse", "--verify", "--quiet", ref); err == nil {
			return true
		}
	}
	return false
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
