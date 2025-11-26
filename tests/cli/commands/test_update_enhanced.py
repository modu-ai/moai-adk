"""Enhanced comprehensive tests for update.py command

This test file targets comprehensive coverage of update.py (786 statements).
Current coverage: 14.12% → Target: 90%+

Test Coverage Strategy:
1. Tool detection functions (_is_installed_via_*)
2. Version comparison and management
3. Cache management and stale detection
4. Template synchronization and backup
5. Settings preservation and restoration
6. Skill detection and restoration
7. Merge strategies (auto vs manual)
8. Migration execution
9. Error handling paths
10. Context building and validation
"""

import json
import subprocess
import urllib.error
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest
from click.testing import CliRunner

from moai_adk.cli.commands.update import (
    UV_TOOL_COMMAND,
    _apply_context_to_file,
    _ask_merge_strategy,
    _build_template_context,
    _clear_uv_package_cache,
    _coalesce,
    _compare_versions,
    _detect_custom_skills,
    _detect_stale_cache,
    _detect_tool_installer,
    _execute_migration_if_needed,
    _execute_upgrade,
    _execute_upgrade_with_retry,
    _extract_project_section,
    _generate_manual_merge_guide,
    _get_current_version,
    _get_latest_version,
    _get_package_config_version,
    _get_project_config_version,
    _get_template_skill_names,
    _is_installed_via_pip,
    _is_installed_via_pipx,
    _is_installed_via_uv_tool,
    _is_placeholder,
    _load_existing_config,
    _preserve_project_metadata,
    _preserve_user_settings,
    _prompt_skill_restore,
    _restore_selected_skills,
    _restore_user_settings,
    _show_installer_not_found_help,
    _show_network_error_help,
    _show_post_update_guidance,
    _show_template_sync_failure_help,
    _show_timeout_error_help,
    _show_upgrade_failure_help,
    _show_version_info,
    _sync_templates,
    _validate_template_substitution,
    _validate_template_substitution_with_rollback,
    get_latest_version,
    set_optimized_false,
    update,
)


