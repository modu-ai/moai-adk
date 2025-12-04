"""
Comprehensive Monitoring System

Real-time monitoring, analytics, and predictive analysis for MoAI-ADK
with automated alerting and optimization capabilities.

Key Features:
- Real-time metrics collection and analysis
- User behavior analytics and pattern recognition
- Predictive analytics and trend analysis
- Automated alerting system
- System health monitoring
- Performance optimization recommendations
- Real-time dashboard interface
"""

import json
import logging
import statistics
import threading
import time
from collections import defaultdict, deque
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
from pathlib import Path
from typing import Any, Callable, Dict, List, Optional, Union

import psutil

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class MetricType(Enum):
    """Types of metrics collected by the monitoring system"""

    SYSTEM_PERFORMANCE = "system_performance"
    USER_BEHAVIOR = "user_behavior"
    TOKEN_USAGE = "token_usage"
    ERROR_RATE = "error_rate"
    RESPONSE_TIME = "response_time"
    MEMORY_USAGE = "memory_usage"
    CPU_USAGE = "cpu_usage"
    THROUGHPUT = "throughput"
    AVAILABILITY = "availability"


class AlertSeverity(Enum):
    """Alert severity levels"""

    LOW = 1
    MEDIUM = 2
    HIGH = 3
    CRITICAL = 4
    EMERGENCY = 5


class HealthStatus(Enum):
    """System health status"""

    HEALTHY = "healthy"
    WARNING = "warning"
    DEGRADED = "degraded"
    CRITICAL = "critical"
    DOWN = "down"


@dataclass
class MetricData:
    """Single metric data point"""

    timestamp: datetime
    metric_type: MetricType
    value: Union[int, float, str, bool]
    tags: Dict[str, str] = field(default_factory=dict)
    source: str = ""
    metadata: Dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "timestamp": self.timestamp.isoformat(),
            "metric_type": self.metric_type.value,
            "value": self.value,
            "tags": self.tags,
            "source": self.source,
            "metadata": self.metadata,
        }


@dataclass
class Alert:
    """Alert definition and data"""

    alert_id: str
    severity: AlertSeverity
    title: str
    description: str
    timestamp: datetime
    metric_type: MetricType
    threshold: float
    current_value: float
    source: str
    tags: Dict[str, str] = field(default_factory=dict)
    resolved: bool = False
    resolved_at: Optional[datetime] = None
    acknowledged: bool = False
    acknowledged_at: Optional[datetime] = None

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
            "tags": self.tags,
            "resolved": self.resolved,
            "resolved_at": self.resolved_at.isoformat() if self.resolved_at else None,
            "acknowledged": self.acknowledged,
            "acknowledged_at": (self.acknowledged_at.isoformat() if self.acknowledged_at else None),
        }


@dataclass
class SystemHealth:
    """System health status information"""

    status: HealthStatus
    timestamp: datetime
    overall_score: float  # 0-100
    component_scores: Dict[str, float] = field(default_factory=dict)
    active_alerts: List[str] = field(default_factory=list)
    recent_metrics: Dict[str, float] = field(default_factory=dict)
    uptime_percentage: float = 100.0
    last_check: Optional[datetime] = None

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "status": self.status.value,
            "timestamp": self.timestamp.isoformat(),
            "overall_score": self.overall_score,
            "component_scores": self.component_scores,
            "active_alerts": self.active_alerts,
            "recent_metrics": self.recent_metrics,
            "uptime_percentage": self.uptime_percentage,
            "last_check": self.last_check.isoformat() if self.last_check else None,
        }


