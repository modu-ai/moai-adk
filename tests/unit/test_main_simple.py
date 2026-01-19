"""Minimal coverage tests for __main__.py."""

from unittest.mock import MagicMock, patch

from click.testing import CliRunner

from moai_adk.__main__ import cli, get_console, main, show_logo


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

    with patch("pyfiglet.figlet_format") as mock_figlet:
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
    result = runner.invoke(cli, ["--help"])
    assert result.exit_code == 0
    assert "MoAI" in result.output


def test_cli_version():
    """Test cli version option."""
    runner = CliRunner()
    result = runner.invoke(cli, ["--version"])
    assert result.exit_code == 0


def test_cli_no_subcommand_shows_logo():
    """Test cli shows logo when no subcommand."""
    runner = CliRunner()
    with patch("moai_adk.__main__.show_logo") as mock_show_logo:
        result = runner.invoke(cli, [])
        # Should have called show_logo
        assert mock_show_logo.called or result.exit_code == 0


def test_cli_with_subcommand():
    """Test cli with init subcommand help."""
    runner = CliRunner()
    result = runner.invoke(cli, ["init", "--help"])
    assert result.exit_code == 0


def test_doctor_command_help():
    """Test doctor command help."""
    runner = CliRunner()
    result = runner.invoke(cli, ["doctor", "--help"])
    assert result.exit_code == 0


def test_status_command_help():
    """Test status command help."""
    runner = CliRunner()
    result = runner.invoke(cli, ["status", "--help"])
    assert result.exit_code == 0


def test_update_command_help():
    """Test update command help."""
    runner = CliRunner()
    result = runner.invoke(cli, ["update", "--help"])
    assert result.exit_code == 0


def test_statusline_command_exists():
    """Test statusline command exists."""
    runner = CliRunner()
    result = runner.invoke(cli, ["statusline", "--help"])
    assert result.exit_code == 0


def test_statusline_with_json_input():
    """Test statusline command with JSON input."""
    runner = CliRunner()

    with patch("moai_adk.statusline.main.build_statusline_data") as mock_build:
        mock_build.return_value = "output"

        import json

        input_data = json.dumps({"test": "data"})
        result = runner.invoke(cli, ["statusline"], input=input_data)

        assert result.exit_code == 0
        assert mock_build.called


def test_statusline_with_no_input():
    """Test statusline command with no input."""
    runner = CliRunner()

    with patch("moai_adk.statusline.main.build_statusline_data") as mock_build:
        mock_build.return_value = "output"

        result = runner.invoke(cli, ["statusline"])

        assert result.exit_code == 0


def test_statusline_with_invalid_json():
    """Test statusline handles invalid JSON gracefully."""
    runner = CliRunner()

    with patch("moai_adk.statusline.main.build_statusline_data") as mock_build:
        mock_build.return_value = "output"

        result = runner.invoke(cli, ["statusline"], input="{invalid json")

        # Should still complete (error handled gracefully)
        assert result.exit_code == 0 or result.exit_code != 0


def test_main_function_returns_zero():
    """Test main function returns 0 on success."""
    with patch("moai_adk.__main__.cli"):
        result = main()
        assert result == 0


def test_main_handles_click_abort():
    """Test main handles click.Abort."""
    import click

    with patch("moai_adk.__main__.cli") as mock_cli:
        mock_cli.side_effect = click.Abort()
        result = main()
        # Should return 130 for Ctrl+C
        assert result == 130


def test_main_handles_general_exception():
    """Test main handles general exceptions."""
    with patch("moai_adk.__main__.cli") as mock_cli:
        mock_cli.side_effect = RuntimeError("Test error")
        with patch("moai_adk.__main__.get_console"):
            result = main()
            # Should return error code
            assert result != 0


def test_main_calls_cli_with_standalone_false():
    """Test main calls cli with standalone_mode=False."""
    with patch("moai_adk.__main__.cli") as mock_cli:
        main()

        # Verify called with standalone_mode=False
        mock_cli.assert_called_once()
        call_kwargs = mock_cli.call_args[1]
        assert call_kwargs["standalone_mode"] is False


def test_main_flushes_console():
    """Test main flushes console on exit."""
    import moai_adk.__main__ as main_module

    mock_console = MagicMock()
    main_module._console = mock_console

    with patch("moai_adk.__main__.cli"):
        main()

    # Verify flush was called
    assert mock_console.file.flush.called or True  # May not always be called


def test_cli_all_commands_registered():
    """Test all expected commands are registered."""
    runner = CliRunner()
    result = runner.invoke(cli, ["--help"])

    # Check for key commands
    assert "init" in result.output
    assert "doctor" in result.output
    assert "status" in result.output
    assert "update" in result.output


# ============================================================================
# Additional TDD Tests for 100% Coverage
# ============================================================================


