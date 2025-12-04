"""Unit tests for Phase 3: CLI Options Enhancement

Tests for --templates-only, --yes, --check, and --force options.
RED Phase: These tests should initially fail.
"""

import json
from unittest.mock import patch

import pytest
from click.testing import CliRunner

from moai_adk.cli.commands.update import (
    update,
)


@pytest.fixture
def runner():
    """Create Click test runner."""
    return CliRunner()


@pytest.fixture
def mock_project_dir(tmp_path):
    """Create a mock project directory with .moai folder."""
    moai_dir = tmp_path / ".moai"
    moai_dir.mkdir()
    config_file = moai_dir / "config.json"
    config_file.write_text(json.dumps({"project": {"name": "test-project"}}))
    return tmp_path


class TestTemplatesOnlyOption:
    """Tests for --templates-only flag."""

    def test_templates_only_skips_upgrade(self, runner, mock_project_dir):
        """Test --templates-only skips package upgrade."""
        with (
            patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
        ):
            mock_sync.return_value = True
            mock_current.return_value = "0.6.1"
            mock_latest.return_value = "0.6.2"

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--templates-only"])

            assert result.exit_code == 0
            mock_sync.assert_called_once()
            # Should not check versions when --templates-only is used
            mock_current.assert_not_called()
            mock_latest.assert_not_called()

    def test_templates_only_no_version_check(self, runner, mock_project_dir):
        """Test --templates-only doesn't check versions."""
        with (
            patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
        ):
            mock_sync.return_value = True

            runner.invoke(update, ["--path", str(mock_project_dir), "--templates-only"])

            # Should not check versions
            mock_current.assert_not_called()

    def test_templates_only_with_force(self, runner, mock_project_dir):
        """Test --templates-only with --force flag."""
        with patch("moai_adk.cli.commands.update._sync_templates") as mock_sync:
            mock_sync.return_value = True

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--templates-only", "--force"])

            assert result.exit_code == 0
            # Force should be passed to _sync_templates
            mock_sync.assert_called_once()
            call_args = mock_sync.call_args
            assert call_args[0][1] is True  # force parameter

    def test_templates_only_sync_failure(self, runner, mock_project_dir):
        """Test --templates-only handles sync failure."""
        with patch("moai_adk.cli.commands.update._sync_templates") as mock_sync:
            mock_sync.return_value = False

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--templates-only"])

            assert result.exit_code != 0
            assert "Template sync failed" in result.output or "Error" in result.output

    def test_templates_only_message(self, runner, mock_project_dir):
        """Test --templates-only shows appropriate message."""
        with patch("moai_adk.cli.commands.update._sync_templates") as mock_sync:
            mock_sync.return_value = True

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--templates-only"])

            assert result.exit_code == 0
            # Should not show upgrade messages
            assert "Upgrading" not in result.output or "upgrade" not in result.output.lower()


class TestCheckOption:
    """Tests for --check flag."""

    def test_check_displays_versions(self, runner, mock_project_dir):
        """Test --check displays current and latest versions."""
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
        ):
            mock_current.return_value = "0.6.1"
            mock_latest.return_value = "0.6.2"

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--check"])

            assert result.exit_code == 0
            assert "0.6.1" in result.output
            assert "0.6.2" in result.output
            assert "Current" in result.output or "current" in result.output
            assert "Latest" in result.output or "latest" in result.output

    def test_check_no_changes_made(self, runner, mock_project_dir):
        """Test --check makes no changes."""
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
            patch("moai_adk.cli.commands.update._execute_upgrade") as mock_upgrade,
        ):
            mock_current.return_value = "0.6.1"
            mock_latest.return_value = "0.6.2"

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--check"])

            # Should not execute upgrade or sync
            mock_sync.assert_not_called()
            mock_upgrade.assert_not_called()
            assert result.exit_code == 0

    def test_check_shows_update_available(self, runner, mock_project_dir):
        """Test --check shows update available message."""
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
        ):
            mock_current.return_value = "0.6.1"
            mock_latest.return_value = "0.6.2"

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--check"])

            assert result.exit_code == 0
            assert (
                "Update available" in result.output
                or "update available" in result.output.lower()
                or "0.6.1" in result.output
                and "0.6.2" in result.output
            )

    def test_check_shows_already_updated(self, runner, mock_project_dir):
        """Test --check shows already up to date message."""
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
        ):
            mock_current.return_value = "0.6.2"
            mock_latest.return_value = "0.6.2"

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--check"])

            assert result.exit_code == 0
            assert "up to date" in result.output.lower() or "Already" in result.output or "already" in result.output

    def test_check_shows_dev_version(self, runner, mock_project_dir):
        """Test --check shows dev version message when current > latest."""
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
        ):
            mock_current.return_value = "0.7.0"
            mock_latest.return_value = "0.6.2"

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--check"])

            assert result.exit_code == 0
            assert "0.7.0" in result.output
            assert "0.6.2" in result.output

    def test_check_handles_network_error(self, runner, mock_project_dir):
        """Test --check handles network error gracefully."""
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
        ):
            mock_current.return_value = "0.6.1"
            mock_latest.side_effect = RuntimeError("Network error")

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--check"])

            # Should handle error gracefully
            assert result.exit_code != 0
            assert "Error" in result.output or "error" in result.output


