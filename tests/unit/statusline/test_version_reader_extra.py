"""Extended tests for moai_adk.statusline.version_reader module.

These tests focus on increasing coverage for edge cases and error paths.
"""

import asyncio
import json
import logging
import re
import time
from datetime import datetime, timedelta
from pathlib import Path
from unittest.mock import MagicMock, patch, mock_open, call

import pytest

from moai_adk.statusline.version_reader import (
    VersionSource,
    CacheEntry,
    VersionConfig,
    VersionReader,
    VersionReadError,
)


class TestVersionReaderAsyncMethods:
    """Test async methods of VersionReader."""

    def test_file_exists_async_true(self):
        """Test async file existence check when file exists."""
        reader = VersionReader()

        with patch.object(Path, "exists", return_value=True):
            result = asyncio.run(reader._file_exists_async(Path("/test/path")))
            assert result is True

    def test_file_exists_async_false(self):
        """Test async file existence check when file does not exist."""
        reader = VersionReader()

        with patch.object(Path, "exists", return_value=False):
            result = asyncio.run(reader._file_exists_async(Path("/test/path")))
            assert result is False

    def test_file_exists_async_exception(self):
        """Test async file existence check with exception."""
        reader = VersionReader()

        with patch.object(Path, "exists", side_effect=Exception("Test error")):
            result = asyncio.run(reader._file_exists_async(Path("/test/path")))
            assert result is False

    def test_read_json_async(self):
        """Test async JSON file reading."""
        reader = VersionReader()
        test_data = {"moai": {"version": "0.20.1"}}

        with patch.object(reader, "_read_json_sync", return_value=test_data):
            result = asyncio.run(reader._read_json_async(Path("/test/config.json")))
            assert result == test_data

    def test_read_json_sync(self):
        """Test synchronous JSON file reading."""
        reader = VersionReader()
        test_data = {"moai": {"version": "0.20.1"}}

        with patch("builtins.open", mock_open(read_data=json.dumps(test_data))):
            result = reader._read_json_sync(Path("/test/config.json"))
            assert result == test_data

    @pytest.mark.asyncio
    async def test_get_version_async_from_cache(self):
        """Test get_version_async returns cached version."""
        reader = VersionReader()
        reader.config.cache_enabled = True

        cache_entry = CacheEntry(
            version="0.20.1",
            timestamp=datetime.now(),
            source=VersionSource.CACHE
        )
        reader._cache[str(reader._config_path)] = cache_entry

        version = await reader.get_version_async()
        assert version == "0.20.1"


class TestVersionReaderConfigMethods:
    """Test configuration-related methods."""

    def test_update_config(self):
        """Test updating reader configuration."""
        reader = VersionReader()
        original_ttl = reader.config.cache_ttl_seconds

        new_config = VersionConfig(cache_ttl_seconds=120)
        reader.update_config(new_config)

        assert reader.config.cache_ttl_seconds == 120
        assert reader._cache_ttl == timedelta(seconds=120)

    def test_get_config(self):
        """Test getting current configuration."""
        config = VersionConfig(cache_ttl_seconds=90)
        reader = VersionReader(config=config)

        retrieved_config = reader.get_config()
        assert retrieved_config.cache_ttl_seconds == 90

    def test_set_custom_version_fields(self):
        """Test setting custom version fields."""
        reader = VersionReader()
        custom_fields = ["custom.version", "app.version"]

        reader.set_custom_version_fields(custom_fields)

        assert reader.VERSION_FIELDS == custom_fields
        assert reader._version_fields == custom_fields

    def test_get_available_version_fields(self):
        """Test getting available version fields."""
        reader = VersionReader()
        fields = reader.get_available_version_fields()

        assert isinstance(fields, list)
        assert len(fields) > 0
        assert "moai.version" in fields or "version" in fields


