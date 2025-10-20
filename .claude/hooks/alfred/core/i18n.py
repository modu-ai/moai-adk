#!/usr/bin/env python3
"""i18n module for hooks

Hooks용 i18n 모듈 (독립 구현)
"""

import json
from functools import lru_cache
from pathlib import Path

# Supported locales
SUPPORTED_LOCALES = {"ko", "en", "ja", "zh", "th"}


@lru_cache(maxsize=5)
def load_messages(locale: str = "ko") -> dict:
    """Load i18n messages for the specified locale.

    Args:
        locale: Language code (ko, en, ja, zh, th)

    Returns:
        Dictionary of translated messages

    Note:
        Uses LRU cache to avoid repeated file I/O
    """
    # Validate locale
    if locale not in SUPPORTED_LOCALES:
        locale = "ko"  # Fallback to Korean

    # Find docs/i18n/{locale}.json from hook location
    repo_root = Path(__file__).parent.parent.parent.parent
    message_file = repo_root / "docs" / "i18n" / f"{locale}.json"

    if not message_file.exists():
        # Fallback to English if locale file not found
        message_file = repo_root / "docs" / "i18n" / "en.json"

    if not message_file.exists():
        # Return empty dict if no message file found
        return {}

    try:
        return json.loads(message_file.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return {}


def t(key: str, locale: str = "ko", **kwargs) -> str:
    """Translate a message key.

    Args:
        key: Message key (e.g., "session_start")
        locale: Language code
        **kwargs: Format variables

    Returns:
        Translated and formatted message

    Examples:
        >>> t("session_start", "en")
        "🚀 MoAI-ADK Session Started"

        >>> t("checkpoint_created", "ko", name="before-merge")
        "🛡️ 체크포인트 생성: before-merge"
    """
    messages = load_messages(locale)
    message = messages.get(key, key)

    # Simple format
    if kwargs:
        try:
            return message.format(**kwargs)
        except (KeyError, ValueError):
            # Return unformatted message if formatting fails
            return message

    return message


def get_supported_locales() -> list[str]:
    """Get list of supported locales.

    Returns:
        List of locale codes
    """
    return list(SUPPORTED_LOCALES)


def validate_locale(locale: str) -> str:
    """Validate and normalize locale code.

    Args:
        locale: Locale code to validate

    Returns:
        Validated locale code (or "ko" if invalid)
    """
    if locale in SUPPORTED_LOCALES:
        return locale
    return "ko"  # Fallback to Korean


__all__ = ["t", "load_messages", "get_supported_locales", "validate_locale"]
