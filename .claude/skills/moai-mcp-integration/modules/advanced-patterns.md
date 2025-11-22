# Advanced MCP Patterns

Enterprise-grade orchestration, monitoring, and optimization patterns.

---

## Multi-Server Orchestration

### Plugin Router Architecture

Dynamically route requests to the best-suited MCP server based on capabilities.

```python
from enum import Enum
from typing import Optional
import asyncio

class ServerCapability(str, Enum):
    """Available MCP server capabilities."""
    GITHUB = "github"
    SLACK = "slack"
    NOTION = "notion"
    FIREBASE = "firebase"
    PLAYWRIGHT = "playwright"

class ServerInfo:
    """Runtime information about an MCP server."""
    def __init__(self, name: str, capabilities: list[ServerCapability]):
        self.name = name
        self.capabilities = capabilities
        self.health_status = "healthy"
        self.latency_ms = 0
        self.error_count = 0

class MCPPluginOrchestrator:
    """Orchestrate multiple MCP servers."""

    def __init__(self):
        self.servers: dict[str, ServerInfo] = {}
        self.routes: dict[str, list[ServerCapability]] = {
            "repository_management": [ServerCapability.GITHUB],
            "notifications": [ServerCapability.SLACK],
            "documentation": [ServerCapability.NOTION],
            "database": [ServerCapability.FIREBASE],
            "web_automation": [ServerCapability.PLAYWRIGHT]
        }

    def discover_servers(self, config: dict) -> dict[str, ServerInfo]:
        """Discover available MCP servers from configuration."""
        discovered = {}

        for server_name, server_config in config.get("mcpServers", {}).items():
            capabilities = [
                ServerCapability(cap)
                for cap in server_config.get("capabilities", [])
            ]

            server = ServerInfo(server_name, capabilities)
            server.health_status = self.check_health(server_name)
            discovered[server_name] = server

        self.servers = discovered
        return discovered

    def find_capable_servers(self, required_capabilities: list[ServerCapability]) -> list[ServerInfo]:
        """Find all servers that support required capabilities."""
        capable = []

        for server in self.servers.values():
            if all(cap in server.capabilities for cap in required_capabilities):
                capable.append(server)

        return sorted(capable, key=lambda s: s.latency_ms)

    def route_request(self, request_type: str, **kwargs) -> dict:
        """Route request to appropriate server."""
        # Get required capabilities for this request type
        required_caps = self.routes.get(
            request_type,
            [ServerCapability.GITHUB]  # default
        )

        # Find capable servers
        capable_servers = self.find_capable_servers(required_caps)

        if not capable_servers:
            raise ValueError(f"No server supports: {required_caps}")

        # Select best server (lowest latency, healthy)
        selected = capable_servers[0]

        if selected.health_status != "healthy":
            # Try next best option
            for server in capable_servers[1:]:
                if server.health_status == "healthy":
                    selected = server
                    break

        # Execute request
        return self.execute_on_server(selected, request_type, **kwargs)

    def execute_on_server(self, server: ServerInfo, request_type: str, **kwargs) -> dict:
        """Execute request on specific server."""
        # Measure latency
        import time
        start = time.time()

        try:
            # Route and execute based on server type
            result = self.dispatch_request(server, request_type, **kwargs)
            server.latency_ms = (time.time() - start) * 1000
            server.error_count = 0
            return result

        except Exception as e:
            server.error_count += 1
            if server.error_count > 3:
                server.health_status = "unhealthy"
            raise

    def dispatch_request(self, server: ServerInfo, request_type: str, **kwargs) -> dict:
        """Dispatch to appropriate handler based on server type."""
        if ServerCapability.GITHUB in server.capabilities:
            return self.github_handler(request_type, **kwargs)
        elif ServerCapability.SLACK in server.capabilities:
            return self.slack_handler(request_type, **kwargs)
        elif ServerCapability.NOTION in server.capabilities:
            return self.notion_handler(request_type, **kwargs)
        else:
            raise ValueError(f"Unknown server: {server.name}")

    def github_handler(self, request_type: str, **kwargs) -> dict:
        """Handle GitHub-specific requests."""
        if request_type == "repository_management":
            return {
                "type": "github",
                "action": kwargs.get("action"),
                "result": "executed"
            }
        raise ValueError(f"Unknown request type: {request_type}")

    def slack_handler(self, request_type: str, **kwargs) -> dict:
        """Handle Slack-specific requests."""
        if request_type == "notifications":
            return {
                "type": "slack",
                "channel": kwargs.get("channel"),
                "status": "sent"
            }
        raise ValueError(f"Unknown request type: {request_type}")

    def notion_handler(self, request_type: str, **kwargs) -> dict:
        """Handle Notion-specific requests."""
        if request_type == "documentation":
            return {
                "type": "notion",
                "page_id": kwargs.get("page_id"),
                "status": "created"
            }
        raise ValueError(f"Unknown request type: {request_type}")

    def check_health(self, server_name: str) -> str:
        """Check server health status."""
        # Implement actual health check logic
        return "healthy"
```

