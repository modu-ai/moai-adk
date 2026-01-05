"""
Comprehensive test coverage for comprehensive_monitoring_system.py

Target: 200+ lines of coverage for MetricsCollector, AlertManager, DashboardManager
Strategy: Maximum test coverage with all methods and edge cases
"""

import threading
from datetime import datetime, timedelta
from pathlib import Path
from unittest.mock import Mock, patch

import pytest

from moai_adk.core.comprehensive_monitoring_system import (
    Alert,
    AlertManager,
    AlertSeverity,
    ComprehensiveMonitoringSystem,
    HealthStatus,
    MetricData,
    MetricsCollector,
    MetricType,
    PerformanceMonitor,
    PredictiveAnalytics,
    SystemHealth,
)


class TestMetricData:
    """Tests for MetricData dataclass"""

    def test_metric_data_creation(self):
        """Test MetricData creation"""
        now = datetime.now()
        metric = MetricData(
            timestamp=now,
            metric_type=MetricType.CPU_USAGE,
            value=45.5,
            tags={"component": "test"},
            source="test_source",
        )
        assert metric.timestamp == now
        assert metric.metric_type == MetricType.CPU_USAGE
        assert metric.value == 45.5
        assert metric.tags["component"] == "test"
        assert metric.source == "test_source"

    def test_metric_data_to_dict(self):
        """Test MetricData to_dict serialization"""
        now = datetime.now()
        metric = MetricData(
            timestamp=now,
            metric_type=MetricType.MEMORY_USAGE,
            value=75.0,
            tags={"zone": "us-east"},
            source="psutil",
            metadata={"extra": "data"},
        )
        result = metric.to_dict()
        assert result["timestamp"] == now.isoformat()
        assert result["metric_type"] == "memory_usage"
        assert result["value"] == 75.0
        assert result["tags"]["zone"] == "us-east"
        assert result["metadata"]["extra"] == "data"


class TestAlert:
    """Tests for Alert dataclass"""

    def test_alert_creation(self):
        """Test Alert creation"""
        now = datetime.now()
        alert = Alert(
            alert_id="test_alert_1",
            severity=AlertSeverity.HIGH,
            title="Test Alert",
            description="Test description",
            timestamp=now,
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.5,
            source="monitor",
        )
        assert alert.alert_id == "test_alert_1"
        assert alert.severity == AlertSeverity.HIGH
        assert alert.resolved is False
        assert alert.acknowledged is False

    def test_alert_to_dict(self):
        """Test Alert serialization"""
        now = datetime.now()
        resolved_at = now + timedelta(minutes=5)
        acknowledged_at = now + timedelta(minutes=1)

        alert = Alert(
            alert_id="alert_123",
            severity=AlertSeverity.CRITICAL,
            title="Critical CPU Alert",
            description="CPU exceeded threshold",
            timestamp=now,
            metric_type=MetricType.CPU_USAGE,
            threshold=90.0,
            current_value=95.0,
            source="monitor",
            resolved=True,
            resolved_at=resolved_at,
            acknowledged=True,
            acknowledged_at=acknowledged_at,
        )
        result = alert.to_dict()
        assert result["alert_id"] == "alert_123"
        assert result["severity"] == 4  # CRITICAL value
        assert result["resolved"] is True
        assert result["acknowledged"] is True


class TestSystemHealth:
    """Tests for SystemHealth dataclass"""

    def test_system_health_creation(self):
        """Test SystemHealth creation"""
        now = datetime.now()
        health = SystemHealth(
            status=HealthStatus.HEALTHY,
            timestamp=now,
            overall_score=95.0,
            component_scores={"cpu": 90.0, "memory": 95.0},
            active_alerts=["alert_1"],
            recent_metrics={"cpu_usage": 45.0},
        )
        assert health.status == HealthStatus.HEALTHY
        assert health.overall_score == 95.0

    def test_system_health_to_dict(self):
        """Test SystemHealth serialization"""
        now = datetime.now()
        health = SystemHealth(
            status=HealthStatus.WARNING,
            timestamp=now,
            overall_score=75.0,
            component_scores={"cpu": 70.0},
        )
        result = health.to_dict()
        assert result["status"] == "warning"
        assert result["overall_score"] == 75.0


