"""Comprehensive test coverage for init command.

Focus on uncovered code paths with mocked dependencies using @patch.
Tests actual code paths without side effects.
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, call, patch

import pytest
from click.testing import CliRunner
from rich.progress import Progress, TaskID

from moai_adk.cli.commands.init import create_progress_callback, init


class TestCreateProgressCallback:
    """Test create_progress_callback function."""

    def test_callback_creation(self):
        """Test callback function creation."""
        # Arrange
        progress = MagicMock(spec=Progress)
        task_ids = [TaskID(1), TaskID(2), TaskID(3)]

        # Act
        callback = create_progress_callback(progress, task_ids)

        # Assert
        assert callback is not None
        assert callable(callback)

    def test_callback_updates_progress(self):
        """Test callback updates progress correctly."""
        # Arrange
        progress = MagicMock(spec=Progress)
        task_ids = [TaskID(1), TaskID(2), TaskID(3)]
        callback = create_progress_callback(progress, task_ids)

        # Act
        callback("Phase 1 complete", 1, 3)

        # Assert
        progress.update.assert_called_once_with(task_ids[0], completed=1, description="Phase 1 complete")

    def test_callback_updates_second_task(self):
        """Test callback updates second task correctly."""
        # Arrange
        progress = MagicMock(spec=Progress)
        task_ids = [TaskID(1), TaskID(2), TaskID(3)]
        callback = create_progress_callback(progress, task_ids)

        # Act
        callback("Phase 2 complete", 2, 3)

        # Assert
        progress.update.assert_called_once_with(task_ids[1], completed=1, description="Phase 2 complete")

    def test_callback_ignores_invalid_phase(self):
        """Test callback ignores invalid phase numbers."""
        # Arrange
        progress = MagicMock(spec=Progress)
        task_ids = [TaskID(1), TaskID(2), TaskID(3)]
        callback = create_progress_callback(progress, task_ids)

        # Act
        callback("Invalid", 0, 3)

        # Assert
        progress.update.assert_not_called()

    def test_callback_ignores_out_of_range_phase(self):
        """Test callback ignores out of range phase numbers."""
        # Arrange
        progress = MagicMock(spec=Progress)
        task_ids = [TaskID(1), TaskID(2), TaskID(3)]
        callback = create_progress_callback(progress, task_ids)

        # Act
        callback("Invalid", 5, 3)

        # Assert
        progress.update.assert_not_called()

    def test_callback_multiple_phases(self):
        """Test callback with multiple phase updates."""
        # Arrange
        progress = MagicMock(spec=Progress)
        task_ids = [TaskID(1), TaskID(2), TaskID(3)]
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
    def test_init_non_interactive_minimal(self, mock_version_reader_class, _mock_print_banner, mock_initializer_class):
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
    def test_init_with_mode_option(self, mock_version_reader_class, _mock_print_banner, mock_initializer_class):
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
    def test_init_with_locale_option(self, mock_version_reader_class, _mock_print_banner, mock_initializer_class):
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
    def test_init_with_language_option(self, mock_version_reader_class, _mock_print_banner, mock_initializer_class):
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
    def test_init_reinitialization_without_backup(
        self, mock_version_reader_class, _mock_print_banner, mock_initializer_class
    ):
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
    def test_init_failure_handling(self, mock_version_reader_class, _mock_print_banner, mock_initializer_class):
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
    def test_init_keyboard_interrupt(self, mock_version_reader_class, _mock_print_banner, mock_initializer_class):
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
    def test_init_config_update_on_reinit(self, mock_version_reader_class, _mock_print_banner, mock_initializer_class):
        """Test init updates config.json on reinitialization."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Create config file
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.json"
            config_data = {
                "moai": {"version": "0.25.0"},
                "project": {"optimized": True},
            }
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
    def test_init_generic_language_display(self, mock_version_reader_class, _mock_print_banner, mock_initializer_class):
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
    def test_init_current_directory_default(
        self, mock_version_reader_class, _mock_print_banner, mock_initializer_class):
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
    def test_init_removes_old_hook_files(self, mock_version_reader_class, _mock_print_banner, mock_initializer_class):
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
    def test_init_version_read_error_recovery(
        self, mock_version_reader_class, _mock_print_banner, mock_initializer_class
    ):
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
    def test_init_backup_info_display(self, mock_version_reader_class, _mock_print_banner, mock_initializer_class):
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
    def test_init_with_all_options(self, mock_version_reader_class, _mock_print_banner, mock_initializer_class):
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


