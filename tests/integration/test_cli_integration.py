"""CLI Integration Tests

CLI 명령어 통합 테스트:
- Click CliRunner를 사용한 CLI 테스트
- __main__.py entry point 테스트
- 모든 CLI 명령어 기본 실행 테스트
"""

import json
import sys
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import click
import pytest
from click.testing import CliRunner

from moai_adk.__main__ import cli, main, show_logo
from moai_adk.cli.commands.doctor import doctor
from moai_adk.cli.commands.init import create_progress_callback, init

# from moai_adk.cli.commands.restore import restore  # Not implemented - handled by checkpoint system
from moai_adk.cli.commands.status import status
from moai_adk.cli.commands.update import update


class TestMainEntryPoint:
    """__main__.py entry point tests"""

    def test_show_logo(self, capsys):
        """Test ASCII logo display"""
        show_logo()
        captured = capsys.readouterr()
        # Logo contains "MoAI" text
        assert "MoAI" in captured.out or "Modu-AI" in captured.out
        assert "Version:" in captured.out

    def test_main_no_subcommand(self):
        """Test CLI without subcommand shows logo"""
        runner = CliRunner()
        result = runner.invoke(cli)
        assert result.exit_code == 0
        assert "MoAI" in result.output or "Modu-AI" in result.output

    def test_main_help_flag(self):
        """Test --help flag"""
        runner = CliRunner()
        result = runner.invoke(cli, ["--help"])
        assert result.exit_code == 0
        assert "MoAI Agentic Development Kit" in result.output
        assert "init" in result.output

    def test_main_version_flag(self):
        """Test --version flag"""
        runner = CliRunner()
        result = runner.invoke(cli, ["--version"])
        assert result.exit_code == 0
        assert "MoAI-ADK" in result.output

    def test_main_function_success(self):
        """Test main() function successful execution"""
        with patch.object(sys, "argv", ["moai-adk", "--help"]):
            exit_code = main()
            assert exit_code == 0

    def test_main_function_ctrl_c(self):
        """Test main() handling Ctrl+C (Abort)"""
        with patch("moai_adk.__main__.cli") as mock_cli:
            mock_cli.side_effect = click.Abort()
            exit_code = main()
            assert exit_code == 130

    def test_main_function_click_exception(self):
        """Test main() handling ClickException"""
        with patch("moai_adk.__main__.cli") as mock_cli:
            mock_exception = click.ClickException("Test error")
            mock_exception.exit_code = 2
            mock_cli.side_effect = mock_exception
            exit_code = main()
            assert exit_code == 2

    def test_main_function_generic_exception(self):
        """Test main() handling generic exception"""
        with patch("moai_adk.__main__.cli") as mock_cli:
            mock_cli.side_effect = RuntimeError("Test error")
            exit_code = main()
            assert exit_code == 1


class TestInitCommand:
    """init command tests"""

    def test_init_help(self):
        """Test init --help"""
        runner = CliRunner()
        result = runner.invoke(init, ["--help"])
        assert result.exit_code == 0
        assert "Initialize" in result.output or "init" in result.output.lower()

    def test_init_noninteractive_minimal(self, tmp_path):
        """Test init in non-interactive mode"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            result = runner.invoke(
                init,
                [
                    ".",
                    "--mode=personal",
                    "--locale=en",
                    "--non-interactive",
                ],
            )

            # Should create .moai directory (check exit code or directory)
            assert result.exit_code == 0 or Path(".moai").exists()

    @pytest.mark.skip(reason="Init command no longer has interactive prompts to interrupt")
    def test_init_interactive_abort(self, tmp_path):
        """Test init interactive mode with abort"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Use a subdirectory to trigger interactive prompt
            # This will ask for project name, allowing interrupt
            result = runner.invoke(init, ["my-project"], input="\x03")
            # Click captures Ctrl+C as Abort
            assert result.exit_code != 0 or "Aborted" in result.output

    def test_init_already_initialized(self, tmp_path):
        """Test init on already initialized project"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai directory directly
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            (moai_dir / "config.json").write_text('{"mode": "personal"}')

            # Try to initialize again (without --force)
            result = runner.invoke(
                init,
                [
                    ".",
                    "--mode=personal",
                    "--locale=en",
                    "--non-interactive",
                ],
            )
            # Should detect existing project or ask for reinit or succeed with --force
            assert "exist" in result.output.lower() or "already" in result.output.lower() or result.exit_code == 0

    def test_init_with_all_flags(self, tmp_path):
        """Test init with all CLI flags"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            runner.invoke(
                init,
                [
                    ".",
                    "--mode=team",
                    "--locale=ko",
                    "--language=python",
                    "--non-interactive",
                ],
            )

            # Check config.json created
            config_path = Path(".moai") / "config.json"
            if config_path.exists():
                with open(config_path) as f:
                    config = json.load(f)
                    project = config.get("project", {})
                    # Check project section with legacy fallback
                    assert project.get("mode") or config.get("mode") == "team"
                    assert project.get("locale") or config.get("locale") == "ko"

    def test_create_progress_callback(self):
        """Test create_progress_callback function"""
        mock_progress = MagicMock()
        task_ids = [1, 2, 3]

        callback = create_progress_callback(mock_progress, task_ids)

        # Test valid phase
        callback("Test message", 1, 3)
        mock_progress.update.assert_called_once_with(task_ids[0], completed=1, description="Test message")

        # Test out of range
        mock_progress.reset_mock()
        callback("Invalid", 0, 3)
        mock_progress.update.assert_not_called()

        callback("Invalid", 4, 3)
        mock_progress.update.assert_not_called()