class MetricsCollector:
    """Collects and manages system metrics"""

    def __init__(self, buffer_size: int = 10000, retention_hours: int = 24):
        self.buffer_size = buffer_size
        self.retention_hours = retention_hours
        self.metrics_buffer: Dict[MetricType, deque] = defaultdict(lambda: deque(maxlen=buffer_size))
        self.aggregated_metrics: Dict[MetricType, Dict[str, Any]] = defaultdict(dict)
        self._lock = threading.Lock()
        self._last_cleanup = datetime.now()

    def add_metric(self, metric: MetricData) -> None:
        """Add a metric to the collection"""
        with self._lock:
            self.metrics_buffer[metric.metric_type].append(metric)
            self._update_aggregated_metrics(metric)
            self._cleanup_old_metrics()

    def _update_aggregated_metrics(self, metric: MetricData) -> None:
        """Update aggregated statistics for a metric type"""
        if metric.metric_type not in self.aggregated_metrics:
            self.aggregated_metrics[metric.metric_type] = {
                "count": 0,
                "sum": 0,
                "min": float("inf"),
                "max": float("-inf"),
                "values": [],
            }

        agg = self.aggregated_metrics[metric.metric_type]

        if isinstance(metric.value, (int, float)):
            agg["count"] += 1
            agg["sum"] += metric.value
            agg["min"] = min(agg["min"], metric.value)
            agg["max"] = max(agg["max"], metric.value)
            agg["values"].append(metric.value)

            # Keep only recent values for statistics
            if len(agg["values"]) > 1000:
                agg["values"] = agg["values"][-1000:]

    def _cleanup_old_metrics(self) -> None:
        """Remove metrics older than retention period"""
        now = datetime.now()
        if (now - self._last_cleanup).seconds < 300:  # Cleanup every 5 minutes
            return

        cutoff_time = now - timedelta(hours=self.retention_hours)

        for metric_type in self.metrics_buffer:
            while self.metrics_buffer[metric_type] and self.metrics_buffer[metric_type][0].timestamp < cutoff_time:
                self.metrics_buffer[metric_type].popleft()

        self._last_cleanup = now

    def get_metrics(
        self,
        metric_type: Optional[MetricType] = None,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
        limit: Optional[int] = None,
    ) -> List[MetricData]:
        """Get metrics with optional filtering"""
        with self._lock:
            if metric_type:
                metrics = list(self.metrics_buffer[metric_type])
            else:
                metrics = []
                for mlist in self.metrics_buffer.values():
                    metrics.extend(mlist)

            # Filter by time range
            if start_time:
                metrics = [m for m in metrics if m.timestamp >= start_time]
            if end_time:
                metrics = [m for m in metrics if m.timestamp <= end_time]

            # Sort by timestamp (newest first)
            metrics.sort(key=lambda m: m.timestamp, reverse=True)

            # Apply limit
            if limit:
                metrics = metrics[:limit]

            return metrics

    def get_statistics(self, metric_type: MetricType, minutes: int = 60) -> Dict[str, Any]:
        """Get statistical summary for a metric type"""
        with self._lock:
            agg = self.aggregated_metrics.get(metric_type, {})

            if not agg or agg["count"] == 0:
                return {
                    "count": 0,
                    "average": None,
                    "min": None,
                    "max": None,
                    "median": None,
                    "std_dev": None,
                }

            values = agg["values"]
            if not values:
                return {
                    "count": agg["count"],
                    "average": agg["sum"] / agg["count"],
                    "min": agg["min"],
                    "max": agg["max"],
                    "median": None,
                    "std_dev": None,
                }

            try:
                return {
                    "count": len(values),
                    "average": statistics.mean(values),
                    "median": statistics.median(values),
                    "min": min(values),
                    "max": max(values),
                    "std_dev": statistics.stdev(values) if len(values) > 1 else 0,
                    "p95": (statistics.quantiles(values, n=20)[18] if len(values) > 20 else max(values)),
                    "p99": (statistics.quantiles(values, n=100)[98] if len(values) > 100 else max(values)),
                }
            except (statistics.StatisticsError, IndexError):
                return {
                    "count": len(values),
                    "average": statistics.mean(values),
                    "median": statistics.median(values),
                    "min": min(values),
                    "max": max(values),
                    "std_dev": 0,
                    "p95": max(values),
                    "p99": max(values),
                }


