#!/usr/bin/env python3
"""Notification and control handlers

Notification, Stop, SubagentStop event handling
"""

import json
from datetime import datetime
from pathlib import Path

from core import HookPayload, HookResult


def _get_command_state_file(cwd: str) -> Path:
    """Get the path to command state tracking file"""
    state_dir = Path(cwd) / ".moai" / "memory"
    state_dir.mkdir(parents=True, exist_ok=True)
    return state_dir / "command-execution-state.json"


def _load_command_state(cwd: str) -> dict:
    """Load current command execution state"""
    try:
        state_file = _get_command_state_file(cwd)
        if state_file.exists():
            with open(state_file, "r", encoding="utf-8") as f:
                return json.load(f)
    except Exception:
        pass
    return {"last_command": None, "last_timestamp": None, "is_running": False}


def _save_command_state(cwd: str, state: dict) -> None:
    """Save command execution state"""
    try:
        state_file = _get_command_state_file(cwd)
        with open(state_file, "w", encoding="utf-8") as f:
            json.dump(state, f, indent=2)
    except Exception:
        pass


def _is_duplicate_command(current_cmd: str, last_cmd: str, last_timestamp: str) -> bool:
    """Check if current command is a duplicate of the last one within 3 seconds"""
    if not last_cmd or not last_timestamp or current_cmd != last_cmd:
        return False

    try:
        last_time = datetime.fromisoformat(last_timestamp)
        current_time = datetime.now()
        time_diff = (current_time - last_time).total_seconds()
        # Consider it a duplicate if same command within 3 seconds
        return time_diff < 3
    except Exception:
        return False


def handle_notification(payload: HookPayload) -> HookResult:
    """Notification event handler

    Detects and warns about duplicate command executions
    (When the same /alfred: command is triggered multiple times within 3 seconds)
    """
    cwd = payload.get("cwd", ".")
    notification = payload.get("notification", {})

    # Extract command information from notification
    current_cmd = None
    if isinstance(notification, dict):
        # Check if notification contains command information
        text = notification.get("text", "") or str(notification)
        if "/alfred:" in text:
            # Extract /alfred: command
            import re

            match = re.search(r"/alfred:\S+", text)
            if match:
                current_cmd = match.group()

    if not current_cmd:
        return HookResult()

    # Load current state
    state = _load_command_state(cwd)
    last_cmd = state.get("last_command")
    last_timestamp = state.get("last_timestamp")

    # Check for duplicate
    if _is_duplicate_command(current_cmd, last_cmd, last_timestamp):
        warning_msg = (
            f"⚠️ Duplicate command detected: '{current_cmd}' "
            f"is running multiple times within 3 seconds.\n"
            f"This may indicate a system issue. Check logs in `.moai/logs/command-invocations.log`"
        )

        # Update state - mark as duplicate detected
        state["duplicate_detected"] = True
        state["duplicate_command"] = current_cmd
        state["duplicate_timestamp"] = datetime.now().isoformat()
        _save_command_state(cwd, state)

        return HookResult(system_message=warning_msg, continue_execution=True)

    # Update state with current command
    state["last_command"] = current_cmd
    state["last_timestamp"] = datetime.now().isoformat()
    state["is_running"] = True
    state["duplicate_detected"] = False
    _save_command_state(cwd, state)

    return HookResult()


def handle_stop(payload: HookPayload) -> HookResult:
    """Stop event handler

    Marks command execution as complete
    """
    cwd = payload.get("cwd", ".")
    state = _load_command_state(cwd)
    state["is_running"] = False
    state["last_timestamp"] = datetime.now().isoformat()
    _save_command_state(cwd, state)

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