class TestToolDetection:
    """Test tool installation detection functions"""

    def test_is_installed_via_uv_tool_success(self):
        """Test UV tool detection when moai-adk is installed."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="moai-adk v0.6.1", stderr="")
            assert _is_installed_via_uv_tool() is True
            mock_run.assert_called_once()

    def test_is_installed_via_uv_tool_not_found(self):
        """Test UV tool detection when moai-adk not in list."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="other-package v1.0", stderr="")
            assert _is_installed_via_uv_tool() is False

    def test_is_installed_via_uv_tool_command_fails(self):
        """Test UV tool detection when command fails."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=1, stdout="", stderr="Error")
            assert _is_installed_via_uv_tool() is False

    def test_is_installed_via_uv_tool_not_installed(self):
        """Test UV tool detection when uv not installed."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = FileNotFoundError("uv not found")
            assert _is_installed_via_uv_tool() is False

    def test_is_installed_via_uv_tool_timeout(self):
        """Test UV tool detection timeout."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.TimeoutExpired(cmd="uv", timeout=5)
            assert _is_installed_via_uv_tool() is False

    def test_is_installed_via_uv_tool_os_error(self):
        """Test UV tool detection with OS error."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = OSError("Permission denied")
            assert _is_installed_via_uv_tool() is False

    def test_is_installed_via_pipx_success(self):
        """Test pipx detection when moai-adk is installed."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="package moai-adk v0.6.1", stderr="")
            assert _is_installed_via_pipx() is True

    def test_is_installed_via_pipx_not_found(self):
        """Test pipx detection when moai-adk not in list."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="other-pkg", stderr="")
            assert _is_installed_via_pipx() is False

    def test_is_installed_via_pipx_command_fails(self):
        """Test pipx detection when command fails."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=1, stdout="", stderr="Error")
            assert _is_installed_via_pipx() is False

    def test_is_installed_via_pipx_not_installed(self):
        """Test pipx detection when pipx not installed."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = FileNotFoundError("pipx not found")
            assert _is_installed_via_pipx() is False

    def test_is_installed_via_pipx_timeout(self):
        """Test pipx detection timeout."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.TimeoutExpired(cmd="pipx", timeout=5)
            assert _is_installed_via_pipx() is False

    def test_is_installed_via_pip_success(self):
        """Test pip detection when moai-adk is installed."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="Name: moai-adk\nVersion: 0.6.1", stderr="")
            assert _is_installed_via_pip() is True

    def test_is_installed_via_pip_not_found(self):
        """Test pip detection when moai-adk not installed."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=1, stdout="", stderr="Package not found")
            assert _is_installed_via_pip() is False

    def test_is_installed_via_pip_not_installed(self):
        """Test pip detection when pip not installed."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = FileNotFoundError("pip not found")
            assert _is_installed_via_pip() is False

    def test_is_installed_via_pip_timeout(self):
        """Test pip detection timeout."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.TimeoutExpired(cmd="pip", timeout=5)
            assert _is_installed_via_pip() is False

    def test_detect_tool_installer_uv_priority(self):
        """Test that uv tool has priority over pipx and pip."""
        with (
            patch("moai_adk.cli.commands.update._is_installed_via_uv_tool", return_value=True),
            patch("moai_adk.cli.commands.update._is_installed_via_pipx", return_value=True),
            patch("moai_adk.cli.commands.update._is_installed_via_pip", return_value=True),
        ):
            result = _detect_tool_installer()
            assert result == UV_TOOL_COMMAND

    def test_detect_tool_installer_pipx_priority(self):
        """Test that pipx has priority over pip."""
        with (
            patch("moai_adk.cli.commands.update._is_installed_via_uv_tool", return_value=False),
            patch("moai_adk.cli.commands.update._is_installed_via_pipx", return_value=True),
            patch("moai_adk.cli.commands.update._is_installed_via_pip", return_value=True),
        ):
            result = _detect_tool_installer()
            assert result == ["pipx", "upgrade", "moai-adk"]

    def test_detect_tool_installer_pip_fallback(self):
        """Test pip as fallback when uv and pipx not available."""
        with (
            patch("moai_adk.cli.commands.update._is_installed_via_uv_tool", return_value=False),
            patch("moai_adk.cli.commands.update._is_installed_via_pipx", return_value=False),
            patch("moai_adk.cli.commands.update._is_installed_via_pip", return_value=True),
        ):
            result = _detect_tool_installer()
            assert result == ["pip", "install", "--upgrade", "moai-adk"]

    def test_detect_tool_installer_none_found(self):
        """Test when no installer is detected."""
        with (
            patch("moai_adk.cli.commands.update._is_installed_via_uv_tool", return_value=False),
            patch("moai_adk.cli.commands.update._is_installed_via_pipx", return_value=False),
            patch("moai_adk.cli.commands.update._is_installed_via_pip", return_value=False),
        ):
            result = _detect_tool_installer()
            assert result is None


class TestVersionFunctions:
    """Test version-related functions"""

    def test_get_current_version_returns_string(self):
        """Test current version returns valid string."""
        version = _get_current_version()
        assert isinstance(version, str)
        assert len(version) > 0

    def test_get_latest_version_success(self):
        """Test fetching latest version from PyPI."""
        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = Mock()
            mock_response.read.return_value = b'{"info": {"version": "0.7.0"}}'
            mock_response.__enter__ = Mock(return_value=mock_response)
            mock_response.__exit__ = Mock(return_value=False)
            mock_urlopen.return_value = mock_response

            result = _get_latest_version()
            assert result == "0.7.0"

    def test_get_latest_version_url_error(self):
        """Test latest version with URL error."""
        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_urlopen.side_effect = urllib.error.URLError("Network error")
            with pytest.raises(RuntimeError, match="Failed to fetch latest version"):
                _get_latest_version()

    def test_get_latest_version_json_error(self):
        """Test latest version with invalid JSON."""
        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = Mock()
            mock_response.read.return_value = b"invalid json"
            mock_response.__enter__ = Mock(return_value=mock_response)
            mock_response.__exit__ = Mock(return_value=False)
            mock_urlopen.return_value = mock_response

            with pytest.raises(RuntimeError, match="Failed to fetch latest version"):
                _get_latest_version()

    def test_get_latest_version_missing_key(self):
        """Test latest version with missing version key."""
        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = Mock()
            mock_response.read.return_value = b'{"info": {"name": "moai-adk"}}'
            mock_response.__enter__ = Mock(return_value=mock_response)
            mock_response.__exit__ = Mock(return_value=False)
            mock_urlopen.return_value = mock_response

            with pytest.raises(RuntimeError, match="Failed to fetch latest version"):
                _get_latest_version()

    def test_get_latest_version_timeout(self):
        """Test latest version with timeout."""
        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_urlopen.side_effect = TimeoutError("Request timeout")
            with pytest.raises(RuntimeError, match="Failed to fetch latest version"):
                _get_latest_version()

    def test_compare_versions_current_less_than_latest(self):
        """Test version comparison: current < latest."""
        result = _compare_versions("0.6.0", "0.7.0")
        assert result == -1

    def test_compare_versions_current_equals_latest(self):
        """Test version comparison: current == latest."""
        result = _compare_versions("0.7.0", "0.7.0")
        assert result == 0

    def test_compare_versions_current_greater_than_latest(self):
        """Test version comparison: current > latest."""
        result = _compare_versions("0.8.0", "0.7.0")
        assert result == 1

    def test_compare_versions_prerelease(self):
        """Test version comparison with prerelease versions."""
        result = _compare_versions("0.7.0a1", "0.7.0")
        assert result == -1

    def test_get_package_config_version_returns_current(self):
        """Test package config version returns __version__."""
        with patch("moai_adk.cli.commands.update.__version__", "0.7.0"):
            result = _get_package_config_version()
            assert result == "0.7.0"

    def test_get_project_config_version_no_config(self, tmp_path):
        """Test project config version when no config exists."""
        result = _get_project_config_version(tmp_path)
        assert result == "0.0.0"

    def test_get_project_config_version_with_template_version(self, tmp_path):
        """Test project config version with template_version field."""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_data = {"project": {"template_version": "0.6.5"}, "moai": {"version": "0.6.0"}}
        (config_dir / "config.json").write_text(json.dumps(config_data))

        result = _get_project_config_version(tmp_path)
        assert result == "0.6.5"

    def test_get_project_config_version_fallback_moai_version(self, tmp_path):
        """Test project config version falls back to moai.version."""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_data = {"moai": {"version": "0.6.2"}, "project": {}}
        (config_dir / "config.json").write_text(json.dumps(config_data))

        result = _get_project_config_version(tmp_path)
        assert result == "0.6.2"

    def test_get_project_config_version_placeholder_values(self, tmp_path):
        """Test project config version with placeholder values."""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_data = {"project": {"template_version": "{{TEMPLATE_VERSION}}"}, "moai": {"version": "{{MOAI_VERSION}}"}}
        (config_dir / "config.json").write_text(json.dumps(config_data))

        result = _get_project_config_version(tmp_path)
        assert result == "0.0.0"

    def test_get_project_config_version_invalid_json(self, tmp_path):
        """Test project config version with invalid JSON."""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        (config_dir / "config.json").write_text("invalid json {")

        with pytest.raises(ValueError, match="Failed to parse project config.json"):
            _get_project_config_version(tmp_path)

    def test_get_latest_version_deprecated(self):
        """Test deprecated get_latest_version function."""
        with patch("moai_adk.cli.commands.update._get_latest_version", return_value="0.7.0"):
            result = get_latest_version()
            assert result == "0.7.0"

    def test_get_latest_version_deprecated_error_returns_none(self):
        """Test deprecated function returns None on error."""
        with patch("moai_adk.cli.commands.update._get_latest_version", side_effect=RuntimeError("Error")):
            result = get_latest_version()
            assert result is None


class TestCacheManagement:
    """Test cache detection and management"""

    def test_detect_stale_cache_when_stale(self):
        """Test stale cache detection when cache is stale."""
        result = _detect_stale_cache("Nothing to upgrade", "0.6.0", "0.7.0")
        assert result is True

    def test_detect_stale_cache_when_not_stale_same_version(self):
        """Test stale cache detection when versions are same."""
        result = _detect_stale_cache("Nothing to upgrade", "0.7.0", "0.7.0")
        assert result is False

    def test_detect_stale_cache_with_upgrade_output(self):
        """Test stale cache detection with upgrade output."""
        result = _detect_stale_cache("Updated moai-adk to 0.7.0", "0.6.0", "0.7.0")
        assert result is False

    def test_detect_stale_cache_empty_output(self):
        """Test stale cache detection with empty output."""
        result = _detect_stale_cache("", "0.6.0", "0.7.0")
        assert result is False

    def test_detect_stale_cache_version_parse_error(self):
        """Test stale cache detection with version parse error."""
        result = _detect_stale_cache("Nothing to upgrade", "invalid", "0.7.0")
        assert result is False

    def test_clear_uv_package_cache_success(self):
        """Test UV cache clearing success."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="Cache cleared", stderr="")
            result = _clear_uv_package_cache("moai-adk")
            assert result is True

    def test_clear_uv_package_cache_failure(self):
        """Test UV cache clearing failure."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=1, stdout="", stderr="Error")
            result = _clear_uv_package_cache("moai-adk")
            assert result is False

    def test_clear_uv_package_cache_timeout(self):
        """Test UV cache clearing timeout."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.TimeoutExpired(cmd="uv", timeout=10)
            result = _clear_uv_package_cache("moai-adk")
            assert result is False

    def test_clear_uv_package_cache_not_found(self):
        """Test UV cache clearing when uv not found."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = FileNotFoundError("uv not found")
            result = _clear_uv_package_cache("moai-adk")
            assert result is False

    def test_clear_uv_package_cache_unexpected_error(self):
        """Test UV cache clearing with unexpected error."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = RuntimeError("Unexpected error")
            result = _clear_uv_package_cache("moai-adk")
            assert result is False


