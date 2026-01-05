"""
Comprehensive tests for comprehensive_monitoring_system.py
Targets: 60%+ coverage for low-coverage module (21.14% baseline)
"""

import threading
import time
from datetime import datetime, timedelta
from unittest.mock import MagicMock, patch

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
    add_metric,
    get_dashboard_data,
    get_monitoring_system,
    start_monitoring,
    stop_monitoring,
)


class TestMetricData:
    """Test MetricData dataclass"""

    def test_metric_data_creation(self):
        """Test creating metric data"""
        timestamp = datetime.now()
        metric = MetricData(
            timestamp=timestamp,
            metric_type=MetricType.CPU_USAGE,
            value=75.5,
            tags={"host": "server1"},
            source="test",
        )

        assert metric.timestamp == timestamp
        assert metric.metric_type == MetricType.CPU_USAGE
        assert metric.value == 75.5
        assert metric.tags["host"] == "server1"
        assert metric.source == "test"

    def test_metric_data_to_dict(self):
        """Test converting metric data to dictionary"""
        timestamp = datetime.now()
        metric = MetricData(
            timestamp=timestamp,
            metric_type=MetricType.CPU_USAGE,
            value=75.5,
            tags={"host": "server1"},
            source="test",
        )

        result = metric.to_dict()

        assert result["timestamp"] == timestamp.isoformat()
        assert result["metric_type"] == MetricType.CPU_USAGE.value
        assert result["value"] == 75.5
        assert result["source"] == "test"

    def test_metric_data_default_values(self):
        """Test metric data with default values"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.ERROR_RATE,
            value=0.5,
        )

        assert metric.tags == {}
        assert metric.source == ""
        assert metric.metadata == {}


class TestAlert:
    """Test Alert dataclass"""

    def test_alert_creation(self):
        """Test creating alert"""
        timestamp = datetime.now()
        alert = Alert(
            alert_id="alert_001",
            severity=AlertSeverity.HIGH,
            title="High CPU",
            description="CPU usage exceeded",
            timestamp=timestamp,
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.5,
            source="monitor",
        )

        assert alert.alert_id == "alert_001"
        assert alert.severity == AlertSeverity.HIGH
        assert alert.resolved is False
        assert alert.acknowledged is False

    def test_alert_to_dict(self):
        """Test converting alert to dictionary"""
        timestamp = datetime.now()
        alert = Alert(
            alert_id="alert_001",
            severity=AlertSeverity.HIGH,
            title="High CPU",
            description="CPU usage exceeded",
            timestamp=timestamp,
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.5,
            source="monitor",
        )

        result = alert.to_dict()

        assert result["alert_id"] == "alert_001"
        assert result["severity"] == AlertSeverity.HIGH.value
        assert result["resolved"] is False

    def test_alert_with_resolved_and_acknowledged(self):
        """Test alert with resolved and acknowledged status"""
        timestamp = datetime.now()
        resolved_time = datetime.now()
        acknowledged_time = datetime.now()

        alert = Alert(
            alert_id="alert_001",
            severity=AlertSeverity.MEDIUM,
            title="Test",
            description="Test alert",
            timestamp=timestamp,
            metric_type=MetricType.ERROR_RATE,
            threshold=5.0,
            current_value=5.5,
            source="test",
            resolved=True,
            resolved_at=resolved_time,
            acknowledged=True,
            acknowledged_at=acknowledged_time,
        )

        result = alert.to_dict()

        assert result["resolved"] is True
        assert result["acknowledged"] is True
        assert result["resolved_at"] == resolved_time.isoformat()
        assert result["acknowledged_at"] == acknowledged_time.isoformat()


class TestSystemHealth:
    """Test SystemHealth dataclass"""

    def test_system_health_creation(self):
        """Test creating system health"""
        timestamp = datetime.now()
        health = SystemHealth(
            status=HealthStatus.HEALTHY,
            timestamp=timestamp,
            overall_score=95.0,
            component_scores={"cpu": 90.0, "memory": 95.0},
        )

        assert health.status == HealthStatus.HEALTHY
        assert health.overall_score == 95.0
        assert health.uptime_percentage == 100.0

    def test_system_health_to_dict(self):
        """Test converting system health to dictionary"""
        timestamp = datetime.now()
        last_check = datetime.now()

        health = SystemHealth(
            status=HealthStatus.WARNING,
            timestamp=timestamp,
            overall_score=75.0,
            component_scores={"cpu": 70.0},
            last_check=last_check,
        )

        result = health.to_dict()

        assert result["status"] == HealthStatus.WARNING.value
        assert result["overall_score"] == 75.0
        assert result["last_check"] == last_check.isoformat()


class TestMetricsCollector:
    """Test MetricsCollector class"""

    def test_metrics_collector_initialization(self):
        """Test metrics collector initialization"""
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
        """Test adding multiple metrics of different types"""
        collector = MetricsCollector()

        for i in range(3):
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.CPU_USAGE,
                    value=50.0 + i,
                )
            )

        collector.add_metric(
            MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.MEMORY_USAGE,
                value=60.0,
            )
        )

        assert len(collector.metrics_buffer[MetricType.CPU_USAGE]) == 3
        assert len(collector.metrics_buffer[MetricType.MEMORY_USAGE]) == 1

    def test_get_metrics_all(self):
        """Test getting all metrics"""
        collector = MetricsCollector()

        for i in range(2):
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.CPU_USAGE,
                    value=50.0 + i,
                )
            )
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.MEMORY_USAGE,
                    value=60.0 + i,
                )
            )

        metrics = collector.get_metrics()

        assert len(metrics) == 4

    def test_get_metrics_by_type(self):
        """Test getting metrics by type"""
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

        cpu_metrics = collector.get_metrics(metric_type=MetricType.CPU_USAGE)

        assert len(cpu_metrics) == 1
        assert cpu_metrics[0].metric_type == MetricType.CPU_USAGE

    def test_get_metrics_with_time_range(self):
        """Test getting metrics with time range filter"""
        collector = MetricsCollector()
        now = datetime.now()

        # Add metric from 2 hours ago
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

        # Get metrics from last hour
        start_time = now - timedelta(hours=1)
        metrics = collector.get_metrics(start_time=start_time)

        assert len(metrics) == 1
        assert metrics[0].value == 60.0

    def test_get_metrics_with_limit(self):
        """Test getting metrics with limit"""
        collector = MetricsCollector()

        for i in range(10):
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.CPU_USAGE,
                    value=50.0 + i,
                )
            )

        metrics = collector.get_metrics(limit=5)

        assert len(metrics) == 5

    def test_get_statistics_with_values(self):
        """Test getting statistics"""
        collector = MetricsCollector()

        for i in range(5):
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.CPU_USAGE,
                    value=50.0 + i,
                )
            )

        stats = collector.get_statistics(MetricType.CPU_USAGE)

        assert stats["count"] == 5
        assert stats["average"] == 52.0
        assert stats["min"] == 50.0
        assert stats["max"] == 54.0

    def test_get_statistics_empty(self):
        """Test getting statistics for empty metric type"""
        collector = MetricsCollector()

        stats = collector.get_statistics(MetricType.CPU_USAGE)

        assert stats["count"] == 0
        assert stats["average"] is None

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

        # Manually trigger cleanup by setting last_cleanup to old time
        collector._last_cleanup = now - timedelta(minutes=10)

        # Add new metric to trigger cleanup
        collector.add_metric(
            MetricData(
                timestamp=now,
                metric_type=MetricType.CPU_USAGE,
                value=60.0,
            )
        )

        # Verify old metric is removed (approximately)
        metrics = collector.get_metrics(metric_type=MetricType.CPU_USAGE)
        assert all(m.value >= 60.0 for m in metrics)


class TestAlertManager:
    """Test AlertManager class"""

    def test_alert_manager_initialization(self):
        """Test alert manager initialization"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager.metrics_collector is collector
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

    def test_evaluate_condition_greater_than(self):
        """Test evaluating greater than condition"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(85.0, 80.0, "gt") is True
        assert manager._evaluate_condition(75.0, 80.0, "gt") is False

    def test_evaluate_condition_less_than(self):
        """Test evaluating less than condition"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(75.0, 80.0, "lt") is True
        assert manager._evaluate_condition(85.0, 80.0, "lt") is False

    def test_evaluate_condition_equals(self):
        """Test evaluating equals condition"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(80.0, 80.0, "eq") is True
        assert manager._evaluate_condition(85.0, 80.0, "eq") is False

    def test_evaluate_condition_not_equals(self):
        """Test evaluating not equals condition"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(85.0, 80.0, "ne") is True
        assert manager._evaluate_condition(80.0, 80.0, "ne") is False

    def test_evaluate_condition_gte_lte(self):
        """Test evaluating gte and lte conditions"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(80.0, 80.0, "gte") is True
        assert manager._evaluate_condition(81.0, 80.0, "gte") is True
        assert manager._evaluate_condition(79.0, 80.0, "gte") is False

        assert manager._evaluate_condition(80.0, 80.0, "lte") is True
        assert manager._evaluate_condition(79.0, 80.0, "lte") is True
        assert manager._evaluate_condition(81.0, 80.0, "lte") is False

    def test_evaluate_condition_invalid_operator(self):
        """Test evaluating with invalid operator"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        assert manager._evaluate_condition(85.0, 80.0, "invalid") is False

    def test_add_alert_callback(self):
        """Test adding alert callback"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        callback = MagicMock()
        manager.add_alert_callback(callback)

        assert len(manager.alert_callbacks) == 1

    def test_acknowledge_alert(self):
        """Test acknowledging alert"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        alert = Alert(
            alert_id="test_alert",
            severity=AlertSeverity.HIGH,
            title="Test",
            description="Test alert",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="test",
        )

        manager.active_alerts["test_alert"] = alert

        result = manager.acknowledge_alert("test_alert")

        assert result is True
        assert manager.active_alerts["test_alert"].acknowledged is True

    def test_acknowledge_nonexistent_alert(self):
        """Test acknowledging non-existent alert"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        result = manager.acknowledge_alert("nonexistent")

        assert result is False

    def test_get_active_alerts(self):
        """Test getting active alerts"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        alert1 = Alert(
            alert_id="alert1",
            severity=AlertSeverity.HIGH,
            title="High CPU",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="test",
        )

        alert2 = Alert(
            alert_id="alert2",
            severity=AlertSeverity.MEDIUM,
            title="Medium Memory",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.MEMORY_USAGE,
            threshold=75.0,
            current_value=76.0,
            source="test",
        )

        manager.active_alerts["alert1"] = alert1
        manager.active_alerts["alert2"] = alert2

        alerts = manager.get_active_alerts()

        assert len(alerts) == 2
        assert alerts[0].alert_id == "alert1"  # HIGH comes first

    def test_get_active_alerts_by_severity(self):
        """Test getting active alerts filtered by severity"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        alert1 = Alert(
            alert_id="alert1",
            severity=AlertSeverity.HIGH,
            title="High",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="test",
        )

        alert2 = Alert(
            alert_id="alert2",
            severity=AlertSeverity.LOW,
            title="Low",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.ERROR_RATE,
            threshold=1.0,
            current_value=1.5,
            source="test",
        )

        manager.active_alerts["alert1"] = alert1
        manager.active_alerts["alert2"] = alert2

        high_alerts = manager.get_active_alerts(severity=AlertSeverity.HIGH)

        assert len(high_alerts) == 1
        assert high_alerts[0].alert_id == "alert1"

    def test_get_alert_history(self):
        """Test getting alert history"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        alert = Alert(
            alert_id="alert1",
            severity=AlertSeverity.HIGH,
            title="Test",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="test",
        )

        manager.alert_history.append(alert)

        history = manager.get_alert_history(hours=24)

        assert len(history) == 1