class AlertManager:
    """Manages alert rules, detection, and notification"""

    def __init__(self, metrics_collector: MetricsCollector):
        self.metrics_collector = metrics_collector
        self.alert_rules: List[Dict[str, Any]] = []
        self.active_alerts: Dict[str, Alert] = {}
        self.alert_history: List[Alert] = []
        self.alert_callbacks: List[Callable[[Alert], None]] = []
        self._lock = threading.Lock()

    def add_alert_rule(
        self,
        name: str,
        metric_type: MetricType,
        threshold: float,
        operator: str = "gt",  # gt, lt, eq, ne
        severity: AlertSeverity = AlertSeverity.MEDIUM,
        window_minutes: int = 5,
        consecutive_violations: int = 1,
        tags: Optional[Dict[str, str]] = None,
        description: Optional[str] = None,
    ) -> None:
        """Add an alert rule"""
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
            "enabled": True,
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

                # Get recent metrics for this rule
                recent_metrics = self.metrics_collector.get_metrics(
                    metric_type=rule["metric_type"],
                    start_time=datetime.now() - timedelta(minutes=rule["window_minutes"]),
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
                        tags=rule["tags"],
                    )

                    self.active_alerts[alert_id] = alert
                    self.alert_history.append(alert)
                    triggered_alerts.append(alert)

                    # Trigger callbacks
                    for callback in self.alert_callbacks:
                        try:
                            callback(alert)
                        except Exception as e:
                            logger.error(f"Error in alert callback: {e}")

                rule["violation_count"] = violations
                rule["last_check"] = datetime.now()

            # Check for resolved alerts
            resolved_alerts = []
            for alert_id, alert in list(self.active_alerts.items()):
                # Check if alert condition is no longer met
                rule = next((r for r in self.alert_rules if r["name"] in alert_id), None)
                if rule:
                    recent_metrics = self.metrics_collector.get_metrics(
                        metric_type=rule["metric_type"],
                        start_time=datetime.now() - timedelta(minutes=1),  # Check last minute
                    )

                    if recent_metrics:
                        latest_value = None
                        for metric in recent_metrics:
                            if isinstance(metric.value, (int, float)):
                                if not self._evaluate_condition(metric.value, rule["threshold"], rule["operator"]):
                                    latest_value = metric.value
                                    break

                        if latest_value is not None:
                            # Alert is resolved
                            alert.resolved = True
                            alert.resolved_at = datetime.now()
                            resolved_alerts.append(alert)
                            del self.active_alerts[alert_id]

        return triggered_alerts

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

    def acknowledge_alert(self, alert_id: str) -> bool:
        """Acknowledge an alert"""
        with self._lock:
            if alert_id in self.active_alerts:
                self.active_alerts[alert_id].acknowledged = True
                self.active_alerts[alert_id].acknowledged_at = datetime.now()
                return True
            return False

    def get_active_alerts(self, severity: Optional[AlertSeverity] = None) -> List[Alert]:
        """Get currently active alerts"""
        alerts = list(self.active_alerts.values())
        if severity:
            alerts = [a for a in alerts if a.severity == severity]
        return sorted(alerts, key=lambda a: (a.severity.value, a.timestamp), reverse=True)

    def get_alert_history(self, hours: int = 24) -> List[Alert]:
        """Get alert history"""
        cutoff_time = datetime.now() - timedelta(hours=hours)
        return [a for a in self.alert_history if a.timestamp >= cutoff_time]


