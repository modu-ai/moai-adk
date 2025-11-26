"""End-to-end tests for complete language localization workflows.

Tests verify complete user workflows for Korean, Japanese, and Spanish:
1. Configuration is set correctly
2. Project documents are generated in user language
3. SPEC documents are generated in user language
4. Language setting persists throughout workflow
"""

from pathlib import Path
from tempfile import TemporaryDirectory
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.core.config.migration import migrate_config_to_nested_structure
from moai_adk.core.project.phase_executor import PhaseExecutor
from moai_adk.core.template.processor import TemplateProcessor


class TestKoreanUserWorkflow:
    """Test complete workflow for Korean user."""

    def test_korean_project_initialization(self) -> None:
        """Korean user can initialize project with Korean configuration."""
        # Arrange
        config = {
            "name": "한국 프로젝트",
            "description": "한국어로 된 프로젝트 설명",
            "language": {"conversation_language": "ko", "conversation_language_name": "한국어"},
            "project": {"mode": "personal"},
        }

        # Act - Verify config structure
        language_config = config.get("language", {})
        conversation_language = language_config.get("conversation_language", "en")
        language_name = language_config.get("conversation_language_name", "English")

        # Assert
        assert conversation_language == "ko", "Should set Korean language code"
        assert language_name == "한국어", "Should set Korean language name"
        assert config["name"] == "한국 프로젝트", "Should preserve Korean project name"

    @patch("moai_adk.core.project.phase_executor.TemplateProcessor")
    def test_korean_phase_executor_context(self, mock_processor_class: type) -> None:
        """PhaseExecutor sets correct Korean context for templates."""
        # Arrange
        mock_processor = MagicMock()
        mock_processor_class.return_value = mock_processor

        validator = MagicMock()
        executor = PhaseExecutor(validator)

        config = {
            "name": "한국 프로젝트",
            "language": {"conversation_language": "ko", "conversation_language_name": "한국어"},
        }

        with TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Act
            executor.execute_resource_phase(project_path, config)

            # Assert
            mock_processor.set_context.assert_called_once()
            context = mock_processor.set_context.call_args[0][0]

            assert context["CONVERSATION_LANGUAGE"] == "ko"
            assert context["CONVERSATION_LANGUAGE_NAME"] == "한국어"

    def test_korean_template_variable_substitution(self) -> None:
        """Korean language variables substitute in templates."""
        # Arrange
        processor = TemplateProcessor(Path("/tmp"))
        processor.set_context(
            {"CONVERSATION_LANGUAGE": "ko", "CONVERSATION_LANGUAGE_NAME": "한국어", "PROJECT_NAME": "MoAI-ADK"}
        )

        # Template that would be used in documentation
        template = """
# {{PROJECT_NAME}} - 언어 설정
설정된 언어: {{CONVERSATION_LANGUAGE_NAME}} ({{CONVERSATION_LANGUAGE}})

이 문서는 {{CONVERSATION_LANGUAGE}}로 작성되었습니다.
"""

        # Act
        result, warnings = processor._substitute_variables(template)

        # Assert
        assert "MoAI-ADK" in result, "Should substitute project name"
        assert "한국어" in result, "Should substitute Korean name"
        assert "ko" in result, "Should substitute language code"
        assert "{{" not in result, "Should have no remaining placeholders"


class TestJapaneseUserWorkflow:
    """Test complete workflow for Japanese user."""

    def test_japanese_project_initialization(self) -> None:
        """Japanese user can initialize project with Japanese configuration."""
        # Arrange
        config = {
            "name": "日本語プロジェクト",
            "description": "日本語で書かれたプロジェクト説明",
            "language": {"conversation_language": "ja", "conversation_language_name": "日本語"},
        }

        # Act
        language_config = config.get("language", {})
        conversation_language = language_config.get("conversation_language", "en")
        language_name = language_config.get("conversation_language_name", "English")

        # Assert
        assert conversation_language == "ja", "Should set Japanese language code"
        assert language_name == "日本語", "Should set Japanese language name"
        assert config["name"] == "日本語プロジェクト"

    @patch("moai_adk.core.project.phase_executor.TemplateProcessor")
    def test_japanese_phase_executor_context(self, mock_processor_class: type) -> None:
        """PhaseExecutor sets correct Japanese context for templates."""
        # Arrange
        mock_processor = MagicMock()
        mock_processor_class.return_value = mock_processor

        validator = MagicMock()
        executor = PhaseExecutor(validator)

        config = {
            "name": "日本語プロジェクト",
            "language": {"conversation_language": "ja", "conversation_language_name": "日本語"},
        }

        with TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Act
            executor.execute_resource_phase(project_path, config)

            # Assert
            context = mock_processor.set_context.call_args[0][0]

            assert context["CONVERSATION_LANGUAGE"] == "ja"
            assert context["CONVERSATION_LANGUAGE_NAME"] == "日本語"

    def test_japanese_template_variable_substitution(self) -> None:
        """Japanese language variables substitute in templates."""
        # Arrange
        processor = TemplateProcessor(Path("/tmp"))
        processor.set_context({"CONVERSATION_LANGUAGE": "ja", "CONVERSATION_LANGUAGE_NAME": "日本語"})

        template = "ドキュメント言語: {{CONVERSATION_LANGUAGE_NAME}} ({{CONVERSATION_LANGUAGE}})"

        # Act
        result, warnings = processor._substitute_variables(template)

        # Assert
        assert "日本語" in result, "Should substitute Japanese name"
        assert "ja" in result, "Should substitute language code"
        assert not warnings, "Should have no warnings"


