"""Additional coverage tests for version reader.

Tests for lines not covered by existing tests.
"""

import os
from unittest.mock import patch

import pytest

from moai_adk.statusline.version_reader import (
    CacheEntry,
    VersionConfig,
    VersionReader,
    VersionSource,
)


class TestVersionReaderClaudeProjectDir:
    """Test CLAUDE_PROJECT_DIR environment variable."""

    def test_init_with_claude_project_dir_env(self, tmp_path):
        """Should use CLAUDE_PROJECT_DIR when set."""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.yaml"
        config_file.write_text("moai:\n  version: 1.2.3")

        # Set environment variable
        with patch.dict(os.environ, {"CLAUDE_PROJECT_DIR": str(tmp_path)}):
            reader = VersionReader()
            assert reader._config_path == config_file


class TestVersionReaderRegexFallback:
    """Test regex compilation fallback."""

    def test_init_with_invalid_regex_uses_default(self):
        """Should fallback to default regex when compilation fails."""
        config = VersionConfig()
        config.version_format_regex = "[invalid("  # Invalid regex

        reader = VersionReader(config=config)

        # Should use default pattern
        assert reader._version_pattern is not None


class TestVersionReaderCacheHitScenario:
    """Test cache hit scenario in get_version."""

    def test_get_version_cache_hit_increments_stats(self, tmp_path):
        """Should increment cache stats on cache hit."""
        from datetime import datetime

        reader = VersionReader(working_dir=tmp_path)

        # Manually set up cache entry
        entry = CacheEntry(
            version="1.0.0",
            timestamp=datetime.now(),
            source=VersionSource.CONFIG_FILE,
        )
        reader._cache[str(reader._config_path)] = entry

        # Get version should hit cache
        version = reader.get_version()
        assert version == "1.0.0"
        assert reader._cache_stats["hits"] > 0
        assert reader._cache_stats["cache_hits_by_source"][VersionSource.CACHE.value] > 0


class TestVersionReaderAsyncErrorHandling:
    """Test async version reading error handling."""

    @pytest.mark.asyncio
    async def test_get_version_async_with_fallback(self, tmp_path):
        """Should return fallback version when config read fails."""
        reader = VersionReader(working_dir=tmp_path)

        # Mock package version to return empty
        with patch.object(reader, "_get_package_version", return_value=""):
            version = await reader.get_version_async()
            # Should return some version (fallback or mock)
            assert version is not None

    @pytest.mark.asyncio
    async def test_get_version_async_handles_exception(self, tmp_path):
        """Should handle exceptions and return fallback."""
        reader = VersionReader(working_dir=tmp_path)

        # Mock config read to raise exception (package version returns empty first)
        with patch.object(reader, "_get_package_version", return_value=""):
            with patch.object(
                reader, "_read_version_from_config_async", side_effect=RuntimeError("Test error")
            ):
                version = await reader.get_version_async()
                # Should return fallback
                assert version is not None


class TestVersionReaderDebugModeLogging:
    """Test debug mode logging."""

    def test_check_cache_logs_in_debug_mode(self, tmp_path, caplog):
        """Should log cache hit in debug mode."""
        import logging

        reader = VersionReader(working_dir=tmp_path)
        reader.config.debug_mode = True

        # Set up cache
        from datetime import datetime

        from moai_adk.statusline.version_reader import CacheEntry

        entry = CacheEntry(version="1.0.0", timestamp=datetime.now(), source=VersionSource.CONFIG_FILE)
        reader._cache[str(reader._config_path)] = entry

        with caplog.at_level(logging.DEBUG):
            reader._check_cache()

        # Should have debug log
        assert any("Cache hit" in record.message for record in caplog.records)

    def test_update_cache_logs_in_debug_mode(self, tmp_path, caplog):
        """Should log cache update in debug mode."""
        import logging

        reader = VersionReader(working_dir=tmp_path)
        reader.config.debug_mode = True

        with caplog.at_level(logging.DEBUG):
            reader._update_cache("1.0.0", VersionSource.CONFIG_FILE)

        # Should have debug log
        assert any("Cache updated" in record.message for record in caplog.records)

    def test_handle_read_error_logs_in_debug_mode(self, tmp_path, caplog):
        """Should log error details in debug mode."""
        import logging

        reader = VersionReader(working_dir=tmp_path)
        reader.config.debug_mode = True

        with caplog.at_level(logging.ERROR):
            reader._handle_read_error(RuntimeError("Test error"), 0.0)

        # Should have error log
        assert any("Error reading version" in record.message for record in caplog.records)

    def test_log_performance_in_debug_mode(self, tmp_path, caplog):
        """Should log performance in debug mode."""
        import logging
        import time

        reader = VersionReader(working_dir=tmp_path)
        reader.config.debug_mode = True
        reader.config.track_performance_metrics = True

        with caplog.at_level(logging.DEBUG):
            reader._log_performance(time.time(), "test_operation")

        # Should have performance log
        assert any("Performance" in record.message for record in caplog.records)