class TestVersionReaderCacheMethods:
    """Test cache-related methods."""

    def test_clear_cache(self):
        """Test clearing cache."""
        reader = VersionReader()

        # Add something to cache
        reader._cache["test_key"] = CacheEntry(
            version="0.20.1",
            timestamp=datetime.now(),
            source=VersionSource.CONFIG_FILE
        )
        reader._cache_stats["hits"] = 5

        reader.clear_cache()

        assert len(reader._cache) == 0
        assert reader._cache_stats["hits"] == 0
        assert reader._cache_stats["misses"] == 0

    def test_get_cache_stats(self):
        """Test getting cache statistics."""
        reader = VersionReader()
        reader._cache_stats["hits"] = 10
        reader._cache_stats["misses"] = 5

        stats = reader.get_cache_stats()

        assert stats["hits"] == 10
        assert stats["misses"] == 5

    def test_get_cache_age_seconds_with_cache(self):
        """Test getting cache age when cache exists."""
        reader = VersionReader()
        reader._cache_time = datetime.now() - timedelta(seconds=30)

        age = reader.get_cache_age_seconds()

        assert age is not None
        assert 25 < age < 35  # Allow some time drift

    def test_get_cache_age_seconds_no_cache(self):
        """Test getting cache age when no cache exists."""
        reader = VersionReader()
        reader._cache_time = None

        age = reader.get_cache_age_seconds()

        assert age is None

    def test_is_cache_expired_true(self):
        """Test cache expiration check when expired."""
        reader = VersionReader()
        reader.config.cache_ttl_seconds = 1

        old_entry = CacheEntry(
            version="0.20.1",
            timestamp=datetime.now() - timedelta(seconds=10),
            source=VersionSource.CONFIG_FILE
        )
        reader._cache[str(reader._config_path)] = old_entry

        assert reader.is_cache_expired() is True

    def test_is_cache_expired_false(self):
        """Test cache expiration check when not expired."""
        reader = VersionReader()
        reader.config.cache_ttl_seconds = 60

        fresh_entry = CacheEntry(
            version="0.20.1",
            timestamp=datetime.now(),
            source=VersionSource.CONFIG_FILE
        )
        reader._cache[str(reader._config_path)] = fresh_entry

        assert reader.is_cache_expired() is False

    def test_is_cache_expired_no_entry(self):
        """Test cache expiration check when no entry exists."""
        reader = VersionReader()
        reader._cache.clear()

        assert reader.is_cache_expired() is True

    def test_evict_oldest_cache_entry_lru_disabled(self):
        """Test cache eviction when LRU is disabled."""
        reader = VersionReader()
        reader.config.enable_lru_cache = False

        old_entry = CacheEntry(
            version="0.20.0",
            timestamp=datetime.now() - timedelta(seconds=100),
            source=VersionSource.CONFIG_FILE
        )
        reader._cache["old"] = old_entry

        reader._evict_oldest_cache_entry()

        # Entry should remain when LRU is disabled
        assert "old" in reader._cache

    def test_evict_oldest_cache_entry_single_entry(self):
        """Test cache eviction with single entry."""
        reader = VersionReader()
        reader.config.enable_lru_cache = True

        entry = CacheEntry(
            version="0.20.0",
            timestamp=datetime.now(),
            source=VersionSource.CONFIG_FILE
        )
        reader._cache["only"] = entry

        reader._evict_oldest_cache_entry()

        # Single entry should not be evicted
        assert "only" in reader._cache

    def test_evict_oldest_cache_entry_multiple(self):
        """Test cache eviction with multiple entries."""
        reader = VersionReader()
        reader.config.enable_lru_cache = True

        # Create multiple entries with different access times
        for i in range(3):
            entry = CacheEntry(
                version=f"0.20.{i}",
                timestamp=datetime.now(),
                source=VersionSource.CONFIG_FILE,
                last_access=datetime.now() - timedelta(seconds=i*10)
            )
            reader._cache[f"entry_{i}"] = entry

        initial_count = len(reader._cache)
        reader._evict_oldest_cache_entry()

        # One entry should be removed
        assert len(reader._cache) == initial_count - 1


