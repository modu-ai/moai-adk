"""
JIT-Enhanced Hook Manager

Integrates Phase 2 JIT Context Loading System with Claude Code hook infrastructure
to provide intelligent, phase-aware hook execution with optimal performance.

Key Features:
- Phase-based hook optimization
- JIT context loading for hooks
- Intelligent skill filtering for hook operations
- Dynamic token budget management
- Real-time performance monitoring
- Smart caching and invalidation
"""

import asyncio
import inspect
import json
import logging
import threading
import time
from collections import defaultdict
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
from pathlib import Path
from typing import Any, Callable, Dict, List, Optional, Set, Tuple

# Import JIT Context Loading System from Phase 2
try:
    from .jit_context_loader import (
        ContextCache as _ImportedContextCache,
    )
    from .jit_context_loader import (
        JITContextLoader as _ImportedJITContextLoader,
    )
    from .jit_context_loader import (
        Phase as _ImportedPhase,
    )
    from .jit_context_loader import (
        TokenBudgetManager as _ImportedTokenBudgetManager,
    )

    JITContextLoader = _ImportedJITContextLoader
    ContextCache = _ImportedContextCache
    TokenBudgetManager = _ImportedTokenBudgetManager
    Phase = _ImportedPhase
    _JIT_AVAILABLE = True
except ImportError:
    _JIT_AVAILABLE = False

    # Fallback for environments where JIT system might not be available
    class JITContextLoader:  # type: ignore[no-redef]
        def __init__(self, *args: Any, **kwargs: Any) -> None:
            pass

    class ContextCache:  # type: ignore[no-redef]
        def __init__(self, max_size: int = 100, max_memory_mb: int = 50) -> None:
            self.max_size = max_size
            self.max_memory_mb = max_memory_mb
            self.hits = 0
            self.misses = 0
            self.cache: dict[Any, Any] = {}

        def get(self, key: Any) -> Any:
            self.misses += 1
            return None

        def put(self, key: Any, value: Any, token_count: int = 0) -> None:
            pass

        def clear(self) -> None:
            pass

        def get_stats(self) -> dict[str, Any]:
            return {"hits": self.hits, "misses": self.misses}

    class TokenBudgetManager:  # type: ignore[no-redef]
        def __init__(self, *args: Any, **kwargs: Any) -> None:
            pass

    # Create Phase enum for hook system (fallback)
    class Phase(Enum):  # type: ignore[no-redef]
        SPEC = "SPEC"
        RED = "RED"
        GREEN = "GREEN"
        REFACTOR = "REFACTOR"
        SYNC = "SYNC"
        DEBUG = "DEBUG"
        PLANNING = "PLANNING"


class HookEvent(Enum):
    """Hook event types from Claude Code"""

    SESSION_START = "SessionStart"
    SESSION_END = "SessionEnd"
    USER_PROMPT_SUBMIT = "UserPromptSubmit"
    PRE_TOOL_USE = "PreToolUse"
    POST_TOOL_USE = "PostToolUse"
    SUBAGENT_START = "SubagentStart"
    SUBAGENT_STOP = "SubagentStop"


class HookPriority(Enum):
    """Hook execution priority levels"""

    CRITICAL = 1  # System-critical hooks (security, validation)
    HIGH = 2  # High-impact hooks (performance optimization)
    NORMAL = 3  # Standard hooks (logging, cleanup)
    LOW = 4  # Optional hooks (analytics, metrics)


@dataclass
class HookMetadata:
    """Metadata for a hook execution"""

    hook_path: str
    event_type: HookEvent
    priority: HookPriority
    estimated_execution_time_ms: float = 0.0
    last_execution_time: Optional[datetime] = None
    success_rate: float = 1.0
    phase_relevance: Dict[Phase, float] = field(default_factory=dict)
    token_cost_estimate: int = 0
    dependencies: Set[str] = field(default_factory=set)
    parallel_safe: bool = True


@dataclass
class HookExecutionResult:
    """Result of hook execution"""

    hook_path: str
    success: bool
    execution_time_ms: float
    token_usage: int
    output: Any
    error_message: Optional[str] = None
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class CircuitBreakerState:
    """Circuit breaker state for hook resilience"""

    failure_count: int = 0
    last_failure_time: Optional[datetime] = None
    state: str = "CLOSED"  # CLOSED, OPEN, HALF_OPEN
    success_threshold: int = 5
    failure_threshold: int = 3
    timeout_seconds: int = 60


class CircuitBreaker:
    """Circuit breaker pattern for failing hooks"""

    def __init__(
        self,
        failure_threshold: int = 3,
        timeout_seconds: int = 60,
        success_threshold: int = 5,
    ):
        self.failure_threshold = failure_threshold
        self.timeout_seconds = timeout_seconds
        self.success_threshold = success_threshold
        self.state = CircuitBreakerState(
            failure_threshold=failure_threshold,
            timeout_seconds=timeout_seconds,
            success_threshold=success_threshold,
        )

    async def call(self, func: Callable, *args, **kwargs) -> Any:
        """Execute function with circuit breaker protection"""
        if self.state.state == "OPEN":
            if self._should_attempt_reset():
                self.state.state = "HALF_OPEN"
            else:
                raise Exception("Circuit breaker is OPEN - call blocked")

        try:
            result = await func(*args, **kwargs) if inspect.iscoroutinefunction(func) else func(*args, **kwargs)
            self._on_success()
            return result
        except Exception as e:
            self._on_failure()
            raise e

    def _should_attempt_reset(self) -> bool:
        """Check if circuit breaker should attempt reset"""
        if self.state.last_failure_time is None:
            return False
        return datetime.now() - self.state.last_failure_time > timedelta(seconds=self.timeout_seconds)

    def _on_success(self) -> None:
        """Handle successful call"""
        self.state.failure_count = 0
        if self.state.state == "HALF_OPEN":
            self.state.success_threshold -= 1
            if self.state.success_threshold <= 0:
                self.state.state = "CLOSED"
                self.state.success_threshold = 5

    def _on_failure(self) -> None:
        """Handle failed call"""
        self.state.failure_count += 1
        self.state.last_failure_time = datetime.now()

        if self.state.failure_count >= self.failure_threshold:
            self.state.state = "OPEN"


class HookResultCache:
    """Advanced result caching with TTL and invalidation"""

    def __init__(self, max_size: int = 1000, default_ttl_seconds: int = 300):
        self.max_size = max_size
        self.default_ttl_seconds = default_ttl_seconds
        self._cache: Dict[str, Tuple[Any, datetime, int]] = {}  # key -> (value, expiry, access_count)
        self._access_times: Dict[str, datetime] = {}
        self._lock = threading.RLock()

    def get(self, key: str) -> Optional[Any]:
        """Get cached value if valid"""
        with self._lock:
            if key not in self._cache:
                return None

            value, expiry, access_count = self._cache[key]

            # Check TTL
            if datetime.now() > expiry:
                del self._cache[key]
                if key in self._access_times:
                    del self._access_times[key]
                return None

            # Update access
            self._cache[key] = (value, expiry, access_count + 1)
            self._access_times[key] = datetime.now()

            return value

    def put(self, key: str, value: Any, ttl_seconds: Optional[int] = None) -> None:
        """Cache value with TTL"""
        with self._lock:
            # Enforce size limit
            if len(self._cache) >= self.max_size:
                self._evict_lru()

            ttl = ttl_seconds or self.default_ttl_seconds
            expiry = datetime.now() + timedelta(seconds=ttl)
            self._cache[key] = (value, expiry, 1)
            self._access_times[key] = datetime.now()

    def invalidate(self, pattern: Optional[str] = None) -> None:
        """Invalidate cache entries"""
        with self._lock:
            if pattern is None:
                self._cache.clear()
                self._access_times.clear()
            else:
                keys_to_remove = [key for key in self._cache.keys() if pattern in key]
                for key in keys_to_remove:
                    del self._cache[key]
                    if key in self._access_times:
                        del self._access_times[key]

    def _evict_lru(self) -> None:
        """Evict least recently used item"""
        if not self._access_times:
            return

        lru_key = min(self._access_times.keys(), key=lambda k: self._access_times[k])
        del self._cache[lru_key]
        del self._access_times[lru_key]

    def get_stats(self) -> Dict[str, Any]:
        """Get cache statistics"""
        with self._lock:
            return {
                "size": len(self._cache),
                "max_size": self.max_size,
                "utilization": len(self._cache) / self.max_size,
            }


