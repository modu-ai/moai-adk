"""
Enhanced Version reader for MoAI-ADK from config.yaml with performance optimizations

Refactored for improved performance, error handling, configurability, and caching strategies
"""

import asyncio
import json
import logging
import os
import re
import time
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
from pathlib import Path
from typing import Any, Dict, List, Optional

try:
    import yaml
except ImportError as e:
    raise ImportError(
        "PyYAML is required for MoAI-ADK hooks. "
        "Install with: pip install pyyaml\n"
        f"Or use: uv run --with pyyaml <hook_script>\n"
        f"Original error: {e}"
    ) from e

logger = logging.getLogger(__name__)


class VersionSource(Enum):
    """Enum for version source tracking"""

    CONFIG_FILE = "config_file"
    FALLBACK = "fallback"
    PACKAGE = "package"
    CACHE = "cache"


@dataclass
class CacheEntry:
    """Cache entry with metadata"""

    version: str
    timestamp: datetime
    source: VersionSource
    access_count: int = 0
    last_access: datetime = field(default_factory=datetime.now)


@dataclass
class VersionConfig:
    """Configuration for version reading behavior with enhanced options"""

    # Cache configuration
    cache_ttl_seconds: int = 60
    cache_enabled: bool = True
    cache_size: int = 50  # Maximum number of cached entries
    enable_lru_cache: bool = True  # Enable least recently used cache eviction

    # Fallback configuration
    fallback_version: str = "unknown"
    fallback_source: VersionSource = VersionSource.FALLBACK

    # Validation configuration
    version_format_regex: str = r"^v?(\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?)$"
    enable_validation: bool = True
    strict_validation: bool = False

    # Performance configuration
    enable_async: bool = True
    enable_batch_reading: bool = True
    batch_size: int = 10
    timeout_seconds: int = 5

    # Debug configuration
    debug_mode: bool = False
    enable_detailed_logging: bool = False
    track_performance_metrics: bool = True

    # Version field priority configuration
    # Order: 1) MoAI package version, 2) Project version, 3) Fallbacks
    version_fields: List[str] = field(
        default_factory=lambda: [
            "moai.version",  # ← 1st priority: MoAI-ADK version
            "project.version",  # ← 2nd priority: Project version
            "version",  # ← 3rd priority: General version
            "project.template_version",  # ← 4th priority: Template version
            "template_version",
        ]
    )


