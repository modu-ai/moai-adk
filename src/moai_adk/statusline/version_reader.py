"""
Version reader for MoAI-ADK from config.json

@CODE:VERSION-READER-001
"""

import json
import logging
from pathlib import Path
from typing import Optional
from datetime import datetime, timedelta

logger = logging.getLogger(__name__)


class VersionReader:
    """Reads MoAI-ADK version from config with 60-second caching"""

    # Configuration
    _CACHE_TTL_SECONDS = 60

    def __init__(self):
        """Initialize version reader"""
        self._version_cache: Optional[str] = None
        self._cache_time: Optional[datetime] = None
        self._cache_ttl = timedelta(seconds=self._CACHE_TTL_SECONDS)
        self._config_path = Path.cwd() / ".moai" / "config.json"

    def get_version(self) -> str:
        """
        Get MoAI-ADK version from config

        Returns:
            Version string (e.g., "0.20.1" or "v0.20.1")
        """
        # Check cache
        if self._is_cache_valid():
            return self._version_cache

        # Read from config file
        version = self._read_version_from_config()
        self._update_cache(version)
        return version

    def _read_version_from_config(self) -> str:
        """
        Read version from .moai/config.json

        Returns:
            Version string or default if not found
        """
        try:
            if not self._config_path.exists():
                logger.debug(f"Config file not found: {self._config_path}")
                return "unknown"

            with open(self._config_path, "r") as f:
                config = json.load(f)

            # Try to get version from moai.version or version field
            version = (
                config.get("moai", {}).get("version")
                or config.get("version")
            )

            if version:
                return version

            logger.debug("Version field not found in config")
            return "unknown"

        except Exception as e:
            logger.debug(f"Error reading version from config: {e}")
            return "unknown"

    def _is_cache_valid(self) -> bool:
        """Check if version cache is still valid"""
        if self._version_cache is None or self._cache_time is None:
            return False
        return datetime.now() - self._cache_time < self._cache_ttl

    def _update_cache(self, version: str) -> None:
        """Update version cache"""
        self._version_cache = version
        self._cache_time = datetime.now()
