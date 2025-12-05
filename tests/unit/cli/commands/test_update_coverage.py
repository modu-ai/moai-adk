"""Comprehensive coverage tests for update command module.

These tests focus on uncovered code paths in src/moai_adk/cli/commands/update.py
with particular attention to version checking, migration, backup operations,
and error handling. Uses @patch for mocking subprocess, file operations, and network calls.

Target Coverage: 70%+
Test Pattern: AAA (Arrange-Act-Assert)
"""

import subprocess
from datetime import datetime
from pathlib import Path
from unittest.mock import MagicMock, patch, call, mock_open
import pytest
import yaml
from packaging import version

from moai_adk.cli.commands.update import (
    _is_installed_via_uv_tool,
    _is_installed_via_pipx,
    _is_installed_via_pip,
    _detect_tool_installer,
    _get_current_version,
    _get_latest_version,
    _compare_versions,
    _get_package_config_version,
    _get_project_config_version,
    _ask_merge_strategy,
    _generate_manual_merge_guide,
    _migrate_legacy_logs,
    _detect_stale_cache,
    _clear_uv_package_cache,
    _execute_upgrade_with_retry,
    UpdateError,
    InstallerNotFoundError,
    NetworkError,
    UpgradeError,
    TemplateSyncError,
)


# ============================================================================
# Test Installer Detection Functions
# ============================================================================


