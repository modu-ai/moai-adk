#!/usr/bin/env python3
"""
MoAI-ADK Unified Checkpoint System (Modularized)

@REQ:CHECKPOINT-SYSTEM-001
@FEATURE:CHECKPOINT-MANAGEMENT-001
@API:GET-CHECKPOINT
@DESIGN:UNIFIED-CHECKPOINT-001
@REFACTOR:MODULAR-ARCHITECTURE-001

Modularized checkpoint system following TRUST principles:
- Single responsibility: Orchestrates modular components
- Clear boundaries: Creation, Management, Recovery modules
- Simplified interface: Maintains backward compatibility
"""

import logging
from pathlib import Path

# Import modular components
from checkpoint_creator import CheckpointCreator, CheckpointInfo, CheckpointCreationError
from checkpoint_manager import CheckpointManager, CheckpointManagerError
from checkpoint_recovery import CheckpointRecovery, CheckpointRecoveryError

logger = logging.getLogger(__name__)


class CheckpointError(Exception):
    """Unified checkpoint system errors"""


class CheckpointSystem:
    """
    Unified checkpoint management system with modular architecture

    Orchestrates three specialized modules:
    - CheckpointCreator: Handles checkpoint creation with Git operations
    - CheckpointManager: Manages metadata, listing, and auto-checkpoint timing
    - CheckpointRecovery: Handles rollback and deletion operations

    Maintains backward compatibility while implementing TRUST principles.
    """

    def __init__(self, project_root: Path | None = None):
        """
        Initialize checkpoint system with modular components

        Args:
            project_root: Project root directory (auto-detected if None)
        """
        self.project_root = project_root

        # Initialize modular components
        self.creator = CheckpointCreator(project_root)
        self.manager = CheckpointManager(project_root)
        self.recovery = CheckpointRecovery(project_root)

    def create_checkpoint(self, message: str, is_auto: bool = False) -> CheckpointInfo:
        """
        Create a new checkpoint

        Args:
            message: Checkpoint message
            is_auto: Whether this is an automatic checkpoint

        Returns:
            CheckpointInfo: Created checkpoint information

        Raises:
            CheckpointError: When checkpoint creation fails
        """
        try:
            # Delegate to creator module
            checkpoint = self.creator.create_checkpoint(message, is_auto)

            # Save metadata through manager module
            self.manager.save_checkpoint_metadata(checkpoint)

            # Clean up old checkpoints through coordinated modules
            self.manager.cleanup_old_checkpoints(
                lambda tag: self.recovery.delete_checkpoint_by_tag_or_index(
                    tag, self.manager.find_checkpoint, self.manager.remove_checkpoint_metadata
                )
            )

            return checkpoint

        except CheckpointCreationError as e:
            raise CheckpointError(str(e))
        except CheckpointManagerError as e:
            raise CheckpointError(str(e))
        except Exception as e:
            logger.error(f"Unexpected error in checkpoint creation: {e}")
            raise CheckpointError(f"Checkpoint creation failed: {e}")

    def list_checkpoints(self, limit: int | None = None) -> list[CheckpointInfo]:
        """
        List checkpoints with optional limit

        Args:
            limit: Maximum number of checkpoints to return

        Returns:
            list[CheckpointInfo]: List of checkpoints sorted by creation time
        """
        return self.manager.list_checkpoints(limit)

    def rollback_to_checkpoint(self, tag_or_index: str) -> CheckpointInfo:
        """
        Rollback to specified checkpoint

        Args:
            tag_or_index: Tag name or numeric index

        Returns:
            CheckpointInfo: The checkpoint that was rolled back to

        Raises:
            CheckpointError: When rollback fails
        """
        try:
            return self.recovery.rollback_by_tag_or_index(
                tag_or_index, self.manager.find_checkpoint
            )
        except CheckpointRecoveryError as e:
            raise CheckpointError(str(e))

    def delete_checkpoint(self, tag_or_index: str) -> bool:
        """
        Delete specified checkpoint

        Args:
            tag_or_index: Tag name or numeric index

        Returns:
            bool: True if deletion was successful
        """
        return self.recovery.delete_checkpoint_by_tag_or_index(
            tag_or_index, self.manager.find_checkpoint, self.manager.remove_checkpoint_metadata
        )

    def should_create_auto_checkpoint(self) -> bool:
        """
        Check if automatic checkpoint should be created

        Returns:
            bool: True if automatic checkpoint creation is needed
        """
        return self.manager.should_create_auto_checkpoint()

    def get_checkpoint_info(self, tag_or_index: str) -> CheckpointInfo | None:
        """
        Get checkpoint information by tag or index

        Args:
            tag_or_index: Tag name or numeric index

        Returns:
            CheckpointInfo | None: Checkpoint info or None if not found
        """
        return self.manager.get_checkpoint_info(tag_or_index)

    # Additional convenience methods for enhanced functionality
    def get_checkpoint_count(self) -> int:
        """Get total number of checkpoints"""
        return self.manager.get_checkpoint_count()

    def get_latest_checkpoint(self) -> CheckpointInfo | None:
        """Get the most recent checkpoint"""
        return self.manager.get_latest_checkpoint()

    def create_stash_before_rollback(self, message: str = "Pre-rollback stash") -> str:
        """Create a Git stash before performing rollback"""
        return self.recovery.create_stash_before_rollback(message)

    def list_stashes(self) -> list[str]:
        """List available Git stashes"""
        return self.recovery.list_stashes()

    def is_clean_working_directory(self) -> bool:
        """Check if working directory is clean"""
        return self.recovery.is_clean_working_directory()


# Backward compatibility convenience functions
def create_checkpoint(
    message: str, project_root: Path | None = None, is_auto: bool = False
) -> CheckpointInfo:
    """
    Convenience function for checkpoint creation

    Args:
        message: Checkpoint message
        project_root: Project root directory (auto-detected if None)
        is_auto: Whether this is an automatic checkpoint

    Returns:
        CheckpointInfo: Created checkpoint information
    """
    system = CheckpointSystem(project_root)
    return system.create_checkpoint(message, is_auto)


def rollback_to_checkpoint(
    tag_or_index: str, project_root: Path | None = None
) -> CheckpointInfo:
    """
    Convenience function for checkpoint rollback

    Args:
        tag_or_index: Tag name or numeric index
        project_root: Project root directory (auto-detected if None)

    Returns:
        CheckpointInfo: The checkpoint that was rolled back to
    """
    system = CheckpointSystem(project_root)
    return system.rollback_to_checkpoint(tag_or_index)


def list_checkpoints(
    limit: int | None = None, project_root: Path | None = None
) -> list[CheckpointInfo]:
    """
    Convenience function for listing checkpoints

    Args:
        limit: Maximum number of checkpoints to return
        project_root: Project root directory (auto-detected if None)

    Returns:
        list[CheckpointInfo]: List of checkpoints sorted by creation time
    """
    system = CheckpointSystem(project_root)
    return system.list_checkpoints(limit)
