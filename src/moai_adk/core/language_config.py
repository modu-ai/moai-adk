"""
Enhanced language configuration for MoAI-ADK.

Supports extended language list with native names and automatic translation capabilities.
"""

from typing import Dict, Optional

# Enhanced language configuration with native names
LANGUAGE_CONFIG: Dict[str, Dict[str, str]] = {
    "en": {
        "name": "English",
        "native_name": "English",
        "code": "en",
        "family": "indo-european",
    },
    "ko": {
        "name": "Korean",
        "native_name": "한국어",
        "code": "ko",
        "family": "koreanic",
    },
    "ja": {
        "name": "Japanese",
        "native_name": "日本語",
        "code": "ja",
        "family": "japonic",
    },
    "es": {
        "name": "Spanish",
        "native_name": "Español",  # Spanish native representation
        "code": "es",
        "family": "indo-european",
    },
    "fr": {
        "name": "French",
        "native_name": "Français",
        "code": "fr",
        "family": "indo-european",
    },
    "de": {
        "name": "German",
        "native_name": "Deutsch",
        "code": "de",
        "family": "indo-european",
    },
    "zh": {
        "name": "Chinese",
        "native_name": "中文",
        "code": "zh",
        "family": "sino-tibetan",
    },
    "pt": {
        "name": "Portuguese",
        "native_name": "Português",
        "code": "pt",
        "family": "indo-european",
    },
    "ru": {
        "name": "Russian",
        "native_name": "Русский",
        "code": "ru",
        "family": "indo-european",
    },
    "it": {
        "name": "Italian",
        "native_name": "Italiano",
        "code": "it",
        "family": "indo-european",
    },
    "ar": {
        "name": "Arabic",
        "native_name": "العربية",
        "code": "ar",
        "family": "afro-asiatic",
    },
    "hi": {
        "name": "Hindi",
        "native_name": "हिन्दी",
        "code": "hi",
        "family": "indo-european",
    },
}


def get_language_info(language_code: str) -> Optional[Dict[str, str]]:
    """Get language information by code.

    Args:
        language_code: ISO language code (e.g., 'ko', 'en')

    Returns:
        Language information dictionary or None if not found
    """
    return LANGUAGE_CONFIG.get(language_code.lower())


def get_native_name(language_code: str) -> str:
    """Get native language name.

    Args:
        language_code: ISO language code

    Returns:
        Native language name or English fallback
    """
    lang_info = get_language_info(language_code)
    return lang_info["native_name"] if lang_info else "English"


def get_english_name(language_code: str) -> str:
    """Get English language name.

    Args:
        language_code: ISO language code

    Returns:
        English language name or fallback
    """
    lang_info = get_language_info(language_code)
    return lang_info["name"] if lang_info else "English"


def get_all_supported_codes() -> list[str]:
    """Get list of all supported language codes."""
    return list(LANGUAGE_CONFIG.keys())


def get_language_family(language_code: str) -> Optional[str]:
    """Get language family for linguistic analysis.

    Args:
        language_code: ISO language code

    Returns:
        Language family string or None
    """
    lang_info = get_language_info(language_code)
    return lang_info.get("family") if lang_info else None


# Language to Claude model mapping for optimal performance
LANGUAGE_MODEL_PREFERENCE: Dict[str, str] = {
    "en": "claude-sonnet-4-5-20250929",  # Best for English
    "ko": "claude-sonnet-4-5-20250929",  # Strong Korean support
    "ja": "claude-sonnet-4-5-20250929",  # Strong Japanese support
    "es": "claude-sonnet-4-5-20250929",  # Strong Spanish support
    "fr": "claude-sonnet-4-5-20250929",  # Strong French support
    "de": "claude-sonnet-4-5-20250929",  # Strong German support
    "zh": "claude-sonnet-4-5-20250929",  # Strong Chinese support
    "pt": "claude-sonnet-4-5-20250929",  # Portuguese support
    "ru": "claude-sonnet-4-5-20250929",  # Russian support
    "it": "claude-sonnet-4-5-20250929",  # Italian support
    "ar": "claude-sonnet-4-5-20250929",  # Arabic support
    "hi": "claude-sonnet-4-5-20250929",  # Hindi support
}


def get_optimal_model(language_code: str) -> str:
    """Get optimal Claude model for specific language.

    Args:
        language_code: ISO language code

    Returns:
        Recommended Claude model identifier
    """
    return LANGUAGE_MODEL_PREFERENCE.get(language_code, "claude-sonnet-4-5-20250929")


# RTL (Right-to-Left) language detection
RTL_LANGUAGES = {"ar", "he", "fa", "ur"}


def is_rtl_language(language_code: str) -> bool:
    """Check if language uses right-to-left script.

    Args:
        language_code: ISO language code

    Returns:
        True if RTL language, False otherwise
    """
    return language_code.lower() in RTL_LANGUAGES


# Translation priorities for descriptions
TRANSLATION_PRIORITY = [
    "en",  # English base
    "ko",  # Korean
    "ja",  # Japanese
    "es",  # Spanish
    "fr",  # French
    "de",  # German
    "zh",  # Chinese
    "pt",  # Portuguese
]


def get_translation_priority() -> list[str]:
    """Get language translation priority order."""
    return TRANSLATION_PRIORITY.copy()