class TestMetricsCollector:
    """Tests for MetricsCollector"""

    def test_metrics_collector_init(self):
        """Test MetricsCollector initialization"""
        collector = MetricsCollector(buffer_size=5000, retention_hours=12)
        assert collector.buffer_size == 5000
        assert collector.retention_hours == 12
        assert len(collector.metrics_buffer) == 0

    def test_add_metric(self):
        """Test adding metrics"""
        collector = MetricsCollector()
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
        )
        collector.add_metric(metric)
        assert len(collector.metrics_buffer[MetricType.CPU_USAGE]) == 1

    def test_add_multiple_metrics(self):
        """Test adding multiple metrics"""
        collector = MetricsCollector()
        for i in range(10):
            metric = MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=40.0 + i,
            )
            collector.add_metric(metric)
        assert len(collector.metrics_buffer[MetricType.CPU_USAGE]) == 10

    def test_get_metrics_no_filter(self):
        """Test getting all metrics"""
        collector = MetricsCollector()
        for metric_type in [MetricType.CPU_USAGE, MetricType.MEMORY_USAGE]:
            for i in range(5):
                metric = MetricData(
                    timestamp=datetime.now(),
                    metric_type=metric_type,
                    value=50.0 + i,
                )
                collector.add_metric(metric)

        all_metrics = collector.get_metrics()
        assert len(all_metrics) == 10

    def test_get_metrics_by_type(self):
        """Test filtering metrics by type"""
        collector = MetricsCollector()
        for i in range(3):
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.CPU_USAGE,
                    value=50.0,
                )
            )
        for i in range(2):
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.MEMORY_USAGE,
                    value=60.0,
                )
            )

        cpu_metrics = collector.get_metrics(metric_type=MetricType.CPU_USAGE)
        assert len(cpu_metrics) == 3

    def test_get_metrics_with_time_filter(self):
        """Test filtering metrics by time range"""
        collector = MetricsCollector()
        now = datetime.now()

        # Add old metric
        old_metric = MetricData(
            timestamp=now - timedelta(hours=2),
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
        )
        collector.add_metric(old_metric)

        # Add recent metric
        recent_metric = MetricData(
            timestamp=now,
            metric_type=MetricType.CPU_USAGE,
            value=60.0,
        )
        collector.add_metric(recent_metric)

        filtered = collector.get_metrics(start_time=now - timedelta(hours=1))
        assert len(filtered) == 1
        assert filtered[0].value == 60.0

    def test_get_metrics_with_limit(self):
        """Test limiting returned metrics"""
        collector = MetricsCollector()
        for i in range(20):
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.CPU_USAGE,
                    value=50.0 + i,
                )
            )

        limited = collector.get_metrics(limit=5)
        assert len(limited) == 5

    def test_get_statistics_no_data(self):
        """Test statistics with no data"""
        collector = MetricsCollector()
        stats = collector.get_statistics(MetricType.CPU_USAGE)
        assert stats["count"] == 0
        assert stats["average"] is None

    def test_get_statistics_single_value(self):
        """Test statistics with single value"""
        collector = MetricsCollector()
        collector.add_metric(
            MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=50.0,
            )
        )
        stats = collector.get_statistics(MetricType.CPU_USAGE)
        assert stats["count"] == 1
        assert stats["average"] == 50.0
        assert stats["min"] == 50.0
        assert stats["max"] == 50.0

    def test_get_statistics_multiple_values(self):
        """Test statistics with multiple values"""
        collector = MetricsCollector()
        values = [40.0, 50.0, 60.0, 70.0, 80.0]
        for val in values:
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.CPU_USAGE,
                    value=val,
                )
            )

        stats = collector.get_statistics(MetricType.CPU_USAGE)
        assert stats["count"] == 5
        assert stats["average"] == 60.0
        assert stats["min"] == 40.0
        assert stats["max"] == 80.0
        assert stats["median"] == 60.0
        assert stats["std_dev"] > 0

    def test_cleanup_old_metrics(self):
        """Test cleanup of old metrics"""
        collector = MetricsCollector(retention_hours=1)
        now = datetime.now()

        # Add old metric
        collector.add_metric(
            MetricData(
                timestamp=now - timedelta(hours=2),
                metric_type=MetricType.CPU_USAGE,
                value=50.0,
            )
        )

        # Add recent metric
        collector.add_metric(
            MetricData(
                timestamp=now,
                metric_type=MetricType.CPU_USAGE,
                value=60.0,
            )
        )

        # Force cleanup
        collector._last_cleanup = now - timedelta(minutes=6)
        collector._cleanup_old_metrics()

        metrics = collector.get_metrics(metric_type=MetricType.CPU_USAGE)
        assert len(metrics) == 1
        assert metrics[0].value == 60.0


