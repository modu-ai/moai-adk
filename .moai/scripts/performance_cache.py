#!/usr/bin/env python3
"""
Performance Cache Module for check-traceability.py
Provides file content caching and modification time tracking to avoid redundant reads.

@FEATURE:PERFORMANCE-CACHE-001 → @TASK:FILE-CACHE-001
"""

import json
import time
import threading
from pathlib import Path
from typing import Dict, Optional, Set
from datetime import datetime


class PerformanceCache:
    """
    File content cache with modification time tracking.
    Thread-safe implementation for concurrent access.
    """

    def __init__(self, cache_dir: Optional[Path] = None):
        self.cache_dir = cache_dir or Path(".moai/cache")
        self.cache_file = self.cache_dir / "file_cache.json"
        self.last_scan_file = self.cache_dir / "last_scan.json"

        # In-memory cache for current session
        self._memory_cache: Dict[str, dict] = {}
        self._lock = threading.RLock()

        # Performance metrics
        self.metrics = {
            'cache_hits': 0,
            'cache_misses': 0,
            'files_scanned': 0,
            'scan_duration': 0.0
        }

        self._ensure_cache_dir()
        self._load_persistent_cache()

    def _ensure_cache_dir(self) -> None:
        """Ensure cache directory exists"""
        self.cache_dir.mkdir(parents=True, exist_ok=True)

    def _load_persistent_cache(self) -> None:
        """Load persistent cache from disk"""
        if self.cache_file.exists():
            try:
                with open(self.cache_file, 'r', encoding='utf-8') as f:
                    data = json.load(f)
                    self._memory_cache = data.get('cache', {})
            except Exception:
                self._memory_cache = {}

    def _save_persistent_cache(self) -> None:
        """Save cache to disk"""
        try:
            cache_data = {
                'cache': self._memory_cache,
                'last_updated': datetime.now().isoformat(),
                'version': '1.0'
            }
            with open(self.cache_file, 'w', encoding='utf-8') as f:
                json.dump(cache_data, f, indent=2)
        except Exception:
            pass  # Fail silently for cache saves

    def get_cached_content(self, file_path: Path) -> Optional[str]:
        """
        Get cached file content if still valid.
        Returns None if file is not cached or has been modified.
        """
        with self._lock:
            file_str = str(file_path)

            if file_str not in self._memory_cache:
                self.metrics['cache_misses'] += 1
                return None

            cached_entry = self._memory_cache[file_str]

            try:
                # Check if file still exists and modification time matches
                if not file_path.exists():
                    # File was deleted, remove from cache
                    del self._memory_cache[file_str]
                    self.metrics['cache_misses'] += 1
                    return None

                current_mtime = file_path.stat().st_mtime
                cached_mtime = cached_entry.get('mtime', 0)

                if current_mtime != cached_mtime:
                    # File was modified, invalidate cache entry
                    del self._memory_cache[file_str]
                    self.metrics['cache_misses'] += 1
                    return None

                # Cache hit
                self.metrics['cache_hits'] += 1
                return cached_entry['content']

            except OSError:
                # File access error, remove from cache
                if file_str in self._memory_cache:
                    del self._memory_cache[file_str]
                self.metrics['cache_misses'] += 1
                return None

    def cache_file_content(self, file_path: Path, content: str) -> None:
        """Cache file content with current modification time"""
        with self._lock:
            try:
                mtime = file_path.stat().st_mtime
                self._memory_cache[str(file_path)] = {
                    'content': content,
                    'mtime': mtime,
                    'cached_at': time.time()
                }
            except OSError:
                pass  # Can't cache if we can't get file stats

    def get_changed_files_since_last_scan(self, files: list) -> Set[str]:
        """
        Return set of files that have changed since last scan.
        If no last scan data exists, returns all files.

        Enhanced with tolerance for small time differences and better error handling.
        """
        try:
            if not self.last_scan_file.exists():
                # First time - consider all files as new
                return set(str(f) for f in files)

            with open(self.last_scan_file, 'r', encoding='utf-8') as f:
                last_scan_data = json.load(f)
                last_scan_mtimes = last_scan_data.get('file_mtimes', {})

            if not last_scan_mtimes:
                # No previous scan data - scan all files
                return set(str(f) for f in files)

            changed_files = set()

            for file_path in files:
                file_str = str(file_path)
                try:
                    current_mtime = Path(file_path).stat().st_mtime
                    last_mtime = last_scan_mtimes.get(file_str, 0)

                    # Use tolerance for floating point precision issues
                    time_diff = abs(current_mtime - last_mtime)
                    if time_diff > 0.001:  # 1ms tolerance
                        changed_files.add(file_str)

                except OSError:
                    # File access error, consider it changed
                    changed_files.add(file_str)

            return changed_files

        except Exception as e:
            # Log the error for debugging
            import logging
            logging.warning(f"Error checking changed files: {e}")
            # If any error occurs, scan all files for safety
            return set(str(f) for f in files)

    def save_scan_timestamp(self, scanned_files: list) -> None:
        """Save timestamp and file mtimes of current scan"""
        try:
            file_mtimes = {}
            for file_path in scanned_files:
                try:
                    file_mtimes[str(file_path)] = Path(file_path).stat().st_mtime
                except OSError:
                    pass

            scan_data = {
                'timestamp': time.time(),
                'file_mtimes': file_mtimes,
                'scan_id': datetime.now().isoformat()
            }

            with open(self.last_scan_file, 'w', encoding='utf-8') as f:
                json.dump(scan_data, f, indent=2)

        except Exception:
            pass  # Fail silently

    def get_cache_stats(self) -> dict:
        """Get cache performance statistics"""
        with self._lock:
            total_requests = self.metrics['cache_hits'] + self.metrics['cache_misses']
            hit_rate = (self.metrics['cache_hits'] / total_requests * 100) if total_requests > 0 else 0

            return {
                'cache_hits': self.metrics['cache_hits'],
                'cache_misses': self.metrics['cache_misses'],
                'hit_rate_percent': round(hit_rate, 1),
                'cached_files': len(self._memory_cache),
                'files_scanned': self.metrics['files_scanned'],
                'scan_duration': self.metrics['scan_duration']
            }

    def clear_cache(self) -> None:
        """Clear all cached data"""
        with self._lock:
            self._memory_cache.clear()
            self.metrics = {
                'cache_hits': 0,
                'cache_misses': 0,
                'files_scanned': 0,
                'scan_duration': 0.0
            }

    def save_to_disk(self) -> None:
        """Explicitly save cache to disk"""
        self._save_persistent_cache()