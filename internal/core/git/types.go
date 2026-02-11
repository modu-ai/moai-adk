package git

import "time"

// GitStatus holds the working tree state.
type GitStatus struct {
	// Staged lists files with changes in the index (staged for commit).
	Staged []string

	// Modified lists files modified in the working tree but not staged.
	Modified []string

	// Untracked lists files not tracked by Git.
	Untracked []string

	// Ahead is the number of local commits not pushed to upstream.
	Ahead int

	// Behind is the number of upstream commits not pulled locally.
	Behind int
}

// Commit represents a Git commit record.
type Commit struct {
	// Hash is the full SHA-1 hash of the commit.
	Hash string

	// Author is the author name of the commit.
	Author string

	// Date is the author date of the commit.
	Date time.Time

	// Message is the first line (subject) of the commit message.
	Message string
}

// Branch represents a Git branch.
type Branch struct {
	// Name is the short branch name (e.g., "main", "feature/login").
	Name string

	// IsRemote indicates whether this is a remote-tracking branch.
	IsRemote bool

	// IsCurrent indicates whether this is the currently checked-out branch.
	IsCurrent bool
}

// Worktree represents a Git worktree entry.
type Worktree struct {
	// Path is the absolute filesystem path to the worktree.
	Path string

	// Branch is the short branch name checked out in the worktree.
	Branch string

	// HEAD is the commit hash at the worktree's HEAD.
	HEAD string
}

// EventType identifies the kind of Git event detected.
type EventType string

const (
	// EventBranchSwitch indicates a branch checkout occurred.
	EventBranchSwitch EventType = "branch_switch"

	// EventNewCommit indicates a new commit was created.
	EventNewCommit EventType = "new_commit"

	// EventMerge indicates a merge occurred.
	EventMerge EventType = "merge"

	// EventRebase indicates a rebase occurred.
	EventRebase EventType = "rebase"
)

// GitEvent represents a detected Git state change.
type GitEvent struct {
	// Type identifies the kind of event.
	Type EventType

	// PreviousBranch is the branch before the event (for branch switches).
	PreviousBranch string

	// CurrentBranch is the branch after the event.
	CurrentBranch string

	// PreviousHEAD is the HEAD commit hash before the event.
	PreviousHEAD string

	// CurrentHEAD is the HEAD commit hash after the event.
	CurrentHEAD string

	// Timestamp is when the event was detected.
	Timestamp time.Time
}

// Repository provides read operations on a Git repository.
type Repository interface {
	// CurrentBranch returns the name of the currently checked-out branch.
	// Returns ErrDetachedHEAD if HEAD is not on a branch.
	CurrentBranch() (string, error)

	// Status returns the working tree status including staged, modified,
	// and untracked files, plus ahead/behind counts relative to upstream.
	Status() (*GitStatus, error)

	// Log returns the most recent n commits from HEAD, newest first.
	// If n exceeds the total number of commits, all available commits are returned.
	Log(n int) ([]Commit, error)

	// Diff returns the unified diff output between two references.
	// References can be branch names, commit hashes, or expressions like HEAD~N.
	Diff(ref1, ref2 string) (string, error)

	// IsClean returns true if the working tree has no uncommitted changes.
	IsClean() (bool, error)

	// Root returns the absolute path to the repository root directory.
	Root() string
}

// BranchManager provides branch lifecycle operations.
type BranchManager interface {
	// Create creates a new local branch from the current HEAD.
	// Returns ErrBranchExists if the branch already exists.
	// Returns ErrInvalidBranchName if the name violates Git ref rules.
	Create(name string) error

	// Switch checks out the specified branch.
	// Returns ErrBranchNotFound if the branch does not exist.
	// Returns ErrDirtyWorkingTree if there are uncommitted changes.
	Switch(name string) error

	// Delete removes a local branch.
	// Returns ErrCannotDeleteCurrentBranch if the branch is checked out.
	// Returns ErrBranchNotFound if the branch does not exist.
	Delete(name string) error

	// List returns all local branches with their current status.
	List() ([]Branch, error)

	// HasConflicts checks whether merging the target branch into the current
	// branch would produce conflicts. This is a read-only dry-run operation.
	// Returns ErrBranchNotFound if the target branch does not exist.
	HasConflicts(target string) (bool, error)

	// MergeBase returns the common ancestor commit hash of two branches.
	// Returns ErrNoMergeBase if no common ancestor exists.
	MergeBase(branch1, branch2 string) (string, error)
}

// WorktreeManager manages Git worktrees for parallel development.
type WorktreeManager interface {
	// Add creates a new worktree at the given path for the given branch.
	// If the branch does not exist, it is created automatically.
	// Returns ErrWorktreePathExists if the path already exists.
	// Returns ErrSystemGitNotFound if git is not in PATH.
	Add(path, branch string) error

	// List returns all active worktrees including the main worktree.
	// Returns ErrSystemGitNotFound if git is not in PATH.
	List() ([]Worktree, error)

	// Remove deletes a worktree at the given path.
	// If force is true, the worktree is removed even with uncommitted changes.
	// Returns ErrWorktreeDirty if the worktree has uncommitted changes and force is false.
	// Returns ErrWorktreeNotFound if no worktree exists at the path.
	Remove(path string, force bool) error

	// Prune removes stale worktree references for deleted directories.
	Prune() error

	// Repair repairs worktree administrative files if they have become
	// corrupted or outdated. This runs 'git worktree repair'.
	Repair() error

	// Root returns the repository root path.
	Root() string

	// Sync fetches the latest changes and merges/rebases a base branch
	// into the worktree at the given path.
	// Strategy can be "merge" (default) or "rebase".
	Sync(wtPath, baseBranch, strategy string) error

	// DeleteBranch deletes a local branch by name.
	// Returns an error if the branch does not exist or is currently checked out.
	DeleteBranch(name string) error

	// IsBranchMerged checks whether a branch has been fully merged into
	// the given base branch.
	IsBranchMerged(branch, base string) (bool, error)
}

// Compile-time interface compliance checks are in each implementation file:
// - manager.go: var _ Repository = (*gitManager)(nil)
// - branch.go:  var _ BranchManager = (*branchManager)(nil)
// - worktree.go: var _ WorktreeManager = (*worktreeManager)(nil)
