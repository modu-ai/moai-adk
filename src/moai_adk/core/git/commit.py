# @CODE:CORE-GIT-001 | SPEC: SPEC-CORE-GIT-001.md | TEST: tests/unit/test_git_utils.py
"""커밋 메시지 포맷팅 함수"""

from typing import Literal

# TDD 단계 타입
TDDStage = Literal["red", "green", "refactor", "docs"]


def format_commit_message(stage: TDDStage | str, description: str, locale: str = "ko") -> str:
    """TDD 단계별 커밋 메시지 생성

    Args:
        stage: TDD 단계 (red|green|refactor|docs)
        description: 커밋 설명
        locale: 로케일 (ko|en), 기본값 "ko"

    Returns:
        이모지와 단계 접두사가 포함된 커밋 메시지

    Examples:
        >>> format_commit_message("red", "사용자 인증 테스트 작성", locale="ko")
        '🔴 RED: 사용자 인증 테스트 작성'
        >>> format_commit_message("green", "Add authentication", locale="en")
        '🟢 GREEN: Add authentication'

    Note:
        - 알 수 없는 로케일은 영어(en)로 폴백
        - 알 수 없는 단계는 None 반환 (KeyError 방지 필요)
    """
    templates = {
        "ko": {
            "red": "🔴 RED: {desc}",
            "green": "🟢 GREEN: {desc}",
            "refactor": "♻️ REFACTOR: {desc}",
            "docs": "📝 DOCS: {desc}",
        },
        "en": {
            "red": "🔴 RED: {desc}",
            "green": "🟢 GREEN: {desc}",
            "refactor": "♻️ REFACTOR: {desc}",
            "docs": "📝 DOCS: {desc}",
        },
    }

    template = templates.get(locale, templates["en"]).get(stage.lower())

    if template is None:
        raise ValueError(f"Unknown TDD stage: {stage}. Valid stages: red, green, refactor, docs")

    return template.format(desc=description)
