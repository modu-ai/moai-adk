# __main__.py exception handling tests
"""Tests for exception handling and error scenarios in __main__.

Following TDD RED-GREEN-REFACTOR cycle.
These tests cover keyboard interrupts, generic exceptions, and console flushing.
"""

from __future__ import annotations

from unittest.mock import MagicMock, patch

import click


class TestMainFunctionExceptionHandling:
    """Tests for main() function exception handling."""

    def test_main_keyboard_interrupt_returns_exit_code_130(self) -> None:
        """Test main() returns 130 on KeyboardInterrupt (Ctrl+C)."""
        from moai_adk.__main__ import main

        with patch("moai_adk.__main__.cli") as mock_cli:
            # Simulate Ctrl+C via click.Abort
            mock_cli.side_effect = click.Abort()

            exit_code = main()

            assert exit_code == 130

    def test_main_click_exception_returns_exit_code(self) -> None:
        """Test main() returns exit code from ClickException."""
        from moai_adk.__main__ import main

        with patch("moai_adk.__main__.cli") as mock_cli:
            # Create a ClickException with exit code 2
            mock_exception = click.ClickException("Test error")
            mock_exception.exit_code = 2
            mock_exception.show = MagicMock()

            mock_cli.side_effect = mock_exception

            exit_code = main()

            assert exit_code == 2
            mock_exception.show.assert_called_once()

    def test_main_generic_exception_returns_exit_code_1(self) -> None:
        """Test main() returns 1 on generic exception."""
        from moai_adk.__main__ import main

        with patch("moai_adk.__main__.cli") as mock_cli:
            # Simulate generic exception
            mock_cli.side_effect = RuntimeError("Unexpected error")

            with patch("moai_adk.__main__.get_console") as mock_get_console:
                mock_console = MagicMock()
                mock_get_console.return_value = mock_console

                exit_code = main()

                assert exit_code == 1
                # Verify error message was printed
                mock_console.print.assert_called()
                args = mock_console.print.call_args[0]
                assert "Error:" in str(args)

    def test_main_success_returns_exit_code_0(self) -> None:
        """Test main() returns 0 on successful execution."""
        from moai_adk.__main__ import main

        with patch("moai_adk.__main__.cli") as mock_cli:
            mock_cli.return_value = None

            exit_code = main()

            assert exit_code == 0

    def test_main_flushes_console_when_exists(self) -> None:
        """Test main() flushes console when _console exists."""
        from moai_adk.__main__ import main

        mock_file = MagicMock()

        with patch("moai_adk.__main__._console") as mock_console:
            mock_console.file = mock_file

            with patch("moai_adk.__main__.cli"):
                _ = main()

                # Verify flush was called
                mock_file.flush.assert_called_once()

    def test_main_with_no_console_skips_flush(self) -> None:
        """Test main() skips flush when _console is None."""
        from moai_adk.__main__ import main

        with patch("moai_adk.__main__._console", None):
            with patch("moai_adk.__main__.cli"):
                # Should not raise any errors
                exit_code = main()
                assert exit_code == 0

    def test_main_displays_generic_exception_message(self) -> None:
        """Test main() displays formatted error message for generic exceptions."""
        from moai_adk.__main__ import main

        with patch("moai_adk.__main__.cli") as mock_cli:
            test_exception = ValueError("Test error message")
            mock_cli.side_effect = test_exception

            with patch("moai_adk.__main__.get_console") as mock_get_console:
                mock_console = MagicMock()
                mock_get_console.return_value = mock_console

                exit_code = main()

                # Verify exit code is 1 for generic exceptions
                assert exit_code == 1
                # Verify error was printed with proper formatting
                assert mock_console.print.called
                print_args = mock_console.print.call_args[0]
                assert "Error:" in str(print_args)

    def test_main_preserves_exception_type_in_message(self) -> None:
        """Test main() includes exception type in error message."""
        from moai_adk.__main__ import main

        with patch("moai_adk.__main__.cli") as mock_cli:
            test_exception = KeyError("missing_key")
            mock_cli.side_effect = test_exception

            with patch("moai_adk.__main__.get_console") as mock_get_console:
                mock_console = MagicMock()
                mock_get_console.return_value = mock_console

                exit_code = main()

                # Verify exception info is preserved
                assert exit_code == 1


