"""Minimal coverage tests for language command."""

from unittest.mock import patch, MagicMock
from click.testing import CliRunner

from moai_adk.cli.commands.language import (
    language,
    list as list_cmd,
    info,
    render_template,
    validate_config,
)


def test_language_group_exists():
    """Test language command group exists."""
    assert callable(language)


def test_list_languages_command():
    """Test list languages subcommand works."""
    runner = CliRunner()
    result = runner.invoke(list_cmd, [])
    assert result.exit_code == 0
    assert "Supported Languages" in result.output or "English Name" in result.output


def test_list_languages_json_output():
    """Test list command with JSON output."""
    runner = CliRunner()
    result = runner.invoke(list_cmd, ["--json-output"])
    assert result.exit_code == 0


def test_info_command_valid_language():
    """Test info command with valid language code."""
    runner = CliRunner()
    result = runner.invoke(info, ["en"])
    assert result.exit_code == 0


def test_info_command_detail_flag():
    """Test info command with detail flag."""
    runner = CliRunner()
    result = runner.invoke(info, ["en", "--detail"])
    assert result.exit_code == 0


def test_info_command_invalid_language():
    """Test info command rejects invalid language code."""
    runner = CliRunner()
    result = runner.invoke(info, ["invalid_xyz"])
    # Should either fail or show error message
    assert "not found" in result.output or "Available codes" in result.output


def test_render_template_with_files():
    """Test render template command."""
    import tempfile
    import json
    from pathlib import Path

    runner = CliRunner()
    with tempfile.TemporaryDirectory() as tmpdir:
        # Create template file
        template_file = Path(tmpdir) / "test.txt"
        template_file.write_text("Hello {{name}}")

        # Create variables file
        vars_file = Path(tmpdir) / "vars.json"
        vars_file.write_text(json.dumps({"name": "World"}))

        result = runner.invoke(render_template, [str(template_file), str(vars_file)])

        # Should show rendered template
        assert result.exit_code == 0


def test_validate_config_with_valid_file():
    """Test validate_config command."""
    import tempfile
    import json
    from pathlib import Path

    runner = CliRunner()
    with tempfile.TemporaryDirectory() as tmpdir:
        # Create config file
        config_file = Path(tmpdir) / "config.json"
        config_file.write_text(
            json.dumps({"language": {"conversation_language": "en"}})
        )

        result = runner.invoke(validate_config, [str(config_file)])

        # Should validate successfully
        assert result.exit_code == 0


def test_validate_config_missing_language_section():
    """Test validate_config with missing language section."""
    import tempfile
    import json
    from pathlib import Path

    runner = CliRunner()
    with tempfile.TemporaryDirectory() as tmpdir:
        # Create config without language section
        config_file = Path(tmpdir) / "config.json"
        config_file.write_text(json.dumps({}))

        result = runner.invoke(validate_config, [str(config_file)])

        # Should report missing language section
        assert result.exit_code == 0
        assert "No 'language' section" in result.output


def test_validate_config_invalid_json():
    """Test validate_config with invalid JSON."""
    import tempfile
    from pathlib import Path

    runner = CliRunner()
    with tempfile.TemporaryDirectory() as tmpdir:
        # Create invalid JSON file
        config_file = Path(tmpdir) / "config.json"
        config_file.write_text("{invalid json")

        result = runner.invoke(validate_config, [str(config_file)])

        # Should report error
        assert result.exit_code != 0 or "Error" in result.output


def test_translate_descriptions_mocked():
    """Test translate_descriptions command with mock."""
    from moai_adk.cli.commands.language import translate_descriptions

    runner = CliRunner()

    with patch(
        "moai_adk.cli.commands.language.ClaudeCLIIntegration"
    ) as mock_integration_class:
        mock_integration = MagicMock()
        mock_integration_class.return_value = mock_integration
        mock_integration.generate_multilingual_descriptions.return_value = {
            "en": "English",
            "ko": "Korean",
        }

        result = runner.invoke(translate_descriptions, ["Test description"])

        assert result.exit_code == 0


def test_execute_command_dry_run():
    """Test execute command in dry-run mode."""
    from moai_adk.cli.commands.language import execute

    runner = CliRunner()

    result = runner.invoke(execute, ["test prompt", "--dry-run"])

    assert result.exit_code == 0
    assert "Dry Run" in result.output


def test_execute_command_with_language():
    """Test execute command with language option."""
    from moai_adk.cli.commands.language import execute

    runner = CliRunner()

    with patch(
        "moai_adk.cli.commands.language.ClaudeCLIIntegration"
    ) as mock_integration_class:
        mock_integration = MagicMock()
        mock_integration_class.return_value = mock_integration
        mock_integration.process_template_command.return_value = {
            "success": True,
            "stdout": "result",
            "stderr": "",
        }

        result = runner.invoke(execute, ["test prompt", "--language", "en"])

        assert result.exit_code == 0
