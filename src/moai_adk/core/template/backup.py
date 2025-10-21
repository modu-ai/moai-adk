# @CODE:TEMPLATE-001 | SPEC: SPEC-INIT-003/spec.md | Chain: TEMPLATE-001
"""Template backup manager (SPEC-INIT-003 v0.3.0).

Creates and manages backups to protect user data during template updates.
"""

from __future__ import annotations

import shutil
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
        return any(
            (self.target_path / item).exists()
            for item in [".moai", ".claude", ".github", "CLAUDE.md"]
        )

    def create_backup(self) -> Path:
        """Create a single backup (always at .moai-backups/backup/).

        Existing backups are overwritten to maintain only one backup copy.

        Returns:
            Backup path (always .moai-backups/backup/).
        """
        backup_path = self.target_path / ".moai-backups" / "backup"

        # Remove existing backup if present
        if backup_path.exists():
            shutil.rmtree(backup_path)

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
            if any(
                rel_path_str.startswith(exclude_dir)
                for exclude_dir in self.BACKUP_EXCLUDE_DIRS
            ):
                continue

            dst_item = dst / rel_path
            if item.is_file():
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(item, dst_item)
            elif item.is_dir():
                dst_item.mkdir(parents=True, exist_ok=True)
