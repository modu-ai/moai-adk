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

    Context Engineering 원칙:
        - 토큰 사용량 > 70% (140,000/200,000) 시 권장
        - 대화 턴 수 > 50 시 권장
        - 긴 세션은 결정/제약/상태 중심으로 요약 후 재시작

    Suggestions:
        - /clear 명령으로 새 세션 시작
        - /new 명령으로 새 대화 시작
        - 핵심 결정사항 요약 후 계속

    Notes:
        - development-guide.md Context Engineering 섹션 기반
        - 성능 향상 및 컨텍스트 관리 개선

    TDD History:
        - RED: PreCompact 메시지 및 제안 테스트
        - GREEN: 한국어 메시지 및 구체적 안내 반환
        - REFACTOR: Context Engineering 원칙 명시 (v0.3.1)
    """
    # Payload에서 토큰 및 턴 정보 추출 시도 (있다면)
    # 현재는 기본 메시지 반환, 향후 payload 구조 확인 후 개선 가능

    suggestions = [
        "`/clear` 명령으로 현재 대화 초기화",
        "`/new` 명령으로 새 대화 시작",
        "핵심 결정사항을 요약한 후 새 세션에서 계속",
    ]

    message = (
        "📊 Context Engineering: Compaction 권장\n\n"
        "컨텍스트 사용량이 증가했습니다. "
        "더 나은 성능과 컨텍스트 관리를 위해 새 세션 시작을 권장합니다.\n\n"
        "**권장사항**: `/clear` 또는 `/new` 명령으로 새로운 대화 세션을 시작하세요."
    )

    return HookResult(message=message, suggestions=suggestions)


__all__ = ["handle_pre_compact"]
