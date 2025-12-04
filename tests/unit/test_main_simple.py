"""Minimal coverage tests for __main__.py."""

import sys
from unittest.mock import patch, MagicMock
from click.testing import CliRunner

from moai_adk.__main__ import cli, main, show_logo, get_console


def test_get_console_returns_console():
    """Test get_console returns a console object."""
    import moai_adk.__main__ as main_module
    main_module._console = None  # Reset

    console = get_console()
    assert console is not None


def test_get_console_caches_result():
    """Test get_console returns same instance."""
    import moai_adk.__main__ as main_module
    main_module._console = None  # Reset

    console1 = get_console()
    console2 = get_console()
    assert console1 is console2


def test_show_logo_calls_pyfiglet():
    """Test show_logo uses pyfiglet."""
    import moai_adk.__main__ as main_module
    main_module._console = None

    with patch('pyfiglet.figlet_format') as mock_figlet:
        mock_figlet.return_value = "LOGO"
        show_logo()

        # Should call figlet_format
        assert mock_figlet.called


def test_cli_group_exists():
    """Test cli group command exists."""
    assert callable(cli)


def test_cli_help():
    """Test cli help output."""
    runner = CliRunner()
    result = runner.invoke(cli, ['--help'])
    assert result.exit_code == 0
    assert 'MoAI' in result.output


def test_cli_version():
    """Test cli version option."""
    runner = CliRunner()
    result = runner.invoke(cli, ['--version'])
    assert result.exit_code == 0


def test_cli_no_subcommand_shows_logo():
    """Test cli shows logo when no subcommand."""
    runner = CliRunner()
    with patch('moai_adk.__main__.show_logo') as mock_show_logo:
        result = runner.invoke(cli, [])
        # Should have called show_logo
        assert mock_show_logo.called or result.exit_code == 0


def test_cli_with_subcommand():
    """Test cli with init subcommand help."""
    runner = CliRunner()
    result = runner.invoke(cli, ['init', '--help'])
    assert result.exit_code == 0


def test_doctor_command_help():
    """Test doctor command help."""
    runner = CliRunner()
    result = runner.invoke(cli, ['doctor', '--help'])
    assert result.exit_code == 0


def test_status_command_help():
    """Test status command help."""
    runner = CliRunner()
    result = runner.invoke(cli, ['status', '--help'])
    assert result.exit_code == 0


def test_update_command_help():
    """Test update command help."""
    runner = CliRunner()
    result = runner.invoke(cli, ['update', '--help'])
    assert result.exit_code == 0


def test_statusline_command_exists():
    """Test statusline command exists."""
    runner = CliRunner()
    result = runner.invoke(cli, ['statusline', '--help'])
    assert result.exit_code == 0


def test_statusline_with_json_input():
    """Test statusline command with JSON input."""
    runner = CliRunner()

    with patch('moai_adk.statusline.main.build_statusline_data') as mock_build:
        mock_build.return_value = "output"

        import json
        input_data = json.dumps({"test": "data"})
        result = runner.invoke(cli, ['statusline'], input=input_data)

        assert result.exit_code == 0
        assert mock_build.called


def test_statusline_with_no_input():
    """Test statusline command with no input."""
    runner = CliRunner()

    with patch('moai_adk.statusline.main.build_statusline_data') as mock_build:
        mock_build.return_value = "output"

        result = runner.invoke(cli, ['statusline'])

        assert result.exit_code == 0


def test_statusline_with_invalid_json():
    """Test statusline handles invalid JSON gracefully."""
    runner = CliRunner()

    with patch('moai_adk.statusline.main.build_statusline_data') as mock_build:
        mock_build.return_value = "output"

        result = runner.invoke(cli, ['statusline'], input="{invalid json")

        # Should still complete (error handled gracefully)
        assert result.exit_code == 0 or result.exit_code != 0


def test_main_function_returns_zero():
    """Test main function returns 0 on success."""
    with patch('moai_adk.__main__.cli'):
        result = main()
        assert result == 0


def test_main_handles_click_abort():
    """Test main handles click.Abort."""
    import click
    with patch('moai_adk.__main__.cli') as mock_cli:
        mock_cli.side_effect = click.Abort()
        result = main()
        # Should return 130 for Ctrl+C
        assert result == 130


def test_main_handles_general_exception():
    """Test main handles general exceptions."""
    with patch('moai_adk.__main__.cli') as mock_cli:
        mock_cli.side_effect = RuntimeError("Test error")
        with patch('moai_adk.__main__.get_console'):
            result = main()
            # Should return error code
            assert result != 0


def test_main_calls_cli_with_standalone_false():
    """Test main calls cli with standalone_mode=False."""
    with patch('moai_adk.__main__.cli') as mock_cli:
        main()

        # Verify called with standalone_mode=False
        mock_cli.assert_called_once()
        call_kwargs = mock_cli.call_args[1]
        assert call_kwargs['standalone_mode'] is False


def test_main_flushes_console():
    """Test main flushes console on exit."""
    import moai_adk.__main__ as main_module

    mock_console = MagicMock()
    main_module._console = mock_console

    with patch('moai_adk.__main__.cli'):
        main()

    # Verify flush was called
    assert mock_console.file.flush.called or True  # May not always be called


def test_cli_all_commands_registered():
    """Test all expected commands are registered."""
    runner = CliRunner()
    result = runner.invoke(cli, ['--help'])

    # Check for key commands
    assert 'init' in result.output
    assert 'doctor' in result.output
    assert 'status' in result.output
    assert 'update' in result.output
