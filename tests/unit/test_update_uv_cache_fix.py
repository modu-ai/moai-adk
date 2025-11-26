"""Unit tests for UV cache fix functionality in update command.

SPEC-UPDATE-CACHE-FIX-001: UV cache auto-refresh on stale PyPI metadata
"""

from __future__ import annotations

import subprocess
import sys
from unittest.mock import Mock


def test_detect_stale_cache_true():
    """
    Test stale cache detection returns True when cache is outdated.

    GIVEN: upgrade output shows "Nothing to upgrade"
      AND: current version < latest version
    WHEN: _detect_stale_cache() is called
    THEN: return True (cache is stale)

    """
    from moai_adk.cli.commands.update import _detect_stale_cache

    # Scenario 1: Minor version difference
    result = _detect_stale_cache(upgrade_output="Nothing to upgrade", current_version="0.8.3", latest_version="0.9.0")
    assert result is True

    # Scenario 2: Patch version difference
    result = _detect_stale_cache(upgrade_output="Nothing to upgrade", current_version="0.8.3", latest_version="0.8.4")
    assert result is True

    # Scenario 3: Major version difference
    result = _detect_stale_cache(upgrade_output="Nothing to upgrade", current_version="0.9.0", latest_version="1.0.0")
    assert result is True


def test_detect_stale_cache_false():
    """
    Test stale cache detection returns False when cache is up-to-date.

    GIVEN: upgrade output OR version equality
    WHEN: _detect_stale_cache() is called
    THEN: return False (cache is not stale)

    """
    from moai_adk.cli.commands.update import _detect_stale_cache

    # Scenario 1: Already up to date (same version)
    result = _detect_stale_cache(upgrade_output="Nothing to upgrade", current_version="0.9.0", latest_version="0.9.0")
    assert result is False

    # Scenario 2: Upgrade succeeded (different message)
    result = _detect_stale_cache(
        upgrade_output="Successfully updated moai-adk 0.8.3 -> 0.9.0", current_version="0.8.3", latest_version="0.9.0"
    )
    assert result is False

    # Scenario 3: Empty output
    result = _detect_stale_cache(upgrade_output="", current_version="0.8.3", latest_version="0.9.0")
    assert result is False

    # Scenario 4: Current version is newer (dev version)
    result = _detect_stale_cache(upgrade_output="Nothing to upgrade", current_version="0.9.1", latest_version="0.9.0")
    assert result is False

    # Scenario 5: Invalid version string (graceful degradation)
    result = _detect_stale_cache(
        upgrade_output="Nothing to upgrade", current_version="invalid-version", latest_version="0.9.0"
    )
    assert result is False


def test_clear_cache_success(monkeypatch):
    """
    Test cache clearing succeeds when subprocess returns 0.

    GIVEN: subprocess.run returns returncode=0
    WHEN: _clear_uv_package_cache() is called
    THEN: return True and log success at DEBUG level

    """
    from moai_adk.cli.commands.update import _clear_uv_package_cache

    # Mock subprocess.run
    mock_result = Mock()
    mock_result.returncode = 0
    mock_result.stderr = ""

    def mock_run(*args, **kwargs):
        return mock_result

    monkeypatch.setattr("subprocess.run", mock_run)

    # Execute
    result = _clear_uv_package_cache("moai-adk")

    # Verify
    assert result is True


def test_clear_cache_failure(monkeypatch):
    """
    Test cache clearing fails when subprocess returns non-zero.

    GIVEN: subprocess.run returns returncode != 0
    WHEN: _clear_uv_package_cache() is called
    THEN: return False and log warning

    """
    from moai_adk.cli.commands.update import _clear_uv_package_cache

    # Scenario 1: Non-zero return code
    mock_result = Mock()
    mock_result.returncode = 1
    mock_result.stderr = "Permission denied"

    monkeypatch.setattr("subprocess.run", lambda *args, **kwargs: mock_result)

    result = _clear_uv_package_cache("moai-adk")
    assert result is False

    # Scenario 2: Timeout expired
    def mock_timeout(*args, **kwargs):
        raise subprocess.TimeoutExpired(cmd=["uv", "cache", "clean", "moai-adk"], timeout=10)

    monkeypatch.setattr("subprocess.run", mock_timeout)
    result = _clear_uv_package_cache("moai-adk")
    assert result is False

    # Scenario 3: FileNotFoundError (uv not installed)
    def mock_not_found(*args, **kwargs):
        raise FileNotFoundError("uv command not found")

    monkeypatch.setattr("subprocess.run", mock_not_found)
    result = _clear_uv_package_cache("moai-adk")
    assert result is False

    # Scenario 4: Generic exception
    def mock_error(*args, **kwargs):
        raise Exception("Unexpected error")

    monkeypatch.setattr("subprocess.run", mock_error)
    result = _clear_uv_package_cache("moai-adk")
    assert result is False


