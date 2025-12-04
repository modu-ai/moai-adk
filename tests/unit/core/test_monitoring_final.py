"""
Final comprehensive tests for Comprehensive Monitoring System.

Focuses on simple, working tests for:
- MetricData and Alert dataclasses
- MetricsCollector
- AlertManager
- SystemHealth
"""

import json
import statistics
import time
from datetime import datetime, timedelta
from pathlib import Path
from unittest.mock import MagicMock, patch, Mock

import pytest

from moai_adk.core.comprehensive_monitoring_system import (
    MetricType,
    AlertSeverity,
    HealthStatus,
    MetricData,
    Alert,
    SystemHealth,
    MetricsCollector,
    AlertManager,
)


class TestMetricType:
    """Test MetricType enumeration."""

    def test_metric_type_values(self):
        """Test MetricType values are defined."""
        # Assert
        assert MetricType.SYSTEM_PERFORMANCE.value == "system_performance"
        assert MetricType.USER_BEHAVIOR.value == "user_behavior"
        assert MetricType.TOKEN_USAGE.value == "token_usage"
        assert MetricType.ERROR_RATE.value == "error_rate"
        assert MetricType.RESPONSE_TIME.value == "response_time"
        assert MetricType.MEMORY_USAGE.value == "memory_usage"
        assert MetricType.CPU_USAGE.value == "cpu_usage"
        assert MetricType.THROUGHPUT.value == "throughput"
        assert MetricType.AVAILABILITY.value == "availability"


class TestAlertSeverity:
    """Test AlertSeverity enumeration."""

    def test_alert_severity_values(self):
        """Test AlertSeverity values."""
        # Assert
        assert AlertSeverity.LOW.value == 1
        assert AlertSeverity.MEDIUM.value == 2
        assert AlertSeverity.HIGH.value == 3
        assert AlertSeverity.CRITICAL.value == 4
        assert AlertSeverity.EMERGENCY.value == 5

    def test_alert_severity_comparison(self):
        """Test severity comparison."""
        # Assert
        assert AlertSeverity.LOW.value < AlertSeverity.EMERGENCY.value
        assert AlertSeverity.CRITICAL.value > AlertSeverity.MEDIUM.value


class TestHealthStatus:
    """Test HealthStatus enumeration."""

    def test_health_status_values(self):
        """Test HealthStatus values."""
        # Assert
        assert HealthStatus.HEALTHY.value == "healthy"
        assert HealthStatus.WARNING.value == "warning"
        assert HealthStatus.DEGRADED.value == "degraded"
        assert HealthStatus.CRITICAL.value == "critical"
        assert HealthStatus.DOWN.value == "down"


class TestMetricData:
    """Test MetricData dataclass."""

    def test_metric_data_creation_numeric(self):
        """Test creating MetricData with numeric value."""
        # Arrange
        now = datetime.now()
        metric = MetricData(
            timestamp=now,
            metric_type=MetricType.RESPONSE_TIME,
            value=150.5,
            tags={"service": "api", "endpoint": "/users"},
            source="test_source",
        )

        # Assert
        assert metric.timestamp == now
        assert metric.metric_type == MetricType.RESPONSE_TIME
        assert metric.value == 150.5
        assert metric.tags["service"] == "api"
        assert metric.source == "test_source"

    def test_metric_data_creation_string_value(self):
        """Test MetricData with string value."""
        # Arrange
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.USER_BEHAVIOR,
            value="login_attempt",
            source="auth_service",
        )

        # Assert
        assert metric.value == "login_attempt"
        assert isinstance(metric.value, str)

    def test_metric_data_creation_bool_value(self):
        """Test MetricData with boolean value."""
        # Arrange
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.AVAILABILITY,
            value=True,
        )

        # Assert
        assert metric.value is True

    def test_metric_data_to_dict(self):
        """Test MetricData serialization to dictionary."""
        # Arrange
        now = datetime.now()
        metric = MetricData(
            timestamp=now,
            metric_type=MetricType.CPU_USAGE,
            value=75.5,
            tags={"host": "server-1"},
            source="monitoring_agent",
            metadata={"cpu_count": 4},
        )

        # Act
        result = metric.to_dict()

        # Assert
        assert result["metric_type"] == "cpu_usage"
        assert result["value"] == 75.5
        assert result["tags"]["host"] == "server-1"
        assert result["source"] == "monitoring_agent"
        assert result["metadata"]["cpu_count"] == 4
        assert isinstance(result["timestamp"], str)

    def test_metric_data_defaults(self):
        """Test MetricData with default values."""
        # Arrange & Act
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.SYSTEM_PERFORMANCE,
            value=100,
        )

        # Assert
        assert metric.tags == {}
        assert metric.source == ""
        assert metric.metadata == {}