class TestInstallerDetection:
    """Test suite for installer detection functions."""

    @patch("subprocess.run")
    def test_is_installed_via_uv_tool_success(self, mock_run):
        """Test successful detection of uv tool installation."""
        # Arrange
        mock_run.return_value = MagicMock(
            returncode=0,
            stdout="moai-adk 0.8.3",
        )

        # Act
        result = _is_installed_via_uv_tool()

        # Assert
        assert result is True
        mock_run.assert_called_once()
        assert mock_run.call_args[0][0] == ["uv", "tool", "list"]

    @patch("subprocess.run")
    def test_is_installed_via_uv_tool_not_found(self, mock_run):
        """Test uv tool detection when moai-adk not in output."""
        # Arrange
        mock_run.return_value = MagicMock(
            returncode=0,
            stdout="other-package 1.0.0",
        )

        # Act
        result = _is_installed_via_uv_tool()

        # Assert
        assert result is False

    @patch("subprocess.run")
    def test_is_installed_via_uv_tool_timeout(self, mock_run):
        """Test uv tool detection with timeout exception."""
        # Arrange
        mock_run.side_effect = subprocess.TimeoutExpired("uv", 5)

        # Act
        result = _is_installed_via_uv_tool()

        # Assert
        assert result is False

    @patch("subprocess.run")
    def test_is_installed_via_uv_tool_not_found_error(self, mock_run):
        """Test uv tool detection when command not found."""
        # Arrange
        mock_run.side_effect = FileNotFoundError()

        # Act
        result = _is_installed_via_uv_tool()

        # Assert
        assert result is False

    @patch("subprocess.run")
    def test_is_installed_via_pipx_success(self, mock_run):
        """Test successful detection of pipx installation."""
        # Arrange
        mock_run.return_value = MagicMock(
            returncode=0,
            stdout="moai-adk 0.8.3",
        )

        # Act
        result = _is_installed_via_pipx()

        # Assert
        assert result is True
        mock_run.assert_called_once()
        assert mock_run.call_args[0][0] == ["pipx", "list"]

    @patch("subprocess.run")
    def test_is_installed_via_pipx_failure(self, mock_run):
        """Test pipx detection failure."""
        # Arrange
        mock_run.return_value = MagicMock(returncode=1, stdout="")

        # Act
        result = _is_installed_via_pipx()

        # Assert
        assert result is False

    @patch("subprocess.run")
    def test_is_installed_via_pip_success(self, mock_run):
        """Test successful detection of pip installation."""
        # Arrange
        mock_run.return_value = MagicMock(returncode=0)

        # Act
        result = _is_installed_via_pip()

        # Assert
        assert result is True
        mock_run.assert_called_once()
        assert mock_run.call_args[0][0] == ["pip", "show", "moai-adk"]

    @patch("subprocess.run")
    def test_is_installed_via_pip_not_found(self, mock_run):
        """Test pip detection when package not found."""
        # Arrange
        mock_run.return_value = MagicMock(returncode=1)

        # Act
        result = _is_installed_via_pip()

        # Assert
        assert result is False

    @patch("moai_adk.cli.commands.update._is_installed_via_uv_tool")
    @patch("moai_adk.cli.commands.update._is_installed_via_pipx")
    @patch("moai_adk.cli.commands.update._is_installed_via_pip")
    def test_detect_tool_installer_uv_priority(self, mock_pip, mock_pipx, mock_uv):
        """Test tool detection with uv tool priority."""
        # Arrange
        mock_uv.return_value = True
        mock_pipx.return_value = True
        mock_pip.return_value = True

        # Act
        result = _detect_tool_installer()

        # Assert
        assert result == ["uv", "tool", "upgrade", "moai-adk"]
        mock_uv.assert_called_once()

    @patch("moai_adk.cli.commands.update._is_installed_via_uv_tool")
    @patch("moai_adk.cli.commands.update._is_installed_via_pipx")
    @patch("moai_adk.cli.commands.update._is_installed_via_pip")
    def test_detect_tool_installer_pipx_fallback(self, mock_pip, mock_pipx, mock_uv):
        """Test tool detection with pipx fallback."""
        # Arrange
        mock_uv.return_value = False
        mock_pipx.return_value = True
        mock_pip.return_value = True

        # Act
        result = _detect_tool_installer()

        # Assert
        assert result == ["pipx", "upgrade", "moai-adk"]

    @patch("moai_adk.cli.commands.update._is_installed_via_uv_tool")
    @patch("moai_adk.cli.commands.update._is_installed_via_pipx")
    @patch("moai_adk.cli.commands.update._is_installed_via_pip")
    def test_detect_tool_installer_pip_fallback(self, mock_pip, mock_pipx, mock_uv):
        """Test tool detection with pip fallback."""
        # Arrange
        mock_uv.return_value = False
        mock_pipx.return_value = False
        mock_pip.return_value = True

        # Act
        result = _detect_tool_installer()

        # Assert
        assert result == ["pip", "install", "--upgrade", "moai-adk"]

    @patch("moai_adk.cli.commands.update._is_installed_via_uv_tool")
    @patch("moai_adk.cli.commands.update._is_installed_via_pipx")
    @patch("moai_adk.cli.commands.update._is_installed_via_pip")
    def test_detect_tool_installer_not_found(self, mock_pip, mock_pipx, mock_uv):
        """Test tool detection when no installer found."""
        # Arrange
        mock_uv.return_value = False
        mock_pipx.return_value = False
        mock_pip.return_value = False

        # Act
        result = _detect_tool_installer()

        # Assert
        assert result is None


# ============================================================================
# Test Version Functions
# ============================================================================


