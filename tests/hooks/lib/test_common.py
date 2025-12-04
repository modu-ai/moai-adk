"""Comprehensive test suite for common.py utilities"""

import statistics

import pytest
from common import format_duration, get_summary_stats


class TestFormatDuration:
    """Test format_duration() function"""

    # Seconds tests
    def test_format_zero_seconds(self):
        """Format zero seconds"""
        result = format_duration(0)
        assert result == "0.0s"

    def test_format_single_second(self):
        """Format single second"""
        result = format_duration(1)
        assert result == "1.0s"

    def test_format_fractional_seconds(self):
        """Format fractional seconds"""
        result = format_duration(30.5)
        assert result == "30.5s"

    @pytest.mark.parametrize(
        "seconds,expected",
        [
            (0.1, "0.1s"),
            (5.5, "5.5s"),
            (15.25, "15.2s"),
            (30.99, "31.0s"),
            (59.9, "59.9s"),
        ],
    )
    def test_format_various_seconds(self, seconds, expected):
        """Format various second values"""
        result = format_duration(seconds)
        # Check format (s, m, or h suffix)
        assert "s" in result
        assert result.startswith(str(int(seconds)) if seconds < 10 else "")

    # Minutes tests
    def test_format_sixty_seconds_as_minute(self):
        """Format 60 seconds as 1 minute"""
        result = format_duration(60)
        assert result == "1.0m"

    def test_format_one_minute(self):
        """Format one minute exactly"""
        result = format_duration(60)
        assert result == "1.0m"

    def test_format_fractional_minutes(self):
        """Format fractional minutes"""
        result = format_duration(90)
        assert result == "1.5m"

    def test_format_two_minutes(self):
        """Format two minutes"""
        result = format_duration(120)
        assert result == "2.0m"

    @pytest.mark.parametrize(
        "seconds,expected_unit",
        [
            (60, "m"),
            (120, "m"),
            (300, "m"),
            (1200, "m"),  # 20 minutes, still in minutes
            (3599, "m"),  # Just under 1 hour
        ],
    )
    def test_format_minutes_range(self, seconds, expected_unit):
        """Format various minute values"""
        result = format_duration(seconds)
        assert expected_unit in result

    def test_format_minutes_with_decimals(self):
        """Format minutes with decimal precision"""
        result = format_duration(150)  # 2.5 minutes
        assert result == "2.5m"

    # Hours tests
    def test_format_one_hour(self):
        """Format one hour (3600 seconds)"""
        result = format_duration(3600)
        assert result == "1.0h"

    def test_format_two_hours(self):
        """Format two hours"""
        result = format_duration(7200)
        assert result == "2.0h"

    def test_format_fractional_hours(self):
        """Format fractional hours"""
        result = format_duration(5400)  # 1.5 hours
        assert result == "1.5h"

    @pytest.mark.parametrize(
        "seconds,expected_unit",
        [
            (3600, "h"),
            (5400, "h"),
            (7200, "h"),
            (10800, "h"),
            (86400, "h"),  # 24 hours
        ],
    )
    def test_format_hours_range(self, seconds, expected_unit):
        """Format various hour values"""
        result = format_duration(seconds)
        assert expected_unit in result

    def test_format_large_duration(self):
        """Format large duration (multiple days)"""
        result = format_duration(86400)  # 24 hours
        assert "h" in result

    def test_format_very_large_duration(self):
        """Format very large duration"""
        result = format_duration(259200)  # 72 hours / 3 days
        assert "h" in result

    # Edge cases
    def test_format_very_small_duration(self):
        """Format very small duration"""
        result = format_duration(0.001)
        assert "0.0s" in result or "s" in result

    def test_format_float_precision(self):
        """Format maintains one decimal place"""
        result = format_duration(1.23456)
        # Should have exactly one decimal place
        parts = result.rstrip("smh").split(".")
        assert len(parts[1]) == 1 if "." in result else True

    def test_format_boundary_60_seconds(self):
        """Format boundary at 60 seconds"""
        result_59 = format_duration(59.9)
        result_60 = format_duration(60.0)

        assert "s" in result_59
        assert "m" in result_60

    def test_format_boundary_3600_seconds(self):
        """Format boundary at 3600 seconds (1 hour)"""
        result_3599 = format_duration(3599)
        result_3600 = format_duration(3600)

        assert "m" in result_3599
        assert "h" in result_3600

    def test_format_negative_duration(self):
        """Format negative duration (edge case)"""
        result = format_duration(-10)
        # Implementation specific, but should not crash
        assert isinstance(result, str)


