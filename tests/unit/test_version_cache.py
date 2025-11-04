# @TEST:CACHE-001, @TEST:CACHE-002, @TEST:CACHE-003, @TEST:CACHE-004, @TEST:CACHE-005
"""
VersionCache 시스템 테스트.

TTL 기반 버전 정보 캐싱 기능을 테스트합니다.

SPEC: SPEC-UPDATE-ENHANCE-001 - SessionStart 버전 체크 시스템 강화
Phase 1: Cache System Implementation
"""

import importlib.util
import json
import sys
import time
from datetime import datetime, timedelta
from pathlib import Path

import pytest


def _load_version_cache_module(module_name: str = "version_cache_module"):
    """Dynamically load core/version_cache.py as a fresh module."""
    repo_root = Path(__file__).resolve().parents[2]

    # First try to load from .claude/hooks/alfred/core (development)
    version_cache_path = repo_root / ".claude" / "hooks" / "alfred" / "core" / "version_cache.py"

    if not version_cache_path.exists():
        # If not found, try templates directory (for testing after installation)
        templates_core_dir = (
            repo_root / "src" / "moai_adk" / "templates" /
            ".claude" / "hooks" / "alfred" / "core"
        )
        version_cache_path = templates_core_dir / "version_cache.py"

    spec = importlib.util.spec_from_file_location(module_name, version_cache_path)
    if spec is None or spec.loader is None:
        raise ImportError(f"Cannot load version_cache module from {version_cache_path}")

    module = importlib.util.module_from_spec(spec)
    sys.modules[module_name] = module
    spec.loader.exec_module(module)
    return module


