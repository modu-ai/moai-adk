package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Manager handles worktree lifecycle operations
type Manager struct {
	RepoDir      string
	WorktreeRoot string
	Registry     *Registry
}

// NewManager creates a worktree manager for the given repository
func NewManager(repoDir string) (*Manager, error) {
	repoName := filepath.Base(repoDir)
	wtRoot, err := detectWorktreeRoot(repoName)
	if err != nil {
		return nil, err
	}

	reg, err := NewRegistry(wtRoot, repoName)
	if err != nil {
		return nil, err
	}

	return &Manager{
		RepoDir:      repoDir,
		WorktreeRoot: wtRoot,
		Registry:     reg,
	}, nil
}

// Create creates a new worktree for a SPEC ID
func (m *Manager) Create(specID, branch, baseBranch string, force bool) (*WorktreeInfo, error) {
	if branch == "" {
		branch = "feature/" + specID
	}
	if baseBranch == "" {
		baseBranch = "main"
	}

	// Check if already exists
	existing := m.Registry.Get(specID)
	if existing != nil && !force {
		return nil, fmt.Errorf("worktree for %s already exists at %s (use --force to override)", specID, existing.Path)
	}

	wtPath := filepath.Join(m.WorktreeRoot, specID)

	// Ensure worktree root exists
	if err := os.MkdirAll(m.WorktreeRoot, 0755); err != nil {
		return nil, fmt.Errorf("failed to create worktree root: %w", err)
	}

	// Fetch latest
	_ = gitCmd(m.RepoDir, "fetch", "origin")

	// Create the worktree with new branch
	args := []string{"worktree", "add", "-b", branch, wtPath, baseBranch}
	if err := gitCmd(m.RepoDir, args...); err != nil {
		// Branch might already exist, try without -b
		args = []string{"worktree", "add", wtPath, branch}
		if err2 := gitCmd(m.RepoDir, args...); err2 != nil {
			return nil, fmt.Errorf("failed to create worktree: %w (also tried existing branch: %v)", err, err2)
		}
	}

	info := NewWorktreeInfo(specID, wtPath, branch)
	if err := m.Registry.Register(info); err != nil {
		return nil, fmt.Errorf("failed to register worktree: %w", err)
	}

	return info, nil
}

// Remove removes a worktree by SPEC ID
func (m *Manager) Remove(specID string, force bool) error {
	info := m.Registry.Get(specID)
	if info == nil {
		return fmt.Errorf("worktree not found: %s", specID)
	}

	// Remove via git
	args := []string{"worktree", "remove", info.Path}
	if force {
		args = append(args, "--force")
	}
	if err := gitCmd(m.RepoDir, args...); err != nil {
		// Fallback: remove directory
		if force {
			if removeErr := os.RemoveAll(info.Path); removeErr != nil {
				return fmt.Errorf("failed to remove worktree: git: %v, rm: %v", err, removeErr)
			}
		} else {
			return fmt.Errorf("failed to remove worktree (use --force): %w", err)
		}
	}

	return m.Registry.Unregister(specID)
}

// List returns all worktrees for the current project
func (m *Manager) List() []*WorktreeInfo {
	return m.Registry.List()
}

