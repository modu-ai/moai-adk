"""
Final comprehensive tests for update command - targeting 55%+ coverage.

Focus areas:
- Version comparison functions
- Installer detection functions
- Config version getters
- Simple helper functions
- Exception classes

All tests use @patch to mock subprocess, file operations, and network calls.
"""

import json
import subprocess
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.cli.commands.update import (
    PIP_COMMAND,
    PIPX_COMMAND,
    TOOL_DETECTION_TIMEOUT,
    UV_TOOL_COMMAND,
    InstallerNotFoundError,
    NetworkError,
    TemplateSyncError,
    UpdateError,
    UpgradeError,
    _compare_versions,
    _detect_tool_installer,
    _get_current_version,
    _get_latest_version,
    _get_package_config_version,
    _get_project_config_version,
    _is_installed_via_pip,
    _is_installed_via_pipx,
    _is_installed_via_uv_tool,
)

# ============================================================================
# Test Exception Classes
# ============================================================================


class TestExceptionClasses:
    """Test custom exception classes."""

    def test_update_error_creation(self):
        """Test creating UpdateError exception."""
        # Arrange & Act
        error = UpdateError("Update failed")

        # Assert
        assert str(error) == "Update failed"
        assert isinstance(error, Exception)

    def test_installer_not_found_error(self):
        """Test InstallerNotFoundError."""
        # Arrange & Act
        error = InstallerNotFoundError("No installer found")

        # Assert
        assert str(error) == "No installer found"
        assert isinstance(error, UpdateError)

    def test_network_error(self):
        """Test NetworkError exception."""
        # Arrange & Act
        error = NetworkError("Network unavailable")

        # Assert
        assert str(error) == "Network unavailable"
        assert isinstance(error, UpdateError)

    def test_upgrade_error(self):
        """Test UpgradeError exception."""
        # Arrange & Act
        error = UpgradeError("Upgrade failed")

        # Assert
        assert str(error) == "Upgrade failed"
        assert isinstance(error, UpdateError)

    def test_template_sync_error(self):
        """Test TemplateSyncError exception."""
        # Arrange & Act
        error = TemplateSyncError("Template sync failed")

        # Assert
        assert str(error) == "Template sync failed"
        assert isinstance(error, UpdateError)


# ============================================================================
# Test Version Comparison
# ============================================================================


class TestVersionComparison:
    """Test version comparison functionality."""

    def test_compare_versions_upgrade_needed(self):
        """Test version comparison when upgrade is needed."""
        # Arrange
        current = "0.1.0"
        latest = "0.2.0"

        # Act
        result = _compare_versions(current, latest)

        # Assert
        assert result == -1

    def test_compare_versions_up_to_date(self):
        """Test version comparison when up to date."""
        # Arrange
        current = "0.2.0"
        latest = "0.2.0"

        # Act
        result = _compare_versions(current, latest)

        # Assert
        assert result == 0

    def test_compare_versions_newer_installed(self):
        """Test version comparison when newer version already installed."""
        # Arrange
        current = "0.3.0"
        latest = "0.2.0"

        # Act
        result = _compare_versions(current, latest)

        # Assert
        assert result == 1

    def test_compare_versions_semantic(self):
        """Test semantic version comparison."""
        # Arrange
        test_cases = [
            ("1.0.0", "1.0.1", -1),
            ("1.0.0", "1.1.0", -1),
            ("1.0.0", "2.0.0", -1),
            ("1.2.3", "1.2.3", 0),
            ("2.0.0", "1.9.9", 1),
        ]

        # Act & Assert
        for current, latest, expected in test_cases:
            result = _compare_versions(current, latest)
            assert result == expected

    def test_compare_versions_prerelease(self):
        """Test version comparison with prerelease versions."""
        # Arrange
        current = "0.1.0"
        latest = "0.2.0a1"

        # Act
        result = _compare_versions(current, latest)

        # Assert
        assert result == -1


# ============================================================================
# Test Current Version Detection
# ============================================================================


class TestCurrentVersionDetection:
    """Test getting current installed version."""

    @patch("moai_adk.cli.commands.update.__version__", "0.6.1")
    def test_get_current_version(self):
        """Test getting current moai-adk version."""
        # Act
        result = _get_current_version()

        # Assert
        assert result == "0.6.1"

    @patch("moai_adk.cli.commands.update.__version__", "1.0.0")
    def test_get_current_version_different(self):
        """Test getting different current version."""
        # Act
        result = _get_current_version()

        # Assert
        assert result == "1.0.0"