class VersionReader:
    """
    Enhanced version reader for MoAI-ADK with advanced caching,
    performance optimization, and comprehensive error handling.

    Features:
    - Multi-level caching with LRU eviction strategy
    - Configurable version field priority
    - Async batch processing for better performance
    - Comprehensive error handling and recovery
    - Version format validation with customizable patterns
    - Performance metrics tracking
    - Graceful degradation strategies
    - Source tracking for debugging
    """

    # Default configuration
    DEFAULT_CONFIG = VersionConfig()

    # Supported version fields in order of priority
    # Order: 1) MoAI package version, 2) Project version, 3) Fallbacks
    DEFAULT_VERSION_FIELDS = [
        "moai.version",  # ← 1st priority: MoAI-ADK version
        "project.version",  # ← 2nd priority: Project version
        "version",  # ← 3rd priority: General version
        "project.template_version",  # ← 4th priority: Template version
        "template_version",
    ]

    def __init__(self, config: Optional[VersionConfig] = None, working_dir: Optional[Path] = None):
        """
        Initialize version reader with enhanced configuration.

        Args:
            config: Version configuration object. If None, uses defaults.
            working_dir: Working directory to search for config. If None, uses environment detection.
        """
        self.config = config or self.DEFAULT_CONFIG

        # Determine working directory with priority:
        # 1. Explicit working_dir parameter
        # 2. CLAUDE_PROJECT_DIR environment variable (set by Claude Code)
        # 3. Current working directory
        if working_dir:
            base_dir = Path(working_dir)
        elif "CLAUDE_PROJECT_DIR" in os.environ:
            base_dir = Path(os.environ["CLAUDE_PROJECT_DIR"])
        else:
            base_dir = Path.cwd()

        self._config_path = base_dir / ".moai" / "config" / "config.yaml"

        # Enhanced caching with LRU support
        self._cache: Dict[str, CacheEntry] = {}
        self._cache_stats: Dict[str, Any] = {
            "hits": 0,
            "misses": 0,
            "errors": 0,
            "cache_hits_by_source": {
                VersionSource.CONFIG_FILE.value: 0,
                VersionSource.CACHE.value: 0,
                VersionSource.FALLBACK.value: 0,
            },
        }

        # Performance tracking
        self._performance_metrics: Dict[str, List[float]] = {
            "read_times": [],
            "validation_times": [],
            "cache_operation_times": [],
        }

        # Version field configuration (backwards compatibility)
        self._version_fields = self.config.version_fields.copy()
        self.VERSION_FIELDS = self._version_fields.copy()

        # Pre-compile regex for performance
        try:
            self._version_pattern = re.compile(self.config.version_format_regex)
        except re.error:
            self._version_pattern = re.compile(self.DEFAULT_CONFIG.version_format_regex)

        # Logging
        self._logger = logging.getLogger(__name__)

        # Backwards compatibility cache attributes
        self._version_cache: Optional[str] = None
        self._cache_time: Optional[datetime] = None
        self._cache_ttl = timedelta(seconds=self.config.cache_ttl_seconds)

        if self.config.debug_mode:
            self._logger.info(f"VersionReader initialized with config: {self.config}")

    def get_version(self, force_refresh: bool = False) -> str:
        """
        Get MoAI-ADK version from config with enhanced caching.

        Args:
            force_refresh: If True, bypass cache and force fresh read

        Returns:
            Version string (e.g., "0.20.1" or "v0.20.1")

        Raises:
            VersionReadError: If version cannot be determined after fallbacks
        """
        # Check if we need to force refresh
        if force_refresh:
            self.clear_cache()

        if self.config.enable_async:
            return asyncio.run(self.get_version_async())
        else:
            return self.get_version_sync()

    def get_version_sync(self) -> str:
        """
        Synchronous version getter for performance-critical paths.

        Priority: installed package version > config file version > fallback

        Returns:
            Version string
        """
        start_time = time.time()

        try:
            # Check cache first
            version = self._check_cache()
            if version is not None:
                self._cache_stats["hits"] += 1
                self._cache_stats["cache_hits_by_source"][VersionSource.CACHE.value] += 1
                return version

            # Priority 1: Try installed package version first (most accurate)
            version = self._get_package_version()
            if version:
                self._update_cache(version, VersionSource.CONFIG_FILE)
                self._cache_stats["misses"] += 1
                return version

            # Priority 2: Read from config file
            version = self._read_version_from_config_sync()
            if not version:
                version = self._get_fallback_version()
            self._update_cache(version, VersionSource.CONFIG_FILE)
            self._cache_stats["misses"] += 1
            return version

        except Exception as e:
            self._handle_read_error(e, start_time)
            return self._get_fallback_version()

        finally:
            self._log_performance(start_time, "sync_read")

    async def get_version_async(self) -> str:
        """
        Async version getter for better performance.

        Priority: installed package version > config file version > fallback

        Returns:
            Version string
        """
        start_time = time.time()

        try:
            # Check cache first
            version = self._check_cache()
            if version is not None:
                self._cache_stats["hits"] += 1
                self._cache_stats["cache_hits_by_source"][VersionSource.CACHE.value] += 1
                return version

            # Priority 1: Try installed package version first (most accurate)
            version = self._get_package_version()
            if version:
                self._update_cache(version, VersionSource.CONFIG_FILE)
                self._cache_stats["misses"] += 1
                return version

            # Priority 2: Read from config file asynchronously
            version = await self._read_version_from_config_async()
            if not version:
                version = self._get_fallback_version()
            self._update_cache(version, VersionSource.CONFIG_FILE)
            self._cache_stats["misses"] += 1
            return version

        except Exception as e:
            self._handle_read_error(e, start_time)
            return self._get_fallback_version()

        finally:
            self._log_performance(start_time, "async_read")

    # Enhanced internal methods
    def _check_cache(self) -> Optional[str]:
        """
        Check cache for valid version entry.

        Returns:
            Version string if cache is valid, None otherwise
        """
        if not self.config.cache_enabled:
            return None

        # Check for existing cache entries
        config_key = str(self._config_path)
        if config_key in self._cache:
            entry = self._cache[config_key]

            # Check if cache entry is still valid
            if self._is_cache_entry_valid(entry):
                entry.access_count += 1
                entry.last_access = datetime.now()
                self._cache_stats["hits"] += 1
                self._cache_stats["cache_hits_by_source"][VersionSource.CACHE.value] += 1

                if self.config.debug_mode:
                    self._logger.debug(f"Cache hit: {entry.version} (source: {entry.source.value})")

                return entry.version

        return None

    def _is_cache_entry_valid(self, entry: CacheEntry) -> bool:
        """
        Check if cache entry is still valid.

        Args:
            entry: Cache entry to validate

        Returns:
            True if cache entry is valid
        """
        # Check TTL
        if self.config.cache_enabled:
            age = datetime.now() - entry.timestamp
            if age.total_seconds() > self.config.cache_ttl_seconds:
                return False

        return True

    def _update_cache(self, version: str, source: VersionSource) -> None:
        """
        Update cache with new version entry.

        Args:
            version: Version string to cache
            source: Source of the version
        """
        if not self.config.cache_enabled:
            return

        config_key = str(self._config_path)
        entry = CacheEntry(version=version, timestamp=datetime.now(), source=source)

        self._cache[config_key] = entry

        # Apply cache size limits with LRU eviction
        if len(self._cache) > self.config.cache_size:
            self._evict_oldest_cache_entry()

        if self.config.debug_mode:
            self._logger.debug(f"Cache updated with version: {version} (source: {source.value})")

    def _evict_oldest_cache_entry(self) -> None:
        """
        Evict the least recently used cache entry.
        """
        if not self.config.enable_lru_cache or len(self._cache) <= 1:
            return

        oldest_entry = None
        oldest_key = None

        for key, entry in self._cache.items():
            if oldest_entry is None or entry.last_access < oldest_entry.last_access:
                oldest_entry = entry
                oldest_key = key

        if oldest_key is not None:
            del self._cache[oldest_key]
            if self.config.debug_mode:
                self._logger.debug(f"Evicted oldest cache entry: {oldest_key}")

    def _handle_read_error(self, error: Exception, start_time: float) -> None:
        """
        Handle read errors with enhanced logging and recovery.

        Args:
            error: Exception that occurred
            start_time: When the operation started
        """
        self._cache_stats["errors"] += 1

        error_msg = f"Error reading version: {error}"
        if self.config.debug_mode:
            self._logger.error(error_msg, exc_info=True)
        else:
            self._logger.warning(error_msg)

        self._log_performance(start_time, "error_read")

    def _log_performance(self, start_time: float, operation: str) -> None:
        """
        Log performance metrics for the operation.

        Args:
            start_time: When the operation started
            operation: Type of operation being logged
        """
        if not self.config.track_performance_metrics:
            return

        duration = time.time() - start_time
        metric_name = f"{operation}_duration"

        if metric_name not in self._performance_metrics:
            self._performance_metrics[metric_name] = []

        self._performance_metrics[metric_name].append(duration)

        if self.config.debug_mode:
            self._logger.debug(f"Performance {operation}: {duration:.4f}s")

    def get_performance_metrics(self) -> Dict[str, Any]:
        """
        Get performance metrics for analysis.

        Returns:
            Dictionary containing performance metrics
        """
        metrics: Dict[str, Any] = {
            "cache_stats": self._cache_stats.copy(),
            "cache_size": len(self._cache),
            "max_cache_size": self.config.cache_size,
            "performance_metrics": {},
        }

        # Calculate average times for each operation
        for operation, times in self._performance_metrics.items():
            if times:
                metrics["performance_metrics"][operation] = {
                    "count": len(times),
                    "average": sum(times) / len(times),
                    "min": min(times),
                    "max": max(times),
                    "total": sum(times),
                }

        return metrics

    def _read_version_from_config_sync(self) -> str:
        """
        Synchronous version of reading from .moai/config/config.yaml.

        Returns:
            Version string or empty string if not found
        """
        try:
            if not self._config_path.exists():
                logger.debug(f"Config file not found: {self._config_path}")
                return ""

            try:
                config_data = self._read_config_sync(self._config_path)
                version = self._extract_version_from_config(config_data)
                return version if version else ""
            except (json.JSONDecodeError, yaml.YAMLError) as e:
                logger.error(f"Invalid config in {self._config_path}: {e}")
                return ""

        except Exception as e:
            logger.error(f"Error reading version from config: {e}")
            return ""

    async def _read_version_from_config_async(self) -> str:
        """
        Read version from .moai/config/config.yaml asynchronously.

        Returns:
            Version string or empty string if not found
        """
        try:
            if not await self._file_exists_async(self._config_path):
                logger.debug(f"Config file not found: {self._config_path}")
                return ""

            try:
                config_data = await self._read_config_async(self._config_path)
                return self._extract_version_from_config(config_data)
            except (json.JSONDecodeError, yaml.YAMLError) as e:
                logger.error(f"Invalid config in {self._config_path}: {e}")
                return ""

        except Exception as e:
            logger.error(f"Error reading version from config: {e}")
            return ""

    def _extract_version_from_config(self, config: Dict[str, Any]) -> str:
        """
        Extract version from config using multiple fallback strategies.

        Args:
            config: Configuration dictionary

        Returns:
            Version string or empty string
        """
        # Try each version field in order of priority
        for field_path in self.VERSION_FIELDS:
            version = self._get_nested_value(config, field_path)
            if version:
                logger.debug(f"Found version in field '{field_path}': {version}")
                return version

        logger.debug("No version field found in config")
        return ""

    def _get_nested_value(self, config: Dict[str, Any], field_path: str) -> Optional[str]:
        """
        Get nested value from config using dot notation.

        Args:
            config: Configuration dictionary
            field_path: Dot-separated path (e.g., "moai.version")

        Returns:
            Value or None if not found
        """
        keys = field_path.split(".")
        current = config

        for key in keys:
            if isinstance(current, dict) and key in current:
                current = current[key]
            else:
                return None

        return str(current) if current is not None else None

    def _format_short_version(self, version: str) -> str:
        """
        Format short version by removing 'v' prefix if present.

        Args:
            version: Version string

        Returns:
            Short version string
        """
        return version[1:] if version.startswith("v") else version

    def _format_display_version(self, version: str) -> str:
        """
        Format display version with proper formatting.

        Args:
            version: Version string

        Returns:
            Display version string
        """
        if version == "unknown":
            return "MoAI-ADK unknown version"
        elif version.startswith("v"):
            return f"MoAI-ADK {version}"
        else:
            return f"MoAI-ADK v{version}"

    def _is_valid_version_format(self, version: str) -> bool:
        """
        Validate version format using regex pattern.

        Args:
            version: Version string to validate

        Returns:
            True if version format is valid
        """
        return bool(self._version_pattern.match(version))

    def _get_fallback_version(self) -> str:
        """
        Get fallback version with graceful degradation.

        Returns:
            Fallback version string
        """
        # Try to get version from package metadata first
        pkg_version = self._get_package_version()
        if pkg_version:
            self._logger.debug(f"Using package metadata version: {pkg_version}")
            return pkg_version

        # Fall back to configured fallback version
        fallback = self.config.fallback_version
        self._logger.debug(f"Using configured fallback version: {fallback}")
        return fallback

    def _get_package_version(self) -> str:
        """
        Get version from installed moai-adk package metadata.

        This allows the statusline to work even when .moai/config/config.json
        is not found, as long as the moai-adk package is installed.

        Returns:
            Version string or empty string if package not found
        """
        try:
            from importlib.metadata import PackageNotFoundError, version

            try:
                pkg_version = version("moai-adk")
                self._logger.debug(f"Found moai-adk package version: {pkg_version}")
                return pkg_version
            except PackageNotFoundError:
                self._logger.debug("moai-adk package not found in metadata")
                return ""
        except ImportError:
            self._logger.debug("importlib.metadata not available")
            return ""
        except Exception as e:
            self._logger.debug(f"Error getting package version: {e}")
            return ""

    async def _file_exists_async(self, path: Path) -> bool:
        """Async file existence check"""
        try:
            loop = asyncio.get_event_loop()
            return await loop.run_in_executor(None, path.exists)
        except Exception:
            return False

    async def _read_json_async(self, path: Path) -> Dict[str, Any]:
        """Async JSON file reading"""
        loop = asyncio.get_event_loop()
        return await loop.run_in_executor(None, self._read_json_sync, path)

    def _read_json_sync(self, path: Path) -> Dict[str, Any]:
        """Synchronous JSON file reading"""
        with open(path, "r", encoding="utf-8", errors="replace") as f:
            return json.load(f)

    async def _read_config_async(self, path: Path) -> Dict[str, Any]:
        """Async config file reading (supports YAML and JSON)"""
        loop = asyncio.get_event_loop()
        return await loop.run_in_executor(None, self._read_config_sync, path)

    def _read_config_sync(self, path: Path) -> Dict[str, Any]:
        """Synchronous config file reading (supports YAML and JSON)"""
        with open(path, "r", encoding="utf-8", errors="replace") as f:
            if path.suffix in (".yaml", ".yml"):
                return yaml.safe_load(f) or {}
            else:
                return json.load(f)

    # Backwards compatibility cache methods
    def clear_cache(self) -> None:
        """Clear version cache (backwards compatibility)"""
        self._version_cache = None
        self._cache_time = None
        # Also clear the main cache dictionary
        self._cache.clear()
        # Reset cache statistics
        self._cache_stats = {
            "hits": 0,
            "misses": 0,
            "errors": 0,
            "cache_hits_by_source": {
                VersionSource.CONFIG_FILE.value: 0,
                VersionSource.CACHE.value: 0,
                VersionSource.FALLBACK.value: 0,
            },
        }
        logger.debug("Version cache cleared")

    def get_cache_stats(self) -> Dict[str, int]:
        """
        Get cache statistics (backwards compatibility).

        Returns:
            Dictionary with cache hit/miss/error counts
        """
        return self._cache_stats.copy()

    def get_cache_age_seconds(self) -> Optional[float]:
        """
        Get cache age in seconds (backwards compatibility).

        Returns:
            Cache age in seconds, or None if no cached version
        """
        if self._cache_time is None:
            return None
        return (datetime.now() - self._cache_time).total_seconds()

    def is_cache_expired(self) -> bool:
        """
        Check if cache is expired (backwards compatibility).

        Returns:
            True if cache is expired
        """
        # Check if cache entry exists and is still valid
        config_key = str(self._config_path)
        if config_key not in self._cache:
            return True
        entry = self._cache[config_key]
        return not self._is_cache_entry_valid(entry)

    def get_config(self) -> VersionConfig:
        """Get current configuration"""
        return self.config

    def update_config(self, config: VersionConfig) -> None:
        """
        Update configuration.

        Args:
            config: New configuration object
        """
        self.config = config
        self._cache_ttl = timedelta(seconds=self.config.cache_ttl_seconds)
        logger.debug("Version reader configuration updated")

    def get_available_version_fields(self) -> list[str]:
        """
        Get list of available version field paths.

        Returns:
            List of version field paths
        """
        return self.VERSION_FIELDS.copy()

    def set_custom_version_fields(self, fields: list[str]) -> None:
        """
        Set custom version field paths.

        Args:
            fields: List of version field paths in order of priority
        """
        self.VERSION_FIELDS = fields.copy()
        self._version_fields = fields.copy()  # Also update internal field list
        logger.debug(f"Custom version fields set: {fields}")


class VersionReadError(Exception):
    """Exception raised when version cannot be read"""

    pass
