"""
Comprehensive test coverage for memory collector module.

Tests for uncovered lines in memory_collector.py:
- psutil not available path (lines 74-76)
- get_memory_info when psutil unavailable (line 89)
- get_memory_info with force_refresh (line 93)
- get_display_string modes (lines 121, 125-133)
- _format_size edge cases (lines 187-194)
- _format_compact (lines 206-209)
- clear_cache (lines 224-227)
- get_cache_age_seconds (lines 236-238)
- Module-level functions (lines 252-255, 264-268)
"""

from datetime import datetime, timedelta
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.statusline.memory_collector import (
    MemoryCollector,
    MemoryInfo,
    get_memory_collector,
    get_memory_display,
)


class TestMemoryCollectorInit:
    """Test MemoryCollector initialization."""

    def test_init_with_default_cache_ttl(self):
        """Test initialization with default cache TTL."""
        # Arrange & Act
        collector = MemoryCollector()

        # Assert
        assert collector._cache_ttl.total_seconds() == 10
        assert collector._cache is None
        assert collector._cache_time is None

    def test_init_with_custom_cache_ttl(self):
        """Test initialization with custom cache TTL."""
        # Arrange & Act
        collector = MemoryCollector(cache_ttl_seconds=30)

        # Assert
        assert collector._cache_ttl.total_seconds() == 30

    def test_init_checks_psutil_available(self):
        """Test that initialization checks for psutil availability."""
        # Arrange & Act
        collector = MemoryCollector()

        # Assert
        assert hasattr(collector, "_psutil_available")
        assert isinstance(collector._psutil_available, bool)


class TestCheckPsutil:
    """Test _check_psutil method (lines 68-76)."""

    def test_check_psutil_available(self):
        """Test when psutil is available."""
        # Arrange & Act
        with patch("moai_adk.statusline.memory_collector.importlib.import_module") as mock_import:
            mock_import.return_value = MagicMock()  # psutil module

            collector = MemoryCollector()

        # Assert
        assert collector._psutil_available is True

    def test_check_psutil_not_available(self):
        """Test when psutil is not available (lines 74-76)."""
        # Arrange & Act
        with patch("moai_adk.statusline.memory_collector.importlib.import_module") as mock_import:
            mock_import.side_effect = ImportError("No module named 'psutil'")

            collector = MemoryCollector()

        # Assert
        assert collector._psutil_available is False


class TestGetMemoryInfo:
    """Test get_memory_info method."""

    def test_get_memory_info_when_psutil_unavailable(self):
        """Test get_memory_info when psutil is not available (line 89)."""
        # Arrange
        collector = MemoryCollector()
        collector._psutil_available = False

        # Act
        result = collector.get_memory_info()

        # Assert
        assert result is None

    def test_get_memory_info_returns_cache_when_valid(self):
        """Test get_memory_info returns cached value when valid."""
        # Arrange
        collector = MemoryCollector()
        collector._psutil_available = True

        cached_info = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=200.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )
        collector._cache = cached_info
        collector._cache_time = datetime.now() - timedelta(seconds=5)

        # Act
        result = collector.get_memory_info()

        # Assert
        assert result == cached_info

    def test_get_memory_info_force_refresh_bypasses_cache(self):
        """Test get_memory_info with force_refresh bypasses cache (line 93)."""
        # Arrange
        collector = MemoryCollector()
        collector._psutil_available = True

        cached_info = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=200.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )
        collector._cache = cached_info
        collector._cache_time = datetime.now() - timedelta(seconds=5)

        with patch.object(collector, "_collect_memory_info") as mock_collect:
            mock_collect.return_value = cached_info

            # Act
            result = collector.get_memory_info(force_refresh=True)

        # Assert
        mock_collect.assert_called_once()
        assert result == cached_info

    def test_get_memory_info_collects_when_cache_expired(self):
        """Test get_memory_info collects when cache is expired."""
        # Arrange
        collector = MemoryCollector()
        collector._psutil_available = True

        cached_info = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=200.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )
        collector._cache = cached_info
        collector._cache_time = datetime.now() - timedelta(seconds=15)  # Expired

        with patch.object(collector, "_collect_memory_info") as mock_collect:
            mock_collect.return_value = cached_info

            # Act
            _ = collector.get_memory_info()

        # Assert
        mock_collect.assert_called_once()

    def test_get_memory_info_updates_cache(self):
        """Test get_memory_info updates cache after collection."""
        # Arrange
        collector = MemoryCollector()
        collector._psutil_available = True

        new_info = MemoryInfo(
            process_rss_mb=150.0,
            process_vms_mb=250.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="150MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )

        with patch.object(collector, "_collect_memory_info", return_value=new_info):
            # Act
            result = collector.get_memory_info()

        # Assert
        assert collector._cache == new_info
        assert result == new_info

    def test_get_memory_info_returns_stale_cache_on_error(self):
        """Test get_memory_info returns stale cache on collection error (line 101)."""
        # Arrange
        collector = MemoryCollector()
        collector._psutil_available = True

        cached_info = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=200.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )
        collector._cache = cached_info
        collector._cache_time = datetime.now() - timedelta(seconds=15)  # Expired

        with patch.object(collector, "_collect_memory_info", side_effect=RuntimeError("Collection error")):
            # Act
            result = collector.get_memory_info()

        # Assert - should return stale cache
        assert result == cached_info