class ConnectionPool:
    """Connection pooling for MCP servers and external resources"""

    def __init__(self, max_connections: int = 10, connection_timeout_seconds: int = 30):
        self.max_connections = max_connections
        self.connection_timeout_seconds = connection_timeout_seconds
        self._pools: Dict[str, List] = defaultdict(list)
        self._active_connections: Dict[str, int] = defaultdict(int)
        self._lock = threading.Lock()

    async def get_connection(self, pool_name: str, connection_factory: Callable) -> Any:
        """Get connection from pool or create new one"""
        with self._lock:
            # Check pool for available connection
            pool = self._pools[pool_name]
            if pool:
                connection = pool.pop()
                self._active_connections[pool_name] += 1
                return connection

            # Check if we can create new connection
            if self._active_connections[pool_name] >= self.max_connections:
                raise Exception(f"Connection pool '{pool_name}' is full")

            self._active_connections[pool_name] += 1

        # Create new connection outside of lock
        try:
            connection = (
                await connection_factory() if inspect.iscoroutinefunction(connection_factory) else connection_factory()
            )
            return connection
        except Exception:
            with self._lock:
                self._active_connections[pool_name] -= 1
            raise

    def return_connection(self, pool_name: str, connection: Any) -> None:
        """Return connection to pool"""
        with self._lock:
            pool = self._pools[pool_name]
            if len(pool) < self.max_connections and connection is not None:
                pool.append(connection)
            self._active_connections[pool_name] -= 1

    def get_pool_stats(self) -> Dict[str, Any]:
        """Get connection pool statistics"""
        with self._lock:
            return {
                "pools": {
                    name: {
                        "available": len(pool),
                        "active": self._active_connections[name],
                        "total": len(pool) + self._active_connections[name],
                    }
                    for name, pool in self._pools.items()
                }
            }


class RetryPolicy:
    """Exponential backoff retry policy"""

    def __init__(
        self,
        max_retries: int = 3,
        base_delay_ms: float = 100,
        max_delay_ms: float = 5000,
        backoff_factor: float = 2.0,
    ):
        self.max_retries = max_retries
        self.base_delay_ms = base_delay_ms
        self.max_delay_ms = max_delay_ms
        self.backoff_factor = backoff_factor

    async def execute_with_retry(self, func: Callable, *args, **kwargs) -> Any:
        """Execute function with retry policy"""
        last_exception = None

        for attempt in range(self.max_retries + 1):
            try:
                result = await func(*args, **kwargs) if inspect.iscoroutinefunction(func) else func(*args, **kwargs)
                return result
            except Exception as e:
                last_exception = e

                if attempt < self.max_retries:
                    delay_ms = min(
                        self.base_delay_ms * (self.backoff_factor**attempt),
                        self.max_delay_ms,
                    )
                    await asyncio.sleep(delay_ms / 1000.0)
                else:
                    break

        raise last_exception


@dataclass
class ResourceUsageMetrics:
    """Resource usage metrics for monitoring"""

    cpu_usage_percent: float = 0.0
    memory_usage_mb: float = 0.0
    disk_io_mb: float = 0.0
    network_io_mb: float = 0.0
    open_files: int = 0
    thread_count: int = 0


@dataclass
class HookPerformanceMetrics:
    """Performance metrics for hook system"""

    total_executions: int = 0
    successful_executions: int = 0
    average_execution_time_ms: float = 0.0
    total_token_usage: int = 0
    cache_hits: int = 0
    cache_misses: int = 0
    phase_distribution: Dict[Phase, int] = field(default_factory=dict)
    event_type_distribution: Dict[HookEvent, int] = field(default_factory=dict)
    circuit_breaker_trips: int = 0
    retry_attempts: int = 0
    resource_usage: ResourceUsageMetrics = field(default_factory=ResourceUsageMetrics)


class ResourceMonitor:
    """Resource usage monitoring for hook system"""

    def __init__(self):
        self._baseline_metrics = self._get_current_metrics()
        self._peak_usage = ResourceUsageMetrics()

    def get_current_metrics(self) -> ResourceUsageMetrics:
        """Get current resource usage metrics"""
        import os

        import psutil

        try:
            process = psutil.Process(os.getpid())

            # Memory usage
            memory_info = process.memory_info()
            memory_mb = memory_info.rss / 1024 / 1024

            # CPU usage (average across all cores)
            cpu_percent = process.cpu_percent()

            # Thread count
            thread_count = process.num_threads()

            # Open file descriptors
            try:
                open_files = process.num_fds()
            except (AttributeError, psutil.AccessDenied):
                open_files = 0

            # Update peak usage
            self._peak_usage.memory_usage_mb = max(self._peak_usage.memory_usage_mb, memory_mb)
            self._peak_usage.cpu_usage_percent = max(self._peak_usage.cpu_usage_percent, cpu_percent)
            self._peak_usage.thread_count = max(self._peak_usage.thread_count, thread_count)
            self._peak_usage.open_files = max(self._peak_usage.open_files, open_files)

            return ResourceUsageMetrics(
                cpu_usage_percent=cpu_percent,
                memory_usage_mb=memory_mb,
                thread_count=thread_count,
                open_files=open_files,
            )
        except Exception:
            return ResourceUsageMetrics()

    def _get_current_metrics(self) -> ResourceUsageMetrics:
        """Get baseline metrics for comparison"""
        return self.get_current_metrics()

    def get_peak_metrics(self) -> ResourceUsageMetrics:
        """Get peak resource usage metrics"""
        return self._peak_usage


class HealthChecker:
    """Health monitoring and check endpoints for hook system"""

    def __init__(self, hook_manager: "JITEnhancedHookManager"):
        self.hook_manager = hook_manager
        self._last_health_check = datetime.now()
        self._health_status = "healthy"

    async def check_system_health(self) -> Dict[str, Any]:
        """Perform comprehensive health check"""
        checks: Dict[str, Dict[str, Any]] = {}
        health_report: Dict[str, Any] = {
            "status": "healthy",
            "timestamp": datetime.now().isoformat(),
            "checks": checks,
        }

        try:
            # Check hook registry
            checks["hook_registry"] = {
                "status": ("healthy" if len(self.hook_manager._hook_registry) > 0 else "warning"),
                "registered_hooks": len(self.hook_manager._hook_registry),
                "events_supported": len(self.hook_manager._hooks_by_event),
            }

            # Check cache health
            cache_stats = self.hook_manager._advanced_cache.get_stats()
            checks["cache"] = {
                "status": "healthy",
                "size": cache_stats["size"],
                "utilization": cache_stats["utilization"],
                "max_size": cache_stats["max_size"],
            }

            # Check connection pools
            pool_stats = self.hook_manager._connection_pool.get_pool_stats()
            checks["connection_pools"] = {
                "status": "healthy",
                "pools": pool_stats["pools"],
            }

            # Check circuit breakers
            tripped_breakers = [
                name for name, cb in self.hook_manager._circuit_breakers.items() if cb.state.state == "OPEN"
            ]

            checks["circuit_breakers"] = {
                "status": "healthy" if len(tripped_breakers) == 0 else "degraded",
                "total_breakers": len(self.hook_manager._circuit_breakers),
                "tripped_breakers": len(tripped_breakers),
                "tripped_breaker_names": tripped_breakers,
            }

            # Check resource usage
            resource_metrics = self.hook_manager._resource_monitor.get_current_metrics()
            checks["resource_usage"] = {
                "status": "healthy",
                "memory_mb": resource_metrics.memory_usage_mb,
                "cpu_percent": resource_metrics.cpu_usage_percent,
                "thread_count": resource_metrics.thread_count,
            }

            # Overall status determination
            statuses = [check["status"] for check in checks.values()]
            if "unhealthy" in statuses:
                health_report["status"] = "unhealthy"
            elif "degraded" in statuses or "warning" in statuses:
                health_report["status"] = "degraded"

            self._health_status = health_report["status"]
            self._last_health_check = datetime.now()

        except Exception as e:
            health_report["status"] = "unhealthy"
            health_report["error"] = str(e)

        return health_report

    def get_health_status(self) -> str:
        """Get current health status"""
        return self._health_status


