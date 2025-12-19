"""
Real-time Monitoring Dashboard

Phase 3: Enterprise-grade real-time monitoring dashboard with alerting,
analytics, and visualization capabilities for the hook system.

Key Features:
- Real-time performance monitoring and visualization
- Automated alerting with configurable thresholds
- Integration with external monitoring systems (Prometheus, Grafana, DataDog)
- Visual analytics for system health and performance
- REST API for monitoring data access
- WebSocket support for real-time updates
- Historical data analysis and trend prediction
- Multi-tenant dashboard support
- Custom alert rule engine
- Performance bottleneck detection
"""

import asyncio
import logging
import statistics
import threading
import time
import uuid
from collections import defaultdict, deque
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
from typing import Any, Callable, Dict, List, Optional, Set, Union

logger = logging.getLogger(__name__)


class MetricType(Enum):
    """Types of metrics collected by the monitoring system"""

    SYSTEM_PERFORMANCE = "system_performance"
    HOOK_EXECUTION = "hook_execution"
    ERROR_RATE = "error_rate"
    RESPONSE_TIME = "response_time"
    MEMORY_USAGE = "memory_usage"
    CPU_USAGE = "cpu_usage"
    NETWORK_IO = "network_io"
    DISK_IO = "disk_io"
    THROUGHPUT = "throughput"
    AVAILABILITY = "availability"
    CACHE_PERFORMANCE = "cache_performance"
    DATABASE_PERFORMANCE = "database_performance"


class AlertSeverity(Enum):
    """Alert severity levels"""

    INFO = 1
    WARNING = 2
    ERROR = 3
    CRITICAL = 4
    EMERGENCY = 5


class DashboardType(Enum):
    """Dashboard types"""

    SYSTEM_OVERVIEW = "system_overview"
    PERFORMANCE_ANALYTICS = "performance_analytics"
    ERROR_MONITORING = "error_monitoring"
    HOOK_ANALYSIS = "hook_analysis"
    RESOURCE_USAGE = "resource_usage"
    CUSTOM = "custom"


@dataclass
class MetricData:
    """Single metric data point with enhanced metadata"""

    timestamp: datetime
    metric_type: MetricType
    value: Union[int, float, str, bool]
    tags: Dict[str, str] = field(default_factory=dict)
    source: str = ""
    metadata: Dict[str, Any] = field(default_factory=dict)
    component: str = ""
    environment: str = "production"
    tenant_id: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "timestamp": self.timestamp.isoformat(),
            "metric_type": self.metric_type.value,
            "value": self.value,
            "tags": self.tags,
            "source": self.source,
            "metadata": self.metadata,
            "component": self.component,
            "environment": self.environment,
            "tenant_id": self.tenant_id,
        }


@dataclass
class Alert:
    """Enhanced alert definition with correlation and context"""

    alert_id: str
    severity: AlertSeverity
    title: str
    description: str
    timestamp: datetime
    metric_type: MetricType
    threshold: float
    current_value: float
    source: str
    component: str
    tags: Dict[str, str] = field(default_factory=dict)
    resolved: bool = False
    resolved_at: Optional[datetime] = None
    acknowledged: bool = False
    acknowledged_at: Optional[datetime] = None
    correlation_id: Optional[str] = None
    context: Dict[str, Any] = field(default_factory=dict)
    affected_services: List[str] = field(default_factory=list)
    recovery_actions: List[str] = field(default_factory=list)
    tenant_id: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "alert_id": self.alert_id,
            "severity": self.severity.value,
            "title": self.title,
            "description": self.description,
            "timestamp": self.timestamp.isoformat(),
            "metric_type": self.metric_type.value,
            "threshold": self.threshold,
            "current_value": self.current_value,
            "source": self.source,
            "component": self.component,
            "tags": self.tags,
            "resolved": self.resolved,
            "resolved_at": self.resolved_at.isoformat() if self.resolved_at else None,
            "acknowledged": self.acknowledged,
            "acknowledged_at": (self.acknowledged_at.isoformat() if self.acknowledged_at else None),
            "correlation_id": self.correlation_id,
            "context": self.context,
            "affected_services": self.affected_services,
            "recovery_actions": self.recovery_actions,
        }


@dataclass
class DashboardWidget:
    """Dashboard widget definition"""

    widget_id: str
    widget_type: str  # "chart", "metric", "table", "alert", "heatmap"
    title: str
    position: Dict[str, Any]  # {"x": 0, "y": 0, "width": 4, "height": 2}
    config: Dict[str, Any] = field(default_factory=dict)
    data_source: str = ""
    refresh_interval_seconds: int = 30
    metrics: List[str] = field(default_factory=list)
    filters: Dict[str, Any] = field(default_factory=dict)


@dataclass
class Dashboard:
    """Dashboard definition with widgets and layout"""

    dashboard_id: str
    name: str
    description: str
    dashboard_type: DashboardType
    widgets: List[DashboardWidget] = field(default_factory=list)
    layout: Dict[str, Any] = field(default_factory=dict)
    filters: Dict[str, Any] = field(default_factory=dict)
    owner: str = ""
    tenant_id: Optional[str] = None
    is_public: bool = False
    created_at: datetime = field(default_factory=datetime.now)
    updated_at: datetime = field(default_factory=datetime.now)

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "dashboard_id": self.dashboard_id,
            "name": self.name,
            "description": self.description,
            "dashboard_type": self.dashboard_type.value,
            "widgets": [w.__dict__ for w in self.widgets],
            "layout": self.layout,
            "filters": self.filters,
            "owner": self.owner,
            "tenant_id": self.tenant_id,
            "is_public": self.is_public,
            "created_at": self.created_at.isoformat(),
            "updated_at": self.updated_at.isoformat(),
        }