class TestGetDisplayString:
    """Test get_display_string method (lines 103-133)."""

    def test_get_display_string_when_psutil_unavailable(self):
        """Test get_display_string when psutil is not available."""
        # Arrange
        collector = MemoryCollector()
        collector._psutil_available = False

        # Act
        result = collector.get_display_string()

        # Assert
        assert result == "N/A"

    def test_get_display_string_when_memory_info_none(self):
        """Test get_display_string when memory_info is None (line 121)."""
        # Arrange
        collector = MemoryCollector()
        collector._psutil_available = False

        # Act
        result = collector.get_display_string()

        # Assert
        assert result == "N/A"

    def test_get_display_string_mode_process(self):
        """Test get_display_string with mode='process' (line 124)."""
        # Arrange
        collector = MemoryCollector()
        collector._psutil_available = True

        memory_info = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=200.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )
        collector._cache = memory_info

        # Act
        result = collector.get_display_string(mode="process")

        # Assert
        assert result == "100MB"

    def test_get_display_string_mode_system(self):
        """Test get_display_string with mode='system' (line 126)."""
        # Arrange
        collector = MemoryCollector()
        collector._psutil_available = True

        memory_info = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=200.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )
        collector._cache = memory_info

        # Act
        result = collector.get_display_string(mode="system")

        # Assert
        assert result == "4.0GB/8.0GB"

    def test_get_display_string_mode_percent(self):
        """Test get_display_string with mode='percent' (line 128)."""
        # Arrange
        collector = MemoryCollector()
        collector._psutil_available = True

        memory_info = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=200.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )
        collector._cache = memory_info

        # Act
        result = collector.get_display_string(mode="percent")

        # Assert
        assert result == "50%"

    def test_get_display_string_mode_compact(self):
        """Test get_display_string with mode='compact' (lines 129-131)."""
        # Arrange
        collector = MemoryCollector()
        collector._psutil_available = True

        memory_info = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=200.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )
        collector._cache = memory_info

        # Act
        result = collector.get_display_string(mode="compact")

        # Assert - should format without "MB" suffix
        assert result == "100M"

    def test_get_display_string_mode_unknown_defaults_to_process(self):
        """Test get_display_string with unknown mode defaults to process (lines 132-133)."""
        # Arrange
        collector = MemoryCollector()
        collector._psutil_available = True

        memory_info = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=200.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )
        collector._cache = memory_info

        # Act
        result = collector.get_display_string(mode="unknown")

        # Assert - should default to process mode
        assert result == "100MB"

    def test_get_display_string_with_force_refresh(self):
        """Test get_display_string with force_refresh parameter."""
        # Arrange
        collector = MemoryCollector()
        collector._psutil_available = True

        memory_info = MemoryInfo(
            process_rss_mb=150.0,
            process_vms_mb=250.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="150MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )

        with patch.object(collector, "get_memory_info", return_value=memory_info) as mock_get:
            # Act
            result = collector.get_display_string(mode="process", force_refresh=True)

        # Assert
        mock_get.assert_called_once_with(force_refresh=True)
        assert result == "150MB"


class TestCollectMemoryInfo:
    """Test _collect_memory_info method."""

    @pytest.fixture
    def mock_psutil(self):
        """Create mock psutil module."""
        with patch("moai_adk.statusline.memory_collector.psutil") as mock:
            yield mock

    def test_collect_memory_info_success(self, mock_psutil):
        """Test successful memory info collection."""
        # Arrange
        mock_process = MagicMock()
        mock_process.memory_info.return_value = MagicMock(rss=100 * 1024 * 1024, vms=200 * 1024 * 1024)  # 100MB, 200MB
        mock_psutil.Process.return_value = mock_process

        mock_vmem = MagicMock()
        mock_vmem.total = 8000 * 1024 * 1024  # 8GB
        mock_vmem.available = 4000 * 1024 * 1024  # 4GB
        mock_vmem.percent = 50.0
        mock_psutil.virtual_memory.return_value = mock_vmem

        collector = MemoryCollector()
        collector._psutil_available = True

        # Act
        result = collector._collect_memory_info()

        # Assert
        assert result is not None
        assert result.process_rss_mb == pytest.approx(100.0, rel=0.1)
        assert result.process_vms_mb == pytest.approx(200.0, rel=0.1)
        assert result.system_total_mb == pytest.approx(8000.0, rel=0.1)
        assert result.system_percent == 50.0
        assert "MB" in result.display_process or "GB" in result.display_process

    def test_collect_memory_info_formats_process_display(self, mock_psutil):
        """Test that process memory is formatted correctly."""
        # Arrange
        mock_process = MagicMock()
        mock_process.memory_info.return_value = MagicMock(rss=128 * 1024 * 1024, vms=256 * 1024 * 1024)
        mock_psutil.Process.return_value = mock_process

        mock_vmem = MagicMock()
        mock_vmem.total = 8000 * 1024 * 1024
        mock_vmem.available = 4000 * 1024 * 1024
        mock_vmem.percent = 50.0
        mock_psutil.virtual_memory.return_value = mock_vmem

        collector = MemoryCollector()
        collector._psutil_available = True

        # Act
        result = collector._collect_memory_info()

        # Assert
        assert result.display_process == "128MB"


class TestFormatSize:
    """Test _format_size method (lines 177-194)."""

    def test_format_size_gb(self):
        """Test formatting size in GB (lines 187-188)."""
        # Arrange
        collector = MemoryCollector()

        # Act
        result = collector._format_size(2048.0)  # 2GB

        # Assert
        assert result == "2.0GB"

    def test_format_size_large_mb(self):
        """Test formatting large MB value (lines 189-190)."""
        # Arrange
        collector = MemoryCollector()

        # Act
        result = collector._format_size(150.0)  # 150MB

        # Assert
        assert result == "150MB"

    def test_format_size_medium_mb(self):
        """Test formatting medium MB value (lines 191-192)."""
        # Arrange
        collector = MemoryCollector()

        # Act
        result = collector._format_size(15.5)  # 15.5MB

        # Assert
        assert result == "15.5MB"

    def test_format_size_small_mb(self):
        """Test formatting small MB value (lines 193-194)."""
        # Arrange
        collector = MemoryCollector()

        # Act
        result = collector._format_size(5.25)  # 5.25MB

        # Assert
        assert result == "5.25MB"

    def test_format_size_exactly_1024_mb(self):
        """Test formatting exactly 1024 MB (boundary case)."""
        # Arrange
        collector = MemoryCollector()

        # Act
        result = collector._format_size(1024.0)

        # Assert
        assert result == "1.0GB"

    def test_format_size_exactly_100_mb(self):
        """Test formatting exactly 100 MB (boundary case)."""
        # Arrange
        collector = MemoryCollector()

        # Act
        result = collector._format_size(100.0)

        # Assert
        assert result == "100MB"

    def test_format_size_exactly_10_mb(self):
        """Test formatting exactly 10 MB (boundary case)."""
        # Arrange
        collector = MemoryCollector()

        # Act
        result = collector._format_size(10.0)

        # Assert
        assert result == "10.0MB"


class TestFormatCompact:
    """Test _format_compact method (lines 196-209)."""

    def test_format_compact_gb(self):
        """Test formatting compact GB (lines 206-207)."""
        # Arrange
        collector = MemoryCollector()

        # Act
        result = collector._format_compact(2048.0)  # 2GB

        # Assert
        assert result == "2.0G"

    def test_format_compact_mb(self):
        """Test formatting compact MB (line 209)."""
        # Arrange
        collector = MemoryCollector()

        # Act
        result = collector._format_compact(150.0)  # 150MB

        # Assert
        assert result == "150M"

    def test_format_compact_exactly_1024_mb(self):
        """Test formatting compact exactly 1024 MB (boundary case)."""
        # Arrange
        collector = MemoryCollector()

        # Act
        result = collector._format_compact(1024.0)

        # Assert
        assert result == "1.0G"


class TestCacheValidation:
    """Test cache validation methods."""

    def test_is_cache_valid_when_cache_none(self):
        """Test cache validation when cache is None (lines 213-214)."""
        # Arrange
        collector = MemoryCollector()
        collector._cache = None
        collector._cache_time = None

        # Act
        result = collector._is_cache_valid()

        # Assert
        assert result is False

    def test_is_cache_valid_when_cache_time_none(self):
        """Test cache validation when cache_time is None."""
        # Arrange
        collector = MemoryCollector()
        collector._cache = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=200.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )
        collector._cache_time = None

        # Act
        result = collector._is_cache_valid()

        # Assert
        assert result is False

    def test_is_cache_valid_within_ttl(self):
        """Test cache validation within TTL."""
        # Arrange
        collector = MemoryCollector()
        collector._cache = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=200.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )
        collector._cache_time = datetime.now() - timedelta(seconds=5)

        # Act
        result = collector._is_cache_valid()

        # Assert
        assert result is True

    def test_is_cache_valid_expired(self):
        """Test cache validation when expired."""
        # Arrange
        collector = MemoryCollector()
        collector._cache = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=200.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )
        collector._cache_time = datetime.now() - timedelta(seconds=15)

        # Act
        result = collector._is_cache_valid()

        # Assert
        assert result is False

    def test_update_cache(self):
        """Test cache update (lines 219-222)."""
        # Arrange
        collector = MemoryCollector()

        memory_info = MemoryInfo(
            process_rss_mb=150.0,
            process_vms_mb=250.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="150MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )

        # Act
        collector._update_cache(memory_info)

        # Assert
        assert collector._cache == memory_info
        assert collector._cache_time is not None


class TestClearCache:
    """Test clear_cache method (lines 224-227)."""

    def test_clear_cache(self):
        """Test clearing cache."""
        # Arrange
        collector = MemoryCollector()
        collector._cache = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=200.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )
        collector._cache_time = datetime.now()

        # Act
        collector.clear_cache()

        # Assert
        assert collector._cache is None
        assert collector._cache_time is None


