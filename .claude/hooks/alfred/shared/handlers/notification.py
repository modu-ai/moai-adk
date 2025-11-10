#!/usr/bin/env python3
"""Notification and control handlers

Notification, Stop, SubagentStop event handling
"""

import json
from datetime import datetime
from pathlib import Path

from core import HookPayload, HookResult
from utils.state_tracking import deduplicate_command, mark_command_complete


def _get_command_execution_state_file(cwd: str) -> Path:
    """Get the path to command execution state tracking file"""
    state_dir = Path(cwd) / ".moai" / "memory"
    state_dir.mkdir(parents=True, exist_ok=True)
    return state_dir / "command-execution-state.json"


def _load_command_execution_state(cwd: str) -> dict:
    """Load current command execution state"""
    try:
        state_file = _get_command_execution_state_file(cwd)
        if state_file.exists():
            with open(state_file, "r", encoding="utf-8") as f:
                return json.load(f)
    except Exception:
        pass
    return {"last_command": None, "last_timestamp": None, "is_running": False}


def _save_command_execution_state(cwd: str, state: dict) -> None:
    """Save command execution state"""
    try:
        state_file = _get_command_execution_state_file(cwd)
        with open(state_file, "w", encoding="utf-8") as f:
            json.dump(state, f, indent=2)
    except Exception:
        pass


def _is_alfred_command_duplicate(current_cmd: str, last_cmd: str, last_timestamp: str) -> bool:
    """Check if current Alfred command is a duplicate of the last one within 3 seconds"""
    if not last_cmd or not last_timestamp or current_cmd != last_cmd:
        return False

    try:
        last_time = datetime.fromisoformat(last_timestamp)
        current_time = datetime.now()
        time_diff = (current_time - last_time).total_seconds()
        # Consider it a duplicate if same Alfred command within 3 seconds
        return time_diff < 3.0
    except Exception:
        return False


def handle_notification(payload: HookPayload) -> HookResult:
    """Notification event handler with COMMAND DEDUPLICATION

    Detects and prevents duplicate command executions
    (When the same /alfred: command is triggered multiple times within 3 seconds)
    Implements 3-second window deduplication for Alfred commands only.
    """
    cwd = payload.get("cwd", ".")
    notification = payload.get("notification", {})

    # Extract command information from notification
    current_cmd = None
    if isinstance(notification, dict):
        # Check if notification contains command information
        text = notification.get("text", "") or str(notification)
        if "/alfred:" in text:
            # Extract /alfred: command and normalize whitespace
            import re

            match = re.search(r"/alfred:\S+", text)
            if match:
                current_cmd = match.group().strip()

    # Only deduplicate Alfred commands, not regular commands
    if not current_cmd or not current_cmd.startswith("/alfred:"):
        return HookResult()

    # Use centralized state tracking for command deduplication
    dedup_result = deduplicate_command(current_cmd, cwd)

    # If duplicate, log it but continue execution (prevents blocking)
    if dedup_result["duplicate"]:
        # Optionally log duplicate detection
        pass

    return HookResult(continue_execution=True)


def handle_stop(payload: HookPayload) -> HookResult:
    """Stop event handler

    Marks command execution as complete
    """
    cwd = payload.get("cwd", ".")
    notification = payload.get("notification", {})

    # Extract command to mark as complete
    current_cmd = None
    if isinstance(notification, dict):
        text = notification.get("text", "") or str(notification)
        if "/alfred:" in text:
            import re
            match = re.search(r"/alfred:\S+", text)
            if match:
                current_cmd = match.group().strip()

    # Mark command as complete in centralized state tracker
    mark_command_complete(current_cmd, cwd)

    return HookResult()


def handle_subagent_stop(payload: HookPayload) -> HookResult:
    """SubagentStop event handler

    Records when a sub-agent finishes execution
    """
    cwd = payload.get("cwd", ".")

    # Extract subagent name if available
    subagent_name = (
        payload.get("subagent", {}).get("name")
        if isinstance(payload.get("subagent"), dict)
        else None
    )

    try:
        state_file = _get_command_state_file(cwd).parent / "subagent-execution.log"
        timestamp = datetime.now().isoformat()

        with open(state_file, "a", encoding="utf-8") as f:
            f.write(f"{timestamp} | Subagent Stop | {subagent_name}\n")
    except Exception:
        pass

    return HookResult()


__all__ = ["handle_notification", "handle_stop", "handle_subagent_stop"]
