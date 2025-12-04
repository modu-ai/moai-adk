# type: ignore
"""
Update checker for MoAI-ADK using PyPI API

"""

import json
import logging
import re
import urllib.request
from dataclasses import dataclass
from datetime import datetime, timedelta
from typing import Optional

logger = logging.getLogger(__name__)


@dataclass
class UpdateInfo:
    """Update information"""

    available: bool
    latest_version: Optional[str]


class UpdateChecker:
    """Checks for MoAI-ADK updates from PyPI with 300-second caching"""

    # Configuration
    _CACHE_TTL_SECONDS = 300
    _PYPI_API_URL = "https://pypi.org/pypi/moai-adk/json"
    _TIMEOUT_SECONDS = 5

    def __init__(self):
        """Initialize update checker"""
        self._cached_info: Optional[UpdateInfo] = None
        self._cache_time: Optional[datetime] = None
        self._cache_ttl = timedelta(seconds=self._CACHE_TTL_SECONDS)
        self._cached_version: Optional[str] = None

    def check_for_update(self, current_version: str) -> UpdateInfo:
        """
        Check for available updates

        Args:
            current_version: Current MoAI-ADK version (e.g., "0.20.1")

        Returns:
            UpdateInfo with availability and latest version
        """
        # Check cache validity (same version)
        if self._is_cache_valid() and self._cached_version == current_version:
            return self._cached_info

        # Fetch latest version from PyPI
        update_info = self._fetch_latest_version(current_version)
        self._update_cache_with(update_info, current_version)
        return update_info

    def _fetch_latest_version(self, current_version: str) -> UpdateInfo:
        """
        Fetch latest version from PyPI API

        Args:
            current_version: Current version string

        Returns:
            UpdateInfo from PyPI or error default
        """
        try:
            with urllib.request.urlopen(self._PYPI_API_URL, timeout=self._TIMEOUT_SECONDS) as response:
                data = json.loads(response.read().decode("utf-8"))

            latest_version = data.get("info", {}).get("version")

            if not latest_version:
                return UpdateInfo(available=False, latest_version=None)

            # Compare versions
            available = self._is_update_available(current_version, latest_version)

            return UpdateInfo(
                available=available,
                latest_version=latest_version if available else None,
            )

        except Exception as e:
            logger.debug(f"Error checking for updates: {e}")
            return UpdateInfo(available=False, latest_version=None)

    @staticmethod
    def _is_update_available(current: str, latest: str) -> bool:
        """
        Compare versions to determine if update is available

        Args:
            current: Current version string
            latest: Latest version string

        Returns:
            True if update is available
        """
        try:
            # Parse version strings (remove 'v' prefix)
            current_clean = current.lstrip("v")
            latest_clean = latest.lstrip("v")

            # Split by dots and convert to integers
            current_parts = [int(x) for x in re.split(r"[^\d]+", current_clean) if x.isdigit()]
            latest_parts = [int(x) for x in re.split(r"[^\d]+", latest_clean) if x.isdigit()]

            # Compare version tuples
            return tuple(latest_parts) > tuple(current_parts)

        except Exception as e:
            logger.debug(f"Error comparing versions: {e}")
            return False

    def _is_cache_valid(self) -> bool:
        """Check if update cache is still valid"""
        if self._cached_info is None or self._cache_time is None:
            return False
        return datetime.now() - self._cache_time < self._cache_ttl

    def _update_cache_with(self, update_info: UpdateInfo, version: str) -> None:
        """Update cache with update info"""
        self._cached_info = update_info
        self._cache_time = datetime.now()
        self._cached_version = version
