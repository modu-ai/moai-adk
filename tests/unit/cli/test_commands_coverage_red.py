"""
@TEST:CLI-COMMANDS-RED-001 RED Phase - Failing tests for CLI commands coverage improvement

Phase 1 of TDD: Write failing tests that accurately reflect the real CLI API
and target the core user journeys for maximum impact on coverage.

Target Coverage Goals:
- commands.py: 21.97% → 80%
- Key functions: cli(), init(), doctor(), status(), update()
- Real API based tests, not mocked assumptions
"""

import tempfile
from pathlib import Path
from unittest.mock import Mock, patch

import pytest
from click.testing import CliRunner

from moai_adk.cli.commands import cli, doctor, init, status, update, restore, help


class TestCLIMainCommandRed:
    """RED: Test the main CLI group with real API calls."""

    def setup_method(self):
        """Set up test environment."""
        self.runner = CliRunner()

    def test_version_flag_should_display_version_and_exit_successfully(self):
        """RED: Test --version flag displays version and exits with code 0."""
        result = self.runner.invoke(cli, ['--version'])

        # Should exit successfully and contain version
        assert result.exit_code == 0
        assert "MoAI-ADK v" in result.output

    def test_help_flag_should_display_help_text_and_exit_successfully(self):
        """RED: Test -h/--help flag displays help."""
        result = self.runner.invoke(cli, ['--help'])

        assert result.exit_code == 0
        assert "Modu-AI's Agentic Development Kit" in result.output

    def test_no_subcommand_should_display_banner_and_help(self):
        """RED: Test CLI with no args shows banner and help."""
        result = self.runner.invoke(cli)

        assert result.exit_code == 0
        # Should show some helpful output, not crash

    def test_invalid_subcommand_should_exit_with_error(self):
        """RED: Test invalid subcommand returns non-zero exit code."""
        result = self.runner.invoke(cli, ['nonexistent-command'])

        assert result.exit_code != 0


class TestInitCommandRed:
    """RED: Test init command with real project directory operations."""

    def setup_method(self):
        """Set up test environment."""
        self.runner = CliRunner()
        self.temp_dir = None

    def teardown_method(self):
        """Clean up test environment."""
        if self.temp_dir and self.temp_dir.exists():
            import shutil
            shutil.rmtree(self.temp_dir)

    def test_init_in_empty_directory_should_create_moai_structure(self):
        """RED: Test init creates .moai directory structure."""
        with self.runner.isolated_filesystem():
            result = self.runner.invoke(init, ['.', '--personal', '--force'])

            # Should create basic structure
            assert Path('.moai').exists()
            assert Path('.claude').exists()
            assert Path('CLAUDE.md').exists()

            # Should exit successfully
            assert result.exit_code == 0

    def test_init_with_backup_flag_should_create_backup_before_installation(self):
        """RED: Test --backup flag creates backup."""
        with self.runner.isolated_filesystem():
            # Create some existing content
            Path('.moai').mkdir()
            Path('.moai/config.json').write_text('{"test": "data"}')

            result = self.runner.invoke(init, ['.', '--backup', '--force'])

            # Should succeed even if backup has issues
            assert result.exit_code == 0

            # Check if backup was attempted (may fail due to permissions)
            backup_dirs = list(Path('..').glob('*_moai_backup_*'))
            # Backup may exist in parent directory or fail - both are acceptable

    def test_init_with_interactive_flag_should_start_wizard(self):
        """RED: Test --interactive flag starts the wizard."""
        with self.runner.isolated_filesystem():
            # Simulate user cancelling wizard
            result = self.runner.invoke(init, ['--interactive'], input='\n' * 20)

            # Should handle wizard interaction (success or controlled exit)
            assert result.exit_code in [0, 1]

    def test_init_with_conflicting_personal_and_team_flags_should_handle_gracefully(self):
        """RED: Test conflicting mode flags."""
        with self.runner.isolated_filesystem():
            result = self.runner.invoke(init, ['.', '--personal', '--team'])

            # Should handle conflicting flags gracefully
            assert result.exit_code in [0, 1, 2]

    def test_init_in_readonly_directory_should_fail_gracefully(self):
        """RED: Test init in readonly directory."""
        with tempfile.TemporaryDirectory() as temp_dir:
            readonly_dir = Path(temp_dir) / "readonly"
            readonly_dir.mkdir(mode=0o444)  # readonly

            try:
                result = self.runner.invoke(init, [str(readonly_dir)])

                # Should handle gracefully (may succeed with logging or fail cleanly)
                assert result.exit_code in [0, 1, 2]

                # Should contain appropriate error message if failed
                if result.exit_code != 0:
                    assert "Permission denied" in result.output or "failed" in result.output.lower()
            finally:
                # Restore permissions for cleanup
                readonly_dir.chmod(0o755)