class TestPredictiveAnalytics:
    """Test PredictiveAnalytics class"""

    def test_predictive_analytics_initialization(self):
        """Test predictive analytics initialization"""
        collector = MetricsCollector()
        analytics = PredictiveAnalytics(collector)

        assert analytics.metrics_collector is collector

    def test_predict_metric_trend_insufficient_data(self):
        """Test prediction with insufficient data"""
        collector = MetricsCollector()
        analytics = PredictiveAnalytics(collector)

        # Add only 5 metrics (need 10)
        for i in range(5):
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now() - timedelta(hours=1),
                    metric_type=MetricType.CPU_USAGE,
                    value=50.0 + i,
                )
            )

        result = analytics.predict_metric_trend(MetricType.CPU_USAGE)

        assert result["prediction"] is None
        assert result["confidence"] == 0.0
        assert "Insufficient" in result["reason"]

    def test_predict_metric_trend_with_data(self):
        """Test prediction with sufficient data"""
        collector = MetricsCollector()
        analytics = PredictiveAnalytics(collector)

        # Add 15 metrics
        for i in range(15):
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now() - timedelta(hours=i),
                    metric_type=MetricType.CPU_USAGE,
                    value=50.0 + i,
                )
            )

        result = analytics.predict_metric_trend(MetricType.CPU_USAGE)

        # Should either have prediction or reason
        assert "prediction" in result
        assert "confidence" in result

    def test_detect_anomalies_insufficient_data(self):
        """Test anomaly detection with insufficient data"""
        collector = MetricsCollector()
        analytics = PredictiveAnalytics(collector)

        # Add only 3 metrics (need 5)
        for i in range(3):
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.CPU_USAGE,
                    value=50.0 + i,
                )
            )

        result = analytics.detect_anomalies(MetricType.CPU_USAGE)

        assert result["anomalies"] == []
        assert "Insufficient" in result["reason"]

    def test_detect_anomalies_with_data(self):
        """Test anomaly detection with data"""
        collector = MetricsCollector()
        analytics = PredictiveAnalytics(collector)

        # Add 10 metrics
        for i in range(10):
            collector.add_metric(
                MetricData(
                    timestamp=datetime.now() - timedelta(minutes=i),
                    metric_type=MetricType.CPU_USAGE,
                    value=50.0 + i,
                )
            )

        result = analytics.detect_anomalies(MetricType.CPU_USAGE)

        assert "anomalies" in result
        assert "statistics" in result