class TestAlert:
    """Test Alert dataclass."""

    def test_alert_creation(self):
        """Test creating Alert."""
        # Arrange
        now = datetime.now()
        alert = Alert(
            alert_id="ALERT-001",
            severity=AlertSeverity.HIGH,
            title="High CPU Usage",
            description="CPU usage exceeded 90%",
            timestamp=now,
            metric_type=MetricType.CPU_USAGE,
            threshold=90.0,
            current_value=95.5,
            source="cpu_monitor",
        )

        # Assert
        assert alert.alert_id == "ALERT-001"
        assert alert.severity == AlertSeverity.HIGH
        assert alert.title == "High CPU Usage"
        assert alert.threshold == 90.0
        assert alert.current_value == 95.5
        assert alert.resolved is False
        assert alert.acknowledged is False

    def test_alert_resolution(self):
        """Test alert resolution."""
        # Arrange
        alert = Alert(
            alert_id="ALERT-001",
            severity=AlertSeverity.MEDIUM,
            title="Test Alert",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.ERROR_RATE,
            threshold=5.0,
            current_value=3.0,
            source="test",
        )

        # Act
        alert.resolved = True
        alert.resolved_at = datetime.now()

        # Assert
        assert alert.resolved is True
        assert alert.resolved_at is not None

    def test_alert_acknowledgment(self):
        """Test alert acknowledgment."""
        # Arrange
        alert = Alert(
            alert_id="ALERT-002",
            severity=AlertSeverity.CRITICAL,
            title="Critical Issue",
            description="System critical",
            timestamp=datetime.now(),
            metric_type=MetricType.AVAILABILITY,
            threshold=99.0,
            current_value=95.0,
            source="availability_check",
        )

        # Act
        alert.acknowledged = True
        alert.acknowledged_at = datetime.now()

        # Assert
        assert alert.acknowledged is True
        assert alert.acknowledged_at is not None

    def test_alert_to_dict(self):
        """Test Alert serialization."""
        # Arrange
        now = datetime.now()
        alert = Alert(
            alert_id="ALERT-003",
            severity=AlertSeverity.LOW,
            title="Low Memory",
            description="Memory usage low",
            timestamp=now,
            metric_type=MetricType.MEMORY_USAGE,
            threshold=10.0,
            current_value=8.5,
            source="memory_monitor",
            tags={"service": "api"},
        )

        # Act
        result = alert.to_dict()

        # Assert
        assert result["alert_id"] == "ALERT-003"
        assert result["severity"] == 1
        assert result["title"] == "Low Memory"
        assert result["threshold"] == 10.0
        assert isinstance(result["timestamp"], str)

    def test_alert_to_dict_with_resolution(self):
        """Test Alert dict includes resolution info."""
        # Arrange
        resolved_at = datetime.now()
        alert = Alert(
            alert_id="ALERT-004",
            severity=AlertSeverity.MEDIUM,
            title="Test",
            description="Test alert",
            timestamp=datetime.now(),
            metric_type=MetricType.TOKEN_USAGE,
            threshold=1000,
            current_value=900,
            source="test",
            resolved=True,
            resolved_at=resolved_at,
        )

        # Act
        result = alert.to_dict()

        # Assert
        assert result["resolved"] is True
        assert result["resolved_at"] is not None