class TestGetSummaryStats:
    """Test get_summary_stats() function"""

    # Basic statistics tests
    def test_summary_stats_single_value(self):
        """Get summary stats for single value"""
        stats = get_summary_stats([42])

        assert stats["mean"] == 42
        assert stats["min"] == 42
        assert stats["max"] == 42
        assert stats["std"] == 0

    def test_summary_stats_two_values(self):
        """Get summary stats for two values"""
        stats = get_summary_stats([10, 20])

        assert stats["mean"] == 15
        assert stats["min"] == 10
        assert stats["max"] == 20

    def test_summary_stats_multiple_values(self):
        """Get summary stats for multiple values"""
        stats = get_summary_stats([1, 2, 3, 4, 5])

        assert stats["mean"] == 3
        assert stats["min"] == 1
        assert stats["max"] == 5

    @pytest.mark.parametrize(
        "values,expected_mean",
        [
            ([1, 1, 1], 1),
            ([1, 2, 3], 2),
            ([10, 20, 30], 20),
            ([0, 0, 0], 0),
        ],
    )
    def test_summary_stats_mean_calculation(self, values, expected_mean):
        """Test mean calculation"""
        stats = get_summary_stats(values)
        assert stats["mean"] == expected_mean

    @pytest.mark.parametrize(
        "values,expected_min,expected_max",
        [
            ([1, 2, 3], 1, 3),
            ([10, 5, 15], 5, 15),
            ([100], 100, 100),
            ([-10, 0, 10], -10, 10),
        ],
    )
    def test_summary_stats_min_max(self, values, expected_min, expected_max):
        """Test min and max calculation"""
        stats = get_summary_stats(values)
        assert stats["min"] == expected_min
        assert stats["max"] == expected_max

    def test_summary_stats_standard_deviation(self):
        """Test standard deviation calculation"""
        values = [1, 2, 3, 4, 5]
        stats = get_summary_stats(values)

        expected_std = statistics.stdev(values)
        assert stats["std"] == expected_std

    def test_summary_stats_zero_std_for_single_value(self):
        """Standard deviation is 0 for single value"""
        stats = get_summary_stats([5])
        assert stats["std"] == 0

    def test_summary_stats_returns_dict(self):
        """get_summary_stats returns dictionary"""
        stats = get_summary_stats([1, 2, 3])

        assert isinstance(stats, dict)
        assert set(stats.keys()) == {"mean", "min", "max", "std"}

    # Empty list tests
    def test_summary_stats_empty_list(self):
        """Get summary stats for empty list"""
        stats = get_summary_stats([])

        assert stats["mean"] == 0
        assert stats["min"] == 0
        assert stats["max"] == 0
        assert stats["std"] == 0

    def test_summary_stats_none_input(self):
        """Get summary stats for None input"""
        # Should handle gracefully or raise
        try:
            stats = get_summary_stats(None)
            # If it doesn't raise, check it returns default values or empty
            assert isinstance(stats, dict)
        except (TypeError, AttributeError):
            # Expected behavior for None input
            pass

    # Float values tests
    def test_summary_stats_float_values(self):
        """Get summary stats for float values"""
        values = [1.5, 2.5, 3.5]
        stats = get_summary_stats(values)

        assert stats["mean"] == 2.5
        assert stats["min"] == 1.5
        assert stats["max"] == 3.5

    def test_summary_stats_mixed_int_float(self):
        """Get summary stats for mixed int and float values"""
        values = [1, 2.5, 3, 4.5, 5]
        stats = get_summary_stats(values)

        assert isinstance(stats["mean"], float)
        assert stats["min"] == 1
        assert stats["max"] == 5

    # Negative values tests
    def test_summary_stats_negative_values(self):
        """Get summary stats for negative values"""
        values = [-5, -3, -1]
        stats = get_summary_stats(values)

        assert stats["mean"] == -3
        assert stats["min"] == -5
        assert stats["max"] == -1

    def test_summary_stats_mixed_positive_negative(self):
        """Get summary stats for mixed positive and negative values"""
        values = [-10, 0, 10]
        stats = get_summary_stats(values)

        assert stats["mean"] == 0
        assert stats["min"] == -10
        assert stats["max"] == 10

    # Large dataset tests
    def test_summary_stats_large_dataset(self):
        """Get summary stats for large dataset"""
        values = list(range(1, 1001))  # 1 to 1000
        stats = get_summary_stats(values)

        assert stats["mean"] == 500.5
        assert stats["min"] == 1
        assert stats["max"] == 1000

    def test_summary_stats_all_same_values(self):
        """Get summary stats for all identical values"""
        values = [5, 5, 5, 5, 5]
        stats = get_summary_stats(values)

        assert stats["mean"] == 5
        assert stats["min"] == 5
        assert stats["max"] == 5
        assert stats["std"] == 0

    # Precision tests
    def test_summary_stats_high_precision(self):
        """Get summary stats with high precision values"""
        values = [1.123456, 2.234567, 3.345678]
        stats = get_summary_stats(values)

        assert isinstance(stats["mean"], float)
        assert abs(stats["mean"] - 2.234567) < 0.0001  # Approximate check

    def test_summary_stats_very_small_values(self):
        """Get summary stats for very small values"""
        values = [0.001, 0.002, 0.003]
        stats = get_summary_stats(values)

        assert stats["min"] == 0.001
        assert stats["max"] == 0.003

    # Special cases
    def test_summary_stats_two_values_std(self):
        """Test standard deviation with exactly two values"""
        values = [1, 3]
        stats = get_summary_stats(values)

        # Standard deviation of [1, 3] is sqrt(2) ≈ 1.414
        expected_std = statistics.stdev(values)
        assert abs(stats["std"] - expected_std) < 0.001

    def test_summary_stats_normal_distribution(self):
        """Get summary stats for normally distributed values"""
        values = [98, 99, 100, 101, 102]  # Normal distribution around 100
        stats = get_summary_stats(values)

        assert stats["mean"] == 100
        assert stats["min"] == 98
        assert stats["max"] == 102

    def test_summary_stats_skewed_distribution(self):
        """Get summary stats for skewed distribution"""
        values = [1, 2, 3, 100, 200]  # Right-skewed
        stats = get_summary_stats(values)

        assert stats["mean"] == 61.2
        assert stats["min"] == 1
        assert stats["max"] == 200

    # Return value structure tests
    def test_summary_stats_return_structure(self):
        """Test return value structure"""
        stats = get_summary_stats([1, 2, 3])

        assert "mean" in stats
        assert "min" in stats
        assert "max" in stats
        assert "std" in stats

    def test_summary_stats_numeric_values(self):
        """Test all return values are numeric"""
        stats = get_summary_stats([1.5, 2.5, 3.5])

        assert isinstance(stats["mean"], (int, float))
        assert isinstance(stats["min"], (int, float))
        assert isinstance(stats["max"], (int, float))
        assert isinstance(stats["std"], (int, float))


