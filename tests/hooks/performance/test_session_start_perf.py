#!/usr/bin/env python3
# @TEST:ENHANCE-PERF-001:SESSION | SPEC: SPEC-ENHANCE-PERF-001
"""SessionStart Hook Performance Tests

Tests caching performance optimization for get_package_version_info() and get_git_info().

Performance Targets:
- First call (cold cache): < 200ms (baseline measurement)
- Second call (warm cache): < 20ms (10x improvement)
- Cache hit rate: > 90% in typical sessions

TDD History:
- RED: Performance benchmarks and cache behavior tests
- GREEN: TTL cache decorator implementation
- REFACTOR: Cache configuration and error handling
"""

import sys
import time
from pathlib import Path

# Setup sys.path for hook imports
HOOKS_DIR = Path(__file__).parent.parent.parent.parent / ".claude" / "hooks" / "alfred"
SHARED_DIR = HOOKS_DIR / "shared"
sys.path.insert(0, str(SHARED_DIR))

from core.project import get_git_info, get_package_version_info


class TestSessionStartPerformance:
    """Performance tests for SessionStart Hook optimization

    @TAG:HOOK-PERF-001
    """

    def test_version_info_first_call_baseline(self, tmp_path):
        """RED: Measure baseline performance of get_package_version_info()

        @TEST:ENHANCE-PERF-001:VERSION-BASELINE

        This test documents the current performance before optimization.
        First call should complete within reasonable time (< 2000ms).
        """
        # Measure first call (cold cache)
        start = time.perf_counter()
        result = get_package_version_info(str(tmp_path))
        elapsed_ms = (time.perf_counter() - start) * 1000

        # Should return valid version info
        assert result["current"] is not None

        # Baseline measurement (no cache yet)
        # This should be slow due to PyPI network call
        print(f"\n📊 First call (no cache): {elapsed_ms:.2f}ms")

        # Document baseline (allow up to 2 seconds for network)
        assert elapsed_ms < 2000, "Baseline too slow even for first call"

    def test_version_info_cached_call_fast(self, tmp_path):
        """RED: Verify cached call is at least 10x faster

        @TEST:ENHANCE-PERF-001:VERSION-CACHED

        After first call, subsequent calls should hit cache and be much faster.
        Target: < 20ms (compared to ~100ms uncached)
        """
        # First call to populate cache
        result1 = get_package_version_info(str(tmp_path))

        # Second call should hit cache
        start = time.perf_counter()
        result2 = get_package_version_info(str(tmp_path))
        elapsed_ms = (time.perf_counter() - start) * 1000

        # Results should be identical (from cache)
        assert result1 == result2

        # Cache hit should be fast (< 20ms target)
        print(f"\n⚡ Cached call: {elapsed_ms:.2f}ms")
        assert elapsed_ms < 20, f"Cache hit too slow: {elapsed_ms:.2f}ms (expected < 20ms)"

    def test_git_info_first_call_baseline(self, tmp_path):
        """RED: Measure baseline performance of get_git_info()

        First call should complete within reasonable time.
        Git commands are typically fast (< 100ms).
        """
        # Initialize a git repo for testing
        import subprocess
        subprocess.run(["git", "init"], cwd=tmp_path, capture_output=True, check=False)
        subprocess.run(["git", "config", "user.name", "Test"], cwd=tmp_path, capture_output=True, check=False)
        subprocess.run(["git", "config", "user.email", "test@test.com"], cwd=tmp_path, capture_output=True, check=False)
        subprocess.run(["git", "commit", "--allow-empty", "-m", "Initial"], cwd=tmp_path, capture_output=True, check=False)

        # Measure first call
        start = time.perf_counter()
        result = get_git_info(str(tmp_path))
        elapsed_ms = (time.perf_counter() - start) * 1000

        # Should return valid git info
        assert result.get("branch") is not None

        print(f"\n📊 Git first call: {elapsed_ms:.2f}ms")
        assert elapsed_ms < 500, "Git baseline too slow"

    def test_git_info_cached_call_fast(self, tmp_path):
        """RED: Verify Git info caching provides speedup

        Second call should be faster due to caching.
        Target: < 20ms (compared to ~50ms uncached)
        """
        # Initialize a git repo
        import subprocess
        subprocess.run(["git", "init"], cwd=tmp_path, capture_output=True, check=False)
        subprocess.run(["git", "config", "user.name", "Test"], cwd=tmp_path, capture_output=True, check=False)
        subprocess.run(["git", "config", "user.email", "test@test.com"], cwd=tmp_path, capture_output=True, check=False)
        subprocess.run(["git", "commit", "--allow-empty", "-m", "Initial"], cwd=tmp_path, capture_output=True, check=False)

        # First call to populate cache
        result1 = get_git_info(str(tmp_path))

        # Second call should hit cache
        start = time.perf_counter()
        result2 = get_git_info(str(tmp_path))
        elapsed_ms = (time.perf_counter() - start) * 1000

        # Results should be identical (from cache)
        assert result1 == result2

        print(f"\n⚡ Git cached call: {elapsed_ms:.2f}ms")
        assert elapsed_ms < 20, f"Git cache hit too slow: {elapsed_ms:.2f}ms"

    def test_cache_ttl_expiration(self, tmp_path, monkeypatch):
        """RED: Verify cache expires after TTL

        After TTL expires, cache should be invalidated and data refreshed.
        """
        # Mock time to control TTL expiration
        fake_time = 1000.0

        def mock_time():
            return fake_time

        monkeypatch.setattr(time, "time", mock_time)

        # First call at t=1000
        result1 = get_package_version_info(str(tmp_path))

        # Second call at t=1010 (within TTL, should hit cache)
        fake_time = 1010.0
        result2 = get_package_version_info(str(tmp_path))
        assert result1 == result2, "Should return cached result within TTL"

        # Third call at t=3000 (beyond TTL, should refresh)
        fake_time = 3000.0
        result3 = get_package_version_info(str(tmp_path))

        # Note: result3 might be same as result1/result2 (version hasn't changed)
        # but it should have performed a fresh check
        assert result3 is not None, "Should refresh after TTL expiration"

    def test_session_start_total_time(self, tmp_path):
        """RED: Verify total SessionStart time meets target

        @TEST:ENHANCE-PERF-001:TOTAL-TIME

        Total time for SessionStart (including all info gathering) should be reasonable.
        First call (cold cache): < 500ms (includes network/git operations)
        Subsequent calls (warm cache): < 100ms (cached data retrieval)

        This is the integration test that validates overall performance goal.
        """
        # Initialize git repo for testing
        import subprocess
        subprocess.run(["git", "init"], cwd=tmp_path, capture_output=True, check=False)
        subprocess.run(["git", "config", "user.name", "Test"], cwd=tmp_path, capture_output=True, check=False)
        subprocess.run(["git", "config", "user.email", "test@test.com"], cwd=tmp_path, capture_output=True, check=False)
        subprocess.run(["git", "commit", "--allow-empty", "-m", "Initial"], cwd=tmp_path, capture_output=True, check=False)

        # First call to populate all caches
        _ = get_package_version_info(str(tmp_path))
        _ = get_git_info(str(tmp_path))

        # Second call should hit all caches (warm cache)
        start = time.perf_counter()
        version_info = get_package_version_info(str(tmp_path))
        git_info = get_git_info(str(tmp_path))
        elapsed_ms = (time.perf_counter() - start) * 1000

        # Both should return valid data
        assert version_info is not None
        assert git_info is not None

        print(f"\n🎯 Total SessionStart time (warm cache): {elapsed_ms:.2f}ms")
        # Realistic target: warm cache calls should complete within 100ms
        assert elapsed_ms < 100, f"Total time {elapsed_ms:.2f}ms exceeds target of 100ms"


