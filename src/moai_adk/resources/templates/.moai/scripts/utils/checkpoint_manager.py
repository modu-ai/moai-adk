#!/usr/bin/env python3
"""
MoAI-ADK Checkpoint Manager Module

@REQ:CHECKPOINT-MANAGEMENT-001
@DESIGN:MODULAR-CHECKPOINT-001
@TASK:CHECKPOINT-MANAGER-MODULE

Single responsibility: Manage checkpoint metadata, listing, and automatic checkpoint timing.
"""

import json
import logging
from datetime import UTC, datetime, timedelta, timezone
from pathlib import Path
from typing import Any

from constants import (
    AUTO_CHECKPOINT_INTERVAL_MINUTES,
    BACKUP_RETENTION_DAYS,
    MAX_CHECKPOINTS,
)
from project_helper import ProjectHelper

# Import CheckpointInfo from creator module
from checkpoint_creator import CheckpointInfo

logger = logging.getLogger(__name__)

# Korean timezone (KST) constant - UTC+9
KST = timezone(timedelta(hours=9))


def get_kst_now() -> datetime:
    """
    Return current time in Korean Standard Time (KST)

    Returns:
        datetime: Current time with KST timezone applied
    """
    return datetime.now(KST)


def convert_utc_to_kst(utc_datetime: datetime) -> datetime:
    """
    Convert UTC time to KST time

    Args:
        utc_datetime: UTC timezone datetime object

    Returns:
        datetime: Datetime object converted to KST

    Raises:
        ValueError: When input is not a datetime object
    """
    if not isinstance(utc_datetime, datetime):
        raise ValueError("Input must be a datetime object")

    if utc_datetime.tzinfo is None:
        # Assume UTC if no timezone info
        utc_datetime = utc_datetime.replace(tzinfo=UTC)

    return utc_datetime.astimezone(KST)


class CheckpointManagerError(Exception):
    """Checkpoint manager specific errors"""


