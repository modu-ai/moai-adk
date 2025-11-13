# type: ignore
"""
Session metrics tracker for statusline

"""

import logging
from datetime import datetime, timedelta
from typing import Optional

logger = logging.getLogger(__name__)


class MetricsTracker:
    """Tracks session duration with 10-second caching"""

    # Configuration
    _CACHE_TTL_SECONDS = 10

    def __init__(self):
        """Initialize metrics tracker"""
        self._session_start: datetime = datetime.now()
        self._duration_cache: Optional[str] = None
        self._cache_time: Optional[datetime] = None
        self._cache_ttl = timedelta(seconds=self._CACHE_TTL_SECONDS)

    def get_duration(self) -> str:
        """
        Get formatted session duration

        Returns:
            Formatted duration string (e.g., "5m", "1h 30m")
        """
        # Check cache validity
        if self._is_cache_valid():
            return self._duration_cache

        # Calculate and format duration
        duration = self._calculate_and_format_duration()
        self._update_cache(duration)
        return duration

    def _calculate_and_format_duration(self) -> str:
        """
        Calculate session duration and format it

        Returns:
            Formatted duration string
        """
        elapsed = datetime.now() - self._session_start
        total_seconds = int(elapsed.total_seconds())

        # Format based on duration range
        if total_seconds < 60:
            return f"{total_seconds}s"
        elif total_seconds < 3600:
            minutes = total_seconds // 60
            seconds = total_seconds % 60
            if seconds > 0:
                return f"{minutes}m {seconds}s"
            return f"{minutes}m"
        else:
            hours = total_seconds // 3600
            minutes = (total_seconds % 3600) // 60
            if minutes > 0:
                return f"{hours}h {minutes}m"
            return f"{hours}h"

    def _is_cache_valid(self) -> bool:
        """Check if duration cache is still valid"""
        if self._duration_cache is None or self._cache_time is None:
            return False
        return datetime.now() - self._cache_time < self._cache_ttl

    def _update_cache(self, duration: str) -> None:
        """Update duration cache"""
        self._duration_cache = duration
        self._cache_time = datetime.now()
