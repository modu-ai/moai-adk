"""Template backup manager (SPEC-INIT-003 v0.3.0).

Creates and manages backups to protect user data during template updates.
"""

from __future__ import annotations

import re
import shutil
from datetime import datetime
from pathlib import Path


class TemplateBackup:
    """Create and manage template backups."""

    # Paths excluded from backups (protect user data)
    BACKUP_EXCLUDE_DIRS = [
        "specs",  # User SPEC documents
        "reports",  # User reports
    ]

    def __init__(self, target_path: Path) -> None:
        """Initialize the backup manager.

        Args:
            target_path: Project path (absolute).
        """
        self.target_path = target_path.resolve()

    @property
    def backup_dir(self) -> Path:
        """Get the backup directory path.

        Returns:
            Path to .moai-backups directory.
        """
        return self.target_path / ".moai-backups"

    def has_existing_files(self) -> bool:
        """Check whether backup-worthy files already exist.

        Returns:
            True when any tracked file exists.
        """
        return any((self.target_path / item).exists() for item in [".moai", ".claude", ".github", "CLAUDE.md"])

    def create_backup(self) -> Path:
        """Create a timestamped backup under .moai-backups/.

        Creates a new timestamped backup directory for each update.
        Maintains backward compatibility by supporting both new and legacy structures.

        Returns:
            Path to timestamped backup directory (e.g., .moai-backups/20241201_143022/).
        """
        # Generate timestamp for backup directory name
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_path = self.target_path / ".moai-backups" / timestamp

        backup_path.mkdir(parents=True, exist_ok=True)

        # Copy backup targets
        for item in [".moai", ".claude", ".github", "CLAUDE.md"]:
            src = self.target_path / item
            if not src.exists():
                continue

            dst = backup_path / item

            if item == ".moai":
                # Copy while skipping protected paths
                self._copy_exclude_protected(src, dst)
            elif src.is_dir():
                shutil.copytree(src, dst, dirs_exist_ok=True)
            else:
                shutil.copy2(src, dst)

        return backup_path

    def get_latest_backup(self) -> Path | None:
        """Get the most recent backup, supporting both new and legacy structures.

        Searches for backups in order of preference:
        1. Latest timestamped backup (new structure)
        2. Legacy backup/ directory (old structure)

        Returns:
            Path to the most recent backup directory, or None if no backups exist.
        """
        # Check for new timestamped backups first
        backup_dir = self.target_path / ".moai-backups"
        if backup_dir.exists():
            # Match pattern: YYYYMMDD_HHMMSS (8 digits + underscore + 6 digits)
            timestamp_pattern = re.compile(r"^\d{8}_\d{6}$")
            timestamped_backups = [d for d in backup_dir.iterdir() if d.is_dir() and timestamp_pattern.match(d.name)]

            if timestamped_backups:
                # Sort by name (timestamp) and return the latest
                latest_backup = max(timestamped_backups, key=lambda x: x.name)
                return latest_backup

        # Fall back to legacy backup/ directory
        legacy_backup = backup_dir / "backup"
        if legacy_backup.exists():
            return legacy_backup

        return None

    def _copy_exclude_protected(self, src: Path, dst: Path) -> None:
        """Copy backup content while excluding protected paths.

        Args:
            src: Source directory.
            dst: Destination directory.
        """
        dst.mkdir(parents=True, exist_ok=True)

        for item in src.rglob("*"):
            rel_path = item.relative_to(src)
            rel_path_str = str(rel_path)

            # Skip excluded paths
            if any(rel_path_str.startswith(exclude_dir) for exclude_dir in self.BACKUP_EXCLUDE_DIRS):
                continue

            dst_item = dst / rel_path
            if item.is_file():
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(item, dst_item)
            elif item.is_dir():
                dst_item.mkdir(parents=True, exist_ok=True)

    def restore_backup(self, backup_path: Path | None = None) -> None:
        """Restore project files from backup.

        Restores .moai, .claude, .github directories and CLAUDE.md file
        from a backup created by create_backup().
        Supports both new timestamped and legacy backup structures.

        Args:
            backup_path: Backup path to restore from.
                        If None, automatically finds the latest backup.

        Raises:
            FileNotFoundError: When no backup is found.
        """
        if backup_path is None:
            backup_path = self.get_latest_backup()

        if backup_path is None or not backup_path.exists():
            raise FileNotFoundError(f"Backup not found: {backup_path}")

        # Restore each item from backup
        for item in [".moai", ".claude", ".github", "CLAUDE.md"]:
            src = backup_path / item
            dst = self.target_path / item

            # Skip if not in backup
            if not src.exists():
                continue

            # Remove current version
            if dst.exists():
                if dst.is_dir():
                    shutil.rmtree(dst)
                else:
                    dst.unlink()

            # Restore from backup
            if src.is_dir():
                shutil.copytree(src, dst, dirs_exist_ok=True)
            else:
                shutil.copy2(src, dst)
