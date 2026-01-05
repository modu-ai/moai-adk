"""
Comprehensive executable tests for Monitoring System.

These tests exercise actual code paths including:
- Metric data collection and serialization
- Alert creation and management
- System health status tracking
- Performance analytics
"""

from datetime import datetime, timedelta

from moai_adk.core.comprehensive_monitoring_system import (
    Alert,
    AlertSeverity,
    HealthStatus,
    MetricData,
    MetricType,
    SystemHealth,
)


class TestMetricType:
    """Test MetricType enum."""

    def test_all_metric_types(self):
        """Test all metric types are defined."""
        assert MetricType.SYSTEM_PERFORMANCE.value == "system_performance"
        assert MetricType.USER_BEHAVIOR.value == "user_behavior"
        assert MetricType.TOKEN_USAGE.value == "token_usage"
        assert MetricType.ERROR_RATE.value == "error_rate"
        assert MetricType.RESPONSE_TIME.value == "response_time"
        assert MetricType.MEMORY_USAGE.value == "memory_usage"
        assert MetricType.CPU_USAGE.value == "cpu_usage"
        assert MetricType.THROUGHPUT.value == "throughput"
        assert MetricType.AVAILABILITY.value == "availability"

    def test_metric_type_enum_values(self):
        """Test metric type values are strings."""
        for metric_type in MetricType:
            assert isinstance(metric_type.value, str)


class TestAlertSeverity:
    """Test AlertSeverity enum."""

    def test_all_alert_severities(self):
        """Test all alert severities are defined."""
        assert AlertSeverity.LOW.value == 1
        assert AlertSeverity.MEDIUM.value == 2
        assert AlertSeverity.HIGH.value == 3
        assert AlertSeverity.CRITICAL.value == 4
        assert AlertSeverity.EMERGENCY.value == 5

    def test_alert_severity_ordering(self):
        """Test alert severity values follow hierarchy."""
        assert AlertSeverity.LOW.value < AlertSeverity.MEDIUM.value
        assert AlertSeverity.MEDIUM.value < AlertSeverity.HIGH.value
        assert AlertSeverity.HIGH.value < AlertSeverity.CRITICAL.value
        assert AlertSeverity.CRITICAL.value < AlertSeverity.EMERGENCY.value


class TestHealthStatus:
    """Test HealthStatus enum."""

    def test_all_health_statuses(self):
        """Test all health statuses are defined."""
        assert HealthStatus.HEALTHY.value == "healthy"
        assert HealthStatus.WARNING.value == "warning"
        assert HealthStatus.DEGRADED.value == "degraded"
        assert HealthStatus.CRITICAL.value == "critical"
        assert HealthStatus.DOWN.value == "down"


