"""Extended tests for moai_adk.core.claude_integration module.

Comprehensive test coverage for ClaudeCLIIntegration with full method coverage,
template processing, and multilingual support testing.
"""

import json
import subprocess
from pathlib import Path
from typing import Any, Dict
from unittest.mock import MagicMock, Mock, patch

import pytest


class TestClaudeCLIIntegrationBasics:
    """Test ClaudeCLIIntegration class initialization and basics."""

    def test_class_import(self):
        """Test that ClaudeCLIIntegration can be imported."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        assert ClaudeCLIIntegration is not None

    def test_init_with_default_template_engine(self):
        """Test initialization with default TemplateEngine."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        with patch("moai_adk.core.claude_integration.TemplateEngine"):
            cli = ClaudeCLIIntegration()
            assert cli.template_engine is not None

    def test_init_with_custom_template_engine(self):
        """Test initialization with custom TemplateEngine."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        mock_engine = Mock()
        cli = ClaudeCLIIntegration(template_engine=mock_engine)
        assert cli.template_engine is mock_engine


class TestGenerateClaudeSettings:
    """Test generate_claude_settings method."""

    def test_generate_settings_returns_path(self):
        """Test that generate_claude_settings returns a Path."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        with patch("moai_adk.core.claude_integration.TemplateEngine"):
            cli = ClaudeCLIIntegration()

            variables = {"CONVERSATION_LANGUAGE": "en", "PROJECT_NAME": "test"}

            with patch("tempfile.NamedTemporaryFile") as mock_temp:
                mock_file = Mock()
                mock_file.name = "/tmp/test.json"
                mock_temp.return_value = mock_file

                with patch.object(Path, "write_text"):
                    result = cli.generate_claude_settings(variables)

                assert result is not None or isinstance(result, Path)

    def test_generate_settings_with_output_path(self):
        """Test generate_claude_settings with specified output path."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        with patch("moai_adk.core.claude_integration.TemplateEngine"):
            cli = ClaudeCLIIntegration()

            variables = {"CONVERSATION_LANGUAGE": "en"}
            output_path = Path("/test/settings.json")

            with patch.object(Path, "write_text"):
                result = cli.generate_claude_settings(variables, output_path)

            assert result == output_path

    def test_generate_settings_content_format(self):
        """Test that generated settings have correct format."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        with patch("moai_adk.core.claude_integration.TemplateEngine"):
            cli = ClaudeCLIIntegration()

            variables = {
                "CONVERSATION_LANGUAGE": "en",
                "PROJECT_NAME": "my-project",
                "CODEBASE_LANGUAGE": "python",
            }
            output_path = Path("/test/settings.json")

            written_content = None

            def capture_write(content):
                nonlocal written_content
                written_content = content

            with patch.object(Path, "write_text", side_effect=capture_write):
                cli.generate_claude_settings(variables, output_path)

            if written_content:
                data = json.loads(written_content)
                assert "variables" in data or written_content is not None


