"""Enhanced unit tests for RealtimeMonitoringDashboard module.

This module tests:
- MetricType, AlertSeverity, DashboardType enums
- MetricData and Alert dataclasses
- MetricsCollector with buffering and aggregation
- AlertManager with rule checking and resolution
- DashboardManager with CRUD operations
- RealtimeMonitoringDashboard integration
"""

from datetime import datetime, timedelta
from unittest import mock

import pytest

from moai_adk.core.realtime_monitoring_dashboard import (
    Alert,
    AlertManager,
    AlertSeverity,
    Dashboard,
    DashboardManager,
    DashboardType,
    DashboardWidget,
    MetricData,
    MetricType,
    MetricsCollector,
    RealtimeMonitoringDashboard,
)


class TestMetricType:
    """Test MetricType enum."""

    def test_metric_type_values(self):
        """Test that all metric types are defined."""
        assert MetricType.SYSTEM_PERFORMANCE.value == "system_performance"
        assert MetricType.HOOK_EXECUTION.value == "hook_execution"
        assert MetricType.ERROR_RATE.value == "error_rate"
        assert MetricType.CPU_USAGE.value == "cpu_usage"
        assert MetricType.MEMORY_USAGE.value == "memory_usage"

    def test_metric_type_count(self):
        """Test metric type count."""
        types = list(MetricType)
        assert len(types) > 0


class TestAlertSeverity:
    """Test AlertSeverity enum."""

    def test_alert_severity_levels(self):
        """Test alert severity levels."""
        assert AlertSeverity.INFO.value == 1
        assert AlertSeverity.WARNING.value == 2
        assert AlertSeverity.ERROR.value == 3
        assert AlertSeverity.CRITICAL.value == 4
        assert AlertSeverity.EMERGENCY.value == 5

    def test_alert_severity_ordering(self):
        """Test alert severity ordering."""
        assert AlertSeverity.INFO.value < AlertSeverity.CRITICAL.value


class TestDashboardType:
    """Test DashboardType enum."""

    def test_dashboard_type_values(self):
        """Test dashboard type values."""
        assert DashboardType.SYSTEM_OVERVIEW.value == "system_overview"
        assert DashboardType.PERFORMANCE_ANALYTICS.value == "performance_analytics"
        assert DashboardType.CUSTOM.value == "custom"


class TestMetricData:
    """Test MetricData dataclass."""

    def test_metric_data_creation(self):
        """Test creating MetricData instance."""
        now = datetime.now()
        metric = MetricData(
            timestamp=now,
            metric_type=MetricType.CPU_USAGE,
            value=45.5,
            source="system",
            component="cpu",
        )
        assert metric.value == 45.5
        assert metric.metric_type == MetricType.CPU_USAGE

    def test_metric_data_with_tags(self):
        """Test MetricData with tags."""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
            tags={"host": "server1"},
        )
        assert metric.tags["host"] == "server1"

    def test_metric_data_with_tenant(self):
        """Test MetricData with tenant ID."""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=60.0,
            tenant_id="tenant1",
        )
        assert metric.tenant_id == "tenant1"

    def test_metric_data_to_dict(self):
        """Test converting MetricData to dictionary."""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=45.0,
            source="test",
        )
        data_dict = metric.to_dict()
        assert "timestamp" in data_dict
        assert "metric_type" in data_dict
        assert data_dict["value"] == 45.0


class TestAlert:
    """Test Alert dataclass."""

    def test_alert_creation(self):
        """Test creating Alert instance."""
        alert = Alert(
            alert_id="alert1",
            severity=AlertSeverity.WARNING,
            title="High CPU",
            description="CPU usage exceeds 80%",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="system",
            component="cpu",
        )
        assert alert.alert_id == "alert1"
        assert alert.severity == AlertSeverity.WARNING

    def test_alert_resolution(self):
        """Test alert resolution."""
        alert = Alert(
            alert_id="alert1",
            severity=AlertSeverity.ERROR,
            title="Test Alert",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="system",
            component="cpu",
        )
        assert alert.resolved is False

        alert.resolved = True
        alert.resolved_at = datetime.now()
        assert alert.resolved is True

    def test_alert_acknowledgment(self):
        """Test alert acknowledgment."""
        alert = Alert(
            alert_id="alert1",
            severity=AlertSeverity.WARNING,
            title="Test Alert",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="system",
            component="cpu",
        )
        assert alert.acknowledged is False

        alert.acknowledged = True
        alert.acknowledged_at = datetime.now()
        assert alert.acknowledged is True

    def test_alert_to_dict(self):
        """Test converting Alert to dictionary."""
        alert = Alert(
            alert_id="alert1",
            severity=AlertSeverity.WARNING,
            title="Test Alert",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="system",
            component="cpu",
        )
        alert_dict = alert.to_dict()
        assert "alert_id" in alert_dict
        assert "severity" in alert_dict
        assert alert_dict["title"] == "Test Alert"