class PerformanceAnomalyDetector:
    """Detect performance anomalies in hook execution"""

    def __init__(self, sensitivity_factor: float = 2.0):
        self.sensitivity_factor = sensitivity_factor
        self._performance_history: Dict[str, List[float]] = defaultdict(list)

    def detect_anomaly(self, hook_path: str, execution_time_ms: float) -> Optional[Dict[str, Any]]:
        """Detect if execution time is anomalous"""
        history = self._performance_history[hook_path]

        if len(history) < 5:
            # Not enough data for detection
            self._performance_history[hook_path].append(execution_time_ms)
            return None

        # Calculate statistics
        mean_time = sum(history) / len(history)
        variance = sum((x - mean_time) ** 2 for x in history) / len(history)
        std_dev = variance**0.5

        # Check for anomaly
        if abs(execution_time_ms - mean_time) > (self.sensitivity_factor * std_dev):
            anomaly_type = "slow" if execution_time_ms > mean_time else "fast"
            return {
                "hook_path": hook_path,
                "anomaly_type": anomaly_type,
                "execution_time_ms": execution_time_ms,
                "mean_time_ms": mean_time,
                "std_dev_ms": std_dev,
                "deviation_factor": abs(execution_time_ms - mean_time) / std_dev,
                "severity": ("high" if abs(execution_time_ms - mean_time) > (3 * std_dev) else "medium"),
            }

        # Update history (keep last 50 entries)
        history.append(execution_time_ms)
        if len(history) > 50:
            history.pop(0)

        return None


