#!/usr/bin/env python3
"""
MoAI-ADK Checkpoint Creator Module

@REQ:CHECKPOINT-CREATION-001
@DESIGN:MODULAR-CHECKPOINT-001
@TASK:CHECKPOINT-CREATOR-MODULE

Single responsibility: Create new checkpoints with proper validation and Git operations.
"""

import logging
import re
from datetime import datetime
from pathlib import Path
from typing import Any

from constants import (
    CHECKPOINT_MESSAGE_MAX_LENGTH,
    CHECKPOINT_TAG_PREFIX,
    ERROR_MESSAGES,
)
from git_helper import GitCommandError, GitHelper
from project_helper import ProjectHelper

logger = logging.getLogger(__name__)

# Korean timezone (KST) constant - UTC+9
from datetime import timezone, timedelta
KST = timezone(timedelta(hours=9))


def get_kst_now() -> datetime:
    """
    Return current time in Korean Standard Time (KST)

    Returns:
        datetime: Current time with KST timezone applied

    Note:
        Always returns timezone-aware datetime for compatibility
        with existing UTC-based checkpoints.
    """
    return datetime.now(KST)


class CheckpointCreationError(Exception):
    """Checkpoint creation specific errors"""


class CheckpointInfo:
    """Checkpoint information data structure"""

    def __init__(self, tag: str, commit_hash: str, message: str, created_at: str,
                 file_count: int = 0, is_auto: bool = False):
        self.tag = tag
        self.commit_hash = commit_hash
        self.message = message
        self.created_at = created_at
        self.file_count = file_count
        self.is_auto = is_auto

    def to_dict(self) -> dict[str, Any]:
        return {"tag": self.tag, "commit_hash": self.commit_hash, "message": self.message,
                "created_at": self.created_at, "file_count": self.file_count, "is_auto": self.is_auto}

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "CheckpointInfo":
        if "tag" in data:  # New format
            return cls(tag=data["tag"], commit_hash=data["commit_hash"], message=data["message"],
                      created_at=data["created_at"], file_count=data.get("file_count", 0),
                      is_auto=data.get("is_auto", False))
        elif "id" in data:  # Legacy format
            return cls(tag=f"moai_cp/{data['id'].replace('checkpoint_', '')}", commit_hash=data["commit"],
                      message=data.get("message", "Legacy checkpoint"), created_at=data["timestamp"],
                      file_count=data.get("files_changed", 0), is_auto=data.get("kind") == "auto")
        else:
            raise ValueError(f"Invalid checkpoint data format: {data}")


