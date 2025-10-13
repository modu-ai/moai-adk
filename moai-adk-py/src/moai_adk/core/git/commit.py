# @CODE:CORE-GIT-001 | SPEC: SPEC-CORE-GIT-001.md | TEST: tests/unit/test_git.py
"""
커밋 메시지 포맷팅 유틸리티.

SPEC: .moai/specs/SPEC-CORE-GIT-001/spec.md
"""

from typing import Literal


def format_commit_message(
    stage: Literal["red", "green", "refactor", "docs"],
    description: str,
    locale: str = "ko",
) -> str:
    """
    TDD 단계별 커밋 메시지 생성.

    Args:
        stage: TDD 단계 (red, green, refactor, docs)
        description: 커밋 설명
        locale: 언어 코드 (ko, en, ja, zh)

    Returns:
        포맷팅된 커밋 메시지

    Examples:
        >>> format_commit_message("red", "사용자 인증 테스트 작성", "ko")
        '🔴 RED: 사용자 인증 테스트 작성'

        >>> format_commit_message("green", "Implement authentication", "en")
        '🟢 GREEN: Implement authentication'

        >>> format_commit_message("refactor", "코드 구조 개선", "ko")
        '♻️ REFACTOR: 코드 구조 개선'
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
        "ja": {
            "red": "🔴 RED: {desc}",
            "green": "🟢 GREEN: {desc}",
            "refactor": "♻️ REFACTOR: {desc}",
            "docs": "📝 DOCS: {desc}",
        },
        "zh": {
            "red": "🔴 RED: {desc}",
            "green": "🟢 GREEN: {desc}",
            "refactor": "♻️ REFACTOR: {desc}",
            "docs": "📝 DOCS: {desc}",
        },
    }

    template = templates.get(locale, templates["en"]).get(stage.lower())
    if not template:
        raise ValueError(f"Invalid stage: {stage}")

    return template.format(desc=description)