class TestMetricData:
    """Test MetricData class."""

    def test_create_metric_data(self):
        """Test creating metric data."""
        now = datetime.now()
        metric = MetricData(
            timestamp=now,
            metric_type=MetricType.CPU_USAGE,
            value=45.5,
            tags={"host": "server-1", "env": "prod"},
            source="monitoring_agent",
        )

        assert metric.timestamp == now
        assert metric.metric_type == MetricType.CPU_USAGE
        assert metric.value == 45.5
        assert metric.tags["host"] == "server-1"
        assert metric.source == "monitoring_agent"

    def test_metric_data_with_different_value_types(self):
        """Test metric data with different value types."""
        now = datetime.now()

        # Integer value
        metric_int = MetricData(
            timestamp=now,
            metric_type=MetricType.ERROR_RATE,
            value=5,
        )
        assert metric_int.value == 5

        # Float value
        metric_float = MetricData(
            timestamp=now,
            metric_type=MetricType.RESPONSE_TIME,
            value=123.45,
        )
        assert metric_float.value == 123.45

        # String value
        metric_str = MetricData(
            timestamp=now,
            metric_type=MetricType.USER_BEHAVIOR,
            value="click",
        )
        assert metric_str.value == "click"

        # Boolean value
        metric_bool = MetricData(
            timestamp=now,
            metric_type=MetricType.AVAILABILITY,
            value=True,
        )
        assert metric_bool.value is True

    def test_metric_data_defaults(self):
        """Test metric data default values."""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.THROUGHPUT,
            value=100,
        )

        assert metric.tags == {}
        assert metric.source == ""
        assert metric.metadata == {}

    def test_metric_data_to_dict(self):
        """Test serializing metric data to dict."""
        now = datetime.now()
        metric = MetricData(
            timestamp=now,
            metric_type=MetricType.MEMORY_USAGE,
            value=2048,
            tags={"service": "api"},
            source="prometheus",
            metadata={"unit": "MB"},
        )

        metric_dict = metric.to_dict()

        assert metric_dict["timestamp"] == now.isoformat()
        assert metric_dict["metric_type"] == "memory_usage"
        assert metric_dict["value"] == 2048
        assert metric_dict["tags"]["service"] == "api"
        assert metric_dict["source"] == "prometheus"
        assert metric_dict["metadata"]["unit"] == "MB"

    def test_metric_data_serialization_roundtrip(self):
        """Test metric data serialization and deserialization."""
        now = datetime.now()
        original = MetricData(
            timestamp=now,
            metric_type=MetricType.TOKEN_USAGE,
            value=5000,
            tags={"user": "alice", "api": "v1"},
            source="tracking",
        )

        # Serialize
        metric_dict = original.to_dict()

        # Reconstruct
        reconstructed = MetricData(
            timestamp=datetime.fromisoformat(metric_dict["timestamp"]),
            metric_type=MetricType(metric_dict["metric_type"]),
            value=metric_dict["value"],
            tags=metric_dict["tags"],
            source=metric_dict["source"],
            metadata=metric_dict["metadata"],
        )

        assert reconstructed.value == original.value
        assert reconstructed.metric_type == original.metric_type
        assert reconstructed.tags == original.tags


class TestAlert:
    """Test Alert class."""

    def test_create_alert(self):
        """Test creating alert."""
        now = datetime.now()
        alert = Alert(
            alert_id="alert-001",
            severity=AlertSeverity.HIGH,
            title="High CPU Usage",
            description="CPU usage exceeded threshold",
            timestamp=now,
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            current_value=95.2,
            source="monitoring_system",
        )

        assert alert.alert_id == "alert-001"
        assert alert.severity == AlertSeverity.HIGH
        assert alert.title == "High CPU Usage"
        assert alert.threshold == 80.0
        assert alert.current_value == 95.2

    def test_alert_defaults(self):
        """Test alert default values."""
        alert = Alert(
            alert_id="alert-002",
            severity=AlertSeverity.CRITICAL,
            title="Critical Issue",
            description="System critical",
            timestamp=datetime.now(),
            metric_type=MetricType.AVAILABILITY,
            threshold=99.0,
            current_value=50.0,
            source="system",
        )

        assert alert.resolved is False
        assert alert.resolved_at is None
        assert alert.acknowledged is False
        assert alert.acknowledged_at is None
        assert alert.tags == {}

    def test_alert_to_dict(self):
        """Test serializing alert to dict."""
        now = datetime.now()
        alert = Alert(
            alert_id="alert-003",
            severity=AlertSeverity.MEDIUM,
            title="Memory Warning",
            description="Memory usage high",
            timestamp=now,
            metric_type=MetricType.MEMORY_USAGE,
            threshold=512.0,
            current_value=600.0,
            source="monitor",
            tags={"component": "api"},
        )

        alert_dict = alert.to_dict()

        assert alert_dict["alert_id"] == "alert-003"
        assert alert_dict["severity"] == 2  # MEDIUM value
        assert alert_dict["title"] == "Memory Warning"
        assert alert_dict["timestamp"] == now.isoformat()
        assert alert_dict["threshold"] == 512.0
        assert alert_dict["current_value"] == 600.0

    def test_alert_acknowledge(self):
        """Test acknowledging alert."""
        alert = Alert(
            alert_id="alert-004",
            severity=AlertSeverity.HIGH,
            title="Test Alert",
            description="Test",
            timestamp=datetime.now(),
            metric_type=MetricType.ERROR_RATE,
            threshold=5.0,
            current_value=10.0,
            source="test",
        )

        assert alert.acknowledged is False

        alert.acknowledged = True
        alert.acknowledged_at = datetime.now()

        assert alert.acknowledged is True
        assert alert.acknowledged_at is not None

    def test_alert_resolution(self):
        """Test resolving alert."""
        alert = Alert(
            alert_id="alert-005",
            severity=AlertSeverity.CRITICAL,
            title="Critical Issue",
            description="Critical",
            timestamp=datetime.now(),
            metric_type=MetricType.THROUGHPUT,
            threshold=100.0,
            current_value=200.0,
            source="test",
        )

        assert alert.resolved is False

        alert.resolved = True
        alert.resolved_at = datetime.now()

        assert alert.resolved is True
        assert alert.resolved_at is not None

    def test_alert_with_tags(self):
        """Test alert with custom tags."""
        alert = Alert(
            alert_id="alert-006",
            severity=AlertSeverity.HIGH,
            title="Service Degradation",
            description="Service response slow",
            timestamp=datetime.now(),
            metric_type=MetricType.RESPONSE_TIME,
            threshold=1000.0,
            current_value=5000.0,
            source="monitoring",
            tags={
                "service": "api",
                "region": "us-west",
                "env": "prod",
            },
        )

        assert alert.tags["service"] == "api"
        assert alert.tags["region"] == "us-west"
        assert len(alert.tags) == 3

    def test_alert_serialization(self):
        """Test alert serialization roundtrip."""
        now = datetime.now()
        resolved_at = now + timedelta(hours=1)

        alert = Alert(
            alert_id="alert-007",
            severity=AlertSeverity.CRITICAL,
            title="Emergency Alert",
            description="Emergency situation",
            timestamp=now,
            metric_type=MetricType.AVAILABILITY,
            threshold=99.9,
            current_value=50.0,
            source="health_check",
            tags={"priority": "emergency"},
            resolved=True,
            resolved_at=resolved_at,
            acknowledged=True,
            acknowledged_at=now + timedelta(minutes=5),
        )

        alert_dict = alert.to_dict()

        assert alert_dict["resolved"] is True
        assert alert_dict["acknowledged"] is True
        assert alert_dict["resolved_at"] == resolved_at.isoformat()


