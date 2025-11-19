#!/usr/bin/env python3
"""Tool usage handlers

PreToolUse, PostToolUse event handling
"""

import json
import sys
from pathlib import Path
from typing import Optional

from lib import HookPayload, HookResult
from lib.checkpoint import create_checkpoint, detect_risky_operation


def get_auto_checkpoint_setting(cwd: str) -> str:
    """Get auto_checkpoint setting from .moai/config/config.json

    Args:
        cwd: Current working directory (project root)

    Returns:
        auto_checkpoint setting value ("disabled", "event-driven", "manual")
        Defaults to "disabled" if not found or error occurs
    """
    try:
        config_path = Path(cwd) / ".moai" / "config" / "config.json"
        if not config_path.exists():
            return "disabled"

        with open(config_path, "r", encoding="utf-8") as f:
            config = json.load(f)

        # Navigate to git_strategy.personal.auto_checkpoint
        git_strategy = config.get("git_strategy", {})
        personal = git_strategy.get("personal", {})
        auto_checkpoint = personal.get("auto_checkpoint", "disabled")

        # Validate setting value
        valid_values = ["disabled", "event-driven", "manual"]
        return auto_checkpoint if auto_checkpoint in valid_values else "disabled"

    except (json.JSONDecodeError, FileNotFoundError, KeyError):
        # If any error occurs, default to disabled for safety
        return "disabled"


def handle_pre_tool_use(payload: HookPayload) -> HookResult:
    """PreToolUse event handler (Event-Driven Checkpoint integration)

    Automatically creates checkpoints before dangerous operations, but only if
    auto_checkpoint is enabled in .moai/config/config.json.
    Called before using the Claude Code tool, it notifies the user when danger is detected.

    Args:
        payload: Claude Code event payload
                 (includes tool, arguments, cwd keys)

    Returns:
        HookResult(
            system_message=checkpoint creation notification (when danger is detected);
            continue_execution=True (always continue operation)
        )

    Checkpoint Triggers:
        - Bash: rm -rf, git merge, git reset --hard
        - Edit/Write: CLAUDE.md, config.json
        - MultiEdit: ≥10 files

    Examples:
        Bash tool (rm -rf) detection:
        → "🛡️ Checkpoint created: before-delete-20251015-143000"

    Notes:
        - Return continue_execution=True even after detection of danger (continue operation)
        - Work continues even when checkpoint fails (ignores)
        - Transparent background operation
        - Respects auto_checkpoint setting from .moai/config/config.json

    """
    tool_name = payload.get("tool", "")
    tool_args = payload.get("arguments", {})
    cwd = payload.get("cwd", ".")

    # Check if auto_checkpoint is disabled
    auto_checkpoint = get_auto_checkpoint_setting(cwd)
    if auto_checkpoint == "disabled":
        # Skip checkpoint creation when disabled
        return HookResult(continue_execution=True)

    # Dangerous operation detection (best-effort)
    try:
        is_risky, operation_type = detect_risky_operation(tool_name, tool_args, cwd)
        # Create checkpoint when danger is detected
        if is_risky:
            checkpoint_branch = create_checkpoint(cwd, operation_type)
            if checkpoint_branch != "checkpoint-failed":
                system_message = (
                    f"🛡️ Checkpoint created: {checkpoint_branch}\n   Operation: {operation_type}"
                )
                return HookResult(system_message=system_message, continue_execution=True)
    except Exception:
        # Do not fail the hook if risk detection errors out
        pass

    return HookResult(continue_execution=True)


def handle_post_tool_use(payload: HookPayload) -> HookResult:
    """PostToolUse event handler (default implementation)"""
    return HookResult()


__all__ = ["handle_pre_tool_use", "handle_post_tool_use"]