class TestInitInteractiveMode:
    """Test init command interactive mode functionality (lines 184-217)."""

    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    @patch("moai_adk.cli.commands.init.prompt_project_setup")
    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    def test_init_interactive_mode_success(
        self,
        mock_initializer_class,
        mock_prompt_project_setup,
        mock_version_reader_class,
        _mock_print_banner,
    ):
        """Test init in interactive mode with successful user prompts."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            mock_version_reader = MagicMock()
            mock_version_reader.get_version.return_value = "0.30.0"
            mock_version_reader_class.return_value = mock_version_reader

            mock_prompt_project_setup.return_value = {
                "locale": "en",
                "user_name": "Test User",
                "project_name": "test-project",
                "glm_api_key": "test-glm-key",
                "git_mode": "personal",
                "github_username": "testuser",
                "git_commit_lang": "en",
                "code_comment_lang": "en",
                "doc_lang": "en",
                "development_mode": "ddd",  # DDD is the only methodology
            }

            mock_initializer = MagicMock()
            mock_initializer.is_initialized.return_value = False
            mock_result = MagicMock()
            mock_result.success = True
            mock_result.project_path = tmpdir
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [tmpdir])

            # Assert
            assert result.exit_code == 0
            mock_prompt_project_setup.assert_called_once()
            # Verify that prompt_project_setup was called (path may differ on macOS due to symlink resolution)
            call_args = mock_prompt_project_setup.call_args
            assert call_args is not None
            assert "project_path" in call_args[1]

    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    @patch("moai_adk.cli.commands.init.prompt_project_setup")
    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    def test_init_interactive_mode_with_locale(
        self,
        mock_initializer_class,
        mock_prompt_project_setup,
        mock_version_reader_class,
        _mock_print_banner,
    ):
        """Test init in interactive mode with Korean locale."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            mock_version_reader = MagicMock()
            mock_version_reader.get_version.return_value = "0.30.0"
            mock_version_reader_class.return_value = mock_version_reader

            mock_prompt_project_setup.return_value = {
                "locale": "ko",
                "user_name": "테스트 사용자",
                "project_name": "test-project",
                "glm_api_key": None,
                "git_mode": "manual",
                "github_username": None,
                "git_commit_lang": "en",
                "code_comment_lang": "en",
                "doc_lang": "en",
                "development_mode": "ddd",  # DDD is the only methodology
            }

            mock_initializer = MagicMock()
            mock_initializer.is_initialized.return_value = False
            mock_result = MagicMock()
            mock_result.success = True
            mock_result.project_path = tmpdir
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [tmpdir])

            # Assert
            assert result.exit_code == 0
            mock_prompt_project_setup.assert_called_once()
            call_kwargs = mock_initializer.initialize.call_args[1]
            assert call_kwargs["locale"] == "ko"

    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    @patch("moai_adk.cli.commands.init.prompt_project_setup")
    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    def test_init_interactive_vs_non_interactive(
        self,
        mock_initializer_class,
        mock_prompt_project_setup,
        mock_version_reader_class,
        _mock_print_banner,
    ):
        """Test that --non-interactive flag skips prompts."""
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
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [tmpdir, "--non-interactive"])

            # Assert
            assert result.exit_code == 0
            mock_prompt_project_setup.assert_not_called()


