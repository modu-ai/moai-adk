package git

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// HasConflicts checks whether merging the target branch into the current
// branch would produce conflicts. This is a read-only dry-run operation
// that does not modify the working tree, staging area, or HEAD.
func (b *branchManager) HasConflicts(target string) (bool, error) {
	b.logger.Debug("checking conflicts", "target", target)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if the target branch exists.
	if !branchExists(ctx, b.root, target) {
		return false, fmt.Errorf("conflict check for %q: %w", target, ErrBranchNotFound)
	}

	// Get the current branch.
	current, err := currentBranch(ctx, b.root)
	if err != nil {
		return false, fmt.Errorf("conflict check: %w", err)
	}

	// Find the merge base.
	base, err := b.MergeBase(current, target)
	if err != nil {
		return false, fmt.Errorf("conflict check: %w", err)
	}

	// Get files changed on each side since the merge base.
	currentFiles, err := changedFiles(ctx, b.root, base, current)
	if err != nil {
		return false, fmt.Errorf("conflict check current files: %w", err)
	}

	targetFiles, err := changedFiles(ctx, b.root, base, target)
	if err != nil {
		return false, fmt.Errorf("conflict check target files: %w", err)
	}

	// Build a set of files changed in the current branch.
	currentSet := make(map[string]bool, len(currentFiles))
	for _, f := range currentFiles {
		currentSet[f] = true
	}

	// Check for overlapping files (both sides modified the same file).
	for _, f := range targetFiles {
		if currentSet[f] {
			b.logger.Debug("conflict detected", "file", f, "current", current, "target", target)
			return true, nil
		}
	}

	b.logger.Debug("no conflicts detected", "current", current, "target", target)
	return false, nil
}

// MergeBase returns the common ancestor commit hash of two branches.
func (b *branchManager) MergeBase(branch1, branch2 string) (string, error) {
	b.logger.Debug("finding merge base", "branch1", branch1, "branch2", branch2)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := execGit(ctx, b.root, "merge-base", branch1, branch2)
	if err != nil {
		// git merge-base exits with code 1 if no merge base is found.
		return "", fmt.Errorf("merge base %s %s: %w", branch1, branch2, ErrNoMergeBase)
	}

	b.logger.Debug("merge base found", "hash", out)
	return out, nil
}

// changedFiles returns the list of files changed between two refs.
func changedFiles(ctx context.Context, dir, ref1, ref2 string) ([]string, error) {
	out, err := execGit(ctx, dir, "diff", "--name-only", ref1, ref2)
	if err != nil {
		return nil, fmt.Errorf("changed files %s..%s: %w", ref1, ref2, err)
	}

	if out == "" {
		return nil, nil
	}

	var files []string
	for line := range strings.SplitSeq(out, "\n") {
		f := strings.TrimSpace(line)
		if f != "" {
			files = append(files, f)
		}
	}
	return files, nil
}
