"""
Comprehensive tests for realtime_monitoring_dashboard.py module.

Provides extensive coverage for:
- MetricData dataclass and serialization
- Alert dataclass and management
- MetricsCollector with multi-tenant support
- AlertManager with correlation and rule evaluation
- DashboardManager with widget management
- RealtimeMonitoringDashboard system integration
- All exception handling and edge cases
"""

import asyncio
import threading
from datetime import datetime, timedelta
from unittest.mock import MagicMock, Mock, patch

import pytest

from src.moai_adk.core.realtime_monitoring_dashboard import (
    Alert,
    AlertManager,
    AlertSeverity,
    Dashboard,
    DashboardManager,
    DashboardType,
    DashboardWidget,
    MetricData,
    MetricsCollector,
    MetricType,
    RealtimeMonitoringDashboard,
    add_hook_metric,
    add_system_metric,
    get_monitoring_dashboard,
    start_monitoring,
    stop_monitoring,
)

# ============================================================================
# MetricData Tests
# ============================================================================


class TestMetricData:
    """Tests for MetricData dataclass"""

    def test_metric_data_creation(self):
        """Test MetricData instantiation with all fields"""
        now = datetime.now()
        metric = MetricData(
            timestamp=now,
            metric_type=MetricType.CPU_USAGE,
            value=75.5,
            tags={"host": "server1"},
            source="psutil",
            metadata={"extra": "info"},
            component="system",
            environment="production",
            tenant_id="tenant1",
        )

        assert metric.timestamp == now
        assert metric.metric_type == MetricType.CPU_USAGE
        assert metric.value == 75.5
        assert metric.tags == {"host": "server1"}
        assert metric.source == "psutil"
        assert metric.component == "system"
        assert metric.environment == "production"
        assert metric.tenant_id == "tenant1"

    def test_metric_data_defaults(self):
        """Test MetricData with default values"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.MEMORY_USAGE,
            value=50,
        )

        assert metric.tags == {}
        assert metric.source == ""
        assert metric.metadata == {}
        assert metric.component == ""
        assert metric.environment == "production"
        assert metric.tenant_id is None

    def test_metric_data_to_dict(self):
        """Test MetricData serialization to dictionary"""
        now = datetime.now()
        metric = MetricData(
            timestamp=now,
            metric_type=MetricType.ERROR_RATE,
            value=2.5,
            tags={"service": "api"},
            source="custom",
            component="api_server",
            tenant_id="t1",
        )

        result = metric.to_dict()

        assert result["timestamp"] == now.isoformat()
        assert result["metric_type"] == "error_rate"
        assert result["value"] == 2.5
        assert result["tags"] == {"service": "api"}
        assert result["source"] == "custom"
        assert result["component"] == "api_server"
        assert result["tenant_id"] == "t1"
        assert result["environment"] == "production"

    def test_metric_data_string_value(self):
        """Test MetricData with string value"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.SYSTEM_PERFORMANCE,
            value="healthy",
        )

        assert metric.value == "healthy"
        result = metric.to_dict()
        assert result["value"] == "healthy"

    def test_metric_data_bool_value(self):
        """Test MetricData with boolean value"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.AVAILABILITY,
            value=True,
        )

        assert metric.value is True
        result = metric.to_dict()
        assert result["value"] is True


# ============================================================================
# Alert Tests
# ============================================================================


class TestAlert:
    """Tests for Alert dataclass"""

    def test_alert_creation(self):
        """Test Alert instantiation with all fields"""
        now = datetime.now()
        alert = Alert(
            alert_id="alert_001",
            severity=AlertSeverity.CRITICAL,
            title="High CPU Usage",
            description="CPU usage exceeded 90%",
            timestamp=now,
            metric_type=MetricType.CPU_USAGE,
            threshold=90.0,
            current_value=95.5,
            source="monitoring_system",
            component="system",
            tags={"host": "server1"},
            resolved=False,
            acknowledged=True,
            acknowledged_at=now,
            correlation_id="corr_001",
            context={"details": "test"},
            affected_services=["api", "worker"],
            recovery_actions=["restart_service"],
            tenant_id="tenant1",
        )

        assert alert.alert_id == "alert_001"
        assert alert.severity == AlertSeverity.CRITICAL
        assert alert.title == "High CPU Usage"
        assert alert.resolved is False
        assert alert.acknowledged is True
        assert alert.tenant_id == "tenant1"

    def test_alert_defaults(self):
        """Test Alert with default values"""
        now = datetime.now()
        alert = Alert(
            alert_id="alert_002",
            severity=AlertSeverity.WARNING,
            title="Warning",
            description="A warning",
            timestamp=now,
            metric_type=MetricType.MEMORY_USAGE,
            threshold=80.0,
            current_value=82.0,
            source="system",
            component="memory",
        )

        assert alert.resolved is False
        assert alert.acknowledged is False
        assert alert.resolved_at is None
        assert alert.acknowledged_at is None
        assert alert.correlation_id is None
        assert alert.context == {}
        assert alert.affected_services == []
        assert alert.recovery_actions == []
        assert alert.tenant_id is None

    def test_alert_to_dict(self):
        """Test Alert serialization to dictionary"""
        now = datetime.now()
        resolved_at = now + timedelta(hours=1)
        alert = Alert(
            alert_id="alert_003",
            severity=AlertSeverity.ERROR,
            title="Test Alert",
            description="Test description",
            timestamp=now,
            metric_type=MetricType.ERROR_RATE,
            threshold=5.0,
            current_value=8.5,
            source="test",
            component="test_component",
            tags={"env": "test"},
            resolved=True,
            resolved_at=resolved_at,
            acknowledged=True,
            acknowledged_at=now,
        )

        result = alert.to_dict()

        assert result["alert_id"] == "alert_003"
        assert result["severity"] == AlertSeverity.ERROR.value
        assert result["timestamp"] == now.isoformat()
        assert result["resolved"] is True
        assert result["resolved_at"] == resolved_at.isoformat()
        assert result["acknowledged"] is True
        assert result["acknowledged_at"] == now.isoformat()

    def test_alert_to_dict_none_dates(self):
        """Test Alert serialization with None dates"""
        now = datetime.now()
        alert = Alert(
            alert_id="alert_004",
            severity=AlertSeverity.INFO,
            title="Info",
            description="Info alert",
            timestamp=now,
            metric_type=MetricType.THROUGHPUT,
            threshold=100.0,
            current_value=150.0,
            source="system",
            component="throughput",
        )

        result = alert.to_dict()

        assert result["resolved_at"] is None
        assert result["acknowledged_at"] is None


# ============================================================================
# Dashboard Tests
# ============================================================================


class TestDashboardWidget:
    """Tests for DashboardWidget dataclass"""

    def test_widget_creation(self):
        """Test DashboardWidget instantiation"""
        widget = DashboardWidget(
            widget_id="w1",
            widget_type="chart",
            title="CPU Chart",
            position={"x": 0, "y": 0, "width": 6, "height": 3},
            config={"chart_type": "line"},
            data_source="metrics_api",
            refresh_interval_seconds=60,
            metrics=["cpu_usage"],
            filters={"host": "server1"},
        )

        assert widget.widget_id == "w1"
        assert widget.widget_type == "chart"
        assert widget.title == "CPU Chart"
        assert widget.refresh_interval_seconds == 60


class TestDashboard:
    """Tests for Dashboard dataclass"""

    def test_dashboard_creation(self):
        """Test Dashboard instantiation"""
        now = datetime.now()
        dashboard = Dashboard(
            dashboard_id="dash1",
            name="System Overview",
            description="System health dashboard",
            dashboard_type=DashboardType.SYSTEM_OVERVIEW,
            owner="admin",
            tenant_id="tenant1",
            is_public=True,
            created_at=now,
            updated_at=now,
        )

        assert dashboard.dashboard_id == "dash1"
        assert dashboard.name == "System Overview"
        assert dashboard.is_public is True
        assert dashboard.owner == "admin"

    def test_dashboard_to_dict(self):
        """Test Dashboard serialization"""
        now = datetime.now()
        widget = DashboardWidget(
            widget_id="w1",
            widget_type="metric",
            title="Health",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
        )

        dashboard = Dashboard(
            dashboard_id="dash2",
            name="Test Dashboard",
            description="Test",
            dashboard_type=DashboardType.PERFORMANCE_ANALYTICS,
            widgets=[widget],
            owner="user1",
            created_at=now,
            updated_at=now,
        )

        result = dashboard.to_dict()

        assert result["dashboard_id"] == "dash2"
        assert result["name"] == "Test Dashboard"
        assert result["dashboard_type"] == "performance_analytics"
        assert len(result["widgets"]) == 1
        assert result["created_at"] == now.isoformat()


# ============================================================================
# MetricsCollector Tests
# ============================================================================


class TestMetricsCollector:
    """Tests for MetricsCollector"""

    def test_metrics_collector_init(self):
        """Test MetricsCollector initialization"""
        collector = MetricsCollector(buffer_size=50000, retention_hours=24)

        assert collector.buffer_size == 50000
        assert collector.retention_hours == 24
        assert len(collector.metrics_buffer) == 0
        assert len(collector.aggregated_metrics) == 0

    def test_add_metric(self):
        """Test adding a single metric"""
        collector = MetricsCollector()
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=75.0,
            component="system",
        )

        collector.add_metric(metric)

        key = "cpu_usage:system"
        assert key in collector.metrics_buffer
        assert len(collector.metrics_buffer[key]) == 1
        assert collector.metrics_buffer[key][0] == metric

    def test_add_metric_aggregation(self):
        """Test metric aggregation"""
        collector = MetricsCollector()

        for i in range(10):
            metric = MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=float(50 + i),
                component="system",
            )
            collector.add_metric(metric)

        key = "cpu_usage:system"
        agg = collector.aggregated_metrics[key]

        assert agg["count"] == 10
        assert agg["min"] == 50.0
        assert agg["max"] == 59.0
        assert agg["sum"] == 545.0

    def test_add_metric_tenant_separation(self):
        """Test metrics are properly separated by tenant"""
        collector = MetricsCollector()

        metric1 = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
            component="system",
            tenant_id="tenant1",
        )

        metric2 = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=60.0,
            component="system",
            tenant_id="tenant2",
        )

        collector.add_metric(metric1)
        collector.add_metric(metric2)

        assert "tenant1" in collector.tenant_metrics
        assert "tenant2" in collector.tenant_metrics

    def test_get_metrics_no_filters(self):
        """Test getting metrics without filters"""
        collector = MetricsCollector()

        for i in range(5):
            metric = MetricData(
                timestamp=datetime.now() - timedelta(minutes=i),
                metric_type=MetricType.MEMORY_USAGE,
                value=50.0 + i,
                component="system",
            )
            collector.add_metric(metric)

        metrics = collector.get_metrics()

        assert len(metrics) == 5
        # Should be sorted by timestamp (newest first)
        assert metrics[0].value == 50.0  # Most recent
        assert metrics[-1].value == 54.0  # Oldest

    def test_get_metrics_by_type(self):
        """Test filtering metrics by type"""
        collector = MetricsCollector()

        cpu_metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
            component="system",
        )

        memory_metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.MEMORY_USAGE,
            value=60.0,
            component="system",
        )

        collector.add_metric(cpu_metric)
        collector.add_metric(memory_metric)

        cpu_metrics = collector.get_metrics(metric_type=MetricType.CPU_USAGE)

        assert len(cpu_metrics) == 1
        assert cpu_metrics[0].metric_type == MetricType.CPU_USAGE

    def test_get_metrics_by_component(self):
        """Test filtering metrics by component"""
        collector = MetricsCollector()

        metric1 = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
            component="system",
        )

        metric2 = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=60.0,
            component="api",
        )

        collector.add_metric(metric1)
        collector.add_metric(metric2)

        system_metrics = collector.get_metrics(component="system")

        assert len(system_metrics) == 1
        assert system_metrics[0].component == "system"

    def test_get_metrics_by_time_range(self):
        """Test filtering metrics by time range"""
        collector = MetricsCollector()
        now = datetime.now()

        metric1 = MetricData(
            timestamp=now - timedelta(hours=2),
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
            component="system",
        )

        metric2 = MetricData(
            timestamp=now - timedelta(minutes=30),
            metric_type=MetricType.CPU_USAGE,
            value=60.0,
            component="system",
        )

        collector.add_metric(metric1)
        collector.add_metric(metric2)

        # Get metrics from last hour
        start_time = now - timedelta(hours=1)
        recent_metrics = collector.get_metrics(start_time=start_time)

        assert len(recent_metrics) == 1
        assert recent_metrics[0].value == 60.0

    def test_get_metrics_by_tags(self):
        """Test filtering metrics by tags"""
        collector = MetricsCollector()

        metric1 = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
            component="system",
            tags={"host": "server1", "env": "prod"},
        )

        metric2 = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=60.0,
            component="system",
            tags={"host": "server2", "env": "prod"},
        )

        collector.add_metric(metric1)
        collector.add_metric(metric2)

        server1_metrics = collector.get_metrics(tags={"host": "server1"})

        assert len(server1_metrics) == 1
        assert server1_metrics[0].tags["host"] == "server1"

    def test_get_metrics_with_limit(self):
        """Test limiting metrics results"""
        collector = MetricsCollector()

        for i in range(10):
            metric = MetricData(
                timestamp=datetime.now() - timedelta(seconds=i),
                metric_type=MetricType.CPU_USAGE,
                value=50.0 + i,
                component="system",
            )
            collector.add_metric(metric)

        limited_metrics = collector.get_metrics(limit=3)

        assert len(limited_metrics) == 3

    def test_get_statistics_basic(self):
        """Test getting statistics for metrics"""
        collector = MetricsCollector()

        values = [50.0, 60.0, 70.0, 80.0, 90.0]
        for val in values:
            metric = MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=val,
                component="system",
            )
            collector.add_metric(metric)

        stats = collector.get_statistics(MetricType.CPU_USAGE, component="system")

        assert stats["count"] == 5
        assert stats["min"] == 50.0
        assert stats["max"] == 90.0
        assert stats["average"] == 70.0

    def test_get_statistics_with_single_value(self):
        """Test statistics with single value (no stdev)"""
        collector = MetricsCollector()

        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
            component="system",
        )
        collector.add_metric(metric)

        stats = collector.get_statistics(MetricType.CPU_USAGE, component="system")

        assert stats["count"] == 1
        assert stats["std_dev"] == 0

    def test_get_statistics_no_metrics(self):
        """Test statistics when no metrics exist"""
        collector = MetricsCollector()

        stats = collector.get_statistics(MetricType.CPU_USAGE)

        assert stats["count"] == 0
        assert stats["average"] is None
        assert stats["min"] is None
        assert stats["max"] is None

    def test_get_statistics_tenant_specific(self):
        """Test getting tenant-specific statistics"""
        collector = MetricsCollector()

        for i in range(5):
            metric = MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=50.0 + i,
                component="system",
                tenant_id="tenant1",
            )
            collector.add_metric(metric)

        stats = collector.get_statistics(
            MetricType.CPU_USAGE, component="system", tenant_id="tenant1"
        )

        assert stats["count"] == 5

    def test_percentile_calculation(self):
        """Test percentile calculation"""
        collector = MetricsCollector()

        values = list(range(1, 101))  # 1 to 100
        for val in values:
            metric = MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=float(val),
                component="system",
            )
            collector.add_metric(metric)

        stats = collector.get_statistics(MetricType.CPU_USAGE, component="system")

        assert "p95" in stats
        assert "p99" in stats
        assert stats["p95"] >= 95.0  # Approximate
        assert stats["p99"] >= 99.0  # Approximate

    def test_cleanup_old_metrics(self):
        """Test old metrics cleanup"""
        collector = MetricsCollector(retention_hours=1)

        now = datetime.now()
        old_metric = MetricData(
            timestamp=now - timedelta(hours=2),
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
            component="system",
        )

        recent_metric = MetricData(
            timestamp=now,
            metric_type=MetricType.CPU_USAGE,
            value=60.0,
            component="system",
        )

        collector.add_metric(old_metric)
        collector.add_metric(recent_metric)

        # Manually trigger cleanup by setting last_cleanup to old time
        collector._last_cleanup = now - timedelta(minutes=10)
        collector._cleanup_old_metrics()

        # Old metric should be removed
        key = "cpu_usage:system"
        metrics = list(collector.metrics_buffer[key])
        assert len(metrics) <= 1

    def test_thread_safety(self):
        """Test thread safety of MetricsCollector"""
        collector = MetricsCollector()

        def add_metrics():
            for i in range(100):
                metric = MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.CPU_USAGE,
                    value=50.0 + i,
                    component="system",
                )
                collector.add_metric(metric)

        threads = [threading.Thread(target=add_metrics) for _ in range(5)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        # Should have metrics from all threads
        assert len(collector.metrics_buffer) > 0


# ============================================================================
# AlertManager Tests
# ============================================================================


class TestAlertManager:
    """Tests for AlertManager"""

    def test_alert_manager_init(self):
        """Test AlertManager initialization"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager.metrics_collector == collector
        assert len(manager.alert_rules) == 0
        assert len(manager.active_alerts) == 0
        assert len(manager.alert_history) == 0

    def test_add_alert_rule(self):
        """Test adding alert rule"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
            severity=AlertSeverity.ERROR,
        )

        assert len(manager.alert_rules) == 1
        rule = manager.alert_rules[0]
        assert rule["name"] == "High CPU"
        assert rule["threshold"] == 80.0
        assert rule["enabled"] is True

    def test_add_alert_rule_with_tenant(self):
        """Test adding tenant-specific alert rule"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="Tenant CPU Alert",
            metric_type=MetricType.CPU_USAGE,
            threshold=75.0,
            tenant_id="tenant1",
        )

        rule = manager.alert_rules[0]
        assert rule["tenant_id"] == "tenant1"

    def test_evaluate_condition_gt(self):
        """Test greater-than condition evaluation"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(85.0, 80.0, "gt") is True
        assert manager._evaluate_condition(75.0, 80.0, "gt") is False
        assert manager._evaluate_condition(80.0, 80.0, "gt") is False

    def test_evaluate_condition_lt(self):
        """Test less-than condition evaluation"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(75.0, 80.0, "lt") is True
        assert manager._evaluate_condition(85.0, 80.0, "lt") is False

    def test_evaluate_condition_eq(self):
        """Test equals condition evaluation"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(80.0, 80.0, "eq") is True
        assert manager._evaluate_condition(81.0, 80.0, "eq") is False

    def test_evaluate_condition_ne(self):
        """Test not-equals condition evaluation"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(81.0, 80.0, "ne") is True
        assert manager._evaluate_condition(80.0, 80.0, "ne") is False

    def test_evaluate_condition_gte(self):
        """Test greater-than-or-equal condition"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(85.0, 80.0, "gte") is True
        assert manager._evaluate_condition(80.0, 80.0, "gte") is True
        assert manager._evaluate_condition(75.0, 80.0, "gte") is False

    def test_evaluate_condition_lte(self):
        """Test less-than-or-equal condition"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(75.0, 80.0, "lte") is True
        assert manager._evaluate_condition(80.0, 80.0, "lte") is True
        assert manager._evaluate_condition(85.0, 80.0, "lte") is False

    def test_evaluate_condition_invalid(self):
        """Test invalid operator"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(85.0, 80.0, "invalid") is False

    def test_check_alerts_no_metrics(self):
        """Test alert checking with no metrics"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
        )

        alerts = manager.check_alerts()

        assert len(alerts) == 0

    def test_check_alerts_violation(self):
        """Test alert triggering on violation"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
            severity=AlertSeverity.ERROR,
        )

        # Add violating metric
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=85.0,
            component="system",
        )
        collector.add_metric(metric)

        alerts = manager.check_alerts()

        assert len(alerts) == 1
        alert = alerts[0]
        assert alert.severity == AlertSeverity.ERROR
        assert alert.current_value == 85.0

    def test_check_alerts_no_violation(self):
        """Test no alert when condition not met"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
        )

        # Add non-violating metric
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=70.0,
            component="system",
        )
        collector.add_metric(metric)

        alerts = manager.check_alerts()

        assert len(alerts) == 0

    def test_check_alerts_consecutive_violations(self):
        """Test consecutive violations requirement"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
            consecutive_violations=3,
        )

        # Add 2 violating metrics
        for _ in range(2):
            metric = MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=85.0,
                component="system",
            )
            collector.add_metric(metric)

        alerts = manager.check_alerts()

        # Should not trigger yet (need 3)
        assert len(alerts) == 0

    def test_check_alerts_cooldown(self):
        """Test alert cooldown period"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
            cooldown_minutes=1,
        )

        # Add violating metric
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=85.0,
            component="system",
        )
        collector.add_metric(metric)

        alerts1 = manager.check_alerts()
        assert len(alerts1) == 1

        # Immediately check again - should not trigger due to cooldown
        alerts2 = manager.check_alerts()
        assert len(alerts2) == 0

    def test_check_alerts_disabled_rule(self):
        """Test disabled rules are not checked"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
        )

        # Disable rule
        manager.alert_rules[0]["enabled"] = False

        # Add violating metric
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=85.0,
            component="system",
        )
        collector.add_metric(metric)

        alerts = manager.check_alerts()

        assert len(alerts) == 0

    def test_add_alert_callback(self):
        """Test adding alert callback"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        callback_mock = Mock()
        manager.add_alert_callback(callback_mock)

        assert callback_mock in manager.alert_callbacks

    def test_alert_callback_execution(self):
        """Test callback is executed on alert"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        callback_mock = Mock()
        manager.add_alert_callback(callback_mock)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
        )

        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=85.0,
            component="system",
        )
        collector.add_metric(metric)

        manager.check_alerts()

        assert callback_mock.called
        assert len(callback_mock.call_args_list) == 1

    def test_acknowledge_alert(self):
        """Test acknowledging an alert"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
        )

        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=85.0,
            component="system",
        )
        collector.add_metric(metric)

        alerts = manager.check_alerts()
        alert_id = alerts[0].alert_id

        result = manager.acknowledge_alert(alert_id)

        assert result is True
        assert manager.active_alerts[alert_id].acknowledged is True

    def test_acknowledge_nonexistent_alert(self):
        """Test acknowledging nonexistent alert returns False"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        result = manager.acknowledge_alert("nonexistent")

        assert result is False

    def test_get_active_alerts(self):
        """Test getting active alerts"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
            severity=AlertSeverity.ERROR,
        )

        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=85.0,
            component="system",
        )
        collector.add_metric(metric)

        manager.check_alerts()

        alerts = manager.get_active_alerts()

        assert len(alerts) == 1

    def test_get_active_alerts_by_severity(self):
        """Test filtering active alerts by severity"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
            severity=AlertSeverity.ERROR,
        )

        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=85.0,
            component="system",
        )
        collector.add_metric(metric)

        manager.check_alerts()

        alerts = manager.get_active_alerts(severity=AlertSeverity.ERROR)

        assert len(alerts) == 1
        assert alerts[0].severity == AlertSeverity.ERROR

    def test_get_alert_history(self):
        """Test getting alert history"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
        )

        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=85.0,
            component="system",
        )
        collector.add_metric(metric)

        manager.check_alerts()

        history = manager.get_alert_history(hours=1)

        assert len(history) > 0

    def test_get_alert_statistics(self):
        """Test getting alert statistics"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
            severity=AlertSeverity.ERROR,
        )

        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=85.0,
            component="system",
        )
        collector.add_metric(metric)

        manager.check_alerts()

        stats = manager.get_alert_statistics(hours=1)

        assert "total_alerts" in stats
        assert "by_severity" in stats
        assert "by_component" in stats


# ============================================================================
# DashboardManager Tests
# ============================================================================


class TestDashboardManager:
    """Tests for DashboardManager"""

    def test_dashboard_manager_init(self):
        """Test DashboardManager initialization"""
        manager = DashboardManager()

        assert len(manager.dashboards) == 0
        assert len(manager.default_dashboards) == 2  # System overview and hook analysis

    def test_create_dashboard(self):
        """Test creating a dashboard"""
        manager = DashboardManager()

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="metric",
            title="CPU",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
        )

        dashboard_id = manager.create_dashboard(
            name="Test Dashboard",
            description="Test",
            dashboard_type=DashboardType.PERFORMANCE_ANALYTICS,
            widgets=[widget],
            owner="user1",
        )

        assert dashboard_id is not None
        assert dashboard_id in manager.dashboards

    def test_get_dashboard(self):
        """Test retrieving a dashboard"""
        manager = DashboardManager()

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="metric",
            title="CPU",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
        )

        dashboard_id = manager.create_dashboard(
            name="Test",
            description="Test",
            dashboard_type=DashboardType.SYSTEM_OVERVIEW,
            widgets=[widget],
        )

        dashboard = manager.get_dashboard(dashboard_id)

        assert dashboard is not None
        assert dashboard.name == "Test"

    def test_get_default_dashboard(self):
        """Test retrieving default dashboard"""
        manager = DashboardManager()

        dashboard = manager.get_dashboard("system_overview")

        assert dashboard is not None
        assert dashboard.name == "System Overview"

    def test_list_dashboards(self):
        """Test listing dashboards"""
        manager = DashboardManager()

        # Create a dashboard first
        widget = DashboardWidget(
            widget_id="test_widget",
            widget_type="metric",
            title="Test",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
            config={},
        )
        manager.create_dashboard(
            name="Test Dashboard",
            description="Test",
            dashboard_type=DashboardType.SYSTEM_OVERVIEW,
            widgets=[widget],
        )

        dashboards = manager.list_dashboards()

        # Should include created dashboard
        assert len(dashboards) > 0

    def test_list_dashboards_by_type(self):
        """Test filtering dashboards by type"""
        manager = DashboardManager()

        dashboards = manager.list_dashboards(dashboard_type=DashboardType.SYSTEM_OVERVIEW)

        assert all(d.dashboard_type == DashboardType.SYSTEM_OVERVIEW for d in dashboards)

    def test_update_dashboard(self):
        """Test updating a dashboard"""
        manager = DashboardManager()

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="metric",
            title="CPU",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
        )

        dashboard_id = manager.create_dashboard(
            name="Test",
            description="Test",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[widget],
        )

        result = manager.update_dashboard(dashboard_id, {"name": "Updated"})

        assert result is True
        updated_dashboard = manager.get_dashboard(dashboard_id)
        assert updated_dashboard.name == "Updated"

    def test_delete_dashboard(self):
        """Test deleting a dashboard"""
        manager = DashboardManager()

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="metric",
            title="CPU",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
        )

        dashboard_id = manager.create_dashboard(
            name="Test",
            description="Test",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[widget],
        )

        result = manager.delete_dashboard(dashboard_id)

        assert result is True
        assert manager.get_dashboard(dashboard_id) is None

    def test_delete_default_dashboard_fails(self):
        """Test that default dashboards cannot be deleted"""
        manager = DashboardManager()

        result = manager.delete_dashboard("system_overview")

        assert result is False


# ============================================================================
# RealtimeMonitoringDashboard Tests
# ============================================================================


class TestRealtimeMonitoringDashboard:
    """Tests for RealtimeMonitoringDashboard"""

    def test_dashboard_init(self):
        """Test RealtimeMonitoringDashboard initialization"""
        dashboard = RealtimeMonitoringDashboard(
            metrics_buffer_size=50000,
            retention_hours=24,
            alert_check_interval=60,
        )

        assert dashboard.metrics_buffer_size == 50000
        assert dashboard.retention_hours == 24
        assert dashboard._running is False

    def test_dashboard_start_stop(self):
        """Test starting and stopping the dashboard"""

        async def test_start_stop():
            dashboard = RealtimeMonitoringDashboard(enable_websocket=False, enable_external_integration=False)

            await dashboard.start()
            assert dashboard._running is True

            dashboard.stop()
            assert dashboard._running is False

        asyncio.run(test_start_stop())

    def test_add_metric(self):
        """Test adding a metric"""
        dashboard = RealtimeMonitoringDashboard()

        dashboard.add_metric(
            MetricType.CPU_USAGE,
            75.5,
            tags={"host": "server1"},
            source="test",
            component="system",
        )

        metrics = dashboard.metrics_collector.get_metrics(metric_type=MetricType.CPU_USAGE)

        assert len(metrics) == 1
        assert metrics[0].value == 75.5

    def test_get_dashboard_data(self):
        """Test getting dashboard data"""
        dashboard = RealtimeMonitoringDashboard()

        # Add some metrics
        for i in range(5):
            dashboard.add_metric(
                MetricType.CPU_USAGE,
                50.0 + i,
                component="system",
            )

        data = dashboard.get_dashboard_data("system_overview")

        assert "dashboard" in data
        assert "widgets_data" in data
        assert data["dashboard"]["name"] == "System Overview"

    def test_get_dashboard_data_not_found(self):
        """Test getting nonexistent dashboard"""
        dashboard = RealtimeMonitoringDashboard()

        data = dashboard.get_dashboard_data("nonexistent")

        assert "error" in data

    def test_get_system_status(self):
        """Test getting system status"""
        dashboard = RealtimeMonitoringDashboard()

        status = dashboard.get_system_status()

        assert "status" in status
        assert "uptime_seconds" in status
        assert "metrics_collected" in status
        assert "active_alerts" in status
        assert "total_dashboards" in status

    def test_create_custom_dashboard(self):
        """Test creating a custom dashboard"""
        dashboard = RealtimeMonitoringDashboard()

        widget_defs = [
            {
                "widget_id": "w1",
                "widget_type": "metric",
                "title": "CPU",
                "position": {"x": 0, "y": 0, "width": 4, "height": 2},
                "config": {"metric": "cpu_usage"},
            }
        ]

        dashboard_id = dashboard.create_custom_dashboard(
            name="Custom",
            description="Custom dashboard",
            widgets=widget_defs,
            owner="user1",
        )

        assert dashboard_id is not None
        created = dashboard.dashboard_manager.get_dashboard(dashboard_id)
        assert created.name == "Custom"

    def test_map_metric_name(self):
        """Test metric name mapping"""
        dashboard = RealtimeMonitoringDashboard()

        metric_type = dashboard._map_metric_name("cpu_usage")
        assert metric_type == MetricType.CPU_USAGE

        metric_type = dashboard._map_metric_name("memory_usage")
        assert metric_type == MetricType.MEMORY_USAGE

        metric_type = dashboard._map_metric_name("unknown")
        assert metric_type is None

    @patch("psutil.cpu_percent")
    @patch("psutil.virtual_memory")
    @patch("psutil.Process")
    def test_collect_system_metrics(self, mock_process, mock_memory, mock_cpu):
        """Test system metrics collection"""
        mock_cpu.return_value = 45.5

        mock_mem_obj = MagicMock()
        mock_mem_obj.percent = 60.5
        mock_memory.return_value = mock_mem_obj

        mock_proc_obj = MagicMock()
        mock_proc_obj.memory_info.return_value = MagicMock(rss=1024 * 1024 * 500)  # 500MB
        mock_process.return_value = mock_proc_obj

        dashboard = RealtimeMonitoringDashboard()
        dashboard._collect_system_metrics()

        # Verify metrics were collected
        cpu_metrics = dashboard.metrics_collector.get_metrics(metric_type=MetricType.CPU_USAGE)
        assert len(cpu_metrics) > 0

    def test_get_widget_data_metric_type(self):
        """Test getting metric widget data"""
        dashboard = RealtimeMonitoringDashboard()

        # Add metrics
        dashboard.add_metric(MetricType.CPU_USAGE, 75.5, component="system")

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="metric",
            title="CPU",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
            config={"metric": "cpu_usage", "format": "percentage"},
        )

        data = dashboard._get_metric_widget_data(widget)

        assert "value" in data
        assert "formatted_value" in data

    def test_get_widget_data_chart_type(self):
        """Test getting chart widget data"""
        dashboard = RealtimeMonitoringDashboard()

        # Add metrics
        for i in range(5):
            dashboard.add_metric(MetricType.CPU_USAGE, 50.0 + i, component="system")

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="chart",
            title="CPU Chart",
            position={"x": 0, "y": 0, "width": 6, "height": 3},
            config={"chart_type": "line"},
            metrics=["cpu_usage"],
        )

        start_time = datetime.now() - timedelta(hours=1)
        end_time = datetime.now()
        data = dashboard._get_chart_widget_data(widget, start_time, end_time)

        assert "type" in data
        assert "data" in data

    def test_get_widget_data_alert_type(self):
        """Test getting alert widget data"""
        dashboard = RealtimeMonitoringDashboard()

        # Trigger an alert
        dashboard.alert_manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
        )

        dashboard.add_metric(MetricType.CPU_USAGE, 85.0, component="system")
        dashboard.alert_manager.check_alerts()

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="alert",
            title="Alerts",
            position={"x": 0, "y": 0, "width": 12, "height": 3},
        )

        data = dashboard._get_alert_widget_data(widget)

        assert "active_count" in data
        assert "severity_breakdown" in data

    def test_default_alert_rules(self):
        """Test that default alert rules are created"""
        dashboard = RealtimeMonitoringDashboard()

        # Should have default rules
        assert len(dashboard.alert_manager.alert_rules) >= 4

        rule_names = [r["name"] for r in dashboard.alert_manager.alert_rules]
        assert "High CPU Usage" in rule_names
        assert "High Memory Usage" in rule_names

    def test_thread_safety_add_metric(self):
        """Test thread safety of metric addition"""
        dashboard = RealtimeMonitoringDashboard()

        def add_metrics():
            for i in range(100):
                dashboard.add_metric(
                    MetricType.CPU_USAGE,
                    50.0 + i,
                    component="system",
                )

        threads = [threading.Thread(target=add_metrics) for _ in range(3)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        metrics = dashboard.metrics_collector.get_metrics()
        assert len(metrics) > 0

    def test_broadcast_alert(self):
        """Test alert broadcasting"""
        dashboard = RealtimeMonitoringDashboard()

        alert = Alert(
            alert_id="test_alert",
            severity=AlertSeverity.ERROR,
            title="Test",
            description="Test alert",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="test",
            component="system",
        )

        # Should not raise exception
        dashboard._broadcast_alert(alert)


# ============================================================================
# Global Function Tests
# ============================================================================


class TestGlobalFunctions:
    """Tests for module-level convenience functions"""

    def test_get_monitoring_dashboard_singleton(self):
        """Test get_monitoring_dashboard returns singleton"""
        dashboard1 = get_monitoring_dashboard()
        dashboard2 = get_monitoring_dashboard()

        assert dashboard1 is dashboard2

    @pytest.mark.asyncio
    async def test_start_monitoring(self):
        """Test start_monitoring function"""
        # Get fresh dashboard by resetting global
        import src.moai_adk.core.realtime_monitoring_dashboard as mod

        mod._monitoring_dashboard = None

        dashboard = get_monitoring_dashboard(enable_websocket=False, enable_external_integration=False)

        await start_monitoring()

        assert dashboard._running is True

        dashboard.stop()

    def test_stop_monitoring(self):
        """Test stop_monitoring function"""
        dashboard = get_monitoring_dashboard(enable_websocket=False, enable_external_integration=False)
        dashboard._running = True

        stop_monitoring()

        assert dashboard._running is False

    def test_add_system_metric(self):
        """Test add_system_metric convenience function"""
        import src.moai_adk.core.realtime_monitoring_dashboard as mod

        mod._monitoring_dashboard = None
        dashboard = get_monitoring_dashboard()

        add_system_metric(
            MetricType.CPU_USAGE,
            75.5,
            component="system",
            tags={"host": "test"},
        )

        metrics = dashboard.metrics_collector.get_metrics(metric_type=MetricType.CPU_USAGE)
        assert len(metrics) > 0

    def test_add_hook_metric(self):
        """Test add_hook_metric convenience function"""
        import src.moai_adk.core.realtime_monitoring_dashboard as mod

        mod._monitoring_dashboard = None
        dashboard = get_monitoring_dashboard()

        add_hook_metric(
            hook_path="test_hook.py",
            execution_time_ms=150.5,
            success=True,
        )

        metrics = dashboard.metrics_collector.get_metrics(metric_type=MetricType.RESPONSE_TIME)
        assert len(metrics) > 0


# ============================================================================
# Edge Cases and Error Handling
# ============================================================================


class TestEdgeCases:
    """Tests for edge cases and error handling"""

    def test_metrics_buffer_overflow(self):
        """Test metrics buffer respects max size"""
        collector = MetricsCollector(buffer_size=10)

        for i in range(20):
            metric = MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=50.0 + i,
                component="system",
            )
            collector.add_metric(metric)

        key = "cpu_usage:system"
        # Buffer should not exceed max size
        assert len(collector.metrics_buffer[key]) <= 10

    def test_empty_metrics_values_list(self):
        """Test statistics with empty values list"""
        collector = MetricsCollector()

        # Add metric with string value
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.SYSTEM_PERFORMANCE,
            value="healthy",
            component="system",
        )
        collector.add_metric(metric)

        stats = collector.get_statistics(MetricType.SYSTEM_PERFORMANCE, component="system")

        assert stats["count"] == 0

    def test_get_widget_data_unsupported_type(self):
        """Test widget data for unsupported widget type"""
        dashboard = RealtimeMonitoringDashboard()

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="unknown",
            title="Unknown",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
        )

        data = dashboard._get_widget_data(widget, datetime.now(), datetime.now())

        assert "error" in data

    def test_check_resolved_alerts(self):
        """Test alert resolution checking"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
        )

        # Trigger alert with high value
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=85.0,
            component="system",
        )
        collector.add_metric(metric)

        alerts = manager.check_alerts()
        assert len(alerts) == 1
        assert len(manager.active_alerts) == 1

    def test_percentile_empty_values(self):
        """Test percentile with empty values"""
        collector = MetricsCollector()

        result = collector._percentile([], 95)

        assert result == 0.0

    def test_percentile_single_value(self):
        """Test percentile with single value"""
        collector = MetricsCollector()

        result = collector._percentile([50.0], 95)

        assert result == 50.0

    def test_alert_without_matching_rule(self):
        """Test checking resolved alerts without matching rule"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        # Manually create alert without a rule
        alert = Alert(
            alert_id="orphan_alert",
            severity=AlertSeverity.WARNING,
            title="Orphan",
            description="Alert without rule",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="test",
            component="system",
        )

        manager.active_alerts["orphan_alert"] = alert

        # Should not crash
        manager._check_resolved_alerts()

    def test_dashboard_get_invalid_metric_widget(self):
        """Test getting metric widget with invalid metric name"""
        dashboard = RealtimeMonitoringDashboard()

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="metric",
            title="Invalid",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
            config={"metric": "invalid_metric"},
        )

        data = dashboard._get_metric_widget_data(widget)

        assert "error" in data

    def test_dashboard_get_table_widget_unknown_type(self):
        """Test getting table widget of unknown type"""
        dashboard = RealtimeMonitoringDashboard()

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="table",
            title="Unknown Table",
            position={"x": 0, "y": 0, "width": 12, "height": 3},
        )

        data = dashboard._get_table_widget_data(widget, datetime.now(), datetime.now())

        assert "error" in data

    def test_concurrent_alert_rule_modifications(self):
        """Test thread safety of alert rule modifications"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        def add_rules():
            for i in range(10):
                manager.add_alert_rule(
                    name=f"Rule {i}",
                    metric_type=MetricType.CPU_USAGE,
                    threshold=80.0,
                )

        threads = [threading.Thread(target=add_rules) for _ in range(3)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        # All rules should be added
        assert len(manager.alert_rules) >= 30

    def test_metrics_filtering_multiple_conditions(self):
        """Test metrics filtering with multiple conditions"""
        collector = MetricsCollector()
        now = datetime.now()

        # Add various metrics
        for i in range(10):
            metric = MetricData(
                timestamp=now - timedelta(seconds=i),
                metric_type=MetricType.CPU_USAGE,
                value=50.0 + i,
                component="system" if i % 2 == 0 else "app",
                tags={"env": "prod" if i % 3 == 0 else "test"},
            )
            collector.add_metric(metric)

        # Filter with multiple conditions
        filtered = collector.get_metrics(
            metric_type=MetricType.CPU_USAGE,
            component="system",
            tags={"env": "prod"},
            limit=5,
        )

        # Should return filtered results
        assert len(filtered) > 0
        assert all(m.component == "system" for m in filtered)

    def test_statistics_calculation_error_handling(self):
        """Test statistics handles calculation errors gracefully"""
        collector = MetricsCollector()

        # Add metrics with standard deviation
        for val in [10.0, 20.0, 30.0, 40.0, 50.0]:
            metric = MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=val,
                component="system",
            )
            collector.add_metric(metric)

        stats = collector.get_statistics(MetricType.CPU_USAGE, component="system")

        # Should include std_dev
        assert "std_dev" in stats
        assert stats["std_dev"] > 0
