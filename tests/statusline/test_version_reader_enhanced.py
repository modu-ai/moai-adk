"""Enhanced tests for version_reader.py - Batch 3 coverage improvements

Focus: Performance metrics, cache management, async operations, error recovery
Target Coverage: 75.0% → 90.0% (+15%)
"""

import json
import time
from datetime import datetime, timedelta
from unittest.mock import patch

import pytest

from moai_adk.statusline.version_reader import (
    CacheEntry,
    VersionConfig,
    VersionReader,
    VersionSource,
)


class TestPerformanceMetrics:
    """Test performance metrics tracking - NEW COVERAGE"""

    def test_performance_metrics_collection(self, tmp_path):
        """Test performance metrics are collected during operations"""
        config = VersionConfig(track_performance_metrics=True)
        reader = VersionReader(config=config, working_dir=tmp_path)

        # Create config
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        (config_dir / "config.json").write_text(json.dumps({"moai": {"version": "0.20.1"}}))

        # Perform operations
        reader.get_version()
        reader.get_version()  # Second call uses cache

        metrics = reader.get_performance_metrics()

        # Check metrics structure
        assert "cache_stats" in metrics
        assert "performance_metrics" in metrics
        assert "cache_size" in metrics

    def test_performance_metrics_statistics(self, tmp_path):
        """Test performance metrics calculate statistics correctly"""
        config = VersionConfig(track_performance_metrics=True)
        reader = VersionReader(config=config, working_dir=tmp_path)

        # Manually add performance data
        reader._performance_metrics["test_operation"] = [0.1, 0.2, 0.15]

        metrics = reader.get_performance_metrics()

        assert "test_operation" in metrics["performance_metrics"]
        stats = metrics["performance_metrics"]["test_operation"]
        assert stats["count"] == 3
        assert 0.1 <= stats["average"] <= 0.2
        assert stats["min"] == 0.1
        assert stats["max"] == 0.2

    def test_performance_metrics_disabled(self, tmp_path):
        """Test performance tracking can be disabled"""
        config = VersionConfig(track_performance_metrics=False)
        reader = VersionReader(config=config, working_dir=tmp_path)

        # Operations should not track performance
        reader._log_performance(time.time(), "test")

        metrics = reader.get_performance_metrics()
        assert metrics["performance_metrics"] == {}


class TestAsyncOperations:
    """Test async version reading - NEW COVERAGE"""

    @pytest.mark.asyncio
    async def test_async_version_reading(self, tmp_path):
        """Test async version reading"""
        config = VersionConfig(enable_async=True)
        reader = VersionReader(config=config, working_dir=tmp_path)

        # Create config
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        (config_dir / "config.json").write_text(json.dumps({"moai": {"version": "0.20.1"}}))

        version = await reader.get_version_async()
        assert version == "0.20.1"

    @pytest.mark.asyncio
    async def test_async_file_exists_check(self, tmp_path):
        """Test async file existence check"""
        reader = VersionReader(working_dir=tmp_path)

        # Test existing file
        test_file = tmp_path / "test.txt"
        test_file.write_text("test")
        assert await reader._file_exists_async(test_file) is True

        # Test non-existing file
        assert await reader._file_exists_async(tmp_path / "nonexistent.txt") is False

    @pytest.mark.asyncio
    async def test_async_json_reading(self, tmp_path):
        """Test async JSON file reading"""
        reader = VersionReader(working_dir=tmp_path)

        test_file = tmp_path / "test.json"
        test_data = {"key": "value"}
        test_file.write_text(json.dumps(test_data))

        result = await reader._read_json_async(test_file)
        assert result == test_data


