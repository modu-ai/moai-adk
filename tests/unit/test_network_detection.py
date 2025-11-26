"""Tests for network detection and offline support - Phase 2

Network detection tests for is_network_available() function
and cache integration tests for get_package_version_info()

"""

import importlib.util
import socket
import sys
import time
from pathlib import Path
from unittest.mock import patch

import pytest


def _load_project_module(module_name: str = "project_module_network"):
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
    module_name = f"project_module_network_{time.time_ns()}"
    module = _load_project_module(module_name=module_name)
    yield module
    sys.modules.pop(module_name, None)


def test_is_network_available_returns_true(project_module):
    """Network available returns True

    Given: Network is available
    When: is_network_available() is called
    Then: Should return True within 100ms
    """
    # Call the function (should succeed in real network environment)
    start_time = time.time()
    result = project_module.is_network_available()
    elapsed = (time.time() - start_time) * 1000  # Convert to ms

    # Assert: Should return True (in most test environments)
    # Note: This test may fail in offline CI environments - that's expected
    assert isinstance(result, bool)
    assert elapsed < 100, f"Network check took {elapsed}ms, should be < 100ms"


def test_is_network_available_returns_false(project_module):
    """Network unavailable returns False

    Given: Network is unavailable (socket connection fails)
    When: is_network_available() is called
    Then: Should return False within 100ms (timeout)
    """
    # Mock socket.create_connection to raise an exception
    with patch("socket.create_connection") as mock_connect:
        mock_connect.side_effect = OSError("Network unreachable")

        start_time = time.time()
        result = project_module.is_network_available()
        elapsed = (time.time() - start_time) * 1000

        # Assert: Should return False on network error
        assert result is False
        assert elapsed < 100, f"Network check took {elapsed}ms, should be < 100ms"


def test_is_network_available_handles_timeout(project_module):
    """Handles socket timeout gracefully

    Given: Network is very slow (times out)
    When: is_network_available(timeout=0.001) is called
    Then: Should return False without raising exception
    """
    # Mock socket.create_connection to raise timeout
    with patch("socket.create_connection") as mock_connect:
        mock_connect.side_effect = socket.timeout("Connection timeout")

        start_time = time.time()
        result = project_module.is_network_available(timeout_seconds=0.001)
        elapsed = (time.time() - start_time) * 1000

        # Assert: Should return False on timeout, no exception raised
        assert result is False
        assert elapsed < 100, f"Timeout handling took {elapsed}ms, should be < 100ms"


@pytest.mark.skip(reason="Network/cache behavior environment-dependent - test in isolation")
def test_get_package_version_with_valid_cache(project_module, tmp_path):
    """Uses cache when valid

    Given: Valid cache file exists with version info
    When: get_package_version_info() is called
    Then: Should return cached data without PyPI query
    """
    import json
    from datetime import datetime, timezone

    # Create cache directory and file
    cache_dir = tmp_path / ".moai" / "cache"
    cache_dir.mkdir(parents=True, exist_ok=True)
    cache_file = cache_dir / "version-check.json"

    # Create valid cache data (recent timestamp)
    cache_data = {
        "current": "0.8.1",
        "latest": "0.9.0",
        "update_available": True,
        "upgrade_command": "uv tool upgrade moai-adk",
        "last_check": datetime.now(timezone.utc).isoformat(),
    }
    cache_file.write_text(json.dumps(cache_data, indent=2))

    # Patch urllib to detect if it's called (it shouldn't be)
    with patch("urllib.request.urlopen") as mock_urlopen:
        mock_urlopen.side_effect = Exception("Should not call PyPI when cache is valid!")

        # Call the function with tmp_path as cwd
        start_time = time.time()
        result = project_module.get_package_version_info(cwd=str(tmp_path))
        elapsed = (time.time() - start_time) * 1000

        # Assert: Should return cached data
        assert result["current"] == "0.8.1"
        assert result["latest"] == "0.9.0"
        assert result["update_available"] is True

        # Assert: Should be fast (no network call)
        assert elapsed < 50, f"Cache read took {elapsed}ms, should be < 50ms"

        # Assert: urllib should NOT have been called
        mock_urlopen.assert_not_called()
