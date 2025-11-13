# type: ignore
#!/usr/bin/env python3
"""
Claude Code Statusline Integration


Main entry point for MoAI-ADK statusline rendering in Claude Code.
Collects all necessary information from the project and renders it
in the specified format for display in the status bar.
"""

import json
import os
import sys
from pathlib import Path
from typing import Optional

from .alfred_detector import AlfredDetector
from .config import StatuslineConfig
from .git_collector import GitCollector
from .metrics_tracker import MetricsTracker
from .renderer import StatuslineData, StatuslineRenderer
from .update_checker import UpdateChecker
from .version_reader import VersionReader


def read_session_context() -> dict:
    """
    Read JSON context from stdin (sent by Claude Code).

    Returns:
        Dictionary containing session information
    """
    try:
        input_data = sys.stdin.read()
        if input_data:
            return json.loads(input_data)
        return {}
    except (json.JSONDecodeError, EOFError, ValueError):
        return {}


def safe_collect_git_info() -> tuple[str, str]:
    """
    Safely collect git information with fallback.

    Returns:
        Tuple of (branch_name, git_status_str)
    """
    try:
        collector = GitCollector()
        git_info = collector.collect_git_info()

        branch = git_info.branch or "unknown"
        git_status = f"+{git_info.staged} M{git_info.modified} ?{git_info.untracked}"

        return branch, git_status
    except Exception:
        return "N/A", ""


def safe_collect_duration() -> str:
    """
    Safely collect session duration with fallback.

    Returns:
        Formatted duration string
    """
    try:
        tracker = MetricsTracker()
        return tracker.get_duration()
    except Exception:
        return "0m"


def safe_collect_alfred_task() -> str:
    """
    Safely collect active Alfred task with fallback.

    Returns:
        Formatted task string
    """
    try:
        detector = AlfredDetector()
        task = detector.detect_active_task()

        if task.command:
            stage_suffix = f"-{task.stage}" if task.stage else ""
            return f"[{task.command.upper()}{stage_suffix}]"
        return ""
    except Exception:
        return ""


def safe_collect_version() -> str:
    """
    Safely collect MoAI-ADK version with fallback.

    Returns:
        Version string
    """
    try:
        reader = VersionReader()
        version = reader.get_version()
        return version or "unknown"
    except Exception:
        return "unknown"


# safe_collect_output_style function removed - no longer needed


def safe_check_update(current_version: str) -> tuple[bool, Optional[str]]:
    """
    Safely check for updates with fallback.

    Args:
        current_version: Current version string

    Returns:
        Tuple of (update_available, latest_version)
    """
    try:
        checker = UpdateChecker()
        update_info = checker.check_for_update(current_version)

        return update_info.available, update_info.latest_version
    except Exception:
        return False, None


def build_statusline_data(session_context: dict, mode: str = "compact") -> str:
    """
    Build complete statusline string from all data sources.

    Collects information from:
    - Claude Code session context
    - Git repository
    - Session metrics
    - Alfred workflow state
    - MoAI-ADK version
    - Update checker

    Args:
        session_context: Context passed from Claude Code
        mode: Display mode (compact, extended, minimal)

    Returns:
        Rendered statusline string
    """
    try:
        # Extract model from session context
        model_info = session_context.get("model", {})
        model = model_info.get("display_name", "Unknown")
        if not model:
            model = model_info.get("name", "Unknown")

        # Extract directory
        cwd = session_context.get("cwd", "")
        if cwd:
            directory = Path(cwd).name or Path(cwd).parent.name or "project"
        else:
            directory = "project"

        # Collect all information
        branch, git_status = safe_collect_git_info()
        duration = safe_collect_duration()
        active_task = safe_collect_alfred_task()
        version = safe_collect_version()
        update_available, latest_version = safe_check_update(version)

        # Build StatuslineData with dynamic fields
        data = StatuslineData(
            model=model,
            version=version,
            memory_usage="256MB",  # TODO: Get actual memory usage
            branch=branch,
            git_status=git_status,
            duration=duration,
            directory=directory,
            active_task=active_task,
            update_available=update_available,
            latest_version=latest_version,
        )

        # Render statusline with labeled sections
        renderer = StatuslineRenderer()
        statusline = renderer.render(data, mode=mode)

        return statusline

    except Exception as e:
        # Graceful degradation on any error
        import logging

        logging.warning(f"Statusline rendering error: {e}")
        return ""


def main():
    """
    Main entry point for Claude Code statusline.

    Reads JSON from stdin, processes all information,
    and outputs the formatted statusline string.
    """
    # Read session context from Claude Code
    session_context = read_session_context()

    # Load configuration
    config = StatuslineConfig()

    # Determine display mode (priority: session context > environment > config > default)
    mode = (
        session_context.get("statusline", {}).get("mode")
        or os.environ.get("MOAI_STATUSLINE_MODE")
        or config.get("statusline.mode")
        or "extended"
    )

    # Build and output statusline
    statusline = build_statusline_data(session_context, mode=mode)
    if statusline:
        print(statusline, end="")


if __name__ == "__main__":
    main()