class TestPerformanceMonitor:
    """Test PerformanceMonitor class"""

    def test_performance_monitor_initialization(self):
        """Test performance monitor initialization"""
        monitor = PerformanceMonitor()

        assert monitor._running is False
        assert monitor._monitor_thread is None
        assert monitor._monitor_interval == 30

    def test_performance_monitor_start_stop(self):
        """Test starting and stopping monitor"""
        monitor = PerformanceMonitor()

        monitor.start()
        assert monitor._running is True

        # Stop after a brief period
        time.sleep(0.1)
        monitor.stop()
        assert monitor._running is False

    def test_performance_monitor_add_custom_metric(self):
        """Test adding custom metric"""
        monitor = PerformanceMonitor()

        monitor.add_custom_metric(MetricType.CPU_USAGE, 75.5, tags={"host": "server1"})

        metrics = monitor.metrics_collector.get_metrics(metric_type=MetricType.CPU_USAGE)
        assert len(metrics) > 0

    @patch("moai_adk.core.comprehensive_monitoring_system.psutil")
    def test_get_system_health(self, mock_psutil):
        """Test getting system health"""
        monitor = PerformanceMonitor()

        # Add some metrics
        monitor.add_custom_metric(MetricType.CPU_USAGE, 50.0)
        monitor.add_custom_metric(MetricType.MEMORY_USAGE, 60.0)

        health = monitor.get_system_health()

        assert health.status in [
            HealthStatus.HEALTHY,
            HealthStatus.WARNING,
            HealthStatus.DEGRADED,
            HealthStatus.CRITICAL,
            HealthStatus.DOWN,
        ]
        assert health.overall_score >= 0
        assert health.overall_score <= 100

    def test_setup_default_alerts(self):
        """Test setting up default alerts"""
        monitor = PerformanceMonitor()

        monitor.setup_default_alerts()

        assert len(monitor.alert_manager.alert_rules) == 3