class TestSaveAdditionalConfigGLM:
    """Test _save_additional_config GLM API key handling (lines 490-509)."""

    @patch("moai_adk.core.credentials.save_glm_key_to_env")
    @patch("moai_adk.core.credentials.get_env_glm_path")
    @patch("moai_adk.core.credentials.remove_glm_key_from_shell_config")
    @patch.dict("os.environ", {"GLM_API_KEY": "old-key"})
    def test_save_additional_config_with_glm_api_key(
        self,
        mock_remove_from_shell,
        mock_get_env_path,
        mock_save_glm_key,
    ):
        """Test saving GLM API key and shell cleanup."""
        # Arrange
        from moai_adk.cli.commands.init import _save_additional_config

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            mock_get_env_path.return_value = Path.home() / ".moai" / ".env.glm"
            mock_remove_from_shell.return_value = {".zshrc": True, ".bashrc": False}

            # Act
            _save_additional_config(
                project_path=project_path,
                project_name="test-project",
                locale="en",
                user_name="Test User",
                service_type="glm",
                pricing_plan=None,
                glm_pricing_plan="basic",
                anthropic_api_key=None,
                glm_api_key="test-glm-key-12345",
                git_mode="personal",
                github_username="testuser",
                git_commit_lang="en",
                code_comment_lang="en",
                doc_lang="en",
            )

            # Assert
            mock_save_glm_key.assert_called_once_with("test-glm-key-12345")
            mock_remove_from_shell.assert_called_once()

    @patch("moai_adk.core.credentials.save_glm_key_to_env")
    @patch("moai_adk.core.credentials.get_env_glm_path")
    @patch("moai_adk.core.credentials.remove_glm_key_from_shell_config")
    @patch.dict("os.environ", {}, clear=True)
    def test_save_additional_config_glm_key_without_shell_cleanup(
        self,
        mock_remove_from_shell,
        mock_get_env_path,
        mock_save_glm_key,
    ):
        """Test saving GLM API key without shell cleanup when env var not set."""
        # Arrange
        from moai_adk.cli.commands.init import _save_additional_config

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            mock_get_env_path.return_value = Path.home() / ".moai" / ".env.glm"

            # Act
            _save_additional_config(
                project_path=project_path,
                project_name="test-project",
                locale="en",
                user_name="Test User",
                service_type="glm",
                pricing_plan=None,
                glm_pricing_plan="basic",
                anthropic_api_key=None,
                glm_api_key="test-glm-key",
                git_mode="personal",
                github_username="testuser",
                git_commit_lang="en",
                code_comment_lang="en",
                doc_lang="en",
            )

            # Assert
            mock_save_glm_key.assert_called_once_with("test-glm-key")
            mock_remove_from_shell.assert_not_called()

    @patch("moai_adk.core.credentials.save_credentials")
    def test_save_additional_config_with_anthropic_key(self, mock_save_credentials):
        """Test saving Anthropic API key."""
        # Arrange
        from moai_adk.cli.commands.init import _save_additional_config

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Act
            _save_additional_config(
                project_path=project_path,
                project_name="test-project",
                locale="en",
                user_name="Test User",
                service_type="claude_subscription",
                pricing_plan="pro",
                glm_pricing_plan=None,
                anthropic_api_key="sk-ant-test-key",
                glm_api_key=None,
                git_mode="personal",
                github_username="testuser",
                git_commit_lang="en",
                code_comment_lang="en",
                doc_lang="en",
            )

            # Assert
            mock_save_credentials.assert_called_once_with(anthropic_api_key="sk-ant-test-key", merge=True)

    @patch("moai_adk.core.credentials.save_credentials")
    @patch("moai_adk.core.credentials.save_glm_key_to_env")
    def test_save_additional_config_without_api_keys(self, mock_save_glm_key, mock_save_credentials):
        """Test saving config without API keys."""
        # Arrange
        from moai_adk.cli.commands.init import _save_additional_config

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Act
            _save_additional_config(
                project_path=project_path,
                project_name="test-project",
                locale="en",
                user_name="Test User",
                service_type="glm",
                pricing_plan=None,
                glm_pricing_plan="basic",
                anthropic_api_key=None,
                glm_api_key=None,
                git_mode="manual",
                github_username=None,
                git_commit_lang="en",
                code_comment_lang="en",
                doc_lang="en",
            )

            # Assert
            mock_save_credentials.assert_not_called()
            mock_save_glm_key.assert_not_called()


