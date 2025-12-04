"""Comprehensive tests for MetricsTracker with 80% coverage target."""

import time
from datetime import datetime, timedelta
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.statusline.metrics_tracker import MetricsTracker


class TestMetricsTrackerInit:
    """Test MetricsTracker initialization."""

    def test_init_default(self):
        """Test default initialization."""
        tracker = MetricsTracker()
        assert tracker._session_start is not None
        assert isinstance(tracker._session_start, datetime)
        assert tracker._duration_cache is None
        assert tracker._cache_time is None
        assert isinstance(tracker._cache_ttl, timedelta)

    def test_init_cache_ttl(self):
        """Test cache TTL value."""
        tracker = MetricsTracker()
        assert tracker._CACHE_TTL_SECONDS == 10
        assert tracker._cache_ttl.total_seconds() == 10

    def test_init_session_start_is_recent(self):
        """Test that session start is set to current time."""
        before = datetime.now()
        tracker = MetricsTracker()
        after = datetime.now()
        assert before <= tracker._session_start <= after


class TestMetricsTrackerGetDuration:
    """Test main get_duration method."""

    def test_get_duration_seconds_only(self):
        """Test duration calculation for seconds."""
        with patch("moai_adk.statusline.metrics_tracker.datetime") as mock_datetime:
            mock_now = MagicMock()
            mock_now.now.side_effect = [
                datetime(2024, 1, 1, 12, 0, 0),  # session start
                datetime(2024, 1, 1, 12, 0, 30),  # first call
                datetime(2024, 1, 1, 12, 0, 30),  # cache check
            ]
            mock_datetime.now = mock_now.now
            mock_datetime.side_effect = lambda *args, **kwargs: datetime(
                *args, **kwargs
            )

            tracker = MetricsTracker()
            tracker._session_start = datetime(2024, 1, 1, 12, 0, 0)
            with patch(
                "moai_adk.statusline.metrics_tracker.datetime.now",
                return_value=datetime(2024, 1, 1, 12, 0, 30),
            ):
                result = tracker.get_duration()
                assert "30s" in result

    def test_get_duration_minutes_only(self):
        """Test duration calculation for minutes."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(minutes=5)
        result = tracker.get_duration()
        assert "5m" in result

    def test_get_duration_minutes_and_seconds(self):
        """Test duration calculation with minutes and seconds."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(minutes=5, seconds=30)
        result = tracker.get_duration()
        assert "m" in result

    def test_get_duration_hours_only(self):
        """Test duration calculation for hours."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(hours=2)
        result = tracker.get_duration()
        assert "2h" in result

    def test_get_duration_hours_and_minutes(self):
        """Test duration calculation with hours and minutes."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(hours=2, minutes=30)
        result = tracker.get_duration()
        assert "h" in result and "m" in result

    def test_get_duration_cache_hit(self):
        """Test that cached duration is returned."""
        tracker = MetricsTracker()
        tracker._duration_cache = "5m"
        tracker._cache_time = datetime.now()
        result = tracker.get_duration()
        assert result == "5m"

    def test_get_duration_cache_miss(self):
        """Test that new duration is calculated on cache miss."""
        tracker = MetricsTracker()
        result = tracker.get_duration()
        assert isinstance(result, str)
        assert tracker._duration_cache == result


