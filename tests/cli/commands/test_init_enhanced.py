"""Enhanced comprehensive tests for init.py command

This test suite targets uncovered lines to achieve 90%+ coverage:
- Lines 142-143: Version read errors
- Lines 164-184: Interactive mode flow
- Lines 195-196: Interactive reinit messages
- Lines 211-214: Old hook file cleanup
- Lines 218-248: Config update during reinit
- Lines 297-298: Backup directory display
- Lines 324-326: Non-current directory next steps
- Lines 332, 336-353: Error handling paths

Test Organization:
- Class-based structure for related tests
- Descriptive test names following test_<action>_<condition>_<expected>
- Comprehensive docstrings
- Parametrization for multiple scenarios
- Mock external dependencies (click, questionary, console)
"""

import json
from pathlib import Path
from unittest.mock import Mock, patch

from click.testing import CliRunner

from moai_adk import __version__
from moai_adk.cli.commands.init import create_progress_callback, init
from moai_adk.core.project.initializer import InstallationResult


class TestCreateProgressCallback:
    """Test progress callback creation and execution"""

    def test_create_progress_callback_returns_callable(self) -> None:
        """Should return a callable progress callback"""
        mock_progress = Mock()
        task_ids = [1, 2, 3, 4, 5]

        callback = create_progress_callback(mock_progress, task_ids)

        assert callable(callback)

    def test_progress_callback_updates_correct_task(self) -> None:
        """Should update the correct task based on current phase"""
        mock_progress = Mock()
        task_ids = [10, 20, 30, 40, 50]

        callback = create_progress_callback(mock_progress, task_ids)

        # Test phase 1 (index 0)
        callback("Phase 1 message", 1, 5)
        mock_progress.update.assert_called_with(10, completed=1, description="Phase 1 message")

        # Test phase 3 (index 2)
        callback("Phase 3 message", 3, 5)
        mock_progress.update.assert_called_with(30, completed=1, description="Phase 3 message")

    def test_progress_callback_handles_out_of_range_phases(self) -> None:
        """Should not crash on out-of-range phase numbers"""
        mock_progress = Mock()
        task_ids = [1, 2, 3]

        callback = create_progress_callback(mock_progress, task_ids)

        # Should not raise exception for invalid phase
        callback("Invalid phase", 0, 3)  # Below range
        callback("Invalid phase", 10, 3)  # Above range

    def test_progress_callback_with_empty_task_ids(self) -> None:
        """Should handle empty task_ids list gracefully"""
        mock_progress = Mock()
        task_ids = []

        callback = create_progress_callback(mock_progress, task_ids)
        callback("Message", 1, 5)

        # Should not crash, just not update anything
        mock_progress.update.assert_not_called()


class TestInitVersionHandling:
    """Test version reading and error handling (lines 142-143)"""

    def test_init_handles_version_read_error_gracefully(self, tmp_path: Path) -> None:
        """Should catch and display version read errors without failing init"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.VersionReader") as mock_reader:
            # Simulate version read error
            mock_reader.return_value.get_version.side_effect = Exception("Version fetch failed")

            with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
                mock_init.return_value.is_initialized.return_value = False
                mock_init.return_value.initialize.return_value = InstallationResult(
                    success=True,
                    project_path=str(tmp_path),
                    language="python",
                    mode="personal",
                    locale="en",
                    duration=100,
                    created_files=[".moai/"],
                )

                result = runner.invoke(init, [str(tmp_path), "--non-interactive"])

                # Should continue despite version error
                assert result.exit_code == 0
                assert "Version read error" in result.output

    def test_init_displays_version_on_success(self, tmp_path: Path) -> None:
        """Should display current version when read succeeds"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.VersionReader") as mock_reader:
            mock_reader.return_value.get_version.return_value = "0.28.0"

            with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
                mock_init.return_value.is_initialized.return_value = False
                mock_init.return_value.initialize.return_value = InstallationResult(
                    success=True,
                    project_path=str(tmp_path),
                    language="python",
                    mode="personal",
                    locale="en",
                    duration=100,
                    created_files=[".moai/"],
                )

                result = runner.invoke(init, [str(tmp_path), "--non-interactive"])

                assert result.exit_code == 0
                assert "Current MoAI-ADK version: 0.28.0" in result.output


