"""TAG linkage manager module for TAG System v2.0.

Implements T4: Linkage Manager for bidirectional TAG↔CODE mapping
with atomic database operations.
"""

from __future__ import annotations

import json
import logging
from pathlib import Path
from typing import Any

from . import tag_validator

# Import atomic_write functions directly
from .atomic_write import atomic_write_json, atomic_write_text

logger = logging.getLogger(__name__)


def spec_document_exists(spec_id: str) -> bool:
    """Check if SPEC document exists in .moai/specs/.

    Args:
        spec_id: SPEC identifier to check

    Returns:
        True if SPEC document exists, False otherwise
    """
    # Get specs directory from current working directory
    specs_dir = Path.cwd() / ".moai" / "specs"

    if not specs_dir.exists():
        return False

    # Check for SPEC directory
    spec_dir = specs_dir / spec_id
    return spec_dir.exists() and spec_dir.is_dir()


class LinkageManager:
    """Manage bidirectional TAG↔CODE mapping database.

    Provides atomic database operations for TAG tracking,
    with support for TAG→Files and File→TAGs lookups.

    Attributes:
        db_path: Path to linkage database file
    """

    def __init__(self, db_path: Path) -> None:
        """Initialize linkage manager.

        Args:
            db_path: Path to linkage database JSON file
        """
        self.db_path = Path(db_path)
        self._ensure_database()

    def _ensure_database(self) -> None:
        """Ensure database file exists with proper structure."""
        if not self.db_path.exists():
            # Create new database with empty structure
            self._write_database({"tags": [], "files": {}})

    def _load_database(self) -> dict[str, Any]:
        """Load database from disk.

        Returns:
            Database dictionary with 'tags' and 'files' keys
        """
        try:
            if not self.db_path.exists():
                return {"tags": [], "files": {}}

            with open(self.db_path, "r", encoding="utf-8") as f:
                data = json.load(f)

            # Validate structure
            if not isinstance(data, dict):
                logger.warning(f"Invalid database structure in {self.db_path}")
                return {"tags": [], "files": {}}

            if "tags" not in data:
                data["tags"] = []
            if "files" not in data:
                data["files"] = {}

            return data

        except (json.JSONDecodeError, OSError) as e:
            logger.error(f"Error loading database {self.db_path}: {e}")
            return {"tags": [], "files": {}}

    def _write_database(self, data: dict[str, Any]) -> bool:
        """Write database to disk atomically.

        Args:
            data: Database dictionary to write

        Returns:
            True if write succeeded, False otherwise
        """
        return atomic_write_json(self.db_path, data)

    def add_tag(self, tag: tag_validator.TAG) -> bool:
        """Add TAG to linkage database.

        Args:
            tag: TAG object to add

        Returns:
            True if TAG added successfully, False otherwise

        Examples:
            >>> tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10)
            >>> manager.add_tag(tag)
            True
        """
        try:
            data = self._load_database()

            # Create TAG entry
            tag_entry = {
                "spec_id": tag.spec_id,
                "verb": tag.verb,
                "file_path": str(tag.file_path),
                "line": tag.line,
            }

            # Check for duplicate
            if tag_entry not in data["tags"]:
                data["tags"].append(tag_entry)

            # Update file index
            file_key = str(tag.file_path)
            if file_key not in data["files"]:
                data["files"][file_key] = []

            if tag.spec_id not in data["files"][file_key]:
                data["files"][file_key].append(tag.spec_id)

            # Write database atomically
            return self._write_database(data)

        except Exception as e:
            logger.error(f"Error adding TAG to database: {e}")
            return False

    def remove_tag(self, tag: tag_validator.TAG) -> bool:
        """Remove specific TAG from database.

        Args:
            tag: TAG object to remove

        Returns:
            True if TAG removed successfully, False otherwise
        """
        try:
            data = self._load_database()

            tag_entry = {
                "spec_id": tag.spec_id,
                "verb": tag.verb,
                "file_path": str(tag.file_path),
                "line": tag.line,
            }

            # Remove from tags list
            if tag_entry in data["tags"]:
                data["tags"].remove(tag_entry)

            # Update file index
            file_key = str(tag.file_path)
            if file_key in data["files"] and tag.spec_id in data["files"][file_key]:
                data["files"][file_key].remove(tag.spec_id)

                # Clean up empty file entries
                if not data["files"][file_key]:
                    del data["files"][file_key]

            return self._write_database(data)

        except Exception as e:
            logger.error(f"Error removing TAG from database: {e}")
            return False

    def remove_file_tags(self, file_path: Path) -> bool:
        """Remove all TAGs for a file.

        Args:
            file_path: Path to file whose TAGs should be removed

        Returns:
            True if TAGs removed successfully, False otherwise

        Examples:
            >>> manager.remove_file_tags(Path("deleted.py"))
            True
        """
        try:
            data = self._load_database()

            file_key = str(file_path)

            # Remove all TAGs for this file
            data["tags"] = [tag for tag in data["tags"] if tag["file_path"] != file_key]

            # Remove file index
            if file_key in data["files"]:
                del data["files"][file_key]

            return self._write_database(data)

        except Exception as e:
            logger.error(f"Error removing file TAGs: {e}")
            return False

    def get_all_tags(self) -> list[dict[str, Any]]:
        """Get all TAGs in database.

        Returns:
            List of TAG dictionaries
        """
        data = self._load_database()
        return data.get("tags", [])

    def get_code_locations(self, spec_id: str) -> list[dict[str, Any]]:
        """Get all code locations for a SPEC-ID.

        Args:
            spec_id: SPEC identifier to query

        Returns:
            List of location dictionaries with file_path, line, and verb

        Examples:
            >>> locations = manager.get_code_locations("SPEC-AUTH-001")
            >>> len(locations)
            3
        """
        data = self._load_database()
        tags = data.get("tags", [])

        # Filter TAGs by spec_id
        locations = [
            {
                "file_path": tag["file_path"],
                "line": tag["line"],
                "verb": tag["verb"],
            }
            for tag in tags
            if tag["spec_id"] == spec_id
        ]

        return locations

    def get_tags_by_file(self, file_path: Path) -> list[dict[str, Any]]:
        """Get all TAGs for a specific file.

        Args:
            file_path: Path to file

        Returns:
            List of TAG dictionaries for the file

        Examples:
            >>> tags = manager.get_tags_by_file(Path("auth.py"))
            >>> len(tags)
            2
        """
        data = self._load_database()
        tags = data.get("tags", [])
        file_key = str(file_path)

        # Filter TAGs by file_path
        return [
            {
                "spec_id": tag["spec_id"],
                "verb": tag["verb"],
                "file_path": tag["file_path"],
                "line": tag["line"],
            }
            for tag in tags
            if tag["file_path"] == file_key
        ]

    def get_all_spec_ids(self) -> list[str]:
        """Get all unique SPEC-IDs in database.

        Returns:
            Sorted list of unique SPEC-ID strings
        """
        data = self._load_database()
        tags = data.get("tags", [])

        spec_ids = {tag["spec_id"] for tag in tags}
        return sorted(spec_ids)

    def find_orphaned_tags(self) -> list[dict[str, Any]]:
        """Find TAGs referencing nonexistent SPEC documents.

        Returns:
            List of orphaned TAG dictionaries

        Examples:
            >>> orphans = manager.find_orphaned_tags()
            >>> len(orphans)
            1
        """
        data = self._load_database()
        tags = data.get("tags", [])

        orphans = [tag for tag in tags if not spec_document_exists(tag["spec_id"])]

        return orphans

    def clear(self) -> bool:
        """Clear all TAGs from database.

        Returns:
            True if database cleared successfully, False otherwise
        """
        return self._write_database({"tags": [], "files": {}})


__all__ = [
    "spec_document_exists",
    "LinkageManager",
    "atomic_write_json",
    "atomic_write_text",
]
