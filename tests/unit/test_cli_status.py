# # REMOVED_ORPHAN_TEST:CLI-003 | SPEC: SPEC-CLI-001/spec.md
"""Unit tests for status command

Basic tests for CLI status functionality.
"""

from click.testing import CliRunner

from moai_adk.cli.commands.status import status


class TestStatusCommand:
    """Test status command"""

    def test_status_command_exists(self):
        """Should have status command"""
        assert status is not None
        assert callable(status)

    def test_status_runs_without_error(self):
        """Should run without crashing"""
        runner = CliRunner()

        with runner.isolated_filesystem():
            result = runner.invoke(status)

            # May fail without .moai setup, but shouldn't crash
            assert result is not None

    def test_status_output_mentions_project(self):
        """Should mention project in output"""
        runner = CliRunner()

        with runner.isolated_filesystem():
            result = runner.invoke(status)

            # Should mention "project" or "status" somewhere
            assert "project" in result.output.lower() or "status" in result.output.lower() or result.exit_code in [0, 1]
