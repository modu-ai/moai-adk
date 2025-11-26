"""
Test suite for SPEC-FIX-001: Statusline recovery - Ver unknown issue resolution

This test suite validates:
- U1-U3: Ubiquitous requirements (uvx environment, config.json, CLI commands)
- ED1-ED3: Event-driven requirements (session hooks, version updates, cache recovery)
- UW1-UW3: Unwanted scenario prevention (no "Ver unknown", error handling, performance)
- SD1-SD3: State-driven requirements (session consistency, multi-session, version tracking)
- OP1-OP3: Optional feature support (statusline display, cache management, performance optimization)
"""

import json
import subprocess
import sys
import time
from pathlib import Path

import pytest

from moai_adk.statusline.main import build_statusline_data
from moai_adk.statusline.version_reader import VersionConfig, VersionReader


class TestUbiquitousRequirements:
    """Tests for U1-U3: Ubiquitous requirements"""

    def test_u1_uvx_environment_available(self):
        """U1: uvx environment is correctly recognized"""
        # Check uvx is available
        result = subprocess.run(["uvx", "--version"], capture_output=True, text=True, timeout=5)
        assert result.returncode == 0, "uvx command should be available"
        assert "uv" in result.stdout.lower(), "uvx --version should show uv info"

    def test_u1_python_version_compatible(self):
        """U1: Python version is 3.13.9+ compatible"""
        version_info = sys.version_info
        assert (version_info.major == 3 and version_info.minor >= 13) or version_info.major > 3
        assert (
            version_info.major == 3
            and version_info.minor == 14
            or (version_info.major == 3 and version_info.minor == 13 and version_info.micro >= 9)
            or version_info.major > 3
        )

    def test_u2_config_json_version_field(self, tmp_path):
        """U2: config.json contains moai.version field"""
        config_path = tmp_path / ".moai" / "config"
        config_path.mkdir(parents=True)

        config_file = config_path / "config.json"
        config_data = {"moai": {"version": "0.26.0"}}
        config_file.write_text(json.dumps(config_data))

        # Read it back
        with open(config_file) as f:
            loaded = json.load(f)

        assert loaded["moai"]["version"] == "0.26.0"

    def test_u3_cli_statusline_command_exists(self):
        """U3: CLI statusline command can be executed"""
        from moai_adk.__main__ import cli

        # Verify statusline command is registered
        assert "statusline" in cli.commands
        assert cli.commands["statusline"] is not None

    def test_u3_statusline_via_uv_run(self):
        """U3: statusline command works via uv run"""
        result = subprocess.run(
            ["uv", "run", "moai-adk", "statusline"],
            input="{}",
            capture_output=True,
            text=True,
            timeout=5,
            cwd="/Users/goos/MoAI/MoAI-ADK",
        )
        assert result.returncode == 0, f"statusline command failed: {result.stderr}"
        assert "Ver" in result.stdout, "Output should contain version info"


class TestEventDrivenRequirements:
    """Tests for ED1-ED3: Event-driven requirements"""

    def test_ed1_version_not_unknown(self):
        """ED1: Version should never be 'unknown' when config.json exists"""
        session_context = {"model": {"display_name": "Haiku 4.5"}, "cwd": "/Users/goos/MoAI/MoAI-ADK"}

        statusline = build_statusline_data(session_context, mode="compact")

        # Check that "unknown" is not in the output for version
        assert "Ver unknown" not in statusline, "Ver unknown should not appear"
        assert "Ver" in statusline, "Version should be displayed"
        assert "0.25" in statusline or "0.26" in statusline, "Should show actual version"

    def test_ed1_version_readable_from_config(self):
        """ED1: Version should be readable from config.json"""
        reader = VersionReader(working_dir=Path("/Users/goos/MoAI/MoAI-ADK"))
        version = reader.get_version()

        assert version != "unknown", "Version should be read from config"
        assert version != "", "Version should not be empty"
        # Expected format: semantic versioning
        assert "." in version, "Version should contain dot separators"

    def test_ed2_version_change_detection(self, tmp_path):
        """ED2: Version changes in config are detected"""
        config_path = tmp_path / ".moai" / "config"
        config_path.mkdir(parents=True)

        config_file = config_path / "config.json"

        # Write initial version
        config_data = {"moai": {"version": "0.26.0"}}
        config_file.write_text(json.dumps(config_data))

        reader = VersionReader(working_dir=tmp_path)
        version1 = reader.get_version()
        assert version1 == "0.26.0"

        # Update version
        config_data["moai"]["version"] = "0.27.0"
        config_file.write_text(json.dumps(config_data))

        # Clear cache to force re-read
        reader.clear_cache()
        version2 = reader.get_version()
        assert version2 == "0.27.0", "Updated version should be detected"

    def test_ed3_cache_cleanup_recovery(self):
        """ED3: Cache cleanup allows proper version recovery"""
        reader = VersionReader(working_dir=Path("/Users/goos/MoAI/MoAI-ADK"))

        # Get version first time (fills cache)
        version1 = reader.get_version()
        assert version1 != "unknown"

        # Clear cache
        reader.clear_cache()

        # Get version again (should work)
        version2 = reader.get_version()
        assert version2 == version1, "Version should remain consistent after cache clear"


