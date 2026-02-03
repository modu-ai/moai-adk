package git

import (
	"context"
	"fmt"
	"log/slog"
	"os"
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
func (w *worktreeManager) Remove(path string) error {
	w.logger.Info("system git fallback", "operation", "worktree remove", "reason", "go-git lacks worktree support")
	w.logger.Debug("removing worktree", "path", path)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := execGit(ctx, w.root, "worktree", "remove", path)
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

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "worktree "):
			// Save the previous entry if present.
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			current = Worktree{Path: strings.TrimPrefix(line, "worktree ")}
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
