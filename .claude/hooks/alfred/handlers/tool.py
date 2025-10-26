#!/usr/bin/env python3
# @CODE:HOOK-TOOL-001 | SPEC: TBD | TEST: tests/hooks/test_handlers.py
"""Tool usage handlers

PreToolUse, PostToolUse event handling
"""

from core import HookPayload, HookResult
from core.checkpoint import create_checkpoint, detect_risky_operation
from core.tags import scan_recent_changes_for_missing_tags


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

    @TAG:CHECKPOINT-EVENT-001
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
                    f"🛡️ Checkpoint created: {checkpoint_branch}\n"
                    f"   Operation: {operation_type}"
                )
                return HookResult(system_message=system_message, continue_execution=True)
    except Exception:
        # Do not fail the hook if risk detection errors out
        pass

    # TAG Guard (gentle): warn when recent changes miss @TAG
    issues = scan_recent_changes_for_missing_tags(cwd)
    if issues:
        # Summarize first few issues for display
        preview = "\n".join(
            f" - {i.path} → 기대 태그: {i.expected}" for i in issues[:5]
        )
        more = "" if len(issues) <= 5 else f"\n (외 {len(issues)-5}건 더 존재)"
        msg = (
            "⚠️ TAG 누락 감지: 생성/수정한 파일 중 @TAG가 없는 항목이 있습니다.\n"
            f"{preview}{more}\n"
            "권장 조치:\n"
            "  1) SPEC/TEST/CODE/DOC 유형에 맞는 @TAG를 파일 상단 주석이나 헤더에 추가\n"
            "  2) rg로 확인: rg '@(SPEC|TEST|CODE|DOC):' -n <경로>\n"
        )
        return HookResult(system_message=msg, continue_execution=True)

    return HookResult(continue_execution=True)


def handle_post_tool_use(payload: HookPayload) -> HookResult:
    """PostToolUse event handler (default implementation)"""
    return HookResult()


__all__ = ["handle_pre_tool_use", "handle_post_tool_use"]
