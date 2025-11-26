"""Integration tests for language configuration flow across commands and agents.

Tests verify that:
1. Language variables are properly substituted in command templates
2. Commands pass language parameters to sub-agents
3. Config migration works correctly
4. All 4 Alfred commands support language parameters
"""

from pathlib import Path
from tempfile import TemporaryDirectory
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.core.config.migration import (
    get_conversation_language,
    get_conversation_language_name,
    migrate_config_to_nested_structure,
)
from moai_adk.core.project.phase_executor import PhaseExecutor
from moai_adk.core.template.processor import TemplateProcessor


class TestConfigMigration:
    """Test configuration migration from flat to nested structure."""

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_migrate_legacy_flat_config_to_nested(self) -> None:
        """Legacy flat config converts to nested structure."""
        # Arrange
        legacy_config = {"conversation_language": "ko", "locale": "ko", "project": {"name": "TestProject"}}

        # Act
        migrated = migrate_config_to_nested_structure(legacy_config)

        # Assert
        assert "language" in migrated, "Should have nested language section"
        assert migrated["language"]["conversation_language"] == "ko"
        assert migrated["language"]["conversation_language_name"] == "한국어"
        assert "conversation_language" not in migrated, "Should remove flat key"
        assert migrated["project"]["name"] == "TestProject", "Should preserve other keys"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_migrate_preserves_language_setting(self) -> None:
        """Migration preserves original language value exactly."""
        languages = [
            ("ko", "한국어"),
            ("ja", "日本語"),
            ("zh", "中文"),
            ("es", "Español"),
            ("en", "English"),
        ]

        for lang_code, lang_name in languages:
            legacy = {"conversation_language": lang_code}
            migrated = migrate_config_to_nested_structure(legacy)
            assert migrated["language"]["conversation_language"] == lang_code
            assert migrated["language"]["conversation_language_name"] == lang_name

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_migrate_already_nested_config_unchanged(self) -> None:
        """Config already in nested structure should not be modified."""
        # Arrange
        nested_config = {"language": {"conversation_language": "ko", "conversation_language_name": "한국어"}}

        # Act
        result = migrate_config_to_nested_structure(nested_config)

        # Assert
        assert result == nested_config, "Should not modify already nested config"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_get_conversation_language_from_nested(self) -> None:
        """Reading language from nested structure works."""
        config = {"language": {"conversation_language": "ja", "conversation_language_name": "日本語"}}

        result = get_conversation_language(config)

        assert result == "ja", "Should read from nested structure"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_get_conversation_language_default_to_english(self) -> None:
        """Missing language config defaults to English."""
        result = get_conversation_language({})
        assert result == "en", "Should default to English"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_get_conversation_language_name(self) -> None:
        """Reading language name works correctly."""
        config = {"language": {"conversation_language": "ko", "conversation_language_name": "한국어"}}

        result = get_conversation_language_name(config)

        assert result == "한국어", "Should read language name"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_get_conversation_language_name_default(self) -> None:
        """Missing language name maps from language code."""
        config = {
            "language": {
                "conversation_language": "ja"
                # conversation_language_name missing
            }
        }

        result = get_conversation_language_name(config)

        assert result == "日本語", "Should map Japanese code to name"


