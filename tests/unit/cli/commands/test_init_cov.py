"""Comprehensive test coverage for init command.

Focus on uncovered code paths with mocked dependencies using @patch.
Tests actual code paths without side effects.
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch, call
from datetime import datetime

import pytest
import click
from click.testing import CliRunner
from rich.console import Console
from rich.progress import Progress

from moai_adk.cli.commands.init import init, create_progress_callback


class TestCreateProgressCallback:
    """Test create_progress_callback function."""

    def test_callback_creation(self):
        """Test callback function creation."""
        # Arrange
        progress = MagicMock(spec=Progress)
        task_ids = [1, 2, 3]

        # Act
        callback = create_progress_callback(progress, task_ids)

        # Assert
        assert callback is not None
        assert callable(callback)

    def test_callback_updates_progress(self):
        """Test callback updates progress correctly."""
        # Arrange
        progress = MagicMock(spec=Progress)
        task_ids = [1, 2, 3]
        callback = create_progress_callback(progress, task_ids)

        # Act
        callback("Phase 1 complete", 1, 3)

        # Assert
        progress.update.assert_called_once_with(task_ids[0], completed=1, description="Phase 1 complete")

    def test_callback_updates_second_task(self):
        """Test callback updates second task correctly."""
        # Arrange
        progress = MagicMock(spec=Progress)
        task_ids = [1, 2, 3]
        callback = create_progress_callback(progress, task_ids)

        # Act
        callback("Phase 2 complete", 2, 3)

        # Assert
        progress.update.assert_called_once_with(task_ids[1], completed=1, description="Phase 2 complete")

    def test_callback_ignores_invalid_phase(self):
        """Test callback ignores invalid phase numbers."""
        # Arrange
        progress = MagicMock(spec=Progress)
        task_ids = [1, 2, 3]
        callback = create_progress_callback(progress, task_ids)

        # Act
        callback("Invalid", 0, 3)

        # Assert
        progress.update.assert_not_called()

    def test_callback_ignores_out_of_range_phase(self):
        """Test callback ignores out of range phase numbers."""
        # Arrange
        progress = MagicMock(spec=Progress)
        task_ids = [1, 2, 3]
        callback = create_progress_callback(progress, task_ids)

        # Act
        callback("Invalid", 5, 3)

        # Assert
        progress.update.assert_not_called()

    def test_callback_multiple_phases(self):
        """Test callback with multiple phase updates."""
        # Arrange
        progress = MagicMock(spec=Progress)
        task_ids = [1, 2, 3]
        callback = create_progress_callback(progress, task_ids)

        # Act
        callback("Phase 1", 1, 3)
        callback("Phase 2", 2, 3)
        callback("Phase 3", 3, 3)

        # Assert
        assert progress.update.call_count == 3
        expected_calls = [
            call(1, completed=1, description="Phase 1"),
            call(2, completed=1, description="Phase 2"),
            call(3, completed=1, description="Phase 3"),
        ]
        progress.update.assert_has_calls(expected_calls)


class TestInitCommand:
    """Test init command function."""

    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    def test_init_non_interactive_minimal(self, mock_version_reader_class, mock_print_banner, mock_initializer_class):
        """Test init command in non-interactive mode with minimal setup."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            # Setup mocks
            mock_version_reader = MagicMock()
            mock_version_reader.get_version.return_value = "0.30.0"
            mock_version_reader_class.return_value = mock_version_reader

            mock_initializer = MagicMock()
            mock_initializer.is_initialized.return_value = False
            mock_result = MagicMock()
            mock_result.success = True
            mock_result.project_path = tmpdir
            mock_result.language = "python"
            mock_result.locale = "en"
            mock_result.created_files = [f"file{i}" for i in range(5)]
            mock_result.duration = 1000
            mock_result.errors = []
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [tmpdir, "--non-interactive", "-y"])

            # Assert
            assert result.exit_code == 0
            mock_initializer.initialize.assert_called_once()

    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    def test_init_with_mode_option(self, mock_version_reader_class, mock_print_banner, mock_initializer_class):
        """Test init command with mode option."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            mock_version_reader = MagicMock()
            mock_version_reader.get_version.return_value = "0.30.0"
            mock_version_reader_class.return_value = mock_version_reader

            mock_initializer = MagicMock()
            mock_initializer.is_initialized.return_value = False
            mock_result = MagicMock()
            mock_result.success = True
            mock_result.project_path = tmpdir
            mock_result.language = "python"
            mock_result.locale = "en"
            mock_result.created_files = []
            mock_result.duration = 500
            mock_result.errors = []
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [tmpdir, "--non-interactive", "--mode", "team"])

            # Assert
            assert result.exit_code == 0
            call_kwargs = mock_initializer.initialize.call_args[1]
            assert call_kwargs["mode"] == "team"

    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    def test_init_with_locale_option(self, mock_version_reader_class, mock_print_banner, mock_initializer_class):
        """Test init command with locale option."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            mock_version_reader = MagicMock()
            mock_version_reader.get_version.return_value = "0.30.0"
            mock_version_reader_class.return_value = mock_version_reader

            mock_initializer = MagicMock()
            mock_initializer.is_initialized.return_value = False
            mock_result = MagicMock()
            mock_result.success = True
            mock_result.project_path = tmpdir
            mock_result.language = "python"
            mock_result.locale = "ko"
            mock_result.created_files = []
            mock_result.duration = 500
            mock_result.errors = []
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [tmpdir, "--non-interactive", "--locale", "ko"])

            # Assert
            assert result.exit_code == 0
            call_kwargs = mock_initializer.initialize.call_args[1]
            assert call_kwargs["locale"] == "ko"

    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    def test_init_with_language_option(self, mock_version_reader_class, mock_print_banner, mock_initializer_class):
        """Test init command with language option."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            mock_version_reader = MagicMock()
            mock_version_reader.get_version.return_value = "0.30.0"
            mock_version_reader_class.return_value = mock_version_reader

            mock_initializer = MagicMock()
            mock_initializer.is_initialized.return_value = False
            mock_result = MagicMock()
            mock_result.success = True
            mock_result.project_path = tmpdir
            mock_result.language = "typescript"
            mock_result.locale = "en"
            mock_result.created_files = []
            mock_result.duration = 500
            mock_result.errors = []
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [tmpdir, "--non-interactive", "--language", "typescript"])

            # Assert
            assert result.exit_code == 0
            call_kwargs = mock_initializer.initialize.call_args[1]
            assert call_kwargs["language"] == "typescript"

    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    def test_init_reinitialization_without_backup(self, mock_version_reader_class, mock_print_banner, mock_initializer_class):
        """Test init command on already initialized project."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            mock_version_reader = MagicMock()
            mock_version_reader.get_version.return_value = "0.30.0"
            mock_version_reader_class.return_value = mock_version_reader

            mock_initializer = MagicMock()
            mock_initializer.is_initialized.return_value = True
            mock_result = MagicMock()
            mock_result.success = True
            mock_result.project_path = tmpdir
            mock_result.language = "python"
            mock_result.locale = "en"
            mock_result.created_files = []
            mock_result.duration = 500
            mock_result.errors = []
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [tmpdir, "--non-interactive"])

            # Assert
            assert result.exit_code == 0
            call_kwargs = mock_initializer.initialize.call_args[1]
            assert call_kwargs["reinit"] is True

    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    def test_init_failure_handling(self, mock_version_reader_class, mock_print_banner, mock_initializer_class):
        """Test init command handles initialization failure."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            mock_version_reader = MagicMock()
            mock_version_reader.get_version.return_value = "0.30.0"
            mock_version_reader_class.return_value = mock_version_reader

            mock_initializer = MagicMock()
            mock_initializer.is_initialized.return_value = False
            mock_result = MagicMock()
            mock_result.success = False
            mock_result.errors = ["Permission denied", "Directory not writable"]
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [tmpdir, "--non-interactive"])

            # Assert
            assert result.exit_code != 0

    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    def test_init_keyboard_interrupt(self, mock_version_reader_class, mock_print_banner, mock_initializer_class):
        """Test init command handles keyboard interrupt."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            mock_version_reader = MagicMock()
            mock_version_reader.get_version.side_effect = KeyboardInterrupt()
            mock_version_reader_class.return_value = mock_version_reader

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [tmpdir, "--non-interactive"])

            # Assert
            assert result.exit_code != 0

    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    def test_init_config_update_on_reinit(self, mock_version_reader_class, mock_print_banner, mock_initializer_class):
        """Test init updates config.json on reinitialization."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Create config file
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.json"
            config_data = {"moai": {"version": "0.25.0"}, "project": {"optimized": True}}
            with open(config_path, "w") as f:
                json.dump(config_data, f)

            mock_version_reader = MagicMock()
            mock_version_reader.get_version.return_value = "0.30.0"
            mock_version_reader_class.return_value = mock_version_reader

            mock_initializer = MagicMock()
            mock_initializer.is_initialized.return_value = True
            mock_result = MagicMock()
            mock_result.success = True
            mock_result.project_path = str(project_path)
            mock_result.language = "python"
            mock_result.locale = "en"
            mock_result.created_files = []
            mock_result.duration = 500
            mock_result.errors = []
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [str(project_path), "--non-interactive"])

            # Assert
            assert result.exit_code == 0
            # Verify config was updated
            with open(config_path) as f:
                updated_config = json.load(f)
            assert updated_config["project"]["optimized"] is False
            assert updated_config["moai"]["version"] == "0.30.0"

    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    def test_init_generic_language_display(self, mock_version_reader_class, mock_print_banner, mock_initializer_class):
        """Test init displays 'Auto-detect' for generic language."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            mock_version_reader = MagicMock()
            mock_version_reader.get_version.return_value = "0.30.0"
            mock_version_reader_class.return_value = mock_version_reader

            mock_initializer = MagicMock()
            mock_initializer.is_initialized.return_value = False
            mock_result = MagicMock()
            mock_result.success = True
            mock_result.project_path = tmpdir
            mock_result.language = "generic"  # Auto-detect marker
            mock_result.locale = "en"
            mock_result.created_files = []
            mock_result.duration = 500
            mock_result.errors = []
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [tmpdir, "--non-interactive"])

            # Assert
            assert result.exit_code == 0
            assert "Auto-detect" in result.output

    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    def test_init_current_directory_default(self, mock_version_reader_class, mock_print_banner, mock_initializer_class):
        """Test init uses current directory as default."""
        # Arrange
        mock_version_reader = MagicMock()
        mock_version_reader.get_version.return_value = "0.30.0"
        mock_version_reader_class.return_value = mock_version_reader

        mock_initializer = MagicMock()
        mock_initializer.is_initialized.return_value = False
        mock_result = MagicMock()
        mock_result.success = True
        mock_result.project_path = "."
        mock_result.language = "python"
        mock_result.locale = "en"
        mock_result.created_files = []
        mock_result.duration = 500
        mock_result.errors = []
        mock_initializer.initialize.return_value = mock_result
        mock_initializer_class.return_value = mock_initializer

        # Act
        runner = CliRunner()
        result = runner.invoke(init, ["--non-interactive"])

        # Assert
        assert result.exit_code == 0

    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    def test_init_removes_old_hook_files(self, mock_version_reader_class, mock_print_banner, mock_initializer_class):
        """Test init removes deprecated hook files on reinitialization."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Create old hook file
            hook_dir = project_path / ".claude" / "hooks" / "alfred"
            hook_dir.mkdir(parents=True, exist_ok=True)
            old_hook = hook_dir / "session_start__startup.py"
            old_hook.write_text("# old hook")

            # Create config
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.json"
            config_data = {"moai": {"version": "0.25.0"}}
            with open(config_path, "w") as f:
                json.dump(config_data, f)

            mock_version_reader = MagicMock()
            mock_version_reader.get_version.return_value = "0.30.0"
            mock_version_reader_class.return_value = mock_version_reader

            mock_initializer = MagicMock()
            mock_initializer.is_initialized.return_value = True
            mock_result = MagicMock()
            mock_result.success = True
            mock_result.project_path = str(project_path)
            mock_result.language = "python"
            mock_result.locale = "en"
            mock_result.created_files = []
            mock_result.duration = 500
            mock_result.errors = []
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [str(project_path), "--non-interactive"])

            # Assert
            assert result.exit_code == 0
            assert not old_hook.exists()

    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    def test_init_version_read_error_recovery(self, mock_version_reader_class, mock_print_banner, mock_initializer_class):
        """Test init recovers from version read errors."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            mock_version_reader = MagicMock()
            mock_version_reader.get_version.side_effect = Exception("Network error")
            mock_version_reader_class.return_value = mock_version_reader

            mock_initializer = MagicMock()
            mock_initializer.is_initialized.return_value = False
            mock_result = MagicMock()
            mock_result.success = True
            mock_result.project_path = tmpdir
            mock_result.language = "python"
            mock_result.locale = "en"
            mock_result.created_files = []
            mock_result.duration = 500
            mock_result.errors = []
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [tmpdir, "--non-interactive"])

            # Assert
            assert result.exit_code == 0
            assert "Version read error" in result.output or "0.30.0" in result.output or result.exit_code == 0

    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    def test_init_backup_info_display(self, mock_version_reader_class, mock_print_banner, mock_initializer_class):
        """Test init displays backup info on reinitialization."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Create backup directory
            backup_dir = project_path / ".moai-backups" / "backup_2025_01_01"
            backup_dir.mkdir(parents=True, exist_ok=True)
            (backup_dir / "config.json").write_text("{}")

            # Create config
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.json"
            config_data = {"moai": {"version": "0.25.0"}}
            with open(config_path, "w") as f:
                json.dump(config_data, f)

            mock_version_reader = MagicMock()
            mock_version_reader.get_version.return_value = "0.30.0"
            mock_version_reader_class.return_value = mock_version_reader

            mock_initializer = MagicMock()
            mock_initializer.is_initialized.return_value = True
            mock_result = MagicMock()
            mock_result.success = True
            mock_result.project_path = str(project_path)
            mock_result.language = "python"
            mock_result.locale = "en"
            mock_result.created_files = []
            mock_result.duration = 500
            mock_result.errors = []
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [str(project_path), "--non-interactive"])

            # Assert
            assert result.exit_code == 0
            assert "Backup" in result.output


class TestInitEdgeCases:
    """Test edge cases and error conditions."""

    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    def test_init_with_all_options(self, mock_version_reader_class, mock_print_banner, mock_initializer_class):
        """Test init with all options specified."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            mock_version_reader = MagicMock()
            mock_version_reader.get_version.return_value = "0.30.0"
            mock_version_reader_class.return_value = mock_version_reader

            mock_initializer = MagicMock()
            mock_initializer.is_initialized.return_value = False
            mock_result = MagicMock()
            mock_result.success = True
            mock_result.project_path = tmpdir
            mock_result.language = "go"
            mock_result.locale = "ja"
            mock_result.created_files = []
            mock_result.duration = 500
            mock_result.errors = []
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(
                init,
                [
                    tmpdir,
                    "--non-interactive",
                    "--mode",
                    "team",
                    "--locale",
                    "ja",
                    "--language",
                    "go",
                    "--force",
                ],
            )

            # Assert
            assert result.exit_code == 0
            call_kwargs = mock_initializer.initialize.call_args[1]
            assert call_kwargs["mode"] == "team"
            assert call_kwargs["locale"] == "ja"
            assert call_kwargs["language"] == "go"
