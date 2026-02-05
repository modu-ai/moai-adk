"""
Memory collector for MoAI-ADK statusline

Collects memory usage information for display in the status bar.
Uses psutil for cross-platform memory monitoring.
"""

import logging
import os
from dataclasses import dataclass
from datetime import datetime, timedelta
from typing import Optional

logger = logging.getLogger(__name__)


@dataclass
class MemoryInfo:
    """Memory usage information"""

    # Process memory (current Python process)
    process_rss_mb: float  # Resident Set Size
    process_vms_mb: float  # Virtual Memory Size

    # System memory
    system_total_mb: float
    system_available_mb: float
    system_percent: float

    # Formatted strings for display
    display_process: str  # e.g., "128MB"
    display_system: str  # e.g., "8.2GB/16GB"
    display_percent: str  # e.g., "51%"

    timestamp: datetime


class MemoryCollector:
    """
    Memory collector for statusline display.

    Features:
    - Process memory monitoring (current Python process)
    - System memory monitoring
    - Caching with configurable TTL
    - Graceful degradation on errors
    - Multiple display format options
    """

    # Cache configuration
    DEFAULT_CACHE_TTL_SECONDS = 10  # Memory doesn't change rapidly

    def __init__(self, cache_ttl_seconds: Optional[int] = None):
        """
        Initialize memory collector.

        Args:
            cache_ttl_seconds: Cache TTL in seconds. If None, uses default.
        """
        self._cache_ttl = timedelta(seconds=cache_ttl_seconds or self.DEFAULT_CACHE_TTL_SECONDS)
        self._cache: Optional[MemoryInfo] = None
        self._cache_time: Optional[datetime] = None
        self._logger = logging.getLogger(__name__)

        # Check if psutil is available
        self._psutil_available = self._check_psutil()

    def _check_psutil(self) -> bool:
        """Check if psutil is available."""
        try:
            import psutil  # noqa: F401

            return True
        except ImportError:
            self._logger.warning("psutil not available, memory monitoring disabled")
            return False

    def get_memory_info(self, force_refresh: bool = False) -> Optional[MemoryInfo]:
        """
        Get memory usage information.

        Args:
            force_refresh: If True, bypass cache and force fresh collection

        Returns:
            MemoryInfo object or None if collection fails
        """
        if not self._psutil_available:
            return None

        # Check cache
        if not force_refresh and self._is_cache_valid():
            return self._cache

        try:
            memory_info = self._collect_memory_info()
            self._update_cache(memory_info)
            return memory_info
        except Exception as e:
            self._logger.error(f"Error collecting memory info: {e}")
            return self._cache  # Return stale cache on error

    def get_display_string(self, mode: str = "process", force_refresh: bool = False) -> str:
        """
        Get formatted memory string for statusline display.

        Args:
            mode: Display mode
                - "process": Process memory (e.g., "128MB")
                - "system": System memory usage (e.g., "8.2GB/16GB")
                - "percent": System memory percentage (e.g., "51%")
                - "compact": Process memory compact (e.g., "128M")
            force_refresh: If True, bypass cache

        Returns:
            Formatted memory string, or "N/A" on error
        """
        memory_info = self.get_memory_info(force_refresh)

        if memory_info is None:
            return "N/A"

        if mode == "process":
            return memory_info.display_process
        elif mode == "system":
            return memory_info.display_system
        elif mode == "percent":
            return memory_info.display_percent
        elif mode == "compact":
            # Compact format without "MB" suffix
            return self._format_compact(memory_info.process_rss_mb)
        else:
            return memory_info.display_process

    def _collect_memory_info(self) -> MemoryInfo:
        """
        Collect memory information using psutil.

        Returns:
            MemoryInfo object with current memory usage
        """
        import psutil

        # Get current process info
        process = psutil.Process(os.getpid())
        process_memory = process.memory_info()

        # Convert to MB
        process_rss_mb = process_memory.rss / (1024 * 1024)
        process_vms_mb = process_memory.vms / (1024 * 1024)

        # Get system memory info
        system_memory = psutil.virtual_memory()
        system_total_mb = system_memory.total / (1024 * 1024)
        system_available_mb = system_memory.available / (1024 * 1024)
        system_percent = system_memory.percent

        # Format display strings
        display_process = self._format_size(process_rss_mb)
        display_system = (
            f"{self._format_size(system_total_mb - system_available_mb)}/{self._format_size(system_total_mb)}"
        )
        display_percent = f"{system_percent:.0f}%"

        return MemoryInfo(
            process_rss_mb=process_rss_mb,
            process_vms_mb=process_vms_mb,
            system_total_mb=system_total_mb,
            system_available_mb=system_available_mb,
            system_percent=system_percent,
            display_process=display_process,
            display_system=display_system,
            display_percent=display_percent,
            timestamp=datetime.now(),
        )

    def _format_size(self, size_mb: float) -> str:
        """
        Format size in MB to human-readable string.

        Args:
            size_mb: Size in megabytes

        Returns:
            Formatted string (e.g., "128MB", "1.2GB")
        """
        if size_mb >= 1024:
            return f"{size_mb / 1024:.1f}GB"
        elif size_mb >= 100:
            return f"{size_mb:.0f}MB"
        elif size_mb >= 10:
            return f"{size_mb:.1f}MB"
        else:
            return f"{size_mb:.2f}MB"

    def _format_compact(self, size_mb: float) -> str:
        """
        Format size in compact format for tight spaces.

        Args:
            size_mb: Size in megabytes

        Returns:
            Compact formatted string (e.g., "128M", "1.2G")
        """
        if size_mb >= 1024:
            return f"{size_mb / 1024:.1f}G"
        else:
            return f"{size_mb:.0f}M"

    def _is_cache_valid(self) -> bool:
        """Check if cache is still valid."""
        if self._cache is None or self._cache_time is None:
            return False

        age = datetime.now() - self._cache_time
        return age < self._cache_ttl

    def _update_cache(self, memory_info: MemoryInfo) -> None:
        """Update cache with new memory info."""
        self._cache = memory_info
        self._cache_time = datetime.now()

    def clear_cache(self) -> None:
        """Clear the memory cache."""
        self._cache = None
        self._cache_time = None

    def get_cache_age_seconds(self) -> Optional[float]:
        """
        Get cache age in seconds.

        Returns:
            Cache age in seconds, or None if no cache
        """
        if self._cache_time is None:
            return None
        return (datetime.now() - self._cache_time).total_seconds()


# Singleton instance for convenience
_default_collector: Optional[MemoryCollector] = None


def get_memory_collector() -> MemoryCollector:
    """
    Get the default memory collector instance.

    Returns:
        MemoryCollector singleton instance
    """
    global _default_collector
    if _default_collector is None:
        _default_collector = MemoryCollector()
    return _default_collector


def get_memory_display(mode: str = "process") -> str:
    """
    Convenience function to get memory display string.

    Args:
        mode: Display mode ("process", "system", "percent", "compact")

    Returns:
        Formatted memory string
    """
    return get_memory_collector().get_display_string(mode)
