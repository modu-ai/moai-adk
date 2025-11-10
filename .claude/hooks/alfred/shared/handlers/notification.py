#!/usr/bin/env python3
"""Notification and control handlers

Notification, Stop, SubagentStop event handling
"""

import json
from datetime import datetime
from pathlib import Path

import time
from core import HookPayload, HookResult, ExecutionResult, get_logger, get_performance_metrics
from utils.state_tracking import deduplicate_command, mark_command_complete, get_state_manager, HookConfiguration


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
    (When the same /alfred: command is triggered multiple times within configurable time window)
    Implements configurable window deduplication for Alfred commands only.

    Args:
        payload: Claude Code event payload containing notification and cwd

    Returns:
        HookResult with continue_execution flag
    """
    logger = get_logger()

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
        logger.debug(f"Non-Alfred command: {current_cmd}")
        return HookResult()

    try:
        # Get state manager with configuration
        state_manager = get_state_manager(cwd)
        config = state_manager.config

        # Use centralized state tracking for command deduplication
        dedup_result = deduplicate_command(current_cmd, cwd, config)

        # Log duplicate detection if enabled
        if dedup_result.duplicate:
            logger.info(f"Command duplicate detected: {current_cmd} (reason: {dedup_result.reason})")
        else:
            logger.debug(f"Command execution allowed: {current_cmd} (reason: {dedup_result.reason})")

        return HookResult(continue_execution=True)

    except Exception as e:
        logger.error(f"Error in command deduplication: {e}")
        # Continue execution despite deduplication failure
        return HookResult(continue_execution=True)


def handle_stop(payload: HookPayload) -> HookResult:
    """Stop event handler

    Marks command execution as complete

    Args:
        payload: Claude Code event payload containing notification and cwd

    Returns:
        HookResult with continue_execution flag
    """
    logger = get_logger()

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

    try:
        # Get state manager with configuration
        state_manager = get_state_manager(cwd)
        config = state_manager.config

        # Mark command as complete in centralized state tracker
        mark_command_complete(current_cmd, cwd, config)
        logger.info(f"Command marked as complete: {current_cmd or 'unknown'}")

    except Exception as e:
        logger.error(f"Failed to mark command as complete: {e}")
        # Continue execution despite failure

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
