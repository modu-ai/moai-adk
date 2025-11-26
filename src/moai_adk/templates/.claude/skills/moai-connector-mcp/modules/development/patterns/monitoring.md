# Monitoring & Observability Patterns

Production monitoring and observability for MCP servers.

---

## Health Checks

### Liveness Probe

```python
from datetime import datetime

@server.resource("health://live")
def liveness() -> dict:
    """Kubernetes liveness probe - is server running?"""
    return {
        "status": "alive",
        "timestamp": datetime.now().isoformat(),
        "version": "1.0.0"
    }
```

### Readiness Probe

```python
@server.resource("health://ready")
def readiness() -> dict:
    """Kubernetes readiness probe - can server handle requests?"""
    # Check critical dependencies
    checks = {
        "database": check_database(),
        "cache": check_cache(),
        "external_api": check_external_api()
    }

    # All checks must pass
    all_ready = all(checks.values())

    if not all_ready:
        raise ValueError(f"Service not ready. Checks: {checks}")

    return {
        "status": "ready",
        "checks": checks,
        "timestamp": datetime.now().isoformat()
    }
```

### Startup Probe

```python
@server.resource("health://startup")
def startup() -> dict:
    """Kubernetes startup probe - is server initialized?"""
    return {
        "status": "started",
        "initialization_complete": True,
        "timestamp": datetime.now().isoformat()
    }
```

---

## Structured Logging

### Logging Configuration

```python
import logging
import json
from datetime import datetime

# Configure structured logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)

logger = logging.getLogger(__name__)

class StructuredLogger:
    @staticmethod
    def log_event(
        event_type: str,
        status: str,
        duration_ms: float,
        details: dict = None
    ):
        """Log structured event."""
        log_entry = {
            "timestamp": datetime.now().isoformat(),
            "event_type": event_type,
            "status": status,
            "duration_ms": duration_ms,
            "details": details or {}
        }

        logger.info(json.dumps(log_entry))

# Usage
@server.tool()
def monitored_operation(param: str) -> dict:
    """Operation with structured logging."""
    import time

    start = time.time()

    try:
        result = execute_operation(param)
        duration = (time.time() - start) * 1000

        StructuredLogger.log_event(
            event_type="operation_executed",
            status="success",
            duration_ms=duration,
            details={"param": param}
        )

        return result

    except Exception as e:
        duration = (time.time() - start) * 1000

        StructuredLogger.log_event(
            event_type="operation_failed",
            status="error",
            duration_ms=duration,
            details={"param": param, "error": str(e)}
        )

        raise
```

---

## Metrics Collection

### Prometheus Metrics

```python
from prometheus_client import Counter, Histogram, Gauge
import time

# Define metrics
request_count = Counter(
    'mcp_tool_requests_total',
    'Total tool requests',
    ['tool_name', 'status']
)

request_duration = Histogram(
    'mcp_tool_duration_seconds',
    'Tool execution duration',
    ['tool_name'],
    buckets=(0.1, 0.5, 1.0, 2.0, 5.0)
)

active_requests = Gauge(
    'mcp_active_requests',
    'Active requests'
)

errors = Counter(
    'mcp_errors_total',
    'Total errors',
    ['error_type']
)

# Instrumented tool
@server.tool()
def measured_operation(param: str) -> dict:
    """Tool with metrics collection."""
    active_requests.inc()
    start = time.time()

    try:
        result = execute_operation(param)
        duration = time.time() - start

        request_count.labels(
            tool_name="measured_operation",
            status="success"
        ).inc()
        request_duration.labels(
            tool_name="measured_operation"
        ).observe(duration)

        return result

    except ValueError as e:
        errors.labels(error_type="validation_error").inc()
        raise

    except Exception as e:
        errors.labels(error_type="unknown").inc()
        raise

    finally:
        active_requests.dec()
```

### Metrics Endpoint

```python
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST

@server.resource("metrics://prometheus")
def prometheus_metrics() -> str:
    """Prometheus metrics endpoint."""
    return generate_latest().decode('utf-8')
```

---

## Request Tracking

### Request ID Propagation

```python
import uuid
from contextvars import ContextVar

# Context variable for request ID
request_id_context: ContextVar[str] = ContextVar('request_id')

def get_request_id() -> str:
    """Get current request ID."""
    try:
        return request_id_context.get()
    except LookupError:
        request_id = str(uuid.uuid4())
        request_id_context.set(request_id)
        return request_id

@server.tool()
def tracked_operation(param: str) -> dict:
    """Operation with request tracking."""
    request_id = get_request_id()

    logger.info(f"[{request_id}] Starting operation with param: {param}")

    try:
        result = execute_operation(param)
        logger.info(f"[{request_id}] Operation succeeded")
        return result

    except Exception as e:
        logger.error(f"[{request_id}] Operation failed: {str(e)}")
        raise
```