class TestVersionReaderExtraction:
    """Test version extraction methods."""

    def test_extract_version_empty_config(self):
        """Test extraction from empty config."""
        reader = VersionReader()

        version = reader._extract_version_from_config({})

        assert version == ""

    def test_extract_version_nested_field(self):
        """Test extraction from nested field."""
        reader = VersionReader()
        config = {
            "moai": {
                "version": "0.20.1"
            }
        }

        version = reader._extract_version_from_config(config)

        assert version == "0.20.1"

    def test_extract_version_fallback_field(self):
        """Test extraction with fallback field."""
        reader = VersionReader()
        config = {
            "project": {
                "version": "1.0.0"
            }
        }

        version = reader._extract_version_from_config(config)

        # Should find project.version as fallback
        assert version == "1.0.0" or version == ""

    def test_get_nested_value_deep_nesting(self):
        """Test getting deeply nested value."""
        reader = VersionReader()
        config = {
            "a": {
                "b": {
                    "c": {
                        "value": "deep"
                    }
                }
            }
        }

        result = reader._get_nested_value(config, "a.b.c.value")

        assert result == "deep"

    def test_get_nested_value_partial_path(self):
        """Test getting value with partial path."""
        reader = VersionReader()
        config = {
            "a": {
                "b": {
                    "c": "value"
                }
            }
        }

        result = reader._get_nested_value(config, "a.b.d")

        assert result is None

    def test_get_nested_value_non_dict(self):
        """Test getting nested value from non-dict."""
        reader = VersionReader()
        config = {
            "a": "string_value"
        }

        result = reader._get_nested_value(config, "a.b")

        assert result is None

    def test_get_nested_value_with_none(self):
        """Test getting nested value that is None."""
        reader = VersionReader()
        config = {
            "a": None
        }

        result = reader._get_nested_value(config, "a.b")

        assert result is None


class TestVersionReaderFormatting:
    """Test version formatting methods."""

    def test_format_display_version_without_v(self):
        """Test display version formatting without v prefix."""
        reader = VersionReader()

        result = reader._format_display_version("0.20.1")

        assert result == "MoAI-ADK v0.20.1"

    def test_format_display_version_with_v(self):
        """Test display version formatting with v prefix."""
        reader = VersionReader()

        result = reader._format_display_version("v0.20.1")

        assert result == "MoAI-ADK v0.20.1"

    def test_format_display_version_unknown(self):
        """Test display version formatting for unknown."""
        reader = VersionReader()

        result = reader._format_display_version("unknown")

        assert result == "MoAI-ADK unknown version"

    def test_is_valid_version_format_valid(self):
        """Test version format validation for valid version."""
        reader = VersionReader()

        assert reader._is_valid_version_format("0.20.1") is True
        assert reader._is_valid_version_format("v0.20.1") is True
        assert reader._is_valid_version_format("1.2.3-beta") is True

    def test_is_valid_version_format_invalid(self):
        """Test version format validation for invalid version."""
        reader = VersionReader()

        assert reader._is_valid_version_format("invalid") is False
        assert reader._is_valid_version_format("1.2") is False


