#!/usr/bin/env python3
"""Tool usage handlers

PreToolUse, PostToolUse event handling
"""

from lib import HookPayload, HookResult
from lib.checkpoint import create_checkpoint, detect_risky_operation


def handle_pre_tool_use(payload: HookPayload) -> HookResult:
    """PreToolUse event handler (Event-Driven Checkpoint integration)

    Automatically creates checkpoints before dangerous operations.
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

    """
    tool_name = payload.get("tool", "")
    tool_args = payload.get("arguments", {})
    cwd = payload.get("cwd", ".")

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