class TestSystemHealth:
    """Test SystemHealth class."""

    def test_create_system_health(self):
        """Test creating system health status."""
        now = datetime.now()
        health = SystemHealth(
            status=HealthStatus.HEALTHY,
            timestamp=now,
            overall_score=95.0,
            component_scores={
                "api": 98.0,
                "database": 92.0,
                "cache": 96.0,
            },
        )

        assert health.status == HealthStatus.HEALTHY
        assert health.overall_score == 95.0
        assert health.component_scores["api"] == 98.0

    def test_system_health_defaults(self):
        """Test system health default values."""
        health = SystemHealth(
            status=HealthStatus.WARNING,
            timestamp=datetime.now(),
            overall_score=75.0,
        )

        assert health.active_alerts == []
        assert health.recent_metrics == {}
        assert health.uptime_percentage == 100.0
        assert health.last_check is None

    def test_system_health_degraded(self):
        """Test degraded system health."""
        health = SystemHealth(
            status=HealthStatus.DEGRADED,
            timestamp=datetime.now(),
            overall_score=50.0,
            component_scores={
                "api": 40.0,
                "database": 60.0,
            },
            active_alerts=["alert-001", "alert-002"],
        )

        assert health.status == HealthStatus.DEGRADED
        assert health.overall_score == 50.0
        assert len(health.active_alerts) == 2

    def test_system_health_critical(self):
        """Test critical system health."""
        health = SystemHealth(
            status=HealthStatus.CRITICAL,
            timestamp=datetime.now(),
            overall_score=25.0,
            active_alerts=["alert-critical-001"],
            uptime_percentage=80.0,
        )

        assert health.status == HealthStatus.CRITICAL
        assert health.overall_score == 25.0
        assert health.uptime_percentage == 80.0

    def test_system_health_down(self):
        """Test down system health."""
        health = SystemHealth(
            status=HealthStatus.DOWN,
            timestamp=datetime.now(),
            overall_score=0.0,
            active_alerts=["alert-down-001", "alert-down-002"],
            uptime_percentage=0.0,
        )

        assert health.status == HealthStatus.DOWN
        assert health.overall_score == 0.0
        assert health.uptime_percentage == 0.0

    def test_system_health_to_dict(self):
        """Test serializing system health to dict."""
        now = datetime.now()
        last_check = now - timedelta(minutes=5)

        health = SystemHealth(
            status=HealthStatus.HEALTHY,
            timestamp=now,
            overall_score=92.5,
            component_scores={
                "api": 95.0,
                "database": 90.0,
            },
            active_alerts=["alert-001"],
            recent_metrics={
                "cpu_usage": 45.0,
                "memory_usage": 60.0,
            },
            uptime_percentage=99.95,
            last_check=last_check,
        )

        health_dict = health.to_dict()

        assert health_dict["status"] == "healthy"
        assert health_dict["overall_score"] == 92.5
        assert health_dict["timestamp"] == now.isoformat()
        assert health_dict["component_scores"]["api"] == 95.0
        assert len(health_dict["active_alerts"]) == 1
        assert health_dict["uptime_percentage"] == 99.95

    def test_system_health_component_scores(self):
        """Test system health with detailed component scores."""
        health = SystemHealth(
            status=HealthStatus.HEALTHY,
            timestamp=datetime.now(),
            overall_score=88.0,
            component_scores={
                "api_gateway": 95.0,
                "auth_service": 92.0,
                "database": 85.0,
                "cache": 90.0,
                "message_queue": 88.0,
                "monitoring": 80.0,
            },
        )

        assert len(health.component_scores) == 6
        assert all(0 <= score <= 100 for score in health.component_scores.values())

    def test_system_health_recent_metrics(self):
        """Test system health with recent metrics."""
        health = SystemHealth(
            status=HealthStatus.WARNING,
            timestamp=datetime.now(),
            overall_score=70.0,
            recent_metrics={
                "cpu_usage": 65.0,
                "memory_usage": 75.0,
                "error_rate": 2.5,
                "response_time_p95": 450.0,
                "requests_per_second": 1200.0,
            },
        )

        assert health.recent_metrics["cpu_usage"] == 65.0
        assert health.recent_metrics["error_rate"] == 2.5
        assert len(health.recent_metrics) == 5

    def test_system_health_uptime_tracking(self):
        """Test system health uptime percentage."""
        # Perfect uptime
        health_perfect = SystemHealth(
            status=HealthStatus.HEALTHY,
            timestamp=datetime.now(),
            overall_score=100.0,
            uptime_percentage=100.0,
        )
        assert health_perfect.uptime_percentage == 100.0

        # Partial downtime
        health_partial = SystemHealth(
            status=HealthStatus.WARNING,
            timestamp=datetime.now(),
            overall_score=80.0,
            uptime_percentage=99.5,
        )
        assert health_partial.uptime_percentage == 99.5

        # Significant downtime
        health_significant = SystemHealth(
            status=HealthStatus.DEGRADED,
            timestamp=datetime.now(),
            overall_score=50.0,
            uptime_percentage=95.0,
        )
        assert health_significant.uptime_percentage == 95.0

    def test_system_health_last_check_timestamp(self):
        """Test tracking last health check time."""
        now = datetime.now()
        check_time = now - timedelta(minutes=1)

        health = SystemHealth(
            status=HealthStatus.HEALTHY,
            timestamp=now,
            overall_score=95.0,
            last_check=check_time,
        )

        assert health.last_check == check_time
        assert (now - health.last_check).total_seconds() > 0

    def test_system_health_score_boundaries(self):
        """Test system health score boundaries."""
        # Test minimum score
        health_min = SystemHealth(
            status=HealthStatus.DOWN,
            timestamp=datetime.now(),
            overall_score=0.0,
        )
        assert health_min.overall_score == 0.0

        # Test maximum score
        health_max = SystemHealth(
            status=HealthStatus.HEALTHY,
            timestamp=datetime.now(),
            overall_score=100.0,
        )
        assert health_max.overall_score == 100.0

        # Test mid-range score
        health_mid = SystemHealth(
            status=HealthStatus.WARNING,
            timestamp=datetime.now(),
            overall_score=50.5,
        )
        assert health_mid.overall_score == 50.5
