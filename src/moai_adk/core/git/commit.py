"""
Commit message formatting utilities.

SPEC: .moai/specs/SPEC-CORE-GIT-001/spec.md
"""

from typing import Literal


def format_commit_message(
    stage: Literal["red", "green", "refactor", "docs"],
    description: str,
    locale: str = "ko",
) -> str:
    """
    Generate a commit message for each TDD stage.

    Args:
        stage: TDD stage (red, green, refactor, docs).
        description: Commit description text.
        locale: Language code (ko, en, ja, zh).

    Returns:
        Formatted commit message.

    Examples:
        >>> format_commit_message("red", "Add failing authentication test", "ko")
        '🔴 RED: Add failing authentication test'

        >>> format_commit_message("green", "Implement authentication", "en")
        '🟢 GREEN: Implement authentication'

        >>> format_commit_message("refactor", "Improve code structure", "ko")
        '♻️ REFACTOR: Improve code structure'
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