class PredictiveAnalytics:
    """Predictive analytics for system performance and user behavior"""

    def __init__(self, metrics_collector: MetricsCollector):
        self.metrics_collector = metrics_collector
        self.models: Dict[str, Dict[str, Any]] = {}
        self.predictions: Dict[str, Dict[str, Any]] = {}

    def predict_metric_trend(
        self,
        metric_type: MetricType,
        hours_ahead: int = 1,
        confidence_threshold: float = 0.7,
    ) -> Dict[str, Any]:
        """Predict metric values for specified hours ahead"""
        try:
            # Get historical data
            historical_metrics = self.metrics_collector.get_metrics(
                metric_type=metric_type, start_time=datetime.now() - timedelta(hours=24)
            )

            if len(historical_metrics) < 10:
                return {
                    "prediction": None,
                    "confidence": 0.0,
                    "reason": "Insufficient historical data",
                }

            # Extract numeric values
            values = []
            timestamps = []
            for metric in historical_metrics:
                if isinstance(metric.value, (int, float)):
                    values.append(metric.value)
                    timestamps.append(metric.timestamp)

            if len(values) < 10:
                return {
                    "prediction": None,
                    "confidence": 0.0,
                    "reason": "Insufficient numeric data points",
                }

            # Simple linear regression for prediction
            import numpy as np

            # Convert timestamps to numeric values (hours ago)
            now = datetime.now()
            x = np.array([(now - ts).total_seconds() / 3600 for ts in timestamps])
            y = np.array(values)

            # Fit linear model
            coeffs = np.polyfit(x, y, 1)

            # Predict future values
            future_x = np.array([-h for h in range(1, hours_ahead + 1)])
            future_y = np.polyval(coeffs, future_x)

            # Calculate confidence based on R-squared
            y_pred = np.polyval(coeffs, x)
            ss_res = np.sum((y - y_pred) ** 2)
            ss_tot = np.sum((y - np.mean(y)) ** 2)
            r_squared = 1 - (ss_res / ss_tot) if ss_tot != 0 else 0

            confidence = max(0, r_squared)

            return {
                "prediction": {
                    "future_values": future_y.tolist(),
                    "trend": ("increasing" if coeffs[0] > 0 else "decreasing" if coeffs[0] < 0 else "stable"),
                    "slope": coeffs[0],
                },
                "confidence": confidence,
                "data_points": len(values),
                "model_type": "linear_regression",
                "reason": f"Linear regression on {len(values)} data points with R²={r_squared:.3f}",
            }

        except Exception as e:
            logger.error(f"Error in predictive analytics: {e}")
            return {
                "prediction": None,
                "confidence": 0.0,
                "reason": f"Analysis error: {str(e)}",
            }

    def detect_anomalies(
        self,
        metric_type: MetricType,
        z_score_threshold: float = 2.0,
        window_minutes: int = 60,
    ) -> Dict[str, Any]:
        """Detect anomalies in metric data using statistical methods"""
        try:
            recent_metrics = self.metrics_collector.get_metrics(
                metric_type=metric_type,
                start_time=datetime.now() - timedelta(minutes=window_minutes),
            )

            values = []
            for metric in recent_metrics:
                if isinstance(metric.value, (int, float)):
                    values.append(metric.value)

            if len(values) < 5:
                return {
                    "anomalies": [],
                    "statistics": {},
                    "reason": "Insufficient data for anomaly detection",
                }

            import numpy as np

            values_array = np.array(values)
            mean = np.mean(values_array)
            std = np.std(values_array)

            if std == 0:
                return {
                    "anomalies": [],
                    "statistics": {"mean": mean, "std": std},
                    "reason": "No variance in data",
                }

            # Detect anomalies using Z-score
            z_scores = np.abs((values_array - mean) / std)
            anomaly_indices = np.where(z_scores > z_score_threshold)[0]

            anomalies = []
            for i, idx in enumerate(anomaly_indices):
                metric = recent_metrics[idx]
                anomalies.append(
                    {
                        "timestamp": metric.timestamp.isoformat(),
                        "value": metric.value,
                        "z_score": float(z_scores[idx]),
                        "deviation": float(values[idx] - mean),
                    }
                )

            return {
                "anomalies": anomalies,
                "statistics": {
                    "mean": float(mean),
                    "std": float(std),
                    "min": float(np.min(values_array)),
                    "max": float(np.max(values_array)),
                    "count": len(values),
                },
                "threshold": z_score_threshold,
                "reason": f"Found {len(anomalies)} anomalies using Z-score > {z_score_threshold}",
            }

        except Exception as e:
            logger.error(f"Error in anomaly detection: {e}")
            return {
                "anomalies": [],
                "statistics": {},
                "reason": f"Analysis error: {str(e)}",
            }


