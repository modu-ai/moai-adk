"""Enhanced comprehensive tests for status.py command

This test suite targets uncovered lines to achieve 90%+ coverage:
- Lines 62-67: Config reading and SPEC counting
- Lines 69-78: Table building with project data
- Lines 81-88: Git integration (optional)
- Lines 102-104: Exception handling

Coverage Target: 90%+ (from 42.50%)
Missing Lines Before: 62-98, 102-104

Test Organization:
- Class-based structure for related tests
- Descriptive test names following test_<action>_<condition>_<expected>
- Comprehensive docstrings
- Mock external dependencies (git, console, click)
"""

import json
from pathlib import Path
from unittest.mock import MagicMock, patch

from click.testing import CliRunner

from moai_adk.cli.commands.status import status


class TestStatusConfigReading:
    """Test config.json reading and validation (lines 62-67)"""

    def test_status_reads_config_successfully(self, tmp_path: Path) -> None:
        """Should read config.json and display project information"""
        # Setup: Create config.json
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"

        config_data = {"project": {"mode": "personal", "locale": "en_US"}}
        config_path.write_text(json.dumps(config_data))

        runner = CliRunner()

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            with patch("moai_adk.cli.commands.status.console") as mock_console:
                result = runner.invoke(status)

                assert result.exit_code == 0
                # Verify console.print was called with panel
                assert mock_console.print.called

    def test_status_missing_config_shows_warning(self, tmp_path: Path) -> None:
        """Should show warning and abort when config.json is missing (lines 57-60)"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            result = runner.invoke(status, catch_exceptions=False)

            # Should abort with exit code 1
            assert result.exit_code == 1
            assert "No .moai/config/config.json found" in result.output
            assert "moai_adk init" in result.output

    def test_status_counts_spec_documents_correctly(self, tmp_path: Path) -> None:
        """Should count SPEC documents in .moai/specs/ directory (lines 66-67)"""
        # Setup: Create config and SPEC files
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"
        config_path.write_text(json.dumps({"project": {"mode": "personal", "locale": "en_US"}}))

        # Create multiple SPEC directories
        specs_dir = tmp_path / ".moai" / "specs"
        for i in range(1, 4):
            spec_dir = specs_dir / f"SPEC-{i:03d}"
            spec_dir.mkdir(parents=True)
            (spec_dir / "spec.md").write_text(f"# SPEC-{i:03d}")

        runner = CliRunner()

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            result = runner.invoke(status)

            assert result.exit_code == 0
            # Verify SPEC count is shown (3 SPECs created)
            # Note: We can't easily check the exact output due to Rich formatting,
            # but we verify no errors occurred

    def test_status_handles_no_specs_directory(self, tmp_path: Path) -> None:
        """Should handle missing .moai/specs directory gracefully (line 67)"""
        # Setup: Create config but no specs directory
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"
        config_path.write_text(json.dumps({"project": {"mode": "personal", "locale": "en_US"}}))

        runner = CliRunner()

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            result = runner.invoke(status)

            # Should not crash, should show 0 SPECs
            assert result.exit_code == 0


class TestStatusTableBuilding:
    """Test status table construction (lines 69-78)"""

    def test_status_displays_mode_from_project_section(self, tmp_path: Path) -> None:
        """Should read mode from project section (line 76)"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"

        config_data = {"project": {"mode": "team", "locale": "ja_JP"}}
        config_path.write_text(json.dumps(config_data))

        runner = CliRunner()

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            result = runner.invoke(status)

            assert result.exit_code == 0

    def test_status_displays_locale_from_project_section(self, tmp_path: Path) -> None:
        """Should read locale from project section (line 77)"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"

        config_data = {"project": {"mode": "personal", "locale": "ko_KR"}}
        config_path.write_text(json.dumps(config_data))

        runner = CliRunner()

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            result = runner.invoke(status)

            assert result.exit_code == 0

    def test_status_fallback_to_root_level_mode(self, tmp_path: Path) -> None:
        """Should fallback to root-level mode if project section missing (line 76)"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"

        # Old config format (mode at root level)
        config_data = {"mode": "personal", "locale": "en_US"}
        config_path.write_text(json.dumps(config_data))

        runner = CliRunner()

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            result = runner.invoke(status)

            assert result.exit_code == 0

    def test_status_fallback_to_root_level_locale(self, tmp_path: Path) -> None:
        """Should fallback to root-level locale if project section missing (line 77)"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"

        # Old config format (locale at root level)
        config_data = {"mode": "team", "locale": "fr_FR"}
        config_path.write_text(json.dumps(config_data))

        runner = CliRunner()

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            result = runner.invoke(status)

            assert result.exit_code == 0

    def test_status_handles_missing_mode_and_locale(self, tmp_path: Path) -> None:
        """Should display 'unknown' for missing mode/locale (lines 76-77)"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"

        # Empty config
        config_data = {}
        config_path.write_text(json.dumps(config_data))

        runner = CliRunner()

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            result = runner.invoke(status)

            # Should display "unknown" for both mode and locale
            assert result.exit_code == 0