class JITEnhancedHookManager:
    """
    Enhanced Hook Manager with JIT Context Loading System integration

    Provides intelligent hook execution with phase-aware optimization,
    token budget management, performance monitoring, and reliability patterns.
    """

    def __init__(
        self,
        hooks_directory: Optional[Path] = None,
        cache_directory: Optional[Path] = None,
        max_concurrent_hooks: int = 5,
        enable_performance_monitoring: bool = True,
        cache_ttl_seconds: int = 300,
        circuit_breaker_threshold: int = 3,
        max_retries: int = 3,
        connection_pool_size: int = 10,
    ):
        """Initialize JIT-Enhanced Hook Manager with Phase 2 optimizations

        Args:
            hooks_directory: Directory containing hook files
            cache_directory: Directory for hook cache and performance data
            max_concurrent_hooks: Maximum number of hooks to execute concurrently
            enable_performance_monitoring: Enable detailed performance tracking
            cache_ttl_seconds: Default TTL for cached results
            circuit_breaker_threshold: Failure threshold for circuit breaker
            max_retries: Maximum retry attempts for failed hooks
            connection_pool_size: Size of connection pool for external resources
        """
        self.hooks_directory = hooks_directory or Path.cwd() / ".claude" / "hooks"
        self.cache_directory = cache_directory or Path.cwd() / ".moai" / "cache" / "hooks"
        self.max_concurrent_hooks = max_concurrent_hooks
        self.enable_performance_monitoring = enable_performance_monitoring

        # Initialize JIT Context Loading System
        self.jit_loader = JITContextLoader()

        # Initialize Phase 2 optimizations
        self._initialize_phase2_optimizations(
            cache_ttl_seconds,
            circuit_breaker_threshold,
            max_retries,
            connection_pool_size,
        )

        # Initialize caches and metadata storage
        self._initialize_caches()

        # Performance tracking
        self.metrics = HookPerformanceMetrics()
        self._performance_lock = threading.Lock()
        self._resource_monitor = ResourceMonitor()

        # Hook registry with metadata
        self._hook_registry: Dict[str, HookMetadata] = {}
        self._hooks_by_event: Dict[HookEvent, List[str]] = {}
        self._circuit_breakers: Dict[str, CircuitBreaker] = {}
        self._retry_policies: Dict[str, RetryPolicy] = {}

        # Initialize hook registry
        self._discover_hooks()

        # Setup health monitoring
        self._health_checker = HealthChecker(self)
        self._logger = logging.getLogger(__name__)

    def _initialize_phase2_optimizations(
        self,
        cache_ttl_seconds: int,
        circuit_breaker_threshold: int,
        max_retries: int,
        connection_pool_size: int,
    ) -> None:
        """Initialize Phase 2 optimization components"""
        # Advanced result cache with TTL
        self._advanced_cache = HookResultCache(max_size=1000, default_ttl_seconds=cache_ttl_seconds)

        # Connection pooling for MCP servers and external resources
        self._connection_pool = ConnectionPool(max_connections=connection_pool_size)

        # Circuit breaker and retry policies
        self.circuit_breaker_threshold = circuit_breaker_threshold
        self.max_retries = max_retries

        # Performance profiling
        self._execution_profiles: Dict[str, List[float]] = defaultdict(list)
        self._anomaly_detector = PerformanceAnomalyDetector()

    def _initialize_caches(self) -> None:
        """Initialize cache directories and data structures"""
        self.cache_directory.mkdir(parents=True, exist_ok=True)

        # Initialize hook result cache
        self._result_cache = ContextCache(max_size=100, max_memory_mb=50)

        # Initialize metadata cache
        self._metadata_cache: Dict[str, Dict[str, Any]] = {}

        # Performance log file
        self._performance_log_path = self.cache_directory / "performance.jsonl"

    def _discover_hooks(self) -> None:
        """Discover and register all available hooks"""
        if not self.hooks_directory.exists():
            return

        for hook_file in self.hooks_directory.rglob("*.py"):
            if hook_file.name.startswith("__") or hook_file.name.startswith("lib/"):
                continue

            hook_path_str = str(hook_file.relative_to(self.hooks_directory))

            # Extract event type from filename
            event_type = self._extract_event_type_from_filename(hook_file.name)
            if event_type:
                self._register_hook(hook_path_str, event_type)

    def _extract_event_type_from_filename(self, filename: str) -> Optional[HookEvent]:
        """Extract hook event type from filename pattern"""
        filename_lower = filename.lower()

        if "session_start" in filename_lower:
            return HookEvent.SESSION_START
        elif "session_end" in filename_lower:
            return HookEvent.SESSION_END
        elif "pre_tool" in filename_lower or "pretool" in filename_lower:
            return HookEvent.PRE_TOOL_USE
        elif "post_tool" in filename_lower or "posttool" in filename_lower:
            return HookEvent.POST_TOOL_USE
        elif "subagent_start" in filename_lower:
            return HookEvent.SUBAGENT_START
        elif "subagent_stop" in filename_lower:
            return HookEvent.SUBAGENT_STOP
        else:
            return None

    def _register_hook(self, hook_path: str, event_type: HookEvent) -> None:
        """Register a hook with metadata"""
        # Generate metadata based on hook characteristics
        metadata = HookMetadata(
            hook_path=hook_path,
            event_type=event_type,
            priority=self._determine_hook_priority(hook_path, event_type),
            estimated_execution_time_ms=self._estimate_execution_time(hook_path),
            phase_relevance=self._determine_phase_relevance(hook_path, event_type),
            token_cost_estimate=self._estimate_token_cost(hook_path),
            parallel_safe=self._is_parallel_safe(hook_path),
        )

        self._hook_registry[hook_path] = metadata

        if event_type not in self._hooks_by_event:
            self._hooks_by_event[event_type] = []
        self._hooks_by_event[event_type].append(hook_path)

    def _determine_hook_priority(self, hook_path: str, event_type: HookEvent) -> HookPriority:
        """Determine hook priority based on its characteristics"""
        filename = hook_path.lower()

        # Security and validation hooks are critical
        if any(keyword in filename for keyword in ["security", "validation", "health_check"]):
            return HookPriority.CRITICAL

        # Performance optimization hooks are high priority
        if any(keyword in filename for keyword in ["performance", "optimizer", "jit"]):
            return HookPriority.HIGH

        # Cleanup and logging hooks are normal priority
        if any(keyword in filename for keyword in ["cleanup", "log", "tracker"]):
            return HookPriority.NORMAL

        # Analytics and metrics are low priority
        if any(keyword in filename for keyword in ["analytics", "metrics", "stats"]):
            return HookPriority.LOW

        # Default priority based on event type
        if event_type == HookEvent.PRE_TOOL_USE:
            return HookPriority.HIGH  # Pre-execution validation is important
        elif event_type == HookEvent.SESSION_START:
            return HookPriority.NORMAL
        else:
            return HookPriority.NORMAL

    def _estimate_execution_time(self, hook_path: str) -> float:
        """Estimate hook execution time based on historical data and characteristics"""
        # Check cache for historical execution time
        cache_key = f"exec_time:{hook_path}"
        if cache_key in self._metadata_cache:
            cached_time = self._metadata_cache[cache_key].get("avg_time_ms")
            if cached_time:
                return cached_time

        # Estimate based on hook characteristics
        filename = hook_path.lower()

        # Hooks with git operations tend to be slower
        if "git" in filename:
            return 200.0  # 200ms estimate for git operations

        # Hooks with network operations are slower
        if any(keyword in filename for keyword in ["fetch", "api", "network"]):
            return 500.0  # 500ms estimate for network operations

        # Hooks with file I/O are moderate
        if any(keyword in filename for keyword in ["read", "write", "parse"]):
            return 50.0  # 50ms estimate for file I/O

        # Simple hooks are fast
        return 10.0  # 10ms estimate for simple operations

    def _determine_phase_relevance(self, hook_path: str, event_type: HookEvent) -> Dict[Phase, float]:
        """Determine hook relevance to different development phases"""
        filename = hook_path.lower()
        relevance = {}

        # Default relevance for all phases
        default_relevance = 0.5

        # SPEC phase relevance
        if any(keyword in filename for keyword in ["spec", "plan", "design", "requirement"]):
            relevance[Phase.SPEC] = 1.0
        else:
            relevance[Phase.SPEC] = default_relevance

        # RED phase relevance (testing)
        if any(keyword in filename for keyword in ["test", "red", "tdd", "assert"]):
            relevance[Phase.RED] = 1.0
        else:
            relevance[Phase.RED] = default_relevance

        # GREEN phase relevance (implementation)
        if any(keyword in filename for keyword in ["implement", "code", "green", "build"]):
            relevance[Phase.GREEN] = 1.0
        else:
            relevance[Phase.GREEN] = default_relevance

        # REFACTOR phase relevance
        if any(keyword in filename for keyword in ["refactor", "optimize", "improve", "clean"]):
            relevance[Phase.REFACTOR] = 1.0
        else:
            relevance[Phase.REFACTOR] = default_relevance

        # SYNC phase relevance (documentation)
        if any(keyword in filename for keyword in ["sync", "doc", "document", "deploy"]):
            relevance[Phase.SYNC] = 1.0
        else:
            relevance[Phase.SYNC] = default_relevance

        # DEBUG phase relevance
        if any(keyword in filename for keyword in ["debug", "error", "troubleshoot", "log"]):
            relevance[Phase.DEBUG] = 1.0
        else:
            relevance[Phase.DEBUG] = default_relevance

        # PLANNING phase relevance
        if any(keyword in filename for keyword in ["plan", "analysis", "strategy"]):
            relevance[Phase.PLANNING] = 1.0
        else:
            relevance[Phase.PLANNING] = default_relevance

        return relevance

    def _estimate_token_cost(self, hook_path: str) -> int:
        """Estimate token cost for hook execution"""
        # Base token cost for any hook
        base_cost = 100

        # Additional cost based on hook characteristics
        filename = hook_path.lower()

        if any(keyword in filename for keyword in ["analysis", "report", "generate"]):
            base_cost += 500  # Higher cost for analysis/generation
        elif any(keyword in filename for keyword in ["log", "simple", "basic"]):
            base_cost += 50  # Lower cost for simple operations

        return base_cost

    def _is_parallel_safe(self, hook_path: str) -> bool:
        """Determine if hook can be executed in parallel"""
        filename = hook_path.lower()

        # Hooks that modify shared state are not parallel safe
        if any(keyword in filename for keyword in ["write", "modify", "update", "delete"]):
            return False

        # Hooks with external dependencies might not be parallel safe
        if any(keyword in filename for keyword in ["database", "network", "api"]):
            return False

        # Most hooks are parallel safe by default
        return True

    async def execute_hooks(
        self,
        event_type: HookEvent,
        context: Dict[str, Any],
        user_input: Optional[str] = None,
        phase: Optional[Phase] = None,
        max_total_execution_time_ms: float = 15000.0,
    ) -> List[HookExecutionResult]:
        """Execute hooks for a specific event with JIT optimization

        Args:
            event_type: Type of hook event
            context: Execution context data
            user_input: User input for phase detection
            phase: Current development phase (if known)
            max_total_execution_time_ms: Maximum total execution time for all hooks

        Returns:
            List of hook execution results
        """
        start_time = time.time()

        # Detect phase if not provided
        if phase is None and user_input:
            try:
                phase = self.jit_loader.phase_detector.detect_phase(user_input)
            except AttributeError:
                # Fallback if JIT loader doesn't have phase detector
                phase = Phase.SPEC

        # Get relevant hooks for this event
        hook_paths = self._hooks_by_event.get(event_type, [])

        # Filter and prioritize hooks based on phase and performance
        prioritized_hooks = self._prioritize_hooks(hook_paths, phase)

        # Load optimized context using JIT system
        optimized_context = await self._load_optimized_context(event_type, context, phase, prioritized_hooks)

        # Execute hooks with optimization
        results = await self._execute_hooks_optimized(prioritized_hooks, optimized_context, max_total_execution_time_ms)

        # Update performance metrics
        if self.enable_performance_monitoring:
            self._update_performance_metrics(event_type, phase, results, start_time)

        return results

    def _prioritize_hooks(self, hook_paths: List[str], phase: Optional[Phase]) -> List[Tuple[str, float]]:
        """Prioritize hooks based on phase relevance and performance characteristics

        Args:
            hook_paths: List of hook file paths
            phase: Current development phase

        Returns:
            List of (hook_path, priority_score) tuples sorted by priority
        """
        hook_priorities = []

        for hook_path in hook_paths:
            metadata = self._hook_registry.get(hook_path)
            if not metadata:
                continue

            # Calculate priority score
            priority_score = 0.0

            # Base priority (lower number = higher priority)
            priority_score += metadata.priority.value * 10

            # Phase relevance bonus
            if phase and phase in metadata.phase_relevance:
                relevance = metadata.phase_relevance[phase]
                priority_score -= relevance * 5  # Higher relevance = lower score (higher priority)

            # Performance penalty (slower hooks get lower priority)
            priority_score += metadata.estimated_execution_time_ms / 100

            # Success rate bonus (more reliable hooks get higher priority)
            if metadata.success_rate < 0.9:
                priority_score += 5  # Penalize unreliable hooks

            hook_priorities.append((hook_path, priority_score))

        # Sort by priority score (lower is better)
        hook_priorities.sort(key=lambda x: x[1])

        return hook_priorities

    async def _load_optimized_context(
        self,
        event_type: HookEvent,
        context: Dict[str, Any],
        phase: Optional[Phase],
        prioritized_hooks: List[Tuple[str, float]],
    ) -> Dict[str, Any]:
        """Load optimized context using JIT system for hook execution

        Args:
            event_type: Hook event type
            context: Original context
            phase: Current development phase
            prioritized_hooks: List of prioritized hooks

        Returns:
            Optimized context with relevant information
        """
        # Create synthetic user input for context loading
        synthetic_input = f"Hook execution for {event_type.value}"
        if phase:
            synthetic_input += f" during {phase.value} phase"

        # Load context using JIT system
        try:
            jit_context, context_metrics = await self.jit_loader.load_context(
                user_input=synthetic_input, context=context
            )
        except (TypeError, AttributeError):
            # Fallback to basic context if JIT loader interface is different
            jit_context = context.copy()

        # Add hook-specific context
        optimized_context = jit_context.copy()
        optimized_context.update(
            {
                "hook_event_type": event_type.value,
                "hook_phase": phase.value if phase else None,
                "hook_execution_mode": "optimized",
                "prioritized_hooks": [hook_path for hook_path, _ in prioritized_hooks[:5]],  # Top 5 hooks
            }
        )

        return optimized_context

    async def _execute_hooks_optimized(
        self,
        prioritized_hooks: List[Tuple[str, float]],
        context: Dict[str, Any],
        max_total_execution_time_ms: float,
    ) -> List[HookExecutionResult]:
        """Execute hooks with optimization and time management

        Args:
            prioritized_hooks: List of (hook_path, priority_score) tuples
            context: Optimized execution context
            max_total_execution_time_ms: Maximum total execution time

        Returns:
            List of hook execution results
        """
        results = []
        remaining_time = max_total_execution_time_ms

        # Separate hooks into parallel-safe and sequential
        parallel_hooks = []
        sequential_hooks = []

        for hook_path, _ in prioritized_hooks:
            metadata = self._hook_registry.get(hook_path)
            if metadata and metadata.parallel_safe:
                parallel_hooks.append(hook_path)
            else:
                sequential_hooks.append(hook_path)

        # Execute parallel hooks first (faster)
        if parallel_hooks and remaining_time > 0:
            parallel_results = await self._execute_hooks_parallel(parallel_hooks, context, remaining_time)
            results.extend(parallel_results)

            # Update remaining time
            total_parallel_time = sum(r.execution_time_ms for r in parallel_results)
            remaining_time -= total_parallel_time

        # Execute sequential hooks with remaining time
        if sequential_hooks and remaining_time > 0:
            sequential_results = await self._execute_hooks_sequential(sequential_hooks, context, remaining_time)
            results.extend(sequential_results)

        return results

    async def _execute_hooks_parallel(
        self, hook_paths: List[str], context: Dict[str, Any], max_total_time_ms: float
    ) -> List[HookExecutionResult]:
        """Execute hooks in parallel with time management"""
        results = []

        # Create semaphore to limit concurrent executions
        semaphore = asyncio.Semaphore(self.max_concurrent_hooks)

        async def execute_single_hook(hook_path: str) -> Optional[HookExecutionResult]:
            async with semaphore:
                try:
                    return await self._execute_single_hook(hook_path, context)
                except Exception as e:
                    return HookExecutionResult(
                        hook_path=hook_path,
                        success=False,
                        execution_time_ms=0.0,
                        token_usage=0,
                        output=None,
                        error_message=str(e),
                    )

        # Execute hooks with timeout
        tasks = [execute_single_hook(hook_path) for hook_path in hook_paths]

        try:
            # Wait for all hooks with total timeout
            completed_results = await asyncio.wait_for(
                asyncio.gather(*tasks, return_exceptions=True),
                timeout=max_total_time_ms / 1000.0,
            )

            for result in completed_results:
                if isinstance(result, HookExecutionResult):
                    results.append(result)
                elif isinstance(result, Exception):
                    # Handle exceptions
                    error_result = HookExecutionResult(
                        hook_path="unknown",
                        success=False,
                        execution_time_ms=0.0,
                        token_usage=0,
                        output=None,
                        error_message=str(result),
                    )
                    results.append(error_result)

        except asyncio.TimeoutError:
            # Some hooks didn't complete in time
            pass

        return results

    async def _execute_hooks_sequential(
        self, hook_paths: List[str], context: Dict[str, Any], max_total_time_ms: float
    ) -> List[HookExecutionResult]:
        """Execute hooks sequentially with time management"""
        results = []
        remaining_time = max_total_time_ms

        for hook_path in hook_paths:
            if remaining_time <= 0:
                break

            try:
                result = await self._execute_single_hook(hook_path, context)
                results.append(result)

                # Update remaining time
                execution_time = result.execution_time_ms
                remaining_time -= execution_time

            except Exception as e:
                error_result = HookExecutionResult(
                    hook_path=hook_path,
                    success=False,
                    execution_time_ms=0.0,
                    token_usage=0,
                    output=None,
                    error_message=str(e),
                )
                results.append(error_result)

        return results

    async def _execute_single_hook(self, hook_path: str, context: Dict[str, Any]) -> HookExecutionResult:
        """Execute a single hook with Phase 2 optimizations

        Args:
            hook_path: Path to hook file
            context: Execution context

        Returns:
            Hook execution result with enhanced monitoring and reliability
        """
        start_time = time.time()
        full_hook_path = self.hooks_directory / hook_path

        try:
            # Get metadata for the hook
            metadata = self._hook_registry.get(hook_path)
            if not metadata:
                raise ValueError(f"Hook metadata not found for {hook_path}")

            # Initialize circuit breaker and retry policy for this hook if needed
            if hook_path not in self._circuit_breakers:
                self._circuit_breakers[hook_path] = CircuitBreaker(
                    failure_threshold=self.circuit_breaker_threshold,
                    timeout_seconds=60,
                    success_threshold=5,
                )
                self._retry_policies[hook_path] = RetryPolicy(
                    max_retries=self.max_retries, base_delay_ms=100, max_delay_ms=5000
                )

            circuit_breaker = self._circuit_breakers[hook_path]
            retry_policy = self._retry_policies[hook_path]

            # Check advanced cache first
            cache_key = f"hook_result:{hook_path}:{hash(str(context))}"
            cached_result = self._advanced_cache.get(cache_key)
            if cached_result:
                if cached_result.success:
                    with self._performance_lock:
                        self.metrics.cache_hits += 1
                    return cached_result

            # Execute with circuit breaker protection and retry logic
            async def execute_hook_with_retry():
                return await self._execute_hook_subprocess(full_hook_path, context, metadata)

            # Apply circuit breaker and retry pattern
            try:
                result = await circuit_breaker.call(retry_policy.execute_with_retry, execute_hook_with_retry)
            except Exception as e:
                # Circuit breaker is OPEN or all retries exhausted
                execution_time = (time.time() - start_time) * 1000
                with self._performance_lock:
                    if circuit_breaker.state.state == "OPEN":
                        self.metrics.circuit_breaker_trips += 1

                self._logger.warning(f"Hook {hook_path} failed due to circuit breaker: {str(e)}")

                return HookExecutionResult(
                    hook_path=hook_path,
                    success=False,
                    execution_time_ms=execution_time,
                    token_usage=0,
                    output=None,
                    error_message=f"Circuit breaker OPEN: {str(e)}",
                    metadata={"circuit_breaker_state": circuit_breaker.state.state},
                )

            # Update resource usage metrics
            current_resources = self._resource_monitor.get_current_metrics()
            with self._performance_lock:
                self.metrics.resource_usage = current_resources

            # Performance anomaly detection
            anomaly = self._anomaly_detector.detect_anomaly(hook_path, result.execution_time_ms)
            if anomaly:
                self._logger.warning(f"Performance anomaly detected for {hook_path}: {anomaly}")
                result.metadata["performance_anomaly"] = anomaly

            # Cache successful results with TTL based on hook characteristics
            if result.success:
                cache_ttl = self._determine_cache_ttl(hook_path, metadata)
                self._advanced_cache.put(cache_key, result, ttl_seconds=cache_ttl)

            # Update cache statistics
            with self._performance_lock:
                self.metrics.cache_misses += 1

            # Update execution profile for performance monitoring
            self._execution_profiles[hook_path].append(result.execution_time_ms)
            if len(self._execution_profiles[hook_path]) > 100:
                self._execution_profiles[hook_path].pop(0)

            # Update metadata
            self._update_hook_metadata(hook_path, result)

            return result

        except Exception as e:
            execution_time = (time.time() - start_time) * 1000
            self._logger.error(f"Unexpected error executing hook {hook_path}: {str(e)}")

            return HookExecutionResult(
                hook_path=hook_path,
                success=False,
                execution_time_ms=execution_time,
                token_usage=0,
                output=None,
                error_message=f"Unexpected error: {str(e)}",
            )

    def _determine_cache_ttl(self, hook_path: str, metadata: HookMetadata) -> int:
        """Determine optimal cache TTL based on hook characteristics"""
        filename = hook_path.lower()

        # Hooks that fetch external data should have shorter TTL
        if any(keyword in filename for keyword in ["fetch", "api", "network", "git"]):
            return 60  # 1 minute

        # Hooks that read static files can have longer TTL
        if any(keyword in filename for keyword in ["read", "parse", "analyze"]):
            return 1800  # 30 minutes

        # Hooks that write or modify data should have very short TTL
        if any(keyword in filename for keyword in ["write", "modify", "update", "create"]):
            return 30  # 30 seconds

        # Default TTL
        return 300  # 5 minutes

    async def _execute_hook_subprocess(
        self, hook_path: Path, context: Dict[str, Any], metadata: HookMetadata
    ) -> HookExecutionResult:
        """Execute hook in isolated subprocess

        Args:
            hook_path: Full path to hook file
            context: Execution context
            metadata: Hook metadata

        Returns:
            Hook execution result
        """
        start_time = time.time()

        try:
            # Prepare input for hook
            hook_input = json.dumps(context)

            # Execute hook with timeout
            timeout_seconds = max(1.0, metadata.estimated_execution_time_ms / 1000.0)

            process = await asyncio.create_subprocess_exec(
                "uv",
                "run",
                str(hook_path),
                stdin=asyncio.subprocess.PIPE,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
                cwd=Path.cwd(),
            )

            try:
                stdout, stderr = await asyncio.wait_for(
                    process.communicate(input=hook_input.encode()),
                    timeout=timeout_seconds,
                )
            except asyncio.TimeoutError:
                process.kill()
                await process.wait()
                raise TimeoutError(f"Hook execution timed out after {timeout_seconds}s")

            execution_time_ms = (time.time() - start_time) * 1000
            success = process.returncode == 0

            # Parse output
            output = None
            if stdout:
                try:
                    output = json.loads(stdout.decode())
                except json.JSONDecodeError:
                    output = stdout.decode()

            error_message = None
            if stderr:
                error_message = stderr.decode().strip()
            elif process.returncode != 0:
                error_message = f"Hook exited with code {process.returncode}"

            return HookExecutionResult(
                hook_path=str(hook_path.relative_to(self.hooks_directory)),
                success=success,
                execution_time_ms=execution_time_ms,
                token_usage=metadata.token_cost_estimate,
                output=output,
                error_message=error_message,
            )

        except Exception as e:
            execution_time_ms = (time.time() - start_time) * 1000

            return HookExecutionResult(
                hook_path=str(hook_path.relative_to(self.hooks_directory)),
                success=False,
                execution_time_ms=execution_time_ms,
                token_usage=metadata.token_cost_estimate,
                output=None,
                error_message=str(e),
            )

    def _update_hook_metadata(self, hook_path: str, result: HookExecutionResult) -> None:
        """Update hook metadata based on execution result"""
        metadata = self._hook_registry.get(hook_path)
        if not metadata:
            return

        # Update execution time estimate
        cache_key = f"exec_time:{hook_path}"
        if cache_key not in self._metadata_cache:
            self._metadata_cache[cache_key] = {"count": 0, "total_time": 0.0}

        cache_entry = self._metadata_cache[cache_key]
        cache_entry["count"] += 1
        cache_entry["total_time"] += result.execution_time_ms
        cache_entry["avg_time_ms"] = cache_entry["total_time"] / cache_entry["count"]

        # Update success rate
        metadata.success_rate = (metadata.success_rate * 0.8) + (1.0 if result.success else 0.0) * 0.2
        metadata.last_execution_time = datetime.now()

    def _update_performance_metrics(
        self,
        event_type: HookEvent,
        phase: Optional[Phase],
        results: List[HookExecutionResult],
        start_time: float,
    ) -> None:
        """Update performance metrics"""
        with self._performance_lock:
            self.metrics.total_executions += len(results)
            self.metrics.successful_executions += sum(1 for r in results if r.success)

            total_execution_time = sum(r.execution_time_ms for r in results)
            self.metrics.average_execution_time_ms = (self.metrics.average_execution_time_ms * 0.9) + (
                total_execution_time / len(results) * 0.1
            )

            self.metrics.total_token_usage += sum(r.token_usage for r in results)

            if phase:
                self.metrics.phase_distribution[phase] = self.metrics.phase_distribution.get(phase, 0) + 1

            self.metrics.event_type_distribution[event_type] = (
                self.metrics.event_type_distribution.get(event_type, 0) + 1
            )

            # Log performance data
            self._log_performance_data(event_type, phase, results, start_time)

    def _log_performance_data(
        self,
        event_type: HookEvent,
        phase: Optional[Phase],
        results: List[HookExecutionResult],
        start_time: float,
    ) -> None:
        """Log performance data to file"""
        log_entry = {
            "timestamp": datetime.now().isoformat(),
            "event_type": event_type.value,
            "phase": phase.value if phase else None,
            "total_hooks": len(results),
            "successful_hooks": sum(1 for r in results if r.success),
            "total_execution_time_ms": sum(r.execution_time_ms for r in results),
            "total_token_usage": sum(r.token_usage for r in results),
            "system_time_ms": (time.time() - start_time) * 1000,
            "results": [
                {
                    "hook_path": r.hook_path,
                    "success": r.success,
                    "execution_time_ms": r.execution_time_ms,
                    "token_usage": r.token_usage,
                    "error_message": r.error_message,
                }
                for r in results
            ],
        }

        try:
            with open(self._performance_log_path, "a") as f:
                f.write(json.dumps(log_entry) + "\n")
        except Exception:
            pass  # Silently fail on logging

    def get_performance_metrics(self) -> HookPerformanceMetrics:
        """Get comprehensive performance metrics with Phase 2 enhancements"""
        with self._performance_lock:
            # Get advanced cache stats
            self._advanced_cache.get_stats()

            # Calculate performance profiles summary
            self._calculate_performance_summary()

            # Get peak resource usage
            self._resource_monitor.get_peak_metrics()

            return HookPerformanceMetrics(
                total_executions=self.metrics.total_executions,
                successful_executions=self.metrics.successful_executions,
                average_execution_time_ms=self.metrics.average_execution_time_ms,
                total_token_usage=self.metrics.total_token_usage,
                cache_hits=self.metrics.cache_hits,
                cache_misses=self.metrics.cache_misses,
                phase_distribution=self.metrics.phase_distribution.copy(),
                event_type_distribution=self.metrics.event_type_distribution.copy(),
                circuit_breaker_trips=self.metrics.circuit_breaker_trips,
                retry_attempts=self.metrics.retry_attempts,
                resource_usage=ResourceUsageMetrics(
                    cpu_usage_percent=self.metrics.resource_usage.cpu_usage_percent,
                    memory_usage_mb=self.metrics.resource_usage.memory_usage_mb,
                    disk_io_mb=self.metrics.resource_usage.disk_io_mb,
                    network_io_mb=self.metrics.resource_usage.network_io_mb,
                    open_files=self.metrics.resource_usage.open_files,
                    thread_count=self.metrics.resource_usage.thread_count,
                ),
            )

    def _calculate_performance_summary(self) -> Dict[str, Any]:
        """Calculate detailed performance summary"""
        hook_performance: Dict[str, Dict[str, Any]] = {}
        summary: Dict[str, Any] = {
            "hook_performance": hook_performance,
            "cache_efficiency": 0.0,
            "overall_health": "healthy",
        }

        # Calculate cache efficiency
        total_cache_requests = self.metrics.cache_hits + self.metrics.cache_misses
        if total_cache_requests > 0:
            summary["cache_efficiency"] = self.metrics.cache_hits / total_cache_requests

        # Calculate per-hook performance statistics
        for hook_path, execution_times in self._execution_profiles.items():
            if execution_times:
                hook_performance[hook_path] = {
                    "avg_time_ms": sum(execution_times) / len(execution_times),
                    "min_time_ms": min(execution_times),
                    "max_time_ms": max(execution_times),
                    "execution_count": len(execution_times),
                    "std_dev_ms": self._calculate_std_dev(execution_times),
                }

        # Determine overall health
        success_rate = self.metrics.successful_executions / max(self.metrics.total_executions, 1)
        if success_rate < 0.9:
            summary["overall_health"] = "degraded"
        elif success_rate < 0.7:
            summary["overall_health"] = "unhealthy"

        return summary

    def _calculate_std_dev(self, values: List[float]) -> float:
        """Calculate standard deviation"""
        if len(values) < 2:
            return 0.0
        mean = sum(values) / len(values)
        variance = sum((x - mean) ** 2 for x in values) / len(values)
        return variance**0.5

    async def get_system_health_report(self) -> Dict[str, Any]:
        """Get comprehensive health report"""
        return await self._health_checker.check_system_health()

    def get_connection_pool_stats(self) -> Dict[str, Any]:
        """Get connection pool statistics"""
        return self._connection_pool.get_pool_stats()

    def get_advanced_cache_stats(self) -> Dict[str, Any]:
        """Get advanced cache statistics"""
        return self._advanced_cache.get_stats()

    def get_circuit_breaker_status(self) -> Dict[str, Any]:
        """Get circuit breaker status for all hooks"""
        return {
            hook_path: {
                "state": cb.state.state,
                "failure_count": cb.state.failure_count,
                "last_failure_time": (cb.state.last_failure_time.isoformat() if cb.state.last_failure_time else None),
                "success_threshold": cb.state.success_threshold,
            }
            for hook_path, cb in self._circuit_breakers.items()
        }

    def get_hook_recommendations(
        self, event_type: Optional[HookEvent] = None, phase: Optional[Phase] = None
    ) -> Dict[str, Any]:
        """Get recommendations for hook optimization

        Args:
            event_type: Specific event type to analyze
            phase: Specific phase to analyze

        Returns:
            Dictionary with optimization recommendations
        """
        recommendations: Dict[str, List[Any]] = {
            "slow_hooks": [],
            "unreliable_hooks": [],
            "phase_mismatched_hooks": [],
            "optimization_suggestions": [],
        }

        # Analyze hook performance
        for hook_path, metadata in self._hook_registry.items():
            if event_type and metadata.event_type != event_type:
                continue

            # Check for slow hooks
            if metadata.estimated_execution_time_ms > 200:
                recommendations["slow_hooks"].append(
                    {
                        "hook_path": hook_path,
                        "estimated_time_ms": metadata.estimated_execution_time_ms,
                        "suggestion": "Consider optimizing or making this hook parallel-safe",
                    }
                )

            # Check for unreliable hooks
            if metadata.success_rate < 0.8:
                recommendations["unreliable_hooks"].append(
                    {
                        "hook_path": hook_path,
                        "success_rate": metadata.success_rate,
                        "suggestion": "Review error handling and improve reliability",
                    }
                )

            # Check for phase mismatch
            if phase:
                relevance = metadata.phase_relevance.get(phase, 0.0)
                if relevance < 0.3:
                    recommendations["phase_mismatched_hooks"].append(
                        {
                            "hook_path": hook_path,
                            "phase": phase.value,
                            "relevance": relevance,
                            "suggestion": "This hook may not be relevant for the current phase",
                        }
                    )

        # Generate optimization suggestions
        if recommendations["slow_hooks"]:
            recommendations["optimization_suggestions"].append(
                "Consider implementing caching for frequently executed slow hooks"
            )

        if recommendations["unreliable_hooks"]:
            recommendations["optimization_suggestions"].append(
                "Add retry logic and better error handling for unreliable hooks"
            )

        if recommendations["phase_mismatched_hooks"]:
            recommendations["optimization_suggestions"].append(
                "Use phase-based hook filtering to skip irrelevant hooks"
            )

        return recommendations

    async def cleanup(self) -> None:
        """Enhanced cleanup with Phase 2 resource management"""
        try:
            # Save comprehensive performance metrics and state
            metrics_file = self.cache_directory / "metrics.json"
            state_file = self.cache_directory / "state.json"

            # Get current metrics
            current_metrics = self.get_performance_metrics()
            health_report = await self.get_system_health_report()

            metrics_data = {
                "timestamp": datetime.now().isoformat(),
                "phase": "phase_2_optimized",
                "metrics": {
                    "total_executions": current_metrics.total_executions,
                    "successful_executions": current_metrics.successful_executions,
                    "average_execution_time_ms": current_metrics.average_execution_time_ms,
                    "total_token_usage": current_metrics.total_token_usage,
                    "cache_hits": current_metrics.cache_hits,
                    "cache_misses": current_metrics.cache_misses,
                    "circuit_breaker_trips": current_metrics.circuit_breaker_trips,
                    "retry_attempts": current_metrics.retry_attempts,
                    "resource_usage": current_metrics.resource_usage.__dict__,
                    "phase_distribution": {k.value: v for k, v in current_metrics.phase_distribution.items()},
                    "event_type_distribution": {k.value: v for k, v in current_metrics.event_type_distribution.items()},
                },
                "health_status": health_report,
                "cache_stats": self.get_advanced_cache_stats(),
                "connection_pool_stats": self.get_connection_pool_stats(),
                "circuit_breaker_status": self.get_circuit_breaker_status(),
                "hook_metadata": {
                    hook_path: {
                        "estimated_execution_time_ms": metadata.estimated_execution_time_ms,
                        "success_rate": metadata.success_rate,
                        "last_execution_time": (
                            metadata.last_execution_time.isoformat() if metadata.last_execution_time else None
                        ),
                        "priority": metadata.priority.value,
                        "parallel_safe": metadata.parallel_safe,
                        "token_cost_estimate": metadata.token_cost_estimate,
                    }
                    for hook_path, metadata in self._hook_registry.items()
                },
                "performance_profiles": {
                    hook_path: {
                        "execution_times": times[-10:],  # Keep last 10 execution times
                        "avg_time_ms": sum(times) / len(times) if times else 0,
                        "count": len(times),
                    }
                    for hook_path, times in self._execution_profiles.items()
                    if times
                },
            }

            # Save metrics
            with open(metrics_file, "w") as f:
                json.dump(metrics_data, f, indent=2)

            # Save state for recovery
            state_data = {
                "timestamp": datetime.now().isoformat(),
                "circuit_breaker_states": self.get_circuit_breaker_status(),
                "cache_config": {
                    "max_size": self._advanced_cache.max_size,
                    "default_ttl_seconds": self._advanced_cache.default_ttl_seconds,
                },
                "connection_pool_config": {
                    "max_connections": self._connection_pool.max_connections,
                    "connection_timeout_seconds": self._connection_pool.connection_timeout_seconds,
                },
                "optimization_config": {
                    "circuit_breaker_threshold": self.circuit_breaker_threshold,
                    "max_retries": self.max_retries,
                },
            }

            with open(state_file, "w") as f:
                json.dump(state_data, f, indent=2)

        except Exception as e:
            if hasattr(self, "_logger"):
                self._logger.error(f"Error during cleanup: {str(e)}")

        finally:
            # Clear all caches and resources
            try:
                self._advanced_cache.invalidate()  # Clear all cache entries
                self._metadata_cache.clear()

                # Reset circuit breakers
                for circuit_breaker in self._circuit_breakers.values():
                    circuit_breaker.state.state = "CLOSED"
                    circuit_breaker.state.failure_count = 0
                    circuit_breaker.state.last_failure_time = None

                # Clear execution profiles
                self._execution_profiles.clear()

            except Exception as e:
                if hasattr(self, "_logger"):
                    self._logger.error(f"Error during cache cleanup: {str(e)}")

            # Log peak resource usage for monitoring
            try:
                peak_resources = self._resource_monitor.get_peak_metrics()
                if hasattr(self, "_logger"):
                    self._logger.info(
                        f"Peak resource usage - Memory: {peak_resources.memory_usage_mb:.1f}MB, "
                        f"CPU: {peak_resources.cpu_usage_percent:.1f}%, "
                        f"Threads: {peak_resources.thread_count}"
                    )
            except Exception:
                pass  # Ignore cleanup logging errors


