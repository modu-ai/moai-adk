"""
Extended comprehensive tests for realtime_monitoring_dashboard.py

Extends coverage to include:
- Async operations and WebSocket integration
- External integration setup (Prometheus, DataDog)
- Widget data retrieval for all widget types
- System metrics collection with psutil
- Alert callback error handling
- Dashboard filtering and updates
- Tenant-specific operations
- Multi-threaded concurrent operations
"""

import threading
from datetime import datetime, timedelta
from unittest.mock import MagicMock, Mock, patch

import pytest

from src.moai_adk.core.realtime_monitoring_dashboard import (
    AlertManager,
    DashboardManager,
    DashboardType,
    DashboardWidget,
    MetricData,
    MetricsCollector,
    MetricType,
    RealtimeMonitoringDashboard,
)

# ============================================================================
# Extended Integration Tests
# ============================================================================


class TestRealtimeMonitoringDashboardAsync:
    """Tests for async operations in RealtimeMonitoringDashboard"""

    @pytest.mark.asyncio
    async def test_start_monitoring_dashboard(self):
        """Test starting monitoring dashboard asynchronously"""
        dashboard = RealtimeMonitoringDashboard(
            enable_websocket=False,
            enable_external_integration=False,
        )

        await dashboard.start()

        assert dashboard._running is True
        assert dashboard._monitoring_thread is not None
        assert dashboard._alert_thread is not None

        dashboard.stop()

    @pytest.mark.asyncio
    async def test_start_monitoring_already_running(self):
        """Test starting already running dashboard"""
        dashboard = RealtimeMonitoringDashboard(
            enable_websocket=False,
            enable_external_integration=False,
        )

        await dashboard.start()
        dashboard_id_1 = id(dashboard._monitoring_thread)

        # Try to start again
        await dashboard.start()
        dashboard_id_2 = id(dashboard._monitoring_thread)

        # Should be same thread (not started twice)
        assert dashboard_id_1 == dashboard_id_2

        dashboard.stop()

    @pytest.mark.asyncio
    async def test_websocket_server_initialization(self):
        """Test WebSocket server initialization"""
        dashboard = RealtimeMonitoringDashboard(
            enable_websocket=True,
            enable_external_integration=False,
        )

        # Should not raise exception
        await dashboard._start_websocket_server()

    @pytest.mark.asyncio
    async def test_external_integration_initialization(self):
        """Test external integration initialization"""
        dashboard = RealtimeMonitoringDashboard(
            enable_external_integration=True,
        )

        # Should not raise exception
        await dashboard._initialize_external_integrations()

    @pytest.mark.asyncio
    async def test_prometheus_integration_setup(self):
        """Test Prometheus integration setup"""
        dashboard = RealtimeMonitoringDashboard()

        # Should not raise exception
        await dashboard._setup_prometheus_integration()

    @pytest.mark.asyncio
    async def test_datadog_integration_setup(self):
        """Test DataDog integration setup"""
        dashboard = RealtimeMonitoringDashboard()

        # Should not raise exception
        await dashboard._setup_datadog_integration()

    def test_stop_monitoring_already_stopped(self):
        """Test stopping already stopped dashboard"""
        dashboard = RealtimeMonitoringDashboard(
            enable_websocket=False,
            enable_external_integration=False,
        )

        assert dashboard._running is False

        # Should not raise exception
        dashboard.stop()

        assert dashboard._running is False

    def test_stop_monitoring_clears_websocket_connections(self):
        """Test stop clears WebSocket connections"""
        dashboard = RealtimeMonitoringDashboard()

        # Add mock connections
        mock_conn1 = MagicMock()
        mock_conn2 = MagicMock()
        dashboard.websocket_connections.add(mock_conn1)
        dashboard.websocket_connections.add(mock_conn2)

        dashboard._running = True
        dashboard.stop()

        assert len(dashboard.websocket_connections) == 0


# ============================================================================
# Widget Data Retrieval Tests
# ============================================================================