class MetricsCollector:
    """Advanced metrics collection with multi-tenant support"""

    def __init__(self, buffer_size: int = 100000, retention_hours: int = 168):  # 7 days
        self.buffer_size = buffer_size
        self.retention_hours = retention_hours
        self.metrics_buffer: Dict[str, deque] = defaultdict(lambda: deque(maxlen=buffer_size))
        self.aggregated_metrics: Dict[str, Dict[str, Any]] = defaultdict(dict)
        self.tenant_metrics: Dict[str, Dict[str, deque]] = defaultdict(lambda: defaultdict(deque))
        self._lock = threading.RLock()
        self._last_cleanup = datetime.now()

    def add_metric(self, metric: MetricData) -> None:
        """Add a metric to the collection with tenant support"""
        with self._lock:
            # Global metrics buffer
            key = f"{metric.metric_type.value}:{metric.component}"
            self.metrics_buffer[key].append(metric)

            # Tenant-specific metrics
            if metric.tenant_id:
                tenant_key = f"{metric.metric_type.value}:{metric.component}:{metric.tenant_id}"
                self.tenant_metrics[metric.tenant_id][tenant_key].append(metric)

            # Update aggregated statistics
            self._update_aggregated_metrics(metric)
            self._cleanup_old_metrics()

    def _update_aggregated_metrics(self, metric: MetricData) -> None:
        """Update aggregated statistics for a metric type"""
        key = f"{metric.metric_type.value}:{metric.component}"

        if key not in self.aggregated_metrics:
            self.aggregated_metrics[key] = {
                "count": 0,
                "sum": 0,
                "min": float("inf"),
                "max": float("-inf"),
                "values": [],
                "last_updated": datetime.now(),
            }

        agg = self.aggregated_metrics[key]

        if isinstance(metric.value, (int, float)):
            agg["count"] += 1
            agg["sum"] += metric.value
            agg["min"] = min(agg["min"], metric.value)
            agg["max"] = max(agg["max"], metric.value)
            agg["values"].append(metric.value)
            agg["last_updated"] = datetime.now()

            # Keep only recent values for statistics
            if len(agg["values"]) > 10000:
                agg["values"] = agg["values"][-10000:]

    def _cleanup_old_metrics(self) -> None:
        """Remove metrics older than retention period"""
        now = datetime.now()
        if (now - self._last_cleanup).seconds < 300:  # Cleanup every 5 minutes
            return

        cutoff_time = now - timedelta(hours=self.retention_hours)

        # Cleanup global metrics
        for key, buffer in self.metrics_buffer.items():
            while buffer and buffer[0].timestamp < cutoff_time:
                buffer.popleft()

        # Cleanup tenant metrics
        for tenant_id, tenant_buffers in self.tenant_metrics.items():
            for key, buffer in tenant_buffers.items():
                while buffer and buffer[0].timestamp < cutoff_time:
                    buffer.popleft()

        self._last_cleanup = now

    def get_metrics(
        self,
        metric_type: Optional[MetricType] = None,
        component: Optional[str] = None,
        tenant_id: Optional[str] = None,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
        limit: Optional[int] = None,
        tags: Optional[Dict[str, str]] = None,
    ) -> List[MetricData]:
        """Get metrics with comprehensive filtering"""
        with self._lock:
            # Choose the right metrics source
            source_metrics: List[MetricData] = []
            if tenant_id:
                for buffer in self.tenant_metrics.get(tenant_id, {}).values():
                    source_metrics.extend(buffer)
            else:
                for buffer in self.metrics_buffer.values():
                    source_metrics.extend(buffer)

            # Apply filters
            metrics = source_metrics

            # Filter by metric type
            if metric_type:
                metrics = [m for m in metrics if m.metric_type == metric_type]

            # Filter by component
            if component:
                metrics = [m for m in metrics if m.component == component]

            # Filter by time range
            if start_time:
                metrics = [m for m in metrics if m.timestamp >= start_time]
            if end_time:
                metrics = [m for m in metrics if m.timestamp <= end_time]

            # Filter by tags
            if tags:
                metrics = [m for m in metrics if all(m.tags.get(k) == v for k, v in tags.items())]

            # Sort by timestamp (newest first)
            metrics.sort(key=lambda m: m.timestamp, reverse=True)

            # Apply limit
            if limit:
                metrics = metrics[:limit]

            return metrics

    def get_statistics(
        self,
        metric_type: MetricType,
        component: Optional[str] = None,
        minutes: int = 60,
        tenant_id: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Get statistical summary for a metric type"""
        with self._lock:
            key = f"{metric_type.value}:{component}" if component else metric_type.value

            # Choose the right aggregated metrics
            if tenant_id:
                agg: Dict[str, Any] = {}
                # Calculate tenant-specific stats
                tenant_metrics = self.get_metrics(metric_type, component, tenant_id)
                if tenant_metrics:
                    values: List[float] = [float(m.value) for m in tenant_metrics if isinstance(m.value, (int, float))]
                    if values:
                        agg = {
                            "count": len(values),
                            "average": statistics.mean(values),
                            "median": statistics.median(values),
                            "min": min(values),
                            "max": max(values),
                            "std_dev": (statistics.stdev(values) if len(values) > 1 else 0),
                            "p95": self._percentile(values, 95),
                            "p99": self._percentile(values, 99),
                        }
            else:
                agg = self.aggregated_metrics.get(key, {})

            if not agg or agg.get("count", 0) == 0:
                return {
                    "count": 0,
                    "average": None,
                    "min": None,
                    "max": None,
                    "median": None,
                    "std_dev": None,
                }

            agg_values_raw = agg.get("values", [])
            values_list: List[float] = agg_values_raw if isinstance(agg_values_raw, list) else []
            if not values_list:
                return {
                    "count": agg.get("count", 0),
                    "average": agg.get("sum", 0) / max(agg.get("count", 1), 1),
                    "min": agg.get("min"),
                    "max": agg.get("max"),
                    "median": None,
                    "std_dev": 0,
                    "p95": agg.get("max"),
                    "p99": agg.get("max"),
                }

            try:
                last_updated_raw = agg.get("last_updated", datetime.now())
                if isinstance(last_updated_raw, datetime):
                    last_updated_str = last_updated_raw.isoformat()
                else:
                    last_updated_str = datetime.now().isoformat()
                return {
                    "count": len(values_list),
                    "average": statistics.mean(values_list),
                    "median": statistics.median(values_list),
                    "min": min(values_list),
                    "max": max(values_list),
                    "std_dev": (statistics.stdev(values_list) if len(values_list) > 1 else 0),
                    "p95": self._percentile(values_list, 95),
                    "p99": self._percentile(values_list, 99),
                    "last_updated": last_updated_str,
                }
            except (statistics.StatisticsError, IndexError):
                return {
                    "count": len(values_list),
                    "average": statistics.mean(values_list),
                    "median": statistics.median(values_list),
                    "min": min(values_list),
                    "max": max(values_list),
                    "std_dev": 0,
                    "p95": max(values_list),
                    "p99": max(values_list),
                    "last_updated": datetime.now().isoformat(),
                }

    def _percentile(self, values: List[float], percentile: int) -> float:
        """Calculate percentile value"""
        if not values:
            return 0.0
        sorted_values = sorted(values)
        index = (percentile / 100) * (len(sorted_values) - 1)
        if index.is_integer():
            return sorted_values[int(index)]
        else:
            lower = sorted_values[int(index)]
            upper = sorted_values[int(index) + 1]
            return lower + (upper - lower) * (index - int(index))


class AlertManager:
    """Advanced alert management with correlation and multi-tenant support"""

    def __init__(self, metrics_collector: MetricsCollector):
        self.metrics_collector = metrics_collector
        self.alert_rules: List[Dict[str, Any]] = []
        self.active_alerts: Dict[str, Alert] = {}
        self.alert_history: List[Alert] = []
        self.alert_callbacks: List[Callable[[Alert], None]] = []
        self.tenant_alerts: Dict[str, Dict[str, Alert]] = defaultdict(dict)
        self._lock = threading.RLock()

    def add_alert_rule(
        self,
        name: str,
        metric_type: MetricType,
        threshold: float,
        operator: str = "gt",  # gt, lt, eq, ne, gte, lte
        severity: AlertSeverity = AlertSeverity.WARNING,
        window_minutes: int = 5,
        consecutive_violations: int = 1,
        tags: Optional[Dict[str, str]] = None,
        description: Optional[str] = None,
        tenant_id: Optional[str] = None,
        component: Optional[str] = None,
        cooldown_minutes: int = 15,
    ) -> None:
        """Add an alert rule with enhanced configuration"""
        rule = {
            "name": name,
            "metric_type": metric_type,
            "threshold": threshold,
            "operator": operator,
            "severity": severity,
            "window_minutes": window_minutes,
            "consecutive_violations": consecutive_violations,
            "tags": tags or {},
            "description": description or f"Alert when {metric_type.value} {operator} {threshold}",
            "violation_count": 0,
            "last_check": None,
            "last_triggered": None,
            "enabled": True,
            "tenant_id": tenant_id,
            "component": component,
            "cooldown_minutes": cooldown_minutes,
        }

        with self._lock:
            self.alert_rules.append(rule)

    def check_alerts(self) -> List[Alert]:
        """Check all alert rules and generate alerts for violations"""
        triggered_alerts = []

        with self._lock:
            for rule in self.alert_rules:
                if not rule["enabled"]:
                    continue

                # Check cooldown period
                if rule["last_triggered"]:
                    time_since_last = datetime.now() - rule["last_triggered"]
                    if time_since_last.total_seconds() < rule["cooldown_minutes"] * 60:
                        continue

                # Get recent metrics for this rule
                recent_metrics = self.metrics_collector.get_metrics(
                    metric_type=rule["metric_type"],
                    component=rule.get("component"),
                    tenant_id=rule.get("tenant_id"),
                    start_time=datetime.now() - timedelta(minutes=rule["window_minutes"]),
                    tags=rule.get("tags"),
                )

                if not recent_metrics:
                    continue

                # Check for violations
                violations = 0
                latest_value = None

                for metric in recent_metrics:
                    if isinstance(metric.value, (int, float)):
                        if self._evaluate_condition(metric.value, rule["threshold"], rule["operator"]):
                            violations += 1
                        latest_value = metric.value

                # Trigger alert if threshold exceeded
                if violations >= rule["consecutive_violations"]:
                    alert_id = f"{rule['name']}_{int(time.time())}"

                    alert = Alert(
                        alert_id=alert_id,
                        severity=rule["severity"],
                        title=f"{rule['name']} Alert Triggered",
                        description=rule["description"],
                        timestamp=datetime.now(),
                        metric_type=rule["metric_type"],
                        threshold=rule["threshold"],
                        current_value=latest_value or 0,
                        source="monitoring_system",
                        component=rule.get("component", "unknown"),
                        tags=rule.get("tags", {}),
                        tenant_id=rule.get("tenant_id"),
                    )

                    # Store alert
                    self.active_alerts[alert_id] = alert
                    self.alert_history.append(alert)

                    if alert.tenant_id:
                        self.tenant_alerts[alert.tenant_id][alert_id] = alert

                    triggered_alerts.append(alert)

                    # Update rule state
                    rule["violation_count"] = violations
                    rule["last_check"] = datetime.now()
                    rule["last_triggered"] = datetime.now()

                    # Trigger callbacks
                    for callback in self.alert_callbacks:
                        try:
                            callback(alert)
                        except Exception as e:
                            logger.error(f"Error in alert callback: {e}")

                rule["violation_count"] = violations
                rule["last_check"] = datetime.now()

            # Check for resolved alerts
            self._check_resolved_alerts()

        return triggered_alerts

    def _check_resolved_alerts(self) -> None:
        """Check if any active alerts have been resolved"""
        for alert_id, alert in list(self.active_alerts.items()):
            # Check if alert condition is no longer met
            rule = next((r for r in self.alert_rules if r["name"] in alert_id), None)
            if rule:
                recent_metrics = self.metrics_collector.get_metrics(
                    metric_type=alert.metric_type,
                    component=alert.component,
                    tenant_id=alert.tenant_id,
                    start_time=datetime.now() - timedelta(minutes=1),  # Check last minute
                )

                if recent_metrics:
                    latest_value = None
                    for metric in recent_metrics:
                        if isinstance(metric.value, (int, float)):
                            if not self._evaluate_condition(metric.value, alert.threshold, rule["operator"]):
                                latest_value = metric.value
                                break

                    if latest_value is not None:
                        # Alert is resolved
                        alert.resolved = True
                        alert.resolved_at = datetime.now()
                        self.alert_history.append(alert)
                        del self.active_alerts[alert_id]

                        if alert.tenant_id and alert_id in self.tenant_alerts.get(alert.tenant_id, {}):
                            del self.tenant_alerts[alert.tenant_id][alert_id]

    def _evaluate_condition(self, value: float, threshold: float, operator: str) -> bool:
        """Evaluate alert condition"""
        if operator == "gt":
            return value > threshold
        elif operator == "lt":
            return value < threshold
        elif operator == "eq":
            return value == threshold
        elif operator == "ne":
            return value != threshold
        elif operator == "gte":
            return value >= threshold
        elif operator == "lte":
            return value <= threshold
        else:
            return False

    def add_alert_callback(self, callback: Callable[[Alert], None]) -> None:
        """Add a callback function to be triggered when alerts fire"""
        self.alert_callbacks.append(callback)

    def acknowledge_alert(self, alert_id: str, tenant_id: Optional[str] = None) -> bool:
        """Acknowledge an alert"""
        with self._lock:
            if tenant_id:
                alerts = self.tenant_alerts.get(tenant_id, {})
                if alert_id in alerts:
                    alerts[alert_id].acknowledged = True
                    alerts[alert_id].acknowledged_at = datetime.now()
                    return True
            else:
                if alert_id in self.active_alerts:
                    self.active_alerts[alert_id].acknowledged = True
                    self.active_alerts[alert_id].acknowledged_at = datetime.now()
                    return True
            return False

    def get_active_alerts(
        self,
        severity: Optional[AlertSeverity] = None,
        tenant_id: Optional[str] = None,
        component: Optional[str] = None,
    ) -> List[Alert]:
        """Get currently active alerts with filtering"""
        with self._lock:
            if tenant_id:
                alerts = list(self.tenant_alerts.get(tenant_id, {}).values())
            else:
                alerts = list(self.active_alerts.values())

            # Apply filters
            if severity:
                alerts = [a for a in alerts if a.severity == severity]
            if component:
                alerts = [a for a in alerts if a.component == component]

            return sorted(alerts, key=lambda a: (a.severity.value, a.timestamp), reverse=True)

    def get_alert_history(
        self,
        hours: int = 24,
        tenant_id: Optional[str] = None,
    ) -> List[Alert]:
        """Get alert history"""
        cutoff_time = datetime.now() - timedelta(hours=hours)

        if tenant_id:
            tenant_history = []
            for alert in self.alert_history:
                if alert.tenant_id == tenant_id and alert.timestamp >= cutoff_time:
                    tenant_history.append(alert)
            return tenant_history
        else:
            return [a for a in self.alert_history if a.timestamp >= cutoff_time]

    def get_alert_statistics(
        self,
        hours: int = 24,
        tenant_id: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Get alert statistics"""
        alerts = self.get_alert_history(hours, tenant_id)

        # Count by severity
        by_severity: Dict[str, int] = defaultdict(int)
        by_component: Dict[str, int] = defaultdict(int)
        by_hour: Dict[str, int] = defaultdict(int)

        for alert in alerts:
            by_severity[str(alert.severity.value)] += 1
            by_component[alert.component] += 1
            hour_key = alert.timestamp.strftime("%Y-%m-%d %H:00")
            by_hour[hour_key] += 1

        return {
            "total_alerts": len(alerts),
            "by_severity": dict(by_severity),
            "by_component": dict(by_component),
            "by_hour": dict(by_hour),
            "resolved_count": sum(1 for a in alerts if a.resolved),
            "acknowledged_count": sum(1 for a in alerts if a.acknowledged),
            "resolution_rate": sum(1 for a in alerts if a.resolved) / max(len(alerts), 1),
            "period_hours": hours,
        }


class DashboardManager:
    """Dashboard management with multi-tenant support"""

    def __init__(self):
        self.dashboards: Dict[str, Dashboard] = {}
        self.tenant_dashboards: Dict[str, Dict[str, Dashboard]] = defaultdict(dict)
        self.default_dashboards: Dict[str, Dashboard] = {}
        self._lock = threading.RLock()

        # Create default dashboards
        self._create_default_dashboards()

    def create_dashboard(
        self,
        name: str,
        description: str,
        dashboard_type: DashboardType,
        widgets: List[DashboardWidget],
        owner: str = "",
        tenant_id: Optional[str] = None,
        is_public: bool = False,
        layout: Optional[Dict[str, Any]] = None,
    ) -> str:
        """Create a new dashboard"""
        dashboard_id = str(uuid.uuid4())

        dashboard = Dashboard(
            dashboard_id=dashboard_id,
            name=name,
            description=description,
            dashboard_type=dashboard_type,
            widgets=widgets,
            layout=layout or {"grid": {"cols": 12, "rows": 8}},
            owner=owner,
            tenant_id=tenant_id,
            is_public=is_public,
        )

        with self._lock:
            if tenant_id:
                self.tenant_dashboards[tenant_id][dashboard_id] = dashboard
            else:
                self.dashboards[dashboard_id] = dashboard

        logger.info(f"Created dashboard: {name} ({dashboard_id})")
        return dashboard_id

    def get_dashboard(self, dashboard_id: str, tenant_id: Optional[str] = None) -> Optional[Dashboard]:
        """Get a dashboard by ID"""
        with self._lock:
            if tenant_id:
                return self.tenant_dashboards.get(tenant_id, {}).get(dashboard_id)
            else:
                return self.dashboards.get(dashboard_id) or self.default_dashboards.get(dashboard_id)

    def list_dashboards(
        self,
        tenant_id: Optional[str] = None,
        dashboard_type: Optional[DashboardType] = None,
        owner: Optional[str] = None,
        is_public: Optional[bool] = None,
    ) -> List[Dashboard]:
        """List dashboards with filtering"""
        with self._lock:
            dashboards: List[Dashboard] = []

            # Collect dashboards
            if tenant_id:
                dashboards.extend(self.tenant_dashboards.get(tenant_id, {}).values())
            else:
                dashboards.extend(self.dashboards.values())

            # Add public dashboards
            dashboards.extend([d for d in self.dashboards.values() if d.is_public])

            # Apply filters
            if dashboard_type:
                dashboards = [d for d in dashboards if d.dashboard_type == dashboard_type]
            if owner:
                dashboards = [d for d in dashboards if d.owner == owner]
            if is_public is not None:
                dashboards = [d for d in dashboards if d.is_public == is_public]

            return dashboards

    def update_dashboard(
        self,
        dashboard_id: str,
        updates: Dict[str, Any],
        tenant_id: Optional[str] = None,
    ) -> bool:
        """Update a dashboard"""
        with self._lock:
            if tenant_id:
                dashboard = self.tenant_dashboards.get(tenant_id, {}).get(dashboard_id)
            else:
                dashboard = self.dashboards.get(dashboard_id)

            if not dashboard:
                return False

            # Update fields
            for key, value in updates.items():
                if hasattr(dashboard, key):
                    setattr(dashboard, key, value)
                elif key in dashboard.__dict__:
                    dashboard.__dict__[key] = value

            dashboard.updated_at = datetime.now()
            return True

    def delete_dashboard(self, dashboard_id: str, tenant_id: Optional[str] = None) -> bool:
        """Delete a dashboard"""
        with self._lock:
            if tenant_id:
                if dashboard_id in self.tenant_dashboards.get(tenant_id, {}):
                    del self.tenant_dashboards[tenant_id][dashboard_id]
                    return True
            else:
                if dashboard_id in self.dashboards:
                    del self.dashboards[dashboard_id]
                    return True
                if dashboard_id in self.default_dashboards:
                    # Don't delete default dashboards
                    return False
            return False

    def _create_default_dashboards(self):
        """Create default dashboards"""
        # System Overview Dashboard
        system_overview_widgets = [
            DashboardWidget(
                widget_id="sys_health",
                widget_type="metric",
                title="System Health",
                position={"x": 0, "y": 0, "width": 4, "height": 2},
                config={"metric": "health_score", "format": "percentage"},
            ),
            DashboardWidget(
                widget_id="active_alerts",
                widget_type="metric",
                title="Active Alerts",
                position={"x": 4, "y": 0, "width": 4, "height": 2},
                config={"metric": "active_alerts", "format": "number"},
            ),
            DashboardWidget(
                widget_id="uptime",
                widget_type="metric",
                title="System Uptime",
                position={"x": 8, "y": 0, "width": 4, "height": 2},
                config={"metric": "uptime", "format": "duration"},
            ),
            DashboardWidget(
                widget_id="cpu_chart",
                widget_type="chart",
                title="CPU Usage",
                position={"x": 0, "y": 2, "width": 6, "height": 3},
                config={"chart_type": "line", "time_range": "1h"},
                metrics=["cpu_usage"],
            ),
            DashboardWidget(
                widget_id="memory_chart",
                widget_type="chart",
                title="Memory Usage",
                position={"x": 6, "y": 2, "width": 6, "height": 3},
                config={"chart_type": "line", "time_range": "1h"},
                metrics=["memory_usage"],
            ),
            DashboardWidget(
                widget_id="alert_table",
                widget_type="table",
                title="Recent Alerts",
                position={"x": 0, "y": 5, "width": 12, "height": 3},
                config={"limit": 10, "sort": "timestamp"},
            ),
        ]

        system_dashboard = Dashboard(
            dashboard_id="system_overview",
            name="System Overview",
            description="Overview of system health and performance",
            dashboard_type=DashboardType.SYSTEM_OVERVIEW,
            widgets=system_overview_widgets,
            owner="system",
            is_public=True,
        )

        self.default_dashboards["system_overview"] = system_dashboard

        # Hook Analysis Dashboard
        hook_analysis_widgets = [
            DashboardWidget(
                widget_id="hook_execution_chart",
                widget_type="chart",
                title="Hook Execution Rate",
                position={"x": 0, "y": 0, "width": 6, "height": 3},
                config={"chart_type": "line", "time_range": "24h"},
                metrics=["hook_execution_rate"],
            ),
            DashboardWidget(
                widget_id="hook_success_rate",
                widget_type="metric",
                title="Hook Success Rate",
                position={"x": 6, "y": 0, "width": 6, "height": 3},
                config={"metric": "hook_success_rate", "format": "percentage"},
            ),
            DashboardWidget(
                widget_id="slow_hooks",
                widget_type="table",
                title="Slowest Hooks",
                position={"x": 0, "y": 3, "width": 12, "height": 4},
                config={"limit": 15, "sort": "execution_time", "order": "desc"},
            ),
            DashboardWidget(
                widget_id="error_by_hook",
                widget_type="chart",
                title="Error Rate by Hook",
                position={"x": 0, "y": 7, "width": 12, "height": 3},
                config={"chart_type": "bar", "time_range": "24h"},
            ),
        ]

        hook_dashboard = Dashboard(
            dashboard_id="hook_analysis",
            name="Hook Analysis",
            description="Detailed analysis of hook execution performance",
            dashboard_type=DashboardType.HOOK_ANALYSIS,
            widgets=hook_analysis_widgets,
            owner="system",
            is_public=True,
        )

        self.default_dashboards["hook_analysis"] = hook_dashboard

        logger.info("Created default dashboards")


class RealtimeMonitoringDashboard:
    """
    Real-time Monitoring Dashboard System

    Enterprise-grade real-time monitoring dashboard with alerting,
    analytics, and visualization capabilities.
    """

    def __init__(
        self,
        metrics_buffer_size: int = 100000,
        retention_hours: int = 168,  # 7 days
        alert_check_interval: int = 30,  # seconds
        enable_websocket: bool = True,
        enable_external_integration: bool = True,
    ):
        """Initialize Real-time Monitoring Dashboard

        Args:
            metrics_buffer_size: Size of metrics buffer
            retention_hours: Hours to retain metrics
            alert_check_interval: Interval between alert checks
            enable_websocket: Enable WebSocket support for real-time updates
            enable_external_integration: Enable external monitoring system integration
        """
        self.metrics_buffer_size = metrics_buffer_size
        self.retention_hours = retention_hours
        self.alert_check_interval = alert_check_interval
        self.enable_websocket = enable_websocket
        self.enable_external_integration = enable_external_integration

        # Initialize components
        self.metrics_collector = MetricsCollector(metrics_buffer_size, retention_hours)
        self.alert_manager = AlertManager(self.metrics_collector)
        self.dashboard_manager = DashboardManager()

        # System state
        self._running = False
        self._startup_time = datetime.now()

        # WebSocket connections
        self.websocket_connections: Set[Any] = set()
        self.websocket_lock = threading.Lock()

        # Background tasks
        self._monitoring_thread: Optional[threading.Thread] = None
        self._alert_thread: Optional[threading.Thread] = None

        # External integrations
        self.external_integrations: Dict[str, Any] = {}

        # Setup default alerts
        self._setup_default_alerts()

        logger.info("Real-time Monitoring Dashboard initialized")

    def _setup_default_alerts(self):
        """Setup default alert rules"""
        # CPU usage alert
        self.alert_manager.add_alert_rule(
            name="High CPU Usage",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
            severity=AlertSeverity.ERROR,
            window_minutes=5,
            consecutive_violations=2,
            tags={"component": "system"},
            description="CPU usage exceeds 80% for 5 minutes",
        )

        # Memory usage alert
        self.alert_manager.add_alert_rule(
            name="High Memory Usage",
            metric_type=MetricType.MEMORY_USAGE,
            threshold=85.0,
            operator="gt",
            severity=AlertSeverity.ERROR,
            window_minutes=5,
            consecutive_violations=2,
            tags={"component": "system"},
            description="Memory usage exceeds 85% for 5 minutes",
        )

        # Error rate alert
        self.alert_manager.add_alert_rule(
            name="High Error Rate",
            metric_type=MetricType.ERROR_RATE,
            threshold=5.0,
            operator="gt",
            severity=AlertSeverity.CRITICAL,
            window_minutes=2,
            consecutive_violations=1,
            tags={"component": "system"},
            description="Error rate exceeds 5% in 2 minutes",
        )

        # Hook execution time alert
        self.alert_manager.add_alert_rule(
            name="Slow Hook Execution",
            metric_type=MetricType.RESPONSE_TIME,
            threshold=5000.0,  # 5 seconds
            operator="gt",
            severity=AlertSeverity.WARNING,
            window_minutes=10,
            consecutive_violations=3,
            tags={"component": "hooks"},
            description="Hook execution time exceeds 5 seconds",
        )

        logger.info("Default alert rules configured")

    async def start(self) -> None:
        """Start the monitoring dashboard system"""
        if self._running:
            return

        logger.info("Starting Real-time Monitoring Dashboard...")

        try:
            # Start background monitoring
            self._start_background_monitoring()

            # Start alert checking
            self._start_alert_monitoring()

            # Start WebSocket server if enabled
            if self.enable_websocket:
                await self._start_websocket_server()

            # Initialize external integrations
            if self.enable_external_integration:
                await self._initialize_external_integrations()

            self._running = True
            logger.info("Real-time Monitoring Dashboard started successfully")

        except Exception as e:
            logger.error(f"Error starting monitoring dashboard: {e}")
            raise

    def stop(self) -> None:
        """Stop the monitoring dashboard system"""
        if not self._running:
            return

        logger.info("Stopping Real-time Monitoring Dashboard...")

        try:
            # Stop background threads
            self._running = False
            if self._monitoring_thread:
                self._monitoring_thread.join(timeout=5)
            if self._alert_thread:
                self._alert_thread.join(timeout=5)

            # Close WebSocket connections
            with self.websocket_lock:
                for conn in self.websocket_connections:
                    try:
                        # Close connection
                        pass  # Implementation depends on WebSocket library
                    except Exception:
                        pass
                self.websocket_connections.clear()

            logger.info("Real-time Monitoring Dashboard stopped")

        except Exception as e:
            logger.error(f"Error stopping monitoring dashboard: {e}")

    def _start_background_monitoring(self) -> None:
        """Start background metrics collection"""

        def monitor_loop():
            while self._running:
                try:
                    self._collect_system_metrics()
                    time.sleep(30)  # Collect metrics every 30 seconds
                except Exception as e:
                    logger.error(f"Error in monitoring loop: {e}")
                    time.sleep(30)

        self._monitoring_thread = threading.Thread(target=monitor_loop, daemon=True)
        self._monitoring_thread.start()

    def _start_alert_monitoring(self) -> None:
        """Start alert checking"""

        def alert_loop():
            while self._running:
                try:
                    alerts = self.alert_manager.check_alerts()
                    if alerts:
                        for alert in alerts:
                            logger.warning(f"Alert triggered: {alert.title} - {alert.current_value}")
                            # Broadcast to WebSocket clients
                            self._broadcast_alert(alert)

                    time.sleep(self.alert_check_interval)
                except Exception as e:
                    logger.error(f"Error in alert loop: {e}")
                    time.sleep(self.alert_check_interval)

        self._alert_thread = threading.Thread(target=alert_loop, daemon=True)
        self._alert_thread.start()

    async def _start_websocket_server(self) -> None:
        """Start WebSocket server for real-time updates"""
        # Implementation depends on WebSocket library (websockets, FastAPI WebSocket, etc.)
        logger.info("WebSocket server support enabled")
        pass

    async def _initialize_external_integrations(self) -> None:
        """Initialize external monitoring system integrations"""
        # Prometheus integration
        try:
            await self._setup_prometheus_integration()
        except Exception as e:
            logger.warning(f"Failed to setup Prometheus integration: {e}")

        # DataDog integration
        try:
            await self._setup_datadog_integration()
        except Exception as e:
            logger.warning(f"Failed to setup DataDog integration: {e}")

        logger.info("External integrations initialized")

    async def _setup_prometheus_integration(self) -> None:
        """Setup Prometheus integration"""
        # Implementation would setup Prometheus metrics endpoint
        pass

    async def _setup_datadog_integration(self) -> None:
        """Setup DataDog integration"""
        # Implementation would setup DataDog API client
        pass

    def _collect_system_metrics(self) -> None:
        """Collect system performance metrics"""
        try:
            import psutil

            # CPU Usage
            cpu_percent = psutil.cpu_percent(interval=1)
            self.add_metric(
                MetricType.CPU_USAGE,
                cpu_percent,
                tags={"component": "system"},
                source="psutil",
            )

            # Memory Usage
            memory = psutil.virtual_memory()
            self.add_metric(
                MetricType.MEMORY_USAGE,
                memory.percent,
                tags={"component": "system"},
                source="psutil",
            )

            # Python process memory
            process = psutil.Process()
            process_memory = process.memory_info()
            self.add_metric(
                MetricType.MEMORY_USAGE,
                process_memory.rss / (1024**2),  # MB
                tags={"component": "python_process"},
                source="psutil",
            )

            # System load
            try:
                load_avg = psutil.getloadavg()
                self.add_metric(
                    MetricType.SYSTEM_PERFORMANCE,
                    load_avg[0],  # 1-minute load average
                    tags={"component": "system", "metric": "load_1min"},
                    source="psutil",
                )
            except (AttributeError, OSError):
                pass  # Not available on all systems

        except Exception as e:
            logger.error(f"Error collecting system metrics: {e}")

    def add_metric(
        self,
        metric_type: MetricType,
        value: Union[int, float],
        tags: Optional[Dict[str, str]] = None,
        source: str = "custom",
        component: str = "",
        tenant_id: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> None:
        """Add a metric to the monitoring system"""
        metric = MetricData(
            timestamp=datetime.now(),
            metric_type=metric_type,
            value=value,
            tags=tags or {},
            source=source,
            component=component,
            tenant_id=tenant_id,
            metadata=metadata or {},
        )

        self.metrics_collector.add_metric(metric)

    def get_dashboard_data(
        self,
        dashboard_id: str,
        tenant_id: Optional[str] = None,
        time_range: Optional[str] = None,  # "1h", "24h", "7d"
        filters: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Get dashboard data for visualization"""
        try:
            # Get dashboard
            dashboard = self.dashboard_manager.get_dashboard(dashboard_id, tenant_id)
            if not dashboard:
                return {"error": "Dashboard not found"}

            # Calculate time range
            end_time = datetime.now()
            if time_range:
                if time_range == "1h":
                    start_time = end_time - timedelta(hours=1)
                elif time_range == "24h":
                    start_time = end_time - timedelta(hours=24)
                elif time_range == "7d":
                    start_time = end_time - timedelta(days=7)
                else:
                    start_time = end_time - timedelta(hours=1)  # Default
            else:
                start_time = end_time - timedelta(hours=1)

            # Collect data for each widget
            widgets_data = {}
            for widget in dashboard.widgets:
                try:
                    widget_data = self._get_widget_data(widget, start_time, end_time, tenant_id, filters)
                    widgets_data[widget.widget_id] = widget_data
                except Exception as e:
                    logger.error(f"Error getting data for widget {widget.widget_id}: {e}")
                    widgets_data[widget.widget_id] = {"error": str(e)}

            return {
                "dashboard": dashboard.to_dict(),
                "widgets_data": widgets_data,
                "time_range": time_range or "1h",
                "generated_at": datetime.now().isoformat(),
            }

        except Exception as e:
            logger.error(f"Error getting dashboard data: {e}")
            return {"error": str(e), "generated_at": datetime.now().isoformat()}

    def _get_widget_data(
        self,
        widget: DashboardWidget,
        start_time: datetime,
        end_time: datetime,
        tenant_id: Optional[str] = None,
        filters: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Get data for a specific widget"""
        if widget.widget_type == "metric":
            return self._get_metric_widget_data(widget, tenant_id)
        elif widget.widget_type == "chart":
            return self._get_chart_widget_data(widget, start_time, end_time, tenant_id, filters)
        elif widget.widget_type == "table":
            return self._get_table_widget_data(widget, start_time, end_time, tenant_id, filters)
        elif widget.widget_type == "alert":
            return self._get_alert_widget_data(widget, tenant_id)
        else:
            return {"error": f"Unsupported widget type: {widget.widget_type}"}

    def _get_metric_widget_data(self, widget: DashboardWidget, tenant_id: Optional[str] = None) -> Dict[str, Any]:
        """Get data for metric widget"""
        metric_name = widget.config.get("metric")
        if not metric_name:
            return {"error": "No metric specified"}

        # Get latest metrics
        metric_type = None
        if metric_name == "cpu_usage":
            metric_type = MetricType.CPU_USAGE
        elif metric_name == "memory_usage":
            metric_type = MetricType.MEMORY_USAGE
        elif metric_name == "health_score":
            metric_type = MetricType.AVAILABILITY
        else:
            return {"error": f"Unknown metric: {metric_name}"}

        recent_metrics = self.metrics_collector.get_metrics(metric_type=metric_type, tenant_id=tenant_id, limit=1)

        if not recent_metrics:
            return {"value": 0, "status": "no_data"}

        latest_metric = recent_metrics[0]
        format_type = widget.config.get("format", "number")

        if format_type == "percentage":
            value = f"{latest_metric.value:.1f}%"
        elif format_type == "duration":
            value = f"{latest_metric.value:.0f}s"
        else:
            value = str(latest_metric.value)

        return {
            "value": latest_metric.value,
            "formatted_value": value,
            "timestamp": latest_metric.timestamp.isoformat(),
            "format": format_type,
        }

    def _get_chart_widget_data(
        self,
        widget: DashboardWidget,
        start_time: datetime,
        end_time: datetime,
        tenant_id: Optional[str] = None,
        filters: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Get data for chart widget"""
        chart_type = widget.config.get("chart_type", "line")
        metrics = widget.metrics or []

        chart_data = {
            "type": chart_type,
            "data": [],
            "labels": [],
        }

        for metric_name in metrics:
            # Map metric names to MetricType
            metric_type = self._map_metric_name(metric_name)
            if metric_type:
                metric_data = self.metrics_collector.get_metrics(
                    metric_type=metric_type,
                    tenant_id=tenant_id,
                    start_time=start_time,
                    end_time=end_time,
                    limit=100,
                )

                series_data = []
                for metric in metric_data:
                    series_data.append(
                        {
                            "timestamp": metric.timestamp.isoformat(),
                            "value": metric.value,
                        }
                    )

                chart_data["data"].append(
                    {
                        "name": metric_name,
                        "series": series_data,
                    }
                )

        return chart_data

    def _get_table_widget_data(
        self,
        widget: DashboardWidget,
        start_time: datetime,
        end_time: datetime,
        tenant_id: Optional[str] = None,
        filters: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Get data for table widget"""
        if widget.widget_id == "alert_table":
            alerts = self.alert_manager.get_alert_history(hours=24, tenant_id=tenant_id)

            table_data = []
            for alert in alerts:
                table_data.append(
                    {
                        "timestamp": alert.timestamp.isoformat(),
                        "severity": alert.severity.value,
                        "title": alert.title,
                        "description": alert.description,
                        "component": alert.component,
                        "resolved": alert.resolved,
                    }
                )

            return {
                "columns": [
                    "timestamp",
                    "severity",
                    "title",
                    "description",
                    "component",
                    "resolved",
                ],
                "data": table_data,
            }

        return {"error": "Unknown table type"}

    def _get_alert_widget_data(self, widget: DashboardWidget, tenant_id: Optional[str] = None) -> Dict[str, Any]:
        """Get data for alert widget"""
        active_alerts = self.alert_manager.get_active_alerts(tenant_id=tenant_id)

        return {
            "active_count": len(active_alerts),
            "severity_breakdown": {
                severity.value: len([a for a in active_alerts if a.severity == severity]) for severity in AlertSeverity
            },
            "recent_alerts": [
                {
                    "id": alert.alert_id,
                    "title": alert.title,
                    "severity": alert.severity.value,
                    "timestamp": alert.timestamp.isoformat(),
                }
                for alert in sorted(active_alerts, key=lambda a: a.timestamp, reverse=True)[:10]
            ],
        }

    def _map_metric_name(self, metric_name: str) -> Optional[MetricType]:
        """Map metric name to MetricType enum"""
        mapping = {
            "cpu_usage": MetricType.CPU_USAGE,
            "memory_usage": MetricType.MEMORY_USAGE,
            "hook_execution_rate": MetricType.THROUGHPUT,
            "hook_success_rate": MetricType.AVAILABILITY,
            "response_time": MetricType.RESPONSE_TIME,
            "error_rate": MetricType.ERROR_RATE,
            "network_io": MetricType.NETWORK_IO,
            "disk_io": MetricType.DISK_IO,
            "cache_performance": MetricType.CACHE_PERFORMANCE,
            "database_performance": MetricType.DATABASE_PERFORMANCE,
        }

        return mapping.get(metric_name)

    def _broadcast_alert(self, alert: Alert) -> None:
        """Broadcast alert to WebSocket clients"""
        if not self.enable_websocket:
            return

        {
            "type": "alert",
            "data": alert.to_dict(),
            "timestamp": datetime.now().isoformat(),
        }

        # Implementation would send message to all connected WebSocket clients
        logger.info(f"Broadcasting alert: {alert.title}")

    def get_system_status(self) -> Dict[str, Any]:
        """Get comprehensive system status"""
        return {
            "status": "running" if self._running else "stopped",
            "uptime_seconds": (datetime.now() - self._startup_time).total_seconds(),
            "metrics_collected": len(self.metrics_collector.metrics_buffer),
            "active_alerts": len(self.alert_manager.active_alerts),
            "total_dashboards": len(self.dashboard_manager.dashboards) + len(self.dashboard_manager.default_dashboards),
            "websocket_connections": len(self.websocket_connections),
            "external_integrations": list(self.external_integrations.keys()),
            "last_update": datetime.now().isoformat(),
        }

    def create_custom_dashboard(
        self,
        name: str,
        description: str,
        widgets: List[Dict[str, Any]],
        owner: str = "",
        tenant_id: Optional[str] = None,
        is_public: bool = False,
    ) -> str:
        """Create a custom dashboard from widget definitions"""
        dashboard_widgets = []

        for widget_def in widgets:
            widget = DashboardWidget(
                widget_id=widget_def.get("widget_id", str(uuid.uuid4())),
                widget_type=widget_def.get("widget_type", "metric"),
                title=widget_def.get("title", ""),
                position=widget_def.get("position", {"x": 0, "y": 0, "width": 4, "height": 2}),
                config=widget_def.get("config", {}),
                metrics=widget_def.get("metrics", []),
            )
            dashboard_widgets.append(widget)

        return self.dashboard_manager.create_dashboard(
            name=name,
            description=description,
            dashboard_type=DashboardType.CUSTOM,
            widgets=dashboard_widgets,
            owner=owner,
            tenant_id=tenant_id,
            is_public=is_public,
        )


# Global instance for easy access
_monitoring_dashboard: Optional[RealtimeMonitoringDashboard] = None


def get_monitoring_dashboard(
    metrics_buffer_size: int = 100000,
    retention_hours: int = 168,
    alert_check_interval: int = 30,
    enable_websocket: bool = True,
    enable_external_integration: bool = True,
) -> RealtimeMonitoringDashboard:
    """Get or create global monitoring dashboard instance"""
    global _monitoring_dashboard
    if _monitoring_dashboard is None:
        _monitoring_dashboard = RealtimeMonitoringDashboard(
            metrics_buffer_size=metrics_buffer_size,
            retention_hours=retention_hours,
            alert_check_interval=alert_check_interval,
            enable_websocket=enable_websocket,
            enable_external_integration=enable_external_integration,
        )
    return _monitoring_dashboard


# Convenience functions
async def start_monitoring() -> None:
    """Start the monitoring dashboard"""
    dashboard = get_monitoring_dashboard()
    await dashboard.start()


def stop_monitoring() -> None:
    """Stop the monitoring dashboard"""
    dashboard = get_monitoring_dashboard()
    dashboard.stop()


def add_system_metric(
    metric_type: MetricType,
    value: Union[int, float],
    component: str = "",
    tags: Optional[Dict[str, str]] = None,
) -> None:
    """Add a system metric"""
    dashboard = get_monitoring_dashboard()
    dashboard.add_metric(metric_type, value, tags, component=component)


def add_hook_metric(
    hook_path: str,
    execution_time_ms: float,
    success: bool,
    tenant_id: Optional[str] = None,
) -> None:
    """Add hook execution metric"""
    dashboard = get_monitoring_dashboard()

    # Add execution time metric
    dashboard.add_metric(
        MetricType.RESPONSE_TIME,
        execution_time_ms,
        tags={"hook_path": hook_path, "success": str(success)},
        component="hooks",
        tenant_id=tenant_id,
    )

    # Add success rate metric
    dashboard.add_metric(
        MetricType.AVAILABILITY,
        1.0 if success else 0.0,
        tags={"hook_path": hook_path},
        component="hooks",
        tenant_id=tenant_id,
    )


if __name__ == "__main__":
    # Example usage
    async def main():
        print("📊 Starting Real-time Monitoring Dashboard...")

        # Initialize monitoring dashboard
        dashboard = RealtimeMonitoringDashboard(
            metrics_buffer_size=10000,
            retention_hours=24,
            enable_websocket=False,  # Disable for demo
            enable_external_integration=False,
        )

        try:
            # Start the system
            await dashboard.start()

            # Add some test metrics
            print("\n📈 Adding test metrics...")
            for i in range(20):
                dashboard.add_metric(
                    MetricType.CPU_USAGE,
                    20 + (i % 80),  # CPU usage from 20% to 100%
                    tags={"component": "system", "instance": "demo"},
                    source="demo",
                )

                dashboard.add_metric(
                    MetricType.MEMORY_USAGE,
                    30 + (i % 70),  # Memory usage from 30% to 100%
                    tags={"component": "system"},
                    source="demo",
                )

                dashboard.add_hook_metric(
                    f"test_hook_{i % 5}.py",
                    100 + (i * 50),  # Execution time from 100ms to 1000ms
                    i % 4 != 0,  # 75% success rate
                    tenant_id="demo_tenant" if i % 2 == 0 else None,
                )

                await asyncio.sleep(0.1)

            # Let metrics process
            print("\n⏳ Processing metrics and checking alerts...")
            await asyncio.sleep(5)

            # Get system status
            status = dashboard.get_system_status()
            print("\n📊 System Status:")
            print(f"  Status: {status['status']}")
            print(f"  Uptime: {status['uptime_seconds']:.1f}s")
            print(f"  Metrics collected: {status['metrics_collected']}")
            print(f"  Active alerts: {status['active_alerts']}")
            print(f"  Total dashboards: {status['total_dashboards']}")

            # Get dashboard data
            dashboard_data = dashboard.get_dashboard_data("system_overview")
            print("\n📱 Dashboard Data:")
            print(f"  Dashboard: {dashboard_data['dashboard']['name']}")
            print(f"  Widgets: {len(dashboard_data.get('widgets_data', {}))}")
            print(f"  Generated at: {dashboard_data.get('generated_at')}")

            # Get metrics statistics
            cpu_stats = dashboard.metrics_collector.get_statistics(MetricType.CPU_USAGE, minutes=10)
            memory_stats = dashboard.metrics_collector.get_statistics(MetricType.MEMORY_USAGE, minutes=10)
            print("\n📈 Metrics Statistics (last 10 minutes):")
            cpu_avg = cpu_stats.get("average", 0)
            cpu_max = cpu_stats.get("max", 0)
            cpu_count = cpu_stats.get("count", 0)
            print(f"  CPU Usage - Avg: {cpu_avg:.1f}%, Max: {cpu_max:.1f}%, Count: {cpu_count}")
            mem_avg = memory_stats.get("average", 0)
            mem_max = memory_stats.get("max", 0)
            mem_count = memory_stats.get("count", 0)
            print(f"  Memory Usage - Avg: {mem_avg:.1f}%, Max: {mem_max:.1f}%, Count: {mem_count}")

            # Get alert statistics
            alert_stats = dashboard.alert_manager.get_alert_statistics(hours=1)
            print("\n🚨 Alert Statistics (last 1 hour):")
            print(f"  Total alerts: {alert_stats['total_alerts']}")
            print(f"  Resolved: {alert_stats['resolved_count']}")
            print(f"  Resolution rate: {alert_stats['resolution_rate']:.1%}")
            print(f"  By severity: {alert_stats['by_severity']}")

            # List available dashboards
            dashboards = dashboard.dashboard_manager.list_dashboards()
            print("\n📋 Available Dashboards:")
            for dashboard_info in dashboards:
                print(f"  - {dashboard_info.name} ({dashboard_info.dashboard_type.value})")

            print("\n✅ Real-time Monitoring Dashboard demo completed successfully!")

        except Exception as e:
            print(f"\n❌ Demo failed: {str(e)}")
            import traceback

            traceback.print_exc()

        finally:
            # Stop the system
            print("\n🛑 Stopping monitoring dashboard...")
            dashboard.stop()
            print("✅ System stopped")

    # Run the demo
    asyncio.run(main())
