"""
Comprehensive tests for ClaudeCLIIntegration module.

Tests the Claude CLI integration with template variable processing,
JSON streaming, and multilingual support.
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, Mock, call, patch

import pytest

from moai_adk.core.claude_integration import ClaudeCLIIntegration
from moai_adk.core.template_engine import TemplateEngine


class TestClaudeCLIIntegrationInit:
    """Test ClaudeCLIIntegration initialization."""

    def test_init_with_default_template_engine(self):
        """Test initialization creates default TemplateEngine."""
        integration = ClaudeCLIIntegration()
        assert integration.template_engine is not None
        assert isinstance(integration.template_engine, TemplateEngine)

    def test_init_with_custom_template_engine(self):
        """Test initialization accepts custom TemplateEngine."""
        custom_engine = TemplateEngine()
        integration = ClaudeCLIIntegration(template_engine=custom_engine)
        assert integration.template_engine is custom_engine


class TestGenerateClaudeSettings:
    """Test generate_claude_settings method."""

    def test_generate_with_default_output_path(self):
        """Test settings generation with auto-generated path."""
        integration = ClaudeCLIIntegration()
        variables = {"PROJECT_NAME": "test-project", "CONVERSATION_LANGUAGE": "en"}

        result_path = integration.generate_claude_settings(variables)

        assert result_path.exists()
        assert result_path.suffix == ".json"

        # Verify content
        with open(result_path, "r") as f:
            settings = json.load(f)

        assert "variables" in settings
        assert settings["variables"]["PROJECT_NAME"] == "test-project"
        assert "template_context" in settings
        assert settings["template_context"]["conversation_language"] == "en"

        # Cleanup
        result_path.unlink()

    def test_generate_with_custom_output_path(self):
        """Test settings generation with specified path."""
        integration = ClaudeCLIIntegration()
        variables = {"PROJECT_NAME": "custom-project", "CODEBASE_LANGUAGE": "python"}

        with tempfile.TemporaryDirectory() as tmpdir:
            output_path = Path(tmpdir) / "custom_settings.json"
            result_path = integration.generate_claude_settings(variables, output_path)

            assert result_path == output_path
            assert result_path.exists()

            with open(result_path, "r") as f:
                settings = json.load(f)

            assert settings["variables"]["PROJECT_NAME"] == "custom-project"
            assert settings["template_context"]["codebase_language"] == "python"

    def test_generate_settings_with_missing_variables(self):
        """Test settings generation with missing optional variables."""
        integration = ClaudeCLIIntegration()
        variables = {"PROJECT_NAME": "minimal"}

        with tempfile.TemporaryDirectory() as tmpdir:
            output_path = Path(tmpdir) / "settings.json"
            result_path = integration.generate_claude_settings(variables, output_path)

            with open(result_path, "r") as f:
                settings = json.load(f)

            # Should use defaults for missing values
            assert settings["template_context"]["conversation_language"] == "en"
            assert settings["template_context"]["codebase_language"] == "python"

    def test_generate_settings_json_format(self):
        """Test that generated settings are valid JSON."""
        integration = ClaudeCLIIntegration()
        variables = {"PROJECT_NAME": "test", "CONVERSATION_LANGUAGE": "ko"}

        with tempfile.TemporaryDirectory() as tmpdir:
            output_path = Path(tmpdir) / "settings.json"
            result_path = integration.generate_claude_settings(variables, output_path)

            # Should be readable as JSON
            with open(result_path, "r", encoding="utf-8") as f:
                settings = json.load(f)

            # Verify structure
            assert isinstance(settings, dict)
            assert isinstance(settings["variables"], dict)
            assert isinstance(settings["template_context"], dict)


class TestProcessTemplateCommand:
    """Test process_template_command method."""

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_process_template_success(self, mock_subprocess_run):
        """Test successful template command processing."""
        # Setup
        mock_subprocess_run.return_value = Mock(
            returncode=0, stdout="output", stderr="", text=True
        )

        integration = ClaudeCLIIntegration()
        command_template = "Hello {{NAME}}"
        variables = {"NAME": "World"}

        # Execute
        with patch.object(
            integration.template_engine, "render_string", return_value="Hello World"
        ):
            result = integration.process_template_command(command_template, variables)

        # Assert
        assert result["success"] is True
        assert result["stdout"] == "output"
        assert result["returncode"] == 0
        assert "processed_command" in result
        assert result["variables_used"] == variables

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_process_template_with_print_mode(self, mock_subprocess_run):
        """Test template processing with print mode enabled."""
        mock_subprocess_run.return_value = Mock(returncode=0, stdout="", stderr="")

        integration = ClaudeCLIIntegration()
        with patch.object(
            integration.template_engine, "render_string", return_value="test"
        ):
            integration.process_template_command("test", {}, print_mode=True)

        # Verify --print flag was added
        call_args = mock_subprocess_run.call_args[0][0]
        assert "--print" in call_args

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_process_template_output_format(self, mock_subprocess_run):
        """Test different output formats."""
        mock_subprocess_run.return_value = Mock(returncode=0, stdout="", stderr="")

        integration = ClaudeCLIIntegration()
        with patch.object(
            integration.template_engine, "render_string", return_value="test"
        ):
            integration.process_template_command(
                "test", {}, output_format="stream-json"
            )

        call_args = mock_subprocess_run.call_args[0][0]
        assert "--output-format" in call_args
        assert "stream-json" in call_args

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_process_template_error_handling(self, mock_subprocess_run):
        """Test error handling in template processing."""
        mock_subprocess_run.return_value = Mock(
            returncode=1, stdout="", stderr="Error occurred"
        )

        integration = ClaudeCLIIntegration()
        with patch.object(
            integration.template_engine, "render_string", return_value="test"
        ):
            result = integration.process_template_command("test", {})

        assert result["success"] is False
        assert result["returncode"] == 1
        assert result["stderr"] == "Error occurred"

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_process_template_exception_handling(self, mock_subprocess_run):
        """Test exception handling during processing."""
        mock_subprocess_run.side_effect = Exception("Command failed")

        integration = ClaudeCLIIntegration()
        with patch.object(
            integration.template_engine, "render_string", return_value="test"
        ):
            result = integration.process_template_command("test", {})

        assert result["success"] is False
        assert "error" in result
        assert "Command failed" in result["error"]


class TestGenerateMultilingualDescriptions:
    """Test generate_multilingual_descriptions method."""

    @patch("moai_adk.core.claude_integration.get_language_info")
    def test_generate_descriptions_default_languages(self, mock_get_lang_info):
        """Test description generation with default languages."""
        mock_get_lang_info.return_value = {"native_name": "Korean", "name": "ko"}

        integration = ClaudeCLIIntegration()
        base_descriptions = {"command1": "A test command"}

        with patch.object(
            integration, "process_template_command"
        ) as mock_process:
            mock_process.return_value = {
                "success": True,
                "stdout": "테스트 명령",
            }

            result = integration.generate_multilingual_descriptions(base_descriptions)

        assert "command1" in result
        assert "en" in result["command1"]
        assert result["command1"]["en"] == "A test command"

    @patch("moai_adk.core.claude_integration.get_language_info")
    def test_generate_descriptions_custom_languages(self, mock_get_lang_info):
        """Test description generation with custom language list."""
        mock_get_lang_info.return_value = {"native_name": "Spanish", "name": "es"}

        integration = ClaudeCLIIntegration()
        base_descriptions = {"test": "Test description"}
        target_languages = ["en", "es"]

        with patch.object(
            integration, "process_template_command"
        ) as mock_process:
            mock_process.return_value = {
                "success": True,
                "stdout": "Descripción de prueba",
            }

            result = integration.generate_multilingual_descriptions(
                base_descriptions, target_languages
            )

        assert "test" in result
        assert "en" in result["test"]

    @patch("moai_adk.core.claude_integration.get_language_info")
    def test_generate_descriptions_failed_translation(self, mock_get_lang_info):
        """Test description generation when translation fails."""
        mock_get_lang_info.return_value = {"native_name": "French", "name": "fr"}

        integration = ClaudeCLIIntegration()
        base_descriptions = {"test": "Original"}

        with patch.object(
            integration, "process_template_command"
        ) as mock_process:
            mock_process.return_value = {"success": False, "error": "Translation failed"}

            result = integration.generate_multilingual_descriptions(base_descriptions)

        # Should still have original English
        assert "test" in result
        assert result["test"]["en"] == "Original"

    def test_generate_descriptions_skips_unsupported_language(self):
        """Test that unsupported languages are skipped."""
        integration = ClaudeCLIIntegration()
        base_descriptions = {"test": "Original"}

        with patch(
            "moai_adk.core.claude_integration.get_language_info", return_value=None
        ):
            result = integration.generate_multilingual_descriptions(
                base_descriptions, ["en", "unsupported"]
            )

        assert "test" in result
        assert "en" in result["test"]
        # Unsupported language should not be in result
        assert "unsupported" not in result["test"]


class TestCreateAgentWithMultilingualSupport:
    """Test create_agent_with_multilingual_support method."""

    def test_create_agent_basic(self):
        """Test basic agent creation."""
        integration = ClaudeCLIIntegration()

        with patch.object(
            integration, "generate_multilingual_descriptions"
        ) as mock_gen:
            mock_gen.return_value = {
                "test-agent": {
                    "en": "Test agent description",
                    "ko": "테스트 에이전트",
                }
            }

            result = integration.create_agent_with_multilingual_support(
                agent_name="test-agent",
                base_description="Test agent description",
                tools=["read", "write"],
            )

        assert result["name"] == "test-agent"
        assert result["description"] == "Test agent description"
        assert "test-agent" in result["descriptions"]
        assert result["multilingual_support"] is True
        assert result["model"] == "sonnet"

    def test_create_agent_custom_model(self):
        """Test agent creation with custom model."""
        integration = ClaudeCLIIntegration()

        with patch.object(
            integration, "generate_multilingual_descriptions"
        ) as mock_gen:
            mock_gen.return_value = {"test": {"en": "Test"}}

            result = integration.create_agent_with_multilingual_support(
                agent_name="test",
                base_description="Test",
                tools=[],
                model="opus",
            )

        assert result["model"] == "opus"

    def test_create_agent_custom_languages(self):
        """Test agent creation with custom target languages."""
        integration = ClaudeCLIIntegration()

        with patch.object(
            integration, "generate_multilingual_descriptions"
        ) as mock_gen:
            mock_gen.return_value = {"test": {"en": "Test", "ko": "테스트"}}

            custom_langs = ["en", "ko"]
            integration.create_agent_with_multilingual_support(
                agent_name="test",
                base_description="Test",
                tools=[],
                target_languages=custom_langs,
            )

            # Verify the custom languages were passed
            mock_gen.assert_called_once()
            call_kwargs = mock_gen.call_args[1]
            assert call_kwargs.get("target_languages") == custom_langs


class TestCreateCommandWithMultilingualSupport:
    """Test create_command_with_multilingual_support method."""

    def test_create_command_basic(self):
        """Test basic command creation."""
        integration = ClaudeCLIIntegration()

        with patch.object(
            integration, "generate_multilingual_descriptions"
        ) as mock_gen:
            mock_gen.return_value = {"test-cmd": {"en": "Test command"}}

            result = integration.create_command_with_multilingual_support(
                command_name="test-cmd",
                base_description="Test command",
                argument_hint=["arg1", "arg2"],
                tools=["read"],
            )

        assert result["name"] == "test-cmd"
        assert result["description"] == "Test command"
        assert result["argument-hint"] == ["arg1", "arg2"]
        assert result["multilingual_support"] is True
        assert result["model"] == "haiku"

    def test_create_command_custom_model(self):
        """Test command creation with custom model."""
        integration = ClaudeCLIIntegration()

        with patch.object(
            integration, "generate_multilingual_descriptions"
        ) as mock_gen:
            mock_gen.return_value = {"test": {"en": "Test"}}

            result = integration.create_command_with_multilingual_support(
                command_name="test",
                base_description="Test",
                argument_hint=[],
                tools=[],
                model="sonnet",
            )

        assert result["model"] == "sonnet"


class TestProcessJsonStreamInput:
    """Test process_json_stream_input method."""

    def test_process_dict_input(self):
        """Test processing dictionary input."""
        integration = ClaudeCLIIntegration()
        input_data = {"key": "value", "number": 42}

        result = integration.process_json_stream_input(input_data)

        assert result == input_data

    def test_process_string_json_input(self):
        """Test processing JSON string input."""
        integration = ClaudeCLIIntegration()
        input_data = '{"key": "value", "number": 42}'

        result = integration.process_json_stream_input(input_data)

        assert result["key"] == "value"
        assert result["number"] == 42

    def test_process_invalid_json_string(self):
        """Test error handling for invalid JSON."""
        integration = ClaudeCLIIntegration()
        invalid_json = '{"key": invalid}'

        with pytest.raises(ValueError, match="Invalid JSON input"):
            integration.process_json_stream_input(invalid_json)

    def test_process_with_variables(self):
        """Test processing with variable substitution."""
        integration = ClaudeCLIIntegration()
        input_data = {"greeting": "Hello {{NAME}}"}
        variables = {"NAME": "World"}

        with patch.object(
            integration.template_engine,
            "render_string",
            return_value="Hello World",
        ):
            result = integration.process_json_stream_input(input_data, variables)

        assert result["greeting"] == "Hello World"

    def test_process_with_partial_variables(self):
        """Test processing when only some values have variables."""
        integration = ClaudeCLIIntegration()
        input_data = {"greeting": "Hello {{NAME}}", "static": "unchanged"}

        with patch.object(
            integration.template_engine, "render_string"
        ) as mock_render:
            mock_render.return_value = "Hello World"
            result = integration.process_json_stream_input(
                input_data, {"NAME": "World"}
            )

        # Only the field with variables should be processed
        assert mock_render.called


class TestExecuteHeadlessCommand:
    """Test execute_headless_command method."""

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_execute_headless_non_streaming(self, mock_subprocess_run):
        """Test headless command execution with non-streaming output."""
        mock_subprocess_run.return_value = Mock(
            returncode=0, stdout="line1\nline2\n", stderr=""
        )

        integration = ClaudeCLIIntegration()
        with patch.object(
            integration.template_engine, "render_string", return_value="prompt"
        ):
            result = integration.execute_headless_command(
                "test prompt", {}, output_format="json"
            )

        assert result["success"] is True
        assert isinstance(result["stdout"], list)
        assert len(result["stdout"]) == 2

    @patch("moai_adk.core.claude_integration.subprocess.Popen")
    def test_execute_headless_streaming(self, mock_popen):
        """Test headless command execution with streaming output."""
        mock_process = MagicMock()
        mock_process.stdout.readline.side_effect = [
            "line1\n",
            "line2\n",
            "",
        ]
        mock_process.stderr.readline.side_effect = ["", "", ""]
        mock_process.poll.side_effect = [None, None, 0]
        mock_popen.return_value = mock_process

        integration = ClaudeCLIIntegration()
        with patch.object(
            integration.template_engine, "render_string", return_value="prompt"
        ):
            result = integration.execute_headless_command(
                "test prompt", {}, output_format="stream-json"
            )

        assert result["success"] is True
        assert isinstance(result["stdout"], list)

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_execute_headless_with_additional_options(self, mock_subprocess_run):
        """Test headless execution with additional CLI options."""
        mock_subprocess_run.return_value = Mock(
            returncode=0, stdout="", stderr=""
        )

        integration = ClaudeCLIIntegration()
        with patch.object(
            integration.template_engine, "render_string", return_value="prompt"
        ):
            integration.execute_headless_command(
                "test",
                {},
                additional_options=["--timeout", "30"],
            )

        call_args = mock_subprocess_run.call_args[0][0]
        assert "--timeout" in call_args
        assert "30" in call_args

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_execute_headless_error(self, mock_subprocess_run):
        """Test error handling in headless execution."""
        mock_subprocess_run.side_effect = Exception("Execution failed")

        integration = ClaudeCLIIntegration()
        with patch.object(
            integration.template_engine, "render_string", return_value="prompt"
        ):
            result = integration.execute_headless_command("test", {})

        assert result["success"] is False
        assert "error" in result


class TestIntegrationWithRealTemplateEngine:
    """Integration tests with actual TemplateEngine."""

    def test_end_to_end_with_template_variables(self):
        """Test end-to-end processing with template variables."""
        integration = ClaudeCLIIntegration()

        with tempfile.TemporaryDirectory() as tmpdir:
            output_path = Path(tmpdir) / "settings.json"
            variables = {
                "PROJECT_NAME": "my-project",
                "CONVERSATION_LANGUAGE": "en",
                "CODEBASE_LANGUAGE": "python",
            }

            result_path = integration.generate_claude_settings(variables, output_path)

            assert result_path.exists()
            with open(result_path) as f:
                settings = json.load(f)

            assert settings["variables"]["PROJECT_NAME"] == "my-project"
            assert settings["template_context"]["codebase_language"] == "python"


class TestEdgeCases:
    """Test edge cases and boundary conditions."""

    def test_empty_variables_dict(self):
        """Test with empty variables dictionary."""
        integration = ClaudeCLIIntegration()

        with tempfile.TemporaryDirectory() as tmpdir:
            output_path = Path(tmpdir) / "settings.json"
            result_path = integration.generate_claude_settings({}, output_path)

            assert result_path.exists()

    def test_unicode_in_variables(self):
        """Test with unicode characters in variables."""
        integration = ClaudeCLIIntegration()
        variables = {"PROJECT_NAME": "프로젝트", "DESCRIPTION": "日本語"}

        with tempfile.TemporaryDirectory() as tmpdir:
            output_path = Path(tmpdir) / "settings.json"
            result_path = integration.generate_claude_settings(variables, output_path)

            with open(result_path, "r", encoding="utf-8") as f:
                settings = json.load(f)

            assert settings["variables"]["PROJECT_NAME"] == "프로젝트"

    @patch("moai_adk.core.claude_integration.subprocess.run")
    def test_large_output_handling(self, mock_subprocess_run):
        """Test handling of large subprocess output."""
        large_output = "x" * 100000
        mock_subprocess_run.return_value = Mock(
            returncode=0, stdout=large_output, stderr=""
        )

        integration = ClaudeCLIIntegration()
        with patch.object(
            integration.template_engine, "render_string", return_value="test"
        ):
            result = integration.process_template_command("test", {})

        assert result["success"] is True
        assert len(result["stdout"]) == len(large_output)