# ============================================================================
# Test Latest Version Detection (Network)
# ============================================================================


class TestLatestVersionDetection:
    """Test fetching latest version from PyPI."""

    @patch("urllib.request.urlopen")
    def test_get_latest_version_success(self, mock_urlopen):
        """Test successfully fetching latest version from PyPI."""
        # Arrange
        mock_response = MagicMock()
        mock_response.read.return_value = json.dumps({"info": {"version": "0.7.0"}}).encode("utf-8")
        mock_response.__enter__.return_value = mock_response
        mock_urlopen.return_value = mock_response

        # Act
        result = _get_latest_version()

        # Assert
        assert result == "0.7.0"

    @patch("urllib.request.urlopen")
    def test_get_latest_version_json_error(self, mock_urlopen):
        """Test handling JSON decode error from PyPI."""
        # Arrange
        mock_response = MagicMock()
        mock_response.read.return_value = b"invalid json"
        mock_response.__enter__.return_value = mock_response
        mock_urlopen.return_value = mock_response

        # Act & Assert
        with pytest.raises(RuntimeError, match="Failed to fetch latest version"):
            _get_latest_version()

    @patch("urllib.request.urlopen")
    def test_get_latest_version_network_error(self, mock_urlopen):
        """Test handling network error when fetching from PyPI."""
        # Arrange
        import urllib.error

        mock_urlopen.side_effect = urllib.error.URLError("Network unreachable")

        # Act & Assert
        with pytest.raises(RuntimeError, match="Failed to fetch latest version"):
            _get_latest_version()

    @patch("urllib.request.urlopen")
    def test_get_latest_version_timeout(self, mock_urlopen):
        """Test handling timeout when fetching from PyPI."""
        # Arrange
        mock_urlopen.side_effect = TimeoutError("Connection timeout")

        # Act & Assert
        with pytest.raises(RuntimeError, match="Failed to fetch latest version"):
            _get_latest_version()


# ============================================================================
# Test Installer Detection
# ============================================================================


class TestInstallerDetection:
    """Test installer detection functions."""

    @patch("subprocess.run")
    def test_is_installed_via_uv_tool_success(self, mock_run):
        """Test successful uv tool detection."""
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
        call_args = mock_run.call_args
        assert call_args[0][0] == ["uv", "tool", "list"]
        assert call_args[1]["timeout"] == TOOL_DETECTION_TIMEOUT

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
    def test_is_installed_via_uv_tool_failure(self, mock_run):
        """Test uv tool detection with non-zero return code."""
        # Arrange
        mock_run.return_value = MagicMock(returncode=1)

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
        """Test successful pipx detection."""
        # Arrange
        mock_run.return_value = MagicMock(
            returncode=0,
            stdout="moai-adk 0.8.3",
        )

        # Act
        result = _is_installed_via_pipx()

        # Assert
        assert result is True

    @patch("subprocess.run")
    def test_is_installed_via_pipx_not_found(self, mock_run):
        """Test pipx detection when moai-adk not installed."""
        # Arrange
        mock_run.return_value = MagicMock(
            returncode=0,
            stdout="other-tool 1.0.0",
        )

        # Act
        result = _is_installed_via_pipx()

        # Assert
        assert result is False

    @patch("subprocess.run")
    def test_is_installed_via_pipx_error(self, mock_run):
        """Test pipx detection with error."""
        # Arrange
        mock_run.side_effect = OSError("Command failed")

        # Act
        result = _is_installed_via_pipx()

        # Assert
        assert result is False

    @patch("subprocess.run")
    def test_is_installed_via_pip_success(self, mock_run):
        """Test successful pip detection."""
        # Arrange
        mock_run.return_value = MagicMock(returncode=0)

        # Act
        result = _is_installed_via_pip()

        # Assert
        assert result is True

    @patch("subprocess.run")
    def test_is_installed_via_pip_not_found(self, mock_run):
        """Test pip detection when package not found."""
        # Arrange
        mock_run.return_value = MagicMock(returncode=1)

        # Act
        result = _is_installed_via_pip()

        # Assert
        assert result is False

    @patch("subprocess.run")
    def test_is_installed_via_pip_error(self, mock_run):
        """Test pip detection with error."""
        # Arrange
        mock_run.side_effect = FileNotFoundError()

        # Act
        result = _is_installed_via_pip()

        # Assert
        assert result is False


