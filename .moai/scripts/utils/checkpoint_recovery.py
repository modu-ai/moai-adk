#!/usr/bin/env python3
"""
MoAI-ADK Checkpoint Recovery Module

@REQ:CHECKPOINT-RECOVERY-001
@DESIGN:MODULAR-CHECKPOINT-001
@TASK:CHECKPOINT-RECOVERY-MODULE

Single responsibility: Handle checkpoint rollback operations and checkpoint deletion.
"""

import logging
from pathlib import Path

from git_helper import GitCommandError, GitHelper
from project_helper import ProjectHelper

# Import CheckpointInfo from creator module
from checkpoint_creator import CheckpointInfo

logger = logging.getLogger(__name__)


class CheckpointRecoveryError(Exception):
    """Checkpoint recovery specific errors"""


class CheckpointRecovery:
    """
    Handles checkpoint rollback and deletion operations

    Single responsibility: Perform Git operations for rollback and
    manage checkpoint deletion including cleanup of associated files.
    """

    def __init__(self, project_root: Path | None = None):
        self.project_root = project_root or ProjectHelper.find_project_root()
        self.git = GitHelper(self.project_root)
        self.checkpoints_dir = self.project_root / ".moai" / "checkpoints"

    def rollback_to_checkpoint(
        self, checkpoint: CheckpointInfo, find_checkpoint_func
    ) -> CheckpointInfo:
        """
        Rollback repository state to specified checkpoint

        Args:
            checkpoint: CheckpointInfo object to rollback to
            find_checkpoint_func: Function to find checkpoint (from manager module)

        Returns:
            CheckpointInfo: The checkpoint that was rolled back to

        Raises:
            CheckpointRecoveryError: When rollback fails
        """
        try:
            # Guard clause: check for uncommitted changes and stash if needed
            if self.git.has_uncommitted_changes():
                stash_id = self.git.stash_push("Pre-rollback stash")
                logger.info(f"Uncommitted changes stashed: {stash_id}")

            # Verify checkpoint exists
            if not checkpoint:
                raise CheckpointRecoveryError("Checkpoint not found")

            # Perform hard reset to checkpoint commit
            self._perform_hard_reset(checkpoint.commit_hash)

            logger.info(f"Rollback completed successfully: {checkpoint.tag}")
            return checkpoint

        except GitCommandError as e:
            logger.error(f"Git rollback operation failed: {e}")
            raise CheckpointRecoveryError(f"Rollback failed: {e}")
        except Exception as e:
            logger.error(f"Unexpected error during rollback: {e}")
            raise CheckpointRecoveryError(f"Rollback error: {e}")

    def rollback_by_tag_or_index(
        self, tag_or_index: str, find_checkpoint_func
    ) -> CheckpointInfo:
        """
        Rollback to checkpoint by tag name or index

        Args:
            tag_or_index: Tag name or numeric index
            find_checkpoint_func: Function to find checkpoint (from manager module)

        Returns:
            CheckpointInfo: The checkpoint that was rolled back to

        Raises:
            CheckpointRecoveryError: When checkpoint not found or rollback fails
        """
        checkpoint = find_checkpoint_func(tag_or_index)
        if not checkpoint:
            raise CheckpointRecoveryError(f"Checkpoint not found: {tag_or_index}")

        return self.rollback_to_checkpoint(checkpoint, find_checkpoint_func)

    def delete_checkpoint(
        self, checkpoint: CheckpointInfo, remove_metadata_func
    ) -> bool:
        """
        Delete checkpoint and clean up associated resources

        Args:
            checkpoint: CheckpointInfo object to delete
            remove_metadata_func: Function to remove metadata (from manager module)

        Returns:
            bool: True if deletion was successful

        Note:
            This method coordinates with the manager module for metadata removal.
        """
        try:
            # Remove Git tag (best effort - tag might not exist)
            self._delete_git_tag(checkpoint.tag)

            # Remove backup file if it exists
            self._delete_backup_file(checkpoint.tag)

            # Remove from metadata (delegated to manager module)
            remove_metadata_func(checkpoint.tag)

            logger.info(f"Checkpoint deleted successfully: {checkpoint.tag}")
            return True

        except Exception as e:
            logger.error(f"Failed to delete checkpoint {checkpoint.tag}: {e}")
            return False

    def delete_checkpoint_by_tag_or_index(
        self, tag_or_index: str, find_checkpoint_func, remove_metadata_func
    ) -> bool:
        """
        Delete checkpoint by tag name or index

        Args:
            tag_or_index: Tag name or numeric index
            find_checkpoint_func: Function to find checkpoint (from manager module)
            remove_metadata_func: Function to remove metadata (from manager module)

        Returns:
            bool: True if deletion was successful
        """
        checkpoint = find_checkpoint_func(tag_or_index)
        if not checkpoint:
            logger.warning(f"Checkpoint not found for deletion: {tag_or_index}")
            return False

        return self.delete_checkpoint(checkpoint, remove_metadata_func)

    def create_stash_before_rollback(self, message: str = "Pre-rollback stash") -> str:
        """Create Git stash before rollback"""
        try:
            if not self.git.has_uncommitted_changes():
                return ""
            stash_id = self.git.stash_push(message)
            logger.info(f"Created stash: {stash_id}")
            return stash_id
        except GitCommandError as e:
            raise CheckpointRecoveryError(f"Stash creation failed: {e}")

    def list_stashes(self) -> list[str]:
        """List available Git stashes"""
        try:
            result = self.git.run_command(["git", "stash", "list"])
            return result.stdout.strip().split("\n") if result.stdout.strip() else []
        except GitCommandError:
            return []

    def is_clean_working_directory(self) -> bool:
        """Check if working directory is clean"""
        try:
            return not self.git.has_uncommitted_changes()
        except Exception:
            return False

    def _perform_hard_reset(self, commit_hash: str) -> None:
        """
        Perform Git hard reset to specified commit

        Args:
            commit_hash: Target commit hash

        Raises:
            CheckpointRecoveryError: When hard reset fails
        """
        try:
            self.git.run_command(["git", "reset", "--hard", commit_hash])
            logger.debug(f"Hard reset performed to: {commit_hash}")
        except GitCommandError as e:
            logger.error(f"Hard reset failed: {e}")
            raise CheckpointRecoveryError(f"Hard reset failed: {e}")

    def _delete_git_tag(self, tag: str) -> None:
        """
        Delete Git tag (best effort)

        Args:
            tag: Tag name to delete

        Note:
            This method does not raise exceptions as tag might not exist.
        """
        try:
            self.git.run_command(["git", "tag", "-d", tag])
            logger.debug(f"Git tag deleted: {tag}")
        except GitCommandError:
            logger.warning(f"Failed to delete Git tag (might not exist): {tag}")

    def _delete_backup_file(self, tag: str) -> None:
        """
        Delete backup file associated with checkpoint

        Args:
            tag: Tag name to find associated backup file

        Note:
            This method does not raise exceptions as backup file might not exist.
        """
        backup_path = self.checkpoints_dir / f"{tag}.tar.gz"
        try:
            if backup_path.exists():
                backup_path.unlink()
                logger.debug(f"Backup file deleted: {backup_path}")
        except Exception as e:
            logger.warning(f"Failed to delete backup file {backup_path}: {e}")

