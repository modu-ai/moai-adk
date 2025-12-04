"""Simple tests for realtime_monitoring_dashboard module.

Tests basic class instantiation, dataclasses, and simple methods
with mocked dependencies.
"""

import unittest
from datetime import datetime
from unittest.mock import MagicMock, Mock, patch

from src.moai_adk.core.realtime_monitoring_dashboard import (
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


class TestMetricData(unittest.TestCase):
    """Test MetricData dataclass."""

    def test_metric_data_creation(self):
        """Test creating a MetricData instance."""
        # Arrange
        now = datetime.now()
        metric = MetricData(
            timestamp=now,
            metric_type=MetricType.CPU_USAGE,
            value=42.5,
            tags={"component": "system"},
            source="test",
            component="test_comp",
            environment="test",
        )

        # Act
        result = metric.to_dict()

        # Assert
        self.assertEqual(result["value"], 42.5)
        self.assertEqual(result["metric_type"], "cpu_usage")
        self.assertEqual(result["source"], "test")
        self.assertIn("timestamp", result)

    def test_metric_data_with_defaults(self):
        """Test MetricData with default values."""
        # Arrange
        now = datetime.now()
        metric = MetricData(
            timestamp=now,
            metric_type=MetricType.MEMORY_USAGE,
            value=85,
        )

        # Act
        result = metric.to_dict()

        # Assert
        self.assertEqual(result["value"], 85)
        self.assertEqual(result["tags"], {})
        self.assertIsNone(result["tenant_id"])


class TestAlert(unittest.TestCase):
    """Test Alert dataclass."""

    def test_alert_creation(self):
        """Test creating an Alert instance."""
        # Arrange
        now = datetime.now()
        alert = Alert(
            alert_id="alert_001",
            severity=AlertSeverity.WARNING,
            title="Test Alert",
            description="Test alert description",
            timestamp=now,
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="test_source",
            component="test_component",
        )

        # Act
        result = alert.to_dict()

        # Assert
        self.assertEqual(result["alert_id"], "alert_001")
        self.assertEqual(result["severity"], 2)  # AlertSeverity.WARNING value
        self.assertEqual(result["current_value"], 85.0)
        self.assertFalse(result["resolved"])

    def test_alert_to_dict(self):
        """Test Alert to_dict conversion."""
        # Arrange
        now = datetime.now()
        alert = Alert(
            alert_id="test_alert",
            severity=AlertSeverity.CRITICAL,
            title="Critical Alert",
            description="Critical issue",
            timestamp=now,
            metric_type=MetricType.ERROR_RATE,
            threshold=5.0,
            current_value=10.0,
            source="system",
            component="api",
            resolved=True,
            resolved_at=now,
        )

        # Act
        result = alert.to_dict()

        # Assert
        self.assertTrue(result["resolved"])
        self.assertIsNotNone(result["resolved_at"])


class TestDashboardWidget(unittest.TestCase):
    """Test DashboardWidget dataclass."""

    def test_widget_creation(self):
        """Test creating a DashboardWidget instance."""
        # Arrange
        widget = DashboardWidget(
            widget_id="widget_001",
            widget_type="chart",
            title="CPU Usage Chart",
            position={"x": 0, "y": 0, "width": 6, "height": 3},
            config={"chart_type": "line"},
            metrics=["cpu_usage"],
        )

        # Act
        # Assert
        self.assertEqual(widget.widget_id, "widget_001")
        self.assertEqual(widget.widget_type, "chart")
        self.assertEqual(widget.title, "CPU Usage Chart")
        self.assertIn("x", widget.position)


class TestDashboard(unittest.TestCase):
    """Test Dashboard dataclass."""

    def test_dashboard_creation(self):
        """Test creating a Dashboard instance."""
        # Arrange
        dashboard = Dashboard(
            dashboard_id="dash_001",
            name="System Overview",
            description="System health dashboard",
            dashboard_type=DashboardType.SYSTEM_OVERVIEW,
            owner="admin",
            is_public=True,
        )

        # Act
        result = dashboard.to_dict()

        # Assert
        self.assertEqual(result["dashboard_id"], "dash_001")
        self.assertEqual(result["name"], "System Overview")
        self.assertTrue(result["is_public"])
        self.assertEqual(result["dashboard_type"], "system_overview")


class TestMetricsCollector(unittest.TestCase):
    """Test MetricsCollector class."""

    def test_metrics_collector_initialization(self):
        """Test MetricsCollector initialization."""
        # Arrange
        # Act
        collector = MetricsCollector(buffer_size=1000, retention_hours=24)

        # Assert
        self.assertEqual(collector.buffer_size, 1000)
        self.assertEqual(collector.retention_hours, 24)
        self.assertIsNotNone(collector.metrics_buffer)
        self.assertIsNotNone(collector.aggregated_metrics)

    def test_add_metric(self):
        """Test adding a metric to the collector."""
        # Arrange
        collector = MetricsCollector()
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
            component="system",
        )

        # Act
        collector.add_metric(metric)

        # Assert
        key = "cpu_usage:system"
        self.assertIn(key, collector.metrics_buffer)
        self.assertGreater(len(collector.metrics_buffer[key]), 0)

    def test_get_metrics_empty(self):
        """Test getting metrics when none exist."""
        # Arrange
        collector = MetricsCollector()

        # Act
        metrics = collector.get_metrics(metric_type=MetricType.CPU_USAGE)

        # Assert
        self.assertEqual(len(metrics), 0)

    def test_get_statistics_empty(self):
        """Test getting statistics with no data."""
        # Arrange
        collector = MetricsCollector()

        # Act
        stats = collector.get_statistics(MetricType.CPU_USAGE)

        # Assert
        self.assertEqual(stats["count"], 0)
        self.assertIsNone(stats["average"])

    def test_percentile_calculation(self):
        """Test percentile calculation."""
        # Arrange
        collector = MetricsCollector()
        values = [1.0, 2.0, 3.0, 4.0, 5.0]

        # Act
        p95 = collector._percentile(values, 95)

        # Assert
        self.assertGreater(p95, 0)
        self.assertLessEqual(p95, 5.0)


