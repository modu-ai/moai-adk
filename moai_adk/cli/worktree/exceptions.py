"""Custom exceptions for Git Worktree CLI."""

from pathlib import Path


class WorktreeError(Exception):
    """Base exception for worktree operations."""

    pass


class WorktreeExistsError(WorktreeError):
    """Raised when trying to create a worktree that already exists."""

    def __init__(self, spec_id: str, path: Path) -> None:
        """Initialize exception with spec ID and path.

        Args:
            spec_id: The SPEC ID that already has a worktree.
            path: The path of the existing worktree.
        """
        self.spec_id = spec_id
        self.path = path
        super().__init__(f"Worktree for '{spec_id}' already exists at {path}")


class WorktreeNotFoundError(WorktreeError):
    """Raised when trying to access a nonexistent worktree."""

    def __init__(self, spec_id: str) -> None:
        """Initialize exception with spec ID.

        Args:
            spec_id: The SPEC ID that was not found.
        """
        self.spec_id = spec_id
        super().__init__(f"Worktree for '{spec_id}' not found")


class UncommittedChangesError(WorktreeError):
    """Raised when trying to remove a worktree with uncommitted changes."""

    def __init__(self, spec_id: str) -> None:
        """Initialize exception with spec ID.

        Args:
            spec_id: The SPEC ID with uncommitted changes.
        """
        self.spec_id = spec_id
        super().__init__(f"Worktree for '{spec_id}' has uncommitted changes. Use --force to remove anyway.")


class GitOperationError(WorktreeError):
    """Raised when a Git operation fails."""

    def __init__(self, message: str) -> None:
        """Initialize exception with error message.

        Args:
            message: Description of the Git operation error.
        """
        super().__init__(f"Git operation failed: {message}")


class MergeConflictError(WorktreeError):
    """Raised when a merge conflict occurs during sync."""

    def __init__(self, spec_id: str, conflicted_files: list[str]) -> None:
        """Initialize exception with conflict details.

        Args:
            spec_id: The SPEC ID with merge conflicts.
            conflicted_files: List of files with conflicts.
        """
        self.spec_id = spec_id
        self.conflicted_files = conflicted_files
        super().__init__(f"Merge conflict in worktree '{spec_id}'. Conflicted files: {', '.join(conflicted_files)}")


class RegistryInconsistencyError(WorktreeError):
    """Raised when registry state is inconsistent with actual Git worktrees."""

    def __init__(self, message: str) -> None:
        """Initialize exception with error message.

        Args:
            message: Description of the inconsistency.
        """
        super().__init__(f"Registry inconsistency: {message}")