// Done completes a worktree workflow: merge into base and remove
func (m *Manager) Done(specID, baseBranch string, push, force bool) (map[string]string, error) {
	if baseBranch == "" {
		baseBranch = "main"
	}

	info := m.Registry.Get(specID)
	if info == nil {
		return nil, fmt.Errorf("worktree not found: %s", specID)
	}

	mergedBranch := info.Branch

	// Fetch latest
	_ = gitCmd(m.RepoDir, "fetch", "origin")

	// Checkout base branch in main repo
	if err := gitCmd(m.RepoDir, "checkout", baseBranch); err != nil {
		return nil, fmt.Errorf("failed to checkout %s: %w", baseBranch, err)
	}

	// Merge worktree branch
	mergeMsg := fmt.Sprintf("Merge %s into %s", mergedBranch, baseBranch)
	if err := gitCmd(m.RepoDir, "merge", mergedBranch, "--no-ff", "-m", mergeMsg); err != nil {
		// Abort merge on conflict
		_ = gitCmd(m.RepoDir, "merge", "--abort")
		return nil, fmt.Errorf("merge conflict: %w", err)
	}

	// Push if requested
	pushed := "false"
	if push {
		if err := gitCmd(m.RepoDir, "push", "origin", baseBranch); err != nil {
			return nil, fmt.Errorf("failed to push: %w", err)
		}
		pushed = "true"
	}

	// Remove worktree
	if err := m.Remove(specID, force); err != nil {
		return nil, fmt.Errorf("failed to remove worktree after merge: %w", err)
	}

	// Delete branch (best effort)
	_ = gitCmd(m.RepoDir, "branch", "-d", mergedBranch)

	return map[string]string{
		"merged_branch": mergedBranch,
		"base_branch":   baseBranch,
		"pushed":        pushed,
	}, nil
}

// Sync merges base branch into a worktree
func (m *Manager) Sync(specID, baseBranch string, rebase bool) error {
	if baseBranch == "" {
		baseBranch = "main"
	}

	info := m.Registry.Get(specID)
	if info == nil {
		return fmt.Errorf("worktree not found: %s", specID)
	}

	// Fetch in worktree
	_ = gitCmd(info.Path, "fetch", "origin")

	target := "origin/" + baseBranch
	if rebase {
		return gitCmd(info.Path, "rebase", target)
	}
	return gitCmd(info.Path, "merge", target)
}

// Recover scans disk for existing worktrees and re-registers them
func (m *Manager) Recover() (int, error) {
	count := 0
	entries, err := os.ReadDir(m.WorktreeRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		dirPath := filepath.Join(m.WorktreeRoot, entry.Name())
		gitPath := filepath.Join(dirPath, ".git")
		if _, statErr := os.Stat(gitPath); statErr != nil {
			continue
		}

		specID := entry.Name()
		if m.Registry.Get(specID) != nil {
			continue // Already registered
		}

		// Try to get branch name
		branch := getBranchName(dirPath)
		info := &WorktreeInfo{
			SpecID:       specID,
			Path:         dirPath,
			Branch:       branch,
			CreatedAt:    "",
			LastAccessed: "",
			Status:       "recovered",
		}
		if err := m.Registry.Register(info); err != nil {
			continue
		}
		count++
	}
	return count, nil
}

// CleanMerged removes worktrees whose branches have been merged into base
func (m *Manager) CleanMerged(baseBranch string) (int, error) {
	if baseBranch == "" {
		baseBranch = "main"
	}

	// Get merged branches
	out, err := gitOutput(m.RepoDir, "branch", "--merged", baseBranch)
	if err != nil {
		return 0, fmt.Errorf("failed to list merged branches: %w", err)
	}

	mergedSet := make(map[string]bool)
	for _, line := range strings.Split(out, "\n") {
		branch := strings.TrimSpace(strings.TrimPrefix(line, "* "))
		if branch != "" {
			mergedSet[branch] = true
		}
	}

	count := 0
	for _, info := range m.Registry.List() {
		if mergedSet[info.Branch] {
			if err := m.Remove(info.SpecID, true); err == nil {
				count++
			}
		}
	}
	return count, nil
}

// detectWorktreeRoot determines the worktree root directory
func detectWorktreeRoot(repoName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	primary := filepath.Join(home, ".moai", "worktrees", repoName)

	// Check if primary location exists and has content
	if _, statErr := os.Stat(filepath.Join(primary, ".moai-worktree-registry.json")); statErr == nil {
		return primary, nil
	}

	// Default to primary
	return primary, nil
}

// gitCmd runs a git command in the specified directory
func gitCmd(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// gitOutput runs a git command and returns its stdout
func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// getBranchName reads the current branch from a git directory
func getBranchName(dir string) string {
	branch, err := gitOutput(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "unknown"
	}
	return branch
}