class TestPhaseExecutorLanguageContext:
    """Test that PhaseExecutor sets correct language context."""

    @patch("moai_adk.core.project.phase_executor.TemplateProcessor")
    def test_executor_sets_korean_language_context(self, mock_processor_class: type) -> None:
        """PhaseExecutor sets Korean language in context."""
        # Arrange
        mock_processor = MagicMock()
        mock_processor_class.return_value = mock_processor

        validator = MagicMock()
        executor = PhaseExecutor(validator)

        config = {
            "language": {"conversation_language": "ko", "conversation_language_name": "한국어"},
            "name": "TestProject",
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

    @patch("moai_adk.core.project.phase_executor.TemplateProcessor")
    def test_executor_sets_japanese_language_context(self, mock_processor_class: type) -> None:
        """PhaseExecutor sets Japanese language in context."""
        # Arrange
        mock_processor = MagicMock()
        mock_processor_class.return_value = mock_processor

        validator = MagicMock()
        executor = PhaseExecutor(validator)

        config = {
            "language": {"conversation_language": "ja", "conversation_language_name": "日本語"},
            "name": "TestProject",
        }

        with TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Act
            executor.execute_resource_phase(project_path, config)

            # Assert
            context = mock_processor.set_context.call_args[0][0]
            assert context["CONVERSATION_LANGUAGE"] == "ja"
            assert context["CONVERSATION_LANGUAGE_NAME"] == "日本語"

    @patch("moai_adk.core.project.phase_executor.TemplateProcessor")
    def test_executor_defaults_to_english_when_missing(self, mock_processor_class: type) -> None:
        """PhaseExecutor defaults to English when language not specified."""
        # Arrange
        mock_processor = MagicMock()
        mock_processor_class.return_value = mock_processor

        validator = MagicMock()
        executor = PhaseExecutor(validator)

        config = {"name": "TestProject"}  # No language config

        with TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Act
            executor.execute_resource_phase(project_path, config)

            # Assert
            context = mock_processor.set_context.call_args[0][0]
            assert context["CONVERSATION_LANGUAGE"] == "en"
            assert context["CONVERSATION_LANGUAGE_NAME"] == "English"


class TestTemplateVariableSubstitution:
    """Test that template variables substitute correctly in commands."""

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_conversation_language_variable_substitution(self) -> None:
        """{{CONVERSATION_LANGUAGE}} variable substitutes correctly."""
        # Arrange
        processor = TemplateProcessor(Path("/tmp"))
        processor.set_context(
            {"CONVERSATION_LANGUAGE": "ko", "CONVERSATION_LANGUAGE_NAME": "한국어", "PROJECT_NAME": "TestProject"}
        )

        template = "Language: {{CONVERSATION_LANGUAGE}}, Project: {{PROJECT_NAME}}"

        # Act
        result, warnings = processor._substitute_variables(template)

        # Assert
        assert "ko" in result, "Should substitute language code"
        assert "TestProject" in result, "Should substitute project name"
        assert "{{" not in result, "Should have no remaining placeholders"
        assert not warnings, "Should have no warnings for complete substitution"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_conversation_language_name_substitution(self) -> None:
        """{{CONVERSATION_LANGUAGE_NAME}} variable substitutes correctly."""
        # Arrange
        processor = TemplateProcessor(Path("/tmp"))
        processor.set_context({"CONVERSATION_LANGUAGE": "ja", "CONVERSATION_LANGUAGE_NAME": "日本語"})

        template = "Language: {{CONVERSATION_LANGUAGE_NAME}}"

        # Act
        result, warnings = processor._substitute_variables(template)

        # Assert
        assert "日本語" in result, "Should substitute Japanese name"
        assert not warnings, "Should have no warnings"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_detects_missing_variables(self) -> None:
        """Detects variables that are not substituted."""
        # Arrange
        processor = TemplateProcessor(Path("/tmp"))
        processor.set_context({"CONVERSATION_LANGUAGE": "ko"})

        template = "Language: {{CONVERSATION_LANGUAGE}}, Missing: {{MISSING_VAR}}"

        # Act
        result, warnings = processor._substitute_variables(template)

        # Assert
        assert "ko" in result, "Should substitute known variable"
        assert "{{MISSING_VAR}}" in result, "Should not substitute unknown variable"
        assert len(warnings) > 0, "Should have warnings for missing variables"
        # Check all warnings for the missing variable (not just the first one)
        warning_text = " ".join(warnings)
        assert "MISSING_VAR" in warning_text, "Warning should mention missing variable"


class TestCommandLanguageParameters:
    """Test that commands properly document language parameters."""

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_0_project_command_has_language_config(self) -> None:
        """0-project command documents language configuration."""
        with open(Path(__file__).parent.parent.parent / ".claude" / "commands" / "alfred" / "0-project.md") as f:
            content = f.read()

        assert "CONVERSATION_LANGUAGE" in content, "Should reference language variable"
        assert "CONVERSATION_LANGUAGE_NAME" in content, "Should reference language name"
        assert "conversation_language" in content, "Should explain language parameter"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_1_plan_command_has_language_config(self) -> None:
        """1-plan command documents language configuration for both steps."""
        with open(Path(__file__).parent.parent.parent / ".claude" / "commands" / "alfred" / "1-plan.md") as f:
            content = f.read()

        # Count occurrences - should be in both STEP 1 and STEP 2
        count = content.count("{{CONVERSATION_LANGUAGE}}")
        assert count >= 2, "Should reference language in both STEP 1 and STEP 2"
        assert "CONVERSATION_LANGUAGE_NAME" in content

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_2_run_command_has_language_config(self) -> None:
        """2-run command documents language configuration."""
        with open(Path(__file__).parent.parent.parent / ".claude" / "commands" / "alfred" / "2-run.md") as f:
            content = f.read()

        assert "CONVERSATION_LANGUAGE" in content, "Should reference language variable"
        assert "Code and technical output MUST be in English" in content

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_3_sync_command_has_language_config(self) -> None:
        """3-sync command documents language configuration."""
        with open(Path(__file__).parent.parent.parent / ".claude" / "commands" / "alfred" / "3-sync.md") as f:
            content = f.read()

        assert "CONVERSATION_LANGUAGE" in content, "Should reference language variable"


# Test execution
if __name__ == "__main__":
    pytest.main([__file__, "-v"])