class TestUpgradeExecution:
    """Test upgrade execution functions"""

    def test_execute_upgrade_with_retry_first_success(self):
        """Test upgrade with retry when first attempt succeeds."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="Successfully upgraded", stderr="")
            result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])
            assert result is True

    def test_execute_upgrade_with_retry_stale_cache_detected(self):
        """Test upgrade with retry when stale cache detected."""
        with (
            patch("subprocess.run") as mock_run,
            patch("moai_adk.cli.commands.update._get_current_version", return_value="0.6.0"),
            patch("moai_adk.cli.commands.update._get_latest_version", return_value="0.7.0"),
            patch("moai_adk.cli.commands.update._clear_uv_package_cache", return_value=True),
        ):

            # First call: stale cache
            # Second call: successful upgrade
            mock_run.side_effect = [
                MagicMock(returncode=0, stdout="Nothing to upgrade", stderr=""),
                MagicMock(returncode=0, stdout="Successfully upgraded", stderr=""),
            ]

            result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])
            assert result is True
            assert mock_run.call_count == 2

    def test_execute_upgrade_with_retry_cache_clear_fails(self):
        """Test upgrade with retry when cache clear fails."""
        with (
            patch("subprocess.run") as mock_run,
            patch("moai_adk.cli.commands.update._get_current_version", return_value="0.6.0"),
            patch("moai_adk.cli.commands.update._get_latest_version", return_value="0.7.0"),
            patch("moai_adk.cli.commands.update._clear_uv_package_cache", return_value=False),
        ):

            mock_run.return_value = MagicMock(returncode=0, stdout="Nothing to upgrade", stderr="")

            result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])
            assert result is False

    def test_execute_upgrade_with_retry_timeout(self):
        """Test upgrade with retry when timeout occurs."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.TimeoutExpired(cmd="uv", timeout=60)

            with pytest.raises(subprocess.TimeoutExpired):
                _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])

    def test_execute_upgrade_with_retry_version_check_fails(self):
        """Test upgrade with retry when version check fails."""
        with (
            patch("subprocess.run") as mock_run,
            patch("moai_adk.cli.commands.update._get_current_version", side_effect=RuntimeError("Error")),
        ):

            mock_run.return_value = MagicMock(returncode=0, stdout="Nothing to upgrade", stderr="")

            result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])
            assert result is True  # Should return original result on version check failure

    def test_execute_upgrade_with_retry_retry_fails(self):
        """Test upgrade with retry when retry also fails."""
        with (
            patch("subprocess.run") as mock_run,
            patch("moai_adk.cli.commands.update._get_current_version", return_value="0.6.0"),
            patch("moai_adk.cli.commands.update._get_latest_version", return_value="0.7.0"),
            patch("moai_adk.cli.commands.update._clear_uv_package_cache", return_value=True),
        ):

            # Both attempts return error
            mock_run.side_effect = [
                MagicMock(returncode=0, stdout="Nothing to upgrade", stderr=""),
                MagicMock(returncode=1, stdout="", stderr="Upgrade failed"),
            ]

            result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])
            assert result is False

    def test_execute_upgrade_success(self):
        """Test basic upgrade execution success."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="Success", stderr="")
            result = _execute_upgrade(["uv", "tool", "upgrade", "moai-adk"])
            assert result is True

    def test_execute_upgrade_failure(self):
        """Test basic upgrade execution failure."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=1, stdout="", stderr="Error")
            result = _execute_upgrade(["uv", "tool", "upgrade", "moai-adk"])
            assert result is False

    def test_execute_upgrade_timeout(self):
        """Test basic upgrade execution timeout."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.TimeoutExpired(cmd="uv", timeout=60)
            with pytest.raises(subprocess.TimeoutExpired):
                _execute_upgrade(["uv", "tool", "upgrade", "moai-adk"])

    def test_execute_upgrade_exception(self):
        """Test basic upgrade execution with exception."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = RuntimeError("Unexpected error")
            result = _execute_upgrade(["uv", "tool", "upgrade", "moai-adk"])
            assert result is False


class TestSettingsPreservation:
    """Test user settings preservation and restoration"""

    def test_preserve_user_settings_with_local_file(self, tmp_path):
        """Test preserving user settings when settings.local.json exists."""
        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        settings_local = claude_dir / "settings.local.json"
        settings_local.write_text('{"user": "test"}')

        result = _preserve_user_settings(tmp_path)

        assert "settings.local.json" in result
        assert result["settings.local.json"] is not None
        assert result["settings.local.json"].exists()

    def test_preserve_user_settings_no_local_file(self, tmp_path):
        """Test preserving user settings when no local file exists."""
        result = _preserve_user_settings(tmp_path)

        assert "settings.local.json" in result
        assert result["settings.local.json"] is None

    def test_preserve_user_settings_read_error(self, tmp_path):
        """Test preserving user settings with read error."""
        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        claude_dir / "settings.local.json"

        with patch.object(Path, "read_text", side_effect=PermissionError("No access")):
            with patch.object(Path, "exists", return_value=True):
                result = _preserve_user_settings(tmp_path)
                assert result["settings.local.json"] is None

    def test_restore_user_settings_success(self, tmp_path):
        """Test restoring user settings successfully."""
        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()

        backup_dir = tmp_path / ".moai-backups" / "settings-backup"
        backup_dir.mkdir(parents=True)
        backup_file = backup_dir / "settings.local.json"
        backup_file.write_text('{"restored": "data"}')

        preserved = {"settings.local.json": backup_file}
        result = _restore_user_settings(tmp_path, preserved)

        assert result is True
        settings_local = claude_dir / "settings.local.json"
        assert settings_local.exists()
        assert "restored" in settings_local.read_text()

    def test_restore_user_settings_no_backup(self, tmp_path):
        """Test restoring user settings when no backup exists."""
        preserved = {"settings.local.json": None}
        result = _restore_user_settings(tmp_path, preserved)
        assert result is True

    def test_restore_user_settings_failure(self, tmp_path):
        """Test restoring user settings with failure."""
        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()

        backup_file = tmp_path / "nonexistent" / "settings.local.json"
        preserved = {"settings.local.json": backup_file}

        result = _restore_user_settings(tmp_path, preserved)
        assert result is False