def test_upgrade_with_retry_stale_cache(monkeypatch):
    """
    Test upgrade with retry logic when stale cache detected.

    GIVEN: first upgrade shows "Nothing to upgrade" but newer version exists
      AND: cache clear succeeds
    WHEN: _execute_upgrade_with_retry() is called
    THEN: cache is cleared and upgrade is retried
      AND: return True if second attempt succeeds

    """
    from moai_adk.cli.commands.update import _execute_upgrade_with_retry

    # Track subprocess.run calls
    call_count = [0]

    # First call: "Nothing to upgrade" (stale cache)
    first_result = Mock()
    first_result.returncode = 0
    first_result.stdout = "Nothing to upgrade"
    first_result.stderr = ""

    # Second call: Successful upgrade after cache clear
    second_result = Mock()
    second_result.returncode = 0
    second_result.stdout = "Updated moai-adk 0.8.3 -> 0.9.1"
    second_result.stderr = ""

    def mock_run(*args, **kwargs):
        call_count[0] += 1
        if call_count[0] == 1:
            return first_result
        return second_result

    monkeypatch.setattr("subprocess.run", mock_run)

    # Mock helper functions - access module from sys.modules
    update_module = sys.modules["moai_adk.cli.commands.update"]
    monkeypatch.setattr(update_module, "_get_current_version", lambda: "0.8.3")
    monkeypatch.setattr(update_module, "_get_latest_version", lambda: "0.9.1")
    monkeypatch.setattr(update_module, "_clear_uv_package_cache", lambda x: True)

    # Execute
    result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])

    # Verify
    assert result is True
    assert call_count[0] == 2  # First attempt + retry


def test_upgrade_without_retry_fresh_cache(monkeypatch):
    """
    Test upgrade succeeds without retry when cache is fresh.

    GIVEN: upgrade succeeds on first attempt
    WHEN: _execute_upgrade_with_retry() is called
    THEN: return True without retry

    """
    from moai_adk.cli.commands.update import _execute_upgrade_with_retry

    # Track subprocess.run calls
    call_count = [0]

    # Mock subprocess.run
    mock_result = Mock()
    mock_result.returncode = 0
    mock_result.stdout = "Updated moai-adk 0.8.3 -> 0.9.0"
    mock_result.stderr = ""

    def mock_run(*args, **kwargs):
        call_count[0] += 1
        return mock_result

    monkeypatch.setattr("subprocess.run", mock_run)

    # Execute
    result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])

    # Verify
    assert result is True
    assert call_count[0] == 1  # No retry needed


def test_upgrade_fails_after_retry(monkeypatch):
    """
    Test upgrade fails after retry when second attempt also fails.

    GIVEN: first upgrade shows "Nothing to upgrade"
      AND: cache clear succeeds
      AND: second upgrade attempt fails
    WHEN: _execute_upgrade_with_retry() is called
    THEN: return False after retry

    """
    from moai_adk.cli.commands.update import _execute_upgrade_with_retry

    # Track subprocess.run calls
    call_count = [0]

    # First call: "Nothing to upgrade"
    first_result = Mock()
    first_result.returncode = 0
    first_result.stdout = "Nothing to upgrade"

    # Second call: Failed upgrade
    second_result = Mock()
    second_result.returncode = 1
    second_result.stdout = ""
    second_result.stderr = "Network error"

    def mock_run(*args, **kwargs):
        call_count[0] += 1
        if call_count[0] == 1:
            return first_result
        return second_result

    monkeypatch.setattr("subprocess.run", mock_run)

    # Mock helper functions - access module from sys.modules
    update_module = sys.modules["moai_adk.cli.commands.update"]
    monkeypatch.setattr(update_module, "_get_current_version", lambda: "0.8.3")
    monkeypatch.setattr(update_module, "_get_latest_version", lambda: "0.9.1")
    monkeypatch.setattr(update_module, "_clear_uv_package_cache", lambda x: True)

    # Execute
    result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])

    # Verify
    assert result is False
    assert call_count[0] == 2


def test_upgrade_cache_clear_fails(monkeypatch):
    """
    Test upgrade when cache clear operation fails.

    GIVEN: first upgrade shows "Nothing to upgrade"
      AND: cache clear fails
    WHEN: _execute_upgrade_with_retry() is called
    THEN: return False without retry

    """
    from moai_adk.cli.commands.update import _execute_upgrade_with_retry

    # Track subprocess.run calls
    call_count = [0]

    # Mock subprocess.run
    first_result = Mock()
    first_result.returncode = 0
    first_result.stdout = "Nothing to upgrade"

    def mock_run(*args, **kwargs):
        call_count[0] += 1
        return first_result

    monkeypatch.setattr("subprocess.run", mock_run)

    # Mock helper functions - access module from sys.modules
    update_module = sys.modules["moai_adk.cli.commands.update"]
    monkeypatch.setattr(update_module, "_get_current_version", lambda: "0.8.3")
    monkeypatch.setattr(update_module, "_get_latest_version", lambda: "0.9.1")
    monkeypatch.setattr(update_module, "_clear_uv_package_cache", lambda x: False)  # Cache clear fails

    # Execute
    result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])

    # Verify
    assert result is False
    assert call_count[0] == 1  # No retry when cache clear fails