class CheckpointManager:
    """
    Manages checkpoint metadata, listing, and automatic checkpoint conditions

    Single responsibility: Handle checkpoint metadata persistence, querying,
    and determining when automatic checkpoints should be created.
    """

    def __init__(self, project_root: Path | None = None):
        self.project_root = project_root or ProjectHelper.find_project_root()
        self.checkpoints_dir = self.project_root / ".moai" / "checkpoints"
        self.metadata_file = self.checkpoints_dir / "metadata.json"
        self.checkpoints_dir.mkdir(parents=True, exist_ok=True)

    def save_checkpoint_metadata(self, checkpoint: CheckpointInfo) -> None:
        """
        Save checkpoint metadata to persistent storage

        Args:
            checkpoint: CheckpointInfo object to save

        Raises:
            CheckpointManagerError: When saving fails
        """
        try:
            metadata = self._load_checkpoint_metadata()

            # Remove existing checkpoint with same tag (update case)
            metadata["checkpoints"] = [
                cp for cp in metadata["checkpoints"]
                if cp.get("tag") != checkpoint.tag
            ]

            # Add new checkpoint
            metadata["checkpoints"].append(checkpoint.to_dict())

            # Write to file
            with open(self.metadata_file, "w", encoding="utf-8") as f:
                json.dump(metadata, f, indent=2, ensure_ascii=False)

            logger.debug(f"Checkpoint metadata saved: {checkpoint.tag}")

        except Exception as e:
            logger.error(f"Failed to save checkpoint metadata: {e}")
            raise CheckpointManagerError(f"Cannot save checkpoint metadata: {e}")

    def list_checkpoints(self, limit: int | None = None) -> list[CheckpointInfo]:
        """
        List all checkpoints, optionally limited

        Args:
            limit: Maximum number of checkpoints to return

        Returns:
            list[CheckpointInfo]: List of checkpoints sorted by creation time (newest first)
        """
        try:
            metadata = self._load_checkpoint_metadata()
            checkpoints = []

            for cp in metadata.get("checkpoints", []):
                try:
                    checkpoint = CheckpointInfo.from_dict(cp)
                    checkpoints.append(checkpoint)
                except Exception as e:
                    logger.warning(f"Failed to parse checkpoint info: {e}, data: {cp}")
                    continue

            # Sort by creation time (newest first)
            checkpoints.sort(key=lambda x: x.created_at, reverse=True)

            if limit:
                checkpoints = checkpoints[:limit]

            return checkpoints

        except Exception as e:
            logger.error(f"Failed to list checkpoints: {e}")
            return []

    def find_checkpoint(self, tag_or_index: str) -> CheckpointInfo | None:
        """
        Find checkpoint by tag name or index

        Args:
            tag_or_index: Tag name or numeric index

        Returns:
            CheckpointInfo | None: Found checkpoint or None if not found
        """
        checkpoints = self.list_checkpoints()

        # Try to find by index if numeric
        if tag_or_index.isdigit():
            index = int(tag_or_index)
            if 0 <= index < len(checkpoints):
                return checkpoints[index]

        # Find by tag name (exact match or suffix match)
        for checkpoint in checkpoints:
            if checkpoint.tag == tag_or_index or checkpoint.tag.endswith(tag_or_index):
                return checkpoint

        return None

    def get_checkpoint_info(self, tag_or_index: str) -> CheckpointInfo | None:
        """
        Get checkpoint information by tag or index

        Args:
            tag_or_index: Tag name or numeric index

        Returns:
            CheckpointInfo | None: Checkpoint info or None if not found
        """
        try:
            return self.find_checkpoint(tag_or_index)
        except Exception as e:
            logger.error(f"Failed to get checkpoint info: {e}")
            return None

    def remove_checkpoint_metadata(self, tag: str) -> None:
        """
        Remove checkpoint from metadata

        Args:
            tag: Tag name to remove

        Raises:
            CheckpointManagerError: When removal fails
        """
        try:
            metadata = self._load_checkpoint_metadata()
            metadata["checkpoints"] = [
                cp for cp in metadata["checkpoints"]
                if cp.get("tag") != tag
            ]

            with open(self.metadata_file, "w", encoding="utf-8") as f:
                json.dump(metadata, f, indent=2, ensure_ascii=False)

            logger.debug(f"Checkpoint metadata removed: {tag}")

        except Exception as e:
            logger.error(f"Failed to remove checkpoint metadata: {e}")
            raise CheckpointManagerError(f"Cannot remove checkpoint metadata: {e}")

    def should_create_auto_checkpoint(self) -> bool:
        """Check if automatic checkpoint should be created based on KST timezone"""
        try:
            checkpoints = self.list_checkpoints(limit=1)
            if not checkpoints:
                return True

            last_time = datetime.fromisoformat(checkpoints[0].created_at)
            now = get_kst_now()

            # Normalize to KST for comparison
            if last_time.tzinfo is None:
                last_time = last_time.replace(tzinfo=UTC).astimezone(KST)
            elif last_time.tzinfo != KST:
                last_time = last_time.astimezone(KST)

            time_diff_minutes = (now - last_time).total_seconds() / 60
            return time_diff_minutes >= AUTO_CHECKPOINT_INTERVAL_MINUTES

        except Exception as e:
            logger.error(f"Failed to check auto checkpoint condition: {e}")
            return False  # Safety: prevent infinite checkpoint creation

    def cleanup_old_checkpoints(self, delete_checkpoint_func) -> None:
        """Clean up old checkpoints based on retention policies"""
        try:
            checkpoints = self.list_checkpoints()

            # Remove checkpoints exceeding maximum count
            if len(checkpoints) > MAX_CHECKPOINTS:
                for checkpoint in checkpoints[MAX_CHECKPOINTS:]:
                    delete_checkpoint_func(checkpoint.tag)

            # Remove old automatic checkpoints based on retention period
            cutoff_date = get_kst_now() - timedelta(days=BACKUP_RETENTION_DAYS)
            for checkpoint in checkpoints:
                created_date = datetime.fromisoformat(checkpoint.created_at)
                if created_date.tzinfo is None:
                    created_date = created_date.replace(tzinfo=UTC).astimezone(KST)
                elif created_date.tzinfo != KST:
                    created_date = created_date.astimezone(KST)

                if created_date < cutoff_date and checkpoint.is_auto:
                    delete_checkpoint_func(checkpoint.tag)

        except Exception as e:
            logger.error(f"Failed to cleanup old checkpoints: {e}")

    def _load_checkpoint_metadata(self) -> dict[str, Any]:
        """
        Load checkpoint metadata from file

        Returns:
            dict: Metadata dictionary with checkpoints list and version info
        """
        if not self.metadata_file.exists():
            return {"checkpoints": [], "version": "1.0"}

        try:
            with open(self.metadata_file, encoding="utf-8") as f:
                return json.load(f)
        except (json.JSONDecodeError, Exception) as e:
            logger.error(f"Failed to load metadata: {e}")
            return {"checkpoints": [], "version": "1.0"}

    def get_checkpoint_count(self) -> int:
        """Get total number of checkpoints"""
        return len(self.list_checkpoints())

    def get_latest_checkpoint(self) -> CheckpointInfo | None:
        """Get the most recent checkpoint"""
        checkpoints = self.list_checkpoints(limit=1)
        return checkpoints[0] if checkpoints else None