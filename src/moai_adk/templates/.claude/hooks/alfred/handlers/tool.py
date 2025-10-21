#!/usr/bin/env python3
"""Tool usage handlers

PreToolUse, PostToolUse event handling
"""

from core import HookPayload, HookResult
from core.checkpoint import create_checkpoint, detect_risky_operation


def handle_pre_tool_use(payload: HookPayload) -> HookResult:
    """PreToolUse event handler (Event-Driven Checkpoint integration)

    Automatically creates checkpoints before dangerous operations.
    Called before using the Claude Code tool, it notifies the user when danger is detected.

    Args:
        payload: Claude Code event payload
                 (includes tool, arguments, cwd keys)

    Returns:
        HookResult(
            message=checkpoint creation notification (when danger is detected);
            blocked=False (always continue operation)
        )

    Checkpoint Triggers:
        - Bash: rm -rf, git merge, git reset --hard
        - Edit/Write: CLAUDE.md, config.json
        - MultiEdit: ≥10 files

    Examples:
        Bash tool (rm -rf) detection:
        → "🛡️ Checkpoint created: before-delete-20251015-143000"

    Notes:
        - Return blocked=False even after detection of danger (continue operation)
        - Work continues even when checkpoint fails (ignores)
        - Transparent background operation

    @TAG:CHECKPOINT-EVENT-001
    """
    tool_name = payload.get("tool", "")
    tool_args = payload.get("arguments", {})
    cwd = payload.get("cwd", ".")

    # Dangerous operation detection
    is_risky, operation_type = detect_risky_operation(tool_name, tool_args, cwd)

    # Create checkpoint when danger is detected
    if is_risky:
        checkpoint_branch = create_checkpoint(cwd, operation_type)

        if checkpoint_branch != "checkpoint-failed":
            message = (
                f"🛡️ Checkpoint created: {checkpoint_branch}\n"
                f"   Operation: {operation_type}"
            )

            return HookResult(message=message, blocked=False)

    return HookResult(blocked=False)


def handle_post_tool_use(payload: HookPayload) -> HookResult:
    """PostToolUse event handler (default implementation)"""
    return HookResult()


__all__ = ["handle_pre_tool_use", "handle_post_tool_use"]