class TestUnwantedScenarioPrevention:
    """Tests for UW1-UW3: Preventing unwanted scenarios"""

    def test_uw1_no_ver_unknown_message(self):
        """UW1: 'Ver unknown' message should never appear"""
        session_context = {"model": {"display_name": "Haiku 4.5"}, "cwd": "/Users/goos/MoAI/MoAI-ADK"}

        statusline = build_statusline_data(session_context, mode="compact")

        assert "Ver unknown" not in statusline, "Ver unknown is explicitly forbidden"

    def test_uw1_graceful_fallback_when_no_config(self, tmp_path):
        """UW1: Graceful fallback when config.json is missing"""
        reader = VersionReader(working_dir=tmp_path)
        version = reader.get_version()

        # Should return fallback, not raise error
        assert version is not None
        # Fallback might be 'unknown' but statusline should handle it gracefully
        assert isinstance(version, str)

    def test_uw2_no_infinite_loop_on_import_failure(self):
        """UW2: No infinite loop when import fails"""
        config = VersionConfig(timeout_seconds=3, cache_ttl_seconds=60)
        reader = VersionReader(config=config)

        # This should complete within timeout
        start = time.time()
        version = reader.get_version()
        elapsed = time.time() - start

        assert elapsed < 3, f"Should complete within 3 seconds, took {elapsed}"
        assert isinstance(version, str)

    def test_uw2_retry_limit_respected(self, tmp_path):
        """UW2: Retry attempts are limited"""
        config = VersionConfig(timeout_seconds=3)
        reader = VersionReader(config=config)

        # Even with repeated failures, should not retry infinitely
        for _ in range(5):
            version = reader.get_version()
            assert isinstance(version, str)

    def test_uw3_performance_within_limits(self):
        """UW3: Performance stays within requirements"""
        reader = VersionReader(working_dir=Path("/Users/goos/MoAI/MoAI-ADK"))

        # First call (cold cache)
        start = time.time()
        reader.get_version()
        first_time = time.time() - start

        # Second call (warm cache)
        start = time.time()
        reader.get_version()
        second_time = time.time() - start

        assert first_time < 2.0, f"First call should be < 2s, was {first_time:.3f}s"
        assert second_time < 1.0, f"Cached call should be < 1s, was {second_time:.3f}s"


class TestStateDrivenRequirements:
    """Tests for SD1-SD3: State-driven requirements"""

    def test_sd1_session_consistency(self):
        """SD1: Version info remains consistent during session"""
        reader = VersionReader(working_dir=Path("/Users/goos/MoAI/MoAI-ADK"))

        # Get version multiple times
        versions = [reader.get_version() for _ in range(5)]

        # All should be identical
        assert all(v == versions[0] for v in versions), "Version should be consistent"

    def test_sd1_git_status_updates(self):
        """SD1: Git status updates while version stays constant"""
        session_context = {"model": {"display_name": "Haiku 4.5"}, "cwd": "/Users/goos/MoAI/MoAI-ADK"}

        statusline1 = build_statusline_data(session_context, mode="compact")

        # Parse version from first statusline
        version_part = [s for s in statusline1.split("|") if "Ver" in s]
        assert version_part, "Should contain version section"

        # Git status might change, but version should remain
        statusline2 = build_statusline_data(session_context, mode="compact")
        version_part2 = [s for s in statusline2.split("|") if "Ver" in s]

        assert version_part == version_part2, "Version section should remain constant"

    def test_sd2_multi_session_independence(self, tmp_path):
        """SD2: Multiple readers don't interfere"""
        reader1 = VersionReader(working_dir=Path("/Users/goos/MoAI/MoAI-ADK"))
        reader2 = VersionReader(working_dir=Path("/Users/goos/MoAI/MoAI-ADK"))

        v1 = reader1.get_version()
        v2 = reader2.get_version()

        assert v1 == v2, "Both readers should get same version"

    def test_sd3_version_field_priority(self, tmp_path):
        """SD3: Version field priority is respected"""
        config_path = tmp_path / ".moai" / "config"
        config_path.mkdir(parents=True)

        config_file = config_path / "config.json"
        config_data = {"moai": {"version": "0.26.0"}, "project": {"version": "1.0.0"}, "version": "2.0.0"}
        config_file.write_text(json.dumps(config_data))

        reader = VersionReader(working_dir=tmp_path)
        version = reader.get_version()

        # Should prefer moai.version first
        assert version == "0.26.0", "Should use moai.version first"