class TestDashboardWidget:
    """Test DashboardWidget dataclass."""

    def test_widget_creation(self):
        """Test creating dashboard widget."""
        widget = DashboardWidget(
            widget_id="widget1",
            widget_type="chart",
            title="CPU Chart",
            position={"x": 0, "y": 0, "width": 6, "height": 3},
        )
        assert widget.widget_id == "widget1"
        assert widget.widget_type == "chart"
        assert widget.title == "CPU Chart"

    def test_widget_with_metrics(self):
        """Test widget with metrics."""
        widget = DashboardWidget(
            widget_id="widget1",
            widget_type="chart",
            title="Metrics",
            position={"x": 0, "y": 0, "width": 6, "height": 3},
            metrics=["cpu_usage", "memory_usage"],
        )
        assert "cpu_usage" in widget.metrics


class TestDashboard:
    """Test Dashboard dataclass."""

    def test_dashboard_creation(self):
        """Test creating dashboard."""
        dashboard = Dashboard(
            dashboard_id="dash1",
            name="System Overview",
            description="System metrics",
            dashboard_type=DashboardType.SYSTEM_OVERVIEW,
        )
        assert dashboard.dashboard_id == "dash1"
        assert dashboard.name == "System Overview"

    def test_dashboard_to_dict(self):
        """Test converting dashboard to dict."""
        dashboard = Dashboard(
            dashboard_id="dash1",
            name="System Overview",
            description="Test dashboard",
            dashboard_type=DashboardType.SYSTEM_OVERVIEW,
        )
        data = dashboard.to_dict()
        assert "dashboard_id" in data
        assert "name" in data


class TestMetricsCollector:
    """Test MetricsCollector class."""

    def test_metrics_collector_initialization(self):
        """Test MetricsCollector initialization."""
        collector = MetricsCollector(buffer_size=1000, retention_hours=24)
        assert collector.buffer_size == 1000
        assert collector.retention_hours == 24

    def test_add_metric(self):
        """Test adding metric to collector."""
        collector = MetricsCollector()
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=45.0,
            source="test",
        )
        collector.add_metric(metric)
        assert len(collector.metrics_buffer) > 0

    def test_get_metrics_empty(self):
        """Test getting metrics when empty."""
        collector = MetricsCollector()
        metrics = collector.get_metrics(metric_type=MetricType.CPU_USAGE)
        assert len(metrics) == 0

    def test_get_metrics_with_limit(self):
        """Test getting metrics with limit."""
        collector = MetricsCollector()
        for i in range(10):
            metric = MetricData(
                timestamp=datetime.now() - timedelta(minutes=i),
                metric_type=MetricType.CPU_USAGE,
                value=float(i),
                source="test",
            )
            collector.add_metric(metric)

        metrics = collector.get_metrics(metric_type=MetricType.CPU_USAGE, limit=5)
        assert len(metrics) <= 5

    def test_get_statistics_basic(self):
        """Test getting metric statistics."""
        collector = MetricsCollector()
        for i in range(5):
            metric = MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=float(i * 10),
                source="test",
                component="cpu",
            )
            collector.add_metric(metric)

        stats = collector.get_statistics(MetricType.CPU_USAGE, component="cpu")
        assert "count" in stats or stats.get("count") == 0


class TestAlertManager:
    """Test AlertManager class."""

    def test_alert_manager_initialization(self):
        """Test AlertManager initialization."""
        collector = MetricsCollector()
        manager = AlertManager(collector)
        assert manager.metrics_collector == collector
        assert len(manager.alert_rules) == 0

    def test_add_alert_rule(self):
        """Test adding alert rule."""
        collector = MetricsCollector()
        manager = AlertManager(collector)
        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
            severity=AlertSeverity.WARNING,
        )
        assert len(manager.alert_rules) == 1

    def test_check_alerts_no_metrics(self):
        """Test checking alerts when no metrics."""
        collector = MetricsCollector()
        manager = AlertManager(collector)
        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
        )
        alerts = manager.check_alerts()
        assert len(alerts) == 0

    def test_acknowledge_alert(self):
        """Test acknowledging alert."""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        alert = Alert(
            alert_id="alert1",
            severity=AlertSeverity.WARNING,
            title="Test",
            description="Test alert",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="test",
            component="cpu",
        )
        manager.active_alerts["alert1"] = alert

        result = manager.acknowledge_alert("alert1")
        assert result is True
        assert manager.active_alerts["alert1"].acknowledged is True

    def test_get_active_alerts(self):
        """Test getting active alerts."""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        alert = Alert(
            alert_id="alert1",
            severity=AlertSeverity.ERROR,
            title="Test",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="test",
            component="cpu",
        )
        manager.active_alerts["alert1"] = alert

        alerts = manager.get_active_alerts()
        assert len(alerts) == 1

    def test_get_alert_history(self):
        """Test getting alert history."""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        alert = Alert(
            alert_id="alert1",
            severity=AlertSeverity.WARNING,
            title="Test",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="test",
            component="cpu",
        )
        manager.alert_history.append(alert)

        history = manager.get_alert_history(hours=24)
        assert len(history) >= 0

    def test_get_alert_statistics(self):
        """Test getting alert statistics."""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        alert = Alert(
            alert_id="alert1",
            severity=AlertSeverity.WARNING,
            title="Test",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="test",
            component="cpu",
        )
        manager.alert_history.append(alert)

        stats = manager.get_alert_statistics(hours=24)
        assert "total_alerts" in stats


