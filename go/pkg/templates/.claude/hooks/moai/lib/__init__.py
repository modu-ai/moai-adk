# type: ignore
"""
Statusline module for Claude Code status display

Provides real-time status information display in Claude Code terminal

Enhanced version includes:
- Improved version reader with caching and error handling
- Async support for better performance
- Configurable version reading behavior
- Comprehensive fallback strategies
"""

__version__ = "0.1.0"

from .alfred_detector import AlfredDetector, AlfredTask
from .config import StatuslineConfig
from .git_collector import GitCollector, GitInfo
from .memory_collector import MemoryCollector, MemoryInfo
from .metrics_tracker import MetricsTracker
from .renderer import StatuslineData, StatuslineRenderer
from .update_checker import UpdateChecker, UpdateInfo
from .version_reader import VersionConfig, VersionReader, VersionReadError

__all__ = [
    "StatuslineRenderer",
    "StatuslineData",
    "StatuslineConfig",
    "GitCollector",
    "GitInfo",
    "MemoryCollector",
    "MemoryInfo",
    "MetricsTracker",
    "AlfredDetector",
    "AlfredTask",
    "VersionReader",
    "VersionConfig",
    "VersionReadError",
    "UpdateChecker",
    "UpdateInfo",
]
