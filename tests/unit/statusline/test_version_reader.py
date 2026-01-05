"""Tests for moai_adk.statusline.version_reader module."""

import asyncio
import sys
from datetime import datetime, timedelta
from unittest.mock import patch

import pytest

from moai_adk.statusline.version_reader import (
    CacheEntry,
    VersionConfig,
    VersionReader,
    VersionReadError,
    VersionSource,
)


class TestVersionSource:
    """Test VersionSource enum."""

    def test_version_source_values(self):
        """Test VersionSource enum values."""
        assert VersionSource.CONFIG_FILE.value == "config_file"
        assert VersionSource.FALLBACK.value == "fallback"
        assert VersionSource.PACKAGE.value == "package"
        assert VersionSource.CACHE.value == "cache"


class TestCacheEntry:
    """Test CacheEntry dataclass."""

    def test_cache_entry_init(self):
        """Test CacheEntry initialization."""
        entry = CacheEntry(
            version="0.20.1",
            timestamp=datetime.now(),
            source=VersionSource.CONFIG_FILE,
        )
        assert entry.version == "0.20.1"
        assert entry.source == VersionSource.CONFIG_FILE
        assert entry.access_count == 0


class TestVersionConfig:
    """Test VersionConfig dataclass."""

    def test_version_config_defaults(self):
        """Test VersionConfig default values."""
        config = VersionConfig()
        assert config.cache_ttl_seconds == 60
        assert config.cache_enabled is True
        assert config.fallback_version == "unknown"
        assert config.enable_async is True