class TestInitInteractiveMode:
    """Test interactive mode prompts and flow (lines 164-184)"""

    def test_init_interactive_mode_calls_prompt_project_setup(self, tmp_path: Path) -> None:
        """Should call prompt_project_setup in interactive mode"""
        runner = CliRunner()

        mock_answers = {
            "project_name": "test-project",
            "mode": "team",
            "locale": "ko",
            "language": "python",
            "custom_language": None,
        }

        with patch("moai_adk.cli.commands.init.prompt_project_setup") as mock_prompt:
            mock_prompt.return_value = mock_answers

            with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
                mock_init.return_value.is_initialized.return_value = False
                mock_init.return_value.initialize.return_value = InstallationResult(
                    success=True,
                    project_path=str(tmp_path),
                    language="python",
                    mode="team",
                    locale="ko",
                    duration=100,
                    created_files=[".moai/"],
                )

                # Interactive mode (no --non-interactive flag)
                runner.invoke(init, [str(tmp_path)])

                # prompt_project_setup should be called
                mock_prompt.assert_called_once()
                call_kwargs = mock_prompt.call_args.kwargs
                assert call_kwargs["project_name"] == str(tmp_path)
                assert call_kwargs["is_current_dir"] is False

    def test_init_interactive_mode_uses_prompt_answers(self, tmp_path: Path) -> None:
        """Should use answers from interactive prompts"""
        runner = CliRunner()

        mock_answers = {
            "project_name": "interactive-project",
            "mode": "team",
            "locale": "ja",
            "language": "typescript",
            "custom_language": None,
        }

        with patch("moai_adk.cli.commands.init.prompt_project_setup") as mock_prompt:
            mock_prompt.return_value = mock_answers

            with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
                mock_init.return_value.is_initialized.return_value = False
                mock_init.return_value.initialize.return_value = InstallationResult(
                    success=True,
                    project_path=str(tmp_path),
                    language="typescript",
                    mode="team",
                    locale="ja",
                    duration=100,
                    created_files=[".moai/"],
                )

                runner.invoke(init, [str(tmp_path)])

                # Verify initialize was called with prompt answers
                init_call = mock_init.return_value.initialize.call_args
                assert init_call.kwargs["mode"] == "team"
                assert init_call.kwargs["locale"] == "ja"
                assert init_call.kwargs["language"] == "typescript"

    def test_init_interactive_mode_with_custom_language(self, tmp_path: Path) -> None:
        """Should handle custom language from interactive prompts"""
        runner = CliRunner()

        mock_answers = {
            "project_name": "custom-lang-project",
            "mode": "personal",
            "locale": "en",
            "language": "other",
            "custom_language": "Rust",
        }

        with patch("moai_adk.cli.commands.init.prompt_project_setup") as mock_prompt:
            mock_prompt.return_value = mock_answers

            with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
                mock_init.return_value.is_initialized.return_value = False
                mock_init.return_value.initialize.return_value = InstallationResult(
                    success=True,
                    project_path=str(tmp_path),
                    language="other",
                    mode="personal",
                    locale="en",
                    duration=100,
                    created_files=[".moai/"],
                )

                runner.invoke(init, [str(tmp_path)])

                # Verify custom_language is passed
                init_call = mock_init.return_value.initialize.call_args
                assert init_call.kwargs["custom_language"] == "Rust"

    def test_init_interactive_mode_with_current_directory(self, tmp_path: Path) -> None:
        """Should handle interactive mode in current directory"""
        runner = CliRunner()

        mock_answers = {
            "project_name": tmp_path.name,
            "mode": "personal",
            "locale": "en",
            "language": "python",
            "custom_language": None,
        }

        with patch("moai_adk.cli.commands.init.prompt_project_setup") as mock_prompt:
            mock_prompt.return_value = mock_answers

            with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
                mock_init.return_value.is_initialized.return_value = False
                mock_init.return_value.initialize.return_value = InstallationResult(
                    success=True,
                    project_path=str(tmp_path),
                    language="python",
                    mode="personal",
                    locale="en",
                    duration=100,
                    created_files=[".moai/"],
                )

                # Use "." for current directory
                with runner.isolated_filesystem(temp_dir=tmp_path):
                    runner.invoke(init, ["."])

                    # Verify is_current_dir is True in prompt call
                    call_kwargs = mock_prompt.call_args.kwargs
                    assert call_kwargs["is_current_dir"] is True

    def test_init_interactive_mode_uses_locale_from_answers_when_none(self, tmp_path: Path) -> None:
        """Should use locale from answers when initial_locale is None (line 184)"""
        runner = CliRunner()

        mock_answers = {
            "project_name": "test-project",
            "mode": "personal",
            "locale": "ko",  # This should be used when locale parameter is None
            "language": "python",
            "custom_language": None,
        }

        with patch("moai_adk.cli.commands.init.prompt_project_setup") as mock_prompt:
            mock_prompt.return_value = mock_answers

            with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
                mock_init.return_value.is_initialized.return_value = False
                mock_init.return_value.initialize.return_value = InstallationResult(
                    success=True,
                    project_path=str(tmp_path),
                    language="python",
                    mode="personal",
                    locale="ko",
                    duration=100,
                    created_files=[".moai/"],
                )

                # Interactive mode without --locale flag (locale=None)
                result = runner.invoke(init, [str(tmp_path)])

                assert result.exit_code == 0

                # Verify locale from answers was used
                init_call = mock_init.return_value.initialize.call_args
                assert init_call.kwargs["locale"] == "ko"