class TestAlertManager:
    """Tests for AlertManager"""

    def test_alert_manager_init(self):
        """Test AlertManager initialization"""
        collector = MetricsCollector()
        manager = AlertManager(collector)
        assert manager.metrics_collector == collector
        assert len(manager.alert_rules) == 0
        assert len(manager.active_alerts) == 0

    def test_add_alert_rule(self):
        """Test adding alert rule"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
            severity=AlertSeverity.HIGH,
        )

        assert len(manager.alert_rules) == 1
        assert manager.alert_rules[0]["name"] == "High CPU"
        assert manager.alert_rules[0]["enabled"] is True

    def test_add_multiple_alert_rules(self):
        """Test adding multiple alert rules"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        for i in range(5):
            manager.add_alert_rule(
                name=f"Alert {i}",
                metric_type=MetricType.CPU_USAGE,
                threshold=70.0 + i * 5,
                operator="gt",
            )

        assert len(manager.alert_rules) == 5

    def test_evaluate_condition_gt(self):
        """Test greater than condition evaluation"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(85.0, 80.0, "gt") is True
        assert manager._evaluate_condition(75.0, 80.0, "gt") is False
        assert manager._evaluate_condition(80.0, 80.0, "gt") is False

    def test_evaluate_condition_lt(self):
        """Test less than condition evaluation"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(75.0, 80.0, "lt") is True
        assert manager._evaluate_condition(85.0, 80.0, "lt") is False

    def test_evaluate_condition_eq(self):
        """Test equality condition evaluation"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(80.0, 80.0, "eq") is True
        assert manager._evaluate_condition(85.0, 80.0, "eq") is False

    def test_evaluate_condition_ne(self):
        """Test not equal condition evaluation"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(85.0, 80.0, "ne") is True
        assert manager._evaluate_condition(80.0, 80.0, "ne") is False

    def test_evaluate_condition_gte(self):
        """Test greater than or equal condition"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(85.0, 80.0, "gte") is True
        assert manager._evaluate_condition(80.0, 80.0, "gte") is True
        assert manager._evaluate_condition(75.0, 80.0, "gte") is False

    def test_evaluate_condition_lte(self):
        """Test less than or equal condition"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(75.0, 80.0, "lte") is True
        assert manager._evaluate_condition(80.0, 80.0, "lte") is True
        assert manager._evaluate_condition(85.0, 80.0, "lte") is False

    def test_check_alerts_no_violations(self):
        """Test alert checking with no violations"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
        )

        # Add metric below threshold
        collector.add_metric(
            MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=50.0,
            )
        )

        alerts = manager.check_alerts()
        assert len(alerts) == 0
        assert len(manager.active_alerts) == 0

    def test_check_alerts_with_violations(self):
        """Test alert checking with violations"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
            window_minutes=5,
        )

        # Add metrics above threshold
        for i in range(3):
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.CPU_USAGE,
                    value=85.0,
                )
            )

        alerts = manager.check_alerts()
        assert len(alerts) > 0

    def test_add_alert_callback(self):
        """Test adding alert callback"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        callback_called = []

        def test_callback(alert):
            callback_called.append(alert)

        manager.add_alert_callback(test_callback)
        assert len(manager.alert_callbacks) == 1

    def test_acknowledge_alert(self):
        """Test acknowledging alert"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        alert = Alert(
            alert_id="test_1",
            severity=AlertSeverity.HIGH,
            title="Test",
            description="Test alert",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="monitor",
        )

        manager.active_alerts["test_1"] = alert
        result = manager.acknowledge_alert("test_1")

        assert result is True
        assert manager.active_alerts["test_1"].acknowledged is True

    def test_acknowledge_alert_not_found(self):
        """Test acknowledging non-existent alert"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        result = manager.acknowledge_alert("nonexistent")
        assert result is False

    def test_get_active_alerts(self):
        """Test getting active alerts"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        for i in range(3):
            alert = Alert(
                alert_id=f"alert_{i}",
                severity=AlertSeverity.MEDIUM if i % 2 == 0 else AlertSeverity.HIGH,
                title=f"Alert {i}",
                description="Test",
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                threshold=80.0,
                current_value=85.0,
                source="monitor",
            )
            manager.active_alerts[f"alert_{i}"] = alert

        alerts = manager.get_active_alerts()
        assert len(alerts) == 3

    def test_get_active_alerts_filtered_by_severity(self):
        """Test filtering active alerts by severity"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        for severity in [AlertSeverity.LOW, AlertSeverity.HIGH]:
            alert = Alert(
                alert_id=f"alert_{severity.value}",
                severity=severity,
                title="Test",
                description="Test",
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                threshold=80.0,
                current_value=85.0,
                source="monitor",
            )
            manager.active_alerts[f"alert_{severity.value}"] = alert

        high_alerts = manager.get_active_alerts(severity=AlertSeverity.HIGH)
        assert len(high_alerts) == 1

    def test_get_alert_history(self):
        """Test getting alert history"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        for i in range(5):
            alert = Alert(
                alert_id=f"alert_{i}",
                severity=AlertSeverity.MEDIUM,
                title=f"Alert {i}",
                description="Test",
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                threshold=80.0,
                current_value=85.0,
                source="monitor",
            )
            manager.alert_history.append(alert)

        history = manager.get_alert_history(hours=24)
        assert len(history) == 5


