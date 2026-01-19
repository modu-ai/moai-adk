"""Additional coverage tests for announcement translator.

Tests for lines not covered by existing tests.
"""

import json
import sys
from pathlib import Path

# The module is in the templates directory
sys.path.insert(
    0,
    str(
        Path(__file__).parent.parent.parent.parent.parent.parent.parent
        / "src"
        / "moai_adk"
        / "templates"
        / ".claude"
        / "hooks"
        / "moai"
        / "shared"
        / "utils"
    ),
)

import announcement_translator


class TestGetLanguageFromConfig:
    """Test get_language_from_config function."""

    def test_returns_default_for_unsupported_language(self, tmp_path):
        """Should return DEFAULT_LANGUAGE when language not supported."""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.yaml"
        config_file.write_text("language:\n  conversation_language: unsupported_lang")

        language = announcement_translator.get_language_from_config(tmp_path)
        assert language == announcement_translator.DEFAULT_LANGUAGE

    def test_returns_default_on_exception(self, tmp_path):
        """Should return DEFAULT_LANGUAGE when exception occurs."""
        # Create invalid YAML
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.yaml"
        config_file.write_text("{ invalid: yaml: content")

        language = announcement_translator.get_language_from_config(tmp_path)
        assert language == announcement_translator.DEFAULT_LANGUAGE


class TestLoadAnnouncementsFromFile:
    """Test load_announcements_from_file function."""

    def test_returns_empty_on_exception(self, tmp_path):
        """Should return empty list on exception."""
        # Create invalid JSON file
        invalid_file = tmp_path / "invalid.json"
        invalid_file.write_text("{ invalid json }")

        announcements = announcement_translator.load_announcements_from_file(invalid_file)
        assert announcements == []

    def test_returns_empty_list_when_missing_field(self, tmp_path):
        """Should return empty list when companyAnnouncements field missing."""
        # Create JSON without companyAnnouncements field
        invalid_file = tmp_path / "incomplete.json"
        invalid_file.write_text('{"otherField": "value"}')

        announcements = announcement_translator.load_announcements_from_file(invalid_file)
        assert announcements == []


class TestTranslateAnnouncements:
    """Test translate_announcements function."""

    def test_fallback_to_english(self, tmp_path):
        """Should fallback to English when requested language file not found."""
        announcements_dir = tmp_path / ".moai" / "announcements"
        announcements_dir.mkdir(parents=True)

        # Create English announcements but not Korean
        english_file = announcements_dir / "en.json"
        english_file.write_text('{"companyAnnouncements": ["English message"]}')

        announcements = announcement_translator.translate_announcements("ko", tmp_path)
        assert announcements == ["English message"]

    def test_fallback_to_default_when_no_files(self, tmp_path):
        """Should fallback to default announcements when no files exist."""
        # Create empty announcements directory
        announcements_dir = tmp_path / ".moai" / "announcements"
        announcements_dir.mkdir(parents=True)

        announcements = announcement_translator.translate_announcements("ko", tmp_path)
        # Should return default announcements
        assert isinstance(announcements, list)


class TestUpdateSettingsAnnouncements:
    """Test update_settings_announcements function."""

    def test_returns_false_when_settings_file_not_exists(self, tmp_path):
        """Should return False when settings.local.json doesn't exist."""
        result = announcement_translator.update_settings_announcements(tmp_path)
        assert result is False

    def test_returns_false_on_exception(self, tmp_path):
        """Should return False when exception occurs."""
        # Create settings file with invalid JSON
        settings_dir = tmp_path / ".claude"
        settings_dir.mkdir(parents=True)
        settings_file = settings_dir / "settings.local.json"
        settings_file.write_text("{ invalid json }")

        result = announcement_translator.update_settings_announcements(tmp_path)
        assert result is False

    def test_successfully_updates_settings(self, tmp_path):
        """Should successfully update settings with announcements."""
        # Create settings file
        settings_dir = tmp_path / ".claude"
        settings_dir.mkdir(parents=True)
        settings_file = settings_dir / "settings.local.json"
        settings_file.write_text('{"existing": "data"}')

        # Create announcements
        announcements_dir = tmp_path / ".moai" / "announcements"
        announcements_dir.mkdir(parents=True)
        english_file = announcements_dir / "en.json"
        english_file.write_text('{"companyAnnouncements": ["Test announcement"]}')

        # Create config with language
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.yaml"
        config_file.write_text("language:\n  conversation_language: en")

        result = announcement_translator.update_settings_announcements(tmp_path)
        assert result is True

        # Verify settings were updated
        with open(settings_file, encoding="utf-8") as f:
            updated_settings = json.load(f)
        assert updated_settings["companyAnnouncements"] == ["Test announcement"]


class TestMainBlock:
    """Test __main__ block."""

    def test_main_with_path_argument(self, tmp_path, monkeypatch, capsys):
        """Should work with path argument."""
        # Set up test project
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.yaml"
        config_file.write_text("language:\n  conversation_language: en")

        announcements_dir = tmp_path / ".moai" / "announcements"
        announcements_dir.mkdir(parents=True)
        english_file = announcements_dir / "en.json"
        english_file.write_text('{"companyAnnouncements": ["Test announcement"]}')

        monkeypatch.setattr("sys.argv", ["announcement_translator.py", str(tmp_path)])

        # Run the __main__ block manually
        test_path = tmp_path
        lang = announcement_translator.get_language_from_config(test_path)
        assert lang == "en"

        announcements = announcement_translator.translate_announcements(lang, test_path)
        assert announcements == ["Test announcement"]

    def test_main_without_path_argument(self, tmp_path, monkeypatch, capsys):
        """Should use current directory when no path argument provided."""
        monkeypatch.chdir(tmp_path)

        # Set up test project
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.yaml"
        config_file.write_text("language:\n  conversation_language: en")

        announcements_dir = tmp_path / ".moai" / "announcements"
        announcements_dir.mkdir(parents=True)
        english_file = announcements_dir / "en.json"
        english_file.write_text('{"companyAnnouncements": ["Test announcement"]}')

        # Run the __main__ block manually using cwd
        test_path = Path.cwd()
        lang = announcement_translator.get_language_from_config(test_path)
        assert lang == "en"

        announcements = announcement_translator.translate_announcements(lang, test_path)
        assert announcements == ["Test announcement"]
