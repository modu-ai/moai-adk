#!/usr/bin/env python3
"""Core module for Alfred Hooks

공통 타입 정의 및 유틸리티 함수
"""

from dataclasses import asdict, dataclass, field
from typing import Any, NotRequired, TypedDict


class HookPayload(TypedDict):
    """Claude Code Hook 이벤트 페이로드 타입 정의

    Claude Code가 Hook 스크립트에 전달하는 데이터 구조.
    이벤트에 따라 필드가 다를 수 있으므로 NotRequired 사용.
    """

    cwd: str
    userPrompt: NotRequired[str]  # UserPromptSubmit 이벤트만 포함
    tool: NotRequired[str]  # PreToolUse/PostToolUse 이벤트
    arguments: NotRequired[dict[str, Any]]  # Tool arguments


@dataclass
class HookResult:
    """Hook 실행 결과"""

    message: str | None = None
    systemMessage: str | None = None  # 사용자에게 직접 표시되는 메시지
    blocked: bool = False
    contextFiles: list[str] = field(default_factory=list)
    suggestions: list[str] = field(default_factory=list)
    exitCode: int = 0

    def to_dict(self) -> dict[str, Any]:
        """일반 Hook용 딕셔너리 변환"""
        return asdict(self)

    def to_user_prompt_submit_dict(self) -> dict[str, Any]:
        """UserPromptSubmit Hook 전용 출력 형식

        Claude Code는 UserPromptSubmit에 대해 특별한 스키마를 요구:
        {
            "hookEventName": "UserPromptSubmit",
            "additionalContext": "string (required)"
        }

        Returns:
            Claude Code UserPromptSubmit Hook 스키마에 맞는 딕셔너리

        Examples:
            >>> result = HookResult(contextFiles=["tests/"])
            >>> result.to_user_prompt_submit_dict()
            {'hookEventName': 'UserPromptSubmit', 'additionalContext': '📎 Context: tests/'}
        """
        # contextFiles를 additionalContext 문자열로 변환
        if self.contextFiles:
            context_str = "\n".join([f"📎 Context: {f}" for f in self.contextFiles])
        else:
            context_str = ""

        # message가 있으면 추가
        if self.message:
            if context_str:
                context_str = f"{self.message}\n\n{context_str}"
            else:
                context_str = self.message

        # 빈 문자열이면 기본값 사용
        if not context_str:
            context_str = ""

        return {
            "hookEventName": "UserPromptSubmit",
            "additionalContext": context_str
        }


__all__ = ["HookPayload", "HookResult"]

# Note: core module exports:
# - HookPayload, HookResult (type definitions)
# - project.py: detect_language, get_git_info, count_specs, get_project_language
# - context.py: get_jit_context
# - checkpoint.py: detect_risky_operation, create_checkpoint, log_checkpoint, list_checkpoints
