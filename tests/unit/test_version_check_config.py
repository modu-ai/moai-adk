"""Tests for Phase 4: Config Integration for Version Check

This module tests configuration-based version checking with:
- Config reading and defaults
- Frequency-based cache TTL
- Disabled version check mode
- Backward compatibility with old configs

"""

import importlib.util
import io
import json
import sys
import time
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest


def _load_project_module(module_name: str = "project_module"):
    """Dynamically load core/project.py as a fresh module."""
    repo_root = Path(__file__).resolve().parents[2]

    # Add hooks directory to sys.path
    hooks_dir = repo_root / "src" / "moai_adk" / "templates" / ".claude" / "hooks" / "alfred"
    if str(hooks_dir) not in sys.path:
        sys.path.insert(0, str(hooks_dir))

    module_path = hooks_dir / "core" / "project.py"

    spec = importlib.util.spec_from_file_location(module_name, module_path)
    module = importlib.util.module_from_spec(spec)
    sys.modules[module_name] = module
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


@pytest.fixture
def project_module():
    """Provide a clean project module instance per test."""
    module_name = f"project_module_{time.time_ns()}"
    module = _load_project_module(module_name=module_name)
    yield module
    sys.modules.pop(module_name, None)


# ========================================
# Phase 4: Config Integration Tests
# ========================================


def test_get_version_check_config_defaults(tmp_path: Path, project_module):
    """Returns defaults if config missing

    Given: No config file exists
    When: get_version_check_config() is called
    Then: Should return defaults (enabled=True, frequency="daily")

    """
    project_dir = tmp_path / "project"
    project_dir.mkdir()

    result = project_module.get_version_check_config(str(project_dir))

    # Assert: Default values returned
    assert result["enabled"] is True
    assert result["frequency"] == "daily"
    assert result["cache_ttl_hours"] == 24


def test_get_version_check_config_custom(tmp_path: Path, project_module):
    """Reads custom frequency from config

    Given: Config file has frequency="weekly"
    When: get_version_check_config() is called
    Then: Should return frequency="weekly" with TTL=168
    """
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir()

    # Create config with custom frequency
    config_path = moai_dir / "config.json"
    config_data = {
        "moai": {"update_check_frequency": "weekly", "version_check": {"enabled": True, "cache_ttl_hours": 168}}
    }
    config_path.write_text(json.dumps(config_data), encoding="utf-8")

    result = project_module.get_version_check_config(str(project_dir))

    # Assert: Custom values returned
    assert result["enabled"] is True
    assert result["frequency"] == "weekly"
    assert result["cache_ttl_hours"] == 168


def test_version_check_disabled(tmp_path: Path, project_module):
    """Skips check when disabled

    Given: version_check.enabled=False in config
    When: get_package_version_info() is called
    Then: Should return only current version (no PyPI query)
    """
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir()

    # Create config with disabled version check
    config_path = moai_dir / "config.json"
    config_data = {"moai": {"version_check": {"enabled": False}}}
    config_path.write_text(json.dumps(config_data), encoding="utf-8")

    # Mock urllib to ensure it's not called
    with patch("urllib.request.urlopen") as mock_urlopen:
        mock_urlopen.side_effect = Exception("Should not call PyPI when disabled!")

        result = project_module.get_package_version_info(cwd=str(project_dir))

        # Assert: Should return current version
        assert result["current"] != "unknown"

        # Assert: Latest should be "unknown" (no PyPI query)
        assert result["latest"] == "unknown"

        # Assert: No update available
        assert result["update_available"] is False

        # Assert: urllib should NOT have been called
        mock_urlopen.assert_not_called()


def test_cache_ttl_by_frequency(tmp_path: Path, project_module):
    """Uses correct TTL for frequency

    Given: Different frequency settings
    When: get_version_check_config() is called
    Then: Should use appropriate TTL (always=0, daily=24, weekly=168, never=inf)
    """
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir()
    config_path = moai_dir / "config.json"

    # Test frequency: always
    config_data = {"moai": {"update_check_frequency": "always"}}
    config_path.write_text(json.dumps(config_data), encoding="utf-8")
    result = project_module.get_version_check_config(str(project_dir))
    assert result["cache_ttl_hours"] == 0

    # Test frequency: daily
    config_data["moai"]["update_check_frequency"] = "daily"
    config_path.write_text(json.dumps(config_data), encoding="utf-8")
    result = project_module.get_version_check_config(str(project_dir))
    assert result["cache_ttl_hours"] == 24

    # Test frequency: weekly
    config_data["moai"]["update_check_frequency"] = "weekly"
    config_path.write_text(json.dumps(config_data), encoding="utf-8")
    result = project_module.get_version_check_config(str(project_dir))
    assert result["cache_ttl_hours"] == 168

    # Test frequency: never
    config_data["moai"]["update_check_frequency"] = "never"
    config_path.write_text(json.dumps(config_data), encoding="utf-8")
    result = project_module.get_version_check_config(str(project_dir))
    assert result["cache_ttl_hours"] == float("inf")