class TestVersionFunctions:
    """Test suite for version checking functions."""

    @patch("moai_adk.cli.commands.update.__version__", "0.8.3")
    def test_get_current_version(self):
        """Test getting current version."""
        # Act
        result = _get_current_version()

        # Assert
        assert result == "0.8.3"

    @patch("urllib.request.urlopen")
    def test_get_latest_version_success(self, mock_urlopen):
        """Test successful fetch of latest version from PyPI."""
        # Arrange
        mock_response = MagicMock()
        mock_response.__enter__ = MagicMock(return_value=mock_response)
        mock_response.__exit__ = MagicMock(return_value=False)
        import json

        mock_response.read.return_value = json.dumps({"info": {"version": "0.9.0"}}).encode("utf-8")
        mock_urlopen.return_value = mock_response

        # Act
        result = _get_latest_version()

        # Assert
        assert result == "0.9.0"
        mock_urlopen.assert_called_once()

    @patch("urllib.request.urlopen")
    def test_get_latest_version_network_error(self, mock_urlopen):
        """Test network error handling when fetching version."""
        # Arrange
        import urllib.error

        mock_urlopen.side_effect = urllib.error.URLError("Connection failed")

        # Act & Assert
        with pytest.raises(RuntimeError):
            _get_latest_version()

    @patch("urllib.request.urlopen")
    def test_get_latest_version_invalid_json(self, mock_urlopen):
        """Test invalid JSON response handling."""
        # Arrange
        mock_response = MagicMock()
        mock_response.__enter__ = MagicMock(return_value=mock_response)
        mock_response.__exit__ = MagicMock(return_value=False)
        mock_response.read.return_value = b"invalid json"
        mock_urlopen.return_value = mock_response

        # Act & Assert
        with pytest.raises(RuntimeError):
            _get_latest_version()

    def test_compare_versions_upgrade_needed(self):
        """Test version comparison when upgrade needed."""
        # Act
        result = _compare_versions("0.8.3", "0.9.0")

        # Assert
        assert result == -1

    def test_compare_versions_up_to_date(self):
        """Test version comparison when up to date."""
        # Act
        result = _compare_versions("0.9.0", "0.9.0")

        # Assert
        assert result == 0

    def test_compare_versions_newer_installed(self):
        """Test version comparison when newer version installed."""
        # Act
        result = _compare_versions("0.10.0", "0.9.0")

        # Assert
        assert result == 1

    def test_compare_versions_prerelease(self):
        """Test version comparison with pre-release versions."""
        # Act
        result = _compare_versions("0.9.0a1", "0.9.0")

        # Assert
        assert result == -1

    @patch("moai_adk.cli.commands.update.__version__", "0.8.0")
    def test_get_package_config_version(self):
        """Test getting package config version."""
        # Act
        result = _get_package_config_version()

        # Assert
        assert result == "0.8.0"


# ============================================================================
# Test Config Version Functions
# ============================================================================


class TestConfigVersionFunctions:
    """Test suite for config version detection."""

    def test_get_project_config_version_no_config(self):
        """Test project config version when config.yaml missing."""
        # Arrange
        with patch("pathlib.Path.exists", return_value=False):
            # Act
            result = _get_project_config_version(Path("/mock/project"))

            # Assert
            assert result == "0.0.0"

    def test_get_project_config_version_with_template_version(self):
        """Test project config version with template_version field."""
        # Arrange
        config_data = {
            "project": {"template_version": "0.8.1"},
            "moai": {"version": "0.8.0"},
        }

        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text") as mock_read:
                mock_read.return_value = yaml.dump(config_data)

                # Act
                result = _get_project_config_version(Path("/mock/project"))

                # Assert
                assert result == "0.8.1"

    def test_get_project_config_version_fallback_moai_version(self):
        """Test project config version with moai version fallback."""
        # Arrange
        config_data = {"project": {}, "moai": {"version": "0.8.0"}}

        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text") as mock_read:
                mock_read.return_value = yaml.dump(config_data)

                # Act
                result = _get_project_config_version(Path("/mock/project"))

                # Assert
                assert result == "0.8.0"

    def test_get_project_config_version_placeholder_value(self):
        """Test project config version with unsubstituted placeholder."""
        # Arrange
        config_data = {"project": {"template_version": "{{VERSION}}"}, "moai": {}}

        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text") as mock_read:
                mock_read.return_value = yaml.dump(config_data)

                # Act
                result = _get_project_config_version(Path("/mock/project"))

                # Assert
                assert result == "0.0.0"

    def test_get_project_config_version_invalid_yaml(self):
        """Test project config version with invalid YAML."""
        # Arrange
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text") as mock_read:
                mock_read.return_value = "invalid: yaml: content:"

                # Act & Assert
                with pytest.raises(ValueError):
                    _get_project_config_version(Path("/mock/project"))


