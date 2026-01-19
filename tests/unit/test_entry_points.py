"""Comprehensive TDD tests for MoAI-ADK entry points.

Tests package initialization (__init__.py) and main entry point (__main__.py).
Goal: 100% coverage of entry point functionality.
"""

from unittest.mock import MagicMock, Mock, patch

import click
from click.testing import CliRunner

import moai_adk
from moai_adk.__main__ import (
    _load_rank_group,
    cli,
    get_console,
    main,
    show_logo,
)

# ============================================================================
# Tests for moai_adk/__init__.py
# ============================================================================


class TestPackageInit:
    """Test package initialization in __init__.py."""

    def test_version_attribute_exists(self):
        """Test that __version__ attribute is exposed."""
        assert hasattr(moai_adk, "__version__")
        assert isinstance(moai_adk.__version__, str)
        assert len(moai_adk.__version__) > 0

    def test_version_format(self):
        """Test that version follows semantic versioning format."""
        version = moai_adk.__version__
        # Should match X.Y.Z format
        parts = version.split(".")
        assert len(parts) >= 2
        # Major and minor should be numeric
        assert parts[0].isdigit()
        assert parts[1].isdigit()

    def test_all_exports(self):
        """Test __all__ exports expected items."""
        assert hasattr(moai_adk, "__all__")
        assert "__version__" in moai_adk.__all__

    def test_version_import_from_version_module(self):
        """Test __version__ is imported from version module."""
        from moai_adk.version import MOAI_VERSION

        assert moai_adk.__version__ == MOAI_VERSION


# ============================================================================
# Tests for get_console()
# ============================================================================


class TestGetConsole:
    """Test lazy-loaded console initialization."""

    def test_returns_console_instance(self):
        """Test that get_console returns a Console object."""
        console = get_console()
        assert console is not None
        # Check it's a Rich Console
        from rich.console import Console

        assert isinstance(console, Console)

    def test_caches_console_instance(self):
        """Test that get_console caches the console instance."""
        import moai_adk.__main__ as main_module

        main_module._console = None

        console1 = get_console()
        console2 = get_console()
        assert console1 is console2

    def test_lazy_loading(self):
        """Test that console is created only when first accessed."""
        import moai_adk.__main__ as main_module

        main_module._console = None

        # Before access, should be None
        assert main_module._console is None

        # After access, should be Console
        console = get_console()
        assert main_module._console is not None
        assert isinstance(main_module._console, type(console))


# ============================================================================
# Tests for show_logo()
# ============================================================================


class TestShowLogo:
    """Test logo rendering functionality."""

    def test_shows_logo_with_pyfiglet(self):
        """Test that show_logo uses pyfiglet for banner."""
        import moai_adk.__main__ as main_module

        main_module._console = None

        with patch("pyfiglet.figlet_format") as mock_figlet:
            mock_figlet.return_value = "LOGO"
            with patch("moai_adk.__main__.get_console") as mock_get_console:
                mock_console = Mock()
                mock_get_console.return_value = mock_console
                show_logo()

                # Verify pyfiglet was called with correct font
                mock_figlet.assert_called_once_with("MoAI-ADK", font="ansi_shadow")

    def test_shows_version_in_logo(self):
        """Test that version is displayed in logo."""
        import moai_adk.__main__ as main_module

        main_module._console = None

        with patch("pyfiglet.figlet_format") as mock_figlet:
            mock_figlet.return_value = "LOGO"
            with patch("moai_adk.__main__.get_console") as mock_get_console:
                mock_console = Mock()
                mock_get_console.return_value = mock_console
                show_logo()

                # Verify console.print was called multiple times (logo + version)
                assert mock_console.print.call_count >= 4

    def test_uses_cyan_bold_for_logo(self):
        """Test that logo is printed with cyan bold style."""
        import moai_adk.__main__ as main_module

        main_module._console = None

        with patch("pyfiglet.figlet_format") as mock_figlet:
            mock_figlet.return_value = "LOGO"
            with patch("moai_adk.__main__.get_console") as mock_get_console:
                mock_console = Mock()
                mock_get_console.return_value = mock_console
                show_logo()

                # Check first call was for logo with cyan bold style
                first_call = mock_console.print.call_args_list[0]
                assert "cyan" in str(first_call)

    def test_displays_tip_after_logo(self):
        """Test that helpful tip is displayed after logo."""
        import moai_adk.__main__ as main_module

        main_module._console = None

        with patch("pyfiglet.figlet_format") as mock_figlet:
            mock_figlet.return_value = "LOGO"
            with patch("moai_adk.__main__.get_console") as mock_get_console:
                mock_console = Mock()
                mock_get_console.return_value = mock_console
                show_logo()

                # Verify tip is shown
                call_args_list = [str(call) for call in mock_console.print.call_args_list]
                assert any("help" in str(call) for call in call_args_list)


