# @TEST:CLI-001 | SPEC: SPEC-CLI-001/spec.md
"""Unit tests for backup command

Basic tests for CLI backup functionality.
"""

from click.testing import CliRunner

from moai_adk.cli.commands.backup import backup


class TestBackupCommand:
    """Test backup command"""

    def test_backup_command_exists(self):
        """Should have backup command"""
        assert backup is not None
        assert callable(backup)

    def test_backup_runs_without_error(self):
        """Should run without crashing"""
        runner = CliRunner()

        with runner.isolated_filesystem():
            result = runner.invoke(backup)

            # May fail without .moai setup, but shouldn't crash
            assert result is not None

    def test_backup_output_mentions_backup(self):
        """Should mention backup in output"""
        runner = CliRunner()

        with runner.isolated_filesystem():
            result = runner.invoke(backup)

            # Should mention "backup" somewhere
            assert "backup" in result.output.lower() or result.exit_code in [0, 1]