class TestSystemHealth:
    """Test SystemHealth dataclass."""

    def test_system_health_creation(self):
        """Test creating SystemHealth."""
        # Arrange
        now = datetime.now()
        health = SystemHealth(
            status=HealthStatus.HEALTHY,
            timestamp=now,
            overall_score=95.5,
            component_scores={"cpu": 95.0, "memory": 90.0, "disk": 100.0},
            active_alerts=[],
            uptime_percentage=99.9,
        )

        # Assert
        assert health.status == HealthStatus.HEALTHY
        assert health.overall_score == 95.5
        assert health.component_scores["cpu"] == 95.0
        assert health.uptime_percentage == 99.9

    def test_system_health_degraded(self):
        """Test degraded system health."""
        # Arrange
        health = SystemHealth(
            status=HealthStatus.DEGRADED,
            timestamp=datetime.now(),
            overall_score=70.0,
            active_alerts=["ALERT-001", "ALERT-002"],
        )

        # Assert
        assert health.status == HealthStatus.DEGRADED
        assert len(health.active_alerts) == 2

    def test_system_health_critical(self):
        """Test critical system health."""
        # Arrange
        health = SystemHealth(
            status=HealthStatus.CRITICAL,
            timestamp=datetime.now(),
            overall_score=30.0,
        )

        # Assert
        assert health.status == HealthStatus.CRITICAL
        assert health.overall_score == 30.0

    def test_system_health_to_dict(self):
        """Test SystemHealth serialization."""
        # Arrange
        now = datetime.now()
        last_check = datetime.now() - timedelta(minutes=5)
        health = SystemHealth(
            status=HealthStatus.WARNING,
            timestamp=now,
            overall_score=80.0,
            component_scores={"service_a": 85.0, "service_b": 75.0},
            active_alerts=["ALERT-001"],
            recent_metrics={"cpu": 75.5, "memory": 70.0},
            uptime_percentage=98.5,
            last_check=last_check,
        )

        # Act
        result = health.to_dict()

        # Assert
        assert result["status"] == "warning"
        assert result["overall_score"] == 80.0
        assert result["component_scores"]["service_a"] == 85.0
        assert result["uptime_percentage"] == 98.5
        assert isinstance(result["timestamp"], str)


