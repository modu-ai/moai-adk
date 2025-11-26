"""
Unit tests for announcement_translator with Haiku dynamic translation support.

Tests coverage:
1. Hardcoded language retrieval (ko, en, ja, zh) - no Claude cost
2. Dynamic translation with Haiku (de, fr, es, etc.)
3. Fallback to English on translation failure
4. Edge cases (invalid language, timeout, Claude CLI not found)
"""

import json
import subprocess

# Import the module under test
import sys
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

lib_path = Path(__file__).parent.parent.parent / ".claude" / "hooks" / "moai" / "lib"
if str(lib_path) not in sys.path:
    sys.path.insert(0, str(lib_path))

try:
    from announcement_translator import (
        ANNOUNCEMENTS_JA,
        ANNOUNCEMENTS_KO,
        ANNOUNCEMENTS_ZH,
        HARDCODED_TRANSLATIONS,
        REFERENCE_ANNOUNCEMENTS_EN,
        auto_translate_and_update,
        copy_settings_to_local,
        get_language_from_config,
        translate_announcements,
        translate_with_haiku,
    )
except ImportError as e:
    # Fallback if module not found (for testing environments)
    pytest.skip(f"announcement_translator module not found: {e}", allow_module_level=True)


class TestHardcodedLanguages:
    """Test retrieval of hardcoded translations (fast path, no Claude cost)."""

    def test_korean_hardcoded(self):
        """Korean translations should return immediately without Claude."""
        result = translate_announcements("ko")
        assert result == ANNOUNCEMENTS_KO
        assert len(result) == 23

    def test_english_hardcoded(self):
        """English translations should return immediately without Claude."""
        result = translate_announcements("en")
        assert result == REFERENCE_ANNOUNCEMENTS_EN
        assert len(result) == 23

    def test_japanese_hardcoded(self):
        """Japanese translations should return immediately without Claude."""
        result = translate_announcements("ja")
        assert result == ANNOUNCEMENTS_JA
        assert len(result) == 23

    def test_chinese_hardcoded(self):
        """Chinese translations should return immediately without Claude."""
        result = translate_announcements("zh")
        assert result == ANNOUNCEMENTS_ZH
        assert len(result) == 23

    def test_all_hardcoded_have_23_items(self):
        """All hardcoded translations should have exactly 23 items."""
        for lang, translations in HARDCODED_TRANSLATIONS.items():
            assert len(translations) == 23, f"Language {lang} has {len(translations)} items, expected 23"