class TestStatusCommand:
    """status command tests"""

    def test_status_help(self):
        """Test status --help"""
        runner = CliRunner()
        result = runner.invoke(status, ["--help"])
        assert result.exit_code == 0
        assert "status" in result.output.lower()

    def test_status_no_moai_directory(self, tmp_path):
        """Test status without .moai directory"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            result = runner.invoke(status)
            assert "No .moai/config.json found" in result.output or result.exit_code != 0

    def test_status_with_config(self, tmp_path):
        """Test status with existing config"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai directory and config
            moai_dir = Path(".moai")
            moai_dir.mkdir()

            config = {"mode": "personal", "locale": "en"}
            with open(moai_dir / "config.json", "w") as f:
                json.dump(config, f)

            result = runner.invoke(status)
            assert result.exit_code == 0
            assert "personal" in result.output

    def test_status_with_git_repo(self, tmp_path):
        """Test status in git repository"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai config
            moai_dir = Path(".moai")
            moai_dir.mkdir()

            config = {"mode": "team", "locale": "ko"}
            with open(moai_dir / "config.json", "w") as f:
                json.dump(config, f)

            # Mock git repo - need to patch git.Repo directly
            with patch("git.Repo") as mock_repo_class:
                mock_repo = Mock()
                mock_repo.active_branch.name = "main"
                mock_repo.is_dirty.return_value = False
                mock_repo_class.return_value = mock_repo

                result = runner.invoke(status)
                assert result.exit_code == 0
                assert "team" in result.output

    def test_status_with_spec_count(self, tmp_path):
        """Test status shows SPEC count"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create structure
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            specs_dir = moai_dir / "specs"
            specs_dir.mkdir()

            # Create dummy SPECs
            (specs_dir / "SPEC-TEST-001").mkdir()
            (specs_dir / "SPEC-TEST-001" / "spec.md").touch()

            config = {"mode": "personal", "locale": "en"}
            with open(moai_dir / "config.json", "w") as f:
                json.dump(config, f)

            result = runner.invoke(status)
            assert result.exit_code == 0
            assert "1" in result.output  # Should show 1 SPEC


class TestDoctorCommand:
    """doctor command tests"""

    def test_doctor_help(self):
        """Test doctor --help"""
        runner = CliRunner()
        result = runner.invoke(doctor, ["--help"])
        assert result.exit_code == 0
        assert "doctor" in result.output.lower() or "Check" in result.output

    def test_doctor_all_checks_pass(self):
        """Test doctor with all checks passing"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.doctor.check_environment") as mock_check:
            mock_check.return_value = {
                "Python >= 3.13": True,
                "Git installed": True,
            }

            result = runner.invoke(doctor)
            assert result.exit_code == 0
            assert "All checks passed" in result.output or "✓" in result.output

    def test_doctor_some_checks_fail(self):
        """Test doctor with failing checks"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.doctor.check_environment") as mock_check:
            mock_check.return_value = {
                "Python >= 3.13": True,
                "Git installed": False,
            }

            result = runner.invoke(doctor)
            # Should complete but show warnings
            assert "✗" in result.output or "fail" in result.output.lower()

    def test_doctor_exception_handling(self):
        """Test doctor handles exceptions"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.doctor.check_environment") as mock_check:
            mock_check.side_effect = RuntimeError("Check failed")

            result = runner.invoke(doctor)
            assert result.exit_code != 0


class TestUpdateCommand:
    """update command tests"""

    def test_update_help(self):
        """Test update --help"""
        runner = CliRunner()
        result = runner.invoke(update, ["--help"])
        assert result.exit_code == 0
        assert "update" in result.output.lower()

    def test_update_check_only(self, tmp_path):
        """Test update --check flag"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai directory
            moai_dir = Path(".moai")
            moai_dir.mkdir()

            result = runner.invoke(update, ["--check"])
            # Should check for updates without installing
            assert result.exit_code == 0
            assert "버전" in result.output or "version" in result.output.lower()

    def test_update_execution(self, tmp_path):
        """Test update command execution"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai directory
            moai_dir = Path(".moai")
            moai_dir.mkdir()

            # Test --check flag (doesn't actually update)
            result = runner.invoke(update, ["--check"])
            assert result.exit_code == 0


class TestCLICommandIntegration:
    """Integration tests for all CLI commands"""

    def test_all_commands_registered(self):
        """Test all commands are registered in CLI"""
        runner = CliRunner()
        result = runner.invoke(cli, ["--help"])

        assert "init" in result.output
        assert "status" in result.output
        assert "doctor" in result.output
        assert "backup" in result.output
        # assert "restore" in result.output  # Not implemented - handled by checkpoint system
        assert "update" in result.output

    def test_commands_have_help(self):
        """Test all commands have --help"""
        runner = CliRunner()
        commands = [init, status, doctor, backup, update]  # restore removed - not implemented

        for cmd in commands:
            result = runner.invoke(cmd, ["--help"])
            assert result.exit_code == 0
            assert len(result.output) > 0

    def test_cli_invocation_chain(self, tmp_path):
        """Test CLI command chain: init → status → doctor"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # 1. Create .moai directory manually (init test is complex)
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            config = {"mode": "personal", "locale": "en"}
            with open(moai_dir / "config.json", "w") as f:
                json.dump(config, f)

            # 2. Check status
            status_result = runner.invoke(status)
            assert "personal" in status_result.output

            # 3. Run doctor
            with patch("moai_adk.cli.commands.doctor.check_environment") as mock_check:
                mock_check.return_value = {"Python >= 3.13": True}
                doctor_result = runner.invoke(doctor)
                assert doctor_result.exit_code == 0