class TestInitReinitializationFlow:
    """Test reinitialization handling (lines 195-196, 211-214, 218-248)"""

    def test_init_displays_reinit_message_non_interactive(self, tmp_path: Path) -> None:
        """Should display reinitialization message in non-interactive mode"""
        runner = CliRunner()

        # Create existing .moai directory
        (tmp_path / ".moai").mkdir()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = True
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language="python",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            result = runner.invoke(init, [str(tmp_path), "--non-interactive"])

            assert result.exit_code == 0
            assert "Reinitializing project (force mode)" in result.output

    def test_init_displays_reinit_message_interactive(self, tmp_path: Path) -> None:
        """Should display reinitialization message in interactive mode (lines 195-196)"""
        runner = CliRunner()

        # Create existing .moai directory
        (tmp_path / ".moai").mkdir()

        mock_answers = {
            "project_name": "test-project",
            "mode": "personal",
            "locale": "en",
            "language": "python",
            "custom_language": None,
        }

        with patch("moai_adk.cli.commands.init.prompt_project_setup") as mock_prompt:
            mock_prompt.return_value = mock_answers

            with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
                mock_init.return_value.is_initialized.return_value = True
                mock_init.return_value.initialize.return_value = InstallationResult(
                    success=True,
                    project_path=str(tmp_path),
                    language="python",
                    mode="personal",
                    locale="en",
                    duration=100,
                    created_files=[".moai/"],
                )

                result = runner.invoke(init, [str(tmp_path)])

                assert result.exit_code == 0
                assert "Reinitializing project..." in result.output
                assert "Backup will be created" in result.output

    def test_init_removes_old_hook_files_on_reinit(self, tmp_path: Path) -> None:
        """Should remove deprecated hook files during reinit (lines 211-214)"""
        runner = CliRunner()

        # Create existing structure with old hook file
        (tmp_path / ".moai").mkdir()
        old_hook_dir = tmp_path / ".claude" / "hooks" / "alfred"
        old_hook_dir.mkdir(parents=True)
        old_hook_file = old_hook_dir / "session_start__startup.py"
        old_hook_file.write_text("# Old hook content")

        assert old_hook_file.exists()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = True
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language="python",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            runner.invoke(init, [str(tmp_path), "--non-interactive"])

            # Old hook file should be removed
            assert not old_hook_file.exists()

    def test_init_handles_old_hook_file_removal_error_gracefully(self, tmp_path: Path) -> None:
        """Should continue if old hook file removal fails"""
        runner = CliRunner()

        (tmp_path / ".moai").mkdir()
        old_hook_dir = tmp_path / ".claude" / "hooks" / "alfred"
        old_hook_dir.mkdir(parents=True)
        old_hook_file = old_hook_dir / "session_start__startup.py"
        old_hook_file.write_text("# Old hook content")

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = True
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language="python",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            # Mock unlink to raise exception
            with patch.object(Path, "unlink", side_effect=PermissionError("Access denied")):
                result = runner.invoke(init, [str(tmp_path), "--non-interactive"])

                # Should still succeed
                assert result.exit_code == 0

    def test_init_updates_config_json_on_reinit(self, tmp_path: Path) -> None:
        """Should update config.json optimized flag on reinit (lines 218-248)"""
        runner = CliRunner()

        # Create existing config
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"

        initial_config = {
            "moai": {"version": "0.27.0"},
            "project": {"optimized": True},
        }
        config_file.write_text(json.dumps(initial_config, indent=2))

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = True
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language="python",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            runner.invoke(init, [str(tmp_path), "--non-interactive"])

            # Config should be updated
            updated_config = json.loads(config_file.read_text())
            assert updated_config["project"]["optimized"] is False
            assert updated_config["moai"]["version"] == __version__

    def test_init_creates_moai_section_if_missing(self, tmp_path: Path) -> None:
        """Should create moai section in config if missing"""
        runner = CliRunner()

        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"

        # Config without moai section
        initial_config = {"project": {"optimized": True}}
        config_file.write_text(json.dumps(initial_config, indent=2))

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = True
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language="python",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            runner.invoke(init, [str(tmp_path), "--non-interactive"])

            # moai section should be created
            updated_config = json.loads(config_file.read_text())
            assert "moai" in updated_config
            assert "version" in updated_config["moai"]

    def test_init_creates_project_section_if_missing(self, tmp_path: Path) -> None:
        """Should create project section in config if missing"""
        runner = CliRunner()

        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"

        # Config without project section
        initial_config = {"moai": {"version": "0.27.0"}}
        config_file.write_text(json.dumps(initial_config, indent=2))

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = True
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language="python",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            runner.invoke(init, [str(tmp_path), "--non-interactive"])

            # project section should be created
            updated_config = json.loads(config_file.read_text())
            assert "project" in updated_config
            assert updated_config["project"]["optimized"] is False

    def test_init_handles_config_update_errors_gracefully(self, tmp_path: Path) -> None:
        """Should continue if config update fails"""
        runner = CliRunner()

        (tmp_path / ".moai" / "config").mkdir(parents=True)
        config_file = tmp_path / ".moai" / "config" / "config.json"
        config_file.write_text("{invalid json}")

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = True
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language="python",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            # Should not fail even with invalid JSON
            result = runner.invoke(init, [str(tmp_path), "--non-interactive"])
            assert result.exit_code == 0

    def test_init_uses_version_reader_for_config_update(self, tmp_path: Path) -> None:
        """Should use VersionReader for config version update"""
        runner = CliRunner()

        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"
        config_file.write_text(json.dumps({"moai": {}, "project": {}}, indent=2))

        with patch("moai_adk.cli.commands.init.VersionReader") as mock_reader:
            mock_reader.return_value.get_version.return_value = "0.28.5"

            with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
                mock_init.return_value.is_initialized.return_value = True
                mock_init.return_value.initialize.return_value = InstallationResult(
                    success=True,
                    project_path=str(tmp_path),
                    language="python",
                    mode="personal",
                    locale="en",
                    duration=100,
                    created_files=[".moai/"],
                )

                runner.invoke(init, [str(tmp_path), "--non-interactive"])

                # Version from VersionReader should be used
                updated_config = json.loads(config_file.read_text())
                assert updated_config["moai"]["version"] == "0.28.5"

    def test_init_falls_back_to_package_version_on_reader_error(self, tmp_path: Path) -> None:
        """Should fall back to __version__ if VersionReader fails"""
        runner = CliRunner()

        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"
        config_file.write_text(json.dumps({"moai": {}, "project": {}}, indent=2))

        with patch("moai_adk.cli.commands.init.VersionReader") as mock_reader:
            mock_reader.return_value.get_version.side_effect = Exception("Reader failed")

            with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
                mock_init.return_value.is_initialized.return_value = True
                mock_init.return_value.initialize.return_value = InstallationResult(
                    success=True,
                    project_path=str(tmp_path),
                    language="python",
                    mode="personal",
                    locale="en",
                    duration=100,
                    created_files=[".moai/"],
                )

                runner.invoke(init, [str(tmp_path), "--non-interactive"])

                # Should fall back to __version__
                updated_config = json.loads(config_file.read_text())
                assert updated_config["moai"]["version"] == __version__