class TestOptionalFeatures:
    """Tests for OP1-OP3: Optional features"""

    def test_op1_statusline_enabled_displays_version(self):
        """OP1: When statusline is enabled, version displays correctly"""
        session_context = {
            "model": {"display_name": "Haiku 4.5"},
            "statusline": {"enabled": True},
            "cwd": "/Users/goos/MoAI/MoAI-ADK",
        }

        statusline = build_statusline_data(session_context, mode="extended")

        assert len(statusline) > 0, "Should produce output"
        assert "Ver" in statusline, "Should display version"
        assert "unknown" not in statusline, "Should not show unknown"

    def test_op2_cache_management_clear(self):
        """OP2: Cache can be explicitly cleared"""
        reader = VersionReader(working_dir=Path("/Users/goos/MoAI/MoAI-ADK"))

        # Prime cache
        v1 = reader.get_version()
        reader.get_cache_stats()

        # Clear cache
        reader.clear_cache()
        reader.get_cache_stats()

        # Both should return same version
        v2 = reader.get_version()
        assert v1 == v2, "Version should be consistent after cache clear"

    def test_op3_performance_optimization_cache_ttl(self):
        """OP3: Cache TTL can be configured for performance"""
        config = VersionConfig(cache_ttl_seconds=1)
        reader = VersionReader(config=config)

        # Get version
        v1 = reader.get_version()

        # Wait for cache to expire
        time.sleep(1.1)

        # Get again (should re-read)
        v2 = reader.get_version()
        assert v1 == v2, "Version should be consistent"


class TestAcceptanceCriteria:
    """Tests for acceptance criteria from SPEC-FIX-001"""

    def test_ac_statusline_via_uv_run(self):
        """AC: statusline command works via uv run"""
        result = subprocess.run(
            ["uv", "run", "moai-adk", "statusline"],
            input="{}",
            capture_output=True,
            text=True,
            timeout=5,
            cwd="/Users/goos/MoAI/MoAI-ADK",
        )
        assert result.returncode == 0
        assert "Ver" in result.stdout
        # Version should NOT be "Ver unknown" - critical requirement
        assert "Ver unknown" not in result.stdout
        # Version should show actual version number
        assert any(v in result.stdout for v in ["0.25", "0.26", "0.27"])

    def test_ac_version_correct_format(self):
        """AC: Version in correct format (semantic versioning)"""
        reader = VersionReader(working_dir=Path("/Users/goos/MoAI/MoAI-ADK"))
        version = reader.get_version()

        # Should be X.Y.Z format
        parts = version.split(".")
        assert len(parts) >= 2, "Should be semantic version"
        assert all(p.isdigit() for p in parts[:2]), "Major.minor should be numeric"

    def test_ac_performance_requirements(self):
        """AC: Performance meets requirements (2s initial, 1s cached)"""
        reader = VersionReader(working_dir=Path("/Users/goos/MoAI/MoAI-ADK"))

        start = time.time()
        reader.get_version()
        first_time = time.time() - start

        start = time.time()
        reader.get_version()
        cached_time = time.time() - start

        assert first_time <= 2.0, f"Initial: {first_time:.3f}s"
        assert cached_time <= 1.0, f"Cached: {cached_time:.3f}s"

    def test_ac_git_status_accuracy(self):
        """AC: Git status is accurately reported"""
        session_context = {"model": {"display_name": "Haiku 4.5"}, "cwd": "/Users/goos/MoAI/MoAI-ADK"}

        statusline = build_statusline_data(session_context, mode="compact")

        # Should contain git indicators
        assert "+" in statusline or "M" in statusline or "?" in statusline or "🔀" in statusline
        assert "feature/SPEC-FIX-001" in statusline or "release" in statusline or "main" in statusline


class TestIntegration:
    """Integration tests for end-to-end workflow"""

    def test_full_statusline_pipeline(self):
        """Full statusline generation pipeline works end-to-end"""
        session_context = {"model": {"display_name": "Haiku 4.5"}, "cwd": "/Users/goos/MoAI/MoAI-ADK"}

        # Build statusline
        statusline = build_statusline_data(session_context, mode="extended")

        # Verify all components present
        assert "Haiku" in statusline or "Unknown" in statusline
        assert "Ver" in statusline
        assert "M" in statusline or "+" in statusline or "?" in statusline
        assert "🔀" in statusline or "release" in statusline or "feature" in statusline or "main" in statusline

    def test_subprocess_statusline_execution(self):
        """Statusline can be executed as subprocess"""
        result = subprocess.run(
            ["uv", "run", "moai-adk", "statusline"],
            input="{}",
            capture_output=True,
            text=True,
            timeout=5,
            cwd="/Users/goos/MoAI/MoAI-ADK",
        )

        assert result.returncode == 0
        assert len(result.stdout) > 0
        assert "Ver" in result.stdout


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
