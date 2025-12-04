"""Comprehensive test coverage for Realtime Monitoring Dashboard.

This module provides extensive unit tests for monitoring, alerting, and analytics functionality.
"""

import asyncio
import pytest
from datetime import datetime, timedelta
from pathlib import Path
from unittest.mock import MagicMock, AsyncMock, patch, Mock
from collections import defaultdict, deque

from moai_adk.core.realtime_monitoring_dashboard import (
    MetricType,
    AlertSeverity,
    DashboardType,
    MetricData,
)


class TestMetricType:
    """Test MetricType enum"""

    def test_metric_type_enum_values(self):
        """Test MetricType has all expected values"""
        assert MetricType.SYSTEM_PERFORMANCE is not None
        assert MetricType.HOOK_EXECUTION is not None
        assert MetricType.ERROR_RATE is not None
        assert MetricType.RESPONSE_TIME is not None
        assert MetricType.MEMORY_USAGE is not None
        assert MetricType.CPU_USAGE is not None
        assert MetricType.NETWORK_IO is not None
        assert MetricType.DISK_IO is not None
        assert MetricType.THROUGHPUT is not None
        assert MetricType.AVAILABILITY is not None
        assert MetricType.CACHE_PERFORMANCE is not None
        assert MetricType.DATABASE_PERFORMANCE is not None

    def test_metric_type_values_are_strings(self):
        """Test metric type values are strings"""
        for metric_type in MetricType:
            assert isinstance(metric_type.value, str)


class TestAlertSeverity:
    """Test AlertSeverity enum"""

    def test_alert_severity_enum_values(self):
        """Test AlertSeverity has all expected values"""
        assert AlertSeverity.INFO is not None
        assert AlertSeverity.WARNING is not None
        assert AlertSeverity.ERROR is not None
        assert AlertSeverity.CRITICAL is not None
        assert AlertSeverity.EMERGENCY is not None

    def test_alert_severity_ordering(self):
        """Test AlertSeverity ordering"""
        assert AlertSeverity.INFO.value < AlertSeverity.WARNING.value
        assert AlertSeverity.WARNING.value < AlertSeverity.ERROR.value
        assert AlertSeverity.ERROR.value < AlertSeverity.CRITICAL.value
        assert AlertSeverity.CRITICAL.value < AlertSeverity.EMERGENCY.value


class TestDashboardType:
    """Test DashboardType enum"""

    def test_dashboard_type_enum_values(self):
        """Test DashboardType has all expected values"""
        assert DashboardType.SYSTEM_OVERVIEW is not None
        assert DashboardType.PERFORMANCE_ANALYTICS is not None
        assert DashboardType.ERROR_MONITORING is not None
        assert DashboardType.HOOK_ANALYSIS is not None
        assert DashboardType.RESOURCE_USAGE is not None
        assert DashboardType.CUSTOM is not None


