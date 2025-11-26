"""Tests for .claude/hooks/alfred/core/project.py module

프로젝트 언어 감지, Git 정보, SPEC 카운팅 테스트

NOTE: These tests require fixing the relative import structure in .claude/hooks/alfred/
Currently skipped due to import path issues - requires refactoring the shared/handlers modules
"""

import importlib.util
import json
import sys
import time
from pathlib import Path

import pytest


def _load_project_module(module_name: str = "project_module"):
    """Dynamically load core/project.py as a fresh module."""
    repo_root = Path(__file__).resolve().parents[2]

    # Add hooks directory and shared directory to sys.path for relative imports
    hooks_dir = repo_root / "src" / "moai_adk" / "templates" / ".claude" / "hooks" / "alfred"
    shared_dir = hooks_dir / "shared"

    for path_entry in [str(hooks_dir), str(shared_dir), str(hooks_dir.parent)]:
        if path_entry not in sys.path:
            sys.path.insert(0, path_entry)

    module_path = hooks_dir / "core" / "project.py"

    spec = importlib.util.spec_from_file_location(module_name, module_path)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"Failed to load module from {module_path}")

    module = importlib.util.module_from_spec(spec)
    sys.modules[module_name] = module
    spec.loader.exec_module(module)
    return module


@pytest.fixture
def project_module():
    """Provide a clean project module instance per test."""
    module_name = f"project_module_{time.time_ns()}"
    module = _load_project_module(module_name=module_name)
    yield module
    sys.modules.pop(module_name, None)


def test_get_project_language_prefers_config(monkeypatch: pytest.MonkeyPatch, tmp_path: Path, project_module):
    """get_project_language()가 .moai/config.json을 우선하는지 확인"""
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir()

    config_path = moai_dir / "config.json"
    config_path.write_text(json.dumps({"language": "python"}), encoding="utf-8")

    assert project_module.get_project_language(str(project_dir)) == "python"


def test_get_project_language_falls_back_to_detect(monkeypatch: pytest.MonkeyPatch, tmp_path: Path, project_module):
    """get_project_language()가 config 없을 때 detect_language()로 fallback하는지 확인"""
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir()

    config_path = moai_dir / "config.json"
    config_path.write_text("not-json", encoding="utf-8")

    monkeypatch.setattr(project_module, "detect_language", lambda cwd: "fallback")
    assert project_module.get_project_language(str(project_dir)) == "fallback"

    config_path.unlink()
    assert project_module.get_project_language(str(project_dir)) == "fallback"


def test_detect_language_python(tmp_path: Path, project_module):
    """detect_language()가 Python 프로젝트를 감지하는지 확인"""
    project_dir = tmp_path / "project"
    project_dir.mkdir()

    # Create pyproject.toml
    (project_dir / "pyproject.toml").write_text("[tool.poetry]\nname = 'test'", encoding="utf-8")

    assert project_module.detect_language(str(project_dir)) == "python"


def test_detect_language_typescript_with_tsconfig(tmp_path: Path, project_module):
    """detect_language()가 tsconfig.json이 있는 TypeScript 프로젝트를 감지하는지 확인"""
    project_dir = tmp_path / "project"
    project_dir.mkdir()

    # Create both package.json and tsconfig.json
    (project_dir / "package.json").write_text(json.dumps({"dependencies": {"typescript": "^5.0.0"}}), encoding="utf-8")
    (project_dir / "tsconfig.json").write_text("{}", encoding="utf-8")

    assert project_module.detect_language(str(project_dir)) == "typescript"


def test_detect_language_javascript_without_tsconfig(tmp_path: Path, project_module):
    """detect_language()가 tsconfig.json 없는 JavaScript 프로젝트를 감지하는지 확인"""
    project_dir = tmp_path / "project"
    project_dir.mkdir()

    # Create package.json without tsconfig.json
    (project_dir / "package.json").write_text(json.dumps({"dependencies": {"react": "^18.0.0"}}), encoding="utf-8")

    assert project_module.detect_language(str(project_dir)) == "javascript"


def test_count_specs(tmp_path: Path, project_module):
    """count_specs()가 SPEC 파일들을 정확히 카운트하는지 확인"""
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir()
    specs_dir = moai_dir / "specs"
    specs_dir.mkdir()

    # Create SPEC directories
    for i in range(1, 4):
        spec_dir = specs_dir / f"SPEC-AUTH-{i:03d}"
        spec_dir.mkdir()
        spec_file = spec_dir / "spec.md"
        spec_file.write_text(f"---\nstatus: {'completed' if i == 1 else 'active'}\n---\n", encoding="utf-8")

    result = project_module.count_specs(str(project_dir))

    assert result["total"] == 3
    assert result["completed"] == 1
    # Calculate expected percentage: 1/3 * 100 = 33
    assert result["percentage"] == 33