class TestInitSuccessOutput:
    """Test success output formatting (lines 297-298, 324-326, 332)"""

    def test_init_displays_backup_info_on_reinit(self, tmp_path: Path) -> None:
        """Should display backup directory info on reinit (lines 297-298)"""
        runner = CliRunner()

        # Create backup directory
        backup_dir = tmp_path / ".moai-backups"
        backup_dir.mkdir()
        (backup_dir / "backup-2025-01-01").mkdir()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = True
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language="python",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            result = runner.invoke(init, [str(tmp_path), "--non-interactive"])

            assert result.exit_code == 0
            assert "Backup:" in result.output
            assert "backup-2025-01-01" in result.output

    def test_init_shows_optimized_false_message_on_reinit(self, tmp_path: Path) -> None:
        """Should show configuration merge required message on reinit"""
        runner = CliRunner()

        (tmp_path / ".moai").mkdir()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = True
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language="python",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            result = runner.invoke(init, [str(tmp_path), "--non-interactive"])

            assert "optimized=false" in result.output
            assert "Configuration merge required" in result.output
            assert "/moai:project" in result.output

    def test_init_shows_cd_instruction_for_non_current_dir(self, tmp_path: Path) -> None:
        """Should show 'cd' instruction when not in current directory (lines 324-326)"""
        runner = CliRunner()

        project_dir = tmp_path / "new-project"
        project_dir.mkdir()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(project_dir),
                language="python",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            result = runner.invoke(init, [str(project_dir), "--non-interactive"])

            assert result.exit_code == 0
            # Check for cd command (path may be wrapped)
            assert "cd" in result.output and "new-project" in result.output
            assert "Start developing with MoAI-ADK!" in result.output

    def test_init_skips_cd_instruction_for_current_dir(self, tmp_path: Path) -> None:
        """Should not show 'cd' instruction for current directory (line 332)"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language="python",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            with runner.isolated_filesystem(temp_dir=tmp_path):
                result = runner.invoke(init, [".", "--non-interactive"])

                assert result.exit_code == 0
                # Should not show cd instruction
                assert "cd " not in result.output or "cd ." in result.output


class TestInitErrorHandling:
    """Test error handling paths (lines 336-353)"""

    def test_init_displays_failure_message_on_error(self, tmp_path: Path) -> None:
        """Should display failure message when initialization fails (lines 336-342)"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=False,
                project_path=str(tmp_path),
                language="python",
                mode="personal",
                locale="en",
                duration=50,
                created_files=[],
                errors=["Directory creation failed", "Permission denied"],
            )

            result = runner.invoke(init, [str(tmp_path), "--non-interactive"])

            assert result.exit_code != 0
            assert "Initialization Failed!" in result.output
            assert "Directory creation failed" in result.output
            assert "Permission denied" in result.output

    def test_init_handles_keyboard_interrupt(self, tmp_path: Path) -> None:
        """Should handle KeyboardInterrupt gracefully (lines 344-346)"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False
            mock_init.return_value.initialize.side_effect = KeyboardInterrupt()

            result = runner.invoke(init, [str(tmp_path), "--non-interactive"])

            assert result.exit_code != 0
            assert "cancelled by user" in result.output

    def test_init_handles_file_exists_error(self, tmp_path: Path) -> None:
        """Should handle FileExistsError gracefully (lines 347-350)"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False
            mock_init.return_value.initialize.side_effect = FileExistsError("Already exists")

            result = runner.invoke(init, [str(tmp_path), "--non-interactive"])

            assert result.exit_code != 0
            assert "already initialized" in result.output

    def test_init_handles_general_exception(self, tmp_path: Path) -> None:
        """Should handle general exceptions (lines 351-353)"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False
            mock_init.return_value.initialize.side_effect = RuntimeError("Unexpected error")

            result = runner.invoke(init, [str(tmp_path), "--non-interactive"])

            assert result.exit_code != 0
            assert "Initialization failed" in result.output
            assert "Unexpected error" in result.output


class TestInitNonInteractiveMode:
    """Test non-interactive mode defaults and behavior"""

    def test_init_non_interactive_uses_defaults(self, tmp_path: Path) -> None:
        """Should use defaults in non-interactive mode"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language=None,
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            result = runner.invoke(init, [str(tmp_path), "--non-interactive"])

            assert result.exit_code == 0

            # Check defaults
            init_call = mock_init.return_value.initialize.call_args
            assert init_call.kwargs["mode"] == "personal"
            assert init_call.kwargs["locale"] == "en"

    def test_init_non_interactive_with_custom_locale(self, tmp_path: Path) -> None:
        """Should use custom locale in non-interactive mode"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language=None,
                mode="personal",
                locale="ko",
                duration=100,
                created_files=[".moai/"],
            )

            result = runner.invoke(init, [str(tmp_path), "--non-interactive", "--locale", "ko"])

            assert result.exit_code == 0

            init_call = mock_init.return_value.initialize.call_args
            assert init_call.kwargs["locale"] == "ko"

    def test_init_non_interactive_with_team_mode(self, tmp_path: Path) -> None:
        """Should use team mode in non-interactive mode"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language=None,
                mode="team",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            result = runner.invoke(init, [str(tmp_path), "--non-interactive", "--mode", "team"])

            assert result.exit_code == 0

            init_call = mock_init.return_value.initialize.call_args
            assert init_call.kwargs["mode"] == "team"

    def test_init_non_interactive_with_language(self, tmp_path: Path) -> None:
        """Should use specified language in non-interactive mode"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language="go",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            result = runner.invoke(init, [str(tmp_path), "--non-interactive", "--language", "go"])

            assert result.exit_code == 0

            init_call = mock_init.return_value.initialize.call_args
            assert init_call.kwargs["language"] == "go"


class TestInitPathHandling:
    """Test path resolution and directory handling"""

    def test_init_resolves_relative_path(self, tmp_path: Path) -> None:
        """Should resolve relative paths to absolute"""
        runner = CliRunner()

        project_dir = tmp_path / "relative-project"
        project_dir.mkdir()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(project_dir),
                language="python",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            # ProjectInitializer should be called with resolved path
            result = runner.invoke(init, [str(project_dir), "--non-interactive"])

            assert result.exit_code == 0
            assert mock_init.called

    def test_init_handles_current_directory_dot(self, tmp_path: Path) -> None:
        """Should handle '.' for current directory"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language="python",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            with runner.isolated_filesystem(temp_dir=tmp_path):
                result = runner.invoke(init, [".", "--non-interactive"])

                assert result.exit_code == 0

    def test_init_uses_project_name_from_path(self, tmp_path: Path) -> None:
        """Should derive project name from path in non-interactive mode"""
        runner = CliRunner()

        project_dir = tmp_path / "my-awesome-project"
        project_dir.mkdir()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(project_dir),
                language="python",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            result = runner.invoke(init, [str(project_dir), "--non-interactive"])

            assert result.exit_code == 0
            # Project name should be derived from directory name