class TestGetCacheAgeSeconds:
    """Test get_cache_age_seconds method (lines 229-238)."""

    def test_get_cache_age_seconds_when_no_cache(self):
        """Test get_cache_age_seconds when cache_time is None (lines 236-238)."""
        # Arrange
        collector = MemoryCollector()
        collector._cache_time = None

        # Act
        result = collector.get_cache_age_seconds()

        # Assert
        assert result is None

    def test_get_cache_age_seconds_with_cache(self):
        """Test get_cache_age_seconds with cache."""
        # Arrange
        collector = MemoryCollector()
        collector._cache_time = datetime.now() - timedelta(seconds=10)

        # Act
        result = collector.get_cache_age_seconds()

        # Assert
        assert result is not None
        assert result >= 9.0  # Allow some timing variance


class TestModuleLevelFunctions:
    """Test module-level functions (lines 252-255, 264-268)."""

    def test_get_memory_collector_singleton(self):
        """Test get_memory_collector returns singleton (lines 252-255)."""
        # Act
        collector1 = get_memory_collector()
        collector2 = get_memory_collector()

        # Assert
        assert collector1 is collector2

    def test_get_memory_collector_returns_instance(self):
        """Test get_memory_collector returns MemoryCollector instance."""
        # Act
        collector = get_memory_collector()

        # Assert
        assert isinstance(collector, MemoryCollector)

    def test_get_memory_display_process_mode(self):
        """Test get_memory_display with process mode (lines 264-268)."""
        # Arrange
        with patch("moai_adk.statusline.memory_collector.get_memory_collector") as mock_get:
            mock_collector = MagicMock()
            mock_collector.get_display_string.return_value = "100MB"
            mock_get.return_value = mock_collector

            # Act
            result = get_memory_display(mode="process")

        # Assert
        mock_collector.get_display_string.assert_called_once_with("process")
        assert result == "100MB"

    def test_get_memory_display_default_mode(self):
        """Test get_memory_display with default mode."""
        # Arrange
        with patch("moai_adk.statusline.memory_collector.get_memory_collector") as mock_get:
            mock_collector = MagicMock()
            mock_collector.get_display_string.return_value = "150M"
            mock_get.return_value = mock_collector

            # Act
            result = get_memory_display()

        # Assert
        mock_collector.get_display_string.assert_called_once_with("process")
        assert result == "150M"