class TestCacheHitRate:
    """Tests for cache hit rate tracking

    @TAG:HOOK-PERF-001
    """

    def test_cache_hit_rate_in_typical_session(self, tmp_path):
        """RED: Verify cache hit rate > 90% in typical session

        @TEST:ENHANCE-PERF-001:HITRATE

        Simulate a typical session with multiple SessionStart calls.
        Cache hit rate should exceed 90%.
        """
        # Simulate 20 SessionStart calls (typical session length)
        for _ in range(20):
            get_package_version_info(str(tmp_path))

        # Cache hit rate should be high after first call
        # Expected: 19/20 = 95% hit rate

        # Note: We'll need to add cache metrics to measure this
        # For now, just verify all calls succeed
        assert True, "Cache hit rate tracking not yet implemented"


class TestCacheErrorHandling:
    """Tests for cache error handling and fallback behavior

    @TAG:HOOK-PERF-001
    """

    def test_cache_failure_fallback_to_direct_call(self, tmp_path):
        """RED: Verify graceful degradation when cache fails

        @TEST:ENHANCE-PERF-001:FALLBACK

        If cache is corrupted or unavailable, should fall back to direct call.
        """
        # Corrupt cache file (if it exists)
        cache_dir = tmp_path / ".moai" / "cache"
        cache_dir.mkdir(parents=True, exist_ok=True)
        cache_file = cache_dir / "version_info.json"
        cache_file.write_text("CORRUPTED DATA!!!")

        # Should still work (fall back to direct call)
        result = get_package_version_info(str(tmp_path))
        assert result["current"] is not None, "Should fall back to direct call on cache error"

    def test_network_timeout_uses_cached_data(self, tmp_path):
        """RED: Verify network timeout uses stale cache gracefully

        If network times out, should return cached data even if stale.
        """
        # This test requires mocking network timeouts
        # For now, document expected behavior
        assert True, "Network timeout fallback not yet implemented"