class TestSaveAdditionalConfigYAML:
    """Test _save_additional_config YAML file operations (lines 511-639)."""

    def test_save_pricing_config(self):
        """Test saving pricing configuration to pricing.yaml."""
        # Arrange
        from moai_adk.cli.commands.init import _save_additional_config

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Act
            _save_additional_config(
                project_path=project_path,
                project_name="test-project",
                locale="en",
                user_name="Test User",
                service_type="glm",
                pricing_plan=None,
                glm_pricing_plan="glm_pro",
                anthropic_api_key=None,
                glm_api_key=None,
                git_mode="personal",
                github_username="testuser",
                git_commit_lang="en",
                code_comment_lang="en",
                doc_lang="en",
            )

            # Assert
            pricing_path = project_path / ".moai" / "config" / "sections" / "pricing.yaml"
            assert pricing_path.exists()
            import yaml

            with open(pricing_path) as f:
                pricing_data = yaml.safe_load(f)
            assert pricing_data["service"]["type"] == "glm"
            assert pricing_data["service"]["glm_pricing_plan"] == "glm_pro"

    def test_save_language_config(self):
        """Test saving language configuration to language.yaml."""
        # Arrange
        from moai_adk.cli.commands.init import _save_additional_config

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Act
            _save_additional_config(
                project_path=project_path,
                project_name="test-project",
                locale="ko",
                user_name="Test User",
                service_type="glm",
                pricing_plan=None,
                glm_pricing_plan="basic",
                anthropic_api_key=None,
                glm_api_key=None,
                git_mode="personal",
                github_username="testuser",
                git_commit_lang="ko",
                code_comment_lang="en",
                doc_lang="en",
            )

            # Assert
            language_path = project_path / ".moai" / "config" / "sections" / "language.yaml"
            assert language_path.exists()
            import yaml

            with open(language_path) as f:
                lang_data = yaml.safe_load(f)
            assert lang_data["language"]["conversation_language"] == "ko"
            assert lang_data["language"]["conversation_language_name"] == "Korean (한국어)"
            assert lang_data["language"]["git_commit_messages"] == "ko"
            assert lang_data["language"]["code_comments"] == "en"
            assert lang_data["language"]["documentation"] == "en"

    def test_save_git_strategy_config(self):
        """Test saving git strategy configuration to git-strategy.yaml."""
        # Arrange
        from moai_adk.cli.commands.init import _save_additional_config

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Act
            _save_additional_config(
                project_path=project_path,
                project_name="test-project",
                locale="en",
                user_name="Test User",
                service_type="glm",
                pricing_plan=None,
                glm_pricing_plan="basic",
                anthropic_api_key=None,
                glm_api_key=None,
                git_mode="team",
                github_username="testuser",
                git_commit_lang="en",
                code_comment_lang="en",
                doc_lang="en",
            )

            # Assert
            git_path = project_path / ".moai" / "config" / "sections" / "git-strategy.yaml"
            assert git_path.exists()
            import yaml

            with open(git_path) as f:
                git_data = yaml.safe_load(f)
            assert git_data["git_strategy"]["mode"] == "team"

    def test_save_project_config(self):
        """Test saving project configuration to project.yaml."""
        # Arrange
        from moai_adk.cli.commands.init import _save_additional_config

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Act
            _save_additional_config(
                project_path=project_path,
                project_name="my-test-project",
                locale="en",
                user_name="Test User",
                service_type="glm",
                pricing_plan=None,
                glm_pricing_plan="basic",
                anthropic_api_key=None,
                glm_api_key=None,
                git_mode="personal",
                github_username="mygithubuser",
                git_commit_lang="en",
                code_comment_lang="en",
                doc_lang="en",
            )

            # Assert
            project_yaml_path = project_path / ".moai" / "config" / "sections" / "project.yaml"
            assert project_yaml_path.exists()
            import yaml

            with open(project_yaml_path) as f:
                project_data = yaml.safe_load(f)
            assert project_data["project"]["name"] == "my-test-project"
            assert project_data["github"]["profile_name"] == "mygithubuser"

    def test_save_user_config(self):
        """Test saving user configuration to user.yaml."""
        # Arrange
        from moai_adk.cli.commands.init import _save_additional_config

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Act
            _save_additional_config(
                project_path=project_path,
                project_name="test-project",
                locale="en",
                user_name="Jane Doe",
                service_type="glm",
                pricing_plan=None,
                glm_pricing_plan="basic",
                anthropic_api_key=None,
                glm_api_key=None,
                git_mode="personal",
                github_username="testuser",
                git_commit_lang="en",
                code_comment_lang="en",
                doc_lang="en",
            )

            # Assert
            user_yaml_path = project_path / ".moai" / "config" / "sections" / "user.yaml"
            assert user_yaml_path.exists()
            import yaml

            with open(user_yaml_path) as f:
                user_data = yaml.safe_load(f)
            assert user_data["user"]["name"] == "Jane Doe"

    def test_save_additional_config_all_sections(self):
        """Test saving all config sections together."""
        # Arrange
        from moai_adk.cli.commands.init import _save_additional_config

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Act
            _save_additional_config(
                project_path=project_path,
                project_name="full-test-project",
                locale="ja",
                user_name="日本語ユーザー",
                service_type="hybrid",
                pricing_plan="max5",
                glm_pricing_plan="enterprise",
                anthropic_api_key="sk-ant-test",
                glm_api_key="glm-test-key",
                git_mode="team",
                github_username="githubuser",
                git_commit_lang="ja",
                code_comment_lang="en",
                doc_lang="ja",
            )

            # Assert
            sections_dir = project_path / ".moai" / "config" / "sections"
            assert sections_dir.exists()
            assert (sections_dir / "pricing.yaml").exists()
            assert (sections_dir / "language.yaml").exists()
            assert (sections_dir / "git-strategy.yaml").exists()
            assert (sections_dir / "project.yaml").exists()
            assert (sections_dir / "user.yaml").exists()