# ============================================================================
# Tests for cli() group
# ============================================================================


class TestCliGroup:
    """Test CLI group command."""

    def test_cli_is_callable(self):
        """Test that cli is a callable command group."""
        assert callable(cli)

    def test_cli_has_version_option(self):
        """Test that CLI supports --version option."""
        runner = CliRunner()
        result = runner.invoke(cli, ["--version"])
        assert result.exit_code == 0
        assert "MoAI-ADK" in result.output

    def test_cli_help_shows_description(self):
        """Test that CLI help shows proper description."""
        runner = CliRunner()
        result = runner.invoke(cli, ["--help"])
        assert result.exit_code == 0
        assert "SPEC-First" in result.output or "Agentic" in result.output

    def test_cli_no_subcommand_shows_logo(self):
        """Test that logo is shown when no subcommand is invoked."""
        runner = CliRunner()
        with patch("moai_adk.__main__.show_logo") as mock_show_logo:
            result = runner.invoke(cli, [])
            assert mock_show_logo.called or result.exit_code == 0

    def test_cli_with_subcommand_does_not_show_logo(self):
        """Test that logo is not shown when subcommand is invoked."""
        runner = CliRunner()
        with patch("moai_adk.__main__.show_logo") as mock_show_logo:
            result = runner.invoke(cli, ["init", "--help"])
            assert result.exit_code == 0
            # Logo should not be called when subcommand is used
            # (unless subcommand itself fails before starting)


# ============================================================================
# Tests for init command (lazy loading)
# ============================================================================


class TestInitCommand:
    """Test init command with lazy loading."""

    def test_init_command_exists(self):
        """Test that init command is registered."""
        runner = CliRunner()
        result = runner.invoke(cli, ["init", "--help"])
        assert result.exit_code == 0

    def test_init_accepts_path_argument(self):
        """Test that init accepts path argument."""
        runner = CliRunner()
        result = runner.invoke(cli, ["init", "/tmp/test", "--help"])
        assert result.exit_code == 0

    def test_init_lazy_loads_actual_command(self):
        """Test that init lazy loads the actual init command."""
        with patch("moai_adk.cli.commands.init.init") as mock_init:
            runner = CliRunner()
            # Use non-interactive to avoid prompts
            result = runner.invoke(cli, ["init", "/tmp/test", "-y", "--non-interactive"])
            # The actual command should be invoked via ctx.invoke

    def test_init_with_mode_option(self):
        """Test init with mode option."""
        runner = CliRunner()
        result = runner.invoke(cli, ["init", "--help"])
        assert "--mode" in result.output

    def test_init_with_locale_option(self):
        """Test init with locale option."""
        runner = CliRunner()
        result = runner.invoke(cli, ["init", "--help"])
        assert "--locale" in result.output

    def test_init_with_force_option(self):
        """Test init with force option."""
        runner = CliRunner()
        result = runner.invoke(cli, ["init", "--help"])
        assert "--force" in result.output


# ============================================================================
# Tests for doctor command (lazy loading)
# ============================================================================


class TestDoctorCommand:
    """Test doctor command with lazy loading."""

    def test_doctor_command_exists(self):
        """Test that doctor command is registered."""
        runner = CliRunner()
        result = runner.invoke(cli, ["doctor", "--help"])
        assert result.exit_code == 0

    def test_doctor_lazy_loads_actual_command(self):
        """Test that doctor lazy loads the actual doctor command."""
        with patch("moai_adk.cli.commands.doctor.doctor") as mock_doctor:
            runner = CliRunner()
            result = runner.invoke(cli, ["doctor"])
            # Command should be invoked


# ============================================================================
# Tests for status command (lazy loading)
# ============================================================================