class TestSkillManagement:
    """Test skill detection and restoration"""

    def test_get_template_skill_names_success(self):
        """Test getting template skill names."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.iterdir") as mock_iterdir:
                # Mock skill directories
                mock_skill1 = Mock(spec=Path)
                mock_skill1.name = "moai-skill-1"
                mock_skill1.is_dir.return_value = True

                mock_skill2 = Mock(spec=Path)
                mock_skill2.name = "moai-skill-2"
                mock_skill2.is_dir.return_value = True

                mock_iterdir.return_value = [mock_skill1, mock_skill2]

                result = _get_template_skill_names()
                assert "moai-skill-1" in result
                assert "moai-skill-2" in result

    def test_get_template_skill_names_no_directory(self):
        """Test getting template skill names when directory doesn't exist."""
        with patch("pathlib.Path.exists", return_value=False):
            result = _get_template_skill_names()
            assert result == set()

    def test_detect_custom_skills_with_custom(self, tmp_path):
        """Test detecting custom skills."""
        skills_dir = tmp_path / ".claude" / "skills"
        skills_dir.mkdir(parents=True)

        # Create skill directories
        (skills_dir / "custom-skill-1").mkdir()
        (skills_dir / "custom-skill-2").mkdir()
        (skills_dir / "moai-skill-template").mkdir()

        template_skills = {"moai-skill-template"}
        result = _detect_custom_skills(tmp_path, template_skills)

        assert "custom-skill-1" in result
        assert "custom-skill-2" in result
        assert "moai-skill-template" not in result

    def test_detect_custom_skills_no_directory(self, tmp_path):
        """Test detecting custom skills when directory doesn't exist."""
        result = _detect_custom_skills(tmp_path, set())
        assert result == []

    def test_prompt_skill_restore_with_yes_flag(self):
        """Test prompting skill restore with --yes flag."""
        custom_skills = ["custom-skill-1", "custom-skill-2"]
        result = _prompt_skill_restore(custom_skills, yes=True)
        assert result == []

    def test_prompt_skill_restore_no_skills(self):
        """Test prompting skill restore with no custom skills."""
        result = _prompt_skill_restore([], yes=False)
        assert result == []

    def test_prompt_skill_restore_user_selection(self):
        """Test prompting skill restore with user selection."""
        custom_skills = ["custom-skill-1", "custom-skill-2"]

        with patch("questionary.checkbox") as mock_checkbox:
            mock_checkbox.return_value.ask.return_value = ["custom-skill-1"]
            result = _prompt_skill_restore(custom_skills, yes=False)
            assert result == ["custom-skill-1"]

    def test_restore_selected_skills_success(self, tmp_path):
        """Test restoring selected skills successfully."""
        backup_path = tmp_path / "backup"
        backup_skills = backup_path / ".claude" / "skills"
        backup_skills.mkdir(parents=True)
        (backup_skills / "custom-skill-1").mkdir()
        (backup_skills / "custom-skill-1" / "SKILL.md").write_text("# Skill 1")

        project_path = tmp_path / "project"
        project_path.mkdir()

        result = _restore_selected_skills(["custom-skill-1"], backup_path, project_path)

        assert result is True
        restored_skill = project_path / ".claude" / "skills" / "custom-skill-1" / "SKILL.md"
        assert restored_skill.exists()

    def test_restore_selected_skills_no_skills(self, tmp_path):
        """Test restoring with no skills selected."""
        result = _restore_selected_skills([], tmp_path, tmp_path)
        assert result is True

    def test_restore_selected_skills_not_in_backup(self, tmp_path):
        """Test restoring skill not in backup."""
        backup_path = tmp_path / "backup"
        project_path = tmp_path / "project"
        project_path.mkdir()

        result = _restore_selected_skills(["nonexistent-skill"], backup_path, project_path)
        assert result is False

    def test_restore_selected_skills_copy_error(self, tmp_path):
        """Test restoring skills with copy error."""
        backup_path = tmp_path / "backup"
        backup_skills = backup_path / ".claude" / "skills"
        backup_skills.mkdir(parents=True)
        (backup_skills / "custom-skill-1").mkdir()

        project_path = tmp_path / "project"
        project_path.mkdir()

        with patch("shutil.copytree", side_effect=PermissionError("No access")):
            result = _restore_selected_skills(["custom-skill-1"], backup_path, project_path)
            assert result is False


class TestMergeStrategies:
    """Test merge strategy selection and execution"""

    def test_ask_merge_strategy_with_yes_flag(self):
        """Test asking merge strategy with --yes flag."""
        result = _ask_merge_strategy(yes=True)
        assert result == "auto"

    def test_ask_merge_strategy_auto_selected(self):
        """Test asking merge strategy with auto selection."""
        with patch("click.prompt", return_value="1"):
            result = _ask_merge_strategy(yes=False)
            assert result == "auto"

    def test_ask_merge_strategy_manual_selected(self):
        """Test asking merge strategy with manual selection."""
        with patch("click.prompt", return_value="2"):
            result = _ask_merge_strategy(yes=False)
            assert result == "manual"

    def test_generate_manual_merge_guide_success(self, tmp_path):
        """Test generating manual merge guide."""
        project_path = tmp_path / "project"
        project_path.mkdir()

        backup_path = project_path / ".moai-backups" / "backup"
        backup_claude = backup_path / ".claude"
        backup_claude.mkdir(parents=True)
        (backup_claude / "settings.json").write_text('{"old": "content"}')

        template_path = tmp_path / "template"
        template_path.mkdir()

        project_claude = project_path / ".claude"
        project_claude.mkdir(parents=True)
        (project_claude / "settings.json").write_text('{"new": "content"}')

        result = _generate_manual_merge_guide(backup_path, template_path, project_path)

        assert result.exists()
        assert result.name == "merge-guide.md"
        content = result.read_text()
        assert "Merge Guide" in content
        assert "Manual Merge Mode" in content

    def test_generate_manual_merge_guide_no_changes(self, tmp_path):
        """Test generating merge guide when no changes detected."""
        project_path = tmp_path / "project"
        project_path.mkdir()

        backup_path = project_path / ".moai-backups" / "backup"
        backup_path.mkdir(parents=True)

        template_path = tmp_path / "template"
        template_path.mkdir()

        result = _generate_manual_merge_guide(backup_path, template_path, project_path)

        assert result.exists()
        content = result.read_text()
        assert "No changes detected" in content