class PerformanceMonitor:
    """System performance monitoring"""

    def __init__(self):
        self.start_time = datetime.now()
        self.metrics_collector = MetricsCollector()
        self.alert_manager = AlertManager(self.metrics_collector)
        self.predictive_analytics = PredictiveAnalytics(self.metrics_collector)
        self._running = False
        self._monitor_thread: Optional[threading.Thread] = None
        self._monitor_interval = 30  # seconds

    def start(self) -> None:
        """Start performance monitoring"""
        if self._running:
            return

        self._running = True
        self._monitor_thread = threading.Thread(target=self._monitor_loop, daemon=True)
        self._monitor_thread.start()
        logger.info("Performance monitoring started")

    def stop(self) -> None:
        """Stop performance monitoring"""
        self._running = False
        if self._monitor_thread:
            self._monitor_thread.join(timeout=5)
        logger.info("Performance monitoring stopped")

    def _monitor_loop(self) -> None:
        """Main monitoring loop"""
        while self._running:
            try:
                self._collect_system_metrics()
                self._check_alerts()
                time.sleep(self._monitor_interval)
            except Exception as e:
                logger.error(f"Error in monitoring loop: {e}")
                time.sleep(self._monitor_interval)

    def _collect_system_metrics(self) -> None:
        """Collect system performance metrics"""
        try:
            # CPU Usage
            cpu_percent = psutil.cpu_percent(interval=1)
            self.metrics_collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.CPU_USAGE,
                    value=cpu_percent,
                    tags={"component": "system"},
                    source="psutil",
                )
            )

            # Memory Usage
            memory = psutil.virtual_memory()
            self.metrics_collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.MEMORY_USAGE,
                    value=memory.percent,
                    tags={"component": "system", "total_gb": memory.total / (1024**3)},
                    source="psutil",
                )
            )

            # Python process memory
            process = psutil.Process()
            process_memory = process.memory_info()
            self.metrics_collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.MEMORY_USAGE,
                    value=process_memory.rss / (1024**2),  # MB
                    tags={"component": "python_process"},
                    source="psutil",
                )
            )

            # System load
            load_avg = psutil.getloadavg()
            self.metrics_collector.add_metric(
                MetricData(
                    timestamp=datetime.now(),
                    metric_type=MetricType.SYSTEM_PERFORMANCE,
                    value=load_avg[0],  # 1-minute load average
                    tags={"component": "system", "metric": "load_1min"},
                    source="psutil",
                )
            )

        except Exception as e:
            logger.error(f"Error collecting system metrics: {e}")

    def _check_alerts(self) -> None:
        """Check for alerts"""
        try:
            alerts = self.alert_manager.check_alerts()
            if alerts:
                for alert in alerts:
                    logger.warning(f"Alert triggered: {alert.title} - {alert.current_value}")

        except Exception as e:
            logger.error(f"Error checking alerts: {e}")

    def add_custom_metric(
        self,
        metric_type: MetricType,
        value: Union[int, float],
        tags: Optional[Dict[str, str]] = None,
        source: str = "custom",
    ) -> None:
        """Add a custom metric"""
        self.metrics_collector.add_metric(
            MetricData(
                timestamp=datetime.now(),
                metric_type=metric_type,
                value=value,
                tags=tags or {},
                source=source,
            )
        )

    def get_system_health(self) -> SystemHealth:
        """Get overall system health status"""
        try:
            # Calculate component scores
            component_scores = {}

            # CPU health
            cpu_metrics = self.metrics_collector.get_metrics(
                MetricType.CPU_USAGE, start_time=datetime.now() - timedelta(minutes=5)
            )
            if cpu_metrics:
                cpu_values = [m.value for m in cpu_metrics if isinstance(m.value, (int, float))]
                if cpu_values:
                    avg_cpu = statistics.mean(cpu_values)
                    cpu_score = max(0, 100 - avg_cpu)  # Lower CPU usage = higher score
                    component_scores["cpu"] = cpu_score

            # Memory health
            memory_metrics = self.metrics_collector.get_metrics(
                MetricType.MEMORY_USAGE,
                start_time=datetime.now() - timedelta(minutes=5),
            )
            if memory_metrics:
                memory_values = [m.value for m in memory_metrics if isinstance(m.value, (int, float))]
                if memory_values:
                    avg_memory = statistics.mean(memory_values)
                    memory_score = max(0, 100 - avg_memory)  # Lower memory usage = higher score
                    component_scores["memory"] = memory_score

            # Error rate health
            error_metrics = self.metrics_collector.get_metrics(
                MetricType.ERROR_RATE, start_time=datetime.now() - timedelta(minutes=10)
            )
            if error_metrics:
                error_values = [m.value for m in error_metrics if isinstance(m.value, (int, float))]
                if error_values:
                    avg_error = statistics.mean(error_values)
                    error_score = max(0, 100 - avg_error * 10)  # Lower error rate = higher score
                    component_scores["error_rate"] = error_score

            # Calculate overall score
            if component_scores:
                overall_score = statistics.mean(component_scores.values())
            else:
                overall_score = 100.0

            # Determine health status
            if overall_score >= 90:
                status = HealthStatus.HEALTHY
            elif overall_score >= 70:
                status = HealthStatus.WARNING
            elif overall_score >= 50:
                status = HealthStatus.DEGRADED
            elif overall_score >= 30:
                status = HealthStatus.CRITICAL
            else:
                status = HealthStatus.DOWN

            # Get active alerts
            active_alerts = list(self.alert_manager.active_alerts.keys())

            # Get recent metrics summary
            recent_metrics = {}
            for metric_type in [
                MetricType.CPU_USAGE,
                MetricType.MEMORY_USAGE,
                MetricType.ERROR_RATE,
            ]:
                recent_metric = self.metrics_collector.get_metrics(metric_type, limit=1)
                if recent_metric and isinstance(recent_metric[0].value, (int, float)):
                    recent_metrics[metric_type.value] = recent_metric[0].value

            return SystemHealth(
                status=status,
                timestamp=datetime.now(),
                overall_score=overall_score,
                component_scores=component_scores,
                active_alerts=active_alerts,
                recent_metrics=recent_metrics,
                last_check=datetime.now(),
            )

        except Exception as e:
            logger.error(f"Error calculating system health: {e}")
            return SystemHealth(
                status=HealthStatus.DOWN,
                timestamp=datetime.now(),
                overall_score=0.0,
                last_check=datetime.now(),
            )

    def setup_default_alerts(self) -> None:
        """Setup default alert rules"""
        # CPU usage alert
        self.alert_manager.add_alert_rule(
            name="High CPU Usage",
            metric_type=MetricType.CPU_USAGE,
            threshold=80.0,
            operator="gt",
            severity=AlertSeverity.HIGH,
            window_minutes=5,
            consecutive_violations=2,
            tags={"component": "cpu"},
        )

        # Memory usage alert
        self.alert_manager.add_alert_rule(
            name="High Memory Usage",
            metric_type=MetricType.MEMORY_USAGE,
            threshold=85.0,
            operator="gt",
            severity=AlertSeverity.HIGH,
            window_minutes=5,
            consecutive_violations=2,
            tags={"component": "memory"},
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
            tags={"component": "errors"},
        )

        logger.info("Default alert rules configured")


