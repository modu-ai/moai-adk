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

        # Import enhanced language configuration
        from ..language_config import LANGUAGE_CONFIG

        # Extract language names from enhanced config
        language_names = {code: info["native_name"] for code, info in LANGUAGE_CONFIG.items()}

        language_name = language_names.get(conversation_language, "English")

        # Create new nested language structure
        config["language"] = {
            "conversation_language": conversation_language,
            "conversation_language_name": language_name,
        }

    # 2. Language settings migration
    # Old: "language": "ko"
    # New: "language": {"conversation_language": "ko", "conversation_language_name": "Korean"}
    if "language" in config and isinstance(config["language"], str):
        old_lang = config["language"]
        lang_names = {
            "ko": "Korean",
            "en": "English",
            "ja": "Japanese",
            "zh": "Chinese",
        }
        config["language"] = {
            "conversation_language": old_lang,
            "conversation_language_name": lang_names.get(old_lang, "English"),
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

    # Import enhanced language configuration
    from ..language_config import LANGUAGE_CONFIG

    # Extract language names from enhanced config
    language_names = {code: info["native_name"] for code, info in LANGUAGE_CONFIG.items()}
    return language_names.get(language_code, "English")


def migrate_config_schema_v0_17_0(config: dict[str, Any]) -> dict[str, Any]:
    """Migrate config schema for v0.16.0 → v0.17.0 (report generation feature).

    Adds new sections:
    - report_generation: Control automatic report generation
    - Enhanced github: auto_delete_branches, spec_git_workflow settings

    This function is backward-compatible and safe for existing configs.

    Args:
        config: Configuration dictionary (may be v0.16.0 or earlier).

    Returns:
        Configuration dictionary with v0.17.0 schema.
    """
    # 1. Add report_generation section if missing (defaults to enabled=true, auto_create=false)
    if "report_generation" not in config:
        config["report_generation"] = {
            "enabled": True,
            "auto_create": False,
            "warn_user": True,
            "user_choice": "Minimal",
            "configured_at": None,  # Will be set when user configures
            "allowed_locations": [
                ".moai/docs/",
                ".moai/reports/",
                ".moai/analysis/",
                ".moai/specs/SPEC-*/",
            ],
            "notes": (
                "Control automatic report generation. 'enabled': turn on/off, "
                "'auto_create': full (true) vs minimal (false) reports. "
                "Helps reduce token usage."
            ),
        }

    # 2. Enhance github section with new fields
    if "github" not in config:
        config["github"] = {}

    github_config = config["github"]

    # Add auto_delete_branches settings if missing
    if "auto_delete_branches" not in github_config:
        github_config["auto_delete_branches"] = None
        github_config["auto_delete_branches_checked"] = False
        github_config["auto_delete_branches_rationale"] = "Not configured"

    # Add spec_git_workflow settings if missing
    if "spec_git_workflow" not in github_config:
        github_config["spec_git_workflow"] = "per_spec"
        github_config["spec_git_workflow_configured"] = False
        github_config["spec_git_workflow_rationale"] = "Ask per SPEC (flexible, user controls each workflow)"

    # Add notes for new fields if missing
    if "notes_new_fields" not in github_config:
        github_config["notes_new_fields"] = (
            "auto_delete_branches: whether to auto-delete feature branches after merge. "
            "spec_git_workflow: 'feature_branch' (auto), 'develop_direct' (direct), "
            "'per_spec' (ask per SPEC)"
        )

    return config


def get_report_generation_config(config: dict[str, Any]) -> dict[str, Any]:
    """Get report generation configuration with safe defaults.

    Args:
        config: Configuration dictionary.

    Returns:
        Report generation configuration with defaults.
    """
    default_config = {
        "enabled": True,
        "auto_create": False,
        "warn_user": True,
        "user_choice": "Minimal",
        "configured_at": None,
        "allowed_locations": [
            ".moai/docs/",
            ".moai/reports/",
            ".moai/analysis/",
            ".moai/specs/SPEC-*/",
        ],
    }

    report_gen = config.get("report_generation", {})
    if isinstance(report_gen, dict):
        # Merge with defaults to ensure all keys exist
        return {**default_config, **report_gen}

    return default_config


def get_spec_git_workflow(config: dict[str, Any]) -> str:
    """Get SPEC git workflow setting with safe default.

    Options:
    - 'per_spec': Ask per SPEC (flexible, user controls)
    - 'feature_branch': Auto-create branch for each SPEC
    - 'develop_direct': Direct commits to develop

    Args:
        config: Configuration dictionary.

    Returns:
        Workflow setting string.
    """
    github_config = config.get("github", {})
    if isinstance(github_config, dict):
        workflow = github_config.get("spec_git_workflow")
        if workflow in ["per_spec", "feature_branch", "develop_direct"]:
            return workflow

    # Default: per_spec (ask user)
    return "per_spec"
