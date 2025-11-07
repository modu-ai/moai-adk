"""
Statusline module for Claude Code status display

@SPEC:CLAUDE-STATUSLINE-001
Provides real-time status information display in Claude Code terminal
"""

__version__ = "0.1.0"

from .renderer import StatuslineRenderer, StatuslineData
from .git_collector import GitCollector, GitInfo
from .metrics_tracker import MetricsTracker
from .alfred_detector import AlfredDetector, AlfredTask
from .version_reader import VersionReader
from .update_checker import UpdateChecker, UpdateInfo

__all__ = [
    "StatuslineRenderer",
    "StatuslineData",
    "GitCollector",
    "GitInfo",
    "MetricsTracker",
    "AlfredDetector",
    "AlfredTask",
    "VersionReader",
    "UpdateChecker",
    "UpdateInfo",
]
