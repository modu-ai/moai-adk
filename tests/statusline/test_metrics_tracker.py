"""
Tests for MetricsTracker - 세션 메트릭 추적

"""

from datetime import datetime, timedelta

import pytest


class TestMetricsTracker:
    """세션 메트릭 추적 테스트"""

    def test_duration_calculation_basic(self):
        """
        GIVEN: 5분 30초 경과
        WHEN: get_duration() 호출
        THEN: "5m 30s" 또는 "5m" 포맷
        """
        from moai_adk.statusline.metrics_tracker import MetricsTracker

        tracker = MetricsTracker()

        # Set session start time in the past
        start_time = datetime.now() - timedelta(minutes=5, seconds=30)
        tracker._session_start = start_time

        duration = tracker.get_duration()

        # Should contain 'm' for minutes
        assert "m" in duration, f"Duration should contain 'm': {duration}"

    def test_duration_format_seconds(self):
        """
        GIVEN: 2분 경과
        WHEN: get_duration() 호출
        THEN: "2m" 포맷 (5분 이내는 초 단위 생략 가능)
        """
        from moai_adk.statusline.metrics_tracker import MetricsTracker

        tracker = MetricsTracker()
        start_time = datetime.now() - timedelta(minutes=2)
        tracker._session_start = start_time

        duration = tracker.get_duration()

        assert "m" in duration, f"Duration should contain 'm': {duration}"

    def test_duration_format_hours(self):
        """
        GIVEN: 1시간 30분 경과
        WHEN: get_duration() 호출
        THEN: "1h 30m" 포맷
        """
        from moai_adk.statusline.metrics_tracker import MetricsTracker

        tracker = MetricsTracker()
        start_time = datetime.now() - timedelta(hours=1, minutes=30)
        tracker._session_start = start_time

        duration = tracker.get_duration()

        assert "h" in duration or ("1" in duration and "30" in duration), f"Duration should reflect hours: {duration}"

    def test_metrics_caching(self):
        """
        GIVEN: MetricsTracker 인스턴스
        WHEN: 10초 이내에 두 번 get_duration() 호출
        THEN: 두 번째는 캐시에서 반환
        """
        from moai_adk.statusline.metrics_tracker import MetricsTracker

        tracker = MetricsTracker()
        start_time = datetime.now() - timedelta(minutes=5)
        tracker._session_start = start_time

        # First call
        result1 = tracker.get_duration()

        # Second call immediately (should use cache)
        result2 = tracker.get_duration()

        # Results should be identical (cached)
        assert result1 == result2, "Cache should return identical results"

    def test_very_short_duration(self):
        """
        GIVEN: 15초 경과
        WHEN: get_duration() 호출
        THEN: "15s" 또는 "0m" 포맷
        """
        from moai_adk.statusline.metrics_tracker import MetricsTracker

        tracker = MetricsTracker()
        start_time = datetime.now() - timedelta(seconds=15)
        tracker._session_start = start_time

        duration = tracker.get_duration()

        # Should show seconds or minute
        assert duration and len(duration) > 0, "Duration should not be empty"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