class TestAlertManager(unittest.TestCase):
    """Test AlertManager class."""

    def setUp(self):
        """Set up test fixtures."""
        self.metrics_collector = MetricsCollector()
        self.alert_manager = AlertManager(self.metrics_collector)

    def test_alert_manager_initialization(self):
        """Test AlertManager initialization."""
        # Arrange
        # Act
        # Assert
        self.assertEqual(len(self.alert_manager.alert_rules), 0)
        self.assertEqual(len(self.alert_manager.active_alerts), 0)

    def test_add_alert_rule(self):
        """Test adding an alert rule."""
        # Arrange
        # Act
        self.alert_manager.add_alert_rule(
            name="Test Alert",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
            severity=AlertSeverity.WARNING,
        )

        # Assert
        self.assertEqual(len(self.alert_manager.alert_rules), 1)
        rule = self.alert_manager.alert_rules[0]
        self.assertEqual(rule["name"], "Test Alert")
        self.assertEqual(rule["threshold"], 80.0)

    def test_add_alert_callback(self):
        """Test adding an alert callback."""
        # Arrange
        callback = Mock()

        # Act
        self.alert_manager.add_alert_callback(callback)

        # Assert
        self.assertEqual(len(self.alert_manager.alert_callbacks), 1)

    def test_acknowledge_alert_nonexistent(self):
        """Test acknowledging a nonexistent alert."""
        # Arrange
        # Act
        result = self.alert_manager.acknowledge_alert("nonexistent_alert")

        # Assert
        self.assertFalse(result)

    def test_get_active_alerts_empty(self):
        """Test getting active alerts when none exist."""
        # Arrange
        # Act
        alerts = self.alert_manager.get_active_alerts()

        # Assert
        self.assertEqual(len(alerts), 0)

    def test_get_alert_history_empty(self):
        """Test getting alert history when empty."""
        # Arrange
        # Act
        history = self.alert_manager.get_alert_history(hours=24)

        # Assert
        self.assertEqual(len(history), 0)

    def test_get_alert_statistics_empty(self):
        """Test getting alert statistics when empty."""
        # Arrange
        # Act
        stats = self.alert_manager.get_alert_statistics(hours=24)

        # Assert
        self.assertEqual(stats["total_alerts"], 0)
        self.assertEqual(stats["resolved_count"], 0)

    def test_evaluate_condition_gt(self):
        """Test greater than condition evaluation."""
        # Arrange
        # Act
        result = self.alert_manager._evaluate_condition(85.0, 80.0, "gt")

        # Assert
        self.assertTrue(result)

    def test_evaluate_condition_lt(self):
        """Test less than condition evaluation."""
        # Arrange
        # Act
        result = self.alert_manager._evaluate_condition(75.0, 80.0, "lt")

        # Assert
        self.assertTrue(result)

    def test_evaluate_condition_eq(self):
        """Test equals condition evaluation."""
        # Arrange
        # Act
        result = self.alert_manager._evaluate_condition(80.0, 80.0, "eq")

        # Assert
        self.assertTrue(result)