# ============================================================================
# Test Merge Strategy and Guide Generation
# ============================================================================


class TestMergeStrategyFunctions:
    """Test suite for merge strategy and guide generation."""

    @patch("click.prompt")
    def test_ask_merge_strategy_auto(self, mock_prompt):
        """Test merge strategy selection for auto merge."""
        # Arrange
        mock_prompt.return_value = "1"

        # Act
        result = _ask_merge_strategy()

        # Assert
        assert result == "auto"

    @patch("click.prompt")
    def test_ask_merge_strategy_manual(self, mock_prompt):
        """Test merge strategy selection for manual merge."""
        # Arrange
        mock_prompt.return_value = "2"

        # Act
        result = _ask_merge_strategy()

        # Assert
        assert result == "manual"

    def test_ask_merge_strategy_yes_flag(self):
        """Test merge strategy with yes flag."""
        # Act
        result = _ask_merge_strategy(yes=True)

        # Assert
        assert result == "auto"

    @patch("pathlib.Path.write_text")
    @patch("pathlib.Path.rglob")
    @patch("pathlib.Path.mkdir")
    @patch("pathlib.Path.exists")
    def test_generate_manual_merge_guide(self, mock_exists, mock_mkdir, mock_rglob, mock_write):
        """Test manual merge guide generation."""
        # Arrange
        mock_exists.return_value = True
        mock_mkdir.return_value = None
        mock_rglob.return_value = []
        mock_write.return_value = None

        # Use real paths for relative_to to work
        backup_path = Path("/tmp/backup")
        template_path = Path("/tmp/template")
        project_path = Path("/tmp")

        # Act
        result = _generate_manual_merge_guide(backup_path, template_path, project_path)

        # Assert
        assert result is not None
        mock_write.assert_called_once()
        written_content = mock_write.call_args[0][0]
        assert "Merge Guide" in written_content
        assert "Manual Merge Mode" in written_content


# ============================================================================
# Test Legacy Log Migration
# ============================================================================


class TestLegacyLogMigration:
    """Test suite for legacy log migration."""

    @patch("pathlib.Path.exists")
    def test_migrate_legacy_logs_no_legacy_files(self, mock_exists):
        """Test migration when no legacy files exist."""
        # Arrange
        mock_exists.return_value = False

        with patch("pathlib.Path.mkdir"):
            # Act
            result = _migrate_legacy_logs(Path("/mock/project"), dry_run=False)

            # Assert
            assert result is True

    @patch("pathlib.Path.exists")
    @patch("pathlib.Path.rglob")
    def test_migrate_legacy_logs_dry_run(self, mock_rglob, mock_exists):
        """Test legacy log migration in dry-run mode."""
        # Arrange
        mock_exists.side_effect = [
            True,  # legacy_memory.exists()
            False,  # legacy_error_logs.exists()
            False,  # legacy_reports.exists()
            False,  # session_file.exists()
        ]
        mock_rglob.return_value = []

        with patch("pathlib.Path.mkdir"):
            # Act
            result = _migrate_legacy_logs(Path("/mock/project"), dry_run=True)

            # Assert
            assert result is True

    @patch("shutil.copy2")
    @patch("pathlib.Path.exists")
    def test_migrate_legacy_logs_actual_migration(self, mock_exists, mock_copy):
        """Test actual legacy log migration."""
        # Arrange
        mock_exists.side_effect = [
            True,  # legacy_memory.exists()
            False,  # legacy_error_logs.exists()
            False,  # legacy_reports.exists()
            True,  # session_file.exists()
            False,  # target_file.exists()
        ]

        with patch("pathlib.Path.mkdir"):
            with patch("pathlib.Path.rglob", return_value=[]):
                with patch("shutil.copystat"):
                    # Act
                    result = _migrate_legacy_logs(Path("/mock/project"), dry_run=False)

                    # Assert
                    assert result is True

    @patch("pathlib.Path.exists")
    def test_migrate_legacy_logs_exception_handling(self, mock_exists):
        """Test exception handling during migration."""
        # Arrange
        mock_exists.side_effect = Exception("Permission denied")

        # Act
        result = _migrate_legacy_logs(Path("/mock/project"), dry_run=False)

        # Assert
        assert result is False