### Usage Example

```python
# Initialize orchestrator
orchestrator = MCPPluginOrchestrator()

# Load configuration
config = {
    "mcpServers": {
        "github": {"capabilities": ["github"]},
        "slack": {"capabilities": ["slack"]},
        "notion": {"capabilities": ["notion"]}
    }
}

# Discover servers
orchestrator.discover_servers(config)

# Route requests automatically
result = orchestrator.route_request(
    "repository_management",
    action="create_issue"
)

# Explicitly request specific capability
github_servers = orchestrator.find_capable_servers(
    [ServerCapability.GITHUB]
)
```

---

## Performance Monitoring

### Metrics Collection Framework

Comprehensive monitoring for MCP server operations.

```python
from dataclasses import dataclass
from datetime import datetime
from typing import Optional
import json

@dataclass
class OperationMetrics:
    """Metrics for a single operation."""
    operation_name: str
    server_name: str
    duration_ms: float
    success: bool
    error_message: Optional[str] = None
    timestamp: datetime = None
    request_size_bytes: int = 0
    response_size_bytes: int = 0

    def __post_init__(self):
        if self.timestamp is None:
            self.timestamp = datetime.now()

class MetricsCollector:
    """Collect and analyze operation metrics."""

    def __init__(self, max_records: int = 10000):
        self.max_records = max_records
        self.metrics: list[OperationMetrics] = []
        self.alerts: list[dict] = []

    def record_operation(self, **kwargs) -> OperationMetrics:
        """Record a completed operation."""
        metrics = OperationMetrics(**kwargs)
        self.metrics.append(metrics)

        # Maintain size limit
        if len(self.metrics) > self.max_records:
            self.metrics.pop(0)

        # Check for anomalies
        self.check_anomalies(metrics)

        return metrics

    def check_anomalies(self, metrics: OperationMetrics):
        """Check for performance anomalies."""
        # High latency alert
        if metrics.duration_ms > 5000 and metrics.success:
            self.alerts.append({
                "type": "high_latency",
                "severity": "warning",
                "operation": metrics.operation_name,
                "server": metrics.server_name,
                "duration_ms": metrics.duration_ms
            })

        # Error rate alert
        recent_errors = sum(
            1 for m in self.metrics[-100:]
            if m.server_name == metrics.server_name and not m.success
        )
        if recent_errors > 10:  # 10% error rate
            self.alerts.append({
                "type": "high_error_rate",
                "severity": "critical",
                "server": metrics.server_name,
                "error_count": recent_errors
            })

    def get_statistics(self, server_name: Optional[str] = None) -> dict:
        """Get aggregated statistics."""
        filtered = [
            m for m in self.metrics
            if server_name is None or m.server_name == server_name
        ]

        if not filtered:
            return {}

        durations = [m.duration_ms for m in filtered]
        successful = [m for m in filtered if m.success]

        return {
            "total_operations": len(filtered),
            "successful": len(successful),
            "failed": len(filtered) - len(successful),
            "success_rate": len(successful) / len(filtered) if filtered else 0,
            "avg_duration_ms": sum(durations) / len(durations),
            "p50_duration_ms": sorted(durations)[len(durations) // 2],
            "p99_duration_ms": sorted(durations)[int(len(durations) * 0.99)],
            "min_duration_ms": min(durations),
            "max_duration_ms": max(durations),
        }

    def get_health_dashboard(self) -> dict:
        """Generate health dashboard."""
        servers = set(m.server_name for m in self.metrics)

        return {
            "timestamp": datetime.now().isoformat(),
            "total_operations": len(self.metrics),
            "servers": {
                server: self.get_statistics(server)
                for server in servers
            },
            "active_alerts": self.alerts[-10:],  # Last 10 alerts
            "system_health": self.calculate_health_score()
        }

    def calculate_health_score(self) -> str:
        """Calculate overall system health score."""
        if not self.metrics:
            return "unknown"

        successful = sum(1 for m in self.metrics if m.success)
        success_rate = successful / len(self.metrics)

        if success_rate >= 0.99:
            return "excellent"
        elif success_rate >= 0.95:
            return "good"
        elif success_rate >= 0.90:
            return "fair"
        else:
            return "poor"

class HealthMonitor:
    """Real-time health monitoring with alerts."""

    def __init__(self, metrics_collector: MetricsCollector):
        self.collector = metrics_collector
        self.thresholds = {
            "error_rate": 0.05,  # 5%
            "latency_p99": 5000,  # 5 seconds
            "latency_p50": 1000,  # 1 second
        }

    def check_health(self) -> dict:
        """Check system health against thresholds."""
        stats = self.collector.get_statistics()

        if not stats:
            return {"status": "no_data"}

        issues = []

        # Check error rate
        if stats.get("success_rate", 1) < (1 - self.thresholds["error_rate"]):
            issues.append({
                "type": "high_error_rate",
                "value": stats["success_rate"],
                "threshold": 1 - self.thresholds["error_rate"]
            })

        # Check latency
        if stats.get("p99_duration_ms", 0) > self.thresholds["latency_p99"]:
            issues.append({
                "type": "high_latency_p99",
                "value": stats["p99_duration_ms"],
                "threshold": self.thresholds["latency_p99"]
            })

        return {
            "status": "healthy" if not issues else "unhealthy",
            "issues": issues,
            "timestamp": datetime.now().isoformat()
        }
```