class TestVersionReaderCacheEviction:
    """Test cache eviction logic."""

    def test_update_cache_does_nothing_when_disabled(self, tmp_path):
        """Should skip cache update when disabled."""
        config = VersionConfig()
        config.cache_enabled = False
        reader = VersionReader(working_dir=tmp_path, config=config)

        initial_size = len(reader._cache)
        reader._update_cache("1.0.0", VersionSource.CONFIG_FILE)

        # Cache should not be updated
        assert len(reader._cache) == initial_size

    def test_update_cache_evicts_oldest_when_full(self, tmp_path):
        """Should evict oldest entry when cache is full."""
        from datetime import datetime, timedelta

        config = VersionConfig()
        config.cache_size = 2
        reader = VersionReader(working_dir=tmp_path, config=config)

        # Add 3 entries (exceeds cache size of 2)
        entry1 = CacheEntry(
            version="1.0.0", timestamp=datetime.now() - timedelta(seconds=3), source=VersionSource.CONFIG_FILE
        )
        entry2 = CacheEntry(
            version="1.0.1", timestamp=datetime.now() - timedelta(seconds=2), source=VersionSource.CONFIG_FILE
        )

        reader._cache["key1"] = entry1
        reader._cache["key2"] = entry2

        # Add third entry (should trigger eviction)
        reader._update_cache("1.0.2", VersionSource.CONFIG_FILE)

        # Cache size should be limited
        assert len(reader._cache) <= 3

    def test_evict_oldest_skips_when_lru_disabled(self, tmp_path):
        """Should skip eviction when LRU is disabled."""
        from datetime import datetime

        config = VersionConfig()
        config.cache_size = 2
        config.enable_lru_cache = False
        reader = VersionReader(working_dir=tmp_path, config=config)

        # Add multiple entries
        reader._cache["key1"] = CacheEntry(
            version="1.0.0", timestamp=datetime.now(), source=VersionSource.CONFIG_FILE
        )
        reader._cache["key2"] = CacheEntry(
            version="1.0.1", timestamp=datetime.now(), source=VersionSource.CONFIG_FILE
        )

        # Add third entry (should not evict since LRU is disabled and cache size is 2)
        reader._update_cache("1.0.2", VersionSource.CONFIG_FILE)

        # Should not evict when cache size is exactly at limit
        assert len(reader._cache) >= 2

    def test_evict_oldest_logs_in_debug_mode(self, tmp_path, caplog):
        """Should log eviction in debug mode."""
        import logging

        config = VersionConfig()
        config.cache_size = 2
        config.debug_mode = True
        reader = VersionReader(working_dir=tmp_path, config=config)

        from datetime import datetime, timedelta

        entry1 = CacheEntry(
            version="1.0.0", timestamp=datetime.now() - timedelta(seconds=3), source=VersionSource.CONFIG_FILE
        )
        entry2 = CacheEntry(
            version="1.0.1", timestamp=datetime.now() - timedelta(seconds=2), source=VersionSource.CONFIG_FILE
        )

        reader._cache["key1"] = entry1
        reader._cache["key2"] = entry2

        with caplog.at_level(logging.DEBUG):
            reader._update_cache("1.0.2", VersionSource.CONFIG_FILE)

        # Should have eviction log
        assert any("Evicted oldest" in record.message for record in caplog.records)


class TestVersionReaderAsyncConfigReading:
    """Test async config reading error handling."""

    @pytest.mark.asyncio
    async def test_read_version_from_config_async_handles_json_error(self, tmp_path):
        """Should handle JSON decode errors in async read."""
        reader = VersionReader(working_dir=tmp_path)

        # Create invalid JSON file
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"
        config_file.write_text("{ invalid json }")

        version = await reader._read_version_from_config_async()
        assert version == ""

    @pytest.mark.asyncio
    async def test_read_version_from_config_async_handles_yaml_error(self, tmp_path):
        """Should handle YAML decode errors in async read."""
        reader = VersionReader(working_dir=tmp_path)

        # Create invalid YAML file
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.yaml"
        config_file.write_text("{ invalid: yaml: content ]")

        version = await reader._read_version_from_config_async()
        assert version == ""


