"""
Backup management module for MoAI-ADK migrations

Creates and manages backups during migration processes
to ensure data safety and enable rollback.
"""

import json
import logging
import shutil
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional

logger = logging.getLogger(__name__)


class BackupManager:
    """Manages backup creation and restoration for migrations"""

    def __init__(self, project_root: Path):
        """
        Initialize backup manager

        Args:
            project_root: Root directory of the project
        """
        self.project_root = Path(project_root)
        self.backup_base_dir = self.project_root / ".moai" / "backups"
        self.backup_base_dir.mkdir(parents=True, exist_ok=True)

    def create_backup(self, description: str = "migration") -> Path:
        """
        Create a full backup of configuration files

        Args:
            description: Description of this backup

        Returns:
            Path to the backup directory
        """
        # Create timestamped backup directory
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_dir = self.backup_base_dir / f"{description}_{timestamp}"
        backup_dir.mkdir(parents=True, exist_ok=True)

        logger.info(f"Creating backup at {backup_dir}")

        # Files to backup
        backup_targets = [
            self.project_root / ".moai" / "config.json",
            self.project_root / ".moai" / "config" / "config.json",
            self.project_root / ".claude" / "statusline-config.yaml",
            self.project_root / ".moai" / "config" / "statusline-config.yaml",
        ]

        backed_up_files = []

        for target in backup_targets:
            if target.exists():
                # Preserve relative path structure in backup
                rel_path = target.relative_to(self.project_root)
                backup_path = backup_dir / rel_path

                # Create parent directories
                backup_path.parent.mkdir(parents=True, exist_ok=True)

                # Copy file
                shutil.copy2(target, backup_path)
                backed_up_files.append(str(rel_path))
                logger.debug(f"Backed up: {rel_path}")

        # Save backup metadata
        metadata = {
            "timestamp": timestamp,
            "description": description,
            "backed_up_files": backed_up_files,
            "project_root": str(self.project_root),
        }

        metadata_path = backup_dir / "backup_metadata.json"
        with open(metadata_path, "w", encoding="utf-8") as f:
            json.dump(metadata, f, indent=2, ensure_ascii=False)

        logger.info(f"✅ Backup created successfully: {backup_dir}")
        return backup_dir

    def list_backups(self) -> List[Dict[str, str]]:
        """
        List all available backups

        Returns:
            List of backup information dictionaries
        """
        backups: List[Dict[str, Any]] = []

        if not self.backup_base_dir.exists():
            return backups

        for backup_dir in sorted(self.backup_base_dir.iterdir(), reverse=True):
            if backup_dir.is_dir():
                metadata_path = backup_dir / "backup_metadata.json"
                if metadata_path.exists():
                    try:
                        with open(metadata_path, "r", encoding="utf-8") as f:
                            metadata = json.load(f)
                            backups.append(
                                {
                                    "path": str(backup_dir),
                                    "timestamp": metadata.get("timestamp", "unknown"),
                                    "description": metadata.get("description", "unknown"),
                                    "files": len(metadata.get("backed_up_files", [])),
                                }
                            )
                    except Exception as e:
                        logger.warning(f"Failed to read backup metadata: {e}")

        return backups

    def restore_backup(self, backup_path: Path) -> bool:
        """
        Restore files from a backup

        Args:
            backup_path: Path to the backup directory

        Returns:
            True if restore was successful, False otherwise
        """
        backup_path = Path(backup_path)

        if not backup_path.exists():
            logger.error(f"Backup directory not found: {backup_path}")
            return False

        metadata_path = backup_path / "backup_metadata.json"
        if not metadata_path.exists():
            logger.error(f"Backup metadata not found: {metadata_path}")
            return False

        try:
            # Read metadata
            with open(metadata_path, "r", encoding="utf-8") as f:
                metadata = json.load(f)

            logger.info(f"Restoring backup from {backup_path}")

            # Restore each file
            for rel_path in metadata.get("backed_up_files", []):
                backup_file = backup_path / rel_path
                target_file = self.project_root / rel_path

                if backup_file.exists():
                    # Create parent directories
                    target_file.parent.mkdir(parents=True, exist_ok=True)

                    # Restore file
                    shutil.copy2(backup_file, target_file)
                    logger.debug(f"Restored: {rel_path}")

            logger.info("✅ Backup restored successfully")
            return True

        except Exception as e:
            logger.error(f"Failed to restore backup: {e}")
            return False

    def cleanup_old_backups(self, keep_count: int = 5) -> int:
        """
        Clean up old backups, keeping only the most recent ones

        Args:
            keep_count: Number of recent backups to keep

        Returns:
            Number of backups deleted
        """
        backups = self.list_backups()

        if len(backups) <= keep_count:
            return 0

        deleted_count = 0
        for backup_info in backups[keep_count:]:
            backup_path = Path(backup_info["path"])
            try:
                shutil.rmtree(backup_path)
                deleted_count += 1
                logger.debug(f"Deleted old backup: {backup_path}")
            except Exception as e:
                logger.warning(f"Failed to delete backup {backup_path}: {e}")

        logger.info(f"Cleaned up {deleted_count} old backups")
        return deleted_count

    def get_latest_backup(self) -> Optional[Path]:
        """
        Get the most recent backup

        Returns:
            Path to the latest backup directory, or None if no backups exist
        """
        backups = self.list_backups()
        if backups:
            return Path(backups[0]["path"])
        return None

    def create_full_project_backup(self, description: str = "pre-update-backup") -> Path:
        """
        Create a complete backup of entire project structure before update

        Backs up:
        - .claude/ (entire directory)
        - .moai/ (entire directory)
        - CLAUDE.md (file)

        Args:
            description: Description of this backup (default: "pre-update-backup")

        Returns:
            Path to the backup directory
        """
        # Create timestamped backup directory
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_dir = self.project_root / ".moai-backups" / f"{description}_{timestamp}"
        backup_dir.mkdir(parents=True, exist_ok=True)

        logger.info(f"Creating full project backup at {backup_dir}")

        # Directories and files to backup
        backup_targets = [
            (self.project_root / ".claude", True),  # (path, is_directory)
            (self.project_root / ".moai", True),
            (self.project_root / "CLAUDE.md", False),
        ]

        backed_up_items = []

        for target_path, is_dir in backup_targets:
            if not target_path.exists():
                continue

            try:
                rel_path = target_path.relative_to(self.project_root)
                backup_path = backup_dir / rel_path

                if is_dir:
                    # Backup directory
                    shutil.copytree(target_path, backup_path, dirs_exist_ok=True)
                    backed_up_items.append(str(rel_path))
                    logger.debug(f"Backed up directory: {rel_path}")
                else:
                    # Backup file
                    backup_path.parent.mkdir(parents=True, exist_ok=True)
                    shutil.copy2(target_path, backup_path)
                    backed_up_items.append(str(rel_path))
                    logger.debug(f"Backed up file: {rel_path}")

            except Exception as e:
                logger.error(f"Failed to backup {target_path}: {e}")
                raise

        # Save backup metadata
        metadata = {
            "timestamp": timestamp,
            "description": description,
            "backed_up_items": backed_up_items,
            "project_root": str(self.project_root),
            "backup_type": "full_project",
        }

        metadata_path = backup_dir / "backup_metadata.json"
        with open(metadata_path, "w", encoding="utf-8") as f:
            json.dump(metadata, f, indent=2, ensure_ascii=False)

        logger.info(f"✅ Full project backup created successfully: {backup_dir}")
        return backup_dir
