#!/usr/bin/env python3
"""Singleton configuration cache for Alfred hooks

Provides efficient caching of frequently accessed configuration data
with automatic invalidation based on file modification time.

Features:
- Singleton pattern for global cache state
- File mtime-based cache invalidation
- Type-safe cache operations
- Graceful degradation on read errors
"""

import json
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any, Optional


class ConfigCache:
    """Singleton cache for configuration data.

    Stores commonly accessed configuration to avoid repeated file I/O.
    Automatically invalidates cached data if source file is modified.

    Usage:
        cache = ConfigCache()
        config = cache.get_config()
        spec_progress = cache.get_spec_progress()
    """

    _instance = None
    _cache = {}
    _mtimes = {}

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance

    @staticmethod
    def get_config() -> Optional[dict[str, Any]]:
        """Get cached config, or load from file if not cached.

        Returns:
            Configuration dict, or None if file doesn't exist
        """
        config_path = Path.cwd() / ".moai" / "config" / "config.json"

        # Check if cache is still valid
        if ConfigCache._is_cache_valid("config", config_path):
            return ConfigCache._cache.get("config")

        # Load from file
        try:
            if not config_path.exists():
                return None

            config = json.loads(config_path.read_text())
            ConfigCache._update_cache("config", config, config_path)
            return config

        except Exception:
            return None

    @staticmethod
    def get_spec_progress() -> dict[str, Any]:
        """Get cached SPEC progress, or compute if not cached.

        Returns:
            Dict with keys: completed, total, percentage
        """
        specs_dir = Path.cwd() / ".moai" / "specs"

        # Check if cache is still valid (5 minute TTL)
        if ConfigCache._is_cache_valid("spec_progress", specs_dir, ttl_minutes=5):
            return ConfigCache._cache.get("spec_progress", {"completed": 0, "total": 0, "percentage": 0})

        # Compute from filesystem
        try:
            if not specs_dir.exists():
                result = {"completed": 0, "total": 0, "percentage": 0}
                ConfigCache._update_cache("spec_progress", result, specs_dir)
                return result

            spec_folders = [d for d in specs_dir.iterdir() if d.is_dir() and d.name.startswith("SPEC-")]
            total = len(spec_folders)

            # Simple completion check - look for spec.md files
            completed = sum(1 for folder in spec_folders if (folder / "spec.md").exists())

            percentage = (completed / total * 100) if total > 0 else 0

            result = {
                "completed": completed,
                "total": total,
                "percentage": round(percentage, 0)
            }

            ConfigCache._update_cache("spec_progress", result, specs_dir)
            return result

        except Exception:
            return {"completed": 0, "total": 0, "percentage": 0}

    @staticmethod
    def _is_cache_valid(key: str, file_path: Path, ttl_minutes: int = 30) -> bool:
        """Check if cached data is still valid.

        Args:
            key: Cache key
            file_path: Path to check for modifications
            ttl_minutes: Time-to-live in minutes

        Returns:
            True if cache exists and is still valid
        """
        if key not in ConfigCache._cache:
            return False

        if not file_path.exists():
            return False

        # Check file modification time
        try:
            current_mtime = file_path.stat().st_mtime
            cached_mtime = ConfigCache._mtimes.get(key)

            if cached_mtime is None:
                return False

            # If file was modified, cache is invalid
            if current_mtime != cached_mtime:
                return False

            # Check TTL
            cached_time = ConfigCache._cache.get(f"{key}_timestamp")
            if cached_time:
                elapsed = datetime.now() - cached_time
                if elapsed > timedelta(minutes=ttl_minutes):
                    return False

            return True

        except Exception:
            return False

    @staticmethod
    def _update_cache(key: str, data: Any, file_path: Path) -> None:
        """Update cache with new data.

        Args:
            key: Cache key
            data: Data to cache
            file_path: Path to track for modifications
        """
        try:
            ConfigCache._cache[key] = data
            ConfigCache._cache[f"{key}_timestamp"] = datetime.now()

            if file_path.exists():
                ConfigCache._mtimes[key] = file_path.stat().st_mtime

        except Exception:
            pass  # Silently fail on cache update

    @staticmethod
    def clear() -> None:
        """Clear all cached data.

        Useful for testing or forcing a refresh.
        """
        ConfigCache._cache.clear()
        ConfigCache._mtimes.clear()

    @staticmethod
    def get_cache_size() -> int:
        """Get current cache size in bytes.

        Returns:
            Approximate size of cached data in bytes
        """
        import sys
        size = sys.getsizeof(ConfigCache._cache) + sys.getsizeof(ConfigCache._mtimes)
        for key, value in ConfigCache._cache.items():
            size += sys.getsizeof(value)
        return size


# Convenience functions for singleton access
def get_cached_config() -> Optional[dict[str, Any]]:
    """Get cached configuration."""
    return ConfigCache.get_config()


def get_cached_spec_progress() -> dict[str, Any]:
    """Get cached SPEC progress."""
    return ConfigCache.get_spec_progress()


def clear_config_cache() -> None:
    """Clear all cached data."""
    ConfigCache.clear()
