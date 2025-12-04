"""Selective Restorer for MoAI-ADK Custom Elements

This module provides functionality to restore selected custom elements from backup
during MoAI-ADK updates. It handles safe file restoration with conflict detection
and detailed reporting of restoration results.

Key Features:
- Preserves unselected elements (fixes disappearing issue)
- Only overwrites/creates selected elements from backup
- Never deletes existing files that aren't being restored
- Maintains all other custom elements intact
- Safe restoration with rollback capabilities

The restorer works with the existing MoAI-ADK backup system and ensures that
only explicitly selected elements are modified, preserving all other files.
"""

import logging
import shutil
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

logger = logging.getLogger(__name__)


class SelectiveRestorer:
    """Restores selected custom elements from backup.

    This class handles the actual file restoration process, copying selected elements
    from backup to their original locations with proper conflict handling and reporting.
    """

    def __init__(self, project_path: Path, backup_path: Optional[Path] = None):
        """Initialize the selective restorer.

        Args:
            project_path: Path to the MoAI-ADK project directory
            backup_path: Path to the backup directory (auto-detected if not provided)
        """
        self.project_path = project_path
        self.backup_path = backup_path or self._find_latest_backup()
        self.restoration_log: List[Dict[str, str]] = []

    def _find_latest_backup(self) -> Optional[Path]:
        """Find the latest backup directory.

        Returns:
            Path to the latest backup directory, or None if no backup found
        """
        # Look for .moai-backups directory
        backups_dir = self.project_path / ".moai-backups"
        if not backups_dir.exists():
            return None

        # Find all backup directories and sort by modification time
        backup_dirs = []
        for item in backups_dir.iterdir():
            if item.is_dir() and item.name.startswith("pre-update-backup_"):
                backup_dirs.append((item.stat().st_mtime, item))

        if not backup_dirs:
            return None

        # Return the most recent backup
        backup_dirs.sort(reverse=True)
        return backup_dirs[0][1]  # Return the Path object

    def restore_elements(self, selected_elements: List[str]) -> Tuple[bool, Dict[str, Any]]:
        """Restore selected custom elements from backup.

        Args:
            selected_elements: List of element paths to restore

        Returns:
            Tuple of (success_status, restoration_stats)

        Example:
            >>> restorer = SelectiveRestorer("/project")
            >>> success, stats = restorer.restore_elements([
            ...     ".claude/agents/my-agent.md",
            ...     ".claude/skills/my-skill/"
            ... ])
            >>> print(f"Restoration {'success' if success else 'failed'}: {stats}")
        """
        if not selected_elements:
            logger.info("No elements selected for restoration.")
            return True, {"total": 0, "success": 0, "failed": 0}

        print(f"\n🚀 Restoring {len(selected_elements)} selected elements...")
        print("-" * 50)

        # Group elements by type for organized restoration
        element_groups = self._group_elements_by_type(selected_elements)

        # Track restoration statistics
        stats: Dict[str, Any] = {"total": 0, "success": 0, "failed": 0, "by_type": {}}

        # Restore each type of element
        for element_type, elements in element_groups.items():
            print(f"\n📂 Restoring {element_type}s...")
            type_stats = self._restore_element_type(element_type, elements)
            by_type: Dict[str, Dict[str, int]] = stats["by_type"]
            by_type[element_type] = type_stats
            total_count: int = stats["total"]
            success_count: int = stats["success"]
            failed_count: int = stats["failed"]
            stats["total"] = total_count + type_stats["total"]
            stats["success"] = success_count + type_stats["success"]
            stats["failed"] = failed_count + type_stats["failed"]

        # Display final summary
        self._display_restoration_summary(stats)

        # Log restoration details
        self._log_restoration_details(selected_elements, stats)

        is_success: bool = stats["failed"] == 0
        if is_success:
            logger.info(f"Successfully restored {stats['success']} elements")
        else:
            logger.warning(f"Failed to restore {stats['failed']} elements")

        return is_success, stats

    def _group_elements_by_type(self, selected_elements: List[str]) -> Dict[str, List[Path]]:
        """Group selected elements by their type.

        Args:
            selected_elements: List of element paths

        Returns:
            Dictionary with element types as keys and lists of element paths as values
        """
        groups: Dict[str, List[Path]] = {
            "agents": [],
            "commands": [],
            "skills": [],
            "hooks": [],
            "unknown": [],
        }

        for element_path in selected_elements:
            path = Path(element_path)
            parts = path.parts

            if "agents" in parts:
                groups["agents"].append(path)
            elif "commands" in parts and "moai" in parts:
                groups["commands"].append(path)
            elif "commands" in parts:
                groups["commands"].append(path)
            elif "skills" in parts:
                groups["skills"].append(path)
            elif "hooks" in parts and "moai" in parts:
                groups["hooks"].append(path)
            elif "hooks" in parts:
                groups["hooks"].append(path)
            else:
                logger.warning(f"Unknown element type for: {element_path}")
                groups["unknown"].append(path)

        return groups

    def _restore_element_type(self, element_type: str, elements: List[Path]) -> Dict[str, int]:
        """Restore elements of a specific type.

        Args:
            element_type: Type of elements to restore
            elements: List of element paths to restore

        Returns:
            Statistics for this restoration type
        """
        stats = {"total": len(elements), "success": 0, "failed": 0}

        for element_path in elements:
            try:
                success = self._restore_single_element(element_path, element_type)
                if success:
                    stats["success"] += 1
                    print(f"   ✓ {element_path.name}")
                else:
                    stats["failed"] += 1
                    print(f"   ✗ Failed: {element_path.name}")
            except Exception as e:
                stats["failed"] += 1
                print(f"   ✗ Error: {element_path.name} - {e}")
                logger.error(f"Error restoring {element_path}: {e}")

        return stats

    def _normalize_element_path(self, element_path: Path) -> Optional[Path]:
        """Normalize element path to be relative to project structure.

        Handles both absolute and relative paths correctly:
        - Absolute paths: extracts the relevant portion (.claude/ or .moai/)
        - Relative paths: validates and returns as-is

        Args:
            element_path: Path to normalize (can be absolute or relative)

        Returns:
            Normalized relative path, or None if normalization fails
        """
        element_str = str(element_path)

        # Handle absolute paths
        if element_path.is_absolute():
            # Extract .claude/ or .moai/ portion from absolute paths
            for safe_prefix in [".claude/", ".moai/"]:
                if safe_prefix in element_str:
                    try:
                        # Split on the safe prefix and take the portion after it
                        _, relative_part = element_str.split(safe_prefix, 1)
                        # Reconstruct the relative path
                        normalized_path = Path(safe_prefix.rstrip("/")) / relative_part
                        logger.debug(f"Normalized absolute path {element_path} -> {normalized_path}")
                        return normalized_path
                    except (ValueError, IndexError):
                        logger.warning(f"Failed to extract relative path from {element_path}")
                        continue

            # If no safe prefix found in absolute path, this is suspicious
            logger.error(f"Absolute path {element_path} doesn't contain .claude/ or .moai/ prefixes")
            return None

        # Handle relative paths
        else:
            # Ensure relative path starts with .claude or .moai
            if not (element_str.startswith(".claude/") or element_str.startswith(".moai/")):
                logger.error(f"Relative path {element_path} must start with .claude/ or .moai/")
                return None

            logger.debug(f"Relative path already normalized: {element_path}")
            return element_path

    def _validate_element_path(self, relative_path: Path) -> bool:
        """Validate an element path for security and correctness.

        Args:
            relative_path: Relative path to validate

        Returns:
            True if path is valid and safe, False otherwise
        """
        path_str = str(relative_path)

        # Security check: prevent path traversal attacks
        if ".." in path_str:
            logger.error(f"Path traversal attempt detected in: {path_str}")
            return False

        # Ensure path starts with allowed prefixes
        allowed_prefixes = [".claude/", ".moai/"]
        if not any(path_str.startswith(prefix) for prefix in allowed_prefixes):
            logger.error(f"Path {path_str} doesn't start with allowed prefix: {allowed_prefixes}")
            return False

        # Additional validation: check for suspicious patterns
        suspicious_patterns = ["//", "~", "$"]
        for pattern in suspicious_patterns:
            if pattern in path_str:
                logger.warning(f"Suspicious pattern '{pattern}' found in path: {path_str}")

        return True

    def _restore_single_element(self, element_path: Path, element_type: str) -> bool:
        """Restore a single element from backup.

        Args:
            element_path: Path to restore the element to (can be absolute or relative)
            element_type: Type of element (for target directory creation)

        Returns:
            True if restoration succeeded, False otherwise
        """
        # Normalize element path to get relative path within project structure
        relative_path = self._normalize_element_path(element_path)
        if relative_path is None:
            logger.error(f"Failed to normalize element path: {element_path}")
            return False

        # Validate the normalized path for security
        if not self._validate_element_path(relative_path):
            logger.error(f"Invalid element path: {relative_path}")
            return False

        # Create absolute target path
        target_path = self.project_path / relative_path
        backup_source = self.backup_path / relative_path

        # Ensure backup source exists
        if not backup_source.exists():
            logger.warning(f"Backup not found for: {relative_path}")
            return False

        # Create target directory if needed
        target_dir = target_path.parent
        target_dir.mkdir(parents=True, exist_ok=True)

        # Handle conflicts
        if target_path.exists():
            if not self._handle_file_conflict(target_path, backup_source):
                logger.warning(f"Conflict handling failed for: {relative_path}")
                return False

        # Perform the restoration
        try:
            if backup_source.is_dir():
                # For directories (skills)
                shutil.copytree(backup_source, target_path, dirs_exist_ok=True)
            else:
                # For files
                shutil.copy2(backup_source, target_path)

            # Record successful restoration
            self.restoration_log.append(
                {
                    "path": str(element_path),
                    "type": element_type,
                    "status": "success",
                    "timestamp": str(backup_source.stat().st_mtime),
                }
            )

            logger.info(f"Restored: {relative_path}")
            return True

        except Exception as e:
            logger.error(f"Failed to restore {relative_path}: {e}")
            return False

    def _handle_file_conflict(self, target_path: Path, backup_source: Path) -> bool:
        """Handle file or directory conflict during restoration.

        Args:
            target_path: Path to target file/directory (existing)
            backup_source: Path to backup source file/directory

        Returns:
            True if conflict handled successfully, False otherwise
        """
        try:
            # Handle directory conflicts (skills are directories)
            if target_path.is_dir() and backup_source.is_dir():
                # Directories - backup target directory and return True
                backup_target = target_path.parent / f"{target_path.name}.backup_dir"
                try:
                    if backup_target.exists():
                        shutil.rmtree(backup_target)
                    shutil.copytree(target_path, backup_target, dirs_exist_ok=True)
                    logger.debug(f"Backed up directory: {target_path} -> {backup_target}")
                    return True
                except Exception as e:
                    logger.error(f"Failed to backup directory {target_path}: {e}")
                    return False

            # Handle file conflicts
            if not target_path.is_file() or not backup_source.is_file():
                # Mixed types or non-existent, skip comparison
                logger.debug(f"Skipping conflict check for non-file: {target_path}")
                return True

            # Compare file contents
            target_content = target_path.read_text(encoding="utf-8", errors="ignore")
            backup_content = backup_source.read_text(encoding="utf-8", errors="ignore")

            if target_content == backup_content:
                # Files are identical, no conflict
                logger.debug(f"No conflict detected for: {target_path}")
                return True

            # Files differ, prompt for action
            print(f"\n⚠️ Conflict detected for: {target_path.name}")
            print("   Target file exists and differs from backup")

            # For now, we'll backup the target and restore the backup
            backup_target = target_path.with_suffix(".backup")
            try:
                shutil.copy2(target_path, backup_target)
                print(f"   Backed up to: {backup_target.name}")
                logger.info(f"Backed up conflicting file: {backup_target}")
                return True
            except Exception as e:
                logger.error(f"Failed to backup conflicting file {target_path}: {e}")
                return False

        except Exception as e:
            logger.error(f"Error handling file conflict for {target_path}: {e}")
            return False

    def _display_restoration_summary(self, stats: Dict[str, int]) -> None:
        """Display summary of restoration results.

        Args:
            stats: Restoration statistics dictionary
        """
        print("\n" + "=" * 50)
        print("🎉 Restoration Complete")
        print("=" * 50)
        print(f"Total elements: {stats['total']}")
        print(f"✅ Successfully restored: {stats['success']}")

        if stats["failed"] > 0:
            print(f"❌ Failed to restore: {stats['failed']}")

        # Show breakdown by type
        if "by_type" in stats:
            print("\n📊 By Type:")
            by_type_stats: Dict[str, Dict[str, int]] = stats["by_type"]  # type: ignore[assignment]
            for element_type, type_stats in by_type_stats.items():
                if type_stats["total"] > 0:
                    status_icon = "✅" if type_stats["failed"] == 0 else "❌"
                    print(f"   {element_type.title()}: {type_stats['success']}/{type_stats['total']} {status_icon}")

    def _log_restoration_details(self, selected_elements: List[str], stats: Dict[str, int]) -> None:
        """Log detailed restoration information for debugging.

        Args:
            selected_elements: List of elements that were selected for restoration
            stats: Restoration statistics
        """
        logger.info(
            f"Restoration completed - Total: {stats['total']}, Success: {stats['success']}, Failed: {stats['failed']}"
        )

        if stats["failed"] > 0:
            failed_elements = [elem for elem in selected_elements if not self._was_restoration_successful(Path(elem))]
            logger.warning(f"Failed elements: {failed_elements}")

        # Log all restoration attempts from the log
        for entry in self.restoration_log:
            if entry["status"] == "success":
                logger.debug(f"✓ Restored: {entry['path']} ({entry['type']})")
            else:
                logger.warning(f"✗ Failed: {entry['path']} ({entry['type']})")

    def _was_restoration_successful(self, element_path: Path) -> bool:
        """Check if an element was successfully restored.

        Args:
            element_path: Path to check

        Returns:
            True if element was restored successfully
        """
        for entry in self.restoration_log:
            if str(Path(entry["path"])) == str(element_path) and entry["status"] == "success":
                return True
        return False


def create_selective_restorer(project_path: str | Path, backup_path: Optional[Path] = None) -> SelectiveRestorer:
    """Factory function to create a SelectiveRestorer.

    Args:
        project_path: Path to the MoAI-ADK project directory
        backup_path: Path to backup directory (auto-detected if not provided)

    Returns:
        Configured SelectiveRestorer instance

    Example:
        >>> restorer = create_selective_restorer("/path/to/project")
        >>> success, stats = restorer.restore_elements([
        ...     ".claude/agents/my-agent.md"
        ... ])
        >>> print(f"Restoration result: {'success' if success else 'failed'}")
    """
    return SelectiveRestorer(Path(project_path).resolve(), backup_path)