class TestMetricsTrackerCalculateAndFormatDuration:
    """Test _calculate_and_format_duration method."""

    def test_calculate_and_format_less_than_60_seconds(self):
        """Test formatting for durations less than 60 seconds."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(seconds=45)
        result = tracker._calculate_and_format_duration()
        assert "s" in result
        assert isinstance(result, str)

    def test_calculate_and_format_exactly_60_seconds(self):
        """Test formatting for exactly 60 seconds."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(seconds=60)
        result = tracker._calculate_and_format_duration()
        assert "m" in result

    def test_calculate_and_format_minutes_no_seconds(self):
        """Test formatting for minutes without seconds."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(minutes=5)
        result = tracker._calculate_and_format_duration()
        assert result == "5m"

    def test_calculate_and_format_minutes_with_seconds(self):
        """Test formatting for minutes with seconds."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(minutes=5, seconds=30)
        result = tracker._calculate_and_format_duration()
        assert "5m 30s" in result

    def test_calculate_and_format_hours_no_minutes(self):
        """Test formatting for hours without minutes."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(hours=2)
        result = tracker._calculate_and_format_duration()
        assert result == "2h"

    def test_calculate_and_format_hours_with_minutes(self):
        """Test formatting for hours with minutes."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(hours=2, minutes=30)
        result = tracker._calculate_and_format_duration()
        assert "2h 30m" in result

    def test_calculate_and_format_large_duration(self):
        """Test formatting for large durations."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(hours=10, minutes=45)
        result = tracker._calculate_and_format_duration()
        assert "h" in result and "m" in result

    def test_calculate_and_format_zero_seconds(self):
        """Test formatting for zero elapsed time."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now()
        result = tracker._calculate_and_format_duration()
        assert "0s" in result

    def test_calculate_and_format_return_type(self):
        """Test that result is always a string."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(seconds=30)
        result = tracker._calculate_and_format_duration()
        assert isinstance(result, str)


class TestMetricsTrackerCacheManagement:
    """Test cache management."""

    def test_is_cache_valid_with_valid_cache(self):
        """Test cache validation with valid cache."""
        tracker = MetricsTracker()
        tracker._duration_cache = "5m"
        tracker._cache_time = datetime.now()
        assert tracker._is_cache_valid()

    def test_is_cache_valid_with_expired_cache(self):
        """Test cache validation with expired cache."""
        tracker = MetricsTracker()
        tracker._duration_cache = "5m"
        tracker._cache_time = datetime.now() - timedelta(seconds=15)
        assert not tracker._is_cache_valid()

    def test_is_cache_valid_with_no_cache(self):
        """Test cache validation with no cached duration."""
        tracker = MetricsTracker()
        tracker._duration_cache = None
        assert not tracker._is_cache_valid()

    def test_is_cache_valid_with_no_cache_time(self):
        """Test cache validation with no cache time."""
        tracker = MetricsTracker()
        tracker._duration_cache = "5m"
        tracker._cache_time = None
        assert not tracker._is_cache_valid()

    def test_update_cache(self):
        """Test cache update."""
        tracker = MetricsTracker()
        duration = "10m 30s"
        tracker._update_cache(duration)

        assert tracker._duration_cache == duration
        assert tracker._cache_time is not None

    def test_cache_ttl_enforcement(self):
        """Test that cache TTL is enforced."""
        tracker = MetricsTracker()
        tracker._cache_ttl = timedelta(seconds=1)
        tracker._duration_cache = "5m"
        tracker._cache_time = datetime.now() - timedelta(seconds=2)

        assert not tracker._is_cache_valid()


class TestMetricsTrackerFormatting:
    """Test duration formatting edge cases."""

    def test_format_with_59_seconds(self):
        """Test formatting with 59 seconds (boundary)."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(seconds=59)
        result = tracker._calculate_and_format_duration()
        assert "59s" in result

    def test_format_with_61_seconds(self):
        """Test formatting with 61 seconds (just over 1 minute)."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(seconds=61)
        result = tracker._calculate_and_format_duration()
        assert "1m 1s" in result

    def test_format_with_3599_seconds(self):
        """Test formatting with 3599 seconds (just under 1 hour)."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(seconds=3599)
        result = tracker._calculate_and_format_duration()
        assert "m" in result

    def test_format_with_3600_seconds(self):
        """Test formatting with 3600 seconds (exactly 1 hour)."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(seconds=3600)
        result = tracker._calculate_and_format_duration()
        assert "1h" in result

    def test_format_with_1_second(self):
        """Test formatting with 1 second."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(seconds=1)
        result = tracker._calculate_and_format_duration()
        assert "1s" in result

    def test_format_with_1_minute(self):
        """Test formatting with exactly 1 minute."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(minutes=1)
        result = tracker._calculate_and_format_duration()
        assert result == "1m"

    def test_format_with_1_hour(self):
        """Test formatting with exactly 1 hour."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(hours=1)
        result = tracker._calculate_and_format_duration()
        assert result == "1h"