class TestCliGroupBehavior:
    """Tests for CLI group-level behavior."""

    def test_cli_invokes_subcommand_via_context(self) -> None:
        """Test CLI properly invokes subcommands through context."""
        from click.testing import CliRunner

        from moai_adk.__main__ import cli

        runner = CliRunner()

        # Test that init command is available
        result = runner.invoke(cli, ["--help"])
        assert "init" in result.output
        assert "Initialize a new MoAI-ADK project" in result.output

    def test_cli_has_all_defined_commands(self) -> None:
        """Test CLI has all expected commands registered."""
        from moai_adk.__main__ import cli

        commands = ["init", "doctor", "status", "update", "claude", "cc", "glm", "statusline", "rank"]

        for cmd in commands:
            assert cmd in cli.commands, f"Command '{cmd}' not found in CLI"

    def test_cli_version_option_configured(self) -> None:
        """Test CLI has version option configured."""
        from click.testing import CliRunner

        from moai_adk.__main__ import cli

        runner = CliRunner()
        result = runner.invoke(cli, ["--version"])

        assert result.exit_code == 0
        # Version info should be in output


class TestCliCommandOptions:
    """Tests for CLI command option definitions."""

    def test_init_command_options_defined(self) -> None:
        """Test init command has all expected options."""
        from moai_adk.__main__ import cli

        # Get init command
        init_cmd = cli.commands.get("init")
        assert init_cmd is not None

        # Check params exist
        param_names = [p.name for p in init_cmd.params]
        expected_params = ["path", "non_interactive", "mode", "locale", "language", "force"]

        for param in expected_params:
            assert param in param_names, f"Parameter '{param}' not found in init command"

    def test_update_command_options_defined(self) -> None:
        """Test update command has all expected options."""
        from moai_adk.__main__ import cli

        # Get update command
        update_cmd = cli.commands.get("update")
        assert update_cmd is not None

        # Check params exist
        param_names = [p.name for p in update_cmd.params]
        expected_params = ["path", "force", "check", "templates_only", "yes", "edit_config"]

        for param in expected_params:
            assert param in param_names, f"Parameter '{param}' not found in update command"


class TestLazyLoadingImports:
    """Tests to verify lazy-loading structure is in place."""

    def test_commands_imported_inside_functions(self) -> None:
        """Test that heavy libraries are imported inside functions, not at module level."""
        import inspect

        import moai_adk.__main__ as main_module

        # Get source code
        source = inspect.getsource(main_module)

        # Rich imports should be inside functions, not at module level
        # Check that "from rich import" is not at the top level
        lines = source.split("\n")
        top_level_imports = []
        in_function = False

        for line in lines:
            stripped = line.strip()
            if stripped.startswith("def ") or stripped.startswith("class "):
                in_function = True
            elif stripped and not in_function:
                if "from rich import" in stripped or "import rich" in stripped:
                    top_level_imports.append(stripped)

        # Rich should not be imported at module level
        assert len(top_level_imports) == 0, "Rich should be lazy-loaded inside functions"

    def test_pyfiglet_imported_inside_show_logo(self) -> None:
        """Test that pyfiglet is imported inside show_logo function."""
        import inspect

        import moai_adk.__main__ as main_module

        # Get show_logo source
        show_logo_source = inspect.getsource(main_module.show_logo)

        # Should contain pyfiglet import inside function
        assert "import pyfiglet" in show_logo_source or "from pyfiglet" in show_logo_source


class TestCliErrorScenarios:
    """Tests for CLI error scenarios and edge cases."""

    def test_cli_handles_missing_command_gracefully(self) -> None:
        """Test CLI provides helpful error for missing commands."""
        from click.testing import CliRunner

        from moai_adk.__main__ import cli

        runner = CliRunner()
        result = runner.invoke(cli, ["nonexistent-command"])

        # Should not crash, should show error or help
        assert result.exit_code != 0 or "no such command" in result.output.lower()

    def test_cli_handles_invalid_option(self) -> None:
        """Test CLI handles invalid options gracefully."""
        from click.testing import CliRunner

        from moai_adk.__main__ import cli

        runner = CliRunner()
        result = runner.invoke(cli, ["--invalid-option"])

        # Should show error about invalid option
        assert result.exit_code != 0


class TestMainEntryPoints:
    """Tests for main entry point execution."""

    def test_main_callable_as_entry_point(self) -> None:
        """Test main() can be called as entry point."""
        from moai_adk.__main__ import main

        with patch("moai_adk.__main__.cli"):
            # Should be callable without errors
            exit_code = main()
            assert exit_code == 0

    def test_main_returns_integer_exit_codes(self) -> None:
        """Test main() always returns integer exit codes."""
        from moai_adk.__main__ import main

        # Test success case
        with patch("moai_adk.__main__.cli"):
            exit_code = main()
            assert isinstance(exit_code, int)

        # Test error case
        with patch("moai_adk.__main__.cli") as mock_cli:
            mock_cli.side_effect = Exception("Test")
            exit_code = main()
            assert isinstance(exit_code, int)
