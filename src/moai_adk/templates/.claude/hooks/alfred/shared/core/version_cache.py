#!/usr/bin/env python3
"""Version information cache with TTL support

TTL-based caching system for version check results to minimize network calls
during SessionStart hook execution.

SPEC: SPEC-UPDATE-ENHANCE-001 - SessionStart version check system enhancement
Phase 1: Cache System Implementation
"""

import json
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


class VersionCache:
    """TTL-based version information cache

    Caches version check results with configurable Time-To-Live (TTL)
    to avoid excessive network calls to PyPI during SessionStart events.

    Attributes:
        cache_dir: Directory to store cache file
        ttl_hours: Time-to-live in hours (default 24)
        cache_file: Path to the cache JSON file

    Examples:
        >>> cache = VersionCache(Path(".moai/cache"), ttl_hours=24)
        >>> cache.save({"current_version": "0.8.1", "latest_version": "0.9.0"})
        True
        >>> cache.is_valid()
        True
        >>> data = cache.load()
        >>> data["current_version"]
        '0.8.1'
    """

    def __init__(self, cache_dir: Path, ttl_hours: int = 4):
        """Initialize cache with TTL in hours

        Args:
            cache_dir: Directory where cache file will be stored
            ttl_hours: Time-to-live in hours (default 4)
        """
        self.cache_dir = Path(cache_dir)
        self.ttl_hours = ttl_hours
        self.cache_file = self.cache_dir / "version-check.json"

    def _calculate_age_hours(self, last_check_iso: str) -> float:
        """Calculate age in hours from ISO timestamp (internal helper)

        Normalizes timezone-aware and naive datetimes for consistent comparison.

        Args:
            last_check_iso: ISO format timestamp string

        Returns:
            Age in hours

        Raises:
            ValueError: If timestamp parsing fails
        """
        last_check = datetime.fromisoformat(last_check_iso)

        # Normalize to naive datetime (remove timezone for comparison)
        if last_check.tzinfo is not None:
            last_check = last_check.replace(tzinfo=None)

        now = datetime.now()
        return (now - last_check).total_seconds() / 3600

    def is_valid(self) -> bool:
        """Check if cache exists and is not expired

        Returns:
            True if cache file exists and is within TTL, False otherwise

        Examples:
            >>> cache = VersionCache(Path(".moai/cache"))
            >>> cache.is_valid()
            False  # No cache file exists yet
        """
        if not self.cache_file.exists():
            return False

        try:
            with open(self.cache_file, "r") as f:
                data = json.load(f)

            age_hours = self._calculate_age_hours(data["last_check"])
            return age_hours < self.ttl_hours

        except (json.JSONDecodeError, KeyError, ValueError, OSError):
            # Corrupted or invalid cache file
            return False

    def load(self) -> dict[str, Any] | None:
        """Load cached version info if valid

        Returns:
            Cached version info dictionary if valid, None otherwise

        Examples:
            >>> cache = VersionCache(Path(".moai/cache"))
            >>> data = cache.load()
            >>> data is None
            True  # No valid cache exists
        """
        if not self.is_valid():
            return None

        try:
            with open(self.cache_file, "r") as f:
                return json.load(f)
        except (json.JSONDecodeError, OSError):
            # Graceful degradation on read errors
            return None

    def save(self, version_info: dict[str, Any]) -> bool:
        """Save version info to cache file

        Creates cache directory if it doesn't exist.
        Updates last_check timestamp to current time if not provided.

        Args:
            version_info: Version information dictionary to cache

        Returns:
            True on successful save, False on error

        Examples:
            >>> cache = VersionCache(Path(".moai/cache"))
            >>> cache.save({"current_version": "0.8.1"})
            True
        """
        try:
            # Create cache directory if it doesn't exist
            self.cache_dir.mkdir(parents=True, exist_ok=True)

            # Update last_check timestamp only if not provided (for testing)
            if "last_check" not in version_info:
                version_info["last_check"] = datetime.now(timezone.utc).isoformat()

            # Write to cache file
            with open(self.cache_file, "w") as f:
                json.dump(version_info, f, indent=2)

            return True

        except (OSError, TypeError):
            # Graceful degradation on write errors
            return False

    def clear(self) -> bool:
        """Clear/remove cache file

        Returns:
            True if cache file was removed or didn't exist, False on error

        Examples:
            >>> cache = VersionCache(Path(".moai/cache"))
            >>> cache.clear()
            True
        """
        try:
            if self.cache_file.exists():
                self.cache_file.unlink()
            return True
        except OSError:
            return False

    def get_age_hours(self) -> float:
        """Get age of cache in hours

        Returns:
            Age in hours, or 0.0 if cache doesn't exist or is invalid

        Examples:
            >>> cache = VersionCache(Path(".moai/cache"))
            >>> cache.get_age_hours()
            0.0  # No cache exists
        """
        if not self.cache_file.exists():
            return 0.0

        try:
            with open(self.cache_file, "r") as f:
                data = json.load(f)

            return self._calculate_age_hours(data["last_check"])

        except (json.JSONDecodeError, KeyError, ValueError, OSError):
            return 0.0


__all__ = ["VersionCache"]