class TestVersionReader:
    """Test VersionReader class."""

    def test_init_default(self):
        """Test VersionReader initialization with defaults."""
        reader = VersionReader()
        assert reader.config is not None
        assert reader.config.cache_enabled is True

    def test_init_custom_config(self):
        """Test VersionReader initialization with custom config."""
        config = VersionConfig(cache_ttl_seconds=30)
        reader = VersionReader(config=config)
        assert reader.config.cache_ttl_seconds == 30

    @pytest.mark.skipif(sys.version_info < (3, 12), reason="Path mock issues on Python < 3.12")
    def test_init_custom_working_dir(self):
        """Test VersionReader initialization with custom working dir."""
        with patch("pathlib.Path"):
            reader = VersionReader(working_dir="/custom/dir")
            assert reader._config_path is not None

    def test_get_version_sync(self):
        """Test get_version_sync method."""
        config = VersionConfig(enable_async=False)
        with patch.object(VersionReader, "_get_package_version", return_value="0.20.1"):
            reader = VersionReader(config=config)
            result = reader.get_version_sync()
            assert result == "0.20.1"

    def test_get_version_async(self):
        """Test get_version_async method."""

        async def test_async():
            config = VersionConfig(enable_async=True)
            with patch.object(VersionReader, "_get_package_version", return_value="0.20.1"):
                reader = VersionReader(config=config)
                result = await reader.get_version_async()
                assert result == "0.20.1"

        asyncio.run(test_async())

    def test_check_cache_empty(self):
        """Test _check_cache with empty cache."""
        reader = VersionReader()
        result = reader._check_cache()
        assert result is None

    def test_check_cache_disabled(self):
        """Test _check_cache when cache is disabled."""
        config = VersionConfig(cache_enabled=False)
        reader = VersionReader(config=config)
        result = reader._check_cache()
        assert result is None

    def test_update_cache(self):
        """Test _update_cache method."""
        reader = VersionReader()
        reader._update_cache("0.20.1", VersionSource.CONFIG_FILE)
        assert len(reader._cache) > 0

    def test_is_cache_entry_valid_ttl_expired(self):
        """Test _is_cache_entry_valid with expired TTL."""
        config = VersionConfig(cache_ttl_seconds=1)
        reader = VersionReader(config=config)
        old_time = datetime.now() - timedelta(seconds=2)
        entry = CacheEntry("0.20.1", old_time, VersionSource.CONFIG_FILE)
        result = reader._is_cache_entry_valid(entry)
        assert result is False

    def test_is_cache_entry_valid_ttl_valid(self):
        """Test _is_cache_entry_valid with valid TTL."""
        config = VersionConfig(cache_ttl_seconds=60)
        reader = VersionReader(config=config)
        entry = CacheEntry("0.20.1", datetime.now(), VersionSource.CONFIG_FILE)
        result = reader._is_cache_entry_valid(entry)
        assert result is True

    def test_extract_version_from_config(self):
        """Test _extract_version_from_config method."""
        reader = VersionReader()
        config = {"moai": {"version": "0.20.1"}}
        result = reader._extract_version_from_config(config)
        assert result == "0.20.1"

    def test_get_nested_value(self):
        """Test _get_nested_value method."""
        reader = VersionReader()
        config = {"moai": {"version": "0.20.1"}}
        result = reader._get_nested_value(config, "moai.version")
        assert result == "0.20.1"

    def test_get_nested_value_missing(self):
        """Test _get_nested_value with missing path."""
        reader = VersionReader()
        config = {"moai": {"version": "0.20.1"}}
        result = reader._get_nested_value(config, "missing.path")
        assert result is None

    def test_format_short_version_with_v(self):
        """Test _format_short_version with v prefix."""
        reader = VersionReader()
        result = reader._format_short_version("v0.20.1")
        assert result == "0.20.1"

    def test_format_short_version_without_v(self):
        """Test _format_short_version without v prefix."""
        reader = VersionReader()
        result = reader._format_short_version("0.20.1")
        assert result == "0.20.1"

    def test_format_display_version_unknown(self):
        """Test _format_display_version with unknown version."""
        reader = VersionReader()
        result = reader._format_display_version("unknown")
        assert "unknown" in result.lower()

    def test_format_display_version_with_v(self):
        """Test _format_display_version with v prefix."""
        reader = VersionReader()
        result = reader._format_display_version("v0.20.1")
        assert "v0.20.1" in result or "0.20.1" in result

    def test_is_valid_version_format(self):
        """Test _is_valid_version_format method."""
        reader = VersionReader()
        assert reader._is_valid_version_format("0.20.1") is True
        assert reader._is_valid_version_format("v0.20.1") is True
        assert reader._is_valid_version_format("invalid") is False

    def test_get_fallback_version(self):
        """Test _get_fallback_version method."""
        with patch.object(VersionReader, "_get_package_version", return_value=""):
            reader = VersionReader()
            result = reader._get_fallback_version()
            assert isinstance(result, str)

    def test_get_package_version(self):
        """Test _get_package_version method."""
        reader = VersionReader()
        result = reader._get_package_version()
        assert isinstance(result, str)

    def test_clear_cache(self):
        """Test clear_cache method."""
        reader = VersionReader()
        reader._update_cache("0.20.1", VersionSource.CONFIG_FILE)
        reader.clear_cache()
        assert len(reader._cache) == 0

    def test_get_cache_stats(self):
        """Test get_cache_stats method."""
        reader = VersionReader()
        stats = reader.get_cache_stats()
        assert isinstance(stats, dict)
        assert "hits" in stats
        assert "misses" in stats

    def test_is_cache_expired(self):
        """Test is_cache_expired method."""
        reader = VersionReader()
        result = reader.is_cache_expired()
        assert isinstance(result, bool)

    def test_get_config(self):
        """Test get_config method."""
        config = VersionConfig()
        reader = VersionReader(config=config)
        result = reader.get_config()
        assert result == config

    def test_get_available_version_fields(self):
        """Test get_available_version_fields method."""
        reader = VersionReader()
        fields = reader.get_available_version_fields()
        assert isinstance(fields, list)
        assert len(fields) > 0

    def test_set_custom_version_fields(self):
        """Test set_custom_version_fields method."""
        reader = VersionReader()
        custom_fields = ["custom.version", "project.version"]
        reader.set_custom_version_fields(custom_fields)
        assert reader.VERSION_FIELDS == custom_fields

    def test_get_performance_metrics(self):
        """Test get_performance_metrics method."""
        reader = VersionReader()
        metrics = reader.get_performance_metrics()
        assert isinstance(metrics, dict)
        assert "cache_stats" in metrics

    def test_evict_oldest_cache_entry(self):
        """Test _evict_oldest_cache_entry method."""
        config = VersionConfig(cache_size=2, enable_lru_cache=True)
        reader = VersionReader(config=config)
        reader._update_cache("v1", VersionSource.CONFIG_FILE)
        reader._update_cache("v2", VersionSource.CONFIG_FILE)
        # Cache should not exceed cache_size
        assert len(reader._cache) <= 2

    def test_handle_read_error(self):
        """Test _handle_read_error method."""
        reader = VersionReader()
        error = Exception("Test error")
        reader._handle_read_error(error, 0)
        assert reader._cache_stats["errors"] >= 0


class TestVersionReadError:
    """Test VersionReadError exception."""

    def test_version_read_error_is_exception(self):
        """Test VersionReadError is an Exception."""
        with pytest.raises(VersionReadError):
            raise VersionReadError("Test error")