### Usage Example

```python
# Initialize monitoring
collector = MetricsCollector()
monitor = HealthMonitor(collector)

# Record operations
collector.record_operation(
    operation_name="query_database",
    server_name="firebase",
    duration_ms=234,
    success=True,
    response_size_bytes=1024
)

# Get statistics
stats = collector.get_statistics("firebase")
print(f"Success rate: {stats['success_rate']:.1%}")
print(f"P99 latency: {stats['p99_duration_ms']}ms")

# Check health
health = monitor.check_health()
print(f"System status: {health['status']}")

# Get dashboard
dashboard = collector.get_health_dashboard()
print(json.dumps(dashboard, indent=2, default=str))
```

---

## Caching Strategy

### Distributed Cache Pattern

```python
from functools import wraps
from datetime import datetime, timedelta
import hashlib

class CacheEntry:
    """Single cache entry with TTL."""
    def __init__(self, value: any, ttl_seconds: int = 300):
        self.value = value
        self.created_at = datetime.now()
        self.ttl_seconds = ttl_seconds

    def is_expired(self) -> bool:
        """Check if cache entry has expired."""
        age = (datetime.now() - self.created_at).total_seconds()
        return age > self.ttl_seconds

class MCPCache:
    """Distributed cache for MCP operations."""

    def __init__(self, max_size: int = 1000):
        self.cache: dict[str, CacheEntry] = {}
        self.max_size = max_size

    def key(self, *args, **kwargs) -> str:
        """Generate cache key from arguments."""
        key_str = str(args) + str(sorted(kwargs.items()))
        return hashlib.md5(key_str.encode()).hexdigest()

    def get(self, key: str):
        """Retrieve from cache."""
        if key not in self.cache:
            return None

        entry = self.cache[key]
        if entry.is_expired():
            del self.cache[key]
            return None

        return entry.value

    def set(self, key: str, value: any, ttl_seconds: int = 300):
        """Store in cache."""
        if len(self.cache) >= self.max_size:
            # Evict oldest expired entries
            expired_keys = [
                k for k, v in self.cache.items()
                if v.is_expired()
            ]
            for k in expired_keys:
                del self.cache[k]

        self.cache[key] = CacheEntry(value, ttl_seconds)

    def cached(self, ttl_seconds: int = 300):
        """Decorator for caching function results."""
        def decorator(func):
            @wraps(func)
            def wrapper(*args, **kwargs):
                cache_key = self.key(*args, **kwargs)
                cached_value = self.get(cache_key)

                if cached_value is not None:
                    return cached_value

                result = func(*args, **kwargs)
                self.set(cache_key, result, ttl_seconds)
                return result

            return wrapper
        return decorator

# Usage
cache = MCPCache()

@cache.cached(ttl_seconds=600)
def get_user_profile(user_id: str) -> dict:
    """Get user profile (cached for 10 minutes)."""
    return fetch_from_database(user_id)
```