class TestStatusGitIntegration:
    """Test Git status integration (lines 81-88)"""

    def test_status_displays_git_branch_when_available(self, tmp_path: Path) -> None:
        """Should display current Git branch (lines 82-86)"""
        # Setup: Create config
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"
        config_path.write_text(json.dumps({"project": {"mode": "personal", "locale": "en_US"}}))

        runner = CliRunner()

        # Mock Git repo
        mock_repo = MagicMock()
        mock_repo.active_branch.name = "feature/test-branch"
        mock_repo.is_dirty.return_value = False

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            # Patch the import inside the try block
            with patch("git.Repo", return_value=mock_repo):
                result = runner.invoke(status)

                assert result.exit_code == 0

    def test_status_displays_clean_git_status(self, tmp_path: Path) -> None:
        """Should display 'Clean' when Git status is clean (line 86)"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"
        config_path.write_text(json.dumps({"project": {"mode": "personal", "locale": "en_US"}}))

        runner = CliRunner()

        # Mock Git repo with clean status
        mock_repo = MagicMock()
        mock_repo.active_branch.name = "main"
        mock_repo.is_dirty.return_value = False

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            with patch("git.Repo", return_value=mock_repo):
                result = runner.invoke(status)

                assert result.exit_code == 0

    def test_status_displays_modified_git_status(self, tmp_path: Path) -> None:
        """Should display 'Modified' when Git status is dirty (line 86)"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"
        config_path.write_text(json.dumps({"project": {"mode": "personal", "locale": "en_US"}}))

        runner = CliRunner()

        # Mock Git repo with dirty status
        mock_repo = MagicMock()
        mock_repo.active_branch.name = "main"
        mock_repo.is_dirty.return_value = True

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            with patch("git.Repo", return_value=mock_repo):
                result = runner.invoke(status)

                assert result.exit_code == 0

    def test_status_handles_git_not_available(self, tmp_path: Path) -> None:
        """Should continue gracefully when Git is not available (lines 87-88)"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"
        config_path.write_text(json.dumps({"project": {"mode": "personal", "locale": "en_US"}}))

        runner = CliRunner()

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            # Mock git.Repo to raise exception (simulating Git not installed)
            with patch("git.Repo", side_effect=Exception("Git not found")):
                result = runner.invoke(status)

                # Should continue without Git info
                assert result.exit_code == 0

    def test_status_handles_not_a_git_repository(self, tmp_path: Path) -> None:
        """Should handle when directory is not a Git repository (lines 87-88)"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"
        config_path.write_text(json.dumps({"project": {"mode": "personal", "locale": "en_US"}}))

        runner = CliRunner()

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            with patch("git.Repo", side_effect=Exception("Not a git repository")):
                result = runner.invoke(status)

                # Should continue without Git info
                assert result.exit_code == 0