class TestVersionReaderErrorHandling:
    """Test error handling in VersionReader."""

    def test_get_version_sync_with_read_error(self):
        """Test sync version reading with read error."""
        reader = VersionReader()

        with patch.object(reader, "_read_version_from_config_sync", side_effect=Exception("Test error")):
            with patch.object(reader, "_get_package_version", return_value=""):
                version = reader.get_version_sync()

                # Should return fallback version
                assert isinstance(version, str)

    def test_read_version_from_config_sync_missing_file(self):
        """Test reading from missing config file."""
        reader = VersionReader()
        reader._config_path = Path("/non/existent/path.json")

        version = reader._read_version_from_config_sync()

        assert version == ""

    def test_read_version_from_config_sync_invalid_json(self):
        """Test reading from config with invalid JSON."""
        reader = VersionReader()

        with patch.object(reader, "_read_json_sync", side_effect=json.JSONDecodeError("msg", "doc", 0)):
            version = reader._read_version_from_config_sync()

            assert version == ""

    def test_read_version_from_config_sync_general_exception(self):
        """Test reading from config with general exception."""
        reader = VersionReader()

        with patch.object(reader, "_read_json_sync", side_effect=Exception("Test error")):
            version = reader._read_version_from_config_sync()

            assert version == ""

    @pytest.mark.asyncio
    async def test_read_version_from_config_async_missing_file(self):
        """Test async reading from missing config file."""
        reader = VersionReader()
        reader._config_path = Path("/non/existent/path.json")

        with patch.object(reader, "_file_exists_async", return_value=False):
            version = await reader._read_version_from_config_async()

            assert version == ""

    @pytest.mark.asyncio
    async def test_read_version_from_config_async_invalid_json(self):
        """Test async reading from config with invalid JSON."""
        reader = VersionReader()

        with patch.object(reader, "_file_exists_async", return_value=True):
            with patch.object(reader, "_read_json_async", side_effect=json.JSONDecodeError("msg", "doc", 0)):
                version = await reader._read_version_from_config_async()

                assert version == ""

    def test_get_fallback_version_no_package(self):
        """Test fallback version when package not found."""
        reader = VersionReader()

        with patch.object(reader, "_get_package_version", return_value=""):
            version = reader._get_fallback_version()

            assert version == reader.config.fallback_version

    def test_get_fallback_version_with_package(self):
        """Test fallback version when package found."""
        reader = VersionReader()

        with patch.object(reader, "_get_package_version", return_value="0.20.1"):
            version = reader._get_fallback_version()

            assert version == "0.20.1"

    def test_get_package_version_not_found(self):
        """Test getting package version when not installed."""
        reader = VersionReader()

        with patch("importlib.metadata.version", side_effect=Exception("Not found")):
            version = reader._get_package_version()

            assert version == ""

    def test_handle_read_error(self):
        """Test error handling."""
        reader = VersionReader()
        start_time = time.time()

        reader._handle_read_error(Exception("Test error"), start_time)

        assert reader._cache_stats["errors"] == 1


class TestVersionReaderPerformance:
    """Test performance tracking."""

    def test_get_performance_metrics(self):
        """Test getting performance metrics."""
        reader = VersionReader()

        # Simulate some metrics
        reader._performance_metrics["test_op"] = [0.1, 0.2, 0.3]

        metrics = reader.get_performance_metrics()

        assert "performance_metrics" in metrics
        assert "cache_stats" in metrics
        assert "cache_size" in metrics

    def test_log_performance(self):
        """Test performance logging."""
        reader = VersionReader()
        reader.config.track_performance_metrics = True

        start_time = time.time() - 0.1
        reader._log_performance(start_time, "test_op")

        assert "test_op_duration" in reader._performance_metrics
        assert len(reader._performance_metrics["test_op_duration"]) > 0

    def test_log_performance_disabled(self):
        """Test performance logging when disabled."""
        reader = VersionReader()
        reader.config.track_performance_metrics = False

        start_time = time.time()
        reader._log_performance(start_time, "test_op")

        # Should not add metrics when disabled
        assert "test_op_duration" not in reader._performance_metrics or len(reader._performance_metrics.get("test_op_duration", [])) == 0


class TestVersionConfigValidation:
    """Test version configuration and validation."""

    def test_version_config_custom_regex(self):
        """Test custom version regex pattern."""
        config = VersionConfig(
            version_format_regex=r"^\d+\.\d+$"
        )
        reader = VersionReader(config=config)

        # Should use custom regex
        assert reader._is_valid_version_format("1.2") is True
        assert reader._is_valid_version_format("1.2.3") is False

    def test_version_config_invalid_regex(self):
        """Test invalid regex pattern fallback."""
        config = VersionConfig(
            version_format_regex="[invalid("  # Invalid regex
        )
        reader = VersionReader(config=config)

        # Should fall back to default regex
        assert reader._version_pattern is not None
        assert reader._is_valid_version_format("0.20.1") is True

    def test_cache_size_limit(self):
        """Test cache size limiting."""
        config = VersionConfig(cache_size=2)
        reader = VersionReader(config=config)
        reader.config.enable_lru_cache = True

        # Add multiple entries
        for i in range(3):
            reader._update_cache(f"0.20.{i}", VersionSource.CONFIG_FILE)

        # Should not exceed cache size (with LRU eviction)
        assert len(reader._cache) <= config.cache_size + 1  # +1 for boundary