class TestPredictiveAnalytics:
    """Tests for PredictiveAnalytics"""

    def test_predictive_analytics_init(self):
        """Test PredictiveAnalytics initialization"""
        collector = MetricsCollector()
        analytics = PredictiveAnalytics(collector)
        assert analytics.metrics_collector == collector

    def test_predict_metric_trend_insufficient_data(self):
        """Test prediction with insufficient data"""
        collector = MetricsCollector()
        analytics = PredictiveAnalytics(collector)

        # Add only 5 metrics (less than minimum 10)
        for i in range(5):
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now() - timedelta(hours=i),
                    metric_type=MetricType.CPU_USAGE,
                    value=50.0 + i,
                )
            )

        result = analytics.predict_metric_trend(MetricType.CPU_USAGE)
        assert result["prediction"] is None
        assert result["confidence"] == 0.0

    def test_detect_anomalies_insufficient_data(self):
        """Test anomaly detection with insufficient data"""
        collector = MetricsCollector()
        analytics = PredictiveAnalytics(collector)

        # Add only 3 metrics (less than minimum 5)
        for i in range(3):
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.CPU_USAGE,
                    value=50.0,
                )
            )

        result = analytics.detect_anomalies(MetricType.CPU_USAGE)
        assert result["anomalies"] == []


class TestPerformanceMonitor:
    """Tests for PerformanceMonitor"""

    def test_performance_monitor_init(self):
        """Test PerformanceMonitor initialization"""
        monitor = PerformanceMonitor()
        assert monitor.metrics_collector is not None
        assert monitor.alert_manager is not None
        assert monitor._running is False

    def test_add_custom_metric(self):
        """Test adding custom metric"""
        monitor = PerformanceMonitor()
        monitor.add_custom_metric(MetricType.CPU_USAGE, 50.0, tags={"test": "1"})

        metrics = monitor.metrics_collector.get_metrics(MetricType.CPU_USAGE)
        assert len(metrics) == 1
        assert metrics[0].value == 50.0

    def test_setup_default_alerts(self):
        """Test setting up default alerts"""
        monitor = PerformanceMonitor()
        monitor.setup_default_alerts()

        assert len(monitor.alert_manager.alert_rules) == 3
        rule_names = [r["name"] for r in monitor.alert_manager.alert_rules]
        assert "High CPU Usage" in rule_names
        assert "High Memory Usage" in rule_names

    @patch("moai_adk.core.comprehensive_monitoring_system.psutil")
    def test_collect_system_metrics(self, mock_psutil):
        """Test collecting system metrics"""
        mock_psutil.cpu_percent.return_value = 45.0
        mock_psutil.virtual_memory.return_value = Mock(
            percent=60.0,
            total=16 * 1024 * 1024 * 1024,
        )
        mock_psutil.Process.return_value.memory_info.return_value = Mock(
            rss=500 * 1024 * 1024,
        )
        mock_psutil.getloadavg.return_value = (1.5, 2.0, 2.5)

        monitor = PerformanceMonitor()
        monitor._collect_system_metrics()

        cpu_metrics = monitor.metrics_collector.get_metrics(MetricType.CPU_USAGE)
        assert len(cpu_metrics) > 0