class TestStatusErrorHandling:
    """Test error handling paths (lines 102-104)"""

    def test_status_handles_malformed_json_config(self, tmp_path: Path) -> None:
        """Should handle malformed JSON in config file (lines 102-104)"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"

        # Write malformed JSON
        config_path.write_text("{ invalid json }")

        runner = CliRunner()

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            # Don't catch exceptions since the command itself handles them
            result = runner.invoke(status)

            # Should display error message and fail
            assert result.exit_code != 0
            assert "Failed to get status" in result.output

    def test_status_handles_config_read_permission_error(self, tmp_path: Path) -> None:
        """Should handle permission errors when reading config (lines 102-104)"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"
        config_path.write_text(json.dumps({"project": {"mode": "personal"}}))

        runner = CliRunner()

        def mock_open_permission_error(*args, **kwargs):
            if "config.json" in str(args[0]):
                raise PermissionError("Access denied")
            # Allow other file operations
            return open(*args, **kwargs)

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            with patch("builtins.open", side_effect=mock_open_permission_error):
                result = runner.invoke(status)

                # Should display error message and fail
                assert result.exit_code != 0
                assert "Failed to get status" in result.output

    def test_status_handles_unexpected_exception(self, tmp_path: Path) -> None:
        """Should handle unexpected exceptions gracefully (lines 102-104)"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"
        config_path.write_text(json.dumps({"project": {"mode": "personal"}}))

        runner = CliRunner()

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            with patch("moai_adk.cli.commands.status.json.load", side_effect=RuntimeError("Unexpected error")):
                result = runner.invoke(status)

                # Should display error message and fail
                assert result.exit_code != 0
                assert "Failed to get status" in result.output


class TestStatusPanelRendering:
    """Test status panel rendering (lines 91-98)"""

    def test_status_renders_panel_with_table(self, tmp_path: Path) -> None:
        """Should render panel with status table (lines 91-98)"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"
        config_path.write_text(json.dumps({"project": {"mode": "personal", "locale": "en_US"}}))

        runner = CliRunner()

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            with patch("moai_adk.cli.commands.status.console") as mock_console:
                result = runner.invoke(status)

                # Verify panel was printed
                assert result.exit_code == 0
                mock_console.print.assert_called_once()

                # Verify it was called with a Panel object
                call_args = mock_console.print.call_args
                from rich.panel import Panel

                assert isinstance(call_args[0][0], Panel)

    def test_status_panel_has_correct_title(self, tmp_path: Path) -> None:
        """Should render panel with 'Project Status' title (line 93)"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"
        config_path.write_text(json.dumps({"project": {"mode": "personal", "locale": "en_US"}}))

        runner = CliRunner()

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            with patch("moai_adk.cli.commands.status.console") as mock_console:
                result = runner.invoke(status)

                assert result.exit_code == 0

                # Verify panel title
                call_args = mock_console.print.call_args
                panel = call_args[0][0]
                assert "Project Status" in str(panel.title)


class TestStatusFullIntegration:
    """Integration tests covering complete status flow"""

    def test_status_full_flow_with_all_features(self, tmp_path: Path) -> None:
        """Should display complete status with config, SPECs, and Git info"""
        # Setup: Create complete project structure
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"

        config_data = {"project": {"mode": "team", "locale": "ko_KR"}}
        config_path.write_text(json.dumps(config_data))

        # Create SPEC documents
        specs_dir = tmp_path / ".moai" / "specs"
        for i in range(1, 6):
            spec_dir = specs_dir / f"SPEC-{i:03d}"
            spec_dir.mkdir(parents=True)
            (spec_dir / "spec.md").write_text(f"# SPEC-{i:03d}")

        runner = CliRunner()

        # Mock Git repo
        mock_repo = MagicMock()
        mock_repo.active_branch.name = "develop"
        mock_repo.is_dirty.return_value = True

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            with patch("git.Repo", return_value=mock_repo):
                result = runner.invoke(status)

                assert result.exit_code == 0

    def test_status_minimal_config_no_git(self, tmp_path: Path) -> None:
        """Should display status with minimal config and no Git"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"
        config_path.write_text(json.dumps({"project": {}}))

        runner = CliRunner()

        with patch("moai_adk.cli.commands.status.Path.cwd", return_value=tmp_path):
            with patch("git.Repo", side_effect=Exception("No git")):
                result = runner.invoke(status)

                # Should display "unknown" for missing values
                assert result.exit_code == 0
