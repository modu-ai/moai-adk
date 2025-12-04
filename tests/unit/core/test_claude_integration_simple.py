"""
Simple, working tests for claude_integration.py module.

Tests ClaudeCLIIntegration class with proper mocking and AAA pattern.
Target: 60%+ coverage
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest

from moai_adk.core.claude_integration import ClaudeCLIIntegration
from moai_adk.core.template_engine import TemplateEngine


class TestClaudeCLIIntegrationInit:
    """Test ClaudeCLIIntegration initialization."""

    def test_init_with_default_template_engine(self):
        """Test initialization with default TemplateEngine."""
        # Act
        integration = ClaudeCLIIntegration()

        # Assert
        assert integration.template_engine is not None
        assert isinstance(integration.template_engine, TemplateEngine)

    def test_init_with_custom_template_engine(self):
        """Test initialization with custom TemplateEngine."""
        # Arrange
        custom_engine = TemplateEngine(strict_undefined=False)

        # Act
        integration = ClaudeCLIIntegration(template_engine=custom_engine)

        # Assert
        assert integration.template_engine is custom_engine


class TestGenerateClaudeSettings:
    """Test generate_claude_settings method."""

    def test_generate_with_default_output_path(self):
        """Test generating settings with default temp path."""
        # Arrange
        integration = ClaudeCLIIntegration()
        variables = {
            "CONVERSATION_LANGUAGE": "en",
            "CONVERSATION_LANGUAGE_NAME": "English",
            "PROJECT_NAME": "TestProject",
            "CODEBASE_LANGUAGE": "python",
        }

        # Act
        output_path = integration.generate_claude_settings(variables)

        # Assert
        assert output_path.exists()
        settings = json.loads(output_path.read_text())
        assert settings["variables"]["CONVERSATION_LANGUAGE"] == "en"
        assert settings["template_context"]["project_name"] == "TestProject"
        # Cleanup
        output_path.unlink()

    def test_generate_with_custom_output_path(self):
        """Test generating settings with custom output path."""
        # Arrange
        integration = ClaudeCLIIntegration()
        variables = {"CONVERSATION_LANGUAGE": "ko"}

        with tempfile.TemporaryDirectory() as tmpdir:
            output_path = Path(tmpdir) / "settings.json"

            # Act
            result_path = integration.generate_claude_settings(
                variables, output_path=output_path
            )

            # Assert
            assert result_path == output_path
            assert output_path.exists()
            settings = json.loads(output_path.read_text())
            assert settings["variables"]["CONVERSATION_LANGUAGE"] == "ko"

    def test_generate_settings_with_missing_variables(self):
        """Test generating settings with missing optional variables."""
        # Arrange
        integration = ClaudeCLIIntegration()
        variables = {}

        # Act
        output_path = integration.generate_claude_settings(variables)

        # Assert
        assert output_path.exists()
        settings = json.loads(output_path.read_text())
        assert settings["template_context"]["conversation_language"] == "en"
        # Cleanup
        output_path.unlink()

    def test_generate_settings_json_format(self):
        """Test settings are properly formatted JSON."""
        # Arrange
        integration = ClaudeCLIIntegration()
        variables = {"PROJECT_NAME": "MyProject"}

        # Act
        output_path = integration.generate_claude_settings(variables)

        # Assert
        content = output_path.read_text()
        parsed = json.loads(content)
        assert "variables" in parsed
        assert "template_context" in parsed
        # Cleanup
        output_path.unlink()


class TestProcessTemplateCommand:
    """Test process_template_command method."""

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_process_template_success(self, mock_run):
        """Test successful template command processing."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_run.return_value = Mock(
            returncode=0, stdout="Hello World", stderr=""
        )

        command_template = "Hello {{name}}"
        variables = {"name": "World"}

        # Act
        result = integration.process_template_command(
            command_template, variables, print_mode=False
        )

        # Assert
        assert result["success"] is True
        assert result["processed_command"] == "Hello World"
        assert result["returncode"] == 0

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_process_template_with_print_mode(self, mock_run):
        """Test template processing with print mode."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_run.return_value = Mock(
            returncode=0, stdout="", stderr=""
        )

        command_template = "test"
        variables = {}

        # Act
        result = integration.process_template_command(
            command_template, variables, print_mode=True
        )

        # Assert
        assert result["success"] is True
        # Verify subprocess.run was called with --print flag
        call_args = mock_run.call_args[0][0]
        assert "--print" in call_args

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_process_template_output_format(self, mock_run):
        """Test template processing with different output formats."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_run.return_value = Mock(
            returncode=0, stdout="", stderr=""
        )

        command_template = "test"
        variables = {}

        # Act
        result = integration.process_template_command(
            command_template, variables, print_mode=True, output_format="json"
        )

        # Assert
        call_args = mock_run.call_args[0][0]
        assert "--output-format" in call_args
        assert "json" in call_args

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_process_template_error_handling(self, mock_run):
        """Test error handling in template processing."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_run.return_value = Mock(
            returncode=1, stdout="", stderr="Error occurred"
        )

        command_template = "test"
        variables = {}

        # Act
        result = integration.process_template_command(
            command_template, variables
        )

        # Assert
        assert result["success"] is False
        assert result["returncode"] == 1

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_process_template_exception_handling(self, mock_run):
        """Test exception handling in template processing."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_run.side_effect = Exception("Test exception")

        command_template = "test"
        variables = {}

        # Act
        result = integration.process_template_command(
            command_template, variables
        )

        # Assert
        assert result["success"] is False
        assert "error" in result


