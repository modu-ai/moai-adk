"""Tests for statusline.memory_collector module."""

from datetime import datetime, timedelta
from unittest.mock import MagicMock, patch

from moai_adk.statusline.memory_collector import (
    MemoryCollector,
    MemoryInfo,
    get_memory_collector,
    get_memory_display,
)


class TestMemoryInfo:
    """Test MemoryInfo dataclass."""

    def test_memory_info_creation(self):
        """Test creating MemoryInfo instance."""
        info = MemoryInfo(
            process_rss_mb=128.0,
            process_vms_mb=500.0,
            system_total_mb=16384.0,
            system_available_mb=8192.0,
            system_percent=50.0,
            display_process="128MB",
            display_system="8.0GB/16GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )
        assert info.process_rss_mb == 128.0
        assert info.display_process == "128MB"
        assert info.system_percent == 50.0


class TestMemoryCollector:
    """Test MemoryCollector class."""

    def test_initialization_default_ttl(self):
        """Test initialization with default TTL."""
        with patch("moai_adk.statusline.memory_collector.MemoryCollector._check_psutil", return_value=True):
            collector = MemoryCollector()
            assert collector._cache_ttl == timedelta(seconds=10)
            assert collector._cache is None
            assert collector._cache_time is None

    def test_initialization_custom_ttl(self):
        """Test initialization with custom TTL."""
        with patch("moai_adk.statusline.memory_collector.MemoryCollector._check_psutil", return_value=True):
            collector = MemoryCollector(cache_ttl_seconds=30)
            assert collector._cache_ttl == timedelta(seconds=30)

    def test_check_psutil_available(self):
        """Test psutil availability check when available."""
        with patch("builtins.__import__", return_value=MagicMock()):
            collector = MemoryCollector()
            assert collector._psutil_available is True

    def test_check_psutil_not_available(self):
        """Test psutil availability check when not available."""
        with patch("builtins.__import__", side_effect=ImportError("No module")):
            collector = MemoryCollector()
            assert collector._psutil_available is False

    def test_get_memory_info_psutil_not_available(self):
        """Test get_memory_info when psutil is not available."""
        with patch("builtins.__import__", side_effect=ImportError("No module")):
            collector = MemoryCollector()
            result = collector.get_memory_info()
            assert result is None

    def test_get_memory_info_uses_cache(self):
        """Test get_memory_info uses valid cache."""
        mock_info = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=400.0,
            system_total_mb=16000.0,
            system_available_mb=8000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="8.0GB/16GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )

        with patch("moai_adk.statusline.memory_collector.MemoryCollector._check_psutil", return_value=True):
            collector = MemoryCollector()
            collector._cache = mock_info
            collector._cache_time = datetime.now()

            result = collector.get_memory_info()
            assert result is mock_info

    def test_get_memory_info_force_refresh(self):
        """Test get_memory_info with force_refresh bypasses cache."""
        mock_info = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=400.0,
            system_total_mb=16000.0,
            system_available_mb=8000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="8.0GB/16GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )

        with patch("moai_adk.statusline.memory_collector.MemoryCollector._check_psutil", return_value=True):
            with patch(
                "moai_adk.statusline.memory_collector.MemoryCollector._collect_memory_info", return_value=mock_info
            ):
                collector = MemoryCollector()
                collector._cache = mock_info
                collector._cache_time = datetime.now()

                result = collector.get_memory_info(force_refresh=True)
                # Should have collected fresh info (cache time updated)
                assert result.process_rss_mb == 100.0

    def test_get_memory_info_cache_expired(self):
        """Test get_memory_info when cache has expired."""
        old_info = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=400.0,
            system_total_mb=16000.0,
            system_available_mb=8000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="8.0GB/16GB",
            display_percent="50%",
            timestamp=datetime.now() - timedelta(seconds=20),
        )

        new_info = MemoryInfo(
            process_rss_mb=200.0,
            process_vms_mb=500.0,
            system_total_mb=16000.0,
            system_available_mb=8000.0,
            system_percent=50.0,
            display_process="200MB",
            display_system="8.0GB/16GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )

        with patch("moai_adk.statusline.memory_collector.MemoryCollector._check_psutil", return_value=True):
            with patch(
                "moai_adk.statusline.memory_collector.MemoryCollector._collect_memory_info", return_value=new_info
            ):
                collector = MemoryCollector(cache_ttl_seconds=10)
                collector._cache = old_info
                collector._cache_time = datetime.now() - timedelta(seconds=15)

                result = collector.get_memory_info()
                assert result.process_rss_mb == 200.0

    def test_get_display_string_process_mode(self):
        """Test get_display_string in process mode."""
        mock_info = MemoryInfo(
            process_rss_mb=128.0,
            process_vms_mb=500.0,
            system_total_mb=16000.0,
            system_available_mb=8000.0,
            system_percent=50.0,
            display_process="128MB",
            display_system="8.0GB/16GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )

        with patch("moai_adk.statusline.memory_collector.MemoryCollector.get_memory_info", return_value=mock_info):
            collector = MemoryCollector()
            result = collector.get_display_string(mode="process")
            assert result == "128MB"

    def test_get_display_string_system_mode(self):
        """Test get_display_string in system mode."""
        mock_info = MemoryInfo(
            process_rss_mb=128.0,
            process_vms_mb=500.0,
            system_total_mb=16000.0,
            system_available_mb=8000.0,
            system_percent=50.0,
            display_process="128MB",
            display_system="8.0GB/16GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )

        with patch("moai_adk.statusline.memory_collector.MemoryCollector.get_memory_info", return_value=mock_info):
            collector = MemoryCollector()
            result = collector.get_display_string(mode="system")
            assert result == "8.0GB/16GB"

    def test_get_display_string_percent_mode(self):
        """Test get_display_string in percent mode."""
        mock_info = MemoryInfo(
            process_rss_mb=128.0,
            process_vms_mb=500.0,
            system_total_mb=16000.0,
            system_available_mb=8000.0,
            system_percent=50.0,
            display_process="128MB",
            display_system="8.0GB/16GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )

        with patch("moai_adk.statusline.memory_collector.MemoryCollector.get_memory_info", return_value=mock_info):
            collector = MemoryCollector()
            result = collector.get_display_string(mode="percent")
            assert result == "50%"

    def test_get_display_string_compact_mode(self):
        """Test get_display_string in compact mode."""
        mock_info = MemoryInfo(
            process_rss_mb=128.0,
            process_vms_mb=500.0,
            system_total_mb=16000.0,
            system_available_mb=8000.0,
            system_percent=50.0,
            display_process="128MB",
            display_system="8.0GB/16GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )

        with patch("moai_adk.statusline.memory_collector.MemoryCollector.get_memory_info", return_value=mock_info):
            collector = MemoryCollector()
            result = collector.get_display_string(mode="compact")
            assert result == "128M"

    def test_get_display_string_no_info(self):
        """Test get_display_string when memory info is None."""
        with patch("moai_adk.statusline.memory_collector.MemoryCollector.get_memory_info", return_value=None):
            collector = MemoryCollector()
            result = collector.get_display_string(mode="process")
            assert result == "N/A"

    def test_format_size_gb(self):
        """Test _format_size for GB sizes."""
        collector = MemoryCollector()
        result = collector._format_size(2048.0)  # 2GB
        assert result == "2.0GB"

    def test_format_size_large_mb(self):
        """Test _format_size for large MB sizes (>= 100)."""
        collector = MemoryCollector()
        result = collector._format_size(500.0)
        assert result == "500MB"

    def test_format_size_medium_mb(self):
        """Test _format_size for medium MB sizes (>= 10)."""
        collector = MemoryCollector()
        result = collector._format_size(50.0)
        assert result == "50.0MB"

    def test_format_size_small_mb(self):
        """Test _format_size for small MB sizes (< 10)."""
        collector = MemoryCollector()
        result = collector._format_size(5.5)
        assert result == "5.50MB"

    def test_format_compact_gb(self):
        """Test _format_compact for GB sizes."""
        collector = MemoryCollector()
        result = collector._format_compact(2048.0)  # 2GB
        assert result == "2.0G"

    def test_format_compact_mb(self):
        """Test _format_compact for MB sizes."""
        collector = MemoryCollector()
        result = collector._format_compact(512.0)
        assert result == "512M"

    def test_is_cache_valid_no_cache(self):
        """Test _is_cache_valid when no cache exists."""
        collector = MemoryCollector()
        assert collector._is_cache_valid() is False

    def test_is_cache_valid_no_cache_time(self):
        """Test _is_cache_valid when cache exists but no time."""
        collector = MemoryCollector()
        mock_info = MagicMock()
        collector._cache = mock_info
        collector._cache_time = None
        assert collector._is_cache_valid() is False

    def test_is_cache_valid_fresh(self):
        """Test _is_cache_valid with fresh cache."""
        collector = MemoryCollector(cache_ttl_seconds=10)
        mock_info = MagicMock()
        collector._cache = mock_info
        collector._cache_time = datetime.now() - timedelta(seconds=5)
        assert collector._is_cache_valid() is True

    def test_is_cache_valid_expired(self):
        """Test _is_cache_valid with expired cache."""
        collector = MemoryCollector(cache_ttl_seconds=10)
        mock_info = MagicMock()
        collector._cache = mock_info
        collector._cache_time = datetime.now() - timedelta(seconds=15)
        assert collector._is_cache_valid() is False

    def test_update_cache(self):
        """Test _update_cache updates cache and time."""
        collector = MemoryCollector()
        mock_info = MagicMock()
        collector._update_cache(mock_info)
        assert collector._cache is mock_info
        assert collector._cache_time is not None

    def test_clear_cache(self):
        """Test clear_cache clears cache and time."""
        collector = MemoryCollector()
        mock_info = MagicMock()
        collector._cache = mock_info
        collector._cache_time = datetime.now()

        collector.clear_cache()
        assert collector._cache is None
        assert collector._cache_time is None

    def test_get_cache_age_no_cache(self):
        """Test get_cache_age_seconds when no cache."""
        collector = MemoryCollector()
        assert collector.get_cache_age_seconds() is None

    def test_get_cache_age_with_cache(self):
        """Test get_cache_age_seconds with cache."""
        collector = MemoryCollector()
        mock_info = MagicMock()
        collector._cache = mock_info
        collector._cache_time = datetime.now() - timedelta(seconds=5)

        age = collector.get_cache_age_seconds()
        assert age >= 5.0
        assert age < 6.0  # Should be close to 5 seconds


class TestModuleFunctions:
    """Test module-level functions."""

    def test_get_memory_collector_singleton(self):
        """Test that get_memory_collector returns singleton."""
        with patch("moai_adk.statusline.memory_collector.MemoryCollector._check_psutil", return_value=True):
            collector1 = get_memory_collector()
            collector2 = get_memory_collector()
            assert collector1 is collector2

    def test_get_memory_display_process(self):
        """Test get_memory_display in process mode."""
        mock_info = MemoryInfo(
            process_rss_mb=128.0,
            process_vms_mb=500.0,
            system_total_mb=16000.0,
            system_available_mb=8000.0,
            system_percent=50.0,
            display_process="128MB",
            display_system="8.0GB/16GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )

        with patch("moai_adk.statusline.memory_collector.get_memory_collector") as mock_get:
            mock_collector = MagicMock()
            mock_collector.get_display_string.return_value = "128MB"
            mock_get.return_value = mock_collector

            result = get_memory_display(mode="process")
            assert result == "128MB"
            mock_collector.get_display_string.assert_called_once_with("process")