# ============================================================================
# Test Cache and Upgrade Functions
# ============================================================================


class TestCacheAndUpgradeFunctions:
    """Test suite for cache and upgrade operations."""

    def test_detect_stale_cache_true(self):
        """Test stale cache detection when true."""
        # Arrange
        upgrade_output = "Nothing to upgrade"
        current_version = "0.8.3"
        latest_version = "0.9.0"

        # Act
        result = _detect_stale_cache(upgrade_output, current_version, latest_version)

        # Assert
        assert result is True

    def test_detect_stale_cache_false_upgraded(self):
        """Test stale cache detection when upgraded."""
        # Arrange
        upgrade_output = "Updated moai-adk from 0.8.3 to 0.9.0"
        current_version = "0.8.3"
        latest_version = "0.9.0"

        # Act
        result = _detect_stale_cache(upgrade_output, current_version, latest_version)

        # Assert
        assert result is False

    def test_detect_stale_cache_false_up_to_date(self):
        """Test stale cache detection when up to date."""
        # Arrange
        upgrade_output = "Nothing to upgrade"
        current_version = "0.9.0"
        latest_version = "0.9.0"

        # Act
        result = _detect_stale_cache(upgrade_output, current_version, latest_version)

        # Assert
        assert result is False

    def test_detect_stale_cache_invalid_version(self):
        """Test stale cache with invalid version."""
        # Arrange
        upgrade_output = "Nothing to upgrade"
        current_version = "invalid"
        latest_version = "0.9.0"

        # Act
        result = _detect_stale_cache(upgrade_output, current_version, latest_version)

        # Assert
        assert result is False

    @patch("subprocess.run")
    def test_clear_uv_package_cache_success(self, mock_run):
        """Test successful cache clearing."""
        # Arrange
        mock_run.return_value = MagicMock(returncode=0)

        # Act
        result = _clear_uv_package_cache("moai-adk")

        # Assert
        assert result is True
        mock_run.assert_called_once()

    @patch("subprocess.run")
    def test_clear_uv_package_cache_failure(self, mock_run):
        """Test cache clearing failure."""
        # Arrange
        mock_run.return_value = MagicMock(returncode=1, stderr="Error")

        # Act
        result = _clear_uv_package_cache("moai-adk")

        # Assert
        assert result is False

    @patch("subprocess.run")
    def test_clear_uv_package_cache_timeout(self, mock_run):
        """Test cache clearing timeout."""
        # Arrange
        mock_run.side_effect = subprocess.TimeoutExpired("uv", 10)

        # Act
        result = _clear_uv_package_cache("moai-adk")

        # Assert
        assert result is False

    @patch("subprocess.run")
    def test_clear_uv_package_cache_not_found(self, mock_run):
        """Test cache clearing when uv not found."""
        # Arrange
        mock_run.side_effect = FileNotFoundError()

        # Act
        result = _clear_uv_package_cache("moai-adk")

        # Assert
        assert result is False


# ============================================================================
# Test Upgrade with Retry
# ============================================================================