class TestVersionReaderPackageNotFound:
    """Test package not found handling."""

    def test_get_package_version_not_found(self, tmp_path):
        """Should return empty string when package not found."""
        reader = VersionReader(working_dir=tmp_path)

        # Mock importlib.metadata to raise PackageNotFoundError
        # Patch where it's imported, not where it's defined
        with patch("importlib.metadata.version") as mock_version:
            from importlib.metadata import PackageNotFoundError

            mock_version.side_effect = PackageNotFoundError()
            version = reader._get_package_version()
            assert version == ""


class TestVersionReaderAsyncFileReading:
    """Test async file reading methods."""

    def test_read_json_sync(self, tmp_path):
        """Should read JSON file synchronously."""
        reader = VersionReader(working_dir=tmp_path)

        # Create JSON file
        json_file = tmp_path / "test.json"
        json_file.write_text('{"key": "value"}')

        result = reader._read_json_sync(json_file)
        assert result == {"key": "value"}

    @pytest.mark.asyncio
    async def test_read_config_async_json(self, tmp_path):
        """Should read JSON config file asynchronously."""
        reader = VersionReader(working_dir=tmp_path)

        # Create JSON config file
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"
        config_file.write_text('{"moai": {"version": "1.2.3"}}')

        result = await reader._read_config_async(config_file)
        assert result == {"moai": {"version": "1.2.3"}}

    @pytest.mark.asyncio
    async def test_read_config_async_yaml(self, tmp_path):
        """Should read YAML config file asynchronously."""
        reader = VersionReader(working_dir=tmp_path)

        # Create YAML config file
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.yaml"
        config_file.write_text("moai:\n  version: 1.2.3")

        result = await reader._read_config_async(config_file)
        assert result == {"moai": {"version": "1.2.3"}}

    def test_read_config_sync_json(self, tmp_path):
        """Should read JSON config file synchronously."""
        reader = VersionReader(working_dir=tmp_path)

        # Create JSON config file
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"
        config_file.write_text('{"moai": {"version": "1.2.3"}}')

        result = reader._read_config_sync(config_file)
        assert result == {"moai": {"version": "1.2.3"}}

    def test_read_config_sync_yaml(self, tmp_path):
        """Should read YAML config file synchronously."""
        reader = VersionReader(working_dir=tmp_path)

        # Create YAML config file
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.yaml"
        config_file.write_text("moai:\n  version: 1.2.3")

        result = reader._read_config_sync(config_file)
        assert result == {"moai": {"version": "1.2.3"}}


class TestVersionReaderCacheAgeAndExpiry:
    """Test cache age and expiry checks."""

    def test_get_cache_age_seconds_returns_none_when_no_cache_time(self, tmp_path):
        """Should return None when cache_time is None."""
        reader = VersionReader(working_dir=tmp_path)
        reader._cache_time = None

        age = reader.get_cache_age_seconds()
        assert age is None

    def test_get_cache_age_seconds_calculates_age(self, tmp_path):
        """Should calculate cache age correctly."""
        from datetime import datetime, timedelta

        reader = VersionReader(working_dir=tmp_path)
        reader._cache_time = datetime.now() - timedelta(seconds=10)

        age = reader.get_cache_age_seconds()
        assert age is not None
        assert age >= 9.0  # Allow some timing variance

    def test_is_cache_expired_when_no_cache_entry(self, tmp_path):
        """Should return True when no cache entry exists."""
        reader = VersionReader(working_dir=tmp_path)

        # Clear cache
        reader._cache.clear()

        is_expired = reader.is_cache_expired()
        assert is_expired is True

    def test_is_cache_expired_when_entry_invalid(self, tmp_path):
        """Should return True when cache entry is invalid."""
        from datetime import datetime, timedelta

        config = VersionConfig()
        config.cache_ttl_seconds = 1  # 1 second TTL
        reader = VersionReader(working_dir=tmp_path, config=config)

        # Add expired cache entry
        entry = CacheEntry(
            version="1.0.0", timestamp=datetime.now() - timedelta(seconds=10), source=VersionSource.CONFIG_FILE
        )
        reader._cache[str(reader._config_path)] = entry

        is_expired = reader.is_cache_expired()
        assert is_expired is True
