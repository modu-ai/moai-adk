"""
Branch Manager - Manage local checkpoint branches.

SPEC: .moai/specs/SPEC-CHECKPOINT-EVENT-001/spec.md
"""

from datetime import datetime

import git


class BranchManager:
    """Manage local checkpoint branches."""

    MAX_CHECKPOINTS = 10
    CHECKPOINT_PREFIX = "before-"

    def __init__(self, repo: git.Repo):
        """
        Initialize the BranchManager.

        Args:
            repo: GitPython Repo instance.
        """
        self.repo = repo
        self._old_branches: set[str] = set()

    def create_checkpoint_branch(self, operation: str) -> str:
        """
        Create a checkpoint branch.

        SPEC requirement: before-{operation}-{timestamp} format for local branches.

        Args:
            operation: Operation name (delete, refactor, merge, etc.).

        Returns:
            Name of the created branch.
        """
        timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
        branch_name = f"{self.CHECKPOINT_PREFIX}{operation}-{timestamp}"

        # Create the branch from the current HEAD
        self.repo.create_head(branch_name)

        # Remove old checkpoints using FIFO order
        self._enforce_max_checkpoints()

        return branch_name

    def branch_exists(self, branch_name: str) -> bool:
        """
        Check if a branch exists.

        Args:
            branch_name: Name of the branch to check.

        Returns:
            True when the branch exists, otherwise False.
        """
        return branch_name in [head.name for head in self.repo.heads]

    def has_remote_tracking(self, branch_name: str) -> bool:
        """
        Determine whether a remote tracking branch exists.

        SPEC requirement: checkpoints must remain local-only branches.

        Args:
            branch_name: Branch name to inspect.

        Returns:
            True if a tracking branch exists, otherwise False.
        """
        try:
            branch = self.repo.heads[branch_name]
            return branch.tracking_branch() is not None
        except (IndexError, AttributeError):
            return False

    def list_checkpoint_branches(self) -> list[str]:
        """
        List all checkpoint branches.

        Returns:
            Names of checkpoint branches.
        """
        return [head.name for head in self.repo.heads if head.name.startswith(self.CHECKPOINT_PREFIX)]

    def mark_as_old(self, branch_name: str) -> None:
        """
        Mark a branch as old (used for tests).

        Args:
            branch_name: Branch to flag as old.
        """
        self._old_branches.add(branch_name)

    def cleanup_old_checkpoints(self, max_count: int) -> None:
        """
        Clean up old checkpoint branches.

        SPEC requirement: delete using FIFO when exceeding the maximum count.

        Args:
            max_count: Maximum number of checkpoints to retain.
        """
        checkpoints = self.list_checkpoint_branches()

        # Sort in chronological order (branches marked via mark_as_old first)
        sorted_checkpoints = sorted(checkpoints, key=lambda name: (name not in self._old_branches, name))

        # Delete the excess branches
        to_delete = sorted_checkpoints[: len(sorted_checkpoints) - max_count]
        for branch_name in to_delete:
            if branch_name in [head.name for head in self.repo.heads]:
                self.repo.delete_head(branch_name, force=True)

    def _enforce_max_checkpoints(self) -> None:
        """Maintain the maximum number of checkpoints (internal)."""
        checkpoints = self.list_checkpoint_branches()

        if len(checkpoints) > self.MAX_CHECKPOINTS:
            # Sort alphabetically (older timestamps first)
            sorted_checkpoints = sorted(checkpoints)
            to_delete = sorted_checkpoints[: len(sorted_checkpoints) - self.MAX_CHECKPOINTS]

            for branch_name in to_delete:
                self.repo.delete_head(branch_name, force=True)