class TestTemplateContext:
    """Test template context building"""

    def test_is_placeholder_true(self):
        """Test placeholder detection for valid placeholders."""
        assert _is_placeholder("{{PROJECT_NAME}}") is True
        assert _is_placeholder("  {{VAR}}  ") is True

    def test_is_placeholder_false(self):
        """Test placeholder detection for non-placeholders."""
        assert _is_placeholder("normal text") is False
        assert _is_placeholder("") is False
        assert _is_placeholder(123) is False
        assert _is_placeholder(None) is False

    def test_coalesce_first_valid_string(self):
        """Test coalesce returns first non-empty, non-placeholder string."""
        result = _coalesce("", "{{PLACEHOLDER}}", "valid", "also-valid", default="default")
        assert result == "valid"

    def test_coalesce_with_default(self):
        """Test coalesce returns default when no valid values."""
        result = _coalesce("", "{{PLACEHOLDER}}", default="default-value")
        assert result == "default-value"

    def test_coalesce_with_non_string(self):
        """Test coalesce handles non-string values."""
        result = _coalesce("", None, 123, default="default")
        assert result == "123"

    def test_extract_project_section_exists(self):
        """Test extracting project section when it exists."""
        config = {"project": {"name": "test", "mode": "personal"}}
        result = _extract_project_section(config)
        assert result == {"name": "test", "mode": "personal"}

    def test_extract_project_section_not_exists(self):
        """Test extracting project section when it doesn't exist."""
        config = {"other": "data"}
        result = _extract_project_section(config)
        assert result == {}

    def test_extract_project_section_not_dict(self):
        """Test extracting project section when it's not a dict."""
        config = {"project": "not a dict"}
        result = _extract_project_section(config)
        assert result == {}

    def test_build_template_context_full(self, tmp_path):
        """Test building template context with all data."""
        existing_config = {
            "project": {
                "name": "test-project",
                "mode": "personal",
                "description": "Test description",
                "version": "1.0.0",
                "created_at": "2025-01-01 00:00:00",
                "locale": "en_US",
                "language": "python",
                "author": "tester",
            },
            "language": {"conversation_language": "en", "conversation_language_name": "English"},
        }

        result = _build_template_context(tmp_path, existing_config, "0.7.0")

        assert result["MOAI_VERSION"] == "0.7.0"
        assert result["PROJECT_NAME"] == "test-project"
        assert result["PROJECT_MODE"] == "personal"
        assert result["PROJECT_DESCRIPTION"] == "Test description"
        assert result["PROJECT_VERSION"] == "1.0.0"
        assert result["CONVERSATION_LANGUAGE"] == "en"
        assert result["CODEBASE_LANGUAGE"] == "python"
        assert result["AUTHOR"] == "tester"

    def test_build_template_context_minimal(self, tmp_path):
        """Test building template context with minimal data."""
        result = _build_template_context(tmp_path, {}, "0.7.0")

        assert result["MOAI_VERSION"] == "0.7.0"
        assert result["PROJECT_NAME"] == tmp_path.name
        assert result["PROJECT_MODE"] == "personal"
        assert result["CONVERSATION_LANGUAGE"] == "en"

    def test_build_template_context_windows_path(self, tmp_path):
        """Test building template context on Windows."""
        with patch("platform.system", return_value="Windows"):
            result = _build_template_context(tmp_path, {}, "0.7.0")
            assert result["PROJECT_DIR"] == "%CLAUDE_PROJECT_DIR%"

    def test_build_template_context_unix_path(self, tmp_path):
        """Test building template context on Unix-like systems."""
        with patch("platform.system", return_value="Darwin"):
            result = _build_template_context(tmp_path, {}, "0.7.0")
            assert result["PROJECT_DIR"] == "$CLAUDE_PROJECT_DIR"


class TestConfigManagement:
    """Test configuration loading and preservation"""

    def test_load_existing_config_success(self, tmp_path):
        """Test loading existing config successfully."""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_data = {"project": {"name": "test"}, "moai": {"version": "0.7.0"}}
        (config_dir / "config.json").write_text(json.dumps(config_data))

        result = _load_existing_config(tmp_path)
        assert result == config_data

    def test_load_existing_config_not_exists(self, tmp_path):
        """Test loading config when it doesn't exist."""
        result = _load_existing_config(tmp_path)
        assert result == {}

    def test_load_existing_config_invalid_json(self, tmp_path):
        """Test loading config with invalid JSON."""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        (config_dir / "config.json").write_text("invalid json {")

        result = _load_existing_config(tmp_path)
        assert result == {}

    def test_set_optimized_false_success(self, tmp_path):
        """Test setting optimized to false."""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_data = {"project": {"optimized": True}}
        (config_dir / "config.json").write_text(json.dumps(config_data))

        set_optimized_false(tmp_path)

        updated_config = json.loads((config_dir / "config.json").read_text())
        assert updated_config["project"]["optimized"] is False

    def test_set_optimized_false_no_config(self, tmp_path):
        """Test setting optimized when config doesn't exist."""
        # Should not raise exception
        set_optimized_false(tmp_path)

    def test_set_optimized_false_invalid_json(self, tmp_path):
        """Test setting optimized with invalid JSON."""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        (config_dir / "config.json").write_text("invalid json")

        # Should not raise exception
        set_optimized_false(tmp_path)

    def test_preserve_project_metadata_success(self, tmp_path):
        """Test preserving project metadata."""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_data = {"project": {}, "moai": {}}
        (config_dir / "config.json").write_text(json.dumps(config_data))

        context = {
            "PROJECT_NAME": "test-project",
            "PROJECT_MODE": "personal",
            "PROJECT_DESCRIPTION": "Test desc",
            "CREATION_TIMESTAMP": "2025-01-01 00:00:00",
        }
        existing_config = {"project": {"optimized": True}}

        _preserve_project_metadata(tmp_path, context, existing_config, "0.7.0")

        updated = json.loads((config_dir / "config.json").read_text())
        assert updated["project"]["name"] == "test-project"
        assert updated["project"]["template_version"] == "0.7.0"
        assert updated["moai"]["version"] == "0.7.0"
        assert updated["project"]["optimized"] is True

    def test_preserve_project_metadata_no_config(self, tmp_path):
        """Test preserving metadata when no config exists."""
        # Should not raise exception
        _preserve_project_metadata(tmp_path, {}, {}, "0.7.0")

    def test_preserve_project_metadata_invalid_json(self, tmp_path):
        """Test preserving metadata with invalid JSON."""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        (config_dir / "config.json").write_text("invalid json")

        # Should not raise exception
        _preserve_project_metadata(tmp_path, {}, {}, "0.7.0")


