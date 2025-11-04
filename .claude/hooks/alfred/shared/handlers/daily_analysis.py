#!/usr/bin/env python3
"""Daily session analysis handler

Runs daily analysis of Claude Code session logs.
Automatically skips if analysis already ran today.
Silent operation (no user message output).
"""

import subprocess
from datetime import datetime
from pathlib import Path
from typing import Any

from core import HookPayload, HookResult


def get_last_analysis_date(cwd: str) -> str | None:
    """Get the date of last successful analysis

    Args:
        cwd: Current working directory

    Returns:
        Date string (YYYY-MM-DD) or None if file doesn't exist
    """
    cache_file = Path(cwd) / ".moai" / "cache" / "last_analysis_date.json"
    if not cache_file.exists():
        return None

    try:
        import json
        data = json.loads(cache_file.read_text())
        return data.get("last_analysis_date")
    except Exception:
        return None


def save_last_analysis_date(cwd: str, date_str: str) -> bool:
    """Save the date of last successful analysis

    Args:
        cwd: Current working directory
        date_str: Date string (YYYY-MM-DD)

    Returns:
        True if successful, False otherwise
    """
    cache_dir = Path(cwd) / ".moai" / "cache"
    cache_dir.mkdir(parents=True, exist_ok=True)
    cache_file = cache_dir / "last_analysis_date.json"

    try:
        import json
        data = {
            "last_analysis_date": date_str,
            "last_analysis_success": True,
        }
        cache_file.write_text(json.dumps(data, indent=2))
        return True
    except Exception:
        return False


def run_session_analyzer(cwd: str) -> bool:
    """Run session analyzer for previous day (--days 1)

    Args:
        cwd: Current working directory

    Returns:
        True if successful, False otherwise
    """
    try:
        result = subprocess.run(
            ["python3", ".moai/scripts/session_analyzer.py", "--days", "1"],
            cwd=cwd,
            timeout=4,  # Leave 1 second buffer for other operations
            capture_output=True,
            text=True,
        )
        return result.returncode == 0
    except subprocess.TimeoutExpired:
        return False
    except FileNotFoundError:
        # session_analyzer.py not found
        return False
    except Exception:
        return False


def handle_daily_analysis(payload: HookPayload) -> HookResult:
    """SessionStart handler for daily session analysis

    Runs analysis of Claude Code session logs from the previous day.
    Silently skips if already analyzed today.
    No user-visible output.

    Args:
        payload: Claude Code event payload (cwd key required)

    Returns:
        HookResult with no system message (silent operation)

    Logic:
        1. Check if analysis already ran today
        2. If yes, skip (silent)
        3. If no, run session_analyzer.py --days 1
        4. Update last_analysis_date on success
        5. Silent operation - no user message

    @TAG:HOOKS-DAILY-ANALYSIS-001
    """
    cwd = payload.get("cwd", ".")
    today = datetime.now().strftime("%Y-%m-%d")

    # Check if already analyzed today
    last_date = get_last_analysis_date(cwd)
    if last_date == today:
        # Already analyzed today - skip silently
        return HookResult(continue_execution=True)

    # Run daily analysis
    success = run_session_analyzer(cwd)

    if success:
        # Save today's date
        save_last_analysis_date(cwd, today)

    # Always return successful continuation (silent operation)
    return HookResult(continue_execution=True)


__all__ = ["handle_daily_analysis"]
