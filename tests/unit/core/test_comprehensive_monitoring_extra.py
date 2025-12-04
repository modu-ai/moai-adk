"""Comprehensive test coverage for Comprehensive Monitoring System.

This module provides extensive unit tests for metrics collection, analytics,
alerting, and predictive analysis functionality.
"""

import asyncio
import pytest
import json
from datetime import datetime, timedelta
from pathlib import Path
from unittest.mock import MagicMock, AsyncMock, patch, Mock
from collections import defaultdict, deque

from moai_adk.core.comprehensive_monitoring_system import (
    MetricType,
    AlertSeverity,
    HealthStatus,
    MetricData,
    Alert,
)


class TestComprehensiveMetricType:
    """Test MetricType enum in comprehensive monitoring system"""

    def test_metric_type_enum_values(self):
        """Test MetricType has all expected values"""
        assert MetricType.SYSTEM_PERFORMANCE is not None
        assert MetricType.USER_BEHAVIOR is not None
        assert MetricType.TOKEN_USAGE is not None
        assert MetricType.ERROR_RATE is not None
        assert MetricType.RESPONSE_TIME is not None
        assert MetricType.MEMORY_USAGE is not None
        assert MetricType.CPU_USAGE is not None
        assert MetricType.THROUGHPUT is not None
        assert MetricType.AVAILABILITY is not None

    def test_metric_type_comprehensive_coverage(self):
        """Test MetricType covers comprehensive monitoring dimensions"""
        metric_types = list(MetricType)

        # Should have metrics for different dimensions
        values = [m.value for m in metric_types]
        assert len(values) == len(set(values))  # All unique


class TestComprehensiveAlertSeverity:
    """Test AlertSeverity enum in comprehensive monitoring"""

    def test_alert_severity_enum_values(self):
        """Test AlertSeverity has all expected values"""
        assert AlertSeverity.LOW is not None
        assert AlertSeverity.MEDIUM is not None
        assert AlertSeverity.HIGH is not None
        assert AlertSeverity.CRITICAL is not None
        assert AlertSeverity.EMERGENCY is not None

    def test_alert_severity_escalation_order(self):
        """Test AlertSeverity follows escalation order"""
        severities = list(AlertSeverity)
        for i in range(len(severities) - 1):
            assert severities[i].value < severities[i + 1].value


class TestHealthStatus:
    """Test HealthStatus enum"""

    def test_health_status_enum_values(self):
        """Test HealthStatus has all expected values"""
        assert HealthStatus.HEALTHY is not None
        assert HealthStatus.WARNING is not None
        assert HealthStatus.DEGRADED is not None
        assert HealthStatus.CRITICAL is not None
        assert HealthStatus.DOWN is not None

    def test_health_status_values(self):
        """Test HealthStatus string values"""
        assert HealthStatus.HEALTHY.value == "healthy"
        assert HealthStatus.WARNING.value == "warning"
        assert HealthStatus.DEGRADED.value == "degraded"
        assert HealthStatus.CRITICAL.value == "critical"
        assert HealthStatus.DOWN.value == "down"


