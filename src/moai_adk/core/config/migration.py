# @CODE:LANG-FIX-001:MIGRATION | SPEC: .moai/specs/SPEC-LANG-FIX-001/spec.md
"""Configuration migration utilities for legacy flat config structure.

Supports migration from legacy flat config.json structure to new nested language structure.
"""

from typing import Any


def migrate_config_to_nested_structure(config: dict[str, Any]) -> dict[str, Any]:
    """Migrate legacy flat config to nested language structure.

    This function handles the transition from legacy flat config:
        "conversation_language": "ko"
        "locale": "ko"

    To new nested structure:
        "language": {
            "conversation_language": "ko",
            "conversation_language_name": "한국어"
        }

    Args:
        config: Configuration dictionary that may have legacy structure.

    Returns:
        Configuration dictionary with nested language structure.
    """
    # If config already has nested language structure, return as-is
    if "language" in config and isinstance(config["language"], dict):
        return config

    # If config has legacy flat structure, migrate it
    if "conversation_language" in config and "language" not in config:
        # Extract conversation language from legacy location
        conversation_language = config.pop("conversation_language", "en")
        config.pop("locale", None)  # Remove legacy locale field

        # Map language codes to language names
        language_names = {
            "en": "English",
            "ko": "한국어",
            "ja": "日本語",
            "zh": "中文",
            "es": "Español",
        }

        language_name = language_names.get(conversation_language, "English")

        # Create new nested language structure
        config["language"] = {
            "conversation_language": conversation_language,
            "conversation_language_name": language_name,
        }

    return config


def get_conversation_language(config: dict[str, Any]) -> str:
    """Get conversation language from config with fallback handling.

    Handles both legacy flat and new nested config structures.

    Args:
        config: Configuration dictionary.

    Returns:
        Language code (e.g., "ko", "en", "ja").
    """
    # First, try to get from nested structure (new format)
    language_config = config.get("language", {})
    if isinstance(language_config, dict):
        result = language_config.get("conversation_language")
        if result:
            return result

    # Fall back to legacy flat structure
    result = config.get("conversation_language")
    if result:
        return result

    # Default to English
    return "en"


def get_conversation_language_name(config: dict[str, Any]) -> str:
    """Get conversation language name from config with fallback handling.

    Handles both legacy flat and new nested config structures.

    Args:
        config: Configuration dictionary.

    Returns:
        Language name (e.g., "한국어", "English").
    """
    # First, try to get from nested structure (new format)
    language_config = config.get("language", {})
    if isinstance(language_config, dict):
        result = language_config.get("conversation_language_name")
        if result:
            return result

    # If we have the language code, try to map it
    language_code = get_conversation_language(config)
    language_names = {
        "en": "English",
        "ko": "한국어",
        "ja": "日本語",
        "zh": "中文",
        "es": "Español",
    }
    return language_names.get(language_code, "English")
