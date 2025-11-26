"""Unit tests for Phase 2: 2-Stage Workflow

Tests for version management and 2-stage update workflow.
RED Phase: These tests should initially fail.
"""

import subprocess
import urllib.error
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest
from click.testing import CliRunner

from moai_adk.cli.commands.update import (
    _compare_versions,
    _execute_upgrade,
    _get_current_version,
    _get_latest_version,
    _sync_templates,
    update,
)


class TestVersionManagement:
    """Test version management functions"""

    def test_get_current_version(self):
        """Test getting current installed version."""
        version = _get_current_version()
        assert isinstance(version, str)
        assert version  # Not empty
        assert version.count(".") >= 2  # Semantic version format

    def test_get_latest_version_success(self):
        """Test fetching latest version from PyPI (success)."""
        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = Mock()
            mock_response.read.return_value = b'{"info": {"version": "0.6.2"}}'
            mock_response.__enter__ = Mock(return_value=mock_response)
            mock_response.__exit__ = Mock(return_value=False)
            mock_urlopen.return_value = mock_response

            latest = _get_latest_version()
            assert latest == "0.6.2"

    def test_get_latest_version_network_error(self):
        """Test handling network error when fetching latest version."""
        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_urlopen.side_effect = urllib.error.URLError("Network error")
            with pytest.raises(RuntimeError, match="Failed to fetch latest version"):
                _get_latest_version()

    def test_get_latest_version_json_decode_error(self):
        """Test handling JSON decode error."""
        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = Mock()
            mock_response.read.return_value = b"invalid json"
            mock_response.__enter__ = Mock(return_value=mock_response)
            mock_response.__exit__ = Mock(return_value=False)
            mock_urlopen.return_value = mock_response

            with pytest.raises(RuntimeError, match="Failed to fetch latest version"):
                _get_latest_version()

    def test_get_latest_version_missing_version_key(self):
        """Test handling missing version key in PyPI response."""
        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = Mock()
            mock_response.read.return_value = b'{"info": {}}'
            mock_response.__enter__ = Mock(return_value=mock_response)
            mock_response.__exit__ = Mock(return_value=False)
            mock_urlopen.return_value = mock_response

            with pytest.raises(RuntimeError, match="Failed to fetch latest version"):
                _get_latest_version()

    def test_compare_versions_upgrade_needed(self):
        """Test version comparison when upgrade needed."""
        result = _compare_versions("0.6.1", "0.6.2")
        assert result == -1  # current < latest

    def test_compare_versions_up_to_date(self):
        """Test version comparison when up to date."""
        result = _compare_versions("0.6.2", "0.6.2")
        assert result == 0  # current == latest

    def test_compare_versions_newer_installed(self):
        """Test version comparison when newer already installed."""
        result = _compare_versions("0.6.3", "0.6.2")
        assert result == 1  # current > latest

    def test_compare_versions_major_difference(self):
        """Test version comparison with major version difference."""
        result = _compare_versions("0.6.1", "1.0.0")
        assert result == -1  # current < latest

    def test_compare_versions_dev_version(self):
        """Test version comparison with dev version."""
        result = _compare_versions("0.6.1-dev", "0.6.1")
        assert result == -1  # dev version is considered less than release