---

## Circuit Breaker Pattern

```python
from enum import Enum
from datetime import datetime, timedelta

class CircuitState(str, Enum):
    """Circuit breaker states."""
    CLOSED = "closed"      # Normal operation
    OPEN = "open"         # Failing, reject requests
    HALF_OPEN = "half_open"  # Testing recovery

class CircuitBreaker:
    """Protect against cascading failures."""

    def __init__(
        self,
        failure_threshold: int = 5,
        reset_timeout_seconds: int = 60
    ):
        self.state = CircuitState.CLOSED
        self.failure_count = 0
        self.failure_threshold = failure_threshold
        self.reset_timeout_seconds = reset_timeout_seconds
        self.last_failure_time = None

    def call(self, func, *args, **kwargs):
        """Execute function with circuit breaker protection."""
        if self.state == CircuitState.OPEN:
            if self.should_attempt_reset():
                self.state = CircuitState.HALF_OPEN
            else:
                raise Exception("Circuit breaker is OPEN")

        try:
            result = func(*args, **kwargs)
            self.on_success()
            return result

        except Exception as e:
            self.on_failure()
            raise

    def on_success(self):
        """Handle successful call."""
        self.failure_count = 0
        self.state = CircuitState.CLOSED

    def on_failure(self):
        """Handle failed call."""
        self.failure_count += 1
        self.last_failure_time = datetime.now()

        if self.failure_count >= self.failure_threshold:
            self.state = CircuitState.OPEN

    def should_attempt_reset(self) -> bool:
        """Check if enough time has passed to attempt reset."""
        if self.last_failure_time is None:
            return False

        reset_time = self.last_failure_time + timedelta(
            seconds=self.reset_timeout_seconds
        )
        return datetime.now() >= reset_time
```

---

## Rate Limiting Pattern

```python
from collections import deque
from datetime import datetime, timedelta

class RateLimiter:
    """Token bucket rate limiting."""

    def __init__(self, requests_per_second: float = 10):
        self.requests_per_second = requests_per_second
        self.requests: deque = deque()
        self.max_tokens = requests_per_second

    def is_allowed(self) -> bool:
        """Check if request is allowed."""
        now = datetime.now()

        # Remove old requests outside window
        window_start = now - timedelta(seconds=1)
        while self.requests and self.requests[0] < window_start:
            self.requests.popleft()

        # Check if limit exceeded
        if len(self.requests) < self.max_tokens:
            self.requests.append(now)
            return True

        return False

    def wait_if_needed(self):
        """Wait until request is allowed."""
        while not self.is_allowed():
            time.sleep(0.01)
```

---

**Last Updated**: 2025-11-22 | Advanced Patterns for Production Use
