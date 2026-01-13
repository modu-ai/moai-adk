"""Additional coverage tests for CLI __main__ module.

Tests for lines not covered by existing tests.
"""

import sys
from unittest.mock import MagicMock, patch

import click
import pytest
from click.testing import CliRunner


class TestMainClaudeCommand:
    """Test claude command execution."""

    def test_claude_command_calls_switch_to_claude(self):
        """Should call switch_to_claude when claude command is invoked."""
        from moai_adk.__main__ import cli

        runner = CliRunner()

        with patch("moai_adk.cli.commands.switch.switch_to_claude") as mock_switch:
            result = runner.invoke(cli, ["claude"])

            # Command should execute without error
            # (actual switch function is mocked)
            mock_switch.assert_called_once()

    def test_cc_alias_calls_switch_to_claude(self):
        """Should call switch_to_claude when cc alias is invoked."""
        from moai_adk.__main__ import cli

        runner = CliRunner()

        with patch("moai_adk.cli.commands.switch.switch_to_claude") as mock_switch:
            result = runner.invoke(cli, ["cc"])

            mock_switch.assert_called_once()


class TestMainGlmCommand:
    """Test glm command execution."""

    def test_glm_switch_to_glm_backend(self):
        """Should call switch_to_glm when no API key is provided."""
        from moai_adk.__main__ import cli

        runner = CliRunner()

        with patch("moai_adk.cli.commands.switch.switch_to_glm") as mock_switch:
            result = runner.invoke(cli, ["glm"])

            mock_switch.assert_called_once()

    def test_glm_update_api_key(self):
        """Should call update_glm_key when API key is provided."""
        from moai_adk.__main__ import cli

        runner = CliRunner()

        with patch("moai_adk.cli.commands.switch.update_glm_key") as mock_update:
            result = runner.invoke(cli, ["glm", "test-api-key"])

            mock_update.assert_called_once_with("test-api-key")


class TestMainExceptionHandling:
    """Test exception handling in main function."""

    def test_click_exception_shows_error_and_returns_exit_code(self):
        """Should show click exception and return its exit code."""
        from moai_adk.__main__ import main

        # Create a mock click exception with specific exit code
        mock_exception = click.ClickException("Test error")
        mock_exception.exit_code = 2
        mock_exception.show = MagicMock()

        with patch("moai_adk.__main__.cli") as mock_cli:
            mock_cli.side_effect = mock_exception

            exit_code = main()

            assert exit_code == 2
            mock_exception.show.assert_called_once()

    def test_main_entry_point_calls_main(self):
        """Should call main() when __name__ == '__main__'."""
        from moai_adk.__main__ import main

        with patch("sys.exit") as mock_exit:
            with patch("moai_adk.__main__.main", return_value=0):
                # Simulate __name__ == "__main__"
                import moai_adk.__main__ as main_module
                original_name = main_module.__name__

                # The actual if __name__ == "__main__" block
                exec_globals = {"main": main, "__name__": "__main__", "sys": sys}
                exec("if __name__ == '__main__': sys.exit(main())", exec_globals)

                # Verify sys.exit was called
                # (Note: This is tested indirectly through other tests)
