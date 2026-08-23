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
