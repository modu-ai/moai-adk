"""
Checkpoint Manager - Event-driven checkpoint system.

SPEC: .moai/specs/SPEC-CHECKPOINT-EVENT-001/spec.md
"""

from datetime import datetime
from pathlib import Path
from typing import Optional

import git

from moai_adk.core.git.branch_manager import BranchManager
from moai_adk.core.git.event_detector import EventDetector


class CheckpointManager:
    """Manage creation and restoration of event-driven checkpoints."""

    def __init__(self, repo: git.Repo, project_root: Path):
        """
        Initialize the CheckpointManager.

        Args:
            repo: GitPython Repo instance.
            project_root: Project root directory.
        """
        self.repo = repo
        self.project_root = project_root
        self.event_detector = EventDetector()
        self.branch_manager = BranchManager(repo)
        self.log_file = project_root / ".moai" / "checkpoints.log"

        # Ensure the log directory exists
        self.log_file.parent.mkdir(parents=True, exist_ok=True)

    def create_checkpoint_if_risky(
        self,
        operation: str,
        deleted_files: Optional[list[str]] = None,
        renamed_files: Optional[list[tuple[str, str]]] = None,
        modified_files: Optional[list[Path]] = None,
    ) -> Optional[str]:
        """
        Create a checkpoint when a risky operation is detected.

        SPEC requirement: automatically create a checkpoint for risky actions.

        Args:
            operation: Operation type.
            deleted_files: Files scheduled for deletion.
            renamed_files: Files that will be renamed.
            modified_files: Files that will be modified.

        Returns:
            Created checkpoint ID (branch name) or None when the operation is safe.
        """
        is_risky = False

        # Identify large deletion operations
        if deleted_files and self.event_detector.is_risky_deletion(deleted_files):
            is_risky = True

        # Identify large-scale refactoring
        if renamed_files and self.event_detector.is_risky_refactoring(renamed_files):
            is_risky = True

        # Check for critical file modifications
        if modified_files:
            for file_path in modified_files:
                if self.event_detector.is_critical_file(file_path):
                    is_risky = True
                    break

        if not is_risky:
            return None

        # Create a checkpoint
        checkpoint_id = self.branch_manager.create_checkpoint_branch(operation)

        # Record checkpoint metadata
        self._log_checkpoint(checkpoint_id, operation)

        return checkpoint_id

    def restore_checkpoint(self, checkpoint_id: str) -> None:
        """
        Restore the repository to the specified checkpoint.

        SPEC requirement: capture the current state as a new checkpoint before restoring.

        Args:
            checkpoint_id: Target checkpoint ID (branch name).
        """
        # Save current state as a safety checkpoint before restoring
        safety_checkpoint = self.branch_manager.create_checkpoint_branch("restore")
        self._log_checkpoint(safety_checkpoint, "restore", is_safety=True)

        # Check out the checkpoint branch
        self.repo.git.checkout(checkpoint_id)

    def list_checkpoints(self) -> list[str]:
        """
        List all checkpoints.

        Returns:
            List of checkpoint IDs.
        """
        return self.branch_manager.list_checkpoint_branches()

    def _log_checkpoint(self, checkpoint_id: str, operation: str, is_safety: bool = False) -> None:
        """
        Append checkpoint metadata to the log file.

        SPEC requirement: write metadata to .moai/checkpoints.log.

        Args:
            checkpoint_id: Checkpoint identifier.
            operation: Operation type.
            is_safety: Whether the checkpoint was created for safety.
        """
        timestamp = datetime.now().isoformat()

        log_entry = f"""---
checkpoint_id: {checkpoint_id}
operation: {operation}
timestamp: {timestamp}
is_safety: {is_safety}
---
"""

        # Append the entry to the log
        with open(self.log_file, "a", encoding="utf-8") as f:
            f.write(log_entry)