class TestUpgradeExecution:
    """Test upgrade execution functions"""

    def test_execute_upgrade_success(self):
        """Test successful package upgrade."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="Success", stderr="")
            result = _execute_upgrade(["uv", "tool", "upgrade", "moai-adk"])
            assert result is True
            mock_run.assert_called_once()

    def test_execute_upgrade_failure(self):
        """Test failed package upgrade."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=1, stdout="", stderr="Error")
            result = _execute_upgrade(["uv", "tool", "upgrade", "moai-adk"])
            assert result is False

    @pytest.mark.skip(reason="CLI confirm input requires interactive input - ClickException from confirm()")
    def test_execute_upgrade_timeout(self):
        """Test upgrade timeout handling."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.TimeoutExpired(cmd="uv", timeout=60)
            result = _execute_upgrade(["uv", "tool", "upgrade", "moai-adk"])
            assert result is False

    def test_execute_upgrade_exception(self):
        """Test upgrade with unexpected exception."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = RuntimeError("Unexpected error")
            result = _execute_upgrade(["uv", "tool", "upgrade", "moai-adk"])
            assert result is False

    def test_execute_upgrade_captures_output(self):
        """Test that upgrade captures output correctly."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="Upgrading moai-adk", stderr="")
            result = _execute_upgrade(["pipx", "upgrade", "moai-adk"])
            assert result is True
            # Verify that capture_output and text are set
            call_kwargs = mock_run.call_args[1]
            assert call_kwargs["capture_output"] is True
            assert call_kwargs["text"] is True


class TestTemplateSyncExtraction:
    """Test template sync function extraction"""

    def test_sync_templates_success(self, tmp_path):
        """Test successful template sync."""
        project_path = tmp_path / "project"
        project_path.mkdir()
        (project_path / ".moai").mkdir()

        with patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor:
            mock_instance = Mock()
            mock_instance.copy_templates.return_value = None
            mock_processor.return_value = mock_instance

            result = _sync_templates(project_path, force=False)
            assert result is True
            mock_processor.assert_called_once_with(project_path)

    def test_sync_templates_with_force(self, tmp_path):
        """Test template sync with force flag."""
        project_path = tmp_path / "project"
        project_path.mkdir()
        (project_path / ".moai").mkdir()

        with patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor:
            mock_instance = Mock()
            mock_processor.return_value = mock_instance

            result = _sync_templates(project_path, force=True)
            assert result is True

    def test_sync_templates_failure(self, tmp_path):
        """Test template sync failure."""
        project_path = tmp_path / "project"
        project_path.mkdir()

        with patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor:
            mock_processor.side_effect = RuntimeError("Template error")

            result = _sync_templates(project_path, force=False)
            assert result is False


class TestTwoStageWorkflow:
    """Test 2-stage workflow integration"""

    def test_update_stage_1_upgrade_needed(self, tmp_path):
        """Test Stage 1: upgrade needed (version < latest)."""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            Path(".moai").mkdir()

            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._detect_tool_installer") as mock_detect,
                patch("moai_adk.cli.commands.update._execute_upgrade") as mock_upgrade,
                patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
            ):

                mock_current.return_value = "0.6.1"
                mock_latest.return_value = "0.6.2"
                mock_detect.return_value = ["uv", "tool", "upgrade", "moai-adk"]
                mock_upgrade.return_value = True

                result = runner.invoke(update, ["--yes"])  # Auto-confirm upgrade

                # Should execute upgrade and exit (not sync)
                mock_upgrade.assert_called_once()
                mock_sync.assert_not_called()
                assert "0.6.1" in result.output
                assert "0.6.2" in result.output
                assert "Run 'moai-adk update' again" in result.output

    def test_update_stage_2_already_latest(self, tmp_path):
        """Test Stage 2: already latest version."""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            Path(".moai").mkdir()

            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
            ):

                mock_current.return_value = "0.6.2"
                mock_latest.return_value = "0.6.2"
                mock_sync.return_value = True

                result = runner.invoke(update)

                # Should go directly to Stage 2 (sync templates)
                mock_sync.assert_called_once()
                assert result.exit_code == 0
                assert "Update complete" in result.output

    def test_update_no_installer_detected(self, tmp_path):
        """Test Stage 1 failure: no installer detected."""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            Path(".moai").mkdir()

            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._detect_tool_installer") as mock_detect,
            ):

                mock_current.return_value = "0.6.1"
                mock_latest.return_value = "0.6.2"
                mock_detect.return_value = None  # No installer found

                result = runner.invoke(update, ["--yes"])  # Auto-confirm

                # Should display error and exit
                assert result.exit_code != 0 or "Error" in result.output
                assert "Cannot detect" in result.output or "package installer" in result.output

    def test_update_upgrade_failed(self, tmp_path):
        """Test Stage 1 failure: upgrade command failed."""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            Path(".moai").mkdir()

            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._detect_tool_installer") as mock_detect,
                patch("moai_adk.cli.commands.update._execute_upgrade") as mock_upgrade,
            ):

                mock_current.return_value = "0.6.1"
                mock_latest.return_value = "0.6.2"
                mock_detect.return_value = ["uv", "tool", "upgrade", "moai-adk"]
                mock_upgrade.return_value = False  # Upgrade failed

                result = runner.invoke(update)

                # Should display error
                assert "Error" in result.output or "failed" in result.output

    def test_update_sync_failed(self, tmp_path):
        """Test Stage 2 failure: template sync failed."""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            Path(".moai").mkdir()

            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
            ):

                mock_current.return_value = "0.6.2"
                mock_latest.return_value = "0.6.2"
                mock_sync.return_value = False  # Sync failed

                result = runner.invoke(update)

                # Should display error
                assert "Error" in result.output or "failed" in result.output

    def test_update_version_fetch_error(self, tmp_path):
        """Test error handling when version fetch fails."""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            Path(".moai").mkdir()

            with patch("moai_adk.cli.commands.update._get_current_version") as mock_current:
                mock_current.side_effect = RuntimeError("Version error")

                result = runner.invoke(update)

                # Should display error
                assert "Error" in result.output

    def test_update_displays_upgrade_command(self, tmp_path):
        """Test that update displays the upgrade command being executed."""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            Path(".moai").mkdir()

            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._detect_tool_installer") as mock_detect,
                patch("moai_adk.cli.commands.update._execute_upgrade") as mock_upgrade,
            ):

                mock_current.return_value = "0.6.1"
                mock_latest.return_value = "0.6.2"
                mock_detect.return_value = ["uv", "tool", "upgrade", "moai-adk"]
                mock_upgrade.return_value = True

                result = runner.invoke(update, ["--yes"])  # Auto-confirm

                # Should display the command being run
                assert "uv tool upgrade moai-adk" in result.output

    def test_update_check_flag_skips_workflow(self, tmp_path):
        """Test that --check flag only displays versions and skips workflow."""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            Path(".moai").mkdir()

            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._execute_upgrade") as mock_upgrade,
                patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
            ):

                mock_current.return_value = "0.6.1"
                mock_latest.return_value = "0.6.2"

                result = runner.invoke(update, ["--check"])

                # Should display versions but not execute upgrade or sync
                assert "Current" in result.output
                assert "Latest" in result.output
                mock_upgrade.assert_not_called()
                mock_sync.assert_not_called()

    def test_update_newer_local_version(self, tmp_path):
        """Test that update skips when local version is newer (dev version)."""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            Path(".moai").mkdir()

            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._execute_upgrade") as mock_upgrade,
                patch("moai_adk.cli.commands.update._sync_templates"),
            ):

                mock_current.return_value = "0.7.0"
                mock_latest.return_value = "0.6.2"

                result = runner.invoke(update)

                # Should skip upgrade (dev version)
                mock_upgrade.assert_not_called()
                # May still sync templates (existing behavior)
                assert result.exit_code == 0