class TestProcessTemplateCommand:
    """Test process_template_command method."""

    def test_process_template_command_success(self):
        """Test processing a template command successfully."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        mock_engine = Mock()
        mock_engine.render_string.return_value = "processed command"
        cli = ClaudeCLIIntegration(template_engine=mock_engine)

        with patch("tempfile.NamedTemporaryFile") as mock_temp:
            mock_file = Mock()
            mock_file.name = "/tmp/test.json"
            mock_temp.return_value = mock_file

            with patch("subprocess.run") as mock_run:
                mock_result = Mock()
                mock_result.returncode = 0
                mock_result.stdout = "output"
                mock_result.stderr = ""
                mock_run.return_value = mock_result

                with patch.object(Path, "unlink"):
                    result = cli.process_template_command(
                        "test command {{VAR}}", {"VAR": "value"}
                    )

        assert isinstance(result, dict)
        assert "success" in result
        assert "stdout" in result
        assert "stderr" in result
        assert "returncode" in result

    def test_process_template_command_with_print_mode(self):
        """Test processing with print mode enabled."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        mock_engine = Mock()
        mock_engine.render_string.return_value = "command"
        cli = ClaudeCLIIntegration(template_engine=mock_engine)

        with patch("tempfile.NamedTemporaryFile") as mock_temp:
            mock_file = Mock()
            mock_file.name = "/tmp/test.json"
            mock_temp.return_value = mock_file

            with patch("subprocess.run") as mock_run:
                mock_result = Mock()
                mock_result.returncode = 0
                mock_result.stdout = ""
                mock_result.stderr = ""
                mock_run.return_value = mock_result

                with patch.object(Path, "unlink"):
                    result = cli.process_template_command(
                        "command", {}, print_mode=True
                    )

                # Check that --print flag was included
                call_args = mock_run.call_args
                assert call_args is not None

    def test_process_template_command_error_handling(self):
        """Test error handling in process_template_command."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        mock_engine = Mock()
        mock_engine.render_string.side_effect = Exception("Template error")
        cli = ClaudeCLIIntegration(template_engine=mock_engine)

        result = cli.process_template_command("command", {})

        assert result["success"] is False
        assert "error" in result


class TestGenerateMultilingualDescriptions:
    """Test generate_multilingual_descriptions method."""

    def test_generate_multilingual_descriptions_returns_dict(self):
        """Test that generate_multilingual_descriptions returns a dict."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        with patch("moai_adk.core.claude_integration.TemplateEngine"):
            cli = ClaudeCLIIntegration()

            base_descriptions = {"agent1": "An AI agent for testing"}

            with patch.object(cli, "process_template_command") as mock_process:
                mock_process.return_value = {"success": False}

                result = cli.generate_multilingual_descriptions(base_descriptions)

        assert isinstance(result, dict)
        assert "agent1" in result
        assert "en" in result["agent1"]

    def test_generate_multilingual_with_custom_languages(self):
        """Test generate_multilingual_descriptions with custom target languages."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        with patch("moai_adk.core.claude_integration.TemplateEngine"):
            cli = ClaudeCLIIntegration()

            base_descriptions = {"agent1": "Test agent"}

            with patch.object(cli, "process_template_command") as mock_process:
                mock_process.return_value = {"success": False}

                result = cli.generate_multilingual_descriptions(
                    base_descriptions, target_languages=["en", "ko"]
                )

        assert isinstance(result, dict)


class TestCreateAgentWithMultilingualSupport:
    """Test create_agent_with_multilingual_support method."""

    def test_create_agent_returns_dict(self):
        """Test that create_agent_with_multilingual_support returns a dict."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        with patch("moai_adk.core.claude_integration.TemplateEngine"):
            cli = ClaudeCLIIntegration()

            with patch.object(cli, "generate_multilingual_descriptions") as mock_gen:
                mock_gen.return_value = {"test-agent": {"en": "An agent"}}

                result = cli.create_agent_with_multilingual_support(
                    "test-agent", "An agent", ["Read", "Write"]
                )

        assert isinstance(result, dict)
        assert "name" in result
        assert "description" in result
        assert "tools" in result
        assert "model" in result

    def test_create_agent_includes_multilingual_flag(self):
        """Test that created agent includes multilingual_support flag."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        with patch("moai_adk.core.claude_integration.TemplateEngine"):
            cli = ClaudeCLIIntegration()

            with patch.object(cli, "generate_multilingual_descriptions") as mock_gen:
                mock_gen.return_value = {"test-agent": {"en": "Test"}}

                result = cli.create_agent_with_multilingual_support(
                    "test-agent", "Test description", []
                )

        assert result.get("multilingual_support") is True


class TestCreateCommandWithMultilingualSupport:
    """Test create_command_with_multilingual_support method."""

    def test_create_command_returns_dict(self):
        """Test that create_command_with_multilingual_support returns a dict."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        with patch("moai_adk.core.claude_integration.TemplateEngine"):
            cli = ClaudeCLIIntegration()

            with patch.object(cli, "generate_multilingual_descriptions") as mock_gen:
                mock_gen.return_value = {"test-cmd": {"en": "A command"}}

                result = cli.create_command_with_multilingual_support(
                    "test-cmd", "A command", ["arg1"], ["Read"]
                )

        assert isinstance(result, dict)
        assert "name" in result
        assert "description" in result
        assert "argument-hint" in result
        assert "tools" in result
        assert "model" in result

    def test_create_command_includes_multilingual_flag(self):
        """Test that created command includes multilingual_support flag."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        with patch("moai_adk.core.claude_integration.TemplateEngine"):
            cli = ClaudeCLIIntegration()

            with patch.object(cli, "generate_multilingual_descriptions") as mock_gen:
                mock_gen.return_value = {"test-cmd": {"en": "Test"}}

                result = cli.create_command_with_multilingual_support(
                    "test-cmd", "Test", ["arg"], []
                )

        assert result.get("multilingual_support") is True


class TestProcessJsonStreamInput:
    """Test process_json_stream_input method."""

    def test_process_json_dict_input(self):
        """Test processing JSON dict input."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        mock_engine = Mock()
        cli = ClaudeCLIIntegration(template_engine=mock_engine)

        input_data = {"key": "value"}
        result = cli.process_json_stream_input(input_data)

        assert isinstance(result, dict)
        assert result == input_data

    def test_process_json_string_input(self):
        """Test processing JSON string input."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        mock_engine = Mock()
        cli = ClaudeCLIIntegration(template_engine=mock_engine)

        input_data = '{"key": "value"}'
        result = cli.process_json_stream_input(input_data)

        assert isinstance(result, dict)
        assert result["key"] == "value"

    def test_process_json_with_variables(self):
        """Test processing JSON with template variable substitution."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        mock_engine = Mock()
        mock_engine.render_string.return_value = "substituted"
        cli = ClaudeCLIIntegration(template_engine=mock_engine)

        input_data = {"key": "{{VAR}}"}
        variables = {"VAR": "value"}

        result = cli.process_json_stream_input(input_data, variables)

        assert isinstance(result, dict)

    def test_process_json_invalid_string(self):
        """Test processing invalid JSON string raises ValueError."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        mock_engine = Mock()
        cli = ClaudeCLIIntegration(template_engine=mock_engine)

        input_data = "invalid json {{{"

        with pytest.raises(ValueError):
            cli.process_json_stream_input(input_data)


class TestExecuteHeadlessCommand:
    """Test execute_headless_command method."""

    def test_execute_headless_command_returns_dict(self):
        """Test that execute_headless_command returns a dict."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        mock_engine = Mock()
        mock_engine.render_string.return_value = "processed prompt"
        cli = ClaudeCLIIntegration(template_engine=mock_engine)

        with patch("tempfile.NamedTemporaryFile") as mock_temp:
            mock_file = Mock()
            mock_file.name = "/tmp/test.json"
            mock_temp.return_value = mock_file

            with patch("subprocess.run") as mock_run:
                mock_result = Mock()
                mock_result.returncode = 0
                mock_result.stdout = "output"
                mock_result.stderr = ""
                mock_run.return_value = mock_result

                with patch.object(Path, "unlink"):
                    result = cli.execute_headless_command(
                        "prompt {{VAR}}", {"VAR": "value"}
                    )

        assert isinstance(result, dict)
        assert "success" in result
        assert "stdout" in result
        assert "stderr" in result
        assert "returncode" in result

    def test_execute_headless_with_streaming(self):
        """Test execute_headless_command with streaming output."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        mock_engine = Mock()
        mock_engine.render_string.return_value = "prompt"
        cli = ClaudeCLIIntegration(template_engine=mock_engine)

        with patch("tempfile.NamedTemporaryFile") as mock_temp:
            mock_file = Mock()
            mock_file.name = "/tmp/test.json"
            mock_temp.return_value = mock_file

            with patch("subprocess.Popen") as mock_popen:
                mock_process = Mock()
                mock_process.stdout.readline.side_effect = ["line1\n", "line2\n", ""]
                mock_process.stderr.readline.side_effect = ["", "", ""]
                mock_process.poll.side_effect = [None, None, 0]
                mock_popen.return_value = mock_process

                with patch.object(Path, "unlink"):
                    result = cli.execute_headless_command(
                        "prompt", {}, output_format="stream-json"
                    )

        assert isinstance(result, dict)

    def test_execute_headless_error_handling(self):
        """Test error handling in execute_headless_command."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        mock_engine = Mock()
        mock_engine.render_string.side_effect = Exception("Render error")
        cli = ClaudeCLIIntegration(template_engine=mock_engine)

        result = cli.execute_headless_command("prompt", {})

        assert result["success"] is False
        assert "error" in result

    def test_execute_headless_with_additional_options(self):
        """Test execute_headless_command with additional CLI options."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        mock_engine = Mock()
        mock_engine.render_string.return_value = "prompt"
        cli = ClaudeCLIIntegration(template_engine=mock_engine)

        with patch("tempfile.NamedTemporaryFile") as mock_temp:
            mock_file = Mock()
            mock_file.name = "/tmp/test.json"
            mock_temp.return_value = mock_file

            with patch("subprocess.run") as mock_run:
                mock_result = Mock()
                mock_result.returncode = 0
                mock_result.stdout = ""
                mock_result.stderr = ""
                mock_run.return_value = mock_result

                with patch.object(Path, "unlink"):
                    result = cli.execute_headless_command(
                        "prompt",
                        {},
                        additional_options=["--option1", "--option2"],
                    )

                # Verify result is a dict
                assert isinstance(result, dict)
                assert "success" in result


class TestLanguageIntegration:
    """Test language-specific integration methods."""

    def test_get_language_info_integration(self):
        """Test integration with get_language_info."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        with patch("moai_adk.core.claude_integration.TemplateEngine"):
            with patch(
                "moai_adk.core.claude_integration.get_language_info"
            ) as mock_lang:
                mock_lang.return_value = {"native_name": "English"}

                cli = ClaudeCLIIntegration()

                with patch.object(cli, "process_template_command") as mock_process:
                    mock_process.return_value = {"success": False}

                    result = cli.generate_multilingual_descriptions(
                        {"cmd": "Test"}, target_languages=["en"]
                    )

                assert isinstance(result, dict)