class TestDynamicTranslationWithHaiku:
    """Test dynamic translation using Claude Haiku for non-hardcoded languages."""

    @patch("subprocess.run")
    def test_german_translation_success(self, mock_run):
        """German translation should call Claude Haiku and parse response."""
        # Mock Claude Haiku response
        mock_response = MagicMock()
        mock_response.returncode = 0
        mock_response.stdout = """1. SPEC zuerst: Verwenden Sie /moai:1-plan, um Anforderungen zuerst zu definieren...
2. TDD Zyklus: RED (Tests schreiben) → GREEN (minimaler Code) → REFACTOR (verbessern)...
3. 4-Phasen-Workflow: /moai:1-plan planen → /moai:2-run implementieren → /moai:3-sync validieren → git bereitstellen
4. TRUST 5 Prinzipien: Test(≥85%) + Lesbar + Einheitlich + Secured(OWASP) + Verfolgbar = Qualität
5. moai Agenten: 19 Expertenmitglieder verwalten automatisch Planung, Implementierung, Tests, Validierung
6. Konversationsstil: /output-style r2d2 (praktischer Partner) oder /output-style yoda (tiefes Lernen)
7. Aufgabenverfolgung: TodoWrite hält den Fortschritt sichtbar und verhindert übersehene Aufgaben
8. 55+ Skill Bibliothek: Verifizierte Muster und Best Practices ermöglichen halluzinationsfreie Implementierung
9. Automatische Validierung: Nach dem Codieren, TRUST 5, Testabdeckung und Sicherheitsprüfungen
10. GitFlow Strategie: feature/SPEC-XXX → develop → main Struktur für sichere Bereitstellungspipeline
11. Context7 Updates: Alle Bibliotheksversionen bleiben aktuell mit Echtzeit-API-Informationen
12. Parallele Verarbeitung: Unabhängige Aufgaben (Tests, Docs, Bereitstellung) laufen gleichzeitig
13. Gesundheitsprüfung: moai-adk doctor diagnostiziert Konfigurations-, Versions- und Abhängigkeitsprobleme
14. Dokumentsynchronisierung: /moai:3-sync auto synchronisiert Tests, Code und Dokumentation automatisch
15. Sprachentrennung: Kommunizieren Sie in Ihrer Sprache, schreiben Sie Code in Projektsprache
16. Mehrsprachige Unterstützung: Koreanisch, Englisch, Japanisch, Spanisch, und 25+ Sprachen
17. Sichere Updates: moai-adk update fügt neue Funktionen hinzu, während bestehende Einstellungen erhalten bleiben
18. Automatische Bereinigung: Sitzungsende bereinigt automatisch .moai/temp/ Cache und Protokolle
19. Fehlerwiederherstellung: Fehlgeschlagene Commits und Merge-Konflikte werden automatisch analysiert
20. Sicherheit an erster Stelle: Umgebungsvariablen, API-Schlüssel, Anmeldedaten werden automatisch in .gitignore hinzugefügt
21. Kontextoptimierung: Effiziente Nutzung des 200K-Token-Fensters ermöglicht die Verarbeitung großer Projekte
22. Schnelles Feedback: Fragen? Fragen Sie sofort - moai interpretiert Absichten und klärt automatisch
23. Sauberer Ausstieg: Beenden Sie Sitzungen mit /clear, um den Kontext zurückzusetzen"""
        mock_run.return_value = mock_response

        result = translate_announcements("de")

        # Verify Claude was called with correct parameters
        mock_run.assert_called_once()
        args, kwargs = mock_run.call_args
        assert args[0] == ["claude", "-p", "--model", "haiku"]
        assert kwargs["timeout"] == 30

        # Verify we got translated list
        assert len(result) == 23
        assert result[0].startswith("SPEC zuerst")

    @patch("subprocess.run")
    def test_french_translation_success(self, mock_run):
        """French translation should work with Haiku."""
        mock_response = MagicMock()
        mock_response.returncode = 0
        # Create 23 numbered translations
        french_items = [f"{i+1}. Traduction française {i+1}" for i in range(23)]
        mock_response.stdout = "\n".join(french_items)
        mock_run.return_value = mock_response

        result = translate_announcements("fr")
        assert len(result) == 23
        assert result[0] == "Traduction française 1"

    @patch("announcement_translator.translate_with_haiku")
    def test_spanish_fallback_on_failure(self, mock_haiku):
        """Spanish should fall back to English if Haiku translation fails."""
        mock_haiku.return_value = None  # Simulate translation failure

        result = translate_announcements("es")

        # Should return English fallback
        assert result == REFERENCE_ANNOUNCEMENTS_EN


class TestTranslationFailures:
    """Test error handling and fallback behavior."""

    @patch("subprocess.run")
    def test_claude_timeout(self, mock_run):
        """Translation should fall back on Claude timeout."""
        mock_run.side_effect = subprocess.TimeoutExpired("claude", 30)

        result = translate_announcements("de")

        # Should fall back to English
        assert result == REFERENCE_ANNOUNCEMENTS_EN

    @patch("subprocess.run")
    def test_claude_cli_not_found(self, mock_run):
        """Translation should fall back if Claude CLI not installed."""
        mock_run.side_effect = FileNotFoundError("claude command not found")

        result = translate_announcements("de")

        # Should fall back to English
        assert result == REFERENCE_ANNOUNCEMENTS_EN

    @patch("subprocess.run")
    def test_invalid_haiku_response(self, mock_run):
        """Translation should fail gracefully on invalid response."""
        mock_response = MagicMock()
        mock_response.returncode = 1
        mock_response.stderr = "Error in translation"
        mock_run.return_value = mock_response

        result = translate_announcements("de")

        # Should fall back to English
        assert result == REFERENCE_ANNOUNCEMENTS_EN

    @patch("subprocess.run")
    def test_incomplete_translation(self, mock_run):
        """Should fall back if only partial translations received."""
        mock_response = MagicMock()
        mock_response.returncode = 0
        # Only 10 items instead of 23
        mock_response.stdout = "\n".join([f"{i+1}. Traduction {i+1}" for i in range(10)])
        mock_run.return_value = mock_response

        result = translate_announcements("de")

        # Should fall back to English because we got only 10 items
        assert result == REFERENCE_ANNOUNCEMENTS_EN