class TestCacheManagement:
    """Test cache management and LRU eviction - NEW COVERAGE"""

    def test_cache_lru_eviction_disabled(self, tmp_path):
        """Test LRU eviction can be disabled"""
        config = VersionConfig(cache_enabled=True, cache_size=2, enable_lru_cache=False)
        reader = VersionReader(config=config, working_dir=tmp_path)

        # Fill cache
        reader._cache["key1"] = CacheEntry(version="0.20.1", timestamp=datetime.now(), source=VersionSource.CONFIG_FILE)
        reader._cache["key2"] = CacheEntry(version="0.20.2", timestamp=datetime.now(), source=VersionSource.CONFIG_FILE)

        # Try to evict (should do nothing)
        reader._evict_oldest_cache_entry()

        # Cache should still have 2 entries
        assert len(reader._cache) == 2

    def test_cache_access_count_tracking(self, tmp_path):
        """Test cache access count is tracked"""
        reader = VersionReader(working_dir=tmp_path)

        cache_key = str(reader._config_path)
        reader._cache[cache_key] = CacheEntry(
            version="0.20.1", timestamp=datetime.now(), source=VersionSource.CONFIG_FILE, access_count=0
        )

        # Access cache multiple times
        reader._check_cache()
        reader._check_cache()

        assert reader._cache[cache_key].access_count == 2

    def test_cache_stats_by_source(self, tmp_path):
        """Test cache statistics track hits by source"""
        reader = VersionReader(working_dir=tmp_path)

        # Populate cache
        cache_key = str(reader._config_path)
        reader._cache[cache_key] = CacheEntry(
            version="0.20.1", timestamp=datetime.now(), source=VersionSource.CONFIG_FILE
        )

        # Hit cache
        reader._check_cache()

        stats = reader.get_cache_stats()
        assert stats["cache_hits_by_source"][VersionSource.CACHE.value] > 0


class TestVersionFieldPriority:
    """Test version field priority and extraction - NEW COVERAGE"""

    def test_nested_value_extraction_nested_path(self, tmp_path):
        """Test extracting deeply nested values"""
        reader = VersionReader(working_dir=tmp_path)
        config = {"level1": {"level2": {"level3": "deep_value"}}}

        value = reader._get_nested_value(config, "level1.level2.level3")
        assert value == "deep_value"

    def test_nested_value_extraction_missing_key(self, tmp_path):
        """Test extraction returns None for missing keys"""
        reader = VersionReader(working_dir=tmp_path)
        config = {"moai": {"version": "0.20.1"}}

        value = reader._get_nested_value(config, "missing.key.path")
        assert value is None

    def test_get_available_version_fields(self, tmp_path):
        """Test getting list of available version fields"""
        reader = VersionReader(working_dir=tmp_path)
        fields = reader.get_available_version_fields()

        assert isinstance(fields, list)
        assert "moai.version" in fields
        assert "project.version" in fields


class TestVersionFormatting:
    """Test version formatting methods - NEW COVERAGE"""

    def test_format_short_version_with_prefix(self, tmp_path):
        """Test formatting version by removing 'v' prefix"""
        reader = VersionReader(working_dir=tmp_path)

        assert reader._format_short_version("v0.20.1") == "0.20.1"
        assert reader._format_short_version("0.20.1") == "0.20.1"

    def test_format_display_version_with_unknown(self, tmp_path):
        """Test formatting display version for unknown"""
        reader = VersionReader(working_dir=tmp_path)

        assert reader._format_display_version("unknown") == "MoAI-ADK unknown version"

    def test_format_display_version_with_prefix(self, tmp_path):
        """Test formatting display version with 'v' prefix"""
        reader = VersionReader(working_dir=tmp_path)

        result = reader._format_display_version("v0.20.1")
        assert "MoAI-ADK" in result
        assert "v0.20.1" in result

    def test_format_display_version_without_prefix(self, tmp_path):
        """Test formatting display version without prefix"""
        reader = VersionReader(working_dir=tmp_path)

        result = reader._format_display_version("0.20.1")
        assert "MoAI-ADK" in result
        assert "0.20.1" in result


class TestVersionValidation:
    """Test version format validation - NEW COVERAGE"""

    def test_valid_version_format(self, tmp_path):
        """Test validating correct version formats"""
        reader = VersionReader(working_dir=tmp_path)

        assert reader._is_valid_version_format("0.20.1") is True
        assert reader._is_valid_version_format("v0.20.1") is True
        assert reader._is_valid_version_format("1.0.0") is True
        assert reader._is_valid_version_format("v1.0.0-alpha") is True

    def test_invalid_version_format(self, tmp_path):
        """Test validating incorrect version formats"""
        reader = VersionReader(working_dir=tmp_path)

        assert reader._is_valid_version_format("invalid") is False
        assert reader._is_valid_version_format("1.0") is False
        assert reader._is_valid_version_format("") is False


