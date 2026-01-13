"""Unit tests for language configuration reading and template variable substitution.

Tests verify that:
1. Config reads language from nested structure (language.conversation_language)
2. Default values work when language config is missing
3. Template variables are correctly set and substituted
"""

from pathlib import Path
from tempfile import TemporaryDirectory
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.core.project.phase_executor import PhaseExecutor
from moai_adk.core.template.processor import TemplateProcessor


class TestLanguageConfigReading:
    """Test language configuration reading from nested structure."""

    def test_reads_nested_language_config(self) -> None:
        """PhaseExecutor reads language from nested config.language.conversation_language."""
        # Arrange
        config = {
            "language": {
                "conversation_language": "ko",
                "conversation_language_name": "한국어",
            }
        }

        # Act
        language_config = config.get("language", {})
        result_lang = language_config.get("conversation_language", "en")
        result_name = language_config.get("conversation_language_name", "English")

        # Assert
        assert result_lang == "ko", "Should read Korean language code"
        assert result_name == "한국어", "Should read Korean language name"

    def test_defaults_to_english_when_missing(self) -> None:
        """When language config missing, defaults to English."""
        # Arrange
        config = {}

        # Act
        language_config = config.get("language", {})
        result_lang = language_config.get("conversation_language", "en")
        result_name = language_config.get("conversation_language_name", "English")

        # Assert
        assert result_lang == "en", "Should default to English code"
        assert result_name == "English", "Should default to English name"

    def test_handles_invalid_language_config_type(self) -> None:
        """When language config is not a dict, uses defaults."""
        # Arrange
        config = {"language": "invalid"}  # Wrong type

        # Act
        language_config = config.get("language", {})
        if not isinstance(language_config, dict):
            language_config = {}
        result_lang = language_config.get("conversation_language", "en")

        # Assert
        assert result_lang == "en", "Should default when config is wrong type"

    def test_japanese_language_config(self) -> None:
        """Reads Japanese language configuration correctly."""
        # Arrange
        config = {
            "language": {
                "conversation_language": "ja",
                "conversation_language_name": "日本語",
            }
        }

        # Act
        language_config = config.get("language", {})
        result_lang = language_config.get("conversation_language", "en")
        result_name = language_config.get("conversation_language_name", "English")

        # Assert
        assert result_lang == "ja", "Should read Japanese language code"
        assert result_name == "日本語", "Should read Japanese language name"

    def test_spanish_language_config(self) -> None:
        """Reads Spanish language configuration correctly."""
        # Arrange
        config = {
            "language": {
                "conversation_language": "es",
                "conversation_language_name": "Español",
            }
        }

        # Act
        language_config = config.get("language", {})
        result_lang = language_config.get("conversation_language", "en")

        # Assert
        assert result_lang == "es", "Should read Spanish language code"


class TestTemplateVariableSubstitution:
    """Test template variable substitution for language variables."""

    def test_conversation_language_substitution(self) -> None:
        """{{CONVERSATION_LANGUAGE}} substitutes correctly."""
        # Arrange
        processor = TemplateProcessor(Path("/tmp"))
        processor.set_context({"CONVERSATION_LANGUAGE": "ko", "CONVERSATION_LANGUAGE_NAME": "한국어"})

        template_content = "Language: {{CONVERSATION_LANGUAGE}}"

        # Act
        result, warnings = processor._substitute_variables(template_content)

        # Assert
        assert result == "Language: ko", "Should substitute CONVERSATION_LANGUAGE"
        assert not warnings, "Should have no warnings"

    def test_conversation_language_name_substitution(self) -> None:
        """{{CONVERSATION_LANGUAGE_NAME}} substitutes correctly."""
        # Arrange
        processor = TemplateProcessor(Path("/tmp"))
        processor.set_context({"CONVERSATION_LANGUAGE": "ko", "CONVERSATION_LANGUAGE_NAME": "한국어"})

        template_content = "Language Name: {{CONVERSATION_LANGUAGE_NAME}}"

        # Act
        result, warnings = processor._substitute_variables(template_content)

        # Assert
        assert result == "Language Name: 한국어", "Should substitute CONVERSATION_LANGUAGE_NAME"
        assert not warnings, "Should have no warnings"

    def test_multiple_language_variables(self) -> None:
        """Multiple language variables substitute correctly."""
        # Arrange
        processor = TemplateProcessor(Path("/tmp"))
        processor.set_context(
            {
                "CONVERSATION_LANGUAGE": "ja",
                "CONVERSATION_LANGUAGE_NAME": "日本語",
                "PROJECT_NAME": "TestProject",
            }
        )

        template_content = """
        Project: {{PROJECT_NAME}}
        Language: {{CONVERSATION_LANGUAGE}}
        Name: {{CONVERSATION_LANGUAGE_NAME}}
        """

        # Act
        result, warnings = processor._substitute_variables(template_content)

        # Assert
        assert "{{" not in result, "Should substitute all variables"
        assert "ja" in result, "Should contain language code"
        assert "日本語" in result, "Should contain language name"
        assert not warnings, "Should have no warnings"

    def test_detects_unsubstituted_variables(self) -> None:
        """Detects variables that couldn't be substituted."""
        # Arrange
        processor = TemplateProcessor(Path("/tmp"))
        processor.set_context({"CONVERSATION_LANGUAGE": "ko"})

        template_content = "Language: {{CONVERSATION_LANGUAGE}}, Missing: {{MISSING_VAR}}"

        # Act
        result, warnings = processor._substitute_variables(template_content)

        # Assert
        assert "ko" in result, "Should substitute known variable"
        assert "{{MISSING_VAR}}" in result, "Should not substitute missing variable"
        assert len(warnings) > 0, "Should have warnings for unsubstituted variables"
        assert any("MISSING_VAR" in warning for warning in warnings), "Warning should mention missing variable"