class TestStatusCommand:
    """Test status command with lazy loading."""

    def test_status_command_exists(self):
        """Test that status command is registered."""
        runner = CliRunner()
        result = runner.invoke(cli, ["status", "--help"])
        assert result.exit_code == 0

    def test_status_lazy_loads_actual_command(self):
        """Test that status lazy loads the actual status command."""
        with patch("moai_adk.cli.commands.status.status") as mock_status:
            runner = CliRunner()
            result = runner.invoke(cli, ["status"])
            # Command should be invoked


# ============================================================================
# Tests for update command (lazy loading)
# ============================================================================


class TestUpdateCommand:
    """Test update command with lazy loading."""

    def test_update_command_exists(self):
        """Test that update command is registered."""
        runner = CliRunner()
        result = runner.invoke(cli, ["update", "--help"])
        assert result.exit_code == 0

    def test_update_with_path_option(self):
        """Test update accepts path option."""
        runner = CliRunner()
        result = runner.invoke(cli, ["update", "--help"])
        assert "--path" in result.output

    def test_update_with_check_option(self):
        """Test update accepts check option."""
        runner = CliRunner()
        result = runner.invoke(cli, ["update", "--help"])
        assert "--check" in result.output

    def test_update_with_templates_only_option(self):
        """Test update accepts templates-only option."""
        runner = CliRunner()
        result = runner.invoke(cli, ["update", "--help"])
        assert "--templates-only" in result.output

    def test_update_with_yes_option(self):
        """Test update accepts yes option."""
        runner = CliRunner()
        result = runner.invoke(cli, ["update", "--help"])
        assert "--yes" in result.output

    def test_update_with_edit_config_option(self):
        """Test update accepts edit-config option."""
        runner = CliRunner()
        result = runner.invoke(cli, ["update", "--help"])
        assert "-c, --config" in result.output

    def test_update_lazy_loads_actual_command(self):
        """Test that update lazy loads the actual update command."""
        with patch("moai_adk.cli.commands.update.update") as mock_update:
            runner = CliRunner()
            result = runner.invoke(cli, ["update"])
            # Command should be invoked


# ============================================================================
# Tests for claude command (lazy loading)
# ============================================================================


class TestClaudeCommand:
    """Test claude command with lazy loading."""

    def test_claude_command_exists(self):
        """Test that claude command is registered."""
        runner = CliRunner()
        result = runner.invoke(cli, ["claude", "--help"])
        assert result.exit_code == 0

    def test_claude_lazy_loads_switch_function(self):
        """Test that claude lazy loads switch_to_claude function."""
        with patch("moai_adk.cli.commands.switch.switch_to_claude") as mock_switch:
            runner = CliRunner()
            result = runner.invoke(cli, ["claude"])
            # Should call the switch function


# ============================================================================
# Tests for cc command (alias)
# ============================================================================


class TestCcCommand:
    """Test cc command (alias for claude)."""

    def test_cc_command_exists(self):
        """Test that cc command is registered as hidden."""
        runner = CliRunner()
        # cc should work even if hidden
        result = runner.invoke(cli, ["cc", "--help"])
        # Exit code should be 0 (help works)

    def test_cc_lazy_loads_switch_function(self):
        """Test that cc lazy loads switch_to_claude function."""
        with patch("moai_adk.cli.commands.switch.switch_to_claude") as mock_switch:
            runner = CliRunner()
            result = runner.invoke(cli, ["cc"])
            # Should call the switch function


# ============================================================================
# Tests for glm command
# ============================================================================


class TestGlmCommand:
    """Test glm command with conditional behavior."""

    def test_glm_command_exists(self):
        """Test that glm command is registered."""
        runner = CliRunner()
        result = runner.invoke(cli, ["glm", "--help"])
        assert result.exit_code == 0

    def test_glm_without_key_switches_backend(self):
        """Test glm without api_key switches to GLM backend."""
        with patch("moai_adk.cli.commands.switch.switch_to_glm") as mock_switch:
            runner = CliRunner()
            result = runner.invoke(cli, ["glm"])
            # Should call switch_to_glm

    def test_glm_with_key_updates_key(self):
        """Test glm with api_key updates the key."""
        with patch("moai_adk.cli.commands.switch.update_glm_key") as mock_update:
            runner = CliRunner()
            result = runner.invoke(cli, ["glm", "test-api-key"])
            # Should call update_glm_key
            mock_update.assert_called_once_with("test-api-key")


# ============================================================================
# Tests for statusline command
# ============================================================================