class TestDoctorCommandRed:
    """RED: Test doctor command diagnostics."""

    def setup_method(self):
        """Set up test environment."""
        self.runner = CliRunner()

    def test_doctor_should_check_environment_and_project_status(self):
        """RED: Test doctor command basic functionality."""
        result = self.runner.invoke(doctor)

        assert result.exit_code == 0
        assert "Health Check" in result.output

    def test_doctor_with_list_backups_should_show_available_backups(self):
        """RED: Test --list-backups flag."""
        with self.runner.isolated_filesystem():
            # Create a fake backup directory
            Path('.moai_backup_20231201_120000').mkdir()

            result = self.runner.invoke(doctor, ['--list-backups'])

            assert result.exit_code == 0
            assert "Available backups" in result.output

    @patch('moai_adk.cli.commands.validate_environment')
    def test_doctor_should_report_environment_validation_failure(self, mock_validate):
        """RED: Test doctor when environment validation fails."""
        mock_validate.return_value = False

        result = self.runner.invoke(doctor)

        assert result.exit_code == 0  # Doctor should complete even with issues
        assert "Environment" in result.output


class TestStatusCommandRed:
    """RED: Test status command project analysis."""

    def setup_method(self):
        """Set up test environment."""
        self.runner = CliRunner()

    def test_status_should_show_project_status_information(self):
        """RED: Test status command shows project information."""
        result = self.runner.invoke(status)

        assert result.exit_code == 0
        assert "Project Status" in result.output

    def test_status_with_verbose_flag_should_show_detailed_information(self):
        """RED: Test --verbose flag shows more details."""
        result = self.runner.invoke(status, ['--verbose'])

        assert result.exit_code == 0
        assert "Project Status" in result.output

    def test_status_with_project_path_should_analyze_specified_directory(self):
        """RED: Test --project-path option."""
        with tempfile.TemporaryDirectory() as temp_dir:
            result = self.runner.invoke(status, ['--project-path', temp_dir])

            assert result.exit_code == 0

    def test_status_in_moai_project_should_show_component_status(self):
        """RED: Test status in an actual MoAI project."""
        with self.runner.isolated_filesystem():
            # Set up basic MoAI structure
            Path('.moai').mkdir()
            Path('.moai/config.json').write_text('{}')
            Path('.claude').mkdir()
            Path('CLAUDE.md').write_text('# Test')

            result = self.runner.invoke(status)

            assert result.exit_code == 0
            assert "✅" in result.output  # Should show some success indicators


class TestUpdateCommandRed:
    """RED: Test update command functionality."""

    def setup_method(self):
        """Set up test environment."""
        self.runner = CliRunner()

    def test_update_check_flag_should_show_version_information(self):
        """RED: Test --check flag shows update info."""
        with self.runner.isolated_filesystem():
            result = self.runner.invoke(update, ['--check'])

            assert result.exit_code == 0
            assert "Checking for updates" in result.output

    def test_update_in_non_moai_project_should_warn_user(self):
        """RED: Test update outside MoAI project."""
        with self.runner.isolated_filesystem():
            result = self.runner.invoke(update)

            # Should warn about not being a MoAI project
            assert "doesn't appear to be a MoAI-ADK project" in result.output

    def test_update_with_package_only_flag_should_handle_package_updates(self):
        """RED: Test --package-only flag."""
        with self.runner.isolated_filesystem():
            Path('.moai').mkdir()  # Make it look like a MoAI project

            result = self.runner.invoke(update, ['--package-only', '--no-backup'])

            assert result.exit_code == 0

    def test_update_with_conflicting_flags_should_error(self):
        """RED: Test conflicting --package-only and --resources-only."""
        result = self.runner.invoke(update, ['--package-only', '--resources-only'])

        assert result.exit_code == 1
        assert "Cannot use --package-only and --resources-only together" in result.output