class CheckpointCreator:
    """
    Handles checkpoint creation with Git operations

    Single responsibility: Create new checkpoints with proper validation,
    message sanitization, and Git tag/commit operations.
    """

    def __init__(self, project_root: Path | None = None):
        self.project_root = project_root or ProjectHelper.find_project_root()
        self.git = GitHelper(self.project_root)

    def create_checkpoint(self, message: str, is_auto: bool = False) -> CheckpointInfo:
        """
        Create a new checkpoint with proper validation and Git operations

        Args:
            message: Checkpoint message
            is_auto: Whether this is an automatic checkpoint

        Returns:
            CheckpointInfo: Created checkpoint information

        Raises:
            CheckpointCreationError: When checkpoint creation fails
        """
        # Input validation with guard clauses
        if not isinstance(message, str):
            raise CheckpointCreationError("Message must be a string")

        if not message.strip():
            raise CheckpointCreationError("Message cannot be empty")

        # Guard clause for Git repository check
        if not self.git.is_git_repo():
            raise CheckpointCreationError(ERROR_MESSAGES["not_git_repo"])

        try:
            # Sanitize message for Git compatibility
            sanitized_message = self._sanitize_checkpoint_message(message)

            # Check for uncommitted changes
            has_changes = self.git.has_uncommitted_changes()
            if not has_changes:
                logger.info("No changes detected, creating tag on current HEAD.")

            # Stage changes if they exist
            if has_changes:
                self._stage_changes_safely()

            # Generate KST-based timestamp and tag name
            kst_now = get_kst_now()
            tag_name = self._generate_unique_tag_name(kst_now)

            # Perform Git operations (commit and tag)
            commit_message = (
                f"📍 {'Auto-' if is_auto else ''}Checkpoint: {sanitized_message}"
            )
            commit_hash = self._create_commit_and_tag(
                has_changes, commit_message, tag_name
            )

            # Create checkpoint information
            checkpoint = CheckpointInfo(
                tag=tag_name,
                commit_hash=commit_hash,
                message=sanitized_message,
                created_at=kst_now.isoformat(),
                file_count=self._count_tracked_files(),
                is_auto=is_auto,
            )

            logger.info(
                f"Checkpoint created successfully: {tag_name} "
                f"(KST: {kst_now.strftime('%Y-%m-%d %H:%M:%S')})"
            )
            return checkpoint

        except GitCommandError as e:
            logger.error(f"Git command execution failed: {e}")
            raise CheckpointCreationError(f"Git command execution failed: {e}")
        except CheckpointCreationError:
            # Re-raise CheckpointCreationError as-is
            raise
        except Exception as e:
            logger.error(f"Unexpected error during checkpoint creation: {e}")
            raise CheckpointCreationError(f"Checkpoint creation failed: {e}")

    def _sanitize_checkpoint_message(self, message: str) -> str:
        """
        Sanitize and validate checkpoint message for Git compatibility

        Args:
            message: Raw checkpoint message

        Returns:
            str: Sanitized message safe for Git operations
        """
        # Remove leading/trailing whitespace
        message = message.strip()

        # Apply length limit
        if len(message) > CHECKPOINT_MESSAGE_MAX_LENGTH:
            message = message[: CHECKPOINT_MESSAGE_MAX_LENGTH - 3] + "..."

        # Remove newline characters (problematic for Git tag messages)
        message = message.replace("\n", " ").replace("\r", " ")

        # Clean up consecutive whitespace
        message = re.sub(r"\s+", " ", message)

        return message

    def _stage_changes_safely(self) -> None:
        """
        Safely stage all changes with proper error handling

        Raises:
            CheckpointCreationError: When staging fails
        """
        try:
            self.git.stage_all_changes()
        except GitCommandError as e:
            logger.error(f"Failed to stage changes: {e}")
            raise CheckpointCreationError(f"Cannot stage changes: {e}")

    def _generate_unique_tag_name(self, kst_now: datetime) -> str:
        """
        Generate unique tag name with KST timestamp

        Args:
            kst_now: Current KST datetime

        Returns:
            str: Unique tag name
        """
        # Generate base timestamp
        timestamp = kst_now.strftime("%Y%m%d_%H%M%S")
        tag_name = f"{CHECKPOINT_TAG_PREFIX}{timestamp}"

        # Check for duplicates and add microseconds if needed
        existing_tags = self._get_existing_tags()
        if tag_name in existing_tags:
            # Add microseconds to avoid collision
            timestamp = kst_now.strftime("%Y%m%d_%H%M%S_%f")[:17]  # 3-digit microseconds
            tag_name = f"{CHECKPOINT_TAG_PREFIX}{timestamp}"

        return tag_name

    def _get_existing_tags(self) -> set:
        """
        Get existing checkpoint tags from Git repository

        Returns:
            set: Set of existing tag names
        """
        try:
            result = self.git.run_command(
                ["git", "tag", "-l", f"{CHECKPOINT_TAG_PREFIX}*"]
            )
            return (
                set(result.stdout.strip().split("\n"))
                if result.stdout.strip()
                else set()
            )
        except GitCommandError:
            logger.warning("Failed to retrieve existing tags, returning empty set")
            return set()

    def _create_commit_and_tag(
        self, has_changes: bool, commit_message: str, tag_name: str
    ) -> str:
        """
        Create Git commit and tag

        Args:
            has_changes: Whether there are uncommitted changes
            commit_message: Commit message
            tag_name: Tag name to create

        Returns:
            str: Commit hash

        Raises:
            CheckpointCreationError: When Git operations fail
        """
        try:
            if has_changes:
                commit_hash = self.git.commit(commit_message, allow_empty=False)
            else:
                commit_hash = self.git.run_command(
                    ["git", "rev-parse", "HEAD"]
                ).stdout.strip()

            self.git.create_tag(tag_name, commit_message)
            return commit_hash

        except GitCommandError as e:
            logger.error(f"Failed to create commit or tag: {e}")
            raise CheckpointCreationError(f"Git operation failed: {e}")

    def _count_tracked_files(self) -> int:
        """
        Count number of tracked files in the repository

        Returns:
            int: Number of tracked files
        """
        try:
            result = self.git.run_command(["git", "ls-files"])
            return len(result.stdout.splitlines())
        except GitCommandError:
            return 0