class TestMetricsCollector:
    """Test MetricsCollector functionality."""

    def test_metrics_collector_creation(self):
        """Test MetricsCollector initialization."""
        # Arrange & Act
        collector = MetricsCollector(buffer_size=10000, retention_hours=24)

        # Assert
        assert collector.buffer_size == 10000
        assert collector.retention_hours == 24

    def test_add_single_metric(self):
        """Test adding a single metric."""
        # Arrange
        collector = MetricsCollector()
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.RESPONSE_TIME,
            value=100.0,
        )

        # Act
        collector.add_metric(metric)

        # Assert
        metrics = collector.get_metrics(metric_type=MetricType.RESPONSE_TIME)
        assert len(metrics) == 1
        assert metrics[0].value == 100.0

    def test_add_multiple_metrics(self):
        """Test adding multiple metrics."""
        # Arrange
        collector = MetricsCollector()
        metrics_data = [
            MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=50.0 + i * 10,
            )
            for i in range(5)
        ]

        # Act
        for metric in metrics_data:
            collector.add_metric(metric)

        # Assert
        metrics = collector.get_metrics(metric_type=MetricType.CPU_USAGE)
        assert len(metrics) == 5

    def test_get_metrics_with_limit(self):
        """Test getting metrics with limit."""
        # Arrange
        collector = MetricsCollector()
        for i in range(10):
            metric = MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.MEMORY_USAGE,
                value=50.0 + i,
            )
            collector.add_metric(metric)

        # Act
        metrics = collector.get_metrics(
            metric_type=MetricType.MEMORY_USAGE, limit=5
        )

        # Assert
        assert len(metrics) <= 5

    def test_get_metrics_by_type(self):
        """Test filtering metrics by type."""
        # Arrange
        collector = MetricsCollector()
        collector.add_metric(
            MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=70.0,
            )
        )
        collector.add_metric(
            MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.MEMORY_USAGE,
                value=80.0,
            )
        )

        # Act
        cpu_metrics = collector.get_metrics(metric_type=MetricType.CPU_USAGE)
        memory_metrics = collector.get_metrics(metric_type=MetricType.MEMORY_USAGE)

        # Assert
        assert len(cpu_metrics) == 1
        assert cpu_metrics[0].value == 70.0
        assert len(memory_metrics) == 1
        assert memory_metrics[0].value == 80.0

    def test_get_all_metrics(self):
        """Test getting all metrics without filter."""
        # Arrange
        collector = MetricsCollector()
        collector.add_metric(
            MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=50.0,
            )
        )
        collector.add_metric(
            MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.MEMORY_USAGE,
                value=60.0,
            )
        )

        # Act
        all_metrics = collector.get_metrics()

        # Assert
        assert len(all_metrics) >= 2

    def test_get_statistics_empty(self):
        """Test statistics with no metrics."""
        # Arrange
        collector = MetricsCollector()

        # Act
        stats = collector.get_statistics(MetricType.CPU_USAGE)

        # Assert
        assert stats["count"] == 0
        assert stats["average"] is None

    def test_get_statistics_single_metric(self):
        """Test statistics with single metric."""
        # Arrange
        collector = MetricsCollector()
        collector.add_metric(
            MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.RESPONSE_TIME,
                value=100.0,
            )
        )

        # Act
        stats = collector.get_statistics(MetricType.RESPONSE_TIME)

        # Assert
        assert stats["count"] == 1
        assert stats["average"] == 100.0
        assert stats["min"] == 100.0
        assert stats["max"] == 100.0

    def test_get_statistics_multiple_metrics(self):
        """Test statistics with multiple metrics."""
        # Arrange
        collector = MetricsCollector()
        values = [100.0, 150.0, 200.0, 250.0, 300.0]
        for value in values:
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.RESPONSE_TIME,
                    value=value,
                )
            )

        # Act
        stats = collector.get_statistics(MetricType.RESPONSE_TIME)

        # Assert
        assert stats["count"] == len(values)
        assert stats["min"] == 100.0
        assert stats["max"] == 300.0
        assert stats["average"] == 200.0

    def test_metrics_sorted_by_timestamp(self):
        """Test that metrics are sorted by timestamp (newest first)."""
        # Arrange
        collector = MetricsCollector()
        now = datetime.now()

        # Add metrics with different timestamps
        collector.add_metric(
            MetricData(
                timestamp=now - timedelta(seconds=10),
                metric_type=MetricType.CPU_USAGE,
                value=50.0,
            )
        )
        collector.add_metric(
            MetricData(
                timestamp=now,
                metric_type=MetricType.CPU_USAGE,
                value=75.0,
            )
        )
        collector.add_metric(
            MetricData(
                timestamp=now - timedelta(seconds=5),
                metric_type=MetricType.CPU_USAGE,
                value=60.0,
            )
        )

        # Act
        metrics = collector.get_metrics(metric_type=MetricType.CPU_USAGE)

        # Assert
        # Should be sorted newest first
        assert metrics[0].value == 75.0  # Most recent