# Global instance for easy access
_jit_hook_manager: Optional[JITEnhancedHookManager] = None


def get_jit_hook_manager() -> JITEnhancedHookManager:
    """Get or create global JIT hook manager instance"""
    global _jit_hook_manager
    if _jit_hook_manager is None:
        _jit_hook_manager = JITEnhancedHookManager()
    return _jit_hook_manager


# Convenience functions for common hook operations
async def execute_session_start_hooks(
    context: Dict[str, Any], user_input: Optional[str] = None
) -> List[HookExecutionResult]:
    """Execute SessionStart hooks with JIT optimization"""
    manager = get_jit_hook_manager()
    return await manager.execute_hooks(HookEvent.SESSION_START, context, user_input=user_input)


async def execute_pre_tool_hooks(
    context: Dict[str, Any], user_input: Optional[str] = None
) -> List[HookExecutionResult]:
    """Execute PreToolUse hooks with JIT optimization"""
    manager = get_jit_hook_manager()
    return await manager.execute_hooks(HookEvent.PRE_TOOL_USE, context, user_input=user_input)


async def execute_session_end_hooks(
    context: Dict[str, Any], user_input: Optional[str] = None
) -> List[HookExecutionResult]:
    """Execute SessionEnd hooks with JIT optimization"""
    manager = get_jit_hook_manager()
    return await manager.execute_hooks(HookEvent.SESSION_END, context, user_input=user_input)