class TestMetricData:
    """Test MetricData dataclass"""

    def test_metric_data_initialization(self):
        """Test MetricData initializes correctly"""
        timestamp = datetime.now()
        metric = MetricData(
            timestamp=timestamp,
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
            tags={"host": "server1"},
            source="system_monitor"
        )
        assert metric.timestamp == timestamp
        assert metric.metric_type == MetricType.CPU_USAGE
        assert metric.value == 50.0
        assert metric.tags["host"] == "server1"

    def test_metric_data_defaults(self):
        """Test MetricData default values"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=50.0
        )
        assert metric.tags == {}
        assert metric.source == ""
        assert metric.environment == "production"

    def test_metric_data_to_dict(self):
        """Test MetricData to_dict conversion"""
        timestamp = datetime.now()
        metric = MetricData(
            timestamp=timestamp,
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
            tags={"host": "server1"},
            source="system_monitor",
            component="cpu_monitor",
            environment="staging"
        )

        metric_dict = metric.to_dict()
        assert metric_dict["metric_type"] == MetricType.CPU_USAGE.value
        assert metric_dict["value"] == 50.0
        assert metric_dict["source"] == "system_monitor"
        assert metric_dict["component"] == "cpu_monitor"
        assert metric_dict["environment"] == "staging"

    def test_metric_data_with_various_value_types(self):
        """Test MetricData with different value types"""
        timestamp = datetime.now()

        # Integer value
        metric_int = MetricData(timestamp=timestamp, metric_type=MetricType.ERROR_RATE, value=5)
        assert metric_int.value == 5

        # Float value
        metric_float = MetricData(timestamp=timestamp, metric_type=MetricType.MEMORY_USAGE, value=512.5)
        assert metric_float.value == 512.5

        # String value
        metric_str = MetricData(timestamp=timestamp, metric_type=MetricType.AVAILABILITY, value="healthy")
        assert metric_str.value == "healthy"

        # Boolean value
        metric_bool = MetricData(timestamp=timestamp, metric_type=MetricType.SYSTEM_PERFORMANCE, value=True)
        assert metric_bool.value is True

    def test_metric_data_with_metadata(self):
        """Test MetricData with metadata"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.RESPONSE_TIME,
            value=150.0,
            metadata={
                "endpoint": "/api/users",
                "method": "GET",
                "status_code": 200
            }
        )
        assert metric.metadata["endpoint"] == "/api/users"
        assert metric.metadata["method"] == "GET"

    def test_metric_data_with_tenant_id(self):
        """Test MetricData with tenant ID"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=40.0,
            tenant_id="tenant_123"
        )
        assert metric.tenant_id == "tenant_123"

    def test_metric_data_serialization_roundtrip(self):
        """Test MetricData serialization roundtrip"""
        timestamp = datetime.now()
        original = MetricData(
            timestamp=timestamp,
            metric_type=MetricType.THROUGHPUT,
            value=1000,
            tags={"region": "us-west"},
            source="api_gateway"
        )

        # Convert to dict
        data_dict = original.to_dict()

        # Verify all fields are serializable
        import json
        json_str = json.dumps(data_dict, default=str)
        assert json_str is not None


class TestMetricDataEdgeCases:
    """Test edge cases for MetricData"""

    def test_metric_data_with_empty_tags(self):
        """Test MetricData with empty tags"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.MEMORY_USAGE,
            value=100.0,
            tags={}
        )
        assert metric.tags == {}

    def test_metric_data_with_none_tenant_id(self):
        """Test MetricData with None tenant ID"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
            tenant_id=None
        )
        assert metric.tenant_id is None

    def test_metric_data_with_zero_value(self):
        """Test MetricData with zero value"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.ERROR_RATE,
            value=0
        )
        assert metric.value == 0

    def test_metric_data_with_negative_value(self):
        """Test MetricData with negative value (edge case)"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.RESPONSE_TIME,
            value=-10  # Invalid but allowed by dataclass
        )
        assert metric.value == -10

    def test_metric_data_with_very_large_value(self):
        """Test MetricData with very large value"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.THROUGHPUT,
            value=999999999999
        )
        assert metric.value == 999999999999

    def test_metric_data_with_float_precision(self):
        """Test MetricData with float precision"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CACHE_PERFORMANCE,
            value=0.123456789
        )
        metric_dict = metric.to_dict()
        assert abs(metric_dict["value"] - 0.123456789) < 0.0000001

    def test_metric_data_component_field(self):
        """Test MetricData component field"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
            component="core_1"
        )
        assert metric.component == "core_1"

    def test_metric_data_environment_field(self):
        """Test MetricData environment field"""
        environments = ["production", "staging", "development", "test"]
        for env in environments:
            metric = MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=50.0,
                environment=env
            )
            assert metric.environment == env

    def test_metric_data_timestamp_variations(self):
        """Test MetricData with different timestamp values"""
        now = datetime.now()
        past = now - timedelta(days=1)
        future = now + timedelta(days=1)

        for ts in [now, past, future]:
            metric = MetricData(
                timestamp=ts,
                metric_type=MetricType.THROUGHPUT,
                value=100
            )
            assert metric.timestamp == ts

    def test_metric_data_large_metadata(self):
        """Test MetricData with large metadata"""
        large_metadata = {f"key_{i}": f"value_{i}" for i in range(100)}
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.RESPONSE_TIME,
            value=100.0,
            metadata=large_metadata
        )
        assert len(metric.metadata) == 100

    def test_metric_data_special_characters_in_tags(self):
        """Test MetricData with special characters in tags"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.DATABASE_PERFORMANCE,
            value=75.0,
            tags={
                "host": "server-01@dc1",
                "service": "api/v1",
                "user": "test@example.com"
            }
        )
        assert "@" in metric.tags["host"]
        assert "/" in metric.tags["service"]

    def test_metric_data_unicode_in_source(self):
        """Test MetricData with unicode in source"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
            source="监控系统"  # Monitoring system in Chinese
        )
        assert metric.source == "监控系统"

    def test_metric_data_to_dict_preserves_types(self):
        """Test that to_dict preserves type information where possible"""
        timestamp = datetime.now()
        metric = MetricData(
            timestamp=timestamp,
            metric_type=MetricType.MEMORY_USAGE,
            value=512.5,
            tags={"key": "value"},
            metadata={"nested": {"value": 123}}
        )

        metric_dict = metric.to_dict()
        assert isinstance(metric_dict["value"], float)
        assert isinstance(metric_dict["tags"], dict)
        assert isinstance(metric_dict["metadata"], dict)
        assert isinstance(metric_dict["timestamp"], str)

    def test_multiple_metrics_independence(self):
        """Test that multiple metrics are independent"""
        metric1 = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=50.0,
            tags={"host": "server1"}
        )

        metric2 = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.MEMORY_USAGE,
            value=100.0,
            tags={"host": "server2"}
        )

        # Modify metric1 tags
        metric1.tags["host"] = "modified"

        # Verify metric2 is unaffected
        assert metric2.tags["host"] == "server2"

    def test_metric_data_all_metric_types(self):
        """Test MetricData works with all MetricType values"""
        timestamp = datetime.now()
        for metric_type in MetricType:
            metric = MetricData(
                timestamp=timestamp,
                metric_type=metric_type,
                value=100.0
            )
            assert metric.metric_type == metric_type

    def test_metric_data_to_dict_iso_timestamp(self):
        """Test that to_dict produces ISO format timestamp"""
        timestamp = datetime(2024, 1, 15, 10, 30, 45)
        metric = MetricData(
            timestamp=timestamp,
            metric_type=MetricType.CPU_USAGE,
            value=50.0
        )

        metric_dict = metric.to_dict()
        assert "2024-01-15" in metric_dict["timestamp"]
        assert "10:30:45" in metric_dict["timestamp"]