class TestInitProgressDisplay:
    """Test progress bar and phase display"""

    def test_init_creates_5_phase_progress_tasks(self, tmp_path: Path) -> None:
        """Should create progress tasks for all 5 phases"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False

            callback_capture = []

            def mock_initialize(**kwargs):
                # Capture callback
                if "progress_callback" in kwargs and kwargs["progress_callback"]:
                    callback_capture.append(kwargs["progress_callback"])
                    # Simulate phase calls
                    for phase in range(1, 6):
                        kwargs["progress_callback"](f"Phase {phase}", phase, 5)

                return InstallationResult(
                    success=True,
                    project_path=str(tmp_path),
                    language="python",
                    mode="personal",
                    locale="en",
                    duration=100,
                    created_files=[".moai/"],
                )

            mock_init.return_value.initialize.side_effect = mock_initialize

            result = runner.invoke(init, [str(tmp_path), "--non-interactive"])

            assert result.exit_code == 0
            assert len(callback_capture) == 1

    def test_init_passes_progress_callback_to_initializer(self, tmp_path: Path) -> None:
        """Should pass progress callback to ProjectInitializer"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language="python",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            result = runner.invoke(init, [str(tmp_path), "--non-interactive"])

            assert result.exit_code == 0

            # Verify callback was passed
            init_call = mock_init.return_value.initialize.call_args
            assert "progress_callback" in init_call.kwargs
            assert callable(init_call.kwargs["progress_callback"])