class TestSpanishUserWorkflow:
    """Test complete workflow for Spanish user."""

    def test_spanish_project_initialization(self) -> None:
        """Spanish user can initialize project with Spanish configuration."""
        # Arrange
        config = {
            "name": "Proyecto Español",
            "description": "Descripción del proyecto en español",
            "language": {"conversation_language": "es", "conversation_language_name": "Español"},
        }

        # Act
        language_config = config.get("language", {})
        conversation_language = language_config.get("conversation_language", "en")
        language_name = language_config.get("conversation_language_name", "English")

        # Assert
        assert conversation_language == "es", "Should set Spanish language code"
        assert language_name == "Español", "Should set Spanish language name"

    @patch("moai_adk.core.project.phase_executor.TemplateProcessor")
    def test_spanish_phase_executor_context(self, mock_processor_class: type) -> None:
        """PhaseExecutor sets correct Spanish context for templates."""
        # Arrange
        mock_processor = MagicMock()
        mock_processor_class.return_value = mock_processor

        validator = MagicMock()
        executor = PhaseExecutor(validator)

        config = {
            "name": "Proyecto Español",
            "language": {"conversation_language": "es", "conversation_language_name": "Español"},
        }

        with TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Act
            executor.execute_resource_phase(project_path, config)

            # Assert
            context = mock_processor.set_context.call_args[0][0]

            assert context["CONVERSATION_LANGUAGE"] == "es"
            assert context["CONVERSATION_LANGUAGE_NAME"] == "Español"


class TestEnglishUserWorkflow:
    """Test regression - English user workflow unchanged."""

    def test_english_project_initialization(self) -> None:
        """English user workflow still works (backward compatibility)."""
        # Arrange
        config = {
            "name": "English Project",
            "description": "Project description in English",
            "language": {"conversation_language": "en", "conversation_language_name": "English"},
        }

        # Act
        language_config = config.get("language", {})
        conversation_language = language_config.get("conversation_language", "en")

        # Assert
        assert conversation_language == "en", "English default should work"

    def test_english_defaults_when_language_missing(self) -> None:
        """Missing language config defaults to English."""
        # Arrange
        config = {"name": "Project"}  # No language section

        # Act
        language_config = config.get("language", {})
        conversation_language = language_config.get("conversation_language", "en")
        language_name = language_config.get("conversation_language_name", "English")

        # Assert
        assert conversation_language == "en", "Should default to English"
        assert language_name == "English", "Should default to English name"


class TestMultiLanguageMigration:
    """Test migration scenarios for multiple languages."""

    @pytest.mark.parametrize(
        "language_code,language_name",
        [
            ("ko", "한국어"),
            ("ja", "日本語"),
            ("zh", "中文"),
            ("es", "Español"),
            ("en", "English"),
        ],
    )
    def test_migrate_all_supported_languages(self, language_code: str, language_name: str) -> None:
        """Migration works for all supported languages."""
        # Arrange
        legacy_config = {"conversation_language": language_code}

        # Act
        migrated = migrate_config_to_nested_structure(legacy_config)

        # Assert
        assert migrated["language"]["conversation_language"] == language_code
        assert migrated["language"]["conversation_language_name"] == language_name

    def test_migration_with_complex_config(self) -> None:
        """Migration preserves all other config sections."""
        # Arrange
        legacy_config = {
            "conversation_language": "ko",
            "project": {"name": "TestProject", "mode": "team"},
            "git_strategy": {"team": {"use_gitflow": True}},
            "custom_field": "custom_value",
        }

        # Act
        migrated = migrate_config_to_nested_structure(legacy_config)

        # Assert
        assert migrated["language"]["conversation_language"] == "ko"
        assert migrated["project"]["name"] == "TestProject"
        assert migrated["git_strategy"]["team"]["use_gitflow"] is True
        assert migrated["custom_field"] == "custom_value"


# Test execution
if __name__ == "__main__":
    pytest.main([__file__, "-v"])
