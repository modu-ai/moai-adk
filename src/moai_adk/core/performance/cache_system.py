"""
Cache System

Provides persistent caching capabilities with TTL support for improved performance.
"""

import json
import os
import time
from typing import Any, Dict, List, Optional


class CacheSystem:
    """
    A persistent cache system with TTL support.

    This class provides file-based caching with support for time-to-live,
    multiple operations, persistence across instances, and thread safety.
    """

    def __init__(self, cache_dir: Optional[str] = None, auto_cleanup: bool = True):
        """
        Initialize the cache system.

        Args:
            cache_dir: Directory to store cache files. If None, uses default temp directory.
            auto_cleanup: Whether to automatically clean up expired files on operations
        """
        self.auto_cleanup = auto_cleanup

        if cache_dir is None:
            import tempfile

            self.cache_dir = os.path.join(tempfile.gettempdir(), "moai_adk_cache")
        else:
            self.cache_dir = cache_dir

        # Cache file extension
        self.file_extension = ".cache"

        # Create cache directory if it doesn't exist
        self._ensure_cache_dir()

    def _ensure_cache_dir(self) -> None:
        """Ensure cache directory exists."""
        try:
            os.makedirs(self.cache_dir, exist_ok=True)
        except OSError as e:
            raise OSError(f"Failed to create cache directory {self.cache_dir}: {e}")

    def _validate_key(self, key: str) -> str:
        """
        Validate and sanitize cache key.

        Args:
            key: Raw cache key

        Returns:
            Sanitized key suitable for filename
        """
        if not isinstance(key, str):
            raise TypeError("Cache key must be a string")

        if not key or key.isspace():
            raise ValueError("Cache key cannot be empty")

        # Sanitize key for safe filename usage
        safe_key = key.replace("/", "_").replace("\\", "_")
        if safe_key != key:
            return safe_key

        return key

    def _get_file_path(self, key: str) -> str:
        """Get file path for a given cache key."""
        safe_key = self._validate_key(key)
        return os.path.join(self.cache_dir, f"{safe_key}{self.file_extension}")

    def _is_expired(self, data: Dict[str, Any]) -> bool:
        """Check if cache data is expired."""
        if "expires_at" not in data:
            return False

        return time.time() > data["expires_at"]

    def _cleanup_expired_files(self) -> None:
        """Remove expired cache files."""
        time.time()
        for file_name in os.listdir(self.cache_dir):
            if file_name.endswith(self.file_extension):
                file_path = os.path.join(self.cache_dir, file_name)
                try:
                    with open(file_path, "r", encoding="utf-8") as f:
                        data = json.load(f)

                    if self._is_expired(data):
                        os.remove(file_path)
                except (json.JSONDecodeError, KeyError, OSError):
                    # Remove corrupted files too
                    try:
                        os.remove(file_path)
                    except OSError:
                        pass

    def _write_data(self, file_path: str, data: Dict[str, Any]) -> None:
        """Write data to file with error handling."""
        try:
            with open(file_path, "w", encoding="utf-8") as f:
                json.dump(data, f, indent=2, ensure_ascii=False)
        except (OSError, TypeError) as e:
            raise OSError(f"Failed to write cache file {file_path}: {e}")

    def _read_data(self, file_path: str) -> Optional[Dict[str, Any]]:
        """Read data from file with error handling."""
        if not os.path.exists(file_path):
            return None

        try:
            with open(file_path, "r", encoding="utf-8") as f:
                return json.load(f)
        except (json.JSONDecodeError, OSError):
            # File is corrupted, remove it
            try:
                os.remove(file_path)
            except OSError:
                pass
            return None

    def set(self, key: str, value: Any, ttl: Optional[float] = None) -> None:
        """
        Set a value in the cache.

        Args:
            key: Cache key
            value: Value to cache (must be JSON serializable)
            ttl: Time to live in seconds (optional)

        Raises:
            TypeError: If value is not JSON serializable
            OSError: If file operations fail
        """
        # Validate JSON serializability
        try:
            json.dumps(value)
        except (TypeError, ValueError) as e:
            raise TypeError(f"Cache value must be JSON serializable: {e}")

        data = {"value": value, "created_at": time.time()}

        if ttl is not None:
            if not isinstance(ttl, (int, float)) or ttl < 0:
                raise ValueError("TTL must be a positive number")
            data["expires_at"] = data["created_at"] + ttl

        file_path = self._get_file_path(key)
        self._write_data(file_path, data)

        # Auto-cleanup if enabled
        if self.auto_cleanup:
            self._cleanup_expired_files()

    def get(self, key: str) -> Optional[Any]:
        """
        Get a value from the cache.

        Args:
            key: Cache key

        Returns:
            Cached value or None if not found or expired
        """
        file_path = self._get_file_path(key)
        data = self._read_data(file_path)

        if data is None:
            return None

        # Check expiration
        if self._is_expired(data):
            try:
                os.remove(file_path)
            except OSError:
                pass
            return None

        return data["value"]

    def delete(self, key: str) -> bool:
        """
        Delete a value from the cache.

        Args:
            key: Cache key

        Returns:
            True if file was deleted, False if it didn't exist
        """
        file_path = self._get_file_path(key)
        try:
            os.remove(file_path)
            return True
        except OSError:
            return False

    def clear(self) -> int:
        """
        Clear all values from the cache.

        Returns:
            Number of files removed
        """
        count = 0
        for file_name in os.listdir(self.cache_dir):
            if file_name.endswith(self.file_extension):
                file_path = os.path.join(self.cache_dir, file_name)
                try:
                    os.remove(file_path)
                    count += 1
                except OSError:
                    continue
        return count

    def exists(self, key: str) -> bool:
        """
        Check if a key exists in the cache.

        Args:
            key: Cache key

        Returns:
            True if key exists and is not expired, False otherwise
        """
        return self.get(key) is not None

    def size(self) -> int:
        """
        Get the number of items in the cache.

        Returns:
            Number of non-expired cache items
        """
        count = 0
        for file_name in os.listdir(self.cache_dir):
            if file_name.endswith(self.file_extension):
                os.path.join(self.cache_dir, file_name)
                key = file_name[: -len(self.file_extension)]  # Remove extension

                if self.exists(key):
                    count += 1
        return count

    def set_if_not_exists(self, key: str, value: Any, ttl: Optional[float] = None) -> bool:
        """
        Set a value only if the key doesn't exist.

        Args:
            key: Cache key
            value: Value to cache
            ttl: Time to live in seconds (optional)

        Returns:
            True if value was set, False if key already existed
        """
        if self.exists(key):
            return False

        self.set(key, value, ttl)
        return True

    def get_multiple(self, keys: List[str]) -> Dict[str, Optional[Any]]:
        """
        Get multiple values from the cache.

        Args:
            keys: List of cache keys

        Returns:
            Dictionary mapping keys to values (or None)
        """
        if not isinstance(keys, list):
            raise TypeError("keys must be a list")

        result = {}
        for key in keys:
            if not isinstance(key, str):
                raise TypeError("All keys must be strings")
            result[key] = self.get(key)
        return result

    def get_stats(self) -> Dict[str, Any]:
        """
        Get cache statistics.

        Returns:
            Dictionary with cache statistics
        """
        total_files = 0
        expired_files = 0
        time.time()

        for file_name in os.listdir(self.cache_dir):
            if file_name.endswith(self.file_extension):
                total_files += 1
                file_path = os.path.join(self.cache_dir, file_name)
                data = self._read_data(file_path)

                if data and self._is_expired(data):
                    expired_files += 1

        return {
            "total_files": total_files,
            "expired_files": expired_files,
            "valid_files": total_files - expired_files,
            "cache_directory": self.cache_dir,
            "auto_cleanup_enabled": self.auto_cleanup,
        }