# # REMOVED_ORPHAN_TEST:OFFLINE-001-05
def test_get_package_version_offline_mode(tmp_path: Path, monkeypatch: pytest.MonkeyPatch, project_module):
    """Returns current version when offline

    Given: Network is unavailable
    When: get_package_version_info() is called
    Then: Should return current version only (no PyPI query)
    """
    from unittest.mock import patch

    # Mock is_network_available to return False
    with patch.object(project_module, "is_network_available", return_value=False):
        # Also mock urllib to ensure it's not called
        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_urlopen.side_effect = Exception("Should not call PyPI in offline mode!")

            result = project_module.get_package_version_info(cwd=str(tmp_path))

            # Assert: Should return current version
            assert result["current"] != "unknown"

            # Assert: Latest should be "unknown" (no PyPI query)
            assert result["latest"] == "unknown"

            # Assert: No update available
            assert result["update_available"] is False

            # Assert: urllib should NOT have been called
            mock_urlopen.assert_not_called()


# ========================================
# Phase 3: Major Version Warning Tests
# ========================================


def test_is_major_version_change_0_to_1(project_module):
    """Detects 0.x → 1.x major version change

    Given: Current version "0.8.1" and latest "1.0.0"
    When: is_major_version_change() is called
    Then: Should return True (major version increased from 0 to 1)

    """
    result = project_module.is_major_version_change("0.8.1", "1.0.0")
    assert result is True


def test_is_major_version_no_change_minor_update(project_module):
    """Minor update not flagged as major version change

    Given: Current version "0.8.1" and latest "0.9.0"
    When: is_major_version_change() is called
    Then: Should return False (major version stayed at 0)
    """
    result = project_module.is_major_version_change("0.8.1", "0.9.0")
    assert result is False


def test_is_major_version_change_1_to_2(project_module):
    """Detects 1.x → 2.x major version change

    Given: Current version "1.2.3" and latest "2.0.0"
    When: is_major_version_change() is called
    Then: Should return True (major version increased from 1 to 2)
    """
    result = project_module.is_major_version_change("1.2.3", "2.0.0")
    assert result is True


def test_is_major_version_change_invalid_versions(project_module):
    """Gracefully handles invalid version strings

    Given: Invalid version strings ("dev", "unknown", non-numeric)
    When: is_major_version_change() is called
    Then: Should return False (no exception, safe fallback)
    """
    # Test various invalid formats
    assert project_module.is_major_version_change("dev", "1.0.0") is False
    assert project_module.is_major_version_change("0.8.1", "unknown") is False
    assert project_module.is_major_version_change("invalid", "also-invalid") is False


def test_get_package_version_includes_release_url(tmp_path: Path, project_module):
    """Release notes URL included in version info

    Given: PyPI response includes project_urls.Changelog
    When: get_package_version_info() is called
    Then: Should include release_notes_url in result dict
    """
    import io
    import json
    from unittest.mock import MagicMock, patch

    # Mock PyPI response with release notes URL
    pypi_data = {
        "info": {"version": "0.9.0", "project_urls": {"Changelog": "https://github.com/modu-ai/moai-adk/releases"}}
    }

    # Create a file-like object for json.load()
    mock_response = MagicMock()
    mock_response.__enter__ = MagicMock(return_value=io.BytesIO(json.dumps(pypi_data).encode()))
    mock_response.__exit__ = MagicMock(return_value=False)

    with patch("urllib.request.urlopen", return_value=mock_response):
        with patch.object(project_module, "is_network_available", return_value=True):
            result = project_module.get_package_version_info(cwd=str(tmp_path))

    # Assert: Should include release_notes_url
    assert "release_notes_url" in result
    assert result["release_notes_url"] is not None
    assert "github.com" in result["release_notes_url"]


def test_get_package_version_includes_major_flag(tmp_path: Path, project_module):
    """Major update flag included in version info

    Given: Latest version is major bump (0.8.1 → 1.0.0)
    When: get_package_version_info() is called
    Then: Should include is_major_update: True
    """
    import io
    import json
    from unittest.mock import MagicMock, patch

    # Mock current version as 0.8.1
    with patch("importlib.metadata.version", return_value="0.8.1"):
        # Mock PyPI response with version 1.0.0
        pypi_data = {
            "info": {"version": "1.0.0", "project_urls": {"Changelog": "https://github.com/modu-ai/moai-adk/releases"}}
        }

        # Create a file-like object for json.load()
        mock_response = MagicMock()
        mock_response.__enter__ = MagicMock(return_value=io.BytesIO(json.dumps(pypi_data).encode()))
        mock_response.__exit__ = MagicMock(return_value=False)

        with patch("urllib.request.urlopen", return_value=mock_response):
            with patch.object(project_module, "is_network_available", return_value=True):
                result = project_module.get_package_version_info(cwd=str(tmp_path))

        # Assert: Should include is_major_update flag
        assert "is_major_update" in result
        assert result["is_major_update"] is True

        # Assert: Update should be available
        assert result["update_available"] is True
        assert result["latest"] == "1.0.0"
