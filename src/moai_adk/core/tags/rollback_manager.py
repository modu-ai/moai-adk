#!/usr/bin/env python3
# @CODE:TAG-ROLLBACK-MANAGER-001 | @SPEC:TAG-ROLLBACK-001
"""TAG system rollback manager

Manager providing safe rollback when problems occur in the TAG policy system.
Supports checkpoint-based recovery and history tracking.

Key Features:
- Checkpoint creation and management
- Safe rollback execution
- History tracking and logging
- Emergency recovery system
"""

import json
import shutil
import time
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional


@dataclass
class Checkpoint:
    """Checkpoint information

    Attributes:
        id: Unique checkpoint ID
        timestamp: Creation time
        description: Checkpoint description
        file_states: File state information
        metadata: Additional metadata
    """
    id: str
    timestamp: datetime
    description: str
    file_states: Dict[str, str]  # {file_path: content_hash}
    metadata: Dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary"""
        return {
            "id": self.id,
            "timestamp": self.timestamp.isoformat(),
            "description": self.description,
            "file_states": self.file_states,
            "metadata": self.metadata
        }


@dataclass
class RollbackConfig:
    """Rollback system configuration

    Attributes:
        checkpoints_dir: Checkpoint storage directory
        max_checkpoints: Maximum number of checkpoints
        auto_cleanup: Enable automatic cleanup
        backup_before_rollback: Create backup before rollback
        rollback_timeout: Rollback timeout (seconds)
    """
    checkpoints_dir: str = ".moai/checkpoints"
    max_checkpoints: int = 10
    auto_cleanup: bool = True
    backup_before_rollback: bool = True
    rollback_timeout: int = 30


class RollbackManager:
    """TAG system rollback manager

    Provides checkpoint-based rollback system to ensure TAG policy system stability.
    Supports rapid and safe recovery when problems occur.

    Usage:
        config = RollbackConfig()
        manager = RollbackManager(config=config)

        # Create checkpoint
        checkpoint_id = manager.create_checkpoint("State before work")

        # Execute rollback
        success = manager.rollback_to_checkpoint(checkpoint_id)

        # Rollback to latest checkpoint
        success = manager.rollback_to_latest()
    """

    def __init__(self, config: Optional[RollbackConfig] = None):
        """Initialize

        Args:
            config: Rollback configuration (default: RollbackConfig())
        """
        self.config = config or RollbackConfig()
        self.checkpoints_dir = Path(self.config.checkpoints_dir)
        self.checkpoints_dir.mkdir(parents=True, exist_ok=True)

    def create_checkpoint(self, description: str,
                         files: Optional[List[str]] = None,
                         metadata: Optional[Dict[str, Any]] = None) -> str:
        """Create checkpoint

        Args:
            description: Checkpoint description
            files: List of files to include (auto-detect if None)
            metadata: Additional metadata

        Returns:
            Created checkpoint ID
        """
        checkpoint_id = self._generate_checkpoint_id()
        timestamp = datetime.now()

        # Collect file states
        if files is None:
            files = self._discover_project_files()

        file_states = self._collect_file_states(files)

        # Create checkpoint
        checkpoint = Checkpoint(
            id=checkpoint_id,
            timestamp=timestamp,
            description=description,
            file_states=file_states,
            metadata=metadata or {}
        )

        # Save checkpoint
        self._save_checkpoint(checkpoint)

        # Create file backup
        self._backup_files(checkpoint_id, files)

        # Clean up old checkpoints
        if self.config.auto_cleanup:
            self._cleanup_old_checkpoints()

        return checkpoint_id

    def rollback_to_checkpoint(self, checkpoint_id: str) -> bool:
        """Rollback to specific checkpoint

        Args:
            checkpoint_id: Checkpoint ID to rollback to

        Returns:
            Success status
        """
        try:
            checkpoint = self._load_checkpoint(checkpoint_id)
            if not checkpoint:
                return False

            # Create backup before rollback
            if self.config.backup_before_rollback:
                self._create_rollback_backup(checkpoint_id)

            # Restore files
            success = self._restore_files(checkpoint)

            if success:
                # Log rollback
                self._log_rollback(checkpoint_id, checkpoint)

            return success

        except Exception:
            return False

    def rollback_to_latest(self) -> bool:
        """Rollback to latest checkpoint

        Returns:
            Success status
        """
        latest_id = self.get_latest_checkpoint_id()
        if not latest_id:
            return False

        return self.rollback_to_checkpoint(latest_id)

    def list_checkpoints(self) -> List[Dict[str, Any]]:
        """List checkpoints

        Returns:
            List of checkpoint information
        """
        checkpoints = []

        for checkpoint_file in self.checkpoints_dir.glob("checkpoint_*.json"):
            try:
                with open(checkpoint_file, 'r', encoding='utf-8') as f:
                    checkpoint_data = json.load(f)
                    checkpoints.append(checkpoint_data)
            except Exception:
                continue

        # Sort by time (newest first)
        checkpoints.sort(key=lambda x: x.get('timestamp', ''), reverse=True)

        return checkpoints

    def get_latest_checkpoint_id(self) -> Optional[str]:
        """Get latest checkpoint ID

        Returns:
            Latest checkpoint ID or None
        """
        checkpoints = self.list_checkpoints()
        return checkpoints[0]['id'] if checkpoints else None

    def delete_checkpoint(self, checkpoint_id: str) -> bool:
        """Delete checkpoint

        Args:
            checkpoint_id: Checkpoint ID to delete

        Returns:
            Success status
        """
        try:
            # Delete checkpoint file
            checkpoint_file = self.checkpoints_dir / f"checkpoint_{checkpoint_id}.json"
            if checkpoint_file.exists():
                checkpoint_file.unlink()

            # Delete backup files
            backup_dir = self.checkpoints_dir / f"backup_{checkpoint_id}"
            if backup_dir.exists():
                shutil.rmtree(backup_dir)

            return True

        except Exception:
            return False

    def emergency_rollback(self) -> bool:
        """Emergency rollback

        Cancel all changes and return to the safest state.

        Returns:
            Success status
        """
        try:
            # Find oldest stable checkpoint
            stable_checkpoints = self._find_stable_checkpoints()
            if not stable_checkpoints:
                return False

            # Rollback to oldest stable checkpoint
            oldest_stable = stable_checkpoints[-1]  # Sorted by oldest
            return self.rollback_to_checkpoint(oldest_stable['id'])

        except Exception:
            return False

    def validate_checkpoint_integrity(self, checkpoint_id: str) -> bool:
        """Validate checkpoint integrity

        Args:
            checkpoint_id: Checkpoint ID to validate

        Returns:
            Integrity status
        """
        try:
            checkpoint = self._load_checkpoint(checkpoint_id)
            if not checkpoint:
                return False

            # Validate file integrity
            backup_dir = self.checkpoints_dir / f"backup_{checkpoint_id}"
            if not backup_dir.exists():
                return False

            # Check backup file existence
            for file_path in checkpoint.file_states.keys():
                backup_file = backup_dir / file_path.replace('/', '_')
                if not backup_file.exists():
                    return False

            return True

        except Exception:
            return False

    def _generate_checkpoint_id(self) -> str:
        """Generate checkpoint ID

        Returns:
            Unique checkpoint ID
        """
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        random_suffix = str(int(time.time() * 1000))[-6:]
        return f"ckpt_{timestamp}_{random_suffix}"

    def _discover_project_files(self) -> List[str]:
        """Auto-discover project files

        Returns:
            List of project file paths
        """
        files = []
        important_patterns = [
            "src/**/*.py",
            "tests/**/*.py",
            "**/*.md",
            "**/*.json",
            "**/*.yml",
            "**/*.yaml",
            ".claude/**/*",
            ".moai/**/*"
        ]

        for pattern in important_patterns:
            for path in Path(".").glob(pattern):
                if path.is_file():
                    files.append(str(path))

        return list(set(files))  # Remove duplicates

    def _collect_file_states(self, files: List[str]) -> Dict[str, str]:
        """Collect file states

        Args:
            files: List of file paths

        Returns:
            {file_path: content_hash} dictionary
        """
        file_states = {}

        for file_path in files:
            try:
                path = Path(file_path)
                if path.exists():
                    content = path.read_text(encoding="utf-8", errors="ignore")
                    # Simple hash (recommend using hashlib in production)
                    content_hash = str(hash(content))
                    file_states[file_path] = content_hash
            except Exception:
                continue

        return file_states

    def _save_checkpoint(self, checkpoint: Checkpoint) -> None:
        """Save checkpoint

        Args:
            checkpoint: Checkpoint to save
        """
        checkpoint_file = self.checkpoints_dir / f"checkpoint_{checkpoint.id}.json"
        with open(checkpoint_file, 'w', encoding='utf-8') as f:
            json.dump(checkpoint.to_dict(), f, indent=2, ensure_ascii=False)

    def _backup_files(self, checkpoint_id: str, files: List[str]) -> None:
        """Create file backup

        Args:
            checkpoint_id: Checkpoint ID
            files: List of files to backup
        """
        backup_dir = self.checkpoints_dir / f"backup_{checkpoint_id}"
        backup_dir.mkdir(exist_ok=True)

        for file_path in files:
            try:
                path = Path(file_path)
                if path.exists():
                    # Convert / to _ in filename to create valid filename
                    backup_name = file_path.replace('/', '_')
                    backup_file = backup_dir / backup_name
                    shutil.copy2(path, backup_file)
            except Exception:
                continue

    def _load_checkpoint(self, checkpoint_id: str) -> Optional[Checkpoint]:
        """Load checkpoint

        Args:
            checkpoint_id: Checkpoint ID to load

        Returns:
            Checkpoint object or None
        """
        try:
            checkpoint_file = self.checkpoints_dir / f"checkpoint_{checkpoint_id}.json"
            if not checkpoint_file.exists():
                return None

            with open(checkpoint_file, 'r', encoding='utf-8') as f:
                data = json.load(f)

            return Checkpoint(
                id=data['id'],
                timestamp=datetime.fromisoformat(data['timestamp']),
                description=data['description'],
                file_states=data['file_states'],
                metadata=data.get('metadata', {})
            )

        except Exception:
            return None

    def _restore_files(self, checkpoint: Checkpoint) -> bool:
        """Restore files

        Args:
            checkpoint: Checkpoint to restore

        Returns:
            Success status
        """
        backup_dir = self.checkpoints_dir / f"backup_{checkpoint.id}"
        if not backup_dir.exists():
            return False

        success_count = 0
        total_files = len(checkpoint.file_states)

        for file_path in checkpoint.file_states.keys():
            try:
                backup_name = file_path.replace('/', '_')
                backup_file = backup_dir / backup_name
                target_file = Path(file_path)

                if backup_file.exists():
                    # Create target directory
                    target_file.parent.mkdir(parents=True, exist_ok=True)
                    # Restore file
                    shutil.copy2(backup_file, target_file)
                    success_count += 1

            except Exception:
                continue

        return success_count == total_files

    def _create_rollback_backup(self, checkpoint_id: str) -> None:
        """Create backup before rollback

        Args:
            checkpoint_id: Checkpoint ID to rollback to
        """
        self.create_checkpoint(
            description=f"Backup before rollback (from {checkpoint_id})",
            metadata={"rollback_from": checkpoint_id}
        )

    def _cleanup_old_checkpoints(self) -> None:
        """Clean up old checkpoints"""
        checkpoints = self.list_checkpoints()

        if len(checkpoints) > self.config.max_checkpoints:
            # Delete oldest checkpoints
            old_checkpoints = checkpoints[self.config.max_checkpoints:]
            for checkpoint in old_checkpoints:
                self.delete_checkpoint(checkpoint['id'])

    def _find_stable_checkpoints(self) -> List[Dict[str, Any]]:
        """Find stable checkpoints

        Returns:
            List of stable checkpoints
        """
        checkpoints = self.list_checkpoints()
        stable_checkpoints = []

        for checkpoint in checkpoints:
            # Select only checkpoints that pass integrity validation
            if self.validate_checkpoint_integrity(checkpoint['id']):
                stable_checkpoints.append(checkpoint)

        return stable_checkpoints

    def _log_rollback(self, checkpoint_id: str, checkpoint: Checkpoint) -> None:
        """Log rollback

        Args:
            checkpoint_id: Rolled back checkpoint ID
            checkpoint: Checkpoint information
        """
        log_file = self.checkpoints_dir / "rollback.log"

        log_entry = {
            "timestamp": datetime.now().isoformat(),
            "checkpoint_id": checkpoint_id,
            "description": checkpoint.description,
            "file_count": len(checkpoint.file_states)
        }

        try:
            with open(log_file, 'a', encoding='utf-8') as f:
                f.write(json.dumps(log_entry, ensure_ascii=False) + '\n')
        except Exception:
            pass

    def get_rollback_history(self) -> List[Dict[str, Any]]:
        """Get rollback history

        Returns:
            List of rollback history
        """
        log_file = self.checkpoints_dir / "rollback.log"
        history = []

        if log_file.exists():
            try:
                with open(log_file, 'r', encoding='utf-8') as f:
                    for line in f:
                        if line.strip():
                            history.append(json.loads(line.strip()))
            except Exception:
                pass

        # Sort by newest first
        history.sort(key=lambda x: x.get('timestamp', ''), reverse=True)
        return history