class TestPackageVersionFallback:
    """Test package version fallback mechanism - NEW COVERAGE"""

    def test_get_package_version_success(self, tmp_path):
        """Test getting version from package metadata"""
        reader = VersionReader(working_dir=tmp_path)

        with patch("importlib.metadata.version", return_value="0.20.1"):
            version = reader._get_package_version()
            assert version == "0.20.1"

    def test_get_package_version_not_found(self, tmp_path):
        """Test handling package not found error"""
        reader = VersionReader(working_dir=tmp_path)

        with patch("importlib.metadata.version", side_effect=Exception("Package not found")):
            version = reader._get_package_version()
            assert version == ""

    def test_fallback_version_with_package(self, tmp_path):
        """Test fallback tries package version first"""
        reader = VersionReader(working_dir=tmp_path)

        with patch.object(reader, "_get_package_version", return_value="0.20.1"):
            fallback = reader._get_fallback_version()
            assert fallback == "0.20.1"

    def test_fallback_version_without_package(self, tmp_path):
        """Test fallback uses configured fallback when package unavailable"""
        config = VersionConfig(fallback_version="dev")
        reader = VersionReader(config=config, working_dir=tmp_path)

        with patch.object(reader, "_get_package_version", return_value=""):
            fallback = reader._get_fallback_version()
            assert fallback == "dev"


class TestConfigurationManagement:
    """Test configuration management - NEW COVERAGE"""

    def test_get_config(self, tmp_path):
        """Test getting current configuration"""
        config = VersionConfig(cache_ttl_seconds=120)
        reader = VersionReader(config=config, working_dir=tmp_path)

        retrieved_config = reader.get_config()
        assert retrieved_config.cache_ttl_seconds == 120

    def test_update_config(self, tmp_path):
        """Test updating configuration"""
        reader = VersionReader(working_dir=tmp_path)
        new_config = VersionConfig(cache_ttl_seconds=180)

        reader.update_config(new_config)

        assert reader.config.cache_ttl_seconds == 180
        assert reader._cache_ttl.total_seconds() == 180


class TestErrorRecovery:
    """Test error recovery and handling - NEW COVERAGE"""

    def test_handle_read_error_tracking(self, tmp_path):
        """Test error tracking in cache stats"""
        reader = VersionReader(working_dir=tmp_path)
        initial_errors = reader._cache_stats["errors"]

        error = Exception("Test error")
        reader._handle_read_error(error, time.time())

        assert reader._cache_stats["errors"] == initial_errors + 1

    def test_handle_read_error_logging_debug_mode(self, tmp_path):
        """Test error logging in debug mode"""
        config = VersionConfig(debug_mode=True)
        reader = VersionReader(config=config, working_dir=tmp_path)

        with patch.object(reader._logger, "error") as mock_error:
            error = Exception("Test error")
            reader._handle_read_error(error, time.time())
            mock_error.assert_called()

    def test_get_cache_age_when_no_cache(self, tmp_path):
        """Test getting cache age when no cache exists"""
        reader = VersionReader(working_dir=tmp_path)
        assert reader.get_cache_age_seconds() is None

    def test_get_cache_age_with_valid_cache(self, tmp_path):
        """Test getting cache age with valid cache"""
        reader = VersionReader(working_dir=tmp_path)
        reader._cache_time = datetime.now() - timedelta(seconds=30)

        age = reader.get_cache_age_seconds()
        assert age is not None
        assert 29 <= age <= 31  # Allow small time variance


class TestVersionExtraction:
    """Test version extraction from config - NEW COVERAGE"""

    def test_extract_version_with_multiple_fallbacks(self, tmp_path):
        """Test version extraction tries multiple fields"""
        reader = VersionReader(working_dir=tmp_path)

        # Config with only template_version
        config = {"template_version": "0.28.0"}
        version = reader._extract_version_from_config(config)
        assert version == "0.28.0"

    def test_extract_version_empty_config(self, tmp_path):
        """Test version extraction from empty config"""
        reader = VersionReader(working_dir=tmp_path)
        version = reader._extract_version_from_config({})
        assert version == ""

    def test_read_version_from_config_sync_missing_file(self, tmp_path):
        """Test sync reading when config file missing"""
        reader = VersionReader(working_dir=tmp_path)
        version = reader._read_version_from_config_sync()
        assert version == ""


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--cov=moai_adk.statusline.version_reader", "--cov-report=term-missing"])