class TestVersionCache:
    """VersionCache TTL 기반 캐싱 테스트."""

    @pytest.fixture
    def cache_dir(self, tmp_path):
        """임시 캐시 디렉토리 생성."""
        cache_path = tmp_path / ".moai" / "cache"
        cache_path.mkdir(parents=True, exist_ok=True)
        return cache_path

    @pytest.fixture
    def version_cache_module(self):
        """Version cache module with dynamic loading."""
        module_name = f"version_cache_module_{time.time_ns()}"
        module = _load_version_cache_module(module_name=module_name)
        yield module
        sys.modules.pop(module_name, None)

    @pytest.fixture
    def version_cache(self, cache_dir, version_cache_module):
        """VersionCache 인스턴스."""
        version_cache_class = version_cache_module.VersionCache
        return version_cache_class(cache_dir=cache_dir, ttl_hours=24)

    @pytest.fixture
    def sample_version_info(self):
        """샘플 버전 정보."""
        return {
            "last_check": datetime.now().isoformat(),
            "current_version": "0.8.1",
            "latest_version": "0.9.0",
            "update_available": True,
            "upgrade_command": "uv tool upgrade moai-adk",
            "release_notes_url": "https://github.com/modu-ai/moai-adk/releases/tag/v0.9.0",
            "is_major_update": False
        }

    # @TEST:CACHE-001: Cache file creation and validation
    def test_cache_file_created_on_save(self, version_cache, cache_dir, sample_version_info):
        """@TEST:CACHE-001: Cache file is created when saving"""
        # Given: A VersionCache instance with empty cache directory
        cache_file = cache_dir / "version-check.json"
        assert not cache_file.exists(), "Cache file should not exist initially"

        # When: save() is called with version info
        result = version_cache.save(sample_version_info)

        # Then: Cache file should be created at .moai/cache/version-check.json
        assert result is True, "save() should return True on success"
        assert cache_file.exists(), "Cache file should be created"

        # Verify file content
        with open(cache_file) as f:
            cached_data = json.load(f)
        assert "last_check" in cached_data
        assert cached_data["current_version"] == "0.8.1"
        assert cached_data["latest_version"] == "0.9.0"

    # @TEST:CACHE-002: TTL validation - within 24 hours
    def test_cache_validity_within_24_hours(self, version_cache, cache_dir, sample_version_info):
        """@TEST:CACHE-002: Cache is valid within TTL"""
        # Given: Cache was created 6 hours ago
        six_hours_ago = datetime.now() - timedelta(hours=6)
        sample_version_info["last_check"] = six_hours_ago.isoformat()
        version_cache.save(sample_version_info)

        # When: is_valid() is called
        result = version_cache.is_valid()

        # Then: Should return True
        assert result is True, "Cache should be valid within 24 hours (6 hours old)"

    # @TEST:CACHE-003: TTL expiration - after 24 hours
    def test_cache_expired_after_24_hours(self, version_cache, cache_dir, sample_version_info):
        """@TEST:CACHE-003: Cache expires after TTL"""
        # Given: Cache was created 48 hours ago
        forty_eight_hours_ago = datetime.now() - timedelta(hours=48)
        sample_version_info["last_check"] = forty_eight_hours_ago.isoformat()
        version_cache.save(sample_version_info)

        # When: is_valid() is called
        result = version_cache.is_valid()

        # Then: Should return False
        assert result is False, "Cache should be expired after 48 hours (> 24 hour TTL)"

    # @TEST:CACHE-004: Load valid cache
    def test_load_returns_valid_cache(self, version_cache, cache_dir, sample_version_info):
        """@TEST:CACHE-004: Load returns cached data if valid"""
        # Given: Valid cache file exists (1 hour old)
        one_hour_ago = datetime.now() - timedelta(hours=1)
        sample_version_info["last_check"] = one_hour_ago.isoformat()
        version_cache.save(sample_version_info)

        # When: load() is called
        loaded_data = version_cache.load()

        # Then: Should return cached version info dict
        assert loaded_data is not None, "load() should return data when cache is valid"
        assert loaded_data["current_version"] == "0.8.1"
        assert loaded_data["latest_version"] == "0.9.0"
        assert loaded_data["update_available"] is True

    # @TEST:CACHE-005: Error handling - corrupted file
    def test_cache_handles_corrupted_file(self, version_cache, cache_dir):
        """@TEST:CACHE-005: Cache handles file corruption gracefully"""
        # Given: Corrupted cache file exists
        cache_file = cache_dir / "version-check.json"
        cache_file.write_text("{ invalid json content !!!")

        # When: load() is called
        loaded_data = version_cache.load()

        # Then: Should return None (graceful degradation)
        assert loaded_data is None, "load() should return None for corrupted cache file"

        # When: is_valid() is called on corrupted cache
        is_valid_result = version_cache.is_valid()

        # Then: Should return False
        assert is_valid_result is False, "is_valid() should return False for corrupted cache"

    # @TEST:CACHE-006: Additional - Cache expiry on boundary (exactly 24 hours)
    def test_cache_expired_at_ttl_boundary(self, version_cache, cache_dir, sample_version_info):
        """Cache should be expired at exactly TTL boundary (24 hours)"""
        # Given: Cache was created exactly 24 hours ago
        exactly_24_hours_ago = datetime.now() - timedelta(hours=24, seconds=1)
        sample_version_info["last_check"] = exactly_24_hours_ago.isoformat()
        version_cache.save(sample_version_info)

        # When: is_valid() is called
        result = version_cache.is_valid()

        # Then: Should return False (strictly greater than TTL)
        assert result is False, "Cache should expire at exactly 24 hours"

    # @TEST:CACHE-007: Additional - Load returns None when expired
    def test_load_returns_none_when_expired(self, version_cache, cache_dir, sample_version_info):
        """load() should return None when cache is expired"""
        # Given: Expired cache file exists (48 hours old)
        forty_eight_hours_ago = datetime.now() - timedelta(hours=48)
        sample_version_info["last_check"] = forty_eight_hours_ago.isoformat()
        version_cache.save(sample_version_info)

        # When: load() is called
        loaded_data = version_cache.load()

        # Then: Should return None
        assert loaded_data is None, "load() should return None when cache is expired"

    # @TEST:CACHE-008: Additional - Clear cache functionality
    def test_clear_removes_cache_file(self, version_cache, cache_dir, sample_version_info):
        """clear() should remove cache file"""
        # Given: Cache file exists
        version_cache.save(sample_version_info)
        cache_file = cache_dir / "version-check.json"
        assert cache_file.exists(), "Cache file should exist before clearing"

        # When: clear() is called
        result = version_cache.clear()

        # Then: Cache file should be removed
        assert result is True, "clear() should return True on success"
        assert not cache_file.exists(), "Cache file should be removed"

    # @TEST:CACHE-009: Additional - Get cache age
    def test_get_age_hours(self, version_cache, cache_dir, sample_version_info):
        """get_age_hours() should return correct age in hours"""
        # Given: Cache created 12 hours ago
        twelve_hours_ago = datetime.now() - timedelta(hours=12)
        sample_version_info["last_check"] = twelve_hours_ago.isoformat()
        version_cache.save(sample_version_info)

        # When: get_age_hours() is called
        age = version_cache.get_age_hours()

        # Then: Should return approximately 12 hours (with tolerance)
        assert 11.5 < age < 12.5, f"Age should be ~12 hours, got {age}"