# ============================================================================
# Test Tool Installer Detection
# ============================================================================


class TestDetectToolInstaller:
    """Test overall tool installer detection."""

    @patch("moai_adk.cli.commands.update._is_installed_via_uv_tool")
    @patch("moai_adk.cli.commands.update._is_installed_via_pipx")
    @patch("moai_adk.cli.commands.update._is_installed_via_pip")
    def test_detect_tool_installer_uv(self, mock_pip, mock_pipx, mock_uv):
        """Test detecting uv tool installer."""
        # Arrange
        mock_uv.return_value = True
        mock_pipx.return_value = False
        mock_pip.return_value = False

        # Act
        result = _detect_tool_installer()

        # Assert
        assert result == UV_TOOL_COMMAND

    @patch("moai_adk.cli.commands.update._is_installed_via_uv_tool")
    @patch("moai_adk.cli.commands.update._is_installed_via_pipx")
    @patch("moai_adk.cli.commands.update._is_installed_via_pip")
    def test_detect_tool_installer_pipx(self, mock_pip, mock_pipx, mock_uv):
        """Test detecting pipx installer."""
        # Arrange
        mock_uv.return_value = False
        mock_pipx.return_value = True
        mock_pip.return_value = False

        # Act
        result = _detect_tool_installer()

        # Assert
        assert result == PIPX_COMMAND

    @patch("moai_adk.cli.commands.update._is_installed_via_uv_tool")
    @patch("moai_adk.cli.commands.update._is_installed_via_pipx")
    @patch("moai_adk.cli.commands.update._is_installed_via_pip")
    def test_detect_tool_installer_pip(self, mock_pip, mock_pipx, mock_uv):
        """Test detecting pip installer."""
        # Arrange
        mock_uv.return_value = False
        mock_pipx.return_value = False
        mock_pip.return_value = True

        # Act
        result = _detect_tool_installer()

        # Assert
        assert result == PIP_COMMAND

    @patch("moai_adk.cli.commands.update._is_installed_via_uv_tool")
    @patch("moai_adk.cli.commands.update._is_installed_via_pipx")
    @patch("moai_adk.cli.commands.update._is_installed_via_pip")
    def test_detect_tool_installer_none(self, mock_pip, mock_pipx, mock_uv):
        """Test when no installer is detected."""
        # Arrange
        mock_uv.return_value = False
        mock_pipx.return_value = False
        mock_pip.return_value = False

        # Act
        result = _detect_tool_installer()

        # Assert
        assert result is None

    @patch("moai_adk.cli.commands.update._is_installed_via_uv_tool")
    @patch("moai_adk.cli.commands.update._is_installed_via_pipx")
    @patch("moai_adk.cli.commands.update._is_installed_via_pip")
    def test_detect_tool_installer_priority(self, mock_pip, mock_pipx, mock_uv):
        """Test that uv tool has priority over pipx and pip."""
        # Arrange
        mock_uv.return_value = True
        mock_pipx.return_value = True  # Also true but should be ignored
        mock_pip.return_value = True  # Also true but should be ignored

        # Act
        result = _detect_tool_installer()

        # Assert
        assert result == UV_TOOL_COMMAND


# ============================================================================
# Test Package Config Version
# ============================================================================


class TestPackageConfigVersion:
    """Test getting package configuration version."""

    @patch("moai_adk.cli.commands.update.__version__", "0.8.0")
    def test_get_package_config_version(self):
        """Test getting package config version."""
        # Act
        result = _get_package_config_version()

        # Assert
        assert result == "0.8.0"

    @patch("moai_adk.cli.commands.update.__version__", "1.2.3")
    def test_get_package_config_version_different(self):
        """Test getting different package config version."""
        # Act
        result = _get_package_config_version()

        # Assert
        assert result == "1.2.3"


# ============================================================================
# Test Project Config Version
# ============================================================================