def get_hook_performance_metrics() -> HookPerformanceMetrics:
    """Get current hook performance metrics"""
    manager = get_jit_hook_manager()
    return manager.get_performance_metrics()


def get_hook_optimization_recommendations(
    event_type: Optional[HookEvent] = None, phase: Optional[Phase] = None
) -> Dict[str, Any]:
    """Get hook optimization recommendations"""
    manager = get_jit_hook_manager()
    return manager.get_hook_recommendations(event_type, phase)


# Phase 2 convenience functions for enhanced monitoring and control


async def get_system_health() -> Dict[str, Any]:
    """Get comprehensive system health report"""
    manager = get_jit_hook_manager()
    return await manager.get_system_health_report()


def get_connection_pool_info() -> Dict[str, Any]:
    """Get connection pool statistics and status"""
    manager = get_jit_hook_manager()
    return manager.get_connection_pool_stats()


def get_cache_performance() -> Dict[str, Any]:
    """Get advanced cache performance metrics"""
    manager = get_jit_hook_manager()
    return manager.get_advanced_cache_stats()


def get_circuit_breaker_info() -> Dict[str, Any]:
    """Get circuit breaker status for all hooks"""
    manager = get_jit_hook_manager()
    return manager.get_circuit_breaker_status()


def invalidate_hook_cache(pattern: Optional[str] = None) -> None:
    """Invalidate hook cache entries"""
    manager = get_jit_hook_manager()
    manager._advanced_cache.invalidate(pattern)