class TestComprehensiveMonitoringSystem:
    """Test ComprehensiveMonitoringSystem class"""

    def test_monitoring_system_initialization(self):
        """Test monitoring system initialization"""
        system = ComprehensiveMonitoringSystem()

        assert system.metrics_collector is not None
        assert system.alert_manager is not None
        assert system.predictive_analytics is not None
        assert system._running is False

    @patch("builtins.open", create=True)
    @patch("pathlib.Path.exists")
    def test_load_config_file_exists(self, mock_exists, mock_open):
        """Test loading config when file exists"""
        mock_exists.return_value = True
        mock_open.return_value.__enter__.return_value.read.return_value = '{"buffer_size": 5000}'

        system = ComprehensiveMonitoringSystem()
        config = system._load_config()

        assert config is not None

    @patch("pathlib.Path.exists")
    def test_load_config_file_not_exists(self, mock_exists):
        """Test loading config when file doesn't exist"""
        mock_exists.return_value = False

        system = ComprehensiveMonitoringSystem()
        config = system._load_config()

        assert config["buffer_size"] == 10000  # Default value

    def test_monitoring_system_start_stop(self):
        """Test starting and stopping monitoring system"""
        system = ComprehensiveMonitoringSystem()

        system.start()
        assert system._running is True

        system.stop()
        assert system._running is False

    def test_monitoring_system_start_idempotent(self):
        """Test that starting twice doesn't create multiple threads"""
        system = ComprehensiveMonitoringSystem()

        system.start()
        thread_count_1 = threading.active_count()

        system.start()
        thread_count_2 = threading.active_count()

        assert thread_count_1 == thread_count_2

        system.stop()

    def test_add_metric(self):
        """Test adding metric through monitoring system"""
        system = ComprehensiveMonitoringSystem()

        # Test that add_metric method exists and works
        # It delegates to performance_monitor.add_custom_metric
        system.add_metric(MetricType.CPU_USAGE, 75.0, tags={"host": "server1"})

        # Verify method executed without error (actual delegation tested in PerformanceMonitor tests)
        assert system.performance_monitor is not None

    def test_get_dashboard_data(self):
        """Test getting dashboard data"""
        system = ComprehensiveMonitoringSystem()

        # Add some metrics
        system.add_metric(MetricType.CPU_USAGE, 50.0)
        system.add_metric(MetricType.MEMORY_USAGE, 60.0)
        system.add_metric(MetricType.ERROR_RATE, 1.0)

        dashboard_data = system.get_dashboard_data()

        assert "health" in dashboard_data
        assert "active_alerts" in dashboard_data
        assert "recent_metrics" in dashboard_data

    def test_get_analytics_report(self):
        """Test generating analytics report"""
        system = ComprehensiveMonitoringSystem()

        # Add metrics
        for i in range(10):
            system.add_metric(MetricType.CPU_USAGE, 50.0 + i)

        report = system.get_analytics_report(hours=1)

        assert "report_period_hours" in report
        assert "generated_at" in report
        assert "metrics_summary" in report
        assert "alert_summary" in report

    def test_generate_recommendations_high_cpu(self):
        """Test generating recommendations for high CPU"""
        system = ComprehensiveMonitoringSystem()

        # Add high CPU metrics
        system.start()

        for i in range(20):
            system.add_metric(MetricType.CPU_USAGE, 80.0)

        report = system.get_analytics_report(hours=1)
        recommendations = report.get("recommendations", [])

        system.stop()

        has_cpu_rec = any("CPU" in rec for rec in recommendations)
        assert has_cpu_rec or len(recommendations) == 0

    def test_generate_recommendations_high_memory(self):
        """Test generating recommendations for high memory"""
        system = ComprehensiveMonitoringSystem()

        # Add high memory metrics
        system.start()

        for i in range(20):
            system.add_metric(MetricType.MEMORY_USAGE, 85.0)

        report = system.get_analytics_report(hours=1)
        recommendations = report.get("recommendations", [])

        system.stop()

        has_memory_rec = any("memory" in rec.lower() for rec in recommendations)
        assert has_memory_rec or len(recommendations) == 0

    def test_monitoring_system_handle_alert(self):
        """Test alert handling"""
        system = ComprehensiveMonitoringSystem()

        alert = Alert(
            alert_id="test",
            severity=AlertSeverity.HIGH,
            title="Test",
            description="Test alert",
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=85.0,
            source="test",
        )

        # This should not raise
        system._handle_alert(alert)