class ComprehensiveMonitoringSystem:
    """Main monitoring system orchestrator"""

    def __init__(self, config_file: Optional[Path] = None):
        self.config_file = config_file or Path.cwd() / ".moai" / "config" / "monitoring.json"
        self.config = self._load_config()

        # Initialize components
        self.metrics_collector = MetricsCollector(
            buffer_size=self.config.get("buffer_size", 10000),
            retention_hours=self.config.get("retention_hours", 24),
        )

        self.alert_manager = AlertManager(self.metrics_collector)
        self.predictive_analytics = PredictiveAnalytics(self.metrics_collector)
        self.performance_monitor = PerformanceMonitor()

        # Initialize monitoring status
        self._running = False
        self._startup_time = datetime.now()

    def _load_config(self) -> Dict[str, Any]:
        """Load monitoring configuration"""
        default_config = {
            "buffer_size": 10000,
            "retention_hours": 24,
            "monitor_interval": 30,
            "alert_check_interval": 60,
            "predictive_analysis_hours": 24,
            "health_check_interval": 300,
            "enable_predictions": True,
            "enable_anomaly_detection": True,
            "auto_optimization": False,
        }

        if self.config_file.exists():
            try:
                with open(self.config_file, "r") as f:
                    config = json.load(f)
                default_config.update(config)
            except Exception as e:
                logger.error(f"Error loading monitoring config: {e}")

        return default_config

    def start(self) -> None:
        """Start the monitoring system"""
        if self._running:
            return

        logger.info("Starting Comprehensive Monitoring System")

        # Start performance monitoring
        self.performance_monitor.start()

        # Setup default alerts
        self.performance_monitor.setup_default_alerts()

        # Setup alert callbacks
        self.alert_manager.add_alert_callback(self._handle_alert)

        self._running = True
        logger.info("Comprehensive Monitoring System started successfully")

    def stop(self) -> None:
        """Stop the monitoring system"""
        if not self._running:
            return

        logger.info("Stopping Comprehensive Monitoring System")

        self.performance_monitor.stop()
        self._running = False

        logger.info("Comprehensive Monitoring System stopped")

    def _handle_alert(self, alert: Alert) -> None:
        """Handle triggered alerts"""
        logger.warning(f"ALERT: {alert.title} - {alert.description}")

        # Here you could add additional alert handling:
        # - Send notifications
        # - Trigger automated responses
        # - Log to external systems
        # - Send to monitoring dashboard

    def add_metric(
        self,
        metric_type: MetricType,
        value: Union[int, float],
        tags: Optional[Dict[str, str]] = None,
        source: str = "user",
    ) -> None:
        """Add a custom metric"""
        self.performance_monitor.add_custom_metric(metric_type, value, tags, source)

    def get_dashboard_data(self) -> Dict[str, Any]:
        """Get data for monitoring dashboard"""
        try:
            # System health
            health = self.performance_monitor.get_system_health()

            # Active alerts
            active_alerts = self.alert_manager.get_active_alerts()

            # Recent metrics summary
            recent_metrics = {}
            for metric_type in [
                MetricType.CPU_USAGE,
                MetricType.MEMORY_USAGE,
                MetricType.ERROR_RATE,
                MetricType.RESPONSE_TIME,
            ]:
                stats = self.metrics_collector.get_statistics(metric_type, minutes=60)
                if stats["count"] > 0:
                    recent_metrics[metric_type.value] = stats

            # Predictions
            predictions = {}
            if self.config.get("enable_predictions", True):
                for metric_type in [MetricType.CPU_USAGE, MetricType.MEMORY_USAGE]:
                    pred = self.predictive_analytics.predict_metric_trend(metric_type, hours_ahead=1)
                    if pred["confidence"] > 0.5:
                        predictions[metric_type.value] = pred

            return {
                "health": health.to_dict(),
                "active_alerts": [alert.to_dict() for alert in active_alerts],
                "recent_metrics": recent_metrics,
                "predictions": predictions,
                "uptime_seconds": (datetime.now() - self._startup_time).total_seconds(),
                "last_update": datetime.now().isoformat(),
            }

        except Exception as e:
            logger.error(f"Error getting dashboard data: {e}")
            return {"error": str(e), "last_update": datetime.now().isoformat()}

    def get_analytics_report(self, hours: int = 24) -> Dict[str, Any]:
        """Generate comprehensive analytics report"""
        try:
            # Overall metrics summary
            summary = {}
            for metric_type in MetricType:
                stats = self.metrics_collector.get_statistics(metric_type, minutes=hours * 60)
                if stats["count"] > 0:
                    summary[metric_type.value] = stats

            # Anomaly detection
            anomalies = {}
            if self.config.get("enable_anomaly_detection", True):
                for metric_type in [
                    MetricType.CPU_USAGE,
                    MetricType.MEMORY_USAGE,
                    MetricType.ERROR_RATE,
                ]:
                    anomaly_result = self.predictive_analytics.detect_anomalies(metric_type)
                    if anomaly_result["anomalies"]:
                        anomalies[metric_type.value] = anomaly_result

            # Alert summary
            alert_history = self.alert_manager.get_alert_history(hours=hours)
            by_severity: Dict[str, int] = {}
            by_metric_type: Dict[str, int] = {}

            for alert in alert_history:
                severity_key = alert.severity.name
                by_severity[severity_key] = by_severity.get(severity_key, 0) + 1

                metric_key = alert.metric_type.value
                by_metric_type[metric_key] = by_metric_type.get(metric_key, 0) + 1

            alert_summary = {
                "total_alerts": len(alert_history),
                "by_severity": by_severity,
                "by_metric_type": by_metric_type,
                "resolved_count": sum(1 for a in alert_history if a.resolved),
                "acknowledged_count": sum(1 for a in alert_history if a.acknowledged),
            }

            return {
                "report_period_hours": hours,
                "generated_at": datetime.now().isoformat(),
                "metrics_summary": summary,
                "anomalies": anomalies,
                "alert_summary": alert_summary,
                "system_health": self.performance_monitor.get_system_health().to_dict(),
                "recommendations": self._generate_recommendations(summary, anomalies),
            }

        except Exception as e:
            logger.error(f"Error generating analytics report: {e}")
            return {"error": str(e), "generated_at": datetime.now().isoformat()}

    def _generate_recommendations(self, metrics_summary: Dict[str, Any], anomalies: Dict[str, Any]) -> List[str]:
        """Generate optimization recommendations based on metrics and anomalies"""
        recommendations = []

        # CPU recommendations
        if MetricType.CPU_USAGE.value in metrics_summary:
            cpu_stats = metrics_summary[MetricType.CPU_USAGE.value]
            if cpu_stats["average"] > 70:
                recommendations.append("High CPU usage detected. Consider optimizing code or scaling resources.")

        # Memory recommendations
        if MetricType.MEMORY_USAGE.value in metrics_summary:
            memory_stats = metrics_summary[MetricType.MEMORY_USAGE.value]
            if memory_stats["average"] > 80:
                recommendations.append(
                    "High memory usage detected. Consider memory optimization or increasing available memory."
                )

        # Error rate recommendations
        if MetricType.ERROR_RATE.value in metrics_summary:
            error_stats = metrics_summary[MetricType.ERROR_RATE.value]
            if error_stats["average"] > 5:
                recommendations.append(
                    "High error rate detected. Review error logs and implement better error handling."
                )

        # Anomaly recommendations
        if anomalies:
            recommendations.append(
                "Anomalies detected in system metrics. Review the detailed anomaly report for specific issues."
            )

        return recommendations


