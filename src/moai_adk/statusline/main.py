# type: ignore
# -*- coding: utf-8 -*-
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
        # Handle Docker/non-interactive environments by checking TTY
        input_data = sys.stdin.read() if not sys.stdin.isatty() else "{}"
        if input_data:
            try:
                return json.loads(input_data)
            except json.JSONDecodeError as e:
                import logging

                logging.error(f"Failed to parse JSON from stdin: {e}")
                logging.debug(f"Input data: {input_data[:200]}")
                return {}
        return {}
    except (EOFError, ValueError) as e:
        import logging

        logging.error(f"Error reading stdin: {e}")
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
    except (OSError, AttributeError, RuntimeError):
        # Git collector errors (file access, attribute, or runtime errors)
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
    except (OSError, AttributeError, ValueError):
        # Metrics tracker errors (file access, attribute, or value errors)
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
    except (OSError, AttributeError, RuntimeError):
        # Alfred detector errors (file access, attribute, or runtime errors)
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
    except (ImportError, AttributeError, OSError):
        # Version reader errors (import, attribute, or file access errors)
        return "unknown"


# safe_collect_output_style function removed - no longer needed


def safe_collect_memory() -> str:
    """
    Safely collect memory usage with fallback.

    Returns:
        Formatted memory usage string (e.g., "128MB")
    """
    try:
        from .memory_collector import MemoryCollector

        collector = MemoryCollector()
        return collector.get_display_string(mode="process")
    except (ImportError, OSError, AttributeError, RuntimeError):
        # Memory collector errors
        return "N/A"


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
    except (OSError, AttributeError, RuntimeError, ValueError):
        # Update checker errors (file access, attribute, runtime, or value errors)
        return False, None


def format_token_count(tokens: int) -> str:
    """
    Format token count for display (e.g., 15234 -> "15K").

    Args:
        tokens: Number of tokens

    Returns:
        Formatted string
    """
    if tokens >= 1000:
        return f"{tokens // 1000}K"
    return str(tokens)


def extract_context_window(session_context: dict) -> str:
    """
    Extract and format context window usage from session context.

    Args:
        session_context: Context passed from Claude Code via stdin

    Returns:
        Formatted context window string (e.g., "15K/200K")
    """
    context_info = session_context.get("context_window", {})

    if not context_info:
        return ""

    # Get context window size
    context_size = context_info.get("context_window_size", 0)
    if not context_size:
        return ""

    # Get current usage
    current_usage = context_info.get("current_usage", {})
    if current_usage:
        # Calculate total current tokens
        input_tokens = current_usage.get("input_tokens", 0)
        cache_creation = current_usage.get("cache_creation_input_tokens", 0)
        cache_read = current_usage.get("cache_read_input_tokens", 0)
        current_tokens = input_tokens + cache_creation + cache_read
    else:
        # Fallback to total tokens
        current_tokens = context_info.get("total_input_tokens", 0)

    if current_tokens > 0:
        return f"{format_token_count(current_tokens)}/{format_token_count(context_size)}"

    return ""


def build_statusline_data(session_context: dict, mode: str = "compact") -> str:
    """
    Build complete statusline string from all data sources.

    Collects information from:
    - Claude Code session context (via stdin)
    - Git repository
    - Session metrics
    - Alfred workflow state
    - MoAI-ADK version
    - Update checker
    - Output style
    - Context window usage

    Args:
        session_context: Context passed from Claude Code via stdin
        mode: Display mode (compact, extended, minimal)

    Returns:
        Rendered statusline string
    """
    try:
        # Extract model from session context with Claude Code version
        model_info = session_context.get("model", {})
        model_name = model_info.get("display_name") or model_info.get("name") or "Unknown"

        # Extract Claude Code version separately for new layout
        claude_version = session_context.get("version", "")
        model = model_name

        # Extract directory
        cwd = session_context.get("cwd", "")
        if cwd:
            directory = Path(cwd).name or Path(cwd).parent.name or "project"
        else:
            directory = "project"

        # Extract output style from session context
        output_style = session_context.get("output_style", {}).get("name", "")

        # Extract context window usage
        context_window = extract_context_window(session_context)

        # Collect all information from local sources
        branch, git_status = safe_collect_git_info()
        duration = safe_collect_duration()
        active_task = safe_collect_alfred_task()
        version = safe_collect_version()
        memory_usage = safe_collect_memory()
        update_available, latest_version = safe_check_update(version)

        # Build StatuslineData with dynamic fields
        data = StatuslineData(
            model=model,
            claude_version=claude_version,
            version=version,
            memory_usage=memory_usage,
            branch=branch,
            git_status=git_status,
            duration=duration,
            directory=directory,
            active_task=active_task,
            output_style=output_style,
            update_available=update_available,
            latest_version=latest_version,
            context_window=context_window,
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
    # Debug mode check
    debug_mode = os.environ.get("MOAI_STATUSLINE_DEBUG") == "1"

    # Read session context from Claude Code
    session_context = read_session_context()

    if debug_mode:
        # Write debug info to stderr for troubleshooting
        sys.stderr.write(f"[DEBUG] Received session_context: {json.dumps(session_context, indent=2)}\n")
        sys.stderr.flush()

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
    if debug_mode:
        sys.stderr.write(f"[DEBUG] Generated statusline: {statusline}\n")
        sys.stderr.flush()

    if statusline:
        print(statusline, end="")


if __name__ == "__main__":
    main()