class TestTemplateValidation:
    """Test template substitution validation"""

    def test_validate_template_substitution_success(self, tmp_path):
        """Test template validation when all variables substituted."""
        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        (claude_dir / "settings.json").write_text('{"project": "test"}')
        (tmp_path / "CLAUDE.md").write_text("# Project: test")

        # Should not raise exception
        _validate_template_substitution(tmp_path)

    def test_validate_template_substitution_with_placeholders(self, tmp_path):
        """Test template validation with unsubstituted placeholders."""
        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        (claude_dir / "settings.json").write_text('{"project": "{{PROJECT_NAME}}"}')

        # Should print warnings but not raise exception
        _validate_template_substitution(tmp_path)

    def test_validate_template_substitution_file_not_exists(self, tmp_path):
        """Test template validation when files don't exist."""
        # Should not raise exception
        _validate_template_substitution(tmp_path)

    def test_validate_template_substitution_with_rollback_success(self, tmp_path):
        """Test template validation with rollback capability - success."""
        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        (claude_dir / "settings.json").write_text('{"project": "test"}')
        (tmp_path / "CLAUDE.md").write_text("# Project: test")

        result = _validate_template_substitution_with_rollback(tmp_path, None)
        assert result is True

    def test_validate_template_substitution_with_rollback_failure(self, tmp_path):
        """Test template validation with rollback capability - failure."""
        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        (claude_dir / "settings.json").write_text('{"project": "{{PROJECT_NAME}}"}')

        backup_path = tmp_path / "backup"
        backup_path.mkdir()

        result = _validate_template_substitution_with_rollback(tmp_path, backup_path)
        assert result is False

    def test_validate_template_substitution_with_rollback_no_backup(self, tmp_path):
        """Test template validation with rollback but no backup."""
        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        (claude_dir / "settings.json").write_text('{"project": "{{PROJECT_NAME}}"}')

        result = _validate_template_substitution_with_rollback(tmp_path, None)
        assert result is False


class TestDisplayFunctions:
    """Test display and help functions"""

    def test_show_version_info(self, capsys):
        """Test showing version information."""
        _show_version_info("0.6.0", "0.7.0")
        captured = capsys.readouterr()
        assert "0.6.0" in captured.out
        assert "0.7.0" in captured.out

    def test_show_installer_not_found_help(self, capsys):
        """Test showing installer not found help."""
        _show_installer_not_found_help()
        captured = capsys.readouterr()
        assert "Cannot detect package installer" in captured.out
        assert "uv tool upgrade" in captured.out

    def test_show_upgrade_failure_help(self, capsys):
        """Test showing upgrade failure help."""
        _show_upgrade_failure_help(["uv", "tool", "upgrade", "moai-adk"])
        captured = capsys.readouterr()
        assert "Upgrade failed" in captured.out
        assert "uv" in captured.out

    def test_show_network_error_help(self, capsys):
        """Test showing network error help."""
        _show_network_error_help()
        captured = capsys.readouterr()
        assert "Cannot reach PyPI" in captured.out

    def test_show_template_sync_failure_help(self, capsys):
        """Test showing template sync failure help."""
        _show_template_sync_failure_help()
        captured = capsys.readouterr()
        assert "Template sync failed" in captured.out

    def test_show_timeout_error_help(self, capsys):
        """Test showing timeout error help."""
        _show_timeout_error_help()
        captured = capsys.readouterr()
        assert "timed out" in captured.out

    def test_show_post_update_guidance(self, tmp_path, capsys):
        """Test showing post-update guidance."""
        backup_path = tmp_path / "backup"
        backup_path.mkdir()

        _show_post_update_guidance(backup_path)
        captured = capsys.readouterr()
        assert "Update complete" in captured.out
        assert "/moai:0-project update" in captured.out


class TestApplyContext:
    """Test context application to files"""

    def test_apply_context_to_file_success(self, tmp_path):
        """Test applying context to file successfully."""
        from moai_adk.core.template.processor import TemplateProcessor

        target_file = tmp_path / "test.md"
        target_file.write_text("# {{PROJECT_NAME}}")

        processor = TemplateProcessor(tmp_path)
        processor.set_context({"PROJECT_NAME": "TestProject"})

        _apply_context_to_file(processor, target_file)

        content = target_file.read_text()
        assert "TestProject" in content
        assert "{{PROJECT_NAME}}" not in content

    def test_apply_context_to_file_no_context(self, tmp_path):
        """Test applying context when no context set."""
        from moai_adk.core.template.processor import TemplateProcessor

        target_file = tmp_path / "test.md"
        target_file.write_text("# Test")

        processor = TemplateProcessor(tmp_path)

        # Should not raise exception
        _apply_context_to_file(processor, target_file)

    def test_apply_context_to_file_not_exists(self, tmp_path):
        """Test applying context when file doesn't exist."""
        from moai_adk.core.template.processor import TemplateProcessor

        target_file = tmp_path / "nonexistent.md"

        processor = TemplateProcessor(tmp_path)
        processor.set_context({"PROJECT_NAME": "Test"})

        # Should not raise exception
        _apply_context_to_file(processor, target_file)

    def test_apply_context_to_file_binary(self, tmp_path):
        """Test applying context to binary file."""
        from moai_adk.core.template.processor import TemplateProcessor

        target_file = tmp_path / "test.bin"
        target_file.write_bytes(b"\x00\x01\x02\x03")

        processor = TemplateProcessor(tmp_path)
        processor.set_context({"PROJECT_NAME": "Test"})

        # Should handle UnicodeDecodeError gracefully
        _apply_context_to_file(processor, target_file)