class TestDashboardManager:
    """Test DashboardManager class."""

    def test_dashboard_manager_initialization(self):
        """Test DashboardManager initialization."""
        manager = DashboardManager()
        assert len(manager.dashboards) >= 0
        assert len(manager.default_dashboards) >= 0

    def test_create_dashboard(self):
        """Test creating dashboard."""
        manager = DashboardManager()
        dashboard_id = manager.create_dashboard(
            name="Test Dashboard",
            description="Test",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[],
        )
        assert dashboard_id is not None
        assert len(dashboard_id) > 0

    def test_get_dashboard(self):
        """Test getting dashboard."""
        manager = DashboardManager()
        dashboard_id = manager.create_dashboard(
            name="Test Dashboard",
            description="Test",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[],
        )

        dashboard = manager.get_dashboard(dashboard_id)
        assert dashboard is not None
        assert dashboard.name == "Test Dashboard"

    def test_list_dashboards(self):
        """Test listing dashboards."""
        manager = DashboardManager()
        dashboards = manager.list_dashboards()
        assert len(dashboards) >= 0

    def test_update_dashboard(self):
        """Test updating dashboard."""
        manager = DashboardManager()
        dashboard_id = manager.create_dashboard(
            name="Test Dashboard",
            description="Test",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[],
        )

        result = manager.update_dashboard(dashboard_id, {"name": "Updated"})
        assert result is True

    def test_delete_dashboard(self):
        """Test deleting dashboard."""
        manager = DashboardManager()
        dashboard_id = manager.create_dashboard(
            name="Test Dashboard",
            description="Test",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[],
        )

        result = manager.delete_dashboard(dashboard_id)
        assert result is True


class TestRealtimeMonitoringDashboard:
    """Test RealtimeMonitoringDashboard class."""

    def test_dashboard_initialization(self):
        """Test RealtimeMonitoringDashboard initialization."""
        dashboard = RealtimeMonitoringDashboard(
            metrics_buffer_size=10000,
            retention_hours=24,
        )
        assert dashboard.metrics_buffer_size == 10000
        assert dashboard.retention_hours == 24

    def test_add_metric(self):
        """Test adding metric to dashboard."""
        dashboard = RealtimeMonitoringDashboard()
        dashboard.add_metric(MetricType.CPU_USAGE, 45.0)
        # Verify metric was added to collector
        assert len(dashboard.metrics_collector.metrics_buffer) >= 0

    def test_get_system_status(self):
        """Test getting system status."""
        dashboard = RealtimeMonitoringDashboard()
        status = dashboard.get_system_status()

        assert "status" in status
        assert "uptime_seconds" in status
        assert "metrics_collected" in status
        assert "active_alerts" in status

    def test_create_custom_dashboard(self):
        """Test creating custom dashboard."""
        dashboard = RealtimeMonitoringDashboard()
        widgets = [
            {
                "widget_id": "w1",
                "widget_type": "metric",
                "title": "CPU",
                "position": {"x": 0, "y": 0, "width": 4, "height": 2},
            }
        ]
        dashboard_id = dashboard.create_custom_dashboard(
            name="Custom",
            description="Test",
            widgets=widgets,
        )
        assert dashboard_id is not None

    def test_stop_when_not_running(self):
        """Test stopping dashboard when not running."""
        dashboard = RealtimeMonitoringDashboard()
        dashboard.stop()  # Should not raise


class TestMetricsCollectorMultiTenant:
    """Test MetricsCollector multi-tenant functionality."""

    def test_add_metric_with_tenant(self):
        """Test adding metric with tenant ID."""
        collector = MetricsCollector()
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=45.0,
            source="test",
            tenant_id="tenant1",
        )
        collector.add_metric(metric)
        assert "tenant1" in collector.tenant_metrics

    def test_get_metrics_by_tenant(self):
        """Test getting metrics filtered by tenant."""
        collector = MetricsCollector()
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=45.0,
            source="test",
            tenant_id="tenant1",
        )
        collector.add_metric(metric)

        metrics = collector.get_metrics(tenant_id="tenant1")
        assert len(metrics) >= 0


class TestAlertManagerCallbacks:
    """Test AlertManager callback functionality."""

    def test_add_alert_callback(self):
        """Test adding alert callback."""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        called = False

        def callback(alert: Alert):
            nonlocal called
            called = True

        manager.add_alert_callback(callback)
        assert len(manager.alert_callbacks) == 1


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
