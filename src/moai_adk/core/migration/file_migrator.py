"""
File migration module for MoAI-ADK version upgrades

Handles the actual file movement and directory creation
during migration processes.
"""

import logging
import shutil
from pathlib import Path
from typing import Any, Dict, List

logger = logging.getLogger(__name__)


class FileMigrator:
    """Handles file operations during migrations"""

    def __init__(self, project_root: Path):
        """
        Initialize file migrator

        Args:
            project_root: Root directory of the project
        """
        self.project_root = Path(project_root)
        self.moved_files: List[Dict[str, str]] = []
        self.created_dirs: List[str] = []

    def create_directory(self, directory: Path) -> bool:
        """
        Create a directory if it doesn't exist

        Args:
            directory: Directory path to create

        Returns:
            True if directory was created or already exists
        """
        try:
            directory = Path(directory)
            directory.mkdir(parents=True, exist_ok=True)
            self.created_dirs.append(str(directory))
            logger.debug(f"Created directory: {directory}")
            return True
        except Exception as e:
            logger.error(f"Failed to create directory {directory}: {e}")
            return False

    def move_file(self, source: Path, destination: Path, copy_instead: bool = True) -> bool:
        """
        Move a file from source to destination

        Args:
            source: Source file path
            destination: Destination file path
            copy_instead: If True, copy instead of move (safer)

        Returns:
            True if operation was successful
        """
        source = Path(source)
        destination = Path(destination)

        if not source.exists():
            logger.warning(f"Source file not found: {source}")
            return False

        if destination.exists():
            logger.info(f"Destination already exists, skipping: {destination}")
            return True

        try:
            # Ensure destination directory exists
            destination.parent.mkdir(parents=True, exist_ok=True)

            # Copy or move file
            if copy_instead:
                shutil.copy2(source, destination)
                logger.debug(f"Copied: {source} → {destination}")
            else:
                shutil.move(str(source), str(destination))
                logger.debug(f"Moved: {source} → {destination}")

            # Record operation
            self.moved_files.append({"from": str(source), "to": str(destination)})

            return True

        except Exception as e:
            logger.error(f"Failed to move file {source} to {destination}: {e}")
            return False

    def delete_file(self, file_path: Path, safe: bool = True) -> bool:
        """
        Delete a file

        Args:
            file_path: Path to the file to delete
            safe: If True, only delete if it's a known safe file

        Returns:
            True if deletion was successful
        """
        file_path = Path(file_path)

        if not file_path.exists():
            logger.debug(f"File already deleted: {file_path}")
            return True

        # Safety check for safe mode
        if safe:
            safe_patterns = [
                ".moai/config.json",
                ".claude/statusline-config.yaml",
            ]
            is_safe = any(str(file_path.relative_to(self.project_root)).endswith(pattern) for pattern in safe_patterns)
            if not is_safe:
                logger.warning(f"Refusing to delete non-safe file: {file_path}")
                return False

        try:
            file_path.unlink()
            logger.debug(f"Deleted: {file_path}")
            return True
        except Exception as e:
            logger.error(f"Failed to delete file {file_path}: {e}")
            return False

    def execute_migration_plan(self, plan: Dict[str, Any]) -> Dict[str, Any]:
        """
        Execute a migration plan

        Args:
            plan: Migration plan dictionary with 'create', 'move', 'cleanup' keys

        Returns:
            Dictionary with execution results
        """
        results: Dict[str, Any] = {
            "success": True,
            "created_dirs": 0,
            "moved_files": 0,
            "cleaned_files": 0,
            "errors": [],
        }

        # Create directories
        for directory in plan.get("create", []):
            dir_path = self.project_root / directory
            if self.create_directory(dir_path):
                results["created_dirs"] += 1
            else:
                results["errors"].append(f"Failed to create directory: {directory}")
                results["success"] = False

        # Move files
        for move_op in plan.get("move", []):
            source = self.project_root / move_op["from"]
            dest = self.project_root / move_op["to"]

            if self.move_file(source, dest, copy_instead=True):
                results["moved_files"] += 1
                logger.info(f"✅ {move_op['description']}: {move_op['from']} → {move_op['to']}")
            else:
                results["errors"].append(f"Failed to move: {move_op['from']} → {move_op['to']}")
                results["success"] = False

        return results

    def cleanup_old_files(self, cleanup_list: List[str], dry_run: bool = False) -> int:
        """
        Clean up old files after successful migration

        Args:
            cleanup_list: List of file paths to clean up
            dry_run: If True, only show what would be deleted

        Returns:
            Number of files cleaned up
        """
        cleaned = 0

        for file_path in cleanup_list:
            full_path = self.project_root / file_path

            if dry_run:
                if full_path.exists():
                    logger.info(f"Would delete: {file_path}")
                    cleaned += 1
            else:
                if self.delete_file(full_path, safe=True):
                    logger.info(f"🗑️  Cleaned up: {file_path}")
                    cleaned += 1

        return cleaned

    def get_migration_summary(self) -> Dict[str, Any]:
        """
        Get summary of migration operations performed

        Returns:
            Dictionary with migration summary
        """
        return {
            "moved_files": len(self.moved_files),
            "created_directories": len(self.created_dirs),
            "operations": self.moved_files,
        }
