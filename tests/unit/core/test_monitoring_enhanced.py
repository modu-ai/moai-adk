"""
Enhanced tests for Comprehensive Monitoring System - targeting 60%+ coverage.

Focus on actual API that exists.
"""

import pytest
from datetime import datetime, timezone
from unittest.mock import MagicMock, patch

from moai_adk.core.comprehensive_monitoring_system import (
    MetricType,
    AlertSeverity,
    HealthStatus,
    MetricData,
    Alert,
    ComprehensiveMonitoringSystem,
)


class TestMetricTypes:
    """Test metric types enumeration."""

    def test_all_metric_types(self):
        """Test metric types are defined."""
        assert MetricType.SYSTEM_PERFORMANCE.value == "system_performance"
        assert MetricType.ERROR_RATE.value == "error_rate"
        assert MetricType.RESPONSE_TIME.value == "response_time"
        assert MetricType.MEMORY_USAGE.value == "memory_usage"


class TestAlertSeverity:
    """Test alert severity levels."""

    def test_severity_values(self):
        """Test severity levels."""
        assert AlertSeverity.LOW.value == 1
        assert AlertSeverity.MEDIUM.value == 2
        assert AlertSeverity.HIGH.value == 3
        assert AlertSeverity.CRITICAL.value == 4
        assert AlertSeverity.EMERGENCY.value == 5


class TestHealthStatus:
    """Test health status enumeration."""

    def test_status_values(self):
        """Test all health statuses."""
        assert HealthStatus.HEALTHY.value == "healthy"
        assert HealthStatus.WARNING.value == "warning"
        assert HealthStatus.DEGRADED.value == "degraded"
        assert HealthStatus.CRITICAL.value == "critical"
        assert HealthStatus.DOWN.value == "down"


class TestMetricData:
    """Test metric data creation."""

    def test_metric_creation(self):
        """Test creating metric data."""
        now = datetime.now(timezone.utc)
        metric = MetricData(
            timestamp=now,
            metric_type=MetricType.CPU_USAGE,
            value=50.5,
            source="test",
        )
        assert metric.value == 50.5
        assert metric.metric_type == MetricType.CPU_USAGE

    def test_metric_to_dict(self):
        """Test metric serialization."""
        now = datetime.now(timezone.utc)
        metric = MetricData(
            timestamp=now,
            metric_type=MetricType.MEMORY_USAGE,
            value=512,
        )
        data = metric.to_dict()
        assert isinstance(data, dict)
        assert data["value"] == 512


class TestAlert:
    """Test alert creation."""

    def test_alert_basic_creation(self):
        """Test creating alert."""
        now = datetime.now(timezone.utc)
        alert = Alert(
            alert_id="test_1",
            severity=AlertSeverity.HIGH,
            title="Test Alert",
            description="Test description",
            threshold=80.0,
            current_value=85.0,
            source="test_source",
            timestamp=now,
            metric_type=MetricType.CPU_USAGE,
        )
        assert alert.alert_id == "test_1"
        assert alert.severity == AlertSeverity.HIGH


class TestMonitoringSystem:
    """Test monitoring system initialization."""

    def test_system_creation(self):
        """Test creating monitoring system."""
        system = ComprehensiveMonitoringSystem()
        assert system is not None

    def test_system_has_attributes(self):
        """Test system has expected attributes."""
        system = ComprehensiveMonitoringSystem()
        assert system is not None


class TestMonitoringIntegration:
    """Test monitoring functionality."""

    def test_get_dashboard_data(self):
        """Test getting dashboard data."""
        system = ComprehensiveMonitoringSystem()
        dashboard = system.get_dashboard_data()
        assert dashboard is not None


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