class TestUpgradeWithRetry:
    """Test suite for upgrade with retry logic."""

    @patch("moai_adk.cli.commands.update._detect_stale_cache")
    @patch("moai_adk.cli.commands.update._clear_uv_package_cache")
    @patch("moai_adk.cli.commands.update._get_latest_version")
    @patch("subprocess.run")
    def test_execute_upgrade_with_retry_success(self, mock_run, mock_latest, mock_clear, mock_detect):
        """Test successful upgrade without retry."""
        # Arrange
        mock_run.return_value = MagicMock(
            returncode=0,
            stdout="Updated moai-adk",
        )
        mock_latest.return_value = "0.9.0"
        mock_detect.return_value = False

        # Act
        result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])

        # Assert
        assert result is True

    @patch("moai_adk.cli.commands.update._detect_stale_cache")
    @patch("moai_adk.cli.commands.update._clear_uv_package_cache")
    @patch("moai_adk.cli.commands.update._get_latest_version")
    @patch("moai_adk.cli.commands.update._get_current_version")
    @patch("subprocess.run")
    def test_execute_upgrade_with_retry_stale_cache(self, mock_run, mock_current, mock_latest, mock_clear, mock_detect):
        """Test upgrade with stale cache detection and retry."""
        # Arrange
        mock_run.side_effect = [
            MagicMock(returncode=0, stdout="Nothing to upgrade"),
            MagicMock(returncode=0, stdout="Updated moai-adk"),
        ]
        mock_current.return_value = "0.8.3"
        mock_latest.return_value = "0.9.0"
        mock_detect.return_value = True
        mock_clear.return_value = True

        # Act
        result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])

        # Assert
        assert result is True
        assert mock_run.call_count == 2

    @patch("moai_adk.cli.commands.update._detect_stale_cache")
    @patch("moai_adk.cli.commands.update._clear_uv_package_cache")
    @patch("moai_adk.cli.commands.update._get_latest_version")
    @patch("moai_adk.cli.commands.update._get_current_version")
    @patch("subprocess.run")
    def test_execute_upgrade_with_retry_cache_clear_fail(
        self, mock_run, mock_current, mock_latest, mock_clear, mock_detect
    ):
        """Test upgrade when cache clear fails."""
        # Arrange
        mock_run.return_value = MagicMock(
            returncode=0,
            stdout="Nothing to upgrade",
        )
        mock_current.return_value = "0.8.3"
        mock_latest.return_value = "0.9.0"
        mock_detect.return_value = True
        mock_clear.return_value = False

        # Act
        result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])

        # Assert
        assert result is False

    @patch("moai_adk.cli.commands.update._detect_stale_cache")
    @patch("subprocess.run")
    def test_execute_upgrade_with_retry_subprocess_error(self, mock_run, mock_detect):
        """Test upgrade with subprocess error."""
        # Arrange
        mock_run.side_effect = subprocess.CalledProcessError(1, "uv")
        mock_detect.return_value = False

        # Act
        result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])

        # Assert
        assert result is False


# ============================================================================
# Test Custom Exceptions
# ============================================================================


class TestCustomExceptions:
    """Test suite for custom exception classes."""

    def test_update_error_exception(self):
        """Test UpdateError exception."""
        # Act & Assert
        with pytest.raises(UpdateError):
            raise UpdateError("Test error")

    def test_installer_not_found_error_exception(self):
        """Test InstallerNotFoundError exception."""
        # Act & Assert
        with pytest.raises(InstallerNotFoundError):
            raise InstallerNotFoundError("No installer found")

    def test_network_error_exception(self):
        """Test NetworkError exception."""
        # Act & Assert
        with pytest.raises(NetworkError):
            raise NetworkError("Network failure")

    def test_upgrade_error_exception(self):
        """Test UpgradeError exception."""
        # Act & Assert
        with pytest.raises(UpgradeError):
            raise UpgradeError("Upgrade failed")

    def test_template_sync_error_exception(self):
        """Test TemplateSyncError exception."""
        # Act & Assert
        with pytest.raises(TemplateSyncError):
            raise TemplateSyncError("Template sync failed")
