package git

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Compile-time interface compliance check.
var _ WorktreeManager = (*worktreeManager)(nil)

// worktreeManager implements the WorktreeManager interface using system git.
type worktreeManager struct {
	root   string
	logger *slog.Logger
}

// NewWorktreeManager creates a new WorktreeManager for the repository at root.
func NewWorktreeManager(root string) *worktreeManager {
	return &worktreeManager{
		root:   root,
		logger: slog.Default().With("module", "git.worktree"),
	}
}

// Add creates a new worktree at the given path for the given branch.
// If the branch does not exist, it is created automatically with -b.
func (w *worktreeManager) Add(path, branch string) error {
	w.logger.Info("system git fallback", "operation", "worktree add", "reason", "go-git lacks worktree support")
	w.logger.Debug("adding worktree", "path", path, "branch", branch)

	// Check if path already exists.
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("add worktree at %q: %w", path, ErrWorktreePathExists)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if the branch already exists.
	if branchExists(ctx, w.root, branch) {
		_, err := execGit(ctx, w.root, "worktree", "add", path, branch)
		if err != nil {
			return fmt.Errorf("add worktree for existing branch %q: %w", branch, err)
		}
	} else {
		// Create a new branch with -b.
		_, err := execGit(ctx, w.root, "worktree", "add", "-b", branch, path)
		if err != nil {
			return fmt.Errorf("add worktree with new branch %q: %w", branch, err)
		}
	}

	w.logger.Debug("worktree added", "path", path, "branch", branch)
	return nil
}

// List returns all active worktrees including the main worktree.
func (w *worktreeManager) List() ([]Worktree, error) {
	w.logger.Info("system git fallback", "operation", "worktree list", "reason", "go-git lacks worktree support")
	w.logger.Debug("listing worktrees")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := execGit(ctx, w.root, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("list worktrees: %w", err)
	}

	worktrees := parsePorcelainWorktreeList(out)
	w.logger.Debug("worktrees listed", "count", len(worktrees))
	return worktrees, nil
}

// Remove deletes a worktree at the given path.
// If force is true, the worktree is removed even with uncommitted changes.
func (w *worktreeManager) Remove(path string, force bool) error {
	w.logger.Info("system git fallback", "operation", "worktree remove", "reason", "go-git lacks worktree support")
	w.logger.Debug("removing worktree", "path", path, "force", force)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)
	_, err := execGit(ctx, w.root, args...)
	if err != nil {
		errStr := err.Error()
		switch {
		case strings.Contains(errStr, "is not a working tree"):
			return fmt.Errorf("remove worktree at %q: %w", path, ErrWorktreeNotFound)
		case strings.Contains(errStr, "contains modified or untracked files"):
			return fmt.Errorf("remove worktree at %q: %w", path, ErrWorktreeDirty)
		default:
			return fmt.Errorf("remove worktree at %q: %w", path, err)
		}
	}

	w.logger.Debug("worktree removed", "path", path)
	return nil
}

// Prune removes stale worktree references for deleted directories.
func (w *worktreeManager) Prune() error {
	w.logger.Info("system git fallback", "operation", "worktree prune", "reason", "go-git lacks worktree support")
	w.logger.Debug("pruning worktrees")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := execGit(ctx, w.root, "worktree", "prune")
	if err != nil {
		return fmt.Errorf("prune worktrees: %w", err)
	}

	w.logger.Debug("worktrees pruned")
	return nil
}

// Repair repairs worktree administrative files if they have become corrupted.
func (w *worktreeManager) Repair() error {
	w.logger.Info("system git fallback", "operation", "worktree repair", "reason", "go-git lacks worktree support")
	w.logger.Debug("repairing worktrees")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := execGit(ctx, w.root, "worktree", "repair")
	if err != nil {
		return fmt.Errorf("repair worktrees: %w", err)
	}

	w.logger.Debug("worktrees repaired")
	return nil
}

// Root returns the repository root path.
func (w *worktreeManager) Root() string {
	return w.root
}

// Sync fetches the latest changes from origin and merges or rebases the
// base branch into the worktree at wtPath.
func (w *worktreeManager) Sync(wtPath, baseBranch, strategy string) error {
	w.logger.Debug("syncing worktree", "path", wtPath, "base", baseBranch, "strategy", strategy)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch latest from origin.
	if _, err := execGit(ctx, wtPath, "fetch", "origin"); err != nil {
		return fmt.Errorf("fetch origin in %q: %w", wtPath, err)
	}

	// Apply sync strategy.
	ref := "origin/" + baseBranch
	switch strategy {
	case "rebase":
		if _, err := execGit(ctx, wtPath, "rebase", ref); err != nil {
			return fmt.Errorf("rebase %s in %q: %w", ref, wtPath, err)
		}
	default: // merge
		if _, err := execGit(ctx, wtPath, "merge", ref, "--no-edit"); err != nil {
			return fmt.Errorf("merge %s in %q: %w", ref, wtPath, err)
		}
	}

	w.logger.Debug("worktree synced", "path", wtPath, "base", baseBranch)
	return nil
}

// DeleteBranch deletes a local branch by name.
func (w *worktreeManager) DeleteBranch(name string) error {
	w.logger.Debug("deleting branch", "name", name)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := execGit(ctx, w.root, "branch", "-d", name); err != nil {
		return fmt.Errorf("delete branch %q: %w", name, err)
	}

	w.logger.Debug("branch deleted", "name", name)
	return nil
}

// IsBranchMerged checks whether a branch has been fully merged into base.
func (w *worktreeManager) IsBranchMerged(branch, base string) (bool, error) {
	w.logger.Debug("checking if branch merged", "branch", branch, "base", base)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := execGit(ctx, w.root, "branch", "--merged", base)
	if err != nil {
		return false, fmt.Errorf("check merged branches: %w", err)
	}

	for line := range strings.SplitSeq(out, "\n") {
		name := strings.TrimSpace(strings.TrimPrefix(line, "* "))
		if name == branch {
			return true, nil
		}
	}
	return false, nil
}

// parsePorcelainWorktreeList parses the output of git worktree list --porcelain.
//
// The porcelain format consists of stanzas separated by blank lines:
//
//	worktree /path/to/worktree
//	HEAD abc123def456
//	branch refs/heads/main
//
//	worktree /tmp/wt-feature
//	HEAD def789abc012
//	branch refs/heads/feature
func parsePorcelainWorktreeList(output string) []Worktree {
	var worktrees []Worktree
	var current Worktree

	lines := strings.SplitSeq(output, "\n")
	for line := range lines {
		switch {
		case strings.HasPrefix(line, "worktree "):
			// Save the previous entry if present.
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			// Normalize the path from git output to use OS-native separators
			current = Worktree{Path: filepath.Clean(strings.TrimPrefix(line, "worktree "))}
		case strings.HasPrefix(line, "HEAD "):
			current.HEAD = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "detached":
			current.Branch = ""
		}
	}

	// Append the last entry.
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees
}
