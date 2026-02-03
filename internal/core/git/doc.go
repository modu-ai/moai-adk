// Package git provides Git repository operations for MoAI-ADK.
//
// It implements three main interfaces:
//   - Repository: read-only operations on a Git repository
//   - BranchManager: branch lifecycle and conflict detection
//   - WorktreeManager: Git worktree management for parallel development
//
// All operations use the system Git binary via os/exec.
// Error handling uses sentinel errors that can be checked with errors.Is().
//
// Usage:
//
//	repo, err := git.NewRepository("/path/to/repo")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	branch, err := repo.CurrentBranch()
//
//	branchMgr := git.NewBranchManager(repo.Root())
//	err = branchMgr.Create("feature/login")
//
//	worktreeMgr := git.NewWorktreeManager(repo.Root())
//	err = worktreeMgr.Add("/tmp/wt", "feature/parallel")
package git