def reset_circuit_breakers(hook_path: Optional[str] = None) -> None:
    """Reset circuit breakers (specific hook or all)"""
    manager = get_jit_hook_manager()

    if hook_path:
        if hook_path in manager._circuit_breakers:
            cb = manager._circuit_breakers[hook_path]
            cb.state.state = "CLOSED"
            cb.state.failure_count = 0
            cb.state.last_failure_time = None
    else:
        for cb in manager._circuit_breakers.values():
            cb.state.state = "CLOSED"
            cb.state.failure_count = 0
            cb.state.last_failure_time = None


async def optimize_hook_system() -> Dict[str, Any]:
    """Run system optimization and return recommendations"""
    manager = get_jit_hook_manager()

    # Get current health and metrics
    health_report = await manager.get_system_health_report()
    metrics = manager.get_performance_metrics()
    manager.get_advanced_cache_stats()

    # Generate optimization recommendations
    optimization_report = {
        "timestamp": datetime.now().isoformat(),
        "health_status": health_report["status"],
        "performance_summary": {
            "success_rate": metrics.successful_executions / max(metrics.total_executions, 1),
            "average_execution_time_ms": metrics.average_execution_time_ms,
            "cache_efficiency": metrics.cache_hits / max(metrics.cache_hits + metrics.cache_misses, 1),
            "circuit_breaker_trips": metrics.circuit_breaker_trips,
        },
        "recommendations": [],
    }

    # Performance recommendations
    if metrics.average_execution_time_ms > 500:
        optimization_report["recommendations"].append("Consider optimizing slow hooks or increasing parallel execution")

    if metrics.cache_hits / max(metrics.cache_hits + metrics.cache_misses, 1) < 0.3:
        optimization_report["recommendations"].append(
            "Cache hit rate is low - consider increasing TTL or reviewing cache strategy"
        )

    if metrics.circuit_breaker_trips > 5:
        optimization_report["recommendations"].append(
            "High circuit breaker activity detected - review hook reliability"
        )

    # Resource recommendations
    if metrics.resource_usage.memory_usage_mb > 500:
        optimization_report["recommendations"].append(
            "High memory usage - consider reducing cache size or optimizing resource usage"
        )

    if health_report["status"] != "healthy":
        optimization_report["recommendations"].append(
            f"System health is {health_report['status']} - review health checks"
        )

    return optimization_report


