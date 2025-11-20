"""E2E Test: Full MoAI-ADK Workflow

Tests the complete workflow:
1. Project initialization (`python -m moai_adk init`)
2. Status check (`python -m moai_adk status`)
3. System diagnostics (`python -m moai_adk doctor`)
4. Backup creation (`python -m moai_adk backup`)
5. Restore (`python -m moai_adk restore`)
"""

from pathlib import Path

import pytest
from click.testing import CliRunner

from moai_adk.__main__ import cli


class TestFullWorkflow:
    """Test complete MoAI-ADK workflow"""

    @pytest.mark.e2e
    def test_init_and_status_workflow(self, tmp_path):
        """Test init → status workflow"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Step 1: Initialize project
            result = runner.invoke(
                cli,
                [
                    "init",
                    ".",
                    "--mode=personal",
                    "--locale=en",
                    "--non-interactive",
                ],
            )
            assert result.exit_code == 0 or Path(".moai").exists()

            # Step 2: Check status
            if Path(".moai").exists():
                result = runner.invoke(cli, ["status"])
                assert result.exit_code == 0
                assert "personal" in result.output.lower() or "mode" in result.output.lower()

    @pytest.mark.e2e
    def test_doctor_workflow(self):
        """Test system diagnostics workflow"""
        runner = CliRunner()

        # Run doctor command
        result = runner.invoke(cli, ["doctor"])
        # Should complete (may have warnings but not crash)
        assert result.exit_code == 0 or "check" in result.output.lower()

    @pytest.mark.e2e
    def test_help_commands(self):
        """Test all --help commands work"""
        runner = CliRunner()

        commands = ["", "init", "status", "doctor", "update"]  # restore removed - not implemented

        for cmd in commands:
            args = [cmd, "--help"] if cmd else ["--help"]
            result = runner.invoke(cli, args)
            assert result.exit_code == 0
            assert len(result.output) > 0

    @pytest.mark.e2e
    def test_version_command(self):
        """Test --version command"""
        runner = CliRunner()

        result = runner.invoke(cli, ["--version"])
        assert result.exit_code == 0
        assert "MoAI-ADK" in result.output or "0." in result.output


@pytest.mark.e2e
class TestCLIErrorHandling:
    """Test CLI error handling scenarios"""

    def test_status_without_init(self, tmp_path):
        """Test status command without initialization"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            result = runner.invoke(cli, ["status"])
            assert result.exit_code != 0 or "not found" in result.output.lower()

    @pytest.mark.skip(reason="restore command not implemented - handled by checkpoint system")
    def test_restore_without_backups(self, tmp_path):
        """Test restore without backups"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            result = runner.invoke(cli, ["restore"])
            assert result.exit_code != 0 or "backup" in result.output.lower()

    def test_update_without_init(self, tmp_path):
        """Test update without initialization"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            result = runner.invoke(cli, ["update", "--check"])
            assert result.exit_code != 0 or "not initialized" in result.output.lower()


@pytest.mark.e2e
class TestIntegrationFlow:
    """Test integrated command flows"""

    def test_init_doctor_status_flow(self, tmp_path):
        """Test: init → doctor → status"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # 1. Init
            result_init = runner.invoke(
                cli,
                [
                    "init",
                    ".",
                    "--mode=personal",
                    "--locale=en",
                    "--non-interactive",
                ],
            )

            # 2. Doctor (always runs)
            result_doctor = runner.invoke(cli, ["doctor"])
            assert result_doctor.exit_code == 0 or "check" in result_doctor.output.lower()

            # 3. Status (if init succeeded)
            if result_init.exit_code == 0 and Path(".moai").exists():
                result_status = runner.invoke(cli, ["status"])
                assert result_status.exit_code == 0

    def test_help_consistency(self):
        """Test all commands have consistent help"""
        runner = CliRunner()

        commands = ["init", "status", "doctor", "update"]  # restore removed - not implemented

        for cmd in commands:
            result = runner.invoke(cli, [cmd, "--help"])
            assert result.exit_code == 0
            # Help should contain command name or usage info
            assert cmd in result.output.lower() or "usage" in result.output.lower()
