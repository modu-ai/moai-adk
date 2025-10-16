#!/usr/bin/env python3
"""User interaction handlers

UserPromptSubmit 이벤트 처리
"""

from core import HookPayload, HookResult
from core.context import get_jit_context


def handle_user_prompt_submit(payload: HookPayload) -> HookResult:
    """UserPromptSubmit 이벤트 핸들러

    사용자 프롬프트를 분석하여 관련 문서를 자동으로 컨텍스트에 추가합니다.
    JIT (Just-in-Time) Retrieval 원칙에 따라 필요한 문서만 로드합니다.

    Args:
        payload: Claude Code 이벤트 페이로드
                 (userPrompt, cwd 키 포함)

    Returns:
        HookResult(
            message=로드된 파일 수 (또는 None),
            contextFiles=추천 문서 경로 리스트
        )

    TDD History:
        - RED: JIT 문서 로딩 시나리오 테스트
        - GREEN: get_jit_context() 호출하여 문서 추천
        - REFACTOR: 메시지 조건부 표시 (파일 있을 때만)
    """
    user_prompt = payload.get("userPrompt", "")
    cwd = payload.get("cwd", ".")
    context_files = get_jit_context(user_prompt, cwd)

    message = f"📎 Loaded {len(context_files)} context file(s)" if context_files else None

    return HookResult(message=message, contextFiles=context_files)


__all__ = ["handle_user_prompt_submit"]