class TestProjectConfigVersion:
    """Test getting project configuration version."""

    def test_get_project_config_version_no_config(self):
        """Test getting version when config.json doesn't exist."""
        # Arrange
        with patch("pathlib.Path.exists", return_value=False):
            project_path = Path("/test/project")

        # Act
        result = _get_project_config_version(project_path)

        # Assert
        assert result == "0.0.0"

    def test_get_project_config_version_from_template_version(self):
        """Test reading version from template_version field."""
        # Arrange
        import tempfile

        config_data = {"project": {"template_version": "0.7.2"}}

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            moai_dir = project_path / ".moai" / "config"
            moai_dir.mkdir(parents=True, exist_ok=True)
            config_file = moai_dir / "config.json"
            config_file.write_text(json.dumps(config_data), encoding="utf-8")

            # Act
            result = _get_project_config_version(project_path)

        # Assert
        assert result == "0.7.2"

    def test_get_project_config_version_from_moai_version(self):
        """Test reading version from moai.version field as fallback."""
        # Arrange
        import tempfile

        config_data = {"moai": {"version": "0.6.5"}}

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            moai_dir = project_path / ".moai" / "config"
            moai_dir.mkdir(parents=True, exist_ok=True)
            config_file = moai_dir / "config.json"
            config_file.write_text(json.dumps(config_data), encoding="utf-8")

            # Act
            result = _get_project_config_version(project_path)

        # Assert
        assert result == "0.6.5"

    def test_get_project_config_version_placeholder(self):
        """Test when version field contains unsubstituted placeholder."""
        # Arrange
        import tempfile

        config_data = {"project": {"template_version": "{{VERSION}}"}}

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            moai_dir = project_path / ".moai" / "config"
            moai_dir.mkdir(parents=True, exist_ok=True)
            config_file = moai_dir / "config.json"
            config_file.write_text(json.dumps(config_data), encoding="utf-8")

            # Act
            result = _get_project_config_version(project_path)

        # Assert
        assert result == "0.0.0"

    def test_get_project_config_version_invalid_json(self):
        """Test handling invalid JSON in config file."""
        # Arrange
        import tempfile

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            moai_dir = project_path / ".moai" / "config"
            moai_dir.mkdir(parents=True, exist_ok=True)
            config_file = moai_dir / "config.json"
            config_file.write_text("invalid json", encoding="utf-8")

            # Act & Assert
            with pytest.raises(ValueError, match="Failed to parse"):
                _get_project_config_version(project_path)


# ============================================================================
# Test Merge Strategy Selection
# ============================================================================


class TestMergeStrategySelection:
    """Test merge strategy selection."""

    def test_ask_merge_strategy_yes_flag(self):
        """Test merge strategy with --yes flag."""
        # Act
        result = _ask_merge_strategy(yes=True)

        # Assert
        assert result == "auto"

    @patch("click.prompt")
    def test_ask_merge_strategy_user_selects_auto(self, mock_prompt):
        """Test user selecting auto merge strategy."""
        # Arrange
        mock_prompt.return_value = "1"

        # Act
        result = _ask_merge_strategy(yes=False)

        # Assert
        assert result == "auto"

    @patch("click.prompt")
    def test_ask_merge_strategy_user_selects_manual(self, mock_prompt):
        """Test user selecting manual merge strategy."""
        # Arrange
        mock_prompt.return_value = "2"

        # Act
        result = _ask_merge_strategy(yes=False)

        # Assert
        assert result == "manual"

    @patch("click.prompt")
    def test_ask_merge_strategy_default(self, mock_prompt):
        """Test merge strategy with default (empty input)."""
        # Arrange
        mock_prompt.return_value = "1"

        # Act
        result = _ask_merge_strategy(yes=False)

        # Assert
        assert result == "auto"


# ============================================================================
# Test Constants
# ============================================================================


class TestConstants:
    """Test module constants."""

    def test_tool_detection_timeout(self):
        """Test tool detection timeout constant."""
        # Assert
        assert TOOL_DETECTION_TIMEOUT == 5

    def test_uv_tool_command(self):
        """Test uv tool command constant."""
        # Assert
        assert UV_TOOL_COMMAND == ["uv", "tool", "upgrade", "moai-adk"]

    def test_pipx_command(self):
        """Test pipx command constant."""
        # Assert
        assert PIPX_COMMAND == ["pipx", "upgrade", "moai-adk"]

    def test_pip_command(self):
        """Test pip command constant."""
        # Assert
        assert PIP_COMMAND == ["pip", "install", "--upgrade", "moai-adk"]


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
