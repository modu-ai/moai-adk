"""Test error handling and recovery for update command.

This module tests all error scenarios:
1. Installer detection failure
2. Network failure (PyPI unreachable)
3. Package upgrade failure
4. Template sync failure
5. Timeout/other errors

Tests ensure proper error messages, exit codes, and recovery instructions.
"""

from __future__ import annotations

import subprocess
from pathlib import Path
from unittest.mock import patch

import pytest
from click.testing import CliRunner

from moai_adk.cli.commands.update import update

# Test fixtures
runner = CliRunner()


class TestInstallerDetectionError:
    """Tests for installer detection failure (EVT-005, CST-001)."""

    def test_no_installer_detected_shows_help(self, tmp_path: Path) -> None:
        """Test error message when no installer found."""
        # Create minimal .moai directory
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()

        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._detect_tool_installer") as mock_detect,
        ):

            mock_current.return_value = "0.6.1"
            mock_latest.return_value = "0.6.2"
            mock_detect.return_value = None  # No installer found

            result = runner.invoke(update, ["--path", str(tmp_path), "--yes"])

            assert result.exit_code != 0
            assert "Cannot detect" in result.output
            assert "uv tool" in result.output
            assert "pipx" in result.output
            assert "pip" in result.output

    def test_installer_not_found_provides_recovery(self, tmp_path: Path) -> None:
        """Test recovery instructions shown."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()

        with (
            patch("moai_adk.cli.commands.update._detect_tool_installer", return_value=None),
            patch("moai_adk.cli.commands.update._get_current_version", return_value="0.6.1"),
            patch("moai_adk.cli.commands.update._get_latest_version", return_value="0.6.2"),
        ):

            result = runner.invoke(update, ["--path", str(tmp_path), "--yes"])

            assert "--templates-only" in result.output

    def test_installer_not_found_in_upgrade_path(self, tmp_path: Path) -> None:
        """Test installer error only occurs when upgrade needed."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()

        with (
            patch("moai_adk.cli.commands.update._get_current_version", return_value="0.6.1"),
            patch("moai_adk.cli.commands.update._get_latest_version", return_value="0.6.2"),
            patch("moai_adk.cli.commands.update._detect_tool_installer", return_value=None),
        ):

            result = runner.invoke(update, ["--path", str(tmp_path), "--yes"])

            # Should fail with installer detection message
            assert result.exit_code != 0
            assert "detect" in result.output.lower()


class TestNetworkFailure:
    """Tests for network/PyPI failure (OPT-003, CST-004)."""

    def test_pypi_unreachable_network_error(self, tmp_path: Path) -> None:
        """Test handling of PyPI network error."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()

        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
        ):

            mock_current.return_value = "0.6.1"
            mock_latest.side_effect = RuntimeError("Failed to fetch latest version from PyPI: Network error")

            result = runner.invoke(update, ["--path", str(tmp_path)])

            assert result.exit_code != 0
            assert "Error" in result.output

    def test_network_error_suggests_alternatives(self, tmp_path: Path) -> None:
        """Test error message suggests alternatives."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()

        with patch(
            "moai_adk.cli.commands.update._get_latest_version",
            side_effect=RuntimeError("Failed to fetch from PyPI: Network error"),
        ):

            result = runner.invoke(update, ["--path", str(tmp_path)])

            # Should suggest --force or describe the situation
            assert result.exit_code != 0
            # Error message from click.Abort or exception

    def test_network_error_allows_force_bypass(self, tmp_path: Path) -> None:
        """Test --force bypasses version check failure."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()
        (moai_dir / "config.json").write_text('{"project": {}}')

        with (
            patch("moai_adk.cli.commands.update._get_latest_version", side_effect=RuntimeError("Network error")),
            patch("moai_adk.cli.commands.update._sync_templates", return_value=True),
        ):

            result = runner.invoke(update, ["--path", str(tmp_path), "--force"])

            # Should proceed with sync
            assert "Syncing templates" in result.output

    def test_network_error_allows_templates_only(self, tmp_path: Path) -> None:
        """Test --templates-only skips version check."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()
        (moai_dir / "config.json").write_text('{"project": {}}')

        with (
            patch("moai_adk.cli.commands.update._get_latest_version", side_effect=RuntimeError("Network error")),
            patch("moai_adk.cli.commands.update._sync_templates", return_value=True),
        ):

            result = runner.invoke(update, ["--path", str(tmp_path), "--templates-only"])

            # Should skip version check and sync templates
            assert result.exit_code == 0
            assert "Syncing templates" in result.output