class TestYesOption:
    """Tests for --yes flag."""

    def test_yes_auto_confirms_upgrade(self, runner, mock_project_dir):
        """Test --yes auto-confirms upgrade prompt."""
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

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--yes"])

            # Should execute upgrade without prompting
            assert result.exit_code == 0
            mock_upgrade.assert_called_once()

    def test_yes_no_confirmation_prompt(self, runner, mock_project_dir):
        """Test --yes doesn't show confirmation prompt."""
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

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--yes"])

            # Should not contain confirmation prompt keywords
            assert "Continue?" not in result.output
            assert "Proceed?" not in result.output
            assert "[y/N]" not in result.output

    def test_yes_with_no_upgrade_needed(self, runner, mock_project_dir):
        """Test --yes when no upgrade needed (goes straight to sync)."""
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
        ):
            mock_current.return_value = "0.6.2"
            mock_latest.return_value = "0.6.2"
            mock_sync.return_value = True

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--yes"])

            assert result.exit_code == 0
            mock_sync.assert_called_once()


class TestForceOption:
    """Tests for --force flag."""

    def test_force_skips_backup(self, runner, mock_project_dir):
        """Test --force skips backup creation."""
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
            patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor,
        ):
            mock_current.return_value = "0.6.2"
            mock_latest.return_value = "0.6.2"
            mock_sync.return_value = True

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--force"])

            assert result.exit_code == 0
            # Should not create backup when --force
            mock_processor.assert_not_called()

    def test_force_passed_to_sync_templates(self, runner, mock_project_dir):
        """Test --force is passed to _sync_templates function."""
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
        ):
            mock_current.return_value = "0.6.2"
            mock_latest.return_value = "0.6.2"
            mock_sync.return_value = True

            runner.invoke(update, ["--path", str(mock_project_dir), "--force"])

            # Force should be passed to _sync_templates
            call_args = mock_sync.call_args
            assert call_args[0][1] is True  # force parameter


class TestCombinedOptions:
    """Tests for combined options."""

    def test_yes_and_force_together(self, runner, mock_project_dir):
        """Test --yes and --force work together."""
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

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--yes", "--force"])

            assert result.exit_code == 0
            mock_upgrade.assert_called_once()

    def test_check_takes_priority_over_templates_only(self, runner, mock_project_dir):
        """Test --check takes priority over --templates-only."""
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
        ):
            mock_current.return_value = "0.6.1"
            mock_latest.return_value = "0.6.2"

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--check", "--templates-only"])

            # Check should take priority - show version info but don't sync
            assert "0.6.1" in result.output or "Current" in result.output
            mock_sync.assert_not_called()

    def test_check_with_yes_ignored(self, runner, mock_project_dir):
        """Test --yes is ignored when --check is used."""
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._execute_upgrade") as mock_upgrade,
        ):
            mock_current.return_value = "0.6.1"
            mock_latest.return_value = "0.6.2"

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--check", "--yes"])

            # Should only check, not upgrade
            assert result.exit_code == 0
            mock_upgrade.assert_not_called()

    def test_templates_only_with_yes(self, runner, mock_project_dir):
        """Test --templates-only with --yes (no prompts expected)."""
        with patch("moai_adk.cli.commands.update._sync_templates") as mock_sync:
            mock_sync.return_value = True

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--templates-only", "--yes"])

            assert result.exit_code == 0
            mock_sync.assert_called_once()


class TestErrorHandling:
    """Tests for error handling scenarios."""

    def test_templates_only_on_uninitialized_project(self, runner, tmp_path):
        """Test --templates-only on project without .moai folder."""
        result = runner.invoke(update, ["--path", str(tmp_path), "--templates-only"])

        # Should detect project is not initialized
        assert result.exit_code != 0

    def test_check_on_uninitialized_project(self, runner, tmp_path):
        """Test --check on project without .moai folder."""
        result = runner.invoke(update, ["--path", str(tmp_path), "--check"])

        # Should detect project is not initialized
        assert result.exit_code != 0

    def test_yes_upgrade_failure(self, runner, mock_project_dir):
        """Test --yes handles upgrade failure."""
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._detect_tool_installer") as mock_detect,
            patch("moai_adk.cli.commands.update._execute_upgrade") as mock_upgrade,
        ):
            mock_current.return_value = "0.6.1"
            mock_latest.return_value = "0.6.2"
            mock_detect.return_value = ["uv", "tool", "upgrade", "moai-adk"]
            mock_upgrade.return_value = False  # Upgrade fails

            result = runner.invoke(update, ["--path", str(mock_project_dir), "--yes"])

            assert result.exit_code != 0
            assert "Error" in result.output or "failed" in result.output.lower()