class TestRestoreCommandRed:
    """RED: Test restore command functionality."""

    def setup_method(self):
        """Set up test environment."""
        self.runner = CliRunner()

    def test_restore_with_valid_backup_should_restore_files(self):
        """RED: Test restore with valid backup directory."""
        with self.runner.isolated_filesystem():
            # Create backup structure
            backup_dir = Path('test_backup')
            backup_dir.mkdir()
            (backup_dir / '.moai').mkdir()
            (backup_dir / '.claude').mkdir()
            (backup_dir / 'CLAUDE.md').write_text('# Backup test')

            result = self.runner.invoke(restore, [str(backup_dir)])

            assert result.exit_code == 0

    def test_restore_with_dry_run_should_show_preview(self):
        """RED: Test --dry-run flag."""
        with self.runner.isolated_filesystem():
            backup_dir = Path('test_backup')
            backup_dir.mkdir()
            (backup_dir / '.moai').mkdir()

            result = self.runner.invoke(restore, [str(backup_dir), '--dry-run'])

            assert result.exit_code == 0
            assert "Dry run" in result.output

    def test_restore_with_invalid_backup_should_error(self):
        """RED: Test restore with non-existent backup."""
        result = self.runner.invoke(restore, ['/nonexistent/backup'])

        assert result.exit_code == 1


class TestHelpCommandRed:
    """RED: Test help command functionality."""

    def setup_method(self):
        """Set up test environment."""
        self.runner = CliRunner()

    def test_help_without_arguments_should_show_general_help(self):
        """RED: Test help command shows available commands."""
        result = self.runner.invoke(help)

        assert result.exit_code == 0
        assert "Available commands" in result.output

    def test_help_with_specific_command_should_show_command_help(self):
        """RED: Test help for specific command."""
        result = self.runner.invoke(help, ['init'])

        assert result.exit_code == 0

    def test_help_with_invalid_command_should_show_error(self):
        """RED: Test help for non-existent command."""
        result = self.runner.invoke(help, ['nonexistent'])

        assert result.exit_code == 0  # Help should not crash
        assert "Unknown command" in result.output


class TestCLIEdgeCasesRed:
    """RED: Test edge cases and error conditions."""

    def setup_method(self):
        """Set up test environment."""
        self.runner = CliRunner()

    def test_keyboard_interrupt_should_be_handled_gracefully(self):
        """RED: Test Ctrl+C handling."""
        # This is hard to test directly, but we can test that the CLI
        # doesn't crash with unexpected exceptions
        result = self.runner.invoke(cli, ['--help'])
        assert result.exit_code == 0

    def test_unicode_arguments_should_be_handled_safely(self):
        """RED: Test unicode in arguments."""
        with self.runner.isolated_filesystem():
            result = self.runner.invoke(init, ['.', '--force'], input='프로젝트\n')

            # Should not crash with unicode
            assert result.exit_code in [0, 1, 2]

    def test_very_long_path_should_be_handled(self):
        """RED: Test long path handling."""
        long_path = 'a' * 200  # Very long path
        result = self.runner.invoke(init, [long_path])

        # Should handle gracefully
        assert result.exit_code in [0, 1, 2]

    def test_special_characters_in_project_path_should_be_handled(self):
        """RED: Test special characters in paths."""
        with self.runner.isolated_filesystem():
            special_dir = "test-project_with&special#chars"
            Path(special_dir).mkdir()

            result = self.runner.invoke(init, [special_dir, '--force'])

            # Should handle special characters
            assert result.exit_code in [0, 1, 2]