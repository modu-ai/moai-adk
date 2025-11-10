# type: ignore
"""
Enhanced Version reader for MoAI-ADK from config.json

@CODE:VERSION-READER-001 | @SPEC:STATUSLINE-VERSION-READER-001 | @TEST:STATUSLINE-VERSION-READER-001
Refactored for improved performance, error handling, and configurability
"""

import json
import logging
import re
import asyncio
from datetime import datetime, timedelta
from pathlib import Path
from typing import Optional, Dict, Any
from dataclasses import dataclass

logger = logging.getLogger(__name__)


@dataclass
class VersionConfig:
    """Configuration for version reading behavior"""
    cache_ttl_seconds: int = 60
    fallback_version: str = "unknown"
    version_format_regex: str = r"^v?(\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?)$"
    cache_enabled: bool = True
    debug_mode: bool = False


class VersionReader:
    """
    Enhanced version reader for MoAI-ADK with improved caching,
    error handling, and configurability.

    Features:
    - Configurable caching with TTL support
    - Multiple version field fallbacks
    - Async support for better performance
    - Comprehensive error handling
    - Version format validation
    - Graceful degradation strategies
    """

    # Default configuration
    DEFAULT_CONFIG = VersionConfig()

    # Supported version fields in order of priority
    VERSION_FIELDS = [
        "project.version",
        "moai.version",
        "version",
        "moai.template_version",
        "template_version"
    ]

    def __init__(self, config: Optional[VersionConfig] = None):
        """
        Initialize version reader with enhanced configuration.

        Args:
            config: Version configuration object. If None, uses defaults.
        """
        self.config = config or self.DEFAULT_CONFIG
        self._version_cache: Optional[str] = None
        self._cache_time: Optional[datetime] = None
        self._cache_ttl = timedelta(seconds=self.config.cache_ttl_seconds)
        self._config_path = Path.cwd() / ".moai" / "config.json"
        self._cache_stats = {
            'hits': 0,
            'misses': 0,
            'errors': 0
        }

        # Pre-compile regex for performance
        try:
            self._version_pattern = re.compile(self.config.version_format_regex)
        except re.error:
            # Fallback to default regex if custom one is invalid
            self._version_pattern = re.compile(self.DEFAULT_CONFIG.version_format_regex)

    def get_version(self) -> str:
        """
        Get MoAI-ADK version from config with caching.

        Returns:
            Version string (e.g., "0.20.1" or "v0.20.1")

        Raises:
            VersionReadError: If version cannot be determined after fallbacks
        """
        return asyncio.run(self.get_version_async())

    async def get_version_async(self) -> str:
        """
        Async version getter for better performance.

        Returns:
            Version string
        """
        if self.config.cache_enabled and self._is_cache_valid():
            self._cache_stats['hits'] += 1
            if self.config.debug_mode:
                logger.debug(f"Version cache hit: {self._version_cache}")
            return self._version_cache

        # Read from config file
        try:
            version = await self._read_version_from_config_async()
            if version and self._is_valid_version_format(version):
                self._update_cache(version)
                self._cache_stats['misses'] += 1
                return version
            else:
                fallback = self._get_fallback_version()
                self._update_cache(fallback)
                self._cache_stats['misses'] += 1
                return fallback
        except Exception as e:
            self._cache_stats['errors'] += 1
            logger.error(f"Error reading version: {e}")
            fallback = self._get_fallback_version()
            self._update_cache(fallback)
            return fallback

    async def _read_version_from_config_async(self) -> str:
        """
        Read version from .moai/config.json asynchronously.

        Returns:
            Version string or empty string if not found
        """
        try:
            if not await self._file_exists_async(self._config_path):
                logger.debug(f"Config file not found: {self._config_path}")
                return ""

            try:
                config_data = await self._read_json_async(self._config_path)
                return self._extract_version_from_config(config_data)
            except json.JSONDecodeError as e:
                logger.error(f"Invalid JSON in config {self._config_path}: {e}")
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
        keys = field_path.split('.')
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
        return version[1:] if version.startswith('v') else version

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
        elif version.startswith('v'):
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
        fallback = self.config.fallback_version
        logger.debug(f"Using fallback version: {fallback}")
        return fallback

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
        with open(path, "r", encoding="utf-8") as f:
            return json.load(f)

    def _is_cache_valid(self) -> bool:
        """
        Check if version cache is still valid.

        Returns:
            True if cache is valid and not expired
        """
        if not self.config.cache_enabled:
            return False

        if self._version_cache is None or self._cache_time is None:
            return False

        return datetime.now() - self._cache_time < self._cache_ttl

    def _update_cache(self, version: str) -> None:
        """
        Update version cache with new version.

        Args:
            version: Version string to cache
        """
        if self.config.cache_enabled:
            self._version_cache = version
            self._cache_time = datetime.now()
            logger.debug(f"Cache updated with version: {version}")

    def clear_cache(self) -> None:
        """Clear version cache"""
        self._version_cache = None
        self._cache_time = None
        logger.debug("Version cache cleared")

    def get_cache_stats(self) -> Dict[str, int]:
        """
        Get cache statistics.

        Returns:
            Dictionary with cache hit/miss/error counts
        """
        return self._cache_stats.copy()

    def get_cache_age_seconds(self) -> Optional[float]:
        """
        Get cache age in seconds.

        Returns:
            Cache age in seconds, or None if no cached version
        """
        if self._cache_time is None:
            return None
        return (datetime.now() - self._cache_time).total_seconds()

    def is_cache_expired(self) -> bool:
        """
        Check if cache is expired.

        Returns:
            True if cache is expired
        """
        return not self._is_cache_valid()

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
        logger.debug(f"Custom version fields set: {fields}")


class VersionReadError(Exception):
    """Exception raised when version cannot be read"""
    pass