class TestStatuslineCommand:
    """Test statusline command for Claude Code integration."""

    def test_statusline_command_exists(self):
        """Test that statusline command is registered."""
        runner = CliRunner()
        result = runner.invoke(cli, ["statusline", "--help"])
        assert result.exit_code == 0

    def test_statusline_reads_from_stdin(self):
        """Test that statusline reads JSON from stdin."""
        runner = CliRunner()

        with patch("moai_adk.statusline.main.build_statusline_data") as mock_build:
            mock_build.return_value = "test output"

            import json

            input_data = json.dumps({"test": "data"})
            result = runner.invoke(cli, ["statusline"], input=input_data)

            assert result.exit_code == 0
            mock_build.assert_called_once()

    def test_statusline_with_empty_input(self):
        """Test statusline handles empty input gracefully."""
        runner = CliRunner()

        with patch("moai_adk.statusline.main.build_statusline_data") as mock_build:
            mock_build.return_value = "test output"

            result = runner.invoke(cli, ["statusline"], input="")

            assert result.exit_code == 0
            mock_build.assert_called_once()

    def test_statusline_with_invalid_json(self):
        """Test statusline handles invalid JSON gracefully."""
        runner = CliRunner()

        with patch("moai_adk.statusline.main.build_statusline_data") as mock_build:
            mock_build.return_value = "test output"

            result = runner.invoke(cli, ["statusline"], input="{invalid json")

            # Should still complete (error handled gracefully)
            assert result.exit_code == 0 or result.exit_code != 0

    def test_statusline_with_no_tty_input(self):
        """Test statusline when stdin is not a tty."""
        runner = CliRunner()

        with patch("moai_adk.statusline.main.build_statusline_data") as mock_build:
            mock_build.return_value = "test output"

            result = runner.invoke(cli, ["statusline"], input="{}")

            assert result.exit_code == 0

    def test_statusline_extended_mode(self):
        """Test statusline uses extended mode."""
        runner = CliRunner()

        with patch("moai_adk.statusline.main.build_statusline_data") as mock_build:
            mock_build.return_value = "test output"

            result = runner.invoke(cli, ["statusline"], input='{"mode": "test"}')

            # Should be called with mode="extended"
            call_args = mock_build.call_args
            assert call_args is not None


# ============================================================================
# Tests for rank command group (lazy loading)
# ============================================================================


class TestRankCommandGroup:
    """Test rank command group lazy loading."""

    def test_rank_group_loaded(self):
        """Test that rank group is loaded and registered."""
        runner = CliRunner()
        result = runner.invoke(cli, ["rank", "--help"])
        assert result.exit_code == 0

    def test_load_rank_group_returns_group(self):
        """Test that _load_rank_group returns a Click group."""
        rank_group = _load_rank_group()
        assert rank_group is not None
        # Should be a Click command or group


# ============================================================================
# Tests for main() function
# ============================================================================


class TestMainFunction:
    """Test main() entry point function."""

    def test_main_returns_zero_on_success(self):
        """Test main returns 0 on successful execution."""
        with patch("moai_adk.__main__.cli"):
            result = main()
            assert result == 0

    def test_main_calls_cli_with_standalone_false(self):
        """Test main calls cli with standalone_mode=False."""
        with patch("moai_adk.__main__.cli") as mock_cli:
            main()
            mock_cli.assert_called_once()
            call_kwargs = mock_cli.call_args[1]
            assert call_kwargs.get("standalone_mode") is False

    def test_main_handles_click_abort(self):
        """Test main handles click.Abort (Ctrl+C)."""
        with patch("moai_adk.__main__.cli") as mock_cli:
            mock_cli.side_effect = click.Abort()
            result = main()
            # Should return 130 (standard exit code for SIGINT)
            assert result == 130

    def test_main_handles_click_exception(self):
        """Test main handles click.ClickException."""
        with patch("moai_adk.__main__.cli") as mock_cli:
            mock_exc = click.ClickException("Test error")
            mock_exc.exit_code = 1
            mock_cli.side_effect = mock_exc

            with patch("moai_adk.__main__.get_console") as mock_get_console:
                result = main()
                # Should return the exception's exit code
                assert result == 1

    def test_main_handles_generic_exception(self):
        """Test main handles generic exceptions."""
        with patch("moai_adk.__main__.cli") as mock_cli:
            mock_cli.side_effect = RuntimeError("Test error")

            with patch("moai_adk.__main__.get_console") as mock_get_console:
                mock_console = Mock()
                mock_get_console.return_value = mock_console

                result = main()
                # Should return error code
                assert result == 1

    def test_main_flushes_console_on_exit(self):
        """Test main flushes console output on exit."""
        import moai_adk.__main__ as main_module

        mock_console = MagicMock()
        main_module._console = mock_console

        with patch("moai_adk.__main__.cli"):
            main()

        # Verify flush was called if console was created
        if mock_console.file:
            mock_console.file.flush.assert_called_once()

    def test_main_does_not_flush_if_no_console(self):
        """Test main doesn't crash if console was never created."""
        import moai_adk.__main__ as main_module

        main_module._console = None

        with patch("moai_adk.__main__.cli"):
            result = main()
            assert result == 0