---

## Performance Monitoring

### Latency Tracking

```python
import time
from collections import deque

class LatencyTracker:
    def __init__(self, window_size: int = 100):
        self.latencies = deque(maxlen=window_size)

    def record(self, duration_ms: float):
        """Record operation duration."""
        self.latencies.append(duration_ms)

    def get_stats(self) -> dict:
        """Get latency statistics."""
        if not self.latencies:
            return {"p50": 0, "p95": 0, "p99": 0, "max": 0}

        sorted_latencies = sorted(self.latencies)
        n = len(sorted_latencies)

        return {
            "count": n,
            "min": min(self.latencies),
            "max": max(self.latencies),
            "p50": sorted_latencies[int(n * 0.5)],
            "p95": sorted_latencies[int(n * 0.95)],
            "p99": sorted_latencies[int(n * 0.99)],
            "avg": sum(self.latencies) / n
        }

tracker = LatencyTracker()

@server.tool()
def performance_tracked_tool(param: str) -> dict:
    """Tool with latency tracking."""
    start = time.time()
    result = execute_operation(param)
    duration = (time.time() - start) * 1000

    tracker.record(duration)
    return result

@server.resource("metrics://latency")
def get_latency_stats() -> dict:
    """Get latency statistics."""
    return tracker.get_stats()
```

---

## Alerting

### Threshold-Based Alerts

```python
class AlertManager:
    def __init__(self):
        self.alerts = []
        self.thresholds = {
            "error_rate": 0.05,  # 5%
            "p99_latency_ms": 5000,
            "active_requests": 1000
        }

    def check_health(self, metrics: dict) -> list[dict]:
        """Check metrics against thresholds."""
        alerts = []

        # Check error rate
        if metrics.get("error_rate", 0) > self.thresholds["error_rate"]:
            alerts.append({
                "severity": "critical",
                "metric": "error_rate",
                "value": metrics["error_rate"],
                "threshold": self.thresholds["error_rate"]
            })

        # Check latency
        if metrics.get("p99_latency_ms", 0) > self.thresholds["p99_latency_ms"]:
            alerts.append({
                "severity": "warning",
                "metric": "p99_latency_ms",
                "value": metrics["p99_latency_ms"],
                "threshold": self.thresholds["p99_latency_ms"]
            })

        # Check active requests
        if metrics.get("active_requests", 0) > self.thresholds["active_requests"]:
            alerts.append({
                "severity": "warning",
                "metric": "active_requests",
                "value": metrics["active_requests"],
                "threshold": self.thresholds["active_requests"]
            })

        self.alerts = alerts
        return alerts

alert_manager = AlertManager()

@server.resource("alerts://active")
def get_active_alerts() -> list[dict]:
    """Get active alerts."""
    return alert_manager.alerts
```

---

## Dashboard Metrics

### Comprehensive Dashboard

```python
@server.resource("dashboard://metrics")
def get_dashboard_metrics() -> dict:
    """Get metrics for dashboard."""
    return {
        "timestamp": datetime.now().isoformat(),
        "health": {
            "status": "healthy",
            "uptime_seconds": get_uptime(),
            "version": "1.0.0"
        },
        "performance": {
            "latency": tracker.get_stats(),
            "throughput_per_minute": calculate_throughput(),
            "active_requests": active_requests._value
        },
        "errors": {
            "total": errors_total._value,
            "recent_errors": get_recent_errors(10)
        },
        "resources": {
            "cpu_percent": get_cpu_usage(),
            "memory_mb": get_memory_usage(),
            "connections": get_connection_count()
        },
        "alerts": get_active_alerts()
    }
```

---

## Best Practices

✅ **Monitoring**:
- Implement health checks for Kubernetes
- Use structured logging
- Collect Prometheus metrics
- Track request IDs
- Monitor latency percentiles
- Set meaningful thresholds

✅ **Alerting**:
- Alert on error rates > 5%
- Alert on P99 latency > 5 seconds
- Monitor resource utilization
- Track active connections
- Log all authentication failures

❌ **Avoid**:
- Logging sensitive data
- Metrics with unbounded labels
- Silent failures without logging
- Ignoring performance degradation
- Collecting too much data

---

**Last Updated**: 2025-11-27