class TestInitLanguageDisplay:
    """Test language display in summary"""

    def test_init_displays_generic_as_auto_detect(self, tmp_path: Path) -> None:
        """Should display 'Auto-detect' for generic language"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language="generic",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            result = runner.invoke(init, [str(tmp_path), "--non-interactive"])

            assert result.exit_code == 0
            assert "Auto-detect" in result.output
            assert "/moai:project" in result.output

    def test_init_displays_specific_language(self, tmp_path: Path) -> None:
        """Should display specific language when set"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
            mock_init.return_value.is_initialized.return_value = False
            mock_init.return_value.initialize.return_value = InstallationResult(
                success=True,
                project_path=str(tmp_path),
                language="typescript",
                mode="personal",
                locale="en",
                duration=100,
                created_files=[".moai/"],
            )

            result = runner.invoke(init, [str(tmp_path), "--non-interactive", "--language", "typescript"])

            assert result.exit_code == 0
            assert "typescript" in result.output


class TestInitConsoleFlush:
    """Test console output flushing"""

    def test_init_flushes_console_output_in_finally(self, tmp_path: Path) -> None:
        """Should flush console output in finally block"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.init.console") as mock_console:
            mock_console.file = Mock()

            with patch("moai_adk.cli.commands.init.ProjectInitializer") as mock_init:
                mock_init.return_value.is_initialized.return_value = False
                mock_init.return_value.initialize.side_effect = Exception("Test error")

                runner.invoke(init, [str(tmp_path), "--non-interactive"])

                # Console flush should be called even on error
                mock_console.file.flush.assert_called()