def test_session_start_respects_config(tmp_path: Path, project_module):
    """SessionStart respects version check config

    Given: version_check.enabled=False
    When: get_package_version_info() is called (simulating SessionStart)
    Then: Should not attempt PyPI query
    """
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir()

    # Create config with disabled version check
    config_path = moai_dir / "config.json"
    config_data = {"moai": {"version_check": {"enabled": False}}}
    config_path.write_text(json.dumps(config_data), encoding="utf-8")

    # Mock network check to ensure it's not even attempted
    with patch.object(project_module, "is_network_available", return_value=True):
        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_urlopen.side_effect = Exception("Should not reach here!")

            result = project_module.get_package_version_info(cwd=str(project_dir))

            # Assert: Should return early without checking network
            assert result["current"] != "unknown"
            assert result["latest"] == "unknown"

            # Network check may or may not be called, but urlopen definitely should not
            mock_urlopen.assert_not_called()


def test_version_check_config_backward_compatible(tmp_path: Path, project_module):
    """Works with old config format

    Given: Old config.json without moai.version_check section
    When: get_version_check_config() is called
    Then: Should provide safe defaults
    """
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir()

    # Create old-style config (no moai.version_check section)
    config_path = moai_dir / "config.json"
    config_data = {"project": {"name": "MoAI-ADK", "language": "python"}}
    config_path.write_text(json.dumps(config_data), encoding="utf-8")

    result = project_module.get_version_check_config(str(project_dir))

    # Assert: Should return safe defaults
    assert result["enabled"] is True
    assert result["frequency"] == "daily"
    assert result["cache_ttl_hours"] == 24


def test_full_version_check_workflow_with_config(tmp_path: Path, project_module):
    """Complete workflow end-to-end

    Given: All components (cache, network, config, handler)
    When: Version check runs with various configs
    Then: Should behave correctly (cache, skip, warn, etc.)
    """
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir()
    cache_dir = moai_dir / "cache"
    cache_dir.mkdir()

    # Test 1: Enabled with daily frequency (should use cache)
    config_path = moai_dir / "config.json"
    config_data = {
        "moai": {"update_check_frequency": "daily", "version_check": {"enabled": True, "cache_ttl_hours": 24}}
    }
    config_path.write_text(json.dumps(config_data), encoding="utf-8")

    # Mock PyPI response
    pypi_data = {
        "info": {"version": "0.9.0", "project_urls": {"Changelog": "https://github.com/modu-ai/moai-adk/releases"}}
    }

    mock_response = MagicMock()
    mock_response.__enter__ = MagicMock(return_value=io.BytesIO(json.dumps(pypi_data).encode()))
    mock_response.__exit__ = MagicMock(return_value=False)

    with patch("urllib.request.urlopen", return_value=mock_response):
        with patch.object(project_module, "is_network_available", return_value=True):
            result = project_module.get_package_version_info(cwd=str(project_dir))

            # Assert: Should have version info
            assert result["current"] != "unknown"
            assert result["latest"] != "unknown"

    # Test 2: Disabled - should not query PyPI
    config_data["moai"]["version_check"]["enabled"] = False
    config_path.write_text(json.dumps(config_data), encoding="utf-8")

    # Clear cache before second test
    cache_file = cache_dir / "version-check.json"
    if cache_file.exists():
        cache_file.unlink()

    with patch("urllib.request.urlopen") as mock_urlopen:
        mock_urlopen.side_effect = Exception("Should not call PyPI!")

        result = project_module.get_package_version_info(cwd=str(project_dir))

        # Assert: Should return current only
        assert result["current"] != "unknown"
        assert result["latest"] == "unknown"
        mock_urlopen.assert_not_called()


def test_backward_compatibility_old_cache(tmp_path: Path, project_module):
    """Works with old cache format

    Given: Cache created by old version (no new fields)
    When: get_package_version_info() is called
    Then: Should handle gracefully and add missing fields
    """
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir()
    cache_dir = moai_dir / "cache"
    cache_dir.mkdir()

    # Create old-style cache (no is_major_update, release_notes_url)
    import time

    cache_path = cache_dir / "version-check.json"
    old_cache = {"latest": "0.9.0", "timestamp": time.time(), "update_available": True}
    cache_path.write_text(json.dumps(old_cache), encoding="utf-8")

    # Create config enabling version check
    config_path = moai_dir / "config.json"
    config_data = {"moai": {"version_check": {"enabled": True}}}
    config_path.write_text(json.dumps(config_data), encoding="utf-8")

    result = project_module.get_package_version_info(cwd=str(project_dir))

    # Assert: Should have all expected fields (with defaults for missing)
    assert result["current"] != "unknown"
    assert result["latest"] != "unknown"
    assert "update_available" in result
    assert "is_major_update" in result  # Should be added even if missing from cache
    assert "release_notes_url" in result  # Should be present