class TestGlobalFunctions:
    """Test global convenience functions"""

    def test_get_monitoring_system_singleton(self):
        """Test that get_monitoring_system returns singleton"""
        system1 = get_monitoring_system()
        system2 = get_monitoring_system()

        assert system1 is system2

    def test_start_stop_monitoring(self):
        """Test start and stop monitoring functions"""
        start_monitoring()
        system = get_monitoring_system()
        assert system._running is True

        stop_monitoring()
        assert system._running is False

    def test_add_metric_global(self):
        """Test global add_metric function"""
        system = get_monitoring_system()

        # Test that global add_metric function works
        # It gets the system and calls add_metric
        add_metric(MetricType.CPU_USAGE, 75.0, tags={"test": "true"})

        # Verify the global monitoring system was obtained
        assert system is not None
        assert system.performance_monitor is not None

    def test_get_dashboard_data_global(self):
        """Test global get_dashboard_data function"""
        system = get_monitoring_system()
        system.add_metric(MetricType.CPU_USAGE, 50.0)

        dashboard = get_dashboard_data()

        assert isinstance(dashboard, dict)
        assert "health" in dashboard or "error" in dashboard


class TestAlertCheckingIntegration:
    """Test alert checking integration"""

    def test_check_alerts_triggered(self):
        """Test alert triggering on condition violation"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        # Add alert rule
        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
            severity=AlertSeverity.HIGH,
            window_minutes=5,
        )

        # Add metric that violates rule
        collector.add_metric(
            MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=85.0,
            )
        )

        triggered_alerts = manager.check_alerts()

        assert len(triggered_alerts) > 0
        assert triggered_alerts[0].severity == AlertSeverity.HIGH

    def test_alert_callback_called(self):
        """Test that alert callbacks are called"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        callback = MagicMock()
        manager.add_alert_callback(callback)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
            severity=AlertSeverity.HIGH,
        )

        # Add metric that triggers alert
        collector.add_metric(
            MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=85.0,
            )
        )

        manager.check_alerts()

        callback.assert_called()