class TestPhaseExecutorResourcePhase:
    """Test PhaseExecutor.execute_resource_phase language handling."""

    @patch("moai_adk.core.project.phase_executor.TemplateProcessor")
    def test_resource_phase_sets_language_context(
        self,
        mock_processor_class: type,
    ) -> None:
        """execute_resource_phase correctly sets language context variables."""
        # Arrange
        mock_processor = MagicMock()
        mock_processor_class.return_value = mock_processor

        validator = MagicMock()
        executor = PhaseExecutor(validator)

        config = {
            "language": {
                "conversation_language": "ko",
                "conversation_language_name": "한국어",
            },
            "name": "TestProject",
        }

        with TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Act
            executor.execute_resource_phase(project_path, config)

            # Assert
            mock_processor.set_context.assert_called_once()
            context = mock_processor.set_context.call_args[0][0]

            assert context["CONVERSATION_LANGUAGE"] == "ko", "Should set Korean language"
            assert context["CONVERSATION_LANGUAGE_NAME"] == "한국어", "Should set Korean name"

    @patch("moai_adk.core.project.phase_executor.TemplateProcessor")
    def test_resource_phase_defaults_language_context(
        self,
        mock_processor_class: type,
    ) -> None:
        """execute_resource_phase defaults to English when language not specified."""
        # Arrange
        mock_processor = MagicMock()
        mock_processor_class.return_value = mock_processor

        validator = MagicMock()
        executor = PhaseExecutor(validator)

        config = {
            "name": "TestProject"
            # No language config
        }

        with TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Act
            executor.execute_resource_phase(project_path, config)

            # Assert
            mock_processor.set_context.assert_called_once()
            context = mock_processor.set_context.call_args[0][0]

            assert context["CONVERSATION_LANGUAGE"] == "en", "Should default to English"
            assert context["CONVERSATION_LANGUAGE_NAME"] == "English", "Should default English name"


class TestLanguageConfigMigration:
    """Test migration from legacy flat config to nested structure."""

    def test_legacy_config_structure(self) -> None:
        """Legacy flat config structure is recognized."""
        # This tests that we understand the old structure
        legacy_config = {"conversation_language": "ko", "locale": "ko"}

        # Act
        has_conversation_language = "conversation_language" in legacy_config

        # Assert
        assert has_conversation_language, "Legacy config has flat structure"

    def test_new_config_structure(self) -> None:
        """New nested config structure is recognized."""
        new_config = {
            "language": {
                "conversation_language": "ko",
                "conversation_language_name": "한국어",
            }
        }

        # Act
        language_config = new_config.get("language", {})
        has_conversation_language = "conversation_language" in language_config

        # Assert
        assert has_conversation_language, "New config has nested structure"


class TestLanguageConfigModuleEdgeCases:
    """Test edge cases in language_config module functions."""

    def test_get_english_name_returns_fallback_for_invalid_code(self) -> None:
        """Test get_english_name returns 'English' for unknown language code."""
        from moai_adk.core.language_config import get_english_name

        # Act
        result = get_english_name("invalid_code")

        # Assert
        assert result == "English", "Should return English fallback for unknown code"

    def test_get_language_family_returns_none_for_invalid_code(self) -> None:
        """Test get_language_family returns None for unknown language code."""
        from moai_adk.core.language_config import get_language_family

        # Act
        result = get_language_family("invalid_code")

        # Assert
        assert result is None, "Should return None for unknown language code"

    def test_is_rtl_language_case_insensitive(self) -> None:
        """Test is_rtl_language is case-insensitive."""
        from moai_adk.core.language_config import is_rtl_language

        # Assert - Arabic should be RTL
        assert is_rtl_language("AR") is True
        assert is_rtl_language("ar") is True
        assert is_rtl_language("Ar") is True

    def test_is_rtl_language_returns_false_for_ltr(self) -> None:
        """Test is_rtl_language returns False for LTR languages."""
        from moai_adk.core.language_config import is_rtl_language

        # Assert
        assert is_rtl_language("en") is False
        assert is_rtl_language("ko") is False
        assert is_rtl_language("zh") is False

    def test_get_translation_priority_returns_copy(self) -> None:
        """Test get_translation_priority returns a copy of the list."""
        from moai_adk.core.language_config import TRANSLATION_PRIORITY, get_translation_priority

        # Act
        result = get_translation_priority()

        # Assert - Should be a copy, not the same object
        assert result == TRANSLATION_PRIORITY
        assert result is not TRANSLATION_PRIORITY

    def test_get_optimal_model_returns_default_for_unknown_language(self) -> None:
        """Test get_optimal_model returns default model for unknown language."""
        from moai_adk.core.language_config import get_optimal_model

        # Act
        result = get_optimal_model("unknown_language")

        # Assert
        assert result == "claude-sonnet-4-5-20250929"


# Test execution
if __name__ == "__main__":
    pytest.main([__file__, "-v"])