class TestMigrationExecution:
    """Test migration execution"""

    def test_execute_migration_if_needed_no_migration(self, tmp_path):
        """Test migration execution when no migration needed."""
        with patch("moai_adk.cli.commands.update.VersionMigrator") as mock_migrator:
            mock_instance = Mock()
            mock_instance.needs_migration.return_value = False
            mock_migrator.return_value = mock_instance

            result = _execute_migration_if_needed(tmp_path, yes=False)
            assert result is True

    def test_execute_migration_if_needed_with_yes(self, tmp_path):
        """Test migration execution with --yes flag."""
        with patch("moai_adk.cli.commands.update.VersionMigrator") as mock_migrator:
            mock_instance = Mock()
            mock_instance.needs_migration.return_value = True
            mock_instance.get_migration_info.return_value = {
                "current_version": "0.23.0",
                "target_version": "0.24.0",
                "file_count": 2,
            }
            mock_instance.migrate_to_v024.return_value = True
            mock_migrator.return_value = mock_instance

            result = _execute_migration_if_needed(tmp_path, yes=True)
            assert result is True

    def test_execute_migration_if_needed_user_confirms(self, tmp_path):
        """Test migration execution when user confirms."""
        with (
            patch("moai_adk.cli.commands.update.VersionMigrator") as mock_migrator,
            patch("click.confirm", return_value=True),
        ):
            mock_instance = Mock()
            mock_instance.needs_migration.return_value = True
            mock_instance.get_migration_info.return_value = {
                "current_version": "0.23.0",
                "target_version": "0.24.0",
                "file_count": 2,
            }
            mock_instance.migrate_to_v024.return_value = True
            mock_migrator.return_value = mock_instance

            result = _execute_migration_if_needed(tmp_path, yes=False)
            assert result is True

    def test_execute_migration_if_needed_user_cancels(self, tmp_path):
        """Test migration execution when user cancels."""
        with (
            patch("moai_adk.cli.commands.update.VersionMigrator") as mock_migrator,
            patch("click.confirm", return_value=False),
        ):
            mock_instance = Mock()
            mock_instance.needs_migration.return_value = True
            mock_instance.get_migration_info.return_value = {
                "current_version": "0.23.0",
                "target_version": "0.24.0",
                "file_count": 2,
            }
            mock_migrator.return_value = mock_instance

            result = _execute_migration_if_needed(tmp_path, yes=False)
            assert result is False

    def test_execute_migration_if_needed_migration_fails(self, tmp_path):
        """Test migration execution when migration fails."""
        with (
            patch("moai_adk.cli.commands.update.VersionMigrator") as mock_migrator,
            patch("click.confirm", return_value=True),
        ):
            mock_instance = Mock()
            mock_instance.needs_migration.return_value = True
            mock_instance.get_migration_info.return_value = {
                "current_version": "0.23.0",
                "target_version": "0.24.0",
                "file_count": 2,
            }
            mock_instance.migrate_to_v024.return_value = False
            mock_migrator.return_value = mock_instance

            result = _execute_migration_if_needed(tmp_path, yes=False)
            assert result is False

    def test_execute_migration_if_needed_exception(self, tmp_path):
        """Test migration execution with exception."""
        with patch("moai_adk.cli.commands.update.VersionMigrator") as mock_migrator:
            mock_migrator.side_effect = RuntimeError("Migration error")

            result = _execute_migration_if_needed(tmp_path, yes=False)
            assert result is False


class TestSyncTemplates:
    """Test template synchronization"""

    def test_sync_templates_success(self, tmp_path):
        """Test successful template sync."""
        (tmp_path / ".moai" / "config").mkdir(parents=True)
        config_data = {"project": {"mode": "personal"}}
        (tmp_path / ".moai" / "config" / "config.json").write_text(json.dumps(config_data))

        with (
            patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor,
            patch("moai_adk.cli.commands.update._get_template_skill_names", return_value=set()),
            patch("moai_adk.cli.commands.update._detect_custom_skills", return_value=[]),
            patch("moai_adk.cli.commands.update.AlfredToMoaiMigrator") as mock_migrator,
        ):

            mock_instance = Mock()
            mock_instance.copy_templates.return_value = None
            mock_processor.return_value = mock_instance

            mock_mig_instance = Mock()
            mock_mig_instance.needs_migration.return_value = False
            mock_migrator.return_value = mock_mig_instance

            result = _sync_templates(tmp_path, force=False, yes=False)
            assert result is True

    def test_sync_templates_with_backup(self, tmp_path):
        """Test template sync with backup creation."""
        (tmp_path / ".moai" / "config").mkdir(parents=True)
        (tmp_path / ".claude").mkdir()
        config_data = {"project": {"mode": "personal"}}
        (tmp_path / ".moai" / "config" / "config.json").write_text(json.dumps(config_data))

        with (
            patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor,
            patch("moai_adk.core.template.backup.TemplateBackup") as mock_backup,
            patch("moai_adk.cli.commands.update.MergeAnalyzer") as mock_analyzer,
            patch("moai_adk.cli.commands.update._get_template_skill_names", return_value=set()),
            patch("moai_adk.cli.commands.update._detect_custom_skills", return_value=[]),
            patch("moai_adk.cli.commands.update.AlfredToMoaiMigrator") as mock_migrator,
        ):

            mock_proc_instance = Mock()
            mock_proc_instance.copy_templates.return_value = None
            mock_processor.return_value = mock_proc_instance

            mock_backup_instance = Mock()
            mock_backup_instance.has_existing_files.return_value = True
            mock_backup_instance.create_backup.return_value = tmp_path / "backup"
            mock_backup.return_value = mock_backup_instance

            mock_analyzer_instance = Mock()
            mock_analyzer_instance.analyze_merge.return_value = {"conflicts": []}
            mock_analyzer_instance.ask_user_confirmation.return_value = True
            mock_analyzer.return_value = mock_analyzer_instance

            mock_mig_instance = Mock()
            mock_mig_instance.needs_migration.return_value = False
            mock_migrator.return_value = mock_mig_instance

            result = _sync_templates(tmp_path, force=False, yes=False)
            assert result is True

    def test_sync_templates_user_cancels(self, tmp_path):
        """Test template sync when user cancels."""
        (tmp_path / ".moai" / "config").mkdir(parents=True)
        (tmp_path / ".claude").mkdir()
        config_data = {"project": {"mode": "personal"}}
        (tmp_path / ".moai" / "config" / "config.json").write_text(json.dumps(config_data))

        with (
            patch("moai_adk.core.template.backup.TemplateBackup") as mock_backup,
            patch("moai_adk.cli.commands.update.MergeAnalyzer") as mock_analyzer,
            patch("moai_adk.cli.commands.update._get_template_skill_names", return_value=set()),
            patch("moai_adk.cli.commands.update._detect_custom_skills", return_value=[]),
            patch("moai_adk.cli.commands.update.TemplateProcessor"),
        ):

            mock_backup_instance = Mock()
            mock_backup_instance.has_existing_files.return_value = True
            mock_backup_instance.create_backup.return_value = tmp_path / "backup"
            mock_backup.return_value = mock_backup_instance

            mock_analyzer_instance = Mock()
            mock_analyzer_instance.analyze_merge.return_value = {"conflicts": []}
            mock_analyzer_instance.ask_user_confirmation.return_value = False
            mock_analyzer.return_value = mock_analyzer_instance

            result = _sync_templates(tmp_path, force=False, yes=False)
            assert result is False

    def test_sync_templates_migration_fails(self, tmp_path):
        """Test template sync when migration fails."""
        (tmp_path / ".moai" / "config").mkdir(parents=True)
        config_data = {"project": {"mode": "personal"}}
        (tmp_path / ".moai" / "config" / "config.json").write_text(json.dumps(config_data))

        with (
            patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor,
            patch("moai_adk.cli.commands.update._get_template_skill_names", return_value=set()),
            patch("moai_adk.cli.commands.update._detect_custom_skills", return_value=[]),
            patch("moai_adk.cli.commands.update.AlfredToMoaiMigrator") as mock_migrator,
            patch("moai_adk.core.template.backup.TemplateBackup") as mock_backup,
        ):

            mock_instance = Mock()
            mock_instance.copy_templates.return_value = None
            mock_processor.return_value = mock_instance

            mock_mig_instance = Mock()
            mock_mig_instance.needs_migration.return_value = True
            mock_mig_instance.execute_migration.return_value = False
            mock_migrator.return_value = mock_mig_instance

            mock_backup_instance = Mock()
            mock_backup.return_value = mock_backup_instance

            tmp_path / "backup"
            result = _sync_templates(tmp_path, force=False, yes=False)
            assert result is False

    def test_sync_templates_validation_fails(self, tmp_path):
        """Test template sync when validation fails."""
        (tmp_path / ".moai" / "config").mkdir(parents=True)
        (tmp_path / ".claude").mkdir()
        config_data = {"project": {"mode": "personal"}}
        (tmp_path / ".moai" / "config" / "config.json").write_text(json.dumps(config_data))
        (tmp_path / ".claude" / "settings.json").write_text('{"var": "{{PLACEHOLDER}}"}')

        with (
            patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor,
            patch("moai_adk.cli.commands.update._get_template_skill_names", return_value=set()),
            patch("moai_adk.cli.commands.update._detect_custom_skills", return_value=[]),
            patch("moai_adk.cli.commands.update.AlfredToMoaiMigrator") as mock_migrator,
            patch("moai_adk.core.template.backup.TemplateBackup") as mock_backup,
        ):

            mock_instance = Mock()
            mock_instance.copy_templates.return_value = None
            mock_processor.return_value = mock_instance

            mock_mig_instance = Mock()
            mock_mig_instance.needs_migration.return_value = False
            mock_migrator.return_value = mock_mig_instance

            mock_backup_instance = Mock()
            mock_backup.return_value = mock_backup_instance

            result = _sync_templates(tmp_path, force=False, yes=False)
            # Should fail due to placeholder validation
            assert result is False


