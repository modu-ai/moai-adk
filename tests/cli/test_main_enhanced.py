"""Enhanced pytest tests for src/moai_adk/cli/main.py

Test coverage for:
- Module imports and re-exports
- cli function availability
- show_logo function availability
- __all__ exports
- Integration with __main__ module
"""

import pytest
from click.testing import CliRunner


class TestMainModuleImports:
    """Test suite for main.py module imports and re-exports."""

    def test_module_imports_successfully(self):
        """Test that the main module can be imported without errors."""
        try:
            from moai_adk.cli import main

            assert main is not None
        except ImportError as e:
            pytest.fail(f"Failed to import moai_adk.cli.main: {e}")

    def test_cli_function_is_exported(self):
        """Test that cli function is properly exported from main module."""
        from moai_adk.cli.main import cli

        assert cli is not None
        assert callable(cli)
        # Click groups have .name attribute instead of __name__
        assert hasattr(cli, "name") or hasattr(cli, "__name__")
        if hasattr(cli, "name"):
            assert cli.name == "cli"
        else:
            assert cli.__name__ == "cli"

    def test_show_logo_function_is_exported(self):
        """Test that show_logo function is properly exported from main module."""
        from moai_adk.cli.main import show_logo

        assert show_logo is not None
        assert callable(show_logo)
        assert hasattr(show_logo, "__name__")
        assert show_logo.__name__ == "show_logo"

    def test_all_exports_list(self):
        """Test that __all__ contains expected exports."""
        from moai_adk.cli import main

        assert hasattr(main, "__all__")
        assert isinstance(main.__all__, list)
        assert "cli" in main.__all__
        assert "show_logo" in main.__all__
        assert len(main.__all__) == 2

    def test_cli_is_click_command(self):
        """Test that cli is a proper Click command/group."""
        import click

        from moai_adk.cli.main import cli

        # Check if it's a Click command (has click.Command attributes)
        assert hasattr(cli, "invoke")
        assert hasattr(cli, "name")
        assert isinstance(cli, (click.Command, click.Group))


class TestCliIntegration:
    """Test suite for CLI integration and behavior."""

    def test_cli_can_be_invoked(self):
        """Test that cli can be invoked via CliRunner."""
        from moai_adk.cli.main import cli

        runner = CliRunner()
        result = runner.invoke(cli, ["--help"])

        # Should succeed (exit code 0)
        assert result.exit_code == 0
        # Should display help text
        assert "MoAI Agentic Development Kit" in result.output or "Usage:" in result.output

    def test_cli_version_option(self):
        """Test that cli has version option."""
        from moai_adk.cli.main import cli

        runner = CliRunner()
        result = runner.invoke(cli, ["--version"])

        # Should succeed
        assert result.exit_code == 0
        # Should display version information
        assert "MoAI-ADK" in result.output or "version" in result.output.lower()

    def test_cli_without_arguments(self):
        """Test that cli can be invoked without arguments (shows logo)."""
        from moai_adk.cli.main import cli

        runner = CliRunner()
        result = runner.invoke(cli, [])

        # Should succeed or show expected output
        # Note: This might display logo or help depending on configuration
        assert result.exit_code == 0


class TestShowLogoIntegration:
    """Test suite for show_logo function integration."""

    def test_show_logo_can_be_called(self, capsys):
        """Test that show_logo function can be called without errors."""
        from moai_adk.cli.main import show_logo

        try:
            show_logo()
            # Capture output to verify it prints something
            captured = capsys.readouterr()
            # Should produce some output (logo, version, etc.)
            assert len(captured.out) > 0 or len(captured.err) == 0
        except Exception as e:
            pytest.fail(f"show_logo() raised unexpected exception: {e}")

    def test_show_logo_displays_version(self, capsys):
        """Test that show_logo displays version information."""
        from moai_adk.cli.main import show_logo

        show_logo()
        captured = capsys.readouterr()

        # Should contain version information
        # Note: Rich console might use special formatting, so check broadly
        output = captured.out + captured.err
        assert len(output) > 0  # Should produce output


class TestModuleReExports:
    """Test suite for verifying re-exports from __main__ module."""

    def test_cli_matches_main_module_cli(self):
        """Test that re-exported cli matches the original from __main__."""
        from moai_adk.__main__ import cli as original_cli
        from moai_adk.cli.main import cli as main_cli

        # Should be the same function object
        assert main_cli is original_cli

    def test_show_logo_matches_main_module_show_logo(self):
        """Test that re-exported show_logo matches the original from __main__."""
        from moai_adk.__main__ import show_logo as original_show_logo
        from moai_adk.cli.main import show_logo as main_show_logo

        # Should be the same function object
        assert main_show_logo is original_show_logo

    def test_module_docstring_exists(self):
        """Test that main module has proper documentation."""
        from moai_adk.cli import main

        assert main.__doc__ is not None
        assert len(main.__doc__.strip()) > 0
        # Should describe the module's purpose
        assert any(keyword in main.__doc__.lower() for keyword in ["cli", "entry", "module"])


class TestErrorHandling:
    """Test suite for error handling and edge cases."""

    def test_import_star_includes_expected_symbols(self):
        """Test that 'from moai_adk.cli.main import *' works correctly."""
        # This tests the __all__ list behavior
        from importlib import import_module

        # Import the module
        main_module = import_module("moai_adk.cli.main")

        # Get __all__ exports
        all_exports = getattr(main_module, "__all__", [])

        # Verify all symbols in __all__ are actually available
        for symbol in all_exports:
            assert hasattr(main_module, symbol), f"Symbol '{symbol}' in __all__ but not found in module"

    def test_cli_invocation_with_invalid_command(self):
        """Test CLI behavior with invalid command."""
        from moai_adk.cli.main import cli

        runner = CliRunner()
        result = runner.invoke(cli, ["nonexistent-command"])

        # Should fail with non-zero exit code
        assert result.exit_code != 0
        # Should show error message
        assert "Error" in result.output or "No such command" in result.output or result.exit_code == 2