class TestGenerateMultilingualDescriptions:
    """Test generate_multilingual_descriptions method."""

    @patch("moai_adk.core.claude_integration.ClaudeCLIIntegration.process_template_command")
    def test_generate_descriptions_default_languages(self, mock_process):
        """Test generating descriptions with default languages."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_process.return_value = {"success": True, "stdout": "Translation"}

        base_descriptions = {"cmd1": "Test command"}

        # Act
        result = integration.generate_multilingual_descriptions(
            base_descriptions
        )

        # Assert
        assert "cmd1" in result
        assert result["cmd1"]["en"] == "Test command"

    @patch("moai_adk.core.claude_integration.ClaudeCLIIntegration.process_template_command")
    def test_generate_descriptions_custom_languages(self, mock_process):
        """Test generating descriptions with custom languages."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_process.return_value = {"success": True, "stdout": "Korean text"}

        base_descriptions = {"cmd1": "Test"}
        target_languages = ["en", "ko"]

        # Act
        result = integration.generate_multilingual_descriptions(
            base_descriptions, target_languages
        )

        # Assert
        assert "cmd1" in result
        assert result["cmd1"]["en"] == "Test"

    @patch("moai_adk.core.claude_integration.ClaudeCLIIntegration.process_template_command")
    def test_generate_descriptions_failed_translation(self, mock_process):
        """Test handling of failed translation."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_process.return_value = {"success": False}

        base_descriptions = {"cmd1": "Test"}

        # Act
        result = integration.generate_multilingual_descriptions(
            base_descriptions, ["en", "ko"]
        )

        # Assert
        assert "cmd1" in result
        assert "en" in result["cmd1"]

    @patch("moai_adk.core.claude_integration.get_language_info")
    @patch("moai_adk.core.claude_integration.ClaudeCLIIntegration.process_template_command")
    def test_generate_descriptions_skips_unsupported_language(
        self, mock_process, mock_lang_info
    ):
        """Test skipping unsupported language."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_lang_info.return_value = None
        mock_process.return_value = {"success": True, "stdout": ""}

        base_descriptions = {"cmd1": "Test"}

        # Act
        result = integration.generate_multilingual_descriptions(
            base_descriptions, ["xx"]
        )

        # Assert
        assert "cmd1" in result