# Global instance for easy access
_monitoring_system: Optional[ComprehensiveMonitoringSystem] = None


def get_monitoring_system() -> ComprehensiveMonitoringSystem:
    """Get or create global monitoring system instance"""
    global _monitoring_system
    if _monitoring_system is None:
        _monitoring_system = ComprehensiveMonitoringSystem()
    return _monitoring_system


# Convenience functions
def start_monitoring() -> None:
    """Start the monitoring system"""
    system = get_monitoring_system()
    system.start()


def stop_monitoring() -> None:
    """Stop the monitoring system"""
    system = get_monitoring_system()
    system.stop()


def add_metric(
    metric_type: MetricType,
    value: Union[int, float],
    tags: Optional[Dict[str, str]] = None,
    source: str = "user",
) -> None:
    """Add a custom metric"""
    system = get_monitoring_system()
    system.add_metric(metric_type, value, tags, source)


def get_dashboard_data() -> Dict[str, Any]:
    """Get monitoring dashboard data"""
    system = get_monitoring_system()
    return system.get_dashboard_data()


if __name__ == "__main__":
    # Example usage
    print("Starting Comprehensive Monitoring System...")

    monitoring = ComprehensiveMonitoringSystem()
    monitoring.start()

    try:
        # Simulate some metrics
        for i in range(10):
            monitoring.add_metric(MetricType.CPU_USAGE, 50 + i * 3)
            monitoring.add_metric(MetricType.MEMORY_USAGE, 60 + i * 2)
            time.sleep(1)

        # Get dashboard data
        dashboard_data = monitoring.get_dashboard_data()
        print(f"System Health: {dashboard_data['health']['status']}")
        print(f"Overall Score: {dashboard_data['health']['overall_score']}")
        print(f"Active Alerts: {len(dashboard_data['active_alerts'])}")

        # Generate analytics report
        report = monitoring.get_analytics_report(hours=1)
        print(f"Analytics Report: {len(report['metrics_summary'])} metric types tracked")

    finally:
        monitoring.stop()
        print("Monitoring stopped.")
