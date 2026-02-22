package git

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// Compile-time interface compliance check.
var _ BranchManager = (*branchManager)(nil)

// branchManager implements the BranchManager interface using the system git binary.
type branchManager struct {
	root   string
	logger *slog.Logger
	mu     sync.Mutex
}

// NewBranchManager creates a new BranchManager for the repository at root.
func NewBranchManager(root string) *branchManager {
	return &branchManager{
		root:   root,
		logger: slog.Default().With("module", "git.branch"),
	}
}

// Create creates a new local branch from the current HEAD.
func (b *branchManager) Create(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.logger.Debug("creating branch", "name", name)

	if err := validateBranchName(name); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if branchExists(ctx, b.root, name) {
		return fmt.Errorf("create branch %q: %w", name, ErrBranchExists)
	}

	_, err := execGit(ctx, b.root, "branch", name)
	if err != nil {
		return fmt.Errorf("create branch %q: %w", name, err)
	}

	b.logger.Debug("branch created", "name", name)
	return nil
}

// Switch checks out the specified branch.
func (b *branchManager) Switch(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.logger.Debug("switching branch", "name", name)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if the target branch exists.
	if !branchExists(ctx, b.root, name) {
		return fmt.Errorf("switch to branch %q: %w", name, ErrBranchNotFound)
	}

	// Check if the working tree is clean.
	clean, err := isWorkingTreeClean(ctx, b.root)
	if err != nil {
		return fmt.Errorf("switch branch: %w", err)
	}
	if !clean {
		return fmt.Errorf("switch to branch %q: %w", name, ErrDirtyWorkingTree)
	}

	_, err = execGit(ctx, b.root, "checkout", name)
	if err != nil {
		return fmt.Errorf("switch to branch %q: %w", name, err)
	}

	b.logger.Debug("branch switched", "name", name)
	return nil
}

// Delete removes a local branch.
func (b *branchManager) Delete(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.logger.Debug("deleting branch", "name", name)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if the branch exists.
	if !branchExists(ctx, b.root, name) {
		return fmt.Errorf("delete branch %q: %w", name, ErrBranchNotFound)
	}

	// Check if it is the current branch.
	current, err := currentBranch(ctx, b.root)
	if err == nil && current == name {
		return fmt.Errorf("delete branch %q: %w", name, ErrCannotDeleteCurrentBranch)
	}

	_, err = execGit(ctx, b.root, "branch", "-d", name)
	if err != nil {
		return fmt.Errorf("delete branch %q: %w", name, err)
	}

	b.logger.Debug("branch deleted", "name", name)
	return nil
}

// List returns all local branches with their current status.
func (b *branchManager) List() ([]Branch, error) {
	b.logger.Debug("listing branches")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get the current branch name (may fail if detached HEAD).
	current, _ := currentBranch(ctx, b.root)

	out, err := execGit(ctx, b.root, "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, fmt.Errorf("list branches: %w", err)
	}

	if out == "" {
		return nil, nil
	}

	var branches []Branch
	lines := strings.SplitSeq(out, "\n")
	for line := range lines {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		branches = append(branches, Branch{
			Name:      name,
			IsRemote:  false,
			IsCurrent: name == current,
		})
	}

	b.logger.Debug("branches listed", "count", len(branches))
	return branches, nil
}

// validateBranchName checks if a branch name follows Git ref naming rules.
func validateBranchName(name string) error {
	if name == "" {
		return fmt.Errorf("empty branch name: %w", ErrInvalidBranchName)
	}
	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("branch name starts with '.': %w", ErrInvalidBranchName)
	}
	if strings.HasSuffix(name, ".lock") {
		return fmt.Errorf("branch name ends with '.lock': %w", ErrInvalidBranchName)
	}
	if strings.HasSuffix(name, "/") {
		return fmt.Errorf("branch name ends with '/': %w", ErrInvalidBranchName)
	}
	if strings.HasSuffix(name, ".") {
		return fmt.Errorf("branch name ends with '.': %w", ErrInvalidBranchName)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("branch name contains '..': %w", ErrInvalidBranchName)
	}
	if strings.Contains(name, "~") {
		return fmt.Errorf("branch name contains '~': %w", ErrInvalidBranchName)
	}
	if strings.Contains(name, "^") {
		return fmt.Errorf("branch name contains '^': %w", ErrInvalidBranchName)
	}
	if strings.Contains(name, ":") {
		return fmt.Errorf("branch name contains ':': %w", ErrInvalidBranchName)
	}
	if strings.Contains(name, " ") {
		return fmt.Errorf("branch name contains space: %w", ErrInvalidBranchName)
	}
	if strings.Contains(name, "\\") {
		return fmt.Errorf("branch name contains '\\': %w", ErrInvalidBranchName)
	}
	if strings.Contains(name, "@{") {
		return fmt.Errorf("branch name contains '@{': %w", ErrInvalidBranchName)
	}
	// Check for ASCII control characters.
	for _, c := range name {
		if c < 0x20 || c == 0x7f {
			return fmt.Errorf("branch name contains control character: %w", ErrInvalidBranchName)
		}
	}
	return nil
}