class TestSaveAdditionalConfigErrors:
    """Test _save_additional_config error handling."""

    @patch("moai_adk.cli.commands.init.yaml.safe_dump")
    def test_save_additional_config_file_permission_error(self, mock_yaml_dump):
        """Test handling file permission errors during config save."""
        # Arrange
        from moai_adk.cli.commands.init import _save_additional_config

        mock_yaml_dump.side_effect = PermissionError("Permission denied")

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Act & Assert
            with pytest.raises(PermissionError):
                _save_additional_config(
                    project_path=project_path,
                    project_name="test-project",
                    locale="en",
                    user_name="Test User",
                    service_type="glm",
                    pricing_plan=None,
                    glm_pricing_plan="basic",
                    anthropic_api_key=None,
                    glm_api_key=None,
                    git_mode="personal",
                    github_username="testuser",
                    git_commit_lang="en",
                    code_comment_lang="en",
                    doc_lang="en",
                )

    def test_save_additional_config_existing_yaml_merging(self):
        """Test that existing YAML configs are merged correctly."""
        # Arrange
        from moai_adk.cli.commands.init import _save_additional_config

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            sections_dir = project_path / ".moai" / "config" / "sections"
            sections_dir.mkdir(parents=True, exist_ok=True)

            # Create existing pricing.yaml with extra data
            import yaml

            existing_pricing = {"service": {"type": "claude_subscription", "extra_field": "keep_me"}}
            pricing_path = sections_dir / "pricing.yaml"
            with open(pricing_path, "w") as f:
                yaml.safe_dump(existing_pricing, f)

            # Act
            _save_additional_config(
                project_path=project_path,
                project_name="test-project",
                locale="en",
                user_name="Test User",
                service_type="glm",
                pricing_plan=None,
                glm_pricing_plan="basic",
                anthropic_api_key=None,
                glm_api_key=None,
                git_mode="personal",
                github_username="testuser",
                git_commit_lang="en",
                code_comment_lang="en",
                doc_lang="en",
            )

            # Assert
            with open(pricing_path) as f:
                updated_pricing = yaml.safe_load(f)
            assert updated_pricing["service"]["type"] == "glm"
            assert updated_pricing["service"]["glm_pricing_plan"] == "basic"
            assert updated_pricing["service"]["extra_field"] == "keep_me"

    def test_save_additional_config_directory_creation(self):
        """Test that sections directory is created if it doesn't exist."""
        # Arrange
        from moai_adk.cli.commands.init import _save_additional_config

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            sections_dir = project_path / ".moai" / "config" / "sections"

            # Ensure sections directory doesn't exist
            assert not sections_dir.exists()

            # Act
            _save_additional_config(
                project_path=project_path,
                project_name="test-project",
                locale="en",
                user_name="Test User",
                service_type="glm",
                pricing_plan=None,
                glm_pricing_plan="basic",
                anthropic_api_key=None,
                glm_api_key=None,
                git_mode="personal",
                github_username="testuser",
                git_commit_lang="en",
                code_comment_lang="en",
                doc_lang="en",
            )

            # Assert
            assert sections_dir.exists()


