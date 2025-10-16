#!/usr/bin/env python3
"""Context compaction handlers

PreCompact 이벤트 처리
"""

from core import HookPayload, HookResult


def handle_pre_compact(payload: HookPayload) -> HookResult:
    """PreCompact 이벤트 핸들러

    컨텍스트가 70% 이상 차면 새 세션 시작을 제안합니다.
    Context Engineering의 Compaction 원칙을 구현합니다.

    Args:
        payload: Claude Code 이벤트 페이로드

    Returns:
        HookResult(
            message=새 세션 시작 제안 메시지,
            suggestions=구체적인 액션 제안 리스트
        )

    Suggestions:
        - /clear 명령으로 새 세션 시작
        - /new 명령으로 새 대화 시작
        - 핵심 결정사항 요약 후 계속

    Notes:
        - Context Engineering: Compaction 원칙 준수
        - 토큰 사용량 > 70% 시 자동 호출
        - 성능 향상 및 컨텍스트 관리 개선

    TDD History:
        - RED: PreCompact 메시지 및 제안 테스트
        - GREEN: 고정 메시지 및 제안 리스트 반환
        - REFACTOR: 사용자 친화적 메시지 개선
    """
    suggestions = [
        "Use `/clear` to start a fresh session",
        "Use `/new` to begin a new conversation",
        "Consider summarizing key decisions before continuing",
    ]

    message = "💡 Tip: Context is getting large. Consider starting a new session for better performance."

    return HookResult(message=message, suggestions=suggestions)


__all__ = ["handle_pre_compact"]