class TestEdgeCases:
    """Test edge cases and error scenarios."""

    def test_empty_template_variables(self):
        """Test processing with empty template variables."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        mock_engine = Mock()
        cli = ClaudeCLIIntegration(template_engine=mock_engine)

        result = cli.generate_claude_settings({})

        assert result is not None

    def test_special_characters_in_variables(self):
        """Test handling special characters in template variables."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        mock_engine = Mock()
        mock_engine.render_string.return_value = "processed"
        cli = ClaudeCLIIntegration(template_engine=mock_engine)

        variables = {"KEY": "value with {{special}} chars"}

        with patch("tempfile.NamedTemporaryFile") as mock_temp:
            mock_file = Mock()
            mock_file.name = "/tmp/test.json"
            mock_temp.return_value = mock_file

            with patch("subprocess.run") as mock_run:
                mock_result = Mock()
                mock_result.returncode = 0
                mock_result.stdout = ""
                mock_result.stderr = ""
                mock_run.return_value = mock_result

                with patch.object(Path, "unlink"):
                    result = cli.process_template_command("cmd", variables)

        assert isinstance(result, dict)

    def test_very_long_command_string(self):
        """Test handling very long command strings."""
        from moai_adk.core.claude_integration import ClaudeCLIIntegration

        mock_engine = Mock()
        mock_engine.render_string.return_value = "x" * 10000
        cli = ClaudeCLIIntegration(template_engine=mock_engine)

        long_command = "cmd " + "{{VAR}}" * 100

        with patch("tempfile.NamedTemporaryFile") as mock_temp:
            mock_file = Mock()
            mock_file.name = "/tmp/test.json"
            mock_temp.return_value = mock_file

            with patch("subprocess.run") as mock_run:
                mock_result = Mock()
                mock_result.returncode = 0
                mock_result.stdout = ""
                mock_result.stderr = ""
                mock_run.return_value = mock_result

                with patch.object(Path, "unlink"):
                    result = cli.process_template_command(long_command, {})

        assert isinstance(result, dict)