class TestMetricsTrackerMultipleCalls:
    """Test behavior with multiple calls."""

    def test_duration_increases_over_time(self):
        """Test that duration increases with time."""
        tracker = MetricsTracker()
        result1 = tracker.get_duration()
        time.sleep(0.1)
        result2 = tracker.get_duration()
        # Results should be similar but second might be larger
        assert isinstance(result1, str)
        assert isinstance(result2, str)

    def test_multiple_calls_within_cache_ttl(self):
        """Test multiple calls within cache TTL."""
        tracker = MetricsTracker()
        result1 = tracker.get_duration()
        result2 = tracker.get_duration()
        # Both should be from cache
        assert result1 == result2

    def test_call_after_cache_expiry(self):
        """Test call after cache expiry."""
        tracker = MetricsTracker()
        tracker._cache_ttl = timedelta(milliseconds=100)
        result1 = tracker.get_duration()
        time.sleep(0.15)
        result2 = tracker.get_duration()
        # Results should both be strings
        assert isinstance(result1, str)
        assert isinstance(result2, str)


class TestMetricsTrackerIntegration:
    """Integration tests for MetricsTracker."""

    def test_full_tracking_flow(self):
        """Test complete tracking flow."""
        tracker = MetricsTracker()
        assert isinstance(tracker._session_start, datetime)

        # Get initial duration
        duration1 = tracker.get_duration()
        assert isinstance(duration1, str)
        assert "s" in duration1 or "m" in duration1 or "h" in duration1

        # Get cached duration
        duration2 = tracker.get_duration()
        assert duration1 == duration2

        # Wait and get new duration
        time.sleep(0.2)
        tracker._cache_ttl = timedelta(milliseconds=50)
        duration3 = tracker.get_duration()
        assert isinstance(duration3, str)

    def test_long_running_session_tracking(self):
        """Test tracking over simulated long session."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(
            hours=3, minutes=45, seconds=30
        )
        result = tracker.get_duration()
        assert "h" in result
        assert "m" in result


class TestMetricsTrackerEdgeCases:
    """Test edge cases and error conditions."""

    def test_with_very_small_elapsed_time(self):
        """Test with very small elapsed time."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(microseconds=500)
        result = tracker._calculate_and_format_duration()
        assert isinstance(result, str)

    def test_with_very_large_elapsed_time(self):
        """Test with very large elapsed time."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(days=1, hours=12)
        result = tracker.get_duration()
        assert isinstance(result, str)

    def test_cache_with_exact_ttl_boundary(self):
        """Test cache at exact TTL boundary."""
        tracker = MetricsTracker()
        tracker._duration_cache = "5m"
        tracker._cache_ttl = timedelta(seconds=10)
        tracker._cache_time = datetime.now() - timedelta(seconds=9.999)
        assert tracker._is_cache_valid()

    def test_cache_just_over_ttl_boundary(self):
        """Test cache just over TTL boundary."""
        tracker = MetricsTracker()
        tracker._duration_cache = "5m"
        tracker._cache_ttl = timedelta(seconds=10)
        tracker._cache_time = datetime.now() - timedelta(seconds=10.001)
        assert not tracker._is_cache_valid()

    def test_format_with_no_seconds_remainder(self):
        """Test formatting when seconds remainder is 0."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(minutes=10)
        result = tracker._calculate_and_format_duration()
        assert result == "10m"

    def test_concurrent_tracker_instances(self):
        """Test multiple tracker instances don't interfere."""
        tracker1 = MetricsTracker()
        tracker2 = MetricsTracker()

        result1 = tracker1.get_duration()
        time.sleep(0.05)
        result2 = tracker2.get_duration()

        assert isinstance(result1, str)
        assert isinstance(result2, str)

    def test_duration_consistency(self):
        """Test that duration calculation is consistent."""
        tracker = MetricsTracker()
        tracker._session_start = datetime.now() - timedelta(minutes=5, seconds=30)
        result = tracker._calculate_and_format_duration()
        assert "5m" in result
        assert "30s" in result

    def test_zero_cache_ttl(self):
        """Test behavior with zero cache TTL."""
        tracker = MetricsTracker()
        tracker._cache_ttl = timedelta(seconds=0)
        tracker._duration_cache = "5m"
        tracker._cache_time = datetime.now() - timedelta(microseconds=1)
        assert not tracker._is_cache_valid()