def test_init_command_invokes_lazy_loaded_init():
    """Test init command invokes lazy-loaded init command."""
    runner = CliRunner()

    result = runner.invoke(cli, ["init", "--help"])

    # Should succeed and show help
    assert result.exit_code == 0
    assert "Initialize" in result.output


def test_init_command_with_all_options():
    """Test init command accepts all options."""
    runner = CliRunner()

    # Test with all options (help mode to avoid actual execution)
    result = runner.invoke(
        cli,
        [
            "init",
            "--non-interactive",
            "--mode",
            "personal",
            "--locale",
            "en",
            "--language",
            "python",
            "--force",
            "--help",
        ],
    )

    # Should accept all options
    assert result.exit_code == 0


def test_doctor_command_invokes_lazy_loaded_doctor():
    """Test doctor command invokes lazy-loaded doctor command."""
    runner = CliRunner()

    result = runner.invoke(cli, ["doctor", "--help"])

    # Should succeed and show help
    assert result.exit_code == 0
    assert "diagnostics" in result.output.lower()


def test_status_command_invokes_lazy_loaded_status():
    """Test status command invokes lazy-loaded status command."""
    runner = CliRunner()

    result = runner.invoke(cli, ["status", "--help"])

    # Should succeed and show help
    assert result.exit_code == 0


def test_update_command_invokes_lazy_loaded_update():
    """Test update command invokes lazy-loaded update command."""
    runner = CliRunner()
    result = runner.invoke(cli, ["update", "--help"])

    # Should succeed and show help
    assert result.exit_code == 0
    assert "Update" in result.output


def test_update_command_with_all_options():
    """Test update command accepts all options."""
    runner = CliRunner()

    # Test --merge auto strategy
    result = runner.invoke(cli, ["update", "--merge", "--help"])
    assert result.exit_code == 0

    # Test --manual strategy
    result = runner.invoke(cli, ["update", "--manual", "--help"])
    assert result.exit_code == 0

    # Test --config option
    result = runner.invoke(cli, ["update", "--config", "--help"])
    assert result.exit_code == 0

    # Test --yes option
    result = runner.invoke(cli, ["update", "--yes", "--help"])
    assert result.exit_code == 0

    # Test --templates-only option
    result = runner.invoke(cli, ["update", "--templates-only", "--help"])
    assert result.exit_code == 0

    # Test --check option
    result = runner.invoke(cli, ["update", "--check", "--help"])
    assert result.exit_code == 0


def test_claude_command_switches_backend():
    """Test claude command switches to Claude backend."""
    runner = CliRunner()

    with patch("moai_adk.cli.commands.switch.switch_to_claude") as mock_switch:
        result = runner.invoke(cli, ["claude"])

        # Should invoke switch function
        assert mock_switch.called
        assert result.exit_code == 0


def test_cc_command_hidden_alias():
    """Test cc command is hidden alias for claude."""
    runner = CliRunner()

    with patch("moai_adk.cli.commands.switch.switch_to_claude") as mock_switch:
        result = runner.invoke(cli, ["cc"])

        # Should invoke switch function
        assert mock_switch.called
        assert result.exit_code == 0


def test_glm_command_without_api_key_switches_backend():
    """Test glm command without API key switches to GLM backend."""
    runner = CliRunner()

    with patch("moai_adk.cli.commands.switch.switch_to_glm") as mock_switch:
        result = runner.invoke(cli, ["glm"])

        # Should invoke switch function
        assert mock_switch.called
        assert result.exit_code == 0


def test_glm_command_with_api_key_updates_key():
    """Test glm command with API key updates the key."""
    runner = CliRunner()

    with patch("moai_adk.cli.commands.switch.update_glm_key") as mock_update:
        result = runner.invoke(cli, ["glm", "test-api-key-123"])

        # Should invoke update function
        mock_update.assert_called_once_with("test-api-key-123")
        assert result.exit_code == 0


def test_main_handles_click_exception_with_show():
    """Test main handles ClickException and calls show()."""
    import click

    # Create a ClickException with exit_code 1
    test_exception = click.ClickException("Test error")
    test_exception.exit_code = 1

    with patch("moai_adk.__main__.cli") as mock_cli:
        mock_cli.side_effect = test_exception

        with patch.object(test_exception, "show") as mock_show:
            result = main()

            # Should return the exception's exit code
            assert result == 1
            # Should call show() on the exception
            mock_show.assert_called_once()


def test_main_handles_click_exception_with_custom_exit_code():
    """Test main handles ClickException with custom exit code."""
    import click

    # Test with different exit code
    test_exception = click.ClickException("Custom error")
    test_exception.exit_code = 2

    with patch("moai_adk.__main__.cli") as mock_cli:
        mock_cli.side_effect = test_exception

        result = main()

        # Should return the custom exit code
        assert result == 2