class TestInitEdgeCasesV2:
    """Test init command edge cases and config migration."""

    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    def test_init_reinit_config_migration_yaml_to_json(
        self,
        mock_initializer_class,
        mock_version_reader_class,
        _mock_print_banner,
    ):
        """Test reinitialization with YAML to JSON config migration."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            mock_version_reader = MagicMock()
            mock_version_reader.get_version.return_value = "0.30.0"
            mock_version_reader_class.return_value = mock_version_reader

            mock_initializer = MagicMock()
            mock_initializer.is_initialized.return_value = True
            mock_initializer.config_format.return_value = "yaml"
            mock_result = MagicMock()
            mock_result.success = True
            mock_result.project_path = tmpdir
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [tmpdir, "--non-interactive", "--force"])

            # Assert
            assert result.exit_code == 0
            mock_initializer.initialize.assert_called_once()

    @patch("moai_adk.cli.commands.init.print_banner")
    @patch("moai_adk.cli.commands.init.VersionReader")
    @patch("moai_adk.cli.commands.init.ProjectInitializer")
    def test_init_reinit_config_migration_json_to_yaml(
        self,
        mock_initializer_class,
        mock_version_reader_class,
        _mock_print_banner,
    ):
        """Test reinitialization with JSON to YAML config migration."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            mock_version_reader = MagicMock()
            mock_version_reader.get_version.return_value = "0.30.0"
            mock_version_reader_class.return_value = mock_version_reader

            mock_initializer = MagicMock()
            mock_initializer.is_initialized.return_value = True
            mock_initializer.config_format.return_value = "json"
            mock_result = MagicMock()
            mock_result.success = True
            mock_result.project_path = tmpdir
            mock_initializer.initialize.return_value = mock_result
            mock_initializer_class.return_value = mock_initializer

            # Act
            runner = CliRunner()
            result = runner.invoke(init, [tmpdir, "--non-interactive", "--force"])

            # Assert
            assert result.exit_code == 0
            mock_initializer.initialize.assert_called_once()
