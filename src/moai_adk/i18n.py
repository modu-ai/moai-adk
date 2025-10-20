# @CODE:I18N-001 | SPEC: SPEC-I18N-001.md | TEST: tests/test_i18n.py
"""
i18n (Internationalization) module for MoAI-ADK.

Provides multilingual message support for 5 languages:
- ko: Korean (default)
- en: English
- ja: Japanese
- zh: Chinese (Simplified)
- th: Thai

다국어 메시지 지원 모듈 (5개 언어):
- ko: 한국어 (기본값)
- en: 영어
- ja: 일본어
- zh: 중국어 (간체)
- th: 태국어
"""

from functools import lru_cache
from pathlib import Path
import json
from typing import Any

# Supported locales
SUPPORTED_LOCALES = ["ko", "en", "ja", "zh", "th"]
DEFAULT_LOCALE = "ko"
FALLBACK_LOCALE = "en"


@lru_cache(maxsize=5)
def load_messages(locale: str = DEFAULT_LOCALE) -> dict[str, Any]:
    """
    Load i18n messages for the specified locale.

    지정된 locale의 i18n 메시지를 로드합니다.

    Args:
        locale: Language code (ko, en, ja, zh, th)

    Returns:
        Dictionary of translated messages

    Raises:
        FileNotFoundError: If message file not found
        json.JSONDecodeError: If JSON parsing fails

    Example:
        >>> messages = load_messages("en")
        >>> messages["session_start"]
        "🚀 MoAI-ADK Session Started"

    Note:
        - Uses LRU cache to improve performance
        - Falls back to English if locale not supported
        - LRU 캐시를 사용하여 성능 최적화
        - 지원하지 않는 locale은 영어로 대체
    """
    # Validate locale
    if locale not in SUPPORTED_LOCALES:
        locale = FALLBACK_LOCALE  # Fallback to English

    # Determine i18n directory
    # Priority: 1. Installed package, 2. Development environment
    try:
        # Try installed package location
        i18n_dir = Path(__file__).parent.parent.parent / "docs" / "i18n"
        if not i18n_dir.exists():
            # Fallback to package data (for pip install)
            import importlib.resources
            i18n_dir = Path(str(importlib.resources.files("moai_adk"))) / ".." / ".." / "docs" / "i18n"
    except Exception:
        # Last resort: relative to current file
        i18n_dir = Path(__file__).parent.parent.parent / "docs" / "i18n"

    message_file = i18n_dir / f"{locale}.json"

    if not message_file.exists():
        raise FileNotFoundError(
            f"Message file not found: {message_file}\n"
            f"Supported locales: {', '.join(SUPPORTED_LOCALES)}"
        )

    # Load and parse JSON
    try:
        return json.loads(message_file.read_text(encoding="utf-8"))
    except json.JSONDecodeError as e:
        raise json.JSONDecodeError(
            f"Failed to parse message file: {message_file}",
            e.doc,
            e.pos,
        )


def t(key: str, locale: str = DEFAULT_LOCALE, **kwargs: Any) -> str:
    """
    Translate a message key.

    메시지 키를 번역합니다.

    Args:
        key: Message key (e.g., "session_start", "error.no_git")
        locale: Language code (ko, en, ja, zh, th)
        **kwargs: Format variables for message interpolation

    Returns:
        Translated and formatted message

    Example:
        >>> t("session_start", "en")
        "🚀 MoAI-ADK Session Started"

        >>> t("checkpoint_created", "ko", name="before-merge")
        "🛡️ 체크포인트 생성: before-merge"

        >>> t("context_loaded", "ja", count=3)
        "📎 3個のコンテキストファイルを読み込みました"

    Note:
        - Supports nested keys with dot notation (e.g., "error.no_git")
        - Returns key itself if translation not found
        - 점 표기법으로 중첩 키 지원 (예: "error.no_git")
        - 번역을 찾지 못하면 키 자체를 반환
    """
    try:
        messages = load_messages(locale)
    except (FileNotFoundError, json.JSONDecodeError):
        # Fallback to default locale
        try:
            messages = load_messages(DEFAULT_LOCALE)
        except Exception:
            # Last resort: return key itself
            return key

    # Navigate nested keys (e.g., "error.no_git" → messages["error"]["no_git"])
    value: Any = messages
    for part in key.split("."):
        if isinstance(value, dict) and part in value:
            value = value[part]
        else:
            # Key not found, return key itself
            return key

    # Ensure we got a string
    if not isinstance(value, str):
        return key

    # Format with variables
    if kwargs:
        try:
            return value.format(**kwargs)
        except (KeyError, ValueError):
            # Format error, return unformatted message
            return value

    return value


def get_supported_locales() -> list[str]:
    """
    Get list of supported locales.

    지원되는 locale 목록을 반환합니다.

    Returns:
        List of locale codes

    Example:
        >>> get_supported_locales()
        ['ko', 'en', 'ja', 'zh', 'th']
    """
    return SUPPORTED_LOCALES.copy()


def get_locale_name(locale: str) -> str:
    """
    Get human-readable name for a locale.

    locale의 사람이 읽을 수 있는 이름을 반환합니다.

    Args:
        locale: Language code

    Returns:
        Locale name in English and native language

    Example:
        >>> get_locale_name("ko")
        "Korean (한국어)"

        >>> get_locale_name("ja")
        "Japanese (日本語)"
    """
    locale_names = {
        "ko": "Korean (한국어)",
        "en": "English",
        "ja": "Japanese (日本語)",
        "zh": "Chinese (中文)",
        "th": "Thai (ไทย)",
    }
    return locale_names.get(locale, locale)


def validate_locale(locale: str) -> str:
    """
    Validate and normalize a locale code.

    locale 코드를 검증하고 정규화합니다.

    Args:
        locale: Language code to validate

    Returns:
        Validated locale code (or fallback)

    Example:
        >>> validate_locale("en")
        "en"

        >>> validate_locale("fr")  # Not supported
        "en"

        >>> validate_locale("EN")  # Uppercase
        "en"

    Note:
        - Returns fallback locale (en) if not supported
        - Converts to lowercase
        - 지원하지 않으면 fallback locale (en) 반환
        - 소문자로 변환
    """
    normalized = locale.lower().strip()
    if normalized in SUPPORTED_LOCALES:
        return normalized
    return FALLBACK_LOCALE