class TestMemoryInfoDataclass:
    """Test MemoryInfo dataclass."""

    def test_memory_info_creation(self):
        """Test MemoryInfo dataclass creation."""
        # Act
        info = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=200.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=datetime.now(),
        )

        # Assert
        assert info.process_rss_mb == 100.0
        assert info.process_vms_mb == 200.0
        assert info.system_total_mb == 8000.0
        assert info.system_percent == 50.0

    def test_memory_info_equality(self):
        """Test MemoryInfo equality."""
        # Arrange
        timestamp = datetime.now()

        # Act
        info1 = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=200.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=timestamp,
        )
        info2 = MemoryInfo(
            process_rss_mb=100.0,
            process_vms_mb=200.0,
            system_total_mb=8000.0,
            system_available_mb=4000.0,
            system_percent=50.0,
            display_process="100MB",
            display_system="4.0GB/8.0GB",
            display_percent="50%",
            timestamp=timestamp,
        )

        # Assert
        assert info1 == info2


class TestEdgeCases:
    """Test edge cases and error conditions."""

    def test_collect_memory_info_with_os_error(self):
        """Test collection when os.getpid fails."""
        # Arrange
        collector = MemoryCollector()
        collector._psutil_available = True

        with patch("moai_adk.statusline.memory_collector.os.getpid", side_effect=OSError("Process error")):
            # Act & Assert - should raise exception
            with pytest.raises(OSError):
                collector._collect_memory_info()

    def test_format_size_zero(self):
        """Test formatting zero size."""
        # Arrange
        collector = MemoryCollector()

        # Act
        result = collector._format_size(0.0)

        # Assert
        assert result == "0.00MB"

    def test_format_compact_zero(self):
        """Test compact formatting zero size."""
        # Arrange
        collector = MemoryCollector()

        # Act
        result = collector._format_compact(0.0)

        # Assert
        assert result == "0M"