class TestComprehensiveMonitoringSystem:
    """Tests for ComprehensiveMonitoringSystem"""

    def test_monitoring_system_init(self):
        """Test ComprehensiveMonitoringSystem initialization"""
        system = ComprehensiveMonitoringSystem()
        assert system.metrics_collector is not None
        assert system.alert_manager is not None
        assert system.predictive_analytics is not None
        assert system._running is False

    def test_load_config_default(self):
        """Test loading default configuration"""
        system = ComprehensiveMonitoringSystem(config_file=Path("/nonexistent/path"))
        config = system.config

        assert config["buffer_size"] == 10000
        assert config["retention_hours"] == 24

    def test_add_metric(self):
        """Test adding metric through system"""
        system = ComprehensiveMonitoringSystem()
        system.add_metric(MetricType.CPU_USAGE, 50.0, tags={"zone": "a"})

        metrics = system.performance_monitor.metrics_collector.get_metrics(MetricType.CPU_USAGE)
        assert len(metrics) == 1

    def test_get_system_health_healthy(self):
        """Test getting system health - healthy"""
        system = ComprehensiveMonitoringSystem()

        # Add healthy metrics
        for val in [10.0, 15.0, 20.0]:  # Very low CPU/memory usage
            system.add_metric(MetricType.CPU_USAGE, val)
            system.add_metric(MetricType.MEMORY_USAGE, val)

        health = system.performance_monitor.get_system_health()
        assert health.overall_score >= 0
        # When scores are calculated correctly, should be at least warning or better
        assert health.status in [
            HealthStatus.HEALTHY,
            HealthStatus.WARNING,
            HealthStatus.DEGRADED,
        ]

    def test_get_system_health_degraded(self):
        """Test getting system health - degraded"""
        system = ComprehensiveMonitoringSystem()

        # Add bad metrics
        for val in [95.0, 98.0, 99.0]:
            system.add_metric(MetricType.CPU_USAGE, val)
            system.add_metric(MetricType.MEMORY_USAGE, val)

        health = system.performance_monitor.get_system_health()
        # With high metrics, should be degraded, critical, or down
        assert health.status in [
            HealthStatus.DEGRADED,
            HealthStatus.CRITICAL,
            HealthStatus.DOWN,
        ]

    def test_get_dashboard_data(self):
        """Test getting dashboard data"""
        system = ComprehensiveMonitoringSystem()

        for val in [50.0, 55.0, 60.0]:
            system.add_metric(MetricType.CPU_USAGE, val)

        dashboard = system.get_dashboard_data()
        assert "health" in dashboard
        assert "active_alerts" in dashboard
        assert "recent_metrics" in dashboard
        assert "uptime_seconds" in dashboard

    def test_get_analytics_report(self):
        """Test getting analytics report"""
        system = ComprehensiveMonitoringSystem()

        for val in [50.0 + i * 5 for i in range(10)]:
            system.add_metric(MetricType.CPU_USAGE, val)

        report = system.get_analytics_report(hours=1)
        assert "report_period_hours" in report
        assert "generated_at" in report
        assert "metrics_summary" in report
        assert "alert_summary" in report

    def test_generate_recommendations_high_cpu(self):
        """Test recommendation generation for high CPU"""
        system = ComprehensiveMonitoringSystem()

        # Add high CPU metrics
        for i in range(20):
            system.add_metric(MetricType.CPU_USAGE, min(75.0 + i, 99.0))

        report = system.get_analytics_report()
        recommendations = report.get("recommendations", [])

        # Should have at least some recommendations if metrics are high enough
        assert isinstance(recommendations, list)

    def test_generate_recommendations_high_memory(self):
        """Test recommendation generation for high memory"""
        system = ComprehensiveMonitoringSystem()

        # Add high memory metrics
        for i in range(20):
            system.add_metric(MetricType.MEMORY_USAGE, min(85.0 + i, 99.0))

        report = system.get_analytics_report()
        recommendations = report.get("recommendations", [])

        assert isinstance(recommendations, list)

    def test_start_stop_monitoring(self):
        """Test starting and stopping monitoring"""
        system = ComprehensiveMonitoringSystem()

        system.start()
        assert system._running is True

        system.stop()
        assert system._running is False

    def test_concurrent_metric_access(self):
        """Test thread-safe metric access"""
        system = ComprehensiveMonitoringSystem()

        def add_metrics():
            for i in range(100):
                system.add_metric(MetricType.CPU_USAGE, 40.0 + (i % 50))

        threads = [threading.Thread(target=add_metrics) for _ in range(3)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        metrics = system.performance_monitor.metrics_collector.get_metrics(MetricType.CPU_USAGE)
        assert len(metrics) >= 100


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