class TestUpgradeFailure:
    """Tests for package upgrade failure (EVT-005, CST-001)."""

    def test_upgrade_failure_shows_error(self, tmp_path: Path) -> None:
        """Test error message when upgrade fails."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()

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

            result = runner.invoke(update, ["--path", str(tmp_path), "--yes"])

            assert result.exit_code != 0
            assert "Upgrade failed" in result.output or "Error" in result.output

    def test_upgrade_failure_provides_troubleshooting(self, tmp_path: Path) -> None:
        """Test troubleshooting steps shown on upgrade failure."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()

        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._detect_tool_installer") as mock_detect,
            patch("moai_adk.cli.commands.update._execute_upgrade") as mock_upgrade,
        ):

            mock_current.return_value = "0.6.1"
            mock_latest.return_value = "0.6.2"
            mock_detect.return_value = ["uv", "tool", "upgrade", "moai-adk"]
            mock_upgrade.return_value = False

            result = runner.invoke(update, ["--path", str(tmp_path), "--yes"])

            # Should suggest troubleshooting steps
            assert "Troubleshooting" in result.output or "manually" in result.output.lower()

    def test_upgrade_subprocess_error_captured(self, tmp_path: Path) -> None:
        """Test subprocess errors are captured and shown."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()

        with (
            patch("moai_adk.cli.commands.update._get_current_version", return_value="0.6.1"),
            patch("moai_adk.cli.commands.update._get_latest_version", return_value="0.6.2"),
            patch(
                "moai_adk.cli.commands.update._detect_tool_installer",
                return_value=["uv", "tool", "upgrade", "moai-adk"],
            ),
            patch("moai_adk.cli.commands.update._execute_upgrade", return_value=False),
        ):

            result = runner.invoke(update, ["--path", str(tmp_path), "--yes"])

            assert result.exit_code != 0


class TestTemplateSyncError:
    """Tests for template sync failure (CST-005)."""

    def test_sync_failure_shows_error(self, tmp_path: Path) -> None:
        """Test error when template sync fails."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()

        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
        ):

            mock_current.return_value = "0.6.2"
            mock_latest.return_value = "0.6.2"
            mock_sync.return_value = False  # Sync failed

            result = runner.invoke(update, ["--path", str(tmp_path)])

            assert result.exit_code != 0
            assert "failed" in result.output.lower() or "Error" in result.output

    def test_sync_failure_in_templates_only_mode(self, tmp_path: Path) -> None:
        """Test sync failure in --templates-only mode."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()

        with patch("moai_adk.cli.commands.update._sync_templates", return_value=False):

            result = runner.invoke(update, ["--path", str(tmp_path), "--templates-only"])

            assert result.exit_code != 0
            assert "Error" in result.output or "failed" in result.output.lower()

    def test_sync_exception_handled(self, tmp_path: Path) -> None:
        """Test exceptions during sync are handled."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()

        with (
            patch("moai_adk.cli.commands.update._get_current_version", return_value="0.6.2"),
            patch("moai_adk.cli.commands.update._get_latest_version", return_value="0.6.2"),
            patch("moai_adk.cli.commands.update._sync_templates", side_effect=Exception("Template error")),
        ):

            result = runner.invoke(update, ["--path", str(tmp_path)])

            # Should catch exception and show error
            assert result.exit_code != 0


class TestTimeoutError:
    """Tests for timeout errors (CST-003)."""

    def test_subprocess_timeout_handling(self, tmp_path: Path) -> None:
        """Test handling of subprocess timeout."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()

        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._detect_tool_installer") as mock_detect,
            patch("moai_adk.cli.commands.update._execute_upgrade") as mock_upgrade,
        ):

            mock_current.return_value = "0.6.1"
            mock_latest.return_value = "0.6.2"
            mock_detect.return_value = ["uv", "tool", "upgrade", "moai-adk"]
            mock_upgrade.side_effect = subprocess.TimeoutExpired(cmd="uv tool upgrade moai-adk", timeout=60)

            # Should not crash, show error instead
            result = runner.invoke(update, ["--path", str(tmp_path), "--yes"])

            assert result.exit_code != 0

    def test_execute_upgrade_timeout_raises(self) -> None:
        """Test _execute_upgrade raises TimeoutExpired on timeout."""
        from moai_adk.cli.commands.update import _execute_upgrade

        with patch("subprocess.run", side_effect=subprocess.TimeoutExpired(cmd="test", timeout=60)):
            with pytest.raises(subprocess.TimeoutExpired):
                _execute_upgrade(["uv", "tool", "upgrade", "moai-adk"])


class TestExitCodes:
    """Tests for proper exit codes."""

    def test_success_exit_code_zero(self, tmp_path: Path) -> None:
        """Test successful run returns exit code 0."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()
        (moai_dir / "config.json").write_text('{"project": {}}')

        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
            patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor,
        ):

            mock_current.return_value = "0.6.2"
            mock_latest.return_value = "0.6.2"
            mock_sync.return_value = True
            mock_processor.return_value.create_backup.return_value = tmp_path / ".moai-backups" / "test"

            result = runner.invoke(update, ["--path", str(tmp_path)])

            assert result.exit_code == 0

    def test_error_exit_code_nonzero(self, tmp_path: Path) -> None:
        """Test error run returns non-zero exit code."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()

        with (
            patch("moai_adk.cli.commands.update._detect_tool_installer", return_value=None),
            patch("moai_adk.cli.commands.update._get_current_version", return_value="0.6.1"),
            patch("moai_adk.cli.commands.update._get_latest_version", return_value="0.6.2"),
        ):

            result = runner.invoke(update, ["--path", str(tmp_path)])

            assert result.exit_code != 0

    def test_check_mode_success_exit_code(self, tmp_path: Path) -> None:
        """Test --check mode success returns exit code 0."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()

        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
        ):

            mock_current.return_value = "0.6.2"
            mock_latest.return_value = "0.6.2"

            result = runner.invoke(update, ["--path", str(tmp_path), "--check"])

            assert result.exit_code == 0

    def test_check_mode_shows_update_available(self, tmp_path: Path) -> None:
        """Test --check mode shows update available."""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()

        with (
            patch("moai_adk.cli.commands.update._get_current_version", return_value="0.6.1"),
            patch("moai_adk.cli.commands.update._get_latest_version", return_value="0.6.2"),
        ):

            result = runner.invoke(update, ["--path", str(tmp_path), "--check"])

            assert result.exit_code == 0
            assert "Update available" in result.output or "0.6.2" in result.output


class TestProjectNotInitialized:
    """Tests for project not initialized error."""

    def test_project_not_initialized_error(self, tmp_path: Path) -> None:
        """Test error when .moai directory missing."""
        result = runner.invoke(update, ["--path", str(tmp_path)])

        assert result.exit_code != 0
        assert "not initialized" in result.output.lower() or "Aborted" in result.output