class TestDashboardManager(unittest.TestCase):
    """Test DashboardManager class."""

    def setUp(self):
        """Set up test fixtures."""
        self.dashboard_manager = DashboardManager()

    def test_dashboard_manager_initialization(self):
        """Test DashboardManager initialization."""
        # Arrange
        # Act
        # Assert
        self.assertGreater(len(self.dashboard_manager.default_dashboards), 0)

    def test_create_dashboard(self):
        """Test creating a dashboard."""
        # Arrange
        widgets = []

        # Act
        dashboard_id = self.dashboard_manager.create_dashboard(
            name="Test Dashboard",
            description="Test dashboard",
            dashboard_type=DashboardType.CUSTOM,
            widgets=widgets,
            owner="test_user",
        )

        # Assert
        self.assertIsNotNone(dashboard_id)
        dashboard = self.dashboard_manager.get_dashboard(dashboard_id)
        self.assertIsNotNone(dashboard)
        self.assertEqual(dashboard.name, "Test Dashboard")

    def test_get_dashboard_default(self):
        """Test getting a default dashboard."""
        # Arrange
        # Act
        dashboard = self.dashboard_manager.get_dashboard("system_overview")

        # Assert
        self.assertIsNotNone(dashboard)
        self.assertEqual(dashboard.name, "System Overview")

    def test_list_dashboards(self):
        """Test listing dashboards."""
        # Arrange
        # Add a new dashboard for this test
        dashboard_id = self.dashboard_manager.create_dashboard(
            name="Test Dashboard",
            description="Test",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[],
        )

        # Act
        dashboards = self.dashboard_manager.list_dashboards(is_public=None)

        # Assert
        self.assertGreater(len(dashboards), 0)

    def test_list_dashboards_by_type(self):
        """Test listing dashboards by type."""
        # Arrange
        # Act
        dashboards = self.dashboard_manager.list_dashboards(dashboard_type=DashboardType.SYSTEM_OVERVIEW)

        # Assert
        for dashboard in dashboards:
            self.assertEqual(dashboard.dashboard_type, DashboardType.SYSTEM_OVERVIEW)

    def test_update_dashboard(self):
        """Test updating a dashboard."""
        # Arrange
        dashboard_id = self.dashboard_manager.create_dashboard(
            name="Original Name",
            description="Original description",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[],
        )

        # Act
        result = self.dashboard_manager.update_dashboard(dashboard_id, {"name": "Updated Name"})

        # Assert
        self.assertTrue(result)
        dashboard = self.dashboard_manager.get_dashboard(dashboard_id)
        self.assertEqual(dashboard.name, "Updated Name")

    def test_delete_dashboard(self):
        """Test deleting a dashboard."""
        # Arrange
        dashboard_id = self.dashboard_manager.create_dashboard(
            name="To Delete",
            description="Delete me",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[],
        )

        # Act
        result = self.dashboard_manager.delete_dashboard(dashboard_id)

        # Assert
        self.assertTrue(result)
        dashboard = self.dashboard_manager.get_dashboard(dashboard_id)
        self.assertIsNone(dashboard)

    def test_delete_default_dashboard_fails(self):
        """Test that deleting a default dashboard fails."""
        # Arrange
        # Act
        result = self.dashboard_manager.delete_dashboard("system_overview")

        # Assert
        self.assertFalse(result)