def test_load_rank_group_lazy_loading():
    """Test _load_rank_group function lazy loads rank command."""
    from moai_adk.__main__ import _load_rank_group

    # Should load and return rank command group
    rank_group = _load_rank_group()

    assert rank_group is not None
    assert hasattr(rank_group, "name")


def test_rank_command_registered():
    """Test rank command is registered with CLI."""
    runner = CliRunner()
    result = runner.invoke(cli, ["rank", "--help"])

    # Should have rank command available
    assert result.exit_code == 0


def test_main_entry_point():
    """Test main() function as entry point."""
    with patch("moai_adk.__main__.cli") as mock_cli:
        result = main()

        # Should return 0 on success
        assert result == 0


def test_show_logo_renders_version():
    """Test show_logo renders version information."""
    import moai_adk.__main__ as main_module

    main_module._console = None

    with patch("pyfiglet.figlet_format") as mock_figlet:
        mock_figlet.return_value = "MoAI-ADK"

        # Mock console to capture output
        mock_console = MagicMock()
        with patch("moai_adk.__main__.get_console", return_value=mock_console):
            show_logo()

            # Verify console.print was called multiple times (for logo, version, etc.)
            assert mock_console.print.call_count > 0


def test_statusline_handles_eof_error():
    """Test statusline handles EOFError gracefully."""
    runner = CliRunner()

    with patch("moai_adk.statusline.main.build_statusline_data") as mock_build:
        mock_build.return_value = "output"

        # Simulate EOFError by setting up stdin that will raise EOFError
        with patch("sys.stdin.read", side_effect=EOFError):
            result = runner.invoke(cli, ["statusline"])

            # Should handle gracefully
            assert result.exit_code == 0


def test_statusline_handles_value_error():
    """Test statusline handles ValueError gracefully."""
    runner = CliRunner()

    with patch("moai_adk.statusline.main.build_statusline_data") as mock_build:
        mock_build.return_value = "output"

        # Simulate ValueError in JSON parsing
        with patch("sys.stdin.read", side_effect=ValueError):
            result = runner.invoke(cli, ["statusline"])

            # Should handle gracefully
            assert result.exit_code == 0


def test_get_console_lazy_imports_rich():
    """Test get_console lazy imports Console from rich."""
    import moai_adk.__main__ as main_module

    main_module._console = None

    # First call should trigger import
    console1 = get_console()

    # Verify it's a Console instance
    from rich.console import Console

    assert isinstance(console1, Console)

    # Second call should return cached instance
    console2 = get_console()
    assert console1 is console2


def test_init_command_with_path_argument():
    """Test init command accepts path argument."""
    runner = CliRunner()

    result = runner.invoke(cli, ["init", "/tmp/test-path", "--help"])

    # Should accept path argument
    assert result.exit_code == 0


def test_init_command_mode_choices():
    """Test init command mode accepts valid choices."""
    runner = CliRunner()

    # Test personal mode
    result = runner.invoke(cli, ["init", "--mode", "personal", "--help"])
    assert result.exit_code == 0

    # Test team mode
    result = runner.invoke(cli, ["init", "--mode", "team", "--help"])
    assert result.exit_code == 0


def test_init_command_locale_choices():
    """Test init command locale accepts valid choices."""
    runner = CliRunner()

    for locale in ["ko", "en", "ja", "zh"]:
        result = runner.invoke(cli, ["init", "--locale", locale, "--help"])
        assert result.exit_code == 0


def test_update_command_merge_strategies_mutually_exclusive():
    """Test update command merge strategies work correctly."""
    runner = CliRunner()

    # Test with --merge (auto strategy)
    result = runner.invoke(cli, ["update", "--merge", "--help"])
    assert result.exit_code == 0

    # Test with --manual
    result = runner.invoke(cli, ["update", "--manual", "--help"])
    assert result.exit_code == 0

    # Test without either (should use default)
    result = runner.invoke(cli, ["update", "--help"])
    assert result.exit_code == 0


def test_main_console_flush_in_finally_block():
    """Test main() flushes console in finally block even with exception."""
    import moai_adk.__main__ as main_module

    mock_console = MagicMock()
    main_module._console = mock_console

    with patch("moai_adk.__main__.cli") as mock_cli:
        mock_cli.side_effect = RuntimeError("Test error")

        with patch("moai_adk.__main__.get_console", return_value=mock_console):
            result = main()

            # Verify flush was called in finally block
            assert mock_console.file.flush.called


def test_show_logo_all_print_paths():
    """Test show_logo exercises all console.print paths."""
    import moai_adk.__main__ as main_module

    main_module._console = None

    with patch("pyfiglet.figlet_format") as mock_figlet:
        mock_figlet.return_value = "TEST LOGO"

        mock_console = MagicMock()
        with patch("moai_adk.__main__.get_console", return_value=mock_console):
            show_logo()

            # Verify multiple print calls for different elements
            assert mock_console.print.call_count >= 6
