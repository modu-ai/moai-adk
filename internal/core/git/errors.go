package git

import "errors"

// Sentinel errors for Git operations.
// All errors can be checked with errors.Is().
var (
	// ErrNotRepository indicates the path is not a valid Git repository.
	ErrNotRepository = errors.New("git: not a git repository")

	// ErrDetachedHEAD indicates HEAD is not pointing to a branch.
	ErrDetachedHEAD = errors.New("git: HEAD is detached")

	// ErrBranchExists indicates a branch with the given name already exists.
	ErrBranchExists = errors.New("git: branch already exists")

	// ErrBranchNotFound indicates the specified branch does not exist.
	ErrBranchNotFound = errors.New("git: branch not found")

	// ErrDirtyWorkingTree indicates the working tree has uncommitted changes.
	ErrDirtyWorkingTree = errors.New("git: working tree has uncommitted changes")

	// ErrCannotDeleteCurrentBranch indicates an attempt to delete the checked-out branch.
	ErrCannotDeleteCurrentBranch = errors.New("git: cannot delete currently checked-out branch")

	// ErrNoMergeBase indicates no common ancestor was found between two branches.
	ErrNoMergeBase = errors.New("git: no common ancestor found")

	// ErrWorktreePathExists indicates the worktree path already exists on disk.
	ErrWorktreePathExists = errors.New("git: worktree path already exists")

	// ErrWorktreeDirty indicates the worktree has uncommitted changes.
	ErrWorktreeDirty = errors.New("git: worktree has uncommitted changes")

	// ErrWorktreeNotFound indicates the specified worktree was not found.
	ErrWorktreeNotFound = errors.New("git: worktree not found")

	// ErrSystemGitNotFound indicates the git binary is not in PATH.
	ErrSystemGitNotFound = errors.New("git: system git binary not found")

	// ErrInvalidBranchName indicates the branch name violates Git ref naming rules.
	ErrInvalidBranchName = errors.New("git: invalid branch name")
)