class TestComprehensiveMetricData:
    """Test MetricData dataclass in comprehensive monitoring"""

    def test_metric_data_initialization(self):
        """Test MetricData initializes correctly"""
        timestamp = datetime.now()
        metric = MetricData(
            timestamp=timestamp,
            metric_type=MetricType.CPU_USAGE,
            value=45.5,
            tags={"host": "server1"},
            source="prometheus"
        )
        assert metric.timestamp == timestamp
        assert metric.metric_type == MetricType.CPU_USAGE
        assert metric.value == 45.5

    def test_metric_data_all_fields(self):
        """Test MetricData with all fields"""
        timestamp = datetime.now()
        metric = MetricData(
            timestamp=timestamp,
            metric_type=MetricType.USER_BEHAVIOR,
            value="click",
            tags={"user_id": "user123", "action": "button_click"},
            source="frontend",
            metadata={"page": "/dashboard", "duration_ms": 150}
        )
        assert metric.tags["user_id"] == "user123"
        assert metric.metadata["page"] == "/dashboard"

    def test_metric_data_to_dict(self):
        """Test MetricData to_dict conversion"""
        timestamp = datetime.now()
        metric = MetricData(
            timestamp=timestamp,
            metric_type=MetricType.MEMORY_USAGE,
            value=1024,
            tags={"component": "api"},
            source="system_monitor"
        )

        metric_dict = metric.to_dict()
        assert metric_dict["metric_type"] == MetricType.MEMORY_USAGE.value
        assert metric_dict["value"] == 1024
        assert metric_dict["source"] == "system_monitor"

    def test_metric_data_various_value_types(self):
        """Test MetricData with different value types"""
        timestamp = datetime.now()

        # Test with integer
        m1 = MetricData(timestamp=timestamp, metric_type=MetricType.ERROR_RATE, value=5)
        assert isinstance(m1.value, int)

        # Test with float
        m2 = MetricData(timestamp=timestamp, metric_type=MetricType.RESPONSE_TIME, value=123.45)
        assert isinstance(m2.value, float)

        # Test with string
        m3 = MetricData(timestamp=timestamp, metric_type=MetricType.SYSTEM_PERFORMANCE, value="optimal")
        assert isinstance(m3.value, str)

        # Test with boolean
        m4 = MetricData(timestamp=timestamp, metric_type=MetricType.AVAILABILITY, value=True)
        assert isinstance(m4.value, bool)

    def test_metric_data_default_tags(self):
        """Test MetricData default tags"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.THROUGHPUT,
            value=1000
        )
        assert metric.tags == {}

    def test_metric_data_default_metadata(self):
        """Test MetricData default metadata"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.TOKEN_USAGE,
            value=5000
        )
        assert metric.metadata == {}

    def test_metric_data_default_source(self):
        """Test MetricData default source"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=50.0
        )
        assert metric.source == ""

    def test_metric_data_with_complex_metadata(self):
        """Test MetricData with complex nested metadata"""
        complex_metadata = {
            "request": {
                "headers": {"content-type": "application/json"},
                "body": {"user_id": 123}
            },
            "response": {
                "status": 200,
                "duration_ms": 45
            }
        }
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.RESPONSE_TIME,
            value=45,
            metadata=complex_metadata
        )
        assert metric.metadata["request"]["headers"]["content-type"] == "application/json"


class TestAlert:
    """Test Alert dataclass"""

    def test_alert_initialization(self):
        """Test Alert initializes correctly"""
        timestamp = datetime.now()
        alert = Alert(
            alert_id="alert_001",
            severity=AlertSeverity.HIGH,
            title="High CPU Usage",
            description="CPU usage exceeded 90%",
            timestamp=timestamp,
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=95.0,
            source="system_monitor"
        )
        assert alert.alert_id == "alert_001"
        assert alert.severity == AlertSeverity.HIGH
        assert alert.title == "High CPU Usage"

    def test_alert_with_all_fields(self):
        """Test Alert with all fields"""
        timestamp = datetime.now()
        alert = Alert(
            alert_id="alert_002",
            severity=AlertSeverity.CRITICAL,
            title="Database Connection Failed",
            description="Unable to connect to primary database",
            timestamp=timestamp,
            metric_type=MetricType.SYSTEM_PERFORMANCE,
            threshold=100.0,
            current_value=150.0,
            source="database_monitor",
            tags={"component": "database_pool"},
            resolved=False
        )
        assert alert.current_value == 150.0
        assert alert.tags["component"] == "database_pool"

    def test_alert_severity_levels(self):
        """Test Alert with different severity levels"""
        timestamp = datetime.now()
        for severity in AlertSeverity:
            alert = Alert(
                alert_id="test",
                severity=severity,
                title="Test Alert",
                description="Test",
                timestamp=timestamp,
                metric_type=MetricType.ERROR_RATE,
                threshold=10.0,
                current_value=15.0,
                source="test"
            )
            assert alert.severity == severity

    def test_alert_metric_type_assignment(self):
        """Test Alert can be associated with MetricType"""
        timestamp = datetime.now()
        alert = Alert(
            alert_id="metric_alert",
            severity=AlertSeverity.MEDIUM,
            title="Memory Alert",
            description="Memory usage high",
            timestamp=timestamp,
            metric_type=MetricType.MEMORY_USAGE,
            threshold=1024.0,
            current_value=2048.0,
            source="system"
        )
        assert alert.metric_type == MetricType.MEMORY_USAGE

    def test_alert_threshold_and_current_value(self):
        """Test Alert tracks threshold and current values"""
        timestamp = datetime.now()
        alert = Alert(
            alert_id="threshold_alert",
            severity=AlertSeverity.HIGH,
            title="Threshold Exceeded",
            description="Value exceeded threshold",
            timestamp=timestamp,
            metric_type=MetricType.THROUGHPUT,
            threshold=80.0,
            current_value=85.5,
            source="metrics"
        )
        assert alert.threshold == 80.0
        assert alert.current_value == 85.5
        assert alert.current_value > alert.threshold

    def test_alert_resolution_tracking(self):
        """Test Alert tracks resolution status"""
        timestamp = datetime.now()
        alert = Alert(
            alert_id="resolv_alert",
            severity=AlertSeverity.HIGH,
            title="Temporary Issue",
            description="Issue resolved",
            timestamp=timestamp,
            metric_type=MetricType.ERROR_RATE,
            threshold=5.0,
            current_value=2.0,
            source="app",
            resolved=True,
            resolved_at=datetime.now()
        )
        assert alert.resolved is True
        assert alert.resolved_at is not None


class TestMetricDataEdgeCases:
    """Test edge cases for MetricData"""

    def test_metric_data_empty_tags_dict(self):
        """Test MetricData with explicitly empty tags"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
            tags={}
        )
        assert metric.tags == {}

    def test_metric_data_large_tags(self):
        """Test MetricData with many tags"""
        large_tags = {f"tag_{i}": f"value_{i}" for i in range(50)}
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.THROUGHPUT,
            value=1000,
            tags=large_tags
        )
        assert len(metric.tags) == 50

    def test_metric_data_special_characters_in_values(self):
        """Test MetricData with special characters in tags"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.ERROR_RATE,
            value=0.05,
            tags={
                "path": "/api/v1/users/@me",
                "query": "filter=status:active&sort=-created_at"
            }
        )
        assert "@" in metric.tags["path"]
        assert "&" in metric.tags["query"]

    def test_metric_data_very_large_number(self):
        """Test MetricData with very large numbers"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.THROUGHPUT,
            value=999999999999999
        )
        assert metric.value == 999999999999999

    def test_metric_data_very_small_number(self):
        """Test MetricData with very small numbers"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.ERROR_RATE,
            value=0.000001
        )
        assert abs(metric.value - 0.000001) < 1e-10

    def test_metric_data_negative_timestamp(self):
        """Test MetricData with past timestamps"""
        past_time = datetime.now() - timedelta(days=365)
        metric = MetricData(
            timestamp=past_time,
            metric_type=MetricType.SYSTEM_PERFORMANCE,
            value=100.0
        )
        assert metric.timestamp < datetime.now()

    def test_metric_data_future_timestamp(self):
        """Test MetricData with future timestamps"""
        future_time = datetime.now() + timedelta(days=365)
        metric = MetricData(
            timestamp=future_time,
            metric_type=MetricType.USER_BEHAVIOR,
            value="prediction"
        )
        assert metric.timestamp > datetime.now()

    def test_metric_data_unicode_in_source(self):
        """Test MetricData with unicode in source"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.TOKEN_USAGE,
            value=1000,
            source="监控系统"
        )
        assert metric.source == "监控系统"

    def test_metric_data_empty_source(self):
        """Test MetricData with empty source"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.RESPONSE_TIME,
            value=100.0,
            source=""
        )
        assert metric.source == ""

    def test_metric_data_to_dict_with_all_fields(self):
        """Test to_dict preserves all fields"""
        timestamp = datetime.now()
        metric = MetricData(
            timestamp=timestamp,
            metric_type=MetricType.CPU_USAGE,
            value=55.5,
            tags={"host": "server1"},
            source="metrics_collector",
            metadata={"region": "us-west-2"}
        )

        metric_dict = metric.to_dict()
        assert "timestamp" in metric_dict
        assert "metric_type" in metric_dict
        assert "value" in metric_dict
        assert "tags" in metric_dict
        assert "source" in metric_dict
        assert "metadata" in metric_dict


class TestAlertEdgeCases:
    """Test edge cases for Alert"""

    def test_alert_minimum_fields(self):
        """Test Alert with only required fields"""
        timestamp = datetime.now()
        alert = Alert(
            alert_id="min_alert",
            severity=AlertSeverity.LOW,
            title="Minimal Alert",
            description="Minimal description",
            timestamp=timestamp,
            metric_type=MetricType.RESPONSE_TIME,
            threshold=100.0,
            current_value=95.0,
            source="test"
        )
        assert alert.alert_id == "min_alert"
        assert alert.metric_type == MetricType.RESPONSE_TIME

    def test_alert_with_zero_threshold(self):
        """Test Alert with zero threshold"""
        timestamp = datetime.now()
        alert = Alert(
            alert_id="zero_alert",
            severity=AlertSeverity.MEDIUM,
            title="Zero Threshold",
            description="Alert with zero threshold",
            timestamp=timestamp,
            metric_type=MetricType.ERROR_RATE,
            threshold=0.0,
            current_value=0.1,
            source="app"
        )
        assert alert.threshold == 0.0

    def test_alert_with_negative_threshold(self):
        """Test Alert with negative threshold (edge case)"""
        timestamp = datetime.now()
        alert = Alert(
            alert_id="negative_alert",
            severity=AlertSeverity.HIGH,
            title="Negative Threshold",
            description="Alert with negative threshold",
            timestamp=timestamp,
            metric_type=MetricType.THROUGHPUT,
            threshold=-10.0,
            current_value=5.0,
            source="system"
        )
        assert alert.threshold == -10.0

    def test_alert_long_title_and_description(self):
        """Test Alert with long title and description"""
        long_title = "A" * 500
        long_description = "B" * 2000
        timestamp = datetime.now()

        alert = Alert(
            alert_id="long_alert",
            severity=AlertSeverity.CRITICAL,
            title=long_title,
            description=long_description,
            timestamp=timestamp,
            metric_type=MetricType.CPU_USAGE,
            threshold=50.0,
            current_value=95.0,
            source="monitor"
        )
        assert len(alert.title) == 500
        assert len(alert.description) == 2000

    def test_alert_special_characters_in_description(self):
        """Test Alert with special characters"""
        special_desc = "Alert: Status=CRITICAL & (CPU>90% OR Memory>80%) !urgent"
        timestamp = datetime.now()
        alert = Alert(
            alert_id="special_alert",
            severity=AlertSeverity.CRITICAL,
            title="Special Characters",
            description=special_desc,
            timestamp=timestamp,
            metric_type=MetricType.SYSTEM_PERFORMANCE,
            threshold=80.0,
            current_value=92.0,
            source="health"
        )
        assert alert.description == special_desc

    def test_alert_unicode_in_title_and_description(self):
        """Test Alert with unicode characters"""
        timestamp = datetime.now()
        alert = Alert(
            alert_id="unicode_alert",
            severity=AlertSeverity.HIGH,
            title="告警 Alert",  # Alert in Chinese
            description="系统故障 System Failure",
            timestamp=timestamp,
            metric_type=MetricType.AVAILABILITY,
            threshold=99.5,
            current_value=98.0,
            source="monitor"
        )
        assert "告警" in alert.title
        assert "系统故障" in alert.description


class TestMetricAndAlertIntegration:
    """Test integration between MetricData and Alert"""

    def test_alert_from_metric_threshold_violation(self):
        """Test creating alert from metric threshold violation"""
        timestamp = datetime.now()
        metric = MetricData(
            timestamp=timestamp,
            metric_type=MetricType.MEMORY_USAGE,
            value=2000.0,
            tags={"host": "server1"}
        )

        # Create alert based on metric
        alert = Alert(
            alert_id=f"alert_{timestamp.isoformat()}",
            severity=AlertSeverity.HIGH,
            title=f"{metric.metric_type.value} Threshold Exceeded",
            description=f"Memory usage {metric.value} MB exceeds limit",
            timestamp=timestamp,
            metric_type=metric.metric_type,
            current_value=metric.value,
            threshold=1500.0,
            source="metric_system"
        )

        assert alert.metric_type == metric.metric_type
        assert alert.current_value == metric.value

    def test_multiple_metrics_trigger_single_alert(self):
        """Test multiple related metrics triggering alert"""
        timestamp = datetime.now()
        metrics = [
            MetricData(timestamp, MetricType.CPU_USAGE, 85.0),
            MetricData(timestamp, MetricType.MEMORY_USAGE, 2048.0),
            MetricData(timestamp, MetricType.RESPONSE_TIME, 500.0)
        ]

        # All metrics indicate system stress
        alert = Alert(
            alert_id="system_stress_alert",
            severity=AlertSeverity.CRITICAL,
            title="System Under High Load",
            description="Multiple performance metrics indicate system stress",
            timestamp=timestamp,
            metric_type=MetricType.SYSTEM_PERFORMANCE,
            threshold=50.0,
            current_value=85.0,
            source="system_monitor",
            tags={"components": ["cpu", "memory", "response_time"]}
        )

        assert alert.severity == AlertSeverity.CRITICAL
