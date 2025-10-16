#!/usr/bin/env python3
"""Tool usage handlers

PreToolUse, PostToolUse 이벤트 처리
"""

from core import HookPayload, HookResult
from core.checkpoint import create_checkpoint, detect_risky_operation


def handle_pre_tool_use(payload: HookPayload) -> HookResult:
    """PreToolUse 이벤트 핸들러 (Event-Driven Checkpoint 통합)

    위험한 작업 전 자동으로 checkpoint를 생성합니다.
    Claude Code tool 사용 전에 호출되며, 위험 감지 시 사용자에게 알립니다.

    Args:
        payload: Claude Code 이벤트 페이로드
                 (tool, arguments, cwd 키 포함)

    Returns:
        HookResult(
            message=checkpoint 생성 알림 (위험 감지 시),
            blocked=False (항상 작업 계속 진행)
        )

    Checkpoint Triggers:
        - Bash: rm -rf, git merge, git reset --hard
        - Edit/Write: CLAUDE.md, config.json
        - MultiEdit: ≥10 files

    Examples:
        Bash tool (rm -rf) 감지:
        → "🛡️ Checkpoint created: before-delete-20251015-143000"

    Notes:
        - 위험 감지 후에도 blocked=False 반환 (작업 계속)
        - Checkpoint 실패 시에도 작업 진행 (무시)
        - 투명한 백그라운드 동작

    @TAG:CHECKPOINT-EVENT-001
    """
    tool_name = payload.get("tool", "")
    tool_args = payload.get("arguments", {})
    cwd = payload.get("cwd", ".")

    # 위험한 작업 감지
    is_risky, operation_type = detect_risky_operation(tool_name, tool_args, cwd)

    # 위험 감지 시 checkpoint 생성
    if is_risky:
        checkpoint_branch = create_checkpoint(cwd, operation_type)

        if checkpoint_branch != "checkpoint-failed":
            message = (
                f"🛡️ Checkpoint created: {checkpoint_branch}\n"
                f"   Operation: {operation_type}\n"
                f"   Restore: /alfred:0-project restore"
            )

            return HookResult(message=message, blocked=False)

    return HookResult(blocked=False)


def handle_post_tool_use(payload: HookPayload) -> HookResult:
    """PostToolUse 이벤트 핸들러 (기본 구현)"""
    return HookResult()


__all__ = ["handle_pre_tool_use", "handle_post_tool_use"]