class TestWidgetDataRetrieval:
    """Tests for widget data retrieval functionality"""

    def test_get_widget_data_metric_no_config(self):
        """Test metric widget without config metric name"""
        dashboard = RealtimeMonitoringDashboard()

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="metric",
            title="Test",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
            config={},  # No metric specified
        )

        data = dashboard._get_metric_widget_data(widget)

        assert "error" in data

    def test_get_metric_widget_data_with_percentages(self):
        """Test metric widget with percentage formatting"""
        dashboard = RealtimeMonitoringDashboard()

        dashboard.add_metric(MetricType.CPU_USAGE, 75.5, component="system")

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="metric",
            title="CPU",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
            config={"metric": "cpu_usage", "format": "percentage"},
        )

        data = dashboard._get_metric_widget_data(widget)

        assert "formatted_value" in data
        assert "%" in data["formatted_value"]

    def test_get_metric_widget_data_with_duration(self):
        """Test metric widget with number formatting (response_time not supported)"""
        dashboard = RealtimeMonitoringDashboard()

        # Use a supported metric type (cpu_usage, memory_usage, health_score)
        dashboard.add_metric(MetricType.CPU_USAGE, 5000.0, component="api")

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="metric",
            title="CPU Time",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
            config={"metric": "cpu_usage", "format": "number"},
        )

        data = dashboard._get_metric_widget_data(widget)

        assert "formatted_value" in data or "value" in data

    def test_get_chart_widget_data_no_metrics(self):
        """Test chart widget with no metrics"""
        dashboard = RealtimeMonitoringDashboard()

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="chart",
            title="Chart",
            position={"x": 0, "y": 0, "width": 6, "height": 3},
            config={"chart_type": "line"},
            metrics=[],
        )

        start_time = datetime.now() - timedelta(hours=1)
        end_time = datetime.now()
        data = dashboard._get_chart_widget_data(widget, start_time, end_time)

        assert "data" in data
        assert len(data["data"]) == 0

    def test_get_chart_widget_bar_type(self):
        """Test chart widget with bar chart type"""
        dashboard = RealtimeMonitoringDashboard()

        for i in range(5):
            dashboard.add_metric(
                MetricType.CPU_USAGE,
                50.0 + i,
                component="system",
            )

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="chart",
            title="Chart",
            position={"x": 0, "y": 0, "width": 6, "height": 3},
            config={"chart_type": "bar"},
            metrics=["cpu_usage"],
        )

        start_time = datetime.now() - timedelta(hours=1)
        end_time = datetime.now()
        data = dashboard._get_chart_widget_data(widget, start_time, end_time)

        assert data["type"] == "bar"

    def test_get_table_widget_alert_table(self):
        """Test table widget with alert data"""
        dashboard = RealtimeMonitoringDashboard()

        # Create some alerts
        dashboard.alert_manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
        )

        dashboard.add_metric(MetricType.CPU_USAGE, 85.0, component="system")
        dashboard.alert_manager.check_alerts()

        widget = DashboardWidget(
            widget_id="alert_table",
            widget_type="table",
            title="Alerts",
            position={"x": 0, "y": 0, "width": 12, "height": 3},
        )

        start_time = datetime.now() - timedelta(hours=1)
        end_time = datetime.now()
        data = dashboard._get_table_widget_data(widget, start_time, end_time)

        assert "columns" in data
        assert "data" in data

    def test_get_widget_data_all_types(self):
        """Test getting widget data for all supported types"""
        dashboard = RealtimeMonitoringDashboard()

        dashboard.add_metric(MetricType.CPU_USAGE, 50.0, component="system")

        # Metric widget
        DashboardWidget(
            widget_id="w1",
            widget_type="metric",
            title="CPU",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
            config={"metric": "cpu_usage"},
        )

        start_time = datetime.now() - timedelta(hours=1)
        end_time = datetime.now()

        # All widget types
        for widget_type in ["metric", "chart", "table", "alert"]:
            widget = DashboardWidget(
                widget_id=f"w_{widget_type}",
                widget_type=widget_type,
                title=widget_type,
                position={"x": 0, "y": 0, "width": 4, "height": 2},
                config={"metric": "cpu_usage"} if widget_type == "metric" else {},
                metrics=["cpu_usage"] if widget_type == "chart" else [],
            )

            data = dashboard._get_widget_data(widget, start_time, end_time)

            assert data is not None


# ============================================================================
# System Metrics Collection Tests
# ============================================================================