if __name__ == "__main__":
    # Example usage and testing with Phase 2 optimizations
    async def test_phase2_optimizations():
        """Test Phase 2 performance optimizations and reliability features"""
        print("🚀 Testing Phase 2 JIT-Enhanced Hook Manager Optimizations")
        print("=" * 60)

        # Initialize with Phase 2 optimizations
        manager = JITEnhancedHookManager(
            cache_ttl_seconds=300,
            circuit_breaker_threshold=3,
            max_retries=2,
            connection_pool_size=5,
        )

        try:
            # Test hook execution with advanced features
            context = {"test": True, "user": "test_user", "session_id": "test_session"}

            print("\n📊 Testing Hook Execution with Phase 2 Optimizations:")
            results = await manager.execute_hooks(
                HookEvent.SESSION_START,
                context,
                user_input="Testing Phase 2 JIT enhanced hook system",
            )

            print(f"Executed {len(results)} hooks")
            for result in results:
                status = "✓" if result.success else "✗"
                print(f"  {result.hook_path}: {status} ({result.execution_time_ms:.1f}ms)")
                if result.metadata.get("performance_anomaly"):
                    anomaly = result.metadata["performance_anomaly"]
                    print(f"    ⚠️  Performance anomaly: {anomaly['anomaly_type']} ({anomaly['severity']})")

            # Show Phase 2 enhanced metrics
            print("\n📈 Phase 2 Performance Metrics:")
            metrics = manager.get_performance_metrics()
            print(f"  Total executions: {metrics.total_executions}")
            print(f"  Success rate: {metrics.successful_executions}/{metrics.total_executions}")
            print(f"  Avg execution time: {metrics.average_execution_time_ms:.1f}ms")
            print(f"  Cache hits: {metrics.cache_hits}, misses: {metrics.cache_misses}")
            print(f"  Circuit breaker trips: {metrics.circuit_breaker_trips}")
            print(f"  Retry attempts: {metrics.retry_attempts}")
            print(f"  Memory usage: {metrics.resource_usage.memory_usage_mb:.1f}MB")
            print(f"  CPU usage: {metrics.resource_usage.cpu_usage_percent:.1f}%")

            # Test health monitoring
            print("\n🏥 System Health Check:")
            health_report = await manager.get_system_health_report()
            print(f"  Overall status: {health_report['status']}")
            for check_name, check_data in health_report["checks"].items():
                status_icon = "✓" if check_data["status"] == "healthy" else "⚠️"
                print(f"  {check_name}: {status_icon} {check_data['status']}")

            # Test cache performance
            print("\n💾 Cache Performance:")
            cache_stats = manager.get_advanced_cache_stats()
            print(f"  Cache size: {cache_stats['size']}/{cache_stats['max_size']}")
            print(f"  Utilization: {cache_stats['utilization']:.1%}")

            # Test circuit breaker status
            print("\n⚡ Circuit Breaker Status:")
            cb_status = manager.get_circuit_breaker_status()
            for hook_name, cb_data in cb_status.items():
                print(f"  {hook_name}: {cb_data['state']} ({cb_data['failure_count']} failures)")

            # Test system optimization
            print("\n🔧 System Optimization:")
            optimization_report = await optimize_hook_system()
            print(f"  Health status: {optimization_report['health_status']}")
            print(f"  Success rate: {optimization_report['performance_summary']['success_rate']:.1%}")
            print(f"  Cache efficiency: {optimization_report['performance_summary']['cache_efficiency']:.1%}")

            if optimization_report["recommendations"]:
                print("  Recommendations:")
                for rec in optimization_report["recommendations"]:
                    print(f"    • {rec}")
            else:
                print("  ✓ No optimization recommendations needed")

            print("\n✅ Phase 2 optimizations test completed successfully!")

        except Exception as e:
            print(f"\n❌ Test failed: {str(e)}")
            import traceback

            traceback.print_exc()

        finally:
            # Enhanced cleanup with Phase 2 features
            print("\n🧹 Cleaning up Phase 2 resources...")
            await manager.cleanup()
            print("✅ Cleanup completed")

    # Run Phase 2 test
    asyncio.run(test_phase2_optimizations())