class TestUpdateCommandIntegration:
    """Integration tests for update command"""

    def test_update_templates_only_success(self, tmp_path):
        """Test update --templates-only flag."""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            moai_dir = Path(".moai")
            moai_dir.mkdir()

            with (
                patch("moai_adk.cli.commands.update._sync_templates", return_value=True),
                patch("moai_adk.cli.commands.update._preserve_user_settings", return_value={}),
                patch("moai_adk.cli.commands.update._restore_user_settings", return_value=True),
            ):

                result = runner.invoke(update, ["--templates-only"])
                assert result.exit_code == 0
                assert "Syncing templates only" in result.output

    def test_update_manual_merge_strategy(self, tmp_path):
        """Test update with --manual merge strategy."""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            moai_dir = Path(".moai") / "config"
            moai_dir.mkdir(parents=True)
            config_data = {"project": {"template_version": "0.6.0"}, "moai": {"version": "0.6.0"}}
            (moai_dir / "config.json").write_text(json.dumps(config_data))

            with (
                patch("moai_adk.cli.commands.update._get_current_version", return_value="0.7.0"),
                patch("moai_adk.cli.commands.update._get_latest_version", return_value="0.7.0"),
                patch("moai_adk.cli.commands.update._get_package_config_version", return_value="0.7.0"),
                patch("moai_adk.cli.commands.update._get_project_config_version", return_value="0.6.0"),
                patch("moai_adk.cli.commands.update._execute_migration_if_needed", return_value=True),
                patch("moai_adk.core.migration.backup_manager.BackupManager") as mock_backup,
                patch("moai_adk.cli.commands.update._generate_manual_merge_guide") as mock_guide,
            ):

                mock_backup_instance = Mock()
                mock_backup_instance.create_full_project_backup.return_value = Path.cwd() / "backup"
                mock_backup.return_value = mock_backup_instance

                mock_guide.return_value = Path.cwd() / ".moai" / "guides" / "merge-guide.md"

                result = runner.invoke(update, ["--manual"])
                assert result.exit_code == 0
                assert "Manual merge mode" in result.output

    def test_update_upgrade_needed(self, tmp_path):
        """Test update when package upgrade is needed."""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            moai_dir = Path(".moai")
            moai_dir.mkdir()

            with (
                patch("moai_adk.cli.commands.update._get_current_version", return_value="0.6.0"),
                patch("moai_adk.cli.commands.update._get_latest_version", return_value="0.7.0"),
                patch(
                    "moai_adk.cli.commands.update._detect_tool_installer",
                    return_value=["uv", "tool", "upgrade", "moai-adk"],
                ),
                patch("moai_adk.cli.commands.update._execute_upgrade", return_value=True),
                patch("click.confirm", return_value=True),
            ):

                result = runner.invoke(update)
                assert result.exit_code == 0
                assert "Upgrading" in result.output
                assert "Upgrade complete" in result.output

    def test_update_installer_not_found(self, tmp_path):
        """Test update when no installer is detected."""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            moai_dir = Path(".moai")
            moai_dir.mkdir()

            with (
                patch("moai_adk.cli.commands.update._get_current_version", return_value="0.6.0"),
                patch("moai_adk.cli.commands.update._get_latest_version", return_value="0.7.0"),
                patch("moai_adk.cli.commands.update._detect_tool_installer", return_value=None),
                patch("click.confirm", return_value=True),
            ):

                result = runner.invoke(update)
                assert result.exit_code != 0
                assert "Cannot detect package installer" in result.output

    def test_update_upgrade_timeout(self, tmp_path):
        """Test update when upgrade times out."""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            moai_dir = Path(".moai")
            moai_dir.mkdir()

            with (
                patch("moai_adk.cli.commands.update._get_current_version", return_value="0.6.0"),
                patch("moai_adk.cli.commands.update._get_latest_version", return_value="0.7.0"),
                patch(
                    "moai_adk.cli.commands.update._detect_tool_installer",
                    return_value=["uv", "tool", "upgrade", "moai-adk"],
                ),
                patch(
                    "moai_adk.cli.commands.update._execute_upgrade",
                    side_effect=subprocess.TimeoutExpired(cmd="uv", timeout=60),
                ),
                patch("click.confirm", return_value=True),
            ):

                result = runner.invoke(update)
                assert result.exit_code != 0
                assert "timed out" in result.output


# End of test file