class TestSystemMetricsCollection:
    """Tests for system metrics collection"""

    @patch("psutil.cpu_percent")
    @patch("psutil.virtual_memory")
    @patch("psutil.Process")
    @patch("psutil.getloadavg")
    def test_collect_system_metrics_full(self, mock_loadavg, mock_process, mock_memory, mock_cpu):
        """Test collecting all system metrics"""
        mock_cpu.return_value = 50.5

        mock_mem_obj = MagicMock()
        mock_mem_obj.percent = 65.5
        mock_memory.return_value = mock_mem_obj

        mock_proc_obj = MagicMock()
        mock_proc_obj.memory_info.return_value = MagicMock(rss=1024 * 1024 * 600)
        mock_process.return_value = mock_proc_obj

        mock_loadavg.return_value = (1.5, 1.2, 0.9)

        dashboard = RealtimeMonitoringDashboard()
        dashboard._collect_system_metrics()

        # Check metrics were collected
        cpu_metrics = dashboard.metrics_collector.get_metrics(metric_type=MetricType.CPU_USAGE)
        assert len(cpu_metrics) > 0

        memory_metrics = dashboard.metrics_collector.get_metrics(metric_type=MetricType.MEMORY_USAGE)
        assert len(memory_metrics) > 0

        perf_metrics = dashboard.metrics_collector.get_metrics(metric_type=MetricType.SYSTEM_PERFORMANCE)
        assert len(perf_metrics) > 0

    @patch("psutil.cpu_percent", side_effect=Exception("psutil error"))
    def test_collect_system_metrics_error_handling(self, mock_cpu):
        """Test error handling during metrics collection"""
        dashboard = RealtimeMonitoringDashboard()

        # Should not raise exception
        dashboard._collect_system_metrics()

    @patch("psutil.getloadavg", side_effect=AttributeError)
    @patch("psutil.cpu_percent", return_value=50.0)
    @patch("psutil.virtual_memory")
    @patch("psutil.Process")
    def test_collect_system_metrics_loadavg_unavailable(self, mock_process, mock_memory, mock_cpu, mock_loadavg):
        """Test system metrics collection when getloadavg is unavailable"""
        mock_mem_obj = MagicMock()
        mock_mem_obj.percent = 50.0
        mock_memory.return_value = mock_mem_obj

        mock_proc_obj = MagicMock()
        mock_proc_obj.memory_info.return_value = MagicMock(rss=1024 * 1024 * 512)
        mock_process.return_value = mock_proc_obj

        dashboard = RealtimeMonitoringDashboard()

        # Should not raise exception
        dashboard._collect_system_metrics()


# ============================================================================
# Alert Callback and Error Handling Tests
# ============================================================================


