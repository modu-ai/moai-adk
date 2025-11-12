"""
Event Detector - Identify risky operations.

SPEC: .moai/specs/SPEC-CHECKPOINT-EVENT-001/spec.md
"""

from pathlib import Path


class EventDetector:
    """Detect potentially risky operations."""

    CRITICAL_FILES = {
        "CLAUDE.md",
        "config.json",
        ".moai/config/config.json",
    }

    CRITICAL_DIRS = {
        ".moai/memory",
    }

    def is_risky_deletion(self, deleted_files: list[str]) -> bool:
        """
        Detect large-scale file deletions.

        SPEC requirement: deleting 10 or more files counts as risky.

        Args:
            deleted_files: Files slated for deletion.

        Returns:
            True when 10 or more files are deleted, otherwise False.
        """
        return len(deleted_files) >= 10

    def is_risky_refactoring(self, renamed_files: list[tuple[str, str]]) -> bool:
        """
        Detect large-scale refactoring.

        SPEC requirement: renaming 10 or more files counts as risky.

        Args:
            renamed_files: List of (old_name, new_name) pairs.

        Returns:
            True when 10 or more files are renamed, otherwise False.
        """
        return len(renamed_files) >= 10

    def is_critical_file(self, file_path: Path) -> bool:
        """
        Determine whether the file is critical.

        SPEC requirement: modifying CLAUDE.md, config.json, or .moai/memory/*.md is risky.

        Args:
            file_path: File path to inspect.

        Returns:
            True when the file is critical, otherwise False.
        """
        # Check whether the file name is in the critical list
        if file_path.name in self.CRITICAL_FILES:
            return True

        # Convert to string for further checks
        path_str = str(file_path)

        # Detect .moai/config/config.json paths
        if ".moai/config/config.json" in path_str or ".moai\\config.json" in path_str:
            return True

        # Detect files inside the .moai/memory/ directory
        for critical_dir in self.CRITICAL_DIRS:
            if critical_dir in path_str or critical_dir.replace("/", "\\") in path_str:
                return True

        return False