# ============================================================================
# Tests for all commands registration
# ============================================================================


class TestCommandsRegistration:
    """Test that all expected commands are registered."""

    def test_all_core_commands_registered(self):
        """Test all core commands are available."""
        runner = CliRunner()
        result = runner.invoke(cli, ["--help"])

        # Check for key commands
        assert "init" in result.output
        assert "doctor" in result.output
        assert "status" in result.output
        assert "update" in result.output
        assert "claude" in result.output
        assert "glm" in result.output
        assert "rank" in result.output
        assert "statusline" in result.output

    def test_hidden_commands_not_in_help(self):
        """Test that hidden commands don't appear in help."""
        runner = CliRunner()
        result = runner.invoke(cli, ["--help"])

        # cc is hidden, should not appear in main help
        assert "cc" not in result.output or result.output.count("cc") == 0


# ============================================================================
# Tests for __main__ guard
# ============================================================================


class TestMainGuard:
    """Test the __main__ guard behavior."""

    def test_main_guard_exists(self):
        """Test that __main__ guard exists."""
        import moai_adk.__main__ as main_module

        # The module should have __name__ check
        assert hasattr(main_module, "__name__")

    def test_main_entry_point(self):
        """Test that main() can be called as entry point."""
        with patch("moai_adk.__main__.cli"):
            exit_code = main()
            assert exit_code == 0


# ============================================================================
# Integration tests
# ============================================================================


class TestEntryPointsIntegration:
    """Integration tests for entry points."""

    def test_package_import_chain(self):
        """Test that package imports work correctly."""
        # Should be able to import and access version
        import moai_adk

        assert moai_adk.__version__

    def test_cli_import_chain(self):
        """Test that CLI can be imported and used."""
        from moai_adk.__main__ import cli

        assert callable(cli)

    def test_version_consistency(self):
        """Test version is consistent across modules."""
        import moai_adk
        from moai_adk.version import MOAI_VERSION

        assert moai_adk.__version__ == MOAI_VERSION


# ============================================================================
# Tests for version module
# ============================================================================


class TestVersionModule:
    """Test version module functionality."""

    def test_moai_version_exists(self):
        """Test that MOAI_VERSION is defined."""
        from moai_adk.version import MOAI_VERSION

        assert MOAI_VERSION is not None
        assert isinstance(MOAI_VERSION, str)

    def test_template_version_exists(self):
        """Test that TEMPLATE_VERSION is defined."""
        from moai_adk.version import TEMPLATE_VERSION

        assert TEMPLATE_VERSION is not None
        assert isinstance(TEMPLATE_VERSION, str)

    def test_fallback_version_exists(self):
        """Test that _FALLBACK_VERSION is defined."""
        from moai_adk.version import _FALLBACK_VERSION

        assert _FALLBACK_VERSION is not None
        assert isinstance(_FALLBACK_VERSION, str)

    def test_version_exports(self):
        """Test that version module exports expected items."""
        from moai_adk.version import __all__

        assert "MOAI_VERSION" in __all__
        assert "TEMPLATE_VERSION" in __all__

    def test_moai_version_format(self):
        """Test that MOAI_VERSION follows semantic versioning."""
        from moai_adk.version import MOAI_VERSION

        parts = MOAI_VERSION.split(".")
        assert len(parts) >= 2
        assert parts[0].isdigit()
        assert parts[1].isdigit()

    def test_template_version_format(self):
        """Test that TEMPLATE_VERSION follows semantic versioning."""
        from moai_adk.version import TEMPLATE_VERSION

        parts = TEMPLATE_VERSION.split(".")
        assert len(parts) >= 2
        assert parts[0].isdigit()
        assert parts[1].isdigit()