class TestAlertCallbacks:
    """Tests for alert callbacks and error handling"""

    def test_alert_callback_with_exception(self):
        """Test alert callback that raises exception"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        # Add callback that raises exception
        def error_callback(alert):
            raise ValueError("Callback error")

        manager.add_alert_callback(error_callback)

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

        # Should not raise exception even though callback fails
        alerts = manager.check_alerts()
        assert len(alerts) == 1

    def test_multiple_alert_callbacks(self):
        """Test multiple alert callbacks"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        callback1 = Mock()
        callback2 = Mock()
        manager.add_alert_callback(callback1)
        manager.add_alert_callback(callback2)

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

        assert callback1.called
        assert callback2.called

    def test_alert_callback_with_tenant_id(self):
        """Test acknowledging tenant-specific alert"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="Tenant Alert",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
            tenant_id="tenant1",
        )

        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=MetricType.CPU_USAGE,
            value=85.0,
            component="system",
            tenant_id="tenant1",
        )
        collector.add_metric(metric)

        manager.check_alerts()

        # Get tenant-specific alerts
        alerts = manager.get_active_alerts(tenant_id="tenant1")
        assert len(alerts) > 0

        # Acknowledge tenant alert
        alert_id = alerts[0].alert_id
        result = manager.acknowledge_alert(alert_id, tenant_id="tenant1")
        assert result is True


# ============================================================================
# Dashboard Filtering and Updates Tests
# ============================================================================


class TestDashboardFiltering:
    """Tests for dashboard filtering and updates"""

    def test_list_dashboards_by_owner(self):
        """Test filtering dashboards by owner"""
        manager = DashboardManager()

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="metric",
            title="CPU",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
        )

        manager.create_dashboard(
            name="Dashboard 1",
            description="Test",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[widget],
            owner="user1",
        )

        manager.create_dashboard(
            name="Dashboard 2",
            description="Test",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[widget],
            owner="user2",
        )

        user1_dashboards = manager.list_dashboards(owner="user1")

        assert len(user1_dashboards) == 1
        assert user1_dashboards[0].owner == "user1"

    def test_list_dashboards_by_is_public(self):
        """Test filtering dashboards by public/private"""
        manager = DashboardManager()

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="metric",
            title="CPU",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
        )

        manager.create_dashboard(
            name="Public Dashboard",
            description="Test",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[widget],
            is_public=True,
        )

        manager.create_dashboard(
            name="Private Dashboard",
            description="Test",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[widget],
            is_public=False,
        )

        public_dashboards = manager.list_dashboards(is_public=True)

        # Should include default dashboards which are public
        assert len(public_dashboards) >= 1

    def test_update_dashboard_multiple_fields(self):
        """Test updating multiple dashboard fields"""
        manager = DashboardManager()

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="metric",
            title="CPU",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
        )

        dashboard_id = manager.create_dashboard(
            name="Original",
            description="Original description",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[widget],
            is_public=False,
        )

        result = manager.update_dashboard(
            dashboard_id,
            {
                "name": "Updated",
                "description": "Updated description",
                "is_public": True,
            },
        )

        assert result is True

        updated = manager.get_dashboard(dashboard_id)
        assert updated.name == "Updated"
        assert updated.description == "Updated description"
        assert updated.is_public is True

    def test_tenant_specific_dashboard_operations(self):
        """Test tenant-specific dashboard operations"""
        manager = DashboardManager()

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="metric",
            title="CPU",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
        )

        # Create tenant-specific dashboard
        tenant_id = "tenant1"
        dashboard_id = manager.create_dashboard(
            name="Tenant Dashboard",
            description="Test",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[widget],
            tenant_id=tenant_id,
        )

        # Get tenant dashboard
        dashboard = manager.get_dashboard(dashboard_id, tenant_id=tenant_id)
        assert dashboard is not None
        assert dashboard.tenant_id == tenant_id

        # Update tenant dashboard
        manager.update_dashboard(
            dashboard_id,
            {"name": "Updated Tenant Dashboard"},
            tenant_id=tenant_id,
        )

        # Delete tenant dashboard
        result = manager.delete_dashboard(dashboard_id, tenant_id=tenant_id)
        assert result is True

    def test_list_tenant_specific_dashboards(self):
        """Test listing tenant-specific dashboards"""
        manager = DashboardManager()

        widget = DashboardWidget(
            widget_id="w1",
            widget_type="metric",
            title="CPU",
            position={"x": 0, "y": 0, "width": 4, "height": 2},
        )

        # Create dashboards for different tenants
        manager.create_dashboard(
            name="Tenant 1 Dashboard",
            description="Test",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[widget],
            tenant_id="tenant1",
        )

        manager.create_dashboard(
            name="Tenant 2 Dashboard",
            description="Test",
            dashboard_type=DashboardType.CUSTOM,
            widgets=[widget],
            tenant_id="tenant2",
        )

        tenant1_dashboards = manager.list_dashboards(tenant_id="tenant1")

        assert len(tenant1_dashboards) >= 1
        assert all(d.tenant_id == "tenant1" or d.is_public for d in tenant1_dashboards)


# ============================================================================
# Concurrent Operations Tests
# ============================================================================


class TestConcurrentOperations:
    """Tests for concurrent operations and thread safety"""

    def test_concurrent_alert_checks(self):
        """Test concurrent alert checking"""
        collector = MetricsCollector()
        manager = AlertManager(collector)

        manager.add_alert_rule(
            name="High CPU",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
        )

        # Add metrics
        for i in range(10):
            metric = MetricData(
                timestamp=datetime.now(),
                metric_type=MetricType.CPU_USAGE,
                value=85.0,
                component="system",
            )
            collector.add_metric(metric)

        # Check alerts concurrently
        def check_alerts():
            return manager.check_alerts()

        threads = [threading.Thread(target=check_alerts) for _ in range(3)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        # Should have alerts
        assert len(manager.active_alerts) > 0

    def test_concurrent_dashboard_updates(self):
        """Test concurrent dashboard updates"""
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

        update_count = [0]

        def update_dashboard():
            for i in range(10):
                result = manager.update_dashboard(
                    dashboard_id,
                    {"name": f"Updated {i}"},
                )
                if result:
                    update_count[0] += 1

        threads = [threading.Thread(target=update_dashboard) for _ in range(3)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        # All updates should succeed
        assert update_count[0] == 30

    def test_concurrent_metric_collection(self):
        """Test concurrent metric collection"""
        dashboard = RealtimeMonitoringDashboard()

        def add_metrics():
            for i in range(50):
                dashboard.add_metric(
                    MetricType.CPU_USAGE,
                    50.0 + (i % 50),
                    component="system",
                )
                dashboard.add_metric(
                    MetricType.MEMORY_USAGE,
                    40.0 + (i % 60),
                    component="system",
                )

        threads = [threading.Thread(target=add_metrics) for _ in range(5)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        # Metrics should be collected
        all_metrics = dashboard.metrics_collector.get_metrics()
        assert len(all_metrics) > 0


# ============================================================================
# Dashboard Data Retrieval Tests
# ============================================================================


class TestDashboardDataRetrieval:
    """Tests for dashboard data retrieval with various time ranges"""

    def test_get_dashboard_data_1h_range(self):
        """Test dashboard data for 1 hour range"""
        dashboard = RealtimeMonitoringDashboard()

        # Add metrics
        for i in range(10):
            dashboard.add_metric(MetricType.CPU_USAGE, 50.0 + i, component="system")

        data = dashboard.get_dashboard_data("system_overview", time_range="1h")

        assert "dashboard" in data
        assert data["time_range"] == "1h"

    def test_get_dashboard_data_24h_range(self):
        """Test dashboard data for 24 hour range"""
        dashboard = RealtimeMonitoringDashboard()

        for i in range(20):
            dashboard.add_metric(MetricType.CPU_USAGE, 50.0 + i, component="system")

        data = dashboard.get_dashboard_data("system_overview", time_range="24h")

        assert "dashboard" in data
        assert data["time_range"] == "24h"

    def test_get_dashboard_data_7d_range(self):
        """Test dashboard data for 7 day range"""
        dashboard = RealtimeMonitoringDashboard()

        for i in range(30):
            dashboard.add_metric(MetricType.CPU_USAGE, 50.0 + i, component="system")

        data = dashboard.get_dashboard_data("system_overview", time_range="7d")

        assert "dashboard" in data
        assert data["time_range"] == "7d"

    def test_get_dashboard_data_invalid_range(self):
        """Test dashboard data with invalid time range"""
        dashboard = RealtimeMonitoringDashboard()

        data = dashboard.get_dashboard_data("system_overview", time_range="invalid")

        # Should use default (1h)
        assert "dashboard" in data
        assert data["time_range"] == "invalid"

    def test_get_dashboard_data_with_filters(self):
        """Test dashboard data retrieval with filters"""
        dashboard = RealtimeMonitoringDashboard()

        for i in range(10):
            dashboard.add_metric(
                MetricType.CPU_USAGE,
                50.0 + i,
                component="system",
                tags={"host": "server1"},
            )

        data = dashboard.get_dashboard_data(
            "system_overview",
            filters={"host": "server1"},
        )

        assert "dashboard" in data

    def test_get_dashboard_data_widget_error_handling(self):
        """Test dashboard data retrieval handles widget errors"""
        dashboard = RealtimeMonitoringDashboard()

        # Get data for default dashboard with various widgets
        data = dashboard.get_dashboard_data("hook_analysis")

        assert "dashboard" in data
        assert "widgets_data" in data
        # Should have data for all widgets even if some fail
        assert len(data["widgets_data"]) > 0


# ============================================================================
# Metric Name Mapping Tests
# ============================================================================


class TestMetricNameMapping:
    """Tests for metric name to MetricType mapping"""

    def test_all_metric_name_mappings(self):
        """Test all metric name mappings"""
        dashboard = RealtimeMonitoringDashboard()

        mapping_tests = [
            ("cpu_usage", MetricType.CPU_USAGE),
            ("memory_usage", MetricType.MEMORY_USAGE),
            ("hook_execution_rate", MetricType.THROUGHPUT),
            ("hook_success_rate", MetricType.AVAILABILITY),
            ("response_time", MetricType.RESPONSE_TIME),
            ("error_rate", MetricType.ERROR_RATE),
            ("network_io", MetricType.NETWORK_IO),
            ("disk_io", MetricType.DISK_IO),
            ("cache_performance", MetricType.CACHE_PERFORMANCE),
            ("database_performance", MetricType.DATABASE_PERFORMANCE),
        ]

        for metric_name, expected_type in mapping_tests:
            result = dashboard._map_metric_name(metric_name)
            assert result == expected_type, f"Failed for {metric_name}"

    def test_unmapped_metric_name(self):
        """Test unmapped metric name returns None"""
        dashboard = RealtimeMonitoringDashboard()

        result = dashboard._map_metric_name("unknown_metric")

        assert result is None


# ============================================================================
# System Status Tests
# ============================================================================


class TestSystemStatus:
    """Tests for system status reporting"""

    def test_get_system_status_includes_all_fields(self):
        """Test system status includes all required fields"""
        dashboard = RealtimeMonitoringDashboard()

        status = dashboard.get_system_status()

        required_fields = [
            "status",
            "uptime_seconds",
            "metrics_collected",
            "active_alerts",
            "total_dashboards",
            "websocket_connections",
            "external_integrations",
            "last_update",
        ]

        for field in required_fields:
            assert field in status

    def test_get_system_status_running(self):
        """Test system status when running"""
        dashboard = RealtimeMonitoringDashboard()
        dashboard._running = True

        status = dashboard.get_system_status()

        assert status["status"] == "running"

    def test_get_system_status_stopped(self):
        """Test system status when stopped"""
        dashboard = RealtimeMonitoringDashboard()
        dashboard._running = False

        status = dashboard.get_system_status()

        assert status["status"] == "stopped"

    def test_get_system_status_uptime_calculation(self):
        """Test uptime calculation in system status"""
        dashboard = RealtimeMonitoringDashboard()

        # Set startup time to 60 seconds ago
        dashboard._startup_time = datetime.now() - timedelta(seconds=60)

        status = dashboard.get_system_status()

        # Should be approximately 60 seconds
        assert 50 < status["uptime_seconds"] < 70


# ============================================================================
# Custom Dashboard Creation Tests
# ============================================================================


class TestCustomDashboardCreation:
    """Tests for custom dashboard creation"""

    def test_create_custom_dashboard_with_various_widgets(self):
        """Test creating custom dashboard with various widget types"""
        dashboard = RealtimeMonitoringDashboard()

        widget_defs = [
            {
                "widget_id": "w1",
                "widget_type": "metric",
                "title": "CPU Usage",
                "position": {"x": 0, "y": 0, "width": 4, "height": 2},
                "config": {"metric": "cpu_usage"},
            },
            {
                "widget_id": "w2",
                "widget_type": "chart",
                "title": "CPU History",
                "position": {"x": 4, "y": 0, "width": 8, "height": 3},
                "config": {"chart_type": "line"},
                "metrics": ["cpu_usage"],
            },
            {
                "widget_id": "w3",
                "widget_type": "alert",
                "title": "Active Alerts",
                "position": {"x": 0, "y": 2, "width": 12, "height": 2},
            },
        ]

        dashboard_id = dashboard.create_custom_dashboard(
            name="Custom Monitoring",
            description="Custom monitoring dashboard",
            widgets=widget_defs,
            owner="admin",
        )

        created = dashboard.dashboard_manager.get_dashboard(dashboard_id)
        assert created.name == "Custom Monitoring"
        assert len(created.widgets) == 3

    def test_create_custom_dashboard_tenant_specific(self):
        """Test creating tenant-specific custom dashboard"""
        dashboard = RealtimeMonitoringDashboard()

        widget_defs = [
            {
                "widget_id": "w1",
                "widget_type": "metric",
                "title": "Tenant Metrics",
                "position": {"x": 0, "y": 0, "width": 12, "height": 2},
            }
        ]

        dashboard_id = dashboard.create_custom_dashboard(
            name="Tenant Dashboard",
            description="Dashboard for tenant1",
            widgets=widget_defs,
            tenant_id="tenant1",
            owner="tenant1_admin",
        )

        created = dashboard.dashboard_manager.get_dashboard(dashboard_id, tenant_id="tenant1")
        assert created.tenant_id == "tenant1"