class TestRealtimeMonitoringDashboard(unittest.TestCase):
    """Test RealtimeMonitoringDashboard class."""

    def setUp(self):
        """Set up test fixtures."""
        self.dashboard = RealtimeMonitoringDashboard(enable_websocket=False, enable_external_integration=False)

    def test_dashboard_initialization(self):
        """Test RealtimeMonitoringDashboard initialization."""
        # Arrange
        # Act
        # Assert
        self.assertIsNotNone(self.dashboard.metrics_collector)
        self.assertIsNotNone(self.dashboard.alert_manager)
        self.assertIsNotNone(self.dashboard.dashboard_manager)
        self.assertFalse(self.dashboard._running)

    def test_add_metric(self):
        """Test adding a metric."""
        # Arrange
        # Act
        self.dashboard.add_metric(MetricType.CPU_USAGE, 45.0, component="test_system")

        # Assert
        metrics = self.dashboard.metrics_collector.get_metrics(metric_type=MetricType.CPU_USAGE)
        self.assertGreater(len(metrics), 0)

    def test_get_system_status(self):
        """Test getting system status."""
        # Arrange
        # Act
        status = self.dashboard.get_system_status()

        # Assert
        self.assertIn("status", status)
        self.assertIn("uptime_seconds", status)
        self.assertIn("metrics_collected", status)
        self.assertEqual(status["status"], "stopped")

    def test_create_custom_dashboard(self):
        """Test creating a custom dashboard."""
        # Arrange
        widgets = [
            {
                "widget_id": "test_widget",
                "widget_type": "metric",
                "title": "Test Widget",
                "position": {"x": 0, "y": 0, "width": 4, "height": 2},
            }
        ]

        # Act
        dashboard_id = self.dashboard.create_custom_dashboard(
            name="Custom Dashboard",
            description="Test custom dashboard",
            widgets=widgets,
            owner="test_user",
        )

        # Assert
        self.assertIsNotNone(dashboard_id)

    def test_get_dashboard_data_not_found(self):
        """Test getting dashboard data for nonexistent dashboard."""
        # Arrange
        # Act
        result = self.dashboard.get_dashboard_data("nonexistent_dashboard")

        # Assert
        self.assertIn("error", result)

    def test_get_dashboard_data_system_overview(self):
        """Test getting data for system overview dashboard."""
        # Arrange
        # Act
        result = self.dashboard.get_dashboard_data("system_overview")

        # Assert
        self.assertIn("dashboard", result)
        self.assertIn("widgets_data", result)

    def test_map_metric_name(self):
        """Test metric name mapping."""
        # Arrange
        # Act
        metric_type = self.dashboard._map_metric_name("cpu_usage")

        # Assert
        self.assertEqual(metric_type, MetricType.CPU_USAGE)

    def test_map_metric_name_unknown(self):
        """Test mapping unknown metric name."""
        # Arrange
        # Act
        metric_type = self.dashboard._map_metric_name("unknown_metric")

        # Assert
        self.assertIsNone(metric_type)

    def test_collect_system_metrics_with_metrics(self):
        """Test collecting system metrics."""
        # Arrange
        # Act - Add metrics directly since psutil is optional
        self.dashboard.add_metric(MetricType.CPU_USAGE, 45.5, component="system")
        self.dashboard.add_metric(MetricType.MEMORY_USAGE, 62.0, component="system")

        # Assert
        metrics = self.dashboard.metrics_collector.get_metrics()
        self.assertGreater(len(metrics), 0)


class TestMonitoringDashboardIntegration(unittest.TestCase):
    """Integration tests for monitoring dashboard."""

    def test_full_monitoring_workflow(self):
        """Test a complete monitoring workflow."""
        # Arrange
        dashboard = RealtimeMonitoringDashboard(enable_websocket=False, enable_external_integration=False)

        # Act
        dashboard.add_metric(MetricType.CPU_USAGE, 45.0, component="test")
        dashboard.add_metric(MetricType.MEMORY_USAGE, 60.0, component="test")

        status = dashboard.get_system_status()
        metrics = dashboard.metrics_collector.get_metrics()

        # Assert
        self.assertEqual(status["status"], "stopped")
        self.assertGreaterEqual(len(metrics), 2)

    def test_alert_workflow(self):
        """Test alert creation and management workflow."""
        # Arrange
        dashboard = RealtimeMonitoringDashboard(enable_websocket=False, enable_external_integration=False)
        # Note: Dashboard initializes with 4 default alert rules

        # Act - Add a custom alert rule
        initial_rule_count = len(dashboard.alert_manager.alert_rules)
        dashboard.alert_manager.add_alert_rule(
            name="High CPU Custom",
            metric_type=MetricType.CPU_USAGE,
            threshold=90.0,
            operator="gt",
            window_minutes=1,
            consecutive_violations=1,
        )

        # Add metrics that will trigger alert
        for i in range(3):
            dashboard.add_metric(MetricType.CPU_USAGE, 95.0, component="test")

        # Assert
        self.assertEqual(len(dashboard.alert_manager.alert_rules), initial_rule_count + 1)
        custom_rule = [r for r in dashboard.alert_manager.alert_rules if r["name"] == "High CPU Custom"][0]
        self.assertEqual(custom_rule["name"], "High CPU Custom")


if __name__ == "__main__":
    unittest.main()