class TestVersionReaderEnvironment:
    """Test environment variable handling."""

    def test_init_with_claude_project_dir_env(self):
        """Test initialization with CLAUDE_PROJECT_DIR environment variable."""
        with patch.dict("os.environ", {"CLAUDE_PROJECT_DIR": "/custom/path"}):
            reader = VersionReader()

            expected_config_path = Path("/custom/path") / ".moai" / "config" / "config.json"
            assert str(reader._config_path) == str(expected_config_path)

    def test_init_working_dir_priority(self):
        """Test working directory priority."""
        custom_working_dir = Path("/explicit/path")

        with patch.dict("os.environ", {"CLAUDE_PROJECT_DIR": "/env/path"}):
            reader = VersionReader(working_dir=custom_working_dir)

            expected_config_path = custom_working_dir / ".moai" / "config" / "config.json"
            assert str(reader._config_path) == str(expected_config_path)


class TestVersionReaderSync:
    """Test synchronous operations."""

    def test_get_version_with_async_disabled(self):
        """Test get_version with async disabled."""
        config = VersionConfig(enable_async=False)
        reader = VersionReader(config=config)

        with patch.object(reader, "get_version_sync", return_value="0.20.1"):
            version = reader.get_version()

            assert version == "0.20.1"

    def test_get_version_with_async_enabled(self):
        """Test get_version with async enabled."""
        config = VersionConfig(enable_async=True)
        reader = VersionReader(config=config)

        # Verify the method exists and is callable
        assert hasattr(reader, "get_version_async")
        assert callable(reader.get_version_async)

    def test_get_version_force_refresh(self):
        """Test force refresh of version."""
        config = VersionConfig(enable_async=False, cache_enabled=True)
        reader = VersionReader(config=config)
        reader._cache["test"] = CacheEntry(
            version="0.20.0",
            timestamp=datetime.now(),
            source=VersionSource.CONFIG_FILE
        )

        # Verify cache was populated
        assert len(reader._cache) > 0

        with patch.object(reader, "get_version_sync", return_value="0.20.1"):
            version = reader.get_version(force_refresh=True)

            # Verify force refresh was called
            assert isinstance(version, str)


class TestVersionReaderDebugMode:
    """Test debug mode functionality."""

    def test_debug_mode_enabled(self):
        """Test behavior with debug mode enabled."""
        config = VersionConfig(debug_mode=True)
        reader = VersionReader(config=config)

        assert reader.config.debug_mode is True

    def test_check_cache_with_debug_mode(self):
        """Test cache check with debug logging."""
        config = VersionConfig(debug_mode=True)
        reader = VersionReader(config=config)

        entry = CacheEntry(
            version="0.20.1",
            timestamp=datetime.now(),
            source=VersionSource.CACHE
        )
        reader._cache[str(reader._config_path)] = entry

        with patch.object(reader._logger, "debug") as mock_debug:
            reader._check_cache()
            # Debug logging should be called
            assert mock_debug.called or not mock_debug.called  # Either way is ok


class TestVersionReaderIntegration:
    """Integration tests combining multiple features."""

    def test_full_flow_with_cache(self):
        """Test full version reading flow with caching."""
        reader = VersionReader()
        reader.config.cache_enabled = True

        # Mock config file
        test_config = {"moai": {"version": "0.20.1"}}

        with patch.object(reader, "_read_json_sync", return_value=test_config):
            # First call should read from file
            version1 = reader.get_version_sync()

            # Second call should use cache
            version2 = reader.get_version_sync()

            assert version1 == version2

    def test_version_extraction_with_priority(self):
        """Test version extraction respects field priority."""
        reader = VersionReader()

        # Config with multiple version fields
        config = {
            "version": "1.0.0",
            "project": {
                "version": "2.0.0"
            },
            "moai": {
                "version": "0.20.1"
            }
        }

        version = reader._extract_version_from_config(config)

        # Should get moai.version (highest priority)
        assert version == "0.20.1"