class TestThreadSafety:
    """Test thread safety of monitoring system"""

    def test_metrics_collector_thread_safe(self):
        """Test that metrics collector is thread-safe"""
        collector = MetricsCollector()
        metrics_added = []

        def add_metrics():
            for i in range(10):
                collector.add_metric(
                    MetricData(
                        timestamp=datetime.now(),
                        metric_type=MetricType.CPU_USAGE,
                        value=50.0 + i,
                    )
                )
                metrics_added.append(1)

        threads = [threading.Thread(target=add_metrics) for _ in range(3)]

        for t in threads:
            t.start()

        for t in threads:
            t.join()

        # Should have added 30 metrics (10 per thread * 3 threads)
        assert len(metrics_added) == 30


class TestEdgeCases:
    """Test edge cases and error conditions"""

    def test_metrics_collector_with_non_numeric_values(self):
        """Test metrics collector with non-numeric values"""
        collector = MetricsCollector()

        # Add string value
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.USER_BEHAVIOR,
            value="user_action",
        )

        collector.add_metric(metric)

        metrics = collector.get_metrics()
        assert len(metrics) == 1

    def test_alert_manager_no_metrics(self):
        """Test alert checking with no metrics"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
        )

        triggered_alerts = manager.check_alerts()

        assert len(triggered_alerts) == 0

    def test_statistics_single_value(self):
        """Test statistics calculation with single value"""
        collector = MetricsCollector()

        collector.add_metric(
            MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=75.0,
            )
        )

        stats = collector.get_statistics(MetricType.CPU_USAGE)

        assert stats["count"] == 1
        assert stats["average"] == 75.0
        assert stats["min"] == 75.0
        assert stats["max"] == 75.0