class TestLanguageDetection:
    """Test language detection from config."""

    def test_detect_korean_from_config(self, tmp_path):
        """Should detect Korean from config file."""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"
        config_file.write_text(json.dumps({"language": {"conversation_language": "ko"}}))

        lang = get_language_from_config(tmp_path)
        assert lang == "ko"

    def test_detect_german_from_config(self, tmp_path):
        """Should detect German from config file."""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"
        config_file.write_text(json.dumps({"language": {"conversation_language": "de"}}))

        lang = get_language_from_config(tmp_path)
        assert lang == "de"

    def test_default_to_english_if_no_config(self, tmp_path):
        """Should default to English if no config found."""
        lang = get_language_from_config(tmp_path)
        assert lang == "en"


class TestSettingsCopy:
    """Test copying settings.json to settings.local.json."""

    def test_copy_with_korean_announcements(self, tmp_path):
        """Should copy settings with Korean announcements."""
        # Create .claude directory
        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()

        # Create settings.json
        settings_file = claude_dir / "settings.json"
        settings_file.write_text(json.dumps({"hooks": {}, "companyAnnouncements": []}))

        # Copy settings with Korean translations
        copy_settings_to_local("ko", ANNOUNCEMENTS_KO, tmp_path)

        # Verify settings.local.json was created
        local_file = claude_dir / "settings.local.json"
        assert local_file.exists()

        # Verify content
        content = json.loads(local_file.read_text())
        assert content["companyAnnouncements"] == ANNOUNCEMENTS_KO


class TestIntegration:
    """Integration tests for full workflow."""

    @patch("announcement_translator.translate_with_haiku")
    def test_full_workflow_hardcoded_language(self, mock_haiku, tmp_path):
        """Full workflow with hardcoded language should not call Haiku."""
        # Setup
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        (config_dir / "config.json").write_text(json.dumps({"language": {"conversation_language": "ko"}}))

        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        (claude_dir / "settings.json").write_text(json.dumps({"hooks": {}}))

        # Run auto update
        auto_translate_and_update(tmp_path)

        # Haiku should not be called for hardcoded language
        mock_haiku.assert_not_called()

        # Verify settings.local.json has Korean announcements
        local_file = claude_dir / "settings.local.json"
        assert local_file.exists()
        content = json.loads(local_file.read_text())
        assert content["companyAnnouncements"] == ANNOUNCEMENTS_KO

    @patch("subprocess.run")
    def test_full_workflow_dynamic_language(self, mock_run, tmp_path):
        """Full workflow with dynamic language should call Haiku."""
        # Setup
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        (config_dir / "config.json").write_text(json.dumps({"language": {"conversation_language": "de"}}))

        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        (claude_dir / "settings.json").write_text(json.dumps({"hooks": {}}))

        # Mock Haiku response
        mock_response = MagicMock()
        mock_response.returncode = 0
        mock_response.stdout = "\n".join([f"{i+1}. Deutsch {i+1}" for i in range(23)])
        mock_run.return_value = mock_response

        # Run auto update
        auto_translate_and_update(tmp_path)

        # Verify Haiku was called
        mock_run.assert_called_once()

        # Verify settings.local.json has German announcements
        local_file = claude_dir / "settings.local.json"
        assert local_file.exists()
        content = json.loads(local_file.read_text())
        assert len(content["companyAnnouncements"]) == 23
        assert content["companyAnnouncements"][0].startswith("Deutsch")


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