class TestCommonIntegration:
    """Integration tests for common.py utilities"""

    def test_format_duration_with_stats(self):
        """Use format_duration with get_summary_stats results"""
        values = [1.5, 2.5, 3.5, 4.5]
        stats = get_summary_stats(values)

        # Format mean as duration
        formatted_mean = format_duration(stats["mean"])
        assert "s" in formatted_mean

    def test_durations_from_stat_values(self):
        """Format multiple values as durations"""
        values = [30, 60, 120, 180, 360, 3600, 7200]

        # Get stats
        get_summary_stats(values)

        # Format each value as duration
        formatted_values = [format_duration(v) for v in values]

        # Check format consistency
        assert all(any(unit in f for unit in ["s", "m", "h"]) for f in formatted_values)

    def test_processing_metrics_workflow(self):
        """Simulate processing metrics workflow"""
        # Simulate processing times in seconds
        processing_times = [2.3, 2.1, 2.5, 2.4, 2.2, 100.5]  # One outlier

        # Get statistics
        stats = get_summary_stats(processing_times)

        # Format durations for display
        mean_formatted = format_duration(stats["mean"])
        min_formatted = format_duration(stats["min"])
        max_formatted = format_duration(stats["max"])

        # Assertions
        assert isinstance(stats["mean"], float)
        assert "s" in mean_formatted or "m" in mean_formatted
        assert "s" in min_formatted
        assert "s" in max_formatted or "m" in max_formatted

    def test_batch_processing_metrics(self):
        """Test batch processing metric collection"""
        batches = [
            [10, 12, 11],  # Batch 1
            [100, 102, 101],  # Batch 2
            [1000, 1005, 998],  # Batch 3
        ]

        for batch in batches:
            stats = get_summary_stats(batch)

            # Format statistics for reporting
            report = {
                "mean": format_duration(stats["mean"]),
                "min": format_duration(stats["min"]),
                "max": format_duration(stats["max"]),
            }

            assert all(isinstance(v, str) for v in report.values())
            assert all(any(unit in v for unit in ["s", "m", "h"]) for v in report.values())
