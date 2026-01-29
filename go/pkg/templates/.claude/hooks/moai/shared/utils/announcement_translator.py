#!/usr/bin/env python3
"""Announcement translator for multi-language companyAnnouncements support.

This module loads language-specific announcements from .moai/announcements/{lang}.json
based on the conversation_language setting in language.yaml.

Supported languages: ko, en, ja, zh
Fallback: en (English)
"""

from __future__ import annotations

import json

# Import atomic_write from lib directory (go up two levels from shared/utils/)
import sys
from pathlib import Path
from pathlib import Path as _Path

_lib_path = _Path(__file__).parent.parent.parent / "lib"
if str(_lib_path) not in sys.path:
    sys.path.insert(0, str(_lib_path))
from atomic_write import atomic_write_json  # noqa: E402

# Default language if not configured
DEFAULT_LANGUAGE = "en"

# Supported languages
SUPPORTED_LANGUAGES = {"ko", "en", "ja", "zh"}


def get_language_from_config(project_path: Path | str) -> str:
    """Get conversation_language from language.yaml config.

    Args:
        project_path: Path to the project root directory

    Returns:
        Language code (ko, en, ja, zh). Defaults to 'en' if not found.
    """
    project_path = Path(project_path)
    language_file = project_path / ".moai" / "config" / "sections" / "language.yaml"

    if not language_file.exists():
        return DEFAULT_LANGUAGE

    try:
        import yaml

        with open(language_file, encoding="utf-8", errors="replace") as f:
            config = yaml.safe_load(f)

        language = config.get("language", {}).get("conversation_language", DEFAULT_LANGUAGE)

        # Validate language code
        if language not in SUPPORTED_LANGUAGES:
            return DEFAULT_LANGUAGE

        return language

    except Exception:
        return DEFAULT_LANGUAGE


def get_announcements_path(project_path: Path | str) -> Path:
    """Get the path to announcements directory.

    Args:
        project_path: Path to the project root directory

    Returns:
        Path to .moai/announcements directory
    """
    project_path = Path(project_path)
    return project_path / ".moai" / "announcements"


def load_announcements_from_file(announcements_file: Path) -> list[str]:
    """Load announcements from a JSON file.

    Args:
        announcements_file: Path to the announcements JSON file

    Returns:
        List of announcement strings
    """
    if not announcements_file.exists():
        return []

    try:
        with open(announcements_file, encoding="utf-8", errors="replace") as f:
            data = json.load(f)
        return data.get("companyAnnouncements", [])
    except Exception:
        return []


def translate_announcements(language: str, project_path: Path | str) -> list[str]:
    """Get announcements for the specified language.

    Args:
        language: Language code (ko, en, ja, zh)
        project_path: Path to the project root directory

    Returns:
        List of announcement strings in the specified language.
        Falls back to English if language file not found.
    """
    project_path = Path(project_path)
    announcements_dir = get_announcements_path(project_path)

    # Try requested language first
    language_file = announcements_dir / f"{language}.json"
    announcements = load_announcements_from_file(language_file)

    if announcements:
        return announcements

    # Fallback to English
    if language != DEFAULT_LANGUAGE:
        english_file = announcements_dir / f"{DEFAULT_LANGUAGE}.json"
        announcements = load_announcements_from_file(english_file)
        if announcements:
            return announcements

    # Final fallback: return default announcements
    return get_default_announcements()


def get_default_announcements() -> list[str]:
    """Get default English announcements as fallback.

    Returns:
        List of default announcement strings in English
    """
    return [
        "🗿 MoAI-ADK: SPEC-First DDD with 48 Skills and Context7 integration",
        "⚡ /moai:alfred: One-stop Plan→Run→Sync automation with intelligent routing",
        "🌳 moai-worktree: Parallel SPEC development - work on multiple features simultaneously",
        "🤖 20 Agents: 8 Expert + 8 Manager + 4 Builder for specialized tasks",
        "📋 Workflow: /moai:1-plan (SPEC) → /moai:2-run (DDD) → /moai:3-sync (Docs)",
        "✅ Quality: TRUST 5 + ≥85% coverage + Ralph Engine (LSP + AST-grep)",
        "📚 Tip: moai update --templates-only syncs latest skills and agents",
        "🏆 moai rank: Track your Claude token usage on rank.mo.ai.kr",
    ]


def update_settings_announcements(project_path: Path | str) -> bool:
    """Update companyAnnouncements in settings.local.json based on language config.

    This function is called during 'moai update' to refresh announcements
    according to the current language setting.

    Args:
        project_path: Path to the project root directory

    Returns:
        True if update successful, False otherwise
    """
    project_path = Path(project_path)
    settings_file = project_path / ".claude" / "settings.local.json"

    if not settings_file.exists():
        return False

    try:
        # Load current settings
        with open(settings_file, encoding="utf-8", errors="replace") as f:
            settings = json.load(f)

        # Get language and announcements
        language = get_language_from_config(project_path)
        announcements = translate_announcements(language, project_path)

        # Update announcements
        settings["companyAnnouncements"] = announcements

        # Write back using atomic write (H3)
        atomic_write_json(settings_file, settings, indent=2, ensure_ascii=False)

        return True

    except Exception:
        return False


# Alias for backward compatibility with update.py
auto_translate_and_update = update_settings_announcements


if __name__ == "__main__":
    # Test the module
    import sys

    if len(sys.argv) > 1:
        test_path = Path(sys.argv[1])
    else:
        test_path = Path.cwd()

    lang = get_language_from_config(test_path)
    print(f"Detected language: {lang}")

    announcements = translate_announcements(lang, test_path)
    print(f"Announcements ({len(announcements)} items):")
    for ann in announcements:
        print(f"  - {ann}")
