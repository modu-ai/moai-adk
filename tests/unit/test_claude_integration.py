"""Test Claude CLI integration functionality."""

import json
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from src.moai_adk.core.claude_integration import ClaudeCLIIntegration
from src.moai_adk.core.template_engine import TemplateEngine


class TestClaudeCLIIntegration:
    """Test Claude CLI integration features."""

    def test_initialization(self):
        """Test ClaudeCLIIntegration initialization."""
        integration = ClaudeCLIIntegration()
        assert integration.template_engine is not None
        assert isinstance(integration.template_engine, TemplateEngine)

    def test_initialization_with_custom_engine(self):
        """Test initialization with custom template engine."""
        custom_engine = TemplateEngine(strict_undefined=False)
        integration = ClaudeCLIIntegration(custom_engine)
        assert integration.template_engine is custom_engine
        assert integration.template_engine.strict_undefined is False

    def test_generate_claude_settings(self):
        """Test Claude settings file generation."""
        integration = ClaudeCLIIntegration()
        variables = {
            "PROJECT_NAME": "TestProject",
            "CONVERSATION_LANGUAGE": "ko",
            "CONVERSATION_LANGUAGE_NAME": "한국어",
        }

        settings_path = integration.generate_claude_settings(variables)

        # Verify file was created
        assert settings_path.exists()

        # Verify content
        settings_content = json.loads(settings_path.read_text(encoding="utf-8"))
        assert "variables" in settings_content
        assert settings_content["variables"]["PROJECT_NAME"] == "TestProject"
        assert settings_content["variables"]["CONVERSATION_LANGUAGE"] == "ko"

        # Verify template context
        context = settings_content["template_context"]
        assert context["conversation_language"] == "ko"
        assert context["conversation_language_name"] == "한국어"
        assert context["project_name"] == "TestProject"

        # Cleanup
        settings_path.unlink()

    def test_generate_claude_settings_with_custom_path(self):
        """Test Claude settings generation with custom output path."""
        integration = ClaudeCLIIntegration()
        variables = {"PROJECT_NAME": "TestProject"}

        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            custom_path = Path(f.name)

        try:
            settings_path = integration.generate_claude_settings(variables, custom_path)
            assert settings_path == custom_path
            assert custom_path.exists()

            # Verify content
            content = json.loads(custom_path.read_text(encoding="utf-8"))
            assert content["variables"]["PROJECT_NAME"] == "TestProject"

        finally:
            if custom_path.exists():
                custom_path.unlink()

    @patch("subprocess.run")
    def test_process_template_command_success(self, mock_run):
        """Test successful template command processing."""
        # Mock subprocess result
        mock_result = MagicMock()
        mock_result.returncode = 0
        mock_result.stdout = '{"response": "success"}'
        mock_result.stderr = ""
        mock_run.return_value = mock_result

        integration = ClaudeCLIIntegration()
        command_template = "Create {{PROJECT_NAME}} in {{CONVERSATION_LANGUAGE}}"
        variables = {"PROJECT_NAME": "MyApp", "CONVERSATION_LANGUAGE": "ko"}

        result = integration.process_template_command(command_template, variables)

        assert result["success"] is True
        assert result["stdout"] == '{"response": "success"}'
        assert result["processed_command"] == "Create MyApp in ko"
        assert result["variables_used"] == variables

        # Verify subprocess was called correctly
        mock_run.assert_called_once()
        args, kwargs = mock_run.call_args
        assert "claude" in args[0]
        assert "--print" in args[0]
        assert "Create MyApp in ko" in args[0][-1]

    @patch("subprocess.run")
    def test_process_template_command_failure(self, mock_run):
        """Test template command processing failure."""
        # Mock subprocess result
        mock_result = MagicMock()
        mock_result.returncode = 1
        mock_result.stdout = ""
        mock_result.stderr = "Error occurred"
        mock_run.return_value = mock_result

        integration = ClaudeCLIIntegration()
        command_template = "Create {{PROJECT_NAME}}"
        variables = {"PROJECT_NAME": "MyApp"}

        result = integration.process_template_command(command_template, variables)

        assert result["success"] is False
        assert result["returncode"] == 1
        assert result["stderr"] == "Error occurred"

    @patch("subprocess.run")
    def test_process_template_command_with_different_formats(self, mock_run):
        """Test command processing with different output formats."""
        mock_result = MagicMock()
        mock_result.returncode = 0
        mock_result.stdout = "Plain text output"
        mock_result.stderr = ""
        mock_run.return_value = mock_result

        integration = ClaudeCLIIntegration()
        variables = {"PROJECT_NAME": "Test"}

        # Test with text format
        result = integration.process_template_command("Test command", variables, print_mode=True, output_format="text")

        assert result["success"] is True

        # Verify correct arguments
        args, _ = mock_run.call_args
        assert "--output-format" in args[0]
        format_index = args[0].index("--output-format")
        assert args[0][format_index + 1] == "text"

    def test_generate_multilingual_descriptions(self):
        """Test multilingual description generation."""
        integration = ClaudeCLIIntegration()
        base_descriptions = {"test_agent": "A test agent for automation", "helper_tool": "A helper tool for developers"}

        # Mock the translation result
        with patch.object(integration, "process_template_command") as mock_translate:
            # Mock translation responses
            mock_translate.side_effect = [
                {"success": True, "stdout": "한국어 설명"},
                {"success": True, "stdout": "한국어 도구 설명"},
                {"success": True, "stdout": "日本語の説明"},
                {"success": True, "stdout": "日本語のツール説明"},
            ]

            result = integration.generate_multilingual_descriptions(base_descriptions, ["en", "ko", "ja"])

            # Verify structure
            assert "test_agent" in result
            assert "helper_tool" in result

            # Verify English base
            assert result["test_agent"]["en"] == "A test agent for automation"
            assert result["helper_tool"]["en"] == "A helper tool for developers"

            # Verify translated versions (if successful)
            if result["test_agent"].get("ko"):
                assert result["test_agent"]["ko"] == "한국어 설명"

    def test_create_agent_with_multilingual_support(self):
        """Test agent creation with multilingual descriptions."""
        integration = ClaudeCLIIntegration()

        with patch.object(integration, "generate_multilingual_descriptions") as mock_gen:
            mock_gen.return_value = {
                "test-agent": {"en": "Base English description", "ko": "한국어 설명", "ja": "日本語の説明"}
            }

            agent_config = integration.create_agent_with_multilingual_support(
                agent_name="test-agent",
                base_description="Base English description",
                tools=["Read", "Write"],
                model="sonnet",
                target_languages=["en", "ko", "ja"],
            )

            # Verify agent configuration
            assert agent_config["name"] == "test-agent"
            assert agent_config["description"] == "Base English description"
            assert agent_config["tools"] == ["Read", "Write"]
            assert agent_config["model"] == "sonnet"
            assert agent_config["multilingual_support"] is True

            # Verify descriptions include all languages
            descriptions = agent_config["descriptions"]
            assert "en" in descriptions
            assert "ko" in descriptions
            assert "ja" in descriptions

    def test_create_command_with_multilingual_support(self):
        """Test command creation with multilingual descriptions."""
        integration = ClaudeCLIIntegration()

        with patch.object(integration, "generate_multilingual_descriptions") as mock_gen:
            mock_gen.return_value = {
                "test-command": {"en": "Test command description", "es": "Descripción del comando de prueba"}
            }

            command_config = integration.create_command_with_multilingual_support(
                command_name="test-command",
                base_description="Test command description",
                argument_hint=["project_name"],
                tools=["Bash"],
                model="haiku",
                target_languages=["en", "es"],
            )

            # Verify command configuration
            assert command_config["name"] == "test-command"
            assert command_config["description"] == "Test command description"
            assert command_config["argument-hint"] == ["project_name"]
            assert command_config["tools"] == ["Bash"]
            assert command_config["model"] == "haiku"
            assert command_config["multilingual_support"] is True

            # Verify descriptions
            descriptions = command_config["descriptions"]
            assert "en" in descriptions
            assert "es" in descriptions

    def test_process_json_stream_input_dict(self):
        """Test processing JSON stream input from dictionary."""
        integration = ClaudeCLIIntegration()
        input_data = {"message": "Hello {{PROJECT_NAME}}", "language": "{{CONVERSATION_LANGUAGE}}"}
        variables = {"PROJECT_NAME": "MyApp", "CONVERSATION_LANGUAGE": "ko"}

        result = integration.process_json_stream_input(input_data, variables)

        assert result["message"] == "Hello MyApp"
        assert result["language"] == "ko"

    def test_process_json_stream_input_string(self):
        """Test processing JSON stream input from JSON string."""
        integration = ClaudeCLIIntegration()
        input_json = '{"message": "Hello {{PROJECT_NAME}}"}'
        variables = {"PROJECT_NAME": "MyApp"}

        result = integration.process_json_stream_input(input_json, variables)

        assert result["message"] == "Hello MyApp"

    def test_process_json_stream_input_invalid_json(self):
        """Test processing invalid JSON string."""
        integration = ClaudeCLIIntegration()
        invalid_json = '{"invalid": json}'

        with pytest.raises(ValueError, match="Invalid JSON input"):
            integration.process_json_stream_input(invalid_json, {})

    @patch("subprocess.Popen")
    def test_execute_headless_command_streaming(self, mock_popen):
        """Test headless command execution with streaming."""
        # Mock streaming process
        mock_process = MagicMock()
        mock_process.poll.side_effect = [None, None, 0]  # Running, then finished
        mock_process.stdout.readline.side_effect = ['{"type": "start"}', '{"type": "progress", "value": 50}', ""]
        mock_process.stderr.readline.side_effect = ["", ""]
        mock_popen.return_value = mock_process

        integration = ClaudeCLIIntegration()
        prompt_template = "Process {{PROJECT_NAME}}"
        variables = {"PROJECT_NAME": "Test"}

        result = integration.execute_headless_command(
            prompt_template, variables, input_format="stream-json", output_format="stream-json"
        )

        assert result["success"] is True
        assert len(result["stdout"]) == 2
        assert '{"type": "start"}' in result["stdout"]
        assert '{"type": "progress"' in result["stdout"][1]

    @patch("subprocess.run")
    def test_execute_headless_command_non_streaming(self, mock_run):
        """Test headless command execution without streaming."""
        mock_result = MagicMock()
        mock_result.returncode = 0
        mock_result.stdout = '{"result": "success"}'
        mock_result.stderr = ""
        mock_run.return_value = mock_result

        integration = ClaudeCLIIntegration()
        prompt_template = "Process {{PROJECT_NAME}}"
        variables = {"PROJECT_NAME": "Test"}

        result = integration.execute_headless_command(
            prompt_template, variables, input_format="text", output_format="json"
        )

        assert result["success"] is True
        assert result["processed_prompt"] == "Process Test"
        assert result["variables"] == variables

    def test_execute_headless_command_with_additional_options(self):
        """Test headless command execution with additional CLI options."""
        integration = ClaudeCLIIntegration()

        with patch.object(integration, "process_template_command") as mock_process:
            mock_process.return_value = {"success": True}

            integration.execute_headless_command(
                "Test command", {"VAR": "value"}, additional_options=["--timeout", "30", "--verbose"]
            )

            # Verify additional options were passed through
            mock_process.assert_called_once()
            call_args = mock_process.call_args
            assert call_args[1]["additional_options"] == ["--timeout", "30", "--verbose"]