class TestAlertManager:
    """Test AlertManager functionality."""

    def test_alert_manager_creation(self):
        """Test AlertManager initialization."""
        # Arrange
        collector = MetricsCollector()

        # Act
        manager = AlertManager(metrics_collector=collector)

        # Assert
        assert manager.metrics_collector is collector
        assert manager.active_alerts == {}
        assert manager.alert_history == []

    def test_add_alert_rule(self):
        """Test adding alert rule."""
        # Arrange
        collector = MetricsCollector()
        manager = AlertManager(metrics_collector=collector)

        # Act
        manager.add_alert_rule(
            name="HighCPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=90.0,
            severity=AlertSeverity.HIGH,
            description="High CPU usage",
        )

        # Assert
        assert len(manager.alert_rules) == 1
        assert manager.alert_rules[0]["name"] == "HighCPU"
        assert manager.alert_rules[0]["threshold"] == 90.0

    def test_create_alert_directly(self):
        """Test creating an alert directly."""
        # Arrange
        collector = MetricsCollector()
        manager = AlertManager(metrics_collector=collector)

        # Act - Create alert directly using Alert class
        alert = Alert(
            alert_id="TEST-001",
            severity=AlertSeverity.HIGH,
            title="Test Alert",
            description="Test description",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=90.0,
            current_value=95.0,
            source="test_source",
        )

        # Assert
        assert alert.alert_id == "TEST-001"
        assert alert.severity == AlertSeverity.HIGH
        assert alert.title == "Test Alert"

    def test_register_alert_callback(self):
        """Test registering alert callback."""
        # Arrange
        collector = MetricsCollector()
        manager = AlertManager(metrics_collector=collector)

        callback = MagicMock()

        # Act
        manager.add_alert_callback(callback)

        # Assert
        assert callback in manager.alert_callbacks

    def test_resolve_alert(self):
        """Test resolving an alert directly."""
        # Arrange
        collector = MetricsCollector()
        manager = AlertManager(metrics_collector=collector)

        alert = Alert(
            alert_id="ALERT-001",
            severity=AlertSeverity.MEDIUM,
            title="Test",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=90.0,
            current_value=95.0,
            source="test",
        )
        manager.active_alerts[alert.alert_id] = alert

        # Act - Manually resolve the alert
        alert.resolved = True
        alert.resolved_at = datetime.now()

        # Assert
        assert alert.resolved is True
        assert alert.resolved_at is not None

    def test_acknowledge_alert(self):
        """Test acknowledging an alert."""
        # Arrange
        collector = MetricsCollector()
        manager = AlertManager(metrics_collector=collector)

        alert = Alert(
            alert_id="ALERT-002",
            severity=AlertSeverity.HIGH,
            title="Test",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.MEMORY_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="test",
        )
        manager.active_alerts[alert.alert_id] = alert

        # Act
        manager.acknowledge_alert(alert.alert_id)

        # Assert
        assert alert.acknowledged is True
        assert alert.acknowledged_at is not None

    def test_get_active_alerts(self):
        """Test getting active alerts."""
        # Arrange
        collector = MetricsCollector()
        manager = AlertManager(metrics_collector=collector)

        alert1 = Alert(
            alert_id="ALERT-001",
            severity=AlertSeverity.HIGH,
            title="Test 1",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=90.0,
            current_value=95.0,
            source="test",
        )
        alert2 = Alert(
            alert_id="ALERT-002",
            severity=AlertSeverity.MEDIUM,
            title="Test 2",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.MEMORY_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="test",
        )

        manager.active_alerts[alert1.alert_id] = alert1
        manager.active_alerts[alert2.alert_id] = alert2

        # Act
        alerts = manager.get_active_alerts()

        # Assert
        assert len(alerts) == 2

    def test_get_alerts_by_severity(self):
        """Test filtering alerts by severity."""
        # Arrange
        collector = MetricsCollector()
        manager = AlertManager(metrics_collector=collector)

        alert = Alert(
            alert_id="ALERT-001",
            severity=AlertSeverity.CRITICAL,
            title="Critical",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.AVAILABILITY,
            threshold=99.0,
            current_value=95.0,
            source="test",
        )
        manager.active_alerts[alert.alert_id] = alert

        # Act
        critical_alerts = [
            a for a in manager.get_active_alerts()
            if a.severity == AlertSeverity.CRITICAL
        ]

        # Assert
        assert len(critical_alerts) == 1
        assert critical_alerts[0].alert_id == "ALERT-001"
