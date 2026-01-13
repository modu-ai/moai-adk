"""Template backup manager (SPEC-INIT-003 v0.3.0).

Creates and manages backups to protect user data during template updates.
"""

from __future__ import annotations

import json
import re
import shutil
from datetime import datetime
from pathlib import Path


class TemplateBackup:
    """Create and manage template backups."""

    # Paths excluded from backups (protect user data) - MUST match TemplateProcessor.PROTECTED_PATHS
    # Note: These are relative paths within .moai/ directory
    BACKUP_EXCLUDE_DIRS = [
        "specs",  # User SPEC documents
        "reports",  # User reports
        "project",  # User project documents (product/structure/tech.md)
        "config/sections",  # User configuration section files (YAML)
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
        tracked_items = [".moai", ".claude", ".github", "CLAUDE.md", ".mcp.json", ".lsp.json", ".git-hooks"]
        return any((self.target_path / item).exists() for item in tracked_items)

    def create_backup(self) -> Path:
        """Create a timestamped backup under .moai-backups/.

        Creates a new timestamped backup directory for each update.
        Maintains backward compatibility by supporting both new and legacy structures.
        Creates backup_metadata.json for backup tracking and restoration.

        Returns:
            Path to timestamped backup directory (e.g., .moai-backups/20241201_143022/).
        """
        # Generate timestamp for backup directory name
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_path = self.target_path / ".moai-backups" / timestamp

        backup_path.mkdir(parents=True, exist_ok=True)

        # Track backed up items for metadata
        backed_up_items: list[str] = []
        excluded_items: list[str] = []

        # Copy backup targets
        for item in [".moai", ".claude", ".github", "CLAUDE.md", ".mcp.json", ".lsp.json", ".git-hooks"]:
            src = self.target_path / item
            if not src.exists():
                continue

            dst = backup_path / item

            if item == ".moai":
                # Copy while skipping protected paths
                excluded = self._copy_exclude_protected(src, dst)
                backed_up_items.append(item)
                excluded_items.extend(excluded)
            elif src.is_dir():
                shutil.copytree(src, dst, dirs_exist_ok=True)
                backed_up_items.append(item)
            else:
                shutil.copy2(src, dst)
                backed_up_items.append(item)

        # Create backup metadata
        metadata = {
            "timestamp": timestamp,
            "description": "template_backup",
            "backed_up_items": backed_up_items,
            "excluded_items": excluded_items,
            "excluded_dirs": self.BACKUP_EXCLUDE_DIRS,
            "project_root": str(self.target_path),
            "backup_type": "template",
        }

        metadata_path = backup_path / "backup_metadata.json"
        with open(metadata_path, "w", encoding="utf-8") as f:
            json.dump(metadata, f, indent=2, ensure_ascii=False)

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

    def list_backups(self) -> list[Path]:
        """List all timestamped backup directories.

        Returns:
            List of backup directory paths, sorted by timestamp (newest first).
        """
        backups: list[Path] = []
        backup_dir = self.target_path / ".moai-backups"

        if not backup_dir.exists():
            return backups

        # Match pattern: YYYYMMDD_HHMMSS (8 digits + underscore + 6 digits)
        timestamp_pattern = re.compile(r"^\d{8}_\d{6}$")
        timestamped_backups = [d for d in backup_dir.iterdir() if d.is_dir() and timestamp_pattern.match(d.name)]

        # Sort by name (timestamp) in descending order (newest first)
        backups = sorted(timestamped_backups, key=lambda x: x.name, reverse=True)
        return backups

    def cleanup_old_backups(self, keep_count: int = 5) -> int:
        """Clean up old backups, keeping only the most recent ones.

        Args:
            keep_count: Number of recent backups to keep (default: 5).

        Returns:
            Number of backups deleted.
        """
        backups = self.list_backups()

        if len(backups) <= keep_count:
            return 0

        deleted_count = 0
        for backup_path in backups[keep_count:]:
            try:
                shutil.rmtree(backup_path)
                deleted_count += 1
            except Exception:
                # Ignore deletion errors and continue with other backups
                pass

        return deleted_count

    def _copy_exclude_protected(self, src: Path, dst: Path) -> list[str]:
        """Copy backup content while excluding protected paths.

        Args:
            src: Source directory.
            dst: Destination directory.

        Returns:
            List of excluded relative paths.
        """
        dst.mkdir(parents=True, exist_ok=True)

        excluded: list[str] = []

        for item in src.rglob("*"):
            rel_path = item.relative_to(src)
            rel_path_str = str(rel_path)

            # Skip excluded paths
            if any(rel_path_str.startswith(exclude_dir) for exclude_dir in self.BACKUP_EXCLUDE_DIRS):
                excluded.append(rel_path_str)
                continue

            dst_item = dst / rel_path
            if item.is_file():
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(item, dst_item)
            elif item.is_dir():
                dst_item.mkdir(parents=True, exist_ok=True)

        return excluded

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
        for item in [".moai", ".claude", ".github", "CLAUDE.md", ".mcp.json", ".lsp.json", ".git-hooks"]:
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