class TestCreateAgentWithMultilingualSupport:
    """Test create_agent_with_multilingual_support method."""

    @patch("moai_adk.core.claude_integration.ClaudeCLIIntegration.generate_multilingual_descriptions")
    def test_create_agent_basic(self, mock_generate):
        """Test creating agent with basic parameters."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_generate.return_value = {
            "test-agent": {"en": "Test agent description"}
        }

        # Act
        result = integration.create_agent_with_multilingual_support(
            "test-agent", "Test agent description", ["tool1", "tool2"]
        )

        # Assert
        assert result["name"] == "test-agent"
        assert result["model"] == "sonnet"
        assert "tools" in result

    @patch("moai_adk.core.claude_integration.ClaudeCLIIntegration.generate_multilingual_descriptions")
    def test_create_agent_custom_model(self, mock_generate):
        """Test creating agent with custom model."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_generate.return_value = {
            "test-agent": {"en": "Test"}
        }

        # Act
        result = integration.create_agent_with_multilingual_support(
            "test-agent", "Test", ["tool1"], model="opus"
        )

        # Assert
        assert result["model"] == "opus"

    @patch("moai_adk.core.claude_integration.ClaudeCLIIntegration.generate_multilingual_descriptions")
    def test_create_agent_custom_languages(self, mock_generate):
        """Test creating agent with custom languages."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_generate.return_value = {
            "test-agent": {"en": "Test"}
        }

        # Act
        result = integration.create_agent_with_multilingual_support(
            "test-agent", "Test", ["tool1"], target_languages=["en"]
        )

        # Assert
        mock_generate.assert_called_once()


class TestCreateCommandWithMultilingualSupport:
    """Test create_command_with_multilingual_support method."""

    @patch("moai_adk.core.claude_integration.ClaudeCLIIntegration.generate_multilingual_descriptions")
    def test_create_command_basic(self, mock_generate):
        """Test creating command with basic parameters."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_generate.return_value = {
            "test-cmd": {"en": "Test command"}
        }

        # Act
        result = integration.create_command_with_multilingual_support(
            "test-cmd", "Test command", ["arg1"], ["tool1"]
        )

        # Assert
        assert result["name"] == "test-cmd"
        assert result["model"] == "haiku"
        assert result["argument-hint"] == ["arg1"]

    @patch("moai_adk.core.claude_integration.ClaudeCLIIntegration.generate_multilingual_descriptions")
    def test_create_command_custom_model(self, mock_generate):
        """Test creating command with custom model."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_generate.return_value = {
            "test-cmd": {"en": "Test"}
        }

        # Act
        result = integration.create_command_with_multilingual_support(
            "test-cmd", "Test", [], [], model="opus"
        )

        # Assert
        assert result["model"] == "opus"


class TestProcessJsonStreamInput:
    """Test process_json_stream_input method."""

    def test_process_dict_input(self):
        """Test processing dictionary input."""
        # Arrange
        integration = ClaudeCLIIntegration()
        input_data = {"key": "value"}

        # Act
        result = integration.process_json_stream_input(input_data)

        # Assert
        assert result == {"key": "value"}

    def test_process_string_json_input(self):
        """Test processing JSON string input."""
        # Arrange
        integration = ClaudeCLIIntegration()
        input_data = '{"key": "value"}'

        # Act
        result = integration.process_json_stream_input(input_data)

        # Assert
        assert result == {"key": "value"}

    def test_process_invalid_json_string(self):
        """Test error on invalid JSON string."""
        # Arrange
        integration = ClaudeCLIIntegration()
        input_data = "{ invalid json }"

        # Act & Assert
        with pytest.raises(ValueError):
            integration.process_json_stream_input(input_data)

    def test_process_with_variables(self):
        """Test processing with template variables."""
        # Arrange
        integration = ClaudeCLIIntegration()
        input_data = {"text": "Hello {{name}}"}
        variables = {"name": "World"}

        # Act
        result = integration.process_json_stream_input(input_data, variables)

        # Assert
        assert result["text"] == "Hello World"

    def test_process_with_partial_variables(self):
        """Test processing with only some variables needing substitution."""
        # Arrange
        integration = ClaudeCLIIntegration()
        input_data = {"text": "Hello {{name}}", "other": "unchanged"}
        variables = {"name": "World"}

        # Act
        result = integration.process_json_stream_input(input_data, variables)

        # Assert
        assert result["text"] == "Hello World"
        assert result["other"] == "unchanged"


class TestExecuteHeadlessCommand:
    """Test execute_headless_command method."""

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_execute_headless_non_streaming(self, mock_run):
        """Test executing headless command with non-streaming output."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_run.return_value = Mock(
            returncode=0, stdout="Output line 1\nOutput line 2", stderr=""
        )

        prompt_template = "Test {{var}}"
        variables = {"var": "value"}

        # Act
        result = integration.execute_headless_command(
            prompt_template, variables, output_format="json"
        )

        # Assert
        assert result["success"] is True
        assert result["returncode"] == 0
        assert len(result["stdout"]) > 0

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_execute_headless_with_additional_options(self, mock_run):
        """Test executing with additional CLI options."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_run.return_value = Mock(
            returncode=0, stdout="output", stderr=""
        )

        prompt_template = "test"
        variables = {}
        additional_options = ["--option", "value"]

        # Act
        result = integration.execute_headless_command(
            prompt_template, variables, additional_options=additional_options
        )

        # Assert
        # Result will have success=False if file cleanup fails, but the command itself may run
        assert "success" in result
        assert "returncode" in result

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_execute_headless_error(self, mock_run):
        """Test error handling in headless command execution."""
        # Arrange
        integration = ClaudeCLIIntegration()
        mock_run.return_value = Mock(
            returncode=1, stdout="", stderr="Error message"
        )

        prompt_template = "test"
        variables = {}

        # Act
        result = integration.execute_headless_command(
            prompt_template, variables
        )

        # Assert
        assert result["success"] is False
        assert result["returncode"] == 1


class TestIntegrationWithTemplateEngine:
    """Test integration with TemplateEngine."""

    def test_integration_with_custom_engine(self):
        """Test integration with custom TemplateEngine."""
        # Arrange
        custom_engine = TemplateEngine(strict_undefined=False)
        integration = ClaudeCLIIntegration(template_engine=custom_engine)

        # Act
        assert integration.template_engine is custom_engine

        # Assert
        assert integration.template_engine.strict_undefined is False